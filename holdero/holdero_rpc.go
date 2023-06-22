package holdero

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/SixofClubsss/dReams/rpc"
	"github.com/deroproject/derohe/cryptography/crypto"
	dero "github.com/deroproject/derohe/rpc"
)

const (
	TourneySCID    = "c2e1ec16aed6f653aef99a06826b2b6f633349807d01fbb74cc0afb5ff99c3c7"
	Holdero100     = "95e69b382044ddc1467e030a80905cf637729612f65624e8d17bf778d4362b8d"
	HolderoSCID    = "e3f37573de94560e126a9020c0a5b3dfc7a4f3a4fbbe369fba93fbd219dc5fe9"
	pHolderoSCID   = "896834d57628d3a65076d3f4d84ddc7c5daf3e86b66a47f018abda6068afe2e6"
	HGCHolderoSCID = "efe646c48977fd776fee73cdd3df147a2668d3b7d965cdb7a187dda4d23005d8"
)

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
}

type holderoValues struct {
	Version   int
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
	Cards     struct {
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
	Winning_hand  []int
	First_try     bool
	Card_delay    bool
	Local_trigger bool
	Flop_trigger  bool
	Turn_trigger  bool
	River_trigger bool
}

var Display displayStrings
var Round holderoValues

// Get Holdero SC data
func FetchHolderoSC() {
	if rpc.Daemon.IsConnected() && Signal.Contract {
		rpcClientD, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		params := dero.GetSC_Params{
			SCID:      Round.Contract,
			Code:      false,
			Variables: true,
		}

		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[FetchHolderoSC]", err)
			return
		}

		var Pot_jv uint64
		Seats_jv := result.VariableStringKeys["Seats at Table:"]
		V_jv := result.VariableStringKeys["V:"]

		if V_jv != nil {
			Round.Version = rpc.IntType(V_jv)
		}

		if Seats_jv != nil && rpc.IntType(Seats_jv) > 0 {
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
			//Back_jv := result.VariableStringKeys["Back:"]
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
			Last_jv := result.VariableStringKeys["Last"]

			if Last_jv != nil {
				Round.Last = int64(rpc.Float64Type(Last_jv))
			} else {
				Round.Last = 0
			}

			if Tourney_jv == nil {
				Round.Tourney = false
				if Chips_jv != nil {
					if rpc.HexToString(Chips_jv) == "ASSET" {
						Round.Asset = true
						if _, ok := result.VariableStringKeys["dReams"].(string); ok {
							Pot_jv = result.Balances[rpc.DreamsSCID]
							Round.AssetID = rpc.DreamsSCID
						} else if _, ok = result.VariableStringKeys["HGC"].(string); ok {
							Pot_jv = result.Balances[rpc.HgcSCID]
							Round.AssetID = rpc.HgcSCID
						}
					} else {
						Round.Asset = false
						Round.AssetID = ""
						Pot_jv = result.Balances["0000000000000000000000000000000000000000000000000000000000000000"]
					}
				} else {
					Round.Asset = false
					Round.AssetID = ""
					Pot_jv = result.Balances["0000000000000000000000000000000000000000000000000000000000000000"]
				}
			} else {
				Round.Tourney = true
				if Chips_jv != nil {
					if rpc.HexToString(Chips_jv) == "ASSET" {
						Round.Asset = true
						Pot_jv = result.Balances[TourneySCID]
					} else {
						Round.Asset = false
						Round.AssetID = ""
						Pot_jv = result.Balances["0000000000000000000000000000000000000000000000000000000000000000"]
					}
				} else {
					Round.Asset = false
					Round.AssetID = ""
					Pot_jv = result.Balances["0000000000000000000000000000000000000000000000000000000000000000"]
				}
			}

			Round.Ante = rpc.Uint64Type(Ante_jv)
			Round.BB = rpc.Uint64Type(BigBlind_jv)
			Round.SB = rpc.Uint64Type(SmallBlind_jv)
			Round.Pot = Pot_jv

			hasFolded(P1F_jv, P2F_jv, P3F_jv, P4F_jv, P5F_jv, P6F_jv)
			allFolded(P1F_jv, P2F_jv, P3F_jv, P4F_jv, P5F_jv, P6F_jv, Seats_jv)

			if !Round.LocalEnd {
				getCommCardValues(Flop1_jv, Flop2_jv, Flop3_jv, TurnCard_jv, RiverCard_jv)
				getPlayerCardValues(P1C1_jv, P1C2_jv, P2C1_jv, P2C2_jv, P3C1_jv, P3C2_jv, P4C1_jv, P4C2_jv, P5C1_jv, P5C2_jv, P6C1_jv, P6C2_jv)
			}

			if !rpc.Startup {
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
				rpc.Wallet.KeyLock = false
			} else {
				Round.Flop = false
			}

			Display.PlayerId = checkPlayerId(getAvatar(1, OneId_jv), getAvatar(2, TwoId_jv), getAvatar(3, ThreeId_jv), getAvatar(4, FourId_jv), getAvatar(5, FiveId_jv), getAvatar(6, SixId_jv))

			if Wager_jv != nil {
				if Round.Bettor == "" {
					Round.Bettor = findBettor(Turn_jv)
				}
				Round.Wager = rpc.Uint64Type(Wager_jv)
				Display.B_Button = "Call/Raise"
				Display.C_Button = "Fold"
			} else {
				Round.Bettor = ""
				Round.Wager = 0
			}

			if Raised_jv != nil {
				if Round.Raiser == "" {
					Round.Raiser = findBettor(Turn_jv)
				}
				Round.Raised = rpc.Uint64Type(Raised_jv)
				Display.B_Button = "Call"
				Display.C_Button = "Fold"
			} else {
				Round.Raiser = ""
				Round.Raised = 0
			}

			if Round.ID == rpc.IntType(Turn_jv)+1 {
				Signal.My_turn = true
			} else if Round.ID == 1 && Turn_jv == Seats_jv {
				Signal.My_turn = true
			} else {
				Signal.My_turn = false
			}

			Display.Pot = rpc.FromAtomic(Pot_jv, 5)
			Display.Seats = fmt.Sprint(Seats_jv)
			Display.Ante = rpc.FromAtomic(Ante_jv, 5)
			Display.Blinds = blindString(BigBlind_jv, SmallBlind_jv)
			Display.Dealer = rpc.AddOne(Dealer_jv)

			Round.SC_seed = fmt.Sprint(Seed_jv)

			if face, ok := Face_jv.(string); ok {
				if face != "nil" {
					var c = &CardSpecs{}
					if err := json.Unmarshal([]byte(rpc.HexToString(face)), c); err == nil {
						Round.Face = c.Faces.Name
						Round.Back = c.Backs.Name
						Round.F_url = c.Faces.Url
						Round.B_url = c.Backs.Url
					}
				}
			} else {
				Round.Face = ""
				Round.Back = ""
				Round.F_url = ""
				Round.B_url = ""
			}

			// // Unused at moment
			// if back, ok := Back_jv.(string); ok {
			// 	if back != "nil" {
			// 		var a = &TableSpecs{}
			// 		json.Unmarshal([]byte(FromHexToString(back)), a)
			// 	}
			// }

			if RevealBool_jv != nil && !Signal.Reveal && !Round.LocalEnd {
				if rpc.AddOne(Turn_jv) == Display.PlayerId {
					Signal.Clicked = true
					Signal.CHeight = rpc.Wallet.Height
					Signal.Reveal = true
					go RevealKey(rpc.Wallet.ClientKey)
				}
			}

			if Turn_jv != Seats_jv {
				Display.Readout = turnReadout(Turn_jv)
				if turn, ok := Turn_jv.(float64); ok {
					Round.Turn = int(turn) + 1
				}
			} else {
				Round.Turn = 1
				Display.Readout = turnReadout(float64(0))
			}

			if End_jv != nil {
				Round.Cards.Key1 = fmt.Sprint(Key1_jv)
				Round.Cards.Key2 = fmt.Sprint(Key2_jv)
				Round.Cards.Key3 = fmt.Sprint(Key3_jv)
				Round.Cards.Key4 = fmt.Sprint(Key4_jv)
				Round.Cards.Key5 = fmt.Sprint(Key5_jv)
				Round.Cards.Key6 = fmt.Sprint(Key6_jv)
				Signal.End = true

			}

			if Round.Version >= 110 && Round.ID == 1 && Times.Kick > 0 && !Signal.My_turn && Round.Pot > 0 && !Round.LocalEnd && !Signal.End {
				if Round.Last != 0 {
					now := time.Now().Unix()
					if now > Round.Last+int64(Times.Kick)+18 {
						if rpc.Wallet.Height > Times.Kick_block+3 {
							TimeOut()
							Times.Kick_block = rpc.Wallet.Height
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
	}
}

// Submit playerId, name, avatar and sit at Holdero table
//   - name and av are for name and avatar in player id string
func SitDown(name, av string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	var player playerId
	player.Id = rpc.Wallet.IdHash
	player.Name = name
	player.Avatar = av

	mar, _ := json.Marshal(player)
	hx := hex.EncodeToString(mar)

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "PlayerEntry"}
	arg2 := dero.Argument{Name: "address", DataType: "S", Value: hx}
	args := dero.Arguments{arg1, arg2}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(Round.Contract, "[Holdero]", args, t, rpc.HighLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[SitDown]", err)
		return
	}

	log.Println("[Holdero] Sit Down TX:", txid)
	rpc.AddLog("Sit Down TX: " + txid.TXID)
}

// Leave Holdero seat on players turn
func Leave() {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	checkoutId := rpc.StringToInt(Display.PlayerId)
	singleNameClear(checkoutId)
	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "PlayerLeave"}
	arg2 := dero.Argument{Name: "id", DataType: "U", Value: checkoutId}
	args := dero.Arguments{arg1, arg2}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(Round.Contract, "[Holdero]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[Leave]", err)
		return
	}

	log.Println("[Holdero] Leave TX:", txid)
	rpc.AddLog("Leave Down TX: " + txid.TXID)
}

// Owner table settings for Holdero
//   - seats defines max players at table
//   - bb, sb and ante define big blind, small blind and antes. Ante can be 0
//   - chips defines if tables is using Dero or assets
//   - name and av are for name and avatar in owners id string
func SetTable(seats int, bb, sb, ante uint64, chips, name, av string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	var player playerId
	player.Id = rpc.Wallet.IdHash
	player.Name = name
	player.Avatar = av

	mar, _ := json.Marshal(player)
	hx := hex.EncodeToString(mar)

	var args dero.Arguments
	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "SetTable"}
	arg2 := dero.Argument{Name: "seats", DataType: "U", Value: seats}
	arg3 := dero.Argument{Name: "bigBlind", DataType: "U", Value: bb}
	arg4 := dero.Argument{Name: "smallBlind", DataType: "U", Value: sb}
	arg5 := dero.Argument{Name: "ante", DataType: "U", Value: ante}
	arg6 := dero.Argument{Name: "address", DataType: "S", Value: hx}
	txid := dero.Transfer_Result{}

	if Round.Version < 110 {
		args = dero.Arguments{arg1, arg2, arg3, arg4, arg5, arg6}
	} else if Round.Version == 110 {
		arg7 := dero.Argument{Name: "chips", DataType: "S", Value: chips}
		args = dero.Arguments{arg1, arg2, arg3, arg4, arg5, arg6, arg7}
	}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(Round.Contract, "[Holdero]", args, t, rpc.HighLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[SetTable]", err)
		return
	}

	log.Println("[Holdero] Set Table TX:", txid)
	rpc.AddLog("Set Table TX: " + txid.TXID)
}

// Submit blinds/ante to deal Holdero hand
func DealHand() (tx string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	if !rpc.Wallet.KeyLock {
		rpc.Wallet.ClientKey = GenerateKey()
	}

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "DealHand"}
	arg2 := dero.Argument{Name: "pcSeed", DataType: "H", Value: rpc.Wallet.ClientKey}
	args := dero.Arguments{arg1, arg2}
	txid := dero.Transfer_Result{}

	var amount uint64

	if Round.Pot == 0 {
		amount = Round.Ante + Round.SB
	} else if Round.Pot == Round.SB || Round.Pot == Round.Ante+Round.SB {
		amount = Round.Ante + Round.BB
	} else {
		amount = Round.Ante
	}

	t := []dero.Transfer{}
	if Round.Asset {
		t1 := dero.Transfer{
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      500,
			Burn:        0,
		}

		if Round.Tourney {
			t2 := dero.Transfer{
				SCID:        crypto.HashHexToHash(TourneySCID),
				Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
				Burn:        amount,
			}
			t = append(t, t1, t2)
		} else {
			t2 := rpc.GetAssetSCIDforTransfer(amount, Round.AssetID)
			if t2.Destination == "" {
				log.Println("[DealHand] Error getting asset SCID for transfer")
				return
			}
			t = append(t, t1, t2)
		}
	} else {
		t1 := dero.Transfer{
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      500,
			Burn:        amount,
		}
		t = append(t, t1)
	}

	fee := rpc.GasEstimate(Round.Contract, "[Holdero]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[DealHand]", err)
		return
	}

	Display.Res = ""
	log.Println("[Holdero] Deal TX:", txid)
	updateStatsWager(float64(amount) / 100000)
	rpc.AddLog("Deal TX: " + txid.TXID)

	return txid.TXID
}

