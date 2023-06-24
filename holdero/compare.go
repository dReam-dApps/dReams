package holdero

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"time"

	"github.com/dReam-dApps/dReams/rpc"
)

type ranker struct {
	p1HandRaw [2]int
	p2HandRaw [2]int
	p3HandRaw [2]int
	p4HandRaw [2]int
	p5HandRaw [2]int
	p6HandRaw [2]int

	pc1 [2]int
	pc2 [2]int

	cc1        [2]int
	cc2        [2]int
	cc3        [2]int
	cc4        [2]int
	cc5        [2]int
	p1HighPair int
	p2HighPair int
	p3HighPair int
	p4HighPair int
	p5HighPair int
	p6HighPair int
	p1LowPair  int
	p2LowPair  int
	p3LowPair  int
	p4LowPair  int
	p5LowPair  int
	p6LowPair  int
	p1Rank     int
	p2Rank     int
	p3Rank     int
	p4Rank     int
	p5Rank     int
	p6Rank     int

	fHighCardArr []int

	p1HighCardArr [5]int
	p2HighCardArr [5]int
	p3HighCardArr [5]int
	p4HighCardArr [5]int
	p5HighCardArr [5]int
	p6HighCardArr [5]int
}

// Gets other player cards and decrypt with their keys after reveal
func KeyCard(hash string, who int) int {
	var keyCheck string
	switch who {
	case 1:
		keyCheck = Round.Cards.Key1
	case 2:
		keyCheck = Round.Cards.Key2
	case 3:
		keyCheck = Round.Cards.Key3
	case 4:
		keyCheck = Round.Cards.Key4
	case 5:
		keyCheck = Round.Cards.Key5
	case 6:
		keyCheck = Round.Cards.Key6
	}

	for i := 1; i < 53; i++ {
		finder := strconv.Itoa(i)
		add := keyCheck + finder + Round.SC_seed
		card := sha256.Sum256([]byte(add))
		str := hex.EncodeToString(card[:])

		if str == hash {
			return i
		}
	}
	return 0

}

// Check if player has revealed cards
func revealed(c [2]int) bool {
	if c[0] != 0 && c[1] != 0 {
		return true
	}

	return false
}

// Main routine triggered when players reveal cards at showdown to payout and end round
func getHands(totalHands int) {
	var r ranker
	r.p1Rank = 100
	r.p1HighPair = 0
	r.p2Rank = 100
	r.p2HighPair = 0
	r.p3Rank = 100
	r.p3HighPair = 0
	r.p4Rank = 100
	r.p4HighPair = 0
	r.p5Rank = 100
	r.p5HighPair = 0
	r.p6Rank = 100
	r.p6HighPair = 0

	r.p1HandRaw = [2]int{KeyCard(Round.Cards.P1C1, 1), KeyCard(Round.Cards.P1C2, 1)}
	r.p2HandRaw = [2]int{KeyCard(Round.Cards.P2C1, 2), KeyCard(Round.Cards.P2C2, 2)}
	r.p3HandRaw = [2]int{KeyCard(Round.Cards.P3C1, 3), KeyCard(Round.Cards.P3C2, 3)}
	r.p4HandRaw = [2]int{KeyCard(Round.Cards.P4C1, 4), KeyCard(Round.Cards.P4C2, 4)}
	r.p5HandRaw = [2]int{KeyCard(Round.Cards.P5C1, 5), KeyCard(Round.Cards.P5C2, 5)}
	r.p6HandRaw = [2]int{KeyCard(Round.Cards.P6C1, 6), KeyCard(Round.Cards.P6C2, 6)}

	switch totalHands {
	case 2:
		r.cc1, r.cc2, r.cc3, r.cc4, r.cc5 = getFlop()
		if !Round.F1 && revealed(r.p1HandRaw) {
			r = getHand1(r)
		}
		if !Round.F2 && revealed(r.p2HandRaw) {
			r = getHand2(r)
		}
	case 3:
		r.cc1, r.cc2, r.cc3, r.cc4, r.cc5 = getFlop()
		if !Round.F1 && revealed(r.p1HandRaw) {
			r = getHand1(r)
		}
		if !Round.F2 && revealed(r.p2HandRaw) {
			r = getHand2(r)
		}
		if !Round.F3 && revealed(r.p3HandRaw) {
			r = getHand3(r)
		}
	case 4:
		r.cc1, r.cc2, r.cc3, r.cc4, r.cc5 = getFlop()
		if !Round.F1 && revealed(r.p1HandRaw) {
			r = getHand1(r)
		}
		if !Round.F2 && revealed(r.p2HandRaw) {
			r = getHand2(r)
		}
		if !Round.F3 && revealed(r.p3HandRaw) {
			r = getHand3(r)
		}
		if !Round.F4 && revealed(r.p4HandRaw) {
			r = getHand4(r)
		}
	case 5:
		r.cc1, r.cc2, r.cc3, r.cc4, r.cc5 = getFlop()
		if !Round.F1 && revealed(r.p1HandRaw) {
			r = getHand1(r)
		}
		if !Round.F2 && revealed(r.p2HandRaw) {
			r = getHand2(r)
		}
		if !Round.F3 && revealed(r.p3HandRaw) {
			r = getHand3(r)
		}
		if !Round.F4 && revealed(r.p4HandRaw) {
			r = getHand4(r)
		}
		if !Round.F5 && revealed(r.p5HandRaw) {
			r = getHand5(r)
		}
	case 6:
		r.cc1, r.cc2, r.cc3, r.cc4, r.cc5 = getFlop()
		if !Round.F1 && revealed(r.p1HandRaw) {
			r = getHand1(r)
		}
		if !Round.F2 && revealed(r.p2HandRaw) {
			r = getHand2(r)
		}
		if !Round.F3 && revealed(r.p3HandRaw) {
			r = getHand3(r)
		}
		if !Round.F4 && revealed(r.p4HandRaw) {
			r = getHand4(r)
		}
		if !Round.F5 && revealed(r.p5HandRaw) {
			r = getHand5(r)
		}
		if !Round.F6 && revealed(r.p6HandRaw) {
			r = getHand6(r)
		}
	}

	Display.Res = compareAll(&r)
}

// Gets community cards for ranker
func getFlop() ([2]int, [2]int, [2]int, [2]int, [2]int) {
	c1 := suitSplit(Round.Flop1)
	c2 := suitSplit(Round.Flop2)
	c3 := suitSplit(Round.Flop3)
	c4 := suitSplit(Round.TurnCard)
	c5 := suitSplit(Round.RiverCard)

	return c1, c2, c3, c4, c5
}

// Set player ranker values

func getHand1(r ranker) ranker {
	r.pc1 = suitSplit(r.p1HandRaw[0])
	r.pc2 = suitSplit(r.p1HandRaw[1])

	var rank int
	rank, r = compareThese(&r)
	copy(r.p1HighCardArr[:], r.fHighCardArr)
	r.p1HighPair, r.p1LowPair = getHighPair(r.p1HighCardArr)
	r.p1Rank = rank

	return r
}

func getHand2(r ranker) ranker {
	r.pc1 = suitSplit(r.p2HandRaw[0])
	r.pc2 = suitSplit(r.p2HandRaw[1])

	var rank int
	rank, r = compareThese(&r)
	copy(r.p2HighCardArr[:], r.fHighCardArr)
	r.p2HighPair, r.p2LowPair = getHighPair(r.p2HighCardArr)
	r.p2Rank = rank

	return r
}

func getHand3(r ranker) ranker {
	r.pc1 = suitSplit(r.p3HandRaw[0])
	r.pc2 = suitSplit(r.p3HandRaw[1])

	var rank int
	rank, r = compareThese(&r)
	copy(r.p3HighCardArr[:], r.fHighCardArr)
	r.p3HighPair, r.p3LowPair = getHighPair(r.p3HighCardArr)
	r.p3Rank = rank

	return r
}

func getHand4(r ranker) ranker {
	r.pc1 = suitSplit(r.p4HandRaw[0])
	r.pc2 = suitSplit(r.p4HandRaw[1])

	var rank int
	rank, r = compareThese(&r)
	copy(r.p4HighCardArr[:], r.fHighCardArr)
	r.p4HighPair, r.p4LowPair = getHighPair(r.p4HighCardArr)
	r.p4Rank = rank

	return r
}

func getHand5(r ranker) ranker {
	r.pc1 = suitSplit(r.p5HandRaw[0])
	r.pc2 = suitSplit(r.p5HandRaw[1])

	var rank int
	rank, r = compareThese(&r)
	copy(r.p5HighCardArr[:], r.fHighCardArr)
	r.p5HighPair, r.p5LowPair = getHighPair(r.p5HighCardArr)
	r.p5Rank = rank

	return r
}

func getHand6(r ranker) ranker {
	r.pc1 = suitSplit(r.p6HandRaw[0])
	r.pc2 = suitSplit(r.p6HandRaw[1])

	var rank int
	rank, r = compareThese(&r)
	copy(r.p6HighCardArr[:], r.fHighCardArr)
	r.p6HighPair, r.p6LowPair = getHighPair(r.p6HighCardArr)
	r.p6Rank = rank

	return r
}

// Search through all hand combinations to find best
func compareThese(r *ranker) (int, ranker) {
	e0Hand := []int{r.cc1[0], r.cc2[0], r.cc3[0], r.cc4[0], r.cc5[0]}
	e0Suits := []int{r.cc1[1], r.cc2[1], r.cc3[1], r.cc4[1], r.cc5[1]}
	fRank := makeHand(e0Hand, e0Suits)
	r.fHighCardArr = e0Hand

	/// Two Hole cards
	e1Hand := []int{r.cc1[0], r.cc2[0], r.cc3[0], r.pc1[0], r.pc2[0]}
	e1Suits := []int{r.cc1[1], r.cc2[1], r.cc3[1], r.pc1[1], r.pc2[1]}
	nRank := makeHand(e1Hand, e1Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e1Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e1Hand, r)

	}

	e2Hand := []int{r.cc1[0], r.cc2[0], r.pc1[0], r.cc4[0], r.pc2[0]}
	e2Suits := []int{r.cc1[1], r.cc2[1], r.pc1[1], r.cc4[1], r.pc2[1]}
	nRank = makeHand(e2Hand, e2Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e2Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e2Hand, r)

	}

	e3Hand := []int{r.cc1[0], r.pc1[0], r.cc3[0], r.cc4[0], r.pc2[0]}
	e3Suits := []int{r.cc1[1], r.pc1[1], r.cc3[1], r.cc4[1], r.pc2[1]}
	nRank = makeHand(e3Hand, e3Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e3Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e3Hand, r)

	}

	e4Hand := []int{r.pc1[0], r.cc2[0], r.cc3[0], r.cc4[0], r.pc2[0]}
	e4Suits := []int{r.pc1[1], r.cc2[1], r.cc3[1], r.cc4[1], r.pc2[1]}
	nRank = makeHand(e4Hand, e4Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e4Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e4Hand, r)

	}

	e5Hand := []int{r.cc1[0], r.cc2[0], r.pc1[0], r.pc2[0], r.cc5[0]}
	e5Suits := []int{r.cc1[1], r.cc2[1], r.pc1[1], r.pc2[1], r.cc5[1]}
	nRank = makeHand(e5Hand, e5Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e5Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e5Hand, r)

	}

	e6Hand := []int{r.cc1[0], r.pc1[0], r.cc3[0], r.pc2[0], r.cc5[0]}
	e6Suits := []int{r.cc1[1], r.pc1[1], r.cc3[1], r.pc2[1], r.cc5[1]}
	nRank = makeHand(e6Hand, e6Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e6Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e6Hand, r)

	}

	e7Hand := []int{r.pc1[0], r.cc2[0], r.cc3[0], r.pc2[0], r.cc5[0]}
	e7Suits := []int{r.pc1[1], r.cc2[1], r.cc3[1], r.pc2[1], r.cc5[1]}
	nRank = makeHand(e7Hand, e7Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e7Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e7Hand, r)

	}

	e8Hand := []int{r.cc1[0], r.pc1[0], r.pc2[0], r.cc4[0], r.cc5[0]}
	e8Suits := []int{r.cc1[1], r.pc1[1], r.pc2[1], r.cc4[1], r.cc5[1]}
	nRank = makeHand(e8Hand, e8Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e8Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e8Hand, r)

	}

	e9Hand := []int{r.pc1[0], r.cc2[0], r.pc2[0], r.cc4[0], r.cc5[0]}
	e9Suits := []int{r.pc1[1], r.cc2[1], r.pc2[1], r.cc4[1], r.cc5[1]}
	nRank = makeHand(e9Hand, e9Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e9Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e9Hand, r)

	}

	e10Hand := []int{r.pc1[0], r.pc2[0], r.cc3[0], r.cc4[0], r.cc5[0]}
	e10Suits := []int{r.pc1[1], r.pc2[1], r.cc3[1], r.cc4[1], r.cc5[1]}
	nRank = makeHand(e10Hand, e10Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e10Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e10Hand, r)

	}

	/// First Hole card
	e11Hand := []int{r.cc1[0], r.cc2[0], r.cc3[0], r.cc4[0], r.pc1[0]}
	e11Suits := []int{r.cc1[1], r.cc2[1], r.cc3[1], r.cc4[1], r.pc1[1]}
	nRank = makeHand(e11Hand, e11Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e11Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e11Hand, r)

	}

	e12Hand := []int{r.cc1[0], r.cc2[0], r.cc3[0], r.pc1[0], r.cc5[0]}
	e12Suits := []int{r.cc1[1], r.cc2[1], r.cc3[1], r.pc1[1], r.cc5[1]}
	nRank = makeHand(e12Hand, e12Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e12Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e12Hand, r)

	}

	e13Hand := []int{r.cc1[0], r.cc2[0], r.pc1[0], r.cc4[0], r.cc5[0]}
	e13Suits := []int{r.cc1[1], r.cc2[1], r.pc1[1], r.cc4[1], r.cc5[1]}
	nRank = makeHand(e13Hand, e13Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e13Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e13Hand, r)

	}

	e14Hand := []int{r.cc1[0], r.pc1[0], r.cc3[0], r.cc4[0], r.cc5[0]}
	e14Suits := []int{r.cc1[1], r.pc1[1], r.cc3[1], r.cc4[1], r.cc5[1]}
	nRank = makeHand(e14Hand, e14Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e14Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e14Hand, r)

	}

	e15Hand := []int{r.pc1[0], r.cc2[0], r.cc3[0], r.cc4[0], r.cc5[0]}
	e15Suits := []int{r.pc1[1], r.cc2[1], r.cc3[1], r.cc4[1], r.cc5[1]}
	nRank = makeHand(e15Hand, e15Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e15Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e15Hand, r)

	}

	/// Second Hole card
	e16Hand := []int{r.cc1[0], r.cc2[0], r.cc3[0], r.cc4[0], r.pc2[0]}
	e16Suits := []int{r.cc1[1], r.cc2[1], r.cc3[1], r.cc4[1], r.pc2[1]}
	nRank = makeHand(e16Hand, e16Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e16Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e16Hand, r)

	}

	e17Hand := []int{r.cc1[0], r.cc2[0], r.cc3[0], r.pc2[0], r.cc5[0]}
	e17Suits := []int{r.cc1[1], r.cc2[1], r.cc3[1], r.pc2[1], r.cc5[1]}
	nRank = makeHand(e17Hand, e17Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e17Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e17Hand, r)

	}

	e18Hand := []int{r.cc1[0], r.cc2[0], r.pc2[0], r.cc4[0], r.cc5[0]}
	e18Suits := []int{r.cc1[1], r.cc2[1], r.pc2[1], r.cc4[1], r.cc5[1]}
	nRank = makeHand(e18Hand, e18Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e18Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e18Hand, r)

	}

	e19Hand := []int{r.cc1[0], r.pc2[0], r.cc3[0], r.cc4[0], r.cc5[0]}
	e19Suits := []int{r.cc1[1], r.pc2[1], r.cc3[1], r.cc4[1], r.cc5[1]}
	nRank = makeHand(e19Hand, e19Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e19Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e19Hand, r)

	}

	e20Hand := []int{r.pc2[0], r.cc2[0], r.cc3[0], r.cc4[0], r.cc5[0]}
	e20Suits := []int{r.pc2[1], r.cc2[1], r.cc3[1], r.cc4[1], r.cc5[1]}
	nRank = makeHand(e20Hand, e20Suits)
	if nRank < fRank {
		fRank = nRank
		r.fHighCardArr = e20Hand
	} else if nRank == fRank {
		fRank = nRank
		r.fHighCardArr = findBest(fRank, r.fHighCardArr, e20Hand, r)

	}
	return fRank, *r
}

