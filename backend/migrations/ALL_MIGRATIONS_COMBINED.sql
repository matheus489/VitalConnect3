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
COMMENT ON TABLE hospitals IS 'Tabela de hospitais integrados ao SIDOT';
COMMENT ON COLUMN hospitals.codigo IS 'Codigo unico do hospital (ex: HGG, HUGO)';
COMMENT ON COLUMN hospitals.config_conexao IS 'Configuracoes de integracao em formato JSON';
COMMENT ON COLUMN hospitals.deleted_at IS 'Soft delete - data de exclusao logica';

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS hospitals;
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
COMMENT ON TABLE users IS 'Usuarios do sistema SIDOT';
COMMENT ON COLUMN users.role IS 'Perfil do usuario: operador, gestor ou admin';
COMMENT ON COLUMN users.hospital_id IS 'Hospital vinculado (opcional para admin)';
COMMENT ON COLUMN users.password_hash IS 'Hash bcrypt da senha (cost factor 12)';

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS users;
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
-- Migration: 005_create_occurrence_history
-- Description: Create occurrence history table for audit trail
-- Created: 2026-01-15

-- UP
CREATE TABLE IF NOT EXISTS occurrence_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    occurrence_id UUID NOT NULL REFERENCES occurrences(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    acao VARCHAR(100) NOT NULL,
    status_anterior occurrence_status,
    status_novo occurrence_status,
    observacoes TEXT,
    desfecho outcome_type,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_occurrence_history_occurrence_id ON occurrence_history(occurrence_id);
CREATE INDEX IF NOT EXISTS idx_occurrence_history_created_at ON occurrence_history(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_occurrence_history_user_id ON occurrence_history(user_id);

-- Composite index for history queries
CREATE INDEX IF NOT EXISTS idx_occurrence_history_listing
    ON occurrence_history(occurrence_id, created_at DESC);

-- Comments
COMMENT ON TABLE occurrence_history IS 'Historico de acoes realizadas em cada ocorrencia';
COMMENT ON COLUMN occurrence_history.acao IS 'Descricao da acao realizada';
COMMENT ON COLUMN occurrence_history.status_anterior IS 'Status antes da transicao (pode ser NULL para criacao)';
COMMENT ON COLUMN occurrence_history.status_novo IS 'Status apos a transicao';
COMMENT ON COLUMN occurrence_history.desfecho IS 'Tipo de desfecho ao concluir a ocorrencia';
COMMENT ON COLUMN occurrence_history.user_id IS 'Usuario que realizou a acao (NULL para acoes automaticas)';

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS occurrence_history;
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
-- Migration: 007_create_notifications
-- Description: Create notifications table for tracking sent notifications
-- Created: 2026-01-15

-- UP
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    occurrence_id UUID NOT NULL REFERENCES occurrences(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    canal notification_channel NOT NULL,
    enviado_em TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status_envio VARCHAR(50) NOT NULL DEFAULT 'enviado',
    erro_mensagem TEXT,
    metadata JSONB DEFAULT '{}'
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_notifications_occurrence_id ON notifications(occurrence_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_canal ON notifications(canal);
CREATE INDEX IF NOT EXISTS idx_notifications_enviado_em ON notifications(enviado_em DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_status_envio ON notifications(status_envio);

-- Composite index for metrics calculation
CREATE INDEX IF NOT EXISTS idx_notifications_metrics
    ON notifications(enviado_em, status_envio)
    WHERE status_envio = 'enviado';

-- Comments
COMMENT ON TABLE notifications IS 'Registro de notificacoes enviadas para auditoria e metricas';
COMMENT ON COLUMN notifications.canal IS 'Canal de notificacao: dashboard ou email';
COMMENT ON COLUMN notifications.status_envio IS 'Status do envio: enviado, falha, pendente';
COMMENT ON COLUMN notifications.erro_mensagem IS 'Mensagem de erro em caso de falha no envio';
COMMENT ON COLUMN notifications.metadata IS 'Dados adicionais da notificacao em JSON';

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS notifications;
-- Migration: 008_add_mobile_phone_to_users
-- Description: Add mobile_phone field to users table for SMS notifications
-- Created: 2026-01-16

-- UP
ALTER TABLE users ADD COLUMN IF NOT EXISTS mobile_phone VARCHAR(16);

-- Add comment for documentation
COMMENT ON COLUMN users.mobile_phone IS 'Mobile phone number in E.164 format (e.g., +5511999999999) for SMS notifications';

-- DOWN (for rollback)
-- ALTER TABLE users DROP COLUMN IF EXISTS mobile_phone;
-- Migration: 008_add_email_notifications
-- Description: Add email_notifications field to users table
-- Created: 2026-01-16

-- UP
ALTER TABLE users
ADD COLUMN IF NOT EXISTS email_notifications BOOLEAN NOT NULL DEFAULT true;

-- Index for filtering by email notification preference
CREATE INDEX IF NOT EXISTS idx_users_email_notifications ON users(email_notifications);

-- Comments
COMMENT ON COLUMN users.email_notifications IS 'Preferencia de notificacao por email (true = ativo, dashboard sempre ativo)';

-- DOWN (for rollback)
-- ALTER TABLE users DROP COLUMN IF EXISTS email_notifications;
-- DROP INDEX IF EXISTS idx_users_email_notifications;
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
-- Migration: 010_migrate_hospital_id_to_user_hospitals
-- Description: Migrate existing hospital_id data to user_hospitals table and remove column
-- Created: 2026-01-16

-- UP

-- Step 1: Migrate existing data from users.hospital_id to user_hospitals
INSERT INTO user_hospitals (user_id, hospital_id, created_at)
SELECT id, hospital_id, CURRENT_TIMESTAMP
FROM users
WHERE hospital_id IS NOT NULL
ON CONFLICT (user_id, hospital_id) DO NOTHING;

-- Step 2: Drop the hospital_id foreign key constraint first
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_hospital_id_fkey;

-- Step 3: Drop the index on hospital_id
DROP INDEX IF EXISTS idx_users_hospital_id;

-- Step 4: Remove the hospital_id column
ALTER TABLE users DROP COLUMN IF EXISTS hospital_id;

-- Comments
COMMENT ON TABLE users IS 'Usuarios do sistema SIDOT - relacao N:N com hospitais via user_hospitals';

-- DOWN (for rollback - complex, requires recreating column)
-- ALTER TABLE users ADD COLUMN hospital_id UUID REFERENCES hospitals(id) ON DELETE SET NULL;
-- CREATE INDEX IF NOT EXISTS idx_users_hospital_id ON users(hospital_id);
-- UPDATE users u SET hospital_id = (SELECT hospital_id FROM user_hospitals uh WHERE uh.user_id = u.id LIMIT 1);
-- Migration: 009_create_user_notification_preferences
-- Description: Create user_notification_preferences table for SMS/Email preferences
-- Created: 2026-01-16

-- UP
CREATE TABLE IF NOT EXISTS user_notification_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sms_enabled BOOLEAN NOT NULL DEFAULT true,
    email_enabled BOOLEAN NOT NULL DEFAULT true,
    dashboard_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_notification_preferences_user_id ON user_notification_preferences(user_id);

-- Comments
COMMENT ON TABLE user_notification_preferences IS 'User notification preferences for SMS, email and dashboard alerts';
COMMENT ON COLUMN user_notification_preferences.sms_enabled IS 'Enable SMS notifications (default: true if mobile_phone present)';
COMMENT ON COLUMN user_notification_preferences.email_enabled IS 'Enable email notifications (default: true)';
COMMENT ON COLUMN user_notification_preferences.dashboard_enabled IS 'Enable dashboard notifications (always true, not editable)';

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS user_notification_preferences;
-- Migration: 008_create_shifts
-- Description: Create shifts table for operator scheduling and notification routing
-- Created: 2026-01-16

-- UP
CREATE TABLE IF NOT EXISTS shifts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hospital_id UUID NOT NULL REFERENCES hospitals(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performant queries
CREATE INDEX IF NOT EXISTS idx_shifts_hospital_id ON shifts(hospital_id);
CREATE INDEX IF NOT EXISTS idx_shifts_user_id ON shifts(user_id);
CREATE INDEX IF NOT EXISTS idx_shifts_day_of_week ON shifts(day_of_week);
CREATE INDEX IF NOT EXISTS idx_shifts_hospital_day ON shifts(hospital_id, day_of_week);

-- Unique constraint to prevent duplicate shifts
CREATE UNIQUE INDEX IF NOT EXISTS idx_shifts_unique_schedule
    ON shifts(hospital_id, user_id, day_of_week, start_time);

-- Comments
COMMENT ON TABLE shifts IS 'Escalas de plantao dos operadores por hospital';
COMMENT ON COLUMN shifts.hospital_id IS 'Hospital ao qual esta escala pertence';
COMMENT ON COLUMN shifts.user_id IS 'Operador escalado para este turno';
COMMENT ON COLUMN shifts.day_of_week IS 'Dia da semana (0=Domingo, 1=Segunda, ..., 6=Sabado)';
COMMENT ON COLUMN shifts.start_time IS 'Horario de inicio do turno (ex: 07:00, 19:00)';
COMMENT ON COLUMN shifts.end_time IS 'Horario de fim do turno (ex: 19:00, 07:00 para turno noturno)';

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS shifts;
-- Migration: 008_create_audit_logs
-- Description: Create audit_logs table for comprehensive system auditing
-- Created: 2026-01-16
-- Requirements: Immutable table (WORM - Write Once, Read Many) for LGPD/CFM compliance

-- UP

-- Create severity enum if not exists
DO $$ BEGIN
    CREATE TYPE audit_severity AS ENUM ('INFO', 'WARN', 'CRITICAL');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create audit_logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    usuario_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_name VARCHAR(255) NOT NULL,
    acao VARCHAR(100) NOT NULL,
    entidade_tipo VARCHAR(100) NOT NULL,
    entidade_id VARCHAR(255) NOT NULL,
    hospital_id UUID REFERENCES hospitals(id) ON DELETE SET NULL,
    severity audit_severity NOT NULL DEFAULT 'INFO',
    detalhes JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT
);

-- Indexes for performance
-- Primary index on timestamp for default ordering (most recent first)
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC);

-- Individual column indexes for filtering
CREATE INDEX IF NOT EXISTS idx_audit_logs_usuario_id ON audit_logs(usuario_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entidade_tipo ON audit_logs(entidade_tipo);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entidade_id ON audit_logs(entidade_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_severity ON audit_logs(severity);
CREATE INDEX IF NOT EXISTS idx_audit_logs_hospital_id ON audit_logs(hospital_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_acao ON audit_logs(acao);

-- Composite index for Gestor queries (filter by hospital, ordered by time)
CREATE INDEX IF NOT EXISTS idx_audit_logs_hospital_timestamp
    ON audit_logs(hospital_id, timestamp DESC);

-- Composite index for entity timeline queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity_timeline
    ON audit_logs(entidade_tipo, entidade_id, timestamp ASC);

-- Table comments
COMMENT ON TABLE audit_logs IS 'Logs de auditoria do sistema - tabela imutavel para conformidade LGPD/CFM';
COMMENT ON COLUMN audit_logs.id IS 'Identificador unico do registro de auditoria';
COMMENT ON COLUMN audit_logs.timestamp IS 'Data/hora do evento';
COMMENT ON COLUMN audit_logs.usuario_id IS 'ID do usuario que realizou a acao (NULL para acoes do sistema)';
COMMENT ON COLUMN audit_logs.actor_name IS 'Nome do usuario ou "SIDOT Bot" para acoes automaticas';
COMMENT ON COLUMN audit_logs.acao IS 'Tipo de acao realizada (ex: regra.update, auth.login)';
COMMENT ON COLUMN audit_logs.entidade_tipo IS 'Tipo de entidade afetada (ex: Ocorrencia, Regra)';
COMMENT ON COLUMN audit_logs.entidade_id IS 'ID da entidade afetada';
COMMENT ON COLUMN audit_logs.hospital_id IS 'Hospital relacionado a acao (para filtro por Gestor)';
COMMENT ON COLUMN audit_logs.severity IS 'Nivel de severidade: INFO, WARN, CRITICAL';
COMMENT ON COLUMN audit_logs.detalhes IS 'Dados contextuais em JSON (sem PII conforme LGPD)';
COMMENT ON COLUMN audit_logs.ip_address IS 'Endereco IP do cliente para rastreabilidade de seguranca';
COMMENT ON COLUMN audit_logs.user_agent IS 'User-Agent do cliente para rastreabilidade';

-- IMMUTABILITY CONSTRAINTS
-- Revoke UPDATE and DELETE permissions to ensure audit trail integrity
-- This simulates WORM (Write Once, Read Many) behavior

-- Create a trigger to prevent UPDATE operations
CREATE OR REPLACE FUNCTION prevent_audit_log_update()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'UPDATE operations are not allowed on audit_logs table';
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER audit_logs_no_update
    BEFORE UPDATE ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION prevent_audit_log_update();

-- Create a trigger to prevent DELETE operations
CREATE OR REPLACE FUNCTION prevent_audit_log_delete()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'DELETE operations are not allowed on audit_logs table';
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER audit_logs_no_delete
    BEFORE DELETE ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION prevent_audit_log_delete();

-- NOTE: For production with 5 years retention and high volume:
-- Consider implementing table partitioning by date range (monthly or quarterly)
-- Example (to be implemented when needed):
-- CREATE TABLE audit_logs (
--     ...
-- ) PARTITION BY RANGE (timestamp);
--
-- CREATE TABLE audit_logs_2026_q1 PARTITION OF audit_logs
--     FOR VALUES FROM ('2026-01-01') TO ('2026-04-01');

-- DOWN (for rollback - requires superuser to bypass triggers)
-- DROP TRIGGER IF EXISTS audit_logs_no_delete ON audit_logs;
-- DROP TRIGGER IF EXISTS audit_logs_no_update ON audit_logs;
-- DROP FUNCTION IF EXISTS prevent_audit_log_delete();
-- DROP FUNCTION IF EXISTS prevent_audit_log_update();
-- DROP TABLE IF EXISTS audit_logs;
-- DROP TYPE IF EXISTS audit_severity;
-- Migration: Create report_audit_logs table for LGPD compliance
-- Description: Tracks all report exports for audit purposes

-- Up Migration
CREATE TABLE IF NOT EXISTS report_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tipo_relatorio VARCHAR(10) NOT NULL CHECK (tipo_relatorio IN ('CSV', 'PDF')),
    filtros JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX idx_report_audit_logs_user_id ON report_audit_logs(user_id);
CREATE INDEX idx_report_audit_logs_created_at ON report_audit_logs(created_at DESC);
CREATE INDEX idx_report_audit_logs_tipo ON report_audit_logs(tipo_relatorio);

-- Comments for documentation
COMMENT ON TABLE report_audit_logs IS 'Audit log for report exports - LGPD compliance';
COMMENT ON COLUMN report_audit_logs.user_id IS 'User who performed the export';
COMMENT ON COLUMN report_audit_logs.tipo_relatorio IS 'Type of report exported (CSV or PDF)';
COMMENT ON COLUMN report_audit_logs.filtros IS 'Filters applied when generating the report (JSON)';
COMMENT ON COLUMN report_audit_logs.created_at IS 'Timestamp when the export was performed';

-- Down Migration (commented out for reference)
-- DROP TABLE IF EXISTS report_audit_logs;
-- Migration: 016_create_push_subscriptions
-- Description: Create push_subscriptions table for FCM tokens
-- Created: 2026-01-16

-- UP
CREATE TABLE IF NOT EXISTS push_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    platform VARCHAR(20) NOT NULL DEFAULT 'web',
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_push_subscriptions_token ON push_subscriptions(token);
CREATE INDEX IF NOT EXISTS idx_push_subscriptions_user_id ON push_subscriptions(user_id);

-- Comments
COMMENT ON TABLE push_subscriptions IS 'FCM push notification subscriptions for users';
COMMENT ON COLUMN push_subscriptions.token IS 'FCM registration token from client';
COMMENT ON COLUMN push_subscriptions.platform IS 'Platform: web, android, ios';
COMMENT ON COLUMN push_subscriptions.user_agent IS 'Browser/device user agent for debugging';

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS push_subscriptions;
-- Migration: 017_add_coordinates_to_hospitals
-- Description: Add latitude and longitude columns for geographic map feature
-- Created: 2026-01-16

-- UP
-- Adicionar campos de coordenadas geograficas para o Dashboard Geografico
ALTER TABLE hospitals ADD COLUMN IF NOT EXISTS latitude DECIMAL(10, 8);
ALTER TABLE hospitals ADD COLUMN IF NOT EXISTS longitude DECIMAL(11, 8);

-- Indice para buscas geograficas (hospitais ativos com coordenadas)
CREATE INDEX IF NOT EXISTS idx_hospitals_coordinates
    ON hospitals(latitude, longitude)
    WHERE deleted_at IS NULL AND ativo = true AND latitude IS NOT NULL AND longitude IS NOT NULL;

-- Comentarios descritivos
COMMENT ON COLUMN hospitals.latitude IS 'Latitude geografica do hospital (formato decimal, ex: -16.6868)';
COMMENT ON COLUMN hospitals.longitude IS 'Longitude geografica do hospital (formato decimal, ex: -49.2648)';

-- DOWN (for rollback)
-- DROP INDEX IF EXISTS idx_hospitals_coordinates;
-- ALTER TABLE hospitals DROP COLUMN IF EXISTS longitude;
-- ALTER TABLE hospitals DROP COLUMN IF EXISTS latitude;
-- Migration: 018_add_telefone_to_hospitals
-- Description: Add telefone column for hospital contact information
-- Created: 2026-01-17

-- UP
-- Adicionar campo de telefone para contato do hospital (recepcao/plantao)
ALTER TABLE hospitals ADD COLUMN IF NOT EXISTS telefone VARCHAR(20);

-- Comentario descritivo
COMMENT ON COLUMN hospitals.telefone IS 'Telefone de contato do hospital - formato brasileiro: (XX) XXXX-XXXX ou (XX) XXXXX-XXXX';

-- DOWN (for rollback)
-- ALTER TABLE hospitals DROP COLUMN IF EXISTS telefone;
-- Migration: 019_create_tenants
-- Description: Create tenants table for multi-tenant support
-- Created: 2026-01-17
-- Requirements: Each tenant represents a Central de Transplantes (e.g., SES-GO, SES-PE, SES-SP)

-- UP

-- Create tenants table (global table, no tenant_id)
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_tenants_slug UNIQUE (slug)
);

-- Index on slug for lookups
CREATE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug);

-- Index on name for display/sorting
CREATE INDEX IF NOT EXISTS idx_tenants_name ON tenants(name);

-- Comments
COMMENT ON TABLE tenants IS 'Centrais de Transplantes - tabela global para suporte multi-tenant';
COMMENT ON COLUMN tenants.id IS 'Identificador unico do tenant (UUID)';
COMMENT ON COLUMN tenants.name IS 'Nome completo da Central de Transplantes (ex: Secretaria de Saude de Goias)';
COMMENT ON COLUMN tenants.slug IS 'Identificador URL-safe unico (ex: ses-go, ses-pe, ses-sp)';
COMMENT ON COLUMN tenants.created_at IS 'Data de criacao do registro';
COMMENT ON COLUMN tenants.updated_at IS 'Data da ultima atualizacao';

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS tenants;
-- Migration: 020_add_tenant_id_to_users
-- Description: Add tenant_id and is_super_admin columns to users table for multi-tenant support
-- Created: 2026-01-17
-- Note: tenant_id will be initially nullable, set with DEFAULT in migration 025, then made NOT NULL in 026

-- UP

-- Add tenant_id column (nullable initially for data migration)
ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- Add is_super_admin column for cross-tenant access
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_super_admin BOOLEAN NOT NULL DEFAULT false;

-- Add foreign key constraint to tenants table
-- Note: This will be enforced after tenants table is populated and data is migrated
ALTER TABLE users
ADD CONSTRAINT fk_users_tenant_id
FOREIGN KEY (tenant_id) REFERENCES tenants(id)
ON DELETE RESTRICT;

-- Create index on tenant_id for query performance
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);

-- Create composite index for tenant + active users
CREATE INDEX IF NOT EXISTS idx_users_tenant_ativo ON users(tenant_id, ativo) WHERE ativo = true;

-- Comments
COMMENT ON COLUMN users.tenant_id IS 'ID do tenant (Central de Transplantes) ao qual o usuario pertence';
COMMENT ON COLUMN users.is_super_admin IS 'Flag para super administradores com acesso cross-tenant';

-- DOWN (for rollback)
-- ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_tenant_id;
-- DROP INDEX IF EXISTS idx_users_tenant_ativo;
-- DROP INDEX IF EXISTS idx_users_tenant_id;
-- ALTER TABLE users DROP COLUMN IF EXISTS is_super_admin;
-- ALTER TABLE users DROP COLUMN IF EXISTS tenant_id;
-- Migration: 021_add_tenant_id_to_hospitals
-- Description: Add tenant_id column to hospitals table for multi-tenant support
-- Created: 2026-01-17
-- Note: tenant_id will be initially nullable, backfilled in migration 025, then made NOT NULL in 026

-- UP

-- Add tenant_id column (nullable initially for data migration)
ALTER TABLE hospitals ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- Add foreign key constraint to tenants table
ALTER TABLE hospitals
ADD CONSTRAINT fk_hospitals_tenant_id
FOREIGN KEY (tenant_id) REFERENCES tenants(id)
ON DELETE RESTRICT;

-- Create index on tenant_id for query performance
CREATE INDEX IF NOT EXISTS idx_hospitals_tenant_id ON hospitals(tenant_id);

-- Create composite index for tenant + active hospitals
CREATE INDEX IF NOT EXISTS idx_hospitals_tenant_ativo ON hospitals(tenant_id, ativo) WHERE ativo = true AND deleted_at IS NULL;

-- Comments
COMMENT ON COLUMN hospitals.tenant_id IS 'ID do tenant (Central de Transplantes) ao qual o hospital pertence';

-- DOWN (for rollback)
-- ALTER TABLE hospitals DROP CONSTRAINT IF EXISTS fk_hospitals_tenant_id;
-- DROP INDEX IF EXISTS idx_hospitals_tenant_ativo;
-- DROP INDEX IF EXISTS idx_hospitals_tenant_id;
-- ALTER TABLE hospitals DROP COLUMN IF EXISTS tenant_id;
-- Migration: 022_add_tenant_id_to_data_tables
-- Description: Add tenant_id column to occurrences, obitos_simulados, triagem_rules, shifts, notifications, audit_logs
-- Created: 2026-01-17
-- Note: tenant_id will be initially nullable, backfilled in migration 025, then made NOT NULL in 026

-- UP

-- =============================================================================
-- OCCURRENCES TABLE
-- =============================================================================
ALTER TABLE occurrences ADD COLUMN IF NOT EXISTS tenant_id UUID;

ALTER TABLE occurrences
ADD CONSTRAINT fk_occurrences_tenant_id
FOREIGN KEY (tenant_id) REFERENCES tenants(id)
ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_occurrences_tenant_id ON occurrences(tenant_id);

-- Composite index for tenant + status queries (most common filter)
CREATE INDEX IF NOT EXISTS idx_occurrences_tenant_status ON occurrences(tenant_id, status);

COMMENT ON COLUMN occurrences.tenant_id IS 'ID do tenant (Central de Transplantes) ao qual a ocorrencia pertence';

-- =============================================================================
-- OBITOS_SIMULADOS TABLE
-- =============================================================================
ALTER TABLE obitos_simulados ADD COLUMN IF NOT EXISTS tenant_id UUID;

ALTER TABLE obitos_simulados
ADD CONSTRAINT fk_obitos_simulados_tenant_id
FOREIGN KEY (tenant_id) REFERENCES tenants(id)
ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_obitos_simulados_tenant_id ON obitos_simulados(tenant_id);

-- Composite index for tenant + unprocessed obitos
CREATE INDEX IF NOT EXISTS idx_obitos_simulados_tenant_processado ON obitos_simulados(tenant_id, processado) WHERE processado = false;

COMMENT ON COLUMN obitos_simulados.tenant_id IS 'ID do tenant (Central de Transplantes) ao qual o obito pertence';

-- =============================================================================
-- TRIAGEM_RULES TABLE
-- =============================================================================
ALTER TABLE triagem_rules ADD COLUMN IF NOT EXISTS tenant_id UUID;

ALTER TABLE triagem_rules
ADD CONSTRAINT fk_triagem_rules_tenant_id
FOREIGN KEY (tenant_id) REFERENCES tenants(id)
ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_triagem_rules_tenant_id ON triagem_rules(tenant_id);

-- Composite index for tenant + active rules
CREATE INDEX IF NOT EXISTS idx_triagem_rules_tenant_ativo ON triagem_rules(tenant_id, ativo, prioridade DESC) WHERE ativo = true;

COMMENT ON COLUMN triagem_rules.tenant_id IS 'ID do tenant (Central de Transplantes) - cada tenant tem regras independentes';

-- =============================================================================
-- SHIFTS TABLE
-- =============================================================================
ALTER TABLE shifts ADD COLUMN IF NOT EXISTS tenant_id UUID;

ALTER TABLE shifts
ADD CONSTRAINT fk_shifts_tenant_id
FOREIGN KEY (tenant_id) REFERENCES tenants(id)
ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_shifts_tenant_id ON shifts(tenant_id);

-- Composite index for tenant + hospital shift queries
CREATE INDEX IF NOT EXISTS idx_shifts_tenant_hospital ON shifts(tenant_id, hospital_id);

COMMENT ON COLUMN shifts.tenant_id IS 'ID do tenant (Central de Transplantes) ao qual o plantao pertence';

-- =============================================================================
-- NOTIFICATIONS TABLE
-- =============================================================================
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS tenant_id UUID;

ALTER TABLE notifications
ADD CONSTRAINT fk_notifications_tenant_id
FOREIGN KEY (tenant_id) REFERENCES tenants(id)
ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_notifications_tenant_id ON notifications(tenant_id);

COMMENT ON COLUMN notifications.tenant_id IS 'ID do tenant (Central de Transplantes) ao qual a notificacao pertence';

-- =============================================================================
-- AUDIT_LOGS TABLE
-- Note: audit_logs has special handling - we need to temporarily disable the update trigger
-- =============================================================================

-- Temporarily disable the no-update trigger for schema changes
DROP TRIGGER IF EXISTS audit_logs_no_update ON audit_logs;

ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- Note: We don't add FK constraint on audit_logs to avoid issues with tenant deletion
-- audit_logs should preserve historical data even if tenant is removed (rare case)
-- ALTER TABLE audit_logs ADD CONSTRAINT fk_audit_logs_tenant_id FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_id ON audit_logs(tenant_id);

-- Composite index for tenant + timestamp queries (primary audit log query pattern)
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_timestamp ON audit_logs(tenant_id, timestamp DESC);

COMMENT ON COLUMN audit_logs.tenant_id IS 'ID do tenant (Central de Transplantes) para filtro de auditoria por tenant';

-- Re-enable the no-update trigger
CREATE TRIGGER audit_logs_no_update
    BEFORE UPDATE ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION prevent_audit_log_update();

-- DOWN (for rollback)
-- DROP TRIGGER IF EXISTS audit_logs_no_update ON audit_logs;
--
-- ALTER TABLE audit_logs DROP COLUMN IF EXISTS tenant_id;
-- DROP INDEX IF EXISTS idx_audit_logs_tenant_timestamp;
-- DROP INDEX IF EXISTS idx_audit_logs_tenant_id;
--
-- ALTER TABLE notifications DROP CONSTRAINT IF EXISTS fk_notifications_tenant_id;
-- ALTER TABLE notifications DROP COLUMN IF EXISTS tenant_id;
-- DROP INDEX IF EXISTS idx_notifications_tenant_id;
--
-- ALTER TABLE shifts DROP CONSTRAINT IF EXISTS fk_shifts_tenant_id;
-- ALTER TABLE shifts DROP COLUMN IF EXISTS tenant_id;
-- DROP INDEX IF EXISTS idx_shifts_tenant_hospital;
-- DROP INDEX IF EXISTS idx_shifts_tenant_id;
--
-- ALTER TABLE triagem_rules DROP CONSTRAINT IF EXISTS fk_triagem_rules_tenant_id;
-- ALTER TABLE triagem_rules DROP COLUMN IF EXISTS tenant_id;
-- DROP INDEX IF EXISTS idx_triagem_rules_tenant_ativo;
-- DROP INDEX IF EXISTS idx_triagem_rules_tenant_id;
--
-- ALTER TABLE obitos_simulados DROP CONSTRAINT IF EXISTS fk_obitos_simulados_tenant_id;
-- ALTER TABLE obitos_simulados DROP COLUMN IF EXISTS tenant_id;
-- DROP INDEX IF EXISTS idx_obitos_simulados_tenant_processado;
-- DROP INDEX IF EXISTS idx_obitos_simulados_tenant_id;
--
-- ALTER TABLE occurrences DROP CONSTRAINT IF EXISTS fk_occurrences_tenant_id;
-- ALTER TABLE occurrences DROP COLUMN IF EXISTS tenant_id;
-- DROP INDEX IF EXISTS idx_occurrences_tenant_status;
-- DROP INDEX IF EXISTS idx_occurrences_tenant_id;
--
-- CREATE TRIGGER audit_logs_no_update
--     BEFORE UPDATE ON audit_logs
--     FOR EACH ROW
--     EXECUTE FUNCTION prevent_audit_log_update();
-- Migration: 023_add_tenant_id_to_user_hospitals
-- Description: Add tenant_id column to user_hospitals junction table for data integrity
-- Created: 2026-01-17
-- Note: tenant_id will be initially nullable, backfilled in migration 025, then made NOT NULL in 026

-- UP

-- Add tenant_id column (nullable initially for data migration)
ALTER TABLE user_hospitals ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- Add foreign key constraint to tenants table
ALTER TABLE user_hospitals
ADD CONSTRAINT fk_user_hospitals_tenant_id
FOREIGN KEY (tenant_id) REFERENCES tenants(id)
ON DELETE RESTRICT;

-- Create index on tenant_id for query performance
CREATE INDEX IF NOT EXISTS idx_user_hospitals_tenant_id ON user_hospitals(tenant_id);

-- Comments
COMMENT ON COLUMN user_hospitals.tenant_id IS 'ID do tenant para garantir integridade - user e hospital devem pertencer ao mesmo tenant';

-- DOWN (for rollback)
-- ALTER TABLE user_hospitals DROP CONSTRAINT IF EXISTS fk_user_hospitals_tenant_id;
-- DROP INDEX IF EXISTS idx_user_hospitals_tenant_id;
-- ALTER TABLE user_hospitals DROP COLUMN IF EXISTS tenant_id;
-- Migration: 024_seed_sesgo_tenant
-- Description: Seed SES-GO tenant as the legacy/first tenant
-- Created: 2026-01-17
-- Note: This creates the SES-GO tenant with a deterministic UUID for reference

-- UP

-- Insert SES-GO tenant with a known UUID (00000000-0000-0000-0000-000000000001)
-- Using ON CONFLICT to make this migration idempotent
INSERT INTO tenants (id, name, slug, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Secretaria de Estado da Saude de Goias',
    'ses-go',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (slug) DO NOTHING;

-- Comment for documentation
COMMENT ON TABLE tenants IS 'Centrais de Transplantes - SES-GO e a primeira Central (legacy data)';

-- DOWN (for rollback)
-- DELETE FROM tenants WHERE slug = 'ses-go';
-- Migration: 025_backfill_tenant_id
-- Description: Backfill tenant_id to all existing records with SES-GO tenant
-- Created: 2026-01-17
-- Note: This assigns all existing data to the SES-GO tenant (legacy migration)

-- UP

-- Define the SES-GO tenant UUID as a constant for this migration
-- Using a DO block to ensure transaction safety
DO $$
DECLARE
    sesgo_tenant_id UUID := '00000000-0000-0000-0000-000000000001';
BEGIN
    -- Verify SES-GO tenant exists
    IF NOT EXISTS (SELECT 1 FROM tenants WHERE id = sesgo_tenant_id) THEN
        RAISE EXCEPTION 'SES-GO tenant not found. Run migration 024 first.';
    END IF;

    -- Backfill users table
    UPDATE users
    SET tenant_id = sesgo_tenant_id
    WHERE tenant_id IS NULL;

    RAISE NOTICE 'Backfilled % users with SES-GO tenant_id', (SELECT COUNT(*) FROM users WHERE tenant_id = sesgo_tenant_id);

    -- Backfill hospitals table
    UPDATE hospitals
    SET tenant_id = sesgo_tenant_id
    WHERE tenant_id IS NULL;

    RAISE NOTICE 'Backfilled % hospitals with SES-GO tenant_id', (SELECT COUNT(*) FROM hospitals WHERE tenant_id = sesgo_tenant_id);

    -- Backfill occurrences table
    UPDATE occurrences
    SET tenant_id = sesgo_tenant_id
    WHERE tenant_id IS NULL;

    RAISE NOTICE 'Backfilled % occurrences with SES-GO tenant_id', (SELECT COUNT(*) FROM occurrences WHERE tenant_id = sesgo_tenant_id);

    -- Backfill obitos_simulados table
    UPDATE obitos_simulados
    SET tenant_id = sesgo_tenant_id
    WHERE tenant_id IS NULL;

    RAISE NOTICE 'Backfilled % obitos_simulados with SES-GO tenant_id', (SELECT COUNT(*) FROM obitos_simulados WHERE tenant_id = sesgo_tenant_id);

    -- Backfill triagem_rules table
    UPDATE triagem_rules
    SET tenant_id = sesgo_tenant_id
    WHERE tenant_id IS NULL;

    RAISE NOTICE 'Backfilled % triagem_rules with SES-GO tenant_id', (SELECT COUNT(*) FROM triagem_rules WHERE tenant_id = sesgo_tenant_id);

    -- Backfill shifts table
    UPDATE shifts
    SET tenant_id = sesgo_tenant_id
    WHERE tenant_id IS NULL;

    RAISE NOTICE 'Backfilled % shifts with SES-GO tenant_id', (SELECT COUNT(*) FROM shifts WHERE tenant_id = sesgo_tenant_id);

    -- Backfill notifications table
    UPDATE notifications
    SET tenant_id = sesgo_tenant_id
    WHERE tenant_id IS NULL;

    RAISE NOTICE 'Backfilled % notifications with SES-GO tenant_id', (SELECT COUNT(*) FROM notifications WHERE tenant_id = sesgo_tenant_id);

    -- Backfill user_hospitals junction table
    UPDATE user_hospitals
    SET tenant_id = sesgo_tenant_id
    WHERE tenant_id IS NULL;

    RAISE NOTICE 'Backfilled % user_hospitals with SES-GO tenant_id', (SELECT COUNT(*) FROM user_hospitals WHERE tenant_id = sesgo_tenant_id);

    -- Handle audit_logs specially (need to disable trigger temporarily)
    -- Note: audit_logs has a no-update trigger, so we drop it, update, and recreate
    DROP TRIGGER IF EXISTS audit_logs_no_update ON audit_logs;

    UPDATE audit_logs
    SET tenant_id = sesgo_tenant_id
    WHERE tenant_id IS NULL;

    RAISE NOTICE 'Backfilled % audit_logs with SES-GO tenant_id', (SELECT COUNT(*) FROM audit_logs WHERE tenant_id = sesgo_tenant_id);

    -- Recreate the audit_logs no-update trigger
    CREATE TRIGGER audit_logs_no_update
        BEFORE UPDATE ON audit_logs
        FOR EACH ROW
        EXECUTE FUNCTION prevent_audit_log_update();

    RAISE NOTICE 'Tenant backfill complete for SES-GO tenant';
END $$;

-- DOWN (for rollback - CAUTION: this removes tenant assignments)
-- Note: Rolling back this migration would require setting all tenant_ids back to NULL
-- This is generally not recommended as it would break tenant isolation
-- DO $$
-- BEGIN
--     DROP TRIGGER IF EXISTS audit_logs_no_update ON audit_logs;
--     UPDATE audit_logs SET tenant_id = NULL;
--     CREATE TRIGGER audit_logs_no_update BEFORE UPDATE ON audit_logs FOR EACH ROW EXECUTE FUNCTION prevent_audit_log_update();
--     UPDATE users SET tenant_id = NULL;
--     UPDATE hospitals SET tenant_id = NULL;
--     UPDATE occurrences SET tenant_id = NULL;
--     UPDATE obitos_simulados SET tenant_id = NULL;
--     UPDATE triagem_rules SET tenant_id = NULL;
--     UPDATE shifts SET tenant_id = NULL;
--     UPDATE notifications SET tenant_id = NULL;
--     UPDATE user_hospitals SET tenant_id = NULL;
-- END $$;
-- Migration: 026_enforce_tenant_id_not_null
-- Description: Add NOT NULL constraint to tenant_id columns after backfill
-- Created: 2026-01-17
-- Note: Run this AFTER migration 025 has backfilled all existing data
-- IMPORTANT: This migration should only be run after confirming all data has tenant_id

-- UP

-- Verify no NULL tenant_ids remain before applying constraints
DO $$
DECLARE
    null_count INTEGER;
BEGIN
    -- Check users
    SELECT COUNT(*) INTO null_count FROM users WHERE tenant_id IS NULL;
    IF null_count > 0 THEN
        RAISE EXCEPTION 'Found % users with NULL tenant_id. Run migration 025 first.', null_count;
    END IF;

    -- Check hospitals
    SELECT COUNT(*) INTO null_count FROM hospitals WHERE tenant_id IS NULL;
    IF null_count > 0 THEN
        RAISE EXCEPTION 'Found % hospitals with NULL tenant_id. Run migration 025 first.', null_count;
    END IF;

    -- Check occurrences
    SELECT COUNT(*) INTO null_count FROM occurrences WHERE tenant_id IS NULL;
    IF null_count > 0 THEN
        RAISE EXCEPTION 'Found % occurrences with NULL tenant_id. Run migration 025 first.', null_count;
    END IF;

    -- Check obitos_simulados
    SELECT COUNT(*) INTO null_count FROM obitos_simulados WHERE tenant_id IS NULL;
    IF null_count > 0 THEN
        RAISE EXCEPTION 'Found % obitos_simulados with NULL tenant_id. Run migration 025 first.', null_count;
    END IF;

    -- Check triagem_rules
    SELECT COUNT(*) INTO null_count FROM triagem_rules WHERE tenant_id IS NULL;
    IF null_count > 0 THEN
        RAISE EXCEPTION 'Found % triagem_rules with NULL tenant_id. Run migration 025 first.', null_count;
    END IF;

    -- Check shifts
    SELECT COUNT(*) INTO null_count FROM shifts WHERE tenant_id IS NULL;
    IF null_count > 0 THEN
        RAISE EXCEPTION 'Found % shifts with NULL tenant_id. Run migration 025 first.', null_count;
    END IF;

    -- Check notifications
    SELECT COUNT(*) INTO null_count FROM notifications WHERE tenant_id IS NULL;
    IF null_count > 0 THEN
        RAISE EXCEPTION 'Found % notifications with NULL tenant_id. Run migration 025 first.', null_count;
    END IF;

    -- Check user_hospitals
    SELECT COUNT(*) INTO null_count FROM user_hospitals WHERE tenant_id IS NULL;
    IF null_count > 0 THEN
        RAISE EXCEPTION 'Found % user_hospitals with NULL tenant_id. Run migration 025 first.', null_count;
    END IF;

    -- Note: audit_logs tenant_id remains nullable to allow for system-level logs
    RAISE NOTICE 'All tenant_id columns verified - no NULL values found';
END $$;

-- Add NOT NULL constraints to tenant_id columns
-- Note: Using COALESCE with a default as fallback during constraint addition
ALTER TABLE users ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE hospitals ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE occurrences ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE obitos_simulados ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE triagem_rules ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE shifts ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE notifications ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE user_hospitals ALTER COLUMN tenant_id SET NOT NULL;

-- Note: audit_logs.tenant_id intentionally remains nullable
-- This allows for cross-tenant audit entries and system-level logs
COMMENT ON COLUMN audit_logs.tenant_id IS 'ID do tenant - nullable para permitir logs de sistema/cross-tenant';

-- Comments
COMMENT ON COLUMN users.tenant_id IS 'ID do tenant (NOT NULL) - cada usuario pertence a exatamente um tenant';
COMMENT ON COLUMN hospitals.tenant_id IS 'ID do tenant (NOT NULL) - cada hospital pertence a exatamente um tenant';
COMMENT ON COLUMN occurrences.tenant_id IS 'ID do tenant (NOT NULL) - cada ocorrencia pertence a exatamente um tenant';
COMMENT ON COLUMN obitos_simulados.tenant_id IS 'ID do tenant (NOT NULL) - cada obito pertence a exatamente um tenant';
COMMENT ON COLUMN triagem_rules.tenant_id IS 'ID do tenant (NOT NULL) - regras de triagem sao independentes por tenant';
COMMENT ON COLUMN shifts.tenant_id IS 'ID do tenant (NOT NULL) - cada plantao pertence a exatamente um tenant';
COMMENT ON COLUMN notifications.tenant_id IS 'ID do tenant (NOT NULL) - cada notificacao pertence a exatamente um tenant';
COMMENT ON COLUMN user_hospitals.tenant_id IS 'ID do tenant (NOT NULL) - garante integridade referencial dentro do tenant';

-- DOWN (for rollback)
-- ALTER TABLE users ALTER COLUMN tenant_id DROP NOT NULL;
-- ALTER TABLE hospitals ALTER COLUMN tenant_id DROP NOT NULL;
-- ALTER TABLE occurrences ALTER COLUMN tenant_id DROP NOT NULL;
-- ALTER TABLE obitos_simulados ALTER COLUMN tenant_id DROP NOT NULL;
-- ALTER TABLE triagem_rules ALTER COLUMN tenant_id DROP NOT NULL;
-- ALTER TABLE shifts ALTER COLUMN tenant_id DROP NOT NULL;
-- ALTER TABLE notifications ALTER COLUMN tenant_id DROP NOT NULL;
-- ALTER TABLE user_hospitals ALTER COLUMN tenant_id DROP NOT NULL;
