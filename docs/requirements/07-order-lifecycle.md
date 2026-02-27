# 07 — Order Lifecycle

## 7.1 Order Status State Machine

```
                        ┌─────────────────────────────────────────────────────┐
                        │                                                     │
                [Customer places order]                                       │
                        │                                                     │
                        ▼                                                     │
                   ┌─────────┐                                                │
                   │ PENDING │  ◄── Payment not yet confirmed (online pay)    │
                   └────┬────┘      OR order just placed (COD)                │
                        │                                                     │
         [Payment confirmed OR COD order]                                     │
                        │                                                     │
                        ▼                                                     │
                   ┌─────────┐                                                │
                   │ CREATED │  ◄── Visible to restaurant, awaiting confirm   │
                   └────┬────┘                                                │
                        │                                                     │
     [Restaurant confirms] OR [Auto-confirm timer expires]                    │
                        │                                                     │
                        ▼                                                     │
                 ┌────────────┐                                               │
                 │ CONFIRMED  │  ◄── Restaurant acknowledged, rider notified  │
                 └─────┬──────┘                                               │
                       │                                                      │
             [Restaurant starts cooking]                                      │
                       │                                                      │
                       ▼                                                      │
                 ┌────────────┐                                               │
                 │ PREPARING  │  ◄── Food being prepared                      │
                 └─────┬──────┘                                               │
                       │                                                      │
              [Restaurant marks food ready]                                   │
                       │                                                      │
                       ▼                                                      │
                 ┌────────────┐                                               │
                 │   READY    │  ◄── Food ready, rider heading to pickup      │
                 └─────┬──────┘                                               │
                       │                                                      │
     [Rider picks up from all restaurants in order]                           │
                       │                                                      │
                       ▼                                                      │
                 ┌────────────┐                                               │
                 │   PICKED   │  ◄── Rider has food, heading to customer      │
                 └─────┬──────┘                                               │
                       │                                                      │
             [Rider marks as delivered]                                       │
                       │                                                      │
                       ▼                                                      │
                 ┌───────────┐                                                │
                 │ DELIVERED │  ◄── Order complete                            │
                 └───────────┘                                                │
                                                                              │
─── Terminal failure states ───────────────────────────────────────────────  │
                                                                              │
PENDING / CREATED  ──[Customer cancels]──►  CANCELLED                        │
CREATED            ──[Restaurant rejects]──► REJECTED                        │
Any active         ──[Admin force cancel]──► CANCELLED                       │
```

---

## 7.2 Restaurant Sub-Order Status (per pickup)

Each restaurant in a multi-restaurant order has its own pickup status:

```
new → confirmed → preparing → ready → picked
         │
       rejected  (only when restaurant-level rejection)
```

The parent order status advances based on all pickups:
- Parent becomes `CONFIRMED` when first pickup is confirmed
- Parent becomes `READY` when ALL pickups are ready
- Parent becomes `PICKED` when ALL pickups are picked

---

## 7.3 Order Creation Flow (Detailed)

```
1. Customer builds cart, proceeds to checkout
2. Customer selects delivery address, payment method, optional promo code
3. Frontend calls POST /api/v1/orders

Server-side atomic order creation:
  a. Validate customer is active
  b. Load all products (with restaurant, category, pricing)
  c. Check all restaurants are currently open (operating hours)
  d. Check product availability for each item
  e. Check stock if inventory-tracked products
  f. Calculate item totals (base + variants + addons) per item
  g. Apply product-level discounts
  h. Validate and apply promo code (if provided):
      - Check code exists and is active
      - Check validity window (start/end date)
      - Check min order amount met
      - Check max usage not exceeded
      - Check user hasn't exceeded per-user limit
      - Calculate promo discount
  i. Calculate delivery charge (based on customer_area from delivery zone)
  j. Calculate VAT
  k. Compute final total
  l. If online payment: create payment intent, return payment URL → status = PENDING
  m. If COD: create order directly → status = CREATED
  n. Reserve stock (decrement inventory)
  o. Create order_items records
  p. Create order_pickups records (one per restaurant)
  q. Enqueue rider_auto_assign job
  r. Send new order notification to restaurants via SSE + push
  s. Return order ID and (for online pay) payment redirect URL
```

