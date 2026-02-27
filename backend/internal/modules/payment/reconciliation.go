package payment

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	gateway "github.com/munchies/platform/backend/internal/platform/payment"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ReconciliationJob reconciles pending payment transactions with their gateways.
type ReconciliationJob struct {
	q        *sqlc.Queries
	gateways map[sqlc.PaymentMethod]gateway.Gateway
	logger   zerolog.Logger
}

// NewReconciliationJob creates a new reconciliation job.
func NewReconciliationJob(q *sqlc.Queries, gateways map[sqlc.PaymentMethod]gateway.Gateway) *ReconciliationJob {
	return &ReconciliationJob{
		q:        q,
		gateways: gateways,
		logger:   log.With().Str("component", "reconciliation").Logger(),
	}
}

// StartReconciliation runs the reconciliation job on a periodic interval.
func (j *ReconciliationJob) StartReconciliation(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	j.logger.Info().Dur("interval", interval).Msg("payment reconciliation job started")

	for {
		select {
		case <-ctx.Done():
			j.logger.Info().Msg("payment reconciliation job stopped")
			return
		case <-ticker.C:
			j.Run(ctx)
		}
	}
}

// Run executes a single reconciliation cycle.
func (j *ReconciliationJob) Run(ctx context.Context) {
	cutoff := time.Now().Add(-15 * time.Minute)

	pending, err := j.q.ListPendingTransactions(ctx, sqlc.ListPendingTransactionsParams{
		Limit:     100,
		OlderThan: cutoff,
	})
	if err != nil {
		j.logger.Error().Err(err).Msg("failed to list pending transactions")
		return
	}

	if len(pending) == 0 {
		j.logger.Debug().Msg("no pending transactions to reconcile")
		return
	}

	var succeeded, failed, skipped int

	for _, txn := range pending {
		result := j.reconcileTransaction(ctx, txn)
		switch result {
		case reconcileSuccess:
			succeeded++
		case reconcileFailed:
			failed++
		case reconcileSkipped:
			skipped++
		}
	}

	j.logger.Info().
		Int("total", len(pending)).
		Int("succeeded", succeeded).
		Int("failed", failed).
		Int("skipped", skipped).
		Msg("reconciliation cycle complete")
}

type reconcileResult int

const (
	reconcileSuccess reconcileResult = iota
	reconcileFailed
	reconcileSkipped
)

func (j *ReconciliationJob) reconcileTransaction(ctx context.Context, txn sqlc.PaymentTransaction) reconcileResult {
	gw, ok := j.gateways[txn.PaymentMethod]
	if !ok {
		j.logger.Warn().
			Str("txn_id", txn.ID.String()).
			Str("method", string(txn.PaymentMethod)).
			Msg("no gateway configured for payment method, skipping")
		return reconcileSkipped
	}

	gwTxnID := ""
	if txn.GatewayTransactionID.Valid {
		gwTxnID = txn.GatewayTransactionID.String
	}
	if gwTxnID == "" {
		j.logger.Warn().
			Str("txn_id", txn.ID.String()).
			Msg("transaction has no gateway ID, skipping")
		return reconcileSkipped
	}

	statusResp, err := gw.QueryStatus(ctx, gwTxnID)
	if err != nil {
		j.logger.Error().
			Err(err).
			Str("txn_id", txn.ID.String()).
			Str("gateway", gw.Name()).
			Msg("ALERT: gateway unreachable during reconciliation")
		return reconcileSkipped
	}

	now := time.Now()

	switch statusResp.Status {
	case sqlc.TxnStatusSuccess:
		return j.handleSuccess(ctx, txn, statusResp, now)
	case sqlc.TxnStatusFailed, sqlc.TxnStatusCancelled:
		return j.handleFailure(ctx, txn, statusResp, now)
	default:
		j.logger.Debug().
			Str("txn_id", txn.ID.String()).
			Str("status", string(statusResp.Status)).
			Msg("transaction still pending at gateway")
		return reconcileSkipped
	}
}

