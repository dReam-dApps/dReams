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
	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/rpc"
)

// Main dReams dApp content struct
type DreamsItems struct {
	LeftLabel  *widget.Label
	RightLabel *widget.Label
	TopLabel   *widget.Label

	Back    fyne.Container
	Front   fyne.Container
	Actions fyne.Container
	DApp    *fyne.Container
}

type TenthAmt struct {
	xwidget.NumericalEntry
	Prefix string
}

// Create new numerical entry with change of 0.1 on up or down key stroke
//   - If entry does not require prefix, pass ""
func TenthAmtEntry(prefix string) *TenthAmt {
	entry := &TenthAmt{}
	entry.ExtendBaseWidget(entry)
	entry.AllowFloat = true
	entry.Prefix = prefix

	return entry
}

// Accepts whole number or '.'
func (e *TenthAmt) TypedRune(r rune) {
	if r >= '0' && r <= '9' {
		e.Entry.TypedRune(r)
		return
	}

	if e.AllowFloat && r == '.' {
		e.Entry.TypedRune(r)
	}
}

// Increment of 0.1 on TypedKey
func (e *TenthAmt) TypedKey(k *fyne.KeyEvent) {
	value := strings.Trim(e.Entry.Text, e.Prefix)
	switch k.Name {
	case fyne.KeyUp:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			e.Entry.SetText(e.Prefix + strconv.FormatFloat(float64(f+0.1), 'f', 1, 64))
		}
	case fyne.KeyDown:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			if f >= 0.1 {
				e.Entry.SetText(e.Prefix + strconv.FormatFloat(float64(f-0.1), 'f', 1, 64))
			}
		}
	}
	e.Entry.TypedKey(k)
}

type WholeAmt struct {
	xwidget.NumericalEntry
	Prefix string
}

// Create new numerical entry with change of 1 on up or down key stroke
//   - If entry does not require prefix, pass ""
func WholeAmtEntry(prefix string) *WholeAmt {
	entry := &WholeAmt{}
	entry.ExtendBaseWidget(entry)
	entry.Prefix = prefix

	return entry
}

// Only accept whole number
func (e *WholeAmt) TypedRune(r rune) {
	if r >= '0' && r <= '9' {
		e.Entry.TypedRune(r)
		return
	}
}

// Increment of 1 on TypedKey
func (e *WholeAmt) TypedKey(k *fyne.KeyEvent) {
	value := strings.Trim(e.Entry.Text, e.Prefix)
	switch k.Name {
	case fyne.KeyUp:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			e.Entry.SetText(e.Prefix + strconv.FormatFloat(float64(f+1), 'f', 0, 64))
		}
	case fyne.KeyDown:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			if f >= 0.1 {
				e.Entry.SetText(e.Prefix + strconv.FormatFloat(float64(f-1), 'f', 0, 64))
			}
		}
	}
	e.Entry.TypedKey(k)
}

type DeroRpcEntries struct {
	Container  *fyne.Container
	Daemon     *widget.SelectEntry
	Wallet     *widget.SelectEntry
	Auth       *widget.Entry
	Balance    *canvas.Text
	Button     *widget.Button
	Disconnect *widget.Check
	Offset     int
}

