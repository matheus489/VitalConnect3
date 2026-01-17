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
