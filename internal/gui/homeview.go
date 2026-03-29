package gui

import (
	"bytes"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/havoczephyr/material-overlay-gui/internal/api"
	"github.com/havoczephyr/material-overlay-gui/internal/service"
	"github.com/havoczephyr/material-overlay-gui/internal/theme"
)

func (a *App) showHome() {
	// ── Top banner: smaller random card + name/buttons ──
	imgWidget := canvas.NewRectangle(theme.ColorBGLight)
	imgWidget.SetMinSize(fyne.NewSize(180, 264))
	imgWidget.CornerRadius = 8
	imgContainer := container.NewCenter(imgWidget)

	randomCardName := widget.NewRichTextFromMarkdown("")

	viewBtn := widget.NewButton("View Card", nil)
	viewBtn.Importance = widget.HighImportance
	viewBtn.Hide()

	refreshBtn := widget.NewButton("Random Card", func() {
		go a.loadRandomCard(imgContainer, randomCardName, viewBtn)
	})

	rightBanner := container.NewVBox(
		sectionHeader("Welcome"),
		widget.NewLabel("Search for a card or browse sets below."),
		layout.NewSpacer(),
		randomCardName,
		container.NewHBox(refreshBtn, viewBtn),
	)

	topBanner := container.NewHSplit(
		container.NewCenter(container.NewPadded(imgContainer)),
		container.NewPadded(rightBanner),
	)
	topBanner.SetOffset(0.25)

	// ── Middle: Set category buttons ──
	latestBtn := widget.NewButton("Latest Sets", func() {
		a.showSetList("Latest Sets", nil)
	})
	latestBtn.Importance = widget.HighImportance

	packsBtn := widget.NewButton("Booster Packs", func() {
		a.showSetList("Booster Packs", func(sets []api.YGOProSet) []api.YGOProSet {
			packs, _ := service.CategorizeRecentSets(sets)
			return packs
		})
	})
	packsBtn.Importance = widget.MediumImportance

	structsBtn := widget.NewButton("Structure Decks", func() {
		a.showSetList("Structure Decks", func(sets []api.YGOProSet) []api.YGOProSet {
			_, structs := service.CategorizeRecentSets(sets)
			return structs
		})
	})
	structsBtn.Importance = widget.MediumImportance

	browseSection := container.NewVBox(
		sectionHeader("Browse Sets"),
		container.NewGridWithColumns(3, latestBtn, packsBtn, structsBtn),
	)

	// ── Bottom: Recent cards ──
	recentHeader := sectionHeader("Recent Cards")
	recentRow := container.NewHBox()
	a.buildRecentRow(recentRow)
	recentSection := container.NewVBox(recentHeader, container.NewHScroll(recentRow))

	// ── Assemble ──
	mainLayout := container.NewVBox(
		container.NewPadded(topBanner),
		container.NewPadded(browseSection),
		container.NewPadded(recentSection),
	)

	a.setContent(container.NewVScroll(mainLayout))

	// Load random card
	go a.loadRandomCard(imgContainer, randomCardName, viewBtn)
}

func (a *App) loadRandomCard(imgContainer *fyne.Container, nameLabel *widget.RichText, viewBtn *widget.Button) {
	ygoproCard, imgData, err := a.svc.LoadRandomCard()
	if err != nil {
		fyne.Do(func() {
			nameLabel.ParseMarkdown("*Failed to load random card*")
		})
		return
	}

	cardName := ygoproCard.Name

	fyne.Do(func() {
		nameLabel.ParseMarkdown("**" + cardName + "**")

		viewBtn.OnTapped = func() {
			a.showCardByName(cardName)
		}
		viewBtn.Show()

		if imgData != nil {
			img := canvas.NewImageFromReader(bytes.NewReader(imgData), cardName)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(180, 264))
			imgContainer.Objects = []fyne.CanvasObject{img}
			imgContainer.Refresh()
		}
	})
}

// showSetList displays a full-page scrollable list of sets with optional filtering.
func (a *App) showSetList(title string, filter func([]api.YGOProSet) []api.YGOProSet) {
	header := sectionHeader(title)
	statusLabel := canvas.NewText("Loading sets...", theme.ColorFGDim)
	statusLabel.TextSize = 13

	listBox := container.NewVBox()
	content := container.NewBorder(
		container.NewPadded(container.NewVBox(header, statusLabel)),
		nil, nil, nil,
		container.NewVScroll(container.NewPadded(listBox)),
	)

	a.setContent(content)

	go func() {
		sets, err := a.svc.FetchRecentSets(100)
		if err != nil {
			fyne.Do(func() {
				statusLabel.Text = "Failed to load sets: " + err.Error()
				statusLabel.Refresh()
			})
			return
		}

		if filter != nil {
			sets = filter(sets)
		}

		limit := 50
		if len(sets) < limit {
			limit = len(sets)
		}

		// Build rows off main thread
		rows := make([]fyne.CanvasObject, limit)
		for i, s := range sets[:limit] {
			rows[i] = a.setRow(s)
		}

		fyne.Do(func() {
			statusLabel.Text = ""
			statusLabel.Refresh()
			for _, row := range rows {
				listBox.Add(row)
			}
			listBox.Refresh()
		})
	}()
}

func (a *App) setRow(set api.YGOProSet) fyne.CanvasObject {
	nameText := canvas.NewText(set.SetName, theme.ColorFG)
	nameText.TextSize = 14

	info := fmt.Sprintf("%s  •  %d cards", set.TCGDate, set.NumOfCards)
	if set.SetCode != "" {
		info = set.SetCode + "  •  " + info
	}
	infoText := canvas.NewText(info, theme.ColorFGDim)
	infoText.TextSize = 12

	label := container.NewVBox(nameText, infoText)

	btn := newTappableImage(label, func() {
		a.showPackDetail(set)
	})

	return container.NewPadded(btn)
}

func (a *App) buildRecentRow(row *fyne.Container) {
	row.Objects = nil
	if len(a.recentCards) == 0 {
		dimText := canvas.NewText("No recent cards", theme.ColorFGDim)
		dimText.TextSize = 13
		row.Add(dimText)
		return
	}
	for _, name := range a.recentCards {
		cardName := name
		btn := widget.NewButton(cardName, func() {
			a.showCardByName(cardName)
		})
		btn.Importance = widget.LowImportance
		row.Add(btn)
	}
}
