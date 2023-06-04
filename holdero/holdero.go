package holdero

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/dwidget"
	"github.com/SixofClubsss/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type settings struct {
	Faces      string
	Backs      string
	Theme      string
	Avatar     string
	FaceUrl    string
	BackUrl    string
	AvatarUrl  string
	ThemeUrl   string
	Shared     bool
	Auto_check bool
	Auto_deal  bool

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
	Tools        *widget.Button
	SharedOn     *widget.RadioGroup
	ThemeImg     canvas.Image
}

type tableObjects struct {
	Sit        *widget.Button
	Leave      *widget.Button
	Deal       *widget.Button
	Bet        *widget.Button
	Check      *widget.Button
	Tournament *widget.Button
	BetEntry   *dwidget.DeroAmts
	Warning    *fyne.Container
}

type swapObjects struct {
	Dreams *widget.Button
	Dero   *widget.Button
	DEntry *dwidget.DeroAmts
}

var Swap swapObjects
var Table tableObjects
var Settings settings
var Poker_name string

func InitTableSettings() {
	rpc.Signal.Startup = true
	rpc.Bacc.Display = true
	rpc.Tarot.Display = true
	rpc.Times.Delay = 30
	rpc.Times.Kick = 0
	rpc.Odds.Run = false
	Settings.Faces = "light/"
	Settings.Backs = "back1.png"
	Settings.Avatar = "None"
	Settings.FaceUrl = ""
	Settings.BackUrl = ""
	Settings.AvatarUrl = ""
	Settings.Auto_deal = false
	Settings.Auto_check = false
	autoBetDefault()
}

// Get current working directory path for prefix
func GetDir() string {
	pre, err := os.Getwd()
	if err != nil {
		log.Println("[GetDir]", err)
		return ""
	}

	return pre
}

