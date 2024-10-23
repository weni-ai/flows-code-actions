package permission

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role string

const (
	ManagerRole Role = "manager"
	ViewerRole  Role = "viewer"
)

type PermissionRole string

const (
	ReadPermission  PermissionRole = "read"
	WritePermission PermissionRole = "write"
)

type UserPermission struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ProjectUUID string             `bson:"project_uuid,omitempty" json:"project_uuid,omitempty"`
	Name        string             `bson:"username" json:"username,omitempty"`
	Email       string             `bson:"email" json:"email,omitempty"`
	Role        Role               `bson:"role" json:"role,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

func NewUserPermission(projectUUID, name, email string, role Role) *UserPermission {
	return &UserPermission{
		ProjectUUID: projectUUID,
		Name:        name,
		Email:       email,
		Role:        role,
	}
}

var accessMatrix = map[Role]map[PermissionRole]bool{
	ManagerRole: {
		ReadPermission:  true,
		WritePermission: true,
	},
	ViewerRole: {
		ReadPermission:  true,
		WritePermission: false,
	},
}

func HasPermission(user *UserPermission, permission PermissionRole) bool {
	return accessMatrix[user.Role][permission]
}

type UserPeermissionUseCase interface {
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
