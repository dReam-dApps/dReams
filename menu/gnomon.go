package menu

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/holdero"
	"github.com/SixofClubsss/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	xwidget "fyne.io/x/fyne/widget"
	"github.com/civilware/Gnomon/indexer"
	"github.com/civilware/Gnomon/storage"
	"github.com/civilware/Gnomon/structures"
)

const (
	NFA_SEARCH_FILTER = `Function init() Uint64
    10  IF EXISTS("owner") == 0 THEN GOTO 20 ELSE GOTO 999
    20  STORE("owner", SIGNER())
    30  STORE("creatorAddr", SIGNER())
    40  STORE("artificerAddr", ADDRESS_RAW("dero1qy0khp9s9yw2h0eu20xmy9lth3zp5cacmx3rwt6k45l568d2mmcf6qgcsevzx"))
    50  IF IS_ADDRESS_VALID(LOAD("artificerAddr")) == 1 THEN GOTO 60 ELSE GOTO 999
    60  STORE("active", 0)
    70  STORE("scBalance", 0)
    80  STORE("cancelBuffer", 300)
    90  STORE("startBlockTime", 0)
    100 STORE("endBlockTime", 0)
    110 STORE("bidCount", 0)
    120 STORE("staticBidIncr", 10000)
    130 STORE("percentBidIncr", 1000)
    140 STORE("listType", "")
    150 STORE("charityDonatePerc", 0)
    160 STORE("startPrice", 0)
    170 STORE("currBidPrice", 0)
    180 STORE("version", "1.1.1")
    500 IF LOAD("charityDonatePerc") + LOAD("artificerFee") + LOAD("royalty") > 100 THEN GOTO 999
    600 SEND_ASSET_TO_ADDRESS(SIGNER(), 1, SCID())
    610 RETURN 0
    999 RETURN 1
End Function`

	g45_search_filter = `STORE("type", "G45-NFT")`
)

type tableStats struct {
	Name    *canvas.Text
	Desc    *canvas.Text
	Version *canvas.Text
	Last    *canvas.Text
	Seats   *canvas.Text
	Open    *canvas.Text
	Image   canvas.Image
}

type gnomon struct {
	Para     int
	Trim     bool
	Fast     bool
	Start    bool
	Init     bool
	Sync     bool
	Syncing  bool
	Checked  bool
	Wait     bool
	Import   bool
	SCIDS    uint64
	Sync_ind *fyne.Animation
	Full_ind *fyne.Animation
	Icon_ind *xwidget.AnimatedGif
	Indexer  *indexer.Indexer
}

var Gnomes gnomon
var Stats tableStats
var Exit_signal bool

// Convert string to int64
func stringToInt64(s string) int64 {
	if s != "" {
		i, err := strconv.Atoi(s)
		if err != nil {
			log.Println("[stringToInt64]", err)
			return 0
		}
		return int64(i)
	}

	return 0
}

// Menu label when Gnomon starting
func startLabel() {
	Assets.Gnomes_sync.Text = (" Starting Gnomon")
	Assets.Gnomes_sync.Refresh()
}

// Menu label when Gomon scans wallet
func checkLabel() {
	Assets.Gnomes_sync.Text = (" Checking for Assets")
	Assets.Gnomes_sync.Refresh()
}

// Menu label when Gnomon is closing
func StopLabel() {
	if Assets.Gnomes_sync != nil {
		Assets.Gnomes_sync.Text = (" Putting Gnomon to Sleep")
		Assets.Gnomes_sync.Refresh()
	}
}

// Menu label when Gnomon is not running
func SleepLabel() {
	Assets.Gnomes_sync.Text = (" Gnomon is Sleeping")
	Assets.Gnomes_sync.Refresh()
}

// dReams app status indicators for wallet, daemon and Gnomon
func StartIndicators() fyne.CanvasObject {
	purple := color.RGBA{105, 90, 205, 210}
	blue := color.RGBA{31, 150, 200, 210}
	alpha := color.RGBA{0, 0, 0, 0}

	g_top := canvas.NewRectangle(color.Black)
	g_top.SetMinSize(fyne.NewSize(57, 10))

	g_bottom := canvas.NewRectangle(color.Black)
	g_bottom.SetMinSize(fyne.NewSize(57, 10))

	Gnomes.Sync_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if Gnomes.Init && !Gnomes.Checked {
				g_top.FillColor = c
				canvas.Refresh(g_top)
				g_bottom.FillColor = c
				canvas.Refresh(g_bottom)
			} else {
				g_top.FillColor = alpha
				canvas.Refresh(g_top)
				g_bottom.FillColor = alpha
				canvas.Refresh(g_bottom)
			}
		})

	Gnomes.Sync_ind.RepeatCount = fyne.AnimationRepeatForever
	Gnomes.Sync_ind.AutoReverse = true

	sync_box := container.NewVBox(
		g_top,
		layout.NewSpacer(),
		g_bottom)

	g_full := canvas.NewRectangle(color.Black)
	g_full.SetMinSize(fyne.NewSize(57, 36))

	Gnomes.Full_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if Gnomes.Init && FastSynced() && Gnomes.Checked {
				g_full.FillColor = c
				canvas.Refresh(g_full)
				sync_box.Hide()
			} else {
				g_full.FillColor = alpha
				canvas.Refresh(g_full)
				sync_box.Show()
			}
		})

	Gnomes.Full_ind.RepeatCount = fyne.AnimationRepeatForever
	Gnomes.Full_ind.AutoReverse = true

	Gnomes.Icon_ind, _ = xwidget.NewAnimatedGifFromResource(bundle.ResourceGnomonGifGif)
	Gnomes.Icon_ind.SetMinSize(fyne.NewSize(36, 36))

	d_rect := canvas.NewRectangle(color.Black)
	d_rect.SetMinSize(fyne.NewSize(36, 36))

	Control.Daemon_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if rpc.Daemon.Connect {
				d_rect.FillColor = c
				canvas.Refresh(d_rect)
			} else {
				d_rect.FillColor = alpha
				canvas.Refresh(d_rect)
			}
		})

	Control.Daemon_ind.RepeatCount = fyne.AnimationRepeatForever
	Control.Daemon_ind.AutoReverse = true

	w_rect := canvas.NewRectangle(color.Black)
	w_rect.SetMinSize(fyne.NewSize(36, 36))

	Control.Wallet_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if rpc.Wallet.Connect {
				w_rect.FillColor = c
				canvas.Refresh(w_rect)
			} else {
				w_rect.FillColor = alpha
				canvas.Refresh(w_rect)
			}
		})

	Control.Wallet_ind.RepeatCount = fyne.AnimationRepeatForever
	Control.Wallet_ind.AutoReverse = true

	d := canvas.NewText("D", bundle.TextColor)
	d.TextStyle.Bold = true
	d.TextSize = 16
	w := canvas.NewText("W", bundle.TextColor)
	w.TextStyle.Bold = true
	w.TextSize = 16

	connect_box := container.NewHBox(
		container.NewMax(d_rect, container.NewCenter(d)),
		container.NewMax(w_rect, container.NewCenter(w)))

	pbot := canvas.NewImageFromResource(bundle.ResourcePokerBotIconPng)
	pbot.SetMinSize(fyne.NewSize(30, 30))
	p_rect := canvas.NewRectangle(alpha)
	p_rect.SetMinSize(fyne.NewSize(36, 36))

	dService := canvas.NewImageFromResource(bundle.ResourceDReamServiceIconPng)
	dService.SetMinSize(fyne.NewSize(30, 30))
	s_rect := canvas.NewRectangle(alpha)
	s_rect.SetMinSize(fyne.NewSize(36, 36))

	service_box := container.NewHBox(
		container.NewMax(p_rect, container.NewCenter(pbot)),
		container.NewMax(s_rect, container.NewCenter(dService)))

	Control.Poker_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if rpc.Odds.Run {
				p_rect.FillColor = c
				pbot.Show()
				canvas.Refresh(p_rect)
			} else {
				p_rect.FillColor = alpha
				pbot.Hide()
				canvas.Refresh(p_rect)
			}
		})

	Control.Service_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if rpc.Wallet.Service {
				s_rect.FillColor = c
				dService.Show()
				canvas.Refresh(s_rect)
			} else {
				s_rect.FillColor = alpha
				dService.Hide()
				canvas.Refresh(s_rect)
			}
		})

	Control.Poker_ind.RepeatCount = fyne.AnimationRepeatForever
	Control.Poker_ind.AutoReverse = true

	Control.Service_ind.RepeatCount = fyne.AnimationRepeatForever
	Control.Service_ind.AutoReverse = true

	top_box := container.NewHBox(layout.NewSpacer(), service_box, connect_box, container.NewMax(g_full, sync_box, Gnomes.Icon_ind))
	place := container.NewVBox(top_box, layout.NewSpacer())

	go func() {
		Gnomes.Sync_ind.Start()
		Gnomes.Full_ind.Start()
		Gnomes.Icon_ind.Start()
		Control.Daemon_ind.Start()
		Control.Wallet_ind.Start()
		Control.Poker_ind.Start()
		Control.Service_ind.Start()
	}()

	return container.NewMax(place)
}

