# 09 — Database Schema

## 9.1 Design Principles

- **PostgreSQL 16+** as the only database
- **SQLC** generates all Go types from SQL queries — no ORM
- **UUID v4** primary keys (using `gen_random_uuid()`)
- **TIMESTAMPTZ** for all timestamps (UTC stored)
- **Soft deletes** via `deleted_at` TIMESTAMPTZ NULLABLE on core entities
- **JSONB** for flexible structured data (snapshots, settings, metadata)
- **TEXT arrays** (`TEXT[]`) for simple tag/string collections
- **UUID arrays** (`UUID[]`) for foreign key sets (avoid join table for small sets)
- **Row-level tenant isolation**: `tenant_id UUID NOT NULL` on every tenant-scoped table
- **Indexes on all FK columns** and common query columns
- **Partial indexes** where appropriate (e.g., active records only)

---

## 9.2 Complete Schema (SQLC-ready)

```sql
-- ============================================================
-- EXTENSIONS
-- ============================================================
CREATE EXTENSION IF NOT EXISTS "pgcrypto";   -- gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pg_trgm";    -- trigram text search

-- ============================================================
-- ENUMS
-- ============================================================
CREATE TYPE tenant_status    AS ENUM ('pending','active','suspended','cancelled');
CREATE TYPE tenant_plan      AS ENUM ('starter','growth','enterprise');
CREATE TYPE user_role        AS ENUM ('customer','tenant_owner','tenant_admin','restaurant_manager','restaurant_staff','rider','platform_admin','platform_support','platform_finance');
CREATE TYPE user_status      AS ENUM ('active','suspended','deleted');
CREATE TYPE restaurant_type  AS ENUM ('restaurant','cloud_kitchen','store','dark_store');
CREATE TYPE product_avail    AS ENUM ('available','unavailable','out_of_stock');
CREATE TYPE price_type       AS ENUM ('flat','variant');
CREATE TYPE order_status     AS ENUM ('pending','created','confirmed','preparing','ready','picked','delivered','cancelled','rejected');
CREATE TYPE pickup_status    AS ENUM ('new','confirmed','preparing','ready','picked','rejected');
CREATE TYPE payment_status   AS ENUM ('unpaid','paid','refunded','partially_refunded');
CREATE TYPE payment_method   AS ENUM ('cod','bkash','aamarpay','sslcommerz','wallet','card');
CREATE TYPE txn_status       AS ENUM ('pending','success','failed','refunded','cancelled');
CREATE TYPE promo_type       AS ENUM ('fixed','percent');
CREATE TYPE promo_apply_on   AS ENUM ('all_items','category','specific_restaurant','delivery_charge');
CREATE TYPE promo_funder     AS ENUM ('vendor','platform','restaurant');
CREATE TYPE invoice_status   AS ENUM ('draft','finalized','paid');
CREATE TYPE issue_type       AS ENUM ('wrong_item','missing_item','quality_issue','late_delivery','other');
CREATE TYPE issue_status     AS ENUM ('open','resolved','closed');
CREATE TYPE accountable      AS ENUM ('restaurant','rider','platform');
CREATE TYPE refund_status    AS ENUM ('pending','approved','rejected','processed');
CREATE TYPE rider_subject    AS ENUM ('attendance_in','attendance_out','picked','delivered','in_hub','location_update');
CREATE TYPE wallet_type      AS ENUM ('credit','debit');
CREATE TYPE wallet_source    AS ENUM ('cashback','referral','welcome','refund','order_payment','admin_adjustment');
CREATE TYPE platform_source  AS ENUM ('web','ios','android','pos');
CREATE TYPE discount_type    AS ENUM ('fixed','percent');
CREATE TYPE vehicle_type     AS ENUM ('bicycle','motorcycle','car');
CREATE TYPE penalty_status   AS ENUM ('pending','cleared','appealed');
CREATE TYPE delivery_model   AS ENUM ('zone_based','distance_based');

-- ============================================================
-- TENANTS
-- ============================================================
CREATE TABLE tenants (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            TEXT        NOT NULL UNIQUE,
    name            TEXT        NOT NULL,
    status          tenant_status NOT NULL DEFAULT 'pending',
    plan            tenant_plan NOT NULL DEFAULT 'starter',
    commission_rate NUMERIC(5,2) NOT NULL DEFAULT 10.00,
    settings        JSONB       NOT NULL DEFAULT '{}',
    domain          TEXT        UNIQUE,
    logo_url        TEXT,
    primary_color   TEXT        DEFAULT '#FF6B35',
    contact_email   TEXT        NOT NULL,
    contact_phone   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_status ON tenants(status);

-- ============================================================
-- USERS
-- ============================================================
CREATE TABLE users (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID        REFERENCES tenants(id),
    phone               TEXT,
    email               TEXT,
    name                TEXT        NOT NULL DEFAULT '',
    password_hash       TEXT,
    role                user_role   NOT NULL DEFAULT 'customer',
    status              user_status NOT NULL DEFAULT 'active',
    avatar_url          TEXT,
    gender              TEXT,
    date_of_birth       DATE,
    device_push_token   TEXT,
    device_info         JSONB       DEFAULT '{}',
    last_login_at       TIMESTAMPTZ,
    referral_code       TEXT        UNIQUE,
    referred_by         UUID        REFERENCES users(id),
    balance             NUMERIC(12,2) NOT NULL DEFAULT 0,
    order_count         INT         NOT NULL DEFAULT 0,
    last_order_at       TIMESTAMPTZ,
    meta                JSONB       DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,
    UNIQUE(tenant_id, phone)
);
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status) WHERE deleted_at IS NULL;

CREATE TABLE user_addresses (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id   UUID        NOT NULL REFERENCES tenants(id),
    label       TEXT        NOT NULL DEFAULT 'Home',
    flat        TEXT,
    address     TEXT        NOT NULL,
    area        TEXT        NOT NULL,
    city        TEXT        NOT NULL DEFAULT 'Dhaka',
    geo_lat     NUMERIC(10,8),
    geo_lng     NUMERIC(11,8),
    is_default  BOOLEAN     NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_user_addresses_user_id ON user_addresses(user_id);

CREATE TABLE otp_verifications (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    phone       TEXT        NOT NULL,
    tenant_id   UUID        REFERENCES tenants(id),
    otp_hash    TEXT        NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    used        BOOLEAN     NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_otp_phone ON otp_verifications(phone, expires_at);

-- ============================================================
-- HUBS & DELIVERY ZONES
-- ============================================================
CREATE TABLE hubs (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID        NOT NULL REFERENCES tenants(id),
    name        TEXT        NOT NULL,
    address     JSONB       NOT NULL DEFAULT '{}',
    geo_lat     NUMERIC(10,8),
    geo_lng     NUMERIC(11,8),
    is_active   BOOLEAN     NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_hubs_tenant_id ON hubs(tenant_id);

CREATE TABLE hub_areas (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    hub_id          UUID        NOT NULL REFERENCES hubs(id) ON DELETE CASCADE,
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    name            TEXT        NOT NULL,
    delivery_charge NUMERIC(10,2) NOT NULL DEFAULT 0,
    geo_polygon     JSONB,
    is_active       BOOLEAN     NOT NULL DEFAULT true
);
CREATE INDEX idx_hub_areas_hub_id ON hub_areas(hub_id);
CREATE INDEX idx_hub_areas_tenant_id ON hub_areas(tenant_id);
CREATE UNIQUE INDEX idx_hub_areas_name_hub ON hub_areas(hub_id, name);

-- ============================================================
-- RESTAURANTS
-- ============================================================
CREATE TABLE restaurants (
    id                    UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             UUID            NOT NULL REFERENCES tenants(id),
    hub_id                UUID            REFERENCES hubs(id),
    managed_by            UUID            REFERENCES users(id),
    name                  TEXT            NOT NULL,
    slug                  TEXT            NOT NULL,
    type                  restaurant_type NOT NULL DEFAULT 'restaurant',
    description           TEXT,
    banner_image_url      TEXT,
    logo_url              TEXT,
    phone                 TEXT,
    address               JSONB           NOT NULL DEFAULT '{}',
    cuisines              TEXT[]          DEFAULT '{}',
    commission_rate       NUMERIC(5,2),
    vat_rate              NUMERIC(5,2)    NOT NULL DEFAULT 0,
    is_vat_included       BOOLEAN         NOT NULL DEFAULT false,
    availability          BOOLEAN         NOT NULL DEFAULT true,
    prep_time             INT             NOT NULL DEFAULT 20,
    prep_time_penalty     NUMERIC(5,2)    NOT NULL DEFAULT 0,
    max_concurrent_orders INT             NOT NULL DEFAULT 10,
    order_prefix          TEXT,
    order_sequence        BIGINT          NOT NULL DEFAULT 0,
    meta_title            TEXT,
    meta_description      TEXT,
    meta_tags             TEXT[]          DEFAULT '{}',
    sort_order            INT             NOT NULL DEFAULT 0,
    is_active             BOOLEAN         NOT NULL DEFAULT true,
    created_at            TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, slug)
);
CREATE INDEX idx_restaurants_tenant_id ON restaurants(tenant_id);
CREATE INDEX idx_restaurants_hub_id ON restaurants(hub_id);
CREATE INDEX idx_restaurants_availability ON restaurants(tenant_id, availability) WHERE is_active = true;
CREATE INDEX idx_restaurants_name_trgm ON restaurants USING gin(name gin_trgm_ops);

CREATE TABLE restaurant_operating_hours (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    restaurant_id   UUID        NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    day_of_week     INT         NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    open_time       TIME        NOT NULL,
    close_time      TIME        NOT NULL,
    is_closed       BOOLEAN     NOT NULL DEFAULT false,
    UNIQUE(restaurant_id, day_of_week)
);

-- ============================================================
-- CATEGORIES
-- ============================================================
CREATE TABLE categories (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    restaurant_id   UUID        REFERENCES restaurants(id) ON DELETE CASCADE,
    name            TEXT        NOT NULL,
    slug            TEXT        NOT NULL,
    image_url       TEXT,
    prep_time       INT         NOT NULL DEFAULT 0,
    sort_order      INT         NOT NULL DEFAULT 0,
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    is_tobacco      BOOLEAN     NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_categories_restaurant_id ON categories(restaurant_id);
CREATE INDEX idx_categories_tenant_id ON categories(tenant_id);

-- ============================================================
-- PRODUCTS
-- ============================================================
CREATE TABLE products (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID            NOT NULL REFERENCES tenants(id),
    restaurant_id   UUID            NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    category_id     UUID            REFERENCES categories(id),
    name            TEXT            NOT NULL,
    slug            TEXT            NOT NULL,
    description     TEXT,
    base_price      NUMERIC(10,2)   NOT NULL,
    vat_rate        NUMERIC(5,2)    NOT NULL DEFAULT 0,
    price_type      price_type      NOT NULL DEFAULT 'flat',
    availability    product_avail   NOT NULL DEFAULT 'available',
    images          TEXT[]          DEFAULT '{}',
    sort_order      INT             NOT NULL DEFAULT 0,
    is_inv_tracked  BOOLEAN         NOT NULL DEFAULT false,
    meta_title      TEXT,
    meta_description TEXT,
    meta_tags       TEXT[]          DEFAULT '{}',
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    UNIQUE(restaurant_id, slug)
);
CREATE INDEX idx_products_restaurant_id ON products(restaurant_id);
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_availability ON products(restaurant_id, availability);
CREATE INDEX idx_products_name_trgm ON products USING gin(name gin_trgm_ops);

CREATE TABLE product_variants (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id  UUID    NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    tenant_id   UUID    NOT NULL REFERENCES tenants(id),
    title       TEXT    NOT NULL,
    min_select  INT     NOT NULL DEFAULT 1,
    max_select  INT     NOT NULL DEFAULT 1,
    sort_order  INT     NOT NULL DEFAULT 0
);
CREATE INDEX idx_product_variants_product_id ON product_variants(product_id);

CREATE TABLE product_variant_items (
    id          UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    variant_id  UUID            NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    tenant_id   UUID            NOT NULL REFERENCES tenants(id),
    name        TEXT            NOT NULL,
    price       NUMERIC(10,2)   NOT NULL DEFAULT 0,
    is_available BOOLEAN        NOT NULL DEFAULT true,
    sort_order  INT             NOT NULL DEFAULT 0
);
CREATE INDEX idx_variant_items_variant_id ON product_variant_items(variant_id);

CREATE TABLE product_addons (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id  UUID    NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    tenant_id   UUID    NOT NULL REFERENCES tenants(id),
    title       TEXT    NOT NULL,
    min_select  INT     NOT NULL DEFAULT 0,
    max_select  INT     NOT NULL DEFAULT 1,
    sort_order  INT     NOT NULL DEFAULT 0
);

CREATE TABLE product_addon_items (
    id          UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    addon_id    UUID            NOT NULL REFERENCES product_addons(id) ON DELETE CASCADE,
    tenant_id   UUID            NOT NULL REFERENCES tenants(id),
    name        TEXT            NOT NULL,
    price       NUMERIC(10,2)   NOT NULL DEFAULT 0,
    is_available BOOLEAN        NOT NULL DEFAULT true
);

CREATE TABLE product_discounts (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id      UUID            NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    tenant_id       UUID            NOT NULL REFERENCES tenants(id),
    discount_type   discount_type   NOT NULL,
    amount          NUMERIC(10,2)   NOT NULL,
    valid_until     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    UNIQUE(product_id)
);

-- ============================================================
-- INVENTORY
-- ============================================================
CREATE TABLE inventory (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id      UUID            NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    tenant_id       UUID            NOT NULL REFERENCES tenants(id),
    restaurant_id   UUID            NOT NULL REFERENCES restaurants(id),
    stock           INT             NOT NULL DEFAULT 0,
    unit_price      NUMERIC(10,2)   NOT NULL DEFAULT 0,
    reorder_level   INT             NOT NULL DEFAULT 5,
    last_updated_at TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by      UUID            REFERENCES users(id),
    UNIQUE(product_id, restaurant_id)
);
CREATE INDEX idx_inventory_restaurant ON inventory(restaurant_id);

-- ============================================================
-- PROMOS
-- ============================================================
CREATE TABLE promos (
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID            NOT NULL REFERENCES tenants(id),
    code                TEXT            NOT NULL,
    auth_key            TEXT,
    description         TEXT,
    promo_type          promo_type      NOT NULL,
    amount              NUMERIC(10,2)   NOT NULL,
    max_discount_amount NUMERIC(10,2),
    cashback_amount     NUMERIC(10,2)   NOT NULL DEFAULT 0,
    apply_on            promo_apply_on  NOT NULL DEFAULT 'all_items',
    funded_by           promo_funder    NOT NULL DEFAULT 'vendor',
    restaurant_id       UUID            REFERENCES restaurants(id),
    category_ids        UUID[]          DEFAULT '{}',
    restaurant_ids      UUID[]          DEFAULT '{}',
    include_stores      BOOLEAN         NOT NULL DEFAULT false,
    min_order_amount    NUMERIC(10,2)   NOT NULL DEFAULT 0,
    max_usage           INT,
    max_usage_per_user  INT             NOT NULL DEFAULT 1,
    eligible_user_ids   UUID[]          DEFAULT '{}',
    usage_count         INT             NOT NULL DEFAULT 0,
    is_active           BOOLEAN         NOT NULL DEFAULT true,
    start_date          TIMESTAMPTZ     NOT NULL,
    end_date            TIMESTAMPTZ     NOT NULL,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, code)
);
CREATE INDEX idx_promos_tenant_id ON promos(tenant_id);
CREATE INDEX idx_promos_code ON promos(tenant_id, code);
CREATE INDEX idx_promos_active ON promos(tenant_id, is_active, end_date);

CREATE TABLE promo_usages (
    id          UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    promo_id    UUID            NOT NULL REFERENCES promos(id),
    user_id     UUID            NOT NULL REFERENCES users(id),
    order_id    UUID            NOT NULL,
    tenant_id   UUID            NOT NULL REFERENCES tenants(id),
    amount_used NUMERIC(10,2)   NOT NULL,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_promo_usage_order ON promo_usages(promo_id, order_id);
CREATE INDEX idx_promo_usage_user ON promo_usages(promo_id, user_id);

-- ============================================================
-- ORDERS
-- ============================================================
CREATE TABLE orders (
    id                      UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id               UUID            NOT NULL REFERENCES tenants(id),
    order_number            TEXT            NOT NULL,
    customer_id             UUID            NOT NULL REFERENCES users(id),
    rider_id                UUID            REFERENCES users(id),
    hub_id                  UUID            REFERENCES hubs(id),
    status                  order_status    NOT NULL DEFAULT 'pending',
    payment_status          payment_status  NOT NULL DEFAULT 'unpaid',
    payment_method          payment_method  NOT NULL,
    delivery_address        JSONB           NOT NULL DEFAULT '{}',
    customer_name           TEXT            NOT NULL,
    customer_phone          TEXT            NOT NULL,
    customer_area           TEXT,
    geo_lat                 NUMERIC(10,8),
    geo_lng                 NUMERIC(11,8),
    subtotal                NUMERIC(12,2)   NOT NULL DEFAULT 0,
    item_discount           NUMERIC(12,2)   NOT NULL DEFAULT 0,
    promo_discount          NUMERIC(12,2)   NOT NULL DEFAULT 0,
    vat                     NUMERIC(12,2)   NOT NULL DEFAULT 0,
    delivery_charge         NUMERIC(12,2)   NOT NULL DEFAULT 0,
    service_charge          NUMERIC(12,2)   NOT NULL DEFAULT 0,
    total                   NUMERIC(12,2)   NOT NULL DEFAULT 0,
    promo_id                UUID            REFERENCES promos(id),
    promo_snapshot          JSONB,
    customer_note           TEXT,
    rider_note              TEXT,
    internal_note           TEXT,
    platform                platform_source NOT NULL DEFAULT 'web',
    is_priority             BOOLEAN         NOT NULL DEFAULT false,
    rejection_reason        TEXT,
    rejected_by             TEXT,
    confirmed_at            TIMESTAMPTZ,
    preparing_at            TIMESTAMPTZ,
    ready_at                TIMESTAMPTZ,
    picked_at               TIMESTAMPTZ,
    delivered_at            TIMESTAMPTZ,
    cancelled_at            TIMESTAMPTZ,
    assigned_at             TIMESTAMPTZ,
    estimated_delivery_min  INT,
    created_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ,
    UNIQUE(tenant_id, order_number)
);
CREATE INDEX idx_orders_tenant_id ON orders(tenant_id);
CREATE INDEX idx_orders_customer_id ON orders(customer_id);
CREATE INDEX idx_orders_rider_id ON orders(rider_id);
CREATE INDEX idx_orders_status ON orders(tenant_id, status);
CREATE INDEX idx_orders_created_at ON orders(tenant_id, created_at DESC);
CREATE INDEX idx_orders_hub_id ON orders(hub_id);

CREATE TABLE order_items (
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id            UUID            NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    tenant_id           UUID            NOT NULL REFERENCES tenants(id),
    restaurant_id       UUID            NOT NULL REFERENCES restaurants(id),
    product_id          UUID            NOT NULL REFERENCES products(id),
    product_snapshot    JSONB           NOT NULL DEFAULT '{}',
    quantity            INT             NOT NULL DEFAULT 1,
    unit_price          NUMERIC(10,2)   NOT NULL,
    variant_price       NUMERIC(10,2)   NOT NULL DEFAULT 0,
    addon_price         NUMERIC(10,2)   NOT NULL DEFAULT 0,
    subtotal            NUMERIC(10,2)   NOT NULL,
    discount            NUMERIC(10,2)   NOT NULL DEFAULT 0,
    promo_discount      NUMERIC(10,2)   NOT NULL DEFAULT 0,
    vat                 NUMERIC(10,2)   NOT NULL DEFAULT 0,
    total               NUMERIC(10,2)   NOT NULL,
    selected_variants   JSONB           DEFAULT '[]',
    selected_addons     JSONB           DEFAULT '[]',
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_restaurant_id ON order_items(restaurant_id);

CREATE TABLE order_pickups (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID            NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    restaurant_id   UUID            NOT NULL REFERENCES restaurants(id),
    tenant_id       UUID            NOT NULL REFERENCES tenants(id),
    order_number    TEXT            NOT NULL,
    status          pickup_status   NOT NULL DEFAULT 'new',
    items           JSONB           NOT NULL DEFAULT '[]',
    items_total     NUMERIC(12,2)   NOT NULL DEFAULT 0,
    commission_rate NUMERIC(5,2)    NOT NULL DEFAULT 0,
    commission_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    vat             NUMERIC(12,2)   NOT NULL DEFAULT 0,
    confirmed_at    TIMESTAMPTZ,
    ready_at        TIMESTAMPTZ,
    picked_at       TIMESTAMPTZ,
    rejected_at     TIMESTAMPTZ,
    rejection_reason TEXT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_order_pickups_order_id ON order_pickups(order_id);
CREATE INDEX idx_order_pickups_restaurant_id ON order_pickups(restaurant_id);
CREATE INDEX idx_order_pickups_status ON order_pickups(restaurant_id, status);

CREATE TABLE order_timeline (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID        NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    tenant_id   UUID        NOT NULL REFERENCES tenants(id),
    event_type  TEXT        NOT NULL,
    old_status  TEXT,
    new_status  TEXT,
    message     TEXT        NOT NULL,
    actor_id    UUID        REFERENCES users(id),
    actor_type  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_order_timeline_order_id ON order_timeline(order_id);

-- ============================================================
-- PAYMENTS
-- ============================================================
CREATE TABLE payment_transactions (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID        NOT NULL REFERENCES tenants(id),
    order_id            UUID        NOT NULL REFERENCES orders(id),
    user_id             UUID        NOT NULL REFERENCES users(id),
    method              payment_method NOT NULL,
    status              txn_status  NOT NULL DEFAULT 'pending',
    amount              NUMERIC(12,2) NOT NULL,
    currency            TEXT        NOT NULL DEFAULT 'BDT',
    gateway_txn_id      TEXT,
    gateway_response    JSONB       DEFAULT '{}',
    ip_address          TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_payment_txn_order_id ON payment_transactions(order_id);
CREATE INDEX idx_payment_txn_tenant_id ON payment_transactions(tenant_id);
CREATE UNIQUE INDEX idx_payment_txn_gateway ON payment_transactions(gateway_txn_id) WHERE gateway_txn_id IS NOT NULL;

CREATE TABLE refunds (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID        NOT NULL REFERENCES tenants(id),
    order_id            UUID        NOT NULL REFERENCES orders(id),
    transaction_id      UUID        REFERENCES payment_transactions(id),
    amount              NUMERIC(12,2) NOT NULL,
    reason              TEXT        NOT NULL,
    status              refund_status NOT NULL DEFAULT 'pending',
    processed_by        UUID        REFERENCES users(id),
    processed_at        TIMESTAMPTZ,
    gateway_refund_id   TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- WALLET
-- ============================================================
CREATE TABLE wallet_transactions (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID            NOT NULL REFERENCES users(id),
    tenant_id       UUID            NOT NULL REFERENCES tenants(id),
    type            wallet_type     NOT NULL,
    source          wallet_source   NOT NULL,
    amount          NUMERIC(12,2)   NOT NULL,
    balance_after   NUMERIC(12,2)   NOT NULL,
    order_id        UUID            REFERENCES orders(id),
    note            TEXT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_wallet_user_id ON wallet_transactions(user_id, created_at DESC);

-- ============================================================
-- RIDERS
-- ============================================================
CREATE TABLE riders (
    user_id         UUID            PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    tenant_id       UUID            NOT NULL REFERENCES tenants(id),
    hub_id          UUID            REFERENCES hubs(id),
    is_available    BOOLEAN         NOT NULL DEFAULT false,
    is_on_duty      BOOLEAN         NOT NULL DEFAULT false,
    vehicle_type    vehicle_type    NOT NULL DEFAULT 'motorcycle',
    nid_number      TEXT,
    license_number  TEXT,
    balance         NUMERIC(12,2)   NOT NULL DEFAULT 0,
    total_earnings  NUMERIC(12,2)   NOT NULL DEFAULT 0,
    order_count     INT             NOT NULL DEFAULT 0,
    rating          NUMERIC(3,2)    NOT NULL DEFAULT 5.0,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_riders_tenant_hub ON riders(tenant_id, hub_id);
CREATE INDEX idx_riders_available ON riders(tenant_id, is_available, is_on_duty) WHERE is_available = true;

CREATE TABLE rider_locations (
    rider_id    UUID            PRIMARY KEY REFERENCES riders(user_id) ON DELETE CASCADE,
    tenant_id   UUID            NOT NULL REFERENCES tenants(id),
    geo_lat     NUMERIC(10,8)   NOT NULL,
    geo_lng     NUMERIC(11,8)   NOT NULL,
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE TABLE rider_travel_logs (
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    rider_id            UUID            NOT NULL REFERENCES riders(user_id),
    tenant_id           UUID            NOT NULL REFERENCES tenants(id),
    order_id            UUID            REFERENCES orders(id),
    geo_lat             NUMERIC(10,8)   NOT NULL,
    geo_lng             NUMERIC(11,8)   NOT NULL,
    subject             rider_subject   NOT NULL,
    distance_from_prev  NUMERIC(10,3),
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_rider_travel_rider_id ON rider_travel_logs(rider_id, created_at DESC);

CREATE TABLE rider_attendance (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    rider_id        UUID        NOT NULL REFERENCES riders(user_id),
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    date            DATE        NOT NULL,
    checked_in_at   TIMESTAMPTZ,
    checked_out_at  TIMESTAMPTZ,
    total_hours     NUMERIC(5,2),
    total_distance  NUMERIC(10,3),
    total_orders    INT         NOT NULL DEFAULT 0,
    UNIQUE(rider_id, date)
);

CREATE TABLE rider_penalties (
    id          UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    rider_id    UUID            NOT NULL REFERENCES riders(user_id),
    tenant_id   UUID            NOT NULL REFERENCES tenants(id),
    order_id    UUID            REFERENCES orders(id),
    reason      TEXT            NOT NULL,
    amount      NUMERIC(10,2)   NOT NULL,
    status      penalty_status  NOT NULL DEFAULT 'pending',
    appeal_note TEXT,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- ============================================================
-- INVOICES
-- ============================================================
CREATE TABLE invoices (
    id                      UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id               UUID            NOT NULL REFERENCES tenants(id),
    restaurant_id           UUID            NOT NULL REFERENCES restaurants(id),
    period_start            DATE            NOT NULL,
    period_end              DATE            NOT NULL,
    total_sales             NUMERIC(12,2)   NOT NULL DEFAULT 0,
    item_discount           NUMERIC(12,2)   NOT NULL DEFAULT 0,
    promo_discount          NUMERIC(12,2)   NOT NULL DEFAULT 0,
    vat_collected           NUMERIC(12,2)   NOT NULL DEFAULT 0,
    commission_rate         NUMERIC(5,2)    NOT NULL,
    commission_amount       NUMERIC(12,2)   NOT NULL DEFAULT 0,
    penalty                 NUMERIC(12,2)   NOT NULL DEFAULT 0,
    adjustment              NUMERIC(12,2)   NOT NULL DEFAULT 0,
    net_payable             NUMERIC(12,2)   NOT NULL DEFAULT 0,
    order_count             INT             NOT NULL DEFAULT 0,
    rejected_order_count    INT             NOT NULL DEFAULT 0,
    status                  invoice_status  NOT NULL DEFAULT 'draft',
    paid_by                 UUID            REFERENCES users(id),
    paid_at                 TIMESTAMPTZ,
    note                    TEXT,
    created_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    UNIQUE(restaurant_id, period_start, period_end)
);
CREATE INDEX idx_invoices_tenant_id ON invoices(tenant_id);
CREATE INDEX idx_invoices_restaurant_id ON invoices(restaurant_id);
CREATE INDEX idx_invoices_status ON invoices(tenant_id, status);

-- ============================================================
-- ORDER ISSUES
-- ============================================================
CREATE TABLE order_issues (
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID            NOT NULL REFERENCES tenants(id),
    order_id            UUID            NOT NULL REFERENCES orders(id),
    type                issue_type      NOT NULL,
    reported_by         UUID            NOT NULL REFERENCES users(id),
    accountable_party   accountable,
    details             TEXT            NOT NULL,
    images              TEXT[]          DEFAULT '{}',
    refund_items        JSONB           DEFAULT '[]',
    refund_amount       NUMERIC(12,2)   NOT NULL DEFAULT 0,
    refund_status       refund_status   NOT NULL DEFAULT 'pending',
    restaurant_penalty  JSONB           DEFAULT '{}',
    rider_penalty       NUMERIC(10,2)   NOT NULL DEFAULT 0,
    status              issue_status    NOT NULL DEFAULT 'open',
    resolved_by         UUID            REFERENCES users(id),
    resolved_at         TIMESTAMPTZ,
    resolution_note     TEXT,
    messages            JSONB           NOT NULL DEFAULT '[]',
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_issues_order_id ON order_issues(order_id);
CREATE INDEX idx_issues_tenant_id ON order_issues(tenant_id, status);

-- ============================================================
-- CONTENT: BANNERS, SECTIONS, STORIES
-- ============================================================
CREATE TABLE banners (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID        NOT NULL REFERENCES tenants(id),
    title       TEXT        NOT NULL,
    image_url   TEXT        NOT NULL,
    link_type   TEXT,
    link_value  TEXT,
    platform    TEXT        NOT NULL DEFAULT 'all',
    sort_order  INT         NOT NULL DEFAULT 0,
    is_active   BOOLEAN     NOT NULL DEFAULT true,
    hub_ids     UUID[]      DEFAULT '{}',
    area_names  TEXT[]      DEFAULT '{}',
    valid_from  TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_banners_tenant_id ON banners(tenant_id, is_active);

CREATE TABLE sections (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID        NOT NULL REFERENCES tenants(id),
    title       TEXT        NOT NULL,
    type        TEXT        NOT NULL,
    item_ids    UUID[]      DEFAULT '{}',
    sort_order  INT         NOT NULL DEFAULT 0,
    is_active   BOOLEAN     NOT NULL DEFAULT true,
    hub_ids     UUID[]      DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_sections_tenant_id ON sections(tenant_id, is_active);

CREATE TABLE stories (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    restaurant_id   UUID        REFERENCES restaurants(id),
    media_url       TEXT        NOT NULL,
    media_type      TEXT        NOT NULL DEFAULT 'image',
    link_type       TEXT,
    link_value      TEXT,
    expires_at      TIMESTAMPTZ NOT NULL,
    sort_order      INT         NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_stories_tenant_id ON stories(tenant_id, expires_at);

-- ============================================================
-- ANALYTICS (DENORMALISED)
-- ============================================================
CREATE TABLE order_analytics (
    id                      UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id               UUID            NOT NULL REFERENCES tenants(id),
    order_id                UUID            NOT NULL REFERENCES orders(id),
    restaurant_ids          UUID[]          DEFAULT '{}',
    customer_id             UUID            NOT NULL,
    rider_id                UUID,
    hub_id                  UUID,
    customer_area           TEXT,
    payment_method          TEXT,
    platform                TEXT,
    promo_code              TEXT,
    subtotal                NUMERIC(12,2)   NOT NULL DEFAULT 0,
    delivery_charge         NUMERIC(12,2)   NOT NULL DEFAULT 0,
    discount                NUMERIC(12,2)   NOT NULL DEFAULT 0,
    promo_discount          NUMERIC(12,2)   NOT NULL DEFAULT 0,
    vat                     NUMERIC(12,2)   NOT NULL DEFAULT 0,
    total                   NUMERIC(12,2)   NOT NULL DEFAULT 0,
    commission              NUMERIC(12,2)   NOT NULL DEFAULT 0,
    confirmation_time_s     INT,
    preparation_time_s      INT,
    pickup_to_delivery_s    INT,
    total_delivery_time_s   INT,
    final_status            TEXT            NOT NULL,
    rejection_reason        TEXT,
    order_date              DATE            NOT NULL,
    order_hour              INT             NOT NULL,
    order_day_of_week       INT             NOT NULL,
    order_month             INT             NOT NULL,
    order_year              INT             NOT NULL,
    completed_at            TIMESTAMPTZ,
    created_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_analytics_tenant_date ON order_analytics(tenant_id, order_date);
CREATE INDEX idx_analytics_restaurant ON order_analytics USING gin(restaurant_ids);
CREATE INDEX idx_analytics_customer ON order_analytics(customer_id);

-- ============================================================
-- NOTIFICATIONS
-- ============================================================
CREATE TABLE notifications (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID        NOT NULL REFERENCES tenants(id),
    user_id     UUID        NOT NULL REFERENCES users(id),
    type        TEXT        NOT NULL,
    title       TEXT        NOT NULL,
    body        TEXT        NOT NULL,
    data        JSONB       DEFAULT '{}',
    is_read     BOOLEAN     NOT NULL DEFAULT false,
    read_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_notifications_user ON notifications(user_id, is_read, created_at DESC);

-- ============================================================
-- SEARCH LOGS
-- ============================================================
CREATE TABLE search_logs (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID        NOT NULL REFERENCES tenants(id),
    user_id     UUID        REFERENCES users(id),
    query       TEXT        NOT NULL,
    result_count INT        NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_search_logs_tenant ON search_logs(tenant_id, created_at DESC);

-- ============================================================
-- RATINGS
-- ============================================================
CREATE TABLE ratings (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    order_id        UUID        NOT NULL REFERENCES orders(id),
    customer_id     UUID        NOT NULL REFERENCES users(id),
    restaurant_id   UUID        NOT NULL REFERENCES restaurants(id),
    rider_id        UUID        REFERENCES users(id),
    restaurant_rating INT       CHECK (restaurant_rating BETWEEN 1 AND 5),
    rider_rating    INT         CHECK (rider_rating BETWEEN 1 AND 5),
    comment         TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(order_id)
);
CREATE INDEX idx_ratings_restaurant ON ratings(restaurant_id);

-- ============================================================
-- AUDIT LOG
-- ============================================================
CREATE TABLE audit_logs (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID        REFERENCES tenants(id),
    user_id     UUID        REFERENCES users(id),
    action      TEXT        NOT NULL,
    entity_type TEXT        NOT NULL,
    entity_id   TEXT        NOT NULL,
    old_values  JSONB,
    new_values  JSONB,
    ip_address  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id, created_at DESC);
```

