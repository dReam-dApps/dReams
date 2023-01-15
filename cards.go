package main

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/table"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type cards struct {
	Hole1 fyne.CanvasObject
	Hole2 fyne.CanvasObject
	Flop1 fyne.CanvasObject
	Flop2 fyne.CanvasObject
	Flop3 fyne.CanvasObject
	Turn  fyne.CanvasObject
	River fyne.CanvasObject

	P1a fyne.CanvasObject
	P1b fyne.CanvasObject

	P2a fyne.CanvasObject
	P2b fyne.CanvasObject

	P3a fyne.CanvasObject
	P3b fyne.CanvasObject

	P4a fyne.CanvasObject
	P4b fyne.CanvasObject

	P5a fyne.CanvasObject
	P5b fyne.CanvasObject

	P6a fyne.CanvasObject
	P6b fyne.CanvasObject

	Layout *fyne.Container
}

var Cards cards

func Card(hash string) int { /// Gets local cards with local key
	for i := 1; i < 53; i++ {
		finder := strconv.Itoa(i)
		add := rpc.Wallet.ClientKey + finder + rpc.Round.SC_seed
		card := sha256.Sum256([]byte(add))
		str := hex.EncodeToString(card[:])

		if str == hash {
			return i
		}

	}
	return 0
}

// / Table cards
func Hole_1(c int, w, h float32) fyne.CanvasObject {
	Cards.Hole1 = DisplayCard(c)
	Cards.Hole1.Resize(fyne.NewSize(165, 225))
	Cards.Hole1.Move(fyne.NewPos(w-335, h-350))

	return Cards.Hole1
}

func Hole_2(c int, w, h float32) fyne.CanvasObject {
	Cards.Hole2 = DisplayCard(c)
	Cards.Hole2.Resize(fyne.NewSize(165, 225))
	Cards.Hole2.Move(fyne.NewPos(w-275, h-350))

	return Cards.Hole2
}

func Flop_1(c int) fyne.CanvasObject {
	Cards.Flop1 = DisplayCard(c)
	Cards.Flop1.Resize(fyne.NewSize(110, 150))
	Cards.Flop1.Move(fyne.NewPos(257, 185))

	return Cards.Flop1
}

func Flop_2(c int) fyne.CanvasObject {
	Cards.Flop2 = DisplayCard(c)
	Cards.Flop2.Resize(fyne.NewSize(110, 150))
	Cards.Flop2.Move(fyne.NewPos(377, 185))

	return Cards.Flop2
}

func Flop_3(c int) fyne.CanvasObject {
	Cards.Flop3 = DisplayCard(c)
	Cards.Flop3.Resize(fyne.NewSize(110, 150))
	Cards.Flop3.Move(fyne.NewPos(497, 185))

	return Cards.Flop3
}

func Turn(c int) fyne.CanvasObject {
	Cards.Turn = DisplayCard(c)
	Cards.Turn.Resize(fyne.NewSize(110, 150))
	Cards.Turn.Move(fyne.NewPos(617, 185))

	return Cards.Turn
}

func River(c int) fyne.CanvasObject {
	Cards.River = DisplayCard(c)
	Cards.River.Resize(fyne.NewSize(110, 150))
	Cards.River.Move(fyne.NewPos(737, 185))

	return Cards.River
}

func P1_a(c int) fyne.CanvasObject {
	Cards.P1a = DisplayCard(c)
	Cards.P1a.Resize(fyne.NewSize(110, 150))
	Cards.P1a.Move(fyne.NewPos(190, 10))

	return Cards.P1a
}

func P1_b(c int) fyne.CanvasObject {
	Cards.P1b = DisplayCard(c)
	Cards.P1b.Resize(fyne.NewSize(110, 150))
	Cards.P1b.Move(fyne.NewPos(242, 10))

	return Cards.P1b
}

func P2_a(c int) fyne.CanvasObject {
	Cards.P2a = DisplayCard(c)
	Cards.P2a.Resize(fyne.NewSize(110, 150))
	Cards.P2a.Move(fyne.NewPos(614, 10))

	return Cards.P2a
}

