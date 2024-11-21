package code

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CodeType string
type LanguageType string

const (
	TypeFlow     CodeType = "flow"
	TypeEndpoint CodeType = "endpoint"

	TypePy LanguageType = "python"
	TypeGo LanguageType = "go"
	TypeJS LanguageType = "javascript"
)

type Code struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	Name        string       `bson:"name" json:"name"`
	Type        CodeType     `bson:"type" json:"type"`
	Source      string       `bson:"source" json:"source"`
	Language    LanguageType `bson:"language" json:"language"`
	URL         string       `bson:"url" json:"url,omitempty"`
	ProjectUUID string       `bson:"project_uuid" json:"project_uuid"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type UseCase interface {
	Create(ctx context.Context, code *Code) (*Code, error)
	GetByID(ctx context.Context, id string) (*Code, error)
	ListProjectCodes(ctx context.Context, projectUUID string, codeType string) ([]Code, error)
	Update(ctx context.Context, id string, name string, source string, codeType string) (*Code, error)
	Delete(ctx context.Context, codeID string) error
}

func NewCodeAction(name, source string, language LanguageType, codeType CodeType, url string, projectUUID string) *Code {
	switch codeType {
	case TypeFlow:
		return NewFlowCode(name, source, language, projectUUID)
	case TypeEndpoint:
		return NewEndpointCode(name, source, language, url, projectUUID)
	}
	return nil
}

func NewFlowCode(name string, source string, language LanguageType, projectUUID string) *Code {
	return &Code{
		Name: name, Type: TypeFlow, Source: source, ProjectUUID: projectUUID, Language: language,
	}
}

func NewEndpointCode(name string, source string, language LanguageType, url string, projectUUID string) *Code {
	return &Code{
		Name: name, Type: TypeEndpoint, Source: source, URL: url, ProjectUUID: projectUUID, Language: language,
	}
}

func (t *CodeType) Validate() error {
	switch *t {
	case TypeFlow:
		return nil
	case TypeEndpoint:
		return nil
	}
	return fmt.Errorf(`code type of (%s) is not valid`, string(*t))
}

func (lang *LanguageType) Validate() error {

	switch *lang {
	case TypePy:
		return nil
	case TypeJS:
		return nil
	case TypeGo:
		return nil
	}
	return fmt.Errorf(`language type (%s) is not valid`, string(*lang))
}
