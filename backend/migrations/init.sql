-- SIDOT Database Initialization
-- This file runs on first container startup

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create custom types for enums
DO $$
BEGIN
    -- User roles
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
        CREATE TYPE user_role AS ENUM ('operador', 'gestor', 'admin');
    END IF;

    -- Occurrence status
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'occurrence_status') THEN
        CREATE TYPE occurrence_status AS ENUM (
            'PENDENTE',
            'EM_ANDAMENTO',
            'ACEITA',
            'RECUSADA',
            'CANCELADA',
            'CONCLUIDA'
        );
    END IF;

    -- Notification channel
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'notification_channel') THEN
        CREATE TYPE notification_channel AS ENUM ('dashboard', 'email');
    END IF;

    -- Outcome type
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'outcome_type') THEN
        CREATE TYPE outcome_type AS ENUM (
            'sucesso_captacao',
            'familia_recusou',
            'contraindicacao_medica',
            'tempo_excedido',
            'outro'
        );
    END IF;
END$$;

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO postgres;
