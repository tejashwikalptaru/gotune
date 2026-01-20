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
//	go build -o build/gotune ./cmd
//
// Run:
//
//	./build/gotune
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tejashwikalptaru/gotune/internal/app"
)

func main() {
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
		fmt.Println("\nShutting down...")
		if err := application.Shutdown(); err != nil {
			fmt.Fprintf(os.Stderr, "Shutdown error: %v\n", err)
		}
		fmt.Println("Shutdown complete")
	}()

	// Run application (blocks until the window closed)
	if err := application.Run(); err != nil {
		log.Printf("Application error: %v", err)
	}

	fmt.Println("Application exited cleanly")
}
