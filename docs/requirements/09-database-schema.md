# 09 — Database Schema

> **Status:** Redesigned (2026-02-27). All tables are live in migrations `000001`–`000018` under `backend/internal/db/migrations/`. The legacy MongoDB-derived schema has been replaced with a fully relational, production-grade PostgreSQL 17 design.

---

## Table of Contents

1. [Design Principles](#1-design-principles)
2. [ENUM Types](#2-enum-types)
3. [Migration Index](#3-migration-index)
4. [Schema Groups](#4-schema-groups)
   - [4.1 Platform & Tenant](#41-platform--tenant)
   - [4.2 Users & Auth](#42-users--auth)
   - [4.3 Geography](#43-geography)
   - [4.4 Restaurants & Catalog](#44-restaurants--catalog)
   - [4.5 Inventory](#45-inventory)
   - [4.6 Orders](#46-orders)
   - [4.7 Riders](#47-riders)
   - [4.8 Promotions](#48-promotions)
   - [4.9 Payments](#49-payments)
   - [4.10 Finance](#410-finance)
   - [4.11 Content & CMS](#411-content--cms)
   - [4.12 Analytics](#412-analytics)
   - [4.13 Notifications & Outbox](#413-notifications--outbox)
5. [Cross-Table FK Map](#5-cross-table-fk-map)
6. [Key Design Decisions](#6-key-design-decisions)

---

## 1. Design Principles

| Principle | Implementation |
|---|---|
| **UUID PKs everywhere** | `gen_random_uuid()` via `pgcrypto` |
| **Soft delete on user data** | `deleted_at TIMESTAMPTZ` on `users`, `restaurants`, `products` |
| **Append-only ledgers** | `wallet_transactions` — never update, always insert |
| **JSONB snapshots** | `order_items.selected_modifiers` preserves modifier state at order time |
| **`updated_at` trigger** | `fn_set_updated_at()` applied to every mutable table — no application-layer responsibility |
| **Circular FK resolution** | Tables with circular dependencies (orders ↔ promos ↔ users) are wired via `ALTER TABLE … ADD CONSTRAINT` in migration `000018` |
| **Transactional outbox** | `outbox_events` ensures side-effects (notifications, analytics) are published exactly once |
| **Idempotency** | `idempotency_keys` prevents duplicate order submissions |
| **Multi-tenancy** | Every domain table carries `tenant_id UUID NOT NULL` for row-level isolation |

---

## 2. ENUM Types

Defined in migrations `000002` and `000003`.

| Type | Values |
|---|---|
| `user_role` | `customer`, `tenant_owner`, `tenant_admin`, `restaurant_manager`, `restaurant_staff`, `rider`, `platform_admin`, `platform_support`, `platform_finance` |
| `user_status` | `active`, `suspended`, `deleted` |
| `gender_type` | `male`, `female`, `other`, `prefer_not_to_say` |
| `tenant_status` | `pending`, `active`, `suspended`, `cancelled` |
| `tenant_plan` | `starter`, `growth`, `enterprise` |
| `subscription_status` | `trialing`, `active`, `past_due`, `cancelled`, `expired` |
| `billing_cycle` | `monthly`, `annual` |
| `restaurant_type` | `restaurant`, `cloud_kitchen`, `store`, `dark_store` |
| `order_status` | `pending`, `created`, `confirmed`, `preparing`, `ready`, `picked`, `delivered`, `cancelled`, `rejected` |
| `pickup_status` | `new`, `confirmed`, `preparing`, `ready`, `picked`, `rejected` |
| `payment_method` | `cod`, `bkash`, `aamarpay`, `sslcommerz`, `wallet`, `card` |
| `payment_status` | `unpaid`, `paid`, `refunded`, `partially_refunded` |
| `txn_status` | `pending`, `success`, `failed`, `refunded`, `cancelled` |
| `refund_status` | `pending`, `approved`, `rejected`, `processed` |
| `wallet_type` | `credit`, `debit` |
| `wallet_source` | `cashback`, `referral`, `welcome`, `refund`, `order_payment`, `admin_adjustment` |
| `invoice_status` | `draft`, `finalized`, `paid` |
| `payout_status` | `pending`, `processing`, `completed`, `failed` |
| `discount_type` | `fixed`, `percent` |
| `promo_type` | `fixed`, `percent` |
| `promo_apply_on` | `all_items`, `category`, `specific_restaurant`, `delivery_charge` |
| `promo_funder` | `vendor`, `platform`, `restaurant` |
| `product_avail` | `available`, `unavailable`, `out_of_stock` |
| `vehicle_type` | `bicycle`, `motorcycle`, `car` |
| `rider_subject` | `attendance_in`, `attendance_out`, `picked`, `delivered`, `in_hub`, `location_update` |
| `penalty_status` | `pending`, `cleared`, `appealed` |
| `issue_type` | `wrong_item`, `missing_item`, `quality_issue`, `late_delivery`, `other` |
| `issue_status` | `open`, `resolved`, `closed` |
| `accountable` | `restaurant`, `rider`, `platform` |
| `actor_type` | `customer`, `restaurant`, `rider`, `platform_admin`, `system` |
| `platform_source` | `web`, `ios`, `android`, `pos` |
| `delivery_model` | `zone_based`, `distance_based` |
| `inventory_adjustment_reason` | `opening_stock`, `purchase`, `manual_adjustment`, `order_reserve`, `order_release`, `order_consume`, `damage_loss`, `stock_return` |
| `notification_channel` | `push`, `sms`, `email`, `in_app` |
| `notification_status` | `pending`, `sent`, `delivered`, `failed`, `read` |
| `outbox_event_status` | `pending`, `processing`, `processed`, `failed` |
| `link_target_type` | `restaurant`, `product`, `category`, `url`, `promo` |
| `media_type` | `image`, `video` |
| `price_type` | `flat`, `variant` *(legacy; superseded by `has_modifiers` boolean on products)* |

---

## 3. Migration Index

| # | File | Tables Created |
|---|---|---|
| 001 | `000001_init_extensions` | — (pgcrypto, pg_trgm extensions) |
| 002 | `000002_create_enums` | — (28 ENUM types) |
| 003 | `000003_add_enum_types` | — (11 ENUM types + `fn_set_updated_at()`) |
| 004 | `000004_create_platform_tables` | `subscription_plans`, `tenants`, `tenant_subscriptions`, `platform_configs`, `tenant_payment_gateways` |
| 005 | `000005_create_users` | `users`, `user_addresses`, `otp_verifications`, `refresh_tokens`, `idempotency_keys` |
| 006 | `000006_create_geography` | `hubs`, `hub_coverage_areas`, `delivery_zone_configs` |
| 007 | `000007_create_restaurants` | `restaurants`, `restaurant_operating_hours`, `restaurant_staff_assignments` |
| 008 | `000008_create_catalog` | `categories`, `products`, `product_modifier_groups`, `product_modifier_options`, `product_discounts` |
| 009 | `000009_create_inventory` | `inventory_items`, `inventory_adjustments` |
| 010 | `000010_create_orders` | `orders`, `order_items`, `order_pickups`, `order_timeline_events`, `order_issues`, `order_issue_messages` |
| 011 | `000011_create_riders` | `riders`, `rider_locations`, `rider_location_history`, `rider_attendance`, `rider_earnings`, `rider_penalties`, `rider_payouts` |
| 012 | `000012_create_promos` | `promos`, `promo_restaurant_restrictions`, `promo_category_restrictions`, `promo_user_eligibility`, `promo_usages` |
| 013 | `000013_create_payments` | `payment_transactions`, `refunds`, `wallet_transactions` |
| 014 | `000014_create_finance` | `invoices` |
| 015 | `000015_create_content` | `banners`, `homepage_sections`, `stories`, `user_favourites`, `reviews` |
| 016 | `000016_create_analytics` | `order_analytics` |
| 017 | `000017_create_notifications` | `notification_preferences`, `notifications`, `outbox_events` |
| 018 | `000018_constraints_and_triggers` | — (deferred FK constraints + full-text index) |

**Total: 54 tables** (excluding `schema_migrations` tracking table)

---

## 4. Schema Groups

### 4.1 Platform & Tenant

#### `subscription_plans`
| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | |
| `name` | TEXT | "Starter", "Growth", "Enterprise" |
| `slug` | TEXT UNIQUE | `starter`, `growth`, `enterprise` |
| `price_monthly` | NUMERIC(12,2) | BDT |
| `price_annual` | NUMERIC(12,2) | BDT |
| `max_restaurants` | INT | NULL = unlimited |
| `max_riders` | INT | NULL = unlimited |
| `commission_rate` | NUMERIC(5,2) | % |
| `features` | JSONB | Feature flags |

#### `tenants`
| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | |
| `slug` | TEXT UNIQUE | URL-safe identifier |
| `name` | TEXT | Display name |
| `status` | `tenant_status` | |
| `plan` | `tenant_plan` | |
| `commission_rate` | NUMERIC(5,2) | Overrides plan default |
| `custom_domain` | TEXT UNIQUE | |
| `timezone` | TEXT | Default `Asia/Dhaka` |
| `currency` | TEXT | Default `BDT` |

#### `tenant_subscriptions`
Billing lifecycle per tenant. Each active tenant has one current subscription.

#### `platform_configs`
Global key-value store. `is_public=true` entries are exposed to the frontend.

#### `tenant_payment_gateways`
Per-tenant enabled gateways and credentials (secrets stored in vault externally; JSONB holds non-secret config).

---

### 4.2 Users & Auth

#### `users`
| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | |
| `tenant_id` | UUID FK → tenants | NULL for platform admins |
| `role` | `user_role` | |
| `name` | TEXT | |
| `phone` | TEXT UNIQUE | E.164 Bangladesh number |
| `email` | TEXT UNIQUE | |
| `avatar_url` | TEXT | |
| `wallet_balance` | NUMERIC(12,2) | Denormalized; source of truth is `wallet_transactions` |
| `last_login_at` | TIMESTAMPTZ | |
| `last_login_ip` | INET | |
| `deleted_at` | TIMESTAMPTZ | Soft delete |

#### `user_addresses`
Multiple delivery addresses per user. `is_default` ensures exactly one default per user.

#### `otp_verifications`
Phone/email OTP codes with expiry and attempt counting. One active OTP per (`user_id`, `purpose`).

#### `refresh_tokens`
JWT refresh token store. Tokens are rotated on use; previous tokens invalidated.

#### `idempotency_keys`
Prevents duplicate submissions. `expires_at` set to 24 h. Response JSONB cached after first successful execution.

---

### 4.3 Geography

#### `hubs`
A hub is a delivery zone grouping. All restaurants and riders belong to a hub.

#### `hub_coverage_areas`
Named polygon sub-areas within a hub for fine-grained zone reporting.

#### `delivery_zone_configs`
Per-tenant delivery fee rules (flat fee, per-km rate, free delivery threshold, surge multiplier).

---

### 4.4 Restaurants & Catalog

#### `restaurants`
| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | |
| `tenant_id` | UUID FK | |
| `hub_id` | UUID FK → hubs | |
| `type` | `restaurant_type` | |
| `slug` | TEXT | Unique within tenant |
| `commission_rate` | NUMERIC(5,2) | NULL = inherit from tenant |
| `rating` | NUMERIC(3,2) | Denormalized avg; updated by batch job |
| `total_reviews` | INT | Denormalized count |
| `is_open` | BOOLEAN | Real-time open/closed toggle |
| `deleted_at` | TIMESTAMPTZ | Soft delete |

#### `restaurant_operating_hours`
Seven rows per restaurant (one per day of week). Break times included.

#### `restaurant_staff_assignments`
Junction table linking `users` to `restaurants` for staff roles.

#### `categories`
Hierarchical with self-referencing `parent_id`. Scoped to tenant.

#### `products`
| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | |
| `restaurant_id` | UUID FK | |
| `category_id` | UUID FK | |
| `has_modifiers` | BOOLEAN | Quick flag — avoids join on list queries |
| `availability` | `product_avail` | |
| `base_price` | NUMERIC(10,2) | |
| `vat_rate` | NUMERIC(5,2) | % |
| `deleted_at` | TIMESTAMPTZ | Soft delete |

#### `product_modifier_groups`
Defines a group of options (e.g., "Size", "Extras"). `min_required ≥ 1` = required variant; `min_required = 0` = optional addon.

#### `product_modifier_options`
Individual options within a group. `additional_price` added to base product price.

#### `product_discounts`
Time-bounded discounts. `applies_to_all=true` ignores `applicable_days`.

---

### 4.5 Inventory

#### `inventory_items`
One row per product-variant. `quantity` is current stock; negative allowed for pre-order.

#### `inventory_adjustments`
Full audit trail for every stock change. `order_id` FK wired in `000018`.

---

### 4.6 Orders

#### `orders`
| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | |
| `tenant_id` | UUID FK | |
| `user_id` | UUID FK → users | |
| `restaurant_id` | UUID FK | |
| `rider_id` | UUID FK → riders | Assigned after confirmation |
| `promo_id` | UUID | FK wired in 000018 |
| `status` | `order_status` | |
| `payment_method` | `payment_method` | |
| `payment_status` | `payment_status` | |
| `subtotal` | NUMERIC(12,2) | Before discounts |
| `item_discount` | NUMERIC(12,2) | Product-level discounts |
| `promo_discount` | NUMERIC(12,2) | Promo-applied discount |
| `vat_total` | NUMERIC(12,2) | |
| `delivery_charge` | NUMERIC(10,2) | |
| `total_amount` | NUMERIC(12,2) | Final charged amount |
| `platform_source` | `platform_source` | |
| `idempotency_key` | TEXT | From `idempotency_keys` table |

#### `order_items`
Line items. `selected_modifiers` JSONB snapshot: `[{"group_id","group_name","option_id","option_name","additional_price"}]` — frozen at order time for historical accuracy.

#### `order_pickups`
One row per (order, restaurant) for multi-restaurant orders. Tracks per-restaurant preparation status.

#### `order_timeline_events`
Append-only status history. Enables full audit trail and SLA calculation.

#### `order_issues`
Dispute/complaint per order. Linked to refund when `refundable=true`.

#### `order_issue_messages`
Threaded messages on an issue between customer, restaurant, and platform.

---

### 4.7 Riders

#### `riders`
Extends `users` (FK `user_id`). Contains delivery-specific fields: `vehicle_type`, `rating`, `total_deliveries`, bank account for payouts.

#### `rider_locations`
Single-row-per-rider real-time GPS. Upserted on every location ping.

#### `rider_location_history`
Full GPS trail. `order_id` FK wired in `000018`.

#### `rider_attendance`
Check-in/check-out records with hub assignment.

#### `rider_earnings`
Per-delivery earnings breakdown. FK to `orders` and `rider_payouts` wired in `000018`.

#### `rider_penalties`
Deductions applied to earnings. FK to `orders` wired in `000018`.

#### `rider_payouts`
Settlement batch records. Rider earnings aggregated and paid via bank transfer.

---

### 4.8 Promotions

#### `promos`
| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | |
| `tenant_id` | UUID FK | |
| `code` | TEXT | Unique within tenant |
| `type` | `promo_type` | `fixed` or `percent` |
| `apply_on` | `promo_apply_on` | |
| `funder` | `promo_funder` | Who bears the discount cost |
| `discount_value` | NUMERIC(10,2) | |
| `max_discount_cap` | NUMERIC(10,2) | Max discount for `percent` type |
| `min_order_value` | NUMERIC(10,2) | |
| `usage_limit` | INT | NULL = unlimited |
| `per_user_limit` | INT | NULL = unlimited |

#### Restriction tables
- `promo_restaurant_restrictions` — junction: promo applies only to listed restaurants
- `promo_category_restrictions` — junction: promo applies only to listed categories
- `promo_user_eligibility` — junction: promo restricted to specific users (targeted promos)

#### `promo_usages`
One row per redemption. `order_id` FK wired in `000018`.

---

### 4.9 Payments

#### `payment_transactions`
One row per payment attempt (multiple attempts possible per order). `order_id` FK wired in `000018`.

#### `refunds`
Refund requests linked to a transaction and optionally to an `order_issue`. `order_id` FK wired in `000018`.

#### `wallet_transactions`
Append-only ledger. `balance_after` stored for audit. `order_id` FK wired in `000018`.

---

### 4.10 Finance

#### `invoices`
Settlement invoice per restaurant per billing period. Contains full revenue breakdown:
- `gross_sales` → `net_sales` (after item and vendor promo discounts)
- `commission_amount` = `net_sales × commission_rate`
- `net_payable` = `net_sales − commission_amount − penalties + adjustment`

---

### 4.11 Content & CMS

#### `banners`
Promotional banners with optional hub scoping, scheduling (`starts_at`/`ends_at`), and deep-link targets.

#### `homepage_sections`
Curated or algorithmic content sections. `content_type` determines what `item_ids` refer to. `filter_rule` JSONB enables dynamic "Top Rated" / "New" sections.

#### `stories`
Short-lived media (image/video) with automatic expiry (`expires_at`).

#### `user_favourites`
Junction table: user ↔ restaurant. PK on `(user_id, restaurant_id)`.

#### `reviews`
One review per `(order_id, user_id)`. `is_published` gates moderation. Restaurant can post a public reply. `order_id` FK wired in `000018`.

---

### 4.12 Analytics

#### `order_analytics`
Denormalized fact table populated by background worker after order reaches a terminal state. Pre-computed time dimensions (`order_date`, `order_hour`, `order_day_of_week`, `order_week`, `order_month`, `order_year`) enable GROUP BY aggregations without expensive date functions. `order_id` FK wired in `000018`.

---

### 4.13 Notifications & Outbox

#### `notification_preferences`
One row per user. Channel and category toggles for notification delivery.

#### `notifications`
Persisted log of all sent notifications. Powers the in-app notification centre and delivery status tracking.

#### `outbox_events`
Transactional outbox pattern. Inserted inside the same DB transaction as the domain event. Background worker polls `status='pending'` rows and publishes to Redis / job queues. Max 5 attempts with exponential backoff via `next_retry_at`.

---

## 5. Cross-Table FK Map

FKs added in migration `000018` (all circular or forward dependencies):

| Table | Column | References |
|---|---|---|
| `orders` | `promo_id` | `promos(id)` |
| `inventory_adjustments` | `order_id` | `orders(id)` |
| `rider_location_history` | `order_id` | `orders(id)` |
| `rider_earnings` | `order_id` | `orders(id)` |
| `rider_earnings` | `payout_id` | `rider_payouts(id)` |
| `rider_penalties` | `order_id` | `orders(id)` |
| `payment_transactions` | `order_id` | `orders(id)` |
| `refunds` | `order_id` | `orders(id)` |
| `wallet_transactions` | `order_id` | `orders(id)` |
| `promo_usages` | `order_id` | `orders(id)` |
| `reviews` | `order_id` | `orders(id)` |
| `order_analytics` | `order_id` | `orders(id)` |
| `platform_configs` | `updated_by` | `users(id)` |

---

## 6. Key Design Decisions

### Modifier Groups (replacing variants + addons)
`product_modifier_groups` + `product_modifier_options` provides a unified model:
- `min_required ≥ 1, max_allowed = 1` → single-select required variant (e.g., "Size: S/M/L")
- `min_required = 0` → optional addon (e.g., "Extra toppings")

`products.has_modifiers` is a boolean quick-flag set by the application when modifier groups are added/removed. This avoids a `LEFT JOIN` on every product listing query.

### JSONB Snapshot in `order_items.selected_modifiers`
Modifier prices and names are snapshotted at order creation time. This ensures historical orders are unaffected by subsequent menu changes.

### Promo Restrictions as Junction Tables
Instead of `UUID[]` arrays, proper junction tables (`promo_restaurant_restrictions`, `promo_category_restrictions`, `promo_user_eligibility`) allow indexed lookups, FK integrity, and easy audit.

### Wallet is Append-Only
`wallet_transactions` is never updated. `users.wallet_balance` is a denormalized read-optimized cache; the true balance is always derived from the append-only ledger.

### Operational Date Rollover
Bangladesh operational day rolls over at 05:00 AM (UTC+6). The `timeutil.OperationalDate()` Go helper handles this. Analytics `order_date` is computed using this logic.

### Deferred FK Strategy
13 FK constraints are applied in migration `000018` to resolve circular dependencies. This is a standard PostgreSQL pattern; `DEFERRABLE INITIALLY DEFERRED` was considered but explicit ordering in migration files is clearer.
