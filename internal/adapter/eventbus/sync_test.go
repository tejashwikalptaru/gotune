package eventbus

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tejashwikalptaru/gotune/internal/domain"
)

// TestNewSyncEventBus tests event bus creation.
func TestNewSyncEventBus(t *testing.T) {
	bus := NewSyncEventBus()

	if bus == nil {
		t.Fatal("NewSyncEventBus returned nil")
	}

	if bus.SubscriberCount() != 0 {
		t.Errorf("Expected 0 subscribers, got %d", bus.SubscriberCount())
	}

	if bus.closed {
		t.Error("New event bus should not be closed")
	}
}

// TestPublishSubscribe tests basic publish/subscribe functionality.
func TestPublishSubscribe(t *testing.T) {
	bus := NewSyncEventBus()
	defer bus.Close()

	var received domain.Event
	var callCount int

	handler := func(event domain.Event) {
		received = event
		callCount++
	}

	// Subscribe to track started events
	subID := bus.Subscribe(domain.EventTrackStarted, handler)

	if subID == "" {
		t.Fatal("Subscribe returned empty subscription ID")
	}

	// Publish a track started event
	track := domain.MusicTrack{ID: "test123", Title: "Test Track"}
	event := domain.NewTrackStartedEvent(track)
	bus.Publish(event)

	// Verify handler was called
	if callCount != 1 {
		t.Errorf("Expected handler to be called once, got %d", callCount)
	}

	if received == nil {
		t.Fatal("Handler did not receive event")
	}

	if received.Type() != domain.EventTrackStarted {
		t.Errorf("Expected EventTrackStarted, got %s", received.Type())
	}

	// Verify event data
	receivedEvent := received.(domain.TrackStartedEvent)
	if receivedEvent.Track.ID != "test123" {
		t.Errorf("Expected track ID test123, got %s", receivedEvent.Track.ID)
	}
}

// TestMultipleSubscribers tests multiple handlers for the same event type.
func TestMultipleSubscribers(t *testing.T) {
	bus := NewSyncEventBus()
	defer bus.Close()

	var callCount1, callCount2, callCount3 int32

	handler1 := func(event domain.Event) {
		atomic.AddInt32(&callCount1, 1)
	}

	handler2 := func(event domain.Event) {
		atomic.AddInt32(&callCount2, 1)
	}

	handler3 := func(event domain.Event) {
		atomic.AddInt32(&callCount3, 1)
	}

	// Subscribe multiple handlers
	bus.Subscribe(domain.EventTrackStarted, handler1)
	bus.Subscribe(domain.EventTrackStarted, handler2)
	bus.Subscribe(domain.EventTrackStarted, handler3)

	// Publish event
	track := domain.MusicTrack{ID: "test", Title: "Test"}
	bus.Publish(domain.NewTrackStartedEvent(track))

	// All handlers should be called
	if atomic.LoadInt32(&callCount1) != 1 {
		t.Errorf("Handler 1: expected 1 call, got %d", callCount1)
	}
	if atomic.LoadInt32(&callCount2) != 1 {
		t.Errorf("Handler 2: expected 1 call, got %d", callCount2)
	}
	if atomic.LoadInt32(&callCount3) != 1 {
		t.Errorf("Handler 3: expected 1 call, got %d", callCount3)
	}
}

// TestUnsubscribe tests unsubscribing handlers.
func TestUnsubscribe(t *testing.T) {
	bus := NewSyncEventBus()
	defer bus.Close()

	var callCount int32

	handler := func(event domain.Event) {
		atomic.AddInt32(&callCount, 1)
	}

	// Subscribe
	subID := bus.Subscribe(domain.EventTrackStarted, handler)

	// Publish - handler should be called
	track := domain.MusicTrack{ID: "test", Title: "Test"}
	bus.Publish(domain.NewTrackStartedEvent(track))

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Expected 1 call before unsubscribe, got %d", callCount)
	}

	// Unsubscribe
	bus.Unsubscribe(subID)

	// Publish again - handler should NOT be called
	bus.Publish(domain.NewTrackStartedEvent(track))

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Expected 1 call after unsubscribe, got %d", callCount)
	}
}

// TestUnsubscribeInvalidID tests unsubscribing with invalid ID (should be no-op).
func TestUnsubscribeInvalidID(t *testing.T) {
	bus := NewSyncEventBus()
	defer bus.Close()

	// Should not panic
	bus.Unsubscribe("invalid-id")
	bus.Unsubscribe("")
}

