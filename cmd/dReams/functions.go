package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	dreams "github.com/SixofClubsss/dReams"
	"github.com/SixofClubsss/dReams/baccarat"
	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/holdero"
	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/prediction"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/tarot"
	"github.com/docopt/docopt-go"
	"github.com/fyne-io/terminal"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Notification struct {
	Title, Content string
}

var cli *terminal.Terminal
var command_line string = `dReams
Platform for Dero dApps, powered by Gnomon.

Usage:
  dReams [options]
  dReams -h | --help

Options:
  -h --help     Show this screen.
  --cli=<false>		dReams option, enables cli app tab.
  --trim=<false>	dReams option, defaults true for minimum index search filters.
  --fastsync=<false>	Gnomon option,  true/false value to define loading at chain height on start up.
  --num-parallel-blocks=<5>   Gnomon option,  defines the number of parallel blocks to index.
  --dbtype=<boltdb>     Gnomon option,  defines type of database 'gravdb' or 'boltdb'.`

var offset int

// Set opts when starting dReams
func flags() (version string) {
	version = rpc.DREAMSv
	arguments, err := docopt.ParseArgs(command_line, nil, version)

	if err != nil {
		log.Fatalf("Error while parsing arguments: %s\n", err)
	}

	dbType := "boltdb"
	if arguments["--dbtype"] != nil {
		if arguments["--dbtype"] == "gravdb" {
			dbType = arguments["--dbtype"].(string)
		}
	}

	trim := true
	if arguments["--trim"] != nil {
		if arguments["--trim"].(string) == "false" {
			trim = false
		}
	}

	fastsync := true
	if arguments["--fastsync"] != nil {
		if arguments["--fastsync"].(string) == "false" {
			fastsync = false
		}
	}

	parallel := 1
	if arguments["--num-parallel-blocks"] != nil {
		s := arguments["--num-parallel-blocks"].(string)
		switch s {
		case "2":
			parallel = 2
		case "3":
			parallel = 3
		case "4":
			parallel = 4
		case "5":
			parallel = 5
		default:
			parallel = 1
		}
	}

	cli := false
	if arguments["--cli"] != nil {
		if arguments["--cli"].(string) == "true" {
			cli = true
		}
	}

	dReams.Cli = cli
	menu.Gnomes.Trim = trim
	menu.Gnomes.Fast = fastsync
	menu.Gnomes.Para = parallel
	menu.Gnomes.DBType = dbType

	return
}

func init() {
	saved := menu.ReadDreamsConfig("dReams")
	if saved.Daemon != nil {
		menu.Control.Daemon_config = saved.Daemon[0]
	}

	holdero.Settings.Favorites = saved.Tables
	menu.Control.Predict_favorites = saved.Predict
	menu.Control.Sports_favorites = saved.Sports

	menu.Market.DreamsFilter = true

	rpc.InitBalances()

	dReams.OS = runtime.GOOS
	prediction.SetPrintColors(dReams.OS)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		menu.CloseAppSignal(true)
		menu.WriteDreamsConfig(rpc.Daemon.Rpc, bundle.AppColor)
		fmt.Println()
		serviceRunning()
		go menu.StopLabel()
		menu.Gnomes.Stop("dReams")
		menu.StopIndicators(indicators)
		time.Sleep(time.Second)
		dReams.StopProcess()
		dReams.Window.Close()
	}()
}

// Starts a Fyne terminal in dReams
func startTerminal() *terminal.Terminal {
	cli = terminal.New()
	go func() {
		_ = cli.RunLocalShell()
	}()

	return cli
}

// Exit running dReams terminal
func exitTerminal() {
	if cli != nil {
		cli.Exit()
	}
}

// Ensure service is shutdown on app close
func serviceRunning() {
	rpc.Wallet.Service = false
	for prediction.Service.Processing {
		log.Println("[dReams] Waiting for service to close")
		time.Sleep(3 * time.Second)
	}
}

// Terminal start info, ascii art for linux
func stamp(v string) {
	if dReams.OS == "linux" {
		fmt.Println(string(bundle.ResourceStampTxt.StaticContent))
	}
	log.Println("[dReams]", v, runtime.GOOS, runtime.GOARCH)
}

