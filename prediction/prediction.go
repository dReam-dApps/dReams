package prediction

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/table"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type predictItems struct {
	Contract    string
	Leaders_map map[string]uint64
	//Leaders_display []string
	Info          *widget.Label
	Prices        *widget.Label
	Predict_list  *widget.List
	Favorite_list *widget.List
	Owned_list    *widget.List
	Leaders_list  *widget.List
	Remove_button *widget.Button
}

var PredictControl predictItems

func PredictConnectedBox() fyne.Widget {
	menu.MenuControl.Predict_check = widget.NewCheck("", func(b bool) {
		if !b {
			// prediction leaderboard
			// table.Actions.NameEntry.Hide()
			// table.Actions.Change.Hide()
			//PredictControl.Leaders_display = []string{}
			table.Actions.Higher.Hide()
			table.Actions.Lower.Hide()

		}
	})
	menu.MenuControl.Predict_check.Disable()

	return menu.MenuControl.Predict_check
}

func PreictionContractEntry() fyne.Widget {
	options := []string{}
	table.Actions.P_contract = widget.NewSelectEntry(options)
	table.Actions.P_contract.PlaceHolder = "Contract Address: "
	table.Actions.P_contract.OnCursorChanged = func() {
		if rpc.Signal.Daemon {
			go func() {
				if len(PredictControl.Contract) == 64 {
					yes, _ := rpc.ValidBetContract(PredictControl.Contract)
					if yes {
						menu.MenuControl.Predict_check.SetChecked(true)
					} else {
						menu.MenuControl.Predict_check.SetChecked(false)
					}
				} else {
					menu.MenuControl.Predict_check.SetChecked(false)
				}
			}()
		}
	}

	this := binding.BindString(&PredictControl.Contract)
	table.Actions.P_contract.Bind(this)

	return table.Actions.P_contract
}

func PredictBox() fyne.CanvasObject {
	// prediction leaderboard
	// table.Actions.NameEntry = widget.NewEntry()
	// table.Actions.NameEntry.SetPlaceHolder("Name")
	// table.Actions.NameEntry.OnChanged = func(input string) {
	// 	table.Actions.NameEntry.Validator = validation.NewRegexp(`\w{3,}`, "Three Letters Minimum")
	// 	table.Actions.NameEntry.Validate()
	// 	table.Actions.NameEntry.Refresh()
	// }
	//
	// table.Actions.Change = widget.NewButton("Change Name", func() {
	// 	if table.Actions.NameEntry.Disabled() {
	// 		table.Actions.NameEntry.Enable()
	// 	} else {
	// 		namePopUp(1)
	// 	}
	// })

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

	// table.Actions.NameEntry.Hide()
	// table.Actions.Change.Hide()
	table.Actions.Higher.Hide()
	table.Actions.Lower.Hide()

	table.Actions.Prediction_box = container.NewVBox(table.Actions.Higher, table.Actions.Lower)
	table.Actions.Prediction_box.Hide()

	return table.Actions.Prediction_box
}

// prediction leaderboard
// func LeadersDisplay() fyne.Widget {
// 	PredictControl.Leaders_display = []string{}
// 	PredictControl.Leaders_list = widget.NewList(
// 		func() int {
// 			return len(PredictControl.Leaders_display)
// 		},
// 		func() fyne.CanvasObject {
// 			return widget.NewLabel("")
// 		},
// 		func(i widget.ListItemID, o fyne.CanvasObject) {
// 			o.(*widget.Label).SetText(PredictControl.Leaders_display[i])
// 		})
//
// 	return PredictControl.Leaders_list
// }

func ShowPredictionControls() {
	menu.DisablePreditions(false)
	table.Actions.Higher.Show()
	table.Actions.Lower.Show()
	// prediction leaderboard
	// table.Actions.NameEntry.Show()
	// table.Actions.NameEntry.Text = menu.CheckPredictionName(PredictControl.Contract)
	// table.Actions.NameEntry.Refresh()
}

func setPredictionControls(str string) (item string) {
	split := strings.Split(str, "   ")
	if len(split) >= 3 {
		trimmed := strings.Trim(split[2], " ")
		if len(trimmed) == 64 {
			item = str
			table.Actions.P_contract.SetText(trimmed)
			go SetPredictionInfo(trimmed)
			if menu.CheckActivePrediction(trimmed) {
				ShowPredictionControls()
			} else {
				menu.DisablePreditions(true)
			}
		}
	}

	return
}

func SetPredictionInfo(scid string) {
	info := GetPrediction(rpc.Signal.Daemon, scid)
	if info != "" {
		PredictControl.Info.SetText(info)
		PredictControl.Info.Refresh()
	}
}

func SetPredictionPrices(d bool) {
	if d {
		_, btc := table.GetPrice("BTC-USDT")
		_, dero := table.GetPrice("DERO-USDT")
		_, xmr := table.GetPrice("XMR-USDT")
		/// custom feed with rpc.Display.P_feed
		prices := "Current Price feed from dReams Client\nBTC: " + btc + "\nDERO: " + dero + "\nXMR: " + xmr

		PredictControl.Prices.SetText(prices)
	}
}

