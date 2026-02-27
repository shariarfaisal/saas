package storefront

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/munchies/platform/backend/internal/modules/tenant"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/pagination"
	"github.com/munchies/platform/backend/internal/pkg/respond"
)

// Handler handles public storefront HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new storefront handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetConfig handles GET /api/v1/storefront/config
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"tenant_id":       t.ID,
		"name":            t.Name,
		"logo_url":        t.LogoUrl,
		"primary_color":   t.PrimaryColor,
		"secondary_color": t.SecondaryColor,
		"currency":        t.Currency,
		"timezone":        t.Timezone,
	})
}

// ListAreas handles GET /api/v1/storefront/areas
func (h *Handler) ListAreas(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	areas, err := h.svc.ListAreas(r.Context(), t.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, areas)
}

// ListRestaurants handles GET /api/v1/storefront/restaurants
func (h *Handler) ListRestaurants(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	areaSlug := r.URL.Query().Get("area")
	page, perPage := parsePagination(r)

	items, meta, err := h.svc.ListRestaurants(r.Context(), t.ID, areaSlug, page, perPage)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: items, Meta: meta})
}

// GetRestaurant handles GET /api/v1/restaurants/{slug}
func (h *Handler) GetRestaurant(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	slug := chi.URLParam(r, "slug")
	res, err := h.svc.GetRestaurant(r.Context(), t.ID, slug)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, res)
}

// GetProduct handles GET /api/v1/products/{id}
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid product id"))
		return
	}
	prod, err := h.svc.GetProduct(r.Context(), id)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, prod)
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
