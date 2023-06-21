package menu

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	dreams "github.com/SixofClubsss/dReams"
	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/rpc"

	"github.com/civilware/Gnomon/indexer"
	"github.com/civilware/Gnomon/storage"
	"github.com/civilware/Gnomon/structures"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
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

// dReams search filters for Gnomon index
func GnomonFilters() (filter []string) {
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

	return
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
//   - If dbType is boltdb, will return nil gravdb
func GnomonGravDB(dbType, dbPath string) *storage.GravitonStore {
	if dbType == "boltdb" {
		return nil
	}

	db, err := storage.NewGravDB(dbPath, "25ms")
	if err != nil {
		log.Fatalf("[GnomonGravDB] %s\n", err)
	}

	return db
}

// Create Gnomon bbolt db with dReams tag
//   - If dbType is not boltdb, will return nil boltdb
func GnomonBoltDB(dbType, dbPath string) *storage.BboltStore {
	if dbType != "boltdb" {
		return nil
	}

	shasum := fmt.Sprintf("%x", sha1.Sum([]byte("dReams")))
	db_name := fmt.Sprintf("gnomondb_bolt_%s_%s.db", "dReams", shasum)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dbPath, 0755); err != nil {
			log.Fatalf("[GnomonBoltDB] %s\n", err)
		}
	}

	db, err := storage.NewBBoltDB(dbPath, filepath.Join(dbPath, db_name))
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	return db
}

