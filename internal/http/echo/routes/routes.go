package routes

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/weni-ai/flows-code-actions/internal/code"
	codeRepoMongo "github.com/weni-ai/flows-code-actions/internal/code/mongodb"
	codeRepoPG "github.com/weni-ai/flows-code-actions/internal/code/pg"
	"github.com/weni-ai/flows-code-actions/internal/codelib"
	codelibRepoMongo "github.com/weni-ai/flows-code-actions/internal/codelib/mongodb"
	codelibRepoPG "github.com/weni-ai/flows-code-actions/internal/codelib/pg"
	"github.com/weni-ai/flows-code-actions/internal/codelog"
	codelogRepoMongo "github.com/weni-ai/flows-code-actions/internal/codelog/mongodb"
	"github.com/weni-ai/flows-code-actions/internal/coderun"
	coderunRepoMongo "github.com/weni-ai/flows-code-actions/internal/coderun/mongodb"
	coderunRepoPG "github.com/weni-ai/flows-code-actions/internal/coderun/pg"
	"github.com/weni-ai/flows-code-actions/internal/coderunner"
	s "github.com/weni-ai/flows-code-actions/internal/http/echo"
	"github.com/weni-ai/flows-code-actions/internal/http/echo/handlers"
	"github.com/weni-ai/flows-code-actions/internal/permission"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Setup(server *s.Server) {
	healthHandler := handlers.NewHealthHandler(server)

	// Setup repositories based on database type
	var codeRepo code.Repository
	var codelibRepo codelib.Repository
	var coderunRepo coderun.Repository
	var codelogRepo codelog.Repository

	if server.Config.DB.Type == "postgres" {
		// Use PostgreSQL repositories
		pgDB := server.SQLDB
		codeRepo = codeRepoPG.NewCodeRepository(pgDB)
		codelibRepo = codelibRepoPG.NewCodeLibRepo(pgDB)
		coderunRepo = coderunRepoPG.NewCodeRunRepository(pgDB)
	} else {
		// Use MongoDB repositories (default)
		mongoDB := server.DB
		codeRepo = codeRepoMongo.NewCodeRepository(mongoDB)
		codelibRepo = codelibRepoMongo.NewCodeLibRepo(mongoDB)
		coderunRepo = coderunRepoMongo.NewCodeRunRepository(mongoDB)
		codelogRepo = codelogRepoMongo.NewCodeLogRepository(mongoDB)
	}

	codeService := code.NewCodeService(server.Config, codeRepo, codelibRepo)
	codeHandler := handlers.NewCodeHandler(codeService)

	coderunService := coderun.NewCodeRunService(coderunRepo)
	coderunHandler := handlers.NewCodeRunHandler(coderunService)

	codelogService := codelog.NewCodeLogService(codelogRepo)
	codelogHandler := handlers.NewCodeLogHandler(codelogService, coderunService)

	server.Services.CodeLogService = codelogService
	server.Services.CodeRunService = coderunService

	coderunnerService := coderunner.NewCodeRunnerService(server.Config, coderunService, codelogService)
	coderunnerHandler := handlers.NewCodeRunnerHandler(codeService, coderunnerService)

	ratelimiter := s.NewRateLimiter(
		server.Redis,
		server.Config.RateLimiterCode.MaxRequests,
		time.Duration(server.Config.RateLimiterCode.Window)*time.Second,
	)

	log := logrus.New()
	server.Echo.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				log.WithFields(logrus.Fields{
					"URI":    v.URI,
					"status": v.Status,
				}).Info("request")
			} else {
				log.WithFields(logrus.Fields{
					"URI":    v.URI,
					"status": v.Status,
					"error":  v.Error,
				}).Error("request error")
			}
			return nil
		},
	}))

	server.Echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://drogasil-demo.netlify.app"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	server.Echo.GET("/", healthHandler.HealthCheck)
	server.Echo.GET("/health", healthHandler.Health)
	server.Echo.GET("/healthz", healthHandler.HealthCheck) // Simplified health check endpoint for monitoring

	server.Echo.POST("/admin/code", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.CreateCode, permission.WritePermission))
	server.Echo.PATCH("/admin/code/:id", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.UpdateCode, permission.WritePermission))
	server.Echo.POST("/code", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.CreateCode, permission.WritePermission))
	server.Echo.GET("/code", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.Find, permission.ReadPermission))
	server.Echo.GET("/code/:id", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.Get, permission.ReadPermission))
	server.Echo.PATCH("/code/:id", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.UpdateCode, permission.WritePermission))
	server.Echo.DELETE("/code/:id", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.Delete, permission.WritePermission))

	server.Echo.GET("/coderun/:id", handlers.ProtectEndpointWithAuthToken(server.Config, coderunHandler.Get, permission.ReadPermission))
	server.Echo.GET("/coderun", handlers.ProtectEndpointWithAuthToken(server.Config, coderunHandler.Find, permission.ReadPermission))

	server.Echo.GET("/codelog/:id", handlers.ProtectEndpointWithAuthToken(server.Config, codelogHandler.Get, permission.ReadPermission))
	server.Echo.GET("/codelog", handlers.ProtectEndpointWithAuthToken(server.Config, codelogHandler.Find, permission.ReadPermission))

	server.Echo.POST("/run/:code_id", handlers.RequireAuthToken(server.Config, coderunnerHandler.RunCode))
	server.Echo.Any("/endpoint/:code_id", coderunnerHandler.RunEndpoint)

	server.Echo.Any("/action/endpoint/:code_id", handlers.LimitByCodeIDMiddleware(coderunnerHandler.ActionEndpoint, *ratelimiter))

	server.Echo.Use(echoprometheus.NewMiddleware("codeactions"))

	server.Echo.GET("/metrics", echoprometheus.NewHandler())
}
