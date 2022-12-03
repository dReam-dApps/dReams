package rpc

import (
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

type times struct {
	Kick       int
	Delay      int
	Kick_block int
}

type displayStrings struct {
	Seats    string
	Pot      string
	Blinds   string
	Ante     string
	Dealer   string
	Turn     string
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

	Preiction string
	P_amt     string
	P_end     string
	P_count   string
	P_pot     string
	P_up      string
	P_down    string
	P_played  string
	P_final   string
	P_mark    string
	P_feed    string
	P_txid    string

	Game    string
	S_count string
	League  string
	S_end   string
	TeamA   string
	TeamB   string
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
	Daemon    string
	Contract  string
	ID        int
	Last      int
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
}

type baccValues struct {
	P_card1  int
	P_card2  int
	P_card3  int
	B_card1  int
	B_card2  int
	B_card3  int
	CHeight  int
	Last     string
	Found    bool
	Display  bool
	Notified bool
}

type predictionValues struct {
	Init   bool
	Mark   bool
	Amount uint64
	Time_a int
	Time_b int
	Time_c int
}

type signals struct {
	Startup   bool
	Daemon    bool
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
	My_turn   bool
	PlacedBet bool
	Paid      bool
	Log       bool
	Clicked   bool
	CHeight   int
}

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

func ToAtomicOne(v string) uint64 {
	f, err := strconv.ParseFloat(v, 64)

	if err != nil {
		log.Println("To Atmoic Conversion Error", err)
		return 0
	}

	ratio := math.Pow(10, float64(1))
	rf := math.Round(f*ratio) / ratio

	u := uint64(math.Round(rf * 100000))

	return u
}

func blindString(b, s interface{}) string {
	bb := b.(float64) / 100000
	sb := s.(float64) / 100000

	x := fmt.Sprintf("%.5f", bb)
	y := fmt.Sprintf("%.5f", sb)

	return x + " / " + y
}

func addOne(v interface{}) string {
	value := int(v.(float64) + 1)
	str := strconv.Itoa(value)

	return str
}

func getAvatar(p int, id interface{}) string {
	if id == nil {
		return "nil"
	}

	str := fmt.Sprint(id)

	if len(str) == 64 {
		return str
	}

	av := fromHextoString(str)
	split := strings.Split(av, "_")
	switch p {
	case 1:
		if len(split) == 2 {
			Round.P1_name = split[1]
		} else if len(split) == 3 {
			Round.P1_name = split[1]
			Round.P1_url = split[2]
		}
	case 2:
		if len(split) == 2 {
			Round.P2_name = split[1]
		} else if len(split) == 3 {
			Round.P2_name = split[1]
			Round.P2_url = split[2]
		}
	case 3:
		if len(split) == 2 {
			Round.P3_name = split[1]
		} else if len(split) == 3 {
			Round.P3_name = split[1]
			Round.P3_url = split[2]
		}
	case 4:
		if len(split) == 2 {
			Round.P4_name = split[1]
		} else if len(split) == 3 {
			Round.P4_name = split[1]
			Round.P4_url = split[2]
		}
	case 5:
		if len(split) == 2 {
			Round.P5_name = split[1]
		} else if len(split) == 3 {
			Round.P5_name = split[1]
			Round.P5_url = split[2]
		}
	case 6:
		if len(split) == 2 {
			Round.P6_name = split[1]
		} else if len(split) == 3 {
			Round.P6_name = split[1]
			Round.P6_url = split[2]
		}
	}

	s := hex.EncodeToString([]byte(split[0]))

	return s
}

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

func potIsEmpty(pot uint64) {
	if pot == 0 {
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
		CardHash.Local1 = ""
		CardHash.Local2 = ""
		CardHash.P1C1 = ""
		CardHash.P1C2 = ""
		CardHash.P2C1 = ""
		CardHash.P2C2 = ""
		CardHash.P3C1 = ""
		CardHash.P3C2 = ""
		CardHash.P4C1 = ""
		CardHash.P4C2 = ""
		CardHash.P5C1 = ""
		CardHash.P5C2 = ""
		CardHash.P6C1 = ""
		CardHash.P6C2 = ""
		CardHash.Key1 = ""
		CardHash.Key2 = ""
		CardHash.Key3 = ""
		CardHash.Key4 = ""
		CardHash.Key5 = ""
		CardHash.Key6 = ""
		Signal.Called = false
		Signal.PlacedBet = false
		Signal.Reveal = false
		Signal.End = false
		Signal.Paid = false
		Signal.Log = false
		Display.Res = ""
	}
}

func tableClosed(seats interface{}) {
	if seats == nil || seats == 0 {
		Display.Seats = ""
		Display.Pot = ""
		Display.Blinds = ""
		Display.Ante = ""
		Display.Dealer = ""
		Display.Turn = ""
		Display.PlayerId = ""
		Round.Winner = ""
		Round.ID = 0
		Signal.Sit = true
	}
}

func tableOpen(seats, full, two, three, four, five, six interface{}) {
	s := int(seats.(float64))
	if s >= 2 && two == nil {
		Signal.Sit = false
	}

	if s >= 3 && three == nil {
		Signal.Sit = false
	}

	if s >= 4 && four == nil {
		Signal.Sit = false
	}

	if s >= 5 && five == nil {
		Signal.Sit = false
	}

	if s == 6 && six == nil {
		Signal.Sit = false
	}

	if full != nil {
		Signal.Sit = true
	}
}

func getCommCardValues(f1, f2, f3, t, r interface{}) {
	if f1 != nil {
		Round.Flop1 = int(f1.(float64))
	}

	if f2 != nil {
		Round.Flop2 = int(f2.(float64))
	}

	if f3 != nil {
		Round.Flop3 = int(f3.(float64))
	}

	if t != nil {
		Round.TurnCard = int(t.(float64))
	}

	if r != nil {
		Round.RiverCard = int(r.(float64))
	}
}

func getPlayerCardValues(a1, a2, b1, b2, c1, c2, d1, d2, e1, e2, f1, f2 interface{}) {
	if a1 != nil {
		if Round.ID == 1 {
			CardHash.Local1 = a1.(string)
			CardHash.Local2 = a2.(string)
		}
		CardHash.P1C1 = a1.(string)
		CardHash.P1C2 = a2.(string)
	}

	if b1 != nil {
		if Round.ID == 2 {
			CardHash.Local1 = b1.(string)
			CardHash.Local2 = b2.(string)
		}
		CardHash.P2C1 = b1.(string)
		CardHash.P2C2 = b2.(string)
	}

	if c1 != nil {
		if Round.ID == 3 {
			CardHash.Local1 = c1.(string)
			CardHash.Local2 = c2.(string)
		}
		CardHash.P3C1 = c1.(string)
		CardHash.P3C2 = c2.(string)
	}

	if d1 != nil {
		if Round.ID == 4 {
			CardHash.Local1 = d1.(string)
			CardHash.Local2 = d2.(string)
		}
		CardHash.P4C1 = d1.(string)
		CardHash.P4C2 = d2.(string)
	}

	if e1 != nil {
		if Round.ID == 5 {
			CardHash.Local1 = e1.(string)
			CardHash.Local2 = e2.(string)
		}
		CardHash.P5C1 = e1.(string)
		CardHash.P5C2 = e2.(string)
	}

	if f1 != nil {
		if Round.ID == 6 {
			CardHash.Local1 = f1.(string)
			CardHash.Local2 = f2.(string)
		}
		CardHash.P6C1 = f1.(string)
		CardHash.P6C2 = f2.(string)
	}

	if Round.ID == 0 {
		CardHash.Local1 = ""
		CardHash.Local2 = ""
	}
}

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

func setSignals(pot uint64, one interface{}) {
	if !Round.LocalEnd {
		if len(CardHash.Local1) != 64 {
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

func hasFolded(one, two, three, four, five, six interface{}) {
	if one != nil {
		Round.F1 = true
		CardHash.P1C1 = ""
		CardHash.P1C2 = ""
	} else {
		Round.F1 = false
	}

	if two != nil {
		Round.F2 = true
		CardHash.P2C1 = ""
		CardHash.P2C2 = ""
	} else {
		Round.F2 = false
	}

	if three != nil {
		Round.F3 = true
		CardHash.P3C1 = ""
		CardHash.P3C2 = ""
	} else {
		Round.F3 = false
	}

	if four != nil {
		Round.F4 = true
		CardHash.P4C1 = ""
		CardHash.P4C2 = ""
	} else {
		Round.F4 = false
	}

	if five != nil {
		Round.F5 = true
		CardHash.P5C1 = ""
		CardHash.P5C2 = ""
	} else {
		Round.F5 = false
	}

	if six != nil {
		Round.F6 = true
		CardHash.P6C1 = ""
		CardHash.P6C2 = ""
	} else {
		Round.F6 = false
	}
}

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
		if !Signal.Log {
			Signal.Log = true
			addLog(Display.Res)
		}

	}
}

func allFoldedWinner() {
	if Round.ID == 1 {
		if Round.LocalEnd && !Signal.Startup {
			if !Signal.Paid {
				Signal.Paid = true
				go func() {
					time.Sleep(time.Duration(Times.Delay) / 2 * time.Second)
					PayOut(Round.Winner)
				}()
			}
		}
	}
}

func winningHand(e interface{}) {
	if e != nil && !Signal.Startup && !Round.LocalEnd {
		go func() {
			getHands(StringToInt(Display.Seats))
		}()
	}
}

// / predictions
func MsToTime(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(0, msInt*int64(time.Millisecond)), nil
}

// / Sports
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
