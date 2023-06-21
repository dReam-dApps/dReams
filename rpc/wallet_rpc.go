package rpc

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/walletapi"
	"github.com/ybbus/jsonrpc/v3"
)

// Add entry to Wallet.LogEntry session log
func AddLog(t string) {
	if Wallet.LogEntry != nil {
		Wallet.LogEntry.SetText(Wallet.LogEntry.Text + "\n\n" + t)
		Wallet.LogEntry.Refresh()
	}
}

// Make gui log for txs with save function
func SessionLog() *fyne.Container {
	Wallet.LogEntry = widget.NewMultiLineEntry()
	Wallet.LogEntry.Disable()
	button := widget.NewButton("Save", func() {
		file_name := fmt.Sprintf("Log-%s", time.Now().Format(time.UnixDate))
		if f, err := os.Create(file_name); err == nil {
			defer f.Close()
			if _, err = f.WriteString(Wallet.LogEntry.Text); err != nil {
				log.Println("[saveLog]", err)
				return
			}

			log.Println("[saveLog] Log File Saved", file_name)
		} else {
			log.Println("[saveLog]", err)
		}
	})
	button.Importance = widget.LowImportance

	cont := container.NewMax(Wallet.LogEntry)
	vbox := container.NewVBox(
		layout.NewSpacer(),
		container.NewBorder(nil, layout.NewSpacer(), nil, button, layout.NewSpacer()))

	return container.NewMax(cont, vbox)
}

func InitBalances() {
	Wallet.TokenBal = make(map[string]uint64)
	Display.Balance = make(map[string]string)
	Display.Balance["Dero"] = "0"
	Display.Balance["dReams"] = "0"
	Display.Balance["HGC"] = "0"
	Display.Balance["TRVL"] = "0"
	Signal.Sit = true
}

// Set wallet rpc client with auth, context and 5 sec cancel
func SetWalletClient(addr, pass string) (jsonrpc.RPCClient, context.Context, context.CancelFunc) {
	client := jsonrpc.NewClientWithOpts("http://"+addr+"/json_rpc", &jsonrpc.RPCClientOpts{
		CustomHeaders: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(pass)),
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

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
			log.Printf("[%s] %s\n", tag, err)
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
		log.Printf("[%s] %s\n", tag, err)
		return
	}

	if (result.Address[0:4] == "dero" || result.Address[0:4] == "deto") && len(result.Address) == 66 {
		Wallet.Connected(true)
		log.Printf("[%s] Wallet Connected\n", tag)
		log.Printf("[%s] Dero Address: %s\n", tag, result.Address)
		Wallet.Address = result.Address
		id := []byte(result.Address)
		hash := sha256.Sum256(id)
		Wallet.idHash = hex.EncodeToString(hash[:])
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
		log.Println("[GetWalletTx]", err)
		return nil
	}

	return &result.Entry
}

// Returns Dero wallet balance
func GetBalance() uint64 {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	var result *rpc.GetBalance_Result
	if err := rpcClientW.CallFor(ctx, &result, "GetBalance"); err != nil {
		log.Println("[GetBalance]", err)
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
		log.Println("[TokenBalance]", err)
		return 0
	}

	return result.Unlocked_Balance
}

// Get Dero balance and all tokens used on dReams platform
func GetDreamsBalances() {
	Wallet.MuB.Lock()
	defer Wallet.MuB.Unlock()

	if Wallet.IsConnected() {
		bal := GetBalance()
		Wallet.Balance = bal
		Display.Balance["Dero"] = FromAtomic(bal, 5)

		dReam_bal := TokenBalance(DreamsSCID)
		Display.Balance["dReams"] = FromAtomic(dReam_bal, 5)
		Wallet.TokenBal["dReams"] = dReam_bal

		trvl_bal := TokenBalance(TrvlSCID)
		Display.Balance["TRVL"] = strconv.Itoa(int(trvl_bal))
		Wallet.TokenBal["TRVL"] = trvl_bal

		hgc_bal := TokenBalance(HgcSCID)
		Display.Balance["HGC"] = FromAtomic(hgc_bal, 5)
		Wallet.TokenBal["HGC"] = hgc_bal

		if Round.Tourney {
			tourney_bal := TokenBalance(TourneySCID)
			Display.Balance["Tournament"] = FromAtomic(tourney_bal, 5)
			Wallet.TokenBal["Tournament"] = tourney_bal
		}

		return
	}

	Display.Balance["Dero"] = "0"
	Wallet.Balance = 0
	Display.Balance["dReams"] = "0"
	Wallet.TokenBal["dReams"] = 0
	Display.Balance["TRVL"] = "0"
	Wallet.TokenBal["TRVL"] = 0
	Display.Balance["HGC"] = "0"
	Wallet.TokenBal["HGC"] = 0
	Display.Balance["Tournament"] = "0"
	Wallet.TokenBal["Tournament"] = 0
}

