package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/prediction"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/table"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Notification struct {
	Title, Content string
}

type save struct {
	Name   string   `json:"name"`
	Daemon []string `json:"daemon"`
}

var offset int

func notification(title, content string, g int) *fyne.Notification {
	switch g {
	case 0:
		rpc.Round.Notified = true
	case 1:
		rpc.Bacc.Notified = true
	}

	return &fyne.Notification{Title: title, Content: content}
}

func isWindows() bool {
	return dReams.os == "windows"
}

func systemTray(w fyne.App) bool {
	if desk, ok := w.(desktop.App); ok {
		m := fyne.NewMenu("MyApp",
			fyne.NewMenuItem("Show", func() {
				log.Println("Tapped show")
			}),
			fyne.NewMenuItem("Reveal Key", func() {
				rpc.RevealKey(rpc.Wallet.ClientKey)
			}))
		desk.SetSystemTrayMenu(m)
		return true
	}
	return false
}

func labelColorBlack(c *fyne.Container) *fyne.Container {
	back := canvas.NewRectangle(color.RGBA{0, 0, 0, 150})
	cont := container.New(layout.NewMaxLayout(), back, c)
	return cont
}

func makeConfig(name, daemon string) (data save) {
	data.Name = name
	switch daemon {
	case "127.0.0.1:10102":
	case "89.38.99.117:10102":
	case "dero-node.mysrv.cloud:10102":
	case "derostats.io:10102":
	default:
		data.Daemon = []string{daemon}
	}
	return data
}

func writeConfig(u save) {
	if u.Daemon != nil && u.Name != "" {
		if u.Daemon[0] != "" {
			file, err := os.Create("config.json")

			if err != nil {
				log.Println(err)
			}

			defer file.Close()
			json, _ := json.MarshalIndent(u, "", " ")

			_, err = file.Write(json)

			if err != nil {
				log.Println("Error writing config file: ", err)
			}
		}
	}
}

func readConfig() (string, string) {
	if !table.FileExists("config.json") {
		return "", ""
	}

	file, err := os.ReadFile("config.json")

	if err != nil {
		log.Println("Error reading config file: ", err)
		return "", ""
	}

	var config save
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Println("Error during unmarshal: ", err)
		return "", ""
	}

	return config.Name, config.Daemon[0]
}

func showBaccCards() *fyne.Container {
	var drawP, drawB int
	if rpc.Bacc.P_card3 == 0 {
		drawP = 99
	} else {
		drawP = rpc.Bacc.P_card3
	}

	if rpc.Bacc.B_card3 == 0 {
		drawB = 99
	} else {
		drawB = rpc.Bacc.B_card3
	}
	w := table.Settings.FaceSelect.SelectedIndex()

	content := *container.NewWithoutLayout(
		PlayerCards(w, BaccSuit(rpc.Bacc.P_card1), BaccSuit(rpc.Bacc.P_card2), BaccSuit(drawP)),
		BankerCards(w, BaccSuit(rpc.Bacc.B_card1), BaccSuit(rpc.Bacc.B_card2), BaccSuit(drawB)),
	)

	rpc.Bacc.Display = true
	table.BaccBuffer(false)

	return &content
}

func clearBaccCards() *fyne.Container {
	content := *container.NewWithoutLayout(
		PlayerCards(0, 99, 99, 99),
		BankerCards(0, 99, 99, 99))

	return &content
}

func showHolderoCards(l1, l2 string) *fyne.Container {
	size := dReams.Window.Content().Size()
	content := *container.NewWithoutLayout(
		Hole_1(Card(l1), size.Width, size.Height),
		Hole_2(Card(l2), size.Width, size.Height),
		P1_a(Is_In(rpc.CardHash.P1C1, 1, rpc.Signal.End)),
		P1_b(Is_In(rpc.CardHash.P1C2, 1, rpc.Signal.End)),
		P2_a(Is_In(rpc.CardHash.P2C1, 2, rpc.Signal.End)),
		P2_b(Is_In(rpc.CardHash.P2C2, 2, rpc.Signal.End)),
		P3_a(Is_In(rpc.CardHash.P3C1, 3, rpc.Signal.End)),
		P3_b(Is_In(rpc.CardHash.P3C2, 3, rpc.Signal.End)),
		P4_a(Is_In(rpc.CardHash.P4C1, 4, rpc.Signal.End)),
		P4_b(Is_In(rpc.CardHash.P4C2, 4, rpc.Signal.End)),
		P5_a(Is_In(rpc.CardHash.P5C1, 5, rpc.Signal.End)),
		P5_b(Is_In(rpc.CardHash.P5C2, 5, rpc.Signal.End)),
		P6_a(Is_In(rpc.CardHash.P6C1, 6, rpc.Signal.End)),
		P6_b(Is_In(rpc.CardHash.P6C2, 6, rpc.Signal.End)),
		Flop_1(rpc.Round.Flop1),
		Flop_2(rpc.Round.Flop2),
		Flop_3(rpc.Round.Flop3),
		River(rpc.Round.TurnCard),
		Turn(rpc.Round.RiverCard))

	return &content
}

