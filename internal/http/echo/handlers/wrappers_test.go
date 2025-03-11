package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/weni-ai/flows-code-actions/config"
	"github.com/weni-ai/flows-code-actions/internal/permission"
)

func TestProtectEndpointWithAuthToken(t *testing.T) {
	// Mock handler that always returns success
	mockHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}

	tests := []struct {
		name           string
		config         *config.Config
		authHeader     string
		mockUserInfo   *httptest.Server
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Auth disabled should pass through",
			config: &config.Config{
				OIDC: config.OIDCConfig{
					AuthEnabled: false,
				},
			},
			authHeader:     "",
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name: "Missing auth header should return unauthorized",
			config: &config.Config{
				OIDC: config.OIDCConfig{
					AuthEnabled: true,
					Host:        "http://localhost",
					Realm:       "test",
				},
			},
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Valid token should succeed",
			config: &config.Config{
				OIDC: config.OIDCConfig{
					AuthEnabled: true,
					Host:        "http://localhost",
					Realm:       "test",
				},
			},
			authHeader: "Bearer valid-token",
			mockUserInfo: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"email": "test@example.com"}`))
			})),
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name: "Invalid token should return unauthorized",
			config: &config.Config{
				OIDC: config.OIDCConfig{
					AuthEnabled: true,
					Host:        "http://localhost",
					Realm:       "test",
				},
			},
			authHeader: "Bearer invalid-token",
			mockUserInfo: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			})),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "UserInfo server error should return error",
			config: &config.Config{
				OIDC: config.OIDCConfig{
					AuthEnabled: true,
					Host:        "http://localhost",
					Realm:       "test",
				},
			},
			authHeader: "Bearer valid-token",
			mockUserInfo: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			})),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Can communicate internally should pass through",
			config: &config.Config{
				OIDC: config.OIDCConfig{
					AuthEnabled: true,
					Host:        "http://localhost",
					Realm:       "test",
				},
			},
			authHeader: "Bearer valid-token",
			mockUserInfo: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"email": "test@example.com", "can_communicate_internally": true}`))
			})),
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockUserInfo != nil {
				defer tt.mockUserInfo.Close()
				tt.config.OIDC.Host = tt.mockUserInfo.URL
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := ProtectEndpointWithAuthToken(tt.config, mockHandler, permission.ReadPermission)
			err := handler(c)

			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					if he.Code != tt.expectedStatus {
						t.Errorf("Expected status code %d, got %d", tt.expectedStatus, he.Code)
					}
				} else {
					t.Errorf("Expected echo.HTTPError, got %v", err)
				}
				return
			}

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedBody != "" && rec.Body.String() != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, rec.Body.String())
			}

			if tt.expectedStatus == http.StatusOK {
				if tt.config.OIDC.AuthEnabled {
					checkPerm := c.Get("check_permission")
					if checkPerm != nil && !checkPerm.(bool) {
						t.Error("Expected check_permission to be set to true")
					}
				}
			}
		})
	}
}
