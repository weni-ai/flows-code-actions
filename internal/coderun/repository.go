package coderun

import "context"

type Repository interface {
	Create(context.Context, *CodeRun) (*CodeRun, error)
	GetByID(context.Context, string) (*CodeRun, error)
	ListByCodeID(context.Context, string, map[string]interface{}) ([]CodeRun, error)
	Update(context.Context, string, *CodeRun) (*CodeRun, error)
	Delete(context.Context, string) error
}
