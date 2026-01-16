-- Migration: 002_create_users
-- Description: Create users table with role enum and hospital relationship
-- Created: 2026-01-15

-- UP
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nome VARCHAR(255) NOT NULL,
    role user_role NOT NULL DEFAULT 'operador',
    hospital_id UUID REFERENCES hospitals(id) ON DELETE SET NULL,
    ativo BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_hospital_id ON users(hospital_id);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_ativo ON users(ativo);

-- Comments
COMMENT ON TABLE users IS 'Usuarios do sistema VitalConnect';
COMMENT ON COLUMN users.role IS 'Perfil do usuario: operador, gestor ou admin';
COMMENT ON COLUMN users.hospital_id IS 'Hospital vinculado (opcional para admin)';
COMMENT ON COLUMN users.password_hash IS 'Hash bcrypt da senha (cost factor 12)';

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS users;
