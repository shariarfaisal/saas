# 04 — Domain Model & Entities

## 4.1 Entity Relationship Overview

```
Tenant
  ├── has many → Restaurants
  │               ├── has many → Products
  │               │               ├── has many → ProductVariants
  │               │               └── has many → ProductAddons
  │               ├── has many → Categories
  │               ├── belongs to → Hub (optional)
  │               └── has many → OperatingHours
  │
  ├── has many → Users (customers, staff, riders)
  │
  ├── has many → Orders
  │               ├── has many → OrderItems
  │               ├── belongs to → User (customer)
  │               ├── belongs to → Rider
  │               ├── has many → OrderPickups (per restaurant)
  │               └── has one  → OrderCharge
  │
  ├── has many → Riders
  │               ├── has many → RiderAttendance
  │               └── has many → RiderTravelLogs
  │
  ├── has many → Promos
  ├── has many → DeliveryZones
  ├── has many → Hubs
  ├── has many → Invoices
  ├── has many → SalesReports
  └── has many → Banners / Sections / Stories
```

---

## 4.2 Core Entities

### 4.2.1 Tenant
See [03-multi-tenancy.md](./03-multi-tenancy.md).

---

### 4.2.2 User

A unified user table for all user types within a tenant. Discriminated by `role`.

```
users
  id                  UUID        PK
  tenant_id           UUID        FK → tenants (nullable for super-admin)
  phone               TEXT        UNIQUE per tenant
  email               TEXT        NULLABLE
  name                TEXT
  password_hash       TEXT        NULLABLE  (for partner/admin users)
  role                ENUM        (customer, tenant_owner, tenant_admin, restaurant_manager, restaurant_staff, rider, platform_admin)
  status              ENUM        (active, suspended, deleted)
  avatar_url          TEXT
  gender              ENUM        (male, female, other) NULLABLE
  date_of_birth       DATE        NULLABLE
  device_push_token   TEXT        NULLABLE
  device_info         JSONB       NULLABLE  -- {platform, version, model}
  last_login_at       TIMESTAMPTZ NULLABLE
  referral_code       TEXT        UNIQUE    NULLABLE
  referred_by         UUID        FK → users NULLABLE
  balance             NUMERIC(12,2) DEFAULT 0  -- wallet/loyalty points
  order_count         INT         DEFAULT 0
  last_order_at       TIMESTAMPTZ NULLABLE
  meta                JSONB       NULLABLE  -- flexible extra data
  created_at          TIMESTAMPTZ
  updated_at          TIMESTAMPTZ
  deleted_at          TIMESTAMPTZ NULLABLE  -- soft delete
```

**Customer-specific data** (saved delivery addresses):
```
user_addresses
  id            UUID  PK
  user_id       UUID  FK → users
  tenant_id     UUID  FK → tenants
  label         TEXT          -- "Home", "Office", etc.
  flat          TEXT  NULLABLE
  address       TEXT
  area          TEXT
  city          TEXT  DEFAULT 'Dhaka'
  geo_lat       NUMERIC(10,8) NULLABLE
  geo_lng       NUMERIC(11,8) NULLABLE
  is_default    BOOLEAN DEFAULT false
  created_at    TIMESTAMPTZ
```

---

### 4.2.3 Hub

A hub is a geographic dispatch zone. Restaurants are assigned to a hub. Orders from a hub area are dispatched from that hub's rider pool.

```
hubs
  id          UUID  PK
  tenant_id   UUID  FK → tenants
  name        TEXT
  address     JSONB
  geo_lat     NUMERIC(10,8)
  geo_lng     NUMERIC(11,8)
  is_active   BOOLEAN DEFAULT true
  created_at  TIMESTAMPTZ
  updated_at  TIMESTAMPTZ

hub_areas
  id              UUID  PK
  hub_id          UUID  FK → hubs
  tenant_id       UUID  FK → tenants
  name            TEXT          -- area name e.g. "Gulshan 1"
  delivery_charge NUMERIC(10,2)
  geo_polygon     JSONB NULLABLE  -- polygon coordinates for geo-fencing
```

