package hub

import (
	"context"

	"github.com/google/uuid"
	"github.com/munchies/platform/backend/internal/db/sqlc"
)

// Repository provides hub data access.
type Repository struct {
	q *sqlc.Queries
}

// NewRepository creates a new hub repository.
func NewRepository(q *sqlc.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) CreateHub(ctx context.Context, arg sqlc.CreateHubParams) (*sqlc.Hub, error) {
	h, err := r.q.CreateHub(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *Repository) GetHubByID(ctx context.Context, id, tenantID uuid.UUID) (*sqlc.Hub, error) {
	h, err := r.q.GetHubByID(ctx, sqlc.GetHubByIDParams{ID: id, TenantID: tenantID})
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *Repository) ListHubsByTenant(ctx context.Context, tenantID uuid.UUID) ([]sqlc.Hub, error) {
	return r.q.ListHubsByTenant(ctx, tenantID)
}

func (r *Repository) UpdateHub(ctx context.Context, arg sqlc.UpdateHubParams) (*sqlc.Hub, error) {
	h, err := r.q.UpdateHub(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *Repository) DeleteHub(ctx context.Context, id, tenantID uuid.UUID) error {
	return r.q.DeleteHub(ctx, sqlc.DeleteHubParams{ID: id, TenantID: tenantID})
}

func (r *Repository) CreateHubArea(ctx context.Context, arg sqlc.CreateHubAreaParams) (*sqlc.HubCoverageArea, error) {
	a, err := r.q.CreateHubArea(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repository) GetHubAreaByID(ctx context.Context, id uuid.UUID) (*sqlc.HubCoverageArea, error) {
	a, err := r.q.GetHubAreaByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repository) ListHubAreas(ctx context.Context, hubID uuid.UUID) ([]sqlc.HubCoverageArea, error) {
	return r.q.ListHubAreas(ctx, hubID)
}

func (r *Repository) UpdateHubArea(ctx context.Context, arg sqlc.UpdateHubAreaParams) (*sqlc.HubCoverageArea, error) {
	a, err := r.q.UpdateHubArea(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repository) DeleteHubArea(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteHubArea(ctx, id)
}

func (r *Repository) GetDeliveryZoneConfig(ctx context.Context, tenantID uuid.UUID) (*sqlc.DeliveryZoneConfig, error) {
	c, err := r.q.GetDeliveryZoneConfig(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) UpsertDeliveryZoneConfig(ctx context.Context, arg sqlc.UpsertDeliveryZoneConfigParams) (*sqlc.DeliveryZoneConfig, error) {
	c, err := r.q.UpsertDeliveryZoneConfig(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
