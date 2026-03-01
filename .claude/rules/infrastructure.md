# Infrastructure & Deployment

## Local Development

```bash
docker-compose up       # PostgreSQL 16 (port 5432) + Redis 7 (port 6379)
cd backend && make dev  # Go API on :8080
cd website && npm run dev   # Customer site on :3000
cd partner && npm run dev   # Partner portal on :3001
cd admin && npm run dev     # Admin panel on :3002
```

## Docker (Backend)

Multi-stage Alpine build:
```dockerfile
# Builder: golang:1.22-alpine
# Runtime: alpine:3.19 with ca-certificates + tzdata
# Binary: CGO_ENABLED=0 -ldflags="-s -w"
```

**Rules:**
- Keep images minimal — Alpine only
- No dev dependencies in runtime image
- Always include `ca-certificates` (for HTTPS) and `tzdata` (for time zones)
- Health check endpoint: `/healthz`

## Environments

| Env | Domain | Purpose |
|-----|--------|---------|
| local | localhost:* | Developer machine |
| development | *.dev.platform.com | Shared dev/integration |
| staging | *.staging.platform.com | Pre-prod testing |
| production | *.platform.com | Live system |

## Production Architecture

```
Cloudflare (CDN + DDoS + WAF)
  → Nginx (TLS termination, rate limiting)
    → Go API cluster (multiple nodes)
      → PostgreSQL (primary + read replica)
      → Redis (Sentinel/Cluster)
```

## Monitoring Stack

- **Metrics:** Prometheus + Grafana dashboards
- **Logs:** Zerolog (structured JSON) → Loki
- **Errors:** Sentry (SENTRY_DSN env var)
- **Traces:** OpenTelemetry
- **Uptime:** UptimeRobot (external)

## Key Environment Variables

```
# Server
PORT, ENVIRONMENT, ALLOWED_ORIGINS

# Database
DATABASE_URL, DB_MAX_OPEN_CONNS, DB_MAX_IDLE_CONNS

# Redis
REDIS_URL

# Auth
JWT_ACCESS_SECRET, JWT_REFRESH_SECRET, JWT_ACCESS_EXPIRY, JWT_REFRESH_EXPIRY

# External
SENTRY_DSN, S3_BUCKET, BKASH_*, FIREBASE_*
```

Never hardcode secrets. All config via environment variables (12-factor app).

## SLOs

- 99.9% API availability
- 99.95% order placement success
- RPO: 15 minutes, RTO: 60 minutes

---
Path scope: Docker, CI/CD, deployment
