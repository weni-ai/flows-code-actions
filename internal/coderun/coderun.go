package coderun

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/weni-ai/flows-code-actions/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CodeRunStatus string

const (
	StatusQueued    CodeRunStatus = "queued"
	StatusStarted   CodeRunStatus = "started"
	StatusCompleted CodeRunStatus = "completed"
	StatusFailed    CodeRunStatus = "failed"
)

type CodeRun struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	CodeID primitive.ObjectID     `bson:"code_id" json:"code_id"`
	Status CodeRunStatus          `bson:"status" json:"status"`
	Result string                 `bson:"result" json:"result"`
	Extra  map[string]interface{} `bson:"extra" json:"extra"`

	Params map[string]interface{} `bson:"params" json:"params"`
	Body   string                 `bson:"body" json:"body"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type UseCase interface {
	Create(ctx context.Context, codeRun *CodeRun) (*CodeRun, error)
	GetByID(ctx context.Context, id string) (*CodeRun, error)
	ListByCodeID(ctx context.Context, codeID string, filter map[string]interface{}) ([]CodeRun, error)
	Update(ctx context.Context, codeRunID string, codeRun *CodeRun) (*CodeRun, error)
	Delete(ctx context.Context, id string) error
	StartCodeRunCleaner(cfg *config.Config) error
}

func NewCodeRun(codeID string, status CodeRunStatus) *CodeRun {
	cID, _ := primitive.ObjectIDFromHex(codeID)
	return &CodeRun{CodeID: cID, Status: status}
}

func (c *CodeRun) StatusCode() (int, error) {
	if extraStatusCode, ok := c.Extra["status_code"]; ok {
		switch v := extraStatusCode.(type) {
		case int:
			return v, nil
		case int32:
			return int(v), nil
		case string:
			if scInt, err := strconv.Atoi(v); err == nil {
				return scInt, nil
			} else {
				return 0, fmt.Errorf("error converting status code to int: %v", err)
			}
		default:
			return 0, fmt.Errorf("unexpected type for status code: %s", reflect.TypeOf(v))
		}
	}
	return 0, fmt.Errorf("status_code couldn't be found")
}

func (c *CodeRun) ResponseContentType() string {
	if extraContentType, ok := c.Extra["content_type"]; ok {
		if contentType, isStr := extraContentType.(string); isStr {
			return contentType
		}
	}
	return "string"
}
