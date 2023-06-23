package prediction

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/SixofClubsss/dReams/holdero"
	"github.com/SixofClubsss/dReams/rpc"
	dero "github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/walletapi"
)

const (
	PredictSCID  = "eaa62b220fa1c411785f43c0c08ec59c761261cb58a0ccedc5b358e5ed2d2c95"
	PPredictSCID = "e5e49c9a6dc1c0dc8a94429a01bf758e705de49487cbd0b3e3550648d2460cdf"
	SportsSCID   = "ad11377c29a863523c1cc50a33ca13e861cc146a7c0496da58deaa1973e0a39f"
	PSportsSCID  = "fffdc4ea6d157880841feab335ab4755edcde4e60fec2fff661009b16f44fa94"
)

// Place higher prediction to SC
//   - addr only needed if dService is placing prediction
func PredictHigher(scid, addr string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	amt := uint64(Predict.Amount)

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "Predict"}
	arg2 := dero.Argument{Name: "pre", DataType: "U", Value: 1}
	arg3 := dero.Argument{Name: "addr", DataType: "S", Value: addr}
	args := dero.Arguments{arg1, arg2, arg3}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        amt,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(scid, "[Predictions]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[PredictHigher]", err)
		return
	}

	log.Println("[Predictions] Prediction TX:", txid)
	rpc.AddLog("Prediction TX: " + txid.TXID)
}

// Place lower prediction to SC
//   - addr only needed if dService is placing prediction
func PredictLower(scid, addr string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	amt := uint64(Predict.Amount)

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "Predict"}
	arg2 := dero.Argument{Name: "pre", DataType: "U", Value: 0}
	arg3 := dero.Argument{Name: "addr", DataType: "S", Value: addr}
	args := dero.Arguments{arg1, arg2, arg3}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        amt,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(scid, "[Predictions]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[PredictLower]", err)
		return
	}

	log.Println("[Predictions] Prediction TX:", txid)
	rpc.AddLog("Prediction TX: " + txid.TXID)
}

// dService prediction place by received tx
//   - amt to send
//   - p is what prediction
//   - addr of placed bet and to send reply message
//   - src and pre_tx used in reply message
func AutoPredict(p int, amt, src uint64, scid, addr, pre_tx string) (tx string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	var hl string
	chopped_scid := scid[:6] + "..." + scid[58:]
	chopped_txid := pre_tx[:6] + "..." + pre_tx[58:]
	switch p {
	case 0:
		hl = "Lower"
	case 1:
		hl = "Higher"
	}

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "Predict"}
	arg2 := dero.Argument{Name: "pre", DataType: "U", Value: p}
	arg3 := dero.Argument{Name: "addr", DataType: "S", Value: addr}
	args := dero.Arguments{arg1, arg2, arg3}
	txid := dero.Transfer_Result{}

	response := dero.Arguments{
		{Name: dero.RPC_DESTINATION_PORT, DataType: dero.DataUint64, Value: uint64(0)},
		{Name: dero.RPC_SOURCE_PORT, DataType: dero.DataUint64, Value: src},
		{Name: dero.RPC_COMMENT, DataType: dero.DataString, Value: fmt.Sprintf("Placed a %s %s bet on %s at height %s,  %s", walletapi.FormatMoney(amt), hl, chopped_scid, rpc.Wallet.Display.Height, chopped_txid)},
	}

	t1 := dero.Transfer{
		Destination: addr,
		Amount:      1,
		Burn:        amt,
		Payload_RPC: response,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(scid, "[AutoPredict]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[AutoPredict]", err)
		return
	}

	log.Println("[AutoPredict] Prediction TX:", txid)
	rpc.AddLog("AutoPredict TX: " + txid.TXID)

	return txid.TXID
}

