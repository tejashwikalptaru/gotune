package service

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tejashwikalptaru/gotune/internal/adapter/audio/mock"
	"github.com/tejashwikalptaru/gotune/internal/adapter/eventbus"
	"github.com/tejashwikalptaru/gotune/internal/domain"
)

// Mock repositories for testing

type mockPlaylistRepository struct {
	mu        sync.RWMutex
	playlists map[string]*domain.Playlist
}

func newMockPlaylistRepository() *mockPlaylistRepository {
	return &mockPlaylistRepository{
		playlists: make(map[string]*domain.Playlist),
	}
}

func (m *mockPlaylistRepository) Save(playlist *domain.Playlist) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.playlists[playlist.ID] = playlist
	return nil
}

func (m *mockPlaylistRepository) Load(id string) (*domain.Playlist, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	playlist, ok := m.playlists[id]
	if !ok {
		return nil, domain.ErrPlaylistEmpty
	}
	return playlist, nil
}

func (m *mockPlaylistRepository) LoadAll() ([]*domain.Playlist, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	playlists := make([]*domain.Playlist, 0, len(m.playlists))
	for _, p := range m.playlists {
		playlists = append(playlists, p)
	}
	return playlists, nil
}

func (m *mockPlaylistRepository) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.playlists, id)
	return nil
}

func (m *mockPlaylistRepository) Exists(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.playlists[id]
	return ok
}

type mockHistoryRepository struct {
	mu           sync.RWMutex
	queue        []domain.MusicTrack
	currentIndex int
}

func newMockHistoryRepository() *mockHistoryRepository {
	return &mockHistoryRepository{
		queue:        make([]domain.MusicTrack, 0),
		currentIndex: -1,
	}
}

func (m *mockHistoryRepository) SaveQueue(tracks []domain.MusicTrack) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queue = tracks
	return nil
}

func (m *mockHistoryRepository) LoadQueue() ([]domain.MusicTrack, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.queue, nil
}

func (m *mockHistoryRepository) SaveCurrentIndex(index int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentIndex = index
	return nil
}

func (m *mockHistoryRepository) LoadCurrentIndex() (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentIndex, nil
}

func (m *mockHistoryRepository) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queue = make([]domain.MusicTrack, 0)
	m.currentIndex = -1
	return nil
}

// Helper to create a test playlist service
func newTestPlaylistService() (*PlaylistService, *PlaybackService, *eventbus.SyncEventBus) {
	engine := mock.NewEngine()
	engine.Initialize(-1, 44100, 0)

	bus := eventbus.NewSyncEventBus()
	playback := NewPlaybackService(engine, bus)
	plRepo := newMockPlaylistRepository()
	histRepo := newMockHistoryRepository()

	playlist := NewPlaylistService(playback, plRepo, histRepo, bus)

	return playlist, playback, bus
}

func TestPlaylistService_AddTrack(t *testing.T) {
	service, _, bus := newTestPlaylistService()
	defer service.Shutdown()

	track := createTestTrack("1", "Song 1", "/test/song1.mp3")

	// Subscribe to events
	var addedEvent domain.TrackAddedEvent
	var updatedEvent domain.PlaylistUpdatedEvent
	bus.Subscribe(domain.EventTrackAdded, func(e domain.Event) {
		addedEvent = e.(domain.TrackAddedEvent)
	})
	bus.Subscribe(domain.EventPlaylistUpdated, func(e domain.Event) {
		updatedEvent = e.(domain.PlaylistUpdatedEvent)
	})

	// Add track without playing
	err := service.AddTrack(track, false)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, 1, service.GetQueueLength())
	assert.Equal(t, track.ID, addedEvent.Track.ID)
	assert.Equal(t, 0, addedEvent.Index)
	assert.Equal(t, 1, len(updatedEvent.Playlist))
}

func TestPlaylistService_AddTrack_PlayImmediately(t *testing.T) {
	service, playback, _ := newTestPlaylistService()
	defer service.Shutdown()

	track := createTestTrack("1", "Song 1", "/test/song1.mp3")

	// Add track and play immediately
	err := service.AddTrack(track, true)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, 1, service.GetQueueLength())
	assert.Equal(t, 0, service.GetCurrentIndex())

	// Verify playback state
	state := playback.GetState()
	assert.NotNil(t, state.CurrentTrack)
	assert.Equal(t, track.ID, state.CurrentTrack.ID)
	assert.Equal(t, domain.StatusPlaying, state.Status)
}

