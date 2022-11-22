package menu

import (
	"image/color"
	"log"
	"math"
	"runtime"
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
	DAEMON_RPC_REMOTE2 = "dero-node.mysrv.cloud:10102"
	DAEMON_RPC_REMOTE3 = "derostats.io:10102"
)

type playerOptions struct {
	list_open      bool
	Daemon_config  string
	Viewing_asset  string
	daemon_check   *widget.Check
	wallet_check   *widget.Check
	contract_check *widget.Check
	contract_input *widget.SelectEntry
	table_options  *widget.List
	table_fav      *widget.List
	holdero_unlock *widget.Button
	holdero_new    *widget.Button
	Bet_unlock     *widget.Button
	Bet_new        *widget.Button
	Bet_menu       *widget.Button
	Claim_button   *widget.Button
	List_button    *widget.Button
	Set_list       *widget.Button
	Stats_box      fyne.Container
}

type resources struct {
	SmallIcon fyne.Resource
	Frame     fyne.Resource
	Back1     fyne.Resource
	Back2     fyne.Resource
	Back3     fyne.Resource
	Back4     fyne.Resource
}

var Resource resources
var TableList []string
var FavoriteList []string
var PlayerControl playerOptions

func GetMenuResources(r1, r2, r3, r4, r5, r6 fyne.Resource) {
	Resource.SmallIcon = r1
	Resource.Frame = r2
	Resource.Back1 = r3
	Resource.Back2 = r4
	Resource.Back3 = r5
	Resource.Back4 = r6
}

func disconnected() {
	rpc.Wallet.PokerOwner = false
	rpc.Wallet.BetOwner = false
	table.Settings.FaceSelect.Options = []string{"Light", "Dark"}
	table.Settings.BackSelect.Options = []string{"Light", "Dark"}
	table.Settings.ThemeSelect.Options = []string{"Main"}
	table.Settings.AvatarSelect.Options = []string{"None"}
	table.Settings.FaceSelect.SetSelectedIndex(0)
	table.Settings.BackSelect.SetSelectedIndex(0)
	table.Settings.FaceSelect.Refresh()
	table.Settings.BackSelect.Refresh()
	table.Settings.ThemeSelect.Refresh()
	table.Settings.AvatarSelect.Refresh()
	table.Assets.Assets = []string{}
	table.Actions.NameEntry.Text = ""
	table.Actions.NameEntry.Enable()
	table.Actions.NameEntry.Refresh()
	Market.Auction_list.UnselectAll()
	Market.Buy_list.UnselectAll()
	Market.Buy_now = []string{}
	Market.Auctions = []string{}
	Market.Buy_list.Refresh()
	Market.Auction_list.Refresh()
	Market.Icon = *canvas.NewImageFromImage(nil)
	Market.Cover = *canvas.NewImageFromImage(nil)
	Market.Viewing = ""
	ResetAuctionInfo()
	AuctionInfo()
	TableList = []string{}
	PlayerControl.table_options.Refresh()
}

func CheckConnection() {
	if rpc.Signal.Daemon {
		PlayerControl.daemon_check.SetChecked(true)
		disableIndex(false)
	} else {
		PlayerControl.daemon_check.SetChecked(false)
		TableList = []string{}
		PlayerControl.table_options.Refresh()
		disableOwnerControls(true)
		disableBaccActions(true)
		disableActions(true)
		disableIndex(true)
		Gnomes.Init = false
		Gnomes.Checked = false
	}

	if rpc.Wallet.Connect {
		PlayerControl.wallet_check.SetChecked(true)
		disableActions(false)
	} else {
		PlayerControl.wallet_check.SetChecked(false)
		TableList = []string{}
		PlayerControl.table_options.Refresh()
		disableOwnerControls(true)
		disableBaccActions(true)
		DisablePreditions(true)
		disableSports(true)
		disableActions(true)
		disconnected()
		Gnomes.Checked = false

	}

	if rpc.Signal.Contract {
		PlayerControl.contract_check.SetChecked(true)
	} else {
		PlayerControl.contract_check.SetChecked(false)
		disableOwnerControls(true)
		rpc.Signal.Sit = true
	}
}