---

### 4.2.4 Restaurant

```
restaurants
  id                  UUID  PK
  tenant_id           UUID  FK → tenants
  hub_id              UUID  FK → hubs NULLABLE
  managed_by          UUID  FK → users  (the restaurant manager)
  name                TEXT
  slug                TEXT  UNIQUE per tenant
  type                ENUM  (restaurant, cloud_kitchen, store, dark_store)
  description         TEXT  NULLABLE
  banner_image_url    TEXT  NULLABLE
  logo_url            TEXT  NULLABLE
  phone               TEXT  NULLABLE
  address             JSONB         -- {flat, street, area, city, geo_lat, geo_lng}
  cuisines            TEXT[]        -- array of cuisine names/ids
  commission_rate     NUMERIC(5,2)  NULLABLE  -- overrides tenant default if set
  vat_rate            NUMERIC(5,2)  DEFAULT 0
  is_vat_included     BOOLEAN DEFAULT false
  availability        BOOLEAN DEFAULT true   -- is restaurant open
  prep_time           INT DEFAULT 20         -- avg prep time in minutes
  prep_time_penalty   NUMERIC(5,2) DEFAULT 0 -- % penalty if prep exceeds limit
  max_concurrent_orders INT DEFAULT 10
  order_prefix        TEXT NULLABLE  -- e.g. "KBC" for order numbering
  meta_title          TEXT NULLABLE
  meta_description    TEXT NULLABLE
  meta_tags           TEXT[] NULLABLE
  sort_order          INT DEFAULT 0
  is_active           BOOLEAN DEFAULT true
  created_at          TIMESTAMPTZ
  updated_at          TIMESTAMPTZ

restaurant_operating_hours
  id              UUID  PK
  restaurant_id   UUID  FK → restaurants
  tenant_id       UUID  FK → tenants
  day_of_week     INT   -- 0=Sunday … 6=Saturday
  open_time       TIME
  close_time      TIME
  is_closed       BOOLEAN DEFAULT false
```

---

### 4.2.5 Category

```
categories
  id              UUID  PK
  tenant_id       UUID  FK → tenants
  restaurant_id   UUID  FK → restaurants NULLABLE  -- null = global category for tenant
  name            TEXT
  slug            TEXT
  image_url       TEXT  NULLABLE
  prep_time       INT   DEFAULT 0    -- extra prep time for this category
  sort_order      INT   DEFAULT 0
  is_active       BOOLEAN DEFAULT true
  is_tobacco      BOOLEAN DEFAULT false  -- age-gated category
  created_at      TIMESTAMPTZ
  updated_at      TIMESTAMPTZ
```

---

### 4.2.6 Product (Menu Item)

