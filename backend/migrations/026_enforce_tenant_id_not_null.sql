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
