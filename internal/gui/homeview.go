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

	// ── Middle: Set browser with tabs ──
	latestList := container.NewVBox(widget.NewLabel("Loading sets..."))
	packsList := container.NewVBox(widget.NewLabel("Loading packs..."))
	structsList := container.NewVBox(widget.NewLabel("Loading structure decks..."))

	setTabs := container.NewAppTabs(
		container.NewTabItem("Latest", container.NewVScroll(container.NewPadded(latestList))),
		container.NewTabItem("Packs", container.NewVScroll(container.NewPadded(packsList))),
		container.NewTabItem("Structure Decks", container.NewVScroll(container.NewPadded(structsList))),
	)
	setTabs.SetTabLocation(container.TabLocationTop)

	// ── Bottom: Recent cards ──
	recentHeader := sectionHeader("Recent Cards")
	recentRow := container.NewHBox()
	a.buildRecentRow(recentRow)

	recentSection := container.NewVBox(recentHeader, container.NewHScroll(recentRow))

	// ── Assemble ──
	mainLayout := container.NewBorder(
		container.NewPadded(topBanner),
		container.NewPadded(recentSection),
		nil, nil,
		container.NewPadded(setTabs),
	)

	a.setContent(mainLayout)

	// Load random card
	go a.loadRandomCard(imgContainer, randomCardName, viewBtn)

	// Load sets in background
	go a.loadSetBrowser(latestList, packsList, structsList)
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
		img.SetMinSize(fyne.NewSize(180, 264))
		imgContainer.Objects = []fyne.CanvasObject{img}
		imgContainer.Refresh()
	}
}

func (a *App) loadSetBrowser(latestBox, packsBox, structsBox *fyne.Container) {
	sets, err := a.svc.FetchRecentSets(100)
	if err != nil {
		latestBox.Objects = nil
		latestBox.Add(widget.NewLabel("Failed to load sets: " + err.Error()))
		latestBox.Refresh()
		return
	}

	packs, structures := service.CategorizeRecentSets(sets)

	// Populate Latest (all sets, top 30)
	latestBox.Objects = nil
	limit := 30
	if len(sets) < limit {
		limit = len(sets)
	}
	for _, s := range sets[:limit] {
		set := s
		latestBox.Add(a.setRow(set))
	}
	latestBox.Refresh()

	// Populate Packs (top 30)
	packsBox.Objects = nil
	limit = 30
	if len(packs) < limit {
		limit = len(packs)
	}
	for _, s := range packs[:limit] {
		set := s
		packsBox.Add(a.setRow(set))
	}
	packsBox.Refresh()

	// Populate Structure Decks (top 30)
	structsBox.Objects = nil
	if len(structures) == 0 {
		dimText := canvas.NewText("No structure decks found.", theme.ColorFGDim)
		dimText.TextSize = 13
		structsBox.Add(dimText)
	} else {
		limit = 30
		if len(structures) < limit {
			limit = len(structures)
		}
		for _, s := range structures[:limit] {
			set := s
			structsBox.Add(a.setRow(set))
		}
	}
	structsBox.Refresh()
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
