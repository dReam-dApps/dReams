package main

import (
	_ "embed"
	"image/color"

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

func setLabels() {
	H.LeftLabel.SetText("Seats: " + rpc.Display.Seats + "      Pot: " + rpc.Display.Pot + "      Blinds: " + rpc.Display.Blinds + "      Ante: " + rpc.Display.Ante + "      Dealer: " + rpc.Display.Dealer + "      Turn: " + rpc.Display.Turn)
	H.RightLabel.SetText(rpc.Display.Readout + "      Player ID: " + rpc.Display.PlayerId + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Wallet.Height)

	B.LeftLabel.SetText("Total Hands Played: " + rpc.Display.Total_w + "      Player Wins: " + rpc.Display.Player_w + "      Ties: " + rpc.Display.Ties + "      Banker Wins: " + rpc.Display.Banker_w + "      Min Bet is " + rpc.Display.BaccMin + " dReams, Max Bet is " + rpc.Display.BaccMax)
	B.RightLabel.SetText("dReams Balance: " + rpc.Wallet.TokenBal + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Wallet.Height)

	P.RightLabel.SetText("dReams Balance: " + rpc.Wallet.TokenBal + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Wallet.Height)

	S.RightLabel.SetText("dReams Balance: " + rpc.Wallet.TokenBal + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Wallet.Height)

	T.LeftLabel.SetText("Total Readings: " + rpc.Display.Readings + "      Click your card for Iluma reading")
	T.RightLabel.SetText("dReams Balance: " + rpc.Wallet.TokenBal + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Wallet.Height)
}

func place() *fyne.Container {
	H.LeftLabel = widget.NewLabel("")
	H.RightLabel = widget.NewLabel("")
	H.TopLabel = widget.NewLabel("")
	H.TopLabel.Move(fyne.NewPos(380, 194))

	B.LeftLabel = widget.NewLabel("")
	B.RightLabel = widget.NewLabel("")

	P.LeftLabel = widget.NewLabel("")
	P.RightLabel = widget.NewLabel("")

	prediction.PredictControl.Info = widget.NewLabel("SCID: \n" + prediction.PredictControl.Contract + "\n")
	prediction.PredictControl.Info.Wrapping = fyne.TextWrapWord
	prediction.PredictControl.Prices = widget.NewLabel("")

	S.LeftLabel = widget.NewLabel("")
	S.RightLabel = widget.NewLabel("")

	T.LeftLabel = widget.NewLabel("")
	T.RightLabel = widget.NewLabel("")

	prediction.SportsControl.Info = widget.NewLabel("SCID: \n" + prediction.SportsControl.Contract + "\n")
	prediction.SportsControl.Info.Wrapping = fyne.TextWrapWord

	menu_tabs := container.NewAppTabs(
		container.NewTabItem("Wallet", placeWall()),
		container.NewTabItem("Contracts", placeContract()),
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
	top_box := container.NewHBox(top, layout.NewSpacer())
	top_bar := container.NewVBox(top_box, layout.NewSpacer())

	menu_bottom := canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
	menu_bottom.SetMinSize(fyne.NewSize(291, 40))
	menu_bottom_box := container.NewHBox(menu_bottom, layout.NewSpacer())
	menu_bottom_bar := container.NewVBox(layout.NewSpacer(), menu_bottom_box)

	tarot_bottom := canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
	tarot_bottom.SetMinSize(fyne.NewSize(135, 40))
	tarot_bottom_box := container.NewHBox(tarot_bottom, layout.NewSpacer())
	tarot_bottom_bar := container.NewVBox(layout.NewSpacer(), tarot_bottom_box)

	alpha := canvas.NewRectangle(color.RGBA{0, 0, 0, 150})
	alpha_box := container.NewMax(top_bar, menu_bottom_bar, tarot_bottom_bar, alpha, menu.StartIndicators())

	tabs := container.NewAppTabs(
		container.NewTabItem("Menu", menu_tabs),
		container.NewTabItem("Holdero", placeHoldero()),
		container.NewTabItem("Baccarat", placeBacc()),
		container.NewTabItem("Predict", placePredict()),
		container.NewTabItem("Sports", placeSports()),
		container.NewTabItem("Tarot", TarotItems(tarot_tabs)),
		container.NewTabItem("Log", rpc.SessionLog()))

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

	max := container.NewMax(alpha_box, tabs)

	return max
}

func placeWall() *container.Split {
	cont := container.NewHScroll(menu.DaemonRpcEntry())
	cont.SetMinSize(fyne.NewSize(340, 35.1875))

	asset_items := container.NewVBox(
		table.DreamsEntry(),
		table.DreamsOpts())

	player_input := container.NewVBox(
		cont,
		menu.WalletRpcEntry(),
		menu.UserPassEntry(),
		menu.RpcConnectButton(),
		layout.NewSpacer(),
		asset_items)

	check_boxes := container.NewVBox(
		menu.DaemonConnectedBox())

	player_box := container.NewHBox(player_input, check_boxes)
	menu_top := container.NewHSplit(player_box, menu.IntroTree())

	menu_bottom := container.NewAdaptiveGrid(1, layout.NewSpacer())

	menu_box := container.NewVSplit(menu_top, menu_bottom)
	menu_box.SetOffset(1)

	return menu_box
}

func placeContract() *container.Split {
	cont := container.NewHScroll(menu.HolderoContractEntry())
	cont.SetMinSize(fyne.NewSize(640, 35.1875))

	unlock_box := container.NewVBox(
		layout.NewSpacer(),
		menu.HolderoUnlockButton(),
		menu.NewTableButton())

	new_box := container.NewVBox(
		layout.NewSpacer(),
		menu.BettingUnlockButton(),
		menu.NewBettingButton())

	grid := container.NewAdaptiveGrid(2, unlock_box, new_box)
	asset_items := container.NewAdaptiveGrid(1, menu.TableStats())

	player_input := container.NewVBox(
		cont,
		asset_items,
		layout.NewSpacer(),
		grid)

	check_box := container.NewVBox(menu.HolderoContractConnectedBox())

	tabs := container.NewAppTabs(
		container.NewTabItem("Tables", menu.TableListings()),
		container.NewTabItem("Favorites", menu.HolderoFavorites()),
		container.NewTabItem("Owned", menu.MyTables()))

	tabs.OnSelected = func(ti *container.TabItem) {
		MenuContractTab(ti)
	}

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	max := container.NewMax(alpha, tabs)

	player_box := container.NewHBox(player_input, check_box)
	menu_top := container.NewHSplit(player_box, max)

	mid := container.NewVBox(layout.NewSpacer(), container.NewAdaptiveGrid(2, menu.NameEntry(), table.TournamentButton()), menu.OwnersBoxMid())

	menu_bottom := container.NewGridWithColumns(3, menu.OwnersBoxLeft(), mid, prediction.OwnerButton())

	menuBox := container.NewVSplit(menu_top, menu_bottom)
	menuBox.SetOffset(1)

	return menuBox
}

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

	player_input := container.NewVBox(
		items_box,
		layout.NewSpacer(),
		table.SetHeaderItems())

	tabs := container.NewAppTabs(
		container.NewTabItem("Owned", menu.AssetList()))

	tabs.OnSelected = func(ti *container.TabItem) {

	}

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	max := container.NewMax(alpha, tabs)

	player_box := container.NewHBox(player_input)
	menu_top := container.NewHSplit(player_box, max)

	menu_bottom := container.NewAdaptiveGrid(1, menu.IndexEntry())

	menu_box := container.NewVSplit(menu_top, menu_bottom)
	menu_box.SetOffset(1)

	return menu_box
}

func placeMarket() *container.Split {
	details := container.NewMax(menu.NfaMarketInfo())

	tabs := container.NewAppTabs(
		container.NewTabItem("Auctions", menu.AuctionListings()),
		container.NewTabItem("Buy Now", menu.BuyNowListings()))

	tabs.SetTabLocation(container.TabLocationTop)
	tabs.OnSelected = func(ti *container.TabItem) {
		MarketTab(ti)
	}
	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	max := container.NewMax(alpha, tabs)

	box := container.NewVBox(
		layout.NewSpacer(),
		details)

	menu_top := container.NewHSplit(box, max)
	menu_top.SetOffset(0)

	menu_bottom := container.NewAdaptiveGrid(1, menu.MarketItems())

	menu_box := container.NewVSplit(menu_top, menu_bottom)
	menu_box.SetOffset(1)

	return menu_box
}

func placeHoldero() *fyne.Container {
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
		table.SitButton(),
		table.LeaveButton(),
		table.DealHandButton(),
		table.CheckButton(),
		table.BetButton(),
		table.BetAmount())

	options := container.NewVBox(layout.NewSpacer(), table.AutoOptions())

	holdero_actions := container.NewHBox(options, layout.NewSpacer(), table.TimeOutWarning(), layout.NewSpacer(), layout.NewSpacer(), &H.ActionButtons)

	H.TableItems = container.NewVBox(
		labelColorBlack(holdero_label),
		&H.TableContent,
		&H.CardsContent,
		layout.NewSpacer(),
		holdero_actions)

	return H.TableItems
}

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

