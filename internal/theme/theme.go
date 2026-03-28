package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// OverlayTheme implements a dark theme for material(Overlay).
type OverlayTheme struct{}

func (t *OverlayTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return ColorBG
	case theme.ColorNameForeground:
		return ColorFG
	case theme.ColorNamePrimary:
		return ColorPrimary
	case theme.ColorNameButton:
		return ColorBGLight
	case theme.ColorNameDisabled:
		return ColorFGDim
	case theme.ColorNameInputBackground:
		return ColorBGLight
	case theme.ColorNameInputBorder:
		return ColorAccent
	case theme.ColorNameSeparator:
		return ColorAccent
	case theme.ColorNameHeaderBackground:
		return ColorBGLight
	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

func (t *OverlayTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *OverlayTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *OverlayTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 4
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 20
	case theme.SizeNameSubHeadingText:
		return 16
	default:
		return theme.DefaultTheme().Size(name)
	}
}
