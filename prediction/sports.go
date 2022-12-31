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
	"github.com/SixofClubsss/dReams/table"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type sportsItems struct {
	Contract      string
	Info          *widget.Label
	Sports_list   *widget.List
	Favorite_list *widget.List
	Owned_list    *widget.List
}

var SportsControl sportsItems

func SportsConnectedBox() fyne.Widget {
	menu.MenuControl.Sports_check = widget.NewCheck("", func(b bool) {
		if !b {
			table.Actions.Game_select.Hide()
			table.Actions.Multi.Hide()
			table.Actions.ButtonA.Hide()
			table.Actions.ButtonB.Hide()
		}
	})
	menu.MenuControl.Sports_check.Disable()

	return menu.MenuControl.Sports_check
}

func SportsContractEntry() fyne.Widget {
	options := []string{""}
	table.Actions.S_contract = widget.NewSelectEntry(options)
	table.Actions.S_contract.PlaceHolder = "Contract Address: "
	table.Actions.S_contract.OnCursorChanged = func() {
		if rpc.Signal.Daemon {
			go func() {
				if len(SportsControl.Contract) == 64 {
					yes, _ := rpc.ValidBetContract(SportsControl.Contract)
					if yes {
						menu.MenuControl.Sports_check.SetChecked(true)
						if !menu.CheckActiveGames(SportsControl.Contract) {
							table.Actions.Game_select.Show()
						} else {
							table.Actions.Game_select.Hide()
						}
					} else {
						menu.MenuControl.Sports_check.SetChecked(false)
					}
				} else {
					menu.MenuControl.Sports_check.SetChecked(false)
				}
			}()
		}
	}

	this := binding.BindString(&SportsControl.Contract)
	table.Actions.S_contract.Bind(this)

	return table.Actions.S_contract
}

func SportsBox() fyne.CanvasObject {
	table.Actions.Game_select = widget.NewSelect(table.Actions.Game_options, func(s string) {
		split := strings.Split(s, "   ")
		a, b := menu.GetSportsTeams(SportsControl.Contract, split[0])
		if table.Actions.Game_select.SelectedIndex() >= 0 {
			table.Actions.Multi.Show()
			table.Actions.ButtonA.Show()
			table.Actions.ButtonB.Show()
			table.Actions.ButtonA.Text = a
			table.Actions.ButtonA.Refresh()
			table.Actions.ButtonB.Text = b
			table.Actions.ButtonB.Refresh()
		} else {
			table.Actions.Multi.Hide()
			table.Actions.ButtonA.Hide()
			table.Actions.ButtonB.Hide()
		}
	})

	table.Actions.Game_select.PlaceHolder = "Select Game #"
	table.Actions.Game_select.Hide()

	var Multi_options = []string{"1x", "3x", "5x"}
	table.Actions.Multi = widget.NewRadioGroup(Multi_options, func(s string) {

	})
	table.Actions.Multi.Horizontal = true
	table.Actions.Multi.Hide()

	table.Actions.ButtonA = widget.NewButton("TEAM A", func() {
		if len(SportsControl.Contract) == 64 {
			confirmPopUp(3, table.Actions.ButtonA.Text, table.Actions.ButtonB.Text)
		}
	})
	table.Actions.ButtonA.Hide()

	table.Actions.ButtonB = widget.NewButton("TEAM B", func() {
		if len(SportsControl.Contract) == 64 {
			confirmPopUp(4, table.Actions.ButtonA.Text, table.Actions.ButtonB.Text)
		}
	})
	table.Actions.ButtonB.Hide()

	sports_muli := container.NewCenter(table.Actions.Multi)
	sports_actions := container.NewVBox(
		sports_muli,
		table.Actions.Game_select,
		table.Actions.ButtonA,
		table.Actions.ButtonB)

	table.Actions.Sports_box = sports_actions
	table.Actions.Sports_box.Hide()

	return table.Actions.Sports_box
}

func setSportsControls(str string) (item string) {
	table.Actions.Game_select.ClearSelected()
	table.Actions.Game_select.Options = []string{}
	table.Actions.Game_select.Refresh()
	split := strings.Split(str, "   ")
	if len(split) >= 3 {
		trimmed := strings.Trim(split[2], " ")
		table.Actions.Sports_box.Show()
		if len(trimmed) == 64 {
			go SetSportsInfo(trimmed)
			item = str
			table.Actions.S_contract.SetText(trimmed)
		}
	}

	return
}

func SetSportsInfo(scid string) {
	info := GetBook(menu.Gnomes.Init, scid)
	SportsControl.Info.SetText(info)
	SportsControl.Info.Refresh()
	table.Actions.Game_select.Refresh()
}

