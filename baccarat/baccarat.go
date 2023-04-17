package baccarat

import (
	"strconv"

	"github.com/SixofClubsss/dReams/dwidget"
	"github.com/SixofClubsss/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type baccObjects struct {
	Actions *fyne.Container
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
		if rpc.Daemon.Connect {
			Table.Actions.Show()
		}
	}

	Table.Actions.Refresh()
}

// Baccarat hand result display
func BaccResult(r string) fyne.Widget {
	label := widget.NewLabel(r)
	label.Move(fyne.NewPos(485, 225))

	return label
}

// Baccarat action objects
func BaccaratButtons() fyne.CanvasObject {
	entry := dwidget.DeroAmtEntry("", 1, 0)
	entry.PlaceHolder = "dReams:"
	entry.SetText("10")
	entry.Validator = validation.NewRegexp(`\d{1,}$`, "Format Not Valid")
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
		rpc.BaccBet(entry.Text, "player")
	})

	banker_button := widget.NewButton("Banker", func() {
		BaccBuffer(true)
		rpc.Bacc.Found = false
		rpc.Bacc.Display = false
		rpc.BaccBet(entry.Text, "banker")
	})

	tie_button := widget.NewButton("Tie", func() {
		BaccBuffer(true)
		rpc.Bacc.Found = false
		rpc.Bacc.Display = false
		rpc.BaccBet(entry.Text, "tie")
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

	search := container.NewVBox(
		layout.NewSpacer(),
		container.NewAdaptiveGrid(2,
			layout.NewSpacer(),
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
