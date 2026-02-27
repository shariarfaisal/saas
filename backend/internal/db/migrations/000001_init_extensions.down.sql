-- 0001_init_extensions.down.sql
-- Remove extensions (use with caution)

DROP EXTENSION IF EXISTS "pg_trgm";
DROP EXTENSION IF EXISTS "pgcrypto";
