package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/SixofClubsss/Baccarat/baccarat"
	"github.com/SixofClubsss/Duels/duel"
	"github.com/SixofClubsss/Grokked/grok"
	"github.com/SixofClubsss/Holdero/holdero"
	"github.com/SixofClubsss/Iluma/tarot"
	"github.com/SixofClubsss/dPrediction/prediction"
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
  --fastsync=<true>	Gnomon option,  true/false value to define loading at chain height on start up.
  --num-parallel-blocks=<1>   Gnomon option,  defines the number of parallel blocks to index.
  --dbtype=<boltdb>     Gnomon option,  defines type of database 'gravdb' or 'boltdb'.`

// Set opts when starting dReams
func flags() {
	arguments, err := docopt.ParseArgs(command_line, nil, rpc.Version().String())
	if err != nil {
		logger.Fatalf("Error while parsing arguments: %s\n", err)
	}

	if dReams.OS() == "linux" {
		fmt.Println(string(bundle.ResourceStampTxt.StaticContent))
	}

	if arguments["--dbtype"] != nil {
		if arguments["--dbtype"] == "gravdb" {
			menu.Gnomes.DBType = arguments["--dbtype"].(string)
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
}

func init() {
	dReams.SetOS()
	menu.InitLogrusLog(logrus.InfoLevel)
	saved := menu.ReadDreamsConfig("dReams")
	if saved.Daemon != nil {
		menu.Control.Daemon_config = saved.Daemon[0]
	}

	holdero.SetFavoriteTables(saved.Tables)
	prediction.Predict.Favorites.SCIDs = saved.Predict
	prediction.Sports.Favorites.SCIDs = saved.Sports

	menu.Market.DreamsFilter = true

	rpc.InitBalances()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		menu.CloseAppSignal(true)
		menu.WriteDreamsConfig(save())
		fmt.Println()
		dappCloseCheck()
		menu.Info.SetStatus("Putting Gnomon to Sleep")
		menu.Gnomes.Stop("dReams")
		menu.StopIndicators(indicators)
		time.Sleep(time.Second)
		dReams.StopProcess()
		dReams.Window.Close()
	}()
}

// Build save struct for local preferences
func save() dreams.SaveData {
	return dreams.SaveData{
		Skin:    bundle.AppColor,
		Daemon:  []string{rpc.Daemon.Rpc},
		Tables:  holdero.GetFavoriteTables(),
		Predict: prediction.Predict.Favorites.SCIDs,
		Sports:  prediction.Sports.Favorites.SCIDs,
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

// Make system tray with opts
//   - Send Dero message menu
//   - Explorer link
//   - Manual reveal key for Holdero
func systemTray(w fyne.App) bool {
	if desk, ok := w.(desktop.App); ok {
		m := fyne.NewMenu("MyApp",
			fyne.NewMenuItem("Send Message", func() {
				if !dReams.IsConfiguring() {
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
						go menu.Info.RefreshPrice(App_Name)
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

// Refresh all menu gui objects
func menuRefresh(offset int) {
	if dReams.OnTab("Menu") && menu.Gnomes.IsInitialized() {
		switch menu.Gnomes.Status() {
		case "initializing":
			menu.Info.SetStatus("Gnomon Initializing")
		case "fastsyncing":
			menu.Info.SetStatus("Gnomon Fastsyncing...")
		case "closing":
			menu.Info.SetStatus("Gnomon Closing...")
		case "indexed":
			if !menu.Gnomes.HasIndex(uint64(menu.ReturnAssetCount())) && !menu.Gnomes.HasChecked() {
				menu.Info.SetStatus("Gnomon Syncing...")
			} else {
				menu.Info.SetStatus("Gnomon Synced")
			}
		case "indexing":
			menu.Info.SetStatus("Gnomon Syncing...")
		}

		if offset == 20 {
			go menu.Info.RefreshPrice(App_Name)
		}

		if offset%3 == 0 && dReams.OnSubTab("Market") && !dReams.IsWindows() && !menu.ClosingApps() {
			menu.FindNFAListings(nil)
		}
	}

	menu.Assets.Stats_box = *container.NewVBox(menu.Assets.Collection, menu.Assets.Name, menu.IconImg(bundle.ResourceAvatarFramePng))
	menu.Assets.Stats_box.Refresh()

	// Update live market info
	if len(menu.Market.Viewing) == 64 && rpc.Wallet.IsConnected() {
		if menu.Market.Tab == "Buy" {
			menu.GetBuyNowDetails(menu.Market.Viewing)
			go menu.RefreshNFAImages()
		} else {
			menu.GetAuctionDetails(menu.Market.Viewing)
			go menu.RefreshNFAImages()
		}
	}

	menu.Info.RefreshDaemon(App_Name)
	menu.Info.RefreshGnomon()
	menu.Info.RefreshWallet()
	menu.Info.RefreshIndexed()

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
		menu.Info.SetStatus("Checking for Assets")
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
		dreams.Theme.Sort()
		dreams.Theme.Select.Options = append([]string{"Main", "Legacy"}, dreams.Theme.Select.Options...)
		sort.Strings(menu.Assets.Assets)
		menu.Assets.Asset_list.Refresh()
		if menu.DappEnabled("Duels") {
			duel.Inventory.SortAll()
		}
		if menu.DappEnabled("Holdero") {
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
			collection, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "collection")
			creator, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "creatorAddr")
			if owner != nil && file != nil && collection != nil && creator != nil {
				if owner[0] == rpc.Wallet.Address && menu.ValidNFA(file[0]) {
					if !menu.IsDreamsNFACreator(creator[0]) {
						return
					}

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
						if menu.DappEnabled("Duels") {
							duel.AddItemsToInventory(scid, header[0], owner[0], collection[0])
						}
					} else if collection[0] == "Dorblings NFA" {
						holdero.Settings.AddAvatar(header[0], owner[0])
						menu.Assets.Add(header[0], scid)
					} else if collection[0] == "DLAMPP" {
						// TODO review after mint
						holdero.Settings.AddAvatar(header[0], owner[0])
						menu.Assets.Add(header[0], scid)
					} else if collection[0] == "High Strangeness" {
						holdero.Settings.AddAvatar(header[0], owner[0])
						menu.Assets.Add(header[0], scid)
						hsCards(owner[0], header[0], check)
						if menu.DappEnabled("Duels") {
							duel.AddItemsToInventory(scid, header[0], owner[0], collection[0])
						}
					} else if collection[0] == "Dero Desperados" {
						holdero.Settings.AddAvatar(header[0], owner[0])
						menu.Assets.Add(header[0], scid)
						if menu.DappEnabled("Duels") {
							duel.AddItemsToInventory(scid, header[0], owner[0], collection[0])
						}
					} else if collection[0] == "Desperado Guns" {
						holdero.Settings.AddAvatar(header[0], owner[0])
						menu.Assets.Add(header[0], scid)
						if menu.DappEnabled("Duels") {
							duel.AddItemsToInventory(scid, header[0], owner[0], collection[0])
						}
					}
				}
			}
		}
	}
}

func hsCards(owner, name, check string) {
	var have_cards bool
	for _, face := range holdero.Settings.CurrentFaces() {
		if face == "HS_Deck" {
			have_cards = true
			break
		}
	}

	if !have_cards {
		holdero.Settings.AddFaces("HS_Deck", owner)
		holdero.Settings.AddBacks("HS_Back", owner)
		holdero.Settings.AddBacks("HS_Back2", owner)
		holdero.Settings.AddBacks("HS_Back3", owner)
		holdero.Settings.AddBacks("HS_Back4", owner)
		holdero.Settings.AddBacks("HS_Back5", owner)
	}

	if check == "GOLDCard" {
		var have_gold bool
		for _, back := range holdero.Settings.CurrentBacks() {
			if back == "HS_Back7" {
				have_gold = true
				break
			}
		}

		if !have_gold {
			holdero.Settings.AddBacks("HS_Back7", owner)
		}
	}

	tower := 0
	switch name {
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
								menu.Assets.Add(agent.Name, scid)
								holdero.Settings.AddAvatar(agent.Name, owner[0])
							}
						} else if minter[0] == menu.Degen_mint && coll[0] == menu.Degen_coll {
							var degen menu.Degen
							if err := json.Unmarshal([]byte(data[0]), &degen); err == nil {
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
	holdero.Disconnected(menu.DappEnabled("Holdero"))
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
	if menu.DappEnabled("Holdero") {
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

	if menu.DappEnabled("Baccarat") {
		bacc := rpc.GetSCCode(rpc.BaccSCID)
		if bacc != "" {
			filter = append(filter, bacc)
		}
	}

	if menu.DappEnabled("dSports and dPredictions") {
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

	// if menu.DappEnabled("DerBnb") {
	// 	bnb := rpc.GetSCCode(rpc.DerBnbSCID)
	// 	if bnb != "" {
	// 		filter = append(filter, bnb)
	// 	}
	// }

	if menu.DappEnabled("Duels") {
		duels := rpc.GetSCCode(duel.DUELSCID)
		if duels != "" {
			filter = append(filter, duels)
		}
	}

	if menu.DappEnabled("Grokked") {
		grok := rpc.GetSCCode(grok.GROKSCID)
		if grok != "" {
			filter = append(filter, grok)
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

			menu.Info.SetStatus("Starting Gnomon")
			rpc.FetchFees()
			filters := gnomonFilters()
			menu.StartGnomon("dReams", menu.Gnomes.DBType, filters, menu.Control.G45_count+menu.Control.NFA_count, menu.Control.NFA_count, menu.G45Index)

			if menu.DappEnabled("dSports and dPredictions") {
				prediction.OnConnected()
			}
		}

		if !b {
			menu.Info.SetStatus("Putting Gnomon to Sleep")
			menu.Gnomes.Stop("dReams")
			menu.Info.SetStatus("Gnomon is Sleeping")
		}
	})
	menu.Control.Daemon_check.Disable()
	menu.Control.Daemon_check.Hide()

	return menu.Control.Daemon_check
}

// Daemon rpc entry object with default options
//   - Bound to rpc.Daemon.Rpc
func daemonRpcEntry() fyne.Widget {
	options := []string{
		"",
		rpc.DAEMON_RPC_DEFAULT,
		rpc.DAEMON_RPC_REMOTE1,
		rpc.DAEMON_RPC_REMOTE2,
		rpc.DAEMON_RPC_REMOTE3,
		rpc.DAEMON_RPC_REMOTE4,
		rpc.DAEMON_RPC_REMOTE5,
		rpc.DAEMON_RPC_REMOTE6,
	}

	if menu.Control.Daemon_config != "" {
		options = append(options, menu.Control.Daemon_config)
	}
	entry := widget.NewSelectEntry(options)
	entry.PlaceHolder = "Daemon RPC: "

	this := binding.BindString(&rpc.Daemon.Rpc)
	entry.Bind(this)

	return entry
}

// Wallet rpc entry object
//   - Bound to rpc.Wallet.Rpc
//   - Changes reset wallet connection and call checkConnection()
func walletRpcEntry() fyne.Widget {
	options := []string{"", "127.0.0.1:10103"}
	entry := widget.NewSelectEntry(options)
	entry.PlaceHolder = "Wallet RPC: "
	entry.OnChanged = func(s string) {
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
	entry.OnChanged = func(s string) {
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
				if menu.DappEnabled("Holdero") {
					holdero.OnConnected()
				}

				if menu.DappEnabled("dSports and dPredictions") {
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
	if menu.DappEnabled("Duels") {
		duel.Inventory.ClearAll()
	}
	checkDreamsNFAs(false, nil)
	checkDreamsG45s(false, nil)
	if menu.DappEnabled("Holdero") {
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

// Returns map of current dApp package versions
func dappVersions(dapps []string) map[string]string {
	versions := make(map[string]string)
	versions["NFA Market"] = rpc.Version().String()
	versions["Gnomon"] = structures.Version.String()
	for _, pkg := range dapps {
		switch pkg {
		case "Holdero":
			versions["Holdero"] = holdero.Version().String()
		case "Baccarat":
			versions["Baccarat"] = baccarat.Version().String()
		case "dSports and dPredictions":
			versions["dSports and dPredictions"] = prediction.Version().String()
		case "Iluma":
			versions["Iluma"] = tarot.Version().String()
		case "Duels":
			versions["Duels"] = duel.Version().String()
		case "Grokked":
			versions["Grokked"] = grok.Version().String()
		}
	}

	return versions
}
