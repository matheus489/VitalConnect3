-- Migration: 010_migrate_hospital_id_to_user_hospitals
-- Description: Migrate existing hospital_id data to user_hospitals table and remove column
-- Created: 2026-01-16

-- UP

-- Step 1: Migrate existing data from users.hospital_id to user_hospitals
INSERT INTO user_hospitals (user_id, hospital_id, created_at)
SELECT id, hospital_id, CURRENT_TIMESTAMP
FROM users
WHERE hospital_id IS NOT NULL
ON CONFLICT (user_id, hospital_id) DO NOTHING;

-- Step 2: Drop the hospital_id foreign key constraint first
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_hospital_id_fkey;

-- Step 3: Drop the index on hospital_id
DROP INDEX IF EXISTS idx_users_hospital_id;

-- Step 4: Remove the hospital_id column
ALTER TABLE users DROP COLUMN IF EXISTS hospital_id;

-- Comments
COMMENT ON TABLE users IS 'Usuarios do sistema SIDOT - relacao N:N com hospitais via user_hospitals';

-- DOWN (for rollback - complex, requires recreating column)
-- ALTER TABLE users ADD COLUMN hospital_id UUID REFERENCES hospitals(id) ON DELETE SET NULL;
-- CREATE INDEX IF NOT EXISTS idx_users_hospital_id ON users(hospital_id);
-- UPDATE users u SET hospital_id = (SELECT hospital_id FROM user_hospitals uh WHERE uh.user_id = u.id LIMIT 1);
