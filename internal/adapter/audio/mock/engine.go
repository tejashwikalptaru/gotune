// Package mock provides a mock implementation of the AudioEngine interface.
// This is used for testing services without requiring the real BASS library.
package mock

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/tejashwikalptaru/gotune/internal/domain"
	"github.com/tejashwikalptaru/gotune/internal/ports"
)

// MockEngine is a mock implementation of the AudioEngine interface.
// It simulates audio playback in memory without actually playing audio.
//
// Thread-safety: This implementation is thread-safe.
type MockEngine struct {
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

// NewMockEngine creates a new mock audio engine.
func NewMockEngine() *MockEngine {
	return &MockEngine{
		tracks:     make(map[domain.TrackHandle]*mockTrack),
		nextHandle: 1,
	}
}

// SetFailInitialize configures the mock to fail initialization (for testing).
func (m *MockEngine) SetFailInitialize(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failInitialize = fail
}

// SetFailLoad configures the mock to fail loading tracks (for testing).
func (m *MockEngine) SetFailLoad(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failLoad = fail
}

// SetFailPlay configures the mock to fail playback (for testing).
func (m *MockEngine) SetFailPlay(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failPlay = fail
}

// Initialize initializes the mock audio engine.
func (m *MockEngine) Initialize(device int, frequency int, flags int) error {
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
func (m *MockEngine) Shutdown() error {
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
func (m *MockEngine) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.initialized
}

// Load loads an audio file and returns a handle.
func (m *MockEngine) Load(filePath string) (domain.TrackHandle, error) {
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
func (m *MockEngine) Unload(handle domain.TrackHandle) error {
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
func (m *MockEngine) Play(handle domain.TrackHandle) error {
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
func (m *MockEngine) Pause(handle domain.TrackHandle) error {
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
func (m *MockEngine) Stop(handle domain.TrackHandle) error {
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
func (m *MockEngine) Status(handle domain.TrackHandle) (domain.PlaybackStatus, error) {
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
func (m *MockEngine) Position(handle domain.TrackHandle) (time.Duration, error) {
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
func (m *MockEngine) Duration(handle domain.TrackHandle) (time.Duration, error) {
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
func (m *MockEngine) Seek(handle domain.TrackHandle, position time.Duration) error {
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
func (m *MockEngine) SetVolume(handle domain.TrackHandle, volume float64) error {
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
func (m *MockEngine) GetVolume(handle domain.TrackHandle) (float64, error) {
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
func (m *MockEngine) GetMetadata(filePath string) (*domain.MusicTrack, error) {
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
func (m *MockEngine) GetLoadedTracks() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.tracks)
}

// SimulateProgress simulates playback progress (for testing).
// This advances the position by the specified duration.
func (m *MockEngine) SimulateProgress(handle domain.TrackHandle, delta time.Duration) error {
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

// Verify that MockEngine implements the AudioEngine interface
var _ ports.AudioEngine = (*MockEngine)(nil)
