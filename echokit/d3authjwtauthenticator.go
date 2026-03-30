package echokit

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/half-ogre/go-kit/kit"
	"github.com/labstack/echo/v4"
)

const (
	d3AuthJWTAuthenticatorContextKey = "go-kit-echokit-d3auth-jwt-authenticated-user"
)

type D3AuthConfig struct {
	BaseURL  string
	Audience string
}

type D3AuthJWTAuthenticator struct {
	config       D3AuthConfig
	jwtValidator *validator.Validator
}

func NewD3AuthJWTAuthenticator(config D3AuthConfig) (Authenticator, error) {
	issuerURL := strings.TrimRight(config.BaseURL, "/")

	jwksURL, err := url.Parse(issuerURL)
	if err != nil {
		return nil, kit.WrapError(err, "failed to parse d3-auth base URL")
	}

	provider := jwks.NewCachingProvider(jwksURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL,
		[]string{config.Audience},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &Auth0CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		return nil, kit.WrapError(err, "failed to create d3-auth JWT validator")
	}

	return &D3AuthJWTAuthenticator{
		config:       config,
		jwtValidator: jwtValidator,
	}, nil
}

func (a *D3AuthJWTAuthenticator) AuthenticateRequest(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return nil
	}

	authHeaderParts := strings.Fields(authHeader)
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return nil
	}

	validateResult, err := a.jwtValidator.ValidateToken(c.Request().Context(), authHeaderParts[1])
	if err != nil {
		return err
	}

	validatedClaims, ok := validateResult.(*validator.ValidatedClaims)
	if !ok {
		return errors.New("failed to cast to ValidatedClaims")
	}

	customClaims, ok := validatedClaims.CustomClaims.(*Auth0CustomClaims)
	if !ok {
		return errors.New("failed to cast custom claims")
	}

	authenticatedUser := AuthenticatedUser{
		Sub:               validatedClaims.RegisteredClaims.Subject,
		Name:              customClaims.Name,
		GivenName:         customClaims.GivenName,
		FamilyName:        customClaims.FamilyName,
		MiddleName:        customClaims.MiddleName,
		Nickname:          customClaims.Nickname,
		PreferredUsername: customClaims.PreferredUsername,
		Email:             customClaims.Email,
		EmailVerified:     customClaims.EmailVerified,
		Picture:           customClaims.Picture,
		UpdatedAt:         customClaims.UpdatedAt,
		Permissions:       customClaims.Permissions,
	}

	c.Set(d3AuthJWTAuthenticatorContextKey, &authenticatedUser)

	return nil
}

func (a *D3AuthJWTAuthenticator) GetAuthenticatedUser(c echo.Context) (*AuthenticatedUser, error) {
	user, ok := c.Get(d3AuthJWTAuthenticatorContextKey).(*AuthenticatedUser)
	if !ok || user == nil {
		return nil, errors.New("no authenticated user")
	}
	return user, nil
}

func (a *D3AuthJWTAuthenticator) HandleNotAuthenticated(c echo.Context) error {
	return c.NoContent(http.StatusUnauthorized)
}

func (a *D3AuthJWTAuthenticator) IsAuthenticated(c echo.Context) (bool, error) {
	user := c.Get(d3AuthJWTAuthenticatorContextKey)
	return user != nil, nil
}
