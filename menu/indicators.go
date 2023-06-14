package menu

import (
	"image/color"
	"time"

	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	xwidget "fyne.io/x/fyne/widget"
)

// Menu label when Gnomon starting
func startLabel() {
	Assets.Gnomes_sync.Text = (" Starting Gnomon")
	Assets.Gnomes_sync.Refresh()
}

// Menu label when Gnomon scans wallet
func checkLabel() {
	Assets.Gnomes_sync.Text = (" Checking for Assets")
	Assets.Gnomes_sync.Refresh()
}

// Menu label when Gnomon is closing
func StopLabel() {
	if Assets.Gnomes_sync != nil {
		Assets.Gnomes_sync.Text = (" Putting Gnomon to Sleep")
		Assets.Gnomes_sync.Refresh()
	}
}

// Menu label when Gnomon is not running
func SleepLabel() {
	Assets.Gnomes_sync.Text = (" Gnomon is Sleeping")
	Assets.Gnomes_sync.Refresh()
}

// dReams app status indicators for wallet, daemon, Gnomon and services
func StartDreamsIndicators() fyne.CanvasObject {
	purple := color.RGBA{105, 90, 205, 210}
	blue := color.RGBA{31, 150, 200, 210}
	alpha := color.RGBA{0, 0, 0, 0}

	g_top := canvas.NewRectangle(color.Black)
	g_top.SetMinSize(fyne.NewSize(57, 10))

	g_bottom := canvas.NewRectangle(color.Black)
	g_bottom.SetMinSize(fyne.NewSize(57, 10))

	Gnomes.Sync_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if Gnomes.IsInitialized() && !Gnomes.HasChecked() {
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
		})

	Gnomes.Sync_ind.RepeatCount = fyne.AnimationRepeatForever
	Gnomes.Sync_ind.AutoReverse = true

	sync_box := container.NewVBox(
		g_top,
		layout.NewSpacer(),
		g_bottom)

	g_full := canvas.NewRectangle(color.Black)
	g_full.SetMinSize(fyne.NewSize(57, 36))

	Gnomes.Full_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if Gnomes.IsInitialized() && Gnomes.HasIndex(2) && Gnomes.HasChecked() {
				g_full.FillColor = c
				canvas.Refresh(g_full)
				sync_box.Hide()
			} else {
				g_full.FillColor = alpha
				canvas.Refresh(g_full)
				sync_box.Show()
			}
		})

	Gnomes.Full_ind.RepeatCount = fyne.AnimationRepeatForever
	Gnomes.Full_ind.AutoReverse = true

	Gnomes.Icon_ind, _ = xwidget.NewAnimatedGifFromResource(bundle.ResourceGnomonGifGif)
	Gnomes.Icon_ind.SetMinSize(fyne.NewSize(36, 36))

	d_rect := canvas.NewRectangle(color.Black)
	d_rect.SetMinSize(fyne.NewSize(36, 36))

	Control.Daemon_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if rpc.Daemon.Connect {
				d_rect.FillColor = c
				canvas.Refresh(d_rect)
			} else {
				d_rect.FillColor = alpha
				canvas.Refresh(d_rect)
			}
		})

	Control.Daemon_ind.RepeatCount = fyne.AnimationRepeatForever
	Control.Daemon_ind.AutoReverse = true

	w_rect := canvas.NewRectangle(color.Black)
	w_rect.SetMinSize(fyne.NewSize(36, 36))

	Control.Wallet_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if rpc.Wallet.IsConnected() {
				w_rect.FillColor = c
				canvas.Refresh(w_rect)
			} else {
				w_rect.FillColor = alpha
				canvas.Refresh(w_rect)
			}
		})

	Control.Wallet_ind.RepeatCount = fyne.AnimationRepeatForever
	Control.Wallet_ind.AutoReverse = true

	d := canvas.NewText(" D ", bundle.TextColor)
	d.TextStyle.Bold = true
	d.Alignment = fyne.TextAlignCenter
	d.TextSize = 16
	w := canvas.NewText(" W ", bundle.TextColor)
	w.TextStyle.Bold = true
	w.Alignment = fyne.TextAlignCenter
	w.TextSize = 16

	connect_box := container.NewHBox(
		container.NewMax(d_rect, container.NewCenter(d)),
		container.NewMax(w_rect, container.NewCenter(w)))

	pbot := canvas.NewImageFromResource(bundle.ResourcePokerBotIconPng)
	pbot.SetMinSize(fyne.NewSize(30, 30))
	p_rect := canvas.NewRectangle(alpha)
	p_rect.SetMinSize(fyne.NewSize(36, 36))

	dService := canvas.NewImageFromResource(bundle.ResourceDReamServiceIconPng)
	dService.SetMinSize(fyne.NewSize(30, 30))
	s_rect := canvas.NewRectangle(alpha)
	s_rect.SetMinSize(fyne.NewSize(36, 36))

	service_box := container.NewHBox(
		container.NewMax(p_rect, container.NewCenter(pbot)),
		container.NewMax(s_rect, container.NewCenter(dService)))

	Control.Poker_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if rpc.Odds.Run {
				p_rect.FillColor = c
				pbot.Show()
				canvas.Refresh(p_rect)
			} else {
				p_rect.FillColor = alpha
				pbot.Hide()
				canvas.Refresh(p_rect)
			}
		})

	Control.Service_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if rpc.Wallet.Service {
				s_rect.FillColor = c
				dService.Show()
				canvas.Refresh(s_rect)
			} else {
				s_rect.FillColor = alpha
				dService.Hide()
				canvas.Refresh(s_rect)
			}
		})

	Control.Poker_ind.RepeatCount = fyne.AnimationRepeatForever
	Control.Poker_ind.AutoReverse = true

	Control.Service_ind.RepeatCount = fyne.AnimationRepeatForever
	Control.Service_ind.AutoReverse = true

	top_box := container.NewHBox(layout.NewSpacer(), service_box, connect_box, container.NewMax(g_full, sync_box, Gnomes.Icon_ind))
	place := container.NewVBox(top_box, layout.NewSpacer())

	go func() {
		Gnomes.Sync_ind.Start()
		Gnomes.Full_ind.Start()
		Gnomes.Icon_ind.Start()
		Control.Daemon_ind.Start()
		Control.Wallet_ind.Start()
		Control.Poker_ind.Start()
		Control.Service_ind.Start()
	}()

	return container.NewMax(place)
}

