package bundle

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// Left arrow icon *canvas.Image
func LeftArrow(size fyne.Size) fyne.CanvasObject {
	leftArrow := canvas.NewImageFromResource(ResourceLeftArrowPng)
	leftArrow.SetMinSize(size)

	return container.NewStack(leftArrow)
}

// Right arrow icon *canvas.Image
func RightArrow(size fyne.Size) fyne.CanvasObject {
	rightArrow := canvas.NewImageFromResource(ResourceRightArrowPng)
	rightArrow.SetMinSize(size)

	return container.NewStack(rightArrow)
}
