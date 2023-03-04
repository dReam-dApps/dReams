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
	TokenBal   string
	TourneyBal string
	Height     string
	Connect    bool
	PokerOwner bool
	BetOwner   bool
	KeyLock    bool
	Service    bool
}

var Wallet wallet
var logEntry = widget.NewMultiLineEntry()

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

func addLog(t string) {
	logEntry.SetText(logEntry.Text + "\n\n" + t)
	logEntry.Refresh()
}

func SessionLog() *fyne.Container {
	logEntry.Disable()
	button := widget.NewButton("Save", func() {
		saveLog(logEntry.Text)
	})

	cont := container.NewMax(logEntry)

	vbox := container.NewVBox(
		layout.NewSpacer(),
		container.NewAdaptiveGrid(2,
			layout.NewSpacer(),
			button))

	max := container.NewMax(cont, vbox)

	return max
}

func saveLog(data string) {
	f, err := os.Create("Log " + time.Now().Format(time.UnixDate))

	if err != nil {
		log.Println("[saveLog]", err)
		return
	}

	defer f.Close()

	_, err = f.WriteString(data)

	if err != nil {
		log.Println("[saveLog]", err)
		return
	}

	log.Println("[dReams] Log File Saved")
}

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

func SetWalletClient(addr, pass string) (jsonrpc.RPCClient, context.Context, context.CancelFunc) { /// user:pass auth
	client := jsonrpc.NewClientWithOpts("http://"+addr+"/json_rpc", &jsonrpc.RPCClientOpts{
		CustomHeaders: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(pass)),
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	return client, ctx, cancel
}

func EchoWallet(wc bool) error { /// echo wallet for connection
	if wc {
		rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
		defer cancel()

		var result string
		params := []string{"Hello", "World", "!"}
		err := rpcClientW.CallFor(ctx, &result, "Echo", params)
		if err != nil {
			Wallet.Connect = false
			log.Println("[dReams]", err)
			return nil
		}

		if result != "WALLET Hello World !" {
			Wallet.Connect = false
		}

		return err
	}
	return nil
}

func GetAddress() error {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	var result *rpc.GetAddress_Result
	err := rpcClientW.CallFor(ctx, &result, "GetAddress")

	if err != nil {
		Wallet.Connect = false
		log.Println("[dReams]", err)
		return nil
	}

	address := len(result.Address)
	if address == 66 {
		Wallet.Connect = true
		log.Println("[dReams] Wallet Connected")
		log.Println("[dReams] Dero Address: " + result.Address)
		Wallet.Address = result.Address
		id := []byte(result.Address)
		hash := sha256.Sum256(id)
		Wallet.idHash = hex.EncodeToString(hash[:])
	} else {
		Wallet.Connect = false
	}

	return err
}

func GetTransaction(txid string) (*rpc.Entry, error) {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	var result *rpc.Get_Transfer_By_TXID_Result
	params := rpc.Get_Transfer_By_TXID_Params{
		TXID: txid,
	}
	err := rpcClientW.CallFor(ctx, &result, "GetTransferbyTXID", params)

	if err != nil {
		log.Println("[dReams]", err)
		return nil, nil
	}

	return &result.Entry, err
}

func GetBalance(wc bool) error { /// get wallet dero balance
	if wc {
		rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
		defer cancel()

		var result *rpc.GetBalance_Result
		err := rpcClientW.CallFor(ctx, &result, "GetBalance")
		if err != nil {
			log.Println("[GetBalance]", err)
			return nil
		}

		Wallet.Balance = result.Unlocked_Balance
		Display.Dero_balance = fromAtomic(result.Unlocked_Balance)

		return err
	}
	return nil
}

func TokenBalance(scid string) (uint64, error) { /// get wallet token balance
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	var result *rpc.GetBalance_Result
	sc := crypto.HashHexToHash(scid)
	params := &rpc.GetBalance_Params{
		SCID: sc,
	}

	err := rpcClientW.CallFor(ctx, &result, "GetBalance", params)
	if err != nil {
		log.Println("[TokenBalance]", err)
		return 0, nil
	}

	return result.Unlocked_Balance, err
}

func DreamsBalance(wc bool) { /// get wallet dReam balance
	if wc {
		bal, _ := TokenBalance(dReamsSCID)
		Wallet.TokenBal = fromAtomic(bal)
	}
}

func TourneyBalance(wc, t bool, scid string) { /// get tournament balance
	if wc && t {
		bal, _ := TokenBalance(scid)
		value := float64(bal)
		Wallet.TourneyBal = fmt.Sprintf("%.2f", value/100000)
	}
}

func TourneyDeposit(bal uint64, name string) error {
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
		fee, _ := GasEstimate(TourneySCID, "[Holdero]", args, t)
		params := &rpc.Transfer_Params{
			Transfers: t,
			SC_ID:     TourneySCID,
			SC_RPC:    args,
			Ringsize:  2,
			Fees:      fee,
		}

		err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
		if err != nil {
			log.Println("[TourneyDeposit]", err)
			return nil
		}

		log.Println("[Holdero] Tournament Deposit TX:", txid)
		addLog("Tournament Deposit TX: " + txid.TXID)

		return err
	}
	log.Println("[Holdero] No Tournament Chips")
	return nil
}

