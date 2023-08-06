package dwidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"github.com/dReam-dApps/dReams/bundle"
)

// Create a new *widget.Label with center alignment
func NewCenterLabel(text string) *widget.Label {
	center := widget.NewLabel(text)
	center.Alignment = fyne.TextAlignCenter

	return center
}

// Create a new *widget.Label with trailing alignment
func NewTrailingLabel(text string) *widget.Label {
	trailing := widget.NewLabel(text)
	trailing.Alignment = fyne.TextAlignTrailing

	return trailing
}

// Create a new *canvas.Text with size and alignment
func NewCanvasText(text string, size float32, align fyne.TextAlign) (canv *canvas.Text) {
	canv = canvas.NewText(text, bundle.TextColor)
	canv.TextSize = size
	canv.Alignment = align

	return
}
