package dreams

import (
	"errors"
	"fmt"
	"image/color"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/civilware/Gnomon/structures"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/rpc"
	"github.com/sirupsen/logrus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// ContainerStack used for building various
// container/label layouts to be placed in main app
type ContainerStack struct {
	LeftLabel  *widget.Label
	RightLabel *widget.Label
	TopLabel   *canvas.Text

	Back    fyne.Container
	Front   fyne.Container
	Actions fyne.Container
	DApp    *fyne.Container
}

// Saved data for users local config.json file
type SaveData struct {
	Skin    color.Gray16 `json:"skin"`
	Daemon  []string     `json:"daemon"`
	Tables  []string     `json:"tables"`
	Predict []string     `json:"predict"`
	Sports  []string     `json:"sports"`
	G45s    []string     `json:"g45s"`
	NFAs    []string     `json:"nfas"`
	DBtype  string       `json:"dbType"`
	Para    int          `json:"paraBlocks"`

	Assets map[string]bool `json:"assets"`
	Dapps  map[string]bool `json:"dapps"`
}

// AppObject holds the main app and channels
type AppObject struct {
	App        fyne.App
	Window     fyne.Window
	Background *fyne.Container
	os         string
	configure  bool
	tab        string
	subTab     string
	quit       chan struct{}
	done       chan struct{}
	receive    chan struct{}
	channels   int
}

// Select widget items for Dero assets
type AssetSelect struct {
	Name   string
	URL    string
	Img    canvas.Image
	Select *widget.Select
}

type counter struct {
	i int
	sync.RWMutex
}

var count counter
var mu sync.RWMutex
var logger = structures.Logger.WithFields(logrus.Fields{})

// Background theme AssetSelect
var Theme AssetSelect

// Add to active channel count
func (c *counter) plus() {
	c.Lock()
	c.i++
	c.Unlock()
}

// Subtract from active channel count
func (c *counter) minus() {
	c.Lock()
	c.i--
	c.Unlock()
}

// Check active channel count
func (c *counter) active() int {
	c.RLock()
	defer c.RUnlock()

	return c.i
}

// Set what OS is being used
func (d *AppObject) SetOS() {
	d.os = runtime.GOOS
}

// Check what OS is set
func (d *AppObject) OS() string {
	return d.os
}

// Set main configure bool
func (d *AppObject) Configure(b bool) {
	mu.Lock()
	d.configure = b
	mu.Unlock()
}

// Check if main app is configuring
func (d *AppObject) IsConfiguring() bool {
	mu.RLock()
	defer mu.RUnlock()

	return d.configure
}

// Set what tab main windows is on
func (d *AppObject) SetTab(name string) {
	mu.Lock()
	d.tab = name
	mu.Unlock()
}

// Check what tab main windows is on
func (d *AppObject) OnTab(name string) bool {
	mu.RLock()
	defer mu.RUnlock()

	return d.tab == name
}

// Set what sub tab is being viewed
func (d *AppObject) SetSubTab(name string) {
	mu.Lock()
	d.subTab = name
	mu.Unlock()
}

// Check what sub tab is being viewed
func (d *AppObject) OnSubTab(name string) bool {
	mu.RLock()
	defer mu.RUnlock()

	return d.subTab == name
}

// Initialize channels
func (d *AppObject) SetChannels(i int) {
	d.receive = make(chan struct{})
	d.done = make(chan struct{})
	d.quit = make(chan struct{})
	d.channels = i
}

// Signal all available channels when we are ready for them to work
func (d *AppObject) SignalChannel() {
	for count.active() < d.channels {
		count.plus()
		d.receive <- struct{}{}
	}
}

// Receive signal for work
func (d *AppObject) Receive() <-chan struct{} {
	return d.receive
}

// Signal back to counter when work is done
func (d *AppObject) WorkDone() {
	count.minus()
}

// Close signal for a dApp
func (d *AppObject) CloseDapp() <-chan struct{} {
	return d.done
}

// Send close signal to all active dApp channels
func (d *AppObject) CloseAllDapps() {
	ch := 0
	for ch < d.channels {
		ch++
		d.done <- struct{}{}
	}

	for count.active() > 0 {
		time.Sleep(time.Second)
	}
}

// Stop the main apps process
func (d *AppObject) StopProcess() {
	d.quit <- struct{}{}
}

// Close signal for main app
func (d *AppObject) Closing() <-chan struct{} {
	return d.quit
}

// Notification pop up for main app
func (d *AppObject) Notification(title, content string) bool {
	d.App.SendNotification(&fyne.Notification{Title: title, Content: content})

	return true
}

// Check if runtime os is windows
func (d *AppObject) IsWindows() bool {
	return d.os == "windows"
}

// Get current working directory path for prefix
func GetDir() (dir string) {
	dir, err := os.Getwd()
	if err != nil {
		logger.Errorln("[GetDir]", err)
	}

	return
}

// Check if path to file exists
//   - tag for log print
func FileExists(path, tag string) bool {
	if _, err := os.Stat(path); err == nil {
		return true

	} else if errors.Is(err, os.ErrNotExist) {
		logger.Errorf("[%s] %s Not Found\n", tag, path)

		return false
	}

	return false
}

// Download image file from url and return as canvas image
func DownloadFile(URL, fileName string) (canvas.Image, error) {
	response, err := http.Get(URL)
	if err != nil {
		return *canvas.NewImageFromImage(nil), err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return *canvas.NewImageFromImage(nil), fmt.Errorf("received %d response code", response.StatusCode)
	}

	return *canvas.NewImageFromReader(response.Body, fileName), nil
}

// dReams app theme selection object
//   - If image is not present locally, it is downloaded
func ThemeSelect() fyne.Widget {
	options := []string{"Main", "Legacy"}
	Theme.Select = widget.NewSelect(options, func(s string) {
		switch Theme.Select.SelectedIndex() {
		case -1:
			Theme.Name = "Main"
		case 0:
			Theme.Name = "Main"
		case 1:
			Theme.Name = "Legacy"
		default:
			Theme.Name = s
		}
		go func() {
			dir := GetDir()
			check := strings.Trim(s, "0123456789")
			if check == "AZYDS" {
				file := dir + "/assets/" + s + "/" + s + ".png"
				if FileExists(file, "dReams") {
					Theme.Img = *canvas.NewImageFromFile(file)
				} else {
					Theme.URL = "https://raw.githubusercontent.com/Azylem/" + s + "/main/" + s + ".png"
					logger.Println("[dReams] Downloading", Theme.URL)
					Theme.Img, _ = DownloadFile(Theme.URL, s)
				}
			} else if check == "SIXART" {
				file := dir + "/assets/" + s + "/" + s + ".png"
				if FileExists(file, "dReams") {
					Theme.Img = *canvas.NewImageFromFile(file)
				} else {
					Theme.URL = "https://raw.githubusercontent.com/SixofClubsss/SIXART/main/" + s + "/" + s + ".png"
					logger.Println("[dReams] Downloading", Theme.URL)
					Theme.Img, _ = DownloadFile(Theme.URL, s)
				}
			} else if check == "HSTheme" {
				file := dir + "/assets/" + s + "/" + s + ".png"
				if FileExists(file, "dReams") {
					Theme.Img = *canvas.NewImageFromFile(file)
				} else {
					Theme.URL = "https://raw.githubusercontent.com/High-Strangeness/High-Strangeness/main/" + s + "/" + s + ".png"
					logger.Println("[dReams] Downloading", Theme.URL)
					Theme.Img, _ = DownloadFile(Theme.URL, s)
				}
			} else if s == "Main" {
				Theme.Img = *canvas.NewImageFromResource(bundle.ResourceBackgroundPng)
			} else if s == "Legacy" {
				Theme.Img = *canvas.NewImageFromResource(bundle.ResourceLegacyBackgroundPng)
			}
		}()
	})
	Theme.Select.PlaceHolder = "Theme"

	return Theme.Select
}

// Add a asset option to a AssetSelect
func (a *AssetSelect) Add(add, check string) {
	if check == rpc.Wallet.Address {
		opts := a.Select.Options
		new_opts := append(opts, add)
		a.Select.Options = new_opts
		a.Select.Refresh()
	}
}

// Clears all assets from select options
func (a *AssetSelect) ClearAll() {
	a.Select.Options = []string{}
	a.Select.Refresh()
}

// Sort the select widgets options
func (a *AssetSelect) Sort() {
	sort.Strings(a.Select.Options)
}
