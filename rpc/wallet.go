package rpc

import (
	"log"
	"sync"

	"fyne.io/fyne/v2/widget"
	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/walletapi"
)

type wallet struct {
	UserPass   string
	idHash     string
	Rpc        string
	Address    string
	ClientKey  string
	Balance    uint64
	TokenBal   map[string]uint64
	Height     int
	Connect    bool
	PokerOwner bool
	BetOwner   bool
	KeyLock    bool
	Service    bool
	MuC        sync.RWMutex
	MuB        sync.RWMutex
	File       *walletapi.Wallet_Disk
	LogEntry   *widget.Entry
}

var Wallet wallet

func (w *wallet) IsConnected() bool {
	w.MuC.RLock()
	defer w.MuC.RUnlock()

	return w.Connect
}

func (w *wallet) Connected(b bool) {
	w.MuC.Lock()
	w.Connect = b
	w.MuC.Unlock()
}

// Get Wallet.Balance, 0 if not connected
func (w *wallet) GetBalance() {
	if w.Connect {
		rpcClientW, ctx, cancel := SetWalletClient(w.Rpc, w.UserPass)
		defer cancel()

		var result *rpc.GetBalance_Result
		if err := rpcClientW.CallFor(ctx, &result, "GetBalance"); err != nil {
			log.Println("[GetBalance]", err)
			w.Balance = 0
			return
		}

		w.Balance = result.Unlocked_Balance

		return
	}

	w.Balance = 0
}

// Get single balance of Wallet.TokenBal[name], 0 if not connected
func (w *wallet) GetTokenBalance(name, scid string) {
	w.MuB.Lock()
	defer w.MuB.Unlock()

	if w.Connect {
		rpcClientW, ctx, cancel := SetWalletClient(w.Rpc, w.UserPass)
		defer cancel()

		params := &rpc.GetBalance_Params{
			SCID: crypto.HashHexToHash(scid),
		}

		var result *rpc.GetBalance_Result
		if err := rpcClientW.CallFor(ctx, &result, "GetBalance", params); err != nil {
			log.Println("[GetTokenBalance]", err)
			w.TokenBal[name] = 0
			return
		}

		w.TokenBal[name] = result.Unlocked_Balance

		return

	}

	w.TokenBal[name] = 0
}

// Read Wallet.TokenBal[name]
func (w *wallet) ReadTokenBalance(name string) uint64 {
	w.MuB.RLock()
	defer w.MuB.RUnlock()

	return w.TokenBal[name]
}
