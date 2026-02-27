-- name: CreateRider :one
INSERT INTO riders (tenant_id, user_id, hub_id, vehicle_type, vehicle_registration, license_number, nid_number, nid_verified)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetRiderByID :one
SELECT * FROM riders WHERE id = $1 AND tenant_id = $2 LIMIT 1;

-- name: GetRiderByUserID :one
SELECT * FROM riders WHERE user_id = $1 AND tenant_id = $2 LIMIT 1;

-- name: ListRidersByTenant :many
SELECT * FROM riders WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: ListRidersByHub :many
SELECT * FROM riders WHERE hub_id = $1 AND tenant_id = $2 ORDER BY created_at DESC;

-- name: ListAvailableRidersByHub :many
SELECT * FROM riders WHERE hub_id = $1 AND tenant_id = $2 AND is_available = true AND is_on_duty = true;

-- name: UpdateRider :one
UPDATE riders SET
  hub_id = COALESCE(sqlc.narg(hub_id), hub_id),
  vehicle_type = COALESCE(sqlc.narg(vehicle_type), vehicle_type),
  vehicle_registration = COALESCE(sqlc.narg(vehicle_registration), vehicle_registration),
  license_number = COALESCE(sqlc.narg(license_number), license_number),
  nid_number = COALESCE(sqlc.narg(nid_number), nid_number),
  nid_verified = COALESCE(sqlc.narg(nid_verified), nid_verified),
  rating_avg = COALESCE(sqlc.narg(rating_avg), rating_avg),
  rating_count = COALESCE(sqlc.narg(rating_count), rating_count)
WHERE id = sqlc.arg(id) AND tenant_id = sqlc.arg(tenant_id)
RETURNING *;

-- name: DeleteRider :exec
DELETE FROM riders WHERE id = $1 AND tenant_id = $2;

-- name: UpdateRiderAvailability :one
UPDATE riders SET is_available = $3
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: UpdateRiderDutyStatus :one
UPDATE riders SET is_on_duty = $3
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: UpdateRiderStats :exec
UPDATE riders SET
  total_order_count = total_order_count + 1,
  total_earnings = total_earnings + $3,
  pending_balance = pending_balance + $4
WHERE id = $1 AND tenant_id = $2;

-- name: CountRidersByTenant :one
SELECT COUNT(*) FROM riders WHERE tenant_id = $1;
