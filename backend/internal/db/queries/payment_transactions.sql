-- name: CreateTransaction :one
INSERT INTO payment_transactions (
    tenant_id, order_id, user_id, payment_method, status, amount, currency,
    gateway_transaction_id, gateway_reference_id, gateway_response, gateway_fee,
    ip_address, user_agent
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: UpdateTransactionStatus :one
UPDATE payment_transactions SET
    status = COALESCE(sqlc.narg(status), status),
    gateway_transaction_id = COALESCE(sqlc.narg(gateway_transaction_id), gateway_transaction_id),
    gateway_reference_id = COALESCE(sqlc.narg(gateway_reference_id), gateway_reference_id),
    gateway_response = COALESCE(sqlc.narg(gateway_response), gateway_response),
    gateway_fee = COALESCE(sqlc.narg(gateway_fee), gateway_fee),
    callback_received_at = COALESCE(sqlc.narg(callback_received_at), callback_received_at)
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: GetTransactionByGatewayID :one
SELECT * FROM payment_transactions
WHERE gateway_transaction_id = $1 AND tenant_id = $2
LIMIT 1;

-- name: GetTransactionByID :one
SELECT * FROM payment_transactions
WHERE id = $1 AND tenant_id = $2
LIMIT 1;

-- name: GetTransactionByOrderID :one
SELECT * FROM payment_transactions
WHERE order_id = $1 AND tenant_id = $2 AND status = 'success'
LIMIT 1;

-- name: ListTransactionsByOrder :many
SELECT * FROM payment_transactions
WHERE order_id = $1
ORDER BY created_at DESC;

-- name: ListPendingTransactions :many
SELECT * FROM payment_transactions
WHERE status = 'pending' AND created_at < sqlc.arg(older_than)::timestamptz
ORDER BY created_at ASC
LIMIT $1;
