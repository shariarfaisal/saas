# 01 — Product Vision

## 1.1 What We Are Building

A **multi-tenant food-commerce SaaS platform** that enables vendors (restaurant groups, cloud kitchens, grocery stores, or any food brand) to launch and operate their own branded online ordering business — without managing infrastructure.

The platform is inspired by the model of FoodPanda, Talabat, and similar aggregators, but operates differently:  
- Vendors **do not** list on a shared marketplace.  
- Each vendor gets **their own branded storefront**, their own partner dashboard, their own rider pool (optional), and their own customer data.  
- The **platform operator** (us) deploys the infrastructure, charges vendors a commission on sales or a subscription fee, and retains full visibility through a super-admin panel.

---

## 1.2 Target Market

**Launch market:** Bangladesh (Dhaka-first, then nationwide)  
**Expansion target:** South Asia, then any emerging market  
**Global standard:** Architecture and features must be on par with global leaders (FoodPanda, Deliveroo, DoorDash) — we do not cut corners on design quality.

**Primary vendor type (Phase 1):** Restaurant groups — brands that operate multiple physical locations / cloud kitchens.  
**Secondary vendor types (Phase 2+):** Grocery delivery, pharmacy, dark stores, general retail.

---

## 1.3 Core Value Proposition

| For Vendors | For End Customers | For Platform Operator |
|-------------|-------------------|----------------------|
| Launch an online food ordering business in days | Order from favourite local brands with fast delivery | Earn commission on every order across all vendors |
| Own your customers — no shared marketplace | Real-time order tracking | Full visibility and control over all tenants |
| Manage multiple restaurants from one dashboard | Multiple payment options (bKash, card, COD) | Scalable revenue model |
| Built-in rider management OR third-party courier | Promotions and loyalty rewards | Plug in new vendors easily |
| Detailed financial and sales reporting | Personalised recommendations | Modular architecture for customisation |

---

## 1.4 Business Model

The platform earns revenue through:

1. **Commission model** — A configurable percentage of each order's subtotal, set per vendor or per restaurant.
2. **Subscription model (future)** — Monthly/annual SaaS fee tiers with feature gates.
3. **Setup fee** — One-time onboarding fee per vendor (optional).

Vendors set their own delivery charges. The platform takes commission from the food subtotal (before delivery charge and taxes).

---

## 1.5 What Makes This Different From the Legacy System

| Legacy (Munchies) | New SaaS Platform |
|-------------------|-------------------|
| Single brand (Munchies only) | Any number of vendors / brands |
| Parse Server + MongoDB | Go + PostgreSQL (typed, performant, maintainable) |
| No tenant isolation | Every resource is tenant-scoped |
| Hard-coded Bangladesh payment gateways | Pluggable payment gateway adapters |
| No subscription/billing module | Platform billing for vendors |
| Admin = internal staff only | Super-admin + vendor-level admin separation |
| No scalable architecture | Designed to scale to millions of orders |
| Technical debt accumulated | Built from scratch with modern practices |

---

## 1.6 Phase Roadmap

### Phase 1 — Core Platform (MVP)
- Vendor onboarding and management
- Restaurant & menu management
- Customer ordering website
- Order management (platform-owned riders)
- Basic payment integration (bKash + COD)
- Partner portal (restaurant dashboard, orders, reports)
- Super admin panel

### Phase 2 — Growth Features
- Customer loyalty / wallet / points system
- Advanced promotions & vouchers
- Push notifications (mobile PWA)
- Rider app (PWA or React Native)
- Multi-zone delivery pricing
- Review & rating system
- SEO-optimised restaurant pages

### Phase 3 — Platform Maturity
- Vendor self-serve onboarding
- Subscription billing & plan management
- Multi-currency support
- Grocery / dark store vertical
- Third-party courier integration (Pathao, Shohoz)
- White-label mobile apps per vendor
- Advanced analytics & BI dashboards
- Webhook system for vendor integrations

---

## 1.7 Non-Goals

- **No shared marketplace / aggregator model** in Phase 1 (customers can't browse all vendors in one app).
- **No microservices** — added complexity is not justified at this stage.
- **No mobile native apps** in Phase 1 (PWA is sufficient for riders; vendor apps via responsive web).
- **No multi-currency** in Phase 1 (BDT only, but architecture must support it later).
