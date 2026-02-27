package delivery

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/munchies/platform/backend/internal/modules/tenant"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/respond"
)

// Handler handles delivery charge HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new delivery handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// CalculateCharge handles POST /api/v1/delivery/charges
func (h *Handler) CalculateCharge(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	var req struct {
		HubID    string `json:"hub_id"`
		AreaSlug string `json:"area_slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.HubID == "" || req.AreaSlug == "" {
		respond.Error(w, apperror.BadRequest("hub_id and area_slug are required"))
		return
	}
	hubID, err := uuid.Parse(req.HubID)
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid hub_id"))
		return
	}

	result, err := h.svc.Calculate(r.Context(), t.ID, hubID, req.AreaSlug)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, result)
}

func toAppError(err error) *apperror.AppError {
	if e, ok := err.(*apperror.AppError); ok {
		return e
	}
	return apperror.Internal("unexpected error", err)
}
