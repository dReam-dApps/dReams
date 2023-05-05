package main

import (
	"fmt"
	"image/color"
	"log"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/SixofClubsss/dReams/baccarat"
	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/holdero"
	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/prediction"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/tarot"
	"github.com/docopt/docopt-go"
	"github.com/fyne-io/terminal"

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

var cli *terminal.Terminal
var command_line string = `dReams
dReam Tables all in one dApp, powered by Gnomon.

Usage:
  dReams [options]
  dReams -h | --help

Options:
  -h --help     Show this screen.
  --cli=<false>	dReams option, enables cli app tab.
  --trim=<false>	dReams option, defaults true for minimum index search filters.
  --fastsync=<false>	Gnomon option,  true/false value to define loading at chain height on start up.
  --num-parallel-blocks=<5>   Gnomon option,  defines the number of parallel blocks to index.`

var offset int

// Set opts when starting dReams
func flags() (version string) {
	version = rpc.DREAMSv
	arguments, err := docopt.ParseArgs(command_line, nil, version)

	if err != nil {
		log.Fatalf("Error while parsing arguments: %s\n", err)
	}

	trim := true
	if arguments["--trim"] != nil {
		if arguments["--trim"].(string) == "false" {
			trim = false
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

	cli := false
	if arguments["--cli"] != nil {
		if arguments["--cli"].(string) == "true" {
			cli = true
		}
	}

	dReams.cli = cli
	menu.Gnomes.Trim = trim
	menu.Gnomes.Fast = fastsync
	menu.Gnomes.Para = parallel

	return
}

func init() {
	saved := menu.ReadDreamsConfig("dReams")
	if saved.Daemon != nil {
		menu.Control.Daemon_config = saved.Daemon[0]
	}

	menu.Control.Holdero_favorites = saved.Tables
	menu.Control.Predict_favorites = saved.Predict
	menu.Control.Sports_favorites = saved.Sports

	menu.Market.DreamsFilter = true

	rpc.Wallet.TokenBal = make(map[string]uint64)
	rpc.Display.Token_balance = make(map[string]string)

	rpc.Signal.Sit = true

	holdero.InitTableSettings()

	dReams.os = runtime.GOOS
	prediction.SetPrintColors(dReams.os)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		menu.Exit_signal = true
		menu.WriteDreamsConfig(rpc.Daemon.Rpc, bundle.AppColor)
		fmt.Println()
		serviceRunning()
		go menu.StopLabel()
		menu.StopGnomon("dReams")
		menu.StopIndicators()
		time.Sleep(time.Second)
		log.Println("[dReams] Closing")
		dReams.Window.Close()
	}()
}

// Starts a Fyne terminal in dReams
func startTerminal() *terminal.Terminal {
	cli = terminal.New()
	go func() {
		_ = cli.RunLocalShell()
	}()

	return cli
}

// Exit running dReams terminal
func exitTerminal() {
	if cli != nil {
		cli.Exit()
	}
}

// Ensure service is shutdown on app close
func serviceRunning() {
	rpc.Wallet.Service = false
	for prediction.Service.Processing {
		log.Println("[dReams] Waiting for service to close")
		time.Sleep(3 * time.Second)
	}
}

// Terminal start info, ascii art for linux
func stamp(v string) {
	if dReams.os == "linux" {
		fmt.Println(string(bundle.ResourceStampTxt.StaticContent))
	}
	log.Println("[dReams]", v, runtime.GOOS, runtime.GOARCH)
}

// Notification switch for dApps
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

// Check if runtime os is windows
func isWindows() bool {
	return dReams.os == "windows"
}

// Make system tray with opts
//   - Send Dero message menu
//   - Explorer link
//   - Manual reveal key for Holdero
func systemTray(w fyne.App) bool {
	if desk, ok := w.(desktop.App); ok {
		m := fyne.NewMenu("MyApp",
			fyne.NewMenuItem("Send Message", func() {
				if !dReams.configure && rpc.Wallet.Connect {
					menu.SendMessageMenu(bundle.ResourceDTGnomonIconPng, bundle.ResourceOwBackgroundPng)
				}
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Explorer", func() {
				link, _ := url.Parse("https://explorer.dero.io")
				fyne.CurrentApp().OpenURL(link)
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Reveal Key", func() {
				go rpc.RevealKey(rpc.Wallet.ClientKey)
			}))
		desk.SetSystemTrayMenu(m)

		return true
	}
	return false
}

// Top label background used on dApp tabs
func labelColorBlack(c *fyne.Container) *fyne.Container {
	var alpha *canvas.Rectangle
	if bundle.AppColor == color.White {
		alpha = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x33})
	} else {
		alpha = canvas.NewRectangle(color.RGBA{0, 0, 0, 150})
	}

	cont := container.New(layout.NewMaxLayout(), alpha, c)

	return cont
}

