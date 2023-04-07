package main

import (
	_ "embed"
	"image/color"
	"strconv"
	"strings"

	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/prediction"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/table"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var H table.Items
var B table.Items
var P table.Items
var S table.Items
var A table.Items
var T table.Items

//go:embed table/iluma/iluma.txt
var iluma_intro string

// If dReams has not been intialized, show this screen
func introScreen() *fyne.Container {
	dReams.configure = true
	title := canvas.NewText("Welcome to dReams", color.White)
	title.TextSize = 18

	intro_label := widget.NewLabel("dReams base app has Holdero, Baccarat and NFA Marketplace\n\nSelect further dApps to add to your dReams")
	intro_label.Wrapping = fyne.TextWrapWord
	intro_label.Alignment = fyne.TextAlignCenter

	dapp_list := []string{"dSports and dPredictions", "Iluma"}
	dapp_checks := widget.NewCheckGroup(dapp_list, func(s []string) {})

	start_button := widget.NewButton("Start dReams", func() {
		menu.MenuControl.Dapp_list = make(map[string]bool)
		for _, name := range dapp_checks.Selected {
			menu.MenuControl.Dapp_list[name] = true
		}

		go func() {
			menu.MenuControl.Dapp_list["dReams"] = true
			dReams.Window.SetContent(
				container.New(layout.NewMaxLayout(),
					background,
					place()))
		}()
	})

	start_button.Importance = widget.LowImportance

	alpha := canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
	intro := container.NewVBox(layout.NewSpacer(), container.NewCenter(title), layout.NewSpacer(), intro_label, container.NewCenter(dapp_checks), layout.NewSpacer(), start_button)
	max := container.NewMax(alpha, intro)

	return max
}

// Select dApps to add or remove from dReams
func dAppScreen() *fyne.Container {
	dReams.configure = true
	title := canvas.NewText("dReams dApps", color.White)
	title.TextSize = 18

	intro_label := widget.NewLabel("Select dApps to add or remove from your dReams")
	intro_label.Wrapping = fyne.TextWrapWord
	intro_label.Alignment = fyne.TextAlignCenter

	enabled_dapps := []string{}
	dapp_list := []string{"dSports and dPredictions", "Iluma"}
	dapp_checks := widget.NewCheckGroup(dapp_list, func(s []string) {})

	for name, enabled := range menu.MenuControl.Dapp_list {
		if enabled {
			enabled_dapps = append(enabled_dapps, name)
			menu.MenuControl.Dapp_list[name] = false
		}
	}

	dapp_checks.SetSelected(enabled_dapps)

	load_button := widget.NewButton("Load Changes", func() {
		for _, name := range dapp_checks.Selected {
			menu.MenuControl.Dapp_list[name] = true
		}

		go func() {
			dReams.Window.Content().(*fyne.Container).Objects[1] = place()
			dReams.Window.Content().(*fyne.Container).Objects[1].Refresh()
		}()
	})

	load_button.Importance = widget.LowImportance

	alpha := canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
	intro := container.NewVBox(layout.NewSpacer(), container.NewCenter(title), layout.NewSpacer(), intro_label, container.NewCenter(dapp_checks), layout.NewSpacer(), load_button)
	max := container.NewMax(alpha, intro)

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

	prediction.PredictControl.Info = widget.NewLabel("SCID: \n" + prediction.PredictControl.Contract + "\n")
	prediction.PredictControl.Info.Wrapping = fyne.TextWrapWord
	prediction.PredictControl.Prices = widget.NewLabel("")

	S.LeftLabel = widget.NewLabel("")
	S.RightLabel = widget.NewLabel("")
	S.RightLabel.SetText("dReams Balance: " + rpc.Display.Token_balance + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Display.Wallet_height)

	T.LeftLabel = widget.NewLabel("")
	T.RightLabel = widget.NewLabel("")
	T.LeftLabel.SetText("Total Readings: " + rpc.Display.Readings + "      Click your card for Iluma reading")
	T.RightLabel.SetText("dReams Balance: " + rpc.Display.Token_balance + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Display.Wallet_height)

	prediction.SportsControl.Info = widget.NewLabel("SCID: \n" + prediction.SportsControl.Contract + "\n")
	prediction.SportsControl.Info.Wrapping = fyne.TextWrapWord

	// dReams menu tabs
	menu_tabs := container.NewAppTabs(
		container.NewTabItem("Wallet", placeWall()),
		container.NewTabItem("dApps", layout.NewSpacer()),
		container.NewTabItem("Assets", placeAssets()),
		container.NewTabItem("Market", placeMarket()))

	menu_tabs.OnSelected = func(ti *container.TabItem) {
		MenuTab(ti)
	}

	menu_tabs.SetTabLocation(container.TabLocationBottom)

	tarot_tabs := container.NewAppTabs(
		container.NewTabItem("Iluma", placeIluma()),
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

	alpha := canvas.NewRectangle(color.RGBA{0, 0, 0, 150})
	alpha_box := container.NewMax(top_bar, menu_bottom_bar, tarot_bottom_bar, alpha)
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

	if menu.MenuControl.Dapp_list["dSports and dPredictions"] {
		tabs.Append(container.NewTabItem("Predict", placePredict()))
		tabs.Append(container.NewTabItem("Sports", placeSports()))
	}

	if menu.MenuControl.Dapp_list["Iluma"] {
		tabs.Append(container.NewTabItem("Tarot", TarotItems(tarot_tabs)))
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

	table.Actions.Dreams = widget.NewButton("Get dReams", nil)
	table.Actions.Dreams.Hide()

	table.Actions.Dero = widget.NewButton("Get Dero", nil)
	table.Actions.Dero.Hide()

	dReams_items := container.NewVBox(
		table.DreamsEntry(),
		container.NewAdaptiveGrid(2, table.Actions.Dreams, table.Actions.Dero))

	user_input_cont := container.NewVBox(
		daemon_cont,
		menu.WalletRpcEntry(),
		menu.UserPassEntry(),
		menu.RpcConnectButton(),
		layout.NewSpacer(),
		dReams_items)

	daemon_check_cont := container.NewVBox(
		menu.DaemonConnectedBox())

	user_input_box := container.NewHBox(user_input_cont, daemon_check_cont)
	menu_top := container.NewHSplit(user_input_box, menu.IntroTree())

	table.Actions.Dreams.OnTapped = func() {
		s := strings.Trim(table.Actions.DEntry.Text, "dReams: ")
		amt, err := strconv.Atoi(s)
		if err == nil && table.Actions.DEntry.Validate() == nil {
			if amt > 0 {
				menu_top.Trailing.(*fyne.Container).Objects[1] = table.DreamsConfirm(1, amt, menu_top, menu.IntroTree())
				menu_top.Trailing.Refresh()
			}
		}
	}

	table.Actions.Dero.OnTapped = func() {
		s := strings.Trim(table.Actions.DEntry.Text, "dReams: ")
		amt, err := strconv.Atoi(s)
		if err == nil && table.Actions.DEntry.Validate() == nil {
			if amt > 0 {
				menu_top.Trailing.(*fyne.Container).Objects[1] = table.DreamsConfirm(2, amt, menu_top, menu.IntroTree())
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
	menu.HolderoControl.Holdero_unlock = widget.NewButton("Unlock Holdero Contract", nil)
	menu.HolderoControl.Holdero_unlock.Hide()

	menu.HolderoControl.Holdero_new = widget.NewButton("New Holdero Table", nil)
	menu.HolderoControl.Holdero_new.Hide()

	unlock_cont := container.NewVBox(
		layout.NewSpacer(),
		menu.HolderoControl.Holdero_unlock,
		menu.HolderoControl.Holdero_new)

	owner_buttons := container.NewAdaptiveGrid(2, container.NewMax(layout.NewSpacer()), unlock_cont)
	owned_tab := container.NewBorder(nil, owner_buttons, nil, nil, menu.MyTables())

	tabs = container.NewAppTabs(
		container.NewTabItem("Tables", layout.NewSpacer()),
		container.NewTabItem("Favorites", menu.HolderoFavorites()),
		container.NewTabItem("Owned", owned_tab),
		container.NewTabItem("View", layout.NewSpacer()))

	tabs.SelectTabIndex(0)
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
					tabs.SelectTabIndex(0)
				} else {
					tabs.SelectTabIndex(0)
				}
			}()
		}
	}

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	max := container.NewMax(alpha, tabs)

	menu.HolderoControl.Holdero_unlock.OnTapped = func() {
		max.Objects[1] = menu.HolderoMenuConfirm(1, max.Objects, tabs)
		max.Objects[1].Refresh()
	}

	menu.HolderoControl.Holdero_new.OnTapped = func() {
		max.Objects[1] = menu.HolderoMenuConfirm(2, max.Objects, tabs)
		max.Objects[1].Refresh()
	}

	player_box := container.NewHBox(player_input, check_box)
	menu_top := container.NewHSplit(player_box, max)

	mid := container.NewVBox(layout.NewSpacer(), container.NewAdaptiveGrid(2, menu.NameEntry(), table.TournamentButton(max.Objects, tabs)), menu.OwnersBoxMid())

	menu_bottom := container.NewGridWithColumns(3, menu.OwnersBoxLeft(max.Objects, tabs), mid, layout.NewSpacer())

	menuBox := container.NewVSplit(menu_top, menu_bottom)
	menuBox.SetOffset(1)

	return menuBox
}

// dReams asset tab layout
func placeAssets() *container.Split {
	asset_items := container.NewVBox(
		table.FaceSelect(),
		table.BackSelect(),
		table.ThemeSelect(),
		table.AvatarSelect(),
		table.SharedDecks(),
		RecheckButton(),
		layout.NewSpacer())

	cont := container.NewHScroll(asset_items)
	cont.SetMinSize(fyne.NewSize(290, 35.1875))

	items_box := container.NewAdaptiveGrid(2, cont, container.NewAdaptiveGrid(1, table.AssetStats()))

	player_input := container.NewVBox(items_box, layout.NewSpacer())

	tabs := container.NewAppTabs(
		container.NewTabItem("Owned", menu.AssetList()))

	tabs.OnSelected = func(ti *container.TabItem) {

	}

	scroll_top := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "arrowUp"), func() {
		table.Assets.Asset_list.ScrollToTop()
	})

	scroll_bottom := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "arrowDown"), func() {
		table.Assets.Asset_list.ScrollToBottom()
	})

	scroll_top.Importance = widget.LowImportance
	scroll_bottom.Importance = widget.LowImportance

	scroll_cont := container.NewVBox(container.NewHBox(layout.NewSpacer(), scroll_top, scroll_bottom))

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	max := container.NewMax(alpha, tabs, scroll_cont)

	player_input.AddObject(table.SetHeaderItems(max.Objects, tabs))
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

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	max := container.NewMax(alpha, tabs, scroll_cont)

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
				menu_top.Trailing.(*fyne.Container).Objects[1] = menu.BidBuyConfirm(scid, amt, 0, menu_top, container.NewMax(alpha, tabs, scroll_cont))
				menu_top.Trailing.(*fyne.Container).Objects[1].Refresh()
			} else if text == "Buy" {
				menu_top.Trailing.(*fyne.Container).Objects[1] = menu.BidBuyConfirm(scid, menu.Market.Buy_amt, 1, menu_top, container.NewMax(alpha, tabs, scroll_cont))
				menu_top.Trailing.(*fyne.Container).Objects[1].Refresh()
			}
		}
	})

	menu.Market.Market_button.Hide()

	menu.Market.Cancel_button = widget.NewButton("Cancel", func() {
		if len(menu.Market.Viewing) == 64 {
			menu.Market.Cancel_button.Hide()
			menu_top.Trailing.(*fyne.Container).Objects[1] = menu.ConfirmCancelClose(menu.Market.Viewing, 1, menu_top, container.NewMax(alpha, tabs, scroll_cont))
			menu_top.Trailing.(*fyne.Container).Objects[1].Refresh()
		}
	})

	menu.Market.Close_button = widget.NewButton("Close", func() {
		if len(menu.Market.Viewing) == 64 {
			menu.Market.Close_button.Hide()
			menu_top.Trailing.(*fyne.Container).Objects[1] = menu.ConfirmCancelClose(menu.Market.Viewing, 0, menu_top, container.NewMax(alpha, tabs, scroll_cont))
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
	H.TableContent = *container.NewWithoutLayout(
		table.HolderoTable(resourceTablePng),
		table.Player1_label(nil, nil, nil),
		table.Player2_label(nil, nil, nil),
		table.Player3_label(nil, nil, nil),
		table.Player4_label(nil, nil, nil),
		table.Player5_label(nil, nil, nil),
		table.Player6_label(nil, nil, nil),
		H.TopLabel,
	)

	holdero_label := container.NewHBox(H.LeftLabel, layout.NewSpacer(), H.RightLabel)

	H.CardsContent = *placeHolderoCards()

	H.ActionButtons = *container.NewVBox(
		layout.NewSpacer(),
		table.SitButton(),
		table.LeaveButton(),
		table.DealHandButton(),
		table.CheckButton(),
		table.BetButton(),
		table.BetAmount())

	options := container.NewVBox(layout.NewSpacer(), table.AutoOptions(), change_screen)

	holdero_actions := container.NewHBox(options, layout.NewSpacer(), table.TimeOutWarning(), layout.NewSpacer(), layout.NewSpacer(), &H.ActionButtons)

	H.TableItems = container.NewVBox(
		labelColorBlack(holdero_label),
		&H.TableContent,
		&H.CardsContent,
		layout.NewSpacer(),
		holdero_actions)

	return H.TableItems
}

// dReams Baccarat tab layout
func placeBacc() *fyne.Container {
	B.TableContent = *container.NewWithoutLayout(
		table.BaccTable(resourceBaccTablePng),
		table.BaccResult(rpc.Display.BaccRes),
	)

	bacc_label := container.NewHBox(B.LeftLabel, layout.NewSpacer(), B.RightLabel)

	B.TableItems = container.NewVBox(
		labelColorBlack(bacc_label),
		&B.TableContent,
		&B.CardsContent,
		layout.NewSpacer(),
		table.BaccaratButtons(),
	)

	return B.TableItems
}

// dReams dPrediction tab layout
func placePredict() *fyne.Container {
	contract_cont := container.NewHScroll(prediction.PreictionContractEntry())
	contract_cont.SetMinSize(fyne.NewSize(600, 35.1875))
	predict_info := container.NewVBox(prediction.PredictControl.Info, prediction.PredictControl.Prices)
	predict_scroll := container.NewScroll(predict_info)
	predict_scroll.SetMinSize(fyne.NewSize(540, 500))

	check_box := container.NewVBox(prediction.PredictConnectedBox())

	hbox := container.NewHBox(contract_cont, check_box)

	table.Actions.Higher = widget.NewButton("Higher", nil)
	table.Actions.Higher.Hide()

	table.Actions.Lower = widget.NewButton("Lower", nil)
	table.Actions.Lower.Hide()

	table.Actions.Prediction_box = container.NewVBox(table.Actions.Higher, table.Actions.Lower)
	table.Actions.Prediction_box.Hide()

	predict_content := container.NewVBox(
		hbox,
		predict_scroll,
		layout.NewSpacer(),
		table.Actions.Prediction_box)

	// leaders_scroll := container.NewScroll(prediction.LeadersDisplay())
	// leaders_scroll.SetMinSize(fyne.NewSize(180, 500))
	// leaders_contnet := container.NewVBox(leaders_scroll)

	menu.MenuControl.Bet_unlock_p = widget.NewButton("Unlock dPrediction Contract", nil)
	menu.MenuControl.Bet_unlock_p.Hide()

	menu.MenuControl.Bet_new_p = widget.NewButton("New dPrediction Contract", nil)
	menu.MenuControl.Bet_new_p.Hide()

	unlock_cont := container.NewVBox(menu.MenuControl.Bet_unlock_p, menu.MenuControl.Bet_new_p)

	owner_buttons := container.NewAdaptiveGrid(2, container.NewMax(prediction.OwnerButtonP()), unlock_cont)
	owned_tab := container.NewBorder(nil, owner_buttons, nil, nil, prediction.PredictionOwned())

	tabs := container.NewAppTabs(
		container.NewTabItem("Contracts", layout.NewSpacer()),
		container.NewTabItem("Favorites", prediction.PredicitionFavorites()),
		container.NewTabItem("Owned", owned_tab))
	// container.NewTabItem("Leaderboard", leaders_contnet))

	tabs.SelectTabIndex(0)
	tabs.Selected().Content = prediction.PredictionListings(tabs)

	tabs.OnSelected = func(ti *container.TabItem) {
		PredictTab(ti)
	}

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	max := container.NewMax(alpha, tabs)

	table.Actions.Higher.OnTapped = func() {
		if len(prediction.PredictControl.Contract) == 64 {
			max.Objects[1] = prediction.ConfirmAction(2, "", "", max.Objects, tabs)
			max.Objects[1].Refresh()
		}
	}

	table.Actions.Lower.OnTapped = func() {
		if len(prediction.PredictControl.Contract) == 64 {
			max.Objects[1] = prediction.ConfirmAction(1, "", "", max.Objects, tabs)
			max.Objects[1].Refresh()
		}
	}

	menu.MenuControl.Bet_unlock_p.OnTapped = func() {
		max.Objects[1] = menu.BettingMenuConfirmP(1, max.Objects, tabs)
		max.Objects[1].Refresh()
	}

	menu.MenuControl.Bet_new_p.OnTapped = func() {
		max.Objects[1] = menu.BettingMenuConfirmP(2, max.Objects, tabs)
		max.Objects[1].Refresh()
	}

	predict_label := container.NewHBox(P.LeftLabel, layout.NewSpacer(), P.RightLabel)
	predict_box := container.NewHSplit(predict_content, max)

	P.TableItems = container.NewVBox(
		labelColorBlack(predict_label),
		predict_box)

	return P.TableItems
}

// dReams dSports tab layout
func placeSports() *fyne.Container {
	cont := container.NewHScroll(prediction.SportsContractEntry())
	cont.SetMinSize(fyne.NewSize(600, 35.1875))
	sports_content := container.NewVBox(prediction.SportsControl.Info)
	sports_scroll := container.NewVScroll(sports_content)
	sports_scroll.SetMinSize(fyne.NewSize(180, 500))

	check_box := container.NewVBox(prediction.SportsConnectedBox())
	hbox := container.NewHBox(cont, check_box)

	table.Actions.Game_select = widget.NewSelect(table.Actions.Game_options, func(s string) {
		split := strings.Split(s, "   ")
		a, b := menu.GetSportsTeams(prediction.SportsControl.Contract, split[0])
		if table.Actions.Game_select.SelectedIndex() >= 0 {
			table.Actions.Multi.Show()
			table.Actions.ButtonA.Show()
			table.Actions.ButtonB.Show()
			table.Actions.ButtonA.Text = a
			table.Actions.ButtonA.Refresh()
			table.Actions.ButtonB.Text = b
			table.Actions.ButtonB.Refresh()
		} else {
			table.Actions.Multi.Hide()
			table.Actions.ButtonA.Hide()
			table.Actions.ButtonB.Hide()
		}
	})

	table.Actions.Game_select.PlaceHolder = "Select Game #"
	table.Actions.Game_select.Hide()

	var Multi_options = []string{"1x", "3x", "5x"}
	table.Actions.Multi = widget.NewRadioGroup(Multi_options, func(s string) {})
	table.Actions.Multi.Horizontal = true
	table.Actions.Multi.Hide()

	table.Actions.ButtonA = widget.NewButton("TEAM A", nil)
	table.Actions.ButtonA.Hide()

	table.Actions.ButtonB = widget.NewButton("TEAM B", nil)
	table.Actions.ButtonB.Hide()

	sports_muli := container.NewCenter(table.Actions.Multi)
	table.Actions.Sports_box = container.NewVBox(
		sports_muli,
		table.Actions.Game_select,
		table.Actions.ButtonA,
		table.Actions.ButtonB)

	table.Actions.Sports_box.Hide()

	sports_left := container.NewVBox(
		hbox,
		sports_scroll,
		layout.NewSpacer(),
		table.Actions.Sports_box)

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

	menu.MenuControl.Bet_unlock_s = widget.NewButton("Unlock dSports Contracts", nil)
	menu.MenuControl.Bet_unlock_s.Hide()

	menu.MenuControl.Bet_new_s = widget.NewButton("New dSports Contract", nil)
	menu.MenuControl.Bet_new_s.Hide()

	unlock_cont := container.NewVBox(
		menu.MenuControl.Bet_unlock_s,
		menu.MenuControl.Bet_new_s)

	owner_buttons := container.NewAdaptiveGrid(2, container.NewMax(prediction.OwnerButtonS()), unlock_cont)
	owned_tab := container.NewBorder(nil, owner_buttons, nil, nil, prediction.SportsOwned())

	tabs := container.NewAppTabs(
		container.NewTabItem("Contracts", layout.NewSpacer()),
		container.NewTabItem("Favorites", prediction.SportsFavorites()),
		container.NewTabItem("Owned", owned_tab),
		container.NewTabItem("Scores", score_tabs),
		container.NewTabItem("Payouts", prediction.SportsPayouts()))

	tabs.SelectTabIndex(0)
	tabs.Selected().Content = prediction.SportsListings(tabs)

	tabs.OnSelected = func(ti *container.TabItem) {

	}

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	max := container.NewMax(alpha, tabs)

	table.Actions.ButtonA.OnTapped = func() {
		if len(prediction.SportsControl.Contract) == 64 {
			max.Objects[1] = prediction.ConfirmAction(3, table.Actions.ButtonA.Text, table.Actions.ButtonB.Text, max.Objects, tabs)
			max.Objects[1].Refresh()
		}
	}
	table.Actions.ButtonA.Hide()

	table.Actions.ButtonB.OnTapped = func() {
		if len(prediction.SportsControl.Contract) == 64 {
			max.Objects[1] = prediction.ConfirmAction(4, table.Actions.ButtonA.Text, table.Actions.ButtonB.Text, max.Objects, tabs)
			max.Objects[1].Refresh()
		}
	}

	menu.MenuControl.Bet_unlock_s.OnTapped = func() {
		max.Objects[1] = menu.BettingMenuConfirmS(1, max.Objects, tabs)
		max.Objects[1].Refresh()
	}

	menu.MenuControl.Bet_new_s.OnTapped = func() {
		max.Objects[1] = menu.BettingMenuConfirmS(2, max.Objects, tabs)
		max.Objects[1].Refresh()
	}

	sports_label := container.NewHBox(S.LeftLabel, layout.NewSpacer(), S.RightLabel)
	sports_box := container.NewHSplit(sports_left, max)

	S.TableItems = container.NewVBox(
		labelColorBlack(sports_label),
		sports_box)

	return S.TableItems
}

// Iluma tab objects
func placeIluma() *fyne.Container {
	var first, second, third bool
	var display int
	img := canvas.NewImageFromResource(resourceIluma82Png)
	intro := widget.NewLabel(iluma_intro)
	scroll := container.NewScroll(intro)

	cont := container.NewGridWithColumns(2, scroll, img)
	alpha := canvas.NewRectangle(color.RGBA{0, 0, 0, 150})
	max := container.NewMax(alpha, cont)

	scroll.OnScrolled = func(p fyne.Position) {
		if p.Y <= 400 {
			second = false
			third = false
			display = 1
		} else if p.Y >= 400 && p.Y <= 800 {
			first = false
			third = false
			display = 2
		} else if p.Y >= 800 {
			first = false
			second = false
			display = 3
		}

		switch display {
		case 1:
			if !first {
				cont.Objects[1] = canvas.NewImageFromResource(resourceIluma82Png)
				cont.Refresh()
				first = true
			}
		case 2:
			if !second {
				cont.Objects[1] = canvas.NewImageFromResource(resourceIluma80Png)
				cont.Refresh()
				second = true
			}
		case 3:
			if !third {
				cont.Objects[1] = canvas.NewImageFromResource(resourceIluma83Png)
				cont.Refresh()
				third = true
			}
		default:

		}
	}

	return max
}

// dReams Tarot tab layout
func placeTarot() *fyne.Container {
	tarot_label := container.NewHBox(T.LeftLabel, layout.NewSpacer(), T.RightLabel)

	T.TableItems = container.NewBorder(
		labelColorBlack(tarot_label),
		nil,
		nil,
		nil,
		table.TarotCardBox(),
	)

	return T.TableItems
}
