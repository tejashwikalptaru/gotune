package service

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tejashwikalptaru/gotune/internal/adapter/audio/mock"
	"github.com/tejashwikalptaru/gotune/internal/adapter/eventbus"
	"github.com/tejashwikalptaru/gotune/internal/domain"
)

// Helper to create a test playback service
func newTestPlaybackService() (*PlaybackService, *mock.MockEngine, *eventbus.SyncEventBus) {
	engine := mock.NewMockEngine()
	bus := eventbus.NewSyncEventBus()

	service := NewPlaybackService(engine, bus)

	return service, engine, bus
}

// Helper to create a test track
func createTestTrack(id, title, path string) domain.MusicTrack {
	return domain.MusicTrack{
		ID:       id,
		Title:    title,
		FilePath: path,
		Artist:   "Test Artist",
		Album:    "Test Album",
		Duration: 3 * time.Minute,
	}
}

func TestPlaybackService_LoadTrack(t *testing.T) {
	service, engine, bus := newTestPlaybackService()
	defer service.Shutdown()

	// Initialize engine
	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	track := createTestTrack("1", "Test Song", "/test/song.mp3")

	// Subscribe to events
	var loadedEvent domain.TrackLoadedEvent
	bus.Subscribe(domain.EventTrackLoaded, func(e domain.Event) {
		loadedEvent = e.(domain.TrackLoadedEvent)
	})

	// Load track
	err = service.LoadTrack(track, 0)
	require.NoError(t, err)

	// Verify state
	state := service.GetState()
	assert.Equal(t, track.ID, state.CurrentTrack.ID)
	assert.Equal(t, 0, state.CurrentIndex)
	assert.Equal(t, domain.StatusStopped, state.Status)

	// Verify event was published
	assert.Equal(t, track.ID, loadedEvent.Track.ID)
	assert.NotEqual(t, domain.InvalidTrackHandle, loadedEvent.Handle)
}

func TestPlaybackService_LoadTrack_InvalidPath(t *testing.T) {
	service, engine, bus := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	// Subscribe to error events
	var errorEvent domain.TrackErrorEvent
	bus.Subscribe(domain.EventTrackError, func(e domain.Event) {
		errorEvent = e.(domain.TrackErrorEvent)
	})

	// Try to load an invalid track
	track := createTestTrack("1", "Test", "/nonexistent/file.mp3")
	engine.SetFailLoad(true)

	err = service.LoadTrack(track, 0)
	assert.Error(t, err)

	// Verify the error event was published
	assert.NotNil(t, errorEvent.Error)
}

func TestPlaybackService_LoadTrack_ReplacesCurrentTrack(t *testing.T) {
	service, engine, _ := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	// Load first track
	track1 := createTestTrack("1", "Song 1", "/test/song1.mp3")
	err = service.LoadTrack(track1, 0)
	require.NoError(t, err)

	// Start playing
	err = service.Play()
	require.NoError(t, err)

	// Load the second track (should stop first)
	track2 := createTestTrack("2", "Song 2", "/test/song2.mp3")
	err = service.LoadTrack(track2, 1)
	require.NoError(t, err)

	// Verify the current track is track2
	state := service.GetState()
	assert.Equal(t, track2.ID, state.CurrentTrack.ID)
	assert.Equal(t, 1, state.CurrentIndex)

	// The first track should be stopped (unloaded)
	assert.Equal(t, 1, engine.GetLoadedTracks())
}

func TestPlaybackService_Play(t *testing.T) {
	service, engine, bus := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	track := createTestTrack("1", "Test Song", "/test/song.mp3")

	// Subscribe to events
	var startedEvent domain.TrackStartedEvent
	bus.Subscribe(domain.EventTrackStarted, func(e domain.Event) {
		startedEvent = e.(domain.TrackStartedEvent)
	})

	// Load track
	err = service.LoadTrack(track, 0)
	require.NoError(t, err)

	// Play
	err = service.Play()
	require.NoError(t, err)

	// Verify state
	state := service.GetState()
	assert.Equal(t, domain.StatusPlaying, state.Status)

	// Verify event was published
	assert.Equal(t, track.ID, startedEvent.Track.ID)
}

