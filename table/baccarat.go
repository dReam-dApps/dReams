package table

import (
	"strconv"

	"github.com/SixofClubsss/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Items struct {
	LeftLabel   *widget.Label
	RightLabel  *widget.Label
	TopLabel    *widget.Label
	BottomLabel *widget.Label

	TableContent  fyne.Container
	CardsContent  fyne.Container
	ActionButtons fyne.Container
	TableItems    *fyne.Container
}

type baccAmt struct {
	NumericalEntry
}

func (e *baccAmt) TypedKey(k *fyne.KeyEvent) {
	switch k.Name {
	case fyne.KeyUp:
		if f, err := strconv.ParseFloat(e.Entry.Text, 64); err == nil {
			e.Entry.SetText(strconv.FormatFloat(float64(f+0.1), 'f', 1, 64))
		}
	case fyne.KeyDown:
		if f, err := strconv.ParseFloat(e.Entry.Text, 64); err == nil {
			if f >= 0.1 {
				e.Entry.SetText(strconv.FormatFloat(float64(f-0.1), 'f', 1, 64))
			}
		}
	}
	e.Entry.TypedKey(k)
}

func BaccBuffer(d bool) {
	if d {
		Actions.Bacc_actions.Hide()
		rpc.Bacc.P_card1 = 99
		rpc.Bacc.P_card2 = 99
		rpc.Bacc.P_card3 = 99
		rpc.Bacc.B_card1 = 99
		rpc.Bacc.B_card2 = 99
		rpc.Bacc.B_card3 = 99
		rpc.Bacc.Last = ""
		rpc.Display.BaccRes = "Wait for Block..."
	} else {
		if rpc.Signal.Daemon {
			Actions.Bacc_actions.Show()
		}
	}

	Actions.Bacc_actions.Refresh()
}

func BaccResult(r string) fyne.Widget {
	label := widget.NewLabel(r)
	label.Move(fyne.NewPos(485, 225))

	return label
}

func BaccaratButtons() fyne.CanvasObject {
	entry := &baccAmt{}
	entry.ExtendBaseWidget(entry)
	entry.PlaceHolder = "dReams:"
	entry.SetText("10")
	entry.Validator = validation.NewRegexp(`\d{1,}$`, "Format Not Valid")
	entry.OnChanged = func(s string) {
		if rpc.Signal.Daemon {
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

	vBox := container.NewVBox(
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

	hBox := container.NewHBox(layout.NewSpacer(), vBox)
	Actions.Bacc_actions = container.NewVBox(layout.NewSpacer(), hBox, search)

	Actions.Bacc_actions.Hide()

	return Actions.Bacc_actions
}

func BaccTable(img fyne.Resource) fyne.CanvasObject {
	table := canvas.NewImageFromResource(img)
	table.Resize(fyne.NewSize(1100, 600))
	table.Move(fyne.NewPos(5, 0))

	return table
}
