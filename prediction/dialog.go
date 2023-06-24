package prediction

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"strconv"
	"strings"
	"time"

	holdero "github.com/SixofClubsss/Holdero"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/menu"
	"github.com/dReam-dApps/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type ownerObjects struct {
	S_end        *dwidget.DeroAmts
	S_amt        *dwidget.DeroAmts
	S_game       *widget.Select
	S_league     *widget.SelectEntry
	S_feed       *widget.SelectEntry
	S_deposit    *dwidget.DeroAmts
	S_set        *widget.Button
	S_cancel     *widget.Button
	P_end        *dwidget.DeroAmts
	P_mark       *widget.Entry
	P_amt        *dwidget.DeroAmts
	P_Name       *widget.SelectEntry
	P_feed       *widget.SelectEntry
	P_deposit    *dwidget.DeroAmts
	P_set        *widget.Button
	P_post       *widget.Button
	P_pay        *widget.Button
	P_cancel     *widget.Button
	Payout_n     *widget.SelectEntry
	Owner_button *widget.Button
	Run_service  *widget.Button
	Service_pay  *widget.Check
	Transactions *widget.Check
	Synced       bool
	Payout_on    bool
	Transact_on  bool
}

var Owner ownerObjects

// Check if prediction is for on chain values
func isOnChainPrediction(s string) bool {
	if s == "DERO-Difficulty" || s == "DERO-Block Time" || s == "DERO-Block Number" {
		return true
	}

	return false
}

// Check which on chain values are required
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

// dPrediction owner control objects for side menu
//   - Pass side menu window to reset to
func predictionOpts(window fyne.Window) fyne.CanvasObject {
	pred := []string{"DERO-BTC", "XMR-BTC", "BTC-USDT", "DERO-USDT", "XMR-USDT", "DERO-Difficulty", "DERO-Block Time", "DERO-Block Number"}
	Owner.P_Name = widget.NewSelectEntry(pred)
	Owner.P_Name.SetPlaceHolder("Name:")
	Owner.P_Name.OnChanged = func(s string) {
		if isOnChainPrediction(s) {
			opts := []string{rpc.DAEMON_RPC_REMOTE1, rpc.DAEMON_RPC_REMOTE2, rpc.DAEMON_RPC_REMOTE5, rpc.DAEMON_RPC_REMOTE6}
			Owner.P_feed.SetOptions(opts)
			if Owner.P_feed.Text != opts[1] {
				Owner.P_feed.SetText(opts[0])
			}
			Owner.P_feed.SetPlaceHolder("Node:")
			Owner.P_feed.Refresh()
		} else {
			opts := []string{"dReams Client"}
			Owner.P_feed.SetOptions(opts)
			Owner.P_feed.SetText(opts[0])
			Owner.P_feed.SetPlaceHolder("Feed:")
			Owner.P_feed.Refresh()
		}
	}

	Owner.P_end = dwidget.DeroAmtEntry("", 1, 0)
	Owner.P_end.SetPlaceHolder("Closes At:")
	Owner.P_end.AllowFloat = false
	Owner.P_end.Validator = validation.NewRegexp(`^\d{10,}$`, "Unix time required")

	Owner.P_mark = widget.NewEntry()
	Owner.P_mark.SetPlaceHolder("Mark:")
	Owner.P_mark.Validator = validation.NewRegexp(`^\d{1,}$`, "Int required")

	Owner.P_amt = dwidget.DeroAmtEntry("", 0.1, 1)
	Owner.P_amt.SetPlaceHolder("Minimum Amount:")
	Owner.P_amt.AllowFloat = true
	Owner.P_amt.Wrapping = fyne.TextTruncate
	Owner.P_amt.Validator = validation.NewRegexp(`^\d{1,}\.\d{1,5}$|^[^0]\d{0,}$`, "Int or float required")

	feeds := []string{"dReams Client"}
	Owner.P_feed = widget.NewSelectEntry(feeds)
	Owner.P_feed.SetPlaceHolder("Feed:")

	Owner.P_deposit = dwidget.DeroAmtEntry("", 0.1, 1)
	Owner.P_deposit.SetPlaceHolder("Deposit Amount:")
	Owner.P_deposit.AllowFloat = true
	Owner.P_deposit.Wrapping = fyne.TextTruncate
	Owner.P_deposit.Validator = validation.NewRegexp(`^\d{1,}\.\d{1,5}$|^[^0]\d{0,}$`, "Int or float required")

	reset := window.Content().(*fyne.Container).Objects[2]

	Owner.P_set = widget.NewButton("Set Prediction", func() {
		if Owner.P_deposit.Validate() == nil && Owner.P_amt.Validate() == nil && Owner.P_end.Validate() == nil && Owner.P_mark.Validate() == nil {
			if len(Predict.Contract) == 64 {
				window.Content().(*fyne.Container).Objects[2] = ownerConfirmAction(2, 100, window, reset)
				window.Content().(*fyne.Container).Objects[2].Refresh()
			}
		}
	})

	Owner.P_cancel = widget.NewButton("Cancel", func() {
		window.Content().(*fyne.Container).Objects[2] = ownerConfirmAction(8, 0, window, reset)
		window.Content().(*fyne.Container).Objects[2].Refresh()
	})

	Owner.P_cancel.Hide()

	Owner.P_post = widget.NewButton("Post", func() {
		go SetPredictionPrices(rpc.Daemon.Connect)
		var a float64
		prediction := Predict.Prediction
		if isOnChainPrediction(prediction) {
			switch onChainPrediction(prediction) {
			case 1:
				a = rpc.GetDifficulty(Predict.Feed)
				window.Content().(*fyne.Container).Objects[2] = ownerConfirmAction(6, a, window, reset)
			case 2:
				a = rpc.GetBlockTime(Predict.Feed)
				window.Content().(*fyne.Container).Objects[2] = ownerConfirmAction(6, a, window, reset)
			case 3:
				d := rpc.DaemonHeight("dReams", Predict.Feed)
				a = float64(d)
				window.Content().(*fyne.Container).Objects[2] = ownerConfirmAction(6, a, window, reset)
			default:

			}

		} else {
			a, _ = holdero.GetPrice(prediction)
			window.Content().(*fyne.Container).Objects[2] = ownerConfirmAction(4, a, window, reset)
		}

		window.Content().(*fyne.Container).Objects[2].Refresh()

	})

	Owner.P_post.Hide()

	Owner.P_pay = widget.NewButton("Prediction Payout", func() {
		go SetPredictionPrices(rpc.Daemon.Connect)
		var a float64
		prediction := Predict.Prediction
		if isOnChainPrediction(prediction) {
			switch onChainPrediction(prediction) {
			case 1:
				a = rpc.GetDifficulty(Predict.Feed)
				window.Content().(*fyne.Container).Objects[2] = ownerConfirmAction(7, a, window, reset)
			case 2:
				a = rpc.GetBlockTime(Predict.Feed)
				window.Content().(*fyne.Container).Objects[2] = ownerConfirmAction(7, a, window, reset)
			case 3:
				d := rpc.DaemonHeight("dReams", Predict.Feed)
				a = float64(d)
				window.Content().(*fyne.Container).Objects[2] = ownerConfirmAction(7, a, window, reset)
			default:

			}

		} else {
			a, _ = holdero.GetPrice(prediction)
			window.Content().(*fyne.Container).Objects[2] = ownerConfirmAction(5, a, window, reset)
		}

		window.Content().(*fyne.Container).Objects[2].Refresh()
	})

	Owner.P_pay.Hide()

	owner_p := container.NewVBox(
		humanTimeConvert(),
		layout.NewSpacer(),
		Owner.P_Name,
		Owner.P_end,
		Owner.P_mark,
		Owner.P_amt,
		Owner.P_feed,
		Owner.P_deposit,
		Owner.P_set,
		layout.NewSpacer(),
		Owner.P_cancel,
		layout.NewSpacer(),
		Owner.P_post,
		layout.NewSpacer(),
		Owner.P_pay,
		layout.NewSpacer(),
	)

	return owner_p
}