func ifBet(w, r uint64) { /// sets bet amount on turn
	if w > 0 && r > 0 && !rpc.Signal.PlacedBet {
		float := float64(w) / 100000
		wager := strconv.FormatFloat(float, 'f', 1, 64)
		table.Actions.BetEntry.SetText(wager)
		rpc.Display.Res = "Bet Raised, " + wager + " to Call "
	} else if w > 0 && !rpc.Signal.PlacedBet {
		float := float64(w) / 100000
		wager := strconv.FormatFloat(float, 'f', 1, 64)
		table.Actions.BetEntry.SetText(wager)
		rpc.Display.Res = "Bet is " + wager
	} else if r > 0 && rpc.Signal.PlacedBet {
		float := float64(r) / 100000
		rasied := strconv.FormatFloat(float, 'f', 1, 64)
		table.Actions.BetEntry.SetText(rasied)
		rpc.Display.Res = "Bet Raised, " + rasied + " to Call"
	} else if w == 0 && !rpc.Signal.Bet {
		float := float64(rpc.Round.BB) / 100000
		this := strconv.FormatFloat(float, 'f', 1, 64)
		table.Actions.BetEntry.SetText(this)
		if !rpc.Signal.Reveal {
			rpc.Display.Res = "Check or Bet"
			table.Actions.BetEntry.Enable()
		}
	} else if !rpc.Signal.Deal {
		rpc.Display.Res = "Deal Hand"
	}

	table.Actions.BetEntry.Refresh()
}

func singleShot(turn, trigger bool) bool {
	if turn && !trigger {
		ifBet(rpc.Round.Wager, rpc.Round.Raised)
		return true
	}

	if !turn {
		return false
	} else {
		return turn
	}
}

func fetch(quit chan struct{}) { /// main loop
	time.Sleep(3 * time.Second)
	var ticker = time.NewTicker(3 * time.Second)
	var trigger bool
	var skip int
	for {
		select {
		case <-ticker.C: /// do on interval
			rpc.Ping()
			rpc.GetBalance(rpc.Wallet.Connect)
			rpc.DreamsBalance(rpc.Wallet.Connect)
			rpc.GetHeight(rpc.Wallet.Connect)
			if !rpc.Signal.Startup {
				menu.CheckConnection()
				rpc.FetchHolderoSC(rpc.Signal.Daemon, rpc.Signal.Contract)
				rpc.FetchBaccSC(rpc.Signal.Daemon)
				rpc.FetchPredictionSC(rpc.Signal.Daemon, prediction.PredictControl.Contract)
				menu.GnomonState(rpc.Signal.Daemon, menu.Gnomes.Init)
				background.Refresh()

				offset++
				if offset == 21 {
					offset = 0
				} else if offset%3 == 0 {
					SportsRefresh(dReams.sports)
				}

				if (rpc.StringToInt(rpc.Display.Turn) == rpc.Round.ID && rpc.StringToInt(rpc.Wallet.Height) > rpc.Signal.CHeight+3) ||
					(rpc.StringToInt(rpc.Display.Turn) != rpc.Round.ID && rpc.Round.ID >= 1) || (!rpc.Signal.My_turn && rpc.Round.ID >= 1) {
					if rpc.Signal.Clicked {
						trigger = false
					}
					rpc.Signal.Clicked = false
				}

				BaccRefresh()
				PredictionRefresh(dReams.predict)

				go MenuRefresh(dReams.menu, menu.Gnomes.Init)
				if !rpc.Signal.Clicked {
					setHolderoLabel()
					table.GetUrls(rpc.Round.F_url, rpc.Round.B_url)
					rpc.Called(rpc.Round.Flop, rpc.Round.Wager)
					trigger = singleShot(rpc.Signal.My_turn, trigger)
					HolderoRefresh()
					skip = 0
				} else {
					waitLabel()
					skip++
					if skip >= 18 {
						rpc.Signal.Clicked = false
						skip = 0
						trigger = false
					}
				}
			}
			if rpc.Signal.Daemon {
				rpc.Signal.Startup = false
			}

		case <-quit: /// exit loop
			log.Println("Closing dReams.")
			ticker.Stop()
			return
		}
	}
}

func setHolderoLabel() {
	H.TopLabel.SetText(rpc.Display.Res)
	H.LeftLabel.SetText("Seats: " + rpc.Display.Seats + "      Pot: " + rpc.Display.Pot + "      Blinds: " + rpc.Display.Blinds + "      Ante: " + rpc.Display.Ante + "      Dealer: " + rpc.Display.Dealer + "      Turn: " + rpc.Display.Turn)
	if rpc.Round.Asset {
		H.RightLabel.SetText(rpc.Display.Readout + "      Player ID: " + rpc.Display.PlayerId + "      dReams Balance: " + rpc.Wallet.TokenBal + "      Height: " + rpc.Wallet.Height)
	} else {
		H.RightLabel.SetText(rpc.Display.Readout + "      Player ID: " + rpc.Display.PlayerId + "      Dero Balance: " + rpc.Wallet.Balance + "      Height: " + rpc.Wallet.Height)
	}

	if rpc.Signal.Contract {
		table.Settings.SharedOn.Enable()
	} else {
		table.Settings.SharedOn.Disable()
	}

	H.TopLabel.Refresh()
	H.LeftLabel.Refresh()
	H.RightLabel.Refresh()
}

