package main

import (
	_ "embed"
	"image/color"
	"log"
	"reflect"
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
)

var H dwidget.DreamsItems
var B dwidget.DreamsItems
var P dwidget.DreamsItems
var S dwidget.DreamsItems
var T dwidget.DreamsItems

// If dReams has not been intialized, show this screen
func introScreen() *fyne.Container {
	dReams.configure = true
	title := canvas.NewText("Welcome to dReams", color.White)
	title.TextSize = 18

	intro_label := widget.NewLabel("dReams base app has:\n\nHoldero\n\nBaccarat\n\nNFA Marketplace\n\n\nSelect further dApps to add to your dReams")
	intro_label.Wrapping = fyne.TextWrapWord
	intro_label.Alignment = fyne.TextAlignCenter

	dApps := rpc.FetchDapps()
	dapp_checks := widget.NewCheckGroup(dApps, func(s []string) {})

	start_button := widget.NewButton("Start dReams", func() {
		menu.Control.Dapp_list = make(map[string]bool)
		for _, name := range dapp_checks.Selected {
			menu.Control.Dapp_list[name] = true
		}

		log.Println("[dReams] Loading dApps")
		go func() {
			menu.Control.Dapp_list["dReams"] = true
			dReams.Window.SetContent(
				container.New(layout.NewMaxLayout(),
					background,
					place()))
		}()
	})

	start_button.Importance = widget.LowImportance

	intro := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(title),
		layout.NewSpacer(),
		intro_label,
		layout.NewSpacer(),
		container.NewCenter(dapp_checks),
		layout.NewSpacer(),
		layout.NewSpacer(),
		start_button)

	max := container.NewMax(menu.Alpha180, intro)

	return max
}

