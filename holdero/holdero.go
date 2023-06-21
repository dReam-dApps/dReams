package holdero

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	dreams "github.com/SixofClubsss/dReams"
	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/dwidget"
	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type holderoObjects struct {
	Contract_entry *widget.SelectEntry
	Table_list     *widget.List
	Favorite_list  *widget.List
	Owned_list     *widget.List
	Holdero_unlock *widget.Button
	Holdero_new    *widget.Button
	Stats_box      fyne.Container
	owner          struct {
		blind_amount uint64
		ante_amount  uint64
		chips        *widget.RadioGroup
		timeout      *widget.Button
		owners_left  *fyne.Container
		owners_mid   *fyne.Container
	}
}

type settings struct {
	Tables        []string
	Favorites     []string
	Owned         []string
	Avatar        string
	AvatarUrl     string
	Synced        bool
	Shared        bool
	Auto_check    bool
	Auto_deal     bool
	P1_avatar_url string
	P2_avatar_url string
	P3_avatar_url string
	P4_avatar_url string
	P5_avatar_url string
	P6_avatar_url string
	Check         *widget.Check
	AvatarSelect  *widget.Select
	Tools         *widget.Button
	SharedOn      *widget.RadioGroup
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
	Stats      struct {
		Name    *canvas.Text
		Desc    *canvas.Text
		Version *canvas.Text
		Last    *canvas.Text
		Seats   *canvas.Text
		Open    *canvas.Text
		Image   canvas.Image
	}
}

var Poker holderoObjects
var Table tableObjects
var Settings settings

func initValues() {
	Times.Delay = 30
	Times.Kick = 0
	Odds.Run = false
	Faces.Name = "light/"
	Backs.Name = "back1.png"
	Settings.Avatar = "None"
	Faces.URL = ""
	Backs.URL = ""
	Settings.AvatarUrl = ""
	Settings.Auto_deal = false
	Settings.Auto_check = false
	Signal.Sit = true
	autoBetDefault()
}

// Holdero SCID entry
//   - Bound to rpc.Round.Contract
//   - Entry text set on list selection
//   - Changes clear table and check if current entry is valid table
func holderoContractEntry() fyne.Widget {
	var wait bool
	Poker.Contract_entry = widget.NewSelectEntry(nil)
	options := []string{""}
	Poker.Contract_entry.SetOptions(options)
	Poker.Contract_entry.PlaceHolder = "Holdero Contract Address: "
	Poker.Contract_entry.OnCursorChanged = func() {
		if rpc.Daemon.IsConnected() && !wait {
			wait = true
			text := Poker.Contract_entry.Text
			clearShared()
			if len(text) == 64 {
				if checkTableOwner(text) {
					disableOwnerControls(false)
					if checkTableVersion(text) >= 110 {
						Poker.owner.chips.Show()
						Poker.owner.timeout.Show()
						Poker.owner.owners_mid.Show()
					} else {
						Poker.owner.chips.Hide()
						Poker.owner.timeout.Hide()
						Poker.owner.owners_mid.Hide()
					}
				} else {
					disableOwnerControls(true)
				}

				if rpc.Wallet.IsConnected() && checkHolderoContract(text) {
					Table.Tournament.Show()
				} else {
					Table.Tournament.Hide()
				}
			} else {
				Signal.Contract = false
				Settings.Check.SetChecked(false)
				Table.Tournament.Hide()
			}
			wait = false
		}
	}

	this := binding.BindString(&Round.Contract)
	Poker.Contract_entry.Bind(this)

	return Poker.Contract_entry
}

// Routine when Holdero SCID is clicked
func setHolderoControls(str string) (item string) {
	split := strings.Split(str, "   ")
	if len(split) >= 3 {
		trimmed := strings.Trim(split[2], " ")
		if len(trimmed) == 64 {
			item = str
			Poker.Contract_entry.SetText(trimmed)
			go getTableStats(trimmed, true)
			Times.Kick_block = rpc.Wallet.Height
		}
	}

	return
}

