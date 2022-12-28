package rpc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/deroproject/derohe/rpc"
	"github.com/ybbus/jsonrpc/v3"
)

const (
	pre          = "http://"
	suff         = "/json_rpc"
	dReamsSCID   = "ad2e7b37c380cc1aed3a6b27224ddfc92a2d15962ca1f4d35e530dba0f9575a9"
	TourneySCID  = "c2e1ec16aed6f653aef99a06826b2b6f633349807d01fbb74cc0afb5ff99c3c7"
	HolderoSCID  = "e3f37573de94560e126a9020c0a5b3dfc7a4f3a4fbbe369fba93fbd219dc5fe9"
	pHolderoSCID = "896834d57628d3a65076d3f4d84ddc7c5daf3e86b66a47f018abda6068afe2e6"
	BaccSCID     = "8289c6109f41cbe1f6d5f27a419db537bf3bf30a25eff285241a36e1ae3e48a4"
	PredictSCID  = "c89c2f514300413fd6922c28591196a7c48b42b07e7f4d7d8d9f7643e253a6ff"
	pPredictSCID = "e5e49c9a6dc1c0dc8a94429a01bf758e705de49487cbd0b3e3550648d2460cdf"
	SportsSCID   = "ad11377c29a863523c1cc50a33ca13e861cc146a7c0496da58deaa1973e0a39f"
	pSportsSCID  = "fffdc4ea6d157880841feab335ab4755edcde4e60fec2fff661009b16f44fa94"
	TarotSCID    = "a6fc0033327073dd54e448192af929466596fce4d689302e558bc85ea8734a82"
	GnomonSCID   = "a05395bb0cf77adc850928b0db00eb5ca7a9ccbafd9a38d021c8d299ad5ce1a4"
	DevAddress   = "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn"
	ArtAddress   = "dero1qy0khp9s9yw2h0eu20xmy9lth3zp5cacmx3rwt6k45l568d2mmcf6qgcsevzx"
)

var Times times
var Display displayStrings
var CardHash hashValue
var Round holderoValues
var Bacc baccValues
var Signal signals
var Predict predictionValues
var Tarot tarotValues

func fromHextoString(h string) string {
	str, err := hex.DecodeString(h)
	if err != nil {
		log.Println("Hex Conversion Error", err)
		return ""
	}
	return string(str)
}

func SetDaemonClient(addr string) (jsonrpc.RPCClient, context.Context, context.CancelFunc) {
	client := jsonrpc.NewClient(pre + addr + suff)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	return client, ctx, cancel
}

func Ping() error { /// ping blockchain for connection
	rpcClientD, ctx, cancel := SetDaemonClient(Round.Daemon)
	defer cancel()

	var result string
	err := rpcClientD.CallFor(ctx, &result, "DERO.Ping")
	if err != nil {
		Signal.Daemon = false
		return nil
	}

	if result == "Pong " {
		Signal.Daemon = true
	} else {
		Signal.Daemon = false
	}

	return err
}

func DaemonHeight(ep string) (uint64, error) {
	rpcClientD, ctx, cancel := SetDaemonClient(ep)
	defer cancel()

	var result *rpc.GetHeight_Result
	err := rpcClientD.CallFor(ctx, &result, "DERO.GetHeight")
	if err != nil {
		log.Println(err)
		return 0, nil
	}

	return result.Height, err
}

func GasEstimate(scid string, args rpc.Arguments, t []rpc.Transfer) (uint64, error) {
	rpcClientD, ctx, cancel := SetDaemonClient(Round.Daemon)
	defer cancel()

	var result *rpc.GasEstimate_Result

	arg1 := rpc.Argument{Name: "SC_ACTION", DataType: "U", Value: 0}
	arg2 := rpc.Argument{Name: "SC_ID", DataType: "H", Value: scid}
	args = append(args, arg1)
	args = append(args, arg2)
	params := rpc.GasEstimate_Params{
		Transfers: t,
		SC_Value:  0,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Signer:    Wallet.Address,
	}

	err := rpcClientD.CallFor(ctx, &result, "DERO.GetGasEstimate", params)
	if err != nil {
		log.Println(err)
		return 0, nil
	}

	log.Println("Gas Fee:", result.GasStorage+120)

	if result.GasStorage < 1200 {
		return result.GasStorage + 120, err
	}

	return 1320, err
}

