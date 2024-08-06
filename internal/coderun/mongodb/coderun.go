package mongodb

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/coderun"
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
	coderun.CreatedAt = time.Now()
	coderun.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(context.Background(), coderun)
	if err != nil {
		return nil, err
	}
	coderun.ID = result.InsertedID.(primitive.ObjectID)
	return coderun, err
}

func (r *codeRunRepo) GetByID(ctx context.Context, id string) (*coderun.CodeRun, error) {
	codeRun := &coderun.CodeRun{}
	coderunID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.Wrap(err, "error on parse id to ObjectID")
	}
	err = r.collection.FindOne(context.Background(), bson.M{"_id": coderunID}).Decode(codeRun)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("coderun not found")
	}
	return codeRun, err
}

func (r *codeRunRepo) ListByCodeID(ctx context.Context, codeID string) ([]coderun.CodeRun, error) {
	codes := []coderun.CodeRun{}
	pcodeID, err := primitive.ObjectIDFromHex(codeID)
	if err != nil {
		return nil, errors.Wrap(err, "error on parse id to ObjectID")
	}
	cstmt, err := r.collection.Find(context.Background(), bson.M{"code_id": pcodeID})
	if err != nil {
		return nil, err
	}
	defer cstmt.Close(context.Background())

	for cstmt.Next(context.Background()) {
		var coderun coderun.CodeRun
		if err := cstmt.Decode(&coderun); err != nil {
			return nil, err
		}
		codes = append(codes, coderun)
	}

	if err := cstmt.Err(); err != nil {
		return nil, err
	}

	return codes, err
}

func (r *codeRunRepo) Update(ctx context.Context, id string, codeRun *coderun.CodeRun) (*coderun.CodeRun, error) {
	codeRun.UpdatedAt = time.Now()
	coderunID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.Wrap(err, "error on parse id to ObjectID")
	}
	_, err = r.collection.UpdateOne(context.Background(), bson.M{"_id": coderunID}, bson.M{"$set": codeRun})
	return codeRun, err
}

func (r *codeRunRepo) Delete(ctx context.Context, id string) error {
	coderunID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.Wrap(err, "error on parse id to ObjectID")
	}
	_, err = r.collection.DeleteOne(context.Background(), bson.M{"_id": coderunID})
	return err
}
