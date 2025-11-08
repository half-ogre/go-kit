package echokit

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/half-ogre/go-kit/kit"
	"github.com/labstack/echo/v4"
)

type Auth0JWTAuthenticator struct {
	config       Auth0Config
	jwtValidator *validator.Validator
}

type Auth0CustomClaims struct {
	Permissions []string `json:"permissions"`
	Picture     string   `json:"picture"`
	Nickname    string   `json:"nickname"`
	Scope       string   `json:"scope"`
}

func (c Auth0CustomClaims) Validate(ctx context.Context) error {
	return nil // Validate does nothing, but is needed to satisfy validator.CustomClaims interface
}

type Auth0JWTAuthenticatorOption func(*Auth0JWTAuthenticator)

func NewAuth0JWTAuthenticator(config Auth0Config) (*Auth0JWTAuthenticator, error) {
	jwtAuthenticator := &Auth0JWTAuthenticator{
		config: config,
	}

	issuerURL, err := url.Parse("https://" + config.Domain + "/")
	if err != nil {
		log.Fatalf("Failed to parse the issuer url: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{config.Audience},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &Auth0CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		return nil, err
	}

	jwtAuthenticator.jwtValidator = jwtValidator

	return jwtAuthenticator, nil
}

func (a *Auth0JWTAuthenticator) GetAuthenticatedUser(c echo.Context) (*AuthenticatedUser, error) {
	ok, err := a.IsAuthenticated(c)
	if err != nil {
		return nil, kit.WrapError(err, "failed to check authentication")
	}

	if !ok {
		return nil, errors.New("no authenticated user")
	}

	session, err := GetSession("fx-jwt-authenticator", c)
	if err != nil {
		return nil, kit.WrapError(err, "error getting auth session")
	}

	if session == nil {
		return nil, errors.New("failed to get auth session")
	}

	authenticatedUserBytes, ok := session.Values["authenticated-user"].([]byte)
	if !ok {
		return nil, errors.New("failed to get authenticated user from session")
	}

	slog.Debug("Auth0JWTAuthenticator#GetAuthenticatedUser:has-authenticated-user", "authenticatedUserBytes", string(authenticatedUserBytes))

	authenticatedUser := AuthenticatedUser{}
	err = json.Unmarshal(authenticatedUserBytes, &authenticatedUser)
	if err != nil {
		return nil, kit.WrapError(err, "failed to unmarshal authenticated user bytes")
	}

	return &authenticatedUser, nil

}

func (a *Auth0JWTAuthenticator) HandleNotAuthenticated(c echo.Context) error {
	return c.NoContent(http.StatusUnauthorized)
}

func (a *Auth0JWTAuthenticator) IsAuthenticated(c echo.Context) (bool, error) {
	session, err := GetSession("fx-jwt-authenticator", c)
	if err != nil {
		return false, kit.WrapError(err, "error getting auth session")
	}

	if session == nil {
		return false, errors.New("failed to get auth session")
	}

	_, ok := session.Values["authenticated-user"]
	if ok {
		return true, nil
	}

	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return false, nil
	}

	authHeaderParts := strings.Fields(authHeader)
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return false, nil
	}

	validateResult, err := a.jwtValidator.ValidateToken(c.Request().Context(), authHeaderParts[1])
	if err != nil {
		return false, err
	}

	validatedClaims, ok := validateResult.(*validator.ValidatedClaims)
	if !ok {
		return false, errors.New("failed to cast to ValidatedClaims")
	}

	customClaims, ok := validatedClaims.CustomClaims.(*Auth0CustomClaims)
	if !ok {
		return false, errors.New("failed to cast custom claims")
	}

	authenticatedUser := AuthenticatedUser{
		Sub:       validatedClaims.RegisteredClaims.Subject,
		Nickname:  customClaims.Nickname,
		AvatarUrl: customClaims.Picture,
	}

	authenticatedUserBytes, err := json.Marshal(authenticatedUser)
	if err != nil {
		return false, kit.WrapError(err, "failed to marshal authenticated user")
	}

	session.Values["authenticated-user"] = authenticatedUserBytes

	err = session.Save(c.Request(), c.Response().Writer)
	if err != nil {
		return false, kit.WrapError(err, "failed to save claims to session")
	}

	return true, nil
}
