package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// WaterAITheme implements fyne.Theme for a dark theme matching the existing frontend
type WaterAITheme struct{}

// NewWaterAITheme creates a new Water AI dark theme
func NewWaterAITheme() fyne.Theme {
	return &WaterAITheme{}
}

// Color returns the color for the specified theme color name
func (t *WaterAITheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Force dark variant
	variant = theme.VariantDark

	// Custom colors matching the existing frontend (#191E1B background)
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{R: 0x19, G: 0x1E, B: 0x1B, A: 0xFF} // #191E1B
	case theme.ColorNameButton:
		return color.RGBA{R: 0x2A, G: 0x2F, B: 0x2C, A: 0xFF} // Slightly lighter
	case theme.ColorNameDisabledButton:
		return color.RGBA{R: 0x1A, G: 0x1A, B: 0x1A, A: 0xFF}
	case theme.ColorNameInputBackground:
		return color.RGBA{R: 0x1E, G: 0x1F, B: 0x23, A: 0xFF} // #1E1F23
	case theme.ColorNameOverlayBackground:
		return color.RGBA{R: 0x1E, G: 0x1F, B: 0x23, A: 0xFF}
	case theme.ColorNameMenuBackground:
		return color.RGBA{R: 0x1E, G: 0x1F, B: 0x23, A: 0xFF}
	case theme.ColorNameHeaderBackground:
		return color.RGBA{R: 0x19, G: 0x1E, B: 0x1B, A: 0xFF}
	case theme.ColorNamePrimary:
		return color.RGBA{R: 0x5E, G: 0xA3, B: 0xFC, A: 0xFF} // Sky blue accent
	case theme.ColorNameFocus:
		return color.RGBA{R: 0x5E, G: 0xA3, B: 0xFC, A: 0x7F}
	case theme.ColorNameSelection:
		return color.RGBA{R: 0x5E, G: 0xA3, B: 0xFC, A: 0x3F}
	case theme.ColorNameHover:
		return color.RGBA{R: 0x3A, G: 0x3B, B: 0x3F, A: 0xFF} // #3A3B3F
	case theme.ColorNamePressed:
		return color.RGBA{R: 0x4A, G: 0x4B, B: 0x4F, A: 0xFF}
	case theme.ColorNameDisabled:
		return color.RGBA{R: 0x66, G: 0x66, B: 0x66, A: 0xFF}
	case theme.ColorNamePlaceHolder:
		return color.RGBA{R: 0x88, G: 0x88, B: 0x88, A: 0xFF}
	case theme.ColorNameSeparator:
		return color.RGBA{R: 0x3A, G: 0x3B, B: 0x3F, A: 0xFF} // #3A3B3F
	case theme.ColorNameScrollBar:
		return color.RGBA{R: 0x3A, G: 0x3B, B: 0x3F, A: 0xFF}
	case theme.ColorNameShadow:
		return color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x40}
	case theme.ColorNameForeground:
		return color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF} // White text
	case theme.ColorNameHyperlink:
		return color.RGBA{R: 0x5E, G: 0xA3, B: 0xFC, A: 0xFF}
	case theme.ColorNameSuccess:
		return color.RGBA{R: 0x4A, G: 0xB3, B: 0x4A, A: 0xFF} // Green
	case theme.ColorNameWarning:
		return color.RGBA{R: 0xFF, G: 0xA5, B: 0x00, A: 0xFF} // Orange
	case theme.ColorNameError:
		return color.RGBA{R: 0xE5, G: 0x4D, B: 0x4D, A: 0xFF} // Red
	case theme.ColorNameInputBorder:
		return color.RGBA{R: 0x3A, G: 0x3B, B: 0x3F, A: 0xFF}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Font returns the font resource for the specified text style
func (t *WaterAITheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon returns the icon resource for the specified icon name
func (t *WaterAITheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns the size for the specified size name
func (t *WaterAITheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 4
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameScrollBarSmall:
		return 6
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 24
	case theme.SizeNameSubHeadingText:
		return 18
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameInputRadius:
		return 4
	case theme.SizeNameSelectionRadius:
		return 4
	default:
		return theme.DefaultTheme().Size(name)
	}
}