func DaemonConnectedBox(b bool) fyne.Widget {
	PlayerControl.daemon_check = widget.NewCheck("", func(b bool) {
		if !Gnomes.Init {
			startGnomon(rpc.Round.Daemon)
		}

		if !b {
			StopGnomon(Gnomes.Init)
		}
	})
	PlayerControl.daemon_check.Disable()

	return PlayerControl.daemon_check
}

func WalletConnectedBox() fyne.Widget {
	PlayerControl.wallet_check = widget.NewCheck("", func(b bool) {})
	PlayerControl.wallet_check.Disable()

	return PlayerControl.wallet_check
}

func HolderoContractConnectedBox() fyne.Widget {
	PlayerControl.contract_check = widget.NewCheck("", func(b bool) {
		if !b {
			disableOwnerControls(true)
		}
	})
	PlayerControl.contract_check.Disable()

	return PlayerControl.contract_check
}

func DaemonRpcEntry() fyne.Widget {
	var options = []string{"", DAEMON_RPC_DEFAULT, DAEMON_RPC_REMOTE1, DAEMON_RPC_REMOTE2, DAEMON_RPC_REMOTE3}
	if PlayerControl.Daemon_config != "" {
		options = append(options, PlayerControl.Daemon_config)
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
	PlayerControl.contract_input = widget.NewSelectEntry(nil)
	options := []string{""}
	PlayerControl.contract_input.SetOptions(options)
	PlayerControl.contract_input.PlaceHolder = "Holdero Contract Address: "
	PlayerControl.contract_input.OnCursorChanged = func() {
		if rpc.Signal.Daemon {
			text := PlayerControl.contract_input.Text
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
			} else {
				rpc.Signal.Contract = false
				PlayerControl.contract_check.SetChecked(false)
			}
		}
	}

	this := binding.BindString(&rpc.Round.Contract)
	PlayerControl.contract_input.Bind(this)

	return PlayerControl.contract_input
}

func RpcConnectButton() fyne.Widget {
	button := widget.NewButton("Connect", func() {
		go func() {
			rpc.Ping()
			rpc.GetAddress()
			CheckConnection()
		}()
	})

	return button
}

func HolderoUnlockButton() fyne.Widget { /// unlock table and upload
	PlayerControl.holdero_unlock = widget.NewButton("Unlock Holdero Contract", func() {
		holderoMenuConfirm(1)
	})

	PlayerControl.holdero_unlock.Hide()

	return PlayerControl.holdero_unlock
}

func NewTableButton() fyne.Widget {
	PlayerControl.holdero_new = widget.NewButton("New Holdero Table", func() {
		holderoMenuConfirm(2)
	})

	PlayerControl.holdero_new.Hide()

	return PlayerControl.holdero_new
}

func BettingUnlockButton() fyne.Widget { /// unlock betting and upload
	PlayerControl.Bet_unlock = widget.NewButton("Unlock Betting Contracts", func() {
		bettingMenuConfirm(1)
	})

	PlayerControl.Bet_unlock.Hide()

	return PlayerControl.Bet_unlock
}

func NewBettingButton() fyne.Widget {
	PlayerControl.Bet_new = widget.NewButton("New Betting Contract", func() {
		bettingMenuConfirm(2)
	})

	PlayerControl.Bet_new.Hide()

	return PlayerControl.Bet_new
}

func TableListings() fyne.Widget { /// tables contracts
	PlayerControl.table_options = widget.NewList(
		func() int {
			return len(TableList)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(TableList[i])
		})

	PlayerControl.table_options.OnSelected = func(id widget.ListItemID) {
		if id != 0 {
			split := strings.Split(TableList[id], "   ")
			PlayerControl.contract_input.SetText(split[2])
			go GetTableStats(split[2], true)
			rpc.Times.Kick_block = rpc.StringToInt(rpc.Wallet.Height)
		}
	}

	return PlayerControl.table_options
}

func FavoriteListings() fyne.Widget {
	PlayerControl.table_fav = widget.NewList(
		func() int {
			return len(FavoriteList)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(FavoriteList[i])
		})
	PlayerControl.table_fav.Resize(fyne.NewSize(360, 680))
	PlayerControl.table_fav.Move(fyne.NewPos(5, 10))
	PlayerControl.table_fav.OnSelected = func(id widget.ListItemID) {
		// if id != 0 {

		// }
	}

	return PlayerControl.table_fav
}

