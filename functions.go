package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/prediction"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/table"
	"github.com/docopt/docopt-go"

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
	Name    string   `json:"name"`
	Daemon  []string `json:"daemon"`
	Tables  []string `json:"tables"`
	Predict []string `json:"predict"`
	Sports  []string `json:"sports"`
}

var command_line string = `dReams
dReam Tables all in one dApp, powered by Gnomon.

Usage:
  dReams [options]
  dReams -h | --help

Options:
  -h --help     Show this screen.
  --trim=<false>	dReams option, set true to trim index search filters.
  --fastsync=<false>	Gnomon option,  true/false value to define loading at chain height on start up.
  --num-parallel-blocks=<5>   Gnomon option,  defines the number of parallel blocks to index.`

var offset int

func flags() {
	arguments, err := docopt.ParseArgs(command_line, nil, "v0.9.2")

	if err != nil {
		log.Fatalf("Error while parsing arguments: %s\n", err)
	}

	trim := false
	if arguments["--trim"] != nil {
		if arguments["--trim"].(string) == "true" {
			trim = true
		}
	}

	fastsync := true
	if arguments["--fastsync"] != nil {
		if arguments["--fastsync"].(string) == "false" {
			fastsync = false
		}
	}

	parallel := 1
	if arguments["--num-parallel-blocks"] != nil {
		s := arguments["--num-parallel-blocks"].(string)
		switch s {
		case "2":
			parallel = 2
		case "3":
			parallel = 3
		case "4":
			parallel = 4
		case "5":
			parallel = 5
		default:
			parallel = 1
		}
	}

	menu.Gnomes.Trim = trim
	menu.Gnomes.Fast = fastsync
	menu.Gnomes.Para = parallel
}

func init() {
	saved := readConfig()

	table.Poker_name = saved.Name

	if saved.Daemon != nil {
		menu.MenuControl.Daemon_config = saved.Daemon[0]
	}

	menu.MenuControl.Holdero_favorites = saved.Tables
	menu.MenuControl.Predict_favorites = saved.Predict
	menu.MenuControl.Sports_favorites = saved.Sports

	rpc.Signal.Sit = true

	table.InitTableSettings()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		writeConfig(makeConfig(table.Poker_name, rpc.Round.Daemon))
		fmt.Println()
		menu.StopGnomon(menu.Gnomes.Init)
		menu.StopIndicators()
		log.Println("[dReams] Closing")
		os.Exit(0)
	}()
}

func stamp() {
	fmt.Println(`♤♡♧♢♧♡♤♡♧♢♧♡♤♡♧♢♧♡♤♡♧♢♧♡♤♡♧♢♧♡♤`)
	fmt.Println(`        dReams v0.9.2`)
	fmt.Println(`   https://dreamtables.net`)
	fmt.Println(`   ©2022-2023 dReam Tables`)
	fmt.Println(`♤♡♧♢♧♡♤♡♧♢♧♡♤♡♧♢♧♡♤♡♧♢♧♡♤♡♧♢♧♡♤`)
}

func notification(title, content string, g int) *fyne.Notification {
	switch g {
	case 0:
		rpc.Round.Notified = true
	case 1:
		rpc.Bacc.Notified = true
	case 2:
		rpc.Tarot.Notified = true
	default:
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
				log.Println("[dReams] Tapped show")
			}),
			fyne.NewMenuItem("Reveal Key", func() {
				go rpc.RevealKey(rpc.Wallet.ClientKey)
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
	// case "dero-node.mysrv.cloud:10102":
	// case "derostats.io:10102":
	case "publicrpc1.dero.io:10102":
	default:
		data.Daemon = []string{daemon}
	}

	data.Tables = menu.MenuControl.Holdero_favorites
	data.Predict = menu.MenuControl.Predict_favorites
	data.Sports = menu.MenuControl.Sports_favorites

	return
}

func writeConfig(u save) {
	if u.Daemon != nil && u.Name != "" {
		if u.Daemon[0] != "" {
			file, err := os.Create("config.json")

			if err != nil {
				log.Println("[writeConfig]", err)
			}

			u.Tables = menu.MenuControl.Holdero_favorites
			u.Predict = menu.MenuControl.Predict_favorites
			u.Sports = menu.MenuControl.Sports_favorites

			defer file.Close()
			json, _ := json.MarshalIndent(u, "", " ")

			_, err = file.Write(json)

			if err != nil {
				log.Println("[writeConfig]", err)
			}
		}
	}
}

