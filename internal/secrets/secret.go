package secrets

import (
	"context"
	"time"
)

type Secret struct {
	ID          string `json:"id,omitempty"` // PostgreSQL UUID (primary key)
	Name        string `json:"name"`
	Value       string `json:"value"`
	ProjectUUID string `json:"project_uuid"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UseCase interface {
	Create(ctx context.Context, secret *Secret) (*Secret, error)
	GetByID(ctx context.Context, id string) (*Secret, error)
	GetByProjectUUID(ctx context.Context, projectUUID string) ([]Secret, error)
	Update(ctx context.Context, id string, name string, value string) (*Secret, error)
	Delete(ctx context.Context, id string) error
}

func NewSecret(name, value, projectUUID string) *Secret {
	return &Secret{
		Name:        name,
		Value:       value,
		ProjectUUID: projectUUID,
	}
}
