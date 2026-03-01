## ADDED Requirements

### Requirement: Unique Modifier Option Names per Group
A `UNIQUE(modifier_group_id, name)` database constraint SHALL be added to `product_modifier_options`. The create and update APIs for modifier options SHALL return `422 DUPLICATE_MODIFIER_OPTION_NAME` before the constraint is hit, with a user-friendly message.

#### Scenario: Duplicate option name rejected at API level
- **WHEN** a restaurant tries to add a modifier option `"No Onion"` to a group that already has `"No Onion"`
- **THEN** the API returns `422 DUPLICATE_MODIFIER_OPTION_NAME` before writing to the database

### Requirement: Cross-Restaurant Category Assignment Prevented
The product create/update service SHALL validate that the `category_id` belongs to the same `restaurant_id` as the product being created/updated. Mismatches SHALL return `422 CATEGORY_RESTAURANT_MISMATCH`.

#### Scenario: Product assigned to wrong restaurant's category
- **WHEN** a request creates a product for restaurant A with a `category_id` that belongs to restaurant B
- **THEN** the API returns `422 CATEGORY_RESTAURANT_MISMATCH`
- **AND** the product is not created

### Requirement: Denormalized Stats Updated via Outbox Events
`products.order_count` SHALL be incremented when an `order.delivered` event is processed. `products.rating_avg` and `products.rating_count` SHALL be updated when a `review.created` event is processed. These updates occur asynchronously via the outbox processor, not inline with the order creation.

#### Scenario: Product order count incremented on delivery
- **WHEN** an order is delivered
- **AND** the outbox processor handles the `order.delivered` event
- **THEN** each product in the order has its `order_count` incremented by the ordered quantity

#### Scenario: Rating average updated on new review
- **WHEN** a customer submits a review with `rating = 5`
- **AND** the outbox processor handles the `review.created` event
- **THEN** the product's `rating_avg` is recalculated as `(current_avg * current_count + 5) / (current_count + 1)`
- **AND** `rating_count` is incremented by 1

### Requirement: Inventory Item Created Timestamp
The `inventory_items` table SHALL have a `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()` column. All new inventory item inserts SHALL populate this field.

#### Scenario: Inventory item creation recorded
- **WHEN** a new inventory item is created
- **THEN** `created_at` is set to the current server timestamp