func CheckForIndex(scid string) (interface{}, error) {
	rpcClientD, ctx, cancel := SetDaemonClient(Round.Daemon)
	defer cancel()

	var result *rpc.GetSC_Result
	params := rpc.GetSC_Params{
		SCID:      GnomonSCID,
		Code:      false,
		Variables: true,
	}

	err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
	if err != nil {
		log.Println(err)
		return nil, nil
	}

	owner := result.VariableStringKeys[scid+"owner"]
	address := DeroAddress(owner)

	return address, err
}

func GetGnomonCode(dc bool, pub int) (string, error) {
	if dc {
		rpcClientD, ctx, cancel := SetDaemonClient(Round.Daemon)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      GnomonSCID,
			Code:      true,
			Variables: false,
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println(err)
			return "", nil
		}

		return result.Code, err
	}
	return "", nil
}

func CheckHolderoContract() error {
	rpcClientD, ctx, cancel := SetDaemonClient(Round.Daemon)
	defer cancel()

	var result *rpc.GetSC_Result
	params := rpc.GetSC_Params{
		SCID:      Round.Contract,
		Code:      false,
		Variables: true,
	}

	err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
	if err != nil {
		log.Println(err)
		return nil
	}

	cards := fmt.Sprint(result.VariableStringKeys["Deck Count:"])
	version := fmt.Sprint(result.VariableStringKeys["V:"])

	v, _ := strconv.Atoi(version)
	c, _ := strconv.Atoi(cards)

	if c > 0 && v >= 100 {
		Signal.Contract = true
	} else {
		Signal.Contract = false
	}

	return err
}

func CheckTournamentTable() (bool, error) {
	rpcClientD, ctx, cancel := SetDaemonClient(Round.Daemon)
	defer cancel()

	var result *rpc.GetSC_Result
	params := rpc.GetSC_Params{
		SCID:      Round.Contract,
		Code:      false,
		Variables: true,
	}

	err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
	if err != nil {
		log.Println(err)
		return false, nil
	}

	tourney := fmt.Sprint(result.VariableStringKeys["Tournament"])
	version := fmt.Sprint(result.VariableStringKeys["V:"])

	t, _ := strconv.Atoi(tourney)
	v, _ := strconv.Atoi(version)

	if t == 1 && v >= 110 {
		return true, err
	}

	return false, err
}

