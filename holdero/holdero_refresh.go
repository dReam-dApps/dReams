package holdero

import (
	"log"
	"strconv"
	"time"

	"fyne.io/fyne/v2/container"
	dreams "github.com/SixofClubsss/dReams"
	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/dwidget"
	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"
)

// Sets bet amount and current bet readout
func ifBet(w, r uint64) {
	if w > 0 && r > 0 && !Signal.PlacedBet {
		float := float64(w) / 100000
		wager := strconv.FormatFloat(float, 'f', 1, 64)
		Table.BetEntry.SetText(wager)
		Display.Res = Round.Raiser + " Raised, " + wager + " to Call "
	} else if w > 0 && !Signal.PlacedBet {
		float := float64(w) / 100000
		wager := strconv.FormatFloat(float, 'f', 1, 64)
		Table.BetEntry.SetText(wager)
		Display.Res = Round.Bettor + " Bet " + wager
	} else if r > 0 && Signal.PlacedBet {
		float := float64(r) / 100000
		raised := strconv.FormatFloat(float, 'f', 1, 64)
		Table.BetEntry.SetText(raised)
		Display.Res = Round.Raiser + " Raised, " + raised + " to Call"
	} else if w == 0 && !Signal.Bet {
		var float float64
		if Round.Ante == 0 {
			float = float64(Round.BB) / 100000
		} else {
			float = float64(Round.Ante) / 100000
		}
		this := strconv.FormatFloat(float, 'f', 1, 64)
		Table.BetEntry.SetText(this)
		if !Signal.Reveal {
			Display.Res = "Check or Bet"
			Table.BetEntry.Enable()
		}
	} else if !Signal.Deal {
		Display.Res = "Deal Hand"
	}

	Table.BetEntry.Refresh()
}

// Single shot triggering ifBet() on players turn
func singleShot(turn, trigger bool) bool {
	if turn && !trigger {
		ifBet(Round.Wager, Round.Raised)
		return true
	}

	if !turn {
		return false
	} else {
		return turn
	}
}

// Main Holdero process
func fetch(H *dwidget.DreamsItems, d dreams.DreamsObject) {
	time.Sleep(3 * time.Second)
	var autoCF, autoD, autoB, trigger bool
	var skip, delay, offset int
	for {
		select {
		case <-d.Receive():
			if !rpc.Wallet.IsConnected() || !rpc.Daemon.IsConnected() {
				disableActions()
				Signal.Contract = false
				Settings.Check.SetChecked(false)
				DisableOwnerControls(true)
				Settings.Tables = []string{}
				Settings.Owned = []string{}
				Settings.Synced = false
				d.WorkDone()
				continue
			}

			if !Settings.Synced && menu.GnomonScan(d.Configure) {
				log.Println("[Holdero] Syncing")
				createTableList()
				Settings.Synced = true
			}

			// Holdero
			Poker.Stats_box = *container.NewVBox(Table.Stats.Name, Table.Stats.Desc, Table.Stats.Version, Table.Stats.Last, Table.Stats.Seats, tableIcon(bundle.ResourceAvatarFramePng))
			Poker.Stats_box.Refresh()

			FetchHolderoSC()
			if (Round.Turn == Round.ID && rpc.Wallet.Height > Signal.CHeight+4) ||
				(Round.Turn != Round.ID && Round.ID >= 1) || (!Signal.My_turn && Round.ID >= 1) {
				if Signal.Clicked {
					trigger = false
					autoCF = false
					autoD = false
					autoB = false
					Signal.Reveal = false
				}
				Signal.Clicked = false
			}

			if !Signal.Clicked {
				if Round.First_try {
					Round.First_try = false
					delay = 0
					Round.Card_delay = false
					go refreshHolderoPlayers(H)
				}

				if Round.Card_delay {
					now := time.Now().Unix()
					delay++
					if delay >= 17 || now > Round.Last+60 {
						delay = 0
						Round.Card_delay = false
					}
				} else {
					setHolderoLabel(H)
					GetUrls(Round.F_url, Round.B_url)
					Called(Round.Flop, Round.Wager)
					trigger = singleShot(Signal.My_turn, trigger)
					holderoRefresh(H, d, offset)
					if Settings.Auto_check && Signal.My_turn && !autoCF {
						if !Signal.Reveal && !Signal.End && !Round.LocalEnd {
							if Round.Cards.Local1 != "" {
								ActionBuffer()
								Check()
								H.TopLabel.Text = "Auto Check/Fold Tx Sent"
								H.TopLabel.Refresh()
								autoCF = true

								go func() {
									if !d.IsWindows() {
										time.Sleep(500 * time.Millisecond)
										Round.Notified = d.Notification("dReams - Holdero", "Auto Check/Fold TX Sent")
									}
								}()
							}
						}
					}

					if Settings.Auto_deal && Signal.My_turn && !autoD && GameIsActive() {
						if !Signal.Reveal && !Signal.End && !Round.LocalEnd {
							if Round.Cards.Local1 == "" {
								autoD = true
								go func() {
									time.Sleep(2100 * time.Millisecond)
									ActionBuffer()
									DealHand()
									H.TopLabel.Text = "Auto Deal Tx Sent"
									H.TopLabel.Refresh()

									if !d.IsWindows() {
										time.Sleep(300 * time.Millisecond)
										Round.Notified = d.Notification("dReams - Holdero", "Auto Deal TX Sent")
									}
								}()
							}
						}
					}

					if Odds.Run && Signal.My_turn && !autoB && GameIsActive() {
						if !Signal.Reveal && !Signal.End && !Round.LocalEnd {
							if Round.Cards.Local1 != "" {
								autoB = true
								go func() {
									time.Sleep(2100 * time.Millisecond)
									ActionBuffer()
									odds, future := MakeOdds()
									BetLogic(odds, future, true)
									H.TopLabel.Text = "Auto Bet Tx Sent"
									H.TopLabel.Refresh()

									if !d.IsWindows() {
										time.Sleep(300 * time.Millisecond)
										Round.Notified = d.Notification("dReams - Holdero", "Auto Bet TX Sent")
									}
								}()
							}
						}
					}

					if Round.ID > 1 && Signal.My_turn && !Signal.End && !Round.LocalEnd {
						now := time.Now().Unix()
						if now > Round.Last+100 {
							Table.Warning.Show()
						} else {
							Table.Warning.Hide()
						}
					} else {
						Table.Warning.Hide()
					}

					skip = 0
				}
			} else {
				waitLabel(H)
				revealingKey(H, d)
				skip++
				if skip >= 20 {
					Signal.Clicked = false
					skip = 0
					trigger = false
					autoCF = false
					autoD = false
					autoB = false
					Signal.Reveal = false
				}
			}

			offset++
			if offset >= 21 {
				offset = 0
			}

			d.WorkDone()
		case <-d.CloseDapp():
			log.Println("[Holdero] Closed")
			return
		}
	}
}