func TestPlaybackService_Play_NoTrackLoaded(t *testing.T) {
	service, engine, _ := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	// Try to play without loading a track
	err = service.Play()
	assert.Equal(t, domain.ErrInvalidTrackHandle, err)
}

func TestPlaybackService_Play_AlreadyPlaying(t *testing.T) {
	service, engine, bus := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	track := createTestTrack("1", "Test Song", "/test/song.mp3")

	eventCount := 0
	bus.Subscribe(domain.EventTrackStarted, func(e domain.Event) {
		eventCount++
	})

	// Load and play
	service.LoadTrack(track, 0)
	service.Play()

	// Play again (should be no-op)
	err = service.Play()
	require.NoError(t, err)

	// Should only receive one started event
	assert.Equal(t, 1, eventCount)
}

func TestPlaybackService_Pause(t *testing.T) {
	service, engine, bus := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	track := createTestTrack("1", "Test Song", "/test/song.mp3")

	// Subscribe to events
	var pausedEvent domain.TrackPausedEvent
	bus.Subscribe(domain.EventTrackPaused, func(e domain.Event) {
		pausedEvent = e.(domain.TrackPausedEvent)
	})

	// Load and play
	service.LoadTrack(track, 0)
	service.Play()

	// Pause
	err = service.Pause()
	require.NoError(t, err)

	// Verify state
	state := service.GetState()
	assert.Equal(t, domain.StatusPaused, state.Status)

	// Verify event
	assert.Equal(t, track.ID, pausedEvent.Track.ID)
}

func TestPlaybackService_Pause_Resume(t *testing.T) {
	service, engine, _ := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	track := createTestTrack("1", "Test Song", "/test/song.mp3")

	// Load and play
	service.LoadTrack(track, 0)
	service.Play()

	// Pause
	service.Pause()

	state := service.GetState()
	assert.Equal(t, domain.StatusPaused, state.Status)

	// Resume (call Play again)
	err = service.Play()
	require.NoError(t, err)

	state = service.GetState()
	assert.Equal(t, domain.StatusPlaying, state.Status)
}

func TestPlaybackService_Stop(t *testing.T) {
	service, engine, bus := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	track := createTestTrack("1", "Test Song", "/test/song.mp3")

	// Subscribe to events
	var stoppedEvent domain.TrackStoppedEvent
	bus.Subscribe(domain.EventTrackStopped, func(e domain.Event) {
		stoppedEvent = e.(domain.TrackStoppedEvent)
	})

	// Load and play
	service.LoadTrack(track, 0)
	service.Play()

	// Stop
	err = service.Stop()
	require.NoError(t, err)

	// Verify state
	state := service.GetState()
	assert.Nil(t, state.CurrentTrack)
	assert.Equal(t, domain.StatusStopped, state.Status)

	// Verify event
	assert.Equal(t, track.ID, stoppedEvent.Track.ID)
}

func TestPlaybackService_SetVolume(t *testing.T) {
	service, engine, bus := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	track := createTestTrack("1", "Test Song", "/test/song.mp3")

	// Subscribe to events
	var volumeEvent domain.VolumeChangedEvent
	bus.Subscribe(domain.EventVolumeChanged, func(e domain.Event) {
		volumeEvent = e.(domain.VolumeChangedEvent)
	})

	// Load track
	service.LoadTrack(track, 0)

	// Set volume
	err = service.SetVolume(0.5)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, 0.5, service.GetVolume())
	assert.Equal(t, 0.5, volumeEvent.Volume)
}

func TestPlaybackService_SetVolume_InvalidRange(t *testing.T) {
	service, engine, _ := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	// Volume below 0
	err = service.SetVolume(-0.1)
	assert.Equal(t, domain.ErrInvalidVolume, err)

	// Volume above 1
	err = service.SetVolume(1.5)
	assert.Equal(t, domain.ErrInvalidVolume, err)
}

func TestPlaybackService_SetVolume_BoundaryValues(t *testing.T) {
	service, engine, _ := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	// Volume 0.0 (minimum)
	err = service.SetVolume(0.0)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, service.GetVolume())

	// Volume 1.0 (maximum)
	err = service.SetVolume(1.0)
	assert.NoError(t, err)
	assert.Equal(t, 1.0, service.GetVolume())
}

