package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"reflect"
	"strings"
	"time"

	"github.com/SixofClubsss/Baccarat/baccarat"
	"github.com/SixofClubsss/Duels/duel"
	"github.com/SixofClubsss/Grokked/grok"
	"github.com/SixofClubsss/Holdero/holdero"
	"github.com/SixofClubsss/Iluma/tarot"
	"github.com/SixofClubsss/dDice/dice"
	"github.com/SixofClubsss/dPrediction/prediction"
	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/gnomes"
	"github.com/dReam-dApps/dReams/menu"
	"github.com/dReam-dApps/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
)

var indicators []*menu.DreamsIndicator

// Boot splash screen
func splashScreen() fyne.CanvasObject {
	text := dwidget.NewCanvasText("Initializing...", 21, fyne.TextAlignCenter)
	text.Color = color.White

	img := canvas.NewImageFromResource(bundle.ResourceFigure1CirclePng)
	img.SetMinSize(fyne.NewSize(180, 180))

	return container.NewStack(dReams.Background, container.NewCenter(img, text), widget.NewProgressBarInfinite())
}

// If dReams has not been initialized, show this screen
//   - User selects dApps and skin to load
func introScreen() *fyne.Container {
	dReams.Configure(true)
	title := canvas.NewText("Welcome to dReams", bundle.TextColor)
	title.Alignment = fyne.TextAlignCenter
	title.TextSize = 18

	var max *fyne.Container
	skin_title := canvas.NewText("Choose your Skin", bundle.TextColor)
	skin_title.Alignment = fyne.TextAlignCenter
	skin_title.TextSize = 18

	skins := widget.NewRadioGroup([]string{"Dark", "Light"}, nil)
	switch bundle.AppColor {
	case color.White:
		skins.SetSelected("Light")
	case color.Black:
		skins.SetSelected("Dark")
	default:
		skins.SetSelected("Dark")
	}

	skins.Horizontal = true
	skins.Required = true

	skins.OnChanged = func(s string) {
		if s == "Light" {
			bundle.AppColor = color.White
		} else {
			bundle.AppColor = color.Black
		}

		dReams.App.Settings().SetTheme(bundle.DeroTheme(bundle.AppColor))
		max.Objects[1].(*container.Split).Leading.(*fyne.Container).Objects[2].(*canvas.Text).Color = bundle.TextColor
		max.Objects[1].(*container.Split).Leading.(*fyne.Container).Objects[2].Refresh()
		max.Objects[1].(*container.Split).Leading.(*fyne.Container).Objects[7].(*canvas.Text).Color = bundle.TextColor
		max.Objects[1].(*container.Split).Leading.(*fyne.Container).Objects[7].Refresh()
		max.Objects[1].(*container.Split).Leading.(*fyne.Container).Objects[9].(*canvas.Text).Color = bundle.TextColor
		max.Objects[1].(*container.Split).Leading.(*fyne.Container).Objects[9].Refresh()
		max.Objects[1].(*container.Split).Trailing.(*fyne.Container).Objects[1].(*canvas.Text).Color = bundle.TextColor
		max.Objects[1].(*container.Split).Trailing.(*fyne.Container).Objects[1].Refresh()
		max.Objects[0] = bundle.NewAlpha180()
		max.Objects[0].Refresh()
	}

	dapp_title := canvas.NewText("Choose dApps to add to your dReams", bundle.TextColor)
	dapp_title.Alignment = fyne.TextAlignCenter
	dapp_title.TextSize = 18

	collection_title := canvas.NewText("Enable asset collections", bundle.TextColor)
	collection_title.Alignment = fyne.TextAlignCenter
	collection_title.TextSize = 18

	dApps := rpc.GetDapps()
	enabled_dapps := make(map[string]bool)

	versions := dappVersions(dApps)

	default_dapps := []string{"Gnomon", "NFA Market"}
	dApps = append(default_dapps, dApps...)

	dapp_checks := widget.NewListWithData(
		binding.BindStringList(&dApps),
		func() fyne.CanvasObject {
			check := widget.NewCheck("", nil)
			check.OnChanged = func(b bool) {
				if b {
					enabled_dapps[check.Text] = true
				} else {
					enabled_dapps[check.Text] = false
				}
			}

			return container.NewAdaptiveGrid(2, check, widget.NewLabel(""))
		},
		func(di binding.DataItem, c fyne.CanvasObject) {
			dat := di.(binding.String)
			str, err := dat.Get()
			if err != nil {
				return
			}

			// Defaults
			if str == "Gnomon" || str == "NFA Market" {
				c.(*fyne.Container).Objects[0].(*widget.Check).OnChanged = nil
				c.(*fyne.Container).Objects[0].(*widget.Check).Disable()
				c.(*fyne.Container).Objects[0].(*widget.Check).SetText(str)
				c.(*fyne.Container).Objects[0].(*widget.Check).SetChecked(true)
				c.(*fyne.Container).Objects[1].(*widget.Label).SetText(versions[str])
				if str == "NFA Market" {
					enabled_dapps[str] = true
				}
				return
			}

			c.(*fyne.Container).Objects[0].(*widget.Check).SetText(str)
			c.(*fyne.Container).Objects[1].(*widget.Label).SetText(versions[str])
		})

	dapp_checks.OnSelected = func(id widget.ListItemID) {
		dapp_checks.Unselect(id)
	}

	dapp_spacer := canvas.NewRectangle(color.Transparent)
	dapp_spacer.SetMinSize(fyne.NewSize(500, 310))
	dapp_box := container.NewStack(dapp_spacer, dapp_checks)

	gnomon_gif, _ := xwidget.NewAnimatedGifFromResource(bundle.ResourceGnomonGifGif)
	gnomon_gif.SetMinSize(fyne.NewSize(100, 100))

	var wait bool
	start_button := widget.NewButton("Start dReams", func() {
		if wait {
			return
		}

		wait = true
		menu.Control.Lock()
		menu.Control.Dapps = make(map[string]bool)

		for _, name := range dApps {
			if name != "NFA Market" {
				menu.Control.Dapps[name] = false
			}
		}

		menu.Control.Dapps = enabled_dapps
		menu.Control.Unlock()

		gnomon_gif.Stop()
		gnomon_gif = nil

		dReams.SetChannels(menu.EnabledDappCount())
		logger.Println("[dReams] Loading dApps")
		go func() {
			dReams.App.Settings().SetTheme(bundle.DeroTheme(bundle.AppColor))
			dReams.Window.SetContent(container.NewStack(dReams.Background, place()))
			if !dReams.Window.FullScreen() {
				dReams.Window.Resize(fyne.NewSize(dreams.MIN_WIDTH, dreams.MIN_HEIGHT))
			}
			wait = false
		}()
	})

	start_button.Importance = widget.LowImportance

	dreams_img := canvas.NewImageFromResource(bundle.ResourceFigure1CirclePng)
	dreams_img.SetMinSize(fyne.NewSize(90, 90))

	powered_label := widget.NewLabel("Powered by")
	powered_label.Alignment = fyne.TextAlignCenter

	gnomon_gif.Start()

	collections_spacer := canvas.NewRectangle(color.Transparent)
	collections_spacer.SetMinSize(fyne.NewSize(10, 750))

	line := canvas.NewLine(bundle.TextColor)
	line_spacer := canvas.NewRectangle(color.Transparent)
	line_spacer.SetMinSize(fyne.NewSize(300, 0))

	under_circle := canvas.NewRadialGradient(color.White, color.Transparent)
	under_circle.SetMinSize(fyne.NewSize(120, 120))

	over_circle := canvas.NewRadialGradient(color.White, color.Transparent)
	over_circle.SetMinSize(fyne.NewSize(130, 130))

	intro := container.NewHSplit(
		container.NewVBox(
			layout.NewSpacer(),
			container.NewHBox(layout.NewSpacer(), container.NewVBox(line_spacer, line), layout.NewSpacer()),
			title,
			container.NewHBox(layout.NewSpacer(), container.NewVBox(line_spacer, line), layout.NewSpacer()),
			container.NewStack(container.NewCenter(under_circle, dreams_img, over_circle)),
			powered_label,
			container.NewCenter(gnomon_gif),
			skin_title,
			container.NewCenter(skins),
			dapp_title,
			container.NewCenter(container.NewVBox(dapp_box)),
			layout.NewSpacer(),
			container.NewHBox(layout.NewSpacer(), container.NewVBox(line_spacer, line), layout.NewSpacer()),
			container.NewCenter(start_button)),
		container.NewVBox(layout.NewSpacer(), collection_title, container.NewStack(collections_spacer, menu.EnabledCollections(true))))

	intro.SetOffset(0.66)

	max = container.NewStack(bundle.Alpha180, intro)

	return max
}

