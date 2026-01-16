-- Migration: 006_create_triagem_rules
-- Description: Create triagem rules table with JSONB for flexible rule configuration
-- Created: 2026-01-15

-- UP
CREATE TABLE IF NOT EXISTS triagem_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(255) NOT NULL,
    descricao TEXT,
    regras JSONB NOT NULL DEFAULT '{}',
    ativo BOOLEAN NOT NULL DEFAULT true,
    prioridade INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- GIN index for JSONB queries
CREATE INDEX IF NOT EXISTS idx_triagem_rules_regras ON triagem_rules USING GIN (regras);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_triagem_rules_ativo ON triagem_rules(ativo);
CREATE INDEX IF NOT EXISTS idx_triagem_rules_prioridade ON triagem_rules(prioridade DESC);

-- Index for active rules ordered by priority
CREATE INDEX IF NOT EXISTS idx_triagem_rules_active_priority
    ON triagem_rules(prioridade DESC)
    WHERE ativo = true;

-- Comments
COMMENT ON TABLE triagem_rules IS 'Regras configuraveis de triagem para elegibilidade de doacao';
COMMENT ON COLUMN triagem_rules.regras IS 'Regras em formato JSONB (ex: {"idade_maxima": 80, "causas_excludentes": [...], "setores_prioritarios": [...]})';
COMMENT ON COLUMN triagem_rules.prioridade IS 'Ordem de aplicacao das regras (maior = primeiro)';

/*
Exemplo de estrutura do campo regras:
{
    "tipo": "idade_maxima",
    "valor": 80,
    "acao": "rejeitar"
}

{
    "tipo": "causas_excludentes",
    "valor": ["Neoplasia maligna", "Sepse", "HIV/AIDS"],
    "acao": "rejeitar"
}

{
    "tipo": "janela_horas",
    "valor": 6,
    "acao": "rejeitar"
}

{
    "tipo": "identificacao_desconhecida",
    "valor": true,
    "acao": "rejeitar"
}

{
    "tipo": "setor_priorizacao",
    "valor": {"UTI": 100, "Emergencia": 80, "Enfermaria": 50},
    "acao": "priorizar"
}
*/

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS triagem_rules;
