package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/coderun"

	_ "github.com/lib/pq"
)

type codeRunRepo struct {
	db *sql.DB
}

// NewCodeRunRepository creates a new PostgreSQL repository for coderun entities
func NewCodeRunRepository(db *sql.DB) coderun.Repository {
	return &codeRunRepo{db: db}
}

func (r *codeRunRepo) Create(ctx context.Context, cr *coderun.CodeRun) (*coderun.CodeRun, error) {
	cr.CreatedAt = time.Now()
	cr.UpdatedAt = time.Now()

	codeUUID := cr.CodeID
	if cr.CodeID != "" && !isUUID(cr.CodeID) {
		var foundUUID sql.NullString
		lookupQuery := `SELECT id FROM codes WHERE mongo_object_id = $1`
		err := r.db.QueryRowContext(ctx, lookupQuery, cr.CodeID).Scan(&foundUUID)

		if err != nil && err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "error looking up code UUID")
		}

		if foundUUID.Valid {
			cr.CodeMongoID = cr.CodeID
			codeUUID = foundUUID.String
		} else {
			codeUUID = ""
		}
	}

	query := `
		INSERT INTO coderuns (mongo_object_id, code_id, code_mongo_id, status, result, extra, params, body, headers, created_at, updated_at)
		VALUES ($1, NULLIF($2, '')::uuid, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`

	// Marshal JSON fields
	extraJSON, err := json.Marshal(cr.Extra)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling extra")
	}

	paramsJSON, err := json.Marshal(cr.Params)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling params")
	}

	headersJSON, err := json.Marshal(cr.Headers)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling headers")
	}

	var id string
	err = r.db.QueryRowContext(ctx, query,
		nullString(cr.MongoObjectID),
		codeUUID,
		nullString(cr.CodeMongoID),
		cr.Status,
		cr.Result,
		extraJSON,
		paramsJSON,
		cr.Body,
		headersJSON,
		cr.CreatedAt,
		cr.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return nil, errors.Wrap(err, "error creating coderun")
	}

	cr.ID = id
	return cr, nil
}

func (r *codeRunRepo) GetByID(ctx context.Context, id string) (*coderun.CodeRun, error) {
	// Try to find by UUID first, then by mongo_object_id
	query := `
		SELECT id, mongo_object_id, code_id, code_mongo_id, status, result, extra, params, body, headers, created_at, updated_at
		FROM coderuns
		WHERE id::text = $1 OR mongo_object_id = $1`

	cr := &coderun.CodeRun{}
	var mongoObjectID, codeID, codeMongoID sql.NullString
	var extraJSON, paramsJSON, headersJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cr.ID,
		&mongoObjectID,
		&codeID,
		&codeMongoID,
		&cr.Status,
		&cr.Result,
		&extraJSON,
		&paramsJSON,
		&cr.Body,
		&headersJSON,
		&cr.CreatedAt,
		&cr.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("coderun not found")
		}
		return nil, errors.Wrap(err, "error getting coderun by id")
	}

	if mongoObjectID.Valid {
		cr.MongoObjectID = mongoObjectID.String
	}
	if codeID.Valid {
		cr.CodeID = codeID.String
	}
	if codeMongoID.Valid {
		cr.CodeMongoID = codeMongoID.String
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(extraJSON, &cr.Extra); err != nil {
		cr.Extra = make(map[string]interface{})
	}

	if err := json.Unmarshal(paramsJSON, &cr.Params); err != nil {
		cr.Params = make(map[string]interface{})
	}

	if err := json.Unmarshal(headersJSON, &cr.Headers); err != nil {
		cr.Headers = make(map[string]interface{})
	}

	return cr, nil
}

