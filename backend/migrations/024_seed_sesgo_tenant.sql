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
