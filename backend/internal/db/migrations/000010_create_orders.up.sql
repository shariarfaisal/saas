-- ============================================================
-- 000010_create_orders.up.sql
-- Orders, items, pickups, timeline events, issues
-- ============================================================

-- ---- Orders ----
CREATE TABLE orders (
    id                          UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                   UUID            NOT NULL REFERENCES tenants(id),
    order_number                TEXT            NOT NULL,         -- e.g. "KBC-001234"
    customer_id                 UUID            NOT NULL REFERENCES users(id),
    rider_id                    UUID            REFERENCES users(id),
    hub_id                      UUID            REFERENCES hubs(id),

    -- Lifecycle status
    status                      order_status    NOT NULL DEFAULT 'pending',
    payment_status              payment_status  NOT NULL DEFAULT 'unpaid',
    payment_method              payment_method  NOT NULL,
    platform                    platform_source NOT NULL DEFAULT 'web',

    -- Delivery snapshot (frozen at order creation)
    delivery_address_id         UUID            REFERENCES user_addresses(id),
    delivery_address            JSONB           NOT NULL,         -- {line1, line2, area, city, geo_lat, geo_lng}
    delivery_recipient_name     TEXT            NOT NULL,
    delivery_recipient_phone    TEXT            NOT NULL,
    delivery_area               TEXT            NOT NULL,         -- for zone charge lookup
    delivery_geo_lat            NUMERIC(10,8),
    delivery_geo_lng            NUMERIC(11,8),

    -- Financial (all amounts in BDT)
    subtotal                    NUMERIC(12,2)   NOT NULL DEFAULT 0.00,  -- sum of item subtotals
    item_discount_total         NUMERIC(12,2)   NOT NULL DEFAULT 0.00,  -- product-level discounts
    promo_discount_total        NUMERIC(12,2)   NOT NULL DEFAULT 0.00,
    vat_total                   NUMERIC(12,2)   NOT NULL DEFAULT 0.00,
    delivery_charge             NUMERIC(10,2)   NOT NULL DEFAULT 0.00,
    service_fee                 NUMERIC(10,2)   NOT NULL DEFAULT 0.00,
    total_amount                NUMERIC(12,2)   NOT NULL DEFAULT 0.00,  -- final customer-paid amount

    -- Promo applied (promo_id FK added in 000018)
    promo_id                    UUID,
    promo_code                  TEXT,
    promo_snapshot              JSONB,                            -- full promo snapshot at order time

    -- Flags
    is_priority                 BOOLEAN         NOT NULL DEFAULT false,
    is_reorder                  BOOLEAN         NOT NULL DEFAULT false,

    -- Notes
    customer_note               TEXT,
    rider_note                  TEXT,
    internal_note               TEXT,

    -- Cancellation / rejection detail
    cancellation_reason         TEXT,
    cancelled_by                actor_type,
    rejection_reason            TEXT,
    rejected_by                 actor_type,

    -- Timing
    auto_confirm_at             TIMESTAMPTZ,                      -- deadline for auto-confirm job
    estimated_delivery_minutes  INT,
    confirmed_at                TIMESTAMPTZ,
    preparing_at                TIMESTAMPTZ,
    ready_at                    TIMESTAMPTZ,
    picked_at                   TIMESTAMPTZ,
    delivered_at                TIMESTAMPTZ,
    cancelled_at                TIMESTAMPTZ,

    created_at                  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at                  TIMESTAMPTZ,

    UNIQUE(tenant_id, order_number)
);

CREATE INDEX idx_orders_tenant_id      ON orders(tenant_id)                         WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_customer_id    ON orders(customer_id);
CREATE INDEX idx_orders_rider_id       ON orders(rider_id)                           WHERE rider_id IS NOT NULL;
CREATE INDEX idx_orders_hub_id         ON orders(hub_id)                             WHERE hub_id IS NOT NULL;
CREATE INDEX idx_orders_status         ON orders(tenant_id, status)                  WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_payment_status ON orders(tenant_id, payment_status);
CREATE INDEX idx_orders_created_at     ON orders(tenant_id, created_at DESC);
CREATE INDEX idx_orders_number         ON orders(order_number);
CREATE INDEX idx_orders_auto_confirm   ON orders(auto_confirm_at)
    WHERE status = 'created' AND auto_confirm_at IS NOT NULL;

