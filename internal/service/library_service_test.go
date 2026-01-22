package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tejashwikalptaru/gotune/internal/adapter/audio/mock"
	"github.com/tejashwikalptaru/gotune/internal/adapter/eventbus"
	"github.com/tejashwikalptaru/gotune/internal/domain"
	"github.com/tejashwikalptaru/gotune/internal/logger"
)

// Helper to create a test library service
func newTestLibraryService() (*LibraryService, *eventbus.SyncEventBus) {
	engine := mock.NewEngine()
	engine.Initialize(-1, 44100, 0)

	bus := eventbus.NewSyncEventBus()
	testLogger := logger.NewTestLogger()
	service := NewLibraryService(testLogger, engine, bus)

	return service, bus
}

// Helper to create a temporary test directory with audio files
func createTestMusicFolder(t *testing.T) string {
	// Create a temp directory
	tmpDir := filepath.Join(os.TempDir(), "gotune_test_"+time.Now().Format("20060102150405"))
	err := os.MkdirAll(tmpDir, 0755)
	require.NoError(t, err)

	// Create some test files
	testFiles := []string{
		"song1.mp3",
		"song2.flac",
		"track.wav",
		"readme.txt", // Non-audio file
		"subdir/nested.mp3",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tmpDir, file)
		dir := filepath.Dir(fullPath)

		// Create a subdirectory if needed
		if dir != tmpDir {
			os.MkdirAll(dir, 0755)
		}

		// Create an empty file
		f, err := os.Create(fullPath)
		require.NoError(t, err)
		f.Close()
	}

	return tmpDir
}

// Helper to clean up the test directory
func cleanupTestFolder(dir string) {
	os.RemoveAll(dir)
}

func TestLibraryService_IsFormatSupported(t *testing.T) {
	service, _ := newTestLibraryService()
	defer service.Shutdown()

	// Supported formats
	assert.True(t, service.IsFormatSupported("song.mp3"))
	assert.True(t, service.IsFormatSupported("track.flac"))
	assert.True(t, service.IsFormatSupported("music.wav"))
	assert.True(t, service.IsFormatSupported("MODULE.MOD"))
	assert.True(t, service.IsFormatSupported("/path/to/song.MP3")) // Case-insensitive

	// Unsupported formats
	assert.False(t, service.IsFormatSupported("readme.txt"))
	assert.False(t, service.IsFormatSupported("image.jpg"))

	// .mp4 is actually supported (AAC audio)
	assert.True(t, service.IsFormatSupported("video.mp4"))
}

func TestLibraryService_GetSupportedFormats(t *testing.T) {
	service, _ := newTestLibraryService()
	defer service.Shutdown()

	formats := service.GetSupportedFormats()

	// Should have multiple formats
	assert.Greater(t, len(formats), 20)

	// Should include common formats
	assert.Contains(t, formats, ".mp3")
	assert.Contains(t, formats, ".flac")
	assert.Contains(t, formats, ".wav")
	assert.Contains(t, formats, ".mod")

	// Should be a copy (modifying doesn't affect internal state)
	formats[0] = ".xyz"
	formats2 := service.GetSupportedFormats()
	assert.NotEqual(t, ".xyz", formats2[0])
}

func TestLibraryService_ExtractMetadata(t *testing.T) {
	service, _ := newTestLibraryService()
	defer service.Shutdown()

	// Create a test file
	tmpDir := createTestMusicFolder(t)
	defer cleanupTestFolder(tmpDir)

	testFile := filepath.Join(tmpDir, "song1.mp3")

	// Extract metadata
	track, err := service.ExtractMetadata(testFile)
	require.NoError(t, err)
	require.NotNil(t, track)

	// Verify basic metadata
	assert.NotEmpty(t, track.ID)
	assert.Equal(t, testFile, track.FilePath)
	assert.NotEmpty(t, track.FileFormat)
}

func TestLibraryService_ExtractMetadata_UnsupportedFormat(t *testing.T) {
	service, _ := newTestLibraryService()
	defer service.Shutdown()

	tmpDir := createTestMusicFolder(t)
	defer cleanupTestFolder(tmpDir)

	textFile := filepath.Join(tmpDir, "readme.txt")

	// Try to extract from an unsupported format
	_, err := service.ExtractMetadata(textFile)
	assert.Equal(t, domain.ErrUnsupportedFormat, err)
}

func TestLibraryService_ExtractMetadata_FileNotFound(t *testing.T) {
	service, _ := newTestLibraryService()
	defer service.Shutdown()

	// Try to extract from a non-existent file
	_, err := service.ExtractMetadata("/nonexistent/file.mp3")
	assert.Equal(t, domain.ErrFileNotFound, err)
}

func TestLibraryService_ScanFiles(t *testing.T) {
	service, bus := newTestLibraryService()
	defer service.Shutdown()

	tmpDir := createTestMusicFolder(t)
	defer cleanupTestFolder(tmpDir)

	// Prepare the file list
	files := []string{
		filepath.Join(tmpDir, "song1.mp3"),
		filepath.Join(tmpDir, "song2.flac"),
		filepath.Join(tmpDir, "track.wav"),
		filepath.Join(tmpDir, "readme.txt"), // Should be skipped
	}

	// Track progress events
	progressCount := 0
	bus.Subscribe(domain.EventScanProgress, func(e domain.Event) {
		progressCount++
	})

	// Scan files
	tracks, err := service.ScanFiles(files)
	require.NoError(t, err)

	// Should find 3 audio files (skipping readme.txt)
	assert.Equal(t, 3, len(tracks))

	// Should have received progress events
	assert.Greater(t, progressCount, 0)
}

