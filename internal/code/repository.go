package code

import "context"

type Repository interface {
	Create(context.Context, *Code) (*Code, error)
	GetByID(context.Context, string) (*Code, error)
	ListByProjectUUID(context.Context, string, string) ([]Code, error)
	Update(context.Context, string, *Code) (*Code, error)
	Delete(context.Context, string) error
}
