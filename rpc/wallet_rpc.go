package rpc

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/ybbus/jsonrpc/v3"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var dreamsV = semver.MustParse("0.10.1-dev") //"0.10.1d"

func Version() semver.Version {
	return dreamsV
}

// Prints session log entry to Wallet.LogEntry and stdout
func PrintLog(format string, a ...any) {
	text := fmt.Sprintf(format, a...)
	if Wallet.LogEntry != nil {
		Wallet.LogEntry.Append(fmt.Sprintf("\n\n%s  %s", time.Now().Format("2006/01/02 15:04:05"), text))
		Wallet.LogEntry.Refresh()
	}

	logger.Println(text)
}

// Prints session error entry to Wallet.LogEntry and stdout
func PrintError(format string, a ...any) {
	text := fmt.Sprintf(format, a...)
	if Wallet.LogEntry != nil {
		Wallet.LogEntry.Append(fmt.Sprintf("\n\n%s  [ERROR] - %s", time.Now().Format("2006/01/02 15:04:05"), text))
		Wallet.LogEntry.Refresh()
	}

	logger.Errorln(text)
}

// Make gui log for txs with save function.
// print out for dApp version if being imported
func SessionLog(tag string, dapp semver.Version) *fyne.Container {
	Wallet.LogEntry = widget.NewMultiLineEntry()
	Wallet.LogEntry.Disable()

	if tag == "dReams" || tag == "NFA Market" {
		text := fmt.Sprintf("[%s] %s  OS: %s  ARCH: %s", tag, Version(), runtime.GOOS, runtime.GOARCH)
		logger.Println(text)
		Wallet.LogEntry.SetText(fmt.Sprintf("%s  %s", time.Now().Format("2006/01/02 15:04:05"), text))
	} else {
		text := fmt.Sprintf("[%s] %s  OS: %s  ARCH: %s  DREAMS: %s", tag, dapp.String(), runtime.GOOS, runtime.GOARCH, Version())
		logger.Println(text)
		Wallet.LogEntry.SetText(fmt.Sprintf("%s  %s", time.Now().Format("2006/01/02 15:04:05"), text))
	}

	button := widget.NewButton("Save", func() {
		file_name := fmt.Sprintf("Log-%s", time.Now().Format("2006-01-02-15:04:05"))
		if f, err := os.Create(file_name); err == nil {
			defer f.Close()
			if _, err = f.WriteString(Wallet.LogEntry.Text); err != nil {
				logger.Errorln("[saveLog]", err)
				return
			}

			logger.Println("[saveLog] Log File Saved", file_name)
		} else {
			logger.Errorln("[saveLog]", err)
		}
	})
	button.Importance = widget.LowImportance

	pad := layout.NewSpacer()
	cont := container.NewStack(Wallet.LogEntry)
	vbox := container.NewVBox(
		layout.NewSpacer(),
		container.NewHBox(pad, container.NewBorder(pad, pad, pad, pad, button)))

	return container.NewStack(cont, vbox)
}

// Initialize balance maps for supported tokens
func InitBalances() {
	Wallet.TokenBal = make(map[string]uint64)
	Wallet.Display.Balance = make(map[string]string)
	SCIDs = make(map[string]string)
	SCIDs["dReams"] = DreamsSCID
	SCIDs["HGC"] = HgcSCID
	SCIDs["TRVL"] = TrvlSCID
	Wallet.Display.Balance["Dero"] = "0"
	Wallet.Display.Balance["dReams"] = "0"
	Wallet.Display.Balance["HGC"] = "0"
	Wallet.Display.Balance["TRVL"] = "0"
}

// Set wallet rpc client with auth, context and 6 sec cancel
func SetWalletClient(addr, pass string) (jsonrpc.RPCClient, context.Context, context.CancelFunc) {
	client := jsonrpc.NewClientWithOpts("http://"+addr+"/json_rpc", &jsonrpc.RPCClientOpts{
		CustomHeaders: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(pass)),
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)

	return client, ctx, cancel
}