func TestPlaylistService_AddTrack_ReplaceCurrentlyPlaying(t *testing.T) {
	service, playback, _ := newTestPlaylistService()
	defer service.Shutdown()

	trackA := createTestTrack("A", "Song A", "/test/songA.mp3")
	trackB := createTestTrack("B", "Song B", "/test/songB.mp3")

	// Add and play first track
	err := service.AddTrack(trackA, true)
	require.NoError(t, err)

	// Verify first track is playing
	state := playback.GetState()
	assert.NotNil(t, state.CurrentTrack)
	assert.Equal(t, "A", state.CurrentTrack.ID)
	assert.Equal(t, domain.StatusPlaying, state.Status)

	// Add second track with playImmediately=true (should replace current)
	err = service.AddTrack(trackB, true)
	require.NoError(t, err)

	// Verify second track is now playing
	state = playback.GetState()
	assert.NotNil(t, state.CurrentTrack)
	assert.Equal(t, "B", state.CurrentTrack.ID)
	assert.Equal(t, domain.StatusPlaying, state.Status)

	// Verify queue has both tracks
	assert.Equal(t, 2, service.GetQueueLength())
	assert.Equal(t, 1, service.GetCurrentIndex())
}

func TestPlaylistService_AddTracks(t *testing.T) {
	service, _, bus := newTestPlaylistService()
	defer service.Shutdown()

	tracks := []domain.MusicTrack{
		createTestTrack("1", "Song 1", "/test/song1.mp3"),
		createTestTrack("2", "Song 2", "/test/song2.mp3"),
		createTestTrack("3", "Song 3", "/test/song3.mp3"),
	}

	// Count events
	addedCount := 0
	bus.Subscribe(domain.EventTrackAdded, func(e domain.Event) {
		addedCount++
	})

	// Add multiple tracks
	err := service.AddTracks(tracks, false)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, 3, service.GetQueueLength())
	assert.Equal(t, 3, addedCount)
}

func TestPlaylistService_AddTracks_PlayFirst(t *testing.T) {
	service, playback, _ := newTestPlaylistService()
	defer service.Shutdown()

	tracks := []domain.MusicTrack{
		createTestTrack("1", "Song 1", "/test/song1.mp3"),
		createTestTrack("2", "Song 2", "/test/song2.mp3"),
		createTestTrack("3", "Song 3", "/test/song3.mp3"),
	}

	// Add multiple tracks and play first
	err := service.AddTracks(tracks, true)
	require.NoError(t, err)

	// Verify queue
	assert.Equal(t, 3, service.GetQueueLength())
	assert.Equal(t, 0, service.GetCurrentIndex())

	// Verify playback state - first track should be playing
	state := playback.GetState()
	assert.NotNil(t, state.CurrentTrack)
	assert.Equal(t, "1", state.CurrentTrack.ID)
	assert.Equal(t, domain.StatusPlaying, state.Status)
}

func TestPlaylistService_RemoveTrack(t *testing.T) {
	service, _, _ := newTestPlaylistService()
	defer service.Shutdown()

	// Add some tracks
	tracks := []domain.MusicTrack{
		createTestTrack("1", "Song 1", "/test/song1.mp3"),
		createTestTrack("2", "Song 2", "/test/song2.mp3"),
		createTestTrack("3", "Song 3", "/test/song3.mp3"),
	}
	service.AddTracks(tracks, false)

	// Remove the middle track
	err := service.RemoveTrack(1)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, 2, service.GetQueueLength())
	queue := service.GetQueue()
	assert.Equal(t, "1", queue[0].ID)
	assert.Equal(t, "3", queue[1].ID)
}

func TestPlaylistService_RemoveTrack_CurrentlyPlaying(t *testing.T) {
	service, playback, _ := newTestPlaylistService()
	defer service.Shutdown()

	// Add and play tracks
	tracks := []domain.MusicTrack{
		createTestTrack("1", "Song 1", "/test/song1.mp3"),
		createTestTrack("2", "Song 2", "/test/song2.mp3"),
	}
	service.AddTracks(tracks, true)

	// Remove the currently playing track (index 0)
	err := service.RemoveTrack(0)
	require.NoError(t, err)

	// Verify playback stopped
	state := playback.GetState()
	assert.Nil(t, state.CurrentTrack)
	assert.Equal(t, -1, service.GetCurrentIndex())
}

