package derbnb

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/holdero"
	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/deroproject/derohe/walletapi"
)

const (
	TIME_FORMAT    = "Mon 02 Jan 2006"
	TOKEN_CONTRACT = `// derbnb property contract v1

Function InitializePrivate() Uint64
	10 IF EXISTS("metadata") THEN GOTO 100
	20 SEND_ASSET_TO_ADDRESS(SIGNER(), 1, SCID())
	30 STORE("metadata","")
	40 STORE("changed", 0)
	50 RETURN 0
	100 RETURN 1
End Function

Function StoreLocation(location String) Uint64
	10 IF ASSETVALUE(SCID()) != 1 THEN GOTO 100
	20 IF location == "" THEN GOTO 100
	30 STORE("location_"+ITOA(LOAD("changed")), location)
	40 IF LOAD("changed") < 6 THEN GOTO 60
	50 DELETE("location_"+ITOA(LOAD("changed")-5))
	60 STORE("changed", LOAD("changed")+1)
	70 SEND_ASSET_TO_ADDRESS(SIGNER(), 1, SCID())
	80 RETURN 0 
	100 RETURN 1
End Function

Function UpdateMetadata(metadata String) Uint64
	10 IF ASSETVALUE(SCID()) != 1 THEN GOTO 100
	20 STORE("metadata", metadata)
	30 SEND_ASSET_TO_ADDRESS(SIGNER(), 1, SCID())
	40 RETURN 0
	100 RETURN 1
End Function`
)

type location_data struct {
	City    string `json:"city"`
	Country string `json:"country"`
}

var current_price uint64
var current_deposit uint64
var listed_properties []string
var property_photos map[string][]string
var my_bookings map[string][]string
var my_properties map[string][]string
var background *fyne.Container

// Run DerBnb as a single dApp
func StartApp() {
	config := menu.ReadDreamsConfig("DerBnb")
	a := app.New()
	a.Settings().SetTheme(bundle.DeroTheme(config.Skin))
	w := a.NewWindow("DerBnb Desktop")
	w.SetIcon(bundle.ResourceDerbnbIconPng)
	w.Resize(fyne.NewSize(1200, 800))
	w.SetMaster()
	quit := make(chan struct{})
	w.SetCloseIntercept(func() {
		menu.WriteDreamsConfig(rpc.Daemon.Rpc, config.Skin)
		quit <- struct{}{}
		w.Close()
	})

	menu.Gnomes.Fast = true

	holdero.Settings.ThemeImg = *canvas.NewImageFromResource(nil)
	background = container.NewMax(&holdero.Settings.ThemeImg)

	go fetch(quit)
	w.SetContent(container.New(layout.NewMaxLayout(), background, LayoutAllItems(false, w, background)))
	w.ShowAndRun()
}

// Main DerBnb process used in StartApp()
func fetch(quit chan struct{}) {
	time.Sleep(6 * time.Second)
	ticker := time.NewTicker(3 * time.Second)

	for {
		select {
		case <-ticker.C: // do on interval
			rpc.Ping()
			rpc.EchoWallet("DerBnb")
			GetProperties()
			rpc.GetBalance()
			if !rpc.Wallet.Connect {
				rpc.Wallet.Balance = 0
			}

			connect_box.RefreshBalance()
			if !rpc.Signal.Startup {
				menu.GnomonEndPoint()
			}

			if rpc.Daemon.Connect && menu.Gnomes.Init {
				connect_box.Disconnect.SetChecked(true)
				height := rpc.DaemonHeight("DerBnb", rpc.Daemon.Rpc)
				if menu.Gnomes.Indexer.LastIndexedHeight >= int64(height)-3 {
					menu.Gnomes.Sync = true
				} else {
					menu.Gnomes.Sync = false
				}
			} else {
				connect_box.Disconnect.SetChecked(false)
			}

			if rpc.Daemon.Connect {
				rpc.Signal.Startup = false
			}

		case <-quit: // exit
			log.Println("[DerBNB] Closing")
			ticker.Stop()
			return
		}
	}
}

// Check if property entry exists already in my_properties
func haveProperty(text string) (have bool) {
	for _, prop := range my_properties[""] {
		if text == prop {
			have = true
		}
	}

	return
}

// Define https or ipfs urls
func propertyImageSource(check string) string {
	if check[0:8] == "https://" {
		return check
	}

	return "https://ipfs.io/ipfs/" + check
}

// Make a formated City, Country location string from SCID
func makeLocationString(scid string) (location string) {
	city, country := getLocation(scid)
	if city != "" && country != "" {
		location = fmt.Sprintf("%s, %s", city, country)
	} else if country != "" {
		location = country
	}

	return
}

