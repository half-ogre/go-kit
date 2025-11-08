package echokit

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
)

func TestErrorHandler(t *testing.T) {
	t.Run("wraps_non_http_error_and_sends_500_response_with_request_id", func(t *testing.T) {
		e := echo.New()
		e.HTTPErrorHandler = ErrorHandler(e)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(echo.HeaderXRequestID, "theRequestID")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		theError := errors.New("the database error")
		e.HTTPErrorHandler(theError, c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), `"message":"Internal Server Error (request_id: theRequestID)"`)
	})

	t.Run("sends_http_error_status_and_message", func(t *testing.T) {
		e := echo.New()
		e.HTTPErrorHandler = ErrorHandler(e)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		httpError := echo.NewHTTPError(http.StatusNotFound, "the resource not found")
		e.HTTPErrorHandler(httpError, c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), `"message":"the resource not found"`)
	})

	t.Run("does_not_handle_already_committed_response", func(t *testing.T) {
		e := echo.New()
		e.HTTPErrorHandler = ErrorHandler(e)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Response().WriteHeader(http.StatusOK)
		c.Response().Committed = true

		theError := errors.New("the error")
		e.HTTPErrorHandler(theError, c)

		// Response should still be 200 from WriteHeader call above
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestRecoverWithErrorHandler(t *testing.T) {
	t.Run("panic_is_logged_at_error_level_with_stack_trace", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.HTTPErrorHandler = ErrorHandler(e)
		e.Use(echomiddleware.RecoverWithConfig(echomiddleware.RecoverConfig{
			LogErrorFunc: PanicLogger,
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(echo.HeaderXRequestID, "theRequestID")
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			panic("the panic message")
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"level":"ERROR"`)
		assert.Contains(t, logOutput, `"msg":"panic recovered"`)
		assert.Contains(t, logOutput, `"error":"the panic message"`)
		assert.Contains(t, logOutput, `"uri":"/test"`)
		assert.Contains(t, logOutput, `"method":"GET"`)
		assert.Contains(t, logOutput, `"stack"`)
	})

	t.Run("panic_is_logged_only_by_recover_middleware", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.HTTPErrorHandler = ErrorHandler(e)
		e.Use(echomiddleware.RecoverWithConfig(echomiddleware.RecoverConfig{
			LogErrorFunc: PanicLogger,
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(echo.HeaderXRequestID, "theRequestID")
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			panic("the panic message")
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		logLines := strings.Split(strings.TrimSpace(logOutput), "\n")
		assert.Len(t, logLines, 1)
		assert.Contains(t, logLines[0], `"msg":"panic recovered"`)
	})

	t.Run("panic_sends_500_response_with_request_id", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.HTTPErrorHandler = ErrorHandler(e)
		e.Use(echomiddleware.RecoverWithConfig(echomiddleware.RecoverConfig{
			LogErrorFunc: PanicLogger,
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(echo.HeaderXRequestID, "theRequestID")
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			panic("the panic message")
		})

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), `"message":"Internal Server Error (request_id: theRequestID)"`)
	})
}
