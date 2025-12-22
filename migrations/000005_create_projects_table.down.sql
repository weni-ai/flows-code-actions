-- Drop projects table
-- Migration: 000005_create_projects_table (DOWN)

-- Drop indexes first
DROP INDEX IF EXISTS idx_projects_authorizations;
DROP INDEX IF EXISTS idx_projects_created_at;
DROP INDEX IF EXISTS idx_projects_name;
DROP INDEX IF EXISTS idx_projects_uuid;

-- Drop table
DROP TABLE IF EXISTS projects;
