-- Restaurant delivery management
ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS delivery_managed_by TEXT NOT NULL DEFAULT 'platform'
    CHECK (delivery_managed_by IN ('platform', 'vendor'));

-- Invoice delivery charge total
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS delivery_charge_total NUMERIC(12,2) NOT NULL DEFAULT 0;

-- Tenant billing day
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS billing_day INT NOT NULL DEFAULT 1
    CHECK (billing_day BETWEEN 1 AND 28);

-- Invoice adjustments table
CREATE TABLE IF NOT EXISTS invoice_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    amount NUMERIC(10,2) NOT NULL,
    direction TEXT NOT NULL DEFAULT 'credit' CHECK (direction IN ('credit', 'debit')),
    reason TEXT NOT NULL,
    created_by_admin_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_invoice_adjustments_invoice ON invoice_adjustments(invoice_id);

-- Subscription invoices table (SaaS billing)
CREATE TABLE IF NOT EXISTS subscription_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    amount NUMERIC(10,2) NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'overdue', 'cancelled')),
    billing_period_start DATE NOT NULL,
    billing_period_end DATE NOT NULL,
    due_date DATE NOT NULL,
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_subscription_invoices_tenant ON subscription_invoices(tenant_id, billing_period_start DESC);

-- COD cash collection records
CREATE TABLE IF NOT EXISTS cash_collection_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    rider_id UUID NOT NULL REFERENCES riders(id),
    order_id UUID NOT NULL REFERENCES orders(id),
    amount NUMERIC(10,2) NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'collected' CHECK (status IN ('collected', 'remitted', 'overdue')),
    collected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    remitted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_cash_collection_tenant ON cash_collection_records(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_cash_collection_rider ON cash_collection_records(rider_id, status);

-- Reconciliation alerts for payment mismatches
CREATE TABLE IF NOT EXISTS reconciliation_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),
    payment_transaction_id UUID REFERENCES payment_transactions(id),
    alert_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'resolved')),
    resolution_notes TEXT,
    resolved_by UUID REFERENCES users(id),
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_reconciliation_alerts_status ON reconciliation_alerts(status, created_at DESC);
