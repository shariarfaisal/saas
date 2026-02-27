-- ============================================================
-- 000006_create_geography.up.sql
-- Hubs, coverage areas, delivery zone configuration
-- ============================================================

-- ---- Dispatch Hubs ----
-- A hub is a geographic dispatch zone; restaurants belong to a hub;
-- riders are pooled per hub.
CREATE TABLE hubs (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name            TEXT        NOT NULL,
    code            TEXT,                       -- short code e.g. 'GUL'
    manager_id      UUID        REFERENCES users(id),
    address_line1   TEXT,
    address_line2   TEXT,
    city            TEXT        NOT NULL DEFAULT 'Dhaka',
    geo_lat         NUMERIC(10,8),
    geo_lng         NUMERIC(11,8),
    contact_phone   TEXT,
    contact_email   TEXT,
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    sort_order      INT         NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_hubs_tenant_id ON hubs(tenant_id) WHERE is_active = true;

CREATE TRIGGER trg_hubs_updated_at
    BEFORE UPDATE ON hubs
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Hub Coverage Areas ----
-- Named delivery areas served by a hub, each with a delivery charge.
-- Used for zone-based delivery pricing.
CREATE TABLE hub_coverage_areas (
    id                          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    hub_id                      UUID          NOT NULL REFERENCES hubs(id) ON DELETE CASCADE,
    tenant_id                   UUID          NOT NULL REFERENCES tenants(id),
    name                        TEXT          NOT NULL,
    slug                        TEXT          NOT NULL,        -- normalised area name for matching
    delivery_charge             NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    min_order_amount            NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    estimated_delivery_minutes  INT           NOT NULL DEFAULT 45,
    geo_polygon                 JSONB,                         -- GeoJSON polygon for geo-fencing
    is_active                   BOOLEAN       NOT NULL DEFAULT true,
    sort_order                  INT           NOT NULL DEFAULT 0,
    UNIQUE(hub_id, slug)
);

CREATE INDEX idx_hub_coverage_areas_hub_id    ON hub_coverage_areas(hub_id)    WHERE is_active = true;
CREATE INDEX idx_hub_coverage_areas_tenant_id ON hub_coverage_areas(tenant_id) WHERE is_active = true;

-- ---- Delivery Zone Configuration ----
-- One row per tenant; controls whether zone-based or distance-based pricing is used.
CREATE TABLE delivery_zone_configs (
    id                      UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id               UUID          NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
    model                   delivery_model NOT NULL DEFAULT 'zone_based',
    -- JSON array of distance tiers when model = 'distance_based'
    -- Example: [{"max_km": 3, "charge": 40}, {"max_km": 5, "charge": 60}, {"max_km": null, "charge": 100}]
    distance_tiers          JSONB         NOT NULL DEFAULT '[]',
    free_delivery_threshold NUMERIC(10,2),  -- order amount above which delivery is free
    created_at              TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_delivery_zone_configs_updated_at
    BEFORE UPDATE ON delivery_zone_configs
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();
