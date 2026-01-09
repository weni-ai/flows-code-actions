-- Create projects table (postgres-first strategy)
-- Stores project information with authorizations

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mongo_object_id TEXT UNIQUE,
    uuid VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    authorizations JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_projects_mongo_object_id ON projects(mongo_object_id);
CREATE INDEX IF NOT EXISTS idx_projects_uuid ON projects(uuid);
CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name);
CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at);

-- Index for JSONB field (GIN index for efficient JSON queries)
CREATE INDEX IF NOT EXISTS idx_projects_authorizations ON projects USING GIN (authorizations);

-- Add comments for documentation
COMMENT ON TABLE projects IS 'Stores project information with user authorizations';
COMMENT ON COLUMN projects.id IS 'Primary key (PostgreSQL native UUID)';
COMMENT ON COLUMN projects.mongo_object_id IS 'MongoDB ObjectID for backward compatibility';
COMMENT ON COLUMN projects.uuid IS 'Project unique identifier (business UUID)';
COMMENT ON COLUMN projects.name IS 'Project name';
COMMENT ON COLUMN projects.authorizations IS 'Array of user authorizations with email and role';
COMMENT ON INDEX idx_projects_uuid IS 'Unique constraint and index on project UUID';
