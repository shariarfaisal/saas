## Context

The order system has been scaffolded through multiple phases but has accumulating design debt that makes it non-deployable. This document captures the precise technical decisions needed to fix the 10 blocking issues before the finance system can produce correct numbers.

## Goals / Non-Goals

**Goals:**
- Fix every issue that corrupts financial data (commission = 0, delivery charge hardcoded, promo not distributed)
- Complete the order lifecycle (MarkDelivered missing)
- Make order creation safe under retries and concurrent load (idempotency, FOR UPDATE)
- Make real-time tracking functional (SSE publishing)
- Fix performance timebomb (order number COUNT(*))

**Non-Goals:**
- Distance-based delivery pricing (Phase 2 — Barikoi integration)
- Scheduled/future orders
- Order modification after creation
- Proof-of-delivery (photo upload)
- Multi-language order status messages

## Decisions

### Decision: Commission at order creation, not invoice time
Commission rate is captured on each `order_pickup` at order creation. This is correct — the rate is frozen at the moment of the order, not recalculated when the invoice is generated later. The invoice service simply sums `commission_amount` from `order_pickups`. The fix is to populate these fields that are currently hardcoded to 0.

```go
// CURRENT (broken) — order/service.go:488
commRatePg.Scan("0")   // always 0
commAmtPg.Scan("0")    // always 0

// FIX
restaurant, _ := s.restaurantRepo.GetByID(ctx, restaurantID)
commRate := restaurant.CommissionRate  // restaurant-specific
if !commRate.Valid {
    tenant, _ := s.tenantRepo.GetByID(ctx, tenantID)
    commRate = tenant.CommissionRate   // fall back to tenant default
}
commAmount := restaurantItems.Total.Mul(commRate).Div(decimal.NewFromInt(100))
```

### Decision: Promo distribution by item weight
When a promo reduces the order total, we distribute the discount proportionally to eligible items. This ensures each `order_items.promo_discount` is correct for per-restaurant settlement — a restaurant only absorbs the promo discount on its own items.

```
promoDiscountPerItem[i] = totalPromoDiscount × (item[i].eligibleSubtotal / totalEligibleSubtotal)
```
Rounding difference (penny) assigned to the last item.

### Decision: Order number via PostgreSQL sequence
Replace `COUNT(*) + 1` with a per-tenant sequence. Sequence name: `order_seq_{tenant_slug}`, created lazily on first order per tenant, or via seed migration for existing tenants.

```sql
CREATE SEQUENCE IF NOT EXISTS order_seq_kacchi_bhai START 1;
SELECT LPAD(NEXTVAL('order_seq_kacchi_bhai')::TEXT, 6, '0');
-- → 'KBC-000001', 'KBC-000002', ... atomically, no race conditions
```

### Decision: Idempotency check using existing `idempotency_keys` table
The table (`000004`) already exists. Order creation handler reads `Idempotency-Key` header, checks the table, returns cached response if found. Key stored as `(tenant_id, user_id, 'POST /orders', key_hash)` with 24h TTL.

### Decision: SSE via Redis pub/sub (existing infrastructure)
The outbox processor already writes events after every order status change. Add a Redis PUBLISH step in the outbox processor for channel `order:{orderID}`. The SSE handler subscribes to that channel via Redis SUBSCRIBE and pushes events to the client. No new infrastructure required.

### Decision: `FOR UPDATE SKIP LOCKED` for stock reservation
Replace the current bare `UPDATE ... WHERE stock_qty - reserved_qty >= qty` with a `SELECT ... FOR UPDATE SKIP LOCKED` to acquire a row lock before the update. This prevents two concurrent requests from both seeing sufficient stock and both succeeding.

### Decision: ConsumeReservedStock on PICKED (not CONFIRMED)
Stock is physically gone when the rider picks it up, not when the restaurant confirms. Consuming on `PICKED` is most accurate. On cancellation, `ReleaseStock` runs to free the reservation. On `DELIVERED`, no additional stock action needed (already consumed at PICKED).

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| Commission change breaks existing invoice drafts | Migration step: backfill `commission_rate`/`commission_amount` on existing `order_pickups` from restaurant/tenant rate |
| Sequence gaps on rollback | Acceptable — gaps in order numbers are normal (bank-grade systems use sequences with gaps) |
| Idempotency key storage growth | TTL-based cleanup job; 24h retention is sufficient |
| SSE fan-out under high load | Redis pub/sub scales well; limit SSE connections per user to 1 per order |

## State Machine (Authoritative)

```
PENDING   → CREATED       (payment confirmed OR COD placed)
CREATED   → CONFIRMED     (restaurant confirms OR auto-confirm timer)
CREATED   → REJECTED      (restaurant rejects with reason)
CREATED   → CANCELLED     (customer cancels OR payment timeout)
CONFIRMED → PREPARING     (restaurant starts cooking)
PREPARING → READY         (restaurant marks ready for pickup)
READY     → PICKED        (rider picks up ALL restaurant portions)
PICKED    → DELIVERED     (rider marks delivered) ← NEW endpoint
PENDING   → CANCELLED     (payment timeout 30min)
Any non-terminal → CANCELLED (admin force-cancel with reason)

Terminal states: DELIVERED, CANCELLED, REJECTED
```

## Migration Plan

1. Migration `000021_order_system_fixes.up.sql`:
   - Add `packaging_fee` column to `orders`
   - Add `UNIQUE(modifier_group_id, name)` to `product_modifier_options`
   - Add `created_at` to `inventory_items`
   - Add pending-payment index
2. Migration `000022_order_sequences.up.sql`:
   - Create per-tenant sequences for active tenants
3. Backfill script: populate `commission_rate`/`commission_amount` on existing `order_pickups` from `restaurants.commission_rate ?? tenants.commission_rate`
4. Deploy order module fixes
5. Verify: create test order → check `order_pickups.commission_amount > 0`

## Open Questions

- Should `packaging_fee` be configurable per restaurant in the `restaurants` table, or come from `platform_configs`?
- Should rider earnings be created on `DELIVERED` event or on the payout cycle? (Recommendation: created on DELIVERED, paid on payout cycle)
- What is the correct auto-confirm default? Currently 5 min in code vs 3 min in Phase 16 proposal.
