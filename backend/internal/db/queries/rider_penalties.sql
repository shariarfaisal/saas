-- name: CreateRiderPenalty :one
INSERT INTO rider_penalties (rider_id, tenant_id, order_id, issue_id, reason, amount)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetPenaltyByID :one
SELECT * FROM rider_penalties WHERE id = $1 AND tenant_id = $2 LIMIT 1;

-- name: ListPenaltiesByRider :many
SELECT * FROM rider_penalties
WHERE rider_id = $1 AND tenant_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: UpdatePenaltyStatus :one
UPDATE rider_penalties SET
  status = $2,
  cleared_at = COALESCE(sqlc.narg(cleared_at), cleared_at),
  cleared_by = COALESCE(sqlc.narg(cleared_by), cleared_by)
WHERE id = sqlc.arg(id) AND tenant_id = sqlc.arg(tenant_id)
RETURNING *;

-- name: AppealPenalty :one
UPDATE rider_penalties SET
  status = 'appealed',
  appeal_note = $3,
  appealed_at = NOW()
WHERE id = $1 AND tenant_id = $2
RETURNING *;