// Do when disconnected
func Disconnected(b bool) {
	if b {
		Round.ID = 0
		Display.PlayerId = ""
		Odds.Run = false
		Faces.Select.Options = []string{"Light", "Dark"}
		Backs.Select.Options = []string{"Light", "Dark"}
		Settings.AvatarSelect.Options = []string{"None"}
		Faces.URL = ""
		Backs.URL = ""
		Settings.AvatarUrl = ""
		Faces.Select.SetSelectedIndex(0)
		Backs.Select.SetSelectedIndex(0)
		Settings.AvatarSelect.SetSelectedIndex(0)
		Faces.Select.Refresh()
		Backs.Select.Refresh()
		Settings.AvatarSelect.Refresh()
		DisableHolderoTools()
	}
}

func disableActions() {
	Poker.Holdero_unlock.Hide()
	Poker.Holdero_new.Hide()
	Table.Tournament.Hide()
	Poker.Holdero_unlock.Refresh()
	Poker.Holdero_new.Refresh()
	Table.Tournament.Refresh()
}

// Disable Holdero owner actions
func DisableOwnerControls(d bool) {
	if d {
		Poker.owner.owners_left.Hide()
		Poker.owner.owners_mid.Hide()
	} else {
		Poker.owner.owners_left.Show()
		Poker.owner.owners_mid.Show()
	}

	Poker.owner.owners_left.Refresh()
	Poker.owner.owners_mid.Refresh()
}

// Sets Holdero table info labels
func setHolderoLabel(H *dwidget.DreamsItems) {
	H.TopLabel.Text = Display.Res
	H.LeftLabel.SetText("Seats: " + Display.Seats + "      Pot: " + Display.Pot + "      Blinds: " + Display.Blinds + "      Ante: " + Display.Ante + "      Dealer: " + Display.Dealer)
	if Round.Asset {
		if Round.Tourney {
			H.RightLabel.SetText(Display.Readout + "      Player ID: " + Display.PlayerId + "      Chip Balance: " + rpc.DisplayBalance("Tournament") + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Display.Wallet_height)
		} else {
			asset_name := rpc.GetAssetSCIDName(Round.AssetID)
			H.RightLabel.SetText(Display.Readout + "      Player ID: " + Display.PlayerId + "      " + asset_name + " Balance: " + rpc.DisplayBalance(asset_name) + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Display.Wallet_height)
		}
	} else {
		H.RightLabel.SetText(Display.Readout + "      Player ID: " + Display.PlayerId + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Display.Wallet_height)
	}

	if Signal.Contract {
		Settings.SharedOn.Enable()
	} else {
		Settings.SharedOn.Disable()
	}

	H.TopLabel.Refresh()
	H.LeftLabel.Refresh()
	H.RightLabel.Refresh()
}

