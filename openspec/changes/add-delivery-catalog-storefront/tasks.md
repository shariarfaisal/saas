## 1. Database & SQLC
- [x] 1.1 Create `hubs.sql` SQLC query file
- [x] 1.2 Create `restaurants.sql` SQLC query file
- [x] 1.3 Create `categories.sql` SQLC query file
- [x] 1.4 Create `products.sql` SQLC query file (includes modifiers and discounts)
- [x] 1.5 Write `hubs.sql.go` (manually generated SQLC implementation)
- [x] 1.6 Write `restaurants.sql.go`
- [x] 1.7 Write `categories.sql.go`
- [x] 1.8 Write `products.sql.go`
- [x] 1.9 Update `querier.go` with all new method signatures

## 2. Modules
- [x] 2.1 Implement `internal/modules/hub/` (handler, service, repository)
- [x] 2.2 Implement `internal/modules/restaurant/` (handler, service, repository)
- [x] 2.3 Implement `internal/modules/catalog/` (handler, service, repository)
- [x] 2.4 Implement `internal/modules/delivery/` (service, handler)
- [x] 2.5 Implement `internal/modules/storefront/` (handler, service)
- [x] 2.6 Implement `internal/modules/media/` (handler stub)

## 3. Routes
- [x] 3.1 Register all partner routes in `server.go`
- [x] 3.2 Register all public storefront routes in `server.go`

## 4. Quality
- [x] 4.1 Build passes (`go build ./...`)
- [x] 4.2 Tests pass (`go test ./...`)
