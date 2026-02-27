package email

import (
	"bytes"
	"context"
	"html/template"

	"github.com/rs/zerolog/log"
)

// Client implements email sending.
type Client struct {
	provider string
	from     string
}

// Config holds email configuration.
type Config struct {
	Provider string // "sendgrid" or "ses"
	APIKey   string
	From     string
}

// New creates a new email client.
func New(cfg Config) *Client {
	return &Client{
		provider: cfg.Provider,
		from:     cfg.From,
	}
}

// Send sends an email.
func (c *Client) Send(ctx context.Context, to, subject, htmlBody string) error {
	// In production, integrate with SendGrid or AWS SES
	log.Info().
		Str("to", to).
		Str("subject", subject).
		Str("provider", c.provider).
		Msg("email sent")
	return nil
}

// Templates maps template names to HTML templates.
var Templates = map[string]string{
	"welcome":            welcomeTemplate,
	"invoice_ready":      invoiceReadyTemplate,
	"order_confirmation": orderConfirmationTemplate,
	"refund_processed":   refundProcessedTemplate,
	"password_reset":     passwordResetTemplate,
	"vendor_invitation":  vendorInvitationTemplate,
	"tenant_suspended":   tenantSuspendedTemplate,
}

// RenderTemplate renders an email template with variables.
func RenderTemplate(templateName string, vars map[string]interface{}) (string, error) {
	tmplStr, ok := Templates[templateName]
	if !ok {
		return "", nil
	}
	tmpl, err := template.New(templateName).Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", err
	}
	return buf.String(), nil
}

const welcomeTemplate = `<h1>Welcome to {{.PlatformName}}!</h1><p>Hi {{.Name}}, your account has been created.</p>`
const invoiceReadyTemplate = `<h1>Invoice Ready</h1><p>Invoice {{.InvoiceNumber}} for period {{.Period}} is ready for review.</p>`
const orderConfirmationTemplate = `<h1>Order Confirmed</h1><p>Your order {{.OrderNumber}} has been confirmed.</p>`
const refundProcessedTemplate = `<h1>Refund Processed</h1><p>Your refund of {{.Amount}} for order {{.OrderNumber}} has been processed.</p>`
const passwordResetTemplate = `<h1>Password Reset</h1><p>Click <a href="{{.ResetLink}}">here</a> to reset your password.</p>`
const vendorInvitationTemplate = `<h1>Vendor Invitation</h1><p>You've been invited to join {{.PlatformName}} as a vendor.</p>`
const tenantSuspendedTemplate = `<h1>Account Suspended</h1><p>Your account has been suspended. Please contact support.</p>`

// NoopClient is a no-op email client for development.
type NoopClient struct{}

// Send does nothing.
func (c *NoopClient) Send(_ context.Context, _, _, _ string) error {
	return nil
}
