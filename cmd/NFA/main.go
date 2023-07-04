package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/menu"
	"github.com/dReam-dApps/dReams/rpc"

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
	menu.InitLogrusLog(runtime.GOOS != "windows")
	config := menu.ReadDreamsConfig(app_tag)

	a := app.New()
	a.Settings().SetTheme(bundle.DeroTheme(config.Skin))
	w := a.NewWindow(app_tag)
	w.Resize(fyne.NewSize(1200, 800))
	w.SetIcon(bundle.ResourceMarketIconPng)
	w.SetMaster()
	quit := make(chan struct{})
	done := make(chan struct{})
	w.SetCloseIntercept(func() {
		menu.WriteDreamsConfig(
			dreams.DreamSave{
				Skin:   config.Skin,
				Daemon: []string{rpc.Daemon.Rpc},
				DBtype: menu.Gnomes.DBType,
			})
		menu.Gnomes.Stop(app_tag)
		quit <- struct{}{}
		if rpc.Wallet.File != nil {
			rpc.Wallet.File.Close_Encrypted_Wallet()
		}
		w.Close()
	})

	menu.Gnomes.Fast = true
	connect_box := dwidget.NewHorizontalEntries(app_tag, 1)
	connect_box.Button.OnTapped = func() {
		rpc.GetAddress(app_tag)
		rpc.Ping()
		if rpc.Daemon.IsConnected() && !menu.Gnomes.IsInitialized() && !menu.Gnomes.Start {
			go menu.StartGnomon(app_tag, menu.Gnomes.DBType, []string{menu.NFA_SEARCH_FILTER}, 0, 0, nil)
			rpc.FetchFees()
			menu.Market.Filters = menu.FetchFilters("market_filter")
		}
	}

	connect_box.Disconnect.OnChanged = func(b bool) {
		if !b {
			menu.Gnomes.Stop(app_tag)
		}
	}

	connect_box.AddDaemonOptions(config.Daemon)
	connect_box.Container.Objects[0].(*fyne.Container).Add(menu.StartIndicators())

	tabs := container.NewAppTabs(
		container.NewTabItem("Market", menu.PlaceMarket()),
		container.NewTabItem("Assets", menu.PlaceAssets(app_tag, nil, bundle.ResourceMarketIconPng, w)),
		container.NewTabItem("Mint", menu.PlaceNFAMint(app_tag, w)))

	tabs.SetTabLocation(container.TabLocationBottom)

	max := container.NewMax(tabs, container.NewVBox(layout.NewSpacer(), connect_box.Container))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println()
		menu.WriteDreamsConfig(
			dreams.DreamSave{
				Skin:   config.Skin,
				Daemon: []string{rpc.Daemon.Rpc},
				DBtype: menu.Gnomes.DBType,
			})
		menu.Gnomes.Stop(app_tag)
		rpc.Wallet.Connected(false)
		quit <- struct{}{}
		if rpc.Wallet.File != nil {
			rpc.Wallet.File.Close_Encrypted_Wallet()
		}
		w.Close()
	}()

	go menu.RunNFAMarket(app_tag, quit, done, connect_box)
	go func() {
		time.Sleep(450 * time.Millisecond)
		w.SetContent(max)
	}()
	w.ShowAndRun()
	<-done
}
