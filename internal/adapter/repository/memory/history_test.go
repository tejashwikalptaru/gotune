package memory

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tejashwikalptaru/gotune/internal/domain"
)

// Helper to create a test history repository
func newTestHistoryRepository() *HistoryRepository {
	// Use Fyne's test app which provides an in-memory preferences backend
	app := test.NewApp()
	prefs := app.Preferences()

	return NewHistoryRepository(prefs)
}

func TestHistoryRepository_SaveAndLoadQueue(t *testing.T) {
	repo := newTestHistoryRepository()

	// Create test tracks
	tracks := []domain.MusicTrack{
		{
			ID:       "track1",
			FilePath: "/music/song1.mp3",
			Title:    "Song 1",
			Artist:   "Artist 1",
		},
		{
			ID:       "track2",
			FilePath: "/music/song2.mp3",
			Title:    "Song 2",
			Artist:   "Artist 2",
		},
	}

	// Save queue
	err := repo.SaveQueue(tracks)
	require.NoError(t, err)

	// Load queue
	loaded, err := repo.LoadQueue()
	require.NoError(t, err)
	require.Equal(t, 2, len(loaded))

	// Verify tracks
	assert.Equal(t, "track1", loaded[0].ID)
	assert.Equal(t, "/music/song1.mp3", loaded[0].FilePath)
	assert.Equal(t, "Song 1", loaded[0].Title)

	assert.Equal(t, "track2", loaded[1].ID)
	assert.Equal(t, "/music/song2.mp3", loaded[1].FilePath)
	assert.Equal(t, "Song 2", loaded[1].Title)
}

func TestHistoryRepository_LoadQueue_Empty(t *testing.T) {
	repo := newTestHistoryRepository()

	// Load when nothing saved
	loaded, err := repo.LoadQueue()
	require.NoError(t, err)
	assert.Equal(t, 0, len(loaded))
}

func TestHistoryRepository_SaveQueue_EmptySlice(t *testing.T) {
	repo := newTestHistoryRepository()

	// Save empty queue
	err := repo.SaveQueue([]domain.MusicTrack{})
	require.NoError(t, err)

	// Load should return empty slice
	loaded, err := repo.LoadQueue()
	require.NoError(t, err)
	assert.Equal(t, 0, len(loaded))
}

func TestHistoryRepository_SaveQueue_OverwritesPrevious(t *testing.T) {
	repo := newTestHistoryRepository()

	// Save first queue
	tracks1 := []domain.MusicTrack{
		{ID: "track1", FilePath: "/music/song1.mp3"},
	}
	err := repo.SaveQueue(tracks1)
	require.NoError(t, err)

	// Save second queue (should overwrite)
	tracks2 := []domain.MusicTrack{
		{ID: "track2", FilePath: "/music/song2.mp3"},
		{ID: "track3", FilePath: "/music/song3.mp3"},
	}
	err = repo.SaveQueue(tracks2)
	require.NoError(t, err)

	// Load should return second queue
	loaded, err := repo.LoadQueue()
	require.NoError(t, err)
	require.Equal(t, 2, len(loaded))
	assert.Equal(t, "track2", loaded[0].ID)
	assert.Equal(t, "track3", loaded[1].ID)
}

func TestHistoryRepository_SaveAndLoadCurrentIndex(t *testing.T) {
	repo := newTestHistoryRepository()

	// Save index
	err := repo.SaveCurrentIndex(5)
	require.NoError(t, err)

	// Load index
	index, err := repo.LoadCurrentIndex()
	require.NoError(t, err)
	assert.Equal(t, 5, index)
}

func TestHistoryRepository_LoadCurrentIndex_NotSaved(t *testing.T) {
	repo := newTestHistoryRepository()

	// Load when nothing saved
	index, err := repo.LoadCurrentIndex()
	require.NoError(t, err)
	assert.Equal(t, -1, index) // Should return -1 when not set
}

func TestHistoryRepository_SaveCurrentIndex_Zero(t *testing.T) {
	repo := newTestHistoryRepository()

	// Save index 0 (valid index)
	err := repo.SaveCurrentIndex(0)
	require.NoError(t, err)

	// Load should return 0
	index, err := repo.LoadCurrentIndex()
	require.NoError(t, err)
	assert.Equal(t, 0, index)
}

func TestHistoryRepository_SaveCurrentIndex_Negative(t *testing.T) {
	repo := newTestHistoryRepository()

	// Save negative index (e.g., -1 for "no track")
	err := repo.SaveCurrentIndex(-1)
	require.NoError(t, err)

	// Load should return -1
	index, err := repo.LoadCurrentIndex()
	require.NoError(t, err)
	assert.Equal(t, -1, index)
}

func TestHistoryRepository_Clear(t *testing.T) {
	repo := newTestHistoryRepository()

	// Save some data
	tracks := []domain.MusicTrack{
		{ID: "track1", FilePath: "/music/song1.mp3"},
	}
	repo.SaveQueue(tracks)
	repo.SaveCurrentIndex(5)

	// Clear
	err := repo.Clear()
	require.NoError(t, err)

	// Verify cleared
	loaded, err := repo.LoadQueue()
	require.NoError(t, err)
	assert.Equal(t, 0, len(loaded))

	index, err := repo.LoadCurrentIndex()
	require.NoError(t, err)
	assert.Equal(t, -1, index)
}

func TestHistoryRepository_SaveLoadCycle(t *testing.T) {
	repo := newTestHistoryRepository()

	// Multiple save/load cycles
	for i := 0; i < 5; i++ {
		tracks := []domain.MusicTrack{
			{ID: "track1", FilePath: "/music/song1.mp3"},
		}

		err := repo.SaveQueue(tracks)
		require.NoError(t, err)

		loaded, err := repo.LoadQueue()
		require.NoError(t, err)
		assert.Equal(t, 1, len(loaded))
	}
}

func TestHistoryRepository_LargQueue(t *testing.T) {
	repo := newTestHistoryRepository()

	// Create large queue (1000 tracks)
	tracks := make([]domain.MusicTrack, 1000)
	for i := 0; i < 1000; i++ {
		tracks[i] = domain.MusicTrack{
			ID:       string(rune(i)),
			FilePath: "/music/song.mp3",
		}
	}

	// Save
	err := repo.SaveQueue(tracks)
	require.NoError(t, err)

	// Load
	loaded, err := repo.LoadQueue()
	require.NoError(t, err)
	assert.Equal(t, 1000, len(loaded))
}
