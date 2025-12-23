package pg

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/permission"

	_ "github.com/lib/pq"
)

type userRepo struct {
	db *sql.DB
}

// NewUserRepository creates a new PostgreSQL repository for user permissions
func NewUserRepository(db *sql.DB) *userRepo {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *permission.UserPermission) (*permission.UserPermission, error) {
	// Check if already exists
	exists, _ := r.Find(ctx, user)
	if exists != nil {
		return nil, errors.New("user permission already exists")
	}

	query := `
		INSERT INTO user_permissions (mongo_object_id, project_uuid, email, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	var id string
	err := r.db.QueryRowContext(ctx, query,
		nullString(user.MongoObjectID),
		user.ProjectUUID,
		user.Email,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return nil, errors.Wrap(err, "error creating user permission")
	}

	user.ID = id
	return user, nil
}

func (r *userRepo) Find(ctx context.Context, user *permission.UserPermission) (*permission.UserPermission, error) {
	query := `
		SELECT id, mongo_object_id, project_uuid, email, role, created_at, updated_at
		FROM user_permissions
		WHERE 1=1`

	args := []interface{}{}

	// Build dynamic query based on provided filters
	if user.Email != "" {
		query += " AND email = $1"
		args = append(args, user.Email)
	}

	if user.ProjectUUID != "" {
		if len(args) == 0 {
			query += " AND project_uuid = $1"
		} else {
			query += " AND project_uuid = $2"
		}
		args = append(args, user.ProjectUUID)
	}

	if len(args) == 0 {
		return nil, errors.New("no filters specified for search user")
	}

	u := &permission.UserPermission{}
	var mongoObjectID sql.NullString

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&u.ID,
		&mongoObjectID,
		&u.ProjectUUID,
		&u.Email,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "error finding user permission")
	}

	if mongoObjectID.Valid {
		u.MongoObjectID = mongoObjectID.String
	}

	return u, nil
}

func (r *userRepo) Update(ctx context.Context, userID string, user *permission.UserPermission) (*permission.UserPermission, error) {
	query := `
		UPDATE user_permissions
		SET mongo_object_id = $2, project_uuid = $3, email = $4, role = $5, updated_at = $6
		WHERE id = $1 OR mongo_object_id = $1
		RETURNING id`

	user.UpdatedAt = time.Now()

	var returnedID string
	err := r.db.QueryRowContext(ctx, query,
		userID,
		nullString(user.MongoObjectID),
		user.ProjectUUID,
		user.Email,
		user.Role,
		user.UpdatedAt,
	).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user permission not found")
		}
		return nil, errors.Wrap(err, "error updating user permission")
	}

	user.ID = returnedID
	return user, nil
}

func (r *userRepo) Delete(ctx context.Context, userID string) error {
	query := `DELETE FROM user_permissions WHERE id = $1 OR mongo_object_id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return errors.Wrap(err, "error deleting user permission")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "error checking affected rows")
	}

	if rowsAffected == 0 {
		return errors.New("user permission not found")
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
