-- name: UpdatePickupStatus :one
UPDATE order_pickups SET status = $3, picked_at = NOW()
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: ListPickupsByOrder :many
SELECT * FROM order_pickups WHERE order_id = $1 AND tenant_id = $2;

-- name: GetPickupByOrderAndRestaurant :one
SELECT * FROM order_pickups
WHERE order_id = $1 AND restaurant_id = $2 AND tenant_id = $3
LIMIT 1;
