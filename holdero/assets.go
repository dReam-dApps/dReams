package holdero

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/SixofClubsss/dReams/dwidget"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/tarot"
	dero "github.com/deroproject/derohe/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type resources struct {
	SmallIcon  fyne.Resource
	Back2      fyne.Resource
	Back3      fyne.Resource
	Back4      fyne.Resource
	Background fyne.Resource
}

type assetObject struct {
	Dreams_bal    *canvas.Text
	Dero_bal      *canvas.Text
	Dero_price    *canvas.Text
	Wall_height   *canvas.Text
	Daem_height   *canvas.Text
	Gnomes_height *canvas.Text
	Gnomes_sync   *canvas.Text
	Gnomes_index  *canvas.Text
	Index_entry   *widget.Entry
	Index_button  *widget.Button
	Index_search  *widget.Button
	Asset_list    *widget.List
	Assets        []string
	Asset_map     map[string]string
	Name          *canvas.Text
	Collection    *canvas.Text
	Descrption    *canvas.Text
	Icon          canvas.Image
	Stats_box     fyne.Container
	Header_box    fyne.Container
}

var Assets assetObject
var Resource resources

// Get resources from main bundled
func GetTableResources(r1, r2, r3, r4, r5, r6, r7, r8 fyne.Resource) {
	Resource.SmallIcon = r1
	Resource.Back2 = r2
	Resource.Back3 = r3
	Resource.Background = r4
	Resource.Back4 = r5
	tarot.Iluma.Background1 = r6
	tarot.Iluma.Background2 = r7
	tarot.Iluma.Back = r8

}

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
				Settings.ThemeImg = *canvas.NewImageFromResource(Resource.Background)
			}
		}()
	})
	Settings.ThemeSelect.PlaceHolder = "Theme"

	return Settings.ThemeSelect
}

