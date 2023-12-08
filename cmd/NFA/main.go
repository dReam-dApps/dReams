package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/civilware/Gnomon/structures"
	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/gnomes"
	"github.com/dReam-dApps/dReams/menu"
	"github.com/dReam-dApps/dReams/rpc"
	"github.com/sirupsen/logrus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

// dApp to run NFA market with full wallet controls from dReams packages

const app_tag = "NFA Market"

func main() {
	n := runtime.NumCPU()
	runtime.GOMAXPROCS(n)
	gnomes.InitLogrusLog(logrus.InfoLevel)
	logger := structures.Logger.WithFields(logrus.Fields{})
	config := menu.ReadDreamsConfig(app_tag)
	gnomon := gnomes.NewGnomes()

	// Initialize Fyne app and window
	var d dreams.AppObject
	d.App = app.NewWithID(fmt.Sprintf("%s Desktop Client", app_tag))
	d.App.Settings().SetTheme(bundle.DeroTheme(config.Skin))
	d.Window = d.App.NewWindow(app_tag)
	d.Window.Resize(fyne.NewSize(1400, 800))
	d.Window.SetIcon(bundle.ResourceMarketIconPng)
	d.Window.CenterOnScreen()
	d.Window.SetMaster()

	// Initialize closing channels and func
	quit := make(chan struct{})
	done := make(chan struct{})
	closeFunc := func() {
		save := dreams.SaveData{
			Skin:   config.Skin,
			DBtype: gnomon.DBStorageType(),
		}

		if rpc.Daemon.Rpc == "" {
			save.Daemon = config.Daemon
		} else {
			save.Daemon = []string{rpc.Daemon.Rpc}
		}

		menu.WriteDreamsConfig(save)
		menu.CloseAppSignal(true)
		gnomon.Stop(app_tag)
		quit <- struct{}{}
		if rpc.Wallet.File != nil {
			rpc.Wallet.File.Close_Encrypted_Wallet()
		}
		d.Window.Close()
	}
	d.Window.SetCloseIntercept(closeFunc)

	// Handle ctrl-c close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println()
		closeFunc()
	}()

	// Initialize vars
	gnomon.SetFastsync(true)

	// Create dwidget connection box with controls
	connect_box := dwidget.NewHorizontalEntries(app_tag, 1)
	connect_box.Button.OnTapped = func() {
		rpc.GetAddress(app_tag)
		rpc.Ping()
		if rpc.Daemon.IsConnected() && !gnomon.IsInitialized() && !gnomon.IsStarting() {
			go gnomes.StartGnomon(app_tag, gnomon.DBStorageType(), []string{gnomes.NFA_SEARCH_FILTER}, 0, 0, nil)
			rpc.FetchFees()
			menu.Market.Filters = menu.FetchFilters("market_filter")
		}
	}

	connect_box.Disconnect.OnChanged = func(b bool) {
		if !b {
			gnomon.Stop(app_tag)
		}
	}

	connect_box.AddDaemonOptions(config.Daemon)
	connect_box.Container.Objects[0].(*fyne.Container).Add(menu.StartIndicators())

	// Layout tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Market", menu.PlaceMarket()),
		container.NewTabItem("Assets", menu.PlaceAssets(app_tag, nil, nil, bundle.ResourceMarketIconPng, &d)),
		container.NewTabItem("Mint", menu.PlaceNFAMint(app_tag, d.Window)),
		container.NewTabItem("Log", rpc.SessionLog(app_tag, rpc.Version())))

	tabs.SetTabLocation(container.TabLocationBottom)

	go menu.RunNFAMarket(app_tag, quit, done, connect_box)
	go func() {
		time.Sleep(450 * time.Millisecond)
		d.Window.SetContent(container.NewStack(tabs, container.NewVBox(layout.NewSpacer(), connect_box.Container)))
	}()
	d.Window.ShowAndRun()
	<-done
	logger.Printf("[%s] Closed\n", app_tag)
}
