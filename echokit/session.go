package echokit

import (
	"fmt"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/half-ogre/go-kit/kit"
	"github.com/labstack/echo/v4"
)

const CONTEXT_KEY_SESSION_STORE = "fx-session-store"

func DeleteSession(name string, c echo.Context) error {
	v := c.Get(CONTEXT_KEY_SESSION_STORE)

	if v == nil {
		return fmt.Errorf("failed to get session store %s fron context", "2")
	}

	sessionStore, ok := v.(sessions.Store)
	if !ok {
		return fmt.Errorf("failed to cast %+v to session store", v)
	}

	s, err := sessionStore.Get(c.Request(), name)
	if err != nil {
		return kit.WrapError(err, "error getting session")
	}

	s.Values = make(map[interface{}]interface{})
	s.Options.MaxAge = -1

	err = s.Save(c.Request(), c.Response().Writer)
	if err != nil {
		return kit.WrapError(err, "failed to delete session")
	}

	return nil
}

func GetSession(name string, c echo.Context) (*sessions.Session, error) {
	v := c.Get(CONTEXT_KEY_SESSION_STORE)

	if v == nil {
		return nil, fmt.Errorf("failed to get session store %s fron context", "2")
	}

	sessionStore, ok := v.(sessions.Store)
	if !ok {
		return nil, fmt.Errorf("failed to cast %+v to session store", v)
	}

	s, err := sessionStore.Get(c.Request(), name)
	if err != nil {
		return nil, kit.WrapError(err, "error getting session")
	}

	return s, nil
}

func NewSessionMiddleware(sessionStore sessions.Store) echo.MiddlewareFunc {
	if sessionStore == nil {
		panic("session store must not be nil")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer context.Clear(c.Request())

			c.Set(CONTEXT_KEY_SESSION_STORE, sessionStore)

			return next(c)
		}
	}
}