// Stop dReams app status indicators
func StopIndicators() {
	Gnomes.Icon_ind.Stop()
	Gnomes.Sync_ind.Stop()
	Gnomes.Full_ind.Stop()
	Control.Daemon_ind.Stop()
	Control.Wallet_ind.Stop()
	Control.Poker_ind.Stop()
	Control.Service_ind.Stop()
}

// dReams search filters for Gnomon index
func searchFilters() (filter []string) {
	if Control.Dapp_list["Holdero"] {
		holdero110 := rpc.GetHoldero110Code(0)
		if holdero110 != "" {
			filter = append(filter, holdero110)
		}

		holdero100 := rpc.GetHoldero100Code()
		if holdero100 != "" {
			filter = append(filter, holdero100)
		}

		holderoHGC := rpc.GetHoldero110Code(2)
		if holderoHGC != "" {
			filter = append(filter, holderoHGC)
		}
	}

	if Control.Dapp_list["Baccarat"] {
		bacc := rpc.GetBaccCode()
		if bacc != "" {
			filter = append(filter, bacc)
		}
	}

	if Control.Dapp_list["dSports and dPredictions"] {
		predict := rpc.GetPredictCode(0)
		if predict != "" {
			filter = append(filter, predict)
		}

		sports := rpc.GetSportsCode(0)
		if sports != "" {
			filter = append(filter, sports)
		}
	}

	gnomon := rpc.GetGnomonCode()
	if gnomon != "" {
		filter = append(filter, gnomon)
	}

	names := rpc.GetNameServiceCode()
	if names != "" {
		filter = append(filter, names)
	}

	ratings := rpc.GetSCCode(rpc.RatingSCID)
	if ratings != "" {
		filter = append(filter, ratings)
	}

	if Control.Dapp_list["DerBnb"] {
		bnb := rpc.GetSCCode(rpc.DerBnbSCID)
		if bnb != "" {
			filter = append(filter, bnb)
		}
	}

	filter = append(filter, NFA_SEARCH_FILTER)
	if !Gnomes.Trim {
		filter = append(filter, g45_search_filter)
	}

	return filter
}

// Manually add SCID to Gnomon index
func manualIndex(scid []string) {
	filters := Gnomes.Indexer.SearchFilter
	Gnomes.Indexer.SearchFilter = []string{}
	scidstoadd := make(map[string]*structures.FastSyncImport)

	for i := range scid {
		owner := rpc.CheckForIndex(scid[i])
		if owner != nil {
			scidstoadd[scid[i]] = &structures.FastSyncImport{}
			scidstoadd[scid[i]].Owner = owner.(string)
		}
	}

	err := Gnomes.Indexer.AddSCIDToIndex(scidstoadd)
	if err != nil {
		log.Printf("Err - %v", err)
	}
	Gnomes.Indexer.SearchFilter = filters
}

// Create Gnomon graviton db with dReams tag
func GnomonDB() *storage.GravitonStore {
	shasum := fmt.Sprintf("%x", sha1.Sum([]byte("dReams")))
	db_folder := fmt.Sprintf("gnomondb\\%s_%s", "dReams", shasum)
	db, _ := storage.NewGravDB(db_folder, "25ms")

	return db
}

// Start Gnomon indexer with or without search filters
//   - End point from rpc.Daemon.Rpc
//   - tag for log print
//   - Passing nil filters with Gnomes.Trim false will run a full Gnomon index
//   - custom func() is for adding specific SCID to index on Gnomon start, Gnomes.Trim false will bypass
//   - lower defines the lower limit of indexed SCIDs from Gnomon search filters before custom adds
//   - upper defines the higher limit when custom indexed SCIDs exist already
func StartGnomon(tag string, filters []string, upper, lower int, custom func()) {
	Gnomes.Start = true
	log.Printf("[%s] Starting Gnomon\n", tag)
	backend := GnomonDB()

	last_height := backend.GetLastIndexHeight()
	runmode := "daemon"
	mbl := false
	closeondisconnect := false

	if filters != nil || !Gnomes.Trim {
		Gnomes.Indexer = indexer.NewIndexer(backend, filters, last_height, rpc.Daemon.Rpc, runmode, mbl, closeondisconnect, Gnomes.Fast)
		go Gnomes.Indexer.StartDaemonMode(Gnomes.Para)
		time.Sleep(3 * time.Second)
		Gnomes.Init = true

		if Gnomes.Trim {
			i := 0
			for {
				contracts := len(Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs())
				if contracts >= upper {
					Gnomes.Trim = false
					break
				}

				if contracts >= lower {
					go custom()
					break
				}
				time.Sleep(1 * time.Second)
				i++
				if i == 60 {
					Gnomes.Trim = false
					log.Printf("[%s] Could not add all custom SCID for index\n", tag)
					break
				}
			}
		}
	}

	Gnomes.Start = false
}

// Manually add G45 collection to Gnomon index
func g45Index() {
	log.Println("[dReams] Adding G45 Collections")
	filters := Gnomes.Indexer.SearchFilter
	Gnomes.Indexer.SearchFilter = []string{}
	scidstoadd := make(map[string]*structures.FastSyncImport)

	a := rpc.GetG45Collection(holdero.ATeam_coll)
	for i := range a {
		scidstoadd[a[i]] = &structures.FastSyncImport{}
	}

	s := rpc.GetG45Collection(holdero.Seals_coll)
	for i := range s {
		scidstoadd[s[i]] = &structures.FastSyncImport{}
	}

	err := Gnomes.Indexer.AddSCIDToIndex(scidstoadd)
	if err != nil {
		log.Printf("Err - %v", err)
	}
	Gnomes.Indexer.SearchFilter = filters
	Gnomes.Trim = false
}

// Update Gnomon endpoint to current rpc.Daemon.Rpc value
func GnomonEndPoint() {
	if rpc.Daemon.Connect && Gnomes.Init && !Gnomes.Syncing {
		Gnomes.Indexer.Endpoint = rpc.Daemon.Rpc
	}
}

