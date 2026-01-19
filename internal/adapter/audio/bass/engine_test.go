// Package bass provides tests for the BASS audio engine adapter.
//
// NOTE: These tests require the BASS library to be available for the current platform
// and architecture. On Apple Silicon (arm64), you may need the arm64 version of the
// BASS library or run tests under Rosetta 2 with GOARCH=amd64.
//
// To run tests on Apple Silicon:
//
//	GOARCH=amd64 go test -v ./internal/adapter/audio/bass/
//
// Or download arm64 BASS library from https://www.un4seen.com/
package bass

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tejashwikalptaru/gotune/internal/domain"
)

// Test audio file paths (relative to project root)
const (
	testDataDir = "../../../../test/testdata/audio"
)

// getTestAudioFile returns the path to a test audio file.
// Creates a minimal WAV file if it doesn't exist.
func getTestAudioFile(t *testing.T) string {
	// Try to find an existing audio file in the project
	testFile := filepath.Join(testDataDir, "test.wav")

	// If the test file doesn't exist, we'll create a minimal valid WAV file
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		// Create a test data directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(testFile), 0755); err != nil {
			t.Skipf("Cannot create test data directory: %v", err)
			return ""
		}

		// Create a minimal valid WAV file (1 second of silence, 44100 Hz, mono, 16-bit)
		wavData := createMinimalWAV(44100, 1) // 1 second
		if err := os.WriteFile(testFile, wavData, 0600); err != nil {
			t.Skipf("Cannot create test WAV file: %v", err)
			return ""
		}
	}

	return testFile
}

// createMinimalWAV creates a minimal valid WAV file with silence.
func createMinimalWAV(sampleRate, durationSeconds int) []byte {
	numSamples := sampleRate * durationSeconds
	dataSize := numSamples * 2 // 16-bit samples = 2 bytes per sample
	fileSize := 36 + dataSize

	wav := make([]byte, 44+dataSize)

	// RIFF header
	copy(wav[0:4], "RIFF")
	writeUint32LE(wav[4:8], uint32(fileSize))
	copy(wav[8:12], "WAVE")

	// fmt chunk
	copy(wav[12:16], "fmt ")
	writeUint32LE(wav[16:20], 16)                   // fmt chunk size
	writeUint16LE(wav[20:22], 1)                    // audio format (PCM)
	writeUint16LE(wav[22:24], 1)                    // num channels (mono)
	writeUint32LE(wav[24:28], uint32(sampleRate))   // sample rate
	writeUint32LE(wav[28:32], uint32(sampleRate*2)) // byte rate
	writeUint16LE(wav[32:34], 2)                    // block align
	writeUint16LE(wav[34:36], 16)                   // bits per sample

	// data chunk
	copy(wav[36:40], "data")
	writeUint32LE(wav[40:44], uint32(dataSize))
	// Data is already zeros (silence)

	return wav
}

func writeUint32LE(buf []byte, val uint32) {
	buf[0] = byte(val)
	buf[1] = byte(val >> 8)
	buf[2] = byte(val >> 16)
	buf[3] = byte(val >> 24)
}

func writeUint16LE(buf []byte, val uint16) {
	buf[0] = byte(val)
	buf[1] = byte(val >> 8)
}

func TestBassEngine_Initialize(t *testing.T) {
	engine := NewEngine()
	require.NotNil(t, engine)

	// Should not be initialized initially
	assert.False(t, engine.IsInitialized())

	// Initialize
	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	// Should be initialized now
	assert.True(t, engine.IsInitialized())

	// Cleanup
	err = engine.Shutdown()
	require.NoError(t, err)

	// Should not be initialized after shutdown
	assert.False(t, engine.IsInitialized())
}

func TestBassEngine_InitializeTwice(t *testing.T) {
	engine := NewEngine()
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	// Second initialization should fail
	err = engine.Initialize(-1, 44100, 0)
	assert.Equal(t, domain.ErrAlreadyInitialized, err)
}

func TestBassEngine_ShutdownNotInitialized(t *testing.T) {
	engine := NewEngine()

	// Shutdown without initialization should fail
	err := engine.Shutdown()
	assert.Equal(t, domain.ErrNotInitialized, err)
}

