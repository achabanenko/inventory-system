-- Add OAuth fields to users table
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50),
ADD COLUMN IF NOT EXISTS oauth_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS avatar_url TEXT;

-- Make password_hash optional for OAuth users
ALTER TABLE users 
ALTER COLUMN password_hash DROP NOT NULL;

-- Temporarily allow users without tenant_id for OAuth flow
-- Users will be assigned to tenants later
-- First, drop any existing foreign key constraints
DO $$ 
BEGIN
    -- Check if foreign key constraint exists and drop it
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'users_tenant_id_fkey' 
        AND table_name = 'users'
    ) THEN
        ALTER TABLE users DROP CONSTRAINT users_tenant_id_fkey;
    END IF;
END $$;

-- Now make tenant_id optional
ALTER TABLE users 
ALTER COLUMN tenant_id DROP NOT NULL;

-- Re-add foreign key constraint but allow NULL values
ALTER TABLE users 
ADD CONSTRAINT users_tenant_id_fkey 
FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

-- Add index for OAuth lookups
CREATE INDEX IF NOT EXISTS idx_users_oauth ON users(oauth_provider, oauth_id) 
WHERE oauth_provider IS NOT NULL AND oauth_id IS NOT NULL;

-- Add index for users without tenants
CREATE INDEX IF NOT EXISTS idx_users_no_tenant ON users(tenant_id) 
WHERE tenant_id IS NULL;
