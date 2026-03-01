# Tasks: Complete Partner Portal Integration

## 0. Shared infrastructure (prerequisite for all API tasks)
- [ ] 0.1 Create `partner/src/lib/types/` directory with TypeScript interfaces for API responses: `Order`, `OrderItem`, `Category`, `Product`, `Variant`, `Addon`, `Invoice`, `Payment`, `Promo`, `Rider`, `RiderEarnings`, `RiderAttendance`, `Penalty`, `Banner`, `Section`, `Story`, `TeamMember`, `NotificationPreferences`
- [ ] 0.2 Create TanStack Query hook files: `use-orders.ts`, `use-menu.ts`, `use-finance.ts`, `use-promotions.ts`, `use-riders.ts`, `use-analytics.ts`, `use-content.ts`, `use-team.ts`, `use-settings.ts`
- [ ] 0.3 Add query key constants in `partner/src/lib/query-keys.ts`
- [ ] 0.4 Verify `NEXT_PUBLIC_API_BASE_URL` is set in `.env.local`

## 1. Backend — Team Management Module
- [ ] 1.1 Create `backend/internal/modules/team/` with `handler.go`, `service.go`, `repository.go`
- [ ] 1.2 Add SQLC queries: `ListTenantMembers`, `CreateInvitation`, `GetInvitationByToken`, `AcceptInvitation`, `UpdateMemberRole`, `RemoveMember`
- [ ] 1.3 Register routes: `GET /partner/team`, `POST /partner/team/invite`, `PUT /partner/team/:userId/role`, `DELETE /partner/team/:userId` in `server.go`
- [ ] 1.4 Wire invitation email via existing email adapter
- [ ] 1.5 Add `POST /auth/invite/accept` handler if not present

## 2. Backend — Notification Preferences API
- [ ] 2.1 Add migration for `notification_preferences JSONB` column on `tenant_users` table (if missing)
- [ ] 2.2 Add SQLC queries: `GetNotificationPreferences`, `UpsertNotificationPreferences`
- [ ] 2.3 Register routes in notification or settings module: `GET /partner/settings/notifications`, `PUT /partner/settings/notifications`

## 3. Backend — Promo Stats Endpoint
- [ ] 3.1 Add SQLC query: `GetPromoStats(promoID)` aggregating usage_count, total_discount_given, unique_users
- [ ] 3.2 Register route: `GET /partner/promos/:id/stats` in promo handler

## 4. Backend — Confirm partner restaurant list endpoint
- [ ] 4.1 Verify `GET /partner/restaurants` exists and returns correct shape (name, logo, availability, cuisines, stats)
- [ ] 4.2 If missing or incomplete, add/fix handler in restaurant module

## 5. Dashboard — Real API integration
- [ ] 5.1 Wire `GET /partner/dashboard/summary` into `page.tsx` using TanStack Query (replace KPI card mock values)
- [ ] 5.2 Wire `use-sse.ts` into `incoming-order-panel.tsx` subscribing to tenant+restaurant order channel
- [ ] 5.3 Wire `use-audio-notification.ts` to fire on SSE new-order events
- [ ] 5.4 Wire `GET /partner/analytics/trends?days=7` into `trend-chart.tsx` component
- [ ] 5.5 Implement 3-minute auto-reject countdown timer with auto-call to reject endpoint on expiry

## 6. Order Management — Real API integration
- [ ] 6.1 Replace `useState(mockOrders)` with `useQuery` calling `GET /partner/orders`
- [ ] 6.2 Wire Accept button → `PATCH /partner/orders/:id/confirm` with optimistic update
- [ ] 6.3 Wire Reject button (with reason input) → `PATCH /partner/orders/:id/reject`
- [ ] 6.4 Wire "Mark Preparing" → `PATCH /partner/orders/:id/preparing`
- [ ] 6.5 Wire "Mark Ready" → `PATCH /partner/orders/:id/ready`
- [ ] 6.6 Wire order detail drawer → `GET /partner/orders/:id`
- [ ] 6.7 Replace history `useState(mockOrders)` with paginated query + search/date filters
- [ ] 6.8 Wire SSE events into kanban to append new order cards in real time

## 7. Menu Management — Real API integration
- [ ] 7.1 Replace `useState(mockCategories)` with `useQuery` calling `GET /partner/restaurants/:id/categories`
- [ ] 7.2 Wire add-category form → `POST /partner/restaurants/:id/categories` with `useMutation`
- [ ] 7.3 Wire delete/edit category → `DELETE` / `PUT /partner/categories/:id`
- [ ] 7.4 Wire drag-drop reorder → `PUT /partner/restaurants/:id/categories/reorder`
- [ ] 7.5 Replace `useState(mockProducts)` with `useQuery` calling `GET /partner/restaurants/:id/products?category_id=`
- [ ] 7.6 Wire product create → `POST /partner/restaurants/:id/products` with image upload step
- [ ] 7.7 Wire product edit → `PUT /partner/products/:id` with image upload step
- [ ] 7.8 Wire product availability toggle → `PATCH /partner/products/:id/availability`
- [ ] 7.9 Complete variant builder: add/remove rows with `name` + `price` fields, include in form payload
- [ ] 7.10 Complete addon builder: add/remove addon groups + items, include in form payload
- [ ] 7.11 Implement CSV validation in bulk-upload modal (required headers, price format), wire upload → `POST /partner/restaurants/:id/products/bulk-upload`

