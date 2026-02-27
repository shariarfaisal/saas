# 06 — Portals & User Roles

## 6.1 Portal Overview

| Portal | Path | Users | Tech | Purpose |
|--------|------|-------|------|---------|
| Customer Website | `website/` | Customers | Next.js (SSR/SSG) | Order food online |
| Partner Portal | `partner/` | Vendor owners, managers, staff | Next.js (SPA) | Manage restaurants, orders, reports |
| Super Admin | `admin/` | Platform staff | Next.js (SPA) | Manage all tenants and platform |
| Rider App (PWA) | `website/` (sub-route) or `partner/` | Delivery riders | Next.js (PWA) | Manage deliveries |

---

## 6.2 Customer Website (`website/`)

**URL pattern:** `[tenant-slug].platform.com`  
**Custom domain (Phase 3):** `order.kacchbhai.com` → maps to tenant via `custom_domain`

### Page Structure
```
/                            → Homepage (banners, sections, restaurant list)
/restaurants                 → All restaurants (filterable, paginated)
/restaurants/[slug]          → Restaurant menu page
/restaurants/[slug]/[product-slug]  → Product detail page
/cart                        → Cart review page
/checkout                    → Checkout page (address, payment, promo)
/orders/[id]                 → Active order tracking page
/orders                      → Order history
/account                     → Profile page
/account/addresses           → Saved addresses
/account/wallet              → Wallet / loyalty balance
/account/favourites          → Favourite restaurants
/auth/login                  → Login (phone OTP)
/auth/register               → Register
/search?q=...                → Search results
```

### SSR vs CSR Strategy
| Page | Strategy | Reason |
|------|----------|--------|
| Homepage | SSR + ISR (revalidate 60s) | SEO + fresh content |
| Restaurant page | SSR + ISR | SEO crucial for restaurant pages |
| Product page | SSR + ISR | SEO |
| Cart / Checkout | CSR | Personalised, no SEO value |
| Order tracking | CSR + SSE | Real-time, no SEO |
| Account pages | CSR | Authenticated, no SEO |
| Search results | SSR | SEO benefit for search indexing |

### Customer Session
- Short-lived JWT in httpOnly cookie
- Refresh token in httpOnly cookie (7 days)
- Cart stored in localStorage (synced on login)
- Phone OTP authentication (no passwords for customers)

---

## 6.3 Partner Portal (`partner/`)

**URL:** `partner.platform.com` (single URL, tenant resolved from JWT)

All pages are client-side rendered (no SSR needed — behind auth).

### Page Structure
```
/auth/login                  → Email + password login
/auth/forgot-password        → Password reset

/ (dashboard)                → Overview KPIs + live order feed

/orders                      → Order list (all restaurants)
/orders/[id]                 → Order detail + actions

/restaurants                 → Restaurant list
/restaurants/new             → Create restaurant
/restaurants/[id]            → Restaurant settings
/restaurants/[id]/menu       → Menu management
/restaurants/[id]/categories → Category management
/restaurants/[id]/hours      → Operating hours
/restaurants/[id]/reports    → Restaurant-specific sales report

/riders                      → Rider management
/riders/new                  → Register new rider
/riders/[id]                 → Rider profile + history
/riders/attendance           → Attendance records
/riders/tracking             → Live rider location map

/promos                      → Promo list
/promos/new                  → Create promo
/promos/[id]                 → Edit promo + usage stats

/inventory                   → Stock management
/inventory/adjust            → Stock adjustment form

/reports                     → Sales & analytics reports
/reports/sales               → Revenue charts
/reports/products            → Top selling products
/reports/riders              → Rider performance

/finance                     → Financial overview
/finance/invoices            → Invoice list
/finance/invoices/[id]       → Invoice detail + PDF download

/content                     → Banners, stories, sections
/content/banners             → Manage banners
/content/stories             → Manage stories
/content/sections            → Manage homepage sections

/settings                    → Tenant settings
/settings/team               → Team members and roles
/settings/notifications      → Notification preferences
/settings/profile            → Vendor profile

/issues                      → Order disputes
/issues/[id]                 → Issue detail + resolution
```

