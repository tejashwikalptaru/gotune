// Package domain contains core business models and logic with no external dependencies.
// This package defines the fundamental entities of the GoTune music player.
package domain

import (
	"time"
)

// MusicTrack represents a single audio track with all its metadata.
// This is the core domain model for individual music files.
type MusicTrack struct {
	// ID is a unique identifier for the track (UUID)
	ID string

	// FilePath is the absolute path to the audio file on the filesystem
	FilePath string

	// Title is the song title (from metadata or filename)
	Title string

	// Artist is the performing artist name
	Artist string

	// Album is the album name
	Album string

	// Duration is the total length of the track
	Duration time.Duration

	// FileFormat is the file extension (mp3, flac, ogg, etc.)
	FileFormat string

	// IsMOD indicates if this is a tracker module file (MOD, XM, IT, S3M)
	IsMOD bool

	// Metadata contains additional track information
	Metadata *TrackMetadata
}

// TrackMetadata contains extended metadata for an audio track.
type TrackMetadata struct {
	// Composer is the song composer
	Composer string

	// Genre is the music genre
	Genre string

	// Year is the release year
	Year int

	// AlbumArt is the embedded album artwork as raw bytes
	AlbumArt []byte

	// BitRate is the audio bit rate in kbps
	BitRate int

	// SampleRate is the audio sample rate in Hz
	SampleRate int

	// TrackNumber is the track number on the album
	TrackNumber int

	// DiscNumber is the disc number for multi-disc albums
	DiscNumber int

	// Comment contains any additional metadata comments
	Comment string
}

// Playlist represents a collection of music tracks.
type Playlist struct {
	// ID is a unique identifier for the playlist (UUID)
	ID string

	// Name is the playlist name
	Name string

	// Tracks is the ordered list of tracks in the playlist
	Tracks []MusicTrack

	// CreatedAt is when the playlist was created
	CreatedAt time.Time

	// UpdatedAt is when the playlist was last modified
	UpdatedAt time.Time
}

// PlaybackState represents the current state of the music player.
// This is the central state object that services manage.
type PlaybackState struct {
	// CurrentTrack is the currently loaded track (nil if none)
	CurrentTrack *MusicTrack

	// CurrentIndex is the index in the queue (0-based, -1 if no track)
	CurrentIndex int

	// Queue is the current playback queue
	Queue []MusicTrack

	// Status is the current playback status
	Status PlaybackStatus

	// Position is the current playback position within the track
	Position time.Duration

	// Volume is the current volume level (0.0 to 1.0)
	Volume float64

	// IsMuted indicates if audio is muted
	IsMuted bool

	// IsLooping indicates if the current track should loop
	IsLooping bool
}

// PlaybackStatus represents the current playback state.
type PlaybackStatus int

const (
	// StatusStopped indicates playback is stopped
	StatusStopped PlaybackStatus = iota

	// StatusPlaying indicates playback is active
	StatusPlaying

	// StatusPaused indicates playback is paused
	StatusPaused

	// StatusStalled indicates playback is stalled/buffering
	StatusStalled
)

// String returns a human-readable representation of the playback status.
func (s PlaybackStatus) String() string {
	switch s {
	case StatusStopped:
		return "stopped"
	case StatusPlaying:
		return "playing"
	case StatusPaused:
		return "paused"
	case StatusStalled:
		return "stalled"
	default:
		return "unknown"
	}
}

// Preferences contain user preferences and settings.
type Preferences struct {
	// Volume is the saved volume level (0.0 to 1.0)
	Volume float64

	// LoopEnabled indicates if loop mode is enabled by default
	LoopEnabled bool

	// Theme is the UI theme (dark, light, system)
	Theme string

	// ScanPaths are directories to scan for music
	ScanPaths []string

	// LastQueueIndex is the last played track index
	LastQueueIndex int

	// LastQueue is the last played queue
	LastQueue []MusicTrack
}

// TrackHandle represents a handle to an audio track in the audio engine.
// This is an opaque identifier used by the audio engine to reference loaded tracks.
type TrackHandle int64

const (
	// InvalidTrackHandle represents an invalid or uninitialized track handle
	InvalidTrackHandle TrackHandle = 0
)

// ScanProgress represents the progress of a music library scan operation.
type ScanProgress struct {
	// CurrentFile is the file currently being scanned
	CurrentFile string

	// FilesScanned is the number of files processed so far
	FilesScanned int

	// TotalFiles is the total number of files to scan (may be -1 if unknown)
	TotalFiles int

	// TracksFound is the number of valid music tracks found
	TracksFound int
}

// IsValid returns true if the scan progress has valid data.
func (p ScanProgress) IsValid() bool {
	return p.FilesScanned >= 0 && p.TracksFound >= 0
}

// Percentage returns the completion percentage (0-100), or -1 if total is unknown.
func (p ScanProgress) Percentage() float64 {
	if p.TotalFiles <= 0 {
		return -1
	}
	return float64(p.FilesScanned) / float64(p.TotalFiles) * 100.0
}
