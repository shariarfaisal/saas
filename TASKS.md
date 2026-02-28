# SaaS Platform — End-to-End Task Checklist

> **How to use this list**
> Each task is a self-contained _session_. Before implementing, write a short OpenSpec proposal (endpoint signatures, DB changes, business rules) then implement. Tasks are ordered so every dependency is built before it is needed. Work top-to-bottom.
>
> Legend: `[ ]` = not started · `[~]` = in progress · `[x]` = done

---

## PHASE 0 — Repository Foundation

- [~] **TASK-001 — Monorepo scaffold & tooling**
  Create the full folder structure under `saas/` (`backend/`, `website/`, `partner/`, `admin/`, `docker-compose.yml`); initialise Go module (`go mod init`), create `Makefile` with `dev`, `build`, `test`, `lint`, `sqlc`, `migrate-up/down/create` targets.

- [~] **TASK-002 — Docker Compose local stack**
  Write `docker-compose.yml` running PostgreSQL 16, Redis 7, and volume mounts for each app; add a `healthcheck` for each service; document `make dev` workflow.

- [~] **TASK-003 — Go application skeleton**
  Bootstrap `cmd/api/main.go` (HTTP server start/graceful shutdown), `internal/config/` (viper + env loading for all config fields defined in doc 13), structured JSON logger (`zerolog`), environment detection (`local/dev/staging/production`).

- [~] **TASK-004 — Database migration framework**
  Set up `golang-migrate` with `internal/db/migrations/`; write `0001_init_extensions.sql` (pgcrypto, pg_trgm); write `0002_create_enums.sql` (all 30 ENUMs from doc 09); confirm migrations run cleanly in `make migrate-up`.

- [~] **TASK-005 — SQLC code-gen pipeline**
  Add `sqlc.yaml` with correct overrides (UUID → `github.com/google/uuid.UUID`, numeric → `shopspring/decimal`, timestamptz → `time.Time`); wire SQLC into `make sqlc`; confirm generated code compiles against empty queries dir.

- [~] **TASK-006 — Shared packages (apperror, pagination, validator, timeutil)**
  Implement typed app-error package with standard codes (NOT_FOUND, FORBIDDEN, VALIDATION_ERROR, CONFLICT, etc.); cursor-based pagination helper; custom validator registration for phone/UUID/decimal; Bangladesh timezone helper.

- [~] **TASK-007 — HTTP server, router & base middleware**
  Set up Chi/Echo router; register global middleware: request-ID injector, structured request logger, CORS (configured from `ALLOWED_ORIGINS`), panic recoverer, content-type enforcer; wire `/healthz` and `/readyz` endpoints.

---

## PHASE 1 — Multi-Tenancy & Authentication

- [x] **TASK-008 — Tenants table, SQLC queries & seed**
      Write migration for `tenants` table; SQLC queries: `CreateTenant`, `GetTenantBySlug`, `GetTenantByID`, `GetTenantByDomain`, `UpdateTenant`, `UpdateTenantStatus`, `ListTenants`; write a seed script for one dev tenant.

- [x] **TASK-009 — Tenant resolver middleware**
      Implement three tenant-resolution strategies in order: (1) subdomain from `Host` header, (2) `tenant_id` JWT claim, (3) `X-Tenant-ID` header (API-key-only); cache resolved tenant in Redis for 60s; inject `TenantContext` into request context; return 403 for suspended/cancelled tenants.

- [x] **TASK-010 — Users & addresses tables, SQLC queries**
      Write migration for `users`, `user_addresses`, `otp_verifications`; SQLC queries: `CreateUser`, `GetUserByPhone`, `GetUserByEmail`, `GetUserByID`, `UpdateUser`, `SoftDeleteUser`, `CreateAddress`, `ListAddresses`, `UpdateAddress`, `DeleteAddress`; add `SetDefaultAddress` logic.

- [x] **TASK-011 — OTP authentication (customer & rider)**
      Implement `POST /api/v1/auth/otp/send` (generate 6-digit OTP, store hashed, send via SMS adapter, rate-limit 3/phone/10min) and `POST /api/v1/auth/otp/verify` (verify hash, create user if first-time, return JWT access+refresh tokens as httpOnly cookies + JSON body).

- [x] **TASK-012 — Email+password authentication (partner & admin)**
      Implement `POST /api/v1/auth/login` (email+password, bcrypt compare, role-check, return JWTs), `POST /api/v1/auth/refresh` (rotate refresh token), `POST /api/v1/auth/logout` (invalidate refresh token via Redis deny-list), `POST /api/v1/auth/password/reset-request` + `POST /api/v1/auth/password/reset` (email-link flow).

- [x] **TASK-013 — JWT middleware & RBAC**
      Implement JWT validation middleware (parse access token, verify signature, check deny-list); implement `RequireRoles(...role)` middleware guard; implement `RequireTenantMatch()` guard (user.tenant_id must match resolved tenant except super-admin); attach `UserContext` to request.

- [x] **TASK-014 — Idempotency key infrastructure**
      Write migration for `idempotency_keys` table; implement Go middleware that reads `Idempotency-Key` header on POST/PATCH, checks DB for existing response (same key + matching hash → replay, different hash → 409), stores new response snapshot after handler completes; apply to `/orders` and payment endpoints.

- [x] **TASK-015 — Current user API & profile management**
      Implement `GET /api/v1/me` (profile), `PATCH /api/v1/me` (update name/email/avatar/dob/gender), `GET/POST/PUT/DELETE /api/v1/me/addresses`, `GET /api/v1/me/wallet` (balance + paginated wallet transactions), `GET /api/v1/me/notifications`, `PATCH /api/v1/me/notifications/:id/read`.

