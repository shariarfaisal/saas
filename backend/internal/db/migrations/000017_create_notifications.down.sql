-- ============================================================
-- 000017_create_notifications.down.sql
-- ============================================================

DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS notification_preferences;
