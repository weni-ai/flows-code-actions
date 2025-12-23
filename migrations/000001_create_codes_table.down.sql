-- Drop codes table
-- Migration: 000001_create_codes_table (DOWN)

-- Drop indexes first
DROP INDEX IF EXISTS idx_codes_created_at;
DROP INDEX IF EXISTS idx_codes_project_type;
DROP INDEX IF EXISTS idx_codes_type;
DROP INDEX IF EXISTS idx_codes_project_uuid;
DROP INDEX IF EXISTS idx_codes_mongo_object_id;

-- Drop table
DROP TABLE IF EXISTS codes;
