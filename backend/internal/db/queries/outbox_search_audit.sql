-- name: CreateNotification :one
INSERT INTO notifications (
    tenant_id, user_id, channel, title, body, image_url,
    action_type, action_payload, status
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetUserDevicePushToken :one
SELECT device_push_token FROM users WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: ClearUserPushToken :exec
UPDATE users SET device_push_token = NULL WHERE id = $1;

-- name: CreateOutboxEvent :one
INSERT INTO outbox_events (
    tenant_id, aggregate_type, aggregate_id, event_type,
    payload, status, max_attempts
) VALUES ($1, $2, $3, $4, $5, 'pending', $6)
RETURNING *;

-- name: ListPendingOutboxEvents :many
SELECT * FROM outbox_events
WHERE status IN ('pending', 'failed')
  AND (next_retry_at IS NULL OR next_retry_at <= NOW())
  AND attempts < max_attempts
ORDER BY created_at ASC
LIMIT $1;

-- name: MarkOutboxEventProcessed :exec
UPDATE outbox_events SET status = 'processed', processed_at = NOW(), attempts = attempts + 1
WHERE id = $1;

-- name: MarkOutboxEventFailed :exec
UPDATE outbox_events SET
    status = CASE WHEN attempts + 1 >= max_attempts THEN 'dead_letter'::outbox_event_status ELSE 'failed'::outbox_event_status END,
    attempts = attempts + 1,
    last_error = $2,
    next_retry_at = NOW() + (POWER(2, attempts + 1) || ' minutes')::interval
WHERE id = $1;

-- name: CreateAuditLog :one
INSERT INTO audit_logs (
    tenant_id, actor_id, actor_type, action, resource_type, resource_id, changes, reason, ip_address
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListAuditLogsByResource :many
SELECT * FROM audit_logs
WHERE resource_type = $1 AND resource_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CreateSearchLog :one
INSERT INTO search_logs (
    tenant_id, user_id, query, search_type, result_count, filters
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetTopSearchTerms :many
SELECT query, COUNT(*)::INT AS search_count
FROM search_logs
WHERE tenant_id = $1 AND created_at >= sqlc.arg(since)::timestamptz
GROUP BY query
ORDER BY search_count DESC
LIMIT $2;

-- name: SearchRestaurants :many
SELECT * FROM restaurants
WHERE tenant_id = $1
  AND deleted_at IS NULL
  AND is_active = true
  AND (name ILIKE '%' || sqlc.arg(query)::text || '%' OR description ILIKE '%' || sqlc.arg(query)::text || '%')
ORDER BY name ASC
LIMIT $2;

-- name: SearchProducts :many
SELECT p.* FROM products p
JOIN restaurants r ON p.restaurant_id = r.id
WHERE p.tenant_id = $1
  AND p.deleted_at IS NULL
  AND p.availability = 'available'
  AND r.is_active = true
  AND r.deleted_at IS NULL
  AND (p.name ILIKE '%' || sqlc.arg(query)::text || '%' OR p.description ILIKE '%' || sqlc.arg(query)::text || '%')
ORDER BY p.name ASC
LIMIT $2;

-- name: PurgeOldNotifications :exec
DELETE FROM notifications WHERE created_at < sqlc.arg(before)::timestamptz;

-- name: PurgeOldOrderTimeline :exec
DELETE FROM order_timeline_events WHERE created_at < sqlc.arg(before)::timestamptz;

-- name: PurgeOldSearchLogs :exec
DELETE FROM search_logs WHERE created_at < sqlc.arg(before)::timestamptz;

-- name: PurgeOldAuditLogs :exec
DELETE FROM audit_logs WHERE created_at < sqlc.arg(before)::timestamptz;

-- name: ListOrderIssuesByTenant :many
SELECT * FROM order_issues
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountOrderIssuesByTenant :one
SELECT COUNT(*) FROM order_issues WHERE tenant_id = $1;

-- name: ListOrderIssuesByOrder :many
SELECT * FROM order_issues
WHERE order_id = $1 AND tenant_id = $2
ORDER BY created_at DESC;

-- name: UpdateOrderIssueStatus :one
UPDATE order_issues SET status = $3, resolution_note = $4, resolved_by_id = $5, resolved_at = NOW()
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: UpdateOrderIssueRefund :one
UPDATE order_issues SET refund_status = $3, refund_amount = $4
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: CreateOrderIssueMessage :one
INSERT INTO order_issue_messages (issue_id, tenant_id, sender_id, message, attachments)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListOrderIssueMessages :many
SELECT * FROM order_issue_messages
WHERE issue_id = $1 AND tenant_id = $2
ORDER BY created_at ASC;

-- name: GetUserWalletBalance :one
SELECT wallet_balance FROM users WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: DebitUserWallet :exec
UPDATE users SET wallet_balance = wallet_balance - sqlc.arg(amount)
WHERE id = $1 AND deleted_at IS NULL AND wallet_balance >= sqlc.arg(amount);

-- name: ListOrdersByStatus :many
SELECT * FROM orders
WHERE tenant_id = $1 AND status = sqlc.arg(order_status)::order_status AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListCreatedOrdersPastTimeout :many
SELECT * FROM orders
WHERE status = 'created'
  AND auto_confirm_at IS NOT NULL
  AND auto_confirm_at < NOW()
  AND deleted_at IS NULL
LIMIT $1;

-- name: ListPendingOrdersPastTimeout :many
SELECT * FROM orders
WHERE status = 'pending'
  AND created_at < sqlc.arg(older_than)::timestamptz
  AND deleted_at IS NULL
LIMIT $1;

-- name: GetNotificationPreferences :one
SELECT * FROM notification_preferences WHERE user_id = $1 LIMIT 1;