// Echo Dero wallet for connection
//   - tag for log print
func EchoWallet(tag string) {
	if Wallet.IsConnected() {
		rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
		defer cancel()

		var result string
		params := []string{"Hello", "World", "!"}
		if err := rpcClientW.CallFor(ctx, &result, "Echo", params); err != nil {
			Wallet.Connected(false)
			PrintError("[%s] %s", tag, err)
			return
		}

		if result != "WALLET Hello World !" {
			Wallet.Connected(false)
		}
	}
}

// Get a wallets Dero address
//   - tag for log print
func GetAddress(tag string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	var result *rpc.GetAddress_Result
	if err := rpcClientW.CallFor(ctx, &result, "GetAddress"); err != nil {
		Wallet.Connected(false)
		PrintError("[%s] %s", tag, err)
		return
	}

	if (result.Address[0:4] == "dero" || result.Address[0:4] == "deto") && len(result.Address) == 66 {
		Wallet.Connected(true)
		PrintLog("[%s] Wallet Connected: %s", tag, result.Address)
		Wallet.Address = result.Address
		id := []byte(result.Address)
		hash := sha256.Sum256(id)
		Wallet.IdHash = hex.EncodeToString(hash[:])
	} else {
		Wallet.Connected(false)
	}
}

// Get wallet tx entry data by txid
func GetWalletTx(txid string) *rpc.Entry {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	var result *rpc.Get_Transfer_By_TXID_Result
	params := rpc.Get_Transfer_By_TXID_Params{
		TXID: txid,
	}

	if err := rpcClientW.CallFor(ctx, &result, "GetTransferbyTXID", params); err != nil {
		logger.Errorln("[GetWalletTx]", err)
		return nil
	}

	return &result.Entry
}

// Get wallet transfers with min/max heights and dst port
func GetWalletTransfers(min, max, dst uint64) *[]rpc.Entry {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	var result *rpc.Get_Transfers_Result
	params := rpc.Get_Transfers_Params{
		Coinbase:        false,
		In:              true,
		Out:             false,
		Min_Height:      min,
		Max_Height:      max,
		DestinationPort: dst,
	}

	if err := rpcClientW.CallFor(ctx, &result, "GetTransfers", params); err != nil {
		logger.Errorln("[GetWalletTx]", err)
		return nil
	}

	return &result.Entries
}

// Returns Dero wallet balance
func GetBalance() uint64 {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	var result *rpc.GetBalance_Result
	if err := rpcClientW.CallFor(ctx, &result, "GetBalance"); err != nil {
		logger.Errorln("[GetBalance]", err)
		return 0
	}

	return result.Unlocked_Balance
}

// Returns wallet balance of token by SCID
func TokenBalance(scid string) uint64 {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	var result *rpc.GetBalance_Result
	params := &rpc.GetBalance_Params{
		SCID: crypto.HashHexToHash(scid),
	}

	if err := rpcClientW.CallFor(ctx, &result, "GetBalance", params); err != nil {
		logger.Errorln("[TokenBalance]", err)
		return 0
	}

	return result.Unlocked_Balance
}

// Get Dero balance and all tokens used on dReams platform
func GetDreamsBalances(assets map[string]string) {
	Wallet.MuB.Lock()
	defer Wallet.MuB.Unlock()

	if Wallet.IsConnected() {
		bal := GetBalance()
		Wallet.Balance = bal
		Wallet.Display.Balance["Dero"] = FromAtomic(bal, 5)

		for name, sc := range assets {
			token_bal := TokenBalance(sc)
			Wallet.Display.Balance[name] = FromAtomic(decimal(name, token_bal))
			Wallet.TokenBal[name] = token_bal
		}

		return
	}

	Wallet.Display.Balance["Dero"] = "0"
	Wallet.Balance = 0
	for name := range assets {
		Wallet.Display.Balance[name] = "0"
		Wallet.TokenBal[name] = 0
	}
}

// Return Display.Balance string of name
func DisplayBalance(name string) string {
	Wallet.MuB.Lock()
	defer Wallet.MuB.Unlock()

	return Wallet.Display.Balance[name]
}

