# Development Guide

This guide provides information for developers working on GoTune.

## Quick Start

```bash
# 1. Clone the repository
git clone <repository-url>
cd gotune

# 2. Setup BASS libraries
./scripts/setup-libs.sh

# 3. Build the application
make build

# 4. Run tests
make test

# 5. Run the application
make run
```

## Development Workflow

### Building During Development

For development with quick iterations:

```bash
# Build and run in one command
make run

# Or build and run separately
make build
./build/gotune

# Or use go run (slower, rebuilds each time)
go run ./cmd
```

### Running with Different Log Levels

GoTune uses structured logging with `log/slog`. Control log verbosity with the `GOTUNE_LOG_LEVEL` environment variable:

```bash
# Debug logging (very verbose - shows all operations)
GOTUNE_LOG_LEVEL=DEBUG ./build/gotune

# Info logging (default - important events only)
GOTUNE_LOG_LEVEL=INFO ./build/gotune

# Warning logging (quiet - only warnings and errors)
GOTUNE_LOG_LEVEL=WARN ./build/gotune

# Error logging (very quiet - only errors)
GOTUNE_LOG_LEVEL=ERROR ./build/gotune
```

The default log level is `INFO` if `GOTUNE_LOG_LEVEL` is not set.

### Development with Verbose Logging

During development, you may want to see all internal operations:

```bash
# Build once
make build

# Run with debug logging
GOTUNE_LOG_LEVEL=DEBUG ./build/gotune
```

## Testing

### Running Tests

```bash
# Run all tests (uses proper library paths automatically)
make test

# Run tests with race detection
make test-race

# Run tests for a specific package
go test ./internal/service -v

# Run a specific test
go test ./internal/service -run TestPlaybackService -v
```

### Test Logging

By default, tests use quiet logging (WARN level) to avoid cluttering test output. Enable debug logging in tests:

```bash
# Enable debug logs in tests
TEST_DEBUG=1 make test

# Or directly with go test
TEST_DEBUG=1 DYLD_LIBRARY_PATH=$(PWD)/build/libs/darwin go test ./internal/... -v
```

### Writing Tests

When writing tests, use the test logger helper:

```go
package service

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "gotune/internal/logger"
)

func TestMyFeature(t *testing.T) {
    // Create test logger (quiet by default, honors TEST_DEBUG env var)
    log := logger.NewTestLogger()

    // Create service with logger
    svc := NewMyService(log, ...)

    // Test your feature
    result := svc.DoSomething()
    assert.NotNil(t, result)
}
```

## Code Quality

### Linting

```bash
# Run linter
make lint

# Auto-fix linting issues where possible
make lint-fix

# Install or update golangci-lint
make lint-install
```

The project uses `golangci-lint` with configuration in `.golangci.yml`.

### Dead Code Detection

```bash
# Check for dead code
make deadcode

# Check dead code including test executables
make deadcode-test

# Unfiltered dead code check
make deadcode-unfiltered
```

### Running All Checks (CI Simulation)

Before pushing code, run the full CI workflow locally:

```bash
# Runs: lint + deadcode + build
make ci-local
```

This simulates what the CI system will check.

## Architecture Overview

GoTune follows **Clean Architecture** with strict layer separation and **Dependency Injection**.

### Directory Structure

```
gotune/
├── cmd/                          # Application entry point
│   └── main.go                  # Production entry point
├── internal/                    # Private application code
│   ├── app/                     # Application layer (DI root)
│   │   ├── app.go              # Dependency injection & orchestration
│   │   └── app_test.go         # Integration tests
│   ├── domain/                  # Domain layer (core business logic)
│   │   ├── models.go           # Domain models (Track, Playlist, etc.)
│   │   ├── events.go           # Event definitions
│   │   └── errors.go           # Domain errors
│   ├── ports/                   # Interface definitions (DI boundaries)
│   │   ├── audio.go            # AudioEngine interface
│   │   ├── eventbus.go         # EventBus interface
│   │   ├── repository.go       # Repository interfaces
│   │   └── ui.go               # UI port interfaces
│   ├── service/                 # Business logic layer
│   │   ├── player_service.go   # Playback control
│   │   ├── playlist_service.go # Playlist management
│   │   ├── library_service.go  # Library scanning
│   │   └── preference_service.go # User preferences
│   ├── adapter/                 # Infrastructure layer
│   │   ├── audio/
│   │   │   ├── bass/           # BASS library adapter
│   │   │   └── mock/           # Mock audio engine for tests
│   │   ├── repository/
│   │   │   └── memory/         # In-memory persistence (Fyne Preferences)
│   │   ├── eventbus/           # Event bus implementation
│   │   └── ui/
│   │       └── fyne/           # Fyne UI adapter (MVP pattern)
│   └── logger/                  # Logging configuration
│       ├── logger.go           # Logger setup and config
│       └── testing.go          # Test logger helper
├── libs/                        # BASS library archives (ZIP files)
└── build/                       # Build artifacts
    └── libs/                    # Extracted BASS libraries
```

