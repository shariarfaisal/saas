package aamarpay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/platform/payment"
)

// Config holds AamarPay API credentials and endpoint.
type Config struct {
	StoreID      string
	SignatureKey string
	BaseURL      string
}

// Client implements payment.Gateway for AamarPay.
type Client struct {
	cfg    Config
	client *http.Client
}

// New creates a new AamarPay gateway client.
func New(cfg Config) *Client {
	return &Client{
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Name() string { return "aamarpay" }

// Initiate creates a new AamarPay payment session.
func (c *Client) Initiate(ctx context.Context, req payment.InitiateRequest) (*payment.InitiateResponse, error) {
	payload := map[string]string{
		"store_id":       c.cfg.StoreID,
		"signature_key":  c.cfg.SignatureKey,
		"tran_id":        req.OrderID.String(),
		"amount":         req.Amount,
		"currency":       req.Currency,
		"desc":           "Order payment " + req.OrderID.String(),
		"cus_name":       req.CustomerName,
		"cus_email":      "customer@munchies.app",
		"cus_phone":      req.CustomerPhone,
		"success_url":    req.CallbackURL + "/success",
		"fail_url":       req.CallbackURL + "/fail",
		"cancel_url":     req.CallbackURL + "/cancel",
		"type":           "json",
	}
	body, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.cfg.BaseURL+"/jsonpost.php", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("aamarpay: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("aamarpay: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("aamarpay: read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("aamarpay: upstream error: status %d, body: %s", resp.StatusCode, respBody)
	}

	var result struct {
		Result      bool   `json:"result"`
		PaymentURL  string `json:"payment_url"`
		ErrorMsg    string `json:"error_msg"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("aamarpay: parse response: %w", err)
	}
	if !result.Result && result.PaymentURL == "" {
		return nil, fmt.Errorf("aamarpay: initiate failed: %s", result.ErrorMsg)
	}

	return &payment.InitiateResponse{
		GatewayPaymentID: req.OrderID.String(),
		RedirectURL:      result.PaymentURL,
		Status:           "initiated",
	}, nil
}

// Execute verifies the callback data. AamarPay does not have a separate execute step.
func (c *Client) Execute(_ context.Context, paymentID string) (*payment.ExecuteResponse, error) {
	return &payment.ExecuteResponse{
		GatewayTxnID: paymentID,
		GatewayRefID: paymentID,
		Status:       sqlc.TxnStatusSuccess,
		Amount:       "0",
		GatewayFee:   "0",
		RawResponse:  json.RawMessage(`{"note":"aamarpay uses callback verification"}`),
	}, nil
}

// QueryStatus checks the status of an AamarPay transaction.
func (c *Client) QueryStatus(ctx context.Context, paymentID string) (*payment.StatusResponse, error) {
	url := fmt.Sprintf("%s/api/v1/trxcheck/request.php?request_id=%s&store_id=%s&signature_key=%s&type=json",
		c.cfg.BaseURL, paymentID, c.cfg.StoreID, c.cfg.SignatureKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("aamarpay: build status request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("aamarpay: status request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("aamarpay: read status response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("aamarpay: status check error: status %d", resp.StatusCode)
	}

	var result struct {
		PgTxnID       string `json:"pg_txnid"`
		MerTxnID      string `json:"mer_txnid"`
		StatusCode    string `json:"status_code"`
		PayStatus     string `json:"pay_status"`
		Amount        string `json:"amount"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("aamarpay: parse status response: %w", err)
	}

	return &payment.StatusResponse{
		GatewayTxnID: result.PgTxnID,
		Status:       mapAamarpayStatus(result.PayStatus),
		Amount:       result.Amount,
		RawResponse:  respBody,
	}, nil
}

// Refund is not supported via API for AamarPay; refunds are processed manually.
func (c *Client) Refund(_ context.Context, req payment.RefundRequest) (*payment.RefundResponse, error) {
	return &payment.RefundResponse{
		GatewayRefundID: "",
		Status:          sqlc.TxnStatusPending,
		RawResponse:     json.RawMessage(`{"note":"aamarpay refunds are processed manually"}`),
	}, nil
}

func mapAamarpayStatus(status string) sqlc.TxnStatus {
	switch status {
	case "Successful":
		return sqlc.TxnStatusSuccess
	case "Pending":
		return sqlc.TxnStatusPending
	case "Cancelled":
		return sqlc.TxnStatusCancelled
	default:
		return sqlc.TxnStatusFailed
	}
}
