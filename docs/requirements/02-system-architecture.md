# 02 — System Architecture

## 2.1 Overview

The platform is a **modular monolith** — one deployable Go binary, cleanly divided into internal modules with strict dependency rules. This gives the simplicity of a monolith and the maintainability of separated concerns, without the operational overhead of microservices.

```
┌──────────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                                 │
│                                                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────┐  ┌──────────┐  │
│  │  website/   │  │  partner/   │  │   admin/     │  │  Rider   │  │
│  │  (Next.js)  │  │  (Next.js)  │  │  (Next.js)   │  │  (PWA)   │  │
│  └──────┬──────┘  └──────┬──────┘  └──────┬───────┘  └────┬─────┘  │
└─────────┼────────────────┼────────────────┼───────────────┼─────────┘
          │                │                │               │
          └────────────────┴───────┬─────────┘               │
                                   │  HTTPS / REST API        │
┌──────────────────────────────────┼──────────────────────────┼────────┐
│                        API GATEWAY / NGINX                           │
│                    (TLS termination, rate limiting)                  │
└──────────────────────────────────┬───────────────────────────────────┘
                                   │
┌──────────────────────────────────▼───────────────────────────────────┐
│                        BACKEND  (Go)                                 │
│                                                                      │
│  ┌───────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │  tenant   │ │  order   │ │  catalog │ │  rider   │ │ finance  │ │
│  │  module   │ │  module  │ │  module  │ │  module  │ │  module  │ │
│  └───────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘ │
│  ┌───────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │   auth    │ │  promo   │ │ delivery │ │ notif.   │ │analytics │ │
│  │  module   │ │  module  │ │  module  │ │  module  │ │  module  │ │
│  └───────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘ │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │              Shared: DB Pool, Redis, Logger, Config            │ │
│  └────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────┘
          │              │              │              │
   ┌──────▼──────┐ ┌─────▼──────┐ ┌────▼───┐  ┌──────▼──────┐
   │ PostgreSQL  │ │   Redis    │ │  S3 /  │  │  External   │
   │  (primary)  │ │ (cache +   │ │  R2    │  │  Services   │
   │             │ │  queues)   │ │(files) │  │(bKash,SMS..)│
   └─────────────┘ └────────────┘ └────────┘  └─────────────┘
```

---

## 2.2 Tech Stack

### Backend
| Component | Choice | Rationale |
|-----------|--------|-----------|
| Language | **Go 1.22+** | Performance, static typing, excellent concurrency, small binary |
| Framework | **Chi** or **Echo** (HTTP router) | Lightweight, idiomatic Go, no magic |
| Database | **PostgreSQL 16+** | Reliable, ACID, JSON support, excellent for complex queries |
| Query layer | **SQLC** | Type-safe SQL, no ORM overhead, full control over queries |
| Migrations | **golang-migrate** | Simple SQL migration files, version-controlled |
| Caching | **Redis 7+** | Session cache, rate limiting, job queues, real-time pub/sub |
| Queue/Workers | **asynq** (Redis-based) | Reliable async jobs, scheduled tasks, retries |
| Auth | **JWT** (access + refresh tokens) | Stateless, works well with multi-tenant |
| File upload | **AWS S3 / Cloudflare R2** | Scalable object storage |
| Real-time | **Server-Sent Events (SSE)** for orders; WebSocket for rider tracking | SSE is simpler for one-directional updates |
| Validation | **go-playground/validator** | Struct-level validation |
| Config | **viper** + environment variables | 12-factor app |
| Logging | **zerolog** or **zap** | Structured JSON logs |
| Observability | **OpenTelemetry** + Prometheus metrics | Production monitoring |

### Frontend (all portals)
| Component | Choice | Rationale |
|-----------|--------|-----------|
| Framework | **Next.js 14+** (App Router) | SSR/SSG for SEO (website), fast SPA for dashboards |
| Language | **TypeScript** | Type safety, better DX |
| Styling | **Tailwind CSS** + **shadcn/ui** | Rapid development, consistent design system |
| State | **Zustand** (client) + **TanStack Query** (server) | Lightweight, no boilerplate |
| Forms | **React Hook Form** + **Zod** | Performant forms with schema validation |
| Maps | **Google Maps / Mapbox** | Rider tracking, delivery zone display |
| Charts | **Recharts** | Analytics dashboards |
| Real-time | **EventSource (SSE)** for orders | Native browser API |

### Infrastructure
| Component | Choice |
|-----------|--------|
| Container | Docker + Docker Compose (dev) |
| Orchestration | Kubernetes or Docker Swarm (prod) |
| CI/CD | GitHub Actions |
| Reverse Proxy | Nginx |
| SSL | Let's Encrypt / Cloudflare |
| Monitoring | Grafana + Prometheus |
| Logging | Loki or ELK stack |
| CDN | Cloudflare |

---

## 2.3 Repository & Folder Structure

