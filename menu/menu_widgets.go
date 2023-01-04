package menu

import (
	"image/color"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"

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
	// DAEMON_RPC_REMOTE2 = "dero-node.mysrv.cloud:10102"
	// DAEMON_RPC_REMOTE3 = "derostats.io:10102"
)

type menuOptions struct {
	list_open         bool
	Daemon_config     string
	Viewing_asset     string
	Holdero_tables    []string
	Holdero_favorites []string
	Holdero_owned     []string
	Predict_contracts []string
	Predict_favorites []string
	Predict_owned     []string
	Sports_contracts  []string
	Sports_favorites  []string
	Sports_owned      []string
	Bet_unlock        *widget.Button
	Bet_new           *widget.Button
	Bet_menu          *widget.Button
	Claim_button      *widget.Button
	List_button       *widget.Button
	Set_list          *widget.Button
	daemon_check      *widget.Check
	holdero_check     *widget.Check
	Predict_check     *widget.Check
	Sports_check      *widget.Check
	Wallet_ind        *fyne.Animation
	Daemon_ind        *fyne.Animation
}

type holderoOptions struct {
	contract_input *widget.SelectEntry
	Table_list     *widget.List
	Favorite_list  *widget.List
	Owned_list     *widget.List
	holdero_unlock *widget.Button
	holdero_new    *widget.Button
	Stats_box      fyne.Container
}

type resources struct {
	SmallIcon fyne.Resource
	Frame     fyne.Resource
	Back1     fyne.Resource
	Back2     fyne.Resource
	Back3     fyne.Resource
	Back4     fyne.Resource
	Gnomon    fyne.Resource
}

var Resource resources
var HolderoControl holderoOptions
var MenuControl menuOptions

func GetMenuResources(r1, r2, r3, r4, r5, r6, r7 fyne.Resource) {
	Resource.SmallIcon = r1
	Resource.Frame = r2
	Resource.Back1 = r3
	Resource.Back2 = r4
	Resource.Back3 = r5
	Resource.Back4 = r6
	Resource.Gnomon = r7
}

func disconnected() {
	rpc.Wallet.PokerOwner = false
	rpc.Wallet.BetOwner = false
	rpc.Round.ID = 0
	rpc.Display.PlayerId = ""
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
	table.Actions.NameEntry.Text = ""
	table.Actions.NameEntry.Enable()
	table.Actions.NameEntry.Refresh()
	Market.Auction_list.UnselectAll()
	Market.Buy_list.UnselectAll()
	Market.Icon = *canvas.NewImageFromImage(nil)
	Market.Cover = *canvas.NewImageFromImage(nil)
	Market.Viewing = ""
	ResetAuctionInfo()
	AuctionInfo()
}

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

