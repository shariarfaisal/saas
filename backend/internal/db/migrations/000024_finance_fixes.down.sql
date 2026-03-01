DROP TABLE IF EXISTS reconciliation_alerts;
DROP TABLE IF EXISTS cash_collection_records;
DROP TABLE IF EXISTS subscription_invoices;
DROP TABLE IF EXISTS invoice_adjustments;
ALTER TABLE tenants DROP COLUMN IF EXISTS billing_day;
ALTER TABLE invoices DROP COLUMN IF EXISTS delivery_charge_total;
ALTER TABLE restaurants DROP COLUMN IF EXISTS delivery_managed_by;
