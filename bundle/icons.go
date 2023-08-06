package bundle

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// Left arrow icon *canvas.Image
func LeftArrow(size fyne.Size) *canvas.Image {
	leftArrow := canvas.NewImageFromResource(ResourceLeftArrowPng)
	leftArrow.SetMinSize(size)

	return leftArrow
}

// Right arrow icon *canvas.Image
func RightArrow(size fyne.Size) *canvas.Image {
	rightArrow := canvas.NewImageFromResource(ResourceRightArrowPng)
	rightArrow.SetMinSize(size)

	return rightArrow
}
