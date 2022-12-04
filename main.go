package main

import (
	"runtime"
	"time"

	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/table"

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
	menu      bool
	holdero   bool
	bacc      bool
	predict   bool
	sports    bool
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
	dReams.App = app.NewWithID(App_ID)
	dReams.App.Settings().SetTheme(Theme())
	dReams.Window = dReams.App.NewWindow(App_Name)
	dReams.Window.SetMaster()
	dReams.Window.Resize(fyne.NewSize(MIN_WIDTH, MIN_HEIGHT))
	dReams.Window.SetFixedSize(false)
	dReams.Window.SetIcon(resourceCardSharkTrayPng)
	dReams.Window.SetMaster()
	quit := make(chan struct{})
	dReams.Window.SetCloseIntercept(func() {
		writeConfig(makeConfig(table.Poker_name, rpc.Round.Daemon))
		menu.StopGnomon(menu.Gnomes.Init)
		quit <- struct{}{}
		time.Sleep(1 * time.Second)
		dReams.Window.Close()
	})

	menu.GetMenuResources(resourceGnomonIconPng, resourceAvatarFramePng, resourceCwBackgroundPng, resourceMwBackgroundPng, resourceOwBackgroundPng, resourceUwBackgroundPng)
	table.GetTableResources(resourceGnomonIconPng, resourceMwBackgroundPng, resourceOwBackgroundPng, resourceBackgroundPng, resourceUwBackgroundPng)

	rpc.Signal.Startup = true
	rpc.Bacc.Display = true
	dReams.menu = true
	table.InitTableSettings()
	table.Settings.ThemeImg = *canvas.NewImageFromResource(resourceBackgroundPng)
	background = container.NewMax(&table.Settings.ThemeImg)

	table.Poker_name, menu.PlayerControl.Daemon_config = readConfig()
	go func() {
		dReams.Window.SetContent(
			container.New(layout.NewMaxLayout(),
				background,
				place()))
	}()
	dReams.os = runtime.GOOS
	if systemTray(dReams.App) {
		dReams.App.(desktop.App).SetSystemTrayIcon(resourceCardSharkTrayPng)
	}
	go fetch(quit)
	dReams.Window.ShowAndRun()
}
