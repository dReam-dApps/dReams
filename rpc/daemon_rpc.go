package rpc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/deroproject/derohe/rpc"
	"github.com/ybbus/jsonrpc/v3"
)

const (
	DREAMSv      = "0.9.5d"
	NameSCID     = "0000000000000000000000000000000000000000000000000000000000000001"
	RatingSCID   = "c66a11ddb22912e92b0a7ab777ed0d343632d9e3c6e8a81452396ca84d2decb6"
	DreamsSCID   = "ad2e7b37c380cc1aed3a6b27224ddfc92a2d15962ca1f4d35e530dba0f9575a9"
	HgcSCID      = "e2e45ce26f70cb551951c855e81a12fee0bb6ebe80ef115c3f50f51e119c02f3"
	TourneySCID  = "c2e1ec16aed6f653aef99a06826b2b6f633349807d01fbb74cc0afb5ff99c3c7"
	HolderoSCID  = "e3f37573de94560e126a9020c0a5b3dfc7a4f3a4fbbe369fba93fbd219dc5fe9"
	pHolderoSCID = "896834d57628d3a65076d3f4d84ddc7c5daf3e86b66a47f018abda6068afe2e6"
	HHolderoSCID = "efe646c48977fd776fee73cdd3df147a2668d3b7d965cdb7a187dda4d23005d8"
	BaccSCID     = "8289c6109f41cbe1f6d5f27a419db537bf3bf30a25eff285241a36e1ae3e48a4"
	PredictSCID  = "eaa62b220fa1c411785f43c0c08ec59c761261cb58a0ccedc5b358e5ed2d2c95"
	pPredictSCID = "e5e49c9a6dc1c0dc8a94429a01bf758e705de49487cbd0b3e3550648d2460cdf"
	SportsSCID   = "ad11377c29a863523c1cc50a33ca13e861cc146a7c0496da58deaa1973e0a39f"
	pSportsSCID  = "fffdc4ea6d157880841feab335ab4755edcde4e60fec2fff661009b16f44fa94"
	TarotSCID    = "a6fc0033327073dd54e448192af929466596fce4d689302e558bc85ea8734a82"
	DerBnbSCID   = "cfbd566d3678dec6e6dfa3a919feae5306ab12af1485e8bcf9320bd5a122b1d3"
	GnomonSCID   = "a05395bb0cf77adc850928b0db00eb5ca7a9ccbafd9a38d021c8d299ad5ce1a4"
	DevAddress   = "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn"
	ArtAddress   = "dero1qy0khp9s9yw2h0eu20xmy9lth3zp5cacmx3rwt6k45l568d2mmcf6qgcsevzx"

	DAEMON_RPC_DEFAULT = "127.0.0.1:10102"
	DAEMON_RPC_REMOTE1 = "89.38.99.117:10102"
	DAEMON_RPC_REMOTE2 = "publicrpc1.dero.io:10102"
	// DAEMON_RPC_REMOTE3 = "dero-node.mysrv.cloud:10102"
	// DAEMON_RPC_REMOTE4 = "derostats.io:10102"
	DAEMON_RPC_REMOTE5 = "85.17.52.28:11012"
	DAEMON_RPC_REMOTE6 = "node.derofoundation.org:11012"
)

type daemon struct {
	Rpc     string
	Connect bool
	Height  uint64
}

var Daemon daemon
var Times times
var Display displayStrings
var Round holderoValues
var Bacc baccValues
var Signal signals
var Predict predictionValues
var Tarot tarotValues

// Convert hex value to string
func fromHextoString(h string) string {
	if str, err := hex.DecodeString(h); err == nil {
		return string(str)
	}

	return ""
}

// Set daemon rpc client with context and 5 sec cancel
func SetDaemonClient(addr string) (jsonrpc.RPCClient, context.Context, context.CancelFunc) {
	client := jsonrpc.NewClient("http://" + addr + "/json_rpc")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	return client, ctx, cancel
}

