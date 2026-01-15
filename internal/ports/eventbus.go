// Package ports define the EventBus interface for event-driven communication.
// The event bus replaces callbacks and enables loose coupling between components.
package ports

import (
	"github.com/tejashwikalptaru/gotune/internal/domain"
)

// EventBus is the interface for publishing and subscribing to events.
// This is the core of the event-driven architecture, replacing the callback system.
//
// The event bus decouples event producers (services) from event consumers (UI, logging, etc.).
// Multiple subscribers can listen to the same event, and subscribers don't know about publishers.
//
// Thread-safety: Implementations must be thread-safe as events may be published and
// subscribed from multiple goroutines simultaneously.
//
// Example usage:
//
//	// In service: Publish an event
//	bus.Publish(domain.NewTrackStartedEvent(track))
//
//	// In UI presenter: Subscribe to events
//	subID := bus.Subscribe(domain.EventTrackStarted, func(event domain.Event) {
//	    e := event.(domain.TrackStartedEvent)
//	    ui.SetPlayState(true)
//	})
//
//	// Later: Unsubscribe
//	bus.Unsubscribe(subID)
type EventBus interface {
	// Publish publishes an event to all subscribers of that event type.
	// The event is delivered to handlers synchronously in the order they subscribed
	// (for synchronous implementations) or asynchronously (for async implementations).
	//
	// This method must not block for long periods. Handlers should process events quickly
	// or dispatch to a background goroutine if long processing is needed.
	Publish(event domain.Event)

	// Subscribe registers a handler for events of the specified type.
	// The handler will be called whenever an event of this type is published.
	//
	// The same handler can be registered multiple times, resulting in multiple calls.
	// Each subscription gets a unique SubscriptionID.
	//
	// eventType: The type of events to listen for (e.g., domain.EventTrackStarted)
	// handler: The function to call when an event is published
	//
	// Returns a SubscriptionID that can be used to unsubscribe later.
	Subscribe(eventType domain.EventType, handler domain.EventHandler) domain.SubscriptionID

	// Unsubscribe removes a previously registered event handler.
	// After unsubscribing, the handler will no longer receive events.
	//
	// If the subscription ID is invalid or already unsubscribed, this is a no-op.
	Unsubscribe(id domain.SubscriptionID)

	// SubscribeAll registers a handler that receives all events regardless of type.
	// This is useful for logging, debugging, or analytics.
	//
	// Returns a SubscriptionID that can be used to unsubscribe later.
	SubscribeAll(handler domain.EventHandler) domain.SubscriptionID

	// HasSubscribers returns true if there are any active subscriptions for the given event type.
	// This can be used to avoid expensive event construction if no one is listening.
	HasSubscribers(eventType domain.EventType) bool

	// Close shuts down the event bus and cleans up resources.
	// After calling Close, no more events should be published or subscribed.
	Close() error
}

// EventFilter is a function that determines if an event should be delivered to a subscriber.
// It returns true if the event should be delivered, false otherwise.
type EventFilter func(event domain.Event) bool

// FilteringEventBus extends EventBus with filtered subscriptions.
// This is optional and not all implementations need to support it.
type FilteringEventBus interface {
	EventBus

	// SubscribeFiltered registers a handler with a filter function.
	// The handler will only be called for events that pass the filter.
	//
	// Example: Only handle events for a specific track
	//	bus.SubscribeFiltered(domain.EventTrackProgress, func(e domain.Event) bool {
	//	    progress := e.(domain.TrackProgressEvent)
	//	    return progress.TrackID == myTrackID
	//	}, handleProgress)
	SubscribeFiltered(eventType domain.EventType, filter EventFilter, handler domain.EventHandler) domain.SubscriptionID
}
