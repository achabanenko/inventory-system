-- Complete OAuth schema migration
-- This script fixes all the missing schema changes for OAuth support

-- 1. Add OAuth fields to users table
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50),
ADD COLUMN IF NOT EXISTS oauth_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS avatar_url TEXT;

-- 2. Make password_hash optional for OAuth users
ALTER TABLE users 
ALTER COLUMN password_hash DROP NOT NULL;

-- 3. Add tenant_id column if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'tenant_id'
    ) THEN
        ALTER TABLE users ADD COLUMN tenant_id UUID;
    END IF;
END $$;

-- 4. Make tenant_id optional (remove NOT NULL constraint)
ALTER TABLE users 
ALTER COLUMN tenant_id DROP NOT NULL;

-- 5. Add foreign key constraint for tenant_id if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'users_tenant_id_fkey' 
        AND table_name = 'users'
    ) THEN
        ALTER TABLE users 
        ADD CONSTRAINT users_tenant_id_fkey 
        FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- 6. Add indexes for OAuth lookups
CREATE INDEX IF NOT EXISTS idx_users_oauth ON users(oauth_provider, oauth_id) 
WHERE oauth_provider IS NOT NULL AND oauth_id IS NOT NULL;

-- 7. Add index for users without tenants
CREATE INDEX IF NOT EXISTS idx_users_no_tenant ON users(tenant_id) 
WHERE tenant_id IS NULL;

-- 8. Add index for tenant lookups
CREATE INDEX IF NOT EXISTS idx_users_tenant ON users(tenant_id) 
WHERE tenant_id IS NOT NULL;

-- 9. Verify the changes
SELECT 
    column_name, 
    is_nullable, 
    data_type 
FROM information_schema.columns 
WHERE table_name = 'users' 
AND column_name IN ('tenant_id', 'oauth_provider', 'oauth_id', 'avatar_url', 'password_hash')
ORDER BY column_name;
