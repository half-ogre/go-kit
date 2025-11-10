package echokit

import "github.com/labstack/echo/v4"

type FakeAuthenticator struct {
	AuthenticateRequestFake    func(c echo.Context) error
	GetAuthenticatedUserFake   func(c echo.Context) (*AuthenticatedUser, error)
	IsAuthenticatedFake        func(c echo.Context) (bool, error)
	HandleNotAuthenticatedFake func(c echo.Context) error
}

func (f *FakeAuthenticator) AuthenticateRequest(c echo.Context) error {
	if f.AuthenticateRequestFake != nil {
		return f.AuthenticateRequestFake(c)
	}
	panic("AuthenticateRequest fake not implemented")
}

func (f *FakeAuthenticator) GetAuthenticatedUser(c echo.Context) (*AuthenticatedUser, error) {
	if f.GetAuthenticatedUserFake != nil {
		return f.GetAuthenticatedUserFake(c)
	}
	panic("GetAuthenticatedUser fake not implemented")
}

func (f *FakeAuthenticator) IsAuthenticated(c echo.Context) (bool, error) {
	if f.IsAuthenticatedFake != nil {
		return f.IsAuthenticatedFake(c)
	}
	panic("IsAuthenticated fake not implemented")
}

func (f *FakeAuthenticator) HandleNotAuthenticated(c echo.Context) error {
	if f.HandleNotAuthenticatedFake != nil {
		return f.HandleNotAuthenticatedFake(c)
	}
	panic("HandleNotAuthenticated fake not implemented")
}
