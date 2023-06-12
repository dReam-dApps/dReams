package menu

import (
	"log"
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

// Shut down Gnomes.Indexer
//   - tag for log print
func (g *gnomon) Stop(tag string) {
	if g.Init && !g.Closing() {
		log.Printf("[%s] Putting Gnomon to Sleep\n", tag)
		g.Indexer.Close()
		g.Init = false
		time.Sleep(1 * time.Second)
		log.Printf("[%s] Gnomon is Sleeping\n", tag)
	}
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
