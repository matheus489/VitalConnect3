-- Migration: 005_create_occurrence_history
-- Description: Create occurrence history table for audit trail
-- Created: 2026-01-15

-- UP
CREATE TABLE IF NOT EXISTS occurrence_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    occurrence_id UUID NOT NULL REFERENCES occurrences(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    acao VARCHAR(100) NOT NULL,
    status_anterior occurrence_status,
    status_novo occurrence_status,
    observacoes TEXT,
    desfecho outcome_type,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_occurrence_history_occurrence_id ON occurrence_history(occurrence_id);
CREATE INDEX IF NOT EXISTS idx_occurrence_history_created_at ON occurrence_history(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_occurrence_history_user_id ON occurrence_history(user_id);

-- Composite index for history queries
CREATE INDEX IF NOT EXISTS idx_occurrence_history_listing
    ON occurrence_history(occurrence_id, created_at DESC);

-- Comments
COMMENT ON TABLE occurrence_history IS 'Historico de acoes realizadas em cada ocorrencia';
COMMENT ON COLUMN occurrence_history.acao IS 'Descricao da acao realizada';
COMMENT ON COLUMN occurrence_history.status_anterior IS 'Status antes da transicao (pode ser NULL para criacao)';
COMMENT ON COLUMN occurrence_history.status_novo IS 'Status apos a transicao';
COMMENT ON COLUMN occurrence_history.desfecho IS 'Tipo de desfecho ao concluir a ocorrencia';
COMMENT ON COLUMN occurrence_history.user_id IS 'Usuario que realizou a acao (NULL para acoes automaticas)';

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS occurrence_history;