func CheckConnection() {
	if rpc.Signal.Daemon {
		MenuControl.daemon_check.SetChecked(true)
		disableIndex(false)
	} else {
		MenuControl.daemon_check.SetChecked(false)
		MenuControl.holdero_check.SetChecked(false)
		MenuControl.Predict_check.SetChecked(false)
		MenuControl.Sports_check.SetChecked(false)
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
		MenuControl.Predict_check.SetChecked(false)
		MenuControl.Sports_check.SetChecked(false)
		rpc.Signal.Contract = false
		clearContractLists()
		disableOwnerControls(true)
		disableBaccActions(true)
		DisablePreditions(true)
		disableSports(true)
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

func DaemonConnectedBox() fyne.Widget {
	MenuControl.daemon_check = widget.NewCheck("", func(b bool) {
		if !Gnomes.Init && !Gnomes.Start {
			startGnomon(rpc.Round.Daemon)
			HolderoControl.contract_input.CursorColumn = 1
			HolderoControl.contract_input.Refresh()
			table.Actions.P_contract.CursorColumn = 1
			table.Actions.P_contract.Refresh()
			table.Actions.S_contract.CursorColumn = 1
			table.Actions.S_contract.Refresh()
		}

		if !b {
			StopGnomon(Gnomes.Init)
		}
	})
	MenuControl.daemon_check.Disable()
	MenuControl.daemon_check.Hide()

	return MenuControl.daemon_check
}

func HolderoContractConnectedBox() fyne.Widget {
	MenuControl.holdero_check = widget.NewCheck("", func(b bool) {
		if !b {
			disableOwnerControls(true)
		}
	})
	MenuControl.holdero_check.Disable()

	return MenuControl.holdero_check
}

func DaemonRpcEntry() fyne.Widget {
	var options = []string{"", DAEMON_RPC_DEFAULT, DAEMON_RPC_REMOTE1, DAEMON_RPC_REMOTE2}
	if MenuControl.Daemon_config != "" {
		options = append(options, MenuControl.Daemon_config)
	}
	entry := widget.NewSelectEntry(options)
	entry.PlaceHolder = "Daemon RPC: "

	this := binding.BindString(&rpc.Round.Daemon)
	entry.Bind(this)

	return entry
}

func WalletRpcEntry() fyne.Widget {
	options := []string{"", "127.0.0.1:10103"}
	entry := widget.NewSelectEntry(options)
	entry.PlaceHolder = "Wallet RPC: "
	entry.OnCursorChanged = func() {
		if rpc.Wallet.Connect {
			rpc.Wallet.Address = ""
			rpc.Wallet.Height = "0"
			rpc.Wallet.Connect = false
			CheckConnection()
		}
	}

	this := binding.BindString(&rpc.Wallet.Rpc)
	entry.Bind(this)

	return entry
}

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

func HolderoContractEntry() fyne.Widget {
	HolderoControl.contract_input = widget.NewSelectEntry(nil)
	options := []string{""}
	HolderoControl.contract_input.SetOptions(options)
	HolderoControl.contract_input.PlaceHolder = "Holdero Contract Address: "
	HolderoControl.contract_input.OnCursorChanged = func() {
		HolderoControl.contract_input.Validate()
		if rpc.Signal.Daemon {
			text := HolderoControl.contract_input.Text
			table.ClearShared()
			if len(text) == 64 {
				go rpc.CheckHolderoContract()
				if checkTableOwner(text) {
					disableOwnerControls(false)
					if checkTableVersion(text) >= 110 {
						ownerControl.chips.Show()
						ownerControl.owners_mid.Show()
					} else {
						ownerControl.chips.Hide()
						ownerControl.owners_mid.Hide()
					}
				} else {
					disableOwnerControls(true)
				}

				tourney, _ := rpc.CheckTournamentTable()
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
		}
	}

	this := binding.BindString(&rpc.Round.Contract)
	HolderoControl.contract_input.Bind(this)

	return HolderoControl.contract_input
}

func RpcConnectButton() fyne.Widget {
	button := widget.NewButton("Connect", func() {
		go func() {
			rpc.Ping()
			rpc.GetAddress()
			CheckConnection()
			HolderoControl.contract_input.CursorColumn = 1
			HolderoControl.contract_input.Refresh()
			table.Actions.P_contract.CursorColumn = 1
			table.Actions.P_contract.Refresh()
			table.Actions.S_contract.CursorColumn = 1
			table.Actions.S_contract.Refresh()
		}()
	})

	return button
}

func HolderoUnlockButton() fyne.Widget { /// unlock table and upload
	HolderoControl.holdero_unlock = widget.NewButton("Unlock Holdero Contract", func() {
		holderoMenuConfirm(1)
	})

	HolderoControl.holdero_unlock.Hide()

	return HolderoControl.holdero_unlock
}

func NewTableButton() fyne.Widget {
	HolderoControl.holdero_new = widget.NewButton("New Holdero Table", func() {
		holderoMenuConfirm(2)
	})

	HolderoControl.holdero_new.Hide()

	return HolderoControl.holdero_new
}

func BettingUnlockButton() fyne.Widget { /// unlock betting and upload
	MenuControl.Bet_unlock = widget.NewButton("Unlock Betting Contracts", func() {
		bettingMenuConfirm(1)
	})

	MenuControl.Bet_unlock.Hide()

	return MenuControl.Bet_unlock
}

func NewBettingButton() fyne.Widget {
	MenuControl.Bet_new = widget.NewButton("New Betting Contract", func() {
		bettingMenuConfirm(2)
	})

	MenuControl.Bet_new.Hide()

	return MenuControl.Bet_new
}

func setHolderoControls(str string) (item string) {
	split := strings.Split(str, "   ")
	if len(split) >= 3 {
		trimmed := strings.Trim(split[2], " ")
		if len(trimmed) == 64 {
			item = str
			HolderoControl.contract_input.SetText(trimmed)
			go GetTableStats(trimmed, true)
			rpc.Times.Kick_block = rpc.StringToInt(rpc.Wallet.Height)
		}
	}

	return
}

func TableListings() fyne.CanvasObject { /// tables contracts
	HolderoControl.Table_list = widget.NewList(
		func() int {
			return len(MenuControl.Holdero_tables)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(MenuControl.Holdero_tables[i])
		})

	var item string

	HolderoControl.Table_list.OnSelected = func(id widget.ListItemID) {
		if id != 0 && Connected() {
			item = setHolderoControls(MenuControl.Holdero_tables[id])
			HolderoControl.Favorite_list.UnselectAll()
			HolderoControl.Owned_list.UnselectAll()
		}
	}

	save := widget.NewButton("Favorite", func() {
		MenuControl.Holdero_favorites = append(MenuControl.Holdero_favorites, item)
		sort.Strings(MenuControl.Holdero_favorites)
	})

	cont := container.NewBorder(
		nil,
		container.NewBorder(nil, nil, nil, save, layout.NewSpacer()),
		nil,
		nil,
		HolderoControl.Table_list)

	return cont
}

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

func NameEntry() fyne.CanvasObject {
	name := widget.NewEntry()
	name.PlaceHolder = "Name:"
	this := binding.BindString(&table.Poker_name)
	name.Bind(this)
	name.Validator = validation.NewRegexp(`^.{3,12}$`, "Format Not Valid")

	return name
}

// / Owner
type tableOwnerOptions struct {
	blindAmount uint64
	anteAmount  uint64
	chips       *widget.RadioGroup
	owners_left *fyne.Container
	owners_mid  *fyne.Container
}

var ownerControl tableOwnerOptions

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

func OwnersBoxLeft() fyne.CanvasObject {
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
			log.Println("Invalid Clean Amount")
		}
	})

	timeout := widget.NewButton("Timeout", func() {
		TimeOutConfirm()
	})

	force := widget.NewButton("Force Start", func() {
		rpc.ForceStat()
	})

	players_items := container.NewAdaptiveGrid(2, player_select, layout.NewSpacer())
	blind_items := container.NewAdaptiveGrid(2, blinds_entry, ownerControl.chips)
	ante_items := container.NewAdaptiveGrid(2, ante_entry, set_button)
	clean_items := container.NewAdaptiveGrid(2, clean_entry, clean_button)
	time_items := container.NewAdaptiveGrid(2, timeout, force)

	ownerControl.owners_left = container.NewVBox(players_items, blind_items, ante_items, clean_items, time_items)
	ownerControl.owners_left.Hide()

	return ownerControl.owners_left
}

