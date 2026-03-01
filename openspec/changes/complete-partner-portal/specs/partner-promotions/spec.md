## ADDED Requirements

### Requirement: Promo List from API
The promotions list page SHALL fetch promos from `GET /partner/promos?restaurant_id=:id` and display them with type, code, usage count, and status badge (active / expired / inactive). The page currently uses `useState(mockPromos)`.

#### Scenario: Promo list loads from API
- **WHEN** a partner navigates to the promotions page
- **THEN** promos are fetched and listed with correct type labels and status badges

### Requirement: Promo Create via API
The create promotion page SHALL submit the validated form to `POST /partner/promos` and redirect to the promo list on success. All form fields (code, type, amount, cap, apply_on, min_order, max_uses, per_user_limit, start/end dates, description, cashback amount) SHALL be included in the request payload.

#### Scenario: Partner creates a percentage discount promo
- **WHEN** the partner fills the form with a valid percentage promo and submits
- **THEN** `POST /partner/promos` is called with the correct payload
- **AND** the partner is redirected to the promo list showing the new promo

#### Scenario: Duplicate promo code
- **WHEN** the partner submits a code that already exists for the tenant
- **THEN** the API returns a 409 conflict
- **AND** an inline error is shown on the Code field

### Requirement: Promo Edit and Deactivate via API
The promo edit page SHALL load promo data from `GET /partner/promos/:id`, submit changes to `PUT /partner/promos/:id`, and support deactivation via `PATCH /partner/promos/:id/deactivate`.

#### Scenario: Partner edits a promo
- **WHEN** the partner changes the discount amount and saves
- **THEN** `PUT /partner/promos/:id` is called with the updated payload
- **AND** a success toast confirms the change

#### Scenario: Partner deactivates a promo
- **WHEN** the partner clicks "Deactivate" on an active promo
- **THEN** `PATCH /partner/promos/:id/deactivate` is called
- **AND** the promo's status badge changes to Inactive

### Requirement: Promo Performance Stats
The promo edit/detail page SHALL display performance statistics fetched from `GET /partner/promos/:id/stats`: total usage count, total discount amount given, and unique users who used the promo. This endpoint SHALL be added to the backend if not yet present.

#### Scenario: Promo stats display real data
- **WHEN** a partner opens a promo for editing
- **THEN** the performance stats section shows real usage figures from the API

### Requirement: Backend — Promo Stats Endpoint
The backend SHALL expose `GET /partner/promos/:id/stats` returning `{ usage_count, total_discount_given, unique_users }` aggregated from the promo usage records for the given promo ID, scoped to the authenticated tenant.

#### Scenario: Stats endpoint returns correct aggregates
- **WHEN** a tenant admin calls `GET /partner/promos/:id/stats`
- **THEN** the response contains usage_count, total_discount_given, and unique_users for that promo