---

## PHASE 2 — Delivery Infrastructure

- [ ] **TASK-016 — Hubs & delivery zone tables, SQLC queries & API**
      Write migration for `hubs` and `hub_areas`; SQLC queries: `CreateHub`, `GetHubByID`, `ListHubsByTenant`, `CreateHubArea`, `GetHubAreaByName`, `ListHubAreas`, `UpdateHubArea`, `DeleteHubArea`; implement partner API `GET/POST/PUT/DELETE /partner/hubs` and `/partner/hubs/:id/areas`.

- [ ] **TASK-017 — Delivery charge calculation service**
      Implement `DeliveryChargeService.Calculate(tenantID, customerArea, hubID)` using zone-based model; handle missing area (return error or default charge per tenant config); expose as internal service used by order creation and the pre-calc endpoint `POST /api/v1/orders/charges/calculate`.

---

## PHASE 3 — Restaurant & Catalog

- [ ] **TASK-018 — Restaurants table, SQLC queries & partner CRUD**
      Write migration for `restaurants` and `restaurant_operating_hours`; SQLC queries covering full CRUD + `GetBySlug`, `ListByTenant`, `ListAvailableByHubAndArea`; implement partner API: `GET/POST /partner/restaurants`, `GET/PUT/DELETE /partner/restaurants/:id`, `PATCH /partner/restaurants/:id/availability`, `GET/PUT /partner/restaurants/:id/hours`.

- [ ] **TASK-019 — Categories CRUD & reorder API**
      Write migration for `categories`; SQLC queries: `CreateCategory`, `ListCategoriesByRestaurant`, `UpdateCategory`, `DeleteCategory`, `ReorderCategories`; implement partner API: `GET/POST /partner/restaurants/:id/categories`, `PUT/DELETE /partner/restaurants/:id/categories/:cat_id`, `PATCH /partner/restaurants/:id/categories/reorder`.

- [ ] **TASK-020 — Products (flat price), availability & image upload**
      Write migration for `products`; SQLC queries for product CRUD + availability update; implement partner API: `GET/POST /partner/restaurants/:id/products`, `GET/PUT/DELETE /partner/products/:id`, `PATCH /partner/products/:id/availability`; wire S3/R2 image upload via `POST /api/v1/media/upload`.

- [ ] **TASK-021 — Product variants & addons**
      Write migration for `product_variants`, `product_variant_items`, `product_addons`, `product_addon_items`; SQLC queries; extend product create/update API to accept variants+addons in a single nested payload; validate `min_select`/`max_select` logic server-side.

- [ ] **TASK-022 — Product discounts**
      Write migration for `product_discounts`; SQLC queries: `UpsertProductDiscount`, `GetActiveDiscount`, `DeleteDiscount`, `ExpireDiscounts`; implement partner API `POST/DELETE /partner/products/:id/discount`; implement discount expiry check consumed by the discount-expiry background job.

- [ ] **TASK-023 — Storefront public catalog API**
      Implement public (no-auth, tenant-scoped) endpoints: `GET /api/v1/storefront/config`, `GET /api/v1/storefront/areas`, `GET /api/v1/storefront/restaurants` (filters: area, cuisine, open-now, sort), `GET /api/v1/restaurants/:slug` (full restaurant + categories + products), `GET /api/v1/products/:id`.

- [ ] **TASK-024 — Menu bulk upload (CSV import)**
      Implement `POST /partner/products/bulk-upload` accepting CSV file; parse, validate, and batch-upsert categories + products + variants + addons; return job-style response with success count and per-row error list.

- [ ] **TASK-025 — Menu duplication between restaurants**
      Implement `POST /partner/restaurants/:id/menu/duplicate` with `source_restaurant_id` in body; copy all categories, products, variants, and addons to target restaurant (same tenant only); return summary of items copied.

---

## PHASE 4 — Inventory

- [ ] **TASK-026 — Inventory tracking (stock management)**
      Write migration for `inventory`; SQLC queries: `GetInventory`, `AdjustStock`, `ListLowStock`, `ReserveStock`, `ReleaseStock`; implement partner API: `GET /partner/inventory`, `POST /partner/inventory/adjust`, `GET /partner/inventory/low-stock`; stock-check hook wired into order creation.

---

## PHASE 5 — Promotions Engine

- [ ] **TASK-027 — Promos table, SQLC queries & partner CRUD**
      Write migration for `promos` and `promo_usages`; SQLC queries: `CreatePromo`, `GetPromoByCode`, `ListPromos`, `UpdatePromo`, `DeactivatePromo`, `IncrementUsage`, `GetUsageByUserAndPromo`; implement partner API: `GET/POST /partner/promos`, `GET/PUT /partner/promos/:id`, `PATCH /partner/promos/:id/deactivate`.

- [ ] **TASK-028 — Promo validation & application service**
      Implement `PromoService.Validate(tenantID, userID, code, cart)` checking: active, date window, min order amount, max_usage, per-user limit, eligible_user_ids; return full discount breakdown; wire into order charge pre-calc endpoint and order creation.

---

## PHASE 6 — Order Core

- [ ] **TASK-029 — Orders, order_items, order_pickups, order_timeline tables & SQLC queries**
      Write migration for all four order tables; SQLC queries: `CreateOrder`, `CreateOrderItems`, `CreateOrderPickups`, `GetOrderByID`, `GetOrderByNumber`, `ListOrdersByCustomer`, `ListOrdersByTenant`, `ListOrdersByRestaurant`, `TransitionOrderStatus`, `AddTimelineEvent`.

