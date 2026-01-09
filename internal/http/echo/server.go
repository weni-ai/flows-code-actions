package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"
	"github.com/weni-ai/flows-code-actions/config"
	"github.com/weni-ai/flows-code-actions/internal/codelog"
	"github.com/weni-ai/flows-code-actions/internal/coderun"
	"github.com/weni-ai/flows-code-actions/internal/permission"
	"go.mongodb.org/mongo-driver/mongo"

	log "github.com/sirupsen/logrus"

	"github.com/labstack/echo/v4"
)

var Permission *EchoPermissionHandler = nil

type EchoPermissionHandler struct {
	permissionService permission.UserPermissionUseCase
}

func NewEchoPermissionHandler(permissionService permission.UserPermissionUseCase) *EchoPermissionHandler {
	return &EchoPermissionHandler{permissionService: permissionService}
}

func CheckPermission(ctx context.Context, c echo.Context, projectUUID string, permissionRole permission.PermissionAccess) error {
	if Permission == nil {
		log.Info("auth permissions disabled")
		return nil
	}
	err := Permission.CheckPermission(ctx, c, projectUUID, permissionRole)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, err)
	}
	return nil
}

func (s *EchoPermissionHandler) CheckPermission(ctx context.Context, c echo.Context, projectUUID string, permissionRole permission.PermissionAccess) error {
	email := c.Get("user_email").(string)
	userPermission, err := s.permissionService.Find(ctx, &permission.UserPermission{ProjectUUID: projectUUID, Email: email})
	if err != nil {
		return errors.New("have'nt permission to access this resource")
	}
	allow := permission.HasPermission(userPermission, permissionRole)
	if !allow {
		return errors.New("have'nt permission to access this resource")
	}
	return nil
}

type Server struct {
	Echo     *echo.Echo
	Config   *config.Config
	DB       *mongo.Database // MongoDB connection (legacy)
	SQLDB    *sql.DB         // PostgreSQL connection
	Redis    *redis.Client
	Locker   *redislock.Client
	Services *Services
}

type Services struct {
	CodeLogService codelog.UseCase
	CodeRunService coderun.UseCase
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		Echo:     echo.New(),
		Config:   cfg,
		Services: &Services{},
	}
}

func (server *Server) Start(addr string) error {
	return server.Echo.Start(":" + addr)
}

func (server *Server) Stop(ctx context.Context) error {
	return server.Echo.Shutdown(ctx)
}

var minIntervalLock = time.Hour * 1

func (server *Server) StartCodeLogCleaner(ctx context.Context, cfg *config.Config) error {
	taskkey := "codelogcleaner"
	_, err := server.Locker.Obtain(ctx, taskkey, minIntervalLock, nil)
	if err != nil {
		log.Println("already has lock for ", taskkey)
		return nil
	}
	err = server.Services.CodeLogService.StartCodeLogCleaner(cfg)
	if err != nil {
		return err
	}
	return nil
}

func (server *Server) StartCodeRunCleaner(ctx context.Context, cfg *config.Config) error {
	taskkey := "coderuncleaner"
	_, err := server.Locker.Obtain(ctx, taskkey, minIntervalLock, nil)
	if err != nil {
		log.Println("already has lock for ", taskkey)
		return nil
	}
	err = server.Services.CodeRunService.StartCodeRunCleaner(cfg)
	if err != nil {
		return err
	}
	return nil
}
