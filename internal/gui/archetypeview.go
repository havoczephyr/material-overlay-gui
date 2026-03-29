package gui

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/havoczephyr/material-overlay-gui/internal/api"
	"github.com/havoczephyr/material-overlay-gui/internal/theme"
)

func (a *App) showArchetype(name string) {
	loading := container.NewCenter(widget.NewLabel("Loading " + name + "..."))
	a.setContent(loading)

	go func() {
		// Fetch article, splash image, and card list in parallel
		var (
			article string
			imgData []byte
			cards   []api.YGOProCard
			artErr  error
			cardErr error
			wg      sync.WaitGroup
		)

		wg.Add(3)

		go func() {
			defer wg.Done()
			article, artErr = a.svc.FetchArchetypeArticle(name)
		}()

		go func() {
			defer wg.Done()
			imgData, _ = a.svc.FetchArchetypeSplashImage(name)
		}()

		go func() {
			defer wg.Done()
			cards, cardErr = a.svc.FetchArchetypeCards(name)
		}()

		wg.Wait()

		// Build the view
		a.renderArchetypeView(name, article, artErr, imgData, cards, cardErr)
	}()
}

func (a *App) renderArchetypeView(name, article string, artErr error, imgData []byte, cards []api.YGOProCard, cardErr error) {
	// ── Top: Splash art (left) + article (right) ──
	var imgWidget fyne.CanvasObject
	if imgData != nil {
		img := canvas.NewImageFromReader(bytes.NewReader(imgData), name)
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(200, 293))
		imgWidget = img
	} else {
		rect := canvas.NewRectangle(theme.ColorBGLight)
		rect.SetMinSize(fyne.NewSize(200, 293))
		rect.CornerRadius = 8
		imgWidget = rect
	}

	nameText := canvas.NewText(name, theme.ColorPrimary)
	nameText.TextSize = 22
	nameText.TextStyle.Bold = true

	var articleWidget fyne.CanvasObject
	if artErr != nil {
		articleWidget = widget.NewLabel("Failed to load article: " + artErr.Error())
	} else if article == "" {
		dimText := canvas.NewText("No article available for this archetype.", theme.ColorFGDim)
		dimText.TextSize = 13
		articleWidget = dimText
	} else {
		articleWidget = a.wikiTextToRichText(article)
	}

	articleBox := container.NewVBox(nameText, articleWidget)
	articleScroll := container.NewVScroll(container.NewPadded(articleBox))

	topSplit := container.NewHSplit(
		container.NewCenter(container.NewPadded(imgWidget)),
		articleScroll,
	)
	topSplit.SetOffset(0.25)

	// ── Bottom: Card list ──
	cardHeader := sectionHeader(fmt.Sprintf("Cards in %s", name))
	cardListBox := container.NewVBox()

	if cardErr != nil {
		cardListBox.Add(widget.NewLabel("Failed to load cards: " + cardErr.Error()))
	} else if len(cards) == 0 {
		dimText := canvas.NewText("No cards found for this archetype.", theme.ColorFGDim)
		dimText.TextSize = 13
		cardListBox.Add(dimText)
	} else {
		for _, c := range cards {
			cardListBox.Add(a.archetypeCardRow(c))
		}
	}

	cardSection := container.NewVBox(cardHeader, cardListBox)
	cardScroll := container.NewVScroll(container.NewPadded(cardSection))

	// Assemble: top section + card list in a vertical split
	mainSplit := container.NewVSplit(topSplit, cardScroll)
	mainSplit.SetOffset(0.4)

	a.setContent(mainSplit)
}

func (a *App) archetypeCardRow(c api.YGOProCard) fyne.CanvasObject {
	cardName := c.Name
	nameBtn := widget.NewButton(cardName, func() {
		a.showCardByName(cardName)
	})
	nameBtn.Importance = widget.LowImportance
	nameBtn.Alignment = widget.ButtonAlignLeading

	typeInfo := c.Type
	if c.Race != "" {
		typeInfo = c.Race + " / " + c.Type
	}

	var details []string
	if c.Attribute != "" {
		details = append(details, c.Attribute)
	}
	details = append(details, typeInfo)

	detailText := canvas.NewText(strings.Join(details, "  •  "), theme.ColorFGDim)
	detailText.TextSize = 12

	return container.NewVBox(nameBtn, detailText)
}