- [ ] **TASK-030 — Order charge pre-calculation endpoint**
      Implement `POST /api/v1/orders/charges/calculate` (authenticated, tenant-scoped): validate cart items + restaurants open + stock, apply product discounts, validate promo code, calculate delivery charge and VAT, return full charge breakdown without creating an order.

- [ ] **TASK-031 — Atomic order creation (COD & wallet)**
      Implement `POST /api/v1/orders` for COD and wallet payment: DB transaction covering stock reservation (`SELECT FOR UPDATE`), promo usage increment, order+items+pickups insert, timeline entry, outbox event insert; return order ID; enqueue rider-assign job; publish SSE event to restaurant channel.

- [ ] **TASK-032 — Order creation for online payment (bKash / AamarPay)**
      Extend order creation to handle `payment_method = bkash | aamarpay`: create order with `status = PENDING`, initiate payment intent via payment adapter, return payment URL to client; on callback success → transition order to `CREATED`; on callback failure → release stock, soft-delete order.

- [ ] **TASK-033 — Order status transitions (restaurant side)**
      Implement partner API: `PATCH /partner/orders/:id/confirm` (CREATED → CONFIRMED, notify rider), `PATCH /partner/orders/:id/reject` (mandatory reason, trigger refund if online payment, partial multi-restaurant rejection logic), `PATCH /partner/orders/:id/ready` (→ READY, notify rider); all transitions write timeline entry.

- [ ] **TASK-034 — Order status transitions (preparing, auto-confirm, admin force-cancel)**
      Implement `PATCH /partner/orders/:id/preparing` by restaurant; implement auto-confirm job: scan CREATED orders past `auto_confirm_after_minutes` timeout, transition with `actor=system`; implement admin force-cancel endpoint with mandatory reason + timeline entry.

- [ ] **TASK-035 — Customer order tracking & cancellation**
      Implement `GET /api/v1/orders/:id` (full detail), `GET /api/v1/orders/:id/tracking` (SSE stream), `PATCH /api/v1/orders/:id/cancel` (allowed in PENDING/CREATED only, trigger refund + stock release), `GET /api/v1/me/orders` (paginated history).

- [ ] **TASK-036 — Multi-restaurant order (pickup coordination)**
      Ensure pickups maintain independent per-restaurant status; parent order status derived from all pickups (READY when all pickups READY, PICKED when all PICKED); rider API `PATCH /rider/orders/:id/picked/:restaurant_id` per restaurant; partner portal shows only own restaurant items.

---

## PHASE 7 — Payments & Refunds

- [ ] **TASK-037 — Payment transactions & refunds tables, SQLC queries**
      Write migration for `payment_transactions` and `refunds`; SQLC queries: `CreateTransaction`, `UpdateTransactionStatus`, `GetTransactionByGatewayID` (for idempotency), `CreateRefund`, `UpdateRefundStatus`, `ListRefundsByOrder`; gateway_txn_id uniqueness enforced at DB level.

- [ ] **TASK-038 — bKash payment adapter & flow**
      Implement `adapters/bkash/` (tokenized API: grant token, create payment, execute payment, query payment, refund); implement `POST /api/v1/payments/bkash/initiate` and idempotent `GET /api/v1/payments/bkash/callback`; on success transition order; on failure release stock.

- [ ] **TASK-039 — AamarPay / SSLCommerz adapter & flow**
      Implement `adapters/aamarpay/` (payment init, success/fail/cancel callback handlers); same idempotency and stock-release guarantees as bKash; wire SSLCommerz behind feature flag as optional second card gateway.

- [ ] **TASK-040 — Refund engine**
      Implement `RefundService.ProcessRefund(orderID, amount, reason, triggeredBy)`: determine payment method, call gateway refund API or credit wallet, mark manual for COD; create `refunds` record, update `payment_transactions`; enqueue refund notification; ledger entry for REFUND_LIABILITY.

- [ ] **TASK-041 — Payment reconciliation job**
      Implement background job polling `PENDING` orders older than 15 min: query gateway API for actual payment status; auto-confirm or auto-cancel with reason; log reconciliation events in order timeline; send alert if gateway unreachable.

---

## PHASE 8 — Rider Module

- [ ] **TASK-042 — Rider profile, hub assignment & SQLC queries**
      Write migration for `riders`, `rider_locations`, `rider_travel_logs`, `rider_attendance`, `rider_penalties`; SQLC queries for full rider CRUD + location upsert + attendance management + penalty CRUD; partner API: `GET/POST/PUT/DELETE /partner/riders`.

- [ ] **TASK-043 — Rider attendance & availability**
      Implement rider API: `POST /rider/attendance/checkin` (attendance record, is_on_duty=true, hub selection), `POST /rider/attendance/checkout` (close record, calc total_hours + distance, set is_on_duty=false), `PATCH /rider/availability` (online/offline toggle without checkout); partner: `GET /partner/riders/attendance`.

- [ ] **TASK-044 — Rider location (WebSocket)**
      Implement `WS /api/v1/rider/ws`: JWT auth on handshake; handle `location` messages (upsert `rider_locations`, append `rider_travel_logs`, calc distance_from_prev using Haversine); handle `status` messages; broadcast to Redis channel `rider:{rider_id}:location`.

- [ ] **TASK-045 — Auto-assignment algorithm**
      Implement `RiderAssignmentService.AutoAssign(orderID)`: fetch available+on-duty riders in order's hub, calculate distance from `rider_locations` to first pickup using Barikoi API (cached), sort by distance, send assignment push to top 3, 60s acceptance timeout, retry up to 3 batches, alert admin if none accept.

