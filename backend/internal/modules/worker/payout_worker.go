package worker

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/shopspring/decimal"
	"github.com/rs/zerolog/log"
)

// GenerateWeeklyPayouts runs on a schedule and creates payout records for riders with unpaid earnings.
// It is intended to be called daily; actual payout generation only occurs on Mondays.
func (w *Worker) GenerateWeeklyPayouts(ctx context.Context) error {
	if time.Now().Weekday() != time.Monday {
		return nil
	}

	riders, err := w.q.ListRidersWithUnpaidEarnings(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	weekAgo := now.AddDate(0, 0, -7)

	created := 0
	for _, r := range riders {
		earnings, err := w.q.ListUnpaidEarningsByRider(ctx, r.RiderID)
		if err != nil {
			log.Error().Err(err).Str("rider_id", r.RiderID.String()).Msg("failed to fetch unpaid earnings")
			continue
		}
		if len(earnings) == 0 {
			continue
		}

		// Sum total earnings
		total := decimal.Zero
		for _, e := range earnings {
			if e.TotalEarning.Valid {
				f, err := e.TotalEarning.Float64Value()
				if err == nil {
					total = total.Add(decimal.NewFromFloat(f.Float64))
				}
			}
		}

		pgTotal := pgtype.Numeric{}
		_ = pgTotal.Scan(total.String())

		payout, err := w.q.CreateRiderPayout(ctx, sqlc.CreateRiderPayoutParams{
			RiderID:       r.RiderID,
			TenantID:      r.TenantID,
			Amount:        pgTotal,
			EarningsFrom:  pgtype.Date{Time: weekAgo, Valid: true},
			EarningsTo:    pgtype.Date{Time: now, Valid: true},
			PaymentMethod: "bank_transfer",
		})
		if err != nil {
			log.Error().Err(err).Str("rider_id", r.RiderID.String()).Msg("failed to create rider payout")
			continue
		}

		if err := w.q.LinkEarningsToPayout(ctx, sqlc.LinkEarningsToPayoutParams{
			RiderID:  r.RiderID,
			PayoutID: pgtype.UUID{Bytes: payout.ID, Valid: true},
		}); err != nil {
			log.Error().Err(err).Str("payout_id", payout.ID.String()).Msg("failed to link earnings to payout")
		}
		created++
	}

	if created > 0 {
		log.Info().Int("count", created).Msg("created weekly rider payouts")
	}
	return nil
}
