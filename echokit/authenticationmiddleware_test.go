package echokit

import (
	"errors"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthenticationMiddleware(t *testing.T) {
	t.Run("sets_authenticator_in_context", func(t *testing.T) {
		fakeAuthenticator := &FakeAuthenticator{
			AuthenticateRequestFake: func(c echo.Context) error {
				return nil
			},
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return false, nil
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		middleware := NewAuthenticationMiddleware(fakeAuthenticator)
		handler := middleware(func(c echo.Context) error {
			authenticator, err := GetAuthenticator(c)
			assert.NoError(t, err)
			assert.Equal(t, fakeAuthenticator, authenticator)
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("returns_an_error_when_AuthenticateRequest_returns_an_error", func(t *testing.T) {
		fakeAuthenticator := &FakeAuthenticator{
			AuthenticateRequestFake: func(c echo.Context) error {
				return errors.New("the fake error")
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		middleware := NewAuthenticationMiddleware(fakeAuthenticator)
		handler := middleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error authenticating request")
		assert.Contains(t, err.Error(), "the fake error")
		_ = rec
	})

	t.Run("returns_an_error_when_IsAuthenticated_returns_an_error", func(t *testing.T) {
		fakeAuthenticator := &FakeAuthenticator{
			AuthenticateRequestFake: func(c echo.Context) error {
				return nil
			},
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return false, errors.New("the fake error")
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		middleware := NewAuthenticationMiddleware(fakeAuthenticator)
		handler := middleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error checking authentication")
		assert.Contains(t, err.Error(), "the fake error")
		_ = rec
	})

	t.Run("calls_next_handler_when_user_is_not_authenticated", func(t *testing.T) {
		nextHandlerCalled := false
		fakeAuthenticator := &FakeAuthenticator{
			AuthenticateRequestFake: func(c echo.Context) error {
				return nil
			},
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return false, nil
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		middleware := NewAuthenticationMiddleware(fakeAuthenticator)
		handler := middleware(func(c echo.Context) error {
			nextHandlerCalled = true
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.NoError(t, err)
		assert.True(t, nextHandlerCalled)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("returns_an_error_when_GetAuthenticatedUser_returns_an_error", func(t *testing.T) {
		fakeAuthenticator := &FakeAuthenticator{
			AuthenticateRequestFake: func(c echo.Context) error {
				return nil
			},
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return true, nil
			},
			GetAuthenticatedUserFake: func(c echo.Context) (*AuthenticatedUser, error) {
				return nil, errors.New("the fake error")
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		middleware := NewAuthenticationMiddleware(fakeAuthenticator)
		handler := middleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error getting authenticated user")
		assert.Contains(t, err.Error(), "the fake error")
		_ = rec
	})

	t.Run("calls_next_handler_when_user_is_authenticated", func(t *testing.T) {
		nextHandlerCalled := false
		fakeAuthenticator := &FakeAuthenticator{
			AuthenticateRequestFake: func(c echo.Context) error {
				return nil
			},
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return true, nil
			},
			GetAuthenticatedUserFake: func(c echo.Context) (*AuthenticatedUser, error) {
				return &AuthenticatedUser{
					Sub:      "theSub",
					Nickname: "theNickname",
				}, nil
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		middleware := NewAuthenticationMiddleware(fakeAuthenticator)
		handler := middleware(func(c echo.Context) error {
			nextHandlerCalled = true
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.NoError(t, err)
		assert.True(t, nextHandlerCalled)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("calls_authenticated_user_callback_when_user_is_authenticated", func(t *testing.T) {
		callbackCalled := false
		var actualUser AuthenticatedUser
		fakeAuthenticator := &FakeAuthenticator{
			AuthenticateRequestFake: func(c echo.Context) error {
				return nil
			},
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return true, nil
			},
			GetAuthenticatedUserFake: func(c echo.Context) (*AuthenticatedUser, error) {
				return &AuthenticatedUser{
					Sub:      "theSub",
					Nickname: "theNickname",
				}, nil
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		middleware := NewAuthenticationMiddleware(fakeAuthenticator, func(opts *AuthenticationMiddlewareOptions) {
			opts.AuthenticatedUserCallback = func(user AuthenticatedUser) error {
				callbackCalled = true
				actualUser = user
				return nil
			}
		})
		handler := middleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.NoError(t, err)
		assert.True(t, callbackCalled)
		assert.Equal(t, "theSub", actualUser.Sub)
		assert.Equal(t, "theNickname", actualUser.Nickname)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("returns_an_error_when_authenticated_user_callback_returns_an_error", func(t *testing.T) {
		fakeAuthenticator := &FakeAuthenticator{
			AuthenticateRequestFake: func(c echo.Context) error {
				return nil
			},
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return true, nil
			},
			GetAuthenticatedUserFake: func(c echo.Context) (*AuthenticatedUser, error) {
				return &AuthenticatedUser{
					Sub:      "theSub",
					Nickname: "theNickname",
				}, nil
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		middleware := NewAuthenticationMiddleware(fakeAuthenticator, func(opts *AuthenticationMiddlewareOptions) {
			opts.AuthenticatedUserCallback = func(user AuthenticatedUser) error {
				return errors.New("the callback error")
			}
		})
		handler := middleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error calling authenticated user callback")
		assert.Contains(t, err.Error(), "the callback error")
		_ = rec
	})

	t.Run("does_not_call_authenticated_user_callback_when_user_is_not_authenticated", func(t *testing.T) {
		callbackCalled := false
		fakeAuthenticator := &FakeAuthenticator{
			AuthenticateRequestFake: func(c echo.Context) error {
				return nil
			},
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return false, nil
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		middleware := NewAuthenticationMiddleware(fakeAuthenticator, func(opts *AuthenticationMiddlewareOptions) {
			opts.AuthenticatedUserCallback = func(user AuthenticatedUser) error {
				callbackCalled = true
				return nil
			}
		})
		handler := middleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.NoError(t, err)
		assert.False(t, callbackCalled)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestGetAuthenticator(t *testing.T) {
	t.Run("returns_authenticator_when_found_in_context", func(t *testing.T) {
		fakeAuthenticator := &FakeAuthenticator{}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")
		c.Set(authenticatorContextKey, fakeAuthenticator)

		authenticator, err := GetAuthenticator(c)

		assert.NoError(t, err)
		assert.Equal(t, fakeAuthenticator, authenticator)
		_ = rec
	})

	t.Run("returns_nil_when_authenticator_not_found_in_context", func(t *testing.T) {
		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		authenticator, err := GetAuthenticator(c)

		assert.NoError(t, err)
		assert.Nil(t, authenticator)
		_ = rec
	})

	t.Run("returns_an_error_when_context_value_cannot_be_cast_to_authenticator", func(t *testing.T) {
		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")
		c.Set(authenticatorContextKey, "not an authenticator")

		authenticator, err := GetAuthenticator(c)

		assert.EqualError(t, err, "failed to cast authenticator from context")
		assert.Nil(t, authenticator)
		_ = rec
	})
}
