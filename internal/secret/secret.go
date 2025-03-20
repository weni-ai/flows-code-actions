package secret

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Secret struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	Name        string `bson:"name" json:"name"`
	Value       string `bson:"value" json:"value"`
	ProjectUUID string `bson:"project_uuid" json:"project_uuid"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type UseCase interface {
	Create(ctx context.Context, secret *Secret) (*Secret, error)
	GetByID(ctx context.Context, id string) (*Secret, error)
	ListProjectSecrets(ctx context.Context, projectUUID string) ([]Secret, error)
	Update(ctx context.Context, id string, name string, value string) (*Secret, error)
	Delete(ctx context.Context, id string) error
}

func NewSecret(name string, value string, projectUUID string) *Secret {
	return &Secret{
		Name:        name,
		Value:       value,
		ProjectUUID: projectUUID,
	}
}
