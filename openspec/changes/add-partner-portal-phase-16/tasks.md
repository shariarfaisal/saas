## 1. Project setup (TASK-080)
- [x] 1.1 Scaffold `partner/` Next.js project with TypeScript + Tailwind + ESLint + Prettier + App Router
- [x] 1.2 Add dependencies (TanStack Query, Zustand, React Hook Form, Zod, Axios, Recharts, Lucide React, shadcn/ui utility deps)
- [x] 1.3 Implement API client with tenant-aware base URL, interceptors, token refresh, `X-Request-ID`
- [x] 1.4 Implement Zustand auth store with restaurant context
- [x] 1.5 Implement protected route layout and Next.js middleware
- [x] 1.6 Implement partner shell layout with sidebar navigation
- [x] 1.7 Add notification bell (polling unread count) and new-order audio notification hook

## 2. Auth (TASK-081)
- [x] 2.1 Build login page (email + password) with React Hook Form + Zod validation
- [x] 2.2 Build forgot/reset password flow pages
- [x] 2.3 Build invitation acceptance page (set password from invite token)
- [x] 2.4 Build multi-restaurant picker (shown after login when user has multiple restaurants)
- [x] 2.5 Implement API route handlers for login, logout, refresh, password-reset

## 3. Dashboard (TASK-082)
- [x] 3.1 Build KPI cards (today's orders, revenue, pending count, avg delivery time)
- [x] 3.2 Build live incoming order panel with SSE connection and audio notification
- [x] 3.3 Build accept/reject buttons with 3-minute countdown timer
- [x] 3.4 Build 7-day trend charts (orders and revenue)
- [x] 3.5 Build quick-action buttons (toggle restaurant availability, view pending issues)

## 4. Restaurant management (TASK-083)
- [x] 4.1 Build restaurant list page with cards and availability toggle
- [x] 4.2 Build restaurant create/edit form (all fields including operating hours scheduler)
- [x] 4.3 Implement branch switcher in sidebar for multi-restaurant scoping

## 5. Menu management (TASK-084)
- [x] 5.1 Build category list panel with drag-drop reorder
- [x] 5.2 Build product grid per category with availability toggles
- [x] 5.3 Build product create/edit sheet (variant builder, addon builder, discount toggle)
- [x] 5.4 Build bulk-upload CSV modal

## 6. Order management (TASK-085)
- [x] 6.1 Build kanban board with columns (New → Confirmed → Preparing → Ready → Picked)
- [x] 6.2 Build order cards with info and action buttons per status
- [x] 6.3 Build order detail drawer (items, customer, rider, payment, timeline)
- [x] 6.4 Build order history table with search and filters

## 7. Rider management (TASK-086)
- [x] 7.1 Build rider list table with status badges and today's orders
- [x] 7.2 Build rider create/edit form
- [x] 7.3 Build rider detail page (stats, attendance, earnings, penalties)
- [x] 7.4 Build attendance calendar and availability toggle

## 8. Promotions (TASK-087)
- [x] 8.1 Build promo list table (code, type, usage count, status)
- [x] 8.2 Build promo create/edit form (all fields including restrictions and cashback)
- [x] 8.3 Build promo performance stats (usage, total discount, unique users)

## 9. Finance (TASK-088)
- [x] 9.1 Build finance summary (current period net payable, YTD totals)
- [x] 9.2 Build invoice list with status badges
- [x] 9.3 Build invoice detail page with full breakdown
- [x] 9.4 Implement PDF download action
- [x] 9.5 Build payment history and outstanding balance alert

## 10. Sales & analytics (TASK-089)
- [x] 10.1 Build sales report with date range picker and CSV export
- [x] 10.2 Build top-selling products table
- [x] 10.3 Build peak hours heatmap
- [x] 10.4 Build order status breakdown chart
- [x] 10.5 Build rider performance table and customer area distribution

## 11. Content management (TASK-090)
- [x] 11.1 Build banners page (image upload, link config, area targeting, dates, drag-drop sort)
- [x] 11.2 Build stories page (media upload, expiry, restaurant link)
- [x] 11.3 Build homepage sections editor (type, display order, restaurant/product selection)

## 12. Settings & team management (TASK-091)
- [x] 12.1 Build settings page (vendor profile, notification preferences)
- [x] 12.2 Build team management (list, invite modal, role selector, remove member)
- [x] 12.3 Add API keys placeholder page

## 13. Quality
- [x] 13.1 Run partner lint/build checks
- [x] 13.2 Capture UI screenshots for key pages
- [x] 13.3 Update TASKS.md Phase 16 checkboxes