// Make Holdero bet
func Bet(amt string) (tx string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "Bet"}
	args := dero.Arguments{arg1}
	txid := dero.Transfer_Result{}

	var t1 dero.Transfer
	if Round.Asset {
		if Round.Tourney {
			t1 = dero.Transfer{
				SCID:        crypto.HashHexToHash(TourneySCID),
				Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
				Burn:        rpc.ToAtomic(amt, 1),
			}
		} else {
			t1 = rpc.GetAssetSCIDforTransfer(rpc.ToAtomic(amt, 1), Round.AssetID)
			if t1.Destination == "" {
				log.Println("[Bet] Error getting asset SCID for transfer")
				return
			}
		}
	} else {
		t1 = dero.Transfer{
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      0,
			Burn:        rpc.ToAtomic(amt, 1),
		}
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(Round.Contract, "[Holdero]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[Bet]", err)
		return
	}

	if f, err := strconv.ParseFloat(amt, 64); err == nil {
		updateStatsWager(f)
	}

	Display.Res = ""
	Signal.PlacedBet = true
	log.Println("[Holdero] Bet TX:", txid)
	rpc.AddLog("Bet TX: " + txid.TXID)

	return txid.TXID
}

// Holdero check and fold
func Check() (tx string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "Bet"}
	args := dero.Arguments{arg1}
	txid := dero.Transfer_Result{}

	var t1 dero.Transfer
	if !Round.Asset {
		t1 = dero.Transfer{
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      0,
			Burn:        0,
		}
	} else {
		if Round.Tourney {
			t1 = dero.Transfer{
				SCID:        crypto.HashHexToHash(TourneySCID),
				Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
				Burn:        0,
			}
		} else {
			t1 = rpc.GetAssetSCIDforTransfer(0, Round.AssetID)
			if t1.Destination == "" {
				log.Println("[Check] Error getting asset SCID for transfer")
				return
			}
		}
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(Round.Contract, "[Holdero]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[Check]", err)
		return
	}

	Display.Res = ""
	log.Println("[Holdero] Check/Fold TX:", txid)
	rpc.AddLog("Check/Fold TX: " + txid.TXID)

	return txid.TXID
}

