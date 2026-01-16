-- Migration: 004_create_occurrences
-- Description: Create occurrences table with LGPD compliant data storage
-- Created: 2026-01-15

-- UP
CREATE TABLE IF NOT EXISTS occurrences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    obito_id UUID NOT NULL REFERENCES obitos_simulados(id) ON DELETE RESTRICT,
    hospital_id UUID NOT NULL REFERENCES hospitals(id) ON DELETE RESTRICT,
    status occurrence_status NOT NULL DEFAULT 'PENDENTE',
    score_priorizacao INTEGER NOT NULL DEFAULT 50,

    -- LGPD compliant fields
    nome_paciente_mascarado VARCHAR(255) NOT NULL,
    dados_completos JSONB NOT NULL,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notificado_em TIMESTAMP WITH TIME ZONE,

    -- Janela de captacao
    data_obito TIMESTAMP WITH TIME ZONE NOT NULL,
    janela_expira_em TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_occurrences_status ON occurrences(status);
CREATE INDEX IF NOT EXISTS idx_occurrences_hospital_id ON occurrences(hospital_id);
CREATE INDEX IF NOT EXISTS idx_occurrences_created_at ON occurrences(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_occurrences_janela_expira ON occurrences(janela_expira_em);
CREATE INDEX IF NOT EXISTS idx_occurrences_obito_id ON occurrences(obito_id);

-- Composite index for common listing queries
CREATE INDEX IF NOT EXISTS idx_occurrences_listing
    ON occurrences(status, hospital_id, created_at DESC);

-- Composite index for pending occurrences (dashboard view)
CREATE INDEX IF NOT EXISTS idx_occurrences_pendentes
    ON occurrences(score_priorizacao DESC, janela_expira_em ASC)
    WHERE status = 'PENDENTE';

-- Comments
COMMENT ON TABLE occurrences IS 'Ocorrencias de obitos elegiveis para doacao de corneas';
COMMENT ON COLUMN occurrences.nome_paciente_mascarado IS 'Nome do paciente com mascara LGPD (ex: Jo** Sil**)';
COMMENT ON COLUMN occurrences.dados_completos IS 'Dados completos do obito em JSON (acesso restrito)';
COMMENT ON COLUMN occurrences.score_priorizacao IS 'Score de priorizacao (UTI=100, Emergencia=80, Outros=50)';
COMMENT ON COLUMN occurrences.janela_expira_em IS 'Timestamp de expiracao da janela de 6 horas';

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS occurrences;
