-- ============================================================
-- 000012_create_promos.up.sql
-- Promo codes, restrictions, per-user eligibility, usage tracking
-- ============================================================

-- ---- Promos ----
CREATE TABLE promos (
    id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID          NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code                TEXT          NOT NULL,
    title               TEXT          NOT NULL,               -- e.g. "Welcome Offer"
    description         TEXT,

    promo_type          promo_type    NOT NULL,
    discount_amount     NUMERIC(10,2) NOT NULL CHECK (discount_amount > 0),
    max_discount_cap    NUMERIC(10,2),                         -- ceiling for percent-type
    cashback_amount     NUMERIC(10,2) NOT NULL DEFAULT 0.00,  -- wallet credit on use
    funded_by           promo_funder  NOT NULL DEFAULT 'platform',
    applies_to          promo_apply_on NOT NULL DEFAULT 'all_items',

    -- Conditions
    min_order_amount    NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    max_total_uses      INT,                                   -- NULL = unlimited
    max_uses_per_user   INT           NOT NULL DEFAULT 1,
    include_stores      BOOLEAN       NOT NULL DEFAULT false,

    -- Validity
    is_active           BOOLEAN       NOT NULL DEFAULT true,
    starts_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    ends_at             TIMESTAMPTZ,                           -- NULL = no expiry

    -- Denormalized usage stats
    total_uses          INT           NOT NULL DEFAULT 0,
    total_discount_given NUMERIC(14,2) NOT NULL DEFAULT 0.00,

    created_by          UUID          REFERENCES users(id),
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    UNIQUE(tenant_id, code)
);

CREATE INDEX idx_promos_tenant_active ON promos(tenant_id, is_active, ends_at);
CREATE INDEX idx_promos_code          ON promos(tenant_id, code);

CREATE TRIGGER trg_promos_updated_at
    BEFORE UPDATE ON promos
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Promo Restaurant Restrictions ----
-- When applies_to = 'specific_restaurant', only these restaurants are eligible.
CREATE TABLE promo_restaurant_restrictions (
    promo_id      UUID NOT NULL REFERENCES promos(id)       ON DELETE CASCADE,
    restaurant_id UUID NOT NULL REFERENCES restaurants(id)  ON DELETE CASCADE,
    PRIMARY KEY (promo_id, restaurant_id)
);

-- ---- Promo Category Restrictions ----
-- When applies_to = 'category', only items in these categories are eligible.
CREATE TABLE promo_category_restrictions (
    promo_id    UUID NOT NULL REFERENCES promos(id)     ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (promo_id, category_id)
);

-- ---- Promo User Eligibility ----
-- When a promo is targeted to specific users only.
CREATE TABLE promo_user_eligibility (
    promo_id UUID NOT NULL REFERENCES promos(id) ON DELETE CASCADE,
    user_id  UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    PRIMARY KEY (promo_id, user_id)
);

-- ---- Promo Usages ----
-- One row per order that used a promo. order_id FK added in 000018.
CREATE TABLE promo_usages (
    id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    promo_id        UUID          NOT NULL REFERENCES promos(id),
    user_id         UUID          NOT NULL REFERENCES users(id),
    order_id        UUID          NOT NULL,                    -- FK added in 000018
    tenant_id       UUID          NOT NULL REFERENCES tenants(id),
    discount_amount NUMERIC(10,2) NOT NULL,
    cashback_amount NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_promo_usages_promo_id    ON promo_usages(promo_id);
CREATE INDEX idx_promo_usages_user_promo  ON promo_usages(user_id, promo_id);
CREATE UNIQUE INDEX uniq_promo_usages_per_order ON promo_usages(order_id, promo_id);
