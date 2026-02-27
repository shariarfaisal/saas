package content

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/modules/tenant"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/pagination"
	"github.com/munchies/platform/backend/internal/pkg/respond"
)

// Handler handles content HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new content handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// --- Banners ---

// ListBanners handles GET /partner/content/banners
func (h *Handler) ListBanners(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	page, perPage := parsePagination(r)
	items, meta, err := h.svc.ListBanners(r.Context(), t.ID, page, perPage)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: items, Meta: meta})
}

// CreateBanner handles POST /partner/content/banners
func (h *Handler) CreateBanner(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	var req struct {
		Title          string      `json:"title"`
		Subtitle       *string     `json:"subtitle"`
		ImageURL       string      `json:"image_url"`
		MobileImageURL *string     `json:"mobile_image_url"`
		LinkType       *string     `json:"link_type"`
		LinkValue      *string     `json:"link_value"`
		Platform       string      `json:"platform"`
		SortOrder      int32       `json:"sort_order"`
		IsActive       bool        `json:"is_active"`
		HubIDs         []uuid.UUID `json:"hub_ids"`
		StartsAt       *time.Time  `json:"starts_at"`
		EndsAt         *time.Time  `json:"ends_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.Title == "" || req.ImageURL == "" {
		respond.Error(w, apperror.BadRequest("title and image_url are required"))
		return
	}
	if req.Platform == "" {
		req.Platform = "all"
	}

	var linkType *sqlc.LinkTargetType
	if req.LinkType != nil {
		lt := sqlc.LinkTargetType(*req.LinkType)
		linkType = &lt
	}

	banner, err := h.svc.CreateBanner(r.Context(), t.ID, CreateBannerRequest{
		Title:          req.Title,
		Subtitle:       req.Subtitle,
		ImageURL:       req.ImageURL,
		MobileImageURL: req.MobileImageURL,
		LinkType:       linkType,
		LinkValue:      req.LinkValue,
		Platform:       req.Platform,
		SortOrder:      req.SortOrder,
		IsActive:       req.IsActive,
		HubIDs:         req.HubIDs,
		StartsAt:       req.StartsAt,
		EndsAt:         req.EndsAt,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusCreated, banner)
}

// UpdateBanner handles PUT /partner/content/banners/:id
func (h *Handler) UpdateBanner(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	bannerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid banner id"))
		return
	}
	var req struct {
		Title          string      `json:"title"`
		Subtitle       *string     `json:"subtitle"`
		ImageURL       string      `json:"image_url"`
		MobileImageURL *string     `json:"mobile_image_url"`
		LinkType       *string     `json:"link_type"`
		LinkValue      *string     `json:"link_value"`
		Platform       string      `json:"platform"`
		SortOrder      int32       `json:"sort_order"`
		IsActive       bool        `json:"is_active"`
		HubIDs         []uuid.UUID `json:"hub_ids"`
		StartsAt       *time.Time  `json:"starts_at"`
		EndsAt         *time.Time  `json:"ends_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	var linkType *sqlc.LinkTargetType
	if req.LinkType != nil {
		lt := sqlc.LinkTargetType(*req.LinkType)
		linkType = &lt
	}

	banner, err := h.svc.UpdateBanner(r.Context(), t.ID, bannerID, CreateBannerRequest{
		Title:          req.Title,
		Subtitle:       req.Subtitle,
		ImageURL:       req.ImageURL,
		MobileImageURL: req.MobileImageURL,
		LinkType:       linkType,
		LinkValue:      req.LinkValue,
		Platform:       req.Platform,
		SortOrder:      req.SortOrder,
		IsActive:       req.IsActive,
		HubIDs:         req.HubIDs,
		StartsAt:       req.StartsAt,
		EndsAt:         req.EndsAt,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, banner)
}

// DeleteBanner handles DELETE /partner/content/banners/:id
func (h *Handler) DeleteBanner(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	bannerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid banner id"))
		return
	}
	if err := h.svc.DeleteBanner(r.Context(), t.ID, bannerID); err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Stories ---

// ListStories handles GET /partner/content/stories
func (h *Handler) ListStories(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	page, perPage := parsePagination(r)
	items, meta, err := h.svc.ListStories(r.Context(), t.ID, page, perPage)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: items, Meta: meta})
}

