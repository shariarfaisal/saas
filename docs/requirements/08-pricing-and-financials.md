# 08 — Pricing & Financials

## 8.1 Revenue Model

### Platform Commission
The platform earns a percentage of each order's **food subtotal** (before delivery charge, after item discounts, before promo discount).

```
platform_commission = (item_subtotal - item_discount) × commission_rate / 100
```

Commission rate hierarchy:
1. Restaurant-specific override (`restaurants.commission_rate`) — if set, used for that restaurant
2. Tenant default (`tenants.commission_rate`) — fallback for all restaurants
3. Platform default — fallback if tenant has no rate set (typically 10–15%)

### Commission Type
- `percentage` — standard % of order subtotal (default)
- `flat` — flat amount per order (future, for specific agreements)

---

## 8.2 Delivery Charge Model

Delivery charges are determined by **delivery zone pricing**. Two models supported:

### Model A — Zone-Based (Area Name)
```
delivery_zones (per tenant or per hub)
  area_name → delivery_charge
```
Example:
```
Gulshan 1   → ৳40
Gulshan 2   → ৳40
Banani      → ৳50
Mirpur 10   → ৳70
```

### Model B — Distance-Based (Phase 2)
Using the Barikoi API distance calculation:
```
if distance <= 3km:  charge = ৳40
if distance <= 5km:  charge = ৳60
if distance <= 8km:  charge = ৳80
else:                charge = ৳100
```

Tenants can configure which model they use.

---

## 8.3 VAT Handling

Per restaurant (set in `restaurants.vat_rate`):

```
if restaurant.is_vat_included:
  # VAT is already in the product price — we extract it
  vat_amount = round(item_total × vat_rate / (100 + vat_rate))
else:
  # VAT is added on top
  vat_amount = round((item_total - item_discount) × vat_rate / 100)
```

VAT collected is tracked separately in invoices and sales reports for compliance.

---

## 8.4 Promo Funding

Who pays for a promo discount determines how it appears in the invoice:

| Promo Created By | Funded By | Invoice Impact |
|-----------------|-----------|----------------|
| Tenant/vendor (for their restaurants) | Vendor | Deducted from vendor's payable |
| Platform (e.g., welcome offer) | Platform | Not deducted from vendor; platform absorbs |
| Restaurant-specific (by restaurant manager) | Restaurant | Deducted from that restaurant's payable |

A `promo.funded_by` field: `ENUM (vendor, platform, restaurant)` tracks this.

---

## 8.5 Settlement Period

Default: **Daily settlement** (platform generates invoice for each restaurant per day).

Configurable per tenant: daily, weekly, bi-weekly.

---

## 8.6 Invoice Calculation

For each restaurant, per settlement period:

```
total_sales         = sum of (item_subtotal - item_discount) for DELIVERED orders
                      where pickup.restaurant_id = this restaurant

vat_collected       = sum of vat for this restaurant's items in delivered orders

vendor_promo_discount = sum of promo_discount for promos funded by vendor for this restaurant's orders

commission_amount   = total_sales × commission_rate / 100

penalty             = sum of restaurant_penalty for order_issues in this period
                      (only where restaurant is accountable and penalty applied)

adjustment          = manual adjustment by admin (positive = extra charge, negative = credit)

net_payable = total_sales
            - commission_amount
            - vendor_promo_discount
            - penalty
            + adjustment
            (VAT is separate — platform handles VAT remittance)
```

**net_payable** is what the platform owes the restaurant.  
If net_payable < 0, the restaurant owes the platform (deducted from next invoice).

---

## 8.7 Restaurant Payment Dashboard

For the partner portal, each restaurant's finance page shows:

| Metric | Description |
|--------|-------------|
| Gross Sales | Total order subtotals in period |
| Product Discounts | Discounts funded from restaurant's pricing |
| Promo Discounts | Vendor-funded promo amounts applied |
| Commission | Platform commission deducted |
| VAT Collected | VAT collected on behalf of government |
| Penalties | Issue penalties applied |
| Adjustments | Admin manual adjustments |
| **Net Payable** | **Amount platform owes the restaurant** |

---

## 8.8 Rider Earnings

Rider earnings are tracked separately from restaurant invoices.

```
rider_earnings_per_order = base_delivery_earning + distance_bonus + peak_hour_bonus

base_delivery_earning   = configured per tenant (e.g., ৳40 per delivery)
distance_bonus          = 0 (default) or extra per km beyond threshold
peak_hour_bonus         = configured per tenant for peak hours (e.g., ৳10 extra)
```

Rider balance accumulates until:
- Weekly batch payout (mobile banking transfer)
- Or manual payout by admin

Rider penalties are deducted from pending balance before payout.

---

## 8.9 Wallet / Loyalty Points

Customer wallet is denominated in the same currency (BDT).

**Earning points:**
- Cashback from promo codes (promo.cashback_amount credited on order delivery)
- Welcome bonus on registration
- Referral bonus (when referred user places first order)

**Spending points:**
- Applied as payment method at checkout (1 point = ৳1)
- Minimum wallet spend per order configurable
- Maximum wallet spend per order configurable (e.g., max 50% of order total)

**Expiry (Phase 2):**
- Points earned expire after 90 days of inactivity

```
wallet_transactions
  id              UUID  PK
  user_id         UUID  FK → users
  tenant_id       UUID  FK → tenants
  type            ENUM  (credit, debit)
  source          ENUM  (cashback, referral, welcome, refund, order_payment, admin_adjustment)
  amount          NUMERIC(12,2)
  balance_after   NUMERIC(12,2)
  order_id        UUID  FK → orders NULLABLE
  note            TEXT  NULLABLE
  created_at      TIMESTAMPTZ
```

---

## 8.10 Subscription Billing (Phase 2)

Vendors will be billed on a monthly/annual basis for platform subscription:

```
subscription_plans
  id              UUID  PK
  name            TEXT          -- "Starter", "Growth", "Enterprise"
  price_monthly   NUMERIC(12,2)
  price_annual    NUMERIC(12,2)
  features        JSONB         -- which features are included
  restaurant_limit INT          -- max number of restaurants
  rider_limit     INT NULLABLE  -- max riders (null = unlimited)
  commission_rate NUMERIC(5,2)  -- platform commission for this plan

tenant_subscriptions
  id              UUID  PK
  tenant_id       UUID  FK → tenants
  plan_id         UUID  FK → subscription_plans
  billing_cycle   ENUM  (monthly, annual)
  status          ENUM  (active, past_due, cancelled)
  current_period_start DATE
  current_period_end   DATE
  next_billing_date    DATE
  payment_method  JSONB         -- stored payment method for auto-billing
  created_at      TIMESTAMPTZ
```
