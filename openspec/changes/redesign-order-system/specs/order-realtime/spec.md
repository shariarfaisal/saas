## ADDED Requirements

### Requirement: SSE Order Stream via Redis Pub/Sub
The `GET /api/v1/orders/:id/stream` endpoint SHALL establish a real SSE connection by subscribing to the Redis pub/sub channel `order:{orderID}`. Events published on that channel SHALL be pushed to the connected client as `text/event-stream`. The outbox event processor SHALL publish status change events to this channel after each state transition. The handler SHALL send a heartbeat `data: ping\n\n` every 30 seconds to detect disconnected clients.

#### Scenario: Customer tracks order in real-time
- **WHEN** a customer opens the order tracking page and connects to `GET /orders/:id/stream`
- **THEN** the server subscribes to Redis channel `order:{orderID}`
- **AND** each subsequent order status change publishes a message to that channel
- **AND** the customer's browser receives the event within 500ms of the status change

#### Scenario: SSE heartbeat keeps connection alive
- **WHEN** no order events occur for 30 seconds
- **THEN** the server sends `data: {"type":"ping"}\n\n`
- **AND** the client uses this to detect connection health

#### Scenario: Client disconnects mid-stream
- **WHEN** a customer closes the tracking page
- **THEN** the server detects the disconnection via context cancellation
- **AND** unsubscribes from the Redis channel within 5 seconds
- **AND** no goroutine leak occurs

### Requirement: Partner Portal Order Stream
The `GET /api/v1/partner/restaurants/:id/orders/stream` endpoint SHALL push all new and updated orders for a specific restaurant to the connected partner portal client. This uses the Redis channel `tenant:{tenantID}:restaurant:{restaurantID}:orders`. New order events include the full order summary payload. Status change events include the order ID and new status.

#### Scenario: Restaurant receives new order notification
- **WHEN** a new order is placed for a restaurant
- **AND** a partner portal client is connected to the restaurant's order stream
- **THEN** the client receives a `new_order` SSE event with the order summary within 1 second

#### Scenario: Partner portal shows live status updates
- **WHEN** a rider picks up an order from restaurant A
- **AND** the partner portal for that restaurant is open
- **THEN** the order kanban card moves to the appropriate column automatically without a page refresh

### Requirement: Outbox Processor Publishes to Redis
The existing outbox event processor (`internal/worker/outbox.go`) SHALL be extended with a Redis PUBLISH step for events of type `order.status_changed` and `order.created`. These SHALL be published to the appropriate channels before the outbox event is marked processed.

#### Scenario: Status change triggers Redis publish
- **WHEN** an order transitions from `confirmed → preparing`
- **THEN** an outbox event `order.status_changed` is created in the same database transaction
- **AND** the outbox processor publishes `{"type":"status_changed","orderID":"...","status":"preparing","actor":"restaurant","timestamp":"..."}` to `order:{orderID}` on Redis
- **AND** the outbox event is marked `processed = true` after successful publish

### Requirement: Rider Assignment Visibility
The order detail API response SHALL include `rider_assignment` object when a rider has been assigned, containing: `rider_id`, `rider_name`, `rider_phone_last4`, `assigned_at`, and `picked_at` (if applicable). The customer order tracking API SHALL include this data.

#### Scenario: Customer sees rider info after assignment
- **WHEN** a rider is assigned to an order
- **AND** the customer calls `GET /orders/:id`
- **THEN** the response includes `rider_assignment.rider_name` and `rider_assignment.rider_phone_last4`
- **AND** actual phone number is not exposed (only last 4 digits for security)
