package delivery

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
)

// Service calculates delivery charges.
type Service struct {
	q *sqlc.Queries
}

// NewService creates a new delivery service.
func NewService(q *sqlc.Queries) *Service {
	return &Service{q: q}
}

// CalculateResult holds the delivery charge calculation result.
type CalculateResult struct {
	DeliveryCharge           pgtype.Numeric `json:"delivery_charge"`
	MinOrderAmount           pgtype.Numeric `json:"min_order_amount"`
	EstimatedDeliveryMinutes int32          `json:"estimated_delivery_minutes"`
	FreeDeliveryThreshold    pgtype.Numeric `json:"free_delivery_threshold"`
}

// Calculate returns the delivery charge for a given tenant, hub, and area slug.
func (s *Service) Calculate(ctx context.Context, tenantID uuid.UUID, hubID uuid.UUID, areaSlug string) (*CalculateResult, error) {
	area, err := s.q.GetHubAreaByName(ctx, sqlc.GetHubAreaByNameParams{
		HubID: hubID,
		Slug:  areaSlug,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("coverage area")
	}
	if err != nil {
		return nil, err
	}
	if !area.IsActive {
		return nil, apperror.BadRequest("area is not active")
	}

	result := &CalculateResult{
		DeliveryCharge:           area.DeliveryCharge,
		MinOrderAmount:           area.MinOrderAmount,
		EstimatedDeliveryMinutes: area.EstimatedDeliveryMinutes,
	}

	// Attempt to load zone config for free delivery threshold
	cfg, err := s.q.GetDeliveryZoneConfig(ctx, tenantID)
	if err == nil {
		result.FreeDeliveryThreshold = cfg.FreeDeliveryThreshold
	}

	return result, nil
}
