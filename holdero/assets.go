package holdero

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/rpc"
	dero "github.com/deroproject/derohe/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// dReams card face selection object for all games
//   - Sets shared face url on selected
//   - If deck is not present locally, it is downloaded
func FaceSelect() fyne.Widget {
	options := []string{"Light", "Dark"}
	Settings.FaceSelect = widget.NewSelect(options, func(s string) {
		switch Settings.FaceSelect.SelectedIndex() {
		case -1:
			Settings.Faces = "light/"
		case 0:
			Settings.Faces = "light/"
		case 1:
			Settings.Faces = "dark/"
		default:
			Settings.Faces = s
		}

		check := strings.Trim(s, "0123456789")
		if check == "AZYPC" {
			Settings.FaceUrl = "https://raw.githubusercontent.com/Azylem/" + s + "/main/" + s + ".zip?raw=true"
			dir := GetDir()
			face := dir + "/cards/" + Settings.Faces + "/card1.png"
			if !FileExists(face, "dReams") {
				log.Println("[dReams] Downloading " + Settings.FaceUrl)
				go GetZipDeck(Settings.Faces, Settings.FaceUrl)
			}
		} else if check == "SIXPC" {
			Settings.FaceUrl = "https://raw.githubusercontent.com/SixofClubsss/" + s + "/main/" + s + ".zip?raw=true"
			dir := GetDir()
			face := dir + "/cards/" + Settings.Faces + "/card1.png"
			if !FileExists(face, "dReams") {
				log.Println("[dReams] Downloading " + Settings.FaceUrl)
				go GetZipDeck(Settings.Faces, Settings.FaceUrl)
			}
		} else if check == "High-Strangeness" {
			Settings.FaceUrl = "https://raw.githubusercontent.com/High-Strangeness/High-Strangeness/main/HS_Deck/HS_Deck.zip?raw=true"
			dir := GetDir()
			face := dir + "/cards/" + Settings.Faces + "/card1.png"
			if !FileExists(face, "dReams") {
				log.Println("[dReams] Downloading " + Settings.FaceUrl)
				go GetZipDeck(Settings.Faces, Settings.FaceUrl)
			}
		} else {
			Settings.FaceUrl = ""
		}
	})

	Settings.FaceSelect.SetSelectedIndex(0)
	Settings.FaceSelect.PlaceHolder = "Faces"

	return Settings.FaceSelect
}

// dReams card back selection object for all games
//   - Sets shared back url on selected
//   - If back is not present locally, it is downloaded
func BackSelect() fyne.Widget {
	options := []string{"Light", "Dark"}
	Settings.BackSelect = widget.NewSelect(options, func(s string) {
		switch Settings.BackSelect.SelectedIndex() {
		case -1:
			Settings.Backs = "back1.png"
		case 0:
			Settings.Backs = "back1.png"
		case 1:
			Settings.Backs = "back2.png"
		default:
			Settings.Backs = s
		}

		go func() {
			check := strings.Trim(s, "0123456789")
			if check == "AZYPCB" {
				Settings.BackUrl = "https://raw.githubusercontent.com/Azylem/" + s + "/main/" + s + ".png"
				dir := GetDir()
				file := dir + "/cards/backs/" + s + ".png"
				if !FileExists(file, "dReams") {
					log.Println("[dReams] Downloading " + Settings.BackUrl)
					downloadFileLocal("cards/backs/"+Settings.Backs+".png", Settings.BackUrl)
				}
			} else if check == "SIXPCB" {
				Settings.BackUrl = "https://raw.githubusercontent.com/SixofClubsss/" + s + "/main/" + s + ".png"
				dir := GetDir()
				back := dir + "/cards/backs/" + s + ".png"
				if !FileExists(back, "dReams") {
					log.Println("[dReams] Downloading " + Settings.BackUrl)
					downloadFileLocal("cards/backs/"+Settings.Backs+".png", Settings.BackUrl)
				}
			} else if check == "High-Strangeness" {
				Settings.BackUrl = "https://raw.githubusercontent.com/High-Strangeness/" + s + "/main/HS_Back/HS_Back.png"
				dir := GetDir()
				back := dir + "/cards/backs/" + s + ".png"
				if !FileExists(back, "dReams") {
					log.Println("[dReams] Downloading " + Settings.BackUrl)
					downloadFileLocal("cards/backs/"+Settings.Backs+".png", Settings.BackUrl)
				}
			} else {
				Settings.BackUrl = ""
			}
		}()
	})

	Settings.BackSelect.SetSelectedIndex(0)
	Settings.BackSelect.PlaceHolder = "Backs"

	return Settings.BackSelect
}

