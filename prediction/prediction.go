package prediction

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	dreams "github.com/SixofClubsss/dReams"
	"github.com/SixofClubsss/dReams/dwidget"
	"github.com/SixofClubsss/dReams/holdero"
	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type predictObjects struct {
	Init          bool
	owner         bool
	Amount        uint64
	Buffer        int64
	Contract      string
	Prediction    string
	Feed          string
	Info          *widget.Label
	Prices        *widget.Label
	Predict_list  *widget.List
	Favorite_list *widget.List
	Owned_list    *widget.List
	Higher        *widget.Button
	Lower         *widget.Button
	Container     *fyne.Container
	Settings      struct {
		Contracts      []string
		Favorites      []string
		Owned          []string
		Unlock         *widget.Button
		New            *widget.Button
		Menu           *widget.Button
		Check          *widget.Check
		Contract_entry *widget.SelectEntry
	}
}

var Predict predictObjects

// Disable dPrediction objects
func disablePredictions(d bool) {
	if d {
		Predict.Container.Hide()
	} else {
		Predict.Container.Show()
	}
	Predict.Container.Refresh()
}

// Check box for dPrediction SCID
//   - Hides prediction controls on disconnect
func PredictConnectedBox() fyne.Widget {
	Predict.Settings.Check = widget.NewCheck("", func(b bool) {
		if !b {
			Predict.Higher.Hide()
			Predict.Lower.Hide()

		}
	})
	Predict.Settings.Check.Disable()

	return Predict.Settings.Check
}

// Entry for dPrediction SCID
//   - Bound to Predict.Contract
//   - Checks for valid SCID on changed
func PredictionContractEntry() fyne.Widget {
	options := []string{}
	Predict.Settings.Contract_entry = widget.NewSelectEntry(options)
	Predict.Settings.Contract_entry.PlaceHolder = "Contract Address: "
	Predict.Settings.Contract_entry.OnCursorChanged = func() {
		if rpc.Daemon.IsConnected() {
			go func() {
				if len(Predict.Contract) == 64 {
					yes := ValidBetContract(Predict.Contract)
					if yes {
						Predict.Settings.Check.SetChecked(true)
					} else {
						Predict.Settings.Check.SetChecked(false)
					}
				} else {
					Predict.Settings.Check.SetChecked(false)
				}
			}()
		}
	}

	this := binding.BindString(&Predict.Contract)
	Predict.Settings.Contract_entry.Bind(this)

	return Predict.Settings.Contract_entry
}

// When called, enable and show dPrediction controls
func ShowPredictionControls() {
	disablePredictions(false)
	Predict.Higher.Show()
	Predict.Lower.Show()
}

// Routine when dPrediction SCID is clicked
//   - Sets label info and controls
//   - item returned for adding and removing favorites
func setPredictionControls(str string) (item string) {
	split := strings.Split(str, "   ")
	if len(split) >= 3 {
		trimmed := strings.Trim(split[2], " ")
		if len(trimmed) == 64 {
			item = str
			Predict.Settings.Contract_entry.SetText(trimmed)
			go SetPredictionInfo(trimmed)
			if CheckActivePrediction(trimmed) {
				ShowPredictionControls()
			} else {
				disablePredictions(true)
			}
		}
	}

	return
}

// Sets dPrediction info label
func SetPredictionInfo(scid string) {
	info := GetPrediction(scid)
	if info != "" {
		Predict.Info.SetText(info)
		Predict.Info.Refresh()
	}
}

// Update price feed for dPrediction display
func SetPredictionPrices(d bool) {
	if d {
		_, btc := holdero.GetPrice("BTC-USDT")
		_, dero := holdero.GetPrice("DERO-USDT")
		_, xmr := holdero.GetPrice("XMR-USDT")
		/// custom feed with rpc.Display.P_feed
		prices := "Current Price feed from dReams Client\nBTC: " + btc + "\nDERO: " + dero + "\nXMR: " + xmr

		Predict.Prices.SetText(prices)
	}
}

