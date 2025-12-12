-- PostgreSQL schema for codes table
-- This table stores code actions with UUID as primary key

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