```
saas/
├── backend/                    # Go API server
│   ├── cmd/
│   │   └── api/
│   │       └── main.go         # Entry point
│   ├── internal/
│   │   ├── config/             # App config loading
│   │   ├── server/             # HTTP server setup, routes
│   │   ├── middleware/         # Auth, CORS, rate limit, tenant resolver
│   │   ├── db/
│   │   │   ├── migrations/     # SQL migration files
│   │   │   ├── queries/        # SQLC query files (.sql)
│   │   │   └── sqlc/           # SQLC generated code (do not edit manually)
│   │   ├── modules/
│   │   │   ├── tenant/         # Vendor/tenant management
│   │   │   ├── auth/           # Authentication & authorization
│   │   │   ├── user/           # Customer & staff user management
│   │   │   ├── catalog/        # Restaurants, menus, products, categories
│   │   │   ├── order/          # Order creation, lifecycle, management
│   │   │   ├── rider/          # Rider management, tracking, assignment
│   │   │   ├── delivery/       # Delivery zones, hubs, charge calculation
│   │   │   ├── payment/        # Payment gateways, transactions
│   │   │   ├── promo/          # Promotions, promo codes, discounts
│   │   │   ├── finance/        # Commissions, invoices, payouts
│   │   │   ├── notification/   # Push, SMS, email notifications
│   │   │   ├── analytics/      # Reports, dashboards, aggregations
│   │   │   ├── media/          # File upload handling
│   │   │   └── search/         # Search functionality
│   │   ├── worker/             # Background job handlers
│   │   ├── pkg/                # Shared internal utilities
│   │   │   ├── apperror/       # Typed error definitions
│   │   │   ├── pagination/     # Cursor/offset pagination helpers
│   │   │   ├── validator/      # Custom validation rules
│   │   │   └── timeutil/       # Time helpers (operational date, etc.)
│   │   └── adapters/           # External service clients
│   │       ├── bkash/
│   │       ├── aamarpay/
│   │       ├── sslcommerz/
│   │       ├── barikoi/        # BD map service
│   │       ├── fcm/            # Firebase push notifications
│   │       └── sms/            # SMS gateway (SSL Wireless, etc.)
│   ├── sqlc.yaml               # SQLC config
│   ├── Makefile
│   ├── Dockerfile
│   └── go.mod
│
├── website/                    # Customer-facing Next.js app
│   ├── app/                    # Next.js App Router
│   ├── components/
│   ├── lib/
│   ├── public/
│   └── package.json
│
├── partner/                    # Vendor/partner portal (Next.js)
│   ├── app/
│   ├── components/
│   ├── lib/
│   └── package.json
│
├── admin/                      # Super admin panel (Next.js)
│   ├── app/
│   ├── components/
│   ├── lib/
│   └── package.json
│
├── docs/
│   └── requirements/           # This documentation
│
└── docker-compose.yml          # Local development setup
```

---

## 2.4 Module Dependency Rules

Modules must not create circular dependencies. The allowed dependency direction is:

```
HTTP Handler → Service → Repository (SQLC) → Database
                ↓
            Adapters (external services)
```

- **Handlers**: Parse requests, validate input, call service, return response.
- **Services**: Business logic only. Orchestrate repositories and adapters. No HTTP concerns.
- **Repositories**: Database access only via SQLC. No business logic.
- **Adapters**: Wrappers for external APIs (payment, SMS, push). Return typed results.

Cross-module communication within the same binary is done via **service interfaces** — not direct struct access. This makes future module extraction easier.

---

## 2.5 Multi-Tenancy Strategy

See [03-multi-tenancy.md](./03-multi-tenancy.md) for full details.

**Summary:** Row-level tenancy. Every table that is tenant-scoped has a `tenant_id` column. The tenant is resolved from the request domain (each vendor gets a subdomain: `vendor.platform.com`) or from the authenticated user's `tenant_id`. Middleware automatically injects `tenant_id` into the request context, and all repository queries filter by it.

---

## 2.6 Authentication Strategy

- **Customers**: Phone number + OTP (primary), email+password optional.
- **Partners/Vendors**: Email + password with 2FA option.
- **Admins**: Email + password with mandatory 2FA.
- **Riders**: Phone number + OTP via dedicated endpoint.

All tokens are **JWT** (short-lived access token 15min + long-lived refresh token 7d, stored in httpOnly cookie). The access token payload contains: `user_id`, `tenant_id`, `role`, `permissions`.

---

## 2.7 API Design Principles

See [10-api-design.md](./10-api-design.md) for full details.

- **REST** over JSON. No GraphQL (complexity not justified).
- All routes versioned: `/api/v1/...`
- Multi-tenant routing: tenant resolved from JWT claim or `X-Tenant-ID` header.
- Consistent error format: `{ "error": { "code": "ORDER_NOT_FOUND", "message": "..." } }`
- Pagination: cursor-based for lists, with `next_cursor` in response.
- All times in UTC ISO 8601.

---

## 2.8 Real-time Architecture

| Use Case | Technology |
|----------|-----------|
| New order notification to partner dashboard | SSE (Server-Sent Events) |
| Order status updates to customer | SSE |
| Rider location tracking | WebSocket (rider → server → admin/partner) |
| Live order board on partner portal | SSE |

SSE connections are scoped per-tenant and per-user. Redis pub/sub is used to fan out events to the correct SSE connection on any server node.

---

## 2.9 Background Jobs

Scheduled and async jobs run via **asynq** (Redis-backed):

| Job | Trigger | Description |
|-----|---------|-------------|
| `order:auto_confirm` | Scheduled (every 1 min) | Auto-confirm orders after timeout |
| `order:auto_cancel` | Scheduled (every 5 min) | Cancel pending orders older than X minutes |
| `rider:auto_assign` | Event-driven (order created) | Find and assign best available rider |
| `invoice:generate` | Scheduled (daily) | Generate daily settlement invoices per restaurant |
| `report:daily` | Scheduled (end of day) | Generate daily sales reports |
| `promo:expiry` | Scheduled (hourly) | Deactivate expired promos |
| `notification:send` | Event-driven | Dispatch queued notifications |
| `product:discount_expiry` | Scheduled (hourly) | Remove expired product discounts |
| `order:analytics_sync` | Event-driven (order complete) | Write to analytics tables |
