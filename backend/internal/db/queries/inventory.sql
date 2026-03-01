-- ============================================================
-- Inventory SQLC Queries
-- ============================================================

-- name: GetInventoryItem :one
SELECT * FROM inventory_items
WHERE id = $1 AND tenant_id = $2;

-- name: GetInventoryByProductAndRestaurant :one
SELECT * FROM inventory_items
WHERE product_id = $1 AND restaurant_id = $2 AND tenant_id = $3;

-- name: ListInventoryByRestaurant :many
SELECT * FROM inventory_items
WHERE restaurant_id = $1 AND tenant_id = $2
ORDER BY updated_at DESC
LIMIT $3 OFFSET $4;

-- name: CountInventoryByRestaurant :one
SELECT COUNT(*) FROM inventory_items
WHERE restaurant_id = $1 AND tenant_id = $2;

-- name: ListLowStock :many
SELECT * FROM inventory_items
WHERE tenant_id = $1 AND restaurant_id = $2
  AND stock_qty - reserved_qty <= reorder_threshold
ORDER BY (stock_qty - reserved_qty) ASC
LIMIT $3 OFFSET $4;

-- name: CountLowStock :one
SELECT COUNT(*) FROM inventory_items
WHERE tenant_id = $1 AND restaurant_id = $2
  AND stock_qty - reserved_qty <= reorder_threshold;

-- name: AdjustStock :one
UPDATE inventory_items
SET stock_qty = stock_qty + sqlc.arg(qty_change)::INT,
    last_restocked_at = CASE WHEN sqlc.arg(qty_change)::INT > 0 THEN NOW() ELSE last_restocked_at END
WHERE id = sqlc.arg(id) AND tenant_id = sqlc.arg(tenant_id)
  AND stock_qty + sqlc.arg(qty_change)::INT >= 0
RETURNING *;

-- name: ReserveStock :one
UPDATE inventory_items
SET reserved_qty = reserved_qty + sqlc.arg(qty)::INT
WHERE id = (
    SELECT id FROM inventory_items
    WHERE product_id = sqlc.arg(product_id) AND restaurant_id = sqlc.arg(restaurant_id)
      AND tenant_id = sqlc.arg(tenant_id)
      AND stock_qty - reserved_qty >= sqlc.arg(qty)::INT
    LIMIT 1
    FOR UPDATE SKIP LOCKED
)
RETURNING *;

-- name: ReleaseStock :one
UPDATE inventory_items
SET reserved_qty = reserved_qty - sqlc.arg(qty)::INT
WHERE product_id = sqlc.arg(product_id) AND restaurant_id = sqlc.arg(restaurant_id)
  AND tenant_id = sqlc.arg(tenant_id)
  AND reserved_qty >= sqlc.arg(qty)::INT
RETURNING *;

-- name: ConsumeReservedStock :one
UPDATE inventory_items
SET stock_qty = stock_qty - sqlc.arg(qty)::INT,
    reserved_qty = reserved_qty - sqlc.arg(qty)::INT
WHERE product_id = sqlc.arg(product_id) AND restaurant_id = sqlc.arg(restaurant_id)
  AND tenant_id = sqlc.arg(tenant_id)
  AND reserved_qty >= sqlc.arg(qty)::INT
  AND stock_qty >= sqlc.arg(qty)::INT
RETURNING *;

-- name: GetInventoryForUpdate :one
SELECT * FROM inventory_items
WHERE product_id = $1 AND restaurant_id = $2
FOR UPDATE;

-- name: CreateInventoryAdjustment :one
INSERT INTO inventory_adjustments (
    inventory_item_id, tenant_id, restaurant_id, order_id,
    adjustment_type, qty_before, qty_change, qty_after,
    cost_price, note, adjusted_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: ListInventoryAdjustments :many
SELECT * FROM inventory_adjustments
WHERE inventory_item_id = $1 AND tenant_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CreateInventoryItem :one
INSERT INTO inventory_items (
    product_id, restaurant_id, tenant_id, stock_qty,
    reserved_qty, cost_price, reorder_threshold
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;
