package dreams

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/SixofClubsss/dReams/bundle"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type DreamsObject struct {
	App        fyne.App
	Window     fyne.Window
	Background *fyne.Container
	OS         string
	Configure  bool
	Menu       bool
	Market     bool
	Cli        bool
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

var active int
var tab string
var mu sync.RWMutex
var Theme AssetSelect

func (d *DreamsObject) SetTab(name string) {
	mu.Lock()
	tab = name
	mu.Unlock()
}

func (d *DreamsObject) OnTab(name string) bool {
	mu.RLock()
	defer mu.RUnlock()

	return tab == name
}

func (d *DreamsObject) SetChannels(i int) {
	d.receive = make(chan struct{})
	d.done = make(chan struct{})
	d.quit = make(chan struct{})
	d.channels = i
}

func (d *DreamsObject) SignalChannel() {
	for active < d.channels {
		active++
		d.receive <- struct{}{}
	}
}

func (d *DreamsObject) Receive() <-chan struct{} {
	return d.receive
}

func (d *DreamsObject) WorkDone() {
	active--
}

func (d *DreamsObject) CloseDapp() <-chan struct{} {
	return d.done
}

func (d *DreamsObject) CloseAllDapps() {
	ch := 0
	for ch < d.channels {
		ch++
		d.done <- struct{}{}
	}
}

func (d *DreamsObject) StopProcess() {
	d.quit <- struct{}{}
}

func (d *DreamsObject) Closing() <-chan struct{} {
	return d.quit
}

// Notification switch for dApps
func (d *DreamsObject) Notification(title, content string) bool {
	// switch g {
	// case 0:
	// 	holdero.Round.Notified = true
	// case 1:
	// 	rpc.Bacc.Notified = true
	// case 2:
	// 	tarot.Iluma.Value.Notified = true
	// default:
	// }

	d.App.SendNotification(&fyne.Notification{Title: title, Content: content})

	return true
}

// Check if runtime os is windows
func (d *DreamsObject) IsWindows() bool {
	return d.OS == "windows"
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
				file := dir + "/cards/" + s + "/" + s + ".png"
				if FileExists(file, "dReams") {
					Theme.Img = *canvas.NewImageFromFile(file)
				} else {
					Theme.URL = "https://raw.githubusercontent.com/Azylem/" + s + "/main/" + s + ".png"
					log.Println("[dReams] Downloading", Theme.URL)
					Theme.Img, _ = DownloadFile(Theme.URL, s)
				}
			} else if check == "SIXART" {
				dir := GetDir()
				file := dir + "/cards/" + s + "/" + s + ".png"
				if FileExists(file, "dReams") {
					Theme.Img = *canvas.NewImageFromFile(file)
				} else {
					Theme.URL = "https://raw.githubusercontent.com/SixofClubsss/SIXART/main/" + s + "/" + s + ".png"
					log.Println("[dReams] Downloading", Theme.URL)
					Theme.Img, _ = DownloadFile(Theme.URL, s)
				}
			} else if check == "HSTheme" {
				dir := GetDir()
				file := dir + "/cards/" + s + "/" + s + ".png"
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