// Place and refresh Baccarat card images
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

	content := *container.NewWithoutLayout(
		PlayerCards(BaccSuit(rpc.Bacc.P_card1), BaccSuit(rpc.Bacc.P_card2), BaccSuit(drawP)),
		BankerCards(BaccSuit(rpc.Bacc.B_card1), BaccSuit(rpc.Bacc.B_card2), BaccSuit(drawB)))

	rpc.Bacc.Display = true
	baccarat.BaccBuffer(false)

	return &content
}

func clearBaccCards() *fyne.Container {
	content := *container.NewWithoutLayout(
		PlayerCards(99, 99, 99),
		BankerCards(99, 99, 99))

	return &content
}

// Place Holdero card images
func placeHolderoCards() *fyne.Container {
	size := dReams.Window.Content().Size()
	Cards.Layout = container.NewWithoutLayout(
		Hole_1(0, size.Width, size.Height),
		Hole_2(0, size.Width, size.Height),
		P1_a(Is_In(rpc.Round.Cards.P1C1, 1, rpc.Signal.End)),
		P1_b(Is_In(rpc.Round.Cards.P1C2, 1, rpc.Signal.End)),
		P2_a(Is_In(rpc.Round.Cards.P2C1, 2, rpc.Signal.End)),
		P2_b(Is_In(rpc.Round.Cards.P2C2, 2, rpc.Signal.End)),
		P3_a(Is_In(rpc.Round.Cards.P3C1, 3, rpc.Signal.End)),
		P3_b(Is_In(rpc.Round.Cards.P3C2, 3, rpc.Signal.End)),
		P4_a(Is_In(rpc.Round.Cards.P4C1, 4, rpc.Signal.End)),
		P4_b(Is_In(rpc.Round.Cards.P4C2, 4, rpc.Signal.End)),
		P5_a(Is_In(rpc.Round.Cards.P5C1, 5, rpc.Signal.End)),
		P5_b(Is_In(rpc.Round.Cards.P5C2, 5, rpc.Signal.End)),
		P6_a(Is_In(rpc.Round.Cards.P6C1, 6, rpc.Signal.End)),
		P6_b(Is_In(rpc.Round.Cards.P6C2, 6, rpc.Signal.End)),
		Flop_1(rpc.Round.Flop1),
		Flop_2(rpc.Round.Flop2),
		Flop_3(rpc.Round.Flop3),
		Turn(rpc.Round.TurnCard),
		River(rpc.Round.RiverCard))

	return Cards.Layout
}