// If hand ranks are the same look to see which is better when comparing
func findBest(rank int, fR, h []int, r *ranker) []int {
	var hand = h
	var swap = fR
	hole := []int{r.pc1[0], r.pc2[0]}
	sort.Ints(hole)
	sort.Ints(swap)
	sort.Ints(hand)

	/// If straight or straight flush
	if rank == 6 || rank == 2 {
		if hole[0] == swap[4]+1 && hole[1] == swap[4]+2 {
			swap = []int{swap[2], swap[3], swap[4], hole[0], hole[1]}
		} else if hole[0] == swap[4]+1 {
			swap = []int{swap[1], swap[2], swap[3], swap[4], hole[0]}
		} else if hole[1] == swap[4]+1 {
			swap = []int{swap[1], swap[2], swap[3], swap[4], hole[1]}
		} else if swap[0] == 2 && swap[1] == 3 && swap[2] == 4 && swap[3] == 5 && swap[4] == 14 && hole[0] == 6 && hole[1] == 7 {
			swap = []int{swap[1], swap[2], swap[3], hole[0], hole[1]}
		} else if swap[0] == 2 && swap[1] == 3 && swap[2] == 4 && swap[3] == 5 && swap[4] == 14 && hole[0] == 6 {
			swap = []int{swap[0], swap[1], swap[2], swap[3], hole[0]}
		} else if swap[0] == 2 && swap[1] == 3 && swap[2] == 4 && swap[3] == 5 && swap[4] == 14 && hole[1] == 6 {
			swap = []int{swap[0], swap[1], swap[2], hole[3], hole[1]}
		} else if hand[0] == swap[0]+1 && hand[1] == swap[1]+1 && hand[2] == swap[2]+1 && hand[3] == swap[3]+1 && hand[4] == swap[4]+1 {
			swap = []int{hand[0], hand[1], hand[2], hand[3], hand[4]}
		}
		/// If full house
	} else if rank == 4 && hole[0] == hole[1] {
		if hole[0] > swap[4] && hole[1] > swap[3] && swap[4] != swap[2] {
			swap = []int{swap[0], swap[1], swap[2], hole[1], hole[0]}
		} else if hole[0] > swap[0] && hole[1] > swap[1] && swap[0] != swap[2] {
			swap = []int{hole[0], hole[1], swap[2], swap[3], swap[4]}
		}
		/// Left overs
	} else if rank == 10 || rank == 9 || rank == 8 || rank == 7 || rank == 5 || rank == 4 || rank == 3 {

		if hand[4] >= swap[4] && hand[3] >= swap[3] && hand[2] >= swap[2] && hand[1] >= swap[1] && hand[0] >= swap[0] {
			swap = []int{hand[0], hand[1], hand[2], hand[3], hand[4]}
		}

	}

	return swap[:]
}

// Gets high pair and low pair from HighCardArr
func getHighPair(h [5]int) (int, int) {
	var highPair int
	var lowPair int

	for i := 0; i < 4; i++ {
		if h[i] == h[i+1] {
			if h[i] > highPair {
				highPair = h[i]
			}
		}
	}

	for i := 0; i < 4; i++ {
		if h[i] == h[i+1] {
			if h[i] < highPair {
				lowPair = h[i]
			}
		}
	}
	return highPair, lowPair
}

// Determines player hand rank after suit slipt
func makeHand(h, s []int) int {
	pHand := h
	pSuits := s

	sort.Ints(pHand)

	/// Royal flush
	if pHand[0] == 10 && pHand[1] == 11 && pHand[2] == 12 && pHand[3] == 13 && pHand[4] == 14 &&
		pSuits[0] == pSuits[1] && pSuits[0] == pSuits[2] && pSuits[0] == pSuits[3] && pSuits[0] == pSuits[4] {

		return 1

	}

	/// Straight flush
	if (pHand[0]+1 == pHand[1] && pHand[1]+1 == pHand[2] && pHand[2]+1 == pHand[3] && pHand[3]+1 == pHand[4] && pHand[0]+4 == pHand[4] &&
		pSuits[0] == pSuits[1] && pSuits[0] == pSuits[2] && pSuits[0] == pSuits[3] && pSuits[0] == pSuits[4]) ||
		(pHand[0] == 2 && pHand[1] == 3 && pHand[2] == 4 && pHand[3] == 5 && pHand[4] == 14 &&
			pSuits[0] == pSuits[1] && pSuits[0] == pSuits[2] && pSuits[0] == pSuits[3] && pSuits[0] == pSuits[4]) {

		return 2
	}

	/// Four of a Kind
	if (pHand[0] == pHand[1] && pHand[1] == pHand[2] && pHand[2] == pHand[3]) ||
		(pHand[1] == pHand[2] && pHand[2] == pHand[3] && pHand[3] == pHand[4]) {

		return 3
	}

	/// Full House
	if (pHand[0] == pHand[1] && pHand[1] == pHand[2] && pHand[3] == pHand[4]) ||
		(pHand[0] == pHand[1] && pHand[2] == pHand[3] && pHand[3] == pHand[4]) {

		return 4
	}

	/// Flush
	if pSuits[0] == pSuits[1] && pSuits[0] == pSuits[2] && pSuits[0] == pSuits[3] && pSuits[0] == pSuits[4] {

		return 5
	}

	/// Straight
	if pHand[0]+1 == pHand[1] && pHand[1]+1 == pHand[2] && pHand[2]+1 == pHand[3] && pHand[3]+1 == pHand[4] && pHand[0]+4 == pHand[4] ||
		pHand[0] == 2 && pHand[1] == 3 && pHand[2] == 4 && pHand[3] == 5 && pHand[4] == 14 {
		return 6
	}

	/// Three of a Kind
	if (pHand[0] == pHand[1] && pHand[1] == pHand[2]) ||
		(pHand[1] == pHand[2] && pHand[2] == pHand[3]) ||
		(pHand[2] == pHand[3] && pHand[3] == pHand[4]) {
		return 7
	}

	/// Two Pair
	if (pHand[0] == pHand[1] && pHand[2] == pHand[3]) ||
		(pHand[1] == pHand[2] && pHand[3] == pHand[4]) ||
		(pHand[0] == pHand[1] && pHand[3] == pHand[4]) {
		return 8
	}

	/// Pair
	if pHand[0] == pHand[1] || pHand[0] == pHand[2] || pHand[0] == pHand[3] || pHand[0] == pHand[4] ||
		pHand[1] == pHand[2] || pHand[1] == pHand[3] || pHand[1] == pHand[4] ||
		pHand[2] == pHand[3] || pHand[2] == pHand[4] ||
		pHand[3] == pHand[4] {
		return 9
	} else {

		return 10
	}
}

// Convert hand rank int to display string
func handToText(rank int) string {
	var handRankText string
	switch rank {
	case 1:
		handRankText = "Royal Flush"
	case 2:
		handRankText = "Straight Flush"
	case 3:
		handRankText = "Four of a Kind"
	case 4:
		handRankText = "Full House"
	case 5:
		handRankText = "Flush"
	case 6:
		handRankText = "Straight"
	case 7:
		handRankText = "Three of a Kind"
	case 8:
		handRankText = "Two Pair"
	case 9:
		handRankText = "Pair"
	case 10:
		handRankText = "High Card"
	}
	return handRankText
}

// Highlights winning Holdero hand at showdown
func highlightHand(a, b int) (hand []int) {
	rank := 100
	community := []int{}
	if Round.Flop1 > 0 {
		community = append(community, Round.Flop1)
	}

	if Round.Flop2 > 0 {
		community = append(community, Round.Flop2)
	}

	if Round.Flop3 > 0 {
		community = append(community, Round.Flop3)
	}

	if Round.TurnCard > 0 {
		community = append(community, Round.TurnCard)
	}

	if Round.RiverCard > 0 {
		community = append(community, Round.RiverCard)
	}

	hole := [2]int{a, b}
	cards := append(community, hole[:]...)
	p := Pool(5, cards)

	var check []int
	for i := range p {
		r, h, _ := compareTheseFive(p[i])
		if check == nil {
			hand = p[i]
			check = h
		}

		if r < rank {
			rank = r
			hand = p[i]
			check = h
		} else if r == rank {
			var better bool
			var r ranker
			r.pc1 = suitSplit(hole[0])
			r.pc2 = suitSplit(hole[1])
			new := findBest(rank, check, h, &r)
			for i := 0; i < 5; i++ {
				if new[i] != check[i] {
					better = true
					break
				}
			}

			if better {
				hand = p[i]
				check = new
			}
		}
	}

	return hand
}

// Payout to winning Holdero hand
func payWinningHand(w int, r *ranker) {
	hand := []int{}
	var winner string
	switch w {
	case 1:
		hand = append(hand, r.p1HandRaw[0], r.p1HandRaw[1])
		winner = "Player1"
	case 2:
		hand = append(hand, r.p2HandRaw[0], r.p2HandRaw[1])
		winner = "Player2"
	case 3:
		hand = append(hand, r.p3HandRaw[0], r.p3HandRaw[1])
		winner = "Player3"
	case 4:
		hand = append(hand, r.p4HandRaw[0], r.p4HandRaw[1])
		winner = "Player4"
	case 5:
		hand = append(hand, r.p5HandRaw[0], r.p5HandRaw[1])
		winner = "Player5"
	case 6:
		hand = append(hand, r.p6HandRaw[0], r.p6HandRaw[1])
		winner = "Player6"
	}

	Round.Winning_hand = highlightHand(hand[0], hand[1])
	updateStatsWins(Round.Pot, winner, false)

	if Round.ID == 1 {
		if !Signal.Paid {
			Signal.Paid = true
			go func() {
				time.Sleep(time.Duration(Times.Delay) * time.Second)
				retry := 0
				for retry < 4 {
					tx := PayOut(winner)
					time.Sleep(time.Second)
					retry += rpc.ConfirmTxRetry(tx, "Holdero", 36)
				}
			}()
		}
	}
}

