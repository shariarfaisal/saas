# Security & Permissions

## Authentication Flow

- OTP-based auth (phone) + password login (partner/admin)
- JWT: access token (15min) + refresh token (7d, httpOnly cookie)
- Token refresh: single-flight pattern (one refresh promise for concurrent 401s)
- Hash refresh tokens with SHA256 before storing in DB
- Passwords: bcrypt hashing

## Authorization Layers (all must pass)

1. **Tenant resolution** — Middleware extracts tenant from subdomain (customer) or JWT claim (partner/admin) or `X-Tenant-ID` header (API)
2. **Authentication** — `authMiddleware.Authenticate` verifies JWT and attaches user to context
3. **Role check** — `RequireRoles(roles ...sqlc.UserRole)` middleware on route groups
4. **Resource ownership** — Service layer verifies resource belongs to user's tenant

```go
// CORRECT — full chain
r.Route("/orders", func(r chi.Router) {
    r.Use(authMiddleware.Authenticate)
    r.Use(auth.RequireRoles(sqlc.UserRoleOwner, sqlc.UserRoleManager))
    r.Post("", orderHandler.CreateOrder)
})

// In handler — always extract and verify tenant
t := tenant.FromContext(r.Context())
if t == nil {
    respond.Error(w, apperror.NotFound("tenant"))
    return
}
```

## User Roles (from doc 06)

| Role | Scope |
|------|-------|
| Platform Admin | Cross-tenant, super-admin panel |
| Restaurant Owner | Tenant-wide, partner portal |
| Manager | Restaurant-level, partner portal |
| Staff | Limited ops, partner portal |
| Customer | Own data only, website |
| Rider | Assigned deliveries, rider app |

## Security Checklist for Every Endpoint

- [ ] Tenant middleware applied to route group
- [ ] Auth middleware applied (unless public endpoint)
- [ ] Role middleware matches required permission level
- [ ] Handler extracts `tenant.FromContext()` and checks nil
- [ ] Handler extracts `auth.UserFromContext()` and checks nil
- [ ] Service passes `tenant_id` to all DB queries
- [ ] Input validated with `go-playground/validator` or Zod (frontend)
- [ ] Rate limiting on sensitive endpoints (OTP, login, password reset)

## Frontend Security

- No sensitive data in localStorage — tokens in httpOnly cookies only
- `withCredentials: true` on all API calls
- `secure: true` on cookies in production
- Zod schema validation on all form inputs before submission
- Next.js middleware redirects unauthenticated users to `/auth/login`

## Common Mistakes

- Forgetting `tenant_id` in a query (data leak)
- Checking role but not tenant (cross-tenant privilege escalation)
- Returning internal error details to client (use `apperror.Internal` which hides internals)
- Missing rate limit on OTP endpoint (abuse vector)
- Storing tokens in localStorage (XSS vulnerability)

---
Path scope: All API endpoints, middleware, auth code
