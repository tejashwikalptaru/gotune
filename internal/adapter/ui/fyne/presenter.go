// Package fyne provides Fyne UI adapter implementations.
// This package implements the UI layer using the Fyne toolkit.
package fyne

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/tejashwikalptaru/gotune/internal/domain"
	"github.com/tejashwikalptaru/gotune/internal/ports"
	"github.com/tejashwikalptaru/gotune/internal/service"
)

// UIView defines the interface for UI updates.
// The actual UI implementation (MainWindow) must implement this interface.
type UIView interface {
	// Playback state updates
	SetPlayState(playing bool)
	SetMuteState(muted bool)
	SetLoopState(enabled bool)
	SetVolume(volume float64)

	// Track information updates
	SetTrackInfo(title, artist, album string)
	SetAlbumArt(imageData []byte)
	ClearAlbumArt()

	// Progress updates
	SetCurrentTime(seconds float64)
	SetTotalTime(seconds float64)
	SetProgress(position, duration float64)

	// Playlist updates
	UpdatePlaylistSelection(index int)

	// Playlist window management
	ShowPlaylistWindow()
	ClosePlaylistWindow()
	IsPlaylistWindowOpen() bool

	// Notifications
	ShowNotification(title, message string)
}

// Presenter implements the Presenter pattern (MVP architecture).
// It coordinates between services and the UI, handling all event-driven updates.
//
// Responsibilities:
// - Subscribe to events from the event bus
// - Map domain events to UI updates
// - Translate UI commands to service method calls
// - Maintain presentation state
//
// Thread-safety: All operations are thread-safe via sync.RWMutex.
type Presenter struct {
	// Dependencies
	logger *slog.Logger

	// Services (injected)
	playbackService   *service.PlaybackService
	playlistService   *service.PlaylistService
	libraryService    *service.LibraryService
	preferenceService *service.PreferenceService

	// Event bus for subscriptions (exported for PlaylistWindow access)
	EventBus ports.EventBus

	// UI view
	view UIView

	// Presentation state
	currentTrack     *domain.MusicTrack
	isPlaying        bool
	progressTicker   *time.Ticker
	stopProgressChan chan bool

	// Concurrency control
	mu           sync.RWMutex
	shutdownOnce sync.Once
}

// NewPresenter creates a new presenter.
func NewPresenter(
	logger *slog.Logger,
	playbackService *service.PlaybackService,
	playlistService *service.PlaylistService,
	libraryService *service.LibraryService,
	preferenceService *service.PreferenceService,
	eventBus ports.EventBus,
	view UIView,
) *Presenter {
	p := &Presenter{
		logger:            logger,
		playbackService:   playbackService,
		playlistService:   playlistService,
		libraryService:    libraryService,
		preferenceService: preferenceService,
		EventBus:          eventBus,
		view:              view,
		stopProgressChan:  make(chan bool, 1),
	}

	// Subscribe to events
	p.subscribeToEvents()

	// Sync UI with current state
	p.syncInitialState()

	// Start progress ticker
	p.startProgressUpdates()

	return p
}

// subscribeToEvents subscribes to all relevant events from the event bus.
func (p *Presenter) subscribeToEvents() {
	subscriptions := map[domain.EventType]domain.EventHandler{
		// Playback events
		domain.EventTrackLoaded:    p.onTrackLoaded,
		domain.EventTrackStarted:   p.onTrackStarted,
		domain.EventTrackPaused:    p.onTrackPaused,
		domain.EventTrackStopped:   p.onTrackStopped,
		domain.EventTrackCompleted: p.onTrackCompleted,

		// Volume events
		domain.EventVolumeChanged: p.onVolumeChanged,
		domain.EventMuteToggled:   p.onMuteToggled,
		domain.EventLoopToggled:   p.onLoopToggled,

		// Playlist events
		domain.EventPlaylistUpdated: p.onPlaylistUpdated,

		// Scan events
		domain.EventScanStarted:   p.onScanStarted,
		domain.EventScanProgress:  p.onScanProgress,
		domain.EventScanCompleted: p.onScanCompleted,
		domain.EventScanCancelled: p.onScanCancelled,
	}

	for eventType, handler := range subscriptions {
		p.EventBus.Subscribe(eventType, handler)
	}
}