// Table owner name and avatar objects
//   - Pass a and f as avatar and its frame resource, shared avatar is set here if image exists
//   - Pass t for player's turn frame resource
func Player1_label(a, f, t fyne.Resource) fyne.CanvasObject {
	var name fyne.CanvasObject
	var avatar fyne.CanvasObject
	var frame fyne.CanvasObject
	var out fyne.CanvasObject
	if rpc.Signal.In1 {
		if rpc.Round.Turn == 1 {
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
			if rpc.Round.Turn == 1 {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		} else {
			avatar = canvas.NewImageFromResource(a)
			if rpc.Round.Turn == 1 {
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
	name.Move(fyne.NewPos(242, 20))

	avatar.Resize(fyne.NewSize(74, 74))
	avatar.Move(fyne.NewPos(359, 50))

	frame.Resize(fyne.NewSize(78, 78))
	frame.Move(fyne.NewPos(357, 48))

	return container.NewWithoutLayout(name, out, avatar, frame)
}

// Player 2 name and avatar objects
//   - Pass a and f as avatar and its frame resource, shared avatar is set here if image exists
//   - Pass t for player's turn frame resource
func Player2_label(a, f, t fyne.Resource) fyne.CanvasObject {
	var name fyne.CanvasObject
	var avatar fyne.CanvasObject
	var frame fyne.CanvasObject
	if rpc.Signal.In2 {
		if rpc.Round.Turn == 2 {
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
			if rpc.Round.Turn == 2 {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		} else {
			avatar = canvas.NewImageFromResource(a)
			if rpc.Round.Turn == 2 {
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
	name.Move(fyne.NewPos(667, 20))

	avatar.Resize(fyne.NewSize(74, 74))
	avatar.Move(fyne.NewPos(782, 50))

	frame.Resize(fyne.NewSize(78, 78))
	frame.Move(fyne.NewPos(780, 48))

	return container.NewWithoutLayout(name, avatar, frame)
}

// Player 3 name and avatar objects
//   - Pass a and f as avatar and its frame resource, shared avatar is set here if image exists
//   - Pass t for player's turn frame resource
func Player3_label(a, f, t fyne.Resource) fyne.CanvasObject {
	var name fyne.CanvasObject
	var avatar fyne.CanvasObject
	var frame fyne.CanvasObject
	if rpc.Signal.In3 {
		if rpc.Round.Turn == 3 {
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
			if rpc.Round.Turn == 3 {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		} else {
			avatar = canvas.NewImageFromResource(a)
			if rpc.Round.Turn == 3 {
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
	name.Move(fyne.NewPos(889, 300))

	avatar.Resize(fyne.NewSize(74, 74))
	avatar.Move(fyne.NewPos(987, 327))

	frame.Resize(fyne.NewSize(78, 78))
	frame.Move(fyne.NewPos(985, 325))

	return container.NewWithoutLayout(name, avatar, frame)
}

// Player 4 name and avatar objects
//   - Pass a and f as avatar and its frame resource, shared avatar is set here if image exists
//   - Pass t for player's turn frame resource
func Player4_label(a, f, t fyne.Resource) fyne.CanvasObject {
	var name fyne.CanvasObject
	var avatar fyne.CanvasObject
	var frame fyne.CanvasObject
	if rpc.Signal.In4 {
		if rpc.Round.Turn == 4 {
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
			if rpc.Round.Turn == 4 {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		} else {
			avatar = canvas.NewImageFromResource(a)
			if rpc.Round.Turn == 4 {
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
	name.Move(fyne.NewPos(688, 555))

	avatar.Resize(fyne.NewSize(74, 74))
	avatar.Move(fyne.NewPos(686, 480))

	frame.Resize(fyne.NewSize(78, 78))
	frame.Move(fyne.NewPos(684, 478))

	return container.NewWithoutLayout(name, avatar, frame)
}

// Player 5 name and avatar objects
//   - Pass a and f as avatar and its frame resource, shared avatar is set here if image exists
//   - Pass t for player's turn frame resource
func Player5_label(a, f, t fyne.Resource) fyne.CanvasObject {
	var name fyne.CanvasObject
	var avatar fyne.CanvasObject
	var frame fyne.CanvasObject
	if rpc.Signal.In5 {
		if rpc.Round.Turn == 5 {
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
			if rpc.Round.Turn == 5 {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		} else {
			avatar = canvas.NewImageFromResource(a)
			if rpc.Round.Turn == 5 {
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
	name.Move(fyne.NewPos(258, 555))

	avatar.Resize(fyne.NewSize(74, 74))
	avatar.Move(fyne.NewPos(257, 480))

	frame.Resize(fyne.NewSize(78, 78))
	frame.Move(fyne.NewPos(255, 478))

	return container.NewWithoutLayout(name, avatar, frame)
}

// Player 6 name and avatar objects
//   - Pass a and f as avatar and its frame resource, shared avatar is set here if image exists
//   - Pass t for player's turn frame resource
func Player6_label(a, f, t fyne.Resource) fyne.CanvasObject {
	var name fyne.CanvasObject
	var avatar fyne.CanvasObject
	var frame fyne.CanvasObject
	if rpc.Signal.In6 {
		if rpc.Round.Turn == 6 {
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
			if rpc.Round.Turn == 6 {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		} else {
			avatar = canvas.NewImageFromResource(a)
			if rpc.Round.Turn == 6 {
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
	name.Move(fyne.NewPos(56, 267))

	avatar.Resize(fyne.NewSize(74, 74))
	avatar.Move(fyne.NewPos(55, 193))

	frame.Resize(fyne.NewSize(78, 78))
	frame.Move(fyne.NewPos(53, 191))

	return container.NewWithoutLayout(name, avatar, frame)
}

// Set Holdero table image resource
func HolderoTable(img fyne.Resource) fyne.CanvasObject {
	table_image := canvas.NewImageFromResource(img)
	table_image.Resize(fyne.NewSize(1100, 600))
	table_image.Move(fyne.NewPos(5, 0))

	return table_image
}

// Holdero object buffer when action triggered
func HolderoButtonBuffer() {
	Table.Sit.Hide()
	Table.Leave.Hide()
	Table.Deal.Hide()
	Table.Bet.Hide()
	Table.Check.Hide()
	Table.BetEntry.Hide()
	Table.Warning.Hide()
	rpc.Display.Res = ""
	rpc.Signal.Clicked = true
	rpc.Signal.CHeight = rpc.Wallet.Height
}

// Checking for current player names at connected Holdero table
//   - If name exists, prompt user to select new name
func CheckNames(seats string) bool {
	if rpc.Round.ID == 1 {
		return true
	}

	err := "[Holdero] Name already used"

	switch seats {
	case "2":
		if Poker_name == rpc.Round.P1_name {
			log.Println(err)
			return false
		}
		return true
	case "3":
		if Poker_name == rpc.Round.P1_name || Poker_name == rpc.Round.P2_name || Poker_name == rpc.Round.P3_name {
			log.Println(err)
			return false
		}
		return true
	case "4":
		if Poker_name == rpc.Round.P1_name || Poker_name == rpc.Round.P2_name || Poker_name == rpc.Round.P3_name || Poker_name == rpc.Round.P4_name {
			log.Println(err)
			return false
		}
		return true
	case "5":
		if Poker_name == rpc.Round.P1_name || Poker_name == rpc.Round.P2_name || Poker_name == rpc.Round.P3_name || Poker_name == rpc.Round.P4_name || Poker_name == rpc.Round.P5_name {
			log.Println(err)
			return false
		}
		return true
	case "6":
		if Poker_name == rpc.Round.P1_name || Poker_name == rpc.Round.P2_name || Poker_name == rpc.Round.P3_name || Poker_name == rpc.Round.P4_name || Poker_name == rpc.Round.P5_name || Poker_name == rpc.Round.P6_name {
			log.Println(err)
			return false
		}
		return true
	default:
		return false
	}
}

// Holdero player sit down button to join current table
func SitButton() fyne.Widget {
	Table.Sit = widget.NewButton("Sit Down", func() {
		if Poker_name != "" {
			if CheckNames(rpc.Display.Seats) {
				rpc.SitDown(Poker_name, Settings.AvatarUrl)
				HolderoButtonBuffer()
			}
		} else {
			log.Println("[Holdero] Pick a name")
		}
	})

	Table.Sit.Hide()

	return Table.Sit
}

// Holdero player leave button to leave current table seat
func LeaveButton() fyne.Widget {
	Table.Leave = widget.NewButton("Leave", func() {
		rpc.Leave()
		HolderoButtonBuffer()
	})

	Table.Leave.Hide()

	return Table.Leave
}

// Holdero player deal hand button
func DealHandButton() fyne.Widget {
	Table.Deal = widget.NewButton("Deal Hand", func() {
		if tx := rpc.DealHand(); tx != "" {
			HolderoButtonBuffer()
		}
	})

	Table.Deal.Hide()

	return Table.Deal
}

// Holdero bet entry amount
//   - Setting the initial value based on if PlacedBet, Wager and Ante
//   - If entry invalid, set to min bet value
func BetAmount() fyne.CanvasObject {
	Table.BetEntry = dwidget.DeroAmtEntry("", 0.1, 1)
	Table.BetEntry.Enable()
	if Table.BetEntry.Text == "" {
		Table.BetEntry.SetText("0.0")
	}
	Table.BetEntry.Validator = validation.NewRegexp(`^\d{1,}\.\d{1,5}$|^[^0.]\d{0,}$`, "Int or float required")
	Table.BetEntry.OnChanged = func(s string) {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			if rpc.Signal.PlacedBet {
				Table.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Raised)/100000, 'f', int(Table.BetEntry.Decimal), 64))
				if Table.BetEntry.Validate() != nil {
					Table.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Raised)/100000, 'f', int(Table.BetEntry.Decimal), 64))
				}
			} else {

				if rpc.Round.Wager > 0 {
					if rpc.Round.Raised > 0 {
						if rpc.Signal.PlacedBet {
							Table.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Raised)/100000, 'f', int(Table.BetEntry.Decimal), 64))
						} else {
							Table.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Wager)/100000, 'f', int(Table.BetEntry.Decimal), 64))
						}
						if Table.BetEntry.Validate() != nil {
							if rpc.Signal.PlacedBet {
								Table.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Raised)/100000, 'f', int(Table.BetEntry.Decimal), 64))
							} else {
								Table.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Wager)/100000, 'f', int(Table.BetEntry.Decimal), 64))
							}
						}
					} else {

						if f < float64(rpc.Round.Wager)/100000 {
							Table.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Wager)/100000, 'f', int(Table.BetEntry.Decimal), 64))
						}

						if Table.BetEntry.Validate() != nil {
							float := f * 100000
							if uint64(float)%10000 == 0 {
								Table.BetEntry.SetText(strconv.FormatFloat(roundFloat(f, 1), 'f', int(Table.BetEntry.Decimal), 64))
							} else if Table.BetEntry.Validate() != nil {
								Table.BetEntry.SetText(strconv.FormatFloat(roundFloat(f, 1), 'f', int(Table.BetEntry.Decimal), 64))
							}
						}
					}
				} else {

					if rpc.Daemon.Connect {
						float := f * 100000
						if uint64(float)%10000 == 0 {
							Table.BetEntry.SetText(strconv.FormatFloat(roundFloat(f, 1), 'f', int(Table.BetEntry.Decimal), 64))
						} else if Table.BetEntry.Validate() != nil {
							Table.BetEntry.SetText(strconv.FormatFloat(roundFloat(f, 1), 'f', int(Table.BetEntry.Decimal), 64))
						}

						if rpc.Round.Ante > 0 {
							if f < float64(rpc.Round.Ante)/100000 {
								Table.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Ante)/100000, 'f', int(Table.BetEntry.Decimal), 64))
							}

							if Table.BetEntry.Validate() != nil {
								Table.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Ante)/100000, 'f', int(Table.BetEntry.Decimal), 64))
							}

						} else {
							if f < float64(rpc.Round.BB)/100000 {
								Table.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.BB)/100000, 'f', int(Table.BetEntry.Decimal), 64))
							}

							if Table.BetEntry.Validate() != nil {
								Table.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.BB)/100000, 'f', int(Table.BetEntry.Decimal), 64))
							}
						}
					}
				}
			}
		} else {
			log.Println("[BetAmount]", err)
			if rpc.Round.Ante == 0 {
				Table.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.BB)/100000, 'f', int(Table.BetEntry.Decimal), 64))
			} else {
				Table.BetEntry.SetText(strconv.FormatFloat(float64(rpc.Round.Ante)/100000, 'f', int(Table.BetEntry.Decimal), 64))
			}
		}
	}

	amt_box := container.NewHScroll(Table.BetEntry)
	amt_box.SetMinSize(fyne.NewSize(100, 40))
	Table.BetEntry.Hide()

	return amt_box

}

