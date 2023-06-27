package main

import (
	"log"
	"runtime"
	"time"

	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/menu"

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
	done := make(chan struct{})

	dReams.Window.SetCloseIntercept(func() {
		menu.CloseAppSignal(true)
		menu.WriteDreamsConfig(save())
		dappCloseCheck()
		go menu.StopLabel()
		menu.Gnomes.Stop("dReams")
		dReams.StopProcess()
		menu.StopIndicators(indicators)
		time.Sleep(time.Second)
		dReams.Window.Close()
	})

	dReams.SetTab("Menu")
	dreams.Theme.Img = *canvas.NewImageFromResource(bundle.ResourceBackgroundPng)
	dReams.Background = container.NewMax(&dreams.Theme.Img)

	dapps := menu.EnabledDapps()
	if dapps == 0 {
		go func() {
			dReams.SetChannels(0)
			time.Sleep(300 * time.Millisecond)
			dReams.Window.SetContent(
				container.New(layout.NewMaxLayout(),
					dReams.Background,
					introScreen()))
		}()
	} else {
		go func() {
			dReams.SetChannels(dapps)
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

	go fetch(done)
	dReams.Window.ShowAndRun()
	<-done
	log.Println("[dReams] Closed")
}
