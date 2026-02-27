package payment

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	gateway "github.com/munchies/platform/backend/internal/platform/payment"
	"github.com/rs/zerolog/log"
)

// ProcessRefund handles refund logic for an order.
func (s *Service) ProcessRefund(ctx context.Context, orderID, tenantID uuid.UUID, amount float64, reason string, triggeredByUserID uuid.UUID) (*sqlc.Refund, error) {
	if amount <= 0 {
		return nil, apperror.BadRequest("refund amount must be positive")
	}

	// 1. Get the order
	order, err := s.q.GetOrderByID(ctx, sqlc.GetOrderByIDParams{
		ID:       orderID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("order")
	}
	if err != nil {
		return nil, apperror.Internal("fetch order", err)
	}

	// Validate order has been paid
	if order.PaymentStatus != sqlc.PaymentStatusPaid {
		return nil, apperror.BadRequest("order is not in paid status")
	}

	// Validate refund amount doesn't exceed order total
	orderTotal, totalErr := order.TotalAmount.Float64Value()
	if totalErr != nil || !orderTotal.Valid {
		return nil, apperror.Internal("parse order total", totalErr)
	}
	if amount > orderTotal.Float64 {
		return nil, apperror.BadRequest("refund amount exceeds order total")
	}

	// 2. Find the successful payment transaction
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

	refundAmount := float64ToNumeric(amount)
	now := time.Now()
	var refundStatus sqlc.RefundStatus
	var gatewayRefundID sql.NullString

	// 3. Process based on payment method
	switch txn.PaymentMethod {
	case sqlc.PaymentMethodBkash, sqlc.PaymentMethodAamarpay, sqlc.PaymentMethodSslcommerz:
		refundStatus, gatewayRefundID, err = s.processGatewayRefund(ctx, txn, amount, reason)
		if err != nil {
			return nil, err
		}

	case sqlc.PaymentMethodWallet:
		if err := s.processWalletRefund(ctx, txn, refundAmount, orderID, tenantID); err != nil {
			return nil, err
		}
		refundStatus = sqlc.RefundStatusProcessed

	case sqlc.PaymentMethodCod:
		refundStatus = sqlc.RefundStatusApproved

	default:
		return nil, apperror.BadRequest("unsupported payment method for refund: " + string(txn.PaymentMethod))
	}

	// 4. Create refund record
	approvedBy := pgtype.UUID{Bytes: triggeredByUserID, Valid: true}
	processedAt := pgtype.Timestamptz{}
	if refundStatus == sqlc.RefundStatusProcessed {
		processedAt = pgtype.Timestamptz{Time: now, InfinityModifier: pgtype.Finite, Valid: true}
	}

	refund, err := s.q.CreateRefund(ctx, sqlc.CreateRefundParams{
		TenantID:        tenantID,
		OrderID:         orderID,
		TransactionID:   txn.ID,
		Amount:          refundAmount,
		Reason:          reason,
		Status:          refundStatus,
		GatewayRefundID: gatewayRefundID,
		ApprovedBy:      approvedBy,
		ApprovedAt:      pgtype.Timestamptz{Time: now, InfinityModifier: pgtype.Finite, Valid: true},
		ProcessedAt:     processedAt,
	})
	if err != nil {
		return nil, apperror.Internal("create refund record", err)
	}

	// 5. Update payment transaction status to refunded
	_, err = s.q.UpdateTransactionStatus(ctx, sqlc.UpdateTransactionStatusParams{
		ID:       txn.ID,
		TenantID: tenantID,
		Status:   sqlc.NullTxnStatus{TxnStatus: sqlc.TxnStatusRefunded, Valid: true},
	})
	if err != nil {
		log.Error().Err(err).Str("txn_id", txn.ID.String()).Msg("failed to update transaction status to refunded")
	}

	// 6. Update order payment status
	paymentStatus := sqlc.PaymentStatusRefunded
	if amount < orderTotal.Float64 {
		paymentStatus = sqlc.PaymentStatusPartiallyRefunded
	}
	_, err = s.q.UpdateOrderPaymentStatus(ctx, sqlc.UpdateOrderPaymentStatusParams{
		ID:            orderID,
		TenantID:      tenantID,
		PaymentStatus: paymentStatus,
	})
	if err != nil {
		log.Error().Err(err).Str("order_id", orderID.String()).Msg("failed to update order payment status")
	}

	// 7. Create timeline event
	metadata, _ := json.Marshal(map[string]interface{}{
		"refund_id":     refund.ID,
		"refund_amount": amount,
		"reason":        reason,
		"status":        string(refundStatus),
	})
	_, err = s.q.CreateTimelineEvent(ctx, sqlc.CreateTimelineEventParams{
		OrderID:     orderID,
		TenantID:    tenantID,
		EventType:   "refund_processed",
		Description: fmt.Sprintf("Refund of %.2f BDT processed: %s", amount, reason),
		ActorID:     pgtype.UUID{Bytes: triggeredByUserID, Valid: true},
		ActorType:   sqlc.ActorTypeSystem,
		Metadata:    metadata,
	})
	if err != nil {
		log.Error().Err(err).Str("order_id", orderID.String()).Msg("failed to create refund timeline event")
	}

	return &refund, nil
}

func (s *Service) processGatewayRefund(ctx context.Context, txn sqlc.PaymentTransaction, amount float64, reason string) (sqlc.RefundStatus, sql.NullString, error) {
	gw, ok := s.gateways[txn.PaymentMethod]
	if !ok {
		return "", sql.NullString{}, apperror.BadRequest("gateway not configured for " + string(txn.PaymentMethod))
	}

	gwTxnID := ""
	if txn.GatewayTransactionID.Valid {
		gwTxnID = txn.GatewayTransactionID.String
	}

	refundResp, err := gw.Refund(ctx, gateway.RefundRequest{
		GatewayTxnID: gwTxnID,
		Amount:       fmt.Sprintf("%.2f", amount),
		Reason:       reason,
		RefundID:     uuid.New().String(),
	})
	if err != nil {
		return "", sql.NullString{}, apperror.Internal("gateway refund failed", err)
	}

	status := sqlc.RefundStatusProcessed
	if refundResp.Status == sqlc.TxnStatusPending {
		status = sqlc.RefundStatusPending
	}

	return status, sql.NullString{String: refundResp.GatewayRefundID, Valid: refundResp.GatewayRefundID != ""}, nil
}

func (s *Service) processWalletRefund(ctx context.Context, txn sqlc.PaymentTransaction, refundAmount pgtype.Numeric, orderID, tenantID uuid.UUID) error {
	// Credit the user's wallet balance
	if err := s.q.CreditUserWallet(ctx, sqlc.CreditUserWalletParams{
		ID:     txn.UserID,
		Amount: refundAmount,
	}); err != nil {
		return apperror.Internal("credit wallet balance", err)
	}

	// Get updated user to retrieve new balance
	user, err := s.q.GetUserByID(ctx, txn.UserID)
	if err != nil {
		return apperror.Internal("fetch user for wallet balance", err)
	}

	// Create wallet transaction record
	_, err = s.q.CreateWalletTransaction(ctx, sqlc.CreateWalletTransactionParams{
		UserID:       txn.UserID,
		TenantID:     tenantID,
		OrderID:      pgtype.UUID{Bytes: orderID, Valid: true},
		Type:         sqlc.WalletTypeCredit,
		Source:       sqlc.WalletSourceRefund,
		Amount:       refundAmount,
		BalanceAfter: user.WalletBalance,
		Description:  sql.NullString{String: "Refund for order " + orderID.String(), Valid: true},
	})
	if err != nil {
		return apperror.Internal("create wallet transaction", err)
	}

	return nil
}

func float64ToNumeric(f float64) pgtype.Numeric {
	// Convert to cents (2 decimal places) for precise integer representation
	cents := new(big.Float).SetPrec(128).SetFloat64(f)
	cents.Mul(cents, new(big.Float).SetInt64(100))
	bi, _ := cents.Int(nil)
	return pgtype.Numeric{
		Int:              bi,
		Exp:              -2,
		Valid:            true,
		InfinityModifier: pgtype.Finite,
	}
}
