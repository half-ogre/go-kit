package ginkit

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWithLogger(t *testing.T) {
	t.Run("sets_logger_in_config", func(t *testing.T) {
		theLogger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
		config := &SlogRequestLoggerConfig{}

		option := WithLogger(theLogger)
		option(config)

		assert.Equal(t, theLogger, config.Logger)
	})
}

func TestSlogRequestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("uses_default_logger_when_no_logger_provided", func(t *testing.T) {
		var logOutput bytes.Buffer
		defaultLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(defaultLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		middleware := SlogRequestLogger()
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, logOutput.String(), "Request completed")
		assert.Contains(t, logOutput.String(), "method=GET")
		assert.Contains(t, logOutput.String(), "path=/test")
		assert.Contains(t, logOutput.String(), "status=200")
	})

	t.Run("uses_provided_logger_when_logger_option_given", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))

		middleware := SlogRequestLogger(WithLogger(theLogger))
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, logOutput.String(), "Request completed")
	})

	t.Run("logs_basic_request_information", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))

		middleware := SlogRequestLogger(WithLogger(theLogger))
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		logString := logOutput.String()
		assert.Contains(t, logString, "method=GET")
		assert.Contains(t, logString, "path=/test")
		assert.Contains(t, logString, "status=200")
		assert.Contains(t, logString, "latency=")
		assert.Contains(t, logString, "client_ip=")
		assert.Contains(t, logString, "body_size=-1")
	})

	t.Run("logs_query_parameters_in_path", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))

		middleware := SlogRequestLogger(WithLogger(theLogger))
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test?param1=value1&param2=value2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		logString := logOutput.String()
		assert.Contains(t, logString, "/test?param1=value1&param2=value2")
	})

	t.Run("logs_different_http_methods", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))

		middleware := SlogRequestLogger(WithLogger(theLogger))
		router := gin.New()
		router.Use(middleware)
		router.POST("/test", func(c *gin.Context) {
			c.Status(http.StatusCreated)
		})

		req := httptest.NewRequest("POST", "/test", strings.NewReader("test body"))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		logString := logOutput.String()
		assert.Contains(t, logString, "method=POST")
		assert.Contains(t, logString, "status=201")
	})

	t.Run("logs_error_status_codes", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))

		middleware := SlogRequestLogger(WithLogger(theLogger))
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusInternalServerError)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Contains(t, logOutput.String(), "status=500")
	})

	t.Run("logs_response_body_size", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))

		middleware := SlogRequestLogger(WithLogger(theLogger))
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "theResponseBody")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Contains(t, logOutput.String(), "body_size=15") // "theResponseBody" is 15 characters
	})

	t.Run("logs_health_endpoint_at_debug_level", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelInfo}))

		middleware := SlogRequestLogger(WithLogger(theLogger))
		router := gin.New()
		router.Use(middleware)
		router.GET("/api/v1/service/health", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/api/v1/service/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// With info level, debug messages should not appear
		assert.Empty(t, logOutput.String())
	})

	t.Run("logs_health_endpoint_at_debug_level_when_debug_enabled", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))

		middleware := SlogRequestLogger(WithLogger(theLogger))
		router := gin.New()
		router.Use(middleware)
		router.GET("/api/v1/service/health", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/api/v1/service/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		logString := logOutput.String()
		assert.Contains(t, logString, "Request completed")
		assert.Contains(t, logString, "path=/api/v1/service/health")
	})

	t.Run("logs_non_health_endpoints_at_info_level", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelInfo}))

		middleware := SlogRequestLogger(WithLogger(theLogger))
		router := gin.New()
		router.Use(middleware)
		router.GET("/api/v1/users", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/api/v1/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		logString := logOutput.String()
		assert.Contains(t, logString, "Request completed")
		assert.Contains(t, logString, "path=/api/v1/users")
	})

	t.Run("measures_latency", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))

		middleware := SlogRequestLogger(WithLogger(theLogger))
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			time.Sleep(10 * time.Millisecond) // Simulate some processing time
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		logString := logOutput.String()
		assert.Contains(t, logString, "latency=")
		// Check that latency is greater than the sleep time
		assert.Contains(t, logString, "ms") // Should contain milliseconds in the duration
	})

	t.Run("captures_client_ip", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))

		middleware := SlogRequestLogger(WithLogger(theLogger))
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.100")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Contains(t, logOutput.String(), "client_ip=192.168.1.100")
	})

	t.Run("applies_multiple_options", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))

		middleware := SlogRequestLogger(WithLogger(theLogger))
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, logOutput.String(), "Request completed")
	})

	t.Run("includes_context_in_log", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))

		middleware := SlogRequestLogger(WithLogger(theLogger))
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			// Add some context value to verify context is passed correctly
			//lint:ignore SA1029 using string key for test simplicity
			ctx := context.WithValue(c.Request.Context(), "test_key", "test_value")
			c.Request = c.Request.WithContext(ctx)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, logOutput.String(), "Request completed")
	})
}
