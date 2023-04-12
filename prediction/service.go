package prediction

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/SixofClubsss/dReams/holdero"
	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"
	dero "github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/walletapi"
	"go.etcd.io/bbolt"
)

// block height of last payload format chance
const PAYLOAD_FORMAT = uint64(1728000)

type service struct {
	Dest_port  uint64
	Debug      bool
	Processing bool
	Last_block int
}

type printColors struct {
	Reset  string
	Yellow string
	Green  string
	Red    string
}

var Service service
var PrintColor printColors

func SetPrintColors(os string) {
	if os != "windows" {
		PrintColor.Reset = "\033[0m"
		PrintColor.Yellow = "\033[33m"
		PrintColor.Green = "\033[32m"
		PrintColor.Red = "\033[31m"
	}
}

func integratedAddress() (uint64, *dero.Address) {
	var err error
	var addr *dero.Address
	if addr, err = dero.NewAddress(rpc.Wallet.Address); err != nil {
		log.Printf("\n[integratedAddress] address could not be parsed: addr:%s err:%s\n", rpc.Wallet.Address, err)
		return 0, nil
	}

	shasum := fmt.Sprintf("%x", sha1.Sum([]byte(addr.String())))
	b := []byte(shasum)

	return binary.BigEndian.Uint64(b), addr
}

func serviceDebug(print bool, tag, str string) {
	if print && Service.Debug {
		log.Println(tag, str)
	}
}

func intgPredictionArgs(scid string, print bool) (higher_arg dero.Arguments, lower_arg dero.Arguments) {
	higher_string := "Higher  "
	lower_string := "Lower  "
	var p_amt []uint64
	var end uint64
	var pre, mark string
	if menu.Gnomes.Init {
		_, init := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_init", menu.Gnomes.Indexer.ChainHeight, true)
		if init != nil && init[0] == 1 {
			predicting, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "predicting", menu.Gnomes.Indexer.ChainHeight, true)
			_, p_end := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_end_at", menu.Gnomes.Indexer.ChainHeight, true)
			_, p_mark := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "mark", menu.Gnomes.Indexer.ChainHeight, true)
			_, p_amt = menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_amount", menu.Gnomes.Indexer.ChainHeight, true)
			if predicting != nil && p_end != nil {
				pre = predicting[0] + "  "
				end = p_end[0]
				if p_mark != nil {
					if isOnChainPrediction(predicting[0]) {
						switch onChainPrediction(predicting[0]) {
						case 2:
							div := float64(p_mark[0]) / 100000
							mark = fmt.Sprintf("%.5f", div) + "  "
						default:
							mark = fmt.Sprintf("%d", p_mark[0]) + "  "

						}
					} else {
						if predicting[0] == "DERO-BTC" || predicting[0] == "XMR-BTC" {
							div := float64(p_mark[0]) / 100000000
							mark = fmt.Sprintf("%.8f", div) + "  "
						} else {
							div := float64(p_mark[0]) / 100
							mark = fmt.Sprintf("%.2f", div) + "  "
						}
					}
				} else {
					mark = "0  "
				}

				ensn := time.Unix(int64(end), 0).UTC()
				format := ensn.Format("2006-01-02 15:04 UTC")

				chopped_scid := scid[:6] + "..." + scid[58:] + "  "

				higher := "p  " + pre + mark + higher_string + chopped_scid + format
				lower := "p  " + pre + mark + lower_string + chopped_scid + format

				amt := uint64(0)
				if p_amt != nil && p_amt[0] != 0 {
					amt = p_amt[0]
				}

				if amt < 1 {
					serviceDebug(print, "[intgPredictionArgs]", fmt.Sprintf("%s Amount less than 1", scid))
					return
				}

				higher_arg = dero.Arguments{
					{Name: dero.RPC_DESTINATION_PORT, DataType: dero.DataUint64, Value: Service.Dest_port},
					{Name: dero.RPC_COMMENT, DataType: dero.DataString, Value: higher},
					{Name: dero.RPC_NEEDS_REPLYBACK_ADDRESS, DataType: dero.DataUint64, Value: uint64(0)},
					{Name: dero.RPC_VALUE_TRANSFER, DataType: dero.DataUint64, Value: amt},
				}

				lower_arg = dero.Arguments{
					{Name: dero.RPC_DESTINATION_PORT, DataType: dero.DataUint64, Value: Service.Dest_port},
					{Name: dero.RPC_COMMENT, DataType: dero.DataString, Value: lower},
					{Name: dero.RPC_NEEDS_REPLYBACK_ADDRESS, DataType: dero.DataUint64, Value: uint64(0)},
					{Name: dero.RPC_VALUE_TRANSFER, DataType: dero.DataUint64, Value: amt},
				}
			} else {
				if Service.Debug {
					serviceDebug(print, "[intgPredictionArgs]", fmt.Sprintf("%s Could not get prediction info", scid))
				}
			}
		} else {
			if Service.Debug {
				serviceDebug(print, "[intgPredictionArgs]", fmt.Sprintf("%s Not initalized", scid))
			}
		}
	} else {
		if Service.Debug {
			serviceDebug(print, "[intgPredictionArgs]", "Gnomon is not initalized")
		}
	}

	return
}