// Shut down Gnomes.Indexer
//   - tag for log print
func StopGnomon(tag string) {
	if Gnomes.Init && !GnomonClosing() {
		log.Printf("[%s] Putting Gnomon to Sleep\n", tag)
		Gnomes.Indexer.Close()
		Gnomes.Init = false
		time.Sleep(1 * time.Second)
		log.Printf("[%s] Gnomon is Sleeping\n", tag)
	}
}

// Check if Gnomon is writing
func GnomonWriting() bool {
	return Gnomes.Indexer.Backend.Writing == 1
}

// Check if Gnomon is closing
func GnomonClosing() bool {
	if !Gnomes.Init {
		return false
	}

	if Gnomes.Indexer.Closing || Gnomes.Indexer.Backend.Closing {
		return true
	}

	return false
}

// Check if Gnomon index contains SCIDs
func FastSynced() bool {
	return Gnomes.SCIDS > 0
}

// Check three connection signals
func Connected() bool {
	if rpc.Daemon.Connect && rpc.Wallet.Connect && Gnomes.Sync {
		return true
	}

	return false
}

// Gnomon will scan connected wallet on start up, then ensure sync
//   - Hold out checking if dReams is in configure
//   - windows disables certain initial sync routines from running on windows os
func GnomonState(windows, config bool) {
	if rpc.Daemon.Connect && Gnomes.Init && !GnomonClosing() {
		contracts := Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
		Gnomes.SCIDS = uint64(len(contracts))
		if FastSynced() && !Gnomes.Trim {
			height := int64(rpc.Wallet.Height)
			if Gnomes.Indexer.ChainHeight >= height-1 && height != 0 && !GnomonClosing() {
				Gnomes.Sync = true
				if !config && rpc.Wallet.Connect && !Gnomes.Checked {
					Gnomes.Syncing = true
					if Control.Dapp_list["dSports and dPredictions"] {
						go CheckBetContractOwners(contracts)
						if !windows {
							go PopulateSports(contracts)
							go PopulatePredictions(contracts)
						}
					}

					if Control.Dapp_list["Holdero"] {
						CreateTableList(Gnomes.Checked, contracts)
						CheckWalletNames(rpc.Wallet.Address)
					}

					go CheckDreamsG45s(Gnomes.Checked, contracts)
					go CheckDreamsNFAs(Gnomes.Checked, contracts)

					if !windows {
						FindNfaListings(contracts)
					}
					Gnomes.Checked = true
					Gnomes.Syncing = false
				}
			} else {
				Gnomes.Sync = false
			}
		}

		if Control.Dapp_list["Holdero"] {
			Assets.Stats_box = *container.NewVBox(Assets.Collection, Assets.Name, IconImg(bundle.ResourceAvatarFramePng))
			Assets.Stats_box.Refresh()
			Poker.Stats_box = *container.NewVBox(Stats.Name, Stats.Desc, Stats.Version, Stats.Last, Stats.Seats, TableIcon(bundle.ResourceAvatarFramePng))
			Poker.Stats_box.Refresh()
		}

		// Update live market info
		if len(Market.Viewing) == 64 && rpc.Wallet.Connect {
			if Market.Tab == "Buy" {
				GetBuyNowDetails(Market.Viewing)
				go RefreshNfaImages()
			} else {
				GetAuctionDetails(Market.Viewing)
				go RefreshNfaImages()
			}
		}
	}
}

// Search Gnomon db for indexed SCID
func searchIndex(scid string) {
	if len(scid) == 64 {
		var found bool
		all := Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
		for sc := range all {
			if scid == sc {
				log.Println("[dReams] " + scid + " Indxed")
				found = true
			}
		}
		if !found {
			log.Println("[dReams] " + scid + " Not Found")
		}
	}
}

// Check wallet for dReams NFAs
//   - Pass scids from db store, can be nil arg
//   - Pass false gc for rechecks
func CheckDreamsNFAs(gc bool, scids map[string]string) {
	if Gnomes.Sync && !gc && !GnomonClosing() {
		go checkLabel()
		if scids == nil {
			scids = Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
		}
		keys := make([]string, len(scids))
		log.Println("[dReams] Checking NFA Assets")
		holdero.Settings.FaceSelect.Options = []string{}
		holdero.Settings.BackSelect.Options = []string{}
		holdero.Settings.ThemeSelect.Options = []string{}
		holdero.Settings.AvatarSelect.Options = []string{}

		i := 0
		for k := range scids {
			if !rpc.Wallet.Connect || GnomonClosing() {
				break
			}
			keys[i] = k
			checkNFAOwner(keys[i])
			i++
		}
		sort.Strings(holdero.Settings.FaceSelect.Options)
		sort.Strings(holdero.Settings.BackSelect.Options)
		sort.Strings(holdero.Settings.ThemeSelect.Options)

		ld := []string{"Light", "Dark"}
		holdero.Settings.FaceSelect.Options = append(ld, holdero.Settings.FaceSelect.Options...)
		holdero.Settings.BackSelect.Options = append(ld, holdero.Settings.BackSelect.Options...)
		holdero.Settings.ThemeSelect.Options = append([]string{"Main"}, holdero.Settings.ThemeSelect.Options...)

		sort.Strings(Assets.Assets)
		Assets.Asset_list.Refresh()
		if Control.Dapp_list["Holdero"] {
			holdero.DisableHolderoTools()
		}
	}
}

// Check wallet for all indexed NFAs
//   - Pass scids from db store, can be nil arg
//   - Pass false gc for rechecks
func CheckAllNFAs(gc bool, scids map[string]string) {
	if Gnomes.Sync && !gc && !GnomonClosing() {
		if scids == nil {
			scids = Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
		}
		keys := make([]string, len(scids))

		i := 0
		assets := []string{}
		for k := range scids {
			if !rpc.Wallet.Connect || GnomonClosing() {
				break
			}

			keys[i] = k
			owner, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(keys[i], "owner", Gnomes.Indexer.ChainHeight, true)
			header, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(keys[i], "nameHdr", Gnomes.Indexer.ChainHeight, true)
			file, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(keys[i], "fileURL", Gnomes.Indexer.ChainHeight, true)
			if owner != nil && header != nil && file != nil {
				if owner[0] == rpc.Wallet.Address && ValidNfa(file[0]) {
					assets = append(assets, header[0]+"   "+keys[i])
				}
			}
			i++
		}

		sort.Strings(assets)
		Assets.Assets = assets
		Assets.Asset_list.Refresh()
	}
}

// Scan all bet contracts to verify if owner
//   - Pass contracts from db store, can be nil arg
func CheckBetContractOwners(contracts map[string]string) {
	if Gnomes.Sync && !Gnomes.Checked && !GnomonClosing() {
		if contracts == nil {
			contracts = Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
		}
		keys := make([]string, len(contracts))

		i := 0
		for k := range contracts {
			keys[i] = k
			verifyBetContractOwner(keys[i], "p")
			verifyBetContractOwner(keys[i], "s")
			if rpc.Wallet.BetOwner {
				break
			}
			i++
		}
	}
}

// Get Gnomon headers of SCID
func GetSCHeaders(scid string) []string {
	if !GnomonClosing() {
		headers, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.GnomonSCID, scid, Gnomes.Indexer.ChainHeight, true)

		if headers != nil {
			split := strings.Split(headers[0], ";")

			if split[0] == "" {
				return nil
			}

			return split
		}
	}
	return nil
}

