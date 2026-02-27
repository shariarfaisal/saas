-- ============================================================
-- 000009_create_inventory.up.sql
-- Per-product-per-restaurant stock tracking
-- ============================================================

-- ---- Inventory Items ----
-- One row per product per restaurant (only for is_inv_tracked products).
CREATE TABLE inventory_items (
    id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id          UUID          NOT NULL REFERENCES products(id)     ON DELETE CASCADE,
    restaurant_id       UUID          NOT NULL REFERENCES restaurants(id)  ON DELETE CASCADE,
    tenant_id           UUID          NOT NULL REFERENCES tenants(id),
    stock_qty           INT           NOT NULL DEFAULT 0 CHECK (stock_qty >= 0),
    reserved_qty        INT           NOT NULL DEFAULT 0 CHECK (reserved_qty >= 0),
    cost_price          NUMERIC(10,2),
    reorder_threshold   INT           NOT NULL DEFAULT 5,
    last_restocked_at   TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE(product_id, restaurant_id)
);

CREATE INDEX idx_inventory_items_restaurant_id ON inventory_items(restaurant_id);
CREATE INDEX idx_inventory_items_low_stock
    ON inventory_items(tenant_id, restaurant_id)
    WHERE stock_qty - reserved_qty <= reorder_threshold;

-- ---- Inventory Adjustments (Audit Log) ----
-- Records every stock change with a reason.
-- order_id FK added in 000018 (circular dependency).
CREATE TABLE inventory_adjustments (
    id                  UUID                        PRIMARY KEY DEFAULT gen_random_uuid(),
    inventory_item_id   UUID                        NOT NULL REFERENCES inventory_items(id),
    tenant_id           UUID                        NOT NULL REFERENCES tenants(id),
    restaurant_id       UUID                        NOT NULL REFERENCES restaurants(id),
    order_id            UUID,                                           -- FK added in 000018
    adjustment_type     inventory_adjustment_reason NOT NULL,
    qty_before          INT                         NOT NULL,
    qty_change          INT                         NOT NULL,           -- positive = add, negative = remove
    qty_after           INT                         NOT NULL,
    cost_price          NUMERIC(10,2),
    note                TEXT,
    adjusted_by         UUID                        REFERENCES users(id),
    created_at          TIMESTAMPTZ                 NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inventory_adjustments_item_id ON inventory_adjustments(inventory_item_id);
CREATE INDEX idx_inventory_adjustments_tenant  ON inventory_adjustments(tenant_id, created_at DESC);
