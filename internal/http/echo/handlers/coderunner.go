package handlers

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/code"
	"github.com/weni-ai/flows-code-actions/internal/coderun"
	"github.com/weni-ai/flows-code-actions/internal/coderunner"
	"github.com/weni-ai/flows-code-actions/internal/metrics"
	"github.com/weni-ai/flows-code-actions/internal/workerpool"
)

type CodeRunnerHandler struct {
	codeService       code.UseCase
	coderunnerService coderunner.UseCase
	workerPool        *workerpool.Pool
}

func NewCodeRunnerHandler(codeService code.UseCase, coderunnerService coderunner.UseCase, workerPool *workerpool.Pool) *CodeRunnerHandler {
	return &CodeRunnerHandler{
		codeService:       codeService,
		coderunnerService: coderunnerService,
		workerPool:        workerPool,
	}
}

func (h *CodeRunnerHandler) RunCode(c echo.Context) error {
	codeID := c.Param("code_id")
	if codeID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("valid id is required").Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	codeAction, err := h.codeService.GetByID(ctx, codeID)
	if err != nil {
		if codeAction == nil {
			return echo.NewHTTPError(http.StatusNotFound, "code not found")
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	result, err := h.coderunnerService.RunCode(ctx, codeID, codeAction.Source, string(codeAction.Language), nil, "", nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	runResult := map[string]interface{}{
		"code_id": codeID,
		"result":  result,
	}

	return c.JSON(http.StatusOK, runResult)
}

func (h *CodeRunnerHandler) RunEndpoint(c echo.Context) error {
	codeID := c.Param("code_id")
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

	result, err := h.coderunnerService.RunCode(ctx, codeID, codeAction.Source, string(codeAction.Language), nil, "", nil)
	if err != nil {
		echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, result.Result)
}

func (h *CodeRunnerHandler) ActionEndpoint(c echo.Context) error {
	start := time.Now()

	codeID := c.Param("code_id")
	if codeID == "" {
		return echo.NewHTTPError(http.StatusNotFound, errors.New("Not Found"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	codeAction, err := h.codeService.GetByID(ctx, codeID)
	if err != nil {
		if codeAction == nil || codeAction.Type == code.TypeFlow {
			return echo.NewHTTPError(http.StatusNotFound, errors.New("Not Found"))
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	defer func() {
		metrics.CodeRunElapsed(codeAction.ProjectUUID, codeID, time.Since(start).Seconds())
		metrics.AddCodeRunCount(codeAction.ProjectUUID, codeID, 1)
	}()

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*time.Duration(codeAction.Timeout))
	defer cancel()

	aheader := map[string]interface{}{}
	for k, v := range c.Request().Header {
		aheader[k] = v
	}

	cparams := map[string]interface{}{}
	for k, v := range c.QueryParams() {
		if len(v) > 0 {
			cparams[k] = v[0]
		}
	}

	abody, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resultCh := make(chan workerpool.Result, 1)
	task := workerpool.Task{
		Ctx: ctx,
		Execute: func(taskCtx context.Context) (*coderun.CodeRun, error) {
			return h.coderunnerService.RunCode(taskCtx, codeID, codeAction.Source, string(codeAction.Language), cparams, string(abody), aheader)
		},
		Result: resultCh,
	}

	if err := h.workerPool.Submit(task); err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, err.Error())
	}

	select {
	case res := <-resultCh:
		if res.Err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, res.Err.Error())
		}

		result := res.Run
		statusCode := http.StatusOK
		contentType := "string"

		if result != nil {
			if sc, err := result.StatusCode(); err == nil {
				statusCode = sc
			}
			contentType = result.ResponseContentType()
		}
		switch contentType {
		case "json":
			c.Response().Header().Set("Content-Type", "application/json; charset=UTF-8")
		case "html":
			c.Response().Header().Set("Content-Type", "text/html; charset=UTF-8")
		default:
			c.Response().Header().Set("Content-Type", "text/plain; charset=UTF-8")
		}

		c.Response().WriteHeader(statusCode)

		var resultContent string
		if result != nil {
			resultContent = result.Result
		} else {
			resultContent = "Internal Server Error"
		}

		if _, err := c.Response().Write([]byte(resultContent)); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return nil
	case <-ctx.Done():
		return echo.NewHTTPError(http.StatusRequestTimeout, "timeout: request context timeout limit exceeded")
	}
}
