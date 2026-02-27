package user

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/munchies/platform/backend/internal/modules/auth"
	"github.com/munchies/platform/backend/internal/modules/tenant"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/pagination"
	"github.com/munchies/platform/backend/internal/pkg/respond"
)

// Handler handles user HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new user handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetMe handles GET /me
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	respond.JSON(w, http.StatusOK, u)
}

// UpdateMe handles PATCH /me
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}

	var req struct {
		Name            *string `json:"name"`
		Email           *string `json:"email"`
		AvatarURL       *string `json:"avatar_url"`
		DevicePushToken *string `json:"device_push_token"`
		DevicePlatform  *string `json:"device_platform"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	updated, err := h.svc.UpdateProfile(r.Context(), u.ID, UpdateProfileRequest{
		Name:            req.Name,
		Email:           req.Email,
		AvatarURL:       req.AvatarURL,
		DevicePushToken: req.DevicePushToken,
		DevicePlatform:  req.DevicePlatform,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, updated)
}

// ListAddresses handles GET /me/addresses
func (h *Handler) ListAddresses(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	addrs, err := h.svc.ListAddresses(r.Context(), u.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, addrs)
}

// CreateAddress handles POST /me/addresses
func (h *Handler) CreateAddress(w http.ResponseWriter, r *http.Request) {
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
		Label          string  `json:"label"`
		RecipientName  *string `json:"recipient_name"`
		RecipientPhone *string `json:"recipient_phone"`
		AddressLine1   string  `json:"address_line1"`
		AddressLine2   *string `json:"address_line2"`
		Area           string  `json:"area"`
		City           string  `json:"city"`
		IsDefault      bool    `json:"is_default"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.AddressLine1 == "" || req.Area == "" {
		respond.Error(w, apperror.BadRequest("address_line1 and area are required"))
		return
	}
	if req.Label == "" {
		req.Label = "Home"
	}
	if req.City == "" {
		req.City = "Dhaka"
	}

	addr, err := h.svc.CreateAddress(r.Context(), u.ID, CreateAddressRequest{
		TenantID:       t.ID,
		Label:          req.Label,
		RecipientName:  req.RecipientName,
		RecipientPhone: req.RecipientPhone,
		AddressLine1:   req.AddressLine1,
		AddressLine2:   req.AddressLine2,
		Area:           req.Area,
		City:           req.City,
		IsDefault:      req.IsDefault,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusCreated, addr)
}

// DeleteAddress handles DELETE /me/addresses/{id}
func (h *Handler) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	addrID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid address id"))
		return
	}
	if err := h.svc.DeleteAddress(r.Context(), u.ID, addrID); err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListWallet handles GET /me/wallet
func (h *Handler) ListWallet(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	page, perPage := parsePagination(r)
	items, meta, err := h.svc.ListWallet(r.Context(), u.ID, page, perPage)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: items, Meta: meta})
}

// ListNotifications handles GET /me/notifications
func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	page, perPage := parsePagination(r)
	items, meta, err := h.svc.ListNotifications(r.Context(), u.ID, page, perPage)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: items, Meta: meta})
}

// MarkNotificationRead handles PATCH /me/notifications/{id}/read
func (h *Handler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	notifID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid notification id"))
		return
	}
	n, err := h.svc.MarkNotificationRead(r.Context(), u.ID, notifID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, n)
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
