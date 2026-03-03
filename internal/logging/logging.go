package logging

import (
	"log/slog"
	"os"
)

// New initializes a new slog.Logger.
func New(debug bool) *slog.Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	
	opts := &slog.HandlerOptions{
		Level: level,
	}
	
	handler := slog.NewTextHandler(os.Stderr, opts)
	return slog.New(handler)
}