- [ ] **TASK-046 — Rider order flow & manual assignment**
      Implement rider API: `GET /rider/orders/active`, `PATCH /rider/orders/:id/picked/:restaurant_id` (per-pickup, check all-picked → transition parent to PICKED), `PATCH /rider/orders/:id/delivered` (→ DELIVERED, update rider stats, trigger analytics sync), `PATCH /rider/orders/:id/issue`; partner: `POST /partner/orders/:id/assign-rider`.

- [ ] **TASK-047 — Rider earnings, travel log & penalties**
      Implement earnings per order (base + distance bonus + peak-hour bonus per tenant config); accumulate in `riders.balance`; `GET /rider/earnings`, `GET /rider/history`; partner API: `GET /partner/riders/:id/travel-log`, `GET /partner/riders/tracking` (live locations), penalty CRUD at `/partner/riders/:id/penalties`.

---

## PHASE 9 — Finance & Invoicing

- [ ] **TASK-048 — Invoices table, SQLC queries & generation logic**
      Write migration for `invoices`; SQLC queries; implement `InvoiceService.GenerateForRestaurant(restaurantID, periodStart, periodEnd)` with exact formula (total_sales, vat_collected, vendor_promo_discount, commission_amount, penalty, adjustment → net_payable); idempotent via UNIQUE constraint on (restaurant_id, period_start, period_end).

- [ ] **TASK-049 — Invoice management API (partner & admin)**
      Implement partner API: `GET /partner/finance/summary`, `GET /partner/finance/invoices`, `GET /partner/finance/invoices/:id`; admin API: `POST /admin/finance/invoices/generate`, `PATCH /admin/finance/invoices/:id/finalize`, `PATCH /admin/finance/invoices/:id/mark-paid`; all mutations require reason + audit_log entry.

- [ ] **TASK-050 — PDF invoice generation & download**
      Implement invoice PDF renderer (Go HTML-to-PDF or chromium headless); expose `GET /partner/finance/invoices/:id/pdf`; cache generated PDFs in S3/R2 by invoice ID; return with `Content-Disposition: attachment` header.

- [ ] **TASK-051 — Ledger entries (financial auditability)**
      Write migration for `ledger_accounts` and `ledger_entries`; seed platform accounts (CUSTOMER_WALLET, PLATFORM_COMMISSION, VENDOR_PAYABLE, REFUND_LIABILITY); implement append-only `LedgerService.Record(...)` called from order completion, refund, wallet credit, commission calculation; never UPDATE entries.

- [ ] **TASK-052 — Wallet management & cashback**
      Implement `WalletService.Credit(...)` and `Debit(...)` with balance-after tracking in `wallet_transactions`; on order DELIVERED trigger cashback credit for eligible promos; enforce max wallet spend per checkout per tenant config; `GET /api/v1/me/wallet` returns balance + paginated history.

---

## PHASE 10 — Notifications & Real-time

- [ ] **TASK-053 — Firebase FCM adapter & push notifications**
      Implement `adapters/fcm/` using Firebase Admin SDK; implement `NotificationService.SendPush(userID, title, body, data)` loading `user.device_push_token`; handle token-not-registered error (clear from DB); persist all notifications to `notifications` table.

- [ ] **TASK-054 — SMS adapter (OTP & order events)**
      Implement `adapters/sms/` for SSL Wireless BD API; implement `SMSService.Send(phone, template, vars)` with all templates from doc 11 §11.6; Redis-based OTP rate limiting; add Twilio as fallback provider config.

- [ ] **TASK-055 — Email adapter (transactional)**
      Implement `adapters/email/` for SendGrid / AWS SES; HTML email templates for all events in doc 11 §11.7 (welcome, invoice_ready, order_confirmation, refund_processed, password_reset, vendor_invitation, tenant_suspended); `EmailService.Send(to, templateName, vars)`.

- [ ] **TASK-056 — SSE infrastructure (real-time order events)**
      Implement `GET /api/v1/events/subscribe` SSE handler: JWT auth, subscribe to Redis pub/sub channels (tenant-wide, user-specific, restaurant-specific), write SSE format events, support `Last-Event-ID` replay from 5-min Redis cache; goroutine-per-connection with context-cancel on disconnect.

- [ ] **TASK-057 — Outbox event processor worker**
      Implement asynq worker polling `outbox_events` with `status=pending`; for each event: publish to Redis pub/sub (→ SSE routing), trigger push/SMS/email notification; mark `processed` on success; exponential backoff on failure; dead-letter after 5 retries.

---

## PHASE 11 — Background Jobs

- [ ] **TASK-058 — asynq worker setup & job registry**
      Initialize asynq server with Redis connection; register all job handlers in central registry; configure concurrency, retry policies, queue priorities per job type; set up asynq web UI for dev monitoring (`localhost:8081`).

- [ ] **TASK-059 — Scheduled jobs (auto-confirm, auto-cancel, expiry, cleanup)**
      Implement and register: `order:auto_confirm` (every 1 min, CREATED orders past timeout), `order:auto_cancel` (every 5 min, PENDING orders past 30 min), `promo:expiry` (hourly), `product:discount_expiry` (hourly), `notifications:cleanup` (daily, purge > 90 days).

- [ ] **TASK-060 — Invoice generation & analytics sync jobs**
      Implement `invoice:daily_generate` (runs at configurable hour, generates invoices for all active tenant restaurants for yesterday); `order:analytics_sync` (event-driven on terminal order state, writes denormalized row to `order_analytics`); `report:daily` (aggregate daily tenant dashboard cache).

- [ ] **TASK-061 — Data retention cleanup jobs**
      Implement scheduled purge jobs for: `order_timeline` > 1 year, `rider_travel_logs` > 6 months, `search_logs` > 90 days, `audit_logs` > 2 years; soft-deleted record hard-purge after grace period; all jobs log purged row count.

