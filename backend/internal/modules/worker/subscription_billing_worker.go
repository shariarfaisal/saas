package worker

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/rs/zerolog/log"
)

// GenerateSubscriptionInvoices creates subscription invoices for tenants on their billing day.
func (w *Worker) GenerateSubscriptionInvoices(ctx context.Context) error {
	today := time.Now().UTC()
	currentDay := today.Day()

	tenants, err := w.q.ListAllActiveTenants(ctx)
	if err != nil {
		return err
	}

	created := 0
	for _, t := range tenants {
		if int(t.BillingDay) != currentDay {
			continue
		}

		periodStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
		periodEnd := periodStart.AddDate(0, 1, -1)
		dueDate := today.AddDate(0, 0, 14)

		_, err := w.q.GetSubscriptionInvoiceByTenantAndPeriod(ctx, sqlc.GetSubscriptionInvoiceByTenantAndPeriodParams{
			TenantID:           t.ID,
			BillingPeriodStart: pgtype.Date{Time: periodStart, Valid: true},
		})
		if err == nil {
			// Invoice already exists for this period
			continue
		}

		pgAmount := pgtype.Numeric{}
		_ = pgAmount.Scan("0")

		_, err = w.q.CreateSubscriptionInvoice(ctx, sqlc.CreateSubscriptionInvoiceParams{
			TenantID:           t.ID,
			Amount:             pgAmount,
			BillingPeriodStart: pgtype.Date{Time: periodStart, Valid: true},
			BillingPeriodEnd:   pgtype.Date{Time: periodEnd, Valid: true},
			DueDate:            pgtype.Date{Time: dueDate, Valid: true},
		})
		if err != nil {
			log.Error().Err(err).Str("tenant_id", t.ID.String()).Msg("failed to create subscription invoice")
			continue
		}
		created++
	}

	if created > 0 {
		log.Info().Int("count", created).Msg("created subscription invoices")
	}
	return nil
}
