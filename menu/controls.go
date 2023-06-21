package menu

import (
	"encoding/json"
	"image/color"
	"log"
	"os"

	dreams "github.com/SixofClubsss/dReams"
	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/rpc"
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
	/// put back
	///u.Tables = Control.Holdero_favorites
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
	if !dreams.FileExists("config/config.json", tag) {
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
			if Control.List_open {
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