// Return asset transfer for SCID
func GetAssetSCIDforTransfer(amt uint64, scid string) (transfer rpc.Transfer) {
	switch scid {
	case DreamsSCID:
		transfer = rpc.Transfer{
			SCID:        crypto.HashHexToHash(DreamsSCID),
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Burn:        amt,
		}
	case HgcSCID:
		transfer = rpc.Transfer{
			SCID:        crypto.HashHexToHash(HgcSCID),
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Burn:        amt,
		}

	default:
		// nothing
	}

	return
}

// Get display name of asset by SCID
func GetAssetSCIDName(scid string) string {
	switch scid {
	case DreamsSCID:
		return "dReams"
	case HgcSCID:
		return "HGC"
	case TrvlSCID:
		return "TRVL"
	default:
		return ""
	}
}

// Gets Dero wallet height
//   - tag for log print
func GetWalletHeight(tag string) {
	if Wallet.IsConnected() {
		rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
		defer cancel()

		var result *rpc.GetHeight_Result
		if err := rpcClientW.CallFor(ctx, &result, "GetHeight"); err != nil {
			logger.Errorf("[%s] %s\n", tag, err)
			return
		}

		Wallet.Height = int(result.Height)
		Wallet.Display.Height = fmt.Sprint(result.Height)
	}
}

// Swap Dero for dReams
//   - amt of Der to swap for dReams
func GetdReams(amt uint64) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "IssueChips"}
	args := rpc.Arguments{arg1}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        amt,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(BaccSCID, "[Swap]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     BaccSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		PrintError("[Swap] DERO-dReams %s", err)
		return
	}

	PrintLog("[Swap] DERO-dReams TX: %s", txid)
}

// Swap dReams for Dero
//   - amt of dReams to swap for Dero
func TradedReams(amt uint64) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "ConvertChips"}
	args := rpc.Arguments{arg1}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		SCID:        crypto.HashHexToHash(DreamsSCID),
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        amt,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(BaccSCID, "[Swap]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     BaccSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		PrintError("[Swap] dReams-DERO %s", err)
		return
	}

	PrintLog("[Swap] dReams-DERO TX: %s", txid)
}

var UnlockFee = uint64(300000)
var ListingFee = uint64(10000)
var MintingFee = uint64(500)
var IlumaFee = uint64(9000)
var LowLimitFee = uint64(1320)
var HighLimitFee = uint64(10000)

// Rate a SC with dReams rating system. Ratings are weight based on transactions Dero amount
//   - amt of Dero for rating
//   - pos defines positive or negative rating
func RateSCID(scid string, amt, pos uint64) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Rate"}
	arg2 := rpc.Argument{Name: "scid", DataType: "S", Value: scid}
	arg3 := rpc.Argument{Name: "pos", DataType: "U", Value: pos}
	args := rpc.Arguments{arg1, arg2, arg3}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        amt,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(RatingSCID, "[RateSCID]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     RatingSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		PrintError("[RateSCID] %s", err)
		return
	}

	PrintLog("[RateSCID] Rate TX: %s", txid)
}

// Set any SC headers on Gnomon SC
//   - name, desc and icon are header params
func SetHeaders(name, desc, icon, scid string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "SetSCIDHeaders"}
	arg2 := rpc.Argument{Name: "name", DataType: "S", Value: name}
	arg3 := rpc.Argument{Name: "descr", DataType: "S", Value: desc}
	arg4 := rpc.Argument{Name: "icon", DataType: "S", Value: icon}
	arg5 := rpc.Argument{Name: "scid", DataType: "S", Value: scid}
	args := rpc.Arguments{arg1, arg2, arg3, arg4, arg5}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        200,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(GnomonSCID, "[SetHeaders]", args, t, HighLimitFee)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_Value:  0,
		SC_ID:     GnomonSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		PrintError("[SetHeaders] %s", err)
		return
	}

	PrintLog("[SetHeaders] Set Headers TX: %s", txid)
}

// Claim transferred NFA token
func ClaimNFA(scid string) (tx string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "ClaimOwnership"}
	args := rpc.Arguments{arg1, arg1}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		SCID:        crypto.HashHexToHash(scid),
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        1,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(scid, "[ClaimNFA]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		PrintError("[ClaimNFA] %s", err)
		return
	}

	PrintLog("[ClaimNFA] Claim TX: %s", txid)

	return txid.TXID
}

