## ADDED Requirements

### Requirement: MarkDelivered Endpoint
The system SHALL provide `PATCH /rider/orders/:id/deliver` allowing the assigned rider to mark an order as delivered. This transitions the order from `PICKED → DELIVERED`, sets `delivered_at` timestamp, creates a timeline event with actor `rider`, and publishes an `order.delivered` outbox event that triggers: rider earnings creation, customer cashback credit (if promo has `cashback_amount > 0`), rating prompt push notification, and contribution to the invoice settlement cycle.

#### Scenario: Rider marks order delivered
- **WHEN** an authenticated rider calls `PATCH /rider/orders/:id/deliver` for an order in `PICKED` status
- **THEN** the order status transitions to `DELIVERED`
- **AND** `delivered_at` is set to the current timestamp
- **AND** a timeline event is created with `event_type = status_changed`, `actor_type = rider`
- **AND** an `order.delivered` outbox event is inserted for async processing

#### Scenario: Rider attempts to deliver non-PICKED order
- **WHEN** a rider calls the deliver endpoint for an order not in `PICKED` status
- **THEN** the API returns `422 INVALID_STATUS_TRANSITION` with a descriptive message

#### Scenario: Wrong rider attempts delivery
- **WHEN** a rider calls the deliver endpoint for an order assigned to a different rider
- **THEN** the API returns `403 FORBIDDEN`

### Requirement: Enforced State Machine
The system SHALL enforce valid status transitions for all order status changes. Attempting an invalid transition SHALL return `422 INVALID_STATUS_TRANSITION`. The valid transitions are:

| From | To | Trigger |
|---|---|---|
| `pending` | `created` | Payment confirmed or COD |
| `created` | `confirmed` | Restaurant confirm or auto-confirm |
| `created` | `rejected` | Restaurant reject with reason |
| `created` | `cancelled` | Customer cancel or payment timeout |
| `confirmed` | `preparing` | Restaurant marks preparing |
| `preparing` | `ready` | Restaurant marks ready |
| `ready` | `picked` | Rider picks up all pickups |
| `picked` | `delivered` | Rider marks delivered |
| Any active | `cancelled` | Admin force-cancel with reason |

#### Scenario: Restaurant tries to skip to READY without PREPARING
- **WHEN** a restaurant calls the `ready` endpoint for an order in `confirmed` status
- **THEN** the API returns `422 INVALID_STATUS_TRANSITION`
- **AND** the order status is unchanged

#### Scenario: Valid transition succeeds
- **WHEN** a restaurant calls `PATCH /partner/orders/:id/preparing` for an order in `confirmed` status
- **THEN** the pickup status transitions to `preparing`
- **AND** the parent order status transitions to `preparing`

### Requirement: Multi-Pickup Status Aggregation (Race-Safe)
Parent order status updates from pickup status changes SHALL be performed in a single atomic transaction using a `SELECT ... FOR UPDATE` on all pickups for the order before evaluating the aggregate. This prevents race conditions when two restaurants update their pickup status concurrently.

#### Scenario: Two restaurants update simultaneously
- **WHEN** restaurant A and restaurant B both mark their pickups as `ready` at the exact same millisecond
- **THEN** exactly one of the two concurrent writes wins the row lock
- **AND** the parent order transitions to `ready` exactly once (not zero times, not twice)

#### Scenario: Parent order reaches READY only when all pickups are ready
- **WHEN** an order has 2 restaurants and only restaurant A marks ready
- **THEN** the parent order status remains `preparing` (or `confirmed`)
- **WHEN** restaurant B also marks ready
- **THEN** the parent order transitions to `ready`

### Requirement: Auto-Cancel PENDING Orders on Payment Timeout
A scheduled job SHALL cancel orders with `status = 'pending'` and `payment_status = 'unpaid'` that were created more than 30 minutes ago. On cancellation: reserved stock is released via `ReleaseStock`, a `CANCELLED` timeline event is created with `actor_type = system`, and an `order.payment_timeout` outbox event is inserted.

#### Scenario: Online payment never completed
- **WHEN** a customer initiates a bKash payment and closes the app without completing
- **AND** 30 minutes elapse
- **THEN** the background job cancels the order
- **AND** reserved stock is released
- **AND** the customer receives a push notification that their order was cancelled due to payment timeout

### Requirement: Payment Timeout Configurable Per Tenant
The payment timeout threshold SHALL be configurable per tenant via `platform_configs` key `payment_timeout_minutes`, defaulting to `30`. Individual tenants MAY override via `tenants.settings->>payment_timeout_minutes`.

#### Scenario: Tenant configures custom payment timeout
- **WHEN** a tenant sets `payment_timeout_minutes = 15` in their settings
- **THEN** PENDING orders for that tenant are auto-cancelled after 15 minutes, not 30