// TestSubscribeAll tests wildcard subscriptions.
func TestSubscribeAll(t *testing.T) {
	bus := NewSyncEventBus()
	defer bus.Close()

	var receivedEvents []domain.Event
	var mu sync.Mutex

	handler := func(event domain.Event) {
		mu.Lock()
		defer mu.Unlock()
		receivedEvents = append(receivedEvents, event)
	}

	// Subscribe to all events
	bus.SubscribeAll(handler)

	// Publish different event types
	track := domain.MusicTrack{ID: "test", Title: "Test"}
	bus.Publish(domain.NewTrackStartedEvent(track))
	bus.Publish(domain.NewTrackPausedEvent(track, 10*time.Second))
	bus.Publish(domain.NewVolumeChangedEvent(0.5))

	// Handler should receive all events
	mu.Lock()
	defer mu.Unlock()

	if len(receivedEvents) != 3 {
		t.Errorf("Expected 3 events, got %d", len(receivedEvents))
	}
}

// TestHasSubscribers tests the HasSubscribers method.
func TestHasSubscribers(t *testing.T) {
	bus := NewSyncEventBus()
	defer bus.Close()

	// No subscribers initially
	if bus.HasSubscribers(domain.EventTrackStarted) {
		t.Error("Expected no subscribers initially")
	}

	// Subscribe
	handler := func(event domain.Event) {}
	bus.Subscribe(domain.EventTrackStarted, handler)

	// Should have subscribers now
	if !bus.HasSubscribers(domain.EventTrackStarted) {
		t.Error("Expected subscribers after subscription")
	}

	// Other event types should still have no subscribers
	if bus.HasSubscribers(domain.EventTrackPaused) {
		t.Error("Expected no subscribers for different event type")
	}
}

// TestHasSubscribersWithWildcard tests HasSubscribers with wildcard subscriptions.
func TestHasSubscribersWithWildcard(t *testing.T) {
	bus := NewSyncEventBus()
	defer bus.Close()

	// Subscribe to all events
	handler := func(event domain.Event) {}
	bus.SubscribeAll(handler)

	// All event types should report having subscribers
	if !bus.HasSubscribers(domain.EventTrackStarted) {
		t.Error("Expected subscribers (wildcard) for EventTrackStarted")
	}

	if !bus.HasSubscribers(domain.EventTrackPaused) {
		t.Error("Expected subscribers (wildcard) for EventTrackPaused")
	}
}

// TestHandlerPanic tests that panicking handlers don't crash the bus.
func TestHandlerPanic(t *testing.T) {
	bus := NewSyncEventBus()
	defer bus.Close()

	var callCount int32

	panicHandler := func(event domain.Event) {
		panic("test panic")
	}

	normalHandler := func(event domain.Event) {
		atomic.AddInt32(&callCount, 1)
	}

	// Subscribe both handlers
	bus.Subscribe(domain.EventTrackStarted, panicHandler)
	bus.Subscribe(domain.EventTrackStarted, normalHandler)

	// Publish event - should not crash, normal handler should still be called
	track := domain.MusicTrack{ID: "test", Title: "Test"}
	bus.Publish(domain.NewTrackStartedEvent(track))

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Expected normal handler to be called despite panic, got %d calls", callCount)
	}
}

// TestClose tests closing the event bus.
func TestClose(t *testing.T) {
	bus := NewSyncEventBus()

	// Add some subscribers
	handler := func(event domain.Event) {}
	bus.Subscribe(domain.EventTrackStarted, handler)
	bus.SubscribeAll(handler)

	if bus.SubscriberCount() == 0 {
		t.Error("Expected subscribers before close")
	}

	// Close the bus
	err := bus.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}

	// All subscribers should be cleared
	if bus.SubscriberCount() != 0 {
		t.Errorf("Expected 0 subscribers after close, got %d", bus.SubscriberCount())
	}

	// Publishing should be a no-op (shouldn't panic)
	track := domain.MusicTrack{ID: "test", Title: "Test"}
	bus.Publish(domain.NewTrackStartedEvent(track))

	// Closing again should return error
	err = bus.Close()
	if err == nil {
		t.Error("Expected error when closing already closed bus")
	}
}

// TestConcurrentPublish tests concurrent event publishing (race condition test).
func TestConcurrentPublish(t *testing.T) {
	bus := NewSyncEventBus()
	defer bus.Close()

	var eventCount int32

	handler := func(event domain.Event) {
		atomic.AddInt32(&eventCount, 1)
	}

	bus.Subscribe(domain.EventTrackStarted, handler)

	// Publish events from multiple goroutines
	const numGoroutines = 10
	const eventsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	track := domain.MusicTrack{ID: "test", Title: "Test"}

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				bus.Publish(domain.NewTrackStartedEvent(track))
			}
		}()
	}

	wg.Wait()

	expectedCount := int32(numGoroutines * eventsPerGoroutine)
	if atomic.LoadInt32(&eventCount) != expectedCount {
		t.Errorf("Expected %d events, got %d", expectedCount, eventCount)
	}
}

