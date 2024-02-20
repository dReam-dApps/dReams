package dwidget

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/rpc"
)

type DeroAmts struct {
	xwidget.NumericalEntry
	Prefix    string
	Increment float64
	Decimal   uint
}

// Create new numerical entry with increment change on up or down key stroke
//   - If entry does not require prefix, pass ""
//   - Increment and Decimal for entry input control
func NewDeroEntry(prefix string, increm float64, decim uint) *DeroAmts {
	entry := &DeroAmts{}
	entry.ExtendBaseWidget(entry)
	entry.AllowFloat = true
	entry.Prefix = prefix
	entry.Increment = increm
	entry.Decimal = decim

	return entry
}

// Accepts whole number or '.'
func (e *DeroAmts) TypedRune(r rune) {
	if r >= '0' && r <= '9' {
		e.Entry.TypedRune(r)
		return
	}

	if e.AllowFloat && r == '.' {
		e.Entry.TypedRune(r)
	}
}

// Increase or decrease of Increment on TypedKey
func (e *DeroAmts) TypedKey(k *fyne.KeyEvent) {
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

type DeroRpcEntries struct {
	Container      *fyne.Container
	Daemon         *widget.SelectEntry
	Wallet         *widget.SelectEntry
	Auth           *widget.Entry
	Balance        *canvas.Text
	Button         *widget.Button
	Disconnect     *widget.Check
	Offset         int
	layout         int
	default_daemon []string
}

// Create a horizontal layout container with entries for daemon, wallet and user:pass
//   - Objects bound to dReams rpc Daemon and Wallet vars with disconnect control
//   - Balance canvas to display wallet balance
//   - Button for OnTapped func()
//   - Offset of 1 puts entries on trailing edge
func NewHorizontalEntries(tag string, offset int) *DeroRpcEntries {
	default_daemon := []string{
		"",
		rpc.DAEMON_RPC_DEFAULT,
		rpc.DAEMON_RPC_REMOTE1,
		rpc.DAEMON_RPC_REMOTE2,
		rpc.DAEMON_RPC_REMOTE5,
		rpc.DAEMON_RPC_REMOTE6,
	}
	daemon_entry := widget.NewSelectEntry(default_daemon)
	daemon_entry.SetPlaceHolder("Daemon RPC:")
	this_daemon := binding.BindString(&rpc.Daemon.Rpc)
	daemon_entry.Bind(this_daemon)

	default_wallet := []string{"127.0.0.1:10103"}
	wallet_entry := widget.NewSelectEntry(default_wallet)
	wallet_entry.SetPlaceHolder("Wallet RPC:")
	this_wallet := binding.BindString(&rpc.Wallet.RPC.Port)
	wallet_entry.Bind(this_wallet)
	wallet_entry.OnChanged = func(s string) {
		if rpc.Wallet.IsConnected() {
			rpc.Wallet.Connected(false)
			rpc.Wallet.CloseConnections(tag)
		}
	}

	pass_entry := widget.NewPasswordEntry()
	pass_entry.SetPlaceHolder("RPC user:pass")
	this_auth := binding.BindString(&rpc.Wallet.RPC.Auth)
	pass_entry.Bind(this_auth)
	pass_entry.OnChanged = func(s string) {
		if rpc.Wallet.IsConnected() {
			rpc.Wallet.Connected(false)
			rpc.Wallet.CloseConnections(tag)
		}
	}

	button := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "confirm"), nil)
	balance := *canvas.NewText(fmt.Sprintf("Balance: %.5f Dero", 0.0), bundle.TextColor)

	control_check := *widget.NewCheck("", nil)
	control_check.Disable()
	control_check.Hide()

	rpc_entry_box := container.NewAdaptiveGrid(3, daemon_entry, wallet_entry, pass_entry)
	rpc_cont := container.NewBorder(nil, nil, &control_check, button, rpc_entry_box)

	d := &DeroRpcEntries{
		Container:      &fyne.Container{},
		Daemon:         daemon_entry,
		Wallet:         wallet_entry,
		Auth:           pass_entry,
		Balance:        &balance,
		Button:         button,
		Disconnect:     &control_check,
		Offset:         offset,
		default_daemon: default_daemon,
		layout:         0,
	}

	if offset == 1 {
		d.Container = container.NewAdaptiveGrid(2, container.NewHBox(layout.NewSpacer(), &balance), rpc_cont)
	} else {
		d.Container = container.NewAdaptiveGrid(2, rpc_cont, container.NewHBox(&balance))
	}

	return d
}

