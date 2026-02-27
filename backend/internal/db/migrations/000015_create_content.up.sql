-- ============================================================
-- 000015_create_content.up.sql
-- Banners, homepage sections, stories, favourites, reviews
-- ============================================================

-- ---- Promotional Banners ----
CREATE TABLE banners (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID            NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    title           TEXT            NOT NULL,
    subtitle        TEXT,
    image_url       TEXT            NOT NULL,
    mobile_image_url TEXT,
    link_type       link_target_type,
    link_value      TEXT,
    platform        TEXT            NOT NULL DEFAULT 'all',  -- 'web', 'app', 'all'
    sort_order      INT             NOT NULL DEFAULT 0,
    is_active       BOOLEAN         NOT NULL DEFAULT true,
    hub_ids         UUID[]          NOT NULL DEFAULT '{}',   -- empty = show everywhere
    starts_at       TIMESTAMPTZ,
    ends_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_banners_tenant_active ON banners(tenant_id, sort_order) WHERE is_active = true;

CREATE TRIGGER trg_banners_updated_at
    BEFORE UPDATE ON banners
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Homepage Sections ----
-- Curated sections like "Trending Now", "New Arrivals", "Best Sellers".
CREATE TABLE homepage_sections (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    title           TEXT        NOT NULL,
    subtitle        TEXT,
    -- 'restaurants', 'products', 'categories', 'banners'
    content_type    TEXT        NOT NULL,
    -- Explicit list of entity UUIDs to display (for manual curation)
    item_ids        UUID[]      NOT NULL DEFAULT '{}',
    -- Or a dynamic filter rule (for algorithmic sections)
    -- e.g. {"type": "top_rated", "limit": 10, "cuisine": "Bengali"}
    filter_rule     JSONB,
    sort_order      INT         NOT NULL DEFAULT 0,
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    hub_ids         UUID[]      NOT NULL DEFAULT '{}',   -- empty = show everywhere
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_homepage_sections_tenant ON homepage_sections(tenant_id, sort_order) WHERE is_active = true;

CREATE TRIGGER trg_homepage_sections_updated_at
    BEFORE UPDATE ON homepage_sections
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Stories ----
-- Short-lived promotional content (Instagram Stories style).
CREATE TABLE stories (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID            NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    restaurant_id   UUID            REFERENCES restaurants(id) ON DELETE SET NULL,
    title           TEXT,
    media_url       TEXT            NOT NULL,
    media_type      media_type      NOT NULL DEFAULT 'image',
    thumbnail_url   TEXT,
    link_type       link_target_type,
    link_value      TEXT,
    expires_at      TIMESTAMPTZ     NOT NULL,
    sort_order      INT             NOT NULL DEFAULT 0,
    view_count      INT             NOT NULL DEFAULT 0,
    is_active       BOOLEAN         NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stories_tenant_active ON stories(tenant_id, sort_order)
    WHERE is_active = true;

-- ---- User Favourites ----
CREATE TABLE user_favourites (
    user_id         UUID        NOT NULL REFERENCES users(id)         ON DELETE CASCADE,
    restaurant_id   UUID        NOT NULL REFERENCES restaurants(id)   ON DELETE CASCADE,
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, restaurant_id)
);

CREATE INDEX idx_user_favourites_user_id ON user_favourites(user_id);

-- ---- Reviews ----
-- Customer reviews for a restaurant, submitted after delivery.
-- One review per order to prevent duplicate submissions.
-- order_id FK added in 000018.
CREATE TABLE reviews (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID        NOT NULL REFERENCES tenants(id),
    order_id            UUID        NOT NULL,                               -- FK added in 000018
    user_id             UUID        NOT NULL REFERENCES users(id),
    restaurant_id       UUID        NOT NULL REFERENCES restaurants(id),
    restaurant_rating   SMALLINT    NOT NULL CHECK (restaurant_rating BETWEEN 1 AND 5),
    rider_rating        SMALLINT    CHECK (rider_rating BETWEEN 1 AND 5),
    comment             TEXT,
    restaurant_reply    TEXT,
    restaurant_reply_at TIMESTAMPTZ,
    images              TEXT[]      NOT NULL DEFAULT '{}',
    is_published        BOOLEAN     NOT NULL DEFAULT false,                  -- pending moderation
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(order_id, user_id)
);

CREATE INDEX idx_reviews_restaurant_id ON reviews(restaurant_id, is_published);
CREATE INDEX idx_reviews_user_id        ON reviews(user_id);

CREATE TRIGGER trg_reviews_updated_at
    BEFORE UPDATE ON reviews
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();
