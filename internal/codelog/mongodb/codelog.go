package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/weni-ai/code-actions/internal/codelog"
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
	codelog.ID = primitive.NewObjectID().Hex()
	codelog.CreatedAt = time.Now()
	codelog.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, codelog)
	return codelog, err
}

func (r *codelogRepo) GetByID(ctx context.Context, id string) (*codelog.CodeLog, error) {
	log := &codelog.CodeLog{}
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(log)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	return log, err
}

func (r *codelogRepo) ListRunLogs(ctx context.Context, runID string) ([]codelog.CodeLog, error) {
	logs := []codelog.CodeLog{}
	c, err := r.collection.Find(ctx, bson.M{"run_id": runID})
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
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(log)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	log.UpdatedAt = time.Now()
	log.Content = content
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": log})
	return log, err
}

func (r *codelogRepo) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
