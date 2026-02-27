package hub

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/slug"
)

// Service implements hub business logic.
type Service struct {
	repo *Repository
}

// NewService creates a new hub service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateHubRequest holds fields for creating a hub.
type CreateHubRequest struct {
	Name         string
	Code         *string
	AddressLine1 *string
	AddressLine2 *string
	City         string
	ContactPhone *string
	ContactEmail *string
	IsActive     bool
	SortOrder    int32
}

// CreateHub creates a new hub for the tenant.
func (s *Service) CreateHub(ctx context.Context, tenantID uuid.UUID, req CreateHubRequest) (*sqlc.Hub, error) {
	return s.repo.CreateHub(ctx, sqlc.CreateHubParams{
		TenantID:     tenantID,
		Name:         req.Name,
		Code:         nullString(req.Code),
		AddressLine1: nullString(req.AddressLine1),
		AddressLine2: nullString(req.AddressLine2),
		City:         req.City,
		ContactPhone: nullString(req.ContactPhone),
		ContactEmail: nullString(req.ContactEmail),
		IsActive:     req.IsActive,
		SortOrder:    req.SortOrder,
	})
}

// GetHub returns a hub by ID.
func (s *Service) GetHub(ctx context.Context, id, tenantID uuid.UUID) (*sqlc.Hub, error) {
	h, err := s.repo.GetHubByID(ctx, id, tenantID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("hub")
	}
	return h, err
}

// ListHubs returns all hubs for a tenant.
func (s *Service) ListHubs(ctx context.Context, tenantID uuid.UUID) ([]sqlc.Hub, error) {
	return s.repo.ListHubsByTenant(ctx, tenantID)
}

// UpdateHubRequest holds updateable hub fields.
type UpdateHubRequest struct {
	Name         *string
	Code         *string
	AddressLine1 *string
	AddressLine2 *string
	City         *string
	ContactPhone *string
	ContactEmail *string
	IsActive     *bool
	SortOrder    *int32
}

// UpdateHub updates a hub.
func (s *Service) UpdateHub(ctx context.Context, id, tenantID uuid.UUID, req UpdateHubRequest) (*sqlc.Hub, error) {
	h, err := s.repo.UpdateHub(ctx, sqlc.UpdateHubParams{
		ID:           id,
		TenantID:     tenantID,
		Name:         nullString(req.Name),
		Code:         nullString(req.Code),
		AddressLine1: nullString(req.AddressLine1),
		AddressLine2: nullString(req.AddressLine2),
		City:         nullString(req.City),
		ContactPhone: nullString(req.ContactPhone),
		ContactEmail: nullString(req.ContactEmail),
		IsActive:     req.IsActive,
		SortOrder:    req.SortOrder,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("hub")
	}
	return h, err
}

// DeleteHub deletes a hub.
func (s *Service) DeleteHub(ctx context.Context, id, tenantID uuid.UUID) error {
	return s.repo.DeleteHub(ctx, id, tenantID)
}

// CreateHubAreaRequest holds fields for creating a hub coverage area.
type CreateHubAreaRequest struct {
	HubID                    uuid.UUID
	TenantID                 uuid.UUID
	Name                     string
	DeliveryCharge           pgtype.Numeric
	MinOrderAmount           pgtype.Numeric
	EstimatedDeliveryMinutes int32
	IsActive                 bool
	SortOrder                int32
}

// CreateHubArea creates a new coverage area for a hub.
func (s *Service) CreateHubArea(ctx context.Context, req CreateHubAreaRequest) (*sqlc.HubCoverageArea, error) {
	slug := slug.Generate(req.Name)
	return s.repo.CreateHubArea(ctx, sqlc.CreateHubAreaParams{
		HubID:                    req.HubID,
		TenantID:                 req.TenantID,
		Name:                     req.Name,
		Slug:                     slug,
		DeliveryCharge:           req.DeliveryCharge,
		MinOrderAmount:           req.MinOrderAmount,
		EstimatedDeliveryMinutes: req.EstimatedDeliveryMinutes,
		IsActive:                 req.IsActive,
		SortOrder:                req.SortOrder,
	})
}

// GetHubArea returns a hub area by ID.
func (s *Service) GetHubArea(ctx context.Context, id uuid.UUID) (*sqlc.HubCoverageArea, error) {
	a, err := s.repo.GetHubAreaByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("hub area")
	}
	return a, err
}

// ListHubAreas returns all coverage areas for a hub.
func (s *Service) ListHubAreas(ctx context.Context, hubID uuid.UUID) ([]sqlc.HubCoverageArea, error) {
	return s.repo.ListHubAreas(ctx, hubID)
}

// UpdateHubAreaRequest holds updateable hub area fields.
type UpdateHubAreaRequest struct {
	Name                     *string
	DeliveryCharge           pgtype.Numeric
	MinOrderAmount           pgtype.Numeric
	EstimatedDeliveryMinutes *int32
	IsActive                 *bool
	SortOrder                *int32
}

// UpdateHubArea updates a hub coverage area.
func (s *Service) UpdateHubArea(ctx context.Context, id uuid.UUID, req UpdateHubAreaRequest) (*sqlc.HubCoverageArea, error) {
	a, err := s.repo.UpdateHubArea(ctx, sqlc.UpdateHubAreaParams{
		ID:                       id,
		Name:                     nullString(req.Name),
		DeliveryCharge:           req.DeliveryCharge,
		MinOrderAmount:           req.MinOrderAmount,
		EstimatedDeliveryMinutes: req.EstimatedDeliveryMinutes,
		IsActive:                 req.IsActive,
		SortOrder:                req.SortOrder,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("hub area")
	}
	return a, err
}

// DeleteHubArea deletes a hub coverage area.
func (s *Service) DeleteHubArea(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteHubArea(ctx, id)
}

// UpsertDeliveryZoneConfigRequest holds fields for upserting the delivery zone config.
type UpsertDeliveryZoneConfigRequest struct {
	Model                 sqlc.DeliveryModel
	DistanceTiers         json.RawMessage
	FreeDeliveryThreshold pgtype.Numeric
}

// UpsertDeliveryZoneConfig upserts the delivery zone config for a tenant.
func (s *Service) UpsertDeliveryZoneConfig(ctx context.Context, tenantID uuid.UUID, req UpsertDeliveryZoneConfigRequest) (*sqlc.DeliveryZoneConfig, error) {
	return s.repo.UpsertDeliveryZoneConfig(ctx, sqlc.UpsertDeliveryZoneConfigParams{
		TenantID:              tenantID,
		Model:                 req.Model,
		DistanceTiers:         req.DistanceTiers,
		FreeDeliveryThreshold: req.FreeDeliveryThreshold,
	})
}

// GetDeliveryZoneConfig returns the delivery zone config for a tenant.
func (s *Service) GetDeliveryZoneConfig(ctx context.Context, tenantID uuid.UUID) (*sqlc.DeliveryZoneConfig, error) {
	c, err := s.repo.GetDeliveryZoneConfig(ctx, tenantID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("delivery zone config")
	}
	return c, err
}

func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}