func NameEntry() fyne.CanvasObject {
	name := widget.NewEntry()
	name.PlaceHolder = "Name:"
	this := binding.BindString(&table.Poker_name)
	name.Bind(this)
	name.Validator = validation.NewRegexp(`^\w{3,10}$`, "Format Not Valid")

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

	PlayerControl.Stats_box = *container.NewVBox(Stats.Name, Stats.Desc, Stats.Version, Stats.Last, Stats.Seats, TableIcon(nil))

	return &PlayerControl.Stats_box
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

	} else {
		table.Assets.Index_button.Show()
		table.Assets.Index_search.Show()
		if rpc.Wallet.Connect {
			PlayerControl.Claim_button.Show()
			table.Assets.Header_box.Show()
			Market.Market_box.Show()
			if PlayerControl.list_open {
				PlayerControl.List_button.Hide()
			}
		} else {
			PlayerControl.List_button.Hide()
			PlayerControl.Claim_button.Hide()
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
		table.Actions.Higher.Hide()
		table.Actions.Lower.Hide()
		table.Actions.Change.Hide()
		table.Actions.NameEntry.Hide()
	} else {
		table.Actions.Higher.Show()
		table.Actions.Lower.Show()
		table.Actions.NameEntry.Show()
	}
	table.Actions.Higher.Refresh()
	table.Actions.Lower.Refresh()
	table.Actions.Change.Refresh()
	table.Actions.NameEntry.Refresh()
}

func disableSports(d bool) {
	if d {
		table.Actions.Sports_box.Hide()
	}

	table.Actions.Sports_box.Refresh()
}

func disableActions(d bool) {
	if d {
		table.Actions.Dreams.Hide()
		table.Actions.Dero.Hide()
		table.Actions.DEntry.Hide()
		PlayerControl.holdero_unlock.Hide()
		PlayerControl.holdero_new.Hide()
		PlayerControl.Bet_new.Hide()
		PlayerControl.Bet_unlock.Hide()
		PlayerControl.Bet_menu.Hide()
	} else {
		table.Actions.Dreams.Show()
		table.Actions.Dero.Show()
		table.Actions.DEntry.Show()
	}

	table.Actions.Dreams.Refresh()
	table.Actions.DEntry.Refresh()
	table.Actions.Dero.Refresh()
	PlayerControl.holdero_unlock.Refresh()
	PlayerControl.holdero_new.Refresh()
	PlayerControl.Bet_new.Refresh()
	PlayerControl.Bet_unlock.Refresh()
	PlayerControl.Bet_menu.Refresh()
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
		PlayerControl.Bet_new.Show()
		PlayerControl.Bet_unlock.Hide()
		PlayerControl.Bet_menu.Show()
	} else {
		rpc.Wallet.BetOwner = false
		PlayerControl.Bet_new.Hide()
		PlayerControl.Bet_unlock.Show()
		PlayerControl.Bet_menu.Hide()
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
	if runtime.GOOS == "windows" {
		cw.Resize(fyne.NewSize(650, 650))
	} else {
		cw.Resize(fyne.NewSize(550, 550))
	}
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

	content := container.NewVBox(label, layout.NewSpacer(), box)

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
	if runtime.GOOS == "windows" {
		cw.Resize(fyne.NewSize(650, 650))
	} else {
		cw.Resize(fyne.NewSize(550, 550))
	}
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

	content := container.NewVBox(label, layout.NewSpacer(), box)

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

	PlayerControl.List_button = widget.NewButton("List Asset", func() {
		listMenu()
	})

	PlayerControl.Claim_button = widget.NewButton("Claim", func() {
		if len(table.Assets.Index_entry.Text) == 64 {
			if isNfa(table.Assets.Index_entry.Text) {
				rpc.ClaimNfa(table.Assets.Index_entry.Text)
			}
		}
	})

	table.Assets.Index_button.Hide()
	table.Assets.Index_search.Hide()
	PlayerControl.List_button.Hide()
	PlayerControl.Claim_button.Hide()

	table.Assets.Gnomes_index = canvas.NewText(" Indexed SCIDs: ", color.White)
	table.Assets.Gnomes_index.TextSize = 18

	bottom_grid := container.NewAdaptiveGrid(3, table.Assets.Gnomes_index, table.Assets.Index_button, table.Assets.Index_search)
	top_grid := container.NewAdaptiveGrid(3, layout.NewSpacer(), PlayerControl.Claim_button, PlayerControl.List_button)
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
		PlayerControl.Viewing_asset = split[1]
		go GetOwnedAssetStats(split[1])
	}

	box := container.NewMax(table.Assets.Asset_list)

	return box
}

