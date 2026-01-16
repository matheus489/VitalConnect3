-- Migration: 007_create_notifications
-- Description: Create notifications table for tracking sent notifications
-- Created: 2026-01-15

-- UP
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    occurrence_id UUID NOT NULL REFERENCES occurrences(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    canal notification_channel NOT NULL,
    enviado_em TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status_envio VARCHAR(50) NOT NULL DEFAULT 'enviado',
    erro_mensagem TEXT,
    metadata JSONB DEFAULT '{}'
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_notifications_occurrence_id ON notifications(occurrence_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_canal ON notifications(canal);
CREATE INDEX IF NOT EXISTS idx_notifications_enviado_em ON notifications(enviado_em DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_status_envio ON notifications(status_envio);

-- Composite index for metrics calculation
CREATE INDEX IF NOT EXISTS idx_notifications_metrics
    ON notifications(enviado_em, status_envio)
    WHERE status_envio = 'enviado';

-- Comments
COMMENT ON TABLE notifications IS 'Registro de notificacoes enviadas para auditoria e metricas';
COMMENT ON COLUMN notifications.canal IS 'Canal de notificacao: dashboard ou email';
COMMENT ON COLUMN notifications.status_envio IS 'Status do envio: enviado, falha, pendente';
COMMENT ON COLUMN notifications.erro_mensagem IS 'Mensagem de erro em caso de falha no envio';
COMMENT ON COLUMN notifications.metadata IS 'Dados adicionais da notificacao em JSON';

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS notifications;