---

## 7.4 Auto-Confirm Timer

Configuration: `tenant.settings.order.auto_confirm_after_minutes` (default: 3 minutes)

When an order is in `CREATED` status for longer than the configured timeout without restaurant action:
- System auto-confirms the order (all pickups set to `confirmed`)
- Order status → `CONFIRMED`
- Timeline entry added: `{actor: "system", message: "Order auto-confirmed"}`
- Rider assignment triggered if not already assigned

---

## 7.5 Rider Assignment

### Auto-Assignment Algorithm
1. Find all riders for the tenant's hub (matching the order's hub)
2. Filter: `is_available = true` AND `is_on_duty = true`
3. Calculate distance from each rider's last known location to the first pickup restaurant
4. Sort by distance (nearest first)
5. Send assignment request to top 3 riders simultaneously (push notification + SSE)
6. First rider to accept gets the order
7. If no rider accepts within 60 seconds, retry with next batch
8. If no rider available after 3 retries, notify tenant_admin

### Manual Assignment (Admin/Manager)
- Admin can manually assign any available rider to any unassigned order
- Timeline entry created: `{actor: user_id, message: "Rider X assigned by Y"}`
- Assigned rider notified via push

---

## 7.6 Multi-Restaurant Order Pickup

When an order contains items from multiple restaurants:
- Each restaurant sees only their own items in partner portal
- Rider collects from each restaurant sequentially
- Rider taps "Picked" for each restaurant individually
- Parent order shows `PICKED` only after all restaurants marked as picked
- Order number shown is: `[ORDER_PREFIX]-[RESTAURANT_PREFIX]-[NUMBER]` per pickup

---

## 7.7 Order Cancellation & Rejection

### Customer Cancellation
- Allowed only in `PENDING` or `CREATED` status
- Not allowed once restaurant has confirmed
- If COD: no financial action needed
- If online payment (paid): auto-trigger refund to original payment method
- Stock released back to inventory

### Restaurant Rejection
- Restaurant can reject at `NEW` or `CONFIRMED` status
- Must provide rejection reason
- If all restaurants reject: order status → `REJECTED`
- If partial rejection in multi-restaurant order: remaining restaurants continue; customer notified of partial rejection
- Refund triggered for rejected items (if online payment)

### Admin Force Cancel
- Super admin or tenant admin can cancel any order in any status except `DELIVERED`
- Full refund triggered
- Timeline entry with reason recorded

---

## 7.8 Payment Flow

### Cash on Delivery (COD)
```
Order created → status: CREATED → ... → DELIVERED
(No payment processing; revenue recorded at delivery)
```

### bKash (Bangladesh Mobile Banking)
```
Order created → status: PENDING
→ User redirected to bKash payment URL
→ bKash callback (success): order status: CREATED, payment_status: paid
→ bKash callback (fail/cancel): stock released, order soft-deleted
→ ... order continues normally
```

### AamarPay / SSLCommerz (Card payment)
```
Order created → status: PENDING
→ User redirected to payment gateway
→ Callback (success): order status: CREATED, payment_status: paid
→ Callback (fail): stock released, order soft-deleted
→ ... continues normally
```

### Wallet / Points
```
Check user.balance >= order.total
→ Deduct from user.balance
→ Order created → status: CREATED
```

---

## 7.9 Order Charge Calculation (Exact Logic)

```
For each order item:
  item_base = product.base_price × quantity
  item_variants = sum(selected_variant_item.price) × quantity  [if variant-type product]
  item_addons = sum(selected_addon.price) × quantity
  item_subtotal = item_base + item_variants + item_addons

  item_discount:
    if product has active discount:
      type=fixed → min(discount.amount × qty, item_subtotal)
      type=percent → round((item_subtotal × discount.amount) / 100)
    else → 0

  item_vat = round(((item_subtotal - item_discount) × vat_rate) / 100)
  item_total = item_subtotal - item_discount + item_vat

order_subtotal = sum(item_subtotal)
order_item_discount = sum(item_discount)
order_vat = sum(item_vat)

promo_discount (if promo applied on items/categories):
  eligible_total = sum of (item_subtotal - item_discount) for applicable items
  if promo_type = fixed:
    promo_discount = min(promo.amount, eligible_total, promo.max_discount_amount)
  if promo_type = percent:
    promo_discount = round(eligible_total × promo.amount / 100)
    promo_discount = min(promo_discount, promo.max_discount_amount)

delivery_charge:
  if promo applied on delivery_charge:
    if promo_type = fixed: delivery_charge = max(0, base_charge - promo.amount)
    if promo_type = percent: delivery_charge = max(0, base_charge - round(base_charge × promo.amount / 100))
  else:
    delivery_charge = zone_charge[customer_area]

order_total = order_subtotal - order_item_discount - promo_discount + order_vat + delivery_charge
```