```
products
  id              UUID  PK
  tenant_id       UUID  FK → tenants
  restaurant_id   UUID  FK → restaurants
  category_id     UUID  FK → categories NULLABLE
  name            TEXT
  slug            TEXT  UNIQUE per restaurant
  description     TEXT  NULLABLE
  base_price      NUMERIC(10,2)
  vat_rate        NUMERIC(5,2) DEFAULT 0
  price_type      ENUM  (flat, variant)
  availability    ENUM  (available, unavailable, out_of_stock)
  images          TEXT[]
  sort_order      INT DEFAULT 0
  meta_title      TEXT NULLABLE
  meta_description TEXT NULLABLE
  meta_tags       TEXT[] NULLABLE
  is_inv_tracked  BOOLEAN DEFAULT false  -- whether stock is managed
  created_at      TIMESTAMPTZ
  updated_at      TIMESTAMPTZ

product_variants       -- e.g. "Size" with options "Small/Medium/Large"
  id              UUID  PK
  product_id      UUID  FK → products
  tenant_id       UUID  FK → tenants
  title           TEXT          -- e.g. "Size"
  min_select      INT DEFAULT 1
  max_select      INT DEFAULT 1
  sort_order      INT DEFAULT 0

product_variant_items  -- individual variant option
  id              UUID  PK
  variant_id      UUID  FK → product_variants
  tenant_id       UUID  FK → tenants
  name            TEXT          -- e.g. "Large"
  price           NUMERIC(10,2) DEFAULT 0  -- extra charge
  is_available    BOOLEAN DEFAULT true
  sort_order      INT DEFAULT 0

product_addons         -- add-on group e.g. "Extra Toppings"
  id              UUID  PK
  product_id      UUID  FK → products
  tenant_id       UUID  FK → tenants
  title           TEXT
  min_select      INT DEFAULT 0
  max_select      INT DEFAULT 1

product_addon_items
  id              UUID  PK
  addon_id        UUID  FK → product_addons
  tenant_id       UUID  FK → tenants
  name            TEXT
  price           NUMERIC(10,2) DEFAULT 0
  is_available    BOOLEAN DEFAULT true

product_discounts
  id              UUID  PK
  product_id      UUID  FK → products
  tenant_id       UUID  FK → tenants
  discount_type   ENUM  (fixed, percent)
  amount          NUMERIC(10,2)
  valid_until     TIMESTAMPTZ NULLABLE
  created_at      TIMESTAMPTZ
```

---

### 4.2.7 Inventory

```
inventory
  id              UUID  PK
  product_id      UUID  FK → products
  tenant_id       UUID  FK → tenants
  restaurant_id   UUID  FK → restaurants
  stock           INT   DEFAULT 0
  unit_price      NUMERIC(10,2)  -- purchase/cost price
  reorder_level   INT DEFAULT 5
  last_updated_at TIMESTAMPTZ
  updated_by      UUID  FK → users
```

---

### 4.2.8 Order

The central entity of the platform.