// dReams app avatar selection object
//   - Sets shared avatar url on selected
func AvatarSelect() fyne.Widget {
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
		} else if ValidAgent(s) {
			agent := getAgentNumber(Assets.Asset_map[s])
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

// Confirm valid A-Team agent number
func ValidAgent(s string) bool {
	if Assets.Asset_map[s] != "" && len(Assets.Asset_map[s]) == 64 {
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

// dReams-Dero swap objects and menu chain display content
func DreamsEntry() fyne.CanvasObject {
	Swap.DEntry = dwidget.WholeAmtEntry("dReams: ")
	Swap.DEntry.PlaceHolder = "dReams:"
	Swap.DEntry.Validator = validation.NewRegexp(`^(dReams: )[^0]\d{0,}$`, "Format Not Valid")
	Swap.DEntry.OnChanged = func(s string) {
		if Swap.DEntry.Validate() != nil {
			Swap.DEntry.SetText("dReams: 1")
		}
	}

	Assets.Gnomes_sync = canvas.NewText("", color.RGBA{31, 150, 200, 210})
	Assets.Gnomes_height = canvas.NewText(" Gnomon Height: ", color.White)
	Assets.Daem_height = canvas.NewText(" Daemon Height: ", color.White)
	Assets.Wall_height = canvas.NewText(" Wallet Height: ", color.White)
	Assets.Dreams_bal = canvas.NewText(" dReams Balance: ", color.White)
	Assets.Dero_bal = canvas.NewText(" Dero Balance: ", color.White)
	price := getOgre("DERO-USDT")
	Assets.Dero_price = canvas.NewText(" Dero Price: $"+price, color.White)

	Assets.Gnomes_sync.TextSize = 18
	Assets.Gnomes_height.TextSize = 18
	Assets.Daem_height.TextSize = 18
	Assets.Wall_height.TextSize = 18
	Assets.Dreams_bal.TextSize = 18
	Assets.Dero_bal.TextSize = 18
	Assets.Dero_price.TextSize = 18
	exLabel := canvas.NewText(" 1 Dero = 333 dReams", color.White)
	exLabel.TextSize = 18

	Swap.DEntry.SetText("dReams: 0")
	Swap.DEntry.Hide()

	box := *container.NewVBox(
		Assets.Gnomes_sync,
		Assets.Gnomes_height,
		Assets.Daem_height,
		Assets.Wall_height,
		Assets.Dreams_bal,
		Assets.Dero_bal,
		Assets.Dero_price, exLabel,
		Swap.DEntry)

	return &box
}

// Tournament deposit button
func TournamentButton(obj []fyne.CanvasObject, tabs *container.AppTabs) fyne.CanvasObject {
	Table.Tournament = widget.NewButton("Tournament", func() {
		obj[1] = tourneyConfirm(obj, tabs)
		obj[1].Refresh()
	})

	Table.Tournament.Hide()

	return Table.Tournament
}

// Confirmation for dReams-Dero swap
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

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	buttons := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewVBox(layout.NewSpacer(), label, layout.NewSpacer(), buttons)

	return container.NewMax(alpha, content)
}

// Holdero tournament chip deposit confirmation
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

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	buttons := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewVBox(layout.NewSpacer(), label, layout.NewSpacer(), buttons)

	return container.NewMax(alpha, content)
}

// Icon image for Holdero tables and asset viewing
func IconImg(res fyne.Resource) *fyne.Container {
	Assets.Icon.SetMinSize(fyne.NewSize(100, 100))
	Assets.Icon.Resize(fyne.NewSize(94, 94))
	Assets.Icon.Move(fyne.NewPos(8, 3))

	frame := canvas.NewImageFromResource(res)
	frame.Resize(fyne.NewSize(100, 100))
	frame.Move(fyne.NewPos(5, 0))

	cont := container.NewWithoutLayout(&Assets.Icon, frame)

	return cont
}

// Display for owned asset info
func AssetStats() fyne.CanvasObject {
	Assets.Collection = canvas.NewText(" Collection: ", color.White)
	Assets.Name = canvas.NewText(" Name: ", color.White)

	Assets.Name.TextSize = 18
	Assets.Collection.TextSize = 18

	Assets.Stats_box = *container.NewVBox(Assets.Collection, Assets.Name, IconImg(nil))

	return &Assets.Stats_box
}

// Set SCID header objects
func SetHeaderItems(obj []fyne.CanvasObject, tabs *container.AppTabs) fyne.CanvasObject {
	name_entry := widget.NewEntry()
	name_entry.PlaceHolder = "Name:"
	descr_entry := widget.NewEntry()
	descr_entry.PlaceHolder = "Description"
	icon_entry := widget.NewEntry()
	icon_entry.PlaceHolder = "Icon:"

	button := widget.NewButton("Set Headers", func() {
		scid := Assets.Index_entry.Text
		if len(scid) == 64 && name_entry.Text != "dReam Tables" {
			obj[1] = setHeaderConfirm(name_entry.Text, descr_entry.Text, icon_entry.Text, scid, obj, tabs)
			obj[1].Refresh()
		}
	})

	contr := container.NewVBox(name_entry, descr_entry, icon_entry, button)
	Assets.Header_box = *container.NewAdaptiveGrid(2, contr)
	Assets.Header_box.Hide()

	return &Assets.Header_box
}

// Confirmation for setting SCID headers
func setHeaderConfirm(name, desc, icon, scid string, obj []fyne.CanvasObject, reset *container.AppTabs) fyne.CanvasObject {
	label := widget.NewLabel("Headers for SCID:\n\n" + scid + "\n\nName: " + name + "\n\nDescription: " + desc + "\n\nIcon: " + icon)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	confirm_button := widget.NewButton("Confirm", func() {
		rpc.SetHeaders(name, desc, icon, scid)
		obj[1] = reset
		obj[1].Refresh()
	})

	cancel_button := widget.NewButton("Cancel", func() {
		obj[1] = reset
		obj[1].Refresh()

	})

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	buttons := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewVBox(layout.NewSpacer(), label, layout.NewSpacer(), buttons)

	return container.NewMax(alpha, content)
}
