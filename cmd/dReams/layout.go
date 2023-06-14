package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SixofClubsss/dReams/baccarat"
	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/derbnb"
	"github.com/SixofClubsss/dReams/dwidget"
	"github.com/SixofClubsss/dReams/holdero"
	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/prediction"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/tarot"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
)

var H dwidget.DreamsItems
var B dwidget.DreamsItems
var P dwidget.DreamsItems
var S dwidget.DreamsItems
var T dwidget.DreamsItems

// If dReams has not been initialized, show this screen
//   - User selects dApps and skin to load
func introScreen() *fyne.Container {
	dReams.configure = true
	title := canvas.NewText("Welcome to dReams", bundle.TextColor)
	title.Alignment = fyne.TextAlignCenter
	title.TextSize = 18

	var max *fyne.Container
	skin_title := canvas.NewText("Choose your Skin", bundle.TextColor)
	skin_title.Alignment = fyne.TextAlignCenter
	skin_title.TextSize = 18

	skins := widget.NewRadioGroup([]string{"Dark", "Light"}, func(s string) {
		if s == "Light" {
			bundle.AppColor = color.White
		} else {
			bundle.AppColor = color.Black
		}

		dReams.App.Settings().SetTheme(bundle.DeroTheme(bundle.AppColor))
		max.Objects[1].(*fyne.Container).Objects[1].(*canvas.Text).Color = bundle.TextColor
		max.Objects[1].(*fyne.Container).Objects[1].Refresh()
		max.Objects[1].(*fyne.Container).Objects[6].(*canvas.Text).Color = bundle.TextColor
		max.Objects[1].(*fyne.Container).Objects[6].Refresh()
		max.Objects[1].(*fyne.Container).Objects[9].(*canvas.Text).Color = bundle.TextColor
		max.Objects[1].(*fyne.Container).Objects[9].Refresh()
		if bundle.AppColor == color.White {
			max.Objects[0] = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x99})
		} else {
			max.Objects[0] = canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
		}
		max.Objects[0].Refresh()
	})

	skins.Horizontal = true
	skins.Required = true

	dapp_title := canvas.NewText("Choose dApps to add to your dReams", bundle.TextColor)
	dapp_title.Alignment = fyne.TextAlignCenter
	dapp_title.TextSize = 18

	dapp_label := widget.NewLabel("dReams base app has:\n\nHoldero\n\nBaccarat\n\nNFA Marketplace")
	dapp_label.Wrapping = fyne.TextWrapWord
	dapp_label.Alignment = fyne.TextAlignCenter

	default_dapps := []string{"NFA Market"}
	default_checks := widget.NewCheckGroup(default_dapps, nil)
	default_checks.SetSelected(default_dapps)
	default_checks.Disable()

	dApps := rpc.FetchDapps()
	dapp_checks := widget.NewCheckGroup(dApps, nil)

	gnomon_gif, _ := xwidget.NewAnimatedGifFromResource(bundle.ResourceGnomonGifGif)
	gnomon_gif.SetMinSize(fyne.NewSize(100, 100))

	start_button := widget.NewButton("Start dReams", func() {
		menu.Control.Dapp_list = make(map[string]bool)

		for _, name := range dApps {
			menu.Control.Dapp_list[name] = false
		}

		for _, name := range dapp_checks.Selected {
			menu.Control.Dapp_list[name] = true
		}

		log.Println("[dReams] Loading dApps")
		go func() {
			dReams.App.Settings().SetTheme(bundle.DeroTheme(bundle.AppColor))
			dReams.Window.SetContent(
				container.New(layout.NewMaxLayout(),
					dReams.background,
					place()))
		}()
	})

	start_button.Importance = widget.LowImportance

	dreams_img := canvas.NewImageFromResource(bundle.ResourceBlueBadge3Png)
	dreams_img.SetMinSize(fyne.NewSize(100, 100))

	powered_label := widget.NewLabel("Powered by")
	powered_label.Alignment = fyne.TextAlignCenter

	gnomon_gif.Start()

	intro := container.NewVBox(
		layout.NewSpacer(),
		title,
		container.NewCenter(dreams_img),
		powered_label,
		container.NewCenter(gnomon_gif),
		layout.NewSpacer(),
		skin_title,
		container.NewCenter(skins),
		layout.NewSpacer(),
		dapp_title,
		container.NewCenter(container.NewVBox(default_checks, dapp_checks)),
		layout.NewSpacer(),
		layout.NewSpacer(),
		start_button)

	max = container.NewMax(bundle.Alpha180, intro)

	return max
}