// Verify if wallet is owner on bet contract
//   - Passed t defines sports or prediction contract
func verifyBetContractOwner(scid, t string) {
	if !GnomonClosing() {
		owner, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "owner", Gnomes.Indexer.ChainHeight, true)
		dev, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "dev", Gnomes.Indexer.ChainHeight, true)
		_, init := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, t+"_init", Gnomes.Indexer.ChainHeight, true)

		if owner != nil && dev != nil && init != nil {
			if dev[0] == rpc.DevAddress && !rpc.Wallet.BetOwner {
				SetBetOwner(owner[0])
			}
		}
	}
}

// Verify if wallet is a co owner on bet contract
func VerifyBetSigner(scid string) bool {
	if Gnomes.Init && Gnomes.Sync && !GnomonClosing() {
		for i := 2; i < 10; i++ {
			if GnomonClosing() {
				break
			}

			signer_addr, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "co_signer"+strconv.Itoa(i), Gnomes.Indexer.ChainHeight, true)
			if signer_addr != nil {
				if signer_addr[0] == rpc.Wallet.Address {
					return true
				}
			}
		}
	}

	return false
}

// Get info for bet contract by SCID
//   - Passed t defines sports or prediction contract
//   - Adding constructed header string to list, owned []string
func checkBetContract(scid, t string, list, owned []string) ([]string, []string) {
	if Gnomes.Init && !GnomonClosing() {
		owner, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "owner", Gnomes.Indexer.ChainHeight, true)
		dev, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "dev", Gnomes.Indexer.ChainHeight, true)
		_, init := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, t+"_init", Gnomes.Indexer.ChainHeight, true)

		if owner != nil && dev != nil && init != nil {
			if dev[0] == rpc.DevAddress {
				headers := GetSCHeaders(scid)
				name := "?"
				desc := "?"
				var hidden bool
				_, restrict := Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.RatingSCID, "restrict", Gnomes.Indexer.ChainHeight, true)
				_, rating := Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.RatingSCID, scid, Gnomes.Indexer.ChainHeight, true)

				if restrict != nil && rating != nil {
					Control.Contract_rating[scid] = rating[0]
					if rating[0] <= restrict[0] {
						hidden = true
					}
				}

				if headers != nil {
					if headers[1] != "" {
						desc = headers[1]
					}

					if headers[0] != "" {
						name = " " + headers[0]
					}

					if headers[0] == "-" {
						hidden = true
					}
				}

				var co_signer bool
				if VerifyBetSigner(scid) {
					co_signer = true
					if !Gnomes.Import {
						Control.Bet_menu_p.Show()
						Control.Bet_menu_s.Show()
					}
				}

				if owner[0] == rpc.Wallet.Address || co_signer {
					owned = append(owned, name+"   "+desc+"   "+scid)
				}

				if !hidden {
					list = append(list, name+"   "+desc+"   "+scid)
				}
			}
		}
	}
	return list, owned
}

// Populate all dReams dPrediction contracts
//   - Pass contracts from db store, can be nil arg
func PopulatePredictions(contracts map[string]string) {
	if rpc.Daemon.Connect && Gnomes.Sync && !GnomonClosing() {
		list := []string{}
		owned := []string{}
		if contracts == nil {
			contracts = Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
		}
		keys := make([]string, len(contracts))

		i := 0
		for k := range contracts {
			keys[i] = k
			list, owned = checkBetContract(keys[i], "p", list, owned)
			i++
		}
		t := len(list)
		list = append(list, " Contracts: "+strconv.Itoa(t))
		sort.Strings(list)
		Control.Predict_contracts = list

		sort.Strings(owned)
		Control.Predict_owned = owned

	}
}

// Populate all dReams dSports contracts
//   - Pass contracts from db store, can be nil arg
func PopulateSports(contracts map[string]string) {
	if rpc.Daemon.Connect && Gnomes.Sync && !GnomonClosing() {
		list := []string{}
		owned := []string{}
		if contracts == nil {
			contracts = Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
		}
		keys := make([]string, len(contracts))

		i := 0
		for k := range contracts {
			keys[i] = k
			list, owned = checkBetContract(keys[i], "s", list, owned)
			i++
		}

		t := len(list)
		list = append(list, " Contracts: "+strconv.Itoa(t))
		sort.Strings(list)
		Control.Sports_contracts = list

		sort.Strings(owned)
		Control.Sports_owned = owned
	}
}

// Check if SCID is a NFA
func isNfa(scid string) bool {
	if Gnomes.Init && Gnomes.Sync && !GnomonClosing() {
		artAddr, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "artificerAddr", Gnomes.Indexer.ChainHeight, true)
		if artAddr != nil {
			return artAddr[0] == rpc.ArtAddress
		}
	}
	return false
}

// Check if SCID is a valid NFA
//   - file != "-"
func ValidNfa(file string) bool {
	return file != "-"
}

// If wallet owns dReams NFA, populate for use in dReams
//   - See games container in menu.PlaceAssets()
func checkNFAOwner(scid string) {
	if !GnomonClosing() {
		owner, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "owner", Gnomes.Indexer.ChainHeight, true)
		header, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.ChainHeight, true)
		file, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "fileURL", Gnomes.Indexer.ChainHeight, true)
		if owner != nil && header != nil && file != nil {
			if owner[0] == rpc.Wallet.Address && ValidNfa(file[0]) {
				check := strings.Trim(header[0], "0123456789")
				if check == "AZYDS" || check == "SIXART" {
					themes := holdero.Settings.ThemeSelect.Options
					new_themes := append(themes, header[0])
					holdero.Settings.ThemeSelect.Options = new_themes
					holdero.Settings.ThemeSelect.Refresh()

					avatars := holdero.Settings.AvatarSelect.Options
					new_avatar := append(avatars, header[0])
					holdero.Settings.AvatarSelect.Options = new_avatar
					holdero.Settings.AvatarSelect.Refresh()
					Assets.Assets = append(Assets.Assets, header[0]+"   "+scid)
				} else if check == "AZYPCB" || check == "SIXPCB" {
					current := holdero.Settings.BackSelect.Options
					new := append(current, header[0])
					holdero.Settings.BackSelect.Options = new
					holdero.Settings.BackSelect.Refresh()
					Assets.Assets = append(Assets.Assets, header[0]+"   "+scid)
				} else if check == "AZYPC" || check == "SIXPC" {
					current := holdero.Settings.FaceSelect.Options
					new := append(current, header[0])
					holdero.Settings.FaceSelect.Options = new
					holdero.Settings.FaceSelect.Refresh()
					Assets.Assets = append(Assets.Assets, header[0]+"   "+scid)
				} else if check == "DBC" {
					current := holdero.Settings.AvatarSelect.Options
					new := append(current, header[0])
					holdero.Settings.AvatarSelect.Options = new
					holdero.Settings.AvatarSelect.Refresh()
					Assets.Assets = append(Assets.Assets, header[0]+"   "+scid)
				} else if check == "HighStrangeness" {
					current_av := holdero.Settings.AvatarSelect.Options
					new_av := append(current_av, header[0])
					holdero.Settings.AvatarSelect.Options = new_av
					holdero.Settings.AvatarSelect.Refresh()
					Assets.Assets = append(Assets.Assets, header[0]+"   "+scid)

					var have_cards bool
					for _, face := range holdero.Settings.FaceSelect.Options {
						if face == "High-Strangeness" {
							have_cards = true
						}
					}

					if !have_cards {
						current_d := holdero.Settings.FaceSelect.Options
						new_d := append(current_d, "High-Strangeness")
						holdero.Settings.FaceSelect.Options = new_d
						holdero.Settings.FaceSelect.Refresh()

						current_b := holdero.Settings.BackSelect.Options
						new_b := append(current_b, "High-Strangeness")
						holdero.Settings.BackSelect.Options = new_b
						holdero.Settings.BackSelect.Refresh()
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
						themes := holdero.Settings.ThemeSelect.Options
						for _, th := range themes {
							if th == "HSTheme"+strconv.Itoa(i) {
								have_theme = true
							}
						}

						if !have_theme {
							new_themes := append(themes, "HSTheme"+strconv.Itoa(i))
							holdero.Settings.ThemeSelect.Options = new_themes
							holdero.Settings.ThemeSelect.Refresh()
						}
					}
				}
			}
		}
	}
}

