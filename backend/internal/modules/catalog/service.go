package catalog

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
	"github.com/munchies/platform/backend/internal/pkg/slug"
)

// Service implements catalog business logic (categories + products).
type Service struct {
	repo *Repository
}

// NewService creates a new catalog service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateCategoryRequest holds fields for creating a category.
type CreateCategoryRequest struct {
	RestaurantID         pgtype.UUID
	ParentID             pgtype.UUID
	Name                 string
	Description          *string
	ImageUrl             *string
	IconUrl              *string
	ExtraPrepTimeMinutes int32
	IsTobacco            bool
	IsActive             bool
	SortOrder            int32
}

// CreateCategory creates a new category.
func (s *Service) CreateCategory(ctx context.Context, tenantID uuid.UUID, req CreateCategoryRequest) (*sqlc.Category, error) {
	catSlug := slug.Generate(req.Name)
	return s.repo.CreateCategory(ctx, sqlc.CreateCategoryParams{
		TenantID:             tenantID,
		RestaurantID:         req.RestaurantID,
		ParentID:             req.ParentID,
		Name:                 req.Name,
		Slug:                 catSlug,
		Description:          nullString(req.Description),
		ImageUrl:             nullString(req.ImageUrl),
		IconUrl:              nullString(req.IconUrl),
		ExtraPrepTimeMinutes: req.ExtraPrepTimeMinutes,
		IsTobacco:            req.IsTobacco,
		IsActive:             req.IsActive,
		SortOrder:            req.SortOrder,
	})
}

// GetCategory returns a category by ID.
func (s *Service) GetCategory(ctx context.Context, id, tenantID uuid.UUID) (*sqlc.Category, error) {
	c, err := s.repo.GetCategoryByID(ctx, id, tenantID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("category")
	}
	return c, err
}

// ListCategories returns all active categories for a restaurant.
func (s *Service) ListCategories(ctx context.Context, restaurantID, tenantID uuid.UUID) ([]sqlc.Category, error) {
	return s.repo.ListCategoriesByRestaurant(ctx, restaurantID, tenantID)
}

// UpdateCategoryRequest holds updateable category fields.
type UpdateCategoryRequest struct {
	Name        *string
	Description *string
	ImageUrl    *string
	SortOrder   *int32
	IsActive    *bool
}