---

## 9.3 Key Indexes Summary

| Table | Index Purpose |
|-------|---------------|
| All tables | `tenant_id` — primary isolation filter |
| `orders` | `(tenant_id, status)`, `(customer_id)`, `(rider_id)`, `created_at DESC` |
| `order_pickups` | `(restaurant_id, status)` — live restaurant order board |
| `products` | `(restaurant_id, availability)`, trigram on `name` for search |
| `restaurants` | `(tenant_id, availability)`, trigram on `name` |
| `promos` | `(tenant_id, code)`, `(tenant_id, is_active, end_date)` |
| `notifications` | `(user_id, is_read, created_at DESC)` |
| `order_analytics` | `(tenant_id, order_date)` for fast aggregations |

---

## 9.4 SQLC Configuration

```yaml
# sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/db/queries"
    schema: "internal/db/migrations"
    gen:
      go:
        package: "db"
        out: "internal/db/sqlc"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
        overrides:
          - db_type: "uuid"
            go_type: "github.com/google/uuid.UUID"
          - db_type: "numeric"
            go_type: "github.com/shopspring/decimal.Decimal"
          - db_type: "timestamptz"
            go_type: "time.Time"
```

---

## 9.5 Critical Production Additions (Recommended Before Launch)

The base schema is strong, but world-class SaaS reliability requires explicit support tables for idempotency, event durability, and financial auditability.