func GetHeight(wc bool) error { /// get wallet height
	if wc {
		rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
		defer cancel()

		var result *rpc.GetHeight_Result
		err := rpcClientW.CallFor(ctx, &result, "GetHeight")
		if err != nil {
			log.Println("[dReams]", err)
			return nil
		}

		Wallet.Height = fmt.Sprint(result.Height)

		return err
	}
	return nil
}

func SitDown(name, av string) error { /// sit at holdero table
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
	fee, _ := GasEstimate(Round.Contract, "[Holdero]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[SitDown]", err)
		return nil
	}

	log.Println("[Holdero] Sit Down TX:", txid)
	addLog("Sit Down TX: " + txid.TXID)

	return err
}

func Leave() error { /// leave holdero table
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
	fee, _ := GasEstimate(Round.Contract, "[Holdero]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[Leave]", err)
		return nil
	}

	log.Println("[Holdero] Leave TX:", txid)
	addLog("Leave Down TX: " + txid.TXID)

	return err
}

func SetTable(seats int, bb, sb, ante uint64, chips, name, av string) error { /// set holdero
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

	if Round.Version < 110 {
		args = rpc.Arguments{arg1, arg2, arg3, arg4, arg5, arg6}
	} else if Round.Version == 110 {
		arg7 := rpc.Argument{Name: "chips", DataType: "S", Value: chips}
		args = rpc.Arguments{arg1, arg2, arg3, arg4, arg5, arg6, arg7}
	}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []rpc.Transfer{t1}
	fee, _ := GasEstimate(Round.Contract, "[Holdero]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[SetTable]", err)
		return nil
	}

	log.Println("[Holdero] Set Table TX:", txid)
	addLog("Set Table TX: " + txid.TXID)

	return err
}

func DealHand() error { /// holdero hand
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

	fee, _ := GasEstimate(Round.Contract, "[Holdero]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[DealHand]", err)
		return nil
	}

	Display.Res = ""
	log.Println("[Holdero] Deal TX:", txid)
	updateStatsWager(float64(amount) / 100000)
	addLog("Deal TX: " + txid.TXID)

	return err
}

func Bet(amt string) error { /// holdero bet
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
	fee, _ := GasEstimate(Round.Contract, "[Holdero]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[Bet]", err)
		return nil
	}

	if f, err := strconv.ParseFloat(amt, 32); err == nil {
		updateStatsWager(float64(f))
	}

	Display.Res = ""
	Signal.PlacedBet = true
	log.Println("[Holdero] Bet TX:", txid)
	addLog("Bet TX: " + txid.TXID)

	return err
}

func Check() error { /// holdero check and fold
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
	fee, _ := GasEstimate(Round.Contract, "[Holdero]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[Check]", err)
		return nil
	}

	Display.Res = ""
	log.Println("[Holdero] Check/Fold TX:", txid)
	addLog("Check/Fold TX: " + txid.TXID)

	return err
}

func PayOut(w string) error { /// holdero single winner
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
	fee, _ := GasEstimate(Round.Contract, "[Holdero]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[PayOut]", err)
		return nil
	}

	log.Println("[Holdero] Payout TX:", txid)
	addLog("Holdero Payout TX: " + txid.TXID)

	return err
}

func PayoutSplit(r ranker, f1, f2, f3, f4, f5, f6 bool) error { /// holdero split winners
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
	fee, _ := GasEstimate(Round.Contract, "[Holdero]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[PayoutSplit]", err)
		return nil
	}

	log.Println("[Holdero] Split Winner TX:", txid)
	addLog("Split Winner TX: " + txid.TXID)

	return err
}

