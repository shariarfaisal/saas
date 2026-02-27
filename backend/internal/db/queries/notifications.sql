-- name: ListNotifications :many
SELECT * FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: CountNotifications :one
SELECT COUNT(*) FROM notifications WHERE user_id = $1;

-- name: MarkNotificationRead :one
UPDATE notifications SET status = 'read' WHERE id = $1 AND user_id = $2 RETURNING *;

-- name: GetNotificationByID :one
SELECT * FROM notifications WHERE id = $1 AND user_id = $2 LIMIT 1;
