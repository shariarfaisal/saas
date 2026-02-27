-- ============================================================
-- 000003_add_enum_types.down.sql
-- ============================================================

DROP FUNCTION IF EXISTS fn_set_updated_at();

DROP TYPE IF EXISTS payout_status;
DROP TYPE IF EXISTS outbox_event_status;
DROP TYPE IF EXISTS media_type;
DROP TYPE IF EXISTS link_target_type;
DROP TYPE IF EXISTS subscription_status;
DROP TYPE IF EXISTS billing_cycle;
DROP TYPE IF EXISTS inventory_adjustment_reason;
DROP TYPE IF EXISTS notification_status;
DROP TYPE IF EXISTS notification_channel;
DROP TYPE IF EXISTS actor_type;
DROP TYPE IF EXISTS gender_type;
