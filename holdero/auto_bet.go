package holdero

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"time"

	"fyne.io/fyne/v2/widget"
	"github.com/SixofClubsss/dReams/rpc"
)

type holdero_stats struct {
	Win      int     `json:"win"`
	Lost     int     `json:"lost"`
	Fold     int     `json:"fold"`
	Push     int     `json:"push"`
	Wagered  float64 `json:"wagered"`
	Earnings float64 `json:"earnings"`
}

type bet_data struct {
	hole      [2]int
	used      [2]bool
	community []int
	hand      []int
	high_pair int
	high_card int
	low_pair  int
	Enabled   bool
	Run       bool
	Bot       Bot_config
	Label     *widget.Label
}

type Bot_config struct {
	Name   string        `json:"name"`
	Bet    [3]float64    `json:"bet"`
	Risk   [3]float64    `json:"risk"`
	Random [2]float64    `json:"random"`
	Luck   float64       `json:"luck"`
	Slow   float64       `json:"slow"`
	Aggr   float64       `json:"aggr"`
	Max    float64       `json:"max"`
	Stats  holdero_stats `json:"stats"`
}

type Player_stats struct {
	Player holdero_stats `json:"stats"`
	Bots   []Bot_config  `json:"bots"`
}

var Stats Player_stats
var Odds bet_data

func rPool(p int, n []int, c []int, cc [][]int) [][]int {
	if len(n) == 0 || p <= 0 {
		return cc
	}

	p--
	for i := range n {
		r := make([]int, len(c)+1)
		copy(r, c)
		r[len(r)-1] = n[i]
		if p == 0 {
			cc = append(cc, r)
		}
		cc = rPool(p, n[i+1:], r, cc)
	}
	return cc
}

// Find all combinations of n within p
func Pool(p int, n []int) [][]int {
	return rPool(p, n, nil, nil)
}

// Make possible cards that are left
func cardNumbers(hole []int) (numbers []int) {
	var used bool
	for i := 1; i < 53; i++ {
		for r := range hole {
			if i == hole[r] {
				used = true
				break
			}
		}

		if !used {
			numbers = append(numbers, i)
		}
		used = false
	}

	return
}

// Compares two hands of five cards
func compareTheseFive(cards []int) (rank int, hand []int, suits []int) {
	c1 := suitSplit(cards[0])
	c2 := suitSplit(cards[1])
	c3 := suitSplit(cards[2])
	c4 := suitSplit(cards[3])
	c5 := suitSplit(cards[4])
	hand = []int{c1[0], c2[0], c3[0], c4[0], c5[0]}
	suits = []int{c1[1], c2[1], c3[1], c4[1], c5[1]}
	rank = makeHand(hand, suits)

	return
}

// Compares two hands of two cards
func compareTheseTwo(cards []int) (rank float64, hand []int) {
	c1 := suitSplit(cards[0])
	c2 := suitSplit(cards[1])
	hand = []int{c1[0], c2[0]}
	suits := []int{c1[1], c2[1]}
	rank = makeHole(hand, suits)

	return
}

// If card matches a value in a combo
func match(c int, combo []int) (used bool) {
	for i := range combo {
		if c == combo[i] {
			used = true
		}
	}

	return
}

// If three cards match values in a combo
func threeMatch(cards, combo []int) bool {
	switch len(cards) {
	case 3:
		if match(cards[0], combo) && match(cards[1], combo) && match(cards[2], combo) {
			return true
		}
	case 4:
		if (match(cards[0], combo) && match(cards[1], combo) && match(cards[2], combo)) ||
			(match(cards[0], combo) && match(cards[1], combo) && match(cards[3], combo)) ||
			(match(cards[0], combo) && match(cards[2], combo) && match(cards[3], combo)) ||
			(match(cards[1], combo) && match(cards[2], combo) && match(cards[3], combo)) {
			return true
		}
	case 5:
		if (match(cards[0], combo) && match(cards[1], combo) && match(cards[2], combo)) ||
			(match(cards[0], combo) && match(cards[1], combo) && match(cards[3], combo)) ||
			(match(cards[0], combo) && match(cards[1], combo) && match(cards[4], combo)) ||
			(match(cards[0], combo) && match(cards[2], combo) && match(cards[3], combo)) ||
			(match(cards[0], combo) && match(cards[2], combo) && match(cards[4], combo)) ||
			(match(cards[0], combo) && match(cards[3], combo) && match(cards[4], combo)) ||
			(match(cards[1], combo) && match(cards[2], combo) && match(cards[3], combo)) ||
			(match(cards[1], combo) && match(cards[2], combo) && match(cards[4], combo)) ||
			(match(cards[1], combo) && match(cards[3], combo) && match(cards[4], combo)) ||
			(match(cards[2], combo) && match(cards[3], combo) && match(cards[4], combo)) {
			return true
		}
	default:

	}

	return false
}

// Make rank for local player current hand
func myHand(cards []int) int {
	rank := 100
	hand := []int{}
	p := Pool(5, cards)
	p1 := suitSplit(Odds.hole[0])
	p2 := suitSplit(Odds.hole[1])

	for i := range p {
		r, h, _ := compareTheseFive(p[i])
		if Odds.hand == nil {
			hand = p[i]
			Odds.hand = h
		}

		if r < rank {
			rank = r
			hand = p[i]
			Odds.hand = h
		} else if r == rank {
			var better bool
			var r ranker
			r.pc1 = p1
			r.pc2 = p2
			new := findBest(rank, Odds.hand, h, &r)
			for i := 0; i < 5; i++ {
				if new[i] != Odds.hand[i] {
					better = true
					break
				}
			}

			if better {
				hand = p[i]
				Odds.hand = new
			}
		}
	}

	Odds.used[0] = match(Odds.hole[0], hand)
	Odds.used[1] = match(Odds.hole[1], hand)

	oddsLog("[myHand]", fmt.Sprintln("You have", handToText(rank)))
	oddsLog("[myHand]", fmt.Sprintln("Hand", hand))
	oddsLog("[myHand]", fmt.Sprintln("Hole", Odds.hole))
	oddsLog("[myHand]", fmt.Sprintln("Used", Odds.used))

	Odds.high_pair, Odds.high_card, Odds.low_pair = getHighs(Odds.hand)

	return rank
}