// Return Display.Balance string of name
func DisplayBalance(name string) string {
	Wallet.MuB.Lock()
	defer Wallet.MuB.Unlock()

	return Display.Balance[name]
}

// Return asset transfer to SCID from Round.AssetID
func GetAssetSCIDforTransfer(amt uint64) (transfer rpc.Transfer) {
	switch Round.AssetID {
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
	default:
		return ""
	}
}

// Deposit tournament chip bal with name to leader board SC
func TourneyDeposit(bal uint64, name string) {
	if bal > 0 {
		rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
		defer cancel()

		arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Deposit"}
		arg2 := rpc.Argument{Name: "name", DataType: "S", Value: name}
		args := rpc.Arguments{arg1, arg2}
		txid := rpc.Transfer_Result{}

		t1 := rpc.Transfer{
			SCID:        crypto.HashHexToHash(TourneySCID),
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      0,
			Burn:        bal,
		}

		t := []rpc.Transfer{t1}
		fee := GasEstimate(TourneySCID, "[Holdero]", args, t, LowLimitFee)
		params := &rpc.Transfer_Params{
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
		AddLog("Tournament Deposit TX: " + txid.TXID)

	} else {
		log.Println("[Holdero] No Tournament Chips")
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
			log.Printf("[%s] %s\n", tag, err)
			return
		}

		Wallet.Height = int(result.Height)
		Display.Wallet_height = fmt.Sprint(result.Height)
	}
}

// Submit playerId, name, avatar and sit at Holdero table
//   - name and av are for name and avatar in player id string
func SitDown(name, av string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	var player playerId
	player.Id = Wallet.idHash
	player.Name = name
	player.Avatar = av

	mar, _ := json.Marshal(player)
	hx := hex.EncodeToString(mar)

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "PlayerEntry"}
	arg2 := rpc.Argument{Name: "address", DataType: "S", Value: hx}
	args := rpc.Arguments{arg1, arg2}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t, HighLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Sit Down TX: " + txid.TXID)
}

// Leave Holdero seat on players turn
func Leave() {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	checkoutId := StringToInt(Display.PlayerId)
	singleNameClear(checkoutId)
	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "PlayerLeave"}
	arg2 := rpc.Argument{Name: "id", DataType: "U", Value: checkoutId}
	args := rpc.Arguments{arg1, arg2}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Leave Down TX: " + txid.TXID)
}

// Owner table settings for Holdero
//   - seats defines max players at table
//   - bb, sb and ante define big blind, small blind and antes. Ante can be 0
//   - chips defines if tables is using Dero or assets
//   - name and av are for name and avatar in owners id string
func SetTable(seats int, bb, sb, ante uint64, chips, name, av string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	var player playerId
	player.Id = Wallet.idHash
	player.Name = name
	player.Avatar = av

	mar, _ := json.Marshal(player)
	hx := hex.EncodeToString(mar)

	var args rpc.Arguments
	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "SetTable"}
	arg2 := rpc.Argument{Name: "seats", DataType: "U", Value: seats}
	arg3 := rpc.Argument{Name: "bigBlind", DataType: "U", Value: bb}
	arg4 := rpc.Argument{Name: "smallBlind", DataType: "U", Value: sb}
	arg5 := rpc.Argument{Name: "ante", DataType: "U", Value: ante}
	arg6 := rpc.Argument{Name: "address", DataType: "S", Value: hx}
	txid := rpc.Transfer_Result{}

	if Round.Version < 110 {
		args = rpc.Arguments{arg1, arg2, arg3, arg4, arg5, arg6}
	} else if Round.Version == 110 {
		arg7 := rpc.Argument{Name: "chips", DataType: "S", Value: chips}
		args = rpc.Arguments{arg1, arg2, arg3, arg4, arg5, arg6, arg7}
	}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t, HighLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Set Table TX: " + txid.TXID)
}