// Refresh Holdero card images
func refreshHolderoCards(l1, l2 string) {
	size := dReams.Window.Content().Size()
	Cards.Layout.Objects[0] = Hole_1(rpc.Card(l1), size.Width, size.Height)
	Cards.Layout.Objects[0].Refresh()

	Cards.Layout.Objects[1] = Hole_2(rpc.Card(l2), size.Width, size.Height)
	Cards.Layout.Objects[1].Refresh()

	Cards.Layout.Objects[2] = P1_a(Is_In(rpc.Round.Cards.P1C1, 1, rpc.Signal.End))
	Cards.Layout.Objects[2].Refresh()

	Cards.Layout.Objects[3] = P1_b(Is_In(rpc.Round.Cards.P1C2, 1, rpc.Signal.End))
	Cards.Layout.Objects[3].Refresh()

	Cards.Layout.Objects[4] = P2_a(Is_In(rpc.Round.Cards.P2C1, 2, rpc.Signal.End))
	Cards.Layout.Objects[4].Refresh()

	Cards.Layout.Objects[5] = P2_b(Is_In(rpc.Round.Cards.P2C2, 2, rpc.Signal.End))
	Cards.Layout.Objects[5].Refresh()

	Cards.Layout.Objects[6] = P3_a(Is_In(rpc.Round.Cards.P3C1, 3, rpc.Signal.End))
	Cards.Layout.Objects[6].Refresh()

	Cards.Layout.Objects[7] = P3_b(Is_In(rpc.Round.Cards.P3C2, 3, rpc.Signal.End))
	Cards.Layout.Objects[7].Refresh()

	Cards.Layout.Objects[8] = P4_a(Is_In(rpc.Round.Cards.P4C1, 4, rpc.Signal.End))
	Cards.Layout.Objects[8].Refresh()

	Cards.Layout.Objects[9] = P4_b(Is_In(rpc.Round.Cards.P4C2, 4, rpc.Signal.End))
	Cards.Layout.Objects[9].Refresh()

	Cards.Layout.Objects[10] = P5_a(Is_In(rpc.Round.Cards.P5C1, 5, rpc.Signal.End))
	Cards.Layout.Objects[10].Refresh()

	Cards.Layout.Objects[11] = P5_b(Is_In(rpc.Round.Cards.P5C2, 5, rpc.Signal.End))
	Cards.Layout.Objects[11].Refresh()

	Cards.Layout.Objects[12] = P6_a(Is_In(rpc.Round.Cards.P6C1, 6, rpc.Signal.End))
	Cards.Layout.Objects[12].Refresh()

	Cards.Layout.Objects[13] = P6_b(Is_In(rpc.Round.Cards.P6C2, 6, rpc.Signal.End))
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

// Sets bet amount and current bet readout
func ifBet(w, r uint64) {
	if w > 0 && r > 0 && !rpc.Signal.PlacedBet {
		float := float64(w) / 100000
		wager := strconv.FormatFloat(float, 'f', 1, 64)
		holdero.Table.BetEntry.SetText(wager)
		rpc.Display.Res = rpc.Round.Raisor + " Raised, " + wager + " to Call "
	} else if w > 0 && !rpc.Signal.PlacedBet {
		float := float64(w) / 100000
		wager := strconv.FormatFloat(float, 'f', 1, 64)
		holdero.Table.BetEntry.SetText(wager)
		rpc.Display.Res = rpc.Round.Bettor + " Bet " + wager
	} else if r > 0 && rpc.Signal.PlacedBet {
		float := float64(r) / 100000
		rasied := strconv.FormatFloat(float, 'f', 1, 64)
		holdero.Table.BetEntry.SetText(rasied)
		rpc.Display.Res = rpc.Round.Raisor + " Raised, " + rasied + " to Call"
	} else if w == 0 && !rpc.Signal.Bet {
		var float float64
		if rpc.Round.Ante == 0 {
			float = float64(rpc.Round.BB) / 100000
		} else {
			float = float64(rpc.Round.Ante) / 100000
		}
		this := strconv.FormatFloat(float, 'f', 1, 64)
		holdero.Table.BetEntry.SetText(this)
		if !rpc.Signal.Reveal {
			rpc.Display.Res = "Check or Bet"
			holdero.Table.BetEntry.Enable()
		}
	} else if !rpc.Signal.Deal {
		rpc.Display.Res = "Deal Hand"
	}

	holdero.Table.BetEntry.Refresh()
}

// Single shot triggering ifBet() on players turn
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

// Main dReams process loop
func fetch(quit chan struct{}) {
	time.Sleep(3 * time.Second)
	var ticker = time.NewTicker(3 * time.Second)
	var autoCF, autoD, autoB, trigger bool
	var skip, delay int
	for {
		select {
		case <-ticker.C: // do on interval
			if !dReams.configure {
				rpc.Ping()
				rpc.EchoWallet("dReams")
				rpc.GetBalance()
				go rpc.GetDreamsBalances()
				rpc.GetWalletHeight("dReams")
				if !rpc.Signal.Startup {
					menu.CheckConnection()
					menu.GnomonEndPoint()
					menu.GnomonState(isWindows(), dReams.configure)
					background.Refresh()

					// Bacc
					if menu.Control.Dapp_list["Baccarat"] {
						rpc.FetchBaccSC()
						BaccRefresh()
					}

					// Holdero
					if menu.Control.Dapp_list["Holdero"] {
						rpc.FetchHolderoSC()
						if (rpc.Round.Turn == rpc.Round.ID && rpc.Wallet.Height > rpc.Signal.CHeight+4) ||
							(rpc.Round.Turn != rpc.Round.ID && rpc.Round.ID >= 1) || (!rpc.Signal.My_turn && rpc.Round.ID >= 1) {
							if rpc.Signal.Clicked {
								trigger = false
								autoCF = false
								autoD = false
								autoB = false
								rpc.Signal.Reveal = false
							}
							rpc.Signal.Clicked = false
						}

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
								if delay >= 15 || now > rpc.Round.Last+45 {
									delay = 0
									rpc.Round.Card_delay = false
								}
							} else {
								setHolderoLabel()
								holdero.GetUrls(rpc.Round.F_url, rpc.Round.B_url)
								rpc.Called(rpc.Round.Flop, rpc.Round.Wager)
								trigger = singleShot(rpc.Signal.My_turn, trigger)
								HolderoRefresh()
								if holdero.Settings.Auto_check && rpc.Signal.My_turn && !autoCF {
									if !rpc.Signal.Reveal && !rpc.Signal.End && !rpc.Round.LocalEnd {
										if rpc.Round.Cards.Local1 != "" {
											holdero.HolderoButtonBuffer()
											rpc.Check()
											H.TopLabel.Text = "Auto Check/Fold Tx Sent"
											H.TopLabel.Refresh()
											autoCF = true

											go func() {
												if !isWindows() {
													time.Sleep(500 * time.Millisecond)
													dReams.App.SendNotification(notification("dReams - Holdero", "Auto Check/Fold TX Sent", 9))
												}
											}()
										}
									}
								}

								if holdero.Settings.Auto_deal && rpc.Signal.My_turn && !autoD && rpc.GameIsActive() {
									if !rpc.Signal.Reveal && !rpc.Signal.End && !rpc.Round.LocalEnd {
										if rpc.Round.Cards.Local1 == "" {
											autoD = true
											go func() {
												time.Sleep(2100 * time.Millisecond)
												holdero.HolderoButtonBuffer()
												rpc.DealHand()
												H.TopLabel.Text = "Auto Deal Tx Sent"
												H.TopLabel.Refresh()

												if !isWindows() {
													time.Sleep(300 * time.Millisecond)
													dReams.App.SendNotification(notification("dReams - Holdero", "Auto Deal TX Sent", 9))
												}
											}()
										}
									}
								}

								if rpc.Odds.Run && rpc.Signal.My_turn && !autoB && rpc.GameIsActive() {
									if !rpc.Signal.Reveal && !rpc.Signal.End && !rpc.Round.LocalEnd {
										if rpc.Round.Cards.Local1 != "" {
											autoB = true
											go func() {
												time.Sleep(2100 * time.Millisecond)
												holdero.HolderoButtonBuffer()
												odds, future := rpc.MakeOdds()
												rpc.BetLogic(odds, future, true)
												H.TopLabel.Text = "Auto Bet Tx Sent"
												H.TopLabel.Refresh()

												if !isWindows() {
													time.Sleep(300 * time.Millisecond)
													dReams.App.SendNotification(notification("dReams - Holdero", "Auto Bet TX Sent", 9))
												}
											}()
										}
									}
								}

								if rpc.Round.ID > 1 && rpc.Signal.My_turn && !rpc.Signal.End && !rpc.Round.LocalEnd {
									now := time.Now().Unix()
									if now > rpc.Round.Last+100 {
										holdero.Table.Warning.Show()
									} else {
										holdero.Table.Warning.Hide()
									}
								} else {
									holdero.Table.Warning.Hide()
								}

								skip = 0
							}
						} else {
							waitLabel()
							revealingKey()
							skip++
							if skip >= 20 {
								rpc.Signal.Clicked = false
								skip = 0
								trigger = false
								autoCF = false
								autoD = false
								autoB = false
								rpc.Signal.Reveal = false
							}
						}
					}

					// Tarot
					if menu.Control.Dapp_list["Iluma"] {
						rpc.FetchTarotSC()
						TarotRefresh()
					}

					// Betting
					if menu.Control.Dapp_list["dSports and dPredictions"] {
						if offset%5 == 0 {
							SportsRefresh(dReams.sports)
						}

						S.RightLabel.SetText("dReams Balance: " + rpc.Display.Token_balance["dReams"] + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Display.Wallet_height)
						PredictionRefresh(dReams.predict)
					}

					// Menu
					go MenuRefresh(dReams.menu)

					offset++
					if offset == 21 {
						offset = 0
					}
				}

				if rpc.Daemon.Connect {
					if rpc.Signal.Startup {
						go refreshPriceDisplay(true)
					}

					rpc.Signal.Startup = false
				}
			}
		case <-quit: // exit loop
			log.Println("[dReams] Closing")
			ticker.Stop()
			return
		}
	}
}

