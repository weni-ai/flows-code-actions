package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/pkg/errors"
	"github.com/weni-ai/code-actions/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func GetMongoDatabase(cf *config.Config) (*mongo.Database, error) {
	mongoClientOptions := options.Client().ApplyURI(cf.DB.URI)
	ctx, ctxCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer ctxCancel()

	mongoClient, err := mongo.Connect(ctx, mongoClientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error on connect to mongo")
	}
	if err := mongoClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, errors.Wrap(err, "mongodb fail to ping")
	} else {
		slog.Info("mongodb OK")
	}
	db := mongoClient.Database(cf.DB.Name)
	return db, nil
}
