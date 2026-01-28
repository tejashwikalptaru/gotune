# GoTune Architecture

GoTune is designed using the principles of **Clean Architecture**. This architectural style promotes a separation of concerns, resulting in a system that is:

- **Independent of Frameworks:** The core business logic is not tied to any specific framework like Fyne.
- **Testable:** Components can be tested in isolation.
- **Independent of UI:** The UI can be swapped out without changing the rest of the system.
- **Independent of Database:** The persistence layer can be easily replaced.
- **Independent of any external agency:** The core business logic doesn't know anything about the outside world.

## Layers

The application is divided into several layers, with a strict dependency rule: **dependencies can only point inwards**.

![Clean Architecture Diagram](https://blog.cleancoder.com/uncle-bob/images/2012-08-13-the-clean-architecture/CleanArchitecture.jpg)

Here's how GoTune's layers map to the Clean Architecture diagram:

```
┌─────────────────────────────────────────┐
│           main.go (Entry Point)         │
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

### 1. Domain

The `internal/domain` package represents the core of the application. It contains the business models, events, and errors.

- **`models.go`:** Defines the core data structures, such as `Track`, `Playlist`, and `Library`. These are plain Go structs with no external dependencies.
- **`events.go`:** Defines the events that are used for communication between different parts of the application.
- **`errors.go`:** Defines custom error types.

The domain layer has **zero dependencies** on any other part of the application.

### 2. Ports

The `internal/ports` package defines the interfaces that allow communication between the application's core and the outside world. These interfaces are the "ports" in the Hexagonal Architecture pattern.

- **`audio.go`:** Defines the `AudioEngine` interface for audio playback.
- **`eventbus.go`:** Defines the `EventBus` interface for event-driven communication.
- **`repository.go`:** Defines the interfaces for data persistence, such as `PlaylistRepository` and `HistoryRepository`.
- **`ui.go`:** Defines the interfaces for the user interface.

Services depend on these interfaces, not on their concrete implementations.

### 3. Service

The `internal/service` package contains the application-specific business logic. It orchestrates the flow of data between the domain and the adapters, using the interfaces defined in the `ports` package.

- **`player_service.go`:** Manages audio playback.
- **`playlist_service.go`:** Manages playlists.
- **`library_service.go`:** Manages the music library.
- **`preference_service.go`:** Manages user preferences.

Services are not aware of the specific implementation details of the adapters. They only know about the interfaces.

### 4. Adapter

The `internal/adapter` package contains the concrete implementations of the interfaces defined in the `ports` package. These are the "adapters" in the Hexagonal Architecture pattern.

- **`audio/bass`:** An implementation of the `AudioEngine` interface using the BASS audio library.
- **`audio/mock`:** A mock implementation of the `AudioEngine` interface for testing.
- **`eventbus`:** An implementation of the `EventBus` interface.
- **`repository/memory`:** An implementation of the repository interfaces using in-memory storage.
- **`ui/fyne`:** An implementation of the UI interfaces using the Fyne GUI framework.

### 5. App

The `internal/app` package is the dependency injection root. It is responsible for creating all the components of the application and wiring them together.

- **`app.go`:** Creates instances of the services, repositories, and other components, and injects them into the components that need them.
- **`version.go`:** Provides version information.

### 6. Main

The `main.go` file is the entry point of the application. It creates an instance of the `Application` from the `app` package and runs it.

## Communication

Communication between the layers is done through two primary mechanisms:

### 1. Dependency Injection

GoTune uses constructor-based dependency injection. This means that a component's dependencies are provided to it through its constructor.

```go
// Service constructor
func NewPlaybackService(
    logger *slog.Logger,
    engine ports.AudioEngine,
    bus ports.EventBus,
) *PlaybackService {
    return &PlaybackService{
        logger: logger,
        engine: engine,
        bus:    bus,
    }
}
```

This makes it easy to replace dependencies with mock implementations for testing.

### 2. Event Bus

GoTune uses an event-driven architecture for communication between services. When something happens in one service, it publishes an event to the event bus. Other services can subscribe to these events and react to them.

```go
// Service A publishes an event
bus.Publish(domain.NewTrackLoadedEvent(track, handle))

// Service B subscribes to the event
bus.Subscribe(domain.EventTrackLoaded, func(event domain.Event) {
    // Handle the event
})
```

This decouples the services from each other and makes it easy to add new features without modifying existing code.

## UI (MVP Pattern)

The UI is implemented using the **Model-View-Presenter (MVP)** pattern.

- **Model:** The services in the `internal/service` package act as the model. They contain the business logic and the application's state.
- **View:** The Fyne widgets in the `internal/adapter/ui/fyne` package act as the view. They are responsible for displaying the UI and capturing user input. They are "dumb" components with no business logic.
- **Presenter:** The `Presenter` in `internal/adapter/ui/fyne/presenter.go` acts as the presenter. It mediates between the model and the view. It listens to events from the services and updates the view accordingly. It also listens to user input from the view and calls the appropriate methods on the services.

This separation of concerns makes the UI easier to test and maintain.
