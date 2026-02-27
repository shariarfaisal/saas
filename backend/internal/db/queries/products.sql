-- name: CreateProduct :one
INSERT INTO products (tenant_id, restaurant_id, category_id, name, slug, description, base_price, vat_rate, availability, images, tags, is_featured, is_inv_tracked, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING *;

-- name: GetProductByID :one
SELECT * FROM products WHERE id = $1 AND tenant_id = $2 LIMIT 1;

-- name: GetProductByIDPublic :one
SELECT * FROM products WHERE id = $1 AND availability = 'available' LIMIT 1;

-- name: ListProductsByRestaurant :many
SELECT * FROM products WHERE restaurant_id = $1 AND tenant_id = $2 ORDER BY sort_order, name LIMIT $3 OFFSET $4;

-- name: CountProductsByRestaurant :one
SELECT COUNT(*) FROM products WHERE restaurant_id = $1 AND tenant_id = $2;

-- name: ListAvailableProductsByRestaurant :many
SELECT * FROM products WHERE restaurant_id = $1 AND availability = 'available' ORDER BY sort_order, name;

-- name: UpdateProduct :one
UPDATE products SET
  name = COALESCE(sqlc.narg(name), name),
  description = COALESCE(sqlc.narg(description), description),
  category_id = COALESCE(sqlc.narg(category_id), category_id),
  base_price = COALESCE(sqlc.narg(base_price), base_price),
  vat_rate = COALESCE(sqlc.narg(vat_rate), vat_rate),
  images = COALESCE(sqlc.narg(images), images),
  tags = COALESCE(sqlc.narg(tags), tags),
  is_featured = COALESCE(sqlc.narg(is_featured), is_featured),
  sort_order = COALESCE(sqlc.narg(sort_order), sort_order)
WHERE id = sqlc.arg(id) AND tenant_id = sqlc.arg(tenant_id)
RETURNING *;

-- name: UpdateProductAvailability :one
UPDATE products SET availability = $2 WHERE id = $1 AND tenant_id = $3 RETURNING *;

-- name: UpdateProductHasModifiers :exec
UPDATE products SET has_modifiers = $2 WHERE id = $1;

-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1 AND tenant_id = $2;

-- name: CreateModifierGroup :one
INSERT INTO product_modifier_groups (product_id, tenant_id, name, description, min_required, max_allowed, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetModifierGroupByID :one
SELECT * FROM product_modifier_groups WHERE id = $1 LIMIT 1;

-- name: ListModifierGroupsByProduct :many
SELECT * FROM product_modifier_groups WHERE product_id = $1 ORDER BY sort_order, name;

-- name: UpdateModifierGroup :one
UPDATE product_modifier_groups SET
  name = COALESCE(sqlc.narg(name), name),
  description = COALESCE(sqlc.narg(description), description),
  min_required = COALESCE(sqlc.narg(min_required), min_required),
  max_allowed = COALESCE(sqlc.narg(max_allowed), max_allowed),
  sort_order = COALESCE(sqlc.narg(sort_order), sort_order)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteModifierGroup :exec
DELETE FROM product_modifier_groups WHERE id = $1;

-- name: CreateModifierOption :one
INSERT INTO product_modifier_options (modifier_group_id, product_id, tenant_id, name, additional_price, is_available, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListModifierOptionsByGroup :many
SELECT * FROM product_modifier_options WHERE modifier_group_id = $1 ORDER BY sort_order, name;

-- name: UpdateModifierOption :one
UPDATE product_modifier_options SET
  name = COALESCE(sqlc.narg(name), name),
  additional_price = COALESCE(sqlc.narg(additional_price), additional_price),
  is_available = COALESCE(sqlc.narg(is_available), is_available),
  sort_order = COALESCE(sqlc.narg(sort_order), sort_order)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteModifierOption :exec
DELETE FROM product_modifier_options WHERE id = $1;

-- name: DeleteModifierOptionsByGroup :exec
DELETE FROM product_modifier_options WHERE modifier_group_id = $1;

-- name: UpsertProductDiscount :one
INSERT INTO product_discounts (product_id, restaurant_id, tenant_id, discount_type, amount, max_discount_cap, starts_at, ends_at, is_active, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (product_id) WHERE is_active = true DO UPDATE SET
  discount_type = EXCLUDED.discount_type,
  amount = EXCLUDED.amount,
  max_discount_cap = EXCLUDED.max_discount_cap,
  starts_at = EXCLUDED.starts_at,
  ends_at = EXCLUDED.ends_at,
  is_active = EXCLUDED.is_active
RETURNING *;

-- name: GetActiveDiscount :one
SELECT * FROM product_discounts
WHERE product_id = $1 AND is_active = true
  AND starts_at <= NOW()
  AND (ends_at IS NULL OR ends_at > NOW())
LIMIT 1;

-- name: DeactivateProductDiscount :exec
UPDATE product_discounts SET is_active = false WHERE product_id = $1 AND is_active = true;

-- name: ExpireDiscounts :exec
UPDATE product_discounts SET is_active = false WHERE ends_at IS NOT NULL AND ends_at < NOW() AND is_active = true;
