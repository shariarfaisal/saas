-- ============================================================
-- 000016_create_analytics.up.sql
-- Denormalized analytics fact table for fast reporting queries.
-- Populated by background job when an order reaches a terminal state.
-- order_id FK added in 000018.
-- ============================================================

CREATE TABLE order_analytics (
    id                      UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id               UUID          NOT NULL REFERENCES tenants(id),
    order_id                UUID          NOT NULL UNIQUE,         -- FK added in 000018

    -- Dimensions
    restaurant_ids          UUID[]        NOT NULL DEFAULT '{}',
    customer_id             UUID          NOT NULL,
    rider_id                UUID,
    hub_id                  UUID,
    delivery_area           TEXT,
    payment_method          TEXT          NOT NULL,
    platform                TEXT          NOT NULL,
    promo_code              TEXT,

    -- Financials
    subtotal                NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    item_discount           NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    promo_discount          NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    delivery_charge         NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    vat_total               NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    total_amount            NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    commission_total        NUMERIC(12,2) NOT NULL DEFAULT 0.00,

    -- SLA timing (seconds; NULL if transition never happened)
    confirmation_duration_s INT,
    preparation_duration_s  INT,
    pickup_to_delivery_s    INT,
    total_fulfillment_s     INT,

    -- Final state
    final_status            TEXT          NOT NULL,
    cancellation_reason     TEXT,

    -- Pre-computed time dimensions for GROUP BY performance
    order_date              DATE          NOT NULL,
    order_hour              SMALLINT      NOT NULL CHECK (order_hour BETWEEN 0 AND 23),
    order_day_of_week       SMALLINT      NOT NULL CHECK (order_day_of_week BETWEEN 0 AND 6),
    order_week              INT           NOT NULL,
    order_month             SMALLINT      NOT NULL CHECK (order_month BETWEEN 1 AND 12),
    order_year              SMALLINT      NOT NULL,

    completed_at            TIMESTAMPTZ,
    created_at              TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_analytics_tenant_date ON order_analytics(tenant_id, order_date DESC);
CREATE INDEX idx_order_analytics_tenant_month ON order_analytics(tenant_id, order_year, order_month);
CREATE INDEX idx_order_analytics_restaurants   ON order_analytics USING gin(restaurant_ids);
CREATE INDEX idx_order_analytics_final_status  ON order_analytics(tenant_id, final_status);
