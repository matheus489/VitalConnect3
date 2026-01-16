-- Migration: 016_create_push_subscriptions
-- Description: Create push_subscriptions table for FCM tokens
-- Created: 2026-01-16

-- UP
CREATE TABLE IF NOT EXISTS push_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    platform VARCHAR(20) NOT NULL DEFAULT 'web',
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_push_subscriptions_token ON push_subscriptions(token);
CREATE INDEX IF NOT EXISTS idx_push_subscriptions_user_id ON push_subscriptions(user_id);

-- Comments
COMMENT ON TABLE push_subscriptions IS 'FCM push notification subscriptions for users';
COMMENT ON COLUMN push_subscriptions.token IS 'FCM registration token from client';
COMMENT ON COLUMN push_subscriptions.platform IS 'Platform: web, android, ios';
COMMENT ON COLUMN push_subscriptions.user_agent IS 'Browser/device user agent for debugging';

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS push_subscriptions;
