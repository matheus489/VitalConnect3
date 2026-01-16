-- Migration: 008_add_email_notifications
-- Description: Add email_notifications field to users table
-- Created: 2026-01-16

-- UP
ALTER TABLE users
ADD COLUMN IF NOT EXISTS email_notifications BOOLEAN NOT NULL DEFAULT true;

-- Index for filtering by email notification preference
CREATE INDEX IF NOT EXISTS idx_users_email_notifications ON users(email_notifications);

-- Comments
COMMENT ON COLUMN users.email_notifications IS 'Preferencia de notificacao por email (true = ativo, dashboard sempre ativo)';

-- DOWN (for rollback)
-- ALTER TABLE users DROP COLUMN IF EXISTS email_notifications;
-- DROP INDEX IF EXISTS idx_users_email_notifications;
