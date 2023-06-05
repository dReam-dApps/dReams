package rpc

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"
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

type displayStrings struct {
	Seats    string
	Pot      string
	Blinds   string
	Ante     string
	Dealer   string
	PlayerId string
	Readout  string
	B_Button string
	C_Button string
	Res      string

	Total_w  string
	Player_w string
	Banker_w string
	Ties     string
	BaccMax  string
	BaccMin  string
	BaccRes  string

	Prediction string
	P_feed     string

	Game    string
	S_count string
	League  string
	S_end   string
	TeamA   string
	TeamB   string

	Readings string

	Dero_balance  string
	Token_balance map[string]string
	Wallet_height string
}

type hashValue struct {
	Local1 string
	Local2 string
	P1C1   string
	P1C2   string
	P2C1   string
	P2C2   string
	P3C1   string
	P3C2   string
	P4C1   string
	P4C2   string
	P5C1   string
	P5C2   string
	P6C1   string
	P6C2   string

	Key1 string
	Key2 string
	Key3 string
	Key4 string
	Key5 string
	Key6 string
}

type holderoValues struct {
	Version   int
	Cards     hashValue
	Contract  string
	ID        int
	Players   int
	Turn      int
	Last      int64
	Pot       uint64
	BB        uint64
	SB        uint64
	Ante      uint64
	Wager     uint64
	Raised    uint64
	Flop1     int
	Flop2     int
	Flop3     int
	TurnCard  int
	RiverCard int
	SC_seed   string
	Winner    string
	Flop      bool
	LocalEnd  bool
	F1        bool
	F2        bool
	F3        bool
	F4        bool
	F5        bool
	F6        bool
	Asset     bool
	Printed   bool
	Notified  bool
	Tourney   bool
	AssetID   string
	Face      string
	Back      string
	F_url     string
	B_url     string
	P1_name   string
	P2_name   string
	P3_name   string
	P4_name   string
	P5_name   string
	P6_name   string
	P1_url    string
	P2_url    string
	P3_url    string
	P4_url    string
	P5_url    string
	P6_url    string
	Bettor    string
	Raiser    string

	Winning_hand  []int
	First_try     bool
	Card_delay    bool
	Local_trigger bool
	Flop_trigger  bool
	Turn_trigger  bool
	River_trigger bool
}

type baccValues struct {
	P_card1  int
	P_card2  int
	P_card3  int
	B_card1  int
	B_card2  int
	B_card3  int
	CHeight  int
	MinBet   float64
	MaxBet   float64
	AssetID  string
	Contract string
	Last     string
	Found    bool
	Display  bool
	Notified bool
}

type tarotValues struct {
	Card1    int
	Card2    int
	Card3    int
	CHeight  int
	Num      int
	Last     string
	Found    bool
	Display  bool
	Notified bool
}

