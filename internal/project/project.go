package project

import (
	"context"
	"time"
)

type Project struct {
	ID            string `json:"id,omitempty"`                              // PostgreSQL UUID (primary key)
	MongoObjectID string `json:"mongo_object_id,omitempty" bson:"_id,omitempty"` // MongoDB ObjectID for backward compatibility
	UUID          string `json:"uuid"`
	Name          string `json:"name"`
	Authorizations []struct {
		UserEmail string `json:"user_email"`
		Role      string `json:"role"`
	} `json:"authorizations"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
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
