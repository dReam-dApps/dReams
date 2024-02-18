package rpc

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/creachadair/jrpc2"
	"github.com/deroproject/derohe/walletapi/xswd"
	"github.com/gorilla/websocket"
	"github.com/ybbus/jsonrpc/v3"
)

// XSWD server connection vars
type XSWDserver struct {
	Port    string
	connect bool
	request bool
	conn    *websocket.Conn
	app     *xswd.ApplicationData
	sync.RWMutex
}

// Default permissions that dReams will request on connection
func dReamsXSWDApplication() *xswd.ApplicationData {
	return &xswd.ApplicationData{
		Id:          hex.EncodeToString(sha256.New().Sum([]byte("dReams")))[0:64],
		Name:        "dReams",
		Description: "dReamLand",
		Url:         "http://dreamdapps.io",
		OnClose:     make(chan bool),
		Permissions: map[string]xswd.Permission{
			"Echo":              xswd.AlwaysAllow,
			"GetAddress":        xswd.AlwaysAllow,
			"HasMethod":         xswd.AlwaysAllow,
			"GetBalance":        xswd.AlwaysAllow,
			"GetTransfers":      xswd.AlwaysAllow,
			"GetTransferbyTXID": xswd.AlwaysAllow,
			"GetHeight":         xswd.AlwaysAllow,
			"transfer":          xswd.AlwaysAllow,
			"Subscribe":         xswd.AlwaysAllow,
		},
		Signature: []byte("645265616d73e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495"),
	}
}

// Initialize websocket
func CreateSocket(port string) (con *websocket.Conn, err error) {
	u := url.URL{Scheme: "ws", Host: port, Path: "/xswd"}
	logger.Printf("[XSWD] Connecting to %s", u.String())

	header := http.Header{}
	header.Set("content-type", "application/json")

	con, _, err = websocket.DefaultDialer.Dial(u.String(), header)

	return
}

// Read message response and unmarshal toType
func (ws *XSWDserver) readAndUnmarshalWallet(toType interface{}) (err error) {
	var message []byte
	_, message, err = ws.conn.ReadMessage()
	if err != nil {
		return
	}

	var v *xswd.RPCResponse
	err = json.Unmarshal(message, &v)
	if err != nil {
		return
	}

	// Handle response error
	if v.Error != nil {
		var result []byte
		result, err = json.Marshal(v.Error)
		if err != nil {
			return fmt.Errorf("%v", v.Error)
		}

		var e *jrpc2.Error
		err = json.Unmarshal(result, &e)
		if err != nil {
			return fmt.Errorf("%v", v.Error)
		}

		return fmt.Errorf("%s", e.Message)
	}

	js, err := json.Marshal(v.Result)
	if err != nil {
		return
	}

	err = json.Unmarshal(js, toType)
	if err != nil {
		return
	}

	return nil
}

// Connect to websocket
func (ws *XSWDserver) ConnectSocket() (response *xswd.AuthorizationResponse, err error) {
	err = ws.conn.WriteJSON(ws.app)
	if err != nil {
		return
	}

	_, message, err := ws.conn.ReadMessage()
	if err != nil {
		return
	}

	err = json.Unmarshal(message, &response)
	if err != nil {
		return
	}

	return
}

// Initialize and connect to websocket
func (ws *XSWDserver) Init() (connected bool) {
	if ws.conn == nil {
		ws.app = dReamsXSWDApplication()

		var err error
		ws.conn, err = CreateSocket(ws.Port)
		if err != nil {
			logger.Errorln("[XSWD]", err)
			return
		}

		ws.connecting(true)
		m, err := ws.ConnectSocket()
		if err != nil {
			logger.Errorln("[XSWD]", err)
			return
		}

		ws.connecting(false)
		if m.Accepted {
			return true
		}

		logger.Println("[XSWD] Wallet denied connection request")
		ws.conn = nil
	} else {
		ws.conn = nil
	}

	return
}

// Check if XSWD server is closed
func (ws *XSWDserver) IsClosed() bool {
	return ws.conn == nil
}

// Check if connecting to XSWD server
func (ws *XSWDserver) IsConnecting() bool {
	ws.RLock()
	defer ws.RUnlock()

	return ws.connect
}

// Check if requesting from XSWD server
func (ws *XSWDserver) IsRequesting() bool {
	ws.RLock()
	defer ws.RUnlock()

	return ws.request
}

// Set connecting to XSWD server
func (ws *XSWDserver) connecting(b bool) {
	ws.Lock()
	ws.connect = b
	ws.Unlock()
}

// Set requesting to XSWD server
func (ws *XSWDserver) requesting(b bool) {
	ws.Lock()
	ws.request = b
	ws.Unlock()
}

// Wrapper for XSWD calls
func (ws *XSWDserver) CallFor(out interface{}, method string, params interface{}) (err error) {
	if ws.IsClosed() {
		return fmt.Errorf("no connection with XSWD server")
	}

	request := jsonrpc.RPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	}

	ws.requesting(true)
	defer func() {
		ws.requesting(false)
	}()

	err = ws.conn.WriteJSON(request)
	if err != nil {
		return
	}

	return ws.readAndUnmarshalWallet(&out)
}