```sql
-- Idempotent writes (checkout, payment-initiation, refunds)
CREATE TABLE idempotency_keys (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID        NOT NULL REFERENCES tenants(id),
    user_id             UUID        REFERENCES users(id),
    endpoint            TEXT        NOT NULL,
    idempotency_key     TEXT        NOT NULL,
    request_hash        TEXT        NOT NULL,
    response_code       INT         NOT NULL,
    response_body       JSONB       NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at          TIMESTAMPTZ NOT NULL,
    UNIQUE(tenant_id, endpoint, idempotency_key)
);
CREATE INDEX idx_idempotency_expires_at ON idempotency_keys(expires_at);

-- Durable domain-event outbox for reliable async side effects
CREATE TABLE outbox_events (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID        REFERENCES tenants(id),
    aggregate_type      TEXT        NOT NULL,  -- order, payment, invoice
    aggregate_id        UUID        NOT NULL,
    event_type          TEXT        NOT NULL,
    payload             JSONB       NOT NULL,
    status              TEXT        NOT NULL DEFAULT 'pending', -- pending,processing,processed,failed
    retry_count         INT         NOT NULL DEFAULT 0,
    next_retry_at       TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at        TIMESTAMPTZ
);
CREATE INDEX idx_outbox_pending ON outbox_events(status, next_retry_at);

-- Optional double-entry ledger for strict financial correctness
CREATE TABLE ledger_accounts (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID        REFERENCES tenants(id),
    account_code        TEXT        NOT NULL,  -- e.g. CUSTOMER_WALLET, PLATFORM_COMMISSION
    account_name        TEXT        NOT NULL,
    currency            TEXT        NOT NULL DEFAULT 'BDT',
    is_active           BOOLEAN     NOT NULL DEFAULT true,
    UNIQUE(tenant_id, account_code)
);

CREATE TABLE ledger_entries (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID        REFERENCES tenants(id),
    reference_type      TEXT        NOT NULL,  -- order,payout,refund,adjustment
    reference_id        UUID        NOT NULL,
    account_id          UUID        NOT NULL REFERENCES ledger_accounts(id),
    direction           TEXT        NOT NULL CHECK (direction IN ('debit','credit')),
    amount              NUMERIC(18,2) NOT NULL CHECK (amount > 0),
    currency            TEXT        NOT NULL DEFAULT 'BDT',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_ledger_ref ON ledger_entries(reference_type, reference_id);
```

### Constraint hardening checklist
- Add `CHECK (amount >= 0)` for all monetary amount columns that represent totals/fees.
- Add explicit `CHECK (end_date >= start_date)` for promos and invoices.
- Add `CHECK (total = subtotal - item_discount - promo_discount + vat + delivery_charge + service_charge)` via generated column or validation trigger.
- Add `UNIQUE(order_id, status)` or event-sequencing protections where duplicate state entries must be blocked.
