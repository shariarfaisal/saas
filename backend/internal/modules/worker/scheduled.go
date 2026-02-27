package worker

import (
	"context"
	"time"

	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/rs/zerolog/log"
)

// AutoConfirmOrders auto-confirms orders past their auto_confirm_at timeout.
func (w *Worker) AutoConfirmOrders(ctx context.Context) error {
	orders, err := w.q.ListCreatedOrdersPastTimeout(ctx, 100)
	if err != nil {
		return err
	}

	confirmed := 0
	for _, order := range orders {
		_, err := w.q.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
			ID:       order.ID,
			TenantID: order.TenantID,
			Status:   sqlc.OrderStatusConfirmed,
		})
		if err != nil {
			log.Error().Err(err).Str("order_id", order.ID.String()).Msg("failed to auto-confirm order")
			continue
		}
		confirmed++
	}

	if confirmed > 0 {
		log.Info().Int("count", confirmed).Msg("auto-confirmed orders")
	}
	return nil
}

// AutoCancelOrders cancels pending orders older than 30 minutes.
func (w *Worker) AutoCancelOrders(ctx context.Context) error {
	olderThan := time.Now().Add(-30 * time.Minute)
	orders, err := w.q.ListPendingOrdersPastTimeout(ctx, sqlc.ListPendingOrdersPastTimeoutParams{
		OlderThan: olderThan,
		Limit:     100,
	})
	if err != nil {
		return err
	}

	cancelled := 0
	for _, order := range orders {
		_, err := w.q.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
			ID:       order.ID,
			TenantID: order.TenantID,
			Status:   sqlc.OrderStatusCancelled,
		})
		if err != nil {
			log.Error().Err(err).Str("order_id", order.ID.String()).Msg("failed to auto-cancel order")
			continue
		}
		cancelled++
	}

	if cancelled > 0 {
		log.Info().Int("count", cancelled).Msg("auto-cancelled pending orders")
	}
	return nil
}

// CleanupNotifications purges notifications older than 90 days.
func (w *Worker) CleanupNotifications(ctx context.Context) error {
	before := time.Now().AddDate(0, 0, -90)
	if err := w.q.PurgeOldNotifications(ctx, before); err != nil {
		return err
	}
	log.Info().Msg("purged old notifications (>90 days)")
	return nil
}

// CleanupOrderTimeline purges order timeline events older than 1 year.
func (w *Worker) CleanupOrderTimeline(ctx context.Context) error {
	before := time.Now().AddDate(-1, 0, 0)
	if err := w.q.PurgeOldOrderTimeline(ctx, before); err != nil {
		return err
	}
	log.Info().Msg("purged old order timeline events (>1 year)")
	return nil
}

// CleanupSearchLogs purges search logs older than 90 days.
func (w *Worker) CleanupSearchLogs(ctx context.Context) error {
	before := time.Now().AddDate(0, 0, -90)
	if err := w.q.PurgeOldSearchLogs(ctx, before); err != nil {
		return err
	}
	log.Info().Msg("purged old search logs (>90 days)")
	return nil
}

// CleanupAuditLogs purges audit logs older than 2 years.
func (w *Worker) CleanupAuditLogs(ctx context.Context) error {
	before := time.Now().AddDate(-2, 0, 0)
	if err := w.q.PurgeOldAuditLogs(ctx, before); err != nil {
		return err
	}
	log.Info().Msg("purged old audit logs (>2 years)")
	return nil
}