// dReams app theme selection object
//   - If image is not present locally, it is downloaded
func ThemeSelect() fyne.Widget {
	options := []string{"Main"}
	Settings.ThemeSelect = widget.NewSelect(options, func(s string) {
		switch Settings.ThemeSelect.SelectedIndex() {
		case -1:
			Settings.Theme = "Main"
		case 0:
			Settings.Theme = "Main"
		default:
			Settings.Theme = s
		}
		go func() {
			check := strings.Trim(s, "0123456789")
			if check == "AZYDS" {
				dir := GetDir()
				file := dir + "/cards/" + s + "/" + s + ".png"
				if FileExists(file, "dReams") {
					Settings.ThemeImg = *canvas.NewImageFromFile(file)
				} else {
					Settings.ThemeUrl = "https://raw.githubusercontent.com/Azylem/" + s + "/main/" + s + ".png"
					log.Println("[dReams] Downloading", Settings.ThemeUrl)
					Settings.ThemeImg, _ = DownloadFile(Settings.ThemeUrl, s)
				}
			} else if check == "SIXART" {
				dir := GetDir()
				file := dir + "/cards/" + s + "/" + s + ".png"
				if FileExists(file, "dReams") {
					Settings.ThemeImg = *canvas.NewImageFromFile(file)
				} else {
					Settings.ThemeUrl = "https://raw.githubusercontent.com/SixofClubsss/" + s + "/main/" + s + ".png"
					log.Println("[dReams] Downloading", Settings.ThemeUrl)
					Settings.ThemeImg, _ = DownloadFile(Settings.ThemeUrl, s)
				}
			} else if check == "HSTheme" {
				dir := GetDir()
				file := dir + "/cards/" + s + "/" + s + ".png"
				if FileExists(file, "dReams") {
					Settings.ThemeImg = *canvas.NewImageFromFile(file)
				} else {
					Settings.ThemeUrl = "https://raw.githubusercontent.com/High-Strangeness/High-Strangeness/main/" + s + "/" + s + ".png"
					log.Println("[dReams] Downloading", Settings.ThemeUrl)
					Settings.ThemeImg, _ = DownloadFile(Settings.ThemeUrl, s)
				}
			}

			if s == "Main" {
				Settings.ThemeImg = *canvas.NewImageFromResource(bundle.ResourceBackgroundPng)
			}
		}()
	})
	Settings.ThemeSelect.PlaceHolder = "Theme"

	return Settings.ThemeSelect
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
			Settings.AvatarUrl = "https://raw.githubusercontent.com/SixofClubsss/" + s + "/main/" + s + "-IC.png"
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
	if rpc.Daemon.Connect {
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
		var agent Agent

		hx, _ := hex.DecodeString(data.(string))
		if err := json.Unmarshal(hx, &agent); err == nil {
			return agent.ID
		}

	}
	return 1200
}

// Check if path to file exists
//   - tag for log print
func FileExists(path, tag string) bool {
	if _, err := os.Stat(path); err == nil {
		return true

	} else if errors.Is(err, os.ErrNotExist) {
		log.Printf("[%s] %s Not Found\n", tag, path)

		return false
	}

	return false
}

