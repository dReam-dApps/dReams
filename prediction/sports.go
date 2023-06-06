package prediction

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type sportsItems struct {
	Contract      string
	Buffer        bool
	Info          *widget.Label
	Sports_list   *widget.List
	Favorite_list *widget.List
	Owned_list    *widget.List
	Payout_log    *widget.Entry

	Game_select  *widget.Select
	Game_options []string
	Multi        *widget.RadioGroup
	ButtonA      *widget.Button
	ButtonB      *widget.Button
	Sports_box   *fyne.Container
}

var Sports sportsItems

// Disable dSports objects
func DisableSports(d bool) {
	if d {
		Sports.Sports_box.Hide()
		menu.Control.Sports_check.SetChecked(false)
	}

	Sports.Sports_box.Refresh()
}

// Check box for dSports SCID
//   - Hides sports controls on disconnect
func SportsConnectedBox() fyne.Widget {
	menu.Control.Sports_check = widget.NewCheck("", func(b bool) {
		if !b {
			Sports.Game_select.Hide()
			Sports.Multi.Hide()
			Sports.ButtonA.Hide()
			Sports.ButtonB.Hide()
		}
	})
	menu.Control.Sports_check.Disable()

	return menu.Control.Sports_check
}

// Entry for dPrediction SCID
//   - Bound to Sports.Contract
//   - Checks for valid SCID on changed
func SportsContractEntry() fyne.Widget {
	options := []string{""}
	menu.Control.S_contract = widget.NewSelectEntry(options)
	menu.Control.S_contract.PlaceHolder = "Contract Address: "
	menu.Control.S_contract.OnCursorChanged = func() {
		if rpc.Daemon.Connect {
			go func() {
				if len(Sports.Contract) == 64 {
					yes := rpc.ValidBetContract(Sports.Contract)
					if yes {
						menu.Control.Sports_check.SetChecked(true)
						if !menu.CheckActiveGames(Sports.Contract) {
							Sports.Game_select.Show()
						} else {
							Sports.Game_select.Hide()
						}
					} else {
						menu.Control.Sports_check.SetChecked(false)
					}
				} else {
					menu.Control.Sports_check.SetChecked(false)
				}
			}()
		}
	}

	this := binding.BindString(&Sports.Contract)
	menu.Control.S_contract.Bind(this)

	return menu.Control.S_contract
}

// Routine when dSports SCID is clicked
//   - Sets label info, controls and payout log
//   - item returned for adding and removing favorite
func setSportsControls(str string) (item string) {
	Sports.Game_select.ClearSelected()
	Sports.Game_select.Options = []string{}
	Sports.Game_select.Refresh()
	split := strings.Split(str, "   ")
	if len(split) >= 3 {
		trimmed := strings.Trim(split[2], " ")
		Sports.Sports_box.Show()
		if len(trimmed) == 64 {
			go SetSportsInfo(trimmed)
			item = str
			menu.Control.S_contract.SetText(trimmed)
			finals := rpc.FetchSportsFinal(trimmed)
			Sports.Payout_log.SetText(formatFinals(trimmed, finals))
		}
	}

	return
}

// Format all dSports final results from SCID
//   - Pass all final strings and splitting for formatting
func formatFinals(scid string, finals []string) (text string) {
	text = "Last Payouts from SCID:\n\n" + scid
	for i := range finals {
		split := strings.Split(finals[i], "   ")
		game := strings.Split(split[1], "_")
		var str string
		l := len(game)
		if l == 4 {
			str = "Game #" + split[0] + "\n" + game[1] + "  Winner: " + WinningTeam(game[1], game[3])
		} else if l >= 2 {
			str = "Game #" + split[0] + "\n" + game[1] + "  Tie"
		}
		text = text + "\n\n" + str + "\nTXID: " + split[2]
	}

	return
}

// Format winning team name from dSports final result string
func WinningTeam(teams, winner string) string {
	split := strings.Split(teams, "--")
	if len(split) >= 2 {
		switch winner {
		case "a":
			return split[0]
		case "b":
			return split[1]
		default:
			return ""
		}
	}
	return ""
}

// Set dSports info label
func SetSportsInfo(scid string) {
	info := GetBook(scid)
	Sports.Info.SetText(info)
	Sports.Info.Refresh()
	Sports.Game_select.Refresh()
}

// List object for populating public dSports contracts, with rating and add favorite controls
//   - Pass tab for action confirmation reset
func SportsListings(tab *container.AppTabs) fyne.CanvasObject {
	Sports.Sports_list = widget.NewList(
		func() int {
			return len(menu.Control.Sports_contracts)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(canvas.NewImageFromImage(nil), widget.NewLabel(""))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*fyne.Container).Objects[1].(*widget.Label).SetText(menu.Control.Sports_contracts[i])
			if menu.Control.Sports_contracts[i][0:2] != "  " {
				var key string
				split := strings.Split(menu.Control.Sports_contracts[i], "   ")
				if len(split) >= 3 {
					trimmed := strings.Trim(split[2], " ")
					if len(trimmed) == 64 {
						key = trimmed
					}
				}

				badge := canvas.NewImageFromResource(menu.DisplayRating(menu.Control.Contract_rating[key]))
				badge.SetMinSize(fyne.NewSize(35, 35))
				o.(*fyne.Container).Objects[0] = badge
			}
		})

	var item string

	Sports.Sports_list.OnSelected = func(id widget.ListItemID) {
		if id != 0 && menu.Connected() {
			item = setSportsControls(menu.Control.Sports_contracts[id])
			Sports.Favorite_list.UnselectAll()
			Sports.Owned_list.UnselectAll()
		} else {
			Sports.Sports_box.Hide()
		}
	}

	save := widget.NewButton("Favorite", func() {
		menu.Control.Sports_favorites = append(menu.Control.Sports_favorites, item)
		sort.Strings(menu.Control.Sports_favorites)
	})

	rate := widget.NewButton("Rate", func() {
		if len(Sports.Contract) == 64 {
			if !menu.CheckOwner(Sports.Contract) {
				reset := tab.Selected().Content
				tab.Selected().Content = menu.RateConfirm(Sports.Contract, tab, reset)
				tab.Selected().Content.Refresh()
			} else {
				log.Println("[dReams] You own this contract")
			}
		}
	})

	cont := container.NewBorder(
		nil,
		container.NewBorder(nil, nil, save, rate, layout.NewSpacer()),
		nil,
		nil,
		Sports.Sports_list)

	return cont
}

// List object for populating favorite dSports contracts, with remove favorite control
func SportsFavorites() fyne.CanvasObject {
	Sports.Favorite_list = widget.NewList(
		func() int {
			return len(menu.Control.Sports_favorites)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(menu.Control.Sports_favorites[i])
		})

	var item string

	Sports.Favorite_list.OnSelected = func(id widget.ListItemID) {
		if menu.Connected() {
			item = setSportsControls(menu.Control.Sports_favorites[id])
			Sports.Sports_list.UnselectAll()
			Sports.Owned_list.UnselectAll()
		} else {
			Sports.Sports_box.Hide()
		}
	}

	remove := widget.NewButton("Remove", func() {
		if len(menu.Control.Sports_favorites) > 0 {
			Sports.Favorite_list.UnselectAll()
			new := menu.Control.Sports_favorites
			for i := range new {
				if new[i] == item {
					copy(new[i:], new[i+1:])
					new[len(new)-1] = ""
					new = new[:len(new)-1]
					menu.Control.Sports_favorites = new
					break
				}
			}
		}
		Sports.Favorite_list.Refresh()
		sort.Strings(menu.Control.Sports_favorites)
	})

	cont := container.NewBorder(
		nil,
		container.NewBorder(nil, nil, nil, remove, layout.NewSpacer()),
		nil,
		nil,
		Sports.Favorite_list)

	return cont
}

// List object for populating owned dSports contracts
func SportsOwned() fyne.CanvasObject {
	Sports.Owned_list = widget.NewList(
		func() int {
			return len(menu.Control.Sports_owned)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(menu.Control.Sports_owned[i])
		})

	Sports.Owned_list.OnSelected = func(id widget.ListItemID) {
		if menu.Connected() {
			setSportsControls(menu.Control.Sports_owned[id])
			Sports.Sports_list.UnselectAll()
			Sports.Favorite_list.UnselectAll()
		} else {
			Sports.Sports_box.Hide()
		}
	}

	return Sports.Owned_list
}

// Log entry for dSports payout results
func SportsPayouts() fyne.CanvasObject {
	Sports.Payout_log = widget.NewMultiLineEntry()
	Sports.Payout_log.Disable()

	return Sports.Payout_log
}

// Gets dSports data from SCID and return formatted info string
func GetBook(scid string) (info string) {
	if menu.Gnomes.Init && !menu.GnomonClosing() && menu.Gnomes.Sync {
		_, initValue := menu.Gnomes.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, "s_init", menu.Gnomes.Indexer.ChainHeight, true)
		if initValue != nil {
			_, playedValue := menu.Gnomes.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, "s_played", menu.Gnomes.Indexer.ChainHeight, true)
			//_, hl := menu.Gnomes.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, "hl", menu.Gnomes.Indexer.ChainHeight, true)
			init := initValue[0]
			played := playedValue[0]

			Sports.Game_options = []string{}
			Sports.Game_select.Options = Sports.Game_options
			played_str := strconv.Itoa(int(played))
			if init == played {
				info = "SCID:\n\n" + scid + "\n\nGames Completed: " + played_str + "\n\nNo current Games\n"
				Sports.Buffer = false
				return
			}

			var single bool
			iv := 1
			for {
				_, s_init := menu.Gnomes.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, "s_init_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
				if s_init != nil {
					game, _ := menu.Gnomes.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, "game_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					league, _ := menu.Gnomes.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, "league_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, s_n := menu.Gnomes.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, "s_#_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, s_amt := menu.Gnomes.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, "s_amount_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, s_end := menu.Gnomes.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, "s_end_at_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, s_total := menu.Gnomes.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, "s_total_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					//s_urlValue, _ := menu.Gnomes.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, "s_url_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, s_ta := menu.Gnomes.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, "team_a_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, s_tb := menu.Gnomes.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, "team_b_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, time_a := menu.Gnomes.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, "time_a", menu.Gnomes.Indexer.ChainHeight, true)
					_, time_b := menu.Gnomes.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, "time_b", menu.Gnomes.Indexer.ChainHeight, true)
					_, buffer := menu.Gnomes.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, "buffer"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)

					team_a := menu.TrimTeamA(game[0])
					team_b := menu.TrimTeamB(game[0])

					now := uint64(time.Now().Unix())
					if now < buffer[0] {
						Sports.Buffer = true
					} else {
						Sports.Buffer = false
					}

					if s_end[0] > now && now > buffer[0] {
						current := Sports.Game_select.Options
						new := append(current, strconv.Itoa(iv)+"   "+game[0])
						Sports.Game_select.Options = new
					}

					eA := fmt.Sprint(s_end[0] * 1000)
					min := fmt.Sprint(float64(s_amt[0]) / 100000)
					n := strconv.Itoa(int(s_n[0]))
					aV := strconv.Itoa(int(s_ta[0]))
					bV := strconv.Itoa(int(s_tb[0]))
					t := strconv.Itoa(int(s_total[0]))
					if !single {
						single = true
						info = "SCID:\n\n" + scid + "\n\nGames Completed: " + played_str + "\nCurrent Games:\n"
					}
					info = info + S_Results(game[0], strconv.Itoa(iv), league[0], min, eA, n, team_a, team_b, aV, bV, t, time_a[0], time_b[0])

				}

				if iv >= int(init) {
					break
				}

				iv++
			}
		}
	}

	return
}

