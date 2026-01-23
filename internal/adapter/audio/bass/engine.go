// Package bass provides a BASS audio library adapter implementing the AudioEngine interface.
package bass

import (
	"log/slog"
	"sync"
	"time"

	"github.com/tejashwikalptaru/gotune/internal/domain"
	"github.com/tejashwikalptaru/gotune/internal/ports"
)

// Engine is the BASS library implementation of the AudioEngine interface.
// It wraps the Un4seen BASS library for audio playback with thread-safe operations.
//
// Thread-safety: This implementation is thread-safe via sync.RWMutex.
type Engine struct {
	// Dependencies
	logger *slog.Logger

	// Configuration
	initialized bool
	device      int
	frequency   int
	flags       int

	// Track management
	tracks map[domain.TrackHandle]*trackInfo
	mu     sync.RWMutex
}

// trackInfo stores information about a loaded track.
type trackInfo struct {
	handle   int64 // BASS channel handle
	filePath string
	isMOD    bool // True if this is a MOD/tracker file
}

// NewEngine creates a new BASS audio engine.
func NewEngine() *Engine {
	return &Engine{
		tracks: make(map[domain.TrackHandle]*trackInfo),
	}
}

// SetLogger sets the logger for this engine.
// This should be called after construction before using the engine.
func (e *Engine) SetLogger(logger *slog.Logger) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.logger = logger
}

// Initialize sets up the BASS audio engine.
func (e *Engine) Initialize(device int, frequency int, flags int) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized {
		return domain.ErrAlreadyInitialized
	}

	err := bassInit(device, frequency, flags)
	if err != nil {
		return err
	}

	e.initialized = true
	e.device = device
	e.frequency = frequency
	e.flags = flags

	return nil
}

// Shutdown releases all BASS engine resources.
func (e *Engine) Shutdown() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.initialized {
		return domain.ErrNotInitialized
	}

	// Stop and free all loaded tracks
	for handle := range e.tracks {
		if err := e.unloadInternal(handle); err != nil {
			if e.logger != nil {
				e.logger.Error("error unloading track during shutdown",
					slog.Int64("handle", int64(handle)),
					slog.Any("error", err))
			}
		}
	}

	err := bassFree()
	if err != nil {
		return err
	}

	e.initialized = false
	e.tracks = make(map[domain.TrackHandle]*trackInfo)

	return nil
}

// IsInitialized returns true if the engine is initialized.
func (e *Engine) IsInitialized() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.initialized
}

// Load loads an audio file and returns a handle.
func (e *Engine) Load(filePath string) (domain.TrackHandle, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.initialized {
		return domain.InvalidTrackHandle, domain.ErrNotInitialized
	}

	if filePath == "" {
		return domain.InvalidTrackHandle, domain.ErrInvalidFilePath
	}

	// Determine if this is a MOD file
	isMOD := isModFile(filePath)

	var bassHandle int64
	var err error

	if isMOD {
		// Load as MOD music
		bassHandle, err = bassMusicLoad(filePath, musicPreScan|musicRamps|streamAutoFree|posReset|posResetEx)
	} else {
		// Load as regular stream
		bassHandle, err = bassStreamCreateFile(filePath, streamAutoFree|posReset|posResetEx)
	}

	if err != nil {
		// Try the opposite method as fallback
		if isMOD {
			bassHandle, err = bassStreamCreateFile(filePath, streamAutoFree|posReset|posResetEx)
			isMOD = false
		} else {
			bassHandle, err = bassMusicLoad(filePath, musicPreScan|musicRamps|streamAutoFree|posReset|posResetEx)
			isMOD = true
		}

		if err != nil {
			return domain.InvalidTrackHandle, err
		}
	}

	// Create a track handle (use bassHandle as the domain handle)
	handle := domain.TrackHandle(bassHandle)

	// Store track info
	e.tracks[handle] = &trackInfo{
		handle:   bassHandle,
		filePath: filePath,
		isMOD:    isMOD,
	}

	return handle, nil
}

