# Munchies SaaS - Claude Code Configuration

Multi-tenant food-commerce SaaS platform. Vendors launch branded online ordering businesses on shared infrastructure. Commission-based revenue model. Target: Bangladesh first, then South Asia.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **Backend** | Go 1.22+, Chi router, PostgreSQL 16+ (SQLC), Redis 7+, JWT auth |
| **Website** | Next.js 16, React 19, TypeScript, Tailwind CSS 4 |
| **Partner Portal** | Next.js 15, React 18, TypeScript, Tailwind CSS 3, shadcn/ui |
| **Admin Panel** | Next.js 15, React 18, TypeScript, Tailwind CSS 3, shadcn/ui |
| **Shared Frontend** | Zustand (state), TanStack Query (server state), React Hook Form + Zod (forms), Axios (HTTP), Recharts (charts) |
| **Infrastructure** | Docker, Docker Compose, GitHub Actions, Nginx, Cloudflare CDN |
| **Observability** | Zerolog, Prometheus, Grafana, Sentry, OpenTelemetry |

## Project Structure

```
saas/
├── backend/          # Go API (modular monolith)
├── website/          # Customer-facing storefront
├── partner/          # Vendor/restaurant management portal
├── admin/            # Super-admin dashboard
├── docs/requirements/  # Complete requirements (01-13)
└── openspec/         # Spec-driven change proposals
```

## OpenSpec System

Read `openspec/AGENTS.md` when the request mentions planning, proposals, specs, new capabilities, breaking changes, or architecture shifts. Use openspec for creating and applying change proposals.

**Active proposals** in `openspec/changes/`:
- `redesign-order-system` — 10 critical production bugs (BLOCKS finance)
- `complete-partner-portal` — Wire mock UI to real APIs (~100+ subtasks)
- `add-customer-website-phase-17` — Customer website (1/10 done)
- `add-auth-and-tenancy` — Backend auth foundation (not started)
- `inventory-promos-orders` — Inventory, promos, order lifecycle (not started)
- `phase-9-14-finance-notifications-analytics` — Finance through analytics (proposal only)

## Architecture — Critical Rules

1. **Multi-tenancy**: Every domain table has `tenant_id UUID NOT NULL`. Every query filters by tenant. No exceptions.
2. **Modular monolith**: `backend/internal/modules/{name}/` — handler.go, service.go, repository.go
3. **SQLC only**: No ORM. All queries in `backend/internal/db/queries/*.sql`, generated to `db/sqlc/`
4. **Layer discipline**: Handler → Service → Repository → DB. Services also call Adapters for external APIs.
5. **Row-level isolation**: Shared DB, shared schema, tenant isolation via WHERE clauses and middleware

## Workflow (RPI)

1. **Research** — Explore codebase, read requirements in `docs/requirements/`, check openspec proposals
2. **Plan** — Use EnterPlanMode for non-trivial tasks. Check if an openspec proposal already exists.
3. **Implement** — Step by step, use TaskCreate to track progress on complex work

## Key Documentation

| Doc | Path |
|-----|------|
| Multi-tenancy | `docs/requirements/03-multi-tenancy.md` |
| Domain model | `docs/requirements/04-domain-model.md` |
| Features | `docs/requirements/05-feature-requirements.md` |
| Roles & portals | `docs/requirements/06-portals-and-roles.md` |
| Order lifecycle | `docs/requirements/07-order-lifecycle.md` |
| Pricing & finance | `docs/requirements/08-pricing-and-financials.md` |
| Database schema | `docs/requirements/09-database-schema.md` |
| API design | `docs/requirements/10-api-design.md` |
| Infrastructure | `docs/requirements/13-infrastructure.md` |

## Commands Reference

```bash
# Backend
cd backend && make dev          # Run API
cd backend && make test         # Tests with coverage + race detection
cd backend && make sqlc         # Regenerate SQLC types
cd backend && make migrate-up   # Run migrations
cd backend && make lint         # golangci-lint

# Frontend (any portal)
npm run dev                     # Dev server
npm run build                   # Production build
npm run lint                    # ESLint

# Infrastructure
docker-compose up               # PostgreSQL + Redis
```

## MCP Servers

Context7, Playwright, Claude in Chrome, DeepWiki — use when relevant.

---
Last updated: 2026-03-02
