-- ============================================================
-- 000008_create_catalog.up.sql
-- Categories, products, modifier groups/options, discounts
-- ============================================================

-- ---- Categories ----
CREATE TABLE categories (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    restaurant_id   UUID        REFERENCES restaurants(id) ON DELETE CASCADE,  -- NULL = tenant-wide
    parent_id       UUID        REFERENCES categories(id),                     -- for nested categories
    name            TEXT        NOT NULL,
    slug            TEXT        NOT NULL,
    description     TEXT,
    image_url       TEXT,
    icon_url        TEXT,
    extra_prep_time_minutes INT NOT NULL DEFAULT 0,
    is_tobacco      BOOLEAN     NOT NULL DEFAULT false,  -- age-gated category
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    sort_order      INT         NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, restaurant_id, slug)
);

CREATE INDEX idx_categories_tenant_id      ON categories(tenant_id);
CREATE INDEX idx_categories_restaurant_id  ON categories(restaurant_id)  WHERE is_active = true;
CREATE INDEX idx_categories_parent_id      ON categories(parent_id)      WHERE parent_id IS NOT NULL;

CREATE TRIGGER trg_categories_updated_at
    BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Products (Menu Items) ----
CREATE TABLE products (
    id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID          NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    restaurant_id       UUID          NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    category_id         UUID          REFERENCES categories(id),
    name                TEXT          NOT NULL,
    slug                TEXT          NOT NULL,
    description         TEXT,
    base_price          NUMERIC(10,2) NOT NULL CHECK (base_price >= 0),
    vat_rate            NUMERIC(5,2)  NOT NULL DEFAULT 0.00,
    -- Whether this product has modifier groups (quick flag to avoid join)
    has_modifiers       BOOLEAN       NOT NULL DEFAULT false,
    availability        product_avail NOT NULL DEFAULT 'available',
    images              TEXT[]        NOT NULL DEFAULT '{}',
    tags                TEXT[]        NOT NULL DEFAULT '{}',
    is_featured         BOOLEAN       NOT NULL DEFAULT false,
    is_inv_tracked      BOOLEAN       NOT NULL DEFAULT false,
    sort_order          INT           NOT NULL DEFAULT 0,
    -- SEO
    meta_title          TEXT,
    meta_description    TEXT,
    -- Denormalized stats
    rating_avg          NUMERIC(3,2)  NOT NULL DEFAULT 0.00,
    rating_count        INT           NOT NULL DEFAULT 0,
    order_count         INT           NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE(restaurant_id, slug)
);

CREATE INDEX idx_products_restaurant_id ON products(restaurant_id);
CREATE INDEX idx_products_category_id   ON products(category_id);
CREATE INDEX idx_products_availability  ON products(restaurant_id, availability)  WHERE availability = 'available';
CREATE INDEX idx_products_name_trgm     ON products USING gin(name gin_trgm_ops);
CREATE INDEX idx_products_featured      ON products(restaurant_id)                WHERE is_featured = true;
CREATE INDEX idx_products_tenant_id     ON products(tenant_id);

CREATE TRIGGER trg_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Product Modifier Groups ----
-- Unified model for both required choices (variants like "Size") and
-- optional extras (add-ons like "Extra Toppings").
--
-- min_required = 0 → optional (add-on style)
-- min_required ≥ 1 → mandatory choice (variant style)
CREATE TABLE product_modifier_groups (
    id              UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id      UUID    NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    tenant_id       UUID    NOT NULL REFERENCES tenants(id),
    name            TEXT    NOT NULL,           -- e.g. "Size", "Crust", "Extras"
    description     TEXT,
    min_required    INT     NOT NULL DEFAULT 0  CHECK (min_required >= 0),
    max_allowed     INT     NOT NULL DEFAULT 1  CHECK (max_allowed >= 1),
    sort_order      INT     NOT NULL DEFAULT 0,
    CONSTRAINT chk_modifier_group_range CHECK (max_allowed >= min_required)
);

CREATE INDEX idx_modifier_groups_product_id ON product_modifier_groups(product_id);

-- ---- Product Modifier Options ----
-- Individual selectable items within a modifier group.
CREATE TABLE product_modifier_options (
    id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    modifier_group_id   UUID          NOT NULL REFERENCES product_modifier_groups(id) ON DELETE CASCADE,
    product_id          UUID          NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    tenant_id           UUID          NOT NULL REFERENCES tenants(id),
    name                TEXT          NOT NULL,  -- e.g. "Large", "Extra Cheese"
    additional_price    NUMERIC(10,2) NOT NULL DEFAULT 0.00 CHECK (additional_price >= 0),
    is_available        BOOLEAN       NOT NULL DEFAULT true,
    sort_order          INT           NOT NULL DEFAULT 0
);

CREATE INDEX idx_modifier_options_group_id   ON product_modifier_options(modifier_group_id);
CREATE INDEX idx_modifier_options_product_id ON product_modifier_options(product_id);

-- ---- Product Discounts ----
-- Time-limited per-product discounts (restaurant or vendor funded).
CREATE TABLE product_discounts (
    id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id          UUID          NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    restaurant_id       UUID          NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    tenant_id           UUID          NOT NULL REFERENCES tenants(id),
    discount_type       discount_type NOT NULL,
    amount              NUMERIC(10,2) NOT NULL CHECK (amount > 0),
    max_discount_cap    NUMERIC(10,2),               -- ceiling for percent-type discounts
    starts_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    ends_at             TIMESTAMPTZ,                 -- NULL = no expiry
    is_active           BOOLEAN       NOT NULL DEFAULT true,
    created_by          UUID          REFERENCES users(id),
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_product_discounts_product_id ON product_discounts(product_id)              WHERE is_active = true;
CREATE INDEX idx_product_discounts_active      ON product_discounts(restaurant_id, ends_at)  WHERE is_active = true;

CREATE TRIGGER trg_product_discounts_updated_at
    BEFORE UPDATE ON product_discounts
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();