```
orders
  id                  UUID  PK
  tenant_id           UUID  FK → tenants
  order_number        TEXT          -- human-readable e.g. "KBC-001234"
  customer_id         UUID  FK → users
  rider_id            UUID  FK → users (rider) NULLABLE
  hub_id              UUID  FK → hubs NULLABLE
  status              ENUM  (pending, created, confirmed, preparing, ready, picked, delivered, cancelled, rejected)
  payment_status      ENUM  (unpaid, paid, refunded, partially_refunded)
  payment_method      ENUM  (cod, bkash, aamarpay, sslcommerz, wallet, card)
  
  -- Customer delivery info
  delivery_address    JSONB         -- snapshot at order time
  customer_name       TEXT
  customer_phone      TEXT
  customer_area       TEXT          -- resolved delivery zone area
  geo_lat             NUMERIC(10,8) NULLABLE
  geo_lng             NUMERIC(11,8) NULLABLE
  
  -- Charge breakdown
  subtotal            NUMERIC(12,2) -- sum of item totals before discounts
  item_discount       NUMERIC(12,2) DEFAULT 0  -- product-level discounts
  promo_discount      NUMERIC(12,2) DEFAULT 0  -- promo code discount
  vat                 NUMERIC(12,2) DEFAULT 0
  delivery_charge     NUMERIC(12,2) DEFAULT 0
  service_charge      NUMERIC(12,2) DEFAULT 0  -- platform fee (future)
  total               NUMERIC(12,2) -- final amount customer pays
  
  -- Promo info (snapshot)
  promo_id            UUID  FK → promos NULLABLE
  promo_snapshot      JSONB NULLABLE  -- snapshot of promo at order time
  
  -- Notes
  customer_note       TEXT  NULLABLE
  rider_note          TEXT  NULLABLE
  internal_note       TEXT  NULLABLE
  
  -- Platform
  platform            ENUM  (web, ios, android, pos)
  is_priority         BOOLEAN DEFAULT false
  
  -- Rejection / cancellation
  rejection_reason    TEXT  NULLABLE
  rejected_by         ENUM  (restaurant, platform, customer, system) NULLABLE
  
  -- Timing
  confirmed_at        TIMESTAMPTZ NULLABLE
  preparing_at        TIMESTAMPTZ NULLABLE
  ready_at            TIMESTAMPTZ NULLABLE
  picked_at           TIMESTAMPTZ NULLABLE
  delivered_at        TIMESTAMPTZ NULLABLE
  cancelled_at        TIMESTAMPTZ NULLABLE
  estimated_delivery_time INT NULLABLE  -- minutes
  
  created_at          TIMESTAMPTZ
  updated_at          TIMESTAMPTZ
  deleted_at          TIMESTAMPTZ NULLABLE

order_items
  id              UUID  PK
  order_id        UUID  FK → orders
  tenant_id       UUID  FK → tenants
  restaurant_id   UUID  FK → restaurants
  product_id      UUID  FK → products
  product_snapshot JSONB  -- full product details at order time
  quantity        INT
  unit_price      NUMERIC(10,2)  -- base price at order time
  variant_price   NUMERIC(10,2) DEFAULT 0
  addon_price     NUMERIC(10,2) DEFAULT 0
  subtotal        NUMERIC(10,2)  -- (unit + variant + addon) * qty
  discount        NUMERIC(10,2) DEFAULT 0
  promo_discount  NUMERIC(10,2) DEFAULT 0
  vat             NUMERIC(10,2) DEFAULT 0
  total           NUMERIC(10,2)
  selected_variants JSONB NULLABLE
  selected_addons   JSONB NULLABLE
  created_at      TIMESTAMPTZ

order_pickups   -- one row per restaurant in a multi-restaurant order
  id              UUID  PK
  order_id        UUID  FK → orders
  restaurant_id   UUID  FK → restaurants
  tenant_id       UUID  FK → tenants
  order_number    TEXT          -- sub order number for this restaurant
  status          ENUM  (new, confirmed, preparing, ready, picked, rejected)
  items           JSONB         -- items from this restaurant
  items_total     NUMERIC(12,2)
  commission_rate NUMERIC(5,2)
  commission_amount NUMERIC(12,2)
  vat             NUMERIC(12,2)
  confirmed_at    TIMESTAMPTZ NULLABLE
  ready_at        TIMESTAMPTZ NULLABLE
  picked_at       TIMESTAMPTZ NULLABLE
  rejected_at     TIMESTAMPTZ NULLABLE
  rejection_reason TEXT NULLABLE
  created_at      TIMESTAMPTZ

order_timeline  -- audit trail of all order state changes
  id              UUID  PK
  order_id        UUID  FK → orders
  tenant_id       UUID  FK → tenants
  event_type      TEXT          -- status_change, rider_assign, note_added, etc.
  old_status      TEXT  NULLABLE
  new_status      TEXT  NULLABLE
  message         TEXT
  actor_id        UUID  FK → users NULLABLE
  actor_type      ENUM  (customer, restaurant, rider, admin, system)
  created_at      TIMESTAMPTZ
```

---

### 4.2.9 Rider

