package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/weni-ai/flows-code-actions/config"
	serverHttp "github.com/weni-ai/flows-code-actions/internal/http/echo"
)

func TestHealthHandler_Health_NoDependencies(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	server := &serverHttp.Server{
		DB:    nil,
		Redis: nil,
		Config: &config.Config{
			EDA: config.EDAConfig{
				RabbitmqURL: "",
			},
		},
	}

	handler := NewHealthHandler(server)

	err := handler.Health(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var response HealthStatus
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "unhealthy", response.Status)
	assert.Equal(t, "unhealthy", response.Services["mongodb"].Status)
	assert.Equal(t, "unhealthy", response.Services["redis"].Status)
	assert.Equal(t, "healthy", response.Services["rabbitmq"].Status)
	assert.Contains(t, response.Services["mongodb"].Message, "not initialized")
	assert.Contains(t, response.Services["redis"].Message, "not initialized")
	assert.Contains(t, response.Services["rabbitmq"].Message, "not configured")
}

func TestHealthHandler_Health_RabbitMQConfiguredWithInvalidURL(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	server := &serverHttp.Server{
		DB:    nil,
		Redis: nil,
		Config: &config.Config{
			EDA: config.EDAConfig{
				RabbitmqURL: "amqp://invalid-host:5672/",
			},
		},
	}

	handler := NewHealthHandler(server)

	err := handler.Health(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var response HealthStatus
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "unhealthy", response.Status)
	assert.Equal(t, "unhealthy", response.Services["mongodb"].Status)
	assert.Equal(t, "unhealthy", response.Services["redis"].Status)
	assert.Equal(t, "unhealthy", response.Services["rabbitmq"].Status)
	assert.Contains(t, response.Services["rabbitmq"].Message, "Failed to connect to RabbitMQ")
}

func TestHealthHandler_Health_ResponseStructure(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	server := &serverHttp.Server{
		DB:    nil,
		Redis: nil,
		Config: &config.Config{
			EDA: config.EDAConfig{
				RabbitmqURL: "",
			},
		},
	}

	handler := NewHealthHandler(server)

	err := handler.Health(c)

	assert.NoError(t, err)

	var response HealthStatus
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, []string{"healthy", "unhealthy"}, response.Status)
	assert.NotNil(t, response.Services)
	assert.Contains(t, response.Services, "mongodb")
	assert.Contains(t, response.Services, "redis")
	assert.Contains(t, response.Services, "rabbitmq")

	for serviceName, service := range response.Services {
		assert.Contains(t, []string{"healthy", "unhealthy"}, service.Status,
			"Service %s should have valid status", serviceName)
		assert.NotEmpty(t, service.Message,
			"Service %s should have a message", serviceName)
	}
}

func TestHealthHandler_checkMongoDB_NilDatabase(t *testing.T) {
	server := &serverHttp.Server{
		DB: nil,
	}
	handler := NewHealthHandler(server)

	result := handler.checkMongoDB(context.Background())

	assert.Equal(t, "unhealthy", result.Status)
	assert.Contains(t, result.Message, "not initialized")
}

func TestHealthHandler_checkRedis_NilRedis(t *testing.T) {
	server := &serverHttp.Server{
		Redis: nil,
	}
	handler := NewHealthHandler(server)

	result := handler.checkRedis(context.Background())

	assert.Equal(t, "unhealthy", result.Status)
	assert.Contains(t, result.Message, "not initialized")
}

func TestHealthHandler_checkRabbitMQ_NotConfigured(t *testing.T) {
	server := &serverHttp.Server{
		Config: &config.Config{
			EDA: config.EDAConfig{
				RabbitmqURL: "",
			},
		},
	}
	handler := NewHealthHandler(server)

	result := handler.checkRabbitMQ(context.Background())

	assert.Equal(t, "healthy", result.Status)
	assert.Contains(t, result.Message, "not configured")
}

func TestHealthHandler_checkRabbitMQ_InvalidURL(t *testing.T) {
	server := &serverHttp.Server{
		Config: &config.Config{
			EDA: config.EDAConfig{
				RabbitmqURL: "invalid-url",
			},
		},
	}
	handler := NewHealthHandler(server)

	result := handler.checkRabbitMQ(context.Background())

	assert.Equal(t, "unhealthy", result.Status)
	assert.Contains(t, result.Message, "Failed to connect to RabbitMQ")
}
