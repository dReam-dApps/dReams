package main

import (
	"math/rand"
	"time"

	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/holdero"
	"github.com/SixofClubsss/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
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

// Set player hole card one image
//   - w and h of main window for resize
func Hole_1(c int, w, h float32) fyne.CanvasObject {
	Cards.Hole1 = DisplayCard(c)
	Cards.Hole1.Resize(fyne.NewSize(165, 225))
	Cards.Hole1.Move(fyne.NewPos(w-335, h-335))

	return Cards.Hole1
}

// Set player hole card two image
//   - w and h of main window for resize
func Hole_2(c int, w, h float32) fyne.CanvasObject {
	Cards.Hole2 = DisplayCard(c)
	Cards.Hole2.Resize(fyne.NewSize(165, 225))
	Cards.Hole2.Move(fyne.NewPos(w-275, h-335))

	return Cards.Hole2
}

// Set first flop card image
func Flop_1(c int) fyne.CanvasObject {
	size := fyne.NewSize(110, 150)
	pos := fyne.NewPos(260, 203)
	Cards.Flop1 = DisplayCard(c)
	Cards.Flop1.Resize(size)
	Cards.Flop1.Move(pos)

	for _, i := range rpc.Round.Winning_hand {
		if c == i {
			highlight := canvas.NewRectangle(bundle.Highlight)
			highlight.Resize(size)
			highlight.Move(pos)

			return container.NewWithoutLayout(Cards.Flop1, highlight)
		}
	}

	return Cards.Flop1
}

// Set second flop card image
func Flop_2(c int) fyne.CanvasObject {
	size := fyne.NewSize(110, 150)
	pos := fyne.NewPos(380, 203)
	Cards.Flop2 = DisplayCard(c)
	Cards.Flop2.Resize(size)
	Cards.Flop2.Move(pos)

	for _, i := range rpc.Round.Winning_hand {
		if c == i {
			highlight := canvas.NewRectangle(bundle.Highlight)
			highlight.Resize(size)
			highlight.Move(pos)

			return container.NewWithoutLayout(Cards.Flop2, highlight)
		}
	}

	return Cards.Flop2
}

// Set third flop card image
func Flop_3(c int) fyne.CanvasObject {
	size := fyne.NewSize(110, 150)
	pos := fyne.NewPos(500, 203)
	Cards.Flop3 = DisplayCard(c)
	Cards.Flop3.Resize(size)
	Cards.Flop3.Move(pos)

	for _, i := range rpc.Round.Winning_hand {
		if c == i {
			highlight := canvas.NewRectangle(bundle.Highlight)
			highlight.Resize(size)
			highlight.Move(pos)

			return container.NewWithoutLayout(Cards.Flop3, highlight)
		}
	}

	return Cards.Flop3
}

// Set turn card image
func Turn(c int) fyne.CanvasObject {
	size := fyne.NewSize(110, 150)
	pos := fyne.NewPos(620, 203)
	Cards.Turn = DisplayCard(c)
	Cards.Turn.Resize(size)
	Cards.Turn.Move(pos)

	for _, i := range rpc.Round.Winning_hand {
		if c == i {
			highlight := canvas.NewRectangle(bundle.Highlight)
			highlight.Resize(size)
			highlight.Move(pos)

			return container.NewWithoutLayout(Cards.Turn, highlight)
		}
	}

	return Cards.Turn
}

// Set river card image
func River(c int) fyne.CanvasObject {
	size := fyne.NewSize(110, 150)
	pos := fyne.NewPos(740, 203)
	Cards.River = DisplayCard(c)
	Cards.River.Resize(size)
	Cards.River.Move(pos)

	for _, i := range rpc.Round.Winning_hand {
		if c == i {
			highlight := canvas.NewRectangle(bundle.Highlight)
			highlight.Resize(size)
			highlight.Move(pos)

			return container.NewWithoutLayout(Cards.River, highlight)
		}
	}

	return Cards.River
}

// Set first players table card one image
func P1_a(c int) fyne.CanvasObject {
	size := fyne.NewSize(110, 150)
	pos := fyne.NewPos(190, 25)
	Cards.P1a = DisplayCard(c)
	Cards.P1a.Resize(size)
	Cards.P1a.Move(pos)

	for _, i := range rpc.Round.Winning_hand {
		if c == i {
			highlight := canvas.NewRectangle(bundle.Highlight)
			highlight.Resize(size)
			highlight.Move(pos)

			return container.NewWithoutLayout(Cards.P1a, highlight)
		}
	}

	return Cards.P1a
}

// Set first players table card two image
func P1_b(c int) fyne.CanvasObject {
	size := fyne.NewSize(110, 150)
	pos := fyne.NewPos(242, 25)
	Cards.P1b = DisplayCard(c)
	Cards.P1b.Resize(size)
	Cards.P1b.Move(pos)

	for _, i := range rpc.Round.Winning_hand {
		if c == i {
			highlight := canvas.NewRectangle(bundle.Highlight)
			highlight.Resize(size)
			highlight.Move(pos)

			return container.NewWithoutLayout(Cards.P1b, highlight)
		}
	}

	return Cards.P1b
}

// Set second players table card one image
func P2_a(c int) fyne.CanvasObject {
	size := fyne.NewSize(110, 150)
	pos := fyne.NewPos(614, 25)
	Cards.P2a = DisplayCard(c)
	Cards.P2a.Resize(size)
	Cards.P2a.Move(pos)

	for _, i := range rpc.Round.Winning_hand {
		if c == i {
			highlight := canvas.NewRectangle(bundle.Highlight)
			highlight.Resize(size)
			highlight.Move(pos)

			return container.NewWithoutLayout(Cards.P2a, highlight)
		}
	}

	return Cards.P2a
}

// Set second players table card two image
func P2_b(c int) fyne.CanvasObject {
	size := fyne.NewSize(110, 150)
	pos := fyne.NewPos(666, 25)
	Cards.P2b = DisplayCard(c)
	Cards.P2b.Resize(size)
	Cards.P2b.Move(pos)

	for _, i := range rpc.Round.Winning_hand {
		if c == i {
			highlight := canvas.NewRectangle(bundle.Highlight)
			highlight.Resize(size)
			highlight.Move(pos)

			return container.NewWithoutLayout(Cards.P2b, highlight)
		}
	}

	return Cards.P2b
}

// Set third players table card one image
func P3_a(c int) fyne.CanvasObject {
	size := fyne.NewSize(110, 150)
	pos := fyne.NewPos(883, 129)
	Cards.P3a = DisplayCard(c)
	Cards.P3a.Resize(size)
	Cards.P3a.Move(pos)

	for _, i := range rpc.Round.Winning_hand {
		if c == i {
			highlight := canvas.NewRectangle(bundle.Highlight)
			highlight.Resize(size)
			highlight.Move(pos)

			return container.NewWithoutLayout(Cards.P3a, highlight)
		}
	}

	return Cards.P3a

}

// Set third players table card two image
func P3_b(c int) fyne.CanvasObject {
	size := fyne.NewSize(110, 150)
	pos := fyne.NewPos(935, 129)
	Cards.P3b = DisplayCard(c)
	Cards.P3b.Resize(size)
	Cards.P3b.Move(pos)

	for _, i := range rpc.Round.Winning_hand {
		if c == i {
			highlight := canvas.NewRectangle(bundle.Highlight)
			highlight.Resize(size)
			highlight.Move(pos)

			return container.NewWithoutLayout(Cards.P3b, highlight)
		}
	}

	return Cards.P3b
}

// Set fourth players table card one image
func P4_a(c int) fyne.CanvasObject {
	size := fyne.NewSize(110, 150)
	pos := fyne.NewPos(766, 383)
	Cards.P4a = DisplayCard(c)
	Cards.P4a.Resize(size)
	Cards.P4a.Move(pos)

	for _, i := range rpc.Round.Winning_hand {
		if c == i {
			highlight := canvas.NewRectangle(bundle.Highlight)
			highlight.Resize(size)
			highlight.Move(pos)

			return container.NewWithoutLayout(Cards.P4a, highlight)
		}
	}

	return Cards.P4a
}

// Set fourth players table card two image
func P4_b(c int) fyne.CanvasObject {
	size := fyne.NewSize(110, 150)
	pos := fyne.NewPos(818, 383)
	Cards.P4b = DisplayCard(c)
	Cards.P4b.Resize(size)
	Cards.P4b.Move(pos)

	for _, i := range rpc.Round.Winning_hand {
		if c == i {
			highlight := canvas.NewRectangle(bundle.Highlight)
			highlight.Resize(size)
			highlight.Move(pos)

			return container.NewWithoutLayout(Cards.P4b, highlight)
		}
	}

	return Cards.P4b
}

// Set fifth players table card one image
func P5_a(c int) fyne.CanvasObject {
	size := fyne.NewSize(110, 150)
	pos := fyne.NewPos(336, 383)
	Cards.P5a = DisplayCard(c)
	Cards.P5a.Resize(size)
	Cards.P5a.Move(pos)

	for _, i := range rpc.Round.Winning_hand {
		if c == i {
			highlight := canvas.NewRectangle(bundle.Highlight)
			highlight.Resize(size)
			highlight.Move(pos)

			return container.NewWithoutLayout(Cards.P5a, highlight)
		}
	}

	return Cards.P5a
}

// Set fifth players table card two image
func P5_b(c int) fyne.CanvasObject {
	size := fyne.NewSize(110, 150)
	pos := fyne.NewPos(388, 383)
	Cards.P5b = DisplayCard(c)
	Cards.P5b.Resize(size)
	Cards.P5b.Move(pos)

	for _, i := range rpc.Round.Winning_hand {
		if c == i {
			highlight := canvas.NewRectangle(bundle.Highlight)
			highlight.Resize(size)
			highlight.Move(pos)

			return container.NewWithoutLayout(Cards.P5b, highlight)
		}
	}

	return Cards.P5b
}

// Set sixth players table card one image
func P6_a(c int) fyne.CanvasObject {
	size := fyne.NewSize(110, 150)
	pos := fyne.NewPos(65, 269)
	Cards.P6a = DisplayCard(c)
	Cards.P6a.Resize(size)
	Cards.P6a.Move(pos)

	for _, i := range rpc.Round.Winning_hand {
		if c == i {
			highlight := canvas.NewRectangle(bundle.Highlight)
			highlight.Resize(size)
			highlight.Move(pos)

			return container.NewWithoutLayout(Cards.P6a, highlight)
		}
	}

	return Cards.P6a
}

// Set sixth players table card two image
func P6_b(c int) fyne.CanvasObject {
	size := fyne.NewSize(110, 150)
	pos := fyne.NewPos(117, 269)
	Cards.P6b = DisplayCard(c)
	Cards.P6b.Resize(size)
	Cards.P6b.Move(pos)

	for _, i := range rpc.Round.Winning_hand {
		if c == i {
			highlight := canvas.NewRectangle(bundle.Highlight)
			highlight.Resize(size)
			highlight.Move(pos)

			return container.NewWithoutLayout(Cards.P6b, highlight)
		}
	}

	return Cards.P6b
}

// Returns int value for player table cards display.
// If player has no card hash values, no cards will be shown show
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

// Returns a custom card face image
//   - face defines which deck to look for
func CustomCard(c int, face string) *canvas.Image {
	dir := holdero.GetDir()
	mid := "/cards/" + face + "/"
	path := dir + mid + cardEnd(c)

	if holdero.FileExists(path, "dReams") {
		return canvas.NewImageFromFile(path)
	}

	return canvas.NewImageFromImage(nil)
}

// Returns a custom card back image
//   - back defines which back to look for
func CustomBack(back string) *canvas.Image {
	dir := holdero.GetDir()
	post := "/cards/backs/" + back + ".png"
	path := dir + post

	if holdero.FileExists(path, "dReams") {
		return canvas.NewImageFromFile(path)
	}

	return canvas.NewImageFromImage(nil)
}

// Used in CustomCard() to build image path
func cardEnd(card int) (suffix string) {
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

// Set Baccarat player card images
func PlayerCards(c1, c2, c3 int) fyne.CanvasObject {
	card1 := DisplayCard(c1)
	card2 := DisplayCard(c2)
	card3 := DisplayCard(c3)

	card1.Resize(fyne.NewSize(110, 150))
	card1.Move(fyne.NewPos(180, 39))

	card2.Resize(fyne.NewSize(110, 150))
	card2.Move(fyne.NewPos(290, 39))

	card3.Resize(fyne.NewSize(110, 150))
	card3.Move(fyne.NewPos(400, 39))

	return container.NewWithoutLayout(card1, card2, card3)
}

// Set Baccarat banker card images
func BankerCards(c1, c2, c3 int) fyne.CanvasObject {
	card1 := DisplayCard(c1)
	card2 := DisplayCard(c2)
	card3 := DisplayCard(c3)

	card1.Resize(fyne.NewSize(110, 150))
	card1.Move(fyne.NewPos(600, 39))

	card2.Resize(fyne.NewSize(110, 150))
	card2.Move(fyne.NewPos(710, 39))

	card3.Resize(fyne.NewSize(110, 150))
	card3.Move(fyne.NewPos(820, 39))

	return container.NewWithoutLayout(card1, card2, card3)
}

// Create a random suit for baccarat card
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

// Switch for Iluma Tarot card image
func TarotCard(c int) *canvas.Image {
	switch c {
	case 1:
		return canvas.NewImageFromResource(bundle.ResourceIluma1Jpg)
	case 2:
		return canvas.NewImageFromResource(bundle.ResourceIluma2Jpg)
	case 3:
		return canvas.NewImageFromResource(bundle.ResourceIluma3Jpg)
	case 4:
		return canvas.NewImageFromResource(bundle.ResourceIluma4Jpg)
	case 5:
		return canvas.NewImageFromResource(bundle.ResourceIluma5Jpg)
	case 6:
		return canvas.NewImageFromResource(bundle.ResourceIluma6Jpg)
	case 7:
		return canvas.NewImageFromResource(bundle.ResourceIluma7Jpg)
	case 8:
		return canvas.NewImageFromResource(bundle.ResourceIluma8Jpg)
	case 9:
		return canvas.NewImageFromResource(bundle.ResourceIluma9Jpg)
	case 10:
		return canvas.NewImageFromResource(bundle.ResourceIluma10Jpg)
	case 11:
		return canvas.NewImageFromResource(bundle.ResourceIluma11Jpg)
	case 12:
		return canvas.NewImageFromResource(bundle.ResourceIluma12Jpg)
	case 13:
		return canvas.NewImageFromResource(bundle.ResourceIluma13Jpg)
	case 14:
		return canvas.NewImageFromResource(bundle.ResourceIluma14Jpg)
	case 15:
		return canvas.NewImageFromResource(bundle.ResourceIluma15Jpg)
	case 16:
		return canvas.NewImageFromResource(bundle.ResourceIluma16Jpg)
	case 17:
		return canvas.NewImageFromResource(bundle.ResourceIluma17Jpg)
	case 18:
		return canvas.NewImageFromResource(bundle.ResourceIluma18Jpg)
	case 19:
		return canvas.NewImageFromResource(bundle.ResourceIluma19Jpg)
	case 20:
		return canvas.NewImageFromResource(bundle.ResourceIluma20Jpg)
	case 21:
		return canvas.NewImageFromResource(bundle.ResourceIluma21Jpg)
	case 22:
		return canvas.NewImageFromResource(bundle.ResourceIluma22Jpg)
	case 23:
		return canvas.NewImageFromResource(bundle.ResourceIluma23Jpg)
	case 24:
		return canvas.NewImageFromResource(bundle.ResourceIluma24Jpg)
	case 25:
		return canvas.NewImageFromResource(bundle.ResourceIluma25Jpg)
	case 26:
		return canvas.NewImageFromResource(bundle.ResourceIluma26Jpg)
	case 27:
		return canvas.NewImageFromResource(bundle.ResourceIluma27Jpg)
	case 28:
		return canvas.NewImageFromResource(bundle.ResourceIluma28Jpg)
	case 29:
		return canvas.NewImageFromResource(bundle.ResourceIluma29Jpg)
	case 30:
		return canvas.NewImageFromResource(bundle.ResourceIluma30Jpg)
	case 31:
		return canvas.NewImageFromResource(bundle.ResourceIluma31Jpg)
	case 32:
		return canvas.NewImageFromResource(bundle.ResourceIluma32Jpg)
	case 33:
		return canvas.NewImageFromResource(bundle.ResourceIluma33Jpg)
	case 34:
		return canvas.NewImageFromResource(bundle.ResourceIluma34Jpg)
	case 35:
		return canvas.NewImageFromResource(bundle.ResourceIluma35Jpg)
	case 36:
		return canvas.NewImageFromResource(bundle.ResourceIluma36Jpg)
	case 37:
		return canvas.NewImageFromResource(bundle.ResourceIluma37Jpg)
	case 38:
		return canvas.NewImageFromResource(bundle.ResourceIluma38Jpg)
	case 39:
		return canvas.NewImageFromResource(bundle.ResourceIluma39Jpg)
	case 40:
		return canvas.NewImageFromResource(bundle.ResourceIluma40Jpg)
	case 41:
		return canvas.NewImageFromResource(bundle.ResourceIluma41Jpg)
	case 42:
		return canvas.NewImageFromResource(bundle.ResourceIluma42Jpg)
	case 43:
		return canvas.NewImageFromResource(bundle.ResourceIluma43Jpg)
	case 44:
		return canvas.NewImageFromResource(bundle.ResourceIluma44Jpg)
	case 45:
		return canvas.NewImageFromResource(bundle.ResourceIluma45Jpg)
	case 46:
		return canvas.NewImageFromResource(bundle.ResourceIluma46Jpg)
	case 47:
		return canvas.NewImageFromResource(bundle.ResourceIluma47Jpg)
	case 48:
		return canvas.NewImageFromResource(bundle.ResourceIluma48Jpg)
	case 49:
		return canvas.NewImageFromResource(bundle.ResourceIluma49Jpg)
	case 50:
		return canvas.NewImageFromResource(bundle.ResourceIluma50Jpg)
	case 51:
		return canvas.NewImageFromResource(bundle.ResourceIluma51Jpg)
	case 52:
		return canvas.NewImageFromResource(bundle.ResourceIluma52Jpg)
	case 53:
		return canvas.NewImageFromResource(bundle.ResourceIluma53Jpg)
	case 54:
		return canvas.NewImageFromResource(bundle.ResourceIluma54Jpg)
	case 55:
		return canvas.NewImageFromResource(bundle.ResourceIluma55Jpg)
	case 56:
		return canvas.NewImageFromResource(bundle.ResourceIluma56Jpg)
	case 57:
		return canvas.NewImageFromResource(bundle.ResourceIluma57Jpg)
	case 58:
		return canvas.NewImageFromResource(bundle.ResourceIluma58Jpg)
	case 59:
		return canvas.NewImageFromResource(bundle.ResourceIluma59Jpg)
	case 60:
		return canvas.NewImageFromResource(bundle.ResourceIluma60Jpg)
	case 61:
		return canvas.NewImageFromResource(bundle.ResourceIluma61Jpg)
	case 62:
		return canvas.NewImageFromResource(bundle.ResourceIluma62Jpg)
	case 63:
		return canvas.NewImageFromResource(bundle.ResourceIluma63Jpg)
	case 64:
		return canvas.NewImageFromResource(bundle.ResourceIluma64Jpg)
	case 65:
		return canvas.NewImageFromResource(bundle.ResourceIluma65Jpg)
	case 66:
		return canvas.NewImageFromResource(bundle.ResourceIluma66Jpg)
	case 67:
		return canvas.NewImageFromResource(bundle.ResourceIluma67Jpg)
	case 68:
		return canvas.NewImageFromResource(bundle.ResourceIluma68Jpg)
	case 69:
		return canvas.NewImageFromResource(bundle.ResourceIluma69Jpg)
	case 70:
		return canvas.NewImageFromResource(bundle.ResourceIluma70Jpg)
	case 71:
		return canvas.NewImageFromResource(bundle.ResourceIluma71Jpg)
	case 72:
		return canvas.NewImageFromResource(bundle.ResourceIluma72Jpg)
	case 73:
		return canvas.NewImageFromResource(bundle.ResourceIluma73Jpg)
	case 74:
		return canvas.NewImageFromResource(bundle.ResourceIluma74Jpg)
	case 75:
		return canvas.NewImageFromResource(bundle.ResourceIluma75Jpg)
	case 76:
		return canvas.NewImageFromResource(bundle.ResourceIluma76Jpg)
	case 77:
		return canvas.NewImageFromResource(bundle.ResourceIluma77Jpg)
	case 78:
		return canvas.NewImageFromResource(bundle.ResourceIluma78Jpg)
	default:
		return canvas.NewImageFromResource(bundle.ResourceIluma81Png)
	}
}

// Main switch used to display playing card images
func DisplayCard(card int) *canvas.Image {
	if !holdero.Settings.Shared || rpc.Round.ID == 1 {
		if card == 99 {
			return canvas.NewImageFromImage(nil)
		}

		if card > 0 {
			i := holdero.Settings.FaceSelect.SelectedIndex()
			switch i {
			case -1:
				return canvas.NewImageFromResource(DisplayLightCard(card))
			case 0:
				return canvas.NewImageFromResource(DisplayLightCard(card))
			case 1:
				return canvas.NewImageFromResource(DisplayDarkCard(card))
			default:
				return CustomCard(card, holdero.Settings.Faces)
			}
		}

		i := holdero.Settings.BackSelect.SelectedIndex()
		switch i {
		case -1:
			return canvas.NewImageFromResource(bundle.ResourceBack1Png)
		case 0:
			return canvas.NewImageFromResource(bundle.ResourceBack1Png)
		case 1:
			return canvas.NewImageFromResource(bundle.ResourceBack2Png)
		default:
			return CustomBack(holdero.Settings.Backs)
		}

	} else {
		if card == 99 {
			return canvas.NewImageFromImage(nil)
		} else if card > 0 {
			return CustomCard(card, rpc.Round.Face)
		} else {
			return CustomBack(rpc.Round.Back)
		}
	}
}

// Switch for standard light deck image
func DisplayLightCard(card int) fyne.Resource {
	if card > 0 && card < 53 {
		switch card {
		case 1:
			return bundle.ResourceLightcard1Png
		case 2:
			return bundle.ResourceLightcard2Png
		case 3:
			return bundle.ResourceLightcard3Png
		case 4:
			return bundle.ResourceLightcard4Png
		case 5:
			return bundle.ResourceLightcard5Png
		case 6:
			return bundle.ResourceLightcard6Png
		case 7:
			return bundle.ResourceLightcard7Png
		case 8:
			return bundle.ResourceLightcard8Png
		case 9:
			return bundle.ResourceLightcard9Png
		case 10:
			return bundle.ResourceLightcard10Png
		case 11:
			return bundle.ResourceLightcard11Png
		case 12:
			return bundle.ResourceLightcard12Png
		case 13:
			return bundle.ResourceLightcard13Png
		case 14:
			return bundle.ResourceLightcard14Png
		case 15:
			return bundle.ResourceLightcard15Png
		case 16:
			return bundle.ResourceLightcard16Png
		case 17:
			return bundle.ResourceLightcard17Png
		case 18:
			return bundle.ResourceLightcard18Png
		case 19:
			return bundle.ResourceLightcard19Png
		case 20:
			return bundle.ResourceLightcard20Png
		case 21:
			return bundle.ResourceLightcard21Png
		case 22:
			return bundle.ResourceLightcard22Png
		case 23:
			return bundle.ResourceLightcard23Png
		case 24:
			return bundle.ResourceLightcard24Png
		case 25:
			return bundle.ResourceLightcard25Png
		case 26:
			return bundle.ResourceLightcard26Png
		case 27:
			return bundle.ResourceLightcard27Png
		case 28:
			return bundle.ResourceLightcard28Png
		case 29:
			return bundle.ResourceLightcard29Png
		case 30:
			return bundle.ResourceLightcard30Png
		case 31:
			return bundle.ResourceLightcard31Png
		case 32:
			return bundle.ResourceLightcard32Png
		case 33:
			return bundle.ResourceLightcard33Png
		case 34:
			return bundle.ResourceLightcard34Png
		case 35:
			return bundle.ResourceLightcard35Png
		case 36:
			return bundle.ResourceLightcard36Png
		case 37:
			return bundle.ResourceLightcard37Png
		case 38:
			return bundle.ResourceLightcard38Png
		case 39:
			return bundle.ResourceLightcard39Png
		case 40:
			return bundle.ResourceLightcard40Png
		case 41:
			return bundle.ResourceLightcard41Png
		case 42:
			return bundle.ResourceLightcard42Png
		case 43:
			return bundle.ResourceLightcard43Png
		case 44:
			return bundle.ResourceLightcard44Png
		case 45:
			return bundle.ResourceLightcard45Png
		case 46:
			return bundle.ResourceLightcard46Png
		case 47:
			return bundle.ResourceLightcard47Png
		case 48:
			return bundle.ResourceLightcard48Png
		case 49:
			return bundle.ResourceLightcard49Png
		case 50:
			return bundle.ResourceLightcard50Png
		case 51:
			return bundle.ResourceLightcard51Png
		case 52:
			return bundle.ResourceLightcard52Png
		}
	}
	return nil
}

// Switch for standard dark deck image
func DisplayDarkCard(card int) fyne.Resource {
	if card > 0 && card < 53 {
		switch card {
		case 1:
			return bundle.ResourceDarkcard1Png
		case 2:
			return bundle.ResourceDarkcard2Png
		case 3:
			return bundle.ResourceDarkcard3Png
		case 4:
			return bundle.ResourceDarkcard4Png
		case 5:
			return bundle.ResourceDarkcard5Png
		case 6:
			return bundle.ResourceDarkcard6Png
		case 7:
			return bundle.ResourceDarkcard7Png
		case 8:
			return bundle.ResourceDarkcard8Png
		case 9:
			return bundle.ResourceDarkcard9Png
		case 10:
			return bundle.ResourceDarkcard10Png
		case 11:
			return bundle.ResourceDarkcard11Png
		case 12:
			return bundle.ResourceDarkcard12Png
		case 13:
			return bundle.ResourceDarkcard13Png
		case 14:
			return bundle.ResourceDarkcard14Png
		case 15:
			return bundle.ResourceDarkcard15Png
		case 16:
			return bundle.ResourceDarkcard16Png
		case 17:
			return bundle.ResourceDarkcard17Png
		case 18:
			return bundle.ResourceDarkcard18Png
		case 19:
			return bundle.ResourceDarkcard19Png
		case 20:
			return bundle.ResourceDarkcard20Png
		case 21:
			return bundle.ResourceDarkcard21Png
		case 22:
			return bundle.ResourceDarkcard22Png
		case 23:
			return bundle.ResourceDarkcard23Png
		case 24:
			return bundle.ResourceDarkcard24Png
		case 25:
			return bundle.ResourceDarkcard25Png
		case 26:
			return bundle.ResourceDarkcard26Png
		case 27:
			return bundle.ResourceDarkcard27Png
		case 28:
			return bundle.ResourceDarkcard28Png
		case 29:
			return bundle.ResourceDarkcard29Png
		case 30:
			return bundle.ResourceDarkcard30Png
		case 31:
			return bundle.ResourceDarkcard31Png
		case 32:
			return bundle.ResourceDarkcard32Png
		case 33:
			return bundle.ResourceDarkcard33Png
		case 34:
			return bundle.ResourceDarkcard34Png
		case 35:
			return bundle.ResourceDarkcard35Png
		case 36:
			return bundle.ResourceDarkcard36Png
		case 37:
			return bundle.ResourceDarkcard37Png
		case 38:
			return bundle.ResourceDarkcard38Png
		case 39:
			return bundle.ResourceDarkcard39Png
		case 40:
			return bundle.ResourceDarkcard40Png
		case 41:
			return bundle.ResourceDarkcard41Png
		case 42:
			return bundle.ResourceDarkcard42Png
		case 43:
			return bundle.ResourceDarkcard43Png
		case 44:
			return bundle.ResourceDarkcard44Png
		case 45:
			return bundle.ResourceDarkcard45Png
		case 46:
			return bundle.ResourceDarkcard46Png
		case 47:
			return bundle.ResourceDarkcard47Png
		case 48:
			return bundle.ResourceDarkcard48Png
		case 49:
			return bundle.ResourceDarkcard49Png
		case 50:
			return bundle.ResourceDarkcard50Png
		case 51:
			return bundle.ResourceDarkcard51Png
		case 52:
			return bundle.ResourceDarkcard52Png
		}
	}
	return nil
}
