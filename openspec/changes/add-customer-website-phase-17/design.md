## Context
The customer website must serve SEO-indexable pages for discovery while supporting authenticated, personalized purchase flows. Tenant context is resolved from host/subdomain at request time and must flow into both server and client logic safely.

## Goals / Non-Goals
- Goals:
  - Build a production-ready Next.js customer app foundation with tenant-aware SSR.
  - Standardize data/access patterns for the remaining customer pages.
  - Preserve security requirements (httpOnly cookie sessions, protected routes, idempotent checkout intent).
- Non-Goals:
  - Re-implement backend business logic in frontend.
  - Introduce new backend APIs outside existing contract assumptions.

## Decisions
- Decision: Use App Router server components for tenant resolution from `headers()` and pass resolved config through a client-safe provider.
  - Alternatives considered: client-side host parsing only (rejected: weak SSR/SEO support).
- Decision: Keep data access abstracted behind `website/src/lib/api/*` fetchers to isolate API contract changes.
  - Alternatives considered: direct fetch calls from all components (rejected: hard to maintain/test).
- Decision: Use Zustand for local UI/session/cart state and TanStack Query for server-state caching.
- Decision: Use Next Metadata API + `next-seo` package for reusable SEO config and JSON-LD utilities.

## Risks / Trade-offs
- Large Phase 17 surface area can lead to partial implementation drift.
  - Mitigation: codify phased tasks and complete baseline architecture first.
- Client/server tenant mismatch risk.
  - Mitigation: resolve tenant once on server and serialize only validated tenant config.
- Payment redirect race conditions.
  - Mitigation: preserve explicit `PENDING` state machine and callback page guards.

## Migration Plan
1. Scaffold `website/` and baseline architecture.
2. Add auth/session plumbing and protected routes.
3. Build discovery/menu/cart/checkout flows.
4. Add payment redirects, tracking, and account surfaces.
5. Validate with lint/build and targeted smoke tests.

## Open Questions
- Final backend endpoint availability and response shape for all storefront resources.
- Final design system parity requirements (if shadcn/ui package should be introduced now vs later).