// Create a vertical layout container with entries for daemon, wallet and user:pass
//   - Objects bound to dReams rpc Daemon and Wallet vars with disconnect control
//   - Balance canvas to display wallet balance
//   - Button for OnTapped func()
//   - Offset of 1 puts entries on top edge
func NewVerticalEntries(tag string, offset int) *DeroRpcEntries {
	default_daemon := []string{
		"",
		rpc.DAEMON_RPC_DEFAULT,
		rpc.DAEMON_RPC_REMOTE1,
		rpc.DAEMON_RPC_REMOTE2,
		rpc.DAEMON_RPC_REMOTE5,
		rpc.DAEMON_RPC_REMOTE6,
	}

	daemon_entry := widget.NewSelectEntry(default_daemon)
	daemon_entry.SetPlaceHolder("Daemon RPC:")
	this_daemon := binding.BindString(&rpc.Daemon.Rpc)
	daemon_entry.Bind(this_daemon)

	default_wallet := []string{"127.0.0.1:10103"}
	wallet_entry := widget.NewSelectEntry(default_wallet)
	wallet_entry.SetPlaceHolder("Wallet RPC:")
	this_wallet := binding.BindString(&rpc.Wallet.RPC.Port)
	wallet_entry.Bind(this_wallet)
	wallet_entry.OnChanged = func(s string) {
		if rpc.Wallet.IsConnected() {
			rpc.Wallet.Connected(false)
			rpc.Wallet.CloseConnections(tag)
		}
	}

	pass_entry := widget.NewPasswordEntry()
	pass_entry.SetPlaceHolder("RPC user:pass")
	this_auth := binding.BindString(&rpc.Wallet.RPC.Auth)
	pass_entry.Bind(this_auth)
	pass_entry.OnChanged = func(s string) {
		if rpc.Wallet.IsConnected() {
			rpc.Wallet.Connected(false)
			rpc.Wallet.CloseConnections(tag)
		}
	}

	button := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "confirm"), nil)
	balance := *canvas.NewText(fmt.Sprintf("Balance: %.5f Dero", 0.0), bundle.TextColor)

	control_check := *widget.NewCheck("", nil)
	control_check.Disable()
	control_check.Hide()

	d := &DeroRpcEntries{
		Container:      &fyne.Container{},
		Daemon:         daemon_entry,
		Wallet:         wallet_entry,
		Auth:           pass_entry,
		Balance:        &balance,
		Button:         button,
		Disconnect:     &control_check,
		Offset:         offset,
		default_daemon: default_daemon,
		layout:         1,
	}

	if offset == 1 {
		d.Container = container.NewVBox(
			daemon_entry,
			wallet_entry,
			pass_entry,
			&balance,
			button,
			&control_check)
	} else {
		d.Container = container.NewVBox(
			&control_check,
			daemon_entry,
			wallet_entry,
			pass_entry,
			&balance,
			button)
	}

	return d
}

// Refresh Balance of DeroRpcEntries
//   - Gets balance from rpc.Wallet.Balance
func (d *DeroRpcEntries) RefreshBalance() {
	d.Balance.Text = (fmt.Sprintf("Balance: %.5f DERO", float64(rpc.Wallet.Balance("DERO"))/100000))
	d.Balance.Refresh()
}

// Add new options to default daemon rpc entry
func (d *DeroRpcEntries) AddDaemonOptions(new_opts []string) {
	current := d.default_daemon
	d.Daemon.SetOptions(append(current, new_opts...))
	d.Daemon.Refresh()
}

// Add canvas object indicators to DeroRpcEntries, switching for layout and Offset
func (d *DeroRpcEntries) AddIndicator(ind fyne.CanvasObject) {
	switch d.layout {
	case 0:
		if d.Offset == 1 {
			d.Container.Objects[0].(*fyne.Container).Add(ind)
			return
		}
		d.Container.Objects[1].(*fyne.Container).Add(ind)
	case 1:
		d.Container.Add(ind)
	default:
		// nothing
	}
}
