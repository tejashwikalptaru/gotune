// Package domain defines events for the event-driven architecture.
// Events replace the callback system and enable loose coupling between components.
package domain

import (
	"time"
)

// Event is the base interface for all events in the system.
// All events must implement this interface to be published via the event bus.
type Event interface {
	// Type returns the event type identifier
	Type() EventType

	// Timestamp returns when the event occurred
	Timestamp() time.Time
}

// EventType is a string identifier for different event types.
type EventType string

// Event type constants define all possible events in the system.
const (
	// Playback events
	EventTrackLoaded    EventType = "track.loaded"
	EventTrackStarted   EventType = "track.started"
	EventTrackPaused    EventType = "track.paused"
	EventTrackStopped   EventType = "track.stopped"
	EventTrackCompleted EventType = "track.completed"
	EventTrackProgress  EventType = "track.progress"
	EventTrackError     EventType = "track.error"
	EventAutoNext       EventType = "track.auto_next"

	// Volume events
	EventVolumeChanged EventType = "volume.changed"
	EventMuteToggled   EventType = "mute.toggled"

	// Playback mode events
	EventLoopToggled EventType = "loop.toggled"

	// Queue/Playlist events
	EventPlaylistUpdated EventType = "playlist.updated"
	EventQueueChanged    EventType = "queue.changed"
	EventTrackAdded      EventType = "track.added"

	// Library scanning events
	EventScanStarted   EventType = "scan.started"
	EventScanProgress  EventType = "scan.progress"
	EventScanCompleted EventType = "scan.completed"
	EventScanCancelled EventType = "scan.cancelled"
)

// EventHandler is a function that handles events.
type EventHandler func(event Event)

// SubscriptionID uniquely identifies an event subscription.
type SubscriptionID string

// baseEvent provides common event functionality.
// All concrete events should embed this struct.
type baseEvent struct {
	timestamp time.Time
}

// Timestamp returns when the event occurred.
func (e baseEvent) Timestamp() time.Time {
	return e.timestamp
}

// newBaseEvent creates a new base event with the current timestamp.
func newBaseEvent() baseEvent {
	return baseEvent{timestamp: time.Now()}
}

// TrackLoadedEvent is published when a track is successfully loaded.
type TrackLoadedEvent struct {
	baseEvent
	Track    MusicTrack
	Handle   TrackHandle
	Duration time.Duration
	Index    int // Queue index
}

// Type returns the event type.
func (e TrackLoadedEvent) Type() EventType {
	return EventTrackLoaded
}

// NewTrackLoadedEvent creates a new TrackLoadedEvent.
func NewTrackLoadedEvent(track MusicTrack, handle TrackHandle, duration time.Duration, index int) TrackLoadedEvent {
	return TrackLoadedEvent{
		baseEvent: newBaseEvent(),
		Track:     track,
		Handle:    handle,
		Duration:  duration,
		Index:     index,
	}
}

// TrackStartedEvent is published when playback starts.
type TrackStartedEvent struct {
	baseEvent
	Track MusicTrack
}

// Type returns the event type.
func (e TrackStartedEvent) Type() EventType {
	return EventTrackStarted
}

// NewTrackStartedEvent creates a new TrackStartedEvent.
func NewTrackStartedEvent(track MusicTrack) TrackStartedEvent {
	return TrackStartedEvent{
		baseEvent: newBaseEvent(),
		Track:     track,
	}
}

// TrackPausedEvent is published when playback is paused.
type TrackPausedEvent struct {
	baseEvent
	Track    MusicTrack
	Position time.Duration
}

// Type returns the event type.
func (e TrackPausedEvent) Type() EventType {
	return EventTrackPaused
}

// NewTrackPausedEvent creates a new TrackPausedEvent.
func NewTrackPausedEvent(track MusicTrack, position time.Duration) TrackPausedEvent {
	return TrackPausedEvent{
		baseEvent: newBaseEvent(),
		Track:     track,
		Position:  position,
	}
}

