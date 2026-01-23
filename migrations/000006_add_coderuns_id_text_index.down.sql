-- Drop expression indexes for text-casted UUID lookups
-- Migration: 000006_add_coderuns_id_text_index (DOWN)

DROP INDEX CONCURRENTLY IF EXISTS idx_coderuns_code_id_text;
DROP INDEX CONCURRENTLY IF EXISTS idx_coderuns_id_text;
