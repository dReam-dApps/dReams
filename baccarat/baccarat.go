package baccarat

import (
	"encoding/hex"
	"encoding/json"
	"image/color"
	"strconv"

	"github.com/SixofClubsss/dReams/dwidget"
	"github.com/SixofClubsss/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type baccObjects struct {
	Actions *fyne.Container
	Map     map[string]string
}

var Table baccObjects

// Baccarat object buffer when action triggered
func BaccBuffer(d bool) {
	if d {
		Table.Actions.Hide()
		rpc.Bacc.P_card1 = 99
		rpc.Bacc.P_card2 = 99
		rpc.Bacc.P_card3 = 99
		rpc.Bacc.B_card1 = 99
		rpc.Bacc.B_card2 = 99
		rpc.Bacc.B_card3 = 99
		rpc.Bacc.Last = ""
		rpc.Display.BaccRes = "Wait for Block..."
	} else {
		if rpc.Daemon.Connect && rpc.Display.BaccRes != "Wait for Block..." {
			Table.Actions.Show()
		}
	}

	Table.Actions.Refresh()
}

// Baccarat hand result display label
func BaccResult(r string) *canvas.Text {
	label := canvas.NewText(r, color.White)
	label.Move(fyne.NewPos(564, 237))
	label.Alignment = fyne.TextAlignCenter

	return label
}

// Baccarat action objects
func BaccaratButtons(w fyne.Window) fyne.CanvasObject {
	entry := dwidget.DeroAmtEntry("", 1, 0)
	entry.PlaceHolder = "dReams:"
	entry.SetText("10")
	entry.Validator = validation.NewRegexp(`^\d{1,}$`, "Int required")
	entry.OnChanged = func(s string) {
		if rpc.Daemon.Connect {
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				if f < 10 {
					entry.SetText("10")
				}

				if f > 250 {
					entry.SetText("250")
				}
			}

			if entry.Validate() != nil {
				entry.SetText("10")
			}
		}

	}

	player_button := widget.NewButton("Player", func() {
		BaccBuffer(true)
		rpc.Bacc.Found = false
		rpc.Bacc.Display = false
		if tx := rpc.BaccBet(entry.Text, "player"); tx == "ID error" {
			dialog.NewInformation("Baccarat", "Select a table", w).Show()
		}
	})

	banker_button := widget.NewButton("Banker", func() {
		BaccBuffer(true)
		rpc.Bacc.Found = false
		rpc.Bacc.Display = false
		if tx := rpc.BaccBet(entry.Text, "banker"); tx == "ID error" {
			dialog.NewInformation("Baccarat", "Select a table", w).Show()
		}
	})

	tie_button := widget.NewButton("Tie", func() {
		BaccBuffer(true)
		rpc.Bacc.Found = false
		rpc.Bacc.Display = false
		if tx := rpc.BaccBet(entry.Text, "tie"); tx == "ID error" {
			dialog.NewInformation("Baccarat", "Select a table", w).Show()
		}
	})

	amt_box := container.NewHScroll(entry)
	amt_box.SetMinSize(fyne.NewSize(100, 40))

	actions := container.NewVBox(
		player_button,
		banker_button,
		tie_button,
		amt_box)

	var searched string
	search_entry := widget.NewEntry()
	search_entry.SetPlaceHolder("TXID:")
	search_button := widget.NewButton("    Search   ", func() {
		txid := search_entry.Text
		if len(txid) == 64 && txid != searched {
			searched = txid
			BaccBuffer(true)
			rpc.Display.BaccRes = "Searching..."
			rpc.Bacc.Found = false
			rpc.Bacc.Display = false
			rpc.FetchBaccHand(txid)
			if !rpc.Bacc.Found {
				rpc.Display.BaccRes = "Hand Not Found"
				BaccBuffer(false)
			}
		}
	})

	table_opts := []string{"dReams"}
	table_select := widget.NewSelect(table_opts, func(s string) {
		switch s {
		case "dReams":
			rpc.Bacc.Contract = rpc.BaccSCID
		default:
			rpc.Bacc.Contract = Table.Map[s]
		}
		rpc.FetchBaccSC()
	})
	table_select.PlaceHolder = "Select Table:"
	table_select.SetSelectedIndex(0)

	search := container.NewVBox(
		layout.NewSpacer(),
		container.NewAdaptiveGrid(2,
			table_select,
			container.NewBorder(nil, nil, nil, search_button, search_entry)))

	Table.Actions = container.NewVBox(
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), actions),
		search)

	Table.Actions.Hide()

	return Table.Actions
}

// Baccarat table image
func BaccTable(img fyne.Resource) fyne.CanvasObject {
	table_img := canvas.NewImageFromResource(img)
	table_img.Resize(fyne.NewSize(1100, 600))
	table_img.Move(fyne.NewPos(5, 0))

	return table_img
}

// Gets list of current Baccarat tables from on chain store and refresh options
func GetBaccTables() {
	if rpc.Daemon.Connect {
		if table_map, ok := rpc.FindStringKey(rpc.RatingSCID, "bacc_tables", rpc.Daemon.Rpc).(string); ok {
			if str, err := hex.DecodeString(table_map); err == nil {
				json.Unmarshal([]byte(str), &Table.Map)
			}
		}

		table_names := make([]string, 0, len(Table.Map))
		for name := range Table.Map {
			table_names = append(table_names, name)
		}

		table_select := Table.Actions.Objects[2].(*fyne.Container).Objects[1].(*fyne.Container).Objects[0].(*widget.Select)
		table_select.Options = []string{"dReams"}
		table_select.Options = append(table_select.Options, table_names...)
		table_select.Refresh()
	}
}
