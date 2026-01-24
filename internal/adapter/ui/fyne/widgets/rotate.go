package widgets

import "fmt"

type Rotator struct {
	text string
	len  int
}

func NewRotator(text string, maxLength int) *Rotator {
	return &Rotator{text: fmt.Sprintf("    %s", text), len: maxLength}
}

func (r *Rotator) Rotate() string {
	if len(r.text) < r.len {
		// do not rotate
		return r.text
	}
	runes := []rune(r.text)
	var newRune []rune
	var temp int32
	for i, v := range runes {
		if i == 0 {
			temp = v
			continue
		}
		newRune = append(newRune, v)
	}

	newRune = append(newRune, temp)
	r.text = string(newRune)
	return r.text
}
