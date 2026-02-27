## ADDED Requirements

### Requirement: Restaurant Management
The system SHALL allow partner users to manage restaurants scoped to their tenant.

#### Scenario: Create restaurant
- **WHEN** a partner sends `POST /partner/restaurants` with required fields
- **THEN** a restaurant is created and returned with 201

#### Scenario: Toggle availability
- **WHEN** a partner sends `PATCH /partner/restaurants/:id/availability`
- **THEN** the restaurant's `is_available` flag is toggled

#### Scenario: Manage operating hours
- **WHEN** a partner sends `PUT /partner/restaurants/:id/hours` with day-of-week entries
- **THEN** operating hours are upserted for the restaurant

### Requirement: Category Management
The system SHALL allow partner users to manage product categories per restaurant.

#### Scenario: Create category
- **WHEN** a partner sends `POST /partner/restaurants/:id/categories`
- **THEN** a category is created

#### Scenario: Reorder categories
- **WHEN** a partner sends `PATCH /partner/restaurants/:id/categories/reorder` with ordered IDs
- **THEN** sort orders are updated accordingly

### Requirement: Product Management
The system SHALL allow partner users to manage products with pricing, variants, and add-ons.

#### Scenario: Create product
- **WHEN** a partner sends `POST /partner/restaurants/:id/products`
- **THEN** a product is created

#### Scenario: Create product with modifiers
- **WHEN** a partner sends product create/update with `modifier_groups` array
- **THEN** modifier groups and their options are created and linked to the product

#### Scenario: Update availability
- **WHEN** a partner sends `PATCH /partner/products/:id/availability`
- **THEN** product availability is updated

### Requirement: Product Discounts
The system SHALL allow partner users to create time-limited discounts on products.

#### Scenario: Create discount
- **WHEN** a partner sends `POST /partner/products/:id/discount`
- **THEN** an active discount is created for the product

#### Scenario: Deactivate discount
- **WHEN** a partner sends `DELETE /partner/products/:id/discount`
- **THEN** any active discount is deactivated

#### Scenario: Expire discounts
- **WHEN** `ExpireDiscounts` is called
- **THEN** all past-end-date discounts are deactivated

### Requirement: Menu Bulk Upload
The system SHALL allow partner users to upload a CSV file to bulk-create categories and products.

#### Scenario: Valid CSV
- **WHEN** a partner uploads a valid CSV with headers: `category_name,name,description,base_price,availability`
- **THEN** categories and products are created; count of created rows returned

### Requirement: Menu Duplication
The system SHALL allow copying all categories and products from one restaurant to another within the same tenant.

#### Scenario: Duplicate menu
- **WHEN** a partner sends `POST /partner/restaurants/:id/menu/duplicate` with `target_restaurant_id`
- **THEN** all categories and products are copied to the target restaurant
