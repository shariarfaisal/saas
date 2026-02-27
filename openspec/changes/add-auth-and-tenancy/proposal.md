# Change: Multi-Tenancy, Authentication & User Profile (TASK-008 through TASK-015)

## Why
The backend foundation (schema, migrations, shared packages) is in place. The platform now needs tenant
resolution, authentication (OTP + email/password), JWT middleware, RBAC guards, idempotency, and the
current-user profile API to become a functional multi-tenant SaaS.

## What Changes
- **TASK-008** — SQLC queries for `tenants` table (`CreateTenant`, `GetTenantBySlug`, `GetTenantByID`,
  `GetTenantByDomain`, `UpdateTenant`, `UpdateTenantStatus`, `ListTenants`) + dev seed script.
- **TASK-009** — Tenant resolver middleware: subdomain → JWT claim → `X-Tenant-ID` header; Redis 60 s cache;
  `TenantContext` injected into request context; 403 for suspended/cancelled tenants.
- **TASK-010** — SQLC queries for `users`, `user_addresses`, `otp_verifications` tables including
  `SetDefaultAddress` transactional logic.
- **TASK-011** — `POST /api/v1/auth/otp/send` + `POST /api/v1/auth/otp/verify`; 6-digit OTP, hashed
  storage, SMS adapter, rate-limit 3/phone/10 min; JWT access + refresh tokens in httpOnly cookies + body.
- **TASK-012** — `POST /api/v1/auth/login`, `POST /api/v1/auth/refresh`, `POST /api/v1/auth/logout`,
  `POST /api/v1/auth/password/reset-request`, `POST /api/v1/auth/password/reset`; bcrypt; Redis deny-list.
- **TASK-013** — `AuthRequired` JWT validation middleware; `RequireRoles(…)` guard; `RequireTenantMatch()`
  guard; `UserContext` attached to request context.
- **TASK-014** — Idempotency key middleware for POST/PATCH: read header, check DB, replay or store response
  snapshot; applied to `/orders` and payment endpoints.
- **TASK-015** — `GET/PATCH /api/v1/me`, `GET/POST/PUT/DELETE /api/v1/me/addresses`,
  `GET /api/v1/me/wallet`, `GET /api/v1/me/notifications`, `PATCH /api/v1/me/notifications/:id/read`.

## Impact
- Affected specs: auth, tenant-resolution, user-profile
- Affected code: `backend/internal/modules/tenant/`, `backend/internal/modules/auth/`,
  `backend/internal/modules/user/`, `backend/internal/middleware/`, `backend/internal/db/queries/`,
  `backend/internal/db/sqlc/`, `backend/cmd/api/main.go`, `backend/internal/server/server.go`,
  `backend/go.mod`
- New dependencies: `golang-jwt/jwt/v5`, `redis/go-redis/v9`, `shopspring/decimal`
- No **BREAKING** changes to existing public interfaces; migrations already applied