func readConfig() (saved save) {
	if !table.FileExists("config.json") {
		return
	}

	file, err := os.ReadFile("config.json")

	if err != nil {
		log.Println("[readConfig]", err)
		return
	}

	var config save
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Println("[readConfig]", err)
		return
	}

	return config
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

func placeHolderoCards() *fyne.Container {
	size := dReams.Window.Content().Size()
	Cards.Layout = container.NewWithoutLayout(
		Hole_1(0, size.Width, size.Height),
		Hole_2(0, size.Width, size.Height),
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
		Turn(rpc.Round.TurnCard),
		River(rpc.Round.RiverCard))

	return Cards.Layout
}

func refreshHolderoCards(l1, l2 string) {
	size := dReams.Window.Content().Size()
	Cards.Layout.Objects[0] = Hole_1(Card(l1), size.Width, size.Height)
	Cards.Layout.Objects[0].Refresh()

	Cards.Layout.Objects[1] = Hole_2(Card(l2), size.Width, size.Height)
	Cards.Layout.Objects[1].Refresh()

	Cards.Layout.Objects[2] = P1_a(Is_In(rpc.CardHash.P1C1, 1, rpc.Signal.End))
	Cards.Layout.Objects[2].Refresh()

	Cards.Layout.Objects[3] = P1_b(Is_In(rpc.CardHash.P1C2, 1, rpc.Signal.End))
	Cards.Layout.Objects[3].Refresh()

	Cards.Layout.Objects[4] = P2_a(Is_In(rpc.CardHash.P2C1, 2, rpc.Signal.End))
	Cards.Layout.Objects[4].Refresh()

	Cards.Layout.Objects[5] = P2_b(Is_In(rpc.CardHash.P2C2, 2, rpc.Signal.End))
	Cards.Layout.Objects[5].Refresh()

	Cards.Layout.Objects[6] = P3_a(Is_In(rpc.CardHash.P3C1, 3, rpc.Signal.End))
	Cards.Layout.Objects[6].Refresh()

	Cards.Layout.Objects[7] = P3_b(Is_In(rpc.CardHash.P3C2, 3, rpc.Signal.End))
	Cards.Layout.Objects[7].Refresh()

	Cards.Layout.Objects[8] = P4_a(Is_In(rpc.CardHash.P4C1, 4, rpc.Signal.End))
	Cards.Layout.Objects[8].Refresh()

	Cards.Layout.Objects[9] = P4_b(Is_In(rpc.CardHash.P4C2, 4, rpc.Signal.End))
	Cards.Layout.Objects[9].Refresh()

	Cards.Layout.Objects[10] = P5_a(Is_In(rpc.CardHash.P5C1, 5, rpc.Signal.End))
	Cards.Layout.Objects[10].Refresh()

	Cards.Layout.Objects[11] = P5_b(Is_In(rpc.CardHash.P5C2, 5, rpc.Signal.End))
	Cards.Layout.Objects[11].Refresh()

	Cards.Layout.Objects[12] = P6_a(Is_In(rpc.CardHash.P6C1, 6, rpc.Signal.End))
	Cards.Layout.Objects[12].Refresh()

	Cards.Layout.Objects[13] = P6_b(Is_In(rpc.CardHash.P6C2, 6, rpc.Signal.End))
	Cards.Layout.Objects[13].Refresh()

	Cards.Layout.Objects[14] = Flop_1(rpc.Round.Flop1)
	Cards.Layout.Objects[14].Refresh()

	Cards.Layout.Objects[15] = Flop_2(rpc.Round.Flop2)
	Cards.Layout.Objects[15].Refresh()

	Cards.Layout.Objects[16] = Flop_3(rpc.Round.Flop3)
	Cards.Layout.Objects[16].Refresh()

	Cards.Layout.Objects[17] = Turn(rpc.Round.TurnCard)
	Cards.Layout.Objects[17].Refresh()

	Cards.Layout.Objects[18] = River(rpc.Round.RiverCard)
	Cards.Layout.Objects[18].Refresh()

	Cards.Layout.Refresh()
}

