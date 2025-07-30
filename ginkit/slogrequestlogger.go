package ginkit

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

type SlogRequestLoggerOption func(*SlogRequestLoggerConfig)

type SlogRequestLoggerConfig struct {
	Logger *slog.Logger
}

func WithLogger(logger *slog.Logger) SlogRequestLoggerOption {
	return func(c *SlogRequestLoggerConfig) {
		c.Logger = logger
	}
}

func SlogRequestLogger(options ...SlogRequestLoggerOption) gin.HandlerFunc {
	config := &SlogRequestLoggerConfig{}
	for _, option := range options {
		option(config)
	}

	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate request processing time
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		// Determine log level based on path
		var logLevel slog.Level
		if path == "/api/v1/service/health" {
			logLevel = slog.LevelDebug
		} else {
			logLevel = slog.LevelInfo
		}

		logger.Log(c.Request.Context(), logLevel, "Request completed",
			"method", method,
			"path", path,
			"status", statusCode,
			"latency", latency,
			"client_ip", clientIP,
			"body_size", bodySize,
		)
	}
}
