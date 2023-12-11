package menu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/civilware/Gnomon/structures"
	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/gnomes"
	"github.com/dReam-dApps/dReams/rpc"
	"github.com/sirupsen/logrus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type menuObjects struct {
	sync.Mutex
	Daemon    string
	Themes    []string
	Dapps     map[string]bool
	Ratings   map[string]uint64
	Indicator struct {
		Wallet *fyne.Animation
		Daemon *fyne.Animation
	}
	Check struct {
		Daemon *widget.Check
	}
}

type exiting struct {
	signal bool
	sync.RWMutex
}

// Background theme AssetSelect
var Theme dreams.AssetSelect

// Control menu indicators, checks, maps and defaults
var Control menuObjects

// Log output with logrus matching Gnomon
var logger = structures.Logger.WithFields(logrus.Fields{})

// Gnomon instance for menu
var gnomon = gnomes.NewGnomes()

// Exit from menu routines
var exit exiting

// Initialize maps and defaults
func init() {
	Theme.Name = "Hex"
	Assets.SCIDs = make(map[string]string)
	Assets.Enabled = make(map[string]bool)
	Control.Dapps = make(map[string]bool)
	Control.Ratings = make(map[string]uint64)
	Control.Themes = []string{"Hex", "Bullet", "Highway", "Glass"}
}

// Check if menu is calling to close
func IsClosing() (close bool) {
	exit.RLock()
	close = exit.signal
	exit.RUnlock()

	return
}

// Set menu closing bool value
func SetClose(value bool) {
	exit.Lock()
	exit.signal = value
	exit.Unlock()
}

// Returns how many dApps are enabled
func EnabledDappCount() (enabled int) {
	for _, b := range Control.Dapps {
		if b {
			enabled++
		}
	}

	return
}

// Returns if a dApp is enabled
func DappEnabled(dapp string) bool {
	if b, ok := Control.Dapps[dapp]; ok && b {
		return true
	}

	return false
}

// Save dReams config.json file for platform wide dApp use
func WriteDreamsConfig(u dreams.SaveData) {
	if u.Daemon != nil && u.Daemon[0] == "" {
		if Control.Daemon != "" {
			u.Daemon[0] = Control.Daemon
		} else {
			u.Daemon[0] = "127.0.0.1:10102"
		}
	}

	file, err := os.Create("config/config.json")
	if err != nil {
		logger.Errorln("[WriteDreamsConfig]", err)
		return
	}
	defer file.Close()

	json, _ := json.MarshalIndent(u, "", " ")
	if _, err = file.Write(json); err != nil {
		logger.Errorln("[WriteDreamsConfig]", err)
	}
}

// Read dReams platform config.json file
//   - tag for log print
//   - Sets up directory if none exists
func ReadDreamsConfig(tag string) (saved dreams.SaveData) {
	if !dreams.FileExists("config/config.json", tag) {
		logger.Printf("[%s] Creating config directory\n", tag)
		mkdir := os.Mkdir("config", 0755)
		if mkdir != nil {
			logger.Errorf("[%s] %s\n", tag, mkdir)
		}

		if config, err := os.Create("config/config.json"); err == nil {
			var save dreams.SaveData
			json, _ := json.MarshalIndent(&save, "", " ")
			if _, err = config.Write(json); err != nil {
				logger.Errorln("[WriteDreamsConfig]", err)
			}
			config.Close()
		}

		gnomon.SetParallel(1)
		gnomon.SetDBStorageType("boltdb")

		return
	}

	gnomon.SetParallel(1)
	gnomon.SetDBStorageType("boltdb")

	file, err := os.ReadFile("config/config.json")
	if err != nil {
		logger.Errorln("[ReadDreamsConfig]", err)
		return
	}

	if err = json.Unmarshal(file, &saved); err != nil {
		logger.Errorln("[ReadDreamsConfig]", err)
		return
	}

	// Old config may have this enabled, set false so we don't block trying to signal this chan
	if saved.Dapps["DerBnb"] {
		saved.Dapps["DerBnb"] = false
	}

	bundle.AppColor = saved.Skin
	Control.Dapps = saved.Dapps

	if saved.Theme != "" {
		Theme.Name = saved.Theme
	}

	if saved.Assets != nil {
		Assets.Enabled = saved.Assets
	}

	if saved.DBtype == "gravdb" {
		gnomon.SetDBStorageType(saved.DBtype)
	}

	if saved.Para > 0 && saved.Para < 6 {
		gnomon.SetParallel(saved.Para)
	}

	return
}

