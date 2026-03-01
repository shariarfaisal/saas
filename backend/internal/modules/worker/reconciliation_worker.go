package worker

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/rs/zerolog/log"
)

// ReconcilePayments finds stale processing payments and creates reconciliation alerts.
func (w *Worker) ReconcilePayments(ctx context.Context) error {
	staleThreshold := time.Now().Add(-1 * time.Hour)
	stale, err := w.q.ListStaleProcessingPayments(ctx, sqlc.ListStaleProcessingPaymentsParams{
		Before: staleThreshold,
		Limit:  100,
	})
	if err != nil {
		return err
	}

	created := 0
	for _, tx := range stale {
		_, err := w.q.CreateReconciliationAlert(ctx, sqlc.CreateReconciliationAlertParams{
			TenantID:             pgtype.UUID{Bytes: tx.TenantID, Valid: true},
			PaymentTransactionID: pgtype.UUID{Bytes: tx.ID, Valid: true},
			AlertType:            "stale_processing_payment",
		})
		if err != nil {
			log.Error().Err(err).Str("tx_id", tx.ID.String()).Msg("failed to create reconciliation alert")
			continue
		}
		created++
	}

	if created > 0 {
		log.Info().Int("count", created).Msg("created reconciliation alerts for stale payments")
	}
	return nil
}
