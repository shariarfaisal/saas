# Project Context

## Purpose
Multi-tenant food-commerce SaaS platform. Vendors (restaurant groups, cloud kitchens) get their own branded online ordering storefront, partner dashboard, and rider management — all on shared infrastructure. Platform earns commission on every order.

**Full documentation:** `docs/requirements/` — this is the authoritative source of truth.

## Tech Stack
- **Backend:** Go 1.22+ (modular monolith), Chi/Echo router, PostgreSQL 16+, SQLC, Redis, asynq
- **Frontend:** Next.js 14+ (App Router), TypeScript, Tailwind CSS, shadcn/ui, TanStack Query
- **Infra:** Docker, Nginx, GitHub Actions CI/CD, Cloudflare CDN

## Project Structure
```
saas/
├── backend/          # Go API server
├── website/          # Customer-facing Next.js app (SSR/SSG)
├── partner/          # Vendor portal Next.js (SPA)
├── admin/            # Super-admin panel Next.js (SPA)
└── docs/requirements/ # Full documentation (read before building)
```

## Project Conventions

### Code Style
- Go: `gofmt` + `golangci-lint`; handler → service → repository layers; no direct DB access in handlers
- TypeScript: strict mode, ESLint + Prettier; named exports; barrel files for modules
- SQL: snake_case table/column names; SQLC for all queries (no raw SQL in Go files outside `queries/`)
- UUIDs as primary keys everywhere

### Architecture Patterns
- Modular monolith — one Go binary, clean module boundaries under `internal/modules/`
- Row-level multi-tenancy — every tenant-scoped table has `tenant_id UUID NOT NULL`
- Repository pattern — all DB access through SQLC-generated typed functions
- No ORM — SQLC only

### Testing Strategy
- Unit tests for service layer (mock repositories)
- Integration tests for API endpoints (test DB)
- Table-driven tests in Go

### Git Workflow
- `main` — production
- `develop` — integration
- Feature branches: `feature/description`
- Commit format: `type(scope): description` (conventional commits)

## Domain Context
- **Tenant** = a vendor (e.g., "Kacchi Bhai" restaurant brand)
- **Restaurant** = a physical or virtual outlet owned by a tenant
- **Hub** = delivery dispatch zone; riders are assigned to hubs
- **Order** has multiple **pickups** (one per restaurant in a multi-restaurant order)
- Order status flow: `pending → created → confirmed → preparing → ready → picked → delivered`
- **Commission** = platform takes % of food subtotal from each order
- **Invoice** = daily financial settlement for each restaurant

## Important Constraints
- No microservices — keep it as a modular monolith
- Multi-tenant by design — tenant_id must be on every scoped table
- Bangladesh market first (BDT currency, bKash + AamarPay payments, Barikoi maps)
- Global-quality architecture (built to scale internationally)
- SQLC only for database queries — no GORM or similar ORMs

## External Dependencies
- **bKash** — mobile banking payment gateway (Bangladesh)
- **AamarPay** — card payment gateway (Bangladesh)
- **Firebase FCM** — push notifications
- **Barikoi** — Bangladesh map/geocoding API
- **Google Maps API** — rider tracking maps (fallback)
- **SSL Wireless** — SMS OTP (Bangladesh)
- **AWS S3 / Cloudflare R2** — file/media storage
- **Redis** — caching, sessions, job queues, pub/sub for SSE
