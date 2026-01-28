// Package logger provides structured logging configuration using log/slog.
package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Config holds logger configuration.
type Config struct {
	Level  slog.Level
	Format string // "text" or "json"
}

// NewLogger creates a configured slog.Logger.
func NewLogger(cfg Config) *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: cfg.Level,
		// Add a source location for debug and error levels
		AddSource: cfg.Level <= slog.LevelDebug,
	}

	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	return slog.New(handler)
}

// DefaultConfig returns the default logger configuration.
// Parses the GOTUNE_LOG_LEVEL environment variable to set the log level.
// Valid values: DEBUG, INFO, WARN, WARNING, ERROR
// Default: INFO
func DefaultConfig() Config {
	level := slog.LevelInfo

	// Parse GOTUNE_LOG_LEVEL env var
	if envLevel := os.Getenv("GOTUNE_LOG_LEVEL"); envLevel != "" {
		switch strings.ToUpper(envLevel) {
		case "DEBUG":
			level = slog.LevelDebug
		case "INFO":
			level = slog.LevelInfo
		case "WARN", "WARNING":
			level = slog.LevelWarn
		case "ERROR":
			level = slog.LevelError
		}
	}

	return Config{
		Level:  level,
		Format: "text",
	}
}
