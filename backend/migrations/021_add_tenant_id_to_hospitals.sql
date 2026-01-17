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
