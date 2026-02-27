-- name: CreateHub :one
INSERT INTO hubs (tenant_id, name, code, manager_id, address_line1, address_line2, city, geo_lat, geo_lng, contact_phone, contact_email, is_active, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetHubByID :one
SELECT * FROM hubs WHERE id = $1 AND tenant_id = $2 LIMIT 1;

-- name: ListHubsByTenant :many
SELECT * FROM hubs WHERE tenant_id = $1 ORDER BY sort_order, name;

-- name: UpdateHub :one
UPDATE hubs SET
  name = COALESCE(sqlc.narg(name), name),
  code = COALESCE(sqlc.narg(code), code),
  address_line1 = COALESCE(sqlc.narg(address_line1), address_line1),
  address_line2 = COALESCE(sqlc.narg(address_line2), address_line2),
  city = COALESCE(sqlc.narg(city), city),
  contact_phone = COALESCE(sqlc.narg(contact_phone), contact_phone),
  contact_email = COALESCE(sqlc.narg(contact_email), contact_email),
  is_active = COALESCE(sqlc.narg(is_active), is_active),
  sort_order = COALESCE(sqlc.narg(sort_order), sort_order)
WHERE id = sqlc.arg(id) AND tenant_id = sqlc.arg(tenant_id)
RETURNING *;

-- name: DeleteHub :exec
DELETE FROM hubs WHERE id = $1 AND tenant_id = $2;

-- name: CreateHubArea :one
INSERT INTO hub_coverage_areas (hub_id, tenant_id, name, slug, delivery_charge, min_order_amount, estimated_delivery_minutes, is_active, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetHubAreaByID :one
SELECT * FROM hub_coverage_areas WHERE id = $1 LIMIT 1;

-- name: GetHubAreaByName :one
SELECT * FROM hub_coverage_areas WHERE hub_id = $1 AND slug = $2 LIMIT 1;

-- name: ListHubAreas :many
SELECT * FROM hub_coverage_areas WHERE hub_id = $1 ORDER BY sort_order, name;

-- name: UpdateHubArea :one
UPDATE hub_coverage_areas SET
  name = COALESCE(sqlc.narg(name), name),
  delivery_charge = COALESCE(sqlc.narg(delivery_charge), delivery_charge),
  min_order_amount = COALESCE(sqlc.narg(min_order_amount), min_order_amount),
  estimated_delivery_minutes = COALESCE(sqlc.narg(estimated_delivery_minutes), estimated_delivery_minutes),
  is_active = COALESCE(sqlc.narg(is_active), is_active),
  sort_order = COALESCE(sqlc.narg(sort_order), sort_order)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteHubArea :exec
DELETE FROM hub_coverage_areas WHERE id = $1;

-- name: GetDeliveryZoneConfig :one
SELECT * FROM delivery_zone_configs WHERE tenant_id = $1 LIMIT 1;

-- name: UpsertDeliveryZoneConfig :one
INSERT INTO delivery_zone_configs (tenant_id, model, distance_tiers, free_delivery_threshold)
VALUES ($1, $2, $3, $4)
ON CONFLICT (tenant_id) DO UPDATE SET
  model = EXCLUDED.model,
  distance_tiers = EXCLUDED.distance_tiers,
  free_delivery_threshold = EXCLUDED.free_delivery_threshold
RETURNING *;