func (r *codeRunRepo) ListByCodeID(ctx context.Context, codeID string, filter map[string]interface{}) ([]coderun.CodeRun, error) {
	// Search by code_id (UUID) or code_mongo_id (MongoDB ObjectID)
	// Use explicit casting for UUID comparison
	query := `
		SELECT id, mongo_object_id, code_id, code_mongo_id, status, result, extra, params, body, headers, created_at, updated_at
		FROM coderuns
		WHERE (code_id::text = $1 OR code_mongo_id = $1)`

	args := []interface{}{codeID}
	argIndex := 2

	// Add date filters
	if after, ok := filter["after"].(time.Time); ok {
		query += " AND created_at >= $2"
		args = append(args, after)
		argIndex++
	}

	if before, ok := filter["before"].(time.Time); ok {
		if argIndex == 2 {
			query += " AND created_at <= $2"
		} else {
			query += " AND created_at <= $3"
		}
		args = append(args, before)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "error listing coderuns by code id")
	}
	defer rows.Close()

	var coderuns []coderun.CodeRun
	for rows.Next() {
		cr := coderun.CodeRun{}
		var mongoObjectID, dbCodeID, codeMongoID sql.NullString
		var extraJSON, paramsJSON, headersJSON []byte

		err := rows.Scan(
			&cr.ID,
			&mongoObjectID,
			&dbCodeID,
			&codeMongoID,
			&cr.Status,
			&cr.Result,
			&extraJSON,
			&paramsJSON,
			&cr.Body,
			&headersJSON,
			&cr.CreatedAt,
			&cr.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error scanning coderun row")
		}

		if mongoObjectID.Valid {
			cr.MongoObjectID = mongoObjectID.String
		}
		if dbCodeID.Valid {
			cr.CodeID = dbCodeID.String
		}
		if codeMongoID.Valid {
			cr.CodeMongoID = codeMongoID.String
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(extraJSON, &cr.Extra); err != nil {
			cr.Extra = make(map[string]interface{})
		}

		if err := json.Unmarshal(paramsJSON, &cr.Params); err != nil {
			cr.Params = make(map[string]interface{})
		}

		if err := json.Unmarshal(headersJSON, &cr.Headers); err != nil {
			cr.Headers = make(map[string]interface{})
		}

		coderuns = append(coderuns, cr)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating coderun rows")
	}

	return coderuns, nil
}

func (r *codeRunRepo) Update(ctx context.Context, id string, cr *coderun.CodeRun) (*coderun.CodeRun, error) {
	cr.UpdatedAt = time.Now()

	codeUUID := cr.CodeID
	if cr.CodeID != "" && !isUUID(cr.CodeID) {
		var foundUUID sql.NullString
		lookupQuery := `SELECT id FROM codes WHERE mongo_object_id = $1`
		err := r.db.QueryRowContext(ctx, lookupQuery, cr.CodeID).Scan(&foundUUID)

		if err != nil && err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "error looking up code UUID")
		}

		if foundUUID.Valid {
			cr.CodeMongoID = cr.CodeID
			codeUUID = foundUUID.String
		} else {
			codeUUID = ""
		}
	}

	query := `
		UPDATE coderuns
		SET mongo_object_id = $2, code_id = NULLIF($3, '')::uuid, code_mongo_id = $4, status = $5, result = $6, 
		    extra = $7, params = $8, body = $9, headers = $10, updated_at = $11
		WHERE id::text = $1 OR mongo_object_id = $1
		RETURNING id`

	// Marshal JSON fields
	extraJSON, err := json.Marshal(cr.Extra)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling extra")
	}

	paramsJSON, err := json.Marshal(cr.Params)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling params")
	}

	headersJSON, err := json.Marshal(cr.Headers)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling headers")
	}

	var returnedID string
	err = r.db.QueryRowContext(ctx, query,
		id,
		nullString(cr.MongoObjectID),
		codeUUID,
		nullString(cr.CodeMongoID),
		cr.Status,
		cr.Result,
		extraJSON,
		paramsJSON,
		cr.Body,
		headersJSON,
		cr.UpdatedAt,
	).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("coderun not found")
		}
		return nil, errors.Wrap(err, "error updating coderun")
	}

	cr.ID = returnedID
	return cr, nil
}

func (r *codeRunRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM coderuns WHERE id::text = $1 OR mongo_object_id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "error deleting coderun")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "error checking affected rows")
	}

	if rowsAffected == 0 {
		return errors.New("coderun not found")
	}

	return nil
}

// nullString converts an empty string to sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// isUUID checks if a string is a valid UUID format
// UUID format: 8-4-4-4-12 hex characters (e.g., 550e8400-e29b-41d4-a716-446655440000)
func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	// Check format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return false
	}
	// Check if all other characters are hex digits
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			continue
		}
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func (r *codeRunRepo) DeleteOlder(ctx context.Context, date time.Time, limit int64) (int64, error) {
	query := `
		DELETE FROM coderuns
		WHERE id IN (
			SELECT id FROM coderuns
			WHERE created_at < $1
			ORDER BY created_at ASC
			LIMIT $2
		)`

	result, err := r.db.ExecContext(ctx, query, date, limit)
	if err != nil {
		return 0, errors.Wrap(err, "error deleting older coderuns")
	}

	deletedCount, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "error getting deleted count")
	}

	return deletedCount, nil
}