// Select dApps to add or remove from dReams
//   - User can change current dApps and skin
func dAppScreen(reset fyne.CanvasObject) *fyne.Container {
	dReams.Configure(true)
	config_title := canvas.NewText("Configure your dReams", bundle.TextColor)
	config_title.Alignment = fyne.TextAlignCenter
	config_title.TextSize = 18

	enabled_dapps := make(map[string]bool)

	gnomon_gif, _ := xwidget.NewAnimatedGifFromResource(bundle.ResourceGnomonGifGif)
	gnomon_gif.SetMinSize(fyne.NewSize(100, 100))

	back_button := widget.NewButton("Back", func() {
		dReams.Configure(false)
		gnomon_gif.Stop()
		gnomon_gif = nil
		menu.RestartGif(gnomes.Indicator.Icon)
		go func() {
			dReams.Window.Content().(*fyne.Container).Objects[1] = reset
		}()
	})

	menu.Control.RLock()
	for name, enabled := range menu.Control.Dapps {
		enabled_dapps[name] = enabled
	}
	menu.Control.RUnlock()

	var wait bool
	var current_skin, skin_choice color.Gray16
	load_button := widget.NewButton("Load Changes", func() {
		if wait {
			return
		}

		wait = true
		rpc.Wallet.CloseConnections("dReams")

		status_text := dwidget.NewCanvasText("Closing dApps...", 21, fyne.TextAlignCenter)
		status_text.Color = color.White

		img := canvas.NewImageFromResource(bundle.ResourceFigure1CirclePng)
		img.SetMinSize(fyne.NewSize(180, 180))

		dReams.Window.Content().(*fyne.Container).Objects[1] = container.NewStack(container.NewCenter(img, status_text), widget.NewProgressBarInfinite())

		logger.Println("[dReams] Closing dApps")
		dReams.CloseAllDapps()
		disconnected()
		menu.Control.Lock()
		menu.Control.Dapps = enabled_dapps
		menu.Control.Unlock()
		dReams.SetChannels(menu.EnabledDappCount())
		menu.SetClose(true)
		gnomon.Checked(false)
		bundle.AppColor = skin_choice
		gnomon_gif.Stop()
		gnomon_gif = nil
		status_text.Text = "Loading dApps..."
		status_text.Refresh()
		go func() {
			time.Sleep(1500 * time.Millisecond)
			menu.SetClose(false)
			logger.Println("[dReams] Loading dApps")
			dReams.App.Settings().SetTheme(bundle.DeroTheme(bundle.AppColor))
			dReams.Window.Content().(*fyne.Container).Objects[1] = place()
			if !dReams.Window.FullScreen() {
				dReams.Window.Resize(fyne.NewSize(dreams.MIN_WIDTH, dreams.MIN_HEIGHT))
			}
			wait = false
		}()
	})

	load_button.Importance = widget.LowImportance
	back_button.Importance = widget.LowImportance

	var dapps_changed bool
	dApps := rpc.GetDapps()
	versions := dappVersions(dApps)

	default_dapps := []string{"Gnomon", "NFA Market"}
	dApps = append(default_dapps, dApps...)

	dapp_checks := widget.NewListWithData(
		binding.BindStringList(&dApps),
		func() fyne.CanvasObject {
			check := widget.NewCheck("", nil)
			check.OnChanged = func(b bool) {
				if b {
					enabled_dapps[check.Text] = true
				} else {
					enabled_dapps[check.Text] = false
				}

				menu.Control.RLock()
				if reflect.DeepEqual(enabled_dapps, menu.Control.Dapps) {
					dapps_changed = false
					if current_skin == skin_choice {
						load_button.Hide()
					}
				} else {
					dapps_changed = true
					load_button.Show()
				}
				menu.Control.RUnlock()
			}

			return container.NewAdaptiveGrid(2, check, widget.NewLabel(""))
		},
		func(di binding.DataItem, c fyne.CanvasObject) {
			dat := di.(binding.String)
			str, err := dat.Get()
			if err != nil {
				return
			}

			// Defaults
			if str == "Gnomon" || str == "NFA Market" {
				c.(*fyne.Container).Objects[0].(*widget.Check).OnChanged = nil
				c.(*fyne.Container).Objects[0].(*widget.Check).Disable()
				c.(*fyne.Container).Objects[0].(*widget.Check).SetText(str)
				c.(*fyne.Container).Objects[0].(*widget.Check).SetChecked(true)
				c.(*fyne.Container).Objects[1].(*widget.Label).SetText(versions[str])
				return
			}

			c.(*fyne.Container).Objects[0].(*widget.Check).SetText(str)
			for name, b := range enabled_dapps {
				if b && name == str {
					c.(*fyne.Container).Objects[0].(*widget.Check).SetChecked(true)
				}
			}
			c.(*fyne.Container).Objects[1].(*widget.Label).SetText(versions[str])
		})

	dapp_checks.OnSelected = func(id widget.ListItemID) {
		dapp_checks.Unselect(id)
	}

	dapp_spacer := canvas.NewRectangle(color.Transparent)
	dapp_spacer.SetMinSize(fyne.NewSize(500, 310))
	dapp_box := container.NewStack(dapp_spacer, dapp_checks)

	skin_title := canvas.NewText("Skin", bundle.TextColor)
	skin_title.Alignment = fyne.TextAlignCenter
	skin_title.TextSize = 18

	skins := widget.NewRadioGroup([]string{"Dark", "Light"}, func(s string) {
		if s == "Light" {
			skin_choice = color.White
		} else {
			skin_choice = color.Black
		}

		if !dapps_changed {
			if skin_choice == current_skin {
				load_button.Hide()
			} else {
				load_button.Show()
			}
		}
	})

	skins.Horizontal = true
	skins.Required = true
	switch bundle.AppColor {
	case color.White:
		skins.SetSelected("Light")
		current_skin = color.White
	case color.Black:
		skins.SetSelected("Dark")
		current_skin = color.Black
	default:

	}

	line := canvas.NewLine(bundle.TextColor)
	line_spacer := canvas.NewRectangle(color.Transparent)
	line_spacer.SetMinSize(fyne.NewSize(300, 0))

	under_circle := canvas.NewRadialGradient(color.White, color.Transparent)
	under_circle.SetMinSize(fyne.NewSize(75, 75))

	over_circle := canvas.NewRadialGradient(color.White, color.Transparent)
	over_circle.SetMinSize(fyne.NewSize(85, 85))

	load_button.Hide()
	back_button.Show()

	title_img := canvas.NewImageFromResource(bundle.ResourceFigure1CirclePng)
	title_img.SetMinSize(fyne.NewSize(60, 60))

	gnomon_gif.Start()

	dapp_title := canvas.NewText("dApps", bundle.TextColor)
	dapp_title.Alignment = fyne.TextAlignCenter
	dapp_title.TextSize = 18

	changes_label := widget.NewLabel("Select dApps to add or remove from your dReams")
	changes_label.Wrapping = fyne.TextWrapWord
	changes_label.Alignment = fyne.TextAlignCenter

	gnomon_label := widget.NewLabel("Adding new dApps may require Gnomon DB resync to index the new contracts")
	gnomon_label.Wrapping = fyne.TextWrapWord
	gnomon_label.Alignment = fyne.TextAlignCenter

	loading_label := widget.NewLabel("Loading changes to dReams will disconnect your wallet")
	loading_label.Alignment = fyne.TextAlignCenter

	config_screen := container.NewVBox(
		config_title,
		container.NewStack(container.NewCenter(under_circle, title_img, over_circle)),
		skin_title,
		container.NewCenter(skins),
		dapp_title,
		changes_label,
		container.NewCenter(container.NewVBox(dapp_box)),
		gnomon_label,
		container.NewCenter(gnomon_gif),
		loading_label,
		container.NewHBox(layout.NewSpacer(), container.NewVBox(line_spacer, line), layout.NewSpacer()),
		container.NewCenter(container.NewHBox(container.NewStack(load_button), container.NewStack(back_button))))

	return container.NewStack(bundle.NewAlpha180(), config_screen)
}