// Make rank for hole card hand
func makeHole(h, s []int) (rank float64) {
	pHand := h
	pSuits := s

	sort.Ints(pHand)

	// chen formula
	switch pHand[1] {
	case 2:
		rank = rank + 1
	case 3:
		rank = rank + 1.5
	case 4:
		rank = rank + 2
	case 5:
		rank = rank + 2.5
	case 6:
		rank = rank + 3
	case 7:
		rank = rank + 3.5
	case 8:
		rank = rank + 4
	case 9:
		rank = rank + 4.5
	case 10:
		rank = rank + 5
	case 11:
		rank = rank + 6
	case 12:
		rank = rank + 7
	case 13:
		rank = rank + 8
	case 14:
		rank = rank + 10
	}

	// pairs
	if pHand[0] == pHand[1] {
		rank = rank * 2
		return
	}

	// suited
	if pSuits[0] == pSuits[1] {
		rank = rank + 2
	}

	// closeness
	switch pHand[1] - pHand[0] {
	case 1:
		if pHand[0] < 11 && pHand[1] < 11 {
			rank++
		}
	case 2:
		rank--
		if pHand[0] < 11 && pHand[1] < 11 {
			rank++
		}
	case 3:
		rank = rank - 2
	case 4:
		rank = rank - 4
	default:
		rank = rank - 5
	}

	return
}

// Evaluate the current board for player outs and bad odds situations
func lookAhead(odds, future float64, rank int, comm []int) (ahead float64) {
	fRank := float64(rank)
	var value, suit []int
	for z := range comm {
		split := suitSplit(comm[z])
		value = append(value, split[0])
		suit = append(suit, split[1])
	}

	sort.Ints(suit)
	sort.Ints(value)
	l := len(value)
	var sure bool
	var flush int
	var straight_outs, flush_outs float64
	if l < 5 {
		var v, s, run []int
		for y := range Odds.hole {
			split := suitSplit(Odds.hole[y])
			v = append(v, split[0])
			s = append(s, split[1])
		}

		full_v := append(v, value...)
		full_s := append(s, suit...)

		run, flush, straight_outs, flush_outs = findMyOuts(len(full_v), rank, full_v, full_s)

		if flush_outs > 0 {
			oddsLog("[lookAhead]", fmt.Sprintln("Flush outs", fmt.Sprintf("%.2f", flush_outs)+"%"))
			if flush_outs > 18 {
				for i := range s {
					if v[i] > 12 && s[i] == flush {
						oddsLog("[lookAhead]", fmt.Sprintln("High flush possible", fmt.Sprintf("%.d High", v[i])))
						sure = true
					}
				}

				if (l == 3 && flush_outs > 30) || future < Odds.Bot.Risk[2] {
					sure = true
				}
			}
		}

		if straight_outs > 0 {
			oddsLog("[lookAhead]", fmt.Sprintln("Straight outs", fmt.Sprintf("%.2f", straight_outs)+"%"))
			if straight_outs > 17 {
				for i := range v {
					if v[i] == run[len(run)-1] || (v[i] == 5 && run[0] == 2) {
						oddsLog("[lookAhead]", fmt.Sprintln("High straight possible", fmt.Sprintf("%.d High", v[i])))
						sure = true
					}
				}
			}

			if (l == 3 && straight_outs > 16) || future < Odds.Bot.Risk[2] {
				sure = true
			}
		}

	} else {
		straight_outs = 0
		flush_outs = 0
	}

	oddsLog("[lookAhead]", fmt.Sprintln("Remaining cards", 5-l))

	if odds > Odds.Bot.Risk[2] {
		po, better := potOdds(odds, straight_outs, flush_outs)
		if better && aggressive(sure) {
			ahead = Odds.Bot.Risk[2] - odds - 1
			oddsLog("[lookAhead]", fmt.Sprintln("Good pot odds on draw", fmt.Sprintf("%.2f", ahead)+"%"))
		} else if po != 0 && !better {
			if future > odds {
				ahead = ahead + Odds.Bot.Risk[1]
				oddsLog("[lookAhead]", fmt.Sprintln("Bad pot odds", "+"+fmt.Sprintf("%.2f", Odds.Bot.Risk[1])+"%"))
			}
		}
	} else {
		pair, trip, sub := pairOnBoard(value, "pairOnBoard")
		if pair > 0 && trip == 0 {
			if rank >= 9 {
				ahead = ahead + 6
				oddsLog("[lookAhead]", fmt.Sprintln("Using board pair", "+"+fmt.Sprintf("%.2f", 6.00)+"%"))
			} else if rank == 8 {
				if Odds.high_pair > pair {
					for i := range value {
						if value[i] > Odds.high_pair {
							ahead++
							oddsLog("[lookAhead]", fmt.Sprintln("Higher card than high pair", "+"+fmt.Sprintf("%.2f", 1.00)+"%"))
						}
					}
				} else {
					ahead = ahead + 6
					oddsLog("[lookAhead]", fmt.Sprintln("High pair is board pair", "+"+fmt.Sprintf("%.2f", 6.00)+"%"))
					for i := range value {
						if value[i] > Odds.low_pair {
							ahead++
							oddsLog("[lookAhead]", fmt.Sprintln("Higher card than low pair", "+"+fmt.Sprintf("%.2f", 1.00)+"%"))
						}
					}
				}
			}
		} else if rank == 8 || rank == 9 {
			for i := range value {
				if value[i] > Odds.high_pair {
					ahead++
					oddsLog("[lookAhead]", fmt.Sprintln("Higher card than high pair", "+"+fmt.Sprintf("%.2f", 1.00)+"%"))
				}

				if rank == 8 {
					if value[i] > Odds.low_pair {
						ahead++
						oddsLog("[lookAhead]", fmt.Sprintln("Higher card than low pair", "+"+fmt.Sprintf("%.2f", 1.00)+"%"))
					}
				}
			}
		}

		if trip > 0 {
			if rank >= 7 {
				var o float64
				if Odds.high_card == 14 {
					if l < 5 {
						o = 3
					} else {
						o = 6
					}
				} else {
					if l < 5 {
						o = 9
					} else {
						o = 15
					}
				}

				ahead = ahead + o
				oddsLog("[lookAhead]", fmt.Sprintln("Using board trip", "+"+fmt.Sprintf("%.2f", o)+"%"))
			}
		}

		_, run3, run4, off3, off4, in3 := runOnBoard(sub, "runOnBoard")
		if run3 {
			if rank > 6 {
				if straight_outs < 1.5 {
					var o float64
					if l < 4 {
						o = fRank
					} else {
						o = (fRank / 2)
					}

					ahead = ahead + o
					oddsLog("[lookAhead]", fmt.Sprintln("Low straight outs", "+"+fmt.Sprintf("%.2f", o)+"%"))
				}
			}
		}

		if run4 {
			if rank > 6 {
				ahead = ahead + fRank
				oddsLog("[lookAhead]", fmt.Sprintln("Worse than four outside", "+"+fmt.Sprintf("%.2f", fRank)+"%"))
			}
		}

		if off3 || in3 {
			if rank > 6 {
				if straight_outs < 1 {
					var o float64
					if l < 4 {
						o = fRank
					} else {
						o = (fRank / 2)
					}

					ahead = ahead + o
					oddsLog("[lookAhead]", fmt.Sprintln("Low inside straight outs", "+"+fmt.Sprintf("%.2f", o)+"%"))
				}
			}
		}

		if off4 {
			if rank > 6 {
				o := (fRank / 2)
				ahead = ahead + o
				oddsLog("[lookAhead]", fmt.Sprintln("Worse than four inside", "+"+fmt.Sprintf("%.2f", o)+"%"))
			}
		}

		_, suit3, suit4 := suitedBoard(l, suit, "suitedBoard")
		if suit3 {
			if rank > 5 {
				if flush_outs < 5 {
					var o float64
					if l < 4 {
						o = fRank
					} else {
						o = (fRank / 2)
					}

					ahead = ahead + o
					oddsLog("[lookAhead]", fmt.Sprintln("Low flush outs", "+"+fmt.Sprintf("%.2f", o)+"%"))
				}
			}
		}

		if suit4 {
			if rank > 5 {
				ahead = ahead + fRank
				oddsLog("[lookAhead]", fmt.Sprintln("Worse than four suited", "+"+fmt.Sprintf("%.2f", fRank)+"%"))
			} else if rank == 5 {
				v := [2]int{}
				for i := range Odds.hole {
					v[i] = suitSplit(Odds.hole[i])[0]
				}

				if Odds.used[0] && !Odds.used[1] {
					if v[0] < 12 {
						o := (Odds.Bot.Risk[1] + 5) - float64(v[0])
						ahead = ahead + o
						oddsLog("[lookAhead]", fmt.Sprintln("Low flush", "+"+fmt.Sprintf("%.2f", o)+"%"))
					}
				}

				if Odds.used[1] && !Odds.used[0] {
					if v[1] < 12 {
						o := (Odds.Bot.Risk[1] + 3) - float64(v[1])
						ahead = ahead + o
						oddsLog("[lookAhead]", fmt.Sprintln("Low flush", "+"+fmt.Sprintf("%.2f", o)+"%"))
					}
				}
			}
		}
	}

	oddsLog("[lookAhead]", fmt.Sprintln("Adding", fmt.Sprintf("%.2f", ahead)+"%"))

	return
}

