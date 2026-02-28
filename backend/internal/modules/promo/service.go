package promo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/pagination"
	"github.com/shopspring/decimal"
)

// Service handles promo business logic.
type Service struct {
	q *sqlc.Queries
}

// NewService creates a new promo service.
func NewService(q *sqlc.Queries) *Service {
	return &Service{q: q}
}

// CreatePromoRequest holds fields for creating a promo.
type CreatePromoRequest struct {
	TenantID       uuid.UUID
	Code           string
	Title          string
	Description    string
	PromoType      sqlc.PromoType
	DiscountAmount decimal.Decimal
	MaxDiscountCap *decimal.Decimal
	CashbackAmount decimal.Decimal
	FundedBy       sqlc.PromoFunder
	AppliesTo      sqlc.PromoApplyOn
	MinOrderAmount decimal.Decimal
	MaxTotalUses   *int32
	MaxUsesPerUser int32
	IncludeStores  bool
	StartsAt       time.Time
	EndsAt         *time.Time
	CreatedBy      uuid.UUID
}

// CreatePromo creates a new promotion.
func (s *Service) CreatePromo(ctx context.Context, req CreatePromoRequest) (*sqlc.Promo, error) {
	discountAmt := pgtype.Numeric{Valid: true}
	_ = discountAmt.Scan(req.DiscountAmount.String())

	var maxCap pgtype.Numeric
	if req.MaxDiscountCap != nil {
		maxCap = pgtype.Numeric{Valid: true}
		_ = maxCap.Scan(req.MaxDiscountCap.String())
	}

	cashback := pgtype.Numeric{Valid: true}
	_ = cashback.Scan(req.CashbackAmount.String())

	minOrder := pgtype.Numeric{Valid: true}
	_ = minOrder.Scan(req.MinOrderAmount.String())

	var endsAt pgtype.Timestamptz
	if req.EndsAt != nil {
		endsAt = pgtype.Timestamptz{Time: *req.EndsAt, Valid: true}
	}

	p, err := s.q.CreatePromo(ctx, sqlc.CreatePromoParams{
		TenantID:       req.TenantID,
		Code:           req.Code,
		Title:          req.Title,
		Description:    sql.NullString{String: req.Description, Valid: req.Description != ""},
		PromoType:      req.PromoType,
		DiscountAmount: discountAmt,
		MaxDiscountCap: maxCap,
		CashbackAmount: cashback,
		FundedBy:       req.FundedBy,
		AppliesTo:      req.AppliesTo,
		MinOrderAmount: minOrder,
		MaxTotalUses:   req.MaxTotalUses,
		MaxUsesPerUser: req.MaxUsesPerUser,
		IncludeStores:  req.IncludeStores,
		IsActive:       true,
		StartsAt:       req.StartsAt,
		EndsAt:         endsAt,
		CreatedBy:      pgtype.UUID{Bytes: req.CreatedBy, Valid: true},
	})
	if err != nil {
		return nil, apperror.Internal("create promo", err)
	}
	return &p, nil
}

// GetPromo returns a promo by ID.
func (s *Service) GetPromo(ctx context.Context, tenantID, promoID uuid.UUID) (*sqlc.Promo, error) {
	p, err := s.q.GetPromoByID(ctx, sqlc.GetPromoByIDParams{
		ID:       promoID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("promo")
	}
	if err != nil {
		return nil, apperror.Internal("get promo", err)
	}
	return &p, nil
}

// ListPromos returns paginated promos for a tenant.
func (s *Service) ListPromos(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]sqlc.Promo, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)

	total, err := s.q.CountPromos(ctx, tenantID)
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("count promos", err)
	}

	promos, err := s.q.ListPromos(ctx, sqlc.ListPromosParams{
		TenantID: tenantID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list promos", err)
	}

	meta := pagination.NewMeta(total, limit, "")
	return promos, meta, nil
}

// UpdatePromoRequest holds fields for updating a promo.
type UpdatePromoRequest struct {
	Title          *string
	Description    *string
	DiscountAmount *decimal.Decimal
	MaxDiscountCap *decimal.Decimal
	CashbackAmount *decimal.Decimal
	MinOrderAmount *decimal.Decimal
	MaxTotalUses   *int32
	MaxUsesPerUser *int32
	StartsAt       *time.Time
	EndsAt         *time.Time
}

