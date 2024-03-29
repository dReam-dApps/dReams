package gnomes

import (
	"crypto/sha1"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/dReam-dApps/dReams/rpc"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"

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

	G45_search_filter = `STORE("type", "G45-NFT")`
)

// Headers from Gnomon SC
type SCHeaders struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IconURL     string `json:"iconURL"`
}

// SCID with Gnomon headers
type SC struct {
	ID      string
	Rating  uint64
	Version uint64
	Header  SCHeaders
}

var logger = structures.Logger.WithFields(logrus.Fields{})

// Enable escape codes for windows Stdout
func enableEscapeCodes() error {
	cmd := exec.Command("cmd", "/c", "echo", "ON")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Initialize logrus logger matching Gnomon log
func InitLogrusLog(level logrus.Level) {
	colors := true
	if runtime.GOOS == "windows" {
		if err := enableEscapeCodes(); err != nil {
			colors = false
			logger.Warnln("[InitLogrusLog] Err enabling escape codes:", err)
		}
	}

	structures.Logger = logrus.Logger{
		Out:   os.Stdout,
		Level: level,
		Formatter: &prefixed.TextFormatter{
			ForceColors:     colors,
			DisableColors:   !colors,
			TimestampFormat: "01/02/2006 15:04:05",
			FullTimestamp:   true,
			ForceFormatting: true,
		},
	}
}

// Manually add SCID(s) to Gnomon index
func AddToIndex(scids []string) (err error) {
	filters := gnomes.Indexer.SearchFilter
	gnomes.Indexer.SearchFilter = []string{}
	scidstoadd := make(map[string]*structures.FastSyncImport)

	for _, sc := range scids {
		owner, _ := gnomes.GetSCIDValuesByKey(rpc.GnomonSCID, sc+"owner")
		if owner != nil {
			scidstoadd[sc] = &structures.FastSyncImport{}
			scidstoadd[sc].Owner = owner[0]
		}
	}

	err = gnomes.Indexer.AddSCIDToIndex(scidstoadd, false, false)
	if err != nil {
		logger.Errorf("[AddToIndex] %v\n", err)
	}
	gnomes.Indexer.SearchFilter = filters

	return
}

// Create a new graviton DB for Gnomon storage
//   - If dbType is boltdb, will return nil gravdb
func NewGravDB(dbType, dbPath string) *storage.GravitonStore {
	if dbType == "boltdb" {
		return nil
	}

	db, err := storage.NewGravDB(dbPath, "25ms")
	if err != nil {
		logger.Fatalf("%s\n", err)
	}

	return db
}

// Create a new bbolt DB with dReams tag for Gnomon storage
//   - If dbType is not boltdb, will return nil boltdb
func NewBoltDB(dbType, dbPath string) *storage.BboltStore {
	if dbType != "boltdb" {
		return nil
	}

	db_name := fmt.Sprintf("gnomon_bolt_%s.db", "dReams")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dbPath, 0755); err != nil {
			logger.Fatalf("[NewBoltDB] %s\n", err)
		}
	}

	db, err := storage.NewBBoltDB(dbPath, db_name)
	if err != nil {
		logger.Fatalf("%s\n", err)
	}

	return db
}

// Start Gnomon indexer with or without search filters
//   - End point from rpc.Daemon.Rpc
//   - tag for log print
//   - dbtype defines gravdb or boltdb
//   - custom func() is for adding specific SCID to index on Gnomon start, gnomes.Fast.Enabled false will bypass
//   - lower defines the lower limit of indexed SCIDs from Gnomon search filters before custom adds
//   - upper defines the higher limit when custom indexed SCIDs exist already
func StartGnomon(tag, dbtype string, filters []string, upper, lower int, custom func()) {
	gnomes.Start = true
	logger.Printf("[%s] Starting Gnomon\n", tag)
	shasum := fmt.Sprintf("%x", sha1.Sum([]byte("dReams")))
	db_path := filepath.Join("datashards", "gnomon", fmt.Sprintf("%s_%s", "dReams", shasum))
	bolt_backend := NewBoltDB(dbtype, db_path)
	grav_backend := NewGravDB(dbtype, db_path)

	var last_height int64
	if dbtype == "boltdb" {
		last_height, _ = bolt_backend.GetLastIndexHeight()
	} else {
		last_height, _ = grav_backend.GetLastIndexHeight()
	}

	if filters != nil {
		exclusions := []string{"bb43c3eb626ee767c9f305772a6666f7c7300441a0ad8538a0799eb4f12ebcd2"}
		gnomes.Indexer = indexer.NewIndexer(grav_backend, bolt_backend, dbtype, filters, last_height, rpc.Daemon.Rpc, "daemon", false, false, &gnomes.Fast, exclusions)
		go gnomes.Indexer.StartDaemonMode(gnomes.Para)
		time.Sleep(3 * time.Second)
		gnomes.Initialized(true)

		if gnomes.Fast.Enabled {
			for {
				contracts := len(gnomes.GetAllOwnersAndSCIDs())
				if contracts >= upper {
					break
				}

				if contracts >= lower && gnomes.IsStatus("indexed") {
					custom()
					break
				}

				time.Sleep(time.Second)

				if !rpc.Daemon.IsConnected() || gnomes.IsClosing() || !gnomes.IsInitialized() {
					logger.Errorf("[%s] Could not add all custom SCIDs to index\n", tag)
					break
				}
			}
		}
	}

	gnomes.Start = false
}

