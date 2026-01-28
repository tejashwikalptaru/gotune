package service

import (
	"io"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tejashwikalptaru/gotune/internal/adapter/eventbus"
	"github.com/tejashwikalptaru/gotune/internal/domain"
)

// Mock preferences repository for testing
type mockPreferencesRepository struct {
	mu        sync.RWMutex
	volume    float64
	loop      bool
	theme     string
	scanPaths []string
}

func newMockPreferencesRepository() *mockPreferencesRepository {
	return &mockPreferencesRepository{
		volume: 0.8, // Default
		loop:   false,
	}
}

func (m *mockPreferencesRepository) SaveVolume(volume float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.volume = volume
	return nil
}

func (m *mockPreferencesRepository) LoadVolume() (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.volume, nil
}

func (m *mockPreferencesRepository) SaveLoopMode(enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loop = enabled
	return nil
}

func (m *mockPreferencesRepository) LoadLoopMode() (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.loop, nil
}

func (m *mockPreferencesRepository) SaveTheme(theme string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.theme = theme
	return nil
}

func (m *mockPreferencesRepository) LoadTheme() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.theme == "" {
		return "system", nil
	}
	return m.theme, nil
}

func (m *mockPreferencesRepository) SaveScanPaths(paths []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.scanPaths = paths
	return nil
}

func (m *mockPreferencesRepository) LoadScanPaths() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.scanPaths == nil {
		return []string{}, nil
	}
	return m.scanPaths, nil
}

func (m *mockPreferencesRepository) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.volume = 0.8
	m.loop = false
	m.theme = ""
	m.scanPaths = nil
	return nil
}

// prefTestLogger returns a logger that discards output for tests
func prefTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Helper to create a test preference service
func newTestPreferenceService() (*PreferenceService, *mockPreferencesRepository) {
	repo := newMockPreferencesRepository()
	bus := eventbus.NewSyncEventBus()
	service := NewPreferenceService(prefTestLogger(), repo, bus)

	return service, repo
}

func TestPreferenceService_GetVolume_Default(t *testing.T) {
	service, _ := newTestPreferenceService()
	defer service.Shutdown()

	// Should return default volume
	volume := service.GetVolume()
	assert.Equal(t, 0.8, volume)
}

func TestPreferenceService_SetVolume(t *testing.T) {
	service, repo := newTestPreferenceService()
	defer service.Shutdown()

	// Set volume
	err := service.SetVolume(0.6)
	require.NoError(t, err)

	// Verify cached value
	assert.Equal(t, 0.6, service.GetVolume())

	// Verify persisted value
	savedVolume, _ := repo.LoadVolume()
	assert.Equal(t, 0.6, savedVolume)
}

func TestPreferenceService_SetVolume_InvalidRange(t *testing.T) {
	service, _ := newTestPreferenceService()
	defer service.Shutdown()

	// Volume below 0
	err := service.SetVolume(-0.1)
	assert.Equal(t, domain.ErrInvalidVolume, err)

	// Volume above 1
	err = service.SetVolume(1.5)
	assert.Equal(t, domain.ErrInvalidVolume, err)

	// Volume should not have changed
	assert.Equal(t, 0.8, service.GetVolume())
}

func TestPreferenceService_SetVolume_BoundaryValues(t *testing.T) {
	service, _ := newTestPreferenceService()
	defer service.Shutdown()

	// Minimum volume
	err := service.SetVolume(0.0)
	require.NoError(t, err)
	assert.Equal(t, 0.0, service.GetVolume())

	// Maximum volume
	err = service.SetVolume(1.0)
	require.NoError(t, err)
	assert.Equal(t, 1.0, service.GetVolume())
}

func TestPreferenceService_GetLoopMode_Default(t *testing.T) {
	service, _ := newTestPreferenceService()
	defer service.Shutdown()

	// Should return default (false)
	assert.False(t, service.GetLoopMode())
}

func TestPreferenceService_SetLoopMode(t *testing.T) {
	service, repo := newTestPreferenceService()
	defer service.Shutdown()

	// Enable loop
	err := service.SetLoopMode(true)
	require.NoError(t, err)

	// Verify cached value
	assert.True(t, service.GetLoopMode())

	// Verify persisted value
	savedLoop, _ := repo.LoadLoopMode()
	assert.True(t, savedLoop)

	// Disable loop
	err = service.SetLoopMode(false)
	require.NoError(t, err)

	assert.False(t, service.GetLoopMode())
}

func TestPreferenceService_GetTheme_Default(t *testing.T) {
	service, _ := newTestPreferenceService()
	defer service.Shutdown()

	// Should return the default theme
	assert.Equal(t, "dark", service.GetTheme())
}

func TestPreferenceService_SetTheme(t *testing.T) {
	service, _ := newTestPreferenceService()
	defer service.Shutdown()

	// Set a light theme
	err := service.SetTheme("light")
	require.NoError(t, err)
	assert.Equal(t, "light", service.GetTheme())

	// Set a dark theme
	err = service.SetTheme("dark")
	require.NoError(t, err)
	assert.Equal(t, "dark", service.GetTheme())
}