// Main compare routine to determine Holdero winner
func compareAll(r *ranker) (end_res string) {
	winningRank := []int{r.p1Rank, r.p2Rank, r.p3Rank, r.p4Rank, r.p5Rank, r.p6Rank}
	sort.Ints(winningRank)

	if r.p1Rank < r.p2Rank && r.p1Rank < r.p3Rank && r.p1Rank < r.p4Rank && r.p1Rank < r.p5Rank && r.p1Rank < r.p6Rank { /// Outright win, player has highest rank
		end_res = Round.P1_name + " Wins with " + handToText(r.p1Rank)
		payWinningHand(1, r)
	} else if r.p2Rank < r.p1Rank && r.p2Rank < r.p3Rank && r.p2Rank < r.p4Rank && r.p2Rank < r.p5Rank && r.p2Rank < r.p6Rank {
		end_res = Round.P2_name + " Wins with " + handToText(r.p2Rank)
		payWinningHand(2, r)
	} else if r.p3Rank < r.p1Rank && r.p3Rank < r.p2Rank && r.p3Rank < r.p4Rank && r.p3Rank < r.p5Rank && r.p3Rank < r.p6Rank {
		end_res = Round.P3_name + " Wins with " + handToText(r.p3Rank)
		payWinningHand(3, r)
	} else if r.p4Rank < r.p1Rank && r.p4Rank < r.p2Rank && r.p4Rank < r.p3Rank && r.p4Rank < r.p5Rank && r.p4Rank < r.p6Rank {
		end_res = Round.P4_name + " Wins with " + handToText(r.p4Rank)
		payWinningHand(4, r)
	} else if r.p5Rank < r.p1Rank && r.p5Rank < r.p2Rank && r.p5Rank < r.p3Rank && r.p5Rank < r.p4Rank && r.p5Rank < r.p6Rank {
		end_res = Round.P5_name + " Wins with " + handToText(r.p5Rank)
		payWinningHand(5, r)
	} else if r.p6Rank < r.p1Rank && r.p6Rank < r.p2Rank && r.p6Rank < r.p3Rank && r.p6Rank < r.p4Rank && r.p6Rank < r.p5Rank {
		end_res = Round.P6_name + " Wins with" + handToText(r.p6Rank)
		payWinningHand(6, r)
	} else {

		highestPair := []int{r.p1HighPair, r.p2HighPair, r.p3HighPair, r.p4HighPair, r.p5HighPair, r.p6HighPair}
		sort.Ints(highestPair)

		if r.p1Rank != winningRank[0] || (winningRank[0] == 9 && r.p1HighPair != highestPair[5]) { /// If player hand is not the highest rank or if doesn't have high pair strip cards
			less1(r)
		}

		if r.p2Rank != winningRank[0] || (winningRank[0] == 9 && r.p2HighPair != highestPair[5]) {
			less2(r)
		}

		if r.p3Rank != winningRank[0] || (winningRank[0] == 9 && r.p3HighPair != highestPair[5]) {
			less3(r)
		}

		if (r.p4Rank != winningRank[0]) || (winningRank[0] == 9 && r.p4HighPair != highestPair[5]) {
			less4(r)
		}

		if r.p5Rank != winningRank[0] || (winningRank[0] == 9 && r.p5HighPair != highestPair[5]) {
			less5(r)
		}

		if r.p6Rank != winningRank[0] || (winningRank[0] == 9 && r.p6HighPair != highestPair[5]) {
			less6(r)
		}

		if winningRank[0] == 10 { /// Compares and strips loosing hands in high card situations
			compare1_2(r)
			compare2_1(r)
			if r.p1HighCardArr[4] > r.p2HighCardArr[4] {
				compare3_1(r)
				compare1_3(r)
			} else {
				compare3_2(r)
				compare2_3(r)
			}

			if r.p1HighCardArr[4] > r.p3HighCardArr[4] {
				compare1_4(r)
				compare4_1(r)
			} else if r.p2HighCardArr[4] > r.p3HighCardArr[4] {
				compare2_4(r)
				compare4_2(r)
			} else {
				compare3_4(r)
				compare4_3(r)
			}

			if r.p1HighCardArr[4] > r.p4HighCardArr[4] {
				compare1_5(r)
				compare5_1(r)
			} else if r.p2HighCardArr[4] > r.p4HighCardArr[4] {
				compare2_5(r)
				compare5_2(r)
			} else if r.p3HighCardArr[4] > r.p4HighCardArr[4] {
				compare3_5(r)
				compare5_3(r)
			} else {
				compare4_5(r)
				compare5_4(r)
			}

			if r.p1HighCardArr[4] > r.p5HighCardArr[4] {
				compare1_6(r)
				compare6_1(r)
			} else if r.p2HighCardArr[4] > r.p5HighCardArr[4] {
				compare2_6(r)
				compare6_2(r)
			} else if r.p3HighCardArr[4] > r.p5HighCardArr[4] {
				compare3_6(r)
				compare6_3(r)
			} else if r.p4HighCardArr[4] > r.p5HighCardArr[4] {
				compare4_6(r)
				compare6_4(r)
			} else {
				compare5_6(r)
				compare6_5(r)
			}
		}

		if r.p1HighPair > r.p2HighPair && r.p1HighPair > r.p3HighPair && r.p1HighPair > r.p4HighPair && r.p1HighPair > r.p5HighPair && r.p1HighPair > r.p6HighPair { /// No outright win, highest pairing first used to compare two hands of same rank
			if r.p1Rank == winningRank[0] {
				end_res = Round.P1_name + " Wins with " + handToText(r.p1Rank)
				payWinningHand(1, r)
			}

		} else if r.p2HighPair > r.p1HighPair && r.p2HighPair > r.p3HighPair && r.p2HighPair > r.p4HighPair && r.p2HighPair > r.p5HighPair && r.p2HighPair > r.p6HighPair {
			if r.p2Rank == winningRank[0] {
				end_res = Round.P2_name + " Wins with " + handToText(r.p2Rank)
				payWinningHand(2, r)
			}

		} else if r.p3HighPair > r.p1HighPair && r.p3HighPair > r.p2HighPair && r.p3HighPair > r.p4HighPair && r.p3HighPair > r.p5HighPair && r.p3HighPair > r.p6HighPair {
			if r.p3Rank == winningRank[0] {
				end_res = Round.P3_name + " Wins with " + handToText(r.p3Rank)
				payWinningHand(3, r)
			}

		} else if r.p4HighPair > r.p1HighPair && r.p4HighPair > r.p2HighPair && r.p4HighPair > r.p3HighPair && r.p4HighPair > r.p5HighPair && r.p4HighPair > r.p6HighPair {
			if r.p4Rank == winningRank[0] {
				end_res = Round.P4_name + " Wins with " + handToText(r.p4Rank)
				payWinningHand(4, r)
			}
		} else if r.p5HighPair > r.p1HighPair && r.p5HighPair > r.p2HighPair && r.p5HighPair > r.p3HighPair && r.p5HighPair > r.p4HighPair && r.p5HighPair > r.p6HighPair {
			if r.p5Rank == winningRank[0] {
				end_res = Round.P5_name + " Wins with " + handToText(r.p5Rank)
				payWinningHand(5, r)
			}
		} else if r.p6HighPair > r.p1HighPair && r.p6HighPair > r.p2HighPair && r.p6HighPair > r.p3HighPair && r.p6HighPair > r.p4HighPair && r.p6HighPair > r.p5HighPair {
			if r.p6Rank == winningRank[0] {
				end_res = Round.P6_name + " Wins with " + handToText(r.p6Rank)
				payWinningHand(6, r)
			}
			/// no high pair winner, if two pair winning rank look for low pair winner
		} else if winningRank[0] == 8 && r.p1LowPair > r.p2LowPair && r.p1LowPair > r.p3LowPair && r.p1LowPair > r.p4LowPair && r.p1LowPair > r.p5LowPair && r.p1LowPair > r.p6LowPair {
			if r.p1Rank == winningRank[0] {
				end_res = Round.P1_name + " Wins with " + handToText(r.p1Rank)
				payWinningHand(1, r)
			}
		} else if winningRank[0] == 8 && r.p2LowPair > r.p1LowPair && r.p2LowPair > r.p3LowPair && r.p2LowPair > r.p4LowPair && r.p2LowPair > r.p5LowPair && r.p2LowPair > r.p6LowPair {
			if r.p2Rank == winningRank[0] {
				end_res = Round.P2_name + " Wins with " + handToText(r.p2Rank)
				payWinningHand(2, r)
			}
		} else if winningRank[0] == 8 && r.p3LowPair > r.p1LowPair && r.p3LowPair > r.p2LowPair && r.p3LowPair > r.p4LowPair && r.p3LowPair > r.p5LowPair && r.p3LowPair > r.p6LowPair {
			if r.p3Rank == winningRank[0] {
				end_res = Round.P3_name + " Wins with " + handToText(r.p3Rank)
				payWinningHand(3, r)
			}
		} else if winningRank[0] == 8 && r.p4LowPair > r.p1LowPair && r.p4LowPair > r.p2LowPair && r.p4LowPair > r.p3LowPair && r.p4LowPair > r.p5LowPair && r.p4LowPair > r.p6LowPair {
			if r.p4Rank == winningRank[0] {
				end_res = Round.P4_name + " Wins with " + handToText(r.p4Rank)
				payWinningHand(4, r)
			}
		} else if winningRank[0] == 8 && r.p5LowPair > r.p1LowPair && r.p5LowPair > r.p2LowPair && r.p5LowPair > r.p3LowPair && r.p5LowPair > r.p4LowPair && r.p5LowPair > r.p6LowPair {
			if r.p5Rank == winningRank[0] {
				end_res = Round.P5_name + " Wins with " + handToText(r.p5Rank)
				payWinningHand(5, r)
			}
		} else if winningRank[0] == 8 && r.p6LowPair > r.p1LowPair && r.p6LowPair > r.p2LowPair && r.p6LowPair > r.p3LowPair && r.p6LowPair > r.p4LowPair && r.p6LowPair > r.p5LowPair {
			if r.p6Rank == winningRank[0] {
				end_res = Round.P6_name + " Wins with " + handToText(r.p6Rank)
				payWinningHand(6, r)
			}

			/// No outright or HighPair win so we compare all left over hands to determine HighCard winner
		} else if (r.p1HighCardArr[4] > r.p2HighCardArr[4] && r.p1HighCardArr[4] > r.p3HighCardArr[4] && r.p1HighCardArr[4] > r.p4HighCardArr[4] && r.p1HighCardArr[4] > r.p5HighCardArr[4] && r.p1HighCardArr[4] > r.p6HighCardArr[4]) ||

			(r.p1HighCardArr[4] >= r.p2HighCardArr[4] && r.p1HighCardArr[4] >= r.p3HighCardArr[4] && r.p1HighCardArr[4] >= r.p4HighCardArr[4] && r.p1HighCardArr[4] >= r.p5HighCardArr[4] && r.p1HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p1HighCardArr[3] > r.p2HighCardArr[3] && r.p1HighCardArr[3] > r.p3HighCardArr[3] && r.p1HighCardArr[3] > r.p4HighCardArr[3] && r.p1HighCardArr[3] > r.p5HighCardArr[3] && r.p1HighCardArr[3] > r.p6HighCardArr[3]) ||

			(r.p1HighCardArr[4] >= r.p2HighCardArr[4] && r.p1HighCardArr[4] >= r.p3HighCardArr[4] && r.p1HighCardArr[4] >= r.p4HighCardArr[4] && r.p1HighCardArr[4] >= r.p5HighCardArr[4] && r.p1HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p1HighCardArr[3] >= r.p2HighCardArr[3] && r.p1HighCardArr[3] >= r.p3HighCardArr[3] && r.p1HighCardArr[3] >= r.p4HighCardArr[3] && r.p1HighCardArr[3] >= r.p5HighCardArr[3] && r.p1HighCardArr[3] >= r.p6HighCardArr[3] &&
				r.p1HighCardArr[2] > r.p2HighCardArr[2] && r.p1HighCardArr[2] > r.p3HighCardArr[2] && r.p1HighCardArr[2] > r.p4HighCardArr[2] && r.p1HighCardArr[2] > r.p5HighCardArr[2] && r.p1HighCardArr[2] > r.p6HighCardArr[2]) ||

			(r.p1HighCardArr[4] >= r.p2HighCardArr[4] && r.p1HighCardArr[4] >= r.p3HighCardArr[4] && r.p1HighCardArr[4] >= r.p4HighCardArr[4] && r.p1HighCardArr[4] >= r.p5HighCardArr[4] && r.p1HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p1HighCardArr[3] >= r.p2HighCardArr[3] && r.p1HighCardArr[3] >= r.p3HighCardArr[3] && r.p1HighCardArr[3] >= r.p4HighCardArr[3] && r.p1HighCardArr[3] >= r.p5HighCardArr[3] && r.p1HighCardArr[3] >= r.p6HighCardArr[3] &&
				r.p1HighCardArr[2] >= r.p2HighCardArr[2] && r.p1HighCardArr[2] >= r.p3HighCardArr[2] && r.p1HighCardArr[2] >= r.p4HighCardArr[2] && r.p1HighCardArr[2] >= r.p5HighCardArr[2] && r.p1HighCardArr[2] >= r.p6HighCardArr[2] &&
				r.p1HighCardArr[1] > r.p2HighCardArr[1] && r.p1HighCardArr[1] > r.p3HighCardArr[1] && r.p1HighCardArr[1] > r.p4HighCardArr[1] && r.p1HighCardArr[1] > r.p5HighCardArr[1] && r.p1HighCardArr[1] > r.p6HighCardArr[1]) ||

			(r.p1HighCardArr[4] >= r.p2HighCardArr[4] && r.p1HighCardArr[4] >= r.p3HighCardArr[4] && r.p1HighCardArr[4] >= r.p4HighCardArr[4] && r.p1HighCardArr[4] >= r.p5HighCardArr[4] && r.p1HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p1HighCardArr[3] >= r.p2HighCardArr[3] && r.p1HighCardArr[3] >= r.p3HighCardArr[3] && r.p1HighCardArr[3] >= r.p4HighCardArr[3] && r.p1HighCardArr[3] >= r.p5HighCardArr[3] && r.p1HighCardArr[3] >= r.p6HighCardArr[3] &&
				r.p1HighCardArr[2] >= r.p2HighCardArr[2] && r.p1HighCardArr[2] >= r.p3HighCardArr[2] && r.p1HighCardArr[2] >= r.p4HighCardArr[2] && r.p1HighCardArr[2] >= r.p5HighCardArr[2] && r.p1HighCardArr[2] >= r.p6HighCardArr[2] &&
				r.p1HighCardArr[1] >= r.p2HighCardArr[1] && r.p1HighCardArr[1] >= r.p3HighCardArr[1] && r.p1HighCardArr[1] >= r.p4HighCardArr[1] && r.p1HighCardArr[1] >= r.p5HighCardArr[1] && r.p1HighCardArr[1] >= r.p6HighCardArr[1] &&
				r.p1HighCardArr[0] > r.p2HighCardArr[0] && r.p1HighCardArr[0] > r.p3HighCardArr[0] && r.p1HighCardArr[0] > r.p4HighCardArr[0] && r.p1HighCardArr[0] > r.p5HighCardArr[0] && r.p1HighCardArr[0] > r.p6HighCardArr[0]) {

			if r.p1Rank == winningRank[0] {
				end_res = Round.P1_name + " Wins with " + handToText(r.p1Rank)
				payWinningHand(1, r)
			}

		} else if (r.p2HighCardArr[4] > r.p1HighCardArr[4] && r.p2HighCardArr[4] > r.p3HighCardArr[4] && r.p2HighCardArr[4] > r.p4HighCardArr[4] && r.p2HighCardArr[4] > r.p5HighCardArr[4] && r.p2HighCardArr[4] > r.p6HighCardArr[4]) ||

			(r.p2HighCardArr[4] >= r.p1HighCardArr[4] && r.p2HighCardArr[4] >= r.p3HighCardArr[4] && r.p2HighCardArr[4] >= r.p4HighCardArr[4] && r.p2HighCardArr[4] >= r.p5HighCardArr[4] && r.p2HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p2HighCardArr[3] > r.p1HighCardArr[3] && r.p2HighCardArr[3] > r.p3HighCardArr[3] && r.p2HighCardArr[3] > r.p4HighCardArr[3] && r.p2HighCardArr[3] > r.p5HighCardArr[3] && r.p2HighCardArr[3] > r.p6HighCardArr[3]) ||

			(r.p2HighCardArr[4] >= r.p1HighCardArr[4] && r.p2HighCardArr[4] >= r.p3HighCardArr[4] && r.p2HighCardArr[4] >= r.p4HighCardArr[4] && r.p2HighCardArr[4] >= r.p5HighCardArr[4] && r.p2HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p2HighCardArr[3] >= r.p1HighCardArr[3] && r.p2HighCardArr[3] >= r.p3HighCardArr[3] && r.p2HighCardArr[3] >= r.p4HighCardArr[3] && r.p2HighCardArr[3] >= r.p5HighCardArr[3] && r.p2HighCardArr[3] >= r.p6HighCardArr[3] &&
				r.p2HighCardArr[2] > r.p1HighCardArr[2] && r.p2HighCardArr[2] > r.p3HighCardArr[2] && r.p2HighCardArr[2] > r.p4HighCardArr[2] && r.p2HighCardArr[2] > r.p5HighCardArr[2] && r.p2HighCardArr[2] > r.p6HighCardArr[2]) ||

			(r.p2HighCardArr[4] >= r.p1HighCardArr[4] && r.p2HighCardArr[4] >= r.p3HighCardArr[4] && r.p2HighCardArr[4] >= r.p4HighCardArr[4] && r.p2HighCardArr[4] >= r.p5HighCardArr[4] && r.p2HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p2HighCardArr[3] >= r.p1HighCardArr[3] && r.p2HighCardArr[3] >= r.p3HighCardArr[3] && r.p2HighCardArr[3] >= r.p4HighCardArr[3] && r.p2HighCardArr[3] >= r.p5HighCardArr[3] && r.p2HighCardArr[3] >= r.p6HighCardArr[3] &&
				r.p2HighCardArr[2] >= r.p1HighCardArr[2] && r.p2HighCardArr[2] >= r.p3HighCardArr[2] && r.p2HighCardArr[2] >= r.p4HighCardArr[2] && r.p2HighCardArr[2] >= r.p5HighCardArr[2] && r.p2HighCardArr[2] >= r.p6HighCardArr[2] &&
				r.p2HighCardArr[1] > r.p1HighCardArr[1] && r.p2HighCardArr[1] > r.p3HighCardArr[1] && r.p2HighCardArr[1] > r.p4HighCardArr[1] && r.p2HighCardArr[1] > r.p5HighCardArr[1] && r.p2HighCardArr[1] > r.p6HighCardArr[1]) ||

			(r.p2HighCardArr[4] >= r.p1HighCardArr[4] && r.p2HighCardArr[4] >= r.p3HighCardArr[4] && r.p2HighCardArr[4] >= r.p4HighCardArr[4] && r.p2HighCardArr[4] >= r.p5HighCardArr[4] && r.p2HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p2HighCardArr[3] >= r.p1HighCardArr[3] && r.p2HighCardArr[3] >= r.p3HighCardArr[3] && r.p2HighCardArr[3] >= r.p4HighCardArr[3] && r.p2HighCardArr[3] >= r.p5HighCardArr[3] && r.p2HighCardArr[3] >= r.p6HighCardArr[3] &&
				r.p2HighCardArr[2] >= r.p1HighCardArr[2] && r.p2HighCardArr[2] >= r.p3HighCardArr[2] && r.p2HighCardArr[2] >= r.p4HighCardArr[2] && r.p2HighCardArr[2] >= r.p5HighCardArr[2] && r.p2HighCardArr[2] >= r.p6HighCardArr[2] &&
				r.p2HighCardArr[1] >= r.p1HighCardArr[1] && r.p2HighCardArr[1] >= r.p3HighCardArr[1] && r.p2HighCardArr[1] >= r.p4HighCardArr[1] && r.p2HighCardArr[1] >= r.p5HighCardArr[1] && r.p2HighCardArr[1] >= r.p6HighCardArr[1] &&
				r.p2HighCardArr[0] > r.p1HighCardArr[0] && r.p2HighCardArr[0] > r.p3HighCardArr[0] && r.p2HighCardArr[0] > r.p4HighCardArr[0] && r.p2HighCardArr[0] > r.p5HighCardArr[0] && r.p2HighCardArr[0] > r.p6HighCardArr[0]) {

			if r.p2Rank == winningRank[0] {
				end_res = Round.P2_name + " Wins with " + handToText(r.p2Rank)
				payWinningHand(2, r)

			}

		} else if (r.p3HighCardArr[4] > r.p1HighCardArr[4] && r.p3HighCardArr[4] > r.p2HighCardArr[4] && r.p3HighCardArr[4] > r.p4HighCardArr[4] && r.p3HighCardArr[4] > r.p5HighCardArr[4] && r.p3HighCardArr[4] > r.p6HighCardArr[4]) ||

			(r.p3HighCardArr[4] >= r.p1HighCardArr[4] && r.p3HighCardArr[4] >= r.p2HighCardArr[4] && r.p3HighCardArr[4] >= r.p4HighCardArr[4] && r.p3HighCardArr[4] >= r.p5HighCardArr[4] && r.p3HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p3HighCardArr[3] > r.p1HighCardArr[3] && r.p3HighCardArr[3] > r.p2HighCardArr[3] && r.p3HighCardArr[3] > r.p4HighCardArr[3] && r.p3HighCardArr[3] > r.p5HighCardArr[3] && r.p3HighCardArr[3] > r.p6HighCardArr[3]) ||

			(r.p3HighCardArr[4] >= r.p1HighCardArr[4] && r.p3HighCardArr[4] >= r.p2HighCardArr[4] && r.p3HighCardArr[4] >= r.p4HighCardArr[4] && r.p3HighCardArr[4] >= r.p5HighCardArr[4] && r.p3HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p3HighCardArr[3] >= r.p1HighCardArr[3] && r.p3HighCardArr[3] >= r.p2HighCardArr[3] && r.p3HighCardArr[3] >= r.p4HighCardArr[3] && r.p3HighCardArr[3] >= r.p5HighCardArr[3] && r.p3HighCardArr[3] >= r.p6HighCardArr[3] &&
				r.p3HighCardArr[2] > r.p1HighCardArr[2] && r.p3HighCardArr[2] > r.p2HighCardArr[2] && r.p3HighCardArr[2] > r.p4HighCardArr[2] && r.p3HighCardArr[2] > r.p5HighCardArr[2] && r.p3HighCardArr[2] > r.p6HighCardArr[2]) ||

			(r.p3HighCardArr[4] >= r.p1HighCardArr[4] && r.p3HighCardArr[4] >= r.p2HighCardArr[4] && r.p3HighCardArr[4] >= r.p4HighCardArr[4] && r.p3HighCardArr[4] >= r.p5HighCardArr[4] && r.p3HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p3HighCardArr[3] >= r.p1HighCardArr[3] && r.p3HighCardArr[3] >= r.p2HighCardArr[3] && r.p3HighCardArr[3] >= r.p4HighCardArr[3] && r.p3HighCardArr[3] >= r.p5HighCardArr[3] && r.p3HighCardArr[3] >= r.p6HighCardArr[3] &&
				r.p3HighCardArr[2] >= r.p1HighCardArr[2] && r.p3HighCardArr[2] >= r.p2HighCardArr[2] && r.p3HighCardArr[2] >= r.p4HighCardArr[2] && r.p3HighCardArr[2] >= r.p5HighCardArr[2] && r.p3HighCardArr[2] >= r.p6HighCardArr[2] &&
				r.p3HighCardArr[1] > r.p1HighCardArr[1] && r.p3HighCardArr[1] > r.p2HighCardArr[1] && r.p3HighCardArr[1] > r.p4HighCardArr[1] && r.p3HighCardArr[1] > r.p5HighCardArr[1] && r.p3HighCardArr[1] > r.p6HighCardArr[1]) ||

			(r.p3HighCardArr[4] >= r.p1HighCardArr[4] && r.p3HighCardArr[4] >= r.p2HighCardArr[4] && r.p3HighCardArr[4] >= r.p4HighCardArr[4] && r.p3HighCardArr[4] >= r.p5HighCardArr[4] && r.p3HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p3HighCardArr[3] >= r.p1HighCardArr[3] && r.p3HighCardArr[3] >= r.p2HighCardArr[3] && r.p3HighCardArr[3] >= r.p4HighCardArr[3] && r.p3HighCardArr[3] >= r.p5HighCardArr[3] && r.p3HighCardArr[3] >= r.p6HighCardArr[3] &&
				r.p3HighCardArr[2] >= r.p1HighCardArr[2] && r.p3HighCardArr[2] >= r.p2HighCardArr[2] && r.p3HighCardArr[2] >= r.p4HighCardArr[2] && r.p3HighCardArr[2] >= r.p5HighCardArr[2] && r.p3HighCardArr[2] >= r.p6HighCardArr[2] &&
				r.p3HighCardArr[1] >= r.p1HighCardArr[1] && r.p3HighCardArr[1] >= r.p2HighCardArr[1] && r.p3HighCardArr[1] >= r.p4HighCardArr[1] && r.p3HighCardArr[1] >= r.p5HighCardArr[1] && r.p3HighCardArr[1] >= r.p6HighCardArr[1] &&
				r.p3HighCardArr[0] > r.p1HighCardArr[0] && r.p3HighCardArr[0] > r.p2HighCardArr[0] && r.p3HighCardArr[0] > r.p4HighCardArr[0] && r.p3HighCardArr[0] > r.p5HighCardArr[0] && r.p3HighCardArr[0] > r.p6HighCardArr[0]) {

			if r.p3Rank == winningRank[0] {
				end_res = Round.P3_name + " Wins with " + handToText(r.p3Rank)
				payWinningHand(3, r)

			}

		} else if (r.p4HighCardArr[4] > r.p1HighCardArr[4] && r.p4HighCardArr[4] > r.p2HighCardArr[4] && r.p4HighCardArr[4] > r.p3HighCardArr[4] && r.p4HighCardArr[4] > r.p5HighCardArr[4] && r.p4HighCardArr[4] > r.p6HighCardArr[4]) ||

			(r.p4HighCardArr[4] >= r.p1HighCardArr[4] && r.p4HighCardArr[4] >= r.p2HighCardArr[4] && r.p4HighCardArr[4] >= r.p3HighCardArr[4] && r.p4HighCardArr[4] >= r.p5HighCardArr[4] && r.p4HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p4HighCardArr[3] > r.p1HighCardArr[3] && r.p4HighCardArr[3] > r.p2HighCardArr[3] && r.p4HighCardArr[3] > r.p3HighCardArr[3] && r.p4HighCardArr[3] > r.p5HighCardArr[3] && r.p4HighCardArr[3] > r.p6HighCardArr[3]) ||

			(r.p4HighCardArr[4] >= r.p1HighCardArr[4] && r.p4HighCardArr[4] >= r.p2HighCardArr[4] && r.p4HighCardArr[4] >= r.p3HighCardArr[4] && r.p4HighCardArr[4] >= r.p5HighCardArr[4] && r.p4HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p4HighCardArr[3] >= r.p1HighCardArr[3] && r.p4HighCardArr[3] >= r.p2HighCardArr[3] && r.p4HighCardArr[3] >= r.p3HighCardArr[3] && r.p4HighCardArr[3] >= r.p5HighCardArr[3] && r.p4HighCardArr[3] >= r.p6HighCardArr[3] &&
				r.p4HighCardArr[2] > r.p1HighCardArr[2] && r.p4HighCardArr[2] > r.p2HighCardArr[2] && r.p4HighCardArr[2] > r.p3HighCardArr[2] && r.p4HighCardArr[2] > r.p5HighCardArr[2] && r.p4HighCardArr[2] > r.p6HighCardArr[2]) ||

			(r.p4HighCardArr[4] >= r.p1HighCardArr[4] && r.p4HighCardArr[4] >= r.p2HighCardArr[4] && r.p4HighCardArr[4] >= r.p3HighCardArr[4] && r.p4HighCardArr[4] >= r.p5HighCardArr[4] && r.p4HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p4HighCardArr[3] >= r.p1HighCardArr[3] && r.p4HighCardArr[3] >= r.p2HighCardArr[3] && r.p4HighCardArr[3] >= r.p3HighCardArr[3] && r.p4HighCardArr[3] >= r.p5HighCardArr[3] && r.p4HighCardArr[3] >= r.p6HighCardArr[3] &&
				r.p4HighCardArr[2] >= r.p1HighCardArr[2] && r.p4HighCardArr[2] >= r.p2HighCardArr[2] && r.p4HighCardArr[2] >= r.p3HighCardArr[2] && r.p4HighCardArr[2] >= r.p5HighCardArr[2] && r.p4HighCardArr[2] >= r.p6HighCardArr[2] &&
				r.p4HighCardArr[1] > r.p1HighCardArr[1] && r.p4HighCardArr[1] > r.p2HighCardArr[1] && r.p4HighCardArr[1] > r.p3HighCardArr[1] && r.p4HighCardArr[1] > r.p5HighCardArr[1] && r.p4HighCardArr[1] > r.p6HighCardArr[1]) ||

			(r.p4HighCardArr[4] >= r.p1HighCardArr[4] && r.p4HighCardArr[4] >= r.p2HighCardArr[4] && r.p4HighCardArr[4] >= r.p3HighCardArr[4] && r.p4HighCardArr[4] >= r.p5HighCardArr[4] && r.p4HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p4HighCardArr[3] >= r.p1HighCardArr[3] && r.p4HighCardArr[3] >= r.p2HighCardArr[3] && r.p4HighCardArr[3] >= r.p3HighCardArr[3] && r.p4HighCardArr[3] >= r.p5HighCardArr[3] && r.p4HighCardArr[3] >= r.p6HighCardArr[3] &&
				r.p4HighCardArr[2] >= r.p1HighCardArr[2] && r.p4HighCardArr[2] >= r.p2HighCardArr[2] && r.p4HighCardArr[2] >= r.p3HighCardArr[2] && r.p4HighCardArr[2] >= r.p5HighCardArr[2] && r.p4HighCardArr[2] >= r.p6HighCardArr[2] &&
				r.p4HighCardArr[1] >= r.p1HighCardArr[1] && r.p4HighCardArr[1] >= r.p2HighCardArr[1] && r.p4HighCardArr[1] >= r.p3HighCardArr[1] && r.p4HighCardArr[1] >= r.p5HighCardArr[1] && r.p4HighCardArr[1] >= r.p6HighCardArr[1] &&
				r.p4HighCardArr[0] > r.p1HighCardArr[0] && r.p4HighCardArr[0] > r.p2HighCardArr[0] && r.p4HighCardArr[0] > r.p3HighCardArr[0] && r.p4HighCardArr[0] > r.p5HighCardArr[0] && r.p4HighCardArr[0] > r.p6HighCardArr[0]) {

			if r.p4Rank == winningRank[0] {
				end_res = Round.P4_name + " Wins with " + handToText(r.p4Rank)
				payWinningHand(4, r)

			}

		} else if (r.p5HighCardArr[4] > r.p1HighCardArr[4] && r.p5HighCardArr[4] > r.p2HighCardArr[4] && r.p5HighCardArr[4] > r.p3HighCardArr[4] && r.p5HighCardArr[4] > r.p4HighCardArr[4] && r.p5HighCardArr[4] > r.p6HighCardArr[4]) ||

			(r.p5HighCardArr[4] >= r.p1HighCardArr[4] && r.p5HighCardArr[4] >= r.p2HighCardArr[4] && r.p5HighCardArr[4] >= r.p3HighCardArr[4] && r.p5HighCardArr[4] >= r.p4HighCardArr[4] && r.p5HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p5HighCardArr[3] > r.p1HighCardArr[3] && r.p5HighCardArr[3] > r.p2HighCardArr[3] && r.p5HighCardArr[3] > r.p3HighCardArr[3] && r.p5HighCardArr[3] > r.p4HighCardArr[3] && r.p5HighCardArr[3] > r.p6HighCardArr[3]) ||

			(r.p5HighCardArr[4] >= r.p1HighCardArr[4] && r.p5HighCardArr[4] >= r.p2HighCardArr[4] && r.p5HighCardArr[4] >= r.p3HighCardArr[4] && r.p5HighCardArr[4] >= r.p4HighCardArr[4] && r.p5HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p5HighCardArr[3] >= r.p1HighCardArr[3] && r.p5HighCardArr[3] >= r.p2HighCardArr[3] && r.p5HighCardArr[3] >= r.p3HighCardArr[3] && r.p5HighCardArr[3] >= r.p4HighCardArr[3] && r.p5HighCardArr[3] >= r.p6HighCardArr[3] &&
				r.p5HighCardArr[2] > r.p1HighCardArr[2] && r.p5HighCardArr[2] > r.p2HighCardArr[2] && r.p5HighCardArr[2] > r.p3HighCardArr[2] && r.p5HighCardArr[2] > r.p4HighCardArr[2] && r.p5HighCardArr[2] > r.p6HighCardArr[2]) ||

			(r.p5HighCardArr[4] >= r.p1HighCardArr[4] && r.p5HighCardArr[4] >= r.p2HighCardArr[4] && r.p5HighCardArr[4] >= r.p3HighCardArr[4] && r.p5HighCardArr[4] >= r.p4HighCardArr[4] && r.p5HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p5HighCardArr[3] >= r.p1HighCardArr[3] && r.p5HighCardArr[3] >= r.p2HighCardArr[3] && r.p5HighCardArr[3] >= r.p3HighCardArr[3] && r.p5HighCardArr[3] >= r.p4HighCardArr[3] && r.p5HighCardArr[3] >= r.p6HighCardArr[3] &&
				r.p5HighCardArr[2] >= r.p1HighCardArr[2] && r.p5HighCardArr[2] >= r.p2HighCardArr[2] && r.p5HighCardArr[2] >= r.p3HighCardArr[2] && r.p5HighCardArr[2] >= r.p4HighCardArr[2] && r.p5HighCardArr[2] >= r.p6HighCardArr[2] &&
				r.p5HighCardArr[1] > r.p1HighCardArr[1] && r.p5HighCardArr[1] > r.p2HighCardArr[1] && r.p5HighCardArr[1] > r.p3HighCardArr[1] && r.p5HighCardArr[1] > r.p4HighCardArr[1] && r.p5HighCardArr[1] > r.p6HighCardArr[1]) ||

			(r.p5HighCardArr[4] >= r.p1HighCardArr[4] && r.p5HighCardArr[4] >= r.p2HighCardArr[4] && r.p5HighCardArr[4] >= r.p3HighCardArr[4] && r.p5HighCardArr[4] >= r.p4HighCardArr[4] && r.p5HighCardArr[4] >= r.p6HighCardArr[4] &&
				r.p5HighCardArr[3] >= r.p1HighCardArr[3] && r.p5HighCardArr[3] >= r.p2HighCardArr[3] && r.p5HighCardArr[3] >= r.p3HighCardArr[3] && r.p5HighCardArr[3] >= r.p4HighCardArr[3] && r.p5HighCardArr[3] >= r.p6HighCardArr[3] &&
				r.p5HighCardArr[2] >= r.p1HighCardArr[2] && r.p5HighCardArr[2] >= r.p2HighCardArr[2] && r.p5HighCardArr[2] >= r.p3HighCardArr[2] && r.p5HighCardArr[2] >= r.p4HighCardArr[2] && r.p5HighCardArr[2] >= r.p6HighCardArr[2] &&
				r.p5HighCardArr[1] >= r.p1HighCardArr[1] && r.p5HighCardArr[1] >= r.p2HighCardArr[1] && r.p5HighCardArr[1] >= r.p3HighCardArr[1] && r.p5HighCardArr[1] >= r.p4HighCardArr[1] && r.p5HighCardArr[1] >= r.p6HighCardArr[1] &&
				r.p5HighCardArr[0] > r.p1HighCardArr[0] && r.p5HighCardArr[0] > r.p2HighCardArr[0] && r.p5HighCardArr[0] > r.p3HighCardArr[0] && r.p5HighCardArr[0] > r.p4HighCardArr[0] && r.p5HighCardArr[0] > r.p6HighCardArr[0]) {

			if r.p5Rank == winningRank[0] {
				end_res = Round.P5_name + " Wins with " + handToText(r.p5Rank)
				payWinningHand(5, r)

			}

		} else if (r.p6HighCardArr[4] > r.p1HighCardArr[4] && r.p6HighCardArr[4] > r.p2HighCardArr[4] && r.p6HighCardArr[4] > r.p3HighCardArr[4] && r.p6HighCardArr[4] > r.p4HighCardArr[4] && r.p6HighCardArr[4] > r.p5HighCardArr[4]) ||

			(r.p6HighCardArr[4] >= r.p1HighCardArr[4] && r.p6HighCardArr[4] >= r.p2HighCardArr[4] && r.p6HighCardArr[4] >= r.p3HighCardArr[4] && r.p6HighCardArr[4] >= r.p4HighCardArr[4] && r.p6HighCardArr[4] >= r.p5HighCardArr[4] &&
				r.p6HighCardArr[3] > r.p1HighCardArr[3] && r.p6HighCardArr[3] > r.p2HighCardArr[3] && r.p6HighCardArr[3] > r.p3HighCardArr[3] && r.p6HighCardArr[3] > r.p4HighCardArr[3] && r.p6HighCardArr[3] > r.p5HighCardArr[3]) ||

			(r.p6HighCardArr[4] >= r.p1HighCardArr[4] && r.p6HighCardArr[4] >= r.p2HighCardArr[4] && r.p6HighCardArr[4] >= r.p3HighCardArr[4] && r.p6HighCardArr[4] >= r.p4HighCardArr[4] && r.p6HighCardArr[4] >= r.p5HighCardArr[4] &&
				r.p6HighCardArr[3] >= r.p1HighCardArr[3] && r.p6HighCardArr[3] >= r.p2HighCardArr[3] && r.p6HighCardArr[3] >= r.p3HighCardArr[3] && r.p6HighCardArr[3] >= r.p4HighCardArr[3] && r.p6HighCardArr[3] >= r.p5HighCardArr[3] &&
				r.p6HighCardArr[2] > r.p1HighCardArr[2] && r.p6HighCardArr[2] > r.p2HighCardArr[2] && r.p6HighCardArr[2] > r.p3HighCardArr[2] && r.p6HighCardArr[2] > r.p4HighCardArr[2] && r.p6HighCardArr[2] > r.p5HighCardArr[2]) ||

			(r.p6HighCardArr[4] >= r.p1HighCardArr[4] && r.p6HighCardArr[4] >= r.p2HighCardArr[4] && r.p6HighCardArr[4] >= r.p3HighCardArr[4] && r.p6HighCardArr[4] >= r.p4HighCardArr[4] && r.p6HighCardArr[4] >= r.p5HighCardArr[4] &&
				r.p6HighCardArr[3] >= r.p1HighCardArr[3] && r.p6HighCardArr[3] >= r.p2HighCardArr[3] && r.p6HighCardArr[3] >= r.p3HighCardArr[3] && r.p6HighCardArr[3] >= r.p4HighCardArr[3] && r.p6HighCardArr[3] >= r.p5HighCardArr[3] &&
				r.p6HighCardArr[2] >= r.p1HighCardArr[2] && r.p6HighCardArr[2] >= r.p2HighCardArr[2] && r.p6HighCardArr[2] >= r.p3HighCardArr[2] && r.p6HighCardArr[2] >= r.p4HighCardArr[2] && r.p6HighCardArr[2] >= r.p5HighCardArr[2] &&
				r.p6HighCardArr[1] > r.p1HighCardArr[1] && r.p6HighCardArr[1] > r.p2HighCardArr[1] && r.p6HighCardArr[1] > r.p3HighCardArr[1] && r.p6HighCardArr[1] > r.p4HighCardArr[1] && r.p6HighCardArr[1] > r.p5HighCardArr[1]) ||

			(r.p6HighCardArr[4] >= r.p1HighCardArr[4] && r.p6HighCardArr[4] >= r.p2HighCardArr[4] && r.p6HighCardArr[4] >= r.p3HighCardArr[4] && r.p6HighCardArr[4] >= r.p4HighCardArr[4] && r.p6HighCardArr[4] >= r.p5HighCardArr[4] &&
				r.p6HighCardArr[3] >= r.p1HighCardArr[3] && r.p6HighCardArr[3] >= r.p2HighCardArr[3] && r.p6HighCardArr[3] >= r.p3HighCardArr[3] && r.p6HighCardArr[3] >= r.p4HighCardArr[3] && r.p6HighCardArr[3] >= r.p5HighCardArr[3] &&
				r.p6HighCardArr[2] >= r.p1HighCardArr[2] && r.p6HighCardArr[2] >= r.p2HighCardArr[2] && r.p6HighCardArr[2] >= r.p3HighCardArr[2] && r.p6HighCardArr[2] >= r.p4HighCardArr[2] && r.p6HighCardArr[2] >= r.p5HighCardArr[2] &&
				r.p6HighCardArr[1] >= r.p1HighCardArr[1] && r.p6HighCardArr[1] >= r.p2HighCardArr[1] && r.p6HighCardArr[1] >= r.p3HighCardArr[1] && r.p6HighCardArr[1] >= r.p4HighCardArr[1] && r.p6HighCardArr[1] >= r.p5HighCardArr[1] &&
				r.p6HighCardArr[0] > r.p1HighCardArr[0] && r.p6HighCardArr[0] > r.p2HighCardArr[0] && r.p6HighCardArr[0] > r.p3HighCardArr[0] && r.p6HighCardArr[0] > r.p4HighCardArr[0] && r.p6HighCardArr[0] > r.p5HighCardArr[0]) {

			if r.p6Rank == winningRank[0] {
				end_res = Round.P6_name + " Wins with " + handToText(r.p6Rank)
				payWinningHand(6, r)

			}

		} else {
			end_res = "Push"
			updateStatsPush(*r, Round.Pot, Round.F1, Round.F2, Round.F3, Round.F4, Round.F5, Round.F6)
			if Round.ID == 1 {
				if !Signal.Paid {
					Signal.Paid = true
					go func() {
						time.Sleep(time.Duration(Times.Delay) * time.Second)
						retry := 0
						for retry < 4 {
							tx := PayoutSplit(*r, Round.F1, Round.F2, Round.F3, Round.F4, Round.F5, Round.F6)
							time.Sleep(time.Second)
							retry += rpc.ConfirmTxRetry(tx, "Holdero", 36)
						}
					}()
				}
			}
		}
	}

	if !Round.Printed {
		Round.Printed = true
		rpc.AddLog(end_res)
	}

	return end_res
}