// Submit blinds/ante to deal Holdero hand
func DealHand() (tx string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	if !Wallet.KeyLock {
		Wallet.ClientKey = GenerateKey()
	}

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "DealHand"}
	arg2 := rpc.Argument{Name: "pcSeed", DataType: "H", Value: Wallet.ClientKey}
	args := rpc.Arguments{arg1, arg2}
	txid := rpc.Transfer_Result{}

	var amount uint64

	if Round.Pot == 0 {
		amount = Round.Ante + Round.SB
	} else if Round.Pot == Round.SB || Round.Pot == Round.Ante+Round.SB {
		amount = Round.Ante + Round.BB
	} else {
		amount = Round.Ante
	}

	t := []rpc.Transfer{}
	if Round.Asset {
		t1 := rpc.Transfer{
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      500,
			Burn:        0,
		}

		if Round.Tourney {
			t2 := rpc.Transfer{
				SCID:        crypto.HashHexToHash(TourneySCID),
				Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
				Burn:        amount,
			}
			t = append(t, t1, t2)
		} else {
			t2 := GetAssetSCIDforTransfer(amount)
			if t2.Destination == "" {
				log.Println("[DealHand] Error getting asset SCID for transfer")
				return
			}
			t = append(t, t1, t2)
		}
	} else {
		t1 := rpc.Transfer{
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      500,
			Burn:        amount,
		}
		t = append(t, t1)
	}

	fee := GasEstimate(Round.Contract, "[Holdero]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Deal TX: " + txid.TXID)

	return txid.TXID
}

// Make Holdero bet
func Bet(amt string) (tx string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Bet"}
	args := rpc.Arguments{arg1}
	txid := rpc.Transfer_Result{}

	var t1 rpc.Transfer
	if Round.Asset {
		if Round.Tourney {
			t1 = rpc.Transfer{
				SCID:        crypto.HashHexToHash(TourneySCID),
				Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
				Burn:        ToAtomic(amt, 1),
			}
		} else {
			t1 = GetAssetSCIDforTransfer(ToAtomic(amt, 1))
			if t1.Destination == "" {
				log.Println("[Bet] Error getting asset SCID for transfer")
				return
			}
		}
	} else {
		t1 = rpc.Transfer{
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      0,
			Burn:        ToAtomic(amt, 1),
		}
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Bet TX: " + txid.TXID)

	return txid.TXID
}

// Holdero check and fold
func Check() (tx string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Bet"}
	args := rpc.Arguments{arg1}
	txid := rpc.Transfer_Result{}

	var t1 rpc.Transfer
	if !Round.Asset {
		t1 = rpc.Transfer{
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      0,
			Burn:        0,
		}
	} else {
		if Round.Tourney {
			t1 = rpc.Transfer{
				SCID:        crypto.HashHexToHash(TourneySCID),
				Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
				Burn:        0,
			}
		} else {
			t1 = GetAssetSCIDforTransfer(0)
			if t1.Destination == "" {
				log.Println("[Check] Error getting asset SCID for transfer")
				return
			}
		}
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Check/Fold TX: " + txid.TXID)

	return txid.TXID
}

// Holdero single winner payout
//   - w defines which player the pot is going to
func PayOut(w string) string {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Winner"}
	arg2 := rpc.Argument{Name: "whoWon", DataType: "S", Value: w}
	args := rpc.Arguments{arg1, arg2}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Holdero Payout TX: " + txid.TXID)

	return txid.TXID
}

// Holdero split winners payout
//   - Pass in ranker from hand and folded bools to determine split
func PayoutSplit(r ranker, f1, f2, f3, f4, f5, f6 bool) string {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
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

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "SplitWinner"}
	arg2 := rpc.Argument{Name: "div", DataType: "U", Value: ways}
	arg3 := rpc.Argument{Name: "split1", DataType: "S", Value: splitWinners[0]}
	arg4 := rpc.Argument{Name: "split2", DataType: "S", Value: splitWinners[1]}
	arg5 := rpc.Argument{Name: "split3", DataType: "S", Value: splitWinners[2]}
	arg6 := rpc.Argument{Name: "split4", DataType: "S", Value: splitWinners[3]}
	arg7 := rpc.Argument{Name: "split5", DataType: "S", Value: splitWinners[4]}
	arg8 := rpc.Argument{Name: "split6", DataType: "S", Value: splitWinners[5]}

	args := rpc.Arguments{arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Split Winner TX: " + txid.TXID)

	return txid.TXID
}

// Reveal Holdero hand key for showdown
func RevealKey(key string) {
	time.Sleep(6 * time.Second)
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "RevealKey"}
	arg2 := rpc.Argument{Name: "pcSeed", DataType: "H", Value: key}
	args := rpc.Arguments{arg1, arg2}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Reveal TX: " + txid.TXID)
}

// Owner can shuffle deck for Holdero, clean above 0 can retrieve balance
func CleanTable(amt uint64) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "CleanTable"}
	arg2 := rpc.Argument{Name: "amount", DataType: "U", Value: amt}
	args := rpc.Arguments{arg1, arg2}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Clean Table TX: " + txid.TXID)
}