func TestPlaylistService_ClearQueue(t *testing.T) {
	service, playback, bus := newTestPlaylistService()
	defer service.Shutdown()

	// Add tracks
	tracks := []domain.MusicTrack{
		createTestTrack("1", "Song 1", "/test/song1.mp3"),
		createTestTrack("2", "Song 2", "/test/song2.mp3"),
	}
	service.AddTracks(tracks, true)

	// Subscribe to queue changed event
	var queueEvent domain.QueueChangedEvent
	bus.Subscribe(domain.EventQueueChanged, func(e domain.Event) {
		queueEvent = e.(domain.QueueChangedEvent)
	})

	// Clear queue
	err := service.ClearQueue()
	require.NoError(t, err)

	// Verify
	assert.Equal(t, 0, service.GetQueueLength())
	assert.Equal(t, -1, service.GetCurrentIndex())
	assert.Equal(t, 0, len(queueEvent.Queue))

	// Verify playback stopped
	state := playback.GetState()
	assert.Nil(t, state.CurrentTrack)
}

func TestPlaylistService_PlayTrackAt(t *testing.T) {
	service, playback, _ := newTestPlaylistService()
	defer service.Shutdown()

	// Add tracks
	tracks := []domain.MusicTrack{
		createTestTrack("1", "Song 1", "/test/song1.mp3"),
		createTestTrack("2", "Song 2", "/test/song2.mp3"),
		createTestTrack("3", "Song 3", "/test/song3.mp3"),
	}
	service.AddTracks(tracks, false)

	// Play track at index 1
	err := service.PlayTrackAt(1)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, 1, service.GetCurrentIndex())
	state := playback.GetState()
	assert.NotNil(t, state.CurrentTrack)
	assert.Equal(t, "2", state.CurrentTrack.ID)
	assert.Equal(t, domain.StatusPlaying, state.Status)
}

func TestPlaylistService_PlayTrackAt_InvalidIndex(t *testing.T) {
	service, _, _ := newTestPlaylistService()
	defer service.Shutdown()

	// Try to play from an empty queue
	err := service.PlayTrackAt(0)
	assert.Equal(t, domain.ErrTrackNotFound, err)

	// Add tracks
	service.AddTrack(createTestTrack("1", "Song 1", "/test/song1.mp3"), false)

	// Try to play an invalid index
	err = service.PlayTrackAt(5)
	assert.Equal(t, domain.ErrTrackNotFound, err)

	err = service.PlayTrackAt(-1)
	assert.Equal(t, domain.ErrTrackNotFound, err)
}

func TestPlaylistService_PlayTrackByPath(t *testing.T) {
	service, playback, _ := newTestPlaylistService()
	defer service.Shutdown()

	// Add tracks
	tracks := []domain.MusicTrack{
		createTestTrack("1", "Song 1", "/test/song1.mp3"),
		createTestTrack("2", "Song 2", "/test/song2.mp3"),
		createTestTrack("3", "Song 3", "/test/song3.mp3"),
	}
	service.AddTracks(tracks, false)

	// Play by path
	index, err := service.PlayTrackByPath("/test/song3.mp3")
	require.NoError(t, err)

	// Verify
	assert.Equal(t, 2, index)
	assert.Equal(t, 2, service.GetCurrentIndex())
	state := playback.GetState()
	assert.NotNil(t, state.CurrentTrack)
	assert.Equal(t, "3", state.CurrentTrack.ID)
}

func TestPlaylistService_PlayTrackByPath_NotFound(t *testing.T) {
	service, _, _ := newTestPlaylistService()
	defer service.Shutdown()

	service.AddTrack(createTestTrack("1", "Song 1", "/test/song1.mp3"), false)

	// Try to play a non-existent path
	index, err := service.PlayTrackByPath("/test/nonexistent.mp3")
	assert.Equal(t, -1, index)
	assert.Equal(t, domain.ErrTrackNotFound, err)
}

