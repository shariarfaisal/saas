## ADDED Requirements

### Requirement: Idempotent Order Creation
The `POST /api/v1/orders` endpoint SHALL require an `Idempotency-Key` header. The server SHALL store the response against `(tenant_id, user_id, idempotency_key)` in the `idempotency_keys` table with 24-hour TTL. Repeated requests with the same key and identical payload SHALL return the original response without creating a duplicate order. Repeated requests with the same key but different payload SHALL return `409 IDEMPOTENCY_KEY_REUSED_WITH_DIFFERENT_PAYLOAD`.

#### Scenario: Mobile client retries after network timeout
- **WHEN** a customer's app sends `POST /orders` with `Idempotency-Key: abc-123`
- **AND** the first request succeeds but the response is lost due to a network error
- **AND** the app retries with the same `Idempotency-Key: abc-123` and same body
- **THEN** the server returns the original order response
- **AND** no second order is created
- **AND** no second payment is charged

#### Scenario: Missing idempotency key
- **WHEN** a client submits `POST /orders` without the `Idempotency-Key` header
- **THEN** the API returns `400 MISSING_IDEMPOTENCY_KEY`

### Requirement: Order Number via Sequence
The system SHALL generate order numbers using a PostgreSQL sequence per tenant (e.g., `order_seq_<tenant_slug>`). The format is `<TENANT_PREFIX>-<6-DIGIT-PADDED-NUMBER>`. This replaces the current `COUNT(*) + 1` approach which causes O(n) table scans and race conditions under concurrent load.

#### Scenario: Concurrent order creation under load
- **WHEN** 100 orders are created simultaneously for the same tenant
- **THEN** each order receives a unique, sequential order number
- **AND** no duplicates or gaps due to race conditions occur

#### Scenario: Order numbers are human-readable
- **WHEN** an order is created for tenant with prefix `KBC`
- **THEN** the order number is formatted as `KBC-000001`, `KBC-000002`, etc.

### Requirement: Stock Consumed on Rider Pickup
The system SHALL call `ConsumeReservedStock` for each inventory-tracked item when a rider marks a restaurant pickup as `PICKED`. This permanently deducts `stock_qty` and reduces `reserved_qty`. Prior to this, only `reserved_qty` is incremented; actual stock is consumed on physical pickup.

#### Scenario: Stock consumed when rider picks up order
- **WHEN** a rider calls the pickup endpoint for a restaurant that had inventory-tracked items
- **THEN** for each tracked item: `inventory_items.stock_qty -= quantity` and `reserved_qty -= quantity`
- **AND** an `inventory_adjustments` record with `adjustment_type = order_consume` and the `order_id` is created

#### Scenario: Non-tracked items are skipped
- **WHEN** an order contains products with `is_inv_tracked = false`
- **THEN** no inventory update occurs for those products

### Requirement: Atomic Stock Reservation with Row Lock
The `ReserveStock` query SHALL use `SELECT ... FOR UPDATE` row locking before updating `reserved_qty` to prevent concurrent overselling. The reservation SHALL fail immediately (not block) for products already locked by another transaction using `SKIP LOCKED`.

#### Scenario: Two simultaneous orders for the last item
- **WHEN** only 1 unit of a product is in stock and two customers submit orders simultaneously
- **THEN** exactly one order succeeds and the other fails with `422 PRODUCT_OUT_OF_STOCK`
- **AND** `stock_qty - reserved_qty >= 0` is always maintained

### Requirement: Stock Released on Cancellation
The system SHALL call `ReleaseStock` for all inventory-tracked items when an order transitions to `CANCELLED` or `REJECTED` (at any stage where stock was reserved but not yet consumed). This restores the available quantity for other orders.

#### Scenario: Customer cancels before pickup
- **WHEN** a customer cancels an order in `CREATED` or `CONFIRMED` status
- **THEN** reserved stock for all inventory-tracked items is released
- **AND** inventory_adjustments records with `adjustment_type = order_release` are created