func placePredict() *fyne.Container {
	cont := container.NewHScroll(prediction.PreictionContractEntry())
	cont.SetMinSize(fyne.NewSize(600, 35.1875))
	predict_info := container.NewVBox(prediction.PredictControl.Info, prediction.PredictControl.Prices)
	predict_scroll := container.NewScroll(predict_info)
	predict_scroll.SetMinSize(fyne.NewSize(540, 500))

	check_box := container.NewVBox(
		prediction.PredictConnectedBox())

	hbox := container.NewHBox(cont, check_box)
	predict_content := container.NewVBox(
		hbox,
		predict_scroll,
		layout.NewSpacer(),
		prediction.PredictBox())

	// leaders_scroll := container.NewScroll(prediction.LeadersDisplay())
	// leaders_scroll.SetMinSize(fyne.NewSize(180, 500))
	// leaders_contnet := container.NewVBox(leaders_scroll)

	tabs := container.NewAppTabs(
		container.NewTabItem("Contracts", prediction.PredictionListings()),
		container.NewTabItem("Favorites", prediction.PredicitionFavorites()),
		container.NewTabItem("Owned", prediction.PredictionOwned()))
	// container.NewTabItem("Leaderboard", leaders_contnet))

	tabs.OnSelected = func(ti *container.TabItem) {
		PredictTab(ti)
	}

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	max := container.NewMax(alpha, tabs)

	predict_label := container.NewHBox(P.LeftLabel, layout.NewSpacer(), P.RightLabel)
	predict_box := container.NewHSplit(predict_content, max)

	P.TableItems = container.NewVBox(
		labelColorBlack(predict_label),
		predict_box)

	return P.TableItems
}