// Ping Dero blockchain for connection
func Ping() {
	rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
	defer cancel()

	var result string
	if err := rpcClientD.CallFor(ctx, &result, "DERO.Ping"); err != nil {
		Daemon.Connect = false
		return
	}

	if result == "Pong " {
		Daemon.Connect = true
	} else {
		Daemon.Connect = false
	}
}

// Get a daemons height
func DaemonHeight(tag, ep string) uint64 {
	rpcClientD, ctx, cancel := SetDaemonClient(ep)
	defer cancel()

	var result *rpc.GetHeight_Result
	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetHeight"); err != nil {
		log.Printf("[%s] %s\n", tag, err)
		return 0
	}

	return result.Height
}

// SC call gas estimate, 1320 Deri max
//   - tag for log print
//   - Pass args and transfers for call
func GasEstimate(scid, tag string, args rpc.Arguments, t []rpc.Transfer, max uint64) uint64 {
	rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
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

	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetGasEstimate", params); err != nil {
		log.Println(tag, err)
		return 0
	}

	log.Println(tag+" Gas Fee:", result.GasStorage+120)

	if result.GasStorage < max {
		return result.GasStorage + 120
	}

	return max + 120
}

// Get single string key result from SCID with daemon input
func FindStringKey(scid, key, daemon string) interface{} {
	rpcClientD, ctx, cancel := SetDaemonClient(daemon)
	defer cancel()

	var result *rpc.GetSC_Result
	params := rpc.GetSC_Params{
		SCID:      scid,
		Code:      false,
		Variables: true,
	}

	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
		log.Println("[FindStringKey]", err)
		return nil
	}

	value := result.VariableStringKeys[key]

	return value
}

// Get list of dReams dApps from contract store
//   - Uses remote daemon if no Daemon.Connect
func FetchDapps() (dApps []string) {
	dApps = []string{"Holdero", "Baccarat", "dSports and dPredictions", "Iluma", "DerBnb"}
	var daemon string
	if !Daemon.Connect {
		daemon = DAEMON_RPC_REMOTE5
	} else {
		daemon = Daemon.Rpc
	}

	if stored, ok := FindStringKey(RatingSCID, "dApps", daemon).(string); ok {
		if h, err := hex.DecodeString(stored); err == nil {
			json.Unmarshal(h, &dApps)
		}
	}

	return
}

// Get platform fees from on chain store
//   - Overwrites defualt fee values with current stored values
func FetchFees() {
	if fee, ok := FindStringKey(RatingSCID, "ContractUnlock", Daemon.Rpc).(float64); ok {
		UnlockFee = uint64(fee)
	} else {
		log.Println("[FetchFees] Could not get current contract unlock fee, using default")
	}

	if fee, ok := FindStringKey(RatingSCID, "ListingFee", Daemon.Rpc).(float64); ok {
		ListingFee = uint64(fee)
	} else {
		log.Println("[FetchFees] Could not get current listing fee, using default")
	}

	if fee, ok := FindStringKey(TarotSCID, "Fee", Daemon.Rpc).(float64); ok {
		IlumaFee = uint64(fee)
	} else {
		log.Println("[FetchFees] Could not get current Iluma fee, using default")
	}

	if fee, ok := FindStringKey(RatingSCID, "LowLimitFee", Daemon.Rpc).(float64); ok {
		LowLimitFee = uint64(fee)
	} else {
		log.Println("[FetchFees] Could not get current low fee limit, using default")
	}

	if fee, ok := FindStringKey(RatingSCID, "HighLimitFee", Daemon.Rpc).(float64); ok {
		HighLimitFee = uint64(fee)
	} else {
		log.Println("[FetchFees] Could not get current high fee limit, using default")
	}
}

// Check Gnomon SC for stored contract owner
func CheckForIndex(scid string) interface{} {
	rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
	defer cancel()

	var result *rpc.GetSC_Result
	params := rpc.GetSC_Params{
		SCID:      GnomonSCID,
		Code:      false,
		Variables: true,
	}

	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
		log.Println("[CheckForIndex]", err)
		return nil
	}

	owner := result.VariableStringKeys[scid+"owner"]
	address := DeroAddress(owner)

	return address
}