func (j *ReconciliationJob) handleSuccess(ctx context.Context, txn sqlc.PaymentTransaction, statusResp *gateway.StatusResponse, now time.Time) reconcileResult {
	// Update transaction to success
	_, err := j.q.UpdateTransactionStatus(ctx, sqlc.UpdateTransactionStatusParams{
		ID:       txn.ID,
		TenantID: txn.TenantID,
		Status:   sqlc.NullTxnStatus{TxnStatus: sqlc.TxnStatusSuccess, Valid: true},
		GatewayResponse: statusResp.RawResponse,
		CallbackReceivedAt: pgtype.Timestamptz{
			Time:             now,
			InfinityModifier: pgtype.Finite,
			Valid:            true,
		},
	})
	if err != nil {
		j.logger.Error().Err(err).Str("txn_id", txn.ID.String()).Msg("failed to update transaction to success")
		return reconcileFailed
	}

	// Update order payment status to paid
	_, err = j.q.UpdateOrderPaymentStatus(ctx, sqlc.UpdateOrderPaymentStatusParams{
		ID:            txn.OrderID,
		TenantID:      txn.TenantID,
		PaymentStatus: sqlc.PaymentStatusPaid,
	})
	if err != nil {
		j.logger.Error().Err(err).Str("order_id", txn.OrderID.String()).Msg("failed to update order payment status")
	}

	// Transition order status to created
	_, err = j.q.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
		ID:       txn.OrderID,
		TenantID: txn.TenantID,
		Status:   sqlc.OrderStatusCreated,
	})
	if err != nil {
		j.logger.Error().Err(err).Str("order_id", txn.OrderID.String()).Msg("failed to update order status")
	}

	// Create timeline event
	metadata, _ := json.Marshal(map[string]string{
		"source":     "reconciliation",
		"gateway":    string(txn.PaymentMethod),
		"gateway_id": statusResp.GatewayTxnID,
	})
	_, err = j.q.CreateTimelineEvent(ctx, sqlc.CreateTimelineEventParams{
		OrderID:   txn.OrderID,
		TenantID:  txn.TenantID,
		EventType: "payment_reconciled",
		PreviousStatus: sqlc.NullOrderStatus{
			OrderStatus: sqlc.OrderStatusPending,
			Valid:       true,
		},
		NewStatus: sqlc.NullOrderStatus{
			OrderStatus: sqlc.OrderStatusCreated,
			Valid:       true,
		},
		Description: "Payment confirmed via reconciliation (" + string(txn.PaymentMethod) + ")",
		ActorType:   sqlc.ActorTypeSystem,
		Metadata:    metadata,
	})
	if err != nil {
		j.logger.Error().Err(err).Str("order_id", txn.OrderID.String()).Msg("failed to create reconciliation timeline event")
	}

	j.logger.Info().
		Str("txn_id", txn.ID.String()).
		Str("order_id", txn.OrderID.String()).
		Msg("transaction reconciled as successful")

	return reconcileSuccess
}

func (j *ReconciliationJob) handleFailure(ctx context.Context, txn sqlc.PaymentTransaction, statusResp *gateway.StatusResponse, now time.Time) reconcileResult {
	_, err := j.q.UpdateTransactionStatus(ctx, sqlc.UpdateTransactionStatusParams{
		ID:       txn.ID,
		TenantID: txn.TenantID,
		Status:   sqlc.NullTxnStatus{TxnStatus: statusResp.Status, Valid: true},
		GatewayResponse: statusResp.RawResponse,
		CallbackReceivedAt: pgtype.Timestamptz{
			Time:             now,
			InfinityModifier: pgtype.Finite,
			Valid:            true,
		},
	})
	if err != nil {
		j.logger.Error().Err(err).Str("txn_id", txn.ID.String()).Msg("failed to update transaction to failed")
		return reconcileFailed
	}

	// Create timeline event for failure
	metadata, _ := json.Marshal(map[string]string{
		"source":  "reconciliation",
		"gateway": string(txn.PaymentMethod),
		"status":  string(statusResp.Status),
	})
	_, err = j.q.CreateTimelineEvent(ctx, sqlc.CreateTimelineEventParams{
		OrderID:     txn.OrderID,
		TenantID:    txn.TenantID,
		EventType:   "payment_failed",
		Description: "Payment " + string(statusResp.Status) + " (detected by reconciliation)",
		ActorType:   sqlc.ActorTypeSystem,
		Metadata:    metadata,
	})
	if err != nil {
		j.logger.Error().Err(err).Str("order_id", txn.OrderID.String()).Msg("failed to create failure timeline event")
	}

	j.logger.Info().
		Str("txn_id", txn.ID.String()).
		Str("order_id", txn.OrderID.String()).
		Str("status", string(statusResp.Status)).
		Msg("transaction reconciled as failed/cancelled")

	return reconcileFailed
}
