package echokit

import (
	"errors"

	"github.com/half-ogre/go-kit/kit"
	"github.com/labstack/echo/v4"
)

func RequireAuthenticated() echo.MiddlewareFunc {
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
			}

			return next(c)
		}
	}
}