---

## PHASE 12 — Order Issues, Ratings & Search

- [ ] **TASK-062 — Order issues & dispute management**
      Write migration for `order_issues`; implement: `POST /api/v1/orders/:id/issue` (customer create), partner view/respond (`GET/POST /partner/issues/:id/message`), admin resolution: `PATCH /admin/issues/:id/resolve`, `PATCH /admin/issues/:id/refund/approve`, `PATCH /admin/issues/:id/refund/reject`; approved refunds trigger `RefundService`.

- [ ] **TASK-063 — Ratings & reviews**
      Write migration for `ratings`; implement `POST /api/v1/orders/:id/rate` (DELIVERED orders only, one-per-order); update aggregate `restaurants.rating` via job or trigger; `GET /api/v1/restaurants/:slug/ratings` (paginated public reviews); partner can respond to reviews.

- [ ] **TASK-064 — Full-text search & autocomplete**
      Implement `GET /api/v1/search?q=&type=` (restaurants + products using pg_trgm GIN index); implement autocomplete suggestions API (top 5 partial-match results); log all searches to `search_logs`; `GET /partner/reports/searches` (top search terms last 30 days).

---

## PHASE 13 — Content Management

- [ ] **TASK-065 — Banners, stories & homepage sections**
      Write migration for `banners`, `stories`, `sections`; implement partner API: `GET/POST/PUT/DELETE /partner/content/banners`, `GET/POST/DELETE /partner/content/stories`, `GET/PUT /partner/content/sections`; storefront API: `GET /api/v1/storefront/banners` (area-filtered, active), `GET /api/v1/storefront/stories`, `GET /api/v1/storefront/sections`.

- [ ] **TASK-066 — Media upload service (S3 / Cloudflare R2)**
      Implement `POST /api/v1/media/upload` (multipart form): validate MIME type (jpg/png/webp/gif/mp4), enforce max size per type, generate UUID filename, upload to S3/R2, return CDN URL; implement `DELETE /api/v1/media/:key`; configure signed upload URLs for large files.

---

## PHASE 14 — Analytics & Reporting

- [ ] **TASK-067 — Partner dashboard KPIs & trend API**
      Implement `GET /partner/dashboard`: today's order counts (all statuses), today's revenue, pending count, avg delivery time; last 7-day orders+revenue trend from `order_analytics`; top 3 products last 7 days; cache per tenant for 2 min in Redis.

- [ ] **TASK-068 — Partner sales report API**
      Implement `GET /partner/reports/sales` (date range, restaurant filter, group by day/week/month): gross sales, item discounts, promo discounts, net sales, VAT, commission, net payable, order count, avg order value, avg delivery time; sourced from `order_analytics`.

- [ ] **TASK-069 — Partner product, peak-hour & rider analytics API**
      Implement `GET /partner/reports/products/top-selling`, `GET /partner/reports/orders/breakdown` (status counts), `GET /partner/reports/peak-hours` (hour × count heatmap data), `GET /partner/reports/riders` (per-rider orders/distance/avg-time); all with optional CSV export.

- [ ] **TASK-070 — Admin cross-tenant analytics API**
      Implement `GET /admin/analytics/overview` (platform totals across all tenants), `GET /admin/analytics/revenue` (commission by period), `GET /admin/analytics/orders` (volume with growth), `GET /admin/analytics/tenants/:id` (tenant deep-dive); super-admin role required.

---

## PHASE 15 — Super Admin Panel (Next.js)

- [x] **TASK-071 — Admin portal project setup**
      Initialise `admin/` Next.js 14 app (TypeScript, Tailwind, shadcn/ui, TanStack Query, Zustand, React Hook Form + Zod); configure API client (base URL, auth interceptor, token refresh, error handling); add `X-Request-ID` header to all requests; ESLint + Prettier.

- [x] **TASK-072 — Admin auth (login, 2FA, session management)**
      Build login page (email + password); mandatory 2FA TOTP flow (QR code setup on first login, TOTP verification on subsequent logins); JWT stored in httpOnly cookie; Zustand auth store; protect all routes via server-side session check; logout + session expiry handling.

- [x] **TASK-073 — Admin platform dashboard page**
      Build dashboard: global KPI cards (total orders today, total commission, active tenants, active riders); revenue trend chart (Recharts); active tenant table with status indicators; system health section (API latency + error rate fetched from `/admin/analytics/overview`).

- [x] **TASK-074 — Tenant management pages (list, create, edit, suspend)**
      Build tenant list (sortable table: name, plan, status, order count, commission rate, date); tenant create form (all fields including plan, commission, slug); tenant detail/edit page; suspend/reinstate modal with reason; tenant analytics drill-down; impersonate button (opens partner portal in new tab).

- [x] **TASK-075 — User management page (cross-tenant)**
      Build user search page (search by phone/email/name with tenant filter); user detail drawer (profile, order summary, account status); suspend user action with reason; GDPR delete flow with confirmation and data-wipe status.

- [x] **TASK-076 — Admin order management (cross-tenant)**
      Build orders page with tenant + status + date filters; order detail modal with full timeline, payment events, rider history, and audit entries; force status override action (dropdown + mandatory reason); link to issue resolution.

- [x] **TASK-077 — Finance management (commissions, invoices, payouts)**
      Build commission ledger page (all tenants, period filter, CSV export); invoice list with generate-invoice action; invoice detail with approve/finalize/mark-paid actions; payout tracking table with settlement status.

