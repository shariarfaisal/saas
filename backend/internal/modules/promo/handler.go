package promo

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/modules/auth"
	"github.com/munchies/platform/backend/internal/modules/tenant"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/pagination"
	"github.com/munchies/platform/backend/internal/pkg/respond"
	"github.com/shopspring/decimal"
)

// Handler handles promo HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new promo handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListPromos handles GET /partner/promos
func (h *Handler) ListPromos(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	page, perPage := parsePagination(r)
	promos, meta, err := h.svc.ListPromos(r.Context(), t.ID, page, perPage)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: promos, Meta: meta})
}

// CreatePromo handles POST /partner/promos
func (h *Handler) CreatePromo(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	var req struct {
		Code           string   `json:"code"`
		Title          string   `json:"title"`
		Description    string   `json:"description"`
		PromoType      string   `json:"promo_type"`
		DiscountAmount string   `json:"discount_amount"`
		MaxDiscountCap *string  `json:"max_discount_cap"`
		CashbackAmount string   `json:"cashback_amount"`
		FundedBy       string   `json:"funded_by"`
		AppliesTo      string   `json:"applies_to"`
		MinOrderAmount string   `json:"min_order_amount"`
		MaxTotalUses   *int32   `json:"max_total_uses"`
		MaxUsesPerUser int32    `json:"max_uses_per_user"`
		IncludeStores  bool     `json:"include_stores"`
		StartsAt       string   `json:"starts_at"`
		EndsAt         *string  `json:"ends_at"`
		RestaurantIDs  []string `json:"restaurant_ids"`
		CategoryIDs    []string `json:"category_ids"`
		UserIDs        []string `json:"user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	if req.Code == "" || req.Title == "" {
		respond.Error(w, apperror.BadRequest("code and title are required"))
		return
	}

	discountAmt, err := decimal.NewFromString(req.DiscountAmount)
	if err != nil || discountAmt.LessThanOrEqual(decimal.Zero) {
		respond.Error(w, apperror.BadRequest("invalid discount_amount"))
		return
	}

	var maxCap *decimal.Decimal
	if req.MaxDiscountCap != nil {
		v, err := decimal.NewFromString(*req.MaxDiscountCap)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid max_discount_cap"))
			return
		}
		maxCap = &v
	}

	cashback := decimal.Zero
	if req.CashbackAmount != "" {
		cashback, err = decimal.NewFromString(req.CashbackAmount)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid cashback_amount"))
			return
		}
	}

	minOrder := decimal.Zero
	if req.MinOrderAmount != "" {
		minOrder, err = decimal.NewFromString(req.MinOrderAmount)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid min_order_amount"))
			return
		}
	}

	startsAt := time.Now()
	if req.StartsAt != "" {
		startsAt, err = time.Parse(time.RFC3339, req.StartsAt)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid starts_at format, use RFC3339"))
			return
		}
	}

	var endsAt *time.Time
	if req.EndsAt != nil {
		t, err := time.Parse(time.RFC3339, *req.EndsAt)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid ends_at format, use RFC3339"))
			return
		}
		endsAt = &t
	}

	if req.MaxUsesPerUser < 1 {
		req.MaxUsesPerUser = 1
	}

	promoType := sqlc.PromoType(req.PromoType)
	if promoType != sqlc.PromoTypeFixed && promoType != sqlc.PromoTypePercent {
		respond.Error(w, apperror.BadRequest("promo_type must be 'fixed' or 'percent'"))
		return
	}

	fundedBy := sqlc.PromoFunder(req.FundedBy)
	if fundedBy == "" {
		fundedBy = sqlc.PromoFunderPlatform
	}

	appliesTo := sqlc.PromoApplyOn(req.AppliesTo)
	if appliesTo == "" {
		appliesTo = sqlc.PromoApplyOnAllItems
	}

	p, err := h.svc.CreatePromo(r.Context(), CreatePromoRequest{
		TenantID:       t.ID,
		Code:           req.Code,
		Title:          req.Title,
		Description:    req.Description,
		PromoType:      promoType,
		DiscountAmount: discountAmt,
		MaxDiscountCap: maxCap,
		CashbackAmount: cashback,
		FundedBy:       fundedBy,
		AppliesTo:      appliesTo,
		MinOrderAmount: minOrder,
		MaxTotalUses:   req.MaxTotalUses,
		MaxUsesPerUser: req.MaxUsesPerUser,
		IncludeStores:  req.IncludeStores,
		StartsAt:       startsAt,
		EndsAt:         endsAt,
		CreatedBy:      u.ID,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	respond.JSON(w, http.StatusCreated, p)
}

// GetPromo handles GET /partner/promos/{id}
func (h *Handler) GetPromo(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	promoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid promo id"))
		return
	}

	p, err := h.svc.GetPromo(r.Context(), t.ID, promoID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, p)
}

// UpdatePromo handles PUT /partner/promos/{id}
func (h *Handler) UpdatePromo(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	promoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid promo id"))
		return
	}

	var req struct {
		Title          *string `json:"title"`
		Description    *string `json:"description"`
		DiscountAmount *string `json:"discount_amount"`
		MaxDiscountCap *string `json:"max_discount_cap"`
		CashbackAmount *string `json:"cashback_amount"`
		MinOrderAmount *string `json:"min_order_amount"`
		MaxTotalUses   *int32  `json:"max_total_uses"`
		MaxUsesPerUser *int32  `json:"max_uses_per_user"`
		StartsAt       *string `json:"starts_at"`
		EndsAt         *string `json:"ends_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	updateReq := UpdatePromoRequest{
		Title:          req.Title,
		Description:    req.Description,
		MaxTotalUses:   req.MaxTotalUses,
		MaxUsesPerUser: req.MaxUsesPerUser,
	}

	if req.DiscountAmount != nil {
		v, err := decimal.NewFromString(*req.DiscountAmount)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid discount_amount"))
			return
		}
		updateReq.DiscountAmount = &v
	}
	if req.MaxDiscountCap != nil {
		v, err := decimal.NewFromString(*req.MaxDiscountCap)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid max_discount_cap"))
			return
		}
		updateReq.MaxDiscountCap = &v
	}
	if req.CashbackAmount != nil {
		v, err := decimal.NewFromString(*req.CashbackAmount)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid cashback_amount"))
			return
		}
		updateReq.CashbackAmount = &v
	}
	if req.MinOrderAmount != nil {
		v, err := decimal.NewFromString(*req.MinOrderAmount)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid min_order_amount"))
			return
		}
		updateReq.MinOrderAmount = &v
	}
	if req.StartsAt != nil {
		t, err := time.Parse(time.RFC3339, *req.StartsAt)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid starts_at format"))
			return
		}
		updateReq.StartsAt = &t
	}
	if req.EndsAt != nil {
		t, err := time.Parse(time.RFC3339, *req.EndsAt)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid ends_at format"))
			return
		}
		updateReq.EndsAt = &t
	}

	p, err := h.svc.UpdatePromo(r.Context(), t.ID, promoID, updateReq)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, p)
}

// DeactivatePromo handles PATCH /partner/promos/{id}/deactivate
func (h *Handler) DeactivatePromo(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	promoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid promo id"))
		return
	}

	p, err := h.svc.DeactivatePromo(r.Context(), t.ID, promoID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, p)
}

// RegisterRoutes registers promo routes on the given router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListPromos)
	r.Post("/", h.CreatePromo)
	r.Get("/{id}", h.GetPromo)
	r.Put("/{id}", h.UpdatePromo)
	r.Patch("/{id}/deactivate", h.DeactivatePromo)
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