// Send bid or buy to NFA SC
//   - bidor defines bid or buy call
func BidBuyNFA(scid, bidor string, amt uint64) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: bidor}
	args := rpc.Arguments{arg1}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        amt,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(scid, "[BidBuyNFA]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		PrintError("[BidBuyNFA] %s", err)
		return
	}

	if bidor == "Bid" {
		PrintLog("[BidBuyNFA] NFA Bid TX: %s", txid)
	} else {
		PrintLog("[BidBuyNFA] NFA Buy TX: %s", txid)
	}
}

// List NFA for auction or sale by SCID
//   - list defines type of listing
//   - char sets charity donation address
//   - dur sets listing duration
//   - amt sets starting price
//   - perc sets percentage to go to charity on sale
func SetNFAListing(scid, list, char string, dur, amt, perc uint64) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Start"}
	arg2 := rpc.Argument{Name: "listType", DataType: "S", Value: strings.ToLower(list)}
	arg3 := rpc.Argument{Name: "duration", DataType: "U", Value: dur}
	arg4 := rpc.Argument{Name: "startPrice", DataType: "U", Value: amt}
	arg5 := rpc.Argument{Name: "charityDonateAddr", DataType: "S", Value: char}
	arg6 := rpc.Argument{Name: "charityDonatePerc", DataType: "U", Value: perc}
	args := rpc.Arguments{arg1, arg2, arg3, arg4, arg5, arg6}
	txid := rpc.Transfer_Result{}

	split_fee := ListingFee / 2

	t1 := rpc.Transfer{
		SCID:        crypto.HashHexToHash(scid),
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        1,
	}

	/// dReams
	t2 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      split_fee,
		Burn:        0,
	}

	/// artificer
	t3 := rpc.Transfer{
		Destination: "dero1qy0khp9s9yw2h0eu20xmy9lth3zp5cacmx3rwt6k45l568d2mmcf6qgcsevzx",
		Amount:      split_fee,
		Burn:        0,
	}

	t := []rpc.Transfer{t1, t2, t3}
	fee := GasEstimate(scid, "[SetNFAListing]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		PrintError("[SetNFAListing] %s", err)
		return
	}

	PrintLog("[SetNFAListing] NFA List TX: %s", txid)
}

// Cancel or close a listed NFA. Can only be canceled within opening buffer period. Can only close listing after expiry
//   - c defines cancel or close call
func CancelCloseNFA(scid, c string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: c}
	args := rpc.Arguments{arg1}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(scid, "[CancelCloseNFA]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		PrintError("[CancelCloseNFA] %s", err)
		return
	}

	if c == "CloseListing" {
		PrintLog("[CancelCloseNFA] Close Listing TX: %s", txid)
	} else {
		PrintLog("[CancelCloseNFA] Cancel Listing TX: %s", txid)
	}
}

// Upload a new NFA SC by string
func UploadNFAContract(code string) (tx string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	txid := rpc.Transfer_Result{}
	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      MintingFee,
	}

	params := &rpc.Transfer_Params{
		Transfers: []rpc.Transfer{t1},
		SC_Code:   code,
		SC_Value:  0,
		SC_RPC:    rpc.Arguments{},
		Ringsize:  2,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		PrintError("[UploadNFAContract] %s", err)
		return
	}

	PrintLog("[UploadNFAContract] TXID: %s", txid)

	return txid.TXID
}

