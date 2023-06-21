package tarot

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"

	"github.com/SixofClubsss/dReams/rpc"
	dero "github.com/deroproject/derohe/rpc"
)

// Get Tarot SC data
func FetchTarotSC() {
	if rpc.Daemon.IsConnected() {
		rpcClientD, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		params := dero.GetSC_Params{
			SCID:      rpc.TarotSCID,
			Code:      false,
			Variables: true,
		}

		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[FetchTarotSC]", err)
			return
		}

		Reading_jv := result.VariableStringKeys["readings:"]
		if Reading_jv != nil {
			Iluma.Value.Readings = fmt.Sprint(Reading_jv)
		}
	}
}

// Find Tarot reading on SC
func FetchReading(tx string) {
	if rpc.Daemon.IsConnected() && len(tx) == 64 {
		rpcClientD, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		params := dero.GetSC_Params{
			SCID:      rpc.TarotSCID,
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
			Display_jv := result.VariableStringKeys["Display"]
			start := rpc.IntType(Reading_jv) - rpc.IntType(Display_jv)
			i := start
			for i < start+45 {
				h := "-readingTXID:"
				w := strconv.Itoa(i)
				TXID_jv := result.VariableStringKeys[w+h]

				if TXID_jv != nil {
					if fmt.Sprint(TXID_jv) == tx {
						Iluma.Value.Found = true
						Iluma.Value.Card1 = findTarotCard(result.VariableStringKeys[w+"-card1:"])
						Iluma.Value.Card2 = findTarotCard(result.VariableStringKeys[w+"-card2:"])
						Iluma.Value.Card3 = findTarotCard(result.VariableStringKeys[w+"-card3:"])
					}
				}
				i++
			}
		}
	}
}

// Find Tarot card from hash value
func findTarotCard(hash interface{}) int {
	if hash != nil {
		for i := 1; i < 79; i++ {
			finder := strconv.Itoa(i)
			card := sha256.Sum256([]byte(finder))
			str := hex.EncodeToString(card[:])

			if str == fmt.Sprint(hash) {
				return i
			}
		}
	}
	return 0
}

// Draw Iluma Tarot reading from SC
//   - num defines one or three card draw
func DrawReading(num int) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "Draw"}
	arg2 := dero.Argument{Name: "num", DataType: "U", Value: num}
	args := dero.Arguments{arg1, arg2}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        rpc.IlumaFee,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(rpc.TarotSCID, "[TarotReading]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     rpc.TarotSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[TarotReading]", err)
		return
	}

	Iluma.Value.Num = num
	Iluma.Value.Last = txid.TXID
	Iluma.Value.Notified = false

	log.Println("[TarotReading] Reading TX:", txid)
	rpc.AddLog("Tarot Reading TX: " + txid.TXID)

	Iluma.Value.CHeight = rpc.Wallet.Height
}
