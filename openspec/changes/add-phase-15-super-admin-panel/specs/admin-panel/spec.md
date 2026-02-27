## ADDED Requirements
### Requirement: Admin Portal Project Foundation
The system SHALL provide an `admin/` Next.js 14 TypeScript project with Tailwind, lint/format tooling, UI primitives, and a shared API client that injects `X-Request-ID` into every request.

#### Scenario: API request tracing
- **WHEN** any admin page performs an API request
- **THEN** the outgoing request includes a unique `X-Request-ID` header
- **AND** API errors are normalized into a standard client error shape

### Requirement: Admin Authentication and Session Enforcement
The system SHALL require authenticated admin sessions with mandatory TOTP two-factor flow and SHALL protect all non-auth routes with server-side session checks.

#### Scenario: First-time 2FA setup
- **WHEN** an admin logs in without TOTP enrollment
- **THEN** the user is redirected to QR setup and must verify TOTP before accessing protected routes

#### Scenario: Session expiry
- **WHEN** an admin session expires or refresh fails
- **THEN** protected route access is denied and the user is redirected to `/auth/login`

### Requirement: Super Admin Operational Pages
The system SHALL provide dashboard, tenant, user, order, finance, issues, and configuration pages with required filtering and sensitive-action reason capture.

#### Scenario: Force order status override
- **WHEN** an admin submits a status override action
- **THEN** a reason is mandatory and included in the mutation payload

#### Scenario: Platform config mutation
- **WHEN** an admin updates feature flags or platform settings
- **THEN** the mutation includes audit metadata to support backend audit-log creation