type predictionValues struct {
	Init   bool
	Amount uint64
	Buffer int64
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

// Returns string value of atomic unit
func fromAtomic(v interface{}) string {
	var value float64

	switch v := v.(type) {
	case uint64:
		value = float64(v)
	case float64:
		value = v
	}

	str := fmt.Sprintf("%.5f", value/100000)

	return str
}

// Returns atomic value of string rounded to one decimal place
func ToAtomicOne(v string) uint64 {
	f, err := strconv.ParseFloat(v, 64)

	if err != nil {
		log.Println("[ToAtomicOne]", err)
		return 0
	}

	ratio := math.Pow(10, float64(1))
	rf := math.Round(f*ratio) / ratio

	u := uint64(math.Round(rf * 100000))

	return u
}

// Make blinds display string
func blindString(b, s interface{}) string {
	bb := b.(float64) / 100000
	sb := s.(float64) / 100000

	x := fmt.Sprintf("%.5f", bb)
	y := fmt.Sprintf("%.5f", sb)

	return x + " / " + y
}

// Returns value plus one as string
func addOne(v interface{}) string {
	value := int(v.(float64) + 1)
	str := strconv.Itoa(value)

	return str
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
		switch p.(float64) {
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

	av := fromHextoString(str)

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

// Check if player Id matches Wallet.idHash
func checkPlayerId(one, two, three, four, five, six string) string {
	var id string
	if Wallet.idHash == one {
		id = "1"
		Round.ID = 1
	} else if Wallet.idHash == two {
		id = "2"
		Round.ID = 2
	} else if Wallet.idHash == three {
		id = "3"
		Round.ID = 3
	} else if Wallet.idHash == four {
		id = "4"
		Round.ID = 4
	} else if Wallet.idHash == five {
		id = "5"
		Round.ID = 5
	} else if Wallet.idHash == six {
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
			Wallet.KeyLock = false
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
	s := int(seats.(float64))
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
		Round.Flop1 = int(f1.(float64))
		if !Round.Flop_trigger {
			Round.Card_delay = true
		}
		Round.Flop_trigger = true
	} else {
		Round.Flop1 = 0
		Round.Flop_trigger = false
	}

	if f2 != nil {
		Round.Flop2 = int(f2.(float64))
	} else {
		Round.Flop2 = 0
	}

	if f3 != nil {
		Round.Flop3 = int(f3.(float64))
	} else {
		Round.Flop3 = 0
	}

	if t != nil {
		Round.TurnCard = int(t.(float64))
		if !Round.Turn_trigger {
			Round.Card_delay = true
		}
		Round.Turn_trigger = true
	} else {
		Round.TurnCard = 0
		Round.Turn_trigger = false
	}

	if r != nil {
		Round.RiverCard = int(r.(float64))
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
			Round.Cards.Local1 = a1.(string)
			Round.Cards.Local2 = a2.(string)
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
		Round.Cards.P1C1 = a1.(string)
		Round.Cards.P1C2 = a2.(string)
	} else {
		Round.Cards.P1C1 = ""
		Round.Cards.P1C2 = ""
	}

	if Round.ID == 2 {
		if b1 != nil {
			Round.Cards.Local1 = b1.(string)
			Round.Cards.Local2 = b2.(string)
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
		Round.Cards.P2C1 = b1.(string)
		Round.Cards.P2C2 = b2.(string)
	} else {
		Round.Cards.P2C1 = ""
		Round.Cards.P2C2 = ""
	}

	if Round.ID == 3 {
		if c1 != nil {
			Round.Cards.Local1 = c1.(string)
			Round.Cards.Local2 = c2.(string)
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
		Round.Cards.P3C1 = c1.(string)
		Round.Cards.P3C2 = c2.(string)
	} else {
		Round.Cards.P3C1 = ""
		Round.Cards.P3C2 = ""
	}

	if Round.ID == 4 {
		if d1 != nil {
			Round.Cards.Local1 = d1.(string)
			Round.Cards.Local2 = d2.(string)
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
		Round.Cards.P4C1 = d1.(string)
		Round.Cards.P4C2 = d2.(string)
	} else {
		Round.Cards.P4C1 = ""
		Round.Cards.P4C2 = ""
	}

	if Round.ID == 5 {
		if e1 != nil {
			Round.Cards.Local1 = e1.(string)
			Round.Cards.Local2 = e2.(string)
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
		Round.Cards.P5C1 = e1.(string)
		Round.Cards.P5C2 = e2.(string)
	} else {
		Round.Cards.P5C1 = ""
		Round.Cards.P5C2 = ""
	}

	if Round.ID == 6 {
		if f1 != nil {
			Round.Cards.Local1 = f1.(string)
			Round.Cards.Local2 = f2.(string)
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
		Round.Cards.P6C1 = f1.(string)
		Round.Cards.P6C2 = f2.(string)
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
func turnReadout(t interface{}) string {
	var s string
	if t != nil {
		if addOne(t) == Display.PlayerId {
			s = "Your Turn"
		} else {
			if addOne(t) == "1" {
				s = "Player 1's Turn"
			} else if addOne(t) == "2" {
				s = "Player 2's Turn"
			} else if addOne(t) == "3" {
				s = "Player 3's Turn"
			} else if addOne(t) == "4" {
				s = "Player 4's Turn"
			} else if addOne(t) == "5" {
				s = "Player 5's Turn"
			} else if addOne(t) == "6" {
				s = "Player 6's Turn"
			}
		}
	}
	return s
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
	if int(s.(float64)) >= 2 {
		if p1 != nil {
			a = int(p1.(float64))
		} else {
			who = "Player1"
			display = Round.P1_name
		}
		if p2 != nil {
			b = int(p2.(float64))
		} else {
			who = "Player2"
			display = Round.P2_name
		}
	}
	if int(s.(float64)) >= 3 {
		if p3 != nil {
			c = int(p3.(float64))
		} else {
			who = "Player3"
			display = Round.P3_name
		}
	}

	if int(s.(float64)) >= 4 {
		if p4 != nil {
			d = int(p4.(float64))
		} else {
			who = "Player4"
			display = Round.P4_name
		}
	}

	if int(s.(float64)) >= 5 {
		if p5 != nil {
			e = int(p5.(float64))
		} else {
			who = "Player5"
			display = Round.P5_name
		}
	}

	if int(s.(float64)) >= 6 {
		if p6 != nil {
			f = int(p6.(float64))
		} else {
			who = "Player6"
			display = Round.P6_name
		}
	}

	i := a + b + c + d + e + f

	if 1+i-int(s.(float64)) == 0 {
		Round.LocalEnd = true
		Round.Winner = who
		Display.Res = display + " Wins, All Players Have Folded"
		if GameIsActive() && Round.Pot > 0 {
			if !Signal.Log {
				Signal.Log = true
				AddLog(Display.Res)
			}

			updateStatsWins(Round.Pot, who, true)
		}
	}
}

// Payout routine when all Holdero players have folded
func allFoldedWinner() {
	if Round.ID == 1 {
		if Round.LocalEnd && !Signal.Startup {
			if !Signal.Paid {
				Signal.Paid = true
				go func() {
					time.Sleep(time.Duration(Times.Delay) * time.Second)
					retry := 0
					for retry < 4 {
						tx := PayOut(Round.Winner)
						time.Sleep(time.Second)
						retry += ConfirmTxRetry(tx, "Holdero", 36)
					}
				}()
			}
		}
	}
}

// If Holdero showdown, trigger the hand ranker routine
func winningHand(e interface{}) {
	if e != nil && !Signal.Startup && !Round.LocalEnd {
		go func() {
			getHands(StringToInt(Display.Seats))
		}()
	}
}

// Convert a millisecond string to time.Time
func MsToTime(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(0, msInt*int64(time.Millisecond)), nil
}

// Return team param string for dSports
func TeamReturn(t int) string {
	var team string
	switch t {
	case 0:
		team = "team_a"
	case 1:
		team = "team_b"
	default:
		team = "none"

	}

	return team
}

// Find Tarot card from hash value
func findTarotCard(hash interface{}) int {
	if hash != nil {
		for i := 1; i < 79; i++ {
			finder := strconv.Itoa(i)
			card := sha256.Sum256([]byte(finder))
			str := hex.EncodeToString(card[:])

			if str == hash.(string) {
				return i
			}
		}
	}
	return 0
}
