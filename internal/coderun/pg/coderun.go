package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/coderun"
	"go.mongodb.org/mongo-driver/bson/primitive"

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
	query := `
		INSERT INTO coderuns (code_id, status, result, extra, params, body, headers, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	cr.CreatedAt = time.Now()
	cr.UpdatedAt = time.Now()

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
		cr.CodeID.Hex(),
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

	// Convert string UUID to ObjectID for compatibility
	if oid, err := primitive.ObjectIDFromHex(id); err == nil {
		cr.ID = oid
	} else {
		cr.ID = primitive.NewObjectID()
	}

	return cr, nil
}

func (r *codeRunRepo) GetByID(ctx context.Context, id string) (*coderun.CodeRun, error) {
	query := `
		SELECT id, code_id, status, result, extra, params, body, headers, created_at, updated_at
		FROM coderuns
		WHERE id = $1`

	cr := &coderun.CodeRun{}
	var dbID, codeID string
	var extraJSON, paramsJSON, headersJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&dbID,
		&codeID,
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

	// Convert UUIDs to ObjectIDs
	if oid, err := primitive.ObjectIDFromHex(dbID); err == nil {
		cr.ID = oid
	} else {
		cr.ID = primitive.NewObjectID()
	}

	if oid, err := primitive.ObjectIDFromHex(codeID); err == nil {
		cr.CodeID = oid
	} else {
		cr.CodeID = primitive.NewObjectID()
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
	query := `
		SELECT id, code_id, status, result, extra, params, body, headers, created_at, updated_at
		FROM coderuns
		WHERE code_id = $1`

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
		var dbID, dbCodeID string
		var extraJSON, paramsJSON, headersJSON []byte

		err := rows.Scan(
			&dbID,
			&dbCodeID,
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

		// Convert UUIDs to ObjectIDs
		if oid, err := primitive.ObjectIDFromHex(dbID); err == nil {
			cr.ID = oid
		} else {
			cr.ID = primitive.NewObjectID()
		}

		if oid, err := primitive.ObjectIDFromHex(dbCodeID); err == nil {
			cr.CodeID = oid
		} else {
			cr.CodeID = primitive.NewObjectID()
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
	query := `
		UPDATE coderuns
		SET code_id = $2, status = $3, result = $4, extra = $5, params = $6, 
		    body = $7, headers = $8, updated_at = $9
		WHERE id = $1`

	cr.UpdatedAt = time.Now()

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

	result, err := r.db.ExecContext(ctx, query,
		id,
		cr.CodeID.Hex(),
		cr.Status,
		cr.Result,
		extraJSON,
		paramsJSON,
		cr.Body,
		headersJSON,
		cr.UpdatedAt,
	)

	if err != nil {
		return nil, errors.Wrap(err, "error updating coderun")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, "error checking affected rows")
	}

	if rowsAffected == 0 {
		return nil, errors.New("coderun not found")
	}

	// Convert string UUID to ObjectID
	if oid, err := primitive.ObjectIDFromHex(id); err == nil {
		cr.ID = oid
	} else {
		cr.ID = primitive.NewObjectID()
	}

	return cr, nil
}

func (r *codeRunRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM coderuns WHERE id = $1`

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
