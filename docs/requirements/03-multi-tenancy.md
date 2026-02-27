# 03 — Multi-Tenancy Model

## 3.1 Tenancy Overview

The platform uses **shared database, shared schema** multi-tenancy with row-level isolation. This is the most cost-efficient approach and easiest to operate for our scale.

Each vendor is a **tenant**. Every resource in the system — restaurants, menus, orders, riders, customers, reports — belongs to exactly one tenant (except platform-wide entities like super-admin config).

```
Platform (Super Admin)
    │
    ├── Tenant A (e.g. "Kacchi Bhai" brand)
    │       ├── Restaurant 1 (Gulshan branch)
    │       ├── Restaurant 2 (Dhanmondi branch)
    │       ├── Restaurant 3 (Cloud kitchen)
    │       ├── Customers (scoped to Tenant A)
    │       ├── Riders (scoped to Tenant A)
    │       └── Orders, Reports, etc.
    │
    ├── Tenant B (e.g. "Pizza Mania" brand)
    │       ├── Restaurant 1
    │       └── ...
    │
    └── Tenant C ...
```

---

## 3.2 Tenant Entity

```
tenants
  id              UUID  PK
  slug            TEXT  UNIQUE        -- used in subdomain: slug.platform.com
  name            TEXT                -- display name
  status          ENUM  (active, suspended, pending, cancelled)
  plan            ENUM  (starter, growth, enterprise)  -- subscription tier
  commission_rate NUMERIC(5,2)        -- platform commission % (default)
  settings        JSONB               -- feature flags, custom config
  domain          TEXT  NULLABLE      -- custom domain (future)
  logo_url        TEXT
  primary_color   TEXT                -- brand color for storefront theming
  contact_email   TEXT
  contact_phone   TEXT
  created_at      TIMESTAMPTZ
  updated_at      TIMESTAMPTZ
```

---

## 3.3 Tenant Resolution

The tenant is resolved on every request through **middleware** before any handler runs:

**Strategy 1 — Subdomain (Customer Website)**
```
kacchbhai.platform.com  →  tenant.slug = 'kacchbhai'
```
The middleware reads `Host` header, strips the platform root domain, queries (or reads from Redis cache) the tenant by slug, and sets `tenant_id` in the request context.

**Strategy 2 — JWT Claim (Partner / Admin Portal)**
The authenticated user's JWT contains `tenant_id`. The middleware reads it directly. For super-admins, `tenant_id` is `null` (they can impersonate any tenant via query param).

**Strategy 3 — X-Tenant-ID Header (API clients / internal)**
Allowed only for requests authenticated with API keys.

```go
// Middleware sets this in context for all downstream handlers
type TenantContext struct {
    TenantID uuid.UUID
    Tenant   *Tenant
}
```

---

## 3.4 Data Isolation Rules

| Table Type | Isolation Method |
|------------|-----------------|
| Tenant-scoped (restaurants, orders, products, etc.) | `tenant_id` column + index on every table; all queries filter by `tenant_id` |
| User table | `tenant_id` column; customers of Tenant A cannot login to Tenant B's website |
| Platform-wide (tenants, super-admin users, platform config) | No `tenant_id`; accessible only to super-admin role |
| Analytics | `tenant_id` scoped; super-admin can query across tenants |

**Critical rule:** Every repository query that touches a tenant-scoped table **must** include `WHERE tenant_id = $tenant_id`. SQLC queries will enforce this at compile time via typed parameters.

---

## 3.5 Vendor Onboarding Flow

### Phase 1 — Super-Admin Managed (MVP)
Super-admin manually creates a tenant through the admin panel.

