package content

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/pagination"
)

// Service implements content management business logic.
type Service struct {
	q *sqlc.Queries
}

// NewService creates a new content service.
func NewService(q *sqlc.Queries) *Service {
	return &Service{q: q}
}

// --- Banners ---

// CreateBanner creates a new banner.
func (s *Service) CreateBanner(ctx context.Context, tenantID uuid.UUID, req CreateBannerRequest) (*sqlc.Banner, error) {
	hubIDs := req.HubIDs
	if hubIDs == nil {
		hubIDs = []uuid.UUID{}
	}
	banner, err := s.q.CreateBanner(ctx, sqlc.CreateBannerParams{
		TenantID:       tenantID,
		Title:          req.Title,
		Subtitle:       toNullString(req.Subtitle),
		ImageUrl:       req.ImageURL,
		MobileImageUrl: toNullString(req.MobileImageURL),
		LinkType:       toNullLinkTargetType(req.LinkType),
		LinkValue:      toNullString(req.LinkValue),
		Platform:       req.Platform,
		SortOrder:      req.SortOrder,
		IsActive:       req.IsActive,
		HubIds:         hubIDs,
		StartsAt:       toPgTimestamptz(req.StartsAt),
		EndsAt:         toPgTimestamptz(req.EndsAt),
	})
	if err != nil {
		return nil, err
	}
	return &banner, nil
}

// GetBanner gets a banner by ID.
func (s *Service) GetBanner(ctx context.Context, tenantID, bannerID uuid.UUID) (*sqlc.Banner, error) {
	banner, err := s.q.GetBannerByID(ctx, sqlc.GetBannerByIDParams{
		ID:       bannerID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("banner")
	}
	if err != nil {
		return nil, err
	}
	return &banner, nil
}

// UpdateBanner updates a banner.
func (s *Service) UpdateBanner(ctx context.Context, tenantID, bannerID uuid.UUID, req CreateBannerRequest) (*sqlc.Banner, error) {
	hubIDs := req.HubIDs
	if hubIDs == nil {
		hubIDs = []uuid.UUID{}
	}
	banner, err := s.q.UpdateBanner(ctx, sqlc.UpdateBannerParams{
		ID:             bannerID,
		TenantID:       tenantID,
		Title:          req.Title,
		Subtitle:       toNullString(req.Subtitle),
		ImageUrl:       req.ImageURL,
		MobileImageUrl: toNullString(req.MobileImageURL),
		LinkType:       toNullLinkTargetType(req.LinkType),
		LinkValue:      toNullString(req.LinkValue),
		Platform:       req.Platform,
		SortOrder:      req.SortOrder,
		IsActive:       req.IsActive,
		HubIds:         hubIDs,
		StartsAt:       toPgTimestamptz(req.StartsAt),
		EndsAt:         toPgTimestamptz(req.EndsAt),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("banner")
	}
	if err != nil {
		return nil, err
	}
	return &banner, nil
}

// DeleteBanner deletes a banner.
func (s *Service) DeleteBanner(ctx context.Context, tenantID, bannerID uuid.UUID) error {
	return s.q.DeleteBanner(ctx, sqlc.DeleteBannerParams{
		ID:       bannerID,
		TenantID: tenantID,
	})
}

// ListBanners lists all banners for a tenant.
func (s *Service) ListBanners(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]sqlc.Banner, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)
	total, err := s.q.CountBannersByTenant(ctx, tenantID)
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("count banners", err)
	}
	items, err := s.q.ListBannersByTenant(ctx, sqlc.ListBannersByTenantParams{
		TenantID: tenantID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list banners", err)
	}
	return items, pagination.NewMeta(total, limit, ""), nil
}

// ListActiveBanners lists active banners for storefront.
func (s *Service) ListActiveBanners(ctx context.Context, tenantID uuid.UUID) ([]sqlc.Banner, error) {
	return s.q.ListActiveBanners(ctx, tenantID)
}

// --- Stories ---

// CreateStory creates a new story.
func (s *Service) CreateStory(ctx context.Context, tenantID uuid.UUID, req CreateStoryRequest) (*sqlc.Story, error) {
	story, err := s.q.CreateStory(ctx, sqlc.CreateStoryParams{
		TenantID:     tenantID,
		RestaurantID: toPgUUIDPtr(req.RestaurantID),
		Title:        toNullString(req.Title),
		MediaUrl:     req.MediaURL,
		MediaType:    req.MediaType,
		ThumbnailUrl: toNullString(req.ThumbnailURL),
		LinkType:     toNullLinkTargetType(req.LinkType),
		LinkValue:    toNullString(req.LinkValue),
		ExpiresAt:    req.ExpiresAt,
		SortOrder:    req.SortOrder,
		IsActive:     req.IsActive,
	})
	if err != nil {
		return nil, err
	}
	return &story, nil
}

