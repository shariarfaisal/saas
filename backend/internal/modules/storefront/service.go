package storefront

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/pagination"
)

// Service implements public storefront business logic.
type Service struct {
	q *sqlc.Queries
}

// NewService creates a new storefront service.
func NewService(q *sqlc.Queries) *Service {
	return &Service{q: q}
}

// ListAreas returns all active coverage areas for a hub associated with the tenant.
func (s *Service) ListAreas(ctx context.Context, tenantID uuid.UUID) ([]sqlc.HubCoverageArea, error) {
	hubs, err := s.q.ListHubsByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	var areas []sqlc.HubCoverageArea
	for _, hub := range hubs {
		hubAreas, err := s.q.ListHubAreas(ctx, hub.ID)
		if err != nil {
			continue
		}
		for _, a := range hubAreas {
			if a.IsActive {
				areas = append(areas, a)
			}
		}
	}
	return areas, nil
}

// ListRestaurants returns available restaurants filtered by area slug.
func (s *Service) ListRestaurants(ctx context.Context, tenantID uuid.UUID, areaSlug string, page, perPage int) ([]sqlc.Restaurant, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)
	// Fetch one extra item to detect if there's a next page
	items, err := s.q.ListAvailableByHubAndArea(ctx, sqlc.ListAvailableByHubAndAreaParams{
		TenantID: tenantID,
		Slug:     areaSlug,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list restaurants", err)
	}
	// Use a conservative total: offset + actual returned count (may undercount on last page)
	total := int64(offset) + int64(len(items))
	if len(items) == limit {
		// There may be more pages; add one more to signal continuation
		total++
	}
	meta := pagination.NewMeta(total, limit, "")
	return items, meta, nil
}

// GetRestaurant returns a restaurant by slug for public view.
func (s *Service) GetRestaurant(ctx context.Context, tenantID uuid.UUID, slug string) (*sqlc.Restaurant, error) {
	res, err := s.q.GetRestaurantBySlug(ctx, sqlc.GetRestaurantBySlugParams{TenantID: tenantID, Slug: slug})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("restaurant")
	}
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// GetProduct returns an available product by ID for public view.
func (s *Service) GetProduct(ctx context.Context, id uuid.UUID) (*sqlc.Product, error) {
	p, err := s.q.GetProductByIDPublic(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("product")
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}
