package codelog

import (
	"context"
	"time"

	"github.com/weni-ai/flows-code-actions/config"
)

type LogType string

const (
	TypeDebug LogType = "debug"
	TypeInfo  LogType = "info"
	TypeError LogType = "error"
)

type CodeLog struct {
	ID string `bson:"_id,omitempty" json:"id,omitempty"`

	RunID  string `bson:"run_id" json:"run_id"`
	CodeID string `bson:"code_id" json:"code_id"`

	Type    LogType `bson:"type" json:"type"`
	Content string  `bson:"content" json:"content"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type UseCase interface {
	Create(ctx context.Context, log *CodeLog) (*CodeLog, error)
	GetByID(ctx context.Context, id string) (*CodeLog, error)
	ListRunLogs(ctx context.Context, runID, codeID string, limit, page int) ([]CodeLog, error)
	Update(ctx context.Context, id string, Content string) (*CodeLog, error)
	Delete(ctx context.Context, id string) error
	StartCodeLogCleaner(*config.Config) error
	Count(ctx context.Context, runID, codeID string) (int64, error)
}

func NewCodeLog(runID string, codeID string, logType LogType, content string) *CodeLog {
	return &CodeLog{
		RunID: runID, CodeID: codeID, Type: logType, Content: content,
	}
}
