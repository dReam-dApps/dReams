package tarot

import (
	_ "embed"
	"fmt"
	"log"
	"math/rand"
	"time"

	dreams "github.com/SixofClubsss/dReams"
	"github.com/SixofClubsss/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type tarot struct {
	Card1   fyne.Container
	Card2   fyne.Container
	Card3   fyne.Container
	Back    fyne.Resource
	Draw1   *widget.Button
	Draw3   *widget.Button
	Search  *fyne.Container
	Actions *fyne.Container
	Label   *widget.Label
	Box     *fyne.Container
	Open    bool
	Value   struct {
		Card1    int
		Card2    int
		Card3    int
		CHeight  int
		Num      int
		Last     string
		Readings string
		Found    bool
		Display  bool
		Notified bool
	}
}

//go:embed text/iluma.txt
var iluma_intro string

var Iluma tarot

func initValues() {
	Iluma.Value.Display = true
}

// Main Tarot process
func fetch(t *dreams.DreamsItems, d dreams.DreamsObject) {
	initValues()
	time.Sleep(3 * time.Second)
	for {
		select {
		case <-d.Receive():
			if !rpc.Wallet.IsConnected() || !rpc.Daemon.IsConnected() {
				disableActions(true)
				tarotRefresh(t)
				d.WorkDone()
				continue
			}

			FetchTarotSC()
			tarotRefresh(t)
			if Iluma.Value.Found && !Iluma.Value.Notified {
				if !d.IsWindows() {
					Iluma.Value.Notified = d.Notification("dReams - Iluma", "Your Reading has Arrived")
				}
			}
			d.WorkDone()
		case <-d.CloseDapp():
			log.Println("[Iluma] Done")
			return
		}
	}
}

// Tarot object buffer when action triggered
func ActionBuffer(d bool) {
	if d {
		Iluma.Card1.Objects[1] = canvas.NewImageFromResource(resourceIluma81Png)
		Iluma.Card1.Refresh()
		Iluma.Card2.Objects[1] = canvas.NewImageFromResource(resourceIluma81Png)
		Iluma.Card2.Refresh()
		Iluma.Card3.Objects[1] = canvas.NewImageFromResource(resourceIluma81Png)
		Iluma.Card3.Refresh()
		Iluma.Draw1.Hide()
		Iluma.Draw3.Hide()
		Iluma.Value.Card1 = 0
		Iluma.Value.Card2 = 0
		Iluma.Value.Card3 = 0
		Iluma.Value.Last = ""
		Iluma.Search.Hide()
	} else {
		if rpc.IsReady() {
			if !Iluma.Open {
				Iluma.Draw1.Show()
				Iluma.Draw3.Show()
			}
			Iluma.Search.Show()
		}
	}

	Iluma.Draw1.Refresh()
	Iluma.Draw3.Refresh()
}

func disableActions(b bool) {
	if b {
		Iluma.Draw1.Hide()
		Iluma.Draw3.Hide()
		Iluma.Search.Hide()
	} else {
		Iluma.Draw1.Show()
		Iluma.Draw3.Show()
		Iluma.Search.Show()
	}

	Iluma.Draw1.Refresh()
	Iluma.Draw3.Refresh()
	Iluma.Search.Refresh()
}

// Refresh all Tarot objects
func tarotRefresh(t *dreams.DreamsItems) {
	t.LeftLabel.SetText("Total Readings: " + Iluma.Value.Readings + "      Click your card for Iluma reading")
	t.RightLabel.SetText("dReams Balance: " + rpc.DisplayBalance("dReams") + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Wallet.Display.Height)

	if !Iluma.Value.Display {
		FetchReading(Iluma.Value.Last)
		Iluma.Box.Refresh()
		if Iluma.Value.Found {
			Iluma.Value.Display = true
			Iluma.Label.SetText("")
			if Iluma.Value.Num == 3 {
				Iluma.Card1.Objects[1] = DisplayCard(Iluma.Value.Card1)
				Iluma.Card2.Objects[1] = DisplayCard(Iluma.Value.Card2)
				Iluma.Card3.Objects[1] = DisplayCard(Iluma.Value.Card3)
			} else {
				Iluma.Card1.Objects[1] = DisplayCard(0)
				Iluma.Card2.Objects[1] = DisplayCard(Iluma.Value.Card1)
				Iluma.Card3.Objects[1] = DisplayCard(0)
			}
			ActionBuffer(false)
			Iluma.Box.Refresh()
		}
	}

	if rpc.Wallet.Height > Iluma.Value.CHeight+3 {
		ActionBuffer(false)
	}

	t.DApp.Refresh()
}

