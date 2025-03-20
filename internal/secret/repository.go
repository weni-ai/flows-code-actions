package secret

import "context"

type Repository interface {
	Create(ctx context.Context, secret *Secret) (*Secret, error)
	GetByID(ctx context.Context, id string) (*Secret, error)
	ListByProjectUUID(ctx context.Context, projectUUID string) ([]Secret, error)
	Update(ctx context.Context, id string, secret *Secret) (*Secret, error)
	Delete(ctx context.Context, id string) error
}
