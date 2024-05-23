package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/code"
	"github.com/weni-ai/flows-code-actions/internal/coderunner"
)

type CodeRunnerHandler struct {
	codeService       code.UseCase
	coderunnerService coderunner.UseCase
}

func NewCodeRunnerHandler(codeService code.UseCase, coderunnerService coderunner.UseCase) *CodeRunnerHandler {
	return &CodeRunnerHandler{
		codeService:       codeService,
		coderunnerService: coderunnerService,
	}
}

func (h *CodeRunnerHandler) RunCode(c echo.Context) error {
	codeID := c.Param("id")
	if codeID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("valid id is required").Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	codeAction, err := h.codeService.GetByID(ctx, codeID)
	if err != nil {
		if codeAction == nil {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	result, err := h.coderunnerService.RunCode(ctx, codeAction.Source, string(codeAction.Language))
	if err != nil {
		echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

func (h *CodeRunnerHandler) RunWebhook(c echo.Context) error {
	codeID := c.Param("id")
	if codeID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("valid id is required"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	codeAction, err := h.codeService.GetByID(ctx, codeID)
	if err != nil {
		if codeAction == nil || codeAction.Type == code.TypeFlow {
			return echo.NewHTTPError(http.StatusNotFound, errors.New("Not Found"))
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	result, err := h.coderunnerService.RunCode(ctx, codeAction.Source, string(codeAction.Language))
	if err != nil {
		echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}