func listMenu() { /// asset listing
	PlayerControl.list_open = true
	aw := fyne.CurrentApp().NewWindow("List NFA")
	aw.Resize(fyne.NewSize(330, 700))
	aw.SetIcon(Resource.SmallIcon)
	PlayerControl.List_button.Hide()
	aw.SetCloseIntercept(func() {
		PlayerControl.list_open = false
		if rpc.Wallet.Connect {
			PlayerControl.List_button.Show()
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

	PlayerControl.Set_list = widget.NewButton("Set Listing", func() {
		if duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
			PlayerControl.Set_list.Hide()
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
		PlayerControl.Set_list)

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
	ocw.Resize(fyne.NewSize(550, 550))
	ocw.SetFixedSize(true)
	label := widget.NewLabel("SCID: " + PlayerControl.Viewing_asset + "\n\nList Type: " + list + "\n\nDuration: " + dur + "\n\nStart Price: " + start + " Dero\n\nDonate Address: " + charAddr + "\n\nDonate Percent: " + charPerc + "\n\nConfirm Asset Listing")
	label.Wrapping = fyne.TextWrapWord
	cancel_button := widget.NewButton("Cancel", func() {
		PlayerControl.Set_list.Show()
		ocw.Close()
	})

	confirm_button := widget.NewButton("Confirm", func() {
		d := uint64(stringToInt64(dur))
		s := ToAtomicFive(start)
		cp := uint64(stringToInt64(charPerc))
		rpc.NfaSetListing(PlayerControl.Viewing_asset, list, charAddr, d, s, cp)
		PlayerControl.Set_list.Show()
		ocw.Close()
	})

	display := container.NewVBox(label, layout.NewSpacer())
	options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewVBox(display, layout.NewSpacer(), options)

	img := *canvas.NewImageFromResource(Resource.Back4)
	ocw.SetContent(
		container.New(layout.NewMaxLayout(),
			&img,
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
		"":                  {"Welcome to dReams"},
		"Welcome to dReams": {"Contracts", "Assets", "Market"},
		"Contracts":         {"Holdero", "Baccarat", "Predictions", "Sports"},
		"Holdero":           {"Multiplayer Texas Hold'em style on chian poker", "No limit, single raise game. Table owners choose game params", "Six players max at a table", "No side pots, must call or fold", "Can use Dero or dReam Tokens", "View table listings or launch your own Holdero contract from the contracts tab"},
		"Baccarat":          {"A popular table game, where closest to 9 wins", "Uses dReam Tokens for betting"},
		"Predictions":       {"Prediction contracts are for binary based predictions, (higher/lower, yes/no)", "Variable time limits allowing for different prediction set ups, each contract runs one prediction at a time", "Current Markets", "dReams Client aggregated price feed", "View active prediction contracts in predictions tab or launch your own prediction contract from the contracts tab"},
		"Current Markets":   {"BTC-USDT", "DERO-USDT", "XMR-USDT"},
		"Sports":            {"Sports contracts are for sports wagers", "Variable time limits, one contract can run miltiple games at the same time", "Current Leagues", "Live game scores, and game schedules", "View active sports contracts in sports tab or launch your own sports contract from the contracts tab"},
		"Current Leagues":   {"FIFA", "NBA", "NFL", "NHL"},
		"Assets":            {"View any owned assets held in wallet", "Put owned assets up for auction or for sale", "Indexer, add custom contracts to your index and search current index db"},
		"Market":            {"View any in game assets up for auction or sale", "Bid on or buy assets", "Cancel or close out any existing listings"},
	}

	tree := widget.NewTreeWithStrings(list)

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	max := container.NewMax(alpha, tree)

	return max
}
