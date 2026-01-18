-- Migration: 030_create_ai_tables
-- Description: Create AI conversation history and action audit log tables
-- Created: 2026-01-18
-- Requirements: Support AI assistant with conversation persistence and action auditing

-- UP

-- Create enum for conversation message roles
DO $$ BEGIN
    CREATE TYPE ai_message_role AS ENUM ('user', 'assistant', 'system');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create enum for AI action status
DO $$ BEGIN
    CREATE TYPE ai_action_status AS ENUM ('pending', 'success', 'failed', 'cancelled');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create ai_conversation_history table
CREATE TABLE IF NOT EXISTS ai_conversation_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id UUID NOT NULL,
    role ai_message_role NOT NULL,
    content TEXT NOT NULL,
    tool_calls JSONB,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for ai_conversation_history
-- Primary lookup by session for conversation retrieval
CREATE INDEX IF NOT EXISTS idx_ai_conversation_session_id
    ON ai_conversation_history(session_id);

-- Tenant isolation and filtering
CREATE INDEX IF NOT EXISTS idx_ai_conversation_tenant_id
    ON ai_conversation_history(tenant_id);

-- User's conversations lookup
CREATE INDEX IF NOT EXISTS idx_ai_conversation_user_id
    ON ai_conversation_history(user_id);

-- Timeline queries ordered by creation
CREATE INDEX IF NOT EXISTS idx_ai_conversation_created_at
    ON ai_conversation_history(created_at DESC);

-- Composite index for user's sessions within tenant
CREATE INDEX IF NOT EXISTS idx_ai_conversation_tenant_user
    ON ai_conversation_history(tenant_id, user_id, created_at DESC);

-- Composite index for session messages ordered by time
CREATE INDEX IF NOT EXISTS idx_ai_conversation_session_timeline
    ON ai_conversation_history(session_id, created_at ASC);

-- Create ai_action_audit_log table
CREATE TABLE IF NOT EXISTS ai_action_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    conversation_id UUID REFERENCES ai_conversation_history(id) ON DELETE SET NULL,
    action_type VARCHAR(100) NOT NULL,
    tool_name VARCHAR(100),
    input_params JSONB NOT NULL DEFAULT '{}',
    output_result JSONB NOT NULL DEFAULT '{}',
    status ai_action_status NOT NULL DEFAULT 'pending',
    execution_time_ms INTEGER,
    error_message TEXT,
    severity audit_severity NOT NULL DEFAULT 'INFO',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for ai_action_audit_log
-- Tenant isolation
CREATE INDEX IF NOT EXISTS idx_ai_audit_tenant_id
    ON ai_action_audit_log(tenant_id);

-- User's action history
CREATE INDEX IF NOT EXISTS idx_ai_audit_user_id
    ON ai_action_audit_log(user_id);

-- Conversation correlation
CREATE INDEX IF NOT EXISTS idx_ai_audit_conversation_id
    ON ai_action_audit_log(conversation_id);

-- Action type filtering (ai.query, ai.tool_execution, ai.confirmation)
CREATE INDEX IF NOT EXISTS idx_ai_audit_action_type
    ON ai_action_audit_log(action_type);

-- Status filtering for pending actions
CREATE INDEX IF NOT EXISTS idx_ai_audit_status
    ON ai_action_audit_log(status);

-- Timeline queries
CREATE INDEX IF NOT EXISTS idx_ai_audit_created_at
    ON ai_action_audit_log(created_at DESC);

-- Composite index for tenant audit queries
CREATE INDEX IF NOT EXISTS idx_ai_audit_tenant_timeline
    ON ai_action_audit_log(tenant_id, created_at DESC);

-- Composite index for user action history within tenant
CREATE INDEX IF NOT EXISTS idx_ai_audit_tenant_user
    ON ai_action_audit_log(tenant_id, user_id, created_at DESC);

