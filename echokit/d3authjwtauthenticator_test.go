package echokit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestD3AuthJWTAuthenticator(t *testing.T) {
	t.Run("returns_not_authenticated_when_no_authorization_header", func(t *testing.T) {
		authenticator := &D3AuthJWTAuthenticator{}

		_, c, _ := makeD3AuthTestContext(http.MethodGet, "/")

		err := authenticator.AuthenticateRequest(c)

		assert.NoError(t, err)
		isAuth, _ := authenticator.IsAuthenticated(c)
		assert.False(t, isAuth)
	})

	t.Run("returns_not_authenticated_when_authorization_header_is_not_bearer", func(t *testing.T) {
		authenticator := &D3AuthJWTAuthenticator{}

		_, c, _ := makeD3AuthTestContext(http.MethodGet, "/")
		c.Request().Header.Set("Authorization", "Basic abc123")

		err := authenticator.AuthenticateRequest(c)

		assert.NoError(t, err)
		isAuth, _ := authenticator.IsAuthenticated(c)
		assert.False(t, isAuth)
	})

	t.Run("returns_authenticated_user_when_set_in_context", func(t *testing.T) {
		authenticator := &D3AuthJWTAuthenticator{}

		_, c, _ := makeD3AuthTestContext(http.MethodGet, "/")
		c.Set(d3AuthJWTAuthenticatorContextKey, &AuthenticatedUser{
			Sub:   "theSubject",
			Email: "theEmail@test.com",
		})

		isAuth, err := authenticator.IsAuthenticated(c)

		assert.NoError(t, err)
		assert.True(t, isAuth)

		user, err := authenticator.GetAuthenticatedUser(c)

		assert.NoError(t, err)
		assert.Equal(t, "theSubject", user.Sub)
		assert.Equal(t, "theEmail@test.com", user.Email)
	})

	t.Run("returns_error_from_GetAuthenticatedUser_when_not_authenticated", func(t *testing.T) {
		authenticator := &D3AuthJWTAuthenticator{}

		_, c, _ := makeD3AuthTestContext(http.MethodGet, "/")

		user, err := authenticator.GetAuthenticatedUser(c)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "no authenticated user", err.Error())
	})

	t.Run("HandleNotAuthenticated_returns_401", func(t *testing.T) {
		authenticator := &D3AuthJWTAuthenticator{}

		_, c, rec := makeD3AuthTestContext(http.MethodGet, "/")

		err := authenticator.HandleNotAuthenticated(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestNewD3AuthJWTAuthenticator(t *testing.T) {
	t.Run("returns_an_error_when_base_url_is_invalid", func(t *testing.T) {
		config := D3AuthConfig{
			BaseURL:  "://invalid",
			Audience: "anAudience",
		}

		authenticator, err := NewD3AuthJWTAuthenticator(config)

		assert.Nil(t, authenticator)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse d3-auth base URL")
	})

	t.Run("creates_authenticator_with_valid_config", func(t *testing.T) {
		config := D3AuthConfig{
			BaseURL:  "https://auth.example.com",
			Audience: "https://api.example.com",
		}

		authenticator, err := NewD3AuthJWTAuthenticator(config)

		assert.NoError(t, err)
		assert.NotNil(t, authenticator)
	})
}

func makeD3AuthTestContext(method, path string) (*echo.Echo, echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return e, c, rec
}