// TrackStoppedEvent is published when playback is stopped.
type TrackStoppedEvent struct {
	baseEvent
	Track MusicTrack
}

// Type returns the event type.
func (e TrackStoppedEvent) Type() EventType {
	return EventTrackStopped
}

// NewTrackStoppedEvent creates a new TrackStoppedEvent.
func NewTrackStoppedEvent(track MusicTrack) TrackStoppedEvent {
	return TrackStoppedEvent{
		baseEvent: newBaseEvent(),
		Track:     track,
	}
}

// TrackCompletedEvent is published when a track finishes playing naturally.
type TrackCompletedEvent struct {
	baseEvent
	Track MusicTrack
}

// Type returns the event type.
func (e TrackCompletedEvent) Type() EventType {
	return EventTrackCompleted
}

// NewTrackCompletedEvent creates a new TrackCompletedEvent.
func NewTrackCompletedEvent(track MusicTrack) TrackCompletedEvent {
	return TrackCompletedEvent{
		baseEvent: newBaseEvent(),
		Track:     track,
	}
}

// TrackProgressEvent is published periodically during playback.
type TrackProgressEvent struct {
	baseEvent
	Position time.Duration
	Duration time.Duration
}

// Type returns the event type.
func (e TrackProgressEvent) Type() EventType {
	return EventTrackProgress
}

// NewTrackProgressEvent creates a new TrackProgressEvent.
func NewTrackProgressEvent(position, duration time.Duration) TrackProgressEvent {
	return TrackProgressEvent{
		baseEvent: newBaseEvent(),
		Position:  position,
		Duration:  duration,
	}
}

// VolumeChangedEvent is published when the volume changes.
type VolumeChangedEvent struct {
	baseEvent
	Volume float64 // 0.0 to 1.0
}

// Type returns the event type.
func (e VolumeChangedEvent) Type() EventType {
	return EventVolumeChanged
}

// NewVolumeChangedEvent creates a new VolumeChangedEvent.
func NewVolumeChangedEvent(volume float64) VolumeChangedEvent {
	return VolumeChangedEvent{
		baseEvent: newBaseEvent(),
		Volume:    volume,
	}
}

// MuteToggledEvent is published when mute is toggled.
type MuteToggledEvent struct {
	baseEvent
	Muted bool
}

// Type returns the event type.
func (e MuteToggledEvent) Type() EventType {
	return EventMuteToggled
}

// NewMuteToggledEvent creates a new MuteToggledEvent.
func NewMuteToggledEvent(muted bool) MuteToggledEvent {
	return MuteToggledEvent{
		baseEvent: newBaseEvent(),
		Muted:     muted,
	}
}

// LoopToggledEvent is published when loop mode is toggled.
type LoopToggledEvent struct {
	baseEvent
	Enabled bool
}

// Type returns the event type.
func (e LoopToggledEvent) Type() EventType {
	return EventLoopToggled
}

// NewLoopToggledEvent creates a new LoopToggledEvent.
func NewLoopToggledEvent(enabled bool) LoopToggledEvent {
	return LoopToggledEvent{
		baseEvent: newBaseEvent(),
		Enabled:   enabled,
	}
}

// PlaylistUpdatedEvent is published when the playlist changes.
type PlaylistUpdatedEvent struct {
	baseEvent
	Playlist []MusicTrack
	Index    int // Current track index
}

// Type returns the event type.
func (e PlaylistUpdatedEvent) Type() EventType {
	return EventPlaylistUpdated
}

// NewPlaylistUpdatedEvent creates a new PlaylistUpdatedEvent.
func NewPlaylistUpdatedEvent(playlist []MusicTrack, index int) PlaylistUpdatedEvent {
	return PlaylistUpdatedEvent{
		baseEvent: newBaseEvent(),
		Playlist:  playlist,
		Index:     index,
	}
}

// QueueChangedEvent is published when the queue changes.
type QueueChangedEvent struct {
	baseEvent
	Queue []MusicTrack
}

// Type returns the event type.
func (e QueueChangedEvent) Type() EventType {
	return EventQueueChanged
}

