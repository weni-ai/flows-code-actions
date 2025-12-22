package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/project"
	"go.mongodb.org/mongo-driver/bson/primitive"

	_ "github.com/lib/pq"
)

type repo struct {
	db *sql.DB
}

// NewProjectRepository creates a new PostgreSQL repository for project entities
func NewProjectRepository(db *sql.DB) *repo {
	return &repo{db: db}
}

func (r *repo) Create(ctx context.Context, proj *project.Project) (*project.Project, error) {
	// Check if project already exists
	exists, _ := r.FindByUUID(ctx, proj.UUID)
	if exists != nil {
		return nil, errors.New("project already exists")
	}

	query := `
		INSERT INTO projects (uuid, name, authorizations, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	proj.CreatedAt = time.Now()
	proj.UpdatedAt = time.Now()

	// Marshal authorizations to JSON
	authJSON, err := json.Marshal(proj.Authorizations)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling authorizations")
	}

	var id string
	err = r.db.QueryRowContext(ctx, query,
		proj.UUID,
		proj.Name,
		authJSON,
		proj.CreatedAt,
		proj.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return nil, errors.Wrap(err, "error creating project")
	}

	// Convert string UUID to ObjectID for compatibility
	if oid, err := primitive.ObjectIDFromHex(id); err == nil {
		proj.ID = oid
	} else {
		proj.ID = primitive.NewObjectID()
	}

	return proj, nil
}

func (r *repo) FindByUUID(ctx context.Context, uuid string) (*project.Project, error) {
	query := `
		SELECT id, uuid, name, authorizations, created_at, updated_at
		FROM projects
		WHERE uuid = $1`

	proj := &project.Project{}
	var dbID string
	var authJSON []byte

	err := r.db.QueryRowContext(ctx, query, uuid).Scan(
		&dbID,
		&proj.UUID,
		&proj.Name,
		&authJSON,
		&proj.CreatedAt,
		&proj.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "error finding project by uuid")
	}

	// Convert string UUID to ObjectID for compatibility
	if oid, err := primitive.ObjectIDFromHex(dbID); err == nil {
		proj.ID = oid
	} else {
		proj.ID = primitive.NewObjectID()
	}

	// Unmarshal authorizations
	if err := json.Unmarshal(authJSON, &proj.Authorizations); err != nil {
		// If unmarshal fails, set empty slice
		proj.Authorizations = []struct {
			UserEmail string `json:"user_email"`
			Role      string `json:"role"`
		}{}
	}

	return proj, nil
}

func (r *repo) Update(ctx context.Context, proj *project.Project) (*project.Project, error) {
	query := `
		UPDATE projects
		SET name = $2, updated_at = $3
		WHERE id = $1`

	proj.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		proj.ID.Hex(),
		proj.Name,
		proj.UpdatedAt,
	)

	if err != nil {
		return nil, errors.Wrap(err, "error updating project")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, "error checking affected rows")
	}

	if rowsAffected == 0 {
		return nil, errors.New("project not found")
	}

	return proj, nil
}

// UpdateAuthorizations updates only the authorizations field
func (r *repo) UpdateAuthorizations(ctx context.Context, uuid string, authorizations interface{}) error {
	query := `
		UPDATE projects
		SET authorizations = $2, updated_at = $3
		WHERE uuid = $1`

	authJSON, err := json.Marshal(authorizations)
	if err != nil {
		return errors.Wrap(err, "error marshaling authorizations")
	}

	result, err := r.db.ExecContext(ctx, query, uuid, authJSON, time.Now())
	if err != nil {
		return errors.Wrap(err, "error updating authorizations")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "error checking affected rows")
	}

	if rowsAffected == 0 {
		return errors.New("project not found")
	}

	return nil
}