// UpdateCategory updates a category.
func (s *Service) UpdateCategory(ctx context.Context, id, tenantID uuid.UUID, req UpdateCategoryRequest) (*sqlc.Category, error) {
	c, err := s.repo.UpdateCategory(ctx, sqlc.UpdateCategoryParams{
		ID:          id,
		TenantID:    tenantID,
		Name:        nullString(req.Name),
		Description: nullString(req.Description),
		ImageUrl:    nullString(req.ImageUrl),
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("category")
	}
	return c, err
}

// DeleteCategory soft-deletes a category.
func (s *Service) DeleteCategory(ctx context.Context, id, tenantID uuid.UUID) error {
	return s.repo.DeleteCategory(ctx, id, tenantID)
}

// ReorderCategoryRequest holds a category sort order update.
type ReorderCategoryRequest struct {
	ID        uuid.UUID `json:"id"`
	SortOrder int32     `json:"sort_order"`
}

// ReorderCategories updates sort orders for multiple categories.
func (s *Service) ReorderCategories(ctx context.Context, tenantID uuid.UUID, items []ReorderCategoryRequest) error {
	for _, item := range items {
		if err := s.repo.UpdateCategorySortOrder(ctx, item.ID, item.SortOrder, tenantID); err != nil {
			return err
		}
	}
	return nil
}

// CreateProductRequest holds fields for creating a product.
type CreateProductRequest struct {
	RestaurantID uuid.UUID
	CategoryID   pgtype.UUID
	Name         string
	Description  *string
	BasePrice    pgtype.Numeric
	VatRate      pgtype.Numeric
	Availability sqlc.ProductAvail
	Images       []string
	Tags         []string
	IsFeatured   bool
	IsInvTracked bool
	SortOrder    int32
}

// CreateProduct creates a new product.
func (s *Service) CreateProduct(ctx context.Context, tenantID uuid.UUID, req CreateProductRequest) (*sqlc.Product, error) {
	prodSlug := slug.Generate(req.Name)
	if req.Availability == "" {
		req.Availability = sqlc.ProductAvailAvailable
	}
	return s.repo.CreateProduct(ctx, sqlc.CreateProductParams{
		TenantID:     tenantID,
		RestaurantID: req.RestaurantID,
		CategoryID:   req.CategoryID,
		Name:         req.Name,
		Slug:         prodSlug,
		Description:  nullString(req.Description),
		BasePrice:    req.BasePrice,
		VatRate:      req.VatRate,
		Availability: req.Availability,
		Images:       req.Images,
		Tags:         req.Tags,
		IsFeatured:   req.IsFeatured,
		IsInvTracked: req.IsInvTracked,
		SortOrder:    req.SortOrder,
	})
}

// GetProduct returns a product by ID.
func (s *Service) GetProduct(ctx context.Context, id, tenantID uuid.UUID) (*sqlc.Product, error) {
	p, err := s.repo.GetProductByID(ctx, id, tenantID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("product")
	}
	return p, err
}

// GetProductPublic returns an available product by ID (no tenant check).
func (s *Service) GetProductPublic(ctx context.Context, id uuid.UUID) (*sqlc.Product, error) {
	p, err := s.repo.GetProductByIDPublic(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("product")
	}
	return p, err
}

// ListProducts returns paginated products for a restaurant.
func (s *Service) ListProducts(ctx context.Context, restaurantID, tenantID uuid.UUID, page, perPage int) ([]sqlc.Product, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)
	total, err := s.repo.CountProductsByRestaurant(ctx, restaurantID, tenantID)
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("count products", err)
	}
	items, err := s.repo.ListProductsByRestaurant(ctx, restaurantID, tenantID, int32(limit), int32(offset))
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list products", err)
	}
	meta := pagination.NewMeta(total, limit, "")
	return items, meta, nil
}

// UpdateProductRequest holds updateable product fields.
type UpdateProductRequest struct {
	Name        *string
	Description *string
	CategoryID  pgtype.UUID
	BasePrice   pgtype.Numeric
	VatRate     pgtype.Numeric
	Images      []string
	Tags        []string
	IsFeatured  *bool
	SortOrder   *int32
}

