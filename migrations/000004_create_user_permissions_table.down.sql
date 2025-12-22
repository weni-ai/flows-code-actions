-- Drop user_permissions table
-- Migration: 000004_create_user_permissions_table (DOWN)

-- Drop indexes first
DROP INDEX IF EXISTS idx_user_permissions_unique_project_email;
DROP INDEX IF EXISTS idx_user_permissions_project_email;
DROP INDEX IF EXISTS idx_user_permissions_email;
DROP INDEX IF EXISTS idx_user_permissions_project_uuid;

-- Drop table
DROP TABLE IF EXISTS user_permissions;
