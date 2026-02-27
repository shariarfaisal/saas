-- ============================================================
-- 000019_add_gateway_txn_uniqueness.up.sql
-- Enforce uniqueness on gateway_transaction_id per tenant
-- for idempotent payment callback processing.
-- ============================================================

CREATE UNIQUE INDEX idx_payment_txns_gateway_id_unique
    ON payment_transactions(tenant_id, gateway_transaction_id)
    WHERE gateway_transaction_id IS NOT NULL;
