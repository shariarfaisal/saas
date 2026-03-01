## ADDED Requirements

### Requirement: Weekly Automated Rider Payout Generation
A scheduled job SHALL run every Monday at 00:00 tenant local time to generate `rider_payouts` records for the previous week. The job SHALL aggregate all `rider_earnings` records for each rider in the period where `payout_id IS NULL`, sum them (minus any unresolved `rider_issues.penalty_amount`), create a `rider_payouts` record with `status = pending`, and link all earnings rows to the payout via `payout_id`.

#### Scenario: Weekly payout job runs
- **WHEN** the payout job runs on Monday for week 2025-W12
- **THEN** for each active rider: a `rider_payouts` record is created with `total_earnings`, `total_penalties`, `net_payout = total_earnings - total_penalties`
- **AND** all `rider_earnings` rows for the period are linked with `payout_id`
- **AND** riders with zero earnings receive no payout record

#### Scenario: Rider with penalty in the period
- **WHEN** a rider had 3000 BDT in earnings but a 200 BDT penalty from a resolved `rider_issues` record
- **THEN** the payout's `net_payout = 2800.00`
- **AND** the penalty is itemized in the payout record

### Requirement: Admin Payout Approval Workflow
The admin SHALL approve or reject pending payouts. `POST /admin/rider-payouts/:id/approve` transitions the payout to `processing` and initiates the bKash disbursement via the payment gateway API. The bKash disbursement webhook (`POST /webhooks/bkash/disburse`) transitions the payout to `completed` or `failed`.

#### Scenario: Admin approves a payout
- **WHEN** an admin calls `POST /admin/rider-payouts/:id/approve`
- **THEN** the payout `status` transitions to `processing`
- **AND** a bKash B2C disbursement request is initiated for the rider's registered mobile number
- **AND** a ledger entry is created: debit `RIDER_LIABILITY`, credit `CASH`

#### Scenario: bKash disbursement succeeds
- **WHEN** bKash sends a success webhook for the disbursement
- **THEN** the payout `status` transitions to `completed`
- **AND** `paid_at` is set to the webhook timestamp

#### Scenario: bKash disbursement fails
- **WHEN** bKash sends a failure webhook (e.g., invalid mobile number)
- **THEN** the payout `status` transitions to `failed`
- **AND** an admin notification is created for manual follow-up
- **AND** the `RIDER_LIABILITY` ledger entry is reversed

### Requirement: Rider Earnings History API
Riders SHALL be able to view their own earnings history. `GET /rider/earnings` returns paginated earnings by order, grouped by week, with totals for the current unpaid period.

#### Scenario: Rider views earnings
- **WHEN** a rider calls `GET /rider/earnings?period=current_week`
- **THEN** the response includes all orders delivered this week with per-order earnings
- **AND** the response includes `weekly_total` and `pending_payout_date`

### Requirement: Rider Earnings on Order Delivered
When an `order.delivered` event is processed, the system SHALL create a `rider_earnings` record for the assigned rider: `order_id`, `rider_id`, `base_earning`, `tip_amount`, `distance_bonus` (if applicable), `total_earning`. The `base_earning` is derived from `platform_configs['rider_base_delivery_earning']` (configurable per tenant).

#### Scenario: Rider earns on delivery
- **WHEN** an order is marked delivered
- **THEN** a `rider_earnings` record is created with `base_earning` from platform config
- **AND** the `RIDER_LIABILITY` ledger account is credited by `total_earning`