func TestPlaylistService_PlayNext(t *testing.T) {
	service, playback, _ := newTestPlaylistService()
	defer service.Shutdown()

	// Add tracks
	tracks := []domain.MusicTrack{
		createTestTrack("1", "Song 1", "/test/song1.mp3"),
		createTestTrack("2", "Song 2", "/test/song2.mp3"),
		createTestTrack("3", "Song 3", "/test/song3.mp3"),
	}
	service.AddTracks(tracks, true) // Start playing first

	// Play next
	err := service.PlayNext()
	require.NoError(t, err)

	// Verify
	assert.Equal(t, 1, service.GetCurrentIndex())
	state := playback.GetState()
	assert.Equal(t, "2", state.CurrentTrack.ID)
}

func TestPlaylistService_PlayNext_EndOfQueue(t *testing.T) {
	service, _, _ := newTestPlaylistService()
	defer service.Shutdown()

	// Add track and play
	service.AddTrack(createTestTrack("1", "Song 1", "/test/song1.mp3"), true)

	// Try to play next (at the end of queue)
	err := service.PlayNext()
	assert.Equal(t, domain.ErrEndOfQueue, err)
}

func TestPlaylistService_PlayNext_EmptyQueue(t *testing.T) {
	service, _, _ := newTestPlaylistService()
	defer service.Shutdown()

	// Try to play next on the empty queue
	err := service.PlayNext()
	assert.Equal(t, domain.ErrQueueEmpty, err)
}

func TestPlaylistService_PlayPrevious(t *testing.T) {
	service, playback, _ := newTestPlaylistService()
	defer service.Shutdown()

	// Add tracks and play the second one
	tracks := []domain.MusicTrack{
		createTestTrack("1", "Song 1", "/test/song1.mp3"),
		createTestTrack("2", "Song 2", "/test/song2.mp3"),
		createTestTrack("3", "Song 3", "/test/song3.mp3"),
	}
	service.AddTracks(tracks, false)
	service.PlayTrackAt(1)

	// Play previous
	err := service.PlayPrevious()
	require.NoError(t, err)

	// Verify
	assert.Equal(t, 0, service.GetCurrentIndex())
	state := playback.GetState()
	assert.Equal(t, "1", state.CurrentTrack.ID)
}

func TestPlaylistService_PlayPrevious_StartOfQueue(t *testing.T) {
	service, _, _ := newTestPlaylistService()
	defer service.Shutdown()

	// Add track and play
	service.AddTrack(createTestTrack("1", "Song 1", "/test/song1.mp3"), true)

	// Try to play previous (at start of queue)
	err := service.PlayPrevious()
	assert.Equal(t, domain.ErrStartOfQueue, err)
}

func TestPlaylistService_GetQueue(t *testing.T) {
	service, _, _ := newTestPlaylistService()
	defer service.Shutdown()

	// Add tracks
	tracks := []domain.MusicTrack{
		createTestTrack("1", "Song 1", "/test/song1.mp3"),
		createTestTrack("2", "Song 2", "/test/song2.mp3"),
	}
	service.AddTracks(tracks, false)

	// Get queue
	queue := service.GetQueue()
	assert.Equal(t, 2, len(queue))
	assert.Equal(t, "1", queue[0].ID)
	assert.Equal(t, "2", queue[1].ID)

	// Verify it's a copy (modifying doesn't affect internal state)
	queue[0].Title = "Modified"
	originalQueue := service.GetQueue()
	assert.NotEqual(t, "Modified", originalQueue[0].Title)
}

func TestPlaylistService_SaveAndLoadQueue(t *testing.T) {
	engine := mock.NewEngine()
	engine.Initialize(-1, 44100, 0)

	bus := eventbus.NewSyncEventBus()
	playback := NewPlaybackService(engine, bus)
	plRepo := newMockPlaylistRepository()
	histRepo := newMockHistoryRepository()

	// First service instance
	service := NewPlaylistService(playback, plRepo, histRepo, bus)

	// Add tracks and play one
	tracks := []domain.MusicTrack{
		createTestTrack("1", "Song 1", "/test/song1.mp3"),
		createTestTrack("2", "Song 2", "/test/song2.mp3"),
	}
	service.AddTracks(tracks, false)
	service.PlayTrackAt(1)

	// Save queue
	err := service.SaveQueue()
	require.NoError(t, err)

	// Shutdown first service
	service.Shutdown()

	// Create a new service instance with SAME repositories
	playback2 := NewPlaybackService(engine, bus)
	service2 := NewPlaylistService(playback2, plRepo, histRepo, bus)
	defer service2.Shutdown()

	err = service2.LoadQueue()
	require.NoError(t, err)

	// Verify loaded state
	assert.Equal(t, 2, service2.GetQueueLength())
	queue := service2.GetQueue()
	assert.Equal(t, "1", queue[0].ID)
	assert.Equal(t, "2", queue[1].ID)
	assert.Equal(t, 1, service2.GetCurrentIndex())
}

