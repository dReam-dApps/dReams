package menu

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/dwidget"
	"github.com/SixofClubsss/dReams/holdero"
	"github.com/SixofClubsss/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type menuObjects struct {
	list_open         bool
	send_open         bool
	msg_open          bool
	Daemon_config     string
	Viewing_asset     string
	Dapp_list         map[string]bool
	Holdero_tables    []string
	Holdero_favorites []string
	Holdero_owned     []string
	Predict_contracts []string
	Predict_favorites []string
	Predict_owned     []string
	Sports_contracts  []string
	Sports_favorites  []string
	Sports_owned      []string
	Contract_rating   map[string]uint64
	Names             *widget.Select
	Bet_unlock_p      *widget.Button
	Bet_unlock_s      *widget.Button
	Bet_new_p         *widget.Button
	Bet_new_s         *widget.Button
	Bet_menu_p        *widget.Button
	Bet_menu_s        *widget.Button
	Send_asset        *widget.Button
	Claim_button      *widget.Button
	List_button       *widget.Button
	daemon_check      *widget.Check
	holdero_check     *widget.Check
	Predict_check     *widget.Check
	P_contract        *widget.SelectEntry
	Sports_check      *widget.Check
	S_contract        *widget.SelectEntry
	Wallet_ind        *fyne.Animation
	Daemon_ind        *fyne.Animation
	Poker_ind         *fyne.Animation
	Service_ind       *fyne.Animation
}

type ownerObjects struct {
	blind_amount uint64
	ante_amount  uint64
	chips        *widget.RadioGroup
	timeout      *widget.Button
	owners_left  *fyne.Container
	owners_mid   *fyne.Container
}

type holderoObjects struct {
	contract_input *widget.SelectEntry
	Table_list     *widget.List
	Favorite_list  *widget.List
	Owned_list     *widget.List
	Holdero_unlock *widget.Button
	Holdero_new    *widget.Button
	Stats_box      fyne.Container
	owner          ownerObjects
}

var Poker holderoObjects
var Control menuObjects

// Connection check for main process
func CheckConnection() {
	if rpc.Daemon.Connect {
		Control.daemon_check.SetChecked(true)
		DisableIndexControls(false)
	} else {
		Control.daemon_check.SetChecked(false)
		if Control.Dapp_list["dSports and dPredictions"] {
			Control.Predict_check.SetChecked(false)
			Control.Sports_check.SetChecked(false)
		}

		rpc.Signal.Contract = false
		clearContractLists()
		if Control.Dapp_list["Holdero"] {
			Control.holdero_check.SetChecked(false)
			disableOwnerControls(true)
		}

		if Control.Dapp_list["Baccarat"] {
			disableBaccActions(true)
		}

		disableActions(true)
		DisableIndexControls(true)
		Gnomes.Init = false
		Gnomes.Checked = false
	}

	if rpc.Wallet.Connect {
		disableActions(false)
	} else {
		rpc.Signal.Contract = false
		clearContractLists()
		if Control.Dapp_list["Holdero"] {
			Control.holdero_check.SetChecked(false)
			disableOwnerControls(true)
		}

		if Control.Dapp_list["Baccarat"] {
			disableBaccActions(true)
		}

		disableActions(true)
		Disconnected()
		Gnomes.Checked = false
	}

	if Control.Dapp_list["Holdero"] {
		if rpc.Signal.Contract {
			Control.holdero_check.SetChecked(true)
		} else {
			Control.holdero_check.SetChecked(false)
			disableOwnerControls(true)
			rpc.Signal.Sit = true
		}
	}
}

// Hiden object, controls Gnomon start and stop based on daemon connection
func DaemonConnectedBox() fyne.Widget {
	Control.daemon_check = widget.NewCheck("", func(b bool) {
		if !Gnomes.Init && !Gnomes.Start {
			go startLabel()
			filters := searchFilters()
			StartGnomon("dReams", filters, 3960, 490, g45Index)
			rpc.FetchFees()
			if Control.Dapp_list["Holdero"] {
				Poker.contract_input.CursorColumn = 1
				Poker.contract_input.Refresh()
			}

			if Control.Dapp_list["dSports and dPredictions"] {
				Control.P_contract.CursorColumn = 1
				Control.P_contract.Refresh()
				Control.S_contract.CursorColumn = 1
				Control.S_contract.Refresh()
			}
		}

		if !b {
			go StopLabel()
			StopGnomon("dReams")
			go SleepLabel()
		}
	})
	Control.daemon_check.Disable()
	Control.daemon_check.Hide()

	return Control.daemon_check
}

// Check box for Holdero SCID connection status
func HolderoContractConnectedBox() fyne.Widget {
	Control.holdero_check = widget.NewCheck("", func(b bool) {
		if !b {
			disableOwnerControls(true)
		}
	})
	Control.holdero_check.Disable()

	return Control.holdero_check
}

// Daemon rpc entry object with default options
//   - Bound to rpc.Daemon.Rpc
func DaemonRpcEntry() fyne.Widget {
	options := []string{"", rpc.DAEMON_RPC_DEFAULT, rpc.DAEMON_RPC_REMOTE1, rpc.DAEMON_RPC_REMOTE2, rpc.DAEMON_RPC_REMOTE5, rpc.DAEMON_RPC_REMOTE6}
	if Control.Daemon_config != "" {
		options = append(options, Control.Daemon_config)
	}
	entry := widget.NewSelectEntry(options)
	entry.PlaceHolder = "Daemon RPC: "

	this := binding.BindString(&rpc.Daemon.Rpc)
	entry.Bind(this)

	return entry
}

// Wallet rpc entry object
//   - Bound to rpc.Wallet.Rpc
//   - Changes reset wallet connection and call CheckConnection()
func WalletRpcEntry() fyne.Widget {
	options := []string{"", "127.0.0.1:10103"}
	entry := widget.NewSelectEntry(options)
	entry.PlaceHolder = "Wallet RPC: "
	entry.OnCursorChanged = func() {
		if rpc.Wallet.Connect {
			rpc.Wallet.Address = ""
			rpc.Display.Wallet_height = "0"
			rpc.Wallet.Height = 0
			rpc.Wallet.Connect = false
			CheckConnection()
		}
	}

	this := binding.BindString(&rpc.Wallet.Rpc)
	entry.Bind(this)

	return entry
}

// Authentication entry object
//   - Bound to rpc.Wallet.UserPass
//   - Changes call rpc.GetAddress() and CheckConnection()
func UserPassEntry() fyne.Widget {
	entry := widget.NewPasswordEntry()
	entry.PlaceHolder = "user:pass"
	entry.OnCursorChanged = func() {
		if rpc.Wallet.Connect {
			rpc.GetAddress("dReams")
			CheckConnection()
		}
	}

	a := binding.BindString(&rpc.Wallet.UserPass)
	entry.Bind(a)

	return entry
}

