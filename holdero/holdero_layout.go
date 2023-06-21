package holdero

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	dreams "github.com/SixofClubsss/dReams"
	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/dwidget"
	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"
)

// Holdero tables menu tab layout
func placeContract(change_screen *fyne.Container, H *dwidget.DreamsItems, d dreams.DreamsObject) *container.Split {
	Settings.Check = widget.NewCheck("", func(b bool) {
		if !b {
			disableOwnerControls(true)
		}
	})
	Settings.Check.Disable()

	check_box := container.NewVBox(Settings.Check)

	var tabs *container.AppTabs
	Poker.Holdero_unlock = widget.NewButton("Unlock Holdero Contract", nil)
	Poker.Holdero_unlock.Hide()

	Poker.Holdero_new = widget.NewButton("New Holdero Table", nil)
	Poker.Holdero_new.Hide()

	unlock_cont := container.NewVBox(
		layout.NewSpacer(),
		Poker.Holdero_unlock,
		Poker.Holdero_new)

	owner_buttons := container.NewAdaptiveGrid(2, container.NewMax(layout.NewSpacer()), unlock_cont)
	owned_tab := container.NewBorder(nil, owner_buttons, nil, nil, myTables())

	tabs = container.NewAppTabs(
		container.NewTabItem("Tables", layout.NewSpacer()),
		container.NewTabItem("Favorites", holderoFavorites()),
		container.NewTabItem("Owned", owned_tab),
		container.NewTabItem("View Table", layout.NewSpacer()))

	tabs.SelectIndex(0)
	tabs.Selected().Content = tableListings(tabs)

	tabs.OnSelected = func(ti *container.TabItem) {
		switch ti.Text {
		case "Tables":
			if rpc.Daemon.IsConnected() {
				go createTableList()
			}

		default:
		}

		if ti.Text == "View Table" {
			go func() {
				if len(Round.Contract) == 64 {
					FetchHolderoSC()
					tables_menu = false
					d.Window.Content().(*fyne.Container).Objects[1].(*fyne.Container).Objects[1].(*container.AppTabs).Selected().Content = change_screen
					d.Window.Content().(*fyne.Container).Objects[1].(*fyne.Container).Objects[1].(*container.AppTabs).Selected().Content.Refresh()
					tabs.SelectIndex(0)
					now := time.Now().Unix()
					if now > Round.Last+33 {
						holderoRefresh(H, d, 0)
					}
				} else {
					tabs.SelectIndex(0)
				}
			}()
		}
	}

	max := container.NewMax(bundle.Alpha120, tabs)

	Poker.Holdero_unlock.OnTapped = func() {
		max.Objects[1] = holderoMenuConfirm(1, max.Objects, tabs)
		max.Objects[1].Refresh()
	}

	Poker.Holdero_new.OnTapped = func() {
		max.Objects[1] = holderoMenuConfirm(2, max.Objects, tabs)
		max.Objects[1].Refresh()
	}

	mid := container.NewVBox(layout.NewSpacer(), container.NewAdaptiveGrid(2, menu.NameEntry(), TournamentButton(max.Objects, tabs)), ownersBoxMid())

	menu_bottom := container.NewGridWithColumns(3, ownersBoxLeft(max.Objects, tabs), mid, layout.NewSpacer())

	contract_cont := container.NewHScroll(holderoContractEntry())
	contract_cont.SetMinSize(fyne.NewSize(640, 35.1875))

	asset_items := container.NewAdaptiveGrid(1, container.NewVBox(displayTableStats()))

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

// Holdero tab layout
func placeHoldero(change_screen *widget.Button, h *dwidget.DreamsItems, d dreams.DreamsObject) *fyne.Container {
	h.Back = *container.NewWithoutLayout(
		HolderoTable(bundle.ResourcePokerTablePng),
		Player1_label(nil, nil, nil),
		Player2_label(nil, nil, nil),
		Player3_label(nil, nil, nil),
		Player4_label(nil, nil, nil),
		Player5_label(nil, nil, nil),
		Player6_label(nil, nil, nil),
		h.TopLabel)

	holdero_label := container.NewHBox(h.LeftLabel, layout.NewSpacer(), h.RightLabel)

	h.Front = *placeHolderoCards(d.Window)

	h.Actions = *container.NewVBox(
		layout.NewSpacer(),
		SitButton(),
		LeaveButton(),
		DealHandButton(),
		CheckButton(),
		BetButton(),
		BetAmount())

	options := container.NewVBox(layout.NewSpacer(), AutoOptions(), change_screen)

	holdero_actions := container.NewHBox(options, layout.NewSpacer(), TimeOutWarning(), layout.NewSpacer(), layout.NewSpacer(), &h.Actions)

	h.DApp = container.NewVBox(
		dwidget.LabelColor(holdero_label),
		&h.Back,
		&h.Front,
		layout.NewSpacer(),
		holdero_actions)

	return h.DApp
}

// Layout all objects for Holdero dApp
func LayoutAllItems(h *dwidget.DreamsItems, d dreams.DreamsObject) *container.Split {
	var holdero_objs *fyne.Container
	var contract_objs *container.Split
	contract_change_screen := widget.NewButton("Tables", nil)
	contract_change_screen.OnTapped = func() {
		go func() {
			tables_menu = true
			d.Window.Content().(*fyne.Container).Objects[1].(*fyne.Container).Objects[1].(*container.AppTabs).Selected().Content = contract_objs
			d.Window.Content().(*fyne.Container).Objects[1].(*fyne.Container).Objects[1].(*container.AppTabs).Selected().Content.Refresh()
		}()
	}

	tables_menu = true
	holdero_objs = placeHoldero(contract_change_screen, h, d)
	contract_objs = placeContract(holdero_objs, h, d)

	// Main process
	go fetch(h, d)

	return contract_objs
}
