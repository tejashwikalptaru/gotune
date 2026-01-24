package memory

import (
	"encoding/json"
	"sync"

	"fyne.io/fyne/v2"
	"github.com/tejashwikalptaru/gotune/internal/domain"
	"github.com/tejashwikalptaru/gotune/internal/ports"
)

// PreferencesRepository implements ports.PreferencesRepository using Fyne preferences.
// This provides a thin wrapper around Fyne's preferences system with proper error handling.
//
// Thread-safe: All operations protected by sync.RWMutex.
type PreferencesRepository struct {
	prefs fyne.Preferences
	mu    sync.RWMutex
}

// NewPreferencesRepository creates a new preferences' repository.
// The preferences parameter should be obtained from fyne.CurrentApp().Preferences().
func NewPreferencesRepository(prefs fyne.Preferences) *PreferencesRepository {
	return &PreferencesRepository{
		prefs: prefs,
	}
}

// SaveVolume persists the volume level.
func (r *PreferencesRepository) SaveVolume(volume float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.prefs.SetFloat("preferences.volume", volume)
	return nil
}

// LoadVolume retrieves the saved volume level.
func (r *PreferencesRepository) LoadVolume() (float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	volume := r.prefs.FloatWithFallback("preferences.volume", 1.0)
	return volume, nil
}

// SaveLoopMode persists the loop mode state.
func (r *PreferencesRepository) SaveLoopMode(enabled bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.prefs.SetBool("preferences.loop", enabled)
	return nil
}

// LoadLoopMode retrieves the saved loop mode state.
func (r *PreferencesRepository) LoadLoopMode() (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	loop := r.prefs.BoolWithFallback("preferences.loop", false)
	return loop, nil
}

// SaveTheme persists the theme preference.
func (r *PreferencesRepository) SaveTheme(theme string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.prefs.SetString("preferences.theme", theme)
	return nil
}

// LoadTheme retrieves the saved theme preference.
func (r *PreferencesRepository) LoadTheme() (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	theme := r.prefs.StringWithFallback("preferences.theme", "system")
	return theme, nil
}

// SaveScanPaths persists the list of directories to scan for music.
func (r *PreferencesRepository) SaveScanPaths(paths []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Serialize paths to JSON
	data, err := json.Marshal(paths)
	if err != nil {
		return domain.NewServiceError("PreferencesRepository", "SaveScanPaths", "failed to marshal paths", err)
	}

	r.prefs.SetString("preferences.scan_paths", string(data))
	return nil
}

// LoadScanPaths retrieves the saved scan paths.
func (r *PreferencesRepository) LoadScanPaths() ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data := r.prefs.String("preferences.scan_paths")
	if data == "" {
		// No saved paths - return empty slice
		return []string{}, nil
	}

	// Deserialize
	var paths []string
	if err := json.Unmarshal([]byte(data), &paths); err != nil {
		return nil, domain.NewServiceError("PreferencesRepository", "LoadScanPaths", "failed to unmarshal paths", err)
	}

	return paths, nil
}

// Clear removes all saved preferences.
func (r *PreferencesRepository) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.prefs.RemoveValue("preferences.volume")
	r.prefs.RemoveValue("preferences.loop")
	r.prefs.RemoveValue("preferences.theme")
	r.prefs.RemoveValue("preferences.scan_paths")

	return nil
}

// Verify interface implementation
var _ ports.PreferencesRepository = (*PreferencesRepository)(nil)
