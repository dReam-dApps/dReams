package menu

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var Username string

// Dero wallet name entry
func NameEntry() fyne.CanvasObject {
	Control.Names = widget.NewSelect([]string{}, func(s string) {
		Username = s
	})

	Control.Names.PlaceHolder = "Wallet names:"

	return container.NewHBox(layout.NewSpacer(), Control.Names)
}
