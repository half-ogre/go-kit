package logkit

import (
	"log/slog"
	"os"
)

func SetDefaultLogger(logLevel slog.Level) (*slog.Logger, *slog.LevelVar) {
	slogLevel := new(slog.LevelVar)
	slogLevel.Set(logLevel)
	slogLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel}))
	slog.SetDefault(slogLogger)
	return slogLogger, slogLevel
}
