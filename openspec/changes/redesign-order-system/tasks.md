# Tasks: redesign-order-system

## Section 0 — Schema & Migrations

- [ ] 0.1 Write migration `000021_order_system_fixes.up.sql`:
  - `ALTER TABLE orders ADD COLUMN packaging_fee NUMERIC(10,2) NOT NULL DEFAULT 0`
  - `ALTER TABLE inventory_items ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
  - `CREATE UNIQUE INDEX uq_modifier_option_name ON product_modifier_options(modifier_group_id, name)`
  - `CREATE INDEX idx_orders_pending_payment ON orders(tenant_id, created_at) WHERE status = 'pending' AND payment_status = 'unpaid'`
- [ ] 0.2 Write migration `000022_order_sequences.up.sql`: create `order_seq_<tenant_slug>` sequence for each existing tenant; add helper function `next_order_number(tenant_id UUID) → TEXT`
- [ ] 0.3 Write migration `000021` down file (drop column, drop index, drop constraint)
- [ ] 0.4 Regenerate SQLC after migration

## Section 1 — Order Number via Sequence

- [ ] 1.1 Add SQLC query `GetNextOrderNumber(tenantID)` → calls `next_order_number(tenant_id)`
- [ ] 1.2 Replace `COUNT(*) + 1` order number logic in `order/service.go:CreateOrder` with sequence call
- [ ] 1.3 Add test: concurrent 100 orders → all order numbers unique, format correct (`KBC-000001`)

## Section 2 — State Machine Enforcement

- [ ] 2.1 Create `internal/modules/order/statemachine.go` with `validTransitions map[OrderStatus][]OrderStatus`
- [ ] 2.2 Add `ValidateTransition(from, to OrderStatus) error` function
- [ ] 2.3 Wire `ValidateTransition` in `UpdatePickupStatus`, `CancelOrder`, and all status-change service methods
- [ ] 2.4 Return `ErrInvalidStatusTransition` (422) when transition not in map
- [ ] 2.5 Unit test all valid transitions pass; all invalid transitions rejected

## Section 3 — Commission Calculation

- [ ] 3.1 Add `GetRestaurantCommissionRate(tenantID, restaurantID)` service helper (looks up `restaurants.commission_rate ?? tenants.commission_rate`)
- [ ] 3.2 In `CreateOrder`, per restaurant group: call helper, compute `commission_amount = items_total × rate / 100`
- [ ] 3.3 Populate `order_pickups.commission_rate` and `order_pickups.commission_amount` at order creation
- [ ] 3.4 Unit test: restaurant with 12% rate → pickup has `commission_rate = 12`, `commission_amount = correct amount`
- [ ] 3.5 Unit test: restaurant with NULL rate falls back to tenant rate
- [ ] 3.6 Write backfill script `scripts/backfill_commission_amounts.go` for existing records (safe to run on live data)

## Section 4 — Zone-Based Delivery Charge

- [ ] 4.1 Add `GetDeliveryCharge(tenantID, deliveryArea string) (decimal.Decimal, error)` method to delivery zone service
- [ ] 4.2 Fall back to `platform_configs['default_delivery_charge']` when no zone match
- [ ] 4.3 Replace hardcoded `60` in `CreateOrder` with `GetDeliveryCharge` call
- [ ] 4.4 Unit test: area with zone mapping → correct charge; unmapped area → platform default

## Section 5 — Promo Discount Distribution

- [ ] 5.1 Add helper `DistributePromoDiscount(items []OrderItem, totalDiscount decimal.Decimal, eligibleItemIDs []UUID) []decimal.Decimal`
- [ ] 5.2 Distribution algorithm: proportional by `item.eligible_subtotal / total_eligible_subtotal`; remainder to last item
- [ ] 5.3 Wire helper in `CreateOrder` after promo validation
- [ ] 5.4 Populate `order_items.promo_discount` for each item
- [ ] 5.5 Unit test: 2 items same price → 50/50 split; 2 items 2:1 price ratio → 2/3 + 1/3
- [ ] 5.6 Unit test: promo `applies_to = specific_restaurant` → only items from that restaurant receive discount

## Section 6 — Wallet Payment Fix

- [ ] 6.1 In `CreateOrder` when `payment_method = wallet`: deduct balance first, return `ErrInsufficientWalletBalance` if low
- [ ] 6.2 Set `order.payment_status = 'paid'` and `order.status = 'created'` (not pending) on wallet orders
- [ ] 6.3 Create `wallet_transactions` debit record with `source = order_payment`, `reference_id = order_id`
- [ ] 6.4 Unit test: wallet order → status = created, payment_status = paid, wallet balance decremented
- [ ] 6.5 Unit test: insufficient balance → 422 INSUFFICIENT_WALLET_BALANCE, order NOT created

## Section 7 — MarkDelivered Endpoint

- [ ] 7.1 Add `MarkDelivered(ctx, orderID, riderID UUID) error` service method
- [ ] 7.2 Validate: order must be in `PICKED` status; rider must be assigned rider
- [ ] 7.3 Transition order to `DELIVERED`, set `delivered_at = NOW()`
- [ ] 7.4 Create `order_timeline_events` record: `event_type = status_changed`, `actor_type = rider`, `actor_id = riderID`
- [ ] 7.5 Insert outbox event `order.delivered` with full order context
- [ ] 7.6 Register `PATCH /rider/orders/:id/deliver` in router with rider auth middleware
- [ ] 7.7 Integration test: rider marks delivered → order DELIVERED, timeline event created, outbox event inserted

## Section 8 — Idempotency on Order Creation

- [ ] 8.1 Add `Idempotency-Key` header validation middleware for `POST /orders`; return 400 if missing
- [ ] 8.2 Add `CheckIdempotencyKey(tenantID, userID, key string) (cachedResp, bool)` method using existing `idempotency_keys` table
- [ ] 8.3 Add `StoreIdempotencyResult(tenantID, userID, key string, response []byte, expiresIn time.Duration)` method
- [ ] 8.4 Store order creation response with 24h TTL on success
- [ ] 8.5 Integration test: duplicate `POST /orders` with same key → second returns same response, one order in DB

## Section 9 — Payment Timeout Job

- [ ] 9.1 Create `internal/worker/payment_timeout_worker.go` implementing asynq periodic task `CancelTimedOutPayments`
- [ ] 9.2 Query: `SELECT id FROM orders WHERE status = 'pending' AND payment_status = 'unpaid' AND created_at < NOW() - INTERVAL '30 min'` using new index
- [ ] 9.3 For each: cancel order, release stock, create timeline event `actor_type = system`, insert outbox `order.payment_timeout`
- [ ] 9.4 Register periodic job in `cmd/worker/main.go` at 1-minute interval
- [ ] 9.5 Add `payment_timeout_minutes` to `platform_configs` with default 30
- [ ] 9.6 Unit test: order older than 30 min with unpaid status → cancelled and stock released

## Section 10 — Stock Consume on PICKED

- [ ] 10.1 Add `ConsumeStockForPickup(ctx, pickupID, items []OrderItem) error` method in inventory service
- [ ] 10.2 Wire call from order service after successful `PICKED` transition for each inventory-tracked item
- [ ] 10.3 Create `inventory_adjustments` record per item: `adjustment_type = order_consume`, `reference_id = order_id`
- [ ] 10.4 On `CANCELLED`/`REJECTED`: call `ReleaseReservedStock` for all items that were reserved
- [ ] 10.5 Integration test: order → PICKED → `stock_qty` decremented, `reserved_qty` decremented

## Section 11 — Atomic Stock Reservation (FOR UPDATE)

- [ ] 11.1 Update `inventory/queries.sql` `ReserveStock`: add `SELECT id FROM inventory_items WHERE id = $1 FOR UPDATE SKIP LOCKED`
- [ ] 11.2 Handle case where `SKIP LOCKED` returns no row → return `ErrStockTemporarilyUnavailable` (retry)
- [ ] 11.3 Regenerate SQLC
- [ ] 11.4 Concurrency test: 50 goroutines reserve same product → no negative stock_qty; exactly `initial_qty` units reserved total

## Section 12 — SSE via Redis Pub/Sub

- [ ] 12.1 Add `PublishOrderEvent(ctx, channel, event OrderEvent) error` to a new `events` package using `go-redis` PUBLISH
- [ ] 12.2 In outbox processor: after marking event processed, call `PublishOrderEvent` for `order.status_changed` and `order.created`
- [ ] 12.3 Rewrite `GET /api/v1/orders/:id/stream` handler to subscribe to Redis `order:{orderID}` and forward to SSE writer
- [ ] 12.4 Add partner portal stream `GET /api/v1/partner/restaurants/:id/orders/stream` subscribing to `tenant:{id}:restaurant:{id}:orders`
- [ ] 12.5 Add 30-second heartbeat `data: {"type":"ping"}\n\n` to both SSE handlers
- [ ] 12.6 Ensure goroutine/connection cleanup on client disconnect via context cancellation
- [ ] 12.7 Integration test: publish event to Redis → SSE client receives within 500ms

## Section 13 — Catalog Integrity

- [ ] 13.1 In product service `CreateProduct`/`UpdateProduct`: add `ValidateCategoryOwnership(restaurantID, categoryID)` check
- [ ] 13.2 In modifier option service `CreateOption`: add `ValidateOptionName(modifierGroupID, name)` uniqueness check (before DB constraint)
- [ ] 13.3 In outbox processor, handle `order.delivered`: increment `products.order_count` for all items
- [ ] 13.4 In outbox processor, handle `review.created`: recalculate `products.rating_avg` and increment `products.rating_count`

## Section 14 — Validation & Smoke Testing

- [ ] 14.1 Run full migration on staging DB; verify schema
- [ ] 14.2 Run commission backfill script; verify `commission_amount > 0` on sample `order_pickups`
- [ ] 14.3 Create end-to-end test: full order lifecycle from PENDING → DELIVERED; verify commission set, promo distributed, stock consumed
- [ ] 14.4 Create end-to-end test: wallet order → verify payment_status = paid, balance decremented
- [ ] 14.5 Verify SSE: connect to stream, change status, confirm event received
- [ ] 14.6 Run `openspec validate redesign-order-system --strict`