// Holdero SCID entry
//   - Bound to rpc.Round.Contract
//   - Entry text set on list selection
//   - Changes clear table and check if current entry is valid table
func HolderoContractEntry() fyne.Widget {
	var wait bool
	Poker.contract_input = widget.NewSelectEntry(nil)
	options := []string{""}
	Poker.contract_input.SetOptions(options)
	Poker.contract_input.PlaceHolder = "Holdero Contract Address: "
	Poker.contract_input.OnCursorChanged = func() {
		if rpc.Daemon.Connect && !wait {
			wait = true
			text := Poker.contract_input.Text
			holdero.ClearShared()
			if len(text) == 64 {
				if CheckTableOwner(text) {
					disableOwnerControls(false)
					if checkTableVersion(text) >= 110 {
						Poker.owner.chips.Show()
						Poker.owner.timeout.Show()
						Poker.owner.owners_mid.Show()
					} else {
						Poker.owner.chips.Hide()
						Poker.owner.timeout.Hide()
						Poker.owner.owners_mid.Hide()
					}
				} else {
					disableOwnerControls(true)
				}

				if rpc.Wallet.Connect && CheckHolderoContract(text) {
					holdero.Table.Tournament.Show()
				} else {
					holdero.Table.Tournament.Hide()
				}
			} else {
				rpc.Signal.Contract = false
				Control.holdero_check.SetChecked(false)
				holdero.Table.Tournament.Hide()
			}
			wait = false
		}
	}

	this := binding.BindString(&rpc.Round.Contract)
	Poker.contract_input.Bind(this)

	return Poker.contract_input
}

// Connect button object for rpc
//   - Pressed calls rpc.Ping(), rpc.GetAddress(), CheckConnection(),
//     checks for Holdero key and clears names for population
func RpcConnectButton() fyne.Widget {
	button := widget.NewButton("Connect", func() {
		go func() {
			rpc.Ping()
			rpc.GetAddress("dReams")
			CheckConnection()
			if Control.Dapp_list["Holdero"] {
				Poker.contract_input.CursorColumn = 1
				Poker.contract_input.Refresh()
				rpc.CheckExisitingKey()
				if len(rpc.Wallet.Address) == 66 {
					Control.Names.ClearSelected()
					Control.Names.Options = []string{}
					Control.Names.Refresh()
					Control.Names.Options = append(Control.Names.Options, rpc.Wallet.Address[0:12])
					if Control.Names.Options != nil {
						Control.Names.SetSelectedIndex(0)
					}
				}
			}

			if Control.Dapp_list["dSports and dPredictions"] {
				Control.P_contract.CursorColumn = 1
				Control.P_contract.Refresh()
				Control.S_contract.CursorColumn = 1
				Control.S_contract.Refresh()
			}
		}()
	})

	return button
}

// Routine when Holdero SCID is clicked
func setHolderoControls(str string) (item string) {
	split := strings.Split(str, "   ")
	if len(split) >= 3 {
		trimmed := strings.Trim(split[2], " ")
		if len(trimmed) == 64 {
			item = str
			Poker.contract_input.SetText(trimmed)
			go GetTableStats(trimmed, true)
			rpc.Times.Kick_block = rpc.Wallet.Height
		}
	}

	return
}

// Display SCID rating from dReams SCID rating system
func DisplayRating(i uint64) fyne.Resource {
	if i > 250000 {
		return bundle.ResourceBlueBadge3Png
	} else if i > 150000 {
		return bundle.ResourceBlueBadge2Png
	} else if i > 90000 {
		return bundle.ResourceBlueBadgePng
	} else if i > 50000 {
		return bundle.ResourceRedBadgePng
	} else {
		return nil
	}
}

// Public Holdero table listings object
func TableListings(tab *container.AppTabs) fyne.CanvasObject {
	Poker.Table_list = widget.NewList(
		func() int {
			return len(Control.Holdero_tables)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(canvas.NewImageFromImage(nil), widget.NewLabel(""))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*fyne.Container).Objects[1].(*widget.Label).SetText(Control.Holdero_tables[i])
			if Control.Holdero_tables[i][0:2] != "  " {
				var key string
				split := strings.Split(Control.Holdero_tables[i], "   ")
				if len(split) >= 3 {
					trimmed := strings.Trim(split[2], " ")
					if len(trimmed) == 64 {
						key = trimmed
					}
				}

				badge := canvas.NewImageFromResource(DisplayRating(Control.Contract_rating[key]))
				badge.SetMinSize(fyne.NewSize(35, 35))
				o.(*fyne.Container).Objects[0] = badge
			}
		})

	var item string

	Poker.Table_list.OnSelected = func(id widget.ListItemID) {
		if id != 0 && Connected() {
			go func() {
				item = setHolderoControls(Control.Holdero_tables[id])
				Poker.Favorite_list.UnselectAll()
				Poker.Owned_list.UnselectAll()
			}()
		}
	}

	save_favorite := widget.NewButton("Favorite", func() {
		Control.Holdero_favorites = append(Control.Holdero_favorites, item)
		sort.Strings(Control.Holdero_favorites)
	})

	rate_contract := widget.NewButton("Rate", func() {
		if len(rpc.Round.Contract) == 64 {
			if !CheckTableOwner(rpc.Round.Contract) {
				reset := tab.Selected().Content
				tab.Selected().Content = RateConfirm(rpc.Round.Contract, tab, reset)
				tab.Selected().Content.Refresh()

			} else {
				log.Println("[dReams] You own this contract")
			}
		}
	})

	tables_cont := container.NewBorder(
		nil,
		container.NewBorder(nil, nil, save_favorite, rate_contract, layout.NewSpacer()),
		nil,
		nil,
		Poker.Table_list)

	return tables_cont
}

// Confirmation for a SCID rating
func RateConfirm(scid string, tab *container.AppTabs, reset fyne.CanvasObject) fyne.CanvasObject {
	label := widget.NewLabel(fmt.Sprintf("Rate your experience with this contract\n\n%s", scid))
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	rating_label := widget.NewLabel("")
	rating_label.Wrapping = fyne.TextWrapWord
	rating_label.Alignment = fyne.TextAlignCenter

	fee_label := widget.NewLabel("")
	fee_label.Wrapping = fyne.TextWrapWord
	fee_label.Alignment = fyne.TextAlignCenter

	var slider *widget.Slider
	confirm := widget.NewButton("Confirm", func() {
		var pos uint64
		if slider.Value > 0 {
			pos = 1
		}

		fee := uint64(math.Abs(slider.Value * 10000))
		rpc.RateSCID(scid, fee, pos)
		tab.Selected().Content = reset
		tab.Selected().Content.Refresh()
	})

	confirm.Hide()

	cancel := widget.NewButton("Cancel", func() {
		tab.Selected().Content = reset
		tab.Selected().Content.Refresh()
	})

	slider = widget.NewSlider(-5, 5)
	slider.Step = 0.5
	slider.OnChanged = func(f float64) {
		if slider.Value != 0 {
			rating_label.SetText(fmt.Sprintf("Rating: %.0f", f*10000))
			fee_label.SetText(fmt.Sprintf("Fee: %.5f Dero", math.Abs(f)/10))
			confirm.Show()
		} else {
			rating_label.SetText("Pick a rating")
			fee_label.SetText("")
			confirm.Hide()
		}
	}

	good := canvas.NewImageFromResource(bundle.ResourceBlueBadge3Png)
	good.SetMinSize(fyne.NewSize(30, 30))
	bad := canvas.NewImageFromResource(bundle.ResourceRedBadgePng)
	bad.SetMinSize(fyne.NewSize(30, 30))

	rate_cont := container.NewBorder(nil, nil, bad, good, slider)

	left := container.NewVBox(confirm)
	right := container.NewVBox(cancel)
	buttons := container.NewAdaptiveGrid(2, left, right)

	content := container.NewVBox(layout.NewSpacer(), label, rating_label, fee_label, layout.NewSpacer(), rate_cont, layout.NewSpacer(), buttons)

	return container.NewMax(content)

}

