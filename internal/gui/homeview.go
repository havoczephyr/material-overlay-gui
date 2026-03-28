package gui

import (
	"bytes"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/havoczephyr/material-overlay-gui/internal/theme"
)

func (a *App) showHome() {
	// Card image area (left)
	imgWidget := placeholderImage()
	imgContainer := container.NewCenter(imgWidget)

	// Random card name label
	randomCardName := widget.NewRichTextFromMarkdown("")

	// "View Card" button - hidden until a card loads
	viewBtn := widget.NewButton("View Card", nil)
	viewBtn.Importance = widget.HighImportance
	viewBtn.Hide()

	// Refresh button
	refreshBtn := widget.NewButton("Random Card", func() {
		go a.loadRandomCard(imgContainer, randomCardName, viewBtn)
	})

	// Recent cards list
	recentHeader := sectionHeader("Recent Cards")
	recentList := container.NewVBox()

	a.buildRecentList(recentList)

	// Right panel
	rightPanel := container.NewVBox(
		sectionHeader("Welcome"),
		widget.NewLabel("Search for a card or load a random one."),
		layout.NewSpacer(),
		randomCardName,
		container.NewHBox(refreshBtn, viewBtn),
		layout.NewSpacer(),
		recentHeader,
		recentList,
	)

	// Main split layout
	leftSide := container.NewCenter(
		container.NewPadded(imgContainer),
	)

	split := container.NewHSplit(leftSide, container.NewPadded(rightPanel))
	split.SetOffset(0.35)

	a.setContent(split)

	// Load a random card on startup
	go a.loadRandomCard(imgContainer, randomCardName, viewBtn)
}

func (a *App) loadRandomCard(imgContainer *fyne.Container, nameLabel *widget.RichText, viewBtn *widget.Button) {
	ygoproCard, imgData, err := a.svc.LoadRandomCard()
	if err != nil {
		nameLabel.ParseMarkdown("*Failed to load random card*")
		return
	}

	nameLabel.ParseMarkdown("**" + ygoproCard.Name + "**")

	viewBtn.OnTapped = func() {
		a.showCardByName(ygoproCard.Name)
	}
	viewBtn.Show()

	if imgData != nil {
		img := canvas.NewImageFromReader(bytes.NewReader(imgData), ygoproCard.Name)
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(250, 365))
		imgContainer.Objects = []fyne.CanvasObject{img}
		imgContainer.Refresh()
	}
}

func (a *App) buildRecentList(box *fyne.Container) {
	box.Objects = nil
	if len(a.recentCards) == 0 {
		dimText := canvas.NewText("No recent cards", theme.ColorFGDim)
		dimText.TextSize = 13
		box.Add(dimText)
		return
	}
	for _, name := range a.recentCards {
		cardName := name
		btn := widget.NewButton(cardName, func() {
			a.showCardByName(cardName)
		})
		btn.Importance = widget.LowImportance
		box.Add(btn)
	}
}
