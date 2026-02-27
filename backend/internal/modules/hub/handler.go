package hub

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/modules/tenant"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/respond"
)

// Handler handles hub HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new hub handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListHubs handles GET /partner/hubs
func (h *Handler) ListHubs(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	hubs, err := h.svc.ListHubs(r.Context(), t.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, hubs)
}

// CreateHub handles POST /partner/hubs
func (h *Handler) CreateHub(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	var req struct {
		Name         string  `json:"name"`
		Code         *string `json:"code"`
		AddressLine1 *string `json:"address_line1"`
		AddressLine2 *string `json:"address_line2"`
		City         string  `json:"city"`
		ContactPhone *string `json:"contact_phone"`
		ContactEmail *string `json:"contact_email"`
		IsActive     bool    `json:"is_active"`
		SortOrder    int32   `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.Name == "" {
		respond.Error(w, apperror.BadRequest("name is required"))
		return
	}
	if req.City == "" {
		req.City = "Dhaka"
	}

	hub, err := h.svc.CreateHub(r.Context(), t.ID, CreateHubRequest{
		Name:         req.Name,
		Code:         req.Code,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		ContactPhone: req.ContactPhone,
		ContactEmail: req.ContactEmail,
		IsActive:     req.IsActive,
		SortOrder:    req.SortOrder,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusCreated, hub)
}

// GetHub handles GET /partner/hubs/{id}
func (h *Handler) GetHub(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid hub id"))
		return
	}
	hub, err := h.svc.GetHub(r.Context(), id, t.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, hub)
}

// UpdateHub handles PUT /partner/hubs/{id}
func (h *Handler) UpdateHub(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid hub id"))
		return
	}

	var req struct {
		Name         *string `json:"name"`
		Code         *string `json:"code"`
		AddressLine1 *string `json:"address_line1"`
		AddressLine2 *string `json:"address_line2"`
		City         *string `json:"city"`
		ContactPhone *string `json:"contact_phone"`
		ContactEmail *string `json:"contact_email"`
		IsActive     *bool   `json:"is_active"`
		SortOrder    *int32  `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	hub, err := h.svc.UpdateHub(r.Context(), id, t.ID, UpdateHubRequest{
		Name:         req.Name,
		Code:         req.Code,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		ContactPhone: req.ContactPhone,
		ContactEmail: req.ContactEmail,
		IsActive:     req.IsActive,
		SortOrder:    req.SortOrder,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, hub)
}

// DeleteHub handles DELETE /partner/hubs/{id}
func (h *Handler) DeleteHub(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid hub id"))
		return
	}
	if err := h.svc.DeleteHub(r.Context(), id, t.ID); err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListHubAreas handles GET /partner/hubs/{id}/areas
func (h *Handler) ListHubAreas(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid hub id"))
		return
	}
	areas, err := h.svc.ListHubAreas(r.Context(), id)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, areas)
}

// CreateHubArea handles POST /partner/hubs/{id}/areas
func (h *Handler) CreateHubArea(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	hubID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid hub id"))
		return
	}

	var req struct {
		Name                     string         `json:"name"`
		DeliveryCharge           pgtype.Numeric `json:"delivery_charge"`
		MinOrderAmount           pgtype.Numeric `json:"min_order_amount"`
		EstimatedDeliveryMinutes int32          `json:"estimated_delivery_minutes"`
		IsActive                 bool           `json:"is_active"`
		SortOrder                int32          `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.Name == "" {
		respond.Error(w, apperror.BadRequest("name is required"))
		return
	}

	area, err := h.svc.CreateHubArea(r.Context(), CreateHubAreaRequest{
		HubID:                    hubID,
		TenantID:                 t.ID,
		Name:                     req.Name,
		DeliveryCharge:           req.DeliveryCharge,
		MinOrderAmount:           req.MinOrderAmount,
		EstimatedDeliveryMinutes: req.EstimatedDeliveryMinutes,
		IsActive:                 req.IsActive,
		SortOrder:                req.SortOrder,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusCreated, area)
}

// UpdateHubArea handles PUT /partner/hubs/{id}/areas/{area_id}
func (h *Handler) UpdateHubArea(w http.ResponseWriter, r *http.Request) {
	areaID, err := uuid.Parse(chi.URLParam(r, "area_id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid area id"))
		return
	}

	var req struct {
		Name                     *string        `json:"name"`
		DeliveryCharge           pgtype.Numeric `json:"delivery_charge"`
		MinOrderAmount           pgtype.Numeric `json:"min_order_amount"`
		EstimatedDeliveryMinutes *int32         `json:"estimated_delivery_minutes"`
		IsActive                 *bool          `json:"is_active"`
		SortOrder                *int32         `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	area, err := h.svc.UpdateHubArea(r.Context(), areaID, UpdateHubAreaRequest{
		Name:                     req.Name,
		DeliveryCharge:           req.DeliveryCharge,
		MinOrderAmount:           req.MinOrderAmount,
		EstimatedDeliveryMinutes: req.EstimatedDeliveryMinutes,
		IsActive:                 req.IsActive,
		SortOrder:                req.SortOrder,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, area)
}

// DeleteHubArea handles DELETE /partner/hubs/{id}/areas/{area_id}
func (h *Handler) DeleteHubArea(w http.ResponseWriter, r *http.Request) {
	areaID, err := uuid.Parse(chi.URLParam(r, "area_id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid area id"))
		return
	}
	if err := h.svc.DeleteHubArea(r.Context(), areaID); err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetDeliveryZoneConfig handles GET /partner/delivery/config
func (h *Handler) GetDeliveryZoneConfig(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	cfg, err := h.svc.GetDeliveryZoneConfig(r.Context(), t.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, cfg)
}

// UpsertDeliveryZoneConfig handles PUT /partner/delivery/config
func (h *Handler) UpsertDeliveryZoneConfig(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	var req struct {
		Model                 sqlc.DeliveryModel `json:"model"`
		DistanceTiers         json.RawMessage    `json:"distance_tiers"`
		FreeDeliveryThreshold pgtype.Numeric     `json:"free_delivery_threshold"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	cfg, err := h.svc.UpsertDeliveryZoneConfig(r.Context(), t.ID, UpsertDeliveryZoneConfigRequest{
		Model:                 req.Model,
		DistanceTiers:         req.DistanceTiers,
		FreeDeliveryThreshold: req.FreeDeliveryThreshold,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, cfg)
}

func toAppError(err error) *apperror.AppError {
	if e, ok := err.(*apperror.AppError); ok {
		return e
	}
	return apperror.Internal("unexpected error", err)
}
