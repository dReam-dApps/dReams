package prediction

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/table"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type psOwnerWidgets struct {
	S_end        *widget.Entry
	S_amt        *table.NumericalEntry
	S_game       *widget.Select
	S_league     *widget.SelectEntry
	S_feed       *widget.SelectEntry
	S_deposit    *table.NumericalEntry
	P_end        *widget.Entry
	P_mark       *widget.Entry
	P_amt        *table.NumericalEntry
	P_Name       *widget.SelectEntry
	P_feed       *widget.SelectEntry
	P_deposit    *table.NumericalEntry
	P_cancel     *widget.Button
	Payout_n     *widget.SelectEntry
	Owner_button *widget.Button
}

var PS_Control psOwnerWidgets

func isOnChainPrediction(s string) bool {
	if s == "DERO-Difficulty" || s == "DERO-Block Time" || s == "DERO-Block Number" {
		return true
	}

	return false
}

func onChainPrediction(s string) int {
	switch s {
	case "DERO-Difficulty":
		return 1
	case "DERO-Block Time":
		return 2
	case "DERO-Block Number":
		return 3
	default:
		return 0
	}
}

func preditctionOpts() fyne.CanvasObject { /// set prediction options
	pred := []string{"BTC-USDT", "DERO-USDT", "XMR-USDT", "DERO-Difficulty", "DERO-Block Time", "DERO-Block Number"}
	PS_Control.P_Name = widget.NewSelectEntry(pred)
	PS_Control.P_Name.SetPlaceHolder("Name:")
	PS_Control.P_Name.OnChanged = func(s string) {
		if isOnChainPrediction(s) {
			opts := []string{menu.DAEMON_RPC_REMOTE1, menu.DAEMON_RPC_REMOTE2}
			PS_Control.P_feed.SetOptions(opts)
			if PS_Control.P_feed.Text != opts[1] {
				PS_Control.P_feed.SetText(opts[0])
			}
			PS_Control.P_feed.SetPlaceHolder("Node:")
			PS_Control.P_feed.Refresh()
		} else {
			opts := []string{"dReams Client"}
			PS_Control.P_feed.SetOptions(opts)
			PS_Control.P_feed.SetText(opts[0])
			PS_Control.P_feed.SetPlaceHolder("Feed:")
			PS_Control.P_feed.Refresh()
		}
	}

	PS_Control.P_end = widget.NewEntry()
	PS_Control.P_end.SetPlaceHolder("Closes At:")
	PS_Control.P_end.Validator = validation.NewRegexp(`^\d{10,}$`, "Format Not Valid")

	PS_Control.P_mark = widget.NewEntry()
	PS_Control.P_mark.SetPlaceHolder("Mark:")
	PS_Control.P_mark.Validator = validation.NewRegexp(`^\d{1,}$`, "Format Not Valid")

	PS_Control.P_amt = table.NilNumericalEntry()
	PS_Control.P_amt.SetPlaceHolder("Minimum Amount:")
	PS_Control.P_amt.Validator = validation.NewRegexp(`^\d{1,}\.\d{1,5}$`, "Format Not Valid")

	feeds := []string{"dReams Client"}
	PS_Control.P_feed = widget.NewSelectEntry(feeds)
	PS_Control.P_feed.SetPlaceHolder("Feed:")

	PS_Control.P_deposit = table.NilNumericalEntry()
	PS_Control.P_deposit.SetPlaceHolder("Deposit Amount:")
	PS_Control.P_deposit.Validator = validation.NewRegexp(`^\d{1,}\.\d{1,5}$`, "Format Not Valid")

	confirm := widget.NewButton("Set Prediction", func() {
		if PS_Control.P_deposit.Validate() == nil && PS_Control.P_amt.Validate() == nil && PS_Control.P_end.Validate() == nil && PS_Control.P_mark.Validate() == nil {
			if len(PredictControl.Contract) == 64 {
				ownerConfirmPopUp(2, 100)
			}
		}
	})

	PS_Control.P_cancel = widget.NewButton("Cancel", func() {
		ownerConfirmPopUp(8, 0)
	})

	PS_Control.P_cancel.Hide()

	owner_p := container.NewVBox(
		humanTimeConvert(),
		layout.NewSpacer(),
		PS_Control.P_Name,
		PS_Control.P_end,
		PS_Control.P_mark,
		PS_Control.P_amt,
		PS_Control.P_feed,
		PS_Control.P_deposit,
		confirm,
		layout.NewSpacer(),
		PS_Control.P_cancel,
		layout.NewSpacer(),
	)

	return owner_p
}

