package dreams

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/civilware/Gnomon/structures"
	"github.com/dReam-dApps/dReams/rpc"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/walletapi/xswd"
	"github.com/sirupsen/logrus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	MIN_WIDTH  = 1400
	MIN_HEIGHT = 800
)

// Saved data for users local config.json file
type SaveData struct {
	Skin   color.Gray16 `json:"skin"`
	Daemon []string     `json:"daemon"`
	Theme  string       `json:"theme"`

	DBtype  string `json:"dbType"`
	Para    int    `json:"paraBlocks"`
	FSForce bool   `json:"fastsyncForce"`
	FSDiff  int64  `json:"fastsyncDiff"`

	Assets map[string]bool `json:"assets"`
	Dapps  map[string]bool `json:"dapps"`
}

// AppObject holds the main app and channels
type AppObject struct {
	App        fyne.App
	Window     fyne.Window
	Background *fyne.Container
	XSWD       *xswd.ApplicationData
	os         string
	configure  bool
	tab        string
	subTab     string
	quit       chan struct{}
	done       chan struct{}
	receive    chan struct{}
	channels   int
	account    struct {
		handlers map[string]func(interface{}) error
	}
}

// Select widget items for Dero assets
type AssetSelect struct {
	Name   string
	URL    string
	Img    canvas.Image
	Select *widget.Select
}

type count struct {
	i       int
	calling bool
	closing bool
	sync.RWMutex
}

var counter count
var mu sync.RWMutex
var ms = 100 * time.Millisecond
var logger = structures.Logger.WithFields(logrus.Fields{})

// Background theme AssetSelect
var Theme AssetSelect

// Add to active channel count
func (c *count) plus() {
	c.Lock()
	c.i++
	c.Unlock()
}

// Subtract from active channel count
func (c *count) minus() {
	c.Lock()
	c.i--
	c.Unlock()
}

// Check active channel count
func (c *count) active() int {
	c.RLock()
	defer c.RUnlock()

	return c.i
}

// Creates a Fyne app returned as AppObject. Takes ID, name, fyne.Theme, fyne.Resources for icon and background images and XSWD application data.
// Window default size is MIN_WIDTH x MIN_HEIGHT, centered and set as master, Theme.Img uses theme resource for background.
func NewFyneApp(id, name string, skin fyne.Theme, icon, theme fyne.Resource, ad *xswd.ApplicationData) AppObject {
	a := app.NewWithID(id)
	a.Settings().SetTheme(skin)

	w := a.NewWindow(name)
	w.Resize(fyne.NewSize(MIN_WIDTH, MIN_HEIGHT))
	w.SetIcon(icon)
	w.CenterOnScreen()
	w.SetMaster()

	Theme.Img = *canvas.NewImageFromResource(theme)

	return AppObject{
		App:        a,
		Window:     w,
		Background: container.NewStack(&Theme.Img),
		XSWD:       ad,
	}
}

// Default balance label with dReams, DERO and wallet height
func SetBalanceLabelText() string {
	return fmt.Sprintf("dReams Balance: %s      DERO Balance: %s      Height: %d", rpc.Wallet.BalanceF("dReams"), rpc.Wallet.BalanceF("DERO"), rpc.Wallet.Height())
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
	if !counter.closing {
		counter.calling = true
		for counter.active() < d.channels {
			counter.plus()
			d.receive <- struct{}{}
		}
		counter.calling = false
	}
}

// Receive signal for work
func (d *AppObject) Receive() <-chan struct{} {
	return d.receive
}

// Signal back to counter when work is done
func (d *AppObject) WorkDone() {
	for counter.calling {
		time.Sleep(ms)
	}
	counter.minus()
}

// Close signal for a dApp
func (d *AppObject) CloseDapp() <-chan struct{} {
	return d.done
}