// Splits cards inside getHand() to make card value and suit value
func suitSplit(card int) [2]int {
	var arrSplit [2]int
	switch card {
	////// Spades
	case 1:
		arrSplit[0] = 14
		arrSplit[1] = 0

	case 2:
		arrSplit[0] = 2
		arrSplit[1] = 0

	case 3:
		arrSplit[0] = 3
		arrSplit[1] = 0

	case 4:
		arrSplit[0] = 4
		arrSplit[1] = 0

	case 5:
		arrSplit[0] = 5
		arrSplit[1] = 0

	case 6:
		arrSplit[0] = 6
		arrSplit[1] = 0

	case 7:
		arrSplit[0] = 7
		arrSplit[1] = 0

	case 8:
		arrSplit[0] = 8
		arrSplit[1] = 0

	case 9:
		arrSplit[0] = 9
		arrSplit[1] = 0

	case 10:
		arrSplit[0] = 10
		arrSplit[1] = 0

	case 11:
		arrSplit[0] = 11
		arrSplit[1] = 0

	case 12:
		arrSplit[0] = 12
		arrSplit[1] = 0

	case 13:
		arrSplit[0] = 13
		arrSplit[1] = 0

		////// Hearts
	case 14:
		arrSplit[0] = 14
		arrSplit[1] = 13

	case 15:
		arrSplit[0] = 2
		arrSplit[1] = 13

	case 16:
		arrSplit[0] = 3
		arrSplit[1] = 13

	case 17:
		arrSplit[0] = 4
		arrSplit[1] = 13

	case 18:
		arrSplit[0] = 5
		arrSplit[1] = 13

	case 19:
		arrSplit[0] = 6
		arrSplit[1] = 13

	case 20:
		arrSplit[0] = 7
		arrSplit[1] = 13

	case 21:
		arrSplit[0] = 8
		arrSplit[1] = 13

	case 22:
		arrSplit[0] = 9
		arrSplit[1] = 13

	case 23:
		arrSplit[0] = 10
		arrSplit[1] = 13

	case 24:
		arrSplit[0] = 11
		arrSplit[1] = 13

	case 25:
		arrSplit[0] = 12
		arrSplit[1] = 13

	case 26:
		arrSplit[0] = 13
		arrSplit[1] = 13

		////// Clubs
	case 27:
		arrSplit[0] = 14
		arrSplit[1] = 26

	case 28:
		arrSplit[0] = 2
		arrSplit[1] = 26

	case 29:
		arrSplit[0] = 3
		arrSplit[1] = 26

	case 30:
		arrSplit[0] = 4
		arrSplit[1] = 26

	case 31:
		arrSplit[0] = 5
		arrSplit[1] = 26

	case 32:
		arrSplit[0] = 6
		arrSplit[1] = 26

	case 33:
		arrSplit[0] = 7
		arrSplit[1] = 26

	case 34:
		arrSplit[0] = 8
		arrSplit[1] = 26

	case 35:
		arrSplit[0] = 9
		arrSplit[1] = 26

	case 36:
		arrSplit[0] = 10
		arrSplit[1] = 26

	case 37:
		arrSplit[0] = 11
		arrSplit[1] = 26

	case 38:
		arrSplit[0] = 12
		arrSplit[1] = 26

	case 39:
		arrSplit[0] = 13
		arrSplit[1] = 26

		////// Diamonds
	case 40:
		arrSplit[0] = 14
		arrSplit[1] = 39

	case 41:
		arrSplit[0] = 2
		arrSplit[1] = 39

	case 42:
		arrSplit[0] = 3
		arrSplit[1] = 39

	case 43:
		arrSplit[0] = 4
		arrSplit[1] = 39

	case 44:
		arrSplit[0] = 5
		arrSplit[1] = 39

	case 45:
		arrSplit[0] = 6
		arrSplit[1] = 39

	case 46:
		arrSplit[0] = 7
		arrSplit[1] = 39

	case 47:
		arrSplit[0] = 8
		arrSplit[1] = 39

	case 48:
		arrSplit[0] = 9
		arrSplit[1] = 39

	case 49:
		arrSplit[0] = 10
		arrSplit[1] = 39

	case 50:
		arrSplit[0] = 11
		arrSplit[1] = 39

	case 51:
		arrSplit[0] = 12
		arrSplit[1] = 39

	case 52:
		arrSplit[0] = 13
		arrSplit[1] = 39

	}

	return arrSplit
}

