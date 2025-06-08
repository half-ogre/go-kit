package echokit

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/labstack/echo/v4"
	echolog "github.com/labstack/gommon/log"
)

type LogWriter struct {
	logger *slog.Logger
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	lw.logger.Info(string(p))
	return len(p), nil
}

func NewEchoSlogLogger(
	logger *slog.Logger,
	level *slog.LevelVar) echo.Logger {
	return &EchoSlogLogger{
		level:  level,
		logger: logger,
		writer: &LogWriter{logger: logger},
	}
}

type EchoSlogLogger struct {
	level  *slog.LevelVar
	logger *slog.Logger
	writer *LogWriter
}

func (l *EchoSlogLogger) Output() io.Writer {
	return l.writer
}

func (l *EchoSlogLogger) SetOutput(w io.Writer) {
	panic("EchoSlogLogger#SetOutput not supported")
}

func (l *EchoSlogLogger) Prefix() string {
	// always return empty; prefix is not supported
	return ""
}

func (l *EchoSlogLogger) SetPrefix(p string) {
	panic("EchoSlogLogger#SetPrefix not supported")
}

func (l *EchoSlogLogger) Level() echolog.Lvl {
	switch l.level.Level() {
	case slog.LevelDebug:
		return echolog.DEBUG
	case slog.LevelInfo:
		return echolog.INFO
	case slog.LevelWarn:
		return echolog.WARN
	case slog.LevelError:
		return echolog.ERROR
	default:
		return echolog.INFO
	}
}

func (l *EchoSlogLogger) SetLevel(v echolog.Lvl) {
	switch v {
	case echolog.DEBUG:
		l.level.Set(slog.LevelDebug)
	case echolog.INFO:
		l.level.Set(slog.LevelInfo)
	case echolog.WARN:
		l.level.Set(slog.LevelWarn)
	case echolog.ERROR:
		l.level.Set(slog.LevelError)
	default:
		l.level.Set(slog.LevelDebug)
	}
}

func (l *EchoSlogLogger) SetHeader(h string) {
	panic("EchoSlogLogger#SetHeader not supported")
}

func (l *EchoSlogLogger) Print(i ...interface{}) {
	l.logger.Info(fmt.Sprint(i...))
}

func (l *EchoSlogLogger) Printf(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l EchoSlogLogger) Printj(j echolog.JSON) {
	l.logger.Info("", "j", j)
}

func (l EchoSlogLogger) Debug(i ...interface{}) {
	l.logger.Debug(fmt.Sprint(i...))
}

func (l EchoSlogLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug(fmt.Sprintf(format, args...))
}

func (l EchoSlogLogger) Debugj(j echolog.JSON) {
	l.logger.Debug("", "j", j)
}

func (l EchoSlogLogger) Info(i ...interface{}) {
	l.logger.Info(fmt.Sprint(i...))
}

func (l EchoSlogLogger) Infof(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l EchoSlogLogger) Infoj(j echolog.JSON) {
	l.logger.Info("", "j", j)
}

func (l EchoSlogLogger) Warn(i ...interface{}) {
	l.logger.Warn(fmt.Sprint(i...))
}

func (l EchoSlogLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

func (l EchoSlogLogger) Warnj(j echolog.JSON) {
	l.logger.Warn("", "j", j)
}

func (l EchoSlogLogger) Error(i ...interface{}) {
	l.logger.Error(fmt.Sprint(i...))
}

func (l EchoSlogLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

func (l EchoSlogLogger) Errorj(j echolog.JSON) {
	l.logger.Error("", "j", j)
}

func (l EchoSlogLogger) Fatal(i ...interface{}) {
	l.logger.Error(fmt.Sprint(i...))
	os.Exit(1)
}

func (l EchoSlogLogger) Fatalj(j echolog.JSON) {
	l.logger.Error("", "j", j)
	os.Exit(1)
}

func (l EchoSlogLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

func (l EchoSlogLogger) Panic(i ...interface{}) {
	l.logger.Error(fmt.Sprint(i...))
	panic(fmt.Sprint(i...))
}

func (l EchoSlogLogger) Panicj(j echolog.JSON) {
	l.logger.Error("", "j", j)
	panic(j)
}

func (l EchoSlogLogger) Panicf(format string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, args...))
	panic(fmt.Sprintf(format, args...))
}
