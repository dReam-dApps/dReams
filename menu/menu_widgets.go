package menu

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/table"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	DAEMON_RPC_DEFAULT = "127.0.0.1:10102"
	DAEMON_RPC_REMOTE1 = "89.38.99.117:10102"
	DAEMON_RPC_REMOTE2 = "publicrpc1.dero.io:10102"
	// DAEMON_RPC_REMOTE3 = "dero-node.mysrv.cloud:10102"
	// DAEMON_RPC_REMOTE4 = "derostats.io:10102"
	DAEMON_RPC_REMOTE5 = "85.17.52.28:11012"
	DAEMON_RPC_REMOTE6 = "node.derofoundation.org:11012"
)

type menuOptions struct {
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
	Sports_check      *widget.Check
	Wallet_ind        *fyne.Animation
	Daemon_ind        *fyne.Animation
	Poker_ind         *fyne.Animation
	Service_ind       *fyne.Animation
}

type holderoOptions struct {
	contract_input *widget.SelectEntry
	Table_list     *widget.List
	Favorite_list  *widget.List
	Owned_list     *widget.List
	Holdero_unlock *widget.Button
	Holdero_new    *widget.Button
	Stats_box      fyne.Container
}

type tableOwnerOptions struct {
	blindAmount uint64
	anteAmount  uint64
	chips       *widget.RadioGroup
	timeout     *widget.Button
	owners_left *fyne.Container
	owners_mid  *fyne.Container
}

type resources struct {
	SmallIcon fyne.Resource
	Frame     fyne.Resource
	Back1     fyne.Resource
	Back2     fyne.Resource
	Back3     fyne.Resource
	Back4     fyne.Resource
	Gnomon    fyne.Resource
	BBadge    fyne.Resource
	B2Badge   fyne.Resource
	B3Badge   fyne.Resource
	RBadge    fyne.Resource
	PBot      fyne.Resource
	dService  fyne.Resource
	Tools     fyne.Resource
}

var Resource resources
var HolderoControl holderoOptions
var MenuControl menuOptions
var ownerControl tableOwnerOptions

// Get menu resources from main
func GetMenuResources(r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11, r12, r13, r14 fyne.Resource) {
	Resource.SmallIcon = r1
	Resource.Frame = r2
	Resource.Back1 = r3
	Resource.Back2 = r4
	Resource.Back3 = r5
	Resource.Back4 = r6
	Resource.Gnomon = r7
	Resource.BBadge = r8
	Resource.B2Badge = r9
	Resource.B3Badge = r10
	Resource.RBadge = r11
	Resource.PBot = r12
	Resource.dService = r13
	Resource.Tools = r14
}

// Do when disconnected
func disconnected() {
	rpc.Wallet.Service = false
	rpc.Wallet.PokerOwner = false
	rpc.Wallet.BetOwner = false
	rpc.Round.ID = 0
	rpc.Display.PlayerId = ""
	rpc.Odds.Run = false
	table.Settings.FaceSelect.Options = []string{"Light", "Dark"}
	table.Settings.BackSelect.Options = []string{"Light", "Dark"}
	table.Settings.ThemeSelect.Options = []string{"Main"}
	table.Settings.AvatarSelect.Options = []string{"None"}
	table.Settings.FaceUrl = ""
	table.Settings.BackUrl = ""
	table.Settings.AvatarUrl = ""
	table.Settings.FaceSelect.SetSelectedIndex(0)
	table.Settings.BackSelect.SetSelectedIndex(0)
	table.Settings.AvatarSelect.SetSelectedIndex(0)
	table.Settings.FaceSelect.Refresh()
	table.Settings.BackSelect.Refresh()
	table.Settings.ThemeSelect.Refresh()
	table.Settings.AvatarSelect.Refresh()
	table.Assets.Assets = []string{}
	table.Assets.Name.Text = (" Name:")
	table.Assets.Name.Refresh()
	table.Assets.Collection.Text = (" Collection:")
	table.Assets.Collection.Refresh()
	table.Assets.Icon = *canvas.NewImageFromImage(nil)
	// prediction leaderboard
	// table.Actions.NameEntry.Text = ""
	// table.Actions.NameEntry.Enable()
	// table.Actions.NameEntry.Refresh()
	table.DisableHolderoTools()
	MenuControl.Names.ClearSelected()
	MenuControl.Names.Options = []string{}
	MenuControl.Names.Refresh()
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
	MenuControl.Holdero_tables = []string{}
	MenuControl.Predict_contracts = []string{}
	MenuControl.Sports_contracts = []string{}
	MenuControl.Holdero_owned = []string{}
	MenuControl.Predict_owned = []string{}
	MenuControl.Sports_owned = []string{}
	Market.Auctions = []string{}
	Market.Buy_now = []string{}
	table.Assets.Assets = []string{}
}

// Connection check for main process
func CheckConnection() {
	if rpc.Signal.Daemon {
		MenuControl.daemon_check.SetChecked(true)
		disableIndex(false)
	} else {
		MenuControl.daemon_check.SetChecked(false)
		MenuControl.holdero_check.SetChecked(false)
		if MenuControl.Dapp_list["dSports and dPredictions"] {
			MenuControl.Predict_check.SetChecked(false)
			MenuControl.Sports_check.SetChecked(false)
		}
		rpc.Signal.Contract = false
		clearContractLists()
		disableOwnerControls(true)
		disableBaccActions(true)
		disableActions(true)
		disableIndex(true)
		Gnomes.Init = false
		Gnomes.Checked = false
	}

	if rpc.Wallet.Connect {
		disableActions(false)
	} else {
		MenuControl.holdero_check.SetChecked(false)
		if MenuControl.Dapp_list["dSports and dPredictions"] {
			MenuControl.Predict_check.SetChecked(false)
			MenuControl.Sports_check.SetChecked(false)
			DisablePreditions(true)
			disableSports(true)
		}
		rpc.Signal.Contract = false
		clearContractLists()
		disableOwnerControls(true)
		disableBaccActions(true)
		disableActions(true)
		disconnected()
		Gnomes.Checked = false
	}

	if rpc.Signal.Contract {
		MenuControl.holdero_check.SetChecked(true)
	} else {
		MenuControl.holdero_check.SetChecked(false)
		disableOwnerControls(true)
		rpc.Signal.Sit = true
	}
}

