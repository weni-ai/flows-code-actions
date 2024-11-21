package handlers

import (
	"net/http"

	serverHttp "github.com/weni-ai/flows-code-actions/internal/http/echo"

	"github.com/labstack/echo/v4"
)

type HealthHandler struct {
	server *serverHttp.Server
}

func NewHealthHandler(server *serverHttp.Server) *HealthHandler {
	return &HealthHandler{server: server}
}

func (h HealthHandler) Health(c echo.Context) error {
	return c.String(http.StatusOK, "OK\n")
}
