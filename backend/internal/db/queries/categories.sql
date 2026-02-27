-- name: CreateCategory :one
INSERT INTO categories (tenant_id, restaurant_id, parent_id, name, slug, description, image_url, icon_url, extra_prep_time_minutes, is_tobacco, is_active, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetCategoryByID :one
SELECT * FROM categories WHERE id = $1 AND tenant_id = $2 LIMIT 1;

-- name: ListCategoriesByRestaurant :many
SELECT * FROM categories WHERE restaurant_id = $1 AND tenant_id = $2 AND is_active = true ORDER BY sort_order, name;

-- name: UpdateCategory :one
UPDATE categories SET
  name = COALESCE(sqlc.narg(name), name),
  description = COALESCE(sqlc.narg(description), description),
  image_url = COALESCE(sqlc.narg(image_url), image_url),
  sort_order = COALESCE(sqlc.narg(sort_order), sort_order),
  is_active = COALESCE(sqlc.narg(is_active), is_active)
WHERE id = sqlc.arg(id) AND tenant_id = sqlc.arg(tenant_id)
RETURNING *;

-- name: DeleteCategory :exec
UPDATE categories SET is_active = false WHERE id = $1 AND tenant_id = $2;

-- name: UpdateCategorySortOrder :exec
UPDATE categories SET sort_order = $2 WHERE id = $1 AND tenant_id = $3;
