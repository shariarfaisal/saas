# Change: Redesign Finance & Settlement System — Invoices, Ledger, Payouts & Revenue

## Why

The finance system schema (migrations 000013, 000014, 000020) is largely correct in structure, but the service layer implementation is almost entirely non-functional:

1. **Invoice fields hardcoded to zero** — `finance/service.go` hardcodes `vendor_promo_discounts = decimal.Zero`, `penalty_amount = decimal.Zero`, `adjustment_amount = decimal.Zero` on every invoice. These fields exist in the schema but are never populated. The settlement formula **produces wrong numbers on every invoice ever generated.**

2. **Commission is always zero** — Invoice sums `commission_amount` from `order_pickups`, but `order_pickups.commission_amount` is always `0` (see `redesign-order-system` bug #2). Finance correctness requires that change to be applied first.

3. **Delivery charge not split in invoices** — The `net_payable` formula in docs says delivery charge belongs to the platform. Currently the invoice doesn't track delivery charge separately at all.

4. **Ledger system is a ghost** — `ledger_accounts` and `ledger_entries` tables exist (migration 000020) but zero rows are ever inserted. No financial audit trail exists.

5. **Rider payouts never generated** — `rider_payouts` table exists but there is no job, service, or endpoint that creates payout records. Riders cannot be paid.

6. **Platform revenue untracked** — No aggregation of commission earned, delivery fees retained, or SaaS subscription revenue. No dashboard, no report.

7. **No reconciliation** — Payments come in via bKash/AamarPay; no daily job cross-checks gateway confirmations against `payment_transactions`. Mismatches are never detected.

8. **No promo funding source separation** — Promos can be platform-funded or vendor-funded (docs/requirements/08). The invoice must separate these so vendors are only charged for vendor-funded promos.

9. **No subscription billing** — The SaaS platform charges restaurants monthly subscription fees. The schema references this but no billing cycle, invoice generation, or collection exists.

## What Changes

### Backend — finance module
- **Fix `GenerateInvoice`**: populate `vendor_promo_discounts` from `promo_usages` where `funded_by = vendor`; populate `penalty_amount` from resolved `order_issues`; populate `adjustment_amount` from `invoice_adjustments` table; populate `delivery_charge_total` from `orders.delivery_charge`; populate `platform_delivery_fee_retained` per business rule
- **Fix net_payable formula**: implement exact formula from docs: `net_payable = gross_sales - item_discounts - vendor_promo_discounts - commission_amount - penalty_amount + adjustment_amount`
- **Add delivery charge split**: platform retains delivery charge by default; vendor gets it if `delivery_managed_by = vendor` (per restaurant setting)
- **Add promo funding classification**: when `promos.funded_by = platform`, the promo discount does NOT reduce `net_payable`; when `funded_by = vendor` it does

### Backend — ledger module (activate the dormant system)
- **Create ledger accounts on tenant creation**: seed `PLATFORM_REVENUE`, `VENDOR_LIABILITY` (one per restaurant), `WALLET_LIABILITY`, `RIDER_LIABILITY` accounts
- **Create ledger entries on order events** via outbox processor: `order.delivered` → debit `VENDOR_LIABILITY`, credit `PLATFORM_REVENUE` (commission amount); credit `RIDER_LIABILITY` (rider earnings estimate)
- **Create ledger entries on invoice events**: `invoice.finalized` → net_payable entry; `invoice.paid` → settlement entry; `refund.processed` → debit `PLATFORM_REVENUE`
- **Create ledger entries on wallet events**: wallet credit/debit → `WALLET_LIABILITY` entries

### Backend — rider payouts
- **Automated payout generation job**: weekly job aggregates `rider_earnings` records for the period, creates `rider_payouts` record per rider with `status = pending`
- **Payout approval flow**: admin approves payout → updates to `processing`; bKash disbursement webhook → `completed`
- **Penalty deductions**: `rider_issues` with `penalty_amount > 0` deducted from payout total
- **API endpoints**: `GET /admin/rider-payouts`, `PATCH /admin/rider-payouts/:id/approve`, `GET /admin/riders/:id/earnings-history`

### Backend — platform revenue tracking
- **Commission revenue report**: aggregate `order_pickups.commission_amount` by tenant and date range
- **Delivery fee revenue report**: aggregate `orders.delivery_charge` for platform-managed deliveries
- **Subscription revenue**: invoice generation and collection for monthly SaaS fees; `subscription_invoices` table
- **API**: `GET /admin/revenue/summary`, `GET /admin/revenue/commission`, `GET /admin/revenue/delivery-fees`

### Backend — payment reconciliation
- **Daily reconciliation job**: cross-reference `payment_transactions` records against gateway callback events; flag any `status = processing` payments older than 1 hour
- **Mismatch alerts**: insert `reconciliation_alerts` record for detected mismatches; admin notification
- **COD collection tracking**: track cash collected by riders in `cash_collection_records`

### Schema additions
- `invoice_adjustments` table: manual adjustment records (amount, reason, created_by, invoice_id)
- `subscription_invoices` table: monthly SaaS billing records per tenant
- `cash_collection_records` table: COD cash by rider per shift
- `reconciliation_alerts` table: mismatch flags for admin review
- New columns on `promos`: `funded_by ENUM('platform', 'vendor')` with default `vendor`
- New column on `restaurants`: `delivery_managed_by ENUM('platform', 'vendor')` with default `platform`

## Impact

- **Affected specs**: `invoice-settlement`, `financial-ledger`, `rider-payouts`, `platform-revenue`, `payment-reconciliation`
- **Affected code**:
  - `backend/internal/modules/finance/service.go` — all major fixes
  - `backend/internal/modules/finance/handler.go` — new payout/revenue endpoints
  - `backend/internal/modules/ledger/` — new module activating dormant tables
  - `backend/internal/worker/payout_worker.go` — new rider payout job
  - `backend/internal/worker/reconciliation_worker.go` — new daily reconciliation job
  - `backend/internal/db/migrations/000023_finance_fixes.up.sql`
- **BREAKING in numbers**: invoice `net_payable` will change once promo, penalty, and adjustment fields are correctly populated
- **Data integrity**: requires `redesign-order-system` to be deployed first (commission data must be non-zero)

## Dependencies

- **Requires**: `redesign-order-system` (commission amounts on pickups must be non-zero)
- **Required by**: `complete-partner-portal` finance feature (partner portal shows correct invoice data)