// Send close signal to all active dApp channels
func (d *AppObject) CloseAllDapps() {
	for counter.calling {
		time.Sleep(ms)
	}
	counter.closing = true
	ch := 0
	for ch < d.channels {
		ch++
		d.done <- struct{}{}
	}

	for counter.active() > 0 {
		time.Sleep(time.Second)
	}
	counter.closing = false
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

// Get the max size for a object while maintaining aspect ratio
func (d *AppObject) GetMaxSize(w, h float32) fyne.Size {
	wRatio := d.Window.Canvas().Size().Width / MIN_WIDTH
	hRatio := d.Window.Canvas().Size().Height / MIN_HEIGHT

	return fyne.NewSize(w*wRatio, h*hRatio)
}

// Add dApp account handler to AppObject
func (d *AppObject) AddAccountHandler(name string, f func(interface{}) error) {
	if d.account.handlers == nil {
		d.account.handlers = make(map[string]func(interface{}) error)
	}

	d.account.handlers[name] = f
}

// Get the current account handlers from AppObject
func (d *AppObject) GetAccountHandlers() map[string]func(interface{}) error {
	return d.account.handlers
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
		logger.Warnf("[%s] %s not found\n", tag, path)

		return false
	}

	return false
}

// Get dero .db file names from wallet directory
func GetDeroAccounts() (prefix string, names []string) {
	prefix = "mainnet"
	if !globals.IsMainnet() {
		prefix = "testnet"
	}

	path := filepath.Join(GetDir(), prefix) + string(filepath.Separator)

	files, err := filepath.Glob(path + "*.db")
	if err != nil {
		logger.Errorln("[dReams]", err)
		return
	}

	for _, f := range files {
		names = append(names, strings.TrimPrefix(f, path))
	}

	return
}

// Download image from URL and return as canvas.Image
func DownloadCanvas(URL, fileName string) (canvas.Image, error) {
	url, err := url.Parse(URL)
	if err != nil {
		return *canvas.NewImageFromImage(nil), err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	response, err := client.Get(url.String())
	if err != nil {
		return *canvas.NewImageFromImage(nil), err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return *canvas.NewImageFromImage(nil), fmt.Errorf("received %d response code", response.StatusCode)
	}

	// if !strings.HasPrefix(response.Header.Get("Content-Type"), "image/") {
	// 	return canvas.NewImageFromImage(nil), fmt.Errorf("%s does not point to an image", URL)
	// }

	var buf bytes.Buffer
	_, err = io.Copy(&buf, response.Body)
	if err != nil {
		return *canvas.NewImageFromImage(nil), err
	}

	return *canvas.NewImageFromReader(&buf, fileName), nil
}

// Download image from URL and return as []byte
func DownloadBytes(URL string) ([]byte, error) {
	url, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}

	client := http.Client{Timeout: 15 * time.Second}
	response, err := client.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download the image: status code %d", response.StatusCode)
	}

	// if !strings.HasPrefix(response.Header.Get("Content-Type"), "image/") {
	// 	return nil, fmt.Errorf("%s does not point to an image", URL)
	// }

	image, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return image, nil
}

// Download image from URL and save as file
func DownloadFile(URL, outPath string) error {
	url, err := url.Parse(URL)
	if err != nil {
		return err
	}

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	client := http.Client{Timeout: 30 * time.Second}
	response, err := client.Get(url.String())
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download the image: status code %d", response.StatusCode)
	}

	_, err = io.Copy(out, response.Body)
	if err != nil {
		return err
	}

	return nil
}

// Unzip a src file into destination
func UnzipFile(src string, destination string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}

	defer func() {
		r.Close()
		os.Remove(src)
	}()

	for _, f := range r.File {
		fpath := filepath.Join(destination, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s is an illegal filepath", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		out, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return filenames, err
		}

	}

	return filenames, nil
}

// Returns Fyne theme icon for name
func FyneIcon(name fyne.ThemeIconName) fyne.Resource {
	return fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), name)
}

// Add a asset option to a AssetSelect
func (a *AssetSelect) Add(add, check string) {
	if check == rpc.Wallet.Address {
		opts := a.Select.Options
		new_opts := append(opts, add)
		a.Select.Options = new_opts
		a.Sort()
		a.Select.Refresh()
	}
}

// Clears all assets from select options
func (a *AssetSelect) ClearAll() {
	a.Select.Options = []string{}
	a.Select.Selected = ""
	a.Select.Refresh()
}

// Sort the select widgets options
func (a *AssetSelect) Sort() {
	sort.Strings(a.Select.Options)
}

// Remove a asset from Select by name
func (a *AssetSelect) RemoveAsset(rm string) {
	index := -1
	for i, item := range a.Select.Options {
		if item == rm {
			index = i
			break
		}
	}

	if index != -1 {
		a.Select.Options = append(a.Select.Options[:index], a.Select.Options[index+1:]...)
	}

	a.Select.Refresh()
}