func sportsOpts() fyne.CanvasObject { /// set sports options
	options := []string{}
	PS_Control.S_game = widget.NewSelect(options, func(s string) {
		var date string
		game := strings.Split(s, "   ")
		for i := range s {
			if i > 3 {
				date = s[0:10]
			}
		}
		comp := date[0:4] + date[5:7] + date[8:10]
		GetGameEnd(comp, game[1], PS_Control.S_league.Text)
	})
	PS_Control.S_game.PlaceHolder = "Game:"

	leagues := []string{"EPL", "NBA", "NFL", "NHL", "Bellator", "UFC"}
	PS_Control.S_league = widget.NewSelectEntry(leagues)
	PS_Control.S_league.OnChanged = func(s string) {
		PS_Control.S_game.Options = []string{}
		PS_Control.S_game.Selected = ""
		if s == "Bellator" || s == "UFC" {
			PS_Control.S_game.PlaceHolder = "Fight:"
		} else {
			PS_Control.S_game.PlaceHolder = "Game:"
		}
		PS_Control.S_game.Refresh()
		switch s {
		case "EPL":
			go GetCurrentWeek("EPL")
		case "NBA":
			go GetCurrentWeek("NBA")
		case "NFL":
			go GetCurrentWeek("NFL")
		case "NHL":
			go GetCurrentWeek("NHL")
		case "UFC":
			go GetCurrentMonth("UFC")
		case "Bellator":
			go GetCurrentMonth("Bellator")
		default:

		}
	}
	PS_Control.S_league.SetPlaceHolder("League:")

	PS_Control.S_end = widget.NewEntry()
	PS_Control.S_end.SetPlaceHolder("Closes At:")
	PS_Control.S_end.Validator = validation.NewRegexp(`^\d{10,}$`, "Format Not Valid")

	PS_Control.S_amt = table.NilNumericalEntry()
	PS_Control.S_amt.Validator = validation.NewRegexp(`^\d{1,}\.\d{1,5}$`, "Format Not Valid")
	PS_Control.S_amt.SetPlaceHolder("Minimum Amount:")

	feeds := []string{"dReams Client"}
	PS_Control.S_feed = widget.NewSelectEntry(feeds)
	PS_Control.S_feed.SetPlaceHolder("Feed:")

	PS_Control.S_deposit = table.NilNumericalEntry()
	PS_Control.S_deposit.SetPlaceHolder("Deposit Amount:")
	PS_Control.S_deposit.Validator = validation.NewRegexp(`^\d{1,}\.\d{1,5}$`, "Format Not Valid")

	confirmButton := widget.NewButton("Set Game", func() {
		if PS_Control.S_deposit.Validate() == nil && PS_Control.S_amt.Validate() == nil && PS_Control.S_end.Validate() == nil {
			if len(SportsControl.Contract) == 64 {
				ownerConfirmPopUp(1, 100)
			}
		}
	})

	sports := container.NewVBox(
		humanTimeConvert(),
		layout.NewSpacer(),
		PS_Control.S_league,
		PS_Control.S_game,
		PS_Control.S_end,
		PS_Control.S_amt,
		PS_Control.S_feed,
		PS_Control.S_deposit,
		confirmButton,
		layout.NewSpacer())

	return sports
}

