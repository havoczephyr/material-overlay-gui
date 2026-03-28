package gui

import (
	"bytes"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/havoczephyr/material-overlay-gui/internal/card"
	"github.com/havoczephyr/material-overlay-gui/internal/theme"
)

func (a *App) showCardByName(name string) {
	// Show loading state
	loading := container.NewCenter(widget.NewLabel("Loading " + name + "..."))
	a.setContent(loading)

	go func() {
		cd, imgData, err := a.svc.LookupCard(name)
		if err != nil {
			errLabel := widget.NewLabel("Failed to load card: " + err.Error())
			a.setContent(container.NewCenter(errLabel))
			return
		}

		a.addRecentCard(cd.Name)
		a.renderCardView(cd, imgData)
	}()
}

func (a *App) renderCardView(cd *card.Card, imgData []byte) {
	// ── Left: Card Image ──
	var imgWidget fyne.CanvasObject
	if imgData != nil {
		img := canvas.NewImageFromReader(bytes.NewReader(imgData), cd.Name)
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(250, 365))
		imgWidget = img
	} else {
		imgWidget = placeholderImage()
	}

	// ── Right: Card Stats ──
	nameText := canvas.NewText(cd.Name, theme.ColorFG)
	nameText.TextSize = 22
	nameText.TextStyle.Bold = true

	statsBox := container.NewVBox(nameText)

	// Card type
	if cd.CardType != "" {
		statsBox.Add(labelValue("Card:  ", cd.CardType))
	}

	// Monster-specific stats
	if cd.IsMonster() {
		if cd.ATK != "" || cd.DEF != "" {
			atkDef := container.NewHBox(
				coloredValue("ATK: ", cd.ATK, theme.ColorATK),
			)
			if cd.IsLink() {
				atkDef.Add(coloredValue("  LINK: ", strings.Join(cd.LinkArrows, ", "), theme.ColorPrimary))
			} else {
				atkDef.Add(coloredValue("  DEF: ", cd.DEF, theme.ColorDEF))
			}
			statsBox.Add(atkDef)
		}

		if cd.Types != "" {
			statsBox.Add(labelValue("Type:  ", cd.Types))
		}

		if cd.Attribute != "" {
			attrRow := container.NewHBox(
				attributeDot(cd.Attribute),
				canvas.NewText(" "+cd.Attribute, theme.AttributeColor(cd.Attribute)),
			)
			statsBox.Add(attrRow)
		}

		if cd.Level != "" {
			statsBox.Add(labelValue("Level:  ", cd.Level))
		}
		if cd.Rank != "" {
			statsBox.Add(labelValue("Rank:  ", cd.Rank))
		}
		if cd.PendulumScale != "" {
			statsBox.Add(labelValue("Pendulum Scale:  ", cd.PendulumScale))
		}
	}

	// Spell/Trap property
	if cd.Property != "" {
		statsBox.Add(labelValue("Property:  ", cd.Property))
	}

	// Status badges
	badgeRow := container.NewHBox()
	if cd.TCGStatus != "" {
		badgeRow.Add(statusBadge("TCG", cd.TCGStatus))
	}
	if cd.OCGStatus != "" {
		badgeRow.Add(statusBadge("OCG", cd.OCGStatus))
	}
	if cd.GenesysCost != "" {
		badgeRow.Add(genesysBadge(cd.GenesysCost))
	}
	if len(badgeRow.Objects) > 0 {
		statsBox.Add(badgeRow)
	}

	// ── Top section: image + stats ──
	topLeft := container.NewCenter(container.NewPadded(imgWidget))
	topRight := container.NewPadded(statsBox)
	topSplit := container.NewHSplit(topLeft, topRight)
	topSplit.SetOffset(0.35)

	// Wrap top section in a sized container for gallery animation
	topWrapper := container.NewGridWrap(fyne.NewSize(960, 400), topSplit)

	// ── Bottom: Tabbed content ──
	tabs := a.buildCardTabs(cd, topWrapper)

	// Full layout
	fullView := container.NewBorder(
		container.NewPadded(topWrapper), nil, nil, nil,
		container.NewPadded(tabs),
	)

	a.setContent(fullView)
}

