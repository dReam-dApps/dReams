package menu

import (
	"image/color"
	"time"

	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/gnomes"
	"github.com/dReam-dApps/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	xwidget "fyne.io/x/fyne/widget"
)

type DreamsIndicator struct {
	Img       *canvas.Image
	Rect      *canvas.Rectangle
	Animation *fyne.Animation
}

// dReams app status indicators for wallet, daemon, Gnomon and services
//   - Pass further DreamsIndicators to add, can be nil
func StartIndicators(add []*DreamsIndicator) fyne.CanvasObject {
	purple := color.RGBA{105, 90, 205, 210}
	blue := color.RGBA{31, 150, 200, 210}
	alpha := color.RGBA{0, 0, 0, 0}

	// Gnomon indicator
	g_top := canvas.NewRectangle(color.Transparent)
	g_top.SetMinSize(fyne.NewSize(57, 10))

	g_bottom := canvas.NewRectangle(color.Transparent)
	g_bottom.SetMinSize(fyne.NewSize(57, 10))

	hover := gnomes.ToolTip(45, nil)

	gnomes.Indicator.Sync = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			init := gnomon.IsInitialized()
			if (init && !gnomon.IsStatus("indexed")) || (init && !gnomon.HasChecked()) {
				g_top.FillColor = c
				canvas.Refresh(g_top)
				g_bottom.FillColor = c
				canvas.Refresh(g_bottom)
			} else {
				g_top.FillColor = alpha
				canvas.Refresh(g_top)
				g_bottom.FillColor = alpha
				canvas.Refresh(g_bottom)
			}
			hover.Refresh()
		})

	gnomes.Indicator.Sync.RepeatCount = fyne.AnimationRepeatForever
	gnomes.Indicator.Sync.AutoReverse = true

	sync_box := container.NewVBox(
		g_top,
		layout.NewSpacer(),
		g_bottom)

	g_full := canvas.NewRectangle(color.Transparent)
	g_full.SetMinSize(fyne.NewSize(57, 36))

	gnomes.Indicator.Full = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if gnomon.IsInitialized() && gnomon.HasIndex(2) && gnomon.HasChecked() && gnomon.IsStatus("indexed") {
				g_full.FillColor = c
				canvas.Refresh(g_full)
				sync_box.Hide()
			} else {
				g_full.FillColor = alpha
				canvas.Refresh(g_full)
				sync_box.Show()
			}
		})

	gnomes.Indicator.Full.RepeatCount = fyne.AnimationRepeatForever
	gnomes.Indicator.Full.AutoReverse = true

	gnomes.Indicator.Icon, _ = xwidget.NewAnimatedGifFromResource(bundle.ResourceGnomonGifGif)
	gnomes.Indicator.Icon.SetMinSize(fyne.NewSize(36, 36))

	// Daemon connection indicator
	d_rect := canvas.NewRectangle(color.Transparent)
	d_rect.SetMinSize(fyne.NewSize(36, 36))

	Control.Indicator.Daemon = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if rpc.Daemon.IsConnected() {
				d_rect.FillColor = c
				canvas.Refresh(d_rect)
			} else {
				d_rect.FillColor = alpha
				canvas.Refresh(d_rect)
			}
		})

	Control.Indicator.Daemon.RepeatCount = fyne.AnimationRepeatForever
	Control.Indicator.Daemon.AutoReverse = true

	// Wallet connection indicator
	w_rect := canvas.NewRectangle(color.Transparent)
	w_rect.SetMinSize(fyne.NewSize(36, 36))

	Control.Indicator.Wallet = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if rpc.Wallet.IsConnected() {
				w_rect.FillColor = c
				canvas.Refresh(w_rect)
			} else {
				w_rect.FillColor = alpha
				canvas.Refresh(w_rect)
			}
		})

	Control.Indicator.Wallet.RepeatCount = fyne.AnimationRepeatForever
	Control.Indicator.Wallet.AutoReverse = true

	d_text := canvas.NewText(" D ", bundle.TextColor)
	d_text.TextStyle.Bold = true
	d_text.Alignment = fyne.TextAlignCenter
	d_text.TextSize = 16

	w_text := canvas.NewText(" W ", bundle.TextColor)
	w_text.TextStyle.Bold = true
	w_text.Alignment = fyne.TextAlignCenter
	w_text.TextSize = 16

	// Tx confirmation indicator
	var c_img *canvas.Image
	if bundle.AppColor == color.White {
		c_img = canvas.NewImageFromResource(bundle.ResourceTxDarkPng)
	} else {
		c_img = canvas.NewImageFromResource(bundle.ResourceTxLightPng)
	}

	c_img.SetMinSize(fyne.NewSize(30, 30))
	c_rect := canvas.NewRectangle(color.Transparent)
	c_rect.SetMinSize(fyne.NewSize(36, 36))

	Control.Indicator.TX = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if rpc.IsConfirmingTx() {
				c_rect.FillColor = c
				c_img.Show()
				canvas.Refresh(c_rect)
			} else {
				c_rect.FillColor = color.Transparent
				c_img.Hide()
				canvas.Refresh(c_rect)
			}
		})

	Control.Indicator.TX.RepeatCount = fyne.AnimationRepeatForever
	Control.Indicator.TX.AutoReverse = true

	connect_box := container.NewHBox(
		container.NewStack(c_rect, container.NewCenter(c_img)),
		container.NewStack(d_rect, container.NewCenter(d_text)),
		container.NewStack(w_rect, container.NewCenter(w_text)))

	additional_inds := container.NewHBox()
	for _, ind := range add {
		additional_inds.Add(container.NewStack(ind.Rect, container.NewCenter(ind.Img)))
	}

	top_box := container.NewHBox(layout.NewSpacer(), additional_inds, connect_box, container.NewStack(g_full, sync_box, gnomes.Indicator.Icon, hover))
	place := container.NewVBox(top_box, layout.NewSpacer())

	go func() {
		gnomes.Indicator.Sync.Start()
		gnomes.Indicator.Full.Start()
		gnomes.Indicator.Icon.Start()
		Control.Indicator.Daemon.Start()
		Control.Indicator.Wallet.Start()
		Control.Indicator.TX.Start()
		for _, ind := range add {
			ind.Animation.Start()
		}
		time.Sleep(time.Second)
		hover.Canvas = fyne.CurrentApp().Driver().CanvasForObject(gnomes.Indicator.Icon)
	}()

	return container.NewStack(place)
}

// Stop dReams app status indicators
func StopIndicators(these []*DreamsIndicator) {
	gnomes.Indicator.Sync.Stop()
	gnomes.Indicator.Full.Stop()
	Control.Indicator.Daemon.Stop()
	Control.Indicator.Wallet.Stop()
	Control.Indicator.TX.Stop()
	for _, ind := range these {
		ind.Animation.Stop()
	}
	if gnomes.Indicator.Icon != nil {
		gnomes.Indicator.Icon.Stop()
	}
}

// Main gif seems to stop when hidden for 5min+
// will use this for now to check if running and restart
func RestartGif(g *xwidget.AnimatedGif) {
	if g != nil {
		g.Start()
	}
}
