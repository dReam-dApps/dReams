package dwidget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dReam-dApps/dReams/bundle"
)

// Top label background used on dApp tabs
func LabelColor(c *fyne.Container) *fyne.Container {
	var alpha *canvas.Rectangle
	if bundle.AppColor == color.White {
		alpha = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x33})
	} else {
		alpha = canvas.NewRectangle(color.RGBA{0, 0, 0, 150})
	}

	return container.NewStack(alpha, c)
}

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
