// Package ports define the UI interface for view abstraction.
// This interface allows the presenter to update the UI without depending on Fyne directly.
package ports

import (
	"github.com/tejashwikalptaru/gotune/internal/domain"
)

// UI is the interface for the user interface layer.
// This abstracts the Fyne UI implementation and allows for testing without a real UI.
//
// The presenter layer will receive events from the event bus and call these methods
// to update the UI accordingly. This creates a clean separation between business logic
// (services), presentation logic (presenter), and view rendering (UI).
//
// Thread-safety: All methods must be called from the main UI thread.
// The Fyne framework handles thread-safety internally.
type UI interface {
	// Display update methods

	// SetTrackInfo updates the displayed track information.
	// This includes title, artist, album, etc.
	SetTrackInfo(track domain.MusicTrack)

	// SetAlbumArt updates the displayed album artwork.
	// imageData: Raw image bytes (JPEG, PNG, etc.)
	SetAlbumArt(imageData []byte)

	// ClearAlbumArt removes the currently displayed album art.
	ClearAlbumArt()

	// SetCurrentTime updates the current playback position display.
	// position: Current position in the track
	SetCurrentTime(position float64)

	// SetTotalTime updates the total track duration display.
	// duration: Total track length
	SetTotalTime(duration float64)

	// Playback state update methods

	// SetPlayState updates the play/pause button state.
	// playing: true if currently playing, false if paused/stopped
	SetPlayState(playing bool)

	// SetProgress updates the progress slider position.
	// current: Current position in seconds
	// total: Total duration in seconds
	SetProgress(current, total float64)

	// Volume/Mute state update methods

	// SetVolume updates the volume slider and display.
	// volume: Volume level (0.0 to 1.0, will be scaled to 0-100 for UI)
	SetVolume(volume float64)

	// SetMuteState updates the mute button state.
	// muted: true if audio is muted, false otherwise
	SetMuteState(muted bool)

	// SetLoopState updates the loop button state.
	// enabled: true if loop mode is enabled, false otherwise
	SetLoopState(enabled bool)

	// Playlist update methods

	// ShowPlaylist displays the playlist window with the given tracks.
	// tracks: The list of tracks in the playlist
	// currentIndex: The index of the currently playing track (-1 if none)
	ShowPlaylist(tracks []domain.MusicTrack, currentIndex int)

	// UpdatePlaylistSelection highlights the currently playing track in the playlist.
	// index: The index of the track to highlight
	UpdatePlaylistSelection(index int)

	// UpdatePlaylistWindow refreshes the playlist window with new data.
	// This is called when tracks are added or the queue changes.
	UpdatePlaylistWindow(tracks []domain.MusicTrack)

	// Notification methods

	// ShowNotification displays a temporary notification to the user.
	// title: Notification title
	// message: Notification message
	ShowNotification(title, message string)

	// ShowError displays an error dialog to the user.
	// title: Error dialog title
	// message: Error message
	ShowError(title, message string)

	// ShowInfo displays an informational dialog.
	// title: Info dialog title
	// message: Info message
	ShowInfo(title, message string)

	// Scan progress methods

	// ShowScanProgress displays the file scanning progress dialog.
	// currentFile: The file currently being scanned
	// filesScanned: Number of files processed
	// totalFiles: Total number of files (-1 if unknown)
	ShowScanProgress(currentFile string, filesScanned, totalFiles int)

	// HideScanProgress closes the file scanning progress dialog.
	HideScanProgress()

	// History/Preferences persistence (delegated to repositories)

	// SaveHistory saves the current playback queue to preferences.
	// This delegates to the history repository via the presenter.
	SaveHistory(data string)

	// GetHistory retrieves the saved playback queue from preferences.
	// This delegates to the history repository via the presenter.
	GetHistory() string

	// Lifecycle methods

	// Run starts the UI event loop.
	// This is a blocking call that runs until the application quits.
	//
	// Returns an error if the UI fails to start.
	Run() error

	// Quit closes the application.
	// This should trigger cleanup and shutdown of all services.
	Quit()

	// Free releases UI resources.
	// Called during application shutdown.
	Free()
}

// UIFactory is a function that creates a UI instance.
// This allows for dependency injection of different UI implementations.
type UIFactory func() (UI, error)