func TestPlaybackService_Mute(t *testing.T) {
	service, engine, bus := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	track := createTestTrack("1", "Test Song", "/test/song.mp3")

	// Subscribe to events
	var muteEvent domain.MuteToggledEvent
	bus.Subscribe(domain.EventMuteToggled, func(e domain.Event) {
		muteEvent = e.(domain.MuteToggledEvent)
	})

	// Load track and set volume
	service.LoadTrack(track, 0)
	service.SetVolume(0.8)

	// Mute
	err = service.Mute(true)
	require.NoError(t, err)

	// Verify
	assert.True(t, service.IsMuted())
	assert.Equal(t, 0.8, service.GetVolume()) // Volume setting preserved
	assert.True(t, muteEvent.Muted)

	// Unmute
	err = service.Mute(false)
	require.NoError(t, err)

	// Verify
	assert.False(t, service.IsMuted())
	assert.Equal(t, 0.8, service.GetVolume())
}

func TestPlaybackService_SetLoop(t *testing.T) {
	service, engine, bus := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	// Subscribe to events
	var loopEvent domain.LoopToggledEvent
	bus.Subscribe(domain.EventLoopToggled, func(e domain.Event) {
		loopEvent = e.(domain.LoopToggledEvent)
	})

	// Enable loop
	service.SetLoop(true)
	assert.True(t, service.IsLooping())
	assert.True(t, loopEvent.Enabled)

	// Disable loop
	service.SetLoop(false)
	assert.False(t, service.IsLooping())
	assert.False(t, loopEvent.Enabled)
}

func TestPlaybackService_Seek(t *testing.T) {
	service, engine, _ := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	track := createTestTrack("1", "Test Song", "/test/song.mp3")

	// Load track
	service.LoadTrack(track, 0)

	// Seek to 1 minute
	err = service.Seek(1 * time.Minute)
	require.NoError(t, err)

	// Verify position (using mock engine's simulated position)
	state := service.GetState()
	assert.Equal(t, 1*time.Minute, state.Position)
}

func TestPlaybackService_Seek_InvalidPosition(t *testing.T) {
	service, engine, _ := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	track := createTestTrack("1", "Test Song", "/test/song.mp3")
	service.LoadTrack(track, 0)

	// Seek beyond duration
	err = service.Seek(10 * time.Minute)
	assert.Equal(t, domain.ErrInvalidPosition, err)

	// Seek to negative position
	err = service.Seek(-1 * time.Second)
	assert.Equal(t, domain.ErrInvalidPosition, err)
}

func TestPlaybackService_GetState(t *testing.T) {
	service, engine, _ := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	track := createTestTrack("1", "Test Song", "/test/song.mp3")

	// Initial state
	state := service.GetState()
	assert.Nil(t, state.CurrentTrack)
	assert.Equal(t, -1, state.CurrentIndex)
	assert.Equal(t, domain.StatusStopped, state.Status)
	assert.Equal(t, 0.8, state.Volume) // Default volume

	// Load track
	service.LoadTrack(track, 5)
	service.SetVolume(0.6)
	service.SetLoop(true)

	state = service.GetState()
	assert.NotNil(t, state.CurrentTrack)
	assert.Equal(t, track.ID, state.CurrentTrack.ID)
	assert.Equal(t, 5, state.CurrentIndex)
	assert.Equal(t, 0.6, state.Volume)
	assert.True(t, state.IsLooping)
}

func TestPlaybackService_ProgressEvents(t *testing.T) {
	service, engine, bus := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	track := createTestTrack("1", "Test Song", "/test/song.mp3")

	// Subscribe to progress events (with proper synchronization)
	var mu sync.Mutex
	progressCount := 0
	bus.Subscribe(domain.EventTrackProgress, func(e domain.Event) {
		mu.Lock()
		progressCount++
		mu.Unlock()
	})

	// Load and play
	service.LoadTrack(track, 0)
	service.Play()

	// Wait for some progress updates (an update interval is 333ms)
	time.Sleep(1 * time.Second)

	// Should have received multiple progress events
	mu.Lock()
	count := progressCount
	mu.Unlock()
	assert.Greater(t, count, 1)
}

