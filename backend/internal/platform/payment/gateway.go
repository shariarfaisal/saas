package payment

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/munchies/platform/backend/internal/db/sqlc"
)

// Gateway defines the common interface for all payment gateways.
type Gateway interface {
	Name() string
	Initiate(ctx context.Context, req InitiateRequest) (*InitiateResponse, error)
	Execute(ctx context.Context, paymentID string) (*ExecuteResponse, error)
	QueryStatus(ctx context.Context, paymentID string) (*StatusResponse, error)
	Refund(ctx context.Context, req RefundRequest) (*RefundResponse, error)
}

// InitiateRequest contains the data needed to start a payment.
type InitiateRequest struct {
	OrderID       uuid.UUID
	Amount        string
	Currency      string
	CallbackURL   string
	CustomerName  string
	CustomerPhone string
}

// InitiateResponse is returned after initiating a payment with the gateway.
type InitiateResponse struct {
	GatewayPaymentID string
	RedirectURL      string
	Status           string
}

// ExecuteResponse is returned after executing/confirming a payment.
type ExecuteResponse struct {
	GatewayTxnID string
	GatewayRefID string
	Status       sqlc.TxnStatus
	Amount       string
	GatewayFee   string
	RawResponse  json.RawMessage
}

// StatusResponse is returned when querying a payment's status.
type StatusResponse struct {
	GatewayTxnID string
	Status       sqlc.TxnStatus
	Amount       string
	RawResponse  json.RawMessage
}

// RefundRequest contains the data needed to issue a refund.
type RefundRequest struct {
	GatewayTxnID string
	Amount       string
	Reason       string
	RefundID     string
}

// RefundResponse is returned after processing a refund.
type RefundResponse struct {
	GatewayRefundID string
	Status          sqlc.TxnStatus
	RawResponse     json.RawMessage
}
