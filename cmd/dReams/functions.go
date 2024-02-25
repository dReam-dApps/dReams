package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/SixofClubsss/Baccarat/baccarat"
	"github.com/SixofClubsss/Duels/duel"
	"github.com/SixofClubsss/Grokked/grok"
	"github.com/SixofClubsss/Holdero/holdero"
	"github.com/SixofClubsss/Iluma/tarot"
	"github.com/SixofClubsss/dDice/dice"
	"github.com/SixofClubsss/dPrediction/prediction"
	"github.com/civilware/Gnomon/structures"
	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/gnomes"
	"github.com/dReam-dApps/dReams/menu"
	"github.com/dReam-dApps/dReams/rpc"
	"github.com/deroproject/derohe/config"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/walletapi/xswd"
	"github.com/docopt/docopt-go"
	"github.com/sirupsen/logrus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var logger = structures.Logger.WithFields(logrus.Fields{})
var command_line string = `dReams
Platform for Dero dApps, powered by Gnomon.

Usage:
  dReams [options]
  dReams -h | --help

Options:
  -h --help             Show this screen.
  --num-parallel-blocks=<1>   Gnomon option,  defines the number of parallel blocks to index.
  --dbtype=<boltdb>     Gnomon option,  defines type of database 'gravdb' or 'boltdb'.
  --testnet=<false>     DERO option, defines if globals.Config is mainnet/testnet.`

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
			gnomon.SetDBStorageType(arguments["--dbtype"].(string))
		}
	}

	if arguments["--num-parallel-blocks"] != nil {
		s := arguments["--num-parallel-blocks"].(string)
		switch s {
		case "2":
			gnomon.SetParallel(2)
		case "3":
			gnomon.SetParallel(3)
		case "4":
			gnomon.SetParallel(4)
		case "5":
			gnomon.SetParallel(5)
		default:
			gnomon.SetParallel(1)
		}
	}

	if arguments["--testnet"] != nil {
		if arguments["--testnet"].(string) == "true" {
			globals.Config = config.Testnet
		}
	}
}

func init() {
	dReams.SetOS()
	gnomes.InitLogrusLog(logrus.InfoLevel)
	saved := menu.ReadDreamsConfig("dReams")
	if saved.Daemon != nil {
		menu.Control.Daemon = saved.Daemon[0]
	}

	holdero.SetFavoriteTables(saved.Tables)
	prediction.Predict.Favorites.SCIDs = saved.Predict
	prediction.Sports.Favorites.SCIDs = saved.Sports

	menu.Market.DreamsFilter = true

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		menu.SetClose(true)
		menu.WriteDreamsConfig(save())
		fmt.Println()
		dappCloseCheck()
		menu.Info.SetStatus("Putting Gnomon to Sleep")
		gnomon.Stop("dReams")
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
		Theme:   dreams.Theme.Name,
		FSForce: gnomon.GetFastsync().ForceFastSync,
		FSDiff:  gnomon.GetFastsync().ForceFastSyncDiff,
		DBtype:  gnomon.DBStorageType(),
		Para:    gnomon.GetParallel(),
		Assets:  menu.Assets.Enabled,
		Dapps:   menu.Control.Dapps,
	}
}

// // Make system tray with opts
// //   - Send Dero message menu
// //   - Explorer link
// //   - Manual reveal key for Holdero
// func systemTray(w fyne.App) bool {
// 	if desk, ok := w.(desktop.App); ok {
// 		m := fyne.NewMenu("MyApp",
// 			fyne.NewMenuItem("Send Message", func() {
// 				if !dReams.IsConfiguring() {
// 					menu.SendMessageMenu("", bundle.ResourceDReamsIconAltPng)
// 				}
// 			}),
// 			fyne.NewMenuItemSeparator(),
// 			fyne.NewMenuItem("Explorer", func() {
// 				link, _ := url.Parse("https://explorer.dero.io")
// 				fyne.CurrentApp().OpenURL(link)
// 			}),
// 			fyne.NewMenuItemSeparator(),
// 			fyne.NewMenuItem("Reveal Key", func() {
// 				go holdero.RevealKey(rpc.Wallet.ClientKey)
// 			}))
// 		desk.SetSystemTrayMenu(m)

// 		return true
// 	}
// 	return false
// }

