package gnomes

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

type Gnomon struct {
	DBType  string
	Para    int
	Fast    bool
	Start   bool
	Init    bool
	Sync    bool
	Syncing bool
	Check   bool
	SCIDS   uint64
	Indexer *indexer.Indexer
	sync.RWMutex
}

var Imported bool
var Sync_ind *fyne.Animation
var Full_ind *fyne.Animation
var Icon_ind *xwidget.AnimatedGif

// GnomonInterface defines the methods of the gnomon type that you want to expose
type Gnomes interface {
	Status() string
	IsStarting() bool
	DBStorageType() string
	SetDBStorageType(s string)
	GetLastHeight() int64
	GetChainHeight() int64
	Stop(tag string)
	IsWriting() bool
	Writing(b bool)
	IsClosing() bool
	Initialized(b bool)
	IsInitialized() bool
	Scanning(b bool)
	IsScanning() bool
	Checked(b bool)
	HasChecked() bool
	IndexContains() map[string]string
	IndexCount() uint64
	ZeroIndexCount()
	HasIndex(u uint64) bool
	Synced(b bool)
	IsSynced() bool
	IsRunning() bool
	IsReady() bool
	SetFastsync(b bool)
	SetParallel(i int)
	GetParallel() int
	SetSearchFilters(filters []string)
	GetSearchFilters() []string
	GetAllOwnersAndSCIDs() map[string]string
	GetSCIDValuesByKey(scid string, key interface{}) (valuesstring []string, valuesuint64 []uint64)
	GetSCIDKeysByValue(scid string, key interface{}) (valuesstring []string, valuesuint64 []uint64)
	GetAllSCIDVariableDetails(scid string) []*structures.SCIDVariable
	AddSCIDToIndex(scids map[string]*structures.FastSyncImport) error
	GetLiveSCIDValuesByKey(scid string, key interface{}) (valuesstring []string, valuesuint64 []uint64, err error)
	ControlPanel(w fyne.Window) *fyne.Container
}

var gnomes Gnomon

func NewGnomes() Gnomes {
	return &gnomes
}

// Returns Status string from Indexer
func (g *Gnomon) Status() (status string) {
	g.RLock()
	defer g.RUnlock()

	if g.Indexer != nil {
		return g.Indexer.Status
	}

	return
}

// Returns true if gnomes is starting
func (g *Gnomon) IsStarting() bool {
	g.RLock()
	defer g.RUnlock()

	return g.Start
}

// Get DBType
func (g *Gnomon) DBStorageType() string {
	return g.DBType
}

// Set DB type to be used
//   - boltdb
//   - gravdb
func (g *Gnomon) SetDBStorageType(s string) {
	g.DBType = s
}

// Get Indexer.LastIndexedHeight
func (g *Gnomon) GetLastHeight() (height int64) {
	if g.Indexer != nil {
		return g.Indexer.LastIndexedHeight
	}

	return
}

// Get Indexer.ChainHeight
func (g *Gnomon) GetChainHeight() (height int64) {
	if g.Indexer != nil {
		return g.Indexer.ChainHeight
	}
	return
}

