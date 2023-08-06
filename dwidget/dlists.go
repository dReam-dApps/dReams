package dwidget

import (
	"sort"

	"fyne.io/fyne/v2/widget"
)

type Lists struct {
	List *widget.List
	All  []uint64
}

// Sort All index
func (l *Lists) SortIndex() {
	sort.Slice(l.All, func(i, j int) bool { return l.All[i] < l.All[j] })
}

// Remove index from All
func (l *Lists) RemoveIndex(e uint64) {
	index := -1
	for i, num := range l.All {
		if num == e {
			index = i
			break
		}
	}

	if index != -1 {
		l.All = append(l.All[:index], l.All[index+1:]...)
	}

	l.List.Refresh()
}

// Check if exists in All
func (l *Lists) Exists(k uint64) bool {
	for _, num := range l.All {
		if num == k {
			return true
		}
	}

	return false
}
