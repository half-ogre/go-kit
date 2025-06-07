package echokit

import (
	"errors"

	"github.com/half-ogre/go-kit/kit"
	"github.com/labstack/echo/v4"
)

const (
	authenticatorContextKey = "github.com/half-ogre/go-kit/echokit/authenticator"
)

type AuthenticatedUser struct {
	Nickname  string
	AvatarUrl string
	Sub       string
}

type Authenticator interface {
	GetAuthenticatedUser(c echo.Context) (*AuthenticatedUser, error)
	IsAuthenticated(c echo.Context) (bool, error)
	HandleNotAuthenticated(c echo.Context) error
}

func NewAuthenticationMiddleware(authenticator Authenticator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(authenticatorContextKey, authenticator)
			return next(c)
		}
	}
}

func GetAuthenticator(c echo.Context) (Authenticator, error) {
	o := c.Get(authenticatorContextKey)
	if o == nil {
		return nil, nil
	}

	authenticator, ok := o.(Authenticator)
	if !ok {
		return nil, errors.New("failed to cast authenticator from context")
	}

	return authenticator, nil
}

type AuthorizationMiddlewareOptions struct {
	AuthenticatedUserCallback func(AuthenticatedUser) error
}

type AuthorizationMiddlewareOption func(*AuthorizationMiddlewareOptions)

func NewAuthorizationMiddleware(authenticator Authenticator, options ...AuthorizationMiddlewareOption) echo.MiddlewareFunc {
	opts := AuthorizationMiddlewareOptions{}

	for _, option := range options {
		option(&opts)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
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

				if opts.AuthenticatedUserCallback != nil {
					err = opts.AuthenticatedUserCallback(*authenticatedUser)
					if err != nil {
						return kit.WrapError(err, "error calling authenticated user callback")
					}
				}
			}

			return next(c)
		}
	}
}

func WithAuthenticatedUserCallback(callback func(AuthenticatedUser) error) func(*AuthorizationMiddlewareOptions) {
	return func(options *AuthorizationMiddlewareOptions) {
		options.AuthenticatedUserCallback = callback
	}
}
