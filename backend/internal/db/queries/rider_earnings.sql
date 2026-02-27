-- name: CreateRiderEarning :one
INSERT INTO rider_earnings (rider_id, tenant_id, order_id, base_earning, distance_bonus, peak_bonus, tip_amount, total_earning)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListEarningsByRider :many
SELECT * FROM rider_earnings
WHERE rider_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetTotalEarningsByRider :one
SELECT COALESCE(SUM(total_earning), 0)::NUMERIC(10,2) AS total
FROM rider_earnings
WHERE rider_id = $1 AND is_paid_out = false;

-- name: ListEarningsByOrder :many
SELECT * FROM rider_earnings WHERE order_id = $1 AND tenant_id = $2;