func OwnersBoxMid() fyne.CanvasObject {
	kick_label := widget.NewLabel("      Auto Kick after")
	k_times := []string{"Off", "2m", "1m"}
	auto_remove := widget.NewSelect(k_times, func(s string) {
		switch s {
		case "Off":
			rpc.Times.Kick = 0
		case "2m":
			rpc.Times.Kick = 120
		case "1m":
			rpc.Times.Kick = 60
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

func TimeOutConfirm() {
	tocw := fyne.CurrentApp().NewWindow("Confirm")
	tocw.SetIcon(Resource.SmallIcon)
	tocw.Resize(fyne.NewSize(330, 150))
	tocw.SetFixedSize(true)
	var confirm_display = widget.NewLabel("")
	confirm_display.Wrapping = fyne.TextWrapWord

	confirm_display.SetText("Confirm Time Out on Current Player")

	cancel_button := widget.NewButton("Cancel", func() {
		tocw.Close()
	})
	confirm_button := widget.NewButton("Confirm", func() {
		rpc.TimeOut()
		tocw.Close()
	})

	display := container.NewVBox(confirm_display, layout.NewSpacer())
	options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewVBox(display, layout.NewSpacer(), options)

	img := *canvas.NewImageFromResource(Resource.Back1)
	tocw.SetContent(
		container.New(layout.NewMaxLayout(),
			&img,
			content))
	tocw.Show()
}

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

func DisablePreditions(d bool) {
	if d {
		table.Actions.Prediction_box.Hide()
	} else {
		table.Actions.Prediction_box.Show()
	}
	table.Actions.Prediction_box.Refresh()
}

func disableSports(d bool) {
	if d {
		table.Actions.Sports_box.Hide()
		MenuControl.Sports_check.SetChecked(false)
	}

	table.Actions.Sports_box.Refresh()
}

func disableActions(d bool) {
	if d {
		table.Actions.Dreams.Hide()
		table.Actions.Dero.Hide()
		table.Actions.DEntry.Hide()
		HolderoControl.holdero_unlock.Hide()
		HolderoControl.holdero_new.Hide()
		MenuControl.Bet_new.Hide()
		MenuControl.Bet_unlock.Hide()
		MenuControl.Bet_menu.Hide()
		table.Actions.Tournament.Hide()
		table.Iluma.Draw1.Hide()
		table.Iluma.Draw3.Hide()
	} else {
		table.Actions.Dreams.Show()
		table.Actions.Dero.Show()
		table.Actions.DEntry.Show()
	}

	table.Actions.Dreams.Refresh()
	table.Actions.DEntry.Refresh()
	table.Actions.Dero.Refresh()
	HolderoControl.holdero_unlock.Refresh()
	HolderoControl.holdero_new.Refresh()
	MenuControl.Bet_new.Refresh()
	MenuControl.Bet_unlock.Refresh()
	MenuControl.Bet_menu.Refresh()
	table.Actions.Tournament.Refresh()
	table.Iluma.Draw1.Refresh()
	table.Iluma.Draw3.Refresh()
}

func disableBaccActions(d bool) {
	if d {
		table.Actions.Bacc_actions.Hide()
	} else {
		table.Actions.Bacc_actions.Show()
	}

	table.Actions.Bacc_actions.Refresh()
}

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

func DisableBetOwner(o string) {

	if o == rpc.Wallet.Address {
		rpc.Wallet.BetOwner = true
		MenuControl.Bet_new.Show()
		MenuControl.Bet_unlock.Hide()
		MenuControl.Bet_menu.Show()
	} else {
		rpc.Wallet.BetOwner = false
		MenuControl.Bet_new.Hide()
		MenuControl.Bet_unlock.Show()
		MenuControl.Bet_menu.Hide()
	}
}

func holderoMenuConfirm(c int) {
	var text string
	switch c {
	case 1:
		text = `You are about to unlock and install your first Holdero Table. 
		
To help support the project, there is a 3 DERO donation attached to preform this action.

Once you've unlocked a table, you can upload as many new tables free of donation

Total transaction will be 3.3 DERO (0.3 gas fee for contract install)

Select a public or private table.

	- Public will show up in indexed list of tables
	- Private will not show up in the list


Confirm to proceed with unlock and install.`
	case 2:
		text = `You are about to install a new table. 

Gas fee to install new table is 0.3 DERO.

Select a public or private table.

	- Public will show up in indexed list of tables
	- Private will not show up in the list


Confirm to proceed with install.`
	}

	cw := fyne.CurrentApp().NewWindow("Confirm")
	cw.Resize(fyne.NewSize(450, 450))
	cw.SetFixedSize(true)
	cw.SetIcon(Resource.SmallIcon)
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord

	var choice *widget.Select

	confirm_button := widget.NewButton("Confirm", func() {
		if choice.SelectedIndex() < 2 && choice.SelectedIndex() >= 0 {
			rpc.UploadHolderoContract(rpc.Signal.Daemon, rpc.Wallet.Connect, choice.SelectedIndex())
			cw.Close()
		}
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
		cw.Close()
	})

	confirm_button.Hide()

	left := container.NewVBox(confirm_button)
	right := container.NewVBox(cancel_button)
	buttons := container.NewAdaptiveGrid(2, left, right)
	box := container.NewVBox(choice, buttons)
	scroll := container.NewVScroll(label)

	content := container.NewBorder(nil, box, nil, nil, scroll)

	img := *canvas.NewImageFromResource(Resource.Back4)
	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	cw.SetContent(
		container.New(layout.NewMaxLayout(),
			&img,
			alpha,
			content))
	cw.Show()
}

func bettingMenuConfirm(c int) {
	var text string
	switch c {
	case 1:
		text = `You are about to unlock and install your first betting contract. 
		
To help support the project, there is a 3 DERO donation attached to preform this action.

Once you've unlocked betting, you can upload as many new betting contracts free of donation

Total transaction will be 3.1 DERO (0.1 gas fee for contract install)

Select a public or private contract.

	- Public will show up in indexed list of contracrs
	- Private will not show up in the list


Choose Predictions or Sports to proceed with unlock and installing chosen contract.`
	case 2:
		text = `You are about to install a new betting contract. 

Gas fee to install new betting contract is 0.1 DERO.

Select a public or private contract.

	- Public will show up in indexed list of contracrs
	- Private will not show up in the list

Choose Predictions or Sports to proceed with install.`
	}

	cw := fyne.CurrentApp().NewWindow("Confirm")
	cw.Resize(fyne.NewSize(450, 450))
	cw.SetFixedSize(true)
	cw.SetIcon(Resource.SmallIcon)
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord

	var choice *widget.Select

	pre_button := widget.NewButton("Predictions", func() {
		if choice.SelectedIndex() < 2 && choice.SelectedIndex() >= 0 {
			rpc.UploadBetContract(rpc.Signal.Daemon, rpc.Wallet.Connect, true, choice.SelectedIndex())
			cw.Close()
		}
	})

	sports_button := widget.NewButton("Sports", func() {
		if choice.SelectedIndex() < 2 && choice.SelectedIndex() >= 0 {
			rpc.UploadBetContract(rpc.Signal.Daemon, rpc.Wallet.Connect, false, choice.SelectedIndex())
			cw.Close()
		}
	})

	options := []string{"Public", "Private"}
	choice = widget.NewSelect(options, func(s string) {
		if s == "Public" || s == "Private" {
			pre_button.Show()
			sports_button.Show()
		} else {
			pre_button.Hide()
			sports_button.Hide()
		}
	})

	cancel_button := widget.NewButton("Cancel", func() {
		cw.Close()
	})

	pre_button.Hide()
	sports_button.Hide()

	left := container.NewVBox(pre_button)
	mid := container.NewVBox(sports_button)
	right := container.NewVBox(cancel_button)
	buttons := container.NewAdaptiveGrid(3, left, mid, right)
	box := container.NewVBox(choice, buttons)
	scroll := container.NewVScroll(label)

	content := container.NewBorder(nil, box, nil, nil, scroll)

	img := *canvas.NewImageFromResource(Resource.Back4)
	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	cw.SetContent(
		container.New(layout.NewMaxLayout(),
			&img,
			alpha,
			content))
	cw.Show()
}

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

	table.Assets.Gnomes_index = canvas.NewText(" Indexed SCIDs: ", color.White)
	table.Assets.Gnomes_index.TextSize = 18

	bottom_grid := container.NewAdaptiveGrid(3, table.Assets.Gnomes_index, table.Assets.Index_button, table.Assets.Index_search)
	top_grid := container.NewAdaptiveGrid(3, layout.NewSpacer(), MenuControl.Claim_button, MenuControl.List_button)
	box := container.NewVBox(top_grid, layout.NewSpacer(), bottom_grid)

	cont := container.NewAdaptiveGrid(2, table.Assets.Index_entry, box)

	return cont
}

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
			go GetOwnedAssetStats(trimmed)
		}
	}

	box := container.NewMax(table.Assets.Asset_list)

	return box
}

