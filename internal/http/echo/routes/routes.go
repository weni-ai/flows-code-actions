package routes

import (
	"github.com/weni-ai/code-actions/internal/code"
	codeRepoMongo "github.com/weni-ai/code-actions/internal/code/mongodb"
	"github.com/weni-ai/code-actions/internal/coderun"
	coderunRepoMongo "github.com/weni-ai/code-actions/internal/coderun/mongodb"
	s "github.com/weni-ai/code-actions/internal/http/echo"
	"github.com/weni-ai/code-actions/internal/http/echo/handlers"

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

	server.Echo.Use(middleware.Logger())

	server.Echo.GET("/health", healthHandler.Health)

	server.Echo.POST("/code", codeHandler.Create)
	server.Echo.GET("/code", codeHandler.Find)
	server.Echo.GET("/code/:id", codeHandler.Get)
	server.Echo.PATCH("/code/:id", codeHandler.Update)
	server.Echo.DELETE("/code/:id", codeHandler.Delete)

	server.Echo.GET("/coderun/:id", coderunHandler.Get)
	server.Echo.GET("/coderun", coderunHandler.Find)
}
