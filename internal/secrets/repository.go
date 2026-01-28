package secrets

import "context"

type Repository interface {
	Create(context.Context, *Secret) (*Secret, error)
	GetByID(context.Context, string) (*Secret, error)
	GetByCodeID(context.Context, string) ([]Secret, error)
	Update(context.Context, string, *Secret) (*Secret, error)
	Delete(context.Context, string) error
}
