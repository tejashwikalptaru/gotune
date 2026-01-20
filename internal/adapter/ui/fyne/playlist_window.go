package fyne

import (
	"fmt"
	"strings"

	fyneapp "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/tejashwikalptaru/gotune/internal/adapter/ui/fyne/widgets"
	"github.com/tejashwikalptaru/gotune/internal/domain"
	"github.com/tejashwikalptaru/gotune/internal/ports"
)

// PlaylistWindow manages the playlist view window.
// It displays the current playback queue with search functionality
// and responds to playlist events for live updates.
type PlaylistWindow struct {
	window      fyneapp.Window
	app         fyneapp.App
	list        *widget.List
	searchEntry *widget.Entry

	// Data state
	data           []domain.MusicTrack // Filtered view (shown in the list)
	mainCollection []domain.MusicTrack // Full queue
	currentIndex   int                 // Selected track index

	// Dependencies
	presenter     *Presenter
	eventBus      ports.EventBus
	subscriptions []domain.SubscriptionID

	// Lifecycle
	onWindowClosed func()
	isVisible      bool
}

// NewPlaylistWindow creates a new playlist window.
// It initializes the UI, subscribes to events, and loads the current queue.
func NewPlaylistWindow(app fyneapp.App, presenter *Presenter, eventBus ports.EventBus) *PlaylistWindow {
	w := &PlaylistWindow{
		app:          app,
		presenter:    presenter,
		eventBus:     eventBus,
		currentIndex: -1,
	}

	// Create the window
	w.window = app.NewWindow("Playlist")
	w.window.Resize(fyneapp.NewSize(500, 600))

	// Build UI
	w.buildUI()

	// Subscribe to events
	w.subscribeToEvents()

	// Set window close handler
	w.window.SetOnClosed(func() {
		w.isVisible = false
		w.unsubscribeFromEvents()
		if w.onWindowClosed != nil {
			w.onWindowClosed()
		}
	})

	// Load initial queue data
	w.loadInitialData()

	return w
}

// buildUI constructs the playlist window UI layout.
func (w *PlaylistWindow) buildUI() {
	// Create the search entry
	w.searchEntry = widget.NewEntry()
	w.searchEntry.SetPlaceHolder("Search...")
	w.searchEntry.OnChanged = func(query string) {
		w.searchCollection(query)
	}

	// Create the list widget
	w.list = widget.NewList(
		func() int {
			return len(w.data)
		},
		func() fyneapp.CanvasObject {
			return w.createCell()
		},
		func(i widget.ListItemID, obj fyneapp.CanvasObject) {
			w.updateCell(i, obj)
		},
	)

	// Create layout
	content := container.NewBorder(
		w.searchEntry, // Top
		nil,           // Bottom
		nil,           // Left
		nil,           // Right
		w.list,        // Center
	)

	w.window.SetContent(content)
}

// createCell creates a new cell for the list.
func (w *PlaylistWindow) createCell() fyneapp.CanvasObject {
	return widgets.NewDoubleTapLabel(w.onCellDoubleTapped)
}

// updateCell updates a list cell with track information.
func (w *PlaylistWindow) updateCell(i widget.ListItemID, obj fyneapp.CanvasObject) {
	label, ok := obj.(*widgets.DoubleTapLabel)
	if !ok {
		return
	}

	if i < 0 || i >= len(w.data) {
		return
	}

	track := w.data[i]
	label.SetIndex(i)

	// Display the track title or filename if the title is empty
	displayText := track.Title
	if displayText == "" {
		displayText = track.FilePath
	}

	label.SetText(displayText)
}

// onCellDoubleTapped handles double-tap events on list cells.
func (w *PlaylistWindow) onCellDoubleTapped(index int) {
	if index < 0 || index >= len(w.data) {
		return
	}

	// When searching, we need to find the actual index in the main collection
	actualIndex := w.findActualIndex(index)
	if actualIndex == -1 {
		return
	}

	// Route through presenter (MVP pattern)
	if w.presenter != nil {
		_ = w.presenter.OnPlaylistTrackSelected(actualIndex)
	}
}

// findActualIndex finds the actual index in the mainCollection for a given filtered data index.
func (w *PlaylistWindow) findActualIndex(filteredIndex int) int {
	if filteredIndex < 0 || filteredIndex >= len(w.data) {
		return -1
	}

	selectedTrack := w.data[filteredIndex]

	// Find this track in the main collection
	for i, track := range w.mainCollection {
		if track.FilePath == selectedTrack.FilePath {
			return i
		}
	}

	return -1
}

// findFilteredIndex finds the filtered data index for a given main collection index.
// Returns -1 if the track at mainIndex is not in the filtered data (filtered out by search).
func (w *PlaylistWindow) findFilteredIndex(mainIndex int) int {
	if mainIndex < 0 || mainIndex >= len(w.mainCollection) {
		return -1
	}

	// If no filter is active, indices are the same
	if w.searchEntry.Text == "" {
		return mainIndex
	}

	// Find the track from main collection
	targetTrack := w.mainCollection[mainIndex]

	// Search for it in the filtered data by comparing file paths
	for i, track := range w.data {
		if track.FilePath == targetTrack.FilePath {
			return i
		}
	}

	// Track is filtered out
	return -1
}

// subscribeToEvents subscribes to playlist-related events.
func (w *PlaylistWindow) subscribeToEvents() {
	w.subscriptions = append(w.subscriptions,
		w.eventBus.Subscribe(domain.EventPlaylistUpdated, w.onPlaylistUpdated),
		w.eventBus.Subscribe(domain.EventTrackAdded, w.onTrackAdded),
	)
}

