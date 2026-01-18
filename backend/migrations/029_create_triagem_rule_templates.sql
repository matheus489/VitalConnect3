-- Migration: 029_create_triagem_rule_templates
-- Description: Create triagem_rule_templates table for master rule templates
-- Created: 2026-01-17
-- Purpose: Allow super admin to create master triagem rules that can be cloned to tenants

-- UP

-- Create triagem_rule_templates table (global table, no tenant_id)
CREATE TABLE IF NOT EXISTS triagem_rule_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nome VARCHAR(255) NOT NULL,
    tipo VARCHAR(50) NOT NULL,
    condicao JSONB NOT NULL,
    descricao TEXT,
    ativo BOOLEAN DEFAULT true,
    prioridade INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index on tipo for filtering by rule type
CREATE INDEX IF NOT EXISTS idx_triagem_rule_templates_tipo ON triagem_rule_templates(tipo);

-- Index on ativo for filtering active/inactive templates
CREATE INDEX IF NOT EXISTS idx_triagem_rule_templates_ativo ON triagem_rule_templates(ativo);

-- Index on nome for searching
CREATE INDEX IF NOT EXISTS idx_triagem_rule_templates_nome ON triagem_rule_templates(nome);

-- GIN index on condicao for JSONB queries
CREATE INDEX IF NOT EXISTS idx_triagem_rule_templates_condicao ON triagem_rule_templates USING GIN (condicao);

-- Combined index for common queries
CREATE INDEX IF NOT EXISTS idx_triagem_rule_templates_ativo_tipo ON triagem_rule_templates(ativo, tipo);

-- Comments
COMMENT ON TABLE triagem_rule_templates IS 'Templates de regras de triagem - modelos globais que podem ser clonados para tenants';
COMMENT ON COLUMN triagem_rule_templates.id IS 'Identificador unico do template (UUID)';
COMMENT ON COLUMN triagem_rule_templates.nome IS 'Nome descritivo da regra template';
COMMENT ON COLUMN triagem_rule_templates.tipo IS 'Tipo da regra (idade_maxima, causas_excludentes, janela_horas, etc)';
COMMENT ON COLUMN triagem_rule_templates.condicao IS 'Condicao/valor da regra em formato JSONB';
COMMENT ON COLUMN triagem_rule_templates.descricao IS 'Descricao detalhada do que a regra faz';
COMMENT ON COLUMN triagem_rule_templates.ativo IS 'Indica se o template esta ativo para uso';
COMMENT ON COLUMN triagem_rule_templates.prioridade IS 'Ordem de execucao da regra (menor = maior prioridade)';
COMMENT ON COLUMN triagem_rule_templates.created_at IS 'Data de criacao do registro';
COMMENT ON COLUMN triagem_rule_templates.updated_at IS 'Data da ultima atualizacao';

-- Seed some default templates
INSERT INTO triagem_rule_templates (nome, tipo, condicao, descricao, ativo, prioridade) VALUES
    (
        'Idade Maxima Padrao',
        'idade_maxima',
        '{"valor": 75, "acao": "rejeitar"}',
        'Rejeita doadores com idade superior a 75 anos',
        true,
        1
    ),
    (
        'Janela de Tempo 6 Horas',
        'janela_horas',
        '{"valor": 6, "acao": "rejeitar"}',
        'Rejeita notificacoes com mais de 6 horas desde o obito',
        true,
        2
    ),
    (
        'Identificacao Desconhecida',
        'identificacao_desconhecida',
        '{"valor": true, "acao": "rejeitar"}',
        'Rejeita casos onde a identificacao do paciente e desconhecida',
        true,
        3
    )
ON CONFLICT DO NOTHING;

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS triagem_rule_templates;