func P2_b(c int) fyne.CanvasObject {
	Cards.P2b = DisplayCard(c)
	Cards.P2b.Resize(fyne.NewSize(110, 150))
	Cards.P2b.Move(fyne.NewPos(666, 10))

	return Cards.P2b
}

func P3_a(c int) fyne.CanvasObject {
	Cards.P3a = DisplayCard(c)
	Cards.P3a.Resize(fyne.NewSize(110, 150))
	Cards.P3a.Move(fyne.NewPos(886, 115))

	return Cards.P3a

}

func P3_b(c int) fyne.CanvasObject {
	Cards.P3b = DisplayCard(c)
	Cards.P3b.Resize(fyne.NewSize(110, 150))
	Cards.P3b.Move(fyne.NewPos(938, 115))

	return Cards.P3b
}

func P4_a(c int) fyne.CanvasObject {
	Cards.P4a = DisplayCard(c)
	Cards.P4a.Resize(fyne.NewSize(110, 150))
	Cards.P4a.Move(fyne.NewPos(766, 361))

	return Cards.P4a
}

func P4_b(c int) fyne.CanvasObject {
	Cards.P4b = DisplayCard(c)
	Cards.P4b.Resize(fyne.NewSize(110, 150))
	Cards.P4b.Move(fyne.NewPos(818, 361))

	return Cards.P4b
}

func P5_a(c int) fyne.CanvasObject {
	Cards.P5a = DisplayCard(c)
	Cards.P5a.Resize(fyne.NewSize(110, 150))
	Cards.P5a.Move(fyne.NewPos(336, 361))

	return Cards.P5a
}

func P5_b(c int) fyne.CanvasObject {
	Cards.P5b = DisplayCard(c)
	Cards.P5b.Resize(fyne.NewSize(110, 150))
	Cards.P5b.Move(fyne.NewPos(388, 361))

	return Cards.P5b
}

func P6_a(c int) fyne.CanvasObject {
	Cards.P6a = DisplayCard(c)
	Cards.P6a.Resize(fyne.NewSize(110, 150))
	Cards.P6a.Move(fyne.NewPos(63, 254))

	return Cards.P6a
}

func P6_b(c int) fyne.CanvasObject {
	Cards.P6b = DisplayCard(c)
	Cards.P6b.Resize(fyne.NewSize(110, 150))
	Cards.P6b.Move(fyne.NewPos(115, 254))

	return Cards.P6b
}

func Is_In(hash string, who int, end bool) int {
	if hash != "" {
		if end {
			return rpc.KeyCard(hash, who)
		} else {
			return 0
		}
	} else {
		return 99
	}
}

func CustomCard(c int, face string) *canvas.Image {
	dir := table.GetDir()
	mid := "/cards/" + face + "/"
	path := dir + mid + cardEnd(c)

	if table.FileExists(path) {
		return canvas.NewImageFromFile(path)
	}

	return canvas.NewImageFromImage(nil)
}

func CustomBack(back string) *canvas.Image {
	dir := table.GetDir()
	post := "/cards/backs/" + back + ".png"
	path := dir + post

	if table.FileExists(path) {
		return canvas.NewImageFromFile(path)
	}

	return canvas.NewImageFromImage(nil)
}

