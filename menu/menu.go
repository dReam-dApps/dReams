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
	"github.com/deroproject/derohe/globals"
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
	sync.RWMutex
	Daemon    string
	Themes    []string
	Dapps     map[string]bool
	Ratings   map[string]uint64
	Indicator struct {
		Wallet *fyne.Animation
		Daemon *fyne.Animation
		TX     *fyne.Animation
	}
	Check struct {
		Daemon *widget.Check
	}
}

type exiting struct {
	signal bool
	sync.RWMutex
}

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
	dreams.Theme.Name = "Hex"
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
	Control.RLock()
	defer Control.RUnlock()
	for _, b := range Control.Dapps {
		if b {
			enabled++
		}
	}

	return
}

// Returns if a dApp is enabled
func DappEnabled(dapp string) bool {
	Control.RLock()
	defer Control.RUnlock()
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
		gnomon.SetFastsync(true, false, 10000)

		return
	}

	gnomon.SetParallel(1)
	gnomon.SetDBStorageType("boltdb")
	gnomon.SetFastsync(true, false, 10000)

	file, err := os.ReadFile("config/config.json")
	if err != nil {
		logger.Errorln("[ReadDreamsConfig]", err)
		return
	}

	if err = json.Unmarshal(file, &saved); err != nil {
		logger.Errorln("[ReadDreamsConfig]", err)
		return
	}

	bundle.AppColor = saved.Skin

	if saved.Dapps != nil {
		// Old config may have this enabled, set false so we don't block trying to signal this chan
		if saved.Dapps["DerBnb"] {
			saved.Dapps["DerBnb"] = false
		}

		Control.Lock()
		Control.Dapps = saved.Dapps
		Control.Unlock()
	}

	if saved.Theme != "" {
		dreams.Theme.Name = saved.Theme
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

	if saved.FSDiff > 0 {
		gnomon.SetFastsync(true, saved.FSForce, saved.FSDiff)
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

// Returns default background theme resource by dreams.Theme.Name
func DefaultBackgroundResource() *fyne.StaticResource {
	switch dreams.Theme.Name {
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

// App background theme selection object
//   - If image is not present locally, it is downloaded
func ThemeSelect(d *dreams.AppObject) fyne.CanvasObject {
	options := Control.Themes
	icon := AssetIcon(bundle.ResourceMarketCirclePng.StaticContent, "", 60)
	var max *fyne.Container
	dreams.Theme.Select = widget.NewSelect(options, nil)
	dreams.Theme.Select.SetSelected(dreams.Theme.Name)
	dreams.Theme.Select.OnChanged = func(s string) {
		switch dreams.Theme.Select.SelectedIndex() {
		case -1:
			dreams.Theme.Name = "Hex"
		default:
			dreams.Theme.Name = s
		}
		go func() {
			dir := dreams.GetDir()
			check := strings.Trim(s, "0123456789")
			scid := Assets.SCIDs[s]
			if check == "AZYDS" {
				file := dir + "/assets/" + s + "/" + s + ".png"
				if dreams.FileExists(file, "dReams") {
					dreams.Theme.Img = *canvas.NewImageFromFile(file)
				} else {
					dreams.Theme.URL = gnomes.GetAssetUrl(1, scid) // "https://raw.githubusercontent.com/Azylem/" + s + "/main/" + s + ".png"
					logger.Println("[dReams] Downloading", dreams.Theme.URL)
					if img, err := dreams.DownloadCanvas(gnomes.GetAssetUrl(0, scid), s); err == nil {
						dreams.Theme.Img = img
					}
				}
				max.Objects[1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = SwitchProfileIcon("AZY-Deroscapes", s, dreams.Theme.URL, 60)
			} else if check == "SIXART" {
				file := dir + "/assets/" + s + "/" + s + ".png"
				if dreams.FileExists(file, "dReams") {
					dreams.Theme.Img = *canvas.NewImageFromFile(file)
				} else {
					dreams.Theme.URL = gnomes.GetAssetUrl(1, scid) // "https://raw.githubusercontent.com/SixofClubsss/SIXART/main/" + s + "/" + s + ".png"
					logger.Println("[dReams] Downloading", dreams.Theme.URL)
					if img, err := dreams.DownloadCanvas(gnomes.GetAssetUrl(0, scid), s); err == nil {
						dreams.Theme.Img = img
					}
				}
				max.Objects[1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = SwitchProfileIcon("SIXART", s, dreams.Theme.URL, 60)
			} else if check == "HSTheme" {
				file := dir + "/assets/" + s + "/" + s + ".png"
				if dreams.FileExists(file, "dReams") {
					dreams.Theme.Img = *canvas.NewImageFromFile(file)
				} else {
					dreams.Theme.URL = "https://raw.githubusercontent.com/High-Strangeness/High-Strangeness/main/" + s + "/" + s + ".png"
					logger.Println("[dReams] Downloading", dreams.Theme.URL)
					if img, err := dreams.DownloadCanvas(dreams.Theme.URL, s); err == nil {
						dreams.Theme.Img = img
					}
				}
				hs_icon := "https://raw.githubusercontent.com/High-Strangeness/High-Strangeness/main/HighStrangeness-IC.jpg"
				max.Objects[1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = SwitchProfileIcon("High Strangeness", "HighStrangeness1", hs_icon, 60)
			} else if s == "Hex" {
				dreams.Theme.Img = *canvas.NewImageFromResource(bundle.ResourceBackground100Png)
				img := canvas.NewImageFromResource(bundle.ResourceMarketCirclePng)
				img.SetMinSize(fyne.NewSize(60, 60))
				max.Objects[1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = img
			} else if s == "Bullet" {
				dreams.Theme.Img = *canvas.NewImageFromResource(bundle.ResourceBackground110Png)
				img := canvas.NewImageFromResource(bundle.ResourceMarketCirclePng)
				img.SetMinSize(fyne.NewSize(60, 60))
				max.Objects[1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = img
			} else if s == "Highway" {
				dreams.Theme.Img = *canvas.NewImageFromResource(bundle.ResourceBackground111Png)
				img := canvas.NewImageFromResource(bundle.ResourceMarketCirclePng)
				img.SetMinSize(fyne.NewSize(60, 60))
				max.Objects[1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = img
			} else if s == "Glass" {
				dreams.Theme.Img = *canvas.NewImageFromResource(bundle.ResourceBackground112Png)
				img := canvas.NewImageFromResource(bundle.ResourceMarketCirclePng)
				img.SetMinSize(fyne.NewSize(60, 60))
				max.Objects[1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = img
			}
			d.Background.Refresh()
		}()
	}
	dreams.Theme.Select.PlaceHolder = "Theme:"
	max = container.NewBorder(nil, nil, icon, nil, container.NewVBox(dreams.Theme.Select))

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

	done := make(chan struct{})
	confirm_button := widget.NewButtonWithIcon("Confirm", dreams.FyneIcon("confirm"), func() {
		var pos uint64
		if slider.Value > 0 {
			pos = 1
		}

		fee := uint64(math.Abs(slider.Value * 10000))
		tx := rpc.RateSCID(scid, fee, pos)
		go ShowTxDialog("Rate SCID", "RateSCID", tx, 3*time.Second, d.Window)

		done <- struct{}{}
	})
	confirm_button.Importance = widget.HighImportance
	confirm_button.Hide()

	cancel_button := widget.NewButtonWithIcon("Cancel", dreams.FyneIcon("cancel"), func() {
		done <- struct{}{}
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

	confirm := dialog.NewCustom("Rate Contract", "", content, d.Window)
	confirm.SetButtons([]fyne.CanvasObject{buttons})
	go ShowConfirmDialog(done, confirm)
}

// Create and show dialog for sent TX, dismiss copies txid to clipboard, dialog will hide after delay
func ShowTxDialog(title, tag, txid string, delay time.Duration, w fyne.Window) {
	var message string
	var button *widget.Button
	if txid != "" {
		message = fmt.Sprintf("TXID: %s", txid)
		button = widget.NewButton("Copy", func() { w.Clipboard().SetContent(txid) })
		if tag != "" {
			go rpc.ConfirmTx(txid, tag, 45)
		}
	} else {
		message = "TX error, check logs"
		button = widget.NewButton("Ok", nil)
	}

	info := dialog.NewCustom(title, message, container.NewVBox(widget.NewLabel(message)), w)
	info.SetButtons([]fyne.CanvasObject{button})

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
				if !rpc.IsReady() || IsClosing() {
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
		c.SetConfirmText("Confirm")
		c.SetDismissText("Cancel")
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
				if !rpc.IsReady() || IsClosing() {
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

// Create and show dialog for general messages
func ShowMessageDialog(title, message string, delay time.Duration, w fyne.Window) {
	info := dialog.NewInformation(title, message, w)
	info.Show()
	time.Sleep(delay)
	info.Hide()
	info = nil
}

// Send Dero asset menu
//   - Asset SCID can be sent as payload to receiver when sending asset
//   - Pass resources for window_icon
func sendAssetMenu(window_icon fyne.Resource, d *dreams.AppObject) {
	Assets.Button.sending = true
	saw := fyne.CurrentApp().NewWindow("Send Asset")
	saw.Resize(fyne.NewSize(330, 680))
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

	viewing_label := widget.NewLabel(fmt.Sprintf("Sending SCID:\n\n%s\n\nEnter destination address below", viewing_asset))
	viewing_label.Wrapping = fyne.TextWrapWord
	viewing_label.Alignment = fyne.TextAlignCenter

	dest_entry := widget.NewMultiLineEntry()
	dest_entry.SetPlaceHolder("Destination Address:")
	dest_entry.Wrapping = fyne.TextWrapWord
	dest_entry.Validator = func(s string) (err error) {
		if _, err = globals.ParseValidateAddress(s); err != nil {
			addr := rpc.GetNameToAddress(strings.TrimSpace(s))
			if _, err = globals.ParseValidateAddress(addr); err != nil {
				send_button.Hide()
				return
			}
		}

		send_button.Show()

		return
	}

	title_line := canvas.NewLine(bundle.TextColor)
	title := container.NewCenter(container.NewVBox(dwidget.NewCanvasText("Sending Asset", 21, fyne.TextAlignCenter), title_line))

	var dest string
	var confirm_open bool
	send_button = widget.NewButton("Send Asset", func() {
		confirm_open = true
		send_asset := viewing_asset

		confirm_button := widget.NewButtonWithIcon("Confirm", dreams.FyneIcon("confirm"), func() {
			tx := rpc.SendAsset(send_asset, dest)
			Assets.Button.sending = false
			go ShowTxDialog("Send Asset", "SendAsset", tx, 3*time.Second, d.Window)

			saw.Close()
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
		confirm_label := widget.NewLabel(fmt.Sprintf("Sending SCID:\n\n%s\n\nDestination: %s", send_asset, dest))
		confirm_label.Wrapping = fyne.TextWrapWord
		confirm_label.Alignment = fyne.TextAlignCenter

		confirm_display := container.NewVBox(layout.NewSpacer(), confirm_label, layout.NewSpacer())
		confirm_options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
		confirm_content := container.NewBorder(title, confirm_options, nil, nil, confirm_display)
		saw.SetContent(
			container.NewStack(
				BackgroundRast("sendAssetMenu"),
				bundle.Alpha180,
				confirm_content))
	})
	send_button.Importance = widget.HighImportance
	send_button.Hide()

	button_spacer := canvas.NewRectangle(color.Transparent)
	button_spacer.SetMinSize(fyne.NewSize(0, 60))

	saw_content = container.NewVBox(
		title,
		viewing_label,
		layout.NewSpacer(),
		container.NewCenter(Assets.Icon),
		button_spacer,
		container.NewStack(dest_entry),
		button_spacer,
		container.NewStack(button_spacer, container.NewVBox(layout.NewSpacer(), container.NewAdaptiveGrid(2, layout.NewSpacer(), container.NewStack(send_button)))))

	go func() {
		for rpc.IsReady() && Assets.Button.sending && !IsClosing() {
			time.Sleep(time.Second)
			if !confirm_open {
				saw_content.Objects[3].(*fyne.Container).Objects[0] = Assets.Icon
				if viewing_asset != Assets.Viewing {
					viewing_asset = Assets.Viewing
					viewing_label.SetText("Sending SCID:\n\n" + viewing_asset + " \n\nEnter destination address below")
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
func listMenu(window_icon fyne.Resource, d *dreams.AppObject) {
	Assets.Button.listing = true
	aw := fyne.CurrentApp().NewWindow("List NFA")
	aw.Resize(fyne.NewSize(330, 680))
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

	duration := dwidget.NewAmountEntry("", 1, 0)
	duration.AllowFloat = false
	duration.SetPlaceHolder("Duration in Hours:")
	duration.Validator = validation.NewRegexp(`^[^0]\d{0,2}$`, "Int required")

	start := dwidget.NewAmountEntry("", 0.1, 1)
	start.AllowFloat = true
	start.SetPlaceHolder("Start Price:")
	start.Validator = validation.NewRegexp(`^\d{1,}\.\d{1,5}$|^[^0]\d{0,}$`, "Int or float required")

	charAddr := widget.NewMultiLineEntry()
	charAddr.Wrapping = fyne.TextWrapWord
	charAddr.SetPlaceHolder("Charity Donation Address:")
	charAddr.Validator = validation.NewRegexp(`^(dero)\w{62}$`, "Int required")

	charPerc := dwidget.NewAmountEntry("", 1, 0)
	charPerc.AllowFloat = false
	charPerc.SetPlaceHolder("Charity Donation %:")
	charPerc.Validator = validation.NewRegexp(`^\d{1,2}$`, "Int required")

	duration.OnChanged = func(s string) {
		if rpc.StringToInt(duration.Text) > 168 {
			duration.SetText("168")
		}
	}

	title_line := canvas.NewLine(bundle.TextColor)
	title := container.NewCenter(container.NewVBox(dwidget.NewCanvasText("List Asset", 21, fyne.TextAlignCenter), title_line))

	var confirm_open bool
	set_button = widget.NewButton("Set Listing", func() {
		if listing.Selected != "" {
			confirm_open = true
			listing_asset := viewing_asset
			artP, royaltyP := GetListingPercents(listing_asset)

			dur := rpc.StringToUint64(duration.Text)
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
				tx := rpc.SetNFAListing(listing_asset, listing.Selected, charAddr.Text, dur, s, cp)
				Assets.Button.listing = false
				if rpc.Wallet.IsConnected() {
					if isNFA(Assets.Viewing) {
						Assets.Button.Send.Show()
						Assets.Button.List.Show()
					}
				}
				go ShowTxDialog(fmt.Sprintf("NFA %s", listing.Selected), "SetNFAListing", tx, 3*time.Second, d.Window)

				aw.Close()
			})
			confirm_button.Importance = widget.HighImportance

			confirm_options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
			confirm_content := container.NewBorder(
				title,
				confirm_options,
				nil,
				nil,
				container.NewVBox(layout.NewSpacer(), confirm_label, layout.NewSpacer()))

			aw.SetContent(
				container.NewStack(
					BackgroundRast("listMenu"),
					bundle.Alpha180,
					confirm_content))
		}
	})
	set_button.Importance = widget.HighImportance
	set_button.Hide()

	button_spacer := canvas.NewRectangle(color.Transparent)
	button_spacer.SetMinSize(fyne.NewSize(0, 36))

	go func() {
		for rpc.IsReady() && Assets.Button.listing && !IsClosing() {
			time.Sleep(time.Second)
			if !confirm_open && isNFA(Assets.Viewing) {
				aw_content.Objects[3].(*fyne.Container).Objects[0] = Assets.Icon
				if viewing_asset != Assets.Viewing {
					viewing_asset = Assets.Viewing
					viewing_label.SetText(fmt.Sprintf("Listing SCID:\n\n%s", viewing_asset))
				}
			}

			if listing.Selected != "" && duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
				set_button.Show()
			} else {
				set_button.Hide()
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
		title,
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
	if dreams.Theme.Img.Resource != nil {
		if img, _, err = image.Decode(bytes.NewReader(dreams.Theme.Img.Resource.Content())); err == nil {
			return canvas.NewRasterFromImage(img)
		}

		if img, _, err = image.Decode(bytes.NewReader(DefaultBackgroundResource().StaticContent)); err == nil {
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
		smw.Resize(fyne.NewSize(330, 680))
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

		ringsize := widget.NewSelect([]string{"2", "16", "32", "64", "128"}, func(s string) {})
		ringsize.PlaceHolder = "Ringsize:"
		ringsize.SetSelectedIndex(1)

		message_entry := widget.NewMultiLineEntry()
		message_entry.SetPlaceHolder("Message:")
		message_entry.Wrapping = fyne.TextWrapWord

		dest_entry := widget.NewMultiLineEntry()
		dest_entry.SetPlaceHolder("Destination Address:")
		dest_entry.Wrapping = fyne.TextWrapWord
		dest_entry.Validator = func(s string) (err error) {
			if _, err = globals.ParseValidateAddress(s); err != nil {
				addr := rpc.GetNameToAddress(strings.TrimSpace(s))
				if _, err = globals.ParseValidateAddress(addr); err != nil {
					send_button.Hide()
					return
				}
			}

			if message_entry.Text != "" {
				send_button.Show()
			} else {
				send_button.Hide()
			}

			return
		}

		message_entry.OnChanged = func(s string) {
			dest_entry.Validate()
		}

		send_button = widget.NewButton("Send Message", func() {
			if message_entry.Text != "" {
				rings := rpc.StringToUint64(ringsize.Selected)
				go rpc.SendMessage(dest_entry.Text, message_entry.Text, rings)
				Assets.Button.messaging = false
				smw.Close()
			}
		})
		send_button.Importance = widget.HighImportance
		send_button.Hide()

		dest_cont := container.NewCenter(container.NewVBox(label, container.NewCenter(ringsize), dest_entry, dwidget.NewSpacer(300, 0)))
		message_cont := container.NewBorder(nil, container.NewStack(send_button), nil, nil, message_entry)

		content := container.NewVSplit(container.NewStack(dest_cont), message_cont)

		go func() {
			for rpc.IsReady() && Assets.Button.messaging && !IsClosing() {
				time.Sleep(time.Second)
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
