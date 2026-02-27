-- ============================================================
-- 000018_constraints_and_triggers.up.sql
-- Cross-table FK constraints that couldn't be added earlier due
-- to circular or forward dependencies, plus subscription plan seed.
-- ============================================================

-- ---- Deferred FK: orders → promos ----
ALTER TABLE orders
    ADD CONSTRAINT fk_orders_promo_id
    FOREIGN KEY (promo_id) REFERENCES promos(id);

-- ---- Deferred FK: inventory_adjustments → orders ----
ALTER TABLE inventory_adjustments
    ADD CONSTRAINT fk_inv_adj_order_id
    FOREIGN KEY (order_id) REFERENCES orders(id);

-- ---- Deferred FK: rider_location_history → orders ----
ALTER TABLE rider_location_history
    ADD CONSTRAINT fk_rider_loc_history_order_id
    FOREIGN KEY (order_id) REFERENCES orders(id);

-- ---- Deferred FK: rider_earnings → orders ----
ALTER TABLE rider_earnings
    ADD CONSTRAINT fk_rider_earnings_order_id
    FOREIGN KEY (order_id) REFERENCES orders(id);

-- ---- Deferred FK: rider_earnings → rider_payouts ----
ALTER TABLE rider_earnings
    ADD CONSTRAINT fk_rider_earnings_payout_id
    FOREIGN KEY (payout_id) REFERENCES rider_payouts(id);

-- ---- Deferred FK: rider_penalties → orders ----
ALTER TABLE rider_penalties
    ADD CONSTRAINT fk_rider_penalties_order_id
    FOREIGN KEY (order_id) REFERENCES orders(id);

-- ---- Deferred FK: payment_transactions → orders ----
ALTER TABLE payment_transactions
    ADD CONSTRAINT fk_payment_txns_order_id
    FOREIGN KEY (order_id) REFERENCES orders(id);

-- ---- Deferred FK: refunds → orders ----
ALTER TABLE refunds
    ADD CONSTRAINT fk_refunds_order_id
    FOREIGN KEY (order_id) REFERENCES orders(id);

-- ---- Deferred FK: wallet_transactions → orders ----
ALTER TABLE wallet_transactions
    ADD CONSTRAINT fk_wallet_txns_order_id
    FOREIGN KEY (order_id) REFERENCES orders(id);

-- ---- Deferred FK: promo_usages → orders ----
ALTER TABLE promo_usages
    ADD CONSTRAINT fk_promo_usages_order_id
    FOREIGN KEY (order_id) REFERENCES orders(id);

-- ---- Deferred FK: reviews → orders ----
ALTER TABLE reviews
    ADD CONSTRAINT fk_reviews_order_id
    FOREIGN KEY (order_id) REFERENCES orders(id);

-- ---- Deferred FK: order_analytics → orders ----
ALTER TABLE order_analytics
    ADD CONSTRAINT fk_order_analytics_order_id
    FOREIGN KEY (order_id) REFERENCES orders(id);

-- ---- Deferred FK: platform_configs → users (updated_by) ----
ALTER TABLE platform_configs
    ADD CONSTRAINT fk_platform_configs_updated_by
    FOREIGN KEY (updated_by) REFERENCES users(id);

-- ---- Full-text search index on users (name) ----
CREATE INDEX idx_users_name_trgm ON users USING gin(name gin_trgm_ops) WHERE deleted_at IS NULL;
