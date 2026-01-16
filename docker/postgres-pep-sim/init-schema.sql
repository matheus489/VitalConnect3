-- VitalConnect PEP Simulator - Schema Initialization
-- Simulates Tasy (Philips) hospital system database schema
-- This is a read-only simulation for development and demo purposes

-- Create schema to simulate Tasy namespace
CREATE SCHEMA IF NOT EXISTS TASY;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Main patient death registry table (simulating Tasy.TB_PACIENTE_OBITO)
CREATE TABLE IF NOT EXISTS TASY.TB_PACIENTE_OBITO (
    -- Primary key - Internal sequential ID
    CD_PACIENTE_OBITO SERIAL PRIMARY KEY,

    -- Patient identification
    CD_PACIENTE BIGINT NOT NULL,
    NM_PACIENTE VARCHAR(255) NOT NULL,
    DT_NASCIMENTO DATE NOT NULL,
    NR_CNS VARCHAR(15),  -- Cartao Nacional de Saude (15 digits)
    NR_CPF VARCHAR(14),  -- CPF with formatting (xxx.xxx.xxx-xx)

    -- Death information
    DT_OBITO TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    DS_CAUSA_MORTIS VARCHAR(500) NOT NULL,
    CD_CID VARCHAR(10),  -- ICD-10 code

    -- Location information
    CD_SETOR VARCHAR(50),
    NR_LEITO VARCHAR(20),
    NR_PRONTUARIO VARCHAR(50),

    -- Metadata
    IE_IDENTIFICACAO_DESCONHECIDA CHAR(1) DEFAULT 'N',
    DT_ATUALIZACAO TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    NM_USUARIO_ATUALIZACAO VARCHAR(100) DEFAULT 'SISTEMA',

    -- Index for efficient queries
    CONSTRAINT chk_identificacao CHECK (IE_IDENTIFICACAO_DESCONHECIDA IN ('S', 'N'))
);

-- Indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_obito_dt_obito ON TASY.TB_PACIENTE_OBITO(DT_OBITO DESC);
CREATE INDEX IF NOT EXISTS idx_obito_cd_paciente ON TASY.TB_PACIENTE_OBITO(CD_PACIENTE);
CREATE INDEX IF NOT EXISTS idx_obito_nr_cns ON TASY.TB_PACIENTE_OBITO(NR_CNS);
CREATE INDEX IF NOT EXISTS idx_obito_dt_atualizacao ON TASY.TB_PACIENTE_OBITO(DT_ATUALIZACAO DESC);

-- View for simplified access (matches VitalConnect field expectations)
CREATE OR REPLACE VIEW TASY.VW_OBITOS_RECENTES AS
SELECT
    CD_PACIENTE_OBITO,
    CD_PACIENTE,
    NM_PACIENTE,
    DT_NASCIMENTO,
    NR_CNS,
    NR_CPF,
    DT_OBITO,
    DS_CAUSA_MORTIS,
    CD_CID,
    CD_SETOR,
    NR_LEITO,
    NR_PRONTUARIO,
    IE_IDENTIFICACAO_DESCONHECIDA,
    DT_ATUALIZACAO
FROM TASY.TB_PACIENTE_OBITO
WHERE DT_OBITO >= CURRENT_TIMESTAMP - INTERVAL '24 hours'
ORDER BY DT_OBITO DESC;

-- Grant read-only permissions (simulate restricted PEP access)
-- In production, the agent user would only have SELECT permissions
GRANT USAGE ON SCHEMA TASY TO postgres;
GRANT SELECT ON ALL TABLES IN SCHEMA TASY TO postgres;

-- Comment on objects
COMMENT ON SCHEMA TASY IS 'Simulated Tasy (Philips) hospital PEP schema';
COMMENT ON TABLE TASY.TB_PACIENTE_OBITO IS 'Patient death registry - simulates real Tasy structure';
COMMENT ON VIEW TASY.VW_OBITOS_RECENTES IS 'Recent deaths view for agent polling (last 24h)';