// Hiden object, controls Gnomon start and stop based on daemon connection
func DaemonConnectedBox() fyne.Widget {
	MenuControl.daemon_check = widget.NewCheck("", func(b bool) {
		if !Gnomes.Init && !Gnomes.Start {
			startGnomon(rpc.Round.Daemon)
			HolderoControl.contract_input.CursorColumn = 1
			HolderoControl.contract_input.Refresh()
			if MenuControl.Dapp_list["dSports and dPredictions"] {
				table.Actions.P_contract.CursorColumn = 1
				table.Actions.P_contract.Refresh()
				table.Actions.S_contract.CursorColumn = 1
				table.Actions.S_contract.Refresh()
			}
		}

		if !b {
			StopGnomon(Gnomes.Init)
		}
	})
	MenuControl.daemon_check.Disable()
	MenuControl.daemon_check.Hide()

	return MenuControl.daemon_check
}

// Check box for Holdero SCID connection status
func HolderoContractConnectedBox() fyne.Widget {
	MenuControl.holdero_check = widget.NewCheck("", func(b bool) {
		if !b {
			disableOwnerControls(true)
		}
	})
	MenuControl.holdero_check.Disable()

	return MenuControl.holdero_check
}

// Daemon rpc entry object
func DaemonRpcEntry() fyne.Widget {
	var options = []string{"", DAEMON_RPC_DEFAULT, DAEMON_RPC_REMOTE1, DAEMON_RPC_REMOTE2, DAEMON_RPC_REMOTE5, DAEMON_RPC_REMOTE6}
	if MenuControl.Daemon_config != "" {
		options = append(options, MenuControl.Daemon_config)
	}
	entry := widget.NewSelectEntry(options)
	entry.PlaceHolder = "Daemon RPC: "

	this := binding.BindString(&rpc.Round.Daemon)
	entry.Bind(this)

	return entry
}

// Wallet rpc entry object
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
func UserPassEntry() fyne.Widget {
	entry := widget.NewPasswordEntry()
	entry.PlaceHolder = "user:pass"
	entry.OnCursorChanged = func() {
		if rpc.Wallet.Connect {
			rpc.GetAddress()
			CheckConnection()
		}
	}

	a := binding.BindString(&rpc.Wallet.UserPass)
	entry.Bind(a)

	return entry
}

// Holdero SCID entry
func HolderoContractEntry() fyne.Widget {
	var wait bool
	HolderoControl.contract_input = widget.NewSelectEntry(nil)
	options := []string{""}
	HolderoControl.contract_input.SetOptions(options)
	HolderoControl.contract_input.PlaceHolder = "Holdero Contract Address: "
	HolderoControl.contract_input.OnCursorChanged = func() {
		if rpc.Signal.Daemon && !wait {
			wait = true
			text := HolderoControl.contract_input.Text
			table.ClearShared()
			if len(text) == 64 {
				if CheckTableOwner(text) {
					disableOwnerControls(false)
					if checkTableVersion(text) >= 110 {
						ownerControl.chips.Show()
						ownerControl.timeout.Show()
						ownerControl.owners_mid.Show()
					} else {
						ownerControl.chips.Hide()
						ownerControl.timeout.Hide()
						ownerControl.owners_mid.Hide()
					}
				} else {
					disableOwnerControls(true)
				}

				tourney := CheckHolderoContract(text)
				if rpc.Wallet.Connect && tourney {
					table.Actions.Tournament.Show()
				} else {
					table.Actions.Tournament.Hide()
				}
			} else {
				rpc.Signal.Contract = false
				MenuControl.holdero_check.SetChecked(false)
				table.Actions.Tournament.Hide()
			}
			wait = false
		}
	}

	this := binding.BindString(&rpc.Round.Contract)
	HolderoControl.contract_input.Bind(this)

	return HolderoControl.contract_input
}

