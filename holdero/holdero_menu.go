package holdero

import (
	"fmt"
	"image/color"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/menu"
	"github.com/dReam-dApps/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func HolderoIndicator() (ind menu.DreamsIndicator) {
	purple := color.RGBA{105, 90, 205, 210}
	blue := color.RGBA{31, 150, 200, 210}
	alpha := &color.RGBA{0, 0, 0, 0}

	ind.Img = canvas.NewImageFromResource(ResourcePokerBotIconPng)
	ind.Img.SetMinSize(fyne.NewSize(30, 30))
	ind.Rect = canvas.NewRectangle(alpha)
	ind.Rect.SetMinSize(fyne.NewSize(36, 36))

	ind.Animation = canvas.NewColorRGBAAnimation(purple, blue,
		time.Second*3, func(c color.Color) {
			if Odds.IsRunning() {
				ind.Rect.FillColor = c
				ind.Img.Show()
				canvas.Refresh(ind.Rect)
			} else {
				ind.Rect.FillColor = alpha
				ind.Img.Hide()
				canvas.Refresh(ind.Rect)
			}
		})

	ind.Animation.RepeatCount = fyne.AnimationRepeatForever
	ind.Animation.AutoReverse = true

	return
}

// Holdero owner control objects, left section
func ownersBoxLeft(obj []fyne.CanvasObject, tabs *container.AppTabs) fyne.CanvasObject {
	players := []string{"Players", "Close Table", "2 Players", "3 Players", "4 Players", "5 Players", "6 Players"}
	player_select := widget.NewSelect(players, func(s string) {})
	player_select.SetSelectedIndex(0)

	blinds_entry := dwidget.DeroAmtEntry("Big Blind: ", 0.1, 1)
	blinds_entry.SetPlaceHolder("Dero:")
	blinds_entry.SetText("Big Blind: 0.0")
	blinds_entry.Validator = validation.NewRegexp(`^(Big Blind: )\d{1,}\.\d{0,1}$|^(Big Blind: )\d{1,}$`, "Int or float required")
	blinds_entry.OnChanged = func(s string) {
		if blinds_entry.Validate() != nil {
			blinds_entry.SetText("Big Blind: 0.0")
			Poker.owner.blind_amount = 0
		} else {
			trimmed := strings.Trim(s, "Biglnd: ")
			if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
				if uint64(f*100000)%10000 == 0 {
					blinds_entry.SetText(blinds_entry.Prefix + strconv.FormatFloat(roundFloat(f, 1), 'f', int(blinds_entry.Decimal), 64))
					Poker.owner.blind_amount = uint64(roundFloat(f*100000, 1))
				} else {
					blinds_entry.SetText(blinds_entry.Prefix + strconv.FormatFloat(roundFloat(f, 1), 'f', int(blinds_entry.Decimal), 64))
				}
			}
		}
	}

	ante_entry := dwidget.DeroAmtEntry("Ante: ", 0.1, 1)
	ante_entry.SetPlaceHolder("Ante:")
	ante_entry.SetText("Ante: 0.0")
	ante_entry.Validator = validation.NewRegexp(`^(Ante: )\d{1,}\.\d{0,1}$|^(Ante: )\d{1,}$`, "Int or float required")
	ante_entry.OnChanged = func(s string) {
		if ante_entry.Validate() != nil {
			ante_entry.SetText("Ante: 0.0")
			Poker.owner.ante_amount = 0
		} else {
			trimmed := strings.Trim(s, ante_entry.Prefix)
			if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
				if uint64(f*100000)%10000 == 0 {
					ante_entry.SetText(ante_entry.Prefix + strconv.FormatFloat(roundFloat(f, 1), 'f', int(ante_entry.Decimal), 64))
					Poker.owner.ante_amount = uint64(roundFloat(f*100000, 1))
				} else {
					ante_entry.SetText(ante_entry.Prefix + strconv.FormatFloat(roundFloat(f, 1), 'f', int(ante_entry.Decimal), 64))
				}
			}
		}
	}

	options := []string{"DERO", "ASSET"}
	Poker.owner.chips = widget.NewRadioGroup(options, nil)
	Poker.owner.chips.SetSelected("DERO")
	Poker.owner.chips.Horizontal = true
	Poker.owner.chips.OnChanged = func(s string) {
		if s == "ASSET" {
			blinds_entry.Increment = 1
			blinds_entry.Decimal = 0
			blinds_entry.SetText("0")
			blinds_entry.Refresh()

			ante_entry.Increment = 1
			ante_entry.Decimal = 0
			ante_entry.SetText("0")
			ante_entry.Refresh()
		} else {
			blinds_entry.Increment = 0.1
			blinds_entry.Decimal = 1
			blinds_entry.Refresh()

			ante_entry.Increment = 0.1
			ante_entry.Decimal = 1
			ante_entry.Refresh()
		}
	}

	set_button := widget.NewButton("Set Table", func() {
		bb := Poker.owner.blind_amount
		sb := Poker.owner.blind_amount / 2
		ante := Poker.owner.ante_amount
		if menu.Username != "" {
			SetTable(player_select.SelectedIndex(), bb, sb, ante, Poker.owner.chips.Selected, menu.Username, Settings.AvatarUrl)
		}
	})

	clean_entry := dwidget.DeroAmtEntry("Clean: ", 1, 0)
	clean_entry.AllowFloat = false
	clean_entry.SetPlaceHolder("Atomic:")
	clean_entry.SetText("Clean: 0")
	clean_entry.Validator = validation.NewRegexp(`^(Clean: )\d{1,}`, "Int required")
	clean_entry.OnChanged = func(s string) {
		if clean_entry.Validate() != nil {
			clean_entry.SetText("Clean: 0")
		}
	}

	clean_button := widget.NewButton("Clean Table", func() {
		trimmed := strings.Trim(clean_entry.Text, "Clean: ")
		c, err := strconv.Atoi(trimmed)
		if err == nil {
			CleanTable(uint64(c))
		} else {
			log.Println("[dReams] Invalid Clean Amount")
		}
	})

	Poker.owner.timeout = widget.NewButton("Timeout", func() {
		obj[1] = timeOutConfirm(obj, tabs)
		obj[1].Refresh()
	})

	force := widget.NewButton("Force Start", func() {
		ForceStat()
	})

	players_items := container.NewAdaptiveGrid(2, player_select, layout.NewSpacer())
	blind_items := container.NewAdaptiveGrid(2, blinds_entry, Poker.owner.chips)
	ante_items := container.NewAdaptiveGrid(2, ante_entry, set_button)
	clean_items := container.NewAdaptiveGrid(2, clean_entry, clean_button)
	time_items := container.NewAdaptiveGrid(2, Poker.owner.timeout, force)

	Poker.owner.owners_left = container.NewVBox(players_items, blind_items, ante_items, clean_items, time_items)
	Poker.owner.owners_left.Hide()

	return Poker.owner.owners_left
}

