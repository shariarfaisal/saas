# Change: Phase 16 Partner Portal (TASK-080 through TASK-091)

## Why
The platform backend exposes tenant-scoped partner APIs for restaurant, order, rider, finance, and content management, but no vendor-facing frontend exists. Phase 16 creates a dedicated Next.js partner portal that allows restaurant owners, managers, and staff to manage their restaurants, process orders in real time, handle riders, promotions, finances, analytics, and content—all within a multi-tenant, multi-restaurant context.

## What Changes
- **TASK-080** — Initialise `partner/` Next.js 14 app (same stack as admin): TypeScript, Tailwind, ESLint, Prettier, TanStack Query, Zustand, React Hook Form + Zod, Axios API client with tenant-aware base URL, protected route layout, notification bell (polling unread count), new-order audio notification setup.
- **TASK-081** — Build partner auth: login page (email+password), forgot/reset password flow, invitation acceptance page (set password from invite token), multi-restaurant context (on login with multiple restaurants show restaurant picker), Zustand auth store with restaurant context, middleware-based route protection.
- **TASK-082** — Build partner dashboard: KPI cards (today's orders, revenue, pending count, avg delivery time), live incoming order panel (SSE-connected, sound + visual badge on new order, accept/reject buttons with 3-min countdown timer), 7-day trend charts, quick-action buttons (toggle restaurant, view pending issues).
- **TASK-083** — Build restaurant management pages: restaurant list (cards with availability toggle), restaurant create/edit form (all fields: name, description, cuisines, address, images, operating hours day-by-day scheduler, VAT rate, prep time), branch switcher in nav sidebar scoping all views to selected restaurant.
- **TASK-084** — Build menu management: left-panel category list with drag-drop reorder, right-panel product grid per category, product create/edit sheet (all fields: name, images, price type, variant builder, addon builder, discount toggle), availability toggle per product/category, bulk-upload CSV modal.
- **TASK-085** — Build live order management board: kanban columns (New → Confirmed → Preparing → Ready → Picked), cards show order number, items summary, time elapsed, customer area, action buttons per status, order detail drawer (items+addons, customer info, rider info, address map, payment, timeline), order history table with search + filters.
- **TASK-086** — Build rider management pages: rider list (table: name, hub, status badge, location, today's orders), rider create/edit form, rider detail page (stats, attendance, live location map, travel-log playback, earnings, penalties), attendance calendar, availability toggle.
- **TASK-087** — Build promotions management pages: promo list (table: code, type, usage count, status), promo create/edit form (all fields: code, type, amount, cap, apply_on, restrictions, date range, per-user limit, cashback amount), promo performance stats (usage, total discount given, unique users).
- **TASK-088** — Build finance pages: finance summary (current period net payable, YTD totals), invoice list with status badges, invoice detail page (full breakdown), PDF download, payment history, outstanding balance alert.
- **TASK-089** — Build sales & analytics pages: sales report (date range picker, grouped by day/week/month, CSV export), top-selling products table, peak hours heatmap, order status breakdown chart, rider performance table, customer area distribution.
- **TASK-090** — Build content management pages: banners page (image upload, link type/value, area targeting, validity dates, sort-order drag-drop), stories page (media upload, expiry, restaurant link), homepage sections editor (type, add restaurants/products, display order).
- **TASK-091** — Build partner settings & team management: settings page (vendor profile, notification preferences per event type), team management (list members with roles, invite by email modal, role selector, remove member), API keys placeholder (Phase 2).

## Impact
- Affected specs: partner-portal
- Affected code: `partner/` (new Next.js app), `TASKS.md` (Phase 16 task status)
- Dependencies added (npm): Next.js 15, React 18, TanStack Query, Zustand, React Hook Form, Zod, Axios, Recharts, Tailwind, clsx, tailwind-merge, class-variance-authority, Lucide React
- Security impact: partner session enforcement with cookie-based auth, tenant-scoped API calls, restaurant-level access isolation, invitation token handling
- No backend API contracts are changed by this proposal; all endpoints already exist under `/api/v1/partner/`