// Find if there is a pair or trip
func pairOnBoard(cards []int, str string) (pair, trip int, sub []int) {
	sub = cards
	for i := range cards {
		l := len(sub)
		if i > l-2 || l < 3 {
			break
		}

		if i <= l-3 {
			if sub[i] == sub[i+1] && sub[i] == sub[i+2] {
				log.Println("[" + str + "] Trip")
				sub = append(sub[0:i], sub[i+2:]...)
				trip = cards[i]
				break
			}
		}

		if sub[i] == sub[i+1] {
			log.Println("[" + str + "] Pair")
			sub = append(sub[0:i], sub[i+1:]...)
			pair = cards[i]
		}
	}

	return
}

// Find if there is a run
func runOnBoard(cards []int, str string) (run []int, run3, run4, off3, off4, in3 bool) {
	l := len(cards)
	for i := range cards {
		if i > l-3 || l < 3 {
			break
		}

		if i <= l-3 && !run4 && !off4 {
			if cards[i] < 11 && cards[i] == cards[i+1]-1 && cards[i] == cards[i+2]-2 {
				run = []int{cards[i], cards[i+1], cards[i+2]}
				log.Println("[" + str + "] Three outside")
				run3 = true
			}

			if (cards[i] == cards[i+1]-1 && cards[i] == cards[i+2]-3) ||
				(cards[i] == cards[i+1]-2 && cards[i] == cards[i+2]-3) ||
				(cards[l-1] == 14 && cards[i] == 2 && cards[i+1] == 3) ||
				(cards[l-1] == 14 && cards[i] == 2 && cards[i+1] == 4) ||
				(cards[l-1] == 14 && cards[i] == 3 && cards[i+1] == 4) {
				run = []int{cards[i], cards[i+1], cards[i+2]}
				log.Println("[" + str + "] Three inside")
				off3 = true
			}

			if cards[i] == cards[i+1]-2 && cards[i] == cards[i+2]-4 {
				run = []int{cards[i], cards[i+1], cards[i+2]}
				log.Println("[" + str + "] Three middle")
				in3 = true
			}
		}

		if l < 4 {
			break
		}

		if i <= l-4 {
			if cards[i] < 11 && cards[i] == cards[i+1]-1 && cards[i] == cards[i+2]-2 && cards[i] == cards[i+3]-3 {
				run = []int{cards[i], cards[i+1], cards[i+2], cards[i+3]}
				log.Println("[" + str + "] Four outside")
				run4 = true
			}

			if (cards[i] == cards[i+1]-1 && cards[i] == cards[i+2]-3 && cards[i] == cards[i+3]-4) ||
				(cards[i] == cards[i+1]-1 && cards[i] == cards[i+2]-2 && cards[i] == cards[i+3]-4) ||
				(cards[i] == cards[i+1]-2 && cards[i] == cards[i+2]-3 && cards[i] == cards[i+3]-4) ||
				(cards[l-1] == 14 && cards[i] == 2 && cards[i+1] == 3 && cards[i+2] == 4) ||
				(cards[l-1] == 14 && cards[i] == 2 && cards[i+1] == 3 && cards[i+2] == 5) ||
				(cards[l-1] == 14 && cards[i] == 2 && cards[i+1] == 4 && cards[i+2] == 5) ||
				(cards[l-1] == 14 && cards[i] == 3 && cards[i+1] == 4 && cards[i+2] == 5) ||
				(cards[i] == 11 && cards[i+1] == 12 && cards[i+2] == 13 && cards[i+3] == 14) {
				run = []int{cards[i], cards[i+1], cards[i+2], cards[i+3]}
				log.Println("[" + str + "] Four inside")
				off4 = true
			}
		}
	}

	return
}