// dService refund if bet void
//   - amt to send
//   - addr to send refund to
//   - src, msg and refund_tx used in reply message
func ServiceRefund(amt, src uint64, scid, addr, msg, refund_tx string) (tx string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	chopped_scid := scid[:6] + "..." + scid[58:]
	chopped_txid := refund_tx[:6] + "..." + refund_tx[58:]
	response := dero.Arguments{
		{Name: dero.RPC_DESTINATION_PORT, DataType: dero.DataUint64, Value: uint64(0)},
		{Name: dero.RPC_SOURCE_PORT, DataType: dero.DataUint64, Value: src},
		{Name: dero.RPC_COMMENT, DataType: dero.DataString, Value: msg + fmt.Sprintf(", refunded %s bet on %s at height %s,  %s", walletapi.FormatMoney(amt), chopped_scid, rpc.Wallet.Display.Height, chopped_txid)},
	}

	t1 := dero.Transfer{
		Destination: addr,
		Amount:      amt,
		Burn:        0,
		Payload_RPC: response,
	}

	txid := dero.Transfer_Result{}
	t := []dero.Transfer{t1}
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_RPC:    dero.Arguments{},
		Ringsize:  16,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[ServiceRefund]", err)
		return
	}

	log.Println("[ServiceRefund] Refund TX:", txid)
	rpc.AddLog("Refund TX: " + txid.TXID)

	return txid.TXID
}

// dService sports book by received tx
//   - amt to send
//   - pre is what team
//   - n is the game number
//   - addr of placed bet and to send reply message
//   - src, abv and tx used in reply message
func AutoBook(amt, pre, src uint64, n, abv, scid, addr, book_tx string) (tx string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	chopped_scid := scid[:6] + "..." + scid[58:]
	chopped_txid := book_tx[:6] + "..." + book_tx[58:]
	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "Book"}
	arg2 := dero.Argument{Name: "pre", DataType: "U", Value: pre}
	arg3 := dero.Argument{Name: "n", DataType: "S", Value: n}
	arg4 := dero.Argument{Name: "addr", DataType: "S", Value: addr}
	args := dero.Arguments{arg1, arg2, arg3, arg4}
	txid := dero.Transfer_Result{}

	response := dero.Arguments{
		{Name: dero.RPC_DESTINATION_PORT, DataType: dero.DataUint64, Value: uint64(0)},
		{Name: dero.RPC_SOURCE_PORT, DataType: dero.DataUint64, Value: src},
		{Name: dero.RPC_COMMENT, DataType: dero.DataString, Value: fmt.Sprintf("Placed a %s %s bet on %s at height %s,  %s", walletapi.FormatMoney(amt), abv, chopped_scid, rpc.Wallet.Display.Height, chopped_txid)},
	}

	t1 := dero.Transfer{
		Destination: addr,
		Amount:      1,
		Burn:        amt,
		Payload_RPC: response,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(scid, "[AutoBook]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[AutoBook]", err)
		return
	}

	log.Println("[AutoBook] Book TX:", txid)
	rpc.AddLog("AutoBook TX: " + txid.TXID)

	return txid.TXID
}

// Owner update for bet SC vars
//   - ta, tb, tc are contracts time limits. Only ta, tb needed for dSports
//   - l is the max bet limit per initialized bet
//   - hl is the max amount of games that can be ran at once
func VarUpdate(scid string, ta, tb, tc, l, hl int) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "VarUpdate"}
	arg2 := dero.Argument{Name: "ta", DataType: "U", Value: ta}
	arg3 := dero.Argument{Name: "tb", DataType: "U", Value: tb}
	arg5 := dero.Argument{Name: "l", DataType: "U", Value: l}

	var args dero.Arguments
	var arg4, arg6 dero.Argument
	if hl > 0 {
		arg4 = dero.Argument{Name: "d", DataType: "U", Value: tc}
		arg6 = dero.Argument{Name: "hl", DataType: "U", Value: hl}
		args = dero.Arguments{arg1, arg2, arg3, arg4, arg5, arg6}
	} else {
		arg4 = dero.Argument{Name: "tc", DataType: "U", Value: tc}
		args = dero.Arguments{arg1, arg2, arg3, arg4, arg5}
	}

	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(scid, "[VarUpdate]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[VarUpdate]", err)
		return
	}

	log.Println("[VarUpdate] VarUpdate TX:", txid)
	rpc.AddLog("VarUpdate TX: " + txid.TXID)
}

