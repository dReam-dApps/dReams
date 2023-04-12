package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/SixofClubsss/dReams/bundle"
)

var defaultTheme fyne.Theme
var _ fyne.Theme = (*configurableTheme)(nil)

type configurableTheme struct {
	colors map[fyne.ThemeColorName]color.Color
	fonts  map[fyne.TextStyle]fyne.Resource
	sizes  map[fyne.ThemeSizeName]float32
}

func Theme() fyne.Theme {
	purple := color.RGBA{105, 90, 205, 210}
	if defaultTheme == nil {
		defaultTheme = &configurableTheme{
			colors: map[fyne.ThemeColorName]color.Color{
				theme.ColorNameBackground:      color.Black,
				theme.ColorNameButton:          color.RGBA{45, 45, 45, 180},
				theme.ColorNameDisabled:        color.White,
				theme.ColorNameDisabledButton:  color.Transparent,
				theme.ColorNameError:           color.NRGBA{R: 0xf4, G: 0x43, B: 0x36, A: 0xff},
				theme.ColorNameFocus:           color.NRGBA{R: 0x88, G: 0xff, B: 0xff, A: 0x22}, // entry highlight
				theme.ColorNameForeground:      color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, // selected tab
				theme.ColorNameHover:           color.NRGBA{R: 0x88, G: 0xff, B: 0xff, A: 0x22}, //button hightlight
				theme.ColorNameInputBackground: color.RGBA{75, 75, 75, 180},
				theme.ColorNamePlaceHolder:     color.NRGBA{R: 0xaa, G: 0xaa, B: 0xaa, A: 0xff},
				theme.ColorNamePressed:         color.NRGBA{A: 0x33},
				theme.ColorNamePrimary:         purple, // tab select color, progress bar
				theme.ColorNameScrollBar:       purple,
				theme.ColorNameSelection:       purple,
				theme.ColorNameShadow:          color.NRGBA{A: 0x88},
			},
			fonts: map[fyne.TextStyle]fyne.Resource{
				{}:                         bundle.ResourceUbuntuRTtf,
				{Bold: true}:               bundle.ResourceVarelaRoundRegularTtf,
				{Bold: true, Italic: true}: bundle.ResourceUbuntuRTtf,
				{Italic: true}:             bundle.ResourceUbuntuRTtf,
				{Monospace: true}:          bundle.ResourceUbuntuRTtf,
			},
			sizes: map[fyne.ThemeSizeName]float32{
				theme.SizeNameInlineIcon:         float32(20),
				theme.SizeNamePadding:            float32(4),
				theme.SizeNameScrollBar:          float32(16),
				theme.SizeNameScrollBarSmall:     float32(3),
				theme.SizeNameSeparatorThickness: float32(1),
				theme.SizeNameText:               float32(14),
				theme.SizeNameHeadingText:        float32(24),
				theme.SizeNameSubHeadingText:     float32(18),
				theme.SizeNameCaptionText:        float32(12),
				theme.SizeNameInputBorder:        float32(2),
			},
		}
	}
	return defaultTheme
}

func (t *configurableTheme) Color(n fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return t.colors[n]
}

func (t *configurableTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.fonts[style]
}

func (t *configurableTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (t *configurableTheme) Size(s fyne.ThemeSizeName) float32 {
	return t.sizes[s]
}