// Round float val to precision
func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

// Holdero place bet button
//   - Input from Table.BetEntry
func BetButton() fyne.Widget {
	Table.Bet = widget.NewButton("Bet", func() {
		if Table.BetEntry.Validate() == nil {
			if tx := rpc.Bet(Table.BetEntry.Text); tx != "" {
				rpc.Signal.Bet = true
				HolderoButtonBuffer()
			}
		}
	})

	Table.Bet.Hide()

	return Table.Bet
}

// Holdero check and fold button
func CheckButton() fyne.Widget {
	Table.Check = widget.NewButton("Check", func() {
		if tx := rpc.Check(); tx != "" {
			rpc.Signal.Bet = true
			HolderoButtonBuffer()
		}
	})

	Table.Check.Hide()

	return Table.Check
}

// Automated options object for Holdero
func AutoOptions() fyne.CanvasObject {
	cf := widget.NewCheck("Auto Check/Fold", func(b bool) {
		if b {
			Settings.Auto_check = true
		} else {
			Settings.Auto_check = false
		}
	})

	deal := widget.NewCheck("Auto Deal", func(b bool) {
		if b {
			Settings.Auto_deal = true
		} else {
			Settings.Auto_deal = false
		}
	})

	checks := container.NewVBox(deal, cf)

	Settings.Tools = widget.NewButton("Tools", func() {
		go holderoTools(deal, cf, Settings.Tools)
	})

	DisableHolderoTools()

	auto := container.NewVBox(checks, Settings.Tools)

	return auto
}

