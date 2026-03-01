## ADDED Requirements

### Requirement: Real-time Order Kanban Board
The orders kanban board SHALL fetch live orders from `GET /partner/orders?restaurant_id=:id&status=new,confirmed,preparing,ready,picked` and display them in the five-column board. Columns SHALL update in real time via SSE events (same channel as dashboard). Status transitions SHALL call the appropriate backend endpoint and optimistically update the card's column.

#### Scenario: Kanban board loads real orders
- **WHEN** a partner navigates to the orders page
- **THEN** orders are fetched from the API and displayed in the correct status columns
- **AND** a loading skeleton is shown while fetching

#### Scenario: Partner accepts a new order
- **WHEN** the partner clicks "Accept" on a New order card
- **THEN** `PATCH /partner/orders/:id/confirm` is called
- **AND** the card moves to the Confirmed column optimistically
- **AND** an error toast and column revert occur if the API call fails

#### Scenario: Partner rejects an order
- **WHEN** the partner clicks "Reject" and provides a rejection reason
- **THEN** `PATCH /partner/orders/:id/reject` is called with the reason
- **AND** the card is removed from the board

#### Scenario: Partner marks order as Ready
- **WHEN** the partner clicks "Mark Ready" on a Preparing card
- **THEN** `PATCH /partner/orders/:id/ready` is called
- **AND** the card moves to the Ready column

#### Scenario: New order arrives via SSE while kanban is open
- **WHEN** an SSE `order.created` event arrives for the current restaurant
- **THEN** the new order card appears in the New column immediately
- **AND** the audio notification plays

### Requirement: Order Detail Drawer with Live Data
The order detail drawer SHALL load full order data from `GET /partner/orders/:id` including order items with addons, customer info, delivery address, payment method and status, assigned rider info, and the order status timeline. The drawer SHALL be accessible by clicking any order card.

#### Scenario: Order detail drawer shows full data
- **WHEN** the partner clicks on an order card
- **THEN** the drawer opens and fetches `GET /partner/orders/:id`
- **AND** all sections (items, customer, rider, payment, timeline) are populated with real data

#### Scenario: Order detail shows rider assignment
- **WHEN** a rider has been assigned to the order
- **THEN** the rider's name, phone, and current status are shown in the drawer

### Requirement: Order History Table with Filtering
The order history tab SHALL fetch paginated past orders from `GET /partner/orders/history?restaurant_id=:id&status=:status&from=:date&to=:date&q=:query` and display them in a searchable, filterable table with pagination controls.

#### Scenario: Partner searches orders by customer phone
- **WHEN** the partner types a phone number in the search field
- **THEN** the table refetches with `q=<phone>` and shows matching orders

#### Scenario: Partner filters by date range
- **WHEN** the partner selects a date range
- **THEN** the table refetches with `from` and `to` params and shows orders in that range

### Requirement: Accept/Reject Countdown Timer
New incoming orders displayed in the dashboard panel and the kanban New column SHALL show a 3-minute countdown timer. When the timer expires with no action, the system SHALL automatically call the reject endpoint with reason `auto_rejected_timeout`.

#### Scenario: Auto-reject on timer expiry
- **WHEN** a new order's 3-minute timer expires
- **THEN** `PATCH /partner/orders/:id/reject` is called with `reason=auto_rejected_timeout`
- **AND** the order card is removed from the board
- **AND** a toast notification informs the partner the order was auto-rejected
