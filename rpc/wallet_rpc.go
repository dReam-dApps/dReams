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

type wallet struct {
	UserPass   string
	idHash     string
	Rpc        string
	Address    string
	ClientKey  string
	Balance    uint64
	TokenBal   uint64
	TourneyBal string
	Height     int
	Connect    bool
	PokerOwner bool
	BetOwner   bool
	KeyLock    bool
	Service    bool
	LogEntry   *widget.Entry
}

var Wallet wallet

func StringToInt(s string) int {
	if s != "" {
		i, err := strconv.Atoi(s)
		if err != nil {
			log.Println("[StringToInt]", err)
			return 0
		}
		return i
	}

	return 0
}

// Add entry to gui log
func AddLog(t string) {
	Wallet.LogEntry.SetText(Wallet.LogEntry.Text + "\n\n" + t)
	Wallet.LogEntry.Refresh()
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

	max := container.NewMax(cont, vbox)

	return max
}

// Save string data to file
// func saveLog(data string) {
// 	file_name := fmt.Sprintf("Log-%s", time.Now().Format(time.UnixDate))
// 	if f, err := os.Create(file_name); err == nil {
// 		defer f.Close()
// 		if _, err = f.WriteString(data); err != nil {
// 			log.Println("[saveLog]", err)
// 			return
// 		}

// 		log.Println("[saveLog] Log File Saved", file_name)
// 	} else {
// 		log.Println("[saveLog]", err)
// 	}
// }

// Get Dero address from keys
func DeroAddress(v interface{}) (address string) {
	switch val := v.(type) {
	case string:
		decd, _ := hex.DecodeString(val)
		p := new(crypto.Point)
		if err := p.DecodeCompressed(decd); err == nil {
			addr := rpc.NewAddressFromKeys(p)
			address = addr.String()
		} else {
			address = string(decd)
		}
	}

	return address
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
	if Wallet.Connect {
		rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
		defer cancel()

		var result string
		params := []string{"Hello", "World", "!"}
		if err := rpcClientW.CallFor(ctx, &result, "Echo", params); err != nil {
			Wallet.Connect = false
			log.Printf("[%s] %s\n", tag, err)
			return
		}

		if result != "WALLET Hello World !" {
			Wallet.Connect = false
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
		Wallet.Connect = false
		log.Printf("[%s] %s\n", tag, err)
		return
	}

	if result.Address[0:4] == "dero" && len(result.Address) == 66 {
		Wallet.Connect = true
		log.Printf("[%s] Wallet Connected\n", tag)
		log.Printf("[%s] Dero Address: %s\n", tag, result.Address)
		Wallet.Address = result.Address
		id := []byte(result.Address)
		hash := sha256.Sum256(id)
		Wallet.idHash = hex.EncodeToString(hash[:])
	} else {
		Wallet.Connect = false
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

// Get wallet Dero balance
func GetBalance() {
	if Wallet.Connect {
		rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
		defer cancel()

		var result *rpc.GetBalance_Result
		if err := rpcClientW.CallFor(ctx, &result, "GetBalance"); err != nil {
			log.Println("[GetBalance]", err)
			return
		}

		Wallet.Balance = result.Unlocked_Balance
		Display.Dero_balance = fromAtomic(result.Unlocked_Balance)
	}
}

// Get wallet balance of token by SCID
func TokenBalance(scid string) uint64 {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	var result *rpc.GetBalance_Result
	sc := crypto.HashHexToHash(scid)
	params := &rpc.GetBalance_Params{
		SCID: sc,
	}

	if err := rpcClientW.CallFor(ctx, &result, "GetBalance", params); err != nil {
		log.Println("[TokenBalance]", err)
		return 0
	}

	return result.Unlocked_Balance
}

// Get wallet dReams token balance
func DreamsBalance() {
	if Wallet.Connect {
		bal := TokenBalance(dReamsSCID)
		Display.Token_balance = fromAtomic(bal)
		Wallet.TokenBal = bal
	}
}

// Get tournament token balance
func TourneyBalance() {
	if Wallet.Connect && Round.Tourney {
		bal := TokenBalance(TourneySCID)
		value := float64(bal)
		Wallet.TourneyBal = fmt.Sprintf("%.2f", value/100000)
	}
}

// Deposit tournament chips to leaderboard SC
func TourneyDeposit(bal uint64, name string) {
	if bal > 0 {
		rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
		defer cancel()

		arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Deposit"}
		arg2 := rpc.Argument{Name: "name", DataType: "S", Value: name}
		args := rpc.Arguments{arg1, arg2}
		txid := rpc.Transfer_Result{}

		scid := crypto.HashHexToHash(TourneySCID)
		t1 := rpc.Transfer{
			SCID:        scid,
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      0,
			Burn:        bal,
		}

		t := []rpc.Transfer{t1}
		fee := GasEstimate(TourneySCID, "[Holdero]", args, t)
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

// Get wallet height
func GetHeight() {
	if Wallet.Connect {
		rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
		defer cancel()

		var result *rpc.GetHeight_Result
		if err := rpcClientW.CallFor(ctx, &result, "GetHeight"); err != nil {
			log.Println("[dReams]", err)
			return
		}

		Wallet.Height = int(result.Height)
		Display.Wallet_height = fmt.Sprint(result.Height)
	}
}

// Submit playerId, name, avatar and sit at Holdero table
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
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t)
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
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t)
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
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t)
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
func DealHand() {
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
			t2 := rpc.Transfer{
				SCID:        crypto.HashHexToHash(dReamsSCID),
				Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
				Burn:        amount,
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

	fee := GasEstimate(Round.Contract, "[Holdero]", args, t)
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
}

// Make Holdero bet
func Bet(amt string) {
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
				Burn:        ToAtomicOne(amt),
			}
		} else {
			t1 = rpc.Transfer{
				SCID:        crypto.HashHexToHash(dReamsSCID),
				Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
				Burn:        ToAtomicOne(amt),
			}
		}
	} else {
		t1 = rpc.Transfer{
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      0,
			Burn:        ToAtomicOne(amt),
		}
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t)
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
}

