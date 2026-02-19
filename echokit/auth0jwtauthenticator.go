package echokit

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/labstack/echo/v4"
)

const (
	auth0JWTAuthenticatorContextKey = "go-kit-echokit-auth0-jwt-authenticated-user"
)

type Auth0JWTAuthenticator struct {
	config       Auth0Config
	jwtValidator *validator.Validator
}

type Auth0CustomClaims struct {
	Name              string   `json:"name"`
	GivenName         string   `json:"given_name"`
	FamilyName        string   `json:"family_name"`
	MiddleName        string   `json:"middle_name"`
	Nickname          string   `json:"nickname"`
	PreferredUsername string   `json:"preferred_username"`
	Email             string   `json:"email"`
	EmailVerified     bool     `json:"email_verified"`
	Picture           string   `json:"picture"`
	UpdatedAt         int64    `json:"updated_at"`
	Permissions       []string `json:"permissions"`
}

func (c Auth0CustomClaims) Validate(ctx context.Context) error {
	return nil // Validate does nothing, but is needed to satisfy validator.CustomClaims interface
}

type Auth0JWTAuthenticatorOption func(*Auth0JWTAuthenticator)

func NewAuth0JWTAuthenticator(config Auth0Config) (Authenticator, error) {
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

func (a *Auth0JWTAuthenticator) AuthenticateRequest(c echo.Context) error {
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

	c.Set(auth0JWTAuthenticatorContextKey, &authenticatedUser)

	return nil
}

func (a *Auth0JWTAuthenticator) GetAuthenticatedUser(c echo.Context) (*AuthenticatedUser, error) {
	user, ok := c.Get(auth0JWTAuthenticatorContextKey).(*AuthenticatedUser)
	if !ok || user == nil {
		return nil, errors.New("no authenticated user")
	}
	return user, nil
}

func (a *Auth0JWTAuthenticator) HandleNotAuthenticated(c echo.Context) error {
	return c.NoContent(http.StatusUnauthorized)
}

func (a *Auth0JWTAuthenticator) IsAuthenticated(c echo.Context) (bool, error) {
	user := c.Get(auth0JWTAuthenticatorContextKey)
	return user != nil, nil
}
