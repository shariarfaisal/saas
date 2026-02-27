-- name: CreateOrderAnalytics :one
INSERT INTO order_analytics (
    tenant_id, order_id, restaurant_ids, customer_id, rider_id, hub_id,
    delivery_area, payment_method, platform, promo_code,
    subtotal, item_discount, promo_discount, delivery_charge, vat_total, total_amount, commission_total,
    confirmation_duration_s, preparation_duration_s, pickup_to_delivery_s, total_fulfillment_s,
    final_status, cancellation_reason,
    order_date, order_hour, order_day_of_week, order_week, order_month, order_year,
    completed_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
    $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30
) RETURNING *;

-- name: GetOrderAnalyticsByOrderID :one
SELECT * FROM order_analytics WHERE order_id = $1 AND tenant_id = $2 LIMIT 1;

-- name: GetDashboardToday :one
SELECT
    COUNT(*)::INT AS total_orders,
    COUNT(CASE WHEN final_status = 'delivered' THEN 1 END)::INT AS delivered_orders,
    COUNT(CASE WHEN final_status = 'cancelled' THEN 1 END)::INT AS cancelled_orders,
    COUNT(CASE WHEN final_status = 'rejected' THEN 1 END)::INT AS rejected_orders,
    COALESCE(SUM(CASE WHEN final_status = 'delivered' THEN total_amount ELSE 0 END), 0)::NUMERIC(14,2) AS today_revenue,
    COALESCE(AVG(CASE WHEN final_status = 'delivered' THEN total_fulfillment_s END), 0)::INT AS avg_delivery_time_s
FROM order_analytics
WHERE tenant_id = $1 AND order_date = CURRENT_DATE;

-- name: GetDashboardTrend :many
SELECT
    order_date,
    COUNT(*)::INT AS order_count,
    COALESCE(SUM(CASE WHEN final_status = 'delivered' THEN total_amount ELSE 0 END), 0)::NUMERIC(14,2) AS revenue
FROM order_analytics
WHERE tenant_id = $1 AND order_date >= sqlc.arg(start_date)::date AND order_date <= sqlc.arg(end_date)::date
GROUP BY order_date
ORDER BY order_date ASC;

-- name: GetTopProducts :many
SELECT
    oi.product_name,
    oi.product_id,
    SUM(oi.quantity)::INT AS total_quantity,
    SUM(oi.item_total)::NUMERIC(14,2) AS total_revenue
FROM order_items oi
JOIN orders o ON oi.order_id = o.id
WHERE o.tenant_id = $1
  AND o.status = 'delivered'
  AND o.delivered_at >= sqlc.arg(start_date)::timestamptz
  AND o.delivered_at < sqlc.arg(end_date)::timestamptz
GROUP BY oi.product_id, oi.product_name
ORDER BY total_quantity DESC
LIMIT $2;

-- name: GetSalesReport :many
SELECT
    order_date,
    COUNT(*)::INT AS order_count,
    COALESCE(SUM(subtotal), 0)::NUMERIC(14,2) AS gross_sales,
    COALESCE(SUM(item_discount), 0)::NUMERIC(14,2) AS item_discounts,
    COALESCE(SUM(promo_discount), 0)::NUMERIC(14,2) AS promo_discounts,
    COALESCE(SUM(subtotal - item_discount - promo_discount), 0)::NUMERIC(14,2) AS net_sales,
    COALESCE(SUM(vat_total), 0)::NUMERIC(14,2) AS vat_total,
    COALESCE(SUM(commission_total), 0)::NUMERIC(14,2) AS commission,
    COALESCE(SUM(total_amount), 0)::NUMERIC(14,2) AS total_revenue,
    COALESCE(AVG(total_amount), 0)::NUMERIC(14,2) AS avg_order_value,
    COALESCE(AVG(total_fulfillment_s), 0)::INT AS avg_delivery_time_s
FROM order_analytics
WHERE tenant_id = $1
  AND order_date >= sqlc.arg(start_date)::date
  AND order_date <= sqlc.arg(end_date)::date
  AND final_status = 'delivered'
