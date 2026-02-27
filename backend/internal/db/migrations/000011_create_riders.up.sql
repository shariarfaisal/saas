-- ============================================================
-- 000011_create_riders.up.sql
-- Rider profiles, real-time location, history, attendance, earnings
-- ============================================================

-- ---- Riders ----
-- Extends the users table for rider-specific data.
-- user_id is both PK and FK (one-to-one with users where role = 'rider').
CREATE TABLE riders (
    id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID          NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    tenant_id           UUID          NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    hub_id              UUID          REFERENCES hubs(id),

    -- Vehicle
    vehicle_type        vehicle_type  NOT NULL DEFAULT 'motorcycle',
    vehicle_registration TEXT,

    -- Identity verification
    license_number      TEXT,
    nid_number          TEXT,
    nid_verified        BOOLEAN       NOT NULL DEFAULT false,

    -- Status
    is_available        BOOLEAN       NOT NULL DEFAULT false,
    is_on_duty          BOOLEAN       NOT NULL DEFAULT false,

    -- Stats (denormalized; updated after each delivery)
    total_order_count   INT           NOT NULL DEFAULT 0,
    total_earnings      NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    pending_balance     NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    rating_avg          NUMERIC(3,2)  NOT NULL DEFAULT 5.00,
    rating_count        INT           NOT NULL DEFAULT 0,

    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_riders_tenant_id  ON riders(tenant_id);
CREATE INDEX idx_riders_hub_id     ON riders(hub_id)                    WHERE hub_id IS NOT NULL;
CREATE INDEX idx_riders_available  ON riders(hub_id, is_available)
    WHERE is_available = true AND is_on_duty = true;

CREATE TRIGGER trg_riders_updated_at
    BEFORE UPDATE ON riders
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Rider Real-Time Location ----
-- Single row per rider, replaced on each GPS update.
CREATE TABLE rider_locations (
    rider_id        UUID          PRIMARY KEY REFERENCES riders(id) ON DELETE CASCADE,
    tenant_id       UUID          NOT NULL REFERENCES tenants(id),
    geo_lat         NUMERIC(10,8) NOT NULL,
    geo_lng         NUMERIC(11,8) NOT NULL,
    heading         NUMERIC(5,2),                -- degrees 0–360
    speed_kmh       NUMERIC(5,2),
    accuracy_meters NUMERIC(8,2),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rider_locations_tenant ON rider_locations(tenant_id, updated_at DESC);

-- ---- Rider Location History ----
-- Append-only GPS trail per rider per shift. order_id FK added in 000018.
CREATE TABLE rider_location_history (
    id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    rider_id            UUID          NOT NULL REFERENCES riders(id) ON DELETE CASCADE,
    tenant_id           UUID          NOT NULL REFERENCES tenants(id),
    order_id            UUID,                                               -- FK added in 000018
    geo_lat             NUMERIC(10,8) NOT NULL,
    geo_lng             NUMERIC(11,8) NOT NULL,
    event_type          rider_subject NOT NULL,
    distance_from_prev_km NUMERIC(8,3),
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rider_location_history_rider_id ON rider_location_history(rider_id, created_at DESC);

-- ---- Rider Attendance ----
-- Daily check-in / check-out records.
CREATE TABLE rider_attendance (
    id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    rider_id            UUID          NOT NULL REFERENCES riders(id) ON DELETE CASCADE,
    tenant_id           UUID          NOT NULL REFERENCES tenants(id),
    work_date           DATE          NOT NULL,
    checked_in_at       TIMESTAMPTZ,
    checked_out_at      TIMESTAMPTZ,
    total_hours         NUMERIC(5,2),
    total_distance_km   NUMERIC(10,3) NOT NULL DEFAULT 0.00,
    completed_orders    INT           NOT NULL DEFAULT 0,
    cancelled_orders    INT           NOT NULL DEFAULT 0,
    earnings            NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE(rider_id, work_date)
);

CREATE INDEX idx_rider_attendance_tenant ON rider_attendance(tenant_id, work_date DESC);

CREATE TRIGGER trg_rider_attendance_updated_at
    BEFORE UPDATE ON rider_attendance
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Rider Earnings (per Order) ----
-- order_id and payout_id FKs added in 000018.
CREATE TABLE rider_earnings (
    id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    rider_id        UUID          NOT NULL REFERENCES riders(id) ON DELETE CASCADE,
    tenant_id       UUID          NOT NULL REFERENCES tenants(id),
    order_id        UUID          NOT NULL,                                 -- FK added in 000018
    base_earning    NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    distance_bonus  NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    peak_bonus      NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    tip_amount      NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    total_earning   NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    is_paid_out     BOOLEAN       NOT NULL DEFAULT false,
    payout_id       UUID,                                                   -- FK added in 000018
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rider_earnings_rider_id ON rider_earnings(rider_id, is_paid_out);
CREATE INDEX idx_rider_earnings_order_id ON rider_earnings(order_id);

-- ---- Rider Penalties ----
-- order_id FK added in 000018.
CREATE TABLE rider_penalties (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    rider_id        UUID            NOT NULL REFERENCES riders(id) ON DELETE CASCADE,
    tenant_id       UUID            NOT NULL REFERENCES tenants(id),
    order_id        UUID,                                                   -- FK added in 000018
    issue_id        UUID            REFERENCES order_issues(id),
    reason          TEXT            NOT NULL,
    amount          NUMERIC(10,2)   NOT NULL CHECK (amount > 0),
    status          penalty_status  NOT NULL DEFAULT 'pending',
    appeal_note     TEXT,
    appealed_at     TIMESTAMPTZ,
    cleared_at      TIMESTAMPTZ,
    cleared_by      UUID            REFERENCES users(id),
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rider_penalties_rider_id ON rider_penalties(rider_id, status);

CREATE TRIGGER trg_rider_penalties_updated_at
    BEFORE UPDATE ON rider_penalties
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Rider Payouts ----
CREATE TABLE rider_payouts (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    rider_id         UUID          NOT NULL REFERENCES riders(id),
    tenant_id        UUID          NOT NULL REFERENCES tenants(id),
    amount           NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    earnings_from    DATE          NOT NULL,
    earnings_to      DATE          NOT NULL,
    payment_method   TEXT          NOT NULL DEFAULT 'bkash',
    payment_reference TEXT,
    status           payout_status NOT NULL DEFAULT 'pending',
    processed_by     UUID          REFERENCES users(id),
    processed_at     TIMESTAMPTZ,
    note             TEXT,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rider_payouts_rider_id ON rider_payouts(rider_id, status);

CREATE TRIGGER trg_rider_payouts_updated_at
    BEFORE UPDATE ON rider_payouts
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();
