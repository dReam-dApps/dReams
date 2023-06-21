package holdero

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/SixofClubsss/dReams/menu"
	coingecko "github.com/superoo7/go-gecko/v3"
)

type ogreFeed struct {
	Success      bool   `json:"success"`
	Initialprice string `json:"initialprice"`
	Price        string `json:"price"`
	High         string `json:"high"`
	Low          string `json:"low"`
	Volume       string `json:"volume"`
	Bid          string `json:"bid"`
	Ask          string `json:"ask"`
}

type kuFeed struct {
	Code string `json:"code"`
	Data struct {
		Time        int64  `json:"time"`
		Sequence    string `json:"sequence"`
		Price       string `json:"price"`
		Size        string `json:"size"`
		BestBid     string `json:"bestBid"`
		BestBidSize string `json:"bestBidSize"`
		BestAsk     string `json:"bestAsk"`
		BestAskSize string `json:"bestAskSize"`
	} `json:"data"`
}

// Main price fetch, returns float and display values
//   - Average from 3 feeds, if not take average from 2, if not TO value takes priority spot
func GetPrice(coin string) (price float64, display string) {
	var t float64
	var k float64
	var g float64
	priceT := getOgre(coin)
	priceK := getKucoin(coin)
	priceG := getGeko(coin)

	if menu.CoinDecimal(coin) == 8 {
		if tf, err := strconv.ParseFloat(priceT, 64); err == nil {
			t = tf * 100000000
		}

		if kf, err := strconv.ParseFloat(priceK, 64); err == nil {
			k = kf * 100000000
		}

		if gf, err := strconv.ParseFloat(priceG, 64); err == nil {
			g = gf * 100000000
		}
	} else {
		if tf, err := strconv.ParseFloat(priceT, 64); err == nil {
			t = tf * 100
		}

		if kf, err := strconv.ParseFloat(priceK, 64); err == nil {
			k = kf * 100
		}

		if gf, err := strconv.ParseFloat(priceG, 64); err == nil {
			g = gf * 100
		}
	}

	if t > 0 && k > 0 && g > 0 {
		price = (t + k + g) / 3
	} else if t > 0 && k > 0 {
		price = (t + k) / 2
	} else if k > 0 && g > 0 {
		price = (k + g) / 2
	} else if t > 0 && g > 0 {
		price = (t + g) / 2
	} else if t > 0 {
		price = t
	} else if k > 0 {
		price = k
	} else if g > 0 {
		price = g
	} else {
		price = 0
		log.Println("[dReams] Error getting price feed")
	}

	if menu.CoinDecimal(coin) == 8 {
		display = fmt.Sprintf("%.8f", price/100000000)
	} else {
		display = fmt.Sprintf("%.2f", price/100)
	}

	return
}

// Get TO coin price feed
func getOgre(coin string) string {
	decimal := 2
	var url string
	var found ogreFeed
	switch coin {
	case "BTC-USDT":
		url = "https://tradeogre.com/api/v1/ticker/usdt-btc"
	case "DERO-USDT":
		url = "https://tradeogre.com/api/v1/ticker/usdt-dero"
	case "XMR-USDT":
		url = "https://tradeogre.com/api/v1/ticker/usdt-xmr"
	case "DERO-BTC":
		url = "https://tradeogre.com/api/v1/ticker/btc-dero"
		decimal = 8
	case "XMR-BTC":
		url = "https://tradeogre.com/api/v1/ticker/btc-xmr"
		decimal = 8
	default:
		return ""
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("[getOgre]", err)
		return ""
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Println("[getOgre]", err)
		return ""
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("[getOgre]", err)
		return ""
	}

	json.Unmarshal(b, &found)

	if s, err := strconv.ParseFloat(found.Price, 64); err == nil {
		if decimal == 8 {
			return fmt.Sprintf("%.8f", s)
		}
		return fmt.Sprintf("%.2f", s)
	}

	return found.Price
}

// Get Kucoin coin price feed
func getKucoin(coin string) string {
	decimal := 2
	var url string
	var found kuFeed
	switch coin {
	case "BTC-USDT":
		url = "https://api.kucoin.com/api/v1/market/orderbook/level1?symbol=BTC-USDT"
	case "DERO-USDT":
		url = "https://api.kucoin.com/api/v1/market/orderbook/level1?symbol=DERO-USDT"
	case "XMR-USDT":
		url = "https://api.kucoin.com/api/v1/market/orderbook/level1?symbol=XMR-USDT"
	case "DERO-BTC":
		url = "https://api.kucoin.com/api/v1/market/orderbook/level1?symbol=DERO-BTC"
		decimal = 8
	case "XMR-BTC":
		url = "https://api.kucoin.com/api/v1/market/orderbook/level1?symbol=XMR-BTC"
		decimal = 8
	default:
		return ""
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("[getKucoin]", err)
		return ""
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Println("[getKucoin]", err)
		return ""
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("[getKucoin]", err)
		return ""
	}

	json.Unmarshal(b, &found)

	if s, err := strconv.ParseFloat(found.Data.Price, 64); err == nil {
		if decimal == 8 {
			return fmt.Sprintf("%.8f", s)
		}
		return fmt.Sprintf("%.2f", s)
	}

	return found.Data.Price
}

// Get coingeko price feed
func getGeko(coin string) string {
	client := &http.Client{Timeout: time.Second * 10}
	CG := coingecko.NewClient(client)

	pair := "usd"
	var url string
	switch coin {
	case "BTC-USDT":
		url = "bitcoin"
	case "DERO-USDT":
		url = "dero"
	case "XMR-USDT":
		url = "monero"
	case "DERO-BTC":
		url = "dero"
		pair = "btc"
	case "XMR-BTC":
		url = "monero"
		pair = "btc"
	default:
		return ""
	}

	price, err := CG.SimpleSinglePrice(url, pair)
	if err != nil {
		log.Println("[getGeko]", err)
		return ""
	}

	if pair == "btc" {
		return fmt.Sprintf("%.8f", price.MarketPrice)
	}

	return fmt.Sprintf("%.2f", price.MarketPrice)
}
