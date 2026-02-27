# 05 — Feature Requirements

## 5.1 Customer Website Features

### 5.1.1 Homepage & Discovery
- **Hero banners**: Rotating promotional banners (image + CTA link)
- **Stories**: Short-lived visual promotions (image/video strips at top)
- **Curated sections**: "Trending Now", "New Arrivals", "Best Sellers", etc.
- **Cuisine/category filters**: Filter restaurants by food type
- **Search**: Search restaurants and menu items (with autocomplete)
- **Restaurant listing**: Cards with name, image, rating, delivery time, min order, offer tags
- **Area/zone detection**: Auto-detect customer area via GPS or manual area selection; only show restaurants that deliver to that area
- **Operating hours**: Real-time open/closed badge on restaurant cards
- **Sorting**: Sort by popularity, rating, delivery time, newest

### 5.1.2 Restaurant Page
- **Restaurant info**: Name, banner, logo, description, cuisines, operating hours, address, phone
- **Category nav**: Sticky horizontal scroll category tabs (clicking jumps to section)
- **Menu items**: Grid or list view of products with image, name, price, availability, discount badge
- **Product detail**: Modal/page with description, images, variant selection, addon selection, quantity selector, total price
- **Search within restaurant**: Filter products within a restaurant's menu
- **Dietary tags**: Vegetarian, vegan, spicy indicators (future)

### 5.1.3 Cart & Checkout
- **Persistent cart**: Cart state preserved in localStorage + synced to session
- **Multi-restaurant cart**: Customer can order from multiple restaurants in one order (they share the same delivery route — hub model)
- **Cart summary**: Items, quantities, item discounts, promo discount, delivery charge, VAT, total
- **Delivery address**: Select from saved addresses or enter new
- **Promo code**: Input field to apply promo code — validated live via API
- **Payment method selection**: COD, bKash, card (AamarPay/SSLCommerz)
- **Order note**: Optional text note to restaurant / rider
- **Estimated delivery time**: Shown at checkout
- **Place order**: Atomic order creation with stock check and charge calculation server-side

### 5.1.4 Order Tracking
- **Active order page**: Real-time status updates via SSE
- **Status timeline**: Visual step-by-step progress (pending → confirmed → preparing → ready → picked → delivered)
- **Rider tracking map**: Live map showing rider position when order is picked (WebSocket / polling)
- **Order details**: All items, charges, restaurant info, rider info
- **Cancel order**: Customer can cancel while status is `pending` or `created` (before restaurant confirms)
- **Call rider**: Display rider phone number when order is picked
- **Order history**: Full list of past orders with status, date, total, quick reorder

### 5.1.5 User Account
- **Register**: Phone OTP, then name + optional email
- **Login**: Phone OTP
- **Profile**: Edit name, email, avatar, date of birth, gender
- **Saved addresses**: CRUD on delivery addresses, set default
- **Saved payment methods**: Save bKash wallet (tokenised agreement)
- **Order history**: Filter by status, date range
- **Loyalty points / wallet balance**: View balance, transaction history
- **Favourite restaurants**: Heart icon to save favourites, dedicated favourites list
- **Reviews**: Submit star rating + comment after delivery

### 5.1.6 Promotions for Customers
- **Promo code input**: On checkout, validated in real-time
- **Automatic offers**: Promos auto-applied if eligible (e.g., first order free delivery)
- **Cashback**: On order completion, eligible cashback credited to wallet
- **Free delivery promos**: Delivery charge zeroed when promo applied on `delivery_charge`

---

## 5.2 Partner Portal Features (Vendor / Restaurant Manager)

### 5.2.1 Dashboard
- **KPIs**: Today's orders, revenue, pending orders, avg delivery time
- **Live order board**: Real-time incoming orders with accept/reject actions
- **Last 7 / 30 days trend**: Orders and revenue graph
- **Notifications**: Invoice ready, penalty applied, low stock alerts
- **Quick actions**: Toggle restaurant availability, view pending issues

### 5.2.2 Restaurant Management
- **Create / edit restaurant**: All fields (name, description, address, cuisines, images, operating hours, VAT, commission view, prep time)
- **Toggle availability**: Instantly open/close restaurant
- **Operating hours**: Day-by-day schedule with open/close times
- **Multiple restaurants**: Vendor can manage all restaurants from one account
- **Branch selector**: Switch context between restaurants quickly

### 5.2.3 Menu Management
- **Categories**: Create, edit, reorder, delete categories; toggle availability
- **Products (menu items)**: Create, edit, delete, toggle availability
  - Basic info: name, description, images
  - Pricing: flat price or variant-based
  - Variants: e.g., Size (Small/Medium/Large) with individual prices
  - Add-ons: e.g., Extra toppings with price
  - Category assignment
  - Sort order drag & drop