// Owner can timeout a player at Holdero table
func TimeOut() {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "TimeOut"}
	args := rpc.Arguments{arg1}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Timeout TX: " + txid.TXID)
}

// Owner can force start a Holdero table with empty seats
func ForceStat() {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "ForceStart"}
	args := rpc.Arguments{arg1}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Force Start TX: " + txid.TXID)
}

// Share asset url at Holdero table
//   - face and back are the names of assets
//   - faceUrl and backUrl are the Urls for those assets
func SharedDeckUrl(face, faceUrl, back, backUrl string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	var cards CardSpecs
	if face != "" && back != "" {
		cards.Faces.Name = face
		cards.Faces.Url = faceUrl
		cards.Backs.Name = back
		cards.Backs.Url = backUrl
	}

	if mar, err := json.Marshal(cards); err == nil {
		arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Deck"}
		arg2 := rpc.Argument{Name: "face", DataType: "S", Value: string(mar)}
		arg3 := rpc.Argument{Name: "back", DataType: "S", Value: "nil"}
		args := rpc.Arguments{arg1, arg2, arg3}
		txid := rpc.Transfer_Result{}

		t1 := rpc.Transfer{
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      0,
			Burn:        0,
		}

		t := []rpc.Transfer{t1}
		fee := GasEstimate(Round.Contract, "[Holdero]", args, t, LowLimitFee)
		params := &rpc.Transfer_Params{
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
		AddLog("Shared TX: " + txid.TXID)
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
	fee := GasEstimate(BaccSCID, "[dReams]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     BaccSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[GetdReams]", err)
		return
	}

	log.Println("[dReams] DERO-dReams", txid)
	AddLog("DERO-dReams " + txid.TXID)
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
	fee := GasEstimate(BaccSCID, "[dReams]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     BaccSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[TradedReams]", err)
		return
	}

	log.Println("[dReams] dReams-DERO TX:", txid)
	AddLog("dReams-DERO TX: " + txid.TXID)
}

var UnlockFee = uint64(300000)
var ListingFee = uint64(10000)
var MintingFee = uint64(500)
var IlumaFee = uint64(9000)
var LowLimitFee = uint64(1320)
var HighLimitFee = uint64(10000)

// Contract unlock transfer
func ownerT3(o bool) (t *rpc.Transfer) {
	if o {
		t = &rpc.Transfer{
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      0,
		}
	} else {
		t = &rpc.Transfer{
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      UnlockFee,
		}
	}

	return
}

// Install new Holdero SC
//   - pub defines public or private SC
func UploadHolderoContract(pub int) {
	if IsReady() {
		rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
		defer cancel()

		code := GetHoldero110Code(pub)
		if code == "" {
			log.Println("[UploadHolderoContract] Could not get SC code")
			return
		}

		args := rpc.Arguments{}
		txid := rpc.Transfer_Result{}

		params := &rpc.Transfer_Params{
			Transfers: []rpc.Transfer{*ownerT3(Wallet.PokerOwner)},
			SC_Code:   code,
			SC_Value:  0,
			SC_RPC:    args,
			Ringsize:  2,
			Fees:      30000,
		}

		if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
			log.Println("[UploadHolderoContract]", err)
			return
		}

		log.Println("[Holdero] Upload TX:", txid)
		AddLog("Holdero Upload TX:" + txid.TXID)
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

	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "PlayBaccarat"}
	arg2 := rpc.Argument{Name: "betOn", DataType: "S", Value: w}
	args := rpc.Arguments{arg1, arg2}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		SCID:        crypto.HashHexToHash(Bacc.AssetID),
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        ToAtomic(amt, 1),
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(Bacc.Contract, "[Baccarat]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
		AddLog("Baccarat Player TX: " + txid.TXID)
	} else if w == "banker" {
		log.Println("[Baccarat] Banker TX:", txid)
		AddLog("Baccarat Banker TX: " + txid.TXID)
	} else {
		log.Println("[Baccarat] Tie TX:", txid)
		AddLog("Baccarat Tie TX: " + txid.TXID)
	}

	Bacc.CHeight = Wallet.Height

	return txid.TXID
}

