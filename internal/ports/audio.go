// Package ports define interfaces for dependency inversion.
// These interfaces allow the core business logic to remain independent of external frameworks.
package ports

import (
	"time"

	"github.com/tejashwikalptaru/gotune/internal/domain"
)

// AudioEngine is the interface for audio playback engines.
// This abstracts the underlying audio library (BASS) and allows for testing with mocks.
//
// Implementations must be thread-safe as they may be called from multiple goroutines.
type AudioEngine interface {
	// Lifecycle methods

	// Initialize sets up the audio engine with the specified configuration.
	// device: Audio device index (-1 for default)
	// frequency: Sample rate in Hz (e.g., 44100 for CD quality)
	// flags: Engine-specific initialization flags
	//
	// Return an error if initialization fails.
	Initialize(device int, frequency int, flags int) error

	// Shutdown releases all audio engine resources.
	// Should be called when the engine is no longer needed.
	//
	// Returns an error if shutdown fails.
	Shutdown() error

	// IsInitialized returns true if the engine has been successfully initialized.
	IsInitialized() bool

	// Track loading methods

	// Load loads an audio file and returns a handle to it.
	// The file remains loaded until Stop is called with the handle.
	//
	// filePath: Absolute path to the audio file
	//
	// Returns a TrackHandle for the loaded track, or an error if loading fails.
	Load(filePath string) (domain.TrackHandle, error)

	// Unload releases resources for a previously loaded track.
	// This is called automatically by Stop, but can be called explicitly if needed.
	//
	// Returns an error if the handle is invalid or unloading fails.
	Unload(handle domain.TrackHandle) error

	// Playback control methods

	// Play starts or resumes playback of the specified track.
	// If the track is stopped, playback starts from the beginning.
	// If the track is paused, playback resumes from the paused position.
	//
	// Returns an error if playback cannot be started.
	Play(handle domain.TrackHandle) error

	// Pause pauses playback of the specified track.
	// The playback position is preserved and can be resumed with Play.
	//
	// Returns an error if the handle is invalid or pausing fails.
	Pause(handle domain.TrackHandle) error

	// Stop stops playback of the specified track and unloads it.
	// The track must be reloaded with Load before it can be played again.
	//
	// Returns an error if the handle is invalid or stopping fails.
	Stop(handle domain.TrackHandle) error

	// State query methods

	// Status returns the current playback status of the specified track.
	//
	// Returns the status or an error if the handle is invalid.
	Status(handle domain.TrackHandle) (domain.PlaybackStatus, error)

	// Position returns the current playback position within the track.
	//
	// Returns the position as a duration or an error if the handle is invalid.
	Position(handle domain.TrackHandle) (time.Duration, error)

	// Duration returns the total duration of the specified track.
	//
	// Returns the duration or an error if the handle is invalid.
	Duration(handle domain.TrackHandle) (time.Duration, error)

	// Seeking methods

	// Seek sets the playback position to the specified time.
	// The position must be within the valid range [0, Duration].
	//
	// Returns an error if the position is invalid or seeking fails.
	Seek(handle domain.TrackHandle, position time.Duration) error

	// Volume control methods

	// SetVolume sets the playback volume for the specified track.
	// volume: Volume level from 0.0 (silent) to 1.0 (full volume)
	//
	// Returns an error if the volume is out of range or the handle is invalid.
	SetVolume(handle domain.TrackHandle, volume float64) error

	// GetVolume returns the current volume level for the specified track.
	//
	// Returns the volume (0.0-1.0), or an error if the handle is invalid.
	GetVolume(handle domain.TrackHandle) (float64, error)

	// Metadata methods

	// GetMetadata extracts metadata from an audio file without loading it for playback.
	// This is used for library scanning to quickly extract track information.
	//
	// filePath: Absolute path to the audio file
	//
	// Returns a MusicTrack with populated metadata, or an error if extraction fails.
	GetMetadata(filePath string) (*domain.MusicTrack, error)

	// Visualization methods

	// GetFFTData retrieves FFT frequency data for visualization.
	// Uses FFT2048 for good resolution, returns 1024 float values representing
	// frequency magnitudes from low to high frequencies.
	//
	// Returns the FFT data or an error if the data cannot be retrieved.
	GetFFTData(handle domain.TrackHandle) ([]float32, error)
}

// AudioEngineFactory is a function that creates an AudioEngine instance.
// This allows for dependency injection of different engine implementations.
type AudioEngineFactory func(config *AudioEngineConfig) (AudioEngine, error)

// AudioEngineConfig contains configuration for creating an audio engine.
type AudioEngineConfig struct {
	// LibraryPath is the path to the native audio library (e.g., libbass.dylib)
	LibraryPath string

	// LibraryName is the name of the library file
	LibraryName string

	// Device is the audio device index (-1 for default)
	Device int

	// Frequency is the sample rate in Hz
	Frequency int

	// Flags are engine-specific initialization flags
	Flags int
}
