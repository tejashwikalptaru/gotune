// Package testutil provides testing utilities for the GoTune application.
package testutil

import (
	"testing"

	"go.uber.org/goleak"
)

// VerifyNoLeaks should be deferred at the start of tests that spawn goroutines.
// It verifies that no goroutines were leaked during the test.
func VerifyNoLeaks(t *testing.T, opts ...goleak.Option) {
	t.Helper()
	goleak.VerifyNone(t, opts...)
}

// IgnoreFyneGoroutines returns goleak options to ignore known Fyne framework goroutines.
// Use this when testing components that use Fyne.
func IgnoreFyneGoroutines() []goleak.Option {
	return []goleak.Option{
		goleak.IgnoreTopFunction("fyne.io/fyne/v2/internal/driver/glfw.(*gLDriver).runGL.func1"),
		goleak.IgnoreTopFunction("fyne.io/fyne/v2/internal/driver/glfw.(*window).RunEventQueue"),
		goleak.IgnoreTopFunction("fyne.io/fyne/v2/internal/animation.(*Runner).runAnimations"),
		goleak.IgnoreAnyFunction("fyne.io/fyne/v2"),
	}
}
