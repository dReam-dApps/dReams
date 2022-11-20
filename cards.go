package main

import (
	"SixofClubsss/dReams/rpc"
	"SixofClubsss/dReams/table"
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

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
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(165, 225))
	card.Move(fyne.NewPos(w-335, h-350))

	return card
}

func Hole_2(c int, w, h float32) fyne.CanvasObject {
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(165, 225))
	card.Move(fyne.NewPos(w-275, h-350))

	return card
}

func Flop_1(c int) fyne.CanvasObject {
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(110, 150))
	card.Move(fyne.NewPos(257, 185))

	return card
}

func Flop_2(c int) fyne.CanvasObject {
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(110, 150))
	card.Move(fyne.NewPos(377, 185))

	return card
}

func Flop_3(c int) fyne.CanvasObject {
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(110, 150))
	card.Move(fyne.NewPos(497, 185))

	return card
}

func River(c int) fyne.CanvasObject {
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(110, 150))
	card.Move(fyne.NewPos(617, 185))

	return card
}

func Turn(c int) fyne.CanvasObject {
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(110, 150))
	card.Move(fyne.NewPos(737, 185))

	return card
}

func P1_a(c int) fyne.CanvasObject {
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(110, 150))
	card.Move(fyne.NewPos(190, 10))

	return card
}

func P1_b(c int) fyne.CanvasObject {
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(110, 150))
	card.Move(fyne.NewPos(242, 10))

	return card
}

func P2_a(c int) fyne.CanvasObject {
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(110, 150))
	card.Move(fyne.NewPos(614, 10))

	return card
}

func P2_b(c int) fyne.CanvasObject {
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(110, 150))
	card.Move(fyne.NewPos(666, 10))

	return card
}

func P3_a(c int) fyne.CanvasObject {
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(110, 150))
	card.Move(fyne.NewPos(886, 115))

	return card

}

func P3_b(c int) fyne.CanvasObject {
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(110, 150))
	card.Move(fyne.NewPos(938, 115))

	return card
}

func P4_a(c int) fyne.CanvasObject {
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(110, 150))
	card.Move(fyne.NewPos(766, 361))

	return card
}

func P4_b(c int) fyne.CanvasObject {
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(110, 150))
	card.Move(fyne.NewPos(818, 361))

	return card
}

func P5_a(c int) fyne.CanvasObject {
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(110, 150))
	card.Move(fyne.NewPos(336, 361))

	return card
}

func P5_b(c int) fyne.CanvasObject {
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(110, 150))
	card.Move(fyne.NewPos(388, 361))

	return card
}

func P6_a(c int) fyne.CanvasObject {
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(110, 150))
	card.Move(fyne.NewPos(63, 254))

	return card
}

func P6_b(c int) fyne.CanvasObject {
	card := DisplayCard(c)
	card.Resize(fyne.NewSize(110, 150))
	card.Move(fyne.NewPos(115, 254))

	return card
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