func listMenu() { /// asset listing
	MenuControl.list_open = true
	aw := fyne.CurrentApp().NewWindow("List NFA")
	aw.Resize(fyne.NewSize(330, 700))
	aw.SetIcon(Resource.SmallIcon)
	MenuControl.List_button.Hide()
	aw.SetCloseIntercept(func() {
		MenuControl.list_open = false
		if rpc.Wallet.Connect {
			MenuControl.List_button.Show()
		}
		aw.Close()
	})
	aw.SetFixedSize(true)

	label := widget.NewLabel("                       Listing fee 0.1 Dero")
	options := []string{"Auction", "Sale"}
	list := widget.NewSelect(options, func(s string) {})
	list.PlaceHolder = "Type:"

	duration := table.NilNumericalEntry()
	duration.SetPlaceHolder("Duration in Hours:")
	duration.Validator = validation.NewRegexp(`^\d{1,}$`, "Format Not Valid")

	start := table.NilNumericalEntry()
	start.SetPlaceHolder("Start Price:")
	start.Validator = validation.NewRegexp(`\d{1,}\.\d{1,5}$`, "Format Not Valid")

	charAddr := widget.NewEntry()
	charAddr.SetPlaceHolder("Charity Donation Address:")
	charAddr.Validator = validation.NewRegexp(`^\w{66,66}$`, "Format Not Valid")

	charPerc := table.NilNumericalEntry()
	charPerc.SetPlaceHolder("Charity Donation %:")
	charPerc.Validator = validation.NewRegexp(`^\d{1,2}$`, "Format Not Valid")

	MenuControl.Set_list = widget.NewButton("Set Listing", func() {
		if duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
			MenuControl.Set_list.Hide()
			confirmAssetList(list.Selected, duration.Text, start.Text, charAddr.Text, charPerc.Text)
		}
	})

	items := container.NewVBox(
		layout.NewSpacer(),
		label,
		list,
		duration,
		start,
		charAddr,
		charPerc,
		MenuControl.Set_list)

	img := *canvas.NewImageFromResource(Resource.Back3)
	aw.SetContent(
		container.New(layout.NewMaxLayout(),
			&img,
			items))
	aw.Show()
}

