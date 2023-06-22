package main

import (
	"log"
	"runtime"
	"time"

	dreams "github.com/SixofClubsss/dReams"
	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/menu"

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

var dReams dreams.DreamsObject

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
	dReams.Window.SetIcon(bundle.ResourceDReamsIconPng)
	dReams.Window.SetMaster()
	done := make(chan struct{})

	dReams.Window.SetCloseIntercept(func() {
		menu.CloseAppSignal(true)
		menu.WriteDreamsConfig(save())
		serviceRunning()
		go menu.StopLabel()
		menu.Gnomes.Stop("dReams")
		dReams.StopProcess()
		menu.StopIndicators(indicators)
		time.Sleep(time.Second)
		dReams.Window.Close()
	})

	dReams.Menu = true

	dreams.Theme.Img = *canvas.NewImageFromResource(bundle.ResourceBackgroundPng)
	dReams.Background = container.NewMax(&dreams.Theme.Img)

	dapps := len(menu.Control.Dapp_list)
	if dapps == 0 {
		go func() {
			time.Sleep(300 * time.Millisecond)
			dReams.Window.SetContent(
				container.New(layout.NewMaxLayout(),
					dReams.Background,
					introScreen()))
		}()
	} else {
		go func() {
			// put back
			//dReams.SetChannels(dapps)
			dReams.SetChannels(4)
			time.Sleep(750 * time.Millisecond)
			dReams.Window.SetContent(
				container.New(layout.NewMaxLayout(),
					dReams.Background,
					place()))

		}()
	}

	if systemTray(dReams.App) {
		dReams.App.(desktop.App).SetSystemTrayIcon(bundle.ResourceTrayIconPng)
	}

	go fetch(dReams, done)
	dReams.Window.ShowAndRun()
	<-done
	log.Println("[dReams] Closed")
}
