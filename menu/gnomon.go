package menu

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

var logger = structures.Logger.WithFields(logrus.Fields{})

// Enable escape codes for windows Stdout
func enableEscapeCodes() error {
	cmd := exec.Command("cmd", "/c", "echo", "ON")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

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

// Manually add SCID to Gnomon index
func AddToIndex(scid []string) (err error) {
	filters := Gnomes.Indexer.SearchFilter
	Gnomes.Indexer.SearchFilter = []string{}
	scidstoadd := make(map[string]*structures.FastSyncImport)

	for _, sc := range scid {
		owner, _ := Gnomes.GetSCIDValuesByKey(rpc.GnomonSCID, sc+"owner")
		if owner != nil {
			scidstoadd[sc] = &structures.FastSyncImport{}
			scidstoadd[sc].Owner = owner[0]
		}
	}

	err = Gnomes.Indexer.AddSCIDToIndex(scidstoadd, false, false)
	if err != nil {
		logger.Errorf("[AddToIndex] %v\n", err)
	}
	Gnomes.Indexer.SearchFilter = filters

	return
}

// Create Gnomon graviton db with dReams tag
//   - If dbType is boltdb, will return nil gravdb
func GnomonGravDB(dbType, dbPath string) *storage.GravitonStore {
	if dbType == "boltdb" {
		return nil
	}

	db, err := storage.NewGravDB(dbPath, "25ms")
	if err != nil {
		logger.Fatalf("[GnomonGravDB] %s\n", err)
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
			logger.Fatalf("[GnomonBoltDB] %s\n", err)
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
//   - custom func() is for adding specific SCID to index on Gnomon start, Gnomes.Fast false will bypass
//   - lower defines the lower limit of indexed SCIDs from Gnomon search filters before custom adds
//   - upper defines the higher limit when custom indexed SCIDs exist already
func StartGnomon(tag, dbtype string, filters []string, upper, lower int, custom func()) {
	Gnomes.Start = true
	logger.Printf("[%s] Starting Gnomon\n", tag)
	shasum := fmt.Sprintf("%x", sha1.Sum([]byte("dReams")))
	db_path := filepath.Join("gnomondb", fmt.Sprintf("%s_%s", "dReams", shasum))
	bolt_backend := GnomonBoltDB(dbtype, db_path)
	grav_backend := GnomonGravDB(dbtype, db_path)

	var last_height int64
	if dbtype == "boltdb" {
		last_height, _ = bolt_backend.GetLastIndexHeight()
	} else {
		last_height, _ = grav_backend.GetLastIndexHeight()
	}

	if filters != nil {
		Gnomes.Indexer = indexer.NewIndexer(grav_backend, bolt_backend, dbtype, filters, last_height, rpc.Daemon.Rpc, "daemon", false, false, Gnomes.Fast, false, nil)
		go Gnomes.Indexer.StartDaemonMode(Gnomes.Para)
		time.Sleep(3 * time.Second)
		Gnomes.Initialized(true)

		if Gnomes.Fast {
			for {
				contracts := len(Gnomes.GetAllOwnersAndSCIDs())
				if contracts >= upper {
					break
				}

				if contracts >= lower {
					custom()
					break
				}

				time.Sleep(time.Second)

				if !rpc.Daemon.IsConnected() || ClosingApps() {
					logger.Errorf("[%s] Could not add all custom SCIDs to index\n", tag)
					break
				}
			}
		}
	}

	Gnomes.Start = false
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

// Gnomon is ready for dApp to preform initial scan
func GnomonScan(config bool) bool {
	if Gnomes.IsSynced() && Gnomes.HasChecked() && !config {
		return true
	}

	return false
}

// Gnomon will scan connected wallet on start up, then ensure sync
//   - Hold out checking if dReams is in configure
//   - Pass scan func for initial Gnomon sync
func GnomonState(config bool, scan func(map[string]string)) {
	if rpc.Daemon.IsConnected() && Gnomes.IsRunning() {
		contracts := Gnomes.IndexContains()
		if Gnomes.HasIndex(2) && !Gnomes.Start {
			height := Gnomes.Indexer.ChainHeight
			if Gnomes.Indexer.LastIndexedHeight >= height-3 && height != 0 {
				Gnomes.Synced(true)
				if !config && rpc.Wallet.IsConnected() && !Gnomes.HasChecked() {
					Gnomes.Scanning(true)

					CheckWalletNames(rpc.Wallet.Address)
					scan(contracts)
					FindNFAListings(contracts)

					Gnomes.Checked(true)
					Gnomes.Scanning(false)
				}
			} else {
				Gnomes.Synced(false)
			}
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
