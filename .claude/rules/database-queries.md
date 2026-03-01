# Database & Queries

## SQLC is Mandatory

All database access goes through SQLC. Never use raw `pgx` queries in services.

**Workflow for new queries:**
1. Write SQL in `backend/internal/db/queries/{entity}.sql`
2. Run `make sqlc` to regenerate Go types
3. Use generated `*sqlc.Queries` methods in repositories/services

**SQLC annotation format:**
```sql
-- name: GetRestaurantByID :one
SELECT * FROM restaurants
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: ListOrdersByTenant :many
SELECT * FROM orders
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
```

## Multi-Tenancy — Every Query Filters by tenant_id

```sql
-- CORRECT
SELECT * FROM restaurants WHERE id = $1 AND tenant_id = $2;

-- WRONG — data leak across tenants
SELECT * FROM restaurants WHERE id = $1;
```

**No exceptions.** Platform-level queries (super-admin) are the only case where tenant_id may be omitted, and those must be in separate admin-specific query files.

## Type Mapping

| PostgreSQL | Go Type |
|-----------|---------|
| `uuid` | `github.com/google/uuid.UUID` |
| `numeric` | `github.com/shopspring/decimal.Decimal` |
| `jsonb` | `encoding/json.RawMessage` |
| `inet` | `net/netip.Addr` |
| `timestamptz` | `pgtype.Timestamptz` |
| nullable columns | `pgtype.*` or `sql.Null*` |

**Type conversion helpers** (already in codebase):
```go
uuidToPgtype(id *uuid.UUID) pgtype.UUID
pgtypeToUUID(p pgtype.UUID) *uuid.UUID
nullString(s *string) sql.NullString
```

## Migration Rules

- Migrations live in `backend/internal/db/migrations/`
- Numbered sequentially: `000001`, `000002`, etc.
- Every `.up.sql` must have a matching `.down.sql`
- Never modify existing migrations — create new ones
- Test with `make migrate-up` and `make migrate-down`

## Query Anti-Patterns

- Never use `SELECT *` in production queries — list columns explicitly
- Never use `COUNT(*)` for generating sequential IDs — use PostgreSQL sequences
- Never hardcode values (commission rates, delivery charges) — read from DB/config
- Always use `deleted_at IS NULL` for soft-deleted entities
- Always use parameterized queries (SQLC handles this)
- Use `FOR UPDATE` when reading before writing in transactions

## Existing Schema Groups (28 ENUM types, 18+ migrations)

Key tables: `tenants`, `users`, `restaurants`, `products`, `orders`, `order_items`, `riders`, `payment_transactions`, `wallet_transactions`, `invoices`, `promos`, `inventory_stocks`, `delivery_zones`, `hubs`

Reference: `docs/requirements/09-database-schema.md`

---
Path scope: All database code, SQLC queries, migrations
