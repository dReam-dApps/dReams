package menu

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dReam-dApps/dReams/rpc"
)

type dispayObjects struct {
	Price  *widget.Entry
	Height struct {
		Wallet *widget.Entry
		Daemon *widget.Entry
		Gnomes *widget.Entry
	}
	Indexed *canvas.Text
	Status  *canvas.Text
}

var Info dispayObjects

func (i *dispayObjects) SetStatus(s string) {
	if i.Status != nil {
		i.Status.Text = s
		i.Status.Refresh()
	}
}

// Refresh Gnomon height display
func (i *dispayObjects) RefreshGnomon() {
	if gnomon.IsRunning() {
		i.Height.Gnomes.Text = fmt.Sprintf("%d", gnomon.GetLastHeight())
	} else {
		i.Height.Gnomes.Text = "0"

	}
	i.Height.Gnomes.Refresh()
}

// Refresh Gnomon indexed scids display
func (i *dispayObjects) RefreshIndexed() {
	if gnomon.IsRunning() {
		i.Indexed.Text = fmt.Sprintf("Indexed SCIDs: %d", gnomon.IndexCount())
	} else {
		i.Indexed.Text = "Indexed SCIDs: 0"

	}
	i.Indexed.Refresh()
}

// Refresh daemon height display
func (i *dispayObjects) RefreshDaemon(tag string) {
	if rpc.Daemon.IsConnected() {
		height := rpc.DaemonHeight(tag, rpc.Daemon.Rpc)
		i.Height.Daemon.Text = fmt.Sprintf("%d", height)
	} else {
		i.Height.Daemon.Text = "0"

	}
	i.Height.Daemon.Refresh()
}

// Refresh current Dero-USDT price
func (i *dispayObjects) RefreshPrice(tag string) {
	if rpc.Daemon.IsConnected() {
		_, price := GetPrice("DERO-USDT", tag)
		i.Price.Text = fmt.Sprintf("$%s", price)
	} else {
		i.Price.Text = "$"
	}
	i.Price.Refresh()
}

// Refresh wallet height display
func (i *dispayObjects) RefreshWallet() {
	if rpc.Wallet.IsConnected() {
		i.Height.Wallet.Text = fmt.Sprintf("%d", rpc.Wallet.Height)
	} else {
		i.Height.Wallet.Text = "0"
	}
	i.Height.Wallet.Refresh()
}

// Set wallet and chain display content for menu
func InfoDisplay() fyne.CanvasObject {
	Info.Status = canvas.NewText("", color.RGBA{31, 150, 200, 210})
	Info.Height.Gnomes = widget.NewEntry()
	Info.Height.Daemon = widget.NewEntry()
	Info.Height.Wallet = widget.NewEntry()
	Info.Price = widget.NewEntry()

	Info.Height.Gnomes.Disable()
	Info.Height.Daemon.Disable()
	Info.Height.Wallet.Disable()
	Info.Price.Disable()

	Info.Status.TextSize = 18
	// Info.Height.Gnomes.TextSize = 18
	// Info.Height.Daemon.TextSize = 18
	// Info.Height.Wallet.TextSize = 18
	// Info.Price.TextSize = 18

	Info.Status.Alignment = fyne.TextAlignCenter
	// Info.Height.Gnomes.Alignment = fyne.TextAlignCenter
	// Info.Height.Daemon.Alignment = fyne.TextAlignCenter
	// Info.Height.Wallet.Alignment = fyne.TextAlignCenter
	// Info.Price.Alignment = fyne.TextAlignCenter

	info_form := []*widget.FormItem{}
	info_form = append(info_form, widget.NewFormItem("", Info.Status))
	info_form = append(info_form, widget.NewFormItem("Gnomon Height", Info.Height.Gnomes))
	info_form = append(info_form, widget.NewFormItem("Daemon Height", Info.Height.Daemon))
	info_form = append(info_form, widget.NewFormItem("Wallet Height", Info.Height.Wallet))
	info_form = append(info_form, widget.NewFormItem("Price", Info.Price))

	return container.NewVBox(widget.NewForm(info_form...))
}
