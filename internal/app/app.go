// Package app provides application-level orchestration and dependency injection.
// This package wires together all components and manages the application lifecycle.
package app

import (
	"fmt"
	"log/slog"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"github.com/tejashwikalptaru/gotune/internal/adapter/audio/bass"
	"github.com/tejashwikalptaru/gotune/internal/adapter/audio/mock"
	"github.com/tejashwikalptaru/gotune/internal/adapter/eventbus"
	"github.com/tejashwikalptaru/gotune/internal/adapter/repository/memory"
	fyneui "github.com/tejashwikalptaru/gotune/internal/adapter/ui/fyne"
	"github.com/tejashwikalptaru/gotune/internal/logger"
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
	logger  *slog.Logger
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

	// LogLevel controls logging verbosity
	LogLevel slog.Level

	// TestFyneApp allows injecting a test Fyne app for testing (nil for production)
	TestFyneApp fyne.App
}

// DefaultConfig returns the default application configuration.
func DefaultConfig() Config {
	loggerCfg := logger.DefaultConfig()
	return Config{
		AppID:        "com.gotune.app",
		AppName:      "Go Tune",
		AudioDevice:  -1,
		SampleRate:   44100,
		UseMockAudio: false,
		LogLevel:     loggerCfg.Level,
	}
}

// NewApplication creates a new application with all dependencies wired.
// This is the main dependency injection function.
func NewApplication(config Config) (*Application, error) {
	app := &Application{}

	// Step 1: Create Fyne application
	if config.TestFyneApp != nil {
		app.fyneApp = config.TestFyneApp
	} else {
		app.fyneApp = fyneapp.NewWithID(config.AppID)
	}

	// Step 1.5: Create logger
	loggerCfg := logger.Config{
		Level:  config.LogLevel,
		Format: "text",
	}
	app.logger = logger.NewLogger(loggerCfg)
	app.logger.Info("initializing application",
		slog.String("app_id", config.AppID),
		slog.String("app_name", config.AppName))

	// Step 2: Create an event bus
	syncBus := eventbus.NewSyncEventBus()
	syncBus.SetLogger(app.logger.With(slog.String("component", "eventbus")))
	app.eventBus = syncBus

	// Step 3: Create an audio engine
	if config.UseMockAudio {
		engine := mock.NewEngine()
		engine.SetLogger(app.logger.With(slog.String("engine", "mock")))
		err := engine.Initialize(config.AudioDevice, config.SampleRate, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize audio engine: %w", err)
		}
		app.audioEngine = engine
	} else {
		engine := bass.NewEngine()
		engine.SetLogger(app.logger.With(slog.String("engine", "bass")))
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
		app.logger.With(slog.String("service", "playback")),
		app.audioEngine,
		app.eventBus,
	)

	app.playlistService = service.NewPlaylistService(
		app.logger.With(slog.String("service", "playlist")),
		app.playbackService,
		app.playlistRepo,
		app.historyRepo,
		app.eventBus,
	)

	app.libraryService = service.NewLibraryService(
		app.logger.With(slog.String("service", "library")),
		app.audioEngine,
		app.eventBus,
	)

	app.preferenceService = service.NewPreferenceService(
		app.logger.With(slog.String("service", "preference")),
		app.preferencesRepo,
		app.eventBus,
	)

	// Step 6: Load saved state
	if err := app.loadSavedState(); err != nil {
		// Non-fatal - just log and continue
		app.logger.Warn("failed to load saved state", slog.Any("error", err))
	}

	// Step 7: Create UI (Phase 8)
	app.mainWindow = fyneui.NewMainWindow(app.fyneApp)

	// Step 8: Create Presenter and wire with UI
	app.presenter = fyneui.NewPresenter(
		app.logger.With(slog.String("component", "presenter")),
		app.playbackService,
		app.playlistService,
		app.libraryService,
		app.preferenceService,
		app.eventBus,
		app.mainWindow,
	)

	// Connect presenter to the main window
	app.mainWindow.SetPresenter(app.presenter)

	// Set callback to save state before window closes
	// This ensures state is persisted even when quitting via Cmd+Q or window close button
	app.mainWindow.SetOnBeforeClose(func() {
		if err := app.saveState(); err != nil {
			app.logger.Warn("failed to save state on close", slog.Any("error", err))
		}
	})

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
		if err := a.playbackService.SetVolume(volume); err != nil {
			a.logger.Warn("failed to set volume", slog.Any("error", err))
		}
	}

	// Load saved loop mode
	loop := a.preferenceService.GetLoopMode()
	a.playbackService.SetLoop(loop)

	return nil
}

// Run starts the application.
// This is called from main.go after the application is created.
func (a *Application) Run() {
	a.logger.Info("GoTune Music Player started")
	a.logger.Info("all services initialized successfully")

	// Show and run UI (blocks until the window is closed)
	a.mainWindow.ShowAndRun()
}

// Shutdown gracefully shuts down the application.
// This should be called via deferring in main.go.
func (a *Application) Shutdown() {
	a.logger.Info("shutting down application")

	// Publish application stopping event
	// a.eventBus.Publish(domain.NewApplicationStoppingEvent())

	// Save the current state
	if err := a.saveState(); err != nil {
		a.logger.Warn("failed to save state", slog.Any("error", err))
	}

	// Shutdown UI and presenter
	if a.presenter != nil {
		a.presenter.Shutdown()
	}

	// Shutdown services (in reverse order of creation)
	if a.preferenceService != nil {
		if err := a.preferenceService.Shutdown(); err != nil {
			a.logger.Warn("failed to shutdown preference service", slog.Any("error", err))
		}
	}

	if a.libraryService != nil {
		if err := a.libraryService.Shutdown(); err != nil {
			a.logger.Warn("failed to shutdown library service", slog.Any("error", err))
		}
	}

	if a.playlistService != nil {
		if err := a.playlistService.Shutdown(); err != nil {
			a.logger.Warn("failed to shutdown playlist service", slog.Any("error", err))
		}
	}

	if a.playbackService != nil {
		if err := a.playbackService.Shutdown(); err != nil {
			a.logger.Warn("failed to shutdown playback service", slog.Any("error", err))
		}
	}

	// Shutdown audio engine
	if a.audioEngine != nil {
		if err := a.audioEngine.Shutdown(); err != nil {
			a.logger.Warn("failed to shutdown audio engine", slog.Any("error", err))
		}
	}

	a.logger.Info("application shutdown complete")
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