func intgSportsArgs(scid string, print bool) (args [][]dero.Arguments) {
	var end uint64
	var league, game, a_string, b_string string
	if menu.Gnomes.Init {
		_, init := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_init", menu.Gnomes.Indexer.ChainHeight, true)
		_, played := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_played", menu.Gnomes.Indexer.ChainHeight, true)
		if init != nil && played != nil {
			if init[0] > played[0] {
				iv := uint64(0)
				_, hl := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "hl", menu.Gnomes.Indexer.ChainHeight, true)
				if hl != nil && played[0] > hl[0]*2 {
					iv = played[0] - hl[0]*2
				}

				for {
					iv++

					if iv > init[0] {
						break
					}

					v := strconv.Itoa(int(iv))
					_, s_init := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_init_"+v, menu.Gnomes.Indexer.ChainHeight, true)
					if s_init != nil && s_init[0] == 1 {
						s_game, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "game_"+v, menu.Gnomes.Indexer.ChainHeight, true)
						s_league, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "league_"+v, menu.Gnomes.Indexer.ChainHeight, true)
						_, s_end := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_end_at_"+v, menu.Gnomes.Indexer.ChainHeight, true)
						if s_game != nil && s_end != nil && s_league != nil {
							league = s_league[0] + "  "
							game = s_game[0] + "  "
							end = s_end[0]
							team_a := menu.TrimTeamA(game)
							team_b := menu.TrimTeamB(game)

							if team_a != "" && team_b != "" {
								a_string = team_a + "  "
								b_string = team_b
							} else {
								serviceDebug(print, "[intgSportsArgs]", fmt.Sprintf("%s Could not get team info", scid))
								continue
							}

						} else {
							serviceDebug(print, "[intgSportsArgs]", fmt.Sprintf("%s Could not get league/game info", scid))
							continue
						}
					} else {
						continue
					}

					utc := time.Unix(int64(end), 0).UTC()
					format := utc.Format("2006-01-02 15:04 UTC")

					chopped_scid := scid[:6] + "..." + scid[58:] + "  "

					team_a := "s" + v + "  " + league + game + a_string + chopped_scid + format
					team_b := "s" + v + "  " + league + game + b_string + chopped_scid + format

					_, s_amt := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_amount_"+v, menu.Gnomes.Indexer.ChainHeight, true)
					amt := uint64(0)
					if s_amt != nil && s_amt[0] != 0 {
						amt = s_amt[0]
					} else {
						serviceDebug(print, "[intgSportsArgs]", "Could not get amount")
						continue
					}

					a_arg := dero.Arguments{
						{Name: dero.RPC_DESTINATION_PORT, DataType: dero.DataUint64, Value: Service.Dest_port},
						{Name: dero.RPC_COMMENT, DataType: dero.DataString, Value: team_a},
						{Name: dero.RPC_NEEDS_REPLYBACK_ADDRESS, DataType: dero.DataUint64, Value: uint64(0)},
						{Name: dero.RPC_VALUE_TRANSFER, DataType: dero.DataUint64, Value: amt},
					}

					b_arg := dero.Arguments{
						{Name: dero.RPC_DESTINATION_PORT, DataType: dero.DataUint64, Value: Service.Dest_port},
						{Name: dero.RPC_COMMENT, DataType: dero.DataString, Value: team_b},
						{Name: dero.RPC_NEEDS_REPLYBACK_ADDRESS, DataType: dero.DataUint64, Value: uint64(0)},
						{Name: dero.RPC_VALUE_TRANSFER, DataType: dero.DataUint64, Value: amt},
					}

					var move []dero.Arguments
					move = append(move, a_arg, b_arg)

					args = append(args, move)
				}
			} else {
				if Service.Debug {
					serviceDebug(print, "[intgSportsArgs]", fmt.Sprintf("%s No games initialized", scid))
				}
			}
		} else {
			if Service.Debug {
				serviceDebug(print, "[intgSportsArgs]", fmt.Sprintf("%s No contract info", scid))
			}
		}
	}

	return
}

func makeIntegratedAddr(print bool) {
	var addr *dero.Address
	Service.Dest_port, addr = integratedAddress()
	if addr == nil {
		log.Println("[makeIntegratedAddr] Could not make addresses")
		return
	}

	service_address := addr.Clone()

	var p_contracts, s_contracts []string
	for _, sc := range menu.Control.Predict_owned {
		split := strings.Split(sc, "   ")
		if len(split) > 2 {
			p_contracts = append(p_contracts, split[2])
		}
	}

	for _, sc := range menu.Control.Sports_owned {
		split := strings.Split(sc, "   ")
		if len(split) > 2 {
			s_contracts = append(s_contracts, split[2])
		}
	}

	var live bool
	for _, sc := range p_contracts {
		higher, lower := intgPredictionArgs(sc, print)
		if higher != nil && lower != nil {
			if print {
				log.Println("[makeIntegratedAddr]", fmt.Sprintf("%d DST Port", higher.Value(dero.RPC_DESTINATION_PORT, dero.DataUint64)))
			}

			service_address.Arguments = higher
			comment := higher.Value(dero.RPC_COMMENT, dero.DataString)
			if print {
				log.Println("[makeIntegratedAddr]", fmt.Sprintf("%s %s \n%s\n", walletapi.FormatMoney(higher.Value(dero.RPC_VALUE_TRANSFER, dero.DataUint64).(uint64)), comment, service_address.String()))
			}

			service_address.Arguments = lower
			comment = lower.Value(dero.RPC_COMMENT, dero.DataString)
			if print {
				log.Println("[makeIntegratedAddr]", fmt.Sprintf("%s %s \n%s\n", walletapi.FormatMoney(lower.Value(dero.RPC_VALUE_TRANSFER, dero.DataUint64).(uint64)), comment, service_address.String()))
			}
			live = true
		}
	}

	for _, sc := range s_contracts {
		all_args := intgSportsArgs(sc, true)
		for _, arg := range all_args {
			if print {
				log.Println("[makeIntegratedAddr]", fmt.Sprintf("%d DST Port", arg[0].Value(dero.RPC_DESTINATION_PORT, dero.DataUint64)))
			}

			service_address.Arguments = arg[0]
			comment := arg[0].Value(dero.RPC_COMMENT, dero.DataString)
			if print {
				log.Println("[makeIntegratedAddr]", fmt.Sprintf("%s %s \n%s\n", walletapi.FormatMoney(arg[0].Value(dero.RPC_VALUE_TRANSFER, dero.DataUint64).(uint64)), comment, service_address.String()))
			}

			service_address.Arguments = arg[1]
			comment = arg[1].Value(dero.RPC_COMMENT, dero.DataString)
			if print {
				log.Println("[makeIntegratedAddr]", fmt.Sprintf("%s %s \n%s\n", walletapi.FormatMoney(arg[1].Value(dero.RPC_VALUE_TRANSFER, dero.DataUint64).(uint64)), comment, service_address.String()))
			}
			live = true
		}
	}

	if !live {
		if print {
			log.Println("[makeIntegratedAddr]", "No addresses")
		}
	}
}

