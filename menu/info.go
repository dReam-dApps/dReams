package menu

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/rpc"
)

type dispayObjects struct {
	Price  *canvas.Text
	Height struct {
		Wallet *canvas.Text
		Daemon *canvas.Text
		Gnomes *canvas.Text
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
	if Gnomes.IsRunning() {
		i.Height.Gnomes.Text = fmt.Sprintf("Gnomon Height: %d", Gnomes.Indexer.LastIndexedHeight)
	} else {
		i.Height.Gnomes.Text = "Gnomon Height: 0"

	}
	i.Height.Gnomes.Refresh()
}

// Refresh Gnomon indexed scids display
func (i *dispayObjects) RefreshIndexed() {
	if Gnomes.IsRunning() {
		i.Indexed.Text = fmt.Sprintf("Indexed SCIDs: %d", Gnomes.IndexCount())
	} else {
		i.Indexed.Text = "Indexed SCIDs: 0"

	}
	i.Indexed.Refresh()
}

// Refresh daemon height display
func (i *dispayObjects) RefreshDaemon(tag string) {
	if rpc.Daemon.IsConnected() {
		height := rpc.DaemonHeight(tag, rpc.Daemon.Rpc)
		i.Height.Daemon.Text = fmt.Sprintf("Daemon Height: %d", height)
	} else {
		i.Height.Daemon.Text = "Daemon Height: 0"

	}
	i.Height.Daemon.Refresh()
}

// Refresh current Dero-USDT price
func (i *dispayObjects) RefreshPrice(tag string) {
	if rpc.Daemon.IsConnected() {
		_, price := GetPrice("DERO-USDT", tag)
		i.Price.Text = fmt.Sprintf("Dero Price: $%s", price)
	} else {
		i.Price.Text = "Dero Price: $"
	}
	i.Price.Refresh()
}

// Refresh wallet height display
func (i *dispayObjects) RefreshWallet() {
	if rpc.Wallet.IsConnected() {
		i.Height.Wallet.Text = fmt.Sprintf("Wallet Height: %s", rpc.Wallet.Display.Height)
	} else {
		i.Height.Wallet.Text = " Wallet Height: 0"
	}
	i.Height.Wallet.Refresh()
}

// Set wallet and chain display content for menu
func InfoDisplay() fyne.CanvasObject {
	Info.Status = canvas.NewText("", color.RGBA{31, 150, 200, 210})
	Info.Height.Gnomes = canvas.NewText("Gnomon Height: ", bundle.TextColor)
	Info.Height.Daemon = canvas.NewText("Daemon Height: ", bundle.TextColor)
	Info.Height.Wallet = canvas.NewText("Wallet Height: ", bundle.TextColor)
	Info.Price = canvas.NewText("Dero Price: $", bundle.TextColor)

	Info.Status.TextSize = 18
	Info.Height.Gnomes.TextSize = 18
	Info.Height.Daemon.TextSize = 18
	Info.Height.Wallet.TextSize = 18
	Info.Price.TextSize = 18

	Info.Status.Alignment = fyne.TextAlignCenter
	Info.Height.Gnomes.Alignment = fyne.TextAlignCenter
	Info.Height.Daemon.Alignment = fyne.TextAlignCenter
	Info.Height.Wallet.Alignment = fyne.TextAlignCenter
	Info.Price.Alignment = fyne.TextAlignCenter

	return container.NewVBox(
		Info.Status,
		Info.Height.Gnomes,
		Info.Height.Daemon,
		Info.Height.Wallet,
		Info.Price)
}
