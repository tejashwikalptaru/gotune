// Package service provides business logic for the GoTune application.
package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/tejashwikalptaru/gotune/internal/domain"
	"github.com/tejashwikalptaru/gotune/internal/ports"
)

// PlaybackService orchestrates audio playback operations.
// It manages the current playing track, volume, mute state, and loop mode.
// All operations are thread-safe via sync.RWMutex.
type PlaybackService struct {
	// Dependencies (injected)
	engine ports.AudioEngine
	bus    ports.EventBus

	// State
	currentTrack   *domain.MusicTrack
	currentHandle  domain.TrackHandle
	currentIndex   int // Index in the playlist (managed by PlaylistService)
	volume         float64
	savedVolume    float64 // Volume before mute
	isMuted        bool
	isLooping      bool
	updateInterval time.Duration

	// Concurrency control
	mu            sync.RWMutex
	stopUpdate    chan struct{}
	updateRunning bool
	manualStop    bool // True if the user explicitly stopped playback
}

// NewPlaybackService creates a new playback service.
func NewPlaybackService(
	engine ports.AudioEngine,
	bus ports.EventBus,
) *PlaybackService {
	service := &PlaybackService{
		engine:         engine,
		bus:            bus,
		currentHandle:  domain.InvalidTrackHandle,
		currentIndex:   -1,
		volume:         0.8,                    // Default 80% volume
		updateInterval: 333 * time.Millisecond, // 3 times per second
		stopUpdate:     make(chan struct{}),
	}

	// Start update routine
	service.startUpdateRoutine()

	return service
}

// LoadTrack loads a track for playback.
// This stops any currently playing track and loads the new one.
func (s *PlaybackService) LoadTrack(track domain.MusicTrack, index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Printf("DEBUG: LoadTrack() called for: %s\n", track.FilePath)

	// Stop the current track if any
	if s.currentHandle != domain.InvalidTrackHandle {
		fmt.Println("DEBUG: Stopping current track")
		s.stopInternal()
	}

	// Load new track
	handle, err := s.engine.Load(track.FilePath)
	if err != nil {
		fmt.Printf("DEBUG: LoadTrack() failed - engine.Load error: %v\n", err)
		s.bus.Publish(domain.NewTrackErrorEvent(track, err))
		return err
	}

	fmt.Printf("DEBUG: Loaded track with handle %d\n", handle)

	// Set volume on a new track
	if err := s.engine.SetVolume(handle, s.volume); err != nil {
		s.engine.Unload(handle)
		return err
	}

	// Get duration
	duration, err := s.engine.Duration(handle)
	if err != nil {
		s.engine.Unload(handle)
		return err
	}

	// Update state
	s.currentTrack = &track
	s.currentHandle = handle
	s.currentIndex = index
	s.manualStop = false

	fmt.Printf("DEBUG: LoadTrack() succeeded, currentHandle set to %d\n", s.currentHandle)

	// Publish event
	s.bus.Publish(domain.NewTrackLoadedEvent(track, handle, duration, index))

	return nil
}

// Play starts or resumes playback of the current track.
func (s *PlaybackService) Play() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Println("DEBUG: Play() called")

	if s.currentHandle == domain.InvalidTrackHandle {
		fmt.Println("DEBUG: Play() failed - invalid track handle")
		return domain.ErrInvalidTrackHandle
	}

	fmt.Printf("DEBUG: Play() attempting with handle %d\n", s.currentHandle)

	// Check the current status
	status, err := s.engine.Status(s.currentHandle)
	if err != nil {
		fmt.Printf("DEBUG: Play() failed - status check error: %v\n", err)
		return err
	}

	fmt.Printf("DEBUG: Current status: %v\n", status)

	// Already playing
	if status == domain.StatusPlaying {
		fmt.Println("DEBUG: Already playing, returning")
		return nil
	}

	// Start/resume playback
	s.manualStop = false
	if err := s.engine.Play(s.currentHandle); err != nil {
		fmt.Printf("DEBUG: Play() failed - engine.Play error: %v\n", err)
		return err
	}

	fmt.Println("DEBUG: Play() succeeded, publishing event")

	// Publish event
	if s.currentTrack != nil {
		s.bus.Publish(domain.NewTrackStartedEvent(*s.currentTrack))
	}

	return nil
}

