package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-module/carbon/v2"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/weni-ai/flows-code-actions/internal/coderun"
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
		err := errors.New("valid id is required")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	codeRun, err := h.codeRunService.GetByID(ctx, codeRunID)
	if err != nil {
		log.WithError(err).Error(err.Error())
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
		err := errors.New("valid code_id is required")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	filter := map[string]interface{}{}
	afterp := c.QueryParam("after")
	if afterp != "" {
		after := carbon.Parse(afterp)
		if !after.IsValid() {
			err := errors.New("invalid after parameter")
			log.WithError(err).Error(err.Error())
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		} else {
			filter["after"] = after.StdTime()
		}
	}
	beforep := c.QueryParam("before")
	if beforep != "" {
		before := carbon.Parse(beforep)
		if !before.IsValid() {
			err := errors.New("invalid before parameter")
			log.WithError(err).Error(err.Error())
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		} else {
			filter["before"] = before.StdTime()
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	codeRuns, err := h.codeRunService.ListByCodeID(ctx, codeID, filter)
	if err != nil {
		log.WithError(err).Error(err.Error())
		if codeRuns == nil {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, codeRuns)
}