func waitLabel() {
	H.TopLabel.SetText("")
	if rpc.Round.Asset {
		H.RightLabel.SetText("Wait for Block" + "      Player ID: " + rpc.Display.PlayerId + "      dReams Balance: " + rpc.Wallet.TokenBal + "      Height: " + rpc.Wallet.Height)
	} else {
		H.RightLabel.SetText("Wait for Block" + "      Player ID: " + rpc.Display.PlayerId + "      Dero Balance: " + rpc.Wallet.Balance + "      Height: " + rpc.Wallet.Height)
	}
	H.TopLabel.Refresh()
	H.RightLabel.Refresh()
}

func HolderoRefresh() {
	go table.ShowAvatar(dReams.holdero)
	H.CardsContent = *container.NewWithoutLayout(showHolderoCards(rpc.CardHash.Local1, rpc.CardHash.Local2))
	if !rpc.Signal.Clicked {
		if rpc.Round.ID == 0 && rpc.Wallet.Connect {
			if rpc.Signal.Sit {
				table.Actions.Sit.Hide()
			} else {
				table.Actions.Sit.Show()
			}
			table.Actions.Leave.Hide()
			table.Actions.Deal.Hide()
			table.Actions.Check.Hide()
			table.Actions.Bet.Hide()
			table.Actions.BetEntry.Hide()
		} else if !rpc.Signal.End && !rpc.Signal.Reveal && rpc.Signal.My_turn {
			if rpc.Signal.Sit {
				table.Actions.Sit.Hide()
			} else {
				table.Actions.Sit.Show()
			}

			if rpc.Signal.Leave {
				table.Actions.Leave.Hide()
			} else {
				table.Actions.Leave.Show()
			}

			if rpc.Signal.Deal {
				table.Actions.Deal.Hide()
			} else {
				table.Actions.Deal.Show()
			}

			table.Actions.Check.SetText(rpc.Display.C_Button)
			table.Actions.Bet.SetText(rpc.Display.B_Button)
			if rpc.Signal.Bet {
				table.Actions.Check.Hide()
				table.Actions.Bet.Hide()
				table.Actions.BetEntry.Hide()
			} else {
				table.Actions.Check.Show()
				table.Actions.Bet.Show()
				table.Actions.BetEntry.Show()
			}

			if !rpc.Round.Notified {
				if !isWindows() {
					dReams.App.SendNotification(notification("dReams - Holdero", "Your Turn", 0))
				}
			}
		} else {
			if rpc.Signal.Sit {
				table.Actions.Sit.Hide()
			} else {
				table.Actions.Sit.Show()
			}
			table.Actions.Leave.Hide()
			table.Actions.Deal.Hide()
			table.Actions.Check.Hide()
			table.Actions.Bet.Hide()
			table.Actions.BetEntry.Hide()

			if rpc.Signal.Reveal && rpc.Signal.My_turn && !rpc.Signal.End {
				if !rpc.Round.Notified {
					rpc.Display.Res = "Revealing Key"
					if !isWindows() {
						dReams.App.SendNotification(notification("dReams - Holdero", "Revealing Key", 0))
					}
				}
			}

			if !rpc.Signal.My_turn && !rpc.Signal.End && !rpc.Round.LocalEnd {
				rpc.Display.Res = ""
				rpc.Round.Notified = false
			}
		}
	}

	go func() {
		H.TableContent = *container.NewWithoutLayout(
			table.HolderoTable(resourceTablePng),
			table.Player1_label(resourceUnknownPng, resourceAvatarFramePng, resourceTurnFramePng),
			table.Player2_label(resourceUnknownPng, resourceAvatarFramePng, resourceTurnFramePng),
			table.Player3_label(resourceUnknownPng, resourceAvatarFramePng, resourceTurnFramePng),
			table.Player4_label(resourceUnknownPng, resourceAvatarFramePng, resourceTurnFramePng),
			table.Player5_label(resourceUnknownPng, resourceAvatarFramePng, resourceTurnFramePng),
			table.Player6_label(resourceUnknownPng, resourceAvatarFramePng, resourceTurnFramePng),
			H.TopLabel,
		)

		H.TableContent.Refresh()
		H.TableItems.Refresh()
	}()
}

func BaccRefresh() {
	B.LeftLabel.SetText("Total Hands Played: " + rpc.Display.Total_w + "      Player Wins: " + rpc.Display.Player_w + "      Ties: " + rpc.Display.Ties + "      Banker Wins: " + rpc.Display.Banker_w + "      Min Bet is " + rpc.Display.BaccMin + " dReam, Max Bet is " + rpc.Display.BaccMax)
	B.RightLabel.SetText("dReams Balance: " + rpc.Wallet.TokenBal + "      Dero Balance: " + rpc.Wallet.Balance + "      Height: " + rpc.Wallet.Height)

	if !rpc.Bacc.Display {
		B.CardsContent = *container.NewWithoutLayout(clearBaccCards())
		rpc.FetchBaccHand(rpc.Signal.Daemon, rpc.Bacc.Last)
		if rpc.Bacc.Found {
			B.CardsContent = *container.NewWithoutLayout(showBaccCards())
		}
	}

	if rpc.StringToInt(rpc.Wallet.Height) > rpc.Bacc.CHeight+3 {
		table.BaccBuffer(false)
	}

	B.TableContent = *container.NewWithoutLayout(
		table.BaccTable(resourceBaccTablePng),
		table.BaccResult(rpc.Display.BaccRes),
	)

	B.TableContent.Refresh()
	B.TableItems.Refresh()

	if rpc.Bacc.Found && !rpc.Bacc.Notified {
		if !isWindows() {
			dReams.App.SendNotification(notification("dReams - Baccarat", rpc.Display.BaccRes, 1))
		}
	}
}

