package rpc

import (
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2/widget"
	"github.com/civilware/Gnomon/structures"
	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/transaction"
	"github.com/deroproject/derohe/walletapi"
	"github.com/sirupsen/logrus"
	"github.com/ybbus/jsonrpc/v3"
)

type wallet struct {
	IdHash   string
	Address  string
	balances map[string]*balance
	height   uint64
	Connect  bool
	muC      sync.RWMutex
	sync.RWMutex
	File     Disk
	LogEntry *widget.Entry
	RPC      RPCserver
	WS       XSWDserver
}

type Disk struct {
	disk *walletapi.Wallet_Disk
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

	if w.File.disk != nil {
		logger.Infof("[%s] Wallet Closed\n", tag)
		w.File.disk.Close_Encrypted_Wallet()
		w.File.disk = nil
	}

	w.Connected(false)
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

// Parse transfer params to match derohe rpcserver.Transfer()
func parseTransferParams(p *rpc.Transfer_Params) (err error) {
	for _, t := range p.Transfers {
		_, err = t.Payload_RPC.CheckPack(transaction.PAYLOAD0_LIMIT)
		if err != nil {
			return
		}
	}

	if len(p.SC_Code) >= 1 { // decode SC from base64 if possible, since json has limitations
		if sc, err := base64.StdEncoding.DecodeString(p.SC_Code); err == nil {
			p.SC_Code = string(sc)
		}
	}

	if p.SC_Code != "" && p.SC_ID == "" {
		p.SC_RPC = append(p.SC_RPC, rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_INSTALL)})
		p.SC_RPC = append(p.SC_RPC, rpc.Argument{Name: rpc.SCCODE, DataType: rpc.DataString, Value: p.SC_Code})
	}

	if p.SC_ID != "" {
		p.SC_RPC = append(p.SC_RPC, rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)})
		p.SC_RPC = append(p.SC_RPC, rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: crypto.HashHexToHash(p.SC_ID)})
		if p.SC_Code != "" {
			p.SC_RPC = append(p.SC_RPC, rpc.Argument{Name: rpc.SCCODE, DataType: rpc.DataString, Value: p.SC_Code})
		}
	}

	return
}

