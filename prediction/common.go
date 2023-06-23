package prediction

import (
	"log"
	"strconv"
	"time"

	dreams "github.com/SixofClubsss/dReams"
	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"
)

func OnConnected() {
	Predict.Settings.Contract_entry.CursorColumn = 1
	Predict.Settings.Contract_entry.Refresh()
	Sports.Settings.Contract_entry.CursorColumn = 1
	Sports.Settings.Contract_entry.Refresh()
}

// Main process for dSports and dPrediction
func fetch(p, s *dreams.DreamsItems, d dreams.DreamsObject) {
	var offset int
	SetPrintColors(d.OS)
	time.Sleep(3 * time.Second)
	for {
		select {
		case <-d.Receive():
			if !rpc.Wallet.IsConnected() || !rpc.Daemon.IsConnected() {
				disableActions()
				Owner.Synced = false
				s.RightLabel.SetText("dReams Balance: " + rpc.DisplayBalance("dReams") + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Wallet.Display.Height)
				p.RightLabel.SetText("dReams Balance: " + rpc.DisplayBalance("dReams") + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Wallet.Display.Height)
				d.WorkDone()
				continue
			}

			if !Owner.Synced && menu.GnomonScan(d.Configure) {
				log.Println("[dPrediction] Syncing")
				contracts := menu.Gnomes.IndexContains()
				go CheckBetContractOwners(contracts)
				if !d.IsWindows() {
					go PopulateSports(contracts)
					go PopulatePredictions(contracts)
				}
				Owner.Synced = true
			}

			// dSports
			if d.OnTab("Sports") {
				if offset%5 == 0 {
					go SetSportsInfo(Sports.Contract)
				}
			}
			s.RightLabel.SetText("dReams Balance: " + rpc.DisplayBalance("dReams") + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Wallet.Display.Height)

			//dPrediction
			if d.OnTab("Predict") {
				if offset%5 == 0 {
					go SetPredictionInfo(Predict.Contract)
				}

				if offset == 11 || Predict.Prices.Text == "" {
					go SetPredictionPrices(rpc.Daemon.Connect)
				}

				p.RightLabel.SetText("dReams Balance: " + rpc.DisplayBalance("dReams") + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Wallet.Display.Height)

				if CheckActivePrediction(Predict.Contract) {
					go ShowPredictionControls()
				} else {
					disablePredictions(true)
				}
			}

			offset++
			if offset >= 21 {
				offset = 0
			}

			d.WorkDone()
		case <-d.CloseDapp():
			log.Println("[dPrediction] Done")
			return
		}
	}
}

func Disconnected() {
	Service.Stop()
	Predict.owner = false
	Owner.Synced = true
}

func disableActions() {
	Predict.Settings.New.Hide()
	Sports.Settings.New.Hide()
	Predict.Settings.Unlock.Hide()
	Sports.Settings.Unlock.Hide()
	Predict.Settings.Menu.Hide()
	Sports.Settings.Menu.Hide()

	Predict.Settings.New.Refresh()
	Sports.Settings.New.Refresh()
	Predict.Settings.Unlock.Refresh()
	Sports.Settings.Unlock.Refresh()
	Predict.Settings.Menu.Refresh()
	Sports.Settings.Menu.Refresh()

	Predict.Settings.Contracts = []string{}
	Sports.Settings.Contracts = []string{}
	Predict.Settings.Owned = []string{}
	Sports.Settings.Owned = []string{}

	Predict.Settings.Check.SetChecked(false)
	disablePredictions(true)
	disableSports(true)
}

// Set objects if bet owner
func setBetOwner(owner string) {
	if owner == rpc.Wallet.Address {
		Predict.owner = true
		Predict.Settings.New.Show()
		Sports.Settings.New.Show()
		Predict.Settings.Unlock.Hide()
		Sports.Settings.Unlock.Hide()
		Predict.Settings.Menu.Show()
		Sports.Settings.Menu.Show()
	} else {
		Predict.owner = false
		Predict.Settings.New.Hide()
		Sports.Settings.New.Hide()
		Predict.Settings.Unlock.Show()
		Sports.Settings.Unlock.Show()
		Predict.Settings.Menu.Hide()
		Sports.Settings.Menu.Hide()
	}
}

// Scan all bet contracts to verify if owner
//   - Pass contracts from db store, can be nil arg
func CheckBetContractOwners(contracts map[string]string) {
	if menu.Gnomes.IsReady() {
		if contracts == nil {
			contracts = menu.Gnomes.GetAllOwnersAndSCIDs()
		}
		keys := make([]string, len(contracts))

		i := 0
		for k := range contracts {
			keys[i] = k
			verifyBetContractOwner(keys[i], "p")
			verifyBetContractOwner(keys[i], "s")
			if Predict.owner {
				break
			}
			i++
		}
	}
}

// Verify if wallet is owner on bet contract
//   - Passed t defines sports or prediction contract
func verifyBetContractOwner(scid, t string) {
	if menu.Gnomes.IsReady() {
		if dev, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "dev"); dev != nil {
			owner, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "owner")
			_, init := menu.Gnomes.GetSCIDValuesByKey(scid, t+"_init")

			if owner != nil && init != nil {
				if dev[0] == rpc.DevAddress && !Predict.owner {
					setBetOwner(owner[0])
				}
			}
		}
	}
}

// Verify if wallet is a co owner on bet contract
func VerifyBetSigner(scid string) bool {
	if menu.Gnomes.IsReady() {
		for i := 2; i < 10; i++ {
			if !menu.Gnomes.IsRunning() {
				break
			}

			signer_addr, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "co_signer"+strconv.Itoa(i))
			if signer_addr != nil {
				if signer_addr[0] == rpc.Wallet.Address {
					return true
				}
			}
		}
	}

	return false
}

// Get info for bet contract by SCID
//   - Passed t defines sports or prediction contract
//   - Adding constructed header string to list, owned []string
func checkBetContract(scid, t string, list, owned []string) ([]string, []string) {
	if menu.Gnomes.IsReady() {
		if dev, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "dev"); dev != nil {
			owner, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "owner")
			_, init := menu.Gnomes.GetSCIDValuesByKey(scid, t+"_init")

			if owner != nil && init != nil {
				if dev[0] == rpc.DevAddress {
					headers := menu.GetSCHeaders(scid)
					name := "?"
					desc := "?"
					var hidden bool
					_, restrict := menu.Gnomes.GetSCIDValuesByKey(rpc.RatingSCID, "restrict")
					_, rating := menu.Gnomes.GetSCIDValuesByKey(rpc.RatingSCID, scid)

					if restrict != nil && rating != nil {
						menu.Control.Lock()
						menu.Control.Contract_rating[scid] = rating[0]
						menu.Control.Unlock()
						if rating[0] <= restrict[0] {
							hidden = true
						}
					}

					if headers != nil {
						if headers[1] != "" {
							desc = headers[1]
						}

						if headers[0] != "" {
							name = " " + headers[0]
						}

						if headers[0] == "-" {
							hidden = true
						}
					}

					var co_signer bool
					if VerifyBetSigner(scid) {
						co_signer = true
						if !menu.Gnomes.Import {
							Predict.Settings.Menu.Show()
							Sports.Settings.Menu.Show()
						}
					}

					if owner[0] == rpc.Wallet.Address || co_signer {
						owned = append(owned, name+"   "+desc+"   "+scid)
					}

					if !hidden {
						list = append(list, name+"   "+desc+"   "+scid)
					}
				}
			}
		}
	}

	return list, owned
}
