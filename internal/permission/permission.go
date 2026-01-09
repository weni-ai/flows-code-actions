package permission

import (
	"context"
	"time"
)

type Role int

const (
	ModeratorRole   Role = Role(3)
	ContributorRole Role = Role(2)
	ViewerRole      Role = Role(1)
)

type PermissionAccess string

const (
	ReadPermission  PermissionAccess = "read"
	WritePermission PermissionAccess = "write"
)

type UserPermission struct {
	ID            string `json:"id,omitempty"`                              // PostgreSQL UUID (primary key)
	MongoObjectID string `json:"mongo_object_id,omitempty" bson:"_id,omitempty"` // MongoDB ObjectID for backward compatibility
	ProjectUUID   string `bson:"project_uuid,omitempty" json:"project_uuid,omitempty"`
	Email         string `bson:"email" json:"email,omitempty"`
	Role          Role   `bson:"role" json:"role,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

func NewUserPermission(projectUUID, email string, role Role) *UserPermission {
	return &UserPermission{
		ProjectUUID: projectUUID,
		Email:       email,
		Role:        role,
	}
}

func HasPermission(user *UserPermission, permission PermissionAccess) bool {
	if permission == ReadPermission && user.Role >= Role(1) {
		return true
	}
	if permission == WritePermission && user.Role >= Role(3) {
		return true
	}
	return false
}

type UserPermissionUseCase interface {
	Create(ctx context.Context, user *UserPermission) (*UserPermission, error)
	Find(ctx context.Context, user *UserPermission) (*UserPermission, error)
	Update(ctx context.Context, id string, user *UserPermission) (*UserPermission, error)
	Delete(ctx context.Context, id string) error
}

type UserPermissionRepository interface {
	Create(context.Context, *UserPermission) (*UserPermission, error)
	Find(context.Context, *UserPermission) (*UserPermission, error)
	Update(context.Context, string, *UserPermission) (*UserPermission, error)
	Delete(context.Context, string) error
}
