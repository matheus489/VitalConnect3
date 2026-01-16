-- Migration: 008_add_mobile_phone_to_users
-- Description: Add mobile_phone field to users table for SMS notifications
-- Created: 2026-01-16

-- UP
ALTER TABLE users ADD COLUMN IF NOT EXISTS mobile_phone VARCHAR(16);

-- Add comment for documentation
COMMENT ON COLUMN users.mobile_phone IS 'Mobile phone number in E.164 format (e.g., +5511999999999) for SMS notifications';

-- DOWN (for rollback)
-- ALTER TABLE users DROP COLUMN IF EXISTS mobile_phone;
