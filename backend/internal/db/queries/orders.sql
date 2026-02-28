-- ============================================================
-- Orders SQLC Queries
-- ============================================================

-- name: CreateOrder :one
INSERT INTO orders (
    tenant_id, order_number, customer_id, status, payment_status,
    payment_method, platform, delivery_address_id, delivery_address,
    delivery_recipient_name, delivery_recipient_phone, delivery_area,
    delivery_geo_lat, delivery_geo_lng, subtotal, item_discount_total,
    promo_discount_total, vat_total, delivery_charge, service_fee,
    total_amount, promo_id, promo_code, promo_snapshot,
    is_priority, is_reorder, customer_note, auto_confirm_at,
    estimated_delivery_minutes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
    $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29
)
RETURNING *;

-- name: CreateOrderItem :one
INSERT INTO order_items (
    order_id, restaurant_id, product_id, tenant_id, product_name,
    product_snapshot, quantity, unit_price, modifier_price,
    item_subtotal, item_discount, item_vat, promo_discount,
    item_total, selected_modifiers, special_instructions
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING *;

-- name: CreateOrderPickup :one
INSERT INTO order_pickups (
    order_id, restaurant_id, tenant_id, pickup_number, status,
    items_subtotal, items_discount, items_vat, items_total,
    commission_rate, commission_amount
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM orders
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: GetOrderByNumber :one
SELECT * FROM orders
WHERE order_number = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: ListOrdersByCustomer :many
SELECT * FROM orders
WHERE customer_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountOrdersByCustomer :one
SELECT COUNT(*) FROM orders
WHERE customer_id = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: ListOrdersByTenant :many
SELECT * FROM orders
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountOrdersByTenant :one
SELECT COUNT(*) FROM orders
WHERE tenant_id = $1 AND deleted_at IS NULL;

-- name: ListOrdersByRestaurant :many
SELECT DISTINCT o.* FROM orders o
JOIN order_pickups op ON o.id = op.order_id
WHERE op.restaurant_id = $1 AND o.tenant_id = $2 AND o.deleted_at IS NULL
ORDER BY o.created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountOrdersByRestaurant :one
SELECT COUNT(DISTINCT o.id) FROM orders o
JOIN order_pickups op ON o.id = op.order_id
WHERE op.restaurant_id = $1 AND o.tenant_id = $2 AND o.deleted_at IS NULL;

-- name: TransitionOrderStatus :one
UPDATE orders SET
    status = sqlc.arg(new_status),
    confirmed_at = CASE WHEN sqlc.arg(new_status) = 'confirmed' THEN NOW() ELSE confirmed_at END,
    preparing_at = CASE WHEN sqlc.arg(new_status) = 'preparing' THEN NOW() ELSE preparing_at END,
    ready_at = CASE WHEN sqlc.arg(new_status) = 'ready' THEN NOW() ELSE ready_at END,
    picked_at = CASE WHEN sqlc.arg(new_status) = 'picked' THEN NOW() ELSE picked_at END,
    delivered_at = CASE WHEN sqlc.arg(new_status) = 'delivered' THEN NOW() ELSE delivered_at END,
    cancelled_at = CASE WHEN sqlc.arg(new_status) = 'cancelled' THEN NOW() ELSE cancelled_at END,
    cancellation_reason = CASE WHEN sqlc.arg(new_status) = 'cancelled' THEN sqlc.narg(cancellation_reason) ELSE cancellation_reason END,
    cancelled_by = CASE WHEN sqlc.arg(new_status) = 'cancelled' THEN sqlc.narg(cancelled_by) ELSE cancelled_by END,
    rejection_reason = CASE WHEN sqlc.arg(new_status) = 'rejected' THEN sqlc.narg(rejection_reason) ELSE rejection_reason END,
    rejected_by = CASE WHEN sqlc.arg(new_status) = 'rejected' THEN sqlc.narg(rejected_by) ELSE rejected_by END
WHERE id = sqlc.arg(id) AND tenant_id = sqlc.arg(tenant_id) AND deleted_at IS NULL
RETURNING *;

-- name: UpdateOrderPaymentStatus :one
UPDATE orders SET
    payment_status = sqlc.arg(payment_status)
WHERE id = sqlc.arg(id) AND tenant_id = sqlc.arg(tenant_id) AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteOrder :exec
UPDATE orders SET deleted_at = NOW()
WHERE id = $1 AND tenant_id = $2;

-- name: AddTimelineEvent :one
INSERT INTO order_timeline_events (
    order_id, tenant_id, event_type, previous_status,
    new_status, description, actor_id, actor_type, metadata
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListTimelineEvents :many
SELECT * FROM order_timeline_events
WHERE order_id = $1 AND tenant_id = $2
ORDER BY created_at ASC;

-- name: GetOrderItemsByOrder :many
SELECT * FROM order_items
WHERE order_id = $1
ORDER BY created_at ASC;

-- name: GetOrderPickupsByOrder :many
SELECT * FROM order_pickups
WHERE order_id = $1
ORDER BY created_at ASC;

-- name: GetOrderPickup :one
SELECT * FROM order_pickups
WHERE order_id = $1 AND restaurant_id = $2;

-- name: TransitionPickupStatus :one
UPDATE order_pickups SET
    status = sqlc.arg(new_status),
    confirmed_at = CASE WHEN sqlc.arg(new_status) = 'confirmed' THEN NOW() ELSE confirmed_at END,
    preparing_at = CASE WHEN sqlc.arg(new_status) = 'preparing' THEN NOW() ELSE preparing_at END,
    ready_at = CASE WHEN sqlc.arg(new_status) = 'ready' THEN NOW() ELSE ready_at END,
    picked_at = CASE WHEN sqlc.arg(new_status) = 'picked' THEN NOW() ELSE picked_at END,
    rejected_at = CASE WHEN sqlc.arg(new_status) = 'rejected' THEN NOW() ELSE rejected_at END,
    rejection_reason = CASE WHEN sqlc.arg(new_status) = 'rejected' THEN sqlc.narg(rejection_reason) ELSE rejection_reason END
WHERE order_id = sqlc.arg(order_id) AND restaurant_id = sqlc.arg(restaurant_id)
RETURNING *;

-- name: GetOrderItemsByRestaurant :many
SELECT oi.* FROM order_items oi
WHERE oi.order_id = $1 AND oi.restaurant_id = $2
ORDER BY oi.created_at ASC;

-- name: ListPendingAutoConfirmOrders :many
SELECT * FROM orders
WHERE status = 'created'
  AND auto_confirm_at IS NOT NULL
  AND auto_confirm_at <= NOW()
  AND deleted_at IS NULL
ORDER BY auto_confirm_at ASC
LIMIT $1;

-- name: UpdateOrderStatus :one
UPDATE orders SET status = $3
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: GenerateOrderNumber :one
SELECT CONCAT(
    sqlc.arg(prefix)::TEXT, '-',
    LPAD((COALESCE(
        (SELECT COUNT(*) + 1 FROM orders WHERE tenant_id = sqlc.arg(tenant_id)),
        1
    ))::TEXT, 6, '0')
) AS order_number;

-- name: GetOrderForUpdate :one
SELECT * FROM orders
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
FOR UPDATE;

-- name: CheckAllPickupsInStatus :one
SELECT COUNT(*) = 0 AS all_in_status
FROM order_pickups
WHERE order_id = $1
  AND status != sqlc.arg(expected_status);

-- name: GetPickupCountByOrder :one
SELECT COUNT(*) FROM order_pickups
WHERE order_id = $1;

-- name: ListPendingPaymentOrders :many
SELECT * FROM orders
WHERE tenant_id = $1 AND payment_status = 'unpaid' AND status = 'pending'
    AND created_at < sqlc.arg(older_than)::timestamptz
ORDER BY created_at ASC
LIMIT $2;

-- name: ListActiveOrdersByRider :many
SELECT * FROM orders
WHERE rider_id = $1 AND tenant_id = $2
  AND status IN ('confirmed', 'preparing', 'ready', 'picked')
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListDeliveredOrdersByRider :many
SELECT * FROM orders
WHERE rider_id = $1 AND tenant_id = $2
  AND status = 'delivered'
  AND deleted_at IS NULL
ORDER BY delivered_at DESC
LIMIT $3 OFFSET $4;

-- name: AssignRiderToOrder :one
UPDATE orders SET rider_id = $3
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: GetOrderForReconciliation :many
SELECT * FROM orders
WHERE payment_status = 'unpaid' AND status IN ('pending', 'created')
    AND created_at < sqlc.arg(older_than)::timestamptz
ORDER BY created_at ASC
LIMIT $1;
