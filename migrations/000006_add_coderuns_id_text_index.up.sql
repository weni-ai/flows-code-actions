-- Add expression indexes for text-casted UUID lookups
-- Migration: 000006_add_coderuns_id_text_index
--
-- Problem: Queries using `WHERE id::text = $1` cannot use the primary key index
-- because PostgreSQL requires an expression index to optimize casted column lookups.
--
-- This index enables efficient lookups when the application queries by ID as text,
-- which happens in dual-ID queries (PostgreSQL UUID or MongoDB ObjectID).

-- Expression index for id::text lookups on coderuns
-- Used by: GetByID, Update, Delete queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_coderuns_id_text ON coderuns((id::text));

-- Expression index for code_id::text lookups on coderuns
-- Used by: ListByCodeID query
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_coderuns_code_id_text ON coderuns((code_id::text));

-- Add comments for documentation
COMMENT ON INDEX idx_coderuns_id_text IS 'Expression index for efficient id::text lookups in dual-ID queries';
COMMENT ON INDEX idx_coderuns_code_id_text IS 'Expression index for efficient code_id::text lookups in ListByCodeID';
