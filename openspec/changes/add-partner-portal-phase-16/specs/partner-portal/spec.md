## ADDED Requirements

### Requirement: Partner Portal Project Setup
The system SHALL provide a partner portal as a Next.js 14+ application with TypeScript, Tailwind CSS, TanStack Query, Zustand, React Hook Form with Zod validation, and Axios API client with tenant-aware base URL configuration.

#### Scenario: Partner portal app initialisation
- **WHEN** a developer clones the repository and runs `npm install && npm run dev` in the `partner/` directory
- **THEN** the Next.js development server starts without errors and serves the partner portal at localhost

#### Scenario: API client sends tenant-scoped requests
- **WHEN** any API call is made from the partner portal
- **THEN** the Axios client includes `X-Request-ID` header and sends credentials (cookies) with every request
- **THEN** on 401 response, the client attempts token refresh before failing

### Requirement: Partner Authentication
The system SHALL provide email+password login, forgot/reset password flow, invitation acceptance with token-based password setup, and multi-restaurant picker for users with access to multiple restaurants.

#### Scenario: Partner login with single restaurant
- **WHEN** a partner user submits valid email and password and has access to one restaurant
- **THEN** an access token cookie is set and the user is redirected to the dashboard

#### Scenario: Partner login with multiple restaurants
- **WHEN** a partner user submits valid credentials and has access to multiple restaurants
- **THEN** a restaurant picker is shown allowing the user to select which restaurant to manage

#### Scenario: Invitation acceptance
- **WHEN** a new team member clicks an invitation link with a valid token
- **THEN** they are shown a page to set their password and upon completion are logged in

#### Scenario: Password reset
- **WHEN** a user clicks "Forgot Password" and submits their email
- **THEN** a password reset link is sent and the user can set a new password via the reset token

### Requirement: Partner Dashboard with Live Orders
The system SHALL display a dashboard with KPI cards, a live incoming order panel connected via SSE, audio notifications for new orders, 7-day trend charts, and quick-action buttons.

#### Scenario: Dashboard KPI display
- **WHEN** a partner user visits the dashboard
- **THEN** they see KPI cards showing today's orders, revenue, pending count, and average delivery time

#### Scenario: Live order notification
- **WHEN** a new order arrives via SSE
- **THEN** the incoming order panel updates with the new order, an audio notification plays, and a visual badge appears

#### Scenario: Order accept/reject with countdown
- **WHEN** a new order appears in the incoming panel
- **THEN** accept and reject buttons are shown with a 3-minute countdown timer

### Requirement: Restaurant Management
The system SHALL allow partners to list, create, edit, and toggle availability of restaurants, with an operating hours day-by-day scheduler and a branch switcher in the sidebar.

#### Scenario: Restaurant list with availability toggle
- **WHEN** a partner visits the restaurants page
- **THEN** restaurants are shown as cards with an availability toggle that calls the PATCH availability API

#### Scenario: Restaurant create/edit form
- **WHEN** a partner creates or edits a restaurant
- **THEN** a form is shown with fields for name, description, cuisines, address, images, operating hours per day, VAT rate, and prep time

#### Scenario: Branch switcher
- **WHEN** a partner has multiple restaurants
- **THEN** a branch switcher in the sidebar allows selecting which restaurant scopes all views

### Requirement: Menu Management
The system SHALL provide menu management with drag-drop category reordering, product grid per category, product create/edit with variant and addon builders, availability toggles, and CSV bulk upload.

#### Scenario: Category drag-drop reorder
- **WHEN** a partner drags a category in the left panel
- **THEN** the category order is updated and persisted via API

#### Scenario: Product create/edit sheet
- **WHEN** a partner clicks to create or edit a product
- **THEN** a slide-in sheet appears with fields for name, images, price type, variants, addons, and discount toggle

#### Scenario: Bulk CSV upload
- **WHEN** a partner uploads a CSV file in the bulk upload modal
- **THEN** the system validates and creates products from the CSV data

### Requirement: Live Order Management Board
The system SHALL display orders as a kanban board with columns (New, Confirmed, Preparing, Ready, Picked), order detail drawer, and order history with search and filters.

#### Scenario: Kanban board display
- **WHEN** a partner visits the orders page
- **THEN** active orders are shown in kanban columns by status with order number, items summary, time elapsed, and customer area

#### Scenario: Order status action
- **WHEN** a partner clicks an action button on an order card (e.g., Confirm, Start Preparing, Mark Ready)
- **THEN** the order status is updated via API and the card moves to the next column

#### Scenario: Order detail drawer
- **WHEN** a partner clicks on an order card
- **THEN** a drawer opens showing items with addons, customer info, rider info, address, payment details, and order timeline

### Requirement: Rider Management
The system SHALL provide rider list, create/edit forms, detail pages with stats and attendance, and availability toggles.

#### Scenario: Rider list display
- **WHEN** a partner visits the riders page
- **THEN** riders are shown in a table with name, hub, status badge, location, and today's order count

#### Scenario: Rider detail page
- **WHEN** a partner clicks on a rider
- **THEN** a detail page shows stats, attendance calendar, earnings, and penalties

### Requirement: Promotions Management
The system SHALL provide promo list, create/edit forms with all restriction fields, and performance statistics.

#### Scenario: Promo create with restrictions
- **WHEN** a partner creates a promotion
- **THEN** a form allows setting code, type, amount, cap, apply_on, restrictions, date range, per-user limit, and cashback amount

#### Scenario: Promo performance stats
- **WHEN** a partner views a promotion's details
- **THEN** usage count, total discount given, and unique users are displayed

### Requirement: Finance Pages
The system SHALL provide finance summary, invoice list with status badges, invoice detail with full breakdown, PDF download, payment history, and outstanding balance alerts.

#### Scenario: Finance summary display
- **WHEN** a partner visits the finance page
- **THEN** current period net payable and YTD totals are displayed

#### Scenario: Invoice detail and PDF
- **WHEN** a partner clicks on an invoice
- **THEN** a detail page shows full breakdown (gross sales, commission, promos, penalties, adjustments, net payable)
- **THEN** a PDF download button is available

### Requirement: Sales and Analytics
The system SHALL provide sales reports with date range picker, grouping options, CSV export, top-selling products, peak hours heatmap, order breakdown chart, rider performance, and customer area distribution.

#### Scenario: Sales report with export
- **WHEN** a partner selects a date range and grouping option
- **THEN** sales data is displayed with all metrics and a CSV export button is available

#### Scenario: Peak hours heatmap
- **WHEN** a partner views the reports section
- **THEN** a heatmap shows order volume by hour and day of week

### Requirement: Content Management
The system SHALL provide banners, stories, and homepage sections management with image upload, targeting, and drag-drop ordering.

#### Scenario: Banner management
- **WHEN** a partner creates a banner
- **THEN** they can upload an image, set link type/value, area targeting, validity dates, and sort order

#### Scenario: Stories management
- **WHEN** a partner creates a story
- **THEN** they can upload media, set expiry, and link to a restaurant

### Requirement: Partner Settings and Team Management
The system SHALL provide vendor profile settings, notification preferences, team member management with role-based invitations, and an API keys placeholder.

#### Scenario: Team invite
- **WHEN** a partner invites a team member via email
- **THEN** an invitation is sent with a role assignment, and the member appears in the team list as pending

#### Scenario: Notification preferences
- **WHEN** a partner configures notification preferences
- **THEN** they can toggle notifications per event type (new orders, status changes, invoices, etc.)