// Pause pauses playback of the current track.
func (s *PlaybackService) Pause() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentHandle == domain.InvalidTrackHandle {
		return domain.ErrInvalidTrackHandle
	}

	// Get the current position before pausing
	position, _ := s.engine.Position(s.currentHandle)

	if err := s.engine.Pause(s.currentHandle); err != nil {
		return err
	}

	// Publish event
	if s.currentTrack != nil {
		s.bus.Publish(domain.NewTrackPausedEvent(*s.currentTrack, position))
	}

	return nil
}

// Stop stops playback and unloads the current track.
func (s *PlaybackService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.stopInternal()
}

// stopInternal stops playback without locking (caller must hold lock).
func (s *PlaybackService) stopInternal() error {
	if s.currentHandle == domain.InvalidTrackHandle {
		return nil
	}

	s.manualStop = true

	// Stop the track
	if err := s.engine.Stop(s.currentHandle); err != nil {
		// Even if stop fails, clear our state
		s.currentHandle = domain.InvalidTrackHandle
		s.currentTrack = nil
		return err
	}

	// Publish event before clearing state
	if s.currentTrack != nil {
		s.bus.Publish(domain.NewTrackStoppedEvent(*s.currentTrack))
	}

	// Clear state
	s.currentHandle = domain.InvalidTrackHandle
	s.currentTrack = nil

	return nil
}

// SetVolume sets the playback volume (0.0 to 1.0).
func (s *PlaybackService) SetVolume(volume float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if volume < 0.0 || volume > 1.0 {
		return domain.ErrInvalidVolume
	}

	s.volume = volume

	// If muted, save the volume but don't apply it
	if s.isMuted {
		s.savedVolume = volume
		s.bus.Publish(domain.NewVolumeChangedEvent(volume))
		return nil
	}

	// Apply volume to the current track if any
	if s.currentHandle != domain.InvalidTrackHandle {
		if err := s.engine.SetVolume(s.currentHandle, volume); err != nil {
			return err
		}
	}

	// Publish event
	s.bus.Publish(domain.NewVolumeChangedEvent(volume))

	return nil
}

// GetVolume returns the current volume (0.0 to 1.0).
func (s *PlaybackService) GetVolume() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.volume
}

// Mute mutes or unmutes playback.
func (s *PlaybackService) Mute(mute bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isMuted == mute {
		return nil // Already in the desired state
	}

	s.isMuted = mute

	// Apply mute/unmute to the current track
	if s.currentHandle != domain.InvalidTrackHandle {
		targetVolume := s.volume
		if mute {
			s.savedVolume = s.volume
			targetVolume = 0.0
		}

		if err := s.engine.SetVolume(s.currentHandle, targetVolume); err != nil {
			return err
		}
	}

	// Publish event
	s.bus.Publish(domain.NewMuteToggledEvent(s.isMuted))

	return nil
}

// IsMuted returns true if playback is muted.
func (s *PlaybackService) IsMuted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.isMuted
}

// SetLoop enables or disables loop mode.
// When enabled, the current track will restart when it finishes.
func (s *PlaybackService) SetLoop(loop bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isLooping == loop {
		return
	}

	s.isLooping = loop

	// Publish event
	s.bus.Publish(domain.NewLoopToggledEvent(loop))
}

// IsLooping returns true if loop mode is enabled.
func (s *PlaybackService) IsLooping() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.isLooping
}