func RevealKey(key string) error { /// holdero reveal
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
	fee, _ := GasEstimate(Round.Contract, "[Holdero]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[RevealKey]", err)
		return nil
	}

	Display.Res = ""
	log.Println("[Holdero] Reveal TX:", txid)
	addLog("Reveal TX: " + txid.TXID)

	return err
}

func CleanTable(amt uint64) error { /// shuffle and clean holdero
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
	fee, _ := GasEstimate(Round.Contract, "[Holdero]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[CleanTable]", err)
		return nil
	}

	log.Println("[Holdero] Clean Table TX:", txid)
	addLog("Clean Table TX: " + txid.TXID)

	return err
}

func TimeOut() error {
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
	fee, _ := GasEstimate(Round.Contract, "[Holdero]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[TimeOut]", err)
		return nil
	}

	log.Println("[Holdero] Timeout TX:", txid)
	addLog("Timeout TX: " + txid.TXID)

	return err
}

func ForceStat() error {
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
	fee, _ := GasEstimate(Round.Contract, "[Holdero]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[ForceStart]", err)
		return nil
	}

	log.Println("[Holdero] Force Start TX:", txid)
	addLog("Force Start TX: " + txid.TXID)

	return err
}

type CardSpecs struct {
	Faces struct {
		Name string `json:"Name"`
		Url  string `json:"Url"`
	} `json:"Faces"`
	Backs struct {
		Name string `json:"Name"`
		Url  string `json:"Url"`
	} `json:"Backs"`
}

type TableSpecs struct {
	MaxBet float64 `json:"Maxbet"`
	MinBuy float64 `json:"Minbuy"`
	MaxBuy float64 `json:"Maxbuy"`
	Time   int     `json:"Time"`
}

func SharedDeckUrl(face, faceUrl, back, backUrl string) error {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	var cards string
	if face == "" || back == "" {
		cards = "nil"
	} else {
		cards = `{"Faces":{"Name":"` + face + `", "Url":"` + faceUrl + `"},"Backs":{"Name":"` + back + `", "Url":"` + backUrl + `"}}`
	}

	specs := "nil"
	// specs := `{"MaxBet":10,"MinBuy":10,"MaxBuy":20, "Time":120}`

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Deck"}
	arg2 := rpc.Argument{Name: "face", DataType: "S", Value: cards}
	arg3 := rpc.Argument{Name: "back", DataType: "S", Value: specs}
	args := rpc.Arguments{arg1, arg2, arg3}
	txid := rpc.Transfer_Result{}

	t1 := rpc.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []rpc.Transfer{t1}
	fee, _ := GasEstimate(Round.Contract, "[Holdero]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     Round.Contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[SharedDeckUrl]", err)
		return nil
	}

	log.Println("[Holdero] Shared TX:", txid)
	addLog("Shared TX: " + txid.TXID)

	return err
}

func GetdReams(amt uint64) error {
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
	fee, _ := GasEstimate(BaccSCID, "[dReams]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     BaccSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[GetdReams]", err)
		return nil
	}

	log.Println("[dReams] Get dReams", txid)
	addLog("Get dReams " + txid.TXID)

	return err
}

func TradedReams(amt uint64) error {
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
	fee, _ := GasEstimate(BaccSCID, "[dReams]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     BaccSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[TradedReams]", err)
		return nil
	}

	log.Println("[dReams] Trade dReams TX:", txid)
	addLog("Trade dReams TX: " + txid.TXID)

	return err
}

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

func UploadHolderoContract(d, w bool, pub int) error {
	if d && w {
		rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
		defer cancel()

		code, code_err := GetHoldero110Code(d, pub)
		if code_err != nil {
			log.Println("[UploadHolderoContract]", code_err)
			return nil
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

		err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
		if err != nil {
			log.Println("[UploadHolderoContract]", err)
			return nil
		}

		log.Println("[Holdero] Upload TX:", txid)
		addLog("Holdero Upload TX:" + txid.TXID)

		return err
	}

	return nil
}

func BaccBet(amt, w string) error {
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
	fee, _ := GasEstimate(BaccSCID, "[Baccarat]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     BaccSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[BaccBet]", err)
		return nil
	}

	Bacc.Last = txid.TXID
	Bacc.Notified = false
	if w == "player" {
		log.Println("[Baccarat] Player TX:", txid)
		addLog("Baccarat Player TX: " + txid.TXID)
	} else if w == "banker" {
		log.Println("[Baccarat] Banker TX:", txid)
		addLog("Baccarat Banker TX: " + txid.TXID)
	} else {
		log.Println("[Baccarat] Tie TX:", txid)
		addLog("Baccarat Tie TX: " + txid.TXID)
	}

	Bacc.CHeight = StringToInt(Wallet.Height)

	return err
}

