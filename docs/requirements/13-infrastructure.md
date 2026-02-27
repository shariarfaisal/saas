# 13 — Infrastructure & DevOps

## 13.1 Environments

| Environment | Purpose | URL Pattern |
|-------------|---------|-------------|
| `local` | Developer machine | `localhost:*` |
| `development` | Shared dev/integration | `*.dev.platform.com` |
| `staging` | Pre-production testing | `*.staging.platform.com` |
| `production` | Live system | `*.platform.com` |

---

## 13.2 Docker Setup (Local Dev)

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16-alpine
    ports: ["5432:5432"]
    environment:
      POSTGRES_DB: platform
      POSTGRES_USER: platform
      POSTGRES_PASSWORD: secret
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
    command: redis-server --appendonly yes

  api:
    build: ./backend
    ports: ["8080:8080"]
    depends_on: [postgres, redis]
    environment:
      DATABASE_URL: postgres://platform:secret@postgres:5432/platform
      REDIS_URL: redis://redis:6379
    volumes:
      - ./backend:/app

  website:
    build: ./website
    ports: ["3000:3000"]
    environment:
      NEXT_PUBLIC_API_URL: http://localhost:8080/api/v1

  partner:
    build: ./partner
    ports: ["3001:3001"]

  admin:
    build: ./admin
    ports: ["3002:3002"]
```

---

## 13.3 Backend Go Configuration

```go
// internal/config/config.go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    JWT      JWTConfig
    Storage  StorageConfig
    Services ExternalServicesConfig
}

type ServerConfig struct {
    Port        int
    Environment string  // local, development, staging, production
    AllowedOrigins []string
}

type DatabaseConfig struct {
    URL             string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
}

type JWTConfig struct {
    AccessTokenSecret  string
    RefreshTokenSecret string
    AccessTokenExpiry  time.Duration  // 15m
    RefreshTokenExpiry time.Duration  // 7d
}

type ExternalServicesConfig struct {
    BkashAppKey      string
    BkashAppSecret   string
    BkashBaseURL     string
    AamarPayStoreID  string
    AamarPayAPIKey   string
    AamarPayBaseURL  string
    FCMServerKey     string
    SMSAPIKey        string
    SMSBaseURL       string
    BarikoiAPIKey    string
    S3Bucket         string
    S3Region         string
    S3Endpoint       string
    AWSAccessKey     string
    AWSSecretKey     string
    SentryDSN        string
}
```

---

## 13.4 Makefile Commands

```makefile
# backend/Makefile

.PHONY: dev build test migrate sqlc

dev:
	go run ./cmd/api/main.go

build:
	go build -o bin/api ./cmd/api/main.go

test:
	go test ./... -v -cover

test-integration:
	go test ./... -v -tags=integration

# Database migrations
migrate-up:
	migrate -path internal/db/migrations -database ${DATABASE_URL} up

migrate-down:
	migrate -path internal/db/migrations -database ${DATABASE_URL} down 1

migrate-create:
	migrate create -ext sql -dir internal/db/migrations -seq $(name)

# SQLC code generation
sqlc:
	sqlc generate

# Linting
lint:
	golangci-lint run ./...

# Full local setup
setup: migrate-up sqlc
```

---

## 13.5 CI/CD Pipeline (GitHub Actions)

```yaml
# .github/workflows/backend.yml
name: Backend CI

on:
  push:
    branches: [main, develop]
    paths: ['backend/**']
  pull_request:
    paths: ['backend/**']

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_DB: platform_test
          POSTGRES_PASSWORD: secret
        options: --health-cmd pg_isready
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: cd backend && go mod download
      - run: cd backend && make migrate-up
      - run: cd backend && make test

  build:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - run: docker build -t platform-api:${{ github.sha }} ./backend
      - run: docker push $REGISTRY/platform-api:${{ github.sha }}
```

---

## 13.6 Production Architecture

```
Internet
    │
    ▼
Cloudflare (CDN + DDoS protection + WAF)
    │
    ├──► Static assets (Next.js build outputs) → Cloudflare Pages
    │
    └──► API traffic
             │
             ▼
         Load Balancer (Nginx / AWS ALB)
             │
        ┌────┴────┐
        │         │
       API       API
      node 1    node 2    (horizontal scaling)
        │         │
        └────┬────┘
             │
         ┌───┴───┐
         │       │
      Primary  Redis
      Postgres   │
         │    (Sentinel
      Read     or Cluster)
      Replica
```

---

## 13.7 Environment Variables (.env)

```bash
# Server
PORT=8080
ENVIRONMENT=production
ALLOWED_ORIGINS=https://*.platform.com

# Database
DATABASE_URL=postgres://user:pass@host:5432/platform?sslmode=require
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5

# Redis
REDIS_URL=redis://:password@redis:6379/0