func DreamService(start uint64, payouts, transfers bool) {
	if rpc.Daemon.Connect && rpc.Wallet.Connect {
		db := boltDB()
		if db == nil {
			log.Println("[dReamService] Closing")
			return
		}
		defer db.Close()

		err := db.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte("BET"))
			return err
		})

		if err != nil {
			log.Printf("[dReamService] err creating bucket. err %s\n", err)
			return
		}

		if start == 0 {
			start = rpc.DaemonHeight(rpc.Daemon.Rpc)
		}

		if start > 0 {
			log.Println("[dReamService] Processing from height", start)
			for i := 5; i > 0; i-- {
				if !rpc.Wallet.Service {
					break
				}
				log.Println("[dReamService] Starting in", i)
				time.Sleep(1 * time.Second)
			}

			if rpc.Wallet.Service {
				log.Println("[dReamService] Starting")
			}

			for rpc.Wallet.Service && rpc.Wallet.Connect && rpc.Daemon.Connect {
				Service.Processing = true
				if transfers {
					processBetTx(start, db, Service.Debug)
				}

				if payouts {
					runPredictionPayouts(Service.Debug)
					runSportsPayouts(Service.Debug)
				}

				for i := 0; i < 10; i++ {
					time.Sleep(1 * time.Second)
					if !rpc.Wallet.Service || !rpc.Wallet.Connect || !rpc.Daemon.Connect {
						break
					}
				}
			}
			Service.Processing = false
			log.Println("[dReamService] Shutting down")
		} else {
			log.Println("[dReamService] Not starting from 0 height")
		}
		log.Println("[dReamService] Done")
	}
	rpc.Wallet.Service = false
}

func runPredictionPayouts(print bool) {
	contracts := menu.Control.Predict_owned
	var pay_queue, post_queue []string
	for i := range contracts {
		if !menu.Gnomes.Init || menu.GnomonClosing() {
			return
		}
		split := strings.Split(contracts[i], "   ")
		if len(split) > 2 {
			_, u := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(split[2], "p_init", menu.Gnomes.Indexer.ChainHeight, true)
			if u != nil {
				if u[0] == 1 {
					serviceDebug(print, "[runPredictionPayouts]", fmt.Sprintf("%s Live", split[2]))
					now := uint64(time.Now().Unix())
					_, end := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(split[2], "p_end_at", menu.Gnomes.Indexer.ChainHeight, true)
					_, time_a := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(split[2], "time_a", menu.Gnomes.Indexer.ChainHeight, true)
					_, time_c := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(split[2], "time_c", menu.Gnomes.Indexer.ChainHeight, true)
					_, mark := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(split[2], "mark", menu.Gnomes.Indexer.ChainHeight, true)
					predict, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(split[2], "predicting", menu.Gnomes.Indexer.ChainHeight, true)
					if end != nil && time_c != nil {
						if now >= end[0]+time_c[0] {
							serviceDebug(print, "[runPredictionPayouts]", "Adding for payout")
							pay_queue = append(pay_queue, split[2])
						} else {
							serviceDebug(print, "[runPredictionPayouts]", fmt.Sprintf("%s Not ready for payout", predict[0]))
						}

						if time_a != nil && mark == nil {
							if now >= end[0] && now <= end[0]+time_a[0] {
								post_queue = append(post_queue, split[2])
							}
						}
					}
				} else {
					serviceDebug(print, "[runPredictionPayouts]", fmt.Sprintf("%s Not live", split[2]))
				}
			}
		}
	}

	for _, sc := range post_queue {
		var sent bool
		var value float64
		GetPrediction(rpc.Daemon.Connect, sc)
		pre := rpc.Display.Prediction
		if isOnChainPrediction(pre) {
			switch onChainPrediction(pre) {
			case 1:
				value = rpc.GetDifficulty(rpc.Display.P_feed)
			case 2:
				value = rpc.GetBlockTime(rpc.Display.P_feed)
			case 3:
				d := rpc.DaemonHeight(rpc.Display.P_feed)
				value = float64(d)
			default:

			}

			if value > 0 {
				sent = true
				switch onChainPrediction(pre) {
				case 1:
					rpc.PostPrediction(sc, int(value))
				case 2:
					rpc.PostPrediction(sc, int(value*100000))
				case 3:
					rpc.PostPrediction(sc, int(value))
				default:
					sent = false
				}

			} else {
				serviceDebug(print, "[runPredictionPayouts]", "0 value from node, not sending")
			}

		} else {
			value, _ = holdero.GetPrice(pre)
			if value > 0 {
				sent = true
				rpc.PostPrediction(sc, int(value))
			} else {
				serviceDebug(print, "[runPredictionPayouts]", "0 price, not posting")
			}
		}

		if sent {
			t := 0
			Service.Last_block = rpc.Wallet.Height
			serviceDebug(print, "[runPredictionPayouts]", "Tx Delay")
			for rpc.Wallet.Height < Service.Last_block+3 {
				if !rpc.Wallet.Service || !rpc.Wallet.Connect || !rpc.Daemon.Connect || t > 36 {
					break
				}

				t++
				time.Sleep(1 * time.Second)
			}
		}
	}

	for _, sc := range pay_queue {
		serviceDebug(print, "[runPredictionPayouts]", fmt.Sprintf("%s Paying out", sc))
		var sent bool
		var amt float64
		GetPrediction(rpc.Daemon.Connect, sc)
		pre := rpc.Display.Prediction
		if isOnChainPrediction(pre) {
			sent = true
			switch onChainPrediction(rpc.Display.Prediction) {
			case 1:
				amt = rpc.GetDifficulty(rpc.Display.P_feed)
			case 2:
				amt = rpc.GetBlockTime(rpc.Display.P_feed)
			case 3:
				d := rpc.DaemonHeight(rpc.Display.P_feed)
				amt = float64(d)
			default:
				sent = false

			}

			if amt > 0 {
				sent = true
				switch onChainPrediction(pre) {
				case 1:
					rpc.EndPrediction(sc, int(amt))
				case 2:
					rpc.EndPrediction(sc, int(amt*100000))
				case 3:
					rpc.EndPrediction(sc, int(amt))
				default:
					sent = false
				}

			} else {
				serviceDebug(print, "[runPredictionPayouts]", "0 value from node, not sending")
			}

		} else {
			amt, _ = holdero.GetPrice(pre)
			if amt > 0 {
				rpc.EndPrediction(sc, int(amt))
				sent = true
			} else {
				serviceDebug(print, "[runPredictionPayouts]", "0 price, not sending")
			}
		}

		if sent {
			t := 0
			Service.Last_block = rpc.Wallet.Height
			serviceDebug(print, "[runPredictionPayouts]", "Tx Delay")
			for rpc.Wallet.Height < Service.Last_block+3 {
				if !rpc.Wallet.Service || !rpc.Wallet.Connect || !rpc.Daemon.Connect || t > 36 {
					break
				}

				t++
				time.Sleep(1 * time.Second)
			}
		}
	}
}

