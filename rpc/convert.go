package rpc

import (
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
)

// Switch to convert interface to int
func intType(v interface{}) (value int) {
	switch v := v.(type) {
	case uint64:
		value = int(v)
	case float64:
		value = int(v)
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			value = int(i)
		}
	}

	return
}

// Switch to convert interface to float64
func float64Type(v interface{}) (value float64) {
	switch v := v.(type) {
	case uint64:
		value = float64(v)
	case float64:
		value = v
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			value = f
		}
	}

	return
}

// Convert hex value to string
func HexToString(h interface{}) string {
	switch h := h.(type) {
	case string:
		if str, err := hex.DecodeString(h); err == nil {
			return string(str)
		}
	}

	return ""
}

// Returns value plus one as string
func AddOne(v interface{}) string {
	return strconv.Itoa(intType(v) + 1)
}

// Convert a millisecond string to time.Time
func MsToTime(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(0, msInt*int64(time.Millisecond)), nil
}

// Convert string to int, returns 0 if err
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

// Convert string to Uint64, returns 0 if err
func StringToUint64(s string) uint64 {
	if s != "" {
		i, err := strconv.Atoi(s)
		if err != nil {
			log.Println("[StringToUint64]", err)
			return 0
		}
		return uint64(i)
	}

	return 0
}

// Returns uint64 atomic value of v rounding to precision
func ToAtomic(v interface{}, precision float64) uint64 {
	ratio := math.Pow(10, precision)
	rf := math.Round(float64Type(v)*ratio) / ratio

	return uint64(math.Round(rf * 100000))
}

// Returns atomic string value of v rounded to precision, walletapi.FormatMoneyPrecision()
func FromAtomic(v interface{}, precision int) string {
	decimals := new(big.Float).SetInt64(100000)
	float_amount, _, _ := big.ParseFloat(fmt.Sprint(float64Type(v)), 10, 0, big.ToZero)
	result := new(big.Float)
	result.Quo(float_amount, decimals)

	return result.Text('f', precision)
}

// Get Dero address from keys
func DeroAddressFromKey(v interface{}) (address string) {
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

	return
}
