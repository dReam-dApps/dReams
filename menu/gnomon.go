package menu

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/table"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
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
	Init     bool
	Sync     bool
	Checked  bool
	SCIDS    uint64
	Indexer  *indexer.Indexer
	Graviton *storage.GravitonStore
}

var Gnomes gnomon

func stringToInt64(s string) int64 {
	if s != "" {
		i, err := strconv.Atoi(s)
		if err != nil {
			log.Println("String Conversion Error", err)
			return 0
		}
		return int64(i)
	}

	return 0
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

	filter = append(filter, nfa_search_filter)
	filter = append(filter, g45_search_filter)

	return filter
}

func manualIndex(scid []string) {
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
}

func GnomonDB() *storage.GravitonStore {
	shasum := fmt.Sprintf("%x", sha1.Sum([]byte("gnomon")))
	db_folder := fmt.Sprintf("gnomondb\\%s_%s", "GNOMON", shasum)
	db := storage.NewGravDB(db_folder, "25ms")

	return db
}

func startGnomon(ep string) {
	log.Println("Starting Gnomon.")
	Gnomes.Graviton = GnomonDB()

	last_indexedheight := Gnomes.Graviton.GetLastIndexHeight()
	daemon_endpoint := ep
	runmode := "daemon"
	mbl := false
	closeondisconnect := false
	fastsync := true

	filters := searchFilters()
	if len(filters) == 7 {
		Gnomes.Indexer = indexer.NewIndexer(Gnomes.Graviton, filters, last_indexedheight, daemon_endpoint, runmode, mbl, closeondisconnect, fastsync)
		go Gnomes.Indexer.StartDaemonMode()
		Gnomes.Init = true
	}
	time.Sleep(3 * time.Second)
}

func StopGnomon(gi bool) {
	if gi {
		log.Println("Putting Gnomon to sleep.")
		Gnomes.Graviton.StoreLastIndexHeight(Gnomes.Graviton.GetLastIndexHeight())
		Gnomes.Indexer.Close()
		Gnomes.Graviton.DB.Close()
		Gnomes.Init = false
		log.Println("Gnomon is asleep.")
		time.Sleep(1 * time.Second)
	}
}

func GnomonState(dc bool) {
	if dc && Gnomes.Init && !Gnomes.Indexer.Closing {
		Gnomes.Graviton.StoreLastIndexHeight(Gnomes.Indexer.LastIndexedHeight)
		Gnomes.SCIDS = uint64(len(Gnomes.Graviton.GetAllOwnersAndSCIDs()))

		if Gnomes.Graviton.GetLastIndexHeight() >= stringToInt64(rpc.Wallet.Height) && stringToInt64(rpc.Wallet.Height) != 0 && !Gnomes.Indexer.Closing {
			Gnomes.Sync = true
			if rpc.Wallet.Connect {
				go CheckBetContract(Gnomes.Sync, Gnomes.Checked)
				CreateTableList(Gnomes.Checked)
				go CheckG45owner(Gnomes.Sync, Gnomes.Checked)
				CheckAssets(Gnomes.Sync, Gnomes.Checked)

			}
		} else {
			Gnomes.Sync = false
		}

		table.Assets.Stats_box = *container.NewVBox(table.Assets.Collection, table.Assets.Name, table.IconImg(Resource.Frame))
		table.Assets.Stats_box.Refresh()
		PlayerControl.Stats_box = *container.NewVBox(Stats.Name, Stats.Desc, Stats.Version, Stats.Last, Stats.Seats, TableIcon(Resource.Frame))
		PlayerControl.Stats_box.Refresh()

		if len(Market.Viewing) == 64 && rpc.Wallet.Connect {
			if Market.Tab == "Buy" {
				GetBuyNowDetails(Market.Viewing)
				BuyNowInfo()
			} else {
				GetAuctionDetails(Market.Viewing)
				AuctionInfo()
			}
		}
	}
}

func searchIndex(scid string) {
	if len(scid) == 64 {
		var found bool
		all := Gnomes.Graviton.GetAllOwnersAndSCIDs()
		for sc := range all {
			if scid == sc {
				log.Println(scid + " Indxed")
				found = true
			}
		}
		if !found {
			log.Println(scid + " Not Found")
		}
	}
}

func CheckAssets(gs, gc bool) {
	if gs && !gc {
		assets := Gnomes.Graviton.GetAllOwnersAndSCIDs()
		keys := make([]string, len(assets))
		log.Println("Checking NFA Assets")

		i := 0
		for k := range assets {
			keys[i] = k
			checkNFAOwner(keys[i])
			i++
		}
		Gnomes.Checked = true
		table.Assets.Asset_list.Refresh()
	}
	sort.Strings(table.Assets.Assets)
}

