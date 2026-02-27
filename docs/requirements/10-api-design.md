# 10 — API Design

## 10.1 Conventions

### Base URL
```
https://api.platform.com/api/v1
```

### Tenant Routing
- Customer-facing APIs: tenant resolved from `Host` header (subdomain)
- Partner/Admin APIs: tenant resolved from `Authorization` JWT claim (`tenant_id`)
- Super-admin APIs: no tenant scope; optional `?tenant_id=` filter

### Authentication
All protected endpoints require:
```
Authorization: Bearer <access_token>
```

### Request / Response Format
- Content-Type: `application/json`
- All IDs: UUID strings
- All timestamps: ISO 8601 UTC (`"2024-01-15T10:30:00Z"`)
- All monetary amounts: string decimal (`"150.00"`) to avoid float precision issues

### Pagination
```json
{
  "data": [...],
  "meta": {
    "total": 245,
    "page": 1,
    "per_page": 20,
    "next_cursor": "eyJpZCI6IjE..."
  }
}
```

### Error Format
```json
{
  "error": {
    "code": "PRODUCT_NOT_FOUND",
    "message": "The requested product does not exist",
    "details": {}
  }
}
```

### Standard HTTP Status Codes
| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created |
| 204 | No Content (delete) |
| 400 | Bad Request (validation error) |
| 401 | Unauthorized (no/invalid token) |
| 403 | Forbidden (insufficient permissions) |
| 404 | Not Found |
| 409 | Conflict (duplicate, out of stock) |
| 422 | Unprocessable Entity (business logic error) |
| 429 | Rate Limited |
| 500 | Internal Server Error |

---

## 10.2 API Route Map

### Auth
```
POST   /api/v1/auth/otp/send              Send OTP to phone
POST   /api/v1/auth/otp/verify            Verify OTP, return tokens (customer)
POST   /api/v1/auth/login                 Email+password login (partner/admin)
POST   /api/v1/auth/refresh               Refresh access token
POST   /api/v1/auth/logout                Invalidate refresh token
POST   /api/v1/auth/password/reset-request  Request password reset email
POST   /api/v1/auth/password/reset        Confirm password reset
```

### Customer — Users
```
GET    /api/v1/me                         Get current user profile
PATCH  /api/v1/me                         Update profile
GET    /api/v1/me/addresses               List saved addresses
POST   /api/v1/me/addresses               Add address
PUT    /api/v1/me/addresses/:id           Update address
DELETE /api/v1/me/addresses/:id           Delete address
GET    /api/v1/me/orders                  Order history
GET    /api/v1/me/favourites              Favourite restaurants
POST   /api/v1/me/favourites/:restaurant_id  Add to favourites
DELETE /api/v1/me/favourites/:restaurant_id  Remove from favourites
GET    /api/v1/me/wallet                  Wallet balance + transactions
GET    /api/v1/me/notifications           User notifications
PATCH  /api/v1/me/notifications/:id/read  Mark notification read
```

### Storefront — Homepage
```
GET    /api/v1/storefront/config          Tenant config (theme, name, logo)
GET    /api/v1/storefront/banners         Active banners for area
GET    /api/v1/storefront/stories         Active stories
GET    /api/v1/storefront/sections        Homepage sections with items
GET    /api/v1/storefront/cuisines        Available cuisine filters
GET    /api/v1/storefront/restaurants     Restaurant listing (with filters)
GET    /api/v1/storefront/areas           Delivery zone areas list
```

### Storefront — Restaurants & Products
```
GET    /api/v1/restaurants/:slug          Restaurant page (menu, info)
GET    /api/v1/restaurants/:slug/products Products with categories
GET    /api/v1/products/:id              Product detail
GET    /api/v1/search?q=&type=           Search (restaurants + products)
```

### Orders
```
POST   /api/v1/orders                    Place order
GET    /api/v1/orders/:id                Order detail (customer)
PATCH  /api/v1/orders/:id/cancel         Customer cancel order
POST   /api/v1/orders/:id/rate           Submit rating after delivery
GET    /api/v1/orders/:id/tracking       SSE stream for real-time updates
```

