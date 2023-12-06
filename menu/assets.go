package menu

import (
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
	"github.com/dReam-dApps/dReams/rpc"
)

type assetObjects struct {
	Swap         *fyne.Container
	Balances     *widget.List
	Index_entry  *widget.Entry
	Index_button *widget.Button
	Index_search *widget.Button
	Asset_list   *widget.List
	Assets       []string
	Asset_map    map[string]string
	Name         *canvas.Text
	Collection   *canvas.Text
	Icon         canvas.Image
	Stats_box    fyne.Container
	Header_box   fyne.Container
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

func (a *assetObjects) Add(name, scid string) {
	a.Assets = append(a.Assets, name+"   "+scid)
	a.Asset_map[name] = scid
}

// Icon image for Holdero tables and asset viewing
//   - Pass res as frame resource
func IconImg(res fyne.Resource) *fyne.Container {
	Assets.Icon.SetMinSize(fyne.NewSize(100, 100))
	Assets.Icon.Resize(fyne.NewSize(94, 94))
	Assets.Icon.Move(fyne.NewPos(7, 3))

	frame := canvas.NewImageFromResource(res)
	frame.Resize(fyne.NewSize(100, 100))
	frame.Move(fyne.NewPos(4, 0))

	return container.NewWithoutLayout(&Assets.Icon, frame)
}

// Display for owned asset info
func AssetStats() fyne.CanvasObject {
	Assets.Collection = canvas.NewText(" Collection: ", bundle.TextColor)
	Assets.Name = canvas.NewText(" Name: ", bundle.TextColor)

	Assets.Name.TextSize = 18
	Assets.Collection.TextSize = 18

	Assets.Stats_box = *container.NewVBox(Assets.Collection, Assets.Name, IconImg(nil))

	return &Assets.Stats_box
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

	label := canvas.NewText("You will need to delete Gnomon DB and resync for changes to take effect ", bundle.TextColor)
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
//   - assets is array of widgets used for asset selections
//   - menu_icon resources for side menus
//   - w for main window dialog
func PlaceAssets(tag string, assets []fyne.Widget, menu_icon fyne.Resource, w fyne.Window) *container.Split {
	items_box := container.NewAdaptiveGrid(2)

	asset_selects := container.NewVBox()
	for _, sel := range assets {
		asset_selects.Add(sel)
	}
	asset_selects.Add(layout.NewSpacer())

	cont := container.NewHScroll(asset_selects)
	cont.SetMinSize(fyne.NewSize(290, 35.1875))

	items_box.Add(cont)

	items_box.Add(container.NewAdaptiveGrid(1, AssetStats()))

	player_input := container.NewVBox(items_box, layout.NewSpacer())

	enable_opts := EnabledCollections(false)

	tabs := container.NewAppTabs(
		container.NewTabItem("Owned", AssetList()))

	if len(asset_selects.Objects) > 1 {
		tabs.Append(container.NewTabItem("Enabled", enable_opts))
	}

	tabs.OnSelected = func(ti *container.TabItem) {
		if ti.Text == "Enabled" {
			if rpc.Daemon.IsConnected() {
				dialog.NewInformation("Assets", "Shut down Gnomon to make changes to asset index", w).Show()
				tabs.Selected().Content = container.NewVScroll(container.NewVBox(dwidget.NewCenterLabel("Currently Enabled:"), dwidget.NewCenterLabel(returnEnabledNames(Control.Enabled_assets))))

				return
			}
			tabs.Selected().Content = enable_opts
		}
	}

	scroll_top := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "arrowUp"), func() {
		Assets.Asset_list.ScrollToTop()
	})

	scroll_bottom := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "arrowDown"), func() {
		Assets.Asset_list.ScrollToBottom()
	})

	scroll_top.Importance = widget.LowImportance
	scroll_bottom.Importance = widget.LowImportance

	scroll_cont := container.NewVBox(container.NewHBox(layout.NewSpacer(), scroll_top, scroll_bottom))

	max := container.NewStack(bundle.Alpha120, tabs, scroll_cont)

	header_name_entry := widget.NewEntry()
	header_name_entry.PlaceHolder = "Name:"
	header_descr_entry := widget.NewEntry()
	header_descr_entry.PlaceHolder = "Description"
	header_icon_entry := widget.NewEntry()
	header_icon_entry.PlaceHolder = "Icon:"

	header_button := widget.NewButton("Set Headers", func() {
		scid := Assets.Index_entry.Text
		if len(scid) == 64 && header_name_entry.Text != "dReam Tables" && header_name_entry.Text != "dReams" {
			if _, ok := rpc.FindStringKey(rpc.GnomonSCID, scid, rpc.Daemon.Rpc).(string); ok {
				max.Objects[1] = setHeaderConfirm(header_name_entry.Text, header_descr_entry.Text, header_icon_entry.Text, scid, max.Objects, tabs)
				max.Objects[1].Refresh()
			} else {
				dialog.NewInformation("Check back soon", "SCID not stored on the main Gnomon SC yet\n\nOnce stored, you can set your SCID headers", w).Show()
			}
		}
	})

	header_contr := container.NewVBox(header_name_entry, header_descr_entry, header_icon_entry, header_button)
	Assets.Header_box = *container.NewAdaptiveGrid(2, header_contr)
	Assets.Header_box.Hide()

	player_input.Add(&Assets.Header_box)

	player_box := container.NewHBox(player_input)

	menu_top := container.NewHSplit(player_box, max)
	menu_bottom := container.NewAdaptiveGrid(1, IndexEntry(menu_icon, w))

	menu_box := container.NewVSplit(menu_top, menu_bottom)
	menu_box.SetOffset(1)

	return menu_box
}

// Confirmation for setting SCID headers
//   - name, desc and icon of SCID header
//   - Pass main window obj to reset to
func setHeaderConfirm(name, desc, icon, scid string, obj []fyne.CanvasObject, reset *container.AppTabs) fyne.CanvasObject {
	label := widget.NewLabel("Headers for SCID:\n\n" + scid + "\n\nName: " + name + "\n\nDescription: " + desc + "\n\nIcon: " + icon)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	confirm_button := widget.NewButtonWithIcon("Confirm", dreams.FyneIcon("confirm"), func() {
		rpc.SetHeaders(name, desc, icon, scid)
		obj[1] = reset
		obj[1].Refresh()
	})
	confirm_button.Importance = widget.HighImportance

	cancel_button := widget.NewButtonWithIcon("Cancel", dreams.FyneIcon("cancel"), func() {
		obj[1] = reset
		obj[1].Refresh()

	})

	alpha := container.NewStack(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	buttons := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewVBox(layout.NewSpacer(), label, layout.NewSpacer(), buttons)

	return container.NewStack(alpha, content)
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
			if isNFA(Assets.Index_entry.Text) {
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

	Info.Indexed = canvas.NewText("Indexed SCIDs: ", bundle.TextColor)
	Info.Indexed.TextSize = 18

	bottom_grid := container.NewAdaptiveGrid(3, Info.Indexed, Assets.Index_button, Assets.Index_search)
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
