package secret

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/secret"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type secretRepo struct {
	collection *mongo.Collection
}

func NewSecretRepository(db *mongo.Database) secret.Repository {
	collection := db.Collection("secrets")
	return &secretRepo{collection: collection}
}

func (r *secretRepo) Create(ctx context.Context, secret *secret.Secret) (*secret.Secret, error) {
	secret.CreatedAt = time.Now()
	secret.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, secret)
	if err != nil {
		return nil, err
	}
	secret.ID = result.InsertedID.(primitive.ObjectID)
	return secret, nil
}

func (r *secretRepo) GetByID(ctx context.Context, id string) (*secret.Secret, error) {
	secret := &secret.Secret{}
	secretID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.Wrap(err, "error on parse id to ObjectID")
	}

	err = r.collection.FindOne(ctx, bson.M{"_id": secretID}).Decode(secret)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	return secret, err
}

func (r *secretRepo) ListByProjectUUID(ctx context.Context, projectUUID string) ([]secret.Secret, error) {
	secrets := []secret.Secret{}
	filter := bson.M{"project_uuid": projectUUID}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var secret secret.Secret
		if err := cursor.Decode(&secret); err != nil {
			return nil, err
		}
		secrets = append(secrets, secret)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return secrets, nil
}

func (r *secretRepo) Update(ctx context.Context, id string, secret *secret.Secret) (*secret.Secret, error) {
	secret.UpdatedAt = time.Now()
	secretID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.Wrap(err, "error on parse id to ObjectID")
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": secretID}, bson.M{"$set": secret})
	return secret, err
}

func (r *secretRepo) Delete(ctx context.Context, id string) error {
	secretID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.Wrap(err, "error on parse id to ObjectID")
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": secretID})
	return err
}