// Find if there is suited
func suitedBoard(l int, cards []int, str string) (suit int, suit3, suit4 bool) {
	suit = 100
	for i := range cards {
		if i > l-3 || l < 3 {
			break
		}

		if i <= l-3 && !suit3 {
			if cards[i] == cards[i+1] && cards[i] == cards[i+2] {
				log.Println("[" + str + "] Three suited")
				suit = cards[i]
				suit3 = true
			}
		}

		if i <= l-4 {
			if cards[i] == cards[i+1] && cards[i] == cards[i+2] && cards[i] == cards[i+3] {
				log.Println("[" + str + "] Four suited")
				suit = cards[i]
				suit4 = true
			}
		}
	}

	return
}

// Gets high pair, high card and low pair
func getHighs(h []int) (highPair int, highCard int, lowPair int) {
	l := len(h)
	for i := range h {
		if i > l-2 {
			break
		}

		if h[i] == h[i+1] {
			if h[i] > highPair {
				highPair = h[i]
			}
		}
	}

	for i := range h {
		if i > l-2 {
			break
		}

		if h[i] == h[i+1] {
			if h[i] > lowPair && h[i] != highPair {
				lowPair = h[i]
			}
		}
	}

	for i := range h {
		if h[i] > highCard && h[i] != highPair && h[i] != lowPair {
			highCard = h[i]
		}
	}

	return
}

// Find possible current hands that could be better to determine base odds %
func countBetter(rank int, comm []int, p [][]int) (better float64, count float64) {
	for i := range p {
		if threeMatch(comm, p[i]) {
			r, h, _ := compareTheseFive(p[i])
			if r < rank {
				better++
			} else if r == rank {
				switch rank {
				case 1:
					// nothing
				case 2:
					for i := range h {
						if Odds.hand[i] < h[i] {
							better++
							break
						}
					}
				case 3:
					_, hc, _ := getHighs(h)
					if hc > Odds.high_card {
						better++
					}
				case 4:
					hp, hc, lp := getHighs(h)
					if hp > Odds.high_pair || (hp == Odds.high_pair && lp > Odds.low_pair) || (hp == Odds.high_pair && lp == Odds.low_pair && hc > Odds.high_card) {
						better++
					}
				case 5:
					for i := range h {
						if Odds.hand[i] < h[i] {
							better++
							break
						}
					}
				case 6:
					for i := range h {
						if Odds.hand[i] < h[i] {
							better++
							break
						}
					}
				case 7:
					hp, hc, _ := getHighs(h)
					if hp > Odds.high_pair || (hp == Odds.high_pair && hc > Odds.high_card) {
						better++
					}
				case 8:
					hp, hc, lp := getHighs(h)
					if hp > Odds.high_pair || (hp == Odds.high_pair && lp > Odds.low_pair) || (hp == Odds.high_pair && lp > Odds.low_pair && hc > Odds.high_card) {
						better++
					}
				case 9:
					hp, hc, _ := getHighs(h)
					if hp > Odds.high_pair || (hp == Odds.high_pair && hc > Odds.high_card) {
						better++
					}
				case 10:
					for i := range h {
						if Odds.hand[i] < h[i] {
							better++
							break
						}
					}
				default:
				}
			}
			count++
		}
	}

	return
}

// Determine base odds % from countBetter()
func betterHands(rank int, comm []int, p [][]int) float64 {
	oddsLog("[betterHands]", fmt.Sprintln("Community", comm))
	better, count := countBetter(rank, comm, p)
	oddsLog("[betterHands]", fmt.Sprintln("Better hands", fmt.Sprintf("%0.2f", better/count*100)+"%"))
	oddsLog("[betterHands]", fmt.Sprintln("Counted", int(count), "hands"))

	return better / count * 100
}

// Find possible future hands that could be better to add to base odds %
func futureHands(p [][]int) float64 {
	var av, count float64
	possible := cardNumbers(append(Odds.community, Odds.hole[:]...))

	for i := range possible {
		new := append(Odds.community, possible[i])
		cards := append(new, Odds.hole[:]...)

		if len(cards) > 4 {
			count++
			rank, _ := myFutureHand(cards)
			b, c := countBetter(rank, new, p)
			av = av + (b / c * 100)
		}
	}
	oddsLog("[futureHands]", fmt.Sprintln("Average draw", fmt.Sprintf("%.2f", av/count)+"%"))

	return av / count
}

// Find possible player future hands for futureHands() compare
func myFutureHand(cards []int) (rank int, hand []int) {
	rank = 100
	p := Pool(5, cards)
	p1 := suitSplit(Odds.hole[0])
	p2 := suitSplit(Odds.hole[1])

	for i := range p {
		r, h, _ := compareTheseFive(p[i])
		if Odds.hand == nil {
			rank = r
			hand = p[i]
		}

		if r < rank {
			rank = r
			hand = p[i]
		} else if r == rank {
			var better bool
			var r ranker
			r.pc1 = p1
			r.pc2 = p2
			new := findBest(rank, Odds.hand, h, &r)
			for i := 0; i < 5; i++ {
				if new[i] != hand[i] {
					better = true
					break
				}
			}

			if better {
				hand = p[i]
			}
		}
	}

	return
}

// Find possible future hole hands that could be better to determine base odds %
func betterHole(cards []int) float64 {
	var better, count float64
	mr, _ := compareTheseTwo(cards)
	numbers := cardNumbers(Odds.hole[:])
	p := Pool(2, numbers)
	oddsLog("[betterHole]", fmt.Sprintln("Hole", cards))
	for i := range p {
		r, _ := compareTheseTwo(p[i])
		if r > mr {
			better++
		}
		count++
	}

	oddsLog("[betterHole]", fmt.Sprintln("Better hands", fmt.Sprintf("%.2f", better/count*100)+"%"))
	oddsLog("[betterHole]", fmt.Sprintln("Counted", int(count), "hands"))

	return better / count * 100
}

