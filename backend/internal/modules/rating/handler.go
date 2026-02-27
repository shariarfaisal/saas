package rating

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

// Handler handles rating HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new rating handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RateOrder handles POST /api/v1/orders/:id/rate
func (h *Handler) RateOrder(w http.ResponseWriter, r *http.Request) {
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
	orderID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid order id"))
		return
	}

	var req struct {
		RestaurantID     string   `json:"restaurant_id"`
		RestaurantRating int16    `json:"restaurant_rating"`
		RiderRating      *int16   `json:"rider_rating"`
		Comment          *string  `json:"comment"`
		Images           []string `json:"images"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	restaurantID, err := uuid.Parse(req.RestaurantID)
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid restaurant_id"))
		return
	}

	if req.RestaurantRating < 1 || req.RestaurantRating > 5 {
		respond.Error(w, apperror.BadRequest("restaurant_rating must be between 1 and 5"))
		return
	}

	images := req.Images
	if images == nil {
		images = []string{}
	}

	review, err := h.svc.CreateRating(r.Context(), t.ID, CreateRatingRequest{
		OrderID:          orderID,
		UserID:           u.ID,
		RestaurantID:     restaurantID,
		RestaurantRating: req.RestaurantRating,
		RiderRating:      req.RiderRating,
		Comment:          req.Comment,
		Images:           images,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusCreated, review)
}

// ListReviews handles GET /api/v1/restaurants/:slug/ratings
func (h *Handler) ListReviews(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	// slug is used as restaurant_id for simplicity
	restaurantID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid restaurant id"))
		return
	}

	page, perPage := parsePagination(r)
	items, meta, err := h.svc.ListByRestaurant(r.Context(), t.ID, restaurantID, page, perPage)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: items, Meta: meta})
}

// RespondToReview handles POST /partner/reviews/:id/respond
func (h *Handler) RespondToReview(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	reviewID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid review id"))
		return
	}

	var req struct {
		Reply string `json:"reply"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.Reply == "" {
		respond.Error(w, apperror.BadRequest("reply is required"))
		return
	}

	review, err := h.svc.RespondToReview(r.Context(), t.ID, reviewID, req.Reply)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, review)
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
