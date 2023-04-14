package bundle

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

var defaultTheme fyne.Theme
var _ fyne.Theme = (*configurableTheme)(nil)

type configurableTheme struct {
	colors map[fyne.ThemeColorName]color.Color
	fonts  map[fyne.TextStyle]fyne.Resource
	sizes  map[fyne.ThemeSizeName]float32
}

var Alpha120 = canvas.NewRectangle(color.RGBA{0, 0, 0, 120})
var Alpha150 = canvas.NewRectangle(color.RGBA{0, 0, 0, 150})
var Alpha180 = canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
var TextColor color.Gray16
var AppColor color.Gray16
var purple = color.RGBA{105, 90, 205, 210}
var blue = color.RGBA{31, 150, 200, 210}

var DeroDarkTheme = &configurableTheme{
	colors: map[fyne.ThemeColorName]color.Color{
		theme.ColorNameBackground:      color.Black,
		theme.ColorNameButton:          color.RGBA{45, 45, 45, 201},
		theme.ColorNameDisabled:        blue,
		theme.ColorNameDisabledButton:  color.Transparent,
		theme.ColorNameError:           color.NRGBA{R: 0xf4, G: 0x33, B: 0x25, A: 0xff},
		theme.ColorNameFocus:           blue,                                            // entry highlight
		theme.ColorNameForeground:      color.White,                                     // text color
		theme.ColorNameHover:           color.NRGBA{R: 0x88, G: 0xff, B: 0xff, A: 0x22}, //button hightlight
		theme.ColorNameInputBackground: color.RGBA{75, 75, 75, 201},                     // entry background
		theme.ColorNamePlaceHolder:     color.RGBA{105, 90, 205, 180},
		theme.ColorNamePressed:         purple,
		theme.ColorNamePrimary:         purple, // tab select color, progress bar
		theme.ColorNameScrollBar:       blue,
		theme.ColorNameSelection:       purple,
		theme.ColorNameShadow:          color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x11},
	},
	fonts: map[fyne.TextStyle]fyne.Resource{
		{}:                         ResourceUbuntuRTtf,
		{Bold: true}:               ResourceVarelaRoundRegularTtf,
		{Bold: true, Italic: true}: ResourceUbuntuRTtf,
		{Italic: true}:             ResourceUbuntuRTtf,
		{Monospace: true}:          ResourceUbuntuRTtf,
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

var DeroLightTheme = &configurableTheme{
	colors: map[fyne.ThemeColorName]color.Color{
		theme.ColorNameBackground:      color.White,
		theme.ColorNameButton:          color.NRGBA{R: 0x0e, G: 0x0c, B: 0x0b, A: 0x55},
		theme.ColorNameDisabled:        color.RGBA{105, 90, 205, 240},
		theme.ColorNameDisabledButton:  color.Transparent,
		theme.ColorNameError:           color.NRGBA{R: 0xf4, G: 0x33, B: 0x25, A: 0xff},
		theme.ColorNameFocus:           purple,                                          // entry highlight
		theme.ColorNameForeground:      color.Black,                                     // text color
		theme.ColorNameHover:           color.NRGBA{R: 0x96, G: 0x5a, B: 0xcd, A: 0x45}, //button hightlight
		theme.ColorNameInputBackground: color.NRGBA{R: 0xf0, G: 0xf0, B: 0xf0, A: 0xa5}, // entry background
		theme.ColorNamePlaceHolder:     color.RGBA{31, 150, 200, 180},
		theme.ColorNamePressed:         color.White,
		theme.ColorNamePrimary:         blue, // tab select color, progress bar
		theme.ColorNameScrollBar:       purple,
		theme.ColorNameSelection:       blue,
		theme.ColorNameShadow:          color.NRGBA{R: 0x0e, G: 0x0c, B: 0x0b, A: 0x44},
	},
	fonts: map[fyne.TextStyle]fyne.Resource{
		{}:                         ResourceUbuntuRTtf,
		{Bold: true}:               ResourceVarelaRoundRegularTtf,
		{Bold: true, Italic: true}: ResourceUbuntuRTtf,
		{Italic: true}:             ResourceUbuntuRTtf,
		{Monospace: true}:          ResourceUbuntuRTtf,
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

func DeroTheme(skin color.Gray16) fyne.Theme {
	if defaultTheme == nil {
		if skin == color.White {
			Alpha120 = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x55})
			Alpha150 = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xaa})
			Alpha180 = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x99})
			defaultTheme = DeroLightTheme
			TextColor = color.Black
		} else {
			Alpha150 = canvas.NewRectangle(color.RGBA{0, 0, 0, 120})
			Alpha150 = canvas.NewRectangle(color.RGBA{0, 0, 0, 150})
			Alpha180 = canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
			defaultTheme = DeroDarkTheme
			TextColor = color.White
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
