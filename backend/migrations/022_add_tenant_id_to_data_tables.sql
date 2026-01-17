-- Migration: 022_add_tenant_id_to_data_tables
-- Description: Add tenant_id column to occurrences, obitos_simulados, triagem_rules, shifts, notifications, audit_logs
-- Created: 2026-01-17
-- Note: tenant_id will be initially nullable, backfilled in migration 025, then made NOT NULL in 026

-- UP

-- =============================================================================
-- OCCURRENCES TABLE
-- =============================================================================
ALTER TABLE occurrences ADD COLUMN IF NOT EXISTS tenant_id UUID;

ALTER TABLE occurrences
ADD CONSTRAINT fk_occurrences_tenant_id
FOREIGN KEY (tenant_id) REFERENCES tenants(id)
ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_occurrences_tenant_id ON occurrences(tenant_id);

-- Composite index for tenant + status queries (most common filter)
CREATE INDEX IF NOT EXISTS idx_occurrences_tenant_status ON occurrences(tenant_id, status);

COMMENT ON COLUMN occurrences.tenant_id IS 'ID do tenant (Central de Transplantes) ao qual a ocorrencia pertence';

-- =============================================================================
-- OBITOS_SIMULADOS TABLE
-- =============================================================================
ALTER TABLE obitos_simulados ADD COLUMN IF NOT EXISTS tenant_id UUID;

ALTER TABLE obitos_simulados
ADD CONSTRAINT fk_obitos_simulados_tenant_id
FOREIGN KEY (tenant_id) REFERENCES tenants(id)
ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_obitos_simulados_tenant_id ON obitos_simulados(tenant_id);

-- Composite index for tenant + unprocessed obitos
CREATE INDEX IF NOT EXISTS idx_obitos_simulados_tenant_processado ON obitos_simulados(tenant_id, processado) WHERE processado = false;

COMMENT ON COLUMN obitos_simulados.tenant_id IS 'ID do tenant (Central de Transplantes) ao qual o obito pertence';

-- =============================================================================
-- TRIAGEM_RULES TABLE
-- =============================================================================
ALTER TABLE triagem_rules ADD COLUMN IF NOT EXISTS tenant_id UUID;

ALTER TABLE triagem_rules
ADD CONSTRAINT fk_triagem_rules_tenant_id
FOREIGN KEY (tenant_id) REFERENCES tenants(id)
ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_triagem_rules_tenant_id ON triagem_rules(tenant_id);

-- Composite index for tenant + active rules
CREATE INDEX IF NOT EXISTS idx_triagem_rules_tenant_ativo ON triagem_rules(tenant_id, ativo, prioridade DESC) WHERE ativo = true;

COMMENT ON COLUMN triagem_rules.tenant_id IS 'ID do tenant (Central de Transplantes) - cada tenant tem regras independentes';

-- =============================================================================
-- SHIFTS TABLE
-- =============================================================================
ALTER TABLE shifts ADD COLUMN IF NOT EXISTS tenant_id UUID;

ALTER TABLE shifts
ADD CONSTRAINT fk_shifts_tenant_id
FOREIGN KEY (tenant_id) REFERENCES tenants(id)
ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_shifts_tenant_id ON shifts(tenant_id);

-- Composite index for tenant + hospital shift queries
CREATE INDEX IF NOT EXISTS idx_shifts_tenant_hospital ON shifts(tenant_id, hospital_id);

COMMENT ON COLUMN shifts.tenant_id IS 'ID do tenant (Central de Transplantes) ao qual o plantao pertence';

-- =============================================================================
-- NOTIFICATIONS TABLE
-- =============================================================================
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS tenant_id UUID;

ALTER TABLE notifications
ADD CONSTRAINT fk_notifications_tenant_id
FOREIGN KEY (tenant_id) REFERENCES tenants(id)
ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_notifications_tenant_id ON notifications(tenant_id);

COMMENT ON COLUMN notifications.tenant_id IS 'ID do tenant (Central de Transplantes) ao qual a notificacao pertence';

-- =============================================================================
-- AUDIT_LOGS TABLE
-- Note: audit_logs has special handling - we need to temporarily disable the update trigger
-- =============================================================================

-- Temporarily disable the no-update trigger for schema changes
DROP TRIGGER IF EXISTS audit_logs_no_update ON audit_logs;

ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- Note: We don't add FK constraint on audit_logs to avoid issues with tenant deletion
-- audit_logs should preserve historical data even if tenant is removed (rare case)
-- ALTER TABLE audit_logs ADD CONSTRAINT fk_audit_logs_tenant_id FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_id ON audit_logs(tenant_id);

-- Composite index for tenant + timestamp queries (primary audit log query pattern)
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_timestamp ON audit_logs(tenant_id, timestamp DESC);

COMMENT ON COLUMN audit_logs.tenant_id IS 'ID do tenant (Central de Transplantes) para filtro de auditoria por tenant';

-- Re-enable the no-update trigger
CREATE TRIGGER audit_logs_no_update
    BEFORE UPDATE ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION prevent_audit_log_update();

-- DOWN (for rollback)
-- DROP TRIGGER IF EXISTS audit_logs_no_update ON audit_logs;
--
-- ALTER TABLE audit_logs DROP COLUMN IF EXISTS tenant_id;
-- DROP INDEX IF EXISTS idx_audit_logs_tenant_timestamp;
-- DROP INDEX IF EXISTS idx_audit_logs_tenant_id;
--
-- ALTER TABLE notifications DROP CONSTRAINT IF EXISTS fk_notifications_tenant_id;
-- ALTER TABLE notifications DROP COLUMN IF EXISTS tenant_id;
-- DROP INDEX IF EXISTS idx_notifications_tenant_id;
--
-- ALTER TABLE shifts DROP CONSTRAINT IF EXISTS fk_shifts_tenant_id;
-- ALTER TABLE shifts DROP COLUMN IF EXISTS tenant_id;
-- DROP INDEX IF EXISTS idx_shifts_tenant_hospital;
-- DROP INDEX IF EXISTS idx_shifts_tenant_id;
--
-- ALTER TABLE triagem_rules DROP CONSTRAINT IF EXISTS fk_triagem_rules_tenant_id;
-- ALTER TABLE triagem_rules DROP COLUMN IF EXISTS tenant_id;
-- DROP INDEX IF EXISTS idx_triagem_rules_tenant_ativo;
-- DROP INDEX IF EXISTS idx_triagem_rules_tenant_id;
--
-- ALTER TABLE obitos_simulados DROP CONSTRAINT IF EXISTS fk_obitos_simulados_tenant_id;
-- ALTER TABLE obitos_simulados DROP COLUMN IF EXISTS tenant_id;
-- DROP INDEX IF EXISTS idx_obitos_simulados_tenant_processado;
-- DROP INDEX IF EXISTS idx_obitos_simulados_tenant_id;
--
-- ALTER TABLE occurrences DROP CONSTRAINT IF EXISTS fk_occurrences_tenant_id;
-- ALTER TABLE occurrences DROP COLUMN IF EXISTS tenant_id;
-- DROP INDEX IF EXISTS idx_occurrences_tenant_status;
-- DROP INDEX IF EXISTS idx_occurrences_tenant_id;
--
-- CREATE TRIGGER audit_logs_no_update
--     BEFORE UPDATE ON audit_logs
--     FOR EACH ROW
--     EXECUTE FUNCTION prevent_audit_log_update();