// TestConcurrentSubscribe tests concurrent subscriptions (race condition test).
func TestConcurrentSubscribe(t *testing.T) {
	bus := NewSyncEventBus()
	defer bus.Close()

	const numGoroutines = 10
	const subscriptionsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	handler := func(event domain.Event) {}

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < subscriptionsPerGoroutine; j++ {
				bus.Subscribe(domain.EventTrackStarted, handler)
			}
		}()
	}

	wg.Wait()

	expectedCount := numGoroutines * subscriptionsPerGoroutine
	if bus.SubscriberCount() != expectedCount {
		t.Errorf("Expected %d subscribers, got %d", expectedCount, bus.SubscriberCount())
	}
}

// TestConcurrentPublishAndSubscribe tests concurrent publishing and subscribing.
func TestConcurrentPublishAndSubscribe(t *testing.T) {
	bus := NewSyncEventBus()
	defer bus.Close()

	var eventCount int32

	handler := func(event domain.Event) {
		atomic.AddInt32(&eventCount, 1)
	}

	const numPublishers = 5
	const numSubscribers = 5
	const eventsPerPublisher = 50

	var wg sync.WaitGroup
	wg.Add(numPublishers + numSubscribers)

	track := domain.MusicTrack{ID: "test", Title: "Test"}

	// Start publishers
	for i := 0; i < numPublishers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < eventsPerPublisher; j++ {
				bus.Publish(domain.NewTrackStartedEvent(track))
				time.Sleep(time.Microsecond) // Small delay to allow interleaving
			}
		}()
	}

	// Start subscribers
	for i := 0; i < numSubscribers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				bus.Subscribe(domain.EventTrackStarted, handler)
				time.Sleep(time.Microsecond)
			}
		}()
	}

	wg.Wait()

	// Should have received events without crashing
	if atomic.LoadInt32(&eventCount) == 0 {
		t.Error("Expected to receive some events")
	}
}

// TestNilEvent tests publishing nil event (should be no-op).
func TestNilEvent(t *testing.T) {
	bus := NewSyncEventBus()
	defer bus.Close()

	var callCount int32

	handler := func(event domain.Event) {
		atomic.AddInt32(&callCount, 1)
	}

	bus.Subscribe(domain.EventTrackStarted, handler)

	// Publishing nil should be a no-op
	bus.Publish(nil)

	if atomic.LoadInt32(&callCount) != 0 {
		t.Errorf("Handler should not be called for nil event, got %d calls", callCount)
	}
}

// TestNilHandler tests that subscribing with nil handler panics.
func TestNilHandler(t *testing.T) {
	bus := NewSyncEventBus()
	defer bus.Close()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when subscribing with nil handler")
		}
	}()

	bus.Subscribe(domain.EventTrackStarted, nil)
}

// TestDifferentEventTypes tests that subscribers only receive their event type.
func TestDifferentEventTypes(t *testing.T) {
	bus := NewSyncEventBus()
	defer bus.Close()

	var startedCount, pausedCount int32

	startedHandler := func(event domain.Event) {
		atomic.AddInt32(&startedCount, 1)
	}

	pausedHandler := func(event domain.Event) {
		atomic.AddInt32(&pausedCount, 1)
	}

	bus.Subscribe(domain.EventTrackStarted, startedHandler)
	bus.Subscribe(domain.EventTrackPaused, pausedHandler)

	track := domain.MusicTrack{ID: "test", Title: "Test"}

	// Publish started event
	bus.Publish(domain.NewTrackStartedEvent(track))

	if atomic.LoadInt32(&startedCount) != 1 {
		t.Errorf("Expected 1 started event, got %d", startedCount)
	}
	if atomic.LoadInt32(&pausedCount) != 0 {
		t.Errorf("Expected 0 paused events, got %d", pausedCount)
	}

	// Publish paused event
	bus.Publish(domain.NewTrackPausedEvent(track, 5*time.Second))

	if atomic.LoadInt32(&startedCount) != 1 {
		t.Errorf("Expected 1 started event after pause, got %d", startedCount)
	}
	if atomic.LoadInt32(&pausedCount) != 1 {
		t.Errorf("Expected 1 paused event, got %d", pausedCount)
	}
}
