package payment

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	gateway "github.com/munchies/platform/backend/internal/platform/payment"
	"github.com/rs/zerolog/log"
)

// Service implements payment business logic.
type Service struct {
	q        *sqlc.Queries
	gateways map[sqlc.PaymentMethod]gateway.Gateway
}

// NewService creates a new payment service.
func NewService(q *sqlc.Queries, gateways map[sqlc.PaymentMethod]gateway.Gateway) *Service {
	return &Service{q: q, gateways: gateways}
}

// InitiatePaymentRequest holds the data needed to start a payment.
type InitiatePaymentRequest struct {
	OrderID  uuid.UUID
	TenantID uuid.UUID
	UserID   uuid.UUID
	Method   sqlc.PaymentMethod
	IPAddr   string
	UserAgent string
}

// InitiatePayment validates the order, creates a pending transaction, and initiates with the gateway.
func (s *Service) InitiatePayment(ctx context.Context, req InitiatePaymentRequest, callbackURL string, customerName, customerPhone string) (*gateway.InitiateResponse, error) {
	gw, ok := s.gateways[req.Method]
	if !ok {
		return nil, apperror.BadRequest("unsupported payment method: " + string(req.Method))
	}

	// Fetch the order
	order, err := s.q.GetOrderByID(ctx, sqlc.GetOrderByIDParams{
		ID:       req.OrderID,
		TenantID: req.TenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("order")
	}
	if err != nil {
		return nil, apperror.Internal("fetch order", err)
	}

	// Validate order state
	if order.PaymentStatus == sqlc.PaymentStatusPaid {
		return nil, apperror.Conflict("order already paid")
	}
	if order.Status != sqlc.OrderStatusPending {
		return nil, apperror.BadRequest("order is not in pending state")
	}

	// Parse amount from pgtype.Numeric
	amountStr := numericToString(order.TotalAmount)

	var ip *netip.Addr
	if req.IPAddr != "" {
		if parsed, err := netip.ParseAddr(req.IPAddr); err == nil {
			ip = &parsed
		}
	}

	// Create pending transaction
	txn, err := s.q.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		TenantID:      req.TenantID,
		OrderID:       req.OrderID,
		UserID:        req.UserID,
		PaymentMethod: req.Method,
		Status:        sqlc.TxnStatusPending,
		Amount:        order.TotalAmount,
		Currency:      "BDT",
		GatewayResponse: json.RawMessage("{}"),
		GatewayFee:    pgtype.Numeric{Int: nil, Exp: 0, NaN: false, InfinityModifier: pgtype.Finite, Valid: false},
		IpAddress:     ip,
		UserAgent:     sql.NullString{String: req.UserAgent, Valid: req.UserAgent != ""},
	})
	if err != nil {
		return nil, apperror.Internal("create transaction", err)
	}

	// Initiate with gateway
	gwResp, err := gw.Initiate(ctx, gateway.InitiateRequest{
		OrderID:       req.OrderID,
		Amount:        amountStr,
		Currency:      "BDT",
		CallbackURL:   callbackURL,
		CustomerName:  customerName,
		CustomerPhone: customerPhone,
	})
	if err != nil {
		// Update transaction to failed on gateway error
		_, _ = s.q.UpdateTransactionStatus(ctx, sqlc.UpdateTransactionStatusParams{
			ID:       txn.ID,
			TenantID: req.TenantID,
			Status:   sqlc.NullTxnStatus{TxnStatus: sqlc.TxnStatusFailed, Valid: true},
			GatewayResponse: json.RawMessage(
				fmt.Sprintf(`{"error": %q}`, err.Error()),
			),
		})
		return nil, apperror.Internal("initiate payment", err)
	}

	// Update transaction with gateway payment ID
	_, _ = s.q.UpdateTransactionStatus(ctx, sqlc.UpdateTransactionStatusParams{
		ID:       txn.ID,
		TenantID: req.TenantID,
		GatewayTransactionID: sql.NullString{
			String: gwResp.GatewayPaymentID,
			Valid:  gwResp.GatewayPaymentID != "",
		},
	})

	return gwResp, nil
}