```
riders          -- extends the users table (role = 'rider')
  user_id         UUID  PK FK → users
  tenant_id       UUID  FK → tenants
  hub_id          UUID  FK → hubs NULLABLE
  is_available    BOOLEAN DEFAULT false  -- currently available to take orders
  is_on_duty      BOOLEAN DEFAULT false
  vehicle_type    ENUM  (bicycle, motorcycle, car)
  nid_number      TEXT  NULLABLE
  license_number  TEXT  NULLABLE
  balance         NUMERIC(12,2) DEFAULT 0  -- earnings balance
  total_earnings  NUMERIC(12,2) DEFAULT 0
  order_count     INT DEFAULT 0
  rating          NUMERIC(3,2) DEFAULT 5.0
  created_at      TIMESTAMPTZ
  updated_at      TIMESTAMPTZ

rider_locations  -- real-time location (updated frequently)
  rider_id        UUID  PK FK → riders
  tenant_id       UUID  FK → tenants
  geo_lat         NUMERIC(10,8)
  geo_lng         NUMERIC(11,8)
  updated_at      TIMESTAMPTZ

rider_travel_logs  -- historical location trail per shift
  id              UUID  PK
  rider_id        UUID  FK → riders
  tenant_id       UUID  FK → tenants
  order_id        UUID  FK → orders NULLABLE
  geo_lat         NUMERIC(10,8)
  geo_lng         NUMERIC(11,8)
  subject         ENUM  (attendance_in, attendance_out, picked, delivered, in_hub, location_update)
  distance_from_prev NUMERIC(10,3) NULLABLE  -- km
  created_at      TIMESTAMPTZ

rider_attendance
  id              UUID  PK
  rider_id        UUID  FK → riders
  tenant_id       UUID  FK → tenants
  date            DATE
  checked_in_at   TIMESTAMPTZ NULLABLE
  checked_out_at  TIMESTAMPTZ NULLABLE
  total_hours     NUMERIC(5,2) NULLABLE
  total_distance  NUMERIC(10,3) NULLABLE  -- km for the day
  total_orders    INT DEFAULT 0
  penalty         JSONB NULLABLE

rider_penalties
  id              UUID  PK
  rider_id        UUID  FK → riders
  tenant_id       UUID  FK → tenants
  order_id        UUID  FK → orders NULLABLE
  reason          TEXT
  amount          NUMERIC(10,2)
  status          ENUM  (pending, cleared, appealed)
  appeal_note     TEXT  NULLABLE
  created_at      TIMESTAMPTZ
```

---

### 4.2.10 Promo

```
promos
  id                  UUID  PK
  tenant_id           UUID  FK → tenants
  code                TEXT  UNIQUE per tenant
  auth_key            TEXT  NULLABLE  -- secondary secret for API-based application
  description         TEXT  NULLABLE
  promo_type          ENUM  (fixed, percent)
  amount              NUMERIC(10,2)
  max_discount_amount NUMERIC(10,2) NULLABLE  -- cap on percent discounts
  cashback_amount     NUMERIC(10,2) DEFAULT 0  -- wallet credit on use
  apply_on            ENUM  (all_items, category, specific_restaurant, delivery_charge)
  
  -- Scope restrictions
  restaurant_id       UUID  FK → restaurants NULLABLE  -- restrict to restaurant
  category_ids        UUID[]  NULLABLE                 -- restrict to categories
  restaurant_ids      UUID[]  NULLABLE                 -- restrict to set of restaurants
  include_stores      BOOLEAN DEFAULT false
  
  -- Usage limits
  min_order_amount    NUMERIC(10,2) DEFAULT 0
  max_usage           INT NULLABLE          -- total uses across all users
  max_usage_per_user  INT DEFAULT 1
  eligible_user_ids   UUID[] NULLABLE       -- only specific users can use
  
  -- Validity
  is_active           BOOLEAN DEFAULT true
  start_date          TIMESTAMPTZ
  end_date            TIMESTAMPTZ
  created_at          TIMESTAMPTZ
  updated_at          TIMESTAMPTZ

promo_usages
  id          UUID  PK
  promo_id    UUID  FK → promos
  user_id     UUID  FK → users
  order_id    UUID  FK → orders
  tenant_id   UUID  FK → tenants
  amount_used NUMERIC(10,2)
  created_at  TIMESTAMPTZ
```

---

### 4.2.11 Payment