// Main odds routine where base odds % and situational odds are combined before passing to BetLog()
func MakeOdds() (odds float64, future float64) {
	fmt.Println()
	Odds.community = []int{}
	Odds.Label.SetText("")
	if Round.Flop1 > 0 {
		Odds.community = append(Odds.community, Round.Flop1)
	}

	if Round.Flop2 > 0 {
		Odds.community = append(Odds.community, Round.Flop2)
	}

	if Round.Flop3 > 0 {
		Odds.community = append(Odds.community, Round.Flop3)
	}

	if Round.TurnCard > 0 {
		Odds.community = append(Odds.community, Round.TurnCard)
	}

	if Round.RiverCard > 0 {
		Odds.community = append(Odds.community, Round.RiverCard)
	}

	Odds.hole = [2]int{Card(Round.Cards.Local1), Card(Round.Cards.Local2)}

	if Odds.hole[:] == nil || Odds.hole[0] == 0 || Odds.hole[1] == 0 {
		oddsLog("[makeOdds]", fmt.Sprintln("No Cards"))
		return 200, 0
	}

	numbers := cardNumbers(Odds.hole[:])
	p := Pool(5, numbers)
	cards := append(Odds.community, Odds.hole[:]...)

	if len(cards) > 4 {
		var cap float64
		rank := myHand(cards)
		o := betterHands(rank, Odds.community, p)

		if !Odds.used[0] && !Odds.used[1] && o != 0 {
			cap = Odds.Bot.Risk[2] - Odds.Bot.Risk[0]
			oddsLog("[makeOdds]", fmt.Sprintln("Using community hand", "+"+fmt.Sprintf("%.2f", cap)+"%"))
		}

		odds = o + cap

		future = futureHands(p)
		la := lookAhead(odds, future, rank, Odds.community)
		if odds > math.Abs(la) || la > odds {
			odds = odds + la
		}

		return

	} else {
		odds = betterHole(Odds.hole[:])

		return odds, 100
	}
}

// Create random values for randomized params in BetLog()
func randomize() (float64, float64, float64) {
	var a, b, c float64
	if Odds.Bot.Random[1] == 1 || Odds.Bot.Random[1] == 3 {
		a = 0 + rand.Float64()*(Odds.Bot.Random[0]-0)
		time.Sleep(9 * time.Millisecond)
	}

	if Odds.Bot.Random[1] == 2 || Odds.Bot.Random[1] == 3 {
		b = 0 + rand.Float64()*(Odds.Bot.Random[0]-0)
		time.Sleep(9 * time.Millisecond)
	}

	c = 0 + rand.Float64()*(Odds.Bot.Random[0]-0)
	i := rand.Intn(20-1) + 1

	if i%2 == 0 {
		a = 0 - a
		b = 0 - b
	}

	return a, b, c
}

// Find players outs for straights and flush
func findMyOuts(l, rank int, value, suit []int) (run []int, flush int, straight_outs, flush_outs float64) {
	sort.Ints(value)
	sort.Ints(suit)
	log.Println("[findMyOuts]", value, suit)
	_, _, sub := pairOnBoard(value, "findMyOuts")
	var run3, run4, off3, off4, in3, suit3, suit4 bool
	run, run3, run4, off3, off4, in3 = runOnBoard(sub, "findMyOuts")
	flush, suit3, suit4 = suitedBoard(l, suit, "findMyOuts")

	if rank > 6 {
		if run4 {
			o := float64(8)
			switch l - 2 {
			case 3:
				straight_outs = o / 47 * 190
			case 4:
				straight_outs = o / 46 * 100
			}
		} else if run3 {
			switch l - 2 {
			case 3:
				o := float64(8)
				t := o / 47
				r := o / 46
				straight_outs = t * r * 100

			default:
			}
		}

		if off4 {
			o := float64(4)
			switch l - 2 {
			case 3:
				straight_outs = o / 47 * 190
			case 4:
				straight_outs = o / 46 * 100
			}
		} else if off3 || in3 {
			switch l - 2 {
			case 3:
				o := float64(8)
				t := o / 47
				r := (o - 4) / 46
				straight_outs = t * r * 100
			default:
			}
		}
	}

	if rank > 5 {
		if suit4 {
			o := float64(9)
			switch l - 2 {
			case 3:
				flush_outs = o / 47 * 190
			case 4:
				flush_outs = o / 46 * 100
			}
		} else if suit3 {
			switch l - 2 {
			case 3:
				o := float64(10)
				t := o / 47
				r := (o - 1) / 46
				flush_outs = t * r * 100
			default:
			}
		}
	}

	return
}

// Check minimum bet at current table
func MinBet() uint64 {
	if Round.Ante == 0 {
		return Round.BB
	}

	return Round.Ante
}

// Check if bet is greater than allowed by Odds.Bot.Max
func maxBet(amt float64) bool {
	return amt > Odds.Bot.Max
}

// If bet is to be called
func callBet(m float64, live bool) bool {
	var amt float64
	if Signal.PlacedBet && Round.Raised > 0 {
		amt = float64(Round.Raised) / 100000
	} else {
		amt = float64(Round.Wager) / 100000
	}

	if maxBet(amt) {
		oddsLog("[callBet]", fmt.Sprintln("Amount higher than max bet", amt))
		return false
	}

	if lowBalance(amt) {
		oddsLog("[callBet]", fmt.Sprintln("Low Balance", amt))
		return false
	}
	curr := "Dero"
	if Round.Asset {
		curr = "Tokens"
	}

	oddsLog("[callBet]", fmt.Sprintln("Call", fmt.Sprintf("%.2f", m)+"x", fmt.Sprintf("%.1f", amt), curr))
	if live {
		Bet(fmt.Sprintf("%.1f", amt))
	}

	return true
}

// If bet is to be raised
func raiseBet(m float64, live bool) bool {
	amt := (float64(Round.Wager) / 100000) * m
	if maxBet(amt) {
		oddsLog("[raiseBet]", fmt.Sprintln("Setting bet amount to max", amt))
		amt = Odds.Bot.Max
	}

	if lowBalance(amt) {
		oddsLog("[raiseBet]", fmt.Sprintln("Low Balance", amt))
		return false
	}

	oddsLog("[raiseBet]", fmt.Sprintln("Raise bet, multiplier and amount", m, fmt.Sprintf("%.1f", amt)))
	if live {
		Bet(fmt.Sprintf("%.1f", amt))
	}

	return true
}

