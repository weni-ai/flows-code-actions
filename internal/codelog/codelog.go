package codelog

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LogType string

const (
	TypeDebug LogType = "debug"
	TypeInfo  LogType = "info"
	TypeError LogType = "error"
)

type CodeLog struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	RunID  primitive.ObjectID `bson:"run_id" json:"run_id"`
	CodeID primitive.ObjectID `bson:"code_id" json:"code_id"`

	Type    LogType `bson:"type" json:"type"`
	Content string  `bson:"content" json:"content"`

	CreatedAt time.Time `bson:"creted_at" json:"creted_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type UseCase interface {
	Create(ctx context.Context, log *CodeLog) (*CodeLog, error)
	GetByID(ctx context.Context, id string) (*CodeLog, error)
	ListRunLogs(ctx context.Context, runID string) ([]CodeLog, error)
	Update(ctx context.Context, id string, Content string) (*CodeLog, error)
	Delete(ctx context.Context, id string) error
}

func NewCodeLog(runID string, codeID string, logType LogType, content string) *CodeLog {
	primitiveRunID, _ := primitive.ObjectIDFromHex(runID)
	primitiveCodeID, _ := primitive.ObjectIDFromHex(codeID)
	return &CodeLog{
		RunID: primitiveRunID, CodeID: primitiveCodeID, Type: logType, Content: content,
	}
}