// ProcessCallback handles an idempotent callback from a payment gateway.
func (s *Service) ProcessCallback(ctx context.Context, gatewayTxnID string, tenantID uuid.UUID, method sqlc.PaymentMethod, gwResponse json.RawMessage) (*sqlc.PaymentTransaction, error) {
	// Look up the transaction
	txn, err := s.q.GetTransactionByGatewayID(ctx, sqlc.GetTransactionByGatewayIDParams{
		GatewayTransactionID: sql.NullString{String: gatewayTxnID, Valid: true},
		TenantID:             tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("transaction")
	}
	if err != nil {
		return nil, apperror.Internal("fetch transaction", err)
	}

	// Idempotency: if already processed, return existing
	if txn.Status == sqlc.TxnStatusSuccess || txn.Status == sqlc.TxnStatusFailed {
		return &txn, nil
	}

	// Execute payment with gateway
	gw, ok := s.gateways[method]
	if !ok {
		return nil, apperror.BadRequest("unsupported payment method")
	}

	execResp, err := gw.Execute(ctx, gatewayTxnID)
	if err != nil {
		log.Error().Err(err).Str("gateway_txn_id", gatewayTxnID).Msg("gateway execute failed")
		return s.markTransactionFailed(ctx, txn, gwResponse)
	}

	now := time.Now()

	if execResp.Status == sqlc.TxnStatusSuccess {
		return s.markTransactionSuccess(ctx, txn, execResp, now)
	}

	return s.markTransactionFailed(ctx, txn, gwResponse)
}

func (s *Service) markTransactionSuccess(ctx context.Context, txn sqlc.PaymentTransaction, execResp *gateway.ExecuteResponse, now time.Time) (*sqlc.PaymentTransaction, error) {
	updated, err := s.q.UpdateTransactionStatus(ctx, sqlc.UpdateTransactionStatusParams{
		ID:       txn.ID,
		TenantID: txn.TenantID,
		Status:   sqlc.NullTxnStatus{TxnStatus: sqlc.TxnStatusSuccess, Valid: true},
		GatewayTransactionID: sql.NullString{
			String: execResp.GatewayTxnID,
			Valid:  execResp.GatewayTxnID != "",
		},
		GatewayReferenceID: sql.NullString{
			String: execResp.GatewayRefID,
			Valid:  execResp.GatewayRefID != "",
		},
		GatewayResponse: execResp.RawResponse,
		CallbackReceivedAt: pgtype.Timestamptz{
			Time:             now,
			InfinityModifier: pgtype.Finite,
			Valid:            true,
		},
	})
	if err != nil {
		return nil, apperror.Internal("update transaction status", err)
	}

	// Update order payment status to paid
	_, err = s.q.UpdateOrderPaymentStatus(ctx, sqlc.UpdateOrderPaymentStatusParams{
		ID:            txn.OrderID,
		TenantID:      txn.TenantID,
		PaymentStatus: sqlc.PaymentStatusPaid,
	})
	if err != nil {
		log.Error().Err(err).Str("order_id", txn.OrderID.String()).Msg("failed to update order payment status")
	}

	// Transition order status to created
	_, err = s.q.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
		ID:       txn.OrderID,
		TenantID: txn.TenantID,
		Status:   sqlc.OrderStatusCreated,
	})
	if err != nil {
		log.Error().Err(err).Str("order_id", txn.OrderID.String()).Msg("failed to update order status")
	}

	// Create timeline event
	_, err = s.q.CreateTimelineEvent(ctx, sqlc.CreateTimelineEventParams{
		OrderID:   txn.OrderID,
		TenantID:  txn.TenantID,
		EventType: "payment_received",
		PreviousStatus: sqlc.NullOrderStatus{
			OrderStatus: sqlc.OrderStatusPending,
			Valid:       true,
		},
		NewStatus: sqlc.NullOrderStatus{
			OrderStatus: sqlc.OrderStatusCreated,
			Valid:       true,
		},
		Description: "Payment received via " + string(txn.PaymentMethod),
		ActorType:   sqlc.ActorTypeSystem,
		Metadata:    json.RawMessage(`{}`),
	})
	if err != nil {
		log.Error().Err(err).Str("order_id", txn.OrderID.String()).Msg("failed to create timeline event")
	}

	return &updated, nil
}

func (s *Service) markTransactionFailed(ctx context.Context, txn sqlc.PaymentTransaction, gwResponse json.RawMessage) (*sqlc.PaymentTransaction, error) {
	now := time.Now()
	updated, err := s.q.UpdateTransactionStatus(ctx, sqlc.UpdateTransactionStatusParams{
		ID:              txn.ID,
		TenantID:        txn.TenantID,
		Status:          sqlc.NullTxnStatus{TxnStatus: sqlc.TxnStatusFailed, Valid: true},
		GatewayResponse: gwResponse,
		CallbackReceivedAt: pgtype.Timestamptz{
			Time:             now,
			InfinityModifier: pgtype.Finite,
			Valid:            true,
		},
	})
	if err != nil {
		return nil, apperror.Internal("update transaction status", err)
	}
	return &updated, nil
}

// GetTransactionByOrder returns the successful transaction for an order.
func (s *Service) GetTransactionByOrder(ctx context.Context, orderID, tenantID uuid.UUID) (*sqlc.PaymentTransaction, error) {
	txn, err := s.q.GetTransactionByOrderID(ctx, sqlc.GetTransactionByOrderIDParams{
		OrderID:  orderID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("payment transaction")
	}
	if err != nil {
		return nil, apperror.Internal("fetch transaction", err)
	}
	return &txn, nil
}

func numericToString(n pgtype.Numeric) string {
	if !n.Valid {
		return "0"
	}
	// Use big.Float for precise string conversion
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return "0"
	}
	return fmt.Sprintf("%.2f", f.Float64)
}
