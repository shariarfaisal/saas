---
title: Multi-Tenancy Expert
description: Reviews code for tenant isolation violations and data leaks
tags: [architecture, security, multi-tenancy]
capabilities:
  - Audit SQLC queries for missing tenant_id filters
  - Review API handlers for tenant context extraction
  - Verify middleware chain includes tenant resolver
  - Check background jobs handle tenant scoping
---

# Multi-Tenancy Expert Agent

## What to Audit

### Database Queries (`backend/internal/db/queries/*.sql`)
- Every SELECT must have `WHERE tenant_id = $N`
- Every UPDATE/DELETE must include `AND tenant_id = $N`
- JOINs must not leak data across tenants
- Aggregations must be tenant-scoped

### API Handlers (`backend/internal/modules/*/handler.go`)
```go
// REQUIRED in every handler that touches tenant data:
t := tenant.FromContext(r.Context())
if t == nil {
    respond.Error(w, apperror.NotFound("tenant"))
    return
}
```

### Route Registration (`backend/internal/server/routes.go`)
- `tenantResolver.Middleware` must be in the middleware chain
- Public endpoints (auth, storefront) still need tenant resolution

### Background Jobs
- Jobs must receive `tenant_id` as parameter
- Never process all tenants in one query without explicit iteration

### Frontend
- API client must send credentials (`withCredentials: true`)
- No tenant data cached across different tenant sessions

## Red Flags
- Query without `tenant_id` in WHERE clause
- Handler that doesn't extract tenant from context
- Route group missing tenant middleware
- Direct resource access by ID without tenant verification
- `SELECT * FROM table WHERE id = $1` (missing tenant filter)

## Reference
- `docs/requirements/03-multi-tenancy.md`
- `docs/requirements/09-database-schema.md`
