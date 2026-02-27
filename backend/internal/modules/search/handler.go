package search

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/munchies/platform/backend/internal/modules/auth"
	"github.com/munchies/platform/backend/internal/modules/tenant"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/respond"
)

// Handler handles search HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new search handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Search handles GET /api/v1/search
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	query := r.URL.Query().Get("q")
	searchType := r.URL.Query().Get("type")

	var userID *uuid.UUID
	if u := auth.UserFromContext(r.Context()); u != nil {
		userID = &u.ID
	}

	result, err := h.svc.Search(r.Context(), t.ID, userID, query, searchType)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, result)
}

// Autocomplete handles GET /api/v1/search/autocomplete
func (h *Handler) Autocomplete(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	query := r.URL.Query().Get("q")
	result, err := h.svc.Autocomplete(r.Context(), t.ID, query)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, result)
}

// TopSearchTerms handles GET /partner/reports/searches
func (h *Handler) TopSearchTerms(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	terms, err := h.svc.GetTopSearchTerms(r.Context(), t.ID, 20)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, terms)
}

func toAppError(err error) *apperror.AppError {
	if e, ok := err.(*apperror.AppError); ok {
		return e
	}
	return apperror.Internal("unexpected error", err)
}