func TestBassEngine_LoadTrack(t *testing.T) {
	testFile := getTestAudioFile(t)
	if testFile == "" {
		t.Skip("No test audio file available")
	}

	engine := NewEngine()
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	// Load track
	handle, err := engine.Load(testFile)
	require.NoError(t, err)
	assert.NotEqual(t, domain.InvalidTrackHandle, handle)

	// Cleanup
	err = engine.Unload(handle)
	assert.NoError(t, err)
}

func TestBassEngine_LoadNotInitialized(t *testing.T) {
	testFile := getTestAudioFile(t)
	if testFile == "" {
		t.Skip("No test audio file available")
	}

	engine := NewEngine()

	// Load without initialization should fail
	handle, err := engine.Load(testFile)
	assert.Equal(t, domain.ErrNotInitialized, err)
	assert.Equal(t, domain.InvalidTrackHandle, handle)
}

func TestBassEngine_LoadInvalidPath(t *testing.T) {
	engine := NewEngine()
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	// Load with an empty path
	handle, err := engine.Load("")
	assert.Equal(t, domain.ErrInvalidFilePath, err)
	assert.Equal(t, domain.InvalidTrackHandle, handle)

	// Load non-existent file
	handle, err = engine.Load("/nonexistent/file.mp3")
	assert.Error(t, err)
	assert.Equal(t, domain.InvalidTrackHandle, handle)
}

func TestBassEngine_UnloadInvalidHandle(t *testing.T) {
	engine := NewEngine()
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	// Unload invalid handle
	err = engine.Unload(domain.InvalidTrackHandle)
	assert.Equal(t, domain.ErrInvalidTrackHandle, err)
}

func TestBassEngine_PlayPauseStop(t *testing.T) {
	testFile := getTestAudioFile(t)
	if testFile == "" {
		t.Skip("No test audio file available")
	}

	engine := NewEngine()
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	handle, err := engine.Load(testFile)
	require.NoError(t, err)
	defer func() {
		if err := engine.Unload(handle); err != nil {
			t.Errorf("Error during engine unload: %v", err)
		}
	}()

	// Initial status should be stopped
	status, err := engine.Status(handle)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusStopped, status)

	// Play
	err = engine.Play(handle)
	require.NoError(t, err)

	// Give it a moment to start playing
	time.Sleep(10 * time.Millisecond)

	status, err = engine.Status(handle)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusPlaying, status)

	// Pause
	err = engine.Pause(handle)
	require.NoError(t, err)

	status, err = engine.Status(handle)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusPaused, status)

	// Resume playing
	err = engine.Play(handle)
	require.NoError(t, err)

	status, err = engine.Status(handle)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusPlaying, status)

	// Stop (this also unloads in our implementation)
	err = engine.Stop(handle)
	require.NoError(t, err)

	// Status check on stopped/unloaded track should fail
	_, err = engine.Status(handle)
	assert.Equal(t, domain.ErrInvalidTrackHandle, err)
}

func TestBassEngine_Volume(t *testing.T) {
	testFile := getTestAudioFile(t)
	if testFile == "" {
		t.Skip("No test audio file available")
	}

	engine := NewEngine()
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	handle, err := engine.Load(testFile)
	require.NoError(t, err)
	defer engine.Unload(handle)

	// Set volume to 0.5
	err = engine.SetVolume(handle, 0.5)
	require.NoError(t, err)

	// Get volume
	volume, err := engine.GetVolume(handle)
	require.NoError(t, err)
	assert.InDelta(t, 0.5, volume, 0.01) // Allow small floating point error

	// Set volume to 1.0
	err = engine.SetVolume(handle, 1.0)
	require.NoError(t, err)

	volume, err = engine.GetVolume(handle)
	require.NoError(t, err)
	assert.InDelta(t, 1.0, volume, 0.01)

	// Set volume to 0.0
	err = engine.SetVolume(handle, 0.0)
	require.NoError(t, err)

	volume, err = engine.GetVolume(handle)
	require.NoError(t, err)
	assert.InDelta(t, 0.0, volume, 0.01)
}

