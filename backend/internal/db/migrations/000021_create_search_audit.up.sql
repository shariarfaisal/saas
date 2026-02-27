-- ============================================================
-- 000021_create_search_audit.up.sql
-- Search logs and audit logs
-- ============================================================

CREATE TABLE search_logs (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID        NOT NULL REFERENCES tenants(id),
    user_id     UUID        REFERENCES users(id),
    query       TEXT        NOT NULL,
    search_type TEXT        NOT NULL DEFAULT 'all',
    result_count INT        NOT NULL DEFAULT 0,
    filters     JSONB       NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_search_logs_tenant ON search_logs(tenant_id, created_at DESC);
CREATE INDEX idx_search_logs_query ON search_logs(tenant_id, query);

CREATE TABLE audit_logs (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID        REFERENCES tenants(id),
    actor_id        UUID        REFERENCES users(id),
    actor_type      actor_type  NOT NULL DEFAULT 'system',
    action          TEXT        NOT NULL,
    resource_type   TEXT        NOT NULL,
    resource_id     UUID,
    changes         JSONB       NOT NULL DEFAULT '{}',
    reason          TEXT,
    ip_address      INET,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_tenant ON audit_logs(tenant_id, created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_actor ON audit_logs(actor_id, created_at DESC);