// Get code of a SC
func GetSCCode(scid string) string {
	if Daemon.Connect {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      scid,
			Code:      true,
			Variables: false,
		}

		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[GetSCCode]", err)
			return ""
		}

		return result.Code
	}
	return ""
}

// Get name service SC code
func GetNameServiceCode() string {
	if Daemon.Connect {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      NameSCID,
			Code:      true,
			Variables: false,
		}

		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[GetNameServiceCode]", err)
			return ""
		}

		return result.Code
	}
	return ""
}

// Get Gnomon SC code
func GetGnomonCode() string {
	if Daemon.Connect {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      GnomonSCID,
			Code:      true,
			Variables: false,
		}

		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[GetGnomonCode]", err)
			return ""
		}

		return result.Code
	}
	return ""
}

// Get all asset SCIDs from collection
func GetG45Collection(scid string) (scids []string) {
	rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
	defer cancel()

	var result *rpc.GetSC_Result
	params := rpc.GetSC_Params{
		SCID:      scid,
		Code:      false,
		Variables: true,
	}

	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
		log.Println("[GetG45Collection]", err)
		return nil
	}

	i := 0
	for {
		n := strconv.Itoa(i)
		asset := result.VariableStringKeys["assets_"+n]

		if asset == nil {
			break
		} else {
			if hx, err := hex.DecodeString(fmt.Sprint(asset)); err != nil {
				log.Println("[GetG45Collection]", err)
				i++
			} else {
				split := strings.Split(string(hx), ",")
				for i := range split {
					sc := strings.Split(split[i], ":")
					trim := strings.Trim(sc[0], `{"`)
					scids = append(scids, trim)
				}
				i++
			}
		}
	}

	return
}

// Get single TX data with GetTransaction
func GetDaemonTx(txid string) *rpc.Tx_Related_Info {
	rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
	defer cancel()

	var result *rpc.GetTransaction_Result
	params := rpc.GetTransaction_Params{
		Tx_Hashes: []string{txid},
	}

	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetTransaction", params); err != nil {
		log.Println("[GetDaemonTx]", err)
		return nil
	}

	if result.Txs != nil {
		return &result.Txs[0]
	}

	return nil
}

// Verify TX signer with GetTransaction
func VerifySigner(txid string) bool {
	rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
	defer cancel()

	var result *rpc.GetTransaction_Result
	params := rpc.GetTransaction_Params{
		Tx_Hashes: []string{txid},
	}

	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetTransaction", params); err != nil {
		log.Println("[VerifySigner]", err)
		return false
	}

	if result.Txs[0].Signer == Wallet.Address {
		return true
	}

	return false
}

