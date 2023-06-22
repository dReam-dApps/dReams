package holdero

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	dreams "github.com/SixofClubsss/dReams"
	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"
	dero "github.com/deroproject/derohe/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var Faces dreams.AssetSelect
var Backs dreams.AssetSelect

// Holdero card face selection object
//   - Sets shared face url on selected
//   - If deck is not present locally, it is downloaded
func FaceSelect() fyne.Widget {
	options := []string{"Light", "Dark"}
	Faces.Select = widget.NewSelect(options, func(s string) {
		switch Faces.Select.SelectedIndex() {
		case -1:
			Faces.Name = "light/"
		case 0:
			Faces.Name = "light/"
		case 1:
			Faces.Name = "dark/"
		default:
			Faces.Name = s
		}

		check := strings.Trim(s, "0123456789")
		if check == "AZYPC" {
			Faces.URL = "https://raw.githubusercontent.com/Azylem/" + s + "/main/" + s + ".zip?raw=true"
			dir := dreams.GetDir()
			face := dir + "/cards/" + Faces.Name + "/card1.png"
			if !dreams.FileExists(face, "dReams") {
				log.Println("[dReams] Downloading " + Faces.URL)
				go GetZipDeck(Faces.Name, Faces.URL)
			}
		} else if check == "SIXPC" {
			Faces.URL = "https://raw.githubusercontent.com/SixofClubsss/" + s + "/main/" + s + ".zip?raw=true"
			dir := dreams.GetDir()
			face := dir + "/cards/" + Faces.Name + "/card1.png"
			if !dreams.FileExists(face, "dReams") {
				log.Println("[dReams] Downloading " + Faces.URL)
				go GetZipDeck(Faces.Name, Faces.URL)
			}
		} else if check == "High-Strangeness" {
			Faces.URL = "https://raw.githubusercontent.com/High-Strangeness/High-Strangeness/main/HS_Deck/HS_Deck.zip?raw=true"
			dir := dreams.GetDir()
			face := dir + "/cards/" + Faces.Name + "/card1.png"
			if !dreams.FileExists(face, "dReams") {
				log.Println("[dReams] Downloading " + Faces.URL)
				go GetZipDeck(Faces.Name, Faces.URL)
			}
		} else {
			Faces.URL = ""
		}
	})

	Faces.Select.SetSelectedIndex(0)
	Faces.Select.PlaceHolder = "Faces"

	return Faces.Select
}

// Holdero card back selection object for all games
//   - Sets shared back url on selected
//   - If back is not present locally, it is downloaded
func BackSelect() fyne.Widget {
	options := []string{"Light", "Dark"}
	Backs.Select = widget.NewSelect(options, func(s string) {
		switch Backs.Select.SelectedIndex() {
		case -1:
			Backs.Name = "back1.png"
		case 0:
			Backs.Name = "back1.png"
		case 1:
			Backs.Name = "back2.png"
		default:
			Backs.Name = s
		}

		go func() {
			check := strings.Trim(s, "0123456789")
			if check == "AZYPCB" {
				Backs.URL = "https://raw.githubusercontent.com/Azylem/" + s + "/main/" + s + ".png"
				dir := dreams.GetDir()
				file := dir + "/cards/backs/" + s + ".png"
				if !dreams.FileExists(file, "dReams") {
					log.Println("[dReams] Downloading " + Backs.URL)
					downloadFileLocal("cards/backs/"+Backs.Name+".png", Backs.URL)
				}
			} else if check == "SIXPCB" {
				Backs.URL = "https://raw.githubusercontent.com/SixofClubsss/" + s + "/main/" + s + ".png"
				dir := dreams.GetDir()
				back := dir + "/cards/backs/" + s + ".png"
				if !dreams.FileExists(back, "dReams") {
					log.Println("[dReams] Downloading " + Backs.URL)
					downloadFileLocal("cards/backs/"+Backs.Name+".png", Backs.URL)
				}
			} else if check == "High-Strangeness" {
				Backs.URL = "https://raw.githubusercontent.com/High-Strangeness/" + s + "/main/HS_Back/HS_Backs.png"
				dir := dreams.GetDir()
				back := dir + "/cards/backs/" + s + ".png"
				if !dreams.FileExists(back, "dReams") {
					log.Println("[dReams] Downloading " + Backs.URL)
					downloadFileLocal("cards/backs/"+Backs.Name+".png", Backs.URL)
				}
			} else {
				Backs.URL = ""
			}
		}()
	})

	Backs.Select.SetSelectedIndex(0)
	Backs.Select.PlaceHolder = "Backs"

	return Backs.Select
}

