package permission

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/permission"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type userRepo struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *userRepo {
	collection := db.Collection("user_permissions")
	return &userRepo{collection: collection}
}

func (r *userRepo) Create(ctx context.Context, user *permission.UserPermission) (*permission.UserPermission, error) {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	exists, _ := r.Find(ctx, user)
	if exists != nil {
		return nil, errors.New("user permission already exists")
	}
	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}
	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil, nil
}

func (r *userRepo) Find(ctx context.Context, user *permission.UserPermission) (*permission.UserPermission, error) {
	u := &permission.UserPermission{}
	filters := bson.M{}
	if user.Email != "" {
		filters["email"] = user.Email
	}
	if user.ProjectUUID != "" {
		filters["project_uuid"] = user.ProjectUUID
	}
	if len(filters) <= 0 {
		return nil, errors.New("no filters specified for search user")
	}
	err := r.collection.FindOne(ctx, filters).Decode(u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	return u, nil
}

func (r *userRepo) Update(ctx context.Context, userID string, user *permission.UserPermission) (*permission.UserPermission, error) {
	user.UpdatedAt = time.Now()
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.Wrap(err, "error on parse user id to object id")
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": user})
	return user, err
}

func (r *userRepo) Delete(ctx context.Context, userID string) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.Wrap(err, "error on parse userID to ObjectID")
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