// List object for populating public dPrediction contracts, with rating and add favorite controls
//   - Pass tab for action confirmation reset
func PredictionListings(tab *container.AppTabs) fyne.CanvasObject {
	Predict.Predict_list = widget.NewList(
		func() int {
			return len(Predict.Settings.Contracts)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(canvas.NewImageFromImage(nil), widget.NewLabel(""))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*fyne.Container).Objects[1].(*widget.Label).SetText(Predict.Settings.Contracts[i])
			if Predict.Settings.Contracts[i][0:2] != "  " {
				var key string
				split := strings.Split(Predict.Settings.Contracts[i], "   ")
				if len(split) >= 3 {
					trimmed := strings.Trim(split[2], " ")
					if len(trimmed) == 64 {
						key = trimmed
					}
				}

				badge := canvas.NewImageFromResource(menu.DisplayRating(menu.Control.Contract_rating[key]))
				badge.SetMinSize(fyne.NewSize(35, 35))
				o.(*fyne.Container).Objects[0] = badge
			}
		})

	var item string

	Predict.Predict_list.OnSelected = func(id widget.ListItemID) {
		if id != 0 && menu.Connected() {
			item = setPredictionControls(Predict.Settings.Contracts[id])
			Predict.Favorite_list.UnselectAll()
			Predict.Owned_list.UnselectAll()
		} else {
			disablePredictions(true)
		}
	}

	save := widget.NewButton("Favorite", func() {
		Predict.Settings.Favorites = append(Predict.Settings.Favorites, item)
		sort.Strings(Predict.Settings.Favorites)
	})

	rate := widget.NewButton("Rate", func() {
		if len(Predict.Contract) == 64 {
			if !menu.CheckOwner(Predict.Contract) {
				reset := tab.Selected().Content
				tab.Selected().Content = menu.RateConfirm(Predict.Contract, tab, reset)
				tab.Selected().Content.Refresh()
			} else {
				log.Println("[dReams] You own this contract")
			}
		}
	})

	cont := container.NewBorder(
		nil,
		container.NewBorder(nil, nil, save, rate, layout.NewSpacer()),
		nil,
		nil,
		Predict.Predict_list)

	return cont
}

// List object for populating favorite dPrediction contracts, with remove favorite control
func PredictionFavorites() fyne.CanvasObject {
	Predict.Favorite_list = widget.NewList(
		func() int {
			return len(Predict.Settings.Favorites)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(Predict.Settings.Favorites[i])
		})

	var item string

	Predict.Favorite_list.OnSelected = func(id widget.ListItemID) {
		if menu.Connected() {
			item = setPredictionControls(Predict.Settings.Favorites[id])
			Predict.Predict_list.UnselectAll()
			Predict.Owned_list.UnselectAll()
		} else {
			disablePredictions(true)
		}
	}

	remove := widget.NewButton("Remove", func() {
		if len(Predict.Settings.Favorites) > 0 {
			Predict.Favorite_list.UnselectAll()
			new := Predict.Settings.Favorites
			for i := range new {
				if new[i] == item {
					copy(new[i:], new[i+1:])
					new[len(new)-1] = ""
					new = new[:len(new)-1]
					Predict.Settings.Favorites = new
					break
				}
			}
		}
		Predict.Favorite_list.Refresh()
		sort.Strings(Predict.Settings.Favorites)
	})

	cont := container.NewBorder(
		nil,
		container.NewBorder(nil, nil, nil, remove, layout.NewSpacer()),
		nil,
		nil,
		Predict.Favorite_list)

	return cont
}

// List object for populating owned dPrediction contracts
func PredictionOwned() fyne.CanvasObject {
	Predict.Owned_list = widget.NewList(
		func() int {
			return len(Predict.Settings.Owned)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(Predict.Settings.Owned[i])
		})

	Predict.Owned_list.OnSelected = func(id widget.ListItemID) {
		if menu.Connected() {
			setPredictionControls(Predict.Settings.Owned[id])
			Predict.Predict_list.UnselectAll()
			Predict.Favorite_list.UnselectAll()
		} else {
			disablePredictions(true)
		}
	}

	return Predict.Owned_list
}