// dReams app avatar selection object
//   - Sets shared avatar url on selected
func AvatarSelect(asset_map map[string]string) fyne.Widget {
	options := []string{"None"}
	Settings.AvatarSelect = widget.NewSelect(options, func(s string) {
		switch Settings.AvatarSelect.SelectedIndex() {
		case -1:
			Settings.Avatar = "None"
		case 0:
			Settings.Avatar = "None"
		default:
			Settings.Avatar = s
		}

		asset_info := ValidAsset(asset_map[s])
		check := strings.Trim(s, " #0123456789")
		if check == "DBC" {
			Settings.AvatarUrl = "https://raw.githubusercontent.com/Azylem/" + s + "/main/" + s + ".PNG"
		} else if check == "HighStrangeness" {
			Settings.AvatarUrl = "https://raw.githubusercontent.com/High-Strangeness/High-Strangeness/main/" + s + "/" + s + ".jpg"
		} else if check == "AZYDS" {
			Settings.AvatarUrl = "https://raw.githubusercontent.com/Azylem/" + s + "/main/" + s + "-IC.png"
		} else if check == "SIXART" {
			Settings.AvatarUrl = "https://raw.githubusercontent.com/SixofClubsss/SIXART/main/" + s + "/" + s + "-IC.png"
		} else if check == "Dero Seals" {
			seal := strings.Trim(s, "Dero Sals#")
			Settings.AvatarUrl = "https://ipfs.io/ipfs/QmP3HnzWpiaBA6ZE8c3dy5ExeG7hnYjSqkNfVbeVW5iEp6/low/" + seal + ".jpg"
		} else if asset_info {
			agent := getAgentNumber(asset_map[s])
			if agent >= 0 && agent < 172 {
				Settings.AvatarUrl = "https://ipfs.io/ipfs/QmaRHXcQwbFdUAvwbjgpDtr5kwGiNpkCM2eDBzAbvhD7wh/low/" + strconv.Itoa(agent) + ".jpg"
			} else if agent < 1200 {
				Settings.AvatarUrl = "https://ipfs.io/ipfs/QmQQyKoE9qDnzybeDCXhyMhwQcPmLaVy3AyYAzzC2zMauW/low/" + strconv.Itoa(agent) + ".jpg"
			}
		} else if s == "None" {
			Settings.AvatarUrl = ""
		}
	})

	Settings.AvatarSelect.PlaceHolder = "Avatar"

	return Settings.AvatarSelect
}

// Confirm if asset map is valid
func ValidAsset(s string) bool {
	if s != "" && len(s) == 64 {
		return true
	}
	return false
}

// Rpc call to get A-Team agent number
func getAgentNumber(scid string) int {
	if rpc.Daemon.IsConnected() {
		rpcClientD, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		params := dero.GetSC_Params{
			SCID:      scid,
			Code:      false,
			Variables: true,
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println("[getAgentNumber]", err)
			return 1200
		}

		data := result.VariableStringKeys["metadata"]
		var agent dreams.Agent

		hx, _ := hex.DecodeString(data.(string))
		if err := json.Unmarshal(hx, &agent); err == nil {
			return agent.ID
		}

	}
	return 1200
}

