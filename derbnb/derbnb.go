package derbnb

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
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

type string_slice_map struct {
	data map[string][]string
	sync.RWMutex
}

var searching_properties bool
var current_price uint64
var current_deposit uint64
var property_filter []string
var listed_properties []string
var property_photos string_slice_map
var my_bookings string_slice_map
var my_properties string_slice_map
var background *fyne.Container

// Run DerBnb as a single dApp
func StartApp() {
	n := runtime.NumCPU()
	runtime.GOMAXPROCS(n)
	config := menu.ReadDreamsConfig("DerBnb")

	a := app.New()
	a.Settings().SetTheme(bundle.DeroTheme(config.Skin))
	w := a.NewWindow("DerBnb Desktop")
	w.SetIcon(bundle.ResourceDerbnbIconPng)
	w.Resize(fyne.NewSize(1200, 800))
	w.SetMaster()
	quit := make(chan struct{})
	done := make(chan struct{})
	w.SetCloseIntercept(func() {
		menu.WriteDreamsConfig(rpc.Daemon.Rpc, config.Skin)
		menu.Gnomes.Stop("DerBnb")
		quit <- struct{}{}
		w.Close()
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println()
		menu.WriteDreamsConfig(rpc.Daemon.Rpc, config.Skin)
		menu.Gnomes.Stop("DerBnb")
		quit <- struct{}{}
		w.Close()
	}()

	menu.Gnomes.Fast = true

	holdero.Settings.ThemeImg = *canvas.NewImageFromResource(nil)
	background = container.NewMax(&holdero.Settings.ThemeImg)

	go fetch(quit, done)
	go func() {
		time.Sleep(450 * time.Millisecond)
		w.SetContent(container.New(layout.NewMaxLayout(), background, LayoutAllItems(false, w, background)))
	}()
	w.ShowAndRun()
	<-done
}

// Main DerBnb process used in StartApp()
func fetch(quit, done chan struct{}) {
	log.Println("[DerBnb]", rpc.DREAMSv, runtime.GOOS, runtime.GOARCH)
	time.Sleep(6 * time.Second)
	ticker := time.NewTicker(3 * time.Second)
	rpc.Wallet.TokenBal = make(map[string]uint64)

	for {
		select {
		case <-ticker.C: // do on interval
			rpc.Ping()
			rpc.EchoWallet("DerBnb")
			GetProperties()
			rpc.Wallet.GetBalance()
			rpc.Wallet.GetTokenBalance("TRVL", rpc.TrvlSCID)

			connect_box.RefreshBalance()
			if !rpc.Signal.Startup {
				menu.GnomonEndPoint()
			}

			if rpc.Daemon.IsConnected() && menu.Gnomes.IsInitialized() {
				connect_box.Disconnect.SetChecked(true)
				if menu.Gnomes.IsRunning() {
					menu.Gnomes.IndexContains()
					if menu.Gnomes.HasIndex(1) {
						menu.Gnomes.Checked(true)
					}
				}

				if menu.Gnomes.Indexer.LastIndexedHeight >= menu.Gnomes.Indexer.ChainHeight-3 {
					menu.Gnomes.Synced(true)
				} else {
					menu.Gnomes.Synced(false)
					menu.Gnomes.Checked(false)
				}
			} else {
				connect_box.Disconnect.SetChecked(false)
			}

			if rpc.Daemon.IsConnected() {
				rpc.Signal.Startup = false
			}

		case <-quit: // exit
			log.Println("[DerBNB] Closing")
			if menu.Gnomes.Icon_ind != nil {
				menu.Gnomes.Icon_ind.Stop()
			}
			ticker.Stop()
			done <- struct{}{}
			return
		}
	}
}

// Check if property entry exists already in my_properties
func haveProperty(text string) (have bool) {
	my_properties.RLock()
	for _, prop := range my_properties.data[""] {
		if text == prop {
			have = true
			break
		}
	}
	my_properties.RUnlock()

	return
}

// Check if string is a member of list
func previouslyAdded(check string, list []string) bool {
	for _, s := range list {
		if check == s {
			return true
		}
	}

	return false
}

// Define https or ipfs urls
func propertyImageSource(check string) string {
	if len(check) < 8 {
		return ""
	}

	if check[0:8] == "https://" {
		return check
	}

	return "https://ipfs.io/ipfs/" + check
}

// Make a formatted City, Country location string from SCID
func makeLocationString(scid string) (location string) {
	city, country := getLocation(scid)
	if city != "" && country != "" {
		location = fmt.Sprintf("%s, %s", city, country)
	} else if country != "" {
		location = country
	}

	return
}

func filterProperty(check interface{}) bool {
	for _, addr := range property_filter {
		if a, ok := check.(string); ok && a == addr {
			return true
		}
	}
	return false
}

func SearchProperties(prefix string, search_city bool) (results []string) {
	if rpc.IsReady() {
		if menu.Gnomes.IsReady() {
			info := menu.Gnomes.GetAllSCIDVariableDetails(rpc.DerBnbSCID)
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

				for _, h := range info[int64(keys[len(keys)-1])] {
					split := strings.Split(h.Key.(string), "_")
					if len(split) > 1 && len(split[0]) == 64 {
						var have bool
						if split[1] == "owner" {
							if filterProperty(h.Value) {
								continue
							}

							if !have {
								location := makeLocationString(split[0])
								if search_city {
									if strings.HasPrefix(location, prefix) {
										list_string := fmt.Sprintf("%s   %s", location, split[0])
										results = append(results, list_string)
									}
								} else {
									if strings.HasPrefix(location, prefix) {
										list_string := fmt.Sprintf("%s   %s", location, split[0])
										results = append(results, list_string)
									}

									if country_check := strings.Split(location, ", "); country_check != nil {
										l := len(country_check)
										if l > 1 {
											if strings.HasPrefix(country_check[l-1], prefix) {
												list_string := fmt.Sprintf("%s   %s", location, split[0])
												results = append(results, list_string)
											}
										}
									}
								}
							}
						}
					}
				}

				searching_properties = true
				sort.Strings(results)
				listed_properties = results
				listings_list.Refresh()
			}
		}
	}

	return
}

