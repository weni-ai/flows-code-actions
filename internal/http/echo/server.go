package server

import (
	"context"

	"github.com/weni-ai/code-actions/config"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/labstack/echo/v4"
)

type Server struct {
	Echo   *echo.Echo
	Config *config.Config
	DB     *mongo.Database
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		Echo:   echo.New(),
		Config: cfg,
	}
}

func (server *Server) Start(addr string) error {
	return server.Echo.Start(":" + addr)
}

func (server *Server) Stop(ctx context.Context) error {
	return server.Echo.Shutdown(ctx)
}
