package handlers

import (
	"context"
	"net/http"
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
	Status   string            `json:"status"`
	Services map[string]Health `json:"services"`
}

type Health struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func NewHealthHandler(server *serverHttp.Server) *HealthHandler {
	return &HealthHandler{server: server}
}

func (h HealthHandler) Health(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	healthStatus := HealthStatus{
		Status:   "healthy",
		Services: make(map[string]Health),
	}

	mongoHealth := h.checkMongoDB(ctx)
	healthStatus.Services["mongodb"] = mongoHealth

	redisHealth := h.checkRedis(ctx)
	healthStatus.Services["redis"] = redisHealth

	rabbitmqHealth := h.checkRabbitMQ(ctx)
	healthStatus.Services["rabbitmq"] = rabbitmqHealth

	allHealthy := true
	for _, service := range healthStatus.Services {
		if service.Status != "healthy" {
			allHealthy = false
			break
		}
	}

	if !allHealthy {
		healthStatus.Status = "unhealthy"
		return c.JSON(http.StatusServiceUnavailable, healthStatus)
	}

	return c.JSON(http.StatusOK, healthStatus)
}

func (h HealthHandler) checkMongoDB(ctx context.Context) Health {
	if h.server.DB == nil {
		return Health{
			Status:  "unhealthy",
			Message: "MongoDB connection not initialized",
		}
	}

	err := h.server.DB.Client().Ping(ctx, nil)
	if err != nil {
		log.WithError(err).Error("MongoDB health check failed")
		return Health{
			Status:  "unhealthy",
			Message: "Failed to ping MongoDB: " + err.Error(),
		}
	}

	return Health{
		Status:  "healthy",
		Message: "MongoDB is connected",
	}
}

func (h HealthHandler) checkRedis(ctx context.Context) Health {
	if h.server.Redis == nil {
		return Health{
			Status:  "unhealthy",
			Message: "Redis connection not initialized",
		}
	}

	_, err := h.server.Redis.Ping(ctx).Result()
	if err != nil {
		log.WithError(err).Error("Redis health check failed")
		return Health{
			Status:  "unhealthy",
			Message: "Failed to ping Redis: " + err.Error(),
		}
	}

	return Health{
		Status:  "healthy",
		Message: "Redis is connected",
	}
}

func (h HealthHandler) checkRabbitMQ(ctx context.Context) Health {
	if h.server.Config.EDA.RabbitmqURL == "" {
		return Health{
			Status:  "healthy",
			Message: "RabbitMQ not configured (optional)",
		}
	}

	conn, err := amqp.Dial(h.server.Config.EDA.RabbitmqURL)
	if err != nil {
		log.WithError(err).Error("RabbitMQ health check failed")
		return Health{
			Status:  "unhealthy",
			Message: "Failed to connect to RabbitMQ: " + err.Error(),
		}
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.WithError(err).Error("RabbitMQ channel creation failed")
		return Health{
			Status:  "unhealthy",
			Message: "Failed to create RabbitMQ channel: " + err.Error(),
		}
	}
	defer ch.Close()

	return Health{
		Status:  "healthy",
		Message: "RabbitMQ is connected",
	}
}