// Display random text when Tarot cards are drawn
func drawText() (text string) {
	i := rand.Intn(6-1) + 1

	switch i {
	case 1:
		text = "Accessing the Akashic Records"
	case 2:
		text = "Consulting your Angels & Ancestors"
	case 3:
		text = "Scanning your Auroa"
	case 4:
		text = "Reading your Light Codes"
	case 5:
		text = "Channeling the Divine"
	case 6:
		text = "Trust in your intuition"
	default:

	}

	return
}

// Confirm Tarot draw of one or three cards
//   - i defines 1 or 3 card draw
func drawConfirm(i int, reset fyne.Container) *fyne.Container {
	label := widget.NewLabel("")
	if i == 3 {
		label.SetText(fmt.Sprintf("You are about to draw three cards\n\nReading fee is %.5f Dero\n\nConfirm", float64(rpc.IlumaFee)/100000))
	} else {
		label.SetText(fmt.Sprintf("You are about to draw one card\n\nReading fee is %.5f Dero\n\nConfirm", float64(rpc.IlumaFee)/100000))
	}

	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	confirm := widget.NewButton("Confirm", func() {
		Iluma.Card2.Objects = reset.Objects
		Iluma.Card2.Refresh()

		if i == 3 {
			ActionBuffer(true)
			Iluma.Value.Found = false
			Iluma.Value.Display = false
			DrawReading(3)
			Iluma.Label.SetText(drawText())
		} else {
			ActionBuffer(true)
			Iluma.Value.Found = false
			Iluma.Value.Display = false
			DrawReading(1)
			Iluma.Label.SetText(drawText())
		}

		Iluma.Open = false
	})

	cancel := widget.NewButton("Cancel", func() {
		Iluma.Open = false
		go func() {
			Iluma.Card2.Objects = reset.Objects
			Iluma.Card2.Refresh()
			Iluma.Draw1.Show()
			Iluma.Draw3.Show()
		}()
	})

	box := container.NewAdaptiveGrid(2, confirm, cancel)
	cont := container.NewBorder(
		nil,
		box,
		nil,
		nil,
		container.NewVBox(layout.NewSpacer(), label, layout.NewSpacer()))

	Iluma.Open = true

	return container.NewMax(cont)
}

// Display Iluma description for Tarot card
//   - card is which card button pressed
//   - text is Iluma reading description
//   - Pass container to reset to
func ilumaDialog(card int, text string, reset fyne.Container) *fyne.Container {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord

	scroll := container.NewVScroll(label)

	Iluma.Open = true
	Iluma.Actions.Hide()
	Iluma.Search.Hide()

	reset_button := widget.NewButton("", func() {
		switch card {
		case 1:
			Iluma.Card1 = reset
		case 2:
			Iluma.Card2 = reset
		case 3:
			Iluma.Card3 = reset
		default:

		}

		Iluma.Open = false
		Iluma.Actions.Show()
		Iluma.Search.Show()
	})
	reset_button.Importance = widget.LowImportance

	return container.NewMax(reset_button, scroll)
}

