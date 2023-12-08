package menu

import (
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"github.com/civilware/Gnomon/structures"
	"github.com/dReam-dApps/dReams/rpc"
)

const (
	Seals_mint = "dero1qyfq8m3rju62tshju60zuc0ymrajwxqajkdh6pw888ejuv94jlfgjqq58px98"
	Seals_coll = "c6fa9a2c95d97da816eb9689a2fb52be385bb1df9e93abe99373ddbd3407129d"
	ATeam_mint = "dero1qyx9748k9wrt89a6rm0zzlayxgs3ndkmvg6m20shqp8ynh54zf2rgqq8yn9hn"
	ATeam_coll = "bbc357bdfe9fc41128fc11ce555eaadbd9b411eca903008396e0de4cc31821c7"
	Degen_coll = "8edea52b9a8a041e3b579ca2d81ea3d3e87e148ba4409273d53039991afa91be"
	Degen_mint = "dero1qy4e7jj4jaaj66pc0vg8h7l0hqelqjxj9ya9qgal03v0phjaycv5yqq8aqgyg"
)

type assetCount struct {
	name    string
	count   int
	creator string
}

// Dero Seals metadata struct
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

// Dero A-Team metadata struct
type Agent struct {
	Attributes struct {
		Color  string `json:"Color"`
		IChing string `json:"I-ching"`
	} `json:"attributes"`
	ID    int    `json:"id"`
	Image string `json:"image"`
	Name  string `json:"name"`
}

type Degen struct {
	Attributes struct {
	} `json:"attributes"`
	ID    int    `json:"id"`
	Image string `json:"image"`
	Name  string `json:"name"`
}

var dReamsG45s = []assetCount{
	{name: "Dero Seals", count: 3500},
	{name: "Dero A-Team", count: 300},
	{name: "Dero Degens", count: 2000},
}

// dReams G45 collections
func IsDreamsG45(check string) bool {
	for _, g := range dReamsG45s {
		if g.name == check {
			return true
		}
	}

	return false
}

// Returns collection SCID by name
func G45CollectionSC(name string) (collection string) {
	switch name {
	case "Dero Seals":
		return Seals_coll
	case "Dero A-Team":
		return ATeam_coll
	case "Dero Degens":
		return Degen_coll
	default:
		return
	}
}

// Returns search filter with all enabled G45s
func ReturnEnabledG45s(assets map[string]bool) (filter []string) {
	for name, enabled := range assets {
		if enabled && IsDreamsG45(name) {
			filter = append(filter, G45CollectionSC(name))
		}
	}

	return
}

// Manually add G45 collections to Gnomon index
func G45Index() {
	if Gnomes.DBType == "boltdb" {
		for Gnomes.IsWriting() {
			time.Sleep(time.Second)
		}
	}
	logger.Println("[dReams] Adding G45 Collections")
	filters := Gnomes.Indexer.SearchFilter
	Gnomes.Indexer.SearchFilter = []string{}
	scidstoadd := make(map[string]*structures.FastSyncImport)

	for _, c := range ReturnEnabledG45s(Control.Enabled_assets) {
		g45 := rpc.GetG45Collection(c)
		for _, sc := range g45 {
			scidstoadd[sc] = &structures.FastSyncImport{}
		}
	}

	err := Gnomes.Indexer.AddSCIDToIndex(scidstoadd, false, false)
	if err != nil {
		logger.Errorf("[G45Index] %v\n", err)
	}
	Gnomes.Indexer.SearchFilter = filters

	if runtime.GOOS != "windows" {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "dReams - Gnomon",
			Content: "Fast sync completed"})
	}
}
