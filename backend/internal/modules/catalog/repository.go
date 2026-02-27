package catalog

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
)

// Repository provides catalog data access.
type Repository struct {
	q *sqlc.Queries
}

// NewRepository creates a new catalog repository.
func NewRepository(q *sqlc.Queries) *Repository {
	return &Repository{q: q}
}

// --- Category ---

func (r *Repository) CreateCategory(ctx context.Context, arg sqlc.CreateCategoryParams) (*sqlc.Category, error) {
	c, err := r.q.CreateCategory(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) GetCategoryByID(ctx context.Context, id, tenantID uuid.UUID) (*sqlc.Category, error) {
	c, err := r.q.GetCategoryByID(ctx, sqlc.GetCategoryByIDParams{ID: id, TenantID: tenantID})
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) ListCategoriesByRestaurant(ctx context.Context, restaurantID, tenantID uuid.UUID) ([]sqlc.Category, error) {
	return r.q.ListCategoriesByRestaurant(ctx, sqlc.ListCategoriesByRestaurantParams{
		RestaurantID: pgtype.UUID{Bytes: restaurantID, Valid: true},
		TenantID:     tenantID,
	})
}

func (r *Repository) UpdateCategory(ctx context.Context, arg sqlc.UpdateCategoryParams) (*sqlc.Category, error) {
	c, err := r.q.UpdateCategory(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) DeleteCategory(ctx context.Context, id, tenantID uuid.UUID) error {
	return r.q.DeleteCategory(ctx, sqlc.DeleteCategoryParams{ID: id, TenantID: tenantID})
}

func (r *Repository) UpdateCategorySortOrder(ctx context.Context, id uuid.UUID, sortOrder int32, tenantID uuid.UUID) error {
	return r.q.UpdateCategorySortOrder(ctx, sqlc.UpdateCategorySortOrderParams{
		ID:        id,
		SortOrder: sortOrder,
		TenantID:  tenantID,
	})
}

// --- Product ---

func (r *Repository) CreateProduct(ctx context.Context, arg sqlc.CreateProductParams) (*sqlc.Product, error) {
	p, err := r.q.CreateProduct(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) GetProductByID(ctx context.Context, id, tenantID uuid.UUID) (*sqlc.Product, error) {
	p, err := r.q.GetProductByID(ctx, sqlc.GetProductByIDParams{ID: id, TenantID: tenantID})
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) GetProductByIDPublic(ctx context.Context, id uuid.UUID) (*sqlc.Product, error) {
	p, err := r.q.GetProductByIDPublic(ctx, id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) ListProductsByRestaurant(ctx context.Context, restaurantID, tenantID uuid.UUID, limit, offset int32) ([]sqlc.Product, error) {
	return r.q.ListProductsByRestaurant(ctx, sqlc.ListProductsByRestaurantParams{
		RestaurantID: restaurantID,
		TenantID:     tenantID,
		Limit:        limit,
		Offset:       offset,
	})
}

func (r *Repository) CountProductsByRestaurant(ctx context.Context, restaurantID, tenantID uuid.UUID) (int64, error) {
	return r.q.CountProductsByRestaurant(ctx, sqlc.CountProductsByRestaurantParams{
		RestaurantID: restaurantID,
		TenantID:     tenantID,
	})
}

func (r *Repository) ListAvailableProductsByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]sqlc.Product, error) {
	return r.q.ListAvailableProductsByRestaurant(ctx, restaurantID)
}

func (r *Repository) UpdateProduct(ctx context.Context, arg sqlc.UpdateProductParams) (*sqlc.Product, error) {
	p, err := r.q.UpdateProduct(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) UpdateProductAvailability(ctx context.Context, id uuid.UUID, avail sqlc.ProductAvail, tenantID uuid.UUID) (*sqlc.Product, error) {
	p, err := r.q.UpdateProductAvailability(ctx, sqlc.UpdateProductAvailabilityParams{
		ID:           id,
		Availability: avail,
		TenantID:     tenantID,
	})
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) UpdateProductHasModifiers(ctx context.Context, id uuid.UUID, has bool) error {
	return r.q.UpdateProductHasModifiers(ctx, sqlc.UpdateProductHasModifiersParams{ID: id, HasModifiers: has})
}

func (r *Repository) DeleteProduct(ctx context.Context, id, tenantID uuid.UUID) error {
	return r.q.DeleteProduct(ctx, sqlc.DeleteProductParams{ID: id, TenantID: tenantID})
}

// --- Modifier group ---

func (r *Repository) CreateModifierGroup(ctx context.Context, arg sqlc.CreateModifierGroupParams) (*sqlc.ProductModifierGroup, error) {
	g, err := r.q.CreateModifierGroup(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *Repository) GetModifierGroupByID(ctx context.Context, id uuid.UUID) (*sqlc.ProductModifierGroup, error) {
	g, err := r.q.GetModifierGroupByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *Repository) ListModifierGroupsByProduct(ctx context.Context, productID uuid.UUID) ([]sqlc.ProductModifierGroup, error) {
	return r.q.ListModifierGroupsByProduct(ctx, productID)
}

func (r *Repository) UpdateModifierGroup(ctx context.Context, arg sqlc.UpdateModifierGroupParams) (*sqlc.ProductModifierGroup, error) {
	g, err := r.q.UpdateModifierGroup(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *Repository) DeleteModifierGroup(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteModifierGroup(ctx, id)
}

// --- Modifier option ---

func (r *Repository) CreateModifierOption(ctx context.Context, arg sqlc.CreateModifierOptionParams) (*sqlc.ProductModifierOption, error) {
	o, err := r.q.CreateModifierOption(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *Repository) ListModifierOptionsByGroup(ctx context.Context, groupID uuid.UUID) ([]sqlc.ProductModifierOption, error) {
	return r.q.ListModifierOptionsByGroup(ctx, groupID)
}

func (r *Repository) UpdateModifierOption(ctx context.Context, arg sqlc.UpdateModifierOptionParams) (*sqlc.ProductModifierOption, error) {
	o, err := r.q.UpdateModifierOption(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *Repository) DeleteModifierOption(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteModifierOption(ctx, id)
}

func (r *Repository) DeleteModifierOptionsByGroup(ctx context.Context, groupID uuid.UUID) error {
	return r.q.DeleteModifierOptionsByGroup(ctx, groupID)
}

// --- Discount ---

func (r *Repository) UpsertProductDiscount(ctx context.Context, arg sqlc.UpsertProductDiscountParams) (*sqlc.ProductDiscount, error) {
	d, err := r.q.UpsertProductDiscount(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *Repository) GetActiveDiscount(ctx context.Context, productID uuid.UUID) (*sqlc.ProductDiscount, error) {
	d, err := r.q.GetActiveDiscount(ctx, productID)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *Repository) DeactivateProductDiscount(ctx context.Context, productID uuid.UUID) error {
	return r.q.DeactivateProductDiscount(ctx, productID)
}