func confirmAssetList(list, dur, start, charAddr, charPerc string) { /// listing confirmation
	ocw := fyne.CurrentApp().NewWindow("Confirm")
	ocw.SetIcon(Resource.SmallIcon)
	ocw.Resize(fyne.NewSize(450, 450))
	ocw.SetFixedSize(true)
	label := widget.NewLabel("SCID: " + MenuControl.Viewing_asset + "\n\nList Type: " + list + "\n\nDuration: " + dur + " Hours\n\nStart Price: " + start + " Dero\n\nDonation Address: " + charAddr + "\n\nDonate Percent: " + charPerc + "\n\nConfirm Asset Listing")
	label.Wrapping = fyne.TextWrapWord
	cancel_button := widget.NewButton("Cancel", func() {
		MenuControl.Set_list.Show()
		ocw.Close()
	})

	confirm_button := widget.NewButton("Confirm", func() {
		d := uint64(stringToInt64(dur))
		s := ToAtomicFive(start)
		cp := uint64(stringToInt64(charPerc))
		rpc.NfaSetListing(MenuControl.Viewing_asset, list, charAddr, d, s, cp)
		MenuControl.Set_list.Show()
		ocw.Close()
	})

	display := container.NewVBox(label, layout.NewSpacer())
	options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewBorder(nil, options, nil, nil, display)

	img := *canvas.NewImageFromResource(Resource.Back4)
	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	ocw.SetContent(
		container.New(layout.NewMaxLayout(),
			&img,
			alpha,
			content))
	ocw.Show()
}

