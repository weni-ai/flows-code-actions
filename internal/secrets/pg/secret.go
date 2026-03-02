package secrets

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/secrets"
)

type secretRepo struct {
	db *sql.DB
}

// NewSecretRepository creates a new PostgreSQL repository for secret entities
func NewSecretRepository(db *sql.DB) secrets.Repository {
	return &secretRepo{db: db}
}

func (r *secretRepo) Create(ctx context.Context, secret *secrets.Secret) (*secrets.Secret, error) {
	query := `
		INSERT INTO secrets (name, value, project_uuid, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id`

	secret.CreatedAt = time.Now()
	secret.UpdatedAt = time.Now()

	var id string
	err := r.db.QueryRowContext(ctx, query,
		secret.Name,
		secret.Value,
		secret.ProjectUUID,
		secret.CreatedAt,
		secret.UpdatedAt,
	).Scan(&id)

	if err != nil {
		// Check for unique constraint violation
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				if strings.Contains(pqErr.Constraint, "uq_secrets_project_name") {
					return nil, errors.Errorf("secret with name '%s' already exists in this project", secret.Name)
				}
			}
		}
		return nil, errors.Wrap(err, "error creating secret")
	}

	secret.ID = id
	return secret, nil
}

func (r *secretRepo) GetByID(ctx context.Context, id string) (*secrets.Secret, error) {
	query := `
		SELECT id, name, value, project_uuid, created_at, updated_at 
		FROM secrets 
		WHERE id = $1`

	secret := &secrets.Secret{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&secret.ID,
		&secret.Name,
		&secret.Value,
		&secret.ProjectUUID,
		&secret.CreatedAt,
		&secret.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("secret not found")
		}
		return nil, errors.Wrap(err, "error getting secret by id")
	}

	return secret, nil
}

func (r *secretRepo) GetByProjectUUID(ctx context.Context, projectUUID string) ([]secrets.Secret, error) {
	query := `
		SELECT id, name, value, project_uuid, created_at, updated_at 
		FROM secrets 
		WHERE project_uuid = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, projectUUID)
	if err != nil {
		return nil, errors.Wrap(err, "error listing secrets by project_uuid")
	}
	defer rows.Close()

	var secretList []secrets.Secret
	for rows.Next() {
		var s secrets.Secret

		err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.Value,
			&s.ProjectUUID,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error scanning secret row")
		}

		secretList = append(secretList, s)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating secret rows")
	}

	return secretList, nil
}

func (r *secretRepo) Update(ctx context.Context, id string, secret *secrets.Secret) (*secrets.Secret, error) {
	query := `
		UPDATE secrets 
		SET name = $2, value = $3, project_uuid = $4, updated_at = $5
		WHERE id = $1
		RETURNING id`

	secret.UpdatedAt = time.Now()

	var returnedID string
	err := r.db.QueryRowContext(ctx, query,
		id,
		secret.Name,
		secret.Value,
		secret.ProjectUUID,
		secret.UpdatedAt,
	).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("secret not found")
		}
		// Check for unique constraint violation (renaming to existing name)
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				if strings.Contains(pqErr.Constraint, "uq_secrets_project_name") {
					return nil, errors.Errorf("secret with name '%s' already exists in this project", secret.Name)
				}
			}
		}
		return nil, errors.Wrap(err, "error updating secret")
	}

	secret.ID = returnedID
	return secret, nil
}

func (r *secretRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM secrets WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "error deleting secret")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "error checking affected rows")
	}

	if rowsAffected == 0 {
		return errors.New("secret not found")
	}

	return nil
}