func PredictionRefresh(tab bool) {
	if tab {
		if rpc.Predict.Init {
			if rpc.Predict.Mark {
				P_initResults(rpc.Display.Preiction, rpc.Display.P_amt, rpc.Display.P_end, rpc.Display.P_count, rpc.Display.P_pot, rpc.Display.P_up, rpc.Display.P_down, rpc.Display.P_played, rpc.Display.P_feed, rpc.Display.P_mark, rpc.Predict.Time_a, rpc.Predict.Time_b, rpc.Predict.Time_c, true)
			} else {
				P_initResults(rpc.Display.Preiction, rpc.Display.P_amt, rpc.Display.P_end, rpc.Display.P_count, rpc.Display.P_pot, rpc.Display.P_up, rpc.Display.P_down, rpc.Display.P_played, rpc.Display.P_feed, "", rpc.Predict.Time_a, rpc.Predict.Time_b, rpc.Predict.Time_c, false)
			}

		} else {
			P_no_initResults(rpc.Display.P_final, rpc.Display.P_txid, rpc.Display.P_played, rpc.Display.P_mark)
		}

		P.RightLabel.SetText("dReams Balance: " + rpc.Wallet.TokenBal + "      Dero Balance: " + rpc.Wallet.Balance + "      Height: " + rpc.Wallet.Height)
	}
}

func SportsRefresh(tab bool) {
	if tab {
		S.RightLabel.SetText("dReams Balance: " + rpc.Wallet.TokenBal + "      Dero Balance: " + rpc.Wallet.Balance + "      Height: " + rpc.Wallet.Height)
		go GetBook(rpc.Signal.Daemon, prediction.SportsControl.Contract)
	}
}

func P_initResults(p, amt, eA, c, to, u, d, r, f, m string, ta, tb, tc int, post bool) { /// prediction info, initialized

	end_time, _ := rpc.MsToTime(eA)
	utc := end_time.String()
	add := rpc.StringToInt(eA)
	end := strconv.Itoa(add + (tc * 1000))
	end_pay, _ := rpc.MsToTime(end)
	rf := strconv.Itoa(tb / 60)

	result, err := strconv.ParseFloat(to, 32)

	if err != nil {
		log.Println("Float Conversion Error", err)
	}

	s := fmt.Sprintf("%.5f", result/100000)
	if post {
		P.TopLabel.SetText("SCID: \n" + prediction.PredictControl.Contract + "\n" + "\n" + p + " Price Posted" +
			"\nMark: " + m + "\nPredictions: " + c +
			"\nRound Pot: " + s + "\nUp Predictions: " + u + "\nDown Predictions: " + d + "\nPayout After: " + end_pay.String() + "\nRefund if not paid within " + rf + " minutes\nRounds Completed: " + r)
	} else {
		pw := strconv.Itoa(ta / 60)
		P.TopLabel.SetText("SCID: \n" + prediction.PredictControl.Contract + "\n" + "\nAccepting " + p + " Predictions " +
			"\nPrediction Amount: " + amt + " Dero\nCloses at: " + utc + "\nPrice posted with in " + pw + " minutes of close\nPredictions: " + c +
			"\nRound Pot: " + s + "\nHigher Predictions: " + u + "\nLower Predictions: " + d + "\nPayout After: " + end_pay.String() + "\nRefund if not paid within " + rf + " minutes\nRounds Completed: " + r)
	}

	if offset == 11 || P.BottomLabel.Text == "" {
		_, btc := table.GetPrice("BTC-USDT")
		_, dero := table.GetPrice("DERO-USDT")
		_, xmr := table.GetPrice("XMR-USDT")
		P.BottomLabel.SetText(f + "\nBTC: " + btc + "\nDERO: " + dero + "\nXMR: " + xmr)
	}
}

func roundResults(fr, m string) string { /// prediction results text
	if len(prediction.PredictControl.Contract) == 64 && fr != "" {
		split := strings.Split(fr, "_")
		var res string
		var def string

		if mark, err := strconv.ParseFloat(m, 64); err == nil {
			if rpc.StringToInt(split[1]) > int(mark*100) {
				res = "Higher "
				def = " > "
			} else if rpc.StringToInt(split[1]) == int(mark*100) {
				res = "Equal "
				def = " == "
			} else {
				res = "Lower "
				def = " < "
			}
		}

		if final, err := strconv.ParseFloat(split[1], 64); err == nil {
			fStr := fmt.Sprintf("%.2f", final/100)

			return split[0] + " " + res + fStr + def + m
		}

	}
	return ""
}

func P_no_initResults(fr, tx, r, m string) { /// prediction info, not initialized
	P.TopLabel.SetText("SCID: \n" + prediction.PredictControl.Contract + "\n" + "\nNot Accepting Predictions\n\nLast Round Mark: " + m +
		"\nLast Round Results: " + roundResults(fr, m) + "\nLast Round TXID: " + tx + "\n\nRounds Completed: " + r)

	P.BottomLabel.SetText("")
}