// Compare two individual hands from ranker for high card situations, if hand being compared is worse strip values

func compare1_2(r *ranker) {
	if (r.p1HighCardArr[4] > r.p2HighCardArr[4]) ||
		(r.p1HighCardArr[4] == r.p2HighCardArr[4] && r.p1HighCardArr[3] > r.p2HighCardArr[3]) ||
		(r.p1HighCardArr[4] == r.p2HighCardArr[4] && r.p1HighCardArr[3] == r.p2HighCardArr[3] && r.p1HighCardArr[2] > r.p2HighCardArr[2]) ||
		(r.p1HighCardArr[4] == r.p2HighCardArr[4] && r.p1HighCardArr[3] == r.p2HighCardArr[3] && r.p1HighCardArr[2] == r.p2HighCardArr[2] && r.p1HighCardArr[1] > r.p2HighCardArr[1]) ||
		(r.p1HighCardArr[4] == r.p2HighCardArr[4] && r.p1HighCardArr[3] == r.p2HighCardArr[3] && r.p1HighCardArr[2] == r.p2HighCardArr[2] && r.p1HighCardArr[1] == r.p2HighCardArr[1] && r.p1HighCardArr[0] > r.p2HighCardArr[0]) {

		less2(r)
	}
}

func compare1_3(r *ranker) {
	if (r.p1HighCardArr[4] > r.p3HighCardArr[4]) ||
		(r.p1HighCardArr[4] == r.p3HighCardArr[4] && r.p1HighCardArr[3] > r.p3HighCardArr[3]) ||
		(r.p1HighCardArr[4] == r.p3HighCardArr[4] && r.p1HighCardArr[3] == r.p3HighCardArr[3] && r.p1HighCardArr[2] > r.p3HighCardArr[2]) ||
		(r.p1HighCardArr[4] == r.p3HighCardArr[4] && r.p1HighCardArr[3] == r.p3HighCardArr[3] && r.p1HighCardArr[2] == r.p3HighCardArr[2] && r.p1HighCardArr[1] > r.p3HighCardArr[1]) ||
		(r.p1HighCardArr[4] == r.p3HighCardArr[4] && r.p1HighCardArr[3] == r.p3HighCardArr[3] && r.p1HighCardArr[2] == r.p3HighCardArr[2] && r.p1HighCardArr[1] == r.p3HighCardArr[1] && r.p1HighCardArr[0] > r.p3HighCardArr[0]) {

		less3(r)
	}
}