// Wallet call switch for RPC, XSWD or walletapi connections
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
	} else if w.File.disk != nil {
		switch method {
		case "transfer":
			result, ok := out.(*rpc.Transfer_Result)
			if !ok {
				return fmt.Errorf("expected out to be *rpc.Transfer_Result, got %T", out)
			}

			if params == nil {
				return fmt.Errorf("params can not be nil for %s", method)
			}

			if p, ok := params[0].(*rpc.Transfer_Params); ok {
				err = parseTransferParams(p)
				if err != nil {
					return
				}

				var tx *transaction.Transaction
				tx, err = w.File.disk.TransferPayload0(p.Transfers, p.Ringsize, false, p.SC_RPC, p.Fees, false)
				if err != nil {
					return
				}

				err = w.File.disk.SendTransaction(tx)
				if err != nil {
					return
				}

				result.TXID = tx.GetHash().String()

			} else {
				err = fmt.Errorf("expected params to be *rpc.Transfer_Params, got %T", params[0])
			}
		case "GetBalance":
			result, ok := out.(*rpc.GetBalance_Result)
			if !ok {
				return fmt.Errorf("expected out to be *rpc.GetBalance_Result, got %T", out)
			}

			if params == nil {
				return fmt.Errorf("params can not be nil for %s", method)
			}

			if p, ok := params[0].(*rpc.GetBalance_Params); ok {
				var unlocked, locked uint64
				if p.SCID.IsZero() {
					unlocked, locked = w.File.disk.Get_Balance()
				} else {
					unlocked, locked = w.File.disk.Get_Balance_scid(p.SCID)
				}

				result.Balance = locked
				result.Unlocked_Balance = unlocked
			} else {
				err = fmt.Errorf("expected out to be *rpc.GetBalance_Params, got %T", params[0])
			}
		case "GetTransfers":
			result, ok := out.(*rpc.Get_Transfers_Result)
			if !ok {
				return fmt.Errorf("expected out to be *rpc.Get_Transfers_Result, got %T", out)
			}

			if params == nil {
				return fmt.Errorf("params can not be nil for %s", method)
			}

			if p, ok := params[0].(*rpc.Get_Transfers_Params); ok {
				result.Entries = w.File.disk.Show_Transfers(p.SCID, p.Coinbase, p.In, p.Out, p.Min_Height, p.Max_Height, p.Sender, p.Receiver, p.DestinationPort, p.SourcePort)
			} else {
				err = fmt.Errorf("expected out to be *rpc.Get_Transfers_Params, got %T", params[0])
			}
		case "GetTransferByTXID":
			result, ok := out.(*rpc.Get_Transfer_By_TXID_Result)
			if !ok {
				return fmt.Errorf("expected out to be *rpc.Get_Transfer_By_TXID_Result, got %T", out)
			}

			if params == nil {
				return fmt.Errorf("params can not be nil for %s", method)
			}

			if p, ok := params[0].(*rpc.Get_Transfer_By_TXID_Params); ok {
				scid, entry := w.File.disk.Get_Payments_TXID(p.SCID, p.TXID)
				result.SCID = scid
				result.Entry = entry
			} else {
				err = fmt.Errorf("expected out to be *rpc.Get_Transfer_By_TXID_Params, got %T", params[0])
			}
		case "GetAddress":
			result, ok := out.(*rpc.GetAddress_Result)
			if !ok {
				return fmt.Errorf("expected out to be *rpc.GetAddress_Result, got %T", out)
			}

			result.Address = w.File.disk.GetAddress().String()

		case "GetHeight":
			result, ok := out.(*rpc.GetHeight_Result)
			if !ok {
				return fmt.Errorf("expected out to be *rpc.GetHeight_Result, got %T", out)
			}

			result.Height = w.File.disk.Get_Height()
		// case "Echo":
		// 	out = "Wallet " + strings.Join(params[0].([]string), " ")

		default:
			err = fmt.Errorf("method %s is not available", method)
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

	if w.RPC.client == nil && w.WS.conn == nil && w.File.disk != nil {
		for name := range w.balances {
			var bal uint64
			if name == "DERO" {
				bal, _ = w.File.disk.Get_Balance()
			} else {
				bal, _ = w.File.disk.Get_Balance_scid(crypto.HashHexToHash(w.balances[name].scid))
			}

			w.balances[name].atomic = bal
			w.balances[name].format = FromAtomic(bal, w.balances[name].decimal)
		}

		return
	}

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
	if w.File.disk != nil {
		w.Lock()
		walletapi.Daemon_Endpoint_Active = Daemon.Rpc
		if err := walletapi.Connect(Daemon.Rpc); err != nil {
			logger.Errorln("[Sync]", err)
			w.Unlock()
			w.Connected(false)
			return
		}

		w.Unlock()
		w.Connected(true)
	} else {
		w.Echo()
	}

	w.GetHeight()
	w.GetAllBalances()
}

func (w *wallet) OpenWalletFile(tag, path, password string) (err error) {
	w.File.disk, err = walletapi.Open_Encrypted_Wallet(path, password)
	if err != nil {
		return
	}

	w.File.disk.SetNetwork(true)
	w.File.disk.SetOnlineMode()
	GetAddress(tag)
	w.Connected(true)

	return
}

// Below are derohe *walletapi.Wallet_Disk methods to expose

func (f *Disk) Encrypt(data []byte) (result []byte, err error) {
	return f.disk.Encrypt(data)
}

func (f *Disk) Decrypt(data []byte) (result []byte, err error) {
	return f.disk.Decrypt(data)
}

func (f *Disk) SignData(data []byte) (result []byte) {
	return f.disk.SignData(data)
}

func (f *Disk) IsNil() bool {
	return f.disk == nil
}
