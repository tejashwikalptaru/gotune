// Package eventbus provides implementations of the EventBus interface.
// This package contains the synchronous event bus implementation.
package eventbus

import (
	"fmt"
	"log/slog"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/tejashwikalptaru/gotune/internal/domain"
	"github.com/tejashwikalptaru/gotune/internal/ports"
)

// SyncEventBus is a synchronous implementation of the EventBus interface.
// Events are delivered to handlers synchronously in the order they were subscribed.
//
// Thread-safety: This implementation is thread-safe. Multiple goroutines can
// publish events and subscribe/unsubscribe handlers concurrently.
//
// Performance: Since handlers are called synchronously, slow handlers will block
// event delivery. Handlers should process events quickly or dispatch to a
// background goroutine if long processing is needed.
type SyncEventBus struct {
	// Dependencies
	logger *slog.Logger

	// subscribers map event types to their subscriptions
	subscribers map[domain.EventType][]subscription

	// allSubscribers contains handlers that receive all events
	allSubscribers []subscription

	// mu protects subscribers and allSubscribers
	mu sync.RWMutex

	// idCounter generates unique subscription IDs
	idCounter uint64

	// closed indicates if the event bus has been closed
	closed bool
}

// a subscription represents a single event subscription.
type subscription struct {
	id      domain.SubscriptionID
	handler domain.EventHandler
}

// NewSyncEventBus creates a new synchronous event bus.
func NewSyncEventBus() *SyncEventBus {
	return &SyncEventBus{
		subscribers:    make(map[domain.EventType][]subscription),
		allSubscribers: make([]subscription, 0),
		idCounter:      0,
	}
}

// SetLogger sets the logger for this event bus.
// This should be called after construction before using the event bus.
func (bus *SyncEventBus) SetLogger(logger *slog.Logger) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.logger = logger
}

// Publish publishes an event to all subscribers of that event type.
// Handlers are called synchronously in the order they subscribed.
//
// If the event bus is closed, this method does nothing.
//
// Panics in handlers are recovered and logged, but do not stop other handlers
// from being called.
func (bus *SyncEventBus) Publish(event domain.Event) {
	if event == nil {
		return
	}

	bus.mu.RLock()
	if bus.closed {
		bus.mu.RUnlock()
		return
	}

	// Get type-specific subscribers
	eventType := event.Type()
	typeSubscribers := make([]subscription, len(bus.subscribers[eventType]))
	copy(typeSubscribers, bus.subscribers[eventType])

	// Get wildcard subscribers
	wildcardSubscribers := make([]subscription, len(bus.allSubscribers))
	copy(wildcardSubscribers, bus.allSubscribers)

	bus.mu.RUnlock()

	// Call type-specific handlers
	for _, sub := range typeSubscribers {
		bus.callHandler(sub.handler, event)
	}

	// Call wildcard handlers
	for _, sub := range wildcardSubscribers {
		bus.callHandler(sub.handler, event)
	}
}

// callHandler calls an event handler and recovers from panics.
func (bus *SyncEventBus) callHandler(handler domain.EventHandler, event domain.Event) {
	defer func() {
		if r := recover(); r != nil {
			// Handler panicked - log it but don't crash
			if bus.logger != nil {
				bus.logger.Error("event handler panicked",
					slog.Any("panic", r),
					slog.String("event_type", string(event.Type())))
			}
		}
	}()

	if bus.logger != nil {
		handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
		bus.logger.Debug("event published",
			slog.String("event_type", string(event.Type())),
			slog.String("handler", handlerName))
	}
	handler(event)
}

// Subscribe registers a handler for events of the specified type.
// Returns a unique subscription ID that can be used to unsubscribe.
//
// The same handler can be registered multiple times with different IDs.
func (bus *SyncEventBus) Subscribe(eventType domain.EventType, handler domain.EventHandler) domain.SubscriptionID {
	if handler == nil {
		panic("event handler cannot be nil")
	}

	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus.closed {
		panic("cannot subscribe to closed event bus")
	}

	// Generate unique subscription ID
	id := domain.SubscriptionID(fmt.Sprintf("sub-%d", atomic.AddUint64(&bus.idCounter, 1)))

	// Create subscription
	sub := subscription{
		id:      id,
		handler: handler,
	}

	// Add to subscribers
	bus.subscribers[eventType] = append(bus.subscribers[eventType], sub)

	return id
}

// Unsubscribe removes a previously registered event handler.
// If the subscription ID is invalid or already unsubscribed, this is a no-op.
func (bus *SyncEventBus) Unsubscribe(id domain.SubscriptionID) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	// Search in type-specific subscribers
	for eventType, subs := range bus.subscribers {
		for i, sub := range subs {
			if sub.id == id {
				// Remove by replacing with the last element and truncating
				subs[i] = subs[len(subs)-1]
				bus.subscribers[eventType] = subs[:len(subs)-1]
				return
			}
		}
	}

	// Search for wildcard subscribers
	for i, sub := range bus.allSubscribers {
		if sub.id == id {
			// Remove by replacing with the last element and truncating
			bus.allSubscribers[i] = bus.allSubscribers[len(bus.allSubscribers)-1]
			bus.allSubscribers = bus.allSubscribers[:len(bus.allSubscribers)-1]
			return
		}
	}
}

// SubscribeAll registers a handler that receives all events regardless of type.
// Returns a unique subscription ID that can be used to unsubscribe.
//
// This is useful for logging, debugging, or analytics.
func (bus *SyncEventBus) SubscribeAll(handler domain.EventHandler) domain.SubscriptionID {
	if handler == nil {
		panic("event handler cannot be nil")
	}

	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus.closed {
		panic("cannot subscribe to closed event bus")
	}

	// Generate unique subscription ID
	id := domain.SubscriptionID(fmt.Sprintf("sub-all-%d", atomic.AddUint64(&bus.idCounter, 1)))

	// Create subscription
	sub := subscription{
		id:      id,
		handler: handler,
	}

	// Add to wildcard subscribers
	bus.allSubscribers = append(bus.allSubscribers, sub)

	return id
}

// HasSubscribers returns true if there are any active subscriptions for the given event type.
// This can be used to avoid expensive event construction if no one is listening.
func (bus *SyncEventBus) HasSubscribers(eventType domain.EventType) bool {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	// Check type-specific subscribers
	if len(bus.subscribers[eventType]) > 0 {
		return true
	}

	// Check wildcard subscribers
	return len(bus.allSubscribers) > 0
}

// Close shuts down the event bus and clears all subscriptions.
// After calling Close, no more events should be published or subscribed.
//
// Returns an error if already closed.
func (bus *SyncEventBus) Close() error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus.closed {
		return fmt.Errorf("event bus already closed")
	}

	bus.closed = true

	// Clear all subscriptions
	bus.subscribers = make(map[domain.EventType][]subscription)
	bus.allSubscribers = make([]subscription, 0)

	return nil
}

// SubscriberCount returns the number of active subscriptions for debugging.
// This counts both type-specific and wildcard subscriptions.
func (bus *SyncEventBus) SubscriberCount() int {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	count := len(bus.allSubscribers)
	for _, subs := range bus.subscribers {
		count += len(subs)
	}
	return count
}

// Verify that SyncEventBus implements the EventBus interface
var _ ports.EventBus = (*SyncEventBus)(nil)
