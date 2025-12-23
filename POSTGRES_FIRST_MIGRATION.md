# PostgreSQL-First Migration Strategy

## Overview

This document describes the migration strategy implemented to transition the codeactions project from MongoDB-centric to a **PostgreSQL-first** architecture, while maintaining backward compatibility with MongoDB.

## Strategy Summary

The migration follows a **postgres-first** approach where:

1. **PostgreSQL UUID** is the primary key (`id UUID PRIMARY KEY DEFAULT gen_random_uuid()`)
2. **MongoDB ObjectID** is stored as a secondary field (`mongo_object_id TEXT UNIQUE`)
3. All models now use `string` for IDs instead of `primitive.ObjectID`
4. Repositories support querying by both PostgreSQL UUID and MongoDB ObjectID

## Database Schema Changes

### Key Changes Across All Tables

All tables now follow this pattern:

```sql
CREATE TABLE table_name (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mongo_object_id TEXT UNIQUE,
    -- other fields...
);

CREATE INDEX idx_table_mongo_object_id ON table_name(mongo_object_id);
```

### Tables Updated

1. **codes** - Code actions (flows and endpoints)
   - Primary key: PostgreSQL UUID
   - Backward compatibility: `mongo_object_id`
   - Migration: `000001_create_codes_table.up.sql`

2. **codelibs** - Code libraries
   - Primary key: PostgreSQL UUID
   - Backward compatibility: `mongo_object_id`
   - Migration: `000002_create_codelibs_table.up.sql`

3. **coderuns** - Code execution runs
   - Primary key: PostgreSQL UUID
   - Backward compatibility: `mongo_object_id`
   - Foreign keys: `code_id` (UUID), `code_mongo_id` (TEXT)
   - Migration: `000003_create_coderuns_table.up.sql`

4. **user_permissions** - User permissions for projects
   - Primary key: PostgreSQL UUID
   - Backward compatibility: `mongo_object_id`
   - Migration: `000004_create_user_permissions_table.up.sql`

5. **projects** - Project information
   - Primary key: PostgreSQL UUID
   - Backward compatibility: `mongo_object_id`
   - Migration: `000005_create_projects_table.up.sql`

## Model Changes

All domain models were updated to use `string` for IDs:

### Before (MongoDB-centric)
```go
type Code struct {
    ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    // ...
}
```

### After (PostgreSQL-first)
```go
type Code struct {
    ID            string `json:"id,omitempty"`                              // PostgreSQL UUID (primary key)
    MongoObjectID string `json:"mongo_object_id,omitempty" bson:"_id,omitempty"` // MongoDB ObjectID for backward compatibility
    // ...
}
```

### Models Updated

- `internal/code/code.go` - Code struct
- `internal/codelib/codelib.go` - CodeLib struct
- `internal/coderun/coderun.go` - CodeRun struct
- `internal/permission/permission.go` - UserPermission struct
- `internal/project/project.go` - Project struct

## Repository Implementation

### PostgreSQL Repositories

All PostgreSQL repositories were updated to:

1. Use `gen_random_uuid()` for generating UUIDs
2. Store `mongo_object_id` for backward compatibility
3. Support querying by both UUID and MongoDB ObjectID
4. Use `sql.NullString` for optional fields

Example from `internal/code/pg/code.go`:

```go
func (r *codeRepo) GetByID(ctx context.Context, id string) (*code.Code, error) {
    // Try to find by UUID first, then by mongo_object_id
    query := `
        SELECT id, mongo_object_id, name, type, source, language, url, project_uuid, timeout, created_at, updated_at 
        FROM codes 
        WHERE id = $1 OR mongo_object_id = $1`
    // ...
}
```

### MongoDB Repositories

MongoDB repositories were updated to:

1. Convert `primitive.ObjectID` to string (hex) when setting IDs
2. Populate both `ID` and `MongoObjectID` fields

Example:

```go
if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
    code.ID = oid.Hex()
    code.MongoObjectID = oid.Hex()
}
```

## Migration Files

All migration files use `gen_random_uuid()` instead of `uuid_generate_v4()`:

```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mongo_object_id TEXT UNIQUE,
    -- ...
);
```

### Migration Files Created/Updated