**Steps:**
1. Super-admin creates tenant record (name, slug, plan, commission rate)
2. System creates tenant owner account (sends invite email to vendor's email)
3. Vendor logs into partner portal, completes profile setup
4. Vendor adds at least one restaurant to get started
5. Super-admin activates tenant (sets status: active)
6. Subdomain goes live

### Phase 2 — Self-Serve Onboarding
1. Vendor visits `platform.com/register`
2. Fills out business registration form
3. Email verification
4. Chooses subdomain (slug)
5. Selects plan
6. Completes restaurant setup wizard
7. Auto-approval OR admin review queue
8. Payment setup for subscription/commission

---

## 3.6 Vendor Roles Within a Tenant

A vendor (tenant) has its own internal user roles:

| Role | Description |
|------|-------------|
| `tenant_owner` | Full access to all restaurants, financials, users |
| `tenant_admin` | Same as owner but cannot delete tenant or change billing |
| `restaurant_manager` | Access only to assigned restaurants (orders, menu, reports) |
| `restaurant_staff` | Read-only access to orders for assigned restaurant |
| `rider` | Delivery rider — dedicated limited access (order pickup/delivery) |

These roles are **tenant-scoped** — a `tenant_admin` of Tenant A has no access to Tenant B.

---

## 3.7 Platform Roles (Super-Admin)

| Role | Description |
|------|-------------|
| `super_admin` | Full platform access — all tenants, all data |
| `platform_support` | Can view all tenants, cannot modify billing or delete data |
| `platform_finance` | Can view and manage invoices, payouts across all tenants |

---

## 3.8 Tenant Feature Flags

Each tenant can have features enabled/disabled via `settings` JSONB:

```json
{
  "features": {
    "rider_management": true,
    "loyalty_points": false,
    "multi_language": false,
    "custom_domain": false,
    "advanced_analytics": true,
    "pos_integration": false
  },
  "order": {
    "auto_confirm_after_minutes": 3,
    "max_concurrent_orders_per_restaurant": 20,
    "allow_scheduled_orders": false
  },
  "delivery": {
    "use_own_riders": true,
    "allow_third_party_courier": false
  }
}
```

---

## 3.9 Tenant Suspension & Offboarding

- **Suspension**: Tenant status set to `suspended`. All API requests return 403. Existing data preserved.
- **Cancellation**: Tenant status set to `cancelled`. Data retained for 90 days, then purged on request.
- **Data export**: Vendors can request a full data export (CSV/JSON) before cancellation.

---

## 3.10 Defense-in-Depth Isolation (World-Class Standard)

Application-layer filtering (`WHERE tenant_id = $tenant_id`) is mandatory but not sufficient alone for enterprise-grade protection. Add database-level safeguards:

1. **PostgreSQL Row-Level Security (RLS)** on tenant-scoped tables
2. **Session tenant context** set at transaction start (`SET LOCAL app.tenant_id = '<uuid>'`)
3. **RLS policy** enforces `tenant_id = current_setting('app.tenant_id')::uuid`
4. **Bypass only for platform service roles** with explicit audited elevation

Representative policy pattern:

```sql
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;

CREATE POLICY orders_tenant_isolation ON orders
USING (tenant_id = current_setting('app.tenant_id')::uuid)
WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);
```

This ensures accidental missing tenant filters in SQL cannot leak cross-tenant data.

### Cross-tenant access policy
- Allowed only for `platform_admin` and explicitly tagged support sessions.
- Every cross-tenant query must emit an `audit_logs` record with reason and operator identity.
- Impersonation sessions must be time-bound and clearly visible in UI.

---

## 3.11 Tenant Data Governance & Lifecycle

### Data residency and retention
- Default region: Bangladesh-adjacent region (`ap-southeast-1`/equivalent) for low latency.
- Tenant data retention policy configurable by plan; default transactional retention 3 years.
- PII minimization: collect only required KYC/contact fields.

### Tenant export SLA
- Full tenant export (`CSV + JSON`) must complete within 24 hours for standard plans.
- Export package includes checksum manifest and schema version.

### Tenant deletion
- Soft-delete phase: 90 days (recoverable)
- Hard-delete phase: cryptographic wipe and deletion report
- Deletion completion certificate stored in admin records
