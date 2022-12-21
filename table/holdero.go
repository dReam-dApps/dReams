package table

import (
	"image/color"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/SixofClubsss/dReams/rpc"

	"fyne.io/fyne/driver/mobile"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/widget"
)

type settings struct {
	Faces     string
	Backs     string
	Theme     string
	Avatar    string
	FaceUrl   string
	BackUrl   string
	AvatarUrl string
	ThemeUrl  string
	Shared    bool

	P1_avatar_url string
	P2_avatar_url string
	P3_avatar_url string
	P4_avatar_url string
	P5_avatar_url string
	P6_avatar_url string

	FaceSelect   *widget.Select
	BackSelect   *widget.Select
	ThemeSelect  *widget.Select
	AvatarSelect *widget.Select
	SharedOn     *widget.RadioGroup
	ThemeImg     canvas.Image
}

var Settings settings
var Poker_name string

func InitTableSettings() {
	rpc.Signal.Startup = true
	rpc.Bacc.Display = true
	rpc.Tarot.Display = true
	rpc.Times.Delay = 30
	rpc.Times.Kick = 0
	Settings.Faces = "light/"
	Settings.Backs = "back1.png"
	Settings.Avatar = "None"
	Settings.FaceUrl = ""
	Settings.BackUrl = ""
	Settings.AvatarUrl = ""
}

func GetDir() string {
	pre, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return ""
	}

	return pre
}