// Holdero check and fold
func Check() {
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
			t1 = rpc.Transfer{
				SCID:        crypto.HashHexToHash(dReamsSCID),
				Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
				Burn:        0,
			}
		}
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t)
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
}

// Holdero single winner payout
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
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t)
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
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t)
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
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t)
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

// Shuffle deck for Holdero, clean above 0 can retrieve balance
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
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t)
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

// Timeout a player at Holdero table
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
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t)
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

// Force start a Holdero table with empty seats
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
	fee := GasEstimate(Round.Contract, "[Holdero]", args, t)
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
		fee := GasEstimate(Round.Contract, "[Holdero]", args, t)
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
	fee := GasEstimate(BaccSCID, "[dReams]", args, t)
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

	log.Println("[dReams] Get dReams", txid)
	AddLog("Get dReams " + txid.TXID)
}

// Swap dReams for Dero
func TradedReams(amt uint64) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "ConvertChips"}
	args := rpc.Arguments{arg1}
	txid := rpc.Transfer_Result{}

	scid := crypto.HashHexToHash(dReamsSCID)
	t1 := rpc.Transfer{
		SCID:        scid,
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        amt,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(BaccSCID, "[dReams]", args, t)
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

	log.Println("[dReams] Trade dReams TX:", txid)
	AddLog("Trade dReams TX: " + txid.TXID)
}

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
			Amount:      300000,
		}
	}

	return
}