func PredictHigher(scid, addr string) error {
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
	fee, _ := GasEstimate(scid, "[Predictions]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[PredictHigher]", err)
		return nil
	}

	log.Println("[Predictions] Prediction TX:", txid)
	addLog("Prediction TX: " + txid.TXID)

	return err
}

func PredictLower(scid, addr string) error {
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
	fee, _ := GasEstimate(scid, "[Predictions]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[PredictLower]", err)
		return nil
	}

	log.Println("[Predictions] Prediction TX:", txid)
	addLog("Prediction TX: " + txid.TXID)

	return err
}

func RateSCID(scid string, amt, pos uint64) error {
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
	fee, _ := GasEstimate(RatingSCID, "[RateSCID]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     RatingSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[RateSCID]", err)
		return nil
	}

	log.Println("[RateSCID] Rate TX:", txid)
	addLog("Rate TX: " + txid.TXID)

	return err
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
// 	fee, _ := GasEstimate(scid, "[Predictions]", args, t)
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
// 	addLog("Name Change TX: " + txid.TXID)
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
// 	fee, _ := GasEstimate(scid, "[Predictions]", args, t)
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
// 	addLog("Remove TX: " + txid.TXID)
//
// 	return err
// }

func AuotPredict(p int, amt, src uint64, scid, addr string) error {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	var hl string
	chopped_scid := scid[:6] + "..." + scid[58:]
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
		{Name: rpc.RPC_COMMENT, DataType: rpc.DataString, Value: fmt.Sprintf("Placed a %s %s bet on %s at height %s", walletapi.FormatMoney(amt), hl, chopped_scid, Wallet.Height)},
	}

	t1 := rpc.Transfer{
		Destination: addr,
		Amount:      1,
		Burn:        amt,
		Payload_RPC: response,
	}

	t := []rpc.Transfer{t1}
	fee, _ := GasEstimate(scid, "[AuotPredict]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[AuotPredict]", err)
		return nil
	}

	log.Println("[AuotPredict] Prediction TX:", txid)
	addLog("AuotPredict TX: " + txid.TXID)

	return err
}

func ServiceRefund(amt, src uint64, scid, addr, msg string) error {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	chopped_scid := scid[:6] + "..." + scid[58:]
	response := rpc.Arguments{
		{Name: rpc.RPC_DESTINATION_PORT, DataType: rpc.DataUint64, Value: uint64(0)},
		{Name: rpc.RPC_SOURCE_PORT, DataType: rpc.DataUint64, Value: src},
		{Name: rpc.RPC_COMMENT, DataType: rpc.DataString, Value: msg + fmt.Sprintf(", refunded %s bet on %s at height %s", walletapi.FormatMoney(amt), chopped_scid, Wallet.Height)},
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

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[ServiceRefund]", err)
		return nil
	}

	log.Println("[ServiceRefund] Refund TX:", txid)
	addLog("Refund TX: " + txid.TXID)

	return err
}

func AuotBook(amt, pre, src uint64, n, abv, scid, addr string) error {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	chopped_scid := scid[:6] + "..." + scid[58:]
	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Book"}
	arg2 := rpc.Argument{Name: "pre", DataType: "U", Value: pre}
	arg3 := rpc.Argument{Name: "n", DataType: "S", Value: n}
	arg4 := rpc.Argument{Name: "addr", DataType: "S", Value: addr}
	args := rpc.Arguments{arg1, arg2, arg3, arg4}
	txid := rpc.Transfer_Result{}

	response := rpc.Arguments{
		{Name: rpc.RPC_DESTINATION_PORT, DataType: rpc.DataUint64, Value: uint64(0)},
		{Name: rpc.RPC_SOURCE_PORT, DataType: rpc.DataUint64, Value: src},
		{Name: rpc.RPC_COMMENT, DataType: rpc.DataString, Value: fmt.Sprintf("Placed a %s %s bet on %s at height %s", walletapi.FormatMoney(amt), abv, chopped_scid, Wallet.Height)},
	}

	t1 := rpc.Transfer{
		Destination: addr,
		Amount:      1,
		Burn:        amt,
		Payload_RPC: response,
	}

	t := []rpc.Transfer{t1}
	fee, _ := GasEstimate(scid, "[AuotBook]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[AuotBook]", err)
		return nil
	}

	log.Println("[AuotBook] Book TX:", txid)
	addLog("AuotBook TX: " + txid.TXID)

	return err
}

