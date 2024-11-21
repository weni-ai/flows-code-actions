package handlers

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/weni-ai/flows-code-actions/internal/code"
)

type CodeHandler struct {
	codeService code.UseCase
}

type CreateCodeActionRequest struct {
	Name        string `json:"name,omitempty"`
	Source      string `json:"source,omitempty"`
	Language    string `json:"language,omitempty"`
	Type        string `json:"type,omitempty"`
	ProjectUUID string `json:"project_uuid,omitempty"`
	URL         string `json:"url,omitempty"`
}

type CodeActionResponse struct {
	ID string `json:"id,omitempty"`

	Name        string `json:"name,omitempty"`
	Source      string `json:"source,omitempty"`
	Language    string `json:"language,omitempty"`
	Type        string `json:"type,omitempty"`
	ProjectUUID string `json:"project_uuid,omitempty"`
	URL         string `json:"url,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func ParseCodeToResponse(newCode *code.Code) CodeActionResponse {
	return CodeActionResponse{
		ID: newCode.ID.Hex(),

		Name:        newCode.Name,
		Source:      newCode.Source,
		Language:    string(newCode.Language),
		Type:        string(newCode.Type),
		URL:         newCode.URL,
		ProjectUUID: newCode.ProjectUUID,

		CreatedAt: newCode.CreatedAt,
		UpdatedAt: newCode.UpdatedAt,
	}
}

type SaveCodeActionRequest struct {
	Name   string `json:"name,omitempty"`
	Source string `json:"source,omitempty"`
	Type   string `json:"type,omitempty"`
}

type SaveCodeActionResponse struct {
	ID string `json:"id,omitempty"`

	Name        string `json:"name,omitempty"`
	Source      string `json:"source,omitempty"`
	Language    string `json:"language,omitempty"`
	Type        string `json:"type,omitempty"`
	ProjectUUID string `json:"project_uuid,omitempty"`
	URL         string `json:"url,omitempty"`

	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func NewCodeHandler(service code.UseCase) *CodeHandler {
	return &CodeHandler{codeService: service}
}

func (h *CodeHandler) CreateCode(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ca := &code.Code{}
	qp := c.QueryParams()
	ca.ProjectUUID = qp.Get("project_uuid")

	if err := CheckPermission(ctx, c, ca.ProjectUUID); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	ca.Name = qp.Get("name")
	ca.Language = code.LanguageType(qp.Get("language"))
	ca.Type = code.CodeType(qp.Get("type"))

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		err = errors.Wrap(err, "failed to read body")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ca.Source = string(body)

	t := code.CodeType(ca.Type)
	if err := t.Validate(); err != nil {
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	lang := code.LanguageType(ca.Language)
	if err := lang.Validate(); err != nil {
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	newCode, err := h.codeService.Create(
		ctx,
		code.NewCodeAction(ca.Name, ca.Source, lang, t, ca.URL, ca.ProjectUUID))
	if err != nil {
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, newCode)
}

func (h *CodeHandler) UpdateCode(c echo.Context) error {
	codeID := c.Param("id")
	if codeID == "" {
		err := errors.New("valid id is required")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ca := &code.Code{}
	qp := c.QueryParams()
	ca.Name = qp.Get("name")
	ca.Language = code.LanguageType(qp.Get("language"))
	ca.Type = code.CodeType(qp.Get("type"))

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		err := errors.Wrap(err, "failed to read body")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ca.Source = string(body)

	t := code.CodeType(ca.Type)
	if err := t.Validate(); err != nil {
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	lang := code.LanguageType(ca.Language)
	if err := lang.Validate(); err != nil {
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	uc, err := h.codeService.GetByID(ctx, codeID)
	if err != nil {
		return err
	}
	if err := CheckPermission(ctx, c, uc.ProjectUUID); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	cd, err := h.codeService.Update(
		ctx,
		codeID, ca.Name, ca.Source, string(ca.Type))
	if err != nil {
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, cd)
}

func (h *CodeHandler) Get(c echo.Context) error {
	codeID := c.Param("id")
	if codeID == "" {
		err := errors.New("valid id is required")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	codeAction, err := h.codeService.GetByID(ctx, codeID)
	if err != nil {
		log.WithError(err).Error(err.Error())
		if codeAction == nil {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := CheckPermission(ctx, c, codeAction.ProjectUUID); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	response := ParseCodeToResponse(codeAction)
	return c.JSON(http.StatusOK, response)
}

func (h *CodeHandler) Find(c echo.Context) error {
	projectUUID := c.QueryParam("project_uuid")
	if projectUUID == "" {
		err := errors.New("valid project_uuid is required")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	codeType := c.QueryParam("code_type")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := CheckPermission(ctx, c, projectUUID); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	codeActions, err := h.codeService.ListProjectCodes(ctx, projectUUID, codeType)
	if err != nil {
		log.WithError(err).Error(err.Error())
		if codeActions == nil {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, codeActions)
}

func (h *CodeHandler) Delete(c echo.Context) error {
	codeID := c.Param("id")
	if codeID == "" {
		err := errors.New("valid id is required")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	uc, err := h.codeService.GetByID(ctx, codeID)
	if err != nil {
		return err
	}
	if err := CheckPermission(ctx, c, uc.ProjectUUID); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	err = h.codeService.Delete(ctx, codeID)
	if err != nil {
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusOK)
}
