# Change: Phase 15 Super Admin Panel (TASK-071 through TASK-079)

## Why
The platform backend now exposes cross-tenant admin APIs, but there is no super-admin frontend project to operate tenants, users, orders, finance, disputes, and platform configuration. Phase 15 requires a dedicated Next.js admin portal with secure authentication, 2FA, and operational pages that align with documented requirements.

## What Changes
- **TASK-071** — Initialize `admin/` Next.js 14 (TypeScript, Tailwind, ESLint, Prettier), configure shadcn/ui primitives, TanStack Query, Zustand, React Hook Form + Zod, and an API client with base URL config, request/response interceptors, token refresh, centralized error mapping, and `X-Request-ID` on every request.
- **TASK-072** — Implement admin authentication UI and session model:
  - `/auth/login` email + password form
  - mandatory TOTP flow (`/auth/totp-setup`, `/auth/totp-verify`) with first-login QR setup path
  - httpOnly-cookie session assumption with route protection via server-side checks + middleware
  - logout handling and session-expiry redirect behavior
- **TASK-073** — Implement dashboard page with global KPI cards, revenue trend chart, active tenant table, and system health block sourced from `/admin/analytics/overview`.
- **TASK-074** — Implement tenant management views: list, create, detail/edit, suspend/reinstate modal with reason, tenant analytics drill-down section, and impersonation action opening partner portal.
- **TASK-075** — Implement platform-wide user management: search with tenant filter, user detail drawer, suspend with reason, GDPR delete flow with explicit confirmation and wipe status.
- **TASK-076** — Implement cross-tenant orders management: filters (tenant/status/date), detail modal (timeline/payment/rider/audit), force status override with mandatory reason, issue-linkout action.
- **TASK-077** — Implement finance management: commission ledger with period filter + CSV export, invoice list + generate invoice action, invoice detail state actions (approve/finalize/mark-paid), payout tracking table.
- **TASK-078** — Implement disputes page: issues queue filters, issue detail with order context + message thread + refund/accountability controls, approve/reject actions, and resolved-history segment.
- **TASK-079** — Implement platform config pages for feature flags, payment gateway config, SMS/email provider config, maintenance mode with confirmation; include audit-log creation payloads on writes.

## Impact
- Affected specs: admin-panel
- Affected code: `admin/` (new Next.js app), `TASKS.md` (Phase 15 task status)
- Dependencies added (npm): TanStack Query, Zustand, React Hook Form, Zod, Axios, Recharts, shadcn/ui utility deps
- Security impact: admin session enforcement, mandatory 2FA pathing, mutation reason fields for sensitive actions, and request ID propagation for traceability
- No backend API contracts are changed by this proposal