// Favorite Holdero tables object
func HolderoFavorites() fyne.CanvasObject {
	Poker.Favorite_list = widget.NewList(
		func() int {
			return len(Control.Holdero_favorites)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(Control.Holdero_favorites[i])
		})

	var item string

	Poker.Favorite_list.OnSelected = func(id widget.ListItemID) {
		if Connected() {
			item = setHolderoControls(Control.Holdero_favorites[id])
			Poker.Table_list.UnselectAll()
			Poker.Owned_list.UnselectAll()
		}
	}

	remove := widget.NewButton("Remove", func() {
		if len(Control.Holdero_favorites) > 0 {
			Poker.Favorite_list.UnselectAll()
			for i := range Control.Holdero_favorites {
				if Control.Holdero_favorites[i] == item {
					copy(Control.Holdero_favorites[i:], Control.Holdero_favorites[i+1:])
					Control.Holdero_favorites[len(Control.Holdero_favorites)-1] = ""
					Control.Holdero_favorites = Control.Holdero_favorites[:len(Control.Holdero_favorites)-1]
					break
				}
			}
		}
		Poker.Favorite_list.Refresh()
		sort.Strings(Control.Holdero_favorites)
	})

	cont := container.NewBorder(
		nil,
		container.NewBorder(nil, nil, nil, remove, layout.NewSpacer()),
		nil,
		nil,
		Poker.Favorite_list)

	return cont
}

// Owned Holdero tables object
func MyTables() fyne.CanvasObject {
	Poker.Owned_list = widget.NewList(
		func() int {
			return len(Control.Holdero_owned)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(Control.Holdero_owned[i])
		})

	Poker.Owned_list.OnSelected = func(id widget.ListItemID) {
		if Connected() {
			setHolderoControls(Control.Holdero_owned[id])
			Poker.Table_list.UnselectAll()
			Poker.Favorite_list.UnselectAll()
		}
	}

	return Poker.Owned_list
}

// Holdero player name entry
func NameEntry() fyne.CanvasObject {
	Control.Names = widget.NewSelect([]string{}, func(s string) {
		holdero.Poker_name = s
	})

	Control.Names.PlaceHolder = "Name:"

	return Control.Names
}

// Round a float to precision
func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

// Holdero owner control objects, left section
func OwnersBoxLeft(obj []fyne.CanvasObject, tabs *container.AppTabs) fyne.CanvasObject {
	players := []string{"Players", "Close Table", "2 Players", "3 Players", "4 Players", "5 Players", "6 Players"}
	player_select := widget.NewSelect(players, func(s string) {})
	player_select.SetSelectedIndex(0)

	blinds_entry := dwidget.DeroAmtEntry("Big Blind: ", 0.1, 1)
	blinds_entry.SetPlaceHolder("Dero:")
	blinds_entry.SetText("Big Blind: 0.0")
	blinds_entry.Validator = validation.NewRegexp(`^(Big Blind: )\d{1,}\.\d{0,1}$|^(Big Blind: )\d{1,}$`, "Int or float required")
	blinds_entry.OnChanged = func(s string) {
		if blinds_entry.Validate() != nil {
			blinds_entry.SetText("Big Blind: 0.0")
			Poker.owner.blind_amount = 0
		} else {
			trimmed := strings.Trim(s, "Biglnd: ")
			if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
				if uint64(f*100000)%10000 == 0 {
					blinds_entry.SetText(blinds_entry.Prefix + strconv.FormatFloat(roundFloat(f, 1), 'f', int(blinds_entry.Decimal), 64))
					Poker.owner.blind_amount = uint64(roundFloat(f*100000, 1))
				} else {
					blinds_entry.SetText(blinds_entry.Prefix + strconv.FormatFloat(roundFloat(f, 1), 'f', int(blinds_entry.Decimal), 64))
				}
			}
		}
	}

	ante_entry := dwidget.DeroAmtEntry("Ante: ", 0.1, 1)
	ante_entry.SetPlaceHolder("Ante:")
	ante_entry.SetText("Ante: 0.0")
	ante_entry.Validator = validation.NewRegexp(`^(Ante: )\d{1,}\.\d{0,1}$|^(Ante: )\d{1,}$`, "Int or float required")
	ante_entry.OnChanged = func(s string) {
		if ante_entry.Validate() != nil {
			ante_entry.SetText("Ante: 0.0")
			Poker.owner.ante_amount = 0
		} else {
			trimmed := strings.Trim(s, ante_entry.Prefix)
			if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
				if uint64(f*100000)%10000 == 0 {
					ante_entry.SetText(ante_entry.Prefix + strconv.FormatFloat(roundFloat(f, 1), 'f', int(ante_entry.Decimal), 64))
					Poker.owner.ante_amount = uint64(roundFloat(f*100000, 1))
				} else {
					ante_entry.SetText(ante_entry.Prefix + strconv.FormatFloat(roundFloat(f, 1), 'f', int(ante_entry.Decimal), 64))
				}
			}
		}
	}

	options := []string{"DERO", "ASSET"}
	Poker.owner.chips = widget.NewRadioGroup(options, nil)
	Poker.owner.chips.SetSelected("DERO")
	Poker.owner.chips.Horizontal = true
	Poker.owner.chips.OnChanged = func(s string) {
		if s == "ASSET" {
			blinds_entry.Increment = 1
			blinds_entry.Decimal = 0
			blinds_entry.SetText("0")
			blinds_entry.Refresh()

			ante_entry.Increment = 1
			ante_entry.Decimal = 0
			ante_entry.SetText("0")
			ante_entry.Refresh()
		} else {
			blinds_entry.Increment = 0.1
			blinds_entry.Decimal = 1
			blinds_entry.Refresh()

			ante_entry.Increment = 0.1
			ante_entry.Decimal = 1
			ante_entry.Refresh()
		}
	}

	set_button := widget.NewButton("Set Table", func() {
		bb := Poker.owner.blind_amount
		sb := Poker.owner.blind_amount / 2
		ante := Poker.owner.ante_amount
		if holdero.Poker_name != "" {
			rpc.SetTable(player_select.SelectedIndex(), bb, sb, ante, Poker.owner.chips.Selected, holdero.Poker_name, holdero.Settings.AvatarUrl)
		}
	})

	clean_entry := dwidget.DeroAmtEntry("Clean: ", 1, 0)
	clean_entry.AllowFloat = false
	clean_entry.SetPlaceHolder("Atomic:")
	clean_entry.SetText("Clean: 0")
	clean_entry.Validator = validation.NewRegexp(`^(Clean: )\d{1,}`, "Int required")
	clean_entry.OnChanged = func(s string) {
		if clean_entry.Validate() != nil {
			clean_entry.SetText("Clean: 0")
		}
	}

	clean_button := widget.NewButton("Clean Table", func() {
		trimmed := strings.Trim(clean_entry.Text, "Clean: ")
		c, err := strconv.Atoi(trimmed)
		if err == nil {
			rpc.CleanTable(uint64(c))
		} else {
			log.Println("[dReams] Invalid Clean Amount")
		}
	})

	Poker.owner.timeout = widget.NewButton("Timeout", func() {
		obj[1] = TimeOutConfirm(obj, tabs)
		obj[1].Refresh()
	})

	force := widget.NewButton("Force Start", func() {
		rpc.ForceStat()
	})

	players_items := container.NewAdaptiveGrid(2, player_select, layout.NewSpacer())
	blind_items := container.NewAdaptiveGrid(2, blinds_entry, Poker.owner.chips)
	ante_items := container.NewAdaptiveGrid(2, ante_entry, set_button)
	clean_items := container.NewAdaptiveGrid(2, clean_entry, clean_button)
	time_items := container.NewAdaptiveGrid(2, Poker.owner.timeout, force)

	Poker.owner.owners_left = container.NewVBox(players_items, blind_items, ante_items, clean_items, time_items)
	Poker.owner.owners_left.Hide()

	return Poker.owner.owners_left
}

