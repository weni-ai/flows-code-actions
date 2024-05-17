package code

import (
	"context"
	"fmt"
	"time"
)

type CodeType string

const (
	TypeFlow     CodeType = "flow"
	TypeEndpoint CodeType = "endpoint"
)

type Code struct {
	ID string `bson:"_id,omitempty" json:"id,omitempty"`

	Name        string   `bson:"name" json:"name"`
	Type        CodeType `bson:"type" json:"type"`
	Source      string   `bson:"source" json:"source"`
	URL         string   `bson:"url" json:"url"`
	ProjectUUID string   `bson:"project_uuid" json:"project_uuid"`

	CreatedAt time.Time `bson:"creted_at" json:"creted_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type UseCase interface {
	Create(ctx context.Context, code *Code) (*Code, error)
	GetByID(ctx context.Context, id string) (*Code, error)
	ListProjectCodes(ctx context.Context, projectUUID string) ([]Code, error)
	Update(ctx context.Context, id string, name string, source string, codeType string) (*Code, error)
	Delete(ctx context.Context, codeID string) error
}

func NewCodeAction(name, source string, codeType CodeType, url string, projectUUID string) *Code {
	switch codeType {
	case TypeFlow:
		return NewFlowCode(name, source, projectUUID)
	case TypeEndpoint:
		return NewEndpointCode(name, source, url, projectUUID)
	}
	return nil
}

func NewFlowCode(name string, source string, projectUUID string) *Code {
	return &Code{
		Name: name, Type: TypeFlow, Source: source, ProjectUUID: projectUUID,
	}
}

func NewEndpointCode(name string, source string, url string, projectUUID string) *Code {
	return &Code{
		Name: name, Type: TypeEndpoint, Source: source, URL: url, ProjectUUID: projectUUID,
	}
}

func (t *CodeType) Validate() error {
	switch *t {
	case TypeFlow:
		return nil
	case TypeEndpoint:
		return nil
	}
	return fmt.Errorf("Code type of %v is not valid", t)
}
