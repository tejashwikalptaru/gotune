package memory

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tejashwikalptaru/gotune/internal/domain"
)

// Helper to create a test playlist repository
func newTestPlaylistRepository() *PlaylistRepository {
	app := test.NewApp()
	prefs := app.Preferences()

	return NewPlaylistRepository(prefs)
}

func TestPlaylistRepository_SaveAndLoad(t *testing.T) {
	repo := newTestPlaylistRepository()

	// Create test playlist
	playlist := &domain.Playlist{
		ID:   "playlist1",
		Name: "My Favorites",
		Tracks: []domain.MusicTrack{
			{ID: "track1", FilePath: "/music/song1.mp3", Title: "Song 1"},
			{ID: "track2", FilePath: "/music/song2.mp3", Title: "Song 2"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save
	err := repo.Save(playlist)
	require.NoError(t, err)

	// Load
	loaded, err := repo.Load("playlist1")
	require.NoError(t, err)
	require.NotNil(t, loaded)

	// Verify
	assert.Equal(t, "playlist1", loaded.ID)
	assert.Equal(t, "My Favorites", loaded.Name)
	assert.Equal(t, 2, len(loaded.Tracks))
	assert.Equal(t, "track1", loaded.Tracks[0].ID)
	assert.Equal(t, "Song 1", loaded.Tracks[0].Title)
}

func TestPlaylistRepository_Load_NotFound(t *testing.T) {
	repo := newTestPlaylistRepository()

	// Try to load non-existent playlist
	_, err := repo.Load("nonexistent")
	assert.Error(t, err)
	// Should be ErrPlaylistNotFound
}

func TestPlaylistRepository_SaveOverwrites(t *testing.T) {
	repo := newTestPlaylistRepository()

	// Save first version
	playlist1 := &domain.Playlist{
		ID:   "playlist1",
		Name: "Original Name",
		Tracks: []domain.MusicTrack{
			{ID: "track1", FilePath: "/music/song1.mp3"},
		},
	}
	err := repo.Save(playlist1)
	require.NoError(t, err)

	// Save updated version with same ID
	playlist2 := &domain.Playlist{
		ID:   "playlist1",
		Name: "Updated Name",
		Tracks: []domain.MusicTrack{
			{ID: "track2", FilePath: "/music/song2.mp3"},
			{ID: "track3", FilePath: "/music/song3.mp3"},
		},
	}
	err = repo.Save(playlist2)
	require.NoError(t, err)

	// Load should return updated version
	loaded, err := repo.Load("playlist1")
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", loaded.Name)
	assert.Equal(t, 2, len(loaded.Tracks))
}

func TestPlaylistRepository_LoadAll_Empty(t *testing.T) {
	repo := newTestPlaylistRepository()

	// Load when nothing saved
	playlists, err := repo.LoadAll()
	require.NoError(t, err)
	assert.Equal(t, 0, len(playlists))
}

func TestPlaylistRepository_LoadAll_Multiple(t *testing.T) {
	repo := newTestPlaylistRepository()

	// Save multiple playlists
	for i := 1; i <= 3; i++ {
		playlist := &domain.Playlist{
			ID:   string(rune('0' + i)),
			Name: "Playlist " + string(rune('0'+i)),
			Tracks: []domain.MusicTrack{
				{ID: "track1", FilePath: "/music/song.mp3"},
			},
		}
		err := repo.Save(playlist)
		require.NoError(t, err)
	}

	// Load all
	playlists, err := repo.LoadAll()
	require.NoError(t, err)
	assert.Equal(t, 3, len(playlists))

	// Verify IDs are present (order doesn't matter)
	ids := make([]string, len(playlists))
	for i, p := range playlists {
		ids[i] = p.ID
	}
	assert.Contains(t, ids, "1")
	assert.Contains(t, ids, "2")
	assert.Contains(t, ids, "3")
}

func TestPlaylistRepository_Delete(t *testing.T) {
	repo := newTestPlaylistRepository()

	// Save playlist
	playlist := &domain.Playlist{
		ID:   "playlist1",
		Name: "Test",
	}
	err := repo.Save(playlist)
	require.NoError(t, err)

	// Verify exists
	assert.True(t, repo.Exists("playlist1"))

	// Delete
	err = repo.Delete("playlist1")
	require.NoError(t, err)

	// Verify deleted
	assert.False(t, repo.Exists("playlist1"))

	// Load should return error
	_, err = repo.Load("playlist1")
	assert.Error(t, err)
}

func TestPlaylistRepository_Delete_NotFound(t *testing.T) {
	repo := newTestPlaylistRepository()

	// Delete non-existent playlist (should be no-op, no error)
	err := repo.Delete("nonexistent")
	assert.NoError(t, err)
}

func TestPlaylistRepository_Exists(t *testing.T) {
	repo := newTestPlaylistRepository()

	// Initially doesn't exist
	assert.False(t, repo.Exists("playlist1"))

	// Save
	playlist := &domain.Playlist{
		ID:   "playlist1",
		Name: "Test",
	}
	repo.Save(playlist)

	// Now exists
	assert.True(t, repo.Exists("playlist1"))

	// Delete
	repo.Delete("playlist1")

	// No longer exists
	assert.False(t, repo.Exists("playlist1"))
}

func TestPlaylistRepository_SaveLoadCycle(t *testing.T) {
	repo := newTestPlaylistRepository()

	playlist := &domain.Playlist{
		ID:   "playlist1",
		Name: "Test",
	}

	// Multiple save/load cycles
	for i := 0; i < 5; i++ {
		err := repo.Save(playlist)
		require.NoError(t, err)

		loaded, err := repo.Load("playlist1")
		require.NoError(t, err)
		assert.Equal(t, "playlist1", loaded.ID)
	}
}

func TestPlaylistRepository_EmptyPlaylist(t *testing.T) {
	repo := newTestPlaylistRepository()

	// Save playlist with no tracks
	playlist := &domain.Playlist{
		ID:     "empty",
		Name:   "Empty Playlist",
		Tracks: []domain.MusicTrack{},
	}
	err := repo.Save(playlist)
	require.NoError(t, err)

	// Load
	loaded, err := repo.Load("empty")
	require.NoError(t, err)
	assert.Equal(t, 0, len(loaded.Tracks))
}

func TestPlaylistRepository_LargePlaylist(t *testing.T) {
	repo := newTestPlaylistRepository()

	// Create playlist with many tracks (1000)
	tracks := make([]domain.MusicTrack, 1000)
	for i := 0; i < 1000; i++ {
		tracks[i] = domain.MusicTrack{
			ID:       string(rune(i)),
			FilePath: "/music/song.mp3",
		}
	}

	playlist := &domain.Playlist{
		ID:     "large",
		Name:   "Large Playlist",
		Tracks: tracks,
	}

	// Save
	err := repo.Save(playlist)
	require.NoError(t, err)

	// Load
	loaded, err := repo.Load("large")
	require.NoError(t, err)
	assert.Equal(t, 1000, len(loaded.Tracks))
}

func TestPlaylistRepository_MultiplePlaylistsIndependent(t *testing.T) {
	repo := newTestPlaylistRepository()

	// Save multiple independent playlists
	playlist1 := &domain.Playlist{
		ID:   "p1",
		Name: "Playlist 1",
	}
	playlist2 := &domain.Playlist{
		ID:   "p2",
		Name: "Playlist 2",
	}

	repo.Save(playlist1)
	repo.Save(playlist2)

	// Verify both exist
	assert.True(t, repo.Exists("p1"))
	assert.True(t, repo.Exists("p2"))

	// Delete one
	repo.Delete("p1")

	// Verify only one deleted
	assert.False(t, repo.Exists("p1"))
	assert.True(t, repo.Exists("p2"))

	// Load other should still work
	loaded, err := repo.Load("p2")
	require.NoError(t, err)
	assert.Equal(t, "Playlist 2", loaded.Name)
}

func TestPlaylistRepository_LoadAllAfterDelete(t *testing.T) {
	repo := newTestPlaylistRepository()

	// Save 3 playlists
	for i := 1; i <= 3; i++ {
		playlist := &domain.Playlist{
			ID:   string(rune('0' + i)),
			Name: "Playlist",
		}
		repo.Save(playlist)
	}

	// Delete one
	repo.Delete("2")

	// LoadAll should return 2
	playlists, err := repo.LoadAll()
	require.NoError(t, err)
	assert.Equal(t, 2, len(playlists))
}

func TestPlaylistRepository_SpecialCharactersInName(t *testing.T) {
	repo := newTestPlaylistRepository()

	// Playlist with special characters
	playlist := &domain.Playlist{
		ID:   "special",
		Name: "Rock & Roll: The 90's \"Best\" of the decade!",
	}

	err := repo.Save(playlist)
	require.NoError(t, err)

	loaded, err := repo.Load("special")
	require.NoError(t, err)
	assert.Equal(t, "Rock & Roll: The 90's \"Best\" of the decade!", loaded.Name)
}
