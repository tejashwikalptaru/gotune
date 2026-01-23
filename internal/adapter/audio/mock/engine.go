// Package mock provides a mock implementation of the AudioEngine interface.
// This is used for testing services without requiring the real BASS library.
package mock

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"github.com/tejashwikalptaru/gotune/internal/domain"
	"github.com/tejashwikalptaru/gotune/internal/ports"
)

// Engine is a mock implementation of the AudioEngine interface.
// It simulates audio playback in memory without actually playing audio.
//
// Thread-safety: This implementation is thread-safe.
type Engine struct {
	// Dependencies
	logger *slog.Logger

	// Configuration
	initialized bool
	device      int
	frequency   int
	flags       int

	// Track state
	tracks     map[domain.TrackHandle]*mockTrack
	nextHandle domain.TrackHandle
	mu         sync.RWMutex

	// Behavior configuration (for testing error scenarios)
	failInitialize bool
	failLoad       bool
	failPlay       bool
}

// mockTrack represents a loaded track in the mock engine.
type mockTrack struct {
	handle   domain.TrackHandle
	filePath string
	duration time.Duration
	position time.Duration
	volume   float64
	status   domain.PlaybackStatus
}

// NewEngine creates a new mock audio engine.
func NewEngine() *Engine {
	return &Engine{
		tracks:     make(map[domain.TrackHandle]*mockTrack),
		nextHandle: 1,
	}
}

// SetLogger sets the logger for this engine.
// This should be called after construction before using the engine.
func (m *Engine) SetLogger(logger *slog.Logger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger = logger
}

// SetFailInitialize configures the mock to fail initialization (for testing).
func (m *Engine) SetFailInitialize(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failInitialize = fail
}

// SetFailLoad configures the mock to fail loading tracks (for testing).
func (m *Engine) SetFailLoad(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failLoad = fail
}

// SetFailPlay configures the mock to fail playback (for testing).
func (m *Engine) SetFailPlay(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failPlay = fail
}

// Initialize initializes the mock audio engine.
func (m *Engine) Initialize(device int, frequency int, flags int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failInitialize {
		return domain.NewAudioEngineError("initialize", "", -1, "mock initialization failed", nil)
	}

	if m.initialized {
		return domain.ErrAlreadyInitialized
	}

	m.initialized = true
	m.device = device
	m.frequency = frequency
	m.flags = flags

	return nil
}

// Shutdown shuts down the mock audio engine.
func (m *Engine) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return domain.ErrNotInitialized
	}

	m.initialized = false
	m.tracks = make(map[domain.TrackHandle]*mockTrack)

	return nil
}

// IsInitialized returns true if the engine is initialized.
func (m *Engine) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.initialized
}

// Load loads an audio file and returns a handle.
func (m *Engine) Load(filePath string) (domain.TrackHandle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return domain.InvalidTrackHandle, domain.ErrNotInitialized
	}

	if m.failLoad {
		return domain.InvalidTrackHandle, domain.NewAudioEngineError("load", filePath, -1, "mock load failed", nil)
	}

	if filePath == "" {
		return domain.InvalidTrackHandle, domain.ErrInvalidFilePath
	}

	// Create a mock track with simulated duration (3 minutes)
	handle := m.nextHandle
	m.nextHandle++

	track := &mockTrack{
		handle:   handle,
		filePath: filePath,
		duration: 3 * time.Minute, // Default duration
		position: 0,
		volume:   1.0, // Full volume
		status:   domain.StatusStopped,
	}

	m.tracks[handle] = track

	return handle, nil
}

// Unload unloads a previously loaded track.
func (m *Engine) Unload(handle domain.TrackHandle) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return domain.ErrNotInitialized
	}

	if _, exists := m.tracks[handle]; !exists {
		return domain.ErrInvalidTrackHandle
	}

	delete(m.tracks, handle)
	return nil
}

// Play starts or resumes playback.
func (m *Engine) Play(handle domain.TrackHandle) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return domain.ErrNotInitialized
	}

	if m.failPlay {
		return domain.ErrPlaybackFailed
	}

	track, exists := m.tracks[handle]
	if !exists {
		return domain.ErrInvalidTrackHandle
	}

	// If stopped, reset position
	if track.status == domain.StatusStopped {
		track.position = 0
	}

	track.status = domain.StatusPlaying
	return nil
}

// Pause pauses playback.
func (m *Engine) Pause(handle domain.TrackHandle) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return domain.ErrNotInitialized
	}

	track, exists := m.tracks[handle]
	if !exists {
		return domain.ErrInvalidTrackHandle
	}

	if track.status == domain.StatusPlaying {
		track.status = domain.StatusPaused
	}

	return nil
}

// Stop stops playback and unloads the track.
func (m *Engine) Stop(handle domain.TrackHandle) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return domain.ErrNotInitialized
	}

	track, exists := m.tracks[handle]
	if !exists {
		return domain.ErrInvalidTrackHandle
	}

	track.status = domain.StatusStopped
	track.position = 0

	// Unload the track
	delete(m.tracks, handle)

	return nil
}

