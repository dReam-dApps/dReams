package dwidget

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/rpc"
	"github.com/deroproject/derohe/config"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/walletapi/xswd"
)

type AmountEntry struct {
	xwidget.NumericalEntry
	Prefix    string
	Increment float64
	Decimal   uint
}

// Create new numerical entry with increment change on up or down key stroke
//   - If entry does not require prefix, pass ""
//   - Increment and Decimal for entry input control
func NewAmountEntry(prefix string, increment float64, decimal uint) *AmountEntry {
	entry := &AmountEntry{}
	entry.ExtendBaseWidget(entry)
	entry.AllowFloat = true
	entry.Prefix = prefix
	entry.Increment = increment
	entry.Decimal = decimal

	return entry
}

// Accepts int or '.'
func (e *AmountEntry) TypedRune(r rune) {
	if r >= '0' && r <= '9' {
		e.Entry.TypedRune(r)
		return
	}

	if e.AllowFloat && r == '.' {
		e.Entry.TypedRune(r)
	}
}

// Increase or decrease of Increment on TypedKey
func (e *AmountEntry) TypedKey(k *fyne.KeyEvent) {
	value := strings.Trim(e.Entry.Text, e.Prefix)
	if e.Decimal > 5 {
		e.Decimal = 5
	}

	switch k.Name {
	case fyne.KeyUp:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			if e.Increment >= 0.00001 {
				e.Entry.SetText(e.Prefix + strconv.FormatFloat(float64(f+e.Increment), 'f', int(e.Decimal), 64))
			}
		}
	case fyne.KeyDown:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			if f >= e.Increment {
				e.Entry.SetText(e.Prefix + strconv.FormatFloat(float64(f-e.Increment), 'f', int(e.Decimal), 64))
			}
		}
	}
	e.Entry.TypedKey(k)
}

// Dero wallet and daemon connection entry objects
type DeroEntries struct {
	Container     *fyne.Container
	Daemon        *widget.SelectEntry
	RPC           *widget.SelectEntry
	XSWD          *widget.SelectEntry
	Auth          *widget.Entry
	Balance       *canvas.Text
	Button        *widget.Button
	Connected     *widget.Check
	offset        int
	layout        int
	daemonOptions []string
}

// Create a horizontal layout container with entries for daemon, wallet RPC or XSWD connections
//   - Objects bound to dReams rpc Daemon and Wallet vars with disconnect control
//   - Balance canvas to display wallet balance
//   - Update Button OnTapped func() as needed, it has default connection func for RPC or XSWD connections
//   - Offset of 1 puts entries on trailing edge
func NewHorizontalEntries(tag string, offset int, d *dreams.AppObject) *DeroEntries {
	entryDaemon, daemons := NewDaemonEntry(nil)

	entryRPC, entryAuth := NewWalletRPCEntries(tag, nil)

	entryXSWD := NewWalletXSWDEntry(tag)

	_, names := dreams.GetDeroAccounts()
	entryDERO := widget.NewSelectEntry(names)
	entryDERO.PlaceHolder = "dero.db path:"

	entryPass := widget.NewPasswordEntry()
	entryPass.PlaceHolder = "Password:"

	selectType := widget.NewSelect([]string{"RPC", "XSWD", "DERO"}, nil)
	selectType.PlaceHolder = "(Select)"
	selectType.SetSelectedIndex(0)

	button := widget.NewButtonWithIcon("", dreams.FyneIcon("confirm"), nil)
	button.OnTapped = onTapped(tag, selectType, entryAuth, entryPass, entryRPC, entryXSWD, entryDERO, button, d)

	balance := canvas.NewText(fmt.Sprintf("%.5f DERO", 0.0), bundle.TextColor)

	check := widget.NewCheck("", nil)
	check.Disable()
	check.Hide()

	layoutRPC := container.NewHBox(
		container.NewStack(NewSpacer(210, 0), entryDaemon),
		container.NewStack(NewSpacer(180, 0), entryRPC),
		container.NewStack(NewSpacer(180, 0), entryAuth),
		selectType)

	layoutXSWD := container.NewHBox(
		container.NewStack(NewSpacer(210, 0), entryDaemon),
		container.NewStack(NewSpacer(180, 0), entryXSWD),
		layout.NewSpacer(),
		selectType)

	layoutDERO := container.NewHBox(
		container.NewStack(NewSpacer(210, 0), entryDaemon),
		container.NewStack(NewSpacer(180, 0), entryDERO),
		container.NewStack(NewSpacer(180, 0), entryPass),
		selectType)

	layoutAll := container.NewBorder(nil, nil, check, button, layoutRPC)

	selectType.OnChanged = func(s string) {
		switch s {
		case "RPC":
			layoutAll.Objects[0] = layoutRPC
		case "XSWD":
			layoutAll.Objects[0] = layoutXSWD
		case "DERO":
			layoutAll.Objects[0] = layoutDERO
			_, names := dreams.GetDeroAccounts()
			layoutDERO.Objects[1].(*fyne.Container).Objects[1].(*widget.SelectEntry).SetOptions(names)
		}
	}

	deroEntries := &DeroEntries{
		Daemon:        entryDaemon,
		RPC:           entryRPC,
		XSWD:          entryXSWD,
		Auth:          entryAuth,
		Balance:       balance,
		Button:        button,
		Connected:     check,
		offset:        offset,
		daemonOptions: daemons,
		layout:        0,
	}

	if offset == 1 {
		deroEntries.Container = container.NewHBox(layout.NewSpacer(), container.NewHBox(layout.NewSpacer(), balance), layoutAll)
	} else {
		deroEntries.Container = container.NewHBox(layoutAll, container.NewHBox(balance))
	}

	return deroEntries
}

