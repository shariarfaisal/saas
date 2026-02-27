package restaurant

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
	"github.com/munchies/platform/backend/internal/pkg/slug"
)

// Service implements restaurant business logic.
type Service struct {
	repo *Repository
}

// NewService creates a new restaurant service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateRestaurantRequest holds fields for creating a restaurant.
type CreateRestaurantRequest struct {
	HubID               pgtype.UUID
	OwnerID             pgtype.UUID
	Name                string
	Type                sqlc.RestaurantType
	Description         *string
	ShortDescription    *string
	BannerImageUrl      *string
	LogoUrl             *string
	GalleryUrls         []string
	Phone               *string
	Email               *string
	AddressLine1        *string
	AddressLine2        *string
	Area                *string
	City                string
	Cuisines            []string
	Tags                []string
	CommissionRate      pgtype.Numeric
	VatRate             pgtype.Numeric
	IsVatInclusive      bool
	MinOrderAmount      pgtype.Numeric
	AvgPrepTimeMinutes  int32
	MaxConcurrentOrders int32
	AutoAcceptOrders    bool
	OrderPrefix         *string
	IsAvailable         bool
	IsFeatured          bool
	IsActive            bool
	SortOrder           int32
}

// CreateRestaurant creates a new restaurant.
func (s *Service) CreateRestaurant(ctx context.Context, tenantID uuid.UUID, req CreateRestaurantRequest) (*sqlc.Restaurant, error) {
	resSlug := slug.Generate(req.Name)
	return s.repo.CreateRestaurant(ctx, sqlc.CreateRestaurantParams{
		TenantID:            tenantID,
		HubID:               req.HubID,
		OwnerID:             req.OwnerID,
		Name:                req.Name,
		Slug:                resSlug,
		Type:                req.Type,
		Description:         nullString(req.Description),
		ShortDescription:    nullString(req.ShortDescription),
		BannerImageUrl:      nullString(req.BannerImageUrl),
		LogoUrl:             nullString(req.LogoUrl),
		GalleryUrls:         req.GalleryUrls,
		Phone:               nullString(req.Phone),
		Email:               nullString(req.Email),
		AddressLine1:        nullString(req.AddressLine1),
		AddressLine2:        nullString(req.AddressLine2),
		Area:                nullString(req.Area),
		City:                req.City,
		Cuisines:            req.Cuisines,
		Tags:                req.Tags,
		CommissionRate:      req.CommissionRate,
		VatRate:             req.VatRate,
		IsVatInclusive:      req.IsVatInclusive,
		MinOrderAmount:      req.MinOrderAmount,
		AvgPrepTimeMinutes:  req.AvgPrepTimeMinutes,
		MaxConcurrentOrders: req.MaxConcurrentOrders,
		AutoAcceptOrders:    req.AutoAcceptOrders,
		OrderPrefix:         nullString(req.OrderPrefix),
		IsAvailable:         req.IsAvailable,
		IsFeatured:          req.IsFeatured,
		IsActive:            req.IsActive,
		SortOrder:           req.SortOrder,
	})
}

// GetRestaurant returns a restaurant by ID.
func (s *Service) GetRestaurant(ctx context.Context, id, tenantID uuid.UUID) (*sqlc.Restaurant, error) {
	res, err := s.repo.GetRestaurantByID(ctx, id, tenantID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("restaurant")
	}
	return res, err
}

// GetRestaurantBySlug returns a restaurant by slug.
func (s *Service) GetRestaurantBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*sqlc.Restaurant, error) {
	res, err := s.repo.GetRestaurantBySlug(ctx, tenantID, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("restaurant")
	}
	return res, err
}

// ListRestaurants returns paginated restaurants for a tenant.
func (s *Service) ListRestaurants(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]sqlc.Restaurant, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)
	total, err := s.repo.CountRestaurantsByTenant(ctx, tenantID)
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("count restaurants", err)
	}
	items, err := s.repo.ListRestaurantsByTenant(ctx, tenantID, int32(limit), int32(offset))
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list restaurants", err)
	}
	meta := pagination.NewMeta(total, limit, "")
	return items, meta, nil
}

// UpdateRestaurantRequest holds updateable restaurant fields.
type UpdateRestaurantRequest struct {
	Name               *string
	Description        *string
	ShortDescription   *string
	BannerImageUrl     *string
	LogoUrl            *string
	Phone              *string
	Email              *string
	AddressLine1       *string
	Area               *string
	City               *string
	Cuisines           []string
	Tags               []string
	MinOrderAmount     pgtype.Numeric
	AvgPrepTimeMinutes *int32
	AutoAcceptOrders   *bool
	IsFeatured         *bool
	SortOrder          *int32
}

// UpdateRestaurant updates a restaurant.
func (s *Service) UpdateRestaurant(ctx context.Context, id, tenantID uuid.UUID, req UpdateRestaurantRequest) (*sqlc.Restaurant, error) {
	res, err := s.repo.UpdateRestaurant(ctx, sqlc.UpdateRestaurantParams{
		ID:                 id,
		TenantID:           tenantID,
		Name:               nullString(req.Name),
		Description:        nullString(req.Description),
		ShortDescription:   nullString(req.ShortDescription),
		BannerImageUrl:     nullString(req.BannerImageUrl),
		LogoUrl:            nullString(req.LogoUrl),
		Phone:              nullString(req.Phone),
		Email:              nullString(req.Email),
		AddressLine1:       nullString(req.AddressLine1),
		Area:               nullString(req.Area),
		City:               nullString(req.City),
		Cuisines:           req.Cuisines,
		Tags:               req.Tags,
		MinOrderAmount:     req.MinOrderAmount,
		AvgPrepTimeMinutes: req.AvgPrepTimeMinutes,
		AutoAcceptOrders:   req.AutoAcceptOrders,
		IsFeatured:         req.IsFeatured,
		SortOrder:          req.SortOrder,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("restaurant")
	}
	return res, err
}

// UpdateAvailability sets the is_available flag on a restaurant.
func (s *Service) UpdateAvailability(ctx context.Context, id, tenantID uuid.UUID, available bool) (*sqlc.Restaurant, error) {
	res, err := s.repo.UpdateRestaurantAvailability(ctx, id, available, tenantID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("restaurant")
	}
	return res, err
}

// DeleteRestaurant soft-deletes a restaurant.
func (s *Service) DeleteRestaurant(ctx context.Context, id, tenantID uuid.UUID) error {
	return s.repo.DeleteRestaurant(ctx, id, tenantID)
}

// UpsertOperatingHour upserts a single operating hour record.
func (s *Service) UpsertOperatingHour(ctx context.Context, arg sqlc.UpsertOperatingHourParams) (*sqlc.RestaurantOperatingHour, error) {
	return s.repo.UpsertOperatingHour(ctx, arg)
}

// ListOperatingHours returns the operating hours for a restaurant.
func (s *Service) ListOperatingHours(ctx context.Context, restaurantID uuid.UUID) ([]sqlc.RestaurantOperatingHour, error) {
	return s.repo.ListOperatingHours(ctx, restaurantID)
}

func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}
