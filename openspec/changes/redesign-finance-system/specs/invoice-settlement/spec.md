## MODIFIED Requirements

### Requirement: Correct Invoice Generation Formula
The `GenerateInvoice` service method SHALL implement the exact settlement formula from docs/requirements/08:

```
gross_sales          = SUM(order_pickups.items_total) for finalized orders in period
item_discounts       = SUM(order_items.discount_amount) for items in period
vendor_promo_discounts = SUM(promo_usages.discount_amount WHERE promo.funded_by = 'vendor' AND restaurant_id matches)
commission_amount    = SUM(order_pickups.commission_amount) (populated by order system)
penalty_amount       = SUM(order_issues.penalty_amount WHERE resolved = true AND restaurant_id matches AND period overlap)
adjustment_amount    = SUM(invoice_adjustments.amount WHERE invoice_id = this invoice)
delivery_charge_total = SUM(orders.delivery_charge) for platform-managed delivery orders in period
net_payable          = gross_sales - item_discounts - vendor_promo_discounts - commission_amount - penalty_amount + adjustment_amount
```

All six fields MUST be computed from actual records, never hardcoded to zero.

#### Scenario: Invoice with vendor-funded promo
- **WHEN** a restaurant has 3 orders in the period, two of which used a vendor-funded promo totaling 150 BDT discount
- **THEN** `invoice.vendor_promo_discounts = 150.00`
- **AND** `net_payable` is reduced by 150.00

#### Scenario: Invoice with platform-funded promo
- **WHEN** a restaurant's orders include a platform-funded promo (e.g., a platform-wide promotion)
- **THEN** `invoice.vendor_promo_discounts` does NOT include the platform-funded discount amount
- **AND** `net_payable` is NOT reduced by the platform promotion

#### Scenario: Invoice with penalties
- **WHEN** a restaurant had an order issue (`order_issues`) resolved against them with `penalty_amount = 200 BDT`
- **THEN** `invoice.penalty_amount = 200.00`
- **AND** `net_payable = gross_sales - ... - 200.00`

#### Scenario: Invoice with manual adjustments
- **WHEN** a finance admin adds a 500 BDT positive adjustment to the invoice (credit to vendor)
- **THEN** `invoice.adjustment_amount = 500.00`
- **AND** `net_payable = gross_sales - ... + 500.00`

### Requirement: Promo Funding Source Classification
All promos SHALL have a `funded_by` field: `platform` or `vendor`. When `funded_by = platform`, the platform absorbs the discount and it is not charged back to the restaurant. When `funded_by = vendor`, the discount reduces the vendor's `net_payable`.

#### Scenario: Platform creates a system-wide promotion
- **WHEN** an admin creates a promo with `funded_by = platform`
- **AND** a customer uses it
- **THEN** the promo discount does not appear in the vendor invoice `vendor_promo_discounts`
- **AND** the platform ledger records the cost as a platform marketing expense

### Requirement: Delivery Charge Disposition per Restaurant
Each restaurant SHALL have a `delivery_managed_by` setting (`platform` or `vendor`). When `platform`, the delivery charge is retained by the platform and does NOT form part of vendor `net_payable`. When `vendor` (self-delivery restaurants), the delivery charge is included in the vendor settlement.

#### Scenario: Platform-managed delivery restaurant
- **WHEN** an order has `delivery_charge = 40.00` for a restaurant where `delivery_managed_by = platform`
- **THEN** the invoice `delivery_charge_retained = 40.00`
- **AND** vendor `net_payable` does NOT include the delivery charge

### Requirement: Invoice Adjustment Records
A new `invoice_adjustments` table SHALL store manual credit/debit adjustments to invoices. Each record has: `invoice_id`, `amount`, `direction (credit|debit)`, `reason`, `created_by_admin_id`, `created_at`. Admin endpoint to add adjustments SHALL be provided. Adjustments on finalized invoices SHALL not be allowed.

#### Scenario: Admin issues a credit adjustment
- **WHEN** an admin calls `POST /admin/invoices/:id/adjustments` with `{amount: 300, direction: "credit", reason: "platform error"}`
- **AND** the invoice is in `draft` status
- **THEN** the adjustment record is created
- **AND** the invoice `adjustment_amount` is recalculated on next `GET` or re-generation

#### Scenario: Adjustment on finalized invoice rejected
- **WHEN** an admin tries to add an adjustment to an invoice with `status = finalized`
- **THEN** the API returns `422 INVOICE_ALREADY_FINALIZED`
