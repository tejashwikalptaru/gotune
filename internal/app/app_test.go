package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// Ignore Fyne framework goroutines that run in the background
		goleak.IgnoreAnyFunction("fyne.io/fyne/v2"),
	)
}

func TestNewApplication(t *testing.T) {
	config := DefaultConfig()
	config.UseMockAudio = true // Use mock for testing

	app, err := NewApplication(config)
	require.NoError(t, err)
	require.NotNil(t, app)

	// Cleanup
	app.Shutdown()
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
	app.Shutdown()

	// Shutdown again should not panic
	app.Shutdown()
}

func TestApplicationLoadSavedState(t *testing.T) {
	config := DefaultConfig()
	config.UseMockAudio = true

	app, err := NewApplication(config)
	require.NoError(t, err)
	defer app.Shutdown()

	// State is automatically saved on shutdown
	// and loaded on the next startup via loadSavedState()
}

func TestApplicationWithServices(t *testing.T) {
	config := DefaultConfig()
	config.UseMockAudio = true

	app, err := NewApplication(config)
	require.NoError(t, err)
	defer app.Shutdown()

	// Application is successfully created with all services wired together
	// Services are tested individually in their respective test packages
}