// UpdatePromo updates a promo.
func (s *Service) UpdatePromo(ctx context.Context, tenantID, promoID uuid.UUID, req UpdatePromoRequest) (*sqlc.Promo, error) {
	params := sqlc.UpdatePromoParams{
		ID:       promoID,
		TenantID: tenantID,
	}

	if req.Title != nil {
		params.Title = sql.NullString{String: *req.Title, Valid: true}
	}
	if req.Description != nil {
		params.Description = sql.NullString{String: *req.Description, Valid: true}
	}
	if req.DiscountAmount != nil {
		params.DiscountAmount = pgtype.Numeric{Valid: true}
		_ = params.DiscountAmount.Scan(req.DiscountAmount.String())
	}
	if req.MaxDiscountCap != nil {
		params.MaxDiscountCap = pgtype.Numeric{Valid: true}
		_ = params.MaxDiscountCap.Scan(req.MaxDiscountCap.String())
	}
	if req.CashbackAmount != nil {
		params.CashbackAmount = pgtype.Numeric{Valid: true}
		_ = params.CashbackAmount.Scan(req.CashbackAmount.String())
	}
	if req.MinOrderAmount != nil {
		params.MinOrderAmount = pgtype.Numeric{Valid: true}
		_ = params.MinOrderAmount.Scan(req.MinOrderAmount.String())
	}
	if req.MaxTotalUses != nil {
		params.MaxTotalUses = req.MaxTotalUses
	}
	if req.MaxUsesPerUser != nil {
		params.MaxUsesPerUser = req.MaxUsesPerUser
	}
	if req.StartsAt != nil {
		params.StartsAt = pgtype.Timestamptz{Time: *req.StartsAt, Valid: true}
	}
	if req.EndsAt != nil {
		params.EndsAt = pgtype.Timestamptz{Time: *req.EndsAt, Valid: true}
	}

	p, err := s.q.UpdatePromo(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("promo")
	}
	if err != nil {
		return nil, apperror.Internal("update promo", err)
	}
	return &p, nil
}

