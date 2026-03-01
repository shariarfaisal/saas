---
title: Database Migration
description: Guide for creating PostgreSQL migrations and SQLC queries
tags: [database, postgresql, migration, sqlc]
---

# Database Migration & Query Workflow

## Creating a New Migration

### 1. Determine Next Number

Check existing: `ls backend/internal/db/migrations/ | tail -5`
Use next sequential number (e.g., `000019`).

### 2. Create Up Migration

`backend/internal/db/migrations/000019_add_feature_table.up.sql`:

```sql
-- Add feature_flags table for tenant-level feature toggles
CREATE TABLE IF NOT EXISTS feature_flags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    feature_key VARCHAR(100) NOT NULL,
    is_enabled  BOOLEAN NOT NULL DEFAULT false,
    metadata    JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (tenant_id, feature_key)
);

-- Index for tenant lookups
CREATE INDEX idx_feature_flags_tenant ON feature_flags(tenant_id);

-- Auto-update trigger
CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON feature_flags
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

### 3. Create Down Migration

`backend/internal/db/migrations/000019_add_feature_table.down.sql`:

```sql
DROP TABLE IF EXISTS feature_flags;
```

### 4. Run Migration

```bash
cd backend && make migrate-up
```

## Adding SQLC Queries

### 1. Create Query File

`backend/internal/db/queries/feature_flags.sql`:

```sql
-- name: GetFeatureFlag :one
SELECT * FROM feature_flags
WHERE tenant_id = $1 AND feature_key = $2;

-- name: ListFeatureFlags :many
SELECT * FROM feature_flags
WHERE tenant_id = $1
ORDER BY feature_key;

-- name: UpsertFeatureFlag :one
INSERT INTO feature_flags (tenant_id, feature_key, is_enabled, metadata)
VALUES ($1, $2, $3, $4)
ON CONFLICT (tenant_id, feature_key)
DO UPDATE SET is_enabled = $3, metadata = $4
RETURNING *;

-- name: DeleteFeatureFlag :exec
DELETE FROM feature_flags
WHERE id = $1 AND tenant_id = $2;
```

### 2. Regenerate Go Types

```bash
cd backend && make sqlc
```

This generates type-safe Go code in `backend/internal/db/sqlc/`.

### 3. Use in Service

```go
flag, err := s.q.GetFeatureFlag(ctx, sqlc.GetFeatureFlagParams{
    TenantID:   uuidToPgtype(&tenantID),
    FeatureKey: "dark_mode",
})
```

## Checklist

- [ ] Every table has `tenant_id UUID NOT NULL` (unless platform-global)
- [ ] Every table has `created_at` and `updated_at` with defaults
- [ ] User-facing tables have `deleted_at TIMESTAMPTZ` for soft deletes
- [ ] Matching `.down.sql` exists and is reversible
- [ ] Appropriate indexes created
- [ ] `update_updated_at_column()` trigger applied
- [ ] SQLC queries include `tenant_id` in WHERE clauses
- [ ] `make sqlc` runs without errors
- [ ] `make migrate-up` and `make migrate-down` both work