// Check if bet can be raised
func canRaise() bool {
	if !Signal.PlacedBet && Round.Raised == 0 && Round.Wager > 0 {
		oddsLog("[canRaise]", "true")
		return true
	}

	return false
}

// Check if wallet balance is to low to call bet
func lowBalance(amt float64) bool {
	if Round.Asset {
		return amt > float64(rpc.Wallet.ReadTokenBalance(rpc.GetAssetSCIDName(Round.AssetID)))/100000
	} else {
		return amt > float64(rpc.Wallet.ReadBalance())/100000
	}
}

// Find if player is in last position
func lastPosition(id int) bool {
	last := 0
	ins := []bool{Signal.In1, Signal.In2, Signal.In3, Signal.In4, Signal.In5, Signal.In6}
	folds := []bool{Round.F1, Round.F2, Round.F3, Round.F4, Round.F5, Round.F6}
	order := []int{}
	dealer := rpc.StringToInt(Display.Dealer)

	for i := range ins {
		if (ins[i] && !folds[i]) || i == dealer-1 {
			order = append(order, i+1)
		}
	}

	if dealer == order[0] {
		last = order[len(order)-1]
	} else {
		for i := range order {
			if order[i] == dealer {
				last = order[i-1]
			}
		}
	}

	if id == last {
		oddsLog("[lastPosition]", fmt.Sprintln("Last move"))
		return true
	}

	return false
}

// Random switch working with Odds.Bot.Slow to keep bets from being placed when hand is good
func slowPlay(future float64) (bool, bool) {
	skip := true
	i := rand.Intn(36-1) + 1
	switch Odds.Bot.Slow {
	case 5:
		if i%5 == 0 {
			skip = false
		}
	case 4:
		if i%4 == 0 {
			skip = false
		}
	case 3:
		if i%3 == 0 || i == 1 || i == 17 || i == 32 {
			skip = false
		}
	case 2:
		if i%2 == 0 || i == 1 || i == 33 {
			skip = false
		}
	case 1:
		skip = false
	}

	if future < Odds.Bot.Risk[2] && lastPosition(Round.ID) {
		if Odds.Bot.Aggr > 3 {
			return false, false
		}
		return false, true

	}

	oddsLog("[slowPlay]", fmt.Sprintln("Slow", skip))

	return skip, false
}

// Random switch working with Odds.Bot.Aggr will trigger bets and raises more often
func aggressive(sure bool) (bet bool) {
	if !sure {
		i := rand.Intn(36-1) + 1
		switch Odds.Bot.Aggr {
		case 5:
			bet = true
		case 4:
			if i%2 == 0 {
				bet = true
			}
		case 3:
			if i%3 == 0 {
				bet = true
			}
		case 2:
			if i%4 == 0 {
				bet = true
			}
		case 1:
			if i%5 == 0 {
				bet = true
			}
		}
	} else {
		bet = true
	}

	oddsLog("[aggressive]", fmt.Sprintln("Aggressive", bet))

	return
}

// Random switch working with Odds.Bot.Aggr will trigger bluffing situations
func bluff() (bet bool) {
	if lastPosition(Round.ID) {
		i := rand.Intn(9-1) + 1
		switch Odds.Bot.Aggr {
		case 5:
			bet = true
		case 4:
			if i%2 == 0 {
				bet = true
			} else {
				bet = false
			}
		case 3:
			if i%3 == 0 {
				bet = true
			} else {
				bet = false
			}
		case 2:
			if i%4 == 0 {
				bet = true
			} else {
				bet = false
			}
		case 1:
			if i%5 == 0 {
				bet = true
			} else {
				bet = false
			}
		}

		oddsLog("[bluff]", fmt.Sprintln("Bluffing", bet))
	}

	return
}

// Find pot odds of situation
func potOdds(odds, straight_outs, flush_outs float64) (po float64, better bool) {
	if Signal.PlacedBet {
		po = float64(Round.Raised) / float64(Round.Pot) * 100
	} else {
		po = float64(Round.Wager) / float64(Round.Pot) * 100
	}

	if po != 0 {
		oddsLog("[potOdds]", fmt.Sprintln("Pot odds", fmt.Sprintf("%.2f", po)+"%"))
		if straight_outs > po || flush_outs > po {
			better = true
		}
	}

	return
}