// Select dApps to add or remove from dReams
func dAppScreen(reset fyne.CanvasObject) *fyne.Container {
	dReams.configure = true
	title := canvas.NewText("dReams dApps", color.White)
	title.TextSize = 18

	changes_label := widget.NewLabel("Select dApps to add or remove from your dReams\n\nLoading dApp changes will disconnect your wallet")
	changes_label.Wrapping = fyne.TextWrapWord
	changes_label.Alignment = fyne.TextAlignCenter

	is_enabled := []string{}
	enabled_dapps := make(map[string]bool)

	var dapp_checks *widget.CheckGroup
	back_button := widget.NewButton("Back", func() {
		dReams.configure = false
		menu.Control.Dapp_list["dReams"] = true
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

	load_button := widget.NewButton("Load Changes", func() {
		rpc.Wallet.Connect = false
		rpc.Wallet.Height = 0
		holdero.InitTableSettings()
		menu.Control.Dapp_list = enabled_dapps
		log.Println("[dReams] Loading dApps")
		menu.Exit_signal = true
		menu.Gnomes.Checked = false
		menu.Disconnected()
		go func() {
			time.Sleep(1500 * time.Millisecond)
			menu.Exit_signal = false
			dReams.Window.Content().(*fyne.Container).Objects[1] = place()
			dReams.Window.Content().(*fyne.Container).Objects[1].Refresh()
		}()
	})

	load_button.Hide()
	load_button.Importance = widget.LowImportance
	back_button.Importance = widget.LowImportance

	dApps := rpc.FetchDapps()
	dapp_checks = widget.NewCheckGroup(dApps, func(s []string) {
		for reset := range enabled_dapps {
			enabled_dapps[reset] = false
		}

		for _, name := range s {
			enabled_dapps[name] = true
		}

		if reflect.DeepEqual(enabled_dapps, menu.Control.Dapp_list) {
			load_button.Hide()
			back_button.Show()
		} else {
			load_button.Show()
			back_button.Hide()
		}
	})

	dapp_checks.SetSelected(is_enabled)

	intro := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(title),
		layout.NewSpacer(),
		changes_label,
		layout.NewSpacer(),
		container.NewCenter(dapp_checks),
		layout.NewSpacer(),
		layout.NewSpacer(),
		container.NewAdaptiveGrid(2, container.NewMax(load_button), back_button))

	max := container.NewMax(menu.Alpha180, intro)

	return max
}

// Main dReams layout
func place() *fyne.Container {
	H.LeftLabel = widget.NewLabel("")
	H.RightLabel = widget.NewLabel("")
	H.TopLabel = widget.NewLabel("")
	H.TopLabel.Move(fyne.NewPos(380, 194))
	H.LeftLabel.SetText("Seats: " + rpc.Display.Seats + "      Pot: " + rpc.Display.Pot + "      Blinds: " + rpc.Display.Blinds + "      Ante: " + rpc.Display.Ante + "      Dealer: " + rpc.Display.Dealer + "      Turn: " + rpc.Display.Turn)
	H.RightLabel.SetText(rpc.Display.Readout + "      Player ID: " + rpc.Display.PlayerId + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Display.Wallet_height)

	B.LeftLabel = widget.NewLabel("")
	B.RightLabel = widget.NewLabel("")
	B.LeftLabel.SetText("Total Hands Played: " + rpc.Display.Total_w + "      Player Wins: " + rpc.Display.Player_w + "      Ties: " + rpc.Display.Ties + "      Banker Wins: " + rpc.Display.Banker_w + "      Min Bet is " + rpc.Display.BaccMin + " dReams, Max Bet is " + rpc.Display.BaccMax)
	B.RightLabel.SetText("dReams Balance: " + rpc.Display.Token_balance + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Display.Wallet_height)

	P.LeftLabel = widget.NewLabel("")
	P.RightLabel = widget.NewLabel("")
	P.RightLabel.SetText("dReams Balance: " + rpc.Display.Token_balance + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Display.Wallet_height)

	prediction.Predict.Info = widget.NewLabel("SCID:\n\n" + prediction.Predict.Contract + "\n")
	prediction.Predict.Info.Wrapping = fyne.TextWrapWord
	prediction.Predict.Prices = widget.NewLabel("")

	S.LeftLabel = widget.NewLabel("")
	S.RightLabel = widget.NewLabel("")
	S.RightLabel.SetText("dReams Balance: " + rpc.Display.Token_balance + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Display.Wallet_height)

	T.LeftLabel = widget.NewLabel("")
	T.RightLabel = widget.NewLabel("")
	T.LeftLabel.SetText("Total Readings: " + rpc.Display.Readings + "      Click your card for Iluma reading")
	T.RightLabel.SetText("dReams Balance: " + rpc.Display.Token_balance + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Display.Wallet_height)

	prediction.Sports.Info = widget.NewLabel("SCID:\n\n" + prediction.Sports.Contract + "\n")
	prediction.Sports.Info.Wrapping = fyne.TextWrapWord

	// dReams menu tabs
	menu_tabs := container.NewAppTabs(
		container.NewTabItem("Wallet", placeWall()),
		container.NewTabItem("dApps", layout.NewSpacer()),
		container.NewTabItem("Assets", placeAssets()),
		container.NewTabItem("Market", placeMarket()))

	menu_tabs.OnSelected = func(ti *container.TabItem) {
		MenuTab(ti)
		if ti.Text == "dApps" {
			if menu.Gnomes.Syncing {
				menu_tabs.SelectIndex(0)
			} else {
				go func() {
					reset := dReams.Window.Content().(*fyne.Container).Objects[1]
					dReams.Window.Content().(*fyne.Container).Objects[1] = dAppScreen(reset)
					dReams.Window.Content().(*fyne.Container).Objects[1].Refresh()
					menu_tabs.SelectIndex(0)
				}()
			}
		}
	}

	menu_tabs.SetTabLocation(container.TabLocationBottom)

	tarot_tabs := container.NewAppTabs(
		container.NewTabItem("Iluma", tarot.PlaceIluma()),
		container.NewTabItem("Reading", placeTarot()))

	tarot_tabs.OnSelected = func(ti *container.TabItem) {
		TarotTab(ti)
	}

	tarot_tabs.SetTabLocation(container.TabLocationBottom)

	top := canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
	top.SetMinSize(fyne.NewSize(465, 40))
	top_bar := container.NewVBox(container.NewMax(top), layout.NewSpacer())

	menu_bottom := canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
	menu_bottom.SetMinSize(fyne.NewSize(268, 40))
	menu_bottom_box := container.NewHBox(menu_bottom, layout.NewSpacer())
	menu_bottom_bar := container.NewVBox(layout.NewSpacer(), menu_bottom_box)

	tarot_bottom := canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
	tarot_bottom.SetMinSize(fyne.NewSize(135, 40))
	tarot_bottom_box := container.NewHBox(tarot_bottom, layout.NewSpacer())
	tarot_bottom_bar := container.NewVBox(layout.NewSpacer(), tarot_bottom_box)
	tarot_bottom.Hide()

	alpha_box := container.NewMax(top_bar, menu_bottom_bar, tarot_bottom_bar, menu.Alpha150)
	if dReams.os != "darwin" {
		alpha_box.Objects = append(alpha_box.Objects, FullScreenSet())
	}
	alpha_box.Objects = append(alpha_box.Objects, menu.StartIndicators())

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

	holdero_objs = placeHoldero(contract_change_screen)
	contract_objs = placeContract(holdero_objs)
	tabs := container.NewAppTabs(
		container.NewTabItem("Menu", menu_tabs),
		container.NewTabItem("Holdero", contract_objs),
		container.NewTabItem("Baccarat", placeBacc()))

	if menu.Control.Dapp_list["dSports and dPredictions"] {
		tabs.Append(container.NewTabItem("Predict", placePredict()))
		tabs.Append(container.NewTabItem("Sports", placeSports()))
	}

	if menu.Control.Dapp_list["Iluma"] {
		tabs.Append(container.NewTabItem("Tarot", TarotItems(tarot_tabs)))
	}

	if menu.Control.Dapp_list["DerBnb"] {
		tabs.Append(container.NewTabItem("DerBnb", derbnb.LayoutAllItems(true, dReams.Window, background)))
	}

	tabs.Append(container.NewTabItem("Log", rpc.SessionLog()))

	tabs.OnSelected = func(ti *container.TabItem) {
		MainTab(ti)
		if ti.Text == "Menu" {
			menu_bottom.Show()
		} else {
			menu_bottom.Hide()
		}

		if ti.Text == "Tarot" {
			tarot_bottom.Show()
		} else {
			tarot_bottom.Hide()
		}
	}

	dReams.configure = false
	max := container.NewMax(alpha_box, tabs)

	return max
}

// dReams wallet layout
func placeWall() *container.Split {
	daemon_cont := container.NewHScroll(menu.DaemonRpcEntry())
	daemon_cont.SetMinSize(fyne.NewSize(340, 35.1875))

	holdero.Swap.Dreams = widget.NewButton("Get dReams", nil)
	holdero.Swap.Dreams.Hide()

	holdero.Swap.Dero = widget.NewButton("Get Dero", nil)
	holdero.Swap.Dero.Hide()

	dReams_items := container.NewVBox(
		holdero.DreamsEntry(),
		container.NewAdaptiveGrid(2, holdero.Swap.Dreams, holdero.Swap.Dero))

	user_input_cont := container.NewVBox(
		daemon_cont,
		menu.WalletRpcEntry(),
		menu.UserPassEntry(),
		menu.RpcConnectButton(),
		layout.NewSpacer(),
		dReams_items)

	menu.Control.Contract_rating = make(map[string]uint64)
	holdero.Assets.Asset_map = make(map[string]string)

	daemon_check_cont := container.NewVBox(menu.DaemonConnectedBox())

	user_input_box := container.NewHBox(user_input_cont, daemon_check_cont)
	menu_top := container.NewHSplit(user_input_box, menu.IntroTree())

	holdero.Swap.Dreams.OnTapped = func() {
		s := strings.Trim(holdero.Swap.DEntry.Text, "dReams: ")
		amt, err := strconv.Atoi(s)
		if err == nil && holdero.Swap.DEntry.Validate() == nil {
			if amt > 0 {
				menu_top.Trailing.(*fyne.Container).Objects[1] = holdero.DreamsConfirm(1, amt, menu_top, menu.IntroTree())
				menu_top.Trailing.Refresh()
			}
		}
	}

	holdero.Swap.Dero.OnTapped = func() {
		s := strings.Trim(holdero.Swap.DEntry.Text, "dReams: ")
		amt, err := strconv.Atoi(s)
		if err == nil && holdero.Swap.DEntry.Validate() == nil {
			if amt > 0 {
				menu_top.Trailing.(*fyne.Container).Objects[1] = holdero.DreamsConfirm(2, amt, menu_top, menu.IntroTree())
				menu_top.Trailing.Refresh()
			}
		}
	}

	menu_bottom := container.NewAdaptiveGrid(1, layout.NewSpacer())
	menu_box := container.NewVSplit(menu_top, menu_bottom)
	menu_box.SetOffset(1)

	return menu_box
}

// Holdero contract tab layout
func placeContract(change_screen *fyne.Container) *container.Split {
	contract_cont := container.NewHScroll(menu.HolderoContractEntry())
	contract_cont.SetMinSize(fyne.NewSize(640, 35.1875))

	asset_items := container.NewAdaptiveGrid(1, container.NewVBox(menu.TableStats()))

	player_input := container.NewVBox(
		contract_cont,
		asset_items,
		layout.NewSpacer())

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

	max := container.NewMax(menu.Alpha120, tabs)

	menu.Poker.Holdero_unlock.OnTapped = func() {
		max.Objects[1] = menu.HolderoMenuConfirm(1, max.Objects, tabs)
		max.Objects[1].Refresh()
	}

	menu.Poker.Holdero_new.OnTapped = func() {
		max.Objects[1] = menu.HolderoMenuConfirm(2, max.Objects, tabs)
		max.Objects[1].Refresh()
	}

	player_box := container.NewHBox(player_input, check_box)
	menu_top := container.NewHSplit(player_box, max)

	mid := container.NewVBox(layout.NewSpacer(), container.NewAdaptiveGrid(2, menu.NameEntry(), holdero.TournamentButton(max.Objects, tabs)), menu.OwnersBoxMid())

	menu_bottom := container.NewGridWithColumns(3, menu.OwnersBoxLeft(max.Objects, tabs), mid, layout.NewSpacer())

	menuBox := container.NewVSplit(menu_top, menu_bottom)
	menuBox.SetOffset(1)

	return menuBox
}

// dReams asset tab layout
func placeAssets() *container.Split {
	asset_items := container.NewVBox(
		holdero.FaceSelect(),
		holdero.BackSelect(),
		holdero.ThemeSelect(),
		holdero.AvatarSelect(),
		holdero.SharedDecks(),
		RecheckButton(),
		layout.NewSpacer())

	cont := container.NewHScroll(asset_items)
	cont.SetMinSize(fyne.NewSize(290, 35.1875))

	items_box := container.NewAdaptiveGrid(2, cont, container.NewAdaptiveGrid(1, holdero.AssetStats()))

	player_input := container.NewVBox(items_box, layout.NewSpacer())

	tabs := container.NewAppTabs(
		container.NewTabItem("Owned", menu.AssetList()))

	tabs.OnSelected = func(ti *container.TabItem) {

	}

	scroll_top := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "arrowUp"), func() {
		holdero.Assets.Asset_list.ScrollToTop()
	})

	scroll_bottom := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "arrowDown"), func() {
		holdero.Assets.Asset_list.ScrollToBottom()
	})

	scroll_top.Importance = widget.LowImportance
	scroll_bottom.Importance = widget.LowImportance

	scroll_cont := container.NewVBox(container.NewHBox(layout.NewSpacer(), scroll_top, scroll_bottom))

	max := container.NewMax(menu.Alpha120, tabs, scroll_cont)

	player_input.Add(holdero.SetHeaderItems(max.Objects, tabs))
	player_box := container.NewHBox(player_input)

	menu_top := container.NewHSplit(player_box, max)
	menu_bottom := container.NewAdaptiveGrid(1, menu.IndexEntry())

	menu_box := container.NewVSplit(menu_top, menu_bottom)
	menu_box.SetOffset(1)

	return menu_box
}