// Make system tray with opts
//   - Send Dero message menu
//   - Explorer link
//   - Manual reveal key for Holdero
func systemTray(w fyne.App) bool {
	if desk, ok := w.(desktop.App); ok {
		m := fyne.NewMenu("MyApp",
			fyne.NewMenuItem("Send Message", func() {
				if !dReams.Configure && rpc.Wallet.IsConnected() {
					menu.SendMessageMenu("", bundle.ResourceDReamsIconAltPng)
				}
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Explorer", func() {
				link, _ := url.Parse("https://explorer.dero.io")
				fyne.CurrentApp().OpenURL(link)
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Reveal Key", func() {
				go holdero.RevealKey(rpc.Wallet.ClientKey)
			}))
		desk.SetSystemTrayMenu(m)

		return true
	}
	return false
}

// Top label background used on dApp tabs
func labelColorBlack(c *fyne.Container) *fyne.Container {
	var alpha *canvas.Rectangle
	if bundle.AppColor == color.White {
		alpha = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x33})
	} else {
		alpha = canvas.NewRectangle(color.RGBA{0, 0, 0, 150})
	}

	cont := container.New(layout.NewMaxLayout(), alpha, c)

	return cont
}

func gnomonScan(contracts map[string]string) {
	CheckDreamsG45s(menu.Gnomes.Check, contracts)
	CheckDreamsNFAs(menu.Gnomes.Check, contracts)
}

// Main dReams process loop
func fetch(d dreams.DreamsObject, done chan struct{}) {
	rpc.Signal.Startup = true
	time.Sleep(3 * time.Second)
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ticker.C: // do on interval
			if !dReams.Configure {
				rpc.Ping()
				rpc.EchoWallet("dReams")
				go rpc.GetDreamsBalances(rpc.SCIDs)
				rpc.GetWalletHeight("dReams")
				if !rpc.Signal.Startup {
					CheckConnection()
					menu.GnomonEndPoint()
					menu.GnomonState(dReams.IsWindows(), dReams.Configure, gnomonScan)
					dReams.Background.Refresh()

					// Betting
					if menu.Control.Dapp_list["dSports and dPredictions"] {
						if offset%5 == 0 {
							SportsRefresh(dReams.Sports)
						}

						S.RightLabel.SetText("dReams Balance: " + rpc.DisplayBalance("dReams") + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Display.Wallet_height)
						PredictionRefresh(dReams.Predict)
					}

					// Menu
					go MenuRefresh(dReams.Menu)

					offset++
					if offset >= 21 {
						offset = 0
					}
				}

				if rpc.Daemon.IsConnected() {
					if rpc.Signal.Startup {
						go refreshPriceDisplay(true)
					}

					rpc.Signal.Startup = false
				}

				dReams.SignalChannel()

			}
		case <-d.Closing(): // exit loop
			log.Println("[dReams] Closing...")
			ticker.Stop()
			dReams.CloseAllDapps()
			time.Sleep(time.Second)
			done <- struct{}{}
			return
		}
	}
}

// Refresh all dPrediction objects
func PredictionRefresh(tab bool) {
	if tab {
		if offset%5 == 0 {
			go prediction.SetPredictionInfo(prediction.Predict.Contract)
		}

		if offset == 11 || prediction.Predict.Prices.Text == "" {
			go prediction.SetPredictionPrices(rpc.Daemon.Connect)
		}

		P.RightLabel.SetText("dReams Balance: " + rpc.DisplayBalance("dReams") + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Display.Wallet_height)

		if menu.CheckActivePrediction(prediction.Predict.Contract) {
			go prediction.ShowPredictionControls()
		} else {
			prediction.DisablePredictions(true)
		}
	}
}

// Refresh all dSports objects
func SportsRefresh(tab bool) {
	if tab {
		go prediction.SetSportsInfo(prediction.Sports.Contract)
	}
}

