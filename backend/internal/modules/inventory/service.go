package inventory

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/pagination"
	"github.com/shopspring/decimal"
)

// Service handles inventory business logic.
type Service struct {
	q *sqlc.Queries
}

// NewService creates a new inventory service.
func NewService(q *sqlc.Queries) *Service {
	return &Service{q: q}
}

// AdjustStockRequest holds fields for a stock adjustment.
type AdjustStockRequest struct {
	InventoryItemID uuid.UUID
	TenantID        uuid.UUID
	RestaurantID    uuid.UUID
	QtyChange       int32
	Reason          sqlc.InventoryAdjustmentReason
	Note            string
	AdjustedBy      uuid.UUID
}

// AdjustStock modifies stock quantity and creates an audit log entry.
func (s *Service) AdjustStock(ctx context.Context, req AdjustStockRequest) (*sqlc.InventoryItem, *sqlc.InventoryAdjustment, error) {
	// Get current item
	item, err := s.q.GetInventoryItem(ctx, sqlc.GetInventoryItemParams{
		ID:       req.InventoryItemID,
		TenantID: req.TenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, apperror.NotFound("inventory item")
	}
	if err != nil {
		return nil, nil, apperror.Internal("get inventory item", err)
	}

	if item.RestaurantID != req.RestaurantID {
		return nil, nil, apperror.Forbidden("inventory item does not belong to this restaurant")
	}

	qtyBefore := item.StockQty
	qtyAfter := qtyBefore + req.QtyChange

	if qtyAfter < 0 {
		return nil, nil, apperror.BadRequest("insufficient stock: cannot reduce below zero")
	}

	// Adjust stock
	updated, err := s.q.AdjustStock(ctx, sqlc.AdjustStockParams{
		QtyChange: req.QtyChange,
		ID:        req.InventoryItemID,
		TenantID:  req.TenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, apperror.BadRequest("insufficient stock: cannot reduce below zero")
	}
	if err != nil {
		return nil, nil, apperror.Internal("adjust stock", err)
	}

	// Create audit log
	var costPrice pgtype.Numeric
	if item.CostPrice.Valid {
		costPrice = item.CostPrice
	}
	adj, err := s.q.CreateInventoryAdjustment(ctx, sqlc.CreateInventoryAdjustmentParams{
		InventoryItemID: req.InventoryItemID,
		TenantID:        req.TenantID,
		RestaurantID:    req.RestaurantID,
		OrderID:         pgtype.UUID{},
		AdjustmentType:  req.Reason,
		QtyBefore:       qtyBefore,
		QtyChange:       req.QtyChange,
		QtyAfter:        qtyAfter,
		CostPrice:       costPrice,
		Note:            sql.NullString{String: req.Note, Valid: req.Note != ""},
		AdjustedBy:      pgtype.UUID{Bytes: req.AdjustedBy, Valid: true},
	})
	if err != nil {
		return nil, nil, apperror.Internal("create inventory adjustment", err)
	}

	return &updated, &adj, nil
}

// ListInventory returns paginated inventory for a restaurant.
func (s *Service) ListInventory(ctx context.Context, tenantID, restaurantID uuid.UUID, page, perPage int) ([]sqlc.InventoryItem, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)

	total, err := s.q.CountInventoryByRestaurant(ctx, sqlc.CountInventoryByRestaurantParams{
		RestaurantID: restaurantID,
		TenantID:     tenantID,
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("count inventory", err)
	}

	items, err := s.q.ListInventoryByRestaurant(ctx, sqlc.ListInventoryByRestaurantParams{
		RestaurantID: restaurantID,
		TenantID:     tenantID,
		Limit:        int32(limit),
		Offset:       int32(offset),
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list inventory", err)
	}

	meta := pagination.NewMeta(total, limit, "")
	return items, meta, nil
}

// ListLowStock returns inventory items below the reorder threshold.
func (s *Service) ListLowStock(ctx context.Context, tenantID, restaurantID uuid.UUID, page, perPage int) ([]sqlc.InventoryItem, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)

	total, err := s.q.CountLowStock(ctx, sqlc.CountLowStockParams{
		TenantID:     tenantID,
		RestaurantID: restaurantID,
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("count low stock", err)
	}

	items, err := s.q.ListLowStock(ctx, sqlc.ListLowStockParams{
		TenantID:     tenantID,
		RestaurantID: restaurantID,
		Limit:        int32(limit),
		Offset:       int32(offset),
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list low stock", err)
	}

	meta := pagination.NewMeta(total, limit, "")
	return items, meta, nil
}

// CheckAndReserveStock validates stock availability and reserves it for an order.
// Used during order creation within a transaction.
func (s *Service) CheckAndReserveStock(ctx context.Context, q *sqlc.Queries, tenantID uuid.UUID, items []StockReservation) error {
	for _, item := range items {
		_, err := q.ReserveStock(ctx, sqlc.ReserveStockParams{
			Qty:          item.Quantity,
			ProductID:    item.ProductID,
			RestaurantID: item.RestaurantID,
			TenantID:     tenantID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return apperror.BadRequest("insufficient stock for product " + item.ProductID.String())
		}
		if err != nil {
			return apperror.Internal("reserve stock", err)
		}
	}
	return nil
}

// ReleaseStockForOrder releases reserved stock when an order is cancelled.
func (s *Service) ReleaseStockForOrder(ctx context.Context, q *sqlc.Queries, tenantID uuid.UUID, items []StockReservation) error {
	for _, item := range items {
		_, err := q.ReleaseStock(ctx, sqlc.ReleaseStockParams{
			Qty:          item.Quantity,
			ProductID:    item.ProductID,
			RestaurantID: item.RestaurantID,
			TenantID:     tenantID,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return apperror.Internal("release stock", err)
		}
	}
	return nil
}

// ConsumeReservedStock permanently deducts reserved stock (called when order is picked up).
func (s *Service) ConsumeReservedStock(ctx context.Context, q *sqlc.Queries, tenantID uuid.UUID, items []StockReservation) error {
	for _, item := range items {
		_, err := q.ConsumeReservedStock(ctx, sqlc.ConsumeReservedStockParams{
			Qty:          item.Quantity,
			ProductID:    item.ProductID,
			RestaurantID: item.RestaurantID,
			TenantID:     tenantID,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return apperror.Internal("consume reserved stock", err)
		}
	}
	return nil
}

// StockReservation represents a stock reservation request for a product.
type StockReservation struct {
	ProductID    uuid.UUID
	RestaurantID uuid.UUID
	Quantity     int32
}

// CreateInventoryItem creates a new inventory tracking record.
func (s *Service) CreateInventoryItem(ctx context.Context, tenantID, productID, restaurantID uuid.UUID, stockQty, reorderThreshold int32, costPrice *decimal.Decimal) (*sqlc.InventoryItem, error) {
	var cp pgtype.Numeric
	if costPrice != nil {
		cp = pgtype.Numeric{Valid: true}
		_ = cp.Scan(costPrice.String())
	}
	item, err := s.q.CreateInventoryItem(ctx, sqlc.CreateInventoryItemParams{
		ProductID:        productID,
		RestaurantID:     restaurantID,
		TenantID:         tenantID,
		StockQty:         stockQty,
		ReservedQty:      0,
		CostPrice:        cp,
		ReorderThreshold: reorderThreshold,
	})
	if err != nil {
		return nil, apperror.Internal("create inventory item", err)
	}
	return &item, nil
}