// Sets Holdero table info labels
func setHolderoLabel() {
	H.TopLabel.Text = rpc.Display.Res
	H.LeftLabel.SetText("Seats: " + rpc.Display.Seats + "      Pot: " + rpc.Display.Pot + "      Blinds: " + rpc.Display.Blinds + "      Ante: " + rpc.Display.Ante + "      Dealer: " + rpc.Display.Dealer)
	if rpc.Round.Asset {
		if rpc.Round.Tourney {
			H.RightLabel.SetText(rpc.Display.Readout + "      Player ID: " + rpc.Display.PlayerId + "      Chip Balance: " + rpc.Display.Token_balance["Tournament"] + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Display.Wallet_height)
		} else {
			asset_name := rpc.GetAssetSCIDName(rpc.Round.AssetID)
			H.RightLabel.SetText(rpc.Display.Readout + "      Player ID: " + rpc.Display.PlayerId + "      " + asset_name + " Balance: " + rpc.Display.Token_balance[asset_name] + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Display.Wallet_height)
		}
	} else {
		H.RightLabel.SetText(rpc.Display.Readout + "      Player ID: " + rpc.Display.PlayerId + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Display.Wallet_height)
	}

	if rpc.Signal.Contract {
		holdero.Settings.SharedOn.Enable()
	} else {
		holdero.Settings.SharedOn.Disable()
	}

	H.TopLabel.Refresh()
	H.LeftLabel.Refresh()
	H.RightLabel.Refresh()
}

