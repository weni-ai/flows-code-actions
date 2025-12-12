package pg

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/codelib"
	"go.mongodb.org/mongo-driver/bson/primitive"

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
		INSERT INTO codelibs (name, language, created_at, updated_at) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id`

	cl.CreatedAt = time.Now()
	cl.UpdatedAt = time.Now()

	var id string
	err := r.db.QueryRowContext(ctx, query,
		cl.Name,
		cl.Language,
		cl.CreatedAt,
		cl.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return nil, errors.Wrap(err, "error creating codelib")
	}

	// Convert string UUID to ObjectID for compatibility
	if oid, err := primitive.ObjectIDFromHex(id); err == nil {
		cl.ID = oid
	} else {
		// If can't convert, create a new ObjectID
		cl.ID = primitive.NewObjectID()
	}

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
		INSERT INTO codelibs (name, language, created_at, updated_at) 
		VALUES ($1, $2, $3, $4) 
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
			cl.Name,
			cl.Language,
			cl.CreatedAt,
			cl.UpdatedAt,
		).Scan(&id)

		if err != nil {
			return nil, errors.Wrap(err, "error executing bulk insert")
		}

		// Convert string UUID to ObjectID for compatibility
		if oid, err := primitive.ObjectIDFromHex(id); err == nil {
			cl.ID = oid
		} else {
			// If can't convert, create a new ObjectID
			cl.ID = primitive.NewObjectID()
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "error committing bulk insert transaction")
	}

	return cls, nil
}

func (r *codelibRepo) List(ctx context.Context, lang *codelib.LanguageType) ([]codelib.CodeLib, error) {
	query := `
		SELECT id, name, language, created_at, updated_at 
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
		var dbID string
		err := rows.Scan(
			&dbID,
			&cl.Name,
			&cl.Language,
			&cl.CreatedAt,
			&cl.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error scanning codelib row")
		}

		// Convert string UUID to ObjectID for compatibility
		if oid, err := primitive.ObjectIDFromHex(dbID); err == nil {
			cl.ID = oid
		} else {
			// If can't convert, create a new ObjectID
			cl.ID = primitive.NewObjectID()
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
		SELECT id, name, language, created_at, updated_at 
		FROM codelibs 
		WHERE name = $1`

	args := []interface{}{name}

	// Add language filter if provided
	if lang != nil {
		query += " AND language = $2"
		args = append(args, string(*lang))
	}

	cl := &codelib.CodeLib{}
	var dbID string
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&dbID,
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

	// Convert string UUID to ObjectID for compatibility
	if oid, err := primitive.ObjectIDFromHex(dbID); err == nil {
		cl.ID = oid
	} else {
		// If can't convert, create a new ObjectID
		cl.ID = primitive.NewObjectID()
	}

	return cl, nil
}
