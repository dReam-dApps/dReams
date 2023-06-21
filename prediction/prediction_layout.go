package prediction

import (
	dreams "github.com/SixofClubsss/dReams"
	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/dwidget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// dPrediction tab layout
func LayoutPredictItems(p *dwidget.DreamsItems, d dreams.DreamsObject) *fyne.Container {
	predict_info := container.NewVBox(Predict.Info, Predict.Prices)
	predict_scroll := container.NewScroll(predict_info)
	predict_scroll.SetMinSize(fyne.NewSize(540, 500))

	check_box := container.NewVBox(PredictConnectedBox())

	contract_scroll := container.NewHScroll(PredictionContractEntry())
	contract_scroll.SetMinSize(fyne.NewSize(600, 35.1875))
	contract_cont := container.NewHBox(contract_scroll, check_box)

	Predict.Higher = widget.NewButton("Higher", nil)
	Predict.Higher.Hide()

	Predict.Lower = widget.NewButton("Lower", nil)
	Predict.Lower.Hide()

	Predict.Container = container.NewVBox(Predict.Higher, Predict.Lower)
	Predict.Container.Hide()

	predict_content := container.NewVBox(
		contract_cont,
		predict_scroll,
		layout.NewSpacer(),
		Predict.Container)

	Predict.Settings.Unlock = widget.NewButton("Unlock dPrediction Contract", nil)
	Predict.Settings.Unlock.Hide()

	Predict.Settings.New = widget.NewButton("New dPrediction Contract", nil)
	Predict.Settings.New.Hide()

	unlock_cont := container.NewVBox(Predict.Settings.Unlock, Predict.Settings.New)

	Predict.Settings.Menu = widget.NewButton("Owner Options", func() {
		go ownersMenu()
	})
	Predict.Settings.Menu.Hide()

	owner_buttons := container.NewAdaptiveGrid(2, container.NewMax(Predict.Settings.Menu), unlock_cont)
	owned_tab := container.NewBorder(nil, owner_buttons, nil, nil, PredictionOwned())

	tabs := container.NewAppTabs(
		container.NewTabItem("Contracts", layout.NewSpacer()),
		container.NewTabItem("Favorites", PredictionFavorites()),
		container.NewTabItem("Owned", owned_tab))

	tabs.SelectIndex(0)
	tabs.Selected().Content = PredictionListings(tabs)

	tabs.OnSelected = func(ti *container.TabItem) {
		switch ti.Text {
		case "Contracts":
			go PopulatePredictions(nil)
		default:
		}
	}

	max := container.NewMax(bundle.Alpha120, tabs)

	Predict.Higher.OnTapped = func() {
		if len(Predict.Contract) == 64 {
			max.Objects[1] = ConfirmAction(2, "", "", max.Objects, tabs)
			max.Objects[1].Refresh()
		}
	}

	Predict.Lower.OnTapped = func() {
		if len(Predict.Contract) == 64 {
			max.Objects[1] = ConfirmAction(1, "", "", max.Objects, tabs)
			max.Objects[1].Refresh()
		}
	}

	Predict.Settings.Unlock.OnTapped = func() {
		max.Objects[1] = newPredictConfirm(1, max.Objects, tabs)
		max.Objects[1].Refresh()
	}

	Predict.Settings.New.OnTapped = func() {
		max.Objects[1] = newSportsConfirm(2, max.Objects, tabs)
		max.Objects[1].Refresh()
	}

	predict_label := container.NewHBox(p.LeftLabel, layout.NewSpacer(), p.RightLabel)
	predict_box := container.NewHSplit(predict_content, max)

	p.DApp = container.NewVBox(
		dwidget.LabelColor(predict_label),
		predict_box)

	return p.DApp
}