// syncInitialState synchronizes the UI with the current application state.
// This is called during presenter initialization to ensure the UI reflects
// the current state of services (volume, loop mode, loaded track, etc.).
func (p *Presenter) syncInitialState() {
	state := p.playbackService.GetState()

	// Update UI with current values
	p.view.SetVolume(state.Volume * 100.0) // Convert from 0.0-1.0 to 0-100
	p.view.SetLoopState(state.IsLooping)
	p.view.SetMuteState(state.IsMuted)

	// If a track is already loaded, update track info
	if state.CurrentTrack != nil {
		p.view.SetTrackInfo(
			state.CurrentTrack.Title,
			state.CurrentTrack.Artist,
			state.CurrentTrack.Album,
		)

		if state.Duration > 0 {
			p.view.SetTotalTime(state.Duration.Seconds())
		}

		// Update album art if available
		if state.CurrentTrack.Metadata != nil &&
			len(state.CurrentTrack.Metadata.AlbumArt) > 0 {
			p.view.SetAlbumArt(state.CurrentTrack.Metadata.AlbumArt)
		} else {
			p.view.ClearAlbumArt()
		}
	}

	// Update play state
	p.view.SetPlayState(state.Status == domain.StatusPlaying)

	// Update progress if track is loaded
	if state.Duration > 0 {
		p.view.SetProgress(state.Position.Seconds(), state.Duration.Seconds())
		p.view.SetCurrentTime(state.Position.Seconds())
	}
}

// Event handlers

func (p *Presenter) onTrackLoaded(event domain.Event) {
	e, ok := event.(domain.TrackLoadedEvent)
	if !ok {
		return
	}

	p.mu.Lock()
	p.currentTrack = &e.Track
	p.mu.Unlock()

	// Update UI
	p.view.SetTrackInfo(e.Track.Title, e.Track.Artist, e.Track.Album)

	// Set total time (convert time.Duration to seconds)
	if e.Duration > 0 {
		seconds := e.Duration.Seconds()
		p.view.SetTotalTime(seconds)
	}

	// Set album art (check the Metadata field)
	if e.Track.Metadata != nil && len(e.Track.Metadata.AlbumArt) > 0 {
		p.view.SetAlbumArt(e.Track.Metadata.AlbumArt)
	} else {
		p.view.ClearAlbumArt()
	}
}

func (p *Presenter) onTrackStarted(event domain.Event) {
	p.mu.Lock()
	p.isPlaying = true
	p.mu.Unlock()

	p.view.SetPlayState(true)
}

func (p *Presenter) onTrackPaused(event domain.Event) {
	p.mu.Lock()
	p.isPlaying = false
	p.mu.Unlock()

	p.view.SetPlayState(false)
}

func (p *Presenter) onTrackStopped(event domain.Event) {
	p.mu.Lock()
	p.isPlaying = false
	p.mu.Unlock()

	p.view.SetPlayState(false)
	p.view.SetCurrentTime(0)
	p.view.SetProgress(0, 1)
}

func (p *Presenter) onTrackCompleted(event domain.Event) {
	// Track completed - the next track will be loaded automatically by PlaylistService
	p.mu.Lock()
	p.isPlaying = false
	p.mu.Unlock()

	// Update UI to show play state (not pause)
	p.view.SetPlayState(false)
}

func (p *Presenter) onVolumeChanged(event domain.Event) {
	e, ok := event.(domain.VolumeChangedEvent)
	if !ok {
		return
	}

	p.view.SetVolume(e.Volume)
}