### Architectural Layers

```
┌─────────────────────────────────────────┐
│         cmd/main.go (Entry Point)       │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│        internal/app (DI Root)           │
│  - Creates all dependencies             │
│  - Wires components together            │
│  - Manages application lifecycle        │
└─────────────────┬───────────────────────┘
                  │
      ┌───────────┴───────────┬──────────────┐
      ▼                       ▼              ▼
┌─────────────┐      ┌──────────────┐  ┌─────────────┐
│   Service   │      │   Adapter    │  │   Domain    │
│    Layer    │      │    Layer     │  │    Layer    │
│             │      │              │  │             │
│ - Player    │      │ - Audio      │  │ - Models    │
│ - Playlist  │      │ - Repository │  │ - Events    │
│ - Library   │      │ - EventBus   │  │ - Errors    │
│ - Preference│      │ - UI (Fyne)  │  │             │
└─────────────┘      └──────────────┘  └─────────────┘
      │                       │
      └───────────┬───────────┘
                  ▼
          ┌──────────────┐
          │    Ports     │
          │ (Interfaces) │
          └──────────────┘
```

### Dependency Flow

All dependencies flow **inward** toward the service layer:

```
cmd/main.go
    → app.NewApplication(config)
        → logger := logger.NewLogger(cfg)
        → eventBus := eventbus.NewSyncEventBus()
        → audioEngine := bass.NewEngine() or mock.NewEngine()
        → repositories := memory.New*(fyneApp.Preferences())
        → services := service.New*(logger, engine, bus, repos)
        → presenter := fyne.NewPresenter(logger, services...)
        → window := fyne.NewMainWindow()
        → presenter.AttachWindow(window)
```

**Key Principles:**
- Services depend only on **ports** (interfaces), never on concrete adapters
- Adapters implement ports
- Domain layer has **zero external dependencies**
- All wiring happens in the `app` package

### Dependency Injection Pattern

GoTune uses **constructor-based dependency injection**:

```go
// Service constructor
func NewPlaybackService(
    logger *slog.Logger,           // Injected logger
    engine ports.AudioEngine,      // Injected audio engine (interface)
    bus ports.EventBus,            // Injected event bus (interface)
) *PlaybackService {
    return &PlaybackService{
        logger: logger,
        engine: engine,
        bus:    bus,
        // ...
    }
}
```

**Benefits:**
- Easy to test (inject mocks)
- Explicit dependencies (visible in constructors)
- No global state
- No framework magic

### Event-Driven Communication

Services communicate via an event bus, not direct method calls:

```go
// Service A publishes an event
bus.Publish(domain.NewTrackLoadedEvent(track, handle))

// Service B subscribes to the event
bus.Subscribe(domain.EventTrackLoaded, func(event domain.Event) {
    // Handle the event
    e := event.(*domain.TrackLoadedEvent)
    // ...
})
```

**Event Types:**
- **Playback events**: track.loaded, track.started, track.paused, track.stopped, track.completed
- **Volume events**: volume.changed, mute.toggled
- **Playlist events**: playlist.updated, queue.changed
- **Library events**: scan.started, scan.progress, scan.completed

**Benefits:**
- Decouples services (no direct dependencies)
- Easy to add new features (just subscribe to events)
- Thread-safe communication

### MVP Pattern in UI

The UI uses the **Model-View-Presenter** pattern:

```
┌──────────────┐       ┌─────────────┐      ┌──────────────┐
│  MainWindow  │◄──────┤  Presenter  │─────►│   Services   │
│   (View)     │       │  (Logic)    │      │   (Model)    │
│              │       │             │      │              │
│ - Dumb       │       │ - Smart     │      │ - Business   │
│ - No logic   │       │ - UI logic  │      │   logic      │
│ - Receives   │       │ - Event     │      │ - State      │
│   updates    │       │   handlers  │      │   management │
└──────────────┘       └─────────────┘      └──────────────┘
```