// Switch for Iluma Tarot card image
func DisplayCard(c int) *canvas.Image {
	switch c {
	case 1:
		return canvas.NewImageFromResource(resourceIluma1Jpg)
	case 2:
		return canvas.NewImageFromResource(resourceIluma2Jpg)
	case 3:
		return canvas.NewImageFromResource(resourceIluma3Jpg)
	case 4:
		return canvas.NewImageFromResource(resourceIluma4Jpg)
	case 5:
		return canvas.NewImageFromResource(resourceIluma5Jpg)
	case 6:
		return canvas.NewImageFromResource(resourceIluma6Jpg)
	case 7:
		return canvas.NewImageFromResource(resourceIluma7Jpg)
	case 8:
		return canvas.NewImageFromResource(resourceIluma8Jpg)
	case 9:
		return canvas.NewImageFromResource(resourceIluma9Jpg)
	case 10:
		return canvas.NewImageFromResource(resourceIluma10Jpg)
	case 11:
		return canvas.NewImageFromResource(resourceIluma11Jpg)
	case 12:
		return canvas.NewImageFromResource(resourceIluma12Jpg)
	case 13:
		return canvas.NewImageFromResource(resourceIluma13Jpg)
	case 14:
		return canvas.NewImageFromResource(resourceIluma14Jpg)
	case 15:
		return canvas.NewImageFromResource(resourceIluma15Jpg)
	case 16:
		return canvas.NewImageFromResource(resourceIluma16Jpg)
	case 17:
		return canvas.NewImageFromResource(resourceIluma17Jpg)
	case 18:
		return canvas.NewImageFromResource(resourceIluma18Jpg)
	case 19:
		return canvas.NewImageFromResource(resourceIluma19Jpg)
	case 20:
		return canvas.NewImageFromResource(resourceIluma20Jpg)
	case 21:
		return canvas.NewImageFromResource(resourceIluma21Jpg)
	case 22:
		return canvas.NewImageFromResource(resourceIluma22Jpg)
	case 23:
		return canvas.NewImageFromResource(resourceIluma23Jpg)
	case 24:
		return canvas.NewImageFromResource(resourceIluma24Jpg)
	case 25:
		return canvas.NewImageFromResource(resourceIluma25Jpg)
	case 26:
		return canvas.NewImageFromResource(resourceIluma26Jpg)
	case 27:
		return canvas.NewImageFromResource(resourceIluma27Jpg)
	case 28:
		return canvas.NewImageFromResource(resourceIluma28Jpg)
	case 29:
		return canvas.NewImageFromResource(resourceIluma29Jpg)
	case 30:
		return canvas.NewImageFromResource(resourceIluma30Jpg)
	case 31:
		return canvas.NewImageFromResource(resourceIluma31Jpg)
	case 32:
		return canvas.NewImageFromResource(resourceIluma32Jpg)
	case 33:
		return canvas.NewImageFromResource(resourceIluma33Jpg)
	case 34:
		return canvas.NewImageFromResource(resourceIluma34Jpg)
	case 35:
		return canvas.NewImageFromResource(resourceIluma35Jpg)
	case 36:
		return canvas.NewImageFromResource(resourceIluma36Jpg)
	case 37:
		return canvas.NewImageFromResource(resourceIluma37Jpg)
	case 38:
		return canvas.NewImageFromResource(resourceIluma38Jpg)
	case 39:
		return canvas.NewImageFromResource(resourceIluma39Jpg)
	case 40:
		return canvas.NewImageFromResource(resourceIluma40Jpg)
	case 41:
		return canvas.NewImageFromResource(resourceIluma41Jpg)
	case 42:
		return canvas.NewImageFromResource(resourceIluma42Jpg)
	case 43:
		return canvas.NewImageFromResource(resourceIluma43Jpg)
	case 44:
		return canvas.NewImageFromResource(resourceIluma44Jpg)
	case 45:
		return canvas.NewImageFromResource(resourceIluma45Jpg)
	case 46:
		return canvas.NewImageFromResource(resourceIluma46Jpg)
	case 47:
		return canvas.NewImageFromResource(resourceIluma47Jpg)
	case 48:
		return canvas.NewImageFromResource(resourceIluma48Jpg)
	case 49:
		return canvas.NewImageFromResource(resourceIluma49Jpg)
	case 50:
		return canvas.NewImageFromResource(resourceIluma50Jpg)
	case 51:
		return canvas.NewImageFromResource(resourceIluma51Jpg)
	case 52:
		return canvas.NewImageFromResource(resourceIluma52Jpg)
	case 53:
		return canvas.NewImageFromResource(resourceIluma53Jpg)
	case 54:
		return canvas.NewImageFromResource(resourceIluma54Jpg)
	case 55:
		return canvas.NewImageFromResource(resourceIluma55Jpg)
	case 56:
		return canvas.NewImageFromResource(resourceIluma56Jpg)
	case 57:
		return canvas.NewImageFromResource(resourceIluma57Jpg)
	case 58:
		return canvas.NewImageFromResource(resourceIluma58Jpg)
	case 59:
		return canvas.NewImageFromResource(resourceIluma59Jpg)
	case 60:
		return canvas.NewImageFromResource(resourceIluma60Jpg)
	case 61:
		return canvas.NewImageFromResource(resourceIluma61Jpg)
	case 62:
		return canvas.NewImageFromResource(resourceIluma62Jpg)
	case 63:
		return canvas.NewImageFromResource(resourceIluma63Jpg)
	case 64:
		return canvas.NewImageFromResource(resourceIluma64Jpg)
	case 65:
		return canvas.NewImageFromResource(resourceIluma65Jpg)
	case 66:
		return canvas.NewImageFromResource(resourceIluma66Jpg)
	case 67:
		return canvas.NewImageFromResource(resourceIluma67Jpg)
	case 68:
		return canvas.NewImageFromResource(resourceIluma68Jpg)
	case 69:
		return canvas.NewImageFromResource(resourceIluma69Jpg)
	case 70:
		return canvas.NewImageFromResource(resourceIluma70Jpg)
	case 71:
		return canvas.NewImageFromResource(resourceIluma71Jpg)
	case 72:
		return canvas.NewImageFromResource(resourceIluma72Jpg)
	case 73:
		return canvas.NewImageFromResource(resourceIluma73Jpg)
	case 74:
		return canvas.NewImageFromResource(resourceIluma74Jpg)
	case 75:
		return canvas.NewImageFromResource(resourceIluma75Jpg)
	case 76:
		return canvas.NewImageFromResource(resourceIluma76Jpg)
	case 77:
		return canvas.NewImageFromResource(resourceIluma77Jpg)
	case 78:
		return canvas.NewImageFromResource(resourceIluma78Jpg)
	default:
		return canvas.NewImageFromResource(resourceIluma81Png)
	}
}