func runSportsPayouts(print bool) {
	contracts := menu.Control.Sports_owned
	for i := range contracts {
		if !menu.Gnomes.Init || menu.GnomonClosing() {
			return
		}
		split := strings.Split(contracts[i], "   ")
		if len(split) > 2 {
			_, init := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(split[2], "s_init", menu.Gnomes.Indexer.ChainHeight, true)
			_, played := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(split[2], "s_played", menu.Gnomes.Indexer.ChainHeight, true)
			if init != nil && played != nil {
				if init[0] > played[0] {
					serviceDebug(print, "[runSportsPayouts]", fmt.Sprintf("%s Live games", split[2]))
					for iv := uint64(1); iv <= init[0]; iv++ {
						num := strconv.Itoa(int(iv))
						game, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(split[2], "game_"+num, menu.Gnomes.Indexer.ChainHeight, true)
						league, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(split[2], "league_"+num, menu.Gnomes.Indexer.ChainHeight, true)
						_, end := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(split[2], "s_end_at_"+num, menu.Gnomes.Indexer.ChainHeight, true)
						_, add := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(split[2], "time_a", menu.Gnomes.Indexer.ChainHeight, true)
						if game != nil && end != nil && add != nil && league != nil {
							if end[0]+add[0] < uint64(time.Now().Unix()) {
								log.Println("[runSportsPayouts] Paying out")
								var win string
								var sent bool
								if league[0] == "Bellator" || league[0] == "UFC" {
									win, _ = GetMmaWinner(game[0], league[0])
								} else {
									win, _ = GetWinner(game[0], league[0])
								}

								if win != "" {
									rpc.EndSports(split[2], num, win)
									sent = true
								} else {
									serviceDebug(print, "[runSportsPayouts]", "Could not get winner")
								}

								if sent {
									t := 0
									Service.Last_block = rpc.Wallet.Height
									serviceDebug(print, "[runSportsPayouts]", "Tx Delay")
									for rpc.Wallet.Height < Service.Last_block+3 {
										if !rpc.Wallet.Service || !rpc.Wallet.Connect || !rpc.Daemon.Connect || t > 36 {
											break
										}

										t++
										time.Sleep(1 * time.Second)
									}
								}
							} else {
								serviceDebug(print, "[runSportsPayouts]", fmt.Sprintf("%s Not ready for payout", game[0]))
							}
						}
					}
				} else {
					serviceDebug(print, "[runSportsPayouts]", fmt.Sprintf("%s Nothing live", split[2]))
				}
			}
		}
	}
}

