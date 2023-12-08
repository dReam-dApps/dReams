package menu

import (
	"bytes"
	"fmt"
	"image/color"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/gnomes"
	"github.com/dReam-dApps/dReams/rpc"
)

type assetObjects struct {
	Headers  *fyne.Container
	Swap     *fyne.Container
	Claim    *fyne.Container
	Balances *widget.List
	List     *widget.List
	Asset    []Asset
	Viewing  string
	SCIDs    map[string]string
	Icon     fyne.CanvasObject
	Button   struct {
		Rescan   *widget.Button
		Send     *widget.Button
		List     *widget.Button
		scanning bool
		sending  bool
		listing  bool
	}
	Index struct {
		Entry  *widget.Entry
		Add    *widget.Button
		Search *widget.Button
	}
}

// Asset info
type Asset struct {
	Name       string `json:"name"`
	Collection string `json:"collection"`
	SCID       string `json:"scid"`
	Type       string `json:"type"`
	Image      []byte `json:"image"`
}

var Assets assetObjects

var dReamsNFAs = []assetCount{
	{name: "AZY-Playing card decks", count: 23, creator: AZY_mint},
	{name: "AZY-Playing card backs", count: 53, creator: AZY_mint},
	{name: "AZY-Deroscapes", count: 10, creator: AZY_mint},
	{name: "Death By Cupcake", count: 8, creator: DCB_mint},
	{name: "SIXPC", count: 9, creator: SIX_mint},
	{name: "SIXPCB", count: 10, creator: SIX_mint},
	{name: "SIXART", count: 17, creator: SIX_mint},
	{name: "High Strangeness", count: 376, creator: HS_mint},
	{name: "Dorblings NFA", count: 110, creator: Dorbling_mint},
	{name: "Dero Desperados", count: 777, creator: Desperado_mint},
	{name: "Desperado Guns", count: 777, creator: Desperado_mint},
	// TODO DLAMPP count
	// {name: "DLAMPP ", count: ?},
}

// Add asset to List and SCIDs
func (a *assetObjects) Add(details Asset, url string) {
	have, err := gnomes.StorageExists(details.Collection, details.Name)
	if err != nil {
		have = false
		logger.Errorln("[AddAsset]", err)
	}

	if have {
		var new Asset
		gnomes.GetStorage(details.Collection, details.Name, &new)
		if new.Image != nil && !bytes.Equal(new.Image, bundle.ResourceMarketCirclePng.StaticContent) {
			details.Image = new.Image
		} else {
			have = false
		}
	}

	if !have {
		if img, err := dreams.DownloadBytes(url); err == nil {
			details.Image = img
		} else {
			details.Image = bundle.ResourceMarketCirclePng.StaticContent
			logger.Errorln("[AddAsset]", err)
		}
	}

	a.Asset = append(a.Asset, details)
	a.SCIDs[details.Name] = details.SCID
}

// Sorts asset list by name
func (a *assetObjects) SortList() {
	sort.Slice(a.Asset, func(i, j int) bool {
		return a.Asset[i].Name < a.Asset[j].Name
	})
	a.List.Refresh()
}

// Check if string is dReams NFA collection
func IsDreamsNFACollection(collection string) bool {
	for _, c := range dReamsNFAs {
		if c.name == collection {
			return true
		}
	}

	return false
}

// Check if string is dReams NFA creator address
func IsDreamsNFACreator(creator string) bool {
	for _, c := range dReamsNFAs {
		if c.creator == creator {
			return true
		}
	}

	return false
}

// Get the nameHdr of a NFA
func GetNFAName(scid string) string {
	if gnomon.IsReady() {
		name, _ := gnomon.GetSCIDValuesByKey(scid, "nameHdr")
		if name != nil {
			return name[0]
		}
	}

	return ""
}

// Check if SCID is a NFA
func isNFA(scid string) bool {
	if gnomon.IsReady() {
		artAddr, _ := gnomon.GetSCIDValuesByKey(scid, "artificerAddr")
		if artAddr != nil {
			return artAddr[0] == rpc.ArtAddress
		}
	}
	return false
}

// Check if SCID is a valid NFA
//   - file != "-"
func ValidNFA(file string) bool {
	return file != "-"
}

