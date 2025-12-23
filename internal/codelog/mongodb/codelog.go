package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/codelog"
	"github.com/weni-ai/flows-code-actions/internal/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type codelogRepo struct {
	collection *mongo.Collection
}

func NewCodeLogRepository(db *mongo.Database) codelog.Repository {
	collection := db.Collection("codelog")
	return &codelogRepo{collection: collection}
}

func (r *codelogRepo) Count(ctx context.Context, runID, codeID string) (int64, error) {
	if runID == "" && codeID == "" {
		return 0, errors.New("must specify a run ID or a code ID")
	}

	findQuery := bson.M{}
	if runID != "" {
		pRunID, err := primitive.ObjectIDFromHex(runID)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse codeID to ObjectID")
		}
		findQuery["run_id"] = pRunID
	}

	if codeID != "" {
		pCodeID, err := primitive.ObjectIDFromHex(codeID)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse codeID to ObjectID")
		}
		findQuery["code_id"] = pCodeID
	}
	count, err := r.collection.CountDocuments(ctx, findQuery)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *codelogRepo) Create(ctx context.Context, codelog *codelog.CodeLog) (*codelog.CodeLog, error) {
	codelog.CreatedAt = time.Now()
	codelog.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, codelog)
	if err != nil {
		return nil, err
	}
	// Convert ObjectID to string
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		codelog.ID = oid.Hex()
	}
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

func (r *codelogRepo) ListRunLogs(ctx context.Context, runID string, codeID string, limit, page int) ([]codelog.CodeLog, error) {
	logs := []codelog.CodeLog{}

	if runID == "" && codeID == "" {
		return nil, errors.New("must specify a run ID or a code ID")
	}

	findQuery := bson.M{}
	if runID != "" {
		pRunID, err := primitive.ObjectIDFromHex(runID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse codeID to ObjectID")
		}
		findQuery["run_id"] = pRunID
	}

	if codeID != "" {
		pCodeID, err := primitive.ObjectIDFromHex(codeID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse codeID to ObjectID")
		}
		findQuery["code_id"] = pCodeID
	}

	options := db.NewMongoPaginate(limit, page).GetpaginatedOpts()
	c, err := r.collection.Find(ctx, findQuery, options)
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

func (r *codelogRepo) DeleteOlder(ctx context.Context, date time.Time, limit int64) (int64, error) {

	qry := bson.M{
		"created_at": bson.M{"$lt": date},
	}

	options := &options.FindOptions{Limit: &limit}
	cursor, err := r.collection.Find(ctx, qry, options)
	if err != nil {
		return 0, fmt.Errorf("find failed: %v", err)
	}
	defer cursor.Close(ctx)
	logs := []codelog.CodeLog{}
	for cursor.Next(ctx) {
		var log codelog.CodeLog
		if err = cursor.Decode(&log); err != nil {
			return 0, fmt.Errorf("failed to parse log from cursor: %v", err)
		}
		logs = append(logs, log)
	}

	if len(logs) > 0 {
		ids := make([]primitive.ObjectID, 0, len(logs))
		for _, log := range logs {
			// Convert string ID to ObjectID
			if oid, err := primitive.ObjectIDFromHex(log.ID); err == nil {
				ids = append(ids, oid)
			}
		}
		if len(ids) > 0 {
			result, err := r.collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": ids}})
			if err != nil {
				return 0, fmt.Errorf("failed to delete logs: %v", err)
			}
			return result.DeletedCount, nil
		}
	}
	return 0, nil
}