```
payment_transactions
  id                  UUID  PK
  tenant_id           UUID  FK → tenants
  order_id            UUID  FK → orders
  user_id             UUID  FK → users
  method              ENUM  (cod, bkash, aamarpay, sslcommerz, wallet, card)
  status              ENUM  (pending, success, failed, refunded, cancelled)
  amount              NUMERIC(12,2)
  currency            TEXT DEFAULT 'BDT'
  gateway_txn_id      TEXT  NULLABLE  -- transaction ID from payment gateway
  gateway_response    JSONB NULLABLE  -- raw gateway response
  ip_address          INET  NULLABLE
  created_at          TIMESTAMPTZ
  updated_at          TIMESTAMPTZ

refunds
  id                  UUID  PK
  tenant_id           UUID  FK → tenants
  order_id            UUID  FK → orders
  transaction_id      UUID  FK → payment_transactions
  amount              NUMERIC(12,2)
  reason              TEXT
  status              ENUM  (pending, processed, rejected)
  processed_by        UUID  FK → users NULLABLE
  processed_at        TIMESTAMPTZ NULLABLE
  gateway_refund_id   TEXT  NULLABLE
  created_at          TIMESTAMPTZ
```

---

### 4.2.12 Invoice & Financial Settlement

```
invoices
  id                  UUID  PK
  tenant_id           UUID  FK → tenants
  restaurant_id       UUID  FK → restaurants
  period_start        DATE
  period_end          DATE
  total_sales         NUMERIC(12,2)   -- gross order subtotals
  item_discount       NUMERIC(12,2)   -- product-level discounts
  promo_discount      NUMERIC(12,2)   -- promo discounts (platform-funded)
  vat_collected       NUMERIC(12,2)
  vat_amount          NUMERIC(12,2)   -- platform collects on behalf
  commission_rate     NUMERIC(5,2)
  commission_amount   NUMERIC(12,2)
  penalty             NUMERIC(12,2) DEFAULT 0
  adjustment          NUMERIC(12,2) DEFAULT 0
  net_payable         NUMERIC(12,2)   -- what platform owes the restaurant
  order_count         INT
  rejected_order_count INT
  status              ENUM  (draft, finalized, paid)
  paid_by             UUID  FK → users NULLABLE
  paid_at             TIMESTAMPTZ NULLABLE
  note                TEXT  NULLABLE
  created_at          TIMESTAMPTZ
  updated_at          TIMESTAMPTZ
```

---

### 4.2.13 Order Issue / Dispute

```
order_issues
  id                  UUID  PK
  tenant_id           UUID  FK → tenants
  order_id            UUID  FK → orders
  type                ENUM  (wrong_item, missing_item, quality_issue, late_delivery, other)
  reported_by         UUID  FK → users
  accountable         ENUM  (restaurant, rider, platform)
  details             TEXT
  images              TEXT[] NULLABLE
  
  refund_items        JSONB NULLABLE  -- which items to refund
  refund_amount       NUMERIC(12,2) DEFAULT 0
  refund_status       ENUM  (pending, approved, rejected, processed) DEFAULT pending
  
  restaurant_penalty  JSONB NULLABLE  -- {id, amount, apply_penalty: bool}
  rider_penalty       NUMERIC(10,2) DEFAULT 0
  
  status              ENUM  (open, resolved, closed)
  resolved_by         UUID  FK → users NULLABLE
  resolved_at         TIMESTAMPTZ NULLABLE
  resolution_note     TEXT  NULLABLE
  messages            JSONB DEFAULT '[]'  -- internal notes thread
  created_at          TIMESTAMPTZ
  updated_at          TIMESTAMPTZ
```

---

### 4.2.14 Homepage & Content

