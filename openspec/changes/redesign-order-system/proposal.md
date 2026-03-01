# Change: Redesign Order System — Lifecycle, Pricing, Reliability & Real-time

## Why

Deep analysis of the backend (`backend/internal/modules/order/`, all order-related migrations, and query files) reveals **blocking production bugs** that make the current order system non-functional in a live environment:

1. **No `MarkDelivered` endpoint exists** — orders can reach `PICKED` but can never complete the lifecycle to `DELIVERED`. Revenue recognition, rider earnings, cashback, and invoice generation all depend on this event.
2. **Commission is hardcoded to `0` on every `order_pickup`** — `service.go:488-491` hardcodes `commission_rate = "0"` and `commission_amount = "0"` at order creation. The finance module reads these fields to generate invoices, so every invoice calculates zero commission.
3. **Delivery charge is hardcoded to 60 BDT** — `service.go` sets `deliveryCharge := decimal.NewFromInt(60)` for all areas, ignoring the `delivery_zones` table entirely.
4. **Promo discount is never distributed to order items** — `order_items.promo_discount` is always `0.00`; the discount exists at order level but is never allocated to line items, breaking per-restaurant settlement.
5. **SSE tracking is a no-op stub** — the handler sends one snapshot then silently waits for disconnect; no status-change events are ever pushed.
6. **Order number uses `COUNT(*)`** — O(n) table scan on every order creation; catastrophic at scale.
7. **No idempotency on order creation** — duplicate requests (e.g., mobile retry on timeout) create duplicate orders and double-charge customers.
8. **Stock is reserved but never consumed** — `ConsumeReservedStock` exists in the inventory SQLC queries but is never called when an order is confirmed/delivered; reserved stock permanently reduces available stock.
9. **Wallet payments are never marked PAID** — COD and wallet both set `payment_status = unpaid` at creation; wallet users are charged but the order shows as unpaid.
10. **No state machine enforcement** — any status → any status transition is accepted; e.g., a restaurant can mark an order "Ready" without ever confirming it.

## What Changes

### Schema additions (migrations)
- `CREATE SEQUENCE` for order number generation per tenant
- `ADD COLUMN packaging_fee NUMERIC(10,2) DEFAULT 0` on `orders` for per-restaurant packaging charges
- `UNIQUE(modifier_group_id, name)` constraint on `product_modifier_options` to prevent duplicate options
- `ADD COLUMN created_at` to `inventory_items` (currently missing)
- Index: `CREATE INDEX idx_orders_pending_payment ON orders(tenant_id, created_at) WHERE status = 'pending' AND payment_status = 'unpaid'`

### Backend — order module
- **Add `MarkDelivered` handler + service method** (`PATCH /rider/orders/:id/deliver`): transitions `PICKED → DELIVERED`, records `delivered_at`, triggers cashback, rider earnings creation, invoice contribution, outbox `order.delivered` event
- **Fix commission calculation**: in `CreateOrder`, look up `restaurant.commission_rate ?? tenant.commission_rate` per restaurant in the cart and populate `order_pickups.commission_rate` and `order_pickups.commission_amount` at creation time
- **Fix delivery charge**: replace hardcoded `60` with `DeliveryZoneService.GetCharge(tenantID, deliveryArea)` lookup; fall back to `platform_configs['default_delivery_charge']`
- **Fix promo discount distribution**: after calculating total promo discount, allocate it proportionally across eligible `order_items` by `item_subtotal` weight and populate each item's `promo_discount` field
- **Enforce state machine**: every status transition validates `previousStatus ∈ allowedPredecessors[newStatus]`; return `422 INVALID_STATUS_TRANSITION` otherwise
- **Fix wallet payment status**: when `payment_method = wallet`, set `payment_status = 'paid'` immediately after wallet deduction; create `wallet_transactions` debit record
- **Add idempotency**: check `idempotency_keys` table before processing `POST /api/v1/orders`; store response snapshot keyed by `(tenant_id, user_id, idempotency_key)` with 24h TTL
- **Fix order number**: replace `COUNT(*) + 1` with `NEXTVAL('tenant_order_seq_<tenantID>')` or a shared sequence with tenant prefix
- **Add payment timeout job**: scheduled job to auto-cancel `PENDING` orders with `payment_status = unpaid` older than 30 minutes; releases reserved stock
- **Fix SSE streaming**: subscribe to Redis pub/sub channel `orders:{orderID}` in the tracking handler; outbox worker publishes on every status change; customer and partner receive live status events

### Backend — inventory module
- **Wire `ConsumeReservedStock`**: call from order service when pickup transitions to `PICKED` (rider picked up) — permanently deducts stock and reduces `reserved_qty`
- **Fix race condition in `ReserveStock`**: add `FOR UPDATE SKIP LOCKED` to the stock-check query to prevent concurrent overselling under high concurrency
- **Fix denormalized stats**: add triggers (or service-layer increments) to update `products.order_count` on order deliver, `products.rating_avg` and `products.rating_count` on review submission

### Backend — catalog module
- **Add `UNIQUE(modifier_group_id, name)` constraint** via migration to prevent duplicate modifier options
- **Validate cross-restaurant category assignment**: service-layer check that `product.restaurant_id == category.restaurant_id` before allow

## Impact

- Affected specs: `order-lifecycle`, `order-pricing`, `order-reliability`, `order-realtime`, `catalog-integrity`
- Affected code:
  - `backend/internal/modules/order/service.go` — major fixes (commission, delivery, promo, state machine, idempotency, delivered)
  - `backend/internal/modules/order/handler.go` — new MarkDelivered handler + SSE fix
  - `backend/internal/modules/inventory/service.go` — ConsumeReservedStock wiring + FOR UPDATE
  - `backend/internal/db/queries/orders.sql` — order number seq query
  - `backend/internal/db/queries/inventory.sql` — FOR UPDATE fix
  - `backend/internal/db/migrations/` — new migration for sequence, constraints, index
  - `backend/internal/modules/worker/` — payment timeout job, SSE publish in outbox processor
- **BREAKING in behavior**: commission rates will now be non-zero on pickups; finance invoice totals will change
- **No breaking API contract changes**: same endpoints, same request/response shapes (adding one new endpoint)
- Finance proposal (`redesign-finance-system`) depends on this change being applied first
