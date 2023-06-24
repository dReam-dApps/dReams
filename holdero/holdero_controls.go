package holdero

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dReam-dApps/dReams/rpc"
)

type times struct {
	Kick       int
	Delay      int
	Kick_block int
}

type playerId struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}
type CardSpecs struct {
	Faces struct {
		Name string `json:"Name"`
		Url  string `json:"Url"`
	} `json:"Faces"`
	Backs struct {
		Name string `json:"Name"`
		Url  string `json:"Url"`
	} `json:"Backs"`
}

type TableSpecs struct {
	MaxBet float64 `json:"Maxbet"`
	MinBuy float64 `json:"Minbuy"`
	MaxBuy float64 `json:"Maxbuy"`
	Time   int     `json:"Time"`
}

type signals struct {
	Startup   bool
	Contract  bool
	Deal      bool
	Bet       bool
	Called    bool
	Reveal    bool
	End       bool
	Sit       bool
	Leave     bool
	In1       bool
	In2       bool
	In3       bool
	In4       bool
	In5       bool
	In6       bool
	Out1      bool
	My_turn   bool
	PlacedBet bool
	Paid      bool
	Log       bool
	Odds      bool
	Clicked   bool
	CHeight   int
}

var Times times
var Signal signals

// Make blinds display string
func blindString(b, s interface{}) string {
	if bb, ok := b.(float64); ok {
		if sb, ok := s.(float64); ok {
			return fmt.Sprintf("%.5f / %.5f", bb/100000, sb/100000)
		}
	}

	return "? / ?"
}

// If Holdero table is closed set vars accordingly
func closedTable() {
	Round.Winner = ""
	Round.Players = 0
	Round.Pot = 0
	Round.ID = 0
	Round.Tourney = false
	Round.P1_url = ""
	Round.P2_url = ""
	Round.P3_url = ""
	Round.P4_url = ""
	Round.P5_url = ""
	Round.P6_url = ""
	Round.P1_name = ""
	Round.P2_name = ""
	Round.P3_name = ""
	Round.P4_name = ""
	Round.P5_name = ""
	Round.P6_name = ""
	Round.Bettor = ""
	Round.Raiser = ""
	Round.Turn = 0
	Round.Last = 0
	Round.Local_trigger = false
	Round.Flop_trigger = false
	Round.Turn_trigger = false
	Round.River_trigger = false
	Signal.Out1 = false
	Signal.Sit = true
	Signal.In1 = false
	Signal.In2 = false
	Signal.In3 = false
	Signal.In4 = false
	Signal.In5 = false
	Signal.In6 = false
	Display.Seats = ""
	Display.Pot = ""
	Display.Blinds = ""
	Display.Ante = ""
	Display.Dealer = ""
	Display.PlayerId = ""
}

// Clear a single players name and avatar values
func singleNameClear(p int) {
	switch p {
	case 1:
		Round.P1_name = ""
		Round.P1_url = ""
	case 2:
		Round.P2_name = ""
		Round.P2_url = ""
	case 3:
		Round.P3_name = ""
		Round.P3_url = ""
	case 4:
		Round.P4_name = ""
		Round.P4_url = ""
	case 5:
		Round.P5_name = ""
		Round.P5_url = ""
	case 6:
		Round.P6_name = ""
		Round.P6_url = ""
	default:

	}
}

