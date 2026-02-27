-- name: CreateAddress :one
INSERT INTO user_addresses (user_id, tenant_id, label, recipient_name, recipient_phone, address_line1, address_line2, area, city, geo_lat, geo_lng, geo_display_addr, is_default)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetAddressByID :one
SELECT * FROM user_addresses WHERE id = $1 AND user_id = $2 LIMIT 1;

-- name: ListAddresses :many
SELECT * FROM user_addresses WHERE user_id = $1 ORDER BY is_default DESC, created_at DESC;

-- name: UpdateAddress :one
UPDATE user_addresses SET
  label = COALESCE(sqlc.narg(label), label),
  recipient_name = COALESCE(sqlc.narg(recipient_name), recipient_name),
  recipient_phone = COALESCE(sqlc.narg(recipient_phone), recipient_phone),
  address_line1 = COALESCE(sqlc.narg(address_line1), address_line1),
  address_line2 = COALESCE(sqlc.narg(address_line2), address_line2),
  area = COALESCE(sqlc.narg(area), area),
  city = COALESCE(sqlc.narg(city), city),
  geo_lat = COALESCE(sqlc.narg(geo_lat), geo_lat),
  geo_lng = COALESCE(sqlc.narg(geo_lng), geo_lng),
  geo_display_addr = COALESCE(sqlc.narg(geo_display_addr), geo_display_addr),
  is_default = COALESCE(sqlc.narg(is_default), is_default)
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
RETURNING *;

-- name: DeleteAddress :exec
DELETE FROM user_addresses WHERE id = $1 AND user_id = $2;

-- name: ClearDefaultAddresses :exec
UPDATE user_addresses SET is_default = false WHERE user_id = $1;