// NewQueueChangedEvent creates a new QueueChangedEvent.
func NewQueueChangedEvent(queue []MusicTrack) QueueChangedEvent {
	return QueueChangedEvent{
		baseEvent: newBaseEvent(),
		Queue:     queue,
	}
}

// TrackAddedEvent is published when a track is added to the queue.
type TrackAddedEvent struct {
	baseEvent
	Track MusicTrack
	Index int
}

// Type returns the event type.
func (e TrackAddedEvent) Type() EventType {
	return EventTrackAdded
}

// NewTrackAddedEvent creates a new TrackAddedEvent.
func NewTrackAddedEvent(track MusicTrack, index int) TrackAddedEvent {
	return TrackAddedEvent{
		baseEvent: newBaseEvent(),
		Track:     track,
		Index:     index,
	}
}

// ScanStartedEvent is published when a library scan starts.
type ScanStartedEvent struct {
	baseEvent
	Path string
}

// Type returns the event type.
func (e ScanStartedEvent) Type() EventType {
	return EventScanStarted
}

// NewScanStartedEvent creates a new ScanStartedEvent.
func NewScanStartedEvent(path string) ScanStartedEvent {
	return ScanStartedEvent{
		baseEvent: newBaseEvent(),
		Path:      path,
	}
}

// ScanProgressEvent is published periodically during a library scan.
type ScanProgressEvent struct {
	baseEvent
	Progress ScanProgress
}

// Type returns the event type.
func (e ScanProgressEvent) Type() EventType {
	return EventScanProgress
}

// NewScanProgressEvent creates a new ScanProgressEvent.
func NewScanProgressEvent(progress ScanProgress) ScanProgressEvent {
	return ScanProgressEvent{
		baseEvent: newBaseEvent(),
		Progress:  progress,
	}
}

// ScanCompletedEvent is published when a library scan completes.
type ScanCompletedEvent struct {
	baseEvent
	TracksFound []MusicTrack
}

// Type returns the event type.
func (e ScanCompletedEvent) Type() EventType {
	return EventScanCompleted
}

// NewScanCompletedEvent creates a new ScanCompletedEvent.
func NewScanCompletedEvent(tracks []MusicTrack) ScanCompletedEvent {
	return ScanCompletedEvent{
		baseEvent:   newBaseEvent(),
		TracksFound: tracks,
	}
}

// ScanCancelledEvent is published when a library scan is canceled.
type ScanCancelledEvent struct {
	baseEvent
	Reason string
}

// Type returns the event type.
func (e ScanCancelledEvent) Type() EventType {
	return EventScanCancelled
}

// NewScanCancelledEvent creates a new ScanCancelledEvent.
func NewScanCancelledEvent(reason string) ScanCancelledEvent {
	return ScanCancelledEvent{
		baseEvent: newBaseEvent(),
		Reason:    reason,
	}
}

// TrackErrorEvent is published when an error occurs with a track.
type TrackErrorEvent struct {
	baseEvent
	Track MusicTrack
	Error error
}

// Type returns the event type.
func (e TrackErrorEvent) Type() EventType {
	return EventTrackError
}

// NewTrackErrorEvent creates a new TrackErrorEvent.
func NewTrackErrorEvent(track MusicTrack, err error) TrackErrorEvent {
	return TrackErrorEvent{
		baseEvent: newBaseEvent(),
		Track:     track,
		Error:     err,
	}
}

// AutoNextEvent is published when a track finishes and the playlist should auto-advance.
// This is used by the PlaybackService to signal the PlaylistService.
type AutoNextEvent struct {
	baseEvent
	Track        MusicTrack
	CurrentIndex int
}

// Type returns the event type.
func (e AutoNextEvent) Type() EventType {
	return EventAutoNext
}

// NewAutoNextEvent creates a new AutoNextEvent.
func NewAutoNextEvent(track MusicTrack, index int) AutoNextEvent {
	return AutoNextEvent{
		baseEvent:    newBaseEvent(),
		Track:        track,
		CurrentIndex: index,
	}
}