func (p *Presenter) onMuteToggled(event domain.Event) {
	e, ok := event.(domain.MuteToggledEvent)
	if !ok {
		return
	}

	p.view.SetMuteState(e.Muted)
}

func (p *Presenter) onLoopToggled(event domain.Event) {
	e, ok := event.(domain.LoopToggledEvent)
	if !ok {
		return
	}

	p.view.SetLoopState(e.Enabled)
}

func (p *Presenter) onPlaylistUpdated(event domain.Event) {
	e, ok := event.(domain.PlaylistUpdatedEvent)
	if !ok {
		return
	}

	// Update playlist selection in the UI
	p.view.UpdatePlaylistSelection(e.Index)
}

func (p *Presenter) onScanStarted(event domain.Event) {
	e, ok := event.(domain.ScanStartedEvent)
	if !ok {
		return
	}

	p.view.ShowNotification("Scan Started", fmt.Sprintf("Scanning: %s", e.Path))
}

func (p *Presenter) onScanProgress(event domain.Event) {
	// Could update a progress bar if UI supports it
	// For now, we just ignore these high-frequency events
}

func (p *Presenter) onScanCompleted(event domain.Event) {
	e, ok := event.(domain.ScanCompletedEvent)
	if !ok {
		return
	}

	message := fmt.Sprintf("Found %d tracks", len(e.TracksFound))
	p.view.ShowNotification("Scan Complete", message)
}

func (p *Presenter) onScanCancelled(event domain.Event) {
	p.view.ShowNotification("Scan Cancelled", "Scan was cancelled")
}

func (p *Presenter) startProgressUpdates() {
	p.progressTicker = time.NewTicker(250 * time.Millisecond)

	go func() {
		for {
			select {
			case <-p.progressTicker.C:
				p.updateProgress()
			case <-p.stopProgressChan:
				return
			}
		}
	}()
}

func (p *Presenter) updateProgress() {
	p.mu.RLock()
	currentTrack := p.currentTrack
	p.mu.RUnlock()

	// Only update if a track is loaded
	if currentTrack == nil {
		return
	}

	state := p.playbackService.GetState()
	if state.Duration <= 0 {
		return
	}

	p.view.SetCurrentTime(state.Position.Seconds())
	p.view.SetProgress(state.Position.Seconds(), state.Duration.Seconds())
}

// UI Command handlers (called by UI)

// OnPlayClicked handles the play button click.
func (p *Presenter) OnPlayClicked() {
	state := p.playbackService.GetState()

	var err error
	if state.Status == domain.StatusPaused {
		// Resume
		err = p.playbackService.Play()
	} else if state.Status == domain.StatusStopped {
		// Start playback
		err = p.playbackService.Play()
	} else {
		// Already playing - pause
		err = p.playbackService.Pause()
	}

	if err != nil {
		// Log error for debugging
		p.logger.Error("play/pause failed", slog.Any("error", err))

		// Show error to the user
		p.view.ShowNotification("Playback Error",
			fmt.Sprintf("Failed to start playback: %v", err))
	}
}

// OnStopClicked handle, the stop button click.
func (p *Presenter) OnStopClicked() {
	if err := p.playbackService.Stop(); err != nil {
		p.logger.Error("stop failed", slog.Any("error", err))
		p.view.ShowNotification("Playback Error",
			fmt.Sprintf("Failed to stop playback: %v", err))
	}
}

// OnNextClicked handles, the next button click.
func (p *Presenter) OnNextClicked() {
	if err := p.playlistService.PlayNext(); err != nil {
		p.logger.Error("next track failed", slog.Any("error", err))
		p.view.ShowNotification("Playlist Error",
			fmt.Sprintf("Failed to play next track: %v", err))
	}
}