// Refresh Gnomon height display
func refreshGnomonDisplay(index_height, c int) {
	if c == 1 {
		height := " Gnomon Height: " + strconv.Itoa(index_height)
		menu.Assets.Gnomes_height.Text = (height)
		menu.Assets.Gnomes_height.Refresh()
	} else {
		menu.Assets.Gnomes_height.Text = (" Gnomon Height: 0")
		menu.Assets.Gnomes_height.Refresh()
	}
}

// Refresh indexed asset count
func refreshIndexDisplay(c bool) {
	if c {
		scids := " Indexed SCIDs: " + strconv.Itoa(int(menu.Gnomes.SCIDS))
		menu.Assets.Gnomes_index.Text = (scids)
		menu.Assets.Gnomes_index.Refresh()
	} else {
		menu.Assets.Gnomes_index.Text = (" Indexed SCIDs: 0")
		menu.Assets.Gnomes_index.Refresh()
	}
}

// Refresh daemon height display
func refreshDaemonDisplay(c bool) {
	if c && rpc.Daemon.IsConnected() {
		dHeight := rpc.DaemonHeight("dReams", rpc.Daemon.Rpc)
		d := strconv.Itoa(int(dHeight))
		menu.Assets.Daem_height.Text = (" Daemon Height: " + d)
		menu.Assets.Daem_height.Refresh()
	} else {
		menu.Assets.Daem_height.Text = (" Daemon Height: 0")
		menu.Assets.Daem_height.Refresh()
	}
}

// Refresh menu wallet display
func refreshWalletDisplay(c bool) {
	if c {
		menu.Assets.Wall_height.Text = (" Wallet Height: " + rpc.Display.Wallet_height)
		menu.Assets.Wall_height.Refresh()
	} else {
		menu.Assets.Wall_height.Text = (" Wallet Height: 0")
		menu.Assets.Wall_height.Refresh()
	}
}

// Refresh current Dero-USDT price
func refreshPriceDisplay(c bool) {
	if c && rpc.Daemon.IsConnected() {
		_, price := holdero.GetPrice("DERO-USDT")
		menu.Assets.Dero_price.Text = (" Dero Price: $" + price)
		menu.Assets.Dero_price.Refresh()
	} else {
		menu.Assets.Dero_price.Text = (" Dero Price: $")
		menu.Assets.Dero_price.Refresh()
	}
}

// Refresh all menu gui objects
func MenuRefresh(tab bool) {
	if tab && menu.Gnomes.IsInitialized() {
		index := menu.Gnomes.Indexer.LastIndexedHeight
		if index < menu.Gnomes.Indexer.ChainHeight-4 || !menu.Gnomes.HasIndex(2) {
			menu.Assets.Gnomes_sync.Text = (" Gnomon Syncing...")
			menu.Assets.Gnomes_sync.Refresh()
		} else {
			menu.Assets.Gnomes_sync.Text = ("")
			menu.Assets.Gnomes_sync.Refresh()
		}
		go refreshGnomonDisplay(int(index), 1)
		go refreshIndexDisplay(true)

		if rpc.Daemon.IsConnected() {
			go refreshDaemonDisplay(true)
		}

		if offset == 20 {
			go refreshPriceDisplay(true)
		}

		if dReams.Menu_tabs.Market && !dReams.IsWindows() && !menu.ClosingApps() {
			menu.FindNfaListings(nil)
		}
	}

	if rpc.Daemon.IsConnected() {
		go refreshDaemonDisplay(true)
	} else {
		go refreshDaemonDisplay(false)
		go refreshGnomonDisplay(0, 0)
	}

	if rpc.Wallet.IsConnected() {
		go refreshWalletDisplay(true)
	} else {
		go refreshWalletDisplay(false)
	}

	menu.Assets.Balances.Refresh()

	if !dReams.Menu {
		menu.Market.Viewing = ""
		menu.Market.Viewing_coll = ""
	}
}

