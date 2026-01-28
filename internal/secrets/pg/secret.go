package secrets

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/secrets"

	_ "github.com/lib/pq"
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
		INSERT INTO secrets (name, value, code_id, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id`

	secret.CreatedAt = time.Now()
	secret.UpdatedAt = time.Now()

	var id string
	err := r.db.QueryRowContext(ctx, query,
		secret.Name,
		secret.Value,
		secret.CodeID,
		secret.CreatedAt,
		secret.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return nil, errors.Wrap(err, "error creating secret")
	}

	secret.ID = id
	return secret, nil
}

func (r *secretRepo) GetByID(ctx context.Context, id string) (*secrets.Secret, error) {
	query := `
		SELECT id, name, value, code_id, created_at, updated_at 
		FROM secrets 
		WHERE id = $1`

	secret := &secrets.Secret{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&secret.ID,
		&secret.Name,
		&secret.Value,
		&secret.CodeID,
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

func (r *secretRepo) GetByCodeID(ctx context.Context, codeID string) ([]secrets.Secret, error) {
	query := `
		SELECT id, name, value, code_id, created_at, updated_at 
		FROM secrets 
		WHERE code_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, codeID)
	if err != nil {
		return nil, errors.Wrap(err, "error listing secrets by code_id")
	}
	defer rows.Close()

	var secretList []secrets.Secret
	for rows.Next() {
		var s secrets.Secret

		err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.Value,
			&s.CodeID,
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
		SET name = $2, value = $3, code_id = $4, updated_at = $5
		WHERE id = $1
		RETURNING id`

	secret.UpdatedAt = time.Now()

	var returnedID string
	err := r.db.QueryRowContext(ctx, query,
		id,
		secret.Name,
		secret.Value,
		secret.CodeID,
		secret.UpdatedAt,
	).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("secret not found")
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