// Update Gnomon endpoint to current rpc.Daemon.Rpc value
func EndPoint() {
	if rpc.Daemon.IsConnected() && gnomes.IsInitialized() && !gnomes.IsScanning() {
		gnomes.Indexer.Endpoint = rpc.Daemon.Rpc
	}
}

// Check if Gnomon and RPC are ready
func IsConnected() bool {
	if rpc.IsReady() && gnomes.IsSynced() {
		return true
	}

	return false
}

// Scan tells dApps if Gnomon is ready for them to preform their initial scan
func Scan(config bool) bool {
	if gnomes.IsSynced() && gnomes.HasChecked() && !config {
		return true
	}

	return false
}

// State checks and maintains Gnomon state (synced/scanning/checked), it will scan connected wallet once synced, then ensure sync
//   - Hold out checking if app is configuring
//   - Pass scan func for initial Gnomon sync
func State(config bool, scan func(map[string]string)) {
	if rpc.Daemon.IsConnected() && gnomes.IsRunning() {
		contracts := gnomes.IndexContains()
		if gnomes.HasIndex(2) && !gnomes.IsStarting() {
			height := gnomes.GetChainHeight()
			if gnomes.GetLastHeight() >= height-3 && height != 0 {
				gnomes.Synced(true)
				if !config && rpc.Wallet.IsConnected() && !gnomes.HasChecked() {
					gnomes.Scanning(true)
					scan(contracts)
					gnomes.Checked(true)
					gnomes.Scanning(false)
				}
			} else {
				gnomes.Synced(false)
			}
		}
	} else {
		gnomes.Synced(false)
	}
}

// Get Gnomon headers of SCID
func GetSCHeaders(scid string) (header SCHeaders) {
	if gnomes.IsRunning() {
		headers, _ := gnomes.GetSCIDValuesByKey(rpc.GnomonSCID, scid)
		if headers != nil {
			split := strings.Split(headers[0], ";")
			switch len(split) {
			case 1:
				header.Name = split[0]
			case 2:
				header.Name = split[0]
				header.Description = split[1]
			case 3:
				header.Name = split[0]
				header.Description = split[1]
				header.IconURL = split[2]
			}
		}
	}
	return
}

// Get a requested NFA url
//   - w of 0 returns "fileURL"
//   - w of 1 returns "iconURLHdr"
//   - w of 2 returns "coverURLHdr"
func GetAssetUrl(w int, scid string) (url string) {
	var link []string
	switch w {
	case 0:
		link, _ = gnomes.GetSCIDValuesByKey(scid, "fileURL")
	case 1:
		link, _ = gnomes.GetSCIDValuesByKey(scid, "iconURLHdr")
	case 2:
		link, _ = gnomes.GetSCIDValuesByKey(scid, "coverURL")
	default:
		// nothing
	}

	if link != nil {
		url = link[0]
	}

	return
}

// Get name, collection and file extension of NFA
func GetAssetInfo(scid string) (name string, collection string, extension string) {
	if n, _ := gnomes.GetSCIDValuesByKey(scid, "nameHdr"); n != nil {
		name = n[0]
	}

	if c, _ := gnomes.GetSCIDValuesByKey(scid, "collection"); c != nil {
		collection = c[0]
	}

	if f, _ := gnomes.GetSCIDValuesByKey(scid, "fileURL"); f != nil {
		extension = filepath.Ext(f[0])
	}

	return
}

// Check owner of any SCID using "owner" key
func CheckOwner(scid string) bool {
	if len(scid) != 64 || !gnomes.IsReady() {
		return false
	}

	owner, _ := gnomes.GetSCIDValuesByKey(scid, "owner")
	if owner != nil {
		return owner[0] == rpc.Wallet.Address
	}

	return false
}
