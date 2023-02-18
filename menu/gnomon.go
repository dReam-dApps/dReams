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

	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/table"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/civilware/Gnomon/indexer"
	"github.com/civilware/Gnomon/storage"
	"github.com/civilware/Gnomon/structures"
)

const (
	nfa_search_filter = `Function init() Uint64
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

type gnomon struct {
	Para     int
	Trim     bool
	Fast     bool
	Start    bool
	Init     bool
	Sync     bool
	Checked  bool
	SCIDS    uint64
	Sync_ind *fyne.Animation
	Full_ind *fyne.Animation
	Icon_ind *fyne.Animation
	Indexer  *indexer.Indexer
}

var Gnomes gnomon

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

func startLabel() {
	table.Assets.Gnomes_sync.Text = (" Starting Gnomon")
	table.Assets.Gnomes_sync.Refresh()
}

func checkLabel() {
	table.Assets.Gnomes_sync.Text = (" Checking for Assets")
	table.Assets.Gnomes_sync.Refresh()
}

func stopLabel() {
	table.Assets.Gnomes_sync.Text = (" Putting Gnomon to Sleep")
	table.Assets.Gnomes_sync.Refresh()
}

func sleepLabel() {
	table.Assets.Gnomes_sync.Text = (" Gnomon is Sleeping")
	table.Assets.Gnomes_sync.Refresh()
}

func StartIndicators() fyne.CanvasObject {
	purple := color.RGBA{105, 90, 205, 210}
	blue := color.RGBA{31, 150, 200, 210}
	alpha := color.RGBA{0, 0, 0, 0}

	g_top := canvas.NewRectangle(color.Black)
	g_top.SetMinSize(fyne.NewSize(150, 12))

	g_bottom := canvas.NewRectangle(color.Black)
	g_bottom.SetMinSize(fyne.NewSize(150, 12))

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
	Gnomes.Sync_ind.Start()

	sync_box := container.NewVBox(
		g_top,
		layout.NewSpacer(),
		g_bottom)

	g_full := canvas.NewRectangle(color.Black)
	g_full.SetMinSize(fyne.NewSize(150, 36))

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
	Gnomes.Full_ind.Start()

	icon := widget.NewIcon(Resource.Gnomon)
	Gnomes.Icon_ind = canvas.NewPositionAnimation(fyne.NewPos(3, 4), fyne.NewPos(112, 1), time.Second*3, func(p fyne.Position) {
		icon.Move(p)
		width := 30 + (p.X / 30)
		icon.Resize(fyne.NewSize(width, width))
	})

	Gnomes.Icon_ind.RepeatCount = fyne.AnimationRepeatForever
	Gnomes.Icon_ind.AutoReverse = true
	Gnomes.Icon_ind.Curve = fyne.AnimationEaseInOut
	Gnomes.Icon_ind.Start()

	d_rect := canvas.NewRectangle(color.Black)
	d_rect.SetMinSize(fyne.NewSize(36, 36))

	MenuControl.Daemon_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if rpc.Signal.Daemon {
				d_rect.FillColor = c
				canvas.Refresh(d_rect)
			} else {
				d_rect.FillColor = alpha
				canvas.Refresh(d_rect)
			}
		})

	MenuControl.Daemon_ind.RepeatCount = fyne.AnimationRepeatForever
	MenuControl.Daemon_ind.AutoReverse = true
	MenuControl.Daemon_ind.Start()

	w_rect := canvas.NewRectangle(color.Black)
	w_rect.SetMinSize(fyne.NewSize(36, 36))

	MenuControl.Wallet_ind = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if rpc.Wallet.Connect {
				w_rect.FillColor = c
				canvas.Refresh(w_rect)
			} else {
				w_rect.FillColor = alpha
				canvas.Refresh(w_rect)
			}
		})
	MenuControl.Wallet_ind.RepeatCount = fyne.AnimationRepeatForever
	MenuControl.Wallet_ind.AutoReverse = true
	MenuControl.Wallet_ind.Start()

	d := canvas.NewText("D", color.White)
	d.TextStyle.Bold = true
	w := canvas.NewText("W", color.White)
	w.TextStyle.Bold = true

	hbox := container.NewHBox(
		container.NewMax(d_rect, container.NewCenter(d)),
		container.NewMax(w_rect, container.NewCenter(w)))

	top_box := container.NewHBox(layout.NewSpacer(), hbox, container.NewMax(g_full, sync_box, icon))
	place := container.NewVBox(top_box, layout.NewSpacer())

	return container.NewMax(place)
}

func StopIndicators() {
	Gnomes.Icon_ind.Stop()
	Gnomes.Sync_ind.Stop()
	Gnomes.Full_ind.Stop()
	MenuControl.Daemon_ind.Stop()
	MenuControl.Wallet_ind.Stop()
}

func searchFilters() (filter []string) {
	holdero110, _ := rpc.GetHoldero110Code(rpc.Signal.Daemon, 0)
	if holdero110 != "" {
		filter = append(filter, holdero110)
	}

	holdero100, _ := rpc.GetHoldero100Code(rpc.Signal.Daemon)
	if holdero100 != "" {
		filter = append(filter, holdero100)
	}

	bacc, _ := rpc.GetBaccCode(rpc.Signal.Daemon)
	if bacc != "" {
		filter = append(filter, bacc)
	}

	predict, _ := rpc.GetPredictCode(rpc.Signal.Daemon, 0)
	if predict != "" {
		filter = append(filter, predict)
	}

	sports, _ := rpc.GetSportsCode(rpc.Signal.Daemon, 0)
	if sports != "" {
		filter = append(filter, sports)
	}

	gnomon, _ := rpc.GetGnomonCode(rpc.Signal.Daemon, 0)
	if sports != "" {
		filter = append(filter, gnomon)
	}

	filter = append(filter, nfa_search_filter)
	if !Gnomes.Trim {
		filter = append(filter, g45_search_filter)
	}

	return filter
}

func manualIndex(scid []string) {
	filters := Gnomes.Indexer.SearchFilter
	Gnomes.Indexer.SearchFilter = []string{}
	scidstoadd := make(map[string]*structures.FastSyncImport)

	for i := range scid {
		owner, _ := rpc.CheckForIndex(scid[i])

		scidstoadd[scid[i]] = &structures.FastSyncImport{}
		scidstoadd[scid[i]].Owner = owner.(string)
	}

	err := Gnomes.Indexer.AddSCIDToIndex(scidstoadd)
	if err != nil {
		log.Printf("Err - %v", err)
	}
	Gnomes.Indexer.SearchFilter = filters
}

func GnomonDB() *storage.GravitonStore {
	shasum := fmt.Sprintf("%x", sha1.Sum([]byte("gnomon")))
	db_folder := fmt.Sprintf("gnomondb\\%s_%s", "GNOMON", shasum)
	db := storage.NewGravDB(db_folder, "25ms")

	return db
}

func startGnomon(ep string) {
	Gnomes.Start = true
	go startLabel()
	log.Println("[dReams] Starting Gnomon")
	backend := GnomonDB()

	last_height := backend.GetLastIndexHeight()
	daemon_endpoint := ep
	runmode := "daemon"
	mbl := false
	closeondisconnect := false

	filters := searchFilters()
	search := len(filters)
	if search == 8 || (Gnomes.Trim && search == 7) {
		table.Assets.Asset_map = make(map[string]string)
		Gnomes.Indexer = indexer.NewIndexer(backend, filters, last_height, daemon_endpoint, runmode, mbl, closeondisconnect, Gnomes.Fast)
		go Gnomes.Indexer.StartDaemonMode(Gnomes.Para)
		time.Sleep(3 * time.Second)
		Gnomes.Init = true

		if Gnomes.Trim {
			i := 0
			for {
				contracts := len(Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs())
				if contracts >= 3960 {
					Gnomes.Trim = false
					break
				}

				if contracts >= 1 {
					go g45Index()
					break
				}
				time.Sleep(1 * time.Second)
				i++
				if i == 30 {
					Gnomes.Trim = false
					log.Println("[dReams] Could not add G45 Collections")
					break
				}
			}
		}
	}

	Gnomes.Start = false
}

func g45Index() {
	log.Println("[dReams] Adding G45 Collections")
	filters := Gnomes.Indexer.SearchFilter
	Gnomes.Indexer.SearchFilter = []string{}
	scidstoadd := make(map[string]*structures.FastSyncImport)

	a, _ := rpc.GetG45Collection(table.ATeam_coll)
	for i := range a {
		scidstoadd[a[i]] = &structures.FastSyncImport{}
	}

	s, _ := rpc.GetG45Collection(table.Seals_coll)
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

func GnomonEndPoint(dc, gi, gs bool) {
	if dc && gi && gs {
		Gnomes.Indexer.Endpoint = rpc.Round.Daemon
	}
}

func StopGnomon(gi bool) {
	if gi && !GnomonClosing() {
		go stopLabel()
		log.Println("[dReams] Putting Gnomon to Sleep")
		Gnomes.Indexer.Close()
		Gnomes.Init = false
		time.Sleep(1 * time.Second)
		log.Println("[dReams] Gnomon is Sleeping")
		go sleepLabel()
	}
}

func GnomonWriting() bool {
	return Gnomes.Indexer.Backend.Writing == 1
}

func GnomonClosing() bool {
	if !Gnomes.Init {
		return false
	}

	if Gnomes.Indexer.Closing || Gnomes.Indexer.Backend.Closing {
		return true
	}

	return false
}

func FastSynced() bool {
	return Gnomes.SCIDS > 0
}

func Connected() bool {
	if rpc.Signal.Daemon && rpc.Wallet.Connect && Gnomes.Sync {
		return true
	}

	return false
}

func GnomonState(dc, gi bool, windows bool) {
	if dc && Gnomes.Init && !GnomonWriting() && !GnomonClosing() {
		contracts := Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
		Gnomes.SCIDS = uint64(len(contracts))
		if FastSynced() && !Gnomes.Trim {
			height := stringToInt64(rpc.Wallet.Height)
			if Gnomes.Indexer.ChainHeight >= height-1 && height != 0 && !GnomonClosing() {
				Gnomes.Sync = true
				if rpc.Wallet.Connect {
					if !Gnomes.Checked {
						go CheckBetContractOwner(Gnomes.Sync, Gnomes.Checked, contracts)
						CreateTableList(Gnomes.Checked, contracts)
						go CheckG45Assets(Gnomes.Sync, Gnomes.Checked, contracts)
						go CheckAssets(Gnomes.Sync, Gnomes.Checked, contracts)
						if !windows {
							go PopulateSports(rpc.Signal.Daemon, Gnomes.Sync, contracts)
							go PopulatePredictions(rpc.Signal.Daemon, Gnomes.Sync, contracts)
							FindNfaListings(Gnomes.Sync, contracts)
						}
						Gnomes.Checked = true
					}
				}
			} else {
				Gnomes.Sync = false
			}
		}

		table.Assets.Stats_box = *container.NewVBox(table.Assets.Collection, table.Assets.Name, table.IconImg(Resource.Frame))
		table.Assets.Stats_box.Refresh()
		HolderoControl.Stats_box = *container.NewVBox(Stats.Name, Stats.Desc, Stats.Version, Stats.Last, Stats.Seats, TableIcon(Resource.Frame))
		HolderoControl.Stats_box.Refresh()

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

func CheckAssets(gs, gc bool, scids map[string]string) {
	if gs && !gc && !GnomonClosing() {
		go checkLabel()
		if scids == nil {
			scids = Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
		}
		keys := make([]string, len(scids))
		log.Println("[dReams] Checking NFA Assets")
		table.Settings.FaceSelect.Options = []string{}
		table.Settings.BackSelect.Options = []string{}
		table.Settings.ThemeSelect.Options = []string{}
		table.Settings.AvatarSelect.Options = []string{}

		i := 0
		for k := range scids {
			keys[i] = k
			checkNFAOwner(keys[i])
			i++
		}
		sort.Strings(table.Settings.FaceSelect.Options)
		sort.Strings(table.Settings.BackSelect.Options)
		sort.Strings(table.Settings.ThemeSelect.Options)

		ld := []string{"Light", "Dark"}
		table.Settings.FaceSelect.Options = append(ld, table.Settings.FaceSelect.Options...)
		table.Settings.BackSelect.Options = append(ld, table.Settings.BackSelect.Options...)
		table.Settings.ThemeSelect.Options = append([]string{"Main"}, table.Settings.ThemeSelect.Options...)

		table.Assets.Asset_list.Refresh()
		table.DisableHolderoTools()
	}
	sort.Strings(table.Assets.Assets)
}

func CheckBetContractOwner(gs, gc bool, contracts map[string]string) {
	if gs && !gc && !GnomonClosing() {
		if contracts == nil {
			contracts = Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
		}
		keys := make([]string, len(contracts))

		i := 0
		for k := range contracts {
			keys[i] = k
			verifyBetContractOwner(keys[i], "p")
			verifyBetContractOwner(keys[i], "s")
			i++
		}
	}
}

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

func verifyBetContractOwner(scid, t string) {
	if !GnomonClosing() {
		owner, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "owner", Gnomes.Indexer.ChainHeight, true)
		dev, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "dev", Gnomes.Indexer.ChainHeight, true)
		_, init := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, t+"_init", Gnomes.Indexer.ChainHeight, true)

		if owner != nil && dev != nil && init != nil {
			if dev[0] == rpc.DevAddress {
				DisableBetOwner(owner[0])
			}
		}
	}
}

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
				if headers != nil {
					if headers[1] != "" {
						desc = headers[1]
					}

					if headers[0] != "" {
						name = " " + headers[0]
					}
				}

				if owner[0] == rpc.Wallet.Address {
					owned = append(owned, name+"   "+desc+"   "+scid)
				}

				list = append(list, name+"   "+desc+"   "+scid)
				DisableBetOwner(owner[0])
			}
		}
	}
	return list, owned
}

func PopulatePredictions(dc, gs bool, contracts map[string]string) {
	if dc && gs && !GnomonClosing() {
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
		MenuControl.Predict_contracts = list

		sort.Strings(owned)
		MenuControl.Predict_owned = owned

	}
}

func PopulateSports(dc, gs bool, contracts map[string]string) {
	if dc && gs && !GnomonClosing() {
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
		MenuControl.Sports_contracts = list

		sort.Strings(owned)
		MenuControl.Sports_owned = owned
	}
}

func isNfa(scid string) bool {
	artAddr, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "artificerAddr", Gnomes.Indexer.ChainHeight, true)
	if artAddr != nil {
		return artAddr[0] == rpc.ArtAddress
	}
	return false
}

func validNfa(file string) bool {
	return file != "-"
}

func checkNFAOwner(scid string) {
	if !GnomonClosing() {
		owner, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "owner", Gnomes.Indexer.ChainHeight, true)
		header, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.ChainHeight, true)
		file, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "fileURL", Gnomes.Indexer.ChainHeight, true)
		if owner != nil && header != nil && file != nil {
			if owner[0] == rpc.Wallet.Address && validNfa(file[0]) {
				check := strings.Trim(header[0], "0123456789")
				if check == "AZYDS" || check == "SIXART" {
					themes := table.Settings.ThemeSelect.Options
					new_themes := append(themes, header[0])
					table.Settings.ThemeSelect.Options = new_themes
					table.Settings.ThemeSelect.Refresh()

					avatars := table.Settings.AvatarSelect.Options
					new_avatar := append(avatars, header[0])
					table.Settings.AvatarSelect.Options = new_avatar
					table.Settings.AvatarSelect.Refresh()
					table.Assets.Assets = append(table.Assets.Assets, header[0]+"   "+scid)
				} else if check == "AZYPCB" || check == "SIXPCB" {
					current := table.Settings.BackSelect.Options
					new := append(current, header[0])
					table.Settings.BackSelect.Options = new
					table.Settings.BackSelect.Refresh()
					table.Assets.Assets = append(table.Assets.Assets, header[0]+"   "+scid)
				} else if check == "AZYPC" || check == "SIXPC" {
					current := table.Settings.FaceSelect.Options
					new := append(current, header[0])
					table.Settings.FaceSelect.Options = new
					table.Settings.FaceSelect.Refresh()
					table.Assets.Assets = append(table.Assets.Assets, header[0]+"   "+scid)
				} else if check == "DBC" {
					current := table.Settings.AvatarSelect.Options
					new := append(current, header[0])
					table.Settings.AvatarSelect.Options = new
					table.Settings.AvatarSelect.Refresh()
					table.Assets.Assets = append(table.Assets.Assets, header[0]+"   "+scid)
				}
			}
		}
	}
}

func GetOwnedAssetStats(scid string) {
	if Gnomes.Init && !GnomonClosing() {
		n, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.LastIndexedHeight, true)
		if n != nil {
			c, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "collection", Gnomes.Indexer.LastIndexedHeight, true)
			//d, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "descrHdr:", Gnomes.Indexer.LastIndexedHeight, true)
			i, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "iconURLHdr", Gnomes.Indexer.LastIndexedHeight, true)

			if n != nil {
				table.Assets.Name.Text = (" Name: " + n[0])
				table.Assets.Name.Refresh()
				MenuControl.List_button.Show()

			} else {
				table.Assets.Name.Text = (" Name: ?")
				table.Assets.Name.Refresh()
			}

			if c != nil {
				table.Assets.Collection.Text = (" Collection: " + c[0])
				table.Assets.Collection.Refresh()
			} else {
				table.Assets.Collection.Text = (" Collection: ?")
				table.Assets.Collection.Refresh()
			}

			if i != nil {
				table.Assets.Icon, _ = table.DownloadFile(i[0], n[0])
			} else {
				table.Assets.Icon = *canvas.NewImageFromImage(nil)
			}

		} else {
			MenuControl.List_button.Hide()
			data, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "metadata", Gnomes.Indexer.LastIndexedHeight, true)
			minter, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "minter", Gnomes.Indexer.LastIndexedHeight, true)
			coll, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "collection", Gnomes.Indexer.LastIndexedHeight, true)
			if data != nil && minter != nil && coll != nil {
				if minter[0] == table.Seals_mint && coll[0] == table.Seals_coll {
					var seal table.Seal
					if err := json.Unmarshal([]byte(data[0]), &seal); err == nil {
						check := strings.Trim(seal.Name, " #0123456789")
						if check == "Dero Seals" {
							table.Assets.Name.Text = (" Name: " + seal.Name)
							table.Assets.Name.Refresh()

							table.Assets.Collection.Text = (" Collection: " + check)
							table.Assets.Collection.Refresh()

							number := strings.Trim(seal.Name, "DeroSals# ")
							table.Assets.Icon, _ = table.DownloadFile("https://ipfs.io/ipfs/QmP3HnzWpiaBA6ZE8c3dy5ExeG7hnYjSqkNfVbeVW5iEp6/low/"+number+".jpg", seal.Name)
						}
					}
				} else if minter[0] == table.ATeam_mint && coll[0] == table.ATeam_coll {
					var agent table.Agent
					if err := json.Unmarshal([]byte(data[0]), &agent); err == nil {
						table.Assets.Name.Text = (" Name: " + agent.Name)
						table.Assets.Name.Refresh()

						table.Assets.Collection.Text = (" Collection: Dero A-Team")
						table.Assets.Collection.Refresh()

						number := strconv.Itoa(agent.ID)
						if agent.ID < 172 {
							table.Assets.Icon, _ = table.DownloadFile("https://ipfs.io/ipfs/QmaRHXcQwbFdUAvwbjgpDtr5kwGiNpkCM2eDBzAbvhD7wh/low/"+number+".jpg", agent.Name)
						} else {
							table.Assets.Icon, _ = table.DownloadFile("https://ipfs.io/ipfs/QmQQyKoE9qDnzybeDCXhyMhwQcPmLaVy3AyYAzzC2zMauW/low/"+number+".jpg", agent.Name)
						}
					}
				}
			}
		}
	}
}

func checkTableOwner(scid string) bool {
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

func checkTableVersion(scid string) uint64 {
	_, v := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "V:", Gnomes.Indexer.LastIndexedHeight, true)

	if v != nil && v[0] >= 100 {
		return v[0]
	}
	return 0
}

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

				if d >= 1 && v == 110 {
					list = append(list, name+"   "+desc+"   "+scid)
				}

				if d >= 1 && v >= 100 {
					if checkTableOwner(scid) {
						owned = append(owned, name+"   "+desc+"   "+scid)
						HolderoControl.holdero_unlock.Hide()
						HolderoControl.holdero_new.Show()
						owner = true
						rpc.Wallet.PokerOwner = true
					}
				}
			}
		}

		if !owner {
			HolderoControl.holdero_unlock.Show()
			HolderoControl.holdero_new.Hide()
			rpc.Wallet.PokerOwner = false
		}

		t := len(list)
		list = append(list, "  Holdero Tables: "+strconv.Itoa(t))
		sort.Strings(list)
		MenuControl.Holdero_tables = list

		sort.Strings(owned)
		MenuControl.Holdero_owned = owned

		HolderoControl.Table_list.Refresh()
		HolderoControl.Owned_list.Refresh()
	}
}

type tableStats struct {
	Name    *canvas.Text
	Desc    *canvas.Text
	Version *canvas.Text
	Last    *canvas.Text
	Seats   *canvas.Text
	Open    *canvas.Text
	Image   canvas.Image
}

var Stats tableStats

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
					Stats.Image, _ = table.DownloadFile(h[2], h[0])
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

func CheckG45Assets(gs, gc bool, g45s map[string]string) {
	if Gnomes.Init && gs && !gc && !GnomonClosing() {
		if g45s == nil {
			g45s = Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
		}
		log.Println("[dReams] Checking G45 Assets")

		for scid := range g45s {
			if GnomonClosing() {
				break
			}
			data, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "metadata", Gnomes.Indexer.LastIndexedHeight, true)
			owner, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "owner", Gnomes.Indexer.LastIndexedHeight, true)
			minter, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "minter", Gnomes.Indexer.LastIndexedHeight, true)
			coll, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "collection", Gnomes.Indexer.LastIndexedHeight, true)
			if data != nil && owner != nil && minter != nil && coll != nil {
				if owner[0] == rpc.Wallet.Address {
					if minter[0] == table.Seals_mint && coll[0] == table.Seals_coll {
						var seal table.Seal
						if err := json.Unmarshal([]byte(data[0]), &seal); err == nil {
							table.Assets.Assets = append(table.Assets.Assets, seal.Name+"   "+scid)
							current := table.Settings.AvatarSelect.Options
							new := append(current, seal.Name)
							table.Settings.AvatarSelect.Options = new
							table.Settings.AvatarSelect.Refresh()
						}
					} else if minter[0] == table.ATeam_mint && coll[0] == table.ATeam_coll {
						var agent table.Agent
						if err := json.Unmarshal([]byte(data[0]), &agent); err == nil {
							table.Assets.Asset_map[agent.Name] = scid
							table.Assets.Assets = append(table.Assets.Assets, agent.Name+"   "+scid)
							current := table.Settings.AvatarSelect.Options
							new := append(current, agent.Name)
							table.Settings.AvatarSelect.Options = new
							table.Settings.AvatarSelect.Refresh()
						}
					}
				}
			}
		}
		sort.Strings(table.Settings.AvatarSelect.Options)
		table.Settings.AvatarSelect.Options = append([]string{"None"}, table.Settings.AvatarSelect.Options...)
		table.Assets.Asset_list.Refresh()

	}
}

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

func CheckPredictionName(scid string) (name string) {
	if len(scid) == 64 && Gnomes.Init && !GnomonClosing() {
		check := Gnomes.Indexer.Backend.GetAllSCIDVariableDetails(scid)
		if check != nil {
			keys := make([]int64, 0, len(check))
			for k := range check {
				keys = append(keys, k)
			}

			sort.Slice(keys, func(i, j int) bool { return keys[i] > keys[j] })
			for val := range check[keys[0]] {
				v := check[keys[0]][val].Key
				if len(v.(string)) == 66 {
					addr := rpc.DeroAddress(v.(string))
					if addr == rpc.Wallet.Address {
						value, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, v, Gnomes.Indexer.ChainHeight, true)
						if value != nil {
							split := strings.Split(value[0], "_")
							name = split[1]
							table.Actions.NameEntry.Disable()
							table.Actions.Change.Show()
							return
						}

					}
				}
			}
		}
	}
	table.Actions.NameEntry.Enable()
	table.Actions.Change.Hide()
	return
}

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

func TrimTeamA(s string) string {
	split := strings.Split(s, "--")

	if len(split) == 2 {
		return split[0]
	}

	return ""

}

func TrimTeamB(s string) string {
	split := strings.Split(s, "--")

	if len(split) == 2 {
		return split[1]
	}
	return ""
}

func FindNfaListings(gs bool, assets map[string]string) {
	if Gnomes.Init && gs && !GnomonClosing() {
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

func checkNfaAuctionListing(scid string) string {
	if !GnomonClosing() {
		listType, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "listType", Gnomes.Indexer.ChainHeight, true)
		header, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.ChainHeight, true)
		coll, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "collection", Gnomes.Indexer.ChainHeight, true)
		desc, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "descrHdr", Gnomes.Indexer.ChainHeight, true)
		if listType != nil && header != nil {
			check := strings.Trim(header[0], "0123456789")
			if check == "AZYDS" || check == "DBC" || check == "AZYPC" || check == "SIXPC" || check == "AZYPCB" || check == "SIXPCB" {
				switch listType[0] {
				case "auction":
					return coll[0] + "   " + header[0] + "   " + desc[0] + "   " + scid
				default:
					return ""
				}
			}
		}
	}

	return ""
}

func checkNfaBuyListing(scid string) string {
	if !GnomonClosing() {
		listType, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "listType", Gnomes.Indexer.ChainHeight, true)
		header, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.ChainHeight, true)
		coll, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "collection", Gnomes.Indexer.ChainHeight, true)
		desc, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "descrHdr", Gnomes.Indexer.ChainHeight, true)
		if listType != nil && header != nil {
			check := strings.Trim(header[0], "0123456789")
			if check == "AZYDS" || check == "DBC" || check == "AZYPC" || check == "SIXPC" || check == "AZYPCB" || check == "SIXPCB" {
				switch listType[0] {
				case "sale":
					return coll[0] + "   " + header[0] + "   " + desc[0] + "   " + scid
				default:
					return ""
				}
			}
		}
	}

	return ""
}

func GetNfaImages(scid string) {
	if len(scid) == 64 {
		name, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.ChainHeight, true)
		icon, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "iconURLHdr", Gnomes.Indexer.LastIndexedHeight, true)
		cover, _ := Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "coverURL", Gnomes.Indexer.LastIndexedHeight, true)
		if icon != nil {
			Market.Icon, _ = table.DownloadFile(icon[0], name[0])
			Market.Cover, _ = table.DownloadFile(cover[0], name[0]+"-cover")
		} else {
			Market.Icon = *canvas.NewImageFromImage(nil)
			Market.Cover = *canvas.NewImageFromImage(nil)
		}
	}
}

func GetAuctionDetails(scid string) {
	if len(scid) == 64 {
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

		if name != nil {
			var ty string
			check := strings.Trim(name[0], "0123456789")
			if check == "AZYPC" || check == "SIXPC" {
				ty = "Playing card deck"
			} else if check == "AZYPCB" || check == "SIXPCB" {
				ty = "Playing card back"
			} else if check == "AZYDS" || check == "SIXART" {
				ty = "Theme/Avatar"
			} else if check == "DBC" {
				ty = "Avatar"
			}

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
				Market.Bid_amt = bid_price[0]
				Market.Bid_price.Text = (" Minimum Bid: " + str)
				Market.Bid_price.Refresh()
			} else {
				Market.Bid_amt = 0
				Market.Bid_price.Text = (" Minimum Bid: ")
				Market.Bid_price.Refresh()
			}

			if owner[0] == rpc.Wallet.Address {
				now := uint64(time.Now().Unix())

				if now < startTime[0]+300 && startTime[0] > 0 {
					Market.Cancel_button.Show()
				} else {
					Market.Cancel_button.Hide()
				}

				if now > endTime[0] && endTime[0] > 0 {
					Market.Close_button.Show()
				} else {
					Market.Close_button.Hide()
				}
			}
		}
	}
}

func GetBuyNowDetails(scid string) {
	if len(scid) == 64 {
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

		if name != nil {
			var ty string
			check := strings.Trim(name[0], "0123456789")
			if check == "AZYPC" || check == "SIXPC" {
				ty = "Playing card deck"
			} else if check == "AZYPCB" || check == "SIXPCB" {
				ty = "Playing card back"
			} else if check == "AZYDS" || check == "SIXART" {
				ty = "Theme/Avatar"
			} else if check == "DBC" {
				ty = "Avatar"
			}

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

			if owner[0] == rpc.Wallet.Address {
				now := uint64(time.Now().Unix())

				if now < startTime[0]+300 && startTime[0] > 0 {
					Market.Cancel_button.Show()
				} else {
					Market.Cancel_button.Hide()
				}

				if now > endTime[0] && endTime[0] > 0 {
					Market.Close_button.Show()
				} else {
					Market.Close_button.Hide()
				}
			}
		}
	}
}
