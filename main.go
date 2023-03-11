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
		serviceRunning()
		menu.StopGnomon(menu.Gnomes.Init)
		quit <- struct{}{}
		menu.StopIndicators()
		time.Sleep(1 * time.Second)
		dReams.Window.Close()
	})

	menu.GetMenuResources(resourceDTGnomonIconPng, resourceAvatarFramePng, resourceCwBackgroundPng, resourceMwBackgroundPng, resourceOwBackgroundPng, resourceUwBackgroundPng, resourceGnomoniconPng, resourceBlueBadgePng, resourceBlueBadge2Png, resourceBlueBadge3Png, resourceRedBadgePng, resourcePokerBoticonPng, resourceDReamServiceiconPng, resourceDReamToolsPng, resourceDReamToolsHalfPng)
	table.GetTableResources(resourceDTGnomonIconPng, resourceMwBackgroundPng, resourceOwBackgroundPng, resourceBackgroundPng, resourceUwBackgroundPng, resourceIlumabackground1Png, resourceIlumabackground2Png, resourceIluma81Png)

	dReams.menu = true

	table.Settings.ThemeImg = *canvas.NewImageFromResource(resourceBackgroundPng)
	background = container.NewMax(&table.Settings.ThemeImg)

	go func() {
		dReams.Window.SetContent(
			container.New(layout.NewMaxLayout(),
				background,
				place()))
	}()

	time.Sleep(500 * time.Millisecond)
	if systemTray(dReams.App) {
		dReams.App.(desktop.App).SetSystemTrayIcon(resourceCardSharkTrayPng)
	}

	go fetch(quit)
	dReams.Window.ShowAndRun()
}
