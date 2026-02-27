# Database & Queries Rule

All database operations must respect multi-tenancy and security requirements.

## Query Rules

1. **Always filter by tenant_id** - Every SELECT must include `WHERE tenant_id = ?`
2. **Prevent data leaks** - Test queries across different tenant contexts
3. **Use parameterized queries** - Prevent SQL injection
4. **Document migrations** - Schema changes need clear documentation

## Review Checklist

When reviewing database code:

- [ ] Query includes tenant filter
- [ ] No hardcoded values
- [ ] Parameterized/prepared statements used
- [ ] Index strategy documented
- [ ] Migration is reversible
- [ ] Test covers multiple tenants

Reference: `docs/requirements/09-database-schema.md`

---

Path scope: All database code