// Holdero owner control objects, middle section
func ownersBoxMid() fyne.CanvasObject {
	kick_label := widget.NewLabel("      Auto Kick after")
	k_times := []string{"Off", "2m", "5m"}
	auto_remove := widget.NewSelect(k_times, func(s string) {
		switch s {
		case "Off":
			Times.Kick = 0
		case "2m":
			Times.Kick = 120
		case "5m":
			Times.Kick = 300
		default:
			Times.Kick = 0
		}
	})
	auto_remove.PlaceHolder = "Kick after"

	pay_label := widget.NewLabel("      Payout Delay")
	p_times := []string{"30s", "60s"}
	delay := widget.NewSelect(p_times, func(s string) {
		switch s {
		case "30s":
			Times.Delay = 30
		case "60s":
			Times.Delay = 60
		default:
			Times.Delay = 30
		}
	})
	delay.PlaceHolder = "Payout delay"

	kick := container.NewVBox(layout.NewSpacer(), kick_label, auto_remove)
	pay := container.NewVBox(layout.NewSpacer(), pay_label, delay)

	Poker.owner.owners_mid = container.NewAdaptiveGrid(2, kick, pay)
	Poker.owner.owners_mid.Hide()

	return Poker.owner.owners_mid
}

// Holdero table icon image with frame
func tableIcon(r fyne.Resource) *fyne.Container {
	Table.Stats.Image.SetMinSize(fyne.NewSize(100, 100))
	Table.Stats.Image.Resize(fyne.NewSize(96, 96))
	Table.Stats.Image.Move(fyne.NewPos(8, 3))

	frame := canvas.NewImageFromResource(r)
	frame.Resize(fyne.NewSize(100, 100))
	frame.Move(fyne.NewPos(5, 0))

	cont := container.NewWithoutLayout(&Table.Stats.Image, frame)

	return cont
}

