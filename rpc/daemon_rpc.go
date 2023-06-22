package rpc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/deroproject/derohe/rpc"
	"github.com/ybbus/jsonrpc/v3"
)

const (
	DREAMSv      = "0.9.6d"
	NameSCID     = "0000000000000000000000000000000000000000000000000000000000000001"
	RatingSCID   = "c66a11ddb22912e92b0a7ab777ed0d343632d9e3c6e8a81452396ca84d2decb6"
	DreamsSCID   = "ad2e7b37c380cc1aed3a6b27224ddfc92a2d15962ca1f4d35e530dba0f9575a9"
	HgcSCID      = "e2e45ce26f70cb551951c855e81a12fee0bb6ebe80ef115c3f50f51e119c02f3"
	TourneySCID  = "c2e1ec16aed6f653aef99a06826b2b6f633349807d01fbb74cc0afb5ff99c3c7"
	HolderoSCID  = "e3f37573de94560e126a9020c0a5b3dfc7a4f3a4fbbe369fba93fbd219dc5fe9"
	pHolderoSCID = "896834d57628d3a65076d3f4d84ddc7c5daf3e86b66a47f018abda6068afe2e6"
	HHolderoSCID = "efe646c48977fd776fee73cdd3df147a2668d3b7d965cdb7a187dda4d23005d8"
	BaccSCID     = "8289c6109f41cbe1f6d5f27a419db537bf3bf30a25eff285241a36e1ae3e48a4"
	PredictSCID  = "eaa62b220fa1c411785f43c0c08ec59c761261cb58a0ccedc5b358e5ed2d2c95"
	PPredictSCID = "e5e49c9a6dc1c0dc8a94429a01bf758e705de49487cbd0b3e3550648d2460cdf"
	SportsSCID   = "ad11377c29a863523c1cc50a33ca13e861cc146a7c0496da58deaa1973e0a39f"
	PSportsSCID  = "fffdc4ea6d157880841feab335ab4755edcde4e60fec2fff661009b16f44fa94"
	TarotSCID    = "a6fc0033327073dd54e448192af929466596fce4d689302e558bc85ea8734a82"
	DerBnbSCID   = "cfbd566d3678dec6e6dfa3a919feae5306ab12af1485e8bcf9320bd5a122b1d3"
	TrvlSCID     = "efacf71e7b5f849653bfa49bfb9dcf7ad3d372944aef33f1e6f54dc95890e3ba"
	GnomonSCID   = "a05395bb0cf77adc850928b0db00eb5ca7a9ccbafd9a38d021c8d299ad5ce1a4"
	DevAddress   = "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn"
	ArtAddress   = "dero1qy0khp9s9yw2h0eu20xmy9lth3zp5cacmx3rwt6k45l568d2mmcf6qgcsevzx"

	DAEMON_RPC_DEFAULT = "127.0.0.1:10102"
	DAEMON_RPC_REMOTE1 = "89.38.99.117:10102"
	DAEMON_RPC_REMOTE2 = "publicrpc1.dero.io:10102"
	DAEMON_RPC_REMOTE3 = "dero-node.mysrv.cloud:10102"
	DAEMON_RPC_REMOTE4 = "derostats.io:10102"
	DAEMON_RPC_REMOTE5 = "85.17.52.28:11012"
	DAEMON_RPC_REMOTE6 = "node.derofoundation.org:11012"
)

type daemon struct {
	Rpc     string
	Connect bool
	Height  uint64
	sync.RWMutex
}

var Daemon daemon
var SCIDs map[string]string
var Display displayStrings
var Signal signals
var Predict predictionValues

func (d *daemon) Connected(b bool) {
	d.Lock()
	d.Connect = b
	d.Unlock()
}

func (d *daemon) IsConnected() bool {
	d.RLock()
	defer d.RUnlock()

	return d.Connect
}

func IsReady() bool {
	if Wallet.IsConnected() && Daemon.IsConnected() {
		return true
	}

	return false
}

// Set daemon rpc client with context and 5 sec cancel
func SetDaemonClient(addr string) (jsonrpc.RPCClient, context.Context, context.CancelFunc) {
	client := jsonrpc.NewClient("http://" + addr + "/json_rpc")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	return client, ctx, cancel
}

// Ping Dero blockchain for connection
func Ping() {
	rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
	defer cancel()

	var result string
	if err := rpcClientD.CallFor(ctx, &result, "DERO.Ping"); err != nil {
		Daemon.Connected(false)
		return
	}

	if result == "Pong " {
		Daemon.Connected(true)
	} else {
		Daemon.Connected(false)
	}
}

