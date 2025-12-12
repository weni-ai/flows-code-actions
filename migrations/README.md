# Database Migrations

This directory contains PostgreSQL migrations in the standard [golang-migrate/migrate](https://github.com/golang-migrate/migrate) format.

## Files

```
migrations/
├── 000001_create_codes_table.up.sql      # Create codes table
├── 000001_create_codes_table.down.sql    # Drop codes table
├── 000002_create_codelibs_table.up.sql   # Create codelibs table
├── 000002_create_codelibs_table.down.sql # Drop codelibs table
└── README.md
```

## How to Use

### CLI Installation

```bash
# Via Go
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Via Homebrew (macOS)
brew install golang-migrate

# Via Docker
docker pull migrate/migrate
```

### Basic Commands

```bash
# Apply all migrations
migrate -path ./migrations -database "postgres://localhost:5432/codeactions?sslmode=disable" up

# Rollback one migration
migrate -path ./migrations -database "postgres://localhost:5432/codeactions?sslmode=disable" down 1

# Check current version
migrate -path ./migrations -database "postgres://localhost:5432/codeactions?sslmode=disable" version

# Apply N specific migrations
migrate -path ./migrations -database "postgres://localhost:5432/codeactions?sslmode=disable" up 2
```

### Via Docker

```bash
# Apply migrations
docker run --rm -v $(pwd)/migrations:/migrations --network host \
  migrate/migrate -path=/migrations/ \
  -database "postgres://localhost:5432/codeactions?sslmode=disable" up

# Check status
docker run --rm -v $(pwd)/migrations:/migrations --network host \
  migrate/migrate -path=/migrations/ \
  -database "postgres://localhost:5432/codeactions?sslmode=disable" version
```

### Create New Migration

```bash
# Create new migration
migrate create -ext sql -dir ./migrations -seq add_new_table

# This will create:
# migrations/000003_add_new_table.up.sql
# migrations/000003_add_new_table.down.sql
```

## Created Tables

### 1. `codes` Table
Stores code actions (flows and endpoints) with their metadata.

**Fields:**
- `id` (UUID) - Primary key
- `name` (VARCHAR) - Code name
- `type` (VARCHAR) - Type: 'flow' or 'endpoint' 
- `source` (TEXT) - Source code
- `language` (VARCHAR) - Language: 'python', 'go', 'javascript'
- `url` (VARCHAR) - URL (for endpoints)
- `project_uuid` (VARCHAR) - Project UUID
- `timeout` (INTEGER) - Execution timeout (5-300s)
- `created_at`, `updated_at` (TIMESTAMP)

**Indexes:**
- `idx_codes_project_uuid` - By project
- `idx_codes_type` - By type
- `idx_codes_project_type` - By project and type
- `idx_codes_created_at` - By creation date

### 2. `codelibs` Table
Stores available code libraries (e.g., Python packages).

**Fields:**
- `id` (UUID) - Primary key
- `name` (VARCHAR) - Library name
- `language` (VARCHAR) - Language: 'python'
- `created_at`, `updated_at` (TIMESTAMP)

**Indexes:**
- `idx_codelibs_name` - By name
- `idx_codelibs_language` - By language
- `idx_codelibs_name_language` - By name and language
- `idx_codelibs_unique_name_language` - Unique constraint

## Usage with Environment Variable

```bash
# Set database URL
export DATABASE_URL="postgres://localhost:5432/codeactions?sslmode=disable"

# Use in commands
migrate -path ./migrations -database $DATABASE_URL up
migrate -path ./migrations -database $DATABASE_URL version
```

## Troubleshooting

### Migration in "Dirty" State
```bash
# Check state
migrate -path ./migrations -database $DATABASE_URL version

# If dirty, force previous version  
migrate -path ./migrations -database $DATABASE_URL force 1

# Try again
migrate -path ./migrations -database $DATABASE_URL up
```

### Backup Before Migrations
```bash
# Create backup
pg_dump $DATABASE_URL > backup_$(date +%Y%m%d_%H%M%S).sql

# Apply migrations
migrate -path ./migrations -database $DATABASE_URL up

# If necessary, restore backup
psql $DATABASE_URL < backup_20231201_120000.sql
```