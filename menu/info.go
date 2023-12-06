package menu

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/dReam-dApps/dReams/bundle"
)

// Set wallet and chain display content for menu
func MenuDisplay() fyne.CanvasObject {
	Assets.Gnomes_sync = canvas.NewText("", color.RGBA{31, 150, 200, 210})
	Assets.Gnomes_height = canvas.NewText(" Gnomon Height: ", bundle.TextColor)
	Assets.Daem_height = canvas.NewText(" Daemon Height: ", bundle.TextColor)
	Assets.Wall_height = canvas.NewText(" Wallet Height: ", bundle.TextColor)
	Assets.Dero_price = canvas.NewText(" Dero Price: $", bundle.TextColor)

	Assets.Gnomes_sync.TextSize = 18
	Assets.Gnomes_height.TextSize = 18
	Assets.Daem_height.TextSize = 18
	Assets.Wall_height.TextSize = 18
	Assets.Dero_price.TextSize = 18

	Assets.Gnomes_sync.Alignment = fyne.TextAlignCenter
	Assets.Gnomes_height.Alignment = fyne.TextAlignCenter
	Assets.Daem_height.Alignment = fyne.TextAlignCenter
	Assets.Wall_height.Alignment = fyne.TextAlignCenter
	Assets.Dero_price.Alignment = fyne.TextAlignCenter

	return container.NewVBox(
		Assets.Gnomes_sync,
		Assets.Gnomes_height,
		Assets.Daem_height,
		Assets.Wall_height,
		Assets.Dero_price)
}