---

## 7.10 Order Number Generation

Format: `[RESTAURANT_PREFIX]-[ZERO_PADDED_SEQUENCE]`
Example: `KBC-001234` (Kacchi Bhai — order #1234)

- Sequence is per-restaurant, incremented atomically in PostgreSQL
- For multi-restaurant orders: each pickup gets its own sub-number
  - Parent order: uses first restaurant's prefix
  - Each pickup: `[PREFIX]-[SEQ]`

---

## 7.11 Estimated Delivery Time

Calculated at order creation:
```
prep_time = max(prep_time of all restaurants in order, considering category prep_time)
travel_time = calculated from rider hub to restaurant to customer (via Barikoi distance API, cached)
buffer_time = 5 minutes

estimated_delivery_time = prep_time + travel_time + buffer_time
```

Updated at each status transition using remaining time logic.

---

## 7.12 Order Issue / Dispute Flow

```
Customer or Admin reports issue
  → order_issue created (status: open)
  → Notified to restaurant (via push + SSE) and platform_admin

Admin reviews issue:
  → Determines accountable party (restaurant / rider / platform)
  → Sets refund amount and refund_items
  → Approves or rejects

If approved:
  → Refund processed (same as cancellation payment reversal)
  → Restaurant penalty deducted from next invoice (if restaurant accountable)
  → Rider penalty recorded (if rider accountable)
  → Customer wallet credited OR payment gateway refund

Issue closed.
```

---

## 7.13 Consistency, Idempotency & Failure-Safe Rules

### Idempotency (mandatory)
- `POST /orders` and payment-initiation endpoints must require `Idempotency-Key` header.
- Server stores `(tenant_id, user_id, endpoint, idempotency_key, request_hash, response_snapshot, expires_at)`.
- Repeated requests with same key and same payload return original response.
- Reused key with different payload returns `409 IDEMPOTENCY_KEY_REUSED_WITH_DIFFERENT_PAYLOAD`.

### Concurrency control
- Stock reservation must be wrapped in a DB transaction (`SELECT ... FOR UPDATE`) to prevent oversell.
- Promo usage counters must be incremented atomically to enforce `max_usage`.
- Order state transitions must use optimistic checks:
  - `UPDATE ... WHERE id = $1 AND status = $expected_old_status`
  - if row count = 0, transition is rejected as stale.

### Outbox pattern for side effects
Order-write transaction must include outbox events for:
- `order.created`
- `order.status_changed`
- `payment.succeeded`
- `payment.failed`

Background workers publish these events after commit, preventing lost notifications during crash/restart windows.

### Payment callback safety
- Gateway callbacks are retried by providers; callback handlers must be idempotent by `gateway_txn_id`.
- Duplicate successful callback must not re-capture payment or append duplicate timeline records.
- Failed callback processing must return non-2xx to trigger provider retry until reconciliation succeeds.

---

## 7.14 Operational Edge Cases & Required Behaviors

1. **Payment success, callback delayed**  
   - `PENDING` order remains recoverable via reconciliation job polling gateway API.

2. **Payment failed after stock reserved**  
   - Stock release job must run immediately and be retried until success.

3. **Rider assigned but unreachable**  
   - Auto-unassign timeout and re-dispatch to next candidate rider batch.

4. **Partial restaurant rejection in multi-restaurant order**  
   - Customer receives partial delivery summary and partial refund breakdown.

5. **ETA breach**  
   - System records SLA breach metrics and can auto-issue compensation per tenant policy.

6. **System maintenance mode during active orders**  
   - New orders blocked; active orders continue with operational overrides available to support/admin.
