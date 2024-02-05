package rpc

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2/widget"
	"github.com/civilware/Gnomon/structures"
	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/walletapi"
	"github.com/sirupsen/logrus"
	"github.com/ybbus/jsonrpc/v3"
)

type wallet struct {
	IdHash    string
	Address   string
	ClientKey string
	Balance   uint64
	TokenBal  map[string]uint64
	Height    int
	Connect   bool
	KeyLock   bool
	muC       sync.RWMutex
	muB       sync.RWMutex
	File      *walletapi.Wallet_Disk
	LogEntry  *widget.Entry
	RPC       RPCserver
	WS        XSWDserver
	Display   struct {
		Balance map[string]string
		Height  string
	}
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
			time.Sleep(100 * time.Millisecond)
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

// Get Wallet.Balance, 0 if not connected
func (w *wallet) GetBalance() {
	w.muB.Lock()
	defer w.muB.Unlock()

	if w.IsConnected() {
		var result *rpc.GetBalance_Result
		if err := w.CallFor(&result, "GetBalance"); err != nil {
			logger.Errorln("[GetBalance]", err)
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
	w.muB.Lock()
	defer w.muB.Unlock()

	if w.IsConnected() {
		params := &rpc.GetBalance_Params{
			SCID: crypto.HashHexToHash(scid),
		}

		var result *rpc.GetBalance_Result
		if err := w.CallFor(&result, "GetBalance", params); err != nil {
			logger.Errorln("[GetTokenBalance]", err)
			w.TokenBal[name] = 0
			return
		}

		w.TokenBal[name] = result.Unlocked_Balance

		return

	}

	w.TokenBal[name] = 0
}

// Read Wallet.Balance
func (w *wallet) ReadBalance() uint64 {
	w.muB.RLock()
	defer w.muB.RUnlock()

	return w.Balance
}

// Read Wallet.TokenBal[name]
func (w *wallet) ReadTokenBalance(name string) uint64 {
	w.muB.RLock()
	defer w.muB.RUnlock()

	return w.TokenBal[name]
}