func PopulatePredictions(dc, gs bool) {
	if dc && gs {
		list := []string{}
		contracts := menu.Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
		keys := make([]string, len(contracts))

		i := 0
		for k := range contracts {
			keys[i] = k
			list = checkBetContractOwner(keys[i], "p", list)
			i++
		}
		t := len(list)
		list = append(list, " Contracts: "+strconv.Itoa(t))
		sort.Strings(list)
		prediction.PredictControl.Contract_list = list
	}
	prediction.PredictControl.Predict_list.Refresh()
}

func checkBetContractOwner(scid, t string, list []string) []string {
	owner, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "owner", menu.Gnomes.Indexer.ChainHeight, true)
	dev, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "dev", menu.Gnomes.Indexer.ChainHeight, true)
	_, init := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, t+"_init", menu.Gnomes.Indexer.ChainHeight, true)

	if owner != nil && dev != nil && init != nil {
		if dev[0] == rpc.DevAddress {
			headers, _ := rpc.GetSCHeaders(scid)
			name := "?"
			desc := "?"
			if headers != nil {
				if headers[1] != "" {
					desc = headers[1]
				}

				if headers[0] != "" {
					name = " " + headers[0]
				}
			}
			list = append(list, name+"   "+desc+"   "+scid)
			menu.DisableBetOwner(owner[0])
		}
	}

	return list
}

func MakeLeaderBoard(dc, gs bool, scid string) {
	if dc && gs && len(scid) == 64 {
		leaders := make(map[string]uint64)
		findLeaders := menu.Gnomes.Indexer.Backend.GetAllSCIDVariableDetails(scid)

		keys := make([]int64, 0, len(findLeaders))
		for k := range findLeaders {
			keys = append(keys, k)

		}

		sort.Slice(keys, func(i, j int) bool { return keys[i] > keys[j] })
		for val := range findLeaders[keys[0]] {
			a := findLeaders[keys[0]][val].Key
			split := strings.Split(a.(string), "_")
			if split[0] == "u" {
				leaders[split[1]] = uint64(findLeaders[keys[0]][val].Value.(float64))
			}
		}

		prediction.PredictControl.Leaders_map = leaders

		printLeaders()
	}
}

func printLeaders() {
	prediction.PredictControl.Leaders_display = []string{" Leaders: " + strconv.Itoa(len(prediction.PredictControl.Leaders_map))}
	keys := make([]string, 0, len(prediction.PredictControl.Leaders_map))

	for key := range prediction.PredictControl.Leaders_map {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return prediction.PredictControl.Leaders_map[keys[i]] > prediction.PredictControl.Leaders_map[keys[j]]
	})

	for _, k := range keys {
		prediction.PredictControl.Leaders_list.Refresh()
		prediction.PredictControl.Leaders_display = append(prediction.PredictControl.Leaders_display, k+": "+strconv.FormatUint(prediction.PredictControl.Leaders_map[k], 10))
	}

	prediction.PredictControl.Leaders_list.Refresh()
}

func PopulateSports(dc, gs bool) {
	if dc && gs {
		list := []string{}
		//prediction.SportsControl.Contract_list = []string{}
		contracts := menu.Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
		keys := make([]string, len(contracts))

		i := 0
		for k := range contracts {
			keys[i] = k
			list = checkBetContractOwner(keys[i], "s", list)
			i++
		}

		t := len(list)
		list = append(list, " Contracts: "+strconv.Itoa(t))
		sort.Strings(list)
		prediction.SportsControl.Contract_list = list
	}
	prediction.SportsControl.Sports_list.Refresh()
}

func GetBook(dc bool, scid string) {
	if dc && !menu.GnomonClosing() && !menu.GnomonWriting() {
		_, initValue := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_init", menu.Gnomes.Indexer.ChainHeight, true)
		if initValue != nil {
			_, playedValue := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_played", menu.Gnomes.Indexer.ChainHeight, true)
			//_, hl := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "hl", menu.Gnomes.Indexer.ChainHeight, true)
			init := initValue[0]
			played := playedValue[0]

			table.Actions.Game_options = []string{}
			table.Actions.Game_select.Options = table.Actions.Game_options
			played_str := strconv.Itoa(int(played))
			if init == played {
				S.TopLabel.SetText("SCID: \n" + scid + "\n\nGames Completed: " + played_str + "\n\nNo current Games\n")
				return
			}

			var single bool
			iv := 1
			for {
				_, s_init := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_init_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
				if s_init != nil {
					game, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "game_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					league, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "league_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, s_n := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_#_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, s_amt := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_amount_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, s_end := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_end_at_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, s_total := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_total_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					//s_urlValue, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_url_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, s_ta := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "team_a_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, s_tb := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "team_b_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, time_a := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "time_a", menu.Gnomes.Indexer.ChainHeight, true)
					_, time_b := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "time_b", menu.Gnomes.Indexer.ChainHeight, true)

					team_a := menu.TrimTeamA(game[0])
					team_b := menu.TrimTeamB(game[0])

					if s_end[0] > uint64(time.Now().Unix()) {
						current := table.Actions.Game_select.Options
						new := append(current, strconv.Itoa(iv)+"   "+game[0])
						table.Actions.Game_select.Options = new
					}

					eA := fmt.Sprint(s_end[0] * 1000)
					min := fmt.Sprint(float64(s_amt[0]) / 100000)
					n := strconv.Itoa(int(s_n[0]))
					aV := strconv.Itoa(int(s_ta[0]))
					bV := strconv.Itoa(int(s_tb[0]))
					t := strconv.Itoa(int(s_total[0]))
					if !single {
						single = true
						S.TopLabel.SetText("SCID: \n" + scid + "\n\nGames Completed: " + played_str + "\nCurrent Games:\n")
					}
					S_Results(game[0], strconv.Itoa(iv), league[0], min, eA, n, team_a, team_b, aV, bV, t, time_a[0], time_b[0])

				}

				if iv >= int(init) {
					break
				}

				iv++
			}
			table.Actions.Game_select.Refresh()
		}
	}
}

