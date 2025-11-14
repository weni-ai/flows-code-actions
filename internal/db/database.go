package db

import (
	"context"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/weni-ai/flows-code-actions/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func GetMongoDatabase(cf *config.Config) (*mongo.Database, error) {
	return GetMongoDatabaseWithRetry(cf, 3)
}

func GetMongoDatabaseWithRetry(cf *config.Config, maxRetries int) (*mongo.Database, error) {
	var lastErr error
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.WithField("attempt", attempt).Info("Attempting MongoDB connection")
		
		db, err := connectToMongoDB(cf)
		if err != nil {
			lastErr = err
			log.WithFields(log.Fields{
				"attempt": attempt,
				"error":   err,
			}).Warn("MongoDB connection failed")
			
			if attempt < maxRetries {
				// Exponential backoff: 2s, 4s, 8s
				waitTime := time.Duration(1<<uint(attempt)) * time.Second
				log.WithField("wait_time", waitTime).Info("Waiting before retry")
				time.Sleep(waitTime)
				continue
			}
		} else {
			log.WithField("attempt", attempt).Info("MongoDB connection successful")
			return db, nil
		}
	}
	
	return nil, errors.Wrapf(lastErr, "failed to connect to MongoDB after %d attempts", maxRetries)
}

func connectToMongoDB(cf *config.Config) (*mongo.Database, error) {
	// Configure client options with robust settings for replica sets
	mongoClientOptions := options.Client().ApplyURI(cf.DB.URI).
		// Connection timeouts
		SetConnectTimeout(30 * time.Second).
		SetServerSelectionTimeout(30 * time.Second).
		SetSocketTimeout(30 * time.Second).
		
		// Connection pool settings
		SetMaxPoolSize(100).
		SetMinPoolSize(5).
		SetMaxConnIdleTime(5 * time.Minute).
		
		// Read preference for replica set resilience
		// SecondaryPreferred allows reading from secondary if primary is unavailable
		SetReadPreference(readpref.SecondaryPreferred()).
		
		// Retry settings
		SetRetryWrites(true).
		SetRetryReads(true).
		
		// Heartbeat for faster detection of topology changes
		SetHeartbeatInterval(10 * time.Second).
		
		// Compression for better performance over slow networks
		SetCompressors([]string{"zstd", "zlib", "snappy"})

	// Create context with timeout
	contextConnectionTimeout := time.Duration(cf.DB.Timeout) * time.Second
	ctx, ctxCancel := context.WithTimeout(context.Background(), contextConnectionTimeout)
	defer ctxCancel()

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(ctx, mongoClientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error connecting to MongoDB")
	}

	// Test connection with ping using secondary preferred read preference
	// This helps when primary is unavailable but secondaries are accessible
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer pingCancel()
	
	if err := mongoClient.Ping(pingCtx, readpref.SecondaryPreferred()); err != nil {
		// Try to close the client before returning error
		if closeErr := mongoClient.Disconnect(context.Background()); closeErr != nil {
			log.WithError(closeErr).Warn("Failed to close MongoDB client after ping failure")
		}
		return nil, errors.Wrap(err, "mongodb ping failed")
	}

	log.Info("MongoDB connection established successfully")
	
	// Get database instance
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
