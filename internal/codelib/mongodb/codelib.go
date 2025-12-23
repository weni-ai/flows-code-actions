package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/weni-ai/flows-code-actions/internal/codelib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type codelibRepo struct {
	collection *mongo.Collection
}

func NewCodeLibRepo(db *mongo.Database) codelib.Repository {
	collection := db.Collection("codelib")
	return &codelibRepo{collection: collection}
}

func (r *codelibRepo) Create(ctx context.Context, cl *codelib.CodeLib) (*codelib.CodeLib, error) {
	cl.CreatedAt = time.Now()
	cl.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, cl)
	if err != nil {
		return nil, err
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		cl.ID = oid.Hex()
		cl.MongoObjectID = oid.Hex()
	}
	return cl, nil
}

func (r *codelibRepo) CreateBulk(ctx context.Context, cls []*codelib.CodeLib) ([]*codelib.CodeLib, error) {
	var ifaces []interface{}
	for _, cl := range cls {
		cl.CreatedAt = time.Now()
		cl.UpdatedAt = time.Now()
		ifaces = append(ifaces, cl)
	}
	_, err := r.collection.InsertMany(ctx, ifaces)
	if err != nil {
		return nil, err
	}
	return cls, nil
}

func (r *codelibRepo) List(ctx context.Context, lang *codelib.LanguageType) ([]codelib.CodeLib, error) {
	libs := []codelib.CodeLib{}
	var filter bson.M
	if lang != nil {
		filter = bson.M{"language": string(*lang)}
	}
	cls, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cls.Close(ctx)

	for cls.Next(ctx) {
		var cl codelib.CodeLib
		if err := cls.Decode(&cl); err != nil {
			return nil, err
		}
		libs = append(libs, cl)
	}
	if err := cls.Err(); err != nil {
		return nil, err
	}
	return libs, err
}

func (r *codelibRepo) Find(ctx context.Context, name string, language *codelib.LanguageType) (*codelib.CodeLib, error) {
	lib := &codelib.CodeLib{}
	var filters bson.M
	if language != nil {
		filters = bson.M{"language": string(*language), "name": name}
	} else {
		filters = bson.M{"name": name}
	}
	err := r.collection.FindOne(ctx, filters).Decode(lib)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	return lib, nil
}
