package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	holdero "github.com/SixofClubsss/Holdero"
	prediction "github.com/SixofClubsss/dPrediction"
	"github.com/civilware/Gnomon/indexer"
	"github.com/civilware/Gnomon/structures"
	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/menu"
	"github.com/dReam-dApps/dReams/rpc"
	"github.com/docopt/docopt-go"
	"github.com/fyne-io/terminal"
	"github.com/sirupsen/logrus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type cliApp struct {
	term    *terminal.Terminal
	enabled bool
}

var cli cliApp
var logger = structures.Logger.WithFields(logrus.Fields{})
var command_line string = `dReams
Platform for Dero dApps, powered by Gnomon.

Usage:
  dReams [options]
  dReams -h | --help

Options:
  -h --help             Show this screen.
  --cli=<false>		dReams option, enables cli app tab.
  --trim=<true>	        dReams option, defaults true for minimum index search filters.
  --fastsync=<true>	Gnomon option,  true/false value to define loading at chain height on start up.
  --num-parallel-blocks=<1>   Gnomon option,  defines the number of parallel blocks to index.
  --dbtype=<boltdb>     Gnomon option,  defines type of database 'gravdb' or 'boltdb'.`

// Set opts when starting dReams
func flags() (version string) {
	version = rpc.DREAMSv
	arguments, err := docopt.ParseArgs(command_line, nil, version)

	if err != nil {
		logger.Fatalf("Error while parsing arguments: %s\n", err)
	}

	if arguments["--dbtype"] != nil {
		if arguments["--dbtype"] == "gravdb" {
			menu.Gnomes.DBType = arguments["--dbtype"].(string)
		}
	}

	menu.Gnomes.Trim = true
	if arguments["--trim"] != nil {
		if arguments["--trim"].(string) == "false" {
			menu.Gnomes.Trim = false
		}
	}

	menu.Gnomes.Fast = true
	if arguments["--fastsync"] != nil {
		if arguments["--fastsync"].(string) == "false" {
			menu.Gnomes.Fast = false
		}
	}

	if arguments["--num-parallel-blocks"] != nil {
		s := arguments["--num-parallel-blocks"].(string)
		switch s {
		case "2":
			menu.Gnomes.Para = 2
		case "3":
			menu.Gnomes.Para = 3
		case "4":
			menu.Gnomes.Para = 4
		case "5":
			menu.Gnomes.Para = 5
		default:
			menu.Gnomes.Para = 1
		}
	}

	cli.enabled = false
	if arguments["--cli"] != nil {
		if arguments["--cli"].(string) == "true" {
			cli.enabled = true
		}
	}

	return
}

func init() {
	arguments := make(map[string]interface{})
	arguments["--debug"] = false
	indexer.InitLog(arguments, os.Stdout)
	saved := menu.ReadDreamsConfig("dReams")
	if saved.Daemon != nil {
		menu.Control.Daemon_config = saved.Daemon[0]
	}

	holdero.Settings.Favorites = saved.Tables
	prediction.Predict.Settings.Favorites = saved.Predict
	prediction.Sports.Settings.Favorites = saved.Sports

	menu.Market.DreamsFilter = true

	rpc.InitBalances()

	dReams.SetOS()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		menu.CloseAppSignal(true)
		menu.WriteDreamsConfig(save())
		fmt.Println()
		dappCloseCheck()
		go menu.StopLabel()
		menu.Gnomes.Stop("dReams")
		menu.StopIndicators(indicators)
		time.Sleep(time.Second)
		dReams.StopProcess()
		dReams.Window.Close()
	}()
}

// Build save struct for local preferences
func save() dreams.DreamSave {
	return dreams.DreamSave{
		Skin:    bundle.AppColor,
		Daemon:  []string{rpc.Daemon.Rpc},
		Tables:  holdero.Settings.Favorites,
		Predict: prediction.Predict.Settings.Favorites,
		Sports:  prediction.Sports.Settings.Favorites,
		DBtype:  menu.Gnomes.DBType,
		Para:    menu.Gnomes.Para,
		Assets:  menu.Control.Enabled_assets,
		Dapps:   menu.Control.Dapp_list,
	}
}

// Starts a Fyne terminal in dReams
func startTerminal() *terminal.Terminal {
	cli.term = terminal.New()
	go func() {
		_ = cli.term.RunLocalShell()
	}()

	return cli.term
}

