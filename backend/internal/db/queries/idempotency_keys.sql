-- name: GetIdempotencyKey :one
SELECT * FROM idempotency_keys
WHERE tenant_id IS NOT DISTINCT FROM $1
  AND user_id IS NOT DISTINCT FROM $2
  AND key = $3
  AND endpoint = $4
LIMIT 1;

-- name: CreateIdempotencyKey :one
INSERT INTO idempotency_keys (tenant_id, user_id, key, endpoint, request_hash, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateIdempotencyKeyResponse :exec
UPDATE idempotency_keys SET response_status = $2, response_body = $3 WHERE id = $1;