// Connect to entered rpc addresses
func RpcConnectButton() fyne.Widget {
	button := widget.NewButton("Connect", func() {
		go func() {
			rpc.Ping()
			rpc.GetAddress()
			CheckConnection()
			HolderoControl.contract_input.CursorColumn = 1
			HolderoControl.contract_input.Refresh()
			if MenuControl.Dapp_list["dSports and dPredictions"] {
				table.Actions.P_contract.CursorColumn = 1
				table.Actions.P_contract.Refresh()
				table.Actions.S_contract.CursorColumn = 1
				table.Actions.S_contract.Refresh()
			}

			rpc.CheckExisitingKey()
			if len(rpc.Wallet.Address) == 66 {
				MenuControl.Names.ClearSelected()
				MenuControl.Names.Options = []string{}
				MenuControl.Names.Refresh()
				MenuControl.Names.Options = append(MenuControl.Names.Options, rpc.Wallet.Address[0:12])
				if MenuControl.Names.Options != nil {
					MenuControl.Names.SetSelectedIndex(0)
				}
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
			HolderoControl.contract_input.SetText(trimmed)
			go GetTableStats(trimmed, true)
			rpc.Times.Kick_block = rpc.Wallet.Height
		}
	}

	return
}

// Display SCID rating from dReams SCID rating system
func DisplayRating(i uint64) fyne.Resource {
	if i > 250000 {
		return Resource.B3Badge
	} else if i > 150000 {
		return Resource.B2Badge
	} else if i > 90000 {
		return Resource.BBadge
	} else if i > 50000 {
		return Resource.RBadge
	} else {
		return nil
	}
}

// Public Holdero table listings object
func TableListings(tab *container.AppTabs) fyne.CanvasObject {
	HolderoControl.Table_list = widget.NewList(
		func() int {
			return len(MenuControl.Holdero_tables)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(canvas.NewImageFromImage(nil), widget.NewLabel(""))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*fyne.Container).Objects[1].(*widget.Label).SetText(MenuControl.Holdero_tables[i])
			if MenuControl.Holdero_tables[i][0:2] != "  " {
				var key string
				split := strings.Split(MenuControl.Holdero_tables[i], "   ")
				if len(split) >= 3 {
					trimmed := strings.Trim(split[2], " ")
					if len(trimmed) == 64 {
						key = trimmed
					}
				}

				badge := canvas.NewImageFromResource(DisplayRating(MenuControl.Contract_rating[key]))
				badge.SetMinSize(fyne.NewSize(35, 35))
				o.(*fyne.Container).Objects[0] = badge
			}
		})

	var item string

	HolderoControl.Table_list.OnSelected = func(id widget.ListItemID) {
		if id != 0 && Connected() {
			go func() {
				item = setHolderoControls(MenuControl.Holdero_tables[id])
				HolderoControl.Favorite_list.UnselectAll()
				HolderoControl.Owned_list.UnselectAll()
			}()
		}
	}

	save_favorite := widget.NewButton("Favorite", func() {
		MenuControl.Holdero_favorites = append(MenuControl.Holdero_favorites, item)
		sort.Strings(MenuControl.Holdero_favorites)
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
		HolderoControl.Table_list)

	return tables_cont
}

// Confrimation for a SCID rating
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

	good := canvas.NewImageFromResource(Resource.B3Badge)
	good.SetMinSize(fyne.NewSize(30, 30))
	bad := canvas.NewImageFromResource(Resource.RBadge)
	bad.SetMinSize(fyne.NewSize(30, 30))

	rate_cont := container.NewBorder(nil, nil, bad, good, slider)

	left := container.NewVBox(confirm)
	right := container.NewVBox(cancel)
	buttons := container.NewAdaptiveGrid(2, left, right)

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	content := container.NewVBox(layout.NewSpacer(), label, rating_label, fee_label, layout.NewSpacer(), rate_cont, layout.NewSpacer(), buttons)

	return container.NewMax(alpha, content)

}

// Favorite Holdero tables object
func HolderoFavorites() fyne.CanvasObject {
	HolderoControl.Favorite_list = widget.NewList(
		func() int {
			return len(MenuControl.Holdero_favorites)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(MenuControl.Holdero_favorites[i])
		})

	var item string

	HolderoControl.Favorite_list.OnSelected = func(id widget.ListItemID) {
		if Connected() {
			item = setHolderoControls(MenuControl.Holdero_favorites[id])
			HolderoControl.Table_list.UnselectAll()
			HolderoControl.Owned_list.UnselectAll()
		}
	}

	remove := widget.NewButton("Remove", func() {
		if len(MenuControl.Holdero_favorites) > 0 {
			HolderoControl.Favorite_list.UnselectAll()
			for i := range MenuControl.Holdero_favorites {
				if MenuControl.Holdero_favorites[i] == item {
					copy(MenuControl.Holdero_favorites[i:], MenuControl.Holdero_favorites[i+1:])
					MenuControl.Holdero_favorites[len(MenuControl.Holdero_favorites)-1] = ""
					MenuControl.Holdero_favorites = MenuControl.Holdero_favorites[:len(MenuControl.Holdero_favorites)-1]
					break
				}
			}
		}
		HolderoControl.Favorite_list.Refresh()
		sort.Strings(MenuControl.Holdero_favorites)
	})

	cont := container.NewBorder(
		nil,
		container.NewBorder(nil, nil, nil, remove, layout.NewSpacer()),
		nil,
		nil,
		HolderoControl.Favorite_list)

	return cont
}

// Owned Holdero tables object
func MyTables() fyne.CanvasObject {
	HolderoControl.Owned_list = widget.NewList(
		func() int {
			return len(MenuControl.Holdero_owned)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(MenuControl.Holdero_owned[i])
		})

	HolderoControl.Owned_list.OnSelected = func(id widget.ListItemID) {
		if Connected() {
			setHolderoControls(MenuControl.Holdero_owned[id])
			HolderoControl.Table_list.UnselectAll()
			HolderoControl.Favorite_list.UnselectAll()
		}
	}

	return HolderoControl.Owned_list
}

// Holdero player name entry
func NameEntry() fyne.CanvasObject {
	MenuControl.Names = widget.NewSelect([]string{}, func(s string) {
		table.Poker_name = s
	})

	MenuControl.Names.PlaceHolder = "Name:"

	return MenuControl.Names
}

// Round a float to precision
func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

type blind struct {
	table.NumericalEntry
}

func (e *blind) TypedKey(k *fyne.KeyEvent) {
	trimmed := strings.Trim(e.Entry.Text, "Biglnd: ")
	switch k.Name {
	case fyne.KeyUp:
		if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
			e.Entry.SetText("Big Blind: " + strconv.FormatFloat(float64(f+0.1), 'f', 1, 64))
		}
	case fyne.KeyDown:
		if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
			if f >= 0.1 {
				e.Entry.SetText("Big Blind: " + strconv.FormatFloat(float64(f-0.1), 'f', 1, 64))
			}
		}
	}
	e.Entry.TypedKey(k)
}

type ante struct {
	table.NumericalEntry
}

func (e *ante) TypedKey(k *fyne.KeyEvent) {
	trimmed := strings.Trim(e.Entry.Text, "Ante: ")
	switch k.Name {
	case fyne.KeyUp:
		if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
			e.Entry.SetText("Ante: " + strconv.FormatFloat(float64(f+0.1), 'f', 1, 64))
		}
	case fyne.KeyDown:
		if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
			if f >= 0.1 {
				e.Entry.SetText("Ante: " + strconv.FormatFloat(float64(f-0.1), 'f', 1, 64))
			}
		}
	}
	e.Entry.TypedKey(k)
}

type cleanAmt struct {
	table.NumericalEntry
}

func (e *cleanAmt) TypedKey(k *fyne.KeyEvent) {
	trimmed := strings.Trim(e.Entry.Text, "Clean: ")
	switch k.Name {
	case fyne.KeyUp:
		if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			e.Entry.SetText("Clean: " + strconv.FormatInt(i+1, 10))
		}
	case fyne.KeyDown:
		if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			if i >= 1 {
				e.Entry.SetText("Clean: " + strconv.FormatInt(i-1, 10))
			}
		}
	}
	e.Entry.TypedKey(k)
}

