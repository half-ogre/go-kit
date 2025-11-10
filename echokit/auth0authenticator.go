package echokit

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/half-ogre/go-kit/kit"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

const (
	auth0AuthenticatorSessionKey = "go-kit-echokit-auth0-authenticator"
)

type Auth0Config struct {
	Audience     string
	CallbackPath string
	ClientId     string
	ClientSecret string
	Domain       string
}

type Auth0Authenticator struct {
	config       Auth0Config
	oauthConfig  *oauth2.Config
	oidcProvider *oidc.Provider
}

type Auth0AuthenticatorOption func(*Auth0Authenticator)

func NewAuth0Authenticator(config Auth0Config) (Authenticator, error) {
	oidcProvider, err := oidc.NewProvider(context.Background(), fmt.Sprintf("https://%s/", config.Domain))
	if err != nil {
		return nil, err
	}

	// RedirectURL is intentionally not set because it is built dynamically based on request host
	oauthConfig := oauth2.Config{
		ClientID:     config.ClientId,
		ClientSecret: config.ClientSecret,
		Endpoint:     oidcProvider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	auth0Authenticator := &Auth0Authenticator{
		config:       config,
		oauthConfig:  &oauthConfig,
		oidcProvider: oidcProvider,
	}

	return auth0Authenticator, nil
}

func (a *Auth0Authenticator) AuthenticateRequest(c echo.Context) error {
	// Unlike JWT authentication, the OAuth authentication flow handles the actual authentication in the callback, so there is nothing to do here
	return nil
}

func (a *Auth0Authenticator) GetAuthenticatedUser(c echo.Context) (*AuthenticatedUser, error) {
	if ok, err := a.IsAuthenticated(c); !ok {
		return nil, err
	} else {
		session, err := GetSession(auth0AuthenticatorSessionKey, c)
		if err != nil {
			return nil, kit.WrapError(err, "error getting auth session")
		}

		if session == nil {
			return nil, errors.New("failed to get auth session")
		}

		claims, ok := session.Values["claims"]
		if !ok {
			return nil, errors.New("failed to get claims from session")
		}

		var claimsMap map[string]interface{}
		err = json.Unmarshal([]byte(claims.(string)), &claimsMap)
		if err != nil {
			return nil, kit.WrapError(err, "failed to unmarshal claims")
		}

		slog.Debug("claims", claims)

		var permissions []string
		if permissionsRaw, ok := claimsMap["permissions"]; ok {
			if permissionsArray, ok := permissionsRaw.([]interface{}); ok {
				for _, p := range permissionsArray {
					if pStr, ok := p.(string); ok {
						permissions = append(permissions, pStr)
					}
				}
			}
		}

		var name, givenName, familyName, middleName, nickname, preferredUsername, email, picture string
		var emailVerified bool
		var updatedAt int64
		if v, ok := claimsMap["name"].(string); ok {
			name = v
		}
		if v, ok := claimsMap["given_name"].(string); ok {
			givenName = v
		}
		if v, ok := claimsMap["family_name"].(string); ok {
			familyName = v
		}
		if v, ok := claimsMap["middle_name"].(string); ok {
			middleName = v
		}
		if v, ok := claimsMap["nickname"].(string); ok {
			nickname = v
		}
		if v, ok := claimsMap["preferred_username"].(string); ok {
			preferredUsername = v
		}
		if v, ok := claimsMap["email"].(string); ok {
			email = v
		}
		if v, ok := claimsMap["email_verified"].(bool); ok {
			emailVerified = v
		}
		if v, ok := claimsMap["picture"].(string); ok {
			picture = v
		}
		if v, ok := claimsMap["updated_at"].(float64); ok {
			updatedAt = int64(v)
		}

		return &AuthenticatedUser{
			Sub:               claimsMap["sub"].(string),
			Name:              name,
			GivenName:         givenName,
			FamilyName:        familyName,
			MiddleName:        middleName,
			Nickname:          nickname,
			PreferredUsername: preferredUsername,
			Email:             email,
			EmailVerified:     emailVerified,
			Picture:           picture,
			UpdatedAt:         updatedAt,
			Permissions:       permissions,
		}, nil
	}
}

func (a *Auth0Authenticator) HandleNotAuthenticated(c echo.Context) error {
	authURL, err := a.GetAuthCodeURL(c)
	if err != nil {
		return kit.WrapError(err, "error getting authentication URL")
	}
	return c.Redirect(http.StatusTemporaryRedirect, authURL.String())
}

func (a *Auth0Authenticator) IsAuthenticated(c echo.Context) (bool, error) {
	session, err := GetSession(auth0AuthenticatorSessionKey, c)
	if err != nil {
		return false, kit.WrapError(err, "error getting auth session")
	}

	if session == nil {
		return false, errors.New("failed to get auth session")
	}

	_, ok := session.Values["access_token"]
	if !ok {
		return false, nil
	}

	return true, nil
}

func (a *Auth0Authenticator) HandleAuthenticationCallback(c echo.Context) (bool, error) {
	session, err := GetSession(auth0AuthenticatorSessionKey, c)
	if err != nil {
		return false, kit.WrapError(err, "failed to get auth session")
	}

	if c.QueryParam("state") != session.Values["state"] {
		return false, fmt.Errorf("query state %s did not match session state %s", c.QueryParam("state"), session.Values["state"])
	}

	callbackOption, err := buildCallbackAuthCodeOption(c, "")
	if err != nil {
		return false, kit.WrapError(err, "failed to build callback auth code option")
	}

	token, err := a.oauthConfig.Exchange(c.Request().Context(), c.QueryParam("code"), callbackOption)
	if err != nil {
		return false, kit.WrapError(err, "failed to exchange token")
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return false, errors.New("no id_token field in oauth2 token")
	}

	verifier := a.oidcProvider.Verifier(&oidc.Config{ClientID: a.oauthConfig.ClientID})
	idToken, err := verifier.Verify(c.Request().Context(), rawIDToken)
	if err != nil {
		return false, kit.WrapError(err, "failed to verify ID token")
	}

	var claimsJSON map[string]interface{}
	if err := idToken.Claims(&claimsJSON); err != nil {
		return false, kit.WrapError(err, "failed to read claims from ID token")
	}

	claimsBytes, err := json.Marshal(claimsJSON)
	if err != nil {
		return false, kit.WrapError(err, "failed to marshal claims")
	}

	session.Values["access_token"] = token.AccessToken
	session.Values["refresh_token"] = token.RefreshToken
	session.Values["expiry"] = token.Expiry.UTC().Format(time.RFC3339)
	session.Values["token_type"] = token.TokenType
	session.Values["claims"] = string(claimsBytes)

	err = session.Save(c.Request(), c.Response().Writer)
	if err != nil {
		return false, kit.WrapError(err, "failed to save user to session")
	}

	return true, nil
}

func (a *Auth0Authenticator) GetAuthCodeURL(c echo.Context) (*url.URL, error) {
	session, err := GetSession(auth0AuthenticatorSessionKey, c)
	if err != nil {
		return nil, kit.WrapError(err, "error getting auth session")
	}

	if session == nil {
		return nil, errors.New("failed to get auth session")
	}

	state, err := generateRandomState()
	if err != nil {
		return nil, kit.WrapError(err, "error generating state")
	}

	session.Values["state"] = state
	err = session.Save(c.Request(), c.Response().Writer)
	if err != nil {
		return nil, kit.WrapError(err, "failed to save state to session")
	}

	callbackOption, err := buildCallbackAuthCodeOption(c, "/auth/callback")
	if err != nil {
		return nil, kit.WrapError(err, "failed to build callback auth code option")
	}

	authCodeUrl, err := url.Parse(a.oauthConfig.AuthCodeURL(state, callbackOption))
	if err != nil {
		return nil, kit.WrapError(err, "failed to parse auth code URL")
	}

	return authCodeUrl, nil
}

func (a *Auth0Authenticator) Login(c echo.Context) error {
	authCodeURL, err := a.GetAuthCodeURL(c)
	if err != nil {
		return kit.WrapError(err, "failed to get auth code URL")
	}

	return c.Redirect(http.StatusTemporaryRedirect, authCodeURL.String())
}

func (a *Auth0Authenticator) Logout(c echo.Context) error {
	logoutUrl, err := url.Parse(fmt.Sprintf("https://%s/v2/logout", a.config.Domain))
	if err != nil {
		return kit.WrapError(err, "failed to parse logout URL")
	}

	returnTo, err := url.Parse("https://" + c.Request().Host)
	if err != nil {
		return kit.WrapError(err, "failed to parse return URL")
	}

	parameters := url.Values{}
	parameters.Add("returnTo", returnTo.String())
	parameters.Add("client_id", a.config.ClientId)
	logoutUrl.RawQuery = parameters.Encode()

	err = DeleteSession(auth0AuthenticatorSessionKey, c)
	if err != nil {
		return kit.WrapError(err, "failed to delete session")
	}

	return c.Redirect(http.StatusTemporaryRedirect, logoutUrl.String())
}

func buildCallbackAuthCodeOption(c echo.Context, path string) (oauth2.AuthCodeOption, error) {
	callbackUrl, err := url.Parse("https://" + c.Request().Host)
	if err != nil {
		return nil, kit.WrapError(err, "failed to parse host %s", c.Request().Host)
	}

	callbackUrl.Path = path
	return oauth2.SetAuthURLParam("redirect_uri", callbackUrl.String()), nil
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	state := base64.StdEncoding.EncodeToString(b)

	return state, nil
}