// Holdero warning label displayed when player is risking being timed out
func TimeOutWarning() *fyne.Container {
	rect := canvas.NewRectangle(color.RGBA{0, 0, 0, 210})
	msg := canvas.NewText("Make your move, or you will be Timed Out", color.RGBA{240, 0, 0, 240})
	msg.TextSize = 15

	Table.Warning = container.NewMax(rect, msg)

	Table.Warning.Hide()

	return container.NewVBox(layout.NewSpacer(), Table.Warning)
}

// Set default params for auto bet functions
func autoBetDefault() {
	rpc.Odds.Bot.Risk[2] = 21
	rpc.Odds.Bot.Risk[1] = 9
	rpc.Odds.Bot.Risk[0] = 3
	rpc.Odds.Bot.Bet[2] = 6
	rpc.Odds.Bot.Bet[1] = 3
	rpc.Odds.Bot.Bet[0] = 1
	rpc.Odds.Bot.Luck = 0
	rpc.Odds.Bot.Slow = 4
	rpc.Odds.Bot.Aggr = 1
	rpc.Odds.Bot.Max = 10
	rpc.Odds.Bot.Random[0] = 0
	rpc.Odds.Bot.Random[1] = 0
}

// Setting current auto bet random option when menu opened
func setRandomOpts(opts *widget.RadioGroup) {
	if rpc.Odds.Bot.Random[0] == 0 {
		opts.Disable()
	} else {
		switch rpc.Odds.Bot.Random[1] {
		case 1:
			opts.SetSelected("Risk")
		case 2:
			opts.SetSelected("Bet")
		case 3:
			opts.SetSelected("Both")
		default:
			opts.SetSelected("")
		}
	}
}

