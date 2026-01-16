// Package service provides business logic for the GoTune application.
package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tejashwikalptaru/gotune/internal/domain"
	"github.com/tejashwikalptaru/gotune/internal/ports"
)

// LibraryService handles music library operations including scanning and metadata extraction.
// All operations are thread-safe via sync.RWMutex.
type LibraryService struct {
	// Dependencies (injected)
	engine ports.AudioEngine
	bus    ports.EventBus

	// State
	scanning      bool
	cancelScan    context.CancelFunc
	scanContext   context.Context
	supportedExts []string

	// Concurrency control
	mu sync.RWMutex
}

// NewLibraryService creates a new library service.
func NewLibraryService(
	engine ports.AudioEngine,
	bus ports.EventBus,
) *LibraryService {
	return &LibraryService{
		engine: engine,
		bus:    bus,
		supportedExts: []string{
			// Common formats
			".mp3", ".mp2", ".mp1",
			".ogg", ".oga",
			".wav", ".aif", ".aiff",
			".flac", ".fla",
			".aac", ".m4a", ".m4b", ".mp4",
			".wma",
			".wv",          // WavPack
			".ape", ".mac", // APE
			".mpc", ".mp+", ".mpp", // Musepack
			".ofr", ".ofs", // OptimFROG
			".tta",         // TTA
			".adx", ".aix", // ADX
			".ac3", // AC3
			".cda", // CD Audio
			// MOD/Tracker formats
			".mod", ".xm", ".it", ".s3m", ".mtm", ".umx", ".mo3",
		},
	}
}

// ScanFolder scans a folder recursively for audio files and extracts metadata.
// Returns a list of tracks found. Publishes progress events during scanning.
func (s *LibraryService) ScanFolder(folderPath string) ([]domain.MusicTrack, error) {
	s.mu.Lock()
	if s.scanning {
		s.mu.Unlock()
		return nil, domain.NewServiceError("LibraryService", "ScanFolder", "scan already in progress", nil)
	}
	s.scanning = true

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	s.scanContext = ctx
	s.cancelScan = cancel
	s.mu.Unlock()

	// Ensure cleanup
	defer func() {
		s.mu.Lock()
		s.scanning = false
		s.scanContext = nil
		s.cancelScan = nil
		s.mu.Unlock()
	}()

	// Publish scan started event
	s.bus.Publish(domain.NewScanStartedEvent(folderPath))

	// Collect all audio files
	files, err := s.collectAudioFiles(ctx, folderPath)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			s.bus.Publish(domain.NewScanCancelledEvent("user cancelled"))
			return nil, domain.ErrScanCancelled
		}
		return nil, err
	}

	// Extract metadata for each file
	tracks := make([]domain.MusicTrack, 0, len(files))
	total := len(files)

	for i, filePath := range files {
		// Check for cancellation
		select {
		case <-ctx.Done():
			s.bus.Publish(domain.NewScanCancelledEvent("user cancelled"))
			return tracks, domain.ErrScanCancelled
		default:
		}

		// Extract metadata
		track, err := s.engine.GetMetadata(filePath)
		if err != nil {
			// Skip files that can't be read but continue scanning
			continue
		}

		if track != nil {
			tracks = append(tracks, *track)
		}

		// Publish progress event
		progress := domain.ScanProgress{
			CurrentFile:  filePath,
			FilesScanned: i + 1,
			TotalFiles:   total,
			TracksFound:  len(tracks),
		}
		s.bus.Publish(domain.NewScanProgressEvent(progress))
	}

	// Publish scan completed event
	s.bus.Publish(domain.NewScanCompletedEvent(tracks))

	return tracks, nil
}

// ScanFiles scans specific files (not a folder) and extracts metadata.
func (s *LibraryService) ScanFiles(filePaths []string) ([]domain.MusicTrack, error) {
	s.mu.Lock()
	if s.scanning {
		s.mu.Unlock()
		return nil, domain.NewServiceError("LibraryService", "ScanFiles", "scan already in progress", nil)
	}
	s.scanning = true

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	s.scanContext = ctx
	s.cancelScan = cancel
	s.mu.Unlock()

	// Ensure cleanup
	defer func() {
		s.mu.Lock()
		s.scanning = false
		s.scanContext = nil
		s.cancelScan = nil
		s.mu.Unlock()
	}()

	tracks := make([]domain.MusicTrack, 0, len(filePaths))
	total := len(filePaths)

	for i, filePath := range filePaths {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return tracks, domain.ErrScanCancelled
		default:
		}

		// Skip unsupported formats
		if !s.IsFormatSupported(filePath) {
			continue
		}

		// Extract metadata
		track, err := s.engine.GetMetadata(filePath)
		if err != nil {
			// Skip files that can't be read
			continue
		}

		if track != nil {
			tracks = append(tracks, *track)
		}

		// Publish progress
		progress := domain.ScanProgress{
			CurrentFile:  filePath,
			FilesScanned: i + 1,
			TotalFiles:   total,
			TracksFound:  len(tracks),
		}
		s.bus.Publish(domain.NewScanProgressEvent(progress))
	}

	return tracks, nil
}

// CancelScan cancels the currently running scan operation.
func (s *LibraryService) CancelScan() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.scanning {
		return domain.NewServiceError("LibraryService", "CancelScan", "no scan in progress", nil)
	}

	if s.cancelScan != nil {
		s.cancelScan()
	}

	return nil
}

// IsScanning returns true if a scan is currently in progress.
func (s *LibraryService) IsScanning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scanning
}

// IsFormatSupported checks if a file format is supported.
func (s *LibraryService) IsFormatSupported(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	for _, supported := range s.supportedExts {
		if ext == supported {
			return true
		}
	}
	return false
}

// GetSupportedFormats returns the list of supported file extensions.
func (s *LibraryService) GetSupportedFormats() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy
	formats := make([]string, len(s.supportedExts))
	copy(formats, s.supportedExts)
	return formats
}

// collectAudioFiles recursively collects all audio files in a directory.
func (s *LibraryService) collectAudioFiles(ctx context.Context, folderPath string) ([]string, error) {
	files := make([]string, 0)

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		if err != nil {
			// Skip files/folders we can't access
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if supported a format
		if s.IsFormatSupported(path) {
			files = append(files, path)
		}

		return nil
	})

	if errors.Is(err, context.Canceled) {
		return files, context.Canceled
	}

	return files, err
}

// ExtractMetadata extracts metadata for a single file.
func (s *LibraryService) ExtractMetadata(filePath string) (*domain.MusicTrack, error) {
	if !s.IsFormatSupported(filePath) {
		return nil, domain.ErrUnsupportedFormat
	}

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, domain.ErrFileNotFound
	}

	return s.engine.GetMetadata(filePath)
}

// Shutdown cleans up resources.
func (s *LibraryService) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cancel any running scan
	if s.scanning && s.cancelScan != nil {
		s.cancelScan()
	}

	return nil
}

// Verify that LibraryService implements the expected interface patterns
var _ interface {
	ScanFolder(string) ([]domain.MusicTrack, error)
	ScanFiles([]string) ([]domain.MusicTrack, error)
	CancelScan() error
	IsScanning() bool
	IsFormatSupported(string) bool
	GetSupportedFormats() []string
	ExtractMetadata(string) (*domain.MusicTrack, error)
	Shutdown() error
} = (*LibraryService)(nil)