// Switch triggered when main tab changes
func MainTab(ti *container.TabItem) {
	switch ti.Text {
	case "Menu":
		dReams.Menu = true
		dReams.Holdero = false
		dReams.Bacc = false
		dReams.Predict = false
		dReams.Sports = false
		dReams.Tarot = false
		if holdero.Round.ID == 1 {
			holdero.Faces.Select.Enable()
			holdero.Backs.Select.Enable()
		}
		go MenuRefresh(dReams.Menu)
	case "Holdero":
		dReams.Menu = false
		dReams.Holdero = true
		dReams.Bacc = false
		dReams.Predict = false
		dReams.Sports = false
		dReams.Tarot = false
	case "Baccarat":
		dReams.Menu = false
		dReams.Holdero = false
		dReams.Bacc = true
		dReams.Predict = false
		dReams.Sports = false
		dReams.Tarot = false
		go func() {
			baccarat.GetBaccTables()
			baccarat.BaccRefresh(&B, dReams)
			if rpc.Wallet.IsConnected() && baccarat.Bacc.Display {
				baccarat.ActionBuffer(false)
			}
		}()
	case "Predict":
		dReams.Menu = false
		dReams.Holdero = false
		dReams.Bacc = false
		dReams.Predict = true
		dReams.Sports = false
		dReams.Tarot = false
		go func() {
			menu.PopulatePredictions(nil)
		}()
		PredictionRefresh(dReams.Predict)
	case "Sports":
		dReams.Menu = false
		dReams.Holdero = false
		dReams.Bacc = false
		dReams.Predict = false
		dReams.Sports = true
		dReams.Tarot = false
		go menu.PopulateSports(nil)
	case "Iluma":
		dReams.Menu = false
		dReams.Holdero = false
		dReams.Bacc = false
		dReams.Predict = false
		dReams.Sports = false
		dReams.Tarot = true
		if tarot.Iluma.Value.Display {
			tarot.ActionBuffer(false)
		}
	}
}

// Switch triggered when menu tab changes
func MenuTab(ti *container.TabItem) {
	switch ti.Text {
	case "Wallet":
		ti.Content.(*container.Split).Leading.(*container.Split).Trailing.Refresh()
		dReams.Menu_tabs.Wallet = true
		dReams.Menu_tabs.Assets = false
		dReams.Menu_tabs.Market = false
	case "Assets":
		dReams.Menu_tabs.Wallet = false
		dReams.Menu_tabs.Assets = true
		dReams.Menu_tabs.Market = false
		menu.Control.Viewing_asset = ""
		menu.Assets.Asset_list.UnselectAll()
	case "Market":
		dReams.Menu_tabs.Wallet = false
		dReams.Menu_tabs.Assets = false
		dReams.Menu_tabs.Market = true
		go menu.FindNfaListings(nil)
		menu.Market.Cancel_button.Hide()
		menu.Market.Close_button.Hide()
		menu.Market.Auction_list.Refresh()
		menu.Market.Buy_list.Refresh()
	}
}

// Switch triggered when dPrediction tab changes
func PredictTab(ti *container.TabItem) {
	switch ti.Text {
	case "Contracts":
		go menu.PopulatePredictions(nil)
	default:
	}
}

// Set and revert main window full screen mode
func FullScreenSet() fyne.CanvasObject {
	var button *widget.Button
	button = widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "viewFullScreen"), func() {
		if dReams.Window.FullScreen() {
			button.Icon = fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "viewFullScreen")
			dReams.Window.SetFullScreen(false)
		} else {
			button.Icon = fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "viewRestore")
			dReams.Window.SetFullScreen(true)
		}
	})

	button.Importance = widget.LowImportance

	cont := container.NewHBox(layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), container.NewVBox(button), layout.NewSpacer())

	return cont
}

