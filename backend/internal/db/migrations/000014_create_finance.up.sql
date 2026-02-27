-- ============================================================
-- 000014_create_finance.up.sql
-- Settlement invoices (one per restaurant per settlement period)
-- ============================================================

CREATE TABLE invoices (
    id                          UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                   UUID            NOT NULL REFERENCES tenants(id),
    restaurant_id               UUID            NOT NULL REFERENCES restaurants(id),
    invoice_number              TEXT            NOT NULL,

    period_start                DATE            NOT NULL,
    period_end                  DATE            NOT NULL,

    -- Revenue
    gross_sales                 NUMERIC(14,2)   NOT NULL DEFAULT 0.00,   -- sum of delivered order subtotals
    item_discounts              NUMERIC(14,2)   NOT NULL DEFAULT 0.00,   -- product-level discounts
    vendor_promo_discounts      NUMERIC(14,2)   NOT NULL DEFAULT 0.00,   -- vendor-funded promo discounts
    net_sales                   NUMERIC(14,2)   NOT NULL DEFAULT 0.00,   -- gross - item_discounts - vendor_promo_discounts
    vat_collected               NUMERIC(14,2)   NOT NULL DEFAULT 0.00,

    -- Deductions
    commission_rate             NUMERIC(5,2)    NOT NULL,
    commission_amount           NUMERIC(14,2)   NOT NULL DEFAULT 0.00,
    penalty_amount              NUMERIC(14,2)   NOT NULL DEFAULT 0.00,

    -- Admin manual adjustment (positive = extra charge, negative = credit)
    adjustment_amount           NUMERIC(14,2)   NOT NULL DEFAULT 0.00,
    adjustment_note             TEXT,

    -- Settlement amount (positive = platform owes restaurant, negative = restaurant owes platform)
    net_payable                 NUMERIC(14,2)   NOT NULL DEFAULT 0.00,

    -- Order counts
    total_orders                INT             NOT NULL DEFAULT 0,
    delivered_orders            INT             NOT NULL DEFAULT 0,
    cancelled_orders            INT             NOT NULL DEFAULT 0,
    rejected_orders             INT             NOT NULL DEFAULT 0,

    -- Lifecycle
    status                      invoice_status  NOT NULL DEFAULT 'draft',
    generated_by                UUID            REFERENCES users(id),
    finalized_by                UUID            REFERENCES users(id),
    finalized_at                TIMESTAMPTZ,
    paid_by                     UUID            REFERENCES users(id),
    paid_at                     TIMESTAMPTZ,
    payment_reference           TEXT,
    notes                       TEXT,

    created_at                  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    UNIQUE(tenant_id, restaurant_id, period_start, period_end)
);

CREATE UNIQUE INDEX uniq_invoice_number ON invoices(tenant_id, invoice_number);
CREATE INDEX idx_invoices_restaurant_id ON invoices(restaurant_id, period_end DESC);
CREATE INDEX idx_invoices_tenant_status ON invoices(tenant_id, status);

CREATE TRIGGER trg_invoices_updated_at
    BEFORE UPDATE ON invoices
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();