// Select dApps to add or remove from dReams
//   - User can change current dApps and skin
func dAppScreen(reset fyne.CanvasObject) *fyne.Container {
	dReams.configure = true
	config_title := canvas.NewText("Configure your dReams", bundle.TextColor)
	config_title.Alignment = fyne.TextAlignCenter
	config_title.TextSize = 18

	is_enabled := []string{}
	enabled_dapps := make(map[string]bool)

	gnomon_gif, _ := xwidget.NewAnimatedGifFromResource(bundle.ResourceGnomonGifGif)
	gnomon_gif.SetMinSize(fyne.NewSize(100, 100))

	back_button := widget.NewButton("Back", func() {
		dReams.configure = false
		gnomon_gif.Stop()
		menu.RestartGif(menu.Gnomes.Icon_ind)
		go func() {
			dReams.Window.Content().(*fyne.Container).Objects[1] = reset
			dReams.Window.Content().(*fyne.Container).Objects[1].Refresh()
		}()
	})

	for name, enabled := range menu.Control.Dapp_list {
		enabled_dapps[name] = enabled
		if enabled {
			is_enabled = append(is_enabled, name)
		}
	}

	var current_skin, skin_choice color.Gray16
	load_button := widget.NewButton("Load Changes", func() {
		rpc.Wallet.Connected(false)
		rpc.Wallet.Height = 0
		menu.Disconnected()
		holdero.InitTableSettings()
		menu.Control.Dapp_list = enabled_dapps
		log.Println("[dReams] Loading dApps")
		menu.CloseAppSignal(true)
		menu.Gnomes.Checked(false)
		bundle.AppColor = skin_choice
		gnomon_gif.Stop()
		go func() {
			time.Sleep(1500 * time.Millisecond)
			menu.CloseAppSignal(false)
			dReams.App.Settings().SetTheme(bundle.DeroTheme(bundle.AppColor))
			dReams.Window.Content().(*fyne.Container).Objects[1] = place()
			dReams.Window.Content().(*fyne.Container).Objects[1].Refresh()
		}()
	})

	load_button.Importance = widget.LowImportance
	back_button.Importance = widget.LowImportance

	var dapps_changed bool
	dApps := rpc.FetchDapps()
	dapp_checks := widget.NewCheckGroup(dApps, func(s []string) {
		for reset := range enabled_dapps {
			enabled_dapps[reset] = false
		}

		for _, name := range s {
			enabled_dapps[name] = true
		}

		if reflect.DeepEqual(enabled_dapps, menu.Control.Dapp_list) {
			dapps_changed = false
			if current_skin == skin_choice {
				load_button.Hide()
				back_button.Show()
			}
		} else {
			dapps_changed = true
			load_button.Show()
			back_button.Hide()
		}
	})

	dapp_checks.SetSelected(is_enabled)

	default_dapps := []string{"NFA Market"}
	default_checks := widget.NewCheckGroup(default_dapps, nil)
	default_checks.SetSelected(default_dapps)
	default_checks.Disable()

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
				back_button.Show()
			} else {
				load_button.Show()
				back_button.Hide()
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

	load_button.Hide()
	back_button.Show()

	dreams_img := canvas.NewImageFromResource(bundle.ResourceBlueBadge3Png)
	dreams_img.SetMinSize(fyne.NewSize(100, 100))

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

	intro := container.NewVBox(
		layout.NewSpacer(),
		config_title,
		container.NewCenter(dreams_img),
		layout.NewSpacer(),
		skin_title,
		container.NewCenter(skins),
		layout.NewSpacer(),
		layout.NewSpacer(),
		dapp_title,
		changes_label,
		container.NewCenter(container.NewVBox(default_checks, dapp_checks)),
		layout.NewSpacer(),
		gnomon_label,
		container.NewCenter(gnomon_gif),
		loading_label,
		layout.NewSpacer(),
		container.NewAdaptiveGrid(2, container.NewMax(load_button), back_button))

	alpha := canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
	if bundle.AppColor == color.White {
		alpha = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x99})
	}

	return container.NewMax(alpha, intro)
}