// Holdero owner control objects, left section
func OwnersBoxLeft(obj []fyne.CanvasObject, tabs *container.AppTabs) fyne.CanvasObject {
	players := []string{"Players", "Close Table", "2 Players", "3 Players", "4 Players", "5 Players", "6 Players"}
	player_select := widget.NewSelect(players, func(s string) {})
	player_select.SetSelectedIndex(0)

	blinds_entry := &blind{}
	blinds_entry.ExtendBaseWidget(blinds_entry)
	blinds_entry.PlaceHolder = "Dero:"
	blinds_entry.SetText("Big Blind: 0.0")
	blinds_entry.Validator = validation.NewRegexp(`^(Big Blind: )\d{1,}\.\d{0,1}`, "Format Not Valid")
	blinds_entry.OnChanged = func(s string) {
		if blinds_entry.Validate() != nil {
			blinds_entry.SetText("Big Blind: 0.0")
			ownerControl.blindAmount = 0
		} else {
			trimmed := strings.Trim(s, "Biglnd: ")
			if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
				if uint64(f*100000)%10000 == 0 {
					blinds_entry.SetText("Big Blind: " + strconv.FormatFloat(roundFloat(f, 1), 'f', 1, 64))
					ownerControl.blindAmount = uint64(roundFloat(f*100000, 1))
				} else {
					blinds_entry.SetText("Big Blind: " + strconv.FormatFloat(roundFloat(f, 1), 'f', 1, 64))
				}
			}
		}
	}

	options := []string{"DERO", "ASSET"}
	ownerControl.chips = widget.NewRadioGroup(options, func(s string) {})
	ownerControl.chips.Horizontal = true

	ante_entry := &ante{}
	ante_entry.ExtendBaseWidget(ante_entry)
	ante_entry.PlaceHolder = "Ante:"
	ante_entry.SetText("Ante: 0.0")
	ante_entry.Validator = validation.NewRegexp(`^(Ante: )\d{1,}\.\d{0,1}`, "Format Not Valid")
	ante_entry.OnChanged = func(s string) {
		if ante_entry.Validate() != nil {
			ante_entry.SetText("Ante: 0.0")
			ownerControl.anteAmount = 0
		} else {
			trimmed := strings.Trim(s, "Ante: ")
			if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
				if uint64(f*100000)%10000 == 0 {
					ante_entry.SetText("Ante: " + strconv.FormatFloat(roundFloat(f, 1), 'f', 1, 64))
					ownerControl.anteAmount = uint64(roundFloat(f*100000, 1))
				} else {
					ante_entry.SetText("Ante: " + strconv.FormatFloat(roundFloat(f, 1), 'f', 1, 64))
				}
			}
		}
	}

	set_button := widget.NewButton("Set Table", func() {
		bb := ownerControl.blindAmount
		sb := ownerControl.blindAmount / 2
		ante := ownerControl.anteAmount
		if table.Poker_name != "" {
			rpc.SetTable(player_select.SelectedIndex(), bb, sb, ante, ownerControl.chips.Selected, table.Poker_name, table.Settings.AvatarUrl)
		}
	})

	clean_entry := &cleanAmt{}
	clean_entry.ExtendBaseWidget(clean_entry)
	clean_entry.PlaceHolder = "Atomic:"
	clean_entry.SetText("Clean: 0")
	clean_entry.Validator = validation.NewRegexp(`^(Clean: )\d{1,}`, "Format Not Valid")
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

	ownerControl.timeout = widget.NewButton("Timeout", func() {
		obj[1] = TimeOutConfirm(obj, tabs)
		obj[1].Refresh()
	})

	force := widget.NewButton("Force Start", func() {
		rpc.ForceStat()
	})

	players_items := container.NewAdaptiveGrid(2, player_select, layout.NewSpacer())
	blind_items := container.NewAdaptiveGrid(2, blinds_entry, ownerControl.chips)
	ante_items := container.NewAdaptiveGrid(2, ante_entry, set_button)
	clean_items := container.NewAdaptiveGrid(2, clean_entry, clean_button)
	time_items := container.NewAdaptiveGrid(2, ownerControl.timeout, force)

	ownerControl.owners_left = container.NewVBox(players_items, blind_items, ante_items, clean_items, time_items)
	ownerControl.owners_left.Hide()

	return ownerControl.owners_left
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

	ownerControl.owners_mid = container.NewAdaptiveGrid(2, kick, pay)
	ownerControl.owners_mid.Hide()

	return ownerControl.owners_mid
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
	Stats.Name = canvas.NewText(" Name: ", color.White)
	Stats.Desc = canvas.NewText(" Description: ", color.White)
	Stats.Version = canvas.NewText(" Table Version: ", color.White)
	Stats.Last = canvas.NewText(" Last Move: ", color.White)
	Stats.Seats = canvas.NewText(" Table Closed ", color.White)

	Stats.Name.TextSize = 18
	Stats.Desc.TextSize = 18
	Stats.Version.TextSize = 18
	Stats.Last.TextSize = 18
	Stats.Seats.TextSize = 18

	HolderoControl.Stats_box = *container.NewVBox(Stats.Name, Stats.Desc, Stats.Version, Stats.Last, Stats.Seats, TableIcon(nil))

	return &HolderoControl.Stats_box
}

// Confirmation of manual Holdero timeout
func TimeOutConfirm(obj []fyne.CanvasObject, reset *container.AppTabs) fyne.CanvasObject {
	var confirm_display = widget.NewLabel("")
	confirm_display.Wrapping = fyne.TextWrapWord

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

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	display := container.NewVScroll(confirm_display)
	options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewBorder(nil, options, nil, nil, display)

	return container.NewMax(alpha, content)
}

// Disable index objects
func disableIndex(d bool) {
	if d {
		table.Assets.Index_button.Hide()
		table.Assets.Index_search.Hide()
		table.Assets.Header_box.Hide()
		Market.Market_box.Hide()
		Gnomes.SCIDS = 0
	} else {
		table.Assets.Index_button.Show()
		table.Assets.Index_search.Show()
		if rpc.Wallet.Connect {
			MenuControl.Claim_button.Show()
			table.Assets.Header_box.Show()
			Market.Market_box.Show()
			if MenuControl.list_open {
				MenuControl.List_button.Hide()
			}
		} else {
			MenuControl.Send_asset.Hide()
			MenuControl.List_button.Hide()
			MenuControl.Claim_button.Hide()
			table.Assets.Header_box.Hide()
			Market.Market_box.Hide()
		}
	}
	table.Assets.Index_button.Refresh()
	table.Assets.Index_search.Refresh()
	table.Assets.Header_box.Refresh()
	Market.Market_box.Refresh()
}

