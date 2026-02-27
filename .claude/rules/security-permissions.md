# Security & Permissions Rule

All endpoints and features must verify user permissions and tenant context.

## Security Checklist

1. **Tenant authorization** - Verify user belongs to requested tenant
2. **Role-based access** - Check user role for the action (see doc 06)
3. **Resource ownership** - Verify resource belongs to user's tenant
4. **Data validation** - Validate all user inputs
5. **Rate limiting** - Apply rate limits to prevent abuse

## Permission Model

Reference: `docs/requirements/06-portals-and-roles.md`

User roles:

- Restaurant Owner/Admin
- Manager
- Staff
- Customer
- Platform Admin

Each role has specific permissions - verify these before implementing.

## Common Mistakes to Avoid

- Forgetting tenant check
- Only checking role, not tenant combination
- Allowing direct resource access by ID without verification
- Missing validation on input

---

Path scope: All API endpoints, user-facing features