// Start Gnomon indexer with or without search filters
//   - End point from rpc.Daemon.Rpc
//   - tag for log print
//   - dbtype defines gravdb or boltdb
//   - Passing nil filters with Gnomes.Trim false will run a full Gnomon index
//   - custom func() is for adding specific SCID to index on Gnomon start, Gnomes.Trim false will bypass
//   - lower defines the lower limit of indexed SCIDs from Gnomon search filters before custom adds
//   - upper defines the higher limit when custom indexed SCIDs exist already
func StartGnomon(tag, dbtype string, filters []string, upper, lower int, custom func()) {
	Gnomes.Start = true
	log.Printf("[%s] Starting Gnomon\n", tag)
	shasum := fmt.Sprintf("%x", sha1.Sum([]byte("dReams")))
	db_path := filepath.Join("gnomondb", fmt.Sprintf("%s_%s", "dReams", shasum))
	bolt_backend := GnomonBoltDB(dbtype, db_path)
	grav_backend := GnomonGravDB(dbtype, db_path)

	var last_height int64
	if dbtype == "boltdb" {
		last_height, _ = bolt_backend.GetLastIndexHeight()
	} else {
		last_height = grav_backend.GetLastIndexHeight()
	}

	runmode := "daemon"
	mbl := false
	closeondisconnect := false

	if filters != nil || !Gnomes.Trim {
		Gnomes.Indexer = indexer.NewIndexer(grav_backend, bolt_backend, dbtype, filters, last_height, rpc.Daemon.Rpc, runmode, mbl, closeondisconnect, Gnomes.Fast)
		go Gnomes.Indexer.StartDaemonMode(Gnomes.Para)
		time.Sleep(3 * time.Second)
		Gnomes.Initialized(true)

		if Gnomes.Trim {
			i := 0
			for {
				contracts := len(Gnomes.GetAllOwnersAndSCIDs())
				if contracts >= upper {
					Gnomes.Trim = false
					break
				}

				if contracts >= lower {
					custom()
					break
				}
				time.Sleep(time.Second)
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
func G45Index() {
	log.Println("[dReams] Adding G45 Collections")
	filters := Gnomes.Indexer.SearchFilter
	Gnomes.Indexer.SearchFilter = []string{}
	scidstoadd := make(map[string]*structures.FastSyncImport)

	a := rpc.GetG45Collection(dreams.ATeam_coll)
	for i := range a {
		scidstoadd[a[i]] = &structures.FastSyncImport{}
	}

	s := rpc.GetG45Collection(dreams.Seals_coll)
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
	if rpc.Daemon.IsConnected() && Gnomes.IsInitialized() && !Gnomes.IsScanning() {
		Gnomes.Indexer.Endpoint = rpc.Daemon.Rpc
	}
}

// Check three connection signals
func Connected() bool {
	if rpc.IsReady() && Gnomes.IsSynced() {
		return true
	}

	return false
}

func GnomonScan(config bool) bool {
	if Gnomes.IsSynced() && !config {
		return true
	}

	return false
}

// Gnomon will scan connected wallet on start up, then ensure sync
//   - Hold out checking if dReams is in configure
//   - windows disables certain initial sync routines from running on windows os
func GnomonState(windows, config bool, scan func(map[string]string)) {
	if rpc.Daemon.IsConnected() && Gnomes.IsRunning() {
		contracts := Gnomes.IndexContains()
		if Gnomes.HasIndex(2) && !Gnomes.Trim {
			height := Gnomes.Indexer.ChainHeight
			if Gnomes.IsRunning() && Gnomes.Indexer.LastIndexedHeight >= height-3 && height != 0 {
				Gnomes.Synced(true)
				if !config && rpc.Wallet.IsConnected() && !Gnomes.Check {
					Gnomes.Scanning(true)
					if Control.Dapp_list["dSports and dPredictions"] {
						go CheckBetContractOwners(contracts)
						if !windows {
							go PopulateSports(contracts)
							go PopulatePredictions(contracts)
						}
					}

					CheckWalletNames(rpc.Wallet.Address)
					scan(contracts)

					if !windows {
						FindNfaListings(contracts)
					}
					Gnomes.Checked(true)
					Gnomes.Scanning(false)
				}
			} else {
				Gnomes.Synced(false)
			}
		}

		Assets.Stats_box = *container.NewVBox(Assets.Collection, Assets.Name, IconImg(bundle.ResourceAvatarFramePng))
		Assets.Stats_box.Refresh()

		// Update live market info
		if len(Market.Viewing) == 64 && rpc.Wallet.IsConnected() {
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
		all := Gnomes.GetAllOwnersAndSCIDs()
		for sc := range all {
			if scid == sc {
				log.Println("[dReams] " + scid + " Indexed")
				found = true
			}
		}
		if !found {
			log.Println("[dReams] " + scid + " Not Found")
		}
	}
}

// Check wallet for all indexed NFAs
//   - Pass scids from db store, can be nil arg
//   - Pass false gc for rechecks
func CheckAllNFAs(gc bool, scids map[string]string) {
	if Gnomes.IsReady() && !gc {
		if scids == nil {
			scids = Gnomes.GetAllOwnersAndSCIDs()
		}
		keys := make([]string, len(scids))

		i := 0
		assets := []string{}
		for k := range scids {
			if !rpc.Wallet.IsConnected() || !Gnomes.IsRunning() {
				break
			}

			keys[i] = k
			if header, _ := Gnomes.GetSCIDValuesByKey(keys[i], "nameHdr"); header != nil {
				owner, _ := Gnomes.GetSCIDValuesByKey(keys[i], "owner")
				file, _ := Gnomes.GetSCIDValuesByKey(keys[i], "fileURL")
				if owner != nil && file != nil {
					if owner[0] == rpc.Wallet.Address && ValidNfa(file[0]) {
						assets = append(assets, header[0]+"   "+keys[i])
					}
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
	if Gnomes.IsReady() && !Gnomes.HasChecked() {
		if contracts == nil {
			contracts = Gnomes.GetAllOwnersAndSCIDs()
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
	if Gnomes.IsRunning() {
		headers, _ := Gnomes.GetSCIDValuesByKey(rpc.GnomonSCID, scid)

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
	if Gnomes.IsReady() {
		if dev, _ := Gnomes.GetSCIDValuesByKey(scid, "dev"); dev != nil {
			owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner")
			_, init := Gnomes.GetSCIDValuesByKey(scid, t+"_init")

			if owner != nil && init != nil {
				if dev[0] == rpc.DevAddress && !rpc.Wallet.BetOwner {
					SetBetOwner(owner[0])
				}
			}
		}
	}
}

// Verify if wallet is a co owner on bet contract
func VerifyBetSigner(scid string) bool {
	if Gnomes.IsReady() {
		for i := 2; i < 10; i++ {
			if !Gnomes.IsRunning() {
				break
			}

			signer_addr, _ := Gnomes.GetSCIDValuesByKey(scid, "co_signer"+strconv.Itoa(i))
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
	if Gnomes.IsReady() {
		if dev, _ := Gnomes.GetSCIDValuesByKey(scid, "dev"); dev != nil {
			owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner")
			_, init := Gnomes.GetSCIDValuesByKey(scid, t+"_init")

			if owner != nil && init != nil {
				if dev[0] == rpc.DevAddress {
					headers := GetSCHeaders(scid)
					name := "?"
					desc := "?"
					var hidden bool
					_, restrict := Gnomes.GetSCIDValuesByKey(rpc.RatingSCID, "restrict")
					_, rating := Gnomes.GetSCIDValuesByKey(rpc.RatingSCID, scid)

					if restrict != nil && rating != nil {
						Control.Lock()
						Control.Contract_rating[scid] = rating[0]
						Control.Unlock()
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
	}

	return list, owned
}

// Populate all dReams dPrediction contracts
//   - Pass contracts from db store, can be nil arg
func PopulatePredictions(contracts map[string]string) {
	if rpc.Daemon.IsConnected() && Gnomes.IsReady() {
		list := []string{}
		owned := []string{}
		if contracts == nil {
			contracts = Gnomes.GetAllOwnersAndSCIDs()
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
	if rpc.Daemon.IsConnected() && Gnomes.IsReady() {
		list := []string{}
		owned := []string{}
		if contracts == nil {
			contracts = Gnomes.GetAllOwnersAndSCIDs()
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
	if Gnomes.IsReady() {
		artAddr, _ := Gnomes.GetSCIDValuesByKey(scid, "artificerAddr")
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

// Get SCID info and update Asset content
func GetOwnedAssetStats(scid string) {
	if Gnomes.IsReady() {
		n, _ := Gnomes.GetSCIDValuesByKey(scid, "nameHdr")
		if n != nil {
			c, _ := Gnomes.GetSCIDValuesByKey(scid, "collection")
			//d, _ := Gnomes.GetSCIDValuesByKey(scid, "descrHdr:")
			i, _ := Gnomes.GetSCIDValuesByKey(scid, "iconURLHdr")

			if n != nil {
				Assets.Name.Text = (" Name: " + n[0])
				Assets.Name.Refresh()
				if !Control.List_open && !Control.send_open {
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
					a, _ = Gnomes.GetSCIDValuesByKey(scid, "fileURL")
				}
			} else {
				Assets.Collection.Text = (" Collection: ?")
				Assets.Collection.Refresh()
			}

			if i != nil {
				if a != nil {
					Assets.Icon, _ = dreams.DownloadFile(a[0], n[0])
				} else {
					Assets.Icon, _ = dreams.DownloadFile(i[0], n[0])
				}
			} else {
				Assets.Icon = *canvas.NewImageFromImage(nil)
			}

		} else {
			Control.List_button.Hide()
			data, _ := Gnomes.GetSCIDValuesByKey(scid, "metadata")
			minter, _ := Gnomes.GetSCIDValuesByKey(scid, "minter")
			coll, _ := Gnomes.GetSCIDValuesByKey(scid, "collection")
			if data != nil && minter != nil && coll != nil {
				if minter[0] == dreams.Seals_mint && coll[0] == dreams.Seals_coll {
					var seal dreams.Seal
					if err := json.Unmarshal([]byte(data[0]), &seal); err == nil {
						check := strings.Trim(seal.Name, " #0123456789")
						if check == "Dero Seals" {
							Assets.Name.Text = (" Name: " + seal.Name)
							Assets.Name.Refresh()

							Assets.Collection.Text = (" Collection: " + check)
							Assets.Collection.Refresh()

							number := strings.Trim(seal.Name, "DeroSals# ")
							Assets.Icon, _ = dreams.DownloadFile("https://ipfs.io/ipfs/QmP3HnzWpiaBA6ZE8c3dy5ExeG7hnYjSqkNfVbeVW5iEp6/low/"+number+".jpg", seal.Name)
						}
					}
				} else if minter[0] == dreams.ATeam_mint && coll[0] == dreams.ATeam_coll {
					var agent dreams.Agent
					if err := json.Unmarshal([]byte(data[0]), &agent); err == nil {
						Assets.Name.Text = (" Name: " + agent.Name)
						Assets.Name.Refresh()

						Assets.Collection.Text = (" Collection: Dero A-Team")
						Assets.Collection.Refresh()

						number := strconv.Itoa(agent.ID)
						if agent.ID < 172 {
							Assets.Icon, _ = dreams.DownloadFile("https://ipfs.io/ipfs/QmaRHXcQwbFdUAvwbjgpDtr5kwGiNpkCM2eDBzAbvhD7wh/low/"+number+".jpg", agent.Name)
						} else {
							Assets.Icon, _ = dreams.DownloadFile("https://ipfs.io/ipfs/QmQQyKoE9qDnzybeDCXhyMhwQcPmLaVy3AyYAzzC2zMauW/low/"+number+".jpg", agent.Name)
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
		link, _ = Gnomes.GetSCIDValuesByKey(scid, "fileURL")
	case 1:
		link, _ = Gnomes.GetSCIDValuesByKey(scid, "iconURLHdr")
	case 2:
		link, _ = Gnomes.GetSCIDValuesByKey(scid, "coverURL")
	default:
		// nothing
	}

	if link != nil {
		url = link[0]
	}

	return
}

// Check owner of any SCID using "owner" key
func CheckOwner(scid string) bool {
	if len(scid) != 64 || !Gnomes.IsReady() {
		return false
	}

	owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner")
	if owner != nil {
		return owner[0] == rpc.Wallet.Address
	}

	return false
}

// Get a wallets registered names
func CheckWalletNames(value string) {
	if Gnomes.IsReady() {
		names, _ := Gnomes.GetSCIDKeysByValue(rpc.NameSCID, value)

		sort.Strings(names)
		Control.Names.Options = append(Control.Names.Options, names...)
	}
}

// Check if dPrediction is live on SCID
func CheckActivePrediction(scid string) bool {
	if len(scid) == 64 && Gnomes.IsReady() {
		_, ends := Gnomes.GetSCIDValuesByKey(scid, "p_end_at")
		_, buff := Gnomes.GetSCIDValuesByKey(scid, "buffer")
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
	if Gnomes.IsReady() {
		_, played := Gnomes.GetSCIDValuesByKey(scid, "s_played")
		_, init := Gnomes.GetSCIDValuesByKey(scid, "s_init")

		if played != nil && init != nil {
			return played[0] == init[0]
		}
	}

	return true
}

func GetSportsAmt(scid, n string) uint64 {
	_, amt := Gnomes.GetSCIDValuesByKey(scid, "s_amount_"+n)
	if amt != nil {
		return amt[0]
	} else {
		return 0
	}
}

// Get current dSports game teams
func GetSportsTeams(scid, n string) (string, string) {
	game, _ := Gnomes.GetSCIDValuesByKey(scid, "game_"+n)

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

// Trim input string to specified len
func TrimStringLen(str string, l int) string {
	if len(str) > l {
		return str[0:l]
	}

	return str
}

// Scan index for any active NFA listings
//   - Pass assets from db store, can be nil arg
func FindNfaListings(assets map[string]string) {
	if Gnomes.IsReady() {
		auction := []string{" Collection,  Name,  Description,  SCID:"}
		buy_now := []string{" Collection,  Name,  Description,  SCID:"}
		my_list := []string{" Collection,  Name,  Description,  SCID:"}
		if assets == nil {
			assets = Gnomes.GetAllOwnersAndSCIDs()
		}
		keys := make([]string, len(assets))

		i := 0
		for k := range assets {
			if !Gnomes.IsRunning() {
				return
			}

			keys[i] = k

			a, owned, expired := checkNfaAuctionListing(keys[i])

			if a != "" && !expired {
				auction = append(auction, a)
			}

			if owned {
				my_list = append(my_list, a)
			}

			b, owned, expired := checkNfaBuyListing(keys[i])

			if b != "" && !expired {
				buy_now = append(buy_now, b)
			}

			if owned {
				my_list = append(my_list, b)
			}

			i++
		}

		if !Gnomes.IsRunning() {
			return
		}

		Market.Auctions = auction
		Market.Buy_now = buy_now
		Market.My_list = my_list
		sort.Strings(Market.Auctions)
		sort.Strings(Market.Buy_now)
		sort.Strings(Market.My_list)

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

// Check NFA listing type and return owner address
//   - Auction returns 1
//   - Sale returns 2
func CheckNFAListingType(scid string) (list int, addr string) {
	if Gnomes.IsReady() {
		if owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner"); owner != nil {
			if listType, _ := Gnomes.GetSCIDValuesByKey(scid, "listType"); listType != nil {
				addr = owner[0]
				switch listType[0] {
				case "auction":
					list = 1
				case "sale":
					list = 2
				default:

				}
			}
		}
	}
	return
}

// Check if NFA SCID is listed for auction
//   - Market.DreamsFilter false for all NFA listings
func checkNfaAuctionListing(scid string) (asset string, owned, expired bool) {
	if Gnomes.IsReady() {
		if creator, _ := Gnomes.GetSCIDValuesByKey(scid, "creatorAddr"); creator != nil {
			listType, _ := Gnomes.GetSCIDValuesByKey(scid, "listType")
			header, _ := Gnomes.GetSCIDValuesByKey(scid, "nameHdr")
			coll, _ := Gnomes.GetSCIDValuesByKey(scid, "collection")
			desc, _ := Gnomes.GetSCIDValuesByKey(scid, "descrHdr")
			if listType != nil && header != nil && coll != nil && desc != nil {
				if Market.DreamsFilter {
					check := strings.Trim(header[0], "0123456789")
					if isDreamsNfaCollection(check) {
						if listType[0] == "auction" {
							desc_check := TrimStringLen(desc[0], 66)
							asset = coll[0] + "   " + header[0] + "   " + desc_check + "   " + scid
							if owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner"); owner != nil {
								if owner[0] == rpc.Wallet.Address {
									owned = true
								}
							}

							if _, endTime := Gnomes.GetSCIDValuesByKey(scid, "endBlockTime"); endTime != nil {
								now := uint64(time.Now().Unix())
								if now > endTime[0] && endTime[0] > 0 {
									expired = true
								}
							}
						}
					}
				} else {
					var hidden bool
					for _, addr := range Market.Filters {
						if creator[0] == addr {
							hidden = true
						}
					}

					if !hidden {
						if listType[0] == "auction" {
							desc_check := TrimStringLen(desc[0], 66)
							asset = coll[0] + "   " + header[0] + "   " + desc_check + "   " + scid
							if owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner"); owner != nil {
								if owner[0] == rpc.Wallet.Address {
									owned = true
								}
							}

							if _, endTime := Gnomes.GetSCIDValuesByKey(scid, "endBlockTime"); endTime != nil {
								now := uint64(time.Now().Unix())
								if now > endTime[0] && endTime[0] > 0 {
									expired = true
								}
							}
						}
					}
				}
			}
		}
	}

	return
}

// Check if NFA SCID is listed as buy now
//   - Market.DreamsFilter false for all NFA listings
func checkNfaBuyListing(scid string) (asset string, owned, expired bool) {
	if Gnomes.IsReady() {
		if creator, _ := Gnomes.GetSCIDValuesByKey(scid, "creatorAddr"); creator != nil {
			listType, _ := Gnomes.GetSCIDValuesByKey(scid, "listType")
			header, _ := Gnomes.GetSCIDValuesByKey(scid, "nameHdr")
			coll, _ := Gnomes.GetSCIDValuesByKey(scid, "collection")
			desc, _ := Gnomes.GetSCIDValuesByKey(scid, "descrHdr")
			if listType != nil && header != nil && coll != nil && desc != nil {
				if Market.DreamsFilter {
					check := strings.Trim(header[0], "0123456789")
					if isDreamsNfaCollection(check) {
						if listType[0] == "sale" {
							desc_check := TrimStringLen(desc[0], 66)
							asset = coll[0] + "   " + header[0] + "   " + desc_check + "   " + scid
							if owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner"); owner != nil {
								if owner[0] == rpc.Wallet.Address {
									owned = true
								}
							}

							if _, endTime := Gnomes.GetSCIDValuesByKey(scid, "endBlockTime"); endTime != nil {
								now := uint64(time.Now().Unix())
								if now > endTime[0] && endTime[0] > 0 {
									expired = true
								}
							}
						}
					}
				} else {
					var hidden bool
					for _, addr := range Market.Filters {
						if creator[0] == addr {
							hidden = true
						}
					}

					if !hidden {
						if listType[0] == "sale" {
							desc_check := TrimStringLen(desc[0], 66)
							asset = coll[0] + "   " + header[0] + "   " + desc_check + "   " + scid
							if owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner"); owner != nil {
								if owner[0] == rpc.Wallet.Address {
									owned = true
								}
							}

							if _, endTime := Gnomes.GetSCIDValuesByKey(scid, "endBlockTime"); endTime != nil {
								now := uint64(time.Now().Unix())
								if now > endTime[0] && endTime[0] > 0 {
									expired = true
								}
							}
						}
					}
				}
			}
		}
	}

	return
}

// Search NFAs in index by name or collection
func SearchNFAsBy(by int, prefix string) (results []string) {
	if Gnomes.IsReady() {
		results = []string{" Collection,  Name,  Description,  SCID:"}
		assets := Gnomes.GetAllOwnersAndSCIDs()
		keys := make([]string, len(assets))

		i := 0
		for k := range assets {
			if !Gnomes.IsReady() {
				return
			}

			keys[i] = k

			if file, _ := Gnomes.GetSCIDValuesByKey(keys[i], "fileURL"); file != nil {
				if ValidNfa(file[0]) {
					if name, _ := Gnomes.GetSCIDValuesByKey(keys[i], "nameHdr"); name != nil {
						coll, _ := Gnomes.GetSCIDValuesByKey(keys[i], "collection")
						desc, _ := Gnomes.GetSCIDValuesByKey(keys[i], "descrHdr")
						if coll != nil && desc != nil {
							switch by {
							case 0:
								if strings.HasPrefix(coll[0], prefix) {
									desc_check := TrimStringLen(desc[0], 66)
									asset := coll[0] + "   " + name[0] + "   " + desc_check + "   " + keys[i]
									results = append(results, asset)
								}
							case 1:
								if strings.HasPrefix(name[0], prefix) {
									desc_check := TrimStringLen(desc[0], 66)
									asset := coll[0] + "   " + name[0] + "   " + desc_check + "   " + keys[i]
									results = append(results, asset)
								}
							}
						}
					}
				}
			}

			i++
		}

		sort.Strings(results)
	}

	return
}

// Get NFA image files
func GetNfaImages(scid string) {
	if Gnomes.IsReady() && len(scid) == 64 {
		name, _ := Gnomes.GetSCIDValuesByKey(scid, "nameHdr")
		icon, _ := Gnomes.GetSCIDValuesByKey(scid, "iconURLHdr")
		cover, _ := Gnomes.GetSCIDValuesByKey(scid, "coverURL")
		if icon != nil {
			Market.Icon, _ = dreams.DownloadFile(icon[0], name[0])
			Market.Cover, _ = dreams.DownloadFile(cover[0], name[0]+"-cover")
		} else {
			Market.Icon = *canvas.NewImageFromImage(nil)
			Market.Cover = *canvas.NewImageFromImage(nil)
		}
	}
}

// Create auction tab info for current asset
func GetAuctionDetails(scid string) {
	if Gnomes.IsReady() && len(scid) == 64 {
		name, _ := Gnomes.GetSCIDValuesByKey(scid, "nameHdr")
		collection, _ := Gnomes.GetSCIDValuesByKey(scid, "collection")
		description, _ := Gnomes.GetSCIDValuesByKey(scid, "descrHdr")
		creator, _ := Gnomes.GetSCIDValuesByKey(scid, "creatorAddr")
		owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner")
		typeHdr, _ := Gnomes.GetSCIDValuesByKey(scid, "typeHdr")
		_, owner_update := Gnomes.GetSCIDValuesByKey(scid, "ownerCanUpdate")
		_, start := Gnomes.GetSCIDValuesByKey(scid, "startPrice")
		_, current := Gnomes.GetSCIDValuesByKey(scid, "currBidAmt")
		_, bid_price := Gnomes.GetSCIDValuesByKey(scid, "currBidPrice")
		_, royalty := Gnomes.GetSCIDValuesByKey(scid, "royalty")
		_, bids := Gnomes.GetSCIDValuesByKey(scid, "bidCount")
		_, endTime := Gnomes.GetSCIDValuesByKey(scid, "endBlockTime")
		_, startTime := Gnomes.GetSCIDValuesByKey(scid, "startBlockTime")
		_, artFee := Gnomes.GetSCIDValuesByKey(scid, "artificerFee")

		if name != nil && collection != nil && start != nil && royalty != nil && endTime != nil && artFee != nil && typeHdr != nil {
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
					ty = typeHdr[0]
				}

				Market.Viewing_coll = check

				Market.Name.SetText(name[0])

				Market.Type.SetText(ty)

				Market.Collection.SetText(collection[0])

				Market.Description.SetText(description[0])

				if Market.Creator.Text != creator[0] {
					Market.Creator.SetText(creator[0])
				}

				if Market.Owner.Text != owner[0] {
					Market.Owner.SetText(owner[0])
				}
				if owner_update[0] == 1 {
					Market.Owner_update.SetText("Yes")
				} else {
					Market.Owner_update.SetText("No")
				}

				Market.Art_fee.SetText(strconv.Itoa(int(artFee[0])) + "%")

				Market.Royalty.SetText(strconv.Itoa(int(royalty[0])) + "%")

				price := float64(start[0])
				str := fmt.Sprintf("%.5f", price/100000)
				Market.Start_price.SetText(str + " Dero")

				Market.Bid_count.SetText(strconv.Itoa(int(bids[0])))

				end, _ := rpc.MsToTime(strconv.Itoa(int(endTime[0]) * 1000))
				Market.End_time.SetText(end.String())

				if current != nil {
					value := float64(current[0])
					str := fmt.Sprintf("%.5f", value/100000)
					Market.Current_bid.SetText(str)
				} else {
					Market.Current_bid.SetText("")
				}

				if bid_price != nil {
					value := float64(bid_price[0])
					str := fmt.Sprintf("%.5f", value/100000)
					if bid_price[0] == 0 {
						Market.Bid_amt = start[0]
					} else {
						Market.Bid_amt = bid_price[0]
					}
					Market.Bid_price.SetText(str)
				} else {
					Market.Bid_amt = 0
					Market.Bid_price.SetText("")
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
	if Gnomes.IsReady() && len(scid) == 64 {
		name, _ := Gnomes.GetSCIDValuesByKey(scid, "nameHdr")
		collection, _ := Gnomes.GetSCIDValuesByKey(scid, "collection")
		description, _ := Gnomes.GetSCIDValuesByKey(scid, "descrHdr")
		creator, _ := Gnomes.GetSCIDValuesByKey(scid, "creatorAddr")
		owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner")
		typeHdr, _ := Gnomes.GetSCIDValuesByKey(scid, "typeHdr")
		_, owner_update := Gnomes.GetSCIDValuesByKey(scid, "ownerCanUpdate")
		_, start := Gnomes.GetSCIDValuesByKey(scid, "startPrice")
		_, royalty := Gnomes.GetSCIDValuesByKey(scid, "royalty")
		_, endTime := Gnomes.GetSCIDValuesByKey(scid, "endBlockTime")
		_, startTime := Gnomes.GetSCIDValuesByKey(scid, "startBlockTime")
		_, artFee := Gnomes.GetSCIDValuesByKey(scid, "artificerFee")

		if name != nil && collection != nil && start != nil && royalty != nil && endTime != nil && artFee != nil && typeHdr != nil {
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
					ty = typeHdr[0]
				}

				Market.Viewing_coll = check

				Market.Name.SetText(name[0])

				Market.Type.SetText(ty)

				Market.Collection.SetText(collection[0])

				Market.Description.SetText(description[0])

				if Market.Creator.Text != creator[0] {
					Market.Creator.SetText(creator[0])
				}

				if Market.Owner.Text != owner[0] {
					Market.Owner.SetText(owner[0])
				}

				if owner_update[0] == 1 {
					Market.Owner_update.SetText("Yes")
				} else {
					Market.Owner_update.SetText("No")
				}

				Market.Art_fee.SetText(strconv.Itoa(int(artFee[0])) + "%")

				Market.Royalty.SetText(strconv.Itoa(int(royalty[0])) + "%")

				Market.Buy_amt = start[0]
				value := float64(start[0])
				str := fmt.Sprintf("%.5f", value/100000)
				Market.Start_price.SetText(str + " Dero")

				Market.Entry.SetText(str)
				Market.Entry.Disable()
				end, _ := rpc.MsToTime(strconv.Itoa(int(endTime[0]) * 1000))
				Market.End_time.SetText(end.String())

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

// Create info for unlisted NFA
func GetUnlistedDetails(scid string) {
	if Gnomes.IsReady() && len(scid) == 64 {
		name, _ := Gnomes.GetSCIDValuesByKey(scid, "nameHdr")
		collection, _ := Gnomes.GetSCIDValuesByKey(scid, "collection")
		description, _ := Gnomes.GetSCIDValuesByKey(scid, "descrHdr")
		creator, _ := Gnomes.GetSCIDValuesByKey(scid, "creatorAddr")
		owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner")
		typeHdr, _ := Gnomes.GetSCIDValuesByKey(scid, "typeHdr")
		_, owner_update := Gnomes.GetSCIDValuesByKey(scid, "ownerCanUpdate")
		_, start := Gnomes.GetSCIDValuesByKey(scid, "startPrice")
		_, royalty := Gnomes.GetSCIDValuesByKey(scid, "royalty")
		_, endTime := Gnomes.GetSCIDValuesByKey(scid, "endBlockTime")
		_, startTime := Gnomes.GetSCIDValuesByKey(scid, "startBlockTime")
		_, artFee := Gnomes.GetSCIDValuesByKey(scid, "artificerFee")

		if name != nil && collection != nil && start != nil && royalty != nil && endTime != nil && artFee != nil && typeHdr != nil {
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
					ty = typeHdr[0]
				}

				Market.Viewing_coll = check

				Market.Name.SetText(name[0])

				Market.Type.SetText(ty)

				Market.Collection.SetText(collection[0])

				Market.Description.SetText(description[0])

				if Market.Creator.Text != creator[0] {
					Market.Creator.SetText(creator[0])
				}

				if Market.Owner.Text != owner[0] {
					Market.Owner.SetText(owner[0])
				}

				if owner_update[0] == 1 {
					Market.Owner_update.SetText("Yes")
				} else {
					Market.Owner_update.SetText("No")
				}

				Market.Art_fee.SetText(strconv.Itoa(int(artFee[0])) + "%")

				Market.Royalty.SetText(strconv.Itoa(int(royalty[0])) + "%")

				Market.Entry.SetText("0")
				Market.Entry.Disable()

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
			}()
		}
	}
}

// Get percentages for a NFA
func GetListingPercents(scid string) (artP float64, royaltyP float64) {
	if Gnomes.IsReady() {
		_, artFee := Gnomes.GetSCIDValuesByKey(scid, "artificerFee")
		_, royalty := Gnomes.GetSCIDValuesByKey(scid, "royalty")

		if artFee != nil && royalty != nil {
			artP = float64(artFee[0]) / 100
			royaltyP = float64(royalty[0]) / 100

			return
		}
	}

	return
}
