package menu

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

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

type xeggexFeed struct {
	TickerID       string `json:"ticker_id"`
	Type           string `json:"type"`
	BaseCurrency   string `json:"base_currency"`
	TargetCurrency string `json:"target_currency"`
	LastPrice      string `json:"last_price"`
	BaseVolume     string `json:"base_volume"`
	TargetVolume   string `json:"target_volume"`
	UsdVolumeEst   string `json:"usd_volume_est"`
	Bid            string `json:"bid"`
	Ask            string `json:"ask"`
	High           string `json:"high"`
	Low            string `json:"low"`
	ChangePercent  string `json:"change_percent"`
}

// Used for placing coin decimal, default returns 2 decimal place
func CoinDecimal(ticker string) int {
	split := strings.Split(ticker, "-")
	if len(split) == 2 {
		switch split[1] {
		case "BTC":
			return 8
		case "DERO":
			return 5
		default:
			return 2
		}
	}

	return 2
}

// Main price fetch, returns float and display string, average from 4 feeds
//   - coin, "BTC-USDT", "DERO-USDT", "XMR-USDT", "DERO-BTC", "XMR-BTC"
func GetPrice(coin, tag string) (price float64, display string) {
	var sum float64
	var count int
	priceT := getOgre(coin)
	priceK := getKucoin(coin)
	priceG := getGeko(coin)
	priceX := getXeggex(coin)

	if CoinDecimal(coin) == 8 {
		if tf, err := strconv.ParseFloat(priceT, 64); err == nil {
			sum += tf * 100000000
			count++
		}

		if kf, err := strconv.ParseFloat(priceK, 64); err == nil {
			sum += kf * 100000000
			count++
		}

		if gf, err := strconv.ParseFloat(priceG, 64); err == nil {
			sum += gf * 100000000
			count++
		}

		if xf, err := strconv.ParseFloat(priceX, 64); err == nil {
			sum += xf * 100000000
			count++
		}
	} else {
		if tf, err := strconv.ParseFloat(priceT, 64); err == nil {
			sum += tf * 100
			count++
		}

		if kf, err := strconv.ParseFloat(priceK, 64); err == nil {
			sum += kf * 100
			count++
		}

		if gf, err := strconv.ParseFloat(priceG, 64); err == nil {
			sum += gf * 100
			count++
		}

		if xf, err := strconv.ParseFloat(priceX, 64); err == nil {
			sum += xf * 100
			count++
		}
	}

	if sum > 0 {
		price = sum / float64(count)
	} else {
		price = 0
		logger.Errorf("[%s] Error getting price feed\n", tag)
	}

	if CoinDecimal(coin) == 8 {
		display = fmt.Sprintf("%.8f", price/100000000)
	} else {
		display = fmt.Sprintf("%.2f", price/100)
	}

	return
}

// Get TO pair feed
func getOgre(coin string) (price string) {
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
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Errorln("[getOgre]", err)
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Errorln("[getOgre]", err)
		return
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorln("[getOgre]", err)
		return
	}

	err = json.Unmarshal(b, &found)
	if err != nil {
		logger.Errorln("[getOgre]", err)
		return
	}

	if s, err := strconv.ParseFloat(found.Price, 64); err == nil {
		if decimal == 8 {
			return fmt.Sprintf("%.8f", s)
		}
		return fmt.Sprintf("%.2f", s)
	}

	return found.Price
}

// Get Kucoin pair feed
func getKucoin(coin string) (price string) {
	decimal := 2
	var url string
	var found kuFeed
	switch coin {
	case "BTC-USDT":
		url = "https://api.kucoin.com/api/v1/market/orderbook/level1?symbol=BTC-USDT"
	case "DERO-USDT":
		// url = "https://api.kucoin.com/api/v1/market/orderbook/level1?symbol=DERO-USDT"
		return
	case "XMR-USDT":
		url = "https://api.kucoin.com/api/v1/market/orderbook/level1?symbol=XMR-USDT"
	case "DERO-BTC":
		// url = "https://api.kucoin.com/api/v1/market/orderbook/level1?symbol=DERO-BTC"
		// decimal = 8
		return
	case "XMR-BTC":
		url = "https://api.kucoin.com/api/v1/market/orderbook/level1?symbol=XMR-BTC"
		decimal = 8
	default:
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Errorln("[getKucoin]", err)
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Errorln("[getKucoin]", err)
		return
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorln("[getKucoin]", err)
		return
	}

	err = json.Unmarshal(b, &found)
	if err != nil {
		logger.Errorln("[getKucoin]", err)
		return
	}

	if s, err := strconv.ParseFloat(found.Data.Price, 64); err == nil {
		if decimal == 8 {
			return fmt.Sprintf("%.8f", s)
		}
		return fmt.Sprintf("%.2f", s)
	}

	return found.Data.Price
}

// Get coingeko pair feed
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
		logger.Errorln("[getGeko]", err)
		return ""
	}

	if pair == "btc" {
		return fmt.Sprintf("%.8f", price.MarketPrice)
	}

	return fmt.Sprintf("%.2f", price.MarketPrice)
}

// Get Xeggex pair feed
func getXeggex(coin string) (price string) {
	decimal := 2
	var url string
	var found xeggexFeed
	switch coin {
	case "BTC-USDT":
		url = "https://api.xeggex.com/api/v2/ticker/BTC/USDT"
	case "DERO-USDT":
		url = "https://api.xeggex.com/api/v2/ticker/DERO/USDT"
	case "XMR-USDT":
		url = "https://api.xeggex.com/api/v2/ticker/XMR/USDT"
	case "DERO-BTC":
		url = "https://api.xeggex.com/api/v2/ticker/DERO/BTC"
		decimal = 8
	case "XMR-BTC":
		url = "https://api.xeggex.com/api/v2/ticker/XMR/BTC"
		decimal = 8
	default:
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Errorln("[getXeggex]", err)
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Errorln("[getXeggex]", err)
		return
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorln("[getXeggex]", err)
		return
	}

	err = json.Unmarshal(b, &found)
	if err != nil {
		logger.Errorln("[getXeggex]", err)
		return
	}

	if s, err := strconv.ParseFloat(found.LastPrice, 64); err == nil {
		if decimal == 8 {
			return fmt.Sprintf("%.8f", s)
		}
		return fmt.Sprintf("%.2f", s)
	}

	return found.LastPrice
}
