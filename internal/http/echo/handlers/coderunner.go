package handlers

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/code"
	"github.com/weni-ai/flows-code-actions/internal/coderunner"
	"github.com/weni-ai/flows-code-actions/internal/metrics"
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

	result, err := h.coderunnerService.RunCode(ctx, codeID, codeAction.Source, string(codeAction.Language), nil, "")
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

	result, err := h.coderunnerService.RunCode(ctx, codeID, codeAction.Source, string(codeAction.Language), nil, "")
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

	res := make(chan error)

	go func() {
		cparams := map[string]interface{}{}

		for k, v := range c.QueryParams() {
			cparams[k] = v[0]
		}

		abody, err := io.ReadAll(c.Request().Body)
		if err != nil {
			res <- echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		result, err := h.coderunnerService.RunCode(ctx, codeID, codeAction.Source, string(codeAction.Language), cparams, string(abody))
		if err != nil {
			res <- echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		statusCode := http.StatusOK
		if sc, err := result.StatusCode(); err == nil {
			statusCode = sc
		}

		contentType := result.ResponseContentType()
		switch contentType {
		case "json":
			c.Response().Header().Set("Content-Type", "application/json; charset=UTF-8")
		case "html":
			c.Response().Header().Set("Content-Type", "text/html; charset=UTF-8")
		default:
			c.Response().Header().Set("Content-Type", "text/plain; charset=UTF-8")
		}

		c.Response().WriteHeader(statusCode)
		_, err = c.Response().Write([]byte(result.Result))
		if err != nil {
			res <- echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		res <- nil
	}()

	select {
	case err := <-res:
		return err
	case <-ctx.Done():
		return echo.NewHTTPError(http.StatusRequestTimeout, "timeout: request context timeout limit exceeded")
	}
}