// Get SCID info and update Asset content
func GetOwnedAssetStats(scid string) {
	if Gnomes.Init && !GnomonClosing() {
		n, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.LastIndexedHeight, true)
		if n != nil {
			c, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "collection", Gnomes.Indexer.LastIndexedHeight, true)
			//d, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "descrHdr:", Gnomes.Indexer.LastIndexedHeight, true)
			i, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "iconURLHdr", Gnomes.Indexer.LastIndexedHeight, true)

			if n != nil {
				Assets.Name.Text = (" Name: " + n[0])
				Assets.Name.Refresh()
				if !Control.list_open && !Control.send_open {
					Control.List_button.Show()
					Control.Send_asset.Show()
				}

			} else {
				Assets.Name.Text = (" Name: ?")
				Assets.Name.Refresh()
			}

			var a []string
			if c != nil {
				Assets.Collection.Text = (" Collection: " + c[0])
				Assets.Collection.Refresh()
				if c[0] == "High Strangeness" {
					a, _ = Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "fileURL", Gnomes.Indexer.ChainHeight, true)
				}
			} else {
				Assets.Collection.Text = (" Collection: ?")
				Assets.Collection.Refresh()
			}

			if i != nil {
				if a != nil {
					Assets.Icon, _ = holdero.DownloadFile(a[0], n[0])
				} else {
					Assets.Icon, _ = holdero.DownloadFile(i[0], n[0])
				}
			} else {
				Assets.Icon = *canvas.NewImageFromImage(nil)
			}

		} else {
			Control.List_button.Hide()
			data, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "metadata", Gnomes.Indexer.LastIndexedHeight, true)
			minter, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "minter", Gnomes.Indexer.LastIndexedHeight, true)
			coll, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "collection", Gnomes.Indexer.LastIndexedHeight, true)
			if data != nil && minter != nil && coll != nil {
				if minter[0] == holdero.Seals_mint && coll[0] == holdero.Seals_coll {
					var seal holdero.Seal
					if err := json.Unmarshal([]byte(data[0]), &seal); err == nil {
						check := strings.Trim(seal.Name, " #0123456789")
						if check == "Dero Seals" {
							Assets.Name.Text = (" Name: " + seal.Name)
							Assets.Name.Refresh()

							Assets.Collection.Text = (" Collection: " + check)
							Assets.Collection.Refresh()

							number := strings.Trim(seal.Name, "DeroSals# ")
							Assets.Icon, _ = holdero.DownloadFile("https://ipfs.io/ipfs/QmP3HnzWpiaBA6ZE8c3dy5ExeG7hnYjSqkNfVbeVW5iEp6/low/"+number+".jpg", seal.Name)
						}
					}
				} else if minter[0] == holdero.ATeam_mint && coll[0] == holdero.ATeam_coll {
					var agent holdero.Agent
					if err := json.Unmarshal([]byte(data[0]), &agent); err == nil {
						Assets.Name.Text = (" Name: " + agent.Name)
						Assets.Name.Refresh()

						Assets.Collection.Text = (" Collection: Dero A-Team")
						Assets.Collection.Refresh()

						number := strconv.Itoa(agent.ID)
						if agent.ID < 172 {
							Assets.Icon, _ = holdero.DownloadFile("https://ipfs.io/ipfs/QmaRHXcQwbFdUAvwbjgpDtr5kwGiNpkCM2eDBzAbvhD7wh/low/"+number+".jpg", agent.Name)
						} else {
							Assets.Icon, _ = holdero.DownloadFile("https://ipfs.io/ipfs/QmQQyKoE9qDnzybeDCXhyMhwQcPmLaVy3AyYAzzC2zMauW/low/"+number+".jpg", agent.Name)
						}
					}
				}
			}
		}
	}
}

// Get a requested NFA url
//   - w of 0 returns "fileURL"
//   - w of 1 returns "iconURLHdr"
//   - w of 2 returns "coverURLHdr"
func GetAssetUrl(w int, scid string) (url string) {
	var link []string
	switch w {
	case 0:
		link, _ = Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "fileURL", Gnomes.Indexer.ChainHeight, true)
	case 1:
		link, _ = Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "iconURLHdr", Gnomes.Indexer.LastIndexedHeight, true)
	case 2:
		link, _ = Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "coverURLHdr", Gnomes.Indexer.LastIndexedHeight, true)
	default:
		// nothing
	}

	if link != nil {
		url = link[0]
	}

	return
}

// Check if wallet owns Holdero table
func CheckTableOwner(scid string) bool {
	if len(scid) != 64 || !Gnomes.Init || GnomonClosing() {
		return false
	}

	check := strings.Trim(scid, " 0123456789")
	if check == "Holdero Tables:" {
		return false
	}

	owner, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "owner:", Gnomes.Indexer.LastIndexedHeight, true)
	if owner != nil {
		return owner[0] == rpc.Wallet.Address
	}

	return false
}

// Check if Holdero table is a tournament table
func CheckHolderoContract(scid string) bool {
	if len(scid) != 64 || !Gnomes.Init || GnomonClosing() {
		return false
	}

	_, deck := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "Deck Count:", Gnomes.Indexer.LastIndexedHeight, true)
	_, version := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "V:", Gnomes.Indexer.LastIndexedHeight, true)
	_, tourney := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "Tournament", Gnomes.Indexer.LastIndexedHeight, true)
	if deck != nil && version != nil && version[0] >= 100 {
		rpc.Signal.Contract = true
	}

	if tourney != nil && tourney[0] == 1 {
		return true
	}

	return false
}

// Check owner of any SCID using "owner" key
func CheckOwner(scid string) bool {
	if len(scid) != 64 || !Gnomes.Init || GnomonClosing() {
		return false
	}

	owner, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "owner", Gnomes.Indexer.LastIndexedHeight, true)
	if owner != nil {
		return owner[0] == rpc.Wallet.Address
	}

	return false
}

// Check Holdero table version
func checkTableVersion(scid string) uint64 {
	_, v := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "V:", Gnomes.Indexer.LastIndexedHeight, true)

	if v != nil && v[0] >= 100 {
		return v[0]
	}
	return 0
}

