package postgre

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/weni-ai/flows-code-actions/config"

	_ "github.com/lib/pq"
)

// GetPostgreDatabase returns a PostgreSQL database connection
func GetPostgreDatabase(cf *config.Config) (*sql.DB, error) {
	return GetPostgreDatabaseWithRetry(cf, 3)
}

// GetPostgreDatabaseWithRetry attempts to connect to PostgreSQL with retry logic
func GetPostgreDatabaseWithRetry(cf *config.Config, maxRetries int) (*sql.DB, error) {
	if maxRetries <= 0 {
		maxRetries = cf.DB.MaxRetries
	}
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.WithField("attempt", attempt).Info("Attempting PostgreSQL connection")

		db, err := connectToPostgreSQL(cf)
		if err != nil {
			lastErr = err
			log.WithFields(log.Fields{
				"attempt": attempt,
				"error":   err,
			}).Warn("PostgreSQL connection failed")

			if attempt < maxRetries {
				// Exponential backoff: 2s, 4s, 8s
				waitTime := time.Duration(1<<uint(attempt)) * time.Second
				log.WithField("wait_time", waitTime).Info("Waiting before retry")
				time.Sleep(waitTime)
				continue
			}
		} else {
			log.WithField("attempt", attempt).Info("PostgreSQL connection successful")
			return db, nil
		}
	}

	return nil, errors.Wrapf(lastErr, "failed to connect to PostgreSQL after %d attempts", maxRetries)
}

// connectToPostgreSQL establishes a connection to PostgreSQL
func connectToPostgreSQL(cf *config.Config) (*sql.DB, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cf.DB.Timeout)*time.Second)
	defer cancel()

	// Open database connection
	db, err := sql.Open("postgres", cf.DB.URI)
	if err != nil {
		return nil, errors.Wrap(err, "error opening PostgreSQL connection")
	}

	// Configure connection pool settings
	db.SetMaxOpenConns(cf.DB.MaxPoolSize)
	db.SetMaxIdleConns(cf.DB.MinPoolSize)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Test connection with ping
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, errors.Wrap(err, "PostgreSQL ping failed")
	}

	log.Info("PostgreSQL connection established successfully")
	return db, nil
}

// PostgrePaginate handles SQL pagination
type PostgrePaginate struct {
	limit  int
	offset int
}

// NewPostgrePaginate creates a new pagination helper
func NewPostgrePaginate(limit, page int) *PostgrePaginate {
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}
	return &PostgrePaginate{
		limit:  limit,
		offset: offset,
	}
}

// GetLimitOffset returns the LIMIT and OFFSET values for SQL queries
func (p *PostgrePaginate) GetLimitOffset() (limit int, offset int) {
	return p.limit, p.offset
}

// ApplyPagination applies pagination to a SQL query
func (p *PostgrePaginate) ApplyPagination(query string) string {
	return fmt.Sprintf("%s LIMIT %d OFFSET %d", query, p.limit, p.offset)
}

// HealthCheckResult represents the result of a PostgreSQL health check
type HealthCheckResult struct {
	Status         string          `json:"status"`          // "healthy", "degraded", or "unhealthy"
	Latency        time.Duration   `json:"latency"`         // Connection latency
	DatabaseStats  *DatabaseStats  `json:"database_stats"`  // Database statistics
	ConnectionPool *ConnectionPool `json:"connection_pool"` // Connection pool status
	Error          string          `json:"error,omitempty"`
	Timestamp      time.Time       `json:"timestamp"`
}

// DatabaseStats contains database-level statistics
type DatabaseStats struct {
	Version         string `json:"version"`
	ActiveQueries   int64  `json:"active_queries"`
	DatabaseSize    string `json:"database_size"`
	ConnectionCount int    `json:"connection_count"`
	Uptime          string `json:"uptime"`
}

// ConnectionPool represents connection pool status
type ConnectionPool struct {
	MaxOpenConns int `json:"max_open_conns"`
	OpenConns    int `json:"open_conns"`
	InUse        int `json:"in_use"`
	Idle         int `json:"idle"`
}

