# Tasks: redesign-finance-system

## Section 0 — Schema & Migrations

- [ ] 0.1 Write migration `000023_finance_fixes.up.sql`:
  - `ALTER TABLE promos ADD COLUMN funded_by TEXT NOT NULL DEFAULT 'vendor'`
  - `ALTER TABLE restaurants ADD COLUMN delivery_managed_by TEXT NOT NULL DEFAULT 'platform'`
  - `ALTER TABLE payment_transactions ADD COLUMN gateway_fee NUMERIC(10,2) NOT NULL DEFAULT 0`
  - `CREATE TABLE invoice_adjustments (id UUID PK, invoice_id UUID FK, amount NUMERIC(10,2), direction TEXT, reason TEXT, created_by_admin_id UUID FK, created_at TIMESTAMPTZ)`
  - `CREATE TABLE subscription_invoices (id UUID PK, tenant_id UUID FK, amount NUMERIC(10,2), status TEXT, billing_period_start DATE, billing_period_end DATE, due_date DATE, paid_at TIMESTAMPTZ, created_at TIMESTAMPTZ)`
  - `CREATE TABLE cash_collection_records (id UUID PK, tenant_id UUID FK, rider_id UUID FK, order_id UUID FK, amount NUMERIC(10,2), status TEXT, collected_at TIMESTAMPTZ, remitted_at TIMESTAMPTZ)`
  - `CREATE TABLE reconciliation_alerts (id UUID PK, tenant_id UUID FK, payment_transaction_id UUID FK, alert_type TEXT, status TEXT DEFAULT 'open', resolution_notes TEXT, resolved_by UUID, resolved_at TIMESTAMPTZ, created_at TIMESTAMPTZ)`
  - `ALTER TABLE tenants ADD COLUMN billing_day INT NOT NULL DEFAULT 1`
  - `ALTER TABLE invoices ADD COLUMN delivery_charge_total NUMERIC(12,2) NOT NULL DEFAULT 0`
- [ ] 0.2 Write down migration for 000023
- [ ] 0.3 Regenerate SQLC for all affected tables

## Section 1 — Fix Invoice Generation

- [ ] 1.1 Add SQLC query `GetVendorPromoDiscountsForInvoice(restaurantID, periodStart, periodEnd)` — joins `promo_usages → promos` filtering `funded_by = 'vendor'`
- [ ] 1.2 Add SQLC query `GetPenaltyAmountForInvoice(restaurantID, periodStart, periodEnd)` — sums `order_issues.penalty_amount WHERE resolved = true`
- [ ] 1.3 Add SQLC query `GetInvoiceAdjustmentTotal(invoiceID)` — sums `invoice_adjustments.amount` by direction
- [ ] 1.4 Add SQLC query `GetDeliveryChargeTotal(restaurantID, periodStart, periodEnd)` — sums `orders.delivery_charge WHERE restaurant.delivery_managed_by = 'platform'`
- [ ] 1.5 Update `finance/service.go:GenerateInvoice` to call all 4 new queries and populate the fields
- [ ] 1.6 Implement net_payable formula exactly as per design.md
- [ ] 1.7 Unit test: invoice with vendor promo → vendor_promo_discounts > 0, platform promo → 0
- [ ] 1.8 Unit test: invoice with penalties → penalty_amount populated
- [ ] 1.9 Unit test: invoice with manual adjustment → adjustment_amount populated
- [ ] 1.10 Integration test: full invoice generation with real data → verify all 6 fields non-zero in expected scenarios

## Section 2 — Invoice Adjustment Endpoint

- [ ] 2.1 Add `POST /admin/invoices/:id/adjustments` handler — creates `invoice_adjustments` record
- [ ] 2.2 Validate invoice is in `draft` status; return 422 if `finalized`
- [ ] 2.3 Add `GET /admin/invoices/:id/adjustments` to list adjustments for an invoice
- [ ] 2.4 Unit test: create adjustment on draft invoice → success; on finalized → 422

## Section 3 — Promo Funding Classification

- [ ] 3.1 Add `funded_by` field to promo create/update API request/response
- [ ] 3.2 Validate `funded_by IN ('platform', 'vendor')` on input
- [ ] 3.3 Update promo create service to persist `funded_by`
- [ ] 3.4 Add `funded_by` to promo list response so partner portal can display it
- [ ] 3.5 Unit test: platform-funded promo usage → not included in vendor invoice

