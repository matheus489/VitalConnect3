-- Migration: 008_create_audit_logs
-- Description: Create audit_logs table for comprehensive system auditing
-- Created: 2026-01-16
-- Requirements: Immutable table (WORM - Write Once, Read Many) for LGPD/CFM compliance

-- UP

-- Create severity enum if not exists
DO $$ BEGIN
    CREATE TYPE audit_severity AS ENUM ('INFO', 'WARN', 'CRITICAL');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create audit_logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    usuario_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_name VARCHAR(255) NOT NULL,
    acao VARCHAR(100) NOT NULL,
    entidade_tipo VARCHAR(100) NOT NULL,
    entidade_id VARCHAR(255) NOT NULL,
    hospital_id UUID REFERENCES hospitals(id) ON DELETE SET NULL,
    severity audit_severity NOT NULL DEFAULT 'INFO',
    detalhes JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT
);

-- Indexes for performance
-- Primary index on timestamp for default ordering (most recent first)
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC);

-- Individual column indexes for filtering
CREATE INDEX IF NOT EXISTS idx_audit_logs_usuario_id ON audit_logs(usuario_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entidade_tipo ON audit_logs(entidade_tipo);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entidade_id ON audit_logs(entidade_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_severity ON audit_logs(severity);
CREATE INDEX IF NOT EXISTS idx_audit_logs_hospital_id ON audit_logs(hospital_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_acao ON audit_logs(acao);

-- Composite index for Gestor queries (filter by hospital, ordered by time)
CREATE INDEX IF NOT EXISTS idx_audit_logs_hospital_timestamp
    ON audit_logs(hospital_id, timestamp DESC);

-- Composite index for entity timeline queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity_timeline
    ON audit_logs(entidade_tipo, entidade_id, timestamp ASC);

-- Table comments
COMMENT ON TABLE audit_logs IS 'Logs de auditoria do sistema - tabela imutavel para conformidade LGPD/CFM';
COMMENT ON COLUMN audit_logs.id IS 'Identificador unico do registro de auditoria';
COMMENT ON COLUMN audit_logs.timestamp IS 'Data/hora do evento';
COMMENT ON COLUMN audit_logs.usuario_id IS 'ID do usuario que realizou a acao (NULL para acoes do sistema)';
COMMENT ON COLUMN audit_logs.actor_name IS 'Nome do usuario ou "VitalConnect Bot" para acoes automaticas';
COMMENT ON COLUMN audit_logs.acao IS 'Tipo de acao realizada (ex: regra.update, auth.login)';
COMMENT ON COLUMN audit_logs.entidade_tipo IS 'Tipo de entidade afetada (ex: Ocorrencia, Regra)';
COMMENT ON COLUMN audit_logs.entidade_id IS 'ID da entidade afetada';
COMMENT ON COLUMN audit_logs.hospital_id IS 'Hospital relacionado a acao (para filtro por Gestor)';
COMMENT ON COLUMN audit_logs.severity IS 'Nivel de severidade: INFO, WARN, CRITICAL';
COMMENT ON COLUMN audit_logs.detalhes IS 'Dados contextuais em JSON (sem PII conforme LGPD)';
COMMENT ON COLUMN audit_logs.ip_address IS 'Endereco IP do cliente para rastreabilidade de seguranca';
COMMENT ON COLUMN audit_logs.user_agent IS 'User-Agent do cliente para rastreabilidade';

-- IMMUTABILITY CONSTRAINTS
-- Revoke UPDATE and DELETE permissions to ensure audit trail integrity
-- This simulates WORM (Write Once, Read Many) behavior

-- Create a trigger to prevent UPDATE operations
CREATE OR REPLACE FUNCTION prevent_audit_log_update()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'UPDATE operations are not allowed on audit_logs table';
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER audit_logs_no_update
    BEFORE UPDATE ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION prevent_audit_log_update();

-- Create a trigger to prevent DELETE operations
CREATE OR REPLACE FUNCTION prevent_audit_log_delete()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'DELETE operations are not allowed on audit_logs table';
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER audit_logs_no_delete
    BEFORE DELETE ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION prevent_audit_log_delete();

-- NOTE: For production with 5 years retention and high volume:
-- Consider implementing table partitioning by date range (monthly or quarterly)
-- Example (to be implemented when needed):
-- CREATE TABLE audit_logs (
--     ...
-- ) PARTITION BY RANGE (timestamp);
--
-- CREATE TABLE audit_logs_2026_q1 PARTITION OF audit_logs
--     FOR VALUES FROM ('2026-01-01') TO ('2026-04-01');

-- DOWN (for rollback - requires superuser to bypass triggers)
-- DROP TRIGGER IF EXISTS audit_logs_no_delete ON audit_logs;
-- DROP TRIGGER IF EXISTS audit_logs_no_update ON audit_logs;
-- DROP FUNCTION IF EXISTS prevent_audit_log_delete();
-- DROP FUNCTION IF EXISTS prevent_audit_log_update();
-- DROP TABLE IF EXISTS audit_logs;
-- DROP TYPE IF EXISTS audit_severity;