// Disable dPrediction objects
func DisablePreditions(d bool) {
	if d {
		table.Actions.Prediction_box.Hide()
	} else {
		table.Actions.Prediction_box.Show()
	}
	table.Actions.Prediction_box.Refresh()
}

// Disable dSports objects
func disableSports(d bool) {
	if d {
		table.Actions.Sports_box.Hide()
		MenuControl.Sports_check.SetChecked(false)
	}

	table.Actions.Sports_box.Refresh()
}

// Disable actions requiring connection
func disableActions(d bool) {
	if d {
		table.Actions.Dreams.Hide()
		table.Actions.Dero.Hide()
		table.Actions.DEntry.Hide()
		HolderoControl.Holdero_unlock.Hide()
		HolderoControl.Holdero_new.Hide()
		table.Actions.Tournament.Hide()

		if MenuControl.Dapp_list["dSports and dPredictions"] {
			MenuControl.Bet_new_p.Hide()
			MenuControl.Bet_new_s.Hide()
			MenuControl.Bet_unlock_p.Hide()
			MenuControl.Bet_unlock_s.Hide()
			MenuControl.Bet_menu_p.Hide()
			MenuControl.Bet_menu_s.Hide()
			MenuControl.Bet_new_p.Refresh()
			MenuControl.Bet_new_s.Refresh()
			MenuControl.Bet_unlock_p.Refresh()
			MenuControl.Bet_unlock_s.Refresh()
			MenuControl.Bet_menu_p.Refresh()
			MenuControl.Bet_menu_s.Refresh()
		}

		if MenuControl.Dapp_list["Iluma"] {
			table.Iluma.Draw1.Hide()
			table.Iluma.Draw3.Hide()
			table.Iluma.Search.Hide()
			table.Iluma.Draw1.Refresh()
			table.Iluma.Draw3.Refresh()
			table.Iluma.Search.Refresh()
		}
	} else {
		table.Actions.Dreams.Show()
		table.Actions.Dero.Show()
		table.Actions.DEntry.Show()
	}

	table.Actions.Dreams.Refresh()
	table.Actions.DEntry.Refresh()
	table.Actions.Dero.Refresh()
	HolderoControl.Holdero_unlock.Refresh()
	HolderoControl.Holdero_new.Refresh()
	table.Actions.Tournament.Refresh()
}

// Disable Baccarat actions
func disableBaccActions(d bool) {
	if d {
		table.Actions.Bacc_actions.Hide()
	} else {
		table.Actions.Bacc_actions.Show()
	}

	table.Actions.Bacc_actions.Refresh()
}

// Disable owner actions
func disableOwnerControls(d bool) {
	if d {
		ownerControl.owners_left.Hide()
		ownerControl.owners_mid.Hide()
	} else {
		ownerControl.owners_left.Show()
		ownerControl.owners_mid.Show()
	}

	ownerControl.owners_left.Refresh()
	ownerControl.owners_mid.Refresh()
}

// Set objects if bet owner
func SetBetOwner(owner string) {
	if MenuControl.Dapp_list["dSports and dPredictions"] {
		if owner == rpc.Wallet.Address {
			rpc.Wallet.BetOwner = true
			MenuControl.Bet_new_p.Show()
			MenuControl.Bet_new_s.Show()
			MenuControl.Bet_unlock_p.Hide()
			MenuControl.Bet_unlock_s.Hide()
			MenuControl.Bet_menu_p.Show()
			MenuControl.Bet_menu_s.Show()
		} else {
			rpc.Wallet.BetOwner = false
			MenuControl.Bet_new_p.Hide()
			MenuControl.Bet_new_s.Hide()
			MenuControl.Bet_unlock_p.Show()
			MenuControl.Bet_unlock_s.Show()
			MenuControl.Bet_menu_p.Hide()
			MenuControl.Bet_menu_s.Hide()
		}
	}
}