// User profile layout with dreams.AssetSelects
func profile() fyne.CanvasObject {
	line := canvas.NewLine(bundle.TextColor)
	form := []*widget.FormItem{}
	form = append(form, widget.NewFormItem("Name", menu.NameEntry()))
	form = append(form, widget.NewFormItem("", layout.NewSpacer()))
	form = append(form, widget.NewFormItem("", container.NewVBox(line)))
	form = append(form, widget.NewFormItem("Avatar", holdero.AvatarSelect(menu.Assets.SCIDs)))
	form = append(form, widget.NewFormItem("Theme", menu.ThemeSelect(&dReams)))
	form = append(form, widget.NewFormItem("Card Deck", holdero.FaceSelect(menu.Assets.SCIDs, &dReams)))
	form = append(form, widget.NewFormItem("Card Back", holdero.BackSelect(menu.Assets.SCIDs)))
	form = append(form, widget.NewFormItem("Dice", dice.DiceSelect(menu.Assets.SCIDs)))
	form = append(form, widget.NewFormItem("", container.NewVBox(line)))

	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(450, 0))

	return container.NewCenter(container.NewBorder(spacer, nil, nil, nil, widget.NewForm(form...)))
}

// TODO move these
var connect_select *container.AppTabs
var menu_tabs *container.AppTabs
var asset_tab *fyne.Container

