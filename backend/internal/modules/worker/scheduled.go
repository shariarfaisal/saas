package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
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

// AutoCancelOrders cancels pending orders older than 30 minutes, releasing reserved stock.
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
		if err := w.cancelPendingOrder(ctx, order); err != nil {
			log.Error().Err(err).Str("order_id", order.ID.String()).Msg("failed to auto-cancel order")
			continue
		}
		cancelled++
	}

	if cancelled > 0 {
		log.Info().Int("count", cancelled).Msg("auto-cancelled pending orders (payment timeout)")
	}
	return nil
}

// cancelPendingOrder cancels a single pending order with stock release and outbox event.
func (w *Worker) cancelPendingOrder(ctx context.Context, order sqlc.Order) error {
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := w.q.WithTx(tx)

	// Release reserved stock
	items, err := qtx.GetOrderItemsByOrder(ctx, order.ID)
	if err != nil {
		return err
	}
	for _, item := range items {
		if _, err := qtx.ReleaseStock(ctx, sqlc.ReleaseStockParams{
			Qty:          item.Quantity,
			ProductID:    item.ProductID,
			RestaurantID: item.RestaurantID,
			TenantID:     order.TenantID,
		}); err != nil {
			log.Warn().Err(err).
				Str("order_id", order.ID.String()).
				Str("product_id", item.ProductID.String()).
				Msg("failed to release stock during payment timeout cancellation")
		}
	}

	// Transition order status to cancelled
	_, err = qtx.TransitionOrderStatus(ctx, sqlc.TransitionOrderStatusParams{
		NewStatus:          sqlc.OrderStatusCancelled,
		CancellationReason: sql.NullString{String: "payment timeout", Valid: true},
		CancelledBy:        sqlc.NullActorType{ActorType: sqlc.ActorTypeSystem, Valid: true},
		RejectionReason:    sql.NullString{},
		RejectedBy:         sqlc.NullActorType{},
		ID:                 order.ID,
		TenantID:           order.TenantID,
	})
	if err != nil {
		return err
	}

	// Add timeline event
	if _, err := qtx.AddTimelineEvent(ctx, sqlc.AddTimelineEventParams{
		OrderID:        order.ID,
		TenantID:       order.TenantID,
		EventType:      "payment_timeout",
		PreviousStatus: sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusPending, Valid: true},
		NewStatus:      sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusCancelled, Valid: true},
		Description:    "Order cancelled due to payment timeout",
		ActorID:        pgtype.UUID{},
		ActorType:      sqlc.ActorTypeSystem,
		Metadata:       json.RawMessage("{}"),
	}); err != nil {
		log.Warn().Err(err).Str("order_id", order.ID.String()).Msg("failed to add timeline event for payment timeout")
	}

	// Create outbox event
	payload, _ := json.Marshal(map[string]interface{}{
		"order_id":  order.ID.String(),
		"tenant_id": order.TenantID.String(),
		"reason":    "payment_timeout",
	})
	if _, err := qtx.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{
		TenantID:      pgtype.UUID{Bytes: order.TenantID, Valid: true},
		AggregateType: "order",
		AggregateID:   order.ID,
		EventType:     "order.payment_timeout",
		Payload:       payload,
		MaxAttempts:   5,
	}); err != nil {
		log.Warn().Err(err).Str("order_id", order.ID.String()).Msg("failed to create outbox event for payment timeout")
	}

	return tx.Commit(ctx)
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