// Holdero single winner payout
//   - w defines which player the pot is going to
func PayOut(w string) string {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "Winner"}
	arg2 := dero.Argument{Name: "whoWon", DataType: "S", Value: w}
	args := dero.Arguments{arg1, arg2}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(Round.Contract, "[Holdero]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[PayOut]", err)
		return ""
	}

	log.Println("[Holdero] Payout TX:", txid)
	rpc.AddLog("Holdero Payout TX: " + txid.TXID)

	return txid.TXID
}

// Holdero split winners payout
//   - Pass in ranker from hand and folded bools to determine split
func PayoutSplit(r ranker, f1, f2, f3, f4, f5, f6 bool) string {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	ways := 0
	splitWinners := [6]string{"Zero", "Zero", "Zero", "Zero", "Zero", "Zero"}

	if r.p1HighCardArr[0] > 0 && !f1 {
		ways = 1
		splitWinners[0] = "Player1"
	}

	if r.p2HighCardArr[0] > 0 && !f2 {
		ways++
		splitWinners[1] = "Player2"
	}

	if r.p3HighCardArr[0] > 0 && !f3 {
		ways++
		splitWinners[2] = "Player3"
	}

	if r.p4HighCardArr[0] > 0 && !f4 {
		ways++
		splitWinners[3] = "Player4"
	}

	if r.p5HighCardArr[0] > 0 && !f5 {
		ways++
		splitWinners[4] = "Player5"
	}

	if r.p6HighCardArr[0] > 0 && !f6 {
		ways++
		splitWinners[5] = "Player6"
	}

	sort.Strings(splitWinners[:])

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "SplitWinner"}
	arg2 := dero.Argument{Name: "div", DataType: "U", Value: ways}
	arg3 := dero.Argument{Name: "split1", DataType: "S", Value: splitWinners[0]}
	arg4 := dero.Argument{Name: "split2", DataType: "S", Value: splitWinners[1]}
	arg5 := dero.Argument{Name: "split3", DataType: "S", Value: splitWinners[2]}
	arg6 := dero.Argument{Name: "split4", DataType: "S", Value: splitWinners[3]}
	arg7 := dero.Argument{Name: "split5", DataType: "S", Value: splitWinners[4]}
	arg8 := dero.Argument{Name: "split6", DataType: "S", Value: splitWinners[5]}

	args := dero.Arguments{arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(Round.Contract, "[Holdero]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[PayoutSplit]", err)
		return ""
	}

	log.Println("[Holdero] Split Winner TX:", txid)
	rpc.AddLog("Split Winner TX: " + txid.TXID)

	return txid.TXID
}

