package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/weni-ai/flows-code-actions/internal/secret"
)

type SecretHandler struct {
	secretService secret.UseCase
}

type CreateSecretRequest struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	ProjectUUID string `json:"project_uuid"`
}

type SecretResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Value       string    `json:"value"`
	ProjectUUID string    `json:"project_uuid"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewSecretHandler(service secret.UseCase) *SecretHandler {
	return &SecretHandler{secretService: service}
}

func (h *SecretHandler) CreateSecret(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var req CreateSecretRequest
	if err := c.Bind(&req); err != nil {
		log.WithError(err).Error("failed to bind request")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := CheckPermission(ctx, c, req.ProjectUUID); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	newSecret, err := h.secretService.Create(ctx, &secret.Secret{
		Name:        req.Name,
		Value:       req.Value,
		ProjectUUID: req.ProjectUUID,
	})
	if err != nil {
		log.WithError(err).Error("failed to create secret")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, SecretResponse{
		ID:          newSecret.ID.Hex(),
		Name:        newSecret.Name,
		Value:       newSecret.Value,
		ProjectUUID: newSecret.ProjectUUID,
		CreatedAt:   newSecret.CreatedAt,
		UpdatedAt:   newSecret.UpdatedAt,
	})
}

func (h *SecretHandler) GetSecret(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	secretID := c.Param("id")
	if secretID == "" {
		err := errors.New("valid id is required")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	secret, err := h.secretService.GetByID(ctx, secretID)
	if err != nil {
		log.WithError(err).Error("failed to get secret")
		if secret == nil {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := CheckPermission(ctx, c, secret.ProjectUUID); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	return c.JSON(http.StatusOK, SecretResponse{
		ID:          secret.ID.Hex(),
		Name:        secret.Name,
		Value:       secret.Value,
		ProjectUUID: secret.ProjectUUID,
		CreatedAt:   secret.CreatedAt,
		UpdatedAt:   secret.UpdatedAt,
	})
}

func (h *SecretHandler) ListSecrets(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	projectUUID := c.QueryParam("project_uuid")
	if projectUUID == "" {
		err := errors.New("valid project_uuid is required")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := CheckPermission(ctx, c, projectUUID); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	secrets, err := h.secretService.ListProjectSecrets(ctx, projectUUID)
	if err != nil {
		log.WithError(err).Error("failed to list secrets")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	response := make([]SecretResponse, len(secrets))
	for i, secret := range secrets {
		response[i] = SecretResponse{
			ID:          secret.ID.Hex(),
			Name:        secret.Name,
			Value:       secret.Value,
			ProjectUUID: secret.ProjectUUID,
			CreatedAt:   secret.CreatedAt,
			UpdatedAt:   secret.UpdatedAt,
		}
	}

	return c.JSON(http.StatusOK, response)
}

func (h *SecretHandler) DeleteSecret(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	secretID := c.Param("id")
	if secretID == "" {
		err := errors.New("valid id is required")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	secret, err := h.secretService.GetByID(ctx, secretID)
	if err != nil {
		return err
	}

	if err := CheckPermission(ctx, c, secret.ProjectUUID); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	if err := h.secretService.Delete(ctx, secretID); err != nil {
		log.WithError(err).Error("failed to delete secret")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *SecretHandler) UpdateSecret(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	secretID := c.Param("id")
	if secretID == "" {
		err := errors.New("valid id is required")
		log.WithError(err).Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var req CreateSecretRequest
	if err := c.Bind(&req); err != nil {
		log.WithError(err).Error("failed to bind request")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	existingSecret, err := h.secretService.GetByID(ctx, secretID)
	if err != nil {
		return err
	}

	if err := CheckPermission(ctx, c, existingSecret.ProjectUUID); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	updatedSecret, err := h.secretService.Update(ctx, secretID, req.Name, req.Value)
	if err != nil {
		log.WithError(err).Error("failed to update secret")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, SecretResponse{
		ID:          updatedSecret.ID.Hex(),
		Name:        updatedSecret.Name,
		Value:       updatedSecret.Value,
		ProjectUUID: updatedSecret.ProjectUUID,
		CreatedAt:   updatedSecret.CreatedAt,
		UpdatedAt:   updatedSecret.UpdatedAt,
	})
}
