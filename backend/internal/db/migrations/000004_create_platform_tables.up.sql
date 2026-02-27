-- ============================================================
-- 000004_create_platform_tables.up.sql
-- Subscription plans, tenants, subscriptions, configs
-- ============================================================

-- ---- Subscription Plans ----
CREATE TABLE subscription_plans (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    name             TEXT          NOT NULL,
    slug             TEXT          NOT NULL UNIQUE,
    description      TEXT,
    price_monthly    NUMERIC(12,2) NOT NULL,
    price_annual     NUMERIC(12,2) NOT NULL,
    max_restaurants  INT           DEFAULT 1,                -- NULL = unlimited
    max_riders       INT,                           -- NULL = unlimited
    commission_rate  NUMERIC(5,2)  NOT NULL DEFAULT 10.00,
    features         JSONB         NOT NULL DEFAULT '{}',
    is_active        BOOLEAN       NOT NULL DEFAULT true,
    sort_order       INT           NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

INSERT INTO subscription_plans (name, slug, price_monthly, price_annual, max_restaurants, commission_rate, sort_order)
VALUES
    ('Starter',    'starter',    0.00,    0.00,    1,    12.00, 1),
    ('Growth',     'growth',     2999.00, 29990.00, 5,   10.00, 2),
    ('Enterprise', 'enterprise', 7999.00, 79990.00, NULL, 8.00, 3);

-- ---- Tenants ----
CREATE TABLE tenants (
    id                    UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    slug                  TEXT          NOT NULL UNIQUE,
    name                  TEXT          NOT NULL,
    status                tenant_status NOT NULL DEFAULT 'pending',
    plan                  tenant_plan   NOT NULL DEFAULT 'starter',
    subscription_plan_id  UUID          REFERENCES subscription_plans(id),
    commission_rate       NUMERIC(5,2)  NOT NULL DEFAULT 10.00,
    settings              JSONB         NOT NULL DEFAULT '{}',
    custom_domain         TEXT          UNIQUE,
    logo_url              TEXT,
    favicon_url           TEXT,
    primary_color         TEXT          NOT NULL DEFAULT '#FF6B35',
    secondary_color       TEXT          NOT NULL DEFAULT '#2C3E50',
    contact_email         TEXT          NOT NULL,
    contact_phone         TEXT,
    address               JSONB,                   -- {line1, area, city, country}
    timezone              TEXT          NOT NULL DEFAULT 'Asia/Dhaka',
    currency              TEXT          NOT NULL DEFAULT 'BDT',
    locale                TEXT          NOT NULL DEFAULT 'en',
    created_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenants_slug   ON tenants(slug);
CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_domain ON tenants(custom_domain) WHERE custom_domain IS NOT NULL;

CREATE TRIGGER trg_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Tenant Subscriptions ----
CREATE TABLE tenant_subscriptions (
    id                    UUID                PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             UUID                NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plan_id               UUID                NOT NULL REFERENCES subscription_plans(id),
    billing_cycle         billing_cycle       NOT NULL DEFAULT 'monthly',
    status                subscription_status NOT NULL DEFAULT 'trialing',
    trial_ends_at         TIMESTAMPTZ,
    current_period_start  DATE                NOT NULL,
    current_period_end    DATE                NOT NULL,
    next_billing_date     DATE,
    cancelled_at          TIMESTAMPTZ,
    cancellation_reason   TEXT,
    created_at            TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenant_subscriptions_tenant_id ON tenant_subscriptions(tenant_id);
CREATE INDEX idx_tenant_subscriptions_status    ON tenant_subscriptions(status);

CREATE TRIGGER trg_tenant_subscriptions_updated_at
    BEFORE UPDATE ON tenant_subscriptions
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Platform Configuration ----
CREATE TABLE platform_configs (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    key         TEXT        NOT NULL UNIQUE,
    value       JSONB       NOT NULL,
    description TEXT,
    is_public   BOOLEAN     NOT NULL DEFAULT false,
    updated_by  UUID,                              -- FK to users added in 000018
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_platform_configs_updated_at
    BEFORE UPDATE ON platform_configs
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Per-Tenant Payment Gateway Configuration ----
CREATE TABLE tenant_payment_gateways (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    gateway     TEXT        NOT NULL,              -- 'bkash', 'aamarpay', 'sslcommerz'
    is_enabled  BOOLEAN     NOT NULL DEFAULT false,
    is_test_mode BOOLEAN    NOT NULL DEFAULT true,
    config      JSONB       NOT NULL DEFAULT '{}', -- encrypted credentials stored externally
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, gateway)
);

CREATE TRIGGER trg_tenant_payment_gateways_updated_at
    BEFORE UPDATE ON tenant_payment_gateways
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();