// Reveal Holdero hand key for showdown
func RevealKey(key string) {
	time.Sleep(6 * time.Second)
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "RevealKey"}
	arg2 := dero.Argument{Name: "pcSeed", DataType: "H", Value: key}
	args := dero.Arguments{arg1, arg2}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(Round.Contract, "[Holdero]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[RevealKey]", err)
		return
	}

	Display.Res = ""
	log.Println("[Holdero] Reveal TX:", txid)
	rpc.AddLog("Reveal TX: " + txid.TXID)
}

// Owner can shuffle deck for Holdero, clean above 0 can retrieve balance
func CleanTable(amt uint64) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "CleanTable"}
	arg2 := dero.Argument{Name: "amount", DataType: "U", Value: amt}
	args := dero.Arguments{arg1, arg2}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(Round.Contract, "[Holdero]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[CleanTable]", err)
		return
	}

	log.Println("[Holdero] Clean Table TX:", txid)
	rpc.AddLog("Clean Table TX: " + txid.TXID)
}

// Owner can timeout a player at Holdero table
func TimeOut() {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "TimeOut"}
	args := dero.Arguments{arg1}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(Round.Contract, "[Holdero]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[TimeOut]", err)
		return
	}

	log.Println("[Holdero] Timeout TX:", txid)
	rpc.AddLog("Timeout TX: " + txid.TXID)
}