func VarUpdate(scid string, ta, tb, tc, l, hl int) error { /// change leaderboard name
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
	fee, _ := GasEstimate(scid, "[VarUpdate]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[VarUpdate]", err)
		return nil
	}

	log.Println("[VarUpdate] VarUpdate TX:", txid)
	addLog("VarUpdate TX: " + txid.TXID)

	return err
}

func AddOwner(scid, addr string) error { /// change leaderboard name
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
	fee, _ := GasEstimate(scid, "[Predictions]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[AddSigner]", err)
		return nil
	}

	log.Println("[Predictions] Add Signer TX:", txid)
	addLog("Add Signer TX: " + txid.TXID)

	return err
}

func RemoveOwner(scid string, num int) error {
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
	fee, _ := GasEstimate(scid, "[Predictions]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[RemoveSigner]", err)
		return nil
	}

	log.Println("[Predictions] Remove Signer TX:", txid)
	addLog("Remove Signer: " + txid.TXID)

	return err
}

func PredictionRefund(scid, tic string) error { /// change leaderboard name
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
	fee, _ := GasEstimate(scid, "[Predictions]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[PredictionRefund]", err)
		return nil
	}

	log.Println("[Predictions] Refund TX:", txid)
	addLog("Refund TX: " + txid.TXID)

	return err
}

func PickTeam(scid, multi, n string, a uint64, pick int) error { /// pick sports team
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
	fee, _ := GasEstimate(scid, "[Sports]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[PickTeam]", err)
		return nil
	}

	log.Println("[Sports] Pick TX:", txid)
	addLog("Pick TX: " + txid.TXID)

	return err
}

func SportsRefund(scid, tic, n string) error { /// change leaderboard name
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
	fee, _ := GasEstimate(scid, "[Sports]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[SportsRefund]", err)
		return nil
	}

	log.Println("[Sports] Refund TX:", txid)
	addLog("Refund TX: " + txid.TXID)

	return err
}

func SetSports(end int, amt, dep uint64, scid, league, game, feed string) error {
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
	fee, _ := GasEstimate(scid, "[Sports]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[SetSports]", err)
		return nil
	}

	log.Println("[Sports] Set TX:", txid)
	addLog("Set Sports TX: " + txid.TXID)

	return err
}

func SetPrediction(end, mark int, amt, dep uint64, scid, predict, feed string) error {
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
	fee, _ := GasEstimate(scid, "[Predictions]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[SetPrediction]", err)
		return nil
	}

	log.Println("[Predictions] Set TX:", txid)
	addLog("Set Prediction TX: " + txid.TXID)

	return err
}

func CancelInitiatedBet(scid string, b int) error {
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
	fee, _ := GasEstimate(scid, tag, args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[CancelInitiatedBet]", err)
		return nil
	}

	if b == 0 {
		log.Println("[Predictions] Cancel TX:", txid)
		addLog("Cancel Prediction TX: " + txid.TXID)
	} else {
		log.Println("[Sports] Cancel TX:", txid)
		addLog("Cancel Sports TX: " + txid.TXID)
	}

	return err
}

func PostPrediction(scid string, price int) error {
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
	fee, _ := GasEstimate(scid, "[Predictions]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[PostPrediction]", err)
		return nil
	}

	log.Println("[Predictions] Post TX:", txid)
	addLog("Post TX: " + txid.TXID)

	return err
}

func EndSports(scid, num, team string) error {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "S_end"}
	arg2 := rpc.Argument{Name: "n", DataType: "S", Value: num}
	arg3 := rpc.Argument{Name: "team", DataType: "S", Value: team}
	args := rpc.Arguments{arg1, arg2, arg3}
	txid := rpc.Transfer_Result{}

	t := []rpc.Transfer{}
	fee, _ := GasEstimate(scid, "[Sports]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[EndSports]", err)
		return nil
	}

	log.Println("[Sports] Payout TX:", txid)
	addLog("Sports Payout TX: " + txid.TXID)

	return err
}