func ifBet(w, r uint64) { /// sets bet amount on turn
	if w > 0 && r > 0 && !rpc.Signal.PlacedBet {
		float := float64(w) / 100000
		wager := strconv.FormatFloat(float, 'f', 1, 64)
		table.Actions.BetEntry.SetText(wager)
		rpc.Display.Res = rpc.Round.Raisor + " Raised, " + wager + " to Call "
	} else if w > 0 && !rpc.Signal.PlacedBet {
		float := float64(w) / 100000
		wager := strconv.FormatFloat(float, 'f', 1, 64)
		table.Actions.BetEntry.SetText(wager)
		rpc.Display.Res = rpc.Round.Bettor + " Bet " + wager
	} else if r > 0 && rpc.Signal.PlacedBet {
		float := float64(r) / 100000
		rasied := strconv.FormatFloat(float, 'f', 1, 64)
		table.Actions.BetEntry.SetText(rasied)
		rpc.Display.Res = rpc.Round.Raisor + " Raised, " + rasied + " to Call"
	} else if w == 0 && !rpc.Signal.Bet {
		var float float64
		if rpc.Round.Ante == 0 {
			float = float64(rpc.Round.BB) / 100000
		} else {
			float = float64(rpc.Round.Ante) / 100000
		}
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
	var autoCF bool
	var autoD bool
	var trigger bool
	var skip int
	var delay int
	setLabels()
	for {
		select {
		case <-ticker.C: /// do on interval
			rpc.Ping()
			rpc.EchoWallet(rpc.Wallet.Connect)
			rpc.GetBalance(rpc.Wallet.Connect)
			go rpc.DreamsBalance(rpc.Wallet.Connect)
			rpc.TourneyBalance(rpc.Wallet.Connect, rpc.Round.Tourney, rpc.TourneySCID)
			rpc.GetHeight(rpc.Wallet.Connect)
			if !rpc.Signal.Startup {
				menu.CheckConnection()
				menu.GnomonEndPoint(rpc.Signal.Daemon, menu.Gnomes.Init, menu.Gnomes.Sync)
				rpc.FetchHolderoSC(rpc.Signal.Daemon, rpc.Signal.Contract)
				rpc.FetchBaccSC(rpc.Signal.Daemon)
				rpc.FetchTarotSC(rpc.Signal.Daemon)
				menu.GnomonState(rpc.Signal.Daemon, menu.Gnomes.Init, isWindows())
				background.Refresh()

				offset++
				if offset == 21 {
					offset = 0
				} else if offset%5 == 0 {
					SportsRefresh(dReams.sports)
				}

				if (rpc.StringToInt(rpc.Display.Turn) == rpc.Round.ID && rpc.StringToInt(rpc.Wallet.Height) > rpc.Signal.CHeight+4) ||
					(rpc.StringToInt(rpc.Display.Turn) != rpc.Round.ID && rpc.Round.ID >= 1) || (!rpc.Signal.My_turn && rpc.Round.ID >= 1) {
					if rpc.Signal.Clicked {
						trigger = false
						autoCF = false
						autoD = false
						rpc.Signal.Reveal = false
					}
					rpc.Signal.Clicked = false
				}

				BaccRefresh()
				PredictionRefresh(dReams.predict)
				S.RightLabel.SetText("dReams Balance: " + rpc.Wallet.TokenBal + "      Dero Balance: " + rpc.Wallet.Balance + "      Height: " + rpc.Wallet.Height)
				TarotRefresh()

				go MenuRefresh(dReams.menu, menu.Gnomes.Init)
				if !rpc.Signal.Clicked {
					if rpc.Round.First_try {
						rpc.Round.First_try = false
						delay = 0
						rpc.Round.Card_delay = false
						go refreshHolderoPlayers()
					}

					if rpc.Round.Card_delay {
						now := time.Now().Unix()
						delay++
						if delay >= 10 || now > rpc.Round.Last+39 {
							delay = 0
							rpc.Round.Card_delay = false
						}
					} else {
						setHolderoLabel()
						table.GetUrls(rpc.Round.F_url, rpc.Round.B_url)
						rpc.Called(rpc.Round.Flop, rpc.Round.Wager)
						trigger = singleShot(rpc.Signal.My_turn, trigger)
						HolderoRefresh()
						if table.Settings.Auto_check && rpc.Signal.My_turn && !autoCF {
							if !rpc.Signal.Reveal && !rpc.Signal.End && !rpc.Round.LocalEnd {
								if rpc.CardHash.Local1 != "" {
									table.HolderoButtonBuffer()
									rpc.Check()
									H.TopLabel.SetText("Auto Check/Fold Tx Sent")
									H.TopLabel.Refresh()
									autoCF = true

									go func() {
										if !isWindows() {
											time.Sleep(500 * time.Millisecond)
											dReams.App.SendNotification(notification("dReams - Holdero", "Auto Check/Fold Tx Sent", 9))
										}
									}()
								}
							}
						}

						if table.Settings.Auto_deal && rpc.Signal.My_turn && !autoD {
							if !rpc.Signal.Reveal && !rpc.Signal.End && !rpc.Round.LocalEnd {
								if rpc.CardHash.Local1 == "" {
									table.HolderoButtonBuffer()
									rpc.DealHand()
									H.TopLabel.SetText("Auto Deal Tx Sent")
									H.TopLabel.Refresh()
									autoD = true

									go func() {
										if !isWindows() {
											time.Sleep(500 * time.Millisecond)
											dReams.App.SendNotification(notification("dReams - Holdero", "Auto Deal Tx Sent", 9))
										}
									}()
								}
							}
						}

						skip = 0
					}
				} else {
					waitLabel()
					revealingKey()
					skip++
					if skip >= 18 {
						rpc.Signal.Clicked = false
						skip = 0
						trigger = false
						autoCF = false
						autoD = false
						rpc.Signal.Reveal = false
					}
				}
			}
			if rpc.Signal.Daemon {
				rpc.Signal.Startup = false
			}

		case <-quit: /// exit loop
			log.Println("[dReams] Closing")
			ticker.Stop()
			return
		}
	}
}

func setHolderoLabel() {
	H.TopLabel.SetText(rpc.Display.Res)
	H.LeftLabel.SetText("Seats: " + rpc.Display.Seats + "      Pot: " + rpc.Display.Pot + "      Blinds: " + rpc.Display.Blinds + "      Ante: " + rpc.Display.Ante + "      Dealer: " + rpc.Display.Dealer + "      Turn: " + rpc.Display.Turn)
	if rpc.Round.Asset {
		if rpc.Round.Tourney {
			H.RightLabel.SetText(rpc.Display.Readout + "      Player ID: " + rpc.Display.PlayerId + "      Chip Balance: " + rpc.Wallet.TourneyBal + "      Height: " + rpc.Wallet.Height)
		} else {
			H.RightLabel.SetText(rpc.Display.Readout + "      Player ID: " + rpc.Display.PlayerId + "      dReams Balance: " + rpc.Wallet.TokenBal + "      Height: " + rpc.Wallet.Height)
		}
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
		if rpc.Round.Tourney {
			H.RightLabel.SetText("Wait for Block" + "      Player ID: " + rpc.Display.PlayerId + "      Chip Balance: " + rpc.Wallet.TourneyBal + "      Height: " + rpc.Wallet.Height)
		} else {
			H.RightLabel.SetText("Wait for Block" + "      Player ID: " + rpc.Display.PlayerId + "      dReams Balance: " + rpc.Wallet.TokenBal + "      Height: " + rpc.Wallet.Height)
		}

	} else {
		H.RightLabel.SetText("Wait for Block" + "      Player ID: " + rpc.Display.PlayerId + "      Dero Balance: " + rpc.Wallet.Balance + "      Height: " + rpc.Wallet.Height)
	}
	H.TopLabel.Refresh()
	H.RightLabel.Refresh()
}

func HolderoRefresh() {
	go table.ShowAvatar(dReams.holdero)
	go refreshHolderoCards(rpc.CardHash.Local1, rpc.CardHash.Local2)
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
			} else if !rpc.Signal.Sit && rpc.Wallet.Connect {
				table.Actions.Sit.Show()
			}
			table.Actions.Leave.Hide()
			table.Actions.Deal.Hide()
			table.Actions.Check.Hide()
			table.Actions.Bet.Hide()
			table.Actions.BetEntry.Hide()

			if !rpc.Signal.My_turn && !rpc.Signal.End && !rpc.Round.LocalEnd {
				rpc.Display.Res = ""
				rpc.Round.Notified = false
			}
		}
	}

	go func() {
		refreshHolderoPlayers()
		H.TableItems.Refresh()
	}()
}