func compare1_4(r *ranker) {
	if (r.p1HighCardArr[4] > r.p4HighCardArr[4]) ||
		(r.p1HighCardArr[4] == r.p4HighCardArr[4] && r.p1HighCardArr[3] > r.p4HighCardArr[3]) ||
		(r.p1HighCardArr[4] == r.p4HighCardArr[4] && r.p1HighCardArr[3] == r.p4HighCardArr[3] && r.p1HighCardArr[2] > r.p4HighCardArr[2]) ||
		(r.p1HighCardArr[4] == r.p4HighCardArr[4] && r.p1HighCardArr[3] == r.p4HighCardArr[3] && r.p1HighCardArr[2] == r.p4HighCardArr[2] && r.p1HighCardArr[1] > r.p4HighCardArr[1]) ||
		(r.p1HighCardArr[4] == r.p4HighCardArr[4] && r.p1HighCardArr[3] == r.p4HighCardArr[3] && r.p1HighCardArr[2] == r.p4HighCardArr[2] && r.p1HighCardArr[1] == r.p4HighCardArr[1] && r.p1HighCardArr[0] > r.p4HighCardArr[0]) {

		less4(r)
	}
}

func compare1_5(r *ranker) {
	if (r.p1HighCardArr[4] > r.p5HighCardArr[4]) ||
		(r.p1HighCardArr[4] == r.p5HighCardArr[4] && r.p1HighCardArr[3] > r.p5HighCardArr[3]) ||
		(r.p1HighCardArr[4] == r.p5HighCardArr[4] && r.p1HighCardArr[3] == r.p5HighCardArr[3] && r.p1HighCardArr[2] > r.p5HighCardArr[2]) ||
		(r.p1HighCardArr[4] == r.p5HighCardArr[4] && r.p1HighCardArr[3] == r.p5HighCardArr[3] && r.p1HighCardArr[2] == r.p5HighCardArr[2] && r.p1HighCardArr[1] > r.p5HighCardArr[1]) ||
		(r.p1HighCardArr[4] == r.p5HighCardArr[4] && r.p1HighCardArr[3] == r.p5HighCardArr[3] && r.p1HighCardArr[2] == r.p5HighCardArr[2] && r.p1HighCardArr[1] == r.p5HighCardArr[1] && r.p1HighCardArr[0] > r.p5HighCardArr[0]) {

		less5(r)
	}
}

func compare1_6(r *ranker) {
	if (r.p1HighCardArr[4] > r.p6HighCardArr[4]) ||
		(r.p1HighCardArr[4] == r.p6HighCardArr[4] && r.p1HighCardArr[3] > r.p6HighCardArr[3]) ||
		(r.p1HighCardArr[4] == r.p6HighCardArr[4] && r.p1HighCardArr[3] == r.p6HighCardArr[3] && r.p1HighCardArr[2] > r.p6HighCardArr[2]) ||
		(r.p1HighCardArr[4] == r.p6HighCardArr[4] && r.p1HighCardArr[3] == r.p6HighCardArr[3] && r.p1HighCardArr[2] == r.p6HighCardArr[2] && r.p1HighCardArr[1] > r.p6HighCardArr[1]) ||
		(r.p1HighCardArr[4] == r.p6HighCardArr[4] && r.p1HighCardArr[3] == r.p6HighCardArr[3] && r.p1HighCardArr[2] == r.p6HighCardArr[2] && r.p1HighCardArr[1] == r.p6HighCardArr[1] && r.p1HighCardArr[0] > r.p6HighCardArr[0]) {

		less6(r)
	}
}

func compare2_1(r *ranker) {
	if (r.p2HighCardArr[4] > r.p1HighCardArr[4]) ||
		(r.p2HighCardArr[4] == r.p1HighCardArr[4] && r.p2HighCardArr[3] > r.p1HighCardArr[3]) ||
		(r.p2HighCardArr[4] == r.p1HighCardArr[4] && r.p2HighCardArr[3] == r.p1HighCardArr[3] && r.p2HighCardArr[2] > r.p1HighCardArr[2]) ||
		(r.p2HighCardArr[4] == r.p1HighCardArr[4] && r.p2HighCardArr[3] == r.p1HighCardArr[3] && r.p2HighCardArr[2] == r.p1HighCardArr[2] && r.p2HighCardArr[1] > r.p1HighCardArr[1]) ||
		(r.p2HighCardArr[4] == r.p1HighCardArr[4] && r.p2HighCardArr[3] == r.p1HighCardArr[3] && r.p2HighCardArr[2] == r.p1HighCardArr[2] && r.p2HighCardArr[1] == r.p1HighCardArr[1] && r.p2HighCardArr[0] > r.p1HighCardArr[0]) {

		less1(r)
	}
}

func compare2_3(r *ranker) {
	if (r.p2HighCardArr[4] > r.p3HighCardArr[4]) ||
		(r.p2HighCardArr[4] == r.p3HighCardArr[4] && r.p2HighCardArr[3] > r.p3HighCardArr[3]) ||
		(r.p2HighCardArr[4] == r.p3HighCardArr[4] && r.p2HighCardArr[3] == r.p3HighCardArr[3] && r.p2HighCardArr[2] > r.p3HighCardArr[2]) ||
		(r.p2HighCardArr[4] == r.p3HighCardArr[4] && r.p2HighCardArr[3] == r.p3HighCardArr[3] && r.p2HighCardArr[2] == r.p3HighCardArr[2] && r.p2HighCardArr[1] > r.p3HighCardArr[1]) ||
		(r.p2HighCardArr[4] == r.p3HighCardArr[4] && r.p2HighCardArr[3] == r.p3HighCardArr[3] && r.p2HighCardArr[2] == r.p3HighCardArr[2] && r.p2HighCardArr[1] == r.p3HighCardArr[1] && r.p2HighCardArr[0] > r.p3HighCardArr[0]) {

		less3(r)
	}
}

func compare2_4(r *ranker) {
	if (r.p2HighCardArr[4] > r.p4HighCardArr[4]) ||
		(r.p2HighCardArr[4] == r.p4HighCardArr[4] && r.p2HighCardArr[3] > r.p4HighCardArr[3]) ||
		(r.p2HighCardArr[4] == r.p4HighCardArr[4] && r.p2HighCardArr[3] == r.p4HighCardArr[3] && r.p2HighCardArr[2] > r.p4HighCardArr[2]) ||
		(r.p2HighCardArr[4] == r.p4HighCardArr[4] && r.p2HighCardArr[3] == r.p4HighCardArr[3] && r.p2HighCardArr[2] == r.p4HighCardArr[2] && r.p2HighCardArr[1] > r.p4HighCardArr[1]) ||
		(r.p2HighCardArr[4] == r.p4HighCardArr[4] && r.p2HighCardArr[3] == r.p4HighCardArr[3] && r.p2HighCardArr[2] == r.p4HighCardArr[2] && r.p2HighCardArr[1] == r.p4HighCardArr[1] && r.p2HighCardArr[0] > r.p4HighCardArr[0]) {

		less4(r)
	}
}

