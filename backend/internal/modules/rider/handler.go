package rider

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
	"github.com/munchies/platform/backend/internal/pkg/respond"
)

// Handler handles rider HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new rider handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ---------- Partner API ----------

// ListRiders handles GET /partner/riders
func (h *Handler) ListRiders(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	limit, offset := parsePagination(r)
	riders, total, err := h.svc.ListRiders(r.Context(), t.ID, limit, offset)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"riders": riders,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// CreateRider handles POST /partner/riders
func (h *Handler) CreateRider(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	var req struct {
		UserID              uuid.UUID `json:"user_id"`
		HubID               *uuid.UUID `json:"hub_id"`
		VehicleType         string    `json:"vehicle_type"`
		VehicleRegistration string    `json:"vehicle_registration"`
		LicenseNumber       string    `json:"license_number"`
		NidNumber           string    `json:"nid_number"`
		NidVerified         bool      `json:"nid_verified"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.UserID == uuid.Nil {
		respond.Error(w, apperror.BadRequest("user_id is required"))
		return
	}
	if req.VehicleType == "" {
		respond.Error(w, apperror.BadRequest("vehicle_type is required"))
		return
	}

	rider, err := h.svc.CreateRider(r.Context(), t.ID, CreateRiderParams{
		UserID:              req.UserID,
		HubID:               req.HubID,
		VehicleType:         sqlc.VehicleType(req.VehicleType),
		VehicleRegistration: req.VehicleRegistration,
		LicenseNumber:       req.LicenseNumber,
		NidNumber:           req.NidNumber,
		NidVerified:         req.NidVerified,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	respond.JSON(w, http.StatusCreated, rider)
}

// GetRider handles GET /partner/riders/{id}
func (h *Handler) GetRider(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid rider ID"))
		return
	}

	rider, err := h.svc.GetRider(r.Context(), id, t.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	respond.JSON(w, http.StatusOK, rider)
}

// UpdateRider handles PUT /partner/riders/{id}
func (h *Handler) UpdateRider(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid rider ID"))
		return
	}

	var req struct {
		HubID               *uuid.UUID `json:"hub_id"`
		VehicleType         string     `json:"vehicle_type"`
		VehicleRegistration string     `json:"vehicle_registration"`
		LicenseNumber       string     `json:"license_number"`
		NidNumber           string     `json:"nid_number"`
		NidVerified         *bool      `json:"nid_verified"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	rider, err := h.svc.UpdateRider(r.Context(), id, t.ID, UpdateRiderParams{
		HubID:               req.HubID,
		VehicleType:         req.VehicleType,
		VehicleRegistration: req.VehicleRegistration,
		LicenseNumber:       req.LicenseNumber,
		NidNumber:           req.NidNumber,
		NidVerified:         req.NidVerified,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	respond.JSON(w, http.StatusOK, rider)
}

// DeleteRider handles DELETE /partner/riders/{id}
func (h *Handler) DeleteRider(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid rider ID"))
		return
	}

	if err := h.svc.DeleteRider(r.Context(), id, t.ID); err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ListAttendance handles GET /partner/riders/attendance
func (h *Handler) ListAttendance(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	dateStr := r.URL.Query().Get("date")
	date := time.Now()
	if dateStr != "" {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid date format, use YYYY-MM-DD"))
			return
		}
		date = parsed
	}

	limit, offset := parsePagination(r)
	records, err := h.svc.ListAttendanceByDate(r.Context(), t.ID, date, limit, offset)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"attendance": records,
		"date":       date.Format("2006-01-02"),
	})
}

// ---------- Rider API ----------

func requireRider(r *http.Request) (*sqlc.User, *sqlc.Tenant, *apperror.AppError) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		return nil, nil, apperror.Unauthorized("authentication required")
	}
	if u.Role != sqlc.UserRoleRider {
		return nil, nil, apperror.Forbidden("rider role required")
	}
	t := tenant.FromContext(r.Context())
	if t == nil {
		return nil, nil, apperror.NotFound("tenant")
	}
	return u, t, nil
}

// CheckIn handles POST /api/v1/rider/attendance/checkin
func (h *Handler) CheckIn(w http.ResponseWriter, r *http.Request) {
	u, t, appErr := requireRider(r)
	if appErr != nil {
		respond.Error(w, appErr)
		return
	}

	var req struct {
		HubID uuid.UUID `json:"hub_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.HubID == uuid.Nil {
		respond.Error(w, apperror.BadRequest("hub_id is required"))
		return
	}

	rider, err := h.svc.GetRiderByUserID(r.Context(), u.ID, t.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	att, err := h.svc.CheckIn(r.Context(), rider.ID, t.ID, req.HubID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	respond.JSON(w, http.StatusOK, att)
}

// CheckOut handles POST /api/v1/rider/attendance/checkout
func (h *Handler) CheckOut(w http.ResponseWriter, r *http.Request) {
	u, t, appErr := requireRider(r)
	if appErr != nil {
		respond.Error(w, appErr)
		return
	}

	rider, err := h.svc.GetRiderByUserID(r.Context(), u.ID, t.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	att, err := h.svc.CheckOut(r.Context(), rider.ID, t.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	respond.JSON(w, http.StatusOK, att)
}

// UpdateAvailability handles PATCH /api/v1/rider/availability
func (h *Handler) UpdateAvailability(w http.ResponseWriter, r *http.Request) {
	u, t, appErr := requireRider(r)
	if appErr != nil {
		respond.Error(w, appErr)
		return
	}

	var req struct {
		IsAvailable bool `json:"is_available"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	rider, err := h.svc.GetRiderByUserID(r.Context(), u.ID, t.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	updated, err := h.svc.UpdateAvailability(r.Context(), rider.ID, t.ID, req.IsAvailable)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"is_available": updated.IsAvailable,
		"is_on_duty":   updated.IsOnDuty,
	})
}

// ---------- Helpers ----------

func parsePagination(r *http.Request) (limit, offset int32) {
	limit = 20
	offset = 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = int32(n)
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}
	return
}

func toAppError(err error) *apperror.AppError {
	if e, ok := err.(*apperror.AppError); ok {
		return e
	}
	return apperror.Internal("unexpected error", err)
}
