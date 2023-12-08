package gnomes

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/dReam-dApps/dReams/bundle"
)

var _ desktop.Hoverable = (*toolTip)(nil)

type toolTip struct {
	hovered bool
	text    *canvas.Text
	Canvas  fyne.Canvas
	popup   *widget.PopUp
	offset  float32
	*canvas.Rectangle
}

// Display Gnomon heights when hovered
func ToolTip(offset float32, can fyne.Canvas) *toolTip {
	rect := canvas.NewRectangle(color.Black)
	rect.SetMinSize(fyne.NewSize(57, 36))

	return &toolTip{
		text:      canvas.NewText("", bundle.TextColor),
		Canvas:    can,
		popup:     &widget.PopUp{},
		offset:    offset,
		Rectangle: rect,
	}
}

func (t *toolTip) MouseIn(event *desktop.MouseEvent) {
	t.hovered = true
	if t.Canvas != nil {
		if gnomes.Indexer != nil {
			t.text.Text = fmt.Sprintf("%d/%d (%s)", gnomes.GetLastHeight(), gnomes.GetChainHeight(), gnomes.Status())
			t.popup = widget.NewPopUp(t.text, t.Canvas)
		} else {
			t.text.Text = "0/0"
			t.popup = widget.NewPopUp(t.text, t.Canvas)
		}

		if !t.popup.Hidden {
			pos := event.AbsolutePosition
			t.popup.ShowAtPosition(fyne.NewPos(pos.X-t.offset/4.5, pos.Y+t.offset))
			t.Refresh()
		}
	}
}

func (t *toolTip) MouseOut() {
	if t.Canvas != nil {
		t.hovered = false
		t.popup.Hide()
		t.popup = nil
		t.Refresh()
	}
}

func (t *toolTip) MouseMoved(event *desktop.MouseEvent) {}

func (t *toolTip) Refresh() {
	if t.hovered {
		if gnomes.Indexer != nil {
			t.text.Text = fmt.Sprintf("%d/%d (%s)", gnomes.GetLastHeight(), gnomes.GetChainHeight(), gnomes.Status())
		} else {
			t.text.Text = "0/0"
		}
		t.text.Refresh()
	}
}