// Public Holdero table listings object
func tableListings(tab *container.AppTabs) fyne.CanvasObject {
	Poker.Table_list = widget.NewList(
		func() int {
			return len(Settings.Tables)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(canvas.NewImageFromImage(nil), widget.NewLabel(""))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*fyne.Container).Objects[1].(*widget.Label).SetText(Settings.Tables[i])
			if Settings.Tables[i][0:2] != "  " {
				var key string
				split := strings.Split(Settings.Tables[i], "   ")
				if len(split) >= 3 {
					trimmed := strings.Trim(split[2], " ")
					if len(trimmed) == 64 {
						key = trimmed
					}
				}

				badge := canvas.NewImageFromResource(menu.DisplayRating(menu.Control.Contract_rating[key]))
				badge.SetMinSize(fyne.NewSize(35, 35))
				o.(*fyne.Container).Objects[0] = badge
			}
		})

	var item string

	Poker.Table_list.OnSelected = func(id widget.ListItemID) {
		if id != 0 && menu.Connected() {
			go func() {
				item = setHolderoControls(Settings.Tables[id])
				Poker.Favorite_list.UnselectAll()
				Poker.Owned_list.UnselectAll()
			}()
		}
	}

	save_favorite := widget.NewButton("Favorite", func() {
		Settings.Favorites = append(Settings.Favorites, item)
		sort.Strings(Settings.Favorites)
	})

	rate_contract := widget.NewButton("Rate", func() {
		if len(Round.Contract) == 64 {
			if !checkTableOwner(Round.Contract) {
				reset := tab.Selected().Content
				tab.Selected().Content = menu.RateConfirm(Round.Contract, tab, reset)
				tab.Selected().Content.Refresh()

			} else {
				log.Println("[dReams] You own this contract")
			}
		}
	})

	tables_cont := container.NewBorder(
		nil,
		container.NewBorder(nil, nil, save_favorite, rate_contract, layout.NewSpacer()),
		nil,
		nil,
		Poker.Table_list)

	return tables_cont
}

// Favorite Holdero tables object
func holderoFavorites() fyne.CanvasObject {
	Poker.Favorite_list = widget.NewList(
		func() int {
			return len(Settings.Favorites)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(Settings.Favorites[i])
		})

	var item string

	Poker.Favorite_list.OnSelected = func(id widget.ListItemID) {
		if menu.Connected() {
			item = setHolderoControls(Settings.Favorites[id])
			Poker.Table_list.UnselectAll()
			Poker.Owned_list.UnselectAll()
		}
	}

	remove := widget.NewButton("Remove", func() {
		if len(Settings.Favorites) > 0 {
			Poker.Favorite_list.UnselectAll()
			for i := range Settings.Favorites {
				if Settings.Favorites[i] == item {
					copy(Settings.Favorites[i:], Settings.Favorites[i+1:])
					Settings.Favorites[len(Settings.Favorites)-1] = ""
					Settings.Favorites = Settings.Favorites[:len(Settings.Favorites)-1]
					break
				}
			}
		}
		Poker.Favorite_list.Refresh()
		sort.Strings(Settings.Favorites)
	})

	cont := container.NewBorder(
		nil,
		container.NewBorder(nil, nil, nil, remove, layout.NewSpacer()),
		nil,
		nil,
		Poker.Favorite_list)

	return cont
}

// Owned Holdero tables object
func myTables() fyne.CanvasObject {
	Poker.Owned_list = widget.NewList(
		func() int {
			return len(Settings.Owned)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(Settings.Owned[i])
		})

	Poker.Owned_list.OnSelected = func(id widget.ListItemID) {
		if menu.Connected() {
			setHolderoControls(Settings.Owned[id])
			Poker.Table_list.UnselectAll()
			Poker.Favorite_list.UnselectAll()
		}
	}

	return Poker.Owned_list
}