// dReams market tabs layout
func placeMarket() *container.Split {
	details := container.NewMax(menu.NfaMarketInfo())

	tabs := container.NewAppTabs(
		container.NewTabItem("Auctions", menu.AuctionListings()),
		container.NewTabItem("Buy Now", menu.BuyNowListings()))

	tabs.SetTabLocation(container.TabLocationTop)
	tabs.OnSelected = func(ti *container.TabItem) {
		MarketTab(ti)
	}

	menu.Market.Tab = "Auction"

	scroll_top := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "arrowUp"), func() {
		switch menu.Market.Tab {
		case "Buy":
			menu.Market.Buy_list.ScrollToTop()
		case "Auction":
			menu.Market.Auction_list.ScrollToTop()
		default:

		}
	})

	scroll_bottom := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "arrowDown"), func() {
		switch menu.Market.Tab {
		case "Buy":
			menu.Market.Buy_list.ScrollToBottom()
		case "Auction":
			menu.Market.Auction_list.ScrollToBottom()
		default:

		}
	})

	scroll_top.Importance = widget.LowImportance
	scroll_bottom.Importance = widget.LowImportance

	scroll_cont := container.NewVBox(container.NewHBox(layout.NewSpacer(), scroll_top, scroll_bottom))

	max := container.NewMax(menu.Alpha120, tabs, scroll_cont)

	details_box := container.NewVBox(layout.NewSpacer(), details)

	menu_top := container.NewHSplit(details_box, max)
	menu_top.SetOffset(0)

	menu.Market.Market_button = widget.NewButton("Bid", func() {
		scid := menu.Market.Viewing
		if len(scid) == 64 {
			text := menu.Market.Market_button.Text
			menu.Market.Market_button.Hide()
			if text == "Bid" {
				amt := menu.ToAtomicFive(menu.Market.Entry.Text)
				menu_top.Trailing.(*fyne.Container).Objects[1] = menu.BidBuyConfirm(scid, amt, 0, menu_top, container.NewMax(menu.Alpha120, tabs, scroll_cont))
				menu_top.Trailing.(*fyne.Container).Objects[1].Refresh()
			} else if text == "Buy" {
				menu_top.Trailing.(*fyne.Container).Objects[1] = menu.BidBuyConfirm(scid, menu.Market.Buy_amt, 1, menu_top, container.NewMax(menu.Alpha120, tabs, scroll_cont))
				menu_top.Trailing.(*fyne.Container).Objects[1].Refresh()
			}
		}
	})

	menu.Market.Market_button.Hide()

	menu.Market.Cancel_button = widget.NewButton("Cancel", func() {
		if len(menu.Market.Viewing) == 64 {
			menu.Market.Cancel_button.Hide()
			menu_top.Trailing.(*fyne.Container).Objects[1] = menu.ConfirmCancelClose(menu.Market.Viewing, 1, menu_top, container.NewMax(menu.Alpha120, tabs, scroll_cont))
			menu_top.Trailing.(*fyne.Container).Objects[1].Refresh()
		}
	})

	menu.Market.Close_button = widget.NewButton("Close", func() {
		if len(menu.Market.Viewing) == 64 {
			menu.Market.Close_button.Hide()
			menu_top.Trailing.(*fyne.Container).Objects[1] = menu.ConfirmCancelClose(menu.Market.Viewing, 0, menu_top, container.NewMax(menu.Alpha120, tabs, scroll_cont))
			menu_top.Trailing.(*fyne.Container).Objects[1].Refresh()
		}
	})

	menu.Market.Market_box = *container.NewAdaptiveGrid(6, menu.MarketEntry(), menu.Market.Market_button, layout.NewSpacer(), layout.NewSpacer(), menu.Market.Close_button, menu.Market.Cancel_button)
	menu.Market.Market_box.Hide()

	menu_bottom := container.NewAdaptiveGrid(1, &menu.Market.Market_box)

	menu_box := container.NewVSplit(menu_top, menu_bottom)
	menu_box.SetOffset(1)

	return menu_box
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
		baccarat.BaccaratButtons())

	return B.DApp
}

