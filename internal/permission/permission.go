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

type Permission string

const (
	ReadPermission  Permission = "read"
	WritePermission Permission = "write"
)

type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ProjectUUID string             `bson:"project_uuid,omitempty" json:"project_uuid,omitempty"`
	Name        string             `bson:"username" json:"username,omitempty"`
	Email       string             `bson:"email" json:"email,omitempty"`
	Role        Role               `bson:"role" json:"role,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

func NewUser(projectUUID, name, email string, role Role) *User {
	return &User{
		ProjectUUID: projectUUID,
		Name:        name,
		Email:       email,
		Role:        role,
	}
}

var accessMatrix = map[Role]map[Permission]bool{
	ManagerRole: {
		ReadPermission:  true,
		WritePermission: true,
	},
	ViewerRole: {
		ReadPermission:  true,
		WritePermission: false,
	},
}

func HasPermission(user *User, permission Permission) bool {
	return accessMatrix[user.Role][permission]
}

type UserUseCase interface {
	Create(ctx context.Context, user *User) (*User, error)
	Find(ctx context.Context, user *User) (*User, error)
	Update(ctx context.Context, id string, user *User) (*User, error)
	Delete(ctx context.Context, id string) error
}

type UserRepository interface {
	Create(context.Context, *User) (*User, error)
	Find(context.Context, *User) (*User, error)
	Update(context.Context, string, *User) (*User, error)
	Delete(context.Context, string) error
}