// Holdero shared cards toggle object
//   - Do not send a blank url
//   - If cards are not present locally, it is downloaded
func SharedDecks() fyne.Widget {
	options := []string{"Shared Decks"}
	Settings.SharedOn = widget.NewRadioGroup(options, func(string) {
		if Settings.Shared || ((len(rpc.Round.Face) < 3 || len(rpc.Round.Back) < 3) && rpc.Round.ID != 1) {
			log.Println("[Holdero] Shared Decks Off")
			Settings.Shared = false
			Settings.FaceSelect.Enable()
			Settings.BackSelect.Enable()
		} else {
			log.Println("[Holdero] Shared Decks On")
			Settings.Shared = true
			if rpc.Round.ID == 1 {
				if Settings.Faces != "" && Settings.FaceUrl != "" && Settings.Backs != "" && Settings.BackUrl != "" {
					rpc.SharedDeckUrl(Settings.Faces, Settings.FaceUrl, Settings.Backs, Settings.BackUrl)
					dir := GetDir()
					back := "/cards/backs/" + Settings.Backs + ".png"
					face := "/cards/" + Settings.Faces + "/card1.png"

					if !FileExists(dir+face, "dReams") {
						go GetZipDeck(Settings.Faces, Settings.FaceUrl)
					}

					if !FileExists(dir+back, "dReams") {
						downloadFileLocal("cards/backs/"+Settings.Backs+".png", Settings.BackUrl)
					}
				}
			} else {
				Settings.FaceSelect.Disable()
				Settings.BackSelect.Disable()
				dir := GetDir()
				back := "/cards/backs/" + rpc.Round.Back + ".png"
				face := "/cards/" + rpc.Round.Face + "/card1.png"

				if !FileExists(dir+face, "dReams") {
					go GetZipDeck(rpc.Round.Face, rpc.Round.F_url)
				}

				if !FileExists(dir+back, "dReams") {
					downloadFileLocal("cards/backs/"+rpc.Round.Back+".png", rpc.Round.B_url)
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

// Confirmation for dReams-Dero swap
//   - c defines swap for Dero or dReams
//   - amt of Dero in atomic units
//   - Pass main window obj to reset to
func DreamsConfirm(c, amt int, obj *container.Split, reset fyne.CanvasObject) fyne.CanvasObject {
	var text string
	dero := float64(amt) / 333
	ratio := math.Pow(10, float64(5))
	x := math.Round(dero*ratio) / ratio
	a := fmt.Sprint(strconv.FormatFloat(dero, 'f', 5, 64))
	switch c {
	case 1:
		text = fmt.Sprintf("You are about to swap %s DERO for %d dReams\n\nConfirm", a, amt)
	case 2:
		text = fmt.Sprintf("You are about to swap %d dReams for %s Dero\n\nConfirm", amt, a)
	}

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	confirm_button := widget.NewButton("Confirm", func() {
		switch c {
		case 1:
			rpc.GetdReams(uint64(x * 100000))
		case 2:
			rpc.TradedReams(uint64(amt * 100000))
		default:

		}

		obj.Trailing.(*fyne.Container).Objects[1] = reset
		obj.Trailing.(*fyne.Container).Objects[1].Refresh()
	})

	cancel_button := widget.NewButton("Cancel", func() {
		obj.Trailing.(*fyne.Container).Objects[1] = reset
		obj.Trailing.(*fyne.Container).Objects[1].Refresh()
	})

	buttons := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewVBox(layout.NewSpacer(), label, layout.NewSpacer(), buttons)

	return container.NewMax(content)
}

// Holdero tournament chip deposit confirmation
//   - Pass main window obj and tabs to reset to
func tourneyConfirm(obj []fyne.CanvasObject, tabs *container.AppTabs) fyne.CanvasObject {
	bal := rpc.TokenBalance(rpc.TourneySCID)
	balance := float64(bal) / 100000
	a := fmt.Sprint(strconv.FormatFloat(balance, 'f', 2, 64))
	text := fmt.Sprintf("You are about to deposit %s Tournament Chips into leaderboard contract\n\nConfirm", a)

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	confirm_button := widget.NewButton("Confirm", func() {
		rpc.TourneyDeposit(bal, Poker_name)
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
