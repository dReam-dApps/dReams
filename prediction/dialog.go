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
	S_set        *widget.Button
	S_cancel     *widget.Button
	P_end        *widget.Entry
	P_mark       *widget.Entry
	P_amt        *table.NumericalEntry
	P_Name       *widget.SelectEntry
	P_feed       *widget.SelectEntry
	P_deposit    *table.NumericalEntry
	P_set        *widget.Button
	P_post       *widget.Button
	P_pay        *widget.Button
	P_cancel     *widget.Button
	Payout_n     *widget.SelectEntry
	Owner_button *widget.Button
	Run_service  *widget.Button
	Service_pay  *widget.Check
	Transactions *widget.Check
	Payout_on    bool
	Transact_on  bool
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
	pred := []string{"DERO-BTC", "XMR-BTC", "BTC-USDT", "DERO-USDT", "XMR-USDT", "DERO-Difficulty", "DERO-Block Time", "DERO-Block Number"}
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
	PS_Control.P_amt.Wrapping = fyne.TextTruncate
	PS_Control.P_amt.Validator = validation.NewRegexp(`^\d{1,}\.\d{1,5}$`, "Format Not Valid")

	feeds := []string{"dReams Client"}
	PS_Control.P_feed = widget.NewSelectEntry(feeds)
	PS_Control.P_feed.SetPlaceHolder("Feed:")

	PS_Control.P_deposit = table.NilNumericalEntry()
	PS_Control.P_deposit.SetPlaceHolder("Deposit Amount:")
	PS_Control.P_deposit.Wrapping = fyne.TextTruncate
	PS_Control.P_deposit.Validator = validation.NewRegexp(`^\d{1,}\.\d{1,5}$`, "Format Not Valid")

	PS_Control.P_set = widget.NewButton("Set Prediction", func() {
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

	PS_Control.P_post = widget.NewButton("Post", func() {
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

	PS_Control.P_post.Hide()

	PS_Control.P_pay = widget.NewButton("Prediction Payout", func() {
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

	PS_Control.P_pay.Hide()

	owner_p := container.NewVBox(
		humanTimeConvert(),
		layout.NewSpacer(),
		PS_Control.P_Name,
		PS_Control.P_end,
		PS_Control.P_mark,
		PS_Control.P_amt,
		PS_Control.P_feed,
		PS_Control.P_deposit,
		PS_Control.P_set,
		layout.NewSpacer(),
		PS_Control.P_cancel,
		layout.NewSpacer(),
		PS_Control.P_post,
		layout.NewSpacer(),
		PS_Control.P_pay,
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

		PS_Control.S_feed.SetText("dReams Client")
		PS_Control.S_feed.Refresh()
	}
	PS_Control.S_league.SetPlaceHolder("League:")

	PS_Control.S_end = widget.NewEntry()
	PS_Control.S_end.SetPlaceHolder("Closes At:")
	PS_Control.S_end.Validator = validation.NewRegexp(`^\d{10,}$`, "Format Not Valid")

	PS_Control.S_amt = table.NilNumericalEntry()
	PS_Control.S_amt.SetPlaceHolder("Minimum Amount:")
	PS_Control.S_amt.Wrapping = fyne.TextTruncate
	PS_Control.S_amt.Validator = validation.NewRegexp(`^\d{1,}\.\d{1,5}$`, "Format Not Valid")

	feeds := []string{"dReams Client"}
	PS_Control.S_feed = widget.NewSelectEntry(feeds)
	PS_Control.S_feed.SetPlaceHolder("Feed:")

	PS_Control.S_deposit = table.NilNumericalEntry()
	PS_Control.S_deposit.SetPlaceHolder("Deposit Amount:")
	PS_Control.S_deposit.Wrapping = fyne.TextTruncate
	PS_Control.S_deposit.Validator = validation.NewRegexp(`^\d{1,}\.\d{1,5}$`, "Format Not Valid")

	PS_Control.S_set = widget.NewButton("Set Game", func() {
		if PS_Control.S_deposit.Validate() == nil && PS_Control.S_amt.Validate() == nil && PS_Control.S_end.Validate() == nil {
			if len(SportsControl.Contract) == 64 {
				ownerConfirmPopUp(1, 100)
			}
		}
	})

	PS_Control.S_cancel = widget.NewButton("Cancel", func() {
		ownerConfirmPopUp(9, 0)
	})

	PS_Control.S_cancel.Hide()

	PS_Control.Payout_n = widget.NewSelectEntry([]string{})
	PS_Control.Payout_n.SetPlaceHolder("Game #")

	sports_confirm := widget.NewButton("Sports Payout", func() {
		if len(SportsControl.Contract) == 64 {
			ownerConfirmPopUp(3, 100)
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
		PS_Control.S_set,
		layout.NewSpacer(),
		PS_Control.S_cancel,
		layout.NewSpacer(),
		PS_Control.Payout_n,
		sports_confirm,
		layout.NewSpacer())

	return sports
}

func serviceOpts() fyne.CanvasObject {
	get_addr := widget.NewButton("Integrated Address", func() {
		go makeIntegratedAddr(true)
	})

	txid := widget.NewMultiLineEntry()
	txid.SetPlaceHolder("TXID:")
	txid.Wrapping = fyne.TextWrapWord
	txid.Validator = validation.NewRegexp(`^\w{64,64}$`, "Format Not Valid")

	process := widget.NewButton("Process Tx", func() {
		if !Service.Processing && !rpc.Wallet.Service {
			if txid.Validate() == nil {
				processSingleTx(txid.Text)
			}
		} else {
			log.Println("[dReams] Stop service to manually process Tx")
		}
	})

	delete := widget.NewButton("Delete Tx", func() {
		if !Service.Processing && !rpc.Wallet.Service {
			if txid.Validate() == nil {
				e, _ := rpc.GetTransaction(txid.Text)
				if e != nil {
					deleteTx("BET", e)
				}
			}
		} else {
			log.Println("[dReams] Stop service to delete Tx")
		}
	})

	entry := &table.NumericalEntry{}
	entry.ExtendBaseWidget(entry)
	entry.SetPlaceHolder("Block #:")
	entry.Wrapping = fyne.TextTruncate
	entry.Validator = validation.NewRegexp(`^[^0]\d{0,}$`, "Format Not Valid")

	var start uint64
	height := widget.NewCheck("Start from current height", func(b bool) {
		if b {
			start, _ = rpc.DaemonHeight(rpc.Round.Daemon)
			entry.SetText(strconv.Itoa(int(start)))
			entry.Disable()
		} else {
			entry.SetText("")
			entry.Enable()
		}
	})
	height.SetChecked(true)

	debug := widget.NewCheck("Debug", func(b bool) {
		if b {
			Service.Debug = true
		} else {
			Service.Debug = false
		}
	})

	view := widget.NewButton("View Tx History", func() {
		if !Service.Processing && !rpc.Wallet.Service {
			if !height.Checked {
				start = uint64(rpc.StringToInt(entry.Text))
			}
			viewProcessedTx(start)
		} else {
			log.Println("[dReams] Stop service to view Tx history")
		}
	})

	PS_Control.Service_pay = widget.NewCheck("Payouts", func(b bool) {
		if b {
			PS_Control.Payout_on = true
		} else {
			PS_Control.Payout_on = false
		}
	})

	if PS_Control.Payout_on {
		PS_Control.Service_pay.SetChecked(true)
		PS_Control.Service_pay.Disable()
	}

	PS_Control.Transactions = widget.NewCheck("Transactions", func(b bool) {
		if b {
			PS_Control.Transact_on = true
		} else {
			PS_Control.Transact_on = false
		}
	})

	if PS_Control.Transact_on {
		PS_Control.Transactions.SetChecked(true)
		PS_Control.Transactions.Disable()
	}

	PS_Control.Run_service = widget.NewButton("Run Service", func() {
		if !rpc.Wallet.Service {
			if entry.Validate() == nil {
				if !height.Checked {
					start = uint64(rpc.StringToInt(entry.Text))
					if start < PAYLOAD_FORMAT {
						start = PAYLOAD_FORMAT
					}
				}
				if PS_Control.Service_pay.Checked || PS_Control.Transactions.Checked {
					rpc.Wallet.Service = true
					PS_Control.Run_service.Hide()
					go servicePopUp(start, PS_Control.Service_pay.Checked, PS_Control.Transactions.Checked)
				} else {
					log.Println("[dReams] Select which services to run")
				}
			} else {
				log.Println("[dReams] Enter service starting height")
			}
		} else {
			log.Println("[dReams] Service already running")
		}
	})

	if rpc.Wallet.Service || Service.Processing {
		PS_Control.Run_service.Hide()
	}

	stop := widget.NewButton("Stop Service", func() {
		if rpc.Wallet.Service {
			log.Println("[dReams] Stopping service")
		}
		rpc.Wallet.Service = false

	})

	box := container.NewVBox(
		layout.NewSpacer(),
		view,
		layout.NewSpacer(),
		txid,
		container.NewAdaptiveGrid(2,
			process,
			delete),
		layout.NewSpacer(),
		get_addr,
		layout.NewSpacer(),
		height,
		entry,
		PS_Control.Service_pay,
		PS_Control.Transactions,
		debug,
		container.NewAdaptiveGrid(2,
			stop,
			PS_Control.Run_service,
		))

	return box
}

func updateOpts() fyne.CanvasObject {
	a_label := widget.NewLabel("Time A         ")
	a := &table.NumericalEntry{}
	a.ExtendBaseWidget(a)
	a.PlaceHolder = "Time A:"
	a.Wrapping = fyne.TextTruncate
	a.Validator = validation.NewRegexp(`[^0]\d{1,}$`, "Format Not Valid")

	b_label := widget.NewLabel("Time B         ")
	b := &table.NumericalEntry{}
	b.ExtendBaseWidget(b)
	b.PlaceHolder = "Time B:"
	b.Wrapping = fyne.TextTruncate
	b.Validator = validation.NewRegexp(`[^0]\d{1,}$`, "Format Not Valid")

	c_label := widget.NewLabel("Time C         ")
	c := &table.NumericalEntry{}
	c.ExtendBaseWidget(c)
	c.PlaceHolder = "Time C:"
	c.Wrapping = fyne.TextTruncate
	c.Validator = validation.NewRegexp(`[^0]\d{1,}$`, "Format Not Valid")

	hl_label := widget.NewLabel("Max Games")
	hl := &table.NumericalEntry{}
	hl.ExtendBaseWidget(hl)
	hl.PlaceHolder = "Max Games:"
	hl.Wrapping = fyne.TextTruncate
	hl.Validator = validation.NewRegexp(`^[^0]\d{0,}$`, "Format Not Valid")

	hl_box := container.NewBorder(nil, nil, hl_label, nil, hl)
	hl_box.Hide()

	// l := &table.NumericalEntry{}
	// l.ExtendBaseWidget(l)
	// l.PlaceHolder = "L:"
	// l.Validator = validation.NewRegexp(`^\d{2,}$`, "Format Not Valid")

	sc := widget.NewSelect([]string{"Prediction", "Sports"}, func(s string) {
		if s == "Sports" {
			c_label.SetText("Delete         ")
			c.Validator = validation.NewRegexp(`[^0]\d{0,}$`, "Format Not Valid")
			hl_box.Show()
		} else {
			c_label.SetText("Time C         ")
			c.Validator = validation.NewRegexp(`[^0]\d{1,}$`, "Format Not Valid")
			hl_box.Hide()
		}
	})
	sc.PlaceHolder = "Select Contract"

	new_owner := widget.NewMultiLineEntry()
	new_owner.Wrapping = fyne.TextWrapWord
	new_owner.SetPlaceHolder("New owner address:")
	add_owner := widget.NewButton("Add Owner", func() {
		if len(new_owner.Text) == 66 {
			switch sc.Selected {
			case "Prediction":
				rpc.AddOwner(PredictControl.Contract, new_owner.Text)
			case "Sports":
				rpc.AddOwner(SportsControl.Contract, new_owner.Text)
			default:
				log.Println("[dReams] Select contract")
			}
		}
	})

	owner_num := &table.NumericalEntry{}
	owner_num.ExtendBaseWidget(owner_num)
	owner_num.PlaceHolder = "Owner #:"
	owner_num.Validator = validation.NewRegexp(`^[^0]\d{0,0}$`, "Format Not Valid")
	owner_num.Wrapping = fyne.TextTruncate

	remove_owner := widget.NewButton("Remove Owner", func() {
		switch sc.Selected {
		case "Prediction":
			rpc.RemoveOwner(PredictControl.Contract, rpc.StringToInt(owner_num.Text))
		case "Sports":
			rpc.RemoveOwner(SportsControl.Contract, rpc.StringToInt(owner_num.Text))
		default:
			log.Println("[dReams] Select contract")
		}
	})

	update_var := widget.NewButton("Update Variables", func() {
		if a.Validate() == nil && b.Validate() == nil && c.Validate() == nil {
			switch sc.Selected {
			case "Prediction":
				rpc.VarUpdate(PredictControl.Contract, rpc.StringToInt(a.Text), rpc.StringToInt(b.Text), rpc.StringToInt(c.Text), 30, 0)
			case "Sports":
				if hl.Validate() == nil {
					rpc.VarUpdate(SportsControl.Contract, rpc.StringToInt(a.Text), rpc.StringToInt(b.Text), rpc.StringToInt(c.Text), 30, rpc.StringToInt(hl.Text))
				}
			default:
				log.Println("[dReams] Select contract")
			}
		}
	})

	return container.NewVBox(
		sc,
		container.NewBorder(nil, nil, a_label, nil, a),
		container.NewBorder(nil, nil, b_label, nil, b),
		container.NewBorder(nil, nil, c_label, nil, c),
		hl_box,
		update_var,
		layout.NewSpacer(),
		new_owner,
		add_owner,
		layout.NewSpacer(),
		container.NewBorder(nil, nil, nil, remove_owner, owner_num),
		layout.NewSpacer())

}

func confirmPopUp(i int, teamA, teamB string) { /// bet action confirmation
	ocw := fyne.CurrentApp().NewWindow("Confirm")
	ocw.SetIcon(menu.Resource.SmallIcon)
	ocw.Resize(fyne.NewSize(330, 330))
	ocw.SetFixedSize(true)
	var confirm_display = widget.NewLabel("")
	confirm_display.Wrapping = fyne.TextWrapWord

	p_scid := PredictControl.Contract
	// prediction leaderboard
	//name := table.Actions.NameEntry.Text

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
		log.Println("[dReams] No Confirm Input")
		confirm_display.SetText("Error")
	}

	cancel_button := widget.NewButton("No", func() {
		ocw.Close()
	})
	confirm_button := widget.NewButton("Yes", func() {
		switch i {
		case 1:
			rpc.PredictLower(p_scid, "")
		case 2:
			rpc.PredictHigher(p_scid, "")
		case 3:
			rpc.PickTeam(s_scid, multi, split[0], menu.GetSportsAmt(s_scid, split[0]), 0)
		case 4:
			rpc.PickTeam(s_scid, multi, split[0], menu.GetSportsAmt(s_scid, split[0]), 1)
		default:

		}
		ocw.Close()
	})

	display := container.NewVScroll(confirm_display)
	options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewBorder(nil, options, nil, nil, display)

	img := *canvas.NewImageFromResource(menu.Resource.Back2)
	ocw.SetContent(
		container.New(layout.NewMaxLayout(),
			&img,
			content))
	ocw.Show()
}

// prediction leaderboard
// func namePopUp(i int) { /// name change confirmation
// 	ncw := fyne.CurrentApp().NewWindow("Confirm")
// 	ncw.SetIcon(menu.Resource.SmallIcon)
// 	ncw.Resize(fyne.NewSize(330, 150))
// 	ncw.SetFixedSize(true)
// 	var confirm_display = widget.NewLabel("")
// 	confirm_display.Wrapping = fyne.TextWrapWord
//
// 	name := table.Actions.NameEntry.Text
//
// 	switch i {
// 	case 1:
// 		confirm_display.SetText("0.1 Dero Fee to Change Name\nNew name: " + name + "\n\nConfirm")
// 	case 2:
// 		confirm_display.SetText("0.1 Dero Fee to Remove Address from contract\n\nConfirm")
// 	default:
// 		confirm_display.SetText("Error")
// 	}
//
// 	cancel_button := widget.NewButton("Cancel", func() {
// 		ncw.Close()
// 	})
//
// 	confirm_button := widget.NewButton("Confirm", func() {
// 		switch i {
// 		case 1:
// 			rpc.NameChange(PredictControl.Contract, name)
// 		case 2:
// 			rpc.RemoveAddress(PredictControl.Contract, name)
// 		default:
//
// 		}
//
// 		ncw.Close()
// 	})
//
// 	display := container.NewVScroll(confirm_display)
// 	options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
// 	content := container.NewBorder(nil, options, nil, nil, display)
//
// 	img := *canvas.NewImageFromResource(menu.Resource.Back1)
// 	ncw.SetContent(
// 		container.New(layout.NewMaxLayout(),
// 			&img,
// 			content))
// 	ncw.Show()
// }

func servicePopUp(start uint64, payout, tranfsers bool) { /// service start confirmation
	scw := fyne.CurrentApp().NewWindow("Starting dReamService")
	scw.Resize(fyne.NewSize(330, 150))
	scw.SetIcon(menu.Resource.SmallIcon)
	scw.SetFixedSize(true)
	scw.SetCloseIntercept(func() {
		rpc.Wallet.Service = false
		scw.Close()
	})

	var pay, transac string
	if tranfsers {
		transac = "process transactions sent to your integrated address"
		if payout {
			transac = transac + " "
		}
	}

	if payout {
		if tranfsers {
			pay = "and "
		}
		pay = pay + "process payouts to contracts"
	}

	str := fmt.Sprintf("This will automatically %s%s.\n\nStarting service from height %d", transac, pay, start)
	confirm_display := widget.NewLabel(str)
	confirm_display.Wrapping = fyne.TextWrapWord

	cancel_button := widget.NewButton("Cancel", func() {
		rpc.Wallet.Service = false
		scw.Close()
	})

	confirm_button := widget.NewButton("Confirm", func() {
		go dReamService(start, payout, tranfsers)
		scw.Close()
	})

	go func() {
		for rpc.Wallet.Connect {
			time.Sleep(1 * time.Second)
		}
		scw.Close()
	}()

	display := container.NewVScroll(confirm_display)
	options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewBorder(nil, options, nil, nil, display)

	img := *canvas.NewImageFromResource(menu.Resource.Back1)
	scw.SetContent(
		container.New(layout.NewMaxLayout(),
			&img,
			content))
	scw.Show()
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

func CheckPredictionStatus(dc bool) {
	if dc && menu.Gnomes.Init && !menu.GnomonClosing() {
		_, ends := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(PredictControl.Contract, "p_end_at", menu.Gnomes.Indexer.ChainHeight, true)
		_, time_a := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(PredictControl.Contract, "time_a", menu.Gnomes.Indexer.ChainHeight, true)
		_, time_c := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(PredictControl.Contract, "time_c", menu.Gnomes.Indexer.ChainHeight, true)
		_, mark := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(PredictControl.Contract, "mark", menu.Gnomes.Indexer.ChainHeight, true)
		if ends != nil && time_a != nil && time_c != nil {
			now := uint64(time.Now().Unix())
			if now >= ends[0] && now <= ends[0]+time_a[0] && mark == nil {
				PS_Control.P_post.Show()
			} else {
				PS_Control.P_post.Hide()
			}

			if now >= ends[0]+time_c[0] {
				PS_Control.P_pay.Show()
			} else {
				PS_Control.P_pay.Hide()
			}
		}

		if ends == nil {
			PS_Control.P_post.Hide()
			PS_Control.P_pay.Hide()
		}
	}
}

func GetActiveGames(dc bool) {
	if dc && menu.Gnomes.Init && !menu.GnomonClosing() {
		options := []string{}
		contracts := menu.Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
		keys := make([]string, len(contracts))

		i := 0
		for k := range contracts {
			keys[i] = k
			owner, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(keys[i], "owner", menu.Gnomes.Indexer.ChainHeight, true)
			if (owner != nil && owner[0] == rpc.Wallet.Address) || menu.VerifyBetSigner(keys[i]) {
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
		container.NewTabItem("Service", serviceOpts()),
		container.NewTabItem("Update", updateOpts()),
	)
	owner_tabs.SetTabLocation(container.TabLocationTop)
	owner_tabs.OnSelected = func(ti *container.TabItem) {
		switch ti.Text {
		case "Sports":
			go GetActiveGames(rpc.Signal.Daemon)
		case "Service":
			go makeIntegratedAddr(false)
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
				if !rpc.Wallet.Connect {
					ticker.Stop()
					ow.Close()
				}

				if rpc.Wallet.Service {
					PS_Control.Run_service.Hide()
					PS_Control.Service_pay.Disable()
					PS_Control.Transactions.Disable()
				}

				if !rpc.Wallet.Service && !Service.Processing {
					PS_Control.Run_service.Show()
					PS_Control.Service_pay.Enable()
					PS_Control.Transactions.Enable()
				}

				CheckPredictionStatus(rpc.Signal.Daemon)
				now := time.Now()
				utime = strconv.Itoa(int(now.Unix()))
				clock.SetText("Unix Time: " + utime)
				if now.Unix() < rpc.Predict.Buffer {
					if rpc.Predict.Init {
						PS_Control.P_set.Hide()
						PS_Control.P_cancel.Show()
					} else {
						PS_Control.P_set.Show()
						PS_Control.P_cancel.Hide()
					}
				} else {
					PS_Control.P_cancel.Hide()
					if rpc.Predict.Init {
						PS_Control.P_set.Hide()
					} else {
						PS_Control.P_set.Show()
					}
				}

				if SportsControl.Buffer {
					PS_Control.S_cancel.Show()
					PS_Control.S_set.Hide()
				} else {
					PS_Control.S_cancel.Hide()
					PS_Control.S_set.Show()
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
	p_dep := PS_Control.P_deposit.Text
	var price string
	if table.CoinDecimal(pre) == 8 {
		price = fmt.Sprintf("%.8f", p/100000000)
	} else {
		price = fmt.Sprintf("%.2f", p/100)
	}

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
		if len(n_split) < 3 {
			log.Println("[dReams] Could not format game string")
			return
		}
		if n_split[1] == "Bellator" || n_split[1] == "UFC" {
			win, team = GetMmaWinner(n_split[2], n_split[1])
		} else {
			win, team = GetWinner(n_split[2], n_split[1])
		}
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
			if onChainPrediction(pre) == 2 || onChainPrediction(p_pre) == 2 { /// one decimal place for block time
				fn = "Node: "
				i := rpc.StringToInt(p_mark) * 10000
				x := float64(i) / 100000
				mark = fmt.Sprintf("%.5f", x)
			} else {
				if isOnChainPrediction(pre) || isOnChainPrediction(p_pre) {
					fn = "Node: "
					mark = p_mark
				} else {
					if table.CoinDecimal(pre) == 8 || table.CoinDecimal(p_pre) == 8 {
						if f, err := strconv.ParseFloat(p_mark, 32); err == nil { /// eight decimal place for btc
							x := f / 100000000
							mark = fmt.Sprintf("%.8f", x)
						}
					} else {
						if f, err := strconv.ParseFloat(p_mark, 32); err == nil {
							x := f / 100
							mark = fmt.Sprintf("%.2f", x)
						}
					}
				}
			}
		}

		confirm_display.SetText("SCID: " + p_scid + "\n\nPredicting: " + p_pre + "\n\nMinimum: " + p_amt + "\n\nCloses At: " + p_end_time.String() + "\n\nMark: " + mark + "\n\n" + fn + p_feed + "\n\nInitial Deposit: " + p_dep + " Dero")

	case 3:
		confirm_display.SetText("SCID: " + s_scid + "\n\nGame: " + PS_Control.Payout_n.Text + "\n\nTeam: " + team + "\n\nConfirm")
	case 4:
		confirm_display.SetText("SCID: " + p_scid + "\n\nFeed from: dReams Client\n\nPost Price: " + price + "\n\nConfirm")
	case 5:
		confirm_display.SetText("SCID: " + p_scid + "\n\nFeed from: dReams Client\n\nFinal Price: " + price + "\n\nConfirm")
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
		confirm_display.SetText("SCID:\n" + p_scid + "\n\nThis will Cancel the current prediction")
	case 9:
		confirm_display.SetText("SCID:\n" + s_scid + "\n\nThis will Cancel the last initiated bet on this contract")
	default:
		log.Println("[dReams] No Confirm Input")
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
			rpc.EndPrediction(p_scid, int(p))
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
				rpc.EndPrediction(p_scid, int(p))
			case 2:
				rpc.EndPrediction(p_scid, int(p*100000))
			case 3:
				rpc.EndPrediction(p_scid, int(p))
			default:
			}
		case 8:
			rpc.CancelInitiatedBet(PredictControl.Contract, 0)
		case 9:
			rpc.CancelInitiatedBet(SportsControl.Contract, 1)
		default:

		}
		ocw.Close()
	})

	display := container.NewVScroll(confirm_display)
	options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewBorder(nil, options, nil, nil, display)

	img := *canvas.NewImageFromResource(menu.Resource.Back2)
	ocw.SetContent(
		container.New(layout.NewMaxLayout(),
			&img,
			content))
	ocw.Show()
}