func S_Results(g, gN, l, min, eA, c, tA, tB, tAV, tBV, total string, a, b uint64) { /// sports info label
	result, err := strconv.ParseFloat(total, 32)

	if err != nil {
		log.Println("Float Conversion Error", err)
	}

	s := fmt.Sprintf("%.5f", result/100000)
	end_time, _ := rpc.MsToTime(eA)
	utc_end := end_time.String()

	pa := strconv.Itoa(int(a/60) / 60)
	rf := strconv.Itoa(int(b/60) / 60)

	S.TopLabel.SetText(S.TopLabel.Text + "\nGame " + gN + " - " + g + "\nLeague: " + l + "\nMinimum: " + min +
		" Dero\nCloses at: " + utc_end + "\nPayout " + pa + " hours after close\nRefund if not paid " + rf + " within hours\nPot Total: " + s + "\nPicks: " + c + "\n" + tA + " Picks: " + tAV + "\n" + tB + " Picks: " + tBV + "\n")

	S.TopLabel.Refresh()
}

func MenuRefresh(tab, gi bool) {
	if tab && gi {
		dHeight, _ := rpc.DaemonHeight()
		var index int
		if !menu.GnomonClosing() && menu.FastSynced() {
			index = int(menu.Gnomes.Indexer.ChainHeight) //int(menu.Gnomes.Indexer.Backend.GetLastIndexHeight())
		}
		table.Assets.Gnomes_height.Text = (" Gnomon Height: " + strconv.Itoa(index))
		table.Assets.Gnomes_height.Refresh()
		table.Assets.Gnomes_index.Text = (" Indexed SCIDs: " + strconv.Itoa(int(menu.Gnomes.SCIDS)))
		table.Assets.Gnomes_index.Refresh()
		table.Assets.Daem_height.Text = (" Daemon Height: " + strconv.Itoa(int(dHeight)))
		table.Assets.Daem_height.Refresh()
		table.Assets.Wall_height.Text = (" Wallet Height: " + rpc.Wallet.Height)
		table.Assets.Wall_height.Refresh()
		table.Assets.Dreams_bal.Text = (" dReams Balance: " + rpc.Wallet.TokenBal)
		table.Assets.Dreams_bal.Refresh()
		table.Assets.Dero_bal.Text = (" Dero Balance: " + rpc.Wallet.Balance)
		table.Assets.Dero_bal.Refresh()
		if offset == 20 {
			_, price := table.GetPrice("DERO-USDT")
			table.Assets.Dero_price.Text = (" Dero Price: $" + price)
			table.Assets.Dero_price.Refresh()
		}

		if dReams.menu_tabs.contracts {
			if offset%3 == 0 {
				go menu.GetTableStats(rpc.Round.Contract, false)
			}
		}

		if dReams.menu_tabs.market && !isWindows() {
			menu.FindNfaListings(menu.Gnomes.Sync)
			menu.Market.Auction_list.Refresh()
			menu.Market.Buy_list.Refresh()
		}
	}

	if !dReams.menu {
		menu.Market.Viewing = ""
	}

}

func RecheckButton() fyne.CanvasObject {
	button := widget.NewButton("Check Assets", func() {
		log.Println("Rechecking Assets")
		go RecheckAssets()
	})

	return button
}

func RecheckAssets() {
	table.Settings.FaceSelect.Options = []string{"Light", "Dark"}
	table.Settings.BackSelect.Options = []string{"Light", "Dark"}
	table.Settings.ThemeSelect.Options = []string{"Main"}
	table.Settings.AvatarSelect.Options = []string{"None"}
	table.Assets.Assets = []string{}
	menu.CheckAssets(menu.Gnomes.Sync, false)
	menu.CheckG45owner(menu.Gnomes.Sync, false)

}

