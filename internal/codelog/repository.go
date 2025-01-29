package codelog

import (
	"context"
	"time"
)

type Repository interface {
	Create(context.Context, *CodeLog) (*CodeLog, error)
	GetByID(context.Context, string) (*CodeLog, error)
	ListRunLogs(context.Context, string) ([]CodeLog, error)
	Update(context.Context, string, string) (*CodeLog, error)
	Delete(context.Context, string) error
	DeleteOlder(context.Context, time.Time, int64) (int64, error)
}