func SportsListings() fyne.CanvasObject { /// sports contract list
	SportsControl.Sports_list = widget.NewList(
		func() int {
			return len(menu.MenuControl.Sports_contracts)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(menu.MenuControl.Sports_contracts[i])
		})

	var item string

	SportsControl.Sports_list.OnSelected = func(id widget.ListItemID) {
		if id != 0 && menu.Connected() {
			item = setSportsControls(menu.MenuControl.Sports_contracts[id])
			SportsControl.Favorite_list.UnselectAll()
			SportsControl.Owned_list.UnselectAll()
		} else {
			table.Actions.Sports_box.Hide()
		}
	}

	save := widget.NewButton("Favorite", func() {
		menu.MenuControl.Sports_favorites = append(menu.MenuControl.Sports_favorites, item)
		sort.Strings(menu.MenuControl.Sports_favorites)
	})

	cont := container.NewBorder(
		nil,
		container.NewBorder(nil, nil, nil, save, layout.NewSpacer()),
		nil,
		nil,
		SportsControl.Sports_list)

	return cont
}

func SportsFavorites() fyne.CanvasObject {
	SportsControl.Favorite_list = widget.NewList(
		func() int {
			return len(menu.MenuControl.Sports_favorites)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(menu.MenuControl.Sports_favorites[i])
		})

	var item string

	SportsControl.Favorite_list.OnSelected = func(id widget.ListItemID) {
		if menu.Connected() {
			item = setSportsControls(menu.MenuControl.Sports_favorites[id])
			SportsControl.Sports_list.UnselectAll()
			SportsControl.Owned_list.UnselectAll()
		} else {
			table.Actions.Sports_box.Hide()
		}
	}

	remove := widget.NewButton("Remove", func() {
		if len(menu.MenuControl.Sports_favorites) > 0 {
			SportsControl.Favorite_list.UnselectAll()
			new := menu.MenuControl.Sports_favorites
			for i := range new {
				if new[i] == item {
					copy(new[i:], new[i+1:])
					new[len(new)-1] = ""
					new = new[:len(new)-1]
					menu.MenuControl.Sports_favorites = new
					break
				}
			}
		}
		SportsControl.Favorite_list.Refresh()
		sort.Strings(menu.MenuControl.Sports_favorites)
	})

	cont := container.NewBorder(
		nil,
		container.NewBorder(nil, nil, nil, remove, layout.NewSpacer()),
		nil,
		nil,
		SportsControl.Favorite_list)

	return cont
}

func SportsOwned() fyne.CanvasObject {
	SportsControl.Owned_list = widget.NewList(
		func() int {
			return len(menu.MenuControl.Sports_owned)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(menu.MenuControl.Sports_owned[i])
		})

	SportsControl.Owned_list.OnSelected = func(id widget.ListItemID) {
		if menu.Connected() {
			setSportsControls(menu.MenuControl.Sports_owned[id])
			SportsControl.Sports_list.UnselectAll()
			SportsControl.Favorite_list.UnselectAll()
		} else {
			table.Actions.Sports_box.Hide()
		}
	}

	return SportsControl.Owned_list
}