// Main dReams layout
func place() *fyne.Container {
	menu.Control.Ratings = make(map[string]uint64)

	var intros []menu.IntroText
	intros = append(intros, menu.MakeMenuIntro(holdero.DreamsMenuIntro())...)
	intros = append(intros, menu.MakeMenuIntro(baccarat.DreamsMenuIntro())...)
	intros = append(intros, menu.MakeMenuIntro(prediction.DreamsMenuIntro())...)
	intros = append(intros, menu.MakeMenuIntro(tarot.DreamsMenuIntro())...)
	// intros = append(intros, menu.MakeMenuIntro(derbnb.DreamsMenuIntro())...)
	intros = append(intros, menu.MakeMenuIntro(duel.DreamsMenuIntro())...)
	intros = append(intros, menu.MakeMenuIntro(grok.DreamsMenuIntro())...)
	intros = append(intros, menu.MakeMenuIntro(dice.DreamsMenuIntro())...)

	indicators = []*menu.DreamsIndicator{}

	// dReams menu tabs
	asset_tab = menu.PlaceAssets("dReams", profile(), rescan, bundle.ResourceDReamsIconAltPng, &dReams)
	menu_tabs = container.NewAppTabs(
		container.NewTabItem("Wallet", placeWall(intros)),
		container.NewTabItem("dApps", layout.NewSpacer()),
		container.NewTabItem("Assets", asset_tab),
		container.NewTabItem("Market", menu.PlaceMarket(&dReams)))

	menu_tabs.OnSelected = func(ti *container.TabItem) {
		switch ti.Text {
		case "Wallet":
			ti.Content.(*container.Split).Leading.(*container.Split).Leading.Refresh()
			dReams.SetSubTab("Wallet")
		case "Assets":
			dReams.SetSubTab("Assets")
			menu.Assets.Viewing = ""
			menu.Assets.List.UnselectAll()
			if _, ok := menu_tabs.Selected().Content.(*fyne.Container).Objects[1].(*container.AppTabs); ok {
				menu_tabs.Selected().Content.(*fyne.Container).Objects[1].(*container.AppTabs).SelectIndex(1)
			}
		case "Market":
			dReams.SetSubTab("Market")
			go menu.FindNFAListings(nil, nil)
			menu.Market.Button.Cancel.Hide()
			menu.Market.Button.Close.Hide()
			menu.Market.List.Auction.Refresh()
			menu.Market.List.Buy.Refresh()
		case "dApps":
			if gnomon.IsScanning() {
				menu_tabs.SelectIndex(0)
				dialog.NewInformation("Gnomon Syncing", "Wait to make dApp changes", dReams.Window).Show()
			} else if rpc.Wallet.WS.IsConnecting() {
				menu_tabs.SelectIndex(0)
				dialog.NewInformation("XSWD Request", "Close connection requests to make dApp changes", dReams.Window).Show()
			} else {
				go func() {
					reset := dReams.Window.Content().(*fyne.Container).Objects[1]
					dapp_screen := dAppScreen(reset)
					dReams.Window.Content().(*fyne.Container).Objects[1] = dapp_screen
					menu_tabs.SelectIndex(0)
				}()
			}
		}
	}

	menu_tabs.SetTabLocation(container.TabLocationBottom)

	top := canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
	top.SetMinSize(fyne.NewSize(465, 40))
	top_bar := container.NewVBox(container.NewStack(top), layout.NewSpacer())

	menu_bottom := canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
	menu_bottom.SetMinSize(fyne.NewSize(268, 37))
	menu_bottom_box := container.NewHBox(menu_bottom, layout.NewSpacer())
	menu_bottom_bar := container.NewVBox(layout.NewSpacer(), menu_bottom_box)

	tarot_bottom := canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
	tarot_bottom.SetMinSize(fyne.NewSize(136, 37))
	tarot_bottom_box := container.NewHBox(tarot_bottom, layout.NewSpacer())
	tarot_bottom_bar := container.NewVBox(layout.NewSpacer(), tarot_bottom_box)
	tarot_bottom.Hide()

	tabs := container.NewAppTabs(container.NewTabItem("Menu", menu_tabs))

	if menu.DappEnabled("Holdero") {
		tabs.Append(container.NewTabItem("Holdero", holdero.LayoutAll(&dReams)))
		indicators = append(indicators, holdero.HolderoIndicator())
	}

	if menu.DappEnabled("Baccarat") {
		tabs.Append(container.NewTabItem("Baccarat", baccarat.LayoutAll(&dReams)))
	}

	if menu.DappEnabled("dSports and dPredictions") {
		tabs.Append(container.NewTabItem("Predict", prediction.LayoutPredictions(&dReams)))
		tabs.Append(container.NewTabItem("Sports", prediction.LayoutSports(&dReams)))
		indicators = append(indicators, prediction.ServiceIndicator())
	}

	if menu.DappEnabled("Iluma") {
		tabs.Append(container.NewTabItem("Iluma", tarot.LayoutAll(&dReams)))
	}

	// // Under development
	// if menu.DappEnabled("DerBnb") {
	// 	tabs.Append(container.NewTabItem("DerBnb", derbnb.LayoutAll(true, &dReams)))
	// }

	if menu.DappEnabled("Duels") {
		tabs.Append(container.NewTabItem("Duels", duel.LayoutAll(menu.Assets.SCIDs, &dReams)))
	}

	if menu.DappEnabled("Grokked") {
		tabs.Append(container.NewTabItem("Grokked", grok.LayoutAll(&dReams)))
	}

	if menu.DappEnabled("Dice") {
		tabs.Append(container.NewTabItem("Dice", dice.LayoutAll(&dReams)))
	}

	tabs.Append(container.NewTabItem("Log", rpc.SessionLog(App_Name, rpc.Version())))

	var fs_button *widget.Button
	fs_button = widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "viewFullScreen"), func() {
		if dReams.Window.FullScreen() {
			fs_button.Icon = fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "viewFullScreen")
			dReams.Window.SetFullScreen(false)
		} else {
			fs_button.Icon = fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "viewRestore")
			dReams.Window.SetFullScreen(true)
		}
	})

	fs_button.Importance = widget.LowImportance

	alpha_box := container.NewStack(top_bar, menu_bottom_bar, tarot_bottom_bar, bundle.Alpha150)
	if dReams.OS() != "darwin" {
		alpha_box.Objects = append(alpha_box.Objects, container.NewHBox(layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), container.NewVBox(fs_button), layout.NewSpacer()))
	}
	alpha_box.Objects = append(alpha_box.Objects, menu.StartIndicators(indicators))

	tabs.OnSelected = func(ti *container.TabItem) {
		dReams.SetTab(ti.Text)
		switch ti.Text {
		case "Baccarat":
			baccarat.OnTabSelected(&dReams)
		}

		if ti.Text == "Menu" {
			holdero.Settings.EnableCardSelects()
			menu_bottom.Show()
			menu_tabs.Items[0].Content.(*container.Split).Leading.(*container.Split).Leading.Refresh()
		} else {
			menu_bottom.Hide()
		}

		if ti.Text == "Iluma" {
			tarot_bottom.Show()
		} else {
			tarot_bottom.Hide()
		}
	}

	dReams.Configure(false)

	return container.NewStack(alpha_box, tabs)
}