func TestBassEngine_VolumeInvalidRange(t *testing.T) {
	testFile := getTestAudioFile(t)
	if testFile == "" {
		t.Skip("No test audio file available")
	}

	engine := NewEngine()
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	handle, err := engine.Load(testFile)
	require.NoError(t, err)
	defer func() {
		if err := engine.Unload(handle); err != nil {
			t.Errorf("Error during engine unload: %v", err)
		}
	}()

	// Volume below 0
	err = engine.SetVolume(handle, -0.1)
	assert.Equal(t, domain.ErrInvalidVolume, err)

	// Volume above 1
	err = engine.SetVolume(handle, 1.1)
	assert.Equal(t, domain.ErrInvalidVolume, err)
}

func TestBassEngine_PositionAndDuration(t *testing.T) {
	testFile := getTestAudioFile(t)
	if testFile == "" {
		t.Skip("No test audio file available")
	}

	engine := NewEngine()
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	handle, err := engine.Load(testFile)
	require.NoError(t, err)
	defer func() {
		if err := engine.Unload(handle); err != nil {
			t.Errorf("Error during engine unload: %v", err)
		}
	}()

	// Get duration
	duration, err := engine.Duration(handle)
	require.NoError(t, err)
	assert.Greater(t, duration, time.Duration(0))

	// Initial position should be 0
	position, err := engine.Position(handle)
	require.NoError(t, err)
	assert.Equal(t, time.Duration(0), position)

	// Play for a bit
	err = engine.Play(handle)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Position should have advanced
	position, err = engine.Position(handle)
	require.NoError(t, err)
	assert.Greater(t, position, time.Duration(0))
	assert.LessOrEqual(t, position, duration)
}

func TestBassEngine_Seek(t *testing.T) {
	testFile := getTestAudioFile(t)
	if testFile == "" {
		t.Skip("No test audio file available")
	}

	engine := NewEngine()
	defer func() {
		if err := engine.Shutdown(); err != nil {
			t.Errorf("Error during engine shutdown: %v", err)
		}
	}()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	handle, err := engine.Load(testFile)
	require.NoError(t, err)
	defer engine.Unload(handle)

	duration, err := engine.Duration(handle)
	require.NoError(t, err)

	// Seek to middle
	seekPos := duration / 2
	err = engine.Seek(handle, seekPos)
	require.NoError(t, err)

	position, err := engine.Position(handle)
	require.NoError(t, err)
	assert.InDelta(t, seekPos.Seconds(), position.Seconds(), 0.1) // Allow 100ms tolerance

	// Seek to beginning
	err = engine.Seek(handle, 0)
	require.NoError(t, err)

	position, err = engine.Position(handle)
	require.NoError(t, err)
	assert.Equal(t, time.Duration(0), position)
}

func TestBassEngine_SeekInvalidPosition(t *testing.T) {
	testFile := getTestAudioFile(t)
	if testFile == "" {
		t.Skip("No test audio file available")
	}

	engine := NewEngine()
	defer engine.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	handle, err := engine.Load(testFile)
	require.NoError(t, err)
	defer engine.Unload(handle)

	duration, err := engine.Duration(handle)
	require.NoError(t, err)

	// Seek beyond duration
	err = engine.Seek(handle, duration+time.Second)
	assert.Equal(t, domain.ErrInvalidPosition, err)

	// Seek to negative position
	err = engine.Seek(handle, -time.Second)
	assert.Equal(t, domain.ErrInvalidPosition, err)
}

func TestBassEngine_GetMetadata(t *testing.T) {
	testFile := getTestAudioFile(t)
	if testFile == "" {
		t.Skip("No test audio file available")
	}

	engine := NewEngine()
	// Metadata extraction doesn't require initialization

	metadata, err := engine.GetMetadata(testFile)
	require.NoError(t, err)
	require.NotNil(t, metadata)

	// Should have basic info
	assert.NotEmpty(t, metadata.FilePath)
	assert.NotEmpty(t, metadata.ID)
	assert.NotEmpty(t, metadata.FileFormat)

	// Duration might be available depending on the file
	// (our generated WAV should have duration)
	if metadata.Duration > 0 {
		assert.Greater(t, metadata.Duration, time.Duration(0))
	}
}

func TestBassEngine_GetMetadataInvalidPath(t *testing.T) {
	engine := NewEngine()

	// Empty path
	metadata, err := engine.GetMetadata("")
	assert.Equal(t, domain.ErrInvalidFilePath, err)
	assert.Nil(t, metadata)

	// Non-existent file
	metadata, err = engine.GetMetadata("/nonexistent/file.mp3")
	assert.Equal(t, domain.ErrFileNotFound, err)
	assert.Nil(t, metadata)
}