func EndPrediction(scid string, price int) error {
	rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
	defer cancel()

	arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "P_end"}
	arg2 := rpc.Argument{Name: "price", DataType: "U", Value: price}
	args := rpc.Arguments{arg1, arg2, arg2}
	txid := rpc.Transfer_Result{}

	t := []rpc.Transfer{}
	fee, _ := GasEstimate(scid, "[Predictions]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_Value:  0,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[EndPrediction]", err)
		return nil
	}

	log.Println("[Predictions] Payout TX:", txid)
	addLog("Prediction Payout TX: " + txid.TXID)

	return err
}

func UploadBetContract(d, w, c bool, pub int) error {
	if d && w {
		rpcClientW, ctx, cancel := SetWalletClient(Wallet.Rpc, Wallet.UserPass)
		defer cancel()

		var code string
		var code_err error

		if c {
			code, code_err = GetPredictCode(d, pub)
			if code_err != nil {
				log.Println("[UploadBetContract]", code_err)
				return nil
			}
		} else {
			code, code_err = GetSportsCode(d, pub)
			if code_err != nil {
				log.Println("[UploadBetContract]", code_err)
				return nil
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
			Fees:      11000,
		}

		err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
		if err != nil {
			log.Println("[UploadBetContract]", err)
			return nil
		}

		if c {
			log.Println("[Predictions] Upload TX:", txid)
			addLog("Prediction Upload TX:" + txid.TXID)
		} else {
			log.Println("[Sports] Upload TX:", txid)
			addLog("Sports Upload TX:" + txid.TXID)
		}

		return err
	}

	return nil
}

func SetHeaders(name, desc, icon, scid string) error {
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
	fee, _ := GasEstimate(GnomonSCID, "[dReams]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_Value:  0,
		SC_ID:     GnomonSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}
	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[SetHeaders]", err)
		return nil
	}
	log.Println("[dReams] Set Headers TX:", txid)
	addLog("Set Headers TX: " + txid.TXID)

	return err
}

func ClaimNfa(scid string) error {
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
	fee, _ := GasEstimate(scid, "[dReams]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[ClaimNfa]", err)
		return nil
	}

	log.Println("[dReams] Claim TX:", txid)
	addLog("Claim TX: " + txid.TXID)

	return err
}

func NfaBidBuy(scid, bidor string, amt uint64) error {
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
	fee, _ := GasEstimate(scid, "[dReams]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[NfaBidBuy]", err)
		return nil
	}

	if bidor == "Bid" {
		log.Println("[dReams] NFA Bid TX:", txid)
		addLog("Bid TX: " + txid.TXID)
	} else {
		log.Println("[dReams] NFA Buy TX:", txid)
		addLog("Buy TX: " + txid.TXID)
	}

	return err
}

func NfaSetListing(scid, list, char string, dur, amt, perc uint64) error {
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
	fee, _ := GasEstimate(scid, "[dReams]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}
	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[NfaSetListing]", err)
		return nil
	}
	log.Println("[dReams] NFA List TX:", txid)
	addLog("NFA List TX: " + txid.TXID)

	return err
}

func NfaCancelClose(scid, c string) error {
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
	fee, _ := GasEstimate(scid, "[dReams]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[NfaCancelClose]", err)
		return nil
	}

	if c == "CloseListing" {
		log.Println("[dReams] Close NFA Listing TX:", txid)
		addLog("Close Listing TX: " + txid.TXID)
	} else {
		log.Println("[dReams] Cancel NFA Listing TX:", txid)
		addLog("Cancel Listing TX: " + txid.TXID)
	}

	return err
}

func TarotReading(num int) error {
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
	fee, _ := GasEstimate(TarotSCID, "[Tarot]", args, t)
	params := &rpc.Transfer_Params{
		Transfers: t,
		SC_ID:     TarotSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[TarotReading]", err)
		return nil
	}

	Tarot.Num = num
	Tarot.Last = txid.TXID
	Tarot.Notified = false

	log.Println("[Tarot] Reading TX:", txid)
	addLog("Reading TX: " + txid.TXID)

	Tarot.CHeight = StringToInt(Wallet.Height)

	return err
}

func SendAsset(scid, dest string, payload bool) error {
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
			{Name: rpc.RPC_COMMENT, DataType: rpc.DataString, Value: fmt.Sprintf("Sent you asset %s at height %s", scid, Wallet.Height)},
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

	err := rpcClientW.CallFor(ctx, &txid, "transfer", params)
	if err != nil {
		log.Println("[SendAsset]", err)
		return nil
	}

	log.Println("[SendAsset] Send Asset TX:", txid)
	addLog("Send Asset TX: " + txid.TXID)

	return err
}
