package memory

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a test preferences repository
func newTestPreferencesRepository() *PreferencesRepository {
	app := test.NewApp()
	prefs := app.Preferences()

	return NewPreferencesRepository(prefs)
}

func TestPreferencesRepository_SaveAndLoadVolume(t *testing.T) {
	repo := newTestPreferencesRepository()

	// Save volume
	err := repo.SaveVolume(0.75)
	require.NoError(t, err)

	// Load volume
	volume, err := repo.LoadVolume()
	require.NoError(t, err)
	assert.Equal(t, 0.75, volume)
}

func TestPreferencesRepository_LoadVolume_Default(t *testing.T) {
	repo := newTestPreferencesRepository()

	// Load when nothing saved - should return default (1.0)
	volume, err := repo.LoadVolume()
	require.NoError(t, err)
	assert.Equal(t, 1.0, volume)
}

func TestPreferencesRepository_SaveVolume_BoundaryValues(t *testing.T) {
	repo := newTestPreferencesRepository()

	// Test minimum
	err := repo.SaveVolume(0.0)
	require.NoError(t, err)

	volume, err := repo.LoadVolume()
	require.NoError(t, err)
	assert.Equal(t, 0.0, volume)

	// Test maximum
	err = repo.SaveVolume(1.0)
	require.NoError(t, err)

	volume, err = repo.LoadVolume()
	require.NoError(t, err)
	assert.Equal(t, 1.0, volume)
}

func TestPreferencesRepository_SaveAndLoadLoopMode(t *testing.T) {
	repo := newTestPreferencesRepository()

	// Save loop enabled
	err := repo.SaveLoopMode(true)
	require.NoError(t, err)

	// Load loop
	loop, err := repo.LoadLoopMode()
	require.NoError(t, err)
	assert.True(t, loop)

	// Save loop disabled
	err = repo.SaveLoopMode(false)
	require.NoError(t, err)

	// Load loop
	loop, err = repo.LoadLoopMode()
	require.NoError(t, err)
	assert.False(t, loop)
}

func TestPreferencesRepository_LoadLoopMode_Default(t *testing.T) {
	repo := newTestPreferencesRepository()

	// Load when nothing saved - should return default (false)
	loop, err := repo.LoadLoopMode()
	require.NoError(t, err)
	assert.False(t, loop)
}

func TestPreferencesRepository_SaveAndLoadTheme(t *testing.T) {
	repo := newTestPreferencesRepository()

	// Save theme
	err := repo.SaveTheme("dark")
	require.NoError(t, err)

	// Load theme
	theme, err := repo.LoadTheme()
	require.NoError(t, err)
	assert.Equal(t, "dark", theme)

	// Change theme
	err = repo.SaveTheme("light")
	require.NoError(t, err)

	// Load theme
	theme, err = repo.LoadTheme()
	require.NoError(t, err)
	assert.Equal(t, "light", theme)
}

func TestPreferencesRepository_LoadTheme_Default(t *testing.T) {
	repo := newTestPreferencesRepository()

	// Load when nothing saved - should return default ("system")
	theme, err := repo.LoadTheme()
	require.NoError(t, err)
	assert.Equal(t, "system", theme)
}

func TestPreferencesRepository_SaveAndLoadScanPaths(t *testing.T) {
	repo := newTestPreferencesRepository()

	// Save paths
	paths := []string{
		"/home/user/Music",
		"/mnt/external/Music",
		"/data/audio",
	}
	err := repo.SaveScanPaths(paths)
	require.NoError(t, err)

	// Load paths
	loaded, err := repo.LoadScanPaths()
	require.NoError(t, err)
	require.Equal(t, 3, len(loaded))
	assert.Equal(t, "/home/user/Music", loaded[0])
	assert.Equal(t, "/mnt/external/Music", loaded[1])
	assert.Equal(t, "/data/audio", loaded[2])
}

func TestPreferencesRepository_LoadScanPaths_Empty(t *testing.T) {
	repo := newTestPreferencesRepository()

	// Load when nothing saved - should return empty slice
	paths, err := repo.LoadScanPaths()
	require.NoError(t, err)
	assert.Equal(t, 0, len(paths))
}

func TestPreferencesRepository_SaveScanPaths_EmptySlice(t *testing.T) {
	repo := newTestPreferencesRepository()

	// Save empty slice
	err := repo.SaveScanPaths([]string{})
	require.NoError(t, err)

	// Load should return empty slice
	paths, err := repo.LoadScanPaths()
	require.NoError(t, err)
	assert.Equal(t, 0, len(paths))
}

func TestPreferencesRepository_SaveScanPaths_OverwritesPrevious(t *testing.T) {
	repo := newTestPreferencesRepository()

	// Save first paths
	paths1 := []string{"/path1", "/path2"}
	err := repo.SaveScanPaths(paths1)
	require.NoError(t, err)

	// Save second paths (should overwrite)
	paths2 := []string{"/path3"}
	err = repo.SaveScanPaths(paths2)
	require.NoError(t, err)

	// Load should return second paths
	loaded, err := repo.LoadScanPaths()
	require.NoError(t, err)
	require.Equal(t, 1, len(loaded))
	assert.Equal(t, "/path3", loaded[0])
}

func TestPreferencesRepository_Clear(t *testing.T) {
	repo := newTestPreferencesRepository()

	// Save all preferences
	repo.SaveVolume(0.5)
	repo.SaveLoopMode(true)
	repo.SaveTheme("dark")
	repo.SaveScanPaths([]string{"/music"})

	// Clear
	err := repo.Clear()
	require.NoError(t, err)

	// Verify all cleared (should return defaults)
	volume, _ := repo.LoadVolume()
	assert.Equal(t, 1.0, volume) // Default

	loop, _ := repo.LoadLoopMode()
	assert.False(t, loop) // Default

	theme, _ := repo.LoadTheme()
	assert.Equal(t, "system", theme) // Default

	paths, _ := repo.LoadScanPaths()
	assert.Equal(t, 0, len(paths)) // Empty
}

func TestPreferencesRepository_MultiplePreferences(t *testing.T) {
	repo := newTestPreferencesRepository()

	// Set multiple preferences
	repo.SaveVolume(0.8)
	repo.SaveLoopMode(true)
	repo.SaveTheme("light")
	repo.SaveScanPaths([]string{"/music", "/audio"})

	// Verify all are saved correctly
	volume, _ := repo.LoadVolume()
	assert.Equal(t, 0.8, volume)

	loop, _ := repo.LoadLoopMode()
	assert.True(t, loop)

	theme, _ := repo.LoadTheme()
	assert.Equal(t, "light", theme)

	paths, _ := repo.LoadScanPaths()
	assert.Equal(t, 2, len(paths))
}

func TestPreferencesRepository_SaveLoadCycle(t *testing.T) {
	repo := newTestPreferencesRepository()

	// Multiple save/load cycles
	for i := 0; i < 10; i++ {
		volume := float64(i) / 10.0

		err := repo.SaveVolume(volume)
		require.NoError(t, err)

		loaded, err := repo.LoadVolume()
		require.NoError(t, err)
		assert.Equal(t, volume, loaded)
	}
}
