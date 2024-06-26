package routes

import (
	"github.com/weni-ai/flows-code-actions/internal/code"
	codeRepoMongo "github.com/weni-ai/flows-code-actions/internal/code/mongodb"
	"github.com/weni-ai/flows-code-actions/internal/codelog"
	codelogRepoMongo "github.com/weni-ai/flows-code-actions/internal/codelog/mongodb"
	"github.com/weni-ai/flows-code-actions/internal/coderun"
	coderunRepoMongo "github.com/weni-ai/flows-code-actions/internal/coderun/mongodb"
	"github.com/weni-ai/flows-code-actions/internal/coderunner"
	s "github.com/weni-ai/flows-code-actions/internal/http/echo"
	"github.com/weni-ai/flows-code-actions/internal/http/echo/handlers"

	"github.com/labstack/echo/v4/middleware"
)

func Setup(server *s.Server) {
	healthHandler := handlers.NewHealthHandler(server)

	codeRepo := codeRepoMongo.NewCodeRepository(server.DB)
	codeService := code.NewCodeService(codeRepo)
	codeHandler := handlers.NewCodeHandler(codeService)

	coderunRepo := coderunRepoMongo.NewCodeRunRepository(server.DB)
	coderunService := coderun.NewCodeRunService(coderunRepo)
	coderunHandler := handlers.NewCodeRunHandler(coderunService)

	codelogRepo := codelogRepoMongo.NewCodeLogRepository(server.DB)
	codelogService := codelog.NewCodeLogService(codelogRepo)
	codelogHandler := handlers.NewCodeLogHandler(codelogService)

	coderunnerService := coderunner.NewCodeRunnerService(coderunService)
	coderunnerHandler := handlers.NewCodeRunnerHandler(codeService, coderunnerService)

	server.Echo.Use(middleware.Logger())

	server.Echo.GET("/health", healthHandler.Health)

	server.Echo.POST("/code", codeHandler.Create)
	server.Echo.GET("/code", codeHandler.Find)
	server.Echo.GET("/code/:id", codeHandler.Get)
	server.Echo.PATCH("/code/:id", codeHandler.Update)
	server.Echo.DELETE("/code/:id", codeHandler.Delete)

	server.Echo.GET("/coderun/:id", coderunHandler.Get)
	server.Echo.GET("/coderun", coderunHandler.Find)

	server.Echo.GET("/codelog/:id", codelogHandler.Get)
	server.Echo.GET("/codelog", codelogHandler.Find)

	server.Echo.POST("/run/:code_id", coderunnerHandler.RunCode)
	server.Echo.POST("/endpoint/:code_id", coderunnerHandler.RunEndpoint)
}
