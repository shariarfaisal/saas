## ADDED Requirements

### Requirement: Backend — Team Management API
The backend SHALL expose the following partner-scoped team management endpoints, restricted to `tenant_owner` and `tenant_admin` roles:
- `GET /partner/team` — list all team members (users with roles scoped to the authenticated tenant)
- `POST /partner/team/invite` — create an invitation record and send an email with an accept link
- `PUT /partner/team/:userId/role` — change an existing member's role
- `DELETE /partner/team/:userId` — remove a member from the tenant (cannot remove self or last owner)

These endpoints SHALL be implemented as a new `team` module under `backend/internal/modules/team/` following the handler → service → repository pattern.

#### Scenario: Tenant admin lists team members
- **WHEN** an authenticated tenant_admin calls `GET /partner/team`
- **THEN** the response lists all users associated with the tenant including their role and status (active / pending)

#### Scenario: Tenant admin invites a new member
- **WHEN** the admin calls `POST /partner/team/invite` with email and role
- **THEN** an invitation record is created in the database
- **AND** an invitation email is sent via the email adapter with a token-based accept link
- **AND** the response returns the invitation details with status `pending`

#### Scenario: Invited user accepts invitation
- **WHEN** the invitee clicks the link and sets their password via `POST /auth/invite/accept`
- **THEN** a user account is created (or linked) with the specified tenant and role
- **AND** the invitation status is updated to `accepted`

#### Scenario: Admin changes a member's role
- **WHEN** the admin calls `PUT /partner/team/:userId/role` with a new role value
- **THEN** the user's role within the tenant is updated
- **AND** the response confirms the new role

#### Scenario: Admin removes a team member
- **WHEN** the admin calls `DELETE /partner/team/:userId`
- **THEN** the user is disassociated from the tenant
- **AND** 400 is returned if the target is the sole owner

### Requirement: Frontend — Team Management with Real API
The team management page (`/team`) SHALL fetch members from `GET /partner/team`, send invitations via `POST /partner/team/invite`, change roles via `PUT /partner/team/:userId/role`, and remove members via `DELETE /partner/team/:userId`. The page currently uses `useState(mockTeam)` with no backend calls.

#### Scenario: Team page shows real member list
- **WHEN** a partner navigates to the team page
- **THEN** all team members with their roles and statuses are displayed from the API

#### Scenario: Partner invites a team member
- **WHEN** the partner enters an email, selects a role, and submits the invite modal
- **THEN** `POST /partner/team/invite` is called
- **AND** the new member appears in the list with "Pending" status badge

#### Scenario: Partner removes a team member
- **WHEN** the partner clicks "Remove" on a non-owner member and confirms
- **THEN** `DELETE /partner/team/:userId` is called
- **AND** the member is removed from the list
