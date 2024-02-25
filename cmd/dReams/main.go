package main

import (
	"runtime"
	"time"

	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/gnomes"
	"github.com/dReam-dApps/dReams/menu"
	"github.com/dReam-dApps/dReams/rpc"
	"github.com/deroproject/derohe/walletapi"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
)

const (
	App_ID   = "dreamdapps.io"
	App_Name = "dReams"
)

var dReams dreams.AppObject
var gnomon = gnomes.NewGnomes()

func main() {
	n := runtime.NumCPU()
	runtime.GOMAXPROCS(n)

	flags()
	dReams.App = app.NewWithID(App_ID)
	dReams.App.Settings().SetTheme(bundle.DeroTheme(bundle.AppColor))
	dReams.Window = dReams.App.NewWindow(App_Name)
	dReams.Window.SetMaster()
	dReams.Window.Resize(fyne.NewSize(dreams.MIN_WIDTH, dreams.MIN_HEIGHT))
	dReams.Window.SetFixedSize(false)
	dReams.Window.SetIcon(bundle.ResourceDReamsIconPng)
	dReams.Window.CenterOnScreen()
	done := make(chan struct{})

	dReams.XSWD = rpc.NewXSWDApplicationData(App_Name, App_ID, App_ID, true)
	dreams.Theme.Img = *canvas.NewImageFromResource(menu.DefaultBackgroundResource())
	dReams.Background = container.NewStack(&dreams.Theme.Img)
	dReams.Window.SetContent(splashScreen())

	close := func() {
		menu.SetClose(true)
		menu.WriteDreamsConfig(save())
		dappCloseCheck()
		menu.Info.SetStatus("Putting Gnomon to Sleep")
		gnomon.Stop("dReams")
		dReams.StopProcess()
		menu.StopIndicators(indicators)
		time.Sleep(time.Second)
		dReams.Window.Close()
	}

	dReams.Window.SetCloseIntercept(func() {
		if gnomon.IsStarting() {
			dReams.Window.RequestFocus()
			dialog.NewConfirm("Gnomon Syncing", "Are you sure you want to close dReams?", func(b bool) {
				if b {
					close()
				}
			}, dReams.Window).Show()
		} else {
			close()
		}
	})

	go walletapi.Initialize_LookupTable(1, 1<<24)

	dReams.SetTab("Menu")

	dapps := menu.EnabledDappCount()
	if dapps == 0 {
		go func() {
			dReams.SetChannels(0)
			time.Sleep(300 * time.Millisecond)
			dReams.Window.SetContent(container.NewStack(dReams.Background, introScreen()))
		}()
	} else {
		go func() {
			dReams.SetChannels(dapps)
			time.Sleep(750 * time.Millisecond)
			dReams.Window.SetContent(container.NewStack(dReams.Background, place()))
			dReams.Window.Resize(fyne.NewSize(dreams.MIN_WIDTH, dreams.MIN_HEIGHT))
		}()
	}

	// if systemTray(dReams.App) {
	// 	dReams.App.(desktop.App).SetSystemTrayIcon(bundle.ResourceTrayIconPng)
	// }

	go fetch(done)
	dReams.Window.ShowAndRun()
	<-done
	logger.Println("[dReams] Closed")
}