// dReams dPrediction tab layout
func placePredict() *fyne.Container {
	contract_cont := container.NewHScroll(prediction.PreictionContractEntry())
	contract_cont.SetMinSize(fyne.NewSize(600, 35.1875))
	predict_info := container.NewVBox(prediction.Predict.Info, prediction.Predict.Prices)
	predict_scroll := container.NewScroll(predict_info)
	predict_scroll.SetMinSize(fyne.NewSize(540, 500))

	check_box := container.NewVBox(prediction.PredictConnectedBox())

	hbox := container.NewHBox(contract_cont, check_box)

	prediction.Predict.Higher = widget.NewButton("Higher", nil)
	prediction.Predict.Higher.Hide()

	prediction.Predict.Lower = widget.NewButton("Lower", nil)
	prediction.Predict.Lower.Hide()

	prediction.Predict.Prediction_box = container.NewVBox(prediction.Predict.Higher, prediction.Predict.Lower)
	prediction.Predict.Prediction_box.Hide()

	predict_content := container.NewVBox(
		hbox,
		predict_scroll,
		layout.NewSpacer(),
		prediction.Predict.Prediction_box)

	// leaders_scroll := container.NewScroll(prediction.LeadersDisplay())
	// leaders_scroll.SetMinSize(fyne.NewSize(180, 500))
	// leaders_contnet := container.NewVBox(leaders_scroll)

	menu.Control.Bet_unlock_p = widget.NewButton("Unlock dPrediction Contract", nil)
	menu.Control.Bet_unlock_p.Hide()

	menu.Control.Bet_new_p = widget.NewButton("New dPrediction Contract", nil)
	menu.Control.Bet_new_p.Hide()

	unlock_cont := container.NewVBox(menu.Control.Bet_unlock_p, menu.Control.Bet_new_p)

	owner_buttons := container.NewAdaptiveGrid(2, container.NewMax(prediction.OwnerButtonP()), unlock_cont)
	owned_tab := container.NewBorder(nil, owner_buttons, nil, nil, prediction.PredictionOwned())

	tabs := container.NewAppTabs(
		container.NewTabItem("Contracts", layout.NewSpacer()),
		container.NewTabItem("Favorites", prediction.PredicitionFavorites()),
		container.NewTabItem("Owned", owned_tab))
	// container.NewTabItem("Leaderboard", leaders_contnet))

	tabs.SelectIndex(0)
	tabs.Selected().Content = prediction.PredictionListings(tabs)

	tabs.OnSelected = func(ti *container.TabItem) {
		PredictTab(ti)
	}

	max := container.NewMax(menu.Alpha120, tabs)

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
		time.Sleep(time.Second)
		for !menu.Exit_signal && menu.Control.Dapp_list["dSports and dPredictions"] {
			if !rpc.Wallet.Connect {
				if menu.Control.Dapp_list["dSports and dPredictions"] {
					menu.Control.Predict_check.SetChecked(false)
					menu.Control.Sports_check.SetChecked(false)
					prediction.DisablePreditions(true)
					prediction.DisableSports(true)
				}
			}
			time.Sleep(time.Second)
		}
	}()

	return P.DApp
}

