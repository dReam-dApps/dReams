package baccarat

import (
	dreams "github.com/SixofClubsss/dReams"
	"github.com/SixofClubsss/dReams/dwidget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

// Baccarat tab layout
func LayoutAllItems(b *dreams.DreamsItems, d dreams.DreamsObject) *fyne.Container {
	b.Back = *container.NewWithoutLayout(
		BaccTable(resourceBaccTablePng),
		baccResult(Display.BaccRes))

	b.Front = *clearBaccCards()

	bacc_label := container.NewHBox(b.LeftLabel, layout.NewSpacer(), b.RightLabel)

	b.DApp = container.NewVBox(
		dwidget.LabelColor(bacc_label),
		&b.Back,
		&b.Front,
		layout.NewSpacer(),
		baccaratButtons(d.Window))

	// Main process
	go fetch(b, d)

	return b.DApp
}