// Make list of public and owned tables
//   - Pass tables from db store, can be nil arg
//   - Pass false gc for rechecks
func CreateTableList(gc bool, tables map[string]string) {
	if Gnomes.Init && !gc && !GnomonClosing() {
		var owner bool
		list := []string{}
		owned := []string{}
		if tables == nil {
			tables = Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
		}

		for scid := range tables {
			if !Gnomes.Init || GnomonClosing() {
				break
			}
			_, valid := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "Deck Count:", Gnomes.Indexer.LastIndexedHeight, true)
			_, version := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "V:", Gnomes.Indexer.LastIndexedHeight, true)

			if valid != nil && version != nil {
				d := valid[0]
				v := version[0]

				headers := GetSCHeaders(scid)
				name := "?"
				desc := "?"
				if headers != nil {
					if headers[1] != "" {
						desc = headers[1]
					}

					if headers[0] != "" {
						name = " " + headers[0]
					}
				}

				var hidden bool
				_, restrict := Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.RatingSCID, "restrict", Gnomes.Indexer.ChainHeight, true)
				_, rating := Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.RatingSCID, scid, Gnomes.Indexer.ChainHeight, true)

				if restrict != nil && rating != nil {
					Control.Contract_rating[scid] = rating[0]
					if rating[0] <= restrict[0] {
						hidden = true
					}
				}

				if d >= 1 && v == 110 && !hidden {
					list = append(list, name+"   "+desc+"   "+scid)
				}

				if d >= 1 && v >= 100 {
					if CheckTableOwner(scid) {
						owned = append(owned, name+"   "+desc+"   "+scid)
						Poker.Holdero_unlock.Hide()
						Poker.Holdero_new.Show()
						owner = true
						rpc.Wallet.PokerOwner = true
					}
				}
			}
		}

		if !owner {
			Poker.Holdero_unlock.Show()
			Poker.Holdero_new.Hide()
			rpc.Wallet.PokerOwner = false
		}

		t := len(list)
		list = append(list, "  Holdero Tables: "+strconv.Itoa(t))
		sort.Strings(list)
		Control.Holdero_tables = list

		sort.Strings(owned)
		Control.Holdero_owned = owned

		Poker.Table_list.Refresh()
		Poker.Owned_list.Refresh()
	}
}

// Get current Holdero table menu stats
func GetTableStats(scid string, single bool) {
	if len(scid) == 64 {
		_, v := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "V:", Gnomes.Indexer.LastIndexedHeight, true)
		_, l := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "Last", Gnomes.Indexer.LastIndexedHeight, true)
		_, s := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "Seats at Table:", Gnomes.Indexer.LastIndexedHeight, true)
		// _, o := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "Open", Gnomes.Indexer.LastIndexedHeight, true)
		// p1, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "Player 1 ID:", Gnomes.Indexer.LastIndexedHeight, true)
		p2, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "Player2 ID:", Gnomes.Indexer.LastIndexedHeight, true)
		p3, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "Player3 ID:", Gnomes.Indexer.LastIndexedHeight, true)
		p4, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "Player4 ID:", Gnomes.Indexer.LastIndexedHeight, true)
		p5, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "Player5 ID:", Gnomes.Indexer.LastIndexedHeight, true)
		p6, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "Player6 ID:", Gnomes.Indexer.LastIndexedHeight, true)
		h := GetSCHeaders(scid)

		if single {
			if h != nil {
				Stats.Name.Text = (" Name: " + h[0])
				Stats.Name.Refresh()
				Stats.Desc.Text = (" Description: " + h[1])
				Stats.Desc.Refresh()
				if len(h[2]) > 6 {
					Stats.Image, _ = holdero.DownloadFile(h[2], h[0])
				} else {
					Stats.Image = *canvas.NewImageFromImage(nil)
				}

			} else {
				Stats.Name.Text = (" Name: ?")
				Stats.Name.Refresh()
				Stats.Desc.Text = (" Description: ?")
				Stats.Desc.Refresh()
				Stats.Image = *canvas.NewImageFromImage(nil)
			}
		}

		if v != nil {
			Stats.Version.Text = (" Table Version: " + strconv.Itoa(int(v[0])))
			Stats.Version.Refresh()
		} else {
			Stats.Version.Text = (" Table Version: ?")
			Stats.Version.Refresh()
		}

		if l != nil {
			time, _ := rpc.MsToTime(strconv.Itoa(int(l[0]) * 1000))
			Stats.Last.Text = (" Last Move: " + time.String())
			Stats.Last.Refresh()
		} else {
			Stats.Last.Text = (" Last Move: ?")
			Stats.Last.Refresh()
		}

		if s != nil {
			if s[0] > 1 {
				sit := 1
				if p2 != nil {
					sit++
				}

				if p3 != nil {
					sit++
				}

				if p4 != nil {
					sit++
				}

				if p5 != nil {
					sit++
				}

				if p6 != nil {
					sit++
				}

				Stats.Seats.Text = (" Seats at Table: " + strconv.Itoa(int(s[0])-sit))
				Stats.Seats.Refresh()
			}
		} else {
			Stats.Seats.Text = (" Table Closed")
			Stats.Seats.Refresh()
		}
	}
}

// Get a wallets registered names
func CheckWalletNames(value string) {
	if Gnomes.Init && Gnomes.Sync && !GnomonClosing() {
		names, _ := Gnomes.Indexer.Backend.GetSCIDKeysByValue(rpc.NameSCID, value, Gnomes.Indexer.LastIndexedHeight, true)

		sort.Strings(names)
		Control.Names.Options = append(Control.Names.Options, names...)
	}
}

// Check if wallet owns in game G45 asset
//   - Pass g45s from db store, can be nil arg
//   - Pass false gc for rechecks
func CheckDreamsG45s(gc bool, g45s map[string]string) {
	if Gnomes.Init && Gnomes.Sync && !gc && !GnomonClosing() {
		if g45s == nil {
			g45s = Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
		}
		log.Println("[dReams] Checking G45 Assets")

		for scid := range g45s {
			if !rpc.Wallet.Connect || GnomonClosing() {
				break
			}
			data, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "metadata", Gnomes.Indexer.LastIndexedHeight, true)
			owner, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "owner", Gnomes.Indexer.LastIndexedHeight, true)
			minter, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "minter", Gnomes.Indexer.LastIndexedHeight, true)
			coll, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "collection", Gnomes.Indexer.LastIndexedHeight, true)
			if data != nil && owner != nil && minter != nil && coll != nil {
				if owner[0] == rpc.Wallet.Address {
					if minter[0] == holdero.Seals_mint && coll[0] == holdero.Seals_coll {
						var seal holdero.Seal
						if err := json.Unmarshal([]byte(data[0]), &seal); err == nil {
							Assets.Assets = append(Assets.Assets, seal.Name+"   "+scid)
							current := holdero.Settings.AvatarSelect.Options
							new := append(current, seal.Name)
							holdero.Settings.AvatarSelect.Options = new
							holdero.Settings.AvatarSelect.Refresh()
						}
					} else if minter[0] == holdero.ATeam_mint && coll[0] == holdero.ATeam_coll {
						var agent holdero.Agent
						if err := json.Unmarshal([]byte(data[0]), &agent); err == nil {
							Assets.Asset_map[agent.Name] = scid
							Assets.Assets = append(Assets.Assets, agent.Name+"   "+scid)
							current := holdero.Settings.AvatarSelect.Options
							new := append(current, agent.Name)
							holdero.Settings.AvatarSelect.Options = new
							holdero.Settings.AvatarSelect.Refresh()
						}
					}
				}
			}
		}
		sort.Strings(holdero.Settings.AvatarSelect.Options)
		holdero.Settings.AvatarSelect.Options = append([]string{"None"}, holdero.Settings.AvatarSelect.Options...)
		Assets.Asset_list.Refresh()
	}
}

// Check if dPrediction is live on SCID
func CheckActivePrediction(scid string) bool {
	if len(scid) == 64 && Gnomes.Init && !GnomonClosing() {
		_, ends := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_end_at", Gnomes.Indexer.ChainHeight, true)
		_, buff := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "buffer", Gnomes.Indexer.ChainHeight, true)
		if ends != nil && buff != nil {
			now := time.Now().Unix()
			if now < int64(ends[0]) && now > int64(buff[0]) {
				return true
			}
		}
	}
	return false
}

