package dwidget

import (
	"sort"

	"fyne.io/fyne/v2/widget"
)

// Fyne List widget, arrays for string or uint64 content
type Lists struct {
	All   []uint64
	SCIDs []string
	*widget.List
}

// Sort All index
//   - Pass reverse true to sort in reverse order
func (l *Lists) SortIndex(reverse bool) {
	if reverse {
		sort.Slice(l.All, func(i, j int) bool { return l.All[i] > l.All[j] })
		return
	}

	sort.Slice(l.All, func(i, j int) bool { return l.All[i] < l.All[j] })
}

// Remove u from All
func (l *Lists) RemoveIndex(u uint64) {
	index := -1
	for i, num := range l.All {
		if num == u {
			index = i
			break
		}
	}

	if index != -1 {
		l.All = append(l.All[:index], l.All[index+1:]...)
	}

	if l.List != nil {
		l.List.Refresh()
	}
}

// Check if u exists in All
func (l *Lists) ExistsIndex(u uint64) bool {
	for _, num := range l.All {
		if num == u {
			return true
		}
	}

	return false
}

// Sort SCIDs array
//   - Pass reverse true to sort in reverse order
func (l *Lists) SortSCIDs(reverse bool) {
	if reverse {
		sort.Slice(l.SCIDs, func(i, j int) bool { return l.SCIDs[i] > l.SCIDs[j] })
		return
	}

	sort.Slice(l.SCIDs, func(i, j int) bool { return l.SCIDs[i] < l.SCIDs[j] })
}

// Remove scid from SCIDs
func (l *Lists) RemoveSCID(scid string) {
	index := -1
	for i, sc := range l.SCIDs {
		if sc == scid {
			index = i
			break
		}
	}

	if index != -1 {
		l.SCIDs = append(l.SCIDs[:index], l.SCIDs[index+1:]...)
	}

	if l.List != nil {
		l.List.Refresh()
	}
}

// Check if scid exists in SCIDs
func (l *Lists) ExistsSCID(scid string) bool {
	for _, sc := range l.SCIDs {
		if sc == scid {
			return true
		}
	}

	return false
}