// Holdero label for waiting for block
func waitLabel() {
	H.TopLabel.Text = ""
	if rpc.Round.Asset {
		if rpc.Round.Tourney {
			H.RightLabel.SetText("Wait for Block" + "      Player ID: " + rpc.Display.PlayerId + "      Chip Balance: " + rpc.Display.Token_balance["Tournament"] + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Display.Wallet_height)
		} else {
			asset_name := rpc.GetAssetSCIDName(rpc.Round.AssetID)
			H.RightLabel.SetText("Wait for Block" + "      Player ID: " + rpc.Display.PlayerId + "      " + asset_name + " Balance: " + rpc.Display.Token_balance[asset_name] + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Display.Wallet_height)
		}

	} else {
		H.RightLabel.SetText("Wait for Block" + "      Player ID: " + rpc.Display.PlayerId + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Display.Wallet_height)
	}
	H.TopLabel.Refresh()
	H.RightLabel.Refresh()
}

// Refresh all Holdero gui objects
func HolderoRefresh() {
	go holdero.ShowAvatar(dReams.holdero)
	go refreshHolderoCards(rpc.Round.Cards.Local1, rpc.Round.Cards.Local2)
	if !rpc.Signal.Clicked {
		if rpc.Round.ID == 0 && rpc.Wallet.Connect {
			if rpc.Signal.Sit {
				holdero.Table.Sit.Hide()
			} else {
				holdero.Table.Sit.Show()
			}
			holdero.Table.Leave.Hide()
			holdero.Table.Deal.Hide()
			holdero.Table.Check.Hide()
			holdero.Table.Bet.Hide()
			holdero.Table.BetEntry.Hide()
		} else if !rpc.Signal.End && !rpc.Signal.Reveal && rpc.Signal.My_turn && rpc.Wallet.Connect {
			if rpc.Signal.Sit {
				holdero.Table.Sit.Hide()
			} else {
				holdero.Table.Sit.Show()
			}

			if rpc.Signal.Leave {
				holdero.Table.Leave.Hide()
			} else {
				holdero.Table.Leave.Show()
			}

			if rpc.Signal.Deal {
				holdero.Table.Deal.Hide()
			} else {
				holdero.Table.Deal.Show()
			}

			holdero.Table.Check.SetText(rpc.Display.C_Button)
			holdero.Table.Bet.SetText(rpc.Display.B_Button)
			if rpc.Signal.Bet {
				holdero.Table.Check.Hide()
				holdero.Table.Bet.Hide()
				holdero.Table.BetEntry.Hide()
			} else {
				holdero.Table.Check.Show()
				holdero.Table.Bet.Show()
				holdero.Table.BetEntry.Show()
			}

			if !rpc.Round.Notified {
				if !isWindows() {
					dReams.App.SendNotification(notification("dReams - Holdero", "Your Turn", 0))
				}
			}
		} else {
			if rpc.Signal.Sit {
				holdero.Table.Sit.Hide()
			} else if !rpc.Signal.Sit && rpc.Wallet.Connect {
				holdero.Table.Sit.Show()
			}
			holdero.Table.Leave.Hide()
			holdero.Table.Deal.Hide()
			holdero.Table.Check.Hide()
			holdero.Table.Bet.Hide()
			holdero.Table.BetEntry.Hide()

			if !rpc.Signal.My_turn && !rpc.Signal.End && !rpc.Round.LocalEnd {
				rpc.Display.Res = ""
				rpc.Round.Notified = false
			}
		}
	}

	if dReams.menu_tabs.contracts {
		if offset%3 == 0 {
			go menu.GetTableStats(rpc.Round.Contract, false)
		}
	}

	go func() {
		refreshHolderoPlayers()
		H.DApp.Refresh()
	}()
}

