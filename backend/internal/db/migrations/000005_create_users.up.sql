-- ============================================================
-- 000005_create_users.up.sql
-- Users, addresses, OTP, refresh tokens, idempotency keys
-- ============================================================

-- ---- Users (unified table for all roles) ----
CREATE TABLE users (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID        REFERENCES tenants(id) ON DELETE CASCADE,  -- NULL for platform admins
    phone               TEXT,
    email               TEXT,
    name                TEXT        NOT NULL DEFAULT '',
    password_hash       TEXT,                     -- NULL for OTP-only customers
    role                user_role   NOT NULL DEFAULT 'customer',
    status              user_status NOT NULL DEFAULT 'active',
    gender              gender_type,
    date_of_birth       DATE,
    avatar_url          TEXT,

    -- Device info (for push notifications)
    device_push_token   TEXT,
    device_platform     TEXT,                     -- 'ios', 'android', 'web'
    device_model        TEXT,
    device_app_version  TEXT,

    -- Auth
    last_login_at       TIMESTAMPTZ,
    last_login_ip       INET,
    email_verified_at   TIMESTAMPTZ,
    phone_verified_at   TIMESTAMPTZ,
    two_factor_enabled  BOOLEAN     NOT NULL DEFAULT false,
    two_factor_secret   TEXT,

    -- Referral
    referral_code       TEXT        UNIQUE,
    referred_by_id      UUID        REFERENCES users(id),

    -- Wallet (denormalized balance; source of truth is wallet_transactions)
    wallet_balance      NUMERIC(12,2) NOT NULL DEFAULT 0.00
                        CHECK (wallet_balance >= 0),

    -- Stats (denormalized for fast reads)
    total_order_count   INT         NOT NULL DEFAULT 0,
    total_spent_amount  NUMERIC(14,2) NOT NULL DEFAULT 0.00,
    last_order_at       TIMESTAMPTZ,

    -- Flexible extra data
    metadata            JSONB       NOT NULL DEFAULT '{}',

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,

    CONSTRAINT uniq_users_tenant_phone UNIQUE (tenant_id, phone),
    CONSTRAINT chk_users_has_contact   CHECK (phone IS NOT NULL OR email IS NOT NULL)
);

CREATE INDEX idx_users_tenant_id     ON users(tenant_id)                WHERE deleted_at IS NULL;
CREATE INDEX idx_users_phone         ON users(phone)                    WHERE phone IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_users_email         ON users(email)                    WHERE email IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_users_role          ON users(tenant_id, role)          WHERE deleted_at IS NULL;
CREATE INDEX idx_users_referral_code ON users(referral_code)            WHERE referral_code IS NOT NULL;
CREATE INDEX idx_users_push_token    ON users(device_push_token)        WHERE device_push_token IS NOT NULL AND deleted_at IS NULL;

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- User Delivery Addresses ----
CREATE TABLE user_addresses (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id        UUID        NOT NULL REFERENCES tenants(id),
    label            TEXT        NOT NULL DEFAULT 'Home',  -- 'Home', 'Office', etc.
    recipient_name   TEXT,
    recipient_phone  TEXT,
    address_line1    TEXT        NOT NULL,
    address_line2    TEXT,
    area             TEXT        NOT NULL,
    city             TEXT        NOT NULL DEFAULT 'Dhaka',
    geo_lat          NUMERIC(10,8),
    geo_lng          NUMERIC(11,8),
    geo_display_addr TEXT,                              -- reverse-geocoded full address
    is_default       BOOLEAN     NOT NULL DEFAULT false,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_addresses_user_id   ON user_addresses(user_id);
CREATE INDEX idx_user_addresses_tenant_id ON user_addresses(tenant_id);

CREATE TRIGGER trg_user_addresses_updated_at
    BEFORE UPDATE ON user_addresses
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- OTP Verifications ----
CREATE TABLE otp_verifications (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID        REFERENCES tenants(id),
    phone       TEXT        NOT NULL,
    purpose     TEXT        NOT NULL DEFAULT 'login',  -- 'login', 'register', 'password_reset'
    otp_hash    TEXT        NOT NULL,
    attempts    INT         NOT NULL DEFAULT 0,
    max_attempts INT        NOT NULL DEFAULT 3,
    expires_at  TIMESTAMPTZ NOT NULL,
    verified_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_otp_phone_purpose ON otp_verifications(phone, purpose, expires_at)
    WHERE verified_at IS NULL;

-- ---- Refresh Tokens ----
CREATE TABLE refresh_tokens (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id   UUID        REFERENCES tenants(id),
    token_hash  TEXT        NOT NULL UNIQUE,
    device_info JSONB,
    ip_address  INET,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id    ON refresh_tokens(user_id)    WHERE revoked_at IS NULL;
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at) WHERE revoked_at IS NULL;

-- ---- API Idempotency Keys ----
-- Prevents duplicate order submissions and payment initiations
CREATE TABLE idempotency_keys (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID        REFERENCES tenants(id),
    user_id         UUID        REFERENCES users(id),
    key             TEXT        NOT NULL,
    endpoint        TEXT        NOT NULL,
    request_hash    TEXT        NOT NULL,
    response_status INT,
    response_body   JSONB,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, user_id, key, endpoint)
);

CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys(expires_at);
