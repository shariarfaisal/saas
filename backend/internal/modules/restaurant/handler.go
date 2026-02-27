package restaurant

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/modules/auth"
	"github.com/munchies/platform/backend/internal/modules/tenant"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/pagination"
	"github.com/munchies/platform/backend/internal/pkg/respond"
	"strconv"
)

// Handler handles restaurant HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new restaurant handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListRestaurants handles GET /partner/restaurants
func (h *Handler) ListRestaurants(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	page, perPage := parsePagination(r)
	items, meta, err := h.svc.ListRestaurants(r.Context(), t.ID, page, perPage)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: items, Meta: meta})
}

// CreateRestaurant handles POST /partner/restaurants
func (h *Handler) CreateRestaurant(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}

	var req struct {
		HubID               *string            `json:"hub_id"`
		Name                string             `json:"name"`
		Type                sqlc.RestaurantType `json:"type"`
		Description         *string            `json:"description"`
		ShortDescription    *string            `json:"short_description"`
		BannerImageUrl      *string            `json:"banner_image_url"`
		LogoUrl             *string            `json:"logo_url"`
		GalleryUrls         []string           `json:"gallery_urls"`
		Phone               *string            `json:"phone"`
		Email               *string            `json:"email"`
		AddressLine1        *string            `json:"address_line1"`
		AddressLine2        *string            `json:"address_line2"`
		Area                *string            `json:"area"`
		City                string             `json:"city"`
		Cuisines            []string           `json:"cuisines"`
		Tags                []string           `json:"tags"`
		CommissionRate      pgtype.Numeric     `json:"commission_rate"`
		VatRate             pgtype.Numeric     `json:"vat_rate"`
		IsVatInclusive      bool               `json:"is_vat_inclusive"`
		MinOrderAmount      pgtype.Numeric     `json:"min_order_amount"`
		AvgPrepTimeMinutes  int32              `json:"avg_prep_time_minutes"`
		MaxConcurrentOrders int32              `json:"max_concurrent_orders"`
		AutoAcceptOrders    bool               `json:"auto_accept_orders"`
		OrderPrefix         *string            `json:"order_prefix"`
		IsAvailable         bool               `json:"is_available"`
		IsFeatured          bool               `json:"is_featured"`
		IsActive            bool               `json:"is_active"`
		SortOrder           int32              `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.Name == "" {
		respond.Error(w, apperror.BadRequest("name is required"))
		return
	}
	if req.Type == "" {
		req.Type = sqlc.RestaurantTypeRestaurant
	}
	if req.City == "" {
		req.City = "Dhaka"
	}
	if req.AvgPrepTimeMinutes == 0 {
		req.AvgPrepTimeMinutes = 30
	}
	if req.MaxConcurrentOrders == 0 {
		req.MaxConcurrentOrders = 10
	}

	hubID := pgtype.UUID{}
	if req.HubID != nil {
		parsed, err := uuid.Parse(*req.HubID)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid hub_id"))
			return
		}
		hubID = pgtype.UUID{Bytes: parsed, Valid: true}
	}
	ownerID := pgtype.UUID{Bytes: u.ID, Valid: true}

	res, err := h.svc.CreateRestaurant(r.Context(), t.ID, CreateRestaurantRequest{
		HubID:               hubID,
		OwnerID:             ownerID,
		Name:                req.Name,
		Type:                req.Type,
		Description:         req.Description,
		ShortDescription:    req.ShortDescription,
		BannerImageUrl:      req.BannerImageUrl,
		LogoUrl:             req.LogoUrl,
		GalleryUrls:         req.GalleryUrls,
		Phone:               req.Phone,
		Email:               req.Email,
		AddressLine1:        req.AddressLine1,
		AddressLine2:        req.AddressLine2,
		Area:                req.Area,
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
		OrderPrefix:         req.OrderPrefix,
		IsAvailable:         req.IsAvailable,
		IsFeatured:          req.IsFeatured,
		IsActive:            req.IsActive,
		SortOrder:           req.SortOrder,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusCreated, res)
}

// GetRestaurant handles GET /partner/restaurants/{id}
func (h *Handler) GetRestaurant(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid restaurant id"))
		return
	}
	res, err := h.svc.GetRestaurant(r.Context(), id, t.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, res)
}

// UpdateRestaurant handles PUT /partner/restaurants/{id}
func (h *Handler) UpdateRestaurant(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid restaurant id"))
		return
	}

	var req struct {
		Name               *string        `json:"name"`
		Description        *string        `json:"description"`
		ShortDescription   *string        `json:"short_description"`
		BannerImageUrl     *string        `json:"banner_image_url"`
		LogoUrl            *string        `json:"logo_url"`
		Phone              *string        `json:"phone"`
		Email              *string        `json:"email"`
		AddressLine1       *string        `json:"address_line1"`
		Area               *string        `json:"area"`
		City               *string        `json:"city"`
		Cuisines           []string       `json:"cuisines"`
		Tags               []string       `json:"tags"`
		MinOrderAmount     pgtype.Numeric `json:"min_order_amount"`
		AvgPrepTimeMinutes *int32         `json:"avg_prep_time_minutes"`
		AutoAcceptOrders   *bool          `json:"auto_accept_orders"`
		IsFeatured         *bool          `json:"is_featured"`
		SortOrder          *int32         `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	res, err := h.svc.UpdateRestaurant(r.Context(), id, t.ID, UpdateRestaurantRequest{
		Name:               req.Name,
		Description:        req.Description,
		ShortDescription:   req.ShortDescription,
		BannerImageUrl:     req.BannerImageUrl,
		LogoUrl:            req.LogoUrl,
		Phone:              req.Phone,
		Email:              req.Email,
		AddressLine1:       req.AddressLine1,
		Area:               req.Area,
		City:               req.City,
		Cuisines:           req.Cuisines,
		Tags:               req.Tags,
		MinOrderAmount:     req.MinOrderAmount,
		AvgPrepTimeMinutes: req.AvgPrepTimeMinutes,
		AutoAcceptOrders:   req.AutoAcceptOrders,
		IsFeatured:         req.IsFeatured,
		SortOrder:          req.SortOrder,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, res)
}

// UpdateAvailability handles PATCH /partner/restaurants/{id}/availability
func (h *Handler) UpdateAvailability(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid restaurant id"))
		return
	}

	var req struct {
		IsAvailable bool `json:"is_available"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	res, err := h.svc.UpdateAvailability(r.Context(), id, t.ID, req.IsAvailable)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, res)
}

// DeleteRestaurant handles DELETE /partner/restaurants/{id}
func (h *Handler) DeleteRestaurant(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid restaurant id"))
		return
	}
	if err := h.svc.DeleteRestaurant(r.Context(), id, t.ID); err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetOperatingHours handles GET /partner/restaurants/{id}/hours
func (h *Handler) GetOperatingHours(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid restaurant id"))
		return
	}
	hours, err := h.svc.ListOperatingHours(r.Context(), id)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, hours)
}

// UpsertOperatingHours handles PUT /partner/restaurants/{id}/hours
func (h *Handler) UpsertOperatingHours(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	restaurantID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid restaurant id"))
		return
	}

	var req []struct {
		DayOfWeek int16       `json:"day_of_week"`
		OpenTime  pgtype.Time `json:"open_time"`
		CloseTime pgtype.Time `json:"close_time"`
		IsClosed  bool        `json:"is_closed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	results := make([]sqlc.RestaurantOperatingHour, 0, len(req))
	for _, item := range req {
		oh, err := h.svc.UpsertOperatingHour(r.Context(), sqlc.UpsertOperatingHourParams{
			RestaurantID: restaurantID,
			TenantID:     t.ID,
			DayOfWeek:    item.DayOfWeek,
			OpenTime:     item.OpenTime,
			CloseTime:    item.CloseTime,
			IsClosed:     item.IsClosed,
		})
		if err != nil {
			respond.Error(w, toAppError(err))
			return
		}
		results = append(results, *oh)
	}
	respond.JSON(w, http.StatusOK, results)
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