func CheckBetContract(gs, gc bool) {
	if gs && !gc {
		contracts := Gnomes.Graviton.GetAllOwnersAndSCIDs()
		keys := make([]string, len(contracts))

		i := 0
		for k := range contracts {
			keys[i] = k
			verifyBetContract(keys[i], "p")
			verifyBetContract(keys[i], "s")
			i++
		}
	}
}

func verifyBetContract(scid, t string) {
	owner, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "owner", Gnomes.Indexer.ChainHeight, true)
	dev, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "dev", Gnomes.Indexer.ChainHeight, true)
	_, init := Gnomes.Graviton.GetSCIDValuesByKey(scid, t+"_init", Gnomes.Indexer.ChainHeight, true)

	if owner != nil && dev != nil && init != nil {
		if dev[0] == rpc.DevAddress {
			DisableBetOwner(owner[0])
		}
	}
}

func isNfa(scid string) bool {
	artAddr, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "artificerAddr", Gnomes.Indexer.ChainHeight, true)
	if artAddr != nil {
		return artAddr[0] == rpc.ArtAddress
	}
	return false
}

func validNfa(file string) bool {
	return file != "-"
}

func checkNFAOwner(scid string) {
	owner, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "owner", Gnomes.Indexer.ChainHeight, true)
	header, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.ChainHeight, true)
	file, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "fileURL", Gnomes.Indexer.ChainHeight, true)
	if owner != nil && header != nil && file != nil {
		if owner[0] == rpc.Wallet.Address && validNfa(file[0]) {
			check := strings.Trim(header[0], "0123456789")
			if check == "AZYDS" {
				current := table.Settings.ThemeSelect.Options
				new := append(current, header[0])
				table.Settings.ThemeSelect.Options = new
				table.Settings.ThemeSelect.Refresh()
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

func GetOwnedAssetStats(scid string) {
	n, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.LastIndexedHeight, true)
	if n != nil {
		c, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "collection", Gnomes.Indexer.LastIndexedHeight, true)
		//d, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "descrHdr:", Gnomes.Indexer.LastIndexedHeight, true)
		i, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "iconURLHdr", Gnomes.Indexer.LastIndexedHeight, true)

		if n != nil {
			table.Assets.Name.Text = (" Name: " + n[0])
			table.Assets.Name.Refresh()
			PlayerControl.List_button.Show()

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
		PlayerControl.List_button.Hide()
		data, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "metadata", Gnomes.Indexer.LastIndexedHeight, true)
		if data != nil {
			var seal Seal
			json.Unmarshal([]byte(data[0]), &seal)
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
	}
}

func checkTableOwner(scid string) bool {
	if len(scid) != 64 {
		return false
	}

	check := strings.Trim(scid, " 0123456789")
	if check == "Holdero Tables:" {
		return false
	}

	owner, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "owner:", Gnomes.Indexer.LastIndexedHeight, true)
	return owner[0] == rpc.Wallet.Address
}

func checkTableVersion(scid string) uint64 {
	_, v := Gnomes.Graviton.GetSCIDValuesByKey(scid, "V:", Gnomes.Indexer.LastIndexedHeight, true)

	if v != nil && v[0] >= 100 {
		return v[0]
	}
	return 0
}

