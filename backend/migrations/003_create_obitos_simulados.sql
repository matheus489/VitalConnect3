-- Migration: 003_create_obitos_simulados
-- Description: Create simulated deaths table as data source for the listener
-- Created: 2026-01-15

-- UP
CREATE TABLE IF NOT EXISTS obitos_simulados (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hospital_id UUID NOT NULL REFERENCES hospitals(id) ON DELETE CASCADE,
    nome_paciente VARCHAR(255) NOT NULL,
    data_nascimento DATE NOT NULL,
    data_obito TIMESTAMP WITH TIME ZONE NOT NULL,
    causa_mortis VARCHAR(500) NOT NULL,
    prontuario VARCHAR(50),
    setor VARCHAR(100),
    leito VARCHAR(50),
    identificacao_desconhecida BOOLEAN NOT NULL DEFAULT false,
    processado BOOLEAN NOT NULL DEFAULT false,
    processado_em TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for listener polling
CREATE INDEX IF NOT EXISTS idx_obitos_simulados_data_obito ON obitos_simulados(data_obito);
CREATE INDEX IF NOT EXISTS idx_obitos_simulados_hospital_id ON obitos_simulados(hospital_id);
CREATE INDEX IF NOT EXISTS idx_obitos_simulados_processado ON obitos_simulados(processado);
CREATE INDEX IF NOT EXISTS idx_obitos_simulados_polling
    ON obitos_simulados(hospital_id, processado, created_at)
    WHERE processado = false;

-- Comments
COMMENT ON TABLE obitos_simulados IS 'Tabela simulada de obitos para demonstracao (substitui integracao real com PEP)';
COMMENT ON COLUMN obitos_simulados.identificacao_desconhecida IS 'Flag para pacientes sem identificacao (indigentes)';
COMMENT ON COLUMN obitos_simulados.processado IS 'Indica se o obito ja foi processado pelo listener';
COMMENT ON COLUMN obitos_simulados.processado_em IS 'Timestamp de quando foi processado pelo listener';
COMMENT ON COLUMN obitos_simulados.setor IS 'Setor do hospital (UTI, Emergencia, Enfermaria, etc)';

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS obitos_simulados;