// unsubscribeFromEvents unsubscribes from all events.
func (w *PlaylistWindow) unsubscribeFromEvents() {
	for _, sub := range w.subscriptions {
		w.eventBus.Unsubscribe(sub)
	}
	w.subscriptions = nil
}

// onPlaylistUpdated handles PlaylistUpdatedEvent.
func (w *PlaylistWindow) onPlaylistUpdated(event domain.Event) {
	playlistEvent, ok := event.(domain.PlaylistUpdatedEvent)
	if !ok {
		return
	}

	fyneapp.Do(func() {
		w.mainCollection = playlistEvent.Playlist
		w.currentIndex = playlistEvent.Index

		// Re-apply search filter if active
		query := w.searchEntry.Text
		if query == "" {
			w.data = w.mainCollection
		} else {
			w.searchCollection(query)
		}

		w.updateWindowTitle()
		w.list.Refresh()

		// Highlight current track (map from main collection index to filtered index)
		if w.currentIndex >= 0 {
			filteredIndex := w.findFilteredIndex(w.currentIndex)
			if filteredIndex >= 0 {
				w.list.Select(filteredIndex)
			} else {
				// Current track is filtered out by search - unselect
				w.list.UnselectAll()
			}
		} else {
			w.list.UnselectAll()
		}
	})
}

// onTrackAdded handles TrackAddedEvent.
func (w *PlaylistWindow) onTrackAdded(event domain.Event) {
	trackEvent, ok := event.(domain.TrackAddedEvent)
	if !ok {
		return
	}

	fyneapp.Do(func() {
		// Add to the main collection
		w.mainCollection = append(w.mainCollection, trackEvent.Track)

		// If no search filter, add to visible data
		query := w.searchEntry.Text
		if query == "" {
			w.data = w.mainCollection
		} else {
			// Check if track matches current search
			if w.matchesSearch(trackEvent.Track, query) {
				w.data = append(w.data, trackEvent.Track)
			}
		}

		w.updateWindowTitle()
		w.list.Refresh()
	})
}

// searchCollection filters the playlist based on the search query.
func (w *PlaylistWindow) searchCollection(query string) {
	query = strings.ToLower(strings.TrimSpace(query))

	if query == "" {
		// No search, show all tracks
		w.data = w.mainCollection
		w.updateWindowTitle()
		w.list.Refresh()
		return
	}

	// Filter tracks
	filtered := make([]domain.MusicTrack, 0)
	for _, track := range w.mainCollection {
		if w.matchesSearch(track, query) {
			filtered = append(filtered, track)
		}
	}

	w.data = filtered
	w.updateWindowTitle()
	w.list.Refresh()
}

// matchesSearch checks if a track matches the search query.
func (w *PlaylistWindow) matchesSearch(track domain.MusicTrack, query string) bool {
	query = strings.ToLower(query)

	// Search across multiple fields
	if strings.Contains(strings.ToLower(track.FilePath), query) {
		return true
	}
	if strings.Contains(strings.ToLower(track.Title), query) {
		return true
	}
	if strings.Contains(strings.ToLower(track.Artist), query) {
		return true
	}
	if strings.Contains(strings.ToLower(track.Album), query) {
		return true
	}

	return false
}

// loadInitialData loads the current queue from the playlist service.
func (w *PlaylistWindow) loadInitialData() {
	if w.presenter == nil || w.presenter.playlistService == nil {
		return
	}

	w.mainCollection = w.presenter.playlistService.GetQueue()
	w.currentIndex = w.presenter.playlistService.GetCurrentIndex()
	w.data = w.mainCollection

	w.updateWindowTitle()
	w.list.Refresh()

	// Highlight current track (map from main collection index to filtered index)
	if w.currentIndex >= 0 {
		filteredIndex := w.findFilteredIndex(w.currentIndex)
		if filteredIndex >= 0 {
			w.list.Select(filteredIndex)
		}
		// If filtered out, don't highlight anything
	}
}

// updateWindowTitle updates the window title with the track count.
func (w *PlaylistWindow) updateWindowTitle() {
	count := len(w.data)
	title := fmt.Sprintf("Playlist (%d items)", count)
	w.window.SetTitle(title)
}

// Show displays the playlist window.
func (w *PlaylistWindow) Show() {
	w.isVisible = true
	w.window.Show()
}

// Close closes the playlist window.
func (w *PlaylistWindow) Close() {
	w.isVisible = false
	w.unsubscribeFromEvents()
	w.window.Close()
}

// IsVisible returns whether the window is currently visible.
func (w *PlaylistWindow) IsVisible() bool {
	return w.isVisible
}

// SetSelected highlights the track at the given index.
func (w *PlaylistWindow) SetSelected(index int) {
	fyneapp.Do(func() {
		w.currentIndex = index
		if index >= 0 {
			filteredIndex := w.findFilteredIndex(index)
			if filteredIndex >= 0 {
				w.list.Select(filteredIndex)
			} else {
				// Current track is filtered out by search
				w.list.UnselectAll()
			}
		} else {
			w.list.UnselectAll()
		}
	})
}

// SetOnWindowClosed sets a callback to be invoked when the window is closed.
// This allows the parent (MainWindow) to be notified and clear its reference.
func (w *PlaylistWindow) SetOnWindowClosed(callback func()) {
	w.onWindowClosed = callback
}
