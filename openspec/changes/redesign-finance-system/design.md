## Context

This document captures implementation decisions for the finance system redesign. These decisions assume `redesign-order-system` is fully deployed — particularly that `order_pickups.commission_amount` is non-zero.

## Goals / Non-Goals

**Goals:**
- Invoice formula correct and complete (all 6 fields populated from real data)
- Double-entry ledger entries for every financial event
- Rider payouts automated end-to-end
- Platform revenue visible to admin
- Daily reconciliation catches mismatches before they compound

**Non-Goals:**
- Multi-currency support (BDT only for now)
- External accounting system integration (QuickBooks, Tally — Phase 3)
- Tax remittance automation (handled manually for now; ledger captures amounts)
- Real-time revenue dashboard (polling-based; SSE in Phase 3)

## Decisions

### Decision: Invoice re-computation vs snapshot
Invoice fields are computed at generation time and stored as snapshots. The stored values do NOT auto-update if underlying data changes. This matches the finance industry standard — invoices are a point-in-time document. Only `adjustment_amount` can change (via explicit adjustment records) while invoice is in `draft`. Once `finalized`, the invoice is immutable.

### Decision: Promo funding source on the `promos` table
Adding `funded_by ENUM('platform', 'vendor') DEFAULT 'vendor'` to the `promos` table is the right location. This is set when the promo is created. The invoice service filters `promo_usages JOIN promos ON promo_usages.promo_id = promos.id WHERE promos.funded_by = 'vendor'` when computing `vendor_promo_discounts`.

### Decision: Ledger uses the existing tables (not rebuild)
`ledger_accounts` and `ledger_entries` in migration 000020 are well-designed. We activate them by:
1. Inserting account seed records on tenant/restaurant creation
2. Writing a `LedgerService` that wraps all entry creation
3. Wiring `LedgerService` calls from the outbox processor

No schema changes to the ledger tables themselves — only additions (account types via the existing `account_type` column).

### Decision: Double-entry with named accounts
Each financial event maps to exactly two ledger entries (one debit, one credit). Account naming convention:
- `PLATFORM_COMMISSION_REVENUE` (singleton per platform)
- `PLATFORM_DELIVERY_REVENUE` (singleton per platform)
- `PLATFORM_SUBSCRIPTION_REVENUE` (singleton per platform)
- `VENDOR_LIABILITY_{restaurant_id}` (one per restaurant)
- `WALLET_LIABILITY` (per tenant)
- `RIDER_LIABILITY` (per tenant)

Account balances can be queried with:
```sql
SELECT account_id, SUM(credit_amount) - SUM(debit_amount) AS balance
FROM ledger_entries WHERE tenant_id = $1 GROUP BY account_id
```

### Decision: Rider payout via bKash B2C (not manual bank transfer)
Bangladesh context: bKash B2C (business-to-customer) is the correct channel for rider payouts. The integration uses the bKash B2B API (disbursement API). The rider must have a registered bKash number in their profile. Fallback: if disbursement fails twice, flag for manual bank transfer.

### Decision: Subscription billing day stored on tenant
`tenants.billing_day INT` (1-28) stores which day of the month billing runs. Default: 1 (first of month). The billing worker runs daily and checks if `TODAY = billing_day` and `last_billed_month != current_month` to avoid double-billing.

### Decision: COD reconciliation window = 24h
Riders are expected to remit cash within 24 hours of collection. After 24h, the record goes `overdue` and an admin alert is created. This is a business rule that may be made configurable via `platform_configs['cod_remittance_window_hours']`.

### Decision: Gateway fee tracking (bKash: 1.5%, AamarPay: 1.2%)
We track the gateway fee that is deducted from the merchant settlement. This is NOT charged to the customer — it reduces platform net revenue. The webhook payloads for both gateways include the fee. If not present in webhook, derive it: `gateway_fee = amount × gateway_fee_rate` using rate stored in `platform_configs`.

## Migration Plan

1. Migration `000023_finance_fixes.up.sql`:
   - `ALTER TABLE promos ADD COLUMN funded_by TEXT NOT NULL DEFAULT 'vendor' CHECK (funded_by IN ('platform', 'vendor'))`
   - `ALTER TABLE restaurants ADD COLUMN delivery_managed_by TEXT NOT NULL DEFAULT 'platform' CHECK (delivery_managed_by IN ('platform', 'vendor'))`
   - `ALTER TABLE payment_transactions ADD COLUMN gateway_fee NUMERIC(10,2) NOT NULL DEFAULT 0`
   - `CREATE TABLE invoice_adjustments (...)`
   - `CREATE TABLE subscription_invoices (...)`
   - `CREATE TABLE cash_collection_records (...)`
   - `CREATE TABLE reconciliation_alerts (...)`
   - Add `tenant_subscriptions` table if not exists

2. Ledger account seeding:
   - Run `SeedLedgerAccounts()` for all existing tenants on deploy
   - Hook into tenant creation and restaurant creation going forward

3. Backfill `promo_usages` linkage (verify existing records have correct promo IDs)

4. Deploy finance service fixes + ledger service activation

5. Smoke test:
   - Create order → deliver → verify ledger entry for commission
   - Finalize invoice → verify all 6 fields non-zero
   - Run payout job → verify `rider_payouts` created
   - Run reconciliation job → verify processing payments checked

## Open Questions

- Should `invoice.delivery_charge_total` be a new column or derived at query time? (Recommendation: store it for historical accuracy — zone prices may change)
- What is the `rider_base_delivery_earning` default? (Bangladesh market: likely 30-50 BDT per delivery)
- Should subscription invoices be separate from restaurant invoices, or merged? (Recommendation: separate — different purpose, different recipient)
- Who receives the subscription invoice — tenant owner or billing contact? (Needs product decision)
