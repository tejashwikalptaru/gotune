package mock

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/tejashwikalptaru/gotune/internal/domain"
)

// TestNewMockEngine tests creating a new mock engine.
func TestNewMockEngine(t *testing.T) {
	engine := NewEngine()

	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}

	if engine.IsInitialized() {
		t.Error("New engine should not be initialized")
	}

	if engine.GetLoadedTracks() != 0 {
		t.Errorf("Expected 0 loaded tracks, got %d", engine.GetLoadedTracks())
	}
}

// TestInitialize tests engine initialization.
func TestInitialize(t *testing.T) {
	engine := NewEngine()

	err := engine.Initialize(-1, 44100, 0)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if !engine.IsInitialized() {
		t.Error("Engine should be initialized")
	}
}

// TestInitializeAlreadyInitialized tests initializing an already initialized engine.
func TestInitializeAlreadyInitialized(t *testing.T) {
	engine := NewEngine()

	err := engine.Initialize(-1, 44100, 0)
	if err != nil {
		t.Fatalf("First Initialize failed: %v", err)
	}

	// Try to initialize again
	err = engine.Initialize(-1, 44100, 0)
	if !errors.Is(err, domain.ErrAlreadyInitialized) {
		t.Errorf("Expected ErrAlreadyInitialized, got %v", err)
	}
}

// TestShutdown tests shutting down the engine.
func TestShutdown(t *testing.T) {
	engine := NewEngine()
	err := engine.Initialize(-1, 44100, 0)
	if err != nil {
		t.Errorf("Initialization failed: %v", err)
	}

	// Load a track
	handle, err := engine.Load("/path/to/test.mp3")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if engine.GetLoadedTracks() != 1 {
		t.Error("Expected 1 loaded track before shutdown")
	}

	// Shutdown
	err = engine.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	if engine.IsInitialized() {
		t.Error("Engine should not be initialized after shutdown")
	}

	if engine.GetLoadedTracks() != 0 {
		t.Errorf("Expected 0 loaded tracks after shutdown, got %d", engine.GetLoadedTracks())
	}

	// Trying to use the handle after shutdown should fail
	_, err = engine.Status(handle)
	if !errors.Is(err, domain.ErrNotInitialized) {
		t.Errorf("Expected ErrNotInitialized, got %v", err)
	}
}

// TestLoad tests loading tracks.
func TestLoad(t *testing.T) {
	engine := NewEngine()
	_ = engine.Initialize(-1, 44100, 0)
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	handle, err := engine.Load("/path/to/test.mp3")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if handle == domain.InvalidTrackHandle {
		t.Error("Load returned invalid handle")
	}

	if engine.GetLoadedTracks() != 1 {
		t.Errorf("Expected 1 loaded track, got %d", engine.GetLoadedTracks())
	}
}

// TestLoadMultipleTracks tests loading multiple tracks.
func TestLoadMultipleTracks(t *testing.T) {
	engine := NewEngine()
	_ = engine.Initialize(-1, 44100, 0)
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	handle1, err := engine.Load("/path/to/test1.mp3")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	handle2, err := engine.Load("/path/to/test2.mp3")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	handle3, err := engine.Load("/path/to/test3.mp3")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if handle1 == handle2 || handle1 == handle3 || handle2 == handle3 {
		t.Error("Handles should be unique")
	}

	if engine.GetLoadedTracks() != 3 {
		t.Errorf("Expected 3 loaded tracks, got %d", engine.GetLoadedTracks())
	}
}

// TestLoadWithoutInitialize tests loading without initialization.
func TestLoadWithoutInitialize(t *testing.T) {
	engine := NewEngine()

	_, err := engine.Load("/path/to/test.mp3")
	if !errors.Is(err, domain.ErrNotInitialized) {
		t.Errorf("Expected ErrNotInitialized, got %v", err)
	}
}