## 8. Finance — Real API integration
- [ ] 8.1 Replace finance summary mock with `useQuery` → `GET /partner/finance/summary`
- [ ] 8.2 Replace invoice list mock with `useQuery` → `GET /partner/finance/invoices`
- [ ] 8.3 Replace invoice detail mock with `useQuery` → `GET /partner/finance/invoices/:id`
- [ ] 8.4 Wire "Download PDF" button → `GET /partner/finance/invoices/:id/pdf` with file download
- [ ] 8.5 Replace payment history mock with `useQuery` → `GET /partner/finance/payments`

## 9. Promotions — Real API integration
- [ ] 9.1 Replace `useState(mockPromos)` with `useQuery` → `GET /partner/promos`
- [ ] 9.2 Wire create form → `POST /partner/promos` with `useMutation` + redirect on success
- [ ] 9.3 Wire edit form → `GET /partner/promos/:id` on load + `PUT /partner/promos/:id` on save
- [ ] 9.4 Wire deactivate → `PATCH /partner/promos/:id/deactivate`
- [ ] 9.5 Wire promo stats → `GET /partner/promos/:id/stats`

## 10. Rider Management — Real API integration
- [ ] 10.1 Replace `useState(mockRiders)` with `useQuery` → `GET /partner/riders`
- [ ] 10.2 Wire availability toggle → `PATCH /partner/riders/:id/availability`
- [ ] 10.3 Wire create form → `POST /partner/riders`
- [ ] 10.4 Wire rider detail stats → `GET /partner/riders/:id`
- [ ] 10.5 Wire earnings section → `GET /partner/riders/:id/earnings`
- [ ] 10.6 Wire attendance calendar → `GET /partner/riders/:id/attendance`
- [ ] 10.7 Wire penalties section → `GET /partner/riders/:id/penalties`
- [ ] 10.8 Add debounced search input → refetch with `q=` param

## 11. Analytics — Real API integration
- [ ] 11.1 Wire sales chart → `GET /partner/analytics/sales` with date range + group_by params
- [ ] 11.2 Wire top products table → `GET /partner/analytics/top-products`
- [ ] 11.3 Wire peak hours heatmap → `GET /partner/analytics/peak-hours`
- [ ] 11.4 Wire order breakdown pie chart → `GET /partner/analytics/order-breakdown`
- [ ] 11.5 Wire rider performance table → `GET /partner/analytics/rider-performance`
- [ ] 11.6 Wire CSV export button → `GET /partner/analytics/sales/export` with file download

## 12. Content Management — Real API integration
- [ ] 12.1 Replace `useState(mockBanners)` with `useQuery` → `GET /partner/content/banners`
- [ ] 12.2 Wire add-banner modal: upload image first → `POST /partner/media/upload`, then `POST /partner/content/banners`
- [ ] 12.3 Wire banner delete → `DELETE /partner/content/banners/:id`
- [ ] 12.4 Wire banner toggle → `PUT /partner/content/banners/:id`
- [ ] 12.5 Wire banner drag-drop reorder → `PUT /partner/content/banners/reorder`
- [ ] 12.6 Replace `useState(mockSections)` with `useQuery` → `GET /partner/content/sections`
- [ ] 12.7 Wire add/edit/delete/toggle sections via content API
- [ ] 12.8 Replace `useState(mockStories)` with `useQuery` → `GET /partner/content/stories`
- [ ] 12.9 Wire add-story modal: upload media → `POST /partner/media/upload`, then `POST /partner/content/stories`
- [ ] 12.10 Wire story delete and toggle via content API

## 13. Team Management — Real API integration
- [ ] 13.1 Replace `useState(mockTeam)` with `useQuery` → `GET /partner/team`
- [ ] 13.2 Wire invite modal → `POST /partner/team/invite`
- [ ] 13.3 Wire role selector change → `PUT /partner/team/:userId/role`
- [ ] 13.4 Wire remove button → `DELETE /partner/team/:userId` with confirmation dialog

## 14. Settings — Real API integration
- [ ] 14.1 Wire vendor profile section → load from `GET /partner/me/profile`, save via `PUT /partner/me/profile`
- [ ] 14.2 Wire notification preferences → load from `GET /partner/settings/notifications`, save via `PUT /partner/settings/notifications`

## 15. Restaurant list page — fix mock
- [ ] 15.1 Replace restaurant list page mock with `useQuery` → `GET /partner/restaurants`

## 16. Quality & validation
- [ ] 16.1 Run `npm run build` in `partner/` and fix any TypeScript errors
- [ ] 16.2 Run `go build ./...` in `backend/` to verify new modules compile
- [ ] 16.3 Run `openspec validate complete-partner-portal --strict`
- [ ] 16.4 Manual smoke test: login → dashboard KPIs load → place test order → kanban updates → order detail opens → finance invoices render