func TestPreferenceService_SetTheme_Invalid(t *testing.T) {
	service, _ := newTestPreferenceService()
	defer service.Shutdown()

	// Try to set an invalid theme
	err := service.SetTheme("blue")
	assert.Error(t, err)

	// Theme should not have changed
	assert.Equal(t, "dark", service.GetTheme())
}

func TestPreferenceService_GetLastFolder_Default(t *testing.T) {
	service, _ := newTestPreferenceService()
	defer service.Shutdown()

	// Should return empty string by default
	assert.Equal(t, "", service.GetLastFolder())
}

func TestPreferenceService_SetLastFolder(t *testing.T) {
	service, _ := newTestPreferenceService()
	defer service.Shutdown()

	// Set the last folder
	err := service.SetLastFolder("/home/user/Music")
	require.NoError(t, err)
	assert.Equal(t, "/home/user/Music", service.GetLastFolder())
}

func TestPreferenceService_ResetToDefaults(t *testing.T) {
	service, repo := newTestPreferenceService()
	defer service.Shutdown()

	// Change all preferences
	service.SetVolume(0.5)
	service.SetLoopMode(true)
	service.SetTheme("light")
	service.SetLastFolder("/some/path")

	// Reset to defaults
	err := service.ResetToDefaults()
	require.NoError(t, err)

	// Verify defaults
	assert.Equal(t, 0.8, service.GetVolume())
	assert.False(t, service.GetLoopMode())
	assert.Equal(t, "dark", service.GetTheme())
	assert.Equal(t, "", service.GetLastFolder())

	// Verify persisted defaults
	savedVolume, _ := repo.LoadVolume()
	assert.Equal(t, 0.8, savedVolume)

	savedLoop, _ := repo.LoadLoopMode()
	assert.False(t, savedLoop)
}

func TestPreferenceService_GetAllPreferences(t *testing.T) {
	service, _ := newTestPreferenceService()
	defer service.Shutdown()

	// Set some preferences
	service.SetVolume(0.7)
	service.SetLoopMode(true)
	service.SetTheme("light")
	service.SetLastFolder("/music")

	// Get all preferences
	prefs := service.GetAllPreferences()

	// Verify all values
	assert.Equal(t, 0.7, prefs["volume"])
	assert.Equal(t, true, prefs["loop"])
	assert.Equal(t, "light", prefs["theme"])
	assert.Equal(t, "/music", prefs["last_folder"])
}

func TestPreferenceService_Persistence(t *testing.T) {
	repo := newMockPreferencesRepository()
	bus := eventbus.NewSyncEventBus()
	testLogger := prefTestLogger()

	// First service instance
	service1 := NewPreferenceService(testLogger, repo, bus)
	service1.SetVolume(0.6)
	service1.SetLoopMode(true)
	service1.Shutdown()

	// Second service instance with the same repository
	service2 := NewPreferenceService(testLogger, repo, bus)
	defer service2.Shutdown()

	// Should load saved preferences
	assert.Equal(t, 0.6, service2.GetVolume())
	assert.True(t, service2.GetLoopMode())
}

// Thread safety tests

func TestPreferenceService_ConcurrentVolumeChanges(t *testing.T) {
	service, _ := newTestPreferenceService()
	defer service.Shutdown()

	// Change volume concurrently
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func(vol float64) {
			service.SetVolume(vol)
			done <- struct{}{}
		}(float64(i) / 10.0)
	}

	// Wait for all
	for i := 0; i < 10; i++ {
		<-done
	}

	// Volume should be one of the set values
	volume := service.GetVolume()
	assert.GreaterOrEqual(t, volume, 0.0)
	assert.LessOrEqual(t, volume, 1.0)
}

func TestPreferenceService_ConcurrentReads(t *testing.T) {
	service, _ := newTestPreferenceService()
	defer service.Shutdown()

	service.SetVolume(0.75)

	// Read concurrently
	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func() {
			vol := service.GetVolume()
			assert.Equal(t, 0.75, vol)
			done <- struct{}{}
		}()
	}

	// Wait for all
	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestPreferenceService_ConcurrentMixedOperations(t *testing.T) {
	service, _ := newTestPreferenceService()
	defer service.Shutdown()

	// Mix reads and writes
	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func(index int) {
			if index%2 == 0 {
				service.SetVolume(0.5)
			} else {
				_ = service.GetVolume()
			}
			done <- struct{}{}
		}(i)
	}

	// Wait for all
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should have a valid volume
	volume := service.GetVolume()
	assert.GreaterOrEqual(t, volume, 0.0)
	assert.LessOrEqual(t, volume, 1.0)
}

func TestPreferenceService_Shutdown(t *testing.T) {
	service, _ := newTestPreferenceService()

	// Set some preferences
	service.SetVolume(0.9)
	service.SetLoopMode(true)

	// Shutdown
	err := service.Shutdown()
	assert.NoError(t, err)

	// Can still read (service doesn't clear on shutdown)
	assert.Equal(t, 0.9, service.GetVolume())
}
