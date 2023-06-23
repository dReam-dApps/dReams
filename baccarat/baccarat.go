package baccarat

import (
	"encoding/hex"
	"encoding/json"
	"image/color"
	"log"
	"strconv"
	"time"

	dreams "github.com/SixofClubsss/dReams"
	"github.com/SixofClubsss/dReams/dwidget"
	"github.com/SixofClubsss/dReams/holdero"
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

func initValues() {
	Bacc.Display = true
}

// Main Baccarat process
func fetch(b *dreams.DreamsItems, d dreams.DreamsObject) {
	initValues()
	time.Sleep(3 * time.Second)
	for {
		select {
		case <-d.Receive():
			if !rpc.Wallet.IsConnected() || !rpc.Daemon.IsConnected() {
				disableBaccActions(true)
				BaccRefresh(b, d)
				d.WorkDone()
				continue
			}

			fetchBaccSC()
			BaccRefresh(b, d)
			d.WorkDone()
		case <-d.CloseDapp():
			log.Println("[Baccarat] Done")
			return
		}
	}
}

// Baccarat object buffer when action triggered
func ActionBuffer(d bool) {
	if d {
		Table.Actions.Hide()
		Bacc.P_card1 = 99
		Bacc.P_card2 = 99
		Bacc.P_card3 = 99
		Bacc.B_card1 = 99
		Bacc.B_card2 = 99
		Bacc.B_card3 = 99
		Bacc.Last = ""
		Display.BaccRes = "Wait for Block..."
	} else {
		if rpc.Daemon.IsConnected() && Display.BaccRes != "Wait for Block..." {
			Table.Actions.Show()
		}
	}

	Table.Actions.Refresh()
}

// Disable Baccarat actions
func disableBaccActions(d bool) {
	if d {
		Table.Actions.Hide()
	} else {
		Table.Actions.Show()
	}

	Table.Actions.Refresh()
}

// Baccarat hand result display label
func baccResult(r string) *canvas.Text {
	label := canvas.NewText(r, color.White)
	label.Move(fyne.NewPos(564, 237))
	label.Alignment = fyne.TextAlignCenter

	return label
}

// Baccarat action objects
func baccaratButtons(w fyne.Window) fyne.CanvasObject {
	entry := dwidget.DeroAmtEntry("", 1, 0)
	entry.PlaceHolder = "dReams:"
	entry.AllowFloat = false
	entry.SetText("10")
	entry.Validator = validation.NewRegexp(`^\d{1,}$`, "Int required")
	entry.OnChanged = func(s string) {
		if rpc.Daemon.IsConnected() {
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				if f < Bacc.MinBet {
					entry.SetText(Display.BaccMin)
				}

				if f > Bacc.MaxBet {
					entry.SetText(Display.BaccMax)
				}
			}

			if entry.Validate() != nil {
				entry.SetText(Display.BaccMin)
			}
		}
	}

	player_button := widget.NewButton("Player", func() {
		ActionBuffer(true)
		Bacc.Found = false
		Bacc.Display = false
		if tx := BaccBet(entry.Text, "player"); tx == "ID error" {
			dialog.NewInformation("Baccarat", "Select a table", w).Show()
		}
	})

	banker_button := widget.NewButton("Banker", func() {
		ActionBuffer(true)
		Bacc.Found = false
		Bacc.Display = false
		if tx := BaccBet(entry.Text, "banker"); tx == "ID error" {
			dialog.NewInformation("Baccarat", "Select a table", w).Show()
		}
	})

	tie_button := widget.NewButton("Tie", func() {
		ActionBuffer(true)
		Bacc.Found = false
		Bacc.Display = false
		if tx := BaccBet(entry.Text, "tie"); tx == "ID error" {
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
	search_button := widget.NewButton("     Search    ", func() {
		txid := search_entry.Text
		if len(txid) == 64 && txid != searched {
			searched = txid
			ActionBuffer(true)
			Display.BaccRes = "Searching..."
			Bacc.Found = false
			Bacc.Display = false
			FetchBaccHand(txid)
			if !Bacc.Found {
				Display.BaccRes = "Hand Not Found"
				ActionBuffer(false)
			}
		}
	})

	Display.BaccMin = "10"
	table_opts := []string{"dReams"}
	table_select := widget.NewSelect(table_opts, func(s string) {
		switch s {
		case "dReams":
			Bacc.Contract = rpc.BaccSCID
		default:
			Bacc.Contract = Table.Map[s]
		}
		fetchBaccSC()
		entry.SetText(Display.BaccMin)
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
	if rpc.Daemon.IsConnected() {
		Table.Map = make(map[string]string)
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

// Place and refresh Baccarat card images
func showBaccCards() *fyne.Container {
	var drawP, drawB int
	if Bacc.P_card3 == 0 {
		drawP = 99
	} else {
		drawP = Bacc.P_card3
	}

	if Bacc.B_card3 == 0 {
		drawB = 99
	} else {
		drawB = Bacc.B_card3
	}

	content := *container.NewWithoutLayout(
		holdero.PlayerCards(holdero.BaccSuit(Bacc.P_card1), holdero.BaccSuit(Bacc.P_card2), holdero.BaccSuit(drawP)),
		holdero.BankerCards(holdero.BaccSuit(Bacc.B_card1), holdero.BaccSuit(Bacc.B_card2), holdero.BaccSuit(drawB)))

	Bacc.Display = true
	ActionBuffer(false)

	return &content
}

func clearBaccCards() *fyne.Container {
	content := *container.NewWithoutLayout(
		holdero.PlayerCards(99, 99, 99),
		holdero.BankerCards(99, 99, 99))

	return &content
}

// Refresh all Baccarat objects
func BaccRefresh(b *dreams.DreamsItems, d dreams.DreamsObject) {
	asset_name := rpc.GetAssetSCIDName(Bacc.AssetID)
	b.LeftLabel.SetText("Total Hands Played: " + Display.Total_w + "      Player Wins: " + Display.Player_w + "      Ties: " + Display.Ties + "      Banker Wins: " + Display.Banker_w + "      Min Bet is " + Display.BaccMin + " " + asset_name + ", Max Bet is " + Display.BaccMax)
	b.RightLabel.SetText(asset_name + " Balance: " + rpc.DisplayBalance(asset_name) + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Wallet.Display.Height)

	if !Bacc.Display {
		b.Front.Objects[0] = clearBaccCards()
		FetchBaccHand(Bacc.Last)
		if Bacc.Found {
			b.Front.Objects[0] = showBaccCards()
		}
		b.Front.Objects[0].Refresh()
	}

	if rpc.Wallet.Height > Bacc.CHeight+3 && !Bacc.Found {
		Display.BaccRes = ""
		ActionBuffer(false)
	}

	b.Back.Objects[1].(*canvas.Text).Text = Display.BaccRes
	b.Back.Objects[1].Refresh()

	b.DApp.Refresh()

	if Bacc.Found && !Bacc.Notified {
		if !d.IsWindows() {
			Bacc.Notified = d.Notification("dReams - Baccarat", Display.BaccRes)
		}
	}
}
