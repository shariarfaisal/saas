## ADDED Requirements

### Requirement: Hub Management
The system SHALL allow partner users to create, list, update, and delete dispatch hubs scoped to their tenant.

#### Scenario: Create hub
- **WHEN** a partner sends `POST /partner/hubs` with `name` and `city`
- **THEN** a hub is created for the resolved tenant and returned with 201

#### Scenario: List hubs
- **WHEN** a partner sends `GET /partner/hubs`
- **THEN** all hubs for the resolved tenant are returned

#### Scenario: Update hub
- **WHEN** a partner sends `PUT /partner/hubs/:id` with updated fields
- **THEN** the hub is updated and returned

#### Scenario: Delete hub
- **WHEN** a partner sends `DELETE /partner/hubs/:id`
- **THEN** the hub is removed

### Requirement: Hub Coverage Area Management
The system SHALL allow partner users to manage coverage areas (delivery zones) for each hub.

#### Scenario: Create coverage area
- **WHEN** a partner sends `POST /partner/hubs/:id/areas` with `name` and `delivery_charge`
- **THEN** a coverage area is created under the hub

#### Scenario: List coverage areas
- **WHEN** a partner sends `GET /partner/hubs/:id/areas`
- **THEN** all coverage areas for the hub are returned

#### Scenario: Update coverage area
- **WHEN** a partner sends `PUT /partner/hubs/:id/areas/:area_id`
- **THEN** the coverage area is updated

#### Scenario: Delete coverage area
- **WHEN** a partner sends `DELETE /partner/hubs/:id/areas/:area_id`
- **THEN** the area is removed

### Requirement: Delivery Charge Calculation
The system SHALL calculate delivery charges based on zone-based model.

#### Scenario: Known area
- **WHEN** `POST /api/v1/orders/charges/calculate` is called with a valid `hub_id` and `area_slug`
- **THEN** the delivery charge and estimated delivery time are returned

#### Scenario: Unknown area
- **WHEN** `area_slug` does not match any active coverage area
- **THEN** a 404 error is returned
