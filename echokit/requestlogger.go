package echokit

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

// RequestLoggerConfig defines the configuration for the request logger middleware.
type RequestLoggerConfig struct {
	// DebugPaths is a list of paths that should be logged at DEBUG level instead of INFO.
	// All other paths will be logged at INFO level.
	DebugPaths []string
}

// RequestLogger returns a middleware that logs all HTTP requests with structured logging.
// All requests are logged at INFO level by default.
func RequestLogger() echo.MiddlewareFunc {
	return RequestLoggerWithConfig(RequestLoggerConfig{})
}

// RequestLoggerWithConfig returns a middleware that logs all HTTP requests with structured logging.
// Paths specified in config.DebugPaths are logged at DEBUG level, all others at INFO level.
func RequestLoggerWithConfig(config RequestLoggerConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			req := c.Request()
			res := c.Response()
			latency := time.Since(start)

			logLevel := slog.LevelInfo
			path := c.Path()
			for _, debugPath := range config.DebugPaths {
				if path == debugPath {
					logLevel = slog.LevelDebug
					break
				}
			}

			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}

			slog.Log(c.Request().Context(), logLevel, "request",
				"id", req.Header.Get(echo.HeaderXRequestID),
				"amzn_trace_id", req.Header.Get("X-Amzn-Trace-Id"),
				"remote_ip", c.RealIP(),
				"x_forwarded_for", req.Header.Get("X-Forwarded-For"),
				"x_forwarded_proto", req.Header.Get("X-Forwarded-Proto"),
				"host", req.Host,
				"method", req.Method,
				"uri", req.RequestURI,
				"user_agent", req.UserAgent(),
				"status", res.Status,
				"error", errMsg,
				"latency", latency.Nanoseconds(),
				"latency_human", latency.String(),
				"bytes_in", req.Header.Get(echo.HeaderContentLength),
				"bytes_out", res.Size,
			)

			return err
		}
	}
}

// PanicLogger logs panics at ERROR level with error message, stack trace, URI, and method.
// This function is meant to be used as the LogErrorFunc in echomiddleware.RecoverConfig.
func PanicLogger(c echo.Context, err error, stack []byte) error {
	slog.Error("panic recovered",
		"error", err.Error(),
		"stack", string(stack),
		"uri", c.Request().RequestURI,
		"method", c.Request().Method,
	)
	return err
}