// Returns name of Holdero player who bet
func findBettor(p interface{}) string {
	if p != nil {
		switch rpc.Float64Type(p) {
		case 0:
			if Round.P6_name != "" && !Round.F6 {
				return Round.P6_name
			} else if Round.P5_name != "" && !Round.F5 {
				return Round.P5_name
			} else if Round.P4_name != "" && !Round.F4 {
				return Round.P4_name
			} else if Round.P3_name != "" && !Round.F3 {
				return Round.P3_name
			} else if Round.P2_name != "" && !Round.F2 {
				return Round.P2_name
			}
		case 1:
			if Round.P1_name != "" && !Round.F1 {
				return Round.P1_name
			} else if Round.P6_name != "" && !Round.F6 {
				return Round.P6_name
			} else if Round.P5_name != "" && !Round.F5 {
				return Round.P5_name
			} else if Round.P4_name != "" && !Round.F4 {
				return Round.P4_name
			} else if Round.P3_name != "" && !Round.F3 {
				return Round.P3_name
			}
		case 2:
			if Round.P2_name != "" && !Round.F2 {
				return Round.P2_name
			} else if Round.P1_name != "" && !Round.F1 {
				return Round.P1_name
			} else if Round.P6_name != "" && !Round.F6 {
				return Round.P6_name
			} else if Round.P5_name != "" && !Round.F5 {
				return Round.P5_name
			} else if Round.P4_name != "" && !Round.F4 {
				return Round.P4_name
			}
		case 3:
			if Round.P3_name != "" && !Round.F3 {
				return Round.P3_name
			} else if Round.P2_name != "" && !Round.F2 {
				return Round.P2_name
			} else if Round.P1_name != "" && !Round.F1 {
				return Round.P1_name
			} else if Round.P6_name != "" && !Round.F6 {
				return Round.P6_name
			} else if Round.P5_name != "" && !Round.F5 {
				return Round.P5_name
			}
		case 4:
			if Round.P4_name != "" && !Round.F4 {
				return Round.P4_name
			} else if Round.P3_name != "" && !Round.F3 {
				return Round.P3_name
			} else if Round.P2_name != "" && !Round.F2 {
				return Round.P2_name
			} else if Round.P1_name != "" && !Round.F1 {
				return Round.P1_name
			} else if Round.P6_name != "" && !Round.F6 {
				return Round.P6_name
			}
		case 5:
			if Round.P5_name != "" && !Round.F5 {
				return Round.P5_name
			} else if Round.P4_name != "" && !Round.F4 {
				return Round.P4_name
			} else if Round.P3_name != "" && !Round.F3 {
				return Round.P3_name
			} else if Round.P2_name != "" && !Round.F2 {
				return Round.P2_name
			} else if Round.P1_name != "" && !Round.F1 {
				return Round.P1_name
			}
		default:
			return ""
		}
	}

	return ""
}

// Gets Holdero player name and avatar, returns player Id string
func getAvatar(p int, id interface{}) string {
	if id == nil {
		singleNameClear(p)
		return "nil"
	}

	str := fmt.Sprint(id)

	if len(str) == 64 {
		return str
	}

	av := rpc.HexToString(str)

	var player playerId

	if err := json.Unmarshal([]byte(av), &player); err != nil {
		log.Println("[getAvatar]", err)
		return ""
	}

	switch p {
	case 1:
		Round.P1_name = player.Name
		Round.P1_url = player.Avatar
	case 2:
		Round.P2_name = player.Name
		Round.P2_url = player.Avatar
	case 3:
		Round.P3_name = player.Name
		Round.P3_url = player.Avatar
	case 4:
		Round.P4_name = player.Name
		Round.P4_url = player.Avatar
	case 5:
		Round.P5_name = player.Name
		Round.P5_url = player.Avatar
	case 6:
		Round.P6_name = player.Name
		Round.P6_url = player.Avatar
	}

	return player.Id
}

// Check if player Id matches rpc.Wallet.IdHash
func checkPlayerId(one, two, three, four, five, six string) string {
	var id string
	if rpc.Wallet.IdHash == one {
		id = "1"
		Round.ID = 1
	} else if rpc.Wallet.IdHash == two {
		id = "2"
		Round.ID = 2
	} else if rpc.Wallet.IdHash == three {
		id = "3"
		Round.ID = 3
	} else if rpc.Wallet.IdHash == four {
		id = "4"
		Round.ID = 4
	} else if rpc.Wallet.IdHash == five {
		id = "5"
		Round.ID = 5
	} else if rpc.Wallet.IdHash == six {
		id = "6"
		Round.ID = 6
	} else {
		id = ""
		Round.ID = 0
	}

	return id
}

// Set Holdero name signals for when player is at table
func setHolderoName(one, two, three, four, five, six interface{}) {
	if one != nil {
		Signal.In1 = true
	} else {
		Signal.In1 = false
	}

	if two != nil {
		Signal.In2 = true
	} else {
		Signal.In2 = false
	}

	if three != nil {
		Signal.In3 = true
	} else {
		Signal.In3 = false
	}

	if four != nil {
		Signal.In4 = true
	} else {
		Signal.In4 = false
	}

	if five != nil {
		Signal.In5 = true
	} else {
		Signal.In5 = false
	}

	if six != nil {
		Signal.In6 = true
	} else {
		Signal.In6 = false
	}
}

