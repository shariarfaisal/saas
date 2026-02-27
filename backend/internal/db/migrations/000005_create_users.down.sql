-- ============================================================
-- 000005_create_users.down.sql
-- ============================================================

DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS otp_verifications;
DROP TABLE IF EXISTS user_addresses;
DROP TABLE IF EXISTS users;
