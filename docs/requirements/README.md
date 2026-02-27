# FoodSaaS Platform — Requirements & Documentation

> **Source of Truth for all development decisions.**  
> Every feature, architecture choice, and design decision lives here.  
> Before building anything, read the relevant section of this documentation.

---

## Table of Contents

| # | Document | Description |
|---|----------|-------------|
| 01 | [Product Vision](./01-product-vision.md) | Why we're building this, target market, goals |
| 02 | [System Architecture](./02-system-architecture.md) | Tech stack, folder structure, service boundaries |
| 03 | [Multi-Tenancy Model](./03-multi-tenancy.md) | How SaaS tenancy works, vendor isolation, onboarding |
| 04 | [Domain Model & Entities](./04-domain-model.md) | All core entities, relationships, business rules |
| 05 | [Feature Requirements](./05-feature-requirements.md) | Full feature list across all portals |
| 06 | [Portals & User Roles](./06-portals-and-roles.md) | All portals, user types, permissions |
| 07 | [Order Lifecycle](./07-order-lifecycle.md) | End-to-end order flow, statuses, state machine |
| 08 | [Pricing & Financials](./08-pricing-and-financials.md) | Delivery charges, commission, invoicing, payouts |
| 09 | [Database Schema](./09-database-schema.md) | PostgreSQL schema design, tables, indexes |
| 10 | [API Design](./10-api-design.md) | REST API conventions, auth, versioning |
| 11 | [Notifications & Real-time](./11-notifications.md) | Push, WebSocket, SMS, email notification system |
| 12 | [Analytics & Reporting](./12-analytics.md) | Reports, dashboards, metrics |
| 13 | [Infrastructure & DevOps](./13-infrastructure.md) | Deployment, CI/CD, environments |
| 14 | [Gap Analysis & Remediation](./14-gap-analysis-and-remediation.md) | Audit findings, risks, priority fixes |
| 15 | [World-Class Quality Gates](./15-world-class-quality-gates.md) | Release criteria, SLAs/SLOs, engineering guardrails |

---

## Quick Reference

**Tech Stack**
- Backend: **Go** (modular monolith) + **PostgreSQL** + **SQLC**
- Frontend: **Next.js 14+** (App Router) for all web portals
- Cache: **Redis**
- Queue: **Redis** (via asynq or similar)
- Real-time: **WebSocket** / SSE
- File Storage: **S3-compatible** (AWS S3 / Cloudflare R2)
- Maps: **Google Maps API** + **Barikoi** (BD local maps)

**Portals**
1. `website/` — Customer-facing ordering website (Next.js)
2. `partner/` — Vendor & restaurant management portal (Next.js)
3. `admin/` — Super-admin platform management (Next.js)
4. `backend/` — Go REST API server

**Key Design Principles**
- **Multi-tenant by design** — every resource is scoped to a tenant
- **Modular monolith** — single deployable backend, clean internal module boundaries
- **No microservices** — simplicity and developer velocity first
- **API-first** — all portals consume the same REST API
- **Extensible** — built to support food delivery today, any commerce vertical tomorrow