// Install new Holdero SC
func UploadHolderoContract(pub int) {
	if Signal.Daemon && Wallet.Connect {
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
func BaccBet(amt, w string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "PlayBaccarat"}
	arg2 := rpc.Argument{Name: "betOn", DataType: "S", Value: w}
	args := rpc.Arguments{arg1, arg2}
	txid := rpc.Transfer_Result{}

	scid := crypto.HashHexToHash(dReamsSCID)
	t1 := rpc.Transfer{
		SCID:        scid,
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        ToAtomicOne(amt),
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(BaccSCID, "[Baccarat]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     BaccSCID,
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
}

// Place higher prediction to SC
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
	fee := GasEstimate(scid, "[Predictions]", args, t)
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
	fee := GasEstimate(scid, "[Predictions]", args, t)
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

// Rate a SC with dReam Tables rating system
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
	fee := GasEstimate(RatingSCID, "[RateSCID]", args, t)
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

// prediction leaderboard
// func NameChange(scid, name string) error { /// change leaderboard name
// 	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
// 	defer cancel()
//
// 	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "NameChange"}
// 	arg2 := rpc.Argument{Name: "name", DataType: "S", Value: name}
// 	args := rpc.Arguments{arg1, arg2}
// 	txid := rpc.Transfer_Result{}
//
// 	t1 := rpc.Transfer{
// 		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
// 		Amount:      0,
// 		Burn:        10000,
// 	}
//
// 	t := []rpc.Transfer{t1}
// 	fee := GasEstimate(scid, "[Predictions]", args, t)
// 	params := &rpc.Transfer_Params{
// 		Transfers: t,
// 		SC_ID:     scid,
// 		SC_RPC:    args,
// 		Ringsize:  2,
// 		Fees:      fee,
// 	}
//
// 	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
// 	if err != nil {
// 		log.Println("[NameChange]", err)
// 		return nil
// 	}
//
// 	log.Println("[Predictions] Name Change TX:", txid)
// 	AddLog("Name Change TX: " + txid.TXID)
//
// 	return err
// }
//
// func RemoveAddress(scid, name string) error {
// 	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
// 	defer cancel()
//
// 	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Remove"}
// 	arg2 := rpc.Argument{Name: "name", DataType: "S", Value: name}
// 	args := rpc.Arguments{arg1, arg2}
// 	txid := rpc.Transfer_Result{}
//
// 	t1 := rpc.Transfer{
// 		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
// 		Amount:      0,
// 		Burn:        10000,
// 	}
//
// 	t := []rpc.Transfer{t1}
// 	fee := GasEstimate(scid, "[Predictions]", args, t)
// 	params := &rpc.Transfer_Params{
// 		Transfers: t,
// 		SC_ID:     scid,
// 		SC_RPC:    args,
// 		Ringsize:  2,
// 		Fees:      fee,
// 	}
//
// 	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
// 	if err != nil {
// 		log.Println("[RemoveAddress]", err)
// 		return nil
// 	}
//
// 	log.Println("[Predictions] Remove TX:", txid)
// 	AddLog("Remove TX: " + txid.TXID)
//
// 	return err
// }

// Service prediction place by received tx
func AutoPredict(p int, amt, src uint64, scid, addr, tx string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	var hl string
	chopped_scid := scid[:6] + "..." + scid[58:]
	chopped_txid := tx[:6] + "..." + tx[58:]
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
	fee := GasEstimate(scid, "[AuotPredict]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[AuotPredict]", err)
		return
	}

	log.Println("[AuotPredict] Prediction TX:", txid)
	AddLog("AuotPredict TX: " + txid.TXID)
}

// Service refund if bet void
func ServiceRefund(amt, src uint64, scid, addr, msg, tx string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	chopped_scid := scid[:6] + "..." + scid[58:]
	chopped_txid := tx[:6] + "..." + tx[58:]
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
}

// Service sports book by received tx
func AutoBook(amt, pre, src uint64, n, abv, scid, addr, tx string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	chopped_scid := scid[:6] + "..." + scid[58:]
	chopped_txid := tx[:6] + "..." + tx[58:]
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
	fee := GasEstimate(scid, "[AuotBook]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[AuotBook]", err)
		return
	}

	log.Println("[AuotBook] Book TX:", txid)
	AddLog("AuotBook TX: " + txid.TXID)
}

// Update bet SC vars
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
	fee := GasEstimate(scid, "[VarUpdate]", args, t)
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

// Add new owner to bet SC
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
	fee := GasEstimate(scid, "[Predictions]", args, t)
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

// Remove owner from bet SC
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
	fee := GasEstimate(scid, "[Predictions]", args, t)
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

// User can refund a void dPrediction payout
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
	fee := GasEstimate(scid, "[Predictions]", args, t)
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

// Book sports team
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
	fee := GasEstimate(scid, "[Sports]", args, t)
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

// User can refund a void dSports payout
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
	fee := GasEstimate(scid, "[Sports]", args, t)
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

// Owner sets a game
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
	fee := GasEstimate(scid, "[Sports]", args, t)
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

// Owner sets a prediction
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
	fee := GasEstimate(scid, "[Predictions]", args, t)
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

// Owner cancel for intiated bet
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
	fee := GasEstimate(scid, tag, args, t)
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
func PostPrediction(scid string, price int) {
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
	fee := GasEstimate(scid, "[Predictions]", args, t)
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
}

// dSports payout
func EndSports(scid, num, team string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "S_end"}
	arg2 := rpc.Argument{Name: "n", DataType: "S", Value: num}
	arg3 := rpc.Argument{Name: "team", DataType: "S", Value: team}
	args := rpc.Arguments{arg1, arg2, arg3}
	txid := rpc.Transfer_Result{}

	t := []rpc.Transfer{}
	fee := GasEstimate(scid, "[Sports]", args, t)
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
}

// dPrediction payout
func EndPrediction(scid string, price int) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "P_end"}
	arg2 := rpc.Argument{Name: "price", DataType: "U", Value: price}
	args := rpc.Arguments{arg1, arg2, arg2}
	txid := rpc.Transfer_Result{}

	t := []rpc.Transfer{}
	fee := GasEstimate(scid, "[Predictions]", args, t)
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
}