// Check for live dSports on SCID
func CheckActiveGames(scid string) bool {
	if Gnomes.Init && !GnomonClosing() {
		_, played := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_played", Gnomes.Indexer.ChainHeight, true)
		_, init := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_init", Gnomes.Indexer.ChainHeight, true)

		if played != nil && init != nil {
			return played[0] == init[0]
		}
	}

	return true
}

func GetSportsAmt(scid, n string) uint64 {
	_, amt := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_amount_"+n, Gnomes.Indexer.ChainHeight, true)
	if amt != nil {
		return amt[0]
	} else {
		return 0
	}
}

// Get current dSports game teams
func GetSportsTeams(scid, n string) (string, string) {
	game, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "game_"+n, Gnomes.Indexer.ChainHeight, true)

	if game != nil {
		team_a := TrimTeamA(game[0])
		team_b := TrimTeamB(game[0])

		if team_a != "" && team_b != "" {
			return team_a, team_b
		}
	}

	return "Team A", "Team B"
}

// Parse dSports game string into team A string
func TrimTeamA(s string) string {
	split := strings.Split(s, "--")

	if len(split) == 2 {
		return split[0]
	}

	return ""

}

// Parse dSports game string into team B string
func TrimTeamB(s string) string {
	split := strings.Split(s, "--")

	if len(split) == 2 {
		return split[1]
	}
	return ""
}

// Scan index for any active NFA listings
//   - Pass assets from db store, can be nil arg
func FindNfaListings(assets map[string]string) {
	if Gnomes.Init && Gnomes.Sync && !GnomonClosing() {
		auction := []string{" Collection,  Name,  Description,  SCID:"}
		buy_now := []string{" Collection,  Name,  Description,  SCID:"}
		if assets == nil {
			assets = Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
		}
		keys := make([]string, len(assets))

		i := 0
		for k := range assets {
			keys[i] = k

			a := checkNfaAuctionListing(keys[i])

			if a != "" {
				auction = append(auction, a)
			}

			b := checkNfaBuyListing(keys[i])

			if b != "" {
				buy_now = append(buy_now, b)
			}

			i++
		}

		if GnomonClosing() {
			return
		}

		Market.Auctions = auction
		Market.Buy_now = buy_now
		sort.Strings(Market.Auctions)
		sort.Strings(Market.Buy_now)

		Market.Auction_list.Refresh()
		Market.Buy_list.Refresh()
	}
}

// dReams NFA collections
func isDreamsNfaCollection(check string) bool {
	if check == "AZYDS" || check == "DBC" || check == "AZYPC" || check == "SIXPC" || check == "AZYPCB" || check == "SIXPCB" || check == "SIXART" || check == "HighStrangeness" {
		return true
	}

	return false
}

// Check if NFA SCID is listed for auction
//   - Market.Filter false for all NFA listings
func checkNfaAuctionListing(scid string) (asset string) {
	if !GnomonClosing() {
		listType, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "listType", Gnomes.Indexer.ChainHeight, true)
		header, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.ChainHeight, true)
		coll, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "collection", Gnomes.Indexer.ChainHeight, true)
		desc, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "descrHdr", Gnomes.Indexer.ChainHeight, true)
		if listType != nil && header != nil {
			if Market.Filter {
				check := strings.Trim(header[0], "0123456789")
				if isDreamsNfaCollection(check) {
					if listType[0] == "auction" {
						asset = coll[0] + "   " + header[0] + "   " + desc[0] + "   " + scid
					}
				}
			} else {
				if listType[0] == "auction" {
					asset = coll[0] + "   " + header[0] + "   " + desc[0] + "   " + scid
				}
			}
		}
	}

	return
}

// Check if NFA SCID is listed as buy now
//   - Market.Filter false for all NFA listings
func checkNfaBuyListing(scid string) (asset string) {
	if !GnomonClosing() {
		listType, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "listType", Gnomes.Indexer.ChainHeight, true)
		header, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.ChainHeight, true)
		coll, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "collection", Gnomes.Indexer.ChainHeight, true)
		desc, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "descrHdr", Gnomes.Indexer.ChainHeight, true)
		if listType != nil && header != nil {
			if Market.Filter {
				check := strings.Trim(header[0], "0123456789")
				if isDreamsNfaCollection(check) {
					if listType[0] == "sale" {
						asset = coll[0] + "   " + header[0] + "   " + desc[0] + "   " + scid
					}
				}
			} else {
				if listType[0] == "sale" {
					asset = coll[0] + "   " + header[0] + "   " + desc[0] + "   " + scid
				}
			}
		}
	}

	return
}

// Get NFA image files
func GetNfaImages(scid string) {
	if !GnomonClosing() && len(scid) == 64 {
		name, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.ChainHeight, true)
		icon, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "iconURLHdr", Gnomes.Indexer.LastIndexedHeight, true)
		cover, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "coverURL", Gnomes.Indexer.LastIndexedHeight, true)
		if icon != nil {
			Market.Icon, _ = holdero.DownloadFile(icon[0], name[0])
			Market.Cover, _ = holdero.DownloadFile(cover[0], name[0]+"-cover")
		} else {
			Market.Icon = *canvas.NewImageFromImage(nil)
			Market.Cover = *canvas.NewImageFromImage(nil)
		}
	}
}