// Table owner name and avatar objects
//   - Pass a and f as avatar and its frame resource, shared avatar is set here if image exists
//   - Pass t for player's turn frame resource
func Player1_label(a, f, t fyne.Resource) fyne.CanvasObject {
	var name fyne.CanvasObject
	var avatar fyne.CanvasObject
	var frame fyne.CanvasObject
	var out fyne.CanvasObject
	if Signal.In1 {
		if Round.Turn == 1 {
			name = canvas.NewText(Round.P1_name, color.RGBA{105, 90, 205, 210})
		} else {
			name = canvas.NewText(Round.P1_name, color.White)
		}
	} else {
		name = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	if a != nil && Signal.In1 {
		if Round.P1_url != "" {
			avatar = &Shared.P1_avatar
			if Round.Turn == 1 {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		} else {
			avatar = canvas.NewImageFromResource(a)
			if Round.Turn == 1 {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		}
	} else {
		avatar = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
		frame = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	if Signal.Out1 {
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
	if Signal.In2 {
		if Round.Turn == 2 {
			name = canvas.NewText(Round.P2_name, color.RGBA{105, 90, 205, 210})
		} else {
			name = canvas.NewText(Round.P2_name, color.White)
		}
	} else {
		name = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	if a != nil && Signal.In2 {
		if Round.P2_url != "" {
			avatar = &Shared.P2_avatar
			if Round.Turn == 2 {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		} else {
			avatar = canvas.NewImageFromResource(a)
			if Round.Turn == 2 {
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
	if Signal.In3 {
		if Round.Turn == 3 {
			name = canvas.NewText(Round.P3_name, color.RGBA{105, 90, 205, 210})
		} else {
			name = canvas.NewText(Round.P3_name, color.White)
		}
	} else {
		name = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	if a != nil && Signal.In3 {
		if Round.P3_url != "" {
			avatar = &Shared.P3_avatar
			if Round.Turn == 3 {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		} else {
			avatar = canvas.NewImageFromResource(a)
			if Round.Turn == 3 {
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
	if Signal.In4 {
		if Round.Turn == 4 {
			name = canvas.NewText(Round.P4_name, color.RGBA{105, 90, 205, 210})
		} else {
			name = canvas.NewText(Round.P4_name, color.White)
		}
	} else {
		name = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	if a != nil && Signal.In4 {
		if Round.P4_url != "" {
			avatar = &Shared.P4_avatar
			if Round.Turn == 4 {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		} else {
			avatar = canvas.NewImageFromResource(a)
			if Round.Turn == 4 {
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
	if Signal.In5 {
		if Round.Turn == 5 {
			name = canvas.NewText(Round.P5_name, color.RGBA{105, 90, 205, 210})
		} else {
			name = canvas.NewText(Round.P5_name, color.White)
		}
	} else {
		name = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	if a != nil && Signal.In5 {
		if Round.P5_url != "" {
			avatar = &Shared.P5_avatar
			if Round.Turn == 5 {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		} else {
			avatar = canvas.NewImageFromResource(a)
			if Round.Turn == 5 {
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
	if Signal.In6 {
		if Round.Turn == 6 {
			name = canvas.NewText(Round.P6_name, color.RGBA{105, 90, 205, 210})
		} else {
			name = canvas.NewText(Round.P6_name, color.White)
		}
	} else {
		name = canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	}

	if a != nil && Signal.In6 {
		if Round.P6_url != "" {
			avatar = &Shared.P6_avatar
			if Round.Turn == 6 {
				frame = canvas.NewImageFromResource(t)
			} else {
				frame = canvas.NewImageFromResource(f)
			}
		} else {
			avatar = canvas.NewImageFromResource(a)
			if Round.Turn == 6 {
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
func ActionBuffer() {
	Table.Sit.Hide()
	Table.Leave.Hide()
	Table.Deal.Hide()
	Table.Bet.Hide()
	Table.Check.Hide()
	Table.BetEntry.Hide()
	Table.Warning.Hide()
	Display.Res = ""
	Signal.Clicked = true
	Signal.CHeight = rpc.Wallet.Height
}

// Checking for current player names at connected Holdero table
//   - If name exists, prompt user to select new name
func checkNames(seats string) bool {
	if Round.ID == 1 {
		return true
	}

	err := "[Holdero] Name already used"

	switch seats {
	case "2":
		if menu.Username == Round.P1_name {
			log.Println(err)
			return false
		}
		return true
	case "3":
		if menu.Username == Round.P1_name || menu.Username == Round.P2_name || menu.Username == Round.P3_name {
			log.Println(err)
			return false
		}
		return true
	case "4":
		if menu.Username == Round.P1_name || menu.Username == Round.P2_name || menu.Username == Round.P3_name || menu.Username == Round.P4_name {
			log.Println(err)
			return false
		}
		return true
	case "5":
		if menu.Username == Round.P1_name || menu.Username == Round.P2_name || menu.Username == Round.P3_name || menu.Username == Round.P4_name || menu.Username == Round.P5_name {
			log.Println(err)
			return false
		}
		return true
	case "6":
		if menu.Username == Round.P1_name || menu.Username == Round.P2_name || menu.Username == Round.P3_name || menu.Username == Round.P4_name || menu.Username == Round.P5_name || menu.Username == Round.P6_name {
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
		if menu.Username != "" {
			if checkNames(Display.Seats) {
				SitDown(menu.Username, Settings.AvatarUrl)
				ActionBuffer()
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
		Leave()
		ActionBuffer()
	})

	Table.Leave.Hide()

	return Table.Leave
}

// Holdero player deal hand button
func DealHandButton() fyne.Widget {
	Table.Deal = widget.NewButton("Deal Hand", func() {
		if tx := DealHand(); tx != "" {
			ActionBuffer()
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
			if Signal.PlacedBet {
				Table.BetEntry.SetText(strconv.FormatFloat(float64(Round.Raised)/100000, 'f', int(Table.BetEntry.Decimal), 64))
				if Table.BetEntry.Validate() != nil {
					Table.BetEntry.SetText(strconv.FormatFloat(float64(Round.Raised)/100000, 'f', int(Table.BetEntry.Decimal), 64))
				}
			} else {

				if Round.Wager > 0 {
					if Round.Raised > 0 {
						if Signal.PlacedBet {
							Table.BetEntry.SetText(strconv.FormatFloat(float64(Round.Raised)/100000, 'f', int(Table.BetEntry.Decimal), 64))
						} else {
							Table.BetEntry.SetText(strconv.FormatFloat(float64(Round.Wager)/100000, 'f', int(Table.BetEntry.Decimal), 64))
						}
						if Table.BetEntry.Validate() != nil {
							if Signal.PlacedBet {
								Table.BetEntry.SetText(strconv.FormatFloat(float64(Round.Raised)/100000, 'f', int(Table.BetEntry.Decimal), 64))
							} else {
								Table.BetEntry.SetText(strconv.FormatFloat(float64(Round.Wager)/100000, 'f', int(Table.BetEntry.Decimal), 64))
							}
						}
					} else {

						if f < float64(Round.Wager)/100000 {
							Table.BetEntry.SetText(strconv.FormatFloat(float64(Round.Wager)/100000, 'f', int(Table.BetEntry.Decimal), 64))
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

					if rpc.Daemon.IsConnected() {
						float := f * 100000
						if uint64(float)%10000 == 0 {
							Table.BetEntry.SetText(strconv.FormatFloat(roundFloat(f, 1), 'f', int(Table.BetEntry.Decimal), 64))
						} else if Table.BetEntry.Validate() != nil {
							Table.BetEntry.SetText(strconv.FormatFloat(roundFloat(f, 1), 'f', int(Table.BetEntry.Decimal), 64))
						}

						if Round.Ante > 0 {
							if f < float64(Round.Ante)/100000 {
								Table.BetEntry.SetText(strconv.FormatFloat(float64(Round.Ante)/100000, 'f', int(Table.BetEntry.Decimal), 64))
							}

							if Table.BetEntry.Validate() != nil {
								Table.BetEntry.SetText(strconv.FormatFloat(float64(Round.Ante)/100000, 'f', int(Table.BetEntry.Decimal), 64))
							}

						} else {
							if f < float64(Round.BB)/100000 {
								Table.BetEntry.SetText(strconv.FormatFloat(float64(Round.BB)/100000, 'f', int(Table.BetEntry.Decimal), 64))
							}

							if Table.BetEntry.Validate() != nil {
								Table.BetEntry.SetText(strconv.FormatFloat(float64(Round.BB)/100000, 'f', int(Table.BetEntry.Decimal), 64))
							}
						}
					}
				}
			}
		} else {
			log.Println("[BetAmount]", err)
			if Round.Ante == 0 {
				Table.BetEntry.SetText(strconv.FormatFloat(float64(Round.BB)/100000, 'f', int(Table.BetEntry.Decimal), 64))
			} else {
				Table.BetEntry.SetText(strconv.FormatFloat(float64(Round.Ante)/100000, 'f', int(Table.BetEntry.Decimal), 64))
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
			if tx := Bet(Table.BetEntry.Text); tx != "" {
				Signal.Bet = true
				ActionBuffer()
			}
		}
	})

	Table.Bet.Hide()

	return Table.Bet
}

// Holdero check and fold button
func CheckButton() fyne.Widget {
	Table.Check = widget.NewButton("Check", func() {
		if tx := Check(); tx != "" {
			Signal.Bet = true
			ActionBuffer()
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
	Odds.Bot.Risk[2] = 21
	Odds.Bot.Risk[1] = 9
	Odds.Bot.Risk[0] = 3
	Odds.Bot.Bet[2] = 6
	Odds.Bot.Bet[1] = 3
	Odds.Bot.Bet[0] = 1
	Odds.Bot.Luck = 0
	Odds.Bot.Slow = 4
	Odds.Bot.Aggr = 1
	Odds.Bot.Max = 10
	Odds.Bot.Random[0] = 0
	Odds.Bot.Random[1] = 0
}

// Setting current auto bet random option when menu opened
func setRandomOpts(opts *widget.RadioGroup) {
	if Odds.Bot.Random[0] == 0 {
		opts.Disable()
	} else {
		switch Odds.Bot.Random[1] {
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
	bm.SetIcon(bundle.ResourceDReamsIconAltPng)
	bm.SetCloseIntercept(func() {
		button.Show()
		bm.Close()
	})

	Stats = readSavedStats()
	config_opts := []string{}
	for i := range Stats.Bots {
		config_opts = append(config_opts, Stats.Bots[i].Name)
	}

	entry := widget.NewSelectEntry(config_opts)
	entry.SetPlaceHolder("Default")
	entry.SetText(Odds.Bot.Name)

	curr := " Dero"
	max_bet := float64(100)
	if Round.Asset {
		curr = " Tokens"
		max_bet = 2500
	}

	mb_label := widget.NewLabel("Max Bet: " + fmt.Sprintf("%.0f", Odds.Bot.Max) + curr)
	mb_slider := widget.NewSlider(1, max_bet)
	mb_slider.SetValue(Odds.Bot.Max)
	mb_slider.OnChanged = func(f float64) {
		go func() {
			min := float64(MinBet()) / 100000
			if min == 0 {
				min = 0.1
			}

			if f < (min*Odds.Bot.Bet[2])*Odds.Bot.Aggr {
				Odds.Bot.Max = (min*Odds.Bot.Bet[2])*Odds.Bot.Aggr + 3
				mb_slider.SetValue(Odds.Bot.Max)
				mb_label.SetText("Max Bet: " + fmt.Sprintf("%.0f", Odds.Bot.Max) + curr)
			} else {
				Odds.Bot.Max = f
				mb_label.SetText("Max Bet: " + fmt.Sprintf("%.0f", f) + curr)
			}
		}()
	}

	rh_label := widget.NewLabel("Risk High: " + fmt.Sprintf("%.0f", Odds.Bot.Risk[2]) + "%")
	rh_slider := widget.NewSlider(1, 90)
	rh_slider.SetValue(Odds.Bot.Risk[2])
	rh_slider.OnChanged = func(f float64) {
		go func() {
			if f < Odds.Bot.Risk[1] {
				Odds.Bot.Risk[2] = Odds.Bot.Risk[1] + 1
				rh_slider.SetValue(Odds.Bot.Risk[2])
			} else {
				Odds.Bot.Risk[2] = f
			}

			rh_label.SetText("Risk High: " + fmt.Sprintf("%.0f", Odds.Bot.Risk[2]) + "%")
		}()
	}

	rm_label := widget.NewLabel("Risk Medium: " + fmt.Sprintf("%.0f", Odds.Bot.Risk[1]) + "%")
	rm_slider := widget.NewSlider(1, 89)
	rm_slider.SetValue(Odds.Bot.Risk[1])
	rm_slider.OnChanged = func(f float64) {
		go func() {
			Odds.Bot.Risk[1] = f
			if f <= Odds.Bot.Risk[0] {
				Odds.Bot.Risk[1] = Odds.Bot.Risk[0] + 1
				rm_slider.SetValue(Odds.Bot.Risk[1])
			}

			if f >= Odds.Bot.Risk[2] {
				Odds.Bot.Risk[2] = f + 1
				rh_slider.SetValue(Odds.Bot.Risk[2])
			}

			rm_label.SetText("Risk Medium: " + fmt.Sprintf("%.0f", Odds.Bot.Risk[1]) + "%")
		}()
	}

	rl_label := widget.NewLabel("Risk Low: " + fmt.Sprintf("%.0f", Odds.Bot.Risk[0]) + "%")
	rl_slider := widget.NewSlider(1, 88)
	rl_slider.SetValue(Odds.Bot.Risk[0])
	rl_slider.OnChanged = func(f float64) {
		go func() {
			if Odds.Bot.Risk[1] <= f {
				rm_slider.SetValue(f + 1)
			}

			Odds.Bot.Risk[0] = f
			rl_label.SetText("Risk Low: " + fmt.Sprintf("%.0f", Odds.Bot.Risk[0]) + "%")
		}()
	}

	bh_label := widget.NewLabel("Bet High: " + fmt.Sprintf("%.0f", Odds.Bot.Bet[2]) + "x")
	bh_slider := widget.NewSlider(1, 100)
	bh_slider.SetValue(Odds.Bot.Bet[2])
	bh_slider.OnChanged = func(f float64) {
		go func() {
			if f < Odds.Bot.Bet[1] {
				Odds.Bot.Bet[2] = Odds.Bot.Bet[1] + 1
				bh_slider.SetValue(Odds.Bot.Bet[2])
			} else {
				Odds.Bot.Bet[2] = f
			}

			min := float64(MinBet()) / 100000
			if min == 0 {
				min = 0.1
			}

			if Odds.Bot.Max < (min*Odds.Bot.Bet[2])*Odds.Bot.Aggr {
				Odds.Bot.Max = (min * Odds.Bot.Bet[2]) * Odds.Bot.Aggr
				mb_slider.SetValue(Odds.Bot.Max)
			}

			bh_label.SetText("Bet High: " + fmt.Sprintf("%.0f", Odds.Bot.Bet[2]) + "x")
		}()
	}

	bm_label := widget.NewLabel("Bet Medium: " + fmt.Sprintf("%.0f", Odds.Bot.Bet[1]) + "x")
	bm_slider := widget.NewSlider(1, 99)
	bm_slider.SetValue(Odds.Bot.Bet[1])
	bm_slider.OnChanged = func(f float64) {
		go func() {
			Odds.Bot.Bet[1] = f
			if f <= Odds.Bot.Bet[0] {
				Odds.Bot.Bet[1] = Odds.Bot.Bet[0] + 1
				bm_slider.SetValue(Odds.Bot.Bet[1])
			}

			if f >= Odds.Bot.Bet[2] {
				Odds.Bot.Bet[2] = f + 1
				bh_slider.SetValue(Odds.Bot.Bet[2])
			}

			bm_label.SetText("Bet Medium: " + fmt.Sprintf("%.0f", Odds.Bot.Bet[1]) + "x")
		}()
	}

	bl_label := widget.NewLabel("Bet Low: " + fmt.Sprintf("%.0f", Odds.Bot.Bet[0]) + "x")
	bl_slider := widget.NewSlider(1, 98)
	bl_slider.SetValue(Odds.Bot.Bet[0])
	bl_slider.OnChanged = func(f float64) {
		go func() {
			if Odds.Bot.Bet[1] <= f {
				bm_slider.SetValue(f + 1)
			}

			Odds.Bot.Bet[0] = f
			bl_label.SetText("Bet Low: " + fmt.Sprintf("%.0f", Odds.Bot.Bet[0]) + "x")
		}()
	}

	luck_label := widget.NewLabel("Luck: " + fmt.Sprintf("%.2f", Odds.Bot.Luck))
	luck_slider := widget.NewSlider(0, 10)
	luck_slider.Step = 0.25
	luck_slider.SetValue(Odds.Bot.Luck)
	luck_slider.OnChanged = func(f float64) {
		go func() {
			Odds.Bot.Luck = f
			luck_label.SetText("Luck: " + fmt.Sprintf("%.2f", f))
		}()
	}

	random_label := widget.NewLabel("Randomize: Off")
	if Odds.Bot.Random[0] == 0 {
		random_label.SetText("Randomize: Off")
	} else {
		random_label.SetText("Randomize: " + fmt.Sprintf("%.2f", Odds.Bot.Random[0]))
	}

	random_opts := widget.NewRadioGroup([]string{"Risk", "Bet", "Both"}, func(s string) {
		switch s {
		case "Risk":
			Odds.Bot.Random[1] = 1
		case "Bet":
			Odds.Bot.Random[1] = 2
		case "Both":
			Odds.Bot.Random[1] = 3
		default:
			Odds.Bot.Random[1] = 0
		}
	})

	setRandomOpts(random_opts)

	random_slider := widget.NewSlider(0, 10)
	random_slider.Step = 0.25
	random_slider.SetValue(Odds.Bot.Random[0])
	random_slider.OnChanged = func(f float64) {
		go func() {
			Odds.Bot.Random[0] = f
			if f >= 0.5 {
				random_label.SetText("Randomize: " + fmt.Sprintf("%.2f", f))
				random_opts.Enable()
			} else {
				Odds.Bot.Random[0] = 0
				Odds.Bot.Random[1] = 0
				random_label.SetText("Randomize: Off")
				random_opts.SetSelected("")
				random_opts.Disable()
			}
		}()
	}

	slow_label := widget.NewLabel("Slowplay: " + fmt.Sprintf("%.0f", Odds.Bot.Slow))
	slow_slider := widget.NewSlider(1, 5)
	slow_slider.SetValue(Odds.Bot.Slow)
	slow_slider.OnChanged = func(f float64) {
		go func() {
			Odds.Bot.Slow = f
			slow_label.SetText("Slowplay: " + fmt.Sprintf("%.0f", f))
		}()
	}

	aggr_label := widget.NewLabel("Aggression: " + fmt.Sprintf("%.0f", Odds.Bot.Aggr))
	aggr_slider := widget.NewSlider(1, 5)
	aggr_slider.SetValue(Odds.Bot.Aggr)
	aggr_slider.OnChanged = func(f float64) {
		go func() {
			Odds.Bot.Aggr = f
			min := float64(MinBet()) / 100000
			if min == 0 {
				min = 0.1
			}

			if Odds.Bot.Max < (min*Odds.Bot.Bet[2])*Odds.Bot.Aggr {
				Odds.Bot.Max = (min * Odds.Bot.Bet[2]) * Odds.Bot.Aggr
				mb_slider.SetValue(Odds.Bot.Max)
			}

			aggr_label.SetText("Aggression: " + fmt.Sprintf("%.0f", f))
		}()
	}

	rem := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "delete"), func() {
		if entry.Text != "" {
			var new []Bot_config
			for i := range Stats.Bots {
				if Stats.Bots[i].Name == entry.Text {
					log.Println("[dReams] Deleting bot config")
					if i > 0 {
						new = append(Stats.Bots[0:i], Stats.Bots[i+1:]...)
						config_opts = append(config_opts[0:i], config_opts[i+1:]...)
						break
					} else {
						if len(config_opts) < 2 {
							new = nil
							config_opts = []string{}
						} else {
							new = Stats.Bots[1:]
							config_opts = append(config_opts[1:2], config_opts[2:]...)
						}
						break
					}
				}
			}

			Stats.Bots = new
			WriteHolderoStats(Stats)
			entry.SetOptions(config_opts)
			entry.SetText("")
		}
	})

	reset := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "viewRefresh"), func() {
		autoBetDefault()
		mb_slider.SetValue(Odds.Bot.Max)
		rh_slider.SetValue(Odds.Bot.Risk[2])
		rm_slider.SetValue(Odds.Bot.Risk[1])
		rl_slider.SetValue(Odds.Bot.Risk[0])
		bh_slider.SetValue(Odds.Bot.Bet[2])
		bm_slider.SetValue(Odds.Bot.Bet[1])
		bl_slider.SetValue(Odds.Bot.Bet[0])
		luck_slider.SetValue(Odds.Bot.Luck)
		random_slider.SetValue(Odds.Bot.Random[0])
		slow_slider.SetValue(Odds.Bot.Slow)
		aggr_slider.SetValue(Odds.Bot.Aggr)
		random_opts.SetSelected("")
		entry.SetText("")
		Odds.Bot.Name = ""
	})

	save := widget.NewButton("Save", func() {
		if entry.Text != "" {
			var ex bool
			for i := range Stats.Bots {
				if entry.Text == Stats.Bots[i].Name {
					ex = true
					log.Println("[dReams] Bot config name exists")
				}
			}

			if !ex {
				Stats.Bots = append(Stats.Bots, Odds.Bot)
				if WriteHolderoStats(Stats) {
					config_opts = append(config_opts, entry.Text)
					entry.SetOptions(config_opts)
					log.Println("[dReams] Saved bot config")
				}
			}
		}
	})

	entry.OnChanged = func(s string) {
		if s != "" {
			Odds.Bot.Name = entry.Text
			for i := range config_opts {
				if s == config_opts[i] {
					for r := range Stats.Bots {
						if Stats.Bots[r].Name == config_opts[i] {
							SetBotConfig(Stats.Bots[r])
							mb_slider.SetValue(Odds.Bot.Max)
							rh_slider.SetValue(Odds.Bot.Risk[2])
							rm_slider.SetValue(Odds.Bot.Risk[1])
							rl_slider.SetValue(Odds.Bot.Risk[0])
							bh_slider.SetValue(Odds.Bot.Bet[2])
							bm_slider.SetValue(Odds.Bot.Bet[1])
							bl_slider.SetValue(Odds.Bot.Bet[0])
							luck_slider.SetValue(Odds.Bot.Luck)
							random_slider.SetValue(Odds.Bot.Random[0])
							slow_slider.SetValue(Odds.Bot.Slow)
							aggr_slider.SetValue(Odds.Bot.Aggr)
							setRandomOpts(random_opts)
						}
					}
				}
			}
		}
	}

	enable := widget.NewCheck("Auto Bet Enabled", func(b bool) {
		if b {
			Odds.Run = true
			if check.Checked {
				check.SetChecked(false)
			}
			check.Disable()
			deal.SetChecked(true)
		} else {
			Odds.Run = false
			check.Enable()
			if deal.Checked {
				deal.SetChecked(false)
			}
		}
	})

	if Odds.Run {
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

	Odds.Label = widget.NewLabel("")
	Odds.Label.Wrapping = fyne.TextWrapWord
	scroll := container.NewVScroll(Odds.Label)
	odds_button := widget.NewButton("Odds", func() {
		odds, future := MakeOdds()
		BetLogic(odds, future, false)
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
			stats_label.SetText("Total Player Stats\n\nWins: " + strconv.Itoa(Stats.Player.Win) + "\n\nLost: " + strconv.Itoa(Stats.Player.Lost) +
				"\n\nFolded: " + strconv.Itoa(Stats.Player.Fold) + "\n\nPush: " + strconv.Itoa(Stats.Player.Push) +
				"\n\nWagered: " + fmt.Sprintf("%.5f", Stats.Player.Wagered) + "\n\nEarnings: " + fmt.Sprintf("%.5f", Stats.Player.Earnings))

			if Odds.Bot.Name != "" {
				for i := range Stats.Bots {
					if Odds.Bot.Name == Stats.Bots[i].Name {
						stats_label.SetText(stats_label.Text + "\n\n\nBot Stats\n\nBot: " + Odds.Bot.Name + "\n\nWins: " + strconv.Itoa(Stats.Bots[i].Stats.Win) +
							"\n\nLost: " + strconv.Itoa(Stats.Bots[i].Stats.Lost) + "\n\nFolded: " + strconv.Itoa(Stats.Bots[i].Stats.Fold) + "\n\nPush: " + strconv.Itoa(Stats.Bots[i].Stats.Push) +
							"\n\nWagered: " + fmt.Sprintf("%.5f", Stats.Bots[i].Stats.Wagered) + "\n\nEarnings: " + fmt.Sprintf("%.5f", Stats.Bots[i].Stats.Earnings))
					}
				}
			}
		}
	}

	go func() {
		for rpc.Wallet.IsConnected() {
			time.Sleep(1 * time.Second)
		}

		button.Show()
		bm.Close()
	}()

	var err error
	var img image.Image
	var rast *canvas.Raster
	if img, _, err = image.Decode(bytes.NewReader(dreams.Theme.Img.Resource.Content())); err != nil {
		if img, _, err = image.Decode(bytes.NewReader(bundle.ResourceBackgroundPng.Content())); err != nil {
			log.Printf("[holderoTools] Fallback %s\n", err)
			source := image.Rect(2, 2, 4, 4)

			rast = canvas.NewRasterFromImage(source)
		} else {
			rast = canvas.NewRasterFromImage(img)
		}
	} else {
		rast = canvas.NewRasterFromImage(img)
	}

	bm.SetContent(
		container.New(layout.NewMaxLayout(),
			rast,
			bundle.Alpha180,
			tabs))
	bm.Show()
}

func DisableHolderoTools() {
	Odds.Enabled = false
	Settings.Tools.Hide()
	if len(Backs.Select.Options) > 2 || len(Faces.Select.Options) > 2 {
		cards := false
		for _, f := range Faces.Select.Options {
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
			for _, b := range Backs.Select.Options {
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
			Odds.Enabled = true
			Settings.Tools.Show()
			if !dreams.FileExists("config/stats.json", "dReams") {
				WriteHolderoStats(Stats)
				log.Println("[dReams] Created stats.json")
			} else {
				Stats = readSavedStats()
			}
		}
	}
}

// Reading saved Holdero stats from config file
func readSavedStats() (saved Player_stats) {
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
