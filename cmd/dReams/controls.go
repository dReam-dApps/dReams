package main

import (
	"fyne.io/fyne/v2/canvas"
	dreams "github.com/SixofClubsss/dReams"
	"github.com/SixofClubsss/dReams/holdero"
	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/prediction"
	"github.com/SixofClubsss/dReams/rpc"
)

// dReams search filters for Gnomon index
func GnomonFilters() (filter []string) {
	if menu.Control.Dapp_list["Holdero"] {
		holdero110 := rpc.GetSCCode(holdero.HolderoSCID)
		if holdero110 != "" {
			filter = append(filter, holdero110)
		}

		holdero100 := rpc.GetSCCode(holdero.Holdero100)
		if holdero100 != "" {
			filter = append(filter, holdero100)
		}

		holderoHGC := rpc.GetSCCode(holdero.HGCHolderoSCID)
		if holderoHGC != "" {
			filter = append(filter, holderoHGC)
		}
	}

	if menu.Control.Dapp_list["Baccarat"] {
		bacc := rpc.GetSCCode(rpc.BaccSCID)
		if bacc != "" {
			filter = append(filter, bacc)
		}
	}

	if menu.Control.Dapp_list["dSports and dPredictions"] {
		predict := rpc.GetSCCode(prediction.PredictSCID)
		if predict != "" {
			filter = append(filter, predict)
		}

		sports := rpc.GetSCCode(prediction.SportsSCID)
		if sports != "" {
			filter = append(filter, sports)
		}
	}

	gnomon := rpc.GetGnomonCode()
	if gnomon != "" {
		filter = append(filter, gnomon)
	}

	names := rpc.GetNameServiceCode()
	if names != "" {
		filter = append(filter, names)
	}

	ratings := rpc.GetSCCode(rpc.RatingSCID)
	if ratings != "" {
		filter = append(filter, ratings)
	}

	if menu.Control.Dapp_list["DerBnb"] {
		bnb := rpc.GetSCCode(rpc.DerBnbSCID)
		if bnb != "" {
			filter = append(filter, bnb)
		}
	}

	filter = append(filter, menu.NFA_SEARCH_FILTER)
	if !menu.Gnomes.Trim {
		filter = append(filter, menu.G45_search_filter)
	}

	return
}

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
		clearContractLists()
		disableActions(true)
		disconnected()
		menu.Gnomes.Checked(false)
	}
}

// Do when disconnected
func disconnected() {
	holdero.Disconnected(menu.Control.Dapp_list["Holdero"])
	prediction.Disconnected()
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
