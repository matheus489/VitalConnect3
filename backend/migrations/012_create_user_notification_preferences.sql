-- Migration: 009_create_user_notification_preferences
-- Description: Create user_notification_preferences table for SMS/Email preferences
-- Created: 2026-01-16

-- UP
CREATE TABLE IF NOT EXISTS user_notification_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sms_enabled BOOLEAN NOT NULL DEFAULT true,
    email_enabled BOOLEAN NOT NULL DEFAULT true,
    dashboard_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_notification_preferences_user_id ON user_notification_preferences(user_id);

-- Comments
COMMENT ON TABLE user_notification_preferences IS 'User notification preferences for SMS, email and dashboard alerts';
COMMENT ON COLUMN user_notification_preferences.sms_enabled IS 'Enable SMS notifications (default: true if mobile_phone present)';
COMMENT ON COLUMN user_notification_preferences.email_enabled IS 'Enable email notifications (default: true)';
COMMENT ON COLUMN user_notification_preferences.dashboard_enabled IS 'Enable dashboard notifications (always true, not editable)';

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS user_notification_preferences;