func GetBook(gi bool, scid string) (info string) {
	if gi && !menu.GnomonClosing() && menu.Gnomes.Sync {
		_, initValue := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_init", menu.Gnomes.Indexer.ChainHeight, true)
		if initValue != nil {
			_, playedValue := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_played", menu.Gnomes.Indexer.ChainHeight, true)
			//_, hl := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "hl", menu.Gnomes.Indexer.ChainHeight, true)
			init := initValue[0]
			played := playedValue[0]

			table.Actions.Game_options = []string{}
			table.Actions.Game_select.Options = table.Actions.Game_options
			played_str := strconv.Itoa(int(played))
			if init == played {
				info = "SCID: \n" + scid + "\n\nGames Completed: " + played_str + "\n\nNo current Games\n"
				return
			}

			var single bool
			iv := 1
			for {
				_, s_init := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_init_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
				if s_init != nil {
					game, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "game_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					league, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "league_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, s_n := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_#_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, s_amt := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_amount_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, s_end := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_end_at_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, s_total := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_total_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					//s_urlValue, _ := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "s_url_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, s_ta := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "team_a_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, s_tb := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "team_b_"+strconv.Itoa(iv), menu.Gnomes.Indexer.ChainHeight, true)
					_, time_a := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "time_a", menu.Gnomes.Indexer.ChainHeight, true)
					_, time_b := menu.Gnomes.Indexer.Backend.GetSCIDValuesByKey(scid, "time_b", menu.Gnomes.Indexer.ChainHeight, true)

					team_a := menu.TrimTeamA(game[0])
					team_b := menu.TrimTeamB(game[0])

					if s_end[0] > uint64(time.Now().Unix()) {
						current := table.Actions.Game_select.Options
						new := append(current, strconv.Itoa(iv)+"   "+game[0])
						table.Actions.Game_select.Options = new
					}

					eA := fmt.Sprint(s_end[0] * 1000)
					min := fmt.Sprint(float64(s_amt[0]) / 100000)
					n := strconv.Itoa(int(s_n[0]))
					aV := strconv.Itoa(int(s_ta[0]))
					bV := strconv.Itoa(int(s_tb[0]))
					t := strconv.Itoa(int(s_total[0]))
					if !single {
						single = true
						info = "SCID: \n" + scid + "\n\nGames Completed: " + played_str + "\nCurrent Games:\n"
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

func S_Results(g, gN, l, min, eA, c, tA, tB, tAV, tBV, total string, a, b uint64) (info string) { /// sports info label
	result, err := strconv.ParseFloat(total, 32)

	if err != nil {
		log.Println("Float Conversion Error", err)
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

func sports(league string) (api string) {
	switch league {
	case "NHL":
		api = "http://site.api.espn.com/apis/site/v2/sports/hockey/nhl/scoreboard"
		// case "FIFA":
		// 	api = "http://site.api.espn.com/apis/site/v2/sports/soccer/fifa.world/scoreboard"
	case "EPL":
		api = "http://site.api.espn.com/apis/site/v2/sports/soccer/eng.1/scoreboard"
	case "NFL":
		api = "http://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard"
	case "NBA":
		api = "http://site.api.espn.com/apis/site/v2/sports/basketball/nba/scoreboard"
	case "UFC":
		api = "http://site.api.espn.com/apis/site/v2/sports/mma/ufc/scoreboard"
	case "Bellator":
		api = "http://site.api.espn.com/apis/site/v2/sports/mma/bellator/scoreboard"
	default:
		api = ""
	}

	return api
}

func GetCurrentWeek(league string) {
	for i := 0; i < 8; i++ {
		now := time.Now().AddDate(0, 0, i)
		date := time.Unix(now.Unix(), 0).String()
		date = date[0:10]
		comp := date[0:4] + date[5:7] + date[8:10]
		switch league {
		case "EPL":
			GetSoccer(comp, league)
		case "NBA":
			GetBasketball(comp, league)
		case "NFL":
			GetFootball(comp, league)
		case "NHL":
			GetHockey(comp, league)
		default:

		}
	}
}

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

func callSoccer(date, league string) (s *soccer) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", sports(league)+"?dates="+date, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Println(err.Error())
		return
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println(err.Error())
		return
	}

	json.Unmarshal(b, &s)

	return s
}

func callMma(date, league string) (m *mma) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", sports(league)+"?dates="+date, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Println(err.Error())
		return
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println(err.Error())
		return
	}

	json.Unmarshal(b, &m)

	return m
}

func callBasketball(date, league string) (bb *basketball) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", sports(league)+"?dates="+date, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Println(err.Error())
		return
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println(err.Error())
		return
	}

	json.Unmarshal(b, &bb)

	return bb
}

func callFootball(date, league string) (f *football) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", sports(league)+"?dates="+date, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Println(err.Error())
		return
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println(err.Error())
		return
	}

	json.Unmarshal(b, &f)

	return f
}

func callHockey(date, league string) (h *hockey) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", sports(league)+"?dates="+date, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Println(err.Error())
		return
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println(err.Error())
		return
	}

	json.Unmarshal(b, &h)

	return h
}

func GetGameEnd(date, game, league string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", sports(league)+"?dates="+date, nil)

	if err != nil {
		log.Println(err.Error())
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Println(err.Error())
		return
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println(err.Error())
		return
	}

	if league == "UFC" || league == "Bellator" {
		var found mma
		json.Unmarshal(b, &found)
		for i := range found.Events {
			trimmed := strings.Trim(found.Events[i].Competitions[0].StartDate, "Z")
			utc_time, err := time.Parse("2006-01-02T15:04", trimmed)
			if err != nil {
				log.Println(err)
			}

			a := found.Events[i].Competitions[0].Competitors[0].Athlete.DisplayName
			b := found.Events[i].Competitions[0].Competitors[1].Athlete.DisplayName
			g := a + "--" + b

			if g == game {
				PS_Control.S_end.SetText(strconv.Itoa(int(utc_time.Unix())))
			}

		}
	} else {
		var found scores
		json.Unmarshal(b, &found)
		for i := range found.Events {
			trimmed := strings.Trim(found.Events[i].Competitions[0].StartDate, "Z")
			utc_time, err := time.Parse("2006-01-02T15:04", trimmed)
			if err != nil {
				log.Println(err)
			}

			a := found.Events[i].Competitions[0].Competitors[0].Team.Abbreviation
			b := found.Events[i].Competitions[0].Competitors[1].Team.Abbreviation
			g := a + "--" + b
			if g == game {
				PS_Control.S_end.SetText(strconv.Itoa(int(utc_time.Unix())))
			}
		}
	}
}

