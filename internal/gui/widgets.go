package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
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
	t := canvas.NewText(text, theme.ColorPrimary)
	t.TextSize = 14
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

// placeholderThumb creates a small placeholder for a gallery thumbnail.
func placeholderThumb() fyne.CanvasObject {
	rect := canvas.NewRectangle(theme.ColorAccent)
	rect.SetMinSize(fyne.NewSize(80, 117))
	rect.CornerRadius = 4
	return rect
}

// thumbnailFrame wraps an image with a border that highlights when selected.
func thumbnailFrame(content fyne.CanvasObject, selected bool) fyne.CanvasObject {
	borderColor := theme.ColorBGLight
	if selected {
		borderColor = theme.ColorPrimary
	}
	border := canvas.NewRectangle(borderColor)
	border.SetMinSize(fyne.NewSize(86, 123))
	border.CornerRadius = 4
	return container.NewStack(border, container.NewCenter(content))
}

// tappableImage is a custom widget that wraps a canvas object and makes it tappable.
type tappableImage struct {
	widget.BaseWidget
	content  fyne.CanvasObject
	onTapped func()
}

func newTappableImage(content fyne.CanvasObject, onTapped func()) *tappableImage {
	t := &tappableImage{content: content, onTapped: onTapped}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tappableImage) Tapped(_ *fyne.PointEvent) {
	if t.onTapped != nil {
		t.onTapped()
	}
}

func (t *tappableImage) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.content)
}

func (t *tappableImage) MinSize() fyne.Size {
	return t.content.MinSize()
}

func (t *tappableImage) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

// tappableButton wraps a widget.Button to show a pointer cursor on hover.
type tappableButton struct {
	widget.BaseWidget
	btn *widget.Button
}

func newTappableButton(btn *widget.Button) *tappableButton {
	t := &tappableButton{btn: btn}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tappableButton) Tapped(e *fyne.PointEvent) {
	t.btn.Tapped(e)
}

func (t *tappableButton) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.btn)
}

func (t *tappableButton) MinSize() fyne.Size {
	return t.btn.MinSize()
}

func (t *tappableButton) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}