// Owner can add new co-owner to bet SC
//   - addr of new co-owner
func AddOwner(scid, addr string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "AddSigner"}
	arg2 := dero.Argument{Name: "new", DataType: "S", Value: addr}
	args := dero.Arguments{arg1, arg2}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(scid, "[Predictions]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[AddSigner]", err)
		return
	}

	log.Println("[Predictions] Add Signer TX:", txid)
	rpc.AddLog("Add Signer TX: " + txid.TXID)
}

// Owner can remove co-owner from bet SC
//   - num defines which co-owner to remove
func RemoveOwner(scid string, num int) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "RemoveSigner"}
	arg2 := dero.Argument{Name: "remove", DataType: "U", Value: num}
	args := dero.Arguments{arg1, arg2}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(scid, "[Predictions]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[RemoveSigner]", err)
		return
	}

	log.Println("[Predictions] Remove Signer TX:", txid)
	rpc.AddLog("Remove Signer: " + txid.TXID)
}

// User can refund a void dPrediction payout from SC
//   - tic is the prediction id string
func PredictionRefund(scid, tic string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "Refund"}
	arg2 := dero.Argument{Name: "tic", DataType: "S", Value: "p-1-1"}
	args := dero.Arguments{arg1, arg2}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(scid, "[Predictions]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[PredictionRefund]", err)
		return
	}

	log.Println("[Predictions] Refund TX:", txid)
	rpc.AddLog("Refund TX: " + txid.TXID)
}

// Book sports team on dSports SC
//   - multi defines 1x, 3x or 5x the minimum
//   - n is the game number
//   - a is amount to book
//   - pick is team to book
func PickTeam(scid, multi, n string, a uint64, pick int) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	var amt uint64
	switch multi {
	case "1x":
		amt = a
	case "3x":
		amt = a * 3
	case "5x":
		amt = a * 5
	default:
		amt = a
	}

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "Book"}
	arg2 := dero.Argument{Name: "n", DataType: "S", Value: n}
	arg3 := dero.Argument{Name: "pre", DataType: "U", Value: pick}
	arg4 := dero.Argument{Name: "addr", DataType: "S", Value: rpc.Wallet.Address}
	args := dero.Arguments{arg1, arg2, arg3, arg4}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        amt,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(scid, "[Sports]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[PickTeam]", err)
		return
	}

	log.Println("[Sports] Pick TX:", txid)
	rpc.AddLog("Pick TX: " + txid.TXID)
}

// User can refund a void dSports payout from SC
//   - tic is the bet id string
//   - n is the game number
func SportsRefund(scid, tic, n string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "Refund"}
	arg2 := dero.Argument{Name: "tic", DataType: "S", Value: tic}
	arg3 := dero.Argument{Name: "n", DataType: "S", Value: n}
	args := dero.Arguments{arg1, arg2, arg3}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(scid, "[Sports]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[SportsRefund]", err)
		return
	}

	log.Println("[Sports] Refund TX:", txid)
	rpc.AddLog("Refund TX: " + txid.TXID)
}

// Owner sets a dSports game
//   - end is unix ending time
//   - amt of single prediction
//   - dep allows owner to add a initial deposit
//   - game is name of game, formatted TEAM--TEAM
//   - feed defines where price api data is sourced from
func SetSports(end int, amt, dep uint64, scid, league, game, feed string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "S_start"}
	arg2 := dero.Argument{Name: "end", DataType: "U", Value: end}
	arg3 := dero.Argument{Name: "amt", DataType: "U", Value: amt}
	arg4 := dero.Argument{Name: "league", DataType: "S", Value: league}
	arg5 := dero.Argument{Name: "game", DataType: "S", Value: game}
	arg6 := dero.Argument{Name: "feed", DataType: "S", Value: feed}
	args := dero.Arguments{arg1, arg2, arg3, arg4, arg5, arg6}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        dep,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(scid, "[Sports]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[SetSports]", err)
		return
	}

	log.Println("[Sports] Set TX:", txid)
	rpc.AddLog("Set Sports TX: " + txid.TXID)
}