// Additional asset type info
func AssetType(collection, typeHdr string) string {
	switch collection {
	case "AZY-Playing card decks", "SIXPC":
		return "Playing card deck"
	case "AZY-Playing card backs", "SIXPCB":
		return "Playing card back"
	case "AZY-Deroscapes", "SIXART":
		return "Theme/Avatar"
	case "Dorblings NFA":
		return "Avatar"
	case "Death By Cupcake", "High Strangeness", "Dero Desperados", "Desperado Guns":
		return "Avatar/Duel"
	default:
		return typeHdr
	}
}

// Parse url for ipfs prefix
func ParseURL(url string) string {
	if strings.HasPrefix(url, "ipfs://") {
		return fmt.Sprintf("https://ipfs.io/ipfs/%s", url[7:])
	}

	return url
}

// Creates framed icon image
func AssetIcon(icon []byte, name string, size float32) fyne.CanvasObject {
	frame := canvas.NewImageFromResource(bundle.ResourceAvatarFramePng)
	frame.SetMinSize(fyne.NewSize(size, size))
	if icon == nil {
		icon = bundle.ResourceMarketCirclePng.StaticContent
	}

	img := canvas.NewImageFromReader(bytes.NewReader(icon), name)
	if img == nil {
		return container.NewStack(frame)
	}

	img.SetMinSize(fyne.NewSize(size, size))
	border := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), img)

	return container.NewStack(border, frame)
}

// Returns search filter with all enabled NFAs
func ReturnEnabledNFAs(assets map[string]bool) (filters []string) {
	for name, enabled := range assets {
		if enabled {
			if IsDreamsNFACollection(name) {
				filters = append(filters, fmt.Sprintf(`450 STORE("collection", "%s`, name))
			}
		}
	}

	return
}

func ReturnAssetCount() (count int) {
	count = Control.NFA_count + Control.G45_count - 10
	if count < 2 {
		count = 2
	}

	return
}

// Options for enabling NFA collection
func enableNFAOpts(asset assetCount) (opts *widget.RadioGroup) {
	onChanged := func(s string) {
		if s == "Yes" {
			Control.Lock()
			Control.Enabled_assets[asset.name] = true
			Control.NFA_count += asset.count
			Control.Unlock()
			return
		}

		Control.Lock()
		defer Control.Unlock()
		Control.Enabled_assets[asset.name] = false
		if Control.NFA_count >= asset.count {
			Control.NFA_count -= asset.count
		}
	}

	if !Control.once {
		opts = widget.NewRadioGroup([]string{"Yes", "No"}, nil)
		opts.Required = true
		opts.Horizontal = true
		if Control.Enabled_assets[asset.name] {
			opts.OnChanged = onChanged
			opts.SetSelected("Yes")
		} else {
			opts.SetSelected("No")
			opts.OnChanged = onChanged
		}

		return
	}

	opts = widget.NewRadioGroup([]string{"Yes", "No"}, nil)
	opts.Required = true
	opts.Horizontal = true
	if Control.Enabled_assets[asset.name] {
		opts.SetSelected("Yes")
	} else {
		opts.SetSelected("No")
	}
	opts.OnChanged = onChanged

	return
}

// Options for enabling G45 collection
func enableG45Opts(asset assetCount) (opts *widget.RadioGroup) {
	onChanged := func(s string) {
		if s == "Yes" {
			Control.Lock()
			Control.Enabled_assets[asset.name] = true
			Control.G45_count += asset.count
			Control.Unlock()
			return
		}

		Control.Lock()
		defer Control.Unlock()
		Control.Enabled_assets[asset.name] = false
		if Control.G45_count >= asset.count {
			Control.G45_count -= asset.count
		}
	}

	if !Control.once {
		opts = widget.NewRadioGroup([]string{"Yes", "No"}, nil)
		opts.Required = true
		opts.Horizontal = true
		if Control.Enabled_assets[asset.name] {
			opts.OnChanged = onChanged
			opts.SetSelected("Yes")
		} else {
			opts.SetSelected("No")
			opts.OnChanged = onChanged
		}

		return
	}

	opts = widget.NewRadioGroup([]string{"Yes", "No"}, nil)
	opts.Required = true
	opts.Horizontal = true
	if Control.Enabled_assets[asset.name] {
		opts.SetSelected("Yes")
	} else {
		opts.SetSelected("No")
	}
	opts.OnChanged = onChanged

	return
}