// Create a vertical layout container with entries for daemon, wallet and user:pass
//   - Objects bound to dReams rpc Daemon and Wallet vars with disconnect control
//   - Balance canvas to display wallet balance
//   - Update Button OnTapped func() as needed, it has default connection func for RPC or XSWD connections
func NewVerticalEntries(tag string, d *dreams.AppObject) *DeroEntries {
	entryDaemon, daemons := NewDaemonEntry(nil)

	entryRPC, entryAuth := NewWalletRPCEntries(tag, nil)

	entryXSWD := NewWalletXSWDEntry(tag)

	_, names := dreams.GetDeroAccounts()
	entryDERO := widget.NewSelectEntry(names)
	entryDERO.PlaceHolder = "dero.db path:"

	entryPass := widget.NewPasswordEntry()
	entryPass.PlaceHolder = "Password:"

	selectType := widget.NewSelect([]string{"RPC", "XSWD", "DERO"}, nil)
	selectType.PlaceHolder = "(Select)"
	selectType.SetSelectedIndex(0)

	button := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "confirm"), nil)
	button.OnTapped = onTapped(tag, selectType, entryAuth, entryPass, entryRPC, entryXSWD, entryDERO, button, d)

	balance := canvas.NewText(fmt.Sprintf("%.5f Dero", 0.0), bundle.TextColor)

	check := widget.NewCheck("", nil)
	check.Disable()
	check.Hide()

	deroEntries := &DeroEntries{
		Daemon:        entryDaemon,
		RPC:           entryRPC,
		XSWD:          entryXSWD,
		Auth:          entryAuth,
		Balance:       balance,
		Button:        button,
		Connected:     check,
		daemonOptions: daemons,
		layout:        1,
	}

	layoutRPC := container.NewVBox(
		entryDaemon,
		entryRPC,
		entryAuth)

	layoutXSWD := container.NewVBox(
		entryDaemon,
		entryXSWD,
		NewSpacer(0, 33))

	layoutDERO := container.NewVBox(
		entryDaemon,
		entryDERO,
		entryPass)

	layoutAll := container.NewVBox(
		layoutRPC,
		layout.NewSpacer(),
		container.NewCenter(balance),
		container.NewAdaptiveGrid(2, selectType, button),
		check)

	selectType.OnChanged = func(s string) {
		switch s {
		case "RPC":
			layoutAll.Objects[0] = layoutRPC
		case "XSWD":
			layoutAll.Objects[0] = layoutXSWD
		case "DERO":
			layoutAll.Objects[0] = layoutDERO
			_, names := dreams.GetDeroAccounts()
			layoutDERO.Objects[1].(*widget.SelectEntry).SetOptions(names)
		}
	}

	deroEntries.Container = container.NewVBox(layoutAll)

	return deroEntries
}

