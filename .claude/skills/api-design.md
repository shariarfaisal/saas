---
title: API Design Skill
description: Reusable knowledge for designing REST APIs following project standards
tags: [api, design, rest]
---

# API Design Skill

Reusable patterns and standards for API design in the Munchies SaaS platform.

## Key References

- **docs/requirements/10-api-design.md** - Complete API design specification
- **docs/requirements/09-database-schema.md** - Database structure reference
- **docs/requirements/04-domain-model.md** - Domain concepts and entities

## API Standards

- RESTful endpoints
- Consistent naming conventions
- Pagination for list endpoints
- Proper HTTP status codes
- Error response format
- Tenant isolation in queries

## Endpoints Pattern

```
/api/v1/:tenantId/resource
/api/v1/:tenantId/resource/:id
```

Always include tenant context in routes.