// UpdateProduct updates a product.
func (s *Service) UpdateProduct(ctx context.Context, id, tenantID uuid.UUID, req UpdateProductRequest) (*sqlc.Product, error) {
	p, err := s.repo.UpdateProduct(ctx, sqlc.UpdateProductParams{
		ID:          id,
		TenantID:    tenantID,
		Name:        nullString(req.Name),
		Description: nullString(req.Description),
		CategoryID:  req.CategoryID,
		BasePrice:   req.BasePrice,
		VatRate:     req.VatRate,
		Images:      req.Images,
		Tags:        req.Tags,
		IsFeatured:  req.IsFeatured,
		SortOrder:   req.SortOrder,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("product")
	}
	return p, err
}

// UpdateProductAvailability sets product availability.
func (s *Service) UpdateProductAvailability(ctx context.Context, id, tenantID uuid.UUID, avail sqlc.ProductAvail) (*sqlc.Product, error) {
	p, err := s.repo.UpdateProductAvailability(ctx, id, avail, tenantID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("product")
	}
	return p, err
}

// DeleteProduct deletes a product.
func (s *Service) DeleteProduct(ctx context.Context, id, tenantID uuid.UUID) error {
	return s.repo.DeleteProduct(ctx, id, tenantID)
}

// CreateModifierGroupRequest holds fields for creating a modifier group.
type CreateModifierGroupRequest struct {
	ProductID   uuid.UUID
	Name        string
	Description *string
	MinRequired int32
	MaxAllowed  int32
	SortOrder   int32
}

// CreateModifierGroup creates a modifier group and updates has_modifiers on the product.
func (s *Service) CreateModifierGroup(ctx context.Context, tenantID uuid.UUID, req CreateModifierGroupRequest) (*sqlc.ProductModifierGroup, error) {
	g, err := s.repo.CreateModifierGroup(ctx, sqlc.CreateModifierGroupParams{
		ProductID:   req.ProductID,
		TenantID:    tenantID,
		Name:        req.Name,
		Description: nullString(req.Description),
		MinRequired: req.MinRequired,
		MaxAllowed:  req.MaxAllowed,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		return nil, err
	}
	_ = s.repo.UpdateProductHasModifiers(ctx, req.ProductID, true)
	return g, nil
}

// ListModifierGroups returns modifier groups for a product with their options.
func (s *Service) ListModifierGroups(ctx context.Context, productID uuid.UUID) ([]sqlc.ProductModifierGroup, error) {
	return s.repo.ListModifierGroupsByProduct(ctx, productID)
}

// UpdateModifierGroupRequest holds updateable modifier group fields.
type UpdateModifierGroupRequest struct {
	Name        *string
	Description *string
	MinRequired *int32
	MaxAllowed  *int32
	SortOrder   *int32
}

// UpdateModifierGroup updates a modifier group.
func (s *Service) UpdateModifierGroup(ctx context.Context, id uuid.UUID, req UpdateModifierGroupRequest) (*sqlc.ProductModifierGroup, error) {
	g, err := s.repo.UpdateModifierGroup(ctx, sqlc.UpdateModifierGroupParams{
		ID:          id,
		Name:        nullString(req.Name),
		Description: nullString(req.Description),
		MinRequired: req.MinRequired,
		MaxAllowed:  req.MaxAllowed,
		SortOrder:   req.SortOrder,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("modifier group")
	}
	return g, err
}

// DeleteModifierGroup deletes a modifier group and its options.
func (s *Service) DeleteModifierGroup(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteModifierOptionsByGroup(ctx, id); err != nil {
		return err
	}
	return s.repo.DeleteModifierGroup(ctx, id)
}

// CreateModifierOptionRequest holds fields for creating a modifier option.
type CreateModifierOptionRequest struct {
	ModifierGroupID uuid.UUID
	ProductID       uuid.UUID
	Name            string
	AdditionalPrice pgtype.Numeric
	IsAvailable     bool
	SortOrder       int32
}

// CreateModifierOption creates a modifier option.
func (s *Service) CreateModifierOption(ctx context.Context, tenantID uuid.UUID, req CreateModifierOptionRequest) (*sqlc.ProductModifierOption, error) {
	return s.repo.CreateModifierOption(ctx, sqlc.CreateModifierOptionParams{
		ModifierGroupID: req.ModifierGroupID,
		ProductID:       req.ProductID,
		TenantID:        tenantID,
		Name:            req.Name,
		AdditionalPrice: req.AdditionalPrice,
		IsAvailable:     req.IsAvailable,
		SortOrder:       req.SortOrder,
	})
}

// ListModifierOptions returns options for a modifier group.
func (s *Service) ListModifierOptions(ctx context.Context, groupID uuid.UUID) ([]sqlc.ProductModifierOption, error) {
	return s.repo.ListModifierOptionsByGroup(ctx, groupID)
}

// UpdateModifierOptionRequest holds updateable modifier option fields.
type UpdateModifierOptionRequest struct {
	Name            *string
	AdditionalPrice pgtype.Numeric
	IsAvailable     *bool
	SortOrder       *int32
}

// UpdateModifierOption updates a modifier option.
func (s *Service) UpdateModifierOption(ctx context.Context, id uuid.UUID, req UpdateModifierOptionRequest) (*sqlc.ProductModifierOption, error) {
	o, err := s.repo.UpdateModifierOption(ctx, sqlc.UpdateModifierOptionParams{
		ID:              id,
		Name:            nullString(req.Name),
		AdditionalPrice: req.AdditionalPrice,
		IsAvailable:     req.IsAvailable,
		SortOrder:       req.SortOrder,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("modifier option")
	}
	return o, err
}

// DeleteModifierOption deletes a modifier option.
func (s *Service) DeleteModifierOption(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteModifierOption(ctx, id)
}

// UpsertDiscountRequest holds fields for upserting a product discount.
type UpsertDiscountRequest struct {
	ProductID      uuid.UUID
	RestaurantID   uuid.UUID
	DiscountType   sqlc.DiscountType
	Amount         pgtype.Numeric
	MaxDiscountCap pgtype.Numeric
	StartsAt       time.Time
	EndsAt         pgtype.Timestamptz
	IsActive       bool
	CreatedBy      pgtype.UUID
}

// UpsertProductDiscount upserts a product discount.
func (s *Service) UpsertProductDiscount(ctx context.Context, tenantID uuid.UUID, req UpsertDiscountRequest) (*sqlc.ProductDiscount, error) {
	return s.repo.UpsertProductDiscount(ctx, sqlc.UpsertProductDiscountParams{
		ProductID:      req.ProductID,
		RestaurantID:   req.RestaurantID,
		TenantID:       tenantID,
		DiscountType:   req.DiscountType,
		Amount:         req.Amount,
		MaxDiscountCap: req.MaxDiscountCap,
		StartsAt:       req.StartsAt,
		EndsAt:         req.EndsAt,
		IsActive:       req.IsActive,
		CreatedBy:      req.CreatedBy,
	})
}

// DeactivateDiscount deactivates all active discounts for a product.
func (s *Service) DeactivateDiscount(ctx context.Context, productID uuid.UUID) error {
	return s.repo.DeactivateProductDiscount(ctx, productID)
}

// DuplicateMenu copies all categories and products from one restaurant to another.
func (s *Service) DuplicateMenu(ctx context.Context, srcRestaurantID, dstRestaurantID, tenantID uuid.UUID) error {
	categories, err := s.repo.ListCategoriesByRestaurant(ctx, srcRestaurantID, tenantID)
	if err != nil {
		return err
	}

	catIDMap := make(map[uuid.UUID]uuid.UUID)
	for _, cat := range categories {
		newCat, err := s.repo.CreateCategory(ctx, sqlc.CreateCategoryParams{
			TenantID:             tenantID,
			RestaurantID:         pgtype.UUID{Bytes: dstRestaurantID, Valid: true},
			ParentID:             cat.ParentID,
			Name:                 cat.Name,
			Slug:                 cat.Slug,
			Description:          cat.Description,
			ImageUrl:             cat.ImageUrl,
			IconUrl:              cat.IconUrl,
			ExtraPrepTimeMinutes: cat.ExtraPrepTimeMinutes,
			IsTobacco:            cat.IsTobacco,
			IsActive:             cat.IsActive,
			SortOrder:            cat.SortOrder,
		})
		if err != nil {
			return err
		}
		catIDMap[cat.ID] = newCat.ID
	}

	products, err := s.repo.ListAvailableProductsByRestaurant(ctx, srcRestaurantID)
	if err != nil {
		return err
	}

	for _, prod := range products {
		newCatID := pgtype.UUID{} // default: no category
		if prod.CategoryID.Valid {
			if mapped, ok := catIDMap[prod.CategoryID.Bytes]; ok {
				newCatID = pgtype.UUID{Bytes: mapped, Valid: true}
			}
			// If category not found in map, product is created without a category
		}
		_, err := s.repo.CreateProduct(ctx, sqlc.CreateProductParams{
			TenantID:     tenantID,
			RestaurantID: dstRestaurantID,
			CategoryID:   newCatID,
			Name:         prod.Name,
			Slug:         prod.Slug,
			Description:  prod.Description,
			BasePrice:    prod.BasePrice,
			VatRate:      prod.VatRate,
			Availability: prod.Availability,
			Images:       prod.Images,
			Tags:         prod.Tags,
			IsFeatured:   prod.IsFeatured,
			IsInvTracked: prod.IsInvTracked,
			SortOrder:    prod.SortOrder,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}