// Creates a basic daemon select entry with default remote options
//   - Entry bound to rpc.Daemon.Rpc
//   - Pass 'defaults' for custom default port options
func NewDaemonEntry(defaults []string) (entry *widget.SelectEntry, daemons []string) {
	daemons = []string{""}
	if defaults != nil {
		daemons = append(daemons, defaults...)
	} else {
		defaults = []string{
			rpc.DAEMON_RPC_DEFAULT,
			rpc.DAEMON_RPC_REMOTE1,
			rpc.DAEMON_RPC_REMOTE2,
			rpc.DAEMON_RPC_REMOTE5,
			rpc.DAEMON_RPC_REMOTE6,
		}

		daemons = append(daemons, defaults...)
	}

	entry = widget.NewSelectEntry(daemons)
	entry.SetPlaceHolder("Daemon RPC:")
	entry.Bind(binding.BindString(&rpc.Daemon.Rpc))

	return
}

// Creates basic wallet RPC port and auth connection entries
//   - Entries bound to rpc.Wallet.RPC.Port and rpc.Wallet.RPC.Auth
//   - Default OnChanged will disconnect wallet
//   - Pass 'defaults' for custom default port options
func NewWalletRPCEntries(tag string, defaults []string) (entryRPC *widget.SelectEntry, entryAuth *widget.Entry) {
	options := []string{""}
	if defaults != nil {
		options = append(options, defaults...)
	} else {
		options = append(options, fmt.Sprintf("127.0.0.1:%d", config.Mainnet.Wallet_RPC_Default_Port))
	}

	entryRPC = widget.NewSelectEntry(options)
	entryRPC.SetPlaceHolder("Wallet RPC:")
	entryRPC.Bind(binding.BindString(&rpc.Wallet.RPC.Port))
	entryRPC.SetText(options[1])
	entryRPC.OnChanged = func(s string) {
		if rpc.Wallet.IsConnected() {
			rpc.Wallet.CloseConnections(tag)
		}
	}

	entryAuth = widget.NewPasswordEntry()
	entryAuth.SetPlaceHolder("RPC user:pass")
	pass_bind := binding.BindString(&rpc.Wallet.RPC.Auth)
	entryAuth.Bind(pass_bind)
	entryAuth.OnChanged = func(s string) {
		if rpc.Wallet.IsConnected() {
			rpc.Wallet.CloseConnections(tag)
		}
	}

	return
}

// Creates basic wallet XSWD connection entry
//   - Entry bound to rpc.Wallet.WS.Port
//   - Default OnChanged will disconnect wallet
func NewWalletXSWDEntry(tag string) (entryXSWD *widget.SelectEntry) {
	options := []string{"", fmt.Sprintf("127.0.0.1:%d", xswd.XSWD_PORT)}

	entryXSWD = widget.NewSelectEntry(options)
	entryXSWD.SetPlaceHolder("XSWD Port:")
	entryXSWD.Bind(binding.BindString(&rpc.Wallet.WS.Port))
	entryXSWD.SetText(options[1])
	entryXSWD.OnChanged = func(s string) {
		if rpc.Wallet.IsConnected() {
			rpc.Wallet.CloseConnections(tag)
		}
	}

	return
}