func FetchHolderoSC(dc, cc bool) error {
	if dc && cc {
		rpcClientD, ctx, cancel := SetDaemonClient(Round.Daemon)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      Round.Contract,
			Code:      false,
			Variables: true,
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println(err)
			return nil
		}

		var Pot_jv uint64
		Seats_jv := result.VariableStringKeys["Seats at Table:"]
		// Count_jv := result.VariableStringKeys["Counter:"]
		Ante_jv := result.VariableStringKeys["Ante:"]
		BigBlind_jv := result.VariableStringKeys["BB:"]
		SmallBlind_jv := result.VariableStringKeys["SB:"]
		Turn_jv := result.VariableStringKeys["Player:"]
		OneId_jv := result.VariableStringKeys["Player1 ID:"]
		TwoId_jv := result.VariableStringKeys["Player2 ID:"]
		ThreeId_jv := result.VariableStringKeys["Player3 ID:"]
		FourId_jv := result.VariableStringKeys["Player4 ID:"]
		FiveId_jv := result.VariableStringKeys["Player5 ID:"]
		SixId_jv := result.VariableStringKeys["Player6 ID:"]
		Dealer_jv := result.VariableStringKeys["Dealer:"]
		Wager_jv := result.VariableStringKeys["Wager:"]
		Raised_jv := result.VariableStringKeys["Raised:"]
		FlopBool_jv := result.VariableStringKeys["Flop"]
		// TurnBool_jv = result.VariableStringKeys["Turn"]
		// RiverBool_jv = result.VariableStringKeys["River"]
		RevealBool_jv := result.VariableStringKeys["Reveal"]
		// Bet_jv := result.VariableStringKeys["Bet"]
		Full_jv := result.VariableStringKeys["Full"]
		// Open_jv := result.VariableStringKeys["Open"]
		Seed_jv := result.VariableStringKeys["HandSeed"]
		Face_jv := result.VariableStringKeys["Face:"]
		Back_jv := result.VariableStringKeys["Back:"]
		V_jv := result.VariableStringKeys["V:"]

		if V_jv != nil {
			Round.Version = int(V_jv.(float64))
		}

		if Seats_jv != nil && int(Seats_jv.(float64)) > 0 {
			Flop1_jv := result.VariableStringKeys["FlopCard1"]
			Flop2_jv := result.VariableStringKeys["FlopCard2"]
			Flop3_jv := result.VariableStringKeys["FlopCard3"]
			TurnCard_jv := result.VariableStringKeys["TurnCard"]
			RiverCard_jv := result.VariableStringKeys["RiverCard"]
			P1C1_jv := result.VariableStringKeys["Player1card1"]
			P1C2_jv := result.VariableStringKeys["Player1card2"]
			P2C1_jv := result.VariableStringKeys["Player2card1"]
			P2C2_jv := result.VariableStringKeys["Player2card2"]
			P3C1_jv := result.VariableStringKeys["Player3card1"]
			P3C2_jv := result.VariableStringKeys["Player3card2"]
			P4C1_jv := result.VariableStringKeys["Player4card1"]
			P4C2_jv := result.VariableStringKeys["Player4card2"]
			P5C1_jv := result.VariableStringKeys["Player5card1"]
			P5C2_jv := result.VariableStringKeys["Player5card2"]
			P6C1_jv := result.VariableStringKeys["Player6card1"]
			P6C2_jv := result.VariableStringKeys["Player6card2"]
			P1F_jv := result.VariableStringKeys["0F"]
			P2F_jv := result.VariableStringKeys["1F"]
			P3F_jv := result.VariableStringKeys["2F"]
			P4F_jv := result.VariableStringKeys["3F"]
			P5F_jv := result.VariableStringKeys["4F"]
			P6F_jv := result.VariableStringKeys["5F"]
			P1Out_jv := result.VariableStringKeys["0SO"]
			// P2Out_jv = result.VariableStringKeys["1SO"]
			// P3Out_jv = result.VariableStringKeys["2SO"]
			// P4Out_jv = result.VariableStringKeys["3SO"]
			// P5Out_jv = result.VariableStringKeys["4SO"]
			// P6Out_jv = result.VariableStringKeys["5SO"]
			Key1_jv := result.VariableStringKeys["Player1Key"]
			Key2_jv := result.VariableStringKeys["Player2Key"]
			Key3_jv := result.VariableStringKeys["Player3Key"]
			Key4_jv := result.VariableStringKeys["Player4Key"]
			Key5_jv := result.VariableStringKeys["Player5Key"]
			Key6_jv := result.VariableStringKeys["Player6Key"]
			End_jv := result.VariableStringKeys["End"]
			Chips_jv := result.VariableStringKeys["Chips"]
			Tourney_jv := result.VariableStringKeys["Tournament"]

			if Tourney_jv == nil {
				Round.Tourney = false
				if Chips_jv != nil {
					if fromHextoString(Chips_jv.(string)) == "ASSET" {
						Round.Asset = true
						Pot_jv = result.Balances[dReamsSCID]
					} else {
						Round.Asset = false
						Pot_jv = result.Balances["0000000000000000000000000000000000000000000000000000000000000000"]
					}
				} else {
					Pot_jv = result.Balances["0000000000000000000000000000000000000000000000000000000000000000"]
				}
			} else {
				Round.Tourney = true
				if Chips_jv != nil {
					if fromHextoString(Chips_jv.(string)) == "ASSET" {
						Round.Asset = true
						Pot_jv = result.Balances[TourneySCID]
					} else {
						Round.Asset = false
						Pot_jv = result.Balances["0000000000000000000000000000000000000000000000000000000000000000"]
					}
				} else {
					Pot_jv = result.Balances["0000000000000000000000000000000000000000000000000000000000000000"]
				}
			}

			Round.Ante = uint64(Ante_jv.(float64))
			Round.BB = uint64(BigBlind_jv.(float64))
			Round.SB = uint64(SmallBlind_jv.(float64))
			Round.Pot = Pot_jv

			hasFolded(P1F_jv, P2F_jv, P3F_jv, P4F_jv, P5F_jv, P6F_jv)
			allFolded(P1F_jv, P2F_jv, P3F_jv, P4F_jv, P5F_jv, P6F_jv, Seats_jv)

			if !Round.LocalEnd {
				getCommCardValues(Flop1_jv, Flop2_jv, Flop3_jv, TurnCard_jv, RiverCard_jv)
				getPlayerCardValues(P1C1_jv, P1C2_jv, P2C1_jv, P2C2_jv, P3C1_jv, P3C2_jv, P4C1_jv, P4C2_jv, P5C1_jv, P5C2_jv, P6C1_jv, P6C2_jv)
			}

			if !Signal.Startup {
				setHolderoName(OneId_jv, TwoId_jv, ThreeId_jv, FourId_jv, FiveId_jv, SixId_jv)
				setSignals(Pot_jv, P1Out_jv)
			}

			if P1Out_jv != nil {
				Signal.Out1 = true
			} else {
				Signal.Out1 = false
			}

			tableOpen(Seats_jv, Full_jv, TwoId_jv, ThreeId_jv, FourId_jv, FiveId_jv, SixId_jv)

			if FlopBool_jv != nil {
				Round.Flop = true
			} else {
				Round.Flop = false
			}

			Display.PlayerId = checkPlayerId(getAvatar(1, OneId_jv), getAvatar(2, TwoId_jv), getAvatar(3, ThreeId_jv), getAvatar(4, FourId_jv), getAvatar(5, FiveId_jv), getAvatar(6, SixId_jv))

			if Wager_jv != nil {
				if Round.Bettor == "" {
					Round.Bettor = findBettor(Turn_jv)
				}
				Round.Wager = uint64(Wager_jv.(float64))
				Display.B_Button = "Call/Raise"
				Display.C_Button = "Fold"
			} else {
				Round.Bettor = ""
				Round.Wager = 0
			}

			if Raised_jv != nil {
				if Round.Raisor == "" {
					Round.Raisor = findBettor(Turn_jv)
				}
				Round.Raised = uint64(Raised_jv.(float64))
				Display.B_Button = "Call"
				Display.C_Button = "Fold"
			} else {
				Round.Raisor = ""
				Round.Raised = 0
			}

			if Round.ID == int(Turn_jv.(float64))+1 {
				Signal.My_turn = true
			} else if Round.ID == 1 && Turn_jv == Seats_jv {
				Signal.My_turn = true
			} else {
				Signal.My_turn = false
			}

			Display.Pot = fromAtomic(Pot_jv)
			Display.Seats = fmt.Sprint(Seats_jv)
			Display.Ante = fromAtomic(Ante_jv)
			Display.Blinds = blindString(BigBlind_jv, SmallBlind_jv)
			Display.Dealer = fmt.Sprint(Dealer_jv.(float64) + 1)

			Round.SC_seed = fmt.Sprint(Seed_jv)

			if Face_jv != nil && Face_jv.(string) != "nil" {
				var c = &CardSpecs{}
				json.Unmarshal([]byte(fromHextoString(Face_jv.(string))), c)
				Round.Face = c.Faces.Name
				Round.Back = c.Backs.Name
				Round.F_url = c.Faces.Url
				Round.B_url = c.Backs.Url
			} else {
				Round.Face = ""
				Round.Back = ""
				Round.F_url = ""
				Round.B_url = ""
			}

			if Back_jv != nil && Back_jv.(string) != "nil" {
				var a = &TableSpecs{}
				json.Unmarshal([]byte(fromHextoString(Back_jv.(string))), a)
			}

			if RevealBool_jv != nil && !Signal.Reveal && !Round.LocalEnd {
				if addOne(Turn_jv) == Display.PlayerId {
					RevealKey(Wallet.ClientKey)
					Signal.Reveal = true
				}
			}

			if Turn_jv != Seats_jv {
				Display.Turn = addOne(Turn_jv)
				Display.Readout = turnReadout(Turn_jv)
			} else {
				Display.Turn = "1"
				Display.Readout = turnReadout(float64(0))
			}

			if End_jv != nil {
				CardHash.Key1 = fmt.Sprint(Key1_jv)
				CardHash.Key2 = fmt.Sprint(Key2_jv)
				CardHash.Key3 = fmt.Sprint(Key3_jv)
				CardHash.Key4 = fmt.Sprint(Key4_jv)
				CardHash.Key5 = fmt.Sprint(Key5_jv)
				CardHash.Key6 = fmt.Sprint(Key6_jv)
				Signal.End = true

			}

			if Round.Version >= 110 && Round.ID == 1 && Times.Kick > 0 && !Signal.My_turn && Round.Pot > 0 {
				Last_jv := result.VariableStringKeys["Last"]
				if Last_jv != nil {
					now := time.Now().Unix()
					if int(now) > int(Last_jv.(float64))+Times.Kick+12 {
						if StringToInt(Wallet.Height) > Times.Kick_block+2 {
							TimeOut()
							Times.Kick_block = StringToInt(Wallet.Height)
						}
					}
				}
			}

			winningHand(End_jv)
		} else {
			closedTable()
		}

		potIsEmpty(Pot_jv)
		allFoldedWinner()

		return err
	}

	return nil
}

