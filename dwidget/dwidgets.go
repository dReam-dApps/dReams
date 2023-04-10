package dwidget

import (
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	xwidget "fyne.io/x/fyne/widget"
)

type TenthAmt struct {
	xwidget.NumericalEntry
	Prefix string
}

// Create new numerical entry with change of 0.1 on up or down key stroke
//   - If entry does not require prefix, pass ""
func TenthAmtEntry(prefix string) *TenthAmt {
	entry := &TenthAmt{}
	entry.ExtendBaseWidget(entry)
	entry.AllowFloat = true
	entry.Prefix = prefix

	return entry
}

// Accepts whole number or '.'
func (e *TenthAmt) TypedRune(r rune) {
	if r >= '0' && r <= '9' {
		e.Entry.TypedRune(r)
		return
	}

	if e.AllowFloat && r == '.' {
		e.Entry.TypedRune(r)
	}
}

// Increment of 0.1 on TypedKey
func (e *TenthAmt) TypedKey(k *fyne.KeyEvent) {
	value := strings.Trim(e.Entry.Text, e.Prefix)
	switch k.Name {
	case fyne.KeyUp:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			e.Entry.SetText(e.Prefix + strconv.FormatFloat(float64(f+0.1), 'f', 1, 64))
		}
	case fyne.KeyDown:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			if f >= 0.1 {
				e.Entry.SetText(e.Prefix + strconv.FormatFloat(float64(f-0.1), 'f', 1, 64))
			}
		}
	}
	e.Entry.TypedKey(k)
}

type WholeAmt struct {
	xwidget.NumericalEntry
	Prefix string
}

// Create new numerical entry with change of 1 on up or down key stroke
//   - If entry does not require prefix, pass ""
func WholeAmtEntry(prefix string) *WholeAmt {
	entry := &WholeAmt{}
	entry.ExtendBaseWidget(entry)
	entry.Prefix = prefix

	return entry
}

// Only accept whole number
func (e *WholeAmt) TypedRune(r rune) {
	if r >= '0' && r <= '9' {
		e.Entry.TypedRune(r)
		return
	}
}

// Increment of 1 on TypedKey
func (e *WholeAmt) TypedKey(k *fyne.KeyEvent) {
	value := strings.Trim(e.Entry.Text, e.Prefix)
	switch k.Name {
	case fyne.KeyUp:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			e.Entry.SetText(e.Prefix + strconv.FormatFloat(float64(f+1), 'f', 0, 64))
		}
	case fyne.KeyDown:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			if f >= 0.1 {
				e.Entry.SetText(e.Prefix + strconv.FormatFloat(float64(f-1), 'f', 0, 64))
			}
		}
	}
	e.Entry.TypedKey(k)
}