// Main bet logic where odds are combined with any random values,
// then situation is evaluated against users params to determine what action to take
func BetLogic(odds, future float64, live bool) {
	if odds == 200 {
		return
	}

	var a, b, c float64
	if Odds.Bot.Random[0] > 0 {
		a, b, c = randomize()
		oddsLog("[BetLogic]", fmt.Sprintln("Randomize", fmt.Sprintf("%.2f", a)+"%", fmt.Sprintf("%.2f", b)+"%", fmt.Sprintf("%.2f", c)+"%"))
	}

	if Odds.Bot.Risk[0]+a < 1 {
		a = 0
	}

	if Odds.Bot.Bet[0]+b < 1 {
		b = 0
	}

	curr := "Dero"
	if Round.Asset {
		curr = "Tokens"
	}

	oddsLog("[BetLogic]", fmt.Sprintln("Wager is", fmt.Sprintf("%.2f", float64(Round.Wager)/100000), curr))
	oddsLog("[BetLogic]", fmt.Sprintln("Odds Calc", fmt.Sprintf("%.2f", odds)+"%"))
	oddsLog("[BetLogic]", fmt.Sprintln("Luck", fmt.Sprintf("%.2f", Odds.Bot.Luck)+"%", fmt.Sprintf("%.2f", c)+"%"))

	l := len(Odds.community)
	var sure bool
	if odds == 0 && l == 5 {
		sure = true
	}

	if (odds-Odds.Bot.Luck)-c < 1 {
		odds = 0.9
	} else {
		odds = (odds - Odds.Bot.Luck) - c
	}

	oddsLog("[BetLogic]", fmt.Sprintln("Final Odds", fmt.Sprintf("%.2f", odds)+"%"))
	oddsLog("[BetLogic]", fmt.Sprintln("Risk1", fmt.Sprintf("%.2f", Odds.Bot.Risk[0]+a)+"%"))
	oddsLog("[BetLogic]", fmt.Sprintln("Risk2", fmt.Sprintf("%.2f", Odds.Bot.Risk[1]+a)+"%"))
	oddsLog("[BetLogic]", fmt.Sprintln("Risk3", fmt.Sprintf("%.2f", Odds.Bot.Risk[2]+a)+"%"))
	oddsLog("[BetLogic]", fmt.Sprintln("Bet1", fmt.Sprintf("%.2f", Odds.Bot.Bet[0]+b)+"x"))
	oddsLog("[BetLogic]", fmt.Sprintln("Bet2", fmt.Sprintf("%.2f", Odds.Bot.Bet[1]+b)+"x"))
	oddsLog("[BetLogic]", fmt.Sprintln("Bet3", fmt.Sprintf("%.2f", Odds.Bot.Bet[2]+b)+"x"))

	if Round.Wager == 0 {
		var amt float64
		var bet bool
		slow, min := slowPlay(future)
		if !slow || sure {
			if odds < (Odds.Bot.Risk[0]+a) && Round.Flop && aggressive(false) && !min {
				amt = (float64(MinBet()) / 100000) * (Odds.Bot.Bet[2] + b)
				oddsLog("[BetLogic]", fmt.Sprintln("Bet High", fmt.Sprintf("%.1f", amt), curr))
				bet = true
			} else if odds < (Odds.Bot.Risk[1]+a) && !min {
				amt = (float64(MinBet()) / 100000) * (Odds.Bot.Bet[1] + b)
				oddsLog("[BetLogic]", fmt.Sprintln("Bet Med", fmt.Sprintf("%.1f", amt), curr))
				bet = true
			} else if odds < (Odds.Bot.Risk[2]+a) || bluff() {
				amt = (float64(MinBet()) / 100000) * (Odds.Bot.Bet[0] + b)
				oddsLog("[BetLogic]", fmt.Sprintln("Bet Low", fmt.Sprintf("%.1f", amt), curr))
				bet = true
			}
		}

		if !bet {
			oddsLog("[BetLogic]", fmt.Sprintln("Check"))
			if live {
				Check()
			}
		} else if live {
			if lowBalance(amt) {
				oddsLog("[BetLogic]", fmt.Sprintln("Low Balance", amt))
			} else if maxBet(amt) {
				oddsLog("[BetLogic]", fmt.Sprintln("Setting bet amount to max", amt))
				Bet(fmt.Sprintf("%.1f", Odds.Bot.Max))
			} else if amt < float64(MinBet())/100000 {
				oddsLog("[BetLogic]", fmt.Sprintln("Setting bet amount to min", amt))
				Bet(fmt.Sprintf("%.1f", float64(MinBet())/100000))
			} else {
				Bet(fmt.Sprintf("%.1f", amt))
			}
		}

	} else if Round.Wager > 0 {
		var bet bool
		if odds < 1 {
			if Round.Wager <= uint64(Odds.Bot.Max)*100000 {
				if canRaise() && aggressive(false) {
					bet = raiseBet((Odds.Bot.Aggr*2)*1.5, live)
				} else {
					bet = callBet(Odds.Bot.Max, live)
				}
			}
		} else if odds < Odds.Bot.Risk[0]+a {
			if Round.Wager <= MinBet()*uint64(Odds.Bot.Bet[2]+b)*uint64(Odds.Bot.Aggr) {
				if canRaise() && aggressive(false) {
					bet = raiseBet(Odds.Bot.Aggr*2, live)
				} else {
					bet = callBet((Odds.Bot.Bet[2]+b)*Odds.Bot.Aggr, live)
				}
			}
		} else if odds < Odds.Bot.Risk[1]+a {
			if Round.Wager <= MinBet()*uint64(Odds.Bot.Bet[1]+b)*uint64(Odds.Bot.Aggr) {
				bet = callBet((Odds.Bot.Bet[1]+b)*Odds.Bot.Aggr, live)
			}
		} else if odds < Odds.Bot.Risk[2]+a {
			if Round.Wager <= MinBet()*uint64(Odds.Bot.Bet[0]+b)*uint64(Odds.Bot.Aggr) {
				bet = callBet((Odds.Bot.Bet[0]+b)*Odds.Bot.Aggr, live)
			}
		} else if !Round.Flop && odds < 50+Odds.Bot.Aggr*12 {
			if Round.Wager <= MinBet()*uint64(Odds.Bot.Bet[0]+b)*uint64(Odds.Bot.Aggr) {
				bet = callBet((Odds.Bot.Bet[0]+b)*Odds.Bot.Aggr, live)
				oddsLog("[BetLogic]", fmt.Sprintln("No Pushover", fmt.Sprintf("%.2f", 50+Odds.Bot.Aggr*12)+"%"))
			}
		} else if l == 3 {
			if Odds.Bot.Aggr > 3 && future < odds {
				if Round.Wager <= MinBet()*uint64(Odds.Bot.Bet[0]+b)*2 {
					bet = callBet((Odds.Bot.Bet[0]+b)*2, live)
				}
			} else if Odds.Bot.Aggr == 3 && odds < 50 && future < odds {
				if Round.Wager <= MinBet()*uint64(Odds.Bot.Bet[0]+b)*2 {
					bet = callBet((Odds.Bot.Bet[0]+b)*2, live)
				}
			} else if Odds.Bot.Aggr < 3 && odds < 40 && future < odds {
				if Round.Wager <= MinBet()*uint64(Odds.Bot.Bet[0]+b) {
					bet = callBet((Odds.Bot.Bet[0] + b), live)
				}
			}

			if bet {
				oddsLog("[BetLogic]", fmt.Sprintln("See one more"))
			}
		}

		if !bet {
			oddsLog("[BetLogic]", fmt.Sprintln("Fold"))
			if live {
				Check()
			}
		}
	}
}

// Prints odds info and adds to gui log
func oddsLog(f, str string) {
	log.Print(f, " ", str)
	Odds.Label.SetText(Odds.Label.Text + str)
}

// Check if current Holdero table is active
func GameIsActive() bool {
	return Round.Players > 1
}