### Real-time in Partner Portal
- **SSE connection** on `/` (dashboard) and `/orders` — receives events for:
  - `order.new` — new order arrives
  - `order.status_changed` — any order status update
  - `notification.new` — platform notification
- Browser sound alert on `order.new`
- Badge counter on orders nav item

### Partner Portal Auth
- Email + password (JWT)
- Refresh token in httpOnly cookie
- Role-based menu visibility (owner sees finance, manager doesn't)

---

## 6.4 Super Admin Panel (`admin/`)

**URL:** `admin.platform.com`

### Page Structure
```
/auth/login                  → Admin login (email + TOTP 2FA)

/ (dashboard)                → Global platform KPIs
/tenants                     → Tenant list
/tenants/new                 → Create tenant + owner
/tenants/[id]                → Tenant settings, config, stats
/tenants/[id]/impersonate    → Impersonate as tenant

/orders                      → All orders (cross-tenant)
/orders/[id]                 → Order detail + master actions

/users                       → All users (cross-tenant)
/users/[id]                  → User profile + actions

/finance                     → Platform revenue
/finance/commissions         → Commission ledger
/finance/invoices            → All invoices (cross-tenant)
/finance/payouts             → Settlement status

/analytics                   → Platform-wide analytics
/analytics/revenue           → Revenue charts
/analytics/orders            → Order trends
/analytics/geographic        → Order map

/content                     → Platform-wide banners/announcements

/config                      → Platform configuration
/config/payment-gateways     → Gateway credentials
/config/sms-email            → Messaging providers
/config/features             → Global feature flags
/config/maintenance          → Maintenance mode

/issues                      → All disputes (cross-tenant)
/issues/[id]                 → Issue resolution

/admins                      → Platform admin user management
```

---

## 6.5 Rider PWA

**URL:** `[tenant-slug].platform.com/rider` (sub-path of customer website, separate Next.js layout)

Since riders are within a tenant's user pool, the rider app resolves tenant from the domain (same as customer website).

### Page Structure
```
/rider/login                 → Phone OTP login
/rider                       → Dashboard (shift summary, available orders)
/rider/attendance            → Check in / check out
/rider/order/[id]            → Active order detail + actions
/rider/history               → Past deliveries
/rider/earnings              → Earnings summary
/rider/profile               → Rider profile
```

### PWA Features
- `manifest.json` for Add to Home Screen
- Service worker for offline capability (order detail cached)
- Background push notifications (Firebase FCM)
- Geolocation API for real-time location sharing

---

## 6.6 Permission Matrix

| Feature | customer | rider | restaurant_staff | restaurant_manager | tenant_admin | tenant_owner | platform_admin |
|---------|----------|-------|-----------------|-------------------|--------------|--------------|----------------|
| Place order | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| View own orders | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ |
| Accept/reject order | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Manage menu | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ |
| View sales report | ❌ | ❌ | ❌ | ✅ (own) | ✅ | ✅ | ✅ |
| Manage riders | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ |
| View invoices | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ |
| Manage promos | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ |
| Manage team | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ |
| Tenant config | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ |
| Manage tenants | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Platform finance | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |

---

## 6.7 Notification Triggers Per Role

| Event | customer | rider | restaurant_manager | tenant_admin | platform_admin |
|-------|----------|-------|-------------------|--------------|----------------|
| New order placed | Email/Push | — | Push + SSE | — | — |
| Order confirmed | Push | — | — | — | — |
| Order picked | Push | — | Push | — | — |
| Order delivered | Push | — | — | — | — |
| Order rejected | Push | — | — | — | — |
| New invoice | — | — | Push + Email | Email | — |
| Low stock | — | — | Push | Push | — |
| Rider assigned | Push | Push | — | — | — |
| Order issue raised | — | — | Push | Push | Push |
| Refund processed | Push + Email | — | — | — | — |
