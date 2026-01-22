// Package service provides business logic for the GoTune application.
package service

import (
	"log/slog"
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
	logger *slog.Logger
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
	updateWg      sync.WaitGroup // WaitGroup to wait for update goroutine to exit
	manualStop    bool           // True if the user explicitly stopped playback
	hasPlayed     bool           // True if the current track has been played
}

// NewPlaybackService creates a new playback service.
func NewPlaybackService(
	logger *slog.Logger,
	engine ports.AudioEngine,
	bus ports.EventBus,
) *PlaybackService {
	service := &PlaybackService{
		logger:         logger,
		engine:         engine,
		bus:            bus,
		currentHandle:  domain.InvalidTrackHandle,
		currentIndex:   -1,
		volume:         0.8,                    // Default 80% volume
		updateInterval: 333 * time.Millisecond, // 3 times per second
		stopUpdate:     make(chan struct{}),
	}

	logger.Debug("playback service initialized")

	// Start update routine
	service.startUpdateRoutine()

	return service
}

// LoadTrack loads a track for playback.
// This stops any currently playing track and loads the new one.
func (s *PlaybackService) LoadTrack(track domain.MusicTrack, index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Debug("loading track", slog.String("file_path", track.FilePath))

	// Stop the current track if any
	if s.currentHandle != domain.InvalidTrackHandle {
		s.logger.Debug("stopping current track")
		if err := s.stopInternal(); err != nil {
			s.logger.Warn("failed to stop current track", slog.Any("error", err))
		}
	}

	// Load new track
	handle, err := s.engine.Load(track.FilePath)
	if err != nil {
		s.logger.Debug("failed to load track", slog.Any("error", err))
		s.bus.Publish(domain.NewTrackErrorEvent(track, err))
		return err
	}

	s.logger.Debug("track loaded successfully", slog.Int64("handle", int64(handle)))

	// Set volume on a new track
	if err := s.engine.SetVolume(handle, s.volume); err != nil {
		if unloadErr := s.engine.Unload(handle); unloadErr != nil {
			s.logger.Warn("failed to unload track after volume error", slog.Any("error", unloadErr))
		}
		return err
	}

	// Get duration
	duration, err := s.engine.Duration(handle)
	if err != nil {
		if unloadErr := s.engine.Unload(handle); unloadErr != nil {
			s.logger.Warn("failed to unload track after duration error", slog.Any("error", unloadErr))
		}
		return err
	}

	// Update state
	s.currentTrack = &track
	s.currentHandle = handle
	s.currentIndex = index
	s.manualStop = false
	s.hasPlayed = false

	s.logger.Debug("loadTrack succeeded", slog.Int64("handle", int64(s.currentHandle)))

	// Publish event
	s.bus.Publish(domain.NewTrackLoadedEvent(track, handle, duration, index))

	return nil
}

// Play starts or resumes playback of the current track.
func (s *PlaybackService) Play() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Debug("play called")

	if s.currentHandle == domain.InvalidTrackHandle {
		s.logger.Debug("play failed - invalid track handle")
		return domain.ErrInvalidTrackHandle
	}

	s.logger.Debug("play attempting", slog.Int64("handle", int64(s.currentHandle)))

	// Check the current status
	status, err := s.engine.Status(s.currentHandle)
	if err != nil {
		s.logger.Debug("play failed - status check error", slog.Any("error", err))
		return err
	}

	s.logger.Debug("current status", slog.Any("status", status))

	// Already playing
	if status == domain.StatusPlaying {
		s.logger.Debug("already playing, returning")
		return nil
	}

	// Start/resume playback
	s.manualStop = false
	s.hasPlayed = true
	if err := s.engine.Play(s.currentHandle); err != nil {
		s.logger.Debug("play failed - engine.Play error", slog.Any("error", err))
		return err
	}

	s.logger.Debug("play succeeded, publishing event")

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
	position, err := s.engine.Position(s.currentHandle)
	if err != nil {
		position = 0 // Default to 0 if position unavailable
	}

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
	s.hasPlayed = false

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
		duration, err := s.engine.Duration(s.currentHandle)
		if err != nil {
			duration = 0 // Default to 0 if duration unavailable
		}
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

		if duration, err := s.engine.Duration(s.currentHandle); err == nil {
			state.Duration = duration
		}
	} else {
		state.Status = domain.StatusStopped
	}

	return state
}

// Shutdown stops playback and cleans up resources.
func (s *PlaybackService) Shutdown() error {
	s.mu.Lock()

	// Stop update routine
	if s.updateRunning {
		close(s.stopUpdate)
		s.updateRunning = false
	}

	// Release lock before waiting for goroutine to exit (to avoid deadlock)
	s.mu.Unlock()

	// Wait for the update goroutine to finish
	s.updateWg.Wait()

	// Acquire lock again to stop the current track
	s.mu.Lock()
	defer s.mu.Unlock()

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
	s.updateWg.Add(1)
	s.mu.Unlock()

	go func() {
		defer s.updateWg.Done()
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

	// Nothing to update if no track loaded
	if s.currentHandle == domain.InvalidTrackHandle || s.currentTrack == nil {
		s.mu.RUnlock()
		return
	}

	// Get current status
	status, err := s.engine.Status(s.currentHandle)
	if err != nil {
		s.mu.RUnlock()
		return
	}

	// Get position and duration
	position, err := s.engine.Position(s.currentHandle)
	if err != nil {
		s.mu.RUnlock()
		return
	}

	duration, err := s.engine.Duration(s.currentHandle)
	if err != nil {
		s.mu.RUnlock()
		return
	}

	// Determine if track finished while holding read lock
	shouldFinish := status == domain.StatusStopped && !s.manualStop && s.hasPlayed
	track := s.currentTrack // Copy pointer for later use

	// Release read lock BEFORE any further processing
	s.mu.RUnlock()

	// Publish progress event (no lock needed - event bus is thread-safe)
	s.bus.Publish(domain.NewTrackProgressEvent(position, duration))

	// Handle track finished with NO lock held
	if shouldFinish && track != nil {
		s.mu.Lock()
		s.handleTrackFinishedWithLock() // Expects write lock, releases it before returning
		// Lock already released by handler - DO NOT UNLOCK HERE
	}
}

// handleTrackFinishedWithLock is called when a track finishes playing naturally.
// Expects write lock held on entry. ALWAYS releases lock before returning.
func (s *PlaybackService) handleTrackFinishedWithLock() {
	if s.currentTrack == nil {
		s.mu.Unlock() // Always unlock before early return
		return
	}

	track := *s.currentTrack
	shouldLoop := s.isLooping
	index := s.currentIndex

	// Reset state
	s.hasPlayed = false

	// Publish completed event
	s.bus.Publish(domain.NewTrackCompletedEvent(track))

	if shouldLoop {
		// Stop internal (expects write lock)
		if err := s.stopInternal(); err != nil {
			s.logger.Warn("failed to stop track in loop", slog.Any("error", err))
		}

		// Release lock before calling public methods
		s.mu.Unlock()

		// Call public methods (they acquire their own locks)
		if err := s.LoadTrack(track, index); err != nil {
			s.logger.Warn("failed to reload track in loop", slog.Any("error", err))
			return
		}
		if err := s.Play(); err != nil {
			s.logger.Warn("failed to play track in loop", slog.Any("error", err))
		}

		// Lock already released
	} else {
		// Release lock before publishing event
		s.mu.Unlock()

		// Publish auto-next event for playlist
		s.bus.Publish(domain.NewAutoNextEvent(track, index))
	}
	// Lock is ALWAYS released before this point
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