// Install new bet contract
func UploadBetContract(c bool, pub int) {
	if Signal.Daemon && Wallet.Connect {
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
	fee := GasEstimate(GnomonSCID, "[dReams]", args, t)
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

	log.Println("[dReams] Set Headers TX:", txid)
	AddLog("Set Headers TX: " + txid.TXID)
}

// Claim transfered NFA token
func ClaimNfa(scid string) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "ClaimOwnership"}
	args := rpc.Arguments{arg1, arg1}
	txid := rpc.Transfer_Result{}

	nfa_sc := crypto.HashHexToHash(scid)
	t1 := rpc.Transfer{
		SCID:        nfa_sc,
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        1,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(scid, "[dReams]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[ClaimNfa]", err)
		return
	}

	log.Println("[dReams] Claim TX:", txid)
	AddLog("Claim TX: " + txid.TXID)
}

// Send bid or buy to NFA SC
func NfaBidBuy(scid, bidor string, amt uint64) {
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
	fee := GasEstimate(scid, "[dReams]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[NfaBidBuy]", err)
		return
	}

	if bidor == "Bid" {
		log.Println("[dReams] NFA Bid TX:", txid)
		AddLog("Bid TX: " + txid.TXID)
	} else {
		log.Println("[dReams] NFA Buy TX:", txid)
		AddLog("Buy TX: " + txid.TXID)
	}
}

// List NFA for auction or sale
func NfaSetListing(scid, list, char string, dur, amt, perc uint64) {
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

	asset_scid := crypto.HashHexToHash(scid)
	t1 := rpc.Transfer{
		SCID:        asset_scid,
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        1,
	}

	/// dReams
	t2 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      5000,
		Burn:        0,
	}

	/// artificer
	t3 := rpc.Transfer{
		Destination: "dero1qy0khp9s9yw2h0eu20xmy9lth3zp5cacmx3rwt6k45l568d2mmcf6qgcsevzx",
		Amount:      5000,
		Burn:        0,
	}

	t := []rpc.Transfer{t1, t2, t3}
	fee := GasEstimate(scid, "[dReams]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[NfaSetListing]", err)
		return
	}

	log.Println("[dReams] NFA List TX:", txid)
	AddLog("NFA List TX: " + txid.TXID)
}

// Cancel listed NFA
func NfaCancelClose(scid, c string) {
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
	fee := GasEstimate(scid, "[dReams]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[NfaCancelClose]", err)
		return
	}

	if c == "CloseListing" {
		log.Println("[dReams] Close NFA Listing TX:", txid)
		AddLog("Close Listing TX: " + txid.TXID)
	} else {
		log.Println("[dReams] Cancel NFA Listing TX:", txid)
		AddLog("Cancel Listing TX: " + txid.TXID)
	}
}

// Get Iluma Tarot reading from SC
func TarotReading(num int) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Draw"}
	arg2 := rpc.Argument{Name: "num", DataType: "U", Value: num}
	args := rpc.Arguments{arg1, arg2}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        10000,
	}

	t := []rpc.Transfer{t1}
	fee := GasEstimate(TarotSCID, "[Tarot]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     TarotSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[TarotReading]", err)
		return
	}

	Tarot.Num = num
	Tarot.Last = txid.TXID
	Tarot.Notified = false

	log.Println("[Tarot] Reading TX:", txid)
	AddLog("Reading TX: " + txid.TXID)

	Tarot.CHeight = Wallet.Height
}

// Send asset to a destination wallet
func SendAsset(scid, dest string, payload bool) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	asset_scid := crypto.HashHexToHash(scid)
	t1 := rpc.Transfer{
		SCID:        asset_scid,
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

// Watch a sent tx and retry 3 times if failed
//   - tag for log print
func ConfirmTx(txid string, tag string, tries int) (retry int) {
	count := 0
	for (tries < 3) && Wallet.Connect && Signal.Daemon {
		count++
		time.Sleep(2 * time.Second)
		if tx := GetDaemonTx(txid); tx != nil {
			if count > 36 {
				break
			}

			if tx.In_pool {
				continue
			} else if !tx.In_pool && tx.Block_Height > 1 && tx.ValidBlock != "" {
				log.Printf("[%s] TX Confirmed\n", tag)
				return 100
			} else if !tx.In_pool && tx.Block_Height == 0 && tx.ValidBlock == "" {
				log.Printf("[%s] TX Failed, Retrying\n", tag)
				time.Sleep(6 * time.Second)
				return 1
			}
		}
	}

	log.Printf("[%s] Could Not Confirm TX\n", tag)

	return 100
}

// Send a message through Dero transaction
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