func refreshHolderoPlayers() {
	H.TableContent.Objects[0] = table.HolderoTable(resourceTablePng)
	H.TableContent.Objects[0].Refresh()

	H.TableContent.Objects[1] = table.Player1_label(resourceUnknownPng, resourceAvatarFramePng, resourceTurnFramePng)
	H.TableContent.Objects[1].Refresh()

	H.TableContent.Objects[2] = table.Player2_label(resourceUnknownPng, resourceAvatarFramePng, resourceTurnFramePng)
	H.TableContent.Objects[2].Refresh()

	H.TableContent.Objects[3] = table.Player3_label(resourceUnknownPng, resourceAvatarFramePng, resourceTurnFramePng)
	H.TableContent.Objects[3].Refresh()

	H.TableContent.Objects[4] = table.Player4_label(resourceUnknownPng, resourceAvatarFramePng, resourceTurnFramePng)
	H.TableContent.Objects[4].Refresh()

	H.TableContent.Objects[5] = table.Player5_label(resourceUnknownPng, resourceAvatarFramePng, resourceTurnFramePng)
	H.TableContent.Objects[5].Refresh()

	H.TableContent.Objects[6] = table.Player6_label(resourceUnknownPng, resourceAvatarFramePng, resourceTurnFramePng)
	H.TableContent.Objects[6].Refresh()

	H.TableContent.Refresh()
}

