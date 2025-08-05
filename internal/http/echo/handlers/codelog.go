package handlers

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/weni-ai/flows-code-actions/internal/codelog"
	"github.com/weni-ai/flows-code-actions/internal/coderun"
)

type CodeLogHandler struct {
	codelogService codelog.UseCase
	coderunService coderun.UseCase
}

func NewCodeLogHandler(codelogService codelog.UseCase, coderunService coderun.UseCase) *CodeLogHandler {
	return &CodeLogHandler{
		codelogService: codelogService,
		coderunService: coderunService,
	}
}

func (h *CodeLogHandler) Get(c echo.Context) error {
	codelogID := c.Param("id")
	if codelogID == "" {
		err := errors.New("valid id is required")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	codelog, err := h.codelogService.GetByID(ctx, codelogID)
	if err != nil {
		log.WithError(err).Error(err.Error())
		if codelog == nil {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, codelog)
}

const (
	perPage = 20
)

func (h *CodeLogHandler) Find(c echo.Context) error {
	runID := c.QueryParam("run_id")
	codeID := c.QueryParam("code_id")
	qpage := c.QueryParam("page")
	if runID == "" && codeID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("valid run_id or code_id is required"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if codeID == "" {
		codeRun, err := h.coderunService.GetByID(ctx, runID)
		if err != nil {
			log.WithError(err).Error(err.Error())
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		codeID = codeRun.CodeID.Hex()
	}

	if qpage == "" {
		qpage = "1"
	}
	page, _ := strconv.Atoi(qpage)
	codeLogs, err := h.codelogService.ListRunLogs(ctx, runID, codeID, perPage, page)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return c.JSON(http.StatusOK, newCodeLogResponse([]codelog.CodeLog{}, 0, page))
		}
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ctx, cancel = context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	total, err := h.codelogService.Count(ctx, runID, codeID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return c.JSON(http.StatusRequestTimeout, err.Error())
		}
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, newCodeLogResponse(codeLogs, total, page))
}

type CodeLogResponse struct {
	Data     []codelog.CodeLog `json:"data"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	LastPage int               `json:"last_page"`
}

func newCodeLogResponse(codeLogs []codelog.CodeLog, total int64, page int) CodeLogResponse {
	return CodeLogResponse{
		Data:     codeLogs,
		Total:    total,
		Page:     page,
		LastPage: int(math.Ceil(float64(total) / float64(perPage))),
	}
}
