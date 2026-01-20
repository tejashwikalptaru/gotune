// Package domain defines domain-specific errors.
// These errors represent business logic failures and are independent of infrastructure.
package domain

import (
	"errors"
	"fmt"
)

// Common errors that services can return.
var (
	// ErrTrackNotFound is returned when a requested track cannot be found.
	ErrTrackNotFound = errors.New("track not found")

	// ErrInvalidTrackHandle is returned when an invalid track handle is used.
	ErrInvalidTrackHandle = errors.New("invalid track handle")

	// ErrPlaylistEmpty is returned when an operation requires a non-empty playlist.
	ErrPlaylistEmpty = errors.New("playlist is empty")

	// ErrQueueEmpty is returned when queue operations are attempted on an empty queue.
	ErrQueueEmpty = errors.New("queue is empty")

	// ErrEndOfQueue is returned when trying to navigate past the end of the queue.
	ErrEndOfQueue = errors.New("end of queue reached")

	// ErrStartOfQueue is returned when trying to navigate before the start of the queue.
	ErrStartOfQueue = errors.New("start of queue reached")

	// ErrInvalidIndex is returned when a queue index is out of bounds.
	ErrInvalidIndex = errors.New("invalid queue index")

	// ErrInvalidVolume is returned when the volume is out of valid range (0.0-1.0).
	ErrInvalidVolume = errors.New("invalid volume: must be between 0.0 and 1.0")

	// ErrInvalidPosition is returned when seeking to an invalid position.
	ErrInvalidPosition = errors.New("invalid playback position")

	// ErrNotInitialized is returned when an operation is attempted on an uninitialized component.
	ErrNotInitialized = errors.New("component not initialized")

	// ErrAlreadyInitialized is returned when attempting to initialize an already initialized component.
	ErrAlreadyInitialized = errors.New("component already initialized")

	// ErrUnsupportedFormat is returned when an audio file format is not supported.
	ErrUnsupportedFormat = errors.New("unsupported audio format")

	// ErrFileNotFound is returned when a file does not exist.
	ErrFileNotFound = errors.New("file not found")

	// ErrInvalidFilePath is returned when a file path is invalid.
	ErrInvalidFilePath = errors.New("invalid file path")

	// ErrDuplicateTrack is returned when attempting to add a track that already exists in the queue.
	ErrDuplicateTrack = errors.New("track already exists in queue")

	// ErrScanCancelled is returned when a library scan is canceled.
	ErrScanCancelled = errors.New("scan cancelled")

	// ErrNoTrackLoaded is returned when playback is attempted with no track loaded.
	ErrNoTrackLoaded = errors.New("no track loaded")

	// ErrPlaybackFailed is returned when playback cannot be started.
	ErrPlaybackFailed = errors.New("playback failed")
)

// AudioEngineError represents an error from the audio engine.
// This wraps low-level audio library errors with additional context.
type AudioEngineError struct {
	Op      string // Operation that failed (e.g., "load", "play", "stop")
	Path    string // File path (if applicable)
	Code    int    // Error code from an underlying library
	Message string // Error message
	Err     error  // Underlying error (if any)
}

// Error implements the error interface.
func (e *AudioEngineError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("audio engine %s failed for '%s': %s (code: %d)", e.Op, e.Path, e.Message, e.Code)
	}
	return fmt.Sprintf("audio engine %s failed: %s (code: %d)", e.Op, e.Message, e.Code)
}

// Unwrap returns the underlying error.
func (e *AudioEngineError) Unwrap() error {
	return e.Err
}

// NewAudioEngineError creates a new AudioEngineError.
func NewAudioEngineError(op, path string, code int, message string, err error) *AudioEngineError {
	return &AudioEngineError{
		Op:      op,
		Path:    path,
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// RepositoryError represents an error from a repository.
// This wraps persistence layer errors with additional context.
type RepositoryError struct {
	Op      string // Operation that failed (e.g., "save", "load", "delete")
	Type    string // Repository type (e.g., "history", "playlist", "preferences")
	Message string // Error message
	Err     error  // Underlying error
}

// Error implements the error interface.
func (e *RepositoryError) Error() string {
	return fmt.Sprintf("repository %s.%s failed: %s", e.Type, e.Op, e.Message)
}

// Unwrap returns the underlying error.
func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// NewRepositoryError creates a new RepositoryError.
func NewRepositoryError(op, repoType, message string, err error) *RepositoryError {
	return &RepositoryError{
		Op:      op,
		Type:    repoType,
		Message: message,
		Err:     err,
	}
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string      // Field that failed validation
	Value   interface{} // Value that failed validation
	Message string      // Error message
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for %s: %s (value: %v)", e.Field, e.Message, e.Value)
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field string, value interface{}, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// ServiceError represents an error from a service layer operation.
type ServiceError struct {
	Service string // Service name (e.g., "PlaybackService", "PlaylistService")
	Op      string // Operation that failed
	Message string // Error message
	Err     error  // Underlying error
}

// Error implements the error interface.
func (e *ServiceError) Error() string {
	return fmt.Sprintf("service %s.%s failed: %s", e.Service, e.Op, e.Message)
}

// Unwrap returns the underlying error.
func (e *ServiceError) Unwrap() error {
	return e.Err
}

// NewServiceError creates a new ServiceError.
func NewServiceError(service, op, message string, err error) *ServiceError {
	return &ServiceError{
		Service: service,
		Op:      op,
		Message: message,
		Err:     err,
	}
}