// Holdero shared cards toggle object
//   - Do not send a blank url
//   - If cards are not present locally, it is downloaded
func SharedDecks() fyne.Widget {
	options := []string{"Shared Decks"}
	Settings.SharedOn = widget.NewRadioGroup(options, func(string) {
		if Settings.Shared || ((len(Round.Face) < 3 || len(Round.Back) < 3) && Round.ID != 1) {
			log.Println("[Holdero] Shared Decks Off")
			Settings.Shared = false
			Faces.Select.Enable()
			Backs.Select.Enable()
		} else {
			log.Println("[Holdero] Shared Decks On")
			Settings.Shared = true
			if Round.ID == 1 {
				if Faces.Name != "" && Faces.URL != "" && Backs.Name != "" && Backs.URL != "" {
					SharedDeckUrl(Faces.Name, Faces.URL, Backs.Name, Backs.URL)
					dir := dreams.GetDir()
					back := "/cards/backs/" + Backs.Name + ".png"
					face := "/cards/" + Faces.Name + "/card1.png"

					if !dreams.FileExists(dir+face, "dReams") {
						go GetZipDeck(Faces.Name, Faces.URL)
					}

					if !dreams.FileExists(dir+back, "dReams") {
						downloadFileLocal("cards/backs/"+Backs.Name+".png", Backs.URL)
					}
				}
			} else {
				Faces.Select.Disable()
				Backs.Select.Disable()
				dir := dreams.GetDir()
				back := "/cards/backs/" + Round.Back + ".png"
				face := "/cards/" + Round.Face + "/card1.png"

				if !dreams.FileExists(dir+face, "dReams") {
					go GetZipDeck(Round.Face, Round.F_url)
				}

				if !dreams.FileExists(dir+back, "dReams") {
					downloadFileLocal("cards/backs/"+Round.Back+".png", Round.B_url)
				}
			}
		}
	})

	Settings.SharedOn.Disable()

	return Settings.SharedOn
}

// Tournament deposit button
//   - Pass main window obj and tabs to reset to
func TournamentButton(obj []fyne.CanvasObject, tabs *container.AppTabs) fyne.CanvasObject {
	Table.Tournament = widget.NewButton("Tournament", func() {
		obj[1] = tourneyConfirm(obj, tabs)
		obj[1].Refresh()
	})

	Table.Tournament.Hide()

	return Table.Tournament
}

// Confirmation for dReams-Dero swap pairs
//   - c defines swap for Dero or dReams
//   - amt of Dero in atomic units
//   - Pass main window obj to reset to
func DreamsConfirm(c, amt float64, obj *fyne.Container, reset fyne.CanvasObject) fyne.CanvasObject {
	var text string
	dero := (amt / 100000) / 333
	ratio := math.Pow(10, float64(5))
	x := math.Round(dero*ratio) / ratio
	a := fmt.Sprint(strconv.FormatFloat(dero, 'f', 5, 64))
	switch c {
	case 1:
		text = fmt.Sprintf("You are about to swap %s DERO for %.5f dReams", a, amt/100000)
	case 2:
		text = fmt.Sprintf("You are about to swap %.5f dReams for %s Dero", amt/100000, a)
	}

	done := false
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	confirm_button := widget.NewButton("Confirm", func() {
		switch c {
		case 1:
			rpc.GetdReams(uint64(x * 100000))
		case 2:
			rpc.TradedReams(uint64(amt))
		default:

		}

		done = true
		obj.Objects[0] = reset
		obj.Objects[0].Refresh()
	})

	cancel_button := widget.NewButton("Cancel", func() {
		done = true
		obj.Objects[0] = reset
		obj.Objects[0].Refresh()
	})

	buttons := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewVBox(layout.NewSpacer(), label, layout.NewSpacer(), buttons)

	go func() {
		for rpc.IsReady() {
			time.Sleep(time.Second)
			if done {
				return
			}
		}

		obj.Objects[0] = reset
		obj.Objects[0].Refresh()
	}()

	return container.NewMax(content)
}

// Holdero tournament chip deposit confirmation
//   - Pass main window obj and tabs to reset to
func tourneyConfirm(obj []fyne.CanvasObject, tabs *container.AppTabs) fyne.CanvasObject {
	bal := rpc.TokenBalance(TourneySCID)
	balance := float64(bal) / 100000
	a := fmt.Sprint(strconv.FormatFloat(balance, 'f', 2, 64))
	text := fmt.Sprintf("You are about to deposit %s Tournament Chips into leader board contract\n\nConfirm", a)

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	confirm_button := widget.NewButton("Confirm", func() {
		TourneyDeposit(bal, menu.Username)
		obj[1] = tabs
		obj[1].Refresh()

	})

	cancel_button := widget.NewButton("Cancel", func() {
		obj[1] = tabs
		obj[1].Refresh()

	})

	buttons := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewVBox(layout.NewSpacer(), label, layout.NewSpacer(), buttons)

	return container.NewMax(content)
}