// Enable asset collection objects
//   - intro used to set label if initial boot screen
func EnabledCollections(intro bool) (obj fyne.CanvasObject) {
	collection_form := []*widget.FormItem{}
	enable_all := widget.NewButton("Enable All", func() {
		for _, item := range collection_form {
			item.Widget.(*widget.RadioGroup).SetSelected("Yes")

		}
	})

	disable_all := widget.NewButton("Disable All", func() {
		for _, item := range collection_form {
			item.Widget.(*widget.RadioGroup).SetSelected("No")
		}
	})

	enable_all.Importance = widget.HighImportance
	disable_all.Importance = widget.HighImportance

	for _, asset := range dReamsNFAs {
		collection_form = append(collection_form, widget.NewFormItem(asset.name, enableNFAOpts(asset)))
	}

	for _, asset := range dReamsG45s {
		collection_form = append(collection_form, widget.NewFormItem(asset.name, enableG45Opts(asset)))
	}

	Control.once = true
	if Control.NFA_count < 3 {
		Control.NFA_count = 3
	}

	label := canvas.NewText("Delete Gnomon DB and resync for changes to take effect", bundle.TextColor)
	label.Alignment = fyne.TextAlignCenter
	if intro {
		label.Text = ""
	}

	return container.NewBorder(
		nil,
		container.NewBorder(nil, nil, enable_all, disable_all, label),
		nil,
		nil,
		container.NewVScroll(container.NewCenter(widget.NewForm(collection_form...))))

}

// Returns string with all enabled asset names formatted for a label
func returnEnabledNames(assets map[string]bool) (text string) {
	var names []string
	for name, enabled := range assets {
		if enabled {
			if IsDreamsNFACollection(name) {
				names = append(names, name)
			}
		}
	}

	for name, enabled := range assets {
		if enabled && IsDreamsG45(name) {
			names = append(names, name)
		}
	}

	sort.Strings(names)

	for _, n := range names {
		text = text + n + "\n\n"
	}

	return
}

