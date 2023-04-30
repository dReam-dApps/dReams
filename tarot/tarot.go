package tarot

import (
	_ "embed"
	"fmt"
	"image/color"
	"math/rand"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/rpc"
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
}

//go:embed iluma/iluma.txt
var iluma_intro string

var Iluma tarot

// Tarot object buffer when action triggered
func TarotBuffer(d bool) {
	if d {
		Iluma.Card1.Objects[1] = canvas.NewImageFromResource(bundle.ResourceIluma81Png)
		Iluma.Card1.Refresh()
		Iluma.Card2.Objects[1] = canvas.NewImageFromResource(bundle.ResourceIluma81Png)
		Iluma.Card2.Refresh()
		Iluma.Card3.Objects[1] = canvas.NewImageFromResource(bundle.ResourceIluma81Png)
		Iluma.Card3.Refresh()
		Iluma.Draw1.Hide()
		Iluma.Draw3.Hide()
		rpc.Tarot.Card1 = 0
		rpc.Tarot.Card2 = 0
		rpc.Tarot.Card3 = 0
		rpc.Tarot.Last = ""
		Iluma.Search.Hide()
	} else {
		if rpc.Daemon.Connect && rpc.Wallet.Connect {
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

// Clickable Tarot card objects
func TarotCardBox() fyne.CanvasObject {
	Iluma.Label = widget.NewLabel("")
	Iluma.Label.Alignment = fyne.TextAlignCenter
	one := widget.NewButton("", func() {
		if rpc.Tarot.Num == 3 && !Iluma.Open && rpc.Tarot.Card1 > 0 {
			c := rpc.Tarot.Card1
			reset := Iluma.Card1
			Iluma.Card1 = *IlumaDialog(1, TarotDescription(c), reset)
		}
	})

	card_back := canvas.NewImageFromResource(bundle.ResourceIluma81Png)

	Iluma.Card1 = *container.NewMax(one, card_back)
	pad1 := container.NewBorder(nil, nil, TarotPadding(), TarotPadding(), &Iluma.Card1)

	two := widget.NewButton("", func() {
		if !Iluma.Open {
			reset := Iluma.Card2
			if rpc.Tarot.Num == 3 && rpc.Tarot.Card2 > 0 {
				c := rpc.Tarot.Card2
				Iluma.Card2 = *IlumaDialog(2, TarotDescription(c), reset)
			}

			if rpc.Tarot.Num == 1 && rpc.Tarot.Card1 > 0 {
				c := rpc.Tarot.Card1
				Iluma.Card2 = *IlumaDialog(2, TarotDescription(c), reset)
			}
		}
	})

	Iluma.Card2 = *container.NewMax(two, card_back)
	pad2 := container.NewBorder(nil, nil, TarotPadding(), TarotPadding(), &Iluma.Card2)

	three := widget.NewButton("", func() {
		if rpc.Tarot.Num == 3 && !Iluma.Open && rpc.Tarot.Card3 > 0 {
			c := rpc.Tarot.Card3
			reset := Iluma.Card3
			Iluma.Card3 = *IlumaDialog(3, TarotDescription(c), reset)
		}
	})

	one.Importance = widget.LowImportance
	two.Importance = widget.LowImportance
	three.Importance = widget.LowImportance

	Iluma.Card3 = *container.NewMax(three, card_back)
	pad3 := container.NewBorder(nil, nil, TarotPadding(), TarotPadding(), &Iluma.Card3)

	actions := container.NewAdaptiveGrid(3,
		layout.NewSpacer(),
		Iluma.Label,
		layout.NewSpacer())

	Iluma.Box = container.NewAdaptiveGrid(3,
		pad1,
		pad2,
		pad3)

	pad := container.NewBorder(
		nil,
		nil,
		layout.NewSpacer(),
		layout.NewSpacer(),
		Iluma.Box)

	alpha := canvas.NewRectangle(color.RGBA{0, 0, 0, 150})
	if bundle.AppColor == color.White {
		alpha = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x55})
	}

	box := *container.NewBorder(
		nil,
		actions,
		nil,
		nil,
		pad)

	max := container.NewMax(alpha, &box)

	return max
}

// Padding section for Tarot card layouts
func TarotPadding() fyne.CanvasObject {
	pad := container.NewHScroll(layout.NewSpacer())
	pad.SetMinSize(fyne.NewSize(40, 0))

	return pad
}

// Display random text when Tarot cards are drawn
func TarotDrawText() (text string) {
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

	return text
}

// Confirm Tarot draw of one or three cards
//   - i defines 1 or 3 card draw
func TarotConfirm(i int, reset fyne.Container) fyne.Container {
	label := widget.NewLabel("")
	if i == 3 {
		label.SetText(fmt.Sprintf("You are about to draw three cards\n\nReading fee is %.5f Dero\n\nConfirm", float64(rpc.IlumaFee)/100000))
	} else {
		label.SetText(fmt.Sprintf("You are about to draw one card\n\nReading fee is %.5f Dero\n\nConfirm", float64(rpc.IlumaFee)/100000))
	}

	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	confirm := widget.NewButton("Confirm", func() {
		Iluma.Card2 = reset
		Iluma.Card2.Refresh()

		if i == 3 {
			TarotBuffer(true)
			rpc.Tarot.Found = false
			rpc.Tarot.Display = false
			rpc.TarotReading(3)
			Iluma.Label.SetText(TarotDrawText())
		} else {
			TarotBuffer(true)
			rpc.Tarot.Found = false
			rpc.Tarot.Display = false
			rpc.TarotReading(1)
			Iluma.Label.SetText(TarotDrawText())
		}

		Iluma.Open = false
	})

	cancel := widget.NewButton("Cancel", func() {
		Iluma.Open = false
		Iluma.Draw1.Show()
		Iluma.Draw3.Show()
		go func() {
			Iluma.Card2 = reset
			Iluma.Card2.Refresh()
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

	return *container.NewMax(cont)
}

// Display Iluma description for Tarot card
//   - card is which card button pressed
//   - text is Iluma reading description
//   - Pass container to reset to
func IlumaDialog(card int, text string, reset fyne.Container) *fyne.Container {
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

// Iluma tab objects, intro description and image scroll
func PlaceIluma() *fyne.Container {
	var first, second, third bool
	var display int
	img := canvas.NewImageFromResource(bundle.ResourceIluma82Png)
	intro := widget.NewLabel(iluma_intro)
	scroll := container.NewScroll(intro)

	alpha := canvas.NewRectangle(color.RGBA{0, 0, 0, 120})
	if bundle.AppColor == color.White {
		alpha = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x55})
	}

	cont := container.NewGridWithColumns(2, scroll, img)

	scroll.OnScrolled = func(p fyne.Position) {
		if p.Y <= 400 {
			second = false
			third = false
			display = 1
		} else if p.Y >= 400 && p.Y <= 800 {
			first = false
			third = false
			display = 2
		} else if p.Y >= 800 {
			first = false
			second = false
			display = 3
		}

		switch display {
		case 1:
			if !first {
				cont.Objects[1] = canvas.NewImageFromResource(bundle.ResourceIluma82Png)
				cont.Refresh()
				first = true
			}
		case 2:
			if !second {
				cont.Objects[1] = canvas.NewImageFromResource(bundle.ResourceIluma80Png)
				cont.Refresh()
				second = true
			}
		case 3:
			if !third {
				cont.Objects[1] = canvas.NewImageFromResource(bundle.ResourceIluma83Png)
				cont.Refresh()
				third = true
			}
		default:

		}
	}

	return container.NewMax(alpha, cont)

}

//go:embed iluma/1.txt
var tarot_txt1 string

//go:embed iluma/2.txt
var tarot_txt2 string

//go:embed iluma/3.txt
var tarot_txt3 string

//go:embed iluma/4.txt
var tarot_txt4 string

//go:embed iluma/5.txt
var tarot_txt5 string

//go:embed iluma/6.txt
var tarot_txt6 string

//go:embed iluma/7.txt
var tarot_txt7 string

//go:embed iluma/8.txt
var tarot_txt8 string

//go:embed iluma/9.txt
var tarot_txt9 string

//go:embed iluma/10.txt
var tarot_txt10 string

//go:embed iluma/11.txt
var tarot_txt11 string

//go:embed iluma/12.txt
var tarot_txt12 string

//go:embed iluma/13.txt
var tarot_txt13 string

//go:embed iluma/14.txt
var tarot_txt14 string

//go:embed iluma/15.txt
var tarot_txt15 string

//go:embed iluma/16.txt
var tarot_txt16 string

//go:embed iluma/17.txt
var tarot_txt17 string

//go:embed iluma/18.txt
var tarot_txt18 string

//go:embed iluma/19.txt
var tarot_txt19 string

//go:embed iluma/20.txt
var tarot_txt20 string

//go:embed iluma/21.txt
var tarot_txt21 string

//go:embed iluma/22.txt
var tarot_txt22 string

//go:embed iluma/23.txt
var tarot_txt23 string

//go:embed iluma/24.txt
var tarot_txt24 string

//go:embed iluma/25.txt
var tarot_txt25 string

//go:embed iluma/26.txt
var tarot_txt26 string

//go:embed iluma/27.txt
var tarot_txt27 string

//go:embed iluma/28.txt
var tarot_txt28 string

//go:embed iluma/29.txt
var tarot_txt29 string

//go:embed iluma/30.txt
var tarot_txt30 string

//go:embed iluma/31.txt
var tarot_txt31 string

//go:embed iluma/32.txt
var tarot_txt32 string

//go:embed iluma/33.txt
var tarot_txt33 string

//go:embed iluma/34.txt
var tarot_txt34 string

//go:embed iluma/35.txt
var tarot_txt35 string

//go:embed iluma/36.txt
var tarot_txt36 string

//go:embed iluma/37.txt
var tarot_txt37 string

//go:embed iluma/38.txt
var tarot_txt38 string

//go:embed iluma/39.txt
var tarot_txt39 string

//go:embed iluma/40.txt
var tarot_txt40 string

//go:embed iluma/41.txt
var tarot_txt41 string

//go:embed iluma/42.txt
var tarot_txt42 string

//go:embed iluma/43.txt
var tarot_txt43 string

//go:embed iluma/44.txt
var tarot_txt44 string

//go:embed iluma/45.txt
var tarot_txt45 string

//go:embed iluma/46.txt
var tarot_txt46 string

//go:embed iluma/47.txt
var tarot_txt47 string

//go:embed iluma/48.txt
var tarot_txt48 string

//go:embed iluma/49.txt
var tarot_txt49 string

//go:embed iluma/50.txt
var tarot_txt50 string

//go:embed iluma/51.txt
var tarot_txt51 string

//go:embed iluma/52.txt
var tarot_txt52 string

//go:embed iluma/53.txt
var tarot_txt53 string

//go:embed iluma/54.txt
var tarot_txt54 string

//go:embed iluma/55.txt
var tarot_txt55 string

//go:embed iluma/56.txt
var tarot_txt56 string

//go:embed iluma/57.txt
var tarot_txt57 string

//go:embed iluma/58.txt
var tarot_txt58 string

//go:embed iluma/59.txt
var tarot_txt59 string

//go:embed iluma/60.txt
var tarot_txt60 string

//go:embed iluma/61.txt
var tarot_txt61 string

//go:embed iluma/62.txt
var tarot_txt62 string

//go:embed iluma/63.txt
var tarot_txt63 string

//go:embed iluma/64.txt
var tarot_txt64 string

//go:embed iluma/65.txt
var tarot_txt65 string

//go:embed iluma/66.txt
var tarot_txt66 string

//go:embed iluma/67.txt
var tarot_txt67 string

//go:embed iluma/68.txt
var tarot_txt68 string

//go:embed iluma/69.txt
var tarot_txt69 string

//go:embed iluma/70.txt
var tarot_txt70 string

//go:embed iluma/71.txt
var tarot_txt71 string

//go:embed iluma/72.txt
var tarot_txt72 string

//go:embed iluma/73.txt
var tarot_txt73 string

//go:embed iluma/74.txt
var tarot_txt74 string

//go:embed iluma/75.txt
var tarot_txt75 string

//go:embed iluma/76.txt
var tarot_txt76 string

//go:embed iluma/77.txt
var tarot_txt77 string

//go:embed iluma/78.txt
var tarot_txt78 string

// Iluma description text switch
func TarotDescription(c int) string {
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