//go:embed text/1.txt
var tarot_txt1 string

//go:embed text/2.txt
var tarot_txt2 string

//go:embed text/3.txt
var tarot_txt3 string

//go:embed text/4.txt
var tarot_txt4 string

//go:embed text/5.txt
var tarot_txt5 string

//go:embed text/6.txt
var tarot_txt6 string

//go:embed text/7.txt
var tarot_txt7 string

//go:embed text/8.txt
var tarot_txt8 string

//go:embed text/9.txt
var tarot_txt9 string

//go:embed text/10.txt
var tarot_txt10 string

//go:embed text/11.txt
var tarot_txt11 string

//go:embed text/12.txt
var tarot_txt12 string

//go:embed text/13.txt
var tarot_txt13 string

//go:embed text/14.txt
var tarot_txt14 string

//go:embed text/15.txt
var tarot_txt15 string

//go:embed text/16.txt
var tarot_txt16 string

//go:embed text/17.txt
var tarot_txt17 string

//go:embed text/18.txt
var tarot_txt18 string

//go:embed text/19.txt
var tarot_txt19 string

//go:embed text/20.txt
var tarot_txt20 string

//go:embed text/21.txt
var tarot_txt21 string

//go:embed text/22.txt
var tarot_txt22 string

//go:embed text/23.txt
var tarot_txt23 string

//go:embed text/24.txt
var tarot_txt24 string

//go:embed text/25.txt
var tarot_txt25 string

//go:embed text/26.txt
var tarot_txt26 string

//go:embed text/27.txt
var tarot_txt27 string

//go:embed text/28.txt
var tarot_txt28 string

//go:embed text/29.txt
var tarot_txt29 string

//go:embed text/30.txt
var tarot_txt30 string

//go:embed text/31.txt
var tarot_txt31 string

//go:embed text/32.txt
var tarot_txt32 string

//go:embed text/33.txt
var tarot_txt33 string

//go:embed text/34.txt
var tarot_txt34 string

//go:embed text/35.txt
var tarot_txt35 string

//go:embed text/36.txt
var tarot_txt36 string

//go:embed text/37.txt
var tarot_txt37 string

//go:embed text/38.txt
var tarot_txt38 string

//go:embed text/39.txt
var tarot_txt39 string

//go:embed text/40.txt
var tarot_txt40 string

//go:embed text/41.txt
var tarot_txt41 string

//go:embed text/42.txt
var tarot_txt42 string

//go:embed text/43.txt
var tarot_txt43 string

//go:embed text/44.txt
var tarot_txt44 string

//go:embed text/45.txt
var tarot_txt45 string

//go:embed text/46.txt
var tarot_txt46 string

//go:embed text/47.txt
var tarot_txt47 string

//go:embed text/48.txt
var tarot_txt48 string

//go:embed text/49.txt
var tarot_txt49 string

//go:embed text/50.txt
var tarot_txt50 string

//go:embed text/51.txt
var tarot_txt51 string

//go:embed text/52.txt
var tarot_txt52 string

//go:embed text/53.txt
var tarot_txt53 string

//go:embed text/54.txt
var tarot_txt54 string

//go:embed text/55.txt
var tarot_txt55 string

//go:embed text/56.txt
var tarot_txt56 string

//go:embed text/57.txt
var tarot_txt57 string