// Send Dero asset to destination address with option to send asset SCID as message to destination as payload
func SendAsset(scid, dest string, payload bool) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	t1 := rpc.Transfer{
		SCID:        crypto.HashHexToHash(scid),
		Destination: dest,
		Amount:      1,
	}

	t := []rpc.Transfer{t1}

	if payload {
		var dstport [8]byte
		rand.Read(dstport[:])

		response := rpc.Arguments{
			{Name: rpc.RPC_DESTINATION_PORT, DataType: rpc.DataUint64, Value: binary.BigEndian.Uint64(dstport[:])},
			{Name: rpc.RPC_SOURCE_PORT, DataType: rpc.DataUint64, Value: uint64(0)},
			{Name: rpc.RPC_COMMENT, DataType: rpc.DataString, Value: fmt.Sprintf("Sent you asset  %s  at height %d", scid, Wallet.Height)},
		}

		t2 := rpc.Transfer{
			Destination: dest,
			Amount:      1,
			Burn:        0,
			Payload_RPC: response,
		}
		t = append(t, t2)
	}

	txid := rpc.Transfer_Result{}

	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_RPC:    rpc.Arguments{},
		Ringsize:  16,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		PrintError("[SendAsset] %s", err)
		return
	}

	PrintLog("[SendAsset] Send Asset TX: %s", txid)
}

// Watch a sent tx and return true if tx is confirmed
//   - tag for log print
//   - timeout is duration of loop in 2sec increment, will break if reached
func ConfirmTx(txid, tag string, timeout int) bool {
	if txid != "" {
		count := 0
		time.Sleep(time.Second)
		for IsReady() {
			count++
			time.Sleep(2 * time.Second)
			if tx := GetDaemonTx(txid); tx != nil {
				if count > timeout {
					break
				}

				if tx.In_pool {
					continue
				} else if !tx.In_pool && tx.Block_Height > 1 && tx.ValidBlock != "" {
					logger.Printf("[%s] TX Confirmed\n", tag)
					return true
				} else if !tx.In_pool && tx.Block_Height == 0 && tx.ValidBlock == "" {
					logger.Warnf("[%s] TX Failed\n", tag)
					return false
				}
			}
		}
	}

	logger.Errorf("[%s] Could Not Confirm TX\n", tag)

	return false
}

// Watch a sent tx with int return for retry count, failed tx returns 1, timeout returns 2
//   - tag for log print
//   - timeout is duration of loop in 2sec increment, will break if reached
func ConfirmTxRetry(txid, tag string, timeout int) (retry int) {
	count := 0
	next_block := Wallet.Height + 1
	time.Sleep(time.Second)
	for IsReady() {
		count++
		time.Sleep(2 * time.Second)
		if tx := GetDaemonTx(txid); tx != nil {
			if count > timeout {
				break
			}

			if tx.In_pool {
				continue
			} else if !tx.In_pool && tx.Block_Height > 1 && tx.ValidBlock != "" {
				logger.Printf("[%s] TX Confirmed\n", tag)
				return 100
			} else if !tx.In_pool && tx.Block_Height == 0 && tx.ValidBlock == "" {
				logger.Warnf("[%s] TX Failed, Retrying next block\n", tag)
				time.Sleep(3 * time.Second)
				for Wallet.Height <= next_block {
					time.Sleep(3 * time.Second)
				}
				return 1
			}
		}
	}

	logger.Errorf("[%s] Could Not Confirm TX\n", tag)

	return 2
}

// Send a message to destination address through Dero transaction, with ringsize selection
func SendMessage(dest, msg string, rings uint64) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	response := rpc.Arguments{
		{Name: rpc.RPC_DESTINATION_PORT, DataType: rpc.DataUint64, Value: 1337},
		{Name: rpc.RPC_SOURCE_PORT, DataType: rpc.DataUint64, Value: uint64(0)},
		{Name: rpc.RPC_COMMENT, DataType: rpc.DataString, Value: msg},
	}

	t1 := rpc.Transfer{
		Destination: dest,
		Amount:      1,
		Burn:        0,
		Payload_RPC: response,
	}

	t := []rpc.Transfer{t1}
	txid := rpc.Transfer_Result{}
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_RPC:    rpc.Arguments{},
		Ringsize:  rings,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		PrintError("[SendMessage] %s", err)
		return
	}

	PrintLog("[SendMessage] Send Message TX: %s", txid)
}

// TODO should put decimal in Wallet.Display

func decimal(name string, bal uint64) (uint64, int) {
	if name == "TRVL" {
		return bal * 100000, 0
	}

	return bal, 5
}