// This is what we want to scan wallet for when Gnomon is synced
func gnomonScan(contracts map[string]string) {
	screen, bar := syncScreen()
	menu_tabs.Items[2].Content = screen
	menu.CheckWalletNames(rpc.Wallet.Address)
	screen.Objects[0].(*fyne.Container).Objects[1].(*canvas.Text).Text = "Syncing NFAs..."
	checkDreamsNFAs(contracts, bar)
	bar.SetValue(0)
	screen.Objects[0].(*fyne.Container).Objects[1].(*canvas.Text).Text = "Syncing G45s..."
	checkDreamsG45s(contracts, bar)
	if gnomon.DBStorageType() == "boltdb" {
		for _, r := range menu.Assets.Asset {
			gnomes.StoreBolt(r.Collection, r.Name, r)
		}
	}
	asset_tab.Objects[1].(*container.AppTabs).SelectIndex(1)
	menu_tabs.Items[2].Content = asset_tab
	menu_tabs.Refresh()
}

// Main dReams process loop
func fetch(done chan struct{}) {
	var offset int
	time.Sleep(3 * time.Second)
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ticker.C: // do on interval
			if !dReams.IsConfiguring() {
				rpc.Ping()
				if !rpc.Wallet.WS.IsRequesting() {
					rpc.Wallet.Sync()
				}

				checkConnection()
				gnomes.EndPoint()
				gnomes.State(dReams.IsConfiguring(), gnomonScan)

				go menuRefresh(offset)

				offset++
				if offset >= 41 {
					offset = 0
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
	if dReams.OnTab("Menu") && gnomon.IsInitialized() {
		switch gnomon.Status() {
		case "initializing":
			menu.Info.SetStatus("Gnomon Initializing")
		case "fastsyncing":
			menu.Info.SetStatus("Gnomon Fastsyncing...")
		case "closing":
			menu.Info.SetStatus("Gnomon Closing...")
		case "indexed":
			if !gnomon.HasIndex(uint64(menu.ReturnAssetCount())) && !gnomon.HasChecked() {
				menu.Info.SetStatus("Gnomon Syncing...")
			} else {
				menu.Info.SetStatus("Gnomon Synced")
			}
		case "indexing":
			menu.Info.SetStatus("Gnomon Syncing...")
		}

		if offset == 40 || menu.Info.Price.Text == "" {
			go menu.Info.RefreshPrice(App_Name)
		}
	}

	menu.Info.RefreshDaemon(App_Name)
	menu.Info.RefreshGnomon()
	menu.Info.RefreshWallet()
	menu.Info.RefreshIndexed()

	menu.Assets.Balances.Refresh()
}

// Check wallet for dReams NFAs
//   - Pass scids from db store, can be nil arg
func checkDreamsNFAs(scids map[string]string, progress *widget.ProgressBar) {
	if gnomon.IsReady() {
		menu.Info.SetStatus("Checking for Assets")
		if scids == nil {
			scids = gnomon.GetAllOwnersAndSCIDs()
		}

		logger.Println("[dReams] Checking NFA Assets")
		dreams.Theme.Select.Options = []string{}
		holdero.Settings.ClearAssets()

		progress.Max = float64(len(scids))

		for sc := range scids {
			if !rpc.Wallet.IsConnected() || !gnomon.IsRunning() {
				break
			}

			checkNFAOwner(sc)
			progress.SetValue(progress.Value + 1)
		}

		holdero.Settings.SortCardAssets()
		dreams.Theme.Sort()
		dreams.Theme.Select.Options = append(menu.Control.Themes, dreams.Theme.Select.Options...)
		dreams.Theme.Select.SetSelected(dreams.Theme.Name)
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
	if gnomon.IsRunning() {
		if header, _ := gnomon.GetSCIDValuesByKey(scid, "nameHdr"); header != nil {
			owner, _ := gnomon.GetSCIDValuesByKey(scid, "owner")
			file, _ := gnomon.GetSCIDValuesByKey(scid, "fileURL")
			collection, _ := gnomon.GetSCIDValuesByKey(scid, "collection")
			creator, _ := gnomon.GetSCIDValuesByKey(scid, "creatorAddr")
			icon, _ := gnomon.GetSCIDValuesByKey(scid, "iconURLHdr")
			if owner != nil && file != nil && collection != nil && creator != nil && icon != nil {
				if owner[0] == rpc.Wallet.Address && menu.ValidNFA(file[0]) {
					if !menu.IsDreamsNFACreator(creator[0]) {
						return
					}

					var add menu.Asset
					add.Name = header[0]
					add.Collection = collection[0]
					add.SCID = scid
					if typeHdr, _ := gnomon.GetSCIDValuesByKey(scid, "typeHdr"); typeHdr != nil {
						add.Type = menu.AssetType(collection[0], typeHdr[0])
					}

					check := strings.Trim(header[0], "0123456789")
					if check == "AZYDS" || check == "SIXART" {
						dreams.Theme.Add(header[0], owner[0])
						holdero.Settings.AddAvatar(header[0], owner[0])
						menu.Assets.Add(add, icon[0])
					} else if check == "AZYPCB" || check == "SIXPCB" {
						holdero.Settings.AddBacks(header[0], owner[0])
						menu.Assets.Add(add, icon[0])
					} else if check == "AZYPC" || check == "SIXPC" {
						holdero.Settings.AddFaces(header[0], owner[0])
						menu.Assets.Add(add, icon[0])
					} else if check == "DBC" {
						holdero.Settings.AddAvatar(header[0], owner[0])
						menu.Assets.Add(add, icon[0])
						if menu.DappEnabled("Duels") {
							duel.AddItemsToInventory(scid, header[0], owner[0], collection[0])
						}
					} else if collection[0] == "Dorblings NFA" {
						holdero.Settings.AddAvatar(header[0], owner[0])
						menu.Assets.Add(add, icon[0])
					} else if collection[0] == "DLAMPP" {
						// TODO review after mint
						holdero.Settings.AddAvatar(header[0], owner[0])
						menu.Assets.Add(add, icon[0])
					} else if collection[0] == "High Strangeness" {
						holdero.Settings.AddAvatar(header[0], owner[0])
						menu.Assets.Add(add, icon[0])
						hsCards(owner[0], header[0], check)
						if menu.DappEnabled("Duels") {
							duel.AddItemsToInventory(scid, header[0], owner[0], collection[0])
						}
					} else if collection[0] == "Dero Desperados" {
						holdero.Settings.AddAvatar(header[0], owner[0])
						menu.Assets.Add(add, icon[0])
						if menu.DappEnabled("Duels") {
							duel.AddItemsToInventory(scid, header[0], owner[0], collection[0])
						}
					} else if collection[0] == "Desperado Guns" {
						holdero.Settings.AddAvatar(header[0], owner[0])
						menu.Assets.Add(add, icon[0])
						if menu.DappEnabled("Duels") {
							duel.AddItemsToInventory(scid, header[0], owner[0], collection[0])
						}
					} else if collection[0] == "dSkullz" {
						holdero.Settings.AddAvatar(header[0], owner[0])
						menu.Assets.Add(add, icon[0])
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
func checkDreamsG45s(g45s map[string]string, progress *widget.ProgressBar) {
	if gnomon.IsReady() {
		if g45s == nil {
			g45s = gnomon.GetAllOwnersAndSCIDs()
		}
		logger.Println("[dReams] Checking G45 Assets")

		progress.Max = float64(len(g45s))

		for scid := range g45s {
			if !rpc.Wallet.IsConnected() || !gnomon.IsRunning() {
				break
			}

			if data, _ := gnomon.GetSCIDValuesByKey(scid, "metadata"); data != nil {
				owner, _ := gnomon.GetSCIDValuesByKey(scid, "owner")
				minter, _ := gnomon.GetSCIDValuesByKey(scid, "minter")
				coll, _ := gnomon.GetSCIDValuesByKey(scid, "collection")
				if owner != nil && minter != nil && coll != nil && owner[0] != "" {
					if owner[0] == rpc.Wallet.Address {
						var add menu.Asset
						add.Type = "Avatar"
						if minter[0] == menu.Seals_mint && coll[0] == menu.Seals_coll {
							var seal menu.Seal
							if err := json.Unmarshal([]byte(data[0]), &seal); err == nil {
								add.Name = seal.Name
								add.Collection = "Dero Seals"
								add.SCID = scid

								menu.Assets.Add(add, menu.ParseURL(seal.Image))
								holdero.Settings.AddAvatar(seal.Name, owner[0])

							}
						} else if minter[0] == menu.ATeam_mint && coll[0] == menu.ATeam_coll {
							var agent menu.Agent
							if err := json.Unmarshal([]byte(data[0]), &agent); err == nil {
								add.Name = agent.Name
								add.Collection = "Dero A-Team"
								add.SCID = scid

								menu.Assets.Add(add, menu.ParseURL(agent.Image))
								holdero.Settings.AddAvatar(agent.Name, owner[0])
							}
						} else if minter[0] == menu.Degen_mint && coll[0] == menu.Degen_coll {
							var degen menu.Degen
							if err := json.Unmarshal([]byte(data[0]), &degen); err == nil {
								add.Name = degen.Name
								add.Collection = "Dero Degens"
								add.SCID = scid

								menu.Assets.Add(add, menu.ParseURL(degen.Image))
								holdero.Settings.AddAvatar(degen.Name, owner[0])
							}
						}
					}
				}
			}

			progress.SetValue(progress.Value + 1)
		}
		holdero.Settings.SortAvatarAsset()
		menu.Assets.List.UnselectAll()
		menu.Assets.SortList()
	}
}

// Connection check for main process
func checkConnection() {
	if rpc.Daemon.IsConnected() {
		menu.Control.Check.Daemon.SetChecked(true)
	} else {
		menu.Control.Check.Daemon.SetChecked(false)
	}

	if rpc.Wallet.IsConnected() {
		if rpc.Daemon.IsConnected() {
			menu.Assets.Swap.Show()
		}

	} else {
		if !rpc.Wallet.WS.IsClosed() {
			if !rpc.Wallet.WS.IsConnecting() && !rpc.Wallet.WS.IsRequesting() {
				rpc.Wallet.CloseConnections("dReams")
			}
		} else if !rpc.Wallet.RPC.IsClosed() {
			rpc.Wallet.CloseConnections("dReams")
		}
		disconnected()
		gnomon.Checked(false)
	}
}

// Do when disconnected
func disconnected() {
	holdero.Disconnected(menu.DappEnabled("Holdero"))
	prediction.Disconnected()
	menu.Assets.Swap.Hide()
	menu.Assets.Names.ClearSelected()
	dreams.Theme.Select.Options = menu.Control.Themes
	dreams.Theme.Select.Refresh()
	menu.Assets.Asset = []menu.Asset{}
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
		grokked := rpc.GetSCCode(grok.GROKSCID)
		if grokked != "" {
			filter = append(filter, grokked)
		}

		grokked = rpc.GetSCCode(grok.GROKOG)
		if grokked != "" {
			filter = append(filter, grokked)
		}
	}

	if menu.DappEnabled("Dice") {
		dice := rpc.GetSCCode(dice.DICESCID)
		if dice != "" {
			filter = append(filter, dice)
		}
	}

	filter = append(filter, menu.ReturnEnabledNFAs(menu.Assets.Enabled)...)

	return
}

// Hidden object, controls Gnomon start and stop based on daemon connection
func daemonConnectedBox() fyne.Widget {
	menu.Control.Check.Daemon = widget.NewCheck("", func(b bool) {
		if !gnomon.IsInitialized() && !gnomon.IsStarting() {
			if rpc.DaemonVersion() == "3.5.3-139.DEROHE.STARGATE+04042023" {
				dialog.NewInformation("Daemon Version", "This daemon may conflict with Gnomon sync", dReams.Window).Show()
			}

			menu.Info.SetStatus("Starting Gnomon")
			rpc.GetFees()
			filters := gnomonFilters()
			gnomes.StartGnomon("dReams", gnomon.DBStorageType(), filters, menu.Assets.Count.G45+menu.Assets.Count.NFA, menu.Assets.Count.NFA, menu.G45Index)
		}

		if !b {
			menu.Info.SetStatus("Putting Gnomon to Sleep")
			gnomon.Stop("dReams")
			menu.Info.SetStatus("Gnomon is Sleeping")
		}
	})
	menu.Control.Check.Daemon.Disable()
	menu.Control.Check.Daemon.Hide()

	return menu.Control.Check.Daemon
}

// Daemon rpc entry object with default options
//   - Bound to rpc.Daemon.Rpc
func daemonRPCEntry() fyne.Widget {
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

	if menu.Control.Daemon != "" {
		options = append(options, menu.Control.Daemon)
	}
	entry := widget.NewSelectEntry(options)
	entry.PlaceHolder = "Daemon: "

	this := binding.BindString(&rpc.Daemon.Rpc)
	entry.Bind(this)

	return entry
}

// Connect/Disconnect objects for RPC
//   - Button OnConnect initializes RPC client and checks connection
//   - Button OnDisconnect closes RPC client if not syncing wallet
//   - Bound entries for Port and Auth
func rpcConnection() fyne.CanvasObject {
	port := "127.0.0.1:10103"
	entryPort := widget.NewSelectEntry([]string{"", port})
	entryAuth := widget.NewPasswordEntry()

	button := widget.NewButton("Connect", nil)
	button.OnTapped = func() {
		if button.Text == "Disconnect" {
			if rpc.Wallet.IsConnected() && gnomon.IsRunning() && !gnomon.HasChecked() {
				dialog.NewInformation("Gnomon Syncing", "Wait for Gnomon to sync before disconnecting", dReams.Window).Show()
				return
			}

			button.Importance = widget.MediumImportance
			entryAuth.Enable()
			entryPort.Enable()
			rpc.Wallet.Connected(false)
			rpc.Wallet.CloseConnections("dReams")
			disconnected()
			button.Text = "Connect"
			button.Refresh()
			connect_select.EnableIndex(1)
			connect_select.EnableIndex(2)
			return
		}

		go func() {
			rpc.Wallet.RPC.Init()
			rpc.GetAddress("dReams")
			checkConnection()
			if rpc.Wallet.IsConnected() {
				button.Importance = widget.HighImportance
				button.Text = "Disconnect"
				button.Refresh()
				entryAuth.Disable()
				entryPort.Disable()
				connect_select.DisableIndex(1)
				connect_select.DisableIndex(2)
			}
		}()
	}

	// OnChanged func for RPC entries
	onChanged := func(s string) {
		if rpc.Wallet.IsConnected() {
			rpc.Wallet.Connected(false)
			go checkConnection()
		}
	}

	// Wallet RPC entry object bound to rpc.Wallet.RPC.Port
	entryPort.PlaceHolder = "Wallet RPC: "
	entryPort.Bind(binding.BindString(&rpc.Wallet.RPC.Port))
	if entryPort.Text == "" {
		entryPort.SetText(port)
	}
	entryPort.OnChanged = onChanged

	// Authentication entry object bound to rpc.Wallet.RPC.Auth
	entryAuth.PlaceHolder = "user:pass"
	entryAuth.Bind(binding.BindString(&rpc.Wallet.RPC.Auth))
	entryAuth.OnChanged = onChanged

	return container.NewVBox(
		dwidget.NewSpacer(300, 0),
		entryPort,
		container.NewBorder(nil, nil, nil, container.NewStack(dwidget.NewSpacer(100, 0), button), entryAuth))
}

// Connect/Disconnect objects for XSWD
//   - Button OnConnect initializes WS and sends connection request
//   - Button OnDisconnect closes WS if not syncing wallet
//   - Bound entry for Port
func xswdConnection() fyne.CanvasObject {
	port := fmt.Sprintf("127.0.0.1:%d", xswd.XSWD_PORT)
	entryPort := widget.NewSelectEntry([]string{"", port})

	button := widget.NewButton("Connect", nil)
	button.OnTapped = func() {
		if button.Text == "Disconnect" {
			if rpc.Wallet.IsConnected() && gnomon.IsRunning() && !gnomon.HasChecked() {
				dialog.NewInformation("Gnomon Syncing", "Wait for Gnomon to sync before disconnecting", dReams.Window).Show()
				return
			}

			button.Importance = widget.MediumImportance
			rpc.Wallet.Connected(false)
			rpc.Wallet.CloseConnections("dReams")
			disconnected()
			entryPort.Enable()
			button.Text = "Connect"
			button.Refresh()
			connect_select.EnableIndex(0)
			connect_select.EnableIndex(2)
			return
		}

		go func() {
			button.Disable()
			entryPort.Disable()
			connect_select.DisableIndex(0)
			connect_select.DisableIndex(2)
			if rpc.Wallet.WS.Init(dReams.XSWD) {
				rpc.GetAddress("dReams")
				if rpc.Wallet.IsConnected() {
					checkConnection()
					button.Importance = widget.HighImportance
					button.Text = "Disconnect"
					button.Refresh()
					button.Enable()
					return
				}

				rpc.Wallet.CloseConnections("dReams")
			}

			entryPort.Enable()
			button.Importance = widget.MediumImportance
			button.Text = "Connect"
			button.Refresh()
			button.Enable()
			connect_select.EnableIndex(0)
			connect_select.EnableIndex(2)
		}()
	}

	// Wallet WS entry object bound to rpc.Wallet.WS.Port
	entryPort.PlaceHolder = "Wallet WS: "
	entryPort.Bind(binding.BindString(&rpc.Wallet.WS.Port))
	if entryPort.Text == "" {
		entryPort.SetText(port)
	}

	return container.NewVBox(
		dwidget.NewSpacer(300, 0),
		entryPort,
		container.NewHBox(layout.NewSpacer(), container.NewStack(dwidget.NewSpacer(100, 0), button)))
}

// Connect/Disconnect objects for walletapi
func accountConnection(d *dreams.AppObject) fyne.CanvasObject {
	_, names := dreams.GetAccounts()

	options := widget.NewSelectEntry(names)
	options.PlaceHolder = "DERO.db:"

	entryPass := widget.NewPasswordEntry()
	entryPass.PlaceHolder = "Password:"

	button := widget.NewButton("Sing in", nil)
	button.OnTapped = func() {
		go func() {
			button.Disable()
			defer func() {
				button.Enable()
			}()

			if button.Text == "Sign out" {
				if rpc.Wallet.IsConnected() && gnomon.IsRunning() && !gnomon.HasChecked() {
					dialog.NewInformation("Gnomon Syncing", "Wait for Gnomon to sync before singing out", dReams.Window).Show()
					return
				}

				rpc.Wallet.CloseConnections(App_Name)
				options.Enable()
				entryPass.Enable()
				connect_select.EnableIndex(0)
				connect_select.EnableIndex(1)
				button.Importance = widget.MediumImportance
				button.Text = "Sing in"
				button.Refresh()

				return
			} else {
				rpc.Ping()
				if !rpc.Daemon.IsConnected() {
					dialog.NewInformation("Select Daemon", "Connect to a daemon", dReams.Window).Show()
					return
				}

				network := "mainnet"
				if !globals.IsMainnet() {
					network = "testnet"
				}

				dir := filepath.Join(dreams.GetDir(), network) + string(filepath.Separator)
				path := filepath.Join(dir, options.Text)
				if strings.HasPrefix(options.Text, "/") {
					path = options.Text
				}

				if err := rpc.Wallet.OpenWalletFile(App_Name, path, entryPass.Text); err != nil {
					logger.Errorf("[%s] %s\n", App_Name, err)
					dialogError := dialog.NewInformation("Error", fmt.Sprintf("%s", err), dReams.Window)
					dialogError.Show()
					return
				}

				options.Disable()
				entryPass.Disable()
				connect_select.DisableIndex(0)
				connect_select.DisableIndex(1)
				button.Importance = widget.HighImportance
				button.Text = "Sign out"
				button.Refresh()
			}
		}()
	}

	return container.NewVBox(
		dwidget.NewSpacer(300, 0),
		options,
		container.NewBorder(nil, nil, nil, container.NewStack(dwidget.NewSpacer(100, 0), button), entryPass))
}

// Rescan func for owned assets list
func rescan() {
	logger.Printf("[%s] Rescaning Assets\n", App_Name)

	menu.Assets.Asset = []menu.Asset{}
	if menu.DappEnabled("Duels") {
		duel.Inventory.ClearAll()
	}
	gnomonScan(gnomon.IndexContains())
	menu.Assets.List.UnselectAll()
	menu.Assets.SortList()
}

func dappCloseCheck() {
	prediction.Service.IsStopped()
	rpc.Wallet.CloseConnections("dReams")
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
		case "Dice":
			versions["Dice"] = dice.Version().String()
		}
	}

	return versions
}

// Splash screen for assets syncing
func syncScreen() (max *fyne.Container, bar *widget.ProgressBar) {
	text := canvas.NewText("Syncing...", color.White)
	text.Alignment = fyne.TextAlignCenter
	text.TextSize = 21

	img := canvas.NewImageFromResource(bundle.ResourceMarketCirclePng)
	img.SetMinSize(fyne.NewSize(150, 150))

	bar = widget.NewProgressBar()
	bar.Max = 4
	bar.TextFormatter = func() string {
		return ""
	}

	max = container.NewBorder(
		dwidget.LabelColor(container.NewVBox(widget.NewLabel(""))),
		nil,
		nil,
		nil,
		container.NewCenter(img, text), bar)

	return
}
