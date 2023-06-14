package menu

import (
	"log"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	xwidget "fyne.io/x/fyne/widget"
	"github.com/civilware/Gnomon/indexer"
	"github.com/civilware/Gnomon/structures"
)

type gnomon struct {
	DBType   string
	Para     int
	Trim     bool
	Fast     bool
	Start    bool
	Init     bool
	Sync     bool
	Syncing  bool
	Check    bool
	Wait     bool
	Import   bool
	SCIDS    uint64
	Sync_ind *fyne.Animation
	Full_ind *fyne.Animation
	Icon_ind *xwidget.AnimatedGif
	Indexer  *indexer.Indexer
	sync.RWMutex
}

var Gnomes gnomon

// Shut down Gnomes.Indexer
//   - tag for log print
func (g *gnomon) Stop(tag string) {
	g.Lock()
	if g.Init && !g.Closing() {
		log.Printf("[%s] Putting Gnomon to Sleep\n", tag)
		g.Indexer.Close()
		g.Init = false
		time.Sleep(1 * time.Second)
		log.Printf("[%s] Gnomon is Sleeping\n", tag)
	}
	g.Unlock()
}

// Check if Gnomon is writing
func (g *gnomon) Writing() bool {
	switch g.Indexer.DBType {
	case "gravdb":
		return g.Indexer.GravDBBackend.Writing == 1
	case "boltdb":
		return g.Indexer.BBSBackend.Writing == 1
	default:
		return g.Indexer.GravDBBackend.Writing == 1
	}
}

// Check if Gnomon is closing
func (g *gnomon) Closing() bool {
	if !g.Init {
		return false
	}

	if g.Indexer.Closing {
		return true
	}

	switch g.Indexer.DBType {
	case "gravdb":
		return g.Indexer.GravDBBackend.Closing
	case "boltdb":
		return g.Indexer.BBSBackend.Closing
	default:
		return false
	}
}

// Set Gnomes.Init var
func (g *gnomon) Initialized(b bool) {
	g.Lock()
	g.Init = b
	g.Unlock()
}

// Check if Gnomes.Init
func (g *gnomon) IsInitialized() bool {
	g.RLock()
	defer g.RUnlock()

	return g.Init
}

// Set Gnomes.Syncing var when scanning wallet
func (g *gnomon) Scanning(b bool) {
	g.Lock()
	g.Syncing = b
	g.Unlock()
}

// Check if Gnomes.Syncing
func (g *gnomon) IsScanning() bool {
	g.RLock()
	defer g.RUnlock()

	return g.Syncing
}

// Set Gnomes.Checked var
func (g *gnomon) Checked(b bool) {
	g.Lock()
	g.Check = b
	g.Unlock()
}

// Check if Gnomes.Checked
func (g *gnomon) HasChecked() bool {
	g.RLock()
	defer g.RUnlock()

	return g.Check
}

// Set Gnomes.SCIDS index count and return GetAllOwnersAndSCIDs()
func (g *gnomon) IndexContains() map[string]string {
	contracts := g.GetAllOwnersAndSCIDs()

	g.Lock()
	g.SCIDS = uint64(len(contracts))
	g.Unlock()

	return contracts
}

// Check if Gnomes index contains SCIDs >= u
func (g *gnomon) HasIndex(u uint64) bool {
	g.RLock()
	defer g.RUnlock()

	return g.SCIDS >= u
}

// Set Gnomes.Sync var
func (g *gnomon) Synced(b bool) {
	g.Lock()
	g.Sync = b
	g.Unlock()
}

// Check if Gnomes.Sync
func (g *gnomon) IsSynced() bool {
	g.RLock()
	defer g.RUnlock()

	return g.Sync
}

// Check if Gnomon is initialized, and not closing
func (g *gnomon) IsRunning() bool {
	g.RLock()
	defer g.RUnlock()

	if g.Init && !g.Closing() {
		return true
	}

	return false
}

// Check if Gnomon is initialized, synced and not closing
func (g *gnomon) IsReady() bool {
	g.RLock()
	defer g.RUnlock()

	if g.Init && g.Sync && !g.Closing() {
		return true
	}

	return false
}

// Method of Gnomon GetAllOwnersAndSCIDs() where DB type is defined by Indexer.DBType
//   - Default is gravdb
func (g *gnomon) GetAllOwnersAndSCIDs() map[string]string {
	switch g.Indexer.DBType {
	case "gravdb":
		return g.Indexer.GravDBBackend.GetAllOwnersAndSCIDs()
	case "boltdb":
		return g.Indexer.BBSBackend.GetAllOwnersAndSCIDs()
	default:
		return g.Indexer.GravDBBackend.GetAllOwnersAndSCIDs()
	}
}

// Method of Gnomon GetSCIDValuesByKey() where DB type is defined by Indexer.DBType
//   - Default is gravdb
func (g *gnomon) GetSCIDValuesByKey(scid string, key interface{}) (valuesstring []string, valuesuint64 []uint64) {
	switch g.Indexer.DBType {
	case "gravdb":
		return g.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, key, g.Indexer.ChainHeight, true)
	case "boltdb":
		return g.Indexer.BBSBackend.GetSCIDValuesByKey(scid, key, g.Indexer.ChainHeight, true)
	default:
		return g.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, key, g.Indexer.ChainHeight, true)
	}
}

// Method of Gnomon GetSCIDKeysByValue() where DB type is defined by Indexer.DBType
//   - Default is gravdb
func (g *gnomon) GetSCIDKeysByValue(scid string, key interface{}) (valuesstring []string, valuesuint64 []uint64) {
	switch g.Indexer.DBType {
	case "gravdb":
		return g.Indexer.GravDBBackend.GetSCIDKeysByValue(scid, key, g.Indexer.ChainHeight, true)
	case "boltdb":
		return g.Indexer.BBSBackend.GetSCIDKeysByValue(scid, key, g.Indexer.ChainHeight, true)
	default:
		return g.Indexer.GravDBBackend.GetSCIDKeysByValue(scid, key, g.Indexer.ChainHeight, true)
	}
}

// Method of Gnomon GetAllSCIDVariableDetails() where DB type is defined by Indexer.DBType
//   - Default is gravdb
func (g *gnomon) GetAllSCIDVariableDetails(scid string) map[int64][]*structures.SCIDVariable {
	switch g.Indexer.DBType {
	case "gravdb":
		return g.Indexer.GravDBBackend.GetAllSCIDVariableDetails(scid)
	case "boltdb":
		return g.Indexer.BBSBackend.GetAllSCIDVariableDetails(scid)
	default:
		return g.Indexer.GravDBBackend.GetAllSCIDVariableDetails(scid)
	}
}
