package inventory

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/modules/auth"
	"github.com/munchies/platform/backend/internal/modules/tenant"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/pagination"
	"github.com/munchies/platform/backend/internal/pkg/respond"
)

// Handler handles inventory HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new inventory handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListInventory handles GET /partner/inventory
func (h *Handler) ListInventory(w http.ResponseWriter, r *http.Request) {
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

	restaurantID, err := uuid.Parse(r.URL.Query().Get("restaurant_id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("restaurant_id is required"))
		return
	}

	page, perPage := parsePagination(r)
	items, meta, err := h.svc.ListInventory(r.Context(), t.ID, restaurantID, page, perPage)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: items, Meta: meta})
}

// AdjustStock handles POST /partner/inventory/adjust
func (h *Handler) AdjustStock(w http.ResponseWriter, r *http.Request) {
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
		InventoryItemID string `json:"inventory_item_id"`
		RestaurantID    string `json:"restaurant_id"`
		QtyChange       int32  `json:"qty_change"`
		Reason          string `json:"reason"`
		Note            string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	itemID, err := uuid.Parse(req.InventoryItemID)
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid inventory_item_id"))
		return
	}
	restaurantID, err := uuid.Parse(req.RestaurantID)
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid restaurant_id"))
		return
	}
	if req.QtyChange == 0 {
		respond.Error(w, apperror.BadRequest("qty_change cannot be zero"))
		return
	}

	reason := sqlc.InventoryAdjustmentReason(req.Reason)
	validReasons := map[sqlc.InventoryAdjustmentReason]bool{
		sqlc.InventoryAdjustmentReasonOpeningStock:      true,
		sqlc.InventoryAdjustmentReasonPurchase:          true,
		sqlc.InventoryAdjustmentReasonManualAdjustment:  true,
		sqlc.InventoryAdjustmentReasonDamageLoss:        true,
		sqlc.InventoryAdjustmentReasonStockReturn:       true,
	}
	if !validReasons[reason] {
		respond.Error(w, apperror.BadRequest("invalid adjustment reason"))
		return
	}

	item, adj, err := h.svc.AdjustStock(r.Context(), AdjustStockRequest{
		InventoryItemID: itemID,
		TenantID:        t.ID,
		RestaurantID:    restaurantID,
		QtyChange:       req.QtyChange,
		Reason:          reason,
		Note:            req.Note,
		AdjustedBy:      u.ID,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"inventory_item": item,
		"adjustment":     adj,
	})
}

// ListLowStock handles GET /partner/inventory/low-stock
func (h *Handler) ListLowStock(w http.ResponseWriter, r *http.Request) {
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

	restaurantID, err := uuid.Parse(r.URL.Query().Get("restaurant_id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("restaurant_id is required"))
		return
	}

	page, perPage := parsePagination(r)
	items, meta, err := h.svc.ListLowStock(r.Context(), t.ID, restaurantID, page, perPage)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: items, Meta: meta})
}

// RegisterRoutes registers inventory routes on the given router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListInventory)
	r.Post("/adjust", h.AdjustStock)
	r.Get("/low-stock", h.ListLowStock)
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
