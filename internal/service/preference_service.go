// Package service provides business logic for the GoTune application.
package service

import (
	"log/slog"
	"sync"

	"github.com/tejashwikalptaru/gotune/internal/domain"
	"github.com/tejashwikalptaru/gotune/internal/ports"
)

// PreferenceService manages application preferences and settings.
// All operations are thread-safe via sync.RWMutex.
type PreferenceService struct {
	// Dependencies (injected)
	logger     *slog.Logger
	repository ports.PreferencesRepository
	bus        ports.EventBus

	// Cached preferences (for performance)
	volume      float64
	loopEnabled bool
	theme       string
	lastFolder  string
	cacheValid  bool

	// Concurrency control
	mu sync.RWMutex
}

// NewPreferenceService creates a new preference service.
func NewPreferenceService(
	logger *slog.Logger,
	repository ports.PreferencesRepository,
	bus ports.EventBus,
) *PreferenceService {
	service := &PreferenceService{
		logger:     logger,
		repository: repository,
		bus:        bus,
		volume:     0.8,    // Default volume
		theme:      "dark", // Default theme
		cacheValid: false,
	}

	logger.Debug("preference service initialized")

	// Load preferences from the repository
	service.loadPreferences()

	return service
}

// loadPreferences loads all preferences from repository into cache.
func (s *PreferenceService) loadPreferences() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Load volume
	if vol, err := s.repository.LoadVolume(); err == nil {
		s.volume = vol
	}

	// Load loop mode
	if loop, err := s.repository.LoadLoopMode(); err == nil {
		s.loopEnabled = loop
	}

	s.cacheValid = true
}

// GetVolume returns the saved volume preference (0.0 to 1.0).
func (s *PreferenceService) GetVolume() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.cacheValid {
		// Try to load from the repository
		if vol, err := s.repository.LoadVolume(); err == nil {
			return vol
		}
	}

	return s.volume
}

// SetVolume saves the volume preference (0.0 to 1.0).
func (s *PreferenceService) SetVolume(volume float64) error {
	if volume < 0.0 || volume > 1.0 {
		return domain.ErrInvalidVolume
	}

	s.mu.Lock()
	s.volume = volume
	s.mu.Unlock()

	// Save to repository
	if err := s.repository.SaveVolume(volume); err != nil {
		return err
	}

	return nil
}

// GetLoopMode returns the saved loop mode preference.
func (s *PreferenceService) GetLoopMode() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.cacheValid {
		// Try to load from the repository
		if loop, err := s.repository.LoadLoopMode(); err == nil {
			return loop
		}
	}

	return s.loopEnabled
}

// SetLoopMode saves the loop mode preference.
func (s *PreferenceService) SetLoopMode(enabled bool) error {
	s.mu.Lock()
	s.loopEnabled = enabled
	s.mu.Unlock()

	// Save to repository
	if err := s.repository.SaveLoopMode(enabled); err != nil {
		return err
	}

	return nil
}

// GetTheme returns the saved theme preference.
func (s *PreferenceService) GetTheme() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.theme
}

// SetTheme saves the theme preference.
func (s *PreferenceService) SetTheme(theme string) error {
	// Validate theme
	validThemes := []string{"light", "dark"}
	isValid := false
	for _, valid := range validThemes {
		if theme == valid {
			isValid = true
			break
		}
	}

	if !isValid {
		return domain.NewValidationError("theme", theme, "must be 'light' or 'dark'")
	}

	s.mu.Lock()
	s.theme = theme
	s.mu.Unlock()

	// Note: Theme saving would require extending the PreferencesRepository interface
	// For now, we just cache it in memory

	return nil
}

// GetLastFolder returns the last opened folder path.
func (s *PreferenceService) GetLastFolder() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.lastFolder
}

// SetLastFolder saves the last opened folder path.
func (s *PreferenceService) SetLastFolder(path string) error {
	s.mu.Lock()
	s.lastFolder = path
	s.mu.Unlock()

	// Note: Last folder saving would require extending the PreferencesRepository interface
	// For now, we just cache it in memory

	return nil
}

// ResetToDefaults resets all preferences to default values.
func (s *PreferenceService) ResetToDefaults() error {
	s.mu.Lock()
	s.volume = 0.8
	s.loopEnabled = false
	s.theme = "dark"
	s.lastFolder = ""
	s.mu.Unlock()

	// Save defaults to repository
	if err := s.repository.SaveVolume(0.8); err != nil {
		return err
	}

	if err := s.repository.SaveLoopMode(false); err != nil {
		return err
	}

	return nil
}

// GetAllPreferences returns all preferences as a map.
func (s *PreferenceService) GetAllPreferences() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"volume":      s.volume,
		"loop":        s.loopEnabled,
		"theme":       s.theme,
		"last_folder": s.lastFolder,
	}
}

// Shutdown cleans up resources.
func (s *PreferenceService) Shutdown() error {
	// No cleanup needed for preference service
	return nil
}

// Verify that PreferenceService implements the expected interface patterns
var _ interface {
	GetVolume() float64
	SetVolume(float64) error
	GetLoopMode() bool
	SetLoopMode(bool) error
	GetTheme() string
	SetTheme(string) error
	GetLastFolder() string
	SetLastFolder(string) error
	ResetToDefaults() error
	GetAllPreferences() map[string]interface{}
	Shutdown() error
} = (*PreferenceService)(nil)
