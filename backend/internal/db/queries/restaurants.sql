-- name: CreateRestaurant :one
INSERT INTO restaurants (tenant_id, hub_id, owner_id, name, slug, type, description, short_description, banner_image_url, logo_url, gallery_urls, phone, email, address_line1, address_line2, area, city, cuisines, tags, commission_rate, vat_rate, is_vat_inclusive, min_order_amount, avg_prep_time_minutes, max_concurrent_orders, auto_accept_orders, order_prefix, is_available, is_featured, is_active, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31)
RETURNING *;

-- name: GetRestaurantByID :one
SELECT * FROM restaurants WHERE id = $1 AND tenant_id = $2 LIMIT 1;

-- name: GetRestaurantBySlug :one
SELECT * FROM restaurants WHERE tenant_id = $1 AND slug = $2 LIMIT 1;

-- name: ListRestaurantsByTenant :many
SELECT * FROM restaurants WHERE tenant_id = $1 AND is_active = true ORDER BY sort_order, name LIMIT $2 OFFSET $3;

-- name: CountRestaurantsByTenant :one
SELECT COUNT(*) FROM restaurants WHERE tenant_id = $1 AND is_active = true;

-- name: ListAvailableByHubAndArea :many
SELECT r.* FROM restaurants r
JOIN hub_coverage_areas hca ON hca.hub_id = r.hub_id AND hca.slug = $2 AND hca.is_active = true
WHERE r.tenant_id = $1 AND r.is_available = true AND r.is_active = true
ORDER BY r.is_featured DESC, r.sort_order, r.name
LIMIT $3 OFFSET $4;

-- name: UpdateRestaurant :one
UPDATE restaurants SET
  name = COALESCE(sqlc.narg(name), name),
  description = COALESCE(sqlc.narg(description), description),
  short_description = COALESCE(sqlc.narg(short_description), short_description),
  banner_image_url = COALESCE(sqlc.narg(banner_image_url), banner_image_url),
  logo_url = COALESCE(sqlc.narg(logo_url), logo_url),
  phone = COALESCE(sqlc.narg(phone), phone),
  email = COALESCE(sqlc.narg(email), email),
  address_line1 = COALESCE(sqlc.narg(address_line1), address_line1),
  area = COALESCE(sqlc.narg(area), area),
  city = COALESCE(sqlc.narg(city), city),
  cuisines = COALESCE(sqlc.narg(cuisines), cuisines),
  tags = COALESCE(sqlc.narg(tags), tags),
  min_order_amount = COALESCE(sqlc.narg(min_order_amount), min_order_amount),
  avg_prep_time_minutes = COALESCE(sqlc.narg(avg_prep_time_minutes), avg_prep_time_minutes),
  auto_accept_orders = COALESCE(sqlc.narg(auto_accept_orders), auto_accept_orders),
  is_featured = COALESCE(sqlc.narg(is_featured), is_featured),
  sort_order = COALESCE(sqlc.narg(sort_order), sort_order)
WHERE id = sqlc.arg(id) AND tenant_id = sqlc.arg(tenant_id)
RETURNING *;

-- name: UpdateRestaurantAvailability :one
UPDATE restaurants SET is_available = $2 WHERE id = $1 AND tenant_id = $3 RETURNING *;

-- name: DeleteRestaurant :exec
UPDATE restaurants SET is_active = false WHERE id = $1 AND tenant_id = $2;

-- name: UpsertOperatingHour :one
INSERT INTO restaurant_operating_hours (restaurant_id, tenant_id, day_of_week, open_time, close_time, is_closed)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (restaurant_id, day_of_week) DO UPDATE SET
  open_time = EXCLUDED.open_time,
  close_time = EXCLUDED.close_time,
  is_closed = EXCLUDED.is_closed
RETURNING *;

-- name: ListOperatingHours :many
SELECT * FROM restaurant_operating_hours WHERE restaurant_id = $1 ORDER BY day_of_week;

-- name: DeleteOperatingHours :exec
DELETE FROM restaurant_operating_hours WHERE restaurant_id = $1;