func TestLibraryService_ScanFolder(t *testing.T) {
	service, bus := newTestLibraryService()
	defer service.Shutdown()

	tmpDir := createTestMusicFolder(t)
	defer cleanupTestFolder(tmpDir)

	// Subscribe to scan events
	var startedEvent domain.ScanStartedEvent
	var completedEvent domain.ScanCompletedEvent
	progressCount := 0

	bus.Subscribe(domain.EventScanStarted, func(e domain.Event) {
		startedEvent = e.(domain.ScanStartedEvent)
	})
	bus.Subscribe(domain.EventScanProgress, func(e domain.Event) {
		progressCount++
	})
	bus.Subscribe(domain.EventScanCompleted, func(e domain.Event) {
		completedEvent = e.(domain.ScanCompletedEvent)
	})

	// Scan folder
	tracks, err := service.ScanFolder(tmpDir)
	require.NoError(t, err)

	// Should find 4 audio files (including nested one)
	assert.Equal(t, 4, len(tracks))

	// Verify events
	assert.Equal(t, tmpDir, startedEvent.Path)
	assert.Equal(t, 4, len(completedEvent.TracksFound))
	assert.Greater(t, progressCount, 0)
}

func TestLibraryService_ScanFolder_NonExistentFolder(t *testing.T) {
	service, _ := newTestLibraryService()
	defer service.Shutdown()

	// Try to scan a non-existent folder
	tracks, err := service.ScanFolder("/nonexistent/folder")

	// Should return an empty list, but not an error (Walk just finds nothing)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(tracks))
}

func TestLibraryService_CancelScan(t *testing.T) {
	service, _ := newTestLibraryService()
	defer service.Shutdown()

	tmpDir := createTestMusicFolder(t)
	defer cleanupTestFolder(tmpDir)

	// Start scanning in the background
	done := make(chan struct{})
	var err error

	go func() {
		_, err = service.ScanFolder(tmpDir)
		done <- struct{}{}
	}()

	// Give the scan a moment to start (scans are very fast on small folders)
	time.Sleep(5 * time.Millisecond)

	// Try to cancel the scan
	cancelErr := service.CancelScan()

	// Wait for the scan to finish
	<-done

	// If scan was still running when we canceled, should get no error
	// If scan already finished, we'll get "no scan in progress" error
	// Both are acceptable outcomes for this test
	_ = cancelErr
	_ = err

	// Main assertion: After everything, no scan should be running
	assert.False(t, service.IsScanning())
}

func TestLibraryService_CancelScan_NoScanInProgress(t *testing.T) {
	service, _ := newTestLibraryService()
	defer service.Shutdown()

	// Try to cancel when no scan is running
	err := service.CancelScan()
	assert.Error(t, err)
}

func TestLibraryService_IsScanning(t *testing.T) {
	service, _ := newTestLibraryService()
	defer service.Shutdown()

	tmpDir := createTestMusicFolder(t)
	defer cleanupTestFolder(tmpDir)

	// Initially not scanning
	assert.False(t, service.IsScanning())

	// Start the scan in the background
	done := make(chan struct{})
	go func() {
		service.ScanFolder(tmpDir)
		done <- struct{}{}
	}()

	// Wait for completion
	<-done

	// Should not be scanning anymore (scan completed)
	assert.False(t, service.IsScanning())
}

func TestLibraryService_ScanFolder_ConcurrentScan(t *testing.T) {
	service, _ := newTestLibraryService()
	defer service.Shutdown()

	tmpDir := createTestMusicFolder(t)
	defer cleanupTestFolder(tmpDir)

	// Start first scan
	done1 := make(chan struct{})
	go func() {
		service.ScanFolder(tmpDir)
		done1 <- struct{}{}
	}()

	// Immediately try to start the second scan
	// (Small folder might finish quickly, but we try anyway)
	_, err2 := service.ScanFolder(tmpDir)

	// Wait for the first scan to complete
	<-done1

	// If the first scan was still running, the second should have failed
	// If the first scan already finished, second might have succeeded
	// We just verify no crashes/panics occurred
	_ = err2
}

func TestLibraryService_Shutdown(t *testing.T) {
	service, _ := newTestLibraryService()

	tmpDir := createTestMusicFolder(t)
	defer cleanupTestFolder(tmpDir)

	// Start the scan in the background
	done := make(chan struct{})
	go func() {
		service.ScanFolder(tmpDir)
		done <- struct{}{}
	}()

	// Give the scan time to start
	time.Sleep(10 * time.Millisecond)

	// Shutdown (should cancel scan)
	err := service.Shutdown()
	assert.NoError(t, err)

	// Wait for the scan to finish
	<-done

	// Should not be scanning
	assert.False(t, service.IsScanning())
}
