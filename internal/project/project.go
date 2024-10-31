package project

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Project struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"id,omitempty"`
	UUID      string             `json:"uuid"`
	Name      string             `json:"name"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

func NewProject(uuid string, name string) *Project {
	return &Project{UUID: uuid, Name: name}
}

type UseCase interface {
	Create(ctx context.Context, project *Project) (*Project, error)
	FindByUUID(ctx context.Context, uuid string) (*Project, error)
	Update(ctx context.Context, project *Project) (*Project, error)
}

type Repository interface {
	Create(context.Context, *Project) (*Project, error)
	FindByUUID(context.Context, string) (*Project, error)
	Update(context.Context, *Project) (*Project, error)
}
