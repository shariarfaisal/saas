-- name: CreateLedgerAccount :one
INSERT INTO ledger_accounts (code, name, account_type, description, is_system)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetLedgerAccountByCode :one
SELECT * FROM ledger_accounts WHERE code = $1 LIMIT 1;

-- name: ListLedgerAccounts :many
SELECT * FROM ledger_accounts ORDER BY code;

-- name: CreateLedgerEntry :one
INSERT INTO ledger_entries (
    tenant_id, account_id, entry_type, reference_type, reference_id,
    debit, credit, balance_after, description, metadata
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: ListLedgerEntriesByAccount :many
SELECT * FROM ledger_entries
WHERE account_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetLastLedgerEntryBalance :one
SELECT balance_after FROM ledger_entries
WHERE account_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: ListLedgerEntriesByReference :many
SELECT * FROM ledger_entries
WHERE reference_type = $1 AND reference_id = $2
ORDER BY created_at ASC;