// Get all DerBnb property info from contract
func GetProperties() {
	if rpc.IsReady() {
		if menu.Gnomes.IsReady() {
			info := menu.Gnomes.GetAllSCIDVariableDetails(rpc.DerBnbSCID)
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

				lp := []string{}

				mpm := make(map[string][]string)
				mpm[""] = []string{}

				mbm := make(map[string][]string)
				mbm[""] = []string{"Completed", "Confirmed", "Requests"}

				var doubles []string
				added_bookings := make(map[string]bool)
				for _, h := range info[int64(keys[len(keys)-1])] {
					split := strings.Split(h.Key.(string), "_")
					l := len(split)
					if l > 1 && len(split[0]) == 64 {
						var have bool
						for _, p := range lp {
							if split[0] == p {
								have = true
								break
							}
						}

						if split[1] == "owner" {
							if filterProperty(h.Value) {
								continue
							}

							if !have && !searching_properties {
								location := makeLocationString(split[0])
								if location != "" {
									list_string := fmt.Sprintf("%s   %s", location, split[0])
									lp = append(lp, list_string)
								}
							}

							// find owned properties
							if h.Value.(string) == rpc.Wallet.Address {
								mpm[""] = append(mpm[""], split[0])
							}
						}

						// find user confirmed booking requests
						if !previouslyAdded(split[0], doubles) && split[1] == "booker" && h.Value.(string) == rpc.Wallet.Address {
							booked, complete := getUserConfirmedBookings(split[0], false)
							mbm["Confirmed"] = append(mbm["Confirmed"], booked...)
							mbm["Completed"] = append(mbm["Completed"], complete...)
							doubles = append(doubles, split[0])
						}

						if l > 3 {
							// find request bookings of property
							if split[1] == "request" && split[2] == "bk" && split[3] == "end" {
								add := true
								for _, prop := range mpm[split[0]] {
									check := strings.Split(prop, "   ")
									if check[0] == split[len(split)-1] {
										add = false
										break
									}
								}

								if add {
									requests := getBookingRequests(split[0], split[len(split)-1], true)
									if requests != "" {
										mpm[split[0]] = append(mpm[split[0]], requests)
									}
								}
							}

							// find user booking requests
							if split[2] == "booker" && h.Value.(string) == rpc.Wallet.Address {
								mbm["Requests"] = append(mbm["Requests"], getBookingRequests(split[0], split[len(split)-1], false))
							}

							// find owner confirmed booking requests
							if !added_bookings[split[0]] && split[1] == "bk" && split[2] == "end" {
								bookings := getOwnerConfirmedBookings(split[0], true)
								if bookings != nil {
									mpm[split[0]] = append(mpm[split[0]], bookings...)
									sort.Strings(mpm[split[0]])
									added_bookings[split[0]] = true
								}
							}
						}
					}
				}
				listing_label.SetText(getInfo(viewing_scid))
				if !searching_properties {
					sort.Strings(lp)
					listed_properties = lp
					listings_list.Refresh()
				}

				my_properties.Lock()
				for sc := range my_properties.data {
					my_properties.data[sc] = []string{}
				}

				for i, s := range mpm {
					my_properties.data[i] = s
				}
				my_properties.Unlock()

				my_properties.RLock()
				property_list.Refresh()
				my_properties.RUnlock()

				my_bookings.Lock()
				for _, curr := range my_bookings.data[""] {
					my_bookings.data[curr] = []string{}
				}

				for i, s := range mbm {
					my_bookings.data[i] = s
				}
				my_bookings.Unlock()

				my_bookings.RLock()
				booking_list.Refresh()
				my_bookings.RUnlock()
			}
		}
	} else {
		listed_properties = []string{}
		listings_list.Refresh()

		my_properties.Lock()
		my_properties.data[""] = []string{}
		my_properties.Unlock()
		my_properties.RLock()
		property_list.Refresh()
		my_properties.RUnlock()

		my_bookings.Lock()
		my_bookings.data[""] = []string{"Completed", "Confirmed", "Requests"}
		for _, curr := range my_bookings.data[""] {
			my_bookings.data[curr] = []string{}
		}
		my_bookings.Unlock()

		my_bookings.RLock()
		booking_list.Refresh()
		my_bookings.RUnlock()
	}
}

