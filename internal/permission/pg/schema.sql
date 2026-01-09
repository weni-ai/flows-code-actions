-- Create user_permissions table (postgres-first strategy)
-- Stores user permissions for projects

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS user_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mongo_object_id TEXT UNIQUE,
    project_uuid VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    role INTEGER NOT NULL CHECK (role IN (1, 2, 3)),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_user_permissions_mongo_object_id ON user_permissions(mongo_object_id);
CREATE INDEX IF NOT EXISTS idx_user_permissions_project_uuid ON user_permissions(project_uuid);
CREATE INDEX IF NOT EXISTS idx_user_permissions_email ON user_permissions(email);
CREATE INDEX IF NOT EXISTS idx_user_permissions_project_email ON user_permissions(project_uuid, email);

-- Unique constraint to prevent duplicate permissions
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_permissions_unique_project_email ON user_permissions(project_uuid, email);

-- Add comments for documentation
COMMENT ON TABLE user_permissions IS 'Stores user permissions for projects';
COMMENT ON COLUMN user_permissions.id IS 'Primary key (PostgreSQL native UUID)';
COMMENT ON COLUMN user_permissions.mongo_object_id IS 'MongoDB ObjectID for backward compatibility';
COMMENT ON COLUMN user_permissions.project_uuid IS 'UUID of the project';
COMMENT ON COLUMN user_permissions.email IS 'User email address';
COMMENT ON COLUMN user_permissions.role IS 'User role: 1=Viewer, 2=Contributor, 3=Moderator';
COMMENT ON INDEX idx_user_permissions_unique_project_email IS 'Ensures one permission per user per project';
