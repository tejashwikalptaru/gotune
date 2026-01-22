// Package logger provides test helpers for structured logging.
package logger

import (
	"log/slog"
	"os"
)

// NewTestLogger creates a logger for tests.
// By default, uses WARN level to keep test output quiet.
// Set TEST_DEBUG environment variable to enable debug logging in tests.
func NewTestLogger() *slog.Logger {
	level := slog.LevelWarn // Quiet by default

	// Allow tests to enable debug logging
	if os.Getenv("TEST_DEBUG") != "" {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
}
