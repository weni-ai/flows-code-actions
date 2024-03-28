package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/weni-ai/code-actions/internal/coderun"
)

type CodeRunHandler struct {
	codeRunService coderun.UseCase
}

func NewCodeRunHandler(service coderun.UseCase) *CodeRunHandler {
	return &CodeRunHandler{codeRunService: service}
}

func (h *CodeRunHandler) Get(c echo.Context) error {
	codeRunID := c.Param("id")
	if codeRunID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("valid id is required").Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	codeRun, err := h.codeRunService.GetByID(ctx, codeRunID)
	if err != nil {
		if codeRun == nil {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, codeRun)
}

func (h *CodeRunHandler) Find(c echo.Context) error {
	codeID := c.QueryParam("code_id")
	if codeID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("valid code_id is required").Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	codeAction, err := h.codeRunService.ListByCodeID(ctx, codeID)
	if err != nil {
		if codeAction == nil {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, codeAction)
}
