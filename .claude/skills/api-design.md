---
title: API Endpoint Implementation
description: Step-by-step guide for adding new API endpoints to the Munchies backend
tags: [api, backend, go]
---

# Implementing a New API Endpoint

## Step-by-step Workflow

### 1. Write SQLC Queries

Create or update `backend/internal/db/queries/{entity}.sql`:

```sql
-- name: CreatePromotion :one
INSERT INTO promotions (
    id, tenant_id, code, discount_type, discount_value,
    min_order_amount, max_discount, usage_limit, starts_at, ends_at
) VALUES (
    gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetPromotionByCode :one
SELECT * FROM promotions
WHERE code = $1 AND tenant_id = $2 AND deleted_at IS NULL;
```

Then run: `cd backend && make sqlc`

### 2. Create Module Structure

```
backend/internal/modules/{name}/
├── handler.go
├── service.go
└── context.go (if needed)
```

### 3. Implement Service

```go
package promo

type Service struct {
    q *sqlc.Queries
}

func NewService(q *sqlc.Queries) *Service {
    return &Service{q: q}
}

func (s *Service) CreatePromotion(ctx context.Context, tenantID uuid.UUID,
    req CreatePromoInput) (*sqlc.Promotion, error) {
    // Validate business rules
    if req.StartsAt.After(req.EndsAt) {
        return nil, apperror.BadRequest("start date must be before end date")
    }
    promo, err := s.q.CreatePromotion(ctx, sqlc.CreatePromotionParams{
        TenantID: uuidToPgtype(&tenantID),
        Code:     req.Code,
    })
    if err != nil {
        return nil, apperror.Internal("create promotion", err)
    }
    return &promo, nil
}
```

### 4. Implement Handler

```go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    t := tenant.FromContext(r.Context())
    if t == nil {
        respond.Error(w, apperror.NotFound("tenant"))
        return
    }
    var req CreatePromoInput
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respond.Error(w, apperror.BadRequest("invalid request body"))
        return
    }
    promo, err := h.svc.CreatePromotion(r.Context(), t.ID, req)
    if err != nil {
        respond.Error(w, toAppError(err))
        return
    }
    respond.JSON(w, http.StatusCreated, promo)
}
```

### 5. Register Routes in `server/routes.go`

```go
promoSvc := promo.NewService(deps.Queries)
promoHandler := promo.NewHandler(promoSvc)

r.Route("/promotions", func(r chi.Router) {
    r.Use(authMiddleware.Authenticate)
    r.Use(auth.RequireRoles(sqlc.UserRoleOwner, sqlc.UserRoleManager))
    r.Post("", promoHandler.Create)
    r.Get("", promoHandler.List)
})
```

### 6. Test and Verify

```bash
make sqlc     # Regenerate types
make build    # Verify compilation
make test     # Run tests
```