func Player1_label(a, f, t fyne.Resource) fyne.CanvasObject {
	var name fyne.CanvasObject
	var avatar fyne.CanvasObject
	var frame fyne.CanvasObject
	var out fyne.CanvasObject
	if rpc.Signal.In1 {
		if rpc.Display.Turn == "1" {
			name = canvas.NewText(rpc.Round.P1_name, color.RGBA{105, 90, 205, 210})
		} else {
			name = canvas.NewText(rpc.Round.P1_name, color.White)
		}
	} else {
		name = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	if a != nil && rpc.Signal.In1 {
		if rpc.Round.P1_url != "" {
			avatar = &Shared.P1_avatar
			if rpc.Display.Turn == "1" {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		} else {
			avatar = canvas.NewImageFromResource(a)
			if rpc.Display.Turn == "1" {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		}
	} else {
		avatar = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
		frame = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	if rpc.Signal.Out1 {
		out = canvas.NewText("Sitting out", color.White)
		out.Resize(fyne.NewSize(100, 25))
		out.Move(fyne.NewPos(253, 45))
	} else {
		out = canvas.NewText("", color.RGBA{0, 0, 0, 0})
	}

	name.Resize(fyne.NewSize(100, 25))
	name.Move(fyne.NewPos(253, 20))

	avatar.Resize(fyne.NewSize(74, 74))
	avatar.Move(fyne.NewPos(359, 22))

	frame.Resize(fyne.NewSize(78, 78))
	frame.Move(fyne.NewPos(357, 20))

	p := container.NewWithoutLayout(name, out, avatar, frame)

	return p
}

func Player2_label(a, f, t fyne.Resource) fyne.CanvasObject {
	var name fyne.CanvasObject
	var avatar fyne.CanvasObject
	var frame fyne.CanvasObject
	if rpc.Signal.In2 {
		if rpc.Display.Turn == "2" {
			name = canvas.NewText(rpc.Round.P2_name, color.RGBA{105, 90, 205, 210})
		} else {
			name = canvas.NewText(rpc.Round.P2_name, color.White)
		}
	} else {
		name = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	if a != nil && rpc.Signal.In2 {
		if rpc.Round.P2_url != "" {
			avatar = &Shared.P2_avatar
			if rpc.Display.Turn == "2" {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		} else {
			avatar = canvas.NewImageFromResource(a)
			if rpc.Display.Turn == "2" {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		}
	} else {
		avatar = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
		frame = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	name.Resize(fyne.NewSize(100, 25))
	name.Move(fyne.NewPos(678, 20))

	avatar.Resize(fyne.NewSize(74, 74))
	avatar.Move(fyne.NewPos(782, 22))

	frame.Resize(fyne.NewSize(78, 78))
	frame.Move(fyne.NewPos(780, 20))

	p := container.NewWithoutLayout(name, avatar, frame)

	return p
}

func Player3_label(a, f, t fyne.Resource) fyne.CanvasObject {
	var name fyne.CanvasObject
	var avatar fyne.CanvasObject
	var frame fyne.CanvasObject
	if rpc.Signal.In3 {
		if rpc.Display.Turn == "3" {
			name = canvas.NewText(rpc.Round.P3_name, color.RGBA{105, 90, 205, 210})
		} else {
			name = canvas.NewText(rpc.Round.P3_name, color.White)
		}
	} else {
		name = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	if a != nil && rpc.Signal.In3 {
		if rpc.Round.P3_url != "" {
			avatar = &Shared.P3_avatar
			if rpc.Display.Turn == "3" {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		} else {
			avatar = canvas.NewImageFromResource(a)
			if rpc.Display.Turn == "3" {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		}
	} else {
		avatar = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
		frame = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	name.Resize(fyne.NewSize(100, 25))
	name.Move(fyne.NewPos(892, 310))

	avatar.Resize(fyne.NewSize(74, 74))
	avatar.Move(fyne.NewPos(997, 312))

	frame.Resize(fyne.NewSize(78, 78))
	frame.Move(fyne.NewPos(995, 310))

	p := container.NewWithoutLayout(name, avatar, frame)

	return p
}

func Player4_label(a, f, t fyne.Resource) fyne.CanvasObject {
	var name fyne.CanvasObject
	var avatar fyne.CanvasObject
	var frame fyne.CanvasObject
	if rpc.Signal.In4 {
		if rpc.Display.Turn == "4" {
			name = canvas.NewText(rpc.Round.P4_name, color.RGBA{105, 90, 205, 210})
		} else {
			name = canvas.NewText(rpc.Round.P4_name, color.White)
		}
	} else {
		name = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	if a != nil && rpc.Signal.In4 {
		if rpc.Round.P4_url != "" {
			avatar = &Shared.P4_avatar
			if rpc.Display.Turn == "4" {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		} else {
			avatar = canvas.NewImageFromResource(a)
			if rpc.Display.Turn == "4" {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		}
	} else {
		avatar = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
		frame = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	name.Resize(fyne.NewSize(100, 25))
	name.Move(fyne.NewPos(765, 555))

	avatar.Resize(fyne.NewSize(74, 74))
	avatar.Move(fyne.NewPos(686, 505))

	frame.Resize(fyne.NewSize(78, 78))
	frame.Move(fyne.NewPos(684, 503))

	p := container.NewWithoutLayout(name, avatar, frame)

	return p
}

func Player5_label(a, f, t fyne.Resource) fyne.CanvasObject {
	var name fyne.CanvasObject
	var avatar fyne.CanvasObject
	var frame fyne.CanvasObject
	if rpc.Signal.In5 {
		if rpc.Display.Turn == "5" {
			name = canvas.NewText(rpc.Round.P5_name, color.RGBA{105, 90, 205, 210})
		} else {
			name = canvas.NewText(rpc.Round.P5_name, color.White)
		}
	} else {
		name = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	if a != nil && rpc.Signal.In5 {
		if rpc.Round.P5_url != "" {
			avatar = &Shared.P5_avatar
			if rpc.Display.Turn == "5" {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		} else {
			avatar = canvas.NewImageFromResource(a)
			if rpc.Display.Turn == "5" {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		}
	} else {
		avatar = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
		frame = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	name.Resize(fyne.NewSize(100, 25))
	name.Move(fyne.NewPos(335, 555))

	avatar.Resize(fyne.NewSize(74, 74))
	avatar.Move(fyne.NewPos(257, 505))

	frame.Resize(fyne.NewSize(78, 78))
	frame.Move(fyne.NewPos(255, 503))

	p := container.NewWithoutLayout(name, avatar, frame)

	return p
}

func Player6_label(a, f, t fyne.Resource) fyne.CanvasObject {
	var name fyne.CanvasObject
	var avatar fyne.CanvasObject
	var frame fyne.CanvasObject
	if rpc.Signal.In6 {
		if rpc.Display.Turn == "6" {
			name = canvas.NewText(rpc.Round.P6_name, color.RGBA{105, 90, 205, 210})
		} else {
			name = canvas.NewText(rpc.Round.P6_name, color.White)
		}
	} else {
		name = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	if a != nil && rpc.Signal.In6 {
		if rpc.Round.P6_url != "" {
			avatar = &Shared.P6_avatar
			if rpc.Display.Turn == "6" {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		} else {
			avatar = canvas.NewImageFromResource(a)
			if rpc.Display.Turn == "6" {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		}
	} else {
		avatar = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
		frame = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	name.Resize(fyne.NewSize(100, 27))
	name.Move(fyne.NewPos(121, 261))

	avatar.Resize(fyne.NewSize(74, 74))
	avatar.Move(fyne.NewPos(42, 212))

	frame.Resize(fyne.NewSize(78, 78))
	frame.Move(fyne.NewPos(40, 210))

	p := container.NewWithoutLayout(name, avatar, frame)

	return p
}

func HolderoTable(img fyne.Resource) fyne.CanvasObject {
	table := canvas.NewImageFromResource(img)
	table.Resize(fyne.NewSize(1100, 600))
	table.Move(fyne.NewPos(5, 0))

	return table
}

type tableWidgets struct {
	Sit      *widget.Button
	Leave    *widget.Button
	Deal     *widget.Button
	Bet      *widget.Button
	Check    *widget.Button
	BetEntry *betAmt

	Bacc_actions *fyne.Container

	Dreams     *widget.Button
	Dero       *widget.Button
	DEntry     *dReamsAmt
	Tournament *widget.Button

	Higher         *widget.Button
	Lower          *widget.Button
	Change         *widget.Button
	Remove         *widget.Button
	NameEntry      *widget.Entry
	Prediction_box *fyne.Container
	P_contract     *widget.SelectEntry

	Game_select  *widget.Select
	Game_options []string
	Multi        *widget.RadioGroup
	ButtonA      *widget.Button
	ButtonB      *widget.Button
	Sports_box   *fyne.Container
	S_contract   *widget.SelectEntry
}

var Actions tableWidgets

func holderoButtonBuffer() {
	Actions.Sit.Hide()
	Actions.Leave.Hide()
	Actions.Deal.Hide()
	Actions.Bet.Hide()
	Actions.Check.Hide()
	Actions.BetEntry.Hide()
	rpc.Display.Res = ""
	rpc.Signal.Clicked = true
	rpc.Signal.CHeight = rpc.StringToInt(rpc.Wallet.Height)
}

func CheckNames(seats string) bool {
	if rpc.Round.ID == 1 {
		return true
	}

	switch seats {
	case "2":
		if Poker_name == rpc.Round.P1_name {
			log.Println("Name already used")
			return false
		}
		return true
	case "3":
		if Poker_name == rpc.Round.P1_name || Poker_name == rpc.Round.P2_name || Poker_name == rpc.Round.P3_name {
			log.Println("Name already used")
			return false
		}
		return true
	case "4":
		if Poker_name == rpc.Round.P1_name || Poker_name == rpc.Round.P2_name || Poker_name == rpc.Round.P3_name || Poker_name == rpc.Round.P4_name {
			log.Println("Name already used")
			return false
		}
		return true
	case "5":
		if Poker_name == rpc.Round.P1_name || Poker_name == rpc.Round.P2_name || Poker_name == rpc.Round.P3_name || Poker_name == rpc.Round.P4_name || Poker_name == rpc.Round.P5_name {
			log.Println("Name already used")
			return false
		}
		return true
	case "6":
		if Poker_name == rpc.Round.P1_name || Poker_name == rpc.Round.P2_name || Poker_name == rpc.Round.P3_name || Poker_name == rpc.Round.P4_name || Poker_name == rpc.Round.P5_name || Poker_name == rpc.Round.P6_name {
			log.Println("Name already used")
			return false
		}
		return true
	default:
		return false
	}
}

func SitButton() fyne.Widget {
	Actions.Sit = widget.NewButton("Sit Down", func() {
		if Poker_name != "" {
			if CheckNames(rpc.Display.Seats) {
				rpc.SitDown(Poker_name, Settings.AvatarUrl)
				holderoButtonBuffer()
			}
		} else {
			log.Println("Pick a name")
		}
	})

	Actions.Sit.Hide()

	return Actions.Sit
}

func LeaveButton() fyne.Widget {
	Actions.Leave = widget.NewButton("Leave", func() {
		rpc.Leave()
		holderoButtonBuffer()
	})

	Actions.Leave.Hide()

	return Actions.Leave
}

func DealHandButton() fyne.Widget {
	Actions.Deal = widget.NewButton("Deal Hand", func() {
		rpc.DealHand()
		holderoButtonBuffer()
	})

	Actions.Deal.Hide()

	return Actions.Deal
}

type betAmt struct {
	NumericalEntry
}

func (e *betAmt) TypedKey(k *fyne.KeyEvent) {
	switch k.Name {
	case fyne.KeyUp:
		if f, err := strconv.ParseFloat(e.Entry.Text, 64); err == nil {
			e.Entry.SetText(strconv.FormatFloat(float64(f+0.1), 'f', 1, 64))
		}
	case fyne.KeyDown:
		if f, err := strconv.ParseFloat(e.Entry.Text, 64); err == nil {
			if f >= 0.1 {
				e.Entry.SetText(strconv.FormatFloat(float64(f-0.1), 'f', 1, 64))
			}
		}
	}
	e.Entry.TypedKey(k)
}

func BetAmount() fyne.CanvasObject {
	Actions.BetEntry = &betAmt{}
	Actions.BetEntry.ExtendBaseWidget(Actions.BetEntry)
	Actions.BetEntry.Enable()
	if Actions.BetEntry.Text == "" {
		Actions.BetEntry.SetText("0.0")
	}
	Actions.BetEntry.Validator = validation.NewRegexp(`[^0]\d{1,}\.\d{1,1}|\d{0,1}\.\d{0,1}$`, "Format Not Valid")
	Actions.BetEntry.OnChanged = func(s string) {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			if rpc.Signal.PlacedBet {
				Actions.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Raised)/100000, 'f', 1, 64))
				if Actions.BetEntry.Validate() != nil {
					Actions.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Raised)/100000, 'f', 1, 64))
				}
			} else {

				if rpc.Round.Wager > 0 {
					if rpc.Round.Raised > 0 {
						if rpc.Signal.PlacedBet {
							Actions.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Raised)/100000, 'f', 1, 64))
						} else {
							Actions.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Wager)/100000, 'f', 1, 64))
						}
						if Actions.BetEntry.Validate() != nil {
							if rpc.Signal.PlacedBet {
								Actions.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Raised)/100000, 'f', 1, 64))
							} else {
								Actions.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Wager)/100000, 'f', 1, 64))
							}
						}
					} else {

						if f < float64(rpc.Round.Wager)/100000 {
							Actions.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Wager)/100000, 'f', 1, 64))
						}

						if Actions.BetEntry.Validate() != nil {
							float := f * 100000
							if uint64(float)%10000 == 0 {
								Actions.BetEntry.SetText(strconv.FormatFloat(roundFloat(f, 1), 'f', 1, 64))
							} else if Actions.BetEntry.Validate() != nil {
								Actions.BetEntry.SetText(strconv.FormatFloat(roundFloat(f, 1), 'f', 1, 64))
							}
						}
					}
				} else {

					if rpc.Signal.Daemon {
						float := f * 100000
						if uint64(float)%10000 == 0 {
							Actions.BetEntry.SetText(strconv.FormatFloat(roundFloat(f, 1), 'f', 1, 64))
						} else if Actions.BetEntry.Validate() != nil {
							Actions.BetEntry.SetText(strconv.FormatFloat(roundFloat(f, 1), 'f', 1, 64))
						}

						if rpc.Round.Ante > 0 {
							if f < float64(rpc.Round.Ante)/100000 {
								Actions.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Ante)/100000, 'f', 1, 64))
							}

							if Actions.BetEntry.Validate() != nil {
								Actions.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Ante)/100000, 'f', 1, 64))
							}

						} else {
							if f < float64(rpc.Round.BB)/100000 {
								Actions.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.BB)/100000, 'f', 1, 64))
							}

							if Actions.BetEntry.Validate() != nil {
								Actions.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.BB)/100000, 'f', 1, 64))
							}
						}
					}
				}
			}
		} else {
			log.Println(err)
		}
	}

	amt_box := container.NewHScroll(Actions.BetEntry)
	amt_box.SetMinSize(fyne.NewSize(100, 40))
	Actions.BetEntry.Hide()

	return amt_box

}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func BetButton() fyne.Widget {
	Actions.Bet = widget.NewButton("Bet", func() {
		if Actions.BetEntry.Validate() == nil {
			rpc.Bet(Actions.BetEntry.Text)
			rpc.Signal.Bet = true
			holderoButtonBuffer()
		}
	})

	Actions.Bet.Hide()

	return Actions.Bet
}

func CheckButton() fyne.Widget {
	Actions.Check = widget.NewButton("Check", func() {
		rpc.Check()
		rpc.Signal.Bet = true
		holderoButtonBuffer()

	})

	Actions.Check.Hide()

	return Actions.Check
}

type NumericalEntry struct {
	widget.Entry
}

func NilNumericalEntry() *NumericalEntry {
	entry := &NumericalEntry{}
	entry.ExtendBaseWidget(entry)

	return entry
}

func (e *NumericalEntry) TypedRune(r rune) {
	if (r >= '0' && r <= '9') || r == '.' {
		e.Entry.TypedRune(r)
	}
}

func (e *NumericalEntry) TypedShortcut(shortcut fyne.Shortcut) {
	paste, ok := shortcut.(*fyne.ShortcutPaste)
	if !ok {
		e.Entry.TypedShortcut(shortcut)
		return
	}

	content := paste.Clipboard.Content()
	if _, err := strconv.ParseFloat(content, 64); err == nil {
		e.Entry.TypedShortcut(shortcut)
	}
}

func (e *NumericalEntry) Keyboard() mobile.KeyboardType {
	return mobile.NumberKeyboard
}
