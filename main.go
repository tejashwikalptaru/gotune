// Package main is the production entry point for GoTune music player.
//
// GoTune is a cross-platform music player with clean architecture:
// - Event-driven communication (no callbacks)
// - Dependency injection for testability
// - MVP pattern for UI decoupling
// - Repository pattern for data persistence
//
// Build:
//
//	go build -o build/gotune .
//
// Run:
//
//	./build/gotune
package main

import (
	"log"
	"log/slog"

	"github.com/tejashwikalptaru/gotune/internal/app"
)

func main() {
	// Log version at startup
	versionInfo := app.GetVersionInfo()
	slog.Info(versionInfo.FullString())

	// Create default configuration
	config := app.DefaultConfig()

	// Use real BASS audio engine
	config.UseMockAudio = false

	// Create the application with dependency injection
	application, err := app.NewApplication(config)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Ensure a graceful shutdown
	defer func() {
		slog.Info("shutting down application")
		application.Shutdown()
		slog.Info("shutdown complete")
	}()

	// Run application (blocks until the window closed)
	application.Run()

	slog.Info("application exited cleanly")
}
