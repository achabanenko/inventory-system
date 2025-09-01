-- Insert default tenant first (if not exists)
INSERT INTO tenants (id, name, slug, is_active, settings, contact)
VALUES (
  gen_random_uuid(),
  'Default Company', 
  'default', 
  true, 
  '{"currency": "USD", "timezone": "UTC"}',
  '{"email": "", "phone": ""}'
) 
ON CONFLICT (slug) DO NOTHING;

-- Insert admin user
INSERT INTO users (id, email, password_hash, name, role, tenant_id, is_active)
SELECT 
  gen_random_uuid(),
  'admin@example.com',
  '$2a$10$yQChQTp1s49xoVrXMkJRFu8dZJY6mHhbaGyg85QmJW5omMmgW29pG', -- bcrypt hash for 'admin123'
  'Admin User',
  'ADMIN',
  t.id,
  true
FROM tenants t 
WHERE t.slug = 'default'
AND NOT EXISTS (SELECT 1 FROM users WHERE email = 'admin@example.com');