CREATE TRIGGER trg_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Order Items ----
CREATE TABLE order_items (
    id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id            UUID          NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    restaurant_id       UUID          NOT NULL REFERENCES restaurants(id),
    product_id          UUID          NOT NULL REFERENCES products(id),
    tenant_id           UUID          NOT NULL REFERENCES tenants(id),

    -- Product snapshot (frozen at order creation for historical accuracy)
    product_name        TEXT          NOT NULL,
    product_snapshot    JSONB         NOT NULL,

    quantity            INT           NOT NULL CHECK (quantity > 0),

    -- Pricing
    unit_price          NUMERIC(10,2) NOT NULL,                   -- base price at order time
    modifier_price      NUMERIC(10,2) NOT NULL DEFAULT 0.00,      -- sum of selected modifier prices
    item_subtotal       NUMERIC(10,2) NOT NULL,                   -- (unit_price + modifier_price) * qty
    item_discount       NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    item_vat            NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    promo_discount      NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    item_total          NUMERIC(10,2) NOT NULL,                   -- final amount for this line

    -- Selected modifier options (snapshot)
    -- Format: [{"group_id":"uuid","group_name":"Size","option_id":"uuid","option_name":"Large","additional_price":30.00}]
    selected_modifiers  JSONB         NOT NULL DEFAULT '[]',

    special_instructions TEXT,
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_items_order_id      ON order_items(order_id);
CREATE INDEX idx_order_items_product_id    ON order_items(product_id);
CREATE INDEX idx_order_items_restaurant_id ON order_items(restaurant_id);

-- ---- Order Pickups ----
-- One row per restaurant in a (potentially multi-restaurant) order.
CREATE TABLE order_pickups (
    id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id            UUID          NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    restaurant_id       UUID          NOT NULL REFERENCES restaurants(id),
    tenant_id           UUID          NOT NULL REFERENCES tenants(id),

    pickup_number       TEXT          NOT NULL,    -- e.g. "KBC-001234"
    status              pickup_status NOT NULL DEFAULT 'new',

    -- Financial breakdown for this restaurant's portion
    items_subtotal      NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    items_discount      NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    items_vat           NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    items_total         NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    commission_rate     NUMERIC(5,2)  NOT NULL DEFAULT 0.00,
    commission_amount   NUMERIC(12,2) NOT NULL DEFAULT 0.00,

    -- Timing
    confirmed_at        TIMESTAMPTZ,
    preparing_at        TIMESTAMPTZ,
    ready_at            TIMESTAMPTZ,
    picked_at           TIMESTAMPTZ,
    rejected_at         TIMESTAMPTZ,
    rejection_reason    TEXT,

    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    UNIQUE(order_id, restaurant_id)
);

CREATE INDEX idx_order_pickups_order_id       ON order_pickups(order_id);
CREATE INDEX idx_order_pickups_restaurant_id  ON order_pickups(restaurant_id);
CREATE INDEX idx_order_pickups_active         ON order_pickups(restaurant_id, status)
    WHERE status IN ('new', 'confirmed', 'preparing');

CREATE TRIGGER trg_order_pickups_updated_at
    BEFORE UPDATE ON order_pickups
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Order Timeline Events (Audit Trail) ----
CREATE TABLE order_timeline_events (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID            NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    tenant_id       UUID            NOT NULL REFERENCES tenants(id),
    event_type      TEXT            NOT NULL,  -- 'status_changed', 'rider_assigned', 'note_added', etc.
    previous_status order_status,
    new_status      order_status,
    description     TEXT            NOT NULL,
    actor_id        UUID            REFERENCES users(id),
    actor_type      actor_type      NOT NULL DEFAULT 'system',
    metadata        JSONB           NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_timeline_order_id   ON order_timeline_events(order_id);
CREATE INDEX idx_order_timeline_created_at ON order_timeline_events(order_id, created_at DESC);

-- ---- Order Issues / Disputes ----
CREATE TABLE order_issues (
    id                      UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id                UUID          NOT NULL REFERENCES orders(id),
    tenant_id               UUID          NOT NULL REFERENCES tenants(id),
    issue_type              issue_type    NOT NULL,
    reported_by_id          UUID          NOT NULL REFERENCES users(id),
    details                 TEXT          NOT NULL,
    evidence_urls           TEXT[]        NOT NULL DEFAULT '{}',

    -- Accountability
    accountable_party       accountable   NOT NULL DEFAULT 'platform',

    -- Refund
    refund_items            JSONB,                -- [{order_item_id, name, qty, amount}]
    refund_amount           NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    refund_status           refund_status NOT NULL DEFAULT 'pending',

    -- Penalties
    restaurant_penalty_amount NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    rider_penalty_amount      NUMERIC(10,2) NOT NULL DEFAULT 0.00,

    -- Resolution
    status                  issue_status  NOT NULL DEFAULT 'open',
    resolution_note         TEXT,
    resolved_by_id          UUID          REFERENCES users(id),
    resolved_at             TIMESTAMPTZ,

    created_at              TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_issues_order_id     ON order_issues(order_id);
CREATE INDEX idx_order_issues_tenant_status ON order_issues(tenant_id, status);

CREATE TRIGGER trg_order_issues_updated_at
    BEFORE UPDATE ON order_issues
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Order Issue Messages (Threaded Discussion) ----
CREATE TABLE order_issue_messages (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id    UUID        NOT NULL REFERENCES order_issues(id) ON DELETE CASCADE,
    tenant_id   UUID        NOT NULL REFERENCES tenants(id),
    sender_id   UUID        NOT NULL REFERENCES users(id),
    message     TEXT        NOT NULL,
    attachments TEXT[]      NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_issue_messages_issue_id ON order_issue_messages(issue_id);
