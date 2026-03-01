## ADDED Requirements

### Requirement: Zone-Based Delivery Charge
The system SHALL calculate delivery charges by looking up the customer's `delivery_area` against the `delivery_zones` table for the tenant. The current hardcoded `60 BDT` SHALL be replaced. If no zone match is found, the system SHALL fall back to `platform_configs['default_delivery_charge']`.

#### Scenario: Customer in a mapped delivery zone
- **WHEN** a customer with `delivery_area = 'Gulshan-1'` places an order for a tenant that has a `Gulshan-1 → 40 BDT` delivery zone mapping
- **THEN** the order's `delivery_charge` is set to `40.00`

#### Scenario: Customer in an unmapped area
- **WHEN** a customer's delivery area is not in the tenant's delivery zones table
- **THEN** the order uses the platform default delivery charge from `platform_configs`

#### Scenario: Promo applied to delivery charge
- **WHEN** a promo with `applies_to = 'delivery_charge'` is applied
- **THEN** the delivery charge is reduced by the promo discount (minimum 0 BDT)
- **AND** the `promo_discount_total` on the order reflects the delivery discount

### Requirement: Commission Calculated at Order Creation
At the moment of order creation, the system SHALL look up the commission rate for each restaurant in the cart and populate `order_pickups.commission_rate` and `order_pickups.commission_amount`. The lookup priority is: `restaurants.commission_rate` (if non-null) → `tenants.commission_rate`. Commission MUST NOT default to `0`.

#### Scenario: Commission set on each pickup at creation
- **WHEN** an order is created with items from restaurant A (commission_rate = 12%) and restaurant B (commission_rate = 8%)
- **THEN** pickup A has `commission_rate = 12.00` and `commission_amount = A.items_total × 0.12`
- **AND** pickup B has `commission_rate = 8.00` and `commission_amount = B.items_total × 0.08`

#### Scenario: Restaurant uses tenant default commission
- **WHEN** a restaurant has no custom commission_rate set (NULL)
- **THEN** the tenant's `commission_rate` is used for that restaurant's pickup

### Requirement: Promo Discount Distributed Across Order Items
When a promo code reduces the order total, the system SHALL allocate the promo discount proportionally across the eligible `order_items` based on each item's share of the eligible subtotal. Each `order_items.promo_discount` field SHALL be set to its proportional share. Rounding difference (penny) SHALL be assigned to the last eligible item.

#### Scenario: Promo distributed proportionally
- **WHEN** a 10% promo is applied to an order with two items:
  - Item A: `eligible_subtotal = 200.00`
  - Item B: `eligible_subtotal = 100.00`
  - `total_promo_discount = 30.00`
- **THEN** `item_A.promo_discount = 20.00` (200/300 × 30)
- **AND** `item_B.promo_discount = 10.00` (100/300 × 30)

#### Scenario: Items from different restaurants have isolated promo share
- **WHEN** a promo `applies_to = 'specific_restaurant'` applies to restaurant A only
- **THEN** only items from restaurant A receive a promo_discount allocation
- **AND** items from restaurant B have `promo_discount = 0.00`

### Requirement: Packaging Fee Support
The system SHALL support per-restaurant packaging fees. A `packaging_fee` column SHALL be added to the `orders` table and populated from `restaurants.packaging_fee` (if set). The packaging fee is added to the order total and shown separately in the charge breakdown.

#### Scenario: Order from restaurant with packaging fee
- **WHEN** a customer orders from a restaurant with `packaging_fee = 10.00`
- **THEN** the order's `packaging_fee = 10.00`
- **AND** the `total_amount` includes the packaging fee
- **AND** the charge calculation endpoint returns `packaging_fee` in the breakdown

### Requirement: Wallet Payment Marked as Paid at Creation
When `payment_method = 'wallet'`, the system SHALL deduct the order total from the customer's wallet balance, set `payment_status = 'paid'` at order creation (not `unpaid`), create a `wallet_transactions` debit record, and set the order's initial status to `CREATED` (not `PENDING`).

#### Scenario: Customer pays with wallet
- **WHEN** a customer places an order with `payment_method = wallet` and has sufficient balance
- **THEN** the order is created with `status = created` and `payment_status = paid`
- **AND** a wallet debit transaction is created with `source = order_payment`
- **AND** the customer's `users.balance` is decremented atomically

#### Scenario: Insufficient wallet balance
- **WHEN** a customer places an order with `payment_method = wallet` and has insufficient balance
- **THEN** the order is not created
- **AND** the API returns `422 INSUFFICIENT_WALLET_BALANCE`