// Owned asset tab layout
//   - tag for log print
//   - profile is canvas object of widgets used to select assets for games, themes, ect
//   - rescan is func used to rescan wallet assets
//   - icon resources for side menus
//   - d for main window dialogs
func PlaceAssets(tag string, profile fyne.CanvasObject, rescan func(), icon fyne.Resource, d *dreams.AppObject) *fyne.Container {
	enable_opts := EnabledCollections(false)

	scid_entry := widget.NewEntry()
	scid_entry.SetPlaceHolder("SCID:")

	line := canvas.NewLine(bundle.TextColor)
	line_spacer := canvas.NewRectangle(color.Transparent)
	line_spacer.SetMinSize(fyne.NewSize(300, 0))

	name_entry := widget.NewEntry()
	name_entry.SetPlaceHolder("Name:")

	descr_entry := widget.NewMultiLineEntry()
	descr_entry.SetPlaceHolder("Description:")

	icon_entry := widget.NewEntry()
	icon_entry.SetPlaceHolder("Icon:")

	header_spacer := canvas.NewRectangle(color.Transparent)
	header_spacer.SetMinSize(fyne.NewSize(580, 30))

	header_button := widget.NewButton("Set Headers", nil)
	header_button.Importance = widget.HighImportance

	headers := container.NewVBox(scid_entry, container.NewVBox(line_spacer, line, line_spacer), header_spacer, container.NewVBox(line_spacer, line, line_spacer), name_entry, descr_entry, icon_entry, container.NewCenter(header_button))
	Assets.Headers = container.NewCenter(headers)
	Assets.Headers.Hide()

	scroll_top := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "arrowUp"), func() {
		Assets.List.ScrollToTop()
	})

	scroll_bottom := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "arrowDown"), func() {
		Assets.List.ScrollToBottom()
	})

	scroll_top.Importance = widget.LowImportance
	scroll_bottom.Importance = widget.LowImportance

	Info.Indexed = canvas.NewText("Indexed SCIDs: ", bundle.TextColor)
	Info.Indexed.TextSize = 18

	scroll_spacer := canvas.NewRectangle(color.Transparent)
	scroll_spacer.SetMinSize(fyne.NewSize(77, 36))
	scroll_buttons := container.NewHBox(scroll_top, scroll_bottom)

	border := container.NewBorder(
		container.NewHBox(layout.NewSpacer(), Info.Indexed, container.NewStack(scroll_spacer, scroll_buttons)),
		nil,
		nil,
		nil)

	var tab *container.TabItem
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("", bundle.ResourceMarketCirclePng, layout.NewSpacer()),
		container.NewTabItem("Owned", AssetList(icon, rescan, d)),
		container.NewTabItem("Profile", container.NewBorder(
			dwidget.NewCanvasText("User Profile", 18, fyne.TextAlignCenter),
			nil,
			nil,
			nil,
			profile)),

		container.NewTabItem("Headers", container.NewBorder(
			dwidget.NewCanvasText("Gnomon SC Headers", 18, fyne.TextAlignCenter),
			nil,
			nil,
			nil,
			Assets.Headers)),

		container.NewTabItem("Index", container.NewBorder(
			dwidget.NewCanvasText("Gnomon Index", 18, fyne.TextAlignCenter),
			nil,
			nil,
			nil,
			container.NewAdaptiveGrid(2, indexEntry(d.Window), enable_opts))))

	tab = tabs.Items[1]
	tabs.Select(tab)
	tabs.DisableIndex(0)
	tabs.SetTabLocation(container.TabLocationLeading)
	tabs.OnSelected = func(ti *container.TabItem) {
		switch ti.Text {
		case "Owned":
			scroll_buttons.Show()
		case "Profile":
			scroll_buttons.Hide()
			if !rpc.Daemon.IsConnected() || !rpc.Wallet.IsConnected() {
				dialog.NewInformation("Profile", "Connect to daemon and wallet to set profile", d.Window).Show()
				tabs.Select(tab)
				return
			}
		case "Headers":
			scroll_buttons.Hide()
			if !rpc.Daemon.IsConnected() || !rpc.Wallet.IsConnected() {
				dialog.NewInformation("Headers", "Connect to daemon and wallet to set SC headers", d.Window).Show()
				tabs.Select(tab)
				return
			}
		case "Enabled":
			scroll_buttons.Show()
			if rpc.Daemon.IsConnected() {
				dialog.NewInformation("Assets", "Shut down Gnomon to make changes to asset index", d.Window).Show()
				tabs.Selected().Content = container.NewBorder(
					dwidget.NewCanvasText("Enabled Assets", 18, fyne.TextAlignCenter),
					nil,
					nil,
					nil,
					container.NewVScroll(container.NewVBox(dwidget.NewCenterLabel(returnEnabledNames(Control.Enabled_assets)))))
				tab = ti
				return
			}
			tabs.Selected().Content = container.NewBorder(dwidget.NewCanvasText("Enabled Assets", 18, fyne.TextAlignCenter), nil, nil, nil, enable_opts)
		case "Index":
			scroll_buttons.Hide()
			if rpc.Daemon.IsConnected() {
				Assets.Index.Entry.Enable()
				enable_opts.(*fyne.Container).Objects[1].(*fyne.Container).Objects[1].(*widget.Button).Hide()
				enable_opts.(*fyne.Container).Objects[1].(*fyne.Container).Objects[2].(*widget.Button).Hide()
				enable_opts.(*fyne.Container).Objects[1].(*fyne.Container).Objects[0].(*canvas.Text).Text = "Shut down Gnomon to enable/disable collections"
				if f, ok := enable_opts.(*fyne.Container).Objects[0].(*container.Scroll).Content.(*fyne.Container).Objects[0].(*widget.Form); ok {
					for _, r := range f.Items {
						r.Widget.(*widget.RadioGroup).Disable()
					}
				}
			} else {
				Assets.Index.Entry.SetText("")
				Assets.Index.Entry.Disable()
				enable_opts.(*fyne.Container).Objects[1].(*fyne.Container).Objects[1].(*widget.Button).Show()
				enable_opts.(*fyne.Container).Objects[1].(*fyne.Container).Objects[2].(*widget.Button).Show()
				enable_opts.(*fyne.Container).Objects[1].(*fyne.Container).Objects[0].(*canvas.Text).Text = "Delete Gnomon DB and resync for changes to take effect"
				if f, ok := enable_opts.(*fyne.Container).Objects[0].(*container.Scroll).Content.(*fyne.Container).Objects[0].(*widget.Form); ok {
					for _, r := range f.Items {
						r.Widget.(*widget.RadioGroup).Enable()
					}
				}
			}
		}

		tab = ti
	}

	header_button.OnTapped = func() {
		scid := scid_entry.Text
		if len(scid) == 64 && name_entry.Text != "dReam Tables" && name_entry.Text != "dReams" {
			if _, ok := rpc.FindStringKey(rpc.GnomonSCID, scid, rpc.Daemon.Rpc).(string); ok {
				setHeaderConfirm(name_entry.Text, descr_entry.Text, icon_entry.Text, scid, d.Window)
			} else {
				dialog.NewInformation("Check back soon", "SCID not stored on the main Gnomon SC yet\n\nOnce stored, you can set your SCID headers", d.Window).Show()
			}
		} else {
			dialog.NewInformation("Not Valid", fmt.Sprintf("SCID %s is not valid", scid), d.Window).Show()
		}
	}

	return container.NewStack(bundle.Alpha120, tabs, border)
}

