package bundle

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

var defaultTheme fyne.Theme
var _ fyne.Theme = (*dTheme)(nil)

type dTheme struct{ variant fyne.ThemeVariant }

var Highlight = color.NRGBA{R: 0x88, G: 0xff, B: 0xff, A: 0x22}
var Alpha120 = canvas.NewRectangle(color.RGBA{0, 0, 0, 120})
var Alpha150 = canvas.NewRectangle(color.RGBA{0, 0, 0, 150})
var Alpha180 = canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
var TextColor color.Gray16
var AppColor color.Gray16
var purple = color.RGBA{105, 90, 205, 210}
var blue = color.RGBA{31, 150, 200, 210}

func NewAlpha120() (alpha *canvas.Rectangle) {
	alpha = canvas.NewRectangle(color.RGBA{0, 0, 0, 120})
	if AppColor == color.White {
		alpha = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x55})
	}

	return
}

func NewAlpha150() (alpha *canvas.Rectangle) {
	alpha = canvas.NewRectangle(color.RGBA{0, 0, 0, 150})
	if AppColor == color.White {
		alpha = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xaa})
	}

	return
}

func NewAlpha180() (alpha *canvas.Rectangle) {
	alpha = canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
	if AppColor == color.White {
		alpha = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x99})
	}

	return
}

func DeroTheme(skin color.Gray16) fyne.Theme {
	if skin == color.White {
		Highlight = color.NRGBA{R: 0x96, G: 0x5a, B: 0xcd, A: 0x45}
		Alpha120 = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x55})
		Alpha150 = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xaa})
		Alpha180 = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x99})
		defaultTheme = dTheme{variant: 1}
		TextColor = color.Black
	} else {
		Highlight = color.NRGBA{R: 0x88, G: 0xff, B: 0xff, A: 0x22}
		Alpha120 = canvas.NewRectangle(color.RGBA{0, 0, 0, 120})
		Alpha150 = canvas.NewRectangle(color.RGBA{0, 0, 0, 150})
		Alpha180 = canvas.NewRectangle(color.RGBA{0, 0, 0, 180})
		defaultTheme = dTheme{variant: 0}
		TextColor = color.White
	}

	return defaultTheme
}

func (t dTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		if t.variant == 1 {
			return color.White
		}
		return color.Black

	case theme.ColorNameButton:
		if t.variant == 1 {
			return color.NRGBA{R: 0x0e, G: 0x0c, B: 0x0b, A: 0x55}
		}
		return color.RGBA{45, 45, 45, 201}

	case theme.ColorNameDisabled:
		if t.variant == 1 {
			return color.RGBA{105, 90, 205, 240}
		}
		return blue

	case theme.ColorNameDisabledButton:
		return color.Transparent

	case theme.ColorNameError:
		return color.NRGBA{R: 0xf4, G: 0x33, B: 0x25, A: 0xff}

	case theme.ColorNameFocus:
		// entry highlight
		if t.variant == 1 {
			return purple
		}
		return blue

	case theme.ColorNameForeground:
		// text color
		if t.variant == 1 {
			return color.Black
		}
		return color.White

	case theme.ColorNameHover:
		// button highlight
		if t.variant == 1 {
			return color.NRGBA{R: 0x96, G: 0x5a, B: 0xcd, A: 0x45}
		}
		return color.NRGBA{R: 0x88, G: 0xff, B: 0xff, A: 0x22}

	case theme.ColorNameInputBackground:
		// entry background
		if t.variant == 1 {
			return color.NRGBA{R: 0xf0, G: 0xf0, B: 0xf0, A: 0xa5}
		}
		return color.RGBA{75, 75, 75, 201}

	case theme.ColorNameInputBorder:
		if t.variant == 1 {
			return color.NRGBA{R: 0x96, G: 0x5a, B: 0xcd, A: 0x45}
		}
		return color.NRGBA{R: 0x88, G: 0xff, B: 0xff, A: 0x22}

	case theme.ColorNameMenuBackground:
		if t.variant == 1 {
			return color.NRGBA{R: 0xf0, G: 0xf0, B: 0xf0, A: 0xfa}
		}
		return color.RGBA{75, 75, 75, 250}

	case theme.ColorNameOverlayBackground:
		if t.variant == 1 {
			return color.White
		}
		return color.Black

	case theme.ColorNamePlaceHolder:
		if t.variant == 1 {
			return color.RGBA{31, 150, 200, 180}
		}
		return color.RGBA{105, 90, 205, 180}

	case theme.ColorNamePressed:
		if t.variant == 1 {
			return color.White
		}
		return purple

	case theme.ColorNamePrimary:
		// tab select color, progress bar
		if t.variant == 1 {
			return blue
		}
		return purple

	case theme.ColorNameScrollBar:
		if t.variant == 1 {
			return purple
		}
		return blue

	case theme.ColorNameSeparator:
		if t.variant == 1 {
			return color.RGBA{105, 90, 205, 240}
		}

		return blue

	case theme.ColorNameSelection:
		if t.variant == 1 {
			return blue
		}
		return purple

	case theme.ColorNameShadow:
		if t.variant == 1 {
			return color.NRGBA{R: 0x0e, G: 0x0c, B: 0x0b, A: 0x44}
		}
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x11}

	default:
		return theme.DefaultTheme().Color(name, variant)

	}
}

func (t dTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t dTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Bold {
		return ResourceVarelaRoundRegularTtf
	}

	return ResourceUbuntuRTtf
}

func (t dTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameInlineIcon:
		return float32(20)

	case theme.SizeNameInnerPadding:
		return float32(8)

	case theme.SizeNameLineSpacing:
		return float32(4)

	case theme.SizeNamePadding:
		return float32(5)

	case theme.SizeNameScrollBar:
		return float32(12)

	case theme.SizeNameScrollBarSmall:
		return float32(3)

	case theme.SizeNameSeparatorThickness:
		return float32(1)

	case theme.SizeNameText:
		return float32(15)

	case theme.SizeNameHeadingText:
		return float32(24)

	case theme.SizeNameSubHeadingText:
		return float32(18)

	case theme.SizeNameCaptionText:
		return float32(12)

	case theme.SizeNameInputBorder:
		return float32(3)

	default:
		return theme.DefaultTheme().Size(name)

	}
}
