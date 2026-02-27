-- name: CreateOrderIssue :one
INSERT INTO order_issues (
    order_id, tenant_id, issue_type, reported_by_id, details,
    accountable_party
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetOrderIssueByID :one
SELECT * FROM order_issues WHERE id = $1 AND tenant_id = $2 LIMIT 1;
