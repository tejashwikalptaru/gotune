package memory

import (
	"encoding/json"
	"log/slog"
	"sync"

	"fyne.io/fyne/v2"
	"github.com/tejashwikalptaru/gotune/internal/domain"
	"github.com/tejashwikalptaru/gotune/internal/ports"
)

// PlaylistRepository implements ports.PlaylistRepository using Fyne preferences.
// Playlists are stored as JSON in preferences with keys like "playlist.<id>".
//
// Thread-safe: All operations protected by sync.RWMutex.
type PlaylistRepository struct {
	prefs  fyne.Preferences
	mu     sync.RWMutex
	logger *slog.Logger
}

// NewPlaylistRepository creates a new playlist repository.
// The preferences parameter should be obtained from fyne.CurrentApp().Preferences().
func NewPlaylistRepository(prefs fyne.Preferences, logger *slog.Logger) *PlaylistRepository {
	return &PlaylistRepository{
		prefs:  prefs,
		logger: logger,
	}
}

// Save persists a playlist.
func (r *PlaylistRepository) Save(playlist *domain.Playlist) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Serialize playlist to JSON
	data, err := json.Marshal(playlist)
	if err != nil {
		return domain.NewServiceError("PlaylistRepository", "Save", "failed to marshal playlist", err)
	}

	// Save to preferences with the key "playlist.<id>"
	key := "playlist." + playlist.ID
	r.prefs.SetString(key, string(data))

	// Update the list of playlist IDs
	ids, err := r.loadPlaylistIDs()
	if err != nil {
		// If loading fails, start with empty slice
		ids = []string{}
	}
	if !contains(ids, playlist.ID) {
		ids = append(ids, playlist.ID)
		if err := r.savePlaylistIDs(ids); err != nil {
			return err
		}
	}

	return nil
}

// Load retrieves a playlist by ID.
func (r *PlaylistRepository) Load(id string) (*domain.Playlist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Load from preferences
	key := "playlist." + id
	data := r.prefs.String(key)
	if data == "" {
		return nil, ports.ErrPlaylistNotFound
	}

	// Deserialize
	var playlist domain.Playlist
	if err := json.Unmarshal([]byte(data), &playlist); err != nil {
		return nil, domain.NewServiceError("PlaylistRepository", "Load", "failed to unmarshal playlist", err)
	}

	return &playlist, nil
}

// LoadAll retrieves all saved playlists.
func (r *PlaylistRepository) LoadAll() ([]*domain.Playlist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Get a list of playlist IDs
	ids, err := r.loadPlaylistIDs()
	if err != nil {
		return nil, err
	}

	// Load each playlist
	playlists := make([]*domain.Playlist, 0, len(ids))
	for _, id := range ids {
		key := "playlist." + id
		data := r.prefs.String(key)
		if data == "" {
			r.logger.Warn("playlist data missing", slog.String("id", id))
			continue // Skip missing playlists
		}

		var playlist domain.Playlist
		if err := json.Unmarshal([]byte(data), &playlist); err != nil {
			r.logger.Warn("playlist corrupted", slog.String("id", id), slog.Any("error", err))
			continue // Skip corrupted playlists
		}

		playlists = append(playlists, &playlist)
	}

	return playlists, nil
}

// Delete removes a playlist by ID.
func (r *PlaylistRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove from preferences
	key := "playlist." + id
	r.prefs.RemoveValue(key)

	// Update the list of playlist IDs
	ids, err := r.loadPlaylistIDs()
	if err != nil {
		// If loading fails, start with empty slice
		ids = []string{}
	}
	ids = remove(ids, id)
	if err := r.savePlaylistIDs(ids); err != nil {
		return err
	}

	return nil
}

// Exists checks if a playlist with the given ID exists.
func (r *PlaylistRepository) Exists(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := "playlist." + id
	data := r.prefs.String(key)
	return data != ""
}

// loadPlaylistIDs loads the list of all playlist IDs.
// Must be called with lock held.
func (r *PlaylistRepository) loadPlaylistIDs() ([]string, error) {
	data := r.prefs.String("playlist._ids")
	if data == "" {
		return []string{}, nil
	}

	var ids []string
	if err := json.Unmarshal([]byte(data), &ids); err != nil {
		return nil, domain.NewServiceError("PlaylistRepository", "loadPlaylistIDs", "failed to unmarshal IDs", err)
	}

	return ids, nil
}

// savePlaylistIDs saves the list of all playlist IDs.
// Must be called with lock held.
func (r *PlaylistRepository) savePlaylistIDs(ids []string) error {
	data, err := json.Marshal(ids)
	if err != nil {
		return domain.NewServiceError("PlaylistRepository", "savePlaylistIDs", "failed to marshal IDs", err)
	}

	r.prefs.SetString("playlist._ids", string(data))
	return nil
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func remove(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// Verify interface implementation
var _ ports.PlaylistRepository = (*PlaylistRepository)(nil)
