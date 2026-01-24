// Package service provides business logic for the GoTune application.
package service

import (
	"log/slog"
	"sync"

	"github.com/tejashwikalptaru/gotune/internal/domain"
	"github.com/tejashwikalptaru/gotune/internal/ports"
)

// PlaylistService manages the playback queue and playlist operations.
// It handles adding tracks, navigation (next/previous), and persistence.
// All operations are thread-safe via sync.RWMutex.
type PlaylistService struct {
	// Dependencies (injected)
	logger     *slog.Logger
	playback   *PlaybackService
	repository ports.PlaylistRepository
	history    ports.HistoryRepository
	bus        ports.EventBus

	// State
	queue        []domain.MusicTrack
	currentIndex int

	// Concurrency control
	mu sync.RWMutex

	// Event subscription
	autoNextSub domain.SubscriptionID
}

// NewPlaylistService creates a new playlist service.
func NewPlaylistService(
	logger *slog.Logger,
	playback *PlaybackService,
	repository ports.PlaylistRepository,
	history ports.HistoryRepository,
	bus ports.EventBus,
) *PlaylistService {
	service := &PlaylistService{
		logger:       logger,
		playback:     playback,
		repository:   repository,
		history:      history,
		bus:          bus,
		queue:        make([]domain.MusicTrack, 0),
		currentIndex: -1,
	}

	logger.Debug("playlist service initialized")

	// Subscribe to auto-next events from the playback service
	service.autoNextSub = bus.Subscribe(domain.EventAutoNext, service.handleAutoNext)

	return service
}

// containsFilePath checks if the queue already contains a track with the given file path.
// Must be called with mutex lock held.
func (s *PlaylistService) containsFilePath(filePath string) bool {
	for _, track := range s.queue {
		if track.FilePath == filePath {
			return true
		}
	}
	return false
}

// AddTrack adds a track to the end of the queue.
// Returns ErrDuplicateTrack if a track with the same FilePath already exists.
func (s *PlaylistService) AddTrack(track domain.MusicTrack, playImmediately bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicates
	if s.containsFilePath(track.FilePath) {
		return domain.ErrDuplicateTrack
	}

	// Add to queue
	s.queue = append(s.queue, track)
	newIndex := len(s.queue) - 1

	// Publish TrackAdded event
	s.bus.Publish(domain.NewTrackAddedEvent(track, newIndex))

	// Play immediately if requested
	if playImmediately {
		s.currentIndex = newIndex
		if err := s.playback.LoadTrack(track, newIndex); err != nil {
			return err
		}
		if err := s.playback.Play(); err != nil {
			return err
		}
		// Publish playlist updated event with a NEW index
		s.bus.Publish(domain.NewPlaylistUpdatedEvent(s.queue, s.currentIndex))
		return nil
	} else {
		// Not playing immediately, publish with an unchanged index
		s.bus.Publish(domain.NewPlaylistUpdatedEvent(s.queue, s.currentIndex))
		return nil
	}
}

// AddTracks adds multiple tracks to the queue, filtering out any duplicates.
// Tracks with FilePath that already exist in the queue are silently skipped.
func (s *PlaylistService) AddTracks(tracks []domain.MusicTrack, playFirst bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(tracks) == 0 {
		return nil
	}

	// Filter out duplicate tracks
	uniqueTracks := make([]domain.MusicTrack, 0, len(tracks))
	for _, track := range tracks {
		if !s.containsFilePath(track.FilePath) {
			uniqueTracks = append(uniqueTracks, track)
		}
	}

	// If all tracks were duplicates, return early
	if len(uniqueTracks) == 0 {
		return nil
	}

	// Add unique tracks
	startIndex := len(s.queue)
	s.queue = append(s.queue, uniqueTracks...)

	// Publish events for each added track
	for i, track := range uniqueTracks {
		s.bus.Publish(domain.NewTrackAddedEvent(track, startIndex+i))
	}

	// Play the first track if requested
	if playFirst && len(uniqueTracks) > 0 {
		s.currentIndex = startIndex
		if err := s.playback.LoadTrack(uniqueTracks[0], startIndex); err != nil {
			return err
		}
		if err := s.playback.Play(); err != nil {
			return err
		}
		// Publish playlist updated event with a NEW index
		s.bus.Publish(domain.NewPlaylistUpdatedEvent(s.queue, s.currentIndex))
		return nil
	} else {
		// Not playing, publish with an unchanged index
		s.bus.Publish(domain.NewPlaylistUpdatedEvent(s.queue, s.currentIndex))
		return nil
	}
}