// Holdero owner control objects, middle section
func OwnersBoxMid() fyne.CanvasObject {
	kick_label := widget.NewLabel("      Auto Kick after")
	k_times := []string{"Off", "2m", "5m"}
	auto_remove := widget.NewSelect(k_times, func(s string) {
		switch s {
		case "Off":
			rpc.Times.Kick = 0
		case "2m":
			rpc.Times.Kick = 120
		case "5m":
			rpc.Times.Kick = 300
		default:
			rpc.Times.Kick = 0
		}
	})
	auto_remove.PlaceHolder = "Kick after"

	pay_label := widget.NewLabel("      Payout Delay")
	p_times := []string{"30s", "60s"}
	delay := widget.NewSelect(p_times, func(s string) {
		switch s {
		case "30s":
			rpc.Times.Delay = 30
		case "60s":
			rpc.Times.Delay = 60
		default:
			rpc.Times.Delay = 30
		}
	})
	delay.PlaceHolder = "Payout delay"

	kick := container.NewVBox(layout.NewSpacer(), kick_label, auto_remove)
	pay := container.NewVBox(layout.NewSpacer(), pay_label, delay)

	Poker.owner.owners_mid = container.NewAdaptiveGrid(2, kick, pay)
	Poker.owner.owners_mid.Hide()

	return Poker.owner.owners_mid
}

// Holdero table icon image with frame
func TableIcon(r fyne.Resource) *fyne.Container {
	Stats.Image.SetMinSize(fyne.NewSize(100, 100))
	Stats.Image.Resize(fyne.NewSize(96, 96))
	Stats.Image.Move(fyne.NewPos(8, 3))

	frame := canvas.NewImageFromResource(r)
	frame.Resize(fyne.NewSize(100, 100))
	frame.Move(fyne.NewPos(5, 0))

	cont := container.NewWithoutLayout(&Stats.Image, frame)

	return cont
}

// Holdero table stats display objects
func TableStats() fyne.CanvasObject {
	Stats.Name = canvas.NewText(" Name: ", bundle.TextColor)
	Stats.Desc = canvas.NewText(" Description: ", bundle.TextColor)
	Stats.Version = canvas.NewText(" Table Version: ", bundle.TextColor)
	Stats.Last = canvas.NewText(" Last Move: ", bundle.TextColor)
	Stats.Seats = canvas.NewText(" Table Closed ", bundle.TextColor)

	Stats.Name.TextSize = 18
	Stats.Desc.TextSize = 18
	Stats.Version.TextSize = 18
	Stats.Last.TextSize = 18
	Stats.Seats.TextSize = 18

	Poker.Stats_box = *container.NewVBox(Stats.Name, Stats.Desc, Stats.Version, Stats.Last, Stats.Seats, TableIcon(nil))

	return &Poker.Stats_box
}

// Confirmation of manual Holdero timeout
func TimeOutConfirm(obj []fyne.CanvasObject, reset *container.AppTabs) fyne.CanvasObject {
	var confirm_display = widget.NewLabel("")
	confirm_display.Wrapping = fyne.TextWrapWord
	confirm_display.Alignment = fyne.TextAlignCenter

	confirm_display.SetText("Confirm Time Out on Current Player")

	cancel_button := widget.NewButton("Cancel", func() {
		obj[1] = reset
		obj[1].Refresh()
	})
	confirm_button := widget.NewButton("Confirm", func() {
		rpc.TimeOut()
		obj[1] = reset
		obj[1].Refresh()
	})

	display := container.NewVBox(layout.NewSpacer(), confirm_display, layout.NewSpacer())
	options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewBorder(nil, options, nil, nil, display)

	return container.NewMax(bundle.Alpha120, content)
}

// Confirmation for Holdero contract installs
func HolderoMenuConfirm(c int, obj []fyne.CanvasObject, tabs *container.AppTabs) fyne.CanvasObject {
	gas_fee := 0.3
	unlock_fee := float64(rpc.UnlockFee) / 100000
	var text string
	switch c {
	case 1:
		Poker.Holdero_unlock.Hide()
		text = `You are about to unlock and install your first Holdero Table
		
To help support the project, there is a ` + fmt.Sprintf("%.5f", unlock_fee) + ` DERO donation attached to preform this action

Once you've unlocked a table, you can upload as many new tables free of donation

Total transaction will be ` + fmt.Sprintf("%0.5f", unlock_fee+gas_fee) + ` DERO (0.3 gas fee for contract install)

Select a public or private table

	- Public will show up in indexed list of tables

	- Private will not show up in the list

	- All standard tables can use dReams or DERO

HGC holders can choose to install a HGC table

	- Public table that uses HGC or DERO`
	case 2:
		Poker.Holdero_new.Hide()
		text = `You are about to install a new table

Gas fee to install new table is 0.3 DERO

Select a public or private table

	- Public will show up in indexed list of tables

	- Private will not show up in the list

	- All standard tables can use dReams or DERO

HGC holders can choose to install a HGC table

	- Public table that uses HGC or DERO`
	}

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	var choice *widget.Select
	confirm_button := widget.NewButton("Confirm", func() {
		if choice.SelectedIndex() < 3 && choice.SelectedIndex() >= 0 {
			rpc.UploadHolderoContract(choice.SelectedIndex())
		}

		if c == 2 {
			Poker.Holdero_new.Show()
		}

		obj[1] = tabs
		obj[1].Refresh()
	})

	options := []string{"Public", "Private"}
	if hgc := rpc.TokenBalance(rpc.HgcSCID); hgc > 0 {
		options = append(options, "HGC")
	}

	choice = widget.NewSelect(options, func(s string) {
		if s == "Public" || s == "Private" || s == "HGC" {
			confirm_button.Show()
		} else {
			confirm_button.Hide()
		}
	})

	cancel_button := widget.NewButton("Cancel", func() {
		switch c {
		case 1:
			Poker.Holdero_unlock.Show()
		case 2:
			Poker.Holdero_new.Show()
		default:

		}

		obj[1] = tabs
		obj[1].Refresh()
	})

	confirm_button.Hide()

	left := container.NewVBox(confirm_button)
	right := container.NewVBox(cancel_button)
	buttons := container.NewAdaptiveGrid(2, left, right)
	actions := container.NewVBox(choice, buttons)
	info_box := container.NewVBox(layout.NewSpacer(), label, layout.NewSpacer())

	content := container.NewBorder(nil, actions, nil, nil, info_box)

	go func() {
		for rpc.Wallet.Connect && rpc.Daemon.Connect {
			time.Sleep(time.Second)
		}

		obj[1] = tabs
		obj[1].Refresh()
	}()

	return container.NewMax(content)
}

