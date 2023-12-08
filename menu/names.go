package menu

import (
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dReam-dApps/dReams/rpc"
)

var Username string

// Dero wallet name entry
func NameEntry() fyne.CanvasObject {
	Assets.Names = widget.NewSelect([]string{}, func(s string) {
		Username = s
	})

	Assets.Names.PlaceHolder = "Wallet names:"

	return container.NewStack(Assets.Names)
}

// Get a wallets registered names
func CheckWalletNames(value string) {
	if gnomon.IsReady() {
		names, _ := gnomon.GetSCIDKeysByValue(rpc.NameSCID, value)

		sort.Strings(names)
		Assets.Names.Options = append(Assets.Names.Options, names...)
	}
}
