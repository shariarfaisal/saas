# Go Backend Patterns

## Handler Pattern

```go
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    // 1. Extract context (tenant, user)
    t := tenant.FromContext(r.Context())
    u := auth.UserFromContext(r.Context())

    // 2. Parse and validate request body
    var req struct {
        Items []OrderItem `json:"items" validate:"required,min=1"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respond.Error(w, apperror.BadRequest("invalid request body"))
        return
    }

    // 3. Call service with parsed data
    order, err := h.svc.CreateOrder(r.Context(), t.ID, u.ID, req.Items)
    if err != nil {
        respond.Error(w, toAppError(err))
        return
    }

    // 4. Return response
    respond.JSON(w, http.StatusCreated, order)
}
```

**Rules:**
- Handlers ONLY parse HTTP, validate input, call service, return response
- Use inline structs for request parsing
- Always check `tenant.FromContext()` and `auth.UserFromContext()` for nil
- Convert all errors via `toAppError(err)` before responding

## Service Pattern

```go
type Service struct {
    q     *sqlc.Queries
    redis *redisclient.Client
}

func NewService(q *sqlc.Queries, redis *redisclient.Client) *Service {
    return &Service{q: q, redis: redis}
}

func (s *Service) CreateOrder(ctx context.Context, tenantID, userID uuid.UUID,
    items []OrderItem) (*Order, error) {
    // Business logic here
    // Return domain types or errors — never HTTP types
}
```

**Rules:**
- Services receive parsed, validated data — never `*http.Request`
- Return domain types and `error` — never write HTTP responses
- Dependencies injected via constructor

## Error Handling

```go
// Use typed errors from apperror package
apperror.NotFound("restaurant")           // 404
apperror.BadRequest("invalid phone")      // 400
apperror.Unauthorized("invalid token")    // 401
apperror.Forbidden("insufficient role")   // 403
apperror.Internal("create order", err)    // 500 (hides internal details)
apperror.RateLimited()                    // 429
apperror.ValidationError("msg", details)  // 400

// Check pgx.ErrNoRows
if errors.Is(err, pgx.ErrNoRows) {
    return apperror.NotFound("order")
}
```

## Routing (Chi)

```go
r.Route("/api/v1", func(r chi.Router) {
    r.Use(tenantResolver.Middleware)

    r.Route("/auth", func(r chi.Router) {
        r.Post("/otp/send", authHandler.SendOTP)
    })

    r.Route("/orders", func(r chi.Router) {
        r.Use(authMiddleware.Authenticate)
        r.Use(auth.RequireRoles(sqlc.UserRoleOwner, sqlc.UserRoleManager))
        r.Post("", orderHandler.CreateOrder)
        r.Get("/{id}", orderHandler.GetOrder)
    })
})
```

## Dependency Injection

All wiring happens in `server.New()`:
```go
svc := order.NewService(deps.Queries, deps.Redis)
handler := order.NewHandler(svc)
// Then register routes
```

No global state. No init(). No service locators.

## Logging (zerolog)

```go
log.Info().Str("tenant_id", t.ID.String()).Msg("order created")
log.Error().Err(err).Str("module", "order").Msg("failed to create order")
```

Always include context: tenant_id, user_id, request_id, module name.

## Response Helpers

```go
respond.JSON(w, http.StatusOK, data)           // Success
respond.Error(w, apperror.BadRequest("msg"))   // Error
// Paginated
respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: items, Meta: meta})
```

## Testing

```bash
make test              # Unit tests with race detection
make test-integration  # Integration tests (requires DB)
```

Tests use standard `testing` package. Test files: `*_test.go` alongside source.

---
Path scope: backend/
