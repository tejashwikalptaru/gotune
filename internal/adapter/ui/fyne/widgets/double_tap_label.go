package widgets

import (
	fyneapp "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// Ensure DoubleTapLabel implements SecondaryTappable interface
var _ fyneapp.SecondaryTappable = (*DoubleTapLabel)(nil)

// DoubleTapLabel is a custom label widget that responds to double-tap gestures.
// It extends the standard Fyne Label widget to support double-tap interaction,
// which is used in the playlist view for track selection.
type DoubleTapLabel struct {
	widget.Label
	doubleTapped     func(index int)
	secondaryTapped  func(index int, pos fyneapp.Position)
	index            int
}

// NewDoubleTapLabel creates a new DoubleTapLabel with the given callback function.
// The callback is invoked when the label is double-tapped, passing the item index.
func NewDoubleTapLabel(doubleTapped func(index int)) *DoubleTapLabel {
	label := &DoubleTapLabel{
		doubleTapped: doubleTapped,
	}
	label.ExtendBaseWidget(label)
	return label
}

// DoubleTapped implements the fyne.DoubleTappable interface.
// It is called when the user double-taps the label.
func (l *DoubleTapLabel) DoubleTapped(_ *fyneapp.PointEvent) {
	if l.doubleTapped != nil {
		l.doubleTapped(l.index)
	}
}

// SetIndex sets the index associated with this label.
// This is typically the position of the item in a list.
func (l *DoubleTapLabel) SetIndex(index int) {
	l.index = index
}

// SetText sets the text content of the label.
func (l *DoubleTapLabel) SetText(text string) {
	l.Label.SetText(text)
}

// SetSecondaryTapped sets the callback function for right-click (secondary tap) events.
func (l *DoubleTapLabel) SetSecondaryTapped(callback func(index int, pos fyneapp.Position)) {
	l.secondaryTapped = callback
}

// TappedSecondary implements the fyne.SecondaryTappable interface.
// It is called when the user right-clicks (or secondary taps) the label.
func (l *DoubleTapLabel) TappedSecondary(pe *fyneapp.PointEvent) {
	if l.secondaryTapped != nil {
		l.secondaryTapped(l.index, pe.AbsolutePosition)
	}
}
