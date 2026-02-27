## ADDED Requirements

### Requirement: Storefront Configuration
The system SHALL expose a public endpoint returning tenant storefront configuration.

#### Scenario: Get config
- **WHEN** `GET /api/v1/storefront/config` is called with a resolved tenant
- **THEN** tenant name, logo, currency, timezone, and delivery config are returned

### Requirement: Storefront Areas
The system SHALL expose a public endpoint listing all active delivery areas for a tenant.

#### Scenario: List areas
- **WHEN** `GET /api/v1/storefront/areas` is called
- **THEN** all active hub coverage areas for the tenant are returned

### Requirement: Storefront Restaurants
The system SHALL expose a public endpoint listing available restaurants with filtering.

#### Scenario: List all
- **WHEN** `GET /api/v1/storefront/restaurants` is called
- **THEN** all available restaurants are returned

#### Scenario: Filter by area
- **WHEN** `GET /api/v1/storefront/restaurants?area=gulshan` is called
- **THEN** only restaurants serving that area are returned

### Requirement: Restaurant Detail with Menu
The system SHALL expose a public endpoint returning full restaurant info with categories and products.

#### Scenario: Get by slug
- **WHEN** `GET /api/v1/restaurants/:slug` is called
- **THEN** full restaurant info, operating hours, categories, and available products are returned

#### Scenario: Not found
- **WHEN** slug does not match any restaurant
- **THEN** 404 is returned

### Requirement: Product Detail
The system SHALL expose a public endpoint returning a single product with its modifier groups.

#### Scenario: Get product
- **WHEN** `GET /api/v1/products/:id` is called
- **THEN** product info with modifier groups and options are returned
