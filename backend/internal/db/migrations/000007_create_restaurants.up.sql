-- ============================================================
-- 000007_create_restaurants.up.sql
-- Restaurants, operating hours, staff assignments
-- ============================================================

-- ---- Restaurants ----
CREATE TABLE restaurants (
    id                      UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id               UUID            NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    hub_id                  UUID            REFERENCES hubs(id),
    owner_id                UUID            REFERENCES users(id),           -- tenant_owner / tenant_admin

    -- Identity
    name                    TEXT            NOT NULL,
    slug                    TEXT            NOT NULL,
    type                    restaurant_type NOT NULL DEFAULT 'restaurant',
    description             TEXT,
    short_description       TEXT,

    -- Media
    banner_image_url        TEXT,
    logo_url                TEXT,
    gallery_urls            TEXT[]          NOT NULL DEFAULT '{}',

    -- Contact & Location
    phone                   TEXT,
    email                   TEXT,
    address_line1           TEXT,
    address_line2           TEXT,
    area                    TEXT,
    city                    TEXT            NOT NULL DEFAULT 'Dhaka',
    geo_lat                 NUMERIC(10,8),
    geo_lng                 NUMERIC(11,8),

    -- Tags
    cuisines                TEXT[]          NOT NULL DEFAULT '{}',
    tags                    TEXT[]          NOT NULL DEFAULT '{}',

    -- Financial
    commission_rate         NUMERIC(5,2),                                   -- NULL → use tenant default
    vat_rate                NUMERIC(5,2)    NOT NULL DEFAULT 0.00,
    is_vat_inclusive        BOOLEAN         NOT NULL DEFAULT false,          -- VAT already in product price?
    min_order_amount        NUMERIC(10,2)   NOT NULL DEFAULT 0.00,

    -- Operations
    avg_prep_time_minutes   INT             NOT NULL DEFAULT 20,
    max_concurrent_orders   INT             NOT NULL DEFAULT 10,
    auto_accept_orders      BOOLEAN         NOT NULL DEFAULT false,

    -- Order numbering (per-restaurant sequence)
    order_prefix            TEXT,                                            -- e.g. 'KBC'
    order_sequence          BIGINT          NOT NULL DEFAULT 0,

    -- Discovery
    is_available            BOOLEAN         NOT NULL DEFAULT true,           -- manual open/close toggle
    is_featured             BOOLEAN         NOT NULL DEFAULT false,
    is_active               BOOLEAN         NOT NULL DEFAULT true,
    sort_order              INT             NOT NULL DEFAULT 0,

    -- SEO
    meta_title              TEXT,
    meta_description        TEXT,
    meta_keywords           TEXT[]          NOT NULL DEFAULT '{}',

    -- Denormalized ratings (refreshed by background job)
    rating_avg              NUMERIC(3,2)    NOT NULL DEFAULT 0.00,
    rating_count            INT             NOT NULL DEFAULT 0,
    total_order_count       INT             NOT NULL DEFAULT 0,

    created_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    UNIQUE(tenant_id, slug)
);

CREATE INDEX idx_restaurants_tenant_id ON restaurants(tenant_id)                      WHERE is_active = true;
CREATE INDEX idx_restaurants_hub_id    ON restaurants(hub_id)                         WHERE hub_id IS NOT NULL;
CREATE INDEX idx_restaurants_available ON restaurants(tenant_id, is_available)         WHERE is_active = true;
CREATE INDEX idx_restaurants_featured  ON restaurants(tenant_id, sort_order)           WHERE is_featured = true AND is_active = true;
CREATE INDEX idx_restaurants_name_trgm ON restaurants USING gin(name gin_trgm_ops);

CREATE TRIGGER trg_restaurants_updated_at
    BEFORE UPDATE ON restaurants
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Restaurant Operating Hours ----
-- One row per day of week (0=Sunday … 6=Saturday)
CREATE TABLE restaurant_operating_hours (
    id            UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    restaurant_id UUID    NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    tenant_id     UUID    NOT NULL REFERENCES tenants(id),
    day_of_week   SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    open_time     TIME    NOT NULL,
    close_time    TIME    NOT NULL,
    is_closed     BOOLEAN NOT NULL DEFAULT false,
    UNIQUE(restaurant_id, day_of_week),
    CONSTRAINT chk_operating_hours_times CHECK (
        is_closed = true OR close_time > open_time
    )
);

CREATE INDEX idx_restaurant_hours_restaurant_id ON restaurant_operating_hours(restaurant_id);

-- ---- Restaurant Staff Assignments ----
-- Which restaurant(s) a staff member (manager / staff) is assigned to.
-- A restaurant_manager may manage multiple restaurants.
CREATE TABLE restaurant_staff_assignments (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    restaurant_id UUID        NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id     UUID        NOT NULL REFERENCES tenants(id),
    role          user_role   NOT NULL
                  CHECK (role IN ('restaurant_manager', 'restaurant_staff')),
    assigned_by   UUID        REFERENCES users(id),
    assigned_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(restaurant_id, user_id)
);

CREATE INDEX idx_restaurant_staff_user_id       ON restaurant_staff_assignments(user_id);
CREATE INDEX idx_restaurant_staff_restaurant_id ON restaurant_staff_assignments(restaurant_id);