**Roles:**
- **View (MainWindow)**: "Dumb" UI that only displays data and captures user input
- **Presenter**: Contains all UI logic, subscribes to domain events, updates the view
- **Model (Services)**: Business logic layer, no knowledge of UI

## Adding a New Feature

Follow these steps to add a new feature:

### 1. Define Domain Model (if needed)

Add new types to `internal/domain/models.go`:

```go
// Example: Adding a favorites feature
type Favorite struct {
    TrackID   string
    AddedAt   time.Time
}
```

### 2. Define Events (if needed)

Add event types to `internal/domain/events.go`:

```go
const (
    EventFavoriteAdded   EventType = "favorite.added"
    EventFavoriteRemoved EventType = "favorite.removed"
)

type FavoriteAddedEvent struct {
    baseEvent
    TrackID string
}
```

### 3. Define Port Interface (if needed)

Add interface to `internal/ports/repository.go`:

```go
type FavoriteRepository interface {
    AddFavorite(trackID string) error
    RemoveFavorite(trackID string) error
    GetAllFavorites() ([]string, error)
    IsFavorite(trackID string) bool
}
```

### 4. Implement in Service Layer

Create `internal/service/favorite_service.go`:

```go
package service

import (
    "log/slog"
    "gotune/internal/domain"
    "gotune/internal/ports"
)

type FavoriteService struct {
    logger *slog.Logger
    repo   ports.FavoriteRepository
    bus    ports.EventBus
}

func NewFavoriteService(
    logger *slog.Logger,
    repo ports.FavoriteRepository,
    bus ports.EventBus,
) *FavoriteService {
    return &FavoriteService{
        logger: logger,
        repo:   repo,
        bus:    bus,
    }
}

func (s *FavoriteService) AddFavorite(trackID string) error {
    s.logger.Info("adding favorite", slog.String("track_id", trackID))

    if err := s.repo.AddFavorite(trackID); err != nil {
        s.logger.Error("failed to add favorite", slog.Any("error", err))
        return err
    }

    // Publish event
    s.bus.Publish(domain.NewFavoriteAddedEvent(trackID))
    return nil
}
```

### 5. Create Adapter (if needed)

Implement repository in `internal/adapter/repository/memory/favorite.go`:

```go
package memory

import (
    "fyne.io/fyne/v2"
    "gotune/internal/domain"
)

type FavoriteRepository struct {
    prefs fyne.Preferences
}

func NewFavoriteRepository(prefs fyne.Preferences) *FavoriteRepository {
    return &FavoriteRepository{prefs: prefs}
}

func (r *FavoriteRepository) AddFavorite(trackID string) error {
    // Implementation...
}
```

### 6. Wire in Dependency Injection

Update `internal/app/app.go`:

```go
type Application struct {
    // ... existing fields ...
    favoriteService *service.FavoriteService
    favoriteRepo    ports.FavoriteRepository
}

func NewApplication(config Config) (*Application, error) {
    // ... existing setup ...

    // Create favorite repository
    app.favoriteRepo = memory.NewFavoriteRepository(app.fyneApp.Preferences())

    // Create favorite service
    app.favoriteService = service.NewFavoriteService(
        app.logger.With(slog.String("service", "favorite")),
        app.favoriteRepo,
        app.eventBus,
    )

    // ... rest of setup ...
}
```

### 7. Add UI (if needed)

Update `internal/adapter/ui/fyne/presenter.go`:

```go
type Presenter struct {
    // ... existing fields ...
    favoriteService *service.FavoriteService
}

func NewPresenter(
    logger *slog.Logger,
    // ... existing params ...
    favoriteService *service.FavoriteService,
) *Presenter {
    // ... setup ...
}

func (p *Presenter) OnAddFavoriteClicked() {
    // Handle UI event
    if err := p.favoriteService.AddFavorite(currentTrackID); err != nil {
        p.logger.Error("failed to add favorite", slog.Any("error", err))
        // Show error dialog
    }
}
```

### 8. Write Tests

Create `internal/service/favorite_service_test.go`:

```go
package service

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "gotune/internal/logger"
)

func TestAddFavorite(t *testing.T) {
    log := logger.NewTestLogger()
    mockRepo := &MockFavoriteRepository{}
    mockBus := &MockEventBus{}

    svc := NewFavoriteService(log, mockRepo, mockBus)

    mockRepo.On("AddFavorite", "track123").Return(nil)
    mockBus.On("Publish", mock.Anything).Return()

    err := svc.AddFavorite("track123")

    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
    mockBus.AssertExpectations(t)
}
```

## BASS Audio Library

### Platform-Specific Configuration

GoTune uses CGO to interface with the BASS audio library. Platform-specific configuration is handled via build tags:

**Files:**
- `internal/adapter/audio/bass/platform_darwin.go` - macOS (build tag: `//go:build darwin`)
- `internal/adapter/audio/bass/platform_linux.go` - Linux (build tag: `//go:build linux`)
- `internal/adapter/audio/bass/platform_windows.go` - Windows (build tag: `//go:build windows`)

**CGO Directives:**
Each platform file contains CGO compiler and linker flags:

```go
//go:build darwin

package bass

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo LDFLAGS: -L${SRCDIR}/../../../../build/libs/darwin -lbass
#include "bass.h"
*/
import "C"
```

Go's build system automatically selects the correct file based on the target platform.

### Using Mock Audio Engine for Tests

For tests that don't require actual audio playback, use the mock engine:

```go
import "gotune/internal/adapter/audio/mock"

func TestSomething(t *testing.T) {
    logger := logger.NewTestLogger()
    engine := mock.NewEngine()
    engine.SetLogger(logger)

    // Use engine in tests
    handle, err := engine.LoadTrack("test.mp3")
    // ...
}
```

The mock engine simulates audio operations without requiring BASS libraries.

## Logging Best Practices

### Using Structured Logging

Use `slog` for all logging with structured key-value pairs:

```go
// Good: Structured logging
s.logger.Info("track loaded",
    slog.String("file_path", track.FilePath),
    slog.Int64("duration_ms", track.Duration),
)

// Bad: String concatenation
s.logger.Info(fmt.Sprintf("Track loaded: %s", track.FilePath))
```

### Log Levels

Choose appropriate log levels:

```go
// Debug: Verbose operational information (disabled in production)
s.logger.Debug("entering function", slog.String("track_id", id))

// Info: Important state changes and events
s.logger.Info("playback started", slog.String("track", track.Title))

// Warn: Recoverable errors, degraded functionality
s.logger.Warn("failed to save state", slog.Any("error", err))

// Error: Errors that prevent operation from completing
s.logger.Error("failed to load track",
    slog.String("file_path", path),
    slog.Any("error", err))
```

### Logger Context

Add context to loggers for better tracing:

```go
// In app.go
service := NewPlaybackService(
    app.logger.With(slog.String("service", "playback")),  // Add context
    engine,
    bus,
)

// All logs from this service will include "service=playback"
```

## Common Development Tasks

### Running a Single Test

```bash
go test ./internal/service -run TestPlaybackService_LoadTrack -v
```

### Running Tests for a Package

```bash
go test ./internal/service -v
```

### Running Tests with Coverage

```bash
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Checking for Race Conditions

```bash
make test-race
```

### Cleaning Build Artifacts

```bash
make clean
```

### Packaging for Distribution

```bash
make package
```

This creates a Fyne-packaged bundle with the BASS libraries included.

## Git Workflow

### Committing Changes

1. Make your changes
2. Run quality checks:
   ```bash
   make ci-local
   ```
3. Run tests:
   ```bash
   make test
   ```
4. Stage and commit:
   ```bash
   git add .
   git commit -m "Add feature: description"
   ```

### Before Pushing

Always run the full test suite and linting:

```bash
make ci-local && make test
```

## Additional Resources

- [BUILD.md](BUILD.md) - Build instructions and troubleshooting
- [Go Documentation](https://golang.org/doc/)
- [Fyne GUI Documentation](https://developer.fyne.io/)
- [BASS Audio Library](https://www.un4seen.com/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [log/slog Documentation](https://pkg.go.dev/log/slog)

## Getting Help

If you encounter issues:

1. Check [BUILD.md](BUILD.md) for build-related problems
2. Run `make test` to verify your environment is set up correctly
3. Check the project's issue tracker
4. Review the architecture documentation above