- [x] **TASK-078 — Issues & dispute resolution page**
      Build issues queue (filter by status/tenant/type); issue detail page with full order context, message thread, refund amount input, accountable party selector, approve/reject buttons; resolved issues history.

- [x] **TASK-079 — Platform config & feature flags page**
      Build settings page: global feature toggles, payment gateway config (credentials, test/live toggle), SMS/email provider config, maintenance mode toggle with confirmation; all changes create audit log.

---

## PHASE 16 — Partner Portal (Next.js)

- [x] **TASK-080 — Partner portal project setup**
      Initialise `partner/` Next.js 14 app (same stack as admin); API client with tenant-aware base URL; Zustand auth store; protected route layout; notification bell (polling unread count); new-order audio notification setup.

- [x] **TASK-081 — Partner auth (login, password reset, invite flow)**
      Build login page (email+password); forgot/reset password flow; invitation acceptance page (set password from invite token); multi-restaurant context: on login with multiple restaurants show restaurant picker.

- [x] **TASK-082 — Partner dashboard (live order board)**
      Build dashboard: KPI cards (today's orders, revenue, pending count, avg delivery time); **live incoming order panel** (SSE-connected, sound + visual badge on new order, accept/reject buttons with 3-min countdown timer); 7-day trend charts; quick-action buttons (toggle restaurant, view pending issues).

- [x] **TASK-083 — Restaurant management pages**
      Build restaurant list (cards with availability toggle); restaurant create/edit form (all fields: name, description, cuisines, address, images, operating hours day-by-day scheduler, VAT rate, prep time); branch switcher in nav sidebar scoping all views to selected restaurant.

- [x] **TASK-084 — Menu management (categories, products, drag-drop reorder)**
      Build menu page: left-panel category list with dnd-kit drag-drop reorder; right-panel product grid per category; product create/edit sheet (all fields: name, images, price type, variant builder, addon builder, discount toggle); availability toggle per product/category; bulk-upload CSV modal.

- [x] **TASK-085 — Live order management board**
      Build orders page: kanban columns (New → Confirmed → Preparing → Ready → Picked); cards show order number, items summary, time elapsed, customer area; action buttons per status; order detail drawer (items+addons, customer info, rider info, address map, payment, timeline); order history table with search + filters.

- [x] **TASK-086 — Rider management pages**
      Build rider list (table: name, hub, status badge, location, today's orders); rider create/edit form; rider detail page (stats, attendance, live location map, travel-log playback, earnings, penalties); attendance calendar; availability toggle.

- [x] **TASK-087 — Promotions management pages**
      Build promo list (table: code, type, usage count, status); promo create/edit form (all fields: code, type, amount, cap, apply_on, restrictions, date range, per-user limit, cashback amount); promo performance stats (usage, total discount given, unique users).

- [x] **TASK-088 — Finance pages (invoices, commission, payouts)**
      Build finance summary (current period net payable, YTD totals); invoice list with status badges; invoice detail page (full breakdown from doc 08 §8.6); PDF download; payment history; outstanding balance alert.

- [x] **TASK-089 — Sales & analytics pages**
      Build reports section: sales report (date range picker, grouped by day/week/month, all metrics from doc 12, CSV export); top-selling products table; peak hours heatmap; order status breakdown chart; rider performance table; customer area distribution.

- [x] **TASK-090 — Content management pages (banners, stories, sections)**
      Build banners page (image upload, link type/value, area targeting, validity dates, sort-order drag-drop); stories page (media upload, expiry, restaurant link); homepage sections editor (type, add restaurants/products, display order).

- [x] **TASK-091 — Partner settings & team management**
      Build settings page (vendor profile, notification preferences per event type); team management (list members with roles, invite by email modal, role selector, remove member); API keys placeholder (Phase 2).

---

## PHASE 17 — Customer Website (Next.js)

- [x] **TASK-092 — Website project setup (SSR-first, SEO-optimized)**
      Initialise `website/` Next.js 14 app (App Router, TypeScript, Tailwind, TanStack Query, Zustand); configure tenant resolution for SSR (read subdomain in server components, pass config to client); add next-seo for metadata; Cloudflare CDN headers; structured data baseline.

- [x] **TASK-093 — Customer auth (phone OTP flow)**
      Build phone OTP modal (phone input → OTP input → success); handle first-time registration (name prompt after OTP verify); JWT stored in httpOnly cookie via Next.js route handler; Zustand auth state; protected-route HOC for checkout/account pages.

- [x] **TASK-094 — Homepage (banners, sections, restaurant listing)**
      Build homepage: hero banner carousel (SSR-fetched, lazy); story strip (horizontal scroll); cuisine filter pills; area selector modal (GPS auto-detect + manual area pick from zone list); restaurant grid (infinite scroll, open/closed badge, rating, delivery time, offer tags); sort controls.

- [x] **TASK-095 — Restaurant page (menu with sticky category nav)**
      Build restaurant page (SSR for SEO): restaurant header (banner, logo, info, cuisines, hours); sticky horizontal category nav (click scrolls, scroll updates active tab via IntersectionObserver); product grid (images, prices, discount badge, availability); search-within-restaurant input.

- [x] **TASK-096 — Product detail modal (variants, addons, quantity)**
      Build product modal/page: image gallery; description; variant group selector (radio or checkbox based on min/max_select); addon group selector; quantity stepper; real-time price calculation (base + variants + addons × qty); "Add to Cart" button with validation.

- [x] **TASK-097 — Cart (persistent, multi-restaurant)**
      Build cart drawer/page: items grouped by restaurant; quantity change + remove; price summary (subtotal, item discounts, promo + delivery placeholders); persist in `localStorage`; sync to session on auth; clear on order success; multi-restaurant info banner.

- [x] **TASK-098 — Checkout page (address, promo, payment)**
      Build checkout page (protected): delivery address selector (saved + add-new with Barikoi map picker); promo code input (live API validation, discount preview); payment method radio (COD, bKash, card); estimated delivery time; full charge breakdown; "Place Order" with loading state + client-side idempotency key.

- [x] **TASK-099 — Payment redirect flows (bKash & AamarPay)**
      Handle bKash redirect (open bKash URL, poll order status on return), AamarPay standard redirect (success/fail/cancel callback pages), wallet instant payment; build success page (tracking link), fail page (retry option), cancel page; clear PENDING state correctly.

- [x] **TASK-100 — Order tracking page (SSE real-time)**
      Build active order tracking: step-progress-bar (Pending → Confirmed → Preparing → Ready → Picked → Delivered) updated via SSE; order details card; rider card (appears when PICKED — name, phone, live location map via polling); cancel button (shown only when cancellable); ETA countdown.

- [x] **TASK-101 — Account pages (profile, addresses, history, wallet, favourites)**
      Build account section: profile edit form; saved addresses CRUD with map picker; order history (filterable, reorder button, detail modal); wallet page (balance card, paginated transactions); favourites grid (heart toggle, quick-remove); notification center list.

---

## PHASE 18 — Rider PWA

- [ ] **TASK-102 — Rider PWA setup (Next.js + next-pwa)**
      Initialise `rider/` Next.js 14 app configured as PWA (next-pwa manifest, service worker, installable on Android); mobile-first Tailwind layout; phone OTP auth; push notification opt-in on first login; register FCM service worker for background push.

- [ ] **TASK-103 — Rider duty management screens**
      Build home screen (off-duty state): hub selector; check-in button; availability toggle (online/offline without checkout); check-out button; today's summary (orders, earnings, distance); profile link.

- [ ] **TASK-104 — Rider active order screen & delivery flow**
      Build active order screen: restaurants list with addresses + items summary; customer address + area; "Navigate to Restaurant" Google Maps deep-link; "Mark as Picked" per restaurant; "Navigate to Customer" link; "Mark as Delivered" button; "Report Issue" button.

- [ ] **TASK-105 — Rider earnings, history & profile screens**
      Build earnings page (today's summary, balance card, weekly earnings chart); delivery history (paginated: date, order number, earning, status); profile page (edit name/avatar/vehicle type, view penalties with appeal-note input).

---

## PHASE 19 — Security Hardening

- [ ] **TASK-106 — PostgreSQL Row-Level Security (RLS) policies**
      Enable RLS on all tenant-scoped tables per doc 03 §3.10; create `SET LOCAL app.tenant_id` session helper; write USING+WITH CHECK policies per table; create platform-role bypass; write integration tests confirming cross-tenant queries return 0 rows without application-level filters.

- [ ] **TASK-107 — Rate limiting hardening (Redis token bucket)**
      Implement Redis-backed rate limiter for all endpoint groups per doc 10 §10.3; add `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `Retry-After` response headers; composite key `(IP, user_id, endpoint)`; test OTP and order endpoints at burst limits.

- [ ] **TASK-108 — Fraud & abuse controls**
      Implement velocity check service (OTP abuse, promo abuse detection — shared device/payment token, COD risk scoring based on cancel rate); implement blocklist framework (`blocked_identifiers` table for phone/IP/device_fingerprint/payment_token); admin UI to manage blocklist.

- [ ] **TASK-109 — Security headers, CORS hardening & input sanitization**
      Add security headers middleware (HSTS, X-Frame-Options, X-Content-Type-Options, CSP, Referrer-Policy); tighten CORS to exact origin allowlist; audit all JSONB inputs; add MIME-type + size validation on file uploads; strip HTML from all free-text inputs.

- [ ] **TASK-110 — Audit log coverage & sensitive action guards**
      Verify `audit_logs` written for: all admin mutations, all partner mutations (refund, force cancel, rider penalty), all auth events (login, logout, OTP send, password reset); add `GuardSensitiveAction` middleware requiring `X-Action-Reason` header on designated high-risk endpoints.

---

## PHASE 20 — Infrastructure & DevOps

- [ ] **TASK-111 — Production Docker images (multi-stage, minimal)**
      Write multi-stage `Dockerfile` per app (Go: build → distroless/alpine < 50MB; Next.js: standalone output mode); add `.dockerignore`; tag with Git SHA; push to container registry in CI.

- [ ] **TASK-112 — Nginx reverse proxy (subdomain routing, SSL)**
      Write `nginx.conf`: upstreams for API + Next.js apps; subdomain routing (`*.platform.com` → website, `partner.*` → partner, `admin.*` → admin, `api.*` → Go API); SSL termination with Let's Encrypt; gzip; WebSocket proxy headers; nginx-level rate limiting.

- [ ] **TASK-113 — GitHub Actions CI/CD pipelines**
      Write workflows: `backend.yml` (test → lint → build → push image → rolling deploy on main), `website.yml`, `partner.yml`, `admin.yml`, `rider.yml` (install → lint → build → deploy); OIDC credentials (no long-lived secrets in CI); PR pipeline (tests only).

- [ ] **TASK-114 — Secrets management (Vault / AWS Secrets Manager)**
      Migrate all production secrets to secrets manager; configure Go app to load secrets at startup via SDK; document secret rotation runbook for DB password + JWT secrets without downtime.

- [ ] **TASK-115 — Prometheus metrics & Grafana dashboards**
      Instrument Go API with `promhttp`; add custom metrics: request duration histogram, orders_created_total, orders_failed_total, rider_assignment_duration, payment_callback_latency, queue depths; import Go + PostgreSQL Grafana dashboards; create custom SLO dashboard.

- [ ] **TASK-116 — Loki log aggregation & Sentry error tracking**
      Configure Promtail/Fluentd → Loki; set up Grafana Loki data source; create log dashboards for order errors and payment failures; integrate Sentry SDK into Go API (with user/tenant context) and all Next.js apps.

- [ ] **TASK-117 — PostgreSQL backup automation & WAL archiving**
      Set up automated daily pg_dump to S3/R2 (30-day retention); enable WAL archiving for PITR (15-min RPO); Redis RDB + AOF backup every 6 hours; object storage versioning; backup health-check cron alerting if last backup > 25 hours old.

- [ ] **TASK-118 — Container orchestration manifests (K8s / Docker Swarm)**
      Write deployment manifests for: API (2+ replicas, resource limits, liveness + readiness probes), worker (1 replica), PostgreSQL (statefulset + PVC), Redis (sentinel mode); horizontal pod autoscaler for API tier; document rollback procedure.

- [ ] **TASK-119 — UptimeRobot + public status page**
      Configure UptimeRobot monitors for API `/healthz`, website homepage, partner portal, admin panel; set alert contacts (email + Slack); set up `status.platform.com` page with incident history; write incident communication templates for P1/P2 outages.

---

## PHASE 21 — Testing & Quality Assurance

- [ ] **TASK-120 — Backend unit tests (business logic layer)**
      Write unit tests for all service-layer logic: order charge calculation (all combinations: variants/addons/promos/VAT/delivery), promo validation rules, invoice calculation formula, rider assignment scoring, delivery charge model; use `testify` + table-driven tests; 80% coverage target on service files.

- [ ] **TASK-121 — Backend integration tests (API + DB)**
      Write integration tests against real PostgreSQL test DB (Docker in CI): auth flows, order creation happy path, cancellation + refund, multi-restaurant pickup, tenant isolation (cross-tenant must return 0 rows), RLS enforcement, idempotency key deduplication, payment callback idempotency.

- [ ] **TASK-122 — Frontend component & E2E tests**
      Write Vitest unit tests for critical components (cart price calc, promo input, order status steps); set up Playwright E2E for critical paths: (1) customer OTP register → COD order → delivered status, (2) partner login → accept order → mark ready, (3) rider picks up → marks delivered.

- [ ] **TASK-123 — Load testing (order placement under concurrency)**
      Write k6 load test scripts: (1) storefront homepage 200 VU 5 min, (2) 50 concurrent order placements (validate no oversell + no duplicate orders), (3) 100 concurrent SSE connections; run against staging; verify p95 latency within SLO targets from doc 13 §13.10.

---

## PHASE 22 — Launch Readiness

- [ ] **TASK-124 — Disaster recovery drill**
      Simulate DB primary failure: verify read-replica promotion time; restore from last backup to clean instance and verify data integrity; document actual RPO/RTO achieved; sign off and store report in ops runbook.

- [ ] **TASK-125 — Production environment checklist run-through**
      Execute doc 15 (World-Class Quality Gates) checklist end-to-end: all release gates, security gates, reliability gates, financial integrity gates, tenant safety gates, performance gates; resolve all failing items; get sign-off before go-live.

- [ ] **TASK-126 — Smoke test suite (automated post-deploy)**
      Write smoke test script (shell or Go) running after every production deploy: (1) health check, (2) OTP send to test phone, (3) catalog API returns tenant data, (4) order pre-calculation returns valid totals, (5) SSE endpoint connects + receives heartbeat, (6) invoice endpoint reachable; alert on any failure within 5 min.

- [ ] **TASK-127 — Tenant onboarding runbook & first-vendor activation**
      Write step-by-step onboarding runbook: super-admin creates tenant → sets commission/plan → invites vendor → vendor sets up restaurant+menu → admin activates → subdomain DNS verified → smoke test storefront; onboard first real vendor end-to-end as production validation.

- [ ] **TASK-128 — Documentation polish & API reference generation**
      Generate OpenAPI 3.0 spec from route map (doc 10); publish to `docs.platform.com`; write vendor onboarding guide, partner portal user guide, and API quickstart; update all `docs/requirements/` with any implementation deviations discovered during build.

---

## Task Count Summary

| Phase     | Tasks   | Scope                              |
| --------- | ------- | ---------------------------------- |
| 0         | 7       | Repository Foundation              |
| 1         | 8       | Multi-Tenancy & Authentication     |
| 2         | 2       | Delivery Infrastructure            |
| 3         | 8       | Restaurant & Catalog               |
| 4         | 1       | Inventory                          |
| 5         | 2       | Promotions Engine                  |
| 6         | 8       | Order Core                         |
| 7         | 5       | Payments & Refunds                 |
| 8         | 6       | Rider Module                       |
| 9         | 5       | Finance & Invoicing                |
| 10        | 5       | Notifications & Real-time          |
| 11        | 4       | Background Jobs                    |
| 12        | 3       | Issues, Ratings & Search           |
| 13        | 2       | Content Management                 |
| 14        | 4       | Analytics & Reporting              |
| 15        | 9       | Admin Panel (Next.js)              |
| 16        | 12      | Partner Portal (Next.js)           |
| 17        | 10      | Customer Website (Next.js)         |
| 18        | 4       | Rider PWA                          |
| 19        | 5       | Security Hardening                 |
| 20        | 9       | Infrastructure & DevOps            |
| 21        | 4       | Testing & QA                       |
| 22        | 5       | Launch Readiness                   |
| **Total** | **128** | **Full production-ready platform** |
