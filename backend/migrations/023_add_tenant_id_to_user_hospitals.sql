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