func processBetTx(start uint64, db *bbolt.DB, print bool) {
	rpcClient, _, _ := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)

	var p_contracts, s_contracts []string
	for _, sc := range menu.Control.Predict_owned {
		split := strings.Split(sc, "   ")
		if len(split) > 2 {
			p_contracts = append(p_contracts, split[2])
		}
	}

	for _, sc := range menu.Control.Sports_owned {
		split := strings.Split(sc, "   ")
		if len(split) > 2 {
			s_contracts = append(s_contracts, split[2])
		}
	}

	var all_args []dero.Arguments
	for _, sc := range p_contracts {
		higher, lower := intgPredictionArgs(sc, Service.Debug)
		if higher != nil && lower != nil {
			all_args = append(all_args, higher, lower)
		}
	}

	for _, sc := range s_contracts {
		sports := intgSportsArgs(sc, Service.Debug)
		for _, arg := range sports {
			all_args = append(all_args, arg...)
		}
	}

	out_params := dero.Get_Transfers_Params{
		Coinbase:   false,
		In:         false,
		Out:        true,
		Min_Height: PAYLOAD_FORMAT,
	}

	var outgoing dero.Get_Transfers_Result
	err := rpcClient.CallFor(context.TODO(), &outgoing, "GetTransfers", out_params)
	if err != nil {
		log.Println("[viewProcessedTx]", err)
		return
	}

	reply_id := checkReplies(outgoing)

	if start < PAYLOAD_FORMAT {
		start = PAYLOAD_FORMAT
	}

	params := dero.Get_Transfers_Params{
		Coinbase:        false,
		In:              true,
		Out:             false,
		Min_Height:      start,
		DestinationPort: Service.Dest_port,
	}

	var transfers dero.Get_Transfers_Result
	err = rpcClient.CallFor(context.TODO(), &transfers, "GetTransfers", params)
	if err != nil {
		log.Println("[processBetTx]", err)
		return
	}

	l := len(transfers.Entries)

	serviceDebug(print, "[processBetTx]", fmt.Sprintf("%d Entries since Height %d", l, start))

	for i, e := range transfers.Entries {
		if !rpc.Wallet.Service {
			break
		}

		if e.Coinbase || !e.Incoming {
			serviceDebug(print, "[processBetTx]", fmt.Sprintf("%s Coinbase or outgoing", e.TXID))
			continue
		}

		var already_processed bool
		db.View(func(tx *bbolt.Tx) error {
			if b := tx.Bucket([]byte("BET")); b != nil {
				if ok := b.Get([]byte(e.TXID)); ok != nil {
					already_processed = true
				}
			}
			return nil
		})

		if already_processed {
			if i > l-10 {
				serviceDebug(print, "[processBetTx]", fmt.Sprintf(PrintColor.Green+"%s Received: %d Already processed"+PrintColor.Reset, e.TXID, e.Height))
				if print {
					var reply_txid string
					for id, repTx := range reply_id {
						if id == e.TXID[:6]+"..."+e.TXID[58:] {
							reply_txid = repTx
							serviceDebug(print, "[processBetTx]", fmt.Sprintf(PrintColor.Yellow+"Replied: %s"+PrintColor.Reset, reply_txid))
							break
						}
					}

					if len(reply_txid) != 64 {
						serviceDebug(print, "[processBetTx]", fmt.Sprintf(PrintColor.Red+"Reply missing for %d blocks"+PrintColor.Reset, rpc.Wallet.Height-int(e.Height)))
					}
				}
			}
			continue
		}

		if !e.Payload_RPC.Has(dero.RPC_DESTINATION_PORT, dero.DataUint64) {
			if i > l-10 {
				serviceDebug(print, "[processBetTx]", fmt.Sprintf(PrintColor.Red+"%s No DST Port"+PrintColor.Reset, e.TXID))
			}
			continue
		}

		if Service.Dest_port != e.Payload_RPC.Value(dero.RPC_DESTINATION_PORT, dero.DataUint64).(uint64) {
			if i > l-10 {
				serviceDebug(print, "[processBetTx]", fmt.Sprintf(PrintColor.Red+"%s Bad DST port"+PrintColor.Reset, e.TXID))
			}
			continue
		}

		if e.Payload_RPC.Has(dero.RPC_COMMENT, dero.DataString) && e.Payload_RPC.Has(dero.RPC_REPLYBACK_ADDRESS, dero.DataAddress) {
			serviceDebug(print, "[processBetTx]", fmt.Sprintf("Processing %s", e.TXID))
			destination_expected := e.Payload_RPC.Value(dero.RPC_REPLYBACK_ADDRESS, dero.DataAddress).(dero.Address).String()
			addr, err := dero.NewAddress(destination_expected)
			if err != nil {
				serviceDebug(print, "[processBetTx]", err.Error())
				storeTx("BET", "done", db, e)
				continue
			}

			// addr.Mainnet = false
			destination_expected = addr.String()
			payload := e.Payload_RPC.Value(dero.RPC_COMMENT, dero.DataString).(string)
			split := strings.Split(payload, "  ")
			if len(split) > 4 {
				serviceDebug(print, "[processBetTx]", fmt.Sprintf("Payload %s", payload))
				serviceDebug(print, "[processBetTx]", fmt.Sprintf("Reply address %s", destination_expected))

				var scid string
				contracts := append(p_contracts, s_contracts...)
				found := false
				for _, sc := range contracts {
					check := sc[:6] + "..." + sc[58:]
					if check == split[len(split)-2] {
						serviceDebug(print, "[processBetTx]", fmt.Sprintf("Found scid %s", sc))
						found = true
						scid = sc
						break
					}
				}

				if found {
					var game_num string
					full_prefix := split[0]
					prefix := strings.Trim(full_prefix, "1234567890")
					if prefix != "p" && prefix != "s" {
						prefix = "nil"
					} else if prefix == "s" {
						game_num = strings.Trim(full_prefix, "s")
						if rpc.StringToInt(game_num) < 1 {
							serviceDebug(print, "[processBetTx]", fmt.Sprintf("%s No game number", e.TXID))
							sendRefund(scid, destination_expected, "No game number", e)
							storeTx("BET", "done", db, e)
							continue
						}
					}

					var amt []uint64
					switch prefix {
					case "p":
						_, amt = menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_amount", menu.Gnomes.Indexer.ChainHeight, true)
					case "s":
						_, amt = menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_amount_"+game_num, menu.Gnomes.Indexer.ChainHeight, true)
					default:
						serviceDebug(print, "[processBetTx]", fmt.Sprintf("%s No prefix", e.TXID))
						sendRefund(scid, destination_expected, "No prefix", e)
						storeTx("BET", "done", db, e)
						continue
					}

					if amt == nil || amt[0] == 0 {
						serviceDebug(print, "[processBetTx]", fmt.Sprintf("%s  Amount is nil", e.TXID))
						sendRefund(scid, destination_expected, "Void", e)
						storeTx("BET", "done", db, e)
						continue
					}

					value_expected := amt[0]
					if e.Amount != value_expected {
						serviceDebug(print, "[processBetTx]", fmt.Sprintf("User transferred %d, we were expecting %d. so we will refund", e.Amount, value_expected)) // this is an unexpected situation
						sendRefund(scid, destination_expected, "Wrong Amount", e)
						storeTx("BET", "done", db, e)
						continue
					}

					var sent bool
					for _, arg := range all_args {
						if arg.Value(dero.RPC_COMMENT, dero.DataString).(string) == payload {
							serviceDebug(print, "[processBetTx]", "Hit payload")

							switch prefix {
							case "p":
								serviceDebug(print, "[processBetTx]", "Payload is prediction")
								switch split[3] {
								case "Higher":
									serviceDebug(print, "[processBetTx]", "Higher arg")
									sent = sendToPrediction(1, scid, destination_expected, e)

								case "Lower":
									serviceDebug(print, "[processBetTx]", "Lower arg")
									sent = sendToPrediction(0, scid, destination_expected, e)

								default:
									sent = true
									serviceDebug(print, "[processBetTx]", fmt.Sprintf("%s No prediction", e.TXID))
									sendRefund(scid, destination_expected, "No prediction", e)
								}

							case "s":
								serviceDebug(print, "[processBetTx]", "Payload is sports")
								var team string
								team_a := menu.TrimTeamA(split[2])
								team_b := menu.TrimTeamB(split[2])
								if split[3] == team_a {
									team = "a"
								} else if split[3] == team_b {
									team = "b"
								} else {
									serviceDebug(print, "[processBetTx]", "Could not get team from payload")
								}

								switch team {
								case "a":
									serviceDebug(print, "[processBetTx]", "Team A arg")
									sent = sendToSports(game_num, team_a, "team_a", scid, destination_expected, e)
								case "b":
									serviceDebug(print, "[processBetTx]", "Team B arg")
									sent = sendToSports(game_num, team_b, "team_b", scid, destination_expected, e)
								default:
									sent = true
									serviceDebug(print, "[processBetTx]", fmt.Sprintf("%s No team", e.TXID))
									sendRefund(scid, destination_expected, "No team", e)

								}

							default:
								sent = true
								serviceDebug(print, "[processBetTx]", fmt.Sprintf("%s No prefix", e.TXID))
								sendRefund(scid, destination_expected, "No prefix", e)

							}

							if sent {
								break
							}
						}
					}

					if !sent {
						serviceDebug(print, "[processBetTx]", fmt.Sprintf("%s Could not make out payload", e.TXID))
						sendRefund(scid, destination_expected, "Bad payload", e)
					}
				} else {
					serviceDebug(print, "[processBetTx]", fmt.Sprintf("%s Scid not found", e.TXID))
				}
			} else {
				serviceDebug(print, "[processBetTx]", fmt.Sprintf("%s Payload format wrong", e.TXID))
			}
		} else {
			serviceDebug(print, "[processBetTx]", fmt.Sprintf("%s No comment or reply address", e.TXID))
		}
		storeTx("BET", "done", db, e)
	}
	serviceDebug(print, "[processBetTx]", "Done\n")
}