func TestBassEngine_MultipleTracksLoaded(t *testing.T) {
	testFile := getTestAudioFile(t)
	if testFile == "" {
		t.Skip("No test audio file available")
	}

	engine := NewEngine()
	defer engine.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	// Load multiple instances of the same track
	handle1, err := engine.Load(testFile)
	require.NoError(t, err)

	handle2, err := engine.Load(testFile)
	require.NoError(t, err)

	// Handles should be different
	assert.NotEqual(t, handle1, handle2)

	// Should have 2 loaded tracks
	assert.Equal(t, 2, engine.GetLoadedTracksCount())

	// Both should be playable
	err = engine.Play(handle1)
	require.NoError(t, err)

	err = engine.Play(handle2)
	require.NoError(t, err)

	// Cleanup
	err = engine.Unload(handle1)
	require.NoError(t, err)

	err = engine.Unload(handle2)
	require.NoError(t, err)

	assert.Equal(t, 0, engine.GetLoadedTracksCount())
}

func TestBassEngine_ShutdownWithLoadedTracks(t *testing.T) {
	testFile := getTestAudioFile(t)
	if testFile == "" {
		t.Skip("No test audio file available")
	}

	engine := NewEngine()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	// Load multiple tracks
	handle1, err := engine.Load(testFile)
	require.NoError(t, err)

	handle2, err := engine.Load(testFile)
	require.NoError(t, err)

	// Start playing one
	err = engine.Play(handle1)
	require.NoError(t, err)

	// Shutdown should clean up all tracks
	err = engine.Shutdown()
	require.NoError(t, err)

	// Should have 0 tracks after shutdown
	assert.Equal(t, 0, engine.GetLoadedTracksCount())

	// Operations on handles should fail after shutdown
	_, err = engine.Status(handle1)
	assert.Equal(t, domain.ErrNotInitialized, err)

	_, err = engine.Status(handle2)
	assert.Equal(t, domain.ErrNotInitialized, err)
}

// Thread safety tests

func TestBassEngine_ConcurrentLoad(t *testing.T) {
	testFile := getTestAudioFile(t)
	if testFile == "" {
		t.Skip("No test audio file available")
	}

	engine := NewEngine()
	defer engine.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	// Load tracks concurrently
	const numGoroutines = 10
	handles := make([]domain.TrackHandle, numGoroutines)
	errors := make([]error, numGoroutines)

	done := make(chan struct{})
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			handles[index], errors[index] = engine.Load(testFile)
			done <- struct{}{}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// All should succeed
	for i := 0; i < numGoroutines; i++ {
		assert.NoError(t, errors[i])
		assert.NotEqual(t, domain.InvalidTrackHandle, handles[i])
	}

	// Cleanup
	for i := 0; i < numGoroutines; i++ {
		if handles[i] != domain.InvalidTrackHandle {
			engine.Unload(handles[i])
		}
	}
}

func TestBassEngine_ConcurrentPlayback(t *testing.T) {
	testFile := getTestAudioFile(t)
	if testFile == "" {
		t.Skip("No test audio file available")
	}

	engine := NewEngine()
	defer engine.Shutdown()

	err := engine.Initialize(-1, 44100, 0)
	require.NoError(t, err)

	// Load multiple tracks
	const numTracks = 5
	handles := make([]domain.TrackHandle, numTracks)
	for i := 0; i < numTracks; i++ {
		handles[i], err = engine.Load(testFile)
		require.NoError(t, err)
	}

	// Play all tracks concurrently
	done := make(chan struct{})
	for i := 0; i < numTracks; i++ {
		go func(handle domain.TrackHandle) {
			engine.Play(handle)
			time.Sleep(50 * time.Millisecond)
			engine.Pause(handle)
			done <- struct{}{}
		}(handles[i])
	}

	// Wait for all
	for i := 0; i < numTracks; i++ {
		<-done
	}

	// All should be paused
	for i := 0; i < numTracks; i++ {
		status, err := engine.Status(handles[i])
		assert.NoError(t, err)
		assert.Equal(t, domain.StatusPaused, status)
	}

	// Cleanup
	for i := 0; i < numTracks; i++ {
		engine.Unload(handles[i])
	}
}