func payoutOpts() fyne.CanvasObject {
	PS_Control.Payout_n = widget.NewSelectEntry([]string{})
	PS_Control.Payout_n.SetPlaceHolder("Game #")

	sports_confirm := widget.NewButton("Sports Payout", func() {
		if len(SportsControl.Contract) == 64 {
			ownerConfirmPopUp(3, 100)
		}
	})

	post_button := widget.NewButton("Post", func() {
		go SetPredictionPrices(rpc.Signal.Daemon)
		var a float64
		prediction := rpc.Display.Prediction
		if isOnChainPrediction(prediction) {
			switch onChainPrediction(prediction) {
			case 1:
				a, _ = rpc.GetDifficulty(rpc.Display.P_feed)
				ownerConfirmPopUp(6, a)
			case 2:
				a, _ = rpc.GetBlockTime(rpc.Display.P_feed)
				ownerConfirmPopUp(6, a)
			case 3:
				d, _ := rpc.DaemonHeight(rpc.Display.P_feed)
				a = float64(d)
				ownerConfirmPopUp(6, a)
			default:

			}

		} else {
			a, _ = table.GetPrice(prediction)
			ownerConfirmPopUp(4, a)
		}

	})

	predict_confirm := widget.NewButton("Prediction Payout", func() {
		go SetPredictionPrices(rpc.Signal.Daemon)
		var a float64
		prediction := rpc.Display.Prediction
		if isOnChainPrediction(prediction) {
			switch onChainPrediction(prediction) {
			case 1:
				a, _ = rpc.GetDifficulty(rpc.Display.P_feed)
				ownerConfirmPopUp(7, a)
			case 2:
				a, _ = rpc.GetBlockTime(rpc.Display.P_feed)
				ownerConfirmPopUp(7, a)
			case 3:
				d, _ := rpc.DaemonHeight(rpc.Display.P_feed)
				a = float64(d)
				ownerConfirmPopUp(7, a)
			default:

			}

		} else {
			a, _ = table.GetPrice(prediction)
			ownerConfirmPopUp(5, a)
		}

	})

	payout := container.NewVBox(
		layout.NewSpacer(),
		PS_Control.Payout_n,
		sports_confirm,
		layout.NewSpacer(),
		post_button,
		layout.NewSpacer(),
		predict_confirm,
		layout.NewSpacer())

	return payout
}

