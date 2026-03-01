## ADDED Requirements

### Requirement: Platform Commission Revenue Report
The admin SHALL have access to a commission revenue report that aggregates `order_pickups.commission_amount` by tenant, restaurant, and time period. The report SHALL support `daily`, `weekly`, `monthly`, and `custom` date ranges.

#### Scenario: Admin views monthly commission report
- **WHEN** an admin calls `GET /admin/revenue/commission?period=monthly&month=2025-03`
- **THEN** the response includes total commission earned per tenant for March 2025
- **AND** a breakdown per restaurant within each tenant
- **AND** a platform-wide total

#### Scenario: Commission report reflects only finalized invoices
- **WHEN** an invoice is in `draft` status
- **THEN** its commission is included in the `unfinalized` bucket, not the `earned` total
- **WHEN** the invoice is finalized
- **THEN** the commission moves to the `earned` bucket

### Requirement: Platform Delivery Fee Revenue
The admin SHALL view delivery charge revenue retained by the platform. This aggregates `orders.delivery_charge` for all orders where `restaurant.delivery_managed_by = platform`.

#### Scenario: Admin views delivery revenue
- **WHEN** an admin calls `GET /admin/revenue/delivery-fees?period=monthly&month=2025-03`
- **THEN** the response includes total delivery fees collected by the platform in March 2025

### Requirement: SaaS Subscription Billing
Each tenant SHALL be billed a monthly subscription fee defined in `tenant_subscriptions` (tier, amount, billing_day). On the billing day, the system SHALL auto-generate a `subscription_invoices` record with `status = pending`. The tenant admin is notified and has 7 days to pay before suspension flag is set.

#### Scenario: Monthly billing cycle runs
- **WHEN** a tenant's billing day arrives
- **THEN** a `subscription_invoices` record is created with `amount = subscription_tier_amount`
- **AND** the tenant admin receives a billing notification
- **AND** `due_date = billing_date + 7 days`

#### Scenario: Tenant pays subscription
- **WHEN** the tenant pays the subscription invoice via bKash/AamarPay
- **THEN** `subscription_invoices.status = paid`
- **AND** a ledger entry credits `PLATFORM_SUBSCRIPTION_REVENUE`

#### Scenario: Tenant overdue on subscription
- **WHEN** a subscription invoice is not paid by `due_date`
- **THEN** `tenants.subscription_status = overdue`
- **AND** partner portal access is restricted to read-only (no new order creation)

### Requirement: Revenue Summary Dashboard
The admin dashboard SHALL display a revenue summary with:
- Total platform revenue (commission + delivery fees + subscription) for current month vs previous month
- Top 5 revenue-generating tenants
- Commission vs delivery fee vs subscription breakdown (pie chart data)
- Outstanding vendor payables (sum of finalized but unpaid invoices)

#### Scenario: Admin loads dashboard
- **WHEN** an admin loads `GET /admin/revenue/summary`
- **THEN** the response includes current month and previous month totals
- **AND** percent change is calculated for each revenue category
