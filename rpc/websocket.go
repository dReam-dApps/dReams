package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/creachadair/jrpc2"
	"github.com/deroproject/derohe/walletapi/rpcserver"
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

// Create XSWD application data for DERO connections
//   - 'allow' true will request AlwaysAllow permissions upon connection for the methods used in rpc package
func NewXSWDApplicationData(name, description, URL string, allow bool) *xswd.ApplicationData {
	id := HashToHexSHA256(name)

	// Add prefix for desktop apps
	if !strings.HasPrefix(URL, "http") {
		URL = "https://" + URL
	}

	permissions := make(map[string]xswd.Permission)
	if allow {
		// Methods used in this package
		methods := []string{
			"Echo",
			"GetAddress",
			"HasMethod",
			"GetBalance",
			"GetTransfers",
			"GetTransferbyTXID",
			"GetHeight",
			"transfer",
			"Subscribe",
		}

		permissions = NewXSWDPermissions(methods)
	}

	return &xswd.ApplicationData{
		Id:          id,
		Name:        name,
		Description: description,
		Url:         URL,
		OnClose:     make(chan bool),
		Permissions: permissions,
		Signature:   []byte(id),
	}
}

// Create a new permissions map for XSWD if methods exists in rpcserver.WalletHandler or is XSWD method
func NewXSWDPermissions(methods []string) map[string]xswd.Permission {
	var exists []string
	for _, m := range methods {
		if _, ok := rpcserver.WalletHandler[m]; ok {
			exists = append(exists, m)
			continue
		}

		// xswd methods
		if m == "Subscribe" || m == "Unsubscribe" || m == "HasMethod" {
			exists = append(exists, m)
		}
	}

	permissions := make(map[string]xswd.Permission)
	for _, e := range exists {
		permissions[e] = xswd.AlwaysAllow
	}

	return permissions
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
func (ws *XSWDserver) Init(app *xswd.ApplicationData) (connected bool) {
	if ws.conn == nil {
		ws.app = app

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
