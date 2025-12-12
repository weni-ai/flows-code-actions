-- PostgreSQL schema for codelibs table
-- This table stores code libraries with UUID as primary key

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS codelibs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    language VARCHAR(50) NOT NULL CHECK (language IN ('python')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_codelibs_name ON codelibs(name);
CREATE INDEX IF NOT EXISTS idx_codelibs_language ON codelibs(language);
CREATE INDEX IF NOT EXISTS idx_codelibs_name_language ON codelibs(name, language);
CREATE INDEX IF NOT EXISTS idx_codelibs_created_at ON codelibs(created_at);

-- Unique constraint to prevent duplicate libraries
CREATE UNIQUE INDEX IF NOT EXISTS idx_codelibs_unique_name_language ON codelibs(name, language);