```
banners
  id              UUID  PK
  tenant_id       UUID  FK → tenants
  title           TEXT
  image_url       TEXT
  link_type       ENUM  (restaurant, product, category, url, promo)
  link_value      TEXT NULLABLE
  platform        ENUM  (web, app, all)
  sort_order      INT DEFAULT 0
  is_active       BOOLEAN DEFAULT true
  hub_ids         UUID[] NULLABLE   -- show only in these hubs
  area_ids        TEXT[] NULLABLE   -- show only in these areas
  valid_from      TIMESTAMPTZ NULLABLE
  valid_until     TIMESTAMPTZ NULLABLE
  created_at      TIMESTAMPTZ

sections         -- curated homepage sections e.g. "Trending Now"
  id              UUID  PK
  tenant_id       UUID  FK → tenants
  title           TEXT
  type            ENUM  (restaurants, products, categories, banners)
  items           UUID[]          -- IDs of content items
  sort_order      INT DEFAULT 0
  is_active       BOOLEAN DEFAULT true
  hub_ids         UUID[] NULLABLE
  created_at      TIMESTAMPTZ
  updated_at      TIMESTAMPTZ

stories          -- short-lived promotional content (like Instagram stories)
  id              UUID  PK
  tenant_id       UUID  FK → tenants
  restaurant_id   UUID  FK → restaurants NULLABLE
  media_url       TEXT
  media_type      ENUM  (image, video)
  link_type       ENUM  (restaurant, product, url) NULLABLE
  link_value      TEXT NULLABLE
  expires_at      TIMESTAMPTZ
  sort_order      INT DEFAULT 0
  created_at      TIMESTAMPTZ
```

---

### 4.2.15 Analytics

```
order_analytics   -- denormalized for fast analytics queries
  id                    UUID  PK
  tenant_id             UUID  FK → tenants
  order_id              UUID  FK → orders
  restaurant_ids        UUID[]
  customer_id           UUID
  rider_id              UUID NULLABLE
  hub_id                UUID NULLABLE
  
  -- Dimensions
  customer_area         TEXT
  payment_method        TEXT
  platform              TEXT
  promo_code            TEXT NULLABLE
  
  -- Order totals
  subtotal              NUMERIC(12,2)
  delivery_charge       NUMERIC(12,2)
  discount              NUMERIC(12,2)
  promo_discount        NUMERIC(12,2)
  vat                   NUMERIC(12,2)
  total                 NUMERIC(12,2)
  commission            NUMERIC(12,2)
  
  -- Timing metrics (seconds)
  confirmation_time_s   INT NULLABLE
  preparation_time_s    INT NULLABLE
  pickup_to_delivery_s  INT NULLABLE
  total_delivery_time_s INT NULLABLE
  
  -- Status
  final_status          TEXT
  rejection_reason      TEXT NULLABLE
  
  -- Time dimensions
  order_date            DATE
  order_hour            INT
  order_day_of_week     INT
  order_month           INT
  order_year            INT
  
  completed_at          TIMESTAMPTZ NULLABLE
  created_at            TIMESTAMPTZ
```

---

## 4.3 Enumerated Types Summary

```sql
CREATE TYPE order_status AS ENUM ('pending','created','confirmed','preparing','ready','picked','delivered','cancelled','rejected');
CREATE TYPE payment_status AS ENUM ('unpaid','paid','refunded','partially_refunded');
CREATE TYPE payment_method AS ENUM ('cod','bkash','aamarpay','sslcommerz','wallet','card');
CREATE TYPE user_role AS ENUM ('customer','tenant_owner','tenant_admin','restaurant_manager','restaurant_staff','rider','platform_admin','platform_support','platform_finance');
CREATE TYPE user_status AS ENUM ('active','suspended','deleted');
CREATE TYPE restaurant_type AS ENUM ('restaurant','cloud_kitchen','store','dark_store');
CREATE TYPE product_availability AS ENUM ('available','unavailable','out_of_stock');
CREATE TYPE promo_type AS ENUM ('fixed','percent');
CREATE TYPE tenant_status AS ENUM ('pending','active','suspended','cancelled');
CREATE TYPE invoice_status AS ENUM ('draft','finalized','paid');
CREATE TYPE rider_subject AS ENUM ('attendance_in','attendance_out','picked','delivered','in_hub','location_update');
```
