-- name: CreateTimelineEvent :one
INSERT INTO order_timeline_events (
    order_id, tenant_id, event_type, previous_status, new_status,
    description, actor_id, actor_type, metadata
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListTimelineByOrder :many
SELECT * FROM order_timeline_events
WHERE order_id = $1
ORDER BY created_at DESC;
