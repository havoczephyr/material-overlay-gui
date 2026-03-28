package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/havoczephyr/material-overlay-gui/internal/api"
	"github.com/havoczephyr/material-overlay-gui/internal/theme"
)

func (a *App) showPackDetail(set api.YGOProSet) {
	// Header
	header := sectionHeader(set.SetName)

	info := fmt.Sprintf("Code: %s  •  %d cards  •  Released: %s", set.SetCode, set.NumOfCards, set.TCGDate)
	infoLabel := canvas.NewText(info, theme.ColorFGDim)
	infoLabel.TextSize = 13

	// Card list container
	cardListBox := container.NewVBox(widget.NewLabel("Loading cards..."))
	cardScroll := container.NewVScroll(container.NewPadded(cardListBox))

	content := container.NewBorder(
		container.NewPadded(container.NewVBox(header, infoLabel)),
		nil, nil, nil,
		cardScroll,
	)

	a.setContent(content)

	// Load cards in background
	go func() {
		cards, err := a.svc.FetchCardsInSet(set.SetName)
		if err != nil {
			cardListBox.Objects = nil
			cardListBox.Add(widget.NewLabel("Failed to load cards: " + err.Error()))
			cardListBox.Refresh()
			return
		}

		if len(cards) == 0 {
			cardListBox.Objects = nil
			dimText := canvas.NewText("No cards found in this set.", theme.ColorFGDim)
			dimText.TextSize = 13
			cardListBox.Add(dimText)
			cardListBox.Refresh()
			return
		}

		cardListBox.Objects = nil
		for _, c := range cards {
			card := c
			cardListBox.Add(a.packCardRow(card, set.SetName))
		}
		cardListBox.Refresh()
	}()
}

func (a *App) packCardRow(card api.YGOProCard, setName string) fyne.CanvasObject {
	// Card name (clickable)
	nameBtn := widget.NewButton(card.Name, func() {
		a.showCardByName(card.Name)
	})
	nameBtn.Importance = widget.LowImportance
	nameBtn.Alignment = widget.ButtonAlignLeading

	// Find rarity for this specific set
	rarity := ""
	for _, cs := range card.CardSets {
		if cs.SetName == setName {
			rarity = cs.SetRarity
			break
		}
	}

	// Type info
	typeInfo := card.Type
	if card.Race != "" {
		typeInfo = card.Race + " / " + card.Type
	}

	// Build the row
	var details []string
	if rarity != "" {
		details = append(details, rarity)
	}
	details = append(details, typeInfo)

	detailText := canvas.NewText(strings.Join(details, "  •  "), theme.ColorFGDim)
	detailText.TextSize = 12

	return container.NewVBox(nameBtn, detailText)
}