func SwitchProfileIcon(collection, name, url string, size float32) (icon *canvas.Image) {
	have, err := gnomes.StorageExists(collection, name)
	if err != nil {
		have = false
		logger.Errorln("[SwitchProfileIcon]", err)
	}

	if have {
		var new Asset
		gnomes.GetStorage(collection, name, &new)
		if new.Image != nil && !bytes.Equal(new.Image, bundle.ResourceMarketCirclePng.StaticContent) {
			icon = canvas.NewImageFromReader(bytes.NewReader(new.Image), name)
			icon.SetMinSize(fyne.NewSize(size, size))
			return
		} else {
			have = false
		}
	}

	if !have {
		if img, err := dreams.DownloadBytes(url); err == nil {
			icon = canvas.NewImageFromReader(bytes.NewReader(img), name)
			icon.SetMinSize(fyne.NewSize(60, 60))
			return
		} else {
			logger.Errorln("[SwitchProfileIcon]", err)
		}
	}

	icon = canvas.NewImageFromResource(bundle.ResourceFigure1CirclePng)
	icon.SetMinSize(fyne.NewSize(60, 60))
	return
}

// Returns default theme resource by Theme.Name
func DefaultThemeResource() *fyne.StaticResource {
	switch Theme.Name {
	case "Hex":
		return bundle.ResourceBackground100Png
	case "Bullet":
		return bundle.ResourceBackground110Png
	case "Highway":
		return bundle.ResourceBackground111Png
	case "Glass":
		return bundle.ResourceBackground112Png
	default:
		return bundle.ResourceBackground100Png
	}
}