func processSingleTx(txid string) {
	if db := boltDB(); db != nil {
		defer db.Close()

		err := db.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte("BET"))
			return err
		})

		if err != nil {
			log.Printf("[dReamService] err creating bucket. err %s\n", err)
			return
		}

		rpcClient, _, _ := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)

		var p_contracts, s_contracts []string
		for _, sc := range menu.Control.Predict_owned {
			split := strings.Split(sc, "   ")
			if len(split) > 2 {
				p_contracts = append(p_contracts, split[2])
			}
		}

		for _, sc := range menu.Control.Sports_owned {
			split := strings.Split(sc, "   ")
			if len(split) > 2 {
				s_contracts = append(s_contracts, split[2])
			}
		}

		var all_args []dero.Arguments
		for _, sc := range p_contracts {
			higher, lower := intgPredictionArgs(sc, Service.Debug)
			if higher != nil && lower != nil {
				all_args = append(all_args, higher, lower)
			}
		}

		for _, sc := range s_contracts {
			sports := intgSportsArgs(sc, Service.Debug)
			for _, arg := range sports {
				all_args = append(all_args, arg...)
			}
		}

		params := dero.Get_Transfer_By_TXID_Params{
			TXID: txid,
		}

		var transfers dero.Get_Transfer_By_TXID_Result
		err = rpcClient.CallFor(context.TODO(), &transfers, "GetTransferbyTXID", params)
		if err != nil {
			log.Println("[processSingleTx]", err)
			return
		}

		log.Println("[processSingleTx] Processing", txid)

		e := transfers.Entry

		if e.Coinbase || !e.Incoming {
			log.Println("[processSingleTx]", e.TXID, "coinbase or outgoing")
			return
		}

		var already_processed bool
		db.View(func(tx *bbolt.Tx) error {
			if b := tx.Bucket([]byte("BET")); b != nil {
				if ok := b.Get([]byte(e.TXID)); ok != nil {
					already_processed = true
				}
			}
			return nil
		})

		if already_processed {
			log.Println("[processSingleTx]", fmt.Sprintf(PrintColor.Green+"%s Received: %d Already processed"+PrintColor.Reset, e.TXID, e.Height))
			return
		}

		if !e.Payload_RPC.Has(dero.RPC_DESTINATION_PORT, dero.DataUint64) {
			log.Println("[processSingleTx]", fmt.Sprintf(PrintColor.Red+"%s No DST Port"+PrintColor.Reset, e.TXID))
			return
		}

		if Service.Dest_port != e.Payload_RPC.Value(dero.RPC_DESTINATION_PORT, dero.DataUint64).(uint64) {
			log.Println("[processSingleTx]", fmt.Sprintf(PrintColor.Red+"%s Bad DST port"+PrintColor.Reset, e.TXID))
			return
		}

		if e.Payload_RPC.Has(dero.RPC_COMMENT, dero.DataString) && e.Payload_RPC.Has(dero.RPC_REPLYBACK_ADDRESS, dero.DataAddress) {
			destination_expected := e.Payload_RPC.Value(dero.RPC_REPLYBACK_ADDRESS, dero.DataAddress).(dero.Address).String()
			addr, err := dero.NewAddress(destination_expected)
			if err != nil {
				log.Println("[processSingleTx] err while while parsing incoming addr", err)
				storeTx("BET", "done", db, e)
				return
			}

			// addr.Mainnet = false
			destination_expected = addr.String()
			payload := e.Payload_RPC.Value(dero.RPC_COMMENT, dero.DataString).(string)
			split := strings.Split(payload, "  ")
			if len(split) > 4 {
				log.Println("[processSingleTx] Payload", payload)
				log.Println("[processSingleTx] Reply addr", destination_expected)

				var scid string
				contracts := append(p_contracts, s_contracts...)
				found := false
				for _, sc := range contracts {
					check := sc[:6] + "..." + sc[58:]
					if check == split[len(split)-2] {
						log.Println("[processSingleTx] Found Scid", sc)
						found = true
						scid = sc
						break
					}
				}

				if found {
					var game_num string
					full_prefix := split[0]
					prefix := strings.Trim(full_prefix, "1234567890")
					if prefix != "p" && prefix != "s" {
						prefix = "nil"
					} else if prefix == "s" {
						game_num = strings.Trim(full_prefix, "s")
						if rpc.StringToInt(game_num) < 1 {
							log.Println("[processSingleTx]", e.TXID, "No game number")
							rpc.ServiceRefund(e.Amount, e.SourcePort, scid, destination_expected, "No game number", e.TXID)
							storeTx("BET", "done", db, e)
							return
						}
					}

					var amt []uint64
					switch prefix {
					case "p":
						_, amt = menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_amount", menu.Gnomes.Indexer.ChainHeight, true)
					case "s":
						_, amt = menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_amount_"+game_num, menu.Gnomes.Indexer.ChainHeight, true)
					default:
						log.Println("[processSingleTx]", e.TXID, "No prefix")
						rpc.ServiceRefund(e.Amount, e.SourcePort, scid, destination_expected, "No prefix", e.TXID)
						storeTx("BET", "done", db, e)
						return
					}

					if amt == nil || amt[0] == 0 {
						log.Println("[processSingleTx]", e.TXID, "amount is nil")
						rpc.ServiceRefund(e.Amount, e.SourcePort, scid, destination_expected, "Void", e.TXID)
						storeTx("BET", "done", db, e)
						return
					}

					value_expected := amt[0]
					if e.Amount != value_expected {
						log.Println(nil, fmt.Sprintf("[processSingleTx] user transferred %d, we were expecting %d. so we will refund", e.Amount, value_expected)) // this is an unexpected situation
						rpc.ServiceRefund(e.Amount, e.SourcePort, scid, destination_expected, "Wrong Amount", e.TXID)
						storeTx("BET", "done", db, e)
						return
					}

					for _, arg := range all_args {
						if arg.Value(dero.RPC_COMMENT, dero.DataString).(string) == payload {
							log.Println("[processSingleTx] Hit payload")

							var sent bool
							switch prefix {
							case "p":
								log.Println("[processSingleTx] Payload is prediction")
								switch split[3] {
								case "Higher":
									log.Println("[processSingleTx] Higher arg")
									sent = sendToPrediction(1, scid, destination_expected, e)

								case "Lower":
									log.Println("[processSingleTx] Lower arg")
									sent = sendToPrediction(0, scid, destination_expected, e)

								default:
									sent = true
									log.Println("[processSingleTx]", e.TXID, "No prediction")
									rpc.ServiceRefund(e.Amount, e.SourcePort, scid, destination_expected, "No prediction", e.TXID)
								}

							case "s":
								log.Println("[processSingleTx] Payload is sports")
								var team string
								team_a := menu.TrimTeamA(split[2])
								team_b := menu.TrimTeamB(split[2])
								if split[3] == team_a {
									team = "a"
								} else if split[3] == team_b {
									team = "b"
								} else {
									log.Println("[processSingleTx] Could not get team from payload")
								}

								switch team {
								case "a":
									log.Println("[processSingleTx] Team A arg")
									sent = sendToSports(game_num, team_a, "team_a", scid, destination_expected, e)
								case "b":
									log.Println("[processSingleTx] Team B arg")
									sent = sendToSports(game_num, team_b, "team_b", scid, destination_expected, e)
								default:
									sent = true
									log.Println("[processSingleTx]", e.TXID, "No team")
									rpc.ServiceRefund(e.Amount, e.SourcePort, scid, destination_expected, "No team", e.TXID)

								}

							default:
								sent = true
								log.Println("[processSingleTx]", e.TXID, "No prefix")
								rpc.ServiceRefund(e.Amount, e.SourcePort, scid, destination_expected, "No prefix", e.TXID)

							}

							if sent {
								break
							}
						} else {
							log.Println("[processSingleTx]", e.TXID, "comment != payload")
						}
					}
				} else {
					log.Println("[processSingleTx]", e.TXID, "scid not found")
				}
			} else {
				log.Println("[processSingleTx]", e.TXID, "Payload format wrong")
			}
		} else {
			log.Println("[processSingleTx]", e.TXID, "No comment or reply address")
		}
		storeTx("BET", "done", db, e)

		log.Printf("[processSingleTx] Done\n\n")
	}
}

