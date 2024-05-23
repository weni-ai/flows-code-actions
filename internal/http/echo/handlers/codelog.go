package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/codelog"
)

type CodeLogHandler struct {
	codelogService codelog.UseCase
}

func NewCodeLogHandler(service codelog.UseCase) *CodeLogHandler {
	return &CodeLogHandler{codelogService: service}
}

func (h *CodeLogHandler) Get(c echo.Context) error {
	codelogID := c.Param("id")
	if codelogID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("valid id is required").Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	codelog, err := h.codelogService.GetByID(ctx, codelogID)
	if err != nil {
		if codelog == nil {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, codelog)
}

func (h *CodeLogHandler) Find(c echo.Context) error {
	runID := c.QueryParam("run_id")
	if runID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("valid run_id is required"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	codeLogs, err := h.codelogService.ListRunLogs(ctx, runID)
	if err != nil {
		if codeLogs == nil {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, codeLogs)
}
