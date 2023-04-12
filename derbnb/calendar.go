package derbnb

import (
	"strings"
	"time"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type tirp_date struct {
	arriving  *canvas.Text
	departing *canvas.Text
}

type add_dates struct {
	starting []*widget.Entry
	ending   []*widget.Entry
}

// Set arriving date before departing date for single pair of entries
func (d *tirp_date) onSelected(t time.Time) {
	if t.Unix() > time.Now().Unix() {
		if t.Before(start_date) || start_date.IsZero() {
			start_date = t
		} else {
			end_date = t
		}

		if !start_date.IsZero() {
			d.arriving.Text = ("Arriving: " + start_date.Format(TIME_FORMAT))
			d.arriving.Refresh()
		}

		if !end_date.IsZero() {
			d.departing.Text = ("Departing: " + end_date.Format(TIME_FORMAT))
			d.departing.Refresh()
		}
	}
}

// Set arriving date before departing date for groups of entries
func (d *add_dates) onSelected(t time.Time) {
	if t.Unix() > time.Now().Unix() {
		for i, wid := range d.starting {
			if !wid.Hidden && wid.Text == "" {
				if d.ending[i].Text != "" {
					trim := strings.TrimPrefix(d.ending[i].Text, "Ending: ")
					if date, err := time.Parse(TIME_FORMAT, trim); err == nil {
						if date.Add(24*time.Hour).Unix() > t.Unix() {
							wid.SetText("Starting: " + t.Format(TIME_FORMAT))
						}
					}
					return
				}
				wid.SetText("Starting: " + t.Format(TIME_FORMAT))
				return
			}
		}

		for i, wid := range d.ending {
			if !wid.Hidden && wid.Text == "" {
				trim := strings.TrimPrefix(d.starting[i].Text, "Starting: ")
				if date, err := time.Parse(TIME_FORMAT, trim); err == nil {
					if date.Add(24*time.Hour).Unix() < t.Unix() {
						wid.SetText("Ending: " + t.Format(TIME_FORMAT))
						return
					}
				}
			}
		}
	}
}