## Section 4 — Ledger Service Activation

- [ ] 4.1 Create `internal/modules/ledger/service.go` with `CreateEntry(ctx, debitAccountID, creditAccountID, amount, referenceType, referenceID, memo)` method
- [ ] 4.2 Add `SeedLedgerAccounts(ctx, tenantID)` method — creates platform accounts + WALLET_LIABILITY + RIDER_LIABILITY
- [ ] 4.3 Add `SeedRestaurantLedgerAccount(ctx, tenantID, restaurantID)` — creates VENDOR_LIABILITY account
- [ ] 4.4 Hook `SeedLedgerAccounts` into tenant creation flow
- [ ] 4.5 Hook `SeedRestaurantLedgerAccount` into restaurant creation flow
- [ ] 4.6 Run `SeedLedgerAccounts` for all existing tenants in a one-time migration/script

## Section 5 — Ledger Entries from Order Events

- [ ] 5.1 In outbox processor, handle `order.delivered`: create commission ledger entry (debit VENDOR_LIABILITY, credit PLATFORM_COMMISSION_REVENUE)
- [ ] 5.2 In outbox processor, handle `order.delivered`: create rider earnings ledger entry (debit RIDER_LIABILITY, credit RIDER_EARNINGS_PAYABLE)
- [ ] 5.3 In outbox processor, handle `order.delivered` for platform-delivery orders: create delivery revenue entry (debit VENDOR_LIABILITY, credit PLATFORM_DELIVERY_REVENUE)
- [ ] 5.4 In finance service, on `invoice.finalized`: create net_payable ledger entry
- [ ] 5.5 In finance service, on `invoice.paid`: create settlement entry
- [ ] 5.6 In wallet service, on every `wallet_transactions` insert: create corresponding ledger entry
- [ ] 5.7 Integration test: deliver order → verify 2-3 balanced ledger entries in DB
- [ ] 5.8 Unit test: sum of all debit entries == sum of all credit entries for any order

## Section 6 — Rider Earnings on Delivery

- [ ] 6.1 In outbox processor handle `order.delivered`: create `rider_earnings` record for assigned rider
- [ ] 6.2 `base_earning` from `platform_configs['rider_base_delivery_earning']` (default 40 BDT)
- [ ] 6.3 `total_earning = base_earning + tip_amount` (tip from order, if any)
- [ ] 6.4 Unit test: delivered order → rider_earnings record created with correct amounts

## Section 7 — Automated Rider Payout Job

- [ ] 7.1 Create `internal/worker/payout_worker.go` implementing `GenerateWeeklyPayouts(ctx)` asynq periodic task
- [ ] 7.2 Aggregate `rider_earnings` by rider where `payout_id IS NULL` for the previous week
- [ ] 7.3 Lookup `rider_issues.penalty_amount` for riders in the period; deduct from total
- [ ] 7.4 Create `rider_payouts` record per rider with `status = pending`, link earnings rows with `payout_id`
- [ ] 7.5 Register as Monday 00:00 cron task in `cmd/worker/main.go`
- [ ] 7.6 Unit test: 5 earnings + 1 penalty → payout created with correct net_payout

## Section 8 — Admin Payout Approval Flow

- [ ] 8.1 Add `GET /admin/rider-payouts` — list all payouts with pagination and status filter
- [ ] 8.2 Add `POST /admin/rider-payouts/:id/approve` — transitions status to `processing`, initiates bKash B2C disbursement
- [ ] 8.3 Add bKash disbursement webhook handler `POST /webhooks/bkash/disburse` — updates payout to `completed` or `failed`
- [ ] 8.4 On `failed` disbursement (2nd attempt): flag for manual bank transfer, notify admin
- [ ] 8.5 On approval: create ledger entry debit RIDER_LIABILITY, credit CASH
- [ ] 8.6 Add `GET /rider/earnings` — rider's own earnings history with week totals
- [ ] 8.7 Integration test: approve payout → disbursement initiated; webhook success → status completed

## Section 9 — Subscription Billing

- [ ] 9.1 Create `internal/worker/subscription_billing_worker.go` daily job that checks `tenants.billing_day`
- [ ] 9.2 Generate `subscription_invoices` record on billing day if not already generated this month
- [ ] 9.3 Send billing notification to tenant admin
- [ ] 9.4 Set `due_date = billing_date + 7 days`; after due date, set `tenants.subscription_status = overdue`
- [ ] 9.5 Add `GET /admin/subscription-invoices` — list all subscription invoices with filters
- [ ] 9.6 Unit test: billing day = 1, today = Mar 1 → invoice generated; run again same day → no duplicate