// Unload releases resources for a loaded track.
func (e *Engine) Unload(handle domain.TrackHandle) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.initialized {
		return domain.ErrNotInitialized
	}

	return e.unloadInternal(handle)
}

// unloadInternal unloads a track without locking (caller must hold lock).
func (e *Engine) unloadInternal(handle domain.TrackHandle) error {
	track, exists := e.tracks[handle]
	if !exists {
		return domain.ErrInvalidTrackHandle
	}

	// Stop the channel first
	if err := bassChannelStop(track.handle); err != nil {
		return err
	}

	// Free the channel
	if track.isMOD {
		bassMusicFree(track.handle)
	} else {
		bassStreamFree(track.handle)
	}

	// Remove from the map
	delete(e.tracks, handle)

	return nil
}

// Play starts or resumes playback.
func (e *Engine) Play(handle domain.TrackHandle) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.logger != nil {
		e.logger.Debug("play called", slog.Int64("handle", int64(handle)))
	}

	if !e.initialized {
		if e.logger != nil {
			e.logger.Debug("engine not initialized")
		}
		return domain.ErrNotInitialized
	}

	track, exists := e.tracks[handle]
	if !exists {
		if e.logger != nil {
			e.logger.Debug("track handle not found", slog.Int64("handle", int64(handle)))
		}
		return domain.ErrInvalidTrackHandle
	}

	if e.logger != nil {
		e.logger.Debug("track info",
			slog.Int64("bass_handle", track.handle),
			slog.String("file_path", track.filePath))
	}

	status := bassChannelIsActive(track.handle)
	if e.logger != nil {
		e.logger.Debug("channel status before play", slog.Any("status", status))
	}

	// If stopped, restart from the beginning
	restart := status == domain.StatusStopped || status == domain.StatusStalled
	if e.logger != nil {
		e.logger.Debug("calling bassChannelPlay", slog.Bool("restart", restart))
	}

	err := bassChannelPlay(track.handle, restart)
	if err != nil {
		if e.logger != nil {
			e.logger.Debug("bassChannelPlay failed", slog.Any("error", err))
		}
	} else {
		if e.logger != nil {
			e.logger.Debug("bassChannelPlay succeeded")
			// Check status after play
			newStatus := bassChannelIsActive(track.handle)
			e.logger.Debug("channel status after play", slog.Any("status", newStatus))
		}
	}

	return err
}

// Pause pauses playback.
func (e *Engine) Pause(handle domain.TrackHandle) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.initialized {
		return domain.ErrNotInitialized
	}

	track, exists := e.tracks[handle]
	if !exists {
		return domain.ErrInvalidTrackHandle
	}

	return bassChannelPause(track.handle)
}

// Stop stops playback and unloads the track.
func (e *Engine) Stop(handle domain.TrackHandle) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.initialized {
		return domain.ErrNotInitialized
	}

	track, exists := e.tracks[handle]
	if !exists {
		return domain.ErrInvalidTrackHandle
	}

	// Fade out effects (smooth stop)
	bassChannelSlideAttribute(track.handle, ChannelAttribFREQ, 1000, 500)
	bassChannelSlideAttribute(track.handle, ChannelAttribVOL|ChannelAttribSLIDELOG, -1, 100)

	// Stop the channel
	err := bassChannelStop(track.handle)
	if err != nil {
		return err
	}

	// Free the channel
	if track.isMOD {
		bassMusicFree(track.handle)
	} else {
		bassStreamFree(track.handle)
	}

	// Remove from tracks
	delete(e.tracks, handle)

	return nil
}

// Status returns the playback status.
func (e *Engine) Status(handle domain.TrackHandle) (domain.PlaybackStatus, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.initialized {
		return domain.StatusStopped, domain.ErrNotInitialized
	}

	track, exists := e.tracks[handle]
	if !exists {
		return domain.StatusStopped, domain.ErrInvalidTrackHandle
	}

	status := bassChannelIsActive(track.handle)
	return status, nil
}

