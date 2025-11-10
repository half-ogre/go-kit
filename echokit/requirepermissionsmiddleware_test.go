package echokit

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRequirePermissions(t *testing.T) {
	t.Run("returns_an_error_when_authenticator_not_found_in_context", func(t *testing.T) {
		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		middleware := RequirePermissions([]string{"thePermission"})
		handler := middleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.EqualError(t, err, "authenticator not found in context")
		_ = rec
	})

	t.Run("returns_an_error_when_IsAuthenticated_returns_an_error", func(t *testing.T) {
		fakeAuthenticator := &FakeAuthenticator{
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return false, assert.AnError
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")
		c.Set(authenticatorContextKey, fakeAuthenticator)

		middleware := RequirePermissions([]string{"thePermission"})
		handler := middleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error checking authentication")
		_ = rec
	})

	t.Run("calls_HandleNotAuthenticated_when_user_is_not_authenticated", func(t *testing.T) {
		handleNotAuthenticatedCalled := false
		fakeAuthenticator := &FakeAuthenticator{
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return false, nil
			},
			HandleNotAuthenticatedFake: func(c echo.Context) error {
				handleNotAuthenticatedCalled = true
				return c.NoContent(http.StatusUnauthorized)
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		c.Set(authenticatorContextKey, fakeAuthenticator)

		middleware := RequirePermissions([]string{"thePermission"})
		handler := middleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.NoError(t, err)
		assert.True(t, handleNotAuthenticatedCalled)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("returns_an_error_when_GetAuthenticatedUser_returns_an_error", func(t *testing.T) {
		fakeAuthenticator := &FakeAuthenticator{
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return true, nil
			},
			GetAuthenticatedUserFake: func(c echo.Context) (*AuthenticatedUser, error) {
				return nil, assert.AnError
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")
		c.Set(authenticatorContextKey, fakeAuthenticator)

		middleware := RequirePermissions([]string{"thePermission"})
		handler := middleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error getting authenticated user")
		_ = rec
	})

	t.Run("calls_HandleNotAuthenticated_when_user_does_not_have_required_permission", func(t *testing.T) {
		handleNotAuthenticatedCalled := false
		fakeAuthenticator := &FakeAuthenticator{
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return true, nil
			},
			GetAuthenticatedUserFake: func(c echo.Context) (*AuthenticatedUser, error) {
				return &AuthenticatedUser{
					Permissions: []string{"aPermission"},
				}, nil
			},
			HandleNotAuthenticatedFake: func(c echo.Context) error {
				handleNotAuthenticatedCalled = true
				return c.NoContent(http.StatusUnauthorized)
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		c.Set(authenticatorContextKey, fakeAuthenticator)

		middleware := RequirePermissions([]string{"thePermission"})
		handler := middleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.NoError(t, err)
		assert.True(t, handleNotAuthenticatedCalled)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("calls_HandleNotAuthenticated_when_user_does_not_have_all_required_permissions", func(t *testing.T) {
		handleNotAuthenticatedCalled := false
		fakeAuthenticator := &FakeAuthenticator{
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return true, nil
			},
			GetAuthenticatedUserFake: func(c echo.Context) (*AuthenticatedUser, error) {
				return &AuthenticatedUser{
					Permissions: []string{"thePermission1"},
				}, nil
			},
			HandleNotAuthenticatedFake: func(c echo.Context) error {
				handleNotAuthenticatedCalled = true
				return c.NoContent(http.StatusUnauthorized)
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		c.Set(authenticatorContextKey, fakeAuthenticator)

		middleware := RequirePermissions([]string{"thePermission1", "thePermission2"})
		handler := middleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.NoError(t, err)
		assert.True(t, handleNotAuthenticatedCalled)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("calls_next_handler_when_user_has_required_permission", func(t *testing.T) {
		nextHandlerCalled := false
		fakeAuthenticator := &FakeAuthenticator{
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return true, nil
			},
			GetAuthenticatedUserFake: func(c echo.Context) (*AuthenticatedUser, error) {
				return &AuthenticatedUser{
					Permissions: []string{"thePermission"},
				}, nil
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		c.Set(authenticatorContextKey, fakeAuthenticator)

		middleware := RequirePermissions([]string{"thePermission"})
		handler := middleware(func(c echo.Context) error {
			nextHandlerCalled = true
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.NoError(t, err)
		assert.True(t, nextHandlerCalled)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("calls_next_handler_when_user_has_all_required_permissions", func(t *testing.T) {
		nextHandlerCalled := false
		fakeAuthenticator := &FakeAuthenticator{
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return true, nil
			},
			GetAuthenticatedUserFake: func(c echo.Context) (*AuthenticatedUser, error) {
				return &AuthenticatedUser{
					Permissions: []string{"thePermission1", "thePermission2"},
				}, nil
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		c.Set(authenticatorContextKey, fakeAuthenticator)

		middleware := RequirePermissions([]string{"thePermission1", "thePermission2"})
		handler := middleware(func(c echo.Context) error {
			nextHandlerCalled = true
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.NoError(t, err)
		assert.True(t, nextHandlerCalled)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("calls_next_handler_when_user_has_required_permissions_plus_additional_permissions", func(t *testing.T) {
		nextHandlerCalled := false
		fakeAuthenticator := &FakeAuthenticator{
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return true, nil
			},
			GetAuthenticatedUserFake: func(c echo.Context) (*AuthenticatedUser, error) {
				return &AuthenticatedUser{
					Permissions: []string{"thePermission1", "thePermission2", "aPermission3"},
				}, nil
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		c.Set(authenticatorContextKey, fakeAuthenticator)

		middleware := RequirePermissions([]string{"thePermission1", "thePermission2"})
		handler := middleware(func(c echo.Context) error {
			nextHandlerCalled = true
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.NoError(t, err)
		assert.True(t, nextHandlerCalled)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("calls_HandleNotAuthenticated_when_user_does_not_have_any_or_permissions", func(t *testing.T) {
		handleNotAuthenticatedCalled := false
		fakeAuthenticator := &FakeAuthenticator{
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return true, nil
			},
			GetAuthenticatedUserFake: func(c echo.Context) (*AuthenticatedUser, error) {
				return &AuthenticatedUser{
					Permissions: []string{"aPermission"},
				}, nil
			},
			HandleNotAuthenticatedFake: func(c echo.Context) error {
				handleNotAuthenticatedCalled = true
				return c.NoContent(http.StatusUnauthorized)
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		c.Set(authenticatorContextKey, fakeAuthenticator)

		middleware := RequirePermissions([]string{"thePermission1"}, []string{"thePermission2"}, []string{"thePermission3"})
		handler := middleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.NoError(t, err)
		assert.True(t, handleNotAuthenticatedCalled)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("calls_next_handler_when_user_has_first_or_permission", func(t *testing.T) {
		nextHandlerCalled := false
		fakeAuthenticator := &FakeAuthenticator{
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return true, nil
			},
			GetAuthenticatedUserFake: func(c echo.Context) (*AuthenticatedUser, error) {
				return &AuthenticatedUser{
					Permissions: []string{"thePermission2"},
				}, nil
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		c.Set(authenticatorContextKey, fakeAuthenticator)

		middleware := RequirePermissions([]string{"thePermission1"}, []string{"thePermission2"}, []string{"thePermission3"})
		handler := middleware(func(c echo.Context) error {
			nextHandlerCalled = true
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.NoError(t, err)
		assert.True(t, nextHandlerCalled)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("calls_next_handler_when_user_has_second_or_permission", func(t *testing.T) {
		nextHandlerCalled := false
		fakeAuthenticator := &FakeAuthenticator{
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return true, nil
			},
			GetAuthenticatedUserFake: func(c echo.Context) (*AuthenticatedUser, error) {
				return &AuthenticatedUser{
					Permissions: []string{"thePermission3"},
				}, nil
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		c.Set(authenticatorContextKey, fakeAuthenticator)

		middleware := RequirePermissions([]string{"thePermission1"}, []string{"thePermission2"}, []string{"thePermission3"})
		handler := middleware(func(c echo.Context) error {
			nextHandlerCalled = true
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.NoError(t, err)
		assert.True(t, nextHandlerCalled)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("calls_next_handler_when_user_has_all_permissions_in_or_permission_set", func(t *testing.T) {
		nextHandlerCalled := false
		fakeAuthenticator := &FakeAuthenticator{
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return true, nil
			},
			GetAuthenticatedUserFake: func(c echo.Context) (*AuthenticatedUser, error) {
				return &AuthenticatedUser{
					Permissions: []string{"thePermission2", "thePermission3"},
				}, nil
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		c.Set(authenticatorContextKey, fakeAuthenticator)

		middleware := RequirePermissions(
			[]string{"thePermission1"},
			[]string{"thePermission2", "thePermission3"},
			[]string{"thePermission4"},
		)
		handler := middleware(func(c echo.Context) error {
			nextHandlerCalled = true
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.NoError(t, err)
		assert.True(t, nextHandlerCalled)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("calls_HandleNotAuthenticated_when_user_has_only_some_permissions_in_or_permission_set", func(t *testing.T) {
		handleNotAuthenticatedCalled := false
		fakeAuthenticator := &FakeAuthenticator{
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return true, nil
			},
			GetAuthenticatedUserFake: func(c echo.Context) (*AuthenticatedUser, error) {
				return &AuthenticatedUser{
					Permissions: []string{"thePermission2"},
				}, nil
			},
			HandleNotAuthenticatedFake: func(c echo.Context) error {
				handleNotAuthenticatedCalled = true
				return c.NoContent(http.StatusUnauthorized)
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		c.Set(authenticatorContextKey, fakeAuthenticator)

		middleware := RequirePermissions(
			[]string{"thePermission1"},
			[]string{"thePermission2", "thePermission3"},
			[]string{"thePermission4"},
		)
		handler := middleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.NoError(t, err)
		assert.True(t, handleNotAuthenticatedCalled)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestRequirePermission(t *testing.T) {
	t.Run("calls_next_handler_when_user_has_required_permission", func(t *testing.T) {
		nextHandlerCalled := false
		fakeAuthenticator := &FakeAuthenticator{
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return true, nil
			},
			GetAuthenticatedUserFake: func(c echo.Context) (*AuthenticatedUser, error) {
				return &AuthenticatedUser{
					Permissions: []string{"thePermission"},
				}, nil
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		c.Set(authenticatorContextKey, fakeAuthenticator)

		middleware := RequirePermission("thePermission")
		handler := middleware(func(c echo.Context) error {
			nextHandlerCalled = true
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.NoError(t, err)
		assert.True(t, nextHandlerCalled)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("calls_next_handler_when_user_has_first_or_permission", func(t *testing.T) {
		nextHandlerCalled := false
		fakeAuthenticator := &FakeAuthenticator{
			IsAuthenticatedFake: func(c echo.Context) (bool, error) {
				return true, nil
			},
			GetAuthenticatedUserFake: func(c echo.Context) (*AuthenticatedUser, error) {
				return &AuthenticatedUser{
					Permissions: []string{"thePermission2"},
				}, nil
			},
		}

		e := echo.New()
		c, rec := NewTestGetRequest(e, "/")

		c.Set(authenticatorContextKey, fakeAuthenticator)

		middleware := RequirePermission("thePermission1", "thePermission2", "thePermission3")
		handler := middleware(func(c echo.Context) error {
			nextHandlerCalled = true
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)

		assert.NoError(t, err)
		assert.True(t, nextHandlerCalled)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