func GetHoldero100Code(dc bool) (string, error) { /// v 1.0.0
	if dc {
		rpcClientD, ctx, cancel := SetDaemonClient(Round.Daemon)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      "95e69b382044ddc1467e030a80905cf637729612f65624e8d17bf778d4362b8d",
			Code:      true,
			Variables: false,
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println(err)
			return "", nil
		}

		return result.Code, err

	}

	return "", nil
}

func GetHoldero110Code(dc bool, pub int) (string, error) { /// v 1.1.0
	if dc {
		rpcClientD, ctx, cancel := SetDaemonClient(Round.Daemon)
		defer cancel()

		var result *rpc.GetSC_Result
		var params rpc.GetSC_Params
		if pub == 1 {
			params = rpc.GetSC_Params{
				SCID:      pHolderoSCID,
				Code:      true,
				Variables: false,
			}
		} else {
			params = rpc.GetSC_Params{
				SCID:      HolderoSCID,
				Code:      true,
				Variables: false,
			}
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println(err)
			return "", nil
		}

		return result.Code, err

	}

	return "", nil
}

func FetchBaccSC(dc bool) error {
	if dc {
		rpcClientD, ctx, cancel := SetDaemonClient(Round.Daemon)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      BaccSCID,
			Code:      false,
			Variables: true,
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println(err)
			return nil
		}

		Total_jv := result.VariableStringKeys["TotalHandsPlayed:"]
		Player_jv := result.VariableStringKeys["Player Wins:"]
		Banker_jv := result.VariableStringKeys["Banker Wins:"]
		Min_jv := result.VariableStringKeys["Min Bet:"]
		Max_jv := result.VariableStringKeys["Max Bet:"]
		Ties_jv := result.VariableStringKeys["Ties:"]
		// Pot_jv = result.Balances[dReamsSCID]
		// Pot_jv = result.Balances["0000000000000000000000000000000000000000000000000000000000000000"]
		if Total_jv != nil {
			Display.Total_w = fmt.Sprint(Total_jv)
		}

		if Player_jv != nil {
			Display.Player_w = fmt.Sprint(Player_jv)
		}

		if Banker_jv != nil {
			Display.Banker_w = fmt.Sprint(Banker_jv)
		}

		if Ties_jv != nil {
			Display.Ties = fmt.Sprint(Ties_jv)
		}

		if Max_jv != nil {
			Display.BaccMax = fmt.Sprintf("%.0f", Max_jv.(float64)/100000)
		}

		if Min_jv != nil {
			Display.BaccMin = fmt.Sprintf("%.0f", Min_jv.(float64)/100000)
		}

		return err
	}

	return nil
}