// Refresh all dPrediction objects
func PredictionRefresh(p *dwidget.DreamsItems, d dreams.DreamsObject) {
	if d.OnTab("Predict") {
		if Predict.Prices.Text == "" {
			go SetPredictionPrices(rpc.Daemon.Connect)
		}

		p.RightLabel.SetText("dReams Balance: " + rpc.DisplayBalance("dReams") + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Wallet.Display.Height)

		if CheckActivePrediction(Predict.Contract) {
			go ShowPredictionControls()
		} else {
			disablePredictions(true)
		}
	}
}

// Formats initialized dPrediction info string
//   - p defines prediction
//   - amt is Dero value
//   - eA is prediction end time
//   - c is number of current predictions
//   - to is current total prediction Dero pot value
//   - u is higher predictions
//   - d is lower predictions
//   - r is total completed prediction rounds
//   - f is prediction feed
//   - m is prediction mark
//   - ta, tb, tc are current contract time frames
func P_initResults(p, amt, eA, c, to, u, d, r, f, m string, ta, tb, tc int) (info string) {
	end_time, _ := rpc.MsToTime(eA)
	utc := end_time.String()
	add := rpc.StringToInt(eA)
	end := strconv.Itoa(add + (tc * 1000))
	end_pay, _ := rpc.MsToTime(end)
	rf := strconv.Itoa(tb / 60)

	result, err := strconv.ParseFloat(to, 32)

	if err != nil {
		log.Println("[Predictions]", err)
	}

	s := fmt.Sprintf("%.5f", result/100000)

	now := time.Now().Unix()
	done := now > end_time.Unix()

	if done {
		mark := m
		if mark == "0" {
			mark = ""
		}

		var wfp string
		if mark == "" {
			wfp = "   Waiting for Mark"
		} else {
			wfp = "   Waiting for Payout"
		}

		if isOnChainPrediction(p) {
			info = "SCID:\n\n" + Predict.Contract + "\n\n" + p + wfp + "\n\nNode: " + f + "\n\nMark: " + mark + "\nRound Pot: " + s +
				" Dero\n\nPredictions: " + c + "\nHigher Predictions: " + u + "\nLower Predictions: " + d +
				"\n\nPayout After: " + end_pay.String() + "\nRefund if not paid within " + rf + " minutes\n\nRounds Completed: " + r
		} else {
			info = "SCID:\n\n" + Predict.Contract + "\n\n" + p + wfp + "\n\nMark: " + mark + "\nRound Pot: " + s +
				" Dero\n\nPredictions: " + c + "\nHigher Predictions: " + u + "\nLower Predictions: " + d +
				"\n\nPayout After: " + end_pay.String() + "\nRefund if not paid within " + rf + " minutes\n\nRounds Completed: " + r
		}

	} else {
		var live string
		if now > Predict.Buffer {
			live = "\n\nAccepting " + p + " Predictions "
		} else {
			left := Predict.Buffer - now
			live = "\n\n" + p + "\nBuffer ends in " + strconv.Itoa(int(left)) + " seconds"
		}

		if amt_f, err := strconv.ParseFloat(amt, 64); err == nil {
			amt = fmt.Sprintf("%.5f", amt_f)
		}

		node := ""
		if isOnChainPrediction(p) {
			node = "\n\nNode: " + f
		}

		if m == "0" {
			pw := strconv.Itoa(ta / 60)
			info = "SCID:\n\n" + Predict.Contract + live + node +
				"\n\nCloses at: " + utc + "\nMark posted with in " + pw + " minutes of close\n\nPrediction Amount: " + amt + " Dero\nRound Pot: " + s + " Dero\n\nPredictions: " + c +
				"\nHigher Predictions: " + u + "\nLower Predictions: " + d + "\n\nPayout After: " + end_pay.String() + "\nRefund if not paid within " + rf + " minutes\n\nRounds Completed: " + r
		} else {
			info = "SCID:\n\n" + Predict.Contract + live + node +
				"\n\nCloses at: " + utc + "\nMark: " + m + "\n\nPrediction Amount: " + amt + " Dero\nRound Pot: " + s + " Dero\n\nPredictions: " + c +
				"\nHigher Predictions: " + u + "\nLower Predictions: " + d + "\n\nPayout After: " + end_pay.String() + "\nRefund if not paid within " + rf + " minutes\n\nRounds Completed: " + r
		}
	}

	return
}