// Place higher prediction to SC
//   - addr only needed if dReamService is placing prediction
func PredictHigher(scid, addr string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	amt := uint64(Predict.Amount)

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Predict"}
	arg2 := rpc.Argument{Name: "pre", DataType: "U", Value: 1}
	arg3 := rpc.Argument{Name: "addr", DataType: "S", Value: addr}
	args := rpc.Arguments{arg1, arg2, arg3}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        amt,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(scid, "[Predictions]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Prediction TX: " + txid.TXID)
}

// Place lower prediction to SC
//   - addr only needed if dReamService is placing prediction
func PredictLower(scid, addr string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	amt := uint64(Predict.Amount)

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Predict"}
	arg2 := rpc.Argument{Name: "pre", DataType: "U", Value: 0}
	arg3 := rpc.Argument{Name: "addr", DataType: "S", Value: addr}
	args := rpc.Arguments{arg1, arg2, arg3}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        amt,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(scid, "[Predictions]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Prediction TX: " + txid.TXID)
}

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
		log.Println("[RateSCID]", err)
		return
	}

	log.Println("[RateSCID] Rate TX:", txid)
	AddLog("Rate TX: " + txid.TXID)
}

// dReamService prediction place by received tx
//   - amt to send
//   - p is what prediction
//   - addr of placed bet and to send reply message
//   - src and pre_tx used in reply message
func AutoPredict(p int, amt, src uint64, scid, addr, pre_tx string) (tx string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
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

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Predict"}
	arg2 := rpc.Argument{Name: "pre", DataType: "U", Value: p}
	arg3 := rpc.Argument{Name: "addr", DataType: "S", Value: addr}
	args := rpc.Arguments{arg1, arg2, arg3}
	txid := rpc.Transfer_Result{}

	response := rpc.Arguments{
		{Name: rpc.RPC_DESTINATION_PORT, DataType: rpc.DataUint64, Value: uint64(0)},
		{Name: rpc.RPC_SOURCE_PORT, DataType: rpc.DataUint64, Value: src},
		{Name: rpc.RPC_COMMENT, DataType: rpc.DataString, Value: fmt.Sprintf("Placed a %s %s bet on %s at height %s,  %s", walletapi.FormatMoney(amt), hl, chopped_scid, Display.Wallet_height, chopped_txid)},
	}

	t1 := rpc.Transfer{
		Destination: addr,
		Amount:      1,
		Burn:        amt,
		Payload_RPC: response,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(scid, "[AutoPredict]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("AutoPredict TX: " + txid.TXID)

	return txid.TXID
}

// dReamService refund if bet void
//   - amt to send
//   - addr to send refund to
//   - src, msg and refund_tx used in reply message
func ServiceRefund(amt, src uint64, scid, addr, msg, refund_tx string) (tx string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	chopped_scid := scid[:6] + "..." + scid[58:]
	chopped_txid := refund_tx[:6] + "..." + refund_tx[58:]
	response := rpc.Arguments{
		{Name: rpc.RPC_DESTINATION_PORT, DataType: rpc.DataUint64, Value: uint64(0)},
		{Name: rpc.RPC_SOURCE_PORT, DataType: rpc.DataUint64, Value: src},
		{Name: rpc.RPC_COMMENT, DataType: rpc.DataString, Value: msg + fmt.Sprintf(", refunded %s bet on %s at height %s,  %s", walletapi.FormatMoney(amt), chopped_scid, Display.Wallet_height, chopped_txid)},
	}

	t1 := rpc.Transfer{
		Destination: addr,
		Amount:      amt,
		Burn:        0,
		Payload_RPC: response,
	}

	txid := rpc.Transfer_Result{}
	t := []rpc.Transfer{t1}
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_RPC:    rpc.Arguments{},
		Ringsize:  16,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[ServiceRefund]", err)
		return
	}

	log.Println("[ServiceRefund] Refund TX:", txid)
	AddLog("Refund TX: " + txid.TXID)

	return txid.TXID
}

// dReamService sports book by received tx
//   - amt to send
//   - pre is what team
//   - n is the game number
//   - addr of placed bet and to send reply message
//   - src, abv and tx used in reply message
func AutoBook(amt, pre, src uint64, n, abv, scid, addr, book_tx string) (tx string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	chopped_scid := scid[:6] + "..." + scid[58:]
	chopped_txid := book_tx[:6] + "..." + book_tx[58:]
	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Book"}
	arg2 := rpc.Argument{Name: "pre", DataType: "U", Value: pre}
	arg3 := rpc.Argument{Name: "n", DataType: "S", Value: n}
	arg4 := rpc.Argument{Name: "addr", DataType: "S", Value: addr}
	args := rpc.Arguments{arg1, arg2, arg3, arg4}
	txid := rpc.Transfer_Result{}

	response := rpc.Arguments{
		{Name: rpc.RPC_DESTINATION_PORT, DataType: rpc.DataUint64, Value: uint64(0)},
		{Name: rpc.RPC_SOURCE_PORT, DataType: rpc.DataUint64, Value: src},
		{Name: rpc.RPC_COMMENT, DataType: rpc.DataString, Value: fmt.Sprintf("Placed a %s %s bet on %s at height %s,  %s", walletapi.FormatMoney(amt), abv, chopped_scid, Display.Wallet_height, chopped_txid)},
	}

	t1 := rpc.Transfer{
		Destination: addr,
		Amount:      1,
		Burn:        amt,
		Payload_RPC: response,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(scid, "[AutoBook]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("AutoBook TX: " + txid.TXID)

	return txid.TXID
}

// Owner update for bet SC vars
//   - ta, tb, tc are contracts time limits. Only ta, tb needed for dSports
//   - l is the max bet limit per initialized bet
//   - hl is the max amount of games that can be ran at once
func VarUpdate(scid string, ta, tb, tc, l, hl int) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "VarUpdate"}
	arg2 := rpc.Argument{Name: "ta", DataType: "U", Value: ta}
	arg3 := rpc.Argument{Name: "tb", DataType: "U", Value: tb}
	arg5 := rpc.Argument{Name: "l", DataType: "U", Value: l}

	var args rpc.Arguments
	var arg4, arg6 rpc.Argument
	if hl > 0 {
		arg4 = rpc.Argument{Name: "d", DataType: "U", Value: tc}
		arg6 = rpc.Argument{Name: "hl", DataType: "U", Value: hl}
		args = rpc.Arguments{arg1, arg2, arg3, arg4, arg5, arg6}
	} else {
		arg4 = rpc.Argument{Name: "tc", DataType: "U", Value: tc}
		args = rpc.Arguments{arg1, arg2, arg3, arg4, arg5}
	}

	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(scid, "[VarUpdate]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("VarUpdate TX: " + txid.TXID)
}

// Owner can add new co-owner to bet SC
//   - addr of new co-owner
func AddOwner(scid, addr string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "AddSigner"}
	arg2 := rpc.Argument{Name: "new", DataType: "S", Value: addr}
	args := rpc.Arguments{arg1, arg2}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(scid, "[Predictions]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Add Signer TX: " + txid.TXID)
}

// Owner can remove co-owner from bet SC
//   - num defines which co-owner to remove
func RemoveOwner(scid string, num int) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "RemoveSigner"}
	arg2 := rpc.Argument{Name: "remove", DataType: "U", Value: num}
	args := rpc.Arguments{arg1, arg2}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(scid, "[Predictions]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Remove Signer: " + txid.TXID)
}