// Owner sets up a dPrediction prediction
//   - end is unix ending time
//   - mark can be predefined or passed as 0 if mark is to be posted live
//   - amt of single prediction
//   - dep allows owner to add a initial deposit
//   - predict is name of what is being predicted
//   - feed defines where price api data is sourced from
func SetPrediction(end, mark int, amt, dep uint64, scid, predict, feed string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "P_start"}
	arg2 := dero.Argument{Name: "end", DataType: "U", Value: end}
	arg3 := dero.Argument{Name: "amt", DataType: "U", Value: amt}
	arg4 := dero.Argument{Name: "predict", DataType: "S", Value: predict}
	arg5 := dero.Argument{Name: "feed", DataType: "S", Value: feed}
	arg6 := dero.Argument{Name: "mark", DataType: "U", Value: mark}
	args := dero.Arguments{arg1, arg2, arg3, arg4, arg5, arg6}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        dep,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(scid, "[Predictions]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[SetPrediction]", err)
		return
	}

	log.Println("[Predictions] Set TX:", txid)
	rpc.AddLog("Set Prediction TX: " + txid.TXID)
}

// Owner cancel for initiated bet for dSports and dPrediction contracts
//   - b defines sports or prediction log print
func CancelInitiatedBet(scid string, b int) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "Cancel"}
	args := dero.Arguments{arg1}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	var tag string
	if b == 0 {
		tag = "[Predictions]"
	} else {
		tag = "[Sports]"
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(scid, tag, args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[CancelInitiatedBet]", err)
		return
	}

	if b == 0 {
		log.Println("[Predictions] Cancel TX:", txid)
		rpc.AddLog("Cancel Prediction TX: " + txid.TXID)
	} else {
		log.Println("[Sports] Cancel TX:", txid)
		rpc.AddLog("Cancel Sports TX: " + txid.TXID)
	}
}

// Post mark to prediction SC
//   - price is the posted mark for prediction
func PostPrediction(scid string, price int) (tx string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "Post"}
	arg2 := dero.Argument{Name: "price", DataType: "U", Value: price}
	args := dero.Arguments{arg1, arg2}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(scid, "[Predictions]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[PostPrediction]", err)
		return
	}

	log.Println("[Predictions] Post TX:", txid)
	rpc.AddLog("Post TX: " + txid.TXID)

	return txid.TXID
}

// dSports SC payout
//   - num is game number
//   - team is winning team for game number
func EndSports(scid, num, team string) (tx string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "S_end"}
	arg2 := dero.Argument{Name: "n", DataType: "S", Value: num}
	arg3 := dero.Argument{Name: "team", DataType: "S", Value: team}
	args := dero.Arguments{arg1, arg2, arg3}
	txid := dero.Transfer_Result{}

	t := []dero.Transfer{}
	fee := rpc.GasEstimate(scid, "[Sports]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[EndSports]", err)
		return
	}

	log.Println("[Sports] Payout TX:", txid)
	rpc.AddLog("Sports Payout TX: " + txid.TXID)

	return txid.TXID
}

// dPrediction SC payout
//   - price is final prediction results
func EndPrediction(scid string, price int) (tx string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "P_end"}
	arg2 := dero.Argument{Name: "price", DataType: "U", Value: price}
	args := dero.Arguments{arg1, arg2, arg2}
	txid := dero.Transfer_Result{}

	t := []dero.Transfer{}
	fee := rpc.GasEstimate(scid, "[Predictions]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_Value:  0,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[EndPrediction]", err)
		return
	}

	log.Println("[Predictions] Payout TX:", txid)
	rpc.AddLog("Prediction Payout TX: " + txid.TXID)

	return txid.TXID
}

