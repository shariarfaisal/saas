# Design: Complete Partner Portal Integration

## Context

The Munchies partner portal was built in Phase 16 as a fully-scaffolded Next.js 14 app with professional UI and Zod-validated forms. The backend (Go modular monolith) exposes all required data under `/partner/*` routes — implemented across phases 4–14. The two systems are almost entirely disconnected. Only restaurant CRUD pages call real APIs. All other 13 feature areas use hardcoded `useState(mockData)`.

Stakeholders: restaurant owners / managers using the portal daily; platform engineers maintaining backend.

Constraints:
- Bangladesh market first (BDT, bKash/AamarPay)
- No ORM; SQLC only on backend
- Multi-tenant by design — all API calls carry tenant context from JWT
- Keep portal as SPA (no SSR for partner portal)

## Goals / Non-Goals

**Goals:**
- Replace every mock data pattern with real TanStack Query hooks
- Wire SSE live order feed into dashboard and kanban
- Add two missing backend APIs (team management, notification preferences)
- Complete variant/addon builder and bulk CSV import logic
- Enable image uploads via existing media API

**Non-Goals:**
- Redesign existing UI/UX — only data wiring changes
- Add new pages not already scaffolded
- Rider live map (GPS tracking) — deferred to Phase 2 as documented
- API keys management — deferred to Phase 2 as documented
- POS integration — deferred to Phase 2

## Decisions

### Decision: TanStack Query as primary data layer
All `useState(mockXxx)` patterns replaced with `useQuery` / `useMutation`. Zustand retains only auth/session state. Custom hook files (`use-orders.ts`, `use-menu.ts`, etc.) own all server state. This matches the installed but unused TanStack Query setup already in `app-providers.tsx`.

### Decision: Type-safe API layer via shared types
Create `partner/src/lib/types/` with TypeScript interfaces matching backend response shapes. Backend Go types are the source of truth; TypeScript types are manually maintained (no codegen in scope for this change).

### Decision: Team management via new backend module
The backend has `users`, `tenant_users`, and `roles` tables but no HTTP endpoints for the partner to list/invite/remove team members. A new `team` module is added to `backend/internal/modules/team/` with handler + service following existing patterns. Invitation creates a `pending_invitations` record and triggers an email via the existing email adapter.

### Decision: Notification preferences stored in user_profiles JSONB
Rather than a new migration, notification preference toggles (8 event types × 2 channels) are stored as JSONB in an existing or new `notification_preferences` column on the `tenant_users` join table. This avoids schema sprawl while keeping preferences per-user-per-tenant.

### Decision: SSE scoping for partner events
The existing `/api/v1/events/subscribe` SSE endpoint publishes to Redis pub/sub channels. Partner portal subscribes to a `tenant:{tenantId}:restaurant:{restaurantId}:orders` channel for new order and status-change events. No new SSE endpoint needed — channel naming convention is added to the auth middleware context.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| Type drift between Go backend and TS frontend | Shared type file; integration tests on key shapes |
| SSE reconnection on mobile/flaky networks | Existing `use-sse.ts` handles reconnect; add exponential backoff |
| Bulk CSV import with large files | Client-side validation before upload; 5MB size cap enforced by media API |
| Team invite email delivery | Use existing email adapter; show pending state in UI if email not confirmed |

## Migration Plan

No database migrations required for orders/menu/finance/promo/rider/analytics/content (all tables exist).  
One optional migration for `notification_preferences` JSONB column on `tenant_users` if not present.  
Rollback: revert backend route registration; frontend can fall back to empty states (no mock data to restore).

## Open Questions

- Does `GET /partner/restaurants` (list all restaurants for logged-in partner) already exist and return the correct shape? (Likely yes based on backend module analysis — needs confirmation)
- Does the promo performance stats endpoint exist at `GET /partner/promos/:id/stats` or must it be added?
- Is `pending_invitations` table present in the DB schema, or does invitation flow use a different mechanism?