// Horizontal layout with daemon, wallet and user:pass entries
//   - Objects bound to dReams rpc Deamon and Wallet vars with disconnect control
//   - Balance canvas to display wallet balance
//   - Button for OnTapped func()
//   - Offset of 1 puts entries on trailing edge
func HorizontalEntries(tag string, offset int) *DeroRpcEntries {
	default_daemon := []string{"", rpc.DAEMON_RPC_DEFAULT, rpc.DAEMON_RPC_REMOTE5, rpc.DAEMON_RPC_REMOTE6}
	daemon_entry := widget.NewSelectEntry(default_daemon)
	daemon_entry.SetPlaceHolder("Daemon RPC:")
	this_daemon := binding.BindString(&rpc.Daemon.Rpc)
	daemon_entry.Bind(this_daemon)

	default_wallet := []string{"127.0.0.1:10103"}
	wallet_entry := widget.NewSelectEntry(default_wallet)
	wallet_entry.SetPlaceHolder("Wallet RPC:")
	this_wallet := binding.BindString(&rpc.Wallet.Rpc)
	wallet_entry.Bind(this_wallet)
	wallet_entry.OnCursorChanged = func() {
		if rpc.Wallet.Connect {
			rpc.Wallet.Address = ""
			rpc.Wallet.Height = 0
			rpc.Wallet.Connect = false
		}
	}

	pass_entry := widget.NewPasswordEntry()
	pass_entry.SetPlaceHolder("RPC user:pass")
	this_auth := binding.BindString(&rpc.Wallet.UserPass)
	pass_entry.Bind(this_auth)
	pass_entry.OnCursorChanged = func() {
		if rpc.Wallet.Connect {
			rpc.GetAddress(tag)
			if !rpc.Wallet.Connect {
				rpc.Wallet.Address = ""
				rpc.Wallet.Height = 0
			}
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
		Container:  &fyne.Container{},
		Daemon:     daemon_entry,
		Wallet:     wallet_entry,
		Auth:       pass_entry,
		Balance:    &balance,
		Button:     button,
		Disconnect: &control_check,
		Offset:     offset,
	}

	if offset == 1 {
		d.Container = container.NewAdaptiveGrid(2, container.NewAdaptiveGrid(2, layout.NewSpacer(), &balance), rpc_cont)
	} else {
		d.Container = container.NewAdaptiveGrid(2, rpc_cont, container.NewAdaptiveGrid(2, &balance, layout.NewSpacer()))
	}

	return d
}

// Verticle layout with daemon, wallet and user:pass entries
//   - Objects bound to dReams rpc Deamon and Wallet vars with disconnect control
//   - Balance canvas to display wallet balance
//   - Button for OnTapped func()
//   - Offset of 1 puts entries on top edge
func VerticleEntries(tag string, offset int) *DeroRpcEntries {
	default_daemon := []string{"", rpc.DAEMON_RPC_DEFAULT, rpc.DAEMON_RPC_REMOTE5, rpc.DAEMON_RPC_REMOTE6}
	daemon_entry := widget.NewSelectEntry(default_daemon)
	daemon_entry.SetPlaceHolder("Daemon RPC:")
	this_daemon := binding.BindString(&rpc.Daemon.Rpc)
	daemon_entry.Bind(this_daemon)

	default_wallet := []string{"127.0.0.1:10103"}
	wallet_entry := widget.NewSelectEntry(default_wallet)
	wallet_entry.SetPlaceHolder("Wallet RPC:")
	this_wallet := binding.BindString(&rpc.Wallet.Rpc)
	wallet_entry.Bind(this_wallet)
	wallet_entry.OnCursorChanged = func() {
		if rpc.Wallet.Connect {
			rpc.Wallet.Address = ""
			rpc.Wallet.Height = 0
			rpc.Wallet.Connect = false
		}
	}

	pass_entry := widget.NewPasswordEntry()
	pass_entry.SetPlaceHolder("RPC user:pass")
	this_auth := binding.BindString(&rpc.Wallet.UserPass)
	pass_entry.Bind(this_auth)
	pass_entry.OnCursorChanged = func() {
		if rpc.Wallet.Connect {
			rpc.GetAddress(tag)
			if !rpc.Wallet.Connect {
				rpc.Wallet.Address = ""
				rpc.Wallet.Height = 0
			}
		}
	}

	button := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "confirm"), nil)
	balance := *canvas.NewText(fmt.Sprintf("Balance: %.5f Dero", 0.0), bundle.TextColor)

	control_check := *widget.NewCheck("", nil)
	control_check.Disable()
	control_check.Hide()

	d := &DeroRpcEntries{
		Container:  &fyne.Container{},
		Daemon:     daemon_entry,
		Wallet:     wallet_entry,
		Auth:       pass_entry,
		Balance:    &balance,
		Button:     button,
		Disconnect: &control_check,
		Offset:     offset,
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
	d.Balance.Text = (fmt.Sprintf("Balance: %.5f Dero", float64(rpc.Wallet.Balance)/100000))
	d.Balance.Refresh()
}
