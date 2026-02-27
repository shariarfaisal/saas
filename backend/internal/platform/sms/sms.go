package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Sender is the interface for SMS providers.
type Sender interface {
	Send(ctx context.Context, phone, message string) error
}

// SSLWireless implements Sender for SSL Wireless (Bangladesh).
type SSLWireless struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewSSLWireless creates a new SSL Wireless SMS sender.
func NewSSLWireless(apiKey, baseURL string) *SSLWireless {
	return &SSLWireless{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Send sends an SMS via SSL Wireless API.
func (s *SSLWireless) Send(ctx context.Context, phone, message string) error {
	payload := map[string]string{
		"api_key": s.apiKey,
		"smsBody": message,
		"msisdn":  phone,
		"csmsId":  fmt.Sprintf("%d", time.Now().UnixNano()),
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/api/v3/send-sms", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("sms: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("sms: send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("sms: upstream error: status %d", resp.StatusCode)
	}
	return nil
}

// NoopSender is a no-op SMS sender for development/testing.
type NoopSender struct{}

func (n *NoopSender) Send(_ context.Context, _, _ string) error {
	return nil
}
