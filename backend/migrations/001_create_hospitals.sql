-- Migration: 001_create_hospitals
-- Description: Create hospitals table with soft delete support
-- Created: 2026-01-15

-- UP
CREATE TABLE IF NOT EXISTS hospitals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(255) NOT NULL,
    codigo VARCHAR(50) NOT NULL,
    endereco TEXT,
    config_conexao JSONB DEFAULT '{}',
    ativo BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_hospitals_codigo ON hospitals(codigo) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_hospitals_ativo ON hospitals(ativo) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE hospitals IS 'Tabela de hospitais integrados ao VitalConnect';
COMMENT ON COLUMN hospitals.codigo IS 'Codigo unico do hospital (ex: HGG, HUGO)';
COMMENT ON COLUMN hospitals.config_conexao IS 'Configuracoes de integracao em formato JSON';
COMMENT ON COLUMN hospitals.deleted_at IS 'Soft delete - data de exclusao logica';

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS hospitals;