func GetBaccCode(dc bool) (string, error) {
	if dc {
		rpcClientD, ctx, cancel := SetDaemonClient(Round.Daemon)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      BaccSCID,
			Code:      true,
			Variables: false,
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println(err)
			return "", nil
		}

		return result.Code, err
	}

	return "", nil
}

func FetchBaccHand(dc bool, tx string) error { /// find played hand
	if dc && tx != "" {
		rpcClientD, ctx, cancel := SetDaemonClient(Round.Daemon)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      BaccSCID,
			Code:      false,
			Variables: true,
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println(err)
			return nil
		}
		Total_jv := result.VariableStringKeys["TotalHandsPlayed:"]
		if Total_jv != nil {
			start := int(Total_jv.(float64))
			i := int(Total_jv.(float64))
			for i < start+24 {
				h := "-Hand#TXID:"
				w := strconv.Itoa(i)
				TXID_jv := result.VariableStringKeys[w+h]

				if TXID_jv != nil {
					if TXID_jv.(string) == tx {
						Bacc.Found = true
						Bacc.P_card1 = int(result.VariableStringKeys[w+"-Player x:"].(float64))
						Bacc.P_card2 = int(result.VariableStringKeys[w+"-Player y:"].(float64))
						Bacc.P_card3 = int(result.VariableStringKeys[w+"-Player z:"].(float64))
						Bacc.B_card1 = int(result.VariableStringKeys[w+"-Banker x:"].(float64))
						Bacc.B_card2 = int(result.VariableStringKeys[w+"-Banker y:"].(float64))
						Bacc.B_card3 = int(result.VariableStringKeys[w+"-Banker z:"].(float64))
						PTotal_jv := result.VariableStringKeys[w+"-Player total:"]
						BTotal_jv := result.VariableStringKeys[w+"-Banker total:"]

						p := int(PTotal_jv.(float64))
						b := int(BTotal_jv.(float64))
						if PTotal_jv.(float64) == BTotal_jv.(float64) {
							Display.BaccRes = "Tie, " + strconv.Itoa(p) + " & " + strconv.Itoa(b)
						} else if PTotal_jv.(float64) > BTotal_jv.(float64) {
							Display.BaccRes = "Player Wins, " + strconv.Itoa(p) + " over " + strconv.Itoa(b)
						} else {
							Display.BaccRes = "Banker Wins, " + strconv.Itoa(b) + " over " + strconv.Itoa(p)
						}
					}
				}
				i++
			}
		}

		return err
	}

	return nil
}

