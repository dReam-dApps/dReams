package holdero

import (
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/canvas"
	dreams "github.com/SixofClubsss/dReams"
	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"
)

// Check if wallet owns Holdero table
func checkTableOwner(scid string) bool {
	if len(scid) != 64 || !menu.Gnomes.IsReady() {
		return false
	}

	check := strings.Trim(scid, " 0123456789")
	if check == "Holdero Tables:" {
		return false
	}

	owner, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "owner:")
	if owner != nil {
		return owner[0] == rpc.Wallet.Address
	}

	return false
}

// Check if Holdero table is a tournament table
func checkHolderoContract(scid string) bool {
	if len(scid) != 64 || !menu.Gnomes.IsReady() {
		return false
	}

	_, deck := menu.Gnomes.GetSCIDValuesByKey(scid, "Deck Count:")
	_, version := menu.Gnomes.GetSCIDValuesByKey(scid, "V:")
	_, tourney := menu.Gnomes.GetSCIDValuesByKey(scid, "Tournament")
	if deck != nil && version != nil && version[0] >= 100 {
		Signal.Contract = true
	}

	if tourney != nil && tourney[0] == 1 {
		return true
	}

	return false
}

// Check Holdero table version
func checkTableVersion(scid string) uint64 {
	_, v := menu.Gnomes.GetSCIDValuesByKey(scid, "V:")

	if v != nil && v[0] >= 100 {
		return v[0]
	}
	return 0
}

// Make list of public and owned tables
func createTableList() {
	if menu.Gnomes.IsReady() {
		var owner bool
		list := []string{}
		owned := []string{}
		tables := menu.Gnomes.GetAllOwnersAndSCIDs()
		for scid := range tables {
			if !menu.Gnomes.IsReady() {
				break
			}

			if _, valid := menu.Gnomes.GetSCIDValuesByKey(scid, "Deck Count:"); valid != nil {
				_, version := menu.Gnomes.GetSCIDValuesByKey(scid, "V:")

				if version != nil {
					d := valid[0]
					v := version[0]

					headers := menu.GetSCHeaders(scid)
					name := "?"
					desc := "?"
					if headers != nil {
						if headers[1] != "" {
							desc = headers[1]
						}

						if headers[0] != "" {
							name = " " + headers[0]
						}
					}

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

					if d >= 1 && v == 110 && !hidden {
						list = append(list, name+"   "+desc+"   "+scid)
					}

					if d >= 1 && v >= 100 {
						if checkTableOwner(scid) {
							owned = append(owned, name+"   "+desc+"   "+scid)
							Poker.Holdero_unlock.Hide()
							Poker.Holdero_new.Show()
							owner = true
							Poker.table_owner = true
						}
					}
				}
			}
		}

		if !owner {
			Poker.Holdero_unlock.Show()
			Poker.Holdero_new.Hide()
			Poker.table_owner = false
		}

		t := len(list)
		list = append(list, "  Holdero Tables: "+strconv.Itoa(t))
		sort.Strings(list)
		Settings.Tables = list

		sort.Strings(owned)
		Settings.Owned = owned

		Poker.Table_list.Refresh()
		Poker.Owned_list.Refresh()
	}
}

// Get current Holdero table menu stats
func getTableStats(scid string, single bool) {
	if menu.Gnomes.IsReady() && len(scid) == 64 {
		_, v := menu.Gnomes.GetSCIDValuesByKey(scid, "V:")
		_, l := menu.Gnomes.GetSCIDValuesByKey(scid, "Last")
		_, s := menu.Gnomes.GetSCIDValuesByKey(scid, "Seats at Table:")
		// _, o := menu.Gnomes.GetSCIDValuesByKey(scid, "Open")
		// p1, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "Player 1 ID:")
		p2, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "Player2 ID:")
		p3, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "Player3 ID:")
		p4, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "Player4 ID:")
		p5, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "Player5 ID:")
		p6, _ := menu.Gnomes.GetSCIDValuesByKey(scid, "Player6 ID:")
		h := menu.GetSCHeaders(scid)

		if single {
			if h != nil {
				Table.Stats.Name.Text = (" Name: " + h[0])
				Table.Stats.Name.Refresh()
				Table.Stats.Desc.Text = (" Description: " + h[1])
				Table.Stats.Desc.Refresh()
				if len(h[2]) > 6 {
					Table.Stats.Image, _ = dreams.DownloadFile(h[2], h[0])
				} else {
					Table.Stats.Image = *canvas.NewImageFromImage(nil)
				}

			} else {
				Table.Stats.Name.Text = (" Name: ?")
				Table.Stats.Name.Refresh()
				Table.Stats.Desc.Text = (" Description: ?")
				Table.Stats.Desc.Refresh()
				Table.Stats.Image = *canvas.NewImageFromImage(nil)
			}
		}

		if v != nil {
			Table.Stats.Version.Text = (" Table Version: " + strconv.Itoa(int(v[0])))
			Table.Stats.Version.Refresh()
		} else {
			Table.Stats.Version.Text = (" Table Version: ?")
			Table.Stats.Version.Refresh()
		}

		if l != nil {
			time, _ := rpc.MsToTime(strconv.Itoa(int(l[0]) * 1000))
			Table.Stats.Last.Text = (" Last Move: " + time.String())
			Table.Stats.Last.Refresh()
		} else {
			Table.Stats.Last.Text = (" Last Move: ?")
			Table.Stats.Last.Refresh()
		}

		if s != nil {
			if s[0] > 1 {
				sit := 1
				if p2 != nil {
					sit++
				}

				if p3 != nil {
					sit++
				}

				if p4 != nil {
					sit++
				}

				if p5 != nil {
					sit++
				}

				if p6 != nil {
					sit++
				}

				Table.Stats.Seats.Text = (" Seats at Table: " + strconv.Itoa(int(s[0])-sit))
				Table.Stats.Seats.Refresh()
			}
		} else {
			Table.Stats.Seats.Text = (" Table Closed")
			Table.Stats.Seats.Refresh()
		}
	}
}