// Confirmation for dPrediction contract installs
func BettingMenuConfirmP(c int, obj []fyne.CanvasObject, tabs *container.AppTabs) fyne.CanvasObject {
	var text string
	gas_fee := 0.125
	unlock_fee := float64(rpc.UnlockFee) / 100000
	switch c {
	case 1:
		text = `You are about to unlock and install your first dPrediction contract 
		
To help support the project, there is a ` + fmt.Sprintf("%.5f", unlock_fee) + ` DERO donation attached to preform this action

Once you've unlocked dPrediction, you can upload as many new prediction or sports contracts free of donation

Total transaction will be ` + fmt.Sprintf("%0.5f", unlock_fee+gas_fee) + ` DERO (0.125 gas fee for contract install)

Select a public or private contract

	- Public will show up in indexed list of contracts

	- Private will not show up in the list`
	case 2:
		text = `You are about to install a new dPrediction contract. 

Gas fee to install is 0.125 DERO

Select a public or private contract

	- Public will show up in indexed list of contracts

	- Private will not show up in the list`
	}

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	var choice *widget.Select

	pre_button := widget.NewButton("Install", func() {
		if choice.SelectedIndex() < 2 && choice.SelectedIndex() >= 0 {
			rpc.UploadBetContract(true, choice.SelectedIndex())
		}

		obj[1] = tabs
		obj[1].Refresh()
	})

	pre_button.Hide()

	options := []string{"Public", "Private"}
	choice = widget.NewSelect(options, func(s string) {
		if s == "Public" || s == "Private" {
			pre_button.Show()
		} else {
			pre_button.Hide()
		}
	})

	cancel_button := widget.NewButton("Cancel", func() {
		obj[1] = tabs
		obj[1].Refresh()
	})

	left := container.NewVBox(pre_button)
	right := container.NewVBox(cancel_button)
	buttons := container.NewAdaptiveGrid(3, left, container.NewVBox(layout.NewSpacer()), right)
	actions := container.NewVBox(choice, buttons)
	info_box := container.NewVBox(layout.NewSpacer(), label, layout.NewSpacer())

	content := container.NewBorder(nil, actions, nil, nil, info_box)

	go func() {
		for rpc.Wallet.Connect && rpc.Daemon.Connect {
			time.Sleep(time.Second)
		}

		obj[1] = tabs
		obj[1].Refresh()
	}()

	return container.NewMax(content)
}

// Confirmation for dSports contract installs
func BettingMenuConfirmS(c int, obj []fyne.CanvasObject, tabs *container.AppTabs) fyne.CanvasObject {
	var text string
	gas_fee := 0.14
	unlock_fee := float64(rpc.UnlockFee) / 100000
	switch c {
	case 1:
		text = `You are about to unlock and install your first dSports contract
		
To help support the project, there is a ` + fmt.Sprintf("%.5f", unlock_fee) + ` DERO donation attached to preform this action

Once you've unlocked dSports, you can upload as many new sports or predictions contracts free of donation

Total transaction will be ` + fmt.Sprintf("%0.5f", unlock_fee+gas_fee) + ` DERO (0.14 gas fee for contract install)

Select a public or private contract

	- Public will show up in indexed list of contracts

	- Private will not show up in the list`
	case 2:
		text = `You are about to install a new dSports contract

Gas fee to install is 0.14 DERO

Select a public or private contract

	- Public will show up in indexed list of contracts

	- Private will not show up in the list`
	}

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	var choice *widget.Select

	sports_button := widget.NewButton("Install", func() {
		if choice.SelectedIndex() < 2 && choice.SelectedIndex() >= 0 {
			rpc.UploadBetContract(false, choice.SelectedIndex())
		}

		obj[1] = tabs
		obj[1].Refresh()
	})

	sports_button.Hide()

	options := []string{"Public", "Private"}
	choice = widget.NewSelect(options, func(s string) {
		if s == "Public" || s == "Private" {
			sports_button.Show()
		} else {
			sports_button.Hide()
		}
	})

	cancel_button := widget.NewButton("Cancel", func() {
		obj[1] = tabs
		obj[1].Refresh()
	})

	left := container.NewVBox(sports_button)
	right := container.NewVBox(cancel_button)
	buttons := container.NewAdaptiveGrid(3, left, container.NewVBox(layout.NewSpacer()), right)
	actions := container.NewVBox(choice, buttons)
	info_box := container.NewVBox(layout.NewSpacer(), label, layout.NewSpacer())

	content := container.NewBorder(nil, actions, nil, nil, info_box)

	go func() {
		for rpc.Wallet.Connect && rpc.Daemon.Connect {
			time.Sleep(time.Second)
		}

		obj[1] = tabs
		obj[1].Refresh()
	}()

	return container.NewMax(content)
}

// Index entry and NFA control objects
//   - Pass window resources for side menu windows
func IndexEntry(window_icon, window_background fyne.Resource) fyne.CanvasObject {
	Assets.Index_entry = widget.NewMultiLineEntry()
	Assets.Index_entry.PlaceHolder = "SCID:"
	Assets.Index_button = widget.NewButton("Add to Index", func() {
		s := strings.Split(Assets.Index_entry.Text, "\n")
		manualIndex(s)
	})

	Assets.Index_search = widget.NewButton("Search Index", func() {
		searchIndex(Assets.Index_entry.Text)
	})

	Control.Send_asset = widget.NewButton("Send Asset", func() {
		go sendAssetMenu(window_icon, window_background)
	})

	Control.List_button = widget.NewButton("List Asset", func() {
		go listMenu(window_icon, window_background)
	})

	Control.Claim_button = widget.NewButton("Claim NFA", func() {
		if len(Assets.Index_entry.Text) == 64 {
			if isNfa(Assets.Index_entry.Text) {
				rpc.ClaimNFA(Assets.Index_entry.Text)
			}
		}
	})

	Assets.Index_button.Hide()
	Assets.Index_search.Hide()
	Control.List_button.Hide()
	Control.Claim_button.Hide()
	Control.Send_asset.Hide()

	Assets.Gnomes_index = canvas.NewText(" Indexed SCIDs: ", bundle.TextColor)
	Assets.Gnomes_index.TextSize = 18

	bottom_grid := container.NewAdaptiveGrid(3, Assets.Gnomes_index, Assets.Index_button, Assets.Index_search)
	top_grid := container.NewAdaptiveGrid(3, container.NewMax(Control.Send_asset), Control.Claim_button, Control.List_button)
	box := container.NewVBox(top_grid, layout.NewSpacer(), bottom_grid)

	return container.NewAdaptiveGrid(2, Assets.Index_entry, box)
}

// Owned asset list object
//   - Sets Control.Viewing_asset and asset stats on selected
func AssetList() fyne.CanvasObject {
	Assets.Asset_list = widget.NewList(
		func() int {
			return len(Assets.Assets)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(Assets.Assets[i])
		})

	Assets.Asset_list.OnSelected = func(id widget.ListItemID) {
		split := strings.Split(Assets.Assets[id], "   ")
		if len(split) >= 2 {
			trimmed := strings.Trim(split[1], " ")
			Control.Viewing_asset = trimmed
			Assets.Icon = *canvas.NewImageFromImage(nil)
			go GetOwnedAssetStats(trimmed)
		}
	}

	return container.NewMax(Assets.Asset_list)
}

