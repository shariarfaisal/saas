## ADDED Requirements

### Requirement: OTP Authentication
The system SHALL support phone-number OTP authentication for customers and riders.
`POST /api/v1/auth/otp/send` SHALL generate a 6-digit OTP, store its bcrypt hash with 10-minute TTL,
send it via the SMS adapter, and enforce a rate limit of 3 requests per phone number per 10 minutes.
`POST /api/v1/auth/otp/verify` SHALL verify the submitted OTP against the stored hash, create the user
record if this is a first-time sign-in, and return JWT access + refresh tokens as httpOnly cookies and
in the JSON response body.

#### Scenario: OTP send success
- **WHEN** a valid Bangladesh phone number is submitted
- **THEN** a 6-digit OTP is sent via SMS and the endpoint returns 200

#### Scenario: OTP rate limited
- **WHEN** 3 OTP requests are made within 10 minutes for the same phone
- **THEN** the endpoint returns 429 with code `RATE_LIMITED`

#### Scenario: OTP verify — new user created
- **WHEN** correct OTP is submitted for a phone not yet registered
- **THEN** a new user record is created and JWT tokens are returned

#### Scenario: OTP verify — existing user
- **WHEN** correct OTP is submitted for an existing user
- **THEN** existing user record is returned and JWT tokens are issued

### Requirement: Email+Password Authentication
The system SHALL support email + password authentication for partner users and platform admins.
`POST /api/v1/auth/login` SHALL validate credentials via bcrypt, enforce role restrictions (only
`tenant_owner`, `tenant_admin`, `restaurant_manager`, `restaurant_staff`, `rider`, `platform_admin`,
`platform_support`, `platform_finance` may use this endpoint), and return JWT tokens.
`POST /api/v1/auth/refresh` SHALL rotate the refresh token.
`POST /api/v1/auth/logout` SHALL add the refresh token to a Redis deny-list.
`POST /api/v1/auth/password/reset-request` and `POST /api/v1/auth/password/reset` SHALL implement an
email-link password reset flow.

#### Scenario: Login success
- **WHEN** valid email + password + allowed role
- **THEN** access and refresh JWT tokens are returned

#### Scenario: Login wrong password
- **WHEN** correct email but wrong password
- **THEN** 401 with code `INVALID_CREDENTIALS`

#### Scenario: Refresh token rotation
- **WHEN** valid refresh token submitted
- **THEN** old token is revoked and a new pair is issued

#### Scenario: Logout
- **WHEN** refresh token submitted to logout
- **THEN** token is added to Redis deny-list; subsequent refresh attempts return 401

### Requirement: JWT Middleware & RBAC
The system SHALL validate Bearer JWT tokens on all protected routes. Expired, tampered, or deny-listed
tokens SHALL return 401. `RequireRoles(…)` SHALL return 403 if the user's role is not in the allowed
list. `RequireTenantMatch()` SHALL return 403 if `user.tenant_id` does not match the resolved tenant
unless the user is a `platform_admin`.

#### Scenario: Missing token
- **WHEN** no Authorization header is present on a protected route
- **THEN** 401 with code `UNAUTHORIZED`

#### Scenario: Role check fail
- **WHEN** authenticated user's role is not in the required list
- **THEN** 403 with code `FORBIDDEN`
