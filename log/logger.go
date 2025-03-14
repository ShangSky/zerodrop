package log

import (
	"log/slog"
	"os"
)

func New() *slog.Logger {
	logHandler := slog.NewJSONHandler(os.Stdout, nil)
	return slog.New(logHandler)
}
