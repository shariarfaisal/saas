# 12 — Analytics & Reporting

## 12.1 Overview

Analytics are built on two layers:
1. **Operational data** — queried directly from the transactional tables (real-time, last 7 days)
2. **Analytics tables** — pre-aggregated/denormalised `order_analytics` table (historical, fast queries)

The `order_analytics` table is populated via a background job when an order reaches a terminal state (`delivered`, `cancelled`, `rejected`).

---

## 12.2 Partner Dashboard Metrics

### Today's KPIs (real-time, from transactional tables)
- Total orders (all statuses)
- Delivered orders
- Rejected orders
- Pending/active orders
- Today's revenue (sum of delivered order totals)
- Average order value
- Average delivery time (minutes)

### 7-Day Trend
- Orders per day (line chart)
- Revenue per day (line chart)
- Rejection rate per day

### Restaurant Performance
- Top 5 products (by order count, last 30 days)
- Peak ordering hours (heatmap by hour-of-day)
- Avg prep time vs promised prep time
- Acceptance rate (confirmed / total incoming)

---

## 12.3 Sales Report (Date Range)

Available via `/partner/reports/sales` with date range parameters.

| Metric | Description |
|--------|-------------|
| Gross Sales | Sum of item subtotals |
| Product Discounts | Restaurant-funded discounts |
| Promo Discounts | Promo code discounts applied |
| Net Sales | Gross - all discounts |
| VAT Collected | VAT on net sales |
| Commission (rate%) | Platform commission deducted |
| Net Payable | After all deductions |
| Total Orders | Count of delivered orders |
| Rejected Orders | Count of rejected orders |
| Avg Order Value | Net Sales / Total Orders |
| Avg Delivery Time | Minutes from confirmed to delivered |

**Grouped by:** Day / Week / Month (configurable)

---

## 12.4 Product Analytics

- **Top selling products** by revenue and quantity (per restaurant, per period)
- **Worst performing** products (low sales, high rejection)
- **Sales by category** breakdown
- **Revenue by product** table with trend

---

## 12.5 Customer Analytics

- **New vs returning** customers
- **Customer lifetime value** (total spend)
- **Repeat rate** (% customers who ordered 2+ times)
- **Geographic distribution** (orders by area/zone)
- **Promo usage** (which promos used how much)

---

## 12.6 Rider Analytics (Partner Portal)

- **Orders per rider** per day/week
- **Avg delivery time per rider**
- **Distance covered per rider**
- **Rider utilisation** (orders per duty hour)
- **On-time delivery rate**

---

## 12.7 Super Admin Analytics (Cross-Tenant)

| Report | Description |
|--------|-------------|
| Platform Revenue | Total commission earned across all tenants |
| Revenue by Tenant | Commission breakdown per vendor |
| Order Volume | Orders across all tenants with growth trend |
| New Tenants | Tenant onboarding trend |
| Active Customers | MAU per tenant |
| Geographic Map | Order density by city/area |
| Rider Fleet | Total riders, utilisation, top performers |

---

## 12.8 Analytics Query Patterns

### SQLC query examples (representative)

```sql
-- Daily revenue for a restaurant
-- name: GetDailySalesForRestaurant :many
SELECT
    order_date,
    COUNT(*) AS order_count,
    SUM(subtotal - item_discount - promo_discount) AS net_sales,
    SUM(commission) AS commission_amount,
    SUM(vat) AS vat_amount
FROM order_analytics
WHERE tenant_id = $1
  AND $2 = ANY(restaurant_ids)
  AND order_date BETWEEN $3 AND $4
  AND final_status = 'delivered'
GROUP BY order_date
ORDER BY order_date;

-- Peak hours
-- name: GetPeakHours :many
SELECT
    order_hour,
    COUNT(*) AS order_count,
    AVG(total) AS avg_order_value
FROM order_analytics
WHERE tenant_id = $1
  AND $2 = ANY(restaurant_ids)
  AND order_date BETWEEN $3 AND $4
  AND final_status = 'delivered'
GROUP BY order_hour
ORDER BY order_hour;

-- Top selling products
-- name: GetTopSellingProducts :many
SELECT
    oi.product_id,
    oi.product_snapshot->>'name' AS product_name,
    SUM(oi.quantity) AS total_quantity,
    SUM(oi.total) AS total_revenue
FROM order_items oi
JOIN orders o ON o.id = oi.order_id
WHERE oi.tenant_id = $1
  AND oi.restaurant_id = $2
  AND o.status = 'delivered'
  AND o.created_at BETWEEN $3 AND $4
GROUP BY oi.product_id, oi.product_snapshot->>'name'
ORDER BY total_quantity DESC
LIMIT 10;
```

---

## 12.9 Data Retention

| Data | Retention Period |
|------|-----------------|
| `order_analytics` | 3 years |
| `order_timeline` | 1 year |
| `rider_travel_logs` | 6 months |
| `search_logs` | 90 days |
| `audit_logs` | 2 years |
| `notifications` | 90 days |
| Active transactional data | Indefinite (with soft deletes) |