// Refresh Holdero player names and avatars
func refreshHolderoPlayers() {
	H.Back.Objects[0] = holdero.HolderoTable(bundle.ResourcePokerTablePng)
	H.Back.Objects[0].Refresh()

	H.Back.Objects[1] = holdero.Player1_label(bundle.ResourceUnknownAvatarPng, bundle.ResourceAvatarFramePng, bundle.ResourceTurnFramePng)
	H.Back.Objects[1].Refresh()

	H.Back.Objects[2] = holdero.Player2_label(bundle.ResourceUnknownAvatarPng, bundle.ResourceAvatarFramePng, bundle.ResourceTurnFramePng)
	H.Back.Objects[2].Refresh()

	H.Back.Objects[3] = holdero.Player3_label(bundle.ResourceUnknownAvatarPng, bundle.ResourceAvatarFramePng, bundle.ResourceTurnFramePng)
	H.Back.Objects[3].Refresh()

	H.Back.Objects[4] = holdero.Player4_label(bundle.ResourceUnknownAvatarPng, bundle.ResourceAvatarFramePng, bundle.ResourceTurnFramePng)
	H.Back.Objects[4].Refresh()

	H.Back.Objects[5] = holdero.Player5_label(bundle.ResourceUnknownAvatarPng, bundle.ResourceAvatarFramePng, bundle.ResourceTurnFramePng)
	H.Back.Objects[5].Refresh()

	H.Back.Objects[6] = holdero.Player6_label(bundle.ResourceUnknownAvatarPng, bundle.ResourceAvatarFramePng, bundle.ResourceTurnFramePng)
	H.Back.Objects[6].Refresh()

	H.Back.Refresh()
}

// Reveal key notification and display
func revealingKey() {
	if rpc.Signal.Reveal && rpc.Signal.My_turn && !rpc.Signal.End {
		if !rpc.Round.Notified {
			rpc.Display.Res = "Revealing Key"
			H.TopLabel.Text = rpc.Display.Res
			H.TopLabel.Refresh()
			if !isWindows() {
				dReams.App.SendNotification(notification("dReams - Holdero", "Revealing Key", 0))
			}
		}
	}
}

// Refresh all Baccarat objects
func BaccRefresh() {
	asset_name := rpc.GetAssetSCIDName(rpc.Bacc.AssetID)
	B.LeftLabel.SetText("Total Hands Played: " + rpc.Display.Total_w + "      Player Wins: " + rpc.Display.Player_w + "      Ties: " + rpc.Display.Ties + "      Banker Wins: " + rpc.Display.Banker_w + "      Min Bet is " + rpc.Display.BaccMin + " dReams, Max Bet is " + rpc.Display.BaccMax)
	B.RightLabel.SetText(asset_name + " Balance: " + rpc.Display.Token_balance[asset_name] + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Display.Wallet_height)

	if !rpc.Bacc.Display {
		B.Front.Objects[0] = clearBaccCards()
		rpc.FetchBaccHand(rpc.Bacc.Last)
		if rpc.Bacc.Found {
			B.Front.Objects[0] = showBaccCards()
		}
		B.Front.Objects[0].Refresh()
	}

	if rpc.Wallet.Height > rpc.Bacc.CHeight+3 && !rpc.Bacc.Found {
		rpc.Display.BaccRes = ""
		baccarat.BaccBuffer(false)
	}

	B.Back.Objects[1].(*canvas.Text).Text = rpc.Display.BaccRes
	B.Back.Objects[1].Refresh()

	B.DApp.Refresh()

	if rpc.Bacc.Found && !rpc.Bacc.Notified {
		if !isWindows() {
			dReams.App.SendNotification(notification("dReams - Baccarat", rpc.Display.BaccRes, 1))
		}
	}
}

// Refresh all dPrediction objects
func PredictionRefresh(tab bool) {
	if tab {
		if offset%5 == 0 {
			go prediction.SetPredictionInfo(prediction.Predict.Contract)
		}

		if offset == 11 || prediction.Predict.Prices.Text == "" {
			go prediction.SetPredictionPrices(rpc.Daemon.Connect)
		}

		P.RightLabel.SetText("dReams Balance: " + rpc.Display.Token_balance["dReams"] + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Display.Wallet_height)

		if menu.CheckActivePrediction(prediction.Predict.Contract) {
			go prediction.ShowPredictionControls()
		} else {
			prediction.DisablePreditions(true)
		}
	}
}

// Refresh all dSports objects
func SportsRefresh(tab bool) {
	if tab {
		go prediction.SetSportsInfo(prediction.Sports.Contract)
	}
}

