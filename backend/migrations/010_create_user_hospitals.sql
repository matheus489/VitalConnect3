-- Migration: 009_create_user_hospitals
-- Description: Create user_hospitals junction table for N:N relationship
-- Created: 2026-01-16

-- UP
CREATE TABLE IF NOT EXISTS user_hospitals (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    hospital_id UUID NOT NULL REFERENCES hospitals(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, hospital_id)
);

-- Indexes for foreign keys and queries
CREATE INDEX IF NOT EXISTS idx_user_hospitals_user_id ON user_hospitals(user_id);
CREATE INDEX IF NOT EXISTS idx_user_hospitals_hospital_id ON user_hospitals(hospital_id);

-- Comments
COMMENT ON TABLE user_hospitals IS 'Tabela de juncao para relacao N:N entre usuarios e hospitais';
COMMENT ON COLUMN user_hospitals.user_id IS 'ID do usuario vinculado';
COMMENT ON COLUMN user_hospitals.hospital_id IS 'ID do hospital vinculado';
COMMENT ON COLUMN user_hospitals.created_at IS 'Data de criacao do vinculo';

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS user_hospitals;
