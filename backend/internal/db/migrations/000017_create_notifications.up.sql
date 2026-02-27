-- ============================================================
-- 000017_create_notifications.up.sql
-- Per-user notification preferences, notification log, outbox
-- ============================================================

-- ---- Notification Preferences ----
CREATE TABLE notification_preferences (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID        NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    tenant_id           UUID        REFERENCES tenants(id),
    -- Channel toggles
    push_enabled        BOOLEAN     NOT NULL DEFAULT true,
    sms_enabled         BOOLEAN     NOT NULL DEFAULT true,
    email_enabled       BOOLEAN     NOT NULL DEFAULT true,
    -- Event category toggles
    order_updates       BOOLEAN     NOT NULL DEFAULT true,
    promotions          BOOLEAN     NOT NULL DEFAULT true,
    system_alerts       BOOLEAN     NOT NULL DEFAULT true,
    invoice_alerts      BOOLEAN     NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_notification_preferences_updated_at
    BEFORE UPDATE ON notification_preferences
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---- Notifications ----
-- Persisted log of all sent notifications (for in-app notification center
-- and delivery tracking).
CREATE TABLE notifications (
    id              UUID                    PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID                    REFERENCES tenants(id),
    user_id         UUID                    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel         notification_channel    NOT NULL,
    title           TEXT                    NOT NULL,
    body            TEXT                    NOT NULL,
    image_url       TEXT,
    -- Deep-link action
    action_type     TEXT,                               -- 'open_order', 'open_promo', 'open_url'
    action_payload  JSONB,
    -- Delivery state
    status          notification_status     NOT NULL DEFAULT 'pending',
    sent_at         TIMESTAMPTZ,
    delivered_at    TIMESTAMPTZ,
    read_at         TIMESTAMPTZ,
    failed_reason   TEXT,
    gateway_message_id TEXT,
    created_at      TIMESTAMPTZ             NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id  ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_pending  ON notifications(status, created_at)    WHERE status = 'pending';
CREATE INDEX idx_notifications_unread   ON notifications(user_id, channel)
    WHERE read_at IS NULL AND status = 'delivered';

-- ---- Transactional Outbox ----
-- Ensures side effects (notifications, analytics sync) are published exactly
-- once even if the process crashes after a DB write but before publishing.
-- Workers poll this table and publish events to Redis pub/sub / job queues.
CREATE TABLE outbox_events (
    id              UUID                PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID                REFERENCES tenants(id),
    aggregate_type  TEXT                NOT NULL,    -- 'order', 'payment', 'user'
    aggregate_id    UUID                NOT NULL,
    event_type      TEXT                NOT NULL,    -- 'order.created', 'payment.succeeded'
    payload         JSONB               NOT NULL,
    status          outbox_event_status NOT NULL DEFAULT 'pending',
    attempts        INT                 NOT NULL DEFAULT 0,
    max_attempts    INT                 NOT NULL DEFAULT 5,
    next_retry_at   TIMESTAMPTZ,
    last_error      TEXT,
    processed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outbox_pending    ON outbox_events(status, next_retry_at)
    WHERE status IN ('pending', 'failed');
CREATE INDEX idx_outbox_aggregate  ON outbox_events(aggregate_type, aggregate_id);
