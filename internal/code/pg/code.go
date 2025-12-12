package code

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/code"
	"go.mongodb.org/mongo-driver/bson/primitive"

	_ "github.com/lib/pq"
)

type codeRepo struct {
	db *sql.DB
}

// NewCodeRepository creates a new PostgreSQL repository for code entities
func NewCodeRepository(db *sql.DB) code.Repository {
	return &codeRepo{db: db}
}

func (r *codeRepo) Create(ctx context.Context, codeAction *code.Code) (*code.Code, error) {
	query := `
		INSERT INTO codes (name, type, source, language, url, project_uuid, timeout, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		RETURNING id`

	codeAction.CreatedAt = time.Now()
	codeAction.UpdatedAt = time.Now()

	var id string
	err := r.db.QueryRowContext(ctx, query,
		codeAction.Name,
		codeAction.Type,
		codeAction.Source,
		codeAction.Language,
		codeAction.URL,
		codeAction.ProjectUUID,
		codeAction.Timeout,
		codeAction.CreatedAt,
		codeAction.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return nil, errors.Wrap(err, "error creating code")
	}

	// Convert string UUID to ObjectID for compatibility
	if oid, err := primitive.ObjectIDFromHex(id); err == nil {
		codeAction.ID = oid
	} else {
		// If can't convert, create a new ObjectID and store the UUID as hex
		codeAction.ID = primitive.NewObjectID()
	}
	return codeAction, nil
}

func (r *codeRepo) GetByID(ctx context.Context, id string) (*code.Code, error) {
	query := `
		SELECT id, name, type, source, language, url, project_uuid, timeout, created_at, updated_at 
		FROM codes 
		WHERE id = $1`

	codeAction := &code.Code{}
	var dbID string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&dbID,
		&codeAction.Name,
		&codeAction.Type,
		&codeAction.Source,
		&codeAction.Language,
		&codeAction.URL,
		&codeAction.ProjectUUID,
		&codeAction.Timeout,
		&codeAction.CreatedAt,
		&codeAction.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("code not found")
		}
		return nil, errors.Wrap(err, "error getting code by id")
	}

	// Convert string UUID to ObjectID for compatibility
	if oid, err := primitive.ObjectIDFromHex(dbID); err == nil {
		codeAction.ID = oid
	} else {
		// If can't convert, create a new ObjectID
		codeAction.ID = primitive.NewObjectID()
	}

	// Set default timeout if not set
	if codeAction.Timeout == 0 {
		codeAction.Timeout = 60
	}

	return codeAction, nil
}

func (r *codeRepo) ListByProjectUUID(ctx context.Context, projectUUID string, codeType string) ([]code.Code, error) {
	query := `
		SELECT id, name, type, source, language, url, project_uuid, timeout, created_at, updated_at 
		FROM codes 
		WHERE project_uuid = $1`

	args := []interface{}{projectUUID}

	// Add code type filter if provided
	if codeType != "" {
		ct := code.CodeType(codeType)
		if err := ct.Validate(); err != nil {
			return nil, err
		}
		query += " AND type = $2"
		args = append(args, codeType)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "error listing codes by project uuid")
	}
	defer rows.Close()

	var codes []code.Code
	for rows.Next() {
		var c code.Code
		var dbID string
		err := rows.Scan(
			&dbID,
			&c.Name,
			&c.Type,
			&c.Source,
			&c.Language,
			&c.URL,
			&c.ProjectUUID,
			&c.Timeout,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error scanning code row")
		}

		// Convert string UUID to ObjectID for compatibility
		if oid, err := primitive.ObjectIDFromHex(dbID); err == nil {
			c.ID = oid
		} else {
			// If can't convert, create a new ObjectID
			c.ID = primitive.NewObjectID()
		}

		// Set default timeout if not set
		if c.Timeout == 0 {
			c.Timeout = 60
		}

		codes = append(codes, c)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating code rows")
	}

	return codes, nil
}

func (r *codeRepo) Update(ctx context.Context, id string, codeAction *code.Code) (*code.Code, error) {
	query := `
		UPDATE codes 
		SET name = $2, type = $3, source = $4, language = $5, url = $6, 
		    project_uuid = $7, timeout = $8, updated_at = $9
		WHERE id = $1`

	codeAction.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		id,
		codeAction.Name,
		codeAction.Type,
		codeAction.Source,
		codeAction.Language,
		codeAction.URL,
		codeAction.ProjectUUID,
		codeAction.Timeout,
		codeAction.UpdatedAt,
	)

	if err != nil {
		return nil, errors.Wrap(err, "error updating code")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, "error checking affected rows")
	}

	if rowsAffected == 0 {
		return nil, errors.New("code not found")
	}

	// Convert string UUID to ObjectID for compatibility
	if oid, err := primitive.ObjectIDFromHex(id); err == nil {
		codeAction.ID = oid
	} else {
		// If can't convert, create a new ObjectID
		codeAction.ID = primitive.NewObjectID()
	}
	return codeAction, nil
}

func (r *codeRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM codes WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "error deleting code")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "error checking affected rows")
	}

	if rowsAffected == 0 {
		return errors.New("code not found")
	}

	return nil
}
