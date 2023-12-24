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
	var wallet, names []string
	if len(value) > 12 {
		wallet = append(wallet, value[0:12])
	}

	if gnomon.IsReady() {
		names, _ = gnomon.GetSCIDKeysByValue(rpc.NameSCID, value)

		sort.Strings(names)
	}

	Assets.Names.Options = append(wallet, names...)

	if len(Assets.Names.Options) > 0 {
		if Username == "" {
			Assets.Names.SetSelectedIndex(0)
		} else {
			var found bool
			for _, name := range Assets.Names.Options {
				if name == Username {
					found = true
					Assets.Names.SetSelected(name)
					break
				}
			}

			if !found {
				Assets.Names.SetSelectedIndex(0)
			}
		}
	}
}
