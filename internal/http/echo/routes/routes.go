package routes

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/weni-ai/flows-code-actions/internal/code"
	codeRepoMongo "github.com/weni-ai/flows-code-actions/internal/code/mongodb"
	codelibRepoMongo "github.com/weni-ai/flows-code-actions/internal/codelib/mongodb"
	"github.com/weni-ai/flows-code-actions/internal/codelog"
	codelogRepoMongo "github.com/weni-ai/flows-code-actions/internal/codelog/mongodb"
	"github.com/weni-ai/flows-code-actions/internal/coderun"
	coderunRepoMongo "github.com/weni-ai/flows-code-actions/internal/coderun/mongodb"
	"github.com/weni-ai/flows-code-actions/internal/coderunner"
	s "github.com/weni-ai/flows-code-actions/internal/http/echo"
	"github.com/weni-ai/flows-code-actions/internal/http/echo/handlers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Setup(server *s.Server) {
	healthHandler := handlers.NewHealthHandler(server)

	codeRepo := codeRepoMongo.NewCodeRepository(server.DB)
	codelibRepo := codelibRepoMongo.NewCodeLibRepo(server.DB)
	codeService := code.NewCodeService(codeRepo, codelibRepo)
	codeHandler := handlers.NewCodeHandler(codeService)

	coderunRepo := coderunRepoMongo.NewCodeRunRepository(server.DB)
	coderunService := coderun.NewCodeRunService(coderunRepo)
	coderunHandler := handlers.NewCodeRunHandler(coderunService)

	codelogRepo := codelogRepoMongo.NewCodeLogRepository(server.DB)
	codelogService := codelog.NewCodeLogService(codelogRepo)
	codelogHandler := handlers.NewCodeLogHandler(codelogService)

	coderunnerService := coderunner.NewCodeRunnerService(server.Config, coderunService, codelogService)
	coderunnerHandler := handlers.NewCodeRunnerHandler(codeService, coderunnerService)

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

	server.Echo.GET("/", healthHandler.Health)
	server.Echo.GET("/health", healthHandler.Health)

	server.Echo.POST("/admin/code", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.CreateByAdmin))
	server.Echo.PATCH("/admin/code/:id", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.UpdateByAdmin))
	server.Echo.POST("/code", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.Create))
	server.Echo.GET("/code", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.Find))
	server.Echo.GET("/code/:id", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.Get))
	server.Echo.PATCH("/code/:id", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.Update))
	server.Echo.DELETE("/code/:id", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.Delete))

	server.Echo.GET("/coderun/:id", handlers.ProtectEndpointWithAuthToken(server.Config, coderunHandler.Get))
	server.Echo.GET("/coderun", handlers.ProtectEndpointWithAuthToken(server.Config, coderunHandler.Find))

	server.Echo.GET("/codelog/:id", handlers.ProtectEndpointWithAuthToken(server.Config, codelogHandler.Get))
	server.Echo.GET("/codelog", handlers.ProtectEndpointWithAuthToken(server.Config, codelogHandler.Find))

	server.Echo.POST("/run/:code_id", handlers.RequireAuthToken(server.Config, coderunnerHandler.RunCode))
	server.Echo.Any("/endpoint/:code_id", coderunnerHandler.RunEndpoint)

	server.Echo.Any("/action/endpoint/:code_id", coderunnerHandler.ActionEndpoint)
}