// PerformHealthCheck performs a comprehensive health check of the PostgreSQL connection
func PerformHealthCheck(cf *config.Config) *HealthCheckResult {
	start := time.Now()
	result := &HealthCheckResult{
		Timestamp: start,
	}

	// Try to connect and ping PostgreSQL
	db, err := connectToPostgreSQL(cf)
	if err != nil {
		result.Status = "unhealthy"
		result.Error = err.Error()
		result.Latency = time.Since(start)
		return result
	}
	defer db.Close()

	// Perform ping to measure latency
	pingStart := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	pingLatency := time.Since(pingStart)
	result.Latency = pingLatency

	if err != nil {
		result.Status = "unhealthy"
		result.Error = fmt.Sprintf("Ping failed: %v", err)
		return result
	}

	result.Status = "healthy"

	// Get database statistics
	result.DatabaseStats = getDatabaseStats(db, ctx)
	result.ConnectionPool = getConnectionPoolStats(db)

	// Determine final health status based on statistics
	if result.DatabaseStats != nil && result.DatabaseStats.ActiveQueries > 1000 {
		result.Status = "degraded"
		if result.Error == "" {
			result.Error = "High number of active queries"
		}
	}

	return result
}

// getDatabaseStats extracts database statistics
func getDatabaseStats(db *sql.DB, ctx context.Context) *DatabaseStats {
	stats := &DatabaseStats{}

	// Get PostgreSQL version
	var version string
	err := db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	if err == nil {
		stats.Version = version
	}

	// Get active queries count
	var activeQueries sql.NullInt64
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pg_stat_activity WHERE state = 'active'").Scan(&activeQueries)
	if err == nil && activeQueries.Valid {
		stats.ActiveQueries = activeQueries.Int64
	}

	// Get database size
	var dbSize sql.NullString
	err = db.QueryRowContext(ctx, "SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&dbSize)
	if err == nil && dbSize.Valid {
		stats.DatabaseSize = dbSize.String
	}

	// Get connection count
	var connCount sql.NullInt64
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pg_stat_activity").Scan(&connCount)
	if err == nil && connCount.Valid {
		stats.ConnectionCount = int(connCount.Int64)
	}

	// Get uptime
	var uptime sql.NullString
	err = db.QueryRowContext(ctx, "SELECT EXTRACT(EPOCH FROM (now() - pg_postmaster_start_time()))::text || ' seconds'").Scan(&uptime)
	if err == nil && uptime.Valid {
		stats.Uptime = uptime.String
	}

	return stats
}

// getConnectionPoolStats extracts connection pool statistics
func getConnectionPoolStats(db *sql.DB) *ConnectionPool {
	dbStats := db.Stats()
	return &ConnectionPool{
		MaxOpenConns: dbStats.MaxOpenConnections,
		OpenConns:    dbStats.OpenConnections,
		InUse:        dbStats.InUse,
		Idle:         dbStats.Idle,
	}
}

// LogHealthStatus logs the health check result in a structured way
func LogHealthStatus(result *HealthCheckResult) {
	fields := log.Fields{
		"status":  result.Status,
		"latency": result.Latency,
	}

	if result.DatabaseStats != nil {
		fields["version"] = result.DatabaseStats.Version
		fields["active_queries"] = result.DatabaseStats.ActiveQueries
		fields["db_size"] = result.DatabaseStats.DatabaseSize
		fields["connections"] = result.DatabaseStats.ConnectionCount
		fields["uptime"] = result.DatabaseStats.Uptime
	}

	if result.ConnectionPool != nil {
		fields["pool_max"] = result.ConnectionPool.MaxOpenConns
		fields["pool_open"] = result.ConnectionPool.OpenConns
		fields["pool_in_use"] = result.ConnectionPool.InUse
		fields["pool_idle"] = result.ConnectionPool.Idle
	}

	if result.Error != "" {
		fields["error"] = result.Error
	}

	switch result.Status {
	case "healthy":
		log.WithFields(fields).Info("PostgreSQL health check: healthy")
	case "degraded":
		log.WithFields(fields).Warn("PostgreSQL health check: degraded")
	case "unhealthy":
		log.WithFields(fields).Error("PostgreSQL health check: unhealthy")
	}
}

// ExecuteInTransaction executes a function within a database transaction
func ExecuteInTransaction(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

// MigrateSchema runs database migrations (basic version)
func MigrateSchema(db *sql.DB, schemaSQL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, schemaSQL)
	if err != nil {
		return errors.Wrap(err, "failed to execute schema migration")
	}

	log.Info("Database schema migration completed successfully")
	return nil
}