// Formats dSports info string
//   - g is game name
//   - gN is game number
//   - l is league
//   - min is minimum Dero wager
//   - eA is game end time
//   - c is current number of picks
//   - tA, tB are team names of A and B
//   - tAV, tBV is total picks for A or B
//   - total is current game Dero pot total
//   - a, b are current contract time frames
func S_Results(g, gN, l, min, eA, c, tA, tB, tAV, tBV, total string, a, b uint64) (info string) { /// sports info label
	result, err := strconv.ParseFloat(total, 32)

	if err != nil {
		log.Println("[Sports]", err)
	}

	s := fmt.Sprintf("%.5f", result/100000)
	end_time, _ := rpc.MsToTime(eA)
	utc_end := end_time.String()

	pa := strconv.Itoa(int(a/60) / 60)
	rf := strconv.Itoa(int(b/60) / 60)

	event := "Game "
	switch l {
	case "Bellator":
		event = "Fight "
	case "UFC":
		event = "Fight "
	default:

	}

	info = ("\n" + event + gN + ": " + g + "\nLeague: " + l + "\nMinimum: " + min +
		" Dero\nCloses at: " + utc_end + "\nPayout " + pa + " hours after close\nRefund if not paid " + rf + " within hours\nPot Total: " + s + "\nPicks: " + c + "\n" + tA + " Picks: " + tAV + "\n" + tB + " Picks: " + tBV + "\n")

	return
}

// Switch for sports api prefix
func sports(league string) (api string) {
	switch league {
	// case "FIFA":
	// 	api = "http://site.api.espn.com/apis/site/v2/sports/soccer/fifa.world/scoreboard"
	case "EPL":
		api = "http://site.api.espn.com/apis/site/v2/sports/soccer/eng.1/scoreboard"
	case "MLS":
		api = "http://site.api.espn.com/apis/site/v2/sports/soccer/usa.1/scoreboard"

	case "NFL":
		api = "http://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard"
	case "NBA":
		api = "http://site.api.espn.com/apis/site/v2/sports/basketball/nba/scoreboard"
	case "NHL":
		api = "http://site.api.espn.com/apis/site/v2/sports/hockey/nhl/scoreboard"
	case "MLB":
		api = "http://site.api.espn.com/apis/site/v2/sports/baseball/mlb/scoreboard"
	case "UFC":
		api = "http://site.api.espn.com/apis/site/v2/sports/mma/ufc/scoreboard"
	case "Bellator":
		api = "http://site.api.espn.com/apis/site/v2/sports/mma/bellator/scoreboard"
	default:
		api = ""
	}

	return api
}

// Gets the games of league for following week
func GetCurrentWeek(league string) {
	for i := 0; i < 8; i++ {
		now := time.Now().AddDate(0, 0, i)
		date := time.Unix(now.Unix(), 0).String()
		date = date[0:10]
		comp := date[0:4] + date[5:7] + date[8:10]
		switch league {
		case "EPL":
			GetSoccer(comp, league)
		case "MLS":
			GetSoccer(comp, league)
		case "NBA":
			GetBasketball(comp, league)
		case "NFL":
			GetFootball(comp, league)
		case "NHL":
			GetHockey(comp, league)
		case "MLB":
			GetBaseball(comp, league)
		default:

		}
	}
}

// Gets the games of league for following month
func GetCurrentMonth(league string) {
	for i := 0; i < 45; i++ {
		now := time.Now().AddDate(0, 0, i)
		date := time.Unix(now.Unix(), 0).String()
		date = date[0:10]
		comp := date[0:4] + date[5:7] + date[8:10]
		switch league {
		case "UFC":
			GetMma(comp, league)
		case "Bellator":
			GetMma(comp, league)
		default:

		}
	}
}

// Call soccer api with chosen date and league
func callSoccer(date, league string) (s *soccer) {
	client := &http.Client{Timeout: 9 * time.Second}
	req, err := http.NewRequest("GET", sports(league)+"?dates="+date, nil)
	if err != nil {
		log.Println("[callSoccer]", err)
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Println("[callSoccer]", err)
		return
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("[callSoccer]", err)
		return
	}

	json.Unmarshal(b, &s)

	return s
}

// Call mma api with chosen date and league
func callMma(date, league string) (m *mma) {
	client := &http.Client{Timeout: 9 * time.Second}
	req, err := http.NewRequest("GET", sports(league)+"?dates="+date, nil)
	if err != nil {
		log.Println("[callMma]", err)
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Println("[callMma]", err)
		return
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("[callMma]", err)
		return
	}

	json.Unmarshal(b, &m)

	return m
}

// Call basketball api with chosen date and league
func callBasketball(date, league string) (bb *basketball) {
	client := &http.Client{Timeout: 9 * time.Second}
	req, err := http.NewRequest("GET", sports(league)+"?dates="+date, nil)
	if err != nil {
		log.Println("[callBasketball]", err)
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Println("[callBasketball]", err)
		return
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("[callBasketball]", err)
		return
	}

	json.Unmarshal(b, &bb)

	return bb
}

// Call basketball api with chosen date and league
func callBaseball(date, league string) (baseb *baseball) {
	client := &http.Client{Timeout: 9 * time.Second}
	req, err := http.NewRequest("GET", sports(league)+"?dates="+date, nil)
	if err != nil {
		log.Println("[callBaseball]", err)
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Println("[callBaseball]", err)
		return
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("[callBaseball]", err)
		return
	}

	json.Unmarshal(b, &baseb)

	return baseb
}

// Call football api with chosen date and league
func callFootball(date, league string) (f *football) {
	client := &http.Client{Timeout: 9 * time.Second}
	req, err := http.NewRequest("GET", sports(league)+"?dates="+date, nil)
	if err != nil {
		log.Println("[callFootball]", err)
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Println("[callFootball]", err)
		return
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("[callFootball]", err)
		return
	}

	json.Unmarshal(b, &f)

	return f
}

// Call hockey api with chosen date and league
func callHockey(date, league string) (h *hockey) {
	client := &http.Client{Timeout: 9 * time.Second}
	req, err := http.NewRequest("GET", sports(league)+"?dates="+date, nil)
	if err != nil {
		log.Println("[callHockey]", err)
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Println("[callHockey]", err)
		return
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("[callHockey]", err)
		return
	}

	json.Unmarshal(b, &h)

	return h
}

// Find and display the end time of selected game
func GetGameEnd(date, game, league string) {
	client := &http.Client{Timeout: 9 * time.Second}
	req, err := http.NewRequest("GET", sports(league)+"?dates="+date, nil)

	if err != nil {
		log.Println("[GetGameEnd]", err)
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Println("[GetGameEnd]", err)
		return
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("[GetGameEnd]", err)
		return
	}

	if league == "UFC" || league == "Bellator" {
		var found mma
		json.Unmarshal(b, &found)
		for i := range found.Events {
			trimmed := strings.Trim(found.Events[i].Competitions[0].StartDate, "Z")
			utc_time, err := time.Parse("2006-01-02T15:04", trimmed)
			if err != nil {
				log.Println("[GetGameEnd]", err)
			}

			for f := range found.Events[i].Competitions {
				a := found.Events[i].Competitions[f].Competitors[0].Athlete.DisplayName
				b := found.Events[i].Competitions[f].Competitors[1].Athlete.DisplayName
				g := a + "--" + b

				if g == game {
					PS_Control.S_end.SetText(strconv.Itoa(int(utc_time.Unix())))
					return
				}
			}

		}
	} else {
		var found scores
		json.Unmarshal(b, &found)
		for i := range found.Events {
			trimmed := strings.Trim(found.Events[i].Competitions[0].StartDate, "Z")
			utc_time, err := time.Parse("2006-01-02T15:04", trimmed)
			if err != nil {
				log.Println("[GetGameEnd]", err)
			}

			a := found.Events[i].Competitions[0].Competitors[0].Team.Abbreviation
			b := found.Events[i].Competitions[0].Competitors[1].Team.Abbreviation
			g := a + "--" + b
			if g == game {
				PS_Control.S_end.SetText(strconv.Itoa(int(utc_time.Unix())))
				return
			}
		}
	}
}

// Call api for scores with chosen date and league
func callScores(date, league string) (s *scores) {
	client := &http.Client{Timeout: 9 * time.Second}
	req, err := http.NewRequest("GET", sports(league)+"?dates="+date, nil)
	if err != nil {
		log.Println("[callScores]", err)
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Println("[callScores]", err)
		return
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("[callScores]", err)
		return
	}

	json.Unmarshal(b, &s)

	return s
}

// Gets past game scores for league and display info
//   - Pass label for display info
func GetScores(label *widget.Label, league string) {
	var single bool
	for i := -1; i < 1; i++ {
		day := time.Now().AddDate(0, 0, i)
		date := time.Unix(day.Unix(), 0).String()
		date = date[0:10]
		comp := date[0:4] + date[5:7] + date[8:10]
		found := callScores(comp, league)
		if found != nil {
			if !single {
				label.SetText(found.Leagues[0].Abbreviation + "\n" + found.Day.Date + "\n")
			}

			for i := range found.Events {
				trimmed := strings.Trim(found.Events[i].Competitions[0].StartDate, "Z")
				utc_time, err := time.Parse("2006-01-02T15:04", trimmed)
				if err != nil {
					log.Println("[GetScores]", err)
				}

				tz, _ := time.LoadLocation("Local")
				local := utc_time.In(tz).String()
				state := found.Events[i].Competitions[0].Status.Type.State
				team_a := found.Events[i].Competitions[0].Competitors[0].Team.Abbreviation
				team_b := found.Events[i].Competitions[0].Competitors[1].Team.Abbreviation
				score_a := found.Events[i].Competitions[0].Competitors[0].Score
				score_b := found.Events[i].Competitions[0].Competitors[1].Score
				period := found.Events[i].Status.Period
				clock := found.Events[i].Competitions[0].Status.DisplayClock
				complete := found.Events[i].Status.Type.Completed

				var format string
				switch league {
				case "EPL":
					format = " Half "
				case "NBA":
					format = " Quarter "
				case "NFL":
					format = " Quarter "
				case "NHL":
					format = " Period "
				default:
				}

				var abv string
				switch period {
				case 0:
					abv = ""
				case 1:
					abv = "st "
				case 2:
					abv = "nd "
				case 3:
					abv = "rd "
				case 4:
					abv = "th "
				default:
					abv = "th "
				}
				if state == "pre" {
					label.SetText(label.Text + team_a + " - " + team_b + "\nStart time: " + local + "\nState: " + state + "\nComplete: " + strconv.FormatBool(complete) + "\n\n")
				} else {
					label.SetText(label.Text + team_a + " - " + team_b + "\nStart time: " + local + "\nState: " + state +
						"\n" + strconv.Itoa(period) + abv + format + " " + clock + "\n" + team_a + ": " + score_a + "\n" + team_b + ": " + score_b + "\nComplete: " + strconv.FormatBool(complete) + "\n\n")
				}

				single = true
			}
		}
	}
	label.Refresh()
}

