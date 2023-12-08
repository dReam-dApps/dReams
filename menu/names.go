package menu

import (
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/dReam-dApps/dReams/rpc"
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

// Get a wallets registered names
func CheckWalletNames(value string) {
	if Gnomes.IsReady() {
		names, _ := Gnomes.GetSCIDKeysByValue(rpc.NameSCID, value)

		sort.Strings(names)
		Control.Names.Options = append(Control.Names.Options, names...)
	}
}
