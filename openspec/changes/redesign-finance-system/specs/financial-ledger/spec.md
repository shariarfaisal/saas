## ADDED Requirements

### Requirement: Ledger Account Seeding on Tenant Creation
When a new tenant is created, the system SHALL automatically seed the following `ledger_accounts` records:
- `PLATFORM_COMMISSION_REVENUE` — tracks commission income
- `PLATFORM_DELIVERY_REVENUE` — tracks delivery charge income
- One `VENDOR_LIABILITY_{restaurantID}` per restaurant created under the tenant (created lazily on restaurant creation)
- `WALLET_LIABILITY` — outstanding wallet balances owed to customers
- `RIDER_LIABILITY` — unpaid rider earnings

#### Scenario: New tenant is created
- **WHEN** a new tenant is provisioned
- **THEN** `PLATFORM_COMMISSION_REVENUE`, `PLATFORM_DELIVERY_REVENUE`, and `WALLET_LIABILITY` ledger accounts are created for that tenant

#### Scenario: New restaurant is added to tenant
- **WHEN** a restaurant is created under a tenant
- **THEN** a `VENDOR_LIABILITY_{restaurantID}` ledger account is created for that restaurant

### Requirement: Ledger Entry on Order Delivered
When an `order.delivered` outbox event is processed, the ledger SHALL record double-entry transactions:

| Entry | Debit | Credit | Amount |
|---|---|---|---|
| Commission earned | `VENDOR_LIABILITY_{restaurantID}` | `PLATFORM_COMMISSION_REVENUE` | `commission_amount` |
| Rider earnings accrued | `VENDOR_LIABILITY_rider` | `RIDER_LIABILITY` | `rider_earnings_estimate` |
| Delivery revenue (platform-managed) | `VENDOR_LIABILITY_customer` | `PLATFORM_DELIVERY_REVENUE` | `delivery_charge` |

#### Scenario: Order delivered triggers ledger entries
- **WHEN** the outbox processor handles `order.delivered`
- **THEN** double-entry ledger entries are created for commission, rider earnings, and delivery revenue
- **AND** the sum of all debit amounts equals the sum of all credit amounts (balanced books)

### Requirement: Ledger Entry on Invoice Events
When an invoice transitions to `finalized`, the system SHALL create a ledger entry for the `net_payable` against the `VENDOR_LIABILITY` account. When an invoice transitions to `paid`, the system SHALL create a settlement entry that zeroes out the liability.

#### Scenario: Invoice finalized
- **WHEN** an invoice's status changes to `finalized`
- **THEN** a ledger entry is created: `VENDOR_LIABILITY_NET = net_payable` (debit/credit depending on sign)

#### Scenario: Invoice paid
- **WHEN** an invoice is marked `paid`
- **THEN** a settlement entry reverses the finalized liability entry, recording actual cash movement

### Requirement: Ledger Entry on Wallet Events
Every `wallet_transactions` record (credit or debit) SHALL generate a corresponding `ledger_entry` on the `WALLET_LIABILITY` account.

#### Scenario: Customer wallet is topped up
- **WHEN** a customer tops up their wallet by 500 BDT
- **THEN** a ledger entry debits `CASH` and credits `WALLET_LIABILITY` for 500 BDT

#### Scenario: Customer uses wallet to pay for order
- **WHEN** a wallet payment is processed
- **THEN** a ledger entry debits `WALLET_LIABILITY` and credits `PLATFORM_COMMISSION_REVENUE` (for commission portion)

### Requirement: Audit Trail via Ledger
All `ledger_entries` records SHALL be immutable — no UPDATE or DELETE is permitted. Corrections are made via reversal entries. The `GET /admin/ledger` endpoint SHALL support filtering by `account_id`, `date_range`, `entry_type`, and pagination.

#### Scenario: Finance admin reviews commission for a month
- **WHEN** an admin calls `GET /admin/ledger?account=PLATFORM_COMMISSION_REVENUE&from=2025-01-01&to=2025-01-31`
- **THEN** all commission ledger entries for January are returned
- **AND** the sum of credit entries matches total commission earned for the period