// Get final result of mma league and display info
//   - Pass label for display info
func GetMmaResults(label *widget.Label, league string) {
	var single bool
	for i := -15; i < 1; i++ {
		day := time.Now().AddDate(0, 0, i)
		date := time.Unix(day.Unix(), 0).String()
		date = date[0:10]
		comp := date[0:4] + date[5:7] + date[8:10]
		found := callMma(comp, league)
		if found != nil {
			if !single {
				label.SetText(found.Leagues[0].Abbreviation + "\n" + found.Day.Date + "\n")
			}

			for i := range found.Events {
				trimmed := strings.Trim(found.Events[i].Competitions[0].StartDate, "Z")
				utc_time, err := time.Parse("2006-01-02T15:04", trimmed)
				if err != nil {
					log.Println("[GetMmaResults]", err)
				}

				tz, _ := time.LoadLocation("Local")
				local := utc_time.In(tz).String()

				for f := range found.Events[i].Competitions {
					state := found.Events[i].Competitions[f].Status.Type.State
					team_a := found.Events[i].Competitions[f].Competitors[0].Athlete.DisplayName
					team_b := found.Events[i].Competitions[f].Competitors[1].Athlete.DisplayName
					winner_a := found.Events[i].Competitions[f].Competitors[0].Winner
					winner_b := found.Events[i].Competitions[f].Competitors[1].Winner
					period := found.Events[i].Competitions[f].Status.Period
					clock := found.Events[i].Competitions[f].Status.DisplayClock
					complete := found.Events[i].Competitions[f].Status.Type.Completed

					var abv string
					switch period {
					case 0:
						abv = ""
					case 1:
						abv = "st "
					case 2:
						abv = "nd "
					case 3:
						abv = "rd "
					case 4:
						abv = "th "
					default:
						abv = "th "
					}
					if state == "pre" {
						label.SetText(label.Text + team_a + " - " + team_b + "\nStart time: " + local + "\nState: " + state + "\nComplete: " + strconv.FormatBool(complete) + "\n\n")
					} else {
						var winner string
						if winner_a {
							winner = team_a
						} else if winner_b {
							winner = team_b
						} else {
							winner = "Draw"
						}
						label.SetText(label.Text + team_a + " - " + team_b + "\nStart time: " + local + "\nState: " + state +
							"\n" + strconv.Itoa(period) + abv + "Round " + " " + clock + "\nWinner: " + winner + "\nComplete: " + strconv.FormatBool(complete) + "\n\n")
					}

					single = true
				}
			}
		}
	}
	label.Refresh()
}

// Gets hockey games for selected league and date
func GetHockey(date, league string) {
	found := callHockey(date, league)
	if found != nil {
		for i := range found.Events {
			pregame := found.Events[i].Competitions[0].Status.Type.State
			trimmed := strings.Trim(found.Events[i].Competitions[0].StartDate, "Z")
			utc_time, err := time.Parse("2006-01-02T15:04", trimmed)
			if err != nil {
				log.Println("[GetHockey]", err)
			}

			tz, _ := time.LoadLocation("Local")

			teamA := found.Events[i].Competitions[0].Competitors[0].Team.Abbreviation
			teamB := found.Events[i].Competitions[0].Competitors[1].Team.Abbreviation

			if !found.Events[i].Status.Type.Completed && pregame == "pre" {
				current := PS_Control.S_game.Options
				new := append(current, utc_time.In(tz).String()[0:16]+"   "+teamA+"--"+teamB)
				PS_Control.S_game.Options = new
			}
		}
	}
}

// Gets hockey games for selected league and adds to options selection
//   - date GetCurrentWeek()
func GetSoccer(date, league string) {
	found := callSoccer(date, league)
	if found != nil {
		for i := range found.Events {
			pregame := found.Events[i].Competitions[0].Status.Type.State

			trimmed := strings.Trim(found.Events[i].Competitions[0].StartDate, "Z")
			utc_time, err := time.Parse("2006-01-02T15:04", trimmed)
			if err != nil {
				log.Println("[GetSoccer]", err)
			}

			tz, _ := time.LoadLocation("Local")

			teamA := found.Events[i].Competitions[0].Competitors[0].Team.Abbreviation
			teamB := found.Events[i].Competitions[0].Competitors[1].Team.Abbreviation

			if !found.Events[i].Status.Type.Completed && pregame == "pre" {
				current := PS_Control.S_game.Options
				new := append(current, utc_time.In(tz).String()[0:16]+"   "+teamA+"--"+teamB)
				PS_Control.S_game.Options = new
			}
		}
	}
}

// Gets and returns the winner of game
//   - league defines api prefix
func GetWinner(game, league string) (win string, team_name string, a_score string, b_score string) {
	for i := -3; i < 1; i++ {
		day := time.Now().AddDate(0, 0, i)
		date := time.Unix(day.Unix(), 0).String()
		date = date[0:10]
		comp := date[0:4] + date[5:7] + date[8:10]

		found := callScores(comp, league)
		if found != nil {
			for i := range found.Events {
				a := found.Events[i].Competitions[0].Competitors[0].Team.Abbreviation
				b := found.Events[i].Competitions[0].Competitors[1].Team.Abbreviation
				g := a + "--" + b

				if g == game {
					if found.Events[i].Status.Type.Completed {
						teamA := found.Events[i].Competitions[0].Competitors[0].Team.Abbreviation
						a_win := found.Events[i].Competitions[0].Competitors[0].Winner
						a_score = found.Events[i].Competitions[0].Competitors[0].Score

						teamB := found.Events[i].Competitions[0].Competitors[1].Team.Abbreviation
						b_win := found.Events[i].Competitions[0].Competitors[1].Winner
						b_score = found.Events[i].Competitions[0].Competitors[1].Score

						if a_win && !b_win {
							return "team_a", teamA, a_score, b_score
						} else if b_win && !a_win {
							return "team_b", teamB, a_score, b_score
						} else if a_score == b_score && a_score != "" && b_score != "" {
							return "", "Tie", a_score, b_score
						}
					}
				}
			}
		}
	}

	win = "invalid"

	return
}

// Gets and returns the winner of mma match
//   - league defines api prefix
func GetMmaWinner(game, league string) (win string, fighter string) {
	for i := -3; i < 1; i++ {
		day := time.Now().AddDate(0, 0, i)
		date := time.Unix(day.Unix(), 0).String()
		date = date[0:10]
		comp := date[0:4] + date[5:7] + date[8:10]

		found := callMma(comp, league)
		if found != nil {
			for i := range found.Events {
				for f := range found.Events[i].Competitions {
					a := found.Events[i].Competitions[f].Competitors[0].Athlete.DisplayName
					b := found.Events[i].Competitions[f].Competitors[1].Athlete.DisplayName
					g := a + "--" + b

					if g == game {
						if found.Events[i].Competitions[f].Status.Type.Completed {
							teamA := found.Events[i].Competitions[f].Competitors[0].Athlete.DisplayName
							a_win := found.Events[i].Competitions[f].Competitors[0].Winner

							teamB := found.Events[i].Competitions[f].Competitors[1].Athlete.DisplayName
							b_win := found.Events[i].Competitions[f].Competitors[1].Winner

							if a_win && !b_win {
								return "team_a", teamA
							} else if b_win && !a_win {
								return "team_b", teamB
							} else {
								return "", "Tie"
							}
						}
					}
				}
			}
		}
	}

	win = "invalid"

	return
}

// Gets football games for selected league and adds to options selection
//   - date GetCurrentWeek()
func GetFootball(date, league string) {
	found := callFootball(date, league)
	if found != nil {
		for i := range found.Events {
			pregame := found.Events[i].Competitions[0].Status.Type.State
			trimmed := strings.Trim(found.Events[i].Competitions[0].StartDate, "Z")
			utc_time, err := time.Parse("2006-01-02T15:04", trimmed)
			if err != nil {
				log.Println("[GetFootball]", err)
			}

			tz, _ := time.LoadLocation("Local")

			teamA := found.Events[i].Competitions[0].Competitors[0].Team.Abbreviation
			teamB := found.Events[i].Competitions[0].Competitors[1].Team.Abbreviation

			if !found.Events[i].Status.Type.Completed && pregame == "pre" {
				current := PS_Control.S_game.Options
				new := append(current, utc_time.In(tz).String()[0:16]+"   "+teamA+"--"+teamB)
				PS_Control.S_game.Options = new
			}
		}
	}
}

// Gets basketball games for selected league and adds to options selection
//   - date GetCurrentWeek()
func GetBasketball(date, league string) {
	found := callBasketball(date, league)
	if found != nil {
		for i := range found.Events {
			pregame := found.Events[i].Competitions[0].Status.Type.State
			trimmed := strings.Trim(found.Events[i].Competitions[0].StartDate, "Z")
			utc_time, err := time.Parse("2006-01-02T15:04", trimmed)
			if err != nil {
				log.Println("[GetBasketball]", err)
			}

			tz, _ := time.LoadLocation("Local")

			teamA := found.Events[i].Competitions[0].Competitors[0].Team.Abbreviation
			teamB := found.Events[i].Competitions[0].Competitors[1].Team.Abbreviation

			if !found.Events[i].Status.Type.Completed && pregame == "pre" {
				current := PS_Control.S_game.Options
				new := append(current, utc_time.In(tz).String()[0:16]+"   "+teamA+"--"+teamB)
				PS_Control.S_game.Options = new
			}
		}
	}
}

// Gets baseball games for selected league and adds to options selection
//   - date GetCurrentWeek()
func GetBaseball(date, league string) {
	found := callBaseball(date, league)
	if found != nil {
		for i := range found.Events {
			pregame := found.Events[i].Competitions[0].Status.Type.State
			trimmed := strings.Trim(found.Events[i].Competitions[0].StartDate, "Z")
			utc_time, err := time.Parse("2006-01-02T15:04", trimmed)
			if err != nil {
				log.Println("[GetBaseball]", err)
			}

			tz, _ := time.LoadLocation("Local")

			teamA := found.Events[i].Competitions[0].Competitors[0].Team.Abbreviation
			teamB := found.Events[i].Competitions[0].Competitors[1].Team.Abbreviation

			if !found.Events[i].Status.Type.Completed && pregame == "pre" {
				current := PS_Control.S_game.Options
				new := append(current, utc_time.In(tz).String()[0:16]+"   "+teamA+"--"+teamB)
				PS_Control.S_game.Options = new
			}
		}
	}
}

// Gets mma matches for selected league and adds to options selection
//   - date GetCurrentMonth()
func GetMma(date, league string) {
	found := callMma(date, league)
	if found != nil {
		for i := range found.Events {
			pregame := found.Events[i].Competitions[0].Status.Type.State
			trimmed := strings.Trim(found.Events[i].Competitions[0].StartDate, "Z")
			utc_time, err := time.Parse("2006-01-02T15:04", trimmed)
			if err != nil {
				log.Println("[GetMma]", err)
			}

			tz, _ := time.LoadLocation("Local")

			for f := range found.Events[i].Competitions {
				fighterA := found.Events[i].Competitions[f].Competitors[0].Athlete.DisplayName
				fighterB := found.Events[i].Competitions[f].Competitors[1].Athlete.DisplayName

				if !found.Events[i].Status.Type.Completed && pregame == "pre" {
					current := PS_Control.S_game.Options
					new := append(current, utc_time.In(tz).String()[0:16]+"   "+fighterA+"--"+fighterB)
					PS_Control.S_game.Options = new
				}
			}
		}
	}
}