// Send Dero asset menu
//   - Asset SCID can be sent as payload to receiver when sending asset
//   - Pass resources for window
func sendAssetMenu(window_icon, background fyne.Resource) {
	Control.send_open = true
	saw := fyne.CurrentApp().NewWindow("Send Asset")
	saw.Resize(fyne.NewSize(330, 700))
	saw.SetIcon(window_icon)
	Control.Send_asset.Hide()
	Control.List_button.Hide()
	saw.SetCloseIntercept(func() {
		Control.send_open = false
		if rpc.Wallet.Connect {
			Control.Send_asset.Show()
			if isNfa(Control.Viewing_asset) {
				Control.List_button.Show()
			}
		}
		saw.Close()
	})
	saw.SetFixedSize(true)

	var saw_content *fyne.Container
	var send_button *widget.Button
	background_img := *canvas.NewImageFromResource(background)

	viewing_asset := Control.Viewing_asset

	viewing_label := widget.NewLabel(fmt.Sprintf("Sending SCID:\n\n%s\n\nEnter destination address below\n\nSCID can be sent to reciever as payload\n\n", viewing_asset))
	viewing_label.Wrapping = fyne.TextWrapWord
	viewing_label.Alignment = fyne.TextAlignCenter

	info_label := widget.NewLabel("Enter all info before sending")
	payload := widget.NewCheck("Send SCID as payload", func(b bool) {})

	dest_entry := widget.NewMultiLineEntry()
	dest_entry.SetPlaceHolder("Destination Address:")
	dest_entry.Wrapping = fyne.TextWrapWord
	dest_entry.Validator = validation.NewRegexp(`^(dero)\w{62}$`, "Invalid Address")
	dest_entry.OnChanged = func(s string) {
		if dest_entry.Validate() == nil {
			info_label.SetText("")
			send_button.Show()
		} else {
			info_label.SetText("Enter destination address.")
			send_button.Hide()
		}
	}

	var dest string
	var confirm_open bool
	send_button = widget.NewButton("Send Asset", func() {
		if dest_entry.Validate() == nil {
			confirm_open = true
			send_asset := viewing_asset
			var load bool
			if payload.Checked {
				load = true
			}

			confirm_button := widget.NewButton("Confirm", func() {
				if dest_entry.Validate() == nil {
					var load bool
					if payload.Checked {
						load = true
					}
					go rpc.SendAsset(send_asset, dest, load)
					saw.Close()
				}
			})

			cancel_button := widget.NewButton("Cancel", func() {
				confirm_open = false
				saw.SetContent(
					container.New(layout.NewMaxLayout(),
						&background_img,
						bundle.Alpha180,
						saw_content))
			})

			dest = dest_entry.Text
			confirm_label := widget.NewLabel(fmt.Sprintf("Sending SCID:\n\n%s\n\nDestination: %s\n\nSending SCID as payload: %t", send_asset, dest, load))
			confirm_label.Wrapping = fyne.TextWrapWord
			confirm_label.Alignment = fyne.TextAlignCenter

			confirm_display := container.NewVBox(confirm_label, layout.NewSpacer())
			confirm_options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
			confirm_content := container.NewBorder(nil, confirm_options, nil, nil, confirm_display)
			saw.SetContent(
				container.New(layout.NewMaxLayout(),
					&background_img,
					bundle.Alpha180,
					confirm_content))
		}
	})
	send_button.Hide()

	icon := Assets.Icon

	saw_content = container.NewVBox(
		viewing_label,
		menuAssetImg(&icon, bundle.ResourceAvatarFramePng),
		layout.NewSpacer(),
		dest_entry,
		container.NewCenter(payload),
		layout.NewSpacer(),
		container.NewAdaptiveGrid(2, layout.NewSpacer(), send_button))

	go func() {
		for rpc.Wallet.Connect && rpc.Daemon.Connect {
			time.Sleep(3 * time.Second)
			if !confirm_open {
				icon = Assets.Icon
				saw_content.Objects[1] = menuAssetImg(&icon, bundle.ResourceAvatarFramePng)
				if viewing_asset != Control.Viewing_asset {
					viewing_asset = Control.Viewing_asset
					viewing_label.SetText("Sending SCID:\n\n" + viewing_asset + " \n\nEnter destination address below\n\nSCID can be sent to reciever as payload\n\n")
				}
				saw_content.Refresh()
			}
		}
		Control.send_open = false
		saw.Close()
	}()

	saw.SetContent(
		container.New(layout.NewMaxLayout(),
			&background_img,
			bundle.Alpha180,
			saw_content))
	saw.Show()
}

// Image for send asset and list menus
//   - Pass res for frame resource
func menuAssetImg(img *canvas.Image, res fyne.Resource) fyne.CanvasObject {
	img.SetMinSize(fyne.NewSize(100, 100))
	img.Resize(fyne.NewSize(94, 94))
	img.Move(fyne.NewPos(118, 3))

	frame := canvas.NewImageFromResource(res)
	frame.Resize(fyne.NewSize(100, 100))
	frame.Move(fyne.NewPos(115, 0))

	cont := container.NewWithoutLayout(img, frame)

	return cont
}