// Seek sets the playback position.
func (s *PlaybackService) Seek(position time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentHandle == domain.InvalidTrackHandle {
		return domain.ErrInvalidTrackHandle
	}

	if err := s.engine.Seek(s.currentHandle, position); err != nil {
		return err
	}

	// Publish progress event with new position
	if s.currentTrack != nil {
		duration, _ := s.engine.Duration(s.currentHandle)
		s.bus.Publish(domain.NewTrackProgressEvent(position, duration))
	}

	return nil
}

// GetState returns the current playback state.
func (s *PlaybackService) GetState() domain.PlaybackState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state := domain.PlaybackState{
		CurrentIndex: s.currentIndex,
		Volume:       s.volume,
		IsMuted:      s.isMuted,
		IsLooping:    s.isLooping,
	}

	// Get current track info
	if s.currentTrack != nil {
		state.CurrentTrack = s.currentTrack
	}

	// Get status and position if the track is loaded
	if s.currentHandle != domain.InvalidTrackHandle {
		if status, err := s.engine.Status(s.currentHandle); err == nil {
			state.Status = status
		}

		if position, err := s.engine.Position(s.currentHandle); err == nil {
			state.Position = position
		}
	} else {
		state.Status = domain.StatusStopped
	}

	return state
}

// Shutdown stops playback and cleans up resources.
func (s *PlaybackService) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop update routine
	if s.updateRunning {
		close(s.stopUpdate)
		s.updateRunning = false
	}

	// Stop the current track
	return s.stopInternal()
}

// startUpdateRoutine starts a goroutine that periodically publishes progress events.
func (s *PlaybackService) startUpdateRoutine() {
	s.mu.Lock()
	if s.updateRunning {
		s.mu.Unlock()
		return
	}
	s.updateRunning = true
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(s.updateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopUpdate:
				return

			case <-ticker.C:
				s.publishProgressUpdate()
			}
		}
	}()
}

// publishProgressUpdate publishes a progress event if a track is playing.
func (s *PlaybackService) publishProgressUpdate() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Nothing to update if no track loaded
	if s.currentHandle == domain.InvalidTrackHandle || s.currentTrack == nil {
		return
	}

	// Get current status
	status, err := s.engine.Status(s.currentHandle)
	if err != nil {
		return
	}

	// Get position and duration
	position, err := s.engine.Position(s.currentHandle)
	if err != nil {
		return
	}

	duration, err := s.engine.Duration(s.currentHandle)
	if err != nil {
		return
	}

	// Publish progress event
	s.bus.Publish(domain.NewTrackProgressEvent(position, duration))

	// Handle track finished (only if not manually stopped)
	if status == domain.StatusStopped && !s.manualStop {
		// s.handleTrackFinished()
	}
}

// handleTrackFinished is called when a track finishes playing naturally.
func (s *PlaybackService) handleTrackFinished() {
	// This is called with read lock held from publishProgressUpdate
	// We need to handle this carefully

	if s.currentTrack == nil {
		return
	}

	track := *s.currentTrack

	// Publish completed event
	s.bus.Publish(domain.NewTrackCompletedEvent(track))

	// If looping, restart the current track
	if s.isLooping {
		// Unlock, reload, and play
		s.mu.RUnlock()
		s.mu.Lock()

		// Reload and play (without changing the index)
		if s.currentTrack != nil {
			s.stopInternal()
			s.LoadTrack(track, s.currentIndex)
			s.Play()
		}

		s.mu.Unlock()
		s.mu.RLock()
	} else {
		// Otherwise, publish the auto-next event (PlaylistService will handle)
		s.bus.Publish(domain.NewAutoNextEvent(track, s.currentIndex))
	}
}

// Verify that PlaybackService implements the expected interface patterns
var _ interface {
	LoadTrack(domain.MusicTrack, int) error
	Play() error
	Pause() error
	Stop() error
	SetVolume(float64) error
	GetVolume() float64
	Mute(bool) error
	IsMuted() bool
	SetLoop(bool)
	IsLooping() bool
	Seek(time.Duration) error
	GetState() domain.PlaybackState
	Shutdown() error
} = (*PlaybackService)(nil)
