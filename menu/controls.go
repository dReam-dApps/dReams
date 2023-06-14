package menu

import (
	"encoding/json"
	"image/color"
	"log"
	"os"

	"fyne.io/fyne/v2/canvas"
	"github.com/SixofClubsss/dReams/baccarat"
	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/holdero"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/tarot"
)

type dReamSave struct {
	Skin    color.Gray16 `json:"skin"`
	Daemon  []string     `json:"daemon"`
	Tables  []string     `json:"tables"`
	Predict []string     `json:"predict"`
	Sports  []string     `json:"sports"`

	Dapps map[string]bool `json:"dapps"`
}

// Save dReams config.json file for platform wide dApp use
func WriteDreamsConfig(daemon string, skin color.Gray16) {
	var u dReamSave
	switch daemon {
	case rpc.DAEMON_RPC_DEFAULT:
	case rpc.DAEMON_RPC_REMOTE1:
	case rpc.DAEMON_RPC_REMOTE2:
	// case menu.DAEMON_RPC_REMOTE3:
	// case menu.DAEMON_RPC_REMOTE4:
	case rpc.DAEMON_RPC_REMOTE5:
	case rpc.DAEMON_RPC_REMOTE6:
	default:
		u.Daemon = []string{daemon}
	}

	u.Skin = skin
	u.Tables = Control.Holdero_favorites
	u.Predict = Control.Predict_favorites
	u.Sports = Control.Sports_favorites
	u.Dapps = Control.Dapp_list

	if u.Daemon != nil {
		if u.Daemon[0] == "" {
			if Control.Daemon_config != "" {
				u.Daemon[0] = Control.Daemon_config
			} else {
				u.Daemon[0] = "127.0.0.1:10102"
			}
		}

		file, err := os.Create("config/config.json")
		if err != nil {
			log.Println("[WriteDreamsConfig]", err)
			return
		}

		defer file.Close()
		json, _ := json.MarshalIndent(u, "", " ")

		if _, err = file.Write(json); err != nil {
			log.Println("[WriteDreamsConfig]", err)
		}
	}
}

// Read dReams platform config.json file
//   - tag for log print
//   - Sets up directory if none exists
func ReadDreamsConfig(tag string) (saved dReamSave) {
	if !holdero.FileExists("config/config.json", tag) {
		log.Printf("[%s] Creating config directory\n", tag)
		mkdir := os.Mkdir("config", 0755)
		if mkdir != nil {
			log.Printf("[%s] %s\n", tag, mkdir)
		}

		if config, err := os.Create("config/config.json"); err == nil {
			var save dReamSave
			json, _ := json.MarshalIndent(&save, "", " ")
			if _, err = config.Write(json); err != nil {
				log.Println("[WriteDreamsConfig]", err)
			}
			config.Close()
		}

		return
	}

	file, err := os.ReadFile("config/config.json")
	if err != nil {
		log.Println("[ReadDreamsConfig]", err)
		return
	}

	if err = json.Unmarshal(file, &saved); err != nil {
		log.Println("[ReadDreamsConfig]", err)
		return
	}

	bundle.AppColor = saved.Skin
	Control.Dapp_list = make(map[string]bool)
	Control.Dapp_list = saved.Dapps

	return
}

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
	holdero.Settings.ThemeSelect.Options = []string{"Main", "Legacy"}
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
	Assets.Assets = []string{}
	Assets.Name.Text = (" Name:")
	Assets.Name.Refresh()
	Assets.Collection.Text = (" Collection:")
	Assets.Collection.Refresh()
	Assets.Icon = *canvas.NewImageFromImage(nil)
	if Control.Dapp_list["Holdero"] {
		holdero.DisableHolderoTools()
		Control.Names.ClearSelected()
		Control.Names.Options = []string{}
		Control.Names.Refresh()
	}
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
	Assets.Assets = []string{}
}

// Disable index objects
func DisableIndexControls(d bool) {
	if d {
		Assets.Index_button.Hide()
		Assets.Index_search.Hide()
		Assets.Header_box.Hide()
		Market.Market_box.Hide()
		Gnomes.SCIDS = 0
	} else {
		Assets.Index_button.Show()
		Assets.Index_search.Show()
		if rpc.Wallet.IsConnected() {
			Control.Claim_button.Show()
			Assets.Header_box.Show()
			Market.Market_box.Show()
			if Control.list_open {
				Control.List_button.Hide()
			}
		} else {
			Control.Send_asset.Hide()
			Control.List_button.Hide()
			Control.Claim_button.Hide()
			Assets.Header_box.Hide()
			Market.Market_box.Hide()
		}
	}
	Assets.Index_button.Refresh()
	Assets.Index_search.Refresh()
	Assets.Header_box.Refresh()
	Market.Market_box.Refresh()
}

// Disable actions requiring connection
func disableActions(d bool) {
	if d {
		Assets.Swap.Hide()
		if Control.Dapp_list["Holdero"] {
			Poker.Holdero_unlock.Hide()
			Poker.Holdero_new.Hide()
			holdero.Table.Tournament.Hide()
			Poker.Holdero_unlock.Refresh()
			Poker.Holdero_new.Refresh()
			holdero.Table.Tournament.Refresh()
		}

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
		if rpc.Daemon.Connect {
			Assets.Swap.Show()
		}
	}

	Assets.Swap.Refresh()
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
