-- Drop secrets table
-- Migration: 000006_create_secrets_table (DOWN)

-- Drop indexes first
DROP INDEX IF EXISTS idx_secrets_created_at;
DROP INDEX IF EXISTS idx_secrets_code_id;

-- Drop table
DROP TABLE IF EXISTS secrets;