// dSports owner control objects for side menu
//   - Pass side menu window to reset to
func sportsOpts(window fyne.Window) fyne.CanvasObject {
	options := []string{}
	Owner.S_game = widget.NewSelect(options, func(s string) {
		var date string
		game := strings.Split(s, "   ")
		for i := range s {
			if i > 3 {
				date = s[0:10]
			}
		}
		comp := date[0:4] + date[5:7] + date[8:10]
		GetGameEnd(comp, game[1], Owner.S_league.Text)
	})
	Owner.S_game.PlaceHolder = "Game:"

	leagues := []string{"EPL", "MLS", "NBA", "NFL", "NHL", "MLB", "Bellator", "UFC"}
	Owner.S_league = widget.NewSelectEntry(leagues)
	Owner.S_league.OnChanged = func(s string) {
		Owner.S_game.Options = []string{}
		Owner.S_game.Selected = ""
		if s == "Bellator" || s == "UFC" {
			Owner.S_game.PlaceHolder = "Fight:"
		} else {
			Owner.S_game.PlaceHolder = "Game:"
		}
		Owner.S_game.Refresh()
		switch s {
		case "EPL":
			go GetCurrentWeek("EPL")
		case "MLS":
			go GetCurrentWeek("MLS")
		case "NBA":
			go GetCurrentWeek("NBA")
		case "NFL":
			go GetCurrentWeek("NFL")
		case "NHL":
			go GetCurrentWeek("NHL")
		case "MLB":
			go GetCurrentWeek("MLB")
		case "UFC":
			go GetCurrentMonth("UFC")
		case "Bellator":
			go GetCurrentMonth("Bellator")
		default:

		}

		Owner.S_feed.SetText("dReams Client")
		Owner.S_feed.Refresh()
	}
	Owner.S_league.SetPlaceHolder("League:")

	Owner.S_end = dwidget.DeroAmtEntry("", 1, 0)
	Owner.S_end.SetPlaceHolder("Closes At:")
	Owner.S_end.Validator = validation.NewRegexp(`^\d{10,}$`, "Unix time required")

	Owner.S_amt = dwidget.DeroAmtEntry("", 0.1, 1)
	Owner.S_amt.SetPlaceHolder("Minimum Amount:")
	Owner.S_amt.AllowFloat = true
	Owner.S_amt.Wrapping = fyne.TextTruncate
	Owner.S_amt.Validator = validation.NewRegexp(`^\d{1,}\.\d{1,5}$|^[^0]\d{0,}$`, "Int or float required")

	feeds := []string{"dReams Client"}
	Owner.S_feed = widget.NewSelectEntry(feeds)
	Owner.S_feed.SetPlaceHolder("Feed:")

	Owner.S_deposit = dwidget.DeroAmtEntry("", 0.1, 1)
	Owner.S_deposit.SetPlaceHolder("Deposit Amount:")
	Owner.S_deposit.AllowFloat = true
	Owner.S_deposit.Wrapping = fyne.TextTruncate
	Owner.S_deposit.Validator = validation.NewRegexp(`^\d{1,}\.\d{1,5}$|^[^0]\d{0,}$`, "Int or float required")

	reset := window.Content().(*fyne.Container).Objects[2]

	Owner.S_set = widget.NewButton("Set Game", func() {
		if Owner.S_deposit.Validate() == nil && Owner.S_amt.Validate() == nil && Owner.S_end.Validate() == nil {
			if len(Sports.Contract) == 64 {
				window.Content().(*fyne.Container).Objects[2] = ownerConfirmAction(1, 100, window, reset)
				window.Content().(*fyne.Container).Objects[2].Refresh()
			}
		}
	})

	Owner.S_cancel = widget.NewButton("Cancel", func() {
		window.Content().(*fyne.Container).Objects[2] = ownerConfirmAction(9, 0, window, reset)
		window.Content().(*fyne.Container).Objects[2].Refresh()
	})

	Owner.S_cancel.Hide()

	Owner.Payout_n = widget.NewSelectEntry([]string{})
	Owner.Payout_n.SetPlaceHolder("Game #")

	sports_confirm := widget.NewButton("Sports Payout", func() {
		if len(Sports.Contract) == 64 {
			window.Content().(*fyne.Container).Objects[2] = ownerConfirmAction(3, 100, window, reset)
			window.Content().(*fyne.Container).Objects[2].Refresh()
		}
	})

	sports := container.NewVBox(
		humanTimeConvert(),
		layout.NewSpacer(),
		Owner.S_league,
		Owner.S_game,
		Owner.S_end,
		Owner.S_amt,
		Owner.S_feed,
		Owner.S_deposit,
		Owner.S_set,
		layout.NewSpacer(),
		Owner.S_cancel,
		layout.NewSpacer(),
		Owner.Payout_n,
		sports_confirm,
		layout.NewSpacer())

	return sports
}

