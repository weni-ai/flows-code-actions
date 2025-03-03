package db

import (
	"context"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/weni-ai/flows-code-actions/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetMongoDatabase(cf *config.Config) (*mongo.Database, error) {
	mongoClientOptions := options.Client().ApplyURI(cf.DB.URI)
	ctx, ctxCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer ctxCancel()

	mongoClient, err := mongo.Connect(ctx, mongoClientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error on connect to mongo")
	}
	if err := mongoClient.Ping(ctx, nil); err != nil {
		return nil, errors.Wrap(err, "mongodb fail to ping")
	} else {
		log.Info("mongodb OK")
	}
	db := mongoClient.Database(cf.DB.Name)
	return db, nil
}

type MongoPaginate struct {
	limit int64
	page  int64
}

func NewMongoPaginate(limit, page int) *MongoPaginate {
	return &MongoPaginate{
		limit: int64(limit),
		page:  int64(page),
	}
}

func (p *MongoPaginate) GetpaginatedOpts() *options.FindOptions {
	l := p.limit
	skip := p.page*p.limit - p.limit
	fOpt := options.FindOptions{Limit: &l, Skip: &skip}
	return &fOpt
}