// Get request to owners and of the renters
//   - stamp is timestamp key of request
func getBookingRequests(scid, stamp string, owned bool) (request string) {
	if menu.Gnomes.IsReady() {
		_, start := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_request_bk_start_"+stamp)
		_, end := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_request_bk_end_"+stamp)
		booker, _ := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_request_booker_"+stamp)

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
//   - stamp is timestamp key of request
func getOwnerConfirmedBookings(scid string, all bool) (bookings []string) {
	if menu.Gnomes.IsReady() {
		_, bk_last := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_bk_last")
		if bk_last != nil {
			i := int(bk_last[0])
			for {
				last := strconv.Itoa(i)
				_, start := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_bk_start_"+last)
				_, end := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_bk_end_"+last)
				booker, _ := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_booker_"+last)
				complete, _ := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_"+last+"_damage_renter")

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
	if menu.Gnomes.IsReady() {
		owner, _ := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_owner")
		_, price := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_price")
		_, dep := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_damage_deposit")
		//_, bk_last := menu.Gnomes.GetSCIDValuesByKey(CONTRACT, scid+"_bk_last")

		if owner != nil && price != nil && dep != nil {
			location := makeLocationString(scid)
			data := getMetadata(scid)
			data_string := fmt.Sprintf("Sq feet: %d\n\nStyle: %s\n\nBedrooms: %d\n\nMax guests: %d\n\nDescription: %s", data.Squarefootage, data.Style, data.NumberOfBedrooms, data.MaxNumberOfGuests, data.Description)
			current_price = price[0]
			current_deposit = dep[0]
			info = fmt.Sprintf("Location: %s\n\nPrice: %s Dero per night\n\nDamage Deposit: %s Dero\n\nOwner: %s\n\n%s", location, walletapi.FormatMoney(price[0]), walletapi.FormatMoney(dep[0]), owner[0], data_string)
		}
	}

	return
}

// Get owner address of DerBnb property
func getOwnerAddress(scid string) (address string) {
	if menu.Gnomes.IsReady() {
		owner, _ := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_owner")
		if owner != nil {
			return owner[0]
		}
	}

	return
}

// Get confirmed  and completed bookings for renters
//   - stamp is timestamp key of request
func getUserConfirmedBookings(scid string, all bool) (confirmed_bookings []string, complete_bookings []string) {
	if menu.Gnomes.IsReady() {
		_, bk_last := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_bk_last")
		if bk_last != nil {
			i := int(bk_last[0])
			for {
				last := strconv.Itoa(i)
				_, start := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_bk_start_"+last)
				_, end := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_bk_end_"+last)
				booker, _ := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_booker_"+last)
				complete, _ := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, scid+"_"+last+"_damage_renter")

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
								complete_bookings = append(complete_bookings, fmt.Sprintf("%s   %s   %s   %s   %s", prefix, last, s.Format(TIME_FORMAT), e.Format(TIME_FORMAT), scid))
							default:

							}
						}
					} else {
						switch prefix {
						case "Booked:":
							confirmed_bookings = append(confirmed_bookings, fmt.Sprintf("%s   %s   %s   %s   %s", prefix, last, s.Format(TIME_FORMAT), e.Format(TIME_FORMAT), booker[0]))
						case "Complete:":
							complete_bookings = append(complete_bookings, fmt.Sprintf("%s   %s   %s   %s   %s", prefix, last, s.Format(TIME_FORMAT), e.Format(TIME_FORMAT), booker[0]))
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

func getUserShares() (shares, epoch, treasury uint64) {
	if menu.Gnomes.IsReady() {
		if _, check_shares := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, rpc.Wallet.Address+"_SHARES"); check_shares != nil {
			shares = check_shares[0]
			if _, check_epoch := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, rpc.Wallet.Address+"_EPOCH"); check_epoch != nil {
				epoch = check_epoch[0]
				if _, check_treasury := menu.Gnomes.GetSCIDValuesByKey(rpc.DerBnbSCID, "TREASURY"); check_treasury != nil {
					treasury = check_treasury[0]
				}
			}
		}
	}

	return
}