-- Table comments
COMMENT ON TABLE ai_conversation_history IS 'Historico de conversas com o assistente AI - isolado por tenant';
COMMENT ON COLUMN ai_conversation_history.id IS 'Identificador unico da mensagem';
COMMENT ON COLUMN ai_conversation_history.tenant_id IS 'ID do tenant (Central de Transplantes)';
COMMENT ON COLUMN ai_conversation_history.user_id IS 'ID do usuario que participou da conversa';
COMMENT ON COLUMN ai_conversation_history.session_id IS 'ID da sessao de conversa para agrupar mensagens';
COMMENT ON COLUMN ai_conversation_history.role IS 'Papel do autor: user, assistant ou system';
COMMENT ON COLUMN ai_conversation_history.content IS 'Conteudo da mensagem';
COMMENT ON COLUMN ai_conversation_history.tool_calls IS 'Chamadas de ferramentas executadas (JSON)';
COMMENT ON COLUMN ai_conversation_history.metadata IS 'Metadados adicionais da mensagem (JSON)';
COMMENT ON COLUMN ai_conversation_history.created_at IS 'Data de criacao da mensagem';

COMMENT ON TABLE ai_action_audit_log IS 'Log de auditoria de acoes do assistente AI - imutavel para compliance';
COMMENT ON COLUMN ai_action_audit_log.id IS 'Identificador unico do registro de auditoria';
COMMENT ON COLUMN ai_action_audit_log.tenant_id IS 'ID do tenant (Central de Transplantes)';
COMMENT ON COLUMN ai_action_audit_log.user_id IS 'ID do usuario que solicitou a acao';
COMMENT ON COLUMN ai_action_audit_log.conversation_id IS 'ID da mensagem de conversa relacionada';
COMMENT ON COLUMN ai_action_audit_log.action_type IS 'Tipo de acao: ai.query, ai.tool_execution, ai.confirmation';
COMMENT ON COLUMN ai_action_audit_log.tool_name IS 'Nome da ferramenta executada (se aplicavel)';
COMMENT ON COLUMN ai_action_audit_log.input_params IS 'Parametros de entrada da acao (JSON)';
COMMENT ON COLUMN ai_action_audit_log.output_result IS 'Resultado da acao (JSON)';
COMMENT ON COLUMN ai_action_audit_log.status IS 'Status da acao: pending, success, failed, cancelled';
COMMENT ON COLUMN ai_action_audit_log.execution_time_ms IS 'Tempo de execucao em milissegundos';
COMMENT ON COLUMN ai_action_audit_log.error_message IS 'Mensagem de erro (se falhou)';
COMMENT ON COLUMN ai_action_audit_log.severity IS 'Nivel de severidade: INFO, WARN, CRITICAL';
COMMENT ON COLUMN ai_action_audit_log.created_at IS 'Data de criacao do registro';

-- DOWN (for rollback)
-- DROP INDEX IF EXISTS idx_ai_audit_tenant_user;
-- DROP INDEX IF EXISTS idx_ai_audit_tenant_timeline;
-- DROP INDEX IF EXISTS idx_ai_audit_created_at;
-- DROP INDEX IF EXISTS idx_ai_audit_status;
-- DROP INDEX IF EXISTS idx_ai_audit_action_type;
-- DROP INDEX IF EXISTS idx_ai_audit_conversation_id;
-- DROP INDEX IF EXISTS idx_ai_audit_user_id;
-- DROP INDEX IF EXISTS idx_ai_audit_tenant_id;
-- DROP TABLE IF EXISTS ai_action_audit_log;
-- DROP INDEX IF EXISTS idx_ai_conversation_session_timeline;
-- DROP INDEX IF EXISTS idx_ai_conversation_tenant_user;
-- DROP INDEX IF EXISTS idx_ai_conversation_created_at;
-- DROP INDEX IF EXISTS idx_ai_conversation_user_id;
-- DROP INDEX IF EXISTS idx_ai_conversation_tenant_id;
-- DROP INDEX IF EXISTS idx_ai_conversation_session_id;
-- DROP TABLE IF EXISTS ai_conversation_history;
-- DROP TYPE IF EXISTS ai_action_status;
-- DROP TYPE IF EXISTS ai_message_role;
