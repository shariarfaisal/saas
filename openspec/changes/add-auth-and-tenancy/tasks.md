## 1. SQLC Queries & Seed (TASK-008, TASK-010)
- [ ] 1.1 Write `internal/db/queries/tenants.sql` with all 7 tenant queries
- [ ] 1.2 Write `internal/db/queries/users.sql` with all user queries + soft-delete
- [ ] 1.3 Write `internal/db/queries/user_addresses.sql` with address queries + SetDefaultAddress
- [ ] 1.4 Write `internal/db/queries/otp_verifications.sql` with OTP queries
- [ ] 1.5 Write `internal/db/queries/refresh_tokens.sql` with token queries
- [ ] 1.6 Write `internal/db/queries/idempotency_keys.sql` with idempotency queries
- [ ] 1.7 Write `internal/db/queries/wallet_transactions.sql` with wallet queries
- [ ] 1.8 Write `internal/db/queries/notifications.sql` with notification queries
- [ ] 1.9 Run `make sqlc` and verify generated code compiles
- [ ] 1.10 Write `backend/scripts/seed_dev.go` dev tenant seed script

## 2. Dependencies (new go modules)
- [ ] 2.1 Add `golang-jwt/jwt/v5` for JWT
- [ ] 2.2 Add `redis/go-redis/v9` for Redis client
- [ ] 2.3 Add `shopspring/decimal` for monetary types (SQLC already configured for it)

## 3. Infrastructure packages
- [ ] 3.1 Create `internal/pkg/respond/respond.go` — JSON response helpers (WriteJSON, WriteError)
- [ ] 3.2 Create `internal/pkg/contextkey/contextkey.go` — typed context keys (TenantCtxKey, UserCtxKey)
- [ ] 3.3 Create `internal/platform/redis/redis.go` — Redis client initialisation
- [ ] 3.4 Create `internal/platform/db/db.go` — pgxpool initialisation helper
- [ ] 3.5 Create `internal/platform/sms/sms.go` — SMS adapter interface + SSL Wireless implementation

## 4. Tenant module (TASK-009)
- [ ] 4.1 Create `internal/modules/tenant/repository.go` — thin repo over SQLC tenant queries
- [ ] 4.2 Create `internal/modules/tenant/middleware.go` — ResolveTenant middleware (subdomain → JWT claim → X-Tenant-ID), Redis cache, 403 guard
- [ ] 4.3 Create `internal/modules/tenant/context.go` — TenantContext type + context helpers

## 5. Auth module (TASK-011, TASK-012, TASK-013)
- [ ] 5.1 Create `internal/modules/auth/token.go` — JWT sign/parse helpers
- [ ] 5.2 Create `internal/modules/auth/service.go` — AuthService (SendOTP, VerifyOTP, Login, Refresh, Logout, ResetRequest, Reset)
- [ ] 5.3 Create `internal/modules/auth/handler.go` — HTTP handlers for auth endpoints
- [ ] 5.4 Create `internal/modules/auth/middleware.go` — AuthRequired, RequireRoles, RequireTenantMatch
- [ ] 5.5 Create `internal/modules/auth/context.go` — UserContext type + context helpers

## 6. User module (TASK-015)
- [ ] 6.1 Create `internal/modules/user/repository.go` — thin repo over SQLC user queries
- [ ] 6.2 Create `internal/modules/user/service.go` — UserService (GetProfile, UpdateProfile, ListAddresses, etc.)
- [ ] 6.3 Create `internal/modules/user/handler.go` — HTTP handlers for /me endpoints

## 7. Idempotency middleware (TASK-014)
- [ ] 7.1 Create `internal/middleware/idempotency.go` — Idempotency-Key middleware

## 8. Wire everything together
- [ ] 8.1 Update `internal/server/server.go` to register all new routes
- [ ] 8.2 Update `cmd/api/main.go` to initialise DB pool and Redis
- [ ] 8.3 Update `internal/config/config.go` if needed

## 9. Tests
- [ ] 9.1 Unit tests for `auth/token.go`
- [ ] 9.2 Unit tests for `tenant/middleware.go`
- [ ] 9.3 Unit tests for `middleware/idempotency.go`
- [ ] 9.4 Verify `go build ./...` clean
- [ ] 9.5 Verify `go test ./...` passes

## 10. OpenSpec update
- [ ] 10.1 Update TASKS.md to mark TASK-008 through TASK-015 as done