func revealingKey() {
	if rpc.Signal.Reveal && rpc.Signal.My_turn && !rpc.Signal.End {
		if !rpc.Round.Notified {
			rpc.Display.Res = "Revealing Key"
			H.TopLabel.SetText(rpc.Display.Res)
			H.TopLabel.Refresh()
			if !isWindows() {
				dReams.App.SendNotification(notification("dReams - Holdero", "Revealing Key", 0))
			}
		}
	}
}

func BaccRefresh() {
	B.LeftLabel.SetText("Total Hands Played: " + rpc.Display.Total_w + "      Player Wins: " + rpc.Display.Player_w + "      Ties: " + rpc.Display.Ties + "      Banker Wins: " + rpc.Display.Banker_w + "      Min Bet is " + rpc.Display.BaccMin + " dReams, Max Bet is " + rpc.Display.BaccMax)
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
		if offset%5 == 0 {
			go prediction.SetPredictionInfo(prediction.PredictControl.Contract)
		}

		if offset == 11 || prediction.PredictControl.Prices.Text == "" {
			go prediction.SetPredictionPrices(rpc.Signal.Daemon)
		}

		P.RightLabel.SetText("dReams Balance: " + rpc.Wallet.TokenBal + "      Dero Balance: " + rpc.Wallet.Balance + "      Height: " + rpc.Wallet.Height)

		if menu.CheckActivePrediction(prediction.PredictControl.Contract) {
			go prediction.ShowPredictionControls()
		} else {
			menu.DisablePreditions(true)
		}
	}
}

func SportsRefresh(tab bool) {
	if tab {
		go prediction.SetSportsInfo(prediction.SportsControl.Contract)
	}
}

