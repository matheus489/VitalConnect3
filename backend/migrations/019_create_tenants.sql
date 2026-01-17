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