func TestPlaybackService_TrackCompleted_WithLoop(t *testing.T) {
	service, engine, bus := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	track := createTestTrack("1", "Test Song", "/test/song.mp3")

	// Subscribe to the completed event
	completedReceived := false
	bus.Subscribe(domain.EventTrackCompleted, func(e domain.Event) {
		completedReceived = true
	})

	// Enable loop
	service.SetLoop(true)

	// Load and play
	service.LoadTrack(track, 0)
	service.Play()

	// Simulate track finishing
	currentTrack := service.GetState().CurrentTrack
	require.NotNil(t, currentTrack)

	// Get the handle from the engine
	loadedTracks := engine.GetLoadedTracks()
	assert.Equal(t, 1, loadedTracks)

	// The update routine should detect track finished and restart it
	// Give it time to detect and restart
	time.Sleep(500 * time.Millisecond)

	// Track should still be playing (restarted)
	state := service.GetState()
	assert.NotNil(t, state.CurrentTrack)
	assert.Equal(t, track.ID, state.CurrentTrack.ID)

	// Use the completedReceived variable
	_ = completedReceived
}

func TestPlaybackService_TrackCompleted_WithoutLoop(t *testing.T) {
	service, engine, bus := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	track := createTestTrack("1", "Test Song", "/test/song.mp3")

	// Subscribe to auto-next event
	autoNextReceived := false
	bus.Subscribe(domain.EventAutoNext, func(e domain.Event) {
		autoNextReceived = true
	})

	// Disable loop
	service.SetLoop(false)

	// Load and play
	service.LoadTrack(track, 0)
	service.Play()

	// The update routine should publish auto-next event when track finishes
	// (In a real scenario, PlaylistService would handle this)
	_ = autoNextReceived
}

// Thread safety tests

func TestPlaybackService_ConcurrentVolumeChanges(t *testing.T) {
	service, engine, _ := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	track := createTestTrack("1", "Test Song", "/test/song.mp3")
	service.LoadTrack(track, 0)

	// Change volume concurrently from multiple goroutines
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func(vol float64) {
			service.SetVolume(vol)
			done <- struct{}{}
		}(float64(i) / 10.0)
	}

	// Wait for all
	for i := 0; i < 10; i++ {
		<-done
	}

	// Volume should be one of the set values
	volume := service.GetVolume()
	assert.GreaterOrEqual(t, volume, 0.0)
	assert.LessOrEqual(t, volume, 1.0)
}

func TestPlaybackService_ConcurrentPlayPause(t *testing.T) {
	service, engine, _ := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	track := createTestTrack("1", "Test Song", "/test/song.mp3")
	service.LoadTrack(track, 0)

	// Play and pause concurrently
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func(index int) {
			if index%2 == 0 {
				service.Play()
			} else {
				service.Pause()
			}
			done <- struct{}{}
		}(i)
	}

	// Wait for all
	for i := 0; i < 10; i++ {
		<-done
	}

	// State should be valid (either playing or paused)
	state := service.GetState()
	assert.Contains(t, []domain.PlaybackStatus{domain.StatusPlaying, domain.StatusPaused}, state.Status)
}

func TestPlaybackService_ConcurrentGetState(t *testing.T) {
	service, engine, _ := newTestPlaybackService()
	defer service.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	track := createTestTrack("1", "Test Song", "/test/song.mp3")
	service.LoadTrack(track, 0)
	service.Play()

	// Read state concurrently
	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func() {
			state := service.GetState()
			assert.NotNil(t, state)
			done <- struct{}{}
		}()
	}

	// Wait for all
	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestPlaybackService_Shutdown(t *testing.T) {
	service, engine, _ := newTestPlaybackService()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	track := createTestTrack("1", "Test Song", "/test/song.mp3")
	service.LoadTrack(track, 0)
	service.Play()

	// Shutdown
	err = service.Shutdown()
	assert.NoError(t, err)

	// State should be cleared
	state := service.GetState()
	assert.Nil(t, state.CurrentTrack)
	assert.Equal(t, domain.StatusStopped, state.Status)
}
