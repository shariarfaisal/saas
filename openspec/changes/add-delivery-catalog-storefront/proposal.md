# Change: Add Delivery Infrastructure, Restaurant Catalog, and Storefront API

## Why
Phase 2 and Phase 3 of the SaaS platform require hub/delivery-zone management, full restaurant catalog CRUD, and a public storefront API so customers can browse restaurants and menus.

## What Changes
- Add hub + hub coverage area + delivery zone config CRUD (partner API)
- Add `DeliveryChargeService.Calculate` + pre-calc endpoint
- Add restaurant + operating hours CRUD (partner API)
- Add category CRUD + reorder (partner API)
- Add product CRUD + availability + image upload stub (partner API)
- Add product modifier groups and options (variants/addons) (partner API)
- Add product discounts with expiry (partner API)
- Add public storefront endpoints (no auth)
- Add CSV bulk menu upload (partner API)
- Add menu duplication between restaurants (partner API)

## Impact
- Affected specs: hub-delivery, restaurant-catalog, storefront
- Affected code: `internal/db/queries/`, `internal/db/sqlc/`, `internal/modules/`, `internal/server/server.go`