type scores struct {
	Leagues []struct {
		ID           string `json:"id"`
		UID          string `json:"uid"`
		Name         string `json:"name"`
		Abbreviation string `json:"abbreviation"`
		MidsizeName  string `json:"midsizeName"`
		Slug         string `json:"slug"`
		Season       struct {
			Year      int    `json:"year"`
			StartDate string `json:"startDate"`
			EndDate   string `json:"endDate"`
			Type      struct {
				ID           string `json:"id"`
				Type         int    `json:"type"`
				Name         string `json:"name"`
				Abbreviation string `json:"abbreviation"`
			} `json:"type"`
		} `json:"season"`
		Logos               []struct{} `json:"logos"`
		CalendarType        string     `json:"calendarType"`
		CalendarIsWhitelist bool       `json:"calendarIsWhitelist"`
		CalendarStartDate   string     `json:"calendarStartDate"`
		CalendarEndDate     string     `json:"calendarEndDate"`
		Calendar            []string   `json:"calendar"`
	} `json:"leagues"`
	Season struct {
		Type int `json:"type"`
		Year int `json:"year"`
	} `json:"season"`
	Day struct {
		Date string `json:"date"`
	} `json:"day"`
	Events []struct {
		ID        string `json:"id"`
		UID       string `json:"uid"`
		Date      string `json:"date"`
		Name      string `json:"name"`
		ShortName string `json:"shortName"`
		Season    struct {
			Year int    `json:"year"`
			Type int    `json:"type"`
			Slug string `json:"slug"`
		} `json:"season"`
		Competitions []struct {
			ID         string `json:"id"`
			UID        string `json:"uid"`
			Date       string `json:"date"`
			StartDate  string `json:"startDate"`
			Attendance int    `json:"attendance"`
			TimeValid  bool   `json:"timeValid"`
			Recent     bool   `json:"recent"`
			Status     struct {
				Clock        float64 `json:"clock"`
				DisplayClock string  `json:"displayClock"`
				Period       int     `json:"period"`
				Type         struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					State       string `json:"state"`
					Completed   bool   `json:"completed"`
					Description string `json:"description"`
					Detail      string `json:"detail"`
					ShortDetail string `json:"shortDetail"`
				} `json:"type"`
			} `json:"status"`
			Venue         struct{}      `json:"venue"`
			Format        struct{}      `json:"format"`
			Notes         []interface{} `json:"notes"`
			GeoBroadcasts []interface{} `json:"geoBroadcasts"`
			Broadcasts    []interface{} `json:"broadcasts"`
			Competitors   []struct {
				ID       string     `json:"id"`
				UID      string     `json:"uid"`
				Type     string     `json:"type"`
				Order    int        `json:"order"`
				HomeAway string     `json:"homeAway"`
				Winner   bool       `json:"winner"`
				Form     string     `json:"form"`
				Score    string     `json:"score"`
				Records  []struct{} `json:"records"`
				Team     struct {
					ID               string     `json:"id"`
					UID              string     `json:"uid"`
					Abbreviation     string     `json:"abbreviation"`
					DisplayName      string     `json:"displayName"`
					ShortDisplayName string     `json:"shortDisplayName"`
					Name             string     `json:"name"`
					Location         string     `json:"location"`
					Color            string     `json:"color"`
					AlternateColor   string     `json:"alternateColor"`
					IsActive         bool       `json:"isActive"`
					Logo             string     `json:"logo"`
					Links            []struct{} `json:"links"`
					Venue            struct{}   `json:"venue"`
				} `json:"team,omitempty"`
				Statistics []struct{} `json:"statistics"`
			} `json:"competitors"`
			Details   []struct{} `json:"details"`
			Headlines []struct{} `json:"headlines"`
		} `json:"competitions"`
		Status struct {
			Clock        float64 `json:"clock"`
			DisplayClock string  `json:"displayClock"`
			Period       int     `json:"period"`
			Type         struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				State       string `json:"state"`
				Completed   bool   `json:"completed"`
				Description string `json:"description"`
				Detail      string `json:"detail"`
				ShortDetail string `json:"shortDetail"`
			} `json:"type"`
		} `json:"status"`
		Links []struct{} `json:"links"`
	} `json:"events"`
}

type soccer struct {
	Leagues []struct {
		ID           string `json:"id"`
		UID          string `json:"uid"`
		Name         string `json:"name"`
		Abbreviation string `json:"abbreviation"`
		MidsizeName  string `json:"midsizeName"`
		Slug         string `json:"slug"`
		Season       struct {
			Year      int    `json:"year"`
			StartDate string `json:"startDate"`
			EndDate   string `json:"endDate"`
			Type      struct {
				ID           string `json:"id"`
				Type         int    `json:"type"`
				Name         string `json:"name"`
				Abbreviation string `json:"abbreviation"`
			} `json:"type"`
		} `json:"season"`
		Logos []struct {
			Href        string   `json:"href"`
			Width       int      `json:"width"`
			Height      int      `json:"height"`
			Alt         string   `json:"alt"`
			Rel         []string `json:"rel"`
			LastUpdated string   `json:"lastUpdated"`
		} `json:"logos"`
		CalendarType        string   `json:"calendarType"`
		CalendarIsWhitelist bool     `json:"calendarIsWhitelist"`
		CalendarStartDate   string   `json:"calendarStartDate"`
		CalendarEndDate     string   `json:"calendarEndDate"`
		Calendar            []string `json:"calendar"`
	} `json:"leagues"`
	Season struct {
		Type int `json:"type"`
		Year int `json:"year"`
	} `json:"season"`
	Day struct {
		Date string `json:"date"`
	} `json:"day"`
	Events []struct {
		ID        string `json:"id"`
		UID       string `json:"uid"`
		Date      string `json:"date"`
		Name      string `json:"name"`
		ShortName string `json:"shortName"`
		Season    struct {
			Year int    `json:"year"`
			Type int    `json:"type"`
			Slug string `json:"slug"`
		} `json:"season"`
		Competitions []struct {
			ID         string `json:"id"`
			UID        string `json:"uid"`
			Date       string `json:"date"`
			StartDate  string `json:"startDate"`
			Attendance int    `json:"attendance"`
			TimeValid  bool   `json:"timeValid"`
			Recent     bool   `json:"recent"`
			Status     struct {
				Clock        float64 `json:"clock"`
				DisplayClock string  `json:"displayClock"`
				Period       int     `json:"period"`
				Type         struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					State       string `json:"state"`
					Completed   bool   `json:"completed"`
					Description string `json:"description"`
					Detail      string `json:"detail"`
					ShortDetail string `json:"shortDetail"`
				} `json:"type"`
			} `json:"status"`
			Venue struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
				Address  struct {
					City    string `json:"city"`
					Country string `json:"country"`
				} `json:"address"`
			} `json:"venue"`
			Format struct {
				Regulation struct {
					Periods int `json:"periods"`
				} `json:"regulation"`
			} `json:"format"`
			Notes         []interface{} `json:"notes"`
			GeoBroadcasts []interface{} `json:"geoBroadcasts"`
			Broadcasts    []interface{} `json:"broadcasts"`
			Competitors   []struct {
				ID       string `json:"id"`
				UID      string `json:"uid"`
				Type     string `json:"type"`
				Order    int    `json:"order"`
				HomeAway string `json:"homeAway"`
				Winner   bool   `json:"winner"`
				Form     string `json:"form"`
				Score    string `json:"score"`
				Records  []struct {
					Name         string `json:"name"`
					Type         string `json:"type"`
					Summary      string `json:"summary"`
					Abbreviation string `json:"abbreviation"`
				} `json:"records"`
				Team struct {
					ID               string `json:"id"`
					UID              string `json:"uid"`
					Abbreviation     string `json:"abbreviation"`
					DisplayName      string `json:"displayName"`
					ShortDisplayName string `json:"shortDisplayName"`
					Name             string `json:"name"`
					Location         string `json:"location"`
					Color            string `json:"color"`
					AlternateColor   string `json:"alternateColor"`
					IsActive         bool   `json:"isActive"`
					Logo             string `json:"logo"`
					Links            []struct {
						Rel        []string `json:"rel"`
						Href       string   `json:"href"`
						Text       string   `json:"text"`
						IsExternal bool     `json:"isExternal"`
						IsPremium  bool     `json:"isPremium"`
					} `json:"links"`
					Venue struct {
						ID string `json:"id"`
					} `json:"venue"`
				} `json:"team,omitempty"`
				Statistics []struct {
					Name         string `json:"name"`
					Abbreviation string `json:"abbreviation"`
					DisplayValue string `json:"displayValue"`
				} `json:"statistics"`
			} `json:"competitors"`
			Details []struct {
				Type struct {
					ID   string `json:"id"`
					Text string `json:"text"`
				} `json:"type"`
				Clock struct {
					Value        float64 `json:"value"`
					DisplayValue string  `json:"displayValue"`
				} `json:"clock"`
				Team struct {
					ID string `json:"id"`
				} `json:"team"`
				ScoreValue       int  `json:"scoreValue"`
				ScoringPlay      bool `json:"scoringPlay"`
				RedCard          bool `json:"redCard"`
				YellowCard       bool `json:"yellowCard"`
				PenaltyKick      bool `json:"penaltyKick"`
				OwnGoal          bool `json:"ownGoal"`
				Shootout         bool `json:"shootout"`
				AthletesInvolved []struct {
					ID          string `json:"id"`
					DisplayName string `json:"displayName"`
					ShortName   string `json:"shortName"`
					FullName    string `json:"fullName"`
					Jersey      string `json:"jersey"`
					Team        struct {
						ID string `json:"id"`
					} `json:"team"`
					Links []struct {
						Rel  []string `json:"rel"`
						Href string   `json:"href"`
					} `json:"links"`
					Position string `json:"position"`
				} `json:"athletesInvolved,omitempty"`
			} `json:"details"`
			Headlines []struct {
				Description   string `json:"description"`
				Type          string `json:"type"`
				ShortLinkText string `json:"shortLinkText"`
			} `json:"headlines"`
		} `json:"competitions"`
		Status struct {
			Clock        float64 `json:"clock"`
			DisplayClock string  `json:"displayClock"`
			Period       int     `json:"period"`
			Type         struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				State       string `json:"state"`
				Completed   bool   `json:"completed"`
				Description string `json:"description"`
				Detail      string `json:"detail"`
				ShortDetail string `json:"shortDetail"`
			} `json:"type"`
		} `json:"status"`
		Links []struct {
			Language   string   `json:"language"`
			Rel        []string `json:"rel"`
			Href       string   `json:"href"`
			Text       string   `json:"text"`
			ShortText  string   `json:"shortText"`
			IsExternal bool     `json:"isExternal"`
			IsPremium  bool     `json:"isPremium"`
		} `json:"links"`
	} `json:"events"`
}