## Section 10 — Platform Revenue Reports

- [ ] 10.1 Add SQLC query `GetCommissionRevenue(tenantID, dateRange)` — sum commission_amount from order_pickups
- [ ] 10.2 Add SQLC query `GetDeliveryFeeRevenue(tenantID, dateRange)` — sum delivery_charge for platform-managed
- [ ] 10.3 Add `GET /admin/revenue/summary` — current month vs previous month totals for all revenue categories
- [ ] 10.4 Add `GET /admin/revenue/commission` — paginated commission breakdown by restaurant
- [ ] 10.5 Add `GET /admin/revenue/delivery-fees` — paginated delivery fee breakdown by tenant
- [ ] 10.6 Unit test: summary includes all 3 revenue categories; percent change calculated correctly

## Section 11 — Payment Reconciliation

- [ ] 11.1 Create `internal/worker/reconciliation_worker.go` daily job at 02:00 UTC
- [ ] 11.2 Query `payment_transactions WHERE status = 'processing' AND created_at < NOW() - INTERVAL '1 hour'`
- [ ] 11.3 For each: call gateway status API (bKash or AamarPay) based on `gateway_type`
- [ ] 11.4 On gateway success + local processing: update to `success`; trigger order completion flow
- [ ] 11.5 On gateway failure + local processing: update to `failed`; trigger auto-cancel
- [ ] 11.6 On mismatch: create `reconciliation_alerts` record, send admin notification
- [ ] 11.7 Add `GET /admin/reconciliation-alerts` — list open alerts with filter by type and date
- [ ] 11.8 Add `PATCH /admin/reconciliation-alerts/:id/resolve` — mark resolved with notes
- [ ] 11.9 Unit test: processing payment older than 1h → gateway queried; mismatch → alert created

## Section 12 — COD Cash Collection

- [ ] 12.1 In `MarkDelivered` handler: if `order.payment_method = 'cod'`, create `cash_collection_records` record with `status = collected`
- [ ] 12.2 Add `POST /rider/cash-remittance` — rider submits collected cash; bulk-update records to `remitted`; create ledger debit entry
- [ ] 12.3 Daily job: flag `cash_collection_records WHERE status = 'collected' AND collected_at < NOW() - INTERVAL '24 hours'` as `overdue`; create admin alert
- [ ] 12.4 Add `GET /admin/cash-collections` — list all collection records with status filter
- [ ] 12.5 Unit test: COD delivery → cash record created; remittance → records updated + ledger entry

## Section 13 — Gateway Fee Tracking

- [ ] 13.1 Update bKash payment success handler: extract `gateway_fee` from webhook payload; store in `payment_transactions.gateway_fee`
- [ ] 13.2 Update AamarPay payment success handler similarly
- [ ] 13.3 If gateway fee not in webhook: derive from `platform_configs['bkash_fee_rate']` (default 0.015) or `aamarPay_fee_rate`
- [ ] 13.4 Update revenue report: show `gross_revenue`, `gateway_fees`, `net_revenue`
- [ ] 13.5 Unit test: bKash webhook with gateway_fee field → stored correctly

## Section 14 — Admin Ledger API

- [ ] 14.1 Add `GET /admin/ledger` — list ledger entries with filters: `account_id`, `from`, `to`, `entry_type`; paginated
- [ ] 14.2 Add `GET /admin/ledger/accounts` — list all ledger accounts with current balance (SUM credit - debit)
- [ ] 14.3 Unit test: create several entries → balance query returns correct sum

## Section 15 — Validation & Smoke Testing

- [ ] 15.1 Run migration 000023 on staging; verify schema
- [ ] 15.2 Run ledger account seeding for existing tenants; verify accounts created
- [ ] 15.3 Create end-to-end test: order lifecycle → deliver → verify commission ledger entry, rider earnings record
- [ ] 15.4 Finalize an invoice: verify all 6 fields non-zero; verify net_payable matches formula
- [ ] 15.5 Run payout job manually; verify `rider_payouts` created with correct net_payout
- [ ] 15.6 Test reconciliation job: create stale processing payment → verify gateway queried, alert created on mismatch
- [ ] 15.7 Run `openspec validate redesign-finance-system --strict`