GROUP BY order_date
ORDER BY order_date ASC;

-- name: GetPeakHours :many
SELECT
    order_hour,
    COUNT(*)::INT AS order_count
FROM order_analytics
WHERE tenant_id = $1
  AND order_date >= sqlc.arg(start_date)::date
  AND order_date <= sqlc.arg(end_date)::date
GROUP BY order_hour
ORDER BY order_hour ASC;

-- name: GetOrderStatusBreakdown :many
SELECT
    final_status,
    COUNT(*)::INT AS order_count
FROM order_analytics
WHERE tenant_id = $1
  AND order_date >= sqlc.arg(start_date)::date
  AND order_date <= sqlc.arg(end_date)::date
GROUP BY final_status;

-- name: GetRiderAnalytics :many
SELECT
    rider_id,
    COUNT(*)::INT AS total_orders,
    COALESCE(AVG(total_fulfillment_s), 0)::INT AS avg_delivery_time_s,
    COALESCE(SUM(delivery_charge), 0)::NUMERIC(14,2) AS total_delivery_revenue
FROM order_analytics
WHERE tenant_id = $1
  AND rider_id IS NOT NULL
  AND order_date >= sqlc.arg(start_date)::date
  AND order_date <= sqlc.arg(end_date)::date
GROUP BY rider_id
ORDER BY total_orders DESC;

-- name: GetAdminOverview :one
SELECT
    COUNT(*)::INT AS total_orders,
    COALESCE(SUM(CASE WHEN final_status = 'delivered' THEN total_amount ELSE 0 END), 0)::NUMERIC(14,2) AS total_revenue,
    COALESCE(SUM(commission_total), 0)::NUMERIC(14,2) AS total_commission,
    COUNT(DISTINCT tenant_id)::INT AS active_tenants,
    COUNT(DISTINCT customer_id)::INT AS unique_customers
FROM order_analytics
WHERE order_date >= sqlc.arg(start_date)::date AND order_date <= sqlc.arg(end_date)::date;

-- name: GetAdminRevenueByPeriod :many
SELECT
    order_date,
    COALESCE(SUM(commission_total), 0)::NUMERIC(14,2) AS commission_revenue,
    COALESCE(SUM(total_amount), 0)::NUMERIC(14,2) AS gross_revenue,
    COUNT(*)::INT AS order_count
FROM order_analytics
WHERE order_date >= sqlc.arg(start_date)::date AND order_date <= sqlc.arg(end_date)::date
GROUP BY order_date
ORDER BY order_date ASC;

-- name: GetAdminOrderVolume :many
SELECT
    order_date,
    COUNT(*)::INT AS order_count,
    COUNT(CASE WHEN final_status = 'delivered' THEN 1 END)::INT AS delivered_count,
    COUNT(CASE WHEN final_status = 'cancelled' THEN 1 END)::INT AS cancelled_count
FROM order_analytics
WHERE order_date >= sqlc.arg(start_date)::date AND order_date <= sqlc.arg(end_date)::date
GROUP BY order_date
ORDER BY order_date ASC;

-- name: GetTenantAnalytics :one
SELECT
    COUNT(*)::INT AS total_orders,
    COALESCE(SUM(CASE WHEN final_status = 'delivered' THEN total_amount ELSE 0 END), 0)::NUMERIC(14,2) AS total_revenue,
    COALESCE(SUM(commission_total), 0)::NUMERIC(14,2) AS total_commission,
    COUNT(DISTINCT customer_id)::INT AS unique_customers,
    COALESCE(AVG(CASE WHEN final_status = 'delivered' THEN total_fulfillment_s END), 0)::INT AS avg_delivery_time_s
FROM order_analytics
WHERE tenant_id = $1
  AND order_date >= sqlc.arg(start_date)::date
  AND order_date <= sqlc.arg(end_date)::date;

-- name: GetPendingOrderCount :one
SELECT COUNT(*)::INT FROM orders
WHERE tenant_id = $1 AND status IN ('pending', 'created') AND deleted_at IS NULL;
