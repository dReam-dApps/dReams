package main

import (
	"SixofClubsss/dReams/menu"
	"SixofClubsss/dReams/rpc"
	"SixofClubsss/dReams/table"
	"time"

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
	ID         = "dReam Tables App"
	Name       = "dReams"
)

type dReamTables struct {
	App       fyne.App
	Window    fyne.Window
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
	dReams.App = app.NewWithID(ID)
	dReams.App.Settings().SetTheme(Theme())
	dReams.Window = dReams.App.NewWindow(Name)
	dReams.Window.SetMaster()
	dReams.Window.Resize(fyne.NewSize(MIN_WIDTH, MIN_HEIGHT))
	dReams.Window.SetFixedSize(false)
	dReams.Window.SetIcon(resourceCardSharkTrayPng)
	dReams.Window.SetMaster()
	quit := make(chan struct{})
	dReams.Window.SetCloseIntercept(func() {
		writeConfig(save{Name: table.Poker_name})
		menu.StopGnomon(menu.Gnomes.Init)
		quit <- struct{}{}
		time.Sleep(1 * time.Second)
		dReams.Window.Close()
	})

	menu.GetMenuResources(resourceGnomonIconPng, resourceAvatarFramePng, resourceCwBackgroundPng, resourceMwBackgroundPng, resourceOwBackgroundPng)
	table.GetTableResources(resourceGnomonIconPng, resourceMwBackgroundPng, resourceOwBackgroundPng, resourceBackgroundPng)

	rpc.Signal.Startup = true
	rpc.Bacc.Display = true
	dReams.menu = true
	table.InitTableSettings()
	table.Settings.ThemeImg = *canvas.NewImageFromResource(resourceBackgroundPng)
	background = container.NewMax(&table.Settings.ThemeImg)

	table.Poker_name = readConfig()
	go func() {
		dReams.Window.SetContent(
			container.New(layout.NewMaxLayout(),
				background,
				place()))
	}()

	if systemTray(dReams.App) {
		dReams.App.(desktop.App).SetSystemTrayIcon(resourceCardSharkTrayPng)
	}
	go fetch(quit)
	dReams.Window.ShowAndRun()
}