// Shut down gnomes.Indexer
//   - tag for log print
func (g *Gnomon) Stop(tag string) {
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

// Check if Indexer is writing
func (g *Gnomon) IsWriting() bool {
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
func (g *Gnomon) Writing(b bool) {
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
func (g *Gnomon) IsClosing() bool {
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
func (g *Gnomon) Initialized(b bool) {
	g.Lock()
	g.Init = b
	g.Unlock()
}

// Check if Gnomes.Init
func (g *Gnomon) IsInitialized() bool {
	g.RLock()
	defer g.RUnlock()

	return g.Init
}

// Set Gnomes.Syncing var when scanning wallet
func (g *Gnomon) Scanning(b bool) {
	g.Lock()
	g.Syncing = b
	g.Unlock()
}

// Check if Gnomes.Syncing
func (g *Gnomon) IsScanning() bool {
	g.RLock()
	defer g.RUnlock()

	return g.Syncing
}

// Set Gnomes.Checked var
func (g *Gnomon) Checked(b bool) {
	g.Lock()
	g.Check = b
	g.Unlock()
}

// Check if Gnomes.Checked
func (g *Gnomon) HasChecked() bool {
	g.RLock()
	defer g.RUnlock()

	return g.Check
}

// Set Gnomes.SCIDS index count and return GetAllOwnersAndSCIDs()
func (g *Gnomon) IndexContains() map[string]string {
	contracts := g.GetAllOwnersAndSCIDs()

	g.Lock()
	g.SCIDS = uint64(len(contracts))
	g.Unlock()

	return contracts
}

// Returns Gnomes.SCIDS
func (g *Gnomon) IndexCount() uint64 {
	g.RLock()
	defer g.RUnlock()

	return g.SCIDS
}

// Set index count to zero
func (g *Gnomon) ZeroIndexCount() {
	g.Lock()
	g.SCIDS = 0
	g.Unlock()
}

// Check if Gnomes index contains SCIDs >= u
func (g *Gnomon) HasIndex(u uint64) bool {
	g.RLock()
	defer g.RUnlock()

	return g.SCIDS >= u
}

// Set Gnomes.Sync var
func (g *Gnomon) Synced(b bool) {
	g.Lock()
	g.Sync = b
	g.Unlock()
}

// Check if Gnomes.Sync
func (g *Gnomon) IsSynced() bool {
	g.RLock()
	defer g.RUnlock()

	return g.Sync
}

// Check if Gnomon is initialized, and not closing
func (g *Gnomon) IsRunning() bool {
	g.RLock()
	defer g.RUnlock()

	if g.Init && !g.IsClosing() {
		return true
	}

	return false
}

// Check if Gnomon is initialized, synced and not closing
func (g *Gnomon) IsReady() bool {
	g.RLock()
	defer g.RUnlock()

	if g.Init && g.Sync && !g.IsClosing() {
		return true
	}

	return false
}

// Set fastsync bool
func (g *Gnomon) SetFastsync(b bool) {
	g.Fast = b
}

// Set Indexer parallel blocks value
func (g *Gnomon) SetParallel(i int) {
	g.Lock()
	g.Para = i
	g.Unlock()
}

// Get Indexer parallel blocks value
func (g *Gnomon) GetParallel() int {
	g.RLock()
	defer g.RUnlock()

	return g.Para
}

// Set Indexer search filters
func (g *Gnomon) SetSearchFilters(filters []string) {
	g.Indexer.SearchFilter = filters
}

// Get Indexer search filters
func (g *Gnomon) GetSearchFilters() []string {
	return g.Indexer.SearchFilter
}

// Method of Gnomon GetAllOwnersAndSCIDs() where DB type is defined by Indexer.DBType
//   - Default is boltdb
func (g *Gnomon) GetAllOwnersAndSCIDs() map[string]string {
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
func (g *Gnomon) GetSCIDValuesByKey(scid string, key interface{}) (valuesstring []string, valuesuint64 []uint64) {
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
func (g *Gnomon) GetSCIDKeysByValue(scid string, key interface{}) (valuesstring []string, valuesuint64 []uint64) {
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
func (g *Gnomon) GetAllSCIDVariableDetails(scid string) []*structures.SCIDVariable {
	switch g.Indexer.DBType {
	case "gravdb":
		return g.Indexer.GravDBBackend.GetAllSCIDVariableDetails(scid)
	case "boltdb":
		return g.Indexer.BBSBackend.GetAllSCIDVariableDetails(scid)
	default:
		return g.Indexer.BBSBackend.GetAllSCIDVariableDetails(scid)
	}
}

// Method of Gnomon Indexer.AddSCIDToIndex()
func (g *Gnomon) AddSCIDToIndex(scids map[string]*structures.FastSyncImport) error {
	return g.Indexer.AddSCIDToIndex(scids, false, false)
}

// Method of Gnomon Indexer.GetSCIDValuesByKey()
func (g *Gnomon) GetLiveSCIDValuesByKey(scid string, key interface{}) (valuesstring []string, valuesuint64 []uint64, err error) {
	var v []*structures.SCIDVariable
	return g.Indexer.GetSCIDValuesByKey(v, scid, key, g.Indexer.ChainHeight)
}

// UI control panel to set Gnomes vars
func (g *Gnomon) ControlPanel(w fyne.Window) *fyne.Container {
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
	gnomes_form = append(gnomes_form, widget.NewFormItem("Parallel Blocks", para))

	return container.NewBorder(nil, container.NewCenter(delete_db), nil, nil, widget.NewForm(gnomes_form...))
}
