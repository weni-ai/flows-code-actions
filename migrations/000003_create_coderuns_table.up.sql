-- Create coderuns table
-- Migration: 000003_create_coderuns_table

CREATE TABLE IF NOT EXISTS coderuns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mongo_object_id TEXT UNIQUE,
    code_id UUID NOT NULL,
    code_mongo_id TEXT,
    status VARCHAR(50) NOT NULL CHECK (status IN ('queued', 'started', 'completed', 'failed')),
    result TEXT,
    extra JSONB DEFAULT '{}'::jsonb,
    params JSONB DEFAULT '{}'::jsonb,
    body TEXT,
    headers JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_coderuns_mongo_object_id ON coderuns(mongo_object_id);
CREATE INDEX IF NOT EXISTS idx_coderuns_code_id ON coderuns(code_id);
CREATE INDEX IF NOT EXISTS idx_coderuns_code_mongo_id ON coderuns(code_mongo_id);
CREATE INDEX IF NOT EXISTS idx_coderuns_status ON coderuns(status);
CREATE INDEX IF NOT EXISTS idx_coderuns_created_at ON coderuns(created_at);
CREATE INDEX IF NOT EXISTS idx_coderuns_code_id_created_at ON coderuns(code_id, created_at);

-- Index for JSONB fields (GIN indexes for efficient JSON queries)
CREATE INDEX IF NOT EXISTS idx_coderuns_extra ON coderuns USING GIN (extra);
CREATE INDEX IF NOT EXISTS idx_coderuns_params ON coderuns USING GIN (params);

-- Add comments for documentation
COMMENT ON TABLE coderuns IS 'Stores code execution runs with their parameters and results';
COMMENT ON COLUMN coderuns.id IS 'Primary key (PostgreSQL native UUID)';
COMMENT ON COLUMN coderuns.mongo_object_id IS 'MongoDB ObjectID for backward compatibility';
COMMENT ON COLUMN coderuns.code_id IS 'Reference to the code being executed (PostgreSQL UUID)';
COMMENT ON COLUMN coderuns.code_mongo_id IS 'Reference to the code (MongoDB ObjectID)';
COMMENT ON COLUMN coderuns.status IS 'Execution status: queued, started, completed, or failed';
COMMENT ON COLUMN coderuns.extra IS 'Extra metadata (e.g., status_code, content_type)';
COMMENT ON COLUMN coderuns.params IS 'Execution parameters';
COMMENT ON COLUMN coderuns.headers IS 'HTTP headers for endpoint executions';
