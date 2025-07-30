package logkit

import (
	"context"
	"log/slog"
)

// SlogWriter implements io.Writer to redirect output to slog
type SlogWriter struct {
	level slog.Level
}

// NewSlogWriter creates a new SlogWriter that logs at the specified level
func NewSlogWriter(level slog.Level) *SlogWriter {
	return &SlogWriter{level: level}
}

func (w *SlogWriter) Write(p []byte) (n int, err error) {
	// Remove trailing newline if present
	message := string(p)
	if len(message) > 0 && message[len(message)-1] == '\n' {
		message = message[:len(message)-1]
	}

	// Log with the specified level
	slog.Log(context.Background(), w.level, "gin_debug", "message", message)
	return len(p), nil
}
