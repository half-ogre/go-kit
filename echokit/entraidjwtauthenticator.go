package echokit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/labstack/echo/v4"
)

const (
	entraIDJWTAuthenticatorContextKey = "go-kit-echokit-entraid-jwt-authenticated-user"
)

type EntraIDJWTAuthenticator struct {
	tenantID     string
	audience     string
	jwtValidator *validator.Validator
}

type EntraIDCustomClaims struct {
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	MiddleName        string `json:"middle_name"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	Picture           string `json:"picture"`
	UpdatedAt         int64  `json:"updated_at"`
	Scp               string `json:"scp"`
}

func (c EntraIDCustomClaims) Validate(ctx context.Context) error {
	return nil
}

// NewEntraIDJWTAuthenticator creates a JWT authenticator for Microsoft Entra ID
func NewEntraIDJWTAuthenticator(tenantID, audience string) (Authenticator, error) {
	issuerURL, err := url.Parse(fmt.Sprintf("https://sts.windows.net/%s/", tenantID))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Entra ID issuer URL: %w", err)
	}

	authorityURL, err := url.Parse(fmt.Sprintf("https://login.microsoftonline.com/%s", tenantID))
	if err != nil {
		return nil, fmt.Errorf("failed to parse authority URL: %w", err)
	}

	provider := jwks.NewCachingProvider(authorityURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{audience},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &EntraIDCustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		return nil, err
	}

	return &EntraIDJWTAuthenticator{
		tenantID:     tenantID,
		audience:     audience,
		jwtValidator: jwtValidator,
	}, nil
}

func (a *EntraIDJWTAuthenticator) AuthenticateRequest(c echo.Context) error {
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
		slog.Debug("Entra ID JWT validation failed", "error", err)
		return err
	}

	validatedClaims, ok := validateResult.(*validator.ValidatedClaims)
	if !ok {
		return errors.New("failed to cast to ValidatedClaims")
	}

	customClaims, ok := validatedClaims.CustomClaims.(*EntraIDCustomClaims)
	if !ok {
		return errors.New("failed to cast custom claims")
	}

	var permissions []string
	if customClaims.Scp != "" {
		permissions = strings.Fields(customClaims.Scp)
	}

	authenticatedUser := AuthenticatedUser{
		Sub:               validatedClaims.RegisteredClaims.Subject,
		Name:              customClaims.Name,
		GivenName:         customClaims.GivenName,
		FamilyName:        customClaims.FamilyName,
		MiddleName:        customClaims.MiddleName,
		Nickname:          "",
		PreferredUsername: customClaims.PreferredUsername,
		Email:             customClaims.Email,
		EmailVerified:     false,
		Picture:           customClaims.Picture,
		UpdatedAt:         customClaims.UpdatedAt,
		Permissions:       map[string][]string{a.audience: permissions},
	}

	c.Set(entraIDJWTAuthenticatorContextKey, &authenticatedUser)

	return nil
}

func (a *EntraIDJWTAuthenticator) GetAuthenticatedUser(c echo.Context) (*AuthenticatedUser, error) {
	user, ok := c.Get(entraIDJWTAuthenticatorContextKey).(*AuthenticatedUser)
	if !ok || user == nil {
		return nil, errors.New("no authenticated user")
	}
	return user, nil
}

func (a *EntraIDJWTAuthenticator) HandleNotAuthenticated(c echo.Context) error {
	return c.NoContent(http.StatusUnauthorized)
}

func (a *EntraIDJWTAuthenticator) IsAuthenticated(c echo.Context) (bool, error) {
	user := c.Get(entraIDJWTAuthenticatorContextKey)
	return user != nil, nil
}