func (a *App) buildCardTabs(cd *card.Card, topWrapper *fyne.Container) *container.AppTabs {
	// Card Text tab
	cardTextContent := a.buildCardTextTab(cd)

	// Lazy-loaded tabs
	tipsContent := container.NewVBox(widget.NewLabel("Loading tips..."))
	triviaContent := container.NewVBox(widget.NewLabel("Loading trivia..."))
	rulingsContent := container.NewVBox(widget.NewLabel("Loading rulings..."))
	errataContent := container.NewVBox(widget.NewLabel("Loading errata..."))

	galleryContent := container.NewVBox(widget.NewLabel("Loading gallery..."))

	tabs := container.NewAppTabs(
		container.NewTabItem("Card Text", container.NewVScroll(container.NewPadded(cardTextContent))),
		container.NewTabItem("Tips", container.NewVScroll(container.NewPadded(tipsContent))),
		container.NewTabItem("Trivia", container.NewVScroll(container.NewPadded(triviaContent))),
		container.NewTabItem("Gallery", container.NewVScroll(container.NewPadded(galleryContent))),
		container.NewTabItem("Rulings", container.NewVScroll(container.NewPadded(rulingsContent))),
		container.NewTabItem("Errata", container.NewVScroll(container.NewPadded(errataContent))),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	// Gallery animation state
	fullHeight := float32(400)
	collapsedHeight := float32(160)
	wasGallery := false
	var currentAnim *fyne.Animation

	// Load tab content on selection
	loaded := map[string]bool{"Card Text": true}
	tabs.OnSelected = func(tab *container.TabItem) {
		isGallery := tab.Text == "Gallery"

		// Handle gallery animation
		if isGallery && !wasGallery {
			// Entering gallery: collapse top section
			if currentAnim != nil {
				currentAnim.Stop()
			}
			currentAnim = canvas.NewSizeAnimation(
				fyne.NewSize(960, fullHeight),
				fyne.NewSize(960, collapsedHeight),
				time.Millisecond*300,
				func(s fyne.Size) {
					topWrapper.Layout = layout.NewGridWrapLayout(s)
					topWrapper.Refresh()
				},
			)
			currentAnim.Curve = fyne.AnimationEaseInOut
			currentAnim.Start()
		} else if !isGallery && wasGallery {
			// Leaving gallery: expand top section back
			if currentAnim != nil {
				currentAnim.Stop()
			}
			currentAnim = canvas.NewSizeAnimation(
				fyne.NewSize(960, collapsedHeight),
				fyne.NewSize(960, fullHeight),
				time.Millisecond*300,
				func(s fyne.Size) {
					topWrapper.Layout = layout.NewGridWrapLayout(s)
					topWrapper.Refresh()
				},
			)
			currentAnim.Curve = fyne.AnimationEaseInOut
			currentAnim.Start()
		}
		wasGallery = isGallery

		// Lazy load content
		if loaded[tab.Text] {
			return
		}
		loaded[tab.Text] = true

		switch tab.Text {
		case "Tips":
			go a.loadTabContent(cd.Name, "tips", tipsContent)
		case "Trivia":
			go a.loadTabContent(cd.Name, "trivia", triviaContent)
		case "Rulings":
			go a.loadTabContent(cd.Name, "rulings", rulingsContent)
		case "Errata":
			go a.loadTabContent(cd.Name, "errata", errataContent)
		case "Gallery":
			go a.loadGallery(cd.Name, galleryContent)
		}
	}

	return tabs
}

func (a *App) buildCardTextTab(cd *card.Card) fyne.CanvasObject {
	content := container.NewVBox()

	// Lore
	if cd.Lore != "" {
		loreHeader := sectionHeader("Effect / Flavor Text")
		loreText := widget.NewLabel(cd.Lore)
		loreText.Wrapping = fyne.TextWrapWord
		content.Add(loreHeader)
		content.Add(loreText)
		content.Add(layout.NewSpacer())
	}

	// Archseries
	if len(cd.Archseries) > 0 {
		content.Add(labelValue("Archseries:  ", strings.Join(cd.Archseries, ", ")))
	}

	// Password
	if cd.Password != "" {
		content.Add(labelValue("Password:  ", cd.Password))
	}

	// Wiki URL
	if cd.WikiURL != "" {
		content.Add(labelValue("Wiki:  ", cd.WikiURL))
	}

	// Card Sets
	if len(cd.CardSets) > 0 {
		content.Add(layout.NewSpacer())
		content.Add(sectionHeader("Card Sets"))

		for _, cs := range cd.CardSets {
			setLine := cs.SetCode + " - " + cs.SetName
			if cs.Rarity != "" {
				setLine += " (" + cs.Rarity + ")"
			}
			setLabel := widget.NewLabel(setLine)
			setLabel.Wrapping = fyne.TextWrapWord
			content.Add(setLabel)
		}
	}

	return content
}

func (a *App) loadTabContent(cardName, tabType string, box *fyne.Container) {
	var rawText string
	var err error

	switch tabType {
	case "tips":
		rawText, err = a.svc.FetchTipsRaw(cardName)
	case "trivia":
		rawText, err = a.svc.FetchTriviaRaw(cardName)
	case "rulings":
		rawText, err = a.svc.FetchRulingsRaw(cardName)
	case "errata":
		rawText, err = a.svc.FetchErrataRaw(cardName)
	}

	box.Objects = nil
	if err != nil {
		box.Add(widget.NewLabel("Failed to load: " + err.Error()))
	} else if rawText == "" {
		dimText := canvas.NewText("No "+tabType+" available for this card.", theme.ColorFGDim)
		dimText.TextSize = 13
		box.Add(dimText)
	} else {
		rt := a.wikiTextToRichText(rawText)
		box.Add(rt)
	}
	box.Refresh()
}