// Refresh all Tarot objects
func TarotRefresh() {
	T.LeftLabel.SetText("Total Readings: " + rpc.Display.Readings + "      Click your card for Iluma reading")
	T.RightLabel.SetText("dReams Balance: " + rpc.Display.Token_balance["dReams"] + "      Dero Balance: " + rpc.Display.Dero_balance + "      Height: " + rpc.Display.Wallet_height)

	if !rpc.Tarot.Display {
		rpc.FetchTarotReading(rpc.Tarot.Last)
		tarot.Iluma.Box.Refresh()
		if rpc.Tarot.Found {
			rpc.Tarot.Display = true
			tarot.Iluma.Label.SetText("")
			if rpc.Tarot.Num == 3 {
				tarot.Iluma.Card1.Objects[1] = TarotCard(rpc.Tarot.Card1)
				tarot.Iluma.Card2.Objects[1] = TarotCard(rpc.Tarot.Card2)
				tarot.Iluma.Card3.Objects[1] = TarotCard(rpc.Tarot.Card3)
			} else {
				tarot.Iluma.Card1.Objects[1] = TarotCard(0)
				tarot.Iluma.Card2.Objects[1] = TarotCard(rpc.Tarot.Card1)
				tarot.Iluma.Card3.Objects[1] = TarotCard(0)
			}
			tarot.TarotBuffer(false)
			tarot.Iluma.Box.Refresh()
		}
	}

	if rpc.Wallet.Height > rpc.Tarot.CHeight+3 {
		tarot.TarotBuffer(false)
	}

	T.DApp.Refresh()

	if rpc.Tarot.Found && !rpc.Tarot.Notified {
		if !isWindows() {
			dReams.App.SendNotification(notification("dReams - Iluma", "Your Reading has Arrvied", 2))
		}
	}
}

// Refresh Gnomon height display
func refreshGnomonDisplay(index_height, c int) {
	if c == 1 {
		height := " Gnomon Height: " + strconv.Itoa(index_height)
		menu.Assets.Gnomes_height.Text = (height)
		menu.Assets.Gnomes_height.Refresh()
	} else {
		menu.Assets.Gnomes_height.Text = (" Gnomon Height: 0")
		menu.Assets.Gnomes_height.Refresh()
	}
}

// Refresh indexed asset count
func refreshIndexDisplay(c bool) {
	if c {
		scids := " Indexed SCIDs: " + strconv.Itoa(int(menu.Gnomes.SCIDS))
		menu.Assets.Gnomes_index.Text = (scids)
		menu.Assets.Gnomes_index.Refresh()
	} else {
		menu.Assets.Gnomes_index.Text = (" Indexed SCIDs: 0")
		menu.Assets.Gnomes_index.Refresh()
	}
}

// Refresh daemon height display
func refreshDaemonDisplay(c bool) {
	if c && rpc.Daemon.Connect {
		dHeight := rpc.DaemonHeight("dReams", rpc.Daemon.Rpc)
		d := strconv.Itoa(int(dHeight))
		menu.Assets.Daem_height.Text = (" Daemon Height: " + d)
		menu.Assets.Daem_height.Refresh()
	} else {
		menu.Assets.Daem_height.Text = (" Daemon Height: 0")
		menu.Assets.Daem_height.Refresh()
	}
}

// Refresh menu wallet display
func refreshWalletDisplay(c bool) {
	if c {
		menu.Assets.Wall_height.Text = (" Wallet Height: " + rpc.Display.Wallet_height)
		menu.Assets.Wall_height.Refresh()
		menu.Assets.Dreams_bal.Text = (" dReams Balance: " + rpc.Display.Token_balance["dReams"])
		menu.Assets.Dreams_bal.Refresh()
		menu.Assets.Dero_bal.Text = (" Dero Balance: " + rpc.Display.Dero_balance)
		menu.Assets.Dero_bal.Refresh()
	} else {
		menu.Assets.Wall_height.Text = (" Wallet Height: 0")
		menu.Assets.Wall_height.Refresh()
		menu.Assets.Dreams_bal.Text = (" dReams Balance: 0")
		menu.Assets.Dreams_bal.Refresh()
		menu.Assets.Dero_bal.Text = (" Dero Balance: 0")
		menu.Assets.Dero_bal.Refresh()
	}
}

// Refresh current Dero-USDT price
func refreshPriceDisplay(c bool) {
	if c && rpc.Daemon.Connect {
		_, price := holdero.GetPrice("DERO-USDT")
		menu.Assets.Dero_price.Text = (" Dero Price: $" + price)
		menu.Assets.Dero_price.Refresh()
	} else {
		menu.Assets.Dero_price.Text = (" Dero Price: $")
		menu.Assets.Dero_price.Refresh()
	}
}

