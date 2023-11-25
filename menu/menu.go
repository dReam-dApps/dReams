package menu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type menuObjects struct {
	G45_count       int
	NFA_count       int
	List_open       bool
	send_open       bool
	msg_open        bool
	once            bool
	Daemon_config   string
	Viewing_asset   string
	Dapp_list       map[string]bool
	Enabled_assets  map[string]bool
	Contract_rating map[string]uint64
	Names           *widget.Select
	Send_asset      *widget.Button
	Claim_button    *widget.Button
	List_button     *widget.Button
	Daemon_check    *widget.Check
	Wallet_ind      *fyne.Animation
	Daemon_ind      *fyne.Animation
	sync.Mutex
}
type exit struct {
	Signal bool
	sync.RWMutex
}

var Control menuObjects
var Exit exit

// Check for app closing signal
func ClosingApps() (close bool) {
	Exit.RLock()
	close = Exit.Signal
	Exit.RUnlock()

	return
}

// Set app closing signal value
func CloseAppSignal(value bool) {
	Exit.Lock()
	Exit.Signal = value
	Exit.Unlock()
}

// Returns how many dApps are enabled
func EnabledDappCount() (enabled int) {
	for _, b := range Control.Dapp_list {
		if b {
			enabled++
		}
	}

	return
}

// Returns if a dApp is enabled
func DappEnabled(dapp string) bool {
	if b, ok := Control.Dapp_list[dapp]; ok && b {
		return true
	}

	return false
}

