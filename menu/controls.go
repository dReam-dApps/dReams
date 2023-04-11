package menu

import (
	"image/color"

	"fyne.io/fyne/v2/canvas"
	"github.com/SixofClubsss/dReams/baccarat"
	"github.com/SixofClubsss/dReams/holdero"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/tarot"
)

var (
	Alpha120 = canvas.NewRectangle(color.RGBA{0, 0, 0, 120})
	Alpha150 = canvas.NewRectangle(color.RGBA{0, 0, 0, 150})
	Alpha180 = canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
)

// Do when disconnected
func Disconnected() {
	rpc.Wallet.Service = false
	rpc.Wallet.PokerOwner = false
	rpc.Wallet.BetOwner = false
	rpc.Wallet.Address = ""
	rpc.Round.ID = 0
	rpc.Display.PlayerId = ""
	rpc.Odds.Run = false
	holdero.Settings.FaceSelect.Options = []string{"Light", "Dark"}
	holdero.Settings.BackSelect.Options = []string{"Light", "Dark"}
	holdero.Settings.ThemeSelect.Options = []string{"Main"}
	holdero.Settings.AvatarSelect.Options = []string{"None"}
	holdero.Settings.FaceUrl = ""
	holdero.Settings.BackUrl = ""
	holdero.Settings.AvatarUrl = ""
	holdero.Settings.FaceSelect.SetSelectedIndex(0)
	holdero.Settings.BackSelect.SetSelectedIndex(0)
	holdero.Settings.AvatarSelect.SetSelectedIndex(0)
	holdero.Settings.FaceSelect.Refresh()
	holdero.Settings.BackSelect.Refresh()
	holdero.Settings.ThemeSelect.Refresh()
	holdero.Settings.AvatarSelect.Refresh()
	holdero.Assets.Assets = []string{}
	holdero.Assets.Name.Text = (" Name:")
	holdero.Assets.Name.Refresh()
	holdero.Assets.Collection.Text = (" Collection:")
	holdero.Assets.Collection.Refresh()
	holdero.Assets.Icon = *canvas.NewImageFromImage(nil)
	// prediction leaderboard
	// holdero.Table.NameEntry.Text = ""
	// holdero.Table.NameEntry.Enable()
	// holdero.Table.NameEntry.Refresh()
	holdero.DisableHolderoTools()
	Control.Names.ClearSelected()
	Control.Names.Options = []string{}
	Control.Names.Refresh()
	Market.Auction_list.UnselectAll()
	Market.Buy_list.UnselectAll()
	Market.Icon = *canvas.NewImageFromImage(nil)
	Market.Cover = *canvas.NewImageFromImage(nil)
	Market.Viewing = ""
	Market.Viewing_coll = ""
	ResetAuctionInfo()
	AuctionInfo()
}

// Clear all contract lists
func clearContractLists() {
	Control.Holdero_tables = []string{}
	Control.Predict_contracts = []string{}
	Control.Sports_contracts = []string{}
	Control.Holdero_owned = []string{}
	Control.Predict_owned = []string{}
	Control.Sports_owned = []string{}
	Market.Auctions = []string{}
	Market.Buy_now = []string{}
	holdero.Assets.Assets = []string{}
}

// Disable index objects
func disableIndex(d bool) {
	if d {
		holdero.Assets.Index_button.Hide()
		holdero.Assets.Index_search.Hide()
		holdero.Assets.Header_box.Hide()
		Market.Market_box.Hide()
		Gnomes.SCIDS = 0
	} else {
		holdero.Assets.Index_button.Show()
		holdero.Assets.Index_search.Show()
		if rpc.Wallet.Connect {
			Control.Claim_button.Show()
			holdero.Assets.Header_box.Show()
			Market.Market_box.Show()
			if Control.list_open {
				Control.List_button.Hide()
			}
		} else {
			Control.Send_asset.Hide()
			Control.List_button.Hide()
			Control.Claim_button.Hide()
			holdero.Assets.Header_box.Hide()
			Market.Market_box.Hide()
		}
	}
	holdero.Assets.Index_button.Refresh()
	holdero.Assets.Index_search.Refresh()
	holdero.Assets.Header_box.Refresh()
	Market.Market_box.Refresh()
}

// Disable actions requiring connection
func disableActions(d bool) {
	if d {
		holdero.Swap.Dreams.Hide()
		holdero.Swap.Dero.Hide()
		holdero.Swap.DEntry.Hide()
		Poker.Holdero_unlock.Hide()
		Poker.Holdero_new.Hide()
		holdero.Table.Tournament.Hide()

		if Control.Dapp_list["dSports and dPredictions"] {
			Control.Bet_new_p.Hide()
			Control.Bet_new_s.Hide()
			Control.Bet_unlock_p.Hide()
			Control.Bet_unlock_s.Hide()
			Control.Bet_menu_p.Hide()
			Control.Bet_menu_s.Hide()
			Control.Bet_new_p.Refresh()
			Control.Bet_new_s.Refresh()
			Control.Bet_unlock_p.Refresh()
			Control.Bet_unlock_s.Refresh()
			Control.Bet_menu_p.Refresh()
			Control.Bet_menu_s.Refresh()
		}

		if Control.Dapp_list["Iluma"] {
			tarot.Iluma.Draw1.Hide()
			tarot.Iluma.Draw3.Hide()
			tarot.Iluma.Search.Hide()
			tarot.Iluma.Draw1.Refresh()
			tarot.Iluma.Draw3.Refresh()
			tarot.Iluma.Search.Refresh()
		}
	} else {
		holdero.Swap.Dreams.Show()
		holdero.Swap.Dero.Show()
		holdero.Swap.DEntry.Show()
	}

	holdero.Swap.Dreams.Refresh()
	holdero.Swap.DEntry.Refresh()
	holdero.Swap.Dero.Refresh()
	Poker.Holdero_unlock.Refresh()
	Poker.Holdero_new.Refresh()
	holdero.Table.Tournament.Refresh()
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

// Disable Holdero owner actions
func disableOwnerControls(d bool) {
	if d {
		Poker.owner.owners_left.Hide()
		Poker.owner.owners_mid.Hide()
	} else {
		Poker.owner.owners_left.Show()
		Poker.owner.owners_mid.Show()
	}

	Poker.owner.owners_left.Refresh()
	Poker.owner.owners_mid.Refresh()
}

// Set objects if bet owner
func SetBetOwner(owner string) {
	if Control.Dapp_list["dSports and dPredictions"] {
		if owner == rpc.Wallet.Address {
			rpc.Wallet.BetOwner = true
			Control.Bet_new_p.Show()
			Control.Bet_new_s.Show()
			Control.Bet_unlock_p.Hide()
			Control.Bet_unlock_s.Hide()
			Control.Bet_menu_p.Show()
			Control.Bet_menu_s.Show()
		} else {
			rpc.Wallet.BetOwner = false
			Control.Bet_new_p.Hide()
			Control.Bet_new_s.Hide()
			Control.Bet_unlock_p.Show()
			Control.Bet_unlock_s.Show()
			Control.Bet_menu_p.Hide()
			Control.Bet_menu_s.Hide()
		}
	}
}
