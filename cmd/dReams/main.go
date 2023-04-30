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
	App_ID     = "dReam Tables App"
	App_Name   = "dReams"
)

type dReamTables struct {
	App       fyne.App
	Window    fyne.Window
	os        string
	configure bool
	menu      bool
	holdero   bool
	bacc      bool
	predict   bool
	sports    bool
	tarot     bool
	menu_tabs struct {
		wallet    bool
		contracts bool
		assets    bool
		market    bool
	}
}

var dReams dReamTables
var background *fyne.Container

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
	quit := make(chan struct{})

	dReams.Window.SetCloseIntercept(func() {
		menu.Exit_signal = true
		menu.WriteDreamsConfig(rpc.Daemon.Rpc, bundle.AppColor)
		serviceRunning()
		go menu.StopLabel()
		menu.StopGnomon("dReams")
		quit <- struct{}{}
		menu.StopIndicators()
		time.Sleep(time.Second)
		dReams.Window.Close()
	})

	dReams.menu = true

	holdero.Settings.ThemeImg = *canvas.NewImageFromResource(bundle.ResourceBackgroundPng)
	background = container.NewMax(&holdero.Settings.ThemeImg)

	if len(menu.Control.Dapp_list) == 0 {
		go func() {
			time.Sleep(300 * time.Millisecond)
			dReams.Window.SetContent(
				container.New(layout.NewMaxLayout(),
					background,
					introScreen()))
		}()
	} else {
		go func() {
			time.Sleep(750 * time.Millisecond)
			dReams.Window.SetContent(
				container.New(layout.NewMaxLayout(),
					background,
					place()))

		}()
	}

	if systemTray(dReams.App) {
		dReams.App.(desktop.App).SetSystemTrayIcon(bundle.ResourceCardSharkTrayPng)
	}

	go fetch(quit)
	dReams.Window.ShowAndRun()
}
