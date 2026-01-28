// Package widgets provides custom Fyne widgets for the GoTune application.
package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// TappableStack is a container that wraps content and responds to secondary (right-click) taps.
// It's used to provide a context menu trigger for the album art / visualizer area.
type TappableStack struct {
	widget.BaseWidget

	content        fyne.CanvasObject
	onSecondaryTap func(*fyne.PointEvent)
}

// NewTappableStack creates a new tappable stack with the given content.
func NewTappableStack(content fyne.CanvasObject, onSecondaryTap func(*fyne.PointEvent)) *TappableStack {
	t := &TappableStack{
		content:        content,
		onSecondaryTap: onSecondaryTap,
	}
	t.ExtendBaseWidget(t)
	return t
}

// CreateRenderer implements fyne.Widget.
func (t *TappableStack) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.content)
}

// Tapped implements fyne.Tappable (primary tap - left click).
// We don't do anything on primary tap.
func (t *TappableStack) Tapped(*fyne.PointEvent) {
	// No action on the primary tap
}

// TappedSecondary implements fyne.SecondaryTappable (right-click).
func (t *TappableStack) TappedSecondary(pe *fyne.PointEvent) {
	if t.onSecondaryTap != nil {
		t.onSecondaryTap(pe)
	}
}

// MouseIn implements desktop.Hoverable.
func (t *TappableStack) MouseIn(*desktop.MouseEvent) {}

// MouseMoved implements desktop.Hoverable.
func (t *TappableStack) MouseMoved(*desktop.MouseEvent) {}

// MouseOut implements desktop.Hoverable.
func (t *TappableStack) MouseOut() {}

// SetContent updates the content of the stack.
func (t *TappableStack) SetContent(content fyne.CanvasObject) {
	t.content = content
	t.Refresh()
}

// Ensure TappableStack implements the required interfaces
var _ fyne.Tappable = (*TappableStack)(nil)
var _ fyne.SecondaryTappable = (*TappableStack)(nil)
var _ desktop.Hoverable = (*TappableStack)(nil)
