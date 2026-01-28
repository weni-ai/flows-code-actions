package secrets

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/weni-ai/flows-code-actions/internal/secrets"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return db, mock
}

func TestNewSecretRepository(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	repo := NewSecretRepository(db)
	assert.NotNil(t, repo)
}

func TestRepoCreate_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewSecretRepository(db)
	ctx := context.Background()

	secret := &secrets.Secret{
		Name:   "API_KEY",
		Value:  "secret-value-123",
		CodeID: "code-uuid-123",
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow("generated-uuid-123")
	mock.ExpectQuery(`INSERT INTO secrets`).
		WithArgs(secret.Name, secret.Value, secret.CodeID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	result, err := repo.Create(ctx, secret)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "generated-uuid-123", result.ID)
	assert.Equal(t, "API_KEY", result.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepoCreate_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewSecretRepository(db)
	ctx := context.Background()

	secret := &secrets.Secret{
		Name:   "API_KEY",
		Value:  "secret-value-123",
		CodeID: "code-uuid-123",
	}

	mock.ExpectQuery(`INSERT INTO secrets`).
		WithArgs(secret.Name, secret.Value, secret.CodeID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	result, err := repo.Create(ctx, secret)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error creating secret")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepoGetByID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewSecretRepository(db)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "value", "code_id", "created_at", "updated_at"}).
		AddRow("secret-uuid-123", "API_KEY", "secret-value", "code-uuid-123", now, now)

	mock.ExpectQuery(`SELECT id, name, value, code_id, created_at, updated_at FROM secrets WHERE id = \$1`).
		WithArgs("secret-uuid-123").
		WillReturnRows(rows)

	result, err := repo.GetByID(ctx, "secret-uuid-123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "secret-uuid-123", result.ID)
	assert.Equal(t, "API_KEY", result.Name)
	assert.Equal(t, "secret-value", result.Value)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepoGetByID_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewSecretRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, name, value, code_id, created_at, updated_at FROM secrets WHERE id = \$1`).
		WithArgs("non-existent-id").
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByID(ctx, "non-existent-id")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "secret not found", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepoGetByCodeID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewSecretRepository(db)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "value", "code_id", "created_at", "updated_at"}).
		AddRow("secret-uuid-1", "API_KEY", "value-1", "code-uuid-123", now, now).
		AddRow("secret-uuid-2", "DB_PASSWORD", "value-2", "code-uuid-123", now, now)

	mock.ExpectQuery(`SELECT id, name, value, code_id, created_at, updated_at FROM secrets WHERE code_id = \$1 ORDER BY created_at DESC`).
		WithArgs("code-uuid-123").
		WillReturnRows(rows)

	result, err := repo.GetByCodeID(ctx, "code-uuid-123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "API_KEY", result[0].Name)
	assert.Equal(t, "DB_PASSWORD", result[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepoGetByCodeID_Empty(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewSecretRepository(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "value", "code_id", "created_at", "updated_at"})

	mock.ExpectQuery(`SELECT id, name, value, code_id, created_at, updated_at FROM secrets WHERE code_id = \$1 ORDER BY created_at DESC`).
		WithArgs("code-uuid-123").
		WillReturnRows(rows)

	result, err := repo.GetByCodeID(ctx, "code-uuid-123")

	assert.NoError(t, err)
	assert.Nil(t, result) // nil slice when no results
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepoUpdate_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewSecretRepository(db)
	ctx := context.Background()

	secret := &secrets.Secret{
		ID:     "secret-uuid-123",
		Name:   "NEW_API_KEY",
		Value:  "new-value",
		CodeID: "code-uuid-123",
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow("secret-uuid-123")
	mock.ExpectQuery(`UPDATE secrets SET name = \$2, value = \$3, code_id = \$4, updated_at = \$5 WHERE id = \$1 RETURNING id`).
		WithArgs("secret-uuid-123", secret.Name, secret.Value, secret.CodeID, sqlmock.AnyArg()).
		WillReturnRows(rows)

	result, err := repo.Update(ctx, "secret-uuid-123", secret)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "secret-uuid-123", result.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepoUpdate_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewSecretRepository(db)
	ctx := context.Background()

	secret := &secrets.Secret{
		ID:     "non-existent-id",
		Name:   "NEW_API_KEY",
		Value:  "new-value",
		CodeID: "code-uuid-123",
	}

	mock.ExpectQuery(`UPDATE secrets SET name = \$2, value = \$3, code_id = \$4, updated_at = \$5 WHERE id = \$1 RETURNING id`).
		WithArgs("non-existent-id", secret.Name, secret.Value, secret.CodeID, sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.Update(ctx, "non-existent-id", secret)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "secret not found", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepoDelete_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewSecretRepository(db)
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM secrets WHERE id = \$1`).
		WithArgs("secret-uuid-123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(ctx, "secret-uuid-123")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepoDelete_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewSecretRepository(db)
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM secrets WHERE id = \$1`).
		WithArgs("non-existent-id").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(ctx, "non-existent-id")

	assert.Error(t, err)
	assert.Equal(t, "secret not found", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepoDelete_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewSecretRepository(db)
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM secrets WHERE id = \$1`).
		WithArgs("secret-uuid-123").
		WillReturnError(sql.ErrConnDone)

	err := repo.Delete(ctx, "secret-uuid-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error deleting secret")
	assert.NoError(t, mock.ExpectationsWereMet())
}
