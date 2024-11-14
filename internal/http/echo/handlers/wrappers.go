package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/config"
	server "github.com/weni-ai/flows-code-actions/internal/http/echo"
	"github.com/weni-ai/flows-code-actions/internal/permission"
)

func RequireAuthToken(conf *config.Config, next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		auth := c.Request().Header.Get("Authorization")
		if conf.AuthToken != "" && fmt.Sprintf("Token %s", conf.AuthToken) != auth {
			return errors.New("invalid or missing authorization token")
		}
		return next(c)
	}
}

func ProtectEndpointWithAuthToken(conf *config.Config, next echo.HandlerFunc, perm permission.PermissionAccess) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !conf.OIDC.AuthEnabled {
			return next(c)
		}
		auth := c.Request().Header.Get("Authorization")
		if auth == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}
		authSplit := strings.Split(auth, " ")
		var token string
		token = authSplit[0]
		if len(authSplit) > 1 {
			token = authSplit[1]
		}

		url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/userinfo", conf.OIDC.Host, conf.OIDC.Realm)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return errors.Wrap(err, "error on create request to get userinfo")
		}
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return errors.Wrap(err, "error on request userinfor")
		}
		if resp.StatusCode == 401 {
			return echo.NewHTTPError(http.StatusUnauthorized, "not authorized")
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "error on read request body")
		}
		if resp.StatusCode != 200 {
			log.Info(string(body))
			return errors.New("error on get userinfo")
		}
		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal userinfo")
		}

		c.Set("check_permission", true)
		c.Set("user_email", result["email"])
		c.Set("perm", string(perm))

		return next(c)
	}
}

func CheckPermission(ctx context.Context, c echo.Context, projectUUID string) error {
	checkPermission := c.Get("check_permission")
	if checkPermission != nil && checkPermission.(bool) {
		perm, ok := c.Get("perm").(string)
		if !ok {
			return errors.New("invalid permission access to check")
		}
		return server.CheckPermission(ctx, c, projectUUID, permission.PermissionAccess(perm))
	}
	return nil
}