// Format dPrediction end result text
//   - fr is the un-split result string
//   - m is prediction mark
func roundResults(fr, m string) string {
	if len(Predict.Contract) == 64 && fr != "" {
		split := strings.Split(fr, "_")
		var res string
		var def string
		var x float64

		if isOnChainPrediction(split[0]) {
			switch onChainPrediction(split[0]) {
			case 1:
				x = 1
			case 2:
				x = 100000
			case 3:
				x = 1
			default:
				x = 1
			}
		} else {
			if menu.CoinDecimal(split[0]) == 8 {
				x = 100000000
			} else {
				x = 100
			}
		}

		if mark, err := strconv.ParseFloat(m, 64); err == nil {
			if rpc.StringToInt(split[1]) > int(mark*x) {
				res = "Higher "
				def = " > "
			} else if rpc.StringToInt(split[1]) == int(mark*x) {
				res = "Equal "
				def = " == "
			} else {
				res = "Lower "
				def = " < "
			}
		}

		if final, err := strconv.ParseFloat(split[1], 64); err == nil {
			var fStr string
			if isOnChainPrediction(split[0]) {
				switch onChainPrediction(split[0]) {
				case 1:
					fStr = fmt.Sprintf("%.0f", final/x)
				case 2:
					fStr = fmt.Sprintf("%.5f", final/x)
				case 3:
					fStr = fmt.Sprintf("%.0f", final/x)
				default:

				}
			} else {
				if menu.CoinDecimal(split[0]) == 8 {
					fStr = fmt.Sprintf("%.8f", final/x)
				} else {
					fStr = fmt.Sprintf("%.2f", final/x)
				}
			}

			return split[0] + " " + res + fStr + def + m
		}

	}
	return ""
}

// Formats non-initialized dPrediction info string
//   - fr is the un-split result string
//   - tx is the previous payout TXID
//   - r is total completed prediction rounds
//   - m is prediction mark
func P_no_initResults(fr, tx, r, m string) (info string) {
	info = "SCID:\n\n" + Predict.Contract + "\n" + "\nRound Completed\n\nRound Mark: " + m +
		"\nRound Results: " + roundResults(fr, m) + "\n\nPayout TXID: " + tx + "\n\nRounds Completed: " + r

	Predict.Prediction = ""

	return
}

// Populate all dReams dPrediction contracts
//   - Pass contracts from db store, can be nil arg
func PopulatePredictions(contracts map[string]string) {
	if rpc.Daemon.IsConnected() && menu.Gnomes.IsReady() {
		list := []string{}
		owned := []string{}
		if contracts == nil {
			contracts = menu.Gnomes.GetAllOwnersAndSCIDs()
		}
		keys := make([]string, len(contracts))

		i := 0
		for k := range contracts {
			keys[i] = k
			list, owned = checkBetContract(keys[i], "p", list, owned)
			i++
		}

		t := len(list)
		list = append(list, " Contracts: "+strconv.Itoa(t))
		sort.Strings(list)
		Predict.Settings.Contracts = list

		sort.Strings(owned)
		Predict.Settings.Owned = owned

	}
}

// Check if dPrediction is live on SCID
func CheckActivePrediction(scid string) bool {
	if len(scid) == 64 && menu.Gnomes.IsReady() {
		_, ends := menu.Gnomes.GetSCIDValuesByKey(scid, "p_end_at")
		_, buff := menu.Gnomes.GetSCIDValuesByKey(scid, "buffer")
		if ends != nil && buff != nil {
			now := time.Now().Unix()
			if now < int64(ends[0]) && now > int64(buff[0]) {
				return true
			}
		}
	}
	return false
}

