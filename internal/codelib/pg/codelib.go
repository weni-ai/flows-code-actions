package pg

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/codelib"

	_ "github.com/lib/pq"
)

type codelibRepo struct {
	db *sql.DB
}

// NewCodeLibRepo creates a new PostgreSQL repository for codelib entities
func NewCodeLibRepo(db *sql.DB) codelib.Repository {
	return &codelibRepo{db: db}
}

func (r *codelibRepo) Create(ctx context.Context, cl *codelib.CodeLib) (*codelib.CodeLib, error) {
	query := `
		INSERT INTO codelibs (mongo_object_id, name, language, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id`

	cl.CreatedAt = time.Now()
	cl.UpdatedAt = time.Now()

	var id string
	err := r.db.QueryRowContext(ctx, query,
		nullString(cl.MongoObjectID),
		cl.Name,
		cl.Language,
		cl.CreatedAt,
		cl.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return nil, errors.Wrap(err, "error creating codelib")
	}

	cl.ID = id
	return cl, nil
}

func (r *codelibRepo) CreateBulk(ctx context.Context, cls []*codelib.CodeLib) ([]*codelib.CodeLib, error) {
	if len(cls) == 0 {
		return cls, nil
	}

	// Start transaction for bulk insert
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error starting transaction")
	}
	defer tx.Rollback()

	// Prepare statement for bulk insert
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO codelibs (mongo_object_id, name, language, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id`)
	if err != nil {
		return nil, errors.Wrap(err, "error preparing bulk insert statement")
	}
	defer stmt.Close()

	now := time.Now()
	for _, cl := range cls {
		cl.CreatedAt = now
		cl.UpdatedAt = now

		var id string
		err := stmt.QueryRowContext(ctx,
			nullString(cl.MongoObjectID),
			cl.Name,
			cl.Language,
			cl.CreatedAt,
			cl.UpdatedAt,
		).Scan(&id)

		if err != nil {
			return nil, errors.Wrap(err, "error executing bulk insert")
		}

		cl.ID = id
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "error committing bulk insert transaction")
	}

	return cls, nil
}

func (r *codelibRepo) List(ctx context.Context, lang *codelib.LanguageType) ([]codelib.CodeLib, error) {
	query := `
		SELECT id, mongo_object_id, name, language, created_at, updated_at 
		FROM codelibs`

	args := []interface{}{}

	// Add language filter if provided
	if lang != nil {
		query += " WHERE language = $1"
		args = append(args, string(*lang))
	}

	query += " ORDER BY name ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "error listing codelibs")
	}
	defer rows.Close()

	var libs []codelib.CodeLib
	for rows.Next() {
		var cl codelib.CodeLib
		var mongoObjectID sql.NullString
		
		err := rows.Scan(
			&cl.ID,
			&mongoObjectID,
			&cl.Name,
			&cl.Language,
			&cl.CreatedAt,
			&cl.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error scanning codelib row")
		}

		if mongoObjectID.Valid {
			cl.MongoObjectID = mongoObjectID.String
		}

		libs = append(libs, cl)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating codelib rows")
	}

	return libs, nil
}

func (r *codelibRepo) Find(ctx context.Context, name string, lang *codelib.LanguageType) (*codelib.CodeLib, error) {
	query := `
		SELECT id, mongo_object_id, name, language, created_at, updated_at 
		FROM codelibs 
		WHERE name = $1`

	args := []interface{}{name}

	// Add language filter if provided
	if lang != nil {
		query += " AND language = $2"
		args = append(args, string(*lang))
	}

	cl := &codelib.CodeLib{}
	var mongoObjectID sql.NullString
	
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&cl.ID,
		&mongoObjectID,
		&cl.Name,
		&cl.Language,
		&cl.CreatedAt,
		&cl.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("codelib not found")
		}
		return nil, errors.Wrap(err, "error finding codelib")
	}

	if mongoObjectID.Valid {
		cl.MongoObjectID = mongoObjectID.String
	}

	return cl, nil
}

// nullString converts an empty string to sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