- **Product discount**: Set fixed or percent discount with expiry date
- **Bulk product upload**: CSV import for mass menu creation
- **Menu duplication**: Copy full menu from one restaurant to another (within same tenant)
- **Availability management**: Mark product as unavailable or out of stock

### 5.2.4 Order Management
- **Incoming orders**: Real-time notification (sound + visual) for new orders
- **Order detail**: Full item list, customer info, delivery address, payment method
- **Accept order**: Restaurant confirms order, starts prep timer
- **Reject order**: With mandatory reason
- **Mark as Ready**: Signal to rider that food is ready for pickup
- **Order history**: Filterable by date, status, restaurant
- **Search orders**: By order number, customer phone
- **Order issues**: View and respond to disputes / refund requests

### 5.2.5 Rider Management (if tenant uses own riders)
- **Rider list**: All riders, status (active/inactive, on duty/off duty), current location
- **Add rider**: Create rider account, assign to hub
- **Assign rider manually**: Assign rider to a specific order
- **Auto-assign settings**: Configure auto-assignment radius, algorithm
- **Rider availability toggle**: Enable/disable per rider
- **Rider travel log**: View rider GPS history per day
- **Rider attendance**: Daily check-in/check-out records
- **Rider earnings**: View daily/monthly earnings per rider
- **Rider penalties**: Create, view, manage penalty records
- **Registration approvals**: Review rider registration applications

### 5.2.6 Sales & Reports
- **Sales report**: Revenue, orders, avg order value by date range
- **Top selling items**: Per restaurant, per date range
- **Order status breakdown**: Delivered vs rejected vs cancelled
- **Peak hours chart**: When most orders arrive
- **Restaurant-level report**: Individual metrics per branch
- **Revenue by area**: Where customers are ordering from
- **CSV export**: Export reports as CSV

### 5.2.7 Promotions
- **Create promo code**: Full configuration (code, type, amount, limits, restrictions, validity)
- **List / manage promos**: Edit, deactivate, view usage stats
- **Promo performance**: How many times used, total discount given

### 5.2.8 Inventory (for stores / inventory-tracked products)
- **Stock management**: Current stock per product
- **Stock adjustment**: Add/remove stock with reason
- **Low stock alerts**: Notify when below reorder level
- **Purchase records**: Log stock purchase events

### 5.2.9 Finance
- **Commission summary**: View platform commission per period
- **Invoices**: View all generated invoices with breakdown
- **Invoice download**: PDF download of each invoice
- **Payment history**: When invoices were settled
- **Payable balance**: Net amount owed to/from platform

### 5.2.10 Content Management
- **Banners**: Upload and manage homepage banners
- **Stories**: Upload story content with expiry
- **Sections**: Curate homepage sections (which restaurants/products to feature)

### 5.2.11 Settings
- **Profile**: Edit vendor profile
- **Notification preferences**: Which events to receive push/email notifications
- **Team management**: Invite team members, assign roles (restaurant_manager, staff)
- **API keys**: (future) Generate API keys for POS integrations

---

## 5.3 Super Admin Panel Features

### 5.3.1 Platform Dashboard
- **Global KPIs**: Total orders today (all tenants), total revenue, active tenants, active riders
- **Revenue breakdown**: Commission earned by platform across all tenants
- **Order map**: Geographic heat map of orders
- **System health**: API latency, error rates, queue depths

### 5.3.2 Tenant Management
- **List all tenants**: With status, plan, order count, revenue
- **Create tenant**: Full onboarding form
- **Edit tenant**: Plan, commission rate, status, features
- **Suspend / reinstate tenant**: With reason and notification
- **Impersonate tenant**: View partner portal as any tenant
- **Tenant analytics**: Deep dive metrics per tenant
- **Commission override**: Per-restaurant commission override

### 5.3.3 User Management (Platform-Wide)
- **Search users across tenants**: By phone, email, name
- **View user profile**: Order history, account status
- **Suspend user**: Block customer or staff account
- **Delete user**: GDPR-compliant data removal

### 5.3.4 Order Management (Platform-Wide)
- **All orders view**: Across all tenants with tenant filter
- **Order detail**: Full timeline, audit log
- **Force status change**: Override any order status (master key operations)
- **Order issue resolution**: Approve/reject refunds and penalties

### 5.3.5 Financial Management
- **Commission ledger**: Platform earnings from all tenants
- **Invoice management**: Generate, finalize, mark paid for any tenant
- **Payout tracking**: Settlement status with all vendors
- **Sales report**: Platform-wide and per-tenant revenue reports

### 5.3.6 Content Management (Platform-Wide)
- **Global banners**: Shown across all storefronts (optional)
- **Platform announcements**: Broadcast messages to all tenants

### 5.3.7 Platform Configuration
- **Delivery zone presets**: Default delivery zone templates for new tenants
- **Payment gateway config**: Configure gateway credentials per tenant
- **SMS / Email provider config**: Twilio, SendGrid, etc.
- **Feature flags**: Enable/disable platform features globally
- **Maintenance mode**: Suspend all ordering activity