- `migrations/000001_create_codes_table.up.sql` ✓
- `migrations/000001_create_codes_table.down.sql` ✓
- `migrations/000002_create_codelibs_table.up.sql` ✓
- `migrations/000002_create_codelibs_table.down.sql` ✓
- `migrations/000003_create_coderuns_table.up.sql` ✓
- `migrations/000003_create_coderuns_table.down.sql` ✓
- `migrations/000004_create_user_permissions_table.up.sql` ✓
- `migrations/000004_create_user_permissions_table.down.sql` ✓
- `migrations/000005_create_projects_table.up.sql` ✓
- `migrations/000005_create_projects_table.down.sql` ✓

## Code Changes Summary

### Files Modified

#### Domain Models
- `internal/code/code.go` - Updated Code struct
- `internal/codelib/codelib.go` - Updated CodeLib struct
- `internal/coderun/coderun.go` - Updated CodeRun struct
- `internal/permission/permission.go` - Updated UserPermission struct
- `internal/project/project.go` - Updated Project struct

#### PostgreSQL Repositories
- `internal/code/pg/code.go` - Updated for postgres-first
- `internal/code/pg/schema.sql` - Updated schema
- `internal/codelib/pg/codelib.go` - Updated for postgres-first
- `internal/codelib/pg/schema.sql` - Updated schema
- `internal/coderun/pg/coderun.go` - Updated for postgres-first
- `internal/coderun/pg/schema.sql` - Updated schema
- `internal/permission/pg/user.go` - Updated for postgres-first
- `internal/permission/pg/schema.sql` - Updated schema
- `internal/project/pg/project.go` - Updated for postgres-first
- `internal/project/pg/schema.sql` - Updated schema

#### MongoDB Repositories (Compatibility)
- `internal/code/mongodb/code.go` - Updated to populate both ID fields
- `internal/codelib/mongodb/codelib.go` - Updated to populate both ID fields
- `internal/coderun/mongodb/coderun.go` - Updated to populate both ID fields
- `internal/permission/mongodb/user.go` - Updated to populate both ID fields
- `internal/project/mongodb/project.go` - Updated to populate both ID fields

#### Services and Handlers
- `internal/coderunner/service.go` - Updated to use string IDs
- `internal/permission/consumer.go` - Updated to use string IDs
- `internal/permission/utils.go` - Updated in-memory repo to use UUID
- `internal/http/echo/handlers/code.go` - Updated to use string IDs
- `internal/http/echo/handlers/codelog.go` - Updated to use string IDs

#### Example Files
- All `internal/*/pg/example.go` files - Updated to use string IDs

## Dependencies Added

- `github.com/google/uuid v1.6.0` - For UUID generation in in-memory repositories

## Benefits of This Approach

1. **PostgreSQL-native**: Uses PostgreSQL's native UUID type and generation
2. **Backward compatible**: Maintains MongoDB ObjectID for existing data
3. **Flexible querying**: Can query by either UUID or MongoDB ObjectID
4. **Clean migration path**: Allows gradual migration from MongoDB to PostgreSQL
5. **Type safety**: Uses standard Go `string` type instead of MongoDB-specific types

## Usage

### Configuration

Set the database type in your environment:

```bash
export FLOWS_CODE_ACTIONS_DB_TYPE=postgres
export FLOWS_CODE_ACTIONS_DB_URI="postgres://user:pass@localhost:5432/codeactions?sslmode=disable"
```

### Running Migrations

```bash
migrate -path migrations -database "postgres://user:pass@localhost:5432/codeactions?sslmode=disable" up
```

### Querying by ID

The repositories now support querying by both PostgreSQL UUID and MongoDB ObjectID:

```go
// Query by PostgreSQL UUID
code, err := codeRepo.GetByID(ctx, "550e8400-e29b-41d4-a716-446655440000")

// Query by MongoDB ObjectID (backward compatibility)
code, err := codeRepo.GetByID(ctx, "507f1f77bcf86cd799439011")
```

## Testing

The project successfully compiles with all changes:

```bash
go build ./...
```

All compilation errors have been resolved, and the codebase is ready for testing with PostgreSQL.

## Next Steps

1. Test the migration with real data
2. Verify all CRUD operations work correctly
3. Test backward compatibility with MongoDB ObjectIDs
4. Performance testing with PostgreSQL
5. Update integration tests to cover both database types

## Notes

- The `pgcrypto` extension is used for `gen_random_uuid()` function
- All indexes include `mongo_object_id` for efficient backward compatibility queries
- Foreign key relationships in `coderuns` table support both UUID and MongoDB ObjectID references
- JSONB fields are used for flexible data structures (e.g., `authorizations`, `extra`, `params`)

