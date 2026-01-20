// Package app provides application-level orchestration and dependency injection.
// This package wires together all components and manages the application lifecycle.
package app

import (
	"fmt"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"github.com/tejashwikalptaru/gotune/internal/adapter/audio/bass"
	"github.com/tejashwikalptaru/gotune/internal/adapter/audio/mock"
	"github.com/tejashwikalptaru/gotune/internal/adapter/eventbus"
	"github.com/tejashwikalptaru/gotune/internal/adapter/repository/memory"
	fyneui "github.com/tejashwikalptaru/gotune/internal/adapter/ui/fyne"
	"github.com/tejashwikalptaru/gotune/internal/ports"
	"github.com/tejashwikalptaru/gotune/internal/service"
)

// Application is the root application structure that holds all dependencies.
// It follows the Dependency Injection pattern with constructor-based injection.
//
// The Application struct is responsible for:
// - Creating and wiring all dependencies
// - Managing the application lifecycle (startup, shutdown)
// - Providing a clean entry point for main.go
type Application struct {
	// Core dependencies
	fyneApp fyne.App

	// Infrastructure
	eventBus    ports.EventBus
	audioEngine ports.AudioEngine

	// Repositories
	historyRepo     ports.HistoryRepository
	playlistRepo    ports.PlaylistRepository
	preferencesRepo ports.PreferencesRepository

	// Services
	playbackService   *service.PlaybackService
	playlistService   *service.PlaylistService
	libraryService    *service.LibraryService
	preferenceService *service.PreferenceService

	// UI (Phase 8)
	presenter  *fyneui.Presenter
	mainWindow *fyneui.MainWindow
}

// Config holds application configuration.
type Config struct {
	// AppID is the unique application identifier
	AppID string

	// AppName is the display name
	AppName string

	// AudioDevice is the audio output device (-1 for default)
	AudioDevice int

	// SampleRate is the audio sample rate
	SampleRate int

	// UseMockAudio determines whether to use a mock audio engine (for testing)
	UseMockAudio bool
}

// DefaultConfig returns the default application configuration.
func DefaultConfig() Config {
	return Config{
		AppID:        "com.gotune.app",
		AppName:      "GoTune",
		AudioDevice:  -1,
		SampleRate:   44100,
		UseMockAudio: false,
	}
}

// NewApplication creates a new application with all dependencies wired.
// This is the main dependency injection function.
func NewApplication(config Config) (*Application, error) {
	app := &Application{}

	// Step 1: Create Fyne application
	app.fyneApp = fyneapp.NewWithID(config.AppID)

	// Step 2: Create an event bus
	app.eventBus = eventbus.NewSyncEventBus()

	// Step 3: Create an audio engine
	if config.UseMockAudio {
		engine := mock.NewEngine()
		err := engine.Initialize(config.AudioDevice, config.SampleRate, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize audio engine: %w", err)
		}
		app.audioEngine = engine
	} else {
		engine := bass.NewEngine()
		err := engine.Initialize(config.AudioDevice, config.SampleRate, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize audio engine: %w", err)
		}
		app.audioEngine = engine
	}

	// Step 4: Create repositories
	prefs := app.fyneApp.Preferences()
	app.historyRepo = memory.NewHistoryRepository(prefs)
	app.playlistRepo = memory.NewPlaylistRepository(prefs)
	app.preferencesRepo = memory.NewPreferencesRepository(prefs)

	// Step 5: Create services (with dependency injection)
	app.playbackService = service.NewPlaybackService(
		app.audioEngine,
		app.eventBus,
	)

	app.playlistService = service.NewPlaylistService(
		app.playbackService,
		app.playlistRepo,
		app.historyRepo,
		app.eventBus,
	)

	app.libraryService = service.NewLibraryService(
		app.audioEngine,
		app.eventBus,
	)

	app.preferenceService = service.NewPreferenceService(
		app.preferencesRepo,
		app.eventBus,
	)

	// Step 6: Load saved state
	if err := app.loadSavedState(); err != nil {
		// Non-fatal - just log and continue
		fmt.Printf("Warning: Failed to load saved state: %v\n", err)
	}

	// Step 7: Create UI (Phase 8)
	app.mainWindow = fyneui.NewMainWindow(app.fyneApp)

	// Step 8: Create Presenter and wire with UI
	app.presenter = fyneui.NewPresenter(
		app.playbackService,
		app.playlistService,
		app.libraryService,
		app.preferenceService,
		app.eventBus,
		app.mainWindow,
	)

	// Connect presenter to the main window
	app.mainWindow.SetPresenter(app.presenter)

	return app, nil
}

// loadSavedState restores the application state from the previous session.
func (a *Application) loadSavedState() error {
	// Load saved queue and position
	err := a.playlistService.LoadQueue()
	if err != nil {
		return fmt.Errorf("failed to load queue: %w", err)
	}

	// Load saved volume
	volume := a.preferenceService.GetVolume()
	if volume > 0 {
		a.playbackService.SetVolume(volume)
	}

	// Load saved loop mode
	loop := a.preferenceService.GetLoopMode()
	a.playbackService.SetLoop(loop)

	return nil
}

// Run starts the application.
// This is called from main.go after the application is created.
func (a *Application) Run() error {
	fmt.Println("GoTune Music Player")
	fmt.Println("All services initialized successfully")

	// Show and run UI (blocks until the window is closed)
	a.mainWindow.ShowAndRun()

	return nil
}

// Shutdown gracefully shuts down the application.
// This should be called via deferring in main.go.
func (a *Application) Shutdown() error {
	fmt.Println("Shutting down application...")

	// Publish application stopping event
	// a.eventBus.Publish(domain.NewApplicationStoppingEvent())

	// Save the current state
	if err := a.saveState(); err != nil {
		fmt.Printf("Warning: Failed to save state: %v\n", err)
	}

	// Shutdown UI and presenter
	if a.presenter != nil {
		a.presenter.Shutdown()
	}

	// Shutdown services (in reverse order of creation)
	if a.preferenceService != nil {
		a.preferenceService.Shutdown()
	}

	if a.libraryService != nil {
		a.libraryService.Shutdown()
	}

	if a.playlistService != nil {
		a.playlistService.Shutdown()
	}

	if a.playbackService != nil {
		a.playbackService.Shutdown()
	}

	// Shutdown audio engine
	if a.audioEngine != nil {
		a.audioEngine.Shutdown()
	}

	fmt.Println("Application shutdown complete")
	return nil
}

// saveState persists the current application state.
func (a *Application) saveState() error {
	// Save current queue
	err := a.playlistService.SaveQueue()
	if err != nil {
		return fmt.Errorf("failed to save queue: %w", err)
	}

	return nil
}

// GetServices returns the service instances.
// This is useful for testing or advanced usage.
func (a *Application) GetServices() (
	*service.PlaybackService,
	*service.PlaylistService,
	*service.LibraryService,
	*service.PreferenceService,
) {
	return a.playbackService, a.playlistService, a.libraryService, a.preferenceService
}

// GetEventBus returns the event bus instance.
func (a *Application) GetEventBus() ports.EventBus {
	return a.eventBus
}

// GetFyneApp returns the Fyne application instance.
func (a *Application) GetFyneApp() fyne.App {
	return a.fyneApp
}
