package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/coderun"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		coderun.ID = oid.Hex()
		coderun.MongoObjectID = oid.Hex()
	}
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

func (r *codeRunRepo) ListByCodeID(ctx context.Context, codeID string, filter map[string]interface{}) ([]coderun.CodeRun, error) {
	codes := []coderun.CodeRun{}
	pcodeID, err := primitive.ObjectIDFromHex(codeID)
	if err != nil {
		return nil, errors.Wrap(err, "error on parse id to ObjectID")
	}
	queryFilter := bson.M{"code_id": pcodeID}

	createdAtFilter := bson.M{}
	if afterFilter, ok := filter["after"]; ok {
		if afterFilterTime, ok := afterFilter.(time.Time); ok {
			createdAtFilter["$gte"] = primitive.NewDateTimeFromTime(afterFilterTime)

		}
	}
	if beforeFilter, ok := filter["before"]; ok {
		if beforeFilterTime, ok := beforeFilter.(time.Time); ok {
			createdAtFilter["$lte"] = primitive.NewDateTimeFromTime(beforeFilterTime)
		}
	}
	if len(createdAtFilter) > 0 {
		queryFilter["created_at"] = createdAtFilter
	}

	cstmt, err := r.collection.Find(context.Background(), queryFilter)
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

func (r *codeRunRepo) DeleteOlder(ctx context.Context, date time.Time, limit int64) (int64, error) {
	qry := bson.M{
		"created_at": bson.M{"$lt": date},
	}
	options := &options.FindOptions{Limit: &limit}
	cursor, err := r.collection.Find(ctx, qry, options)
	if err != nil {
		return 0, fmt.Errorf("find failed: %v", err)
	}
	defer cursor.Close(ctx)
	runs := []coderun.CodeRun{}
	for cursor.Next(ctx) {
		var run coderun.CodeRun
		if err := cursor.Decode(&run); err != nil {
			return 0, err
		}
		runs = append(runs, run)
	}
	if len(runs) > 0 {
		ids := make([]primitive.ObjectID, len(runs))
		for i, run := range runs {
			oid, err := primitive.ObjectIDFromHex(run.ID)
			if err != nil {
				continue
			}
			ids[i] = oid
		}
		res, err := r.collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": ids}})
		if err != nil {
			return 0, fmt.Errorf("failed to delete runs: %v", err)
		}
		return res.DeletedCount, nil
	}
	return 0, nil
}
