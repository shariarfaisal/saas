-- name: ListWalletTransactions :many
SELECT * FROM wallet_transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: CountWalletTransactions :one
SELECT COUNT(*) FROM wallet_transactions WHERE user_id = $1;
