package project

import (
	"context"
	"errors"
	"time"

	"github.com/weni-ai/flows-code-actions/internal/project"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type repo struct {
	collection *mongo.Collection
}

func NewProjectRepository(db *mongo.Database) *repo {
	collection := db.Collection("project")
	return &repo{collection: collection}
}

func (r *repo) Create(ctx context.Context, project *project.Project) (*project.Project, error) {
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, project)
	if err != nil {
		return nil, err
	}
	project.ID = result.InsertedID.(primitive.ObjectID)
	return project, nil
}

func (r *repo) FindByUUID(ctx context.Context, uuid string) (*project.Project, error) {
	p := &project.Project{}
	filters := bson.M{"uuid": uuid}
	err := r.collection.FindOne(ctx, filters).Decode(p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	return p, nil
}

func (r *repo) Update(ctx context.Context, project *project.Project) (*project.Project, error) {
	filter := bson.M{"_id": project.ID}
	update := bson.M{
		"$set": bson.M{
			"name":       project.Name,
			"updated_at": time.Now(),
		},
	}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}
	if result.ModifiedCount == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return project, nil
}
