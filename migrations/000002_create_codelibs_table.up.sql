-- Create codelibs table
-- Migration: 000002_create_codelibs_table

CREATE TABLE IF NOT EXISTS codelibs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mongo_object_id TEXT UNIQUE,
    name VARCHAR(255) NOT NULL,
    language VARCHAR(50) NOT NULL CHECK (language IN ('python')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_codelibs_mongo_object_id ON codelibs(mongo_object_id);
CREATE INDEX IF NOT EXISTS idx_codelibs_name ON codelibs(name);
CREATE INDEX IF NOT EXISTS idx_codelibs_language ON codelibs(language);
CREATE INDEX IF NOT EXISTS idx_codelibs_name_language ON codelibs(name, language);
CREATE INDEX IF NOT EXISTS idx_codelibs_created_at ON codelibs(created_at);

-- Unique constraint to prevent duplicate libraries
CREATE UNIQUE INDEX IF NOT EXISTS idx_codelibs_unique_name_language ON codelibs(name, language);

-- Add comments for documentation
COMMENT ON TABLE codelibs IS 'Stores code libraries (e.g., Python packages) available for use';
COMMENT ON COLUMN codelibs.id IS 'Primary key (PostgreSQL native UUID)';
COMMENT ON COLUMN codelibs.mongo_object_id IS 'MongoDB ObjectID for backward compatibility';
COMMENT ON COLUMN codelibs.name IS 'Name of the library (e.g., numpy, pandas)';
COMMENT ON COLUMN codelibs.language IS 'Programming language of the library';
COMMENT ON INDEX idx_codelibs_unique_name_language IS 'Prevents duplicate libraries for the same language';
