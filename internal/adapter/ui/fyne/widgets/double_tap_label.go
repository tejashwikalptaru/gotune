package widgets

import (
	fyneapp "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// DoubleTapLabel is a custom label widget that responds to double-tap gestures.
// It extends the standard Fyne Label widget to support double-tap interaction,
// which is used in the playlist view for track selection.
type DoubleTapLabel struct {
	widget.Label
	doubleTapped func(index int)
	index        int
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