// Holdero label for waiting for block
func waitLabel(H *dwidget.DreamsItems) {
	H.TopLabel.Text = ""
	if Round.Asset {
		if Round.Tourney {
			H.RightLabel.SetText("Wait for Block" + "      Player ID: " + Display.PlayerId + "      Chip Balance: " + rpc.DisplayBalance("Tournament") + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Display.Wallet_height)
		} else {
			asset_name := rpc.GetAssetSCIDName(Round.AssetID)
			H.RightLabel.SetText("Wait for Block" + "      Player ID: " + Display.PlayerId + "      " + asset_name + " Balance: " + rpc.DisplayBalance(asset_name) + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Display.Wallet_height)
		}

	} else {
		H.RightLabel.SetText("Wait for Block" + "      Player ID: " + Display.PlayerId + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Display.Wallet_height)
	}
	H.TopLabel.Refresh()
	H.RightLabel.Refresh()
}

// Refresh all Holdero gui objects
func holderoRefresh(h *dwidget.DreamsItems, d dreams.DreamsObject, offset int) {
	go ShowAvatar(d.Holdero)
	go refreshHolderoCards(Round.Cards.Local1, Round.Cards.Local2, d.Window)
	if !Signal.Clicked {
		if Round.ID == 0 && rpc.Wallet.IsConnected() {
			if Signal.Sit {
				Table.Sit.Hide()
			} else {
				Table.Sit.Show()
			}
			Table.Leave.Hide()
			Table.Deal.Hide()
			Table.Check.Hide()
			Table.Bet.Hide()
			Table.BetEntry.Hide()
		} else if !Signal.End && !Signal.Reveal && Signal.My_turn && rpc.Wallet.IsConnected() {
			if Signal.Sit {
				Table.Sit.Hide()
			} else {
				Table.Sit.Show()
			}

			if Signal.Leave {
				Table.Leave.Hide()
			} else {
				Table.Leave.Show()
			}

			if Signal.Deal {
				Table.Deal.Hide()
			} else {
				Table.Deal.Show()
			}

			Table.Check.SetText(Display.C_Button)
			Table.Bet.SetText(Display.B_Button)
			if Signal.Bet {
				Table.Check.Hide()
				Table.Bet.Hide()
				Table.BetEntry.Hide()
			} else {
				Table.Check.Show()
				Table.Bet.Show()
				Table.BetEntry.Show()
			}

			if !Round.Notified {
				if !d.IsWindows() {
					Round.Notified = d.Notification("dReams - Holdero", "Your Turn")
				}
			}
		} else {
			if Signal.Sit {
				Table.Sit.Hide()
			} else if !Signal.Sit && rpc.Wallet.IsConnected() {
				Table.Sit.Show()
			}
			Table.Leave.Hide()
			Table.Deal.Hide()
			Table.Check.Hide()
			Table.Bet.Hide()
			Table.BetEntry.Hide()

			if !Signal.My_turn && !Signal.End && !Round.LocalEnd {
				Display.Res = ""
				Round.Notified = false
			}
		}
	}

	// put back
	if d.Menu_tabs.Contracts {
		if offset%3 == 0 {
			go getTableStats(Round.Contract, false)
		}
	}
	offset++

	go func() {
		refreshHolderoPlayers(h)
		h.DApp.Refresh()
	}()
}

// Refresh Holdero player names and avatars
func refreshHolderoPlayers(H *dwidget.DreamsItems) {
	H.Back.Objects[0] = HolderoTable(bundle.ResourcePokerTablePng)
	H.Back.Objects[0].Refresh()

	H.Back.Objects[1] = Player1_label(bundle.ResourceUnknownAvatarPng, bundle.ResourceAvatarFramePng, bundle.ResourceTurnFramePng)
	H.Back.Objects[1].Refresh()

	H.Back.Objects[2] = Player2_label(bundle.ResourceUnknownAvatarPng, bundle.ResourceAvatarFramePng, bundle.ResourceTurnFramePng)
	H.Back.Objects[2].Refresh()

	H.Back.Objects[3] = Player3_label(bundle.ResourceUnknownAvatarPng, bundle.ResourceAvatarFramePng, bundle.ResourceTurnFramePng)
	H.Back.Objects[3].Refresh()

	H.Back.Objects[4] = Player4_label(bundle.ResourceUnknownAvatarPng, bundle.ResourceAvatarFramePng, bundle.ResourceTurnFramePng)
	H.Back.Objects[4].Refresh()

	H.Back.Objects[5] = Player5_label(bundle.ResourceUnknownAvatarPng, bundle.ResourceAvatarFramePng, bundle.ResourceTurnFramePng)
	H.Back.Objects[5].Refresh()

	H.Back.Objects[6] = Player6_label(bundle.ResourceUnknownAvatarPng, bundle.ResourceAvatarFramePng, bundle.ResourceTurnFramePng)
	H.Back.Objects[6].Refresh()

	H.Back.Refresh()
}

// Reveal key notification and display
func revealingKey(H *dwidget.DreamsItems, d dreams.DreamsObject) {
	if Signal.Reveal && Signal.My_turn && !Signal.End {
		if !Round.Notified {
			Display.Res = "Revealing Key"
			H.TopLabel.Text = Display.Res
			H.TopLabel.Refresh()

			if !d.IsWindows() {
				Round.Notified = d.Notification("dReams - Holdero", "Revealing Key")
			}
		}
	}
}
