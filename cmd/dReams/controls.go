package main

import (
	"fyne.io/fyne/v2/canvas"
	dreams "github.com/SixofClubsss/dReams"
	"github.com/SixofClubsss/dReams/holdero"
	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"
)

// Connection check for main process
func CheckConnection() {
	if rpc.Daemon.IsConnected() {
		menu.Control.Daemon_check.SetChecked(true)
		menu.DisableIndexControls(false)
	} else {
		menu.Control.Daemon_check.SetChecked(false)

		disableActions(true)
		menu.DisableIndexControls(true)
	}

	if rpc.Wallet.IsConnected() {
		disableActions(false)
	} else {
		holdero.Signal.Contract = false
		clearContractLists()

		disableActions(true)
		disconnected()
		menu.Gnomes.Checked(false)
	}
}

// Do when disconnected
func disconnected() {
	holdero.Disconnected(menu.Control.Dapp_list["Holdero"])
	rpc.Wallet.Service = false
	rpc.Wallet.PokerOwner = false
	rpc.Wallet.BetOwner = false
	rpc.Wallet.Address = ""
	dreams.Theme.Select.Options = []string{"Main", "Legacy"}
	dreams.Theme.Select.Refresh()
	menu.Assets.Assets = []string{}
	menu.Assets.Name.Text = (" Name:")
	menu.Assets.Name.Refresh()
	menu.Assets.Collection.Text = (" Collection:")
	menu.Assets.Collection.Refresh()
	menu.Assets.Icon = *canvas.NewImageFromImage(nil)
	menu.Market.Auction_list.UnselectAll()
	menu.Market.Buy_list.UnselectAll()
	menu.Market.Icon = *canvas.NewImageFromImage(nil)
	menu.Market.Cover = *canvas.NewImageFromImage(nil)
	menu.Market.Viewing = ""
	menu.Market.Viewing_coll = ""
	menu.ResetAuctionInfo()
	menu.AuctionInfo()
}

// Clear all contract lists
func clearContractLists() {
	menu.Market.Auctions = []string{}
	menu.Market.Buy_now = []string{}
	menu.Assets.Assets = []string{}
}

// Disable actions requiring connection
func disableActions(d bool) {
	if d {
		menu.Assets.Swap.Hide()
	} else {
		if rpc.Daemon.IsConnected() {
			menu.Assets.Swap.Show()
		}
	}

	menu.Assets.Swap.Refresh()
}