func confirmPopUp(i int, teamA, teamB string) { /// bet action confirmation
	ocw := fyne.CurrentApp().NewWindow("Confirm")
	ocw.SetIcon(menu.Resource.SmallIcon)
	ocw.Resize(fyne.NewSize(330, 330))
	ocw.SetFixedSize(true)
	var confirm_display = widget.NewLabel("")
	confirm_display.Wrapping = fyne.TextWrapWord

	p_scid := PredictControl.Contract
	name := table.Actions.NameEntry.Text

	s_scid := SportsControl.Contract
	split := strings.Split(table.Actions.Game_select.Selected, "   ")
	multi := table.Actions.Multi.Selected

	switch i {
	case 1:
		float := float64(rpc.Predict.Amount)
		amt := float / 100000
		a := fmt.Sprintf("%.5f", amt)

		confirm_display.SetText("SCID: " + p_scid + "\n\nLower prediction for " + a + " Dero\n\nConfirm")
	case 2:
		float := float64(rpc.Predict.Amount)
		amt := float / 100000
		a := fmt.Sprintf("%.5f", amt)

		confirm_display.SetText("SCID: " + p_scid + "\n\nHiger prediction for " + a + " Dero\n\nConfirm")
	case 3:
		game := table.Actions.Game_select.Selected
		val := float64(menu.GetSportsAmt(s_scid, split[0]))
		var x string

		switch multi {
		case "3x":
			x = fmt.Sprint(val * 3 / 100000)
		case "5x":
			x = fmt.Sprint(val * 5 / 100000)
		default:
			x = fmt.Sprint(val / 100000)
		}
		confirm_display.SetText("SCID: " + s_scid + "\n\nBetting on Game # " + game + "\n\n" + teamA + " for " + x + " Dero\n\nConfirm")
	case 4:
		game := table.Actions.Game_select.Selected
		val := float64(menu.GetSportsAmt(s_scid, split[0]))
		var x string

		switch multi {
		case "3x":
			x = fmt.Sprint(val * 3 / 100000)
		case "5x":
			x = fmt.Sprint(val * 5 / 100000)
		default:
			x = fmt.Sprint(val / 100000)
		}
		confirm_display.SetText("SCID: " + s_scid + "\n\nBetting on Game # " + game + "\n\n" + teamB + " for " + x + " Dero\n\nConfirm")
	default:
		log.Println("No Confirm Input")
		confirm_display.SetText("Error")
	}

	cancel_button := widget.NewButton("No", func() {
		ocw.Close()
	})
	confirm_button := widget.NewButton("Yes", func() {
		switch i {
		case 1:
			rpc.PredictLower(p_scid, name)
		case 2:
			rpc.PredictHigher(p_scid, name)
		case 3:
			rpc.PickTeam(s_scid, multi, split[0], menu.GetSportsAmt(s_scid, split[0]), 0)
		case 4:
			rpc.PickTeam(s_scid, multi, split[0], menu.GetSportsAmt(s_scid, split[0]), 1)
		default:

		}
		ocw.Close()
	})

	display := container.NewVBox(confirm_display, layout.NewSpacer())
	options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewVBox(display, layout.NewSpacer(), options)

	img := *canvas.NewImageFromResource(menu.Resource.Back2)
	ocw.SetContent(
		container.New(layout.NewMaxLayout(),
			&img,
			content))
	ocw.Show()
}

func namePopUp(i int) { /// name change confirmation
	ncw := fyne.CurrentApp().NewWindow("Confirm")
	ncw.SetIcon(menu.Resource.SmallIcon)
	ncw.Resize(fyne.NewSize(330, 150))
	ncw.SetFixedSize(true)
	var confirm_display = widget.NewLabel("")
	confirm_display.Wrapping = fyne.TextWrapWord

	name := table.Actions.NameEntry.Text

	switch i {
	case 1:
		confirm_display.SetText("0.1 Dero Fee to Change Name\nNew name: " + name + "\n\nConfirm")
	case 2:
		confirm_display.SetText("0.1 Dero Fee to Remove Address from contract\n\nConfirm")
	default:
		confirm_display.SetText("Error")
	}

	cancel_button := widget.NewButton("Cancel", func() {
		ncw.Close()
	})

	confirm_button := widget.NewButton("Confirm", func() {
		switch i {
		case 1:
			rpc.NameChange(PredictControl.Contract, name)
		case 2:
			rpc.RemoveAddress(PredictControl.Contract, name)
		default:

		}

		ncw.Close()
	})

	display := container.NewVBox(confirm_display, layout.NewSpacer())
	options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewVBox(display, layout.NewSpacer(), options)

	img := *canvas.NewImageFromResource(menu.Resource.Back1)
	ncw.SetContent(
		container.New(layout.NewMaxLayout(),
			&img,
			content))
	ncw.Show()
}

func humanTimeConvert() fyne.CanvasObject {
	entry := widget.NewEntry()
	res := widget.NewEntry()
	res.Disable()
	button := widget.NewButton("Human Time", func() {
		time := time.Unix(int64(rpc.StringToInt(entry.Text)), 0).String()
		res.SetText(time)
	})

	split := container.NewHSplit(entry, button)
	box := container.NewVBox(res, split)

	return box
}

func OwnerButton() fyne.CanvasObject {
	menu.MenuControl.Bet_menu = widget.NewButton("Bet Contract Options", func() {
		ownersMenu()
	})
	menu.MenuControl.Bet_menu.Hide()

	box := container.NewVBox(layout.NewSpacer(), menu.MenuControl.Bet_menu)

	return box
}