func ToAtomicFive(v string) uint64 {
	f, err := strconv.ParseFloat(v, 64)

	if err != nil {
		log.Println("To Atmoic Conversion Error", err)
		return 0
	}

	ratio := math.Pow(10, float64(5))
	rf := math.Round(f*ratio) / ratio

	return uint64(math.Round(rf * 100000))
}

func IntroTree() fyne.CanvasObject {
	list := map[string][]string{
		"":                      {"Welcome to dReams"},
		"Welcome to dReams":     {"Get Started", "Contracts", "Assets", "Market"},
		"Get Started":           {"You will need a Dero wallet to play", "Can use local daemon, or remote daemon options are availible in drop down", "Enter daemon rpc address, wallet rpc address and user:pass", "Press connect, D & W indicators at top right of screen will light up on successful connection", "On first start start up of app, Gnomon will take ~10 seconds to create your local db", "Gnomon idicator will have a stripe when starting or syncing, indicator will turn solid when startup, sync and scan are completed"},
		"Contracts":             {"Holdero", "Baccarat", "Predictions", "Sports", "Tarot"},
		"Holdero":               {"Multiplayer Texas Hold'em style on chian poker", "No limit, single raise game. Table owners choose game params", "Six players max at a table", "No side pots, must call or fold", "Public and private tables can use Dero or dReam Tokens", "Tournament tables can be set up to use any Token", "View table listings or launch your own Holdero contract from the contracts tab"},
		"Baccarat":              {"A popular table game, where closest to 9 wins", "Uses dReam Tokens for betting"},
		"Predictions":           {"Prediction contracts are for binary based predictions, (higher/lower, yes/no)", "How predictions works", "Current Markets", "dReams Client aggregated price feed", "View active prediction contracts in predictions tab or launch your own prediction contract from the contracts tab"},
		"How predictions works": {"P2P predictions", "Variable time limits allowing for different prediction set ups, each contract runs one prediction at a time", "Click a contract from the list to view it", "Closes at, is when the contract will stop accepting predictions", "Mark (price or value you are predicting on) can be set on prediction initialization or it can given live", "Posted with in, is the acceptable time frame to post the live Mark", "If Mark is not posted, prediction is voided and you will be refunded", "Payout after, is when the Final price is posted and compared to the mark to determine winners", "If the final price is not posted with in refund time frame, prediction is void and you will be refunded"},
		"Current Markets":       {"DERO-BTC", "XMR-BTC", "BTC-USDT", "DERO-USDT", "XMR-USDT", "DERO-Difficulty", "DERO-Block Time", "DERO-Block Number"},
		"Sports":                {"Sports contracts are for sports wagers", "How sports works", "Current Leagues", "Live game scores, and game schedules", "View active sports contracts in sports tab or launch your own sports contract from the contracts tab"},
		"How sports works":      {"P2P betting", "Variable time limits, one contract can run miltiple games at the same time", "Click a contract from the list to view it", "Any active games on the contract will populate, you can pick which game you'd like to play from the drop down", "Closes at, is when the contrcts stops accepting picks", "Default payout time after close is 4hr, this is when winner will be posted from client feed", "Default refund time is 8hr after close, meaning if winner is not provided past that time you will be refunded", "A Tie refunds pot to all all participants"},
		"Current Leagues":       {"EPL", "NBA", "NFL", "NHL", "Bellator", "UFC"},
		"Tarot":                 {"On chian Tarot readings", "Iluma cards and readings created by Kalina Lux"},
		"Assets":                {"View any owned assets held in wallet", "Put owned assets up for auction or for sale", "Indexer, add custom contracts to your index and search current index db"},
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
