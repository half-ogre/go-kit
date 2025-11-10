package echokit

import (
	"errors"
	"log/slog"
	"slices"

	"github.com/half-ogre/go-kit/kit"
	"github.com/labstack/echo/v4"
)

func RequirePermissions(permissions []string, orPermissions ...[]string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authenticator, err := GetAuthenticator(c)
			if err != nil {
				return kit.WrapError(err, "error getting authenticator")
			}

			if authenticator == nil {
				return errors.New("authenticator not found in context")
			}

			isAuthenticated, err := authenticator.IsAuthenticated(c)
			if err != nil {
				return kit.WrapError(err, "error checking authentication")
			}

			if !isAuthenticated {
				return authenticator.HandleNotAuthenticated(c)
			} else {
				authenticatedUser, err := authenticator.GetAuthenticatedUser(c)
				if err != nil {
					return kit.WrapError(err, "error getting authenticated user")
				}

				slog.Debug("checking user permissions", "user", authenticatedUser)

				hasPermissions := checkPermissions(authenticatedUser.Permissions, permissions)
				if !hasPermissions {
					for _, orPerms := range orPermissions {
						if checkPermissions(authenticatedUser.Permissions, orPerms) {
							hasPermissions = true
							break
						}
					}
				}

				if !hasPermissions {
					return authenticator.HandleNotAuthenticated(c)
				}
			}

			return next(c)
		}
	}
}

func RequirePermission(permission string, orPermission ...string) echo.MiddlewareFunc {
	orPermissions := [][]string{}
	for _, orP := range orPermission {
		orPermissions = append(orPermissions, []string{orP})
	}

	return RequirePermissions([]string{permission}, orPermissions...)
}

func checkPermissions(userPermissions []string, requiredPermissions []string) bool {
	for _, required := range requiredPermissions {
		found := slices.Contains(userPermissions, required)
		if !found {
			return false
		}
	}
	return true
}