// DeleteStory deletes a story.
func (s *Service) DeleteStory(ctx context.Context, tenantID, storyID uuid.UUID) error {
	return s.q.DeleteStory(ctx, sqlc.DeleteStoryParams{
		ID:       storyID,
		TenantID: tenantID,
	})
}

// ListStories lists all stories for a tenant.
func (s *Service) ListStories(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]sqlc.Story, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)
	total, err := s.q.CountStoriesByTenant(ctx, tenantID)
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("count stories", err)
	}
	items, err := s.q.ListStoriesByTenant(ctx, sqlc.ListStoriesByTenantParams{
		TenantID: tenantID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list stories", err)
	}
	return items, pagination.NewMeta(total, limit, ""), nil
}

// ListActiveStories lists active stories for storefront.
func (s *Service) ListActiveStories(ctx context.Context, tenantID uuid.UUID) ([]sqlc.Story, error) {
	return s.q.ListActiveStories(ctx, tenantID)
}

// --- Sections ---

// GetSection gets a section by ID.
func (s *Service) GetSection(ctx context.Context, tenantID, sectionID uuid.UUID) (*sqlc.HomepageSection, error) {
	section, err := s.q.GetSectionByID(ctx, sqlc.GetSectionByIDParams{
		ID:       sectionID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("section")
	}
	if err != nil {
		return nil, err
	}
	return &section, nil
}

// UpdateSection updates a homepage section.
func (s *Service) UpdateSection(ctx context.Context, tenantID, sectionID uuid.UUID, req UpdateSectionRequest) (*sqlc.HomepageSection, error) {
	itemIDs := req.ItemIDs
	if itemIDs == nil {
		itemIDs = []uuid.UUID{}
	}
	hubIDs := req.HubIDs
	if hubIDs == nil {
		hubIDs = []uuid.UUID{}
	}
	filterRule, _ := json.Marshal(req.FilterRule)

	section, err := s.q.UpdateSection(ctx, sqlc.UpdateSectionParams{
		ID:          sectionID,
		TenantID:    tenantID,
		Title:       req.Title,
		Subtitle:    toNullString(req.Subtitle),
		ContentType: req.ContentType,
		ItemIds:     itemIDs,
		FilterRule:  filterRule,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
		HubIds:      hubIDs,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("section")
	}
	if err != nil {
		return nil, err
	}
	return &section, nil
}

// ListSections lists all sections for a tenant.
func (s *Service) ListSections(ctx context.Context, tenantID uuid.UUID) ([]sqlc.HomepageSection, error) {
	return s.q.ListSectionsByTenant(ctx, tenantID)
}

// ListActiveSections lists active sections for storefront.
func (s *Service) ListActiveSections(ctx context.Context, tenantID uuid.UUID) ([]sqlc.HomepageSection, error) {
	return s.q.ListActiveSections(ctx, tenantID)
}

// Request types

type CreateBannerRequest struct {
	Title          string
	Subtitle       *string
	ImageURL       string
	MobileImageURL *string
	LinkType       *sqlc.LinkTargetType
	LinkValue      *string
	Platform       string
	SortOrder      int32
	IsActive       bool
	HubIDs         []uuid.UUID
	StartsAt       *time.Time
	EndsAt         *time.Time
}

type CreateStoryRequest struct {
	RestaurantID *uuid.UUID
	Title        *string
	MediaURL     string
	MediaType    sqlc.MediaType
	ThumbnailURL *string
	LinkType     *sqlc.LinkTargetType
	LinkValue    *string
	ExpiresAt    time.Time
	SortOrder    int32
	IsActive     bool
}

type UpdateSectionRequest struct {
	Title       string
	Subtitle    *string
	ContentType string
	ItemIDs     []uuid.UUID
	FilterRule  interface{}
	SortOrder   int32
	IsActive    bool
	HubIDs      []uuid.UUID
}

func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func toNullLinkTargetType(lt *sqlc.LinkTargetType) sqlc.NullLinkTargetType {
	if lt == nil {
		return sqlc.NullLinkTargetType{}
	}
	return sqlc.NullLinkTargetType{LinkTargetType: *lt, Valid: true}
}

func toPgUUIDPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func toPgTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}