func cardEnd(card int) string {
	var suffix string
	if card > 0 && card < 53 {
		switch card {
		case 1:
			suffix = "card1.png"
		case 2:
			suffix = "card2.png"
		case 3:
			suffix = "card3.png"
		case 4:
			suffix = "card4.png"
		case 5:
			suffix = "card5.png"
		case 6:
			suffix = "card6.png"
		case 7:
			suffix = "card7.png"
		case 8:
			suffix = "card8.png"
		case 9:
			suffix = "card9.png"
		case 10:
			suffix = "card10.png"
		case 11:
			suffix = "card11.png"
		case 12:
			suffix = "card12.png"
		case 13:
			suffix = "card13.png"
		case 14:
			suffix = "card14.png"
		case 15:
			suffix = "card15.png"
		case 16:
			suffix = "card16.png"
		case 17:
			suffix = "card17.png"
		case 18:
			suffix = "card18.png"
		case 19:
			suffix = "card19.png"
		case 20:
			suffix = "card20.png"
		case 21:
			suffix = "card21.png"
		case 22:
			suffix = "card22.png"
		case 23:
			suffix = "card23.png"
		case 24:
			suffix = "card24.png"
		case 25:
			suffix = "card25.png"
		case 26:
			suffix = "card26.png"
		case 27:
			suffix = "card27.png"
		case 28:
			suffix = "card28.png"
		case 29:
			suffix = "card29.png"
		case 30:
			suffix = "card30.png"
		case 31:
			suffix = "card31.png"
		case 32:
			suffix = "card32.png"
		case 33:
			suffix = "card33.png"
		case 34:
			suffix = "card34.png"
		case 35:
			suffix = "card35.png"
		case 36:
			suffix = "card36.png"
		case 37:
			suffix = "card37.png"
		case 38:
			suffix = "card38.png"
		case 39:
			suffix = "card39.png"
		case 40:
			suffix = "card40.png"
		case 41:
			suffix = "card41.png"
		case 42:
			suffix = "card42.png"
		case 43:
			suffix = "card43.png"
		case 44:
			suffix = "card44.png"
		case 45:
			suffix = "card45.png"
		case 46:
			suffix = "card46.png"
		case 47:
			suffix = "card47.png"
		case 48:
			suffix = "card48.png"
		case 49:
			suffix = "card49.png"
		case 50:
			suffix = "card50.png"
		case 51:
			suffix = "card51.png"
		case 52:
			suffix = "card52.png"
		}
	} else {
		suffix = "card1.png"
	}
	return suffix

}

func PlayerCards(dl, c1, c2, c3 int) fyne.CanvasObject {
	card1 := DisplayCard(c1)
	card2 := DisplayCard(c2)
	card3 := DisplayCard(c3)

	card1.Resize(fyne.NewSize(110, 150))
	card1.Move(fyne.NewPos(180, 20))

	card2.Resize(fyne.NewSize(110, 150))
	card2.Move(fyne.NewPos(290, 20))

	card3.Resize(fyne.NewSize(110, 150))
	card3.Move(fyne.NewPos(400, 20))

	box := container.NewWithoutLayout(
		card1,
		card2,
		card3,
	)

	return box
}

func BankerCards(dl, c1, c2, c3 int) fyne.CanvasObject {
	card1 := DisplayCard(c1)
	card2 := DisplayCard(c2)
	card3 := DisplayCard(c3)

	card1.Resize(fyne.NewSize(110, 150))
	card1.Move(fyne.NewPos(600, 20))

	card2.Resize(fyne.NewSize(110, 150))
	card2.Move(fyne.NewPos(710, 20))

	card3.Resize(fyne.NewSize(110, 150))
	card3.Move(fyne.NewPos(820, 20))

	box := container.NewWithoutLayout(
		card1,
		card2,
		card3,
	)

	return box
}

