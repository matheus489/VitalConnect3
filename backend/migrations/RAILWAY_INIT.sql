-- =====================================================
-- SIDOT - SCRIPT DE INICIALIZACAO PARA RAILWAY
-- Execute este script no PostgreSQL do Railway
-- =====================================================

-- Habilitar extensao UUID (necessaria para gen_random_uuid e uuid_generate_v4)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Criar tipos ENUM
DO $$ BEGIN
    CREATE TYPE user_role AS ENUM ('operador', 'gestor', 'admin');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
    CREATE TYPE occurrence_status AS ENUM ('PENDENTE', 'EM_ANDAMENTO', 'ACEITA', 'RECUSADA', 'CANCELADA', 'CONCLUIDA');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
    CREATE TYPE outcome_type AS ENUM ('doacao_realizada', 'nao_autorizado_familia', 'contraindicacao_medica', 'janela_expirada', 'outros');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
    CREATE TYPE notification_channel AS ENUM ('dashboard', 'email', 'sms', 'push');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
    CREATE TYPE audit_severity AS ENUM ('INFO', 'WARN', 'CRITICAL');
EXCEPTION WHEN duplicate_object THEN null; END $$;

-- =====================================================
-- TABELAS PRINCIPAIS
-- =====================================================

-- TENANTS (Multi-tenant support)
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_tenants_slug UNIQUE (slug)
);

-- Seed tenant SES-GO
INSERT INTO tenants (id, name, slug) VALUES
('00000000-0000-0000-0000-000000000001', 'Secretaria de Estado da Saude de Goias', 'ses-go')
ON CONFLICT (slug) DO NOTHING;

-- HOSPITALS
CREATE TABLE IF NOT EXISTS hospitals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    nome VARCHAR(255) NOT NULL,
    codigo VARCHAR(50) NOT NULL,
    endereco TEXT,
    telefone VARCHAR(20),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    config_conexao JSONB DEFAULT '{}',
    ativo BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_hospitals_codigo ON hospitals(codigo) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_hospitals_tenant_id ON hospitals(tenant_id);

-- USERS
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nome VARCHAR(255) NOT NULL,
    role user_role NOT NULL DEFAULT 'operador',
    mobile_phone VARCHAR(16),
    email_notifications BOOLEAN NOT NULL DEFAULT true,
    is_super_admin BOOLEAN NOT NULL DEFAULT false,
    ativo BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);

-- USER_HOSPITALS (junction table)
CREATE TABLE IF NOT EXISTS user_hospitals (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    hospital_id UUID NOT NULL REFERENCES hospitals(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, hospital_id)
);