func viewProcessedTx(start uint64) {
	if db := boltDB(); db != nil {
		defer db.Close()

		err := db.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte("BET"))
			return err
		})

		if err != nil {
			log.Printf("[dReamService] err creating bucket. err %s\n", err)
			return
		}

		rpcClient, _, _ := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)

		out_params := dero.Get_Transfers_Params{
			Coinbase:   false,
			In:         false,
			Out:        true,
			Min_Height: PAYLOAD_FORMAT,
		}

		var outgoing dero.Get_Transfers_Result
		err = rpcClient.CallFor(context.TODO(), &outgoing, "GetTransfers", out_params)
		if err != nil {
			log.Println("[viewProcessedTx]", err)
			return
		}

		reply_id := checkReplies(outgoing)

		in_params := dero.Get_Transfers_Params{
			Coinbase:        false,
			In:              true,
			Out:             false,
			Min_Height:      start,
			DestinationPort: Service.Dest_port,
		}

		var transfers dero.Get_Transfers_Result
		err = rpcClient.CallFor(context.TODO(), &transfers, "GetTransfers", in_params)
		if err != nil {
			log.Println("[ViewProcessedTx] Could not obtain gettransfers from wallet", err)
			return
		}

		log.Println("[ViewProcessedTx] Viewing", len(transfers.Entries), "Entries from Height", strconv.Itoa(int(start)))

		for _, e := range transfers.Entries {
			if e.Coinbase || !e.Incoming {
				log.Println("[ViewProcessedTx]", e.TXID, "coinbase or outgoing")
				continue
			}

			var already_processed bool
			db.View(func(tx *bbolt.Tx) error {
				if b := tx.Bucket([]byte("BET")); b != nil {
					if ok := b.Get([]byte(e.TXID)); ok != nil {
						already_processed = true
					}
				}
				return nil
			})

			var replied bool
			var reply_txid string
			for id, repTx := range reply_id {
				if id == e.TXID[:6]+"..."+e.TXID[58:] {
					replied = true
					reply_txid = repTx
				}
			}

			when := e.Height
			if already_processed {
				log.Println("[ViewProcessedTx]", fmt.Sprintf(PrintColor.Green+"%s Received: %d Already processed"+PrintColor.Reset, e.TXID, when))
				if replied {
					log.Println("[ViewProcessedTx]", fmt.Sprintf(PrintColor.Yellow+"Replied: %s"+PrintColor.Reset, reply_txid))
				}
			} else {
				log.Println("[ViewProcessedTx]", fmt.Sprintf(PrintColor.Red+"%s Received: %d Not processed"+PrintColor.Reset, e.TXID, when))
			}
		}
		log.Println("[ViewProcessedTx] Done")
	}
}