// TestUnload tests unloading tracks.
func TestUnload(t *testing.T) {
	engine := NewEngine()
	_ = engine.Initialize(-1, 44100, 0)
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	handle, err := engine.Load("/path/to/test.mp3")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	err = engine.Unload(handle)
	if err != nil {
		t.Errorf("Unload failed: %v", err)
	}

	if engine.GetLoadedTracks() != 0 {
		t.Errorf("Expected 0 loaded tracks after unload, got %d", engine.GetLoadedTracks())
	}

	// Using the handle after unloading should fail
	_, err = engine.Status(handle)
	if !errors.Is(err, domain.ErrInvalidTrackHandle) {
		t.Errorf("Expected ErrInvalidTrackHandle, got %v", err)
	}
}

// TestPlay tests starting playback.
func TestPlay(t *testing.T) {
	engine := NewEngine()
	_ = engine.Initialize(-1, 44100, 0)
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	handle, err := engine.Load("/path/to/test.mp3")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	err = engine.Play(handle)
	if err != nil {
		t.Errorf("Play failed: %v", err)
	}

	status, err := engine.Status(handle)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if status != domain.StatusPlaying {
		t.Errorf("Expected StatusPlaying, got %v", status)
	}
}

// TestPause tests pausing playback.
func TestPause(t *testing.T) {
	engine := NewEngine()
	_ = engine.Initialize(-1, 44100, 0)
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	handle, err := engine.Load("/path/to/test.mp3")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	_ = engine.Play(handle)

	err = engine.Pause(handle)
	if err != nil {
		t.Errorf("Pause failed: %v", err)
	}

	status, err := engine.Status(handle)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if status != domain.StatusPaused {
		t.Errorf("Expected StatusPaused, got %v", status)
	}
}

// TestStop tests stopping playback.
func TestStop(t *testing.T) {
	engine := NewEngine()
	_ = engine.Initialize(-1, 44100, 0)
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	handle, err := engine.Load("/path/to/test.mp3")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	_ = engine.Play(handle)

	err = engine.Stop(handle)
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}

	// Stop should unload the track
	if engine.GetLoadedTracks() != 0 {
		t.Errorf("Expected 0 loaded tracks after stop, got %d", engine.GetLoadedTracks())
	}

	// Using the handle after stop should fail
	_, err = engine.Status(handle)
	if !errors.Is(err, domain.ErrInvalidTrackHandle) {
		t.Errorf("Expected ErrInvalidTrackHandle after stop, got %v", err)
	}
}

// TestDuration tests getting track duration.
func TestDuration(t *testing.T) {
	engine := NewEngine()
	_ = engine.Initialize(-1, 44100, 0)
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	handle, err := engine.Load("/path/to/test.mp3")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	duration, err := engine.Duration(handle)
	if err != nil {
		t.Errorf("Duration failed: %v", err)
	}

	if duration != 3*time.Minute {
		t.Errorf("Expected 3 minute duration, got %v", duration)
	}
}

// TestPosition tests getting and setting position.
func TestPosition(t *testing.T) {
	engine := NewEngine()
	_ = engine.Initialize(-1, 44100, 0)
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	handle, err := engine.Load("/path/to/test.mp3")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Initial position should be 0
	pos, err := engine.Position(handle)
	if err != nil {
		t.Errorf("Position failed: %v", err)
	}
	if pos != 0 {
		t.Errorf("Expected initial position 0, got %v", pos)
	}

	// Seek to 1 minute
	err = engine.Seek(handle, time.Minute)
	if err != nil {
		t.Errorf("Seek failed: %v", err)
	}

	// Check a new position
	pos, err = engine.Position(handle)
	if err != nil {
		t.Fatalf("Position failed: %v", err)
	}
	if pos != time.Minute {
		t.Errorf("Expected position 1m, got %v", pos)
	}
}

