package main

import (
	"runtime"
	"time"

	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/holdero"
	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
)

const (
	MIN_WIDTH  = 1400
	MIN_HEIGHT = 800
	App_ID     = "dReams Platform"
	App_Name   = "dReams"
)

type dReamsObjects struct {
	App        fyne.App
	Window     fyne.Window
	background *fyne.Container
	os         string
	configure  bool
	menu       bool
	holdero    bool
	bacc       bool
	predict    bool
	sports     bool
	tarot      bool
	cli        bool
	quit       chan struct{}
	menu_tabs  struct {
		wallet    bool
		contracts bool
		assets    bool
		market    bool
	}
}

var dReams dReamsObjects

func main() {
	n := runtime.NumCPU()
	runtime.GOMAXPROCS(n)

	v := flags()
	stamp(v)
	dReams.App = app.NewWithID(App_ID)
	dReams.App.Settings().SetTheme(bundle.DeroTheme(bundle.AppColor))
	dReams.Window = dReams.App.NewWindow(App_Name)
	dReams.Window.SetMaster()
	dReams.Window.Resize(fyne.NewSize(MIN_WIDTH, MIN_HEIGHT))
	dReams.Window.SetFixedSize(false)
	dReams.Window.SetIcon(bundle.ResourceCardSharkTrayPng)
	dReams.Window.SetMaster()
	dReams.quit = make(chan struct{})
	done := make(chan struct{})

	dReams.Window.SetCloseIntercept(func() {
		menu.CloseAppSignal(true)
		menu.WriteDreamsConfig(rpc.Daemon.Rpc, bundle.AppColor)
		serviceRunning()
		go menu.StopLabel()
		menu.StopGnomon("dReams")
		dReams.quit <- struct{}{}
		menu.StopIndicators()
		time.Sleep(time.Second)
		dReams.Window.Close()
	})

	dReams.menu = true

	holdero.Settings.ThemeImg = *canvas.NewImageFromResource(bundle.ResourceBackgroundPng)
	dReams.background = container.NewMax(&holdero.Settings.ThemeImg)

	if len(menu.Control.Dapp_list) == 0 {
		go func() {
			time.Sleep(300 * time.Millisecond)
			dReams.Window.SetContent(
				container.New(layout.NewMaxLayout(),
					dReams.background,
					introScreen()))
		}()
	} else {
		go func() {
			time.Sleep(750 * time.Millisecond)
			dReams.Window.SetContent(
				container.New(layout.NewMaxLayout(),
					dReams.background,
					place()))

		}()
	}

	if systemTray(dReams.App) {
		dReams.App.(desktop.App).SetSystemTrayIcon(bundle.ResourceCardSharkTrayPng)
	}

	go fetch(dReams.quit, done)
	dReams.Window.ShowAndRun()
	<-done
}
