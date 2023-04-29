package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/dwidget"
	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"
)

// dApp to run NFA market with full wallet controls from dReams packages

const app_tag = "NFA Market"

func main() {
	n := runtime.NumCPU()
	runtime.GOMAXPROCS(n)
	config := menu.ReadDreamsConfig(app_tag)

	a := app.New()
	a.Settings().SetTheme(bundle.DeroTheme(config.Skin))
	w := a.NewWindow(app_tag)
	w.Resize(fyne.NewSize(1200, 800))
	w.SetMaster()
	quit := make(chan struct{})
	w.SetCloseIntercept(func() {
		menu.WriteDreamsConfig(rpc.Daemon.Rpc, config.Skin)
		menu.StopGnomon(app_tag)
		quit <- struct{}{}
		if rpc.Wallet.File != nil {
			rpc.Wallet.File.Close_Encrypted_Wallet()
		}
		w.Close()
	})

	menu.Gnomes.Fast = true
	connect_box := dwidget.HorizontalEntries(app_tag, 1)
	connect_box.Button.OnTapped = func() {
		rpc.GetAddress(app_tag)
		rpc.Ping()
		if rpc.Daemon.Connect && !menu.Gnomes.Init && !menu.Gnomes.Start {
			go menu.StartGnomon(app_tag, []string{menu.NFA_SEARCH_FILTER}, 0, 0, nil)
			rpc.FetchFees()
			menu.FetchFilters()
		}
	}

	connect_box.AddDaemonOptions(config.Daemon)
	connect_box.Container.Objects[0].(*fyne.Container).Add(menu.StartIndicators())

	tabs := container.NewAppTabs(
		container.NewTabItem("Market", menu.PlaceMarket()),
		container.NewTabItem("Assets", menu.PlaceAssets(app_tag, false, nil, nil, nil)),
		container.NewTabItem("Mint", menu.PlaceNFAMint(app_tag, w)))

	tabs.SetTabLocation(container.TabLocationBottom)

	max := container.NewMax(tabs, container.NewVBox(layout.NewSpacer(), connect_box.Container))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println()
		menu.WriteDreamsConfig(rpc.Daemon.Rpc, config.Skin)
		menu.StopGnomon(app_tag)
		rpc.Wallet.Connect = false
		quit <- struct{}{}
		if rpc.Wallet.File != nil {
			rpc.Wallet.File.Close_Encrypted_Wallet()
		}
		w.Close()
	}()

	go menu.RunNFAMarket(app_tag, quit, connect_box)
	go func() {
		time.Sleep(450 * time.Millisecond)
		w.SetContent(max)
	}()
	w.ShowAndRun()
}