// User can refund a void dPrediction payout from SC
//   - tic is the prediction id string
func PredictionRefund(scid, tic string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Refund"}
	arg2 := rpc.Argument{Name: "tic", DataType: "S", Value: "p-1-1"}
	args := rpc.Arguments{arg1, arg2}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(scid, "[Predictions]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Refund TX: " + txid.TXID)
}

// Book sports team on dSports SC
//   - multi defines 1x, 3x or 5x the minimum
//   - n is the game number
//   - a is amount to book
//   - pick is team to book
func PickTeam(scid, multi, n string, a uint64, pick int) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
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

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Book"}
	arg2 := rpc.Argument{Name: "n", DataType: "S", Value: n}
	arg3 := rpc.Argument{Name: "pre", DataType: "U", Value: pick}
	arg4 := rpc.Argument{Name: "addr", DataType: "S", Value: Wallet.Address}
	args := rpc.Arguments{arg1, arg2, arg3, arg4}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        amt,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(scid, "[Sports]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Pick TX: " + txid.TXID)
}

// User can refund a void dSports payout from SC
//   - tic is the bet id string
//   - n is the game number
func SportsRefund(scid, tic, n string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Refund"}
	arg2 := rpc.Argument{Name: "tic", DataType: "S", Value: tic}
	arg3 := rpc.Argument{Name: "n", DataType: "S", Value: n}
	args := rpc.Arguments{arg1, arg2, arg3}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(scid, "[Sports]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Refund TX: " + txid.TXID)
}

