-- Migration: 028_create_system_settings
-- Description: Create system_settings table for global platform configuration
-- Created: 2026-01-17
-- Purpose: Store global settings like SMTP, SMS, FCM configurations

-- UP

-- Create system_settings table (global table, no tenant_id)
CREATE TABLE IF NOT EXISTS system_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(100) NOT NULL,
    value JSONB NOT NULL,
    description TEXT,
    is_encrypted BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_system_settings_key UNIQUE (key)
);

-- Index on key for lookups
CREATE INDEX IF NOT EXISTS idx_system_settings_key ON system_settings(key);

-- Index on is_encrypted for filtering
CREATE INDEX IF NOT EXISTS idx_system_settings_is_encrypted ON system_settings(is_encrypted);

-- Comments
COMMENT ON TABLE system_settings IS 'Configuracoes globais do sistema - tabela global para super admin';
COMMENT ON COLUMN system_settings.id IS 'Identificador unico da configuracao (UUID)';
COMMENT ON COLUMN system_settings.key IS 'Chave unica da configuracao (ex: smtp_config, twilio_config, fcm_config)';
COMMENT ON COLUMN system_settings.value IS 'Valor da configuracao em formato JSONB';
COMMENT ON COLUMN system_settings.description IS 'Descricao da configuracao para documentacao';
COMMENT ON COLUMN system_settings.is_encrypted IS 'Indica se o valor esta criptografado (para dados sensiveis)';
COMMENT ON COLUMN system_settings.created_at IS 'Data de criacao do registro';
COMMENT ON COLUMN system_settings.updated_at IS 'Data da ultima atualizacao';

-- Seed initial settings with empty configurations
INSERT INTO system_settings (key, value, description, is_encrypted) VALUES
    ('smtp_config', '{"host": "", "port": 587, "user": "", "password": "", "from_address": "", "from_name": "VitalConnect"}', 'Configuracao do servidor SMTP para envio de emails', true),
    ('twilio_config', '{"account_sid": "", "auth_token": "", "from_number": ""}', 'Configuracao do Twilio para envio de SMS', true),
    ('fcm_config', '{"server_key": ""}', 'Configuracao do Firebase Cloud Messaging para push notifications', true)
ON CONFLICT (key) DO NOTHING;

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS system_settings;
