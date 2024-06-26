package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
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

type CreateCodeActionResponse struct {
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

func (h *CodeHandler) Create(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ca := new(CreateCodeActionRequest)
	if err := c.Bind(ca); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	t := code.CodeType(ca.Type)
	if err := t.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	lang := code.LanguageType(ca.Language)
	if err := lang.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	newCode, err := h.codeService.Create(
		ctx,
		code.NewCodeAction(ca.Name, ca.Source, lang, t, ca.URL, ca.ProjectUUID))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	response := CreateCodeActionResponse{
		ID: newCode.ID,

		Name:        newCode.Name,
		Source:      newCode.Source,
		Language:    string(newCode.Language),
		Type:        string(newCode.Type),
		URL:         newCode.URL,
		ProjectUUID: newCode.ProjectUUID,

		CreatedAt: newCode.CreatedAt.Format(time.DateTime),
		UpdatedAt: newCode.UpdatedAt.Format(time.DateTime),
	}
	return c.JSON(http.StatusCreated, response)
}

func (h *CodeHandler) Get(c echo.Context) error {
	codeID := c.Param("id")
	if codeID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("valid id is required").Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	codeAction, err := h.codeService.GetByID(ctx, codeID)
	if err != nil {
		if codeAction == nil {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, codeAction)
}

func (h *CodeHandler) Find(c echo.Context) error {
	projectUUID := c.QueryParam("project_uuid")
	if projectUUID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("valid project_uuid is required").Error())
	}
	codeType := c.QueryParam("code_type")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	codeAction, err := h.codeService.ListProjectCodes(ctx, projectUUID, codeType)
	if err != nil {
		if codeAction == nil {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, codeAction)
}

func (h *CodeHandler) Update(c echo.Context) error {
	codeID := c.Param("id")
	if codeID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("valid id is required").Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sca := new(SaveCodeActionRequest)
	if err := c.Bind(sca); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	cd, err := h.codeService.Update(ctx, codeID, sca.Name, sca.Source, sca.Type)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "failed to save code").Error())
	}

	response := CreateCodeActionResponse{
		ID: cd.ID,

		Name:     cd.Name,
		Source:   cd.Source,
		Language: string(cd.Language),
		Type:     string(cd.Type),
		URL:      cd.URL,

		CreatedAt: cd.CreatedAt.Format(time.DateTime),
		UpdatedAt: cd.UpdatedAt.Format(time.DateTime),
	}
	return c.JSON(http.StatusOK, response)
}

func (h *CodeHandler) Delete(c echo.Context) error {
	codeID := c.Param("id")
	if codeID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("valid id is required").Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := h.codeService.Delete(ctx, codeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusOK)
}