// RemoveTrack removes a track at the specified index.
func (s *PlaylistService) RemoveTrack(index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if index < 0 || index >= len(s.queue) {
		return domain.ErrTrackNotFound
	}

	// Remove track
	s.queue = append(s.queue[:index], s.queue[index+1:]...)

	// Adjust the current index if needed
	if s.currentIndex == index {
		// Stopped playing the removed track
		if err := s.playback.Stop(); err != nil {
			s.logger.Warn("failed to stop playback", slog.Any("error", err))
		}
		s.currentIndex = -1
	} else if s.currentIndex > index {
		// Shift index down
		s.currentIndex--
	}

	// Publish event
	s.bus.Publish(domain.NewPlaylistUpdatedEvent(s.queue, s.currentIndex))

	return nil
}

// ClearQueue removes all tracks from the queue.
func (s *PlaylistService) ClearQueue() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop playback
	if err := s.playback.Stop(); err != nil {
		s.logger.Warn("failed to stop playback", slog.Any("error", err))
	}

	// Clear queue
	s.queue = make([]domain.MusicTrack, 0)
	s.currentIndex = -1

	// Publish event
	s.bus.Publish(domain.NewQueueChangedEvent(s.queue))

	return nil
}

// PlayTrackAt plays the track at the specified index.
func (s *PlaylistService) PlayTrackAt(index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if index < 0 || index >= len(s.queue) {
		return domain.ErrTrackNotFound
	}

	s.currentIndex = index
	track := s.queue[index]

	// Load and play
	if err := s.playback.LoadTrack(track, index); err != nil {
		return err
	}

	if err := s.playback.Play(); err != nil {
		return err
	}

	// Publish playlist updated event
	s.bus.Publish(domain.NewPlaylistUpdatedEvent(s.queue, s.currentIndex))

	return nil
}

// PlayTrackByPath plays a track from the queue by its file path.
// Returns the index of the track, or -1 if not found.
func (s *PlaylistService) PlayTrackByPath(filePath string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find track in the queue
	index := -1
	for i, track := range s.queue {
		if track.FilePath == filePath {
			index = i
			break
		}
	}

	if index == -1 {
		return -1, domain.ErrTrackNotFound
	}

	s.currentIndex = index
	track := s.queue[index]

	// Load and play
	if err := s.playback.LoadTrack(track, index); err != nil {
		return index, err
	}

	if err := s.playback.Play(); err != nil {
		return index, err
	}

	return index, nil
}

// PlayNext plays the next track in the queue.
func (s *PlaylistService) PlayNext() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.queue) == 0 {
		return domain.ErrQueueEmpty
	}

	// Check if there's a next track
	if s.currentIndex >= len(s.queue)-1 {
		return domain.ErrEndOfQueue
	}

	s.currentIndex++
	track := s.queue[s.currentIndex]

	// Load and play
	if err := s.playback.LoadTrack(track, s.currentIndex); err != nil {
		return err
	}

	if err := s.playback.Play(); err != nil {
		return err
	}

	// Publish playlist updated event
	s.bus.Publish(domain.NewPlaylistUpdatedEvent(s.queue, s.currentIndex))

	return nil
}

// PlayPrevious plays the previous track in the queue.
func (s *PlaylistService) PlayPrevious() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.queue) == 0 {
		return domain.ErrQueueEmpty
	}

	// Check if there's a previous track
	if s.currentIndex <= 0 {
		return domain.ErrStartOfQueue
	}

	s.currentIndex--
	track := s.queue[s.currentIndex]

	// Load and play
	if err := s.playback.LoadTrack(track, s.currentIndex); err != nil {
		return err
	}

	if err := s.playback.Play(); err != nil {
		return err
	}

	// Publish playlist updated event
	s.bus.Publish(domain.NewPlaylistUpdatedEvent(s.queue, s.currentIndex))

	return nil
}

// GetQueue returns a copy of the current queue.
func (s *PlaylistService) GetQueue() []domain.MusicTrack {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modification
	queue := make([]domain.MusicTrack, len(s.queue))
	copy(queue, s.queue)
	return queue
}

// GetCurrentIndex returns the index of the currently playing track.
func (s *PlaylistService) GetCurrentIndex() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.currentIndex
}

// GetQueueLength returns the number of tracks in the queue.
func (s *PlaylistService) GetQueueLength() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.queue)
}

// SaveQueue saves the current queue to the history repository.
func (s *PlaylistService) SaveQueue() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.history.SaveQueue(s.queue); err != nil {
		return err
	}

	if err := s.history.SaveCurrentIndex(s.currentIndex); err != nil {
		return err
	}

	return nil
}

