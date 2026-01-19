# Gemini Codebase Understanding

This document outlines the understanding of the GoTune project by the Gemini AI assistant.

## Project Overview

GoTune is a music player application. Based on the file structure, it appears to be a desktop application built with the Fyne toolkit for the graphical user interface and the BASS audio library for music playback.

## Key Dependencies

-   `fyne.io/fyne/v2`: A cross-platform GUI toolkit for Go. This is used for building the user interface.
-   `github.com/dhowden/tag`: A library for reading metadata from audio files (e.g., artist, title, album).
-   `github.com/stretchr/testify`: A testing toolkit for Go.

## Architecture

The project follows a clean architecture pattern, separating concerns into different layers:

-   `cmd/`: The application's entry point.
-   `internal/`: Contains the core application logic.
    -   `adapter/`: Adapters to external services and libraries (e.g., audio engine, UI).
    -   `app/`: The main application logic.
    -   `domain/`: Core domain models, events, and errors.
    -   `infrastructure/`: Configuration and platform-specific code.
    -   `ports/`: Interfaces for the application's core services.
    -   `service/`: Implementations of the services defined in `ports/`.
-   `res/`: application resources like icons and logos
-   `scripts/`: contains scripts for building and setting up dependencies
-   `test/`: contains test data
-   `build/`: contains the build output of the application
-   `libs/`: contains the bass library dependencies

This structure promotes loose coupling and testability.

## UI (Fyne)

The UI is implemented in the `internal/adapter/ui/fyne` directory. Key files include:

-   `main_window.go`: Defines the main application window.
-   `presenter.go`: Handles the presentation logic, updating the UI based on events from the application.
-   `dialogs.go`: Contains dialog windows for user interaction.

## Audio (BASS)

The audio playback functionality is implemented in the `internal/adapter/audio/bass` directory. Key files include:

-   `engine.go`: The core audio engine, responsible for playing, pausing, and stopping music.
-   `bindings.go`: Go bindings for the BASS library.
-   `platform_*.go`: Platform-specific code for loading the BASS library.

## Next Steps

The next step is to investigate the reported threading issue with the Fyne library. I will examine the code in `internal/adapter/ui/fyne` and the services that interact with it to identify any UI updates that are not being performed on the main Fyne thread.
