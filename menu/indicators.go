package menu

import (
	"fmt"
	"image/color"
	"time"

	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
)

type DreamsIndicator struct {
	Img       *canvas.Image
	Rect      *canvas.Rectangle
	Animation *fyne.Animation
}

// dReams app status indicators for wallet, daemon, Gnomon and services
//   - Pass further DreamsIndicators to add
func StartDreamsIndicators(add []DreamsIndicator) fyne.CanvasObject {
	purple := color.RGBA{105, 90, 205, 210}
	blue := color.RGBA{31, 150, 200, 210}
	alpha := color.RGBA{0, 0, 0, 0}

	g_top := canvas.NewRectangle(color.Black)
	g_top.SetMinSize(fyne.NewSize(57, 10))

	g_bottom := canvas.NewRectangle(color.Black)
	g_bottom.SetMinSize(fyne.NewSize(57, 10))

	hover := gnomonToolTip(45, nil)

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
			hover.Refresh()
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
			if rpc.Daemon.IsConnected() {
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
		container.NewStack(d_rect, container.NewCenter(d)),
		container.NewStack(w_rect, container.NewCenter(w)))

	additional_inds := container.NewHBox()
	for _, ind := range add {
		additional_inds.Add(container.NewStack(ind.Rect, container.NewCenter(ind.Img)))
	}

	top_box := container.NewHBox(layout.NewSpacer(), additional_inds, connect_box, container.NewStack(g_full, sync_box, Gnomes.Icon_ind, hover))
	place := container.NewVBox(top_box, layout.NewSpacer())

	go func() {
		Gnomes.Sync_ind.Start()
		Gnomes.Full_ind.Start()
		Gnomes.Icon_ind.Start()
		Control.Daemon_ind.Start()
		Control.Wallet_ind.Start()
		for _, ind := range add {
			ind.Animation.Start()
		}
		time.Sleep(time.Second)
		hover.canvas = fyne.CurrentApp().Driver().CanvasForObject(Gnomes.Icon_ind)
	}()

	return container.NewStack(place)
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

	hover := gnomonToolTip(-45, nil)

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
			hover.Refresh()
		})

	Gnomes.Full_ind.RepeatCount = fyne.AnimationRepeatForever
	Gnomes.Full_ind.AutoReverse = true

	Gnomes.Icon_ind, _ = xwidget.NewAnimatedGifFromResource(bundle.ResourceGnomonGifGif)
	Gnomes.Icon_ind.SetMinSize(fyne.NewSize(36, 36))

	d_rect := canvas.NewRectangle(color.Black)
	d_rect.SetMinSize(fyne.NewSize(36, 36))

	Control.Daemon_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if rpc.Daemon.IsConnected() {
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
		container.NewStack(d_rect, container.NewCenter(d)),
		container.NewStack(w_rect, container.NewCenter(w)))

	top_box := container.NewHBox(layout.NewSpacer(), connect_box, container.NewStack(g_full, sync_box, Gnomes.Icon_ind, hover))
	place := container.NewVBox(top_box, layout.NewSpacer())

	go func() {
		Gnomes.Sync_ind.Start()
		Gnomes.Full_ind.Start()
		Gnomes.Icon_ind.Start()
		Control.Daemon_ind.Start()
		Control.Wallet_ind.Start()
		time.Sleep(time.Second)
		hover.canvas = fyne.CurrentApp().Driver().CanvasForObject(Gnomes.Icon_ind)
	}()

	return container.NewStack(place)
}

// Stop dReams app status indicators
func StopIndicators(these []DreamsIndicator) {
	Gnomes.Sync_ind.Stop()
	Gnomes.Full_ind.Stop()
	Control.Daemon_ind.Stop()
	Control.Wallet_ind.Stop()
	for _, ind := range these {
		ind.Animation.Stop()
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

var _ desktop.Hoverable = (*toolTip)(nil)

type toolTip struct {
	hovered bool
	text    *canvas.Text
	canvas  fyne.Canvas
	popup   *widget.PopUp
	offset  float32
	*canvas.Rectangle
}

// Display Gnomon heights when hovered
func gnomonToolTip(offset float32, can fyne.Canvas) *toolTip {
	rect := canvas.NewRectangle(color.Black)
	rect.SetMinSize(fyne.NewSize(57, 36))

	return &toolTip{
		text:      canvas.NewText("", bundle.TextColor),
		canvas:    can,
		popup:     &widget.PopUp{},
		offset:    offset,
		Rectangle: rect,
	}
}

func (t *toolTip) MouseIn(event *desktop.MouseEvent) {
	t.hovered = true
	if t.canvas != nil {
		if Gnomes.Indexer != nil {
			t.text.Text = fmt.Sprintf("%d/%d (%s)", Gnomes.Indexer.LastIndexedHeight, Gnomes.Indexer.ChainHeight, Gnomes.Status())
			t.popup = widget.NewPopUp(t.text, t.canvas)
		} else {
			t.text.Text = "0/0"
			t.popup = widget.NewPopUp(t.text, t.canvas)
		}

		if !t.popup.Hidden {
			pos := event.AbsolutePosition
			t.popup.ShowAtPosition(fyne.NewPos(pos.X-t.offset/4.5, pos.Y+t.offset))
			t.Refresh()
		}
	}
}

func (t *toolTip) MouseOut() {
	if t.canvas != nil {
		t.hovered = false
		t.popup.Hide()
		t.popup = nil
		t.Refresh()
	}
}

func (t *toolTip) MouseMoved(event *desktop.MouseEvent) {}

func (t *toolTip) Refresh() {
	if t.hovered {
		if Gnomes.Indexer != nil {
			t.text.Text = fmt.Sprintf("%d/%d (%s)", Gnomes.Indexer.LastIndexedHeight, Gnomes.Indexer.ChainHeight, Gnomes.Status())
		} else {
			t.text.Text = "0/0"
		}
		t.text.Refresh()
	}
}
