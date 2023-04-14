package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	config := menu.ReadDreamsConfig(app_tag)
	a := app.New()
	a.Settings().SetTheme(bundle.DeroTheme(config.Skin))
	w := a.NewWindow(app_tag)
	w.Resize(fyne.NewSize(1200, 800))
	w.SetMaster()
	quit := make(chan struct{})
	w.SetCloseIntercept(func() {
		menu.WriteDreamsConfig(rpc.Daemon.Rpc, config.Skin)
		quit <- struct{}{}
		w.Close()
	})

	menu.Gnomes.Fast = true
	connect_box := dwidget.HorizontalEntries(app_tag, 1)
	connect_box.Button.OnTapped = func() {
		rpc.GetAddress(app_tag)
		rpc.Ping()
		if rpc.Daemon.Connect && !menu.Gnomes.Init && !menu.Gnomes.Start {
			go menu.StartGnomon(app_tag, []string{menu.NFA_SEARCH_FILTER}, 0, 0, nil)
		}
	}

	tabs := container.NewAppTabs(
		container.NewTabItem("Market", menu.PlaceMarket()),
		container.NewTabItem("Assets", menu.PlaceAssets(app_tag, false, nil, nil, nil)))

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
		log.Printf("[%s] Closing", app_tag)
		w.Close()
	}()

	go menu.RunNFAMarketRoutine(app_tag, quit, connect_box)
	w.SetContent(max)
	w.ShowAndRun()
}