// Exit running dReams terminal
func exitTerminal() {
	if cli.term != nil {
		cli.term.Exit()
	}
}

// Terminal start info, ascii art for linux
func stamp(v string) {
	if dReams.OS() == "linux" {
		fmt.Println(string(bundle.ResourceStampTxt.StaticContent))
	}
	logger.Println("[dReams]", v, runtime.GOOS, runtime.GOARCH)
}

// Make system tray with opts
//   - Send Dero message menu
//   - Explorer link
//   - Manual reveal key for Holdero
func systemTray(w fyne.App) bool {
	if desk, ok := w.(desktop.App); ok {
		m := fyne.NewMenu("MyApp",
			fyne.NewMenuItem("Send Message", func() {
				if !dReams.IsConfiguring() && rpc.Wallet.IsConnected() {
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

// This is what we want to scan wallet for when Gnomon is synced
func gnomonScan(contracts map[string]string) {
	checkDreamsNFAs(menu.Gnomes.Check, contracts)
	checkDreamsG45s(menu.Gnomes.Check, contracts)
}

// Main dReams process loop
func fetch(done chan struct{}) {
	var offset int
	rpc.Startup = true
	time.Sleep(3 * time.Second)
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ticker.C: // do on interval
			if !dReams.IsConfiguring() {
				rpc.Ping()
				rpc.EchoWallet("dReams")
				go rpc.GetDreamsBalances(rpc.SCIDs)
				rpc.GetWalletHeight("dReams")
				if !rpc.Startup {
					checkConnection()
					menu.GnomonEndPoint()
					menu.GnomonState(dReams.IsConfiguring(), gnomonScan)
					dReams.Background.Refresh()

					go menuRefresh(offset)

					offset++
					if offset >= 21 {
						offset = 0
					}
				}

				if rpc.Daemon.IsConnected() {
					if rpc.Startup {
						go refreshPriceDisplay(true)
					}

					rpc.Startup = false
				}

				dReams.SignalChannel()

			}
		case <-dReams.Closing(): // exit loop
			logger.Println("[dReams] Closing...")
			ticker.Stop()
			dReams.CloseAllDapps()
			time.Sleep(time.Second)
			done <- struct{}{}
			return
		}
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
	if c {
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
		menu.Assets.Wall_height.Text = (" Wallet Height: " + rpc.Wallet.Display.Height)
		menu.Assets.Wall_height.Refresh()
	} else {
		menu.Assets.Wall_height.Text = (" Wallet Height: 0")
		menu.Assets.Wall_height.Refresh()
	}
}

// Refresh current Dero-USDT price
func refreshPriceDisplay(c bool) {
	if c && rpc.Daemon.IsConnected() {
		_, price := menu.GetPrice("DERO-USDT", "dReams")
		menu.Assets.Dero_price.Text = (" Dero Price: $" + price)
		menu.Assets.Dero_price.Refresh()
	} else {
		menu.Assets.Dero_price.Text = (" Dero Price: $")
		menu.Assets.Dero_price.Refresh()
	}
}

// Refresh all menu gui objects
func menuRefresh(offset int) {
	if dReams.OnTab("Menu") && menu.Gnomes.IsInitialized() {
		index := menu.Gnomes.Indexer.LastIndexedHeight
		if index < menu.Gnomes.Indexer.ChainHeight-4 || !menu.Gnomes.HasIndex(uint64(menu.ReturnAssetCount())) {
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

		if offset%3 == 0 && dReams.OnSubTab("Market") && !dReams.IsWindows() && !menu.ClosingApps() {
			menu.FindNfaListings(nil)
		}
	}

	menu.Assets.Stats_box = *container.NewVBox(menu.Assets.Collection, menu.Assets.Name, menu.IconImg(bundle.ResourceAvatarFramePng))
	menu.Assets.Stats_box.Refresh()

	// Update live market info
	if len(menu.Market.Viewing) == 64 && rpc.Wallet.IsConnected() {
		if menu.Market.Tab == "Buy" {
			menu.GetBuyNowDetails(menu.Market.Viewing)
			go menu.RefreshNfaImages()
		} else {
			menu.GetAuctionDetails(menu.Market.Viewing)
			go menu.RefreshNfaImages()
		}
	}

	if rpc.Daemon.IsConnected() {
		go refreshDaemonDisplay(true)
	} else {
		go refreshDaemonDisplay(false)
		go refreshGnomonDisplay(0, 0)
		go refreshIndexDisplay(false)
	}

	if rpc.Wallet.IsConnected() {
		go refreshWalletDisplay(true)
	} else {
		go refreshWalletDisplay(false)
	}

	menu.Assets.Balances.Refresh()

	if !dReams.OnTab("Menu") {
		menu.Market.Viewing = ""
		menu.Market.Viewing_coll = ""
	}
}

// Check wallet for dReams NFAs
//   - Pass scids from db store, can be nil arg
//   - Pass false gc for rechecks
func checkDreamsNFAs(gc bool, scids map[string]string) {
	if menu.Gnomes.IsReady() && !gc {
		menu.Assets.Gnomes_sync.Text = (" Checking for Assets")
		menu.Assets.Gnomes_sync.Refresh()
		if scids == nil {
			scids = menu.Gnomes.GetAllOwnersAndSCIDs()
		}

		logger.Println("[dReams] Checking NFA Assets")
		dreams.Theme.Select.Options = []string{}
		holdero.Settings.ClearAssets()

		for sc := range scids {
			if !rpc.Wallet.IsConnected() || !menu.Gnomes.IsRunning() {
				break
			}

			checkNFAOwner(sc)
		}

		holdero.Settings.SortCardAsset()
		sort.Strings(dreams.Theme.Select.Options)
		dreams.Theme.Select.Options = append([]string{"Main", "Legacy"}, dreams.Theme.Select.Options...)
		sort.Strings(menu.Assets.Assets)
		menu.Assets.Asset_list.Refresh()
		if menu.Control.Dapp_list["Holdero"] {
			holdero.DisableHolderoTools()
		}
	}
}

// If wallet owns dReams NFA, populate for use in dReams
//   - See asset_selects container in menu.PlaceAssets()
func checkNFAOwner(scid string) {
	if menu.Gnomes.IsRunning() {
		if header, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "nameHdr"); header != nil {
			owner, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "owner")
			file, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "fileURL")
			if owner != nil && file != nil {
				if owner[0] == rpc.Wallet.Address && menu.ValidNfa(file[0]) {
					check := strings.Trim(header[0], "0123456789")
					if check == "AZYDS" || check == "SIXART" {
						dreams.Theme.Add(header[0], owner[0])
						holdero.Settings.AddAvatar(header[0], owner[0])
						menu.Assets.Add(header[0], scid)
					} else if check == "AZYPCB" || check == "SIXPCB" {
						holdero.Settings.AddBacks(header[0], owner[0])
						menu.Assets.Add(header[0], scid)
					} else if check == "AZYPC" || check == "SIXPC" {
						holdero.Settings.AddFaces(header[0], owner[0])
						menu.Assets.Add(header[0], scid)
					} else if check == "DBC" {
						holdero.Settings.AddAvatar(header[0], owner[0])
						menu.Assets.Add(header[0], scid)
					} else if check == "HighStrangeness" {
						holdero.Settings.AddAvatar(header[0], owner[0])
						menu.Assets.Add(header[0], scid)

						var have_cards bool
						for _, face := range holdero.Settings.CurrentFaces() {
							if face == "High-Strangeness" {
								have_cards = true
							}
						}

						if !have_cards {
							holdero.Settings.AddFaces("High-Strangeness", owner[0])
							holdero.Settings.AddBacks("High-Strangeness", owner[0])
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
func checkDreamsG45s(gc bool, g45s map[string]string) {
	if menu.Gnomes.IsReady() && !gc {
		if g45s == nil {
			g45s = menu.Gnomes.GetAllOwnersAndSCIDs()
		}
		logger.Println("[dReams] Checking G45 Assets")

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
						if minter[0] == menu.Seals_mint && coll[0] == menu.Seals_coll {
							var seal menu.Seal
							if err := json.Unmarshal([]byte(data[0]), &seal); err == nil {
								menu.Assets.Add(seal.Name, scid)
								holdero.Settings.AddAvatar(seal.Name, owner[0])
							}
						} else if minter[0] == menu.ATeam_mint && coll[0] == menu.ATeam_coll {
							var agent menu.Agent
							if err := json.Unmarshal([]byte(data[0]), &agent); err == nil {
								menu.Assets.Asset_map[agent.Name] = scid
								menu.Assets.Add(agent.Name, scid)
								holdero.Settings.AddAvatar(agent.Name, owner[0])
							}
						} else if minter[0] == menu.Degen_mint && coll[0] == menu.Degen_coll {
							var degen menu.Degen
							if err := json.Unmarshal([]byte(data[0]), &degen); err == nil {
								menu.Assets.Asset_map[degen.Name] = scid
								menu.Assets.Add(degen.Name, scid)
								holdero.Settings.AddAvatar(degen.Name, owner[0])
							}
						}
					}
				}
			}
		}
		holdero.Settings.SortAvatarAsset()
		menu.Assets.Asset_list.Refresh()
	}
}

// Connection check for main process
func checkConnection() {
	if rpc.Daemon.IsConnected() {
		menu.Control.Daemon_check.SetChecked(true)
		menu.DisableIndexControls(false)
	} else {
		menu.Control.Daemon_check.SetChecked(false)
		disableActions(true)
		disconnected()
		menu.DisableIndexControls(true)
	}

	if rpc.Wallet.IsConnected() {
		if rpc.Daemon.IsConnected() {
			disableActions(false)
		}
	} else {
		disableActions(true)
		disconnected()
		menu.Gnomes.Checked(false)
	}
}

// Do when disconnected
func disconnected() {
	menu.Market.Auctions = []string{}
	menu.Market.Buy_now = []string{}
	holdero.Disconnected(menu.Control.Dapp_list["Holdero"])
	prediction.Disconnected()
	rpc.Wallet.Address = ""
	dreams.Theme.Select.Options = []string{"Main", "Legacy"}
	dreams.Theme.Select.Refresh()
	menu.Assets.Assets = []string{}
	menu.Assets.Name.Text = (" Name:")
	menu.Assets.Name.Refresh()
	menu.Assets.Collection.Text = (" Collection:")
	menu.Assets.Collection.Refresh()
	menu.Assets.Icon = *canvas.NewImageFromImage(nil)
	menu.Market.Auction_list.UnselectAll()
	menu.Market.Buy_list.UnselectAll()
	menu.Market.Icon = *canvas.NewImageFromImage(nil)
	menu.Market.Cover = *canvas.NewImageFromImage(nil)
	menu.Market.Viewing = ""
	menu.Market.Viewing_coll = ""
	menu.ResetAuctionInfo()
	menu.AuctionInfo()
}

// Disable actions requiring connection
func disableActions(d bool) {
	if d {
		menu.Assets.Swap.Hide()
	} else {
		menu.Assets.Swap.Show()
	}

	menu.Assets.Swap.Refresh()
}

// dReams search filters for Gnomon index
func gnomonFilters() (filter []string) {
	if menu.Control.Dapp_list["Holdero"] {
		holdero110 := rpc.GetSCCode(holdero.HolderoSCID)
		if holdero110 != "" {
			filter = append(filter, holdero110)
		}

		holdero100 := rpc.GetSCCode(holdero.Holdero100)
		if holdero100 != "" {
			filter = append(filter, holdero100)
		}

		holderoHGC := rpc.GetSCCode(holdero.HGCHolderoSCID)
		if holderoHGC != "" {
			filter = append(filter, holderoHGC)
		}
	}

	if menu.Control.Dapp_list["Baccarat"] {
		bacc := rpc.GetSCCode(rpc.BaccSCID)
		if bacc != "" {
			filter = append(filter, bacc)
		}
	}

	if menu.Control.Dapp_list["dSports and dPredictions"] {
		predict := rpc.GetSCCode(prediction.PredictSCID)
		if predict != "" {
			filter = append(filter, predict)
		}

		sports := rpc.GetSCCode(prediction.SportsSCID)
		if sports != "" {
			filter = append(filter, sports)
		}
	}

	gnomon := rpc.GetSCCode(rpc.GnomonSCID)
	if gnomon != "" {
		filter = append(filter, gnomon)
	}

	names := rpc.GetSCCode(rpc.NameSCID)
	if names != "" {
		filter = append(filter, names)
	}

	ratings := rpc.GetSCCode(rpc.RatingSCID)
	if ratings != "" {
		filter = append(filter, ratings)
	}

	if menu.Control.Dapp_list["DerBnb"] {
		bnb := rpc.GetSCCode(rpc.DerBnbSCID)
		if bnb != "" {
			filter = append(filter, bnb)
		}
	}

	filter = append(filter, menu.ReturnEnabledNFAs(menu.Control.Enabled_assets)...)

	return
}

// Hidden object, controls Gnomon start and stop based on daemon connection
func daemonConnectedBox() fyne.Widget {
	menu.Control.Daemon_check = widget.NewCheck("", func(b bool) {
		if !menu.Gnomes.IsInitialized() && !menu.Gnomes.Start {
			if rpc.DaemonVersion() == "3.5.3-139.DEROHE.STARGATE+04042023" {
				dialog.NewInformation("Daemon Version", "This daemon may conflict with Gnomon sync", dReams.Window).Show()
			}

			menu.Assets.Gnomes_sync.Text = (" Starting Gnomon")
			menu.Assets.Gnomes_sync.Refresh()
			filters := gnomonFilters()
			menu.StartGnomon("dReams", menu.Gnomes.DBType, filters, menu.Control.G45_count+menu.Control.NFA_count, menu.Control.NFA_count, menu.G45Index)
			rpc.FetchFees()

			if menu.Control.Dapp_list["dSports and dPredictions"] {
				prediction.OnConnected()
			}
		}

		if !b {
			go menu.StopLabel()
			menu.Gnomes.Stop("dReams")
			menu.Assets.Gnomes_sync.Text = (" Gnomon is Sleeping")
			menu.Assets.Gnomes_sync.Refresh()
		}
	})
	menu.Control.Daemon_check.Disable()
	menu.Control.Daemon_check.Hide()

	return menu.Control.Daemon_check
}

// Wallet rpc entry object
//   - Bound to rpc.Wallet.Rpc
//   - Changes reset wallet connection and call checkConnection()
func walletRpcEntry() fyne.Widget {
	options := []string{"", "127.0.0.1:10103"}
	entry := widget.NewSelectEntry(options)
	entry.PlaceHolder = "Wallet RPC: "
	entry.OnCursorChanged = func() {
		if rpc.Wallet.IsConnected() {
			rpc.Wallet.Address = ""
			rpc.Wallet.Display.Height = "0"
			rpc.Wallet.Height = 0
			rpc.Wallet.Connected(false)
			go checkConnection()
		}
	}

	entry.Bind(binding.BindString(&rpc.Wallet.Rpc))

	return entry
}

// Authentication entry object
//   - Bound to rpc.Wallet.UserPass
//   - Changes call rpc.GetAddress() and checkConnection()
func userPassEntry() fyne.Widget {
	entry := widget.NewPasswordEntry()
	entry.PlaceHolder = "user:pass"
	entry.OnCursorChanged = func() {
		if rpc.Wallet.IsConnected() {
			rpc.GetAddress("dReams")
			go checkConnection()
		}
	}

	entry.Bind(binding.BindString(&rpc.Wallet.UserPass))

	return entry
}

// Connect button object for rpc
//   - Pressed calls rpc.Ping(), rpc.GetAddress(), checkConnection(),
//   - dapp.OnConnected() funcs get called here
func rpcConnectButton() fyne.Widget {
	var wait bool
	button := widget.NewButton("Connect", func() {
		go func() {
			if !wait {
				wait = true
				rpc.Ping()
				rpc.GetAddress("dReams")
				checkConnection()
				if menu.Control.Dapp_list["Holdero"] {
					holdero.OnConnected()
				}

				if menu.Control.Dapp_list["dSports and dPredictions"] {
					prediction.OnConnected()
				}
				wait = false

				return
			}

			if !rpc.Wallet.IsConnected() {
				logger.Warnf("[dReams] Syncing, please wait")
			}
		}()
	})

	return button
}

// dReams recheck owned assets routine
func recheckDreamsAssets() {
	menu.Gnomes.Wait = true
	menu.Assets.Assets = []string{}
	checkDreamsNFAs(false, nil)
	checkDreamsG45s(false, nil)
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
func recheckButton(tag string, recheck func()) (button fyne.Widget) {
	button = widget.NewButton("Check Assets", func() {
		if !menu.Gnomes.Wait {
			logger.Printf("[%s] Rechecking Assets\n", tag)
			go recheck()
		}
	})

	return
}

func dappCloseCheck() {
	prediction.Service.IsStopped()
}
