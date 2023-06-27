package dreams

import (
	"errors"
	"fmt"
	"image/color"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type DreamsItems struct {
	LeftLabel  *widget.Label
	RightLabel *widget.Label
	TopLabel   *canvas.Text

	Back    fyne.Container
	Front   fyne.Container
	Actions fyne.Container
	DApp    *fyne.Container
}

type DreamSave struct {
	Skin    color.Gray16 `json:"skin"`
	Daemon  []string     `json:"daemon"`
	Tables  []string     `json:"tables"`
	Predict []string     `json:"predict"`
	Sports  []string     `json:"sports"`

	Dapps map[string]bool `json:"dapps"`
}

type DreamsObject struct {
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
func (d *DreamsObject) SetOS() {
	d.os = runtime.GOOS
}

// Check what OS is set
func (d *DreamsObject) OS() string {
	return d.os
}

// Set dReams configure bool
func (d *DreamsObject) Configure(b bool) {
	mu.Lock()
	d.configure = b
	mu.Unlock()
}

// Check if dReams is configuring
func (d *DreamsObject) IsConfiguring() bool {
	mu.RLock()
	defer mu.RUnlock()

	return d.configure
}

// Set what tab main windows is on
func (d *DreamsObject) SetTab(name string) {
	mu.Lock()
	d.tab = name
	mu.Unlock()
}

// Check what tab main windows is on
func (d *DreamsObject) OnTab(name string) bool {
	mu.RLock()
	defer mu.RUnlock()

	return d.tab == name
}

// Set what sub tab is being viewed
func (d *DreamsObject) SetSubTab(name string) {
	mu.Lock()
	d.subTab = name
	mu.Unlock()
}

// Check what sub tab is being viewed
func (d *DreamsObject) OnSubTab(name string) bool {
	mu.RLock()
	defer mu.RUnlock()

	return d.subTab == name
}

// Initialize channels
func (d *DreamsObject) SetChannels(i int) {
	d.receive = make(chan struct{})
	d.done = make(chan struct{})
	d.quit = make(chan struct{})
	d.channels = i
}

// Signal all available channels when we are ready for them to work
func (d *DreamsObject) SignalChannel() {
	for count.active() < d.channels {
		count.plus()
		d.receive <- struct{}{}
	}
}

// Receive signal for work
func (d *DreamsObject) Receive() <-chan struct{} {
	return d.receive
}

// Signal back to counter when work is done
func (d *DreamsObject) WorkDone() {
	count.minus()
}

// Close signal for a dApp
func (d *DreamsObject) CloseDapp() <-chan struct{} {
	return d.done
}

// Send close signal to all active dApp channels
func (d *DreamsObject) CloseAllDapps() {
	ch := 0
	for ch < d.channels {
		ch++
		d.done <- struct{}{}
	}

	for count.active() > 0 {
		time.Sleep(time.Second)
	}
}

// Stop the main dReams process
func (d *DreamsObject) StopProcess() {
	d.quit <- struct{}{}
}

// Close signal for dReams
func (d *DreamsObject) Closing() <-chan struct{} {
	return d.quit
}

// Notification pop up for dReams app
func (d *DreamsObject) Notification(title, content string) bool {
	d.App.SendNotification(&fyne.Notification{Title: title, Content: content})

	return true
}

// Check if runtime os is windows
func (d *DreamsObject) IsWindows() bool {
	return d.os == "windows"
}

// Get current working directory path for prefix
func GetDir() (dir string) {
	dir, err := os.Getwd()
	if err != nil {
		log.Println("[GetDir]", err)
	}

	return
}

// Check if path to file exists
//   - tag for log print
func FileExists(path, tag string) bool {
	if _, err := os.Stat(path); err == nil {
		return true

	} else if errors.Is(err, os.ErrNotExist) {
		log.Printf("[%s] %s Not Found\n", tag, path)

		return false
	}

	return false
}

// Download image file from url and return as canvas image
func DownloadFile(Url, fileName string) (canvas.Image, error) {
	response, err := http.Get(Url)
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
			check := strings.Trim(s, "0123456789")
			if check == "AZYDS" {
				dir := GetDir()
				file := dir + "/assets/" + s + "/" + s + ".png"
				if FileExists(file, "dReams") {
					Theme.Img = *canvas.NewImageFromFile(file)
				} else {
					Theme.URL = "https://raw.githubusercontent.com/Azylem/" + s + "/main/" + s + ".png"
					log.Println("[dReams] Downloading", Theme.URL)
					Theme.Img, _ = DownloadFile(Theme.URL, s)
				}
			} else if check == "SIXART" {
				dir := GetDir()
				file := dir + "/assets/" + s + "/" + s + ".png"
				if FileExists(file, "dReams") {
					Theme.Img = *canvas.NewImageFromFile(file)
				} else {
					Theme.URL = "https://raw.githubusercontent.com/SixofClubsss/SIXART/main/" + s + "/" + s + ".png"
					log.Println("[dReams] Downloading", Theme.URL)
					Theme.Img, _ = DownloadFile(Theme.URL, s)
				}
			} else if check == "HSTheme" {
				dir := GetDir()
				file := dir + "/assets/" + s + "/" + s + ".png"
				if FileExists(file, "dReams") {
					Theme.Img = *canvas.NewImageFromFile(file)
				} else {
					Theme.URL = "https://raw.githubusercontent.com/High-Strangeness/High-Strangeness/main/" + s + "/" + s + ".png"
					log.Println("[dReams] Downloading", Theme.URL)
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

func (a *AssetSelect) Add(add, check string) {
	if check == rpc.Wallet.Address {
		opts := a.Select.Options
		new_opts := append(opts, add)
		a.Select.Options = new_opts
		a.Select.Refresh()
	}
}
