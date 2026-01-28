// Package bass provides metadata extraction for audio files.
package bass

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dhowden/tag"
	"github.com/tejashwikalptaru/gotune/internal/domain"
)

// MOD file extensions
var modFormats = []string{
	".mod", ".xm", ".it", ".s3m", ".mtm", ".umx", ".mo3",
}

// isModFile checks if the file is a MOD/tracker format.
func isModFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	for _, modExt := range modFormats {
		if ext == modExt {
			return true
		}
	}
	return false
}

// extractMetadata extracts metadata from an audio file.
// This handles both MOD files and regular audio files.
func extractMetadata(filePath string) (*domain.MusicTrack, error) {
	if filePath == "" {
		return nil, domain.ErrInvalidFilePath
	}

	// Check if a file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, domain.ErrFileNotFound
	}

	filename := filepath.Base(filePath)
	ext := filepath.Ext(filePath)
	isMOD := isModFile(filePath)

	// Create a base track
	track := &domain.MusicTrack{
		ID:         generateTrackID(),
		FilePath:   filePath,
		Title:      filename,
		FileFormat: ext,
		IsMOD:      isMOD,
		Metadata:   &domain.TrackMetadata{},
	}

	if isMOD {
		// Extract MOD metadata
		return extractMODMetadata(track)
	}

	// Extract regular audio file metadata
	return extractAudioMetadata(track)
}

// extractMODMetadata extracts metadata from a MOD/tracker file.
func extractMODMetadata(track *domain.MusicTrack) (*domain.MusicTrack, error) {
	// Load the MOD file to extract tags
	handle, err := bassMusicLoad(track.FilePath, streamDecodeOnly|streamAutoFree)
	if err != nil {
		// If loading fails, return basic metadata
		return track, nil
	}
	defer bassMusicFree(handle)

	// Extract MOD-specific tags
	name := strings.TrimSpace(bassChannelGetTags(handle, TagMusicNAME))
	if name != "" {
		track.Title = name
	}

	author := strings.TrimSpace(bassChannelGetTags(handle, TagMusicAUTH))
	if author != "" {
		track.Artist = author
		track.Metadata.Composer = author
	}

	message := strings.TrimSpace(bassChannelGetTags(handle, TagMusicMESSAGE))
	if message != "" {
		track.Metadata.Comment = message
	}

	instrument := strings.TrimSpace(bassChannelGetTags(handle, TagMusicINST))
	if instrument != "" {
		// Store instrument info in comment if comment is empty
		if track.Metadata.Comment == "" {
			track.Metadata.Comment = "Instrument: " + instrument
		}
	}

	// Get duration
	lengthBytes := bassChannelGetLength(handle)
	track.Duration = bassChannelBytes2Seconds(handle, lengthBytes)

	return track, nil
}

// extractAudioMetadata extracts metadata from a regular audio file.
func extractAudioMetadata(track *domain.MusicTrack) (*domain.MusicTrack, error) {
	file, err := os.Open(track.FilePath)
	if err != nil {
		// If we can't open the file, return basic metadata
		return track, nil
	}
	defer file.Close()

	// Use dhowden/tag library to extract metadata
	metadata, err := tag.ReadFrom(file)
	if err != nil || metadata == nil {
		// If tag reading fails, return basic metadata
		return track, nil
	}

	// Extract metadata fields
	if title := strings.TrimSpace(metadata.Title()); title != "" {
		track.Title = title
	}

	if artist := strings.TrimSpace(metadata.Artist()); artist != "" {
		track.Artist = artist
	}

	if album := strings.TrimSpace(metadata.Album()); album != "" {
		track.Album = album
	}

	// Extended metadata
	track.Metadata.Composer = strings.TrimSpace(metadata.Composer())
	track.Metadata.Genre = strings.TrimSpace(metadata.Genre())

	if year := metadata.Year(); year > 0 {
		track.Metadata.Year = year
	}

	// Track and disc numbers
	trackNum, _ := metadata.Track()
	track.Metadata.TrackNumber = trackNum

	discNum, _ := metadata.Disc()
	track.Metadata.DiscNumber = discNum

	// Album art
	if picture := metadata.Picture(); picture != nil {
		track.Metadata.AlbumArt = picture.Data
	}

	// Format-specific metadata
	if format := metadata.Format(); format != tag.UnknownFormat {
		// Could extract bit rate, sample rate, etc. from the format if needed
		track.Metadata.SampleRate = 44100 // Default assumption
	}

	return track, nil
}

// generateTrackID generates a unique ID for a track
func generateTrackID() string {
	// Generate a random 8-byte hex string for uniqueness
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// If random generation fails, fall back to a timestamp-based ID
		b = []byte(fmt.Sprintf("%d", time.Now().UnixNano()))
	}
	return fmt.Sprintf("track-%s", hex.EncodeToString(b))
}
