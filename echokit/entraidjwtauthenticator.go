package echokit

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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

const (
	entraIDJWTAuthenticatorSessionKey = "go-kit-echokit-entraid-jwt-authenticator"
)

type EntraIDJWTAuthenticator struct {
	tenantID     string
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
	// Entra ID v1.0 issuer URL: https://sts.windows.net/{tenantId}/
	issuerURL, err := url.Parse(fmt.Sprintf("https://sts.windows.net/%s/", tenantID))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Entra ID issuer URL: %w", err)
	}

	// For JWKS discovery, use login.microsoftonline.com (where OpenID config is hosted)
	// The provider will fetch /.well-known/openid-configuration from this URL
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
		jwtValidator: jwtValidator,
	}, nil
}

func (a *EntraIDJWTAuthenticator) AuthenticateRequest(c echo.Context) error {
	session, err := GetSession(entraIDJWTAuthenticatorSessionKey, c)
	if err != nil {
		return kit.WrapError(err, "error getting auth session")
	}

	if session == nil {
		return errors.New("failed to get auth session")
	}

	_, ok := session.Values["authenticated-user"]
	if ok {
		return nil
	}

	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return nil
	}

	authHeaderParts := strings.Fields(authHeader)
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return nil
	}

	// Decode token to see claims for debugging
	tokenString := authHeaderParts[1]
	parts := strings.Split(tokenString, ".")
	if len(parts) == 3 {
		// Decode claims (base64url encoded)
		claims, decodeErr := base64.RawURLEncoding.DecodeString(parts[1])
		if decodeErr == nil {
			slog.Info("JWT claims", "claims", string(claims))
		}
	}

	validateResult, err := a.jwtValidator.ValidateToken(c.Request().Context(), authHeaderParts[1])
	if err != nil {
		slog.Error("JWT validation failed", "error", err)
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
		EmailVerified:     false, // Entra ID doesn't provide email_verified in standard claims
		Picture:           customClaims.Picture,
		UpdatedAt:         customClaims.UpdatedAt,
		Permissions:       permissions,
	}

	authenticatedUserBytes, err := json.Marshal(authenticatedUser)
	if err != nil {
		return kit.WrapError(err, "failed to marshal authenticated user")
	}

	session.Values["authenticated-user"] = authenticatedUserBytes

	err = session.Save(c.Request(), c.Response().Writer)
	if err != nil {
		return kit.WrapError(err, "failed to save claims to session")
	}

	return nil
}

func (a *EntraIDJWTAuthenticator) GetAuthenticatedUser(c echo.Context) (*AuthenticatedUser, error) {
	ok, err := a.IsAuthenticated(c)
	if err != nil {
		return nil, kit.WrapError(err, "failed to check authentication")
	}

	if !ok {
		return nil, errors.New("no authenticated user")
	}

	session, err := GetSession(entraIDJWTAuthenticatorSessionKey, c)
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

	slog.Debug("EntraIDJWTAuthenticator#GetAuthenticatedUser:has-authenticated-user", "authenticatedUserBytes", string(authenticatedUserBytes))

	authenticatedUser := AuthenticatedUser{}
	err = json.Unmarshal(authenticatedUserBytes, &authenticatedUser)
	if err != nil {
		return nil, kit.WrapError(err, "failed to unmarshal authenticated user bytes")
	}

	return &authenticatedUser, nil
}

func (a *EntraIDJWTAuthenticator) HandleNotAuthenticated(c echo.Context) error {
	return c.NoContent(http.StatusUnauthorized)
}

func (a *EntraIDJWTAuthenticator) IsAuthenticated(c echo.Context) (bool, error) {
	session, err := GetSession(entraIDJWTAuthenticatorSessionKey, c)
	if err != nil {
		return false, kit.WrapError(err, "error getting auth session")
	}

	if session == nil {
		return false, errors.New("failed to get auth session")
	}

	_, ok := session.Values["authenticated-user"]
	return ok, nil
}