func BaccSuit(card int) int {
	if card == 99 {
		return 99
	}

	var suited int

	seed := rand.NewSource(time.Now().UnixNano())
	y := rand.New(seed)
	x := y.Intn(4) + 1

	if card == 0 {
		seed := rand.NewSource(time.Now().UnixNano())
		y := rand.New(seed)
		x := y.Intn(16) + 1

		switch x {
		case 1:
			suited = 10
		case 2:
			suited = 11
		case 3:
			suited = 12
		case 4:
			suited = 13
		case 5:
			suited = 23
		case 6:
			suited = 24
		case 7:
			suited = 25
		case 8:
			suited = 26
		case 9:
			suited = 36
		case 10:
			suited = 37
		case 11:
			suited = 38
		case 12:
			suited = 39
		case 13:
			suited = 49
		case 14:
			suited = 50
		case 15:
			suited = 51
		case 16:
			suited = 52
		}

		return suited
	}

	switch card {
	case 1:
		switch x {
		case 1:
			suited = 1
		case 2:
			suited = 14
		case 3:
			suited = 27
		case 4:
			suited = 40
		}
	case 2:
		switch x {
		case 1:
			suited = 2
		case 2:
			suited = 15
		case 3:
			suited = 28
		case 4:
			suited = 41
		}
	case 3:
		switch x {
		case 1:
			suited = 3
		case 2:
			suited = 16
		case 3:
			suited = 29
		case 4:
			suited = 42
		}
	case 4:
		switch x {
		case 1:
			suited = 4
		case 2:
			suited = 17
		case 3:
			suited = 30
		case 4:
			suited = 43
		}
	case 5:
		switch x {
		case 1:
			suited = 5
		case 2:
			suited = 18
		case 3:
			suited = 31
		case 4:
			suited = 44
		}
	case 6:
		switch x {
		case 1:
			suited = 6
		case 2:
			suited = 19
		case 3:
			suited = 32
		case 4:
			suited = 45
		}
	case 7:
		switch x {
		case 1:
			suited = 7
		case 2:
			suited = 20
		case 3:
			suited = 33
		case 4:
			suited = 46
		}
	case 8:
		switch x {
		case 1:
			suited = 8
		case 2:
			suited = 21
		case 3:
			suited = 34
		case 4:
			suited = 47
		}
	case 9:
		switch x {
		case 1:
			suited = 9
		case 2:
			suited = 22
		case 3:
			suited = 35
		case 4:
			suited = 48
		}
	case 10:
		switch x {
		case 1:
			suited = 10
		case 2:
			suited = 23
		case 3:
			suited = 36
		case 4:
			suited = 49
		}
	case 11:
		switch x {
		case 1:
			suited = 11
		case 2:
			suited = 24
		case 3:
			suited = 37
		case 4:
			suited = 50
		}
	case 12:
		switch x {
		case 1:
			suited = 12
		case 2:
			suited = 25
		case 3:
			suited = 38
		case 4:
			suited = 51
		}
	case 13:
		switch x {
		case 1:
			suited = 13
		case 2:
			suited = 26
		case 3:
			suited = 39
		case 4:
			suited = 52
		}
	}

	return suited
}

func TarotCard(c int) *canvas.Image {
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

func TarotItems(tabs *container.AppTabs) fyne.CanvasObject {
	search_entry := widget.NewEntry()
	search_entry.SetPlaceHolder("TXID:")
	search_button := widget.NewButton("    Search   ", func() {
		txid := search_entry.Text
		if len(txid) == 64 {
			signer, _ := rpc.VerifySigner(search_entry.Text)
			if signer {
				rpc.Tarot.Display = true
				table.Iluma.Label.SetText("")
				rpc.FetchTarotReading(rpc.Signal.Daemon, txid)
				if rpc.Tarot.T_card2 != 0 && rpc.Tarot.T_card3 != 0 {
					table.Iluma.Card1.Objects[1] = TarotCard(rpc.Tarot.T_card1)
					table.Iluma.Card2.Objects[1] = TarotCard(rpc.Tarot.T_card2)
					table.Iluma.Card3.Objects[1] = TarotCard(rpc.Tarot.T_card3)
					rpc.Tarot.Num = 3
				} else {
					table.Iluma.Card1.Objects[1] = TarotCard(0)
					table.Iluma.Card2.Objects[1] = TarotCard(rpc.Tarot.T_card1)
					table.Iluma.Card3.Objects[1] = TarotCard(0)
					rpc.Tarot.Num = 1
				}
				table.Iluma.Box.Refresh()
			} else {
				log.Println("[Tarot] This is not your reading")
			}
		}
	})

	table.Iluma.Draw1 = widget.NewButton("Draw One", func() {
		if !table.Iluma.Open {
			table.TarotConfirm(1)
		}
	})

	table.Iluma.Draw3 = widget.NewButton("Draw Three", func() {
		if !table.Iluma.Open {
			table.TarotConfirm(3)
		}
	})

	draw_cont := container.NewAdaptiveGrid(5,
		layout.NewSpacer(),
		layout.NewSpacer(),
		table.Iluma.Draw1,
		table.Iluma.Draw3,
		layout.NewSpacer())

	table.Iluma.Search = container.NewBorder(nil, nil, nil, search_button, search_entry)

	table.Iluma.Actions = container.NewVBox(
		layout.NewSpacer(),
		container.NewAdaptiveGrid(2, draw_cont, table.Iluma.Search))

	table.Iluma.Search.Hide()
	table.Iluma.Actions.Hide()
	max := container.NewMax(tabs, table.Iluma.Actions)

	return max
}
