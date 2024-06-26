package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/weni-ai/flows-code-actions/internal/code"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type codeRepo struct {
	collection *mongo.Collection
}

func NewCodeRepository(db *mongo.Database) code.Repository {
	collection := db.Collection("code")
	return &codeRepo{collection: collection}
}

func (r *codeRepo) Create(ctx context.Context, code *code.Code) (*code.Code, error) {
	code.ID = primitive.NewObjectID().Hex()
	code.CreatedAt = time.Now()
	code.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, code)
	return code, err
}

func (r *codeRepo) GetByID(ctx context.Context, id string) (*code.Code, error) {
	codeAction := &code.Code{}
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(codeAction)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	return codeAction, err
}

func (r *codeRepo) ListByProjectUUID(ctx context.Context, projectUUID string, codeType string) ([]code.Code, error) {
	codes := []code.Code{}
	filter := bson.M{"project_uuid": projectUUID}
	if codeType != "" {
		ct := code.CodeType(codeType)
		if err := ct.Validate(); err != nil {
			return nil, err
		}
		filter["type"] = codeType
	}
	c, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer c.Close(ctx)

	for c.Next(ctx) {
		var code code.Code
		if err := c.Decode(&code); err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}

	if err := c.Err(); err != nil {
		return nil, err
	}

	return codes, err
}

func (r *codeRepo) Update(ctx context.Context, id string, codeAction *code.Code) (*code.Code, error) {
	codeAction.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": codeAction})
	return codeAction, err
}

func (r *codeRepo) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
