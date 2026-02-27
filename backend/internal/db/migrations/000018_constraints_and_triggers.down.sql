-- ============================================================
-- 000018_constraints_and_triggers.down.sql
-- ============================================================

DROP INDEX IF EXISTS idx_users_name_trgm;

ALTER TABLE platform_configs    DROP CONSTRAINT IF EXISTS fk_platform_configs_updated_by;
ALTER TABLE order_analytics     DROP CONSTRAINT IF EXISTS fk_order_analytics_order_id;
ALTER TABLE reviews             DROP CONSTRAINT IF EXISTS fk_reviews_order_id;
ALTER TABLE promo_usages        DROP CONSTRAINT IF EXISTS fk_promo_usages_order_id;
ALTER TABLE wallet_transactions DROP CONSTRAINT IF EXISTS fk_wallet_txns_order_id;
ALTER TABLE refunds             DROP CONSTRAINT IF EXISTS fk_refunds_order_id;
ALTER TABLE payment_transactions DROP CONSTRAINT IF EXISTS fk_payment_txns_order_id;
ALTER TABLE rider_penalties     DROP CONSTRAINT IF EXISTS fk_rider_penalties_order_id;
ALTER TABLE rider_earnings      DROP CONSTRAINT IF EXISTS fk_rider_earnings_payout_id;
ALTER TABLE rider_earnings      DROP CONSTRAINT IF EXISTS fk_rider_earnings_order_id;
ALTER TABLE rider_location_history DROP CONSTRAINT IF EXISTS fk_rider_loc_history_order_id;
ALTER TABLE inventory_adjustments DROP CONSTRAINT IF EXISTS fk_inv_adj_order_id;
ALTER TABLE orders              DROP CONSTRAINT IF EXISTS fk_orders_promo_id;