func compare2_5(r *ranker) {
	if (r.p2HighCardArr[4] > r.p5HighCardArr[4]) ||
		(r.p2HighCardArr[4] == r.p5HighCardArr[4] && r.p2HighCardArr[3] > r.p5HighCardArr[3]) ||
		(r.p2HighCardArr[4] == r.p5HighCardArr[4] && r.p2HighCardArr[3] == r.p5HighCardArr[3] && r.p2HighCardArr[2] > r.p5HighCardArr[2]) ||
		(r.p2HighCardArr[4] == r.p5HighCardArr[4] && r.p2HighCardArr[3] == r.p5HighCardArr[3] && r.p2HighCardArr[2] == r.p5HighCardArr[2] && r.p2HighCardArr[1] > r.p5HighCardArr[1]) ||
		(r.p2HighCardArr[4] == r.p5HighCardArr[4] && r.p2HighCardArr[3] == r.p5HighCardArr[3] && r.p2HighCardArr[2] == r.p5HighCardArr[2] && r.p2HighCardArr[1] == r.p5HighCardArr[1] && r.p2HighCardArr[0] > r.p5HighCardArr[0]) {

		less5(r)
	}
}

func compare2_6(r *ranker) {
	if (r.p2HighCardArr[4] > r.p6HighCardArr[4]) ||
		(r.p2HighCardArr[4] == r.p6HighCardArr[4] && r.p2HighCardArr[3] > r.p6HighCardArr[3]) ||
		(r.p2HighCardArr[4] == r.p6HighCardArr[4] && r.p2HighCardArr[3] == r.p6HighCardArr[3] && r.p2HighCardArr[2] > r.p6HighCardArr[2]) ||
		(r.p2HighCardArr[4] == r.p6HighCardArr[4] && r.p2HighCardArr[3] == r.p6HighCardArr[3] && r.p2HighCardArr[2] == r.p6HighCardArr[2] && r.p2HighCardArr[1] > r.p6HighCardArr[1]) ||
		(r.p2HighCardArr[4] == r.p6HighCardArr[4] && r.p2HighCardArr[3] == r.p6HighCardArr[3] && r.p2HighCardArr[2] == r.p6HighCardArr[2] && r.p2HighCardArr[1] == r.p6HighCardArr[1] && r.p2HighCardArr[0] > r.p6HighCardArr[0]) {

		less6(r)
	}
}

func compare3_1(r *ranker) {
	if (r.p3HighCardArr[4] > r.p1HighCardArr[4]) ||
		(r.p3HighCardArr[4] == r.p1HighCardArr[4] && r.p3HighCardArr[3] > r.p1HighCardArr[3]) ||
		(r.p3HighCardArr[4] == r.p1HighCardArr[4] && r.p3HighCardArr[3] == r.p1HighCardArr[3] && r.p3HighCardArr[2] > r.p1HighCardArr[2]) ||
		(r.p3HighCardArr[4] == r.p1HighCardArr[4] && r.p3HighCardArr[3] == r.p1HighCardArr[3] && r.p3HighCardArr[2] == r.p1HighCardArr[2] && r.p3HighCardArr[1] > r.p1HighCardArr[1]) ||
		(r.p3HighCardArr[4] == r.p1HighCardArr[4] && r.p3HighCardArr[3] == r.p1HighCardArr[3] && r.p3HighCardArr[2] == r.p1HighCardArr[2] && r.p3HighCardArr[1] == r.p1HighCardArr[1] && r.p3HighCardArr[0] > r.p1HighCardArr[0]) {

		less1(r)
	}
}

func compare3_2(r *ranker) {
	if (r.p3HighCardArr[4] > r.p2HighCardArr[4]) ||
		(r.p3HighCardArr[4] == r.p2HighCardArr[4] && r.p3HighCardArr[3] > r.p2HighCardArr[3]) ||
		(r.p3HighCardArr[4] == r.p2HighCardArr[4] && r.p3HighCardArr[3] == r.p2HighCardArr[3] && r.p3HighCardArr[2] > r.p2HighCardArr[2]) ||
		(r.p3HighCardArr[4] == r.p2HighCardArr[4] && r.p3HighCardArr[3] == r.p2HighCardArr[3] && r.p3HighCardArr[2] == r.p2HighCardArr[2] && r.p3HighCardArr[1] > r.p2HighCardArr[1]) ||
		(r.p3HighCardArr[4] == r.p2HighCardArr[4] && r.p3HighCardArr[3] == r.p2HighCardArr[3] && r.p3HighCardArr[2] == r.p2HighCardArr[2] && r.p3HighCardArr[1] == r.p2HighCardArr[1] && r.p3HighCardArr[0] > r.p2HighCardArr[0]) {

		less2(r)
	}
}

func compare3_4(r *ranker) {
	if (r.p3HighCardArr[4] > r.p4HighCardArr[4]) ||
		(r.p3HighCardArr[4] == r.p4HighCardArr[4] && r.p3HighCardArr[3] > r.p4HighCardArr[3]) ||
		(r.p3HighCardArr[4] == r.p4HighCardArr[4] && r.p3HighCardArr[3] == r.p4HighCardArr[3] && r.p3HighCardArr[2] > r.p4HighCardArr[2]) ||
		(r.p3HighCardArr[4] == r.p4HighCardArr[4] && r.p3HighCardArr[3] == r.p4HighCardArr[3] && r.p3HighCardArr[2] == r.p4HighCardArr[2] && r.p3HighCardArr[1] > r.p4HighCardArr[1]) ||
		(r.p3HighCardArr[4] == r.p4HighCardArr[4] && r.p3HighCardArr[3] == r.p4HighCardArr[3] && r.p3HighCardArr[2] == r.p4HighCardArr[2] && r.p3HighCardArr[1] == r.p4HighCardArr[1] && r.p3HighCardArr[0] > r.p4HighCardArr[0]) {

		less4(r)
	}
}

func compare3_5(r *ranker) {
	if (r.p3HighCardArr[4] > r.p5HighCardArr[4]) ||
		(r.p3HighCardArr[4] == r.p5HighCardArr[4] && r.p3HighCardArr[3] > r.p5HighCardArr[3]) ||
		(r.p3HighCardArr[4] == r.p5HighCardArr[4] && r.p3HighCardArr[3] == r.p5HighCardArr[3] && r.p3HighCardArr[2] > r.p5HighCardArr[2]) ||
		(r.p3HighCardArr[4] == r.p5HighCardArr[4] && r.p3HighCardArr[3] == r.p5HighCardArr[3] && r.p3HighCardArr[2] == r.p5HighCardArr[2] && r.p3HighCardArr[1] > r.p5HighCardArr[1]) ||
		(r.p3HighCardArr[4] == r.p5HighCardArr[4] && r.p3HighCardArr[3] == r.p5HighCardArr[3] && r.p3HighCardArr[2] == r.p5HighCardArr[2] && r.p3HighCardArr[1] == r.p5HighCardArr[1] && r.p3HighCardArr[0] > r.p5HighCardArr[0]) {

		less5(r)
	}
}

func compare3_6(r *ranker) {
	if (r.p3HighCardArr[4] > r.p6HighCardArr[4]) ||
		(r.p3HighCardArr[4] == r.p6HighCardArr[4] && r.p3HighCardArr[3] > r.p6HighCardArr[3]) ||
		(r.p3HighCardArr[4] == r.p6HighCardArr[4] && r.p3HighCardArr[3] == r.p6HighCardArr[3] && r.p3HighCardArr[2] > r.p6HighCardArr[2]) ||
		(r.p3HighCardArr[4] == r.p6HighCardArr[4] && r.p3HighCardArr[3] == r.p6HighCardArr[3] && r.p3HighCardArr[2] == r.p6HighCardArr[2] && r.p3HighCardArr[1] > r.p6HighCardArr[1]) ||
		(r.p3HighCardArr[4] == r.p6HighCardArr[4] && r.p3HighCardArr[3] == r.p6HighCardArr[3] && r.p3HighCardArr[2] == r.p6HighCardArr[2] && r.p3HighCardArr[1] == r.p6HighCardArr[1] && r.p3HighCardArr[0] > r.p6HighCardArr[0]) {

		less6(r)
	}
}

func compare4_1(r *ranker) {
	if (r.p4HighCardArr[4] > r.p1HighCardArr[4]) ||
		(r.p4HighCardArr[4] == r.p1HighCardArr[4] && r.p4HighCardArr[3] > r.p1HighCardArr[3]) ||
		(r.p4HighCardArr[4] == r.p1HighCardArr[4] && r.p4HighCardArr[3] == r.p1HighCardArr[3] && r.p4HighCardArr[2] > r.p1HighCardArr[2]) ||
		(r.p4HighCardArr[4] == r.p1HighCardArr[4] && r.p4HighCardArr[3] == r.p1HighCardArr[3] && r.p4HighCardArr[2] == r.p1HighCardArr[2] && r.p4HighCardArr[1] > r.p1HighCardArr[1]) ||
		(r.p4HighCardArr[4] == r.p1HighCardArr[4] && r.p4HighCardArr[3] == r.p1HighCardArr[3] && r.p4HighCardArr[2] == r.p1HighCardArr[2] && r.p4HighCardArr[1] == r.p1HighCardArr[1] && r.p4HighCardArr[0] > r.p1HighCardArr[0]) {

		less1(r)
	}
}

func compare4_2(r *ranker) {
	if (r.p4HighCardArr[4] > r.p2HighCardArr[4]) ||
		(r.p4HighCardArr[4] == r.p2HighCardArr[4] && r.p4HighCardArr[3] > r.p2HighCardArr[3]) ||
		(r.p4HighCardArr[4] == r.p2HighCardArr[4] && r.p4HighCardArr[3] == r.p2HighCardArr[3] && r.p4HighCardArr[2] > r.p2HighCardArr[2]) ||
		(r.p4HighCardArr[4] == r.p2HighCardArr[4] && r.p4HighCardArr[3] == r.p2HighCardArr[3] && r.p4HighCardArr[2] == r.p2HighCardArr[2] && r.p4HighCardArr[1] > r.p2HighCardArr[1]) ||
		(r.p4HighCardArr[4] == r.p2HighCardArr[4] && r.p4HighCardArr[3] == r.p2HighCardArr[3] && r.p4HighCardArr[2] == r.p2HighCardArr[2] && r.p4HighCardArr[1] == r.p2HighCardArr[1] && r.p4HighCardArr[0] > r.p2HighCardArr[0]) {

		less2(r)
	}
}

func compare4_3(r *ranker) {
	if (r.p4HighCardArr[4] > r.p3HighCardArr[4]) ||
		(r.p4HighCardArr[4] == r.p3HighCardArr[4] && r.p4HighCardArr[3] > r.p3HighCardArr[3]) ||
		(r.p4HighCardArr[4] == r.p3HighCardArr[4] && r.p4HighCardArr[3] == r.p3HighCardArr[3] && r.p4HighCardArr[2] > r.p3HighCardArr[2]) ||
		(r.p4HighCardArr[4] == r.p3HighCardArr[4] && r.p4HighCardArr[3] == r.p3HighCardArr[3] && r.p4HighCardArr[2] == r.p3HighCardArr[2] && r.p4HighCardArr[1] > r.p3HighCardArr[1]) ||
		(r.p4HighCardArr[4] == r.p3HighCardArr[4] && r.p4HighCardArr[3] == r.p3HighCardArr[3] && r.p4HighCardArr[2] == r.p3HighCardArr[2] && r.p4HighCardArr[1] == r.p3HighCardArr[1] && r.p4HighCardArr[0] > r.p3HighCardArr[0]) {

		less3(r)
	}
}

func compare4_5(r *ranker) {
	if (r.p4HighCardArr[4] > r.p5HighCardArr[4]) ||
		(r.p4HighCardArr[4] == r.p5HighCardArr[4] && r.p4HighCardArr[3] > r.p5HighCardArr[3]) ||
		(r.p4HighCardArr[4] == r.p5HighCardArr[4] && r.p4HighCardArr[3] == r.p5HighCardArr[3] && r.p4HighCardArr[2] > r.p5HighCardArr[2]) ||
		(r.p4HighCardArr[4] == r.p5HighCardArr[4] && r.p4HighCardArr[3] == r.p5HighCardArr[3] && r.p4HighCardArr[2] == r.p5HighCardArr[2] && r.p4HighCardArr[1] > r.p5HighCardArr[1]) ||
		(r.p4HighCardArr[4] == r.p5HighCardArr[4] && r.p4HighCardArr[3] == r.p5HighCardArr[3] && r.p4HighCardArr[2] == r.p5HighCardArr[2] && r.p4HighCardArr[1] == r.p5HighCardArr[1] && r.p4HighCardArr[0] > r.p5HighCardArr[0]) {

		less5(r)
	}
}

func compare4_6(r *ranker) {
	if (r.p4HighCardArr[4] > r.p6HighCardArr[4]) ||
		(r.p4HighCardArr[4] == r.p6HighCardArr[4] && r.p4HighCardArr[3] > r.p6HighCardArr[3]) ||
		(r.p4HighCardArr[4] == r.p6HighCardArr[4] && r.p4HighCardArr[3] == r.p6HighCardArr[3] && r.p4HighCardArr[2] > r.p6HighCardArr[2]) ||
		(r.p4HighCardArr[4] == r.p6HighCardArr[4] && r.p4HighCardArr[3] == r.p6HighCardArr[3] && r.p4HighCardArr[2] == r.p6HighCardArr[2] && r.p4HighCardArr[1] > r.p6HighCardArr[1]) ||
		(r.p4HighCardArr[4] == r.p6HighCardArr[4] && r.p4HighCardArr[3] == r.p6HighCardArr[3] && r.p4HighCardArr[2] == r.p6HighCardArr[2] && r.p4HighCardArr[1] == r.p6HighCardArr[1] && r.p4HighCardArr[0] > r.p6HighCardArr[0]) {

		less6(r)
	}
}

