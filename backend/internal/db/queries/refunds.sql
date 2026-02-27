-- name: CreateRefund :one
INSERT INTO refunds (
    tenant_id, order_id, transaction_id, issue_id, amount, reason, status,
    gateway_refund_id, approved_by, approved_at, processed_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: UpdateRefundStatus :one
UPDATE refunds SET
    status = COALESCE(sqlc.narg(status), status),
    gateway_refund_id = COALESCE(sqlc.narg(gateway_refund_id), gateway_refund_id),
    processed_at = COALESCE(sqlc.narg(processed_at), processed_at)
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: GetRefundByID :one
SELECT * FROM refunds
WHERE id = $1 AND tenant_id = $2
LIMIT 1;

-- name: ListRefundsByOrder :many
SELECT * FROM refunds
WHERE order_id = $1 AND tenant_id = $2
ORDER BY created_at DESC;

-- name: ApproveRefund :one
UPDATE refunds SET
    status = 'approved',
    approved_by = sqlc.arg(approved_by),
    approved_at = sqlc.arg(approved_at)
WHERE id = $1 AND tenant_id = $2
RETURNING *;