// Confirmation dialog for setting SCID headers
//   - name, desc and icon of SCID header on Gnomon SC
func setHeaderConfirm(name, desc, icon, scid string, w fyne.Window) {
	text := fmt.Sprintf("Headers for SCID:\n\n%s\n\nName: %s\n\nDescription: %s\n\nIcon: %s", scid, name, desc, icon)
	done := make(chan struct{})
	confirm := dialog.NewConfirm("Set Headers", text, func(b bool) {
		if b {
			rpc.SetHeaders(name, desc, icon, scid)
		}
		done <- struct{}{}
	}, w)

	go ShowConfirmDialog(done, confirm)
}

// Index entry and NFA control objects
//   - Pass window resources for side menu windows
func indexEntry(w fyne.Window) fyne.CanvasObject {
	Assets.Index.Entry = widget.NewMultiLineEntry()
	Assets.Index.Entry.SetPlaceHolder("Add SCID(s):")
	Assets.Index.Add = widget.NewButton("Add to Index", func() {
		if gnomon.IsReady() {
			s := strings.Split(Assets.Index.Entry.Text, "\n")
			if err := gnomes.AddToIndex(s); err == nil {
				dialog.NewInformation("Added to Index", "SCIDs added", w).Show()
			} else {
				dialog.NewInformation("Error", "Error adding SCIDs to index", w).Show()
			}
		}
	})

	Assets.Index.Search = widget.NewButton("Search Index", func() {
		if gnomon.IsReady() {
			scid := Assets.Index.Entry.Text
			if len(scid) == 64 {
				var found bool
				all := gnomon.GetAllOwnersAndSCIDs()
				for sc := range all {
					if scid == sc {
						dialog.NewInformation("Found", fmt.Sprintf("SCID %s found", scid), w).Show()
						logger.Printf("[Search] %s Found\n", scid)
						found = true
					}
				}
				if !found {
					dialog.NewInformation("Not Found", fmt.Sprintf("Index does not contain SCID %s", scid), w).Show()
					logger.Errorf("[Search] %s Not Found\n", scid)
				}
			} else {
				dialog.NewInformation("Not Valid", fmt.Sprintf("SCID %s is not valid", scid), w).Show()
				logger.Errorf("[Search] %s Not Found\n", scid)
			}
		}
	})

	Assets.Index.Add.Hide()
	Assets.Index.Search.Hide()

	return container.NewBorder(
		nil,
		container.NewCenter(container.NewHBox(Assets.Index.Add, Assets.Index.Search)),
		nil,
		nil,
		Assets.Index.Entry)
}

// Disable index objects
func DisableIndexControls(d bool) {
	if d {
		Assets.Index.Add.Hide()
		Assets.Index.Search.Hide()
		Assets.Headers.Hide()
		Market.Market_box.Hide()
		gnomon.ZeroIndexCount()
	} else {
		Assets.Index.Add.Show()
		Assets.Index.Search.Show()
		if rpc.Wallet.IsConnected() {
			Assets.Headers.Show()
			Assets.Claim.Show()
			Market.Market_box.Show()
			if !Assets.Button.scanning && gnomon.HasChecked() {
				Assets.Button.Rescan.Show()
			} else {
				Assets.Button.Rescan.Hide()
			}
			if Assets.Button.listing {
				Assets.Button.List.Hide()
			}
			if Assets.Button.sending {
				Assets.Button.Send.Hide()
			}
		} else {
			Assets.Button.Send.Hide()
			Assets.Button.List.Hide()
			Assets.Claim.Hide()
			Market.Market_box.Hide()
			Assets.Button.Rescan.Hide()
		}
	}
	Assets.Index.Add.Refresh()
	Assets.Index.Search.Refresh()
	Assets.Headers.Refresh()
	Market.Market_box.Refresh()
	Assets.Button.Rescan.Refresh()
}