// OnPreviousClicked handles the previous button click.
func (p *Presenter) OnPreviousClicked() {
	if err := p.playlistService.PlayPrevious(); err != nil {
		p.logger.Error("previous track failed", slog.Any("error", err))
		p.view.ShowNotification("Playlist Error",
			fmt.Sprintf("Failed to play previous track: %v", err))
	}
}

// OnVolumeChanged handles volume slider changes.
func (p *Presenter) OnVolumeChanged(volume float64) {
	// Normalize from 0-100 to 0.0-1.0
	normalized := volume / 100.0
	if err := p.playbackService.SetVolume(normalized); err != nil {
		p.logger.Error("volume change failed", slog.Any("error", err))
		p.view.ShowNotification("Volume Error",
			fmt.Sprintf("Failed to change volume: %v", err))
	}
	p.preferenceService.SetVolume(normalized)
}

// OnMuteClicked handles the mute button click.
func (p *Presenter) OnMuteClicked() {
	state := p.playbackService.GetState()
	p.playbackService.Mute(!state.IsMuted)
}

// OnLoopClicked handles the loop button click.
func (p *Presenter) OnLoopClicked() {
	state := p.playbackService.GetState()
	newLoopState := !state.IsLooping
	p.playbackService.SetLoop(newLoopState)
	p.preferenceService.SetLoopMode(newLoopState)
}

// OnSeekRequested handles seek requests from the progress slider.
func (p *Presenter) OnSeekRequested(position float64) {
	// Convert seconds to time.Duration
	positionDuration := time.Duration(position * float64(time.Second))
	if err := p.playbackService.Seek(positionDuration); err != nil {
		p.logger.Error("seek failed", slog.Any("error", err))
		p.view.ShowNotification("Seek Error",
			fmt.Sprintf("Failed to seek: %v", err))
	}
}

// OnFileOpened handles file open requests.
func (p *Presenter) OnFileOpened(filePath string) error {
	// Extract metadata
	track, err := p.libraryService.ExtractMetadata(filePath)
	if err != nil {
		return err
	}

	// Add to the playlist and play immediately
	err = p.playlistService.AddTrack(*track, true)
	if err != nil {
		// Silently ignore duplicate errors (user preference)
		if err == domain.ErrDuplicateTrack {
			return nil
		}
		return err
	}

	return nil
}

// OnFolderOpened handles folder open requests.
func (p *Presenter) OnFolderOpened(folderPath string) error {
	// Scan folder
	tracks, err := p.libraryService.ScanFolder(folderPath)
	if err != nil {
		return err
	}

	// Add all tracks to the playlist (don't play first automatically)
	err = p.playlistService.AddTracks(tracks, false)
	if err != nil {
		return err
	}

	// Save the last folder
	p.preferenceService.SetLastFolder(folderPath)

	return nil
}

// OnTrackSelected handles track selection from playlist.
func (p *Presenter) OnTrackSelected(trackPath string) error {
	// PlayTrackByPath returns (index, error)
	_, err := p.playlistService.PlayTrackByPath(trackPath)
	return err
}

// OnPlaylistMenuClicked handles "View Playlist" menu action.
func (p *Presenter) OnPlaylistMenuClicked() {
	p.view.ShowPlaylistWindow()
}

// OnPlaylistTrackSelected handles track selection from playlist window by index.
func (p *Presenter) OnPlaylistTrackSelected(index int) error {
	return p.playlistService.PlayTrackAt(index)
}

// GetQueue returns the current queue.
func (p *Presenter) GetQueue() []domain.MusicTrack {
	return p.playlistService.GetQueue()
}

// Shutdown cleans up resources.
// It's safe to call multiple times (idempotent).
func (p *Presenter) Shutdown() {
	p.shutdownOnce.Do(func() {
		// Stop the ticker first to prevent new iterations
		if p.progressTicker != nil {
			p.progressTicker.Stop()
		}

		// Close channel to signal goroutine to exit
		close(p.stopProgressChan)
	})
}