// Save dReams config.json file for platform wide dApp use
func WriteDreamsConfig(u dreams.SaveData) {
	if u.Daemon != nil && u.Daemon[0] == "" {
		if Control.Daemon_config != "" {
			u.Daemon[0] = Control.Daemon_config
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

		Gnomes.Para = 1
		Gnomes.DBType = "boltdb"
		Control.Dapp_list = make(map[string]bool)
		Control.Enabled_assets = make(map[string]bool)

		return
	}

	Gnomes.Para = 1
	Gnomes.DBType = "boltdb"
	Control.Dapp_list = make(map[string]bool)
	Control.Enabled_assets = make(map[string]bool)

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
	Control.Dapp_list = saved.Dapps

	if saved.Assets != nil {
		Control.Enabled_assets = saved.Assets
	}

	if saved.DBtype == "gravdb" {
		Gnomes.DBType = saved.DBtype
	}

	if saved.Para > 0 && saved.Para < 6 {
		Gnomes.Para = saved.Para
	}

	return
}

// Daemon rpc entry object with default options
//   - Bound to rpc.Daemon.Rpc
func DaemonRpcEntry() fyne.Widget {
	options := []string{
		"",
		rpc.DAEMON_RPC_DEFAULT,
		rpc.DAEMON_RPC_REMOTE1,
		rpc.DAEMON_RPC_REMOTE2,
		rpc.DAEMON_RPC_REMOTE3,
		rpc.DAEMON_RPC_REMOTE4,
		rpc.DAEMON_RPC_REMOTE5,
		rpc.DAEMON_RPC_REMOTE6,
	}

	if Control.Daemon_config != "" {
		options = append(options, Control.Daemon_config)
	}
	entry := widget.NewSelectEntry(options)
	entry.PlaceHolder = "Daemon RPC: "

	this := binding.BindString(&rpc.Daemon.Rpc)
	entry.Bind(this)

	return entry
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
func RateConfirm(scid string, tab *container.AppTabs, reset fyne.CanvasObject) fyne.CanvasObject {
	label := widget.NewLabel(fmt.Sprintf("Rate your experience with this contract\n\n%s", scid))
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	rating_label := widget.NewLabel("")
	rating_label.Wrapping = fyne.TextWrapWord
	rating_label.Alignment = fyne.TextAlignCenter

	fee_label := widget.NewLabel("")
	fee_label.Wrapping = fyne.TextWrapWord
	fee_label.Alignment = fyne.TextAlignCenter

	var slider *widget.Slider
	confirm := widget.NewButton("Confirm", func() {
		var pos uint64
		if slider.Value > 0 {
			pos = 1
		}

		fee := uint64(math.Abs(slider.Value * 10000))
		rpc.RateSCID(scid, fee, pos)
		tab.Selected().Content = reset
		tab.Selected().Content.Refresh()
	})

	confirm.Hide()

	cancel := widget.NewButton("Cancel", func() {
		tab.Selected().Content = reset
		tab.Selected().Content.Refresh()
	})

	slider = widget.NewSlider(-5, 5)
	slider.Step = 0.5
	slider.OnChanged = func(f float64) {
		if slider.Value != 0 {
			rating_label.SetText(fmt.Sprintf("Rating: %.0f", f*10000))
			fee_label.SetText(fmt.Sprintf("Fee: %.5f Dero", math.Abs(f)/10))
			confirm.Show()
		} else {
			rating_label.SetText("Pick a rating")
			fee_label.SetText("")
			confirm.Hide()
		}
	}

	good := canvas.NewImageFromResource(bundle.ResourceBlueBadge3Png)
	good.SetMinSize(fyne.NewSize(30, 30))
	bad := canvas.NewImageFromResource(bundle.ResourceRedBadgePng)
	bad.SetMinSize(fyne.NewSize(30, 30))

	rate_cont := container.NewBorder(nil, nil, bad, good, slider)

	left := container.NewVBox(confirm)
	right := container.NewVBox(cancel)
	buttons := container.NewAdaptiveGrid(2, left, right)

	content := container.NewVBox(layout.NewSpacer(), label, rating_label, fee_label, layout.NewSpacer(), rate_cont, layout.NewSpacer(), buttons)

	return container.NewStack(content)

}

var Username string

// Dero wallet name entry
func NameEntry() fyne.CanvasObject {
	Control.Names = widget.NewSelect([]string{}, func(s string) {
		Username = s
	})

	Control.Names.PlaceHolder = "Wallet names:"

	return container.NewHBox(layout.NewSpacer(), Control.Names)
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

// Index entry and NFA control objects
//   - Pass window resources for side menu windows
func IndexEntry(window_icon fyne.Resource, w fyne.Window) fyne.CanvasObject {
	Assets.Index_entry = widget.NewMultiLineEntry()
	Assets.Index_entry.PlaceHolder = "SCID:"
	Assets.Index_button = widget.NewButton("Add to Index", func() {
		if Gnomes.IsReady() {
			s := strings.Split(Assets.Index_entry.Text, "\n")
			if err := manualIndex(s); err == nil {
				dialog.NewInformation("Added to Index", "SCIDs added", w).Show()
			} else {
				dialog.NewInformation("Error", "Error adding SCIDs to index", w).Show()
			}
		}
	})

	Assets.Index_search = widget.NewButton("Search Index", func() {
		searchIndex(Assets.Index_entry.Text)
	})

	Control.Send_asset = widget.NewButton("Send Asset", func() {
		go sendAssetMenu(window_icon)
	})

	Control.List_button = widget.NewButton("List Asset", func() {
		go listMenu(window_icon)
	})

	Control.Claim_button = widget.NewButton("Claim NFA", func() {
		if len(Assets.Index_entry.Text) == 64 {
			if isNfa(Assets.Index_entry.Text) {
				if tx := rpc.ClaimNFA(Assets.Index_entry.Text); tx != "" {
					go ShowTxDialog("Claim NFA", fmt.Sprintf("TX: %s", tx), tx, 3*time.Second, w)
				} else {
					dialog.NewInformation("Claim NFA", "TX Error", w).Show()
				}

				return
			}

			dialog.NewInformation("Claim NFA", "Could not validate SCID as NFA", w).Show()
			return
		}

		dialog.NewInformation("Claim NFA", "Not a valid SCID", w).Show()
	})

	Assets.Index_button.Hide()
	Assets.Index_search.Hide()
	Control.List_button.Hide()
	Control.Claim_button.Hide()
	Control.Send_asset.Hide()

	Assets.Gnomes_index = canvas.NewText(" Indexed SCIDs: ", bundle.TextColor)
	Assets.Gnomes_index.TextSize = 18

	bottom_grid := container.NewAdaptiveGrid(3, Assets.Gnomes_index, Assets.Index_button, Assets.Index_search)
	top_grid := container.NewAdaptiveGrid(3, container.NewStack(Control.Send_asset), Control.Claim_button, Control.List_button)
	box := container.NewVBox(top_grid, layout.NewSpacer(), bottom_grid)

	return container.NewAdaptiveGrid(2, Assets.Index_entry, box)
}

// Disable index objects
func DisableIndexControls(d bool) {
	if d {
		Assets.Index_button.Hide()
		Assets.Index_search.Hide()
		Assets.Header_box.Hide()
		Market.Market_box.Hide()
		Gnomes.SCIDS = 0
	} else {
		Assets.Index_button.Show()
		Assets.Index_search.Show()
		if rpc.Wallet.IsConnected() {
			Control.Claim_button.Show()
			Assets.Header_box.Show()
			Market.Market_box.Show()
			if Control.List_open {
				Control.List_button.Hide()
			}
		} else {
			Control.Send_asset.Hide()
			Control.List_button.Hide()
			Control.Claim_button.Hide()
			Assets.Header_box.Hide()
			Market.Market_box.Hide()
		}
	}
	Assets.Index_button.Refresh()
	Assets.Index_search.Refresh()
	Assets.Header_box.Refresh()
	Market.Market_box.Refresh()
}

// Owned asset list object
//   - Sets Control.Viewing_asset and asset stats on selected
func AssetList() fyne.CanvasObject {
	Assets.Asset_list = widget.NewList(
		func() int {
			return len(Assets.Assets)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if len(Assets.Assets) > 0 {
				o.(*widget.Label).SetText(Assets.Assets[i])
			}
		})

	Assets.Asset_list.OnSelected = func(id widget.ListItemID) {
		split := strings.Split(Assets.Assets[id], "   ")
		if len(split) >= 2 {
			trimmed := strings.Trim(split[1], " ")
			Control.Viewing_asset = trimmed
			Assets.Icon = *canvas.NewImageFromImage(nil)
			go GetOwnedAssetStats(trimmed)
		}
	}

	return container.NewStack(Assets.Asset_list)
}

// Send Dero asset menu
//   - Asset SCID can be sent as payload to receiver when sending asset
//   - Pass resources for window_icon
func sendAssetMenu(window_icon fyne.Resource) {
	Control.send_open = true
	saw := fyne.CurrentApp().NewWindow("Send Asset")
	saw.Resize(fyne.NewSize(330, 700))
	saw.SetIcon(window_icon)
	Control.Send_asset.Hide()
	Control.List_button.Hide()
	saw.SetCloseIntercept(func() {
		Control.send_open = false
		if rpc.Wallet.IsConnected() {
			Control.Send_asset.Show()
			if isNfa(Control.Viewing_asset) {
				Control.List_button.Show()
			}
		}
		saw.Close()
	})
	saw.SetFixedSize(true)

	var saw_content *fyne.Container
	var send_button *widget.Button

	viewing_asset := Control.Viewing_asset

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
			info_label.SetText("Enter destination address.")
			send_button.Hide()
		}
	}

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

			confirm_button := widget.NewButton("Confirm", func() {
				if dest_entry.Validate() == nil {
					var load bool
					if payload.Checked {
						load = true
					}
					go rpc.SendAsset(send_asset, dest, load)
					Control.send_open = false
					saw.Close()
				}
			})

			cancel_button := widget.NewButton("Cancel", func() {
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
	send_button.Hide()

	icon := Assets.Icon

	saw_content = container.NewVBox(
		viewing_label,
		menuAssetImg(&icon, bundle.ResourceAvatarFramePng),
		layout.NewSpacer(),
		dest_entry,
		container.NewCenter(payload),
		layout.NewSpacer(),
		container.NewAdaptiveGrid(2, layout.NewSpacer(), send_button))

	go func() {
		for rpc.IsReady() && Control.send_open {
			time.Sleep(2 * time.Second)
			if !confirm_open {
				icon = Assets.Icon
				saw_content.Objects[1] = menuAssetImg(&icon, bundle.ResourceAvatarFramePng)
				if viewing_asset != Control.Viewing_asset {
					viewing_asset = Control.Viewing_asset
					viewing_label.SetText("Sending SCID:\n\n" + viewing_asset + " \n\nEnter destination address below\n\nSCID can be sent to receiver as payload\n\n")
				}
				saw_content.Refresh()
			}
		}
		Control.send_open = false
		saw.Close()
	}()

	saw.SetContent(
		container.NewStack(
			BackgroundRast("sendAssetMenu"),
			bundle.Alpha180,
			saw_content))
	saw.Show()
}

// Image for send asset and list menus
//   - Pass res for frame resource
func menuAssetImg(img *canvas.Image, res fyne.Resource) fyne.CanvasObject {
	img.SetMinSize(fyne.NewSize(100, 100))
	img.Resize(fyne.NewSize(94, 94))
	img.Move(fyne.NewPos(118, 3))

	frame := canvas.NewImageFromResource(res)
	frame.Resize(fyne.NewSize(100, 100))
	frame.Move(fyne.NewPos(115, 0))

	cont := container.NewWithoutLayout(img, frame)

	return cont
}

// NFA listing menu
//   - Pass resources for menu window to match main
func listMenu(window_icon fyne.Resource) {
	Control.List_open = true
	aw := fyne.CurrentApp().NewWindow("List NFA")
	aw.Resize(fyne.NewSize(330, 700))
	aw.SetIcon(window_icon)
	Control.List_button.Hide()
	Control.Send_asset.Hide()
	aw.SetCloseIntercept(func() {
		Control.List_open = false
		if rpc.Wallet.IsConnected() {
			Control.Send_asset.Show()
			if isNfa(Control.Viewing_asset) {
				Control.List_button.Show()
			}
		}
		aw.Close()
	})
	aw.SetFixedSize(true)

	var aw_content *fyne.Container
	var set_list *widget.Button

	viewing_asset := Control.Viewing_asset
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

	charAddr := widget.NewEntry()
	charAddr.SetPlaceHolder("Charity Donation Address:")
	charAddr.Validator = validation.NewRegexp(`^(dero)\w{62}$`, "Int required")

	charPerc := dwidget.NewDeroEntry("", 1, 0)
	charPerc.AllowFloat = false
	charPerc.SetPlaceHolder("Charity Donation %:")
	charPerc.Validator = validation.NewRegexp(`^\d{1,2}$`, "Int required")
	charPerc.OnChanged = func(s string) {
		if listing.Selected != "" && duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
			set_list.Show()
		} else {
			set_list.Hide()
		}
	}

	duration.OnChanged = func(s string) {
		if rpc.StringToInt(s) > 168 {
			duration.SetText("168")
		}

		if listing.Selected != "" && duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
			set_list.Show()
		} else {
			set_list.Hide()
		}
	}

	start.OnChanged = func(s string) {
		if listing.Selected != "" && duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
			set_list.Show()
		} else {
			set_list.Hide()
		}
	}

	charAddr.OnChanged = func(s string) {
		if listing.Selected != "" && duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
			set_list.Show()
		} else {
			set_list.Hide()
		}
	}

	listing.OnChanged = func(s string) {
		if listing.Selected != "" && duration.Validate() == nil && start.Validate() == nil && charAddr.Validate() == nil && charPerc.Validate() == nil {
			set_list.Show()
		} else {
			set_list.Hide()
		}
	}

	var confirm_open bool
	set_list = widget.NewButton("Set Listing", func() {
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

				cancel_button := widget.NewButton("Cancel", func() {
					confirm_open = false
					aw.SetContent(
						container.NewStack(
							BackgroundRast("listMenu"),
							bundle.Alpha180,
							aw_content))
				})

				confirm_button := widget.NewButton("Confirm", func() {
					rpc.SetNFAListing(listing_asset, listing.Selected, charAddr.Text, d, s, cp)
					Control.List_open = false
					if rpc.Wallet.IsConnected() {
						Control.Send_asset.Show()
						if isNfa(Control.Viewing_asset) {
							Control.List_button.Show()
						}
					}
					aw.Close()
				})

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
	set_list.Hide()

	icon := Assets.Icon

	go func() {
		for rpc.IsReady() && Control.List_open {
			time.Sleep(2 * time.Second)
			if !confirm_open && isNfa(Control.Viewing_asset) {
				icon = Assets.Icon
				aw_content.Objects[2] = menuAssetImg(&icon, bundle.ResourceAvatarFramePng)
				if viewing_asset != Control.Viewing_asset {
					viewing_asset = Control.Viewing_asset
					viewing_label.SetText(fmt.Sprintf("Listing SCID:\n\n%s", viewing_asset))
				}
				aw_content.Refresh()
			}
		}
		Control.List_open = false
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
		menuAssetImg(&icon, bundle.ResourceAvatarFramePng),
		layout.NewSpacer(),
		layout.NewSpacer(),
		listing,
		duration,
		start,
		container.NewCenter(enable_donations),
		charAddr,
		charPerc,
		container.NewCenter(fee_label),
		container.NewAdaptiveGrid(2, layout.NewSpacer(), set_list))

	aw.SetContent(
		container.NewStack(
			BackgroundRast("listMenu"),
			bundle.Alpha180,
			aw_content))
	aw.Show()
}

type IntroText struct {
	name    string
	content []string
}

// Create menu tree items for dApps
func MakeMenuIntro(items map[string][]string) (entries []IntroText) {
	var menu_entry IntroText
	for name, e := range items {
		menu_entry.name = name
		menu_entry.content = e
		entries = append(entries, menu_entry)
	}

	return
}

// Menu instruction tree
func IntroTree(intros []IntroText) fyne.CanvasObject {
	list := map[string][]string{
		"":                        {"Welcome to dReams"},
		"Welcome to dReams":       {"Get Started", "dApps", "Assets", "Market"},
		"Get Started":             {"Visit dero.io for daemon and wallet download info", "Connecting", "FAQ"},
		"Connecting":              {"Daemon", "Wallet"},
		"FAQ":                     {"Can't connect", "How to resync Gnomon DB", "Can't see any tables, contracts or market info", "How to see terminal log", "Visit dreamdapps.io for further documentation"},
		"Can't connect":           {"Using a local daemon will yield the best results", "If you are using a remote daemon, try changing daemons", "Any connection errors can be found in terminal log"},
		"How to resync Gnomon DB": {"Go to Gnomon options in Menu", "If Gnomon is running you will be prompted to shut it down to make changes", "Click the delete DB button", "Reconnect to a daemon to resync", "Any sync errors can be found in terminal log"},

		"Can't see any tables, contracts or market info": {"Make sure daemon, wallet and Gnomon indicators are lit up solid", "If you've added new dApps to your dReams, a Gnomon resync will add them to your index", "Look in the asset tab for number of indexed SCIDs, it should be above 0", "Make sure your collection or dApp is enabled", "Try resyncing", "Any errors can be found in terminal log"},

		"How to see terminal log": {"Windows", "Mac", "Linux"},
		"Windows":                 {"Open powershell or command prompt", "Navigate to dReams directory", `Start dReams using       .\dReams-windows-amd64.exe`},
		"Mac":                     {"Open a terminal", "Navigate to dReams directory", `Start dReams using       ./dReams-macos-amd64`},
		"Linux":                   {"Open a terminal", "Navigate to dReams directory", `Start dReams using       ./dReams-linux-amd64`},
		"Daemon":                  {"Using local daemon will give best performance while using dReams", "Remote daemon options are available in drop down if a local daemon is not available", "Enter daemon address and the D light in top right will light up if connection is successful", "Once daemon is connected Gnomon will start up, the Gnomon indicator light will have a stripe in middle"},
		"Wallet":                  {"Set up and register a Dero wallet", "Your wallet will need to be running rpc server", "Using cli, start your wallet with flags --rpc-server --rpc-login=user:pass", "With Engram, turn on cyberdeck to start rpc server", "In dReams enter your wallet rpc address and rpc user:pass", "Press connect and the W light in top right will light up if connection is successful", "Once wallet is connected and Gnomon is running, Gnomon will sync with wallet", "The Gnomon indicator will turn solid when this is complete, everything is now connected"},

		"dApps":         {"Loading dApps", "Holdero", "Baccarat", "Predictions", "Sports", "dService", "Iluma", "Asset Duels", "Contract Ratings"},
		"Loading dApps": {"You can add or remove dApps in the dApps tab", "Loading changes will disconnect your wallet", "Gnomon will continue to run, but may need to be resynced to index any new dApps added", "Your dApp preferences will be saved in local config file", "Loading only the dApps you are using will increase Gnomon and dReams performance"},
	}

	for i := range intros {
		list[intros[i].name] = intros[i].content
	}

	list["Contract Ratings"] = []string{
		"dReams has a public rating store on chain for multiplayer contracts",
		"Players can rate other contracts positively or negatively",
		"Four rating tiers, tier two being the starting tier for all contracts",
		"Each rating transaction is weight based by its Dero value",
		"Contracts that fall below tier one will no longer populate in the public index"}

	list["Assets"] = []string{
		"Enabling assets collections",
		"View any owned assets held in wallet",
		"Put owned assets up for auction or for sale",
		"Send assets privately to another wallet",
		"Indexer, add custom contracts to your index and search current index DB"}

	list["Enabling assets collections"] = []string{
		"You can enable or disable indexing of any asset collection in the Asset/Enabled tab",
		"Changes will require Gnomon DB to be resynced to take effect",
		"Your collection preferences will be saved in local config file",
		"Loading only the asset collections you are using will increase Gnomon and dReams performance"}

	list["Market"] = []string{
		"View any in game assets up for auction or sale",
		"Search all NFAs",
		"Bid on or buy assets",
		"Cancel or close out any existing listings",
		"Create NFA charity auctions and sales"}

	tree := widget.NewTreeWithStrings(list)

	tree.OnBranchClosed = func(uid widget.TreeNodeID) {
		tree.UnselectAll()
		if uid == "Welcome to dReams" {
			tree.CloseAllBranches()
		}
	}

	tree.OnBranchOpened = func(uid widget.TreeNodeID) {
		tree.Select(uid)
	}

	tree.OpenBranch("Welcome to dReams")

	max := container.NewStack(tree)

	return max
}

// Used for placing coin decimal, default returns 2 decimal place
func CoinDecimal(ticker string) int {
	split := strings.Split(ticker, "-")
	if len(split) == 2 {
		switch split[1] {
		case "BTC":
			return 8
		case "DERO":
			return 5
		default:
			return 2
		}
	}

	return 2
}

func CreateSwapContainer(pair string) (*dwidget.DeroAmts, *fyne.Container) {
	split := strings.Split(pair, "-")
	if len(split) != 2 {
		return dwidget.NewDeroEntry("", 0, 0), container.NewStack(widget.NewLabel("Invalid Pair"))
	}

	incr := 0.1
	switch split[0] {
	case "dReams":
		incr = 1
	}

	color1 := color.RGBA{0, 0, 0, 0}
	color2 := color.RGBA{0, 0, 0, 0}
	image1 := canvas.NewImageFromResource(bundle.ResourceSwapFrame1Png)
	image2 := canvas.NewImageFromResource(bundle.ResourceSwapFrame2Png)

	rect2 := canvas.NewRectangle(color2)
	rect2.SetMinSize(fyne.NewSize(200, 100))
	swap2_label := canvas.NewText(split[1], bundle.TextColor)
	swap2_label.Alignment = fyne.TextAlignCenter
	swap2_label.TextSize = 18
	swap2_entry := dwidget.NewDeroEntry("", incr, uint(CoinDecimal(split[0])))
	swap2_entry.SetText("0")
	swap2_entry.Disable()

	pad2 := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), swap2_entry)

	swap2 := container.NewBorder(nil, pad2, nil, nil, container.NewCenter(swap2_label))
	cont2 := container.NewStack(rect2, image2, swap2)

	rect1 := canvas.NewRectangle(color1)
	rect1.SetMinSize(fyne.NewSize(200, 100))
	swap1_label := canvas.NewText(split[0], bundle.TextColor)
	swap1_label.Alignment = fyne.TextAlignCenter
	swap1_label.TextSize = 18
	swap1_entry := dwidget.NewDeroEntry("", incr, uint(CoinDecimal(split[0])))
	swap1_entry.SetText("0")
	swap1_entry.Validator = validation.NewRegexp(`^\d{1,}\.\d{1,5}$|^[^0.]\d{0,}$`, "Int or float required")
	swap1_entry.OnChanged = func(s string) {
		switch pair {
		case "DERO-dReams", "dReams-DERO":
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				ex := float64(333)
				if split[0] == "dReams" {
					new := f / ex
					swap2_entry.SetText(fmt.Sprintf("%.5f", new))
					return
				}

				new := f * ex
				swap2_entry.SetText(fmt.Sprintf("%.5f", new))

			}
		default:
			// other pairs
		}
	}

	pad1 := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), swap1_entry)

	swap1 := container.NewBorder(nil, pad1, nil, nil, container.NewCenter(swap1_label))
	cont1 := container.NewStack(rect1, image1, swap1)

	return swap1_entry, container.NewAdaptiveGrid(2, cont1, cont2)
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

		if img, _, err = image.Decode(bytes.NewReader(bundle.ResourceBackgroundPng.Content())); err == nil {
			return canvas.NewRasterFromImage(img)
		}

		logger.Warnf("[%s] Fallback %s\n", tag, err)
	}

	return canvas.NewRasterFromImage(image.Rect(2, 2, 4, 4))

}

// Send Dero message menu
func SendMessageMenu(dest string, window_icon fyne.Resource) {
	if !Control.msg_open && rpc.Wallet.IsConnected() {
		Control.msg_open = true
		smw := fyne.CurrentApp().NewWindow("Send Message")
		smw.Resize(fyne.NewSize(330, 700))
		smw.SetIcon(window_icon)
		smw.SetCloseIntercept(func() {
			Control.msg_open = false
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
				Control.msg_open = false
				smw.Close()
			}
		})
		send_button.Hide()

		dest_cont := container.NewVBox(label, ringsize, dest_entry)
		message_cont := container.NewBorder(nil, send_button, nil, nil, message_entry)

		content := container.NewVSplit(dest_cont, message_cont)

		go func() {
			for rpc.IsReady() && Control.msg_open {
				time.Sleep(2 * time.Second)
			}
			Control.msg_open = false
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