// dReam Tools menu for Holdero
//   - deal check and button widgets are passed when setting auto objects for control
func holderoTools(deal, check *widget.Check, button *widget.Button) {
	button.Hide()
	bm := fyne.CurrentApp().NewWindow("Holdero Tools")
	bm.Resize(fyne.NewSize(330, 700))
	bm.SetFixedSize(true)
	bm.SetIcon(bundle.ResourceDTGnomonIconPng)
	bm.SetCloseIntercept(func() {
		button.Show()
		bm.Close()
	})

	rpc.Stats = ReadSavedStats()
	config_opts := []string{}
	for i := range rpc.Stats.Bots {
		config_opts = append(config_opts, rpc.Stats.Bots[i].Name)
	}

	entry := widget.NewSelectEntry(config_opts)
	entry.SetPlaceHolder("Default")
	entry.SetText(rpc.Odds.Bot.Name)

	curr := " Dero"
	max_bet := float64(100)
	if rpc.Round.Asset {
		curr = " Tokens"
		max_bet = 2500
	}

	mb_label := widget.NewLabel("Max Bet: " + fmt.Sprintf("%.0f", rpc.Odds.Bot.Max) + curr)
	mb_slider := widget.NewSlider(1, max_bet)
	mb_slider.SetValue(rpc.Odds.Bot.Max)
	mb_slider.OnChanged = func(f float64) {
		go func() {
			min := float64(rpc.MinBet()) / 100000
			if min == 0 {
				min = 0.1
			}

			if f < (min*rpc.Odds.Bot.Bet[2])*rpc.Odds.Bot.Aggr {
				rpc.Odds.Bot.Max = (min*rpc.Odds.Bot.Bet[2])*rpc.Odds.Bot.Aggr + 3
				mb_slider.SetValue(rpc.Odds.Bot.Max)
				mb_label.SetText("Max Bet: " + fmt.Sprintf("%.0f", rpc.Odds.Bot.Max) + curr)
			} else {
				rpc.Odds.Bot.Max = f
				mb_label.SetText("Max Bet: " + fmt.Sprintf("%.0f", f) + curr)
			}
		}()
	}

	rh_label := widget.NewLabel("Risk High: " + fmt.Sprintf("%.0f", rpc.Odds.Bot.Risk[2]) + "%")
	rh_slider := widget.NewSlider(1, 90)
	rh_slider.SetValue(rpc.Odds.Bot.Risk[2])
	rh_slider.OnChanged = func(f float64) {
		go func() {
			if f < rpc.Odds.Bot.Risk[1] {
				rpc.Odds.Bot.Risk[2] = rpc.Odds.Bot.Risk[1] + 1
				rh_slider.SetValue(rpc.Odds.Bot.Risk[2])
			} else {
				rpc.Odds.Bot.Risk[2] = f
			}

			rh_label.SetText("Risk High: " + fmt.Sprintf("%.0f", rpc.Odds.Bot.Risk[2]) + "%")
		}()
	}

	rm_label := widget.NewLabel("Risk Medium: " + fmt.Sprintf("%.0f", rpc.Odds.Bot.Risk[1]) + "%")
	rm_slider := widget.NewSlider(1, 89)
	rm_slider.SetValue(rpc.Odds.Bot.Risk[1])
	rm_slider.OnChanged = func(f float64) {
		go func() {
			rpc.Odds.Bot.Risk[1] = f
			if f <= rpc.Odds.Bot.Risk[0] {
				rpc.Odds.Bot.Risk[1] = rpc.Odds.Bot.Risk[0] + 1
				rm_slider.SetValue(rpc.Odds.Bot.Risk[1])
			}

			if f >= rpc.Odds.Bot.Risk[2] {
				rpc.Odds.Bot.Risk[2] = f + 1
				rh_slider.SetValue(rpc.Odds.Bot.Risk[2])
			}

			rm_label.SetText("Risk Medium: " + fmt.Sprintf("%.0f", rpc.Odds.Bot.Risk[1]) + "%")
		}()
	}

	rl_label := widget.NewLabel("Risk Low: " + fmt.Sprintf("%.0f", rpc.Odds.Bot.Risk[0]) + "%")
	rl_slider := widget.NewSlider(1, 88)
	rl_slider.SetValue(rpc.Odds.Bot.Risk[0])
	rl_slider.OnChanged = func(f float64) {
		go func() {
			if rpc.Odds.Bot.Risk[1] <= f {
				rm_slider.SetValue(f + 1)
			}

			rpc.Odds.Bot.Risk[0] = f
			rl_label.SetText("Risk Low: " + fmt.Sprintf("%.0f", rpc.Odds.Bot.Risk[0]) + "%")
		}()
	}

	bh_label := widget.NewLabel("Bet High: " + fmt.Sprintf("%.0f", rpc.Odds.Bot.Bet[2]) + "x")
	bh_slider := widget.NewSlider(1, 100)
	bh_slider.SetValue(rpc.Odds.Bot.Bet[2])
	bh_slider.OnChanged = func(f float64) {
		go func() {
			if f < rpc.Odds.Bot.Bet[1] {
				rpc.Odds.Bot.Bet[2] = rpc.Odds.Bot.Bet[1] + 1
				bh_slider.SetValue(rpc.Odds.Bot.Bet[2])
			} else {
				rpc.Odds.Bot.Bet[2] = f
			}

			min := float64(rpc.MinBet()) / 100000
			if min == 0 {
				min = 0.1
			}

			if rpc.Odds.Bot.Max < (min*rpc.Odds.Bot.Bet[2])*rpc.Odds.Bot.Aggr {
				rpc.Odds.Bot.Max = (min * rpc.Odds.Bot.Bet[2]) * rpc.Odds.Bot.Aggr
				mb_slider.SetValue(rpc.Odds.Bot.Max)
			}

			bh_label.SetText("Bet High: " + fmt.Sprintf("%.0f", rpc.Odds.Bot.Bet[2]) + "x")
		}()
	}

	bm_label := widget.NewLabel("Bet Medium: " + fmt.Sprintf("%.0f", rpc.Odds.Bot.Bet[1]) + "x")
	bm_slider := widget.NewSlider(1, 99)
	bm_slider.SetValue(rpc.Odds.Bot.Bet[1])
	bm_slider.OnChanged = func(f float64) {
		go func() {
			rpc.Odds.Bot.Bet[1] = f
			if f <= rpc.Odds.Bot.Bet[0] {
				rpc.Odds.Bot.Bet[1] = rpc.Odds.Bot.Bet[0] + 1
				bm_slider.SetValue(rpc.Odds.Bot.Bet[1])
			}

			if f >= rpc.Odds.Bot.Bet[2] {
				rpc.Odds.Bot.Bet[2] = f + 1
				bh_slider.SetValue(rpc.Odds.Bot.Bet[2])
			}

			bm_label.SetText("Bet Medium: " + fmt.Sprintf("%.0f", rpc.Odds.Bot.Bet[1]) + "x")
		}()
	}

	bl_label := widget.NewLabel("Bet Low: " + fmt.Sprintf("%.0f", rpc.Odds.Bot.Bet[0]) + "x")
	bl_slider := widget.NewSlider(1, 98)
	bl_slider.SetValue(rpc.Odds.Bot.Bet[0])
	bl_slider.OnChanged = func(f float64) {
		go func() {
			if rpc.Odds.Bot.Bet[1] <= f {
				bm_slider.SetValue(f + 1)
			}

			rpc.Odds.Bot.Bet[0] = f
			bl_label.SetText("Bet Low: " + fmt.Sprintf("%.0f", rpc.Odds.Bot.Bet[0]) + "x")
		}()
	}

	luck_label := widget.NewLabel("Luck: " + fmt.Sprintf("%.2f", rpc.Odds.Bot.Luck))
	luck_slider := widget.NewSlider(0, 10)
	luck_slider.Step = 0.25
	luck_slider.SetValue(rpc.Odds.Bot.Luck)
	luck_slider.OnChanged = func(f float64) {
		go func() {
			rpc.Odds.Bot.Luck = f
			luck_label.SetText("Luck: " + fmt.Sprintf("%.2f", f))
		}()
	}

	random_label := widget.NewLabel("Randomize: Off")
	if rpc.Odds.Bot.Random[0] == 0 {
		random_label.SetText("Randomize: Off")
	} else {
		random_label.SetText("Randomize: " + fmt.Sprintf("%.2f", rpc.Odds.Bot.Random[0]))
	}

	random_opts := widget.NewRadioGroup([]string{"Risk", "Bet", "Both"}, func(s string) {
		switch s {
		case "Risk":
			rpc.Odds.Bot.Random[1] = 1
		case "Bet":
			rpc.Odds.Bot.Random[1] = 2
		case "Both":
			rpc.Odds.Bot.Random[1] = 3
		default:
			rpc.Odds.Bot.Random[1] = 0
		}
	})

	setRandomOpts(random_opts)

	random_slider := widget.NewSlider(0, 10)
	random_slider.Step = 0.25
	random_slider.SetValue(rpc.Odds.Bot.Random[0])
	random_slider.OnChanged = func(f float64) {
		go func() {
			rpc.Odds.Bot.Random[0] = f
			if f >= 0.5 {
				random_label.SetText("Randomize: " + fmt.Sprintf("%.2f", f))
				random_opts.Enable()
			} else {
				rpc.Odds.Bot.Random[0] = 0
				rpc.Odds.Bot.Random[1] = 0
				random_label.SetText("Randomize: Off")
				random_opts.SetSelected("")
				random_opts.Disable()
			}
		}()
	}

	slow_label := widget.NewLabel("Slowplay: " + fmt.Sprintf("%.0f", rpc.Odds.Bot.Slow))
	slow_slider := widget.NewSlider(1, 5)
	slow_slider.SetValue(rpc.Odds.Bot.Slow)
	slow_slider.OnChanged = func(f float64) {
		go func() {
			rpc.Odds.Bot.Slow = f
			slow_label.SetText("Slowplay: " + fmt.Sprintf("%.0f", f))
		}()
	}

	aggr_label := widget.NewLabel("Aggression: " + fmt.Sprintf("%.0f", rpc.Odds.Bot.Aggr))
	aggr_slider := widget.NewSlider(1, 5)
	aggr_slider.SetValue(rpc.Odds.Bot.Aggr)
	aggr_slider.OnChanged = func(f float64) {
		go func() {
			rpc.Odds.Bot.Aggr = f
			min := float64(rpc.MinBet()) / 100000
			if min == 0 {
				min = 0.1
			}

			if rpc.Odds.Bot.Max < (min*rpc.Odds.Bot.Bet[2])*rpc.Odds.Bot.Aggr {
				rpc.Odds.Bot.Max = (min * rpc.Odds.Bot.Bet[2]) * rpc.Odds.Bot.Aggr
				mb_slider.SetValue(rpc.Odds.Bot.Max)
			}

			aggr_label.SetText("Aggression: " + fmt.Sprintf("%.0f", f))
		}()
	}

	rem := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "delete"), func() {
		if entry.Text != "" {
			var new []rpc.Bot_config
			for i := range rpc.Stats.Bots {
				if rpc.Stats.Bots[i].Name == entry.Text {
					log.Println("[dReams] Deleting bot config")
					if i > 0 {
						new = append(rpc.Stats.Bots[0:i], rpc.Stats.Bots[i+1:]...)
						config_opts = append(config_opts[0:i], config_opts[i+1:]...)
						break
					} else {
						if len(config_opts) < 2 {
							new = nil
							config_opts = []string{}
						} else {
							new = rpc.Stats.Bots[1:]
							config_opts = append(config_opts[1:2], config_opts[2:]...)
						}
						break
					}
				}
			}

			rpc.Stats.Bots = new
			rpc.WriteHolderoStats(rpc.Stats)
			entry.SetOptions(config_opts)
			entry.SetText("")
		}
	})

	reset := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "viewRefresh"), func() {
		autoBetDefault()
		mb_slider.SetValue(rpc.Odds.Bot.Max)
		rh_slider.SetValue(rpc.Odds.Bot.Risk[2])
		rm_slider.SetValue(rpc.Odds.Bot.Risk[1])
		rl_slider.SetValue(rpc.Odds.Bot.Risk[0])
		bh_slider.SetValue(rpc.Odds.Bot.Bet[2])
		bm_slider.SetValue(rpc.Odds.Bot.Bet[1])
		bl_slider.SetValue(rpc.Odds.Bot.Bet[0])
		luck_slider.SetValue(rpc.Odds.Bot.Luck)
		random_slider.SetValue(rpc.Odds.Bot.Random[0])
		slow_slider.SetValue(rpc.Odds.Bot.Slow)
		aggr_slider.SetValue(rpc.Odds.Bot.Aggr)
		random_opts.SetSelected("")
		entry.SetText("")
		rpc.Odds.Bot.Name = ""
	})

	save := widget.NewButton("Save", func() {
		if entry.Text != "" {
			var ex bool
			for i := range rpc.Stats.Bots {
				if entry.Text == rpc.Stats.Bots[i].Name {
					ex = true
					log.Println("[dReams] Bot config name exists")
				}
			}

			if !ex {
				rpc.Stats.Bots = append(rpc.Stats.Bots, rpc.Odds.Bot)
				if rpc.WriteHolderoStats(rpc.Stats) {
					config_opts = append(config_opts, entry.Text)
					entry.SetOptions(config_opts)
					log.Println("[dReams] Saved bot config")
				}
			}
		}
	})

	entry.OnChanged = func(s string) {
		if s != "" {
			rpc.Odds.Bot.Name = entry.Text
			for i := range config_opts {
				if s == config_opts[i] {
					for r := range rpc.Stats.Bots {
						if rpc.Stats.Bots[r].Name == config_opts[i] {
							rpc.SetBotConfig(rpc.Stats.Bots[r])
							mb_slider.SetValue(rpc.Odds.Bot.Max)
							rh_slider.SetValue(rpc.Odds.Bot.Risk[2])
							rm_slider.SetValue(rpc.Odds.Bot.Risk[1])
							rl_slider.SetValue(rpc.Odds.Bot.Risk[0])
							bh_slider.SetValue(rpc.Odds.Bot.Bet[2])
							bm_slider.SetValue(rpc.Odds.Bot.Bet[1])
							bl_slider.SetValue(rpc.Odds.Bot.Bet[0])
							luck_slider.SetValue(rpc.Odds.Bot.Luck)
							random_slider.SetValue(rpc.Odds.Bot.Random[0])
							slow_slider.SetValue(rpc.Odds.Bot.Slow)
							aggr_slider.SetValue(rpc.Odds.Bot.Aggr)
							setRandomOpts(random_opts)
						}
					}
				}
			}
		}
	}

	enable := widget.NewCheck("Auto Bet Enabled", func(b bool) {
		if b {
			rpc.Odds.Run = true
			if check.Checked {
				check.SetChecked(false)
			}
			check.Disable()
			deal.SetChecked(true)
		} else {
			rpc.Odds.Run = false
			check.Enable()
			if deal.Checked {
				deal.SetChecked(false)
			}
		}
	})

	if rpc.Odds.Run {
		enable.SetChecked(true)
	}

	config_buttons := container.NewBorder(nil, nil, nil, container.NewHBox(reset, rem), save)

	top_box := container.NewAdaptiveGrid(2,
		container.NewVBox(
			luck_label,
			luck_slider,
			slow_label,
			slow_slider,
			aggr_label,
			aggr_slider,
			mb_label,
			mb_slider,
			layout.NewSpacer(),
			enable),

		container.NewVBox(
			random_label,
			random_slider,
			random_opts,
			layout.NewSpacer(),
			entry,
			config_buttons))

	rpc.Odds.Label = widget.NewLabel("")
	rpc.Odds.Label.Wrapping = fyne.TextWrapWord
	scroll := container.NewVScroll(rpc.Odds.Label)
	odds_button := widget.NewButton("Odds", func() {
		odds, future := rpc.MakeOdds()
		rpc.BetLogic(odds, future, false)
	})

	r_box := container.NewVBox(
		rh_label,
		rh_slider,
		rm_label,
		rm_slider,
		rl_label,
		rl_slider)

	b_box := container.NewVBox(
		bh_label,
		bh_slider,
		bm_label,
		bm_slider,
		bl_label,
		bl_slider)

	bet_bot := container.NewVBox(
		r_box,
		layout.NewSpacer(),
		b_box,
		layout.NewSpacer(),
		top_box)

	odds_box := container.NewVBox(layout.NewSpacer(), odds_button)
	max := container.NewMax(scroll, odds_box)

	stats_label := widget.NewLabel("")

	tabs := container.NewAppTabs(
		container.NewTabItem("Bot", container.NewBorder(nil, nil, nil, nil, bet_bot)),
		container.NewTabItem("Odds", max),
		container.NewTabItem("Stats", stats_label),
	)

	tabs.OnSelected = func(ti *container.TabItem) {
		switch ti.Text {
		case "Stats":
			stats_label.SetText("Total Player Stats\n\nWins: " + strconv.Itoa(rpc.Stats.Player.Win) + "\n\nLost: " + strconv.Itoa(rpc.Stats.Player.Lost) +
				"\n\nFolded: " + strconv.Itoa(rpc.Stats.Player.Fold) + "\n\nPush: " + strconv.Itoa(rpc.Stats.Player.Push) +
				"\n\nWagered: " + fmt.Sprintf("%.5f", rpc.Stats.Player.Wagered) + "\n\nEarnings: " + fmt.Sprintf("%.5f", rpc.Stats.Player.Earnings))

			if rpc.Odds.Bot.Name != "" {
				for i := range rpc.Stats.Bots {
					if rpc.Odds.Bot.Name == rpc.Stats.Bots[i].Name {
						stats_label.SetText(stats_label.Text + "\n\n\nBot Stats\n\nBot: " + rpc.Odds.Bot.Name + "\n\nWins: " + strconv.Itoa(rpc.Stats.Bots[i].Stats.Win) +
							"\n\nLost: " + strconv.Itoa(rpc.Stats.Bots[i].Stats.Lost) + "\n\nFolded: " + strconv.Itoa(rpc.Stats.Bots[i].Stats.Fold) + "\n\nPush: " + strconv.Itoa(rpc.Stats.Bots[i].Stats.Push) +
							"\n\nWagered: " + fmt.Sprintf("%.5f", rpc.Stats.Bots[i].Stats.Wagered) + "\n\nEarnings: " + fmt.Sprintf("%.5f", rpc.Stats.Bots[i].Stats.Earnings))
					}
				}
			}
		}
	}

	go func() {
		for rpc.Wallet.Connect {
			time.Sleep(1 * time.Second)
		}

		button.Show()
		bm.Close()
	}()

	img := *canvas.NewImageFromResource(bundle.ResourceOwBackgroundPng)
	bm.SetContent(
		container.New(layout.NewMaxLayout(),
			&img,
			bundle.Alpha180,
			tabs))
	bm.Show()
}