func ValidBetContract(scid string) (bool, error) {
	rpcClientD, ctx, cancel := SetDaemonClient(Round.Daemon)
	defer cancel()

	var result *rpc.GetSC_Result
	params := rpc.GetSC_Params{
		SCID:      scid,
		Code:      false,
		Variables: true,
	}

	err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
	if err != nil {
		log.Println(err)
		return false, nil
	}

	d := fmt.Sprint(result.VariableStringKeys["dev"])

	if DeroAddress(d) != DevAddress {
		return false, err
	}

	return true, err
}

func FetchPredictionFinal(d bool, scid string) (string, error) {
	if d {
		rpcClientD, ctx, cancel := SetDaemonClient(Round.Daemon)
		defer cancel()

		params := &rpc.GetSC_Params{
			SCID:      scid,
			Code:      false,
			Variables: true,
		}

		var result *rpc.GetSC_Result
		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println(err)
			return "", nil
		}

		p_txid := result.VariableStringKeys["p_final_txid"]

		var txid string
		if p_txid != nil {
			txid = fmt.Sprint(p_txid)

		}
		return txid, nil
	}

	return "", nil
}

func GetPredictCode(dc bool, pub int) (string, error) {
	if dc {
		rpcClientD, ctx, cancel := SetDaemonClient(Round.Daemon)
		defer cancel()

		var result *rpc.GetSC_Result
		var params rpc.GetSC_Params
		if pub == 1 {
			params = rpc.GetSC_Params{
				SCID:      pPredictSCID,
				Code:      true,
				Variables: false,
			}
		} else {
			params = rpc.GetSC_Params{
				SCID:      PredictSCID,
				Code:      true,
				Variables: false,
			}
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println(err)
			return "", nil
		}

		return result.Code, err
	}
	return "", nil
}