func placeSports() *fyne.Container {
	cont := container.NewHScroll(prediction.SportsContractEntry())
	cont.SetMinSize(fyne.NewSize(600, 35.1875))
	sports_content := container.NewVBox(prediction.SportsControl.Info)
	sports_scroll := container.NewVScroll(sports_content)
	sports_scroll.SetMinSize(fyne.NewSize(180, 500))

	check_box := container.NewVBox(
		prediction.SportsConnectedBox())
	hbox := container.NewHBox(cont, check_box)

	sports_left := container.NewVBox(
		hbox,
		sports_scroll,
		layout.NewSpacer(),
		prediction.SportsBox())

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
		container.NewTabItem("UFC", ufc_scroll),
	)

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

	tabs := container.NewAppTabs(
		container.NewTabItem("Contracts", prediction.SportsListings()),
		container.NewTabItem("Favorites", prediction.SportsFavorites()),
		container.NewTabItem("Owned", prediction.SportsOwned()),
		container.NewTabItem("Scores", score_tabs),
		container.NewTabItem("Payouts", prediction.SportsPayouts()))

	tabs.OnSelected = func(ti *container.TabItem) {

	}

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	max := container.NewMax(alpha, tabs)

	sports_label := container.NewHBox(S.LeftLabel, layout.NewSpacer(), S.RightLabel)
	sports_box := container.NewHSplit(sports_left, max)

	S.TableItems = container.NewVBox(
		labelColorBlack(sports_label),
		sports_box)

	return S.TableItems
}

//go:embed table/iluma/iluma.txt
var iluma_intro string

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