// App theme selection object
//   - If image is not present locally, it is downloaded
func ThemeSelect(d *dreams.AppObject) fyne.CanvasObject {
	options := Control.Themes
	icon := AssetIcon(bundle.ResourceMarketCirclePng.StaticContent, "", 60)
	var max *fyne.Container
	Theme.Select = widget.NewSelect(options, nil)
	Theme.Select.SetSelected(Theme.Name)
	Theme.Select.OnChanged = func(s string) {
		switch Theme.Select.SelectedIndex() {
		case -1:
			Theme.Name = "Hex"
		default:
			Theme.Name = s
		}
		go func() {
			dir := dreams.GetDir()
			check := strings.Trim(s, "0123456789")
			scid := Assets.SCIDs[s]
			if check == "AZYDS" {
				file := dir + "/assets/" + s + "/" + s + ".png"
				if dreams.FileExists(file, "dReams") {
					Theme.Img = *canvas.NewImageFromFile(file)
				} else {
					Theme.URL = gnomes.GetAssetUrl(1, scid) // "https://raw.githubusercontent.com/Azylem/" + s + "/main/" + s + ".png"
					logger.Println("[dReams] Downloading", Theme.URL)
					if img, err := dreams.DownloadCanvas(gnomes.GetAssetUrl(0, scid), s); err == nil {
						Theme.Img = img
					}
				}
				max.Objects[1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = SwitchProfileIcon("AZY-Deroscapes", s, Theme.URL, 60)
			} else if check == "SIXART" {
				file := dir + "/assets/" + s + "/" + s + ".png"
				if dreams.FileExists(file, "dReams") {
					Theme.Img = *canvas.NewImageFromFile(file)
				} else {
					Theme.URL = gnomes.GetAssetUrl(1, scid) // "https://raw.githubusercontent.com/SixofClubsss/SIXART/main/" + s + "/" + s + ".png"
					logger.Println("[dReams] Downloading", Theme.URL)
					if img, err := dreams.DownloadCanvas(gnomes.GetAssetUrl(0, scid), s); err == nil {
						Theme.Img = img
					}
				}
				max.Objects[1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = SwitchProfileIcon("SIXART", s, Theme.URL, 60)
			} else if check == "HSTheme" {
				file := dir + "/assets/" + s + "/" + s + ".png"
				if dreams.FileExists(file, "dReams") {
					Theme.Img = *canvas.NewImageFromFile(file)
				} else {
					Theme.URL = "https://raw.githubusercontent.com/High-Strangeness/High-Strangeness/main/" + s + "/" + s + ".png"
					logger.Println("[dReams] Downloading", Theme.URL)
					if img, err := dreams.DownloadCanvas(Theme.URL, s); err == nil {
						Theme.Img = img
					}
				}
				hs_icon := "https://raw.githubusercontent.com/High-Strangeness/High-Strangeness/main/HighStrangeness-IC.jpg"
				max.Objects[1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = SwitchProfileIcon("High Strangeness", "HighStrangeness1", hs_icon, 60)
			} else if s == "Hex" {
				Theme.Img = *canvas.NewImageFromResource(bundle.ResourceBackground100Png)
				img := canvas.NewImageFromResource(bundle.ResourceMarketCirclePng)
				img.SetMinSize(fyne.NewSize(60, 60))
				max.Objects[1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = img
			} else if s == "Bullet" {
				Theme.Img = *canvas.NewImageFromResource(bundle.ResourceBackground110Png)
				img := canvas.NewImageFromResource(bundle.ResourceMarketCirclePng)
				img.SetMinSize(fyne.NewSize(60, 60))
				max.Objects[1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = img
			} else if s == "Highway" {
				Theme.Img = *canvas.NewImageFromResource(bundle.ResourceBackground111Png)
				img := canvas.NewImageFromResource(bundle.ResourceMarketCirclePng)
				img.SetMinSize(fyne.NewSize(60, 60))
				max.Objects[1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = img
			} else if s == "Glass" {
				Theme.Img = *canvas.NewImageFromResource(bundle.ResourceBackground112Png)
				img := canvas.NewImageFromResource(bundle.ResourceMarketCirclePng)
				img.SetMinSize(fyne.NewSize(60, 60))
				max.Objects[1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = img
			}
			d.Background.Refresh()
		}()
	}
	Theme.Select.PlaceHolder = "Theme:"
	max = container.NewBorder(nil, nil, icon, nil, container.NewVBox(Theme.Select))

	return max
}

// Display SCID rating from dReams SCID rating system
func DisplayRating(i uint64) fyne.Resource {
	if i > 250000 {
		return bundle.ResourceBlueBadge3Png
	} else if i > 150000 {
		return bundle.ResourceBlueBadge2Png
	} else if i > 90000 {
		return bundle.ResourceBlueBadgePng
	} else if i > 50000 {
		return bundle.ResourceRedBadgePng
	} else {
		return nil
	}
}

// Confirmation for a SCID rating
func RateConfirm(scid string, d *dreams.AppObject) {
	label := widget.NewLabel(fmt.Sprintf("Rate your experience with this contract\n\n%s", scid))
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	rating_label := widget.NewLabel("Pick a rating")
	rating_label.Wrapping = fyne.TextWrapWord
	rating_label.Alignment = fyne.TextAlignCenter

	fee_label := widget.NewLabel("")
	fee_label.Wrapping = fyne.TextWrapWord
	fee_label.Alignment = fyne.TextAlignCenter

	var slider *widget.Slider
	var confirm *dialog.CustomDialog

	confirm_button := widget.NewButtonWithIcon("Confirm", dreams.FyneIcon("confirm"), func() {
		var pos uint64
		if slider.Value > 0 {
			pos = 1
		}

		fee := uint64(math.Abs(slider.Value * 10000))
		rpc.RateSCID(scid, fee, pos)
		confirm.Hide()
		confirm = nil
	})
	confirm_button.Importance = widget.HighImportance
	confirm_button.Hide()

	cancel_button := widget.NewButtonWithIcon("Cancel", dreams.FyneIcon("cancel"), func() {
		confirm.Hide()
		confirm = nil
	})

	slider = widget.NewSlider(-5, 5)
	slider.Step = 0.5
	slider.OnChanged = func(f float64) {
		if slider.Value != 0 {
			rating_label.SetText(fmt.Sprintf("Rating: %.0f", f*10000))
			fee_label.SetText(fmt.Sprintf("Fee: %.5f Dero", math.Abs(f)/10))
			confirm_button.Show()
		} else {
			rating_label.SetText("Pick a rating")
			fee_label.SetText("")
			confirm_button.Hide()
		}
	}

	good := canvas.NewImageFromResource(bundle.ResourceBlueBadge3Png)
	good.SetMinSize(fyne.NewSize(30, 30))
	bad := canvas.NewImageFromResource(bundle.ResourceRedBadgePng)
	bad.SetMinSize(fyne.NewSize(30, 30))

	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(400, 100))

	rate_cont := container.NewBorder(spacer, nil, bad, good, slider)

	left := container.NewVBox(confirm_button)
	right := container.NewVBox(cancel_button)
	buttons := container.NewAdaptiveGrid(2, left, right)

	content := container.NewVBox(layout.NewSpacer(), label, rating_label, fee_label, rate_cont, layout.NewSpacer())

	confirm = dialog.NewCustom("Rate Contract", "", content, d.Window)
	confirm.SetButtons([]fyne.CanvasObject{buttons})
	confirm.Show()

	go func() {
		for rpc.IsReady() {
			time.Sleep(time.Second)
		}

		if confirm != nil {
			confirm.Hide()
			confirm = nil
		}
	}()
}

// Create and show dialog for sent TX, dismiss copies txid to clipboard, dialog will hide after delay
func ShowTxDialog(title, message, txid string, delay time.Duration, w fyne.Window) {
	info := dialog.NewInformation(title, message, w)
	info.SetDismissText("Copy")
	info.SetOnClosed(func() {
		w.Clipboard().SetContent(txid)
	})
	info.Show()
	time.Sleep(delay)
	info.Hide()
	info = nil
}

// Shows a passed Fyne CustomDialog or ConfirmDialog and closes it if connection is lost
func ShowConfirmDialog(done chan struct{}, confirm interface{}) {
	switch c := confirm.(type) {
	case *dialog.CustomDialog:
		c.Show()
		for {
			select {
			case <-done:
				if c != nil {
					c.Hide()
					c = nil
				}
				return
			default:
				if !rpc.IsReady() {
					if c != nil {
						c.Hide()
						c = nil
					}
					return
				}
				time.Sleep(time.Second)
			}
		}
	case *dialog.ConfirmDialog:
		c.Show()
		for {
			select {
			case <-done:
				if c != nil {
					c.Hide()
					c = nil
				}
				return
			default:
				if !rpc.IsReady() {
					if c != nil {
						c.Hide()
						c = nil
					}
					return
				}
				time.Sleep(time.Second)
			}
		}
	default:
		// Nothing
	}
}

// Send Dero asset menu
//   - Asset SCID can be sent as payload to receiver when sending asset
//   - Pass resources for window_icon
func sendAssetMenu(window_icon fyne.Resource) {
	Assets.Button.sending = true
	saw := fyne.CurrentApp().NewWindow("Send Asset")
	saw.Resize(fyne.NewSize(330, 700))
	saw.SetIcon(window_icon)
	Assets.Button.Send.Hide()
	Assets.Button.List.Hide()
	saw.SetCloseIntercept(func() {
		Assets.Button.sending = false
		if rpc.Wallet.IsConnected() {
			if isNFA(Assets.Viewing) {
				Assets.Button.Send.Show()
				Assets.Button.List.Show()
			}
		}
		saw.Close()
	})
	saw.SetFixedSize(true)

	var saw_content *fyne.Container
	var send_button *widget.Button

	viewing_asset := Assets.Viewing

	viewing_label := widget.NewLabel(fmt.Sprintf("Sending SCID:\n\n%s\n\nEnter destination address below\n\nSCID can be sent to receiver as payload\n\n", viewing_asset))
	viewing_label.Wrapping = fyne.TextWrapWord
	viewing_label.Alignment = fyne.TextAlignCenter

	info_label := widget.NewLabel("Enter all info before sending")
	payload := widget.NewCheck("Send SCID as payload", func(b bool) {})

	dest_entry := widget.NewMultiLineEntry()
	dest_entry.SetPlaceHolder("Destination Address:")
	dest_entry.Wrapping = fyne.TextWrapWord
	dest_entry.Validator = validation.NewRegexp(`^(dero)\w{62}$`, "Invalid Address")
	dest_entry.OnChanged = func(s string) {
		if dest_entry.Validate() == nil {
			info_label.SetText("")
			send_button.Show()
		} else {
			info_label.SetText("Enter destination address")
			send_button.Hide()
		}
	}

	entry_clear := widget.NewButtonWithIcon("", dreams.FyneIcon("contentUndo"), func() {
		dest_entry.SetText("")
	})
	entry_clear.Importance = widget.LowImportance

	var dest string
	var confirm_open bool
	send_button = widget.NewButton("Send Asset", func() {
		if dest_entry.Validate() == nil {
			confirm_open = true
			send_asset := viewing_asset
			var load bool
			if payload.Checked {
				load = true
			}

			confirm_button := widget.NewButtonWithIcon("Confirm", dreams.FyneIcon("confirm"), func() {
				if dest_entry.Validate() == nil {
					var load bool
					if payload.Checked {
						load = true
					}
					go rpc.SendAsset(send_asset, dest, load)
					Assets.Button.sending = false
					saw.Close()
				}
			})
			confirm_button.Importance = widget.HighImportance

			cancel_button := widget.NewButtonWithIcon("Cancel", dreams.FyneIcon("cancel"), func() {
				confirm_open = false
				saw.SetContent(
					container.NewStack(
						BackgroundRast("sendAssetMenu"),
						bundle.Alpha180,
						saw_content))
			})

			dest = dest_entry.Text
			confirm_label := widget.NewLabel(fmt.Sprintf("Sending SCID:\n\n%s\n\nDestination: %s\n\nSending SCID as payload: %t", send_asset, dest, load))
			confirm_label.Wrapping = fyne.TextWrapWord
			confirm_label.Alignment = fyne.TextAlignCenter

			confirm_display := container.NewVBox(confirm_label, layout.NewSpacer())
			confirm_options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
			confirm_content := container.NewBorder(nil, confirm_options, nil, nil, confirm_display)
			saw.SetContent(
				container.NewStack(
					BackgroundRast("sendAssetMenu"),
					bundle.Alpha180,
					confirm_content))
		}
	})
	send_button.Importance = widget.HighImportance
	send_button.Hide()

	button_spacer := canvas.NewRectangle(color.Transparent)
	button_spacer.SetMinSize(fyne.NewSize(0, 90))

	saw_content = container.NewVBox(
		viewing_label,
		layout.NewSpacer(),
		container.NewCenter(Assets.Icon),
		button_spacer,
		container.NewStack(
			dest_entry,
			container.NewVBox(
				layout.NewSpacer(),
				container.NewHBox(layout.NewSpacer(), container.NewBorder(nil, layout.NewSpacer(), nil, layout.NewSpacer(), entry_clear)))),
		container.NewCenter(payload),
		container.NewStack(button_spacer, container.NewVBox(layout.NewSpacer(), container.NewAdaptiveGrid(2, layout.NewSpacer(), container.NewStack(send_button)))))

	go func() {
		for rpc.IsReady() && Assets.Button.sending {
			time.Sleep(time.Second)
			if !confirm_open {
				saw_content.Objects[2].(*fyne.Container).Objects[0] = Assets.Icon
				if viewing_asset != Assets.Viewing {
					viewing_asset = Assets.Viewing
					viewing_label.SetText("Sending SCID:\n\n" + viewing_asset + " \n\nEnter destination address below\n\nSCID can be sent to receiver as payload\n\n")
				}

			}
		}
		Assets.Button.sending = false
		saw.Close()
	}()

	saw.SetContent(
		container.NewStack(
			BackgroundRast("sendAssetMenu"),
			bundle.Alpha180,
			saw_content))
	saw.Show()
}

// NFA listing menu
//   - Pass resources for menu window to match main
func listMenu(window_icon fyne.Resource) {
	Assets.Button.listing = true
	aw := fyne.CurrentApp().NewWindow("List NFA")
	aw.Resize(fyne.NewSize(330, 700))
	aw.SetIcon(window_icon)
	Assets.Button.List.Hide()
	Assets.Button.Send.Hide()
	aw.SetCloseIntercept(func() {
		Assets.Button.listing = false
		if rpc.Wallet.IsConnected() {
			if isNFA(Assets.Viewing) {
				Assets.Button.Send.Show()
				Assets.Button.List.Show()
			}
		}
		aw.Close()
	})
	aw.SetFixedSize(true)

	var aw_content *fyne.Container
	var set_button *widget.Button

	viewing_asset := Assets.Viewing
	viewing_label := widget.NewLabel(fmt.Sprintf("Listing SCID:\n\n%s", viewing_asset))
	viewing_label.Wrapping = fyne.TextWrapWord
	viewing_label.Alignment = fyne.TextAlignCenter

	fee_label := widget.NewLabel(fmt.Sprintf("Listing fee %.5f Dero", float64(rpc.ListingFee)/100000))

	listing_options := []string{"Auction", "Sale"}
	listing := widget.NewSelect(listing_options, nil)
	listing.PlaceHolder = "Type:"

	duration := dwidget.NewDeroEntry("", 1, 0)
	duration.AllowFloat = false
	duration.SetPlaceHolder("Duration in Hours:")
	duration.Validator = validation.NewRegexp(`^[^0]\d{0,2}$`, "Int required")

	start := dwidget.NewDeroEntry("", 0.1, 1)
	start.AllowFloat = true
	start.SetPlaceHolder("Start Price:")
	start.Validator = validation.NewRegexp(`^\d{1,}\.\d{1,5}$|^[^0]\d{0,}$`, "Int or float required")

	charAddr := widget.NewMultiLineEntry()
	charAddr.Wrapping = fyne.TextWrapWord
	charAddr.SetPlaceHolder("Charity Donation Address:")
	charAddr.Validator = validation.NewRegexp(`^(dero)\w{62}$`, "Int required")

	charPerc := dwidget.NewDeroEntry("", 1, 0)
	charPerc.AllowFloat = false
	charPerc.SetPlaceHolder("Charity Donation %:")
	charPerc.Validator = validation.NewRegexp(`^\d{1,2}$`, "Int required")
	charPerc.OnChanged = func(s string) {
		if listing.Selected != "" && duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
			set_button.Show()
		} else {
			set_button.Hide()
		}
	}

	duration.OnChanged = func(s string) {
		if rpc.StringToInt(s) > 168 {
			duration.SetText("168")
		}

		if listing.Selected != "" && duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
			set_button.Show()
		} else {
			set_button.Hide()
		}
	}

	start.OnChanged = func(s string) {
		if listing.Selected != "" && duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
			set_button.Show()
		} else {
			set_button.Hide()
		}
	}

	charAddr.OnChanged = func(s string) {
		if listing.Selected != "" && duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
			set_button.Show()
		} else {
			set_button.Hide()
		}
	}

	listing.OnChanged = func(s string) {
		if listing.Selected != "" && duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
			set_button.Show()
		} else {
			set_button.Hide()
		}
	}

	var confirm_open bool
	set_button = widget.NewButton("Set Listing", func() {
		if duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
			if listing.Selected != "" {
				confirm_open = true
				listing_asset := viewing_asset
				artP, royaltyP := GetListingPercents(listing_asset)

				d := rpc.StringToUint64(duration.Text)
				s := rpc.ToAtomic(start.Text, 5)
				sp := float64(s) / 100000
				cp := rpc.StringToUint64(charPerc.Text)

				art_gets := (float64(s) * artP) / 100000
				royalty_gets := (float64(s) * royaltyP) / 100000
				char_gets := float64(s) * (float64(cp) / 100) / 100000

				total := sp - art_gets - royalty_gets - char_gets

				first_line := fmt.Sprintf("Listing SCID:\n\n%s\n\nList Type: %s\n\nDuration: %s Hours\n\nStart Price: %0.5f Dero\n\n", listing_asset, listing.Selected, duration.Text, sp)
				second_line := fmt.Sprintf("Artificer Fee: %.0f%s - %0.5f Dero\n\nRoyalties: %.0f%s - %0.5f Dero\n\n", artP*100, "%", art_gets, royaltyP*100, "%", royalty_gets)
				third_line := fmt.Sprintf("Charity Address: %s\n\nCharity Percent: %s%s - %0.5f Dero\n\nYou will receive %.5f Dero if asset sells at start price", charAddr.Text, charPerc.Text, "%", char_gets, total)

				confirm_label := widget.NewLabel(first_line + second_line + third_line)
				confirm_label.Wrapping = fyne.TextWrapWord
				confirm_label.Alignment = fyne.TextAlignCenter

				cancel_button := widget.NewButtonWithIcon("Cancel", dreams.FyneIcon("cancel"), func() {
					confirm_open = false
					aw.SetContent(
						container.NewStack(
							BackgroundRast("listMenu"),
							bundle.Alpha180,
							aw_content))
				})

				confirm_button := widget.NewButtonWithIcon("Confirm", dreams.FyneIcon("confirm"), func() {
					rpc.SetNFAListing(listing_asset, listing.Selected, charAddr.Text, d, s, cp)
					Assets.Button.listing = false
					if rpc.Wallet.IsConnected() {
						if isNFA(Assets.Viewing) {
							Assets.Button.Send.Show()
							Assets.Button.List.Show()
						}
					}
					aw.Close()
				})
				confirm_button.Importance = widget.HighImportance

				confirm_options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
				confirm_content := container.NewBorder(nil, confirm_options, nil, nil, confirm_label)

				aw.SetContent(
					container.NewStack(
						BackgroundRast("listMenu"),
						bundle.Alpha180,
						confirm_content))
			}
		}
	})
	set_button.Importance = widget.HighImportance
	set_button.Hide()

	button_spacer := canvas.NewRectangle(color.Transparent)
	button_spacer.SetMinSize(fyne.NewSize(0, 36))

	go func() {
		for rpc.IsReady() && Assets.Button.listing {
			time.Sleep(time.Second)
			if !confirm_open && isNFA(Assets.Viewing) {
				aw_content.Objects[2].(*fyne.Container).Objects[0] = Assets.Icon
				if viewing_asset != Assets.Viewing {
					viewing_asset = Assets.Viewing
					viewing_label.SetText(fmt.Sprintf("Listing SCID:\n\n%s", viewing_asset))
				}
			}
		}
		Assets.Button.listing = false
		aw.Close()
	}()

	charAddr.Disable()
	charPerc.Disable()
	charAddr.SetText(rpc.Wallet.Address)
	charPerc.SetText("0")

	enable_donations := widget.NewCheck("Enable Donations", func(b bool) {
		if b {
			charAddr.Enable()
			charPerc.Enable()
			charAddr.SetText("")
			charPerc.SetText("")
		} else {
			charAddr.Disable()
			charPerc.Disable()
			charAddr.SetText(rpc.Wallet.Address)
			charPerc.SetText("0")
		}
	})

	aw_content = container.NewVBox(
		viewing_label,
		layout.NewSpacer(),
		container.NewCenter(Assets.Icon),
		button_spacer,
		listing,
		duration,
		start,
		container.NewCenter(enable_donations),
		charAddr,
		charPerc,
		container.NewCenter(fee_label),
		container.NewStack(button_spacer, container.NewVBox(layout.NewSpacer(), container.NewAdaptiveGrid(2, layout.NewSpacer(), container.NewStack(set_button)))))

	aw.SetContent(
		container.NewStack(
			BackgroundRast("listMenu"),
			bundle.Alpha180,
			aw_content))
	aw.Show()
}