// When Holdero pot is empty set vars accordingly
func potIsEmpty(pot uint64) {
	if pot == 0 {
		if !Signal.My_turn {
			rpc.Wallet.KeyLock = false
		}
		Round.Winning_hand = []int{}
		Round.Flop1 = 0
		Round.Flop2 = 0
		Round.Flop3 = 0
		Round.TurnCard = 0
		Round.RiverCard = 0
		Round.LocalEnd = false
		Round.Wager = 0
		Round.Raised = 0
		Round.Winner = ""
		Round.Printed = false
		Round.Cards.Local1 = ""
		Round.Cards.Local2 = ""
		Round.Cards.P1C1 = ""
		Round.Cards.P1C2 = ""
		Round.Cards.P2C1 = ""
		Round.Cards.P2C2 = ""
		Round.Cards.P3C1 = ""
		Round.Cards.P3C2 = ""
		Round.Cards.P4C1 = ""
		Round.Cards.P4C2 = ""
		Round.Cards.P5C1 = ""
		Round.Cards.P5C2 = ""
		Round.Cards.P6C1 = ""
		Round.Cards.P6C2 = ""
		Round.Cards.Key1 = ""
		Round.Cards.Key2 = ""
		Round.Cards.Key3 = ""
		Round.Cards.Key4 = ""
		Round.Cards.Key5 = ""
		Round.Cards.Key6 = ""
		Signal.Called = false
		Signal.PlacedBet = false
		Signal.Reveal = false
		Signal.End = false
		Signal.Paid = false
		Signal.Log = false
		Signal.Odds = false
		Display.Res = ""
		Round.Bettor = ""
		Round.Raiser = ""
		Round.Local_trigger = false
		Round.Flop_trigger = false
		Round.Turn_trigger = false
		Round.River_trigger = false
	}
}

// Sets Holdero sit signal if table has open seats
func tableOpen(seats, full, two, three, four, five, six interface{}) {
	players := 1
	if two != nil {
		players++
	}

	if three != nil {
		players++
	}

	if four != nil {
		players++
	}

	if five != nil {
		players++
	}

	if six != nil {
		players++
	}

	if Signal.Out1 {
		players--
	}

	Round.Players = players

	if Round.ID > 1 {
		Signal.Sit = true
		return
	}
	s := rpc.IntType(seats)
	if s >= 2 && two == nil && Round.ID != 1 {
		Signal.Sit = false
	}

	if s >= 3 && three == nil && Round.ID != 1 {
		Signal.Sit = false
	}

	if s >= 4 && four == nil && Round.ID != 1 {
		Signal.Sit = false
	}

	if s >= 5 && five == nil && Round.ID != 1 {
		Signal.Sit = false
	}

	if s == 6 && six == nil && Round.ID != 1 {
		Signal.Sit = false
	}

	if full != nil {
		Signal.Sit = true
	}
}

// Gets Holdero community card values
func getCommCardValues(f1, f2, f3, t, r interface{}) {
	if f1 != nil {
		Round.Flop1 = rpc.IntType(f1)
		if !Round.Flop_trigger {
			Round.Card_delay = true
		}
		Round.Flop_trigger = true
	} else {
		Round.Flop1 = 0
		Round.Flop_trigger = false
	}

	if f2 != nil {
		Round.Flop2 = rpc.IntType(f2)
	} else {
		Round.Flop2 = 0
	}

	if f3 != nil {
		Round.Flop3 = rpc.IntType(f3)
	} else {
		Round.Flop3 = 0
	}

	if t != nil {
		Round.TurnCard = rpc.IntType(t)
		if !Round.Turn_trigger {
			Round.Card_delay = true
		}
		Round.Turn_trigger = true
	} else {
		Round.TurnCard = 0
		Round.Turn_trigger = false
	}

	if r != nil {
		Round.RiverCard = rpc.IntType(r)
		if !Round.River_trigger {
			Round.Card_delay = true
		}
		Round.River_trigger = true
	} else {
		Round.RiverCard = 0
		Round.River_trigger = false
	}
}

