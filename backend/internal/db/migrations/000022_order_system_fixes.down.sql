DROP INDEX IF EXISTS idx_orders_pending_payment;
DROP INDEX IF EXISTS uq_modifier_option_name;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS created_at;
ALTER TABLE orders DROP COLUMN IF EXISTS packaging_fee;