// Main dReams layout
func place() *fyne.Container {
	H.LeftLabel = widget.NewLabel("")
	H.RightLabel = widget.NewLabel("")
	H.TopLabel = canvas.NewText(rpc.Display.Res, color.White)
	H.TopLabel.Move(fyne.NewPos(387, 204))
	H.LeftLabel.SetText("Seats: " + rpc.Display.Seats + "      Pot: " + rpc.Display.Pot + "      Blinds: " + rpc.Display.Blinds + "      Ante: " + rpc.Display.Ante + "      Dealer: " + rpc.Display.Dealer)
	H.RightLabel.SetText(rpc.Display.Readout + "      Player ID: " + rpc.Display.PlayerId + "      Dero Balance: " + rpc.Display.Balance["Dero"] + "      Height: " + rpc.Display.Wallet_height)

	B.LeftLabel = widget.NewLabel("")
	B.RightLabel = widget.NewLabel("")
	B.LeftLabel.SetText("Total Hands Played: " + rpc.Display.Total_w + "      Player Wins: " + rpc.Display.Player_w + "      Ties: " + rpc.Display.Ties + "      Banker Wins: " + rpc.Display.Banker_w + "      Min Bet is " + rpc.Display.BaccMin + " dReams, Max Bet is " + rpc.Display.BaccMax)
	B.RightLabel.SetText("dReams Balance: " + rpc.Display.Balance["dReams"] + "      Dero Balance: " + rpc.Display.Balance["Dero"] + "      Height: " + rpc.Display.Wallet_height)

	P.LeftLabel = widget.NewLabel("")
	P.RightLabel = widget.NewLabel("")
	P.RightLabel.SetText("dReams Balance: " + rpc.Display.Balance["dReams"] + "      Dero Balance: " + rpc.Display.Balance["Dero"] + "      Height: " + rpc.Display.Wallet_height)

	prediction.Predict.Info = widget.NewLabel("SCID:\n\n" + prediction.Predict.Contract + "\n")
	prediction.Predict.Info.Wrapping = fyne.TextWrapWord
	prediction.Predict.Prices = widget.NewLabel("")

	S.LeftLabel = widget.NewLabel("")
	S.RightLabel = widget.NewLabel("")
	S.RightLabel.SetText("dReams Balance: " + rpc.Display.Balance["dReams"] + "      Dero Balance: " + rpc.Display.Balance["Dero"] + "      Height: " + rpc.Display.Wallet_height)

	T.LeftLabel = widget.NewLabel("")
	T.RightLabel = widget.NewLabel("")
	T.LeftLabel.SetText("Total Readings: " + rpc.Display.Readings + "      Click your card for Iluma reading")
	T.RightLabel.SetText("dReams Balance: " + rpc.Display.Balance["dReams"] + "      Dero Balance: " + rpc.Display.Balance["Dero"] + "      Height: " + rpc.Display.Wallet_height)

	prediction.Sports.Info = widget.NewLabel("SCID:\n\n" + prediction.Sports.Contract + "\n")
	prediction.Sports.Info.Wrapping = fyne.TextWrapWord

	// dReams menu tabs
	menu_tabs := container.NewAppTabs(
		container.NewTabItem("Wallet", placeWall()),
		container.NewTabItem("dApps", layout.NewSpacer()),
		container.NewTabItem("Assets", menu.PlaceAssets("dReams", true, menu.RecheckDreamsAssets, bundle.ResourceDReamsIconAltPng, dReams.Window)),
		container.NewTabItem("Market", menu.PlaceMarket()))

	menu_tabs.OnSelected = func(ti *container.TabItem) {
		MenuTab(ti)
		if ti.Text == "dApps" {
			if menu.Gnomes.IsScanning() {
				menu_tabs.SelectIndex(0)
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
	top_bar := container.NewVBox(container.NewMax(top), layout.NewSpacer())

	menu_bottom := canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
	menu_bottom.SetMinSize(fyne.NewSize(268, 37))
	menu_bottom_box := container.NewHBox(menu_bottom, layout.NewSpacer())
	menu_bottom_bar := container.NewVBox(layout.NewSpacer(), menu_bottom_box)

	tarot_bottom := canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
	tarot_bottom.SetMinSize(fyne.NewSize(136, 37))
	tarot_bottom_box := container.NewHBox(tarot_bottom, layout.NewSpacer())
	tarot_bottom_bar := container.NewVBox(layout.NewSpacer(), tarot_bottom_box)
	tarot_bottom.Hide()

	alpha_box := container.NewMax(top_bar, menu_bottom_bar, tarot_bottom_bar, bundle.Alpha150)
	if dReams.os != "darwin" {
		alpha_box.Objects = append(alpha_box.Objects, FullScreenSet())
	}
	alpha_box.Objects = append(alpha_box.Objects, menu.StartDreamsIndicators())

	tabs := container.NewAppTabs(container.NewTabItem("Menu", menu_tabs))

	if menu.Control.Dapp_list["Holdero"] {
		var holdero_objs *fyne.Container
		var contract_objs *container.Split
		contract_change_screen := widget.NewButton("Contracts", nil)
		contract_change_screen.OnTapped = func() {
			go func() {
				dReams.menu_tabs.contracts = true
				dReams.Window.Content().(*fyne.Container).Objects[1].(*fyne.Container).Objects[1].(*container.AppTabs).Selected().Content = contract_objs
				dReams.Window.Content().(*fyne.Container).Objects[1].(*fyne.Container).Objects[1].(*container.AppTabs).Selected().Content.Refresh()
			}()
		}

		dReams.menu_tabs.contracts = true
		holdero_objs = placeHoldero(contract_change_screen)
		contract_objs = placeContract(holdero_objs)

		tabs.Append(container.NewTabItem("Holdero", contract_objs))
	}

	if menu.Control.Dapp_list["Baccarat"] {
		tabs.Append(container.NewTabItem("Baccarat", placeBacc()))
	}

	if menu.Control.Dapp_list["dSports and dPredictions"] {
		tabs.Append(container.NewTabItem("Predict", placePredict()))
		tabs.Append(container.NewTabItem("Sports", placeSports()))
	}

	if menu.Control.Dapp_list["Iluma"] {
		tabs.Append(container.NewTabItem("Iluma", placeTarot()))
	}

	if menu.Control.Dapp_list["DerBnb"] {
		tabs.Append(container.NewTabItem("DerBnb", derbnb.LayoutAllItems(true, dReams.Window, dReams.background)))
	}

	if dReams.cli {
		exitTerminal()
		tabs.Append(container.NewTabItem("Cli", startTerminal()))
	}

	tabs.Append(container.NewTabItem("Log", rpc.SessionLog()))

	tabs.OnSelected = func(ti *container.TabItem) {
		MainTab(ti)
		if ti.Text == "Menu" {
			menu_bottom.Show()
			menu_tabs.Items[0].Content.(*container.Split).Leading.(*container.Split).Trailing.Refresh()
		} else {
			menu_bottom.Hide()
		}

		if ti.Text == "Iluma" {
			tarot_bottom.Show()
		} else {
			tarot_bottom.Hide()
		}
	}

	dReams.configure = false

	return container.NewMax(alpha_box, tabs)
}

// dReams wallet layout
func placeWall() *container.Split {
	daemon_cont := container.NewHScroll(menu.DaemonRpcEntry())
	daemon_cont.SetMinSize(fyne.NewSize(340, 35.1875))

	user_input_cont := container.NewVBox(
		daemon_cont,
		menu.WalletRpcEntry(),
		menu.UserPassEntry(),
		menu.RpcConnectButton(),
		layout.NewSpacer(),
		menu.MenuDisplay())

	menu.Control.Contract_rating = make(map[string]uint64)
	menu.Assets.Asset_map = make(map[string]string)

	daemon_check_cont := container.NewVBox(menu.DaemonConnectedBox())

	user_input_box := container.NewHBox(user_input_cont, daemon_check_cont)
	connect_tabs := container.NewAppTabs(container.NewTabItem("Connect", container.NewCenter(user_input_box)))

	menu_top := container.NewHSplit(container.NewMax(bundle.Alpha120, menu.IntroTree()), connect_tabs)
	menu_top.SetOffset(0.66)

	menu_bottom := container.NewAdaptiveGrid(1, placeSwap())
	menu_box := container.NewVSplit(menu_top, menu_bottom)
	menu_box.SetOffset(0.5)

	return menu_box
}

func placeSwap() *container.Split {
	pair_opts := []string{"DERO-dReams", "dReams-DERO"}
	select_pair := widget.NewSelect(pair_opts, nil)
	select_pair.PlaceHolder = "Pairs"
	select_pair.SetSelectedIndex(0)

	assets := []string{}
	for asset := range rpc.Display.Balance {
		assets = append(assets, asset)
	}

	sort.Strings(assets)

	menu.Assets.Balances = widget.NewList(
		func() int {
			return len(assets)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(assets[i] + fmt.Sprintf(": %s", rpc.Display.Balance[assets[i]]))
		})

	balance_tabs := container.NewAppTabs(container.NewTabItem("Balances", menu.Assets.Balances))

	var swap_entry *dwidget.DeroAmts
	var swap_boxes *fyne.Container

	max := container.NewMax()
	swap_tabs := container.NewAppTabs()

	swap_button := widget.NewButton("Swap", nil)
	swap_button.OnTapped = func() {
		switch select_pair.Selected {
		case "DERO-dReams":
			f, err := strconv.ParseFloat(swap_entry.Text, 64)
			if err == nil && swap_entry.Validate() == nil {
				if amt := (f * 333) * 100000; amt > 0 {
					max.Objects[0] = holdero.DreamsConfirm(1, amt, max, swap_tabs)
					max.Refresh()
				}
			}
		case "dReams-DERO":
			f, err := strconv.ParseFloat(swap_entry.Text, 64)
			if err == nil && swap_entry.Validate() == nil {
				if amt := f * 100000; amt > 0 {
					max.Objects[0] = holdero.DreamsConfirm(2, amt, max, swap_tabs)
					max.Refresh()
				}
			}
		}
	}

	swap_entry, swap_boxes = menu.CreateSwapContainer(select_pair.Selected)
	menu.Assets.Swap = container.NewBorder(select_pair, swap_button, nil, nil, swap_boxes)
	menu.Assets.Swap.Hide()

	select_pair.OnChanged = func(s string) {
		split := strings.Split(s, "-")
		if len(split) != 2 {
			return
		}

		swap_entry, swap_boxes = menu.CreateSwapContainer(s)

		menu.Assets.Swap.Objects[0] = swap_boxes
		menu.Assets.Swap.Refresh()
	}

	alpha := canvas.NewRectangle(color.RGBA{0, 0, 0, 120})
	if bundle.AppColor == color.White {
		alpha = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x55})
	}

	swap_tabs = container.NewAppTabs(container.NewTabItem("Swap", container.NewCenter(menu.Assets.Swap)))
	max.Add(swap_tabs)

	full := container.NewHSplit(container.NewMax(alpha, balance_tabs), max)
	full.SetOffset(0.66)

	return full
}

// Holdero contract tab layout
func placeContract(change_screen *fyne.Container) *container.Split {
	check_box := container.NewVBox(menu.HolderoContractConnectedBox())

	var tabs *container.AppTabs
	menu.Poker.Holdero_unlock = widget.NewButton("Unlock Holdero Contract", nil)
	menu.Poker.Holdero_unlock.Hide()

	menu.Poker.Holdero_new = widget.NewButton("New Holdero Table", nil)
	menu.Poker.Holdero_new.Hide()

	unlock_cont := container.NewVBox(
		layout.NewSpacer(),
		menu.Poker.Holdero_unlock,
		menu.Poker.Holdero_new)

	owner_buttons := container.NewAdaptiveGrid(2, container.NewMax(layout.NewSpacer()), unlock_cont)
	owned_tab := container.NewBorder(nil, owner_buttons, nil, nil, menu.MyTables())

	tabs = container.NewAppTabs(
		container.NewTabItem("Tables", layout.NewSpacer()),
		container.NewTabItem("Favorites", menu.HolderoFavorites()),
		container.NewTabItem("Owned", owned_tab),
		container.NewTabItem("View", layout.NewSpacer()))

	tabs.SelectIndex(0)
	tabs.Selected().Content = menu.TableListings(tabs)

	tabs.OnSelected = func(ti *container.TabItem) {
		MenuContractTab(ti)
		if ti.Text == "View" {
			go func() {
				if len(rpc.Round.Contract) == 64 {
					rpc.FetchHolderoSC()
					dReams.menu_tabs.contracts = false
					dReams.Window.Content().(*fyne.Container).Objects[1].(*fyne.Container).Objects[1].(*container.AppTabs).Selected().Content = change_screen
					dReams.Window.Content().(*fyne.Container).Objects[1].(*fyne.Container).Objects[1].(*container.AppTabs).Selected().Content.Refresh()
					tabs.SelectIndex(0)
					now := time.Now().Unix()
					if now > rpc.Round.Last+33 {
						HolderoRefresh()
					}
				} else {
					tabs.SelectIndex(0)
				}
			}()
		}
	}

	max := container.NewMax(bundle.Alpha120, tabs)

	menu.Poker.Holdero_unlock.OnTapped = func() {
		max.Objects[1] = menu.HolderoMenuConfirm(1, max.Objects, tabs)
		max.Objects[1].Refresh()
	}

	menu.Poker.Holdero_new.OnTapped = func() {
		max.Objects[1] = menu.HolderoMenuConfirm(2, max.Objects, tabs)
		max.Objects[1].Refresh()
	}

	mid := container.NewVBox(layout.NewSpacer(), container.NewAdaptiveGrid(2, menu.NameEntry(), holdero.TournamentButton(max.Objects, tabs)), menu.OwnersBoxMid())

	menu_bottom := container.NewGridWithColumns(3, menu.OwnersBoxLeft(max.Objects, tabs), mid, layout.NewSpacer())

	contract_cont := container.NewHScroll(menu.HolderoContractEntry())
	contract_cont.SetMinSize(fyne.NewSize(640, 35.1875))

	asset_items := container.NewAdaptiveGrid(1, container.NewVBox(menu.TableStats()))

	player_input := container.NewVBox(
		contract_cont,
		asset_items,
		layout.NewSpacer())

	player_box := container.NewHBox(player_input, check_box)
	menu_top := container.NewHSplit(player_box, max)

	menuBox := container.NewVSplit(menu_top, menu_bottom)
	menuBox.SetOffset(1)

	return menuBox
}

// dReams Holdero tab layout
func placeHoldero(change_screen *widget.Button) *fyne.Container {
	H.Back = *container.NewWithoutLayout(
		holdero.HolderoTable(bundle.ResourcePokerTablePng),
		holdero.Player1_label(nil, nil, nil),
		holdero.Player2_label(nil, nil, nil),
		holdero.Player3_label(nil, nil, nil),
		holdero.Player4_label(nil, nil, nil),
		holdero.Player5_label(nil, nil, nil),
		holdero.Player6_label(nil, nil, nil),
		H.TopLabel)

	holdero_label := container.NewHBox(H.LeftLabel, layout.NewSpacer(), H.RightLabel)

	H.Front = *placeHolderoCards()

	H.Actions = *container.NewVBox(
		layout.NewSpacer(),
		holdero.SitButton(),
		holdero.LeaveButton(),
		holdero.DealHandButton(),
		holdero.CheckButton(),
		holdero.BetButton(),
		holdero.BetAmount())

	options := container.NewVBox(layout.NewSpacer(), holdero.AutoOptions(), change_screen)

	holdero_actions := container.NewHBox(options, layout.NewSpacer(), holdero.TimeOutWarning(), layout.NewSpacer(), layout.NewSpacer(), &H.Actions)

	H.DApp = container.NewVBox(
		labelColorBlack(holdero_label),
		&H.Back,
		&H.Front,
		layout.NewSpacer(),
		holdero_actions)

	return H.DApp
}

// dReams Baccarat tab layout
func placeBacc() *fyne.Container {
	B.Back = *container.NewWithoutLayout(
		baccarat.BaccTable(bundle.ResourceBaccTablePng),
		baccarat.BaccResult(rpc.Display.BaccRes))

	B.Front = *clearBaccCards()

	bacc_label := container.NewHBox(B.LeftLabel, layout.NewSpacer(), B.RightLabel)

	B.DApp = container.NewVBox(
		labelColorBlack(bacc_label),
		&B.Back,
		&B.Front,
		layout.NewSpacer(),
		baccarat.BaccaratButtons(dReams.Window))

	return B.DApp
}

// dReams dPrediction tab layout
func placePredict() *fyne.Container {
	predict_info := container.NewVBox(prediction.Predict.Info, prediction.Predict.Prices)
	predict_scroll := container.NewScroll(predict_info)
	predict_scroll.SetMinSize(fyne.NewSize(540, 500))

	check_box := container.NewVBox(prediction.PredictConnectedBox())

	contract_scroll := container.NewHScroll(prediction.PredictionContractEntry())
	contract_scroll.SetMinSize(fyne.NewSize(600, 35.1875))
	contract_cont := container.NewHBox(contract_scroll, check_box)

	prediction.Predict.Higher = widget.NewButton("Higher", nil)
	prediction.Predict.Higher.Hide()

	prediction.Predict.Lower = widget.NewButton("Lower", nil)
	prediction.Predict.Lower.Hide()

	prediction.Predict.Prediction_box = container.NewVBox(prediction.Predict.Higher, prediction.Predict.Lower)
	prediction.Predict.Prediction_box.Hide()

	predict_content := container.NewVBox(
		contract_cont,
		predict_scroll,
		layout.NewSpacer(),
		prediction.Predict.Prediction_box)

	menu.Control.Bet_unlock_p = widget.NewButton("Unlock dPrediction Contract", nil)
	menu.Control.Bet_unlock_p.Hide()

	menu.Control.Bet_new_p = widget.NewButton("New dPrediction Contract", nil)
	menu.Control.Bet_new_p.Hide()

	unlock_cont := container.NewVBox(menu.Control.Bet_unlock_p, menu.Control.Bet_new_p)

	owner_buttons := container.NewAdaptiveGrid(2, container.NewMax(prediction.OwnerButtonP()), unlock_cont)
	owned_tab := container.NewBorder(nil, owner_buttons, nil, nil, prediction.PredictionOwned())

	tabs := container.NewAppTabs(
		container.NewTabItem("Contracts", layout.NewSpacer()),
		container.NewTabItem("Favorites", prediction.PredictionFavorites()),
		container.NewTabItem("Owned", owned_tab))

	tabs.SelectIndex(0)
	tabs.Selected().Content = prediction.PredictionListings(tabs)

	tabs.OnSelected = func(ti *container.TabItem) {
		PredictTab(ti)
	}

	max := container.NewMax(bundle.Alpha120, tabs)

	prediction.Predict.Higher.OnTapped = func() {
		if len(prediction.Predict.Contract) == 64 {
			max.Objects[1] = prediction.ConfirmAction(2, "", "", max.Objects, tabs)
			max.Objects[1].Refresh()
		}
	}

	prediction.Predict.Lower.OnTapped = func() {
		if len(prediction.Predict.Contract) == 64 {
			max.Objects[1] = prediction.ConfirmAction(1, "", "", max.Objects, tabs)
			max.Objects[1].Refresh()
		}
	}

	menu.Control.Bet_unlock_p.OnTapped = func() {
		max.Objects[1] = menu.BettingMenuConfirmP(1, max.Objects, tabs)
		max.Objects[1].Refresh()
	}

	menu.Control.Bet_new_p.OnTapped = func() {
		max.Objects[1] = menu.BettingMenuConfirmP(2, max.Objects, tabs)
		max.Objects[1].Refresh()
	}

	predict_label := container.NewHBox(P.LeftLabel, layout.NewSpacer(), P.RightLabel)
	predict_box := container.NewHSplit(predict_content, max)

	P.DApp = container.NewVBox(
		labelColorBlack(predict_label),
		predict_box)

	go func() {
		time.Sleep(2 * time.Second)
		for !menu.ClosingApps() && menu.Control.Dapp_list["dSports and dPredictions"] {
			if !rpc.Wallet.IsConnected() && !rpc.Signal.Startup {
				menu.Control.Predict_check.SetChecked(false)
				menu.Control.Sports_check.SetChecked(false)
				prediction.DisablePredictions(true)
				prediction.DisableSports(true)
			}
			time.Sleep(time.Second)
		}
	}()

	return P.DApp
}

// dReams dSports tab layout
func placeSports() *fyne.Container {
	sports_content := container.NewVBox(prediction.Sports.Info)
	sports_scroll := container.NewVScroll(sports_content)
	sports_scroll.SetMinSize(fyne.NewSize(180, 500))

	check_box := container.NewVBox(prediction.SportsConnectedBox())

	contract_scroll := container.NewHScroll(prediction.SportsContractEntry())
	contract_scroll.SetMinSize(fyne.NewSize(600, 35.1875))
	contract_cont := container.NewHBox(contract_scroll, check_box)

	prediction.Sports.Game_select = widget.NewSelect(prediction.Sports.Game_options, func(s string) {
		split := strings.Split(s, "   ")
		a, b := menu.GetSportsTeams(prediction.Sports.Contract, split[0])
		if prediction.Sports.Game_select.SelectedIndex() >= 0 {
			prediction.Sports.Multi.Show()
			prediction.Sports.ButtonA.Show()
			prediction.Sports.ButtonB.Show()
			prediction.Sports.ButtonA.Text = a
			prediction.Sports.ButtonA.Refresh()
			prediction.Sports.ButtonB.Text = b
			prediction.Sports.ButtonB.Refresh()
		} else {
			prediction.Sports.Multi.Hide()
			prediction.Sports.ButtonA.Hide()
			prediction.Sports.ButtonB.Hide()
		}
	})

	prediction.Sports.Game_select.PlaceHolder = "Select Game #"
	prediction.Sports.Game_select.Hide()

	var Multi_options = []string{"1x", "3x", "5x"}
	prediction.Sports.Multi = widget.NewRadioGroup(Multi_options, func(s string) {})
	prediction.Sports.Multi.Horizontal = true
	prediction.Sports.Multi.Hide()

	prediction.Sports.ButtonA = widget.NewButton("TEAM A", nil)
	prediction.Sports.ButtonA.Hide()

	prediction.Sports.ButtonB = widget.NewButton("TEAM B", nil)
	prediction.Sports.ButtonB.Hide()

	sports_multi := container.NewCenter(prediction.Sports.Multi)
	prediction.Sports.Sports_box = container.NewVBox(
		sports_multi,
		prediction.Sports.Game_select,
		prediction.Sports.ButtonA,
		prediction.Sports.ButtonB)

	prediction.Sports.Sports_box.Hide()

	sports_left := container.NewVBox(
		contract_cont,
		sports_scroll,
		layout.NewSpacer(),
		prediction.Sports.Sports_box)

	epl := widget.NewLabel("")
	epl.Wrapping = fyne.TextWrapWord
	epl_scroll := container.NewVScroll(epl)
	mls := widget.NewLabel("")
	mls.Wrapping = fyne.TextWrapWord
	mls_scroll := container.NewVScroll(mls)
	nba := widget.NewLabel("")
	nba.Wrapping = fyne.TextWrapWord
	nba_scroll := container.NewVScroll(nba)
	nfl := widget.NewLabel("")
	nfl.Wrapping = fyne.TextWrapWord
	nfl_scroll := container.NewVScroll(nfl)
	nhl := widget.NewLabel("")
	nhl.Wrapping = fyne.TextWrapWord
	nhl_scroll := container.NewVScroll(nhl)
	mlb := widget.NewLabel("")
	mlb.Wrapping = fyne.TextWrapWord
	mlb_scroll := container.NewVScroll(mlb)
	bellator := widget.NewLabel("")
	bellator.Wrapping = fyne.TextWrapWord
	bellator_scroll := container.NewVScroll(bellator)
	ufc := widget.NewLabel("")
	ufc.Wrapping = fyne.TextWrapWord
	ufc_scroll := container.NewVScroll(ufc)
	score_tabs := container.NewAppTabs(
		container.NewTabItem("EPL", epl_scroll),
		container.NewTabItem("MLS", mls_scroll),
		container.NewTabItem("NBA", nba_scroll),
		container.NewTabItem("NFL", nfl_scroll),
		container.NewTabItem("NHL", nhl_scroll),
		container.NewTabItem("MLB", mlb_scroll),
		container.NewTabItem("Bellator", bellator_scroll),
		container.NewTabItem("UFC", ufc_scroll))

	score_tabs.OnSelected = func(ti *container.TabItem) {
		switch ti.Text {
		case "EPL":
			go prediction.GetScores(epl, "EPL")
		case "MLS":
			go prediction.GetScores(mls, "MLS")
		case "NBA":
			go prediction.GetScores(nba, "NBA")
		case "NFL":
			go prediction.GetScores(nfl, "NFL")
		case "NHL":
			go prediction.GetScores(nhl, "NHL")
		case "MLB":
			go prediction.GetScores(mlb, "MLB")
		case "Bellator":
			go prediction.GetMmaResults(bellator, "Bellator")
		case "UFC":
			go prediction.GetMmaResults(ufc, "UFC")
		default:
		}
	}

	menu.Control.Bet_unlock_s = widget.NewButton("Unlock dSports Contracts", nil)
	menu.Control.Bet_unlock_s.Hide()

	menu.Control.Bet_new_s = widget.NewButton("New dSports Contract", nil)
	menu.Control.Bet_new_s.Hide()

	unlock_cont := container.NewVBox(
		menu.Control.Bet_unlock_s,
		menu.Control.Bet_new_s)

	owner_buttons := container.NewAdaptiveGrid(2, container.NewMax(prediction.OwnerButtonS()), unlock_cont)
	owned_tab := container.NewBorder(nil, owner_buttons, nil, nil, prediction.SportsOwned())

	tabs := container.NewAppTabs(
		container.NewTabItem("Contracts", layout.NewSpacer()),
		container.NewTabItem("Favorites", prediction.SportsFavorites()),
		container.NewTabItem("Owned", owned_tab),
		container.NewTabItem("Scores", score_tabs),
		container.NewTabItem("Payouts", prediction.SportsPayouts()))

	tabs.SelectIndex(0)
	tabs.Selected().Content = prediction.SportsListings(tabs)

	tabs.OnSelected = func(ti *container.TabItem) {

	}

	max := container.NewMax(bundle.Alpha120, tabs)

	prediction.Sports.ButtonA.OnTapped = func() {
		if len(prediction.Sports.Contract) == 64 {
			max.Objects[1] = prediction.ConfirmAction(3, prediction.Sports.ButtonA.Text, prediction.Sports.ButtonB.Text, max.Objects, tabs)
			max.Objects[1].Refresh()
		}
	}
	prediction.Sports.ButtonA.Hide()

	prediction.Sports.ButtonB.OnTapped = func() {
		if len(prediction.Sports.Contract) == 64 {
			max.Objects[1] = prediction.ConfirmAction(4, prediction.Sports.ButtonA.Text, prediction.Sports.ButtonB.Text, max.Objects, tabs)
			max.Objects[1].Refresh()
		}
	}

	menu.Control.Bet_unlock_s.OnTapped = func() {
		max.Objects[1] = menu.BettingMenuConfirmS(1, max.Objects, tabs)
		max.Objects[1].Refresh()
	}

	menu.Control.Bet_new_s.OnTapped = func() {
		max.Objects[1] = menu.BettingMenuConfirmS(2, max.Objects, tabs)
		max.Objects[1].Refresh()
	}

	sports_label := container.NewHBox(S.LeftLabel, layout.NewSpacer(), S.RightLabel)
	sports_box := container.NewHSplit(sports_left, max)

	S.DApp = container.NewVBox(
		labelColorBlack(sports_label),
		sports_box)

	return S.DApp
}

// dReams Tarot tab layout
func placeTarot() fyne.CanvasObject {
	search_entry := widget.NewEntry()
	search_entry.SetPlaceHolder("TXID:")
	search_button := widget.NewButton("    Search   ", func() {
		txid := search_entry.Text
		if len(txid) == 64 {
			signer := rpc.VerifySigner(search_entry.Text)
			if signer {
				rpc.Tarot.Display = true
				tarot.Iluma.Label.SetText("")
				rpc.FetchTarotReading(txid)
				if rpc.Tarot.Card2 != 0 && rpc.Tarot.Card3 != 0 {
					tarot.Iluma.Card1.Objects[1] = TarotCard(rpc.Tarot.Card1)
					tarot.Iluma.Card2.Objects[1] = TarotCard(rpc.Tarot.Card2)
					tarot.Iluma.Card3.Objects[1] = TarotCard(rpc.Tarot.Card3)
					rpc.Tarot.Num = 3
				} else {
					tarot.Iluma.Card1.Objects[1] = TarotCard(0)
					tarot.Iluma.Card2.Objects[1] = TarotCard(rpc.Tarot.Card1)
					tarot.Iluma.Card3.Objects[1] = TarotCard(0)
					rpc.Tarot.Num = 1
				}
				tarot.Iluma.Box.Refresh()
			} else {
				log.Println("[Iluma] This is not your reading")
			}
		}
	})

	tarot_label := container.NewHBox(T.LeftLabel, layout.NewSpacer(), T.RightLabel)

	T.DApp = container.NewBorder(
		labelColorBlack(tarot_label),
		nil,
		nil,
		nil,
		tarot.TarotCardBox())

	reset := tarot.Iluma.Card2

	tarot.Iluma.Draw1 = widget.NewButton("Draw One", func() {
		if !tarot.Iluma.Open {
			tarot.Iluma.Draw1.Hide()
			tarot.Iluma.Draw3.Hide()
			tarot.Iluma.Card2 = *tarot.TarotConfirm(1, reset)
		}
	})

	tarot.Iluma.Draw3 = widget.NewButton("Draw Three", func() {
		if !tarot.Iluma.Open {
			tarot.Iluma.Draw1.Hide()
			tarot.Iluma.Draw3.Hide()
			tarot.Iluma.Card2 = *tarot.TarotConfirm(3, reset)
		}
	})

	tarot.Iluma.Draw1.Hide()
	tarot.Iluma.Draw3.Hide()

	draw_cont := container.NewAdaptiveGrid(5,
		layout.NewSpacer(),
		layout.NewSpacer(),
		tarot.Iluma.Draw1,
		tarot.Iluma.Draw3,
		layout.NewSpacer())

	tarot.Iluma.Search = container.NewBorder(nil, nil, nil, search_button, search_entry)

	tarot.Iluma.Actions = container.NewVBox(
		layout.NewSpacer(),
		container.NewAdaptiveGrid(2, draw_cont, tarot.Iluma.Search))

	tarot.Iluma.Search.Hide()
	tarot.Iluma.Actions.Hide()

	tarot_tabs := container.NewAppTabs(
		container.NewTabItem("Iluma", tarot.PlaceIluma()),
		container.NewTabItem("Reading", T.DApp))

	tarot_tabs.OnSelected = func(ti *container.TabItem) {
		TarotTab(ti)
	}

	tarot_tabs.SetTabLocation(container.TabLocationBottom)

	return container.NewMax(tarot_tabs, tarot.Iluma.Actions)
}