// Check wallet for dReams NFAs
//   - Pass scids from db store, can be nil arg
//   - Pass false gc for rechecks
func CheckDreamsNFAs(gc bool, scids map[string]string) {
	if menu.Gnomes.IsReady() && !gc {
		menu.Assets.Gnomes_sync.Text = (" Checking for Assets")
		menu.Assets.Gnomes_sync.Refresh()

		if scids == nil {
			scids = menu.Gnomes.GetAllOwnersAndSCIDs()
		}
		keys := make([]string, len(scids))
		log.Println("[dReams] Checking NFA Assets")
		holdero.Faces.Select.Options = []string{}
		holdero.Backs.Select.Options = []string{}
		dreams.Theme.Select.Options = []string{}
		holdero.Settings.AvatarSelect.Options = []string{}

		i := 0
		for k := range scids {
			if !rpc.Wallet.IsConnected() || !menu.Gnomes.IsRunning() {
				break
			}
			keys[i] = k
			checkNFAOwner(keys[i])
			i++
		}
		sort.Strings(holdero.Faces.Select.Options)
		sort.Strings(holdero.Backs.Select.Options)
		sort.Strings(dreams.Theme.Select.Options)

		ld := []string{"Light", "Dark"}
		holdero.Faces.Select.Options = append(ld, holdero.Faces.Select.Options...)
		holdero.Backs.Select.Options = append(ld, holdero.Backs.Select.Options...)
		dreams.Theme.Select.Options = append([]string{"Main", "Legacy"}, dreams.Theme.Select.Options...)

		sort.Strings(menu.Assets.Assets)
		menu.Assets.Asset_list.Refresh()
		if menu.Control.Dapp_list["Holdero"] {
			holdero.DisableHolderoTools()
		}
	}
}

// If wallet owns dReams NFA, populate for use in dReams
//   - See games container in menu.PlaceAssets()
func checkNFAOwner(scid string) {
	if menu.Gnomes.IsRunning() {
		if header, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "nameHdr"); header != nil {
			owner, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "owner")
			file, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "fileURL")
			if owner != nil && file != nil {
				if owner[0] == rpc.Wallet.Address && menu.ValidNfa(file[0]) {
					check := strings.Trim(header[0], "0123456789")
					if check == "AZYDS" || check == "SIXART" {
						themes := dreams.Theme.Select.Options
						new_themes := append(themes, header[0])
						dreams.Theme.Select.Options = new_themes
						dreams.Theme.Select.Refresh()

						avatars := holdero.Settings.AvatarSelect.Options
						new_avatar := append(avatars, header[0])
						holdero.Settings.AvatarSelect.Options = new_avatar
						holdero.Settings.AvatarSelect.Refresh()
						menu.Assets.Assets = append(menu.Assets.Assets, header[0]+"   "+scid)
					} else if check == "AZYPCB" || check == "SIXPCB" {
						current := holdero.Backs.Select.Options
						new := append(current, header[0])
						holdero.Backs.Select.Options = new
						holdero.Backs.Select.Refresh()
						menu.Assets.Assets = append(menu.Assets.Assets, header[0]+"   "+scid)
					} else if check == "AZYPC" || check == "SIXPC" {
						current := holdero.Faces.Select.Options
						new := append(current, header[0])
						holdero.Faces.Select.Options = new
						holdero.Faces.Select.Refresh()
						menu.Assets.Assets = append(menu.Assets.Assets, header[0]+"   "+scid)
					} else if check == "DBC" {
						current := holdero.Settings.AvatarSelect.Options
						new := append(current, header[0])
						holdero.Settings.AvatarSelect.Options = new
						holdero.Settings.AvatarSelect.Refresh()
						menu.Assets.Assets = append(menu.Assets.Assets, header[0]+"   "+scid)
					} else if check == "HighStrangeness" {
						current_av := holdero.Settings.AvatarSelect.Options
						new_av := append(current_av, header[0])
						holdero.Settings.AvatarSelect.Options = new_av
						holdero.Settings.AvatarSelect.Refresh()
						menu.Assets.Assets = append(menu.Assets.Assets, header[0]+"   "+scid)

						var have_cards bool
						for _, face := range holdero.Faces.Select.Options {
							if face == "High-Strangeness" {
								have_cards = true
							}
						}

						if !have_cards {
							current_d := holdero.Faces.Select.Options
							new_d := append(current_d, "High-Strangeness")
							holdero.Faces.Select.Options = new_d
							holdero.Faces.Select.Refresh()

							current_b := holdero.Backs.Select.Options
							new_b := append(current_b, "High-Strangeness")
							holdero.Backs.Select.Options = new_b
							holdero.Backs.Select.Refresh()
						}

						tower := 0
						switch header[0] {
						case "HighStrangeness363":
							tower = 4
						case "HighStrangeness364":
							tower = 8
						case "HighStrangeness365":
							tower = 12
						default:
						}

						var have_theme bool
						for i := tower; i > 0; i-- {
							themes := dreams.Theme.Select.Options
							for _, th := range themes {
								if th == "HSTheme"+strconv.Itoa(i) {
									have_theme = true
								}
							}

							if !have_theme {
								new_themes := append(themes, "HSTheme"+strconv.Itoa(i))
								dreams.Theme.Select.Options = new_themes
								dreams.Theme.Select.Refresh()
							}
						}
					}
				}
			}
		}
	}
}