func DisableHolderoTools() {
	rpc.Odds.Enabled = false
	Settings.Tools.Hide()
	if len(Settings.BackSelect.Options) > 2 || len(Settings.FaceSelect.Options) > 2 {
		cards := false
		for _, f := range Settings.FaceSelect.Options {
			asset := strings.Trim(f, "0123456789")
			switch asset {
			case "AZYPC":
				cards = true
			case "SIXPC":
				cards = true
			default:

			}

			if cards {
				break
			}
		}

		if !cards {
			for _, b := range Settings.BackSelect.Options {
				asset := strings.Trim(b, "0123456789")
				switch asset {
				case "AZYPCB":
					cards = true
				case "SIXPCB":
					cards = true
				default:

				}

				if cards {
					break
				}
			}
		}

		if cards {
			rpc.Odds.Enabled = true
			Settings.Tools.Show()
			if !FileExists("config/stats.json", "dReams") {
				rpc.WriteHolderoStats(rpc.Stats)
				log.Println("[dReams] Created stats.json")
			} else {
				rpc.Stats = ReadSavedStats()
			}
		}
	}
}

// Reading saved Holdero stats from config file
func ReadSavedStats() (saved rpc.Player_stats) {
	file, err := os.ReadFile("config/stats.json")

	if err != nil {
		log.Println("[readSavedStats]", err)
		return
	}

	err = json.Unmarshal(file, &saved)
	if err != nil {
		log.Println("[readSavedStats]", err)
		return
	}

	return
}
