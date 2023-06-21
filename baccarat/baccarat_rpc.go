package baccarat

import (
	"fmt"
	"log"
	"strconv"

	"github.com/SixofClubsss/dReams/rpc"
	"github.com/deroproject/derohe/cryptography/crypto"
	dero "github.com/deroproject/derohe/rpc"
)

type displayStrings struct {
	Total_w  string
	Player_w string
	Banker_w string
	Ties     string
	BaccMax  string
	BaccMin  string
	BaccRes  string
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

var Display displayStrings
var Bacc baccValues

// Get Baccarat SC data
func fetchBaccSC() {
	if rpc.Daemon.IsConnected() {
		rpcClientD, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		params := dero.GetSC_Params{
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

		if max, ok := Max_jv.(float64); ok {
			Display.BaccMax = fmt.Sprintf("%.0f", max/100000)
			Bacc.MaxBet = max / 100000
		} else {
			Display.BaccMax = "250"
			Bacc.MaxBet = 250
		}

		if min, ok := Min_jv.(float64); ok {
			Display.BaccMin = fmt.Sprintf("%.0f", min/100000)
			Bacc.MinBet = min / 100000
		} else {
			Display.BaccMin = "10"
			Bacc.MinBet = 10
		}
	}
}

// Get Baccarat SC code
func GetBaccCode() string {
	if rpc.Daemon.IsConnected() {
		rpcClientD, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		params := dero.GetSC_Params{
			SCID:      rpc.BaccSCID,
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
	if rpc.Daemon.IsConnected() && tx != "" {
		rpcClientD, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		params := dero.GetSC_Params{
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
			Display_jv := result.VariableStringKeys["display"]
			start := rpc.IntType(Total_jv) - rpc.IntType(Display_jv)
			i := start
			for i < start+45 {
				h := "-Hand#TXID:"
				w := strconv.Itoa(i)
				TXID_jv := result.VariableStringKeys[w+h]

				if TXID_jv != nil {
					if TXID_jv.(string) == tx {
						Bacc.Found = true
						Bacc.P_card1 = rpc.IntType(result.VariableStringKeys[w+"-Player x:"])
						Bacc.P_card2 = rpc.IntType(result.VariableStringKeys[w+"-Player y:"])
						Bacc.P_card3 = rpc.IntType(result.VariableStringKeys[w+"-Player z:"])
						Bacc.B_card1 = rpc.IntType(result.VariableStringKeys[w+"-Banker x:"])
						Bacc.B_card2 = rpc.IntType(result.VariableStringKeys[w+"-Banker y:"])
						Bacc.B_card3 = rpc.IntType(result.VariableStringKeys[w+"-Banker z:"])
						PTotal_jv := result.VariableStringKeys[w+"-Player total:"]
						BTotal_jv := result.VariableStringKeys[w+"-Banker total:"]

						p := rpc.IntType(PTotal_jv)
						b := rpc.IntType(BTotal_jv)
						if rpc.IntType(PTotal_jv) == rpc.IntType(BTotal_jv) {
							Display.BaccRes = fmt.Sprintf("Hand# %s Tie, %d & %d", w, p, b)
						} else if rpc.IntType(PTotal_jv) > rpc.IntType(BTotal_jv) {
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

// Place Baccarat bet
//   - amt to bet
//   - w defines where bet is placed (player, banker or tie)
func BaccBet(amt, w string) (tx string) {
	if Bacc.AssetID == "" || len(Bacc.AssetID) != 64 {
		log.Println("[Baccarat] Asset ID error")
		return "ID error"
	}

	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "PlayBaccarat"}
	arg2 := dero.Argument{Name: "betOn", DataType: "S", Value: w}
	args := dero.Arguments{arg1, arg2}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		SCID:        crypto.HashHexToHash(Bacc.AssetID),
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        rpc.ToAtomic(amt, 1),
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(Bacc.Contract, "[Baccarat]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     Bacc.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[BaccBet]", err)
		return
	}

	Bacc.Last = txid.TXID
	Bacc.Notified = false
	if w == "player" {
		log.Println("[Baccarat] Player TX:", txid)
		rpc.AddLog("Baccarat Player TX: " + txid.TXID)
	} else if w == "banker" {
		log.Println("[Baccarat] Banker TX:", txid)
		rpc.AddLog("Baccarat Banker TX: " + txid.TXID)
	} else {
		log.Println("[Baccarat] Tie TX:", txid)
		rpc.AddLog("Baccarat Tie TX: " + txid.TXID)
	}

	Bacc.CHeight = rpc.Wallet.Height

	return txid.TXID
}