### Payments
```
POST   /api/v1/payments/bkash/initiate      Initiate bKash payment
GET    /api/v1/payments/bkash/callback      bKash callback (webhook)
POST   /api/v1/payments/aamarpay/initiate   Initiate AamarPay
POST   /api/v1/payments/aamarpay/success    AamarPay success callback
POST   /api/v1/payments/aamarpay/fail       AamarPay fail callback
POST   /api/v1/payments/aamarpay/cancel     AamarPay cancel callback
POST   /api/v1/payments/bkash/wallet/save   Save bKash wallet (agreement)
```

### Order Charges (pre-order calculation)
```
POST   /api/v1/orders/charges/calculate     Calculate order total before placing
```

---

### Partner API (prefix: `/api/v1/partner/`)

```
# Dashboard
GET    /partner/dashboard                   KPIs + quick stats

# Restaurants
GET    /partner/restaurants                 List managed restaurants
POST   /partner/restaurants                 Create restaurant
GET    /partner/restaurants/:id             Restaurant detail
PUT    /partner/restaurants/:id             Update restaurant
PATCH  /partner/restaurants/:id/availability  Toggle open/close
DELETE /partner/restaurants/:id             Deactivate restaurant

# Operating Hours
GET    /partner/restaurants/:id/hours
PUT    /partner/restaurants/:id/hours

# Categories
GET    /partner/restaurants/:id/categories
POST   /partner/restaurants/:id/categories
PUT    /partner/restaurants/:id/categories/:cat_id
DELETE /partner/restaurants/:id/categories/:cat_id
PATCH  /partner/restaurants/:id/categories/reorder

# Products
GET    /partner/restaurants/:id/products
POST   /partner/restaurants/:id/products
GET    /partner/products/:id
PUT    /partner/products/:id
DELETE /partner/products/:id
PATCH  /partner/products/:id/availability
POST   /partner/products/bulk-upload          CSV upload
POST   /partner/restaurants/:id/menu/duplicate  Duplicate from another restaurant

# Orders
GET    /partner/orders                      All orders (filterable)
GET    /partner/orders/:id                  Order detail
PATCH  /partner/orders/:id/confirm          Confirm order (restaurant)
PATCH  /partner/orders/:id/reject           Reject order
PATCH  /partner/orders/:id/ready            Mark food ready
GET    /partner/orders/stream               SSE stream for live order updates

# Riders
GET    /partner/riders                      Rider list
POST   /partner/riders                      Create rider account
GET    /partner/riders/:id                  Rider detail
PUT    /partner/riders/:id                  Update rider
PATCH  /partner/riders/:id/availability     Toggle available
DELETE /partner/riders/:id                  Deactivate rider
GET    /partner/riders/tracking             Live rider locations (all active)
GET    /partner/riders/:id/travel-log       Travel log for date
GET    /partner/riders/attendance           Attendance records
POST   /partner/orders/:id/assign-rider     Manually assign rider

# Promos
GET    /partner/promos
POST   /partner/promos
GET    /partner/promos/:id
PUT    /partner/promos/:id
PATCH  /partner/promos/:id/deactivate

# Inventory
GET    /partner/inventory
POST   /partner/inventory/adjust
GET    /partner/inventory/low-stock

# Reports
GET    /partner/reports/sales
GET    /partner/reports/products/top-selling
GET    /partner/reports/orders/breakdown
GET    /partner/reports/peak-hours
GET    /partner/reports/riders

# Finance
GET    /partner/finance/summary
GET    /partner/finance/invoices
GET    /partner/finance/invoices/:id
GET    /partner/finance/invoices/:id/pdf

# Issues
GET    /partner/issues
GET    /partner/issues/:id
POST   /partner/issues/:id/message

# Content
GET    /partner/content/banners
POST   /partner/content/banners
PUT    /partner/content/banners/:id
DELETE /partner/content/banners/:id

GET    /partner/content/stories
POST   /partner/content/stories
DELETE /partner/content/stories/:id

GET    /partner/content/sections
PUT    /partner/content/sections

# Settings
GET    /partner/settings
PUT    /partner/settings
GET    /partner/team
POST   /partner/team/invite
DELETE /partner/team/:user_id
```

---

