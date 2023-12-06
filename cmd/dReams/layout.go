package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"reflect"
	"time"

	"github.com/SixofClubsss/Baccarat/baccarat"
	"github.com/SixofClubsss/Duels/duel"
	"github.com/SixofClubsss/Grokked/grok"
	"github.com/SixofClubsss/Holdero/holdero"
	"github.com/SixofClubsss/Iluma/tarot"
	"github.com/SixofClubsss/dPrediction/prediction"
	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
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

var indicators []menu.DreamsIndicator

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

	dApps := rpc.FetchDapps()
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
		menu.Control.Dapp_list = make(map[string]bool)

		for _, name := range dApps {
			menu.Control.Dapp_list[name] = false
		}

		menu.Control.Dapp_list = enabled_dapps

		dReams.SetChannels(menu.EnabledDappCount())
		logger.Println("[dReams] Loading dApps")
		go func() {
			dReams.App.Settings().SetTheme(bundle.DeroTheme(bundle.AppColor))
			dReams.Window.SetContent(
				container.NewStack(
					dReams.Background,
					place()))
			wait = false
		}()
	})

	start_button.Importance = widget.LowImportance

	dreams_img := canvas.NewImageFromResource(bundle.ResourceBlueBadge3Png)
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
		menu.RestartGif(menu.Gnomes.Icon_ind)
		go func() {
			dReams.Window.Content().(*fyne.Container).Objects[1] = reset
			dReams.Window.Content().(*fyne.Container).Objects[1].Refresh()
		}()
	})

	for name, enabled := range menu.Control.Dapp_list {
		enabled_dapps[name] = enabled
	}

	var wait bool
	var current_skin, skin_choice color.Gray16
	load_button := widget.NewButton("Load Changes", func() {
		if wait {
			return
		}

		wait = true
		rpc.Wallet.Connected(false)
		rpc.Wallet.Height = 0

		status_text := canvas.NewText("Closing dApps...", color.NRGBA{R: 0xf0, G: 0xf0, B: 0xf0, A: 0xaa})
		status_text.TextSize = 21
		status_text.Alignment = fyne.TextAlignCenter

		img := canvas.NewImageFromResource(bundle.ResourceFigure1CirclePng)
		img.SetMinSize(fyne.NewSize(180, 180))

		dReams.Window.Content().(*fyne.Container).Objects[1] = container.NewStack(container.NewCenter(container.NewBorder(nil, status_text, nil, nil, img)), widget.NewProgressBarInfinite())
		dReams.Window.Content().(*fyne.Container).Objects[1].Refresh()

		logger.Println("[dReams] Closing dApps")
		dReams.CloseAllDapps()
		disconnected()
		menu.Control.Dapp_list = enabled_dapps
		dReams.SetChannels(menu.EnabledDappCount())
		menu.CloseAppSignal(true)
		menu.Gnomes.Checked(false)
		bundle.AppColor = skin_choice
		gnomon_gif.Stop()
		status_text.Text = "Loading dApps..."
		status_text.Refresh()
		go func() {
			time.Sleep(1500 * time.Millisecond)
			menu.CloseAppSignal(false)
			logger.Println("[dReams] Loading dApps")
			dReams.App.Settings().SetTheme(bundle.DeroTheme(bundle.AppColor))
			dReams.Window.Content().(*fyne.Container).Objects[1] = place()
			dReams.Window.Content().(*fyne.Container).Objects[1].Refresh()
			wait = false
		}()
	})

	load_button.Importance = widget.LowImportance
	back_button.Importance = widget.LowImportance

	var dapps_changed bool
	dApps := rpc.FetchDapps()
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

				if reflect.DeepEqual(enabled_dapps, menu.Control.Dapp_list) {
					dapps_changed = false
					if current_skin == skin_choice {
						load_button.Hide()
					}
				} else {
					dapps_changed = true
					load_button.Show()
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

// Main dReams layout
func place() *fyne.Container {
	menu.Control.Contract_rating = make(map[string]uint64)
	menu.Assets.Asset_map = make(map[string]string)

	asset_selects := []fyne.Widget{
		holdero.FaceSelect(),
		holdero.BackSelect(),
		dreams.ThemeSelect(),
		holdero.AvatarSelect(menu.Assets.Asset_map),
		holdero.SharedDecks(),
		recheckButton("dReams", recheckDreamsAssets),
	}

	var intros []menu.IntroText
	intros = append(intros, menu.MakeMenuIntro(holdero.DreamsMenuIntro())...)
	intros = append(intros, menu.MakeMenuIntro(baccarat.DreamsMenuIntro())...)
	intros = append(intros, menu.MakeMenuIntro(prediction.DreamsMenuIntro())...)
	intros = append(intros, menu.MakeMenuIntro(tarot.DreamsMenuIntro())...)
	// intros = append(intros, menu.MakeMenuIntro(derbnb.DreamsMenuIntro())...)
	intros = append(intros, menu.MakeMenuIntro(grok.DreamsMenuIntro())...)
	intros = append(intros, menu.MakeMenuIntro(duel.DreamsMenuIntro())...)

	// dReams menu tabs
	menu_tabs := container.NewAppTabs(
		container.NewTabItem("Wallet", placeWall(intros)),
		container.NewTabItem("dApps", layout.NewSpacer()),
		container.NewTabItem("Assets", menu.PlaceAssets("dReams", asset_selects, bundle.ResourceDReamsIconAltPng, dReams.Window)),
		container.NewTabItem("Market", menu.PlaceMarket()))

	menu_tabs.OnSelected = func(ti *container.TabItem) {
		switch ti.Text {
		case "Wallet":
			ti.Content.(*container.Split).Leading.(*container.Split).Leading.Refresh()
			dReams.SetSubTab("Wallet")
		case "Assets":
			dReams.SetSubTab("Assets")
			menu.Control.Viewing_asset = ""
			menu.Assets.Asset_list.UnselectAll()
			menu_tabs.Selected().Content.(*container.Split).Leading.(*container.Split).Trailing.(*fyne.Container).Objects[1].(*container.AppTabs).SelectIndex(0)
		case "Market":
			dReams.SetSubTab("Market")
			go menu.FindNFAListings(nil)
			menu.Market.Cancel_button.Hide()
			menu.Market.Close_button.Hide()
			menu.Market.Auction_list.Refresh()
			menu.Market.Buy_list.Refresh()
		case "dApps":
			if menu.Gnomes.IsScanning() {
				menu_tabs.SelectIndex(0)
				dialog.NewInformation("Gnomon Syncing", "Please wait to make dApp changes", dReams.Window).Show()
			} else {
				go func() {
					reset := dReams.Window.Content().(*fyne.Container).Objects[1]
					dapp_screen := dAppScreen(reset)
					dReams.Window.Content().(*fyne.Container).Objects[1] = dapp_screen
					dReams.Window.Content().(*fyne.Container).Objects[1].Refresh()
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
		tabs.Append(container.NewTabItem("Holdero", holdero.LayoutAllItems(&dReams)))
		indicators = append(indicators, holdero.HolderoIndicator())
	}

	if menu.DappEnabled("Baccarat") {
		tabs.Append(container.NewTabItem("Baccarat", baccarat.LayoutAllItems(&dReams)))
	}

	if menu.DappEnabled("dSports and dPredictions") {
		tabs.Append(container.NewTabItem("Predict", prediction.LayoutPredictItems(&dReams)))
		tabs.Append(container.NewTabItem("Sports", prediction.LayoutSportsItems(&dReams)))
		indicators = append(indicators, prediction.ServiceIndicator())
	}

	if menu.DappEnabled("Iluma") {
		tabs.Append(container.NewTabItem("Iluma", tarot.LayoutAllItems(&dReams)))
	}

	// // Under development
	// if menu.DappEnabled("DerBnb") {
	// 	tabs.Append(container.NewTabItem("DerBnb", derbnb.LayoutAllItems(true, &dReams)))
	// }

	if menu.DappEnabled("Duels") {
		tabs.Append(container.NewTabItem("Duels", duel.LayoutAllItems(menu.Assets.Asset_map, &dReams)))
	}

	if menu.DappEnabled("Grokked") {
		tabs.Append(container.NewTabItem("Grokked", grok.LayoutAllItems(&dReams)))
	}

	if cli.enabled {
		exitTerminal()
		tabs.Append(container.NewTabItem("Cli", startTerminal()))
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
	alpha_box.Objects = append(alpha_box.Objects, menu.StartDreamsIndicators(indicators))

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
	daemon_cont := container.NewHScroll(daemonRpcEntry())
	daemon_cont.SetMinSize(fyne.NewSize(340, 35.1875))

	user_input_cont := container.NewVBox(
		daemon_cont,
		walletRpcEntry(),
		userPassEntry(),
		rpcConnectButton(),
		layout.NewSpacer(),
		menu.InfoDisplay())

	daemon_check_cont := container.NewVBox(daemonConnectedBox())

	user_input_box := container.NewHBox(user_input_cont, daemon_check_cont)
	connect_tabs := container.NewAppTabs(
		container.NewTabItem("Connect", container.NewCenter(user_input_box)),
		container.NewTabItem("Gnomon", container.NewCenter(menu.Gnomes.ControlPanel(dReams.Window))))

	connect_tabs.OnSelected = func(ti *container.TabItem) {
		if ti.Text == "Gnomon" {
			if rpc.Daemon.IsConnected() {
				status := menu.Gnomes.Status()
				if (menu.Gnomes.Start && status != "indexing") || menu.Gnomes.IsScanning() {
					if status == "indexed" {
						status = "scanning wallet"
					}
					dialog.NewInformation("Gnomon Syncing", fmt.Sprintf("%s, please wait...", status), dReams.Window).Show()
					connect_tabs.SelectIndex(0)
				} else if menu.Gnomes.IsInitialized() {
					dialog.NewConfirm("Gnomon Running", "Shut down Gnomon to make changes", func(b bool) {
						if b {
							daemon_cont.Content.(*widget.SelectEntry).SetText("")
							daemon_check_cont.Objects[0].(*widget.Check).SetChecked(false)
						} else {
							connect_tabs.SelectIndex(0)
						}
					}, dReams.Window).Show()
				}
			}
		}
	}

	menu_top := container.NewHSplit(container.NewStack(bundle.Alpha120, menu.IntroTree(intros)), connect_tabs)
	menu_top.SetOffset(0.66)

	menu_bottom := container.NewAdaptiveGrid(1, holdero.PlaceSwap(&dReams))
	menu_box := container.NewVSplit(menu_top, menu_bottom)
	menu_box.SetOffset(0.5)

	return menu_box
}
