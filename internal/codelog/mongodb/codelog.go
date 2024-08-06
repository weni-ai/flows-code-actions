package mongodb

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/codelog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type codelogRepo struct {
	collection *mongo.Collection
}

func NewCodeLogRepository(db *mongo.Database) codelog.Repository {
	collection := db.Collection("codelog")
	return &codelogRepo{collection: collection}
}

func (r *codelogRepo) Create(ctx context.Context, codelog *codelog.CodeLog) (*codelog.CodeLog, error) {
	codelog.CreatedAt = time.Now()
	codelog.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, codelog)
	if err != nil {
		return nil, err
	}
	codelog.ID = result.InsertedID.(primitive.ObjectID)
	return codelog, err
}

func (r *codelogRepo) GetByID(ctx context.Context, id string) (*codelog.CodeLog, error) {
	log := &codelog.CodeLog{}
	codelogID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse id to ObjectID")
	}
	err = r.collection.FindOne(ctx, bson.M{"_id": codelogID}).Decode(log)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	return log, err
}

func (r *codelogRepo) ListRunLogs(ctx context.Context, runID string) ([]codelog.CodeLog, error) {
	logs := []codelog.CodeLog{}
	pRunID, err := primitive.ObjectIDFromHex(runID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse runID to ObjectID")
	}
	c, err := r.collection.Find(ctx, bson.M{"run_id": pRunID})
	if err != nil {
		return nil, err
	}
	defer c.Close(ctx)

	for c.Next(ctx) {
		var log codelog.CodeLog
		if err := c.Decode(&log); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	if err := c.Err(); err != nil {
		return nil, err
	}

	return logs, err
}

func (r *codelogRepo) Update(ctx context.Context, id string, content string) (*codelog.CodeLog, error) {
	log := &codelog.CodeLog{}
	codelogID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse id to ObjectID")
	}
	err = r.collection.FindOne(ctx, bson.M{"_id": codelogID}).Decode(log)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	log.UpdatedAt = time.Now()
	log.Content = content
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": log})
	return log, err
}

func (r *codelogRepo) Delete(ctx context.Context, id string) error {
	codelogID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.Wrap(err, "failed to parse id to ObjectID")
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": codelogID})
	return err
}
