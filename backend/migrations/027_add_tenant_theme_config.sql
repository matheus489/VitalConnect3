-- Migration: 027_add_tenant_theme_config
-- Description: Extend tenants table with theme configuration and branding fields
-- Created: 2026-01-17
-- Purpose: Support super-admin backoffice for tenant UI customization

-- UP

-- Add theme configuration and branding columns to tenants table
ALTER TABLE tenants
ADD COLUMN IF NOT EXISTS theme_config JSONB DEFAULT '{}',
ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true,
ADD COLUMN IF NOT EXISTS logo_url TEXT,
ADD COLUMN IF NOT EXISTS favicon_url TEXT;

-- Set default theme_config structure for existing tenants
UPDATE tenants
SET theme_config = '{
  "theme": {
    "colors": {
      "primary": "#2563eb",
      "secondary": "#64748b",
      "background": "#ffffff",
      "foreground": "#0f172a",
      "muted": "#f1f5f9",
      "accent": "#f59e0b"
    },
    "fonts": {
      "body": "Inter",
      "heading": "Inter"
    }
  },
  "layout": {
    "sidebar": [],
    "topbar": {
      "show_user_info": true,
      "show_tenant_logo": true
    },
    "dashboard_widgets": []
  }
}'::jsonb
WHERE theme_config = '{}'::jsonb OR theme_config IS NULL;

-- Create GIN index on theme_config for efficient JSONB queries
CREATE INDEX IF NOT EXISTS idx_tenants_theme_config ON tenants USING GIN (theme_config);

-- Create index on is_active for filtering active/inactive tenants
CREATE INDEX IF NOT EXISTS idx_tenants_is_active ON tenants(is_active);

-- Comments
COMMENT ON COLUMN tenants.theme_config IS 'JSONB configuration for tenant UI theming - includes colors, fonts, sidebar items, and dashboard widgets';
COMMENT ON COLUMN tenants.is_active IS 'Flag to enable/disable tenant access (soft delete behavior)';
COMMENT ON COLUMN tenants.logo_url IS 'URL to tenant logo image for branding';
COMMENT ON COLUMN tenants.favicon_url IS 'URL to tenant favicon for browser tab';

-- DOWN (for rollback)
-- DROP INDEX IF EXISTS idx_tenants_is_active;
-- DROP INDEX IF EXISTS idx_tenants_theme_config;
-- ALTER TABLE tenants DROP COLUMN IF EXISTS favicon_url;
-- ALTER TABLE tenants DROP COLUMN IF EXISTS logo_url;
-- ALTER TABLE tenants DROP COLUMN IF EXISTS is_active;
-- ALTER TABLE tenants DROP COLUMN IF EXISTS theme_config;