// Gets Holdero player card hash values
func getPlayerCardValues(a1, a2, b1, b2, c1, c2, d1, d2, e1, e2, f1, f2 interface{}) {
	if Round.ID == 1 {
		if a1 != nil {
			Round.Cards.Local1 = fmt.Sprint(a1)
			Round.Cards.Local2 = fmt.Sprint(a2)
			if !Round.Local_trigger {
				Round.Card_delay = true
			}
			Round.Local_trigger = true
		} else {
			Round.Cards.Local1 = ""
			Round.Cards.Local2 = ""
			Round.Local_trigger = false
		}
	}

	if a1 != nil {
		Round.Cards.P1C1 = fmt.Sprint(a1)
		Round.Cards.P1C2 = fmt.Sprint(a2)
	} else {
		Round.Cards.P1C1 = ""
		Round.Cards.P1C2 = ""
	}

	if Round.ID == 2 {
		if b1 != nil {
			Round.Cards.Local1 = fmt.Sprint(b1)
			Round.Cards.Local2 = fmt.Sprint(b2)
			if !Round.Local_trigger {
				Round.Card_delay = true
			}
			Round.Local_trigger = true
		} else {
			Round.Cards.Local1 = ""
			Round.Cards.Local2 = ""
			Round.Local_trigger = false
		}
	}

	if b1 != nil {
		Round.Cards.P2C1 = fmt.Sprint(b1)
		Round.Cards.P2C2 = fmt.Sprint(b2)
	} else {
		Round.Cards.P2C1 = ""
		Round.Cards.P2C2 = ""
	}

	if Round.ID == 3 {
		if c1 != nil {
			Round.Cards.Local1 = fmt.Sprint(c1)
			Round.Cards.Local2 = fmt.Sprint(c2)
			if !Round.Local_trigger {
				Round.Card_delay = true
			}
			Round.Local_trigger = true
		} else {
			Round.Cards.Local1 = ""
			Round.Cards.Local2 = ""
			Round.Local_trigger = false
		}
	}

	if c1 != nil {
		Round.Cards.P3C1 = fmt.Sprint(c1)
		Round.Cards.P3C2 = fmt.Sprint(c2)
	} else {
		Round.Cards.P3C1 = ""
		Round.Cards.P3C2 = ""
	}

	if Round.ID == 4 {
		if d1 != nil {
			Round.Cards.Local1 = fmt.Sprint(d1)
			Round.Cards.Local2 = fmt.Sprint(d2)
			if !Round.Local_trigger {
				Round.Card_delay = true
			}
			Round.Local_trigger = true
		} else {
			Round.Cards.Local1 = ""
			Round.Cards.Local2 = ""
			Round.Local_trigger = false
		}
	}

	if d1 != nil {
		Round.Cards.P4C1 = fmt.Sprint(d1)
		Round.Cards.P4C2 = fmt.Sprint(d2)
	} else {
		Round.Cards.P4C1 = ""
		Round.Cards.P4C2 = ""
	}

	if Round.ID == 5 {
		if e1 != nil {
			Round.Cards.Local1 = fmt.Sprint(e1)
			Round.Cards.Local2 = fmt.Sprint(e2)
			if !Round.Local_trigger {
				Round.Card_delay = true
			}
			Round.Local_trigger = true
		} else {
			Round.Cards.Local1 = ""
			Round.Cards.Local2 = ""
			Round.Local_trigger = false
		}
	}

	if e1 != nil {
		Round.Cards.P5C1 = fmt.Sprint(e1)
		Round.Cards.P5C2 = fmt.Sprint(e2)
	} else {
		Round.Cards.P5C1 = ""
		Round.Cards.P5C2 = ""
	}

	if Round.ID == 6 {
		if f1 != nil {
			Round.Cards.Local1 = fmt.Sprint(f1)
			Round.Cards.Local2 = fmt.Sprint(f2)
			if !Round.Local_trigger {
				Round.Card_delay = true
			}
			Round.Local_trigger = true
		} else {
			Round.Cards.Local1 = ""
			Round.Cards.Local2 = ""
			Round.Local_trigger = false
		}
	}

	if f1 != nil {
		Round.Cards.P6C1 = fmt.Sprint(f1)
		Round.Cards.P6C2 = fmt.Sprint(f2)
	} else {
		Round.Cards.P6C1 = ""
		Round.Cards.P6C2 = ""
	}

	if Round.ID == 0 {
		Round.Cards.Local1 = ""
		Round.Cards.Local2 = ""
	}
}

// If Holdero player has called set Signal.Called, and reset Signal.PlacedBet when no wager
func Called(fb bool, w uint64) {
	if w == 0 {
		if fb {
			Signal.Called = true
		} else {
			Signal.Called = false
		}

		if Signal.Called {
			Round.Raised = 0
			Round.Wager = 0
			Signal.PlacedBet = false
			Signal.Called = false
		}

		Display.B_Button = "Bet"
		Display.C_Button = "Check"
	}
}

// Holdero players turn display string
func turnReadout(t interface{}) (turn string) {
	if t != nil {
		switch rpc.AddOne(t) {
		case Display.PlayerId:
			turn = "Your Turn"
		case "1":
			turn = "Player 1's Turn"
		case "2":
			turn = "Player 2's Turn"
		case "3":
			turn = "Player 3's Turn"
		case "4":
			turn = "Player 4's Turn"
		case "5":
			turn = "Player 5's Turn"
		case "6":
			turn = "Player 6's Turn"
		}
	}

	return
}

