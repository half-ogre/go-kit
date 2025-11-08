package echokit

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRequestLogger(t *testing.T) {
	t.Run("logs_successful_request_at_info_level", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLogger())

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"level":"INFO"`)
		assert.Contains(t, logOutput, `"msg":"request"`)
		assert.Contains(t, logOutput, `"method":"GET"`)
		assert.Contains(t, logOutput, `"uri":"/test"`)
		assert.Contains(t, logOutput, `"status":200`)
	})

	t.Run("logs_at_info_level_by_default_when_no_debug_paths_configured", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLogger())

		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rec := httptest.NewRecorder()

		e.GET("/api/health", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"level":"INFO"`)
		assert.Contains(t, logOutput, `"msg":"request"`)
	})

	t.Run("logs_configured_paths_at_debug_level", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLoggerWithConfig(RequestLoggerConfig{
			DebugPaths: []string{"/api/health"},
		}))

		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rec := httptest.NewRecorder()

		e.GET("/api/health", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"level":"DEBUG"`)
		assert.Contains(t, logOutput, `"msg":"request"`)
	})

	t.Run("logs_multiple_configured_paths_at_debug_level", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLoggerWithConfig(RequestLoggerConfig{
			DebugPaths: []string{"/api/health", "/api/ready"},
		}))

		e.GET("/api/health", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})
		e.GET("/api/ready", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})
		e.GET("/api/other", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		// Test /api/health
		logBuf.Reset()
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"level":"DEBUG"`)

		// Test /api/ready
		logBuf.Reset()
		req = httptest.NewRequest(http.MethodGet, "/api/ready", nil)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		logOutput = logBuf.String()
		assert.Contains(t, logOutput, `"level":"DEBUG"`)

		// Test /api/other
		logBuf.Reset()
		req = httptest.NewRequest(http.MethodGet, "/api/other", nil)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		logOutput = logBuf.String()
		assert.Contains(t, logOutput, `"level":"INFO"`)
	})

	t.Run("logs_request_id_from_header", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLogger())

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"id":`)
	})

	t.Run("logs_remote_ip", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLogger())

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"remote_ip":"192.168.1.100"`)
	})

	t.Run("logs_amzn_trace_id_when_present", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLogger())

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Amzn-Trace-Id", "Root=1-67890abc-12345678901234567890abcd")
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"amzn_trace_id":"Root=1-67890abc-12345678901234567890abcd"`)
	})

	t.Run("logs_empty_amzn_trace_id_when_not_present", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLogger())

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"amzn_trace_id":""`)
	})

	t.Run("logs_x_forwarded_for_when_present", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLogger())

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.2")
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"x_forwarded_for":"203.0.113.1, 198.51.100.2"`)
	})

	t.Run("logs_empty_x_forwarded_for_when_not_present", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLogger())

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"x_forwarded_for":""`)
	})

	t.Run("logs_x_forwarded_proto_when_present", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLogger())

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Forwarded-Proto", "https")
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"x_forwarded_proto":"https"`)
	})

	t.Run("logs_empty_x_forwarded_proto_when_not_present", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLogger())

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"x_forwarded_proto":""`)
	})

	t.Run("logs_latency_in_nanoseconds_and_human_readable", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLogger())

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"latency":`)
		assert.Contains(t, logOutput, `"latency_human":`)
	})

	t.Run("logs_http_error_in_error_field", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLogger())

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusNotFound, "the resource not found")
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"level":"INFO"`)
		assert.Contains(t, logOutput, `"error":"code=404, message=the resource not found"`)
	})

	t.Run("logs_non_http_error_in_error_field", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLogger())

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			return errors.New("the database error")
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"level":"INFO"`)
		assert.Contains(t, logOutput, `"error":"the database error"`)
	})

	t.Run("logs_empty_error_field_when_no_error", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLogger())

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"error":""`)
	})

	t.Run("logs_host_and_user_agent", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLogger())

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Host = "example.com"
		req.Header.Set("User-Agent", "test-agent/1.0")
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"host":"example.com"`)
		assert.Contains(t, logOutput, `"user_agent":"test-agent/1.0"`)
	})

	t.Run("logs_bytes_in_and_bytes_out", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLogger())

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{"message": "hello"})
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"bytes_in":`)
		assert.Contains(t, logOutput, `"bytes_out":`)
	})

	t.Run("logs_panic_in_error_field_after_recover_middleware", func(t *testing.T) {
		var logBuf bytes.Buffer
		testLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(testLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		e := echo.New()
		e.Use(echomiddleware.RequestID())
		e.Use(RequestLogger())
		e.Use(echomiddleware.RecoverWithConfig(echomiddleware.RecoverConfig{
			LogErrorFunc: PanicLogger,
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		e.GET("/test", func(c echo.Context) error {
			panic("the panic message")
		})

		e.ServeHTTP(rec, req)

		logOutput := logBuf.String()
		// Should have two log entries: one from Recover at ERROR, one from RequestLogger at INFO
		assert.Contains(t, logOutput, `"level":"ERROR"`)
		assert.Contains(t, logOutput, `"msg":"panic recovered"`)
		assert.Contains(t, logOutput, `"level":"INFO"`)
		assert.Contains(t, logOutput, `"msg":"request"`)
		assert.Contains(t, logOutput, `"error":"the panic message"`)
	})
}