// Holdero table stats display objects
func displayTableStats() fyne.CanvasObject {
	Table.Stats.Name = canvas.NewText(" Name: ", bundle.TextColor)
	Table.Stats.Desc = canvas.NewText(" Description: ", bundle.TextColor)
	Table.Stats.Version = canvas.NewText(" Table Version: ", bundle.TextColor)
	Table.Stats.Last = canvas.NewText(" Last Move: ", bundle.TextColor)
	Table.Stats.Seats = canvas.NewText(" Table Closed ", bundle.TextColor)

	Table.Stats.Name.TextSize = 18
	Table.Stats.Desc.TextSize = 18
	Table.Stats.Version.TextSize = 18
	Table.Stats.Last.TextSize = 18
	Table.Stats.Seats.TextSize = 18

	Poker.Stats_box = *container.NewVBox(Table.Stats.Name, Table.Stats.Desc, Table.Stats.Version, Table.Stats.Last, Table.Stats.Seats, tableIcon(nil))

	return &Poker.Stats_box
}

// Confirmation of manual Holdero timeout
func timeOutConfirm(obj []fyne.CanvasObject, reset *container.AppTabs) fyne.CanvasObject {
	var confirm_display = widget.NewLabel("")
	confirm_display.Wrapping = fyne.TextWrapWord
	confirm_display.Alignment = fyne.TextAlignCenter

	confirm_display.SetText("Confirm Time Out on Current Player")

	cancel_button := widget.NewButton("Cancel", func() {
		obj[1] = reset
		obj[1].Refresh()
	})
	confirm_button := widget.NewButton("Confirm", func() {
		TimeOut()
		obj[1] = reset
		obj[1].Refresh()
	})

	display := container.NewVBox(layout.NewSpacer(), confirm_display, layout.NewSpacer())
	options := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewBorder(nil, options, nil, nil, display)

	return container.NewMax(bundle.Alpha120, content)
}

// Confirmation for Holdero contract installs
func holderoMenuConfirm(c int, obj []fyne.CanvasObject, tabs *container.AppTabs) fyne.CanvasObject {
	gas_fee := 0.3
	unlock_fee := float64(rpc.UnlockFee) / 100000
	var text string
	switch c {
	case 1:
		Poker.Holdero_unlock.Hide()
		text = `You are about to unlock and install your first Holdero Table
		
To help support the project, there is a ` + fmt.Sprintf("%.5f", unlock_fee) + ` DERO donation attached to preform this action

Unlocking a Holdero table gives you unlimited access to table uploads and all base level owner features

Total transaction will be ` + fmt.Sprintf("%0.5f", unlock_fee+gas_fee) + ` DERO (0.30000 gas fee for contract install)


Select a public or private table

Public will show up in indexed list of tables

Private will not show up in the list

All standard tables can use dReams or DERO


HGC holders can choose to install a HGC table

Public table that uses HGC or DERO`
	case 2:
		Poker.Holdero_new.Hide()
		text = `You are about to install a new Holdero table

Gas fee to install new table is 0.30000 DERO


Select a public or private table

Public will show up in indexed list of tables

Private will not show up in the list

All standard tables can use dReams or DERO


HGC holders can choose to install a HGC table

Public table that uses HGC or DERO`
	}

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	var choice *widget.Select
	confirm_button := widget.NewButton("Confirm", func() {
		if choice.SelectedIndex() < 3 && choice.SelectedIndex() >= 0 {
			uploadHolderoContract(choice.SelectedIndex())
		}

		if c == 2 {
			Poker.Holdero_new.Show()
		}

		obj[1] = tabs
		obj[1].Refresh()
	})

	options := []string{"Public", "Private"}
	if hgc := rpc.TokenBalance(rpc.HgcSCID); hgc > 0 {
		options = append(options, "HGC")
	}

	choice = widget.NewSelect(options, func(s string) {
		if s == "Public" || s == "Private" || s == "HGC" {
			confirm_button.Show()
		} else {
			confirm_button.Hide()
		}
	})

	cancel_button := widget.NewButton("Cancel", func() {
		switch c {
		case 1:
			Poker.Holdero_unlock.Show()
		case 2:
			Poker.Holdero_new.Show()
		default:

		}

		obj[1] = tabs
		obj[1].Refresh()
	})

	confirm_button.Hide()

	left := container.NewVBox(confirm_button)
	right := container.NewVBox(cancel_button)
	buttons := container.NewAdaptiveGrid(2, left, right)
	actions := container.NewVBox(choice, buttons)
	info_box := container.NewVBox(layout.NewSpacer(), label, layout.NewSpacer())

	content := container.NewBorder(nil, actions, nil, nil, info_box)

	go func() {
		for rpc.IsReady() {
			time.Sleep(time.Second)
		}

		obj[1] = tabs
		obj[1].Refresh()
	}()

	return container.NewMax(content)
}