// Gets dPrediction data from SCID and return formatted info string
func GetPrediction(scid string) (info string) {
	if rpc.Daemon.IsConnected() && menu.Gnomes.IsReady() {
		predicting, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "predicting")
		url, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "p_url")
		final, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "p_final")
		//final_tx, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "p_final_txid")
		_, amt := menu.Gnomes.GetSCIDValuesByKey(scid, "p_amount")
		_, init := menu.Gnomes.GetSCIDValuesByKey(scid, "p_init")
		_, up := menu.Gnomes.GetSCIDValuesByKey(scid, "p_up")
		_, down := menu.Gnomes.GetSCIDValuesByKey(scid, "p_down")
		_, count := menu.Gnomes.GetSCIDValuesByKey(scid, "p_#")
		_, end := menu.Gnomes.GetSCIDValuesByKey(scid, "p_end_at")
		_, buffer := menu.Gnomes.GetSCIDValuesByKey(scid, "buffer")
		_, pot := menu.Gnomes.GetSCIDValuesByKey(scid, "p_total")
		_, rounds := menu.Gnomes.GetSCIDValuesByKey(scid, "p_played")
		_, mark := menu.Gnomes.GetSCIDValuesByKey(scid, "mark")
		_, time_a := menu.Gnomes.GetSCIDValuesByKey(scid, "time_a")
		_, time_b := menu.Gnomes.GetSCIDValuesByKey(scid, "time_b")
		_, time_c := menu.Gnomes.GetSCIDValuesByKey(scid, "time_c")

		var pre, p_played, p_final, p_mark string
		if init != nil {
			if init[0] == 1 {
				Predict.Init = true

				if buffer != nil {
					Predict.Buffer = int64(buffer[0])
				}

				Predict.Amount = amt[0]
				if predicting != nil {
					pre = predicting[0]
				}
				Predict.Prediction = pre

				p_amt := fmt.Sprint(float64(Predict.Amount) / 100000)

				p_down := fmt.Sprint(down[0])
				p_up := fmt.Sprint(up[0])
				p_count := fmt.Sprint(count[0])

				p_pot := fmt.Sprint(pot[0])
				p_played := fmt.Sprint(rounds[0])

				var p_feed string
				if url != nil {
					Predict.Feed = url[0]
					p_feed = url[0]
				}

				if mark != nil {
					if predicting != nil {
						if isOnChainPrediction(predicting[0]) {
							i := onChainPrediction(predicting[0])
							switch i {
							case 1:
								p_mark = fmt.Sprintf("%d", mark[0])
							case 2:
								p_mark = fmt.Sprintf("%.5f", float64(mark[0])/100000)
							case 3:
								p_mark = fmt.Sprintf("%d", mark[0])
							}
						} else {
							if menu.CoinDecimal(pre) == 8 {
								p_mark = fmt.Sprintf("%.8f", float64(mark[0])/100000000)
							} else {
								p_mark = fmt.Sprintf("%.2f", float64(mark[0])/100)
							}
						}
					}
				} else {
					p_mark = "0"
				}

				var p_end string
				if init[0] == 1 {
					end_at := uint(end[0])
					p_end = fmt.Sprint(end_at * 1000)
				}

				info = P_initResults(pre, p_amt, p_end, p_count, p_pot, p_up, p_down, p_played, p_feed, p_mark, int(time_a[0]), int(time_b[0]), int(time_c[0]))

			} else {
				if final != nil {
					p_final = final[0]
				}

				Predict.Init = false
				txid := FetchPredictionFinal(scid)

				if mark != nil {
					split := strings.Split(p_final, "_")
					if isOnChainPrediction(split[0]) {
						i := onChainPrediction(split[0])

						switch i {
						case 1:
							p_mark = fmt.Sprintf("%d", mark[0])
						case 2:
							p_mark = fmt.Sprintf("%.5f", float64(mark[0])/100000)
						case 3:
							p_mark = fmt.Sprintf("%d", mark[0])
						}
					} else {
						if menu.CoinDecimal(split[0]) == 8 {
							p_mark = fmt.Sprintf("%.8f", float64(mark[0])/100000000)
						} else {
							p_mark = fmt.Sprintf("%.2f", float64(mark[0])/100)
						}
					}

				} else {
					p_mark = "0"
				}

				if rounds != nil {
					p_played = fmt.Sprint(rounds[0])
				}

				info = P_no_initResults(p_final, txid, p_played, p_mark)
			}
		}
	}

	return
}
