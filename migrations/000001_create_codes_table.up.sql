-- Create codes table
-- Migration: 000001_create_codes_table

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('flow', 'endpoint')),
    source TEXT NOT NULL,
    language VARCHAR(50) NOT NULL CHECK (language IN ('python', 'go', 'javascript')),
    url VARCHAR(512),
    project_uuid VARCHAR(255) NOT NULL,
    timeout INTEGER NOT NULL DEFAULT 60 CHECK (timeout >= 5 AND timeout <= 300),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_codes_project_uuid ON codes(project_uuid);
CREATE INDEX IF NOT EXISTS idx_codes_type ON codes(type);
CREATE INDEX IF NOT EXISTS idx_codes_project_type ON codes(project_uuid, type);
CREATE INDEX IF NOT EXISTS idx_codes_created_at ON codes(created_at);

-- Add comments for documentation
COMMENT ON TABLE codes IS 'Stores code actions (flows and endpoints) with their metadata';
COMMENT ON COLUMN codes.type IS 'Type of code: flow or endpoint';
COMMENT ON COLUMN codes.language IS 'Programming language: python, go, or javascript';
COMMENT ON COLUMN codes.timeout IS 'Execution timeout in seconds (5-300)';
COMMENT ON COLUMN codes.project_uuid IS 'UUID of the project this code belongs to';
