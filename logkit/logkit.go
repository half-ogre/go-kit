package logkit

import (
	"log/slog"
	"os"
)

func SetDefualtLogger(logLevel slog.Level) {
	slogLevel := new(slog.LevelVar)
	slogLevel.Set(logLevel)
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel})))
}