// Get Holdero SC data
func FetchHolderoSC() {
	if Daemon.Connect && Signal.Contract {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
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
			Last_jv := result.VariableStringKeys["Last"]

			if Last_jv != nil {
				Round.Last = int64(Last_jv.(float64))
			} else {
				Round.Last = 0
			}

			if Tourney_jv == nil {
				Round.Tourney = false
				if Chips_jv != nil {
					if fromHextoString(Chips_jv.(string)) == "ASSET" {
						Round.Asset = true
						if _, ok := result.VariableStringKeys["dReams"].(string); ok {
							Pot_jv = result.Balances[DreamsSCID]
							Round.AssetID = DreamsSCID
						} else if _, ok = result.VariableStringKeys["HGC"].(string); ok {
							Pot_jv = result.Balances[HgcSCID]
							Round.AssetID = HgcSCID
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
					if fromHextoString(Chips_jv.(string)) == "ASSET" {
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
				Wallet.KeyLock = false
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
					Signal.Clicked = true
					Signal.CHeight = Wallet.Height
					Signal.Reveal = true
					go RevealKey(Wallet.ClientKey)
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
						if Wallet.Height > Times.Kick_block+3 {
							TimeOut()
							Times.Kick_block = Wallet.Height
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

// Code for v1.0.0 Holdero SC
func GetHoldero100Code() string {
	if Daemon.Connect {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      "95e69b382044ddc1467e030a80905cf637729612f65624e8d17bf778d4362b8d",
			Code:      true,
			Variables: false,
		}

		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[GetHoldero100Code]", err)
			return ""
		}

		return result.Code

	}

	return ""
}

// Code for v1.1.0 Holdero public or private SC
//   - version defines which type of Holdero contratc
//   - 0 for standard public
//   - 1 for standard private
//   - 2 for HGC
func GetHoldero110Code(version int) string {
	if Daemon.Connect {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
		defer cancel()

		var result *rpc.GetSC_Result
		var params rpc.GetSC_Params
		switch version {
		case 0:
			params = rpc.GetSC_Params{
				SCID:      HolderoSCID,
				Code:      true,
				Variables: false,
			}
		case 1:
			params = rpc.GetSC_Params{
				SCID:      pHolderoSCID,
				Code:      true,
				Variables: false,
			}
		case 2:
			params = rpc.GetSC_Params{
				SCID:      HHolderoSCID,
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

// Get Baccarat SC data
func FetchBaccSC() {
	if Daemon.Connect {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      Bacc.Contract,
			Code:      false,
			Variables: true,
		}

		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[FetchBaccSC]", err)
			return
		}

		Asset_jv := result.VariableStringKeys["tokenSCID"]
		Total_jv := result.VariableStringKeys["TotalHandsPlayed:"]
		Player_jv := result.VariableStringKeys["Player Wins:"]
		Banker_jv := result.VariableStringKeys["Banker Wins:"]
		Min_jv := result.VariableStringKeys["Min Bet:"]
		Max_jv := result.VariableStringKeys["Max Bet:"]
		Ties_jv := result.VariableStringKeys["Ties:"]
		// Pot_jv = result.Balances[dReamsSCID]
		// Pot_jv = result.Balances["0000000000000000000000000000000000000000000000000000000000000000"]
		if Asset_jv != nil {
			Bacc.AssetID = fmt.Sprint(Asset_jv)
		}

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
	}
}

// Get Baccarat SC code
func GetBaccCode() string {
	if Daemon.Connect {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      BaccSCID,
			Code:      true,
			Variables: false,
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println("[GetBaccCode]", err)
			return ""
		}

		return result.Code
	}

	return ""
}

// Find played Baccarat hand
func FetchBaccHand(tx string) {
	if Daemon.Connect && tx != "" {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      Bacc.Contract,
			Code:      false,
			Variables: true,
		}

		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[FetchBaccHand]", err)
			return
		}

		Total_jv := result.VariableStringKeys["TotalHandsPlayed:"]
		if Total_jv != nil {
			start := int(Total_jv.(float64)) - 21
			i := start
			for i < start+45 {
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
							Display.BaccRes = fmt.Sprintf("Hand# %s Tie, %d & %d", w, p, b)
						} else if PTotal_jv.(float64) > BTotal_jv.(float64) {
							Display.BaccRes = fmt.Sprintf("Hand# %s Player Wins, %d over %d", w, p, b)
						} else {
							Display.BaccRes = fmt.Sprintf("Hand# %s Banker Wins, %d over %d", w, b, p)
						}
					}
				}
				i++
			}
		}
	}
}

// Check dSports/dPrediction SC for dev address
func ValidBetContract(scid string) bool {
	rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
	defer cancel()

	var result *rpc.GetSC_Result
	params := rpc.GetSC_Params{
		SCID:      scid,
		Code:      false,
		Variables: true,
	}

	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
		log.Println("[ValidBetContract]", err)
		return false
	}

	d := fmt.Sprint(result.VariableStringKeys["dev"])

	return DeroAddress(d) == DevAddress
}

// Get dPrediction final TXID
func FetchPredictionFinal(scid string) (txid string) {
	if Daemon.Connect {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
		defer cancel()

		params := &rpc.GetSC_Params{
			SCID:      scid,
			Code:      false,
			Variables: true,
		}

		var result *rpc.GetSC_Result
		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[FetchPredictionFinal]", err)
			return ""
		}

		p_txid := result.VariableStringKeys["p_final_txid"]

		if p_txid != nil {
			txid = fmt.Sprint(p_txid)

		}
		return txid
	}

	return ""
}

