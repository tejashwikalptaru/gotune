// Package memory provides in-memory repository implementations using Fyne preferences.
package memory

import (
	"encoding/json"
	"sync"

	"fyne.io/fyne/v2"
	"github.com/tejashwikalptaru/gotune/internal/domain"
	"github.com/tejashwikalptaru/gotune/internal/ports"
)

// HistoryRepository implements ports.HistoryRepository using Fyne preferences.
//
// Fyne preferences automatically use OS-specific app data directories:
// - macOS: ~/Library/Preferences/com.gotune.app.plist
// - Linux: ~/.config/gotune/
// - Windows: %APPDATA%\gotune\
//
// Thread-safe: All operations protected by sync.RWMutex.
type HistoryRepository struct {
	prefs fyne.Preferences
	mu    sync.RWMutex
}

// NewHistoryRepository creates a new history repository.
// The preferences parameter should be obtained from fyne.CurrentApp().Preferences().
func NewHistoryRepository(prefs fyne.Preferences) *HistoryRepository {
	return &HistoryRepository{
		prefs: prefs,
	}
}

// SaveQueue persists the current playback queue.
func (r *HistoryRepository) SaveQueue(tracks []domain.MusicTrack) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Serialize tracks to JSON
	data, err := json.Marshal(tracks)
	if err != nil {
		return domain.NewServiceError("HistoryRepository", "SaveQueue", "failed to marshal tracks", err)
	}

	// Save to preferences
	r.prefs.SetString("history.queue", string(data))

	return nil
}

// LoadQueue retrieves the last saved playback queue.
func (r *HistoryRepository) LoadQueue() ([]domain.MusicTrack, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Load from preferences
	data := r.prefs.String("history.queue")
	if data == "" {
		// No saved queue - return empty slice
		return []domain.MusicTrack{}, nil
	}

	// Deserialize
	var tracks []domain.MusicTrack
	if err := json.Unmarshal([]byte(data), &tracks); err != nil {
		return nil, domain.NewServiceError("HistoryRepository", "LoadQueue", "failed to unmarshal tracks", err)
	}

	return tracks, nil
}

// SaveCurrentIndex persists the current track index in the queue.
func (r *HistoryRepository) SaveCurrentIndex(index int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store index+1 to distinguish between "not set" (0) and "saved 0" (1)
	// This is because Fyne returns 0 if the key doesn't exist
	r.prefs.SetInt("history.current_index", index+1)
	return nil
}

// LoadCurrentIndex retrieves the last saved track index.
func (r *HistoryRepository) LoadCurrentIndex() (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Load the stored value (which is index+1)
	storedValue := r.prefs.Int("history.current_index")
	if storedValue == 0 {
		// Key doesn't exist - return -1 to indicate "no track selected"
		return -1, nil
	}

	// Subtract 1 to get the actual index
	return storedValue - 1, nil
}

// Clear removes all saved history data.
func (r *HistoryRepository) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.prefs.RemoveValue("history.queue")
	r.prefs.RemoveValue("history.current_index")

	return nil
}

// Verify interface implementation
var _ ports.HistoryRepository = (*HistoryRepository)(nil)