// TestSeekInvalidPosition tests seeking invalid positions.
func TestSeekInvalidPosition(t *testing.T) {
	engine := NewEngine()
	_ = engine.Initialize(-1, 44100, 0)
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	handle, err := engine.Load("/path/to/test.mp3")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Seek to negative position
	err = engine.Seek(handle, -1*time.Second)
	if !errors.Is(err, domain.ErrInvalidPosition) {
		t.Errorf("Expected ErrInvalidPosition for negative position, got %v", err)
	}

	// Seek beyond duration
	err = engine.Seek(handle, 10*time.Minute)
	if !errors.Is(err, domain.ErrInvalidPosition) {
		t.Errorf("Expected ErrInvalidPosition for position beyond duration, got %v", err)
	}
}

// TestVolume tests volume control.
func TestVolume(t *testing.T) {
	engine := NewEngine()
	_ = engine.Initialize(-1, 44100, 0)
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	handle, err := engine.Load("/path/to/test.mp3")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// The default volume should be 1.0
	vol, err := engine.GetVolume(handle)
	if err != nil {
		t.Errorf("GetVolume failed: %v", err)
	}
	if vol != 1.0 {
		t.Errorf("Expected default volume 1.0, got %v", vol)
	}

	// Set volume to 0.5
	err = engine.SetVolume(handle, 0.5)
	if err != nil {
		t.Errorf("SetVolume failed: %v", err)
	}

	// Check new volume
	vol, err = engine.GetVolume(handle)
	if err != nil {
		t.Fatalf("GetVolume failed: %v", err)
	}
	if vol != 0.5 {
		t.Errorf("Expected volume 0.5, got %v", vol)
	}
}

// TestVolumeInvalidRange tests setting volume out of range.
func TestVolumeInvalidRange(t *testing.T) {
	engine := NewEngine()
	_ = engine.Initialize(-1, 44100, 0)
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	handle, err := engine.Load("/path/to/test.mp3")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Volume too low
	err = engine.SetVolume(handle, -0.1)
	if !errors.Is(err, domain.ErrInvalidVolume) {
		t.Errorf("Expected ErrInvalidVolume for negative volume, got %v", err)
	}

	// Volume too high
	err = engine.SetVolume(handle, 1.1)
	if !errors.Is(err, domain.ErrInvalidVolume) {
		t.Errorf("Expected ErrInvalidVolume for volume > 1.0, got %v", err)
	}
}

// TestGetMetadata tests extracting metadata.
func TestGetMetadata(t *testing.T) {
	engine := NewEngine()

	track, err := engine.GetMetadata("/path/to/test.mp3")
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}

	if track == nil {
		t.Fatal("GetMetadata returned nil track")
	}

	if track.FilePath != "/path/to/test.mp3" {
		t.Errorf("Expected FilePath /path/to/test.mp3, got %s", track.FilePath)
	}

	if track.Title != "test" {
		t.Errorf("Expected Title 'test', got %s", track.Title)
	}

	if track.FileFormat != ".mp3" {
		t.Errorf("Expected FileFormat .mp3, got %s", track.FileFormat)
	}

	if track.IsMOD {
		t.Error("MP3 file should not be marked as MOD")
	}
}

// TestGetMetadataMODFormat tests MOD format detection.
func TestGetMetadataMODFormat(t *testing.T) {
	engine := NewEngine()

	track, err := engine.GetMetadata("/path/to/test.mod")
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}

	if !track.IsMOD {
		t.Error("MOD file should be marked as MOD")
	}

	if track.FileFormat != ".mod" {
		t.Errorf("Expected FileFormat .mod, got %s", track.FileFormat)
	}
}

