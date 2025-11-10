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
	Sub               string
	Name              string
	GivenName         string
	FamilyName        string
	MiddleName        string
	Nickname          string
	PreferredUsername string
	Email             string
	EmailVerified     bool
	Picture           string
	UpdatedAt         int64
	Permissions       []string
}

type Authenticator interface {
	AuthenticateRequest(c echo.Context) error
	GetAuthenticatedUser(c echo.Context) (*AuthenticatedUser, error)
	IsAuthenticated(c echo.Context) (bool, error)
	HandleNotAuthenticated(c echo.Context) error
}

type AuthenticationMiddlewareOptions struct {
	AuthenticatedUserCallback func(AuthenticatedUser) error
}

type AuthenticationMiddlewareOption func(*AuthenticationMiddlewareOptions)

func NewAuthenticationMiddleware(authenticator Authenticator, options ...AuthenticationMiddlewareOption) echo.MiddlewareFunc {
	opts := AuthenticationMiddlewareOptions{}

	for _, option := range options {
		option(&opts)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(authenticatorContextKey, authenticator)

			err := authenticator.AuthenticateRequest(c)
			if err != nil {
				return kit.WrapError(err, "error authenticating request")
			}

			isAuthenticated, err := authenticator.IsAuthenticated(c)
			if err != nil {
				return kit.WrapError(err, "error checking authentication")
			}

			if isAuthenticated {
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