// Position returns the current playback position.
func (e *Engine) Position(handle domain.TrackHandle) (time.Duration, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.initialized {
		return 0, domain.ErrNotInitialized
	}

	track, exists := e.tracks[handle]
	if !exists {
		return 0, domain.ErrInvalidTrackHandle
	}

	posBytes := bassChannelGetPosition(track.handle)
	duration := bassChannelBytes2Seconds(track.handle, posBytes)

	return duration, nil
}

// Duration returns the total track duration.
func (e *Engine) Duration(handle domain.TrackHandle) (time.Duration, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.initialized {
		return 0, domain.ErrNotInitialized
	}

	track, exists := e.tracks[handle]
	if !exists {
		return 0, domain.ErrInvalidTrackHandle
	}

	lengthBytes := bassChannelGetLength(track.handle)
	duration := bassChannelBytes2Seconds(track.handle, lengthBytes)

	return duration, nil
}

// Seek sets the playback position.
func (e *Engine) Seek(handle domain.TrackHandle, position time.Duration) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.initialized {
		return domain.ErrNotInitialized
	}

	track, exists := e.tracks[handle]
	if !exists {
		return domain.ErrInvalidTrackHandle
	}

	// Get duration to validate position
	lengthBytes := bassChannelGetLength(track.handle)
	duration := bassChannelBytes2Seconds(track.handle, lengthBytes)

	if position < 0 || position > duration {
		return domain.ErrInvalidPosition
	}

	// Convert position to bytes
	posBytes := bassChannelSeconds2Bytes(track.handle, position)

	return bassChannelSetPosition(track.handle, posBytes)
}

// SetVolume sets the playback volume (0.0 to 1.0).
func (e *Engine) SetVolume(handle domain.TrackHandle, volume float64) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.initialized {
		return domain.ErrNotInitialized
	}

	track, exists := e.tracks[handle]
	if !exists {
		return domain.ErrInvalidTrackHandle
	}

	if volume < 0.0 || volume > 1.0 {
		return domain.ErrInvalidVolume
	}

	// BASS uses 0.0 to 1.0 for volume, so no conversion needed
	return bassChannelSetAttribute(track.handle, ChannelAttribVOL, float32(volume))
}

// GetVolume returns the current volume (0.0 to 1.0).
func (e *Engine) GetVolume(handle domain.TrackHandle) (float64, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.initialized {
		return 0, domain.ErrNotInitialized
	}

	track, exists := e.tracks[handle]
	if !exists {
		return 0, domain.ErrInvalidTrackHandle
	}

	volume, err := bassChannelGetAttribute(track.handle, ChannelAttribVOL)
	if err != nil {
		return 0, err
	}

	return float64(volume), nil
}

// GetMetadata extracts metadata from an audio file without loading it for playback.
func (e *Engine) GetMetadata(filePath string) (*domain.MusicTrack, error) {
	// Metadata extraction is handled by the metadata.go file
	// This is a separate concern from playback
	return extractMetadata(filePath)
}

// GetLoadedTracksCount returns the number of currently loaded tracks (for debugging).
func (e *Engine) GetLoadedTracksCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.tracks)
}

// GetFFTData retrieves FFT frequency data for visualization.
// Uses BASS_DATA_FFT2048 for good resolution, returns 1024 float values.
// The returned values represent frequency magnitudes from low to high frequencies.
func (e *Engine) GetFFTData(handle domain.TrackHandle) ([]float32, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.initialized {
		return nil, domain.ErrNotInitialized
	}

	track, exists := e.tracks[handle]
	if !exists {
		return nil, domain.ErrInvalidTrackHandle
	}

	// FFT2048 returns 1024 float values (half of FFT size)
	buffer := make([]float32, 1024)
	result := bassChannelGetData(track.handle, buffer, dataFFT2048)
	if result == -1 {
		return nil, domain.ErrFFTDataUnavailable
	}

	return buffer, nil
}

// Verify that Engine implements the AudioEngine interface
var _ ports.AudioEngine = (*Engine)(nil)