func TarotRefresh() {
	T.LeftLabel.SetText("Total Readings: " + rpc.Display.Readings + "      Click your card for Iluma reading")
	T.RightLabel.SetText("dReams Balance: " + rpc.Wallet.TokenBal + "      Dero Balance: " + rpc.Wallet.Balance + "      Height: " + rpc.Wallet.Height)

	if !rpc.Tarot.Display {
		rpc.FetchTarotReading(rpc.Signal.Daemon, rpc.Tarot.Last)
		table.Iluma.Box.Refresh()
		if rpc.Tarot.Found {
			rpc.Tarot.Display = true
			table.Iluma.Label.SetText("")
			if rpc.Tarot.Num == 3 {
				table.Iluma.Card1.Objects[1] = TarotCard(rpc.Tarot.T_card1)
				table.Iluma.Card2.Objects[1] = TarotCard(rpc.Tarot.T_card2)
				table.Iluma.Card3.Objects[1] = TarotCard(rpc.Tarot.T_card3)
			} else {
				table.Iluma.Card1.Objects[1] = TarotCard(0)
				table.Iluma.Card2.Objects[1] = TarotCard(rpc.Tarot.T_card1)
				table.Iluma.Card3.Objects[1] = TarotCard(0)
			}
			table.TarotBuffer(false)
			table.Iluma.Box.Refresh()
		}
	}

	if rpc.StringToInt(rpc.Wallet.Height) > rpc.Tarot.CHeight+3 {
		table.TarotBuffer(false)
	}

	T.TableItems.Refresh()

	if rpc.Tarot.Found && !rpc.Tarot.Notified {
		if !isWindows() {
			dReams.App.SendNotification(notification("dReams - Tarot", "Your Reading has Arrvied", 2))
		}
	}
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

func refreshGnomonDisplay(index, c int) {
	if c == 1 {
		height := " Gnomon Height: " + strconv.Itoa(index)
		table.Assets.Gnomes_height.Text = (height)
		table.Assets.Gnomes_height.Refresh()
	} else {
		table.Assets.Gnomes_height.Text = (" Gnomon Height: 0")
		table.Assets.Gnomes_height.Refresh()
	}
}

func refreshIndexDisplay(c bool) {
	if c {
		scids := " Indexed SCIDs: " + strconv.Itoa(int(menu.Gnomes.SCIDS))
		table.Assets.Gnomes_index.Text = (scids)
		table.Assets.Gnomes_index.Refresh()
	} else {
		table.Assets.Gnomes_index.Text = (" Indexed SCIDs: 0")
		table.Assets.Gnomes_index.Refresh()
	}
}

func refreshDaemonDisplay(c bool) {
	if c && rpc.Signal.Daemon {
		dHeight, _ := rpc.DaemonHeight(rpc.Round.Daemon)
		d := strconv.Itoa(int(dHeight))
		table.Assets.Daem_height.Text = (" Daemon Height: " + d)
		table.Assets.Daem_height.Refresh()
	} else {
		table.Assets.Daem_height.Text = (" Daemon Height: 0")
		table.Assets.Daem_height.Refresh()
	}
}

func refreshWalletDisplay(c bool) {
	if c {
		table.Assets.Wall_height.Text = (" Wallet Height: " + rpc.Wallet.Height)
		table.Assets.Wall_height.Refresh()
		table.Assets.Dreams_bal.Text = (" dReams Balance: " + rpc.Wallet.TokenBal)
		table.Assets.Dreams_bal.Refresh()
		table.Assets.Dero_bal.Text = (" Dero Balance: " + rpc.Wallet.Balance)
		table.Assets.Dero_bal.Refresh()
	} else {
		table.Assets.Wall_height.Text = (" Wallet Height: 0")
		table.Assets.Wall_height.Refresh()
		table.Assets.Dreams_bal.Text = (" dReams Balance: 0")
		table.Assets.Dreams_bal.Refresh()
		table.Assets.Dero_bal.Text = (" Dero Balance: 0")
		table.Assets.Dero_bal.Refresh()
	}
}

func refreshPriceDisplay(c bool) {
	if c && rpc.Signal.Daemon {
		_, price := table.GetPrice("DERO-USDT")
		table.Assets.Dero_price.Text = (" Dero Price: $" + price)
		table.Assets.Dero_price.Refresh()
	} else {
		table.Assets.Dero_price.Text = (" Dero Price: $")
		table.Assets.Dero_price.Refresh()
	}
}

func MenuRefresh(tab, gi bool) {
	if tab && gi {
		var index int
		if !menu.GnomonClosing() && menu.FastSynced() {
			index = int(menu.Gnomes.Indexer.ChainHeight)
		}

		if !menu.FastSynced() {
			table.Assets.Gnomes_sync.Text = (" Gnomon Syncing... ")
			table.Assets.Gnomes_sync.Refresh()
		} else {
			if !menu.GnomonClosing() {
				table.Assets.Gnomes_sync.Text = ("")
				table.Assets.Gnomes_sync.Refresh()
			}
		}
		go refreshGnomonDisplay(index, 1)
		go refreshIndexDisplay(true)

		if rpc.Signal.Daemon {
			go refreshDaemonDisplay(true)
		}

		if offset == 20 {
			go refreshPriceDisplay(true)
		}

		if dReams.menu_tabs.contracts {
			if offset%3 == 0 {
				go menu.GetTableStats(rpc.Round.Contract, false)
			}
		}

		if dReams.menu_tabs.market && !isWindows() {
			menu.FindNfaListings(menu.Gnomes.Sync, nil)
		}
	}

	if rpc.Signal.Daemon {
		go refreshDaemonDisplay(true)
	} else {
		go refreshDaemonDisplay(false)
		go refreshGnomonDisplay(0, 0)
	}

	if rpc.Wallet.Connect {
		go refreshWalletDisplay(true)
	} else {
		go refreshWalletDisplay(false)
	}

	if !dReams.menu {
		menu.Market.Viewing = ""
	}
}

func RecheckButton() fyne.CanvasObject {
	button := widget.NewButton("Check Assets", func() {
		log.Println("[dReams] Rechecking Assets")
		go RecheckAssets()
	})

	return button
}

func RecheckAssets() {
	table.Assets.Assets = []string{}
	menu.CheckAssets(menu.Gnomes.Sync, false, nil)
	menu.CheckG45Assets(menu.Gnomes.Sync, false, nil)
	sort.Strings(table.Assets.Assets)
	table.Assets.Asset_list.UnselectAll()
	table.Assets.Asset_list.Refresh()

}

func MainTab(ti *container.TabItem) {
	switch ti.Text {
	case "Menu":
		dReams.menu = true
		dReams.holdero = false
		dReams.bacc = false
		dReams.predict = false
		dReams.sports = false
		dReams.tarot = false
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
		dReams.tarot = false
		go func() {
			now := time.Now().Unix()
			rpc.FetchHolderoSC(rpc.Signal.Daemon, rpc.Signal.Contract)
			if now > rpc.Round.Last+33 {
				HolderoRefresh()
			}
		}()
	case "Baccarat":
		dReams.menu = false
		dReams.holdero = false
		dReams.bacc = true
		dReams.predict = false
		dReams.sports = false
		dReams.tarot = false
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
		dReams.tarot = false
		go func() {
			table.Actions.NameEntry.Text = menu.CheckPredictionName(prediction.PredictControl.Contract)
			table.Actions.NameEntry.Refresh()
			menu.PopulatePredictions(rpc.Signal.Daemon, menu.Gnomes.Sync, nil)
		}()
		PredictionRefresh(dReams.predict)
	case "Sports":
		dReams.menu = false
		dReams.holdero = false
		dReams.bacc = false
		dReams.predict = false
		dReams.sports = true
		dReams.tarot = false
		go menu.PopulateSports(rpc.Signal.Daemon, menu.Gnomes.Sync, nil)
	case "Tarot":
		dReams.menu = false
		dReams.holdero = false
		dReams.bacc = false
		dReams.predict = false
		dReams.sports = false
		dReams.tarot = true
		if rpc.Wallet.Connect && rpc.Tarot.Display {
			table.TarotBuffer(false)
		}
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
		go menu.PopulatePredictions(rpc.Signal.Daemon, menu.Gnomes.Sync, nil)
		if rpc.Wallet.Connect && menu.Gnomes.Checked {
			go menu.CreateTableList(false, nil)
		}
	case "Assets":
		dReams.menu_tabs.wallet = false
		dReams.menu_tabs.contracts = false
		dReams.menu_tabs.assets = true
		dReams.menu_tabs.market = false
		menu.MenuControl.Viewing_asset = ""
		table.Assets.Asset_list.UnselectAll()
	case "Market":
		dReams.menu_tabs.wallet = false
		dReams.menu_tabs.contracts = false
		dReams.menu_tabs.assets = false
		dReams.menu_tabs.market = true
		go menu.FindNfaListings(menu.Gnomes.Sync, nil)
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
			go menu.CreateTableList(false, nil)
		}

	default:
	}
}

func MarketTab(ti *container.TabItem) {
	switch ti.Text {
	case "Auctions":
		go menu.FindNfaListings(menu.Gnomes.Sync, nil)
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
		go menu.FindNfaListings(menu.Gnomes.Sync, nil)
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
		go menu.PopulatePredictions(rpc.Signal.Daemon, menu.Gnomes.Sync, nil)
	case "Leaderboard":
		go MakeLeaderBoard(rpc.Signal.Daemon, menu.Gnomes.Sync, prediction.PredictControl.Contract)

	default:
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
			case -1:
				return canvas.NewImageFromResource(DisplayLightCard(card))
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
		case -1:
			return canvas.NewImageFromResource(resourceBack1Png)
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
