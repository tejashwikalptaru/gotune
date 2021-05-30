package extended

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type DoubleTapLabel struct {
	widget.Label
	doubleTapped func(index int)
	index int
}

func NewDoubleTapLabel(f func(int)) *DoubleTapLabel {
	label := &DoubleTapLabel{doubleTapped: f}
	label.ExtendBaseWidget(label)
	return label
}

func (label *DoubleTapLabel) DoubleTapped(_ *fyne.PointEvent) {
	if label.doubleTapped != nil {
		label.doubleTapped(label.index)
	}
}

func (label *DoubleTapLabel) Set(text string, index int) {
	label.SetText(text)
	label.index = index
}