// Check if wallet owns in game G45 asset
//   - Pass g45s from db store, can be nil arg
//   - Pass false gc for rechecks
func CheckDreamsG45s(gc bool, g45s map[string]string) {
	if menu.Gnomes.IsReady() && !gc {
		if g45s == nil {
			g45s = menu.Gnomes.GetAllOwnersAndSCIDs()
		}
		log.Println("[dReams] Checking G45 Assets")

		for scid := range g45s {
			if !rpc.Wallet.IsConnected() || !menu.Gnomes.IsRunning() {
				break
			}

			if data, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "metadata"); data != nil {
				owner, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "owner")
				minter, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "minter")
				coll, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "collection")
				if owner != nil && minter != nil && coll != nil {
					if owner[0] == rpc.Wallet.Address {
						if minter[0] == dreams.Seals_mint && coll[0] == dreams.Seals_coll {
							var seal dreams.Seal
							if err := json.Unmarshal([]byte(data[0]), &seal); err == nil {
								menu.Assets.Assets = append(menu.Assets.Assets, seal.Name+"   "+scid)
								current := holdero.Settings.AvatarSelect.Options
								new := append(current, seal.Name)
								holdero.Settings.AvatarSelect.Options = new
								holdero.Settings.AvatarSelect.Refresh()
							}
						} else if minter[0] == dreams.ATeam_mint && coll[0] == dreams.ATeam_coll {
							var agent dreams.Agent
							if err := json.Unmarshal([]byte(data[0]), &agent); err == nil {
								menu.Assets.Asset_map[agent.Name] = scid
								menu.Assets.Assets = append(menu.Assets.Assets, agent.Name+"   "+scid)
								current := holdero.Settings.AvatarSelect.Options
								new := append(current, agent.Name)
								holdero.Settings.AvatarSelect.Options = new
								holdero.Settings.AvatarSelect.Refresh()
							}
						}
					}
				}
			}
		}
		sort.Strings(holdero.Settings.AvatarSelect.Options)
		holdero.Settings.AvatarSelect.Options = append([]string{"None"}, holdero.Settings.AvatarSelect.Options...)
		menu.Assets.Asset_list.Refresh()
	}
}

// Hidden object, controls Gnomon start and stop based on daemon connection
func DaemonConnectedBox() fyne.Widget {
	menu.Control.Daemon_check = widget.NewCheck("", func(b bool) {
		if !menu.Gnomes.IsInitialized() && !menu.Gnomes.Start {
			//go startLabel()
			menu.Assets.Gnomes_sync.Text = (" Starting Gnomon")
			menu.Assets.Gnomes_sync.Refresh()
			filters := menu.GnomonFilters()
			menu.StartGnomon("dReams", menu.Gnomes.DBType, filters, 3960, 490, menu.G45Index)
			rpc.FetchFees()
			if menu.Control.Dapp_list["Holdero"] {
				holdero.Poker.Contract_entry.CursorColumn = 1
				holdero.Poker.Contract_entry.Refresh()
			}

			if menu.Control.Dapp_list["dSports and dPredictions"] {
				menu.Control.P_contract.CursorColumn = 1
				menu.Control.P_contract.Refresh()
				menu.Control.S_contract.CursorColumn = 1
				menu.Control.S_contract.Refresh()
			}
		}

		if !b {
			go menu.StopLabel()
			menu.Gnomes.Stop("dReams")
			go menu.SleepLabel()
		}
	})
	menu.Control.Daemon_check.Disable()
	menu.Control.Daemon_check.Hide()

	return menu.Control.Daemon_check
}

