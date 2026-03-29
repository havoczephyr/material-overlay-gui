package gui

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/havoczephyr/material-overlay-gui/internal/api"
	"github.com/havoczephyr/material-overlay-gui/internal/theme"
)

func (a *App) showArchetype(name string) {
	loading := container.NewCenter(widget.NewLabel("Loading " + name + "..."))
	a.setContent(loading)

	go func() {
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
		a.renderArchetypeView(name, article, artErr, imgData, cards, cardErr)
	}()
}

func (a *App) renderArchetypeView(name, article string, artErr error, imgData []byte, cards []api.YGOProCard, cardErr error) {
	// ── Article section (splash art + article) ──
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
		dimText := canvas.NewText("No article available.", theme.ColorFGDim)
		dimText.TextSize = 13
		articleWidget = dimText
	} else {
		articleWidget = a.wikiArticleToContent(article)
	}

	// Sticky header: image + title fixed at top, article scrolls underneath
	header := container.NewVBox(
		container.NewCenter(imgWidget),
		container.NewCenter(nameText),
	)
	articleSection := container.NewBorder(
		container.NewPadded(header), nil, nil, nil,
		container.NewVScroll(container.NewPadded(articleWidget)),
	)

	// ── Cards section ──
	cardHeader := sectionHeader(fmt.Sprintf("Cards in %s", name))
	cardListBox := container.NewVBox()

	if cardErr != nil {
		cardListBox.Add(widget.NewLabel("Failed to load cards: " + cardErr.Error()))
	} else if len(cards) == 0 {
		dimText := canvas.NewText("No cards found.", theme.ColorFGDim)
		dimText.TextSize = 13
		cardListBox.Add(dimText)
	} else {
		for _, c := range cards {
			cardListBox.Add(a.archetypeCardRow(c))
		}
	}

	cardSection := container.NewVBox(cardHeader, cardListBox)
	cardScroll := container.NewVScroll(container.NewPadded(cardSection))

	// ── Animated divider with toggle ──
	splitHeight := float32(300)
	expandedHeight := float32(650)
	collapsedHeight := float32(0)
	currentHeight := splitHeight
	state := 0 // 0=split, 1=article expanded, 2=cards expanded

	articleWrapper := container.NewGridWrap(fyne.NewSize(960, splitHeight), articleSection)

	var currentAnim *fyne.Animation
	animateTo := func(target float32) {
		if currentAnim != nil {
			currentAnim.Stop()
		}
		from := currentHeight
		currentAnim = canvas.NewSizeAnimation(
			fyne.NewSize(960, from),
			fyne.NewSize(960, target),
			time.Millisecond*300,
			func(s fyne.Size) {
				articleWrapper.Layout = layout.NewGridWrapLayout(s)
				articleWrapper.Refresh()
				a.contentBox.Refresh()
			},
		)
		currentAnim.Curve = fyne.AnimationEaseInOut
		currentAnim.Start()
		currentHeight = target
	}

	downBtn := widget.NewButton("▼", nil)
	upBtn := widget.NewButton("▲", nil)

	downBtn.OnTapped = func() {
		switch state {
		case 0: // split → article expanded
			animateTo(expandedHeight)
			state = 1
		case 2: // cards expanded → split
			animateTo(splitHeight)
			state = 0
		}
	}
	upBtn.OnTapped = func() {
		switch state {
		case 0: // split → cards expanded
			animateTo(collapsedHeight)
			state = 2
		case 1: // article expanded → split
			animateTo(splitHeight)
			state = 0
		}
	}

	toggleBg := canvas.NewRectangle(theme.ColorBGLight)
	toggleBg.SetMinSize(fyne.NewSize(0, 32))
	toggleBar := container.NewStack(
		toggleBg,
		container.NewCenter(container.NewHBox(downBtn, upBtn)),
	)

	// ── Assemble: article (top, animated) + toggle + cards (center) ──
	bottomSection := container.NewBorder(toggleBar, nil, nil, nil, cardScroll)
	fullView := container.NewBorder(articleWrapper, nil, nil, nil, bottomSection)

	a.setContent(fullView)
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
