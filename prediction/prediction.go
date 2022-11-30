package prediction

import (
	"strings"

	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/table"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/widget"
)

type predictItems struct {
	Contract        string
	Leaders_map     map[string]uint64
	Leaders_display []string
	Contract_list   []string
	connected_box   *widget.Check
	Leaders_list    *widget.List
	Predict_list    *widget.List
	Remove_button   *widget.Button
}

var PredictControl predictItems

func PredictConnectedBox() fyne.Widget {
	PredictControl.connected_box = widget.NewCheck("", func(b bool) {})
	PredictControl.connected_box.Disable()

	return PredictControl.connected_box
}

func PreictionContractEntry() fyne.Widget {
	options := []string{}
	table.Actions.P_contract = widget.NewSelectEntry(options)
	table.Actions.P_contract.PlaceHolder = "Contract Address: "
	table.Actions.P_contract.OnCursorChanged = func() {
		if rpc.Signal.Daemon {
			yes, _ := rpc.CheckBetContract(PredictControl.Contract)
			if yes {
				PredictControl.connected_box.SetChecked(true)
			} else {
				PredictControl.connected_box.SetChecked(false)
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

func PredictionListings() fyne.Widget { /// prediction contract list
	PredictControl.Predict_list = widget.NewList(
		func() int {
			return len(PredictControl.Contract_list)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(PredictControl.Contract_list[i])
		})

	PredictControl.Predict_list.OnSelected = func(id widget.ListItemID) {
		if id != 0 {
			if rpc.Signal.Daemon && rpc.Wallet.Connect {
				split := strings.Split(PredictControl.Contract_list[id], "   ")
				trimmed := strings.Trim(split[2], " ")
				if len(trimmed) == 64 {
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

	return PredictControl.Predict_list
}

func Remove() fyne.Widget {
	PredictControl.Remove_button = widget.NewButton("Remove", func() {
		namePopUp(2)
	})

	PredictControl.Remove_button.Hide()

	return PredictControl.Remove_button
}