// Status returns the playback status.
func (m *Engine) Status(handle domain.TrackHandle) (domain.PlaybackStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return domain.StatusStopped, domain.ErrNotInitialized
	}

	track, exists := m.tracks[handle]
	if !exists {
		return domain.StatusStopped, domain.ErrInvalidTrackHandle
	}

	return track.status, nil
}

// Position returns the current playback position.
func (m *Engine) Position(handle domain.TrackHandle) (time.Duration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return 0, domain.ErrNotInitialized
	}

	track, exists := m.tracks[handle]
	if !exists {
		return 0, domain.ErrInvalidTrackHandle
	}

	return track.position, nil
}

// Duration returns the total track duration.
func (m *Engine) Duration(handle domain.TrackHandle) (time.Duration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return 0, domain.ErrNotInitialized
	}

	track, exists := m.tracks[handle]
	if !exists {
		return 0, domain.ErrInvalidTrackHandle
	}

	return track.duration, nil
}

// Seek sets the playback position.
func (m *Engine) Seek(handle domain.TrackHandle, position time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return domain.ErrNotInitialized
	}

	track, exists := m.tracks[handle]
	if !exists {
		return domain.ErrInvalidTrackHandle
	}

	if position < 0 || position > track.duration {
		return domain.ErrInvalidPosition
	}

	track.position = position
	return nil
}

// SetVolume sets the playback volume.
func (m *Engine) SetVolume(handle domain.TrackHandle, volume float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return domain.ErrNotInitialized
	}

	track, exists := m.tracks[handle]
	if !exists {
		return domain.ErrInvalidTrackHandle
	}

	if volume < 0.0 || volume > 1.0 {
		return domain.ErrInvalidVolume
	}

	track.volume = volume
	return nil
}

// GetVolume returns the current volume.
func (m *Engine) GetVolume(handle domain.TrackHandle) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return 0, domain.ErrNotInitialized
	}

	track, exists := m.tracks[handle]
	if !exists {
		return 0, domain.ErrInvalidTrackHandle
	}

	return track.volume, nil
}

// GetMetadata extracts mock metadata from a file path.
func (m *Engine) GetMetadata(filePath string) (*domain.MusicTrack, error) {
	if filePath == "" {
		return nil, domain.ErrInvalidFilePath
	}

	// Extract filename for mock metadata
	filename := filepath.Base(filePath)
	ext := filepath.Ext(filename)
	nameWithoutExt := filename[:len(filename)-len(ext)]

	// Create mock metadata
	track := &domain.MusicTrack{
		ID:         fmt.Sprintf("mock-%s", nameWithoutExt),
		FilePath:   filePath,
		Title:      nameWithoutExt,
		Artist:     "Mock Artist",
		Album:      "Mock Album",
		Duration:   3 * time.Minute,
		FileFormat: ext,
		IsMOD:      isMODFormat(ext),
		Metadata: &domain.TrackMetadata{
			Composer:   "Mock Composer",
			Genre:      "Mock Genre",
			Year:       2024,
			BitRate:    320,
			SampleRate: 44100,
		},
	}

	return track, nil
}

// isMODFormat checks if the file extension is a MOD format.
func isMODFormat(ext string) bool {
	modFormats := []string{".mod", ".xm", ".it", ".s3m", ".mtm", ".umx"}
	for _, format := range modFormats {
		if ext == format {
			return true
		}
	}
	return false
}

// GetLoadedTracks returns the number of currently loaded tracks (for testing).
func (m *Engine) GetLoadedTracks() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.tracks)
}

// SimulateProgress simulates playback progress (for testing).
// This advances the position by the specified duration.
func (m *Engine) SimulateProgress(handle domain.TrackHandle, delta time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	track, exists := m.tracks[handle]
	if !exists {
		return domain.ErrInvalidTrackHandle
	}

	if track.status != domain.StatusPlaying {
		return fmt.Errorf("track is not playing")
	}

	track.position += delta
	if track.position > track.duration {
		track.position = track.duration
		track.status = domain.StatusStopped
	}

	return nil
}

// GetFFTData returns mock FFT data for visualization.
// In the mock engine, this returns a simple simulated waveform.
func (m *Engine) GetFFTData(handle domain.TrackHandle) ([]float32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, domain.ErrNotInitialized
	}

	track, exists := m.tracks[handle]
	if !exists {
		return nil, domain.ErrInvalidTrackHandle
	}

	// Only return data if playing
	if track.status != domain.StatusPlaying {
		return nil, domain.ErrFFTDataUnavailable
	}

	// Return mock FFT data (1024 values simulating frequency data)
	data := make([]float32, 1024)
	for i := range data {
		// Simulate decreasing intensity at higher frequencies
		data[i] = float32(0.5) * (1.0 - float32(i)/1024.0)
	}

	return data, nil
}

// Verify that Engine implements the AudioEngine interface
var _ ports.AudioEngine = (*Engine)(nil)