func CreateTableList(gc bool) {
	if !gc {
		var owner bool
		TableList = []string{}
		tables := Gnomes.Graviton.GetAllOwnersAndSCIDs()

		for scid := range tables {
			_, valid := Gnomes.Graviton.GetSCIDValuesByKey(scid, "Deck Count:", Gnomes.Indexer.LastIndexedHeight, true)
			_, version := Gnomes.Graviton.GetSCIDValuesByKey(scid, "V:", Gnomes.Indexer.LastIndexedHeight, true)

			if valid != nil && version != nil {
				d := valid[0]
				v := version[0]
				if d >= 1 && v >= 100 {
					TableList = append(TableList, scid)
					if checkTableOwner(scid) {
						PlayerControl.holdero_unlock.Hide()
						PlayerControl.holdero_new.Show()
						owner = true
						rpc.Wallet.PokerOwner = true
					}
				}
			}
		}

		if !owner {
			PlayerControl.holdero_unlock.Show()
			PlayerControl.holdero_new.Hide()
			rpc.Wallet.PokerOwner = false
		}

		t := len(TableList)
		TableList = append(TableList, "  Holdero Tables: "+strconv.Itoa(t))
		sort.Strings(TableList)

		PlayerControl.table_options.Refresh()
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
		_, v := Gnomes.Graviton.GetSCIDValuesByKey(scid, "V:", Gnomes.Indexer.LastIndexedHeight, true)
		_, l := Gnomes.Graviton.GetSCIDValuesByKey(scid, "Last", Gnomes.Indexer.LastIndexedHeight, true)
		_, s := Gnomes.Graviton.GetSCIDValuesByKey(scid, "Seats at Table:", Gnomes.Indexer.LastIndexedHeight, true)
		// _, o := Gnomes.Graviton.GetSCIDValuesByKey(scid, "Open", Gnomes.Indexer.LastIndexedHeight, true)
		// p1, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "Player 1 ID:", Gnomes.Indexer.LastIndexedHeight, true)
		p2, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "Player2 ID:", Gnomes.Indexer.LastIndexedHeight, true)
		p3, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "Player3 ID:", Gnomes.Indexer.LastIndexedHeight, true)
		p4, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "Player4 ID:", Gnomes.Indexer.LastIndexedHeight, true)
		p5, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "Player5 ID:", Gnomes.Indexer.LastIndexedHeight, true)
		p6, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "Player6 ID:", Gnomes.Indexer.LastIndexedHeight, true)
		h, _ := rpc.GetSCHeaders(scid)

		if single {
			if h != nil {
				Stats.Name.Text = (" Name: " + h[0])
				Stats.Name.Refresh()
				Stats.Desc.Text = (" Description: " + h[1])
				Stats.Desc.Refresh()
				Stats.Image, _ = table.DownloadFile(h[2], h[0])

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

type Seal struct {
	Attributes struct {
		Eyes        string `json:"Eyes"`
		FacialHair  string `json:"Facial Hair"`
		HairAndHats string `json:"Hair And Hats"`
		Shirts      string `json:"Shirts"`
	} `json:"attributes"`
	ID    int     `json:"id"`
	Image string  `json:"image"`
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

func CheckG45owner(gs, gc bool) {
	if gs && !gc {
		g45s := Gnomes.Graviton.GetAllOwnersAndSCIDs()
		log.Println("Checking G45 Assets")

		for scid := range g45s {
			var seal Seal
			data, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "metadata", Gnomes.Indexer.LastIndexedHeight, true)
			owner, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "owner", Gnomes.Indexer.LastIndexedHeight, true)
			if data != nil && owner != nil {
				if owner[0] == rpc.Wallet.Address {
					json.Unmarshal([]byte(data[0]), &seal)
					check := strings.Trim(seal.Name, " #0123456789")
					if check == "Dero Seals" {
						table.Assets.Assets = append(table.Assets.Assets, seal.Name+"   "+scid)
						current := table.Settings.AvatarSelect.Options
						new := append(current, seal.Name)
						table.Settings.AvatarSelect.Options = new
						table.Settings.AvatarSelect.Refresh()
					}
				}
			}
		}
		table.Assets.Asset_list.Refresh()
	}
}

func CheckPredictionName(scid string) (name string) {
	if scid != "" {
		check := Gnomes.Graviton.GetAllSCIDVariableDetails(scid)
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
						value, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, v, Gnomes.Indexer.ChainHeight, true)
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

func GetSportsAmt(scid, n string) uint64 {
	_, amt := Gnomes.Graviton.GetSCIDValuesByKey(scid, "s_amount_"+n, Gnomes.Indexer.ChainHeight, true)
	if amt != nil {
		return amt[0]
	} else {
		return 0
	}
}

func GetSportsTeams(scid, n string) (string, string) {
	game, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "game_"+n, Gnomes.Indexer.ChainHeight, true)

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
	split := strings.Split(s, "-")

	return split[0]
}

func TrimTeamB(s string) string {
	split := strings.Split(s, "-")

	return split[1]
}

func FindNfaListings(gs bool) {
	if gs {
		Market.Auctions = []string{" Collection,  Name,  Description,  SCID:"}
		Market.Buy_now = []string{" Collection,  Name,  Description, SCID:"}
		assets := Gnomes.Graviton.GetAllOwnersAndSCIDs()
		keys := make([]string, len(assets))

		i := 0
		for k := range assets {
			keys[i] = k
			checkNfaListing(keys[i])
			i++
		}

		sort.Strings(Market.Auctions)
		sort.Strings(Market.Buy_now)
	}
}

func checkNfaListing(scid string) {
	listType, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "listType", Gnomes.Indexer.ChainHeight, true)
	header, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.ChainHeight, true)
	coll, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "collection", Gnomes.Indexer.ChainHeight, true)
	desc, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "descrHdr", Gnomes.Indexer.ChainHeight, true)
	if listType != nil && header != nil {
		check := strings.Trim(header[0], "0123456789")
		if check == "AZYDS" || check == "DBC" || check == "AZYPC" || check == "SIXPC" || check == "AZYPCB" || check == "SIXPCB" {
			switch listType[0] {
			case "auction":
				Market.Auctions = append(Market.Auctions, coll[0]+"   "+header[0]+"   "+desc[0]+"   "+scid)
			case "sale":
				Market.Buy_now = append(Market.Buy_now, coll[0]+"   "+header[0]+"   "+desc[0]+"   "+scid)
			}
		}
	}
}