// Wallet rpc entry object
//   - Bound to rpc.Wallet.Rpc
//   - Changes reset wallet connection and call CheckConnection()
func WalletRpcEntry() fyne.Widget {
	options := []string{"", "127.0.0.1:10103"}
	entry := widget.NewSelectEntry(options)
	entry.PlaceHolder = "Wallet RPC: "
	entry.OnCursorChanged = func() {
		if rpc.Wallet.IsConnected() {
			rpc.Wallet.Address = ""
			rpc.Display.Wallet_height = "0"
			rpc.Wallet.Height = 0
			rpc.Wallet.Connected(false)
			go CheckConnection()
		}
	}

	this := binding.BindString(&rpc.Wallet.Rpc)
	entry.Bind(this)

	return entry
}

// Authentication entry object
//   - Bound to rpc.Wallet.UserPass
//   - Changes call rpc.GetAddress() and CheckConnection()
func UserPassEntry() fyne.Widget {
	entry := widget.NewPasswordEntry()
	entry.PlaceHolder = "user:pass"
	entry.OnCursorChanged = func() {
		if rpc.Wallet.IsConnected() {
			rpc.GetAddress("dReams")
			go CheckConnection()
		}
	}

	a := binding.BindString(&rpc.Wallet.UserPass)
	entry.Bind(a)

	return entry
}

// Connect button object for rpc
//   - Pressed calls rpc.Ping(), rpc.GetAddress(), CheckConnection(),
//     checks for Holdero key and clears names for population
func RpcConnectButton() fyne.Widget {
	var wait bool
	button := widget.NewButton("Connect", func() {
		go func() {
			if !wait {
				wait = true
				rpc.Ping()
				rpc.GetAddress("dReams")
				CheckConnection()
				if menu.Control.Dapp_list["Holdero"] {
					holdero.Poker.Contract_entry.CursorColumn = 1
					holdero.Poker.Contract_entry.Refresh()
					if len(rpc.Wallet.Address) == 66 {
						holdero.CheckExistingKey()
						menu.Control.Names.ClearSelected()
						menu.Control.Names.Options = []string{}
						menu.Control.Names.Refresh()
						menu.Control.Names.Options = append(menu.Control.Names.Options, rpc.Wallet.Address[0:12])
						if menu.Control.Names.Options != nil {
							menu.Control.Names.SetSelectedIndex(0)
						}
					}
				}

				if menu.Control.Dapp_list["dSports and dPredictions"] {
					menu.Control.P_contract.CursorColumn = 1
					menu.Control.P_contract.Refresh()
					menu.Control.S_contract.CursorColumn = 1
					menu.Control.S_contract.Refresh()
				}
				wait = false
			}
		}()
	})

	return button
}

// dReams recheck owned assets routine
func RecheckDreamsAssets() {
	menu.Gnomes.Wait = true
	menu.Assets.Assets = []string{}
	CheckDreamsNFAs(false, nil)
	CheckDreamsG45s(false, nil)
	if menu.Control.Dapp_list["Holdero"] {
		if rpc.Wallet.IsConnected() {
			menu.Control.Names.Options = []string{rpc.Wallet.Address[0:12]}
			menu.CheckWalletNames(rpc.Wallet.Address)
		}
	}
	sort.Strings(menu.Assets.Assets)
	menu.Assets.Asset_list.UnselectAll()
	menu.Assets.Asset_list.Refresh()
	menu.Gnomes.Wait = false
}

// Recheck owned assets button
//   - tag for log print
//   - pass recheck for desired check
func RecheckButton(tag string, recheck func()) (button fyne.Widget) {
	button = widget.NewButton("Check Assets", func() {
		if !menu.Gnomes.Wait {
			log.Printf("[%s] Rechecking Assets\n", tag)
			go recheck()
		}
	})

	return
}
