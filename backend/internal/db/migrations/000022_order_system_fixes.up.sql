-- Add packaging_fee column to orders
ALTER TABLE orders ADD COLUMN IF NOT EXISTS packaging_fee NUMERIC(10,2) NOT NULL DEFAULT 0;

-- Add created_at to inventory_items if missing
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Add unique index on modifier options (if not exists)
CREATE UNIQUE INDEX IF NOT EXISTS uq_modifier_option_name ON product_modifier_options(modifier_group_id, name);

-- Index for pending payment timeout queries
CREATE INDEX IF NOT EXISTS idx_orders_pending_payment ON orders(tenant_id, created_at) WHERE status = 'pending' AND payment_status = 'unpaid';
