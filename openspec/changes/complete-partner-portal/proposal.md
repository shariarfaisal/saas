# Change: Complete Partner Portal — Full API Integration & Backend Gaps

## Why

The partner portal (`partner/`) was scaffolded in Phase 16 with comprehensive UI for all 10 feature areas (dashboard, orders, menu, finance, promotions, riders, analytics, content, team, settings). However, only the **Restaurant management** pages are connected to real backend APIs. Every other feature — orders, menu, finance, promotions, riders, analytics, content, team, and settings — operates on hardcoded `useState(mockData)` with **zero API calls**. Additionally, two backend capabilities required by the portal (team/staff management CRUD and notification preferences API) are either missing or incomplete. This change bridges the full gap between the working backend and the non-functional frontend.

## Gap Summary

### Partner Portal (Frontend) — 13 of 14 feature areas disconnected

| Feature Area | Current Status | Gap |
|---|---|---|
| Auth / Login | ✅ Real API | — |
| Restaurant Management | ✅ Real API | Minor: list page still mock |
| Dashboard KPIs | ❌ Mock data | Needs `GET /partner/dashboard/summary` integration |
| Live Order Feed | ❌ Not wired | SSE hook exists (`use-sse.ts`) but not used |
| Order Management | ❌ Mock data | Full kanban & history need order API |
| Menu / Catalog | ❌ Mock data | Categories + products CRUD; variant & addon builders broken |
| Finance | ❌ Mock data | Invoice list/detail, payment history, PDF download |
| Promotions | ❌ Mock data | Promo CRUD + performance stats |
| Rider Management | ❌ Mock data | Rider list/detail, attendance, earnings, penalties |
| Analytics / Reports | ❌ Mock data | Sales charts, top products, heatmap, CSV export |
| Content Management | ❌ Mock data | Banners, sections, stories — no persistence |
| Team Management | ❌ Mock data | Invite / list / remove — no backend calls |
| Settings | ❌ Mock data | Notification preferences — no persistence |
| Image / File Uploads | ❌ UI only | No upload logic; media API exists but not called |
| Bulk CSV Import | ❌ Stub | Modal exists; no validation or upload logic |
| Invoice PDF Download | ❌ Stub | Button exists; no PDF generation call |

### Backend — Missing / Incomplete APIs

| Gap | Notes |
|---|---|
| Team / Staff management API | No `GET/POST/DELETE /partner/team` endpoints; roles exist in DB but no CRUD surface |
| Notification preferences API | No `GET/PUT /partner/settings/notifications` endpoints |
| Partner SSE event endpoint | SSE infra exists (`/api/v1/events/subscribe`) but partner-scoped channel unclear |
| Restaurant list API for partner | Backend `GET /partner/restaurants` exists; partner portal list page uses mock |
| Promo performance stats | Backend tracks usage; no `GET /partner/promos/:id/stats` endpoint confirmed |

## What Changes

### Backend additions
- **Team management API** (`/partner/team`): list members, invite by email, assign/change role, remove member
- **Notification preferences API** (`/partner/settings/notifications`): get and update per-user push/email toggle map
- **Promo performance stats endpoint** (`GET /partner/promos/:id/stats`): usage count, total discount given, unique users
- **Partner SSE channel scope** clarified: tenant+restaurant-scoped events for new orders, status changes

### Partner portal — replace mock data with real API
- Dashboard: KPI summary + 7-day trend + live incoming order SSE panel
- Orders: kanban status board, order detail drawer, history table — all wired to order API
- Menu: category CRUD, product CRUD, variant/addon builder logic completed, bulk CSV import, image upload
- Finance: invoice list/detail, payment history, PDF download — wired to finance API
- Promotions: promo list, create/edit/deactivate — wired to promo API; performance stats page
- Riders: rider list, create/edit, detail (stats/earnings/attendance/penalties) — wired to rider API
- Analytics: sales report, top products, peak hours heatmap, order breakdown, CSV export — wired to analytics API
- Content: banner/section/story CRUD with image upload — wired to content API
- Team: invite member, list, remove, role change — wired to new team API
- Settings: vendor profile, notification preferences — wired to new prefs API

### Infrastructure / DX
- Introduce TanStack Query hooks (currently set up but unused) for all data fetching
- Replace all `useState(mockXxx)` patterns with `useQuery` / `useMutation` hooks
- Centralise API types in `src/lib/types/` aligned to backend response shapes
- Add `use-orders.ts`, `use-menu.ts`, `use-finance.ts`, etc. custom hook files
- SSE `use-sse.ts` hook wired into dashboard incoming-order panel and order kanban

## Impact

- Affected specs: `partner-dashboard`, `partner-order-management`, `partner-menu-management`, `partner-finance`, `partner-promotions`, `partner-riders`, `partner-analytics`, `partner-content`, `partner-team`, `partner-settings`
- Affected code:
  - `partner/src/app/(dashboard)/` — all page.tsx files replacing mock state
  - `partner/src/hooks/` — new query/mutation hook files
  - `partner/src/lib/types/` — new type definitions
  - `backend/internal/modules/` — new `team/` module handlers; additions to `notification/` and `promo/` handlers
- **No breaking changes** to existing public APIs or database schema
- No new database migrations required (team members exist in `users` table; notification prefs JSONB column may be added to `tenant_users` or `user_profiles`)