-- OBITOS_SIMULADOS
CREATE TABLE IF NOT EXISTS obitos_simulados (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
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

CREATE INDEX IF NOT EXISTS idx_obitos_simulados_tenant_id ON obitos_simulados(tenant_id);
CREATE INDEX IF NOT EXISTS idx_obitos_simulados_processado ON obitos_simulados(processado);

-- OCCURRENCES
CREATE TABLE IF NOT EXISTS occurrences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    obito_id UUID NOT NULL REFERENCES obitos_simulados(id) ON DELETE RESTRICT,
    hospital_id UUID NOT NULL REFERENCES hospitals(id) ON DELETE RESTRICT,
    status occurrence_status NOT NULL DEFAULT 'PENDENTE',
    score_priorizacao INTEGER NOT NULL DEFAULT 50,
    nome_paciente_mascarado VARCHAR(255) NOT NULL,
    dados_completos JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notificado_em TIMESTAMP WITH TIME ZONE,
    data_obito TIMESTAMP WITH TIME ZONE NOT NULL,
    janela_expira_em TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_occurrences_tenant_id ON occurrences(tenant_id);
CREATE INDEX IF NOT EXISTS idx_occurrences_status ON occurrences(status);

-- OCCURRENCE_HISTORY
CREATE TABLE IF NOT EXISTS occurrence_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    occurrence_id UUID NOT NULL REFERENCES occurrences(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    acao VARCHAR(100) NOT NULL,
    status_anterior occurrence_status,
    status_novo occurrence_status,
    observacoes TEXT,
    desfecho outcome_type,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_occurrence_history_occurrence_id ON occurrence_history(occurrence_id);

-- TRIAGEM_RULES
CREATE TABLE IF NOT EXISTS triagem_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    nome VARCHAR(255) NOT NULL,
    descricao TEXT,
    regras JSONB NOT NULL DEFAULT '{}',
    ativo BOOLEAN NOT NULL DEFAULT true,
    prioridade INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_triagem_rules_tenant_id ON triagem_rules(tenant_id);

-- NOTIFICATIONS
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    occurrence_id UUID NOT NULL REFERENCES occurrences(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    canal notification_channel NOT NULL,
    enviado_em TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status_envio VARCHAR(50) NOT NULL DEFAULT 'enviado',
    erro_mensagem TEXT,
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_notifications_tenant_id ON notifications(tenant_id);

-- SHIFTS
CREATE TABLE IF NOT EXISTS shifts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    hospital_id UUID NOT NULL REFERENCES hospitals(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_shifts_tenant_id ON shifts(tenant_id);

-- AUDIT_LOGS
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
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

CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_id ON audit_logs(tenant_id);

-- Triggers para imutabilidade do audit_logs
CREATE OR REPLACE FUNCTION prevent_audit_log_update() RETURNS TRIGGER AS $$
BEGIN RAISE EXCEPTION 'UPDATE not allowed on audit_logs'; END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION prevent_audit_log_delete() RETURNS TRIGGER AS $$
BEGIN RAISE EXCEPTION 'DELETE not allowed on audit_logs'; END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS audit_logs_no_update ON audit_logs;
CREATE TRIGGER audit_logs_no_update BEFORE UPDATE ON audit_logs FOR EACH ROW EXECUTE FUNCTION prevent_audit_log_update();

DROP TRIGGER IF EXISTS audit_logs_no_delete ON audit_logs;
CREATE TRIGGER audit_logs_no_delete BEFORE DELETE ON audit_logs FOR EACH ROW EXECUTE FUNCTION prevent_audit_log_delete();

-- REPORT_AUDIT_LOGS
CREATE TABLE IF NOT EXISTS report_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tipo_relatorio VARCHAR(10) NOT NULL CHECK (tipo_relatorio IN ('CSV', 'PDF')),
    filtros JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- PUSH_SUBSCRIPTIONS
CREATE TABLE IF NOT EXISTS push_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    platform VARCHAR(20) NOT NULL DEFAULT 'web',
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_push_subscriptions_token ON push_subscriptions(token);

-- USER_NOTIFICATION_PREFERENCES
CREATE TABLE IF NOT EXISTS user_notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sms_enabled BOOLEAN NOT NULL DEFAULT true,
    email_enabled BOOLEAN NOT NULL DEFAULT true,
    dashboard_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_notification_preferences_user_id ON user_notification_preferences(user_id);

-- =====================================================
-- CRIAR USUARIO ADMIN INICIAL
-- Senha: admin123 (hash bcrypt)
-- =====================================================
INSERT INTO users (id, tenant_id, email, password_hash, nome, role, ativo)
VALUES (
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    '00000000-0000-0000-0000-000000000001',
    'admin@sidot.gov.br',
    '$2a$10$N9qo8uLOickgx2ZMRZoMye/eDKsYqLhLrJC9L7uVvKdLdJYK6H5VO',
    'Administrador SIDOT',
    'admin',
    true
)
ON CONFLICT (email) DO NOTHING;

-- =====================================================
-- FIM DO SCRIPT
-- Usuario: admin@sidot.gov.br
-- Senha: admin123
-- =====================================================
