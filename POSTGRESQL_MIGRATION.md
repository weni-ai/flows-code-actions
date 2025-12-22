# PostgreSQL Migration Guide

This guide explains how to use PostgreSQL instead of MongoDB in the Code Actions application.

## Overview

The application now supports both MongoDB and PostgreSQL databases. You can choose which database to use by setting the `DATABASE_TYPE` environment variable.

## Configuration

### Environment Variables

```bash
# Database Type (mongodb or postgres)
DATABASE_TYPE=postgres

# PostgreSQL Connection
DATABASE_URI=postgres://user:password@localhost:5432/codeactions?sslmode=disable
DATABASE_NAME=codeactions

# Connection Pool Settings
DATABASE_MAX_POOL_SIZE=100
DATABASE_MIN_POOL_SIZE=10
DATABASE_TIMEOUT=10
DATABASE_MAX_RETRIES=3
DATABASE_CONNECT_TIMEOUT_SECONDS=10
```

### Using MongoDB (Default)

```bash
DATABASE_TYPE=mongodb
DATABASE_URI=mongodb://localhost:27017
DATABASE_NAME=flow-code-actions
```

### Using PostgreSQL

```bash
DATABASE_TYPE=postgres
DATABASE_URI=postgres://localhost:5432/codeactions?sslmode=disable
DATABASE_NAME=codeactions
```

## Database Setup

### 1. Run Migrations

Using the official migrate CLI:

```bash
# Install migrate CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path ./migrations -database "postgres://localhost:5432/codeactions?sslmode=disable" up
```

Using Docker:

```bash
docker run --rm -v $(pwd)/migrations:/migrations --network host \
  migrate/migrate -path=/migrations/ \
  -database "postgres://localhost:5432/codeactions?sslmode=disable" up
```

### 2. Verify Tables

```bash
psql postgres://localhost:5432/codeactions -c "\dt"
```

You should see:
- `codes` - Code actions
- `codelibs` - Code libraries
- `coderuns` - Code executions
- `user_permissions` - User permissions
- `projects` - Projects

## Implemented Repositories

### ✅ PostgreSQL Repositories Available

1. **Code** (`internal/code/pg`)
   - Create, Read, Update, Delete code actions
   - List by project UUID with type filtering

2. **CodeLib** (`internal/codelib/pg`)
   - Manage Python libraries
   - Bulk operations support

3. **CodeRun** (`internal/coderun/pg`)
   - Track code executions
   - Filter by date ranges
   - Cleanup old runs

4. **Permission** (`internal/permission/pg`)
   - User permissions management
   - Role-based access control

5. **Project** (`internal/project/pg`)
   - Project management
   - User authorizations

### ⚠️ MongoDB Only (Not Yet Implemented for PostgreSQL)

- **CodeLog** (`internal/codelog`) - Still uses MongoDB

## Running the Application

### With PostgreSQL

```bash
# Set environment variables
export DATABASE_TYPE=postgres
export DATABASE_URI="postgres://localhost:5432/codeactions?sslmode=disable"

# Run application
go run cmd/codeactions/main.go
```

### With Docker Compose

```yaml
version: '3.8'
services:
  app:
    build: .
    environment:
      - DATABASE_TYPE=postgres
      - DATABASE_URI=postgres://postgres:password@db:5432/codeactions?sslmode=disable
      - DATABASE_NAME=codeactions
    depends_on:
      - db
  
  db:
    image: postgres:15
    environment:
      - POSTGRES_DB=codeactions
      - POSTGRES_PASSWORD=password
    ports:
      - "5432:5432"
```

## Migration from MongoDB to PostgreSQL

### Step 1: Export Data from MongoDB

```javascript
// Export codes
mongoexport --db=flow-code-actions --collection=code --out=codes.json

// Export codelibs
mongoexport --db=flow-code-actions --collection=codelib --out=codelibs.json

// Export coderuns
mongoexport --db=flow-code-actions --collection=coderun --out=coderuns.json

// Export permissions
mongoexport --db=flow-code-actions --collection=user_permissions --out=permissions.json

// Export projects
mongoexport --db=flow-code-actions --collection=project --out=projects.json
```

### Step 2: Transform and Import to PostgreSQL

Create a migration script to:
1. Parse JSON exports
2. Transform ObjectIDs to UUIDs
3. Insert into PostgreSQL tables

Example transformation script (Python):

```python
import json
import psycopg2
from uuid import uuid4

# Connect to PostgreSQL
conn = psycopg2.connect("postgres://localhost:5432/codeactions")
cur = conn.cursor()

# Load MongoDB export
with open('codes.json') as f:
    codes = [json.loads(line) for line in f]

# Transform and insert
for code in codes:
    cur.execute("""
        INSERT INTO codes (id, name, type, source, language, url, project_uuid, timeout, created_at, updated_at)
        VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
    """, (
        str(uuid4()),  # Generate new UUID
        code['name'],
        code['type'],
        code['source'],
        code['language'],
        code.get('url', ''),
        code['project_uuid'],
        code.get('timeout', 60),
        code['created_at'],
        code['updated_at']
    ))

conn.commit()
```

## Testing

### Test PostgreSQL Connection

```bash
# Test with psql
psql "postgres://localhost:5432/codeactions?sslmode=disable" -c "SELECT version();"
```

### Run Application Tests

```bash
# Set test database
export DATABASE_TYPE=postgres
export DATABASE_URI="postgres://localhost:5432/codeactions_test?sslmode=disable"

# Run tests
go test ./...
```

## Troubleshooting

### Connection Issues

```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Check connection
psql "postgres://localhost:5432/codeactions?sslmode=disable"
```

### Migration Issues

```bash
# Check migration status
migrate -path ./migrations -database "postgres://..." version

# Force migration version (if needed)
migrate -path ./migrations -database "postgres://..." force 1
```

### Application Logs

The application logs will indicate which database type is being used:

```
INFO Using PostgreSQL database
```

or

```
INFO Using MongoDB database
```

## Performance Considerations

### PostgreSQL Advantages
- JSONB indexes for fast JSON queries
- Better transaction support
- More mature backup/restore tools
- Better for relational data

### Indexes

All tables have appropriate indexes:
- Primary key indexes (UUID)
- Foreign key indexes
- GIN indexes for JSONB fields
- Composite indexes for common queries

## Backup and Restore

### PostgreSQL Backup

```bash
# Full backup
pg_dump postgres://localhost:5432/codeactions > backup.sql

# Backup with compression
pg_dump postgres://localhost:5432/codeactions | gzip > backup.sql.gz
```

### PostgreSQL Restore

```bash
# Restore from backup
psql postgres://localhost:5432/codeactions < backup.sql

# Restore from compressed backup
gunzip -c backup.sql.gz | psql postgres://localhost:5432/codeactions
```

## Support

For issues or questions:
1. Check application logs
2. Verify database connection
3. Check migration status
4. Review this documentation
