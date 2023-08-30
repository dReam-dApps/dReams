package menu

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/civilware/Gnomon/indexer"
	"github.com/civilware/Gnomon/structures"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
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
	if g.IsInitialized() && !g.IsClosing() {
		logger.Printf("[%s] Putting Gnomon to Sleep\n", tag)
		g.Lock()
		g.Indexer.Close()
		g.Init = false
		g.Check = false
		g.Unlock()
		logger.Printf("[%s] Gnomon is Sleeping\n", tag)
	}
}

// Check if Gnomon is writing
func (g *gnomon) IsWriting() bool {
	g.RLock()
	defer g.RUnlock()
	switch g.Indexer.DBType {
	case "gravdb":
		return g.Indexer.GravDBBackend.Writing == 1
	case "boltdb":
		return g.Indexer.BBSBackend.Writing == 1
	default:
		return g.Indexer.BBSBackend.Writing == 1
	}
}

// Set Indexer.Backend.Writing var,
// if set true will wait if Indexer is writing already
func (g *gnomon) Writing(b bool) {
	for b && g.IsWriting() {
		time.Sleep(30 * time.Millisecond)
	}

	i := 0
	if b {
		i = 1
	}

	g.Lock()
	defer g.Unlock()
	switch g.Indexer.DBType {
	case "gravdb":
		g.Indexer.GravDBBackend.Writing = i
	case "boltdb":
		g.Indexer.BBSBackend.Writing = i
	default:
		g.Indexer.BBSBackend.Writing = i
	}
}

// Check if Gnomon is closing
func (g *gnomon) IsClosing() bool {
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
		return g.Indexer.BBSBackend.Closing
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

	if g.Init && !g.IsClosing() {
		return true
	}

	return false
}

// Check if Gnomon is initialized, synced and not closing
func (g *gnomon) IsReady() bool {
	g.RLock()
	defer g.RUnlock()

	if g.Init && g.Sync && !g.IsClosing() {
		return true
	}

	return false
}

// Method of Gnomon GetAllOwnersAndSCIDs() where DB type is defined by Indexer.DBType
//   - Default is boltdb
func (g *gnomon) GetAllOwnersAndSCIDs() map[string]string {
	switch g.Indexer.DBType {
	case "gravdb":
		return g.Indexer.GravDBBackend.GetAllOwnersAndSCIDs()
	case "boltdb":
		return g.Indexer.BBSBackend.GetAllOwnersAndSCIDs()
	default:
		return g.Indexer.BBSBackend.GetAllOwnersAndSCIDs()
	}
}

// Method of Gnomon GetSCIDValuesByKey() where DB type is defined by Indexer.DBType
//   - Default is boltdb
func (g *gnomon) GetSCIDValuesByKey(scid string, key interface{}) (valuesstring []string, valuesuint64 []uint64) {
	switch g.Indexer.DBType {
	case "gravdb":
		return g.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, key, g.Indexer.ChainHeight, true)
	case "boltdb":
		return g.Indexer.BBSBackend.GetSCIDValuesByKey(scid, key, g.Indexer.ChainHeight, true)
	default:
		return g.Indexer.BBSBackend.GetSCIDValuesByKey(scid, key, g.Indexer.ChainHeight, true)
	}
}

// Method of Gnomon GetSCIDKeysByValue() where DB type is defined by Indexer.DBType
//   - Default is boltdb
func (g *gnomon) GetSCIDKeysByValue(scid string, key interface{}) (valuesstring []string, valuesuint64 []uint64) {
	switch g.Indexer.DBType {
	case "gravdb":
		return g.Indexer.GravDBBackend.GetSCIDKeysByValue(scid, key, g.Indexer.ChainHeight, true)
	case "boltdb":
		return g.Indexer.BBSBackend.GetSCIDKeysByValue(scid, key, g.Indexer.ChainHeight, true)
	default:
		return g.Indexer.BBSBackend.GetSCIDKeysByValue(scid, key, g.Indexer.ChainHeight, true)
	}
}

// Method of Gnomon GetAllSCIDVariableDetails() where DB type is defined by Indexer.DBType
//   - Default is boltdb
func (g *gnomon) GetAllSCIDVariableDetails(scid string) map[int64][]*structures.SCIDVariable {
	switch g.Indexer.DBType {
	case "gravdb":
		return g.Indexer.GravDBBackend.GetAllSCIDVariableDetails(scid)
	case "boltdb":
		return g.Indexer.BBSBackend.GetAllSCIDVariableDetails(scid)
	default:
		return g.Indexer.BBSBackend.GetAllSCIDVariableDetails(scid)
	}
}

// UI control panel to set Gnomes vars
func (g *gnomon) ControlPanel(w fyne.Window) *fyne.Container {
	db := widget.NewRadioGroup([]string{"boltdb", "gravdb"}, func(s string) {
		g.DBType = s
	})
	db.Horizontal = true
	db.SetSelected(g.DBType)

	fast := widget.NewRadioGroup([]string{"true", "false"}, func(s string) {
		if b, err := strconv.ParseBool(s); err == nil {
			g.Fast = b

			return
		}

		g.Fast = true
	})
	fast.Horizontal = true
	fast.SetSelected(strconv.FormatBool(g.Fast))

	trim := widget.NewRadioGroup([]string{"true", "false"}, func(s string) {
		if b, err := strconv.ParseBool(s); err == nil {
			g.Trim = b

			return
		}

		g.Trim = true
	})
	trim.Horizontal = true

	para := widget.NewSelect([]string{"1", "2", "3", "4", "5"}, func(s string) {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			g.Para = int(i)

			return
		}

		g.Para = 1
	})

	if g.Para < 6 && g.Para > 1 {
		para.SetSelectedIndex(g.Para - 1)
	} else {
		para.SetSelectedIndex(0)
	}

	delete_db := widget.NewButton("Delete DB", func() {
		dialog.NewConfirm("Delete DB", "This will delete your current Gnomon DB", func(b bool) {
			if b {
				os.RemoveAll(filepath.Clean("gnomondb"))
			}
		}, w).Show()
	})

	gnomes_form := []*widget.FormItem{}
	gnomes_form = append(gnomes_form, widget.NewFormItem("DB Type", db))
	gnomes_form = append(gnomes_form, widget.NewFormItem("Fastsync", fast))
	gnomes_form = append(gnomes_form, widget.NewFormItem("Pruned Index", trim))
	gnomes_form = append(gnomes_form, widget.NewFormItem("Parallel Blocks", para))

	return container.NewBorder(nil, container.NewCenter(delete_db), nil, nil, widget.NewForm(gnomes_form...))
}