// Confirmation for Holdero contract installs
func HolderoMenuConfirm(c int, obj []fyne.CanvasObject, tabs *container.AppTabs) fyne.CanvasObject {
	var text string
	switch c {
	case 1:
		HolderoControl.Holdero_unlock.Hide()
		text = `You are about to unlock and install your first Holdero Table
		
To help support the project, there is a 3 DERO donation attached to preform this action

Once you've unlocked a table, you can upload as many new tables free of donation

Total transaction will be 3.3 DERO (0.3 gas fee for contract install)

Select a public or private table

	- Public will show up in indexed list of tables

	- Private will not show up in the list

Confirm`
	case 2:
		HolderoControl.Holdero_new.Hide()
		text = `You are about to install a new table

Gas fee to install new table is 0.3 DERO

Select a public or private table

	- Public will show up in indexed list of tables

	- Private will not show up in the list

Confirm`
	}

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	var choice *widget.Select

	confirm_button := widget.NewButton("Confirm", func() {
		if choice.SelectedIndex() < 2 && choice.SelectedIndex() >= 0 {
			rpc.UploadHolderoContract(choice.SelectedIndex())
		}

		if c == 2 {
			HolderoControl.Holdero_new.Show()
		}

		obj[1] = tabs
		obj[1].Refresh()
	})

	options := []string{"Public", "Private"}
	choice = widget.NewSelect(options, func(s string) {
		if s == "Public" || s == "Private" {
			confirm_button.Show()
		} else {
			confirm_button.Hide()
		}
	})

	cancel_button := widget.NewButton("Cancel", func() {
		switch c {
		case 1:
			HolderoControl.Holdero_unlock.Show()
		case 2:
			HolderoControl.Holdero_new.Show()
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

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	content := container.NewBorder(nil, actions, nil, nil, info_box)

	return container.NewMax(alpha, content)
}

// Confirmation for dPrediction contract installs
func BettingMenuConfirmP(c int, obj []fyne.CanvasObject, tabs *container.AppTabs) fyne.CanvasObject {
	var text string
	switch c {
	case 1:
		text = `You are about to unlock and install your first dPrediction contract 
		
To help support the project, there is a 3 DERO donation attached to preform this action

Once you've unlocked dPrediction, you can upload as many new prediction or sports contracts free of donation

Total transaction will be 3.125 DERO (0.125 gas fee for contract install)

Select a public or private contract

	- Public will show up in indexed list of contracts

	- Private will not show up in the list
	
Confirm`
	case 2:
		text = `You are about to install a new dPrediction contract. 

Gas fee to install is 0.125 DERO

Select a public or private contract

	- Public will show up in indexed list of contracts

	- Private will not show up in the list
	
Confirm`
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

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	content := container.NewBorder(nil, actions, nil, nil, info_box)

	return container.NewMax(alpha, content)
}

// Confirmation for dSports contract installs
func BettingMenuConfirmS(c int, obj []fyne.CanvasObject, tabs *container.AppTabs) fyne.CanvasObject {
	var text string
	switch c {
	case 1:
		text = `You are about to unlock and install your first dSports contract
		
To help support the project, there is a 3 DERO donation attached to preform this action

Once you've unlocked dSports, you can upload as many new sports or predictions contracts free of donation

Total transaction will be 3.14 DERO (0.14 gas fee for contract install)

Select a public or private contract

	- Public will show up in indexed list of contracts

	- Private will not show up in the list
	
Confirm`
	case 2:
		text = `You are about to install a new dSports contract

Gas fee to install is 0.14 DERO

Select a public or private contract

	- Public will show up in indexed list of contracts

	- Private will not show up in the list
	
Confirm`
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

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	content := container.NewBorder(nil, actions, nil, nil, info_box)

	return container.NewMax(alpha, content)
}

// Index entry objects
func IndexEntry() fyne.CanvasObject {
	table.Assets.Index_entry = widget.NewMultiLineEntry()
	table.Assets.Index_entry.PlaceHolder = "SCID:"
	table.Assets.Index_button = widget.NewButton("Add to Index", func() {
		s := strings.Split(table.Assets.Index_entry.Text, "\n")
		manualIndex(s)
	})

	table.Assets.Index_search = widget.NewButton("Search Index", func() {
		searchIndex(table.Assets.Index_entry.Text)
	})

	MenuControl.Send_asset = widget.NewButton("Send Asset", func() {
		go sendAssetMenu()
	})

	MenuControl.List_button = widget.NewButton("List Asset", func() {
		listMenu()
	})

	MenuControl.Claim_button = widget.NewButton("Claim NFA", func() {
		if len(table.Assets.Index_entry.Text) == 64 {
			if isNfa(table.Assets.Index_entry.Text) {
				rpc.ClaimNfa(table.Assets.Index_entry.Text)
			}
		}
	})

	table.Assets.Index_button.Hide()
	table.Assets.Index_search.Hide()
	MenuControl.List_button.Hide()
	MenuControl.Claim_button.Hide()
	MenuControl.Send_asset.Hide()

	table.Assets.Gnomes_index = canvas.NewText(" Indexed SCIDs: ", color.White)
	table.Assets.Gnomes_index.TextSize = 18

	bottom_grid := container.NewAdaptiveGrid(3, table.Assets.Gnomes_index, table.Assets.Index_button, table.Assets.Index_search)
	top_grid := container.NewAdaptiveGrid(3, container.NewMax(MenuControl.Send_asset), MenuControl.Claim_button, MenuControl.List_button)
	box := container.NewVBox(top_grid, layout.NewSpacer(), bottom_grid)

	cont := container.NewAdaptiveGrid(2, table.Assets.Index_entry, box)

	return cont
}

// Owned asset list object
func AssetList() fyne.CanvasObject {
	table.Assets.Asset_list = widget.NewList(
		func() int {
			return len(table.Assets.Assets)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(table.Assets.Assets[i])
		})

	table.Assets.Asset_list.OnSelected = func(id widget.ListItemID) {
		split := strings.Split(table.Assets.Assets[id], "   ")
		if len(split) >= 2 {
			trimmed := strings.Trim(split[1], " ")
			MenuControl.Viewing_asset = trimmed
			table.Assets.Icon = *canvas.NewImageFromImage(nil)
			go GetOwnedAssetStats(trimmed)
		}
	}

	box := container.NewMax(table.Assets.Asset_list)

	return box
}

// Send Dero asset menu
func sendAssetMenu() {
	MenuControl.send_open = true
	saw := fyne.CurrentApp().NewWindow("Send Asset")
	saw.Resize(fyne.NewSize(330, 700))
	saw.SetIcon(Resource.SmallIcon)
	MenuControl.Send_asset.Hide()
	MenuControl.List_button.Hide()
	saw.SetCloseIntercept(func() {
		MenuControl.send_open = false
		if rpc.Wallet.Connect {
			MenuControl.Send_asset.Show()
			if isNfa(MenuControl.Viewing_asset) {
				MenuControl.List_button.Show()
			}
		}
		saw.Close()
	})
	saw.SetFixedSize(true)

	var saw_content *fyne.Container
	var send_button *widget.Button
	img := *canvas.NewImageFromResource(Resource.Back3)
	alpha := canvas.NewRectangle(color.RGBA{0, 0, 0, 180})

	viewing_asset := MenuControl.Viewing_asset

	viewing_label := widget.NewLabel(fmt.Sprintf("Sending SCID:\n%s\n\nEnter destination address below.\n\nSCID can be sent to reciever as payload.\n\n", viewing_asset))
	viewing_label.Wrapping = fyne.TextWrapWord
	viewing_label.Alignment = fyne.TextAlignCenter

	info_label := widget.NewLabel("Enter all info before sending")
	payload := widget.NewCheck("Send SCID as payload", func(b bool) {})

	dest_entry := widget.NewMultiLineEntry()
	dest_entry.SetPlaceHolder("Destination Address:")
	dest_entry.Wrapping = fyne.TextWrapWord
	dest_entry.Validator = validation.NewRegexp(`^(dero)\w{62}`, "Invalid Address")
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
						&img,
						alpha,
						saw_content))
			})

			dest = dest_entry.Text
			confirm_label := widget.NewLabel(fmt.Sprintf("Sending SCID:\n%s\n\nDestination: %s\n\nSending SCID as payload: %t", send_asset, dest, load))
			confirm_label.Wrapping = fyne.TextWrapWord
			confirm_label.Alignment = fyne.TextAlignCenter

			confirm_display := container.NewVBox(confirm_label, layout.NewSpacer())
			confirm_options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
			confirm_content := container.NewBorder(nil, confirm_options, nil, nil, confirm_display)
			saw.SetContent(
				container.New(layout.NewMaxLayout(),
					&img,
					alpha,
					confirm_content))
		}
	})
	send_button.Hide()

	icon := table.Assets.Icon

	saw_content = container.NewVBox(
		viewing_label,
		menuAssetImg(&icon, Resource.Frame),
		layout.NewSpacer(),
		dest_entry,
		container.NewCenter(payload),
		layout.NewSpacer(),
		container.NewAdaptiveGrid(2, layout.NewSpacer(), send_button))

	go func() {
		for rpc.Wallet.Connect && rpc.Signal.Daemon {
			time.Sleep(3 * time.Second)
			if !confirm_open {
				icon = table.Assets.Icon
				saw_content.Objects[1] = menuAssetImg(&icon, Resource.Frame)
				if viewing_asset != MenuControl.Viewing_asset {
					viewing_asset = MenuControl.Viewing_asset
					viewing_label.SetText("Sending SCID:\n" + viewing_asset + " \n\nEnter destination address below.\n\nSCID can be sent to reciever as payload.\n\n")
				}
				saw_content.Refresh()
			}
		}
		MenuControl.send_open = false
		saw.Close()
	}()

	saw.SetContent(
		container.New(layout.NewMaxLayout(),
			&img,
			alpha,
			saw_content))
	saw.Show()
}