// Get a daemons height
func DaemonHeight(tag, ep string) uint64 {
	rpcClientD, ctx, cancel := SetDaemonClient(ep)
	defer cancel()

	var result *rpc.GetHeight_Result
	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetHeight"); err != nil {
		log.Printf("[%s] %s\n", tag, err)
		return 0
	}

	return result.Height
}

// SC call gas estimate, 1320 Deri max
//   - tag for log print
//   - Pass args and transfers for call
func GasEstimate(scid, tag string, args rpc.Arguments, t []rpc.Transfer, max uint64) uint64 {
	rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
	defer cancel()

	var result *rpc.GasEstimate_Result

	arg1 := rpc.Argument{Name: "SC_ACTION", DataType: "U", Value: 0}
	arg2 := rpc.Argument{Name: "SC_ID", DataType: "H", Value: scid}
	args = append(args, arg1)
	args = append(args, arg2)
	params := rpc.GasEstimate_Params{
		Transfers: t,
		SC_Value:  0,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Signer:    Wallet.Address,
	}

	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetGasEstimate", params); err != nil {
		log.Println(tag, err)
		return 0
	}

	log.Println(tag+" Gas Fee:", result.GasStorage+120)

	if result.GasStorage < max {
		return result.GasStorage + 120
	}

	return max + 120
}

// Get single string key result from SCID with daemon input
func FindStringKey(scid, key, daemon string) interface{} {
	rpcClientD, ctx, cancel := SetDaemonClient(daemon)
	defer cancel()

	var result *rpc.GetSC_Result
	params := rpc.GetSC_Params{
		SCID:      scid,
		Code:      false,
		Variables: true,
	}

	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
		log.Println("[FindStringKey]", err)
		return nil
	}

	return result.VariableStringKeys[key]
}

// Get list of dReams dApps from contract store
//   - Uses remote daemon if no Daemon.Connect
func FetchDapps() (dApps []string) {
	dApps = []string{"Holdero", "Baccarat", "dSports and dPredictions", "Iluma", "DerBnb"}
	var daemon string
	if !Daemon.IsConnected() {
		daemon = DAEMON_RPC_REMOTE5
	} else {
		daemon = Daemon.Rpc
	}

	if stored, ok := FindStringKey(RatingSCID, "dApps", daemon).(string); ok {
		if h, err := hex.DecodeString(stored); err == nil {
			json.Unmarshal(h, &dApps)
		}
	}

	return
}

// Get platform fees from on chain store
//   - Overwrites default fee values with current stored values
func FetchFees() {
	if fee, ok := FindStringKey(RatingSCID, "ContractUnlock", Daemon.Rpc).(float64); ok {
		UnlockFee = uint64(fee)
	} else {
		log.Println("[FetchFees] Could not get current contract unlock fee, using default")
	}

	if fee, ok := FindStringKey(RatingSCID, "ListingFee", Daemon.Rpc).(float64); ok {
		ListingFee = uint64(fee)
	} else {
		log.Println("[FetchFees] Could not get current listing fee, using default")
	}

	if fee, ok := FindStringKey(TarotSCID, "Fee", Daemon.Rpc).(float64); ok {
		IlumaFee = uint64(fee)
	} else {
		log.Println("[FetchFees] Could not get current Iluma fee, using default")
	}

	if fee, ok := FindStringKey(RatingSCID, "LowLimitFee", Daemon.Rpc).(float64); ok {
		LowLimitFee = uint64(fee)
	} else {
		log.Println("[FetchFees] Could not get current low fee limit, using default")
	}

	if fee, ok := FindStringKey(RatingSCID, "HighLimitFee", Daemon.Rpc).(float64); ok {
		HighLimitFee = uint64(fee)
	} else {
		log.Println("[FetchFees] Could not get current high fee limit, using default")
	}
}

// Check Gnomon SC for stored contract owner
func CheckForIndex(scid string) interface{} {
	rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
	defer cancel()

	var result *rpc.GetSC_Result
	params := rpc.GetSC_Params{
		SCID:      GnomonSCID,
		Code:      false,
		Variables: true,
	}

	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
		log.Println("[CheckForIndex]", err)
		return nil
	}

	return DeroAddressFromKey(result.VariableStringKeys[scid+"owner"])
}

// Get code of a SC
func GetSCCode(scid string) string {
	if Daemon.IsConnected() {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      scid,
			Code:      true,
			Variables: false,
		}

		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[GetSCCode]", err)
			return ""
		}

		return result.Code
	}
	return ""
}

