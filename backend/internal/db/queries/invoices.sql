-- name: CreateInvoice :one
INSERT INTO invoices (
    tenant_id, restaurant_id, invoice_number, period_start, period_end,
    gross_sales, item_discounts, vendor_promo_discounts, net_sales, vat_collected,
    commission_rate, commission_amount, penalty_amount, adjustment_amount, adjustment_note,
    net_payable, total_orders, delivered_orders, cancelled_orders, rejected_orders,
    status, generated_by, notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
) RETURNING *;

-- name: GetInvoiceByID :one
SELECT * FROM invoices WHERE id = $1 AND tenant_id = $2 LIMIT 1;

-- name: ListInvoicesByRestaurant :many
SELECT * FROM invoices
WHERE restaurant_id = $1 AND tenant_id = $2
ORDER BY period_end DESC
LIMIT $3 OFFSET $4;

-- name: CountInvoicesByRestaurant :one
SELECT COUNT(*) FROM invoices WHERE restaurant_id = $1 AND tenant_id = $2;

-- name: ListInvoicesByTenant :many
SELECT * FROM invoices
WHERE tenant_id = $1
ORDER BY period_end DESC
LIMIT $2 OFFSET $3;

-- name: CountInvoicesByTenant :one
SELECT COUNT(*) FROM invoices WHERE tenant_id = $1;

-- name: FinalizeInvoice :one
UPDATE invoices SET status = 'finalized', finalized_by = $3, finalized_at = NOW(), notes = COALESCE($4, notes)
WHERE id = $1 AND tenant_id = $2 AND status = 'draft'
RETURNING *;

-- name: MarkInvoicePaid :one
UPDATE invoices SET status = 'paid', paid_by = $3, paid_at = NOW(), payment_reference = $4, notes = COALESCE($5, notes)
WHERE id = $1 AND tenant_id = $2 AND status = 'finalized'
RETURNING *;

-- name: GetInvoiceByPeriod :one
SELECT * FROM invoices
WHERE tenant_id = $1 AND restaurant_id = $2 AND period_start = $3 AND period_end = $4
LIMIT 1;

-- name: GetFinanceSummary :one
SELECT
    COALESCE(SUM(CASE WHEN status = 'draft' THEN net_payable ELSE 0 END), 0)::NUMERIC(14,2) AS pending_amount,
    COALESCE(SUM(CASE WHEN status = 'finalized' THEN net_payable ELSE 0 END), 0)::NUMERIC(14,2) AS finalized_amount,
    COALESCE(SUM(CASE WHEN status = 'paid' THEN net_payable ELSE 0 END), 0)::NUMERIC(14,2) AS paid_amount,
    COUNT(CASE WHEN status = 'draft' THEN 1 END)::INT AS draft_count,
    COUNT(CASE WHEN status = 'finalized' THEN 1 END)::INT AS finalized_count,
    COUNT(CASE WHEN status = 'paid' THEN 1 END)::INT AS paid_count
FROM invoices WHERE tenant_id = $1;

-- name: ListOrderItemsByOrderIDs :many
SELECT * FROM order_items
WHERE order_id = ANY(sqlc.arg(order_ids)::uuid[])
  AND restaurant_id = $1;

-- name: ListOrderItemsByRestaurantAndPeriod :many
SELECT oi.* FROM order_items oi
JOIN orders o ON oi.order_id = o.id
WHERE o.tenant_id = $1
  AND oi.restaurant_id = $2
  AND o.status = 'delivered'
  AND o.deleted_at IS NULL
  AND o.delivered_at >= sqlc.arg(period_start)::timestamptz
  AND o.delivered_at < sqlc.arg(period_end)::timestamptz;

-- name: CountOrdersByRestaurantAndPeriod :one
SELECT
    COUNT(*)::INT AS total_orders,
    COUNT(CASE WHEN status = 'delivered' THEN 1 END)::INT AS delivered_orders,
    COUNT(CASE WHEN status = 'cancelled' THEN 1 END)::INT AS cancelled_orders,
    COUNT(CASE WHEN status = 'rejected' THEN 1 END)::INT AS rejected_orders
FROM orders
WHERE tenant_id = $1
  AND deleted_at IS NULL
  AND created_at >= sqlc.arg(period_start)::timestamptz
  AND created_at < sqlc.arg(period_end)::timestamptz;

-- name: GetOrderPickupsByRestaurantAndPeriod :many
SELECT op.* FROM order_pickups op
JOIN orders o ON op.order_id = o.id
WHERE o.tenant_id = $1
  AND op.restaurant_id = $2
  AND o.status = 'delivered'
  AND o.deleted_at IS NULL
  AND o.delivered_at >= sqlc.arg(period_start)::timestamptz
  AND o.delivered_at < sqlc.arg(period_end)::timestamptz;

-- name: ListActiveRestaurantsByTenant :many
SELECT id, tenant_id, name FROM restaurants
WHERE tenant_id = $1 AND deleted_at IS NULL AND is_active = true;
