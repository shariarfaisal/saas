package order

import (
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
)

// validTransitions defines allowed predecessor states for each target status.
var validTransitions = map[sqlc.OrderStatus][]sqlc.OrderStatus{
	sqlc.OrderStatusCreated:   {sqlc.OrderStatusPending},
	sqlc.OrderStatusConfirmed: {sqlc.OrderStatusCreated},
	sqlc.OrderStatusPreparing: {sqlc.OrderStatusConfirmed, sqlc.OrderStatusCreated},
	sqlc.OrderStatusReady:     {sqlc.OrderStatusPreparing, sqlc.OrderStatusConfirmed},
	sqlc.OrderStatusPicked:    {sqlc.OrderStatusReady, sqlc.OrderStatusPreparing},
	sqlc.OrderStatusDelivered: {sqlc.OrderStatusPicked},
	sqlc.OrderStatusRejected:  {sqlc.OrderStatusCreated},
	sqlc.OrderStatusCancelled: {
		sqlc.OrderStatusPending,
		sqlc.OrderStatusCreated,
		sqlc.OrderStatusConfirmed,
		sqlc.OrderStatusPreparing,
		sqlc.OrderStatusReady,
	},
}

// ValidateTransition checks if transitioning from `from` to `to` is allowed.
func ValidateTransition(from, to sqlc.OrderStatus) error {
	allowed, ok := validTransitions[to]
	if !ok {
		return apperror.UnprocessableEntity("unknown target status: " + string(to))
	}
	for _, s := range allowed {
		if s == from {
			return nil
		}
	}
	return apperror.UnprocessableEntity("invalid status transition from " + string(from) + " to " + string(to))
}
