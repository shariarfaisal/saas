-- ============================================================
-- 000003_add_enum_types.up.sql
-- Additional ENUM types beyond the base 28 in 000002
-- ============================================================

CREATE TYPE gender_type AS ENUM (
    'male', 'female', 'other', 'prefer_not_to_say'
);

-- Who performed an action (for timeline events, audit logs)
CREATE TYPE actor_type AS ENUM (
    'customer', 'restaurant', 'rider', 'platform_admin', 'system'
);

-- Notification delivery channels
CREATE TYPE notification_channel AS ENUM (
    'push', 'sms', 'email', 'in_app'
);

-- Lifecycle state of a notification
CREATE TYPE notification_status AS ENUM (
    'pending', 'sent', 'delivered', 'failed', 'read'
);

-- Why an inventory record changed
CREATE TYPE inventory_adjustment_reason AS ENUM (
    'opening_stock',
    'purchase',
    'manual_adjustment',
    'order_reserve',
    'order_release',
    'order_consume',
    'damage_loss',
    'stock_return'
);

-- Subscription billing cadence
CREATE TYPE billing_cycle AS ENUM ('monthly', 'annual');

-- Tenant subscription lifecycle state
CREATE TYPE subscription_status AS ENUM (
    'trialing', 'active', 'past_due', 'cancelled'
);

-- What a banner / story / section link points to
CREATE TYPE link_target_type AS ENUM (
    'restaurant', 'product', 'category', 'url', 'promo'
);

-- Media asset type for stories
CREATE TYPE media_type AS ENUM ('image', 'video');

-- Transactional outbox event processing state
CREATE TYPE outbox_event_status AS ENUM (
    'pending', 'processing', 'processed', 'failed', 'dead_letter'
);

-- Rider payout lifecycle state
CREATE TYPE payout_status AS ENUM (
    'pending', 'processing', 'completed', 'failed'
);

-- ============================================================
-- Reusable trigger function — keeps updated_at current
-- ============================================================
CREATE OR REPLACE FUNCTION fn_set_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;