// Create a new raster from image, looking for dreams.Theme.Img.Resource
// and will fallback to bundle.ResourceBackgroundPng if err
func BackgroundRast(tag string) *canvas.Raster {
	var err error
	var img image.Image
	if Theme.Img.Resource != nil {
		if img, _, err = image.Decode(bytes.NewReader(Theme.Img.Resource.Content())); err == nil {
			return canvas.NewRasterFromImage(img)
		}

		if img, _, err = image.Decode(bytes.NewReader(DefaultThemeResource().StaticContent)); err == nil {
			return canvas.NewRasterFromImage(img)
		}

		logger.Warnf("[%s] Fallback %s\n", tag, err)
	}

	return canvas.NewRasterFromImage(image.Rect(2, 2, 4, 4))

}

// Send Dero message menu
func SendMessageMenu(dest string, window_icon fyne.Resource) {
	if !Assets.Button.messaging && rpc.Wallet.IsConnected() {
		Assets.Button.messaging = true
		smw := fyne.CurrentApp().NewWindow("Send Message")
		smw.Resize(fyne.NewSize(330, 700))
		smw.SetIcon(window_icon)
		smw.SetCloseIntercept(func() {
			Assets.Button.messaging = false
			smw.Close()
		})
		smw.SetFixedSize(true)

		var send_button *widget.Button

		label := widget.NewLabel("Sending Message:\n\nEnter ringsize and destination address below")
		label.Wrapping = fyne.TextWrapWord
		label.Alignment = fyne.TextAlignCenter

		ringsize := widget.NewSelect([]string{"2", "16", "32", "64"}, func(s string) {})
		ringsize.PlaceHolder = "Ringsize:"
		ringsize.SetSelectedIndex(1)

		message_entry := widget.NewMultiLineEntry()
		message_entry.SetPlaceHolder("Message:")
		message_entry.Wrapping = fyne.TextWrapWord

		dest_entry := widget.NewMultiLineEntry()
		dest_entry.SetPlaceHolder("Destination Address:")
		dest_entry.Wrapping = fyne.TextWrapWord
		dest_entry.Validator = validation.NewRegexp(`^(dero)\w{62}$`, "Invalid Address")
		dest_entry.OnChanged = func(s string) {
			if dest_entry.Validate() == nil && message_entry.Text != "" {
				send_button.Show()
			} else {
				send_button.Hide()
			}
		}

		message_entry.OnChanged = func(s string) {
			if s != "" && dest_entry.Validate() == nil {
				send_button.Show()
			} else {
				send_button.Hide()
			}
		}

		send_button = widget.NewButton("Send Message", func() {
			if dest_entry.Validate() == nil && message_entry.Text != "" {
				rings := rpc.StringToUint64(ringsize.Selected)
				go rpc.SendMessage(dest_entry.Text, message_entry.Text, rings)
				Assets.Button.messaging = false
				smw.Close()
			}
		})
		send_button.Hide()

		dest_cont := container.NewVBox(label, ringsize, dest_entry)
		message_cont := container.NewBorder(nil, send_button, nil, nil, message_entry)

		content := container.NewVSplit(dest_cont, message_cont)

		go func() {
			for rpc.IsReady() && Assets.Button.messaging {
				time.Sleep(2 * time.Second)
			}
			Assets.Button.messaging = false
			smw.Close()
		}()

		if dest != "" {
			dest_entry.SetText(dest)
		}

		smw.SetContent(
			container.NewStack(
				BackgroundRast("SendMessageMenu"),
				bundle.Alpha180,
				content))
		smw.Show()
	}
}