// Image for send asset and list menus
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
func listMenu() {
	MenuControl.list_open = true
	aw := fyne.CurrentApp().NewWindow("List NFA")
	aw.Resize(fyne.NewSize(330, 700))
	aw.SetIcon(Resource.SmallIcon)
	MenuControl.List_button.Hide()
	MenuControl.Send_asset.Hide()
	aw.SetCloseIntercept(func() {
		MenuControl.list_open = false
		if rpc.Wallet.Connect {
			MenuControl.Send_asset.Show()
			if isNfa(MenuControl.Viewing_asset) {
				MenuControl.List_button.Show()
			}
		}
		aw.Close()
	})
	aw.SetFixedSize(true)

	var aw_content *fyne.Container
	var set_list *widget.Button
	aw_img := *canvas.NewImageFromResource(Resource.Back3)
	alpha := canvas.NewRectangle(color.RGBA{0, 0, 0, 180})

	viewing_asset := MenuControl.Viewing_asset
	viewing_label := widget.NewLabel(fmt.Sprintf("Listing SCID: %s", viewing_asset))
	viewing_label.Wrapping = fyne.TextWrapWord
	viewing_label.Alignment = fyne.TextAlignCenter

	fee_label := widget.NewLabel("Listing fee 0.1 Dero")

	listing_options := []string{"Auction", "Sale"}
	listing := widget.NewSelect(listing_options, func(s string) {})
	listing.PlaceHolder = "Type:"

	duration := table.NilNumericalEntry()
	duration.SetPlaceHolder("Duration in Hours:")
	duration.Validator = validation.NewRegexp(`^[^0]\d{0,2}$`, "Int required")

	start := table.NilNumericalEntry()
	start.SetPlaceHolder("Start Price:")
	start.Validator = validation.NewRegexp(`\d{1,}\.\d{1,5}$`, "Float required")

	charAddr := widget.NewEntry()
	charAddr.SetPlaceHolder("Charity Donation Address:")
	charAddr.Validator = validation.NewRegexp(`^\w{66,66}$`, "Int required")

	charPerc := table.NilNumericalEntry()
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

				first_line := fmt.Sprintf("Listing SCID:\n%s\n\nList Type: %s\n\nDuration: %s Hours\n\nStart Price: %0.5f Dero\n\n", listing_asset, listing.Selected, duration.Text, sp)
				second_line := fmt.Sprintf("Artificer Fee: %.0f%s - %0.5f Dero\n\nRoyalties: %.0f%s - %0.5f Dero\n\n", artP*100, "%", art_gets, royaltyP*100, "%", royalty_gets)
				third_line := fmt.Sprintf("Chairity Address: %s\n\nCharity Percent: %s%s - %0.5f Dero\n\nYou will receive %.5f Dero if asset sells at start price", charAddr.Text, charPerc.Text, "%", char_gets, total)

				confirm_label := widget.NewLabel(first_line + second_line + third_line)
				confirm_label.Wrapping = fyne.TextWrapWord
				confirm_label.Alignment = fyne.TextAlignCenter

				cancel_button := widget.NewButton("Cancel", func() {
					confirm_open = false
					aw.SetContent(
						container.New(layout.NewMaxLayout(),
							&aw_img,
							alpha,
							aw_content))
				})

				confirm_button := widget.NewButton("Confirm", func() {
					rpc.NfaSetListing(listing_asset, listing.Selected, charAddr.Text, d, s, cp)
					MenuControl.list_open = false
					if rpc.Wallet.Connect {
						MenuControl.Send_asset.Show()
						if isNfa(MenuControl.Viewing_asset) {
							MenuControl.List_button.Show()
						}
					}
					aw.Close()
				})

				confirm_display := container.NewVBox(confirm_label, layout.NewSpacer())
				confirm_options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
				confirm_content := container.NewBorder(nil, confirm_options, nil, nil, confirm_display)

				aw.SetContent(
					container.New(layout.NewMaxLayout(),
						&aw_img,
						alpha,
						confirm_content))
			}
		}
	})
	set_list.Hide()

	icon := table.Assets.Icon

	go func() {
		for rpc.Wallet.Connect && rpc.Signal.Daemon {
			time.Sleep(3 * time.Second)
			if !confirm_open {
				icon = table.Assets.Icon
				aw_content.Objects[2] = menuAssetImg(&icon, Resource.Frame)
				if viewing_asset != MenuControl.Viewing_asset {
					viewing_asset = MenuControl.Viewing_asset
					viewing_label.SetText(fmt.Sprintf("Listing SCID: %s\n", viewing_asset))
				}
				aw_content.Refresh()
			}
		}
		MenuControl.list_open = false
		aw.Close()
	}()

	aw_content = container.NewVBox(
		viewing_label,
		layout.NewSpacer(),
		menuAssetImg(&icon, Resource.Frame),
		layout.NewSpacer(),
		layout.NewSpacer(),
		listing,
		duration,
		start,
		charAddr,
		charPerc,
		container.NewAdaptiveGrid(2, layout.NewSpacer(), container.NewCenter(fee_label)),
		container.NewAdaptiveGrid(2, layout.NewSpacer(), set_list))

	aw.SetContent(
		container.New(layout.NewMaxLayout(),
			&aw_img,
			alpha,
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
		"":                      {"Welcome to dReams"},
		"Welcome to dReams":     {"Get Started", "Contracts", "Assets", "Market"},
		"Get Started":           {"You will need a Dero wallet to play, visit dero.io for more info", "Can use local daemon, or remote daemon options are availible in drop down", "Enter daemon rpc address, wallet rpc address and user:pass", "Press connect, D & W indicators at top right of screen will light up on successful connection", "On first start start up of app, Gnomon will take ~10 seconds to create your local db", "Gnomon idicator will have a stripe when starting or syncing, indicator will turn solid when startup, sync and scan are completed"},
		"Contracts":             {"Holdero", "Baccarat", "Predictions", "Sports", "dReam Service", "Tarot", "Contract Ratings"},
		"Holdero":               {"Multiplayer Texas Hold'em style on chian poker", "No limit, single raise game. Table owners choose game params", "Six players max at a table", "No side pots, must call or fold", "Public and private tables can use Dero or dReam Tokens", "dReam Tools", "Tournament tables can be set up to use any Token", "View table listings or launch your own Holdero contract from the contracts tab"},
		"dReam Tools":           {"A suite of tools for Holdero, unlocked with ownership of AZY or SIX playing card assets (Requires one deck or two backs)", "Odds calculator", "Bot player with 12 customizable parameters", "Track playing stats for users and bot players"},
		"Baccarat":              {"A popular table game, where closest to 9 wins", "Uses dReam Tokens for betting"},
		"Predictions":           {"Prediction contracts are for binary based predictions, (higher/lower, yes/no)", "How predictions works", "Current Markets", "dReams Client aggregated price feed", "View active prediction contracts in predictions tab or launch your own prediction contract from the contracts tab"},
		"How predictions works": {"P2P predictions", "Variable time limits allowing for different prediction set ups, each contract runs one prediction at a time", "Click a contract from the list to view it", "Closes at, is when the contract will stop accepting predictions", "Mark (price or value you are predicting on) can be set on prediction initialization or it can given live", "Posted with in, is the acceptable time frame to post the live Mark", "If Mark is not posted, prediction is voided and you will be refunded", "Payout after, is when the Final price is posted and compared to the mark to determine winners", "If the final price is not posted with in refund time frame, prediction is void and you will be refunded"},
		"Current Markets":       {"DERO-BTC", "XMR-BTC", "BTC-USDT", "DERO-USDT", "XMR-USDT", "DERO-Difficulty", "DERO-Block Time", "DERO-Block Number"},
		"Sports":                {"Sports contracts are for sports wagers", "How sports works", "Current Leagues", "Live game scores, and game schedules", "View active sports contracts in sports tab or launch your own sports contract from the contracts tab"},
		"How sports works":      {"P2P betting", "Variable time limits, one contract can run miltiple games at the same time", "Click a contract from the list to view it", "Any active games on the contract will populate, you can pick which game you'd like to play from the drop down", "Closes at, is when the contrcts stops accepting picks", "Default payout time after close is 4hr, this is when winner will be posted from client feed", "Default refund time is 8hr after close, meaning if winner is not provided past that time you will be refunded", "A Tie refunds pot to all all participants"},
		"Current Leagues":       {"EPL", "NBA", "NFL", "NHL", "Bellator", "UFC"},
		"dReam Service":         {"dReam Service is unlocked for all betting contract owners", "Full automation of contract posts and payouts", "Integrated address service allows bets to be placed thorugh a Dero transaction to sent to service", "Multiple owners can be added to contracts and multiple service wallets can be ran on one contract"},
		"Tarot":                 {"On chian Tarot readings", "Iluma cards and readings created by Kalina Lux"},
		"Contract Ratings":      {"Holdero and public betting contracts each have a rating stored on chain", "Players can rate other contracts positively or negatively", "Four rating tiers, tier two being the starting tier for all contracts", "Each rating transaction is weight based by its Dero value", "Contracts that fall below tier one will no longer populate in the public index"},
		"Assets":                {"View any owned assets held in wallet", "Put owned assets up for auction or for sale", "Send assets privately to another wallet", "Indexer, add custom contracts to your index and search current index db"},
		"Market":                {"View any in game assets up for auction or sale", "Bid on or buy assets", "Cancel or close out any existing listings"},
	}

	tree := widget.NewTreeWithStrings(list)

	tree.OnBranchClosed = func(uid widget.TreeNodeID) {
		tree.UnselectAll()
	}

	tree.OnBranchOpened = func(uid widget.TreeNodeID) {
		tree.Select(uid)
	}

	tree.OpenBranch("Welcome to dReams")

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	max := container.NewMax(alpha, tree)

	return max
}