func compare5_1(r *ranker) {
	if (r.p5HighCardArr[4] > r.p1HighCardArr[4]) ||
		(r.p5HighCardArr[4] == r.p1HighCardArr[4] && r.p5HighCardArr[3] > r.p1HighCardArr[3]) ||
		(r.p5HighCardArr[4] == r.p1HighCardArr[4] && r.p5HighCardArr[3] == r.p1HighCardArr[3] && r.p5HighCardArr[2] > r.p1HighCardArr[2]) ||
		(r.p5HighCardArr[4] == r.p1HighCardArr[4] && r.p5HighCardArr[3] == r.p1HighCardArr[3] && r.p5HighCardArr[2] == r.p1HighCardArr[2] && r.p5HighCardArr[1] > r.p1HighCardArr[1]) ||
		(r.p5HighCardArr[4] == r.p1HighCardArr[4] && r.p5HighCardArr[3] == r.p1HighCardArr[3] && r.p5HighCardArr[2] == r.p1HighCardArr[2] && r.p5HighCardArr[1] == r.p1HighCardArr[1] && r.p5HighCardArr[0] > r.p1HighCardArr[0]) {

		less1(r)
	}
}

func compare5_2(r *ranker) {
	if (r.p5HighCardArr[4] > r.p2HighCardArr[4]) ||
		(r.p5HighCardArr[4] == r.p2HighCardArr[4] && r.p5HighCardArr[3] > r.p2HighCardArr[3]) ||
		(r.p5HighCardArr[4] == r.p2HighCardArr[4] && r.p5HighCardArr[3] == r.p2HighCardArr[3] && r.p5HighCardArr[2] > r.p2HighCardArr[2]) ||
		(r.p5HighCardArr[4] == r.p2HighCardArr[4] && r.p5HighCardArr[3] == r.p2HighCardArr[3] && r.p5HighCardArr[2] == r.p2HighCardArr[2] && r.p5HighCardArr[1] > r.p2HighCardArr[1]) ||
		(r.p5HighCardArr[4] == r.p2HighCardArr[4] && r.p5HighCardArr[3] == r.p2HighCardArr[3] && r.p5HighCardArr[2] == r.p2HighCardArr[2] && r.p5HighCardArr[1] == r.p2HighCardArr[1] && r.p5HighCardArr[0] > r.p2HighCardArr[0]) {

		less2(r)
	}
}

func compare5_3(r *ranker) {
	if (r.p5HighCardArr[4] > r.p3HighCardArr[4]) ||
		(r.p5HighCardArr[4] == r.p3HighCardArr[4] && r.p5HighCardArr[3] > r.p3HighCardArr[3]) ||
		(r.p5HighCardArr[4] == r.p3HighCardArr[4] && r.p5HighCardArr[3] == r.p3HighCardArr[3] && r.p5HighCardArr[2] > r.p3HighCardArr[2]) ||
		(r.p5HighCardArr[4] == r.p3HighCardArr[4] && r.p5HighCardArr[3] == r.p3HighCardArr[3] && r.p5HighCardArr[2] == r.p3HighCardArr[2] && r.p5HighCardArr[1] > r.p3HighCardArr[1]) ||
		(r.p5HighCardArr[4] == r.p3HighCardArr[4] && r.p5HighCardArr[3] == r.p3HighCardArr[3] && r.p5HighCardArr[2] == r.p3HighCardArr[2] && r.p5HighCardArr[1] == r.p3HighCardArr[1] && r.p5HighCardArr[0] > r.p3HighCardArr[0]) {

		less3(r)
	}
}

func compare5_4(r *ranker) {
	if (r.p5HighCardArr[4] > r.p4HighCardArr[4]) ||
		(r.p5HighCardArr[4] == r.p4HighCardArr[4] && r.p5HighCardArr[3] > r.p4HighCardArr[3]) ||
		(r.p5HighCardArr[4] == r.p4HighCardArr[4] && r.p5HighCardArr[3] == r.p4HighCardArr[3] && r.p5HighCardArr[2] > r.p4HighCardArr[2]) ||
		(r.p5HighCardArr[4] == r.p4HighCardArr[4] && r.p5HighCardArr[3] == r.p4HighCardArr[3] && r.p5HighCardArr[2] == r.p4HighCardArr[2] && r.p5HighCardArr[1] > r.p4HighCardArr[1]) ||
		(r.p5HighCardArr[4] == r.p4HighCardArr[4] && r.p5HighCardArr[3] == r.p4HighCardArr[3] && r.p5HighCardArr[2] == r.p4HighCardArr[2] && r.p5HighCardArr[1] == r.p4HighCardArr[1] && r.p5HighCardArr[0] > r.p4HighCardArr[0]) {

		less4(r)
	}
}

func compare5_6(r *ranker) {
	if (r.p5HighCardArr[4] > r.p6HighCardArr[4]) ||
		(r.p5HighCardArr[4] == r.p6HighCardArr[4] && r.p5HighCardArr[3] > r.p6HighCardArr[3]) ||
		(r.p5HighCardArr[4] == r.p6HighCardArr[4] && r.p5HighCardArr[3] == r.p6HighCardArr[3] && r.p5HighCardArr[2] > r.p6HighCardArr[2]) ||
		(r.p5HighCardArr[4] == r.p6HighCardArr[4] && r.p5HighCardArr[3] == r.p6HighCardArr[3] && r.p5HighCardArr[2] == r.p6HighCardArr[2] && r.p5HighCardArr[1] > r.p6HighCardArr[1]) ||
		(r.p5HighCardArr[4] == r.p6HighCardArr[4] && r.p5HighCardArr[3] == r.p6HighCardArr[3] && r.p5HighCardArr[2] == r.p6HighCardArr[2] && r.p5HighCardArr[1] == r.p6HighCardArr[1] && r.p5HighCardArr[0] > r.p6HighCardArr[0]) {

		less6(r)
	}
}

func compare6_1(r *ranker) {
	if (r.p6HighCardArr[4] > r.p1HighCardArr[4]) ||
		(r.p6HighCardArr[4] == r.p1HighCardArr[4] && r.p6HighCardArr[3] > r.p1HighCardArr[3]) ||
		(r.p6HighCardArr[4] == r.p1HighCardArr[4] && r.p6HighCardArr[3] == r.p1HighCardArr[3] && r.p6HighCardArr[2] > r.p1HighCardArr[2]) ||
		(r.p6HighCardArr[4] == r.p1HighCardArr[4] && r.p6HighCardArr[3] == r.p1HighCardArr[3] && r.p6HighCardArr[2] == r.p1HighCardArr[2] && r.p6HighCardArr[1] > r.p1HighCardArr[1]) ||
		(r.p6HighCardArr[4] == r.p1HighCardArr[4] && r.p6HighCardArr[3] == r.p1HighCardArr[3] && r.p6HighCardArr[2] == r.p1HighCardArr[2] && r.p6HighCardArr[1] == r.p1HighCardArr[1] && r.p6HighCardArr[0] > r.p1HighCardArr[0]) {

		less1(r)
	}
}

func compare6_2(r *ranker) {
	if (r.p6HighCardArr[4] > r.p2HighCardArr[4]) ||
		(r.p6HighCardArr[4] == r.p2HighCardArr[4] && r.p6HighCardArr[3] > r.p2HighCardArr[3]) ||
		(r.p6HighCardArr[4] == r.p2HighCardArr[4] && r.p6HighCardArr[3] == r.p2HighCardArr[3] && r.p6HighCardArr[2] > r.p2HighCardArr[2]) ||
		(r.p6HighCardArr[4] == r.p2HighCardArr[4] && r.p6HighCardArr[3] == r.p2HighCardArr[3] && r.p6HighCardArr[2] == r.p2HighCardArr[2] && r.p6HighCardArr[1] > r.p2HighCardArr[1]) ||
		(r.p6HighCardArr[4] == r.p2HighCardArr[4] && r.p6HighCardArr[3] == r.p2HighCardArr[3] && r.p6HighCardArr[2] == r.p2HighCardArr[2] && r.p6HighCardArr[1] == r.p2HighCardArr[1] && r.p6HighCardArr[0] > r.p2HighCardArr[0]) {

		less2(r)
	}
}

func compare6_3(r *ranker) {
	if (r.p6HighCardArr[4] > r.p3HighCardArr[4]) ||
		(r.p6HighCardArr[4] == r.p3HighCardArr[4] && r.p6HighCardArr[3] > r.p3HighCardArr[3]) ||
		(r.p6HighCardArr[4] == r.p3HighCardArr[4] && r.p6HighCardArr[3] == r.p3HighCardArr[3] && r.p6HighCardArr[2] > r.p3HighCardArr[2]) ||
		(r.p6HighCardArr[4] == r.p3HighCardArr[4] && r.p6HighCardArr[3] == r.p3HighCardArr[3] && r.p6HighCardArr[2] == r.p3HighCardArr[2] && r.p6HighCardArr[1] > r.p3HighCardArr[1]) ||
		(r.p6HighCardArr[4] == r.p3HighCardArr[4] && r.p6HighCardArr[3] == r.p3HighCardArr[3] && r.p6HighCardArr[2] == r.p3HighCardArr[2] && r.p6HighCardArr[1] == r.p3HighCardArr[1] && r.p6HighCardArr[0] > r.p3HighCardArr[0]) {

		less3(r)
	}
}

func compare6_4(r *ranker) {
	if (r.p6HighCardArr[4] > r.p4HighCardArr[4]) ||
		(r.p6HighCardArr[4] == r.p4HighCardArr[4] && r.p6HighCardArr[3] > r.p4HighCardArr[3]) ||
		(r.p6HighCardArr[4] == r.p4HighCardArr[4] && r.p6HighCardArr[3] == r.p4HighCardArr[3] && r.p6HighCardArr[2] > r.p4HighCardArr[2]) ||
		(r.p6HighCardArr[4] == r.p4HighCardArr[4] && r.p6HighCardArr[3] == r.p4HighCardArr[3] && r.p6HighCardArr[2] == r.p4HighCardArr[2] && r.p6HighCardArr[1] > r.p4HighCardArr[1]) ||
		(r.p6HighCardArr[4] == r.p4HighCardArr[4] && r.p6HighCardArr[3] == r.p4HighCardArr[3] && r.p6HighCardArr[2] == r.p4HighCardArr[2] && r.p6HighCardArr[1] == r.p4HighCardArr[1] && r.p6HighCardArr[0] > r.p4HighCardArr[0]) {

		less4(r)
	}
}

func compare6_5(r *ranker) {
	if (r.p6HighCardArr[4] > r.p5HighCardArr[4]) ||
		(r.p6HighCardArr[4] == r.p5HighCardArr[4] && r.p6HighCardArr[3] > r.p5HighCardArr[3]) ||
		(r.p6HighCardArr[4] == r.p5HighCardArr[4] && r.p6HighCardArr[3] == r.p5HighCardArr[3] && r.p6HighCardArr[2] > r.p5HighCardArr[2]) ||
		(r.p6HighCardArr[4] == r.p5HighCardArr[4] && r.p6HighCardArr[3] == r.p5HighCardArr[3] && r.p6HighCardArr[2] == r.p5HighCardArr[2] && r.p6HighCardArr[1] > r.p5HighCardArr[1]) ||
		(r.p6HighCardArr[4] == r.p5HighCardArr[4] && r.p6HighCardArr[3] == r.p5HighCardArr[3] && r.p6HighCardArr[2] == r.p5HighCardArr[2] && r.p6HighCardArr[1] == r.p5HighCardArr[1] && r.p6HighCardArr[0] > r.p5HighCardArr[0]) {

		less5(r)
	}
}

// Strip player values for ranker

func less6(r *ranker) {
	r.p6HighCardArr[0] = 0
	r.p6HighCardArr[1] = 0
	r.p6HighCardArr[2] = 0
	r.p6HighCardArr[3] = 0
	r.p6HighCardArr[4] = 0
	r.p6HighPair = 0
}

func less5(r *ranker) {
	r.p5HighCardArr[0] = 0
	r.p5HighCardArr[1] = 0
	r.p5HighCardArr[2] = 0
	r.p5HighCardArr[3] = 0
	r.p5HighCardArr[4] = 0
	r.p5HighPair = 0
}

func less4(r *ranker) {
	r.p4HighCardArr[0] = 0
	r.p4HighCardArr[1] = 0
	r.p4HighCardArr[2] = 0
	r.p4HighCardArr[3] = 0
	r.p4HighCardArr[4] = 0
	r.p4HighPair = 0
}

func less3(r *ranker) {
	r.p3HighCardArr[0] = 0
	r.p3HighCardArr[1] = 0
	r.p3HighCardArr[2] = 0
	r.p3HighCardArr[3] = 0
	r.p3HighCardArr[4] = 0
	r.p3HighPair = 0
}

func less2(r *ranker) {
	r.p2HighCardArr[0] = 0
	r.p2HighCardArr[1] = 0
	r.p2HighCardArr[2] = 0
	r.p2HighCardArr[3] = 0
	r.p2HighCardArr[4] = 0
	r.p2HighPair = 0
}

func less1(r *ranker) {
	r.p1HighCardArr[0] = 0
	r.p1HighCardArr[1] = 0
	r.p1HighCardArr[2] = 0
	r.p1HighCardArr[3] = 0
	r.p1HighCardArr[4] = 0
	r.p1HighPair = 0
}
