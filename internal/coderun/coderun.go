package coderun

import (
	"context"
	"time"
)

type CodeRunStatus string

const (
	StatusQueued    CodeRunStatus = "queued"
	StatusStarted   CodeRunStatus = "started"
	StatusCompleted CodeRunStatus = "completed"
	StatusFailed    CodeRunStatus = "failed"
)

type CodeRun struct {
	ID string `bson:"_id,omitempty" json:"_id,omitempty"`

	CodeID string        `bson:"code_id" json:"code_id"`
	Status CodeRunStatus `bson:"status" json:"status"`
	Result string        `bson:"result" json:"result"`

	CreatedAt time.Time `bson:"creted_at" json:"creted_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type UseCase interface {
	Create(ctx context.Context, codeRun *CodeRun) (*CodeRun, error)
	GetByID(ctx context.Context, id string) (*CodeRun, error)
	ListByCodeID(ctx context.Context, codeID string) ([]CodeRun, error)
	Update(ctx context.Context, codeRunID string, codeRun *CodeRun) (*CodeRun, error)
	Delete(ctx context.Context, id string) error
}

func NewCodeRun(codeID string, status CodeRunStatus) *CodeRun {
	return &CodeRun{CodeID: codeID, Status: status}
}
