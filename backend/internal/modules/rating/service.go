package rating

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/pagination"
	"github.com/shopspring/decimal"
)

// Service implements rating business logic.
type Service struct {
	q *sqlc.Queries
}

// NewService creates a new rating service.
func NewService(q *sqlc.Queries) *Service {
	return &Service{q: q}
}

// CreateRating creates a rating for a delivered order. One rating per order.
func (s *Service) CreateRating(ctx context.Context, tenantID uuid.UUID, req CreateRatingRequest) (*sqlc.Review, error) {
	// Verify order exists and is delivered
	order, err := s.q.GetOrderByID(ctx, sqlc.GetOrderByIDParams{
		ID:       req.OrderID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("order")
	}
	if err != nil {
		return nil, err
	}
	if order.Status != sqlc.OrderStatusDelivered {
		return nil, apperror.BadRequest("can only rate delivered orders")
	}

	// Check if already rated
	_, err = s.q.GetReviewByOrderAndUser(ctx, sqlc.GetReviewByOrderAndUserParams{
		OrderID: req.OrderID,
		UserID:  req.UserID,
	})
	if err == nil {
		return nil, apperror.Conflict("order already rated")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	review, err := s.q.CreateReview(ctx, sqlc.CreateReviewParams{
		TenantID:         tenantID,
		OrderID:          req.OrderID,
		UserID:           req.UserID,
		RestaurantID:     req.RestaurantID,
		RestaurantRating: req.RestaurantRating,
		RiderRating:      req.RiderRating,
		Comment:          req.Comment,
		Images:           req.Images,
	})
	if err != nil {
		return nil, err
	}

	// Update restaurant aggregate rating
	go s.updateRestaurantRating(context.Background(), req.RestaurantID)

	return &review, nil
}

// updateRestaurantRating recalculates and updates the aggregate rating.
func (s *Service) updateRestaurantRating(ctx context.Context, restaurantID uuid.UUID) {
	avg, err := s.q.GetRestaurantAvgRating(ctx, restaurantID)
	if err != nil {
		return
	}
	avgRating, _ := avg.AvgRating.Float64()
	s.q.UpdateRestaurantRating(ctx, sqlc.UpdateRestaurantRatingParams{
		ID:          restaurantID,
		RatingAvg:   decimal.NewFromFloat(avgRating),
		RatingCount: int32(avg.ReviewCount),
	})
}

// ListByRestaurant returns paginated public reviews for a restaurant.
func (s *Service) ListByRestaurant(ctx context.Context, tenantID, restaurantID uuid.UUID, page, perPage int) ([]sqlc.Review, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)
	total, err := s.q.CountReviewsByRestaurant(ctx, sqlc.CountReviewsByRestaurantParams{
		RestaurantID: restaurantID,
		TenantID:     tenantID,
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("count reviews", err)
	}
	items, err := s.q.ListReviewsByRestaurant(ctx, sqlc.ListReviewsByRestaurantParams{
		RestaurantID: restaurantID,
		TenantID:     tenantID,
		Limit:        int32(limit),
		Offset:       int32(offset),
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list reviews", err)
	}
	return items, pagination.NewMeta(total, limit, ""), nil
}

// RespondToReview allows a partner to respond to a review.
func (s *Service) RespondToReview(ctx context.Context, tenantID, reviewID uuid.UUID, reply string) (*sqlc.Review, error) {
	review, err := s.q.UpdateRestaurantReply(ctx, sqlc.UpdateRestaurantReplyParams{
		ID:              reviewID,
		TenantID:        tenantID,
		RestaurantReply: &reply,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("review")
	}
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// CreateRatingRequest holds fields for creating a rating.
type CreateRatingRequest struct {
	OrderID          uuid.UUID
	UserID           uuid.UUID
	RestaurantID     uuid.UUID
	RestaurantRating int16
	RiderRating      *int16
	Comment          *string
	Images           []string
}