func GetActiveGames(dc bool) {
	if dc {
		options := []string{}
		contracts := menu.Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
		keys := make([]string, len(contracts))

		i := 0
		for k := range contracts {
			keys[i] = k
			owner, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(keys[i], "owner", menu.Gnomes.Indexer.ChainHeight, true)
			if owner != nil && owner[0] == rpc.Wallet.Address {
				if len(keys[i]) == 64 {
					_, init := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(keys[i], "s_init", menu.Gnomes.Indexer.ChainHeight, true)
					if init != nil {
						for ic := uint64(1); ic <= init[0]; ic++ {
							num := strconv.Itoa(int(ic))
							game, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(keys[i], "game_"+num, menu.Gnomes.Indexer.ChainHeight, true)
							league, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(keys[i], "league_"+num, menu.Gnomes.Indexer.ChainHeight, true)
							_, end := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(keys[i], "s_end_at_"+num, menu.Gnomes.Indexer.ChainHeight, true)
							_, add := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(keys[i], "time_a", menu.Gnomes.Indexer.ChainHeight, true)
							if game != nil && end != nil && add != nil {
								if end[0]+add[0] < uint64(time.Now().Unix()) {
									options = append(options, num+"   "+league[0]+"   "+game[0])
								}
							}
						}

					}
				}
				i++
			}
		}
		PS_Control.Payout_n.SetOptions(options)
	}
}