// Owner sets a dSports game
//   - end is unix ending time
//   - amt of single prediction
//   - dep allows owner to add a initial deposit
//   - game is name of game, formatted TEAM--TEAM
//   - feed defines where price api data is sourced from
func SetSports(end int, amt, dep uint64, scid, league, game, feed string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "S_start"}
	arg2 := rpc.Argument{Name: "end", DataType: "U", Value: end}
	arg3 := rpc.Argument{Name: "amt", DataType: "U", Value: amt}
	arg4 := rpc.Argument{Name: "league", DataType: "S", Value: league}
	arg5 := rpc.Argument{Name: "game", DataType: "S", Value: game}
	arg6 := rpc.Argument{Name: "feed", DataType: "S", Value: feed}
	args := rpc.Arguments{arg1, arg2, arg3, arg4, arg5, arg6}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        dep,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(scid, "[Sports]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Set Sports TX: " + txid.TXID)
}

// Owner sets up a dPrediction prediction
//   - end is unix ending time
//   - mark can be predefined or passed as 0 if mark is to be posted live
//   - amt of single prediction
//   - dep allows owner to add a initial deposit
//   - predict is name of what is being predicted
//   - feed defines where price api data is sourced from
func SetPrediction(end, mark int, amt, dep uint64, scid, predict, feed string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "P_start"}
	arg2 := rpc.Argument{Name: "end", DataType: "U", Value: end}
	arg3 := rpc.Argument{Name: "amt", DataType: "U", Value: amt}
	arg4 := rpc.Argument{Name: "predict", DataType: "S", Value: predict}
	arg5 := rpc.Argument{Name: "feed", DataType: "S", Value: feed}
	arg6 := rpc.Argument{Name: "mark", DataType: "U", Value: mark}
	args := rpc.Arguments{arg1, arg2, arg3, arg4, arg5, arg6}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        dep,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(scid, "[Predictions]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Set Prediction TX: " + txid.TXID)
}

// Owner cancel for initiated bet for dSports and dPrediction contracts
//   - b defines sports or prediction log print
func CancelInitiatedBet(scid string, b int) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Cancel"}
	args := rpc.Arguments{arg1}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
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

	t := []rpc.Transfer{t1}
	fee := GasEstimate(scid, tag, args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
		AddLog("Cancel Prediction TX: " + txid.TXID)
	} else {
		log.Println("[Sports] Cancel TX:", txid)
		AddLog("Cancel Sports TX: " + txid.TXID)
	}
}

// Post mark to prediction SC
//   - price is the posted mark for prediction
func PostPrediction(scid string, price int) (tx string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Post"}
	arg2 := rpc.Argument{Name: "price", DataType: "U", Value: price}
	args := rpc.Arguments{arg1, arg2}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(scid, "[Predictions]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Post TX: " + txid.TXID)

	return txid.TXID
}

// dSports SC payout
//   - num is game number
//   - team is winning team for game number
func EndSports(scid, num, team string) (tx string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "S_end"}
	arg2 := rpc.Argument{Name: "n", DataType: "S", Value: num}
	arg3 := rpc.Argument{Name: "team", DataType: "S", Value: team}
	args := rpc.Arguments{arg1, arg2, arg3}
	txid := rpc.Transfer_Result{}

	t := []rpc.Transfer{}
	fee := GasEstimate(scid, "[Sports]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Sports Payout TX: " + txid.TXID)

	return txid.TXID
}

