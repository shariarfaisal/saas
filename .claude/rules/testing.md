# Testing Rules

## Backend (Go)

**Test location:** `*_test.go` alongside source files
**Run:** `make test` (unit) or `make test-integration` (with DB)
**Flags:** `-race -coverprofile=coverage.out`

### What to test:
- Middleware (auth, tenant resolver, rate limiting)
- Service methods (business logic with mocked DB)
- Utility packages (pagination, validation, error types, time)
- Config loading

### Test pattern:
```go
func TestSendOTP_InvalidPhone(t *testing.T) {
    svc := NewService(mockQueries, mockRedis, mockSMS, tokenCfg)
    err := svc.SendOTP(ctx, &tenantID, "", "login")
    if err == nil {
        t.Fatal("expected error for empty phone")
    }
}
```

### Multi-tenant test rule:
Every test that touches data must verify tenant isolation — test that tenant A cannot access tenant B's resources.

## Frontend (TypeScript)

No test framework is configured yet. When adding tests:
- Use Vitest (compatible with Next.js + TypeScript)
- Cypress for E2E testing
- Test critical flows: auth, order creation, payment

## CI (GitHub Actions)

Backend CI runs on ubuntu-latest with:
- Go 1.22
- PostgreSQL 16 service container
- Migrations before tests
- Race detection enabled

---
Path scope: All test code
