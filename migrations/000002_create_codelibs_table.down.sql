-- Drop codelibs table
-- Migration: 000002_create_codelibs_table (DOWN)

-- Drop indexes first
DROP INDEX IF EXISTS idx_codelibs_unique_name_language;
DROP INDEX IF EXISTS idx_codelibs_created_at;
DROP INDEX IF EXISTS idx_codelibs_name_language;
DROP INDEX IF EXISTS idx_codelibs_language;
DROP INDEX IF EXISTS idx_codelibs_name;

-- Drop table
DROP TABLE IF EXISTS codelibs;
