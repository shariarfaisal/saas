## 1. Project setup (TASK-080)
- [ ] 1.1 Scaffold `partner/` Next.js project with TypeScript + Tailwind + ESLint + Prettier + App Router
- [ ] 1.2 Add dependencies (TanStack Query, Zustand, React Hook Form, Zod, Axios, Recharts, Lucide React, shadcn/ui utility deps)
- [ ] 1.3 Implement API client with tenant-aware base URL, interceptors, token refresh, `X-Request-ID`
- [ ] 1.4 Implement Zustand auth store with restaurant context
- [ ] 1.5 Implement protected route layout and Next.js middleware
- [ ] 1.6 Implement partner shell layout with sidebar navigation
- [ ] 1.7 Add notification bell (polling unread count) and new-order audio notification hook

## 2. Auth (TASK-081)
- [ ] 2.1 Build login page (email + password) with React Hook Form + Zod validation
- [ ] 2.2 Build forgot/reset password flow pages
- [ ] 2.3 Build invitation acceptance page (set password from invite token)
- [ ] 2.4 Build multi-restaurant picker (shown after login when user has multiple restaurants)
- [ ] 2.5 Implement API route handlers for login, logout, refresh, password-reset

## 3. Dashboard (TASK-082)
- [ ] 3.1 Build KPI cards (today's orders, revenue, pending count, avg delivery time)
- [ ] 3.2 Build live incoming order panel with SSE connection and audio notification
- [ ] 3.3 Build accept/reject buttons with 3-minute countdown timer
- [ ] 3.4 Build 7-day trend charts (orders and revenue)
- [ ] 3.5 Build quick-action buttons (toggle restaurant availability, view pending issues)

## 4. Restaurant management (TASK-083)
- [ ] 4.1 Build restaurant list page with cards and availability toggle
- [ ] 4.2 Build restaurant create/edit form (all fields including operating hours scheduler)
- [ ] 4.3 Implement branch switcher in sidebar for multi-restaurant scoping

## 5. Menu management (TASK-084)
- [ ] 5.1 Build category list panel with drag-drop reorder
- [ ] 5.2 Build product grid per category with availability toggles
- [ ] 5.3 Build product create/edit sheet (variant builder, addon builder, discount toggle)
- [ ] 5.4 Build bulk-upload CSV modal

## 6. Order management (TASK-085)
- [ ] 6.1 Build kanban board with columns (New → Confirmed → Preparing → Ready → Picked)
- [ ] 6.2 Build order cards with info and action buttons per status
- [ ] 6.3 Build order detail drawer (items, customer, rider, payment, timeline)
- [ ] 6.4 Build order history table with search and filters

## 7. Rider management (TASK-086)
- [ ] 7.1 Build rider list table with status badges and today's orders
- [ ] 7.2 Build rider create/edit form
- [ ] 7.3 Build rider detail page (stats, attendance, earnings, penalties)
- [ ] 7.4 Build attendance calendar and availability toggle

## 8. Promotions (TASK-087)
- [ ] 8.1 Build promo list table (code, type, usage count, status)
- [ ] 8.2 Build promo create/edit form (all fields including restrictions and cashback)
- [ ] 8.3 Build promo performance stats (usage, total discount, unique users)

## 9. Finance (TASK-088)
- [ ] 9.1 Build finance summary (current period net payable, YTD totals)
- [ ] 9.2 Build invoice list with status badges
- [ ] 9.3 Build invoice detail page with full breakdown
- [ ] 9.4 Implement PDF download action
- [ ] 9.5 Build payment history and outstanding balance alert

## 10. Sales & analytics (TASK-089)
- [ ] 10.1 Build sales report with date range picker and CSV export
- [ ] 10.2 Build top-selling products table
- [ ] 10.3 Build peak hours heatmap
- [ ] 10.4 Build order status breakdown chart
- [ ] 10.5 Build rider performance table and customer area distribution

## 11. Content management (TASK-090)
- [ ] 11.1 Build banners page (image upload, link config, area targeting, dates, drag-drop sort)
- [ ] 11.2 Build stories page (media upload, expiry, restaurant link)
- [ ] 11.3 Build homepage sections editor (type, display order, restaurant/product selection)

## 12. Settings & team management (TASK-091)
- [ ] 12.1 Build settings page (vendor profile, notification preferences)
- [ ] 12.2 Build team management (list, invite modal, role selector, remove member)
- [ ] 12.3 Add API keys placeholder page

## 13. Quality
- [ ] 13.1 Run partner lint/build checks
- [ ] 13.2 Capture UI screenshots for key pages
- [ ] 13.3 Update TASKS.md Phase 16 checkboxes
