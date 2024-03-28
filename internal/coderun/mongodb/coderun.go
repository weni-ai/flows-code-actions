package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/weni-ai/code-actions/internal/coderun"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type codeRunRepo struct {
	collection *mongo.Collection
}

func NewCodeRunRepository(db *mongo.Database) coderun.Repository {
	collection := db.Collection("coderun")
	return &codeRunRepo{collection: collection}
}

func (r *codeRunRepo) Create(ctx context.Context, coderun *coderun.CodeRun) (*coderun.CodeRun, error) {
	coderun.ID = primitive.NewObjectID().Hex()
	coderun.CreatedAt = time.Now()
	coderun.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(context.Background(), coderun)
	return coderun, err
}

func (r *codeRunRepo) GetByID(ctx context.Context, id string) (*coderun.CodeRun, error) {
	codeAction := &coderun.CodeRun{}
	err := r.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(codeAction)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("coderun not found")
	}
	return codeAction, err
}

func (r *codeRunRepo) ListByCodeID(ctx context.Context, codeID string) ([]coderun.CodeRun, error) {
	codes := []coderun.CodeRun{}
	c, err := r.collection.Find(context.Background(), bson.M{"code_id": codeID})
	if err != nil {
		return nil, err
	}
	defer c.Close(context.Background())

	for c.Next(context.Background()) {
		var coderun coderun.CodeRun
		if err := c.Decode(&coderun); err != nil {
			return nil, err
		}
		codes = append(codes, coderun)
	}

	if err := c.Err(); err != nil {
		return nil, err
	}

	return codes, err
}

func (r *codeRunRepo) Update(ctx context.Context, id string, codeAction *coderun.CodeRun) error {
	codeAction.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$set": codeAction})
	return err
}

func (r *codeRunRepo) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}
