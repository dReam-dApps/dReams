package prediction

import (
	"sort"
	"strings"

	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/table"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type predictItems struct {
	Contract        string
	Leaders_map     map[string]uint64
	Leaders_display []string
	Leaders_list    *widget.List
	Predict_list    *widget.List
	Remove_button   *widget.Button
}

var PredictControl predictItems

func PredictConnectedBox() fyne.Widget {
	menu.MenuControl.Predict_check = widget.NewCheck("", func(b bool) {})
	menu.MenuControl.Predict_check.Disable()

	return menu.MenuControl.Predict_check
}

func PreictionContractEntry() fyne.Widget {
	options := []string{}
	table.Actions.P_contract = widget.NewSelectEntry(options)
	table.Actions.P_contract.PlaceHolder = "Contract Address: "
	table.Actions.P_contract.OnCursorChanged = func() {
		if rpc.Signal.Daemon {
			yes, _ := rpc.CheckBetContract(PredictControl.Contract)
			if yes {
				menu.MenuControl.Predict_check.SetChecked(true)
			} else {
				menu.MenuControl.Predict_check.SetChecked(false)
			}
		}
	}

	this := binding.BindString(&PredictControl.Contract)
	table.Actions.P_contract.Bind(this)

	return table.Actions.P_contract
}

func PredictBox() fyne.CanvasObject {
	table.Actions.NameEntry = widget.NewEntry()
	table.Actions.NameEntry.SetPlaceHolder("Name")
	table.Actions.NameEntry.OnChanged = func(input string) {
		table.Actions.NameEntry.Validator = validation.NewRegexp(`\w{3,}`, "Three Letters Minimum")
		table.Actions.NameEntry.Validate()
		table.Actions.NameEntry.Refresh()
	}

	table.Actions.Change = widget.NewButton("Change Name", func() {
		if table.Actions.NameEntry.Disabled() {
			table.Actions.NameEntry.Enable()
		} else {
			namePopUp(1)
		}
	})

	table.Actions.Higher = widget.NewButton("Higher", func() {
		if len(PredictControl.Contract) == 64 {
			confirmPopUp(2, "", "")
		}
	})

	table.Actions.Lower = widget.NewButton("Lower", func() {
		if len(PredictControl.Contract) == 64 {
			confirmPopUp(1, "", "")
		}
	})

	table.Actions.NameEntry.Hide()
	table.Actions.Change.Hide()
	table.Actions.Higher.Hide()
	table.Actions.Lower.Hide()

	table.Actions.Prediction_box = container.NewVBox(table.Actions.NameEntry, table.Actions.Change, table.Actions.Higher, table.Actions.Lower)
	table.Actions.Prediction_box.Hide()

	return table.Actions.Prediction_box
}

func LeadersDisplay() fyne.Widget {
	PredictControl.Leaders_display = []string{}
	PredictControl.Leaders_list = widget.NewList(
		func() int {
			return len(PredictControl.Leaders_display)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(PredictControl.Leaders_display[i])
		})

	return PredictControl.Leaders_list
}

func PredictionListings() fyne.CanvasObject { /// prediction contract list
	PredictControl.Predict_list = widget.NewList(
		func() int {
			return len(menu.MenuControl.Predict_contracts)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(menu.MenuControl.Predict_contracts[i])
		})

	var item string

	PredictControl.Predict_list.OnSelected = func(id widget.ListItemID) {
		if id != 0 {
			if menu.Connected() {
				split := strings.Split(menu.MenuControl.Predict_contracts[id], "   ")
				trimmed := strings.Trim(split[2], " ")
				if len(trimmed) == 64 {
					item = menu.MenuControl.Predict_contracts[id]
					table.Actions.P_contract.SetText(trimmed)
					if menu.CheckActivePrediction(trimmed) {
						menu.DisablePreditions(false)
						table.Actions.Higher.Show()
						table.Actions.Lower.Show()
						table.Actions.NameEntry.Show()
						table.Actions.NameEntry.Text = menu.CheckPredictionName(PredictControl.Contract)
						table.Actions.NameEntry.Refresh()
					} else {
						menu.DisablePreditions(true)
					}
				}
			} else {
				menu.DisablePreditions(true)
			}
		}
	}

	save := widget.NewButton("Favorite", func() {
		menu.MenuControl.Predict_favorites = append(menu.MenuControl.Predict_favorites, item)
		sort.Strings(menu.MenuControl.Predict_favorites)
	})

	cont := container.NewBorder(
		nil,
		container.NewBorder(nil, nil, nil, save, layout.NewSpacer()),
		nil,
		nil,
		PredictControl.Predict_list)

	return cont
}

func PredicitFavorites() fyne.CanvasObject {
	favorites := widget.NewList(
		func() int {
			return len(menu.MenuControl.Predict_favorites)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(menu.MenuControl.Predict_favorites[i])
		})

	var item string

	favorites.OnSelected = func(id widget.ListItemID) {
		if menu.Connected() {
			split := strings.Split(menu.MenuControl.Predict_favorites[id], "   ")
			if len(split) >= 3 {
				trimmed := strings.Trim(split[2], " ")
				if len(trimmed) == 64 {
					if len(trimmed) == 64 {
						item = menu.MenuControl.Predict_favorites[id]
						table.Actions.P_contract.SetText(trimmed)
						if menu.CheckActivePrediction(trimmed) {
							menu.DisablePreditions(false)
							table.Actions.Higher.Show()
							table.Actions.Lower.Show()
							table.Actions.NameEntry.Show()
							table.Actions.NameEntry.Text = menu.CheckPredictionName(PredictControl.Contract)
							table.Actions.NameEntry.Refresh()
						} else {
							menu.DisablePreditions(true)
						}
					}
				}
			} else {
				menu.DisablePreditions(true)
			}
		}
	}

	remove := widget.NewButton("Remove", func() {
		if len(menu.MenuControl.Predict_favorites) > 0 {
			favorites.UnselectAll()
			new := menu.MenuControl.Predict_favorites
			for i := range new {
				if new[i] == item {
					copy(new[i:], new[i+1:])
					new[len(new)-1] = ""
					new = new[:len(new)-1]
					menu.MenuControl.Predict_favorites = new
					break
				}
			}
		}
		favorites.Refresh()
		sort.Strings(menu.MenuControl.Predict_favorites)
	})

	cont := container.NewBorder(
		nil,
		container.NewBorder(nil, nil, nil, remove, layout.NewSpacer()),
		nil,
		nil,
		favorites)

	return cont
}

func Remove() fyne.Widget {
	PredictControl.Remove_button = widget.NewButton("Remove", func() {
		namePopUp(2)
	})

	PredictControl.Remove_button.Hide()

	return PredictControl.Remove_button
}