// OnTapped default function for dwidget wallet connection button
func onTapped(tag string, selectType *widget.Select, entryAuth, entryPass *widget.Entry, entryRPC, entryXSWD, entryDERO *widget.SelectEntry, button *widget.Button, d *dreams.AppObject) func() {
	return func() {
		if selectType.SelectedIndex() < 0 {
			dialog.NewInformation("Select Connect Type", "Select RPC or XSWD wallet connection", d.Window).Show()
			return
		}

		switch selectType.Selected {
		case "RPC":
			// Disconnect from RPC
			if !rpc.Wallet.RPC.IsClosed() {
				button.Importance = widget.MediumImportance
				entryRPC.Enable()
				entryAuth.Enable()
				rpc.Wallet.CloseConnections(tag)
				button.Icon = dreams.FyneIcon("confirm")
				button.Refresh()
				selectType.Enable()
				return
			}

			// Connect to RPC
			rpc.Wallet.RPC.Init()
			rpc.GetAddress(tag)
			rpc.Ping()
			if rpc.Wallet.IsConnected() {
				button.Importance = widget.HighImportance
				button.Icon = dreams.FyneIcon("cancel")
				button.Refresh()
				entryRPC.Disable()
				entryAuth.Disable()
				selectType.Disable()
			} else {
				rpc.Wallet.CloseConnections(tag)
			}
		case "XSWD":
			// Disconnect from XSWD
			if !rpc.Wallet.WS.IsClosed() {
				button.Importance = widget.MediumImportance
				rpc.Wallet.CloseConnections(tag)
				entryXSWD.Enable()
				button.Icon = dreams.FyneIcon("confirm")
				button.Refresh()
				selectType.Enable()
				return
			}

			// Connect to XSWD
			go func() {
				button.Disable()
				entryXSWD.Disable()
				selectType.Disable()
				// TODO pass app data
				if rpc.Wallet.WS.Init(d.XSWD) {
					rpc.GetAddress(tag)
					if rpc.Wallet.IsConnected() {
						button.Importance = widget.HighImportance
						button.Icon = dreams.FyneIcon("cancel")
						button.Refresh()
						button.Enable()
						return
					}

					rpc.Wallet.CloseConnections(tag)
				}

				entryXSWD.Enable()
				button.Importance = widget.MediumImportance
				button.Icon = dreams.FyneIcon("confirm")
				button.Refresh()
				button.Enable()
				selectType.Enable()
			}()
		case "DERO":
			go func() {
				button.Disable()
				entryPass.Disable()
				entryDERO.Disable()
				selectType.Disable()
				defer func() {
					button.Enable()
				}()

				// Close wallet
				if !rpc.Wallet.File.IsNil() {
					rpc.Wallet.CloseConnections(tag)
					selectType.Enable()
					entryPass.Enable()
					entryDERO.Enable()
					button.Importance = widget.MediumImportance
					button.Icon = dreams.FyneIcon("confirm")
					button.Refresh()

					return
				} else {
					rpc.Ping()
					// Check if connected to daemon
					if !rpc.Daemon.IsConnected() {
						dialog.NewInformation("Select Daemon", "Connect to a daemon", d.Window).Show()
						entryPass.Enable()
						entryDERO.Enable()
						selectType.Enable()
						return
					}

					network := "mainnet"
					if !globals.IsMainnet() {
						network = "testnet"
					}

					dir := filepath.Join(dreams.GetDir(), network) + string(filepath.Separator)
					path := filepath.Join(dir, entryDERO.Text)
					if strings.HasPrefix(entryDERO.Text, string(filepath.Separator)) {
						path = entryDERO.Text
					}

					// Open wallet
					if err := rpc.Wallet.OpenWalletFile(tag, path, entryPass.Text); err != nil {
						dialogError := dialog.NewInformation("Error", fmt.Sprintf("%s", err), d.Window)
						dialogError.Show()
						entryPass.Enable()
						entryDERO.Enable()
						selectType.Enable()
						return
					}

					entryDERO.Disable()
					entryPass.Disable()
					button.Importance = widget.HighImportance
					button.Icon = dreams.FyneIcon("cancel")
					button.Refresh()
				}
			}()
		}
	}
}

// Refresh Balance of DeroEntries
//   - Gets balance from rpc.Wallet.Balance("DERO")
func (d *DeroEntries) RefreshBalance() {
	d.Balance.Text = (fmt.Sprintf("%.5f DERO", float64(rpc.Wallet.Balance("DERO"))/100000))
	d.Balance.Refresh()
}

// Add 'new' options to default daemon rpc entry
func (d *DeroEntries) AddDaemonOptions(new []string) {
	current := d.daemonOptions
	d.Daemon.SetOptions(append(current, new...))
	d.Daemon.Refresh()
}

// Add canvas object indicators to DeroEntries, switching for layout
func (d *DeroEntries) AddIndicator(ind fyne.CanvasObject) {
	switch d.layout {
	case 0:
		d.Container.Objects[1].(*fyne.Container).Add(ind)
	case 1:
		d.Container.Add(container.NewCenter(ind))
	default:
		// nothing
	}
}
