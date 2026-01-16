-- Migration: Create report_audit_logs table for LGPD compliance
-- Description: Tracks all report exports for audit purposes

-- Up Migration
CREATE TABLE IF NOT EXISTS report_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tipo_relatorio VARCHAR(10) NOT NULL CHECK (tipo_relatorio IN ('CSV', 'PDF')),
    filtros JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX idx_report_audit_logs_user_id ON report_audit_logs(user_id);
CREATE INDEX idx_report_audit_logs_created_at ON report_audit_logs(created_at DESC);
CREATE INDEX idx_report_audit_logs_tipo ON report_audit_logs(tipo_relatorio);

-- Comments for documentation
COMMENT ON TABLE report_audit_logs IS 'Audit log for report exports - LGPD compliance';
COMMENT ON COLUMN report_audit_logs.user_id IS 'User who performed the export';
COMMENT ON COLUMN report_audit_logs.tipo_relatorio IS 'Type of report exported (CSV or PDF)';
COMMENT ON COLUMN report_audit_logs.filtros IS 'Filters applied when generating the report (JSON)';
COMMENT ON COLUMN report_audit_logs.created_at IS 'Timestamp when the export was performed';

-- Down Migration (commented out for reference)
-- DROP TABLE IF EXISTS report_audit_logs;
