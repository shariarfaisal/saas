-- name: GetOrderByID :one
SELECT * FROM orders
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateOrderStatus :one
UPDATE orders SET status = $3
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: UpdateOrderPaymentStatus :one
UPDATE orders SET payment_status = $3
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: ListPendingPaymentOrders :many
SELECT * FROM orders
WHERE tenant_id = $1 AND payment_status = 'unpaid' AND status = 'pending'
    AND created_at < sqlc.arg(older_than)::timestamptz
ORDER BY created_at ASC
LIMIT $2;

-- name: GetOrderForReconciliation :many
SELECT * FROM orders
WHERE payment_status = 'unpaid' AND status IN ('pending', 'created')
    AND created_at < sqlc.arg(older_than)::timestamptz
ORDER BY created_at ASC
LIMIT $1;
