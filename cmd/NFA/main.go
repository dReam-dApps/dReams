package main

import (
	"fmt"
	"image/color"
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
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
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
	a := app.NewWithID(fmt.Sprintf("%s Desktop Client", app_tag))
	a.Settings().SetTheme(bundle.DeroTheme(config.Skin))
	w := a.NewWindow(app_tag)
	w.Resize(fyne.NewSize(1400, 800))
	w.SetIcon(bundle.ResourceMarketIconPng)
	w.CenterOnScreen()
	w.SetMaster()

	// Initialize dReams AppObject
	menu.Theme.Img = *canvas.NewImageFromResource(menu.DefaultThemeResource())
	d := dreams.AppObject{
		App:        a,
		Window:     w,
		Background: container.NewStack(&menu.Theme.Img),
	}

	// Enable calling RunNFAMarket
	enabled := menu.EnabledDappCount()
	if enabled < 1 {
		enabled = 1
	}
	d.SetChannels(enabled)

	// Initialize closing channels and func
	done := make(chan struct{})
	closeFunc := func() {
		save := dreams.SaveData{
			Skin:   config.Skin,
			DBtype: gnomon.DBStorageType(),
			Theme:  menu.Theme.Name,
		}

		if rpc.Daemon.Rpc == "" {
			save.Daemon = config.Daemon
		} else {
			save.Daemon = []string{rpc.Daemon.Rpc}
		}

		menu.WriteDreamsConfig(save)
		menu.SetClose(true)
		gnomon.Stop(app_tag)
		d.StopProcess()
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
	gnomon.SetFastsync(true, true, 10000)
	gnomon.SetDBStorageType("boltdb")

	// Create dwidget connection box with controls
	connect_box := dwidget.NewHorizontalEntries(app_tag, 1)
	connect_box.Button.OnTapped = func() {
		rpc.GetAddress(app_tag)
		rpc.Ping()
		if rpc.Daemon.IsConnected() && !gnomon.IsInitialized() && !gnomon.IsStarting() {
			go gnomes.StartGnomon(app_tag, gnomon.DBStorageType(), []string{gnomes.NFA_SEARCH_FILTER}, 0, 0, nil)
			rpc.GetFees()
			menu.Market.Filters = menu.GetFilters("market_filter")
		}
	}

	connect_box.Disconnect.OnChanged = func(b bool) {
		if !b {
			gnomon.Stop(app_tag)
		}
	}

	connect_box.AddDaemonOptions(config.Daemon)
	connect_box.Container.Objects[0].(*fyne.Container).Add(menu.StartIndicators())

	// Layout asset profile objects
	line := canvas.NewLine(bundle.TextColor)
	form := []*widget.FormItem{}
	form = append(form, widget.NewFormItem("Name", menu.NameEntry()))
	form = append(form, widget.NewFormItem("", layout.NewSpacer()))
	form = append(form, widget.NewFormItem("", container.NewVBox(line)))
	form = append(form, widget.NewFormItem("Theme", menu.ThemeSelect(&d)))
	form = append(form, widget.NewFormItem("", layout.NewSpacer()))
	form = append(form, widget.NewFormItem("", container.NewVBox(line)))

	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(450, 0))

	profile := container.NewCenter(container.NewBorder(spacer, nil, nil, nil, widget.NewForm(form...)))

	// Initialize asset rescan func
	rescan := func() {
		menu.CheckAllNFAs(nil)
	}

	// Layout tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Market", menu.PlaceMarket(&d)),
		container.NewTabItem("Assets", menu.PlaceAssets(app_tag, profile, rescan, bundle.ResourceMarketIconPng, &d)),
		container.NewTabItem("Mint", menu.PlaceNFAMint(app_tag, d.Window)),
		container.NewTabItem("Log", rpc.SessionLog(app_tag, rpc.Version())))

	tabs.SetTabLocation(container.TabLocationBottom)

	// For RunNFAMarket routine
	d.SetSubTab("Market")

	// NFA Market routine, signals RunNFAMarket
	go func() {
		synced := false
		time.Sleep(3 * time.Second)
		ticker := time.NewTicker(3 * time.Second)
		for {
			select {
			case <-ticker.C:
				rpc.Ping()
				rpc.EchoWallet(app_tag)
				go rpc.GetDreamsBalances(rpc.SCIDs)
				rpc.GetWalletHeight(app_tag)

				// Refresh Dero balance and Gnomon endpoint
				connect_box.RefreshBalance()
				if !rpc.Startup {
					gnomes.EndPoint()
				}

				if rpc.Daemon.IsConnected() && gnomon.IsRunning() {
					rpc.Startup = false
					connect_box.Disconnect.SetChecked(true)

					// Check Gnomon index for SCs
					gnomon.IndexContains()
					if gnomon.HasIndex(1) {
						gnomon.Checked(true)
					}

					// Check Gnomon index for sync
					if gnomon.GetLastHeight() >= gnomon.GetChainHeight()-3 {
						gnomon.Synced(true)
					} else {
						synced = false
						gnomon.Synced(false)
						gnomon.Checked(false)
					}

					// Check wallet for all owned NFAs and store icons in boltdb
					if gnomon.IsSynced() {
						if !synced {
							menu.CheckAllNFAs(nil)
							menu.Assets.List.Refresh()
							if gnomon.DBStorageType() == "boltdb" {
								for _, r := range menu.Assets.Asset {
									gnomes.StoreBolt(r.Collection, r.Name, r)
								}
							}
							synced = true
						}
					}

				} else {
					connect_box.Disconnect.SetChecked(false)
				}

				d.SignalChannel()

			case <-d.Closing():
				logger.Printf("[%s] Closing...", app_tag)
				ticker.Stop()
				d.CloseAllDapps()
				time.Sleep(time.Second)
				done <- struct{}{}
				return
			}
		}
	}()

	go func() {
		time.Sleep(450 * time.Millisecond)
		d.Window.SetContent(container.NewStack(d.Background, tabs, container.NewVBox(layout.NewSpacer(), connect_box.Container)))
	}()
	d.Window.ShowAndRun()
	<-done
	logger.Printf("[%s] Closed\n", app_tag)
}