type hockey struct {
	Leagues []struct {
		ID           string `json:"id"`
		UID          string `json:"uid"`
		Name         string `json:"name"`
		Abbreviation string `json:"abbreviation"`
		Slug         string `json:"slug"`
		Season       struct {
			Year      int    `json:"year"`
			StartDate string `json:"startDate"`
			EndDate   string `json:"endDate"`
			Type      struct {
				ID           string `json:"id"`
				Type         int    `json:"type"`
				Name         string `json:"name"`
				Abbreviation string `json:"abbreviation"`
			} `json:"type"`
		} `json:"season"`
		Logos []struct {
			Href        string   `json:"href"`
			Width       int      `json:"width"`
			Height      int      `json:"height"`
			Alt         string   `json:"alt"`
			Rel         []string `json:"rel"`
			LastUpdated string   `json:"lastUpdated"`
		} `json:"logos"`
		CalendarType        string   `json:"calendarType"`
		CalendarIsWhitelist bool     `json:"calendarIsWhitelist"`
		CalendarStartDate   string   `json:"calendarStartDate"`
		CalendarEndDate     string   `json:"calendarEndDate"`
		Calendar            []string `json:"calendar"`
	} `json:"leagues"`
	Season struct {
		Type int `json:"type"`
		Year int `json:"year"`
	} `json:"season"`
	Day struct {
		Date string `json:"date"`
	} `json:"day"`
	Events []struct {
		ID        string `json:"id"`
		UID       string `json:"uid"`
		Date      string `json:"date"`
		Name      string `json:"name"`
		ShortName string `json:"shortName"`
		Season    struct {
			Year int    `json:"year"`
			Type int    `json:"type"`
			Slug string `json:"slug"`
		} `json:"season"`
		Competitions []struct {
			ID         string `json:"id"`
			UID        string `json:"uid"`
			Date       string `json:"date"`
			Attendance int    `json:"attendance"`
			Type       struct {
				ID           string `json:"id"`
				Abbreviation string `json:"abbreviation"`
			} `json:"type"`
			TimeValid   bool `json:"timeValid"`
			NeutralSite bool `json:"neutralSite"`
			Recent      bool `json:"recent"`
			Venue       struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
				Address  struct {
					City    string `json:"city"`
					State   string `json:"state"`
					Country string `json:"country"`
				} `json:"address"`
				Capacity int  `json:"capacity"`
				Indoor   bool `json:"indoor"`
			} `json:"venue"`
			Competitors []struct {
				ID       string `json:"id"`
				UID      string `json:"uid"`
				Type     string `json:"type"`
				Order    int    `json:"order"`
				HomeAway string `json:"homeAway"`
				Winner   bool   `json:"winner"`
				Team     struct {
					ID               string `json:"id"`
					UID              string `json:"uid"`
					Location         string `json:"location"`
					Name             string `json:"name"`
					Abbreviation     string `json:"abbreviation"`
					DisplayName      string `json:"displayName"`
					ShortDisplayName string `json:"shortDisplayName"`
					Color            string `json:"color"`
					AlternateColor   string `json:"alternateColor"`
					IsActive         bool   `json:"isActive"`
					Venue            struct {
						ID string `json:"id"`
					} `json:"venue"`
					Links []struct {
						Rel        []string `json:"rel"`
						Href       string   `json:"href"`
						Text       string   `json:"text"`
						IsExternal bool     `json:"isExternal"`
						IsPremium  bool     `json:"isPremium"`
					} `json:"links"`
					Logo string `json:"logo"`
				} `json:"team"`
				Score      string `json:"score"`
				Linescores []struct {
					Value float64 `json:"value"`
				} `json:"linescores"`
				Statistics []struct {
					Name         string `json:"name"`
					Abbreviation string `json:"abbreviation"`
					DisplayValue string `json:"displayValue"`
				} `json:"statistics"`
				Leaders []struct {
					Name             string `json:"name"`
					DisplayName      string `json:"displayName"`
					ShortDisplayName string `json:"shortDisplayName"`
					Abbreviation     string `json:"abbreviation"`
					Leaders          []struct {
						DisplayValue string  `json:"displayValue"`
						Value        float64 `json:"value"`
						Athlete      struct {
							ID          string `json:"id"`
							FullName    string `json:"fullName"`
							DisplayName string `json:"displayName"`
							ShortName   string `json:"shortName"`
							Links       []struct {
								Rel  []string `json:"rel"`
								Href string   `json:"href"`
							} `json:"links"`
							Headshot string `json:"headshot"`
							Jersey   string `json:"jersey"`
							Position struct {
								Abbreviation string `json:"abbreviation"`
							} `json:"position"`
							Team struct {
								ID string `json:"id"`
							} `json:"team"`
							Active bool `json:"active"`
						} `json:"athlete"`
						Team struct {
							ID string `json:"id"`
						} `json:"team"`
					} `json:"leaders"`
				} `json:"leaders"`
				Probables []struct {
					Name             string `json:"name"`
					DisplayName      string `json:"displayName"`
					ShortDisplayName string `json:"shortDisplayName"`
					Abbreviation     string `json:"abbreviation"`
					PlayerID         int    `json:"playerId"`
					Athlete          struct {
						ID          string `json:"id"`
						FullName    string `json:"fullName"`
						DisplayName string `json:"displayName"`
						ShortName   string `json:"shortName"`
						Links       []struct {
							Rel  []string `json:"rel"`
							Href string   `json:"href"`
						} `json:"links"`
						Headshot string `json:"headshot"`
						Jersey   string `json:"jersey"`
						Position string `json:"position"`
						Team     struct {
							ID string `json:"id"`
						} `json:"team"`
					} `json:"athlete"`
					Status struct {
						ID           string `json:"id"`
						Name         string `json:"name"`
						Type         string `json:"type"`
						Abbreviation string `json:"abbreviation"`
					} `json:"status"`
					Statistics []interface{} `json:"statistics"`
				} `json:"probables"`
				Records []struct {
					Name         string `json:"name"`
					Abbreviation string `json:"abbreviation"`
					Type         string `json:"type"`
					Summary      string `json:"summary"`
				} `json:"records"`
			} `json:"competitors"`
			Notes  []interface{} `json:"notes"`
			Status struct {
				Clock        float64 `json:"clock"`
				DisplayClock string  `json:"displayClock"`
				Period       int     `json:"period"`
				Type         struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					State       string `json:"state"`
					Completed   bool   `json:"completed"`
					Description string `json:"description"`
					Detail      string `json:"detail"`
					ShortDetail string `json:"shortDetail"`
				} `json:"type"`
				FeaturedAthletes []struct {
					Name             string `json:"name"`
					DisplayName      string `json:"displayName"`
					ShortDisplayName string `json:"shortDisplayName"`
					Abbreviation     string `json:"abbreviation"`
					PlayerID         int    `json:"playerId"`
					Athlete          struct {
						ID          string `json:"id"`
						FullName    string `json:"fullName"`
						DisplayName string `json:"displayName"`
						ShortName   string `json:"shortName"`
						Links       []struct {
							Rel  []string `json:"rel"`
							Href string   `json:"href"`
						} `json:"links"`
						Headshot string `json:"headshot"`
						Jersey   string `json:"jersey"`
						Position string `json:"position"`
						Team     struct {
							ID string `json:"id"`
						} `json:"team"`
					} `json:"athlete"`
					Team struct {
						ID string `json:"id"`
					} `json:"team"`
					Statistics []struct {
						Name         string `json:"name"`
						Abbreviation string `json:"abbreviation"`
						DisplayValue string `json:"displayValue"`
					} `json:"statistics"`
				} `json:"featuredAthletes"`
			} `json:"status"`
			Broadcasts []struct {
				Market string   `json:"market"`
				Names  []string `json:"names"`
			} `json:"broadcasts"`
			Format struct {
				Regulation struct {
					Periods int `json:"periods"`
				} `json:"regulation"`
			} `json:"format"`
			StartDate     string `json:"startDate"`
			GeoBroadcasts []struct {
				Type struct {
					ID        string `json:"id"`
					ShortName string `json:"shortName"`
				} `json:"type"`
				Market struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"market"`
				Media struct {
					ShortName string `json:"shortName"`
				} `json:"media"`
				Lang   string `json:"lang"`
				Region string `json:"region"`
			} `json:"geoBroadcasts"`
			Headlines []struct {
				Description   string `json:"description"`
				Type          string `json:"type"`
				ShortLinkText string `json:"shortLinkText"`
				Video         []struct {
					ID        int    `json:"id"`
					Source    string `json:"source"`
					Headline  string `json:"headline"`
					Thumbnail string `json:"thumbnail"`
					Duration  int    `json:"duration"`
					Tracking  struct {
						SportName    string `json:"sportName"`
						LeagueName   string `json:"leagueName"`
						CoverageType string `json:"coverageType"`
						TrackingName string `json:"trackingName"`
						TrackingID   string `json:"trackingId"`
					} `json:"tracking"`
					DeviceRestrictions struct {
						Type    string   `json:"type"`
						Devices []string `json:"devices"`
					} `json:"deviceRestrictions"`
					GeoRestrictions struct {
						Type      string   `json:"type"`
						Countries []string `json:"countries"`
					} `json:"geoRestrictions"`
					Links struct {
						API struct {
							Self struct {
								Href string `json:"href"`
							} `json:"self"`
							Artwork struct {
								Href string `json:"href"`
							} `json:"artwork"`
						} `json:"api"`
						Web struct {
							Href  string `json:"href"`
							Short struct {
								Href string `json:"href"`
							} `json:"short"`
							Self struct {
								Href string `json:"href"`
							} `json:"self"`
						} `json:"web"`
						Source struct {
							Mezzanine struct {
								Href string `json:"href"`
							} `json:"mezzanine"`
							Flash struct {
								Href string `json:"href"`
							} `json:"flash"`
							Hds struct {
								Href string `json:"href"`
							} `json:"hds"`
							Hls struct {
								Href string `json:"href"`
								Hd   struct {
									Href string `json:"href"`
								} `json:"HD"`
							} `json:"HLS"`
							Hd struct {
								Href string `json:"href"`
							} `json:"HD"`
							Full struct {
								Href string `json:"href"`
							} `json:"full"`
							Href string `json:"href"`
						} `json:"source"`
						Mobile struct {
							Alert struct {
								Href string `json:"href"`
							} `json:"alert"`
							Source struct {
								Href string `json:"href"`
							} `json:"source"`
							Href      string `json:"href"`
							Streaming struct {
								Href string `json:"href"`
							} `json:"streaming"`
							ProgressiveDownload struct {
								Href string `json:"href"`
							} `json:"progressiveDownload"`
						} `json:"mobile"`
					} `json:"links"`
				} `json:"video"`
			} `json:"headlines"`
		} `json:"competitions"`
		Links []struct {
			Language   string   `json:"language"`
			Rel        []string `json:"rel"`
			Href       string   `json:"href"`
			Text       string   `json:"text"`
			ShortText  string   `json:"shortText"`
			IsExternal bool     `json:"isExternal"`
			IsPremium  bool     `json:"isPremium"`
		} `json:"links"`
		Status struct {
			Clock        float64 `json:"clock"`
			DisplayClock string  `json:"displayClock"`
			Period       int     `json:"period"`
			Type         struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				State       string `json:"state"`
				Completed   bool   `json:"completed"`
				Description string `json:"description"`
				Detail      string `json:"detail"`
				ShortDetail string `json:"shortDetail"`
			} `json:"type"`
		} `json:"status"`
	} `json:"events"`
}

type football struct {
	Leagues []struct {
		ID           string `json:"id"`
		UID          string `json:"uid"`
		Name         string `json:"name"`
		Abbreviation string `json:"abbreviation"`
		Slug         string `json:"slug"`
		Season       struct {
			Year      int    `json:"year"`
			StartDate string `json:"startDate"`
			EndDate   string `json:"endDate"`
			Type      struct {
				ID           string `json:"id"`
				Type         int    `json:"type"`
				Name         string `json:"name"`
				Abbreviation string `json:"abbreviation"`
			} `json:"type"`
		} `json:"season"`
		Logos []struct {
			Href        string   `json:"href"`
			Width       int      `json:"width"`
			Height      int      `json:"height"`
			Alt         string   `json:"alt"`
			Rel         []string `json:"rel"`
			LastUpdated string   `json:"lastUpdated"`
		} `json:"logos"`
		CalendarType        string `json:"calendarType"`
		CalendarIsWhitelist bool   `json:"calendarIsWhitelist"`
		CalendarStartDate   string `json:"calendarStartDate"`
		CalendarEndDate     string `json:"calendarEndDate"`
		Calendar            []struct {
			Label     string `json:"label"`
			Value     string `json:"value"`
			StartDate string `json:"startDate"`
			EndDate   string `json:"endDate"`
			Entries   []struct {
				Label          string `json:"label"`
				AlternateLabel string `json:"alternateLabel"`
				Detail         string `json:"detail"`
				Value          string `json:"value"`
				StartDate      string `json:"startDate"`
				EndDate        string `json:"endDate"`
			} `json:"entries"`
		} `json:"calendar"`
	} `json:"leagues"`
	Season struct {
		Type int `json:"type"`
		Year int `json:"year"`
	} `json:"season"`
	Week struct {
		Number     int `json:"number"`
		TeamsOnBye []struct {
			ID               string `json:"id"`
			UID              string `json:"uid"`
			Location         string `json:"location"`
			Name             string `json:"name"`
			Abbreviation     string `json:"abbreviation"`
			DisplayName      string `json:"displayName"`
			ShortDisplayName string `json:"shortDisplayName"`
			IsActive         bool   `json:"isActive"`
			Links            []struct {
				Rel        []string `json:"rel"`
				Href       string   `json:"href"`
				Text       string   `json:"text"`
				IsExternal bool     `json:"isExternal"`
				IsPremium  bool     `json:"isPremium"`
			} `json:"links"`
			Logo string `json:"logo"`
		} `json:"teamsOnBye"`
	} `json:"week"`
	Events []struct {
		ID        string `json:"id"`
		UID       string `json:"uid"`
		Date      string `json:"date"`
		Name      string `json:"name"`
		ShortName string `json:"shortName"`
		Season    struct {
			Year int    `json:"year"`
			Type int    `json:"type"`
			Slug string `json:"slug"`
		} `json:"season"`
		Week struct {
			Number int `json:"number"`
		} `json:"week"`
		Competitions []struct {
			ID         string `json:"id"`
			UID        string `json:"uid"`
			Date       string `json:"date"`
			Attendance int    `json:"attendance"`
			Type       struct {
				ID           string `json:"id"`
				Abbreviation string `json:"abbreviation"`
			} `json:"type"`
			TimeValid             bool `json:"timeValid"`
			NeutralSite           bool `json:"neutralSite"`
			ConferenceCompetition bool `json:"conferenceCompetition"`
			Recent                bool `json:"recent"`
			Venue                 struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
				Address  struct {
					City  string `json:"city"`
					State string `json:"state"`
				} `json:"address"`
				Capacity int  `json:"capacity"`
				Indoor   bool `json:"indoor"`
			} `json:"venue"`
			Competitors []struct {
				ID       string `json:"id"`
				UID      string `json:"uid"`
				Type     string `json:"type"`
				Order    int    `json:"order"`
				HomeAway string `json:"homeAway"`
				Winner   bool   `json:"winner"`
				Team     struct {
					ID               string `json:"id"`
					UID              string `json:"uid"`
					Location         string `json:"location"`
					Name             string `json:"name"`
					Abbreviation     string `json:"abbreviation"`
					DisplayName      string `json:"displayName"`
					ShortDisplayName string `json:"shortDisplayName"`
					Color            string `json:"color"`
					AlternateColor   string `json:"alternateColor"`
					IsActive         bool   `json:"isActive"`
					Venue            struct {
						ID string `json:"id"`
					} `json:"venue"`
					Links []struct {
						Rel        []string `json:"rel"`
						Href       string   `json:"href"`
						Text       string   `json:"text"`
						IsExternal bool     `json:"isExternal"`
						IsPremium  bool     `json:"isPremium"`
					} `json:"links"`
					Logo string `json:"logo"`
				} `json:"team"`
				Score      string `json:"score"`
				Linescores []struct {
					Value float64 `json:"value"`
				} `json:"linescores"`
				Statistics []interface{} `json:"statistics"`
				Records    []struct {
					Name         string `json:"name"`
					Abbreviation string `json:"abbreviation,omitempty"`
					Type         string `json:"type"`
					Summary      string `json:"summary"`
				} `json:"records"`
			} `json:"competitors"`
			Notes  []interface{} `json:"notes"`
			Status struct {
				Clock        float64 `json:"clock"`
				DisplayClock string  `json:"displayClock"`
				Period       int     `json:"period"`
				Type         struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					State       string `json:"state"`
					Completed   bool   `json:"completed"`
					Description string `json:"description"`
					Detail      string `json:"detail"`
					ShortDetail string `json:"shortDetail"`
				} `json:"type"`
			} `json:"status"`
			Broadcasts []struct {
				Market string   `json:"market"`
				Names  []string `json:"names"`
			} `json:"broadcasts"`
			Leaders []struct {
				Name             string `json:"name"`
				DisplayName      string `json:"displayName"`
				ShortDisplayName string `json:"shortDisplayName"`
				Abbreviation     string `json:"abbreviation"`
				Leaders          []struct {
					DisplayValue string  `json:"displayValue"`
					Value        float64 `json:"value"`
					Athlete      struct {
						ID          string `json:"id"`
						FullName    string `json:"fullName"`
						DisplayName string `json:"displayName"`
						ShortName   string `json:"shortName"`
						Links       []struct {
							Rel  []string `json:"rel"`
							Href string   `json:"href"`
						} `json:"links"`
						Headshot string `json:"headshot"`
						Jersey   string `json:"jersey"`
						Position struct {
							Abbreviation string `json:"abbreviation"`
						} `json:"position"`
						Team struct {
							ID string `json:"id"`
						} `json:"team"`
						Active bool `json:"active"`
					} `json:"athlete"`
					Team struct {
						ID string `json:"id"`
					} `json:"team"`
				} `json:"leaders"`
			} `json:"leaders"`
			Format struct {
				Regulation struct {
					Periods int `json:"periods"`
				} `json:"regulation"`
			} `json:"format"`
			StartDate     string `json:"startDate"`
			GeoBroadcasts []struct {
				Type struct {
					ID        string `json:"id"`
					ShortName string `json:"shortName"`
				} `json:"type"`
				Market struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"market"`
				Media struct {
					ShortName string `json:"shortName"`
				} `json:"media"`
				Lang   string `json:"lang"`
				Region string `json:"region"`
			} `json:"geoBroadcasts"`
			Headlines []struct {
				Description   string `json:"description"`
				Type          string `json:"type"`
				ShortLinkText string `json:"shortLinkText"`
			} `json:"headlines"`
		} `json:"competitions"`
		Links []struct {
			Language   string   `json:"language"`
			Rel        []string `json:"rel"`
			Href       string   `json:"href"`
			Text       string   `json:"text"`
			ShortText  string   `json:"shortText"`
			IsExternal bool     `json:"isExternal"`
			IsPremium  bool     `json:"isPremium"`
		} `json:"links"`
		Status struct {
			Clock        float64 `json:"clock"`
			DisplayClock string  `json:"displayClock"`
			Period       int     `json:"period"`
			Type         struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				State       string `json:"state"`
				Completed   bool   `json:"completed"`
				Description string `json:"description"`
				Detail      string `json:"detail"`
				ShortDetail string `json:"shortDetail"`
			} `json:"type"`
		} `json:"status"`
	} `json:"events"`
}