// NFA listing menu
//   - Pass resources for menu window to match main
func listMenu(window_icon, background fyne.Resource) {
	Control.list_open = true
	aw := fyne.CurrentApp().NewWindow("List NFA")
	aw.Resize(fyne.NewSize(330, 700))
	aw.SetIcon(window_icon)
	Control.List_button.Hide()
	Control.Send_asset.Hide()
	aw.SetCloseIntercept(func() {
		Control.list_open = false
		if rpc.Wallet.Connect {
			Control.Send_asset.Show()
			if isNfa(Control.Viewing_asset) {
				Control.List_button.Show()
			}
		}
		aw.Close()
	})
	aw.SetFixedSize(true)

	var aw_content *fyne.Container
	var set_list *widget.Button
	background_img := *canvas.NewImageFromResource(background)

	viewing_asset := Control.Viewing_asset
	viewing_label := widget.NewLabel(fmt.Sprintf("Listing SCID:\n\n%s", viewing_asset))
	viewing_label.Wrapping = fyne.TextWrapWord
	viewing_label.Alignment = fyne.TextAlignCenter

	fee_label := widget.NewLabel(fmt.Sprintf("Listing fee %.5f Dero", float64(rpc.ListingFee)/100000))

	listing_options := []string{"Auction", "Sale"}
	listing := widget.NewSelect(listing_options, func(s string) {})
	listing.PlaceHolder = "Type:"

	duration := dwidget.DeroAmtEntry("", 1, 0)
	duration.AllowFloat = false
	duration.SetPlaceHolder("Duration in Hours:")
	duration.Validator = validation.NewRegexp(`^[^0]\d{0,2}$`, "Int required")

	start := dwidget.DeroAmtEntry("", 0.1, 1)
	start.AllowFloat = true
	start.SetPlaceHolder("Start Price:")
	start.Validator = validation.NewRegexp(`^\d{1,}\.\d{1,5}$|^[^0]\d{0,}$`, "Int or float required")

	charAddr := widget.NewEntry()
	charAddr.SetPlaceHolder("Charity Donation Address:")
	charAddr.Validator = validation.NewRegexp(`^(dero)\w{62}$`, "Int required")

	charPerc := dwidget.DeroAmtEntry("", 1, 0)
	charPerc.AllowFloat = false
	charPerc.SetPlaceHolder("Charity Donation %:")
	charPerc.Validator = validation.NewRegexp(`^\d{1,2}$`, "Int required")
	charPerc.OnChanged = func(s string) {
		if listing.Selected != "" && duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
			set_list.Show()
		} else {
			set_list.Hide()
		}
	}

	duration.OnChanged = func(s string) {
		if rpc.StringToInt(s) > 168 {
			duration.SetText("168")
		}

		if listing.Selected != "" && duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
			set_list.Show()
		} else {
			set_list.Hide()
		}
	}

	start.OnChanged = func(s string) {
		if listing.Selected != "" && duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
			set_list.Show()
		} else {
			set_list.Hide()
		}
	}

	charAddr.OnChanged = func(s string) {
		if listing.Selected != "" && duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
			set_list.Show()
		} else {
			set_list.Hide()
		}
	}

	var confirm_open bool
	set_list = widget.NewButton("Set Listing", func() {
		if duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
			if listing.Selected != "" {
				confirm_open = true
				listing_asset := viewing_asset
				artP, royaltyP := GetListingPercents(listing_asset)

				d := uint64(stringToInt64(duration.Text))
				s := ToAtomicFive(start.Text)
				sp := float64(s) / 100000
				cp := uint64(stringToInt64(charPerc.Text))

				art_gets := (float64(s) * artP) / 100000
				royalty_gets := (float64(s) * royaltyP) / 100000
				char_gets := float64(s) * (float64(cp) / 100) / 100000

				total := sp - art_gets - royalty_gets - char_gets

				first_line := fmt.Sprintf("Listing SCID:\n\n%s\n\nList Type: %s\n\nDuration: %s Hours\n\nStart Price: %0.5f Dero\n\n", listing_asset, listing.Selected, duration.Text, sp)
				second_line := fmt.Sprintf("Artificer Fee: %.0f%s - %0.5f Dero\n\nRoyalties: %.0f%s - %0.5f Dero\n\n", artP*100, "%", art_gets, royaltyP*100, "%", royalty_gets)
				third_line := fmt.Sprintf("Chairity Address: %s\n\nCharity Percent: %s%s - %0.5f Dero\n\nYou will receive %.5f Dero if asset sells at start price", charAddr.Text, charPerc.Text, "%", char_gets, total)

				confirm_label := widget.NewLabel(first_line + second_line + third_line)
				confirm_label.Wrapping = fyne.TextWrapWord
				confirm_label.Alignment = fyne.TextAlignCenter

				cancel_button := widget.NewButton("Cancel", func() {
					confirm_open = false
					aw.SetContent(
						container.New(layout.NewMaxLayout(),
							&background_img,
							bundle.Alpha180,
							aw_content))
				})

				confirm_button := widget.NewButton("Confirm", func() {
					rpc.SetNFAListing(listing_asset, listing.Selected, charAddr.Text, d, s, cp)
					Control.list_open = false
					if rpc.Wallet.Connect {
						Control.Send_asset.Show()
						if isNfa(Control.Viewing_asset) {
							Control.List_button.Show()
						}
					}
					aw.Close()
				})

				confirm_options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
				confirm_content := container.NewBorder(nil, confirm_options, nil, nil, confirm_label)

				aw.SetContent(
					container.New(layout.NewMaxLayout(),
						&background_img,
						bundle.Alpha180,
						confirm_content))
			}
		}
	})
	set_list.Hide()

	icon := Assets.Icon

	go func() {
		for rpc.Wallet.Connect && rpc.Daemon.Connect {
			time.Sleep(3 * time.Second)
			if !confirm_open && isNfa(Control.Viewing_asset) {
				icon = Assets.Icon
				aw_content.Objects[2] = menuAssetImg(&icon, bundle.ResourceAvatarFramePng)
				if viewing_asset != Control.Viewing_asset {
					viewing_asset = Control.Viewing_asset
					viewing_label.SetText(fmt.Sprintf("Listing SCID:\n\n%s", viewing_asset))
				}
				aw_content.Refresh()
			}
		}
		Control.list_open = false
		aw.Close()
	}()

	aw_content = container.NewVBox(
		viewing_label,
		layout.NewSpacer(),
		menuAssetImg(&icon, bundle.ResourceAvatarFramePng),
		layout.NewSpacer(),
		layout.NewSpacer(),
		listing,
		duration,
		start,
		charAddr,
		charPerc,
		container.NewCenter(fee_label),
		container.NewAdaptiveGrid(2, layout.NewSpacer(), set_list))

	aw.SetContent(
		container.New(layout.NewMaxLayout(),
			&background_img,
			bundle.Alpha180,
			aw_content))
	aw.Show()
}

// Convert string to atomic value
func ToAtomicFive(v string) uint64 {
	f, err := strconv.ParseFloat(v, 64)

	if err != nil {
		log.Println("[ToAtomicFive]", err)
		return 0
	}

	ratio := math.Pow(10, float64(5))
	rf := math.Round(f*ratio) / ratio

	return uint64(math.Round(rf * 100000))
}