func GetSportsCode(dc bool, pub int) (string, error) {
	if dc {
		rpcClientD, ctx, cancel := SetDaemonClient(Round.Daemon)
		defer cancel()

		var result *rpc.GetSC_Result
		var params rpc.GetSC_Params
		if pub == 1 {
			params = rpc.GetSC_Params{
				SCID:      pSportsSCID,
				Code:      true,
				Variables: false,
			}
		} else {
			params = rpc.GetSC_Params{
				SCID:      SportsSCID,
				Code:      true,
				Variables: false,
			}
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println(err)
			return "", nil
		}

		return result.Code, err
	}
	return "", nil
}

func FetchTarotSC(dc bool) error {
	if dc {
		rpcClientD, ctx, cancel := SetDaemonClient(Round.Daemon)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      TarotSCID,
			Code:      false,
			Variables: true,
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println(err)
			return nil
		}

		Reading_jv := result.VariableStringKeys["readings:"]
		if Reading_jv != nil {
			Display.Readings = fmt.Sprint(Reading_jv)
		}

		return err
	}

	return nil
}

func FetchTarotReading(dc bool, tx string) error {
	if dc && len(tx) == 64 {
		rpcClientD, ctx, cancel := SetDaemonClient(Round.Daemon)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      TarotSCID,
			Code:      false,
			Variables: true,
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println(err)
			return nil
		}

		Reading_jv := result.VariableStringKeys["readings:"]
		if Reading_jv != nil {
			start := int(Reading_jv.(float64))
			i := int(Reading_jv.(float64))
			for i < start+24 {
				h := "-readingTXID:"
				w := strconv.Itoa(i)
				TXID_jv := result.VariableStringKeys[w+h]

				if TXID_jv != nil {
					if TXID_jv.(string) == tx {
						Tarot.Found = true
						Tarot.T_card1 = findTarotCard(result.VariableStringKeys[w+"-card1:"])
						Tarot.T_card2 = findTarotCard(result.VariableStringKeys[w+"-card2:"])
						Tarot.T_card3 = findTarotCard(result.VariableStringKeys[w+"-card3:"])
					}
				}
				i++
			}
		}

		return err
	}

	return nil
}

func GetDifficulty(ep string) (float64, error) {
	rpcClientD, ctx, cancel := SetDaemonClient(ep)
	defer cancel()

	var result *rpc.GetInfo_Result
	err := rpcClientD.CallFor(ctx, &result, "DERO.GetInfo")
	if err != nil {
		log.Println(err)
		return 0, nil
	}

	return float64(result.Difficulty), err
}

func GetBlockTime(ep string) (float64, error) {
	rpcClientD, ctx, cancel := SetDaemonClient(ep)
	defer cancel()

	var result *rpc.GetInfo_Result
	err := rpcClientD.CallFor(ctx, &result, "DERO.GetInfo")
	if err != nil {
		log.Println(err)
		return 0, nil
	}

	return float64(result.AverageBlockTime50), err
}