func PredictionListings() fyne.CanvasObject { /// prediction contract list
	PredictControl.Predict_list = widget.NewList(
		func() int {
			return len(menu.MenuControl.Predict_contracts)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(canvas.NewImageFromImage(nil), widget.NewLabel(""))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*fyne.Container).Objects[1].(*widget.Label).SetText(menu.MenuControl.Predict_contracts[i])
			if menu.MenuControl.Predict_contracts[i][0:2] != "  " {
				var key string
				split := strings.Split(menu.MenuControl.Predict_contracts[i], "   ")
				if len(split) >= 3 {
					trimmed := strings.Trim(split[2], " ")
					if len(trimmed) == 64 {
						key = trimmed
					}
				}

				badge := canvas.NewImageFromResource(menu.DisplayRating(menu.MenuControl.Contract_rating[key]))
				badge.SetMinSize(fyne.NewSize(35, 35))
				o.(*fyne.Container).Objects[0] = badge
			}
		})

	var item string

	PredictControl.Predict_list.OnSelected = func(id widget.ListItemID) {
		if id != 0 && menu.Connected() {
			item = setPredictionControls(menu.MenuControl.Predict_contracts[id])
			PredictControl.Favorite_list.UnselectAll()
			PredictControl.Owned_list.UnselectAll()
		} else {
			menu.DisablePreditions(true)
		}
	}

	save := widget.NewButton("Favorite", func() {
		menu.MenuControl.Predict_favorites = append(menu.MenuControl.Predict_favorites, item)
		sort.Strings(menu.MenuControl.Predict_favorites)
	})

	rate := widget.NewButton("Rate", func() {
		if len(PredictControl.Contract) == 64 {
			if !menu.CheckOwner(PredictControl.Contract) {
				menu.RateConfirm(PredictControl.Contract)
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
		PredictControl.Predict_list)

	return cont
}

func PredicitionFavorites() fyne.CanvasObject {
	PredictControl.Favorite_list = widget.NewList(
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

	PredictControl.Favorite_list.OnSelected = func(id widget.ListItemID) {
		if menu.Connected() {
			item = setPredictionControls(menu.MenuControl.Predict_favorites[id])
			PredictControl.Predict_list.UnselectAll()
			PredictControl.Owned_list.UnselectAll()
		} else {
			menu.DisablePreditions(true)
		}
	}

	remove := widget.NewButton("Remove", func() {
		if len(menu.MenuControl.Predict_favorites) > 0 {
			PredictControl.Favorite_list.UnselectAll()
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
		PredictControl.Favorite_list.Refresh()
		sort.Strings(menu.MenuControl.Predict_favorites)
	})

	cont := container.NewBorder(
		nil,
		container.NewBorder(nil, nil, nil, remove, layout.NewSpacer()),
		nil,
		nil,
		PredictControl.Favorite_list)

	return cont
}

func PredictionOwned() fyne.CanvasObject {
	PredictControl.Owned_list = widget.NewList(
		func() int {
			return len(menu.MenuControl.Predict_owned)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(menu.MenuControl.Predict_owned[i])
		})

	PredictControl.Owned_list.OnSelected = func(id widget.ListItemID) {
		if menu.Connected() {
			setPredictionControls(menu.MenuControl.Predict_owned[id])
			PredictControl.Predict_list.UnselectAll()
			PredictControl.Favorite_list.UnselectAll()
		} else {
			menu.DisablePreditions(true)
		}

	}

	return PredictControl.Owned_list
}

// prediction leaderboard
// func Remove() fyne.Widget {
// 	PredictControl.Remove_button = widget.NewButton("Remove", func() {
// 		namePopUp(2)
// 	})
//
// 	PredictControl.Remove_button.Hide()
//
// 	return PredictControl.Remove_button
// }

func P_initResults(p, amt, eA, c, to, u, d, r, f, m string, ta, tb, tc int) (info string) { /// prediction info, initialized
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
			info = "SCID: \n" + PredictControl.Contract + "\n\n" + p + wfp + "\n\nNode: " + f + "\n\nMark: " + mark + "\nRound Pot: " + s +
				"\n\nPredictions: " + c + "\nHigher Predictions: " + u + "\nLower Predictions: " + d +
				"\n\nPayout After: " + end_pay.String() + "\nRefund if not paid within " + rf + " minutes\n\nRounds Completed: " + r
		} else {
			info = "SCID: \n" + PredictControl.Contract + "\n\n" + p + wfp + "\n\nMark: " + mark + "\nRound Pot: " + s +
				"\n\nPredictions: " + c + "\nHigher Predictions: " + u + "\nLower Predictions: " + d +
				"\n\nPayout After: " + end_pay.String() + "\nRefund if not paid within " + rf + " minutes\n\nRounds Completed: " + r
		}

	} else {
		var live string
		if now > rpc.Predict.Buffer {
			live = "\n\nAccepting " + p + " Predictions "
		} else {
			left := rpc.Predict.Buffer - now
			live = "\n\n" + p + "\nBuffer ends in " + strconv.Itoa(int(left)) + " seconds"
		}

		node := ""
		if isOnChainPrediction(p) {
			node = "\n\nNode: " + f
		}

		if m == "0" {
			pw := strconv.Itoa(ta / 60)
			info = "SCID: \n" + PredictControl.Contract + live + node +
				"\n\nCloses at: " + utc + "\nMark posted with in " + pw + " minutes of close\n\nPrediction Amount: " + amt + " Dero\nRound Pot: " + s + " \n\nPredictions: " + c +
				"\nHigher Predictions: " + u + "\nLower Predictions: " + d + "\n\nPayout After: " + end_pay.String() + "\nRefund if not paid within " + rf + " minutes\n\nRounds Completed: " + r
		} else {
			info = "SCID: \n" + PredictControl.Contract + live + node +
				"\n\nCloses at: " + utc + "\nMark: " + m + "\n\nPrediction Amount: " + amt + " Dero\nRound Pot: " + s + "\n\nPredictions: " + c +
				"\nHigher Predictions: " + u + "\nLower Predictions: " + d + "\n\nPayout After: " + end_pay.String() + "\nRefund if not paid within " + rf + " minutes\n\nRounds Completed: " + r
		}
	}

	return
}

func roundResults(fr, m string) string { /// prediction results text
	if len(PredictControl.Contract) == 64 && fr != "" {
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
			if table.CoinDecimal(split[0]) == 8 {
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
				if table.CoinDecimal(split[0]) == 8 {
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

func P_no_initResults(fr, tx, r, m string) (info string) { /// prediction info, not initialized
	info = "SCID: \n" + PredictControl.Contract + "\n" + "\nRound Completed\n\nRound Mark: " + m +
		"\nRound Results: " + roundResults(fr, m) + "\n\nPayout TXID: " + tx + "\n\nRounds Completed: " + r

	rpc.Display.Prediction = ""

	return
}

func GetPrediction(d bool, scid string) (info string) {
	if d && menu.Gnomes.Init && !menu.GnomonClosing() && menu.Gnomes.Sync {
		predicting, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "predicting", menu.Gnomes.Indexer.ChainHeight, true)
		url, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_url", menu.Gnomes.Indexer.ChainHeight, true)
		final, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_final", menu.Gnomes.Indexer.ChainHeight, true)
		//final_tx, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_final_txid", menu.Gnomes.Indexer.ChainHeight, true)
		_, amt := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_amount", menu.Gnomes.Indexer.ChainHeight, true)
		_, init := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_init", menu.Gnomes.Indexer.ChainHeight, true)
		_, up := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_up", menu.Gnomes.Indexer.ChainHeight, true)
		_, down := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_down", menu.Gnomes.Indexer.ChainHeight, true)
		_, count := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_#", menu.Gnomes.Indexer.ChainHeight, true)
		_, end := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_end_at", menu.Gnomes.Indexer.ChainHeight, true)
		_, buffer := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "buffer", menu.Gnomes.Indexer.ChainHeight, true)
		_, pot := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_total", menu.Gnomes.Indexer.ChainHeight, true)
		_, rounds := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_played", menu.Gnomes.Indexer.ChainHeight, true)
		_, mark := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "mark", menu.Gnomes.Indexer.ChainHeight, true)
		_, time_a := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "time_a", menu.Gnomes.Indexer.ChainHeight, true)
		_, time_b := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "time_b", menu.Gnomes.Indexer.ChainHeight, true)
		_, time_c := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "time_c", menu.Gnomes.Indexer.ChainHeight, true)

		var pre, p_played, p_final, p_mark string
		if init != nil {
			if init[0] == 1 {
				rpc.Predict.Init = true

				if buffer != nil {
					rpc.Predict.Buffer = int64(buffer[0])
				}

				rpc.Predict.Amount = amt[0]
				if predicting != nil {
					pre = predicting[0]
				}
				rpc.Display.Prediction = pre

				p_amt := fmt.Sprint(float64(rpc.Predict.Amount) / 100000)

				p_down := fmt.Sprint(down[0])
				p_up := fmt.Sprint(up[0])
				p_count := fmt.Sprint(count[0])

				p_pot := fmt.Sprint(pot[0])
				p_played := fmt.Sprint(rounds[0])

				var p_feed string
				if url != nil {
					rpc.Display.P_feed = url[0]
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
							if table.CoinDecimal(pre) == 8 {
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

				rpc.Predict.Init = false
				txid, _ := rpc.FetchPredictionFinal(rpc.Signal.Daemon, scid)

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
						if table.CoinDecimal(split[0]) == 8 {
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
