-- 0002_create_enums.down.sql
-- Drop all platform ENUM types in reverse order

DROP TYPE IF EXISTS delivery_model;
DROP TYPE IF EXISTS penalty_status;
DROP TYPE IF EXISTS vehicle_type;
DROP TYPE IF EXISTS discount_type;
DROP TYPE IF EXISTS platform_source;
DROP TYPE IF EXISTS wallet_source;
DROP TYPE IF EXISTS wallet_type;
DROP TYPE IF EXISTS rider_subject;
DROP TYPE IF EXISTS refund_status;
DROP TYPE IF EXISTS accountable;
DROP TYPE IF EXISTS issue_status;
DROP TYPE IF EXISTS issue_type;
DROP TYPE IF EXISTS invoice_status;
DROP TYPE IF EXISTS promo_funder;
DROP TYPE IF EXISTS promo_apply_on;
DROP TYPE IF EXISTS promo_type;
DROP TYPE IF EXISTS txn_status;
DROP TYPE IF EXISTS payment_method;
DROP TYPE IF EXISTS payment_status;
DROP TYPE IF EXISTS pickup_status;
DROP TYPE IF EXISTS order_status;
DROP TYPE IF EXISTS price_type;
DROP TYPE IF EXISTS product_avail;
DROP TYPE IF EXISTS restaurant_type;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_role;
DROP TYPE IF EXISTS tenant_plan;
DROP TYPE IF EXISTS tenant_status;
