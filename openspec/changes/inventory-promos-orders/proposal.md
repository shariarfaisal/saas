# Proposal: Inventory, Promotions Engine & Order Core

**ID:** inventory-promos-orders
**Status:** approved
**Created:** 2026-02-27

## Summary

Implement Phases 4–6 of the Munchies SaaS platform covering inventory stock management,
promotions engine with validation, and full order lifecycle including multi-restaurant
pickup coordination.

## Motivation

The platform requires:
1. **Inventory tracking** so restaurants can manage stock levels and prevent overselling
2. **Promotions engine** for marketing campaigns with validation rules
3. **Order core** for the complete order lifecycle from charge calculation through delivery

## Scope

### Phase 4 — Inventory (TASK-026)
- SQLC queries for inventory CRUD and stock operations
- Partner API for inventory management
- Stock-check hook wired into order creation

### Phase 5 — Promotions Engine (TASK-027, TASK-028)
- SQLC queries for promo CRUD and usage tracking
- Partner API for promo management
- PromoService.Validate() with comprehensive validation
- Integration with order charge calculation and creation

### Phase 6 — Order Core (TASK-029 through TASK-036)
- SQLC queries for orders, items, pickups, timeline
- Charge pre-calculation endpoint
- Atomic order creation (COD, wallet, online payment)
- Order status transitions (restaurant, customer, admin, system)
- Customer order tracking and cancellation
- Multi-restaurant pickup coordination

## Design

### Database Schema
All tables already exist via migrations 000009 (inventory), 000010 (orders), 000012 (promos).
No new migrations needed — only SQLC queries and application code.

### Module Structure
```
internal/modules/
├── inventory/
│   ├── handler.go    — HTTP handlers for partner inventory API
│   └── service.go    — Business logic for stock operations
├── promo/
│   ├── handler.go    — HTTP handlers for partner promo API
│   └── service.go    — Business logic + Validate()
└── order/
    ├── handler.go    — HTTP handlers for customer & partner order API
    └── service.go    — Order lifecycle business logic
```

### API Endpoints

#### Inventory (Partner)
- `GET /partner/inventory` — List inventory for restaurant
- `POST /partner/inventory/adjust` — Adjust stock quantity
- `GET /partner/inventory/low-stock` — List items below reorder threshold

#### Promotions (Partner)
- `GET /partner/promos` — List promos for tenant
- `POST /partner/promos` — Create promo
- `GET /partner/promos/{id}` — Get promo details
- `PUT /partner/promos/{id}` — Update promo
- `PATCH /partner/promos/{id}/deactivate` — Deactivate promo

#### Orders (Customer)
- `POST /api/v1/orders/charges/calculate` — Pre-calculate charges
- `POST /api/v1/orders` — Create order
- `GET /api/v1/orders/{id}` — Get order details
- `GET /api/v1/orders/{id}/tracking` — SSE tracking stream
- `PATCH /api/v1/orders/{id}/cancel` — Cancel order
- `GET /api/v1/me/orders` — Order history

#### Orders (Partner/Restaurant)
- `PATCH /partner/orders/{id}/confirm` — Confirm order
- `PATCH /partner/orders/{id}/reject` — Reject order
- `PATCH /partner/orders/{id}/preparing` — Mark preparing
- `PATCH /partner/orders/{id}/ready` — Mark ready

#### Orders (Rider)
- `PATCH /rider/orders/{id}/picked/{restaurantID}` — Mark pickup per restaurant

#### Orders (Admin)
- `PATCH /admin/orders/{id}/force-cancel` — Force cancel with reason

### Key Business Rules

#### Inventory
- Stock cannot go negative (CHECK constraint)
- Reserved qty tracks stock held for pending orders
- Low stock = stock_qty - reserved_qty <= reorder_threshold
- Adjustments create audit log entries

#### Promo Validation
1. Promo must be active (is_active = true)
2. Current time within starts_at/ends_at window
3. Cart total >= min_order_amount
4. total_uses < max_total_uses (if set)
5. User usage count < max_uses_per_user
6. User must be in eligible_user_ids (if restricted)
7. Restaurant/category restrictions apply per applies_to

#### Order Creation (Atomic)
1. Validate cart items exist and restaurants are open
2. Check and reserve stock (SELECT FOR UPDATE)
3. Validate and apply promo code
4. Calculate delivery charge and VAT
5. Insert order + items + pickups + timeline in single transaction
6. Insert outbox event for async processing
7. Increment promo usage if applicable

#### Order Status Transitions
- PENDING → CREATED (payment confirmed for online)
- CREATED → CONFIRMED (restaurant confirms or auto-confirm timeout)
- CREATED → REJECTED (restaurant rejects with reason)
- CONFIRMED → PREPARING (restaurant starts preparing)
- PREPARING → READY (food ready for pickup)
- READY → PICKED (rider picks up)
- PICKED → DELIVERED (rider delivers)
- PENDING/CREATED → CANCELLED (customer cancels)
- Any → CANCELLED (admin force-cancel)

#### Multi-Restaurant Orders
- Each restaurant has independent pickup status
- Parent order READY when ALL pickups READY
- Parent order PICKED when ALL pickups PICKED
- Partner portal shows only own restaurant's items

## Tasks

1. Create SQLC query files (inventory.sql, promos.sql, orders.sql)
2. Generate SQLC code
3. Implement inventory module (handler + service)
4. Implement promo module (handler + service)
5. Implement order module (handler + service)
6. Register routes in server.go
7. Test compilation and validate

## Scenarios

### Scenario: Partner adjusts inventory stock
Given a restaurant manager is authenticated
When they POST /partner/inventory/adjust with product_id, qty_change, reason
Then the inventory stock is adjusted
And an audit log entry is created
And the response shows the updated stock level

### Scenario: Partner creates a promo code
Given a tenant admin is authenticated
When they POST /partner/promos with code, type, discount, conditions
Then the promo is created with the specified rules
And the response returns the promo details

### Scenario: Customer validates promo at checkout
Given an authenticated customer with a cart
When they apply a promo code during charge calculation
Then the system validates all promo conditions
And returns the discount breakdown if valid
Or returns an error describing why the promo is invalid

### Scenario: Customer creates a COD order
Given an authenticated customer with valid cart items
When they POST /api/v1/orders with payment_method=cod
Then stock is reserved for all items
And promo usage is recorded if applicable
And the order is created with status=CREATED
And timeline entry is added
And outbox event is inserted for rider assignment

### Scenario: Restaurant confirms an order
Given a restaurant manager viewing a new order
When they PATCH /partner/orders/{id}/confirm
Then the order pickup status transitions to CONFIRMED
And a timeline entry is recorded
And the rider is notified

### Scenario: Multi-restaurant order pickup
Given an order spanning two restaurants
When restaurant A marks their pickup as READY
Then only restaurant A's pickup status changes to READY
And the parent order remains in CONFIRMED/PREPARING
When restaurant B also marks READY
Then the parent order transitions to READY

### Scenario: Customer cancels a pending order
Given a customer with a PENDING or CREATED order
When they PATCH /api/v1/orders/{id}/cancel
Then the order is cancelled
And reserved stock is released
And refund is triggered if payment was made online