// dService control objects for side menu
//   - Pass side menu window to reset to
func serviceOpts(window fyne.Window) fyne.CanvasObject {
	get_addr := widget.NewButton("Integrated Address", func() {
		go makeIntegratedAddr(true)
	})

	txid := widget.NewMultiLineEntry()
	txid.SetPlaceHolder("TXID:")
	txid.Wrapping = fyne.TextWrapWord
	txid.Validator = validation.NewRegexp(`^\w{64,64}$`, "Invalid TXID")

	process := widget.NewButton("Process Tx", func() {
		if !Service.IsProcessing() && !Service.IsRunning() {
			if txid.Validate() == nil {
				processSingleTx(txid.Text)
			}
		} else {
			log.Println("[dReams] Stop service to manually process Tx")
		}
	})

	delete := widget.NewButton("Delete Tx", func() {
		if !Service.IsProcessing() && !Service.IsRunning() {
			if txid.Validate() == nil {
				e := rpc.GetWalletTx(txid.Text)
				if e != nil {
					if db := boltDB(); db != nil {
						defer db.Close()
						deleteTx("BET", db, *e)
					}
				}
			}
		} else {
			log.Println("[dReams] Stop service to delete Tx")
		}
	})

	store := widget.NewButton("Store Tx", func() {
		if !Service.IsProcessing() && !Service.IsRunning() {
			if txid.Validate() == nil {
				e := rpc.GetWalletTx(txid.Text)
				if e != nil {
					if db := boltDB(); db != nil {
						defer db.Close()
						storeTx("BET", "done", db, *e)
					}
				}
			}
		} else {
			log.Println("[dReams] Stop service to store Tx")
		}
	})

	entry := dwidget.DeroAmtEntry("", 1, 0)
	entry.SetPlaceHolder("Block #:")
	entry.AllowFloat = false
	entry.Wrapping = fyne.TextTruncate
	entry.Validator = validation.NewRegexp(`^[^0]\d{0,}$`, "Int required")

	var start uint64
	height := widget.NewCheck("Start from current height", func(b bool) {
		if b {
			start = rpc.DaemonHeight("dReams", rpc.Daemon.Rpc)
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
		if !Service.IsProcessing() && !Service.IsRunning() {
			if !height.Checked {
				start = uint64(rpc.StringToInt(entry.Text))
			}
			viewProcessedTx(start)
		} else {
			log.Println("[dReams] Stop service to view Tx history")
		}
	})

	Owner.Service_pay = widget.NewCheck("Payouts", func(b bool) {
		if b {
			Owner.Payout_on = true
		} else {
			Owner.Payout_on = false
		}
	})

	if Owner.Payout_on {
		Owner.Service_pay.SetChecked(true)
		Owner.Service_pay.Disable()
	}

	Owner.Transactions = widget.NewCheck("Transactions", func(b bool) {
		if b {
			Owner.Transact_on = true
		} else {
			Owner.Transact_on = false
		}
	})

	if Owner.Transact_on {
		Owner.Transactions.SetChecked(true)
		Owner.Transactions.Disable()
	}

	reset := window.Content().(*fyne.Container).Objects[2]

	Owner.Run_service = widget.NewButton("Run Service", func() {
		if !Service.IsRunning() {
			if entry.Validate() == nil {
				if !height.Checked {
					start = uint64(rpc.StringToInt(entry.Text))
					if start < PAYLOAD_FORMAT {
						start = PAYLOAD_FORMAT
					}
				}

				if Owner.Service_pay.Checked || Owner.Transactions.Checked {
					go func() {
						Service.Start()
						Owner.Run_service.Hide()
						window.Content().(*fyne.Container).Objects[2] = serviceRunConfirm(start, Owner.Service_pay.Checked, Owner.Transactions.Checked, window, reset)
						window.Content().(*fyne.Container).Objects[2].Refresh()
					}()
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

	if Service.IsRunning() || Service.IsProcessing() {
		Owner.Run_service.Hide()
	}

	stop := widget.NewButton("Stop Service", func() {
		if Service.IsRunning() {
			log.Println("[dReams] Stopping service")
		}
		Service.Stop()

	})

	box := container.NewVBox(
		layout.NewSpacer(),
		view,
		layout.NewSpacer(),
		txid,
		container.NewAdaptiveGrid(3,
			process,
			delete,
			store),
		layout.NewSpacer(),
		get_addr,
		layout.NewSpacer(),
		height,
		entry,
		Owner.Service_pay,
		Owner.Transactions,
		debug,
		container.NewAdaptiveGrid(2,
			stop,
			Owner.Run_service,
		))

	return box
}

// SCID update objects for side menu
func updateOpts() fyne.CanvasObject {
	a_label := widget.NewLabel("Time A         ")
	a := dwidget.DeroAmtEntry("", 1, 0)
	a.SetPlaceHolder("Time A:")
	a.AllowFloat = false
	a.Wrapping = fyne.TextTruncate
	a.Validator = validation.NewRegexp(`[^0]\d{1,}$`, "Int required")

	b_label := widget.NewLabel("Time B         ")
	b := dwidget.DeroAmtEntry("", 1, 0)
	b.SetPlaceHolder("Time B:")
	b.AllowFloat = false
	b.Wrapping = fyne.TextTruncate
	b.Validator = validation.NewRegexp(`[^0]\d{1,}$`, "Int required")

	c_label := widget.NewLabel("Time C         ")
	c := dwidget.DeroAmtEntry("", 1, 0)
	c.SetPlaceHolder("Time C:")
	c.AllowFloat = false
	c.Wrapping = fyne.TextTruncate
	c.Validator = validation.NewRegexp(`[^0]\d{1,}$`, "Int required")

	hl_label := widget.NewLabel("Max Games")
	hl := dwidget.DeroAmtEntry("", 1, 0)
	hl.SetPlaceHolder("Max Games:")
	hl.AllowFloat = false
	hl.Wrapping = fyne.TextTruncate
	hl.Validator = validation.NewRegexp(`^[^0]\d{0,}$`, "Int required")

	hl_box := container.NewBorder(nil, nil, hl_label, nil, hl)
	hl_box.Hide()

	// l := dwidget.WholeAmtEntry()
	// l.PlaceHolder = "L:"
	// l.Validator = validation.NewRegexp(`^\d{2,}$`, "Format Not Valid")

	sc := widget.NewSelect([]string{"Prediction", "Sports"}, func(s string) {
		if s == "Sports" {
			c_label.SetText("Delete         ")
			c.Validator = validation.NewRegexp(`[^0]\d{0,}$`, "Int required")
			hl_box.Show()
		} else {
			c_label.SetText("Time C         ")
			c.Validator = validation.NewRegexp(`[^0]\d{1,}$`, "Int required")
			hl_box.Hide()
		}
	})
	sc.PlaceHolder = "Select Contract"

	new_owner := widget.NewMultiLineEntry()
	new_owner.Validator = validation.NewRegexp(`^(dero)\w{62}$`, "Invalid Address")
	new_owner.Wrapping = fyne.TextWrapWord
	new_owner.SetPlaceHolder("New owner address:")
	add_owner := widget.NewButton("Add Owner", func() {
		if new_owner.Validate() == nil {
			switch sc.Selected {
			case "Prediction":
				AddOwner(Predict.Contract, new_owner.Text)
			case "Sports":
				AddOwner(Sports.Contract, new_owner.Text)
			default:
				log.Println("[dReams] Select contract")
			}
		}
	})

	owner_num := dwidget.DeroAmtEntry("", 1, 0)
	owner_num.SetPlaceHolder("Owner #:")
	owner_num.AllowFloat = false
	owner_num.Validator = validation.NewRegexp(`^[^0]\d{0,0}$`, "Int required")
	owner_num.Wrapping = fyne.TextTruncate

	remove_owner := widget.NewButton("Remove Owner", func() {
		switch sc.Selected {
		case "Prediction":
			RemoveOwner(Predict.Contract, rpc.StringToInt(owner_num.Text))
		case "Sports":
			RemoveOwner(Sports.Contract, rpc.StringToInt(owner_num.Text))
		default:
			log.Println("[dReams] Select contract")
		}
	})

	update_var := widget.NewButton("Update Variables", func() {
		if a.Validate() == nil && b.Validate() == nil && c.Validate() == nil {
			switch sc.Selected {
			case "Prediction":
				VarUpdate(Predict.Contract, rpc.StringToInt(a.Text), rpc.StringToInt(b.Text), rpc.StringToInt(c.Text), 30, 0)
			case "Sports":
				if hl.Validate() == nil {
					VarUpdate(Sports.Contract, rpc.StringToInt(a.Text), rpc.StringToInt(b.Text), rpc.StringToInt(c.Text), 30, rpc.StringToInt(hl.Text))
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

// dSports and dPrediction action confirmation
//   - i defines the action to be confirmed
//   - teamA, teamB needed only for dSports confirmations
//   - Pass main window obj and tabs to reset to
func ConfirmAction(i int, teamA, teamB string, obj []fyne.CanvasObject, tabs *container.AppTabs) fyne.CanvasObject {
	var confirm_display = widget.NewLabel("")
	confirm_display.Wrapping = fyne.TextWrapWord
	confirm_display.Alignment = fyne.TextAlignCenter

	p_scid := Predict.Contract

	s_scid := Sports.Contract
	split := strings.Split(Sports.Game_select.Selected, "   ")
	multi := Sports.Multi.Selected

	switch i {
	case 1:
		float := float64(Predict.Amount)
		amt := float / 100000
		confirm_display.SetText(fmt.Sprintf("SCID:\n\n%s\n\nLower prediction for %.5f Dero\n\nConfirm", p_scid, amt))
	case 2:
		float := float64(Predict.Amount)
		amt := float / 100000
		confirm_display.SetText(fmt.Sprintf("SCID:\n\n%s\n\nHigher prediction for %.5f Dero\n\nConfirm", p_scid, amt))
	case 3:
		game := Sports.Game_select.Selected
		val := float64(GetSportsAmt(s_scid, split[0]))
		var x string

		switch multi {
		case "3x":
			x = fmt.Sprint(val * 3 / 100000)
		case "5x":
			x = fmt.Sprint(val * 5 / 100000)
		default:
			x = fmt.Sprint(val / 100000)
		}

		confirm_display.SetText(fmt.Sprintf("SCID:\n\n%s\n\nBetting on Game # %s\n\n%s for %s Dero\n\nConfirm", s_scid, game, teamA, x))
	case 4:
		game := Sports.Game_select.Selected
		val := float64(GetSportsAmt(s_scid, split[0]))
		var x string

		switch multi {
		case "3x":
			x = fmt.Sprint(val * 3 / 100000)
		case "5x":
			x = fmt.Sprint(val * 5 / 100000)
		default:
			x = fmt.Sprint(val / 100000)
		}

		confirm_display.SetText(fmt.Sprintf("SCID:\n\n%s\n\nBetting on Game # %s\n\n%s for %s Dero\n\nConfirm", s_scid, game, teamB, x))
	default:
		log.Println("[dReams] No Confirm Input")
		confirm_display.SetText("Error")
	}

	cancel_button := widget.NewButton("Cancel", func() {
		obj[1] = tabs
		obj[1].Refresh()
	})

	confirm_button := widget.NewButton("Confirm", func() {
		switch i {
		case 1:
			PredictLower(p_scid, "")
		case 2:
			PredictHigher(p_scid, "")
		case 3:
			PickTeam(s_scid, multi, split[0], GetSportsAmt(s_scid, split[0]), 0)
		case 4:
			PickTeam(s_scid, multi, split[0], GetSportsAmt(s_scid, split[0]), 1)
		default:

		}

		obj[1] = tabs
		obj[1].Refresh()
	})

	display := container.NewVBox(layout.NewSpacer(), confirm_display, layout.NewSpacer())
	options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewBorder(nil, options, nil, nil, display)

	go func() {
		for rpc.IsReady() {
			time.Sleep(time.Second)
		}

		obj[1] = tabs
		obj[1].Refresh()
	}()

	return container.NewMax(bundle.Alpha120, content)
}

// dReam Service start confirmation
//   - start is starting height to run service
//   - payout and transfers, params for service
//   - Pass side window to reset to
func serviceRunConfirm(start uint64, payout, transfers bool, window fyne.Window, reset fyne.CanvasObject) fyne.CanvasObject {
	var pay, transac string
	if transfers {
		transac = "process transactions sent to your integrated address"
		if payout {
			transac = transac + " "
		}
	}

	if payout {
		if transfers {
			pay = "and "
		}
		pay = pay + "process payouts to contracts"
	}

	str := fmt.Sprintf("This will automatically %s%s.\n\nStarting service from height %d", transac, pay, start)
	confirm_display := widget.NewLabel(str)
	confirm_display.Wrapping = fyne.TextWrapWord
	confirm_display.Alignment = fyne.TextAlignCenter

	cancel_button := widget.NewButton("Cancel", func() {
		Service.Stop()
		window.Content().(*fyne.Container).Objects[2] = reset
		window.Content().(*fyne.Container).Objects[2].Refresh()
	})

	confirm_button := widget.NewButton("Confirm", func() {
		go RunService(start, payout, transfers)
		window.Content().(*fyne.Container).Objects[2] = reset
		window.Content().(*fyne.Container).Objects[2].Refresh()
	})

	display := container.NewVBox(layout.NewSpacer(), confirm_display, layout.NewSpacer())
	options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewBorder(nil, options, nil, nil, display)

	return container.NewMax(content)
}

// Convert unix time to human readable time
func humanTimeConvert() fyne.CanvasObject {
	entry := dwidget.DeroAmtEntry("", 1, 0)
	entry.AllowFloat = false
	entry.SetPlaceHolder("Unix time:")
	entry.Validator = validation.NewRegexp(`^\d{10,}$`, "Unix time required")
	res := widget.NewEntry()
	res.Disable()
	button := widget.NewButton("Human Time", func() {
		if entry.Validate() == nil {
			time := time.Unix(int64(rpc.StringToInt(entry.Text)), 0).String()
			res.SetText(time)
		}
	})

	split := container.NewHSplit(entry, button)
	box := container.NewVBox(res, split)

	return box
}

// Check dPrediction SCID for live status
func CheckPredictionStatus() {
	if rpc.Daemon.IsConnected() && menu.Gnomes.IsReady() {
		_, ends := menu.Gnomes.GetSCIDValuesByKey(Predict.Contract, "p_end_at")
		_, time_a := menu.Gnomes.GetSCIDValuesByKey(Predict.Contract, "time_a")
		_, time_c := menu.Gnomes.GetSCIDValuesByKey(Predict.Contract, "time_c")
		_, mark := menu.Gnomes.GetSCIDValuesByKey(Predict.Contract, "mark")
		if ends != nil && time_a != nil && time_c != nil {
			now := uint64(time.Now().Unix())
			if now >= ends[0] && now <= ends[0]+time_a[0] && mark == nil {
				Owner.P_post.Show()
			} else {
				Owner.P_post.Hide()
			}

			if now >= ends[0]+time_c[0] {
				Owner.P_pay.Show()
			} else {
				Owner.P_pay.Hide()
			}
		}

		if ends == nil {
			Owner.P_post.Hide()
			Owner.P_pay.Hide()
		}
	}
}

// Check dSports SCID for active games
func GetActiveGames() {
	if rpc.Daemon.IsConnected() && menu.Gnomes.IsReady() {
		options := []string{}
		contracts := menu.Gnomes.GetAllOwnersAndSCIDs()
		keys := make([]string, len(contracts))

		i := 0
		for k := range contracts {
			keys[i] = k
			owner, _ := menu.Gnomes.GetSCIDValuesByKey(keys[i], "owner")
			if (owner != nil && owner[0] == rpc.Wallet.Address) || VerifyBetSigner(keys[i]) {
				if len(keys[i]) == 64 {
					_, init := menu.Gnomes.GetSCIDValuesByKey(keys[i], "s_init")
					if init != nil {
						for ic := uint64(1); ic <= init[0]; ic++ {
							num := strconv.Itoa(int(ic))
							game, _ := menu.Gnomes.GetSCIDValuesByKey(keys[i], "game_"+num)
							league, _ := menu.Gnomes.GetSCIDValuesByKey(keys[i], "league_"+num)
							_, end := menu.Gnomes.GetSCIDValuesByKey(keys[i], "s_end_at_"+num)
							_, add := menu.Gnomes.GetSCIDValuesByKey(keys[i], "time_a")
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
		Owner.Payout_n.SetOptions(options)
	}
}

// Bet contract owner control menu
func ownersMenu() {
	ow := fyne.CurrentApp().NewWindow("Bet Contracts")
	ow.Resize(fyne.NewSize(330, 700))
	ow.SetIcon(bundle.ResourceDReamsIconAltPng)
	Predict.Settings.Menu.Hide()
	Sports.Settings.Menu.Hide()
	quit := make(chan struct{})
	ow.SetCloseIntercept(func() {
		Predict.Settings.Menu.Show()
		Sports.Settings.Menu.Show()
		quit <- struct{}{}
		ow.Close()
	})
	ow.SetFixedSize(true)

	owner_tabs := container.NewAppTabs(
		container.NewTabItem("Predict", layout.NewSpacer()),
		container.NewTabItem("Sports", layout.NewSpacer()),
		container.NewTabItem("Service", layout.NewSpacer()),
		container.NewTabItem("Update", updateOpts()),
	)
	owner_tabs.SetTabLocation(container.TabLocationTop)
	owner_tabs.OnSelected = func(ti *container.TabItem) {
		switch ti.Text {
		case "Sports":
			go GetActiveGames()
		case "Service":
			go makeIntegratedAddr(false)
		}
	}

	var utime string
	clock := widget.NewEntry()
	clock.Disable()

	entry := dwidget.DeroAmtEntry("", 1, 0)
	entry.AllowFloat = false
	entry.SetPlaceHolder("Hours to close:")
	entry.Validator = validation.NewRegexp(`^\d{1,}$`, "Int required")
	button := widget.NewButton("Add Hours", func() {
		if entry.Validate() == nil {
			i := rpc.StringToInt(entry.Text)
			u := rpc.StringToInt(utime)
			r := u + (i * 3600)

			switch owner_tabs.SelectedIndex() {
			case 0:
				Owner.P_end.SetText(strconv.Itoa(r))
			case 1:
				Owner.S_end.SetText(strconv.Itoa(r))
			}
		}
	})

	go func() {
		var ticker = time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				if !rpc.Wallet.IsConnected() {
					ticker.Stop()
					ow.Close()
				}

				if Service.IsRunning() {
					Owner.Run_service.Hide()
					Owner.Service_pay.Disable()
					Owner.Transactions.Disable()
				}

				if !Service.IsRunning() && !Service.IsProcessing() {
					Owner.Run_service.Show()
					Owner.Service_pay.Enable()
					Owner.Transactions.Enable()
				}

				CheckPredictionStatus()
				now := time.Now()
				utime = strconv.Itoa(int(now.Unix()))
				clock.SetText("Unix Time: " + utime)
				if now.Unix() < Predict.Buffer {
					if Predict.Init {
						Owner.P_set.Hide()
						Owner.P_cancel.Show()
					} else {
						Owner.P_set.Show()
						Owner.P_cancel.Hide()
					}
				} else {
					Owner.P_cancel.Hide()
					if Predict.Init {
						Owner.P_set.Hide()
					} else {
						Owner.P_set.Show()
					}
				}

				if Sports.Buffer {
					Owner.S_cancel.Show()
					Owner.S_set.Hide()
				} else {
					Owner.S_cancel.Hide()
					Owner.S_set.Show()
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

	alpha := canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
	if bundle.AppColor == color.White {
		alpha = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x99})
	}

	go func() {
		time.Sleep(200 * time.Millisecond)
		ow.SetContent(
			container.New(
				layout.NewMaxLayout(),
				menu.BackgroundRast("ownersMenu"),
				alpha,
				border))

		owner_tabs.SelectIndex(2)
		owner_tabs.Selected().Content = serviceOpts(ow)
		owner_tabs.SelectIndex(1)
		owner_tabs.Selected().Content = sportsOpts(ow)
		owner_tabs.SelectIndex(0)
		owner_tabs.Selected().Content = predictionOpts(ow)

		time.Sleep(time.Second)
		markets := []string{}
		if stored, ok := rpc.FindStringKey(rpc.RatingSCID, "prediction_markets", rpc.Daemon.Rpc).(string); ok {
			if h, err := hex.DecodeString(stored); err == nil {
				if err = json.Unmarshal(h, &markets); err == nil {
					Owner.P_Name.SetOptions(markets)
				}
			}
		}

		leagues := []string{}
		if stored, ok := rpc.FindStringKey(rpc.RatingSCID, "sports_leagues", rpc.Daemon.Rpc).(string); ok {
			if h, err := hex.DecodeString(stored); err == nil {
				if err = json.Unmarshal(h, &leagues); err == nil {
					Owner.S_league.SetOptions(leagues)
				}
			}
		}
	}()

	ow.Show()
}

// Bet contract owner action confirmation
//   - i defines action to be confirmed
//   - p for prediction price
//   - Pass side window to reset to
func ownerConfirmAction(i int, p float64, window fyne.Window, reset fyne.CanvasObject) fyne.CanvasObject {
	var confirm_display = widget.NewLabel("")
	confirm_display.Wrapping = fyne.TextWrapWord
	confirm_display.Alignment = fyne.TextAlignCenter

	pre := Predict.Prediction
	p_scid := Predict.Contract
	p_pre := Owner.P_Name.Text
	p_amt := Owner.P_amt.Text
	if p_amt_f, err := strconv.ParseFloat(p_amt, 64); err == nil {
		p_amt = fmt.Sprintf("%.5f", p_amt_f)
	}
	p_mark := Owner.P_mark.Text
	p_end := Owner.P_end.Text
	p_end_time, _ := rpc.MsToTime(p_end + "000")
	p_feed := Owner.P_feed.Text
	p_dep := Owner.P_deposit.Text
	if p_dep_f, err := strconv.ParseFloat(p_dep, 64); err == nil {
		p_dep = fmt.Sprintf("%.5f", p_dep_f)
	}
	var price string
	if menu.CoinDecimal(pre) == 8 {
		price = fmt.Sprintf("%.8f", p/100000000)
	} else {
		price = fmt.Sprintf("%.2f", p/100)
	}

	var s_game string
	s_scid := Sports.Contract
	game_split := strings.Split(Owner.S_game.Selected, "   ")
	if len(game_split) > 1 {
		s_game = game_split[1]
	} else {
		s_game = game_split[0]
	}

	s_league := Owner.S_league.Text
	s_amt := Owner.S_amt.Text
	if s_amt_f, err := strconv.ParseFloat(s_amt, 64); err == nil {
		s_amt = fmt.Sprintf("%.5f", s_amt_f)
	}
	s_end := Owner.S_end.Text
	s_end_time, _ := rpc.MsToTime(s_end + "000")
	s_feed := Owner.S_feed.Text
	n_split := strings.Split(Owner.Payout_n.Text, "   ")
	s_pay_n := n_split[0]
	s_dep := Owner.S_deposit.Text
	if s_dep_f, err := strconv.ParseFloat(s_dep, 64); err == nil {
		s_dep = fmt.Sprintf("%.5f", s_dep_f)
	}

	var win, team, a_score, b_score, payout_str string
	if i == 3 {
		if len(n_split) < 3 {
			log.Println("[dReams] Could not format game string")
			i = 100
		}
		if n_split[1] == "Bellator" || n_split[1] == "UFC" {
			win, team = GetMmaWinner(n_split[2], n_split[1])
			payout_str = fmt.Sprintf("SCID:\n\n%s\n\nFight: %s\n\nWinner: %s\n\nConfirm", s_scid, Owner.Payout_n.Text, team)
		} else {
			win, team, a_score, b_score = GetWinner(n_split[2], n_split[1])
			payout_str = fmt.Sprintf("SCID:\n\n%s\n\nGame: %s\n\n%s: %s\n%s: %s\n\nWinner: %s\n\nConfirm", s_scid, Owner.Payout_n.Text, TrimTeamA(n_split[2]), a_score, TrimTeamB(n_split[2]), b_score, team)
		}
	}

	switch i {
	case 1:
		confirm_display.SetText("SCID:\n\n" + s_scid + "\n\nGame: " + s_game + "\n\nMinimum: " + s_amt + " Dero\n\nCloses At: " + s_end_time.String() + "\n\nFeed: " + s_feed + "\n\nInitial Deposit: " + s_dep + " Dero")
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
					if menu.CoinDecimal(pre) == 8 || menu.CoinDecimal(p_pre) == 8 {
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

		confirm_display.SetText("SCID:\n\n" + p_scid + "\n\nPredicting: " + p_pre + "\n\nMinimum: " + p_amt + " Dero\n\nCloses At: " + p_end_time.String() + "\n\nMark: " + mark + "\n\n" + fn + p_feed + "\n\nInitial Deposit: " + p_dep + " Dero")

	case 3:
		confirm_display.SetText(payout_str)
	case 4:
		confirm_display.SetText("SCID:\n\n" + p_scid + "\n\nFeed from: dReams Client\n\nPost Price: " + price + "\n\nConfirm")
	case 5:
		confirm_display.SetText("SCID:\n\n" + p_scid + "\n\nFeed from: dReams Client\n\nFinal Price: " + price + "\n\nConfirm")
	case 6:
		switch onChainPrediction(pre) {
		case 1:
			confirm_display.SetText("SCID:\n\n" + p_scid + "\n\n" + pre + ": " + fmt.Sprintf("%.0f", p) + "\n\nNode: " + Predict.Feed + "\n\nConfirm Post")
		case 2:
			confirm_display.SetText("SCID:\n\n" + p_scid + "\n\n" + pre + ": " + fmt.Sprintf("%.5f", p) + "\n\nNode: " + Predict.Feed + "\n\nConfirm Post")
		case 3:
			confirm_display.SetText("SCID:\n\n" + p_scid + "\n\n" + pre + ": " + fmt.Sprintf("%.0f", p) + "\n\nNode: " + Predict.Feed + "\n\nConfirm Post")
		}

	case 7:
		switch onChainPrediction(pre) {
		case 1:
			confirm_display.SetText("SCID:\n\n" + p_scid + "\n\n" + pre + ": " + fmt.Sprintf("%.0f", p) + "\n\nNode: " + Predict.Feed + "\n\nConfirm Payout")
		case 2:
			confirm_display.SetText("SCID:\n\n" + p_scid + "\n\n" + pre + ": " + fmt.Sprintf("%.5f", p) + "\n\nNode: " + Predict.Feed + "\n\nConfirm Payout")
		case 3:
			confirm_display.SetText("SCID:\n\n" + p_scid + "\n\n" + pre + ": " + fmt.Sprintf("%.0f", p) + "\n\nNode: " + Predict.Feed + "\n\nConfirm Payout")
		}

	case 8:
		confirm_display.SetText("SCID:\n\n" + p_scid + "\n\nThis will Cancel the current prediction")
	case 9:
		confirm_display.SetText("SCID:\n\n" + s_scid + "\n\nThis will Cancel the last initiated bet on this contract")
	default:
		log.Println("[dReams] No Confirm Input")
		confirm_display.SetText("Error")
	}

	cancel_button := widget.NewButton("Cancel", func() {
		window.Content().(*fyne.Container).Objects[2] = reset
		window.Content().(*fyne.Container).Objects[2].Refresh()
	})

	confirm_button := widget.NewButton("Confirm", func() {
		Owner.Payout_n.SetText("")
		switch i {
		case 1:
			SetSports(rpc.StringToInt(s_end), rpc.ToAtomic(s_amt, 5), rpc.ToAtomic(s_dep, 5), s_scid, s_league, s_game, s_feed)
		case 2:
			if onChainPrediction(pre) == 2 || onChainPrediction(p_pre) == 2 { /// decimal of one place for block time
				SetPrediction(rpc.StringToInt(p_end), rpc.StringToInt(p_mark)*10000, rpc.ToAtomic(p_amt, 5), rpc.ToAtomic(p_dep, 5), p_scid, p_pre, p_feed)
			} else {
				SetPrediction(rpc.StringToInt(p_end), rpc.StringToInt(p_mark), rpc.ToAtomic(p_amt, 5), rpc.ToAtomic(p_dep, 5), p_scid, p_pre, p_feed)
			}
		case 3:
			EndSports(s_scid, s_pay_n, win)
		case 4:
			PostPrediction(p_scid, int(p))
		case 5:
			EndPrediction(p_scid, int(p))
		case 6:
			switch onChainPrediction(pre) {
			case 1:
				PostPrediction(p_scid, int(p))
			case 2:
				PostPrediction(p_scid, int(p*100000))
			case 3:
				PostPrediction(p_scid, int(p))
			default:
			}
		case 7:
			switch onChainPrediction(pre) {
			case 1:
				EndPrediction(p_scid, int(p))
			case 2:
				EndPrediction(p_scid, int(p*100000))
			case 3:
				EndPrediction(p_scid, int(p))
			default:
			}
		case 8:
			CancelInitiatedBet(Predict.Contract, 0)
		case 9:
			CancelInitiatedBet(Sports.Contract, 1)
		default:

		}

		window.Content().(*fyne.Container).Objects[2] = reset
		window.Content().(*fyne.Container).Objects[2].Refresh()
	})

	display := container.NewVBox(layout.NewSpacer(), confirm_display, layout.NewSpacer())
	options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewBorder(nil, options, nil, nil, display)

	return container.NewMax(content)
}

// Confirmation for dPrediction contract installs
func newPredictConfirm(c int, obj []fyne.CanvasObject, tabs *container.AppTabs) fyne.CanvasObject {
	var text string
	gas_fee := 0.125
	unlock_fee := float64(rpc.UnlockFee) / 100000
	switch c {
	case 1:
		text = `You are about to unlock and install your first dPrediction contract 
		
To help support the project, there is a ` + fmt.Sprintf("%.5f", unlock_fee) + ` DERO donation attached to preform this action

Unlocking dPrediction or dSports gives you unlimited access to bet contract uploads and all base level owner features

Total transaction will be ` + fmt.Sprintf("%0.5f", unlock_fee+gas_fee) + ` DERO (0.12500 gas fee for contract install)


Select a public or private contract

Public will show up in indexed list of contracts

Private will not show up in the list`
	case 2:
		text = `You are about to install a new dPrediction contract

Gas fee to install contract is 0.12500 DERO


Select a public or private contract

Public will show up in indexed list of contracts

Private will not show up in the list`
	}

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	var choice *widget.Select

	pre_button := widget.NewButton("Install", func() {
		if choice.SelectedIndex() < 2 && choice.SelectedIndex() >= 0 {
			UploadBetContract(true, choice.SelectedIndex())
		}

		obj[1] = tabs
		obj[1].Refresh()
	})

	pre_button.Hide()

	options := []string{"Public", "Private"}
	choice = widget.NewSelect(options, func(s string) {
		if s == "Public" || s == "Private" {
			pre_button.Show()
		} else {
			pre_button.Hide()
		}
	})

	cancel_button := widget.NewButton("Cancel", func() {
		obj[1] = tabs
		obj[1].Refresh()
	})

	left := container.NewVBox(pre_button)
	right := container.NewVBox(cancel_button)
	buttons := container.NewAdaptiveGrid(3, left, container.NewVBox(layout.NewSpacer()), right)
	actions := container.NewVBox(choice, buttons)
	info_box := container.NewVBox(layout.NewSpacer(), label, layout.NewSpacer())

	content := container.NewBorder(nil, actions, nil, nil, info_box)

	go func() {
		for rpc.IsReady() {
			time.Sleep(time.Second)
		}

		obj[1] = tabs
		obj[1].Refresh()
	}()

	return container.NewMax(content)
}

// Confirmation for dSports contract installs
func newSportsConfirm(c int, obj []fyne.CanvasObject, tabs *container.AppTabs) fyne.CanvasObject {
	var text string
	gas_fee := 0.14
	unlock_fee := float64(rpc.UnlockFee) / 100000
	switch c {
	case 1:
		text = `You are about to unlock and install your first dSports contract
		
To help support the project, there is a ` + fmt.Sprintf("%.5f", unlock_fee) + ` DERO donation attached to preform this action

Unlocking dPrediction or dSports gives you unlimited access to bet contract uploads and all base level owner features

Total transaction will be ` + fmt.Sprintf("%0.5f", unlock_fee+gas_fee) + ` DERO (0.14000 gas fee for contract install)


Select a public or private contract

Public will show up in indexed list of contracts

Private will not show up in the list`
	case 2:
		text = `You are about to install a new dSports contract

Gas fee to install contract is 0.14000 DERO


Select a public or private contract

Public will show up in indexed list of contracts

Private will not show up in the list`
	}

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	var choice *widget.Select

	sports_button := widget.NewButton("Install", func() {
		if choice.SelectedIndex() < 2 && choice.SelectedIndex() >= 0 {
			UploadBetContract(false, choice.SelectedIndex())
		}

		obj[1] = tabs
		obj[1].Refresh()
	})

	sports_button.Hide()

	options := []string{"Public", "Private"}
	choice = widget.NewSelect(options, func(s string) {
		if s == "Public" || s == "Private" {
			sports_button.Show()
		} else {
			sports_button.Hide()
		}
	})

	cancel_button := widget.NewButton("Cancel", func() {
		obj[1] = tabs
		obj[1].Refresh()
	})

	left := container.NewVBox(sports_button)
	right := container.NewVBox(cancel_button)
	buttons := container.NewAdaptiveGrid(3, left, container.NewVBox(layout.NewSpacer()), right)
	actions := container.NewVBox(choice, buttons)
	info_box := container.NewVBox(layout.NewSpacer(), label, layout.NewSpacer())

	content := container.NewBorder(nil, actions, nil, nil, info_box)

	go func() {
		for rpc.IsReady() {
			time.Sleep(time.Second)
		}

		obj[1] = tabs
		obj[1].Refresh()
	}()

	return container.NewMax(content)
}