// dPrediction SC payout
//   - price is final prediction results
func EndPrediction(scid string, price int) (tx string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "P_end"}
	arg2 := rpc.Argument{Name: "price", DataType: "U", Value: price}
	args := rpc.Arguments{arg1, arg2, arg2}
	txid := rpc.Transfer_Result{}

	t := []rpc.Transfer{}
	fee := GasEstimate(scid, "[Predictions]", args, t, LowLimitFee)
	params := &rpc.Transfer_Params{
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
	AddLog("Prediction Payout TX: " + txid.TXID)

	return txid.TXID
}

// Install new bet SC
//   - c defines dSports or dPrediction contract
//   - pub defines public or private contract
func UploadBetContract(c bool, pub int) {
	if IsReady() {
		rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
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

		args := rpc.Arguments{}
		txid := rpc.Transfer_Result{}

		params := &rpc.Transfer_Params{
			Transfers: []rpc.Transfer{*ownerT3(Wallet.BetOwner)},
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
			AddLog("Prediction Upload TX:" + txid.TXID)
		} else {
			log.Println("[Sports] Upload TX:", txid)
			AddLog("Sports Upload TX:" + txid.TXID)
		}
	}
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
		log.Println("[SetHeaders]", err)
		return
	}

	log.Println("[SetHeaders] Set Headers TX:", txid)
	AddLog("Set Headers TX: " + txid.TXID)
}

// Claim transferred NFA token
func ClaimNFA(scid string) {
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
		log.Println("[ClaimNFA]", err)
		return
	}

	log.Println("[ClaimNFA] Claim TX:", txid)
	AddLog("NFA Claim TX: " + txid.TXID)
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
		log.Println("[BidBuyNFA]", err)
		return
	}

	if bidor == "Bid" {
		log.Println("[BidBuyNFA] NFA Bid TX:", txid)
		AddLog("NFA Bid TX: " + txid.TXID)
	} else {
		log.Println("[BidBuyNFA] NFA Buy TX:", txid)
		AddLog("NFA Buy TX: " + txid.TXID)
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
		log.Println("[SetNFAListing]", err)
		return
	}

	log.Println("[SetNFAListing] NFA List TX:", txid)
	AddLog("NFA List TX: " + txid.TXID)
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
		log.Println("[CancelCloseNFA]", err)
		return
	}

	if c == "CloseListing" {
		log.Println("[CancelCloseNFA] Close NFA Listing TX:", txid)
		AddLog("NFA Close Listing TX: " + txid.TXID)
	} else {
		log.Println("[CancelCloseNFA] Cancel NFA Listing TX:", txid)
		AddLog("NFA Cancel Listing TX: " + txid.TXID)
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
		log.Println("[UploadNFAContract]", err)
		return
	}

	log.Println("[UploadNFAContract] TXID:", txid)

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
			{Name: rpc.RPC_COMMENT, DataType: rpc.DataString, Value: fmt.Sprintf("Sent you asset %s at height %d", scid, Wallet.Height)},
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
		log.Println("[SendAsset]", err)
		return
	}

	log.Println("[SendAsset] Send Asset TX:", txid)
	AddLog("Send Asset TX: " + txid.TXID)
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
					log.Printf("[%s] TX Confirmed\n", tag)
					return true
				} else if !tx.In_pool && tx.Block_Height == 0 && tx.ValidBlock == "" {
					log.Printf("[%s] TX Failed\n", tag)
					return false
				}
			}
		}
	}

	log.Printf("[%s] Could Not Confirm TX\n", tag)

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
				log.Printf("[%s] TX Confirmed\n", tag)
				return 100
			} else if !tx.In_pool && tx.Block_Height == 0 && tx.ValidBlock == "" {
				log.Printf("[%s] TX Failed, Retrying next block\n", tag)
				time.Sleep(3 * time.Second)
				for Wallet.Height <= next_block {
					time.Sleep(3 * time.Second)
				}
				return 1
			}
		}
	}

	log.Printf("[%s] Could Not Confirm TX\n", tag)

	return 2
}

// Send a message to destination address through Dero transaction, with ringsize selection
func SendMessage(dest, msg string, rings uint64) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	var dstport [8]byte
	rand.Read(dstport[:])

	response := rpc.Arguments{
		{Name: rpc.RPC_DESTINATION_PORT, DataType: rpc.DataUint64, Value: binary.BigEndian.Uint64(dstport[:])},
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
		log.Println("[SendMessage]", err)
		return
	}

	log.Println("[SendMessage] Send Message TX:", txid)
}
