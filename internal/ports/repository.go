// Package ports define repository interfaces for data persistence abstraction.
// These interfaces enable the repository pattern and allow swapping persistence mechanisms.
package ports

import (
	"github.com/tejashwikalptaru/gotune/internal/domain"
)

// PlaylistRepository handles the persistence of playlists.
// Implementations can use files, databases, or in-memory storage.
//
// Thread-safety: Implementations must be thread-safe.
type PlaylistRepository interface {
	// Save persists a playlist.
	// If a playlist with the same ID exists, it is replaced.
	//
	// Returns an error if saving fails.
	Save(playlist *domain.Playlist) error

	// Load retrieves a playlist by ID.
	// If the playlist doesn't exist, returns (nil, domain.ErrPlaylistNotFound).
	//
	// Returns the playlist or an error if loading fails.
	Load(id string) (*domain.Playlist, error)

	// LoadAll retrieves all saved playlists.
	//
	// Returns a slice of playlists (empty if none exist), or an error if loading fails.
	LoadAll() ([]*domain.Playlist, error)

	// Delete removes a playlist by ID.
	// If the playlist doesn't exist, this is a no-op (no error).
	//
	// Returns an error if deletion fails.
	Delete(id string) error

	// Exists checks if a playlist with the given ID exists.
	//
	// Returns true if the playlist exists, false otherwise.
	Exists(id string) bool
}

// HistoryRepository handles the persistence of playback history (queue and position).
// This replaces the hardcoded file paths in the original implementation.
//
// Thread-safety: Implementations must be thread-safe.
type HistoryRepository interface {
	// SaveQueue persists in the current playback queue.
	// This allows restoring the queue when the application restarts.
	//
	// Returns an error if saving fails.
	SaveQueue(tracks []domain.MusicTrack) error

	// LoadQueue retrieves the last saved playback queue.
	// If no queue was saved, returns an empty slice (not an error).
	//
	// Returns the queue or an error if loading fails.
	LoadQueue() ([]domain.MusicTrack, error)

	// SaveCurrentIndex persists the current track index in the queue.
	//
	// Returns an error if saving fails.
	SaveCurrentIndex(index int) error

	// LoadCurrentIndex retrieves the last saved track index.
	// If no index was saved, returns -1 (not an error).
	//
	// Returns the index or an error if loading fails.
	LoadCurrentIndex() (int, error)

	// Clear removes all saved history data.
	//
	// Returns an error if clearing fails.
	Clear() error
}

// PreferencesRepository handles the persistence of user preferences.
// This abstracts the Fyne preferences storage.
//
// Thread-safety: Implementations must be thread-safe.
type PreferencesRepository interface {
	// Volume preferences

	// SaveVolume persists at the volume level.
	//
	// Returns an error if saving fails.
	SaveVolume(volume float64) error

	// LoadVolume retrieves the saved volume level.
	// If no volume was saved, returns 1.0 (full volume) as default.
	//
	// Returns the volume or an error if loading fails.
	LoadVolume() (float64, error)

	// Loop mode preferences

	// SaveLoopMode persists in the loop mode state.
	//
	// Returns an error if saving fails.
	SaveLoopMode(enabled bool) error

	// LoadLoopMode retrieves the saved loop mode state.
	// If no loop mode was saved, returns false as default.
	//
	// Returns the loop mode, or an error if loading fails.
	LoadLoopMode() (bool, error)

	// Theme preferences

	// SaveTheme persists the theme preference.
	//
	// Returns an error if saving fails.
	SaveTheme(theme string) error

	// LoadTheme retrieves the saved theme preference.
	// If no theme was saved, returns "system" as default.
	//
	// Returns the theme or an error if loading fails.
	LoadTheme() (string, error)

	// Scan paths preferences

	// SaveScanPaths persists the list of directories to scan for music.
	//
	// Returns an error if saving fails.
	SaveScanPaths(paths []string) error

	// LoadScanPaths retrieves the saved scan paths.
	// If no paths were saved, returns an empty slice (not an error).
	//
	// Returns the paths or an error if loading fails.
	LoadScanPaths() ([]string, error)

	// Utility methods

	// Clear removes all saved preferences.
	//
	// Returns an error if clearing fails.
	Clear() error
}

// ErrPlaylistNotFound is returned when a requested playlist doesn't exist.
var ErrPlaylistNotFound = domain.ErrTrackNotFound // Reuse domain error
