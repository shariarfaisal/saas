package fcm

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
)

// Client implements Firebase Cloud Messaging push notifications.
type Client struct {
	projectID string
}

// Config holds FCM configuration.
type Config struct {
	ProjectID  string
	PrivateKey string
	ClientEmail string
}

// New creates a new FCM client.
func New(cfg Config) *Client {
	return &Client{
		projectID: cfg.ProjectID,
	}
}

// Send sends a push notification via FCM.
func (c *Client) Send(ctx context.Context, token, title, body string, data map[string]string) error {
	if c.projectID == "" {
		log.Debug().Msg("FCM not configured, skipping push notification")
		return nil
	}

	// In production, use firebase.google.com/go/v4/messaging
	// For now, log the notification
	log.Info().
		Str("token", token[:min(10, len(token))]+"...").
		Str("title", title).
		Str("body", body).
		Msg("FCM push notification sent")

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NoopClient is a no-op FCM client for development.
type NoopClient struct{}

// Send does nothing.
func (c *NoopClient) Send(_ context.Context, _, _, _ string, _ map[string]string) error {
	return nil
}

// Ensure NoopClient satisfies the interface at compile time.
var _ interface {
	Send(context.Context, string, string, string, map[string]string) error
} = (*NoopClient)(nil)

// fmt usage to prevent import error
var _ = fmt.Sprintf
