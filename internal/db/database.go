package db

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/weni-ai/flows-code-actions/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func GetMongoDatabase(cf *config.Config) (*mongo.Database, error) {
	return GetMongoDatabaseWithRetry(cf, 3)
}

func GetMongoDatabaseWithRetry(cf *config.Config, maxRetries int) (*mongo.Database, error) {
	if maxRetries <= 0 {
		maxRetries = cf.DB.MaxRetries
	}
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
		// Connection timeouts - using config values
		SetConnectTimeout(time.Duration(cf.DB.ConnectTimeoutSeconds) * time.Second).
		SetServerSelectionTimeout(time.Duration(cf.DB.ServerSelectionTimeout) * time.Second).
		SetSocketTimeout(time.Duration(cf.DB.SocketTimeoutSeconds) * time.Second).

		// Connection pool settings - using config values
		SetMaxPoolSize(uint64(cf.DB.MaxPoolSize)).
		SetMinPoolSize(uint64(cf.DB.MinPoolSize)).
		SetMaxConnIdleTime(5 * time.Minute).

		// Read preference for replica set resilience
		// SecondaryPreferred allows reading from secondary if primary is unavailable
		SetReadPreference(readpref.SecondaryPreferred()).

		// Retry settings
		SetRetryWrites(true).
		SetRetryReads(true).

		// Heartbeat for faster detection of topology changes - using config value
		SetHeartbeatInterval(time.Duration(cf.DB.HeartbeatIntervalSeconds) * time.Second).

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

// HealthCheckResult represents the result of a MongoDB health check
type HealthCheckResult struct {
	Status     string            `json:"status"`      // "healthy", "degraded", or "unhealthy"
	Latency    time.Duration     `json:"latency"`     // Connection latency
	ReplicaSet *ReplicaSetStatus `json:"replica_set"` // Replica set information
	Error      string            `json:"error,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
}

// ReplicaSetStatus contains information about the replica set topology
type ReplicaSetStatus struct {
	Type           string         `json:"type"`
	ServersCount   int            `json:"servers_count"`
	PrimaryCount   int            `json:"primary_count"`
	SecondaryCount int            `json:"secondary_count"`
	UnknownCount   int            `json:"unknown_count"`
	Servers        []ServerStatus `json:"servers"`
}

// ServerStatus represents the status of a single server in the replica set
type ServerStatus struct {
	Address string        `json:"address"`
	Type    string        `json:"type"`
	RTT     time.Duration `json:"rtt"`
	Error   string        `json:"error,omitempty"`
}

// PerformHealthCheck performs a comprehensive health check of the MongoDB connection
func PerformHealthCheck(cf *config.Config) *HealthCheckResult {
	start := time.Now()
	result := &HealthCheckResult{
		Timestamp: start,
	}

	// Try to connect and ping MongoDB
	db, err := connectToMongoDB(cf)
	if err != nil {
		result.Status = "unhealthy"
		result.Error = err.Error()
		result.Latency = time.Since(start)
		return result
	}

	// Get MongoDB client to check replica set status
	client := db.Client()

	// Perform ping to measure latency
	pingStart := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, readpref.Primary())
	pingLatency := time.Since(pingStart)
	result.Latency = pingLatency

	if err != nil {
		// Primary ping failed, try secondary
		secondaryErr := client.Ping(ctx, readpref.SecondaryPreferred())
		if secondaryErr != nil {
			result.Status = "unhealthy"
			result.Error = fmt.Sprintf("Primary ping failed: %v, Secondary ping failed: %v", err, secondaryErr)
			return result
		}
		// Secondary ping succeeded, but primary failed - degraded state
		result.Status = "degraded"
		result.Error = fmt.Sprintf("Primary unavailable: %v", err)
	} else {
		result.Status = "healthy"
	}

	// Get replica set topology information
	result.ReplicaSet = getReplicaSetStatus(client)

	// Determine final health status based on replica set
	if result.ReplicaSet != nil {
		if result.ReplicaSet.PrimaryCount == 0 {
			result.Status = "degraded"
			if result.Error == "" {
				result.Error = "No primary server available"
			}
		}
		if result.ReplicaSet.UnknownCount > 0 {
			if result.Status == "healthy" {
				result.Status = "degraded"
			}
			if result.Error == "" {
				result.Error = fmt.Sprintf("%d server(s) in unknown state", result.ReplicaSet.UnknownCount)
			}
		}
	}

	return result
}

// getReplicaSetStatus extracts replica set topology information
func getReplicaSetStatus(client *mongo.Client) *ReplicaSetStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := &ReplicaSetStatus{
		Servers: make([]ServerStatus, 0),
	}

	// Try to get topology from admin command
	var adminResult bson.M
	err := client.Database("admin").RunCommand(ctx, bson.D{bson.E{Key: "hello", Value: 1}}).Decode(&adminResult)
	if err != nil {
		log.WithError(err).Warn("Failed to get replica set status")
		return result
	}

	// Parse basic replica set information
	if setName, ok := adminResult["setName"].(string); ok && setName != "" {
		result.Type = "ReplicaSet"
	} else {
		result.Type = "Standalone"
		return result
	}

	// For detailed server information, we'd need access to the topology description
	// This is a simplified version that at least tells us if we're in a replica set
	if primary, ok := adminResult["primary"].(string); ok && primary != "" {
		result.PrimaryCount = 1
		result.Servers = append(result.Servers, ServerStatus{
			Address: primary,
			Type:    "Primary",
		})
	}

	if hosts, ok := adminResult["hosts"].(bson.A); ok {
		result.ServersCount = len(hosts)
		result.SecondaryCount = len(hosts) - result.PrimaryCount
	}

	return result
}

// LogHealthStatus logs the health check result in a structured way
func LogHealthStatus(result *HealthCheckResult) {
	fields := log.Fields{
		"status":  result.Status,
		"latency": result.Latency,
	}

	if result.ReplicaSet != nil {
		fields["rs_type"] = result.ReplicaSet.Type
		fields["rs_servers"] = result.ReplicaSet.ServersCount
		fields["rs_primaries"] = result.ReplicaSet.PrimaryCount
		fields["rs_secondaries"] = result.ReplicaSet.SecondaryCount
		fields["rs_unknown"] = result.ReplicaSet.UnknownCount
	}

	if result.Error != "" {
		fields["error"] = result.Error
	}

	switch result.Status {
	case "healthy":
		log.WithFields(fields).Info("MongoDB health check: healthy")
	case "degraded":
		log.WithFields(fields).Warn("MongoDB health check: degraded")
	case "unhealthy":
		log.WithFields(fields).Error("MongoDB health check: unhealthy")
	}
}