// Create auction tab info for current asset
func GetAuctionDetails(scid string) {
	if !GnomonClosing() && len(scid) == 64 {
		name, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.ChainHeight, true)
		collection, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "collection", Gnomes.Indexer.ChainHeight, true)
		description, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "descrHdr", Gnomes.Indexer.ChainHeight, true)
		creator, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "creatorAddr", Gnomes.Indexer.ChainHeight, true)
		owner, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "owner", Gnomes.Indexer.ChainHeight, true)
		_, owner_update := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "ownerCanUpdate", Gnomes.Indexer.ChainHeight, true)
		_, start := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "startPrice", Gnomes.Indexer.ChainHeight, true)
		_, current := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "currBidAmt", Gnomes.Indexer.ChainHeight, true)
		_, bid_price := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "currBidPrice", Gnomes.Indexer.ChainHeight, true)
		_, royalty := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "royalty", Gnomes.Indexer.ChainHeight, true)
		_, bids := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "bidCount", Gnomes.Indexer.ChainHeight, true)
		_, endTime := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "endBlockTime", Gnomes.Indexer.ChainHeight, true)
		_, startTime := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "startBlockTime", Gnomes.Indexer.ChainHeight, true)
		_, artFee := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "artificerFee", Gnomes.Indexer.ChainHeight, true)

		if name != nil && collection != nil && start != nil && royalty != nil && endTime != nil && artFee != nil {
			go func() {
				var ty string
				check := strings.Trim(name[0], "0123456789")
				if check == "AZYPC" || check == "SIXPC" {
					ty = "Playing card deck"
				} else if check == "AZYPCB" || check == "SIXPCB" {
					ty = "Playing card back"
				} else if check == "AZYDS" || check == "SIXART" {
					ty = "Theme/Avatar"
				} else if check == "DBC" || check == "HighStrangeness" {
					ty = "Avatar"
				} else {
					return
				}

				Market.Viewing_coll = check
				Market.Name.Text = (" Name: " + name[0])
				Market.Name.Refresh()
				Market.Type.Text = (" Asset Type: " + ty)
				Market.Type.Refresh()
				Market.Collection.Text = (" Collection: " + collection[0])
				Market.Collection.Refresh()
				Market.Description.Text = (" Description: " + description[0])
				Market.Description.Refresh()

				Market.Creator.Text = (" Creator: " + creator[0])
				Market.Creator.Refresh()
				Market.Owner.Text = (" Owner: " + owner[0])
				Market.Owner.Refresh()
				if owner_update[0] == 1 {
					Market.Owner_update.Text = (" Owner can update: Yes")
				} else {
					Market.Owner_update.Text = (" Owner can update: No")
				}
				Market.Owner_update.Refresh()

				Market.Art_fee.Text = (" Artificer Fee: " + strconv.Itoa(int(artFee[0])) + "%")
				Market.Art_fee.Refresh()

				Market.Royalty.Text = (" Royalty: " + strconv.Itoa(int(royalty[0])) + "%")
				Market.Royalty.Refresh()
				price := float64(start[0])
				str := fmt.Sprintf("%.5f", price/100000)
				Market.Start_price.Text = (" Start Price: " + str + " Dero")
				Market.Start_price.Refresh()
				Market.Bid_count.Text = (" Bids: " + strconv.Itoa(int(bids[0])))
				Market.Bid_count.Refresh()

				end, _ := rpc.MsToTime(strconv.Itoa(int(endTime[0]) * 1000))
				Market.End_time.Text = (" Ends At: " + end.String())
				Market.End_time.Refresh()

				if current != nil {
					value := float64(current[0])
					str := fmt.Sprintf("%.5f", value/100000)
					Market.Current_bid.Text = (" Current Bid: " + str)
					Market.Current_bid.Refresh()
				} else {
					Market.Current_bid.Text = (" Current Bid: ")
					Market.Current_bid.Refresh()
				}

				if bid_price != nil {
					value := float64(bid_price[0])
					str := fmt.Sprintf("%.5f", value/100000)
					if bid_price[0] == 0 {
						Market.Bid_amt = start[0]
					} else {
						Market.Bid_amt = bid_price[0]
					}
					Market.Bid_price.Text = (" Minimum Bid: " + str)
					Market.Bid_price.Refresh()
				} else {
					Market.Bid_amt = 0
					Market.Bid_price.Text = (" Minimum Bid: ")
					Market.Bid_price.Refresh()
				}

				if amt, err := strconv.ParseFloat(Market.Entry.Text, 64); err == nil {
					value := float64(Market.Bid_amt) / 100000
					if amt == 0 || amt < value {
						amt := fmt.Sprintf("%.5f", value)
						Market.Entry.SetText(amt)
					}
				}

				now := uint64(time.Now().Unix())
				if owner[0] == rpc.Wallet.Address {
					if now < startTime[0]+300 && startTime[0] > 0 && !Market.Confirming {
						Market.Cancel_button.Show()
					} else {
						Market.Cancel_button.Hide()
					}

					if now > endTime[0] && endTime[0] > 0 && !Market.Confirming {
						Market.Close_button.Show()
					} else {
						Market.Close_button.Hide()
					}
				} else {
					Market.Close_button.Hide()
					Market.Cancel_button.Hide()
				}

				Market.Market_button.Show()
				if now > endTime[0] || Market.Confirming {
					Market.Market_button.Hide()
				}
			}()
		}
	}
}

// Create buy now tab info for current asset
func GetBuyNowDetails(scid string) {
	if !GnomonClosing() && len(scid) == 64 {
		name, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.ChainHeight, true)
		collection, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "collection", Gnomes.Indexer.ChainHeight, true)
		description, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "descrHdr", Gnomes.Indexer.ChainHeight, true)
		creator, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "creatorAddr", Gnomes.Indexer.ChainHeight, true)
		owner, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "owner", Gnomes.Indexer.ChainHeight, true)
		_, owner_update := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "ownerCanUpdate", Gnomes.Indexer.ChainHeight, true)
		_, start := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "startPrice", Gnomes.Indexer.ChainHeight, true)
		_, royalty := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "royalty", Gnomes.Indexer.ChainHeight, true)
		_, endTime := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "endBlockTime", Gnomes.Indexer.ChainHeight, true)
		_, startTime := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "startBlockTime", Gnomes.Indexer.ChainHeight, true)
		_, artFee := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "artificerFee", Gnomes.Indexer.ChainHeight, true)

		if name != nil && collection != nil && start != nil && royalty != nil && endTime != nil && artFee != nil {
			go func() {
				var ty string
				check := strings.Trim(name[0], "0123456789")
				if check == "AZYPC" || check == "SIXPC" {
					ty = "Playing card deck"
				} else if check == "AZYPCB" || check == "SIXPCB" {
					ty = "Playing card back"
				} else if check == "AZYDS" || check == "SIXART" {
					ty = "Theme/Avatar"
				} else if check == "DBC" || check == "HighStrangeness" {
					ty = "Avatar"
				} else {
					return
				}

				Market.Viewing_coll = check
				Market.Name.Text = (" Name: " + name[0])
				Market.Name.Refresh()
				Market.Type.Text = (" Asset Type: " + ty)
				Market.Type.Refresh()
				Market.Collection.Text = (" Collection: " + collection[0])
				Market.Collection.Refresh()
				Market.Description.Text = (" Description: " + description[0])
				Market.Description.Refresh()

				Market.Creator.Text = (" Creator: " + creator[0])
				Market.Creator.Refresh()
				Market.Owner.Text = (" Owner: " + owner[0])
				Market.Owner.Refresh()
				if owner_update[0] == 1 {
					Market.Owner_update.Text = (" Owner can update: Yes")
				} else {
					Market.Owner_update.Text = (" Owner can update: No")
				}
				Market.Owner_update.Refresh()

				Market.Art_fee.Text = (" Artificer Fee: " + strconv.Itoa(int(artFee[0])) + "%")
				Market.Art_fee.Refresh()

				Market.Royalty.Text = (" Royalty: " + strconv.Itoa(int(royalty[0])) + "%")
				Market.Royalty.Refresh()
				Market.Buy_amt = start[0]
				value := float64(start[0])
				str := fmt.Sprintf("%.5f", value/100000)
				Market.Start_price.Text = (" Buy now for: " + str + " Dero")
				Market.Start_price.Refresh()

				Market.Entry.SetText(str)
				Market.Entry.Disable()
				end, _ := rpc.MsToTime(strconv.Itoa(int(endTime[0]) * 1000))
				Market.End_time.Text = (" Ends At: " + end.String())
				Market.End_time.Refresh()

				now := uint64(time.Now().Unix())
				if owner[0] == rpc.Wallet.Address {
					if now < startTime[0]+300 && startTime[0] > 0 && !Market.Confirming {
						Market.Cancel_button.Show()
					} else {
						Market.Cancel_button.Hide()
					}

					if now > endTime[0] && endTime[0] > 0 && !Market.Confirming {
						Market.Close_button.Show()
					} else {
						Market.Close_button.Hide()
					}
				} else {
					Market.Close_button.Hide()
					Market.Cancel_button.Hide()
				}

				Market.Market_button.Show()
				if now > endTime[0] || Market.Confirming {
					Market.Market_button.Hide()
				}
			}()
		}
	}
}

// Get percentages for a NFA
func GetListingPercents(scid string) (artP float64, royaltyP float64) {
	if Gnomes.Init && Gnomes.Sync && !GnomonClosing() {
		_, artFee := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "artificerFee", Gnomes.Indexer.ChainHeight, true)
		_, royalty := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "royalty", Gnomes.Indexer.ChainHeight, true)

		if artFee != nil && royalty != nil {
			artP = float64(artFee[0]) / 100
			royaltyP = float64(royalty[0]) / 100

			return
		}
	}

	return
}