// Owner can force start a Holdero table with empty seats
func ForceStat() {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "ForceStart"}
	args := dero.Arguments{arg1}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(Round.Contract, "[Holdero]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[ForceStart]", err)
		return
	}

	log.Println("[Holdero] Force Start TX:", txid)
	rpc.AddLog("Force Start TX: " + txid.TXID)
}

// Share asset url at Holdero table
//   - face and back are the names of assets
//   - faceUrl and backUrl are the Urls for those assets
func SharedDeckUrl(face, faceUrl, back, backUrl string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	var cards CardSpecs
	if face != "" && back != "" {
		cards.Faces.Name = face
		cards.Faces.Url = faceUrl
		cards.Backs.Name = back
		cards.Backs.Url = backUrl
	}

	if mar, err := json.Marshal(cards); err == nil {
		arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "Deck"}
		arg2 := dero.Argument{Name: "face", DataType: "S", Value: string(mar)}
		arg3 := dero.Argument{Name: "back", DataType: "S", Value: "nil"}
		args := dero.Arguments{arg1, arg2, arg3}
		txid := dero.Transfer_Result{}

		t1 := dero.Transfer{
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      0,
			Burn:        0,
		}

		t := []dero.Transfer{t1}
		fee := rpc.GasEstimate(Round.Contract, "[Holdero]", args, t, rpc.LowLimitFee)
		params := &dero.Transfer_Params{
			Transfers: t,
			SC_ID:     Round.Contract,
			SC_RPC:    args,
			Ringsize:  2,
			Fees:      fee,
		}

		if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
			log.Println("[SharedDeckUrl]", err)
			return
		}

		log.Println("[Holdero] Shared TX:", txid)
		rpc.AddLog("Shared TX: " + txid.TXID)
	}
}