# JWT
JWT_ACCESS_SECRET=<256-bit-secret>
JWT_REFRESH_SECRET=<256-bit-secret>
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# Storage
S3_BUCKET=platform-media
S3_REGION=ap-southeast-1
S3_ENDPOINT=https://s3.amazonaws.com
AWS_ACCESS_KEY_ID=...
AWS_SECRET_ACCESS_KEY=...

# Payment — bKash
BKASH_APP_KEY=...
BKASH_APP_SECRET=...
BKASH_USERNAME=...
BKASH_PASSWORD=...
BKASH_BASE_URL=https://tokenized.pay.bka.sh/v1.2.0-beta

# Payment — AamarPay
AAMARPAY_STORE_ID=...
AAMARPAY_API_KEY=...
AAMARPAY_BASE_URL=https://secure.aamarpay.com

# Firebase (Push)
FIREBASE_PROJECT_ID=...
FIREBASE_PRIVATE_KEY=...
FIREBASE_CLIENT_EMAIL=...

# SMS
SMS_API_KEY=...
SMS_FROM=PlatformName

# Maps
BARIKOI_API_KEY=...
GOOGLE_MAPS_API_KEY=...

# Monitoring
SENTRY_DSN=https://...
```

---

## 13.8 Monitoring & Alerting

| Tool | Purpose |
|------|---------|
| Prometheus | Metrics collection (request rate, latency, error rate, queue depth) |
| Grafana | Dashboards and visualisation |
| Loki | Log aggregation |
| Sentry | Error tracking and alerting |
| UptimeRobot | External uptime monitoring |

**Key alerts:**
- API p95 latency > 1s (warning), > 3s (critical)
- Error rate > 1% (warning), > 5% (critical)
- Postgres connection pool > 80% utilisation
- Redis memory > 80%
- Disk usage > 80%
- SSL certificate expiry < 14 days
- Order queue depth > 100 (unprocessed jobs)

---

## 13.9 Security Checklist

- [ ] All secrets in environment variables, never in code
- [ ] HTTPS everywhere (enforce via Cloudflare)
- [ ] httpOnly, Secure, SameSite=Strict cookies for auth tokens
- [ ] Rate limiting on all auth and order endpoints
- [ ] SQL injection protection via SQLC (parameterised queries only)
- [ ] Input validation via go-playground/validator
- [ ] CORS restricted to known origins
- [ ] JWT signed with RS256 or HS256 with strong secrets
- [ ] Tenant isolation enforced at repository layer (not just handler)
- [ ] File uploads: type validation, size limits, virus scan (Phase 2)
- [ ] Admin 2FA mandatory
- [ ] Audit log for all sensitive mutations
- [ ] Dependency scanning (Dependabot / Snyk)
- [ ] Regular PostgreSQL backups (daily automated, 30-day retention)
- [ ] Principle of least privilege for DB user (no DDL in app user)

---

## 13.10 Reliability Objectives (SLO / SLA)

### Platform SLOs
- API availability: **99.9% monthly**
- Order placement success rate (non-user-error): **99.95%**
- Payment callback processing latency p95: **< 10s**
- Partner live order board freshness (SSE lag): **< 3s p95**

### Internal SLIs to track
- Request success ratio by endpoint group
- Queue processing delay (event age)
- DB transaction contention / lock waits
- Redis pub/sub delivery lag

### SLA communication
- Public status page for tenants
- Incident communication templates for P1/P2 outages

---

## 13.11 Backup, Disaster Recovery & Continuity

### Backup policy
- PostgreSQL full backup: daily
- WAL archiving / point-in-time recovery enabled
- Redis snapshot + AOF backup every 6 hours
- Object storage versioning enabled for media

### Recovery objectives
- **RPO**: 15 minutes (max data loss)
- **RTO**: 60 minutes (service restoration)

### DR posture
- Warm standby database in secondary region
- IaC templates to recreate API + worker stack quickly
- Quarterly restore drills (must be documented and signed off)

---

## 13.12 Secrets, Access & Environment Hardening

- Secrets manager required in non-local environments (AWS Secrets Manager / Vault).
- No long-lived static cloud credentials in CI; use OIDC short-lived tokens.
- Production access must require SSO + MFA and be role-scoped.
- Break-glass credentials must be rotated after every emergency use.
- Separate service accounts:
  - API read/write account
  - migration account
  - analytics read account

---

## 13.13 Release Governance

### Promotion model
`develop` → `staging` → `production` with mandatory quality gates.

### Production deployment gates
- All required tests pass
- DB migration backward compatibility check passes
- Security scan passes (no critical unresolved findings)
- SLO/error-budget status healthy
- Rollback plan attached to release

### Post-deploy verification
- Smoke tests: auth, checkout, payment callback, order board, rider updates
- Canary monitoring window before full rollout
- Automatic rollback trigger thresholds predefined