// dReams dSports tab layout
func placeSports() *fyne.Container {
	cont := container.NewHScroll(prediction.SportsContractEntry())
	cont.SetMinSize(fyne.NewSize(600, 35.1875))
	sports_content := container.NewVBox(prediction.Sports.Info)
	sports_scroll := container.NewVScroll(sports_content)
	sports_scroll.SetMinSize(fyne.NewSize(180, 500))

	check_box := container.NewVBox(prediction.SportsConnectedBox())
	hbox := container.NewHBox(cont, check_box)

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

	sports_muli := container.NewCenter(prediction.Sports.Multi)
	prediction.Sports.Sports_box = container.NewVBox(
		sports_muli,
		prediction.Sports.Game_select,
		prediction.Sports.ButtonA,
		prediction.Sports.ButtonB)

	prediction.Sports.Sports_box.Hide()

	sports_left := container.NewVBox(
		hbox,
		sports_scroll,
		layout.NewSpacer(),
		prediction.Sports.Sports_box)

	epl := widget.NewLabel("")
	epl.Wrapping = fyne.TextWrapWord
	epl_scroll := container.NewVScroll(epl)
	nba := widget.NewLabel("")
	nba.Wrapping = fyne.TextWrapWord
	nba_scroll := container.NewVScroll(nba)
	nfl := widget.NewLabel("")
	nfl.Wrapping = fyne.TextWrapWord
	nfl_scroll := container.NewVScroll(nfl)
	nhl := widget.NewLabel("")
	nhl.Wrapping = fyne.TextWrapWord
	nhl_scroll := container.NewVScroll(nhl)
	bellator := widget.NewLabel("")
	bellator.Wrapping = fyne.TextWrapWord
	bellator_scroll := container.NewVScroll(bellator)
	ufc := widget.NewLabel("")
	ufc.Wrapping = fyne.TextWrapWord
	ufc_scroll := container.NewVScroll(ufc)
	score_tabs := container.NewAppTabs(
		container.NewTabItem("EPL", epl_scroll),
		container.NewTabItem("NBA", nba_scroll),
		container.NewTabItem("NFL", nfl_scroll),
		container.NewTabItem("NHL", nhl_scroll),
		container.NewTabItem("Bellator", bellator_scroll),
		container.NewTabItem("UFC", ufc_scroll))

	score_tabs.OnSelected = func(ti *container.TabItem) {
		switch ti.Text {
		case "EPL":
			go prediction.GetScores(epl, "EPL")
		case "NBA":
			go prediction.GetScores(nba, "NBA")
		case "NFL":
			go prediction.GetScores(nfl, "NFL")
		case "NHL":
			go prediction.GetScores(nhl, "NHL")
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

	max := container.NewMax(menu.Alpha120, tabs)

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
func placeTarot() *fyne.Container {
	tarot_label := container.NewHBox(T.LeftLabel, layout.NewSpacer(), T.RightLabel)

	T.DApp = container.NewBorder(
		labelColorBlack(tarot_label),
		nil,
		nil,
		nil,
		tarot.TarotCardBox())

	return T.DApp
}