### Rider API (prefix: `/api/v1/rider/`)
```
GET    /rider/profile                   Rider profile + stats
GET    /rider/orders/active             Current assigned order
PATCH  /rider/orders/:id/picked/:restaurant_id  Mark restaurant pickup
PATCH  /rider/orders/:id/delivered      Mark as delivered
PATCH  /rider/orders/:id/issue          Report order issue
POST   /rider/location                  Update current location
POST   /rider/attendance/checkin        Start shift
POST   /rider/attendance/checkout       End shift
PATCH  /rider/availability              Toggle online/offline
GET    /rider/history                   Delivery history
GET    /rider/earnings                  Earnings summary
```

---

### Admin API (prefix: `/api/v1/admin/`)
```
# Tenants
GET    /admin/tenants
POST   /admin/tenants
GET    /admin/tenants/:id
PUT    /admin/tenants/:id
PATCH  /admin/tenants/:id/status
POST   /admin/tenants/:id/impersonate

# Users
GET    /admin/users
GET    /admin/users/:id
PATCH  /admin/users/:id/status

# Orders (cross-tenant)
GET    /admin/orders
GET    /admin/orders/:id
PATCH  /admin/orders/:id/status   (master override)

# Finance
GET    /admin/finance/commissions
GET    /admin/finance/invoices
POST   /admin/finance/invoices/generate
PATCH  /admin/finance/invoices/:id/finalize
PATCH  /admin/finance/invoices/:id/mark-paid

# Issues
GET    /admin/issues
GET    /admin/issues/:id
PATCH  /admin/issues/:id/resolve
PATCH  /admin/issues/:id/refund/approve
PATCH  /admin/issues/:id/refund/reject

# Analytics
GET    /admin/analytics/overview
GET    /admin/analytics/revenue
GET    /admin/analytics/orders
GET    /admin/analytics/tenants/:id

# Config
GET    /admin/config
PUT    /admin/config
POST   /admin/config/maintenance/enable
POST   /admin/config/maintenance/disable
```

---

## 10.3 Rate Limiting

| Endpoint Group | Limit |
|----------------|-------|
| OTP send | 3 per phone per 10 minutes |
| Order placement | 5 per user per minute |
| General API | 100 per IP per minute |
| Partner API | 500 per user per minute |
| Admin API | 1000 per user per minute |
| Search | 30 per IP per minute |

---

## 10.4 Webhooks (Phase 2)

Vendors can register webhook URLs to receive event notifications:

```
POST /partner/webhooks              Register webhook
GET  /partner/webhooks              List webhooks
DELETE /partner/webhooks/:id        Delete webhook
```

Events delivered:
- `order.created`
- `order.status_changed`
- `order.cancelled`
- `payment.success`
- `payment.refunded`

---

## 10.5 Idempotency & Retry Contract

### Required header (write endpoints)
For all non-safe mutations:
- `POST /orders`
- payment initiation endpoints
- refund endpoints
- invoice finalization/mark-paid endpoints

Client must send:
```
Idempotency-Key: <uuid-or-random-unique-token>
```

### Server behavior
1. First request: process and persist response snapshot.
2. Retry with same key + same payload: return original response.
3. Retry with same key + different payload: `409`.
4. Key retention TTL: 24 hours minimum.

### Recommended client retry policy
- 5xx / network timeout: exponential backoff with jitter.
- 4xx validation errors: no retry.

---

## 10.6 Webhook Security Standard

Webhook delivery must use signed payloads:

Headers:
```
X-Webhook-Event: order.created
X-Webhook-Timestamp: 1700000000
X-Webhook-Signature: sha256=<hex_hmac>
```

Signature formula:
```
signature = HMAC_SHA256(webhook_secret, timestamp + "." + raw_body)
```

Receiver requirements:
- Reject if timestamp skew > 5 minutes.
- Reject invalid signatures (`401`).
- Enforce idempotency using webhook event `id`.

Provider delivery guarantees:
- At-least-once delivery
- Retry schedule: 1m, 5m, 15m, 1h, 6h, 24h (then mark failed)
- Delivery logs exposed to tenant in partner portal

---

## 10.7 API Evolution & Backward Compatibility

- Breaking changes require new version prefix (`/api/v2`), never silent in-place change.
- Additive fields are allowed in existing versions.
- Deprecation policy:
  1. Mark deprecated in docs
  2. Add deprecation response header
  3. Maintain compatibility for minimum 90 days
  4. Remove after migration window

Standard headers:
```
X-Request-ID: <trace-id>
X-API-Version: v1
Deprecation: true|false
Sunset: <ISO datetime>   # when endpoint is planned to be removed
```