// Check dSports/dPrediction SC for dev address
func ValidBetContract(scid string) bool {
	rpcClientD, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
	defer cancel()

	var result *dero.GetSC_Result
	params := dero.GetSC_Params{
		SCID:      scid,
		Code:      false,
		Variables: true,
	}

	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
		log.Println("[ValidBetContract]", err)
		return false
	}

	d := fmt.Sprint(result.VariableStringKeys["dev"])

	return rpc.DeroAddressFromKey(d) == rpc.DevAddress
}

// Get dPrediction final TXID
func FetchPredictionFinal(scid string) (txid string) {
	if rpc.Daemon.IsConnected() {
		rpcClientD, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		params := &dero.GetSC_Params{
			SCID:      scid,
			Code:      false,
			Variables: true,
		}

		var result *dero.GetSC_Result
		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[FetchPredictionFinal]", err)
			return ""
		}

		p_txid := result.VariableStringKeys["p_final_txid"]

		if p_txid != nil {
			txid = fmt.Sprint(p_txid)
		}
	}

	return
}

// Get dPrediction SC code for public and private SC
//   - pub defines public or private contract
func GetPredictCode(pub int) string {
	if rpc.Daemon.IsConnected() {
		rpcClientD, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		var params dero.GetSC_Params
		if pub == 1 {
			params = dero.GetSC_Params{
				SCID:      PPredictSCID,
				Code:      true,
				Variables: false,
			}
		} else {
			params = dero.GetSC_Params{
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
	if rpc.Daemon.IsConnected() {
		rpcClientD, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		var params dero.GetSC_Params
		if pub == 1 {
			params = dero.GetSC_Params{
				SCID:      PSportsSCID,
				Code:      true,
				Variables: false,
			}
		} else {
			params = dero.GetSC_Params{
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
	if rpc.Daemon.IsConnected() {
		rpcClientD, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		params := &dero.GetSC_Params{
			SCID:      scid,
			Code:      false,
			Variables: true,
		}

		var result *dero.GetSC_Result
		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[FetchSportsFinal]", err)
			return
		}

		played := result.VariableStringKeys["s_played"]
		if played != nil {
			start := rpc.IntType(played) - 4
			i := start
			for {
				str := fmt.Sprint(i)
				game := result.VariableStringKeys["s_final_"+str]
				s_txid := result.VariableStringKeys["s_final_txid_"+str]

				if s_txid != nil && game != nil {
					if decode, err := hex.DecodeString(fmt.Sprint(game)); err == nil {
						final := str + "   " + string(decode) + "   " + fmt.Sprint(s_txid)
						finals = append(finals, final)
					}
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

// Install new bet SC
//   - c defines dSports or dPrediction contract
//   - pub defines public or private contract
func UploadBetContract(c bool, pub int) {
	if rpc.IsReady() {
		rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
		defer cancel()

		var fee uint64
		var code string

		if c {
			fee = 12500
			code = GetPredictCode(pub)
			if code == "" {
				log.Println("[UploadBetContract] Could not get SC code")
				return
			}
		} else {
			fee = 14500
			code = GetSportsCode(pub)
			if code == "" {
				log.Println("[UploadBetContract] Could not get SC code")
				return
			}
		}

		args := dero.Arguments{}
		txid := dero.Transfer_Result{}

		params := &dero.Transfer_Params{
			Transfers: []dero.Transfer{*holdero.OwnerT3(Predict.owner)},
			SC_Code:   code,
			SC_Value:  0,
			SC_RPC:    args,
			Ringsize:  2,
			Fees:      fee,
		}

		if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
			log.Println("[UploadBetContract]", err)
			return
		}

		if c {
			log.Println("[Predictions] Upload TX:", txid)
			rpc.AddLog("Prediction Upload TX:" + txid.TXID)
		} else {
			log.Println("[Sports] Upload TX:", txid)
			rpc.AddLog("Sports Upload TX:" + txid.TXID)
		}
	}
}