// Menu instruction tree
func IntroTree() fyne.CanvasObject {
	list := map[string][]string{
		"":                        {"Welcome to dReams"},
		"Welcome to dReams":       {"Get Started", "dApps", "Assets", "Market"},
		"Get Started":             {"Visit dero.io for daemon and wallet download info", "Connecting", "FAQ"},
		"Connecting":              {"Daemon", "Wallet"},
		"FAQ":                     {"Can't connect", "How to resync Gnomon db", "Can't see any tables, contracts or market info", "How to see terminal log"},
		"Can't connect":           {"Using a local daemon will yeild the best results", "If you are using a remote daemon, try changing daemons", "Any connection errors can be found in terminal log"},
		"How to resync Gnomon db": {"Shut down dReams", "Find and delete the Gnomon db folder that is in your dReams directory", "Restart dReams and connect to resync db", "Any sync errors can be found in terminal log"},
		"Can't see any tables, contracts or market info": {"Make sure daemon, wallet and Gnomon indicators are lit up solid", "If you've added new dApps to your dReams, a Gnomon resync will add them to your index", "Look in the asset tab for number of indexed SCIDs", "If indexed SCIDs is less than 4000 your db is not fully synced", "Try resyncing", "Any errors can be found in terminal log"},
		"How to see terminal log":                        {"Windows", "Mac", "Linux"},
		"Windows":                                        {"Open powershell or command prompt", "Navigate to dReams directory", `Start dReams using       .\dReams-windows-amd64.exe`},
		"Mac":                                            {"Open a terminal", "Navigate to dReams directory", `Start dReams using       ./dReams-macos-amd64`},
		"Linux":                                          {"Open a terminal", "Navigate to dReams directory", `Start dReams using       ./dReams-linux-amd64`},
		"Daemon":                                         {"Using local daemon will give best performance while using dReams", "Remote daemon options are available in drop down if a local daemon is not available", "Enter daemon address and the D light in top right will light up if connection is successful", "Once daemon is connected Gnomon will start up, the Gnomon indicator light will have a stripe in middle"},
		"Wallet":                                         {"Set up and register a Dero wallet", "Your wallet will need to be running rpc server", "Using cli, start your wallet with flags --rpc-server --rpc-login=user:pass", "With Engram, turn on cyberdeck to start rpc server", "In dReams enter your wallet rpc address and rpc user:pass", "Press connect and the W light in top right will light up if connection is successful", "Once wallet is connected and Gnomon is running, Gnomon will sync with wallet", "The Gnomon indicator will turn solid when this is complete, everything is now connected"},

		"dApps":                 {"Holdero", "Baccarat", "Predictions", "Sports", "dReam Service", "Tarot", "DerBnb", "Contract Ratings"},
		"Holdero":               {"Multiplayer Texas Hold'em style on chian poker", "No limit, single raise game. Table owners choose game params", "Six players max at a table", "No side pots, must call or fold", "Standard tables can be public or private, and can use Dero or dReam Tokens", "dReam Tools", "Tournament tables can be set up to use any Token", "View table listings or launch your own Holdero contract in the owned tab"},
		"dReam Tools":           {"A suite of tools for Holdero, unlocked with ownership of a AZY or SIX playing card assets", "Odds calculator", "Bot player with 12 customizable parameters", "Track playing stats for user and bot players"},
		"Baccarat":              {"A popular table game, where closest to 9 wins", "Bet on player, banker or tie as the winnng outcome", "Select table with bottom left drop down to choose currency"},
		"Predictions":           {"Prediction contracts are for binary based predictions, (higher/lower, yes/no)", "How predictions works", "Current Markets", "dReams Client aggregated price feed", "View active prediction contracts in predictions tab or launch your own prediction contract in the owned tab"},
		"How predictions works": {"P2P predictions", "Variable time limits allowing for different prediction set ups, each contract runs one prediction at a time", "Click a contract from the list to view it", "Closes at, is when the contract will stop accepting predictions", "Mark (price or value you are predicting on) can be set on prediction initialization or it can given live", "Posted with in, is the acceptable time frame to post the live Mark", "If Mark is not posted, prediction is voided and you will be refunded", "Payout after, is when the Final price is posted and compared to the mark to determine winners", "If the final price is not posted with in refund time frame, prediction is void and you will be refunded"},
		"Current Markets":       {"DERO-BTC", "XMR-BTC", "BTC-USDT", "DERO-USDT", "XMR-USDT", "DERO-Difficulty", "DERO-Block Time", "DERO-Block Number"},
		"Sports":                {"Sports contracts are for sports wagers", "How sports works", "Current Leagues", "Live game scores, and game schedules", "View active sports contracts in sports tab or launch your own sports contract in the owned tab"},
		"How sports works":      {"P2P betting", "Variable time limits, one contract can run multiple games at the same time", "Click a contract from the list to view it", "Any active games on the contract will populate, you can pick which game you'd like to play from the drop down", "Closes at, is when the contracts stops accepting picks", "Default payout time after close is 4hr, this is when winner will be posted from client feed", "Default refund time is 8hr after close, meaning if winner is not provided past that time you will be refunded", "A Tie refunds pot to all all participants"},
		"Current Leagues":       {"EPL", "MLS", "FIFA", "NBA", "NFL", "NHL", "MLB", "Bellator", "UFC"},
		"dReam Service":         {"dReam Service is unlocked for all betting contract owners", "Full automation of contract posts and payouts", "Integrated address service allows bets to be placed thorugh a Dero transaction to sent to service", "Multiple owners can be added to contracts and multiple service wallets can be ran on one contract", "Stand alone cli app availible for streamlined use"},
		"Tarot":                 {"On chian Tarot readings", "Iluma cards and readings created by Kalina Lux"},
		"DerBnb":                {"A property rental platform", "Users can mint properties as contracts and list for rentals", "Property owners can choose rates, damage deposits and availabilty dates", "Dero messaging helps owners and renters facilitate the final details of rental privately", "Rating system for properties"},
		"Contract Ratings":      {"dReam Tables has a public rating store on chain for multiplayer contracts", "Players can rate other contracts positively or negatively", "Four rating tiers, tier two being the starting tier for all contracts", "Each rating transaction is weight based by its Dero value", "Contracts that fall below tier one will no longer populate in the public index"},
		"Assets":                {"View any owned assets held in wallet", "Put owned assets up for auction or for sale", "Send assets privately to another wallet", "Indexer, add custom contracts to your index and search current index db"},
		"Market":                {"View any in game assets up for auction or sale", "Bid on or buy assets", "Cancel or close out any existing listings"},
	}

	tree := widget.NewTreeWithStrings(list)

	tree.OnBranchClosed = func(uid widget.TreeNodeID) {
		tree.UnselectAll()
		if uid == "Welcome to dReams" {
			tree.CloseAllBranches()
		}
	}

	tree.OnBranchOpened = func(uid widget.TreeNodeID) {
		tree.Select(uid)
	}

	tree.OpenBranch("Welcome to dReams")

	max := container.NewMax(tree)

	return max
}

// Send Dero message menu
func SendMessageMenu(window_icon, background fyne.Resource) {
	if !Control.msg_open {
		Control.msg_open = true
		smw := fyne.CurrentApp().NewWindow("Send Asset")
		smw.Resize(fyne.NewSize(330, 700))
		smw.SetIcon(window_icon)
		smw.SetCloseIntercept(func() {
			Control.msg_open = false
			smw.Close()
		})
		smw.SetFixedSize(true)

		var send_button *widget.Button
		img := *canvas.NewImageFromResource(background)

		label := widget.NewLabel("Sending Message:\n\nEnter ringsize and destination address below")
		label.Wrapping = fyne.TextWrapWord
		label.Alignment = fyne.TextAlignCenter

		ringsize := widget.NewSelect([]string{"16", "32", "64"}, func(s string) {})
		ringsize.PlaceHolder = "Ringsize:"
		ringsize.SetSelectedIndex(1)

		message_entry := widget.NewMultiLineEntry()
		message_entry.SetPlaceHolder("Message:")
		message_entry.Wrapping = fyne.TextWrapWord

		dest_entry := widget.NewMultiLineEntry()
		dest_entry.SetPlaceHolder("Destination Address:")
		dest_entry.Wrapping = fyne.TextWrapWord
		dest_entry.Validator = validation.NewRegexp(`^(dero)\w{62}$`, "Invalid Address")
		dest_entry.OnChanged = func(s string) {
			if dest_entry.Validate() == nil && message_entry.Text != "" {
				send_button.Show()
			} else {
				send_button.Hide()
			}
		}

		message_entry.OnChanged = func(s string) {
			if s != "" && dest_entry.Validate() == nil {
				send_button.Show()
			} else {
				send_button.Hide()
			}
		}

		send_button = widget.NewButton("Send Message", func() {
			if dest_entry.Validate() == nil && message_entry.Text != "" {
				rings := stringToInt64(ringsize.Selected)
				go rpc.SendMessage(dest_entry.Text, dest_entry.Text, uint64(rings))
				Control.msg_open = false
				smw.Close()
			}
		})
		send_button.Hide()

		dest_cont := container.NewVBox(label, ringsize, dest_entry)
		message_cont := container.NewBorder(nil, send_button, nil, nil, message_entry)

		content := container.NewVSplit(dest_cont, message_cont)

		go func() {
			for rpc.Wallet.Connect && rpc.Daemon.Connect {
				time.Sleep(3 * time.Second)
			}
			Control.msg_open = false
			smw.Close()
		}()

		smw.SetContent(
			container.New(layout.NewMaxLayout(),
				&img,
				bundle.Alpha180,
				content))
		smw.Show()
	}
}
