## ADDED Requirements

### Requirement: Current User Profile
The system SHALL expose `GET /api/v1/me` returning the authenticated user's full profile and
`PATCH /api/v1/me` allowing the user to update name, email, avatar URL, date of birth, and gender.

#### Scenario: Get profile
- **WHEN** authenticated user calls GET /api/v1/me
- **THEN** the user's profile data is returned (excluding password_hash)

#### Scenario: Update profile
- **WHEN** authenticated user calls PATCH /api/v1/me with partial data
- **THEN** only provided fields are updated and the updated profile is returned

### Requirement: Address Management
The system SHALL expose `GET/POST /api/v1/me/addresses` and `PUT/DELETE /api/v1/me/addresses/:id`
for managing a user's saved delivery addresses. Setting a new default address SHALL unset all other
defaults for the same user atomically.

#### Scenario: List addresses
- **WHEN** authenticated user calls GET /api/v1/me/addresses
- **THEN** all saved addresses for that user are returned ordered by is_default DESC

#### Scenario: Set default
- **WHEN** user creates or updates an address with is_default=true
- **THEN** all other addresses for that user have is_default set to false atomically

### Requirement: Wallet View
The system SHALL expose `GET /api/v1/me/wallet` returning the user's current wallet balance and a
paginated list of wallet transactions.

#### Scenario: Get wallet
- **WHEN** authenticated user calls GET /api/v1/me/wallet
- **THEN** balance and paginated transaction history are returned

### Requirement: Notification View
The system SHALL expose `GET /api/v1/me/notifications` returning paginated in-app notifications and
`PATCH /api/v1/me/notifications/:id/read` to mark a single notification as read.

#### Scenario: Mark notification read
- **WHEN** authenticated user calls PATCH /api/v1/me/notifications/:id/read
- **THEN** the notification status is updated to 'read' and the updated record is returned