func MainTab(ti *container.TabItem) {
	switch ti.Text {
	case "Menu":
		dReams.menu = true
		dReams.holdero = false
		dReams.bacc = false
		dReams.predict = false
		dReams.sports = false
		if rpc.Round.ID == 1 {
			table.Settings.FaceSelect.Enable()
			table.Settings.BackSelect.Enable()
		}
		go MenuRefresh(dReams.menu, menu.Gnomes.Init)
	case "Holdero":
		dReams.menu = false
		dReams.holdero = true
		dReams.bacc = false
		dReams.predict = false
		dReams.sports = false
		HolderoRefresh()
	case "Baccarat":
		dReams.menu = false
		dReams.holdero = false
		dReams.bacc = true
		dReams.predict = false
		dReams.sports = false
		BaccRefresh()
		if rpc.Wallet.Connect && rpc.Bacc.Display {
			table.BaccBuffer(false)
		}
	case "Predict":
		dReams.menu = false
		dReams.holdero = false
		dReams.bacc = false
		dReams.predict = true
		dReams.sports = false
		table.Actions.NameEntry.Text = menu.CheckPredictionName(prediction.PredictControl.Contract)
		table.Actions.NameEntry.Refresh()
		go PopulatePredictions(rpc.Signal.Daemon, menu.Gnomes.Sync)
		PredictionRefresh(dReams.predict)
	case "Sports":
		dReams.menu = false
		dReams.holdero = false
		dReams.bacc = false
		dReams.predict = false
		dReams.sports = true
		go PopulateSports(rpc.Signal.Daemon, menu.Gnomes.Sync)
		go GetBook(rpc.Signal.Daemon, prediction.SportsControl.Contract)
	}
}

func MenuTab(ti *container.TabItem) {
	switch ti.Text {
	case "Wallet":
		dReams.menu_tabs.wallet = true
		dReams.menu_tabs.contracts = false
		dReams.menu_tabs.assets = false
		dReams.menu_tabs.market = false
	case "Contracts":
		dReams.menu_tabs.wallet = false
		dReams.menu_tabs.contracts = true
		dReams.menu_tabs.assets = false
		dReams.menu_tabs.market = false
		go PopulatePredictions(rpc.Signal.Daemon, menu.Gnomes.Sync)
		if rpc.Wallet.Connect && menu.Gnomes.Checked {
			go menu.CreateTableList(false)
		}
	case "Assets":
		dReams.menu_tabs.wallet = false
		dReams.menu_tabs.contracts = false
		dReams.menu_tabs.assets = true
		dReams.menu_tabs.market = false
		menu.PlayerControl.Viewing_asset = ""
		table.Assets.Asset_list.UnselectAll()
	case "Market":
		dReams.menu_tabs.wallet = false
		dReams.menu_tabs.contracts = false
		dReams.menu_tabs.assets = false
		dReams.menu_tabs.market = true
		go menu.FindNfaListings(menu.Gnomes.Sync)
		menu.Market.Cancel_button.Hide()
		menu.Market.Close_button.Hide()
		menu.Market.Auction_list.Refresh()
		menu.Market.Buy_list.Refresh()
	}
}

func MenuContractTab(ti *container.TabItem) {
	switch ti.Text {
	case "Tables":
		if rpc.Signal.Daemon {
			go menu.CreateTableList(false)
		}
	}
}

func MarketTab(ti *container.TabItem) {
	switch ti.Text {
	case "Auctions":
		go menu.FindNfaListings(menu.Gnomes.Sync)
		menu.Market.Tab = "Auction"
		menu.Market.Auction_list.UnselectAll()
		menu.Market.Viewing = ""
		menu.Market.Market_button.Text = "Bid"
		menu.Market.Market_button.Refresh()
		menu.Market.Entry.SetText("0.0")
		menu.Market.Entry.Enable()
		menu.ResetAuctionInfo()
		menu.AuctionInfo()
	case "Buy Now":
		go menu.FindNfaListings(menu.Gnomes.Sync)
		menu.Market.Tab = "Buy"
		menu.Market.Buy_list.UnselectAll()
		menu.Market.Viewing = ""
		menu.Market.Market_button.Text = "Buy"
		menu.Market.Entry.Disable()
		menu.Market.Market_button.Refresh()
		menu.Market.Details_box.Refresh()
		menu.ResetBuyInfo()
		menu.BuyNowInfo()
	}
}

func PredictTab(ti *container.TabItem) {
	switch ti.Text {
	case "Contracts":
		go PopulatePredictions(rpc.Signal.Daemon, menu.Gnomes.Sync)
	case "Leaderboard":
		go MakeLeaderBoard(rpc.Signal.Daemon, menu.Gnomes.Sync, prediction.PredictControl.Contract)
	}
}

func DisplayCard(card int) *canvas.Image {
	if !table.Settings.Shared || rpc.Round.ID == 1 {
		if card == 99 {
			return canvas.NewImageFromImage(nil)
		}

		if card > 0 {
			i := table.Settings.FaceSelect.SelectedIndex()
			switch i {
			case 0:
				return canvas.NewImageFromResource(DisplayLightCard(card))
			case 1:
				return canvas.NewImageFromResource(DisplayDarkCard(card))
			default:
				return CustomCard(card, table.Settings.Faces)
			}
		}

		i := table.Settings.BackSelect.SelectedIndex()
		switch i {
		case 0:
			return canvas.NewImageFromResource(resourceBack1Png)
		case 1:
			return canvas.NewImageFromResource(resourceBack2Png)
		default:
			return CustomBack(table.Settings.Backs)
		}

	} else {
		if card == 99 {
			return canvas.NewImageFromImage(nil)
		} else if card > 0 {
			return CustomCard(card, rpc.Round.Face)
		} else {
			return CustomBack(rpc.Round.Back)
		}
	}
}

