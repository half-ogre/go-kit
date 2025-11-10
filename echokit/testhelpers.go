package echokit

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/labstack/echo/v4"
)

// NewTestGetRequest creates a test GET request with the given path
func NewTestGetRequest(e *echo.Echo, path string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// NewTestPostJSONRequest creates a test POST request with JSON body
func NewTestPostJSONRequest(e *echo.Echo, path string, body string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// NewTestPutJSONRequest creates a test PUT request with JSON body
func NewTestPutJSONRequest(e *echo.Echo, path string, body string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// NewTestPatchJSONRequest creates a test PATCH request with JSON body
func NewTestPatchJSONRequest(e *echo.Echo, path string, body string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPatch, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// NewTestDeleteRequest creates a test DELETE request
func NewTestDeleteRequest(e *echo.Echo, path string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}