// CreateStory handles POST /partner/content/stories
func (h *Handler) CreateStory(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	var req struct {
		RestaurantID *uuid.UUID `json:"restaurant_id"`
		Title        *string    `json:"title"`
		MediaURL     string     `json:"media_url"`
		MediaType    string     `json:"media_type"`
		ThumbnailURL *string    `json:"thumbnail_url"`
		LinkType     *string    `json:"link_type"`
		LinkValue    *string    `json:"link_value"`
		ExpiresAt    time.Time  `json:"expires_at"`
		SortOrder    int32      `json:"sort_order"`
		IsActive     bool       `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.MediaURL == "" {
		respond.Error(w, apperror.BadRequest("media_url is required"))
		return
	}

	var linkType *sqlc.LinkTargetType
	if req.LinkType != nil {
		lt := sqlc.LinkTargetType(*req.LinkType)
		linkType = &lt
	}

	story, err := h.svc.CreateStory(r.Context(), t.ID, CreateStoryRequest{
		RestaurantID: req.RestaurantID,
		Title:        req.Title,
		MediaURL:     req.MediaURL,
		MediaType:    sqlc.MediaType(req.MediaType),
		ThumbnailURL: req.ThumbnailURL,
		LinkType:     linkType,
		LinkValue:    req.LinkValue,
		ExpiresAt:    req.ExpiresAt,
		SortOrder:    req.SortOrder,
		IsActive:     req.IsActive,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusCreated, story)
}

// DeleteStory handles DELETE /partner/content/stories/:id
func (h *Handler) DeleteStory(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	storyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid story id"))
		return
	}
	if err := h.svc.DeleteStory(r.Context(), t.ID, storyID); err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Sections ---

// ListSections handles GET /partner/content/sections
func (h *Handler) ListSections(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	sections, err := h.svc.ListSections(r.Context(), t.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, sections)
}

// UpdateSection handles PUT /partner/content/sections/:id
func (h *Handler) UpdateSection(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	sectionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid section id"))
		return
	}
	var req struct {
		Title       string      `json:"title"`
		Subtitle    *string     `json:"subtitle"`
		ContentType string      `json:"content_type"`
		ItemIDs     []uuid.UUID `json:"item_ids"`
		FilterRule  interface{} `json:"filter_rule"`
		SortOrder   int32       `json:"sort_order"`
		IsActive    bool        `json:"is_active"`
		HubIDs      []uuid.UUID `json:"hub_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	section, err := h.svc.UpdateSection(r.Context(), t.ID, sectionID, UpdateSectionRequest{
		Title:       req.Title,
		Subtitle:    req.Subtitle,
		ContentType: req.ContentType,
		ItemIDs:     req.ItemIDs,
		FilterRule:  req.FilterRule,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
		HubIDs:      req.HubIDs,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, section)
}

// --- Storefront ---

// StorefrontBanners handles GET /api/v1/storefront/banners
func (h *Handler) StorefrontBanners(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	banners, err := h.svc.ListActiveBanners(r.Context(), t.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, banners)
}

// StorefrontStories handles GET /api/v1/storefront/stories
func (h *Handler) StorefrontStories(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	stories, err := h.svc.ListActiveStories(r.Context(), t.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, stories)
}

// StorefrontSections handles GET /api/v1/storefront/sections
func (h *Handler) StorefrontSections(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	sections, err := h.svc.ListActiveSections(r.Context(), t.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, sections)
}

func parsePagination(r *http.Request) (page, perPage int) {
	q := r.URL.Query()
	page, _ = strconv.Atoi(q.Get("page"))
	perPage, _ = strconv.Atoi(q.Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = pagination.DefaultPageSize
	}
	return page, perPage
}

func toAppError(err error) *apperror.AppError {
	if e, ok := err.(*apperror.AppError); ok {
		return e
	}
	return apperror.Internal("unexpected error", err)
}
