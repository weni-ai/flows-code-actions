-- Drop coderuns table
-- Migration: 000003_create_coderuns_table (DOWN)

-- Drop indexes first
DROP INDEX IF EXISTS idx_coderuns_params;
DROP INDEX IF EXISTS idx_coderuns_extra;
DROP INDEX IF EXISTS idx_coderuns_code_id_created_at;
DROP INDEX IF EXISTS idx_coderuns_created_at;
DROP INDEX IF EXISTS idx_coderuns_status;
DROP INDEX IF EXISTS idx_coderuns_code_mongo_id;
DROP INDEX IF EXISTS idx_coderuns_code_id;
DROP INDEX IF EXISTS idx_coderuns_mongo_object_id;

-- Drop table
DROP TABLE IF EXISTS coderuns;