//go:embed text/58.txt
var tarot_txt58 string

//go:embed text/59.txt
var tarot_txt59 string

//go:embed text/60.txt
var tarot_txt60 string

//go:embed text/61.txt
var tarot_txt61 string

//go:embed text/62.txt
var tarot_txt62 string

//go:embed text/63.txt
var tarot_txt63 string

//go:embed text/64.txt
var tarot_txt64 string

//go:embed text/65.txt
var tarot_txt65 string

//go:embed text/66.txt
var tarot_txt66 string

//go:embed text/67.txt
var tarot_txt67 string

//go:embed text/68.txt
var tarot_txt68 string

//go:embed text/69.txt
var tarot_txt69 string

//go:embed text/70.txt
var tarot_txt70 string

//go:embed text/71.txt
var tarot_txt71 string

//go:embed text/72.txt
var tarot_txt72 string

//go:embed text/73.txt
var tarot_txt73 string

//go:embed text/74.txt
var tarot_txt74 string

//go:embed text/75.txt
var tarot_txt75 string

//go:embed text/76.txt
var tarot_txt76 string

//go:embed text/77.txt
var tarot_txt77 string

//go:embed text/78.txt
var tarot_txt78 string

// Iluma description text switch
func ilumaDescription(c int) string {
	switch c {
	case 1:
		return tarot_txt1
	case 2:
		return tarot_txt2
	case 3:
		return tarot_txt3
	case 4:
		return tarot_txt4
	case 5:
		return tarot_txt5
	case 6:
		return tarot_txt6
	case 7:
		return tarot_txt7
	case 8:
		return tarot_txt8
	case 9:
		return tarot_txt9
	case 10:
		return tarot_txt10
	case 11:
		return tarot_txt11
	case 12:
		return tarot_txt12
	case 13:
		return tarot_txt13
	case 14:
		return tarot_txt14
	case 15:
		return tarot_txt15
	case 16:
		return tarot_txt16
	case 17:
		return tarot_txt17
	case 18:
		return tarot_txt18
	case 19:
		return tarot_txt19
	case 20:
		return tarot_txt20
	case 21:
		return tarot_txt21
	case 22:
		return tarot_txt22
	case 23:
		return tarot_txt23
	case 24:
		return tarot_txt24
	case 25:
		return tarot_txt25
	case 26:
		return tarot_txt26
	case 27:
		return tarot_txt27
	case 28:
		return tarot_txt28
	case 29:
		return tarot_txt29
	case 30:
		return tarot_txt30
	case 31:
		return tarot_txt31
	case 32:
		return tarot_txt32
	case 33:
		return tarot_txt33
	case 34:
		return tarot_txt34
	case 35:
		return tarot_txt35
	case 36:
		return tarot_txt36
	case 37:
		return tarot_txt37
	case 38:
		return tarot_txt38
	case 39:
		return tarot_txt39
	case 40:
		return tarot_txt40
	case 41:
		return tarot_txt41
	case 42:
		return tarot_txt42
	case 43:
		return tarot_txt43
	case 44:
		return tarot_txt44
	case 45:
		return tarot_txt45
	case 46:
		return tarot_txt46
	case 47:
		return tarot_txt47
	case 48:
		return tarot_txt48
	case 49:
		return tarot_txt49
	case 50:
		return tarot_txt50
	case 51:
		return tarot_txt51
	case 52:
		return tarot_txt52
	case 53:
		return tarot_txt53
	case 54:
		return tarot_txt54
	case 55:
		return tarot_txt55
	case 56:
		return tarot_txt56
	case 57:
		return tarot_txt57
	case 58:
		return tarot_txt58
	case 59:
		return tarot_txt59
	case 60:
		return tarot_txt60
	case 61:
		return tarot_txt61
	case 62:
		return tarot_txt62
	case 63:
		return tarot_txt63
	case 64:
		return tarot_txt64
	case 65:
		return tarot_txt65
	case 66:
		return tarot_txt66
	case 67:
		return tarot_txt67
	case 68:
		return tarot_txt68
	case 69:
		return tarot_txt69
	case 70:
		return tarot_txt70
	case 71:
		return tarot_txt71
	case 72:
		return tarot_txt72
	case 73:
		return tarot_txt73
	case 74:
		return tarot_txt74
	case 75:
		return tarot_txt75
	case 76:
		return tarot_txt76
	case 77:
		return tarot_txt77
	case 78:
		return tarot_txt78
	}

	return ""
}