type basketball struct {
	Leagues []struct {
		ID           string `json:"id"`
		UID          string `json:"uid"`
		Name         string `json:"name"`
		Abbreviation string `json:"abbreviation"`
		Slug         string `json:"slug"`
		Season       struct {
			Year      int    `json:"year"`
			StartDate string `json:"startDate"`
			EndDate   string `json:"endDate"`
			Type      struct {
				ID           string `json:"id"`
				Type         int    `json:"type"`
				Name         string `json:"name"`
				Abbreviation string `json:"abbreviation"`
			} `json:"type"`
		} `json:"season"`
		Logos []struct {
			Href        string   `json:"href"`
			Width       int      `json:"width"`
			Height      int      `json:"height"`
			Alt         string   `json:"alt"`
			Rel         []string `json:"rel"`
			LastUpdated string   `json:"lastUpdated"`
		} `json:"logos"`
		CalendarType        string   `json:"calendarType"`
		CalendarIsWhitelist bool     `json:"calendarIsWhitelist"`
		CalendarStartDate   string   `json:"calendarStartDate"`
		CalendarEndDate     string   `json:"calendarEndDate"`
		Calendar            []string `json:"calendar"`
	} `json:"leagues"`
	Season struct {
		Type int `json:"type"`
		Year int `json:"year"`
	} `json:"season"`
	Day struct {
		Date string `json:"date"`
	} `json:"day"`
	Events []struct {
		ID        string `json:"id"`
		UID       string `json:"uid"`
		Date      string `json:"date"`
		Name      string `json:"name"`
		ShortName string `json:"shortName"`
		Season    struct {
			Year int    `json:"year"`
			Type int    `json:"type"`
			Slug string `json:"slug"`
		} `json:"season"`
		Competitions []struct {
			ID         string `json:"id"`
			UID        string `json:"uid"`
			Date       string `json:"date"`
			Attendance int    `json:"attendance"`
			Type       struct {
				ID           string `json:"id"`
				Abbreviation string `json:"abbreviation"`
			} `json:"type"`
			TimeValid             bool `json:"timeValid"`
			NeutralSite           bool `json:"neutralSite"`
			ConferenceCompetition bool `json:"conferenceCompetition"`
			Recent                bool `json:"recent"`
			Venue                 struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
				Address  struct {
					City  string `json:"city"`
					State string `json:"state"`
				} `json:"address"`
				Capacity int  `json:"capacity"`
				Indoor   bool `json:"indoor"`
			} `json:"venue"`
			Competitors []struct {
				ID       string `json:"id"`
				UID      string `json:"uid"`
				Type     string `json:"type"`
				Order    int    `json:"order"`
				HomeAway string `json:"homeAway"`
				Team     struct {
					ID               string `json:"id"`
					UID              string `json:"uid"`
					Location         string `json:"location"`
					Name             string `json:"name"`
					Abbreviation     string `json:"abbreviation"`
					DisplayName      string `json:"displayName"`
					ShortDisplayName string `json:"shortDisplayName"`
					Color            string `json:"color"`
					AlternateColor   string `json:"alternateColor"`
					IsActive         bool   `json:"isActive"`
					Venue            struct {
						ID string `json:"id"`
					} `json:"venue"`
					Links []struct {
						Rel        []string `json:"rel"`
						Href       string   `json:"href"`
						Text       string   `json:"text"`
						IsExternal bool     `json:"isExternal"`
						IsPremium  bool     `json:"isPremium"`
					} `json:"links"`
					Logo string `json:"logo"`
				} `json:"team"`
				Score      string `json:"score"`
				Statistics []struct {
					Name             string `json:"name"`
					Abbreviation     string `json:"abbreviation"`
					DisplayValue     string `json:"displayValue"`
					RankDisplayValue string `json:"rankDisplayValue,omitempty"`
				} `json:"statistics"`
				Records []struct {
					Name         string `json:"name"`
					Abbreviation string `json:"abbreviation,omitempty"`
					Type         string `json:"type"`
					Summary      string `json:"summary"`
				} `json:"records"`
				Leaders []struct {
					Name             string `json:"name"`
					DisplayName      string `json:"displayName"`
					ShortDisplayName string `json:"shortDisplayName"`
					Abbreviation     string `json:"abbreviation"`
					Leaders          []struct {
						DisplayValue string  `json:"displayValue"`
						Value        float64 `json:"value"`
						Athlete      struct {
							ID          string `json:"id"`
							FullName    string `json:"fullName"`
							DisplayName string `json:"displayName"`
							ShortName   string `json:"shortName"`
							Links       []struct {
								Rel  []string `json:"rel"`
								Href string   `json:"href"`
							} `json:"links"`
							Headshot string `json:"headshot"`
							Jersey   string `json:"jersey"`
							Position struct {
								Abbreviation string `json:"abbreviation"`
							} `json:"position"`
							Team struct {
								ID string `json:"id"`
							} `json:"team"`
							Active bool `json:"active"`
						} `json:"athlete"`
						Team struct {
							ID string `json:"id"`
						} `json:"team"`
					} `json:"leaders"`
				} `json:"leaders"`
			} `json:"competitors"`
			Notes  []interface{} `json:"notes"`
			Status struct {
				Clock        float64 `json:"clock"`
				DisplayClock string  `json:"displayClock"`
				Period       int     `json:"period"`
				Type         struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					State       string `json:"state"`
					Completed   bool   `json:"completed"`
					Description string `json:"description"`
					Detail      string `json:"detail"`
					ShortDetail string `json:"shortDetail"`
				} `json:"type"`
			} `json:"status"`
			Broadcasts []struct {
				Market string   `json:"market"`
				Names  []string `json:"names"`
			} `json:"broadcasts"`
			Format struct {
				Regulation struct {
					Periods int `json:"periods"`
				} `json:"regulation"`
			} `json:"format"`
			Tickets []struct {
				Summary         string `json:"summary"`
				NumberAvailable int    `json:"numberAvailable"`
				Links           []struct {
					Href string `json:"href"`
				} `json:"links"`
			} `json:"tickets"`
			StartDate     string `json:"startDate"`
			GeoBroadcasts []struct {
				Type struct {
					ID        string `json:"id"`
					ShortName string `json:"shortName"`
				} `json:"type"`
				Market struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"market"`
				Media struct {
					ShortName string `json:"shortName"`
				} `json:"media"`
				Lang   string `json:"lang"`
				Region string `json:"region"`
			} `json:"geoBroadcasts"`
			Odds []struct {
				Provider struct {
					ID       string `json:"id"`
					Name     string `json:"name"`
					Priority int    `json:"priority"`
				} `json:"provider"`
				Details   string  `json:"details"`
				OverUnder float64 `json:"overUnder"`
			} `json:"odds"`
		} `json:"competitions"`
		Links []struct {
			Language   string   `json:"language"`
			Rel        []string `json:"rel"`
			Href       string   `json:"href"`
			Text       string   `json:"text"`
			ShortText  string   `json:"shortText"`
			IsExternal bool     `json:"isExternal"`
			IsPremium  bool     `json:"isPremium"`
		} `json:"links"`
		Status struct {
			Clock        float64 `json:"clock"`
			DisplayClock string  `json:"displayClock"`
			Period       int     `json:"period"`
			Type         struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				State       string `json:"state"`
				Completed   bool   `json:"completed"`
				Description string `json:"description"`
				Detail      string `json:"detail"`
				ShortDetail string `json:"shortDetail"`
			} `json:"type"`
		} `json:"status"`
	} `json:"events"`
}

type baseball struct {
	Leagues []struct {
		ID           string `json:"id"`
		UID          string `json:"uid"`
		Name         string `json:"name"`
		Abbreviation string `json:"abbreviation"`
		MidsizeName  string `json:"midsizeName"`
		Slug         string `json:"slug"`
		Season       struct {
			Year      int    `json:"year"`
			StartDate string `json:"startDate"`
			EndDate   string `json:"endDate"`
			Type      struct {
				ID           string `json:"id"`
				Type         int    `json:"type"`
				Name         string `json:"name"`
				Abbreviation string `json:"abbreviation"`
			} `json:"type"`
		} `json:"season"`
		Logos []struct {
			Href        string   `json:"href"`
			Width       int      `json:"width"`
			Height      int      `json:"height"`
			Alt         string   `json:"alt"`
			Rel         []string `json:"rel"`
			LastUpdated string   `json:"lastUpdated"`
		} `json:"logos"`
		CalendarType        string   `json:"calendarType"`
		CalendarIsWhitelist bool     `json:"calendarIsWhitelist"`
		CalendarStartDate   string   `json:"calendarStartDate"`
		CalendarEndDate     string   `json:"calendarEndDate"`
		Calendar            []string `json:"calendar"`
	} `json:"leagues"`
	Events []struct {
		ID        string `json:"id"`
		UID       string `json:"uid"`
		Date      string `json:"date"`
		Name      string `json:"name"`
		ShortName string `json:"shortName"`
		Season    struct {
			Year int    `json:"year"`
			Type int    `json:"type"`
			Slug string `json:"slug"`
		} `json:"season"`
		Competitions []struct {
			ID         string `json:"id"`
			UID        string `json:"uid"`
			Date       string `json:"date"`
			Attendance int    `json:"attendance"`
			Type       struct {
				ID           string `json:"id"`
				Abbreviation string `json:"abbreviation"`
			} `json:"type"`
			TimeValid             bool `json:"timeValid"`
			NeutralSite           bool `json:"neutralSite"`
			ConferenceCompetition bool `json:"conferenceCompetition"`
			Recent                bool `json:"recent"`
			WasSuspended          bool `json:"wasSuspended"`
			Venue                 struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
				Address  struct {
					City  string `json:"city"`
					State string `json:"state"`
				} `json:"address"`
				Capacity int  `json:"capacity"`
				Indoor   bool `json:"indoor"`
			} `json:"venue"`
			Competitors []struct {
				ID       string `json:"id"`
				UID      string `json:"uid"`
				Type     string `json:"type"`
				Order    int    `json:"order"`
				HomeAway string `json:"homeAway"`
				Winner   bool   `json:"winner"`
				Team     struct {
					ID               string `json:"id"`
					UID              string `json:"uid"`
					Location         string `json:"location"`
					Name             string `json:"name"`
					Abbreviation     string `json:"abbreviation"`
					DisplayName      string `json:"displayName"`
					ShortDisplayName string `json:"shortDisplayName"`
					Color            string `json:"color"`
					AlternateColor   string `json:"alternateColor"`
					IsActive         bool   `json:"isActive"`
					Links            []struct {
						Rel        []string `json:"rel"`
						Href       string   `json:"href"`
						Text       string   `json:"text"`
						IsExternal bool     `json:"isExternal"`
						IsPremium  bool     `json:"isPremium"`
					} `json:"links"`
					Logo string `json:"logo"`
				} `json:"team"`
				Score      string `json:"score"`
				Linescores []struct {
					Value float64 `json:"value"`
				} `json:"linescores"`
				Statistics []struct {
					Name         string `json:"name"`
					Abbreviation string `json:"abbreviation"`
					DisplayValue string `json:"displayValue"`
				} `json:"statistics"`
				Leaders []struct {
					Name             string `json:"name"`
					DisplayName      string `json:"displayName"`
					ShortDisplayName string `json:"shortDisplayName"`
					Abbreviation     string `json:"abbreviation"`
					Leaders          []struct {
						DisplayValue string  `json:"displayValue"`
						Value        float64 `json:"value"`
						Athlete      struct {
							ID          string `json:"id"`
							FullName    string `json:"fullName"`
							DisplayName string `json:"displayName"`
							ShortName   string `json:"shortName"`
							Links       []struct {
								Rel  []string `json:"rel"`
								Href string   `json:"href"`
							} `json:"links"`
							Headshot string `json:"headshot"`
							Jersey   string `json:"jersey"`
							Position struct {
								Abbreviation string `json:"abbreviation"`
							} `json:"position"`
							Team struct {
								ID string `json:"id"`
							} `json:"team"`
							Active bool `json:"active"`
						} `json:"athlete"`
						Team struct {
							ID string `json:"id"`
						} `json:"team"`
					} `json:"leaders"`
				} `json:"leaders"`
				Probables []struct {
					Name             string `json:"name"`
					DisplayName      string `json:"displayName"`
					ShortDisplayName string `json:"shortDisplayName"`
					Abbreviation     string `json:"abbreviation"`
					PlayerID         int    `json:"playerId"`
					Athlete          struct {
						ID          string `json:"id"`
						FullName    string `json:"fullName"`
						DisplayName string `json:"displayName"`
						ShortName   string `json:"shortName"`
						Links       []struct {
							Rel  []string `json:"rel"`
							Href string   `json:"href"`
						} `json:"links"`
						Headshot string `json:"headshot"`
						Jersey   string `json:"jersey"`
						Position string `json:"position"`
						Team     struct {
							ID string `json:"id"`
						} `json:"team"`
					} `json:"athlete"`
					Statistics []struct {
						Name         string `json:"name"`
						Abbreviation string `json:"abbreviation"`
						DisplayValue string `json:"displayValue"`
					} `json:"statistics"`
				} `json:"probables"`
				Hits    int `json:"hits"`
				Errors  int `json:"errors"`
				Records []struct {
					Name         string `json:"name"`
					Abbreviation string `json:"abbreviation,omitempty"`
					Type         string `json:"type"`
					Summary      string `json:"summary"`
				} `json:"records"`
			} `json:"competitors"`
			Notes  []any `json:"notes"`
			Status struct {
				Clock        float64 `json:"clock"`
				DisplayClock string  `json:"displayClock"`
				Period       int     `json:"period"`
				Type         struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					State       string `json:"state"`
					Completed   bool   `json:"completed"`
					Description string `json:"description"`
					Detail      string `json:"detail"`
					ShortDetail string `json:"shortDetail"`
				} `json:"type"`
				FeaturedAthletes []struct {
					Name             string `json:"name"`
					DisplayName      string `json:"displayName"`
					ShortDisplayName string `json:"shortDisplayName"`
					Abbreviation     string `json:"abbreviation"`
					PlayerID         int    `json:"playerId"`
					Athlete          struct {
						ID          string `json:"id"`
						FullName    string `json:"fullName"`
						DisplayName string `json:"displayName"`
						ShortName   string `json:"shortName"`
						Links       []struct {
							Rel  []string `json:"rel"`
							Href string   `json:"href"`
						} `json:"links"`
						Headshot string `json:"headshot"`
						Jersey   string `json:"jersey"`
						Position string `json:"position"`
						Team     struct {
							ID string `json:"id"`
						} `json:"team"`
					} `json:"athlete"`
					Team struct {
						ID string `json:"id"`
					} `json:"team"`
					Statistics []struct {
						Name         string `json:"name"`
						Abbreviation string `json:"abbreviation"`
						DisplayValue string `json:"displayValue"`
					} `json:"statistics"`
				} `json:"featuredAthletes"`
			} `json:"status"`
			Broadcasts []struct {
				Market string   `json:"market"`
				Names  []string `json:"names"`
			} `json:"broadcasts"`
			Leaders []struct {
				Name             string `json:"name"`
				DisplayName      string `json:"displayName"`
				ShortDisplayName string `json:"shortDisplayName"`
				Abbreviation     string `json:"abbreviation"`
				Leaders          []struct {
					DisplayValue string  `json:"displayValue"`
					Value        float64 `json:"value"`
					Athlete      struct {
						ID          string `json:"id"`
						FullName    string `json:"fullName"`
						DisplayName string `json:"displayName"`
						ShortName   string `json:"shortName"`
						Links       []struct {
							Rel  []string `json:"rel"`
							Href string   `json:"href"`
						} `json:"links"`
						Headshot string `json:"headshot"`
						Jersey   string `json:"jersey"`
						Position struct {
							Abbreviation string `json:"abbreviation"`
						} `json:"position"`
						Team struct {
							ID string `json:"id"`
						} `json:"team"`
						Active bool `json:"active"`
					} `json:"athlete"`
					Team struct {
						ID string `json:"id"`
					} `json:"team"`
				} `json:"leaders"`
			} `json:"leaders"`
			Format struct {
				Regulation struct {
					Periods int `json:"periods"`
				} `json:"regulation"`
			} `json:"format"`
			StartDate     string `json:"startDate"`
			GeoBroadcasts []struct {
				Type struct {
					ID        string `json:"id"`
					ShortName string `json:"shortName"`
				} `json:"type"`
				Market struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"market"`
				Media struct {
					ShortName string `json:"shortName"`
				} `json:"media"`
				Lang   string `json:"lang"`
				Region string `json:"region"`
			} `json:"geoBroadcasts"`
			Headlines []struct {
				Description   string `json:"description"`
				Type          string `json:"type"`
				ShortLinkText string `json:"shortLinkText"`
				Video         []struct {
					ID        int    `json:"id"`
					Source    string `json:"source"`
					Headline  string `json:"headline"`
					Thumbnail string `json:"thumbnail"`
					Duration  int    `json:"duration"`
					Tracking  struct {
						SportName    string `json:"sportName"`
						LeagueName   string `json:"leagueName"`
						CoverageType string `json:"coverageType"`
						TrackingName string `json:"trackingName"`
						TrackingID   string `json:"trackingId"`
					} `json:"tracking"`
					DeviceRestrictions struct {
						Type    string   `json:"type"`
						Devices []string `json:"devices"`
					} `json:"deviceRestrictions"`
					GeoRestrictions struct {
						Type      string   `json:"type"`
						Countries []string `json:"countries"`
					} `json:"geoRestrictions"`
					Links struct {
						API struct {
							Self struct {
								Href string `json:"href"`
							} `json:"self"`
							Artwork struct {
								Href string `json:"href"`
							} `json:"artwork"`
						} `json:"api"`
						Web struct {
							Href  string `json:"href"`
							Short struct {
								Href string `json:"href"`
							} `json:"short"`
							Self struct {
								Href string `json:"href"`
							} `json:"self"`
						} `json:"web"`
						Source struct {
							Mezzanine struct {
								Href string `json:"href"`
							} `json:"mezzanine"`
							Flash struct {
								Href string `json:"href"`
							} `json:"flash"`
							Hds struct {
								Href string `json:"href"`
							} `json:"hds"`
							Hls struct {
								Href string `json:"href"`
								Hd   struct {
									Href string `json:"href"`
								} `json:"HD"`
							} `json:"HLS"`
							Hd struct {
								Href string `json:"href"`
							} `json:"HD"`
							Full struct {
								Href string `json:"href"`
							} `json:"full"`
							Href string `json:"href"`
						} `json:"source"`
						Mobile struct {
							Alert struct {
								Href string `json:"href"`
							} `json:"alert"`
							Source struct {
								Href string `json:"href"`
							} `json:"source"`
							Href      string `json:"href"`
							Streaming struct {
								Href string `json:"href"`
							} `json:"streaming"`
							ProgressiveDownload struct {
								Href string `json:"href"`
							} `json:"progressiveDownload"`
						} `json:"mobile"`
					} `json:"links"`
				} `json:"video"`
			} `json:"headlines"`
		} `json:"competitions"`
		Links []struct {
			Language   string   `json:"language"`
			Rel        []string `json:"rel"`
			Href       string   `json:"href"`
			Text       string   `json:"text"`
			ShortText  string   `json:"shortText"`
			IsExternal bool     `json:"isExternal"`
			IsPremium  bool     `json:"isPremium"`
		} `json:"links"`
		Status struct {
			Clock        float64 `json:"clock"`
			DisplayClock string  `json:"displayClock"`
			Period       int     `json:"period"`
			Type         struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				State       string `json:"state"`
				Completed   bool   `json:"completed"`
				Description string `json:"description"`
				Detail      string `json:"detail"`
				ShortDetail string `json:"shortDetail"`
			} `json:"type"`
		} `json:"status"`
		Weather struct {
			DisplayValue    string `json:"displayValue"`
			Temperature     int    `json:"temperature"`
			HighTemperature int    `json:"highTemperature"`
			ConditionID     string `json:"conditionId"`
			Link            struct {
				Language   string   `json:"language"`
				Rel        []string `json:"rel"`
				Href       string   `json:"href"`
				Text       string   `json:"text"`
				ShortText  string   `json:"shortText"`
				IsExternal bool     `json:"isExternal"`
				IsPremium  bool     `json:"isPremium"`
			} `json:"link"`
		} `json:"weather,omitempty"`
	} `json:"events"`
}