// TestSimulateProgress tests simulating playback progress.
func TestSimulateProgress(t *testing.T) {
	engine := NewEngine()
	_ = engine.Initialize(-1, 44100, 0)
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	handle, err := engine.Load("/path/to/test.mp3")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	_ = engine.Play(handle)

	// Simulate 30 seconds of progress
	err = engine.SimulateProgress(handle, 30*time.Second)
	if err != nil {
		t.Errorf("SimulateProgress failed: %v", err)
	}

	pos, err := engine.Position(handle)
	if err != nil {
		t.Fatalf("Position failed: %v", err)
	}
	if pos != 30*time.Second {
		t.Errorf("Expected position 30s, got %v", pos)
	}

	// Simulate progress beyond duration
	err = engine.SimulateProgress(handle, 5*time.Minute)
	if err != nil {
		t.Errorf("SimulateProgress failed: %v", err)
	}

	// Position should be capped at duration
	pos, err = engine.Position(handle)
	if err != nil {
		t.Fatalf("Position failed: %v", err)
	}
	duration, err := engine.Duration(handle)
	if err != nil {
		t.Fatalf("Duration failed: %v", err)
	}
	if pos != duration {
		t.Errorf("Expected position at duration %v, got %v", duration, pos)
	}

	// Track should be stopped after reaching the end
	status, err := engine.Status(handle)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if status != domain.StatusStopped {
		t.Errorf("Expected StatusStopped after reaching end, got %v", status)
	}
}

// TestFailInitialize tests configured initialization failure.
func TestFailInitialize(t *testing.T) {
	engine := NewEngine()
	engine.SetFailInitialize(true)

	err := engine.Initialize(-1, 44100, 0)
	if err == nil {
		t.Error("Expected initialization to fail")
	}

	if engine.IsInitialized() {
		t.Error("Engine should not be initialized after failed init")
	}
}

// TestFailLoad tests configured load failure.
func TestFailLoad(t *testing.T) {
	engine := NewEngine()
	_ = engine.Initialize(-1, 44100, 0)
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	engine.SetFailLoad(true)

	_, err := engine.Load("/path/to/test.mp3")
	if err == nil {
		t.Error("Expected load to fail")
	}
}

// TestFailPlay tests configured playback failure.
func TestFailPlay(t *testing.T) {
	engine := NewEngine()
	_ = engine.Initialize(-1, 44100, 0)
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	handle, err := engine.Load("/path/to/test.mp3")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	engine.SetFailPlay(true)

	err = engine.Play(handle)
	if err == nil {
		t.Error("Expected play to fail")
	}
}

// TestConcurrentLoad tests concurrent track loading.
func TestConcurrentLoad(t *testing.T) {
	engine := NewEngine()
	_ = engine.Initialize(-1, 44100, 0)
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	const numGoroutines = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	handles := make([]domain.TrackHandle, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			handle, err := engine.Load("/path/to/test.mp3")
			if err != nil {
				t.Errorf("Load failed: %v", err)
			}
			handles[index] = handle
		}(i)
	}

	wg.Wait()

	if engine.GetLoadedTracks() != numGoroutines {
		t.Errorf("Expected %d loaded tracks, got %d", numGoroutines, engine.GetLoadedTracks())
	}

	// All handles should be unique
	seen := make(map[domain.TrackHandle]bool)
	for _, handle := range handles {
		if seen[handle] {
			t.Error("Duplicate handle detected in concurrent loading")
		}
		seen[handle] = true
	}
}

// TestConcurrentPlayback tests concurrent playback operations.
func TestConcurrentPlayback(t *testing.T) {
	engine := NewEngine()
	_ = engine.Initialize(-1, 44100, 0)
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	// Load multiple tracks
	const numTracks = 5
	handles := make([]domain.TrackHandle, numTracks)
	for i := 0; i < numTracks; i++ {
		handle, err := engine.Load("/path/to/test.mp3")
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		handles[i] = handle
	}

	// Concurrently play, pause, and stop tracks
	var wg sync.WaitGroup
	wg.Add(numTracks * 3)

	for _, handle := range handles {
		h := handle
		go func() {
			defer wg.Done()
			_ = engine.Play(h)
		}()
		go func() {
			defer wg.Done()
			time.Sleep(time.Millisecond)
			_ = engine.Pause(h)
		}()
		go func() {
			defer wg.Done()
			time.Sleep(2 * time.Millisecond)
			_ = engine.Stop(h)
		}()
	}

	wg.Wait()

	// All tracks should be stopped (and unloaded)
	if engine.GetLoadedTracks() != 0 {
		t.Errorf("Expected 0 tracks after stopping all, got %d", engine.GetLoadedTracks())
	}
}