func TestPlaylistService_MoveTrack(t *testing.T) {
	service, _, _ := newTestPlaylistService()
	defer service.Shutdown()

	// Add tracks
	tracks := []domain.MusicTrack{
		createTestTrack("1", "Song 1", "/test/song1.mp3"),
		createTestTrack("2", "Song 2", "/test/song2.mp3"),
		createTestTrack("3", "Song 3", "/test/song3.mp3"),
	}
	service.AddTracks(tracks, false)

	// Move track from index 0 to index 2
	err := service.MoveTrack(0, 2)
	require.NoError(t, err)

	// Verify order
	queue := service.GetQueue()
	assert.Equal(t, "2", queue[0].ID)
	assert.Equal(t, "3", queue[1].ID)
	assert.Equal(t, "1", queue[2].ID)
}

func TestPlaylistService_MoveTrack_InvalidIndices(t *testing.T) {
	service, _, _ := newTestPlaylistService()
	defer service.Shutdown()

	service.AddTrack(createTestTrack("1", "Song 1", "/test/song1.mp3"), false)

	// Invalid from index
	err := service.MoveTrack(-1, 0)
	assert.Equal(t, domain.ErrTrackNotFound, err)

	// Invalid to index
	err = service.MoveTrack(0, 5)
	assert.Equal(t, domain.ErrTrackNotFound, err)
}

func TestPlaylistService_AutoNext(t *testing.T) {
	service, playback, bus := newTestPlaylistService()
	defer service.Shutdown()

	// Add tracks
	tracks := []domain.MusicTrack{
		createTestTrack("1", "Song 1", "/test/song1.mp3"),
		createTestTrack("2", "Song 2", "/test/song2.mp3"),
	}
	service.AddTracks(tracks, true) // Starts playing index 0

	// Initial state
	assert.Equal(t, 0, service.GetCurrentIndex())

	// Simulate auto-next event (track finished playing)
	bus.Publish(domain.NewAutoNextEvent(tracks[0], 0))

	// Give the handler time to process the event and start the next track
	// The handler advances to the next track and loads/plays it
	// Note: In real scenario, this would be triggered by playback service
	// when track finishes playing

	// Verify current index advanced to next track
	assert.Equal(t, 1, service.GetCurrentIndex())

	// Verify playback loaded the next track
	state := playback.GetState()
	if state.CurrentTrack != nil {
		assert.Equal(t, "2", state.CurrentTrack.ID)
	}
}

// Thread safety tests

func TestPlaylistService_ConcurrentAddTracks(t *testing.T) {
	service, _, _ := newTestPlaylistService()
	defer service.Shutdown()

	// Add tracks concurrently
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func(index int) {
			track := createTestTrack(string(rune('0'+index)), "Song", "/test/song.mp3")
			service.AddTrack(track, false)
			done <- struct{}{}
		}(i)
	}

	// Wait for all
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 10 tracks
	assert.Equal(t, 10, service.GetQueueLength())
}

func TestPlaylistService_ConcurrentRemove(t *testing.T) {
	service, _, _ := newTestPlaylistService()
	defer service.Shutdown()

	// Add many tracks
	for i := 0; i < 20; i++ {
		track := createTestTrack(string(rune('0'+i)), "Song", "/test/song.mp3")
		service.AddTrack(track, false)
	}

	// Remove tracks concurrently
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			service.RemoveTrack(0) // Always remove first
			done <- struct{}{}
		}()
	}

	// Wait for all
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 10 tracks remaining
	assert.Equal(t, 10, service.GetQueueLength())
}

func TestPlaylistService_Shutdown(t *testing.T) {
	service, _, _ := newTestPlaylistService()

	// Add tracks
	tracks := []domain.MusicTrack{
		createTestTrack("1", "Song 1", "/test/song1.mp3"),
		createTestTrack("2", "Song 2", "/test/song2.mp3"),
	}
	service.AddTracks(tracks, false)

	// Shutdown
	err := service.Shutdown()
	assert.NoError(t, err)
}