// Get all DerBnb property info from contract
func GetProperties() {
	if rpc.Wallet.Connect && rpc.Daemon.Connect {
		if menu.Gnomes.Init && menu.Gnomes.Sync && !menu.GnomonClosing() {
			info := menu.Gnomes.Indexer.Backend.GetAllSCIDVariableDetails(rpc.DerBnbSCID)
			if info != nil {
				i := 0
				keys := make([]int, len(info))
				for k := range info {
					keys[i] = int(k)
					i++
				}

				// If no info, return
				if len(keys) == 0 {
					return
				}

				sort.Ints(keys)
				listed_properties = []string{}
				my_properties[""] = []string{}
				my_bookings[""] = []string{"Completed", "Confirmed", "Requests"}

				for sc := range my_properties {
					my_properties[sc] = []string{}
				}

				for _, curr := range my_bookings[""] {
					my_bookings[curr] = []string{}
				}

				var double bool
				added_bookings := make(map[string]bool)
				for _, h := range info[int64(keys[len(keys)-1])] {
					split := strings.Split(h.Key.(string), "_")
					l := len(split)
					if l > 1 && len(split[0]) == 64 {
						var have bool
						for _, p := range listed_properties {
							if split[0] == p {
								have = true
								break
							}
						}

						if split[1] == "owner" {
							if !have {
								location := makeLocationString(split[0])
								if location != "" {
									list_string := fmt.Sprintf("%s   %s", location, split[0])
									listed_properties = append(listed_properties, list_string)
								}
							}

							// find owned properties
							if h.Value.(string) == rpc.Wallet.Address {
								my_properties[""] = append(my_properties[""], split[0])
							}
						}

						// find userr confirmed booking requests
						if !double && split[1] == "booker" && h.Value.(string) == rpc.Wallet.Address {
							booked, complete := getUserConfirmedBookings(split[0], false)
							my_bookings["Confirmed"] = append(my_bookings["Confirmed"], booked...)
							my_bookings["Completed"] = append(my_bookings["Completed"], complete...)
							double = true
						}

						if l > 3 {
							// find request bookings of property
							if split[1] == "request" && split[2] == "bk" && split[3] == "end" {
								add := true
								for _, prop := range my_properties[split[0]] {
									check := strings.Split(prop, "   ")
									if check[0] == split[len(split)-1] {
										add = false
										break
									}
								}

								if add {
									requests := getBookingRequests(split[0], split[len(split)-1], true)
									if requests != "" {
										my_properties[split[0]] = append(my_properties[split[0]], requests)
									}
								}
							}

							// find user booking requests
							if split[2] == "booker" && h.Value.(string) == rpc.Wallet.Address {
								my_bookings["Requests"] = append(my_bookings["Requests"], getBookingRequests(split[0], split[len(split)-1], false))
							}

							// find owner confirmed booking requests
							if !added_bookings[split[0]] && split[1] == "bk" && split[2] == "end" {
								bookings := getOwnerConfirmedBookings(split[0], true)
								if bookings != nil {
									my_properties[split[0]] = append(my_properties[split[0]], bookings...)
									sort.Strings(my_properties[split[0]])
									added_bookings[split[0]] = true
								}
							}
						}
					}
				}
				listing_label.SetText(getInfo(viewing_scid))
				booking_list.Refresh()
				property_list.Refresh()
				listings_list.Refresh()
			}
		}
	} else {
		listed_properties = []string{}
		my_properties[""] = []string{}
		my_bookings[""] = []string{"Completed", "Confirmed", "Requests"}
		for _, curr := range my_bookings[""] {
			my_bookings[curr] = []string{}
		}
		booking_list.Refresh()
		property_list.Refresh()
		listings_list.Refresh()
	}
}

// Get request to owners and of the renters
//   - stamp is timestap key of request
func getBookingRequests(scid, stamp string, owned bool) (request string) {
	if menu.Gnomes.Init && menu.Gnomes.Sync && !menu.GnomonClosing() {
		_, start := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_request_bk_start_"+stamp, menu.Gnomes.Indexer.ChainHeight, true)
		_, end := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_request_bk_end_"+stamp, menu.Gnomes.Indexer.ChainHeight, true)
		booker, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_request_booker_"+stamp, menu.Gnomes.Indexer.ChainHeight, true)

		if start != nil && end != nil && booker != nil {
			s := time.Unix(int64(start[0]), 0)
			e := time.Unix(int64(end[0]), 0)
			final := booker[0]
			if !owned {
				final = scid
			}

			return fmt.Sprintf("Request:   %s   %s   %s   %s", stamp, s.Format(TIME_FORMAT), e.Format(TIME_FORMAT), final)
		}
	}

	return
}