func boltDB() *bbolt.DB {
	db_name := fmt.Sprintf("config/dReamService_%s.bbolt.db", rpc.Wallet.Address)
	db, err := bbolt.Open(db_name, 0600, nil)
	if err != nil {
		log.Printf("[dReamService] could not open db err:%s\n", err)
		return nil
	}

	return db
}

func storeTx(bucket, value string, db *bbolt.DB, e dero.Entry) {
	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		return b.Put([]byte(e.TXID), []byte(value))
	})

	if err != nil {
		log.Println("[storeTx]", bucket, err)
	} else {
		log.Println("[storeTx]", e.TXID, bucket, "Stored")
	}
}

func deleteTx(bucket string, db *bbolt.DB, e dero.Entry) {
	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		return b.Delete([]byte(e.TXID))
	})

	if err != nil {
		log.Println("[deleteTx]", bucket, err)
	} else {
		log.Println("[deleteTx]", e.TXID, bucket, "Deleted")
	}
}

func sendToPrediction(pre int, scid, destination_expected string, e dero.Entry) bool {
	waitForBlock()
	_, end := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_end_at", menu.Gnomes.Indexer.ChainHeight, true)
	_, buffer := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "buffer", menu.Gnomes.Indexer.ChainHeight, true)
	_, limit := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "limit", menu.Gnomes.Indexer.ChainHeight, true)
	_, played := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "p_#", menu.Gnomes.Indexer.ChainHeight, true)
	if end == nil || buffer == nil || limit == nil || played == nil {
		return false
	}

	now := time.Now().Unix()
	if now > int64(end[0]) {
		rpc.ServiceRefund(e.Amount, e.SourcePort, scid, destination_expected, "Past Deadline", e.TXID)
	} else if now < int64(buffer[0]) {
		rpc.ServiceRefund(e.Amount, e.SourcePort, scid, destination_expected, "Before Buffer", e.TXID)
	} else if played[0] >= limit[0] {
		rpc.ServiceRefund(e.Amount, e.SourcePort, scid, destination_expected, "Bet Limit Reached", e.TXID)
	} else {
		rpc.AutoPredict(pre, e.Amount, e.SourcePort, scid, destination_expected, e.TXID)
	}

	Service.Last_block = rpc.Wallet.Height

	t := 0
	if Service.Debug {
		log.Println("[sendToPrediction] Tx delay")
	}

	for rpc.Wallet.Height < Service.Last_block+3 {
		if !rpc.Wallet.Service || !rpc.Wallet.Connect || !rpc.Daemon.Connect || t > 36 {
			break
		}

		t++
		time.Sleep(1 * time.Second)
	}

	return true
}

func sendToSports(n, abv, team, scid, destination_expected string, e dero.Entry) bool {
	waitForBlock()
	_, end := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_end_at_"+n, menu.Gnomes.Indexer.ChainHeight, true)
	_, buffer := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "buffer"+n, menu.Gnomes.Indexer.ChainHeight, true)
	_, limit := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "limit", menu.Gnomes.Indexer.ChainHeight, true)
	_, played := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_#_"+n, menu.Gnomes.Indexer.ChainHeight, true)
	if end == nil || buffer == nil || limit == nil || played == nil {
		return false
	}

	var pre uint64
	if team == "team_a" {
		pre = 0
	} else if team == "team_b" {
		pre = 1
	}

	now := time.Now().Unix()
	if now > int64(end[0]) {
		rpc.ServiceRefund(e.Amount, e.SourcePort, scid, destination_expected, "Past Deadline", e.TXID)
	} else if now < int64(buffer[0]) {
		rpc.ServiceRefund(e.Amount, e.SourcePort, scid, destination_expected, "Before Buffer", e.TXID)
	} else if played[0] >= limit[0] {
		rpc.ServiceRefund(e.Amount, e.SourcePort, scid, destination_expected, "Bet Limit Reached", e.TXID)
	} else {
		rpc.AutoBook(e.Amount, pre, e.SourcePort, n, abv, scid, destination_expected, e.TXID)
	}

	Service.Last_block = rpc.Wallet.Height

	t := 0
	if Service.Debug {
		log.Println("[sendToSports] Tx delay")
	}

	for rpc.Wallet.Height < Service.Last_block+3 {
		if !rpc.Wallet.Service || !rpc.Wallet.Connect || !rpc.Daemon.Connect || t > 36 {
			break
		}

		t++
		time.Sleep(1 * time.Second)
	}

	return true
}

func sendRefund(scid, addr, msg string, e dero.Entry) {
	waitForBlock()
	rpc.ServiceRefund(e.Amount, e.SourcePort, scid, addr, msg, e.TXID)
	Service.Last_block = rpc.Wallet.Height

	t := 0
	if Service.Debug {
		log.Println("[sendRefund] Tx delay")
	}
	for rpc.Wallet.Height < Service.Last_block+3 {
		if !rpc.Wallet.Service || !rpc.Wallet.Connect || !rpc.Daemon.Connect || t > 36 {
			break
		}

		t++
		time.Sleep(1 * time.Second)
	}
}

func waitForBlock() {
	i := 0
	if Service.Debug && rpc.Wallet.Height < Service.Last_block+3 {
		log.Println("[waitForBlock] Waiting for block")
	}

	for rpc.Wallet.Height < Service.Last_block+3 && i < 20 {
		i++
		time.Sleep(3 * time.Second)
	}
}

func checkReplies(outgoing dero.Get_Transfers_Result) (reply_id map[string]string) {
	reply_id = make(map[string]string)
	for _, out := range outgoing.Entries {
		if out.Payload_RPC.Has(dero.RPC_COMMENT, dero.DataString) {
			comm := out.Payload_RPC.Value(dero.RPC_COMMENT, dero.DataString).(string)
			split := strings.Split(comm, ",  ")
			if len(split) == 2 {
				if len(split[1]) == 15 {
					reply_id[split[1]] = out.TXID
				}
			}
		}
	}

	return
}