// LoadQueue loads the queue from the history repository.
func (s *PlaylistService) LoadQueue() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Load queue
	queue, err := s.history.LoadQueue()
	if err != nil {
		return err
	}

	// Load the current index
	index, err := s.history.LoadCurrentIndex()
	if err != nil {
		// Default to -1 if not found
		index = -1
	}

	// Update state
	s.queue = queue
	s.currentIndex = index

	// Publish event
	s.bus.Publish(domain.NewPlaylistUpdatedEvent(s.queue, s.currentIndex))

	return nil
}

// MoveTrack moves a track from one index to another.
func (s *PlaylistService) MoveTrack(fromIndex, toIndex int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if fromIndex < 0 || fromIndex >= len(s.queue) {
		return domain.ErrTrackNotFound
	}

	if toIndex < 0 || toIndex >= len(s.queue) {
		return domain.ErrTrackNotFound
	}

	if fromIndex == toIndex {
		return nil
	}

	// Remove track from the old position
	track := s.queue[fromIndex]
	s.queue = append(s.queue[:fromIndex], s.queue[fromIndex+1:]...)

	// Insert at a new position
	s.queue = append(s.queue[:toIndex], append([]domain.MusicTrack{track}, s.queue[toIndex:]...)...)

	// Adjust the current index if needed
	switch {
	case s.currentIndex == fromIndex:
		s.currentIndex = toIndex
	case fromIndex < s.currentIndex && toIndex >= s.currentIndex:
		s.currentIndex--
	case fromIndex > s.currentIndex && toIndex <= s.currentIndex:
		s.currentIndex++
	}

	// Publish event
	s.bus.Publish(domain.NewPlaylistUpdatedEvent(s.queue, s.currentIndex))

	return nil
}

// Shuffle randomizes the order of tracks in the queue (except currently playing).
func (s *PlaylistService) Shuffle() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.queue) <= 1 {
		return nil
	}

	// TODO: Implement shuffle algorithm
	// For now, just publish update event
	s.bus.Publish(domain.NewPlaylistUpdatedEvent(s.queue, s.currentIndex))

	return nil
}

// handleAutoNext is called when a track finishes playing and auto-next is requested.
func (s *PlaylistService) handleAutoNext(event domain.Event) {
	autoNextEvent, ok := event.(domain.AutoNextEvent)
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify the event is for the current track
	if autoNextEvent.CurrentIndex != s.currentIndex {
		return
	}

	// Check if there's a next track
	if s.currentIndex >= len(s.queue)-1 {
		// End of queue - stop playback to clean up state
		s.mu.Unlock()
		if err := s.playback.Stop(); err != nil {
			s.logger.Warn("failed to stop playback at end of queue", slog.Any("error", err))
		}
		s.mu.Lock()
		return
	}

	// Play the next track
	s.currentIndex++
	track := s.queue[s.currentIndex]

	// Load and play (unlock first to avoid deadlock)
	s.mu.Unlock()
	if err := s.playback.LoadTrack(track, s.currentIndex); err != nil {
		s.logger.Warn("failed to load next track", slog.Any("error", err))
		s.mu.Lock()
		return
	}
	if err := s.playback.Play(); err != nil {
		s.logger.Warn("failed to play next track", slog.Any("error", err))
	}

	// Publish playlist updated event
	s.bus.Publish(domain.NewPlaylistUpdatedEvent(s.queue, s.currentIndex))

	s.mu.Lock()
}

// Shutdown cleans up resources.
func (s *PlaylistService) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Unsubscribe from events
	s.bus.Unsubscribe(s.autoNextSub)

	// Save queue before shutdown (the best effort)
	if err := s.history.SaveQueue(s.queue); err != nil {
		s.logger.Warn("failed to save queue on shutdown", slog.Any("error", err))
	}
	if err := s.history.SaveCurrentIndex(s.currentIndex); err != nil {
		s.logger.Warn("failed to save current index on shutdown", slog.Any("error", err))
	}

	return nil
}

// Verify that PlaylistService implements the expected interface patterns
var _ interface {
	AddTrack(domain.MusicTrack, bool) error
	AddTracks([]domain.MusicTrack, bool) error
	RemoveTrack(int) error
	ClearQueue() error
	PlayTrackAt(int) error
	PlayTrackByPath(string) (int, error)
	PlayNext() error
	PlayPrevious() error
	GetQueue() []domain.MusicTrack
	GetCurrentIndex() int
	GetQueueLength() int
	SaveQueue() error
	LoadQueue() error
	MoveTrack(int, int) error
	Shuffle() error
	Shutdown() error
} = (*PlaylistService)(nil)
