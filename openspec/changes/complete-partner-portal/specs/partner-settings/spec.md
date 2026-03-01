## ADDED Requirements

### Requirement: Backend — Notification Preferences API
The backend SHALL expose:
- `GET /partner/settings/notifications` — return the authenticated user's notification preference map (8 event types × push/email boolean toggles) scoped to the tenant
- `PUT /partner/settings/notifications` — accept and persist the full preference map

Preferences SHALL be stored as a JSONB column `notification_preferences` on the `tenant_users` join table (or equivalent). A missing column SHALL be added via migration.

#### Scenario: Partner retrieves notification preferences
- **WHEN** an authenticated partner calls `GET /partner/settings/notifications`
- **THEN** the response contains a JSON map of event types to channel preferences with current values

#### Scenario: Partner updates notification preferences
- **WHEN** the partner calls `PUT /partner/settings/notifications` with an updated preference map
- **THEN** the preferences are persisted and the response confirms the saved values

### Requirement: Frontend — Settings Notification Preferences with API
The settings page notification preferences section SHALL fetch current preferences from `GET /partner/settings/notifications` on load and submit changes via `PUT /partner/settings/notifications`. The 8 event-type toggle rows currently use local React state with no persistence.

#### Scenario: Settings page loads real preferences
- **WHEN** a partner navigates to Settings
- **THEN** the notification toggle states are populated from the API response

#### Scenario: Partner saves notification preference changes
- **WHEN** the partner toggles some preferences and clicks Save
- **THEN** `PUT /partner/settings/notifications` is called with the full preference map
- **AND** a success toast confirms the changes are saved

### Requirement: Vendor Profile Update via API
The vendor profile section in Settings SHALL fetch profile data from `GET /partner/me/profile` and submit changes via `PUT /partner/me/profile`. Fields include business name, contact email, phone, registration number, and address. The section currently uses local state with no API calls.

#### Scenario: Settings page loads vendor profile
- **WHEN** a partner navigates to Settings
- **THEN** the profile fields are pre-filled with data from the API

#### Scenario: Partner saves profile changes
- **WHEN** the partner updates a field and clicks Save Profile
- **THEN** `PUT /partner/me/profile` is called with the updated values
- **AND** the auth store is updated with the new business name if changed