// Send Dero message menu
func SendMessageMenu() {
	if !MenuControl.msg_open {
		MenuControl.msg_open = true
		smw := fyne.CurrentApp().NewWindow("Send Asset")
		smw.Resize(fyne.NewSize(330, 700))
		smw.SetIcon(Resource.SmallIcon)
		smw.SetCloseIntercept(func() {
			MenuControl.msg_open = false
			smw.Close()
		})
		smw.SetFixedSize(true)

		var send_button *widget.Button
		img := *canvas.NewImageFromResource(Resource.Back3)
		alpha := canvas.NewRectangle(color.RGBA{0, 0, 0, 180})

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
		dest_entry.Validator = validation.NewRegexp(`^(dero)\w{62}`, "Invalid Address")
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
				MenuControl.msg_open = false
				smw.Close()
			}
		})
		send_button.Hide()

		dest_cont := container.NewVBox(label, ringsize, dest_entry)
		message_cont := container.NewBorder(nil, send_button, nil, nil, message_entry)

		content := container.NewVSplit(dest_cont, message_cont)

		go func() {
			for rpc.Wallet.Connect && rpc.Signal.Daemon {
				time.Sleep(3 * time.Second)
			}
			MenuControl.msg_open = false
			smw.Close()
		}()

		smw.SetContent(
			container.New(layout.NewMaxLayout(),
				&img,
				alpha,
				content))
		smw.Show()
	}
}
