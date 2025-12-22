-- Create projects table
-- Migration: 000005_create_projects_table

CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    uuid VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    authorizations JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_projects_uuid ON projects(uuid);
CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name);
CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at);

-- Index for JSONB field (GIN index for efficient JSON queries)
CREATE INDEX IF NOT EXISTS idx_projects_authorizations ON projects USING GIN (authorizations);

-- Add comments for documentation
COMMENT ON TABLE projects IS 'Stores project information with user authorizations';
COMMENT ON COLUMN projects.uuid IS 'Project unique identifier (business UUID)';
COMMENT ON COLUMN projects.name IS 'Project name';
COMMENT ON COLUMN projects.authorizations IS 'Array of user authorizations with email and role';
COMMENT ON INDEX idx_projects_uuid IS 'Unique constraint and index on project UUID';