// Sets Holdero action signals
func setSignals(pot uint64, one interface{}) {
	if !Round.LocalEnd {
		if len(Round.Cards.Local1) != 64 {
			Signal.Deal = false
			Signal.Leave = false
			Signal.Bet = true
		} else {
			Signal.Deal = true
			Signal.Leave = true
			if pot != 0 {
				Signal.Bet = false
			} else {
				Signal.Bet = true
			}
		}
	} else {
		Signal.Deal = true
		Signal.Leave = true
		Signal.Bet = true
	}

	if Round.ID > 1 {
		Signal.Sit = true
	}

	if Round.ID == 1 {
		if one != nil {
			Signal.Sit = false
		} else {
			Signal.Sit = true
		}
	}
}

// If Holdero player has folded, set Round folded bools and clear cards
func hasFolded(one, two, three, four, five, six interface{}) {
	if one != nil {
		Round.F1 = true
		Round.Cards.P1C1 = ""
		Round.Cards.P1C2 = ""
	} else {
		Round.F1 = false
	}

	if two != nil {
		Round.F2 = true
		Round.Cards.P2C1 = ""
		Round.Cards.P2C2 = ""
	} else {
		Round.F2 = false
	}

	if three != nil {
		Round.F3 = true
		Round.Cards.P3C1 = ""
		Round.Cards.P3C2 = ""
	} else {
		Round.F3 = false
	}

	if four != nil {
		Round.F4 = true
		Round.Cards.P4C1 = ""
		Round.Cards.P4C2 = ""
	} else {
		Round.F4 = false
	}

	if five != nil {
		Round.F5 = true
		Round.Cards.P5C1 = ""
		Round.Cards.P5C2 = ""
	} else {
		Round.F5 = false
	}

	if six != nil {
		Round.F6 = true
		Round.Cards.P6C1 = ""
		Round.Cards.P6C2 = ""
	} else {
		Round.F6 = false
	}
}

// Determine if all players have folded and trigger payout
func allFolded(p1, p2, p3, p4, p5, p6, s interface{}) {
	var a, b, c, d, e, f int
	var who string
	var display string
	seats := rpc.IntType(s)
	if seats >= 2 {
		if p1 != nil {
			a = rpc.IntType(p1)
		} else {
			who = "Player1"
			display = Round.P1_name
		}
		if p2 != nil {
			b = rpc.IntType(p2)
		} else {
			who = "Player2"
			display = Round.P2_name
		}
	}
	if seats >= 3 {
		if p3 != nil {
			c = rpc.IntType(p3)
		} else {
			who = "Player3"
			display = Round.P3_name
		}
	}

	if seats >= 4 {
		if p4 != nil {
			d = rpc.IntType(p4)
		} else {
			who = "Player4"
			display = Round.P4_name
		}
	}

	if seats >= 5 {
		if p5 != nil {
			e = rpc.IntType(p5)
		} else {
			who = "Player5"
			display = Round.P5_name
		}
	}

	if seats >= 6 {
		if p6 != nil {
			f = rpc.IntType(p6)
		} else {
			who = "Player6"
			display = Round.P6_name
		}
	}

	i := a + b + c + d + e + f

	if 1+i-seats == 0 {
		Round.LocalEnd = true
		Round.Winner = who
		Display.Res = display + " Wins, All Players Have Folded"
		if GameIsActive() && Round.Pot > 0 {
			if !Signal.Log {
				Signal.Log = true
				rpc.AddLog(Display.Res)
			}

			updateStatsWins(Round.Pot, who, true)
		}
	}
}

// Payout routine when all Holdero players have folded
func allFoldedWinner() {
	if Round.ID == 1 {
		if Round.LocalEnd && !rpc.Startup {
			if !Signal.Paid {
				Signal.Paid = true
				go func() {
					time.Sleep(time.Duration(Times.Delay) * time.Second)
					retry := 0
					for retry < 4 {
						tx := PayOut(Round.Winner)
						time.Sleep(time.Second)
						retry += rpc.ConfirmTxRetry(tx, "Holdero", 36)
					}
				}()
			}
		}
	}
}

// If Holdero showdown, trigger the hand ranker routine
func winningHand(e interface{}) {
	if e != nil && !rpc.Startup && !Round.LocalEnd {
		go func() {
			getHands(rpc.StringToInt(Display.Seats))
		}()
	}
}
