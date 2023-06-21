package main

import (
	"fyne.io/fyne/v2/canvas"
	dreams "github.com/SixofClubsss/dReams"
	"github.com/SixofClubsss/dReams/baccarat"
	"github.com/SixofClubsss/dReams/holdero"
	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/tarot"
)

// Connection check for main process
func CheckConnection() {
	if rpc.Daemon.IsConnected() {
		menu.Control.Daemon_check.SetChecked(true)
		menu.DisableIndexControls(false)
	} else {
		menu.Control.Daemon_check.SetChecked(false)
		if menu.Control.Dapp_list["dSports and dPredictions"] {
			menu.Control.Predict_check.SetChecked(false)
			menu.Control.Sports_check.SetChecked(false)
		}

		if menu.Control.Dapp_list["Baccarat"] {
			disableBaccActions(true)
		}

		disableActions(true)
		menu.DisableIndexControls(true)
	}

	if rpc.Wallet.IsConnected() {
		disableActions(false)
	} else {
		holdero.Signal.Contract = false
		clearContractLists()
		if menu.Control.Dapp_list["Holdero"] {
			holdero.Settings.Check.SetChecked(false)
			holdero.DisableOwnerControls(true)
		}

		if menu.Control.Dapp_list["Baccarat"] {
			disableBaccActions(true)
		}

		disableActions(true)
		disconnected()
		menu.Gnomes.Checked(false)
	}

	if menu.Control.Dapp_list["Holdero"] {
		if holdero.Signal.Contract {
			holdero.Settings.Check.SetChecked(true)
		} else {
			holdero.Settings.Check.SetChecked(false)
			holdero.DisableOwnerControls(true)
			holdero.Signal.Sit = true
		}
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
	/// put back
	//holdero.Settings.Tables = []string{}
	//Control.Predict_contracts = []string{}
	//Control.Sports_contracts = []string{}
	//holdero.Settings.Owned = []string{}
	//Control.Predict_owned = []string{}
	//Control.Sports_owned = []string{}
	menu.Market.Auctions = []string{}
	menu.Market.Buy_now = []string{}
	menu.Assets.Assets = []string{}
}

// Disable actions requiring connection
func disableActions(d bool) {
	if d {
		menu.Assets.Swap.Hide()

		if menu.Control.Dapp_list["dSports and dPredictions"] {
			menu.Control.Bet_new_p.Hide()
			menu.Control.Bet_new_s.Hide()
			menu.Control.Bet_unlock_p.Hide()
			menu.Control.Bet_unlock_s.Hide()
			menu.Control.Bet_menu_p.Hide()
			menu.Control.Bet_menu_s.Hide()
			menu.Control.Bet_new_p.Refresh()
			menu.Control.Bet_new_s.Refresh()
			menu.Control.Bet_unlock_p.Refresh()
			menu.Control.Bet_unlock_s.Refresh()
			menu.Control.Bet_menu_p.Refresh()
			menu.Control.Bet_menu_s.Refresh()
		}

		if menu.Control.Dapp_list["Iluma"] {
			tarot.Iluma.Draw1.Hide()
			tarot.Iluma.Draw3.Hide()
			tarot.Iluma.Search.Hide()
			tarot.Iluma.Draw1.Refresh()
			tarot.Iluma.Draw3.Refresh()
			tarot.Iluma.Search.Refresh()
		}
	} else {
		if rpc.Daemon.IsConnected() {
			menu.Assets.Swap.Show()
		}
	}

	menu.Assets.Swap.Refresh()
}

// Disable Baccarat actions
func disableBaccActions(d bool) {
	if d {
		baccarat.Table.Actions.Hide()
	} else {
		baccarat.Table.Actions.Show()
	}

	baccarat.Table.Actions.Refresh()
}