// Deposit tournament chip bal with name to leader board SC
func TourneyDeposit(bal uint64, name string) {
	if bal > 0 {
		rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
		defer cancel()

		arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "Deposit"}
		arg2 := dero.Argument{Name: "name", DataType: "S", Value: name}
		args := dero.Arguments{arg1, arg2}
		txid := dero.Transfer_Result{}

		t1 := dero.Transfer{
			SCID:        crypto.HashHexToHash(TourneySCID),
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      0,
			Burn:        bal,
		}

		t := []dero.Transfer{t1}
		fee := rpc.GasEstimate(TourneySCID, "[Holdero]", args, t, rpc.LowLimitFee)
		params := &dero.Transfer_Params{
			Transfers: t,
			SC_ID:     TourneySCID,
			SC_RPC:    args,
			Ringsize:  2,
			Fees:      fee,
		}

		if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
			log.Println("[TourneyDeposit]", err)
			return
		}

		log.Println("[Holdero] Tournament Deposit TX:", txid)
		rpc.AddLog("Tournament Deposit TX: " + txid.TXID)

	} else {
		log.Println("[Holdero] No Tournament Chips")
	}
}

// Code latest SC code for Holdero public or private SC
//   - version defines which type of Holdero contract
//   - 0 for standard public
//   - 1 for standard private
//   - 2 for HGC
func GetHolderoCode(version int) string {
	if rpc.Daemon.IsConnected() {
		rpcClientD, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		var params dero.GetSC_Params
		switch version {
		case 0:
			params = dero.GetSC_Params{
				SCID:      HolderoSCID,
				Code:      true,
				Variables: false,
			}
		case 1:
			params = dero.GetSC_Params{
				SCID:      pHolderoSCID,
				Code:      true,
				Variables: false,
			}
		case 2:
			params = dero.GetSC_Params{
				SCID:      HGCHolderoSCID,
				Code:      true,
				Variables: false,
			}
		default:

		}

		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[GetHoldero110Code]", err)
			return ""
		}

		return result.Code

	}

	return ""
}

var unlockFee = uint64(300000)

// Contract unlock transfer
func OwnerT3(o bool) (t *dero.Transfer) {
	if o {
		t = &dero.Transfer{
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      0,
		}
	} else {
		if fee, ok := rpc.FindStringKey(rpc.RatingSCID, "ContractUnlock", rpc.Daemon.Rpc).(float64); ok {
			unlockFee = uint64(fee)
		} else {
			log.Println("[FetchFees] Could not get current contract unlock fee, using default")
		}

		t = &dero.Transfer{
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      unlockFee,
		}
	}

	return
}

// Install new Holdero SC
//   - pub defines public or private SC
func uploadHolderoContract(pub int) {
	if rpc.IsReady() {
		rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
		defer cancel()

		code := GetHolderoCode(pub)
		if code == "" {
			log.Println("[uploadHolderoContract] Could not get SC code")
			return
		}

		args := dero.Arguments{}
		txid := dero.Transfer_Result{}

		params := &dero.Transfer_Params{
			Transfers: []dero.Transfer{*OwnerT3(Poker.table_owner)},
			SC_Code:   code,
			SC_Value:  0,
			SC_RPC:    args,
			Ringsize:  2,
		}

		if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
			log.Println("[uploadHolderoContract]", err)
			return
		}

		log.Println("[Holdero] Upload TX:", txid)
		rpc.AddLog("Holdero Upload TX:" + txid.TXID)
	}
}