type mma struct {
	Leagues []struct {
		ID           string `json:"id"`
		UID          string `json:"uid"`
		Name         string `json:"name"`
		Abbreviation string `json:"abbreviation"`
		Slug         string `json:"slug"`
		Season       struct {
			Year      int    `json:"year"`
			StartDate string `json:"startDate"`
			EndDate   string `json:"endDate"`
			Type      struct {
				ID           string `json:"id"`
				Type         int    `json:"type"`
				Name         string `json:"name"`
				Abbreviation string `json:"abbreviation"`
			} `json:"type"`
		} `json:"season"`
		Logos []struct {
			Href        string   `json:"href"`
			Width       int      `json:"width"`
			Height      int      `json:"height"`
			Alt         string   `json:"alt"`
			Rel         []string `json:"rel"`
			LastUpdated string   `json:"lastUpdated"`
		} `json:"logos"`
		CalendarType        string `json:"calendarType"`
		CalendarIsWhitelist bool   `json:"calendarIsWhitelist"`
		CalendarStartDate   string `json:"calendarStartDate"`
		CalendarEndDate     string `json:"calendarEndDate"`
		Calendar            []struct {
			Label     string `json:"label"`
			StartDate string `json:"startDate"`
			EndDate   string `json:"endDate"`
			Event     struct {
				Ref string `json:"$ref"`
			} `json:"event"`
		} `json:"calendar"`
	} `json:"leagues"`
	Season struct {
		Type int `json:"type"`
		Year int `json:"year"`
	} `json:"season"`
	Day struct {
		Date string `json:"date"`
	} `json:"day"`
	Events []struct {
		ID        string `json:"id"`
		UID       string `json:"uid"`
		Date      string `json:"date"`
		Name      string `json:"name"`
		ShortName string `json:"shortName"`
		Season    struct {
			Year int    `json:"year"`
			Type int    `json:"type"`
			Slug string `json:"slug"`
		} `json:"season"`
		Competitions []struct {
			ID          string `json:"id"`
			UID         string `json:"uid"`
			Date        string `json:"date"`
			EndDate     string `json:"endDate"`
			TimeValid   bool   `json:"timeValid"`
			NeutralSite bool   `json:"neutralSite"`
			Recent      bool   `json:"recent"`
			Venue       struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
				Address  struct {
					City  string `json:"city"`
					State string `json:"state"`
				} `json:"address"`
				Indoor bool `json:"indoor"`
			} `json:"venue"`
			Competitors []struct {
				ID      string `json:"id"`
				UID     string `json:"uid"`
				Type    string `json:"type"`
				Order   int    `json:"order"`
				Winner  bool   `json:"winner"`
				Athlete struct {
					FullName    string `json:"fullName"`
					DisplayName string `json:"displayName"`
					ShortName   string `json:"shortName"`
					Flag        struct {
						Href string   `json:"href"`
						Alt  string   `json:"alt"`
						Rel  []string `json:"rel"`
					} `json:"flag"`
				} `json:"athlete"`
			} `json:"competitors"`
			Status struct {
				Clock        float64 `json:"clock"`
				DisplayClock string  `json:"displayClock"`
				Period       int     `json:"period"`
				Type         struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					State       string `json:"state"`
					Completed   bool   `json:"completed"`
					Description string `json:"description"`
					Detail      string `json:"detail"`
					ShortDetail string `json:"shortDetail"`
				} `json:"type"`
			} `json:"status"`
			Broadcasts []struct {
				Market string   `json:"market"`
				Names  []string `json:"names"`
			} `json:"broadcasts"`
			Format struct {
				Regulation struct {
					Periods int `json:"periods"`
				} `json:"regulation"`
			} `json:"format"`
			StartDate     string `json:"startDate"`
			GeoBroadcasts []struct {
				Type struct {
					ID        string `json:"id"`
					ShortName string `json:"shortName"`
				} `json:"type"`
				Market struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"market"`
				Media struct {
					ShortName string `json:"shortName"`
				} `json:"media"`
				Lang   string `json:"lang"`
				Region string `json:"region"`
			} `json:"geoBroadcasts"`
			Type struct {
				ID           string `json:"id"`
				Abbreviation string `json:"abbreviation"`
			} `json:"type,omitempty"`
		} `json:"competitions"`
		Links []struct {
			Language   string   `json:"language"`
			Rel        []string `json:"rel"`
			Href       string   `json:"href"`
			Text       string   `json:"text"`
			ShortText  string   `json:"shortText"`
			IsExternal bool     `json:"isExternal"`
			IsPremium  bool     `json:"isPremium"`
		} `json:"links"`
		Venues []struct {
			ID       string `json:"id"`
			FullName string `json:"fullName"`
			Address  struct {
				City  string `json:"city"`
				State string `json:"state"`
			} `json:"address"`
		} `json:"venues"`
		Status struct {
			Type struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				State       string `json:"state"`
				Completed   bool   `json:"completed"`
				Description string `json:"description"`
				Detail      string `json:"detail"`
				ShortDetail string `json:"shortDetail"`
			} `json:"type"`
		} `json:"status"`
	} `json:"events"`
}
