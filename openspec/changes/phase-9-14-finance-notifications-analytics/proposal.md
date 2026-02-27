# Change: Phases 9–14 — Finance, Notifications, Background Jobs, Issues/Ratings/Search, Content, Analytics

## Why
The platform has core order, payment, and restaurant management in place. To reach production readiness,
we need financial settlement (invoices, ledger, wallets), real-time notifications (FCM, SMS, email, SSE),
background job processing (asynq), order issue resolution, ratings/reviews, full-text search,
content management (banners/stories), and analytics/reporting APIs.

## What Changes

### Phase 9 — Finance & Invoicing
- **TASK-048** — SQLC queries for `invoices` table; `InvoiceService.GenerateForRestaurant(restaurantID, periodStart, periodEnd)` with exact settlement formula; idempotent via UNIQUE constraint.
- **TASK-049** — Partner API: `GET /partner/finance/summary`, `GET /partner/finance/invoices`, `GET /partner/finance/invoices/:id`; Admin API: `POST /admin/finance/invoices/generate`, `PATCH /admin/finance/invoices/:id/finalize`, `PATCH /admin/finance/invoices/:id/mark-paid`; audit_log entries.
- **TASK-050** — PDF invoice renderer (Go HTML template → PDF); `GET /partner/finance/invoices/:id/pdf`; S3/R2 caching; `Content-Disposition: attachment`.
- **TASK-051** — Migration for `ledger_accounts` and `ledger_entries`; seed platform accounts; append-only `LedgerService.Record(...)`.
- **TASK-052** — `WalletService.Credit/Debit` with balance-after tracking; cashback on DELIVERED orders; max wallet spend enforcement; `GET /api/v1/me/wallet`.

### Phase 10 — Notifications & Real-time
- **TASK-053** — `adapters/fcm/` using Firebase Admin SDK; `NotificationService.SendPush`; handle token-not-registered; persist to `notifications` table.
- **TASK-054** — `adapters/sms/` for SSL Wireless BD API; `SMSService.Send(phone, template, vars)`; Redis OTP rate limiting; Twilio fallback config.
- **TASK-055** — `adapters/email/` for SendGrid/AWS SES; HTML templates; `EmailService.Send(to, templateName, vars)`.
- **TASK-056** — SSE handler `GET /api/v1/events/subscribe`; JWT auth; Redis pub/sub channels; Last-Event-ID replay.
- **TASK-057** — Outbox event processor worker; publish to Redis pub/sub; trigger push/SMS/email; exponential backoff; dead-letter after 5 retries.

### Phase 11 — Background Jobs
- **TASK-058** — asynq server setup with Redis; job handler registry; queue priorities; concurrency config.
- **TASK-059** — Scheduled jobs: auto-confirm, auto-cancel, promo/discount expiry, notification cleanup.
- **TASK-060** — Invoice daily generation; order analytics sync; daily report cache.
- **TASK-061** — Data retention cleanup: order_timeline > 1yr, rider_travel_logs > 6mo, search_logs > 90d, audit_logs > 2yr.

### Phase 12 — Order Issues, Ratings & Search
- **TASK-062** — Order issues CRUD; customer create, partner view/respond, admin resolve/approve/reject refund.
- **TASK-063** — Ratings migration & API; one-per-order; aggregate restaurant rating; partner response.
- **TASK-064** — Full-text search with pg_trgm GIN; autocomplete; search logging; partner search reports.

### Phase 13 — Content Management
- **TASK-065** — Banners, stories, sections CRUD (partner); storefront GET APIs (area-filtered, active).
- **TASK-066** — Media upload service: MIME validation, size limits, UUID filename, S3/R2 upload, CDN URL return.

### Phase 14 — Analytics & Reporting
- **TASK-067** — Partner dashboard KPIs: today's counts/revenue, 7-day trends, top products; Redis cache.
- **TASK-068** — Partner sales report: date range, restaurant filter, group by day/week/month.
- **TASK-069** — Partner product/peak-hour/rider analytics with optional CSV export.
- **TASK-070** — Admin cross-tenant analytics: platform totals, commission revenue, order volume, tenant deep-dive.

## Impact
- Affected code: `backend/internal/modules/` (new: finance, notification, issue, rating, search, content, analytics, worker),
  `backend/internal/platform/` (new: fcm, email adapters), `backend/internal/db/queries/`, `backend/internal/db/sqlc/`,
  `backend/internal/db/migrations/`, `backend/internal/server/server.go`, `backend/cmd/api/main.go`, `backend/go.mod`
- New dependencies: `github.com/hibiken/asynq` (background jobs), `github.com/SebastiaanKlipworthy/go-wkhtmltopdf` or HTML template approach
- New migrations: 000020 (ledger tables), 000021 (search logs, additional indexes)
- No **BREAKING** changes to existing public interfaces
