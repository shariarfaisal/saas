package bkash

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/platform/payment"
)

// Config holds bKash API credentials and endpoint.
type Config struct {
	AppKey    string
	AppSecret string
	BaseURL   string
	Username  string
	Password  string
}

// Client implements payment.Gateway for bKash tokenized checkout.
type Client struct {
	cfg    Config
	client *http.Client

	mu           sync.Mutex
	token        string
	tokenExpiry  time.Time
}

// New creates a new bKash gateway client.
func New(cfg Config) *Client {
	return &Client{
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Name() string { return "bkash" }

// grantToken obtains or returns a cached auth token. Thread-safe.
func (c *Client) grantToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.tokenExpiry) {
		return c.token, nil
	}

	payload := map[string]string{
		"app_key":    c.cfg.AppKey,
		"app_secret": c.cfg.AppSecret,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.cfg.BaseURL+"/tokenized/checkout/token/grant", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("bkash: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("username", c.cfg.Username)
	req.Header.Set("password", c.cfg.Password)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("bkash: token request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("bkash: read token response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bkash: token grant failed: status %d, body: %s", resp.StatusCode, respBody)
	}

	var result struct {
		IDToken      string `json:"id_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		StatusCode   string `json:"statusCode"`
		StatusMessage string `json:"statusMessage"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("bkash: parse token response: %w", err)
	}
	if result.IDToken == "" {
		return "", fmt.Errorf("bkash: token grant failed: %s", result.StatusMessage)
	}

	c.token = result.IDToken
	// Refresh 60s before expiry for safety
	if result.ExpiresIn <= 0 {
		result.ExpiresIn = 3600
	}
	c.tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn-60) * time.Second)

	return c.token, nil
}

// doAuthedRequest performs an authenticated POST to the bKash API.
func (c *Client) doAuthedRequest(ctx context.Context, path string, payload interface{}) ([]byte, error) {
	token, err := c.grantToken(ctx)
	if err != nil {
		return nil, err
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("bkash: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	req.Header.Set("X-App-Key", c.cfg.AppKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bkash: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("bkash: read response body: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("bkash: upstream error: status %d, body: %s", resp.StatusCode, respBody)
	}
	return respBody, nil
}

// Initiate creates a new bKash payment.
func (c *Client) Initiate(ctx context.Context, req payment.InitiateRequest) (*payment.InitiateResponse, error) {
	payload := map[string]string{
		"mode":                "0011",
		"payerReference":      req.OrderID.String(),
		"callbackURL":         req.CallbackURL,
		"amount":              req.Amount,
		"currency":            req.Currency,
		"intent":              "sale",
		"merchantInvoiceNumber": req.OrderID.String(),
	}

	respBody, err := c.doAuthedRequest(ctx, "/tokenized/checkout/create", payload)
	if err != nil {
		return nil, err
	}

	var result struct {
		PaymentID     string `json:"paymentID"`
		BkashURL      string `json:"bkashURL"`
		StatusCode    string `json:"statusCode"`
		StatusMessage string `json:"statusMessage"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("bkash: parse initiate response: %w", err)
	}
	if result.PaymentID == "" {
		return nil, fmt.Errorf("bkash: initiate failed: %s", result.StatusMessage)
	}

	return &payment.InitiateResponse{
		GatewayPaymentID: result.PaymentID,
		RedirectURL:      result.BkashURL,
		Status:           result.StatusCode,
	}, nil
}

// Execute confirms a bKash payment after customer approval.
func (c *Client) Execute(ctx context.Context, paymentID string) (*payment.ExecuteResponse, error) {
	payload := map[string]string{
		"paymentID": paymentID,
	}

	respBody, err := c.doAuthedRequest(ctx, "/tokenized/checkout/execute", payload)
	if err != nil {
		return nil, err
	}

	var result struct {
		PaymentID          string `json:"paymentID"`
		TrxID              string `json:"trxID"`
		TransactionStatus  string `json:"transactionStatus"`
		Amount             string `json:"amount"`
		Charge             string `json:"charge"`
		StatusCode         string `json:"statusCode"`
		StatusMessage      string `json:"statusMessage"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("bkash: parse execute response: %w", err)
	}

	status := mapBkashStatus(result.TransactionStatus)

	return &payment.ExecuteResponse{
		GatewayTxnID: result.TrxID,
		GatewayRefID: result.PaymentID,
		Status:       status,
		Amount:       result.Amount,
		GatewayFee:   result.Charge,
		RawResponse:  respBody,
	}, nil
}

// QueryStatus checks the current status of a bKash payment.
func (c *Client) QueryStatus(ctx context.Context, paymentID string) (*payment.StatusResponse, error) {
	payload := map[string]string{
		"paymentID": paymentID,
	}

	respBody, err := c.doAuthedRequest(ctx, "/tokenized/checkout/payment/status", payload)
	if err != nil {
		return nil, err
	}

	var result struct {
		PaymentID         string `json:"paymentID"`
		TrxID             string `json:"trxID"`
		TransactionStatus string `json:"transactionStatus"`
		Amount            string `json:"amount"`
		StatusCode        string `json:"statusCode"`
		StatusMessage     string `json:"statusMessage"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("bkash: parse status response: %w", err)
	}

	return &payment.StatusResponse{
		GatewayTxnID: result.TrxID,
		Status:       mapBkashStatus(result.TransactionStatus),
		Amount:       result.Amount,
		RawResponse:  respBody,
	}, nil
}

// Refund processes a refund for a bKash payment.
func (c *Client) Refund(ctx context.Context, req payment.RefundRequest) (*payment.RefundResponse, error) {
	payload := map[string]string{
		"paymentID": req.GatewayTxnID,
		"amount":    req.Amount,
		"trxID":     req.GatewayTxnID,
		"sku":       req.RefundID,
		"reason":    req.Reason,
	}

	respBody, err := c.doAuthedRequest(ctx, "/tokenized/checkout/payment/refund", payload)
	if err != nil {
		return nil, err
	}

	var result struct {
		RefundTrxID       string `json:"refundTrxID"`
		TransactionStatus string `json:"transactionStatus"`
		StatusCode        string `json:"statusCode"`
		StatusMessage     string `json:"statusMessage"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("bkash: parse refund response: %w", err)
	}

	return &payment.RefundResponse{
		GatewayRefundID: result.RefundTrxID,
		Status:          mapBkashStatus(result.TransactionStatus),
		RawResponse:     respBody,
	}, nil
}

func mapBkashStatus(status string) sqlc.TxnStatus {
	switch status {
	case "Completed":
		return sqlc.TxnStatusSuccess
	case "Initiated", "Pending", "Authorized":
		return sqlc.TxnStatusPending
	case "Cancelled":
		return sqlc.TxnStatusCancelled
	default:
		return sqlc.TxnStatusFailed
	}
}
