package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApplication(t *testing.T) {
	config := DefaultConfig()
	config.UseMockAudio = true // Use mock for testing

	app, err := NewApplication(config)
	require.NoError(t, err)
	require.NotNil(t, app)

	// Verify all services were created
	playback, playlist, library, preference := app.GetServices()
	assert.NotNil(t, playback)
	assert.NotNil(t, playlist)
	assert.NotNil(t, library)
	assert.NotNil(t, preference)

	// Verify event bus was created
	assert.NotNil(t, app.GetEventBus())

	// Verify Fyne app was created
	assert.NotNil(t, app.GetFyneApp())

	// Cleanup
	err = app.Shutdown()
	assert.NoError(t, err)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "com.gotune.app", config.AppID)
	assert.Equal(t, "GoTune", config.AppName)
	assert.Equal(t, -1, config.AudioDevice)
	assert.Equal(t, 44100, config.SampleRate)
	assert.False(t, config.UseMockAudio)
}

func TestApplicationLifecycle(t *testing.T) {
	config := DefaultConfig()
	config.UseMockAudio = true

	// Create
	app, err := NewApplication(config)
	require.NoError(t, err)

	// Run would normally block, but we're not calling it in test

	// Shutdown
	err = app.Shutdown()
	assert.NoError(t, err)

	// Shutdown again should not panic
	err = app.Shutdown()
	assert.NoError(t, err)
}

func TestApplicationLoadSavedState(t *testing.T) {
	config := DefaultConfig()
	config.UseMockAudio = true

	app, err := NewApplication(config)
	require.NoError(t, err)
	defer app.Shutdown()

	// Get services
	playback, playlist, _, preference := app.GetServices()

	// Set some state
	playback.SetVolume(0.75)
	playback.SetLoop(true)
	preference.SetVolume(0.75)
	preference.SetLoopMode(true)

	// Add track to playlist
	// (would need actual tracks, but this shows the API)
	_ = playlist
	_ = preference

	// State is automatically saved on shutdown
	// and loaded on next startup via loadSavedState()
}

func TestApplicationWithServices(t *testing.T) {
	config := DefaultConfig()
	config.UseMockAudio = true

	app, err := NewApplication(config)
	require.NoError(t, err)
	defer app.Shutdown()

	playback, _, library, preference := app.GetServices()

	// Test that services work together
	volume := preference.GetVolume()
	assert.InDelta(t, 1.0, volume, 0.01) // Default volume from repository is 1.0

	// Set volume via service
	err = playback.SetVolume(0.6)
	assert.NoError(t, err)

	// Test library service
	formats := library.GetSupportedFormats()
	assert.Greater(t, len(formats), 20)
}
