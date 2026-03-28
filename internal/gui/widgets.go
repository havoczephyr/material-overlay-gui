package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/havoczephyr/material-overlay-gui/internal/theme"
)

// statusBadge creates a colored badge for a card status (Unlimited, Limited, etc.).
func statusBadge(label, status string) fyne.CanvasObject {
	bg := canvas.NewRectangle(theme.StatusColor(status))
	bg.SetMinSize(fyne.NewSize(0, 24))
	bg.CornerRadius = 4

	text := canvas.NewText(label+": "+status, color.Black)
	text.TextStyle.Bold = true
	text.TextSize = 12
	text.Alignment = fyne.TextAlignCenter

	return container.NewStack(bg, container.NewCenter(text))
}

// genesysBadge creates a colored badge for the Genesys point cost.
func genesysBadge(cost string) fyne.CanvasObject {
	if cost == "" {
		return widget.NewLabel("")
	}
	bg := canvas.NewRectangle(theme.GenesysCostColor(cost))
	bg.SetMinSize(fyne.NewSize(0, 24))
	bg.CornerRadius = 4

	text := canvas.NewText("GEN: "+cost+"pts", color.Black)
	text.TextStyle.Bold = true
	text.TextSize = 12
	text.Alignment = fyne.TextAlignCenter

	return container.NewStack(bg, container.NewCenter(text))
}

// attributeDot creates a small colored dot for the card attribute.
func attributeDot(attr string) fyne.CanvasObject {
	r := canvas.NewRectangle(theme.AttributeColor(attr))
	r.SetMinSize(fyne.NewSize(14, 14))
	r.CornerRadius = 7
	return r
}

// labelValue creates a styled label-value pair.
func labelValue(label, value string) fyne.CanvasObject {
	l := canvas.NewText(label, theme.ColorFGDim)
	l.TextSize = 13

	v := canvas.NewText(value, theme.ColorFG)
	v.TextSize = 14
	v.TextStyle.Bold = true

	return container.NewHBox(l, v)
}

// coloredValue creates a label-value pair with a custom color for the value.
func coloredValue(label string, value string, valueColor color.Color) fyne.CanvasObject {
	l := canvas.NewText(label, theme.ColorFGDim)
	l.TextSize = 13

	v := canvas.NewText(value, valueColor)
	v.TextSize = 14
	v.TextStyle.Bold = true

	return container.NewHBox(l, v)
}

// sectionHeader creates a styled section header.
func sectionHeader(text string) fyne.CanvasObject {
	t := canvas.NewText(text, theme.ColorGold)
	t.TextSize = 16
	t.TextStyle.Bold = true
	return t
}

// placeholderImage creates a placeholder rectangle for when no image is loaded.
func placeholderImage() fyne.CanvasObject {
	rect := canvas.NewRectangle(theme.ColorBGLight)
	rect.SetMinSize(fyne.NewSize(250, 365))
	rect.CornerRadius = 8
	return rect
}