// Get name service SC code
func GetNameServiceCode() string {
	if Daemon.IsConnected() {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      NameSCID,
			Code:      true,
			Variables: false,
		}

		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[GetNameServiceCode]", err)
			return ""
		}

		return result.Code
	}
	return ""
}

// Get Gnomon SC code
func GetGnomonCode() string {
	if Daemon.IsConnected() {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      GnomonSCID,
			Code:      true,
			Variables: false,
		}

		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[GetGnomonCode]", err)
			return ""
		}

		return result.Code
	}
	return ""
}

// Get all asset SCIDs from collection
func GetG45Collection(scid string) (scids []string) {
	rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
	defer cancel()

	var result *rpc.GetSC_Result
	params := rpc.GetSC_Params{
		SCID:      scid,
		Code:      false,
		Variables: true,
	}

	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
		log.Println("[GetG45Collection]", err)
		return nil
	}

	i := 0
	for {
		n := strconv.Itoa(i)
		asset := result.VariableStringKeys["assets_"+n]

		if asset == nil {
			break
		} else {
			if hx, err := hex.DecodeString(fmt.Sprint(asset)); err != nil {
				log.Println("[GetG45Collection]", err)
				i++
			} else {
				split := strings.Split(string(hx), ",")
				for i := range split {
					sc := strings.Split(split[i], ":")
					trim := strings.Trim(sc[0], `{"`)
					scids = append(scids, trim)
				}
				i++
			}
		}
	}

	return
}

// Get single TX data with GetTransaction
func GetDaemonTx(txid string) *rpc.Tx_Related_Info {
	rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
	defer cancel()

	var result *rpc.GetTransaction_Result
	params := rpc.GetTransaction_Params{
		Tx_Hashes: []string{txid},
	}

	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetTransaction", params); err != nil {
		log.Println("[GetDaemonTx]", err)
		return nil
	}

	if result.Txs != nil {
		return &result.Txs[0]
	}

	return nil
}

// Verify TX signer with GetTransaction
func VerifySigner(txid string) bool {
	rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
	defer cancel()

	var result *rpc.GetTransaction_Result
	params := rpc.GetTransaction_Params{
		Tx_Hashes: []string{txid},
	}

	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetTransaction", params); err != nil {
		log.Println("[VerifySigner]", err)
		return false
	}

	return result.Txs[0].Signer == Wallet.Address
}

// Code for v1.0.0 Holdero SC
func GetHoldero100Code() string {
	if Daemon.IsConnected() {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
		defer cancel()

		var result *rpc.GetSC_Result
		params := rpc.GetSC_Params{
			SCID:      "95e69b382044ddc1467e030a80905cf637729612f65624e8d17bf778d4362b8d",
			Code:      true,
			Variables: false,
		}

		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[GetHoldero100Code]", err)
			return ""
		}

		return result.Code

	}

	return ""
}

// Code for v1.1.0 Holdero public or private SC
//   - version defines which type of Holdero contract
//   - 0 for standard public
//   - 1 for standard private
//   - 2 for HGC
func GetHoldero110Code(version int) string {
	if Daemon.IsConnected() {
		rpcClientD, ctx, cancel := SetDaemonClient(Daemon.Rpc)
		defer cancel()

		var result *rpc.GetSC_Result
		var params rpc.GetSC_Params
		switch version {
		case 0:
			params = rpc.GetSC_Params{
				SCID:      HolderoSCID,
				Code:      true,
				Variables: false,
			}
		case 1:
			params = rpc.GetSC_Params{
				SCID:      pHolderoSCID,
				Code:      true,
				Variables: false,
			}
		case 2:
			params = rpc.GetSC_Params{
				SCID:      HHolderoSCID,
				Code:      true,
				Variables: false,
			}
		default:

		}

		if err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			log.Println("[GetHoldero110Code]", err)
			return ""
		}

		return result.Code

	}

	return ""
}

// Get difficulty from a daemon
func GetDifficulty(ep string) float64 {
	rpcClientD, ctx, cancel := SetDaemonClient(ep)
	defer cancel()

	var result *rpc.GetInfo_Result
	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetInfo"); err != nil {
		log.Println("[GetDifficulty]", err)
		return 0
	}

	return float64(result.Difficulty)
}

// Get average block time from a daemon
func GetBlockTime(ep string) float64 {
	rpcClientD, ctx, cancel := SetDaemonClient(ep)
	defer cancel()

	var result *rpc.GetInfo_Result
	if err := rpcClientD.CallFor(ctx, &result, "DERO.GetInfo"); err != nil {
		log.Println("[GetBlockTime]", err)
		return 0
	}

	return float64(result.AverageBlockTime50)
}