func GetAuctionImages(scid string) {
	if len(scid) == 64 {
		name, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.ChainHeight, true)
		icon, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "iconURLHdr", Gnomes.Indexer.LastIndexedHeight, true)
		cover, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "coverURL", Gnomes.Indexer.LastIndexedHeight, true)
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
		name, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.ChainHeight, true)
		collection, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "collection", Gnomes.Indexer.ChainHeight, true)
		description, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "descrHdr", Gnomes.Indexer.ChainHeight, true)
		creator, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "creatorAddr", Gnomes.Indexer.ChainHeight, true)
		owner, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "owner", Gnomes.Indexer.ChainHeight, true)
		_, owner_update := Gnomes.Graviton.GetSCIDValuesByKey(scid, "ownerCanUpdate", Gnomes.Indexer.ChainHeight, true)
		_, start := Gnomes.Graviton.GetSCIDValuesByKey(scid, "startPrice", Gnomes.Indexer.ChainHeight, true)
		_, current := Gnomes.Graviton.GetSCIDValuesByKey(scid, "currBidAmt", Gnomes.Indexer.ChainHeight, true)
		_, bid_price := Gnomes.Graviton.GetSCIDValuesByKey(scid, "currBidPrice", Gnomes.Indexer.ChainHeight, true)
		_, royalty := Gnomes.Graviton.GetSCIDValuesByKey(scid, "royalty", Gnomes.Indexer.ChainHeight, true)
		_, bids := Gnomes.Graviton.GetSCIDValuesByKey(scid, "bidCount", Gnomes.Indexer.ChainHeight, true)
		_, endTime := Gnomes.Graviton.GetSCIDValuesByKey(scid, "endBlockTime", Gnomes.Indexer.ChainHeight, true)
		_, startTime := Gnomes.Graviton.GetSCIDValuesByKey(scid, "startBlockTime", Gnomes.Indexer.ChainHeight, true)

		if name != nil {
			Market.Name.Text = (" Name: " + name[0])
			Market.Name.Refresh()
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

func GetBuyNowImages(scid string) {
	if len(scid) == 64 {
		name, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.ChainHeight, true)
		icon, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "iconURLHdr", Gnomes.Indexer.LastIndexedHeight, true)
		cover, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "coverURL", Gnomes.Indexer.LastIndexedHeight, true)
		if icon != nil {
			Market.Icon, _ = table.DownloadFile(icon[0], name[0])
			Market.Cover, _ = table.DownloadFile(cover[0], name[0]+"-cover")
		} else {
			Market.Icon = *canvas.NewImageFromImage(nil)
			Market.Cover = *canvas.NewImageFromImage(nil)

		}
	}
}

func GetBuyNowDetails(scid string) {
	if len(scid) == 64 {
		name, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "nameHdr", Gnomes.Indexer.ChainHeight, true)
		collection, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "collection", Gnomes.Indexer.ChainHeight, true)
		description, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "descrHdr", Gnomes.Indexer.ChainHeight, true)
		creator, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "creatorAddr", Gnomes.Indexer.ChainHeight, true)
		owner, _ := Gnomes.Graviton.GetSCIDValuesByKey(scid, "owner", Gnomes.Indexer.ChainHeight, true)
		_, owner_update := Gnomes.Graviton.GetSCIDValuesByKey(scid, "ownerCanUpdate", Gnomes.Indexer.ChainHeight, true)
		_, start := Gnomes.Graviton.GetSCIDValuesByKey(scid, "startPrice", Gnomes.Indexer.ChainHeight, true)
		_, royalty := Gnomes.Graviton.GetSCIDValuesByKey(scid, "royalty", Gnomes.Indexer.ChainHeight, true)
		_, endTime := Gnomes.Graviton.GetSCIDValuesByKey(scid, "endBlockTime", Gnomes.Indexer.ChainHeight, true)
		_, startTime := Gnomes.Graviton.GetSCIDValuesByKey(scid, "startBlockTime", Gnomes.Indexer.ChainHeight, true)

		if name != nil {
			Market.Name.Text = (" Name: " + name[0])
			Market.Name.Refresh()
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