func DisplayLightCard(card int) fyne.Resource {
	if card > 0 && card < 53 {
		switch card {
		case 1:
			return resourceLightcard1Png
		case 2:
			return resourceLightcard2Png
		case 3:
			return resourceLightcard3Png
		case 4:
			return resourceLightcard4Png
		case 5:
			return resourceLightcard5Png
		case 6:
			return resourceLightcard6Png
		case 7:
			return resourceLightcard7Png
		case 8:
			return resourceLightcard8Png
		case 9:
			return resourceLightcard9Png
		case 10:
			return resourceLightcard10Png
		case 11:
			return resourceLightcard11Png
		case 12:
			return resourceLightcard12Png
		case 13:
			return resourceLightcard13Png
		case 14:
			return resourceLightcard14Png
		case 15:
			return resourceLightcard15Png
		case 16:
			return resourceLightcard16Png
		case 17:
			return resourceLightcard17Png
		case 18:
			return resourceLightcard18Png
		case 19:
			return resourceLightcard19Png
		case 20:
			return resourceLightcard20Png
		case 21:
			return resourceLightcard21Png
		case 22:
			return resourceLightcard22Png
		case 23:
			return resourceLightcard23Png
		case 24:
			return resourceLightcard24Png
		case 25:
			return resourceLightcard25Png
		case 26:
			return resourceLightcard26Png
		case 27:
			return resourceLightcard27Png
		case 28:
			return resourceLightcard28Png
		case 29:
			return resourceLightcard29Png
		case 30:
			return resourceLightcard30Png
		case 31:
			return resourceLightcard31Png
		case 32:
			return resourceLightcard32Png
		case 33:
			return resourceLightcard33Png
		case 34:
			return resourceLightcard34Png
		case 35:
			return resourceLightcard35Png
		case 36:
			return resourceLightcard36Png
		case 37:
			return resourceLightcard37Png
		case 38:
			return resourceLightcard38Png
		case 39:
			return resourceLightcard39Png
		case 40:
			return resourceLightcard40Png
		case 41:
			return resourceLightcard41Png
		case 42:
			return resourceLightcard42Png
		case 43:
			return resourceLightcard43Png
		case 44:
			return resourceLightcard44Png
		case 45:
			return resourceLightcard45Png
		case 46:
			return resourceLightcard46Png
		case 47:
			return resourceLightcard47Png
		case 48:
			return resourceLightcard48Png
		case 49:
			return resourceLightcard49Png
		case 50:
			return resourceLightcard50Png
		case 51:
			return resourceLightcard51Png
		case 52:
			return resourceLightcard52Png
		}
	}
	return nil
}

func DisplayDarkCard(card int) fyne.Resource {
	if card > 0 && card < 53 {
		switch card {
		case 1:
			return resourceDarkcard1Png
		case 2:
			return resourceDarkcard2Png
		case 3:
			return resourceDarkcard3Png
		case 4:
			return resourceDarkcard4Png
		case 5:
			return resourceDarkcard5Png
		case 6:
			return resourceDarkcard6Png
		case 7:
			return resourceDarkcard7Png
		case 8:
			return resourceDarkcard8Png
		case 9:
			return resourceDarkcard9Png
		case 10:
			return resourceDarkcard10Png
		case 11:
			return resourceDarkcard11Png
		case 12:
			return resourceDarkcard12Png
		case 13:
			return resourceDarkcard13Png
		case 14:
			return resourceDarkcard14Png
		case 15:
			return resourceDarkcard15Png
		case 16:
			return resourceDarkcard16Png
		case 17:
			return resourceDarkcard17Png
		case 18:
			return resourceDarkcard18Png
		case 19:
			return resourceDarkcard19Png
		case 20:
			return resourceDarkcard20Png
		case 21:
			return resourceDarkcard21Png
		case 22:
			return resourceDarkcard22Png
		case 23:
			return resourceDarkcard23Png
		case 24:
			return resourceDarkcard24Png
		case 25:
			return resourceDarkcard25Png
		case 26:
			return resourceDarkcard26Png
		case 27:
			return resourceDarkcard27Png
		case 28:
			return resourceDarkcard28Png
		case 29:
			return resourceDarkcard29Png
		case 30:
			return resourceDarkcard30Png
		case 31:
			return resourceDarkcard31Png
		case 32:
			return resourceDarkcard32Png
		case 33:
			return resourceDarkcard33Png
		case 34:
			return resourceDarkcard34Png
		case 35:
			return resourceDarkcard35Png
		case 36:
			return resourceDarkcard36Png
		case 37:
			return resourceDarkcard37Png
		case 38:
			return resourceDarkcard38Png
		case 39:
			return resourceDarkcard39Png
		case 40:
			return resourceDarkcard40Png
		case 41:
			return resourceDarkcard41Png
		case 42:
			return resourceDarkcard42Png
		case 43:
			return resourceDarkcard43Png
		case 44:
			return resourceDarkcard44Png
		case 45:
			return resourceDarkcard45Png
		case 46:
			return resourceDarkcard46Png
		case 47:
			return resourceDarkcard47Png
		case 48:
			return resourceDarkcard48Png
		case 49:
			return resourceDarkcard49Png
		case 50:
			return resourceDarkcard50Png
		case 51:
			return resourceDarkcard51Png
		case 52:
			return resourceDarkcard52Png
		}
	}
	return nil
}
