package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/weni-ai/flows-code-actions/internal/secrets"
)

type SecretHandler struct {
	secretService secrets.UseCase
}

type CreateSecretRequest struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	ProjectUUID string `json:"project_uuid"`
}

type UpdateSecretRequest struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type SecretResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	ProjectUUID string `json:"project_uuid"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ParseSecretToResponse(secret *secrets.Secret) SecretResponse {
	return SecretResponse{
		ID:          secret.ID,
		Name:        secret.Name,
		Value:       secret.Value,
		ProjectUUID: secret.ProjectUUID,
		CreatedAt:   secret.CreatedAt,
		UpdatedAt:   secret.UpdatedAt,
	}
}

func NewSecretHandler(secretService secrets.UseCase) *SecretHandler {
	return &SecretHandler{
		secretService: secretService,
	}
}

func (h *SecretHandler) CreateSecret(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var req CreateSecretRequest
	if err := c.Bind(&req); err != nil {
		err = errors.Wrap(err, "failed to parse request body")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if req.Name == "" {
		err := errors.New("name is required")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if req.Value == "" {
		err := errors.New("value is required")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if req.ProjectUUID == "" {
		err := errors.New("project_uuid is required")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := CheckPermission(ctx, c, req.ProjectUUID); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	secret := secrets.NewSecret(req.Name, req.Value, req.ProjectUUID)
	newSecret, err := h.secretService.Create(ctx, secret)
	if err != nil {
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, ParseSecretToResponse(newSecret))
}

func (h *SecretHandler) GetSecret(c echo.Context) error {
	secretID := c.Param("id")
	if secretID == "" {
		err := errors.New("valid id is required")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	secret, err := h.secretService.GetByID(ctx, secretID)
	if err != nil {
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	if err := CheckPermission(ctx, c, secret.ProjectUUID); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	return c.JSON(http.StatusOK, ParseSecretToResponse(secret))
}

func (h *SecretHandler) FindSecretsByProjectUUID(c echo.Context) error {
	projectUUID := c.QueryParam("project_uuid")
	if projectUUID == "" {
		err := errors.New("valid project_uuid is required")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := CheckPermission(ctx, c, projectUUID); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	secretList, err := h.secretService.GetByProjectUUID(ctx, projectUUID)
	if err != nil {
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	responses := make([]SecretResponse, len(secretList))
	for i, s := range secretList {
		responses[i] = ParseSecretToResponse(&s)
	}

	return c.JSON(http.StatusOK, responses)
}

func (h *SecretHandler) UpdateSecret(c echo.Context) error {
	secretID := c.Param("id")
	if secretID == "" {
		err := errors.New("valid id is required")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get secret to check permission
	existingSecret, err := h.secretService.GetByID(ctx, secretID)
	if err != nil {
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	if err := CheckPermission(ctx, c, existingSecret.ProjectUUID); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	var req UpdateSecretRequest
	if err := c.Bind(&req); err != nil {
		err = errors.Wrap(err, "failed to parse request body")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	updatedSecret, err := h.secretService.Update(ctx, secretID, req.Name, req.Value)
	if err != nil {
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, ParseSecretToResponse(updatedSecret))
}

func (h *SecretHandler) DeleteSecret(c echo.Context) error {
	secretID := c.Param("id")
	if secretID == "" {
		err := errors.New("valid id is required")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get secret to check permission
	existingSecret, err := h.secretService.GetByID(ctx, secretID)
	if err != nil {
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	if err := CheckPermission(ctx, c, existingSecret.ProjectUUID); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	err = h.secretService.Delete(ctx, secretID)
	if err != nil {
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}