// DeactivatePromo deactivates a promo.
func (s *Service) DeactivatePromo(ctx context.Context, tenantID, promoID uuid.UUID) (*sqlc.Promo, error) {
	p, err := s.q.DeactivatePromo(ctx, sqlc.DeactivatePromoParams{
		ID:       promoID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("promo")
	}
	if err != nil {
		return nil, apperror.Internal("deactivate promo", err)
	}
	return &p, nil
}

// CartItem represents an item in the customer's cart for promo validation.
type CartItem struct {
	ProductID    uuid.UUID
	RestaurantID uuid.UUID
	CategoryID   uuid.UUID
	Quantity     int32
	UnitPrice    decimal.Decimal
	ItemSubtotal decimal.Decimal
}

// PromoValidationResult contains the result of promo validation.
type PromoValidationResult struct {
	Valid          bool            `json:"valid"`
	PromoID        uuid.UUID       `json:"promo_id"`
	Code           string          `json:"code"`
	PromoType      sqlc.PromoType  `json:"promo_type"`
	DiscountAmount decimal.Decimal `json:"discount_amount"`
	CashbackAmount decimal.Decimal `json:"cashback_amount"`
	ErrorMessage   string          `json:"error_message,omitempty"`
}

// Validate validates a promo code against the given cart and user context.
func (s *Service) Validate(ctx context.Context, tenantID, userID uuid.UUID, code string, cartTotal decimal.Decimal, cartItems []CartItem) (*PromoValidationResult, error) {
	// 1. Get active promo by code
	promo, err := s.q.GetActivePromoByCode(ctx, sqlc.GetActivePromoByCodeParams{
		Code:     code,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return &PromoValidationResult{
			Valid:        false,
			ErrorMessage: "promo code not found or expired",
		}, nil
	}
	if err != nil {
		return nil, apperror.Internal("get promo", err)
	}

	// 2. Check max_total_uses
	if promo.MaxTotalUses != nil && promo.TotalUses >= *promo.MaxTotalUses {
		return &PromoValidationResult{
			Valid:        false,
			PromoID:      promo.ID,
			Code:         promo.Code,
			ErrorMessage: "promo usage limit reached",
		}, nil
	}

	// 3. Check per-user limit
	userUsageCount, err := s.q.GetUsageCountByUserAndPromo(ctx, sqlc.GetUsageCountByUserAndPromoParams{
		UserID:  userID,
		PromoID: promo.ID,
	})
	if err != nil {
		return nil, apperror.Internal("get usage count", err)
	}
	if userUsageCount >= int64(promo.MaxUsesPerUser) {
		return &PromoValidationResult{
			Valid:        false,
			PromoID:      promo.ID,
			Code:         promo.Code,
			ErrorMessage: "you have already used this promo code the maximum number of times",
		}, nil
	}

	// 4. Check min order amount
	minOrderAmt := decimal.Zero
	if promo.MinOrderAmount.Valid {
		_ = minOrderAmt.Scan(promo.MinOrderAmount.Int.String())
	}
	if cartTotal.LessThan(minOrderAmt) {
		return &PromoValidationResult{
			Valid:        false,
			PromoID:      promo.ID,
			Code:         promo.Code,
			ErrorMessage: "minimum order amount of " + minOrderAmt.String() + " not met",
		}, nil
	}

	// 5. Check user eligibility (if restricted)
	eligibleUsers, err := s.q.ListPromoUserEligibility(ctx, promo.ID)
	if err != nil {
		return nil, apperror.Internal("list promo eligibility", err)
	}
	if len(eligibleUsers) > 0 {
		eligible := false
		for _, eid := range eligibleUsers {
			if eid == userID {
				eligible = true
				break
			}
		}
		if !eligible {
			return &PromoValidationResult{
				Valid:        false,
				PromoID:      promo.ID,
				Code:         promo.Code,
				ErrorMessage: "you are not eligible for this promo",
			}, nil
		}
	}

	// 6. Check restaurant restrictions
	if promo.AppliesTo == sqlc.PromoApplyOnSpecificRestaurant {
		restrictedRestaurants, err := s.q.ListPromoRestaurantRestrictions(ctx, promo.ID)
		if err != nil {
			return nil, apperror.Internal("list restaurant restrictions", err)
		}
		if len(restrictedRestaurants) > 0 {
			restaurantMap := make(map[uuid.UUID]bool)
			for _, rid := range restrictedRestaurants {
				restaurantMap[rid] = true
			}
			for _, item := range cartItems {
				if !restaurantMap[item.RestaurantID] {
					return &PromoValidationResult{
						Valid:        false,
						PromoID:      promo.ID,
						Code:         promo.Code,
						ErrorMessage: "promo is not valid for all restaurants in your cart",
					}, nil
				}
			}
		}
	}

	// 7. Check category restrictions
	if promo.AppliesTo == sqlc.PromoApplyOnCategory {
		restrictedCategories, err := s.q.ListPromoCategoryRestrictions(ctx, promo.ID)
		if err != nil {
			return nil, apperror.Internal("list category restrictions", err)
		}
		if len(restrictedCategories) > 0 {
			categoryMap := make(map[uuid.UUID]bool)
			for _, cid := range restrictedCategories {
				categoryMap[cid] = true
			}
			for _, item := range cartItems {
				if !categoryMap[item.CategoryID] {
					return &PromoValidationResult{
						Valid:        false,
						PromoID:      promo.ID,
						Code:         promo.Code,
						ErrorMessage: "promo is not valid for all categories in your cart",
					}, nil
				}
			}
		}
	}

	// 8. Calculate discount
	discountAmt := decimal.Zero
	if promo.DiscountAmount.Valid {
		_ = discountAmt.Scan(promo.DiscountAmount.Int.String())
	}

	var finalDiscount decimal.Decimal
	if promo.PromoType == sqlc.PromoTypePercent {
		finalDiscount = cartTotal.Mul(discountAmt).Div(decimal.NewFromInt(100))
		// Apply max discount cap
		if promo.MaxDiscountCap.Valid {
			maxCap := decimal.Zero
			_ = maxCap.Scan(promo.MaxDiscountCap.Int.String())
			if finalDiscount.GreaterThan(maxCap) {
				finalDiscount = maxCap
			}
		}
	} else {
		finalDiscount = discountAmt
	}

	// Don't let discount exceed cart total
	if finalDiscount.GreaterThan(cartTotal) {
		finalDiscount = cartTotal
	}

	cashback := decimal.Zero
	if promo.CashbackAmount.Valid {
		_ = cashback.Scan(promo.CashbackAmount.Int.String())
	}

	return &PromoValidationResult{
		Valid:          true,
		PromoID:        promo.ID,
		Code:           promo.Code,
		PromoType:      promo.PromoType,
		DiscountAmount: finalDiscount,
		CashbackAmount: cashback,
	}, nil
}