func callScores(date, league string) (s *scores) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", sports(league)+"?dates="+date, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Println(err.Error())
		return
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println(err.Error())
		return
	}

	json.Unmarshal(b, &s)

	return s
}

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
					log.Println(err)
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
					log.Println(err)
				}

				tz, _ := time.LoadLocation("Local")
				local := utc_time.In(tz).String()
				state := found.Events[i].Competitions[0].Status.Type.State
				team_a := found.Events[i].Competitions[0].Competitors[0].Athlete.DisplayName
				team_b := found.Events[i].Competitions[0].Competitors[1].Athlete.DisplayName
				winner_a := found.Events[i].Competitions[0].Competitors[0].Winner
				winner_b := found.Events[i].Competitions[0].Competitors[1].Winner
				period := found.Events[i].Competitions[0].Status.Period
				clock := found.Events[i].Competitions[0].Status.DisplayClock
				complete := found.Events[i].Status.Type.Completed

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
	label.Refresh()
}

func GetHockey(date, league string) {
	found := callHockey(date, league)
	if found != nil {
		for i := range found.Events {
			pregame := found.Events[i].Competitions[0].Status.Type.State
			trimmed := strings.Trim(found.Events[i].Competitions[0].StartDate, "Z")
			utc_time, err := time.Parse("2006-01-02T15:04", trimmed)
			if err != nil {
				log.Println(err)
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

func GetSoccer(date, league string) {
	found := callSoccer(date, league)
	if found != nil {
		for i := range found.Events {
			pregame := found.Events[i].Competitions[0].Status.Type.State

			trimmed := strings.Trim(found.Events[i].Competitions[0].StartDate, "Z")
			utc_time, err := time.Parse("2006-01-02T15:04", trimmed)
			if err != nil {
				log.Println(err)
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

func GetWinner(game, league string) (string, string) {
	for i := -2; i < 1; i++ {
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

						teamB := found.Events[i].Competitions[0].Competitors[1].Team.Abbreviation
						b_win := found.Events[i].Competitions[0].Competitors[1].Winner

						if a_win && !b_win {
							return "team_a", teamA
						} else if b_win && !a_win {
							return "team_b", teamB
						} else {
							return "", ""
						}
					}
				}
			}
		}
	}
	return "", ""
}

func GetMmaWinner(game, league string) (string, string) {
	for i := -2; i < 1; i++ {
		day := time.Now().AddDate(0, 0, i)
		date := time.Unix(day.Unix(), 0).String()
		date = date[0:10]
		comp := date[0:4] + date[5:7] + date[8:10]

		found := callMma(comp, league)
		if found != nil {
			for i := range found.Events {
				a := found.Events[i].Competitions[0].Competitors[0].Athlete.DisplayName
				b := found.Events[i].Competitions[0].Competitors[1].Athlete.DisplayName
				g := a + "--" + b

				if g == game {
					if found.Events[i].Status.Type.Completed {
						teamA := found.Events[i].Competitions[0].Competitors[0].Athlete.DisplayName
						a_win := found.Events[i].Competitions[0].Competitors[0].Winner

						teamB := found.Events[i].Competitions[0].Competitors[1].Athlete.DisplayName
						b_win := found.Events[i].Competitions[0].Competitors[1].Winner

						if a_win && !b_win {
							return "team_a", teamA
						} else if b_win && !a_win {
							return "team_b", teamB
						} else {
							return "", ""
						}
					}
				}
			}
		}
	}
	return "", ""
}

func GetFootball(date, league string) {
	found := callFootball(date, league)
	if found != nil {
		for i := range found.Events {
			pregame := found.Events[i].Competitions[0].Status.Type.State
			trimmed := strings.Trim(found.Events[i].Competitions[0].StartDate, "Z")
			utc_time, err := time.Parse("2006-01-02T15:04", trimmed)
			if err != nil {
				log.Println(err)
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

func GetBasketball(date, league string) {
	found := callBasketball(date, league)
	if found != nil {
		for i := range found.Events {
			pregame := found.Events[i].Competitions[0].Status.Type.State
			trimmed := strings.Trim(found.Events[i].Competitions[0].StartDate, "Z")
			utc_time, err := time.Parse("2006-01-02T15:04", trimmed)
			if err != nil {
				log.Println(err)
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

func GetMma(date, league string) {
	found := callMma(date, league)
	if found != nil {
		for i := range found.Events {
			pregame := found.Events[i].Competitions[0].Status.Type.State
			trimmed := strings.Trim(found.Events[i].Competitions[0].StartDate, "Z")
			utc_time, err := time.Parse("2006-01-02T15:04", trimmed)
			if err != nil {
				log.Println(err)
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
