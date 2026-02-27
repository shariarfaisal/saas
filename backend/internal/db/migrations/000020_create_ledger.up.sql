-- ============================================================
-- 000020_create_ledger.up.sql
-- Ledger accounts and entries for financial auditability
-- ============================================================

-- ---- Ledger Account Types ----
CREATE TYPE ledger_account_type AS ENUM (
    'asset', 'liability', 'revenue', 'expense'
);

CREATE TYPE ledger_entry_type AS ENUM (
    'order_revenue', 'commission', 'refund', 'wallet_credit', 'wallet_debit',
    'vendor_payout', 'delivery_fee', 'penalty', 'adjustment'
);

-- ---- Ledger Accounts ----
CREATE TABLE ledger_accounts (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    code            TEXT            NOT NULL UNIQUE,
    name            TEXT            NOT NULL,
    account_type    ledger_account_type NOT NULL,
    description     TEXT,
    is_system       BOOLEAN         NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- ---- Ledger Entries (Append-Only) ----
CREATE TABLE ledger_entries (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID            REFERENCES tenants(id),
    account_id      UUID            NOT NULL REFERENCES ledger_accounts(id),
    entry_type      ledger_entry_type NOT NULL,
    reference_type  TEXT            NOT NULL,
    reference_id    UUID            NOT NULL,
    debit           NUMERIC(14,2)   NOT NULL DEFAULT 0.00,
    credit          NUMERIC(14,2)   NOT NULL DEFAULT 0.00,
    balance_after   NUMERIC(14,2)   NOT NULL,
    description     TEXT            NOT NULL,
    metadata        JSONB           NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_entries_account ON ledger_entries(account_id, created_at DESC);
CREATE INDEX idx_ledger_entries_reference ON ledger_entries(reference_type, reference_id);
CREATE INDEX idx_ledger_entries_tenant ON ledger_entries(tenant_id, created_at DESC);
