## ADDED Requirements

### Requirement: Daily Payment Reconciliation Job
A scheduled job SHALL run daily at 02:00 UTC to reconcile `payment_transactions` records against gateway callback records. For each `payment_transactions` record with `status = processing` older than 1 hour, the job SHALL query the gateway status API (bKash or AamarPay) and update the local record.

#### Scenario: Payment stuck in processing
- **WHEN** a `payment_transactions` record has `status = processing` for more than 1 hour
- **THEN** the reconciliation job queries the gateway for the transaction status
- **WHEN** the gateway confirms success
- **THEN** the local record is updated to `success` and the order is updated accordingly
- **WHEN** the gateway confirms failure
- **THEN** the local record is updated to `failed` and the order is auto-cancelled with stock release

### Requirement: Mismatch Alert Creation
When the reconciliation job detects a payment record that succeeded at the gateway but has `status = pending/failed` locally (or vice versa), it SHALL create a `reconciliation_alerts` record and trigger an admin notification.

#### Scenario: Payment succeeded at gateway but failed locally
- **WHEN** the reconciliation job finds a gateway-successful payment that is `failed` in local records
- **THEN** a `reconciliation_alerts` record is created with `type = gateway_success_local_failure`
- **AND** the alert is surfaced in the admin dashboard
- **AND** the finance admin receives a push/email notification

#### Scenario: Admin resolves mismatch
- **WHEN** an admin calls `PATCH /admin/reconciliation-alerts/:id/resolve` with a resolution action
- **THEN** the alert `status` transitions to `resolved`
- **AND** the corrective action (manual refund or payment capture) is recorded in `resolution_notes`

### Requirement: COD Cash Collection Tracking
For Cash-on-Delivery orders, the system SHALL track cash collection per rider shift. A `cash_collection_records` table stores: `rider_id`, `order_id`, `amount`, `collected_at`, `remitted_at` (nullable), `status (collected|remitted|overdue)`. After delivery, a cash record is automatically created. Riders remit cash in bulk at end of shift.

#### Scenario: COD order delivered
- **WHEN** a rider marks a COD order as delivered
- **THEN** a `cash_collection_records` record is created with `status = collected`, `amount = order.total_amount`

#### Scenario: Rider remits cash at end of shift
- **WHEN** a rider calls `POST /rider/cash-remittance` with total cash amount
- **THEN** all unremitted `cash_collection_records` for that rider are marked `status = remitted`, `remitted_at = NOW()`
- **AND** a `CASH` ledger debit entry is created

#### Scenario: Cash overdue alert
- **WHEN** a `cash_collection_records` record has `status = collected` for more than 24 hours
- **THEN** the record transitions to `status = overdue`
- **AND** an admin alert is created

### Requirement: Gateway Fee Tracking
Each `payment_transactions` record SHALL track the gateway fee charged by bKash/AamarPay. A `gateway_fee` column (NUMERIC 10,2) SHALL be added. On successful payment, the gateway fee SHALL be populated from the webhook payload. The platform revenue report SHALL show net revenue after gateway fees.

#### Scenario: bKash payment processes with gateway fee
- **WHEN** a bKash payment webhook arrives with `gateway_fee = 15.00`
- **THEN** `payment_transactions.gateway_fee = 15.00`
- **AND** the net platform revenue for this transaction is `commission_amount - gateway_fee`