// Dero status indicators for wallet, daemon and Gnomon
func StartIndicators() fyne.CanvasObject {
	purple := color.RGBA{105, 90, 205, 210}
	blue := color.RGBA{31, 150, 200, 210}
	alpha := color.RGBA{0, 0, 0, 0}

	g_top := canvas.NewRectangle(color.Black)
	g_top.SetMinSize(fyne.NewSize(57, 10))

	g_bottom := canvas.NewRectangle(color.Black)
	g_bottom.SetMinSize(fyne.NewSize(57, 10))

	Gnomes.Sync_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if Gnomes.IsInitialized() && !Gnomes.HasChecked() {
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
		})

	Gnomes.Sync_ind.RepeatCount = fyne.AnimationRepeatForever
	Gnomes.Sync_ind.AutoReverse = true

	sync_box := container.NewVBox(
		g_top,
		layout.NewSpacer(),
		g_bottom)

	g_full := canvas.NewRectangle(color.Black)
	g_full.SetMinSize(fyne.NewSize(57, 36))

	Gnomes.Full_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if Gnomes.IsInitialized() && Gnomes.HasIndex(1) && Gnomes.HasChecked() {
				g_full.FillColor = c
				canvas.Refresh(g_full)
				sync_box.Hide()
			} else {
				g_full.FillColor = alpha
				canvas.Refresh(g_full)
				sync_box.Show()
			}
		})

	Gnomes.Full_ind.RepeatCount = fyne.AnimationRepeatForever
	Gnomes.Full_ind.AutoReverse = true

	Gnomes.Icon_ind, _ = xwidget.NewAnimatedGifFromResource(bundle.ResourceGnomonGifGif)
	Gnomes.Icon_ind.SetMinSize(fyne.NewSize(36, 36))

	d_rect := canvas.NewRectangle(color.Black)
	d_rect.SetMinSize(fyne.NewSize(36, 36))

	Control.Daemon_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if rpc.Daemon.Connect {
				d_rect.FillColor = c
				canvas.Refresh(d_rect)
			} else {
				d_rect.FillColor = alpha
				canvas.Refresh(d_rect)
			}
		})

	Control.Daemon_ind.RepeatCount = fyne.AnimationRepeatForever
	Control.Daemon_ind.AutoReverse = true

	w_rect := canvas.NewRectangle(color.Black)
	w_rect.SetMinSize(fyne.NewSize(36, 36))

	Control.Wallet_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if rpc.Wallet.IsConnected() {
				w_rect.FillColor = c
				canvas.Refresh(w_rect)
			} else {
				w_rect.FillColor = alpha
				canvas.Refresh(w_rect)
			}
		})

	Control.Wallet_ind.RepeatCount = fyne.AnimationRepeatForever
	Control.Wallet_ind.AutoReverse = true

	d := canvas.NewText(" D ", bundle.TextColor)
	d.TextStyle.Bold = true
	d.Alignment = fyne.TextAlignCenter
	d.TextSize = 16
	w := canvas.NewText(" W ", bundle.TextColor)
	w.TextStyle.Bold = true
	w.Alignment = fyne.TextAlignCenter
	w.TextSize = 16

	connect_box := container.NewHBox(
		container.NewMax(d_rect, container.NewCenter(d)),
		container.NewMax(w_rect, container.NewCenter(w)))

	top_box := container.NewHBox(layout.NewSpacer(), connect_box, container.NewMax(g_full, sync_box, Gnomes.Icon_ind))
	place := container.NewVBox(top_box, layout.NewSpacer())

	go func() {
		Gnomes.Sync_ind.Start()
		Gnomes.Full_ind.Start()
		Gnomes.Icon_ind.Start()
		Control.Daemon_ind.Start()
		Control.Wallet_ind.Start()
	}()

	return container.NewMax(place)
}

// Stop dReams app status indicators
func StopIndicators() {
	Gnomes.Sync_ind.Stop()
	Gnomes.Full_ind.Stop()
	Control.Daemon_ind.Stop()
	Control.Wallet_ind.Stop()
	if Control.Poker_ind != nil {
		Control.Poker_ind.Stop()
	}
	if Control.Service_ind != nil {
		Control.Service_ind.Stop()
	}
	if Gnomes.Icon_ind != nil {
		Gnomes.Icon_ind.Stop()
	}
}

// Main gif seems to stop when hidden for 5min+
// will use this for now to check if running and restart
func RestartGif(g *xwidget.AnimatedGif) {
	if g != nil {
		g.Start()
	}
}