// Owned asset list object
//   - Sets Assets.Viewing and buttons visibility on selected
//   - rescan is func placed in button to rescan wallet assets
func AssetList(icon fyne.Resource, rescan func(), d *dreams.AppObject) fyne.CanvasObject {
	Assets.List = widget.NewList(
		func() int {
			return len(Assets.Asset)
		},
		func() fyne.CanvasObject {
			return container.NewStack(
				container.NewBorder(
					nil,
					nil,
					container.NewCenter(canvas.NewImageFromImage(nil)),
					nil,
					container.NewBorder(
						widget.NewLabel(""),
						widget.NewLabel(""),
						nil,
						nil,
					)))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			a := Assets.Asset
			if i > len(a)-1 {
				return
			}

			header := fmt.Sprintf("%s   %s   %s", a[i].Name, a[i].Collection, a[i].SCID)
			if o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*widget.Label).Text != header {
				o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*widget.Label).SetText(header)
				o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[1].(*widget.Label).SetText(fmt.Sprintf("Type: %s", a[i].Type))
				o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[1].(*fyne.Container).Objects[0] = AssetIcon(a[i].Image, a[i].Name, 70)
				o.Refresh()
			}
		})

	Assets.List.OnSelected = func(id widget.ListItemID) {
		if len(Assets.Asset[id].SCID) == 64 {
			Assets.Viewing = Assets.Asset[id].SCID
			if !isNFA(Assets.Asset[id].SCID) {
				Assets.Button.List.Hide()
				Assets.Button.Send.Hide()
			} else {
				if !Assets.Button.listing && !Assets.Button.sending {
					Assets.Button.List.Show()
					Assets.Button.Send.Show()
				}
			}
			Assets.Icon = AssetIcon(Assets.Asset[id].Image, Assets.Asset[id].Name, 100)
		}
	}

	Assets.Button.Send = widget.NewButton("Send Asset", func() {
		go sendAssetMenu(icon)
	})

	Assets.Button.List = widget.NewButton("List Asset", func() {
		go listMenu(icon)
	})

	Assets.Button.Send.Importance = widget.HighImportance
	Assets.Button.List.Importance = widget.HighImportance
	Assets.Button.List.Hide()
	Assets.Button.Send.Hide()

	entry := widget.NewEntry()
	entry.SetPlaceHolder("Claim NFA:")

	claim_button := widget.NewButton("Claim", func() {
		if len(entry.Text) == 64 {
			if isNFA(entry.Text) {
				if tx := rpc.ClaimNFA(entry.Text); tx != "" {
					go ShowTxDialog("Claim NFA", fmt.Sprintf("TX: %s", tx), tx, 3*time.Second, d.Window)
				} else {
					dialog.NewInformation("Claim NFA", "TX Error", d.Window).Show()
				}

				return
			}

			dialog.NewInformation("Claim NFA", "Could not validate SCID as NFA", d.Window).Show()
			return
		}

		dialog.NewInformation("Claim NFA", "Not a valid SCID", d.Window).Show()
	})

	Assets.Claim = container.NewBorder(nil, nil, nil, claim_button, entry)
	Assets.Claim.Hide()

	Assets.Button.Rescan = widget.NewButton("Rescan", func() {
		go func() {
			Assets.Button.scanning = true
			Assets.Button.Rescan.Hide()
			Assets.Button.List.Hide()
			Assets.Button.Send.Hide()
			rescan()
			Assets.Button.scanning = false
		}()
	})
	Assets.Button.Rescan.Importance = widget.LowImportance
	Assets.Button.Rescan.Hide()

	return container.NewBorder(
		nil,
		container.NewAdaptiveGrid(2,
			Assets.Claim,
			container.NewAdaptiveGrid(5, layout.NewSpacer(), container.NewStack(layout.NewSpacer(), Assets.Button.Rescan), layout.NewSpacer(), Assets.Button.Send, Assets.Button.List)),
		nil,
		nil,
		Assets.List)
}
