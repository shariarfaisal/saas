package restaurant

import (
	"context"

	"github.com/google/uuid"
	"github.com/munchies/platform/backend/internal/db/sqlc"
)

// Repository provides restaurant data access.
type Repository struct {
	q *sqlc.Queries
}

// NewRepository creates a new restaurant repository.
func NewRepository(q *sqlc.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) CreateRestaurant(ctx context.Context, arg sqlc.CreateRestaurantParams) (*sqlc.Restaurant, error) {
	res, err := r.q.CreateRestaurant(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *Repository) GetRestaurantByID(ctx context.Context, id, tenantID uuid.UUID) (*sqlc.Restaurant, error) {
	res, err := r.q.GetRestaurantByID(ctx, sqlc.GetRestaurantByIDParams{ID: id, TenantID: tenantID})
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *Repository) GetRestaurantBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*sqlc.Restaurant, error) {
	res, err := r.q.GetRestaurantBySlug(ctx, sqlc.GetRestaurantBySlugParams{TenantID: tenantID, Slug: slug})
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *Repository) ListRestaurantsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]sqlc.Restaurant, error) {
	return r.q.ListRestaurantsByTenant(ctx, sqlc.ListRestaurantsByTenantParams{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	})
}

func (r *Repository) CountRestaurantsByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return r.q.CountRestaurantsByTenant(ctx, tenantID)
}

func (r *Repository) ListAvailableByHubAndArea(ctx context.Context, tenantID uuid.UUID, areaSlug string, limit, offset int32) ([]sqlc.Restaurant, error) {
	return r.q.ListAvailableByHubAndArea(ctx, sqlc.ListAvailableByHubAndAreaParams{
		TenantID: tenantID,
		Slug:     areaSlug,
		Limit:    limit,
		Offset:   offset,
	})
}

func (r *Repository) UpdateRestaurant(ctx context.Context, arg sqlc.UpdateRestaurantParams) (*sqlc.Restaurant, error) {
	res, err := r.q.UpdateRestaurant(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *Repository) UpdateRestaurantAvailability(ctx context.Context, id uuid.UUID, available bool, tenantID uuid.UUID) (*sqlc.Restaurant, error) {
	res, err := r.q.UpdateRestaurantAvailability(ctx, sqlc.UpdateRestaurantAvailabilityParams{
		ID:          id,
		IsAvailable: available,
		TenantID:    tenantID,
	})
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *Repository) DeleteRestaurant(ctx context.Context, id, tenantID uuid.UUID) error {
	return r.q.DeleteRestaurant(ctx, sqlc.DeleteRestaurantParams{ID: id, TenantID: tenantID})
}

func (r *Repository) UpsertOperatingHour(ctx context.Context, arg sqlc.UpsertOperatingHourParams) (*sqlc.RestaurantOperatingHour, error) {
	oh, err := r.q.UpsertOperatingHour(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &oh, nil
}

func (r *Repository) ListOperatingHours(ctx context.Context, restaurantID uuid.UUID) ([]sqlc.RestaurantOperatingHour, error) {
	return r.q.ListOperatingHours(ctx, restaurantID)
}

func (r *Repository) DeleteOperatingHours(ctx context.Context, restaurantID uuid.UUID) error {
	return r.q.DeleteOperatingHours(ctx, restaurantID)
}
