-- Create secrets table
-- Migration: 000006_create_secrets_table

CREATE TABLE IF NOT EXISTS secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    code_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_secrets_code_id ON secrets(code_id);
CREATE INDEX IF NOT EXISTS idx_secrets_created_at ON secrets(created_at);

-- Add comments for documentation
COMMENT ON TABLE secrets IS 'Stores secrets associated with code actions';
COMMENT ON COLUMN secrets.id IS 'Primary key (PostgreSQL native UUID)';
COMMENT ON COLUMN secrets.name IS 'Secret name/key';
COMMENT ON COLUMN secrets.value IS 'Secret value (encrypted at application level if needed)';
COMMENT ON COLUMN secrets.code_id IS 'Reference to the associated code action';
