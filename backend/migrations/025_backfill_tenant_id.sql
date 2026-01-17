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
