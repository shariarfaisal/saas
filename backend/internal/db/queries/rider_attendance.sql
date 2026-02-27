-- name: CreateAttendance :one
INSERT INTO rider_attendance (rider_id, tenant_id, work_date, checked_in_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetActiveAttendance :one
SELECT * FROM rider_attendance
WHERE rider_id = $1 AND checked_out_at IS NULL
ORDER BY checked_in_at DESC
LIMIT 1;

-- name: UpdateAttendanceCheckout :one
UPDATE rider_attendance SET
  checked_out_at = $2,
  total_hours = $3,
  total_distance_km = $4
WHERE id = $1
RETURNING *;

-- name: UpdateAttendanceStats :exec
UPDATE rider_attendance SET
  completed_orders = completed_orders + 1,
  earnings = earnings + $2
WHERE id = $1;

-- name: ListAttendanceByRider :many
SELECT * FROM rider_attendance
WHERE rider_id = $1
ORDER BY work_date DESC
LIMIT $2 OFFSET $3;

-- name: ListAttendanceByTenant :many
SELECT * FROM rider_attendance
WHERE tenant_id = $1 AND work_date = sqlc.arg(work_date)::date
ORDER BY checked_in_at DESC
LIMIT $2 OFFSET $3;