### 5.3.8 Analytics & BI
- **Cross-tenant analytics**: Revenue, orders, growth across all tenants
- **Retention metrics**: Customer retention per tenant
- **Geographic analytics**: Order density maps
- **Rider performance**: Fleet metrics across all tenants
- **Export reports**: Raw CSV data exports

---

## 5.4 Rider Portal / App Features

The rider experience is a **Progressive Web App (PWA)** — mobile-first, installable on phone.

### 5.4.1 Authentication
- Phone OTP login
- Push notifications opt-in

### 5.4.2 Duty Management
- **Attendance**: Check-in to start shift, check-out to end
- **Hub selection**: Rider selects their hub at check-in
- **Availability toggle**: Go online/offline without checking out

### 5.4.3 Order Delivery
- **New order notification**: Push notification + in-app alert
- **Order detail**: Pick-up location(s), items summary, customer address
- **Accept / decline order request** (if manual request system)
- **Navigate to restaurant**: Google Maps deep link to restaurant
- **Mark as Picked**: Tap when food collected from all restaurants
- **Navigate to customer**: Maps link to delivery address
- **Mark as Delivered**: Tap on successful delivery
- **Report issue**: Flag problem with order (missing item, customer unavailable, etc.)

### 5.4.4 Earnings & History
- **Today's summary**: Orders completed, distance, earnings
- **Order history**: Past deliveries with amounts
- **Earnings balance**: Total accumulated earnings

### 5.4.5 Profile
- **Edit profile**: Name, avatar, vehicle type
- **View penalties**: Any active penalties with details and appeal option

---

## 5.5 Cross-Cutting Features

### Notifications
- **Push notifications** (Firebase FCM): New order, order status change, promo
- **SMS**: OTP, order confirmation
- **Email**: Invoice, registration confirmation, password reset
- **In-app**: Partner portal notification center, customer order updates

### Search
- **Full-text search**: Restaurant names, menu items, cuisines
- **Autocomplete**: Suggestions as user types
- **Search logs**: Track popular searches for analytics

### Reviews & Ratings
- Customers rate order (1–5 stars) after delivery
- Optional text comment
- Rating linked to restaurant and optionally to rider
- Aggregate rating displayed on restaurant card and page
- Partner can respond to reviews

### Multi-language (Phase 2)
- English and Bengali support
- i18n from the start in frontend code

### SEO (Customer Website)
- Server-side rendered restaurant pages
- Structured data (JSON-LD) for restaurant schema
- Open Graph tags for social sharing
- Sitemap generation per tenant

### Accessibility
- WCAG 2.1 AA compliance target
- Keyboard navigation
- Screen reader support

### Performance Targets
- Customer website: LCP < 2.5s, FID < 100ms, CLS < 0.1
- API response: p95 < 200ms for read endpoints, p95 < 500ms for write endpoints
- Order placement: < 1s end-to-end

---

## 5.6 Enterprise & Trust Requirements (Missing-Core Additions)

### 5.6.1 Payments & Financial Integrity
- **Idempotent checkout**: Duplicate order-submit clicks must not create duplicate orders/charges.
- **Exactly-once payment capture**: Payment callback retries must remain side-effect safe.
- **Immutable financial ledger**: Commission, refunds, adjustments, and payouts must be derived from append-only ledger entries.
- **Finance reconciliation console**: Admin tools to reconcile gateway transactions vs internal orders daily.

### 5.6.2 Fraud & Abuse Controls
- Velocity controls for OTP, promo abuse, and order placement by user/device/IP.
- Promo abuse detection: shared device / shared payment method / repeated cancellation pattern checks.
- Risk scoring for COD abuse (high cancellation users can be restricted to online payment).
- Blocklist framework: phone, device fingerprint, IP, payment token.

### 5.6.3 Support & Operations
- **Support console** with full order timeline, payment events, rider events, and communication history.
- **Safe admin tools**: refunds, manual order adjustments, and re-dispatch actions must require reason input and create audit logs.
- **Customer communications log**: every outbound push/SMS/email event traceable per order.

### 5.6.4 Configurability for Multi-Vertical Commerce
- Product model must support non-food goods via configurable attributes (weight, shelf life, age restriction, fulfillment mode).
- Checkout policy engine: per-tenant toggles for scheduled orders, minimum order amount, slot-based delivery, and packaging fee.
- Tax/fee model abstraction to support non-BD tax systems without schema rewrites.

### 5.6.5 Platform Reliability UX
- Graceful checkout recovery when payment gateway is degraded (clear retry state, no silent failures).
- Customer-facing incident messaging in checkout/order tracking for known outages.
- Automatic recovery flows for stale `PENDING` orders after payment callbacks are delayed.