// Get confirmed bookings for owners
//   - stamp is timestap key of request
func getOwnerConfirmedBookings(scid string, all bool) (bookings []string) {
	if menu.Gnomes.Init && menu.Gnomes.Sync && !menu.GnomonClosing() {
		_, bk_last := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_bk_last", menu.Gnomes.Indexer.ChainHeight, true)
		if bk_last != nil {
			i := int(bk_last[0])
			for {
				last := strconv.Itoa(i)
				_, start := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_bk_start_"+last, menu.Gnomes.Indexer.ChainHeight, true)
				_, end := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_bk_end_"+last, menu.Gnomes.Indexer.ChainHeight, true)
				booker, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_booker_"+last, menu.Gnomes.Indexer.ChainHeight, true)
				complete, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_"+last+"_damage_renter", menu.Gnomes.Indexer.ChainHeight, true)

				if start != nil && end != nil && booker != nil {
					prefix := "Booked:"
					if complete != nil {
						prefix = "Complete:"
					}

					s := time.Unix(int64(start[0]), 0)
					e := time.Unix(int64(end[0]), 0)

					if !all {
						if booker[0] == rpc.Wallet.Address {
							bookings = append(bookings, fmt.Sprintf("%s   %s   %s   %s   %s", prefix, last, s.Format(TIME_FORMAT), e.Format(TIME_FORMAT), scid))
						}
					} else {
						bookings = append(bookings, fmt.Sprintf("%s   %s   %s   %s   %s", prefix, last, s.Format(TIME_FORMAT), e.Format(TIME_FORMAT), booker[0]))
					}
				}

				i--
				if i <= 0 {
					break
				}
			}
		}
	}

	return
}

// Get property info from DerBnb contract for display
func getInfo(scid string) (info string) {
	if menu.Gnomes.Init && menu.Gnomes.Sync && !menu.GnomonClosing() {
		owner, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_owner", menu.Gnomes.Indexer.ChainHeight, true)
		_, price := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_price", menu.Gnomes.Indexer.ChainHeight, true)
		_, dep := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_damage_deposit", menu.Gnomes.Indexer.ChainHeight, true)
		//_, bk_last := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(CONTRACT, scid+"_bk_last", menu.Gnomes.Indexer.ChainHeight, true)

		if owner != nil && price != nil && dep != nil {
			location := makeLocationString(scid)
			data := getMetadata(scid)
			data_string := fmt.Sprintf("Sq feet: %d\n\nStyle: %s\n\nBedrooms: %d\n\nMax guests: %d", data.Squarefootage, data.Style, data.NumberOfBedrooms, data.MaxNumberOfGuests)
			current_price = price[0]
			current_deposit = dep[0]
			info = fmt.Sprintf("Location: %s\n\nPrice: %s Dero per night\n\nDamage Deposit: %s Dero\n\nOwner: %s\n\n%s", location, walletapi.FormatMoney(price[0]), walletapi.FormatMoney(dep[0]), owner[0], data_string)
		}
	}

	return
}

// Get owner address of DerBnb property
func getOwnerAddress(scid string) (address string) {
	if menu.Gnomes.Init && menu.Gnomes.Sync && !menu.GnomonClosing() {
		owner, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_owner", menu.Gnomes.Indexer.ChainHeight, true)
		if owner != nil {
			return owner[0]
		}
	}

	return
}

// Get confirmed  and completed bookings for renters
//   - stamp is timestap key of request
func getUserConfirmedBookings(scid string, all bool) (confirmed_bookings []string, complete_bookinds []string) {
	if menu.Gnomes.Init && menu.Gnomes.Sync && !menu.GnomonClosing() {
		_, bk_last := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_bk_last", menu.Gnomes.Indexer.ChainHeight, true)
		if bk_last != nil {
			i := int(bk_last[0])
			for {
				last := strconv.Itoa(i)
				_, start := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_bk_start_"+last, menu.Gnomes.Indexer.ChainHeight, true)
				_, end := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_bk_end_"+last, menu.Gnomes.Indexer.ChainHeight, true)
				booker, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_booker_"+last, menu.Gnomes.Indexer.ChainHeight, true)
				complete, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_"+last+"_damage_renter", menu.Gnomes.Indexer.ChainHeight, true)

				if start != nil && end != nil && booker != nil {
					prefix := "Booked:"
					if complete != nil {
						prefix = "Complete:"
					}

					s := time.Unix(int64(start[0]), 0)
					e := time.Unix(int64(end[0]), 0)

					if !all {
						if booker[0] == rpc.Wallet.Address {
							switch prefix {
							case "Booked:":
								confirmed_bookings = append(confirmed_bookings, fmt.Sprintf("%s   %s   %s   %s   %s", prefix, last, s.Format(TIME_FORMAT), e.Format(TIME_FORMAT), scid))
							case "Complete:":
								complete_bookinds = append(complete_bookinds, fmt.Sprintf("%s   %s   %s   %s   %s", prefix, last, s.Format(TIME_FORMAT), e.Format(TIME_FORMAT), scid))
							default:

							}
						}
					} else {
						switch prefix {
						case "Booked:":
							confirmed_bookings = append(confirmed_bookings, fmt.Sprintf("%s   %s   %s   %s   %s", prefix, last, s.Format(TIME_FORMAT), e.Format(TIME_FORMAT), booker[0]))
						case "Complete:":
							complete_bookinds = append(complete_bookinds, fmt.Sprintf("%s   %s   %s   %s   %s", prefix, last, s.Format(TIME_FORMAT), e.Format(TIME_FORMAT), booker[0]))
						default:

						}
					}
				}

				i--
				if i <= 0 {
					break
				}
			}
		}
	}

	return
}
