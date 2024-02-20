package rpc

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2/widget"
	"github.com/civilware/Gnomon/structures"
	"github.com/deroproject/derohe/walletapi"
	"github.com/sirupsen/logrus"
	"github.com/ybbus/jsonrpc/v3"
)

type wallet struct {
	IdHash    string
	Address   string
	ClientKey string
	balances  map[string]*balance
	height    uint64
	Connect   bool
	KeyLock   bool
	muC       sync.RWMutex
	sync.RWMutex
	File     *walletapi.Wallet_Disk
	LogEntry *widget.Entry
	RPC      RPCserver
	WS       XSWDserver
}

type balance struct {
	decimal int
	atomic  uint64
	format  string
	scid    string
}

var Wallet wallet
var logger = structures.Logger.WithFields(logrus.Fields{})

// Close all connections to wallet
func (w *wallet) CloseConnections(tag string) {
	if w.RPC.client != nil {
		logger.Infof("[%s] RPC Closed\n", tag)
		w.RPC.client = nil
		w.RPC.cancel = nil
	}

	if w.WS.conn != nil {
		logger.Infof("[%s] XSWD Closed\n", tag)
		w.WS.conn.Close()
		w.WS.conn = nil
	}
}

// Check if wallet is connected
func (w *wallet) IsConnected() bool {
	w.muC.RLock()
	defer w.muC.RUnlock()

	return w.Connect
}

// Set wallet connection
func (w *wallet) Connected(b bool) {
	w.muC.Lock()
	w.Connect = b
	if !b {
		w.height = 0
		w.Address = ""
	}
	w.muC.Unlock()
}

// Wallet call switch for RPC or XSWD connections
func (w *wallet) CallFor(out interface{}, method string, params ...interface{}) (err error) {
	if w.RPC.client != nil {
		if err = w.RPC.CallFor(&out, method, params...); err != nil {
			return
		}
	} else if w.WS.conn != nil {
		for w.WS.IsRequesting() {
			time.Sleep(500 * time.Millisecond)
			logger.Warnln("[XSWD] Request sleep...")
		}

		if err = w.WS.CallFor(&out, method, jsonrpc.Params(params...)); err != nil {
			return
		}
	} else {
		err = fmt.Errorf("wallet not connected for %s", method)
	}

	return
}

// Call EchoWallet if wallet is connected and set Connected
func (w *wallet) Echo() {
	if w.IsConnected() {
		w.Connected(EchoWallet())
	}
}

// Call GetWalletHeight if wallet is connected and set height
func (w *wallet) GetHeight() {
	if w.IsConnected() {
		w.height = GetWalletHeight()
	}
}

// Returns all names of wallet.balances map
func (w *wallet) Balances() (all []string) {
	w.RLock()
	defer w.RUnlock()

	for name := range w.balances {
		all = append(all, name)
	}

	return
}

// Returns balance in atomic units
func (w *wallet) Balance(name string) (atomic uint64) {
	w.RLock()
	defer w.RUnlock()

	if w.balances[name] != nil {
		atomic = w.balances[name].atomic
	}

	return
}

// Returns balance string of name formatted to decimal place
func (w *wallet) BalanceF(name string) (balance string) {
	w.RLock()
	defer w.RUnlock()

	if w.balances[name] != nil {
		return w.balances[name].format
	} else {
		return "0.00000"
	}
}

// Returns wallet.height
func (w *wallet) Height() uint64 {
	return w.height
}

// Add a scid with balances data to wallet.balances map
func (w *wallet) AddSCID(name, scid string, decimal int) {
	w.Lock()
	w.balances[name] = &balance{decimal: decimal, scid: scid}
	w.Unlock()
}

// Get DERO balance and all assets in wallet.Balances
func (w *wallet) GetAllBalances() {
	w.Lock()
	defer w.Unlock()

	if w.IsConnected() {
		for name := range w.balances {
			var bal uint64
			if name == "DERO" {
				bal = GetBalance()
			} else {
				bal = GetAssetBalance(w.balances[name].scid)
			}

			w.balances[name].atomic = bal
			w.balances[name].format = FromAtomic(bal, w.balances[name].decimal)
		}

		return
	}

	for name := range w.balances {
		w.balances[name].atomic = 0
		w.balances[name].format = FromAtomic(0, w.balances[name].decimal)
	}
}

// Sync calls wallet.Echo, wallet.GetHeight and wallet.GetAllBalances if wallet is connected
func (w *wallet) Sync() {
	w.Echo()
	w.GetHeight()
	w.GetAllBalances()
}