// Get dPrediction SC code for public and private SC
//   - pub defines public or private contract
func GetPredictCode(pub int) string {
	if Daemon.Connect {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
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

		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[GetPredictCode]", err)
			return ""
		}

		return result.Code
	}
	return ""
}

// Get dSports SC code for public and private SC
//   - pub defines public or private contract
func GetSportsCode(pub int) string {
	if Daemon.Connect {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
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

		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[GetSportsCode]", err)
			return ""
		}

		return result.Code
	}
	return ""
}

// Get recent dSports final results and TXIDs
func FetchSportsFinal(scid string) (finals []string) {
	if Daemon.Connect {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
		defer cancel()

		params := &rpc.GetSC_Params{
			SCID:      scid,
			Code:      false,
			Variables: true,
		}

		var result *rpc.GetSC_Result
		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[FetchSportsFinal]", err)
			return
		}

		played := result.VariableStringKeys["s_played"]
		if played != nil {
			start := int(played.(float64)) - 4
			i := start
			for {
				str := fmt.Sprint(i)
				game := result.VariableStringKeys["s_final_"+str]
				s_txid := result.VariableStringKeys["s_final_txid_"+str]

				if s_txid != nil && game != nil {
					decode, _ := hex.DecodeString(game.(string))
					final := str + "   " + string(decode) + "   " + s_txid.(string)
					finals = append(finals, final)
				}

				i++
				if i > start+4 {
					break
				}
			}
		}
	}

	return
}

// Get Tarot SC data
func FetchTarotSC() {
	if Daemon.Connect {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      TarotSCID,
			Code:      false,
			Variables: true,
		}

		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[FetchTarotSC]", err)
			return
		}

		Reading_jv := result.VariableStringKeys["readings:"]
		if Reading_jv != nil {
			Display.Readings = fmt.Sprint(Reading_jv)
		}
	}
}

// Find Tarot reading on SC
func FetchTarotReading(tx string) {
	if Daemon.Connect && len(tx) == 64 {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      TarotSCID,
			Code:      false,
			Variables: true,
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println("[FetchTarotReading]", err)
			return
		}

		Reading_jv := result.VariableStringKeys["readings:"]
		if Reading_jv != nil {
			start := int(Reading_jv.(float64)) - 21
			i := start
			for i < start+45 {
				h := "-readingTXID:"
				w := strconv.Itoa(i)
				TXID_jv := result.VariableStringKeys[w+h]

				if TXID_jv != nil {
					if TXID_jv.(string) == tx {
						Tarot.Found = true
						Tarot.Card1 = findTarotCard(result.VariableStringKeys[w+"-card1:"])
						Tarot.Card2 = findTarotCard(result.VariableStringKeys[w+"-card2:"])
						Tarot.Card3 = findTarotCard(result.VariableStringKeys[w+"-card3:"])
					}
				}
				i++
			}
		}
	}
}

// Get difficulty from a daemon
func GetDifficulty(ep string) float64 {
	rpcClientD, ctx, cancel := SetDaemonClient(ep)
	defer cancel()

	var result *rpc.GetInfo_Result
	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetInfo"); err != nil {
		log.Println("[GetDifficulty]", err)
		return 0
	}

	return float64(result.Difficulty)
}

// Get average block time from a daemon
func GetBlockTime(ep string) float64 {
	rpcClientD, ctx, cancel := SetDaemonClient(ep)
	defer cancel()

	var result *rpc.GetInfo_Result
	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetInfo"); err != nil {
		log.Println("[GetBlockTime]", err)
		return 0
	}

	return float64(result.AverageBlockTime50)
}
