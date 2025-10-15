package handlers

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	serverHttp "github.com/weni-ai/flows-code-actions/internal/http/echo"

	"github.com/labstack/echo/v4"
)

type HealthHandler struct {
	server *serverHttp.Server
}

type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version,omitempty"`
	Uptime    time.Duration     `json:"uptime"`
	Services  map[string]Health `json:"services"`
	System    SystemInfo        `json:"system"`
}

type Health struct {
	Status    string        `json:"status"`
	Message   string        `json:"message"`
	Latency   time.Duration `json:"latency"`
	Timestamp time.Time     `json:"timestamp"`
}

type SystemInfo struct {
	GoVersion     string `json:"go_version"`
	NumGoroutines int    `json:"num_goroutines"`
	MemoryUsage   string `json:"memory_usage"`
	CPUCount      int    `json:"cpu_count"`
}

var (
	healthCache      *HealthStatus
	healthCacheMutex sync.RWMutex
	lastHealthCheck  time.Time
	cacheExpiry      = 30 * time.Second
	startTime        = time.Now()
)

func NewHealthHandler(server *serverHttp.Server) *HealthHandler {
	return &HealthHandler{server: server}
}

func (h HealthHandler) Health(c echo.Context) error {
	healthCacheMutex.RLock()
	if healthCache != nil && time.Since(lastHealthCheck) < cacheExpiry {
		defer healthCacheMutex.RUnlock()
		status := http.StatusOK
		if healthCache.Status != "healthy" {
			status = http.StatusServiceUnavailable
		}
		return c.JSON(status, healthCache)
	}
	healthCacheMutex.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	healthStatus := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime),
		Services:  make(map[string]Health),
		System:    h.getSystemInfo(),
	}

	var wg sync.WaitGroup
	serviceMutex := sync.Mutex{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		mongoHealth := h.checkMongoDB(ctx)
		serviceMutex.Lock()
		healthStatus.Services["mongodb"] = mongoHealth
		serviceMutex.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		redisHealth := h.checkRedis(ctx)
		serviceMutex.Lock()
		healthStatus.Services["redis"] = redisHealth
		serviceMutex.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		rabbitmqHealth := h.checkRabbitMQ(ctx)
		serviceMutex.Lock()
		healthStatus.Services["rabbitmq"] = rabbitmqHealth
		serviceMutex.Unlock()
	}()

	wg.Wait()

	allHealthy := true
	for _, service := range healthStatus.Services {
		if service.Status != "healthy" {
			allHealthy = false
			break
		}
	}

	if !allHealthy {
		healthStatus.Status = "unhealthy"
	}

	healthCacheMutex.Lock()
	healthCache = &healthStatus
	lastHealthCheck = time.Now()
	healthCacheMutex.Unlock()

	status := http.StatusOK
	if healthStatus.Status != "healthy" {
		status = http.StatusServiceUnavailable
	}

	return c.JSON(status, healthStatus)
}

// getSystemInfo collect system information
func (h HealthHandler) getSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemInfo{
		GoVersion:     runtime.Version(),
		NumGoroutines: runtime.NumGoroutine(),
		MemoryUsage:   formatBytes(m.Alloc),
		CPUCount:      runtime.NumCPU(),
	}
}

// formatBytes convert bytes to a readable format
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return "< 1 KB"
	}

	suffixes := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	div := uint64(1)
	exp := 0

	for b >= unit*div && exp < len(suffixes)-1 {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %s", float64(b)/float64(div), suffixes[exp])
}

func (h HealthHandler) checkMongoDB(ctx context.Context) Health {
	start := time.Now()
	timestamp := start

	if h.server.DB == nil {
		return Health{
			Status:    "unhealthy",
			Message:   "MongoDB connection not initialized",
			Latency:   0,
			Timestamp: timestamp,
		}
	}

	mongoCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := h.server.DB.Client().Ping(mongoCtx, nil)

	latency := time.Since(start)

	if err != nil {
		log.WithError(err).Error("MongoDB health check failed")

		log.WithFields(log.Fields{
			"latency": latency,
			"error":   err.Error(),
		}).Warn("MongoDB ping failed - connection pool will handle reconnection automatically")

		return Health{
			Status:    "unhealthy",
			Message:   fmt.Sprintf("MongoDB ping failed: %v (auto-reconnect enabled)", err.Error()),
			Latency:   latency,
			Timestamp: timestamp,
		}
	}

	return Health{
		Status:    "healthy",
		Message:   fmt.Sprintf("MongoDB is connected (ping: %v)", latency),
		Latency:   latency,
		Timestamp: timestamp,
	}
}

func (h HealthHandler) checkRedis(ctx context.Context) Health {
	start := time.Now()
	timestamp := start

	if h.server.Redis == nil {
		return Health{
			Status:    "unhealthy",
			Message:   "Redis connection not initialized",
			Latency:   0,
			Timestamp: timestamp,
		}
	}

	redisCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := h.server.Redis.Ping(redisCtx).Result()
	latency := time.Since(start)

	if err != nil {
		log.WithError(err).Error("Redis health check failed")
		return Health{
			Status:    "unhealthy",
			Message:   "Failed to ping Redis: " + err.Error(),
			Latency:   latency,
			Timestamp: timestamp,
		}
	}

	return Health{
		Status:    "healthy",
		Message:   fmt.Sprintf("Redis is connected (ping: %v)", latency),
		Latency:   latency,
		Timestamp: timestamp,
	}
}

func (h HealthHandler) checkRabbitMQ(ctx context.Context) Health {
	start := time.Now()
	timestamp := start

	if h.server.Config.EDA.RabbitmqURL == "" {
		return Health{
			Status:    "healthy",
			Message:   "RabbitMQ not configured (optional)",
			Latency:   0,
			Timestamp: timestamp,
		}
	}

	rabbitmqCtx, cancel := context.WithTimeout(ctx, 7*time.Second)
	defer cancel()

	type connResult struct {
		conn *amqp.Connection
		err  error
	}
	connCh := make(chan connResult, 1)

	go func() {
		conn, err := amqp.Dial(h.server.Config.EDA.RabbitmqURL)
		connCh <- connResult{conn: conn, err: err}
	}()

	var conn *amqp.Connection
	select {
	case result := <-connCh:
		if result.err != nil {
			latency := time.Since(start)
			log.WithError(result.err).Error("RabbitMQ health check failed")
			return Health{
				Status:    "unhealthy",
				Message:   "Failed to connect to RabbitMQ: " + result.err.Error(),
				Latency:   latency,
				Timestamp: timestamp,
			}
		}
		conn = result.conn
	case <-rabbitmqCtx.Done():
		latency := time.Since(start)
		return Health{
			Status:    "unhealthy",
			Message:   "RabbitMQ connection timeout",
			Latency:   latency,
			Timestamp: timestamp,
		}
	}

	defer conn.Close()

	latency := time.Since(start)
	return Health{
		Status:    "healthy",
		Message:   fmt.Sprintf("RabbitMQ is connected (latency: %v)", latency),
		Latency:   latency,
		Timestamp: timestamp,
	}
}

// HealthCheck endpoint simplified for basic monitoring (without external dependencies)
func (h HealthHandler) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now(),
		"uptime":    time.Since(startTime),
		"version":   runtime.Version(),
	})
}