// Refresh all menu gui objects
func MenuRefresh(tab bool) {
	if tab && menu.Gnomes.Init {
		var index int
		if !menu.GnomonClosing() && menu.FastSynced() {
			index = int(menu.Gnomes.Indexer.LastIndexedHeight)
		}

		if !menu.FastSynced() {
			menu.Assets.Gnomes_sync.Text = (" Gnomon Syncing... ")
			menu.Assets.Gnomes_sync.Refresh()
		} else {
			if !menu.GnomonClosing() {
				menu.Assets.Gnomes_sync.Text = ("")
				menu.Assets.Gnomes_sync.Refresh()
			}
		}
		go refreshGnomonDisplay(index, 1)
		go refreshIndexDisplay(true)

		if rpc.Daemon.Connect {
			go refreshDaemonDisplay(true)
		}

		if offset == 20 {
			go refreshPriceDisplay(true)
		}

		if dReams.menu_tabs.market && !isWindows() && !menu.Exit_signal {
			menu.FindNfaListings(nil)
		}
	}

	if rpc.Daemon.Connect {
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
		menu.Market.Viewing_coll = ""
	}
}

// Switch triggered when main tab changes
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
			holdero.Settings.FaceSelect.Enable()
			holdero.Settings.BackSelect.Enable()
		}
		go MenuRefresh(dReams.menu)
	case "Holdero":
		dReams.menu = false
		dReams.holdero = true
		dReams.bacc = false
		dReams.predict = false
		dReams.sports = false
		dReams.tarot = false
	case "Baccarat":
		dReams.menu = false
		dReams.holdero = false
		dReams.bacc = true
		dReams.predict = false
		dReams.sports = false
		dReams.tarot = false
		go func() {
			baccarat.GetBaccTables()
			BaccRefresh()
			if rpc.Wallet.Connect && rpc.Bacc.Display {
				baccarat.BaccBuffer(false)
			}
		}()
	case "Predict":
		dReams.menu = false
		dReams.holdero = false
		dReams.bacc = false
		dReams.predict = true
		dReams.sports = false
		dReams.tarot = false
		go func() {
			menu.PopulatePredictions(nil)
		}()
		PredictionRefresh(dReams.predict)
	case "Sports":
		dReams.menu = false
		dReams.holdero = false
		dReams.bacc = false
		dReams.predict = false
		dReams.sports = true
		dReams.tarot = false
		go menu.PopulateSports(nil)
	case "Iluma":
		dReams.menu = false
		dReams.holdero = false
		dReams.bacc = false
		dReams.predict = false
		dReams.sports = false
		dReams.tarot = true
		if rpc.Tarot.Display {
			tarot.TarotBuffer(false)
		}
	}
}

// Switch triggered when menu tab changes
func MenuTab(ti *container.TabItem) {
	switch ti.Text {
	case "Wallet":
		ti.Content.(*container.Split).Leading.(*container.Split).Trailing.Refresh()
		dReams.menu_tabs.wallet = true
		dReams.menu_tabs.assets = false
		dReams.menu_tabs.market = false
	case "Assets":
		dReams.menu_tabs.wallet = false
		dReams.menu_tabs.assets = true
		dReams.menu_tabs.market = false
		menu.Control.Viewing_asset = ""
		menu.Assets.Asset_list.UnselectAll()
	case "Market":
		dReams.menu_tabs.wallet = false
		dReams.menu_tabs.assets = false
		dReams.menu_tabs.market = true
		go menu.FindNfaListings(nil)
		menu.Market.Cancel_button.Hide()
		menu.Market.Close_button.Hide()
		menu.Market.Auction_list.Refresh()
		menu.Market.Buy_list.Refresh()
	}
}

// Switch triggered when Holdero contracts tab changes
func MenuContractTab(ti *container.TabItem) {
	switch ti.Text {
	case "Tables":
		if rpc.Daemon.Connect {
			go menu.CreateTableList(false, nil)
		}

	default:
	}
}

// Switch triggered when dPrediction tab changes
func PredictTab(ti *container.TabItem) {
	switch ti.Text {
	case "Contracts":
		go menu.PopulatePredictions(nil)
	default:
	}
}

// Switch triggered when Tarot tab changes
func TarotTab(ti *container.TabItem) {
	switch ti.Text {
	case "Iluma":
		tarot.Iluma.Actions.Hide()
	case "Reading":
		tarot.Iluma.Actions.Show()

	default:
	}
}

// Set and revert main window fullscreen mode
func FullScreenSet() fyne.CanvasObject {
	var button *widget.Button
	button = widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "viewFullScreen"), func() {
		if dReams.Window.FullScreen() {
			button.Icon = fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "viewFullScreen")
			dReams.Window.SetFullScreen(false)
		} else {
			button.Icon = fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "viewRestore")
			dReams.Window.SetFullScreen(true)
		}
	})

	button.Importance = widget.LowImportance

	cont := container.NewHBox(layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), container.NewVBox(button), layout.NewSpacer())

	return cont
}
