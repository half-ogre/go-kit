package logkit

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSlogWriter(t *testing.T) {
	t.Run("creates_writer_with_specified_level", func(t *testing.T) {
		theLevel := slog.LevelWarn

		writer := NewSlogWriter(theLevel)

		assert.NotNil(t, writer)
		assert.Equal(t, theLevel, writer.level)
	})

	t.Run("creates_writer_with_debug_level", func(t *testing.T) {
		theLevel := slog.LevelDebug

		writer := NewSlogWriter(theLevel)

		assert.Equal(t, theLevel, writer.level)
	})

	t.Run("creates_writer_with_info_level", func(t *testing.T) {
		theLevel := slog.LevelInfo

		writer := NewSlogWriter(theLevel)

		assert.Equal(t, theLevel, writer.level)
	})

	t.Run("creates_writer_with_error_level", func(t *testing.T) {
		theLevel := slog.LevelError

		writer := NewSlogWriter(theLevel)

		assert.Equal(t, theLevel, writer.level)
	})
}

func TestSlogWriter_Write(t *testing.T) {
	t.Run("writes_message_to_slog", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(theLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		writer := NewSlogWriter(slog.LevelInfo)
		theMessage := []byte("theLogMessage")

		n, err := writer.Write(theMessage)

		assert.NoError(t, err)
		assert.Equal(t, len(theMessage), n)
		logString := logOutput.String()
		assert.Contains(t, logString, "gin_debug")
		assert.Contains(t, logString, "message=theLogMessage")
	})

	t.Run("removes_trailing_newline", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(theLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		writer := NewSlogWriter(slog.LevelInfo)
		theMessageWithNewline := []byte("theLogMessage\n")

		n, err := writer.Write(theMessageWithNewline)

		assert.NoError(t, err)
		assert.Equal(t, len(theMessageWithNewline), n)
		logString := logOutput.String()
		assert.Contains(t, logString, "message=theLogMessage")
		// Verify the logged message doesn't contain the trailing newline
		assert.NotContains(t, strings.Split(logString, "message=")[1], "\\n")
	})

	t.Run("handles_empty_message", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(theLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		writer := NewSlogWriter(slog.LevelInfo)
		emptyMessage := []byte("")

		n, err := writer.Write(emptyMessage)

		assert.NoError(t, err)
		assert.Equal(t, 0, n)
		logString := logOutput.String()
		assert.Contains(t, logString, "gin_debug")
		assert.Contains(t, logString, "message=")
	})

	t.Run("handles_message_with_only_newline", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(theLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		writer := NewSlogWriter(slog.LevelInfo)
		newlineOnly := []byte("\n")

		n, err := writer.Write(newlineOnly)

		assert.NoError(t, err)
		assert.Equal(t, 1, n)
		logString := logOutput.String()
		assert.Contains(t, logString, "gin_debug")
		assert.Contains(t, logString, "message=")
	})

	t.Run("preserves_internal_newlines", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(theLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		writer := NewSlogWriter(slog.LevelInfo)
		theMultilineMessage := []byte("line1\nline2\nline3")

		n, err := writer.Write(theMultilineMessage)

		assert.NoError(t, err)
		assert.Equal(t, len(theMultilineMessage), n)
		logString := logOutput.String()
		assert.Contains(t, logString, "message=\"line1\\nline2\\nline3\"")
	})

	t.Run("logs_at_debug_level", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(theLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		writer := NewSlogWriter(slog.LevelDebug)
		theMessage := []byte("theDebugMessage")

		n, err := writer.Write(theMessage)

		assert.NoError(t, err)
		assert.Equal(t, len(theMessage), n)
		logString := logOutput.String()
		assert.Contains(t, logString, "level=DEBUG")
		assert.Contains(t, logString, "message=theDebugMessage")
	})

	t.Run("logs_at_info_level", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(theLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		writer := NewSlogWriter(slog.LevelInfo)
		theMessage := []byte("theInfoMessage")

		n, err := writer.Write(theMessage)

		assert.NoError(t, err)
		assert.Equal(t, len(theMessage), n)
		logString := logOutput.String()
		assert.Contains(t, logString, "level=INFO")
		assert.Contains(t, logString, "message=theInfoMessage")
	})

	t.Run("logs_at_warn_level", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(theLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		writer := NewSlogWriter(slog.LevelWarn)
		theMessage := []byte("theWarnMessage")

		n, err := writer.Write(theMessage)

		assert.NoError(t, err)
		assert.Equal(t, len(theMessage), n)
		logString := logOutput.String()
		assert.Contains(t, logString, "level=WARN")
		assert.Contains(t, logString, "message=theWarnMessage")
	})

	t.Run("logs_at_error_level", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(theLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		writer := NewSlogWriter(slog.LevelError)
		theMessage := []byte("theErrorMessage")

		n, err := writer.Write(theMessage)

		assert.NoError(t, err)
		assert.Equal(t, len(theMessage), n)
		logString := logOutput.String()
		assert.Contains(t, logString, "level=ERROR")
		assert.Contains(t, logString, "message=theErrorMessage")
	})

	t.Run("respects_logger_level_filtering", func(t *testing.T) {
		var logOutput bytes.Buffer
		// Set logger to only log WARN and above
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelWarn}))
		slog.SetDefault(theLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		writer := NewSlogWriter(slog.LevelDebug)
		theMessage := []byte("theDebugMessage")

		n, err := writer.Write(theMessage)

		assert.NoError(t, err)
		assert.Equal(t, len(theMessage), n)
		// Debug message should be filtered out by the logger
		assert.Empty(t, logOutput.String())
	})

	t.Run("uses_background_context", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(theLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		writer := NewSlogWriter(slog.LevelInfo)
		theMessage := []byte("theContextMessage")

		n, err := writer.Write(theMessage)

		assert.NoError(t, err)
		assert.Equal(t, len(theMessage), n)
		// Verify that the message was logged (context.Background() allows logging)
		logString := logOutput.String()
		assert.Contains(t, logString, "gin_debug")
		assert.Contains(t, logString, "message=theContextMessage")
	})

	t.Run("handles_unicode_characters", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(theLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		writer := NewSlogWriter(slog.LevelInfo)
		theUnicodeMessage := []byte("こんにちは 🌍")

		n, err := writer.Write(theUnicodeMessage)

		assert.NoError(t, err)
		assert.Equal(t, len(theUnicodeMessage), n)
		logString := logOutput.String()
		assert.Contains(t, logString, "gin_debug")
		assert.Contains(t, logString, "こんにちは 🌍")
	})

	t.Run("handles_special_characters", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(theLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		writer := NewSlogWriter(slog.LevelInfo)
		theSpecialMessage := []byte("message with \"quotes\" and 'apostrophes' and \\backslashes\\")

		n, err := writer.Write(theSpecialMessage)

		assert.NoError(t, err)
		assert.Equal(t, len(theSpecialMessage), n)
		logString := logOutput.String()
		assert.Contains(t, logString, "gin_debug")
		// The exact escaping depends on the handler, but the content should be there
		assert.Contains(t, logString, "quotes")
		assert.Contains(t, logString, "apostrophes")
		assert.Contains(t, logString, "backslashes")
	})

	t.Run("returns_correct_byte_count", func(t *testing.T) {
		var logOutput bytes.Buffer
		theLogger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(theLogger)
		t.Cleanup(func() { slog.SetDefault(slog.Default()) })

		writer := NewSlogWriter(slog.LevelInfo)
		theMessage := []byte("test message content")

		n, err := writer.Write(theMessage)

		assert.NoError(t, err)
		assert.Equal(t, len(theMessage), n)
	})
}