func ownersMenu() { /// bet owners menu
	ow := fyne.CurrentApp().NewWindow("Bet Contracts")
	ow.Resize(fyne.NewSize(330, 700))
	ow.SetIcon(menu.Resource.SmallIcon)
	menu.MenuControl.Bet_menu.Hide()
	quit := make(chan struct{})
	ow.SetCloseIntercept(func() {
		menu.MenuControl.Bet_menu.Show()
		quit <- struct{}{}
		ow.Close()
	})
	ow.SetFixedSize(true)

	owner_tabs := container.NewAppTabs(
		container.NewTabItem("Predict", preditctionOpts()),
		container.NewTabItem("Sports", sportsOpts()),
		container.NewTabItem("Payout", payoutOpts()),
	)
	owner_tabs.SetTabLocation(container.TabLocationTop)
	owner_tabs.OnSelected = func(ti *container.TabItem) {
		switch ti.Text {
		case "Payout":
			go GetActiveGames(rpc.Signal.Daemon)
		}
	}

	var utime string
	clock := widget.NewEntry()
	clock.Disable()

	entry := widget.NewEntry()
	entry.Validator = validation.NewRegexp(`^\d{1,}$`, "Format Not Valid")
	button := widget.NewButton("Add Hours", func() {
		if entry.Validate() == nil {
			i := rpc.StringToInt(entry.Text)
			u := rpc.StringToInt(utime)
			r := u + (i * 3600)

			switch owner_tabs.SelectedIndex() {
			case 0:
				PS_Control.P_end.SetText(strconv.Itoa(r))
			case 1:
				PS_Control.S_end.SetText(strconv.Itoa(r))
			}
		}
	})

	go func() {
		var ticker = time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				now := time.Now()
				utime = strconv.Itoa(int(now.Unix()))
				clock.SetText("Unix Time: " + utime)
				if now.Unix() < rpc.Predict.Buffer {
					if rpc.Predict.Init {
						PS_Control.P_cancel.Show()
					} else {
						PS_Control.P_cancel.Hide()
					}
				} else {
					PS_Control.P_cancel.Hide()
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	bottom_split := container.NewHSplit(entry, button)
	bottom_box := container.NewVBox(clock, bottom_split)

	border := container.NewBorder(nil, bottom_box, nil, nil, owner_tabs)

	img := *canvas.NewImageFromResource(menu.Resource.Back3)
	ow.SetContent(
		container.New(
			layout.NewMaxLayout(),
			&img,
			border))
	ow.Show()
}

func ownerConfirmPopUp(i int, p float64) { /// bet owner action confirmation
	ocw := fyne.CurrentApp().NewWindow("Confirm")
	ocw.SetIcon(menu.Resource.SmallIcon)
	ocw.Resize(fyne.NewSize(330, 330))
	ocw.SetFixedSize(true)
	var confirm_display = widget.NewLabel("")
	confirm_display.Wrapping = fyne.TextWrapWord

	pre := rpc.Display.Prediction
	p_scid := PredictControl.Contract
	p_pre := PS_Control.P_Name.Text
	p_amt := PS_Control.P_amt.Text
	p_mark := PS_Control.P_mark.Text
	p_end := PS_Control.P_end.Text
	p_end_time, _ := rpc.MsToTime(p_end + "000")
	p_feed := PS_Control.P_feed.Text
	price := fmt.Sprintf("%.2f", p/100)
	p_dep := PS_Control.P_deposit.Text

	var s_game string
	s_scid := SportsControl.Contract
	game_split := strings.Split(PS_Control.S_game.Selected, "   ")
	if len(game_split) > 1 {
		s_game = game_split[1]
	} else {
		s_game = game_split[0]
	}

	s_league := PS_Control.S_league.Text
	s_amt := PS_Control.S_amt.Text
	s_end := PS_Control.S_end.Text
	s_end_time, _ := rpc.MsToTime(s_end + "000")
	s_feed := PS_Control.S_feed.Text
	n_split := strings.Split(PS_Control.Payout_n.Text, "   ")
	s_pay_n := n_split[0]
	s_dep := PS_Control.S_deposit.Text

	var win, team string
	if i == 3 {
		win, team = GetWinner(n_split[2], n_split[1])
	}

	switch i {
	case 1:
		confirm_display.SetText("SCID: " + s_scid + "\n\nGame: " + s_game + "\n\nMinimum: " + s_amt + "\n\nCloses At: " + s_end_time.String() + "\n\nFeed: " + s_feed + "\n\nInitial Deposit: " + s_dep + " Dero")
	case 2:
		fn := "Feed: "
		var mark string
		if p_mark == "0" || p_mark == "" {
			mark = "Not Set"
		} else {
			if onChainPrediction(pre) == 2 || onChainPrediction(p_pre) == 2 { /// decimal of one place for block time
				fn = "Node: "
				i := rpc.StringToInt(p_mark) * 10000
				x := float64(i) / 100000
				mark = fmt.Sprintf("%.5f", x)
			} else {
				if isOnChainPrediction(pre) || isOnChainPrediction(p_pre) {
					fn = "Node: "
					mark = p_mark
				} else {
					if f, err := strconv.ParseFloat(p_mark, 32); err == nil {
						x := f / 100
						mark = fmt.Sprintf("%.2f", x)
					}
				}
			}
		}

		confirm_display.SetText("SCID: " + p_scid + "\n\nPredicting: " + p_pre + "\n\nMinimum: " + p_amt + "\n\nCloses At: " + p_end_time.String() + "\n\nMark: " + mark + "\n\n" + fn + p_feed + "\n\nInitial Deposit: " + p_dep + " Dero")

	case 3:
		confirm_display.SetText("SCID: " + s_scid + "\n\nGame: " + PS_Control.Payout_n.Text + "\nTeam: " + team + "\n\nConfirm")
	case 4:
		confirm_display.SetText("SCID: " + p_scid + "Feed from: dReams Client\n\nPost Price: " + price + "\n\nConfirm")
	case 5:
		confirm_display.SetText("SCID: " + p_scid + "Feed from: dReams Client\n\nFinal Price: " + price + "\n\nConfirm")
	case 6:
		switch onChainPrediction(pre) {
		case 1:
			confirm_display.SetText("SCID: " + p_scid + "\n\n" + pre + ": " + fmt.Sprintf("%.0f", p) + "\n\nNode: " + rpc.Display.P_feed + "\n\nConfirm Post")
		case 2:
			confirm_display.SetText("SCID: " + p_scid + "\n\n" + pre + ": " + fmt.Sprintf("%.5f", p) + "\n\nNode: " + rpc.Display.P_feed + "\n\nConfirm Post")
		case 3:
			confirm_display.SetText("SCID: " + p_scid + "\n\n" + pre + ": " + fmt.Sprintf("%.0f", p) + "\n\nNode: " + rpc.Display.P_feed + "\n\nConfirm Post")
		}

	case 7:
		switch onChainPrediction(pre) {
		case 1:
			confirm_display.SetText("SCID: " + p_scid + "\n\n" + pre + ": " + fmt.Sprintf("%.0f", p) + "\n\nNode: " + rpc.Display.P_feed + "\n\nConfirm Payout")
		case 2:
			confirm_display.SetText("SCID: " + p_scid + "\n\n" + pre + ": " + fmt.Sprintf("%.5f", p) + "\n\nNode: " + rpc.Display.P_feed + "\n\nConfirm Payout")
		case 3:
			confirm_display.SetText("SCID: " + p_scid + "\n\n" + pre + ": " + fmt.Sprintf("%.0f", p) + "\n\nNode: " + rpc.Display.P_feed + "\n\nConfirm Payout")
		}

	case 8:
		confirm_display.SetText("SCID: " + p_scid + "\n\nThis will Cancel the current prediction")
	default:
		log.Println("No Confirm Input")
		confirm_display.SetText("Error")
	}

	cancel_button := widget.NewButton("Cancel", func() {
		ocw.Close()
	})

	confirm_button := widget.NewButton("Confirm", func() {
		PS_Control.Payout_n.SetText("")
		switch i {
		case 1:
			rpc.SetSports(rpc.StringToInt(s_end), menu.ToAtomicFive(s_amt), menu.ToAtomicFive(s_dep), s_scid, s_league, s_game, s_feed)
		case 2:
			if onChainPrediction(pre) == 2 || onChainPrediction(p_pre) == 2 { /// decimal of one place for block time
				rpc.SetPrediction(rpc.StringToInt(p_end), rpc.StringToInt(p_mark)*10000, menu.ToAtomicFive(p_amt), menu.ToAtomicFive(p_dep), p_scid, p_pre, p_feed)
			} else {
				rpc.SetPrediction(rpc.StringToInt(p_end), rpc.StringToInt(p_mark), menu.ToAtomicFive(p_amt), menu.ToAtomicFive(p_dep), p_scid, p_pre, p_feed)
			}
		case 3:
			rpc.EndSports(s_scid, s_pay_n, win)
		case 4:
			rpc.PostPrediction(p_scid, int(p))
		case 5:
			rpc.EndPredition(p_scid, int(p))
		case 6:
			switch onChainPrediction(pre) {
			case 1:
				rpc.PostPrediction(p_scid, int(p))
			case 2:
				rpc.PostPrediction(p_scid, int(p*100000))
			case 3:
				rpc.PostPrediction(p_scid, int(p))
			default:
			}
		case 7:
			switch onChainPrediction(pre) {
			case 1:
				rpc.EndPredition(p_scid, int(p))
			case 2:
				rpc.EndPredition(p_scid, int(p*100000))
			case 3:
				rpc.EndPredition(p_scid, int(p))
			default:
			}
		case 8:
			rpc.CancelPrediction(PredictControl.Contract)
		default:

		}
		ocw.Close()
	})

	display := container.NewVBox(confirm_display, layout.NewSpacer())
	options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewVBox(display, layout.NewSpacer(), options)

	img := *canvas.NewImageFromResource(menu.Resource.Back2)
	ocw.SetContent(
		container.New(layout.NewMaxLayout(),
			&img,
			content))
	ocw.Show()
}
