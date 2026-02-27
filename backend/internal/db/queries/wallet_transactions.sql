-- name: ListWalletTransactions :many
SELECT * FROM wallet_transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: CountWalletTransactions :one
SELECT COUNT(*) FROM wallet_transactions WHERE user_id = $1;

-- name: CreateWalletTransaction :one
INSERT INTO wallet_transactions (
    user_id, tenant_id, order_id, type, source, amount, balance_after, description, expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: CreditUserWallet :exec
UPDATE users SET wallet_balance = wallet_balance + sqlc.arg(amount) WHERE id = $1 AND deleted_at IS NULL;
