package dwidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// ContainerStack used for building various
// container/label layouts to be placed in a main app
type ContainerStack struct {
	Left     stackLabel
	Right    stackLabel
	TopLabel *canvas.Text

	Back    fyne.Container
	Front   fyne.Container
	Actions fyne.Container
	DApp    *fyne.Container
}

type stackLabel struct {
	Label  *widget.Label
	update func() string
}

// Updates label text with internal update()
func (l *stackLabel) UpdateText() {
	l.Label.SetText(l.update())
}

// Used to set a standard update func for label text
func (l *stackLabel) SetUpdate(updateText func() string) {
	l.update = updateText
	l.Label.SetText(l.update())
}