// Set config to stored values
func SetBotConfig(opt Bot_config) {
	Odds.Bot.Risk[2] = opt.Risk[2]
	Odds.Bot.Risk[1] = opt.Risk[1]
	Odds.Bot.Risk[0] = opt.Risk[0]
	Odds.Bot.Bet[2] = opt.Bet[2]
	Odds.Bot.Bet[1] = opt.Bet[1]
	Odds.Bot.Bet[0] = opt.Bet[0]
	Odds.Bot.Luck = opt.Luck
	Odds.Bot.Slow = opt.Slow
	Odds.Bot.Aggr = opt.Aggr
	Odds.Bot.Max = opt.Max
	Odds.Bot.Random[0] = opt.Random[0]
	Odds.Bot.Random[1] = opt.Random[1]
}

// Save config of current values
func SaveBotConfig(i int, opt Bot_config) {
	Stats.Bots[i].Risk[2] = opt.Risk[2]
	Stats.Bots[i].Risk[1] = opt.Risk[1]
	Stats.Bots[i].Risk[0] = opt.Risk[0]
	Stats.Bots[i].Bet[2] = opt.Bet[2]
	Stats.Bots[i].Bet[1] = opt.Bet[1]
	Stats.Bots[i].Bet[0] = opt.Bet[0]
	Stats.Bots[i].Luck = opt.Luck
	Stats.Bots[i].Slow = opt.Slow
	Stats.Bots[i].Aggr = opt.Aggr
	Stats.Bots[i].Max = opt.Max
	Stats.Bots[i].Random[0] = opt.Random[0]
	Stats.Bots[i].Random[1] = opt.Random[1]
}

// Write Holdero stats to file
func WriteHolderoStats(config Player_stats) bool {
	file, err := os.Create("config/stats.json")
	if err != nil {
		log.Println("[WriteHolderoStats]", err)
		return false
	}

	defer file.Close()
	json, _ := json.MarshalIndent(config, "", "")

	_, err = file.Write(json)
	if err != nil {
		log.Println("[WriteHolderoStats]", err)
		return false
	}

	return true
}

// Update win or loss of Holdero stats
func updateStatsWins(amt uint64, player string, fold bool) {
	if Odds.Enabled && !Signal.Odds {
		if "Player"+Display.PlayerId == player {
			Stats.Player.Win++
			Stats.Player.Earnings = Stats.Player.Earnings + float64(amt)/100000
			if Odds.Bot.Name != "" && Odds.Run {
				for i := range Stats.Bots {
					if Odds.Bot.Name == Stats.Bots[i].Name {
						Stats.Bots[i].Stats.Win++
						Stats.Bots[i].Stats.Earnings = Stats.Bots[i].Stats.Earnings + float64(amt)/100000
						SaveBotConfig(i, Odds.Bot)
					}
				}
			}
		} else {
			if !fold {
				Stats.Player.Lost++
			} else {
				Stats.Player.Fold++
			}
			for i := range Stats.Bots {
				if Odds.Bot.Name == Stats.Bots[i].Name {
					if !fold {
						Stats.Bots[i].Stats.Lost++
					} else {
						Stats.Bots[i].Stats.Fold++
					}
					SaveBotConfig(i, Odds.Bot)
				}
			}
		}

		WriteHolderoStats(Stats)
		Signal.Odds = true
	}
}

// Update wager of Holdero stats
func updateStatsWager(amt float64) {
	if Odds.Enabled {
		Stats.Player.Wagered = Stats.Player.Wagered + amt
		if Odds.Bot.Name != "" && Odds.Run {
			for i := range Stats.Bots {
				if Odds.Bot.Name == Stats.Bots[i].Name {
					Stats.Bots[i].Stats.Wagered = Stats.Bots[i].Stats.Wagered + amt
					SaveBotConfig(i, Odds.Bot)

				}
			}
		}

		WriteHolderoStats(Stats)
	}
}

// Update Holdero stats when push
func updateStatsPush(r ranker, amt uint64, f1, f2, f3, f4, f5, f6 bool) {
	if Odds.Enabled && !Signal.Odds {
		fold := false
		ways := float64(0)
		winners := [6]string{"Zero", "Zero", "Zero", "Zero", "Zero", "Zero"}

		if r.p1HighCardArr[0] > 0 && !f1 {
			ways++
			winners[0] = "Player1"
		} else {
			if Round.ID == 1 {
				fold = true
			}
		}

		if r.p2HighCardArr[0] > 0 && !f2 {
			ways++
			winners[1] = "Player2"
		} else {
			if Round.ID == 2 {
				fold = true
			}
		}

		if r.p3HighCardArr[0] > 0 && !f3 {
			ways++
			winners[2] = "Player3"
		} else {
			if Round.ID == 3 {
				fold = true
			}
		}

		if r.p4HighCardArr[0] > 0 && !f4 {
			ways++
			winners[3] = "Player4"
		} else {
			if Round.ID == 4 {
				fold = true
			}
		}

		if r.p5HighCardArr[0] > 0 && !f5 {
			ways++
			winners[4] = "Player5"
		} else {
			if Round.ID == 5 {
				fold = true
			}
		}

		if r.p6HighCardArr[0] > 0 && !f6 {
			ways++
			winners[5] = "Player6"
		} else {
			if Round.ID == 6 {
				fold = true
			}
		}

		var in bool
		for i := range winners {
			if "Player"+Display.PlayerId == winners[i] {
				in = true
			}
		}

		if in {
			Stats.Player.Push++
			Stats.Player.Earnings = Stats.Player.Earnings + float64(amt)/100000/ways
			if Odds.Bot.Name != "" && Odds.Run {
				for i := range Stats.Bots {
					if Odds.Bot.Name == Stats.Bots[i].Name {
						Stats.Bots[i].Stats.Push++
						Stats.Bots[i].Stats.Earnings = Stats.Bots[i].Stats.Earnings + float64(amt)/100000/ways
						SaveBotConfig(i, Odds.Bot)
					}
				}
			}

			WriteHolderoStats(Stats)
		} else {
			if !fold {
				Stats.Player.Lost++
			} else {
				Stats.Player.Fold++
			}

			for i := range Stats.Bots {
				if Odds.Bot.Name == Stats.Bots[i].Name {
					if !fold {
						Stats.Bots[i].Stats.Lost++
					} else {
						Stats.Bots[i].Stats.Fold++
					}
					SaveBotConfig(i, Odds.Bot)
				}
			}

			WriteHolderoStats(Stats)

		}
		Signal.Odds = true
	}
}