// dReams wallet layout
func placeWall(intros []menu.IntroText) *container.Split {
	daemon_entry := daemonRPCEntry()

	layoutRPC := container.NewVBox(layout.NewSpacer(), rpcConnection())
	layoutXSWD := container.NewVBox(layout.NewSpacer(), xswdConnection())
	layoutFile := container.NewVBox(layout.NewSpacer(), accountConnection())

	connect_select = container.NewAppTabs(
		container.NewTabItem("RPC", layoutRPC),
		container.NewTabItem("XSWD", layoutXSWD),
		container.NewTabItem("DERO", layoutFile))

	connect_select.SetTabLocation(container.TabLocationLeading)
	connect_select.OnSelected = func(ti *container.TabItem) {
		switch ti.Text {
		case "DERO":
			_, names := dreams.GetDeroAccounts()
			layoutFile.Objects[1].(*fyne.Container).Objects[1].(*widget.SelectEntry).SetOptions(names)
		}
	}

	daemon_check_cont := container.NewVBox(daemonConnectedBox())

	connect_tab := container.NewCenter(
		container.NewVBox(
			container.NewStack(
				dwidget.NewSpacer(0, 120),
				container.NewVBox(container.NewBorder(nil, nil, dwidget.NewSpacer(53, 0), nil, daemon_entry)),
				connect_select),
			menu.InfoDisplay()))

	connect_tabs := container.NewAppTabs(
		container.NewTabItem("Connect", connect_tab),
		container.NewTabItem("Gnomon", container.NewCenter(gnomon.ControlPanel(dReams.Window))))

	connect_tabs.OnSelected = func(ti *container.TabItem) {
		if ti.Text == "Gnomon" {
			if rpc.Daemon.IsConnected() {
				status := gnomon.Status()
				if (gnomon.IsStarting() && status != "indexing") || gnomon.IsScanning() {
					if status == "indexed" {
						status = "scanning wallet"
					}
					status = fmt.Sprintf("%s%s", strings.ToUpper(string(status[0])), status[1:])
					dialog.NewInformation("Gnomon Syncing", fmt.Sprintf("%s, please wait...", status), dReams.Window).Show()
					connect_tabs.SelectIndex(0)
				} else if gnomon.IsInitialized() {
					dialog.NewConfirm("Gnomon Running", "Shut down Gnomon to make changes", func(b bool) {
						if b {
							daemon_entry.(*widget.SelectEntry).SetText("")
							daemon_check_cont.Objects[0].(*widget.Check).SetChecked(false)
						} else {
							connect_tabs.SelectIndex(0)
						}
					}, dReams.Window).Show()
				}
			}
		}
	}

	menu_top := container.NewHSplit(menu.IntroTree(intros), connect_tabs)
	menu_top.SetOffset(0.66)

	menu_bottom := container.NewAdaptiveGrid(1, holdero.PlaceSwap(&dReams))
	menu_box := container.NewVSplit(menu_top, menu_bottom)
	menu_box.SetOffset(0.5)

	return menu_box
}
