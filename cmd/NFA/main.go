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
	"github.com/deroproject/derohe/walletapi"
	"github.com/sirupsen/logrus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// dApp to run NFA market with full wallet controls from dReams packages

const (
	appName = "NFA Market"
	appID   = "dreamdapps.io.nfa"
)

func main() {
	n := runtime.NumCPU()
	runtime.GOMAXPROCS(n)

	// Initialize logrus logger to stdout
	logger := structures.Logger.WithFields(logrus.Fields{})
	gnomes.InitLogrusLog(logrus.InfoLevel)

	// Read config.json file
	config := menu.ReadDreamsConfig(appName)

	// New gnomes instance for app
	gnomon := gnomes.NewGnomes()

	// Initialize Fyne app and window as dreams.AppObject
	d := dreams.NewFyneApp(
		appID,
		appName,
		bundle.DeroTheme(config.Skin),
		bundle.ResourceMarketIconPng,
		menu.DefaultBackgroundResource(),
		rpc.NewXSWDApplicationData(appName, "Non-Fungible Asset Market", appID, true))

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
			Theme:  dreams.Theme.Name,
		}

		if rpc.Daemon.Rpc == "" {
			save.Daemon = config.Daemon
		} else {
			save.Daemon = []string{rpc.Daemon.Rpc}
		}

		menu.WriteDreamsConfig(save)
		menu.SetClose(true)
		gnomon.Stop(appName)
		d.StopProcess()
		rpc.Wallet.CloseConnections(appName)
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

	// Initialize Gnomon vars
	gnomon.SetFastsync(true, true, 3000)
	gnomon.SetDBStorageType("boltdb")

	// Create dwidget connection box, using default OnTapped for RPC/XSWD connections
	connection := dwidget.NewHorizontalEntries(appName, 1, &d)

	// Gnomon controlled by daemon connection
	connection.Connected.OnChanged = func(b bool) {
		if b {
			if rpc.Daemon.IsConnected() && !gnomon.IsInitialized() && !gnomon.IsStarting() {
				go gnomes.StartGnomon(appName, gnomon.DBStorageType(), []string{gnomes.NFA_SEARCH_FILTER}, 0, 0, nil)
				rpc.GetFees()
				menu.Market.Filters = menu.GetFilters("market_filter")
			}
		} else {
			gnomon.Stop(appName)
		}
	}

	// Set any saved daemon configs
	connection.AddDaemonOptions(config.Daemon)

	// Adding dReams indicator panel for wallet, daemon and Gnomon
	connection.AddIndicator(menu.StartIndicators(nil))

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
		container.NewTabItem("Assets", menu.PlaceAssets(appName, profile, rescan, bundle.ResourceMarketIconPng, &d)),
		container.NewTabItem("Mint", menu.PlaceNFAMint(appName, d.Window)),
		container.NewTabItem("Log", rpc.SessionLog(appName, rpc.Version())))

	tabs.SetTabLocation(container.TabLocationBottom)

	go walletapi.Initialize_LookupTable(1, 1<<24)

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
				rpc.Wallet.Sync()

				// Refresh Dero balance and Gnomon endpoint
				connection.RefreshBalance()

				if rpc.Daemon.IsConnected() {
					connection.Connected.SetChecked(true)
					gnomes.EndPoint()

					if gnomon.IsRunning() {
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
					}

				} else {
					gnomon.Synced(false)
					connection.Connected.SetChecked(false)
				}

				d.SignalChannel()

			case <-d.Closing():
				logger.Printf("[%s] Closing...", appName)
				ticker.Stop()
				d.CloseAllDapps()
				time.Sleep(time.Second)
				done <- struct{}{}
				return
			}
		}
	}()

	// Start app and place layout
	go func() {
		time.Sleep(450 * time.Millisecond)
		d.Window.SetContent(container.NewStack(d.Background, tabs, container.NewVBox(layout.NewSpacer(), connection.Container)))
	}()
	d.Window.ShowAndRun()
	<-done
	logger.Printf("[%s] Closed\n", appName)
}
