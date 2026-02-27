-- ============================================================
-- 000013_create_payments.up.sql
-- Payment transactions, refunds, wallet transactions
-- ============================================================

-- ---- Payment Transactions ----
-- One row per payment attempt (may have multiple attempts per order).
-- order_id FK added in 000018.
CREATE TABLE payment_transactions (
    id                      UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id               UUID            NOT NULL REFERENCES tenants(id),
    order_id                UUID            NOT NULL,                   -- FK added in 000018
    user_id                 UUID            NOT NULL REFERENCES users(id),
    payment_method          payment_method  NOT NULL,
    status                  txn_status      NOT NULL DEFAULT 'pending',
    amount                  NUMERIC(12,2)   NOT NULL CHECK (amount > 0),
    currency                TEXT            NOT NULL DEFAULT 'BDT',

    -- Gateway data
    gateway_transaction_id  TEXT,
    gateway_reference_id    TEXT,
    gateway_response        JSONB,
    gateway_fee             NUMERIC(10,2)   NOT NULL DEFAULT 0.00,

    -- Request context
    ip_address              INET,
    user_agent              TEXT,

    -- Idempotency: callback must be processed exactly once
    callback_received_at    TIMESTAMPTZ,

    created_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payment_txns_order_id    ON payment_transactions(order_id);
CREATE INDEX idx_payment_txns_user_id     ON payment_transactions(user_id);
CREATE INDEX idx_payment_txns_gateway_id  ON payment_transactions(gateway_transaction_id)
    WHERE gateway_transaction_id IS NOT NULL;
CREATE INDEX idx_payment_txns_status      ON payment_transactions(tenant_id, status, created_at DESC);

CREATE TRIGGER trg_payment_transactions_updated_at
    BEFORE UPDATE ON payment_transactions
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Refunds ----
-- order_id FK added in 000018.
CREATE TABLE refunds (
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID            NOT NULL REFERENCES tenants(id),
    order_id            UUID            NOT NULL,                       -- FK added in 000018
    transaction_id      UUID            NOT NULL REFERENCES payment_transactions(id),
    issue_id            UUID            REFERENCES order_issues(id),
    amount              NUMERIC(12,2)   NOT NULL CHECK (amount > 0),
    reason              TEXT            NOT NULL,
    status              refund_status   NOT NULL DEFAULT 'pending',
    gateway_refund_id   TEXT,
    approved_by         UUID            REFERENCES users(id),
    approved_at         TIMESTAMPTZ,
    processed_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refunds_order_id       ON refunds(order_id);
CREATE INDEX idx_refunds_transaction_id ON refunds(transaction_id);
CREATE INDEX idx_refunds_status         ON refunds(tenant_id, status);

CREATE TRIGGER trg_refunds_updated_at
    BEFORE UPDATE ON refunds
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Wallet Transactions ----
-- Append-only ledger. order_id FK added in 000018.
-- balance_after provides a running total for audit purposes.
CREATE TABLE wallet_transactions (
    id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID          NOT NULL REFERENCES users(id),
    tenant_id       UUID          NOT NULL REFERENCES tenants(id),
    order_id        UUID,                                               -- FK added in 000018
    type            wallet_type   NOT NULL,
    source          wallet_source NOT NULL,
    amount          NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    balance_after   NUMERIC(12,2) NOT NULL,
    description     TEXT,
    expires_at      TIMESTAMPTZ,                                        -- for cashback expiry (Phase 2)
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wallet_txns_user_id  ON wallet_transactions(user_id, created_at DESC);
CREATE INDEX idx_wallet_txns_order_id ON wallet_transactions(order_id) WHERE order_id IS NOT NULL